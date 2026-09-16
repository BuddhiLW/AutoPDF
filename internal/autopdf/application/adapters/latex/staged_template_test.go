// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package latex_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/latex"
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/infrastructure/adapters"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
)

// pdfWritingExecutor stands in for the LaTeX engine: it answers `which` and,
// for a compile, drops a non-empty PDF where the adapter expects one.
type pdfWritingExecutor struct {
	pdfPath string
	sawArgs [][]string
}

func (e *pdfWritingExecutor) Execute(ctx context.Context, cmd ports.Command) (ports.CommandResult, error) {
	e.sawArgs = append(e.sawArgs, cmd.Args)
	if cmd.Executable == "which" {
		return ports.NewCommandResult("/usr/bin/pdflatex", "", 0, 0), nil
	}
	if err := os.WriteFile(e.pdfPath, []byte("%PDF-1.4 stub"), 0o600); err != nil {
		return ports.CommandResult{}, err
	}
	return ports.NewCommandResult("", "", 0, 0), nil
}

// A job name equal to the source template's base name used to make the staged
// file and the template the same path, so the compile overwrote the input and
// then deleted it.
func TestCompileNeverWritesOverTheSourceTemplate(t *testing.T) {
	workingDir := t.TempDir()
	templatePath := filepath.Join(workingDir, "doc.tex")
	original := []byte(`\documentclass{article}\begin{document}delim[[.a]]\end{document}`)
	if err := os.WriteFile(templatePath, original, 0o600); err != nil {
		t.Fatalf("write template: %v", err)
	}

	outputPath := filepath.Join(workingDir, "doc.pdf")
	executor := &pdfWritingExecutor{pdfPath: outputPath}
	adapter := latex.NewLaTeXCompilerAdapterWithWorkingDir(
		&config.Config{}, adapters.NewOSFileSystem(), executor, workingDir)

	opts := ports.NewCompileOptions("pdflatex", outputPath, workingDir).
		WithJobName("doc")

	if _, err := adapter.Compile(context.Background(), "rendered content", opts); err != nil {
		t.Fatalf("compile: %v", err)
	}

	survived, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("the source template must survive a compile: %v", err)
	}
	if string(survived) != string(original) {
		t.Fatalf("the source template must be unchanged, got %q", survived)
	}
}

func TestCompileRemovesOnlyItsOwnStagedFile(t *testing.T) {
	workingDir := t.TempDir()
	outputPath := filepath.Join(workingDir, "doc.pdf")
	executor := &pdfWritingExecutor{pdfPath: outputPath}
	adapter := latex.NewLaTeXCompilerAdapterWithWorkingDir(
		&config.Config{}, adapters.NewOSFileSystem(), executor, workingDir)

	opts := ports.NewCompileOptions("pdflatex", outputPath, workingDir).
		WithJobName("doc")

	if _, err := adapter.Compile(context.Background(), "rendered content", opts); err != nil {
		t.Fatalf("compile: %v", err)
	}

	staged := filepath.Join(workingDir, "doc"+ports.StagedTemplateSuffix)
	if _, err := os.Stat(staged); !os.IsNotExist(err) {
		t.Fatalf("the staged file must be removed after a non-debug compile: %v", err)
	}
}

func TestDebugModeKeepsTheStagedFileUnderItsOwnName(t *testing.T) {
	workingDir := t.TempDir()
	outputPath := filepath.Join(workingDir, "doc.pdf")
	executor := &pdfWritingExecutor{pdfPath: outputPath}
	adapter := latex.NewLaTeXCompilerAdapterWithWorkingDir(
		&config.Config{}, adapters.NewOSFileSystem(), executor, workingDir)

	opts := ports.NewCompileOptions("pdflatex", outputPath, workingDir).
		WithJobName("doc").
		WithDebug(true)

	if _, err := adapter.Compile(context.Background(), "rendered content", opts); err != nil {
		t.Fatalf("compile: %v", err)
	}

	staged := filepath.Join(workingDir, "doc"+ports.StagedTemplateSuffix)
	if _, err := os.Stat(staged); err != nil {
		t.Fatalf("debug mode must keep the staged file: %v", err)
	}
}
