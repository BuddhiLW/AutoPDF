// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package latex

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	application "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// LaTeXCompilerAdapter wraps the existing LaTeX compiler
// Now DIP-compliant: depends on ports, not low-level libraries
type LaTeXCompilerAdapter struct {
	config     *config.Config
	fileSystem application.FileSystem
	executor   application.CommandExecutor
	workingDir string // NEW: Configurable working directory
}

// NewLaTeXCompilerAdapter creates a new LaTeX compiler adapter
func NewLaTeXCompilerAdapter(
	cfg *config.Config,
	fileSystem application.FileSystem,
	executor application.CommandExecutor,
) *LaTeXCompilerAdapter {
	return &LaTeXCompilerAdapter{
		config:     cfg,
		fileSystem: fileSystem,
		executor:   executor,
		workingDir: "/tmp/autopdf", // Default working directory
	}
}

// NewLaTeXCompilerAdapterWithWorkingDir creates a new LaTeX compiler adapter with custom working directory
func NewLaTeXCompilerAdapterWithWorkingDir(
	cfg *config.Config,
	fileSystem application.FileSystem,
	executor application.CommandExecutor,
	workingDir string,
) *LaTeXCompilerAdapter {
	return &LaTeXCompilerAdapter{
		config:     cfg,
		fileSystem: fileSystem,
		executor:   executor,
		workingDir: workingDir, // Use provided working directory
	}
}

// Compile compiles LaTeX content to PDF
func (lca *LaTeXCompilerAdapter) Compile(ctx context.Context, content string, engine string, outputPath string, debugEnabled bool) (string, error) {
	if content == "" {
		return "", errors.New("no LaTeX content provided")
	}

	// Determine the engine to use
	if engine == "" {
		engine = "pdflatex" // Default engine
	}

	// Verify that the engine is installed
	if _, err := lca.executor.Execute(ctx, application.NewCommand("which", []string{engine}, "")); err != nil {
		return "", fmt.Errorf("LaTeX engine not found: %s", engine)
	}

	// Use configurable working directory
	workingDir := lca.workingDir

	// Ensure working directory exists
	if err := lca.fileSystem.MkdirAll(ctx, workingDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create working directory: %w", err)
	}

	// Generate a concrete file name
	var concreteFileName string
	if debugEnabled {
		// In debug mode, create a persistent concrete file with a memorable name
		concreteFileName = "autopdf-concrete.tex"
		if outputPath != "" {
			baseName := filepath.Base(outputPath)
			concreteFileName = "autopdf-concrete-" + strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".tex"
		}
	} else {
		// Normal mode: use temporary file
		concreteFileName = "autopdf_temp.tex"
		if outputPath != "" {
			baseName := filepath.Base(outputPath)
			concreteFileName = "autopdf_" + strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".tex"
		}
	}
	concreteFile := filepath.Join(workingDir, concreteFileName)

	// Write the content to the concrete file
	if err := lca.fileSystem.WriteFile(ctx, concreteFile, []byte(content), 0644); err != nil {
		return "", err
	}

	// Only clean up temp file if not in debug mode
	if !debugEnabled {
		defer func() {
			if err := lca.fileSystem.Remove(ctx, concreteFile); err != nil {
				// Log cleanup error but don't fail
			}
		}()
	}

	// Determine output PDF path
	var pdfPath string
	if outputPath != "" {
		pdfPath = outputPath
		if !strings.HasSuffix(pdfPath, ".pdf") {
			pdfPath = pdfPath + ".pdf"
		}
	} else {
		// Default output path
		baseName := strings.TrimSuffix(filepath.Base(concreteFile), ".tex")
		pdfPath = filepath.Join(workingDir, baseName+".pdf")
	}

	// Create output directory if needed
	outputDir := filepath.Dir(pdfPath)
	if err := lca.fileSystem.MkdirAll(ctx, outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get the base name for the LaTeX job
	baseName := strings.TrimSuffix(filepath.Base(pdfPath), ".pdf")

	// Create command to run LaTeX
	var cmdStr string
	if outputDir == "." {
		cmdStr = fmt.Sprintf("%s -interaction=nonstopmode -jobname=%s %s", engine, baseName, concreteFile)
	} else {
		cmdStr = fmt.Sprintf("%s -interaction=nonstopmode -jobname=%s -output-directory=%s %s", engine, baseName, outputDir, concreteFile)
	}

	cmd := application.NewCommand("sh", []string{"-c", cmdStr}, workingDir).
		WithTimeout(5 * time.Minute)

	// Run the LaTeX command
	result, err := lca.executor.Execute(ctx, cmd)
	if err != nil {
		// Check if PDF was created despite the error
		if _, statErr := lca.fileSystem.Stat(ctx, pdfPath); statErr != nil {
			// Include LaTeX's actual error output in the error message
			errorDetails := fmt.Sprintf(
				"LaTeX compilation failed:\nCommand: %s\nWorking Dir: %s\nStderr:\n%s\nStdout:\n%s",
				cmdStr, workingDir, result.Stderr, result.Stdout,
			)
			return "", fmt.Errorf("%s: %w", errorDetails, err)
		}
		// PDF was created, so continue
	}

	// Check if output PDF exists and has content
	fileInfo, statErr := lca.fileSystem.Stat(ctx, pdfPath)
	if statErr != nil {
		if lca.fileSystem.IsNotExist(statErr) {
			return "", errors.New("PDF output file was not created")
		}
		return "", fmt.Errorf("error checking output file: %w", statErr)
	}

	// Check if file is empty
	if fileInfo.Size() == 0 {
		return "", errors.New("PDF output file was created but is empty")
	}

	return pdfPath, nil
}
