// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"
	"log"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain"
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
func (os *OptionsService) ExecuteOptions(ctx context.Context, options domain.BuildOptions) error {
	// Execute clean option
	if options.Clean.Enabled {
		if err := os.Cleaner.CleanAux(ctx, options.Clean.Target); err != nil {
			return fmt.Errorf("failed to clean auxiliary files: %w", err)
		}
		log.Printf("Cleaned auxiliary files in: %s", options.Clean.Target)
	}

	// Execute verbose option
	if options.Verbose.Enabled {
		os.Logger.SetVerbosity(options.Verbose.Level)
		log.Printf("Verbose logging enabled at level %d", options.Verbose.Level)
	}

	// Execute debug option
	if options.Debug.Enabled {
		os.Debugger.EnableDebug(options.Debug.Output)
		log.Printf("Debug information enabled, output to: %s", options.Debug.Output)
	}

	// Execute force option
	if options.Force.Enabled {
		os.Forcer.SetForceMode(options.Force.Overwrite)
		log.Printf("Force mode enabled, overwrite: %t", options.Force.Overwrite)
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
