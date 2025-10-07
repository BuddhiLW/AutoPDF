// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package multiple

import (
	"context"
	"fmt"
	"time"

	"github.com/BuddhiLW/AutoPDF/configs"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/result_collector"
	parallelService "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/services/parallel"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/parallel"
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

	// Create domain services
	resultCollector := result_collector.NewResultCollectorAdapter()
	orchestrator := parallelService.NewParallelExecutionOrchestrator()

	// Create parallel compilation service
	parallelSvc := parallelService.NewParallelCompilationService(
		orchestrator,
		resultCollector,
		nil, // strategies would be injected here
	)

	// Create parallel compilation request
	request := parallel.ParallelCompilationRequest{
		ConfigurationFile: configFile,
		TemplateFiles:     templateFiles,
		MaxConcurrency:    4,
		Timeout:           30 * time.Second,
	}

	// Execute parallel compilation
	result, err := parallelSvc.CompileTemplates(ctx, request)
	if err != nil {
		return fmt.Errorf("parallel compilation failed: %w", err)
	}

	// Log results
	logger.InfoWithFields("Parallel compilation completed",
		"success_count", result.SuccessCount,
		"failure_count", result.FailureCount,
		"total_duration", result.TotalDuration,
	)

	// Report successful builds
	for _, buildResult := range result.SuccessfulBuilds {
		logger.InfoWithFields("Template compiled successfully",
			"template", buildResult.TemplateFile,
			"pdf", buildResult.PDFPath,
			"duration", buildResult.Duration,
		)
	}

	// Report failures
	for _, buildFailure := range result.FailedBuilds {
		logger.ErrorWithFields("Template compilation failed",
			"template", buildFailure.TemplateFile,
			"error", buildFailure.Error,
			"duration", buildFailure.Duration,
		)
	}

	// Return error if any builds failed
	if result.FailureCount > 0 {
		return fmt.Errorf("failed to compile %d templates", result.FailureCount)
	}

	logger.InfoWithFields("All templates compiled successfully", "count", result.SuccessCount)
	return nil
}
