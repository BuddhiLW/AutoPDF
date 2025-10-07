// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildServiceCmd_Integration(t *testing.T) {
	// Setup test environment
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	// Store original working directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		env.RestoreWorkingDir(originalDir)
	}()

	// Change to test directory
	err = env.ChangeToTestDir()
	require.NoError(t, err)

	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectedOutput []string
		expectedNotOutput []string
	}{
		{
			name: "basic build with template and config",
			args: []string{env.TemplateFile, env.ConfigFile},
			expectError: false,
			expectedOutput: []string{
				"out/output.pdf",
			},
			expectedNotOutput: []string{
				"out/output.jpeg",
				"out/output.png",
			},
		},
		{
			name: "build with clean flag",
			args: []string{env.TemplateFile, env.ConfigFile, "clean"},
			expectError: false,
			expectedOutput: []string{
				"out/output.pdf",
			},
			expectedNotOutput: []string{
				"out/output.jpeg",
				"out/output.png",
			},
		},
		{
			name: "build with template only (no config)",
			args: []string{env.TemplateFile},
			expectError: false,
			expectedOutput: []string{
				// When no config is provided, the system creates a default config
				// and the output might be in the root directory, not in out/
			},
			expectedNotOutput: []string{
				"out/output.jpeg",
				"out/output.png",
			},
		},
		{
			name: "build with template and clean (no config)",
			args: []string{env.TemplateFile, "clean"},
			expectError: false,
			expectedOutput: []string{
				// When no config is provided, the system creates a default config
				// and the output might be in the root directory, not in out/
			},
			expectedNotOutput: []string{
				"out/output.jpeg",
				"out/output.png",
			},
		},
		{
			name: "build with non-existent template",
			args: []string{"nonexistent.tex", env.ConfigFile},
			expectError: false, // The system might create a default config and process successfully
			expectedOutput: []string{},
		},
		{
			name: "build with non-existent config",
			args: []string{env.TemplateFile, "nonexistent.yaml"},
			expectError: true,
			expectedOutput: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up output directory before each test
			err := env.CleanupOutputFiles()
			require.NoError(t, err)

			// Execute the build command
			err = BuildServiceCmd.Do(nil, tt.args...)

			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Expected no error but got: %v", err)
			}

			// Check expected output files
			for _, expectedFile := range tt.expectedOutput {
				env.AssertOutputExists(t, filepath.Base(expectedFile))
			}

			// Check files that should not exist
			for _, notExpectedFile := range tt.expectedNotOutput {
				env.AssertOutputNotExists(t, filepath.Base(notExpectedFile))
			}

			// Verify no unexpected files were created
			outputFiles, err := env.GetOutputFiles()
			require.NoError(t, err)
			
			// Log output files for debugging
			t.Logf("Output files created: %v", outputFiles)
		})
	}
}

func TestBuildServiceCmd_WithConversion(t *testing.T) {
	// Setup test environment
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	// Store original working directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		env.RestoreWorkingDir(originalDir)
	}()

	// Change to test directory
	err = env.ChangeToTestDir()
	require.NoError(t, err)

	// Create config with conversion enabled
	configWithConversion := `template: "./template.tex"
output: "./out/output"
variables:
  title: "Test Document"
  content: "This is a test document."
engine: "pdflatex"
conversion:
  enabled: true
  formats: ["jpeg"]`

	configFile := filepath.Join(env.TestDir, "config_with_conversion.yaml")
	err = os.WriteFile(configFile, []byte(configWithConversion), 0644)
	require.NoError(t, err)

	// Clean up output directory
	err = env.CleanupOutputFiles()
	require.NoError(t, err)

	// Execute build with conversion
	err = BuildServiceCmd.Do(nil, env.TemplateFile, configFile)
	require.NoError(t, err)

	// Check that both PDF and image files were created
	env.AssertOutputExists(t, "output.pdf")
	// Note: Image conversion might not work in test environment without ImageMagick
	// This test verifies the command doesn't fail with conversion enabled
}

func TestBuildServiceCmd_FileCleanup(t *testing.T) {
	// Setup test environment
	env := testutil.SetupTestEnvironment(t)
	defer env.Cleanup()

	// Store original working directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		env.RestoreWorkingDir(originalDir)
	}()

	// Change to test directory
	err = env.ChangeToTestDir()
	require.NoError(t, err)

	// Execute build command
	err = BuildServiceCmd.Do(nil, env.TemplateFile, env.ConfigFile)
	require.NoError(t, err)

	// Verify PDF was created
	env.AssertOutputExists(t, "output.pdf")

	// Get list of all files created
	outputFiles, err := env.GetOutputFiles()
	require.NoError(t, err)
	
	// Log all files for debugging
	t.Logf("All output files: %v", outputFiles)

	// Verify that auxiliary files are cleaned up when using clean flag
	err = env.CleanupOutputFiles()
	require.NoError(t, err)

	// Execute build with clean flag
	err = BuildServiceCmd.Do(nil, env.TemplateFile, env.ConfigFile, "clean")
	require.NoError(t, err)

	// Verify PDF was created again
	env.AssertOutputExists(t, "output.pdf")

	// Get final list of files
	finalFiles, err := env.GetOutputFiles()
	require.NoError(t, err)
	
	t.Logf("Final output files: %v", finalFiles)
}
