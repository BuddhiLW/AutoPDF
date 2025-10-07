// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package clean

import (
	"fmt"
	"log"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters"
	services "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/services"
	resultPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/result"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/wiring"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
)

// CleanServiceCmd is the clean command that removes LaTeX auxiliary files
var CleanServiceCmd = &bonzai.Cmd{
	Name:    `clean`,
	Alias:   `c`,
	Short:   `remove LaTeX auxiliary files`,
	Usage:   `[DIRECTORY]`,
	MinArgs: 0,
	MaxArgs: 1,
	Long: `
The clean command manages LaTeX auxiliary file cleaning with persistent settings.

Subcommands:
  on     - Enable automatic cleaning (persistent)
  off    - Disable automatic cleaning (persistent)  
  switch - Toggle automatic cleaning (persistent)
  status - Show current clean settings

Without subcommands, it performs immediate cleaning of auxiliary files.
These include .aux, .log, .toc, .lof, .lot, .out, .nav, .snm, .synctex.gz, 
.fls, .fdb_latexmk, .bbl, .blg, .run.xml, .bcf, .idx, .ilg, .ind, .brf, 
.vrb, .xdv, .dvi and other temporary files.

Examples:
  autopdf clean                    # Clean current directory immediately
  autopdf clean ./output          # Clean specific directory immediately
  autopdf clean on                # Enable persistent auto-cleaning
  autopdf clean off               # Disable persistent auto-cleaning
  autopdf clean switch            # Toggle persistent auto-cleaning
  autopdf clean status            # Show current clean settings
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
		CleanOnCmd,
		CleanOffCmd,
		CleanSwitchCmd,
		CleanStatusCmd,
	},
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		log.Println("Cleaning LaTeX auxiliary files...")

		// Determine directory to clean
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		// Build the cleaner service
		serviceBuilder := wiring.NewServiceBuilder()
		cleaner := serviceBuilder.BuildCleanerService(dir)

		// Execute the clean operation
		result, err := cleaner.Clean()
		if err != nil {
			log.Printf("Error cleaning auxiliary files: %s", err)
			return err
		}

		// Handle the result
		resultHandler := resultPkg.NewResultHandler()
		return resultHandler.HandleCleanResult(result)
	},
}

// CleanOnCmd enables persistent auto-cleaning
var CleanOnCmd = &bonzai.Cmd{
	Name:    `on`,
	Short:   `enable persistent auto-cleaning`,
	Usage:   ``,
	MinArgs: 0,
	MaxArgs: 0,
	Long: `
Enable persistent automatic cleaning of LaTeX auxiliary files.
This setting will be remembered across CLI sessions.

Examples:
  autopdf clean on
`,
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		persistentService := services.NewPersistentService()

		if err := persistentService.SetCleanEnabled(true); err != nil {
			return fmt.Errorf("failed to enable persistent cleaning: %w", err)
		}

		// Create logger for user feedback
		logger := adapters.NewLoggerAdapter(adapters.Detailed, "stdout")
		logger.Info("✅ Auto-cleaning enabled (persistent)")
		logger.Info("LaTeX auxiliary files will be automatically cleaned after compilation.")
		return nil
	},
}

// CleanOffCmd disables persistent auto-cleaning
var CleanOffCmd = &bonzai.Cmd{
	Name:    `off`,
	Short:   `disable persistent auto-cleaning`,
	Usage:   ``,
	MinArgs: 0,
	MaxArgs: 0,
	Long: `
Disable persistent automatic cleaning of LaTeX auxiliary files.
This setting will be remembered across CLI sessions.

Examples:
  autopdf clean off
`,
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		persistentService := services.NewPersistentService()

		if err := persistentService.SetCleanEnabled(false); err != nil {
			return fmt.Errorf("failed to disable persistent cleaning: %w", err)
		}

		// Create logger for user feedback
		logger := adapters.NewLoggerAdapter(adapters.Detailed, "stdout")
		logger.Info("❌ Auto-cleaning disabled (persistent)")
		logger.Info("LaTeX auxiliary files will not be automatically cleaned.")
		return nil
	},
}

// CleanSwitchCmd toggles persistent auto-cleaning
var CleanSwitchCmd = &bonzai.Cmd{
	Name:    `switch`,
	Short:   `toggle persistent auto-cleaning`,
	Usage:   ``,
	MinArgs: 0,
	MaxArgs: 0,
	Long: `
Toggle persistent automatic cleaning of LaTeX auxiliary files.
This setting will be remembered across CLI sessions.

Examples:
  autopdf clean switch
`,
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		persistentService := services.NewPersistentService()

		enabled, err := persistentService.ToggleClean()
		if err != nil {
			return fmt.Errorf("failed to toggle persistent cleaning: %w", err)
		}

		// Create logger for user feedback
		logger := adapters.NewLoggerAdapter(adapters.Detailed, "stdout")
		if enabled {
			logger.Info("✅ Auto-cleaning enabled (persistent)")
			logger.Info("LaTeX auxiliary files will be automatically cleaned after compilation.")
		} else {
			logger.Info("❌ Auto-cleaning disabled (persistent)")
			logger.Info("LaTeX auxiliary files will not be automatically cleaned.")
		}
		return nil
	},
}

// CleanStatusCmd shows current clean settings
var CleanStatusCmd = &bonzai.Cmd{
	Name:    `status`,
	Short:   `show current clean settings`,
	Usage:   ``,
	MinArgs: 0,
	MaxArgs: 0,
	Long: `
Show the current persistent clean settings and status.

Examples:
  autopdf clean status
`,
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		persistentService := services.NewPersistentService()
		status := persistentService.GetStatus()

		// Create logger for user feedback
		logger := adapters.NewLoggerAdapter(adapters.Detailed, "stdout")
		logger.Info("🧹 AutoPDF Clean Settings")
		logger.Info("=========================")

		cleanStatus := status["clean"].(map[string]interface{})
		enabled := cleanStatus["enabled"].(bool)

		if enabled {
			logger.Info("✅ Auto-cleaning: ENABLED")
			logger.Info("   LaTeX auxiliary files will be automatically cleaned after compilation.")
		} else {
			logger.Info("❌ Auto-cleaning: DISABLED")
			logger.Info("   LaTeX auxiliary files will not be automatically cleaned.")
		}

		logger.InfoWithFields("📁 Config file", "path", persistentService.GetConfigPath())

		return nil
	},
}
