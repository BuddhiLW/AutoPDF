// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package latex

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
)

// StagedSuffix names the rendered file the LaTeX backend writes next to the
// source before invoking the engine. It must never collide with a source
// template's own name: the working directory is the template's directory and
// the staged file is removed after a non-debug compile. The PDF name comes
// from -jobname, so the suffix is not observable in the output.
const StagedSuffix = ".autopdf.tex"

// stagedSourcePath is the path of the staged LaTeX source for jobName.
func stagedSourcePath(workingDir, jobName string) string {
	return filepath.Join(workingDir, jobName+StagedSuffix)
}

// probeToolchain runs a probe command and reports a failed probe as a
// *ports.ToolchainMissingError naming binary. A cancelled ctx wins.
func probeToolchain(ctx context.Context, executor ports.CommandExecutor, binary string, probe ports.Command) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	result, err := executor.Execute(ctx, probe)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	if err == nil && result.ExitCode == 0 {
		return nil
	}
	var missing *ports.ToolchainMissingError
	if errors.As(err, &missing) && missing.Binary == binary {
		return missing
	}
	cause := err
	if cause == nil {
		cause = fmt.Errorf("probe %s exited with status %d", probe.Executable, result.ExitCode)
	}
	return &ports.ToolchainMissingError{Binary: binary, Err: cause}
}

// verifyPDF checks that a non-empty file exists at pdfPath.
func verifyPDF(ctx context.Context, fileSystem ports.FileSystem, pdfPath string) error {
	info, err := fileSystem.Stat(ctx, pdfPath)
	if err != nil {
		if fileSystem.IsNotExist(err) {
			return fmt.Errorf("PDF output file was not created: %s", pdfPath)
		}
		return fmt.Errorf("error checking output file %s: %w", pdfPath, err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("PDF output file was created but is empty: %s", pdfPath)
	}
	return nil
}

// discardOutput removes whatever sits at pdfPath so that a failed compile
// never leaves a PDF a caller could mistake for its output. It runs even when
// ctx is already cancelled.
func discardOutput(ctx context.Context, fileSystem ports.FileSystem, pdfPath string) {
	_ = fileSystem.Remove(context.WithoutCancel(ctx), pdfPath)
}

// removeStaged deletes the staged source; it runs even when ctx is cancelled.
func removeStaged(ctx context.Context, fileSystem ports.FileSystem, stagedPath string) {
	_ = fileSystem.Remove(context.WithoutCancel(ctx), stagedPath)
}

// asToolchainMissing returns err when it is a *ports.ToolchainMissingError.
func asToolchainMissing(err error) (*ports.ToolchainMissingError, bool) {
	var missing *ports.ToolchainMissingError
	if errors.As(err, &missing) {
		return missing, true
	}
	return nil, false
}
