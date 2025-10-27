// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package latex

import (
	"context"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/decorators"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/services"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/ports"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/valueobjects"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// LaTeXCompilerAdapterV2 is the refactored LaTeX compiler adapter following CLARITY principles
type LaTeXCompilerAdapterV2 struct {
	config          *config.Config
	fileSystem      ports.FileSystem
	clock           ports.Clock
	logger          ports.DebugLogger
	commandExecutor ports.CommandExecutor

	// Micro-services
	engineValidator *services.EngineValidator
	tempFileManager *services.TempFileManager
	commandBuilder  *services.LaTeXCommandBuilder
	outputValidator *services.OutputValidator
}

// NewLaTeXCompilerAdapterV2 creates a new LaTeX compiler adapter with dependency injection
func NewLaTeXCompilerAdapterV2(
	cfg *config.Config,
	fileSystem ports.FileSystem,
	clock ports.Clock,
	logger ports.DebugLogger,
	commandExecutor ports.CommandExecutor,
) *LaTeXCompilerAdapterV2 {
	return &LaTeXCompilerAdapterV2{
		config:          cfg,
		fileSystem:      fileSystem,
		clock:           clock,
		logger:          logger,
		commandExecutor: commandExecutor,
		engineValidator: services.NewEngineValidator(),
		tempFileManager: services.NewTempFileManager(fileSystem),
		commandBuilder:  services.NewLaTeXCommandBuilder(fileSystem),
		outputValidator: services.NewOutputValidator(fileSystem),
	}
}

// Compile implements the LaTeXCompiler interface
func (lca *LaTeXCompilerAdapterV2) Compile(ctx context.Context, compCtx valueobjects.CompilationContext) (string, error) {
	return lca.CompileWithPorts(ctx, compCtx)
}

// CompileWithPorts compiles LaTeX content using the new port-based architecture
func (lca *LaTeXCompilerAdapterV2) CompileWithPorts(ctx context.Context, compCtx valueobjects.CompilationContext) (string, error) {
	// 1. Validate engine
	if err := lca.engineValidator.Validate(compCtx.Engine()); err != nil {
		return "", err
	}

	// 2. Create temp file or use existing file
	var tempFile string
	var cleanup func()
	var err error

	if compCtx.IsFilePath() {
		// Content is already a file path (processed template)
		tempFile = compCtx.Content()
		cleanup = func() {} // No cleanup needed for external files
	} else {
		// Content is LaTeX text, create temp file
		tempFile, cleanup, err = lca.tempFileManager.Create(compCtx)
		if err != nil {
			return "", err
		}
		defer cleanup()
	}

	// 3. Build command
	cmd := lca.commandBuilder.Build(compCtx, tempFile)

	// Store command in context for debug decorators
	ctx = valueobjects.WithCommand(ctx, cmd)

	// 4. Execute LaTeX compilation
	_, err = lca.commandExecutor.Execute(ctx, cmd)
	if err != nil {
		lca.logger.Warn("LaTeX compilation failed", "error", err)
		// Continue to check if PDF was created despite error
	}

	// 5. Get the actual PDF output path from context
	pdfOutputPath := compCtx.OutputPath()

	// 6. Validate the PDF output (not the temp file)
	validatedPath, err := lca.outputValidator.Validate(pdfOutputPath)
	if err != nil {
		lca.logger.Warn("PDF validation failed",
			"expected_path", pdfOutputPath,
			"temp_file", tempFile,
			"error", err)
		return "", err
	}

	lca.logger.Info("PDF compilation successful",
		"output_path", validatedPath,
		"temp_file", tempFile)

	return validatedPath, nil
}

// CreateCompilerWithDebugDecorators creates a compiler with debug decorators if enabled
func CreateCompilerWithDebugDecorators(
	baseCompiler *LaTeXCompilerAdapterV2,
	debugConfig valueobjects.DebugConfig,
) decorators.LaTeXCompiler {
	compiler := decorators.LaTeXCompiler(baseCompiler)

	// Wrap with debug decorators if enabled
	if debugConfig.IsEnabled() {
		compiler = decorators.NewDebugFileWriterDecorator(
			compiler,
			baseCompiler.fileSystem,
			baseCompiler.clock,
			baseCompiler.logger,
			debugConfig,
		)
		compiler = decorators.NewDebugLogWriterDecorator(
			compiler,
			baseCompiler.fileSystem,
			baseCompiler.clock,
			baseCompiler.logger,
			debugConfig,
		)
	}

	return compiler
}
