// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package wiring

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BuddhiLW/AutoPDF/internal/application"
	"github.com/BuddhiLW/AutoPDF/internal/application/adapters"
	appAdapters "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/args"
	"github.com/BuddhiLW/AutoPDF/internal/converter"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/rwxrob/bonzai/futil"
	"go.uber.org/zap"
)

// ServiceBuilder handles the construction of the application service
type ServiceBuilder struct{}

// NewServiceBuilder creates a new service builder
func NewServiceBuilder() *ServiceBuilder {
	return &ServiceBuilder{}
}

// BuildDocumentService constructs the DocumentService with all required adapters
func (sb *ServiceBuilder) BuildDocumentService(cfg *config.Config) *application.DocumentService {
	return &application.DocumentService{
		TemplateProcessor: adapters.NewTemplateProcessorAdapter(cfg),
		LaTeXCompiler:     adapters.NewLaTeXCompilerAdapter(cfg),
		Converter:         adapters.NewConverterAdapter(cfg),
		Cleaner:           adapters.NewCleanerAdapter(),
	}
}

// BuildRequest constructs a BuildRequest from the parsed arguments and config
func (sb *ServiceBuilder) BuildRequest(args *args.BuildArgs, cfg *config.Config) application.BuildRequest {
	return application.BuildRequest{
		TemplatePath: cfg.Template.String(),
		ConfigPath:   args.ConfigFile,
		Variables:    map[string]string(cfg.Variables),
		Engine:       cfg.Engine.String(),
		OutputPath:   cfg.Output.String(),
		DoConvert:    cfg.Conversion.Enabled,
		DoClean:      args.Options.Clean.Enabled,
		Conversion: application.ConversionSettings{
			Enabled: cfg.Conversion.Enabled,
			Formats: cfg.Conversion.Formats,
		},
	}
}

// NewConvertServiceBuilder creates a new convert service builder
func NewConvertServiceBuilder() *ServiceBuilder {
	return &ServiceBuilder{}
}

// BuildConverterService constructs the converter service
func (sb *ServiceBuilder) BuildConverterService(args *args.ConvertArgs) *converter.Converter {
	// Create a config with the provided formats
	cfg := &config.Config{
		Conversion: config.Conversion{
			Enabled: true,
			Formats: args.Formats,
		},
	}

	return converter.NewConverter(cfg)
}

// BuildCleanerService constructs the cleaner service
func (sb *ServiceBuilder) BuildCleanerService(directory string) *CleanerService {
	return &CleanerService{
		Directory: directory,
	}
}

// CleanerService handles cleaning of LaTeX auxiliary files
type CleanerService struct {
	Directory string
}

// Clean removes all auxiliary files in the specified directory
func (cs *CleanerService) Clean() (*CleanResult, error) {
	// Check if directory exists
	if !futil.IsDir(cs.Directory) {
		return nil, fmt.Errorf("directory does not exist: %s", cs.Directory)
	}

	var cleanedFiles []string
	var errors []string

	// Walk through directory and remove auxiliary files
	err := filepath.Walk(cs.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Remove file if it has an auxiliary extension
		if isAuxFile(info.Name()) {
			if err := os.Remove(path); err != nil {
				errors = append(errors, fmt.Sprintf("failed to remove %s: %v", path, err))
			} else {
				cleanedFiles = append(cleanedFiles, path)
				log.Printf("Removed: %s", path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error cleaning auxiliary files: %w", err)
	}

	return &CleanResult{
		Directory:    cs.Directory,
		CleanedFiles: cleanedFiles,
		Errors:       errors,
		FilesRemoved: len(cleanedFiles),
	}, nil
}

// CleanResult represents the result of a clean operation
type CleanResult struct {
	Directory    string
	CleanedFiles []string
	Errors       []string
	FilesRemoved int
}

// isAuxFile checks if a file is a LaTeX auxiliary file
func isAuxFile(filename string) bool {
	auxiliaryExtensions := []string{
		".aux", ".log", ".toc", ".lof", ".lot", ".out", ".nav", ".snm",
		".synctex.gz", ".fls", ".fdb_latexmk", ".bbl", ".blg", ".run.xml",
		".bcf", ".idx", ".ilg", ".ind", ".brf", ".vrb", ".xdv", ".dvi",
	}

	for _, ext := range auxiliaryExtensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

// BuildVerboseService constructs the verbose service
func (sb *ServiceBuilder) BuildVerboseService(level int, logger *appAdapters.LoggerAdapter) *VerboseService {
	return &VerboseService{
		Level:  level,
		Logger: logger,
	}
}

// BuildDebugService constructs the debug service
func (sb *ServiceBuilder) BuildDebugService(output string) *DebugService {
	return &DebugService{
		Output: output,
	}
}

// BuildForceService constructs the force service
func (sb *ServiceBuilder) BuildForceService(enabled bool) *ForceService {
	return &ForceService{
		Enabled: enabled,
	}
}

// VerboseService handles verbose logging configuration
type VerboseService struct {
	Level  int
	Logger *appAdapters.LoggerAdapter
}

// SetVerboseLevel sets the verbose logging level
func (vs *VerboseService) SetVerboseLevel() (*VerboseResult, error) {
	// Log the verbose level change using the logger
	if vs.Logger != nil {
		vs.Logger.Info("Verbose logging level configured",
			zap.Int("level", vs.Level),
			zap.Bool("enabled", vs.Level > 0),
		)
	}

	levelDescriptions := map[int]string{
		0: "Silent (only errors)",
		1: "Basic information (warnings and above)",
		2: "Detailed information (info and above)",
		3: "Debug information (debug and above)",
		4: "Maximum verbosity (all logs with full introspection)",
	}

	description := levelDescriptions[vs.Level]
	if description == "" {
		description = "Unknown level"
	}

	return &VerboseResult{
		Level:       vs.Level,
		Description: description,
		Enabled:     vs.Level > 0,
	}, nil
}

// VerboseResult represents the result of setting verbose level
type VerboseResult struct {
	Level       int
	Description string
	Enabled     bool
}

// DebugService handles debug output configuration
type DebugService struct {
	Output string
}

// EnableDebug enables debug output
func (ds *DebugService) EnableDebug() (*DebugResult, error) {
	// In a real implementation, this would configure debug output
	// For now, we'll just return a result indicating debug was enabled

	return &DebugResult{
		Output:  ds.Output,
		Enabled: true,
	}, nil
}

// DebugResult represents the result of enabling debug
type DebugResult struct {
	Output  string
	Enabled bool
}

// ForceService handles force operations configuration
type ForceService struct {
	Enabled bool
}

// SetForceMode sets the force mode
func (fs *ForceService) SetForceMode() (*ForceResult, error) {
	// In a real implementation, this would configure force operations
	// For now, we'll just return a result indicating force was set

	return &ForceResult{
		Enabled: fs.Enabled,
	}, nil
}

// ForceResult represents the result of setting force mode
type ForceResult struct {
	Enabled bool
}
