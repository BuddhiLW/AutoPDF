// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package latex_test

import (
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/latex"
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports/compilertest"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
)

const contractDocument = `\documentclass{article}\begin{document}compilertest\end{document}`

var latexTemplateExtensions = []string{".tex", ".latex"}

func TestLaTeXCompilerAdapterHonoursTheDocumentCompilerContract(t *testing.T) {
	for _, passes := range []int{1, 3} {
		passes := passes
		t.Run(map[int]string{1: "SinglePass", 3: "ThreePasses"}[passes], func(t *testing.T) {
			compilertest.Run(t, func(t *testing.T, fs ports.FileSystem, exec ports.CommandExecutor) compilertest.Subject {
				return compilertest.Subject{
					Compiler:           latex.NewLaTeXCompilerAdapterWithWorkingDir(&config.Config{}, fs, exec, t.TempDir()),
					Engine:             "pdflatex",
					Backend:            ports.NewLaTeXOptions(passes, false),
					Content:            contractDocument,
					Binaries:           []string{"pdflatex"},
					TemplateExtensions: latexTemplateExtensions,
				}
			})
		})
	}
}

func TestLatexmkCompilerAdapterHonoursTheDocumentCompilerContract(t *testing.T) {
	compilertest.Run(t, func(t *testing.T, fs ports.FileSystem, exec ports.CommandExecutor) compilertest.Subject {
		return compilertest.Subject{
			Compiler:           latex.NewLatexmkCompilerAdapter(exec, fs),
			Engine:             "xelatex",
			Backend:            ports.NewLaTeXOptions(1, true),
			Content:            contractDocument,
			Binaries:           []string{"latexmk", "xelatex"},
			TemplateExtensions: latexTemplateExtensions,
		}
	})
}

func TestLaTeXAdapterRunsOneEngineInvocationPerPass(t *testing.T) {
	fs := compilertest.NewMemFS()
	exec := compilertest.NewFakeExecutor(fs)
	exec.SetPDFPath("/w/out.pdf")
	adapter := latex.NewLaTeXCompilerAdapterWithWorkingDir(&config.Config{}, fs, exec, "/w")

	opts := ports.NewCompileOptions("lualatex", "/w/out.pdf", "/w").
		WithJobName("out").
		WithBackendOptions(ports.NewLaTeXOptions(3, false))
	if _, err := adapter.Compile(t.Context(), contractDocument, opts); err != nil {
		t.Fatalf("compile: %v", err)
	}

	cmds := exec.CompileCommands()
	if len(cmds) != 3 {
		t.Fatalf("ran %d engine invocations, want 3", len(cmds))
	}
	for _, c := range cmds {
		if c.Executable != "lualatex" {
			t.Errorf("executable = %q, want lualatex", c.Executable)
		}
	}
}

func TestLaTeXOptionsDefaultToOnePassWhenAbsent(t *testing.T) {
	fs := compilertest.NewMemFS()
	exec := compilertest.NewFakeExecutor(fs)
	exec.SetPDFPath("/w/out.pdf")
	adapter := latex.NewLaTeXCompilerAdapterWithWorkingDir(&config.Config{}, fs, exec, "/w")

	opts := ports.NewCompileOptions("pdflatex", "/w/out.pdf", "/w").WithJobName("out")
	if _, err := adapter.Compile(t.Context(), contractDocument, opts); err != nil {
		t.Fatalf("compile: %v", err)
	}
	if n := len(exec.CompileCommands()); n != 1 {
		t.Fatalf("ran %d engine invocations, want 1", n)
	}
}
