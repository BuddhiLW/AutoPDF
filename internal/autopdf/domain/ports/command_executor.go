// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package ports

import "context"

// CommandExecutor abstracts command execution
type CommandExecutor interface {
	Execute(ctx context.Context, cmd Command) (output []byte, err error)
}

// Command represents an executable command
type Command interface {
	String() string
	GetOutput() []byte // Capture command output for debugging
	GetError() error   // Capture execution error for debugging
}
