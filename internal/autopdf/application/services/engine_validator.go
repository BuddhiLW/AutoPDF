// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"errors"
	"os/exec"
)

// EngineValidator validates LaTeX engine availability
type EngineValidator struct{}

// NewEngineValidator creates a new engine validator
func NewEngineValidator() *EngineValidator {
	return &EngineValidator{}
}

// Validate checks if the LaTeX engine is available
func (v *EngineValidator) Validate(engine string) error {
	if _, err := exec.LookPath(engine); err != nil {
		return errors.New("LaTeX engine not found: " + engine)
	}
	return nil
}
