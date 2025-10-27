// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package factories

import (
	"fmt"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/factories"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/valueobjects"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/watch"
	external_pdf_service "github.com/BuddhiLW/AutoPDF/pkg/api/adapters/external_pdf_service"
	"github.com/BuddhiLW/AutoPDF/pkg/api/adapters/pdf_validator"
	"github.com/BuddhiLW/AutoPDF/pkg/api/adapters/template_processor"
	"github.com/BuddhiLW/AutoPDF/pkg/api/adapters/variable_resolver"
	"github.com/BuddhiLW/AutoPDF/pkg/api/adapters/watch_service"
	"github.com/BuddhiLW/AutoPDF/pkg/api/application"
	apiconfig "github.com/BuddhiLW/AutoPDF/pkg/api/config"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"go.uber.org/zap"
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

	// Check if V2 compiler should be used
	// Strategy Pattern: Auto-select V2 when format files configured (smart default)
	debugConfig := apiconfig.LoadDebugConfigFromEnv()
	var externalService generation.PDFGenerationService

	// Debug: Log config state for troubleshooting
	f.logger.Info("Compiler selection debug",
		zap.String("format_file", f.config.FormatFile.String()),
		zap.Bool("format_file_is_empty", f.config.FormatFile.IsEmpty()),
		zap.Bool("v2_explicitly_enabled", debugConfig.IsV2CompilerEnabled()),
	)

	// Use V2 compiler if:
	// 1. Explicitly enabled via AUTOPDF_USE_V2_COMPILER=true, OR
	// 2. Format file is configured (smart default for performance)
	useV2Compiler := debugConfig.IsV2CompilerEnabled() || !f.config.FormatFile.IsEmpty()

	if useV2Compiler {
		reason := "explicitly enabled"
		if !f.config.FormatFile.IsEmpty() {
			reason = "format file configured (auto-enabled)"
		}
		f.logger.Info("Using V2 LaTeX compiler (CLARITY-refactored)", zap.String("reason", reason))

		var err error
		externalService, err = f.CreateExternalServiceV2(debugConfig)
		if err != nil {
			f.logger.Warn("Failed to create V2 compiler, falling back to legacy", zap.Error(err))
			externalService = external_pdf_service.NewExternalPDFServiceAdapter(f.config)
		}
	} else {
		f.logger.Info("Using legacy LaTeX compiler (no format file, V2 not enabled)")
		externalService = external_pdf_service.NewExternalPDFServiceAdapter(f.config)
	}

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

// CreateExternalServiceV2 creates an external PDF service using V2 compiler
func (f *PDFGenerationServiceFactory) CreateExternalServiceV2(
	debugConfig *apiconfig.APIDebugConfig,
) (generation.PDFGenerationService, error) {
	// Convert API debug config to value object
	voDebugConfig, err := valueobjects.NewDebugConfig(
		debugConfig.Enabled,
		debugConfig.GetConcreteFileDirectory(),
		debugConfig.GetLogDirectory(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create debug config: %w", err)
	}

	// Create LaTeX compiler factory
	compilerFactory := factories.NewLaTeXCompilerFactory(
		f.config,
		voDebugConfig,
	)

	// Create V2 compiler with decorators
	v2Compiler := compilerFactory.CreateCompiler()

	// Wrap V2 compiler in external service adapter
	return external_pdf_service.NewExternalPDFServiceWithV2Compiler(
		f.config,
		v2Compiler,
	), nil
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

// CreatePooledApplicationService creates a PDF generation service with warm worker pool.
// This provides significant performance improvements for high-volume PDF generation:
// - Cold compilation: ~1.2-2.5s
// - Warm compilation: ~100-300ms (4-12x faster)
//
// The pool is configured via PoolConfig and maintains persistent LaTeX processes.
// This method is intended for long-running services (web servers, batch processors).
//
// Note: This requires the cartas-backend worker pool implementation.
// If not available, use CreateApplicationService() for standard (cold) compilation.
//
// Example usage:
//   poolConfig, _ := worker.NewPoolConfig(
//       worker.WithMinWorkers(2),
//       worker.WithMaxWorkers(10),
//   )
//   service, _ := factory.CreatePooledApplicationService(poolConfig)
//   defer service.Close() // Important: shutdown pool on exit
//
// This is commented out by default since it creates a circular dependency.
// Uncomment and import when using the worker pool:
//
// import (
//     workerdom "github.com/BuddhiLW/cartas-backend/pkg/latex/worker/domain"
//     workerinfra "github.com/BuddhiLW/cartas-backend/pkg/latex/worker/infrastructure"
// )
//
// func (f *PDFGenerationServiceFactory) CreatePooledApplicationService(
//     poolConfig *workerdom.PoolConfig,
// ) (*application.PDFGenerationApplicationService, error) {
//     // Create pooled LaTeX adapter
//     pooledAdapter, err := workerinfra.NewPooledLaTeXAdapter(poolConfig)
//     if err != nil {
//         return nil, fmt.Errorf("failed to create pooled adapter: %w", err)
//     }
//
//     // Create other adapters (same as CreateApplicationService)
//     templateAdapter := template_processor.NewTemplateProcessorAdapter(f.config, f.logger)
//     variableResolver := variable_resolver.NewVariableResolverAdapter(f.config, f.logger)
//     pdfValidator := pdf_validator.NewPDFValidatorAdapter()
//
//     // Use pooled adapter instead of external service
//     // This requires modifying external_pdf_service to accept custom compiler
//     externalService := external_pdf_service.NewExternalPDFServiceWithCompiler(
//         f.config,
//         pooledAdapter, // Inject pooled compiler
//     )
//
//     // Create watch service dependencies
//     watchService := &minimalWatchService{}
//     watchManager := watch_service.NewWatchModeManagerAdapter(f.logger)
//     watchServiceAdapter := watch_service.NewWatchServiceAdapter(watchService, watchManager, f.logger)
//
//     // Create application service
//     return application.NewPDFGenerationApplicationService(
//         templateAdapter,
//         variableResolver,
//         pdfValidator,
//         externalService,
//         watchServiceAdapter,
//         watchManager,
//         f.logger,
//     ), nil
// }
