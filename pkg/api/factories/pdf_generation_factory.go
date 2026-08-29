// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package factories

import (
	"context"
	"time"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/logger"
	autopdfports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/domain/watch"
	infraadapters "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/infrastructure/adapters"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	external_pdf_service "github.com/BuddhiLW/AutoPDF/v2/pkg/api/adapters/external_pdf_service"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/adapters/pdf_validator"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/adapters/template_processor"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/adapters/variable_resolver"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/adapters/watch_service"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/application"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/domain/generation"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
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
		f.resolvePortLogger(),
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

// resolvePortLogger returns the Logger port this factory injects into the
// services it builds: an externally supplied port takes precedence, otherwise
// the factory's own LoggerAdapter is bridged to the port. Never returns nil.
func (f *PDFGenerationServiceFactory) resolvePortLogger() autopdfports.Logger {
	if f.portLogger != nil {
		return f.portLogger
	}
	// NewLoggerPortAdapter yields a no-op Logger when f.logger is nil.
	return infraadapters.NewLoggerPortAdapter(f.logger)
}

// CreateExternalService creates an external PDF service
func (f *PDFGenerationServiceFactory) CreateExternalService() generation.PDFGenerationService {
	return external_pdf_service.NewExternalPDFServiceAdapterWithLogger(f.config, f.debugEnabled, f.resolvePortLogger())
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
