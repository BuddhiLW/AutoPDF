// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package cleaner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/backend"
)

// CleanerAdapter removes the auxiliary files the registered backends produce.
type CleanerAdapter struct {
	auxExtensions []string
}

// NewCleanerAdapter creates a cleaner over the default backend registry's
// auxiliary extensions.
func NewCleanerAdapter() *CleanerAdapter {
	return NewCleanerAdapterWithExtensions(backend.Default().AuxExtensions())
}

// NewCleanerAdapterWithExtensions creates a cleaner that removes files with
// the given auxiliary extensions.
func NewCleanerAdapterWithExtensions(auxExtensions []string) *CleanerAdapter {
	return &CleanerAdapter{auxExtensions: append([]string(nil), auxExtensions...)}
}

// AuxExtensions returns the extensions this cleaner removes.
func (ca *CleanerAdapter) AuxExtensions() []string {
	return append([]string(nil), ca.auxExtensions...)
}

// Clean removes auxiliary files for a given PDF
func (ca *CleanerAdapter) Clean(ctx context.Context, pdfPath string) error {
	// Get the directory of the PDF
	dir := filepath.Dir(pdfPath)

	auxExtensions := ca.auxExtensions

	// Get the base name without extension
	baseName := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))

	// Find and remove auxiliary files
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	removedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()

		// Check if it's an auxiliary file for this PDF
		if strings.HasPrefix(fileName, baseName) {
			for _, ext := range auxExtensions {
				if strings.HasSuffix(fileName, ext) {
					filePath := filepath.Join(dir, fileName)
					if err := os.Remove(filePath); err != nil {
						// Log warning but continue
						fmt.Printf("Warning: failed to remove %s: %v\n", filePath, err)
					} else {
						removedCount++
					}
					break
				}
			}
		}
	}

	if removedCount > 0 {
		fmt.Printf("Successfully cleaned %d auxiliary files in: %s\n", removedCount, dir)
	} else {
		fmt.Printf("No auxiliary files found to clean in: %s\n", dir)
	}

	return nil
}
