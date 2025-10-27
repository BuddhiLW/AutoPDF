// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"context"
	"os/exec"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/ports"
)

// ExecCommandExecutor implements CommandExecutor using os/exec
type ExecCommandExecutor struct{}

// NewExecCommandExecutor creates a new exec command executor
func NewExecCommandExecutor() ports.CommandExecutor {
	return &ExecCommandExecutor{}
}

// Execute runs a command and returns its output
func (e *ExecCommandExecutor) Execute(ctx context.Context, cmd ports.Command) ([]byte, error) {
	execCmd := cmd.(*ExecCommand)
	output, err := execCmd.Cmd.CombinedOutput()

	// Store output and error for debug logging
	execCmd.output = output
	execCmd.execErr = err

	return output, err
}

// ExecCommand implements Command using exec.Cmd
type ExecCommand struct {
	Cmd     *exec.Cmd
	output  []byte // Captured command output
	execErr error  // Captured execution error
}

// NewExecCommand creates a new exec command
func NewExecCommand(cmd *exec.Cmd) ports.Command {
	return &ExecCommand{Cmd: cmd}
}

// NewExecCommandFromString creates a new exec command from a string
func NewExecCommandFromString(cmdStr string) ports.Command {
	return &ExecCommand{Cmd: exec.Command("sh", "-c", cmdStr)}
}

// NewExecCommandFromArgs creates a new exec command from command name and args
func NewExecCommandFromArgs(name string, args []string) ports.Command {
	return &ExecCommand{Cmd: exec.Command(name, args...)}
}

// String returns the command string representation
func (c *ExecCommand) String() string {
	return c.Cmd.String()
}

// GetOutput returns the captured command output
func (c *ExecCommand) GetOutput() []byte {
	return c.output
}

// GetError returns the captured execution error
func (c *ExecCommand) GetError() error {
	return c.execErr
}

// SetWorkingDirectory sets the working directory for a command
func SetWorkingDirectory(cmd ports.Command, workDir string) {
	if execCmd, ok := cmd.(*ExecCommand); ok {
		execCmd.Cmd.Dir = workDir
	}
}
