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
	Compile(ctx context.Context, content string, opts CompileOptions) (string, error)
}

// CompileOptions represents LaTeX compilation parameters
// Value Object following DDD principles - immutable and validated
type CompileOptions struct {
	Engine     string // pdflatex, xelatex, lualatex
	OutputPath string
	WorkingDir string
	Passes     int  // Number of compilation passes
	UseLatexmk bool // Whether to use latexmk
	JobName    string
	Cleanup    bool // Whether to cleanup aux files
	Debug      bool // Whether debug mode is enabled
}

// NewCompileOptions creates a validated CompileOptions with defaults
func NewCompileOptions(engine, outputPath, workingDir string) CompileOptions {
	return CompileOptions{
		Engine:     engine,
		OutputPath: outputPath,
		WorkingDir: workingDir,
		Passes:     1,
		UseLatexmk: false,
		JobName:    "document",
		Cleanup:    true,
		Debug:      false,
	}
}

// WithPasses sets the number of compilation passes
func (opts CompileOptions) WithPasses(passes int) CompileOptions {
	if passes < 1 {
		passes = 1
	}
	if passes > 10 {
		passes = 10
	}
	opts.Passes = passes
	return opts
}

// WithLatexmk enables latexmk usage
func (opts CompileOptions) WithLatexmk(useLatexmk bool) CompileOptions {
	opts.UseLatexmk = useLatexmk
	return opts
}

// WithJobName sets the job name
func (opts CompileOptions) WithJobName(jobName string) CompileOptions {
	if jobName == "" {
		jobName = "document"
	}
	opts.JobName = jobName
	return opts
}

// WithCleanup sets cleanup behavior
func (opts CompileOptions) WithCleanup(cleanup bool) CompileOptions {
	opts.Cleanup = cleanup
	return opts
}

// WithDebug sets debug mode
func (opts CompileOptions) WithDebug(debug bool) CompileOptions {
	opts.Debug = debug
	return opts
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

	// Symlink operations
	Symlink(ctx context.Context, oldname, newname string) error

	// Synchronization
	Sync(ctx context.Context) error
}

// PathOperations abstracts path manipulation operations (DIP for AutoPDF)
// This ensures the application layer doesn't depend on path/filepath package
type PathOperations interface {
	// Path manipulation
	Dir(path string) string
	Base(path string) string
	Ext(path string) string
	Join(elem ...string) string
	IsAbs(path string) bool
	Clean(path string) string
}

// FileInfo abstracts file information
type FileInfo interface {
	Name() string
	Size() int64
	Mode() fs.FileMode
	ModTime() time.Time
	IsDir() bool
	Sys() interface{}
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

// RebuildService orchestrates document rebuilding from file changes
// This port abstracts rebuild orchestration following DIP principle
// Pure transport types - no domain dependencies
type RebuildService interface {
	Rebuild(ctx context.Context, templatePath, configPath string) (RebuildResult, error)
}

// RebuildResult represents the result of a rebuild operation
type RebuildResult struct {
	PDFPath string
	Success bool
	Error   error
}

// Logger provides structured logging capabilities (DIP for logging)
// This ensures adapters can log without depending on concrete logging libraries
type Logger interface {
	// Debug logs a debug-level message
	Debug(ctx context.Context, msg string, fields ...LogField)
	// Info logs an info-level message
	Info(ctx context.Context, msg string, fields ...LogField)
	// Warn logs a warning-level message
	Warn(ctx context.Context, msg string, fields ...LogField)
	// Error logs an error-level message
	Error(ctx context.Context, msg string, fields ...LogField)
}

// LogField represents a single logging field (key-value pair)
type LogField struct {
	Key   string
	Value interface{}
}

// NewLogField creates a new log field
func NewLogField(key string, value interface{}) LogField {
	return LogField{Key: key, Value: value}
}
