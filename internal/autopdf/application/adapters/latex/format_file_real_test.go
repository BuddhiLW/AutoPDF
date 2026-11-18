// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package latex_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/latex"
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/infrastructure/adapters"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
)

// RealisticPreamble loads several common packages, the case format files
// are meant to speed up.
const realisticPreamble = `\documentclass[11pt]{article}
\usepackage[T1]{fontenc}
\usepackage{lmodern}
\usepackage{amsmath,amssymb}
\usepackage{graphicx}
\usepackage{xcolor}
\usepackage{tikz}
\usepackage{booktabs}
\usepackage[margin=2cm]{geometry}
\usepackage{hyperref}
`

const realisticBody = `\begin{document}
\section{Report}\label{s}
Value $x^2 + y^2 = z^2$ in \textcolor{red}{red}, see \ref{s}.
\begin{tabular}{ll}\toprule a & b \\ \bottomrule\end{tabular}
\tikz\draw (0,0) -- (1,1);
\end{document}
`

func requireMylatexformat(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("pdflatex"); err != nil {
		t.Skip("pdflatex is not installed")
	}
	kpse, err := exec.LookPath("kpsewhich")
	if err != nil {
		t.Skip("kpsewhich is not installed")
	}
	out, err := exec.Command(kpse, "mylatexformat.ltx").Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		t.Skip("mylatexformat.ltx is not installed")
	}
}

func pdfText(t *testing.T, path string) string {
	t.Helper()
	tool, err := exec.LookPath("pdftotext")
	if err != nil {
		t.Skip("pdftotext is not installed")
	}
	out, err := exec.Command(tool, path, "-").Output()
	if err != nil {
		t.Fatalf("pdftotext %s: %v", path, err)
	}
	return string(out)
}

func fmtLogLine(t *testing.T, logPath string) string {
	t.Helper()
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read %s: %v", logPath, err)
	}
	first, _, _ := strings.Cut(string(data), "\n")
	return first
}

func TestRealPdflatexFormatFileMatchesAPlainCompileAndIsReused(t *testing.T) {
	requireMylatexformat(t)
	work := t.TempDir()
	doc := realisticPreamble + realisticBody
	adapter := latex.NewLaTeXCompilerAdapterWithWorkingDir(
		&config.Config{}, adapters.NewOSFileSystem(), adapters.NewOSCommandExecutor(), work)
	compile := func(name string, formatFile bool) string {
		out := filepath.Join(work, name+".pdf")
		opts := ports.NewCompileOptions("pdflatex", out, work).WithJobName(name).
			WithBackendOptions(ports.NewLaTeXOptions(2, false).WithFormatFile(formatFile))
		path, err := adapter.Compile(context.Background(), doc, opts)
		if err != nil {
			t.Fatalf("compile %s: %v", name, err)
		}
		return path
	}

	plain := compile("plain", false)
	first := compile("first", true)
	fmtName := latex.FormatName("pdflatex", realisticPreamble)
	fmtPath := filepath.Join(work, latex.FormatCacheDir, fmtName+".fmt")
	built, err := os.Stat(fmtPath)
	if err != nil {
		t.Fatalf("the first build must leave %s: %v", fmtPath, err)
	}
	time.Sleep(10 * time.Millisecond)
	second := compile("second", true)

	if want, got := pdfText(t, plain), pdfText(t, first); want != got {
		t.Errorf("format-file output differs from the plain compile:\nplain:\n%s\nformat:\n%s", want, got)
	}
	if want, got := pdfText(t, plain), pdfText(t, second); want != got {
		t.Errorf("cached-format output differs from the plain compile")
	}
	if line := fmtLogLine(t, filepath.Join(work, "second.log")); !strings.Contains(line, "format="+fmtName) {
		t.Errorf("the second build must load the cached format, log says %q", line)
	}
	if line := fmtLogLine(t, filepath.Join(work, "plain.log")); strings.Contains(line, fmtName) {
		t.Errorf("the plain build must not load a format, log says %q", line)
	}
	again, err := os.Stat(fmtPath)
	if err != nil || !again.ModTime().Equal(built.ModTime()) {
		t.Errorf("the second build must reuse %s, not rebuild it", fmtPath)
	}
	_ = second
}

func TestRealXelatexFormatFileFallsBackWhenTheDumpFails(t *testing.T) {
	requireMylatexformat(t)
	if _, err := exec.LookPath("xelatex"); err != nil {
		t.Skip("xelatex is not installed")
	}
	work := t.TempDir()
	adapter := latex.NewLaTeXCompilerAdapterWithWorkingDir(
		&config.Config{}, adapters.NewOSFileSystem(), adapters.NewOSCommandExecutor(), work)
	doc := "\\documentclass{article}\n\\usepackage{fontspec}\n\\begin{document}Fallback\\end{document}\n"
	opts := ports.NewCompileOptions("xelatex", filepath.Join(work, "x.pdf"), work).WithJobName("x").
		WithBackendOptions(ports.NewLaTeXOptions(1, false).WithFormatFile(true))
	path, err := adapter.Compile(context.Background(), doc, opts)
	if err != nil {
		t.Fatalf("a preamble xelatex cannot dump must still compile: %v", err)
	}
	if !strings.Contains(pdfText(t, path), "Fallback") {
		t.Error("the fallback PDF lacks the document text")
	}
}
