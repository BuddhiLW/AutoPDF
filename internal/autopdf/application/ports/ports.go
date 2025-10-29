// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"io/fs"
	"time"
)

// TemplateProcessor processes templates with variables
// Pure transport types - no domain dependencies
type TemplateProcessor interface {
	Process(ctx context.Context, templatePath string, variables map[string]string) (string, error)
}

// LaTeXCompiler compiles LaTeX content to PDF
// Pure transport types - no domain dependencies
type LaTeXCompiler interface {
	Compile(ctx context.Context, content string, engine string, outputPath string, debugEnabled bool) (string, error)
}

// Converter converts PDFs to images
// Pure transport types - no domain dependencies
type Converter interface {
	ConvertToImages(ctx context.Context, pdfPath string, formats []string) ([]string, error)
}

// Cleaner removes auxiliary files
// Pure transport types - no domain dependencies
type Cleaner interface {
	Clean(ctx context.Context, pdfPath string) error
}

// FontValidator checks if required fonts are available
type FontValidator interface {
	ValidateFonts(ctx context.Context, fontNames []string) (ValidationResult, error)
}

// ValidationResult represents font validation results
type ValidationResult struct {
	AllAvailable bool
	Missing      []string
	Available    []string
}

// FileSystem abstracts file system operations (DIP for AutoPDF)
// This ensures the application layer doesn't depend on os package
type FileSystem interface {
	// File operations
	WriteFile(ctx context.Context, path string, data []byte, perm fs.FileMode) error
	ReadFile(ctx context.Context, path string) ([]byte, error)
	Remove(ctx context.Context, path string) error
	Stat(ctx context.Context, path string) (FileInfo, error)
	IsNotExist(err error) bool

	// Directory operations
	MkdirAll(ctx context.Context, path string, perm fs.FileMode) error

	// Synchronization
	Sync(ctx context.Context) error
}

// FileInfo abstracts file information
type FileInfo interface {
	Name() string
	Size() int64
	Mode() fs.FileMode
	ModTime() time.Time
	IsDir() bool
}

// CommandExecutor abstracts command execution (DIP for AutoPDF)
// This ensures the application layer doesn't depend on os/exec package
type CommandExecutor interface {
	// Execute runs a command and returns output
	Execute(ctx context.Context, cmd Command) (CommandResult, error)
}

// Command is a Value Object representing a command to execute
type Command struct {
	Executable string
	Args       []string
	Dir        string
	Env        []string
	Timeout    time.Duration
}

// NewCommand creates a validated command
func NewCommand(executable string, args []string, dir string) Command {
	return Command{
		Executable: executable,
		Args:       args,
		Dir:        dir,
		Timeout:    time.Minute * 5, // Default 5 minute timeout
	}
}

// WithTimeout sets the timeout for the command
func (c Command) WithTimeout(duration time.Duration) Command {
	c.Timeout = duration
	return c
}

// WithEnv sets environment variables for the command
func (c Command) WithEnv(env []string) Command {
	c.Env = env
	return c
}

// CommandResult represents the result of a command execution
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
	Success  bool
}

// NewCommandResult creates a command result
func NewCommandResult(stdout, stderr string, exitCode int, duration time.Duration) CommandResult {
	return CommandResult{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
		Duration: duration,
		Success:  exitCode == 0,
	}
}
