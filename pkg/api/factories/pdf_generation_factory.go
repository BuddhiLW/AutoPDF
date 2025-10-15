// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package factories

import (
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	external_pdf_service "github.com/BuddhiLW/AutoPDF/pkg/api/adapters/external_pdf_service"
	"github.com/BuddhiLW/AutoPDF/pkg/api/adapters/pdf_validator"
	"github.com/BuddhiLW/AutoPDF/pkg/api/adapters/template_processor"
	"github.com/BuddhiLW/AutoPDF/pkg/api/adapters/variable_resolver"
	"github.com/BuddhiLW/AutoPDF/pkg/api/application"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// PDFGenerationServiceFactory creates PDF generation services
type PDFGenerationServiceFactory struct {
	config *config.Config
	logger *logger.LoggerAdapter
}

// NewPDFGenerationServiceFactory creates a new factory
func NewPDFGenerationServiceFactory(cfg *config.Config, logger *logger.LoggerAdapter) *PDFGenerationServiceFactory {
	return &PDFGenerationServiceFactory{
		config: cfg,
		logger: logger,
	}
}

// CreateApplicationService creates a PDF generation application service
func (f *PDFGenerationServiceFactory) CreateApplicationService() *application.PDFGenerationApplicationService {
	// Create adapters
	templateAdapter := template_processor.NewTemplateProcessorAdapter(f.config, f.logger)
	variableResolver := variable_resolver.NewVariableResolverAdapter(f.config, f.logger)
	pdfValidator := pdf_validator.NewPDFValidatorAdapter()
	externalService := external_pdf_service.NewExternalPDFServiceAdapter(f.config)

	// Create application service
	return application.NewPDFGenerationApplicationService(
		templateAdapter,
		variableResolver,
		pdfValidator,
		externalService,
		f.logger,
	)
}

// CreateTemplateService creates a template processing service
func (f *PDFGenerationServiceFactory) CreateTemplateService() generation.TemplateProcessingService {
	return template_processor.NewTemplateProcessorAdapter(f.config, f.logger)
}

// CreateVariableResolver creates a variable resolver
func (f *PDFGenerationServiceFactory) CreateVariableResolver() generation.VariableResolver {
	return variable_resolver.NewVariableResolverAdapter(f.config, f.logger)
}

// CreatePDFValidator creates a PDF validator
func (f *PDFGenerationServiceFactory) CreatePDFValidator() generation.PDFValidator {
	return pdf_validator.NewPDFValidatorAdapter()
}

// CreateExternalService creates an external PDF service
func (f *PDFGenerationServiceFactory) CreateExternalService() generation.PDFGenerationService {
	return external_pdf_service.NewExternalPDFServiceAdapter(f.config)
}

// CreateCompleteService creates a complete service with all dependencies
func (f *PDFGenerationServiceFactory) CreateCompleteService() *application.PDFGenerationApplicationService {
	return f.CreateApplicationService()
}
