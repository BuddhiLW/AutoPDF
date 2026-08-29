package preview

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/render/latex"
)

// Options controls workspace placement and idle expiry.
type Options struct {
	WorkspaceRoot    string
	IdleTTL          time.Duration
	KeepWorkspace    bool
	Resolver         AssetResolver
	WorkspaceFactory WorkspaceFactory
}

type job struct {
	ctx        context.Context
	revision   uint64
	projection latex.Projection
	reply      chan Result
}

// Session serializes compiles within one persistent LaTeX workspace.
type Session struct {
	compiler       Compiler
	resolver       AssetResolver
	workspace      string
	workspaceLease Workspace
	idleTTL        time.Duration

	mu           sync.Mutex
	closed       bool
	revision     uint64
	pending      *job
	activeCancel context.CancelFunc
	written      map[string]string
	pages        []Page
	timer        *time.Timer
	wake         chan struct{}
	stop         chan struct{}
	done         chan struct{}
	closeOnce    sync.Once
}

// NewSession creates a private workspace retained across preview revisions.
func NewSession(compiler Compiler, options Options) (*Session, error) {
	if compiler == nil {
		return nil, ErrCompilerRequired
	}
	factory := options.WorkspaceFactory
	if factory == nil {
		factory = localWorkspaceFactory{root: options.WorkspaceRoot, keep: options.KeepWorkspace}
	}
	workspace, err := factory.Create()
	if err != nil {
		return nil, fmt.Errorf("create preview workspace lease: %w", err)
	}
	if workspace == nil || strings.TrimSpace(workspace.Path()) == "" {
		if workspace != nil {
			_ = workspace.Close()
		}
		return nil, ErrWorkspaceRequired
	}
	session := &Session{
		compiler: compiler, resolver: options.Resolver, workspace: workspace.Path(), workspaceLease: workspace,
		idleTTL: options.IdleTTL,
		written: make(map[string]string), wake: make(chan struct{}, 1),
		stop: make(chan struct{}), done: make(chan struct{}),
	}
	go session.run()
	session.mu.Lock()
	session.touchLocked()
	session.mu.Unlock()
	return session, nil
}

// Workspace returns the stable directory used for this session.
func (session *Session) Workspace() string { return session.workspace }

// Closed reports whether Close or idle expiry ended the session.
func (session *Session) Closed() bool {
	session.mu.Lock()
	defer session.mu.Unlock()
	return session.closed
}

// Submit coalesces queued work, cancels an active older compile, and returns a
// one-shot result channel. Only the latest revision may publish output.
func (session *Session) Submit(ctx context.Context, projection latex.Projection) <-chan Result {
	if ctx == nil {
		ctx = context.Background()
	}
	reply := make(chan Result, 1)
	session.mu.Lock()
	if session.closed {
		session.mu.Unlock()
		deliver(reply, Result{Err: ErrClosed})
		return reply
	}
	session.revision++
	entry := &job{ctx: ctx, revision: session.revision, projection: projection, reply: reply}
	if session.pending != nil {
		deliver(session.pending.reply, Result{Revision: session.pending.revision, Err: ErrSuperseded})
	}
	session.pending = entry
	if session.activeCancel != nil {
		session.activeCancel()
	}
	session.touchLocked()
	session.mu.Unlock()
	session.notify()
	return reply
}

func (session *Session) run() {
	defer close(session.done)
	for {
		select {
		case <-session.stop:
			return
		case <-session.wake:
		}
		for {
			entry, compileCtx, cancel := session.takePending()
			if entry == nil {
				break
			}
			result := session.render(compileCtx, entry)
			cancel()
			session.mu.Lock()
			closed := session.closed
			latest := session.revision == entry.revision
			session.activeCancel = nil
			if latest && !closed && result.Err == nil {
				session.pages = clonePages(result.currentPages)
			}
			session.mu.Unlock()
			if closed {
				result.Result.Err = ErrClosed
			} else if !latest {
				result.Result.Err = ErrSuperseded
			}
			deliver(entry.reply, result.Result)
		}
	}
}

type renderedResult struct {
	Result
	currentPages []Page
}

func (session *Session) render(ctx context.Context, entry *job) renderedResult {
	started := time.Now()
	materializeStarted := time.Now()
	dirty, err := session.materialize(ctx, entry.projection)
	if err != nil {
		return renderedResult{Result: Result{Revision: entry.revision, Projection: entry.projection.Hash, Err: err}}
	}
	materializeElapsed := time.Since(materializeStarted)
	output, err := session.compiler.Compile(ctx, CompileRequest{
		Revision: entry.revision, Workspace: session.workspace, Main: entry.projection.Main, DirtyFiles: dirty,
		SourceMap: cloneSourceMap(entry.projection.SourceMap),
	})
	output.Timings.Materialize = materializeElapsed
	output.Timings.Total = time.Since(started)
	if err != nil {
		return renderedResult{Result: Result{Revision: entry.revision, Projection: entry.projection.Hash, DirtyFiles: dirty, Diagnostics: output.Diagnostics, Timings: output.Timings, Err: err}}
	}
	pages := normalizePages(output.Pages)
	session.mu.Lock()
	previous := clonePages(session.pages)
	session.mu.Unlock()
	changed, removed := diffPages(previous, pages)
	return renderedResult{
		Result:       Result{Revision: entry.revision, Projection: entry.projection.Hash, PDF: append([]byte(nil), output.PDF...), ChangedPages: changed, RemovedPages: removed, Diagnostics: append([]Diagnostic(nil), output.Diagnostics...), DirtyFiles: dirty, Timings: output.Timings},
		currentPages: pages,
	}
}

func cloneSourceMap(source map[string]latex.SourceLocation) map[string]latex.SourceLocation {
	if len(source) == 0 {
		return nil
	}
	result := make(map[string]latex.SourceLocation, len(source))
	for componentID, location := range source {
		result[componentID] = location
	}
	return result
}

// Done closes when manual close or idle expiry has released the workspace.
func (session *Session) Done() <-chan struct{} { return session.done }

func (session *Session) takePending() (*job, context.Context, context.CancelFunc) {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.closed || session.pending == nil {
		return nil, nil, func() {}
	}
	entry := session.pending
	session.pending = nil
	compileCtx, cancel := context.WithCancel(context.Background())
	stopCaller := context.AfterFunc(entry.ctx, cancel)
	wrappedCancel := func() { stopCaller(); cancel() }
	session.activeCancel = wrappedCancel
	return entry, compileCtx, wrappedCancel
}

func (session *Session) notify() {
	select {
	case session.wake <- struct{}{}:
	default:
	}
}

func (session *Session) touchLocked() {
	if session.idleTTL <= 0 {
		return
	}
	if session.timer != nil {
		session.timer.Stop()
	}
	session.timer = time.AfterFunc(session.idleTTL, func() { _ = session.Close() })
}

// Close cancels work and removes only this session's private workspace.
func (session *Session) Close() error {
	var closeErr error
	session.closeOnce.Do(func() {
		session.mu.Lock()
		session.closed = true
		if session.timer != nil {
			session.timer.Stop()
		}
		if session.activeCancel != nil {
			session.activeCancel()
		}
		if session.pending != nil {
			deliver(session.pending.reply, Result{Revision: session.pending.revision, Err: ErrClosed})
			session.pending = nil
		}
		close(session.stop)
		session.mu.Unlock()
		<-session.done
		closeErr = session.workspaceLease.Close()
	})
	return closeErr
}

func normalizePages(pages []Page) []Page {
	result := clonePages(pages)
	sort.Slice(result, func(i, j int) bool { return result[i].Number < result[j].Number })
	for index := range result {
		if result[index].Hash == "" {
			digest := sha256.Sum256(append([]byte(result[index].MediaType+"\x00"), result[index].Data...))
			result[index].Hash = hex.EncodeToString(digest[:])
		}
	}
	return result
}

func diffPages(previous, current []Page) ([]Page, []int) {
	old := make(map[int]string, len(previous))
	for _, page := range previous {
		old[page.Number] = page.Hash
	}
	changed := make([]Page, 0)
	currentNumbers := make(map[int]struct{}, len(current))
	for _, page := range current {
		currentNumbers[page.Number] = struct{}{}
		if old[page.Number] != page.Hash {
			changed = append(changed, page)
		}
	}
	removed := make([]int, 0)
	for number := range old {
		if _, exists := currentNumbers[number]; !exists {
			removed = append(removed, number)
		}
	}
	sort.Ints(removed)
	return changed, removed
}

func clonePages(pages []Page) []Page {
	result := append([]Page(nil), pages...)
	for index := range result {
		result[index].Data = append([]byte(nil), result[index].Data...)
	}
	return result
}

func deliver(channel chan Result, result Result) {
	select {
	case channel <- result:
	default:
	}
	close(channel)
}
