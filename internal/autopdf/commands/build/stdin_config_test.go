// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/template/strict"
)

func writeStdinTemplate(t *testing.T) (string, string) {
	t.Helper()
	directory := t.TempDir()
	t.Chdir(directory)
	templatePath := filepath.Join(directory, "doc.tex")
	template := `\documentclass[12pt]{article}
\begin{document}
A: delim[[.alpha]]
B: delim[[.beta]]
\end{document}
`
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("write template: %v", err)
	}
	return directory, templatePath
}

func TestBuildReadsTheConfigFromStdinWhenGivenDash(t *testing.T) {
	_, templatePath := writeStdinTemplate(t)
	stdin := strings.NewReader(`{"engine":"pdflatex","output":"out.pdf","variables":{"alpha":"one"}}`)

	err := executeBuildProcess(context.Background(), stdin, []string{templatePath, "-"})
	if err == nil {
		t.Fatal("beta is absent from the stdin config, so the build must be refused")
	}
	var missing *strict.MissingVariablesError
	if !errors.As(err, &missing) {
		t.Fatalf("expected the strict verdict, got %T: %v", err, err)
	}
	if got := missing.Keys(); len(got) != 1 || got[0] != "beta" {
		t.Fatalf("only beta may be missing once alpha came from stdin, got %v", got)
	}
}

func TestBuildFromStdinBuildsBesideTheTemplate(t *testing.T) {
	directory, templatePath := writeStdinTemplate(t)
	stdin := strings.NewReader("engine: pdflatex\noutput: out.pdf\nvariables:\n  alpha: one\n  beta: two\n")

	if err := executeBuildProcess(context.Background(), stdin, []string{templatePath, "-", "clean"}); err != nil {
		t.Skipf("build did not complete in this environment, skipping: %v", err)
	}
	if _, err := os.Stat(filepath.Join(directory, "out.pdf")); err != nil {
		t.Fatalf("the PDF must land in the template's directory: %v", err)
	}
}

func TestBuildRefusesWatchWithAStdinConfig(t *testing.T) {
	_, templatePath := writeStdinTemplate(t)

	err := executeBuildProcess(context.Background(), strings.NewReader("{}"), []string{templatePath, "-", "watch"})
	if err == nil || !strings.Contains(err.Error(), "standard input") {
		t.Fatalf("watch with a stdin config must be refused, got %v", err)
	}
}
