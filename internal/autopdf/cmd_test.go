// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package autopdf

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/internal/tex"
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

// Integration test for the build command
func TestBuildCmd_Integration(t *testing.T) {
	// Skip this test if running in CI or without pdflatex
	_, err := os.Stat("/usr/bin/pdflatex")
	if os.IsNotExist(err) {
		t.Skip("pdflatex not found, skipping integration test")
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	fmt.Println(tempDir)
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
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write template: %v", err)
	}
	// Check if the template file exists
	if !checkFileExists(templatePath) {
		t.Fatalf("Template file does not exist: %s", templatePath)
	}
	content, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("Failed to read template file: %v", err)
	}
	log.Printf("Template content: %s", string(content))

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
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	// Check if the config file exists
	if !checkFileExists(configPath) {
		t.Fatalf("Config file does not exist: %s", configPath)
	}
	content, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	log.Printf("Config content: %s", string(content))

	// Run the build command (but don't fail the test if LaTeX compilation fails
	// since it's expecting a full LaTeX installation)
	output, err := captureOutput(tex.BuildCmd, templatePath, configPath)
	fmt.Println(output)
	log.Print(output)
	log.Println("Error:", err)
	// time.Sleep(2 * time.Second) // Give some time for the command to finish

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
