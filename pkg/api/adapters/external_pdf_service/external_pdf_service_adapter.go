// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	autopdfports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/adapters"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/domain"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/domain/generation"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
)

// TODO: This file needs to be updated after internal application layer restructuring
// ExternalPDFServiceAdapter implements domain.PDFGenerationService
// This adapter bridges the API layer with the internal application layer
type ExternalPDFServiceAdapter struct {
	config       *config.Config
	debugEnabled bool                // Store debug from config
	logger       autopdfports.Logger // Store logger for passing to InternalApplicationAdapter
}

// NewExternalPDFServiceAdapter creates a new external PDF service adapter
func NewExternalPDFServiceAdapter(cfg *config.Config, debugEnabled bool) *ExternalPDFServiceAdapter {
	return NewExternalPDFServiceAdapterWithLogger(cfg, debugEnabled, nil)
}

// NewExternalPDFServiceAdapterWithLogger creates a new external PDF service adapter with logger
func NewExternalPDFServiceAdapterWithLogger(cfg *config.Config, debugEnabled bool, logger autopdfports.Logger) *ExternalPDFServiceAdapter {
	return &ExternalPDFServiceAdapter{
		config:       cfg,
		debugEnabled: debugEnabled,
		logger:       logger,
	}
}

// Generate generates a PDF using the internal application layer
func (epsa *ExternalPDFServiceAdapter) Generate(ctx context.Context, req generation.PDFGenerationRequest) (generation.PDFGenerationResult, error) {
	// Merge debug: request overrides config default
	debugEnabled := epsa.debugEnabled
	if req.Options.Debug.Enabled {
		debugEnabled = true
	}

	// Convert domain request to config
	cfg, err := epsa.createConfigFromRequest(ctx, req)
	if err != nil {
		return generation.PDFGenerationResult{Success: false}, err
	}

	// Create internal application adapter with logger
	internalAdapter := adapters.NewInternalApplicationAdapterWithLogger(cfg, epsa.logger)

	// Extract working directory from request options
	workingDir := req.Options.WorkingDir

	// Generate PDF using the internal adapter with working directory
	pdfBytes, paths, err := internalAdapter.GeneratePDFContext(ctx, cfg, config.Template(req.TemplatePath), debugEnabled, workingDir)
	if err != nil {
		return generation.PDFGenerationResult{
			Success: false,
			Error: domain.PDFGenerationError{
				Code:    domain.ErrCodePDFGenerationFailed,
				Message: "PDF generation failed",
				Details: api.NewErrorDetails(api.ErrorCategoryGeneration, api.ErrorSeverityHigh).
					WithError(err),
			},
		}, err
	}

	// Convert result to domain result
	imagePaths := make([]string, 0, len(paths))
	for _, path := range paths {
		imagePaths = append(imagePaths, path)
	}

	return generation.PDFGenerationResult{
		PDFPath:    req.OutputPath,
		ImagePaths: imagePaths,
		Success:    true,
		Metadata: generation.PDFMetadata{
			FileSize:    int64(len(pdfBytes)),
			PageCount:   1, // Would need proper PDF parsing
			GeneratedAt: time.Now(),
			Engine:      req.Engine,
			Template:    req.TemplatePath,
		},
	}, nil
}

// GenerateWithWorkingDir generates a PDF using the internal application layer with custom working directory
func (epsa *ExternalPDFServiceAdapter) GenerateWithWorkingDir(ctx context.Context, req generation.PDFGenerationRequest, workingDir string) (generation.PDFGenerationResult, error) {
	// Merge debug: request overrides config default
	debugEnabled := epsa.debugEnabled
	if req.Options.Debug.Enabled {
		debugEnabled = true
	}

	// Convert domain request to config
	cfg, err := epsa.createConfigFromRequest(ctx, req)
	if err != nil {
		return generation.PDFGenerationResult{Success: false}, err
	}

	// Create internal application adapter with logger
	internalAdapter := adapters.NewInternalApplicationAdapterWithLogger(cfg, epsa.logger)

	// Generate PDF using the internal adapter with custom working directory
	pdfBytes, paths, err := internalAdapter.GeneratePDFContext(ctx, cfg, config.Template(req.TemplatePath), debugEnabled, workingDir)
	if err != nil {
		return generation.PDFGenerationResult{
			Success: false,
			Error: domain.PDFGenerationError{
				Code:    domain.ErrCodePDFGenerationFailed,
				Message: "PDF generation failed",
				Details: api.NewErrorDetails(api.ErrorCategoryGeneration, api.ErrorSeverityHigh).
					WithError(err),
			},
		}, err
	}

	// Convert result to domain result
	imagePaths := make([]string, 0, len(paths))
	for _, path := range paths {
		imagePaths = append(imagePaths, path)
	}

	return generation.PDFGenerationResult{
		PDFPath:    req.OutputPath,
		ImagePaths: imagePaths,
		Success:    true,
		Metadata: generation.PDFMetadata{
			FileSize:    int64(len(pdfBytes)),
			PageCount:   1, // Would need proper PDF parsing
			GeneratedAt: time.Now(),
			Engine:      req.Engine,
			Template:    req.TemplatePath,
		},
	}, nil
}

// ValidateRequest validates a PDF generation request
func (epsa *ExternalPDFServiceAdapter) ValidateRequest(req generation.PDFGenerationRequest) error {
	if req.TemplatePath == "" {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodeTemplateNotFound,
			Message: api.ErrTemplatePathRequired,
			Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh),
		}
	}

	if req.Engine == "" {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodeEngineNotFound,
			Message: api.ErrEngineRequired,
			Details: api.NewErrorDetails(api.ErrorCategoryGeneration, api.ErrorSeverityHigh),
		}
	}

	if req.OutputPath == "" {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodeOutputPathInvalid,
			Message: api.ErrOutputPathRequired,
			Details: api.NewErrorDetails(api.ErrorCategoryGeneration, api.ErrorSeverityHigh),
		}
	}

	// Check if template exists
	if _, err := os.Stat(req.TemplatePath); os.IsNotExist(err) {
		// Format the error message properly to avoid literal %s
		errorMessage := fmt.Sprintf(api.ErrTemplateFileNotFound, req.TemplatePath)
		return domain.PDFGenerationError{
			Code:    domain.ErrCodeTemplateNotFound,
			Message: errorMessage,
			Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
				WithTemplatePath(req.TemplatePath),
		}
	}

	// Check if output directory exists or can be created
	outputDir := filepath.Dir(req.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		// Format the error message properly to avoid literal %s
		errorMessage := fmt.Sprintf(api.ErrOutputPathInvalid, err.Error())
		return domain.PDFGenerationError{
			Code:    domain.ErrCodeOutputPathInvalid,
			Message: errorMessage,
			Details: api.NewErrorDetails(api.ErrorCategoryGeneration, api.ErrorSeverityHigh).
				WithOutputPath(req.OutputPath).
				WithError(err),
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

// createConfigFromRequest creates a config from a PDF generation request
func (epsa *ExternalPDFServiceAdapter) createConfigFromRequest(ctx context.Context, req generation.PDFGenerationRequest) (*config.Config, error) {
	cfg := &config.Config{
		Template: config.Template(req.TemplatePath),
		Output:   config.Output(req.OutputPath),
		Engine:   config.Engine(req.Engine),
		Conversion: config.Conversion{
			Enabled: req.Options.Conversion.Enabled,
			Formats: req.Options.Conversion.Formats,
		},
		Variables:  *config.NewVariables(),
		Passes:     req.Options.Passes,
		UseLatexmk: req.Options.UseLatexmk,
	}

	// DEBUG: Log extracted config values (if logger is available)
	// Note: epsa.logger is ports.Logger, not cartas-backend logger, so these logs
	// will go to AutoPDF's logger, not cartas-backend's logger
	if epsa.logger != nil {
		epsa.logger.Info(ctx, "Extracting config from PDF generation request",
			autopdfports.NewLogField("request_passes", req.Options.Passes),
			autopdfports.NewLogField("request_use_latexmk", req.Options.UseLatexmk),
			autopdfports.NewLogField("config_passes", cfg.Passes),
			autopdfports.NewLogField("config_use_latexmk", cfg.UseLatexmk))
	}

	// Set variables from request (now using TemplateVariables)
	if req.Variables != nil {
		// Convert TemplateVariables to flattened map
		flattened := req.Variables.Flatten()
		for key, value := range flattened {
			if err := cfg.Variables.SetString(key, value); err != nil {
				return nil, fmt.Errorf("set request variable %q: %w", key, err)
			}
		}
	}

	return cfg, nil
}
