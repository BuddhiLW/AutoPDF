// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package factories

import (
	"github.com/BuddhiLW/AutoPDF/pkg/api/adapters"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// PDFGenerationServiceFactory creates PDF generation services
type PDFGenerationServiceFactory struct {
	config *config.Config
}

// NewPDFGenerationServiceFactory creates a new factory
func NewPDFGenerationServiceFactory(cfg *config.Config) *PDFGenerationServiceFactory {
	return &PDFGenerationServiceFactory{
		config: cfg,
	}
}

// CreateApplicationService creates a PDF generation application service
func (f *PDFGenerationServiceFactory) CreateApplicationService() *services.PDFGenerationApplicationService {
	// Create adapters
	templateAdapter := adapters.NewTemplateProcessorAdapter(f.config)
	variableResolver := adapters.NewVariableResolverAdapter(f.config)
	pdfValidator := adapters.NewPDFValidatorAdapter()
	externalService := adapters.NewExternalPDFServiceAdapter(f.config)

	// Create application service
	return services.NewPDFGenerationApplicationService(
		templateAdapter,
		variableResolver,
		pdfValidator,
		externalService,
	)
}

// CreateTemplateService creates a template processing service
func (f *PDFGenerationServiceFactory) CreateTemplateService() domain.TemplateProcessingService {
	return adapters.NewTemplateProcessorAdapter(f.config)
}

// CreateVariableResolver creates a variable resolver
func (f *PDFGenerationServiceFactory) CreateVariableResolver() domain.VariableResolver {
	return adapters.NewVariableResolverAdapter(f.config)
}

// CreatePDFValidator creates a PDF validator
func (f *PDFGenerationServiceFactory) CreatePDFValidator() domain.PDFValidator {
	return adapters.NewPDFValidatorAdapter()
}

// CreateExternalService creates an external PDF service
func (f *PDFGenerationServiceFactory) CreateExternalService() domain.PDFGenerationService {
	return adapters.NewExternalPDFServiceAdapter(f.config)
}

// CreateCompleteService creates a complete service with all dependencies
func (f *PDFGenerationServiceFactory) CreateCompleteService() *services.PDFGenerationApplicationService {
	return f.CreateApplicationService()
}
