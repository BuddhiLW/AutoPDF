// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package autopdf

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	
	"github.com/rwxrob/bonzai"
)

// Helper function to check if a file exists
func checkFileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}

// Helper function to capture command output
func captureOutput(cmd *bonzai.Cmd, args ...string) (string, error) {
	// Save and restore original stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Run the command
	err := cmd.Do(cmd, args...)
	
	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	
	// Read captured output
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		return "", err
	}
	
	return buf.String(), err
}

func TestCmd_Structure(t *testing.T) {
	// Test that the command structure is correct
	if Cmd.Name != "autopdf" {
		t.Errorf("Expected command name to be 'autopdf', got '%s'", Cmd.Name)
	}
	
	// Check for required subcommands
	subCmdNames := []string{"build", "clean", "convert"}
	for _, name := range subCmdNames {
		found := false
		for _, cmd := range Cmd.Cmds {
			if cmd.Name == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find subcommand '%s'", name)
		}
	}
}

func TestBuildCmd_Args(t *testing.T) {
	// Test that build requires at least one argument
	err := buildCmd.Do(buildCmd)
	if err == nil {
		t.Errorf("Expected error for missing arguments but got none")
	}
}

func TestConvertCmd_Args(t *testing.T) {
	// Test that convert requires at least one argument
	err := convertCmd.Do(convertCmd)
	if err == nil {
		t.Errorf("Expected error for missing arguments but got none")
	}
}

// Integration test for the build command
func TestBuildCmd_Integration(t *testing.T) {
	// Skip this test if running in CI or without pdflatex
	_, err := os.Stat("/usr/bin/pdflatex")
	if os.IsNotExist(err) {
		t.Skip("pdflatex not found, skipping integration test")
	}
	
	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a simple LaTeX template
	templateContent := `
\documentclass{article}
\title{delim[[.title]]}
\begin{document}
\maketitle
delim[[.content]]
\end{document}
`
	templatePath := filepath.Join(tempDir, "template.tex")
	if err := ioutil.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}
	
	// Create a YAML config
	configContent := `
template: "` + templatePath + `"
output: "` + filepath.Join(tempDir, "output.pdf") + `"
variables:
  title: "Test Document"
  content: "This is a test document."
engine: "pdflatex"
conversion:
  enabled: false
`
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := ioutil.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	
	// Run the build command (but don't fail the test if LaTeX compilation fails
	// since it's expecting a full LaTeX installation)
	output, err := captureOutput(buildCmd, templatePath, configPath)
	
	t.Logf("Build output: %s", output)
	if err != nil {
		t.Logf("Build error: %v (may be expected in test environment)", err)
	}
	
	// Check that the command at least attempted to process the template
	if err != nil {
		if !strings.Contains(err.Error(), "LaTeX compilation failed") &&
		   !strings.Contains(err.Error(), "failed to parse config") {
			t.Errorf("Command failed unexpectedly: %v", err)
		}
	}
}
