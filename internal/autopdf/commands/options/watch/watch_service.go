// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package watch

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/BuddhiLW/AutoPDF/configs"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/debounce"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/pattern_matcher"
	ports "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/ports"
	persistentService "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/services/persistent"
	watchService "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/services/watch"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common"
	argsPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/args"
	configPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/config"
	wiringPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/wiring"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/options/watch/exclude"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/options/watch/interval"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/options"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/watch"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
)

// WatchServiceCmd handles file watching and automatic rebuilding
var WatchServiceCmd = &bonzai.Cmd{
	Name:    `watch`,
	Alias:   `w`,
	Short:   `watch files and auto-rebuild on changes`,
	Usage:   `TEMPLATE [CONFIG] [OPTIONS...]`,
	MinArgs: 1,
	MaxArgs: 10, // Allow up to 10 arguments for multiple options
	Long: `
The watch command monitors template and configuration files for changes and automatically
rebuilds the PDF when modifications are detected.

This is particularly useful during development when you want to see changes immediately
without manually rebuilding each time.

Features:
- Monitors template files for changes
- Monitors configuration files for changes
- Automatic PDF regeneration on file changes
- Debounced file system events (prevents multiple rebuilds)
- Configurable watch patterns and exclusions via subcommands

Examples:
  autopdf watch template.tex
  autopdf watch template.tex config.yaml
  autopdf watch template.tex exclude "*.aux" "*.log"
  autopdf watch template.tex interval 1s
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
		exclude.ExcludeServiceCmd,
		interval.IntervalServiceCmd,
	},
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		// Create standardized logger and context
		ctx, logger := common.CreateStandardLoggerContext()
		defer logger.Sync()

		// Execute the streamlined watch process
		return ExecuteWatchProcess(ctx, args)
	},
}

// WatchConfig holds configuration for the watch command
type WatchConfig struct {
	TemplateFile string
	ConfigFile   string
	Interval     time.Duration
	Exclude      []string
	Include      []string
}

// ExecuteWatchProcess orchestrates file watching and automatic rebuilding
// This is the public API that maintains backward compatibility
func ExecuteWatchProcess(ctx context.Context, args []string) error {
	return ExecuteWatchProcessWithOptions(ctx, args, options.NewBuildOptions())
}

// ExecuteWatchProcessWithOptions orchestrates file watching with explicit BuildOptions
// Following CLARITY: explicit dependency injection of options for better control
func ExecuteWatchProcessWithOptions(ctx context.Context, args []string, buildOpts options.BuildOptions) error {
	// Create logger with options from BuildOptions
	// Following CLARITY: use options explicitly rather than relying on context
	logger := createLoggerFromOptions(buildOpts)
	ctx = context.WithValue(ctx, configs.LoggerKey, logger)

	logger.InfoWithFields("Starting file watcher", "args", args)

	// Filter out options first (they're ignored by watch but prevent errors)
	argsParser := argsPkg.NewArgsParser()
	cleanArgs, _, err := argsParser.ParseArgsWithOptions(args)
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Parse watch arguments from cleaned args
	watchConfig, err := parseWatchArgs(cleanArgs)
	if err != nil {
		return fmt.Errorf("failed to parse watch arguments: %w", err)
	}

	// Create domain services
	patternMatcher := pattern_matcher.NewPatternMatcherAdapter()
	// Configure pattern matcher with inclusion/exclusion patterns from WatchConfig
	// Following CLARITY: explicit configuration of dependencies
	patternMatcher.ConfigureInclusions(watchConfig.Include)
	patternMatcher.ConfigureExclusions(watchConfig.Exclude)
	debounceStrategy := debounce.NewDebounceStrategyAdapter(watchConfig.Interval)

	// Create rebuild service adapter following DIP
	// Following CLARITY: compose services via dependency injection
	configResolver := configPkg.NewConfigResolver()
	serviceBuilder := wiringPkg.NewServiceBuilder()
	rebuildService := NewDocumentRebuildAdapter(configResolver, serviceBuilder, logger)

	// Resolve absolute paths for template and config
	absTemplatePath, err := filepath.Abs(watchConfig.TemplateFile)
	if err != nil {
		return fmt.Errorf("failed to resolve template path: %w", err)
	}
	absConfigPath, err := filepath.Abs(watchConfig.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to resolve config path: %w", err)
	}

	changeProcessor := createFileChangeProcessor(ctx, logger, rebuildService, absTemplatePath, absConfigPath)

	// Create watch application service
	watchSvc := watchService.NewWatchApplicationService(
		patternMatcher,
		debounceStrategy,
		changeProcessor,
		logger,
	)

	// Configure the service
	// Following CLARITY: use absolute paths for consistency with FileChangeProcessorImpl
	domainConfig := watch.WatchConfiguration{
		TemplateFile:      absTemplatePath, // Use absolute path (not relative)
		ConfigFile:        absConfigPath,   // Use absolute path (not relative)
		DebounceInterval:  watchConfig.Interval,
		ExclusionPatterns: watchConfig.Exclude,
		InclusionPatterns: watchConfig.Include,
	}

	// Start watching
	if err := watchSvc.StartWatching(domainConfig); err != nil {
		return fmt.Errorf("failed to start watching: %w", err)
	}
	defer watchSvc.StopWatching()

	logger.InfoWithFields("File watcher started successfully",
		"template", watchConfig.TemplateFile,
		"config", watchConfig.ConfigFile,
		"interval", watchConfig.Interval,
	)

	// Keep the process running
	select {}
}

// createFileChangeProcessor creates a file change processor
// Following CLARITY: explicit dependencies via constructor (Dependency Injection)
// Following DIP: depends on RebuildService port (abstraction)
func createFileChangeProcessor(
	ctx context.Context,
	logger *logger.LoggerAdapter,
	rebuildService ports.RebuildService,
	templateFile, configFile string,
) watch.FileChangeProcessor {
	return &FileChangeProcessorImpl{
		ctx:            ctx,
		logger:         logger,
		rebuildService: rebuildService,
		templateFile:   templateFile,
		configFile:     configFile,
	}
}

// FileChangeProcessorImpl implements the FileChangeProcessor interface
// Following DIP: depends on RebuildService port (abstraction), not concrete implementation
// Following SRP: single responsibility - process file change events and trigger rebuilds
type FileChangeProcessorImpl struct {
	ctx            context.Context
	logger         *logger.LoggerAdapter
	rebuildService ports.RebuildService
	templateFile   string
	configFile     string
}

// ProcessChange processes a file change event and triggers rebuild if needed
// Following CLARITY: clear intent - check if rebuild needed, then trigger
func (p *FileChangeProcessorImpl) ProcessChange(event watch.FileChangeEvent) error {
	p.logger.InfoWithFields("Processing file change",
		"file", event.FilePath,
		"operation", event.Operation,
		"timestamp", event.Timestamp,
	)

	// Check if this file change should trigger a rebuild
	if !p.shouldRebuild(event) {
		p.logger.DebugWithFields("File change does not require rebuild",
			"file", event.FilePath,
		)
		return nil
	}

	// Trigger rebuild using injected RebuildService
	// Following DIP: depends on abstraction, not concrete implementation
	result, err := p.rebuildService.Rebuild(p.ctx, p.templateFile, p.configFile)
	if err != nil {
		p.logger.ErrorWithFields("Rebuild triggered but failed",
			"file", event.FilePath,
			"error", err,
			"rebuild_result", result,
		)
		return fmt.Errorf("rebuild failed: %w", err)
	}

	if !result.Success {
		p.logger.WarnWithFields("Rebuild completed with errors",
			"file", event.FilePath,
			"pdf_path", result.PDFPath,
			"error", result.Error,
		)
		return result.Error
	}

	p.logger.InfoWithFields("Rebuild completed successfully",
		"file", event.FilePath,
		"pdf_path", result.PDFPath,
	)
	return nil
}

// shouldRebuild determines if a file change event should trigger a rebuild
// Following CLARITY: explicit intent - clearly defines when to rebuild
func (p *FileChangeProcessorImpl) shouldRebuild(event watch.FileChangeEvent) bool {
	// Normalize paths for comparison (resolve to absolute paths)
	eventPath, err := filepath.Abs(event.FilePath)
	if err != nil {
		// If we can't resolve, use original path
		eventPath = event.FilePath
	}

	// Rebuild if the changed file matches template or config file exactly
	if eventPath == p.templateFile || eventPath == p.configFile {
		return true
	}

	// Rebuild if it's a watched file type (assets like .cls, images, etc.)
	// The watch service filters by inclusion patterns, so if we get here, it's a watched file
	// This ensures changes to .cls files, images, and other assets trigger rebuild
	return true
}

// CanProcess determines if this processor can handle the event
func (p *FileChangeProcessorImpl) CanProcess(event watch.FileChangeEvent) bool {
	// Only process write operations for now
	return event.Operation == watch.WriteOp
}

// parseWatchArgs parses command line arguments for watch command
func parseWatchArgs(args []string) (*WatchConfig, error) {
	config := &WatchConfig{
		TemplateFile: args[0],
		ConfigFile:   "autopdf.yaml", // Default config
		Interval:     500 * time.Millisecond,
		Exclude:      []string{"*.aux", "*.log", "*.out", "*.toc", "*.fdb_latexmk", "*.fls", "*.synctex.gz"},
		Include:      []string{"*.tex", "*.yaml", "*.yml", "*.cls", "*.png", "*.jpg", "*.jpeg", "*.pdf"},
	}

	// Parse config file if provided
	if len(args) > 1 {
		arg := args[1]
		if strings.HasSuffix(arg, ".yaml") || strings.HasSuffix(arg, ".yml") {
			config.ConfigFile = arg
		}
	}

	return config, nil
}

// createLoggerFromOptions creates a logger adapter based on BuildOptions
// Following CLARITY: explicit logger creation from options, with fallback to persistent flags
func createLoggerFromOptions(buildOpts options.BuildOptions) *logger.LoggerAdapter {
	var logLevel logger.LogLevel = logger.Detailed // Default

	// If verbose is enabled in options, use that level
	if buildOpts.Verbose.Enabled {
		logLevel = logger.LogLevel(buildOpts.Verbose.Level)
	} else {
		// Otherwise, check persistent flags
		persistentSvc := persistentService.NewPersistentService()
		logLevel = persistentSvc.GetVerboseLevel()
	}

	// Determine output destination
	output := "stdout"
	if buildOpts.Debug.Enabled {
		output = buildOpts.Debug.Output
	}

	return logger.NewLoggerAdapter(logLevel, output)
}
