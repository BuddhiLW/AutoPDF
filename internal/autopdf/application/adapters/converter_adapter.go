// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// ConverterAdapter wraps the existing PDF converter
type ConverterAdapter struct {
	config *config.Config
}

// NewConverterAdapter creates a new converter adapter
func NewConverterAdapter(cfg *config.Config) *ConverterAdapter {
	return &ConverterAdapter{
		config: cfg,
	}
}

// ConvertToImages converts a PDF to images
func (ca *ConverterAdapter) ConvertToImages(ctx context.Context, pdfPath string, formats []string) ([]string, error) {
	if pdfPath == "" {
		return nil, errors.New("no PDF file specified")
	}

	// Ensure the file has .pdf extension
	if !strings.HasSuffix(pdfPath, ".pdf") {
		pdfPath = fmt.Sprintf("%s.pdf", pdfPath)
	}

	// Check if the file exists
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("PDF file does not exist: %s", pdfPath)
	}

	// Get the formats for conversion
	if len(formats) == 0 {
		formats = []string{"png"} // Default format
	}

	outputFiles := []string{}
	dir := filepath.Dir(pdfPath)
	baseName := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))

	// Try to find ImageMagick's convert tool
	_, err := exec.LookPath("convert")
	if err == nil {
		// ImageMagick approach
		for _, format := range formats {
			outputPath := filepath.Join(dir, fmt.Sprintf("%s.%s", baseName, format))
			cmd := exec.Command("convert",
				"-density", "300",
				pdfPath,
				outputPath)

			if err := cmd.Run(); err != nil {
				return outputFiles, fmt.Errorf("image conversion failed for %s: %w", format, err)
			}

			outputFiles = append(outputFiles, outputPath)
		}

		return outputFiles, nil
	}

	// Try pdftoppm for PDF to image conversion (part of poppler-utils)
	_, err = exec.LookPath("pdftoppm")
	if err == nil {
		for _, format := range formats {
			// pdftoppm has specific flags for different formats
			var outputPrefix string
			var args []string

			switch format {
			case "png":
				outputPrefix = filepath.Join(dir, baseName)
				args = []string{"-png", pdfPath, outputPrefix}
			case "jpg", "jpeg":
				outputPrefix = filepath.Join(dir, baseName)
				args = []string{"-jpeg", pdfPath, outputPrefix}
			default:
				continue // Skip unsupported formats
			}

			cmd := exec.Command("pdftoppm", args...)

			if err := cmd.Run(); err != nil {
				return outputFiles, fmt.Errorf("image conversion failed for %s: %w", format, err)
			}

			// pdftoppm adds numeric suffixes to outputs for multi-page documents
			// For simplicity, we'll just note the base name
			outputFiles = append(outputFiles, fmt.Sprintf("%s-*.%s", outputPrefix, format))
		}

		return outputFiles, nil
	}

	return nil, errors.New("no suitable conversion tool found (tried 'convert' and 'pdftoppm')")
}
