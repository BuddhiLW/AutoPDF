// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package contract

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLIPublicContract_Build tests the public contract of the build command
// This ensures that the CLI interface remains stable across refactoring
func TestCLIPublicContract_Build(t *testing.T) {
	// Skip if autopdf binary is not available
	if _, err := exec.LookPath("autopdf"); err != nil {
		t.Skip("autopdf binary not found in PATH, skipping public contract test")
	}

	tests := []struct {
		name          string
		args          []string
		expectSuccess bool
		expectPDFFile bool
		expectStdout  string
	}{
		{
			name:          "build with template and config",
			args:          []string{"build", "testdata/template.tex", "testdata/config.yaml"},
			expectSuccess: true,
			expectPDFFile: true,
			expectStdout:  "Successfully built PDF:",
		},
		{
			name:          "build with clean option",
			args:          []string{"build", "testdata/template.tex", "testdata/config.yaml", "clean"},
			expectSuccess: true,
			expectPDFFile: true,
			expectStdout:  "Successfully built PDF:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for the test
			tempDir := t.TempDir()

			// Copy test files to temp directory
			copyTestFiles(t, tempDir)

			// Change to temp directory
			oldDir, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(oldDir)

			err = os.Chdir(tempDir)
			require.NoError(t, err)

			// Run the command
			cmd := exec.Command("autopdf", tt.args...)
			output, err := cmd.CombinedOutput()

			if tt.expectSuccess {
				assert.NoError(t, err, "Command should succeed")
				assert.Equal(t, 0, cmd.ProcessState.ExitCode(), "Exit code should be 0")
			} else {
				assert.Error(t, err, "Command should fail")
				assert.NotEqual(t, 0, cmd.ProcessState.ExitCode(), "Exit code should not be 0")
			}

			// Check stdout contains expected string
			if tt.expectStdout != "" {
				assert.Contains(t, string(output), tt.expectStdout, "Stdout should contain expected string")
			}

			// Check PDF file exists
			if tt.expectPDFFile {
				outputPath := "output.pdf" // Default from test config
				assert.FileExists(t, outputPath, "PDF file should exist")

				// Check file is not empty
				stat, err := os.Stat(outputPath)
				require.NoError(t, err)
				assert.Greater(t, stat.Size(), int64(0), "PDF file should not be empty")
			}
		})
	}
}

// TestCLIPublicContract_NoFlags ensures no flags are accepted (Bonzai philosophy)
func TestCLIPublicContract_NoFlags(t *testing.T) {
	// Skip if autopdf binary is not available
	if _, err := exec.LookPath("autopdf"); err != nil {
		t.Skip("autopdf binary not found in PATH, skipping public contract test")
	}

	// Test that flags are not accepted
	cmd := exec.Command("autopdf", "build", "--help")
	err := cmd.Run()

	// Command should fail because flags are not supported
	assert.Error(t, err, "Command with flags should fail")
	assert.NotEqual(t, 0, cmd.ProcessState.ExitCode(), "Exit code should not be 0")
}

// TestCLIPublicContract_PositionalArgs ensures positional args work correctly
func TestCLIPublicContract_PositionalArgs(t *testing.T) {
	// Skip if autopdf binary is not available
	if _, err := exec.LookPath("autopdf"); err != nil {
		t.Skip("autopdf binary not found in PATH, skipping public contract test")
	}

	tests := []struct {
		name          string
		args          []string
		expectSuccess bool
	}{
		{
			name:          "build with one arg (template only)",
			args:          []string{"build", "template.tex"},
			expectSuccess: true, // Should create default config
		},
		{
			name:          "build with two args (template + config)",
			args:          []string{"build", "template.tex", "config.yaml"},
			expectSuccess: true,
		},
		{
			name:          "build with three args (template + config + clean)",
			args:          []string{"build", "template.tex", "config.yaml", "clean"},
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for the test
			tempDir := t.TempDir()

			// Copy test files to temp directory
			copyTestFiles(t, tempDir)

			// Change to temp directory
			oldDir, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(oldDir)

			err = os.Chdir(tempDir)
			require.NoError(t, err)

			// Run the command
			cmd := exec.Command("autopdf", tt.args...)
			err = cmd.Run()

			if tt.expectSuccess {
				assert.NoError(t, err, "Command should succeed")
			} else {
				assert.Error(t, err, "Command should fail")
			}
		})
	}
}

// Helper function to copy test files to temp directory
func copyTestFiles(t *testing.T, destDir string) {
	// Create testdata directory
	testdataDir := filepath.Join(destDir, "testdata")
	err := os.MkdirAll(testdataDir, 0755)
	require.NoError(t, err)

	// Create a simple template file
	templateContent := `\documentclass{scrartcl}
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}

\title{delim[[.title]]}
\begin{document}
\maketitle
delim[[.content]]
\end{document}`

	templatePath := filepath.Join(testdataDir, "template.tex")
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	require.NoError(t, err)

	// Create a simple config file
	configContent := `template: ""
output: "output.pdf"
variables:
  title: "Test Document"
  content: "This is a test document."
engine: "pdflatex"
conversion:
  enabled: false
  formats: []
`

	configPath := filepath.Join(testdataDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)
}
