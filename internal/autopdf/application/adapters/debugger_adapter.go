// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"fmt"
	"log"
	"os"
)

// DebuggerAdapter adapts debug functionality to the application.DebuggerPort interface
type DebuggerAdapter struct {
	enabled bool
	output  string
	file    *os.File
}

// NewDebuggerAdapter creates a new debugger adapter
func NewDebuggerAdapter() *DebuggerAdapter {
	return &DebuggerAdapter{
		enabled: false,
		output:  "stdout",
	}
}

// EnableDebug enables debug output to the specified destination
func (da *DebuggerAdapter) EnableDebug(output string) {
	da.enabled = true
	da.output = output

	if output != "stdout" && output != "stderr" {
		// Open file for debug output
		file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Printf("Warning: Could not open debug output file %s: %v", output, err)
			da.output = "stdout" // Fallback to stdout
		} else {
			da.file = file
		}
	}
}

// Debug logs a debug message
func (da *DebuggerAdapter) Debug(message string, args ...interface{}) {
	if !da.enabled {
		return
	}

	formattedMessage := fmt.Sprintf("[DEBUG] %s\n", fmt.Sprintf(message, args...))

	switch da.output {
	case "stdout":
		fmt.Print(formattedMessage)
	case "stderr":
		fmt.Fprint(os.Stderr, formattedMessage)
	default:
		if da.file != nil {
			da.file.WriteString(formattedMessage)
		}
	}
}

// Close closes any open debug output file
func (da *DebuggerAdapter) Close() error {
	if da.file != nil {
		return da.file.Close()
	}
	return nil
}
