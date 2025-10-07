// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"context"

	"github.com/BuddhiLW/AutoPDF/internal/tex"
)

// CleanerAdapter adapts the tex.Cleaner to the application.CleanerPort interface
type CleanerAdapter struct {
	cleaner *tex.Cleaner
}

// NewCleanerAdapter creates a new cleaner adapter
func NewCleanerAdapter() *CleanerAdapter {
	return &CleanerAdapter{}
}

// CleanAux implements the CleanerPort interface
func (ca *CleanerAdapter) CleanAux(ctx context.Context, target string) error {
	// Create a new cleaner for the target directory
	cleaner := tex.NewCleaner(target)
	return cleaner.Clean()
}
