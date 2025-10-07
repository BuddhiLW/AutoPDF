// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CleanerAdapter wraps the existing cleaner
type CleanerAdapter struct {
}

// NewCleanerAdapter creates a new cleaner adapter
func NewCleanerAdapter() *CleanerAdapter {
	return &CleanerAdapter{}
}

// Clean removes auxiliary files for a given PDF
func (ca *CleanerAdapter) Clean(ctx context.Context, pdfPath string) error {
	// Get the directory of the PDF
	dir := filepath.Dir(pdfPath)

	// Define auxiliary file extensions
	auxExtensions := []string{
		".aux", ".log", ".toc", ".lof", ".lot", ".out", ".nav", ".snm",
		".synctex.gz", ".fls", ".fdb_latexmk", ".bbl", ".blg", ".run.xml",
		".bcf", ".idx", ".ilg", ".ind", ".brf", ".vrb", ".xdv", ".dvi",
	}

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
