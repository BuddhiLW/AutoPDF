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

// Compile compiles LaTeX content to PDF using the new CompileOptions interface
func (lca *LaTeXCompilerAdapter) Compile(ctx context.Context, content string, opts application.CompileOptions) (string, error) {
	if content == "" {
		return "", errors.New("no LaTeX content provided")
	}

	// Use working directory from options if provided, otherwise use adapter's default
	workingDir := opts.WorkingDir
	if workingDir == "" {
		workingDir = lca.workingDir
	}

	// Ensure working directory exists
	if err := lca.fileSystem.MkdirAll(ctx, workingDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create working directory: %w", err)
	}

	// Verify that the engine is installed
	if _, err := lca.executor.Execute(ctx, application.NewCommand("which", []string{opts.Engine}, "")); err != nil {
		return "", fmt.Errorf("LaTeX engine not found: %s", opts.Engine)
	}

	// Generate concrete file name using job name
	concreteFileName := fmt.Sprintf("%s.tex", opts.JobName)
	concreteFile := filepath.Join(workingDir, concreteFileName)

	// Write the content to the concrete file
	if err := lca.fileSystem.WriteFile(ctx, concreteFile, []byte(content), 0644); err != nil {
		return "", err
	}

	// Only clean up temp file if not in debug mode
	if !opts.Debug {
		defer func() {
			if err := lca.fileSystem.Remove(ctx, concreteFile); err != nil {
				// Log cleanup error but don't fail
			}
		}()
	}

	// Determine output PDF path
	pdfPath := opts.OutputPath
	if pdfPath == "" {
		pdfPath = filepath.Join(workingDir, fmt.Sprintf("%s.pdf", opts.JobName))
	}

	// Create output directory if needed
	outputDir := filepath.Dir(pdfPath)
	if err := lca.fileSystem.MkdirAll(ctx, outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get the base name for the LaTeX job
	baseName := strings.TrimSuffix(filepath.Base(pdfPath), ".pdf")

	// Run multiple passes if requested
	for pass := 1; pass <= opts.Passes; pass++ {
		// Create command to run LaTeX
		var cmdStr string
		if outputDir == "." {
			cmdStr = fmt.Sprintf("%s -interaction=nonstopmode -jobname=%s %s", opts.Engine, baseName, concreteFile)
		} else {
			cmdStr = fmt.Sprintf("%s -interaction=nonstopmode -jobname=%s -output-directory=%s %s", opts.Engine, baseName, outputDir, concreteFile)
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
					"LaTeX compilation failed (pass %d/%d):\nCommand: %s\nWorking Dir: %s\nStderr:\n%s\nStdout:\n%s",
					pass, opts.Passes, cmdStr, workingDir, result.Stderr, result.Stdout,
				)
				return "", fmt.Errorf("%s: %w", errorDetails, err)
			}
			// PDF was created, so continue
		}

		// Check if output PDF exists and has content
		fileInfo, statErr := lca.fileSystem.Stat(ctx, pdfPath)
		if statErr != nil {
			if lca.fileSystem.IsNotExist(statErr) {
				return "", fmt.Errorf("PDF output file was not created (pass %d/%d)", pass, opts.Passes)
			}
			return "", fmt.Errorf("error checking output file (pass %d/%d): %w", pass, opts.Passes, statErr)
		}

		// Check if file is empty
		if fileInfo.Size() == 0 {
			return "", fmt.Errorf("PDF output file was created but is empty (pass %d/%d)", pass, opts.Passes)
		}
	}

	return pdfPath, nil
}

// CompileLegacy provides backward compatibility with the old interface
func (lca *LaTeXCompilerAdapter) CompileLegacy(ctx context.Context, content string, engine string, outputPath string, debugEnabled bool) (string, error) {
	opts := application.NewCompileOptions(engine, outputPath, lca.workingDir).
		WithDebug(debugEnabled)

	return lca.Compile(ctx, content, opts)
}
