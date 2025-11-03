// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"path/filepath"

	"github.com/BuddhiLW/AutoPDF/configs"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/services/document"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common"
	argsPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/args"
	configPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/config"
	resultPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/result"
	wiringPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/wiring"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/convert"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/options/config"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/options/multiple"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/options/watch"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
)

// BuildServiceCmd is the thin CLI layer that delegates to the application service
var BuildServiceCmd = &bonzai.Cmd{
	Name:    `build`,
	Alias:   `b`,
	Short:   `process template and compile to PDF`,
	Usage:   `TEMPLATE [CONFIG] [OPTIONS...]`,
	MinArgs: 1,
	MaxArgs: 10, // Allow up to 10 arguments for multiple options
	Long: `
The build command processes a template file using variables from a configuration,
compiles the processed template to LaTeX, and produces a PDF output.

This version uses the application service layer for better separation of concerns.

If no configuration file is provided, it will look for autopdf.yaml in the current directory.

Available options:
- clean: Remove auxiliary LaTeX files after compilation
- verbose: Enable verbose logging
- debug: Enable debug information output
- force: Force operations (overwrite existing files)

Examples:
  autopdf build template.tex
  autopdf build template.tex config.yaml
  autopdf build template.tex config.yaml clean
  autopdf build template.tex clean verbose debug
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
		convert.ConvertServiceCmd,
		config.ConfigServiceCmd,
	},
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		// Create standardized logger and context
		ctx, logger := common.CreateStandardLoggerContext()
		defer logger.Sync()

		// Execute the streamlined build process
		return executeBuildProcess(ctx, args)
	},
}

// executeBuildProcess orchestrates the entire build process with minimal logging overhead
func executeBuildProcess(ctx context.Context, args []string) error {
	// Parse arguments with logging
	argsParser := argsPkg.NewArgsParser()
	buildArgs, err := argsParser.ParseBuildArgsWithLogging(ctx, args)
	if err != nil {
		return err
	}

	// If watch is enabled, delegate to watch mode
	if buildArgs.Options.Watch.Enabled {
		return executeWatchedBuild(ctx, buildArgs)
	}

	// Normal build flow
	return executeSingleBuild(ctx, buildArgs)
}

// executeSingleBuild performs a single build operation
func executeSingleBuild(ctx context.Context, buildArgs *argsPkg.BuildArgs) error {
	// Resolve and load configuration with logging
	configResolver := configPkg.NewConfigResolver()
	cfg, err := configResolver.LoadConfigWithLogging(ctx, buildArgs.TemplateFile, buildArgs.ConfigFile)
	if err != nil {
		return err
	}

	// Build and execute with logging
	// Use template's directory as working directory for CLI to find assets (.cls files, images)
	templateDir := filepath.Dir(cfg.Template.String())
	serviceBuilder := wiringPkg.NewServiceBuilder()
	svc := serviceBuilder.BuildDocumentServiceWithWorkingDir(cfg, templateDir)
	req := serviceBuilder.BuildRequest(buildArgs, cfg)

	result, err := svc.Build(ctx, req)
	if err != nil {
		return configs.BuildError
	}

	// Handle result and delegation
	resultHandler := resultPkg.NewResultHandler()
	if err := resultHandler.HandleBuildResult(result); err != nil {
		return err
	}

	// Handle delegation if needed
	return handleDelegation(ctx, buildArgs, result)
}

// executeWatchedBuild performs build in watch mode
func executeWatchedBuild(ctx context.Context, buildArgs *argsPkg.BuildArgs) error {
	logger := configs.GetLoggerFromContext(ctx)
	logger.InfoWithFields("Starting watched build mode",
		"template", buildArgs.TemplateFile,
		"config", buildArgs.ConfigFile,
		"interval", buildArgs.Options.Watch.Interval,
	)

	// For now, delegate to the standalone watch command
	// This maintains compatibility while we could enhance it later
	watchArgs := []string{buildArgs.TemplateFile}
	if buildArgs.ConfigFile != "" {
		watchArgs = append(watchArgs, buildArgs.ConfigFile)
	}

	// Pass BuildOptions directly to watch service
	// This preserves verbose level, debug settings, and other options
	// Following CLARITY: explicit option passing instead of string-based args
	return watch.ExecuteWatchProcessWithOptions(ctx, watchArgs, buildArgs.Options)
}

// handleDelegation manages subcommand delegation using the new flexible approach
func handleDelegation(ctx context.Context, buildArgs *argsPkg.BuildArgs, result document.BuildResult) error {
	remainingArgs := buildArgs.GetRemainingArgs()
	if len(remainingArgs) == 0 {
		return nil
	}

	subcommand := remainingArgs[0]

	// Special handling for convert command - replace subcommand with PDF path
	if subcommand == "convert" {
		remainingArgs[0] = result.PDFPath
	}

	// Create command map for delegation
	availableCommands := common.CreateCommandMap(
		convert.ConvertServiceCmd,
		config.ConfigServiceCmd,
		multiple.MultipleServiceCmd,
		watch.WatchServiceCmd,
	)

	// Delegate using the flexible approach
	return common.HandleSubcommandDelegation(ctx, subcommand, remainingArgs, availableCommands)
}
