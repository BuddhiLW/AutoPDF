// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/BuddhiLW/AutoPDF/internal/tex"
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

	// Create a config with the specified engine
	cfg := &config.Config{
		Engine: config.Engine(engine),
		Output: config.Output(outputPath),
	}

	// Create the compiler
	compiler := tex.NewCompiler(cfg)

	// Compile the LaTeX file
	pdfPath, err := compiler.Compile(tempFile)
	if err != nil {
		return "", err
	}

	// If output path is specified and different from compiled path, move the PDF
	if outputPath != "" && pdfPath != outputPath {
		// Ensure output path has .pdf extension
		if !strings.HasSuffix(outputPath, ".pdf") {
			outputPath = outputPath + ".pdf"
		}

		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return "", err
		}

		pdfData, err := os.ReadFile(pdfPath)
		if err != nil {
			return "", err
		}

		if err := os.WriteFile(outputPath, pdfData, 0644); err != nil {
			return "", err
		}

		// Clean up the original compiled PDF if it's different
		if pdfPath != outputPath {
			os.Remove(pdfPath)
		}

		pdfPath = outputPath
	}

	return pdfPath, nil
}
