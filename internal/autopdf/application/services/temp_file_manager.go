// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/ports"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/valueobjects"
)

// TempFileManager handles temporary file creation and cleanup
type TempFileManager struct {
	fileSystem ports.FileSystem
}

// NewTempFileManager creates a new temp file manager
func NewTempFileManager(fileSystem ports.FileSystem) *TempFileManager {
	return &TempFileManager{
		fileSystem: fileSystem,
	}
}

// Create creates a temporary file and returns its path and cleanup function
func (m *TempFileManager) Create(ctx valueobjects.CompilationContext) (string, func(), error) {
	// Use /tmp for temp files, not current working directory
	tempDir := "/tmp"

	// Generate concrete file name
	concreteFileName := m.generateFileName(ctx)
	concreteFile := filepath.Join(tempDir, concreteFileName)

	// Write content to file
	if err := m.fileSystem.WriteFile(concreteFile, []byte(ctx.Content()), valueobjects.GetFilePermission()); err != nil {
		return "", nil, err
	}

	// Return cleanup function
	cleanup := func() {
		if !ctx.DebugMode() {
			m.fileSystem.Remove(concreteFile)
		}
	}

	return concreteFile, cleanup, nil
}

// generateFileName generates an appropriate file name based on context
func (m *TempFileManager) generateFileName(ctx valueobjects.CompilationContext) string {
	// Always use unique temp file names with random suffix
	// This prevents conflicts and ensures each compilation has its own file
	randomSuffix := fmt.Sprintf("%d", time.Now().UnixNano()%1000000000)

	if ctx.DebugMode() {
		// Debug mode: use timestamp for easier identification
		timestamp := time.Now().Format("20060102-150405")
		return fmt.Sprintf("autopdf-processed-%s.tex", timestamp)
	}

	// Normal mode: use random suffix for uniqueness
	return fmt.Sprintf("autopdf-processed-%s.tex", randomSuffix)
}
