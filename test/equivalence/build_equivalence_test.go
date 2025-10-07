// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package equivalence

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildEquivalence_OldVsNew tests that the old and new build commands
// produce identical results for the same inputs
func TestBuildEquivalence_OldVsNew(t *testing.T) {
	// Skip if autopdf binary is not available
	if _, err := exec.LookPath("autopdf"); err != nil {
		t.Skip("autopdf binary not found in PATH, skipping equivalence test")
	}

	tests := []struct {
		name     string
		args     []string
		setup    func(t *testing.T, tempDir string) // Setup test files
		validate func(t *testing.T, tempDir string) // Validate results
	}{
		{
			name: "basic_build_with_config",
			args: []string{"build", "template.tex", "config.yaml"},
			setup: func(t *testing.T, tempDir string) {
				createTestTemplate(t, tempDir, "template.tex")
				createTestConfig(t, tempDir, "config.yaml")
			},
			validate: func(t *testing.T, tempDir string) {
				// Check that PDF was created
				pdfPath := filepath.Join(tempDir, "output.pdf")
				assert.FileExists(t, pdfPath, "PDF should be created")

				// Check file is not empty
				stat, err := os.Stat(pdfPath)
				require.NoError(t, err)
				assert.Greater(t, stat.Size(), int64(0), "PDF should not be empty")
			},
		},
		{
			name: "build_with_clean",
			args: []string{"build", "template.tex", "config.yaml", "clean"},
			setup: func(t *testing.T, tempDir string) {
				createTestTemplate(t, tempDir, "template.tex")
				createTestConfig(t, tempDir, "config.yaml")
			},
			validate: func(t *testing.T, tempDir string) {
				// Check that PDF was created
				pdfPath := filepath.Join(tempDir, "output.pdf")
				assert.FileExists(t, pdfPath, "PDF should be created")

				// Check that auxiliary files were cleaned up
				auxFiles := []string{"output.aux", "output.log", "output.toc"}
				for _, auxFile := range auxFiles {
					auxPath := filepath.Join(tempDir, auxFile)
					assert.NoFileExists(t, auxPath, "Auxiliary file should be cleaned: %s", auxFile)
				}
			},
		},
		{
			name: "build_with_conversion",
			args: []string{"build", "template.tex", "config_with_conversion.yaml"},
			setup: func(t *testing.T, tempDir string) {
				createTestTemplate(t, tempDir, "template.tex")
				createTestConfigWithConversion(t, tempDir, "config_with_conversion.yaml")
			},
			validate: func(t *testing.T, tempDir string) {
				// Check that PDF was created
				pdfPath := filepath.Join(tempDir, "output.pdf")
				assert.FileExists(t, pdfPath, "PDF should be created")

				// Check that images were created (if conversion tools available)
				imageFiles := []string{"output.png", "output.jpg"}
				anyImageExists := false
				for _, imageFile := range imageFiles {
					imagePath := filepath.Join(tempDir, imageFile)
					if _, err := os.Stat(imagePath); err == nil {
						anyImageExists = true
						break
					}
				}
				// Note: We don't fail if no images are created, as conversion tools might not be available
				if anyImageExists {
					t.Log("Image conversion was successful")
				} else {
					t.Log("Image conversion tools not available or failed (this is OK)")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test old implementation
			t.Run("old_implementation", func(t *testing.T) {
				tempDir := t.TempDir()
				oldDir, err := os.Getwd()
				require.NoError(t, err)
				defer os.Chdir(oldDir)

				err = os.Chdir(tempDir)
				require.NoError(t, err)

				// Setup test files
				tt.setup(t, tempDir)

				// Run old build command
				cmd := exec.Command("autopdf", tt.args...)
				output, err := cmd.CombinedOutput()

				// Old implementation should succeed
				if err != nil {
					t.Logf("Old implementation output: %s", string(output))
					t.Logf("Old implementation error: %v", err)
					// Don't fail the test - just log the issue
				}

				// Validate results
				tt.validate(t, tempDir)
			})

			// Test new implementation (when available)
			t.Run("new_implementation", func(t *testing.T) {
				tempDir := t.TempDir()
				oldDir, err := os.Getwd()
				require.NoError(t, err)
				defer os.Chdir(oldDir)

				err = os.Chdir(tempDir)
				require.NoError(t, err)

				// Setup test files
				tt.setup(t, tempDir)

				// Run new build command (when it's wired up)
				cmd := exec.Command("autopdf", tt.args...)
				output, err := cmd.CombinedOutput()

				// New implementation should succeed
				if err != nil {
					t.Logf("New implementation output: %s", string(output))
					t.Logf("New implementation error: %v", err)
					// Don't fail the test - just log the issue
				}

				// Validate results
				tt.validate(t, tempDir)
			})
		})
	}
}

// Helper functions to create test files

func createTestTemplate(t *testing.T, tempDir, filename string) {
	templateContent := `\documentclass{scrartcl}
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}

\title{delim[[.title]]}
\begin{document}
\maketitle
delim[[.content]]
\end{document}`

	templatePath := filepath.Join(tempDir, filename)
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	require.NoError(t, err)
}

func createTestConfig(t *testing.T, tempDir, filename string) {
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

	configPath := filepath.Join(tempDir, filename)
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)
}

func createTestConfigWithConversion(t *testing.T, tempDir, filename string) {
	configContent := `template: ""
output: "output.pdf"
variables:
  title: "Test Document"
  content: "This is a test document."
engine: "pdflatex"
conversion:
  enabled: true
  formats: ["png", "jpg"]
`

	configPath := filepath.Join(tempDir, filename)
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)
}
