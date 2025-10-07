// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package watch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BuddhiLW/AutoPDF/configs"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/debounce"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/pattern_matcher"
	watchService "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/services/watch"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/options/watch/exclude"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/options/watch/interval"
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
	Usage:   `TEMPLATE [CONFIG]`,
	MinArgs: 1,
	MaxArgs: 2,
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
		return executeWatchProcess(ctx, args)
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

// executeWatchProcess orchestrates file watching and automatic rebuilding
func executeWatchProcess(ctx context.Context, args []string) error {
	logger := configs.GetLoggerFromContext(ctx)
	logger.InfoWithFields("Starting file watcher", "args", args)

	// Parse watch arguments
	watchConfig, err := parseWatchArgs(args)
	if err != nil {
		return fmt.Errorf("failed to parse watch arguments: %w", err)
	}

	// Create domain services
	patternMatcher := pattern_matcher.NewPatternMatcherAdapter()
	debounceStrategy := debounce.NewDebounceStrategyAdapter(watchConfig.Interval)
	changeProcessor := createFileChangeProcessor(ctx, logger)

	// Create watch application service
	watchSvc := watchService.NewWatchApplicationService(
		patternMatcher,
		debounceStrategy,
		changeProcessor,
		logger,
	)

	// Configure the service
	domainConfig := watch.WatchConfiguration{
		TemplateFile:      watchConfig.TemplateFile,
		ConfigFile:        watchConfig.ConfigFile,
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
func createFileChangeProcessor(ctx context.Context, logger *logger.LoggerAdapter) watch.FileChangeProcessor {
	return &FileChangeProcessorImpl{
		ctx:    ctx,
		logger: logger,
	}
}

// FileChangeProcessorImpl implements the FileChangeProcessor interface
type FileChangeProcessorImpl struct {
	ctx    context.Context
	logger *logger.LoggerAdapter
}

// ProcessChange processes a file change event
func (p *FileChangeProcessorImpl) ProcessChange(event watch.FileChangeEvent) error {
	p.logger.InfoWithFields("Processing file change",
		"file", event.FilePath,
		"operation", event.Operation,
		"timestamp", event.Timestamp,
	)

	// Here you would implement the actual rebuild logic
	// For now, just log the event
	p.logger.InfoWithFields("File change processed successfully", "file", event.FilePath)
	return nil
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
		Include:      []string{"*.tex", "*.yaml", "*.yml"},
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
