// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package args

import (
	"testing"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain"
	"github.com/stretchr/testify/assert"
)

func TestArgsParser_ParseBuildArgs_WithOptions_Simple(t *testing.T) {
	parser := NewArgsParser()

	tests := []struct {
		name        string
		args        []string
		expectError bool
		validate    func(t *testing.T, result *BuildArgs)
	}{
		{
			name:        "template only",
			args:        []string{"template.tex"},
			expectError: false,
			validate: func(t *testing.T, result *BuildArgs) {
				assert.Equal(t, "template.tex", result.TemplateFile)
				assert.Equal(t, "", result.ConfigFile)
				assert.False(t, result.Options.Clean.Enabled)
				assert.False(t, result.Options.Verbose.Enabled)
				assert.False(t, result.Options.Debug.Enabled)
				assert.False(t, result.Options.Force.Enabled)
			},
		},
		{
			name:        "template with config",
			args:        []string{"template.tex", "config.yaml"},
			expectError: false,
			validate: func(t *testing.T, result *BuildArgs) {
				assert.Equal(t, "template.tex", result.TemplateFile)
				assert.Equal(t, "config.yaml", result.ConfigFile)
				assert.False(t, result.Options.Clean.Enabled)
			},
		},
		{
			name:        "template with clean option",
			args:        []string{"template.tex", "clean"},
			expectError: false,
			validate: func(t *testing.T, result *BuildArgs) {
				assert.Equal(t, "template.tex", result.TemplateFile)
				assert.True(t, result.Options.Clean.Enabled)
				assert.Equal(t, ".", result.Options.Clean.Target)
			},
		},
		{
			name:        "template with verbose option",
			args:        []string{"template.tex", "verbose"},
			expectError: false,
			validate: func(t *testing.T, result *BuildArgs) {
				assert.Equal(t, "template.tex", result.TemplateFile)
				assert.True(t, result.Options.Verbose.Enabled)
				assert.Equal(t, 2, result.Options.Verbose.Level)
			},
		},
		{
			name:        "template with debug option",
			args:        []string{"template.tex", "debug"},
			expectError: false,
			validate: func(t *testing.T, result *BuildArgs) {
				assert.Equal(t, "template.tex", result.TemplateFile)
				assert.True(t, result.Options.Debug.Enabled)
				assert.Equal(t, "stdout", result.Options.Debug.Output)
			},
		},
		{
			name:        "template with force option",
			args:        []string{"template.tex", "force"},
			expectError: false,
			validate: func(t *testing.T, result *BuildArgs) {
				assert.Equal(t, "template.tex", result.TemplateFile)
				assert.True(t, result.Options.Force.Enabled)
				assert.True(t, result.Options.Force.Overwrite)
			},
		},
		{
			name:        "template with multiple options",
			args:        []string{"template.tex", "clean", "verbose", "debug"},
			expectError: false,
			validate: func(t *testing.T, result *BuildArgs) {
				assert.Equal(t, "template.tex", result.TemplateFile)
				assert.True(t, result.Options.Clean.Enabled)
				assert.True(t, result.Options.Verbose.Enabled)
				assert.True(t, result.Options.Debug.Enabled)
				assert.False(t, result.Options.Force.Enabled)
			},
		},
		{
			name:        "template with config and multiple options",
			args:        []string{"template.tex", "config.yaml", "clean", "verbose"},
			expectError: false,
			validate: func(t *testing.T, result *BuildArgs) {
				assert.Equal(t, "template.tex", result.TemplateFile)
				assert.Equal(t, "config.yaml", result.ConfigFile)
				assert.True(t, result.Options.Clean.Enabled)
				assert.True(t, result.Options.Verbose.Enabled)
				assert.False(t, result.Options.Debug.Enabled)
				assert.False(t, result.Options.Force.Enabled)
			},
		},
		{
			name:        "template with all options",
			args:        []string{"template.tex", "clean", "verbose", "debug", "force"},
			expectError: false,
			validate: func(t *testing.T, result *BuildArgs) {
				assert.Equal(t, "template.tex", result.TemplateFile)
				assert.True(t, result.Options.Clean.Enabled)
				assert.True(t, result.Options.Verbose.Enabled)
				assert.True(t, result.Options.Debug.Enabled)
				assert.True(t, result.Options.Force.Enabled)
			},
		},
		{
			name:        "no arguments",
			args:        []string{},
			expectError: true,
			validate: func(t *testing.T, result *BuildArgs) {
				assert.Nil(t, result)
			},
		},
		{
			name:        "invalid option",
			args:        []string{"template.tex", "invalid"},
			expectError: true,
			validate: func(t *testing.T, result *BuildArgs) {
				assert.Nil(t, result)
			},
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
				assert.NotNil(t, result)
				tt.validate(t, result)
			}
		})
	}
}

func TestArgsParser_IsOption(t *testing.T) {
	parser := NewArgsParser()

	tests := []struct {
		arg      string
		expected bool
	}{
		{"clean", true},
		{"verbose", true},
		{"debug", true},
		{"force", true},
		{"config.yaml", false},
		{"template.tex", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			result := parser.isOption(tt.arg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestArgsParser_SetOption(t *testing.T) {
	parser := NewArgsParser()

	tests := []struct {
		name     string
		option   string
		validate func(t *testing.T, options *domain.BuildOptions)
	}{
		{
			name:   "clean option",
			option: "clean",
			validate: func(t *testing.T, options *domain.BuildOptions) {
				assert.True(t, options.Clean.Enabled)
				assert.Equal(t, ".", options.Clean.Target)
			},
		},
		{
			name:   "verbose option",
			option: "verbose",
			validate: func(t *testing.T, options *domain.BuildOptions) {
				assert.True(t, options.Verbose.Enabled)
				assert.Equal(t, 2, options.Verbose.Level)
			},
		},
		{
			name:   "debug option",
			option: "debug",
			validate: func(t *testing.T, options *domain.BuildOptions) {
				assert.True(t, options.Debug.Enabled)
				assert.Equal(t, "stdout", options.Debug.Output)
			},
		},
		{
			name:   "force option",
			option: "force",
			validate: func(t *testing.T, options *domain.BuildOptions) {
				assert.True(t, options.Force.Enabled)
				assert.True(t, options.Force.Overwrite)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := domain.NewBuildOptions()
			parser.setOption(&options, tt.option)
			tt.validate(t, &options)
		})
	}
}

func TestArgsParser_SetOption_Multiple(t *testing.T) {
	parser := NewArgsParser()

	// Test setting multiple options
	options := domain.NewBuildOptions()
	parser.setOption(&options, "clean")
	parser.setOption(&options, "verbose")
	parser.setOption(&options, "debug")

	assert.True(t, options.Clean.Enabled)
	assert.True(t, options.Verbose.Enabled)
	assert.True(t, options.Debug.Enabled)
	assert.False(t, options.Force.Enabled)
}
