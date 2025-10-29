// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	application "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/ports"
)

// OSCommandExecutor implements CommandExecutor using os/exec package
// This follows the Adapter pattern to bridge infrastructure and application layer
//
// Design Principles:
// - Adapter Pattern: Bridges os/exec to application port
// - DIP: Infrastructure depends on application, not vice versa
// - Telemetry: Captures stdout/stderr for observability
type OSCommandExecutor struct{}

// NewOSCommandExecutor creates a new OS command executor
func NewOSCommandExecutor() *OSCommandExecutor {
	return &OSCommandExecutor{}
}

// Execute implements CommandExecutor interface
func (e *OSCommandExecutor) Execute(ctx context.Context, cmd application.Command) (application.CommandResult, error) {
	startTime := time.Now()

	// Create context with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, cmd.Timeout)
	defer cancel()

	// Parse the command (handle shell execution for complex commands)
	var execCmd *exec.Cmd
	if len(cmd.Args) > 0 && cmd.Args[0] == "sh" {
		// Shell command execution
		execCmd = exec.CommandContext(cmdCtx, cmd.Executable, cmd.Args...)
	} else {
		// Direct command execution
		execCmd = exec.CommandContext(cmdCtx, cmd.Executable, cmd.Args...)
	}

	execCmd.Dir = cmd.Dir

	// Set environment if provided
	if len(cmd.Env) > 0 {
		execCmd.Env = cmd.Env
	}

	// Capture both stdout and stderr
	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	// Execute command
	err := execCmd.Run()
	duration := time.Since(startTime)

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	// Create result
	result := application.NewCommandResult(
		stdout.String(),
		stderr.String(),
		exitCode,
		duration,
	)

	return result, err
}

// MockCommandExecutor implements CommandExecutor for testing
type MockCommandExecutor struct {
	results map[string]application.CommandResult
	errors  map[string]error
}

// NewMockCommandExecutor creates a new mock command executor
func NewMockCommandExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{
		results: make(map[string]application.CommandResult),
		errors:  make(map[string]error),
	}
}

// Execute implements CommandExecutor interface
func (m *MockCommandExecutor) Execute(ctx context.Context, cmd application.Command) (application.CommandResult, error) {
	// Create command signature
	signature := fmt.Sprintf("%s %s", cmd.Executable, strings.Join(cmd.Args, " "))

	// Check for predefined error
	if err, exists := m.errors[signature]; exists {
		return application.CommandResult{}, err
	}

	// Check for predefined result
	if result, exists := m.results[signature]; exists {
		return result, nil
	}

	// Default success result
	return application.NewCommandResult("", "", 0, time.Millisecond), nil
}

// SetResult configures a mock result for a command
func (m *MockCommandExecutor) SetResult(signature string, result application.CommandResult) {
	m.results[signature] = result
}

// SetError configures a mock error for a command
func (m *MockCommandExecutor) SetError(signature string, err error) {
	m.errors[signature] = err
}
