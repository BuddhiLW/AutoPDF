// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package factories

import (
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/latex"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/decorators"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/ports"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/valueobjects"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/infrastructure/adapters"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// LaTeXCompilerFactory creates LaTeX compilers with proper dependency injection
type LaTeXCompilerFactory struct {
	config      *config.Config
	fileSystem  ports.FileSystem
	clock       ports.Clock
	logger      ports.DebugLogger
	executor    ports.CommandExecutor
	debugConfig valueobjects.DebugConfig
}

// NewLaTeXCompilerFactory creates a new LaTeX compiler factory
func NewLaTeXCompilerFactory(
	cfg *config.Config,
	debugConfig valueobjects.DebugConfig,
) *LaTeXCompilerFactory {
	return &LaTeXCompilerFactory{
		config:      cfg,
		fileSystem:  adapters.NewOSFileSystemAdapter(),
		clock:       adapters.NewRealClockAdapter(),
		logger:      adapters.NewStdoutLoggerAdapter(),
		executor:    adapters.NewExecCommandExecutor(),
		debugConfig: debugConfig,
	}
}

// CreateCompiler creates a LaTeX compiler with optional debug decorators
func (f *LaTeXCompilerFactory) CreateCompiler() decorators.LaTeXCompiler {
	// Create base compiler
	baseCompiler := latex.NewLaTeXCompilerAdapterV2(
		f.config,
		f.fileSystem,
		f.clock,
		f.logger,
		f.executor,
	)

	// Wrap with debug decorators if enabled
	return latex.CreateCompilerWithDebugDecorators(baseCompiler, f.debugConfig)
}

// CreateCompilerWithCustomDependencies creates a compiler with custom dependencies
func (f *LaTeXCompilerFactory) CreateCompilerWithCustomDependencies(
	fileSystem ports.FileSystem,
	clock ports.Clock,
	logger ports.DebugLogger,
	executor ports.CommandExecutor,
) decorators.LaTeXCompiler {
	// Create base compiler with custom dependencies
	baseCompiler := latex.NewLaTeXCompilerAdapterV2(
		f.config,
		fileSystem,
		clock,
		logger,
		executor,
	)

	// Wrap with debug decorators if enabled
	return latex.CreateCompilerWithDebugDecorators(baseCompiler, f.debugConfig)
}
