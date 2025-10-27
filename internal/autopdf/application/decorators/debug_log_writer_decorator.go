// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package decorators

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/services"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/ports"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/valueobjects"
)

// DebugLogWriterDecorator handles log file creation for debugging
type DebugLogWriterDecorator struct {
	wrapped        LaTeXCompiler
	fileSystem     ports.FileSystem
	clock          ports.Clock
	logger         ports.DebugLogger
	debugConfig    valueobjects.DebugConfig
	commandBuilder *services.LaTeXCommandBuilder
}

// NewDebugLogWriterDecorator creates a new debug log writer decorator
func NewDebugLogWriterDecorator(
	wrapped LaTeXCompiler,
	fileSystem ports.FileSystem,
	clock ports.Clock,
	logger ports.DebugLogger,
	debugConfig valueobjects.DebugConfig,
) LaTeXCompiler {
	return &DebugLogWriterDecorator{
		wrapped:        wrapped,
		fileSystem:     fileSystem,
		clock:          clock,
		logger:         logger,
		debugConfig:    debugConfig,
		commandBuilder: services.NewLaTeXCommandBuilder(fileSystem),
	}
}

// Compile compiles LaTeX content and creates debug log files if enabled
func (d *DebugLogWriterDecorator) Compile(ctx context.Context, compCtx valueobjects.CompilationContext) (string, error) {
	// Capture start time
	startTime := d.clock.Now()

	// Delegate to wrapped compiler
	result, err := d.wrapped.Compile(ctx, compCtx)

	// After: Write debug log if enabled
	if d.debugConfig.IsEnabled() {
		// Get the real command from context (set by the wrapped compiler)
		cmd, ok := valueobjects.GetCommandFromContext(ctx)
		if !ok {
			// Fallback: build command using the same command builder for logging
			// Use a placeholder temp file path for logging purposes
			placeholderTempFile := "/tmp/autopdf-processed-placeholder.tex"
			cmd = d.commandBuilder.Build(compCtx, placeholderTempFile)
		}

		d.writeLogFile(ctx, compCtx, result, err, startTime, cmd)
	}

	return result, err
}

// writeLogFile writes the compilation log to the log directory
func (d *DebugLogWriterDecorator) writeLogFile(
	ctx context.Context,
	compCtx valueobjects.CompilationContext,
	result string,
	compilationErr error,
	startTime time.Time,
	cmd ports.Command,
) {
	logDir := d.debugConfig.LogDirectory()
	if err := d.fileSystem.MkdirAll(logDir, valueobjects.GetDirectoryPermission()); err != nil {
		d.logger.Warn("Failed to create log directory %s: %v", logDir, err)
		return
	}

	// Create timestamped log file
	timestamp := d.clock.Format(d.clock.Now(), valueobjects.GetDebugFileTimestampFormat())
	logFile := filepath.Join(logDir, fmt.Sprintf("request-%s.log", timestamp))

	// Prepare log content
	logContent := d.buildLogContent(ctx, compCtx, result, compilationErr, startTime, cmd)

	if err := d.fileSystem.WriteFile(logFile, []byte(logContent), valueobjects.GetFilePermission()); err != nil {
		d.logger.Warn("Failed to create debug log file %s: %v", logFile, err)
	} else {
		d.logger.Info("Debug: Created log file at %s", logFile)
	}
}

// buildLogContent builds the log file content
func (d *DebugLogWriterDecorator) buildLogContent(
	ctx context.Context,
	compCtx valueobjects.CompilationContext,
	result string,
	compilationErr error,
	startTime time.Time,
	cmd ports.Command,
) string {
	duration := d.clock.Now().Sub(startTime)

	logContent := "=== AutoPDF LaTeX Compilation Log ===\n"
	logContent += fmt.Sprintf("Timestamp: %s\n", d.clock.Format(d.clock.Now(), valueobjects.GetLogTimestampFormat()))
	logContent += fmt.Sprintf("Duration: %v\n", duration)
	logContent += fmt.Sprintf("Engine: %s\n", compCtx.Engine())

	// Add format file info if present
	if compCtx.HasFormatFile() {
		logContent += fmt.Sprintf("Format File: %s\n", compCtx.FormatFile())
	} else {
		logContent += "Format File: (none - legacy compilation)\n"
	}

	// Add working directory if present
	if workDir := compCtx.WorkingDirectory(); workDir != "" {
		logContent += fmt.Sprintf("Working Dir: %s\n", workDir)
	}

	logContent += fmt.Sprintf("Output Path: %s\n", compCtx.OutputPath())

	// Extract command from context
	logContent += "\n=== Command ===\n"
	logContent += fmt.Sprintf("%s\n", cmd.String())

	// Add error information
	if compilationErr != nil {
		logContent += "\n=== Compilation Error ===\n"
		logContent += fmt.Sprintf("%v\n", compilationErr)
	} else {
		logContent += "\n=== Result ===\n"
		logContent += fmt.Sprintf("Success: PDF created at %s\n", result)
	}

	// Add LaTeX output
	logContent += "\n=== LaTeX Output ===\n"
	if output := cmd.GetOutput(); len(output) > 0 {
		logContent += string(output)
	} else {
		logContent += "(no output captured)\n"
	}

	return logContent
}
