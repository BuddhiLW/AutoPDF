// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package latex

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// LaTeXCompilerAdapter wraps the existing LaTeX compiler
type LaTeXCompilerAdapter struct {
	config *config.Config
}

// NewLaTeXCompilerAdapter creates a new LaTeX compiler adapter
func NewLaTeXCompilerAdapter(cfg *config.Config) *LaTeXCompilerAdapter {
	return &LaTeXCompilerAdapter{
		config: cfg,
	}
}

// Compile compiles LaTeX content to PDF
func (lca *LaTeXCompilerAdapter) Compile(ctx context.Context, content string, engine string, outputPath string) (string, error) {
	if content == "" {
		return "", errors.New("no LaTeX content provided")
	}

	// Determine the engine to use
	if engine == "" {
		engine = "pdflatex" // Default engine
	}

	// Verify that the engine is installed
	if _, err := exec.LookPath(engine); err != nil {
		return "", fmt.Errorf("LaTeX engine not found: %s", engine)
	}

	// Create a temporary file for the LaTeX content
	tempDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Generate a temp file name
	tempFileName := "autopdf_temp.tex"
	if outputPath != "" {
		baseName := filepath.Base(outputPath)
		tempFileName = "autopdf_" + strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".tex"
	}
	tempFile := filepath.Join(tempDir, tempFileName)

	// Write the content to the temp file
	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		return "", err
	}

	// Ensure temp file is cleaned up
	defer os.Remove(tempFile)

	// Determine output PDF path
	var pdfPath string
	if outputPath != "" {
		pdfPath = outputPath
		if !strings.HasSuffix(pdfPath, ".pdf") {
			pdfPath = pdfPath + ".pdf"
		}
	} else {
		// Default output path
		baseName := strings.TrimSuffix(filepath.Base(tempFile), ".tex")
		pdfPath = filepath.Join(tempDir, baseName+".pdf")
	}

	// Create output directory if needed
	outputDir := filepath.Dir(pdfPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get the base name for the LaTeX job
	baseName := strings.TrimSuffix(filepath.Base(pdfPath), ".pdf")

	// Create command to run LaTeX
	var cmd *exec.Cmd
	if outputDir == "." {
		cmdStr := fmt.Sprintf("%s -interaction=nonstopmode -jobname=%s %s", engine, baseName, tempFile)
		cmd = exec.Command("sh", "-c", cmdStr)
	} else {
		cmdStr := fmt.Sprintf("%s -interaction=nonstopmode -jobname=%s -output-directory=%s %s", engine, baseName, outputDir, tempFile)
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	// Run the LaTeX command
	if err := cmd.Run(); err != nil {
		// Check if PDF was created despite the error
		if _, statErr := os.Stat(pdfPath); statErr != nil {
			return "", fmt.Errorf("LaTeX compilation failed: %w", err)
		}
		// PDF was created, so continue
	}

	// Check if output PDF exists and has content
	fileInfo, statErr := os.Stat(pdfPath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
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
