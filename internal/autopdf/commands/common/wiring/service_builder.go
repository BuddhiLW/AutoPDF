// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package wiring

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BuddhiLW/AutoPDF/v2/configs"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/cleaner"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/converter"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/template"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/backend"
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	documentService "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/services/document"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/commands/common/args"
	infraadapters "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/infrastructure/adapters"

	// Legacy converter functionality now integrated into adapters
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
	apperrors "github.com/BuddhiLW/AutoPDF/v2/pkg/errors"
	"github.com/rwxrob/bonzai/futil"
	"go.uber.org/zap"
)

// ServiceBuilder handles the construction of the application service
type ServiceBuilder struct {
	registry   *backend.Registry
	fileSystem ports.FileSystem      // nil means the OS file system
	executor   ports.CommandExecutor // nil means the OS command executor
}

// WithToolchain makes the compilers this builder wires run over fileSystem and
// executor instead of the OS ones (nil keeps the OS default).
func (sb *ServiceBuilder) WithToolchain(fileSystem ports.FileSystem, executor ports.CommandExecutor) *ServiceBuilder {
	sb.fileSystem = fileSystem
	sb.executor = executor
	return sb
}

// NewServiceBuilder creates a service builder over the default backend registry.
func NewServiceBuilder() *ServiceBuilder {
	return NewServiceBuilderWithRegistry(backend.Default())
}

// NewServiceBuilderWithRegistry creates a service builder that resolves
// backends through registry (nil means backend.Default()).
func NewServiceBuilderWithRegistry(registry *backend.Registry) *ServiceBuilder {
	if registry == nil {
		registry = backend.Default()
	}
	return &ServiceBuilder{registry: registry}
}

func (sb *ServiceBuilder) backends() *backend.Registry {
	if sb == nil || sb.registry == nil {
		return backend.Default()
	}
	return sb.registry
}

// BuildDocumentService constructs the DocumentService with all required adapters
func (sb *ServiceBuilder) BuildDocumentService(cfg *config.Config) *documentService.DocumentService {
	return sb.BuildDocumentServiceWithWorkingDir(cfg, "")
}

// BuildDocumentServiceWithWorkingDir constructs the DocumentService with template directory as working dir
// This is used by CLI to ensure LaTeX can find assets (.cls files, images) in the template's directory
func (sb *ServiceBuilder) BuildDocumentServiceWithWorkingDir(cfg *config.Config, workingDir string) *documentService.DocumentService {
	// Create infrastructure adapters (DIP: Application depends on abstractions)
	var fileSystem ports.FileSystem = infraadapters.NewOSFileSystem()
	var executor ports.CommandExecutor = infraadapters.NewOSCommandExecutor()
	if sb.fileSystem != nil {
		fileSystem = sb.fileSystem
	}
	if sb.executor != nil {
		executor = sb.executor
	}

	return &documentService.DocumentService{
		TemplateProcessor: template.NewTemplateProcessorAdapter(cfg),
		Compilers: backend.CompilerSource{
			Resolver: sb.backends(),
			Deps:     backend.CompilerDeps{FileSystem: fileSystem, Executor: executor, WorkingDir: workingDir},
		},
		Converter:    converter.NewConverterAdapter(cfg),
		Cleaner:      sb.cleanerFor(cfg),
		PathOps:      infraadapters.NewOSPathOperations(),
		FileSystem:   infraadapters.NewOSFileSystem(),
		ErrorFactory: apperrors.NewDomainErrorFactory(nil),
	}
}

// cleanerFor removes only the auxiliary files of the backend cfg selects; an
// unresolvable engine falls back to every registered backend's extensions.
func (sb *ServiceBuilder) cleanerFor(cfg *config.Config) *cleaner.CleanerAdapter {
	if cfg != nil {
		if b, err := sb.backends().Resolve(cfg.Engine.String(), cfg.Template.String()); err == nil {
			return cleaner.NewCleanerAdapterWithExtensions(b.AuxExtensions)
		}
	}
	return cleaner.NewCleanerAdapterWithExtensions(sb.backends().AuxExtensions())
}

// BackendOptions resolves the backend for the config's engine and template
// and builds its options from the config's settings. An unknown engine, or a
// setting the backend cannot honour (passes > 1 or use_latexmk with typst),
// is refused.
func (sb *ServiceBuilder) BackendOptions(cfg *config.Config) (ports.BackendOptions, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config must not be nil")
	}
	b, err := sb.backends().Resolve(cfg.Engine.String(), cfg.Template.String())
	if err != nil {
		return nil, err
	}
	return b.Options(backend.Settings{Passes: cfg.Passes, UseLatexmk: cfg.UseLatexmk})
}

// BuildRequest constructs a BuildRequest from the parsed arguments and config.
// Backend options come from the backend the registry resolves for the config's
// engine and template; a configuration that backend refuses is an error.
func (sb *ServiceBuilder) BuildRequest(args *args.BuildArgs, cfg *config.Config) (documentService.BuildRequest, error) {
	backendOptions, err := sb.BackendOptions(cfg)
	if err != nil {
		return documentService.BuildRequest{}, err
	}
	return documentService.BuildRequest{
		TemplatePath: cfg.Template.String(),
		ConfigPath:   args.ConfigFile,
		Variables:    &cfg.Variables, // Use complex variables from pkg/
		Engine:       cfg.Engine.String(),
		OutputPath:   cfg.Output.String(),
		WorkingDir:   filepath.Dir(cfg.Template.String()),
		DoClean:      args.Options.Clean.Enabled,
		DebugEnabled: args.Options.Debug.Enabled, // Pass debug option for persistent concrete files
		Backend:      backendOptions,
		Conversion: documentService.ConversionSettings{
			Enabled: cfg.Conversion.Enabled,
			Formats: cfg.Conversion.Formats,
		},
	}, nil
}

// NewConvertServiceBuilder creates a new convert service builder
func NewConvertServiceBuilder() *ServiceBuilder {
	return &ServiceBuilder{}
}

// BuildConverterService constructs the converter service
func (sb *ServiceBuilder) BuildConverterService(args *args.ConvertArgs) *converter.ConverterAdapter {
	// Create a config with the provided formats
	cfg := &config.Config{
		Conversion: config.Conversion{
			Enabled: true,
			Formats: args.Formats,
		},
	}

	return converter.NewConverterAdapter(cfg)
}

// BuildCleanerService constructs the cleaner service
func (sb *ServiceBuilder) BuildCleanerService(directory string) *CleanerService {
	return &CleanerService{
		Directory:     directory,
		AuxExtensions: sb.backends().AuxExtensions(),
	}
}

// CleanerService handles cleaning of backend auxiliary files
type CleanerService struct {
	Directory     string
	AuxExtensions []string // nil means the default registry's projection
}

// Clean removes all auxiliary files in the specified directory
func (cs *CleanerService) Clean() (*CleanResult, error) {
	// Check if directory exists
	if !futil.IsDir(cs.Directory) {
		return nil, fmt.Errorf(configs.ErrDirectoryNotExists, cs.Directory)
	}

	auxExtensions := cs.AuxExtensions
	if auxExtensions == nil {
		auxExtensions = backend.Default().AuxExtensions()
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
		if isAuxFile(info.Name(), auxExtensions) {
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

// isAuxFile checks if a file carries one of the auxiliary extensions
func isAuxFile(filename string, auxiliaryExtensions []string) bool {
	for _, ext := range auxiliaryExtensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

// BuildVerboseService constructs the verbose service
func (sb *ServiceBuilder) BuildVerboseService(level int, loggerAdapter *logger.LoggerAdapter) *VerboseService {
	return &VerboseService{
		Level:  level,
		Logger: loggerAdapter,
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
	Logger *logger.LoggerAdapter
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

	levelDescriptions := configs.VerboseLevelDescriptions

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

// BuildConfigService creates a config service
func (sb *ServiceBuilder) BuildConfigService() *ConfigService {
	return &ConfigService{}
}

// ConfigService handles configuration operations
type ConfigService struct{}

// HandleConfig processes configuration operations
func (cs *ConfigService) HandleConfig(args ...string) (*ConfigResult, error) {
	// For now, just return a basic result
	// Future: implement actual config operations
	return &ConfigResult{
		ConfigFile: "",
		Valid:      true,
		Message:    "Configuration handled successfully",
	}, nil
}

// ConfigResult represents the result of config operations
type ConfigResult struct {
	ConfigFile string
	Valid      bool
	Message    string
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
