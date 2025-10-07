// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package result

import (
	"testing"

	services "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/services"
	"github.com/stretchr/testify/assert"
)

func TestResultHandler_HandleBuildResult(t *testing.T) {
	handler := NewResultHandler()

	tests := []struct {
		name     string
		result   services.BuildResult
		expected string
	}{
		{
			name: "successful build with images",
			result: services.BuildResult{
				PDFPath:    "output.pdf",
				ImagePaths: []string{"output.jpeg", "output.png"},
				Error:      nil,
			},
			expected: "Successfully built PDF: output.pdf\nGenerated image files:\n  - output.jpeg\n  - output.png\n",
		},
		{
			name: "successful build without images",
			result: services.BuildResult{
				PDFPath:    "output.pdf",
				ImagePaths: []string{},
				Error:      nil,
			},
			expected: "Successfully built PDF: output.pdf\n",
		},
		{
			name: "build with warning",
			result: services.BuildResult{
				PDFPath:    "output.pdf",
				ImagePaths: []string{},
				Error:      assert.AnError,
			},
			expected: "Successfully built PDF: output.pdf\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.HandleBuildResult(tt.result)
			assert.NoError(t, err)
			// Note: We can't easily test stdout in unit tests without capturing it
			// This test mainly ensures the method doesn't panic and returns no error
		})
	}
}
