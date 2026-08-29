package preview_test

import (
	"context"
	"errors"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/preview"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/latex"
)

type compilerFunc func(context.Context, preview.CompileRequest) (preview.CompileOutput, error)

func (function compilerFunc) Compile(ctx context.Context, request preview.CompileRequest) (preview.CompileOutput, error) {
	return function(ctx, request)
}

type workspaceFactory struct{ workspace *workspace }

func (factory workspaceFactory) Create() (preview.Workspace, error) { return factory.workspace, nil }

type workspace struct {
	path   string
	closed bool
}

func (workspace *workspace) Path() string { return workspace.path }
func (workspace *workspace) Close() error {
	workspace.closed = true
	return nil
}

func projection(hash, content string) latex.Projection {
	return latex.Projection{Main: "main.tex", Hash: hash, Files: []latex.File{{Path: "main.tex", Content: []byte(content)}}}
}

func TestSessionReusesWorkspaceWritesOnlyDirtyFilesAndDiffsPages(t *testing.T) {
	var mu sync.Mutex
	var requests []preview.CompileRequest
	compiler := compilerFunc(func(_ context.Context, request preview.CompileRequest) (preview.CompileOutput, error) {
		mu.Lock()
		requests = append(requests, request)
		mu.Unlock()
		data, err := os.ReadFile(request.Workspace + "/main.tex")
		if err != nil {
			return preview.CompileOutput{}, err
		}
		return preview.CompileOutput{PDF: data, Pages: []preview.Page{{Number: 1, MediaType: "image/png", Data: data}}}, nil
	})
	session, err := preview.NewSession(compiler, preview.Options{WorkspaceRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	first := <-session.Submit(context.Background(), projection("one", "A"))
	second := <-session.Submit(context.Background(), projection("one", "A"))
	third := <-session.Submit(context.Background(), projection("two", "B"))
	if first.Err != nil || second.Err != nil || third.Err != nil {
		t.Fatalf("unexpected results: %v %v %v", first.Err, second.Err, third.Err)
	}
	if len(first.ChangedPages) != 1 || len(second.ChangedPages) != 0 || len(third.ChangedPages) != 1 {
		t.Fatalf("page diffs wrong: %d %d %d", len(first.ChangedPages), len(second.ChangedPages), len(third.ChangedPages))
	}
	mu.Lock()
	defer mu.Unlock()
	if !reflect.DeepEqual(requests[0].DirtyFiles, []string{"main.tex"}) || len(requests[1].DirtyFiles) != 0 || !reflect.DeepEqual(requests[2].DirtyFiles, []string{"main.tex"}) {
		t.Fatalf("incremental writes wrong: %+v", requests)
	}
	if requests[0].Workspace != requests[2].Workspace {
		t.Fatal("workspace was not retained")
	}
}

func TestSessionCancelsSupersededRevisionAndPublishesLatest(t *testing.T) {
	started := make(chan struct{})
	compiler := compilerFunc(func(ctx context.Context, request preview.CompileRequest) (preview.CompileOutput, error) {
		if request.Revision == 1 {
			close(started)
			<-ctx.Done()
			return preview.CompileOutput{}, ctx.Err()
		}
		return preview.CompileOutput{Pages: []preview.Page{{Number: 1, Data: []byte("latest")}}}, nil
	})
	session, err := preview.NewSession(compiler, preview.Options{WorkspaceRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	old := session.Submit(context.Background(), projection("old", "old"))
	<-started
	latest := session.Submit(context.Background(), projection("latest", "latest"))
	if result := <-old; !errors.Is(result.Err, preview.ErrSuperseded) {
		t.Fatalf("expected superseded result, got %v", result.Err)
	}
	if result := <-latest; result.Err != nil || result.Revision != 2 || len(result.ChangedPages) != 1 {
		t.Fatalf("latest result wrong: %+v", result)
	}
}

func TestSessionExpiresAndRejectsUnsafeProjection(t *testing.T) {
	compiler := compilerFunc(func(context.Context, preview.CompileRequest) (preview.CompileOutput, error) {
		return preview.CompileOutput{}, nil
	})
	session, err := preview.NewSession(compiler, preview.Options{WorkspaceRoot: t.TempDir(), IdleTTL: 20 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for !session.Closed() && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if !session.Closed() {
		t.Fatal("session did not expire")
	}
	if result := <-session.Submit(context.Background(), projection("late", "late")); !errors.Is(result.Err, preview.ErrClosed) {
		t.Fatalf("expected closed error, got %v", result.Err)
	}

	unsafe, err := preview.NewSession(compiler, preview.Options{WorkspaceRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer unsafe.Close()
	result := <-unsafe.Submit(context.Background(), latex.Projection{Main: "main.tex", Files: []latex.File{{Path: "../escape", Content: []byte("x")}}})
	if !errors.Is(result.Err, preview.ErrUnsafePath) {
		t.Fatalf("expected unsafe path error, got %v", result.Err)
	}
}

func TestSessionUsesInjectedWorkspaceLifecycle(t *testing.T) {
	compiler := compilerFunc(func(context.Context, preview.CompileRequest) (preview.CompileOutput, error) {
		return preview.CompileOutput{}, nil
	})
	lease := &workspace{path: t.TempDir()}
	session, err := preview.NewSession(compiler, preview.Options{WorkspaceFactory: workspaceFactory{workspace: lease}})
	if err != nil {
		t.Fatal(err)
	}
	if session.Workspace() != lease.path {
		t.Fatalf("workspace mismatch: %s", session.Workspace())
	}
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}
	if !lease.closed {
		t.Fatal("workspace lease was not closed")
	}
}
