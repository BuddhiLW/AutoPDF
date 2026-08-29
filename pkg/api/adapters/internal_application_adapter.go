// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/cleaner"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/converter"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/latex"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/template"
	autopdfports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	documentService "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/services/document"
	infraadapters "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/infrastructure/adapters"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
	errors "github.com/BuddhiLW/AutoPDF/v2/pkg/errors"
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
	return iaa.GeneratePDFContext(context.Background(), cfg, template, debugEnabled, workingDir)
}

// GeneratePDFWithWorkingDir generates a PDF using the internal application layer with custom working directory
func (iaa *InternalApplicationAdapter) GeneratePDFWithWorkingDir(cfg *config.Config, template config.Template, debugEnabled bool, workingDir string) ([]byte, map[string]string, error) {
	return iaa.GeneratePDFContext(context.Background(), cfg, template, debugEnabled, workingDir)
}

// GeneratePDFContext runs one generation in an isolated workspace and propagates ctx
// through template processing, compilation, conversion, and cleanup.
func (iaa *InternalApplicationAdapter) GeneratePDFContext(ctx context.Context, cfg *config.Config, template config.Template, debugEnabled bool, workingDir string) ([]byte, map[string]string, error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("config must not be nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	// Collect: translate compatibility inputs into one canonical configuration.
	mergedCfg := iaa.mergeConfig(ctx, cfg, template)
	// Promote: apply output conventions before crossing into effectful execution.
	requestedOutput := normalizePDFOutput(mergedCfg.Output.String())

	workspace, err := os.MkdirTemp("", "autopdf-request-")
	if err != nil {
		return nil, nil, fmt.Errorf("create request workspace: %w", err)
	}
	defer func() { _ = os.RemoveAll(workspace) }()

	assetDir := workingDir
	if assetDir == "" {
		assetDir = filepath.Dir(mergedCfg.Template.String())
	}
	stageName := "output.pdf"
	if requestedOutput != "" {
		stageName = filepath.Base(requestedOutput)
	}
	if err := linkWorkspaceAssets(assetDir, workspace, mergedCfg.Template.String(), stageName); err != nil {
		return nil, nil, err
	}

	buildCfg := *mergedCfg
	buildCfg.Output = config.Output(filepath.Join(workspace, stageName))
	configPath := filepath.Join(workspace, "config.yaml")
	if err := iaa.createConfigFile(&buildCfg, configPath); err != nil {
		return nil, nil, fmt.Errorf("create request config: %w", err)
	}

	docService := iaa.createDocumentServiceWithWorkingDir(ctx, &buildCfg, workspace, iaa.logger)
	req := documentService.BuildRequest{
		TemplatePath: buildCfg.Template.String(),
		ConfigPath:   configPath,
		Variables:    &buildCfg.Variables,
		Engine:       buildCfg.Engine.String(),
		OutputPath:   buildCfg.Output.String(),
		WorkingDir:   workspace,
		DoClean:      false,
		DebugEnabled: debugEnabled,
		Passes:       buildCfg.Passes,
		UseLatexmk:   buildCfg.UseLatexmk,
		Conversion: documentService.ConversionSettings{
			Enabled: buildCfg.Conversion.Enabled,
			Formats: append([]string(nil), buildCfg.Conversion.Formats...),
		},
	}

	// Pipeline: process, compile, and optionally convert through DocumentService.
	result, buildErr := docService.Build(ctx, req)
	if buildErr != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, nil, ctxErr
		}
		if _, statErr := os.Stat(buildCfg.Output.String()); os.IsNotExist(statErr) {
			return nil, nil, buildErr
		}
	}

	stagePDF := result.PDFPath
	if stagePDF == "" {
		stagePDF = buildCfg.Output.String()
	}
	pdfBytes, err := os.ReadFile(stagePDF)
	if err != nil {
		return nil, nil, fmt.Errorf("read generated PDF: %w", err)
	}
	if len(pdfBytes) == 0 {
		return nil, nil, fmt.Errorf("generated PDF is empty")
	}
	if len(pdfBytes) < 5 || string(pdfBytes[:5]) != "%PDF-" {
		return nil, nil, fmt.Errorf("generated file is not a valid PDF")
	}

	// Boundary: atomically publish requested artifacts; workspace files never escape.
	paths := make(map[string]string, len(result.ImagePaths))
	if requestedOutput != "" {
		if err := writeFileAtomic(ctx, requestedOutput, pdfBytes, 0644); err != nil {
			return nil, nil, fmt.Errorf("publish generated PDF: %w", err)
		}
		for index, imagePath := range result.ImagePaths {
			targetPath := publishedImagePath(stagePDF, requestedOutput, imagePath)
			if err := copyFileAtomic(ctx, imagePath, targetPath); err != nil {
				return nil, nil, fmt.Errorf("publish converted image: %w", err)
			}
			paths[fmt.Sprintf("image_%d", index)] = targetPath
		}
	}

	return pdfBytes, paths, nil
}

// mergeConfig merges the provided config with defaults and template
func (iaa *InternalApplicationAdapter) mergeConfig(ctx context.Context, cfg *config.Config, template config.Template) *config.Config {
	defaultCfg := config.GetDefaultConfig()

	// Log incoming config values for debugging
	if iaa.logger != nil {
		iaa.logger.Info(ctx, "Merging AutoPDF config",
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
	// Apply defaults for Passes and UseLatexmk if not set (zero values)
	if merged.Passes < 1 {
		merged.Passes = defaultCfg.Passes
	}
	// UseLatexmk defaults to false if not explicitly set, which is fine
	// But we preserve the value from cfg if it was set

	// Log merged config values for debugging
	if iaa.logger != nil {
		iaa.logger.Info(ctx, "Merged AutoPDF config",
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
	if err := yaml.NewEncoder(file).Encode(cfg); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

// createDocumentServiceWithWorkingDir creates the internal document service with custom working directory
// logger can be nil if no logging is needed
func (iaa *InternalApplicationAdapter) createDocumentServiceWithWorkingDir(ctx context.Context, cfg *config.Config, workingDir string, logger autopdfports.Logger) *documentService.DocumentService {
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
			logger.Info(ctx, "Using latexmk adapter for multi-pass compilation",
				autopdfports.NewLogField("passes", cfg.Passes),
				autopdfports.NewLogField("engine", string(cfg.Engine)),
				autopdfports.NewLogField("working_dir", workingDir))
		}
		latexAdapter = latex.NewLatexmkCompilerAdapterWithLogger(executor, fileSystem, logger)
	} else {
		// Use regular LaTeX adapter (manual passes)
		// Log selection for debugging
		if logger != nil {
			logger.Info(ctx, "Using manual LaTeX adapter (non-latexmk)",
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

func normalizePDFOutput(path string) string {
	if path == "" || strings.EqualFold(filepath.Ext(path), ".pdf") {
		return path
	}
	return path + ".pdf"
}

func linkWorkspaceAssets(assetDir, workspace, templatePath, stageName string) error {
	if assetDir == "" {
		return nil
	}
	assetDir, err := filepath.Abs(assetDir)
	if err != nil {
		return fmt.Errorf("resolve asset directory: %w", err)
	}
	workspace, err = filepath.Abs(workspace)
	if err != nil {
		return fmt.Errorf("resolve workspace: %w", err)
	}
	if assetDir == workspace {
		return nil
	}

	entries, err := os.ReadDir(assetDir)
	if err != nil {
		return fmt.Errorf("read asset directory: %w", err)
	}
	templatePath, err = filepath.Abs(templatePath)
	if err != nil {
		return fmt.Errorf("resolve template path: %w", err)
	}
	for _, entry := range entries {
		if entry.Name() == "config.yaml" || entry.Name() == stageName {
			continue
		}
		source := filepath.Join(assetDir, entry.Name())
		if source == templatePath {
			continue
		}
		target := filepath.Join(workspace, entry.Name())
		if err := os.Symlink(source, target); err != nil {
			return fmt.Errorf("link workspace asset %q: %w", entry.Name(), err)
		}
	}
	return nil
}

func publishedImagePath(stagePDF, requestedPDF, stageImage string) string {
	stageBase := strings.TrimSuffix(filepath.Base(stagePDF), filepath.Ext(stagePDF))
	imageExt := filepath.Ext(stageImage)
	imageBase := strings.TrimSuffix(filepath.Base(stageImage), imageExt)
	suffix := strings.TrimPrefix(imageBase, stageBase)
	requestedBase := strings.TrimSuffix(filepath.Base(requestedPDF), filepath.Ext(requestedPDF))
	return filepath.Join(filepath.Dir(requestedPDF), requestedBase+suffix+imageExt)
}

func copyFileAtomic(ctx context.Context, source, target string) error {
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return writeFileAtomic(ctx, target, data, 0644)
}

func writeFileAtomic(ctx context.Context, target string, data []byte, mode os.FileMode) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	directory := filepath.Dir(target)
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(directory, ".autopdf-publish-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(mode); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, target)
}
