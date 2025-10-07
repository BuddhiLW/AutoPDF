// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestEnvironment manages test file system state
type TestEnvironment struct {
	TestDir      string
	OutputDir    string
	ConfigFile   string
	TemplateFile string
	Cleanup      func()
}

// SetupTestEnvironment creates a controlled test environment
func SetupTestEnvironment(t *testing.T) *TestEnvironment {
	// Create a temporary directory for this test
	testDir := t.TempDir()

	// Create output directory within test directory
	outputDir := filepath.Join(testDir, "out")
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Try to use existing test data if available
	testConfig := DefaultTestConfig()
	testDataDir := testConfig.GetTestDataDir()

	var templateFile, configFile string

	if testDataDir != "" {
		// Copy from existing test data
		sourceTemplate := filepath.Join(testDataDir, "template.tex")
		sourceConfig := filepath.Join(testDataDir, "config.yaml")

		templateFile = filepath.Join(testDir, "template.tex")
		configFile = filepath.Join(testDir, "config.yaml")

		// Copy template file
		if _, err := os.Stat(sourceTemplate); err == nil {
			templateData, err := os.ReadFile(sourceTemplate)
			if err == nil {
				err = os.WriteFile(templateFile, templateData, 0644)
				if err != nil {
					t.Fatalf("Failed to copy template file: %v", err)
				}
			}
		}

		// Copy config file
		if _, err := os.Stat(sourceConfig); err == nil {
			configData, err := os.ReadFile(sourceConfig)
			if err == nil {
				// Copy config as-is, it should work with relative paths
				err = os.WriteFile(configFile, configData, 0644)
				if err != nil {
					t.Fatalf("Failed to copy config file: %v", err)
				}
			}
		}
	}

	// Fallback: create default files if test data not available
	if templateFile == "" {
		templateFile = filepath.Join(testDir, "template.tex")
		templateContent := `\documentclass{article}
\title{delim[[.title]]}
\begin{document}
\maketitle
delim[[.content]]
\end{document}`
		err = os.WriteFile(templateFile, []byte(templateContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create template file: %v", err)
		}
	}

	if configFile == "" {
		configFile = filepath.Join(testDir, "config.yaml")
		configContent := fmt.Sprintf(`template: "./template.tex"
output: "./out/output"
variables:
  title: "Test Document"
  content: "This is a test document."
engine: "pdflatex"
conversion:
  enabled: false`)
		err = os.WriteFile(configFile, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}
	}

	// Create cleanup function
	cleanup := func() {
		if testConfig.ShouldCleanup() {
			// Remove the entire test directory
			os.RemoveAll(testDir)
		}
	}

	return &TestEnvironment{
		TestDir:      testDir,
		OutputDir:    outputDir,
		ConfigFile:   configFile,
		TemplateFile: templateFile,
		Cleanup:      cleanup,
	}
}

// ChangeToTestDir changes the current working directory to the test directory
func (te *TestEnvironment) ChangeToTestDir() error {
	return os.Chdir(te.TestDir)
}

// RestoreWorkingDir restores the original working directory
func (te *TestEnvironment) RestoreWorkingDir(originalDir string) error {
	return os.Chdir(originalDir)
}

// GetOutputFiles returns all files in the output directory
func (te *TestEnvironment) GetOutputFiles() ([]string, error) {
	var files []string
	err := filepath.Walk(te.OutputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(te.TestDir, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}
		return nil
	})
	return files, err
}

// CleanupOutputFiles removes all files from the output directory
func (te *TestEnvironment) CleanupOutputFiles() error {
	entries, err := os.ReadDir(te.OutputDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		filePath := filepath.Join(te.OutputDir, entry.Name())
		err := os.RemoveAll(filePath)
		if err != nil {
			return err
		}
	}
	return nil
}

// AssertOutputExists checks if a specific output file exists
func (te *TestEnvironment) AssertOutputExists(t *testing.T, filename string) {
	filePath := filepath.Join(te.OutputDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Expected output file %s does not exist", filename)
	}
}

// AssertOutputNotExists checks if a specific output file does not exist
func (te *TestEnvironment) AssertOutputNotExists(t *testing.T, filename string) {
	filePath := filepath.Join(te.OutputDir, filename)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("Expected output file %s to not exist, but it does", filename)
	}
}
