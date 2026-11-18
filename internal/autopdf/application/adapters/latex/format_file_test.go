// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package latex_test

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/latex"
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports/compilertest"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
)

const fmtWork = "/w"

const fmtDoc = "\\documentclass{article}\n\\usepackage{mystyle}\n\\usepackage{xcolor}\n" +
	"\\begin{document}\nHello\n\\end{document}\n"

// fmtExecutor records every command. An -ini run dumps a stub format and a
// recorder file naming mystyle.sty; an engine run writes the PDF.
type fmtExecutor struct {
	mu          sync.Mutex
	fs          *compilertest.MemFS
	pdfPath     string
	iniFails    bool
	fmtUseFails bool
	cmds        []ports.Command
}

func (e *fmtExecutor) Execute(ctx context.Context, cmd ports.Command) (ports.CommandResult, error) {
	e.mu.Lock()
	e.cmds = append(e.cmds, cmd)
	e.mu.Unlock()
	switch {
	case cmd.Executable == "which":
		return ports.NewCommandResult("/usr/bin/"+cmd.Args[0], "", 0, 0), nil
	case slices.Contains(cmd.Args, "--version"):
		return ports.NewCommandResult("pdfTeX 3.14 (fake)\nmore\n", "", 0, 0), nil
	case slices.Contains(cmd.Args, "-ini"):
		if e.iniFails {
			return ports.NewCommandResult("", "! Can't \\dump a format with native fonts", 1, 0), errors.New("exit status 1")
		}
		job := strings.TrimPrefix(argWithPrefix(cmd.Args, "-jobname="), "-jobname=")
		out := filepath.Join(cmd.Dir, latex.FormatCacheDir, job)
		e.fs.Seed(out+".fmt", []byte("fake format"))
		e.fs.Seed(out+".fls", []byte("PWD /w\nINPUT /usr/share/texmf/tex/latex/base/article.cls\n"+
			"INPUT ./mystyle.sty\nINPUT mystyle.sty\nINPUT "+cmd.Args[len(cmd.Args)-1]+"\n"))
		return ports.NewCommandResult("", "", 0, 0), nil
	}
	if e.fmtUseFails && argWithPrefix(cmd.Args, "-fmt=") != "" {
		return ports.NewCommandResult("", "! format mismatch", 1, 0), errors.New("exit status 1")
	}
	e.fs.Seed(e.pdfPath, compilertest.FakePDF)
	return ports.NewCommandResult("", "", 0, 0), nil
}

func (e *fmtExecutor) withArg(prefix string) []ports.Command {
	e.mu.Lock()
	defer e.mu.Unlock()
	var out []ports.Command
	for _, c := range e.cmds {
		if argWithPrefix(c.Args, prefix) != "" {
			out = append(out, c)
		}
	}
	return out
}

func (e *fmtExecutor) compiles() []ports.Command {
	e.mu.Lock()
	defer e.mu.Unlock()
	var out []ports.Command
	for _, c := range e.cmds {
		if c.Executable == "pdflatex" && !slices.Contains(c.Args, "-ini") && !slices.Contains(c.Args, "--version") {
			out = append(out, c)
		}
	}
	return out
}

func argWithPrefix(args []string, prefix string) string {
	for _, a := range args {
		if strings.HasPrefix(a, prefix) {
			return a
		}
	}
	return ""
}

func newFmtFixture(t *testing.T) (*latex.LaTeXCompilerAdapter, *fmtExecutor, *compilertest.MemFS, ports.CompileOptions) {
	t.Helper()
	fs := compilertest.NewMemFS()
	fs.Seed(filepath.Join(fmtWork, "mystyle.sty"), []byte(`\newcommand\x{1}`))
	ex := &fmtExecutor{fs: fs, pdfPath: "/out/doc.pdf"}
	adapter := latex.NewLaTeXCompilerAdapterWithWorkingDir(&config.Config{}, fs, ex, fmtWork)
	opts := ports.NewCompileOptions("pdflatex", "/out/doc.pdf", fmtWork).WithJobName("doc").
		WithBackendOptions(ports.NewLaTeXOptions(1, false).WithFormatFile(true))
	return adapter, ex, fs, opts
}

func mustCompile(t *testing.T, adapter *latex.LaTeXCompilerAdapter, content string, opts ports.CompileOptions) {
	t.Helper()
	if _, err := adapter.Compile(context.Background(), content, opts); err != nil {
		t.Fatalf("compile: %v", err)
	}
}

func TestFormatFileIsOffByDefault(t *testing.T) {
	adapter, ex, _, opts := newFmtFixture(t)
	mustCompile(t, adapter, fmtDoc, opts.WithBackendOptions(ports.NewLaTeXOptions(1, false)))
	if n := len(ex.withArg("-ini")); n != 0 {
		t.Fatalf("ran %d -ini commands without format_file", n)
	}
	if n := len(ex.withArg("-fmt=")); n != 0 {
		t.Fatalf("passed -fmt %d times without format_file", n)
	}
}

func TestFormatFileBuildsASandboxedFormatOnceAndReusesIt(t *testing.T) {
	adapter, ex, fs, opts := newFmtFixture(t)
	mustCompile(t, adapter, fmtDoc, opts)
	mustCompile(t, adapter, fmtDoc, opts)

	inis := ex.withArg("-ini")
	if len(inis) != 1 {
		t.Fatalf("ran %d -ini commands over two builds, want 1", len(inis))
	}
	ini := inis[0]
	if ini.Executable != "pdflatex" || ini.Dir != fmtWork {
		t.Errorf("-ini ran %s in %s", ini.Executable, ini.Dir)
	}
	for _, want := range []string{"-no-shell-escape", "-interaction=nonstopmode", "-recorder",
		"-output-directory=" + latex.FormatCacheDir, "&pdflatex", "mylatexformat.ltx"} {
		if !slices.Contains(ini.Args, want) {
			t.Errorf("-ini args %v lack %s", ini.Args, want)
		}
	}
	assertSandboxed(t, ini, "-no-shell-escape")

	compiles := ex.withArg("-fmt=")
	if len(compiles) != 2 {
		t.Fatalf("%d engine runs used the format, want 2", len(compiles))
	}
	fmtArg := strings.TrimPrefix(argWithPrefix(compiles[0].Args, "-fmt="), "-fmt=")
	if filepath.IsAbs(fmtArg) || !strings.HasPrefix(fmtArg, latex.FormatCacheDir+"/pdflatex-") {
		t.Errorf("-fmt=%s must name the cache relative to the working dir", fmtArg)
	}
	if argWithPrefix(compiles[1].Args, "-fmt=") != "-fmt="+fmtArg {
		t.Errorf("second build used %v, want the cached %s", compiles[1].Args, fmtArg)
	}
	for _, c := range compiles {
		assertSandboxed(t, c, "-no-shell-escape")
	}
	if !fs.Exists(filepath.Join(fmtWork, fmtArg+".fmt")) || !fs.Exists(filepath.Join(fmtWork, fmtArg+".deps")) {
		t.Errorf("format and dependency manifest must be cached under %s", latex.FormatCacheDir)
	}
	for _, p := range fs.WrittenPaths() {
		if strings.HasSuffix(p, latex.StagedSuffix) && fs.Exists(p) {
			t.Errorf("staged file %s was left behind", p)
		}
	}
}

func TestFormatFileKeyChangesWithThePreambleAndTheEngine(t *testing.T) {
	a := latex.FormatName("pdflatex", "\\documentclass{article}\n")
	if b := latex.FormatName("pdflatex", "\\documentclass{report}\n"); a == b {
		t.Error("a different preamble must yield a different format")
	}
	if b := latex.FormatName("/usr/bin/pdflatex", "\\documentclass{article}\n"); a != b {
		t.Error("the engine path must not change the key, only its name")
	}
	if b := latex.FormatName("lualatex", "\\documentclass{article}\n"); a == b {
		t.Error("a different engine must yield a different format")
	}
}

func TestFormatFileRebuildsWhenThePreambleChanges(t *testing.T) {
	adapter, ex, _, opts := newFmtFixture(t)
	mustCompile(t, adapter, fmtDoc, opts)
	mustCompile(t, adapter, strings.Replace(fmtDoc, "xcolor", "graphicx", 1), opts)
	mustCompile(t, adapter, strings.Replace(fmtDoc, "Hello", "Changed body", 1), opts)
	if n := len(ex.withArg("-ini")); n != 2 {
		t.Fatalf("ran %d -ini commands, want 2 (body edits reuse the format)", n)
	}
}

func TestFormatFileRebuildsWhenALocalPreambleInputChanges(t *testing.T) {
	adapter, ex, fs, opts := newFmtFixture(t)
	mustCompile(t, adapter, fmtDoc, opts)
	fs.Seed(filepath.Join(fmtWork, "mystyle.sty"), []byte(`\newcommand\x{2}`))
	mustCompile(t, adapter, fmtDoc, opts)
	if n := len(ex.withArg("-ini")); n != 2 {
		t.Fatalf("ran %d -ini commands, want 2 after mystyle.sty changed", n)
	}
}

func TestFormatFileBuildFailureFallsBackToAPlainCompile(t *testing.T) {
	adapter, ex, _, opts := newFmtFixture(t)
	ex.iniFails = true
	mustCompile(t, adapter, fmtDoc, opts)
	mustCompile(t, adapter, fmtDoc, opts)
	if n := len(ex.withArg("-fmt=")); n != 0 {
		t.Fatalf("%d engine runs used a format that failed to build", n)
	}
	if n := len(ex.compiles()); n != 2 {
		t.Fatalf("ran %d plain compiles, want 2", n)
	}
	if n := len(ex.withArg("-ini")); n != 1 {
		t.Fatalf("ran %d -ini commands, want 1: a failed dump is remembered per engine version", n)
	}
}

func TestFormatFileUseFailureRetriesWithoutTheFormat(t *testing.T) {
	adapter, ex, fs, opts := newFmtFixture(t)
	ex.fmtUseFails = true
	mustCompile(t, adapter, fmtDoc, opts)
	compiles := ex.compiles()
	if len(compiles) != 2 || argWithPrefix(compiles[1].Args, "-fmt=") != "" {
		t.Fatalf("want a -fmt run then a plain retry, got %d runs", len(compiles))
	}
	fmtArg := strings.TrimPrefix(argWithPrefix(compiles[0].Args, "-fmt="), "-fmt=")
	if fs.Exists(filepath.Join(fmtWork, fmtArg+".fmt")) {
		t.Error("a format the engine could not use must be discarded")
	}
}

func TestFormatFileWithoutAPreambleCompilesPlainly(t *testing.T) {
	adapter, ex, _, opts := newFmtFixture(t)
	mustCompile(t, adapter, "\\begin{document}x\\end{document}", opts)
	if n := len(ex.withArg("-ini")); n != 0 {
		t.Fatalf("ran %d -ini commands for a document without a preamble", n)
	}
}

func TestFormatFileStopsAtEndofdump(t *testing.T) {
	adapter, ex, fs, opts := newFmtFixture(t)
	doc := "\\documentclass{article}\n\\csname endofdump\\endcsname\n\\newcommand\\who{A}\n\\begin{document}\\who\\end{document}\n"
	mustCompile(t, adapter, doc, opts.WithDebug(true))
	mustCompile(t, adapter, strings.Replace(doc, "{A}", "{B}", 1), opts.WithDebug(true))
	if n := len(ex.withArg("-ini")); n != 1 {
		t.Fatalf("ran %d -ini commands, want 1: lines after \\endofdump are not dumped", n)
	}
	for _, p := range fs.WrittenPaths() {
		data, _ := fs.Contents(p)
		if strings.Contains(p, "pdflatex-") && strings.Contains(string(data), `\who`) {
			t.Errorf("staged preamble %s holds content after \\endofdump", p)
		}
	}
}
