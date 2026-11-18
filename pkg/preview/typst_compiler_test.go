package preview_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	typstadapter "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/typst"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/preview"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/latex"
)

// fakeTypst plays `typst compile`: pages of the main source are separated by
// "#pagebreak()", a .pdf target receives a PDF, a {p} target one PNG per page
// whose bytes are that page's source.
type fakeTypst struct {
	mu        sync.Mutex
	processes []preview.Process
	stderr    string
	exitCode  int
}

func (fake *fakeTypst) Run(_ context.Context, process preview.Process) (preview.ProcessOutput, error) {
	fake.mu.Lock()
	fake.processes = append(fake.processes, process)
	fake.mu.Unlock()
	if fake.exitCode != 0 {
		return preview.ProcessOutput{Stderr: fake.stderr, ExitCode: fake.exitCode}, errors.New("exit status " + strconv.Itoa(fake.exitCode))
	}
	var source, target string
	for index, arg := range process.Args {
		if strings.HasSuffix(arg, ".typ") {
			source, target = arg, process.Args[index+1]
		}
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return preview.ProcessOutput{ExitCode: 1}, err
	}
	pages := strings.Split(string(data), "#pagebreak()")
	if strings.Contains(target, "{p}") {
		for index, page := range pages {
			path := strings.ReplaceAll(target, "{p}", strconv.Itoa(index+1))
			if err := os.WriteFile(path, []byte(page), 0o644); err != nil {
				return preview.ProcessOutput{ExitCode: 1}, err
			}
		}
	} else if err := os.WriteFile(target, append([]byte("%PDF-1.7\n"), data...), 0o644); err != nil {
		return preview.ProcessOutput{ExitCode: 1}, err
	}
	return preview.ProcessOutput{Stderr: fake.stderr}, nil
}

func (fake *fakeTypst) calls() []preview.Process {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	return slices.Clone(fake.processes)
}

func writeMain(t *testing.T, workspace, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(workspace, "main.typ"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestTypstCompilerProducesPDFAndNativePages(t *testing.T) {
	runner := &fakeTypst{}
	compiler, err := preview.NewTypstCompiler(preview.TypstCompilerOptions{Runner: runner, PPI: 96, FontPaths: []string{"/fonts"}, Inputs: map[string]string{"k": "v"}})
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	writeMain(t, workspace, "one#pagebreak()two")
	output, err := compiler.Compile(context.Background(), preview.CompileRequest{Revision: 1, Workspace: workspace, Main: "main.typ"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(output.PDF), "%PDF-") {
		t.Fatalf("no PDF: %q", output.PDF)
	}
	if len(output.Pages) != 2 || string(output.Pages[0].Data) != "one" || string(output.Pages[1].Data) != "two" {
		t.Fatalf("pages wrong: %+v", output.Pages)
	}
	for index, page := range output.Pages {
		if page.Number != index+1 || page.MediaType != "image/png" || page.Hash == "" {
			t.Fatalf("page %d identity wrong: %+v", index, page)
		}
	}
	calls := runner.calls()
	if len(calls) != 2 {
		t.Fatalf("expected PDF and PNG runs, got %d", len(calls))
	}
	for _, call := range calls {
		if call.Executable != "typst" || call.Dir != workspace || call.Args[0] != "compile" {
			t.Fatalf("unexpected process: %+v", call)
		}
		joined := strings.Join(call.Args, " ")
		for _, want := range []string{"--root " + workspace, "--diagnostic-format short", "--font-path /fonts", "--input k=v", filepath.Join(workspace, "main.typ")} {
			if !strings.Contains(joined, want) {
				t.Fatalf("args %q lack %q", joined, want)
			}
		}
	}
	png := strings.Join(calls[1].Args, " ")
	if !strings.Contains(png, "{p}.png --format png --ppi 96") {
		t.Fatalf("PNG run args wrong: %q", png)
	}
	entries, _ := os.ReadDir(workspace)
	if len(entries) != 1 {
		t.Fatalf("compiler left output in the workspace: %v", entries)
	}
}

func TestTypstCompilerMapsShortDiagnostics(t *testing.T) {
	workspace := t.TempDir()
	runner := &fakeTypst{
		exitCode: 1,
		stderr: filepath.Join(workspace, "fragments/body.typ") + ":7:3: error: unknown variable: foo\n" +
			"  hint: did you mean bar?\nerror: no fonts\n",
	}
	compiler, err := preview.NewTypstCompiler(preview.TypstCompilerOptions{Runner: runner})
	if err != nil {
		t.Fatal(err)
	}
	writeMain(t, workspace, "x")
	output, err := compiler.Compile(context.Background(), preview.CompileRequest{
		Workspace: workspace, Main: "main.typ",
		SourceMap: map[string]preview.SourceLocation{"body": {Path: "fragments/body.typ", Line: 1}},
	})
	if !errors.Is(err, typstadapter.ErrCompileFailed) {
		t.Fatalf("expected typst compile error, got %v", err)
	}
	if len(output.Diagnostics) != 2 {
		t.Fatalf("diagnostics wrong: %+v", output.Diagnostics)
	}
	first := output.Diagnostics[0]
	if first.ComponentID != "body" || first.Path != "fragments/body.typ" || first.Line != 7 || first.Column != 3 ||
		first.Severity != "error" || first.Message != "unknown variable: foo" || len(first.Hints) != 1 {
		t.Fatalf("located diagnostic wrong: %+v", first)
	}
	if second := output.Diagnostics[1]; second.Path != "" || second.Message != "no fonts" {
		t.Fatalf("unlocated diagnostic wrong: %+v", second)
	}
	if len(runner.calls()) != 1 {
		t.Fatal("a failed PDF compile must not rasterize")
	}
}

func TestTypstCompilerReportsWarningsOnSuccess(t *testing.T) {
	runner := &fakeTypst{stderr: "main.typ:1:1: warning: unused\n"}
	compiler, err := preview.NewTypstCompiler(preview.TypstCompilerOptions{Runner: runner})
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	writeMain(t, workspace, "x")
	output, err := compiler.Compile(context.Background(), preview.CompileRequest{Workspace: workspace, Main: "main.typ"})
	if err != nil || len(output.Diagnostics) != 1 || output.Diagnostics[0].Severity != "warning" {
		t.Fatalf("warning not reported: %+v %v", output.Diagnostics, err)
	}
}

func TestTypstCompilerCancellationAndUnsafeMain(t *testing.T) {
	started := make(chan struct{})
	runner := runnerFunc(func(ctx context.Context, _ preview.Process) (preview.ProcessOutput, error) {
		close(started)
		<-ctx.Done()
		return preview.ProcessOutput{}, ctx.Err()
	})
	compiler, err := preview.NewTypstCompiler(preview.TypstCompilerOptions{Runner: runner})
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	writeMain(t, workspace, "x")
	if _, err := compiler.Compile(context.Background(), preview.CompileRequest{Workspace: workspace, Main: "../main.typ"}); !errors.Is(err, preview.ErrUnsafePath) {
		t.Fatalf("expected unsafe path, got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, compileErr := compiler.Compile(ctx, preview.CompileRequest{Workspace: workspace, Main: "main.typ"})
		done <- compileErr
	}()
	<-started
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}

func TestNewTypstCompilerRefusesInvalidInputs(t *testing.T) {
	if _, err := preview.NewTypstCompiler(preview.TypstCompilerOptions{Runner: &fakeTypst{}, Inputs: map[string]string{"a=b": "c"}}); !errors.Is(err, typstadapter.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func typstProjection(hash, content string) latex.Projection {
	return latex.Projection{Main: "main.typ", Hash: hash, Files: []latex.File{{Path: "main.typ", Content: []byte(content)}}}
}

func TestTypstCompilerSessionReportsOnlyChangedPages(t *testing.T) {
	compiler, err := preview.NewTypstCompiler(preview.TypstCompilerOptions{Runner: &fakeTypst{}})
	if err != nil {
		t.Fatal(err)
	}
	session, err := preview.NewSession(compiler, preview.Options{WorkspaceRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	first := <-session.Submit(context.Background(), typstProjection("1", "a#pagebreak()b#pagebreak()c"))
	second := <-session.Submit(context.Background(), typstProjection("2", "a#pagebreak()B#pagebreak()c"))
	third := <-session.Submit(context.Background(), typstProjection("3", "a#pagebreak()B"))
	for _, result := range []preview.Result{first, second, third} {
		if result.Err != nil {
			t.Fatal(result.Err)
		}
	}
	if len(first.ChangedPages) != 3 {
		t.Fatalf("first revision must send every page: %+v", first.ChangedPages)
	}
	if len(second.ChangedPages) != 1 || second.ChangedPages[0].Number != 2 || string(second.ChangedPages[0].Data) != "B" {
		t.Fatalf("second revision must send only page 2: %+v", second.ChangedPages)
	}
	if len(third.ChangedPages) != 0 || !slices.Equal(third.RemovedPages, []int{3}) {
		t.Fatalf("third revision must remove page 3 only: %+v %v", third.ChangedPages, third.RemovedPages)
	}
}

func TestTypstCompilerRealBinary(t *testing.T) {
	if _, err := exec.LookPath("typst"); err != nil {
		t.Skip("typst unavailable on PATH")
	}
	compiler, err := preview.NewTypstCompiler(preview.TypstCompilerOptions{PPI: 36})
	if err != nil {
		t.Fatal(err)
	}
	session, err := preview.NewSession(compiler, preview.Options{WorkspaceRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	first := <-session.Submit(context.Background(), typstProjection("1", "= One\nHello\n#pagebreak()\n= Two\nWorld\n"))
	if first.Err != nil {
		t.Fatalf("first compile: %v %+v", first.Err, first.Diagnostics)
	}
	if !strings.HasPrefix(string(first.PDF), "%PDF-") || len(first.ChangedPages) != 2 {
		t.Fatalf("first revision wrong: pdf=%d pages=%d", len(first.PDF), len(first.ChangedPages))
	}
	for _, page := range first.ChangedPages {
		if !strings.HasPrefix(string(page.Data), "\x89PNG") {
			t.Fatalf("page %d is not a PNG", page.Number)
		}
	}
	second := <-session.Submit(context.Background(), typstProjection("2", "= One\nHello\n#pagebreak()\n= Two\nChanged\n"))
	if second.Err != nil {
		t.Fatalf("second compile: %v", second.Err)
	}
	if len(second.ChangedPages) != 1 || second.ChangedPages[0].Number != 2 {
		t.Fatalf("only page 2 should change: %+v", len(second.ChangedPages))
	}
	broken := <-session.Submit(context.Background(), typstProjection("3", "#unknownfn()\n"))
	if broken.Err == nil || len(broken.Diagnostics) == 0 || broken.Diagnostics[0].Path != "main.typ" || broken.Diagnostics[0].Line != 1 {
		t.Fatalf("broken source diagnostics wrong: %v %+v", broken.Err, broken.Diagnostics)
	}
}
