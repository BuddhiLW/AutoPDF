// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"os"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
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
	logger           *logger.LoggerAdapter

	// Encapsulated services
	watchService generation.WatchService

	// Guards for validation and conditional logic
	requestGuard        ValidationGuard
	pdfValidationGuard  *PDFValidationGuard
	contentPreviewGuard *ContentPreviewGuard
}

// NewPDFOrchestrationService creates a new orchestration service
func NewPDFOrchestrationService(
	templateService generation.TemplateProcessingService,
	variableResolver generation.VariableResolver,
	pdfValidator generation.PDFValidator,
	externalService generation.PDFGenerationService,
	watchService watch.WatchService,
	watchManager generation.WatchModeManager,
	logger *logger.LoggerAdapter,
) *PDFOrchestrationService {
	// Initialize guards
	requestGuard := NewRequestValidationGuard(templateService, variableResolver)
	pdfValidationGuard := NewPDFValidationGuard()
	contentPreviewGuard := NewContentPreviewGuard()

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
		requestGuard:        requestGuard,
		pdfValidationGuard:  pdfValidationGuard,
		contentPreviewGuard: contentPreviewGuard,
	}
}

// GeneratePDF orchestrates the complete PDF generation workflow
func (s *PDFOrchestrationService) GeneratePDF(ctx context.Context, req generation.PDFGenerationRequest) (generation.PDFGenerationResult, error) {
	var variableCount int
	if req.Variables != nil {
		variableCount = req.Variables.Len()
	}

	s.logger.InfoWithFields("Starting PDF generation",
		"template_path", req.TemplatePath,
		"engine", req.Engine,
		"output_path", req.OutputPath,
		"variable_count", variableCount,
		"debug_enabled", req.Options.Debug.Enabled,
	)

	// Step 1: Validate the request using guard
	if err := s.requestGuard.Validate(ctx, req); err != nil {
		return generation.PDFGenerationResult{
			Success: false,
			Error:   err,
		}, err
	}

	// Step 2: Resolve complex variables to simple key-value pairs
	s.logger.DebugWithFields("Starting variable resolution",
		"input_variable_count", variableCount,
	)

	simpleVariables, err := s.variableResolver.Resolve(req.Variables)
	if err != nil {
		return generation.PDFGenerationResult{
			Success: false,
			Error: domain.VariableResolutionError{
				Code:    domain.ErrCodeVariableInvalid,
				Message: api.ErrVariableResolutionFailed,
				Details: api.NewErrorDetails(api.ErrorCategoryVariable, api.ErrorSeverityHigh).
					WithError(err),
			},
		}, err
	}

	s.logger.DebugWithFields("Variable resolution completed",
		"resolved_variable_count", len(simpleVariables),
	)

	// Step 3: Process template with resolved variables
	// Log variables being processed
	s.logger.DebugWithFields("Processing template with variables",
		"variables", simpleVariables,
		"variable_count", len(simpleVariables),
	)

	processedContent, err := s.templateService.Process(ctx, req.TemplatePath, simpleVariables)
	if err != nil {
		return generation.PDFGenerationResult{
			Success: false,
			Error: domain.TemplateProcessingError{
				Code:    domain.ErrCodeTemplateInvalid,
				Message: api.ErrTemplateProcessingFailed,
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
	defer os.Remove(tempFile.Name()) // Clean up temporary file after generation

	// Log processed content using guard
	contentLen := len(processedContent)
	previewLen := s.contentPreviewGuard.GetPreviewLength(contentLen)
	s.logger.DebugWithFields("Processed template content",
		"content_length", contentLen,
		"preview_length", previewLen,
		"preview", processedContent[:previewLen],
	)

	if _, err := tempFile.WriteString(processedContent); err != nil {
		tempFile.Close()
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
	tempFile.Close()

	// Log temporary file path
	s.logger.InfoWithFields("Using temporary template file",
		"temp_file", tempFile.Name(),
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
		return generation.PDFGenerationResult{
			Success: false,
			Error: domain.PDFGenerationError{
				Code:    domain.ErrCodePDFGenerationFailed,
				Message: api.ErrPDFGenerationFailed,
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

	s.logger.InfoWithFields("PDF generation completed",
		"success", result.Success,
		"pdf_path", result.PDFPath,
		"image_count", len(result.ImagePaths),
	)

	// Step 6: Handle watch mode if enabled using encapsulated service
	if s.watchService.ShouldStartWatchMode(req, result) {
		if err := s.watchService.StartWatchMode(ctx, req); err != nil {
			s.logger.ErrorWithFields("Failed to start watch mode",
				"error", err,
				"template_path", req.TemplatePath,
			)
			// Don't fail the generation, just log the error
		} else {
			s.logger.InfoWithFields("Watch mode started successfully",
				"template_path", req.TemplatePath,
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
