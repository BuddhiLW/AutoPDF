// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package args

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertArgsParser_ParseConvertArgs(t *testing.T) {
	parser := NewConvertArgsParser()

	tests := []struct {
		name        string
		args        []string
		expected    *ConvertArgs
		expectError bool
	}{
		{
			name: "single PDF file with default format",
			args: []string{"document.pdf"},
			expected: &ConvertArgs{
				PDFFile: "document.pdf",
				Formats: []string{"png"},
			},
			expectError: false,
		},
		{
			name: "PDF with single format",
			args: []string{"document.pdf", "jpeg"},
			expected: &ConvertArgs{
				PDFFile: "document.pdf",
				Formats: []string{"jpeg"},
			},
			expectError: false,
		},
		{
			name: "PDF with multiple formats",
			args: []string{"document.pdf", "png", "jpeg", "gif"},
			expected: &ConvertArgs{
				PDFFile: "document.pdf",
				Formats: []string{"png", "jpeg", "gif"},
			},
			expectError: false,
		},
		{
			name: "PDF with mixed case formats",
			args: []string{"document.pdf", "PNG", "Jpeg", "GIF"},
			expected: &ConvertArgs{
				PDFFile: "document.pdf",
				Formats: []string{"png", "jpeg", "gif"},
			},
			expectError: false,
		},
		{
			name: "PDF with formats with spaces",
			args: []string{"document.pdf", " png ", " jpeg "},
			expected: &ConvertArgs{
				PDFFile: "document.pdf",
				Formats: []string{"png", "jpeg"},
			},
			expectError: false,
		},
		{
			name:        "PDF with unsupported format",
			args:        []string{"document.pdf", "invalid"},
			expected:    nil,
			expectError: true,
		},
		{
			name:        "PDF with empty format",
			args:        []string{"document.pdf", ""},
			expected:    nil,
			expectError: true,
		},
		{
			name:        "no arguments",
			args:        []string{},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseConvertArgs(tt.args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestConvertArgsParser_IsValidFormat(t *testing.T) {
	parser := NewConvertArgsParser()

	tests := []struct {
		format   string
		expected bool
	}{
		{"png", true},
		{"jpeg", true},
		{"jpg", true},
		{"gif", true},
		{"bmp", true},
		{"tiff", true},
		{"webp", true},
		{"PNG", true}, // Case insensitive
		{"JPEG", true},
		{"invalid", false},
		{"", false},
		{"pdf", false},
		{"doc", false},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := parser.isValidFormat(tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}
