// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/cleaner"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/converter"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/latex"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/template"
	autopdfports "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/ports"
	documentService "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/services/document"
	infraadapters "github.com/BuddhiLW/AutoPDF/internal/autopdf/infrastructure/adapters"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	errors "github.com/BuddhiLW/AutoPDF/pkg/errors"
	"gopkg.in/yaml.v3"
)

// InternalApplicationAdapter bridges the API layer with the internal application layer
// This adapter follows the Adapter pattern from GoF and maintains separation of concerns
type InternalApplicationAdapter struct {
	config *config.Config
	logger autopdfports.Logger // Optional logger for transparency
}

// NewInternalApplicationAdapter creates a new adapter
func NewInternalApplicationAdapter(cfg *config.Config) *InternalApplicationAdapter {
	return &InternalApplicationAdapter{
		config: cfg,
		logger: nil,
	}
}

// NewInternalApplicationAdapterWithLogger creates a new adapter with logger
func NewInternalApplicationAdapterWithLogger(cfg *config.Config, logger autopdfports.Logger) *InternalApplicationAdapter {
	return &InternalApplicationAdapter{
		config: cfg,
		logger: logger,
	}
}

// GeneratePDF generates a PDF using the internal application layer
// This method maintains the same signature as the original GeneratePDF function
// workingDir is optional - if empty, uses default behavior
func (iaa *InternalApplicationAdapter) GeneratePDF(cfg *config.Config, template config.Template, debugEnabled bool, workingDir string) ([]byte, map[string]string, error) {
	// Merge configuration
	mergedCfg := iaa.mergeConfig(cfg, template)

	// Create temporary config file
	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := iaa.createConfigFile(mergedCfg, configPath); err != nil {
		return nil, nil, err
	}
	defer os.Remove(configPath)

	// Create document service using the internal application layer
	// Use working directory if provided, otherwise use default
	var docService *documentService.DocumentService
	if workingDir != "" {
		docService = iaa.createDocumentServiceWithWorkingDir(mergedCfg, workingDir, iaa.logger)
	} else {
		docService = iaa.createDocumentService(mergedCfg)
	}

	// Create build request
	req := documentService.BuildRequest{
		TemplatePath: mergedCfg.Template.String(),
		ConfigPath:   configPath,
		Variables:    &mergedCfg.Variables,
		Engine:       mergedCfg.Engine.String(),
		OutputPath:   mergedCfg.Output.String(),
		WorkingDir:   workingDir, // Pass working directory to BuildRequest
		DoConvert:    mergedCfg.Conversion.Enabled,
		DoClean:      false, // Don't clean for API usage
		DebugEnabled: debugEnabled,
		Passes:       mergedCfg.Passes,
		UseLatexmk:   mergedCfg.UseLatexmk,
		Conversion: documentService.ConversionSettings{
			Enabled: mergedCfg.Conversion.Enabled,
			Formats: mergedCfg.Conversion.Formats,
		},
	}

	// Build the document
	ctx := context.Background()
	result, err := docService.Build(ctx, req)
	if err != nil {
		// Check if output file exists despite errors
		if _, statErr := os.Stat(mergedCfg.Output.String()); os.IsNotExist(statErr) {
			return nil, nil, err
		}
		// If file exists, continue with reading it
	}

	// Update output path if it was changed by the service
	if result.PDFPath != "" {
		mergedCfg.Output = config.Output(result.PDFPath)
	}

	// Read the generated PDF
	pdfBytes, err := os.ReadFile(mergedCfg.Output.String())
	if err != nil {
		return nil, nil, err
	}

	// Verify PDF content
	if len(pdfBytes) == 0 {
		return nil, nil, fmt.Errorf("generated PDF is empty")
	}

	// Basic PDF header validation
	if len(pdfBytes) < 5 || string(pdfBytes[0:5]) != "%PDF-" {
		return nil, nil, fmt.Errorf("generated file is not a valid PDF")
	}

	// Handle conversion results
	paths := make(map[string]string)
	if result.ImagePaths != nil {
		for i, imagePath := range result.ImagePaths {
			paths[fmt.Sprintf("image_%d", i)] = imagePath
		}
	}

	return pdfBytes, paths, nil
}

// GeneratePDFWithWorkingDir generates a PDF using the internal application layer with custom working directory
func (iaa *InternalApplicationAdapter) GeneratePDFWithWorkingDir(cfg *config.Config, template config.Template, debugEnabled bool, workingDir string) ([]byte, map[string]string, error) {
	// Merge configuration
	mergedCfg := iaa.mergeConfig(cfg, template)

	// Create temporary config file
	tmpDir := os.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := iaa.createConfigFile(mergedCfg, configPath); err != nil {
		return nil, nil, err
	}
	defer os.Remove(configPath)

	// Create document service using the internal application layer with custom working directory
	docService := iaa.createDocumentServiceWithWorkingDir(mergedCfg, workingDir, iaa.logger)

	// Create build request
	req := documentService.BuildRequest{
		TemplatePath: mergedCfg.Template.String(),
		ConfigPath:   configPath,
		Variables:    &mergedCfg.Variables,
		Engine:       mergedCfg.Engine.String(),
		OutputPath:   mergedCfg.Output.String(),
		WorkingDir:   workingDir, // Pass working directory to BuildRequest
		DoConvert:    mergedCfg.Conversion.Enabled,
		DoClean:      false, // Don't clean for API usage
		DebugEnabled: debugEnabled,
		Passes:       mergedCfg.Passes,
		UseLatexmk:   mergedCfg.UseLatexmk,
		Conversion: documentService.ConversionSettings{
			Enabled: mergedCfg.Conversion.Enabled,
			Formats: mergedCfg.Conversion.Formats,
		},
	}

	// Build the document
	ctx := context.Background()
	result, err := docService.Build(ctx, req)
	if err != nil {
		// Check if output file exists despite errors
		if _, statErr := os.Stat(mergedCfg.Output.String()); os.IsNotExist(statErr) {
			return nil, nil, err
		}
		// If file exists, continue with reading it
	}

	// Update output path if it was changed by the service
	if result.PDFPath != "" {
		mergedCfg.Output = config.Output(result.PDFPath)
	}

	// Read the generated PDF
	pdfBytes, err := os.ReadFile(mergedCfg.Output.String())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read generated PDF: %w", err)
	}

	// Handle conversion results
	paths := make(map[string]string)
	if result.ImagePaths != nil {
		for i, imagePath := range result.ImagePaths {
			paths[fmt.Sprintf("image_%d", i)] = imagePath
		}
	}

	return pdfBytes, paths, nil
}

// mergeConfig merges the provided config with defaults and template
func (iaa *InternalApplicationAdapter) mergeConfig(cfg *config.Config, template config.Template) *config.Config {
	defaultCfg := config.GetDefaultConfig()

	// Log incoming config values for debugging
	if iaa.logger != nil {
		iaa.logger.Info(context.Background(), "Merging AutoPDF config",
			autopdfports.NewLogField("incoming_passes", cfg.Passes),
			autopdfports.NewLogField("incoming_use_latexmk", cfg.UseLatexmk),
			autopdfports.NewLogField("default_passes", defaultCfg.Passes),
			autopdfports.NewLogField("default_use_latexmk", defaultCfg.UseLatexmk))
	}

	merged := &config.Config{
		Template:   cfg.Template,
		Output:     cfg.Output,
		Variables:  cfg.Variables,
		Engine:     cfg.Engine,
		Conversion: cfg.Conversion,
		Passes:     cfg.Passes,     // Preserve Passes from template config
		UseLatexmk: cfg.UseLatexmk, // Preserve UseLatexmk from template config
	}

	// Apply template if not set
	if merged.Template == "" {
		merged.Template = template
	}

	// Apply defaults for missing values
	if merged.Variables.VariableSet == nil {
		merged.Variables = defaultCfg.Variables
	}
	if merged.Engine == "" {
		merged.Engine = defaultCfg.Engine
	}
	if merged.Output == "" {
		tmpDir := os.TempDir()
		tmpOutDir := filepath.Join(tmpDir, "out")
		os.MkdirAll(tmpOutDir, 0755)
		merged.Output = config.Output(filepath.Join(tmpOutDir, "output.pdf"))
	}

	// Apply defaults for Passes and UseLatexmk if not set (zero values)
	if merged.Passes < 1 {
		merged.Passes = defaultCfg.Passes
	}
	// UseLatexmk defaults to false if not explicitly set, which is fine
	// But we preserve the value from cfg if it was set

	// Log merged config values for debugging
	if iaa.logger != nil {
		iaa.logger.Info(context.Background(), "Merged AutoPDF config",
			autopdfports.NewLogField("merged_passes", merged.Passes),
			autopdfports.NewLogField("merged_use_latexmk", merged.UseLatexmk),
			autopdfports.NewLogField("merged_engine", string(merged.Engine)))
	}

	return merged
}

// createConfigFile creates a temporary config file
func (iaa *InternalApplicationAdapter) createConfigFile(cfg *config.Config, configPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create and write config file
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Encode config to YAML
	return yaml.NewEncoder(file).Encode(cfg)
}

// createDocumentService creates the internal document service
func (iaa *InternalApplicationAdapter) createDocumentService(cfg *config.Config) *documentService.DocumentService {
	// Use default working directory for backward compatibility
	return iaa.createDocumentServiceWithWorkingDir(cfg, "/tmp/autopdf", iaa.logger)
}

// createDocumentServiceWithWorkingDir creates the internal document service with custom working directory
// logger can be nil if no logging is needed
func (iaa *InternalApplicationAdapter) createDocumentServiceWithWorkingDir(cfg *config.Config, workingDir string, logger autopdfports.Logger) *documentService.DocumentService {
	// Create infrastructure adapters (DIP: Application depends on abstractions)
	fileSystem := infraadapters.NewOSFileSystem()
	executor := infraadapters.NewOSCommandExecutor()

	// Create adapters using the internal application layer
	templateAdapter := template.NewTemplateProcessorAdapter(cfg)

	// Select LaTeX compiler based on UseLatexmk flag
	var latexAdapter autopdfports.LaTeXCompiler
	if cfg.UseLatexmk {
		// Use latexmk adapter with logger for transparency
		// Log selection for debugging
		if logger != nil {
			logger.Info(context.Background(), "Using latexmk adapter for multi-pass compilation",
				autopdfports.NewLogField("passes", cfg.Passes),
				autopdfports.NewLogField("engine", string(cfg.Engine)),
				autopdfports.NewLogField("working_dir", workingDir))
		}
		latexAdapter = latex.NewLatexmkCompilerAdapterWithLogger(executor, fileSystem, logger)
	} else {
		// Use regular LaTeX adapter (manual passes)
		// Log selection for debugging
		if logger != nil {
			logger.Info(context.Background(), "Using manual LaTeX adapter (non-latexmk)",
				autopdfports.NewLogField("passes", cfg.Passes),
				autopdfports.NewLogField("engine", string(cfg.Engine)),
				autopdfports.NewLogField("working_dir", workingDir))
		}
		latexAdapter = latex.NewLaTeXCompilerAdapterWithWorkingDir(cfg, fileSystem, executor, workingDir)
	}

	converterAdapter := converter.NewConverterAdapter(cfg)
	cleanerAdapter := cleaner.NewCleanerAdapter()

	// Create document service
	return &documentService.DocumentService{
		TemplateProcessor: templateAdapter,
		LaTeXCompiler:     latexAdapter,
		Converter:         converterAdapter,
		Cleaner:           cleanerAdapter,
		PathOps:           infraadapters.NewOSPathOperations(),
		FileSystem:        infraadapters.NewOSFileSystem(),
		ErrorFactory:      errors.NewDomainErrorFactory(nil),
	}
}
