package preview_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/BuddhiLW/AutoPDF/pkg/preview"
	"github.com/BuddhiLW/AutoPDF/pkg/render/latex"
)

type runnerFunc func(context.Context, preview.Process) (preview.ProcessOutput, error)

func (function runnerFunc) Run(ctx context.Context, process preview.Process) (preview.ProcessOutput, error) {
	return function(ctx, process)
}

type rasterizerFunc func(context.Context, string, string, map[int]string) (preview.RasterOutput, error)

func (function rasterizerFunc) Rasterize(ctx context.Context, pdf, workspace string, previous map[int]string) (preview.RasterOutput, error) {
	return function(ctx, pdf, workspace, previous)
}

func TestLaTeXCompilerCancellationTerminatesRunner(t *testing.T) {
	started := make(chan struct{})
	runner := runnerFunc(func(ctx context.Context, _ preview.Process) (preview.ProcessOutput, error) {
		close(started)
		<-ctx.Done()
		return preview.ProcessOutput{}, ctx.Err()
	})
	compiler, err := preview.NewLaTeXCompiler(preview.LaTeXCompilerOptions{
		Runner: runner,
		Rasterizer: rasterizerFunc(func(context.Context, string, string, map[int]string) (preview.RasterOutput, error) {
			t.Fatal("rasterizer must not run")
			return preview.RasterOutput{}, nil
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	if err := os.WriteFile(filepath.Join(workspace, "main.tex"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, compileErr := compiler.Compile(ctx, preview.CompileRequest{Workspace: workspace, Main: "main.tex"})
		done <- compileErr
	}()
	<-started
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}

func TestLaTeXCompilerMapsGeneratedDiagnostics(t *testing.T) {
	runner := runnerFunc(func(context.Context, preview.Process) (preview.ProcessOutput, error) {
		return preview.ProcessOutput{Stderr: "fragments/body.tex:7: Undefined control sequence"}, errors.New("exit 1")
	})
	compiler, err := preview.NewLaTeXCompiler(preview.LaTeXCompilerOptions{Runner: runner, Rasterizer: rasterizerFunc(func(context.Context, string, string, map[int]string) (preview.RasterOutput, error) {
		return preview.RasterOutput{}, nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	if err := os.WriteFile(filepath.Join(workspace, "main.tex"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	output, err := compiler.Compile(context.Background(), preview.CompileRequest{
		Workspace: workspace, Main: "main.tex",
		SourceMap: map[string]latex.SourceLocation{"body": {Path: "fragments/body.tex", Line: 1}},
	})
	if err == nil || len(output.Diagnostics) != 1 || output.Diagnostics[0].ComponentID != "body" || output.Diagnostics[0].Line != 7 {
		t.Fatalf("diagnostic mapping failed: %+v %v", output.Diagnostics, err)
	}
}

func TestLaTeXCompilerWarmIntegrationOmitsUnchangedRasterPayloads(t *testing.T) {
	for _, executable := range []string{"pdflatex", "pdftoppm"} {
		if _, err := exec.LookPath(executable); err != nil {
			t.Skip(executable + " unavailable")
		}
	}
	compiler, err := preview.NewLaTeXCompiler(preview.LaTeXCompilerOptions{Engine: "pdflatex", DPI: 72})
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	source := []byte("\\documentclass{article}\n\\begin{document}\nHello warm preview.\n\\end{document}\n")
	if err := os.WriteFile(filepath.Join(workspace, "main.tex"), source, 0o644); err != nil {
		t.Fatal(err)
	}
	request := preview.CompileRequest{Revision: 1, Workspace: workspace, Main: "main.tex", DirtyFiles: []string{"main.tex"}}
	first, err := compiler.Compile(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	request.Revision = 2
	request.DirtyFiles = nil
	second, err := compiler.Compile(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Pages) != 1 || len(first.Pages[0].Data) == 0 {
		t.Fatalf("first compile did not raster page: %+v", first.Pages)
	}
	if len(second.Pages) != 1 || second.Pages[0].Hash != first.Pages[0].Hash || len(second.Pages[0].Data) != 0 {
		t.Fatalf("unchanged page was rasterized again: %+v", second.Pages)
	}
	if second.Timings.TeX <= 0 || second.Timings.Fingerprint <= 0 {
		t.Fatalf("missing layered timings: %+v", second.Timings)
	}
}
