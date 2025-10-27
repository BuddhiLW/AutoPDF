// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"fmt"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/ports"
)

// StdoutLoggerAdapter implements DebugLogger using fmt.Printf
type StdoutLoggerAdapter struct{}

// NewStdoutLoggerAdapter creates a new stdout logger adapter
func NewStdoutLoggerAdapter() ports.DebugLogger {
	return &StdoutLoggerAdapter{}
}

// Warn logs a warning message
func (l *StdoutLoggerAdapter) Warn(msg string, args ...interface{}) {
	fmt.Printf("WARNING: "+msg+"\n", args...)
}

// Info logs an info message
func (l *StdoutLoggerAdapter) Info(msg string, args ...interface{}) {
	fmt.Printf("INFO: "+msg+"\n", args...)
}
