// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package latex

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	application "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
)

// DefaultEngine is the LaTeX engine used when CompileOptions.Engine is empty.
const DefaultEngine = "pdflatex"

// LaTeXCompilerAdapter compiles LaTeX through an engine binary
// (pdflatex, xelatex, lualatex), running LaTeXOptions.Passes passes.
// It depends only on the FileSystem and CommandExecutor ports.
type LaTeXCompilerAdapter struct {
	config     *config.Config
	fileSystem application.FileSystem
	executor   application.CommandExecutor
	workingDir string // used when CompileOptions.WorkingDir is empty
	logger     application.Logger
}

var _ application.DocumentCompiler = (*LaTeXCompilerAdapter)(nil)

// WithLogger makes the adapter report format-file cache events to logger
// (nil discards them) and returns the adapter.
func (lca *LaTeXCompilerAdapter) WithLogger(logger application.Logger) *LaTeXCompilerAdapter {
	lca.logger = logger
	return lca
}

func (lca *LaTeXCompilerAdapter) log() application.Logger {
	if lca.logger == nil {
		return application.NewNoOpLogger()
	}
	return lca.logger
}

// NewLaTeXCompilerAdapter creates a new LaTeX compiler adapter
func NewLaTeXCompilerAdapter(
	cfg *config.Config,
	fileSystem application.FileSystem,
	executor application.CommandExecutor,
) *LaTeXCompilerAdapter {
	return &LaTeXCompilerAdapter{
		config:     cfg,
		fileSystem: fileSystem,
		executor:   executor,
		workingDir: "/tmp/autopdf", // Default working directory
	}
}

// NewLaTeXCompilerAdapterWithWorkingDir creates a new LaTeX compiler adapter with custom working directory
func NewLaTeXCompilerAdapterWithWorkingDir(
	cfg *config.Config,
	fileSystem application.FileSystem,
	executor application.CommandExecutor,
	workingDir string,
) *LaTeXCompilerAdapter {
	return &LaTeXCompilerAdapter{
		config:     cfg,
		fileSystem: fileSystem,
		executor:   executor,
		workingDir: workingDir,
	}
}

// Compile compiles LaTeX content to PDF and returns the PDF path.
// On any error no PDF is left at the output path.
func (lca *LaTeXCompilerAdapter) Compile(ctx context.Context, content string, opts application.CompileOptions) (pdfPath string, err error) {
	workingDir := opts.WorkingDir
	if workingDir == "" {
		workingDir = lca.workingDir
	}
	jobName := opts.JobName
	if jobName == "" {
		jobName = "document"
	}
	engine := opts.Engine
	if engine == "" {
		engine = DefaultEngine
	}

	target := opts.OutputPath
	if target == "" {
		target = filepath.Join(workingDir, jobName+".pdf")
	}
	pdfPath = target
	defer func() {
		if err != nil {
			discardOutput(ctx, lca.fileSystem, target)
			pdfPath = ""
		}
	}()

	if err := ctx.Err(); err != nil {
		return "", err
	}
	if content == "" {
		return "", errors.New("no LaTeX content provided")
	}
	latexOpts, err := application.LaTeXOptionsFrom(opts)
	if err != nil {
		return "", err
	}

	if err := lca.fileSystem.MkdirAll(ctx, workingDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create working directory: %w", err)
	}

	if err := probeToolchain(ctx, lca.executor, engine,
		application.NewCommand("which", []string{engine}, "")); err != nil {
		return "", err
	}

	stagedPath := stagedSourcePath(workingDir, jobName)
	if err := lca.fileSystem.WriteFile(ctx, stagedPath, []byte(content), 0644); err != nil {
		return "", err
	}
	if !opts.Debug {
		defer removeStaged(ctx, lca.fileSystem, stagedPath)
	}

	outputDir := filepath.Dir(pdfPath)
	if err := lca.fileSystem.MkdirAll(ctx, outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// A PDF left by an earlier run must not pass for this run's output.
	discardOutput(ctx, lca.fileSystem, pdfPath)

	run := passRun{
		engine:     engine,
		workingDir: workingDir,
		outputDir:  outputDir,
		baseName:   strings.TrimSuffix(filepath.Base(pdfPath), ".pdf"),
		input:      filepath.Base(stagedPath),
		pdfPath:    pdfPath,
		passes:     latexOpts.Passes,
	}

	if latexOpts.FormatFile {
		run.format = lca.prepareFormat(ctx, engine, workingDir, content, opts.Debug)
	}
	err = lca.runPasses(ctx, run)
	if err != nil && run.format != "" && ctx.Err() == nil {
		if _, missing := asToolchainMissing(err); !missing {
			lca.log().Warn(ctx, "format_file: compile with the cached format failed; retrying without it",
				application.NewLogField("format", run.format),
				application.NewLogField("error", err.Error()))
			discardOutput(ctx, lca.fileSystem, pdfPath)
			failedFormat := run.format
			run.format = ""
			err = lca.runPasses(ctx, run)
			if err == nil {
				lca.discardFormat(ctx, workingDir, failedFormat)
			}
		}
	}
	if err != nil {
		return "", err
	}
	return pdfPath, nil
}

// passRun is one engine invocation sequence of a compile.
type passRun struct {
	engine     string
	workingDir string
	outputDir  string
	baseName   string
	input      string // staged source, relative to workingDir
	pdfPath    string
	passes     int
	format     string // -fmt value relative to workingDir; "" runs the engine's own format
}

func (lca *LaTeXCompilerAdapter) runPasses(ctx context.Context, run passRun) error {
	for pass := 1; pass <= run.passes; pass++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		args := []string{"-interaction=nonstopmode", noShellEscape, "-jobname=" + run.baseName}
		if run.outputDir != "." {
			args = append(args, "-output-directory="+run.outputDir)
		}
		if run.format != "" {
			args = append(args, "-fmt="+run.format)
		}
		// Relative to workingDir: openin_any=p refuses an absolute input path.
		args = append(args, run.input)
		cmdStr := run.engine + " " + strings.Join(args, " ")

		cmd := application.NewCommand(run.engine, args, run.workingDir).
			WithTimeout(5 * time.Minute).
			WithEnv(sandboxEnv())

		result, execErr := lca.executor.Execute(ctx, cmd)
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("LaTeX compilation cancelled (pass %d/%d): %w", pass, run.passes, ctxErr)
		}
		if missing, ok := asToolchainMissing(execErr); ok {
			return missing
		}
		if execErr == nil && result.ExitCode != 0 {
			execErr = fmt.Errorf("exit status %d", result.ExitCode)
		}
		if execErr != nil {
			// nonstopmode LaTeX exits non-zero on recoverable errors yet still
			// writes the PDF; only a missing PDF makes the pass fail.
			if _, statErr := lca.fileSystem.Stat(ctx, run.pdfPath); statErr != nil {
				errorDetails := fmt.Sprintf(
					"LaTeX compilation failed (pass %d/%d):\nCommand: %s\nWorking Dir: %s\nStderr:\n%s\nStdout:\n%s",
					pass, run.passes, cmdStr, run.workingDir, result.Stderr, result.Stdout,
				)
				return fmt.Errorf("%s: %w", errorDetails, execErr)
			}
		}

		if err := verifyPDF(ctx, lca.fileSystem, run.pdfPath); err != nil {
			return fmt.Errorf("pass %d/%d: %w", pass, run.passes, err)
		}
	}
	return nil
}

// CompileLegacy provides backward compatibility with the old interface
func (lca *LaTeXCompilerAdapter) CompileLegacy(ctx context.Context, content string, engine string, outputPath string, debugEnabled bool) (string, error) {
	opts := application.NewCompileOptions(engine, outputPath, lca.workingDir).
		WithDebug(debugEnabled)

	return lca.Compile(ctx, content, opts)
}
