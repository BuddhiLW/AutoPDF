// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/api/domain"
)

// PDFGenerationApplicationService orchestrates PDF generation using domain services
type PDFGenerationApplicationService struct {
	templateService  domain.TemplateProcessingService
	variableResolver domain.VariableResolver
	pdfValidator     domain.PDFValidator
	externalService  domain.PDFGenerationService
}

// NewPDFGenerationApplicationService creates a new application service
func NewPDFGenerationApplicationService(
	templateService domain.TemplateProcessingService,
	variableResolver domain.VariableResolver,
	pdfValidator domain.PDFValidator,
	externalService domain.PDFGenerationService,
) *PDFGenerationApplicationService {
	return &PDFGenerationApplicationService{
		templateService:  templateService,
		variableResolver: variableResolver,
		pdfValidator:     pdfValidator,
		externalService:  externalService,
	}
}

// GeneratePDF orchestrates the complete PDF generation workflow
func (s *PDFGenerationApplicationService) GeneratePDF(ctx context.Context, req domain.PDFGenerationRequest) (domain.PDFGenerationResult, error) {
	// Step 1: Validate the request
	if err := s.validateRequest(req); err != nil {
		return domain.PDFGenerationResult{
			Success: false,
			Error:   err,
		}, err
	}

	// Step 2: Resolve complex variables to simple key-value pairs
	simpleVariables, err := s.variableResolver.Resolve(req.Variables)
	if err != nil {
		return domain.PDFGenerationResult{
			Success: false,
			Error: domain.VariableResolutionError{
				Code:    domain.ErrCodeVariableInvalid,
				Message: "Failed to resolve variables",
				Details: map[string]interface{}{"error": err.Error()},
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
		return domain.PDFGenerationResult{
			Success: false,
			Error: domain.TemplateProcessingError{
				Code:    domain.ErrCodeTemplateInvalid,
				Message: "Failed to process template",
				Details: map[string]interface{}{"template": req.TemplatePath, "error": err.Error()},
			},
		}, err
	}

	// Step 4: Generate PDF using external service
	generationReq := domain.PDFGenerationRequest{
		TemplatePath: req.TemplatePath,
		Variables:    req.Variables, // Use original complex variables
		Engine:       req.Engine,
		OutputPath:   req.OutputPath,
		Options:      req.Options,
	}

	result, err := s.externalService.Generate(ctx, generationReq)
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

	// Step 5: Validate the generated PDF
	if result.Success && result.PDFPath != "" {
		if err := s.pdfValidator.Validate(result.PDFPath); err != nil {
			return domain.PDFGenerationResult{
				Success: false,
				Error: domain.PDFGenerationError{
					Code:    domain.ErrCodePDFValidationFailed,
					Message: "Generated PDF validation failed",
					Details: map[string]interface{}{"pdf_path": result.PDFPath, "error": err.Error()},
				},
			}, err
		}

		// Get metadata for the result
		metadata, err := s.pdfValidator.GetMetadata(result.PDFPath)
		if err != nil {
			// Log warning but don't fail
			metadata = domain.PDFMetadata{
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
func (s *PDFGenerationApplicationService) validateRequest(req domain.PDFGenerationRequest) error {
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
			Message: "Template validation failed",
			Details: map[string]interface{}{"template": req.TemplatePath, "error": err.Error()},
		}
	}

	// Validate variables
	if err := s.variableResolver.Validate(req.Variables); err != nil {
		return domain.VariableResolutionError{
			Code:    domain.ErrCodeVariableInvalid,
			Message: "Variable validation failed",
			Details: map[string]interface{}{"error": err.Error()},
		}
	}

	return nil
}
