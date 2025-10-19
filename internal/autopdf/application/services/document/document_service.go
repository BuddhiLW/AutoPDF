// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package document

import (
	"context"
	"fmt"

	ports "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// DocumentService orchestrates the document generation workflow
// It depends on ports (interfaces) for flexibility and testability
type DocumentService struct {
	TemplateProcessor ports.TemplateProcessor
	LaTeXCompiler     ports.LaTeXCompiler
	Converter         ports.Converter
	Cleaner           ports.Cleaner
}

// BuildRequest encapsulates all parameters for building a document
type BuildRequest struct {
	TemplatePath string
	ConfigPath   string
	Variables    *config.Variables // Use complex variables from pkg/
	Engine       string
	OutputPath   string
	DoConvert    bool
	DoClean      bool
	DebugEnabled bool // Enable debug mode for persistent concrete files
	Conversion   ConversionSettings
}

// ConversionSettings holds conversion options
type ConversionSettings struct {
	Enabled bool
	Formats []string
}

// BuildResult encapsulates the results of building a document
type BuildResult struct {
	PDFPath    string
	ImagePaths []string
	Success    bool
	Error      error
}

// Build orchestrates the entire document generation workflow:
// 1. Process template with variables
// 2. Compile LaTeX to PDF
// 3. Optionally convert PDF to images
// 4. Optionally clean auxiliary files
func (s *DocumentService) Build(ctx context.Context, req BuildRequest) (BuildResult, error) {
	// Step 1: Convert complex variables to simple map for template processing
	simpleVariables := make(map[string]string)
	if req.Variables != nil {
		// Flatten complex variables to simple key-value pairs
		simpleVariables = req.Variables.Flatten()
	}

	// Step 2: Process template
	processedContent, err := s.TemplateProcessor.Process(ctx, req.TemplatePath, simpleVariables)
	if err != nil {
		return BuildResult{
			Success: false,
			Error:   fmt.Errorf("template processing failed: %w", err),
		}, err
	}

	// Step 3: Compile LaTeX to PDF
	pdfPath, err := s.LaTeXCompiler.Compile(ctx, processedContent, req.Engine, req.OutputPath, req.DebugEnabled)
	if err != nil {
		return BuildResult{
			Success: false,
			Error:   fmt.Errorf("LaTeX compilation failed: %w", err),
		}, err
	}

	result := BuildResult{
		PDFPath: pdfPath,
		Success: true,
	}

	// Step 4: Optionally convert PDF to images
	if req.DoConvert && req.Conversion.Enabled {
		imagePaths, err := s.Converter.ConvertToImages(ctx, pdfPath, req.Conversion.Formats)
		if err != nil {
			// Log warning but don't fail the build
			result.Error = fmt.Errorf("PDF conversion failed: %w", err)
		} else {
			result.ImagePaths = imagePaths
		}
	}

	// Step 5: Optionally clean auxiliary files
	if req.DoClean {
		if err := s.Cleaner.Clean(ctx, pdfPath); err != nil {
			// Log warning but don't fail the build
			result.Error = fmt.Errorf("cleanup failed: %w", err)
		}
	}

	return result, nil
}

// ConvertDocument converts an existing PDF to images
func (s *DocumentService) ConvertDocument(ctx context.Context, pdfPath string, formats []string) ([]string, error) {
	return s.Converter.ConvertToImages(ctx, pdfPath, formats)
}

// CleanDocument removes auxiliary files for a given PDF
func (s *DocumentService) CleanDocument(ctx context.Context, pdfPath string) error {
	return s.Cleaner.Clean(ctx, pdfPath)
}
