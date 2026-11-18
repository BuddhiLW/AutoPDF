// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

// Package typst is the Typst backend: a ports.DocumentCompiler that runs
// `typst compile` through the FileSystem and CommandExecutor ports.
package typst

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
)

const (
	// DefaultBinary is the typst executable the adapter runs.
	DefaultBinary = "typst"
	// EngineName is the only engine name the Typst backend accepts in
	// CompileOptions.Engine (empty is accepted too).
	EngineName = "typst"
	// DefaultWorkingDir is used when neither CompileOptions.WorkingDir nor
	// WithWorkingDir is set.
	DefaultWorkingDir = "/tmp/autopdf"
	// DefaultTimeout bounds one typst run.
	DefaultTimeout = 5 * time.Minute
)

// ErrUnsupportedEngine is returned when CompileOptions.Engine names an engine
// that is not Typst.
var ErrUnsupportedEngine = errors.New("unsupported engine for the typst backend")

// ErrInvalidInput is returned for a TypstOptions.Inputs key typst cannot
// carry (empty, or containing '=').
var ErrInvalidInput = errors.New("invalid typst input")

// TypstCompilerAdapter compiles Typst source to PDF with one `typst compile`
// run. It depends only on the FileSystem and CommandExecutor ports.
type TypstCompilerAdapter struct {
	fileSystem ports.FileSystem
	executor   ports.CommandExecutor
	binary     string
	workingDir string
	timeout    time.Duration
}

var _ ports.DocumentCompiler = (*TypstCompilerAdapter)(nil)

// Option configures a TypstCompilerAdapter.
type Option func(*TypstCompilerAdapter)

// WithWorkingDir sets the working directory used when
// CompileOptions.WorkingDir is empty.
func WithWorkingDir(dir string) Option {
	return func(a *TypstCompilerAdapter) {
		if dir != "" {
			a.workingDir = dir
		}
	}
}

// WithBinary sets the typst executable (name on PATH or a path).
func WithBinary(binary string) Option {
	return func(a *TypstCompilerAdapter) {
		if binary != "" {
			a.binary = binary
		}
	}
}

// WithTimeout bounds one typst run.
func WithTimeout(d time.Duration) Option {
	return func(a *TypstCompilerAdapter) {
		if d > 0 {
			a.timeout = d
		}
	}
}

// NewTypstCompilerAdapter creates a Typst compiler over the given ports.
func NewTypstCompilerAdapter(fileSystem ports.FileSystem, executor ports.CommandExecutor, opts ...Option) *TypstCompilerAdapter {
	a := &TypstCompilerAdapter{
		fileSystem: fileSystem,
		executor:   executor,
		binary:     DefaultBinary,
		workingDir: DefaultWorkingDir,
		timeout:    DefaultTimeout,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Binary returns the typst executable the adapter runs.
func (a *TypstCompilerAdapter) Binary() string { return a.binary }

// Compile compiles Typst content to PDF and returns the PDF path.
// On any error no PDF is left at the output path.
func (a *TypstCompilerAdapter) Compile(ctx context.Context, content string, opts ports.CompileOptions) (pdfPath string, err error) {
	workingDir := opts.WorkingDir
	if workingDir == "" {
		workingDir = a.workingDir
	}
	jobName := opts.JobName
	if jobName == "" {
		jobName = "document"
	}
	target := opts.OutputPath
	if target == "" {
		target = filepath.Join(workingDir, jobName+".pdf")
	}
	pdfPath = target
	defer func() {
		if err != nil {
			discardOutput(ctx, a.fileSystem, target)
			pdfPath = ""
		}
	}()

	if err := ctx.Err(); err != nil {
		return "", err
	}
	if opts.Engine != "" && opts.Engine != EngineName {
		return "", fmt.Errorf("%w: %q (want %q)", ErrUnsupportedEngine, opts.Engine, EngineName)
	}
	typstOpts, err := ports.TypstOptionsFrom(opts)
	if err != nil {
		return "", err
	}
	if content == "" {
		return "", errors.New("no Typst content provided")
	}

	if err := a.fileSystem.MkdirAll(ctx, workingDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create working directory: %w", err)
	}

	if err := probeToolchain(ctx, a.executor, a.binary,
		ports.NewCommand(a.binary, []string{"--version"}, "")); err != nil {
		return "", err
	}

	stagedPath := stagedSourcePath(workingDir, jobName)
	if err := a.fileSystem.WriteFile(ctx, stagedPath, []byte(content), 0644); err != nil {
		return "", err
	}
	if !opts.Debug {
		defer removeStaged(ctx, a.fileSystem, stagedPath)
	}

	if err := a.fileSystem.MkdirAll(ctx, filepath.Dir(pdfPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// A PDF left by an earlier run must not pass for this run's output.
	discardOutput(ctx, a.fileSystem, pdfPath)

	root := typstOpts.Root
	if root == "" {
		root = workingDir
	}
	args, err := CompileArgs(absPath(root), absPath(stagedPath), absPath(pdfPath), typstOpts)
	if err != nil {
		return "", err
	}
	cmd := ports.NewCommand(a.binary, args, workingDir).WithTimeout(a.timeout)

	if err := ctx.Err(); err != nil {
		return "", err
	}
	result, execErr := a.executor.Execute(ctx, cmd)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return "", fmt.Errorf("typst compilation cancelled: %w", ctxErr)
	}
	if missing, ok := asToolchainMissing(execErr); ok {
		return "", missing
	}
	if execErr != nil || result.ExitCode != 0 {
		output := joinOutput(result.Stderr, result.Stdout)
		return "", &CompileError{
			Command:     a.binary + " " + strings.Join(args, " "),
			ExitCode:    result.ExitCode,
			Diagnostics: ParseDiagnostics(output),
			Output:      output,
			Err:         execErr,
		}
	}

	if err := verifyPDF(ctx, a.fileSystem, pdfPath); err != nil {
		return "", err
	}
	return pdfPath, nil
}

// CompileArgs builds the `typst` argument list for one compile:
//
//	compile --root R --diagnostic-format short [--font-path P]... [--input k=v]... SRC OUT
//
// Font paths keep their order; inputs are sorted by key so the command line
// is deterministic.
func CompileArgs(root, source, output string, opts ports.TypstOptions) ([]string, error) {
	args := []string{"compile", "--root", root, "--diagnostic-format", "short"}
	for _, p := range opts.FontPaths {
		if p == "" {
			continue
		}
		args = append(args, "--font-path", p)
	}
	keys := make([]string, 0, len(opts.Inputs))
	for k := range opts.Inputs {
		if k == "" || strings.Contains(k, "=") {
			return nil, fmt.Errorf("%w: key %q", ErrInvalidInput, k)
		}
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		args = append(args, "--input", k+"="+opts.Inputs[k])
	}
	return append(args, source, output), nil
}

// absPath makes p absolute so typst resolves it the same way the FileSystem
// port does, whatever its own working directory is.
func absPath(p string) string {
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return p
}

func joinOutput(stderr, stdout string) string {
	switch {
	case stderr == "":
		return stdout
	case stdout == "":
		return stderr
	default:
		return strings.TrimRight(stderr, "\n") + "\n" + stdout
	}
}
