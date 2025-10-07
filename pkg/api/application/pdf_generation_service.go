// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/api"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
)

// PDFGenerationApplicationService orchestrates PDF generation using domain services
type PDFGenerationApplicationService struct {
	templateService  generation.TemplateProcessingService
	variableResolver generation.VariableResolver
	pdfValidator     generation.PDFValidator
	externalService  generation.PDFGenerationService
}

// NewPDFGenerationApplicationService creates a new application service
func NewPDFGenerationApplicationService(
	templateService generation.TemplateProcessingService,
	variableResolver generation.VariableResolver,
	pdfValidator generation.PDFValidator,
	externalService generation.PDFGenerationService,
) *PDFGenerationApplicationService {
	return &PDFGenerationApplicationService{
		templateService:  templateService,
		variableResolver: variableResolver,
		pdfValidator:     pdfValidator,
		externalService:  externalService,
	}
}

// GeneratePDF orchestrates the complete PDF generation workflow
func (s *PDFGenerationApplicationService) GeneratePDF(ctx context.Context, req generation.PDFGenerationRequest) (generation.PDFGenerationResult, error) {
	// Step 1: Validate the request
	if err := s.validateRequest(req); err != nil {
		return generation.PDFGenerationResult{
			Success: false,
			Error:   err,
		}, err
	}

	// Step 2: Resolve complex variables to simple key-value pairs
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

	// Step 3: Process template with resolved variables
	// Convert simple variables to interface{} map for template processing
	interfaceVars := make(map[string]interface{})
	for k, v := range simpleVariables {
		interfaceVars[k] = v
	}
	_, err = s.templateService.Process(ctx, req.TemplatePath, interfaceVars)
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

	// Step 4: Generate PDF using external service
	generationReq := generation.PDFGenerationRequest{
		TemplatePath: req.TemplatePath,
		Variables:    req.Variables, // Use original complex variables
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

	// Step 5: Validate the generated PDF
	if result.Success && result.PDFPath != "" {
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

	return result, nil
}

// ValidateTemplate validates a template file
func (s *PDFGenerationApplicationService) ValidateTemplate(templatePath string) error {
	return s.templateService.ValidateTemplate(templatePath)
}

// GetTemplateVariables extracts variables from a template
func (s *PDFGenerationApplicationService) GetTemplateVariables(templatePath string) ([]string, error) {
	return s.templateService.GetTemplateVariables(templatePath)
}

// GetSupportedEngines returns supported LaTeX engines
func (s *PDFGenerationApplicationService) GetSupportedEngines() []string {
	return s.externalService.GetSupportedEngines()
}

// GetSupportedFormats returns supported output formats
func (s *PDFGenerationApplicationService) GetSupportedFormats() []string {
	return s.externalService.GetSupportedFormats()
}

// validateRequest validates the PDF generation request
func (s *PDFGenerationApplicationService) validateRequest(req generation.PDFGenerationRequest) error {
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

	// Validate template
	if err := s.templateService.ValidateTemplate(req.TemplatePath); err != nil {
		return domain.TemplateProcessingError{
			Code:    domain.ErrCodeTemplateInvalid,
			Message: api.ErrTemplateValidationFailed,
			Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
				WithTemplatePath(req.TemplatePath).
				WithError(err),
		}
	}

	// Validate variables
	if err := s.variableResolver.Validate(req.Variables); err != nil {
		return domain.VariableResolutionError{
			Code:    domain.ErrCodeVariableInvalid,
			Message: api.ErrVariableValidationFailed,
			Details: api.NewErrorDetails(api.ErrorCategoryVariable, api.ErrorSeverityHigh).
				WithError(err),
		}
	}

	return nil
}
