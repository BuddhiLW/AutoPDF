// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"path/filepath"

	ports "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/ports"
)

// OSPathOperations implements PathOperations using the standard library
type OSPathOperations struct{}

// NewOSPathOperations creates a new OS path operations adapter
func NewOSPathOperations() ports.PathOperations {
	return &OSPathOperations{}
}

// Dir returns the directory part of the path
func (p *OSPathOperations) Dir(path string) string {
	return filepath.Dir(path)
}

// Base returns the last element of the path
func (p *OSPathOperations) Base(path string) string {
	return filepath.Base(path)
}

// Ext returns the file extension
func (p *OSPathOperations) Ext(path string) string {
	return filepath.Ext(path)
}

// Join joins path elements
func (p *OSPathOperations) Join(elem ...string) string {
	return filepath.Join(elem...)
}

// IsAbs returns true if the path is absolute
func (p *OSPathOperations) IsAbs(path string) bool {
	return filepath.IsAbs(path)
}

// Clean returns the shortest path name equivalent to path
func (p *OSPathOperations) Clean(path string) string {
	return filepath.Clean(path)
}
