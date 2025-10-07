// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"context"

	"github.com/BuddhiLW/AutoPDF/internal/converter"
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
	// Create a config with the specified formats
	cfg := &config.Config{
		Conversion: config.Conversion{
			Enabled: true,
			Formats: formats,
		},
	}

	// Create the converter
	conv := converter.NewConverter(cfg)

	// Convert the PDF to images
	imagePaths, err := conv.ConvertPDFToImages(pdfPath)
	if err != nil {
		return nil, err
	}

	return imagePaths, nil
}
