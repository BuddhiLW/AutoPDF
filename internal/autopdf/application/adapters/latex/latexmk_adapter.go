// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package latex

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
)

// LatexmkCompilerAdapter implements DocumentCompiler using latexmk
// This adapter provides multi-pass compilation with automatic dependency tracking
type LatexmkCompilerAdapter struct {
	commandExecutor ports.CommandExecutor
	fileSystem      ports.FileSystem
	logger          ports.Logger // Optional logger for transparency
}

// NewLatexmkCompilerAdapter creates a new latexmk compiler adapter
func NewLatexmkCompilerAdapter(commandExecutor ports.CommandExecutor, fileSystem ports.FileSystem) *LatexmkCompilerAdapter {
	return NewLatexmkCompilerAdapterWithLogger(commandExecutor, fileSystem, nil)
}

// NewLatexmkCompilerAdapterWithLogger creates a new latexmk compiler adapter with logger
func NewLatexmkCompilerAdapterWithLogger(commandExecutor ports.CommandExecutor, fileSystem ports.FileSystem, logger ports.Logger) *LatexmkCompilerAdapter {
	return &LatexmkCompilerAdapter{
		commandExecutor: commandExecutor,
		fileSystem:      fileSystem,
		logger:          logger,
	}
}

var _ ports.DocumentCompiler = (*LatexmkCompilerAdapter)(nil)

// Compile compiles LaTeX content using latexmk and returns the PDF path.
// On any error no PDF is left at the output path.
func (a *LatexmkCompilerAdapter) Compile(ctx context.Context, content string, opts ports.CompileOptions) (pdfPath string, err error) {
	if opts.JobName == "" {
		opts.JobName = "document"
	}
	if opts.OutputPath != "" {
		// latexmk names the PDF from -jobname; keep it equal to OutputPath's base.
		opts.JobName = strings.TrimSuffix(filepath.Base(opts.OutputPath), ".pdf")
	}
	target := filepath.Join(a.outputDir(opts), opts.JobName+".pdf")
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
	if content == "" {
		return "", errors.New("no LaTeX content provided")
	}
	if _, err := ports.LaTeXOptionsFrom(opts); err != nil {
		return "", err
	}

	engine := opts.Engine
	if engine == "" {
		engine = DefaultEngine
	}
	for _, binary := range []string{"latexmk", engine} {
		if err := probeToolchain(ctx, a.commandExecutor, binary,
			ports.NewCommand("which", []string{binary}, "")); err != nil {
			return "", err
		}
	}

	// Write content to the staged .tex file. The suffix keeps it off the source
	// template's path; latexmk names its output from -jobname, not from this file.
	texPath := stagedSourcePath(opts.WorkingDir, opts.JobName)
	err = a.fileSystem.WriteFile(ctx, texPath, []byte(content), 0644)
	if err != nil {
		a.logError(ctx, "Failed to write LaTeX content to file",
			ports.NewLogField("tex_path", texPath),
			ports.NewLogField("error", err.Error()))
		return "", fmt.Errorf("failed to write LaTeX content: %w", err)
	}
	if !opts.Debug {
		defer removeStaged(ctx, a.fileSystem, texPath)
	}

	// A PDF left by an earlier run must not pass for this run's output.
	discardOutput(ctx, a.fileSystem, pdfPath)

	// Build latexmk command
	cmd := a.buildLatexmkCommand(opts, texPath)

	// Log command being executed for transparency
	cmdString := a.formatCommand(cmd)
	a.logInfo(ctx, "Executing latexmk command",
		ports.NewLogField("command", cmdString),
		ports.NewLogField("engine", opts.Engine),
		ports.NewLogField("working_dir", opts.WorkingDir),
		ports.NewLogField("job_name", opts.JobName),
		ports.NewLogField("tex_path", texPath))

	// Execute latexmk
	result, err := a.commandExecutor.Execute(ctx, cmd)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return "", fmt.Errorf("latexmk compilation cancelled: %w", ctxErr)
	}
	if missing, ok := asToolchainMissing(err); ok {
		return "", missing
	}
	if err != nil {
		// Build log fields - handle case where result might be empty
		logFields := []ports.LogField{
			ports.NewLogField("command", cmdString),
			ports.NewLogField("working_dir", opts.WorkingDir),
			ports.NewLogField("error", err.Error()),
		}
		// Only include result fields if result is available
		if result.Stdout != "" {
			logFields = append(logFields, ports.NewLogField("stdout", result.Stdout))
		}
		if result.Stderr != "" {
			logFields = append(logFields, ports.NewLogField("stderr", result.Stderr))
		}
		if result.ExitCode != 0 {
			logFields = append(logFields, ports.NewLogField("exit_code", result.ExitCode))
		}
		if result.Duration > 0 {
			logFields = append(logFields, ports.NewLogField("duration", result.Duration.String()))
		}

		a.logError(ctx, "latexmk execution failed", logFields...)
		return "", fmt.Errorf("latexmk execution failed: %w", err)
	}

	if !result.Success {
		// Log detailed failure information
		a.logError(ctx, "latexmk compilation failed",
			ports.NewLogField("command", cmdString),
			ports.NewLogField("working_dir", opts.WorkingDir),
			ports.NewLogField("stdout", result.Stdout),
			ports.NewLogField("stderr", result.Stderr),
			ports.NewLogField("exit_code", result.ExitCode),
			ports.NewLogField("duration", result.Duration.String()))
		return "", fmt.Errorf("latexmk compilation failed: %s", result.Stderr)
	}

	// Log successful compilation
	a.logInfo(ctx, "latexmk compilation successful",
		ports.NewLogField("command", cmdString),
		ports.NewLogField("working_dir", opts.WorkingDir),
		ports.NewLogField("duration", result.Duration.String()),
		ports.NewLogField("exit_code", result.ExitCode))

	// Log stdout/stderr even on success (may contain warnings)
	if result.Stdout != "" {
		a.logDebug(ctx, "latexmk stdout",
			ports.NewLogField("stdout", result.Stdout))
	}
	if result.Stderr != "" {
		a.logWarn(ctx, "latexmk stderr (may contain warnings)",
			ports.NewLogField("stderr", result.Stderr))
	}

	if err := verifyPDF(ctx, a.fileSystem, pdfPath); err != nil {
		return "", err
	}

	// Cleanup aux files if requested and not in debug mode
	if opts.Cleanup && !opts.Debug {
		err = a.cleanupAuxFiles(ctx, opts)
		if err != nil {
			// Log cleanup error but don't fail the compilation
			a.logWarn(ctx, "Failed to cleanup auxiliary files",
				ports.NewLogField("working_dir", opts.WorkingDir),
				ports.NewLogField("error", err.Error()))
		}
	}

	return pdfPath, nil
}

// buildLatexmkCommand constructs the latexmk command with appropriate options
func (a *LatexmkCompilerAdapter) buildLatexmkCommand(opts ports.CompileOptions, texPath string) ports.Command {
	outputDir := a.outputDir(opts)

	args := []string{
		"-silent", // Suppress output except errors (for clean logs)
		"-f",      // Force compilation even if output is up to date (ensures multi-pass)
		"-interaction=nonstopmode",
		"-latexoption=-interaction=nonstopmode",
		"-jobname=" + opts.JobName,
		"-outdir=" + outputDir,
	}

	// Map engine to latexmk flag
	switch opts.Engine {
	case "pdflatex":
		args = append(args, "-pdflatex")
	case "xelatex":
		args = append(args, "-xelatex")
	case "lualatex":
		args = append(args, "-lualatex")
	default:
		// Default to xelatex if unknown engine
		args = append(args, "-xelatex")
	}

	// Add the .tex file
	args = append(args, texPath)

	return ports.NewCommand("latexmk", args, opts.WorkingDir).
		WithTimeout(5 * time.Minute)
}

// outputDir is where latexmk writes the PDF (-outdir): OutputPath's directory,
// or WorkingDir when OutputPath is empty or a bare file name.
func (a *LatexmkCompilerAdapter) outputDir(opts ports.CompileOptions) string {
	outputDir := filepath.Dir(opts.OutputPath)
	if outputDir == "." || outputDir == "" {
		outputDir = opts.WorkingDir
	}
	return outputDir
}

// cleanupAuxFiles runs latexmk -c to clean auxiliary files
func (a *LatexmkCompilerAdapter) cleanupAuxFiles(ctx context.Context, opts ports.CompileOptions) error {
	cleanupCmd := ports.NewCommand("latexmk", []string{
		"-c",
		"-outdir=" + opts.WorkingDir,
	}, opts.WorkingDir).WithTimeout(30 * time.Second)

	cleanupCmdString := a.formatCommand(cleanupCmd)
	a.logDebug(ctx, "Cleaning auxiliary files",
		ports.NewLogField("command", cleanupCmdString),
		ports.NewLogField("working_dir", opts.WorkingDir))

	result, err := a.commandExecutor.Execute(ctx, cleanupCmd)
	if err != nil {
		a.logWarn(ctx, "Cleanup command execution failed",
			ports.NewLogField("command", cleanupCmdString),
			ports.NewLogField("error", err.Error()))
		return err
	}

	if !result.Success {
		a.logWarn(ctx, "Cleanup command returned non-zero exit code",
			ports.NewLogField("command", cleanupCmdString),
			ports.NewLogField("exit_code", result.ExitCode),
			ports.NewLogField("stderr", result.Stderr))
	}

	return nil
}

// IsAvailable checks if latexmk is available in the system
func (a *LatexmkCompilerAdapter) IsAvailable(ctx context.Context) bool {
	cmd := ports.NewCommand("latexmk", []string{"--version"}, "")
	result, err := a.commandExecutor.Execute(ctx, cmd)
	return err == nil && result.Success
}

// CheckLatexmkAvailability is a package-level function to check if latexmk is available
func CheckLatexmkAvailability() bool {
	_, err := exec.LookPath("latexmk")
	return err == nil
}

// formatCommand formats a command as a string for logging
// Returns a human-readable command string like: "latexmk -xelatex -silent -f document.tex (dir: /tmp/work, timeout: 5m0s)"
func (a *LatexmkCompilerAdapter) formatCommand(cmd ports.Command) string {
	// Build readable command string with args joined by spaces
	cmdStr := cmd.Executable
	if len(cmd.Args) > 0 {
		argsStr := ""
		for i, arg := range cmd.Args {
			if i > 0 {
				argsStr += " "
			}
			argsStr += arg
		}
		cmdStr = fmt.Sprintf("%s %s", cmd.Executable, argsStr)
	}

	// Add context: working directory and timeout
	return fmt.Sprintf("%s (dir: %s, timeout: %v)", cmdStr, cmd.Dir, cmd.Timeout)
}

// logInfo logs an info message if logger is available
func (a *LatexmkCompilerAdapter) logInfo(ctx context.Context, msg string, fields ...ports.LogField) {
	if a.logger != nil {
		a.logger.Info(ctx, msg, fields...)
	}
}

// logDebug logs a debug message if logger is available
func (a *LatexmkCompilerAdapter) logDebug(ctx context.Context, msg string, fields ...ports.LogField) {
	if a.logger != nil {
		a.logger.Debug(ctx, msg, fields...)
	}
}

// logWarn logs a warning message if logger is available
func (a *LatexmkCompilerAdapter) logWarn(ctx context.Context, msg string, fields ...ports.LogField) {
	if a.logger != nil {
		a.logger.Warn(ctx, msg, fields...)
	}
}

// logError logs an error message if logger is available
func (a *LatexmkCompilerAdapter) logError(ctx context.Context, msg string, fields ...ports.LogField) {
	if a.logger != nil {
		a.logger.Error(ctx, msg, fields...)
	}
}
