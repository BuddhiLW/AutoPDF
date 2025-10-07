// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// ExternalPDFServiceAdapter implements domain.PDFGenerationService
// This adapter bridges the API layer with the internal application layer
type ExternalPDFServiceAdapter struct {
	config *config.Config
}

// NewExternalPDFServiceAdapter creates a new external PDF service adapter
func NewExternalPDFServiceAdapter(cfg *config.Config) *ExternalPDFServiceAdapter {
	return &ExternalPDFServiceAdapter{
		config: cfg,
	}
}

// Generate generates a PDF using the internal application layer
func (epsa *ExternalPDFServiceAdapter) Generate(ctx context.Context, req domain.PDFGenerationRequest) (domain.PDFGenerationResult, error) {
	// Convert domain request to internal application request
	// Convert simple variables to complex variables
	complexVars := config.NewVariables()
	for key, value := range req.Variables {
		complexVars.SetString(key, fmt.Sprintf("%v", value))
	}

	appReq := application.BuildRequest{
		TemplatePath: req.TemplatePath,
		ConfigPath:   "", // Not needed for API usage
		Variables:    complexVars,
		Engine:       req.Engine,
		OutputPath:   req.OutputPath,
		DoConvert:    req.Options.DoConvert,
		DoClean:      req.Options.DoClean,
		Conversion: application.ConversionSettings{
			Enabled: req.Options.Conversion.Enabled,
			Formats: req.Options.Conversion.Formats,
		},
	}

	// Create internal application service
	docService := epsa.createDocumentService()

	// Generate PDF using internal service
	result, err := docService.Build(ctx, appReq)
	if err != nil {
		return domain.PDFGenerationResult{
			Success: false,
			Error: domain.PDFGenerationError{
				Code:    domain.ErrCodePDFGenerationFailed,
				Message: "PDF generation failed",
				Details: map[string]interface{}{"error": err.Error()},
			},
		}, err
	}

	// Convert internal result to domain result
	domainResult := domain.PDFGenerationResult{
		PDFPath:    result.PDFPath,
		ImagePaths: result.ImagePaths,
		Success:    result.Success,
		Error:      result.Error,
		Metadata: domain.PDFMetadata{
			FileSize:    epsa.getFileSize(result.PDFPath),
			PageCount:   1, // Would need proper PDF parsing
			GeneratedAt: time.Now(),
			Engine:      req.Engine,
			Template:    req.TemplatePath,
		},
	}

	return domainResult, nil
}

// ValidateRequest validates a PDF generation request
func (epsa *ExternalPDFServiceAdapter) ValidateRequest(req domain.PDFGenerationRequest) error {
	if req.TemplatePath == "" {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodeTemplateNotFound,
			Message: "Template path is required",
		}
	}

	if req.Engine == "" {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodeEngineNotFound,
			Message: "LaTeX engine is required",
		}
	}

	if req.OutputPath == "" {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodeOutputPathInvalid,
			Message: "Output path is required",
		}
	}

	// Check if template exists
	if _, err := os.Stat(req.TemplatePath); os.IsNotExist(err) {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodeTemplateNotFound,
			Message: "Template file does not exist",
			Details: map[string]interface{}{"template": req.TemplatePath},
		}
	}

	// Check if output directory exists or can be created
	outputDir := filepath.Dir(req.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodeOutputPathInvalid,
			Message: "Cannot create output directory",
			Details: map[string]interface{}{"output_dir": outputDir, "error": err.Error()},
		}
	}

	return nil
}

// GetSupportedEngines returns supported LaTeX engines
func (epsa *ExternalPDFServiceAdapter) GetSupportedEngines() []string {
	return []string{
		"pdflatex",
		"xelatex",
		"lualatex",
		"latex",
	}
}

// GetSupportedFormats returns supported output formats
func (epsa *ExternalPDFServiceAdapter) GetSupportedFormats() []string {
	return []string{
		"pdf",
		"png",
		"jpeg",
		"jpg",
		"gif",
		"svg",
	}
}

// createDocumentService creates the internal document service
func (epsa *ExternalPDFServiceAdapter) createDocumentService() *application.DocumentService {
	// Create adapters for internal application layer
	templateAdapter := adapters.NewTemplateProcessorAdapter(epsa.config)
	latexAdapter := adapters.NewLaTeXCompilerAdapter(epsa.config)
	converterAdapter := adapters.NewConverterAdapter(epsa.config)
	cleanerAdapter := adapters.NewCleanerAdapter()

	// Create document service
	return &application.DocumentService{
		TemplateProcessor: templateAdapter,
		LaTeXCompiler:     latexAdapter,
		Converter:         converterAdapter,
		Cleaner:           cleanerAdapter,
	}
}

// getFileSize gets the file size of a PDF
func (epsa *ExternalPDFServiceAdapter) getFileSize(pdfPath string) int64 {
	if pdfPath == "" {
		return 0
	}

	fileInfo, err := os.Stat(pdfPath)
	if err != nil {
		return 0
	}

	return fileInfo.Size()
}

// CreateConfigFromRequest creates a config from a PDF generation request
func (epsa *ExternalPDFServiceAdapter) CreateConfigFromRequest(req domain.PDFGenerationRequest) *config.Config {
	cfg := &config.Config{
		Template: config.Template(req.TemplatePath),
		Output:   config.Output(req.OutputPath),
		Engine:   config.Engine(req.Engine),
		Conversion: config.Conversion{
			Enabled: req.Options.Conversion.Enabled,
			Formats: req.Options.Conversion.Formats,
		},
		Variables: *config.NewVariables(),
	}

	// Set variables from request
	for key, value := range req.Variables {
		cfg.Variables.SetString(key, fmt.Sprintf("%v", value))
	}

	return cfg
}
