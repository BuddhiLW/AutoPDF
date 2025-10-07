// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"context"
	"path/filepath"

	"github.com/BuddhiLW/AutoPDF/internal/tex"
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

	// Create the cleaner
	cleaner := tex.NewCleaner(dir)

	// Clean the auxiliary files
	if err := cleaner.Clean(); err != nil {
		return err
	}

	return nil
}
