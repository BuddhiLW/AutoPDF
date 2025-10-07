// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package verbose

import (
	"fmt"
	"strconv"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	persistentService "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/services/persistent"
	resultPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/result"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/wiring"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
)

// VerboseServiceCmd is the verbose command that sets verbose logging level
var VerboseServiceCmd = &bonzai.Cmd{
	Name:    `verbose`,
	Alias:   `v`,
	Short:   `set verbose logging level`,
	Usage:   `[LEVEL]`,
	MinArgs: 0,
	MaxArgs: 1,
	Long: `
The verbose command sets the verbose logging level for AutoPDF operations.
This affects the amount of detail shown during build, convert, and other operations.

Levels:
  0 - Silent (only errors)
  1 - Basic information (warnings and above)
  2 - Detailed information (info and above) - default
  3 - Debug information (debug and above)
  4 - Maximum verbosity (all logs with full introspection)

Examples:
  autopdf verbose
  autopdf verbose 3
  autopdf verbose 1
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
	},
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		// Create persistent service
		persistentSvc := persistentService.NewPersistentService()

		// Default to current level if no level specified
		level := int(persistentSvc.GetVerboseLevel())

		// Parse level from arguments if provided
		if len(args) > 0 {
			parsedLevel, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid verbose level '%s': must be a number between 0-4", args[0])
			}
			if parsedLevel < 0 || parsedLevel > 4 {
				return fmt.Errorf("verbose level must be between 0-4, got: %d", parsedLevel)
			}
			level = parsedLevel
		}

		// Create logger adapter with the specified level
		loggerAdapter := logger.NewLoggerAdapter(logger.LogLevel(level), "stdout")

		// Log the verbose level change
		loggerAdapter.InfoWithFields("Setting verbose logging level",
			"level_name", logger.LogLevel(level).String(),
			"level", level,
		)

		// Persist the verbose level
		if err := persistentSvc.SetVerboseLevel(logger.LogLevel(level)); err != nil {
			loggerAdapter.ErrorWithFields("Failed to persist verbose level",
				"error", err,
			)
			return fmt.Errorf("failed to persist verbose level: %w", err)
		}

		// Build the verbose service
		serviceBuilder := wiring.NewServiceBuilder()
		verboseService := serviceBuilder.BuildVerboseService(level, loggerAdapter)

		// Execute the verbose operation
		result, err := verboseService.SetVerboseLevel()
		if err != nil {
			loggerAdapter.ErrorWithFields("Error setting verbose level",
				"error", err,
			)
			return err
		}

		// Handle the result
		resultHandler := resultPkg.NewResultHandler()
		return resultHandler.HandleVerboseResult(result)
	},
}
