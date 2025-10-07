// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package args

import (
	"fmt"
	"strings"
)

// ConvertArgs represents the parsed arguments for the convert command
type ConvertArgs struct {
	PDFFile string
	Formats []string
}

// ConvertArgsParser handles parsing of convert command line arguments
type ConvertArgsParser struct{}

// NewConvertArgsParser creates a new convert argument parser
func NewConvertArgsParser() *ConvertArgsParser {
	return &ConvertArgsParser{}
}

// ParseConvertArgs parses the convert command arguments
func (cap *ConvertArgsParser) ParseConvertArgs(args []string) (*ConvertArgs, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("PDF file is required")
	}

	convertArgs := &ConvertArgs{
		PDFFile: args[0],
		Formats: []string{"png"}, // Default format
	}

	// Parse formats if provided
	if len(args) > 1 {
		formats := args[1:]
		// Validate and normalize formats
		validFormats := []string{}
		for _, format := range formats {
			normalized := strings.ToLower(strings.TrimSpace(format))
			if cap.isValidFormat(normalized) {
				validFormats = append(validFormats, normalized)
			} else {
				return nil, fmt.Errorf("unsupported format: %s", format)
			}
		}
		if len(validFormats) > 0 {
			convertArgs.Formats = validFormats
		}
	}

	return convertArgs, nil
}

// isValidFormat checks if the format is supported
func (cap *ConvertArgsParser) isValidFormat(format string) bool {
	supportedFormats := map[string]bool{
		"png":  true,
		"jpeg": true,
		"jpg":  true,
		"gif":  true,
		"bmp":  true,
		"tiff": true,
		"webp": true,
	}
	return supportedFormats[strings.ToLower(format)]
}
