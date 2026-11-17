// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"time"
)

// TemplateProcessor processes templates with variables
// Pure transport types - no domain dependencies
type TemplateProcessor interface {
	Process(ctx context.Context, templatePath string, variables map[string]string) (string, error)
}

// DocumentCompiler compiles rendered document source to a PDF and returns the
// PDF path. It is backend-neutral: LaTeX, Typst and any later backend implement
// it, and every implementation must pass compilertest.Run.
//
// Contract:
//   - success: a non-empty PDF exists at the returned path, which equals
//     opts.OutputPath when OutputPath is set;
//   - failure: an error is returned and no PDF is left at the output path;
//   - a cancelled ctx stops the compile and the error matches ctx.Err();
//   - the staged source file never collides with the source template and is
//     removed unless opts.Debug;
//   - a missing toolchain yields a *ToolchainMissingError naming the binary.
type DocumentCompiler interface {
	Compile(ctx context.Context, content string, opts CompileOptions) (string, error)
}

// LaTeXCompiler is the former name of DocumentCompiler.
//
// Deprecated: use DocumentCompiler.
type LaTeXCompiler = DocumentCompiler

// BackendOptions is a backend-specific option value carried by CompileOptions.
// Each backend defines its own option type and reads only that type; a value
// meant for another backend is refused with a *BackendOptionsMismatchError.
type BackendOptions interface {
	// BackendName names the backend these options belong to.
	BackendName() string
}

// CompileOptions holds the backend-neutral compile parameters.
// Backend-specific knobs travel in Backend (e.g. LaTeXOptions).
type CompileOptions struct {
	Engine     string // engine name, e.g. pdflatex, xelatex, lualatex
	OutputPath string
	WorkingDir string
	JobName    string
	Cleanup    bool           // Whether to cleanup aux files
	Debug      bool           // Whether debug mode is enabled (keeps the staged file)
	Backend    BackendOptions // nil means the backend's defaults
}

// NewCompileOptions creates a CompileOptions with defaults and no backend options.
func NewCompileOptions(engine, outputPath, workingDir string) CompileOptions {
	return CompileOptions{
		Engine:     engine,
		OutputPath: outputPath,
		WorkingDir: workingDir,
		JobName:    "document",
		Cleanup:    true,
		Debug:      false,
	}
}

// WithBackendOptions sets the backend-specific options.
func (opts CompileOptions) WithBackendOptions(backend BackendOptions) CompileOptions {
	opts.Backend = backend
	return opts
}

// LaTeXBackendName is the BackendName of LaTeXOptions.
const LaTeXBackendName = "latex"

// LaTeX pass bounds.
const (
	MinLaTeXPasses = 1
	MaxLaTeXPasses = 10
)

// LaTeXOptions are the LaTeX-only compile knobs.
type LaTeXOptions struct {
	Passes     int  // Number of compilation passes (1..10)
	UseLatexmk bool // Whether latexmk drives the compile
}

// BackendName implements BackendOptions.
func (LaTeXOptions) BackendName() string { return LaTeXBackendName }

// DefaultLaTeXOptions returns one pass without latexmk.
func DefaultLaTeXOptions() LaTeXOptions {
	return LaTeXOptions{Passes: MinLaTeXPasses}
}

// NewLaTeXOptions returns LaTeXOptions with passes clamped to [1, 10].
func NewLaTeXOptions(passes int, useLatexmk bool) LaTeXOptions {
	return LaTeXOptions{Passes: passes, UseLatexmk: useLatexmk}.normalized()
}

func (o LaTeXOptions) normalized() LaTeXOptions {
	if o.Passes < MinLaTeXPasses {
		o.Passes = MinLaTeXPasses
	}
	if o.Passes > MaxLaTeXPasses {
		o.Passes = MaxLaTeXPasses
	}
	return o
}

// LaTeXOptionsFrom reads the LaTeX options carried by opts. Nil backend options
// yield DefaultLaTeXOptions; options of another backend are refused.
func LaTeXOptionsFrom(opts CompileOptions) (LaTeXOptions, error) {
	switch b := opts.Backend.(type) {
	case nil:
		return DefaultLaTeXOptions(), nil
	case LaTeXOptions:
		return b.normalized(), nil
	case *LaTeXOptions:
		if b == nil {
			return DefaultLaTeXOptions(), nil
		}
		return b.normalized(), nil
	default:
		return LaTeXOptions{}, &BackendOptionsMismatchError{Want: LaTeXBackendName, Got: b.BackendName()}
	}
}

// ErrBackendOptionsMismatch is matched (errors.Is) by every *BackendOptionsMismatchError.
var ErrBackendOptionsMismatch = errors.New("backend options mismatch")

// BackendOptionsMismatchError reports backend options handed to a compiler of
// a different backend.
type BackendOptionsMismatchError struct {
	Want string // backend the compiler implements
	Got  string // backend the options belong to
}

func (e *BackendOptionsMismatchError) Error() string {
	return fmt.Sprintf("backend options mismatch: %s compiler was given %s options", e.Want, e.Got)
}

// Is makes errors.Is(err, ErrBackendOptionsMismatch) hold.
func (e *BackendOptionsMismatchError) Is(target error) bool {
	return target == ErrBackendOptionsMismatch
}

// ErrToolchainMissing is matched (errors.Is) by every *ToolchainMissingError.
var ErrToolchainMissing = errors.New("toolchain missing")

// ToolchainMissingError reports that a backend's binary is not available.
// CommandExecutor implementations return it when the executable cannot be
// found; compilers return it when their toolchain probe fails.
type ToolchainMissingError struct {
	Binary string // the executable that is missing, e.g. pdflatex, typst
	Err    error  // underlying cause, may be nil
}

func (e *ToolchainMissingError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("toolchain missing: %s not found: %v", e.Binary, e.Err)
	}
	return fmt.Sprintf("toolchain missing: %s not found", e.Binary)
}

// Unwrap returns the underlying cause.
func (e *ToolchainMissingError) Unwrap() error { return e.Err }

// Is makes errors.Is(err, ErrToolchainMissing) hold.
func (e *ToolchainMissingError) Is(target error) bool {
	return target == ErrToolchainMissing
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

// NoOpLogger is the null-object implementation of Logger. It discards every
// message. Callers use it in place of a nil Logger so that an unconfigured
// dependency degrades to silence rather than to a nil-interface panic.
type NoOpLogger struct{}

// NewNoOpLogger returns a Logger that discards all messages.
func NewNoOpLogger() Logger { return &NoOpLogger{} }

// Debug discards the message.
func (n *NoOpLogger) Debug(ctx context.Context, msg string, fields ...LogField) {}

// Info discards the message.
func (n *NoOpLogger) Info(ctx context.Context, msg string, fields ...LogField) {}

// Warn discards the message.
func (n *NoOpLogger) Warn(ctx context.Context, msg string, fields ...LogField) {}

// Error discards the message.
func (n *NoOpLogger) Error(ctx context.Context, msg string, fields ...LogField) {}
