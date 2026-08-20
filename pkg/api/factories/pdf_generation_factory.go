// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package factories

import (
	"context"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	autopdfports "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/watch"
	infraadapters "github.com/BuddhiLW/AutoPDF/internal/autopdf/infrastructure/adapters"
	"github.com/BuddhiLW/AutoPDF/pkg/api"
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
	portLogger   autopdfports.Logger // Optional logger for latexmk transparency
}

// NewPDFGenerationServiceFactory creates a new factory
func NewPDFGenerationServiceFactory(cfg *config.Config, publicLogger api.Logger, debugEnabled bool) *PDFGenerationServiceFactory {
	legacyLogger := logger.NewLoggerAdapter(logger.Silent, "stderr")
	return &PDFGenerationServiceFactory{
		config:       cfg,
		logger:       legacyLogger,
		debugEnabled: debugEnabled,
		portLogger:   adaptPublicLogger(publicLogger),
	}
}

type publicLoggerPort struct{ logger api.Logger }

func adaptPublicLogger(logger api.Logger) autopdfports.Logger {
	if logger == nil {
		return nil
	}
	return &publicLoggerPort{logger: logger}
}

func publicFields(fields []autopdfports.LogField) []api.LogField {
	result := make([]api.LogField, len(fields))
	for index, field := range fields {
		result[index] = api.NewLogField(field.Key, field.Value)
	}
	return result
}

func (adapter *publicLoggerPort) Debug(ctx context.Context, message string, fields ...autopdfports.LogField) {
	adapter.logger.Debug(ctx, message, publicFields(fields)...)
}
func (adapter *publicLoggerPort) Info(ctx context.Context, message string, fields ...autopdfports.LogField) {
	adapter.logger.Info(ctx, message, publicFields(fields)...)
}
func (adapter *publicLoggerPort) Warn(ctx context.Context, message string, fields ...autopdfports.LogField) {
	adapter.logger.Warn(ctx, message, publicFields(fields)...)
}
func (adapter *publicLoggerPort) Error(ctx context.Context, message string, fields ...autopdfports.LogField) {
	adapter.logger.Error(ctx, message, publicFields(fields)...)
}

// CreateApplicationService creates a PDF generation application service
func (f *PDFGenerationServiceFactory) CreateApplicationService() *application.PDFGenerationApplicationService {
	// Create adapters
	templateAdapter := template_processor.NewTemplateProcessorAdapter(f.config, f.logger)
	variableResolver := variable_resolver.NewVariableResolverAdapter(f.config, f.logger)
	pdfValidator := pdf_validator.NewPDFValidatorAdapter()

	// Create external service with logger (for latexmk transparency)
	externalService := f.CreateExternalService()

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
	// Use portLogger if set (from cartas-backend), otherwise convert AutoPDF logger adapter
	var portLogger autopdfports.Logger
	if f.portLogger != nil {
		portLogger = f.portLogger
	} else if f.logger != nil {
		portLogger = infraadapters.NewLoggerPortAdapter(f.logger)
	}
	return external_pdf_service.NewExternalPDFServiceAdapterWithLogger(f.config, f.debugEnabled, portLogger)
}

// CreateCompleteService creates a complete service with all dependencies
func (f *PDFGenerationServiceFactory) CreateCompleteService() *application.PDFGenerationApplicationService {
	return f.CreateApplicationService()
}

// minimalWatchService provides a minimal implementation for factory usage
type minimalWatchService struct{}

func (m *minimalWatchService) StartWatching(config watch.WatchConfiguration) error {
	return watch_service.ErrWatchModeUnavailable
}

func (m *minimalWatchService) StopWatching() error {
	return watch_service.ErrWatchModeUnavailable
}

func (m *minimalWatchService) ConfigureExclusions(patterns []string) error {
	return watch_service.ErrWatchModeUnavailable
}

func (m *minimalWatchService) ConfigureInterval(interval time.Duration) error {
	return watch_service.ErrWatchModeUnavailable
}
