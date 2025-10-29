// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package factories

import (
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/watch"
	external_pdf_service "github.com/BuddhiLW/AutoPDF/pkg/api/adapters/external_pdf_service"
	"github.com/BuddhiLW/AutoPDF/pkg/api/adapters/pdf_validator"
	"github.com/BuddhiLW/AutoPDF/pkg/api/adapters/template_processor"
	"github.com/BuddhiLW/AutoPDF/pkg/api/adapters/variable_resolver"
	"github.com/BuddhiLW/AutoPDF/pkg/api/adapters/watch_service"
	"github.com/BuddhiLW/AutoPDF/pkg/api/application"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// PDFGenerationServiceFactory creates PDF generation services
type PDFGenerationServiceFactory struct {
	config       *config.Config
	logger       *logger.LoggerAdapter
	debugEnabled bool
}

// NewPDFGenerationServiceFactory creates a new factory
func NewPDFGenerationServiceFactory(cfg *config.Config, logger *logger.LoggerAdapter, debugEnabled bool) *PDFGenerationServiceFactory {
	return &PDFGenerationServiceFactory{
		config:       cfg,
		logger:       logger,
		debugEnabled: debugEnabled,
	}
}

// CreateApplicationService creates a PDF generation application service
func (f *PDFGenerationServiceFactory) CreateApplicationService() *application.PDFGenerationApplicationService {
	// Create adapters
	templateAdapter := template_processor.NewTemplateProcessorAdapter(f.config, f.logger)
	variableResolver := variable_resolver.NewVariableResolverAdapter(f.config, f.logger)
	pdfValidator := pdf_validator.NewPDFValidatorAdapter()
	externalService := external_pdf_service.NewExternalPDFServiceAdapter(f.config, f.debugEnabled)

	// Create watch service dependencies
	// For factory usage, create a minimal watch service
	watchService := &minimalWatchService{}
	watchManager := watch_service.NewWatchModeManagerAdapter(f.logger)
	watchServiceAdapter := watch_service.NewWatchServiceAdapter(watchService, watchManager, f.logger)

	// Create application service
	return application.NewPDFGenerationApplicationService(
		templateAdapter,
		variableResolver,
		pdfValidator,
		externalService,
		watchServiceAdapter,
		watchManager,
		f.logger,
		f.debugEnabled,
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
	return external_pdf_service.NewExternalPDFServiceAdapter(f.config, f.debugEnabled)
}

// CreateCompleteService creates a complete service with all dependencies
func (f *PDFGenerationServiceFactory) CreateCompleteService() *application.PDFGenerationApplicationService {
	return f.CreateApplicationService()
}

// minimalWatchService provides a minimal implementation for factory usage
type minimalWatchService struct{}

func (m *minimalWatchService) StartWatching(config watch.WatchConfiguration) error {
	return nil
}

func (m *minimalWatchService) StopWatching() error {
	return nil
}

func (m *minimalWatchService) ConfigureExclusions(patterns []string) error {
	return nil
}

func (m *minimalWatchService) ConfigureInterval(interval time.Duration) error {
	return nil
}
