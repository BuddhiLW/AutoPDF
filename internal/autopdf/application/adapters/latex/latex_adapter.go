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
}

var _ application.DocumentCompiler = (*LaTeXCompilerAdapter)(nil)

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

	baseName := strings.TrimSuffix(filepath.Base(pdfPath), ".pdf")

	for pass := 1; pass <= latexOpts.Passes; pass++ {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		args := []string{"-interaction=nonstopmode", noShellEscape, "-jobname=" + baseName}
		if outputDir != "." {
			args = append(args, "-output-directory="+outputDir)
		}
		// Relative to workingDir: openin_any=p refuses an absolute input path.
		args = append(args, filepath.Base(stagedPath))
		cmdStr := engine + " " + strings.Join(args, " ")

		cmd := application.NewCommand(engine, args, workingDir).
			WithTimeout(5 * time.Minute).
			WithEnv(sandboxEnv())

		result, execErr := lca.executor.Execute(ctx, cmd)
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", fmt.Errorf("LaTeX compilation cancelled (pass %d/%d): %w", pass, latexOpts.Passes, ctxErr)
		}
		if missing, ok := asToolchainMissing(execErr); ok {
			return "", missing
		}
		if execErr == nil && result.ExitCode != 0 {
			execErr = fmt.Errorf("exit status %d", result.ExitCode)
		}
		if execErr != nil {
			// nonstopmode LaTeX exits non-zero on recoverable errors yet still
			// writes the PDF; only a missing PDF makes the pass fail.
			if _, statErr := lca.fileSystem.Stat(ctx, pdfPath); statErr != nil {
				errorDetails := fmt.Sprintf(
					"LaTeX compilation failed (pass %d/%d):\nCommand: %s\nWorking Dir: %s\nStderr:\n%s\nStdout:\n%s",
					pass, latexOpts.Passes, cmdStr, workingDir, result.Stderr, result.Stdout,
				)
				return "", fmt.Errorf("%s: %w", errorDetails, execErr)
			}
		}

		if err := verifyPDF(ctx, lca.fileSystem, pdfPath); err != nil {
			return "", fmt.Errorf("pass %d/%d: %w", pass, latexOpts.Passes, err)
		}
	}

	return pdfPath, nil
}

// CompileLegacy provides backward compatibility with the old interface
func (lca *LaTeXCompilerAdapter) CompileLegacy(ctx context.Context, content string, engine string, outputPath string, debugEnabled bool) (string, error) {
	opts := application.NewCompileOptions(engine, outputPath, lca.workingDir).
		WithDebug(debugEnabled)

	return lca.Compile(ctx, content, opts)
}
