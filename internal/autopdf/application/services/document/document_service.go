// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package document

import (
	"context"

	ports "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	apperrors "github.com/BuddhiLW/AutoPDF/pkg/errors"
)

// DocumentService orchestrates the document generation workflow
// It depends on ports (interfaces) for flexibility and testability
type DocumentService struct {
	TemplateProcessor ports.TemplateProcessor
	LaTeXCompiler     ports.LaTeXCompiler
	Converter         ports.Converter
	Cleaner           ports.Cleaner
	PathOps           ports.PathOperations
	FileSystem        ports.FileSystem
	ErrorFactory      *apperrors.DomainErrorFactory
}

// BuildRequest encapsulates all parameters for building a document
type BuildRequest struct {
	TemplatePath string
	ConfigPath   string
	Variables    *config.Variables // Use complex variables from pkg/
	Engine       string
	OutputPath   string
	WorkingDir   string // Working directory for LaTeX compilation (where assets are symlinked)
	DoConvert    bool
	DoClean      bool
	DebugEnabled bool // Enable debug mode for persistent concrete files
	Passes       int  // Number of compilation passes
	UseLatexmk   bool // Whether to use latexmk
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
			Error:   s.ErrorFactory.TemplateProcessingFailed(req.TemplatePath, err),
		}, err
	}

	// Step 3: Compile LaTeX to PDF
	// Note: Asset symlinks should be set up by the calling strategy (e.g., AutoPDFGenerationStrategy)
	// Use the working directory from BuildRequest if provided, otherwise derive from output path
	workingDir := req.WorkingDir
	if workingDir == "" {
		workingDir = s.PathOps.Dir(req.OutputPath)
	}

	// Extract jobname from output path (base filename without extension)
	// This ensures PDF/JPEG use the custom tag naming instead of hardcoded "document"
	outputBaseName := s.PathOps.Base(req.OutputPath)
	ext := s.PathOps.Ext(outputBaseName)
	// Remove extension from base filename to get jobname
	jobName := outputBaseName
	if ext != "" && len(jobName) > len(ext) {
		jobName = jobName[:len(jobName)-len(ext)]
	}
	// Fallback to "document" if extraction fails
	if jobName == "" {
		jobName = "document"
	}

	compileOptions := ports.NewCompileOptions(req.Engine, req.OutputPath, workingDir).
		WithDebug(req.DebugEnabled).
		WithPasses(req.Passes).
		WithLatexmk(req.UseLatexmk).
		WithJobName(jobName) // Set jobname from output path

	pdfPath, err := s.LaTeXCompiler.Compile(ctx, processedContent, compileOptions)
	if err != nil {
		return BuildResult{
			Success: false,
			Error:   s.ErrorFactory.LaTeXCompilationFailed(req.OutputPath, err),
		}, err
	}

	result := BuildResult{
		PDFPath: pdfPath,
		Success: true,
	}

	// Step 5: Optionally convert PDF to images
	if req.DoConvert && req.Conversion.Enabled {
		imagePaths, err := s.Converter.ConvertToImages(ctx, pdfPath, req.Conversion.Formats)
		if err != nil {
			// Log warning but don't fail the build
			result.Error = s.ErrorFactory.PDFConversionFailed(pdfPath, err)
		} else {
			result.ImagePaths = imagePaths
		}
	}

	// Step 6: Optionally clean auxiliary files
	if req.DoClean {
		if err := s.Cleaner.Clean(ctx, pdfPath); err != nil {
			// Log warning but don't fail the build
			result.Error = s.ErrorFactory.CleanupFailed(pdfPath, err)
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
