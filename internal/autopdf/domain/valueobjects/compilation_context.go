// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package valueobjects

import (
	"errors"
	"strings"
)

// CompilationContext encapsulates compilation parameters
// Value Object: Immutable compilation configuration following DDD principles
type CompilationContext struct {
	content          string
	engine           string
	outputPath       string
	debugMode        bool
	workingDirectory string // Directory from which LaTeX will be executed
	formatFile       string // Optional precompiled format file path (.fmt)
}

// NewCompilationContext creates a new compilation context
func NewCompilationContext(content, engine, outputPath string, debug bool) (CompilationContext, error) {
	if strings.TrimSpace(content) == "" {
		return CompilationContext{}, errors.New("content cannot be empty")
	}

	if strings.TrimSpace(engine) == "" {
		engine = "pdflatex" // Default engine
	}

	return CompilationContext{
		content:          strings.TrimSpace(content),
		engine:           strings.TrimSpace(engine),
		outputPath:       strings.TrimSpace(outputPath),
		debugMode:        debug,
		workingDirectory: "", // Empty means use current directory
	}, nil
}

// NewCompilationContextWithWorkDir creates a compilation context with working directory
func NewCompilationContextWithWorkDir(content, engine, outputPath, workDir string, debug bool) (CompilationContext, error) {
	ctx, err := NewCompilationContext(content, engine, outputPath, debug)
	if err != nil {
		return CompilationContext{}, err
	}
	ctx.workingDirectory = strings.TrimSpace(workDir)
	return ctx, nil
}

// NewCompilationContextWithFormatFile creates a compilation context with format file and template file path
// This is used when the template has already been processed and written to a file
func NewCompilationContextWithFormatFile(templateFilePath, engine, outputPath, workDir, formatFile string, debug bool) (CompilationContext, error) {
	if strings.TrimSpace(templateFilePath) == "" {
		return CompilationContext{}, errors.New("template file path cannot be empty")
	}

	if strings.TrimSpace(engine) == "" {
		engine = "pdflatex" // Default engine
	}

	return CompilationContext{
		content:          templateFilePath, // Store file path as content for format-aware compilation
		engine:           strings.TrimSpace(engine),
		outputPath:       strings.TrimSpace(outputPath),
		debugMode:        debug,
		workingDirectory: strings.TrimSpace(workDir),
		formatFile:       strings.TrimSpace(formatFile),
	}, nil
}

// Content returns the LaTeX content
func (c CompilationContext) Content() string {
	return c.content
}

// Engine returns the LaTeX engine
func (c CompilationContext) Engine() string {
	return c.engine
}

// OutputPath returns the output path
func (c CompilationContext) OutputPath() string {
	return c.outputPath
}

// DebugMode returns true if debug mode is enabled
func (c CompilationContext) DebugMode() bool {
	return c.debugMode
}

// WorkingDirectory returns the working directory for LaTeX execution
func (c CompilationContext) WorkingDirectory() string {
	return c.workingDirectory
}

// FormatFile returns the precompiled format file path
func (c CompilationContext) FormatFile() string {
	return c.formatFile
}

// HasFormatFile returns true if a format file is configured
// Predicate Method: Explicit intent for format file availability check
func (c CompilationContext) HasFormatFile() bool {
	return strings.TrimSpace(c.formatFile) != ""
}

// IsFilePath returns true if the content field contains a file path instead of LaTeX content
// This is used when the template has already been processed and written to a file
func (c CompilationContext) IsFilePath() bool {
	// If we have a format file and the content doesn't contain LaTeX commands,
	// it's likely a file path
	return c.HasFormatFile() && !strings.Contains(c.content, "\\documentclass")
}
