// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package force

import (
	"os"

	"github.com/BuddhiLW/AutoPDF/configs"
)

// ForcerAdapter adapts force operations to the application.ForcerPort interface
type ForcerAdapter struct {
	forceMode bool
	overwrite bool
}

// NewForcerAdapter creates a new forcer adapter
func NewForcerAdapter() *ForcerAdapter {
	return &ForcerAdapter{
		forceMode: configs.DefaultForceEnabled,
		overwrite: configs.DefaultForceEnabled,
	}
}

// SetForceMode sets the force mode with overwrite setting
func (fa *ForcerAdapter) SetForceMode(overwrite bool) {
	fa.forceMode = true
	fa.overwrite = overwrite
}

// ShouldOverwrite returns whether files should be overwritten
func (fa *ForcerAdapter) ShouldOverwrite() bool {
	return fa.forceMode && fa.overwrite
}

// ShouldForce returns whether force mode is enabled
func (fa *ForcerAdapter) ShouldForce() bool {
	return fa.forceMode
}

// CheckFileExists checks if a file exists and handles force mode
func (fa *ForcerAdapter) CheckFileExists(filepath string) (bool, error) {
	_, err := os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // File doesn't exist
		}
		return false, err // Some other error
	}

	// File exists
	if fa.ShouldForce() && fa.ShouldOverwrite() {
		return false, nil // Force overwrite, treat as if file doesn't exist
	}

	return true, nil // File exists and we're not forcing overwrite
}
