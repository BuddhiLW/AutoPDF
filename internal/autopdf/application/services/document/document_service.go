// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package document

import (
	"context"
	"errors"

	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
	apperrors "github.com/BuddhiLW/AutoPDF/v2/pkg/errors"
)

// DocumentService orchestrates the document generation workflow
// It depends on ports (interfaces) for flexibility and testability
type DocumentService struct {
	TemplateProcessor ports.TemplateProcessor
	Compiler          ports.DocumentCompiler // used when Compilers is nil
	Compilers         CompilerResolver       // picks the compiler per request (engine, template); wins over Compiler
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
	DoClean      bool
	DebugEnabled bool                 // Enable debug mode for persistent concrete files
	Backend      ports.BackendOptions // Backend-specific options (e.g. ports.LaTeXOptions); nil = backend defaults
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

	// Step 3: Compile the rendered document to PDF
	// Note: Asset symlinks should be set up by the calling strategy (e.g., AutoPDFGenerationStrategy)
	// Use the working directory from BuildRequest if provided, otherwise derive from output path
	outputPath := req.OutputPath
	if outputPath != "" && s.PathOps.Ext(outputPath) == "" {
		outputPath += ".pdf"
	}

	workingDir := req.WorkingDir
	if workingDir == "" {
		workingDir = s.PathOps.Dir(outputPath)
	}

	// Extract jobname from output path (base filename without extension)
	// This ensures PDF/JPEG use the custom tag naming instead of hardcoded "document"
	jobName := "document"
	if outputPath != "" {
		outputBaseName := s.PathOps.Base(outputPath)
		ext := s.PathOps.Ext(outputBaseName)
		// Remove extension from base filename to get jobname
		jobName = outputBaseName
		if ext != "" && len(jobName) > len(ext) {
			jobName = jobName[:len(jobName)-len(ext)]
		}
		if jobName == "" {
			jobName = "document"
		}
	}

	compileOptions := ports.NewCompileOptions(req.Engine, outputPath, workingDir).
		WithDebug(req.DebugEnabled).
		WithBackendOptions(req.Backend).
		WithJobName(jobName) // Set jobname from output path

	compiler, err := s.compilerFor(req)
	if err != nil {
		return BuildResult{
			Success: false,
			Error:   s.ErrorFactory.LaTeXCompilationFailed(outputPath, err),
		}, err
	}

	pdfPath, err := compiler.Compile(ctx, processedContent, compileOptions)
	if err != nil {
		return BuildResult{
			Success: false,
			Error:   s.ErrorFactory.LaTeXCompilationFailed(outputPath, err),
		}, err
	}

	result := BuildResult{
		PDFPath: pdfPath,
		Success: true,
	}

	// Step 5: Optionally convert PDF to images
	if req.Conversion.Enabled {
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

// CompilerResolver chooses the DocumentCompiler for a request's engine and
// template, built for the request's backend options.
type CompilerResolver interface {
	CompilerFor(engine, templatePath string, opts ports.BackendOptions) (ports.DocumentCompiler, error)
}

// CompilerResolverFunc adapts a function to CompilerResolver.
type CompilerResolverFunc func(engine, templatePath string, opts ports.BackendOptions) (ports.DocumentCompiler, error)

// CompilerFor implements CompilerResolver.
func (f CompilerResolverFunc) CompilerFor(engine, templatePath string, opts ports.BackendOptions) (ports.DocumentCompiler, error) {
	return f(engine, templatePath, opts)
}

// ErrNoCompiler is returned when a DocumentService has neither Compilers nor Compiler.
var ErrNoCompiler = errors.New("document service has no compiler")

func (s *DocumentService) compilerFor(req BuildRequest) (ports.DocumentCompiler, error) {
	if s.Compilers != nil {
		return s.Compilers.CompilerFor(req.Engine, req.TemplatePath, req.Backend)
	}
	if s.Compiler == nil {
		return nil, ErrNoCompiler
	}
	return s.Compiler, nil
}

// ConvertDocument converts an existing PDF to images
func (s *DocumentService) ConvertDocument(ctx context.Context, pdfPath string, formats []string) ([]string, error) {
	return s.Converter.ConvertToImages(ctx, pdfPath, formats)
}

// CleanDocument removes auxiliary files for a given PDF
func (s *DocumentService) CleanDocument(ctx context.Context, pdfPath string) error {
	return s.Cleaner.Clean(ctx, pdfPath)
}
