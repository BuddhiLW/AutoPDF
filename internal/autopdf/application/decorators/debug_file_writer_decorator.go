// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package decorators

import (
	"context"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/ports"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/valueobjects"
)

// LaTeXCompiler defines the interface for LaTeX compilation
type LaTeXCompiler interface {
	Compile(ctx context.Context, compCtx valueobjects.CompilationContext) (string, error)
}

// DebugFileWriterDecorator handles concrete file creation for debugging
type DebugFileWriterDecorator struct {
	wrapped     LaTeXCompiler
	fileSystem  ports.FileSystem
	clock       ports.Clock
	logger      ports.DebugLogger
	debugConfig valueobjects.DebugConfig
}

// NewDebugFileWriterDecorator creates a new debug file writer decorator
func NewDebugFileWriterDecorator(
	wrapped LaTeXCompiler,
	fileSystem ports.FileSystem,
	clock ports.Clock,
	logger ports.DebugLogger,
	debugConfig valueobjects.DebugConfig,
) LaTeXCompiler {
	return &DebugFileWriterDecorator{
		wrapped:     wrapped,
		fileSystem:  fileSystem,
		clock:       clock,
		logger:      logger,
		debugConfig: debugConfig,
	}
}

// Compile compiles LaTeX content
func (d *DebugFileWriterDecorator) Compile(ctx context.Context, compCtx valueobjects.CompilationContext) (string, error) {
	// Delegate to wrapped compiler
	return d.wrapped.Compile(ctx, compCtx)
}
