// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package valueobjects

import (
	"context"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/ports"
)

// Context key for storing the LaTeX command for debug logging
type contextKey string

const commandContextKey contextKey = "latex_command"

// WithCommand stores a LaTeX command in the context for debug logging
func WithCommand(ctx context.Context, cmd ports.Command) context.Context {
	return context.WithValue(ctx, commandContextKey, cmd)
}

// GetCommandFromContext extracts the LaTeX command from context for debug logging
func GetCommandFromContext(ctx context.Context) (ports.Command, bool) {
	cmd, ok := ctx.Value(commandContextKey).(ports.Command)
	return cmd, ok
}
