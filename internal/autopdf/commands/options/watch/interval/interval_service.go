// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package interval

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/debounce"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common"
	"github.com/rwxrob/bonzai"
	"github.com/rwxrob/bonzai/cmds/help"
	"github.com/rwxrob/bonzai/comp"
)

// IntervalServiceCmd handles debounce interval configuration for watch command
var IntervalServiceCmd = &bonzai.Cmd{
	Name:    `interval`,
	Alias:   `i`,
	Short:   `set debounce interval for watch command`,
	Usage:   `[DURATION]`,
	MinArgs: 0,
	MaxArgs: 1,
	Long: `
The interval command sets the debounce interval for the watch command.

This controls how long to wait after a file change before triggering a rebuild.
This prevents multiple rebuilds from rapid file changes.

Duration formats supported:
- 500ms, 1s, 2s, 5s
- 1000ms, 2000ms
- Default: 500ms

Examples:
  autopdf watch template.tex interval 1s
  autopdf watch template.tex interval 500ms
  autopdf watch template.tex interval 2s
`,
	Comp: comp.Cmds,
	Cmds: []*bonzai.Cmd{
		help.Cmd,
	},
	Do: func(cmd *bonzai.Cmd, args ...string) error {
		// Create standardized logger and context
		ctx, logger := common.CreateStandardLoggerContext()
		defer logger.Sync()

		// Execute the interval process
		return executeIntervalProcess(ctx, args, logger)
	},
}

// executeIntervalProcess handles interval configuration
func executeIntervalProcess(ctx context.Context, args []string, logger *logger.LoggerAdapter) error {
	// Create debounce strategy adapter
	debounceStrategy := debounce.NewDebounceStrategyAdapter(500 * time.Millisecond)

	if len(args) == 0 {
		// Show current interval
		currentInterval := debounceStrategy.GetInterval()
		logger.InfoWithFields("Current debounce interval", "interval", currentInterval)
		return nil
	}

	// Parse interval
	duration, err := parseDuration(args[0])
	if err != nil {
		logger.ErrorWithFields("Invalid duration", "input", args[0], "error", err)
		return fmt.Errorf("invalid duration: %s", args[0])
	}

	// Validate duration using domain constraints
	if duration < 100*time.Millisecond {
		logger.WarnWithFields("Duration too short", "duration", duration)
		return fmt.Errorf("duration too short, minimum is 100ms")
	}

	if duration > 10*time.Second {
		logger.WarnWithFields("Duration too long", "duration", duration)
		return fmt.Errorf("duration too long, maximum is 10s")
	}

	// Configure the debounce strategy
	debounceStrategy.ConfigureInterval(duration)

	logger.InfoWithFields("Debounce interval set", "interval", duration)
	return nil
}

// parseDuration parses a duration string
func parseDuration(input string) (time.Duration, error) {
	// Handle common formats
	input = strings.ToLower(strings.TrimSpace(input))

	// If it's just a number, assume milliseconds
	if num, err := strconv.Atoi(input); err == nil {
		return time.Duration(num) * time.Millisecond, nil
	}

	// Parse with time.ParseDuration
	return time.ParseDuration(input)
}
