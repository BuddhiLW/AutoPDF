// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package typst_test

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/typst"
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports/compilertest"
)

func TestCompileArgs(t *testing.T) {
	cases := []struct {
		name string
		opts ports.TypstOptions
		want []string
	}{
		{
			name: "defaults",
			opts: ports.TypstOptions{},
			want: []string{"compile", "--root", "/r", "--diagnostic-format", "short", "/w/doc.autopdf.typ", "/o/doc.pdf"},
		},
		{
			name: "font paths keep their order and skip blanks",
			opts: ports.TypstOptions{FontPaths: []string{"/z", "", "/a"}},
			want: []string{"compile", "--root", "/r", "--diagnostic-format", "short",
				"--font-path", "/z", "--font-path", "/a",
				"/w/doc.autopdf.typ", "/o/doc.pdf"},
		},
		{
			name: "inputs are sorted by key",
			opts: ports.TypstOptions{Inputs: map[string]string{"zeta": "1", "alpha": "a=b", "mid": ""}},
			want: []string{"compile", "--root", "/r", "--diagnostic-format", "short",
				"--input", "alpha=a=b", "--input", "mid=", "--input", "zeta=1",
				"/w/doc.autopdf.typ", "/o/doc.pdf"},
		},
		{
			name: "fonts before inputs",
			opts: ports.TypstOptions{FontPaths: []string{"/f"}, Inputs: map[string]string{"k": "v"}},
			want: []string{"compile", "--root", "/r", "--diagnostic-format", "short",
				"--font-path", "/f", "--input", "k=v",
				"/w/doc.autopdf.typ", "/o/doc.pdf"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := typst.CompileArgs("/r", "/w/doc.autopdf.typ", "/o/doc.pdf", tc.opts)
			if err != nil {
				t.Fatalf("CompileArgs: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("args\n got  %q\n want %q", got, tc.want)
			}
		})
	}
}

func TestCompileArgsIsDeterministic(t *testing.T) {
	opts := ports.TypstOptions{Inputs: map[string]string{"a": "1", "b": "2", "c": "3", "d": "4", "e": "5"}}
	first, _ := typst.CompileArgs("/r", "/s", "/o", opts)
	for i := 0; i < 50; i++ {
		got, _ := typst.CompileArgs("/r", "/s", "/o", opts)
		if !slices.Equal(got, first) {
			t.Fatalf("run %d: %q != %q", i, got, first)
		}
	}
}

func TestCompileArgsRefusesInputKeysTypstCannotCarry(t *testing.T) {
	for _, key := range []string{"", "a=b"} {
		_, err := typst.CompileArgs("/r", "/s", "/o", ports.TypstOptions{Inputs: map[string]string{key: "v"}})
		if !errors.Is(err, typst.ErrInvalidInput) {
			t.Errorf("key %q: err = %v, want ErrInvalidInput", key, err)
		}
	}
}

func newAdapter(t *testing.T) (*typst.TypstCompilerAdapter, *compilertest.MemFS, *compilertest.FakeExecutor) {
	t.Helper()
	fs := compilertest.NewMemFS()
	exec := compilertest.NewFakeExecutor(fs)
	exec.SetPDFPath("/w/out.pdf")
	return typst.NewTypstCompilerAdapter(fs, exec, typst.WithWorkingDir("/w")), fs, exec
}

func TestAdapterRunsOneTypstCompileWithStagedSourceAndRoot(t *testing.T) {
	adapter, _, exec := newAdapter(t)
	opts := ports.NewCompileOptions("typst", "/w/out.pdf", "/w").
		WithJobName("out").
		WithBackendOptions(ports.TypstOptions{FontPaths: []string{"/fonts"}, Inputs: map[string]string{"k": "v"}})

	path, err := adapter.Compile(t.Context(), contractDocument, opts)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if path != "/w/out.pdf" {
		t.Errorf("path = %q", path)
	}
	cmds := exec.CompileCommands()
	if len(cmds) != 1 {
		t.Fatalf("ran %d typst compiles, want 1 (single pass)", len(cmds))
	}
	want := []string{"compile", "--root", "/w", "--diagnostic-format", "short",
		"--font-path", "/fonts", "--input", "k=v", "/w/out" + typst.StagedSuffix, "/w/out.pdf"}
	if cmds[0].Executable != "typst" || !slices.Equal(cmds[0].Args, want) {
		t.Fatalf("command = %s %q, want typst %q", cmds[0].Executable, cmds[0].Args, want)
	}
	if cmds[0].Dir != "/w" {
		t.Errorf("command dir = %q, want /w", cmds[0].Dir)
	}
	probes := 0
	for _, c := range exec.Commands() {
		if c.Executable == "typst" && slices.Equal(c.Args, []string{"--version"}) {
			probes++
		}
	}
	if probes != 1 {
		t.Errorf("ran %d `typst --version` probes, want 1", probes)
	}
}

func TestAdapterUsesExplicitRoot(t *testing.T) {
	adapter, _, exec := newAdapter(t)
	opts := ports.NewCompileOptions("", "/w/out.pdf", "/w").WithJobName("out").
		WithBackendOptions(&ports.TypstOptions{Root: "/project"})
	if _, err := adapter.Compile(t.Context(), contractDocument, opts); err != nil {
		t.Fatalf("compile: %v", err)
	}
	args := exec.CompileCommands()[0].Args
	if args[1] != "--root" || args[2] != "/project" {
		t.Fatalf("args = %q, want --root /project", args)
	}
}

func TestAdapterRefusesLaTeXOnlyOptions(t *testing.T) {
	for name, latexOpts := range map[string]ports.LaTeXOptions{
		"passes>1": ports.NewLaTeXOptions(3, false),
		"latexmk":  ports.NewLaTeXOptions(1, true),
		"defaults": ports.DefaultLaTeXOptions(),
	} {
		t.Run(name, func(t *testing.T) {
			adapter, fs, exec := newAdapter(t)
			fs.Seed("/w/out.pdf", []byte("%PDF stale"))
			opts := ports.NewCompileOptions("typst", "/w/out.pdf", "/w").WithJobName("out").
				WithBackendOptions(latexOpts)
			_, err := adapter.Compile(t.Context(), contractDocument, opts)
			var mismatch *ports.BackendOptionsMismatchError
			if !errors.As(err, &mismatch) {
				t.Fatalf("err = %v, want *BackendOptionsMismatchError", err)
			}
			if mismatch.Want != ports.TypstBackendName || mismatch.Got != ports.LaTeXBackendName {
				t.Errorf("mismatch = %+v", mismatch)
			}
			if n := len(exec.Commands()); n != 0 {
				t.Errorf("refused options ran %d commands", n)
			}
			if fs.Exists("/w/out.pdf") {
				t.Error("a refused compile left a PDF")
			}
		})
	}
}

func TestAdapterRefusesALaTeXEngine(t *testing.T) {
	adapter, _, exec := newAdapter(t)
	opts := ports.NewCompileOptions("pdflatex", "/w/out.pdf", "/w").WithJobName("out")
	_, err := adapter.Compile(t.Context(), contractDocument, opts)
	if !errors.Is(err, typst.ErrUnsupportedEngine) {
		t.Fatalf("err = %v, want ErrUnsupportedEngine", err)
	}
	if n := len(exec.Commands()); n != 0 {
		t.Errorf("ran %d commands", n)
	}
}

// diagExecutor fails every non-probe command with the given stderr.
type diagExecutor struct {
	stderr string
}

func (d diagExecutor) Execute(ctx context.Context, cmd ports.Command) (ports.CommandResult, error) {
	if slices.Contains(cmd.Args, "--version") {
		return ports.NewCommandResult("typst 0.13.1\n", "", 0, 0), nil
	}
	return ports.NewCommandResult("", d.stderr, 1, 0), errors.New("exit status 1")
}

func TestAdapterReportsFailuresAsCompileErrorWithDiagnostics(t *testing.T) {
	fs := compilertest.NewMemFS()
	stderr := "/w/out.autopdf.typ:3:5: error: unknown variable: foo\n" +
		"/w/out.autopdf.typ:7:1: warning: unused label\n"
	adapter := typst.NewTypstCompilerAdapter(fs, diagExecutor{stderr: stderr}, typst.WithWorkingDir("/w"))
	fs.Seed("/w/out.pdf", []byte("%PDF stale"))

	opts := ports.NewCompileOptions("typst", "/w/out.pdf", "/w").WithJobName("out")
	path, err := adapter.Compile(t.Context(), contractDocument, opts)
	if path != "" {
		t.Errorf("path = %q, want empty", path)
	}
	var ce *typst.CompileError
	if !errors.As(err, &ce) {
		t.Fatalf("err = %v, want *typst.CompileError", err)
	}
	if !errors.Is(err, typst.ErrCompileFailed) {
		t.Error("errors.Is(err, ErrCompileFailed) = false")
	}
	if ce.ExitCode != 1 || len(ce.Diagnostics) != 2 {
		t.Fatalf("CompileError = %+v", ce)
	}
	errs := ce.Errors()
	if len(errs) != 1 || errs[0].Line != 3 || errs[0].Column != 5 || errs[0].Message != "unknown variable: foo" {
		t.Errorf("Errors() = %+v", errs)
	}
	if !strings.Contains(err.Error(), "3:5: error: unknown variable: foo") {
		t.Errorf("message %q lacks the diagnostic", err.Error())
	}
	if fs.Exists("/w/out.pdf") {
		t.Error("a failed compile left a PDF")
	}
	if fs.Exists("/w/out" + typst.StagedSuffix) {
		t.Error("staged file survived a non-debug failed compile")
	}
}

func TestAdapterRejectsEmptyContent(t *testing.T) {
	adapter, _, exec := newAdapter(t)
	_, err := adapter.Compile(t.Context(), "", ports.NewCompileOptions("typst", "/w/out.pdf", "/w"))
	if err == nil {
		t.Fatal("empty content compiled")
	}
	if n := len(exec.Commands()); n != 0 {
		t.Errorf("ran %d commands", n)
	}
}

func TestAdapterDefaultsOutputToJobNameInWorkingDir(t *testing.T) {
	fs := compilertest.NewMemFS()
	exec := compilertest.NewFakeExecutor(fs)
	exec.SetPDFPath("/w/report.pdf")
	adapter := typst.NewTypstCompilerAdapter(fs, exec)
	opts := ports.NewCompileOptions("typst", "", "/w").WithJobName("report")
	path, err := adapter.Compile(t.Context(), contractDocument, opts)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if path != "/w/report.pdf" {
		t.Fatalf("path = %q, want /w/report.pdf", path)
	}
}
