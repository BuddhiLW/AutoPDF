// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"
	"os"
	"time"

	ports "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/watch"
	"github.com/BuddhiLW/AutoPDF/pkg/api"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
)

// PDFOrchestrationService encapsulates all orchestration concerns
type PDFOrchestrationService struct {
	// Core services
	templateService  generation.TemplateProcessingService
	variableResolver generation.VariableResolver
	pdfValidator     generation.PDFValidator
	externalService  generation.PDFGenerationService
	logger           ports.Logger

	// Encapsulated services
	watchService generation.WatchService

	// Guards for validation and conditional logic
	requestGuard       ValidationGuard
	pdfValidationGuard *PDFValidationGuard
}

// NewPDFOrchestrationService creates a new orchestration service
func NewPDFOrchestrationService(
	templateService generation.TemplateProcessingService,
	variableResolver generation.VariableResolver,
	pdfValidator generation.PDFValidator,
	externalService generation.PDFGenerationService,
	watchService watch.WatchService,
	watchManager generation.WatchModeManager,
	logger ports.Logger,
) *PDFOrchestrationService {
	if logger == nil {
		logger = ports.NewNoOpLogger()
	}

	// Initialize guards
	requestGuard := NewRequestValidationGuard(templateService, variableResolver)
	pdfValidationGuard := NewPDFValidationGuard()

	// Initialize encapsulated watch service
	encapsulatedWatchService := NewWatchService(watchService, watchManager)

	return &PDFOrchestrationService{
		templateService:  templateService,
		variableResolver: variableResolver,
		pdfValidator:     pdfValidator,
		externalService:  externalService,
		logger:           logger,

		// Encapsulated services
		watchService: encapsulatedWatchService,

		// Initialize guards
		requestGuard:       requestGuard,
		pdfValidationGuard: pdfValidationGuard,
	}
}

// GeneratePDF orchestrates the complete PDF generation workflow
func (s *PDFOrchestrationService) GeneratePDF(ctx context.Context, req generation.PDFGenerationRequest) (generation.PDFGenerationResult, error) {
	var variableCount int
	if req.Variables != nil {
		variableCount = req.Variables.Len()
	}

	s.logger.Info(ctx, "Starting PDF generation",
		ports.NewLogField("template_path", req.TemplatePath),
		ports.NewLogField("engine", req.Engine),
		ports.NewLogField("output_path", req.OutputPath),
		ports.NewLogField("variable_count", variableCount),
		ports.NewLogField("debug_enabled", req.Options.Debug.Enabled),
	)

	// Step 1: Validate the request using guard
	if err := s.requestGuard.Validate(ctx, req); err != nil {
		return generation.PDFGenerationResult{
			Success: false,
			Error:   err,
		}, err
	}

	// Step 2: Resolve complex variables to simple key-value pairs
	s.logger.Debug(ctx, "Starting variable resolution",
		ports.NewLogField("input_variable_count", variableCount),
	)

	simpleVariables, err := s.variableResolver.Resolve(req.Variables)
	if err != nil {
		// Format the error message properly to avoid literal %s
		errorMessage := fmt.Sprintf(api.ErrVariableResolutionFailed, err.Error())
		return generation.PDFGenerationResult{
			Success: false,
			Error: domain.VariableResolutionError{
				Code:    domain.ErrCodeVariableInvalid,
				Message: errorMessage,
				Details: api.NewErrorDetails(api.ErrorCategoryVariable, api.ErrorSeverityHigh).
					WithError(err),
			},
		}, err
	}

	s.logger.Debug(ctx, "Variable resolution completed",
		ports.NewLogField("resolved_variable_count", len(simpleVariables)),
	)

	// Step 3: Process template without logging variable names or values.
	s.logger.Debug(ctx, "Processing template", ports.NewLogField("variable_count", len(simpleVariables)))

	processedContent, err := s.templateService.Process(ctx, req.TemplatePath, simpleVariables)
	if err != nil {
		// Format the error message properly to avoid literal %s
		errorMessage := fmt.Sprintf(api.ErrTemplateProcessingFailed, err.Error())
		return generation.PDFGenerationResult{
			Success: false,
			Error: domain.TemplateProcessingError{
				Code:    domain.ErrCodeTemplateInvalid,
				Message: errorMessage,
				Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
					WithTemplatePath(req.TemplatePath).
					WithError(err),
			},
		}, err
	}

	// Write processed template to a temporary file
	// This ensures the LaTeX engine uses the processed content with variables replaced
	tempFile, err := os.CreateTemp("", "autopdf-processed-*.tex")
	if err != nil {
		return generation.PDFGenerationResult{
			Success: false,
			Error: domain.TemplateProcessingError{
				Code:    domain.ErrCodeTemplateInvalid,
				Message: "Failed to create temporary file for processed template",
				Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
					WithTemplatePath(req.TemplatePath).
					WithError(err),
			},
		}, err
	}
	defer func() { _ = os.Remove(tempFile.Name()) }() // Clean up after generation.

	// Log size only; processed content may contain secrets or personal data.
	contentLen := len(processedContent)
	s.logger.Debug(ctx, "Processed template content",
		ports.NewLogField("content_length", contentLen),
	)

	if _, err := tempFile.WriteString(processedContent); err != nil {
		_ = tempFile.Close()
		return generation.PDFGenerationResult{
			Success: false,
			Error: domain.TemplateProcessingError{
				Code:    domain.ErrCodeTemplateInvalid,
				Message: "Failed to write processed template to temporary file",
				Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
					WithTemplatePath(req.TemplatePath).
					WithError(err),
			},
		}, err
	}
	if err := tempFile.Close(); err != nil {
		return generation.PDFGenerationResult{Success: false}, fmt.Errorf("close processed template: %w", err)
	}

	// Log temporary file path
	s.logger.Info(ctx, "Using temporary template file",
		ports.NewLogField("temp_file", tempFile.Name()),
	)

	// Step 4: Generate PDF using external service with processed template
	generationReq := generation.PDFGenerationRequest{
		TemplatePath: tempFile.Name(), // Use processed template file
		Variables:    req.Variables,   // Keep original variables for metadata
		Engine:       req.Engine,
		OutputPath:   req.OutputPath,
		Options:      req.Options,
	}

	result, err := s.externalService.Generate(ctx, generationReq)
	if err != nil {
		// Format the error message properly to avoid literal %s
		errorMessage := fmt.Sprintf(api.ErrPDFGenerationFailed, err.Error())
		return generation.PDFGenerationResult{
			Success: false,
			Error: domain.PDFGenerationError{
				Code:    domain.ErrCodePDFGenerationFailed,
				Message: errorMessage,
				Details: api.NewErrorDetails(api.ErrorCategoryGeneration, api.ErrorSeverityHigh).
					WithError(err),
			},
		}, err
	}

	// Step 5: Validate the generated PDF using guard
	if s.pdfValidationGuard.ShouldValidatePDF(result) {
		if err := s.pdfValidator.Validate(result.PDFPath); err != nil {
			return generation.PDFGenerationResult{
				Success: false,
				Error: domain.PDFGenerationError{
					Code:    domain.ErrCodePDFValidationFailed,
					Message: api.ErrPDFValidationFailed,
					Details: api.NewErrorDetails(api.ErrorCategoryPDF, api.ErrorSeverityHigh).
						WithFilePath(result.PDFPath).
						WithError(err),
				},
			}, err
		}

		// Get metadata for the result
		metadata, err := s.pdfValidator.GetMetadata(result.PDFPath)
		if err != nil {
			// Log warning but don't fail
			metadata = generation.PDFMetadata{
				GeneratedAt: time.Now(),
				Engine:      req.Engine,
				Template:    req.TemplatePath,
			}
		}
		result.Metadata = metadata
	}

	s.logger.Info(ctx, "PDF generation completed",
		ports.NewLogField("success", result.Success),
		ports.NewLogField("pdf_path", result.PDFPath),
		ports.NewLogField("image_count", len(result.ImagePaths)),
	)

	// Step 6: Handle watch mode if enabled using encapsulated service
	if s.watchService.ShouldStartWatchMode(req, result) {
		if err := s.watchService.StartWatchMode(ctx, req); err != nil {
			s.logger.Error(ctx, "Failed to start watch mode",
				ports.NewLogField("error", err),
				ports.NewLogField("template_path", req.TemplatePath),
			)
			// Don't fail the generation, just log the error
		} else {
			s.logger.Info(ctx, "Watch mode started successfully",
				ports.NewLogField("template_path", req.TemplatePath),
			)
		}
	}

	return result, nil
}

// ValidateTemplate validates a template file
func (s *PDFOrchestrationService) ValidateTemplate(templatePath string) error {
	return s.templateService.ValidateTemplate(templatePath)
}

// GetTemplateVariables extracts variables from a template
func (s *PDFOrchestrationService) GetTemplateVariables(templatePath string) ([]string, error) {
	return s.templateService.GetTemplateVariables(templatePath)
}

// GetSupportedEngines returns supported LaTeX engines
func (s *PDFOrchestrationService) GetSupportedEngines() []string {
	return s.externalService.GetSupportedEngines()
}

// GetSupportedFormats returns supported output formats
func (s *PDFOrchestrationService) GetSupportedFormats() []string {
	return s.externalService.GetSupportedFormats()
}

// GetActiveWatchModes returns information about active watch modes
func (s *PDFOrchestrationService) GetActiveWatchModes() map[string]generation.WatchInstanceInfo {
	return s.watchService.GetActiveWatchModes()
}

// StopWatchMode stops a specific watch mode
func (s *PDFOrchestrationService) StopWatchMode(watchID string) error {
	return s.watchService.StopWatchMode(watchID)
}

// StopAllWatchModes stops all active watch modes
func (s *PDFOrchestrationService) StopAllWatchModes() error {
	return s.watchService.StopAllWatchModes()
}
