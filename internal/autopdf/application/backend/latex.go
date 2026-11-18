// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/latex"
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
)

// LaTeX engine names; the first is the default engine of the default backend.
const (
	EnginePDFLaTeX = "pdflatex"
	EngineXeLaTeX  = "xelatex"
	EngineLuaLaTeX = "lualatex"
)

// LaTeX returns the LaTeX backend descriptor.
func LaTeX() Backend {
	return Backend{
		Name:               ports.LaTeXBackendName,
		EngineNames:        []string{EnginePDFLaTeX, EngineXeLaTeX, EngineLuaLaTeX},
		TemplateExtensions: []string{".tex", ".latex"},
		StagedSuffix:       latex.StagedSuffix,
		AuxExtensions: []string{
			".aux", ".log", ".toc", ".lof", ".lot", ".out", ".nav", ".snm",
			".synctex.gz", ".fls", ".fdb_latexmk", ".bbl", ".blg", ".run.xml",
			".bcf", ".idx", ".ilg", ".ind", ".brf", ".vrb", ".xdv", ".dvi",
		},
		NewCompiler: newLaTeXCompiler,
		NewOptions: func(s Settings) (ports.BackendOptions, error) {
			if s.FormatFile && s.UseLatexmk {
				return nil, &UnsupportedSettingError{Backend: ports.LaTeXBackendName, Setting: "format_file with use_latexmk", Value: true}
			}
			return ports.NewLaTeXOptions(s.Passes, s.UseLatexmk).WithFormatFile(s.FormatFile), nil
		},
	}
}

func newLaTeXCompiler(deps CompilerDeps, opts ports.BackendOptions) (ports.DocumentCompiler, error) {
	latexOpts, err := ports.LaTeXOptionsFrom(ports.CompileOptions{Backend: opts})
	if err != nil {
		return nil, err
	}
	if latexOpts.UseLatexmk {
		return latex.NewLatexmkCompilerAdapterWithLogger(deps.Executor, deps.FileSystem, deps.Logger), nil
	}
	if deps.WorkingDir == "" {
		return latex.NewLaTeXCompilerAdapter(&config.Config{}, deps.FileSystem, deps.Executor).WithLogger(deps.Logger), nil
	}
	return latex.NewLaTeXCompilerAdapterWithWorkingDir(&config.Config{}, deps.FileSystem, deps.Executor, deps.WorkingDir).WithLogger(deps.Logger), nil
}
