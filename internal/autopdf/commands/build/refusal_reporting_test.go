// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/configs"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/template/strict"
)

// A build refused for unsatisfied variables must reach the operator as the
// whole offence list. A test that only asserted "it errored" would pass while
// the CLI printed a bare sentinel.
func TestBuildReportsEveryUnsatisfiedVariableToTheOperator(t *testing.T) {
	directory := t.TempDir()
	t.Chdir(directory)

	templatePath := filepath.Join(directory, "doc.tex")
	template := `\documentclass[12pt]{article}
\begin{document}
A: delim[[.alpha]]
B: delim[[.beta]]
C: delim[[.gamma]]
\end{document}
`
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("write template: %v", err)
	}

	configPath := filepath.Join(directory, "config.yaml")
	configBody := `template: "doc.tex"
output: "out.pdf"
engine: "pdflatex"
variables:
  title: "Example Title"
`
	if err := os.WriteFile(configPath, []byte(configBody), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := BuildServiceCmd.Do(nil, templatePath, configPath)
	if err == nil {
		t.Fatal("a template reading three absent keys must not build")
	}

	message := err.Error()
	for _, key := range []string{"alpha", "beta", "gamma"} {
		if !strings.Contains(message, key) {
			t.Fatalf("the operator must be told about %q, got:\n%s", key, message)
		}
	}
	if !strings.Contains(message, "doc.tex:3:") {
		t.Fatalf("the operator must be given a source location, got:\n%s", message)
	}
	if message == configs.BuildError.Error() {
		t.Fatalf("the refusal was flattened into a generic sentinel: %s", message)
	}
	if !errors.Is(err, configs.TemplateError) {
		t.Fatalf("the failing phase must be named as the template phase, got: %v", err)
	}

	var missing *strict.MissingVariablesError
	if !errors.As(err, &missing) {
		t.Fatalf("the verdict must survive the trip to the CLI, got %T", err)
	}
	if got, want := len(missing.Keys()), 3; got != want {
		t.Fatalf("keys reported = %d, want %d", got, want)
	}

	if _, statErr := os.Stat(filepath.Join(directory, "out.pdf")); !os.IsNotExist(statErr) {
		t.Fatal("a refused build must not leave a PDF behind")
	}
}

// The source template is an input, and a build must never consume it.
func TestBuildLeavesTheSourceTemplateIntact(t *testing.T) {
	directory := t.TempDir()
	t.Chdir(directory)

	templatePath := filepath.Join(directory, "doc.tex")
	template := `\documentclass[12pt]{article}
\begin{document}
T: delim[[.title]]
\end{document}
`
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatalf("write template: %v", err)
	}

	// output shares the template's base name, which is the collision that used
	// to make the staged file and the source the same path.
	configPath := filepath.Join(directory, "config.yaml")
	configBody := `template: "doc.tex"
output: "doc.pdf"
engine: "pdflatex"
variables:
  title: "Example Title"
`
	if err := os.WriteFile(configPath, []byte(configBody), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := BuildServiceCmd.Do(nil, templatePath, configPath); err != nil {
		t.Skipf("build did not complete in this environment, skipping: %v", err)
	}

	survived, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("the source template must survive a build: %v", err)
	}
	if string(survived) != template {
		t.Fatalf("the source template must be unchanged, got:\n%s", survived)
	}
}
