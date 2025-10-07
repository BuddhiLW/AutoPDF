// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package exclude

import (
	"context"
	"fmt"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
)

// ExcludeServiceCmd handles file exclusion patterns for watch command
var ExcludeServiceCmd = &bonzai.Cmd{
	Name:    `exclude`,
	Alias:   `x`,
	Short:   `manage file exclusion patterns for watch`,
	Usage:   `[PATTERN...]`,
	MinArgs: 0,
	MaxArgs: 20,
	Long: `
The exclude command manages file exclusion patterns for the watch command.

This allows you to specify which files should be ignored when watching for changes.
Common patterns to exclude are LaTeX auxiliary files and temporary files.

Examples:
  autopdf watch template.tex exclude "*.aux" "*.log"
  autopdf watch template.tex exclude "*.aux" "*.log" "*.out" "*.toc"
  autopdf watch template.tex exclude "*.aux" "*.log" "*.synctex.gz"
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
	},
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		// Create standardized logger and context
		ctx, logger := common.CreateStandardLoggerContext()
		defer logger.Sync()

		// Execute the exclude process
		return executeExcludeProcess(ctx, args, logger)
	},
}

// executeExcludeProcess handles exclusion pattern management
func executeExcludeProcess(ctx context.Context, args []string, logger *adapters.LoggerAdapter) error {
	// Create pattern matcher adapter
	patternMatcher := adapters.NewPatternMatcherAdapter()

	if len(args) == 0 {
		// Show current exclusion patterns
		currentPatterns := patternMatcher.GetExclusionPatterns()
		logger.InfoWithFields("Current exclusion patterns", "patterns", currentPatterns)
		return nil
	}

	// Process exclusion patterns
	patterns := args
	logger.InfoWithFields("Setting exclusion patterns", "patterns", patterns)

	// Validate patterns using domain logic
	for _, pattern := range patterns {
		if !patternMatcher.ValidatePattern(pattern) {
			logger.WarnWithFields("Invalid pattern", "pattern", pattern)
			return fmt.Errorf("invalid pattern: %s", pattern)
		}
	}

	// Configure the pattern matcher
	patternMatcher.ConfigureExclusions(patterns)

	logger.InfoWithFields("Exclusion patterns set successfully", "patterns", patterns)
	return nil
}
