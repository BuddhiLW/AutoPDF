// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package latex_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/latex"
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/infrastructure/adapters"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
)

// recordingExecutor keeps every command it is handed and writes a stub PDF
// where the adapter will look for one.
type recordingExecutor struct {
	pdfPath string
	seen    []ports.Command
}

func (e *recordingExecutor) Execute(ctx context.Context, cmd ports.Command) (ports.CommandResult, error) {
	e.seen = append(e.seen, cmd)
	if cmd.Executable == "which" {
		return ports.NewCommandResult("/usr/bin/pdflatex", "", 0, 0), nil
	}
	if err := os.WriteFile(e.pdfPath, []byte("%PDF-1.4 stub"), 0o600); err != nil {
		return ports.CommandResult{}, err
	}
	return ports.NewCommandResult("", "", 0, 0), nil
}

func (e *recordingExecutor) compileCommands(executable string) []ports.Command {
	var out []ports.Command
	for _, c := range e.seen {
		if c.Executable == executable && !slices.Contains(c.Args, "-c") {
			out = append(out, c)
		}
	}
	return out
}

func assertSandboxed(t *testing.T, cmd ports.Command, shellFlag string) {
	t.Helper()
	if !slices.Contains(cmd.Args, shellFlag) {
		t.Errorf("%s must carry %s, got %v", cmd.Executable, shellFlag, cmd.Args)
	}
	for _, want := range []string{"openin_any=p", "openout_any=p", "shell_escape=f"} {
		if !slices.Contains(cmd.Env, want) {
			t.Errorf("%s env must carry %s", cmd.Executable, want)
		}
	}
	input := cmd.Args[len(cmd.Args)-1]
	if filepath.IsAbs(input) {
		t.Errorf("the staged input must be passed relative to the working dir, got %s", input)
	}
}

func TestLatexmkRunsWithoutShellEscapeAndParanoidReads(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "doc.pdf")
	ex := &recordingExecutor{pdfPath: out}
	adapter := latex.NewLatexmkCompilerAdapter(ex, adapters.NewOSFileSystem())
	opts := ports.NewCompileOptions("pdflatex", out, dir).WithJobName("doc")

	if _, err := adapter.Compile(context.Background(), "x", opts); err != nil {
		t.Fatalf("compile: %v", err)
	}
	cmds := ex.compileCommands("latexmk")
	if len(cmds) == 0 {
		t.Fatal("no latexmk command was run")
	}
	assertSandboxed(t, cmds[0], "-latexoption=-no-shell-escape")
}

func TestEngineRunsWithoutShellEscapeAndParanoidReads(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "doc.pdf")
	ex := &recordingExecutor{pdfPath: out}
	adapter := latex.NewLaTeXCompilerAdapterWithWorkingDir(
		&config.Config{}, adapters.NewOSFileSystem(), ex, dir)
	opts := ports.NewCompileOptions("pdflatex", out, dir).WithJobName("doc")

	if _, err := adapter.Compile(context.Background(), "x", opts); err != nil {
		t.Fatalf("compile: %v", err)
	}
	cmds := ex.compileCommands("pdflatex")
	if len(cmds) == 0 {
		t.Fatal("no pdflatex command was run")
	}
	for _, c := range cmds {
		assertSandboxed(t, c, "-no-shell-escape")
	}
}

func realEngine(t *testing.T) (*latex.LaTeXCompilerAdapter, string, string) {
	t.Helper()
	if _, err := exec.LookPath("pdflatex"); err != nil {
		t.Skip("pdflatex is not installed")
	}
	work := t.TempDir()
	outDir := t.TempDir()
	adapter := latex.NewLaTeXCompilerAdapterWithWorkingDir(
		&config.Config{}, adapters.NewOSFileSystem(), adapters.NewOSCommandExecutor(), work)
	return adapter, work, filepath.Join(outDir, "doc.pdf")
}

func TestRealPdflatexRefusesAnAbsoluteInput(t *testing.T) {
	adapter, work, out := realEngine(t)
	opts := ports.NewCompileOptions("pdflatex", out, work).WithJobName("doc")
	doc := "\\documentclass{article}\\begin{document}X\\input{/etc/hostname}Y\\end{document}\n"

	if _, err := adapter.Compile(context.Background(), doc, opts); err == nil {
		t.Fatal("a template that \\input's /etc/hostname must not compile")
	}
}

func TestRealPdflatexStillCompilesIntoASeparateOutputDir(t *testing.T) {
	adapter, work, out := realEngine(t)
	opts := ports.NewCompileOptions("pdflatex", out, work).WithJobName("doc")
	opts.Passes = 2
	doc := "\\documentclass{article}\\begin{document}\\section{A}\\label{a}See \\ref{a}.\\end{document}\n"

	path, err := adapter.Compile(context.Background(), doc, opts)
	if err != nil {
		t.Fatalf("a plain two-pass document must compile under the sandbox: %v", err)
	}
	if info, err := os.Stat(path); err != nil || info.Size() == 0 {
		t.Fatalf("expected a PDF at %s: %v", path, err)
	}
}
