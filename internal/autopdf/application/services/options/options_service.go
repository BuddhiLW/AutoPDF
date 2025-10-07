// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"context"
	"fmt"

	"github.com/BuddhiLW/AutoPDF/configs"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/options"
)

// OptionsService handles the execution of build options
type OptionsService struct {
	Cleaner  CleanerPort
	Logger   LoggerPort
	Debugger DebuggerPort
	Forcer   ForcerPort
}

// NewOptionsService creates a new options service
func NewOptionsService(cleaner CleanerPort, logger LoggerPort, debugger DebuggerPort, forcer ForcerPort) *OptionsService {
	return &OptionsService{
		Cleaner:  cleaner,
		Logger:   logger,
		Debugger: debugger,
		Forcer:   forcer,
	}
}

// ExecuteOptions executes all enabled options
func (os *OptionsService) ExecuteOptions(ctx context.Context, buildOptions options.BuildOptions) error {
	// Get logger from context
	logger := configs.GetLoggerFromContext(ctx)

	// Execute clean option
	if buildOptions.Clean.Enabled {
		if err := os.Cleaner.CleanAux(ctx, buildOptions.Clean.Target); err != nil {
			logger.ErrorWithFields("Failed to clean auxiliary files",
				"target", buildOptions.Clean.Target,
				"error", err)
			return fmt.Errorf("failed to clean auxiliary files: %w", err)
		}
		logger.InfoWithFields("Cleaned auxiliary files", "target", buildOptions.Clean.Target)
	}

	// Execute verbose option
	if buildOptions.Verbose.Enabled {
		os.Logger.SetVerbosity(buildOptions.Verbose.Level)
		logger.InfoWithFields("Verbose logging enabled", "level", buildOptions.Verbose.Level)
	}

	// Execute debug option
	if buildOptions.Debug.Enabled {
		os.Debugger.EnableDebug(buildOptions.Debug.Output)
		logger.InfoWithFields("Debug information enabled", "output", buildOptions.Debug.Output)
	}

	// Execute force option
	if buildOptions.Force.Enabled {
		os.Forcer.SetForceMode(buildOptions.Force.Overwrite)
		logger.InfoWithFields("Force mode enabled", "overwrite", buildOptions.Force.Overwrite)
	}

	return nil
}

// Port interfaces for dependency injection
type CleanerPort interface {
	CleanAux(ctx context.Context, target string) error
}

type LoggerPort interface {
	SetVerbosity(level int)
	Log(level int, message string, args ...interface{})
}

type DebuggerPort interface {
	EnableDebug(output string)
	Debug(message string, args ...interface{})
}

type ForcerPort interface {
	SetForceMode(overwrite bool)
	ShouldOverwrite() bool
}
