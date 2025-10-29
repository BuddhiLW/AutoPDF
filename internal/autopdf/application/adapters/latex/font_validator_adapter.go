// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package latex

import (
	"context"
	"os/exec"
	"strings"

	application "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/ports"
)

// FontValidatorAdapter implements ports.FontValidator using fc-list
type FontValidatorAdapter struct{}

// NewFontValidatorAdapter creates a new font validator adapter
func NewFontValidatorAdapter() *FontValidatorAdapter {
	return &FontValidatorAdapter{}
}

// ValidateFonts checks if required fonts are available using fc-list
func (f *FontValidatorAdapter) ValidateFonts(ctx context.Context, fontNames []string) (application.ValidationResult, error) {
	result := application.ValidationResult{
		AllAvailable: true,
		Missing:      []string{},
		Available:    []string{},
	}

	// Get all available fonts once
	cmd := exec.CommandContext(ctx, "fc-list", ":", "family")
	output, err := cmd.Output()
	if err != nil {
		// If fc-list fails, return error but don't block compilation
		return result, err
	}

	availableFonts := strings.ToLower(string(output))

	// Check each required font
	for _, fontName := range fontNames {
		if strings.Contains(availableFonts, strings.ToLower(fontName)) {
			result.Available = append(result.Available, fontName)
		} else {
			result.Missing = append(result.Missing, fontName)
			result.AllAvailable = false
		}
	}

	return result, nil
}

