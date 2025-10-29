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
	documentService "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/services/document"
	infraadapters "github.com/BuddhiLW/AutoPDF/internal/autopdf/infrastructure/adapters"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"gopkg.in/yaml.v3"
)

// InternalApplicationAdapter bridges the API layer with the internal application layer
// This adapter follows the Adapter pattern from GoF and maintains separation of concerns
type InternalApplicationAdapter struct {
	config *config.Config
}

// NewInternalApplicationAdapter creates a new adapter
func NewInternalApplicationAdapter(cfg *config.Config) *InternalApplicationAdapter {
	return &InternalApplicationAdapter{
		config: cfg,
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
		docService = iaa.createDocumentServiceWithWorkingDir(mergedCfg, workingDir)
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
		DoConvert:    mergedCfg.Conversion.Enabled,
		DoClean:      false, // Don't clean for API usage
		DebugEnabled: debugEnabled,
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
	docService := iaa.createDocumentServiceWithWorkingDir(mergedCfg, workingDir)

	// Create build request
	req := documentService.BuildRequest{
		TemplatePath: mergedCfg.Template.String(),
		ConfigPath:   configPath,
		Variables:    &mergedCfg.Variables,
		Engine:       mergedCfg.Engine.String(),
		OutputPath:   mergedCfg.Output.String(),
		DoConvert:    mergedCfg.Conversion.Enabled,
		DoClean:      false, // Don't clean for API usage
		DebugEnabled: debugEnabled,
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

	merged := &config.Config{
		Template:   cfg.Template,
		Output:     cfg.Output,
		Variables:  cfg.Variables,
		Engine:     cfg.Engine,
		Conversion: cfg.Conversion,
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
	return iaa.createDocumentServiceWithWorkingDir(cfg, "/tmp/autopdf")
}

// createDocumentServiceWithWorkingDir creates the internal document service with custom working directory
func (iaa *InternalApplicationAdapter) createDocumentServiceWithWorkingDir(cfg *config.Config, workingDir string) *documentService.DocumentService {
	// Create infrastructure adapters (DIP: Application depends on abstractions)
	fileSystem := infraadapters.NewOSFileSystem()
	executor := infraadapters.NewOSCommandExecutor()

	// Create adapters using the internal application layer
	templateAdapter := template.NewTemplateProcessorAdapter(cfg)
	latexAdapter := latex.NewLaTeXCompilerAdapterWithWorkingDir(cfg, fileSystem, executor, workingDir)
	converterAdapter := converter.NewConverterAdapter(cfg)
	cleanerAdapter := cleaner.NewCleanerAdapter()

	// Create document service
	return &documentService.DocumentService{
		TemplateProcessor: templateAdapter,
		LaTeXCompiler:     latexAdapter,
		Converter:         converterAdapter,
		Cleaner:           cleanerAdapter,
	}
}
