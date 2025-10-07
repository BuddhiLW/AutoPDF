// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package args

import (
	"testing"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain"
	"github.com/stretchr/testify/assert"
)

func TestArgsParser_ParseBuildArgs(t *testing.T) {
	parser := NewArgsParser()

	tests := []struct {
		name        string
		args        []string
		expected    *BuildArgs
		expectError bool
	}{
		{
			name: "minimal args - template only",
			args: []string{"template.tex"},
			expected: &BuildArgs{
				TemplateFile: "template.tex",
				ConfigFile:   "",
				Options:      domain.NewBuildOptions(),
			},
			expectError: false,
		},
		{
			name: "template and config",
			args: []string{"template.tex", "config.yaml"},
			expected: &BuildArgs{
				TemplateFile: "template.tex",
				ConfigFile:   "config.yaml",
				Options:      domain.NewBuildOptions(),
			},
			expectError: false,
		},
		{
			name: "template, config, and clean",
			args: []string{"template.tex", "config.yaml", "clean"},
			expected: &BuildArgs{
				TemplateFile: "template.tex",
				ConfigFile:   "config.yaml",
				Options: func() domain.BuildOptions {
					opts := domain.NewBuildOptions()
					opts.EnableClean(".")
					return opts
				}(),
			},
			expectError: false,
		},
		{
			name: "template and clean (no config)",
			args: []string{"template.tex", "clean"},
			expected: &BuildArgs{
				TemplateFile: "template.tex",
				ConfigFile:   "",
				Options: func() domain.BuildOptions {
					opts := domain.NewBuildOptions()
					opts.EnableClean(".")
					return opts
				}(),
			},
			expectError: false,
		},
		{
			name:        "no args",
			args:        []string{},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseBuildArgs(tt.args)

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
