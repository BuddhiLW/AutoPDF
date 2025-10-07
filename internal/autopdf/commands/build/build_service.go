// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"

	"github.com/BuddhiLW/AutoPDF/configs"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters"
	argsPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/args"
	configPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/config"
	resultPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/result"
	wiringPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/wiring"
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
	},
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		// Create logger adapter with default detailed level
		logger := adapters.NewLoggerAdapter(adapters.Detailed, "stdout")
		defer logger.Sync()

		logger.InfoWithFields("Starting AutoPDF build process",
			"args", args,
		)

		// Parse arguments
		logger.Debug("Parsing command line arguments")
		argsParser := argsPkg.NewArgsParser()
		buildArgs, err := argsParser.ParseBuildArgs(args)
		if err != nil {
			logger.ErrorWithFields("Failed to parse arguments", "error", err)
			return err
		}
		logger.InfoWithFields("Arguments parsed successfully",
			"template_file", buildArgs.TemplateFile,
			"config_file", buildArgs.ConfigFile,
			"options", buildArgs.Options,
		)

		// Resolve config file
		logger.Debug("Resolving configuration file")
		configResolver := configPkg.NewConfigResolver()
		configFile, err := configResolver.ResolveConfigFile(buildArgs.TemplateFile, buildArgs.ConfigFile)
		if err != nil {
			logger.ErrorWithFields("Failed to resolve config file", "error", err)
			return err
		}
		logger.InfoWithFields("Configuration file resolved",
			"config_file", configFile,
		)

		// Load and resolve config
		logger.Debug("Loading configuration")
		cfg, err := configResolver.LoadConfig(configFile)
		if err != nil {
			logger.ErrorWithFields("Failed to load configuration", "error", err)
			return err
		}
		logger.LogConfigBuilding(configFile, cfg.Variables.Flatten())

		// Resolve template path
		logger.Debug("Resolving template path")
		err = configResolver.ResolveTemplatePath(cfg, buildArgs.TemplateFile, configFile)
		if err != nil {
			logger.ErrorWithFields("Failed to resolve template path", "error", err)
			return err
		}
		logger.LogDataMapping(cfg.Template.String(), cfg.Variables.Flatten())

		// Build the application service
		logger.Debug("Building application service")
		serviceBuilder := wiringPkg.NewServiceBuilder()
		svc := serviceBuilder.BuildDocumentService(cfg)

		// Build the request
		logger.Debug("Building service request")
		req := serviceBuilder.BuildRequest(buildArgs, cfg)
		logger.InfoWithFields("Service request built",
			"template_path", req.TemplatePath,
			"output_path", req.OutputPath,
			"engine", req.Engine,
			"do_convert", req.DoConvert,
			"do_clean", req.DoClean,
		)

		// Execute the build
		logger.Info("Executing PDF build process")
		ctx := context.Background()
		result, err := svc.Build(ctx, req)
		if err != nil {
			logger.ErrorWithFields("PDF build failed", "error", err)
			return configs.BuildError
		}

		logger.InfoWithFields("PDF build completed successfully",
			"pdf_path", result.PDFPath,
			"image_paths", result.ImagePaths,
		)

		// Handle the result
		resultHandler := resultPkg.NewResultHandler()
		return resultHandler.HandleBuildResult(result)
	},
}
