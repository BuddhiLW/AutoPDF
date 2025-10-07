// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package multiple

import (
	"context"
	"fmt"
	"sync"

	"github.com/BuddhiLW/AutoPDF/configs"
	services "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/services"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common"
	argsPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/args"
	configPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/config"
	resultPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/result"
	wiringPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/wiring"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
)

// MultipleServiceCmd handles multiple template compilation in parallel
var MultipleServiceCmd = &bonzai.Cmd{
	Name:    `multiple`,
	Alias:   `m`,
	Short:   `compile multiple templates in parallel`,
	Usage:   `CONFIG [TEMPLATE1] [TEMPLATE2] ... [OPTIONS...]`,
	MinArgs: 2,
	MaxArgs: 20,
	Long: `
The multiple command compiles multiple LaTeX templates in parallel using the same configuration.

This is useful for batch processing multiple documents with the same variables and settings.

Examples:
  autopdf multiple config.yaml template1.tex template2.tex
  autopdf multiple config.yaml *.tex
  autopdf multiple config.yaml template1.tex template2.tex clean verbose
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
	},
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		// Create standardized logger and context
		ctx, logger := common.CreateStandardLoggerContext()
		defer logger.Sync()

		// Execute the streamlined multiple process
		return executeMultipleProcess(ctx, args)
	},
}

// executeMultipleProcess orchestrates parallel template compilation
func executeMultipleProcess(ctx context.Context, args []string) error {
	logger := configs.GetLoggerFromContext(ctx)
	logger.InfoWithFields("Starting parallel template compilation", "args", args)

	// Parse arguments to get config file and template files
	configFile := args[0]
	templateFiles := args[1:]

	// Load configuration once for all templates
	configResolver := configPkg.NewConfigResolver()
	cfg, err := configResolver.LoadConfigWithLogging(ctx, templateFiles[0], configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create wait group for parallel execution
	var wg sync.WaitGroup
	results := make(chan services.BuildResult, len(templateFiles))
	errors := make(chan error, len(templateFiles))

	// Compile each template in parallel
	for _, templateFile := range templateFiles {
		wg.Add(1)
		go func(template string) {
			defer wg.Done()

			logger.InfoWithFields("Compiling template", "template", template)

			// Build service for this template
			serviceBuilder := wiringPkg.NewServiceBuilder()
			svc := serviceBuilder.BuildDocumentService(cfg)

			// Create build request for this template
			buildArgs := &argsPkg.BuildArgs{
				TemplateFile: template,
				ConfigFile:   configFile,
			}
			req := serviceBuilder.BuildRequest(buildArgs, cfg)

			// Execute build
			result, err := svc.Build(ctx, req)
			if err != nil {
				errors <- fmt.Errorf("failed to build %s: %w", template, err)
				return
			}

			results <- result
			logger.InfoWithFields("Template compiled successfully", "template", template, "pdf", result.PDFPath)
		}(templateFile)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(results)
	close(errors)

	// Collect results
	var buildResults []services.BuildResult
	var buildErrors []error

	for result := range results {
		buildResults = append(buildResults, result)
	}

	for err := range errors {
		buildErrors = append(buildErrors, err)
	}

	// Handle results
	resultHandler := resultPkg.NewResultHandler()

	// Report successful builds
	for _, result := range buildResults {
		if err := resultHandler.HandleBuildResult(result); err != nil {
			logger.WarnWithFields("Failed to handle result", "error", err)
		}
	}

	// Report errors
	if len(buildErrors) > 0 {
		logger.ErrorWithFields("Some templates failed to compile", "error_count", len(buildErrors))
		for _, err := range buildErrors {
			logger.ErrorWithFields("Build error", "error", err)
		}
		return fmt.Errorf("failed to compile %d templates", len(buildErrors))
	}

	logger.InfoWithFields("All templates compiled successfully", "count", len(buildResults))
	return nil
}
