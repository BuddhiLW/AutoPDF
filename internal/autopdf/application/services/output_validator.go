// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"errors"
	"os"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/ports"
)

// OutputValidator validates PDF output
type OutputValidator struct {
	fileSystem ports.FileSystem
}

// NewOutputValidator creates a new output validator
func NewOutputValidator(fileSystem ports.FileSystem) *OutputValidator {
	return &OutputValidator{
		fileSystem: fileSystem,
	}
}

// Validate checks if the output PDF exists and has content
func (v *OutputValidator) Validate(pdfPath string) (string, error) {
	// Check if output PDF exists and has content
	fileInfo, statErr := v.fileSystem.Stat(pdfPath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return "", errors.New("PDF output file was not created")
		}
		return "", errors.New("error checking output file: " + statErr.Error())
	}

	// Check if file is empty
	if fileInfo.Size() == 0 {
		return "", errors.New("PDF output file was created but is empty")
	}

	return pdfPath, nil
}
