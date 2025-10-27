// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/pkg/api/application"
	"github.com/BuddhiLW/AutoPDF/pkg/api/builders"
	apiconfig "github.com/BuddhiLW/AutoPDF/pkg/api/config"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
	"github.com/BuddhiLW/AutoPDF/pkg/api/factories"
	autopdfconfig "github.com/BuddhiLW/AutoPDF/pkg/config"
)

// PDFGenerationAPIService provides a clean API for PDF generation
type PDFGenerationAPIService struct {
	appService  *application.PDFGenerationApplicationService
	config      *autopdfconfig.Config
	debugConfig *apiconfig.APIDebugConfig
}

// NewPDFGenerationAPIService creates a new API service
func NewPDFGenerationAPIService(cfg *autopdfconfig.Config, logger *logger.LoggerAdapter) *PDFGenerationAPIService {
	// Load debug configuration from environment
	debugConfig := apiconfig.LoadDebugConfigFromEnv()

	// Create factory
	factory := factories.NewPDFGenerationServiceFactory(cfg, logger)

	// Create application service
	appService := factory.CreateApplicationService()

	return &PDFGenerationAPIService{
		appService:  appService,
		config:      cfg,
		debugConfig: debugConfig,
	}
}

// GeneratePDF generates a PDF using the builder pattern
func (s *PDFGenerationAPIService) GeneratePDF(ctx context.Context, templatePath string, outputPath string, variables map[string]interface{}) ([]byte, map[string]string, error) {
	// Build request using builder pattern
	request := builders.NewPDFGenerationRequestBuilder().
		WithTemplate(templatePath).
		WithOutput(outputPath).
		WithEngine(s.config.Engine.String()).
		WithVariables(variables).
		WithConversion(s.config.Conversion.Enabled, s.config.Conversion.Formats...).
		WithCleanup(false). // Don't clean for API usage
		WithTimeout(30 * time.Second).
		Build()

	// Generate PDF
	result, err := s.appService.GeneratePDF(ctx, request)
	if err != nil {
		return nil, nil, err
	}

	if !result.Success {
		return nil, nil, result.Error
	}

	// Read PDF bytes
	pdfBytes, err := os.ReadFile(result.PDFPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read generated PDF: %w", err)
	}

	// Create image paths map
	imagePaths := make(map[string]string)
	for i, imagePath := range result.ImagePaths {
		imagePaths[fmt.Sprintf("image_%d", i)] = imagePath
	}

	return pdfBytes, imagePaths, nil
}

// GeneratePDFWithOptions generates a PDF with custom options
func (s *PDFGenerationAPIService) GeneratePDFWithOptions(ctx context.Context, options PDFGenerationOptions) ([]byte, map[string]string, error) {
	// Build request with custom options
	verboseLevel := 0
	if options.Verbose {
		verboseLevel = 1
	}

	debugOptions := generation.DebugOptions{
		Enabled: options.Debug,
	}

	request := builders.NewPDFGenerationRequestBuilder().
		WithTemplate(options.TemplatePath).
		WithOutput(options.OutputPath).
		WithEngine(options.Engine).
		WithVariables(options.Variables).
		WithConversion(options.Conversion.Enabled, options.Conversion.Formats...).
		WithCleanup(options.Cleanup).
		WithTimeout(options.Timeout).
		WithVerbose(verboseLevel).
		WithDebug(debugOptions).
		Build()

	// Generate PDF
	result, err := s.appService.GeneratePDF(ctx, request)
	if err != nil {
		return nil, nil, err
	}

	if !result.Success {
		return nil, nil, result.Error
	}

	// Read PDF bytes
	pdfBytes, err := os.ReadFile(result.PDFPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read generated PDF: %w", err)
	}

	// Create image paths map
	imagePaths := make(map[string]string)
	for i, imagePath := range result.ImagePaths {
		imagePaths[fmt.Sprintf("image_%d", i)] = imagePath
	}

	return pdfBytes, imagePaths, nil
}

// GeneratePDFFromStruct generates a PDF from a struct (converts struct to variables automatically)
func (s *PDFGenerationAPIService) GeneratePDFFromStruct(ctx context.Context, templatePath string, outputPath string, data interface{}) ([]byte, map[string]string, error) {
	// Build debug options from environment config
	debugOptions := generation.DebugOptions{
		Enabled:            s.debugConfig.Enabled,
		LogToFile:          s.debugConfig.Enabled, // Enable log files when debug is on
		LogFilePath:        s.debugConfig.GetLogDirectory(),
		CreateConcreteFile: s.debugConfig.Enabled, // Create concrete .tex files for inspection
		RequestID:          fmt.Sprintf("request-%d", time.Now().Unix()),
	}

	// Build request using builder pattern with struct conversion
	requestBuilder := builders.NewPDFGenerationRequestBuilder().
		WithTemplate(templatePath).
		WithOutput(outputPath).
		WithEngine(s.config.Engine.String()).
		WithVariablesFromStruct(data). // Convert struct to TemplateVariables
		WithConversion(s.config.Conversion.Enabled, s.config.Conversion.Formats...).
		WithCleanup(false). // Don't clean for API usage
		WithTimeout(30 * time.Second).
		WithDebug(debugOptions) // Add debug options here

	// Add format file if configured (DIP: depends on config abstraction)
	if !s.config.FormatFile.IsEmpty() {
		requestBuilder = requestBuilder.WithFormatFile(s.config.FormatFile.String())
	}

	// Add working directory if configured (DIP: depends on config abstraction)
	if !s.config.WorkingDir.IsEmpty() {
		requestBuilder = requestBuilder.WithWorkingDir(s.config.WorkingDir.String())
	}

	request := requestBuilder.Build()

	// Generate PDF
	result, err := s.appService.GeneratePDF(ctx, request)
	if err != nil {
		return nil, nil, err
	}

	if !result.Success {
		return nil, nil, result.Error
	}

	// Read PDF bytes
	pdfBytes, err := os.ReadFile(result.PDFPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read generated PDF: %w", err)
	}

	// Create image paths map
	imagePaths := make(map[string]string)
	for i, imagePath := range result.ImagePaths {
		imagePaths[fmt.Sprintf("image_%d", i)] = imagePath
	}

	return pdfBytes, imagePaths, nil
}

// ValidateTemplate validates a template file
func (s *PDFGenerationAPIService) ValidateTemplate(templatePath string) error {
	return s.appService.ValidateTemplate(templatePath)
}

// GetTemplateVariables extracts variables from a template
func (s *PDFGenerationAPIService) GetTemplateVariables(templatePath string) ([]string, error) {
	return s.appService.GetTemplateVariables(templatePath)
}

// GetSupportedEngines returns supported LaTeX engines
func (s *PDFGenerationAPIService) GetSupportedEngines() []string {
	return s.appService.GetSupportedEngines()
}

// GetSupportedFormats returns supported output formats
func (s *PDFGenerationAPIService) GetSupportedFormats() []string {
	return s.appService.GetSupportedFormats()
}

// PDFGenerationOptions represents options for PDF generation
type PDFGenerationOptions struct {
	TemplatePath string
	OutputPath   string
	Engine       string
	Variables    map[string]interface{}
	Conversion   ConversionOptions
	Cleanup      bool
	Timeout      time.Duration
	Verbose      bool
	Debug        bool
}

// ConversionOptions represents conversion options
type ConversionOptions struct {
	Enabled bool
	Formats []string
}

// NewPDFGenerationOptions creates new PDF generation options
func NewPDFGenerationOptions(templatePath, outputPath string) *PDFGenerationOptions {
	return &PDFGenerationOptions{
		TemplatePath: templatePath,
		OutputPath:   outputPath,
		Engine:       "pdflatex",
		Variables:    make(map[string]interface{}),
		Conversion: ConversionOptions{
			Enabled: false,
			Formats: []string{},
		},
		Cleanup: false,
		Timeout: 30 * time.Second,
		Verbose: false,
		Debug:   false,
	}
}

// WithEngine sets the LaTeX engine
func (o *PDFGenerationOptions) WithEngine(engine string) *PDFGenerationOptions {
	o.Engine = engine
	return o
}

// WithVariable sets a variable
func (o *PDFGenerationOptions) WithVariable(key string, value interface{}) *PDFGenerationOptions {
	if o.Variables == nil {
		o.Variables = make(map[string]interface{})
	}
	o.Variables[key] = value
	return o
}

// WithVariables sets multiple variables
func (o *PDFGenerationOptions) WithVariables(variables map[string]interface{}) *PDFGenerationOptions {
	o.Variables = variables
	return o
}

// WithConversion enables conversion
func (o *PDFGenerationOptions) WithConversion(enabled bool, formats ...string) *PDFGenerationOptions {
	o.Conversion.Enabled = enabled
	o.Conversion.Formats = formats
	return o
}

// WithCleanup enables cleanup
func (o *PDFGenerationOptions) WithCleanup(enabled bool) *PDFGenerationOptions {
	o.Cleanup = enabled
	return o
}

// WithTimeout sets timeout
func (o *PDFGenerationOptions) WithTimeout(timeout time.Duration) *PDFGenerationOptions {
	o.Timeout = timeout
	return o
}

// WithVerbose enables verbose logging
func (o *PDFGenerationOptions) WithVerbose(enabled bool) *PDFGenerationOptions {
	o.Verbose = enabled
	return o
}

// WithDebug enables debug logging
func (o *PDFGenerationOptions) WithDebug(enabled bool) *PDFGenerationOptions {
	o.Debug = enabled
	return o
}
