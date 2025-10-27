// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/ports"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/valueobjects"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/infrastructure/adapters"
)

// LaTeXCommandBuilder builds LaTeX command strings
type LaTeXCommandBuilder struct {
	fileSystem ports.FileSystem
}

// NewLaTeXCommandBuilder creates a new LaTeX command builder
func NewLaTeXCommandBuilder(fileSystem ports.FileSystem) *LaTeXCommandBuilder {
	return &LaTeXCommandBuilder{
		fileSystem: fileSystem,
	}
}

// Build creates a LaTeX command for the given context and temp file
// Strategy Pattern: Selects format-aware or legacy compilation based on context
func (b *LaTeXCommandBuilder) Build(ctx valueobjects.CompilationContext, tempFile string) ports.Command {
	// Determine output PDF path
	pdfPath := b.determineOutputPath(ctx, tempFile)

	// Create output directory if needed
	outputDir := filepath.Dir(pdfPath)
	b.fileSystem.MkdirAll(outputDir, valueobjects.GetDirectoryPermission())

	// Get the base name for the LaTeX job
	baseName := strings.TrimSuffix(filepath.Base(pdfPath), ".pdf")

	// Build command using appropriate strategy
	var cmd ports.Command
	if ctx.HasFormatFile() {
		// Format-aware compilation (4-12x faster)
		cmd = b.buildFormatCommand(ctx, tempFile, baseName, outputDir)
	} else {
		// Legacy compilation (backward compatible)
		cmd = b.buildLegacyCommand(ctx, tempFile, baseName, outputDir)
	}

	// Set working directory if specified
	if workDir := ctx.WorkingDirectory(); workDir != "" {
		adapters.SetWorkingDirectory(cmd, workDir)
	}

	return cmd
}

// buildFormatCommand builds a LaTeX command with format file support
// Compose Pattern: Extends base command with format-specific flags
func (b *LaTeXCommandBuilder) buildFormatCommand(
	ctx valueobjects.CompilationContext,
	tempFile, baseName, outputDir string,
) ports.Command {
	args := []string{
		fmt.Sprintf("-fmt=%s", ctx.FormatFile()),
		"-interaction=nonstopmode",
		"-halt-on-error",
		"-file-line-error",
		"-no-shell-escape",
		fmt.Sprintf("-jobname=%s", baseName),
	}

	// Add output directory if not current directory and no working directory specified
	// When working directory is set, LaTeX will write to that directory by default
	if outputDir != "." && ctx.WorkingDirectory() == "" {
		args = append(args, fmt.Sprintf("-output-directory=%s", outputDir))
	}

	// Add input file (must be absolute path)
	args = append(args, tempFile)

	// Debug logging
	fmt.Printf("DEBUG LaTeX command: engine=%s, inputTex=%s, outputDir=%s, jobName=%s\n",
		ctx.Engine(), tempFile, outputDir, baseName)

	return adapters.NewExecCommandFromArgs(ctx.Engine(), args)
}

// buildLegacyCommand builds a standard LaTeX command without format file
// Preserves original behavior for backward compatibility
func (b *LaTeXCommandBuilder) buildLegacyCommand(
	ctx valueobjects.CompilationContext,
	tempFile, baseName, outputDir string,
) ports.Command {
	args := []string{
		"-interaction=nonstopmode",
		"-halt-on-error",
		"-file-line-error",
		"-no-shell-escape",
		fmt.Sprintf("-jobname=%s", baseName),
	}

	// Add output directory if not current directory and no working directory specified
	// When working directory is set, LaTeX will write to that directory by default
	if outputDir != "." && ctx.WorkingDirectory() == "" {
		args = append(args, fmt.Sprintf("-output-directory=%s", outputDir))
	}

	// Add input file (must be absolute path)
	args = append(args, tempFile)

	// Debug logging
	fmt.Printf("DEBUG LaTeX command: engine=%s, inputTex=%s, outputDir=%s, jobName=%s\n",
		ctx.Engine(), tempFile, outputDir, baseName)

	return adapters.NewExecCommandFromArgs(ctx.Engine(), args)
}

// determineOutputPath determines the output PDF path
func (b *LaTeXCommandBuilder) determineOutputPath(ctx valueobjects.CompilationContext, tempFile string) string {
	if ctx.OutputPath() != "" {
		pdfPath := ctx.OutputPath()
		if !strings.HasSuffix(pdfPath, ".pdf") {
			pdfPath = pdfPath + ".pdf"
		}
		return pdfPath
	}

	// Default output path
	baseName := strings.TrimSuffix(filepath.Base(tempFile), ".tex")
	tempDir, _ := b.fileSystem.Getwd()
	return filepath.Join(tempDir, baseName+".pdf")
}
