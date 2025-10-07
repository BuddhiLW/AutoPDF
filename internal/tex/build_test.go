package tex

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildCmd_Structure(t *testing.T) {
	if BuildCmd == nil {
		t.Fatal("BuildCmd should not be nil")
	}

	if BuildCmd.Name != "build" {
		t.Errorf("Expected BuildCmd.Name to be 'build', got '%s'", BuildCmd.Name)
	}

	if BuildCmd.Alias != "b" {
		t.Errorf("Expected BuildCmd.Alias to be 'b', got '%s'", BuildCmd.Alias)
	}

	if BuildCmd.MinArgs != 1 {
		t.Errorf("Expected BuildCmd.MinArgs to be 1, got %d", BuildCmd.MinArgs)
	}

	if BuildCmd.MaxArgs != 3 {
		t.Errorf("Expected BuildCmd.MaxArgs to be 3, got %d", BuildCmd.MaxArgs)
	}
}

func TestBuildCmd_Do_WithTemplateOnly(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file
	templateContent := `\documentclass{article}
\title{delim[[.title]]}
\begin{document}
\maketitle
delim[[.content]]
\end{document}`
	templatePath := filepath.Join(tempDir, "template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Test with template only (should create default config)
	args := []string{templatePath}
	err = BuildCmd.Do(BuildCmd, args...)
	if err != nil {
		t.Errorf("BuildCmd.Do failed with template only: %v", err)
	}
}

func TestBuildCmd_Do_WithTemplateAndConfig(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file with variables
	templateContent := `\documentclass{scrartcl}
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}

\title{delim[[.title]]}
\begin{document}
\maketitle
delim[[.content]]
\end{document}`
	templatePath := filepath.Join(tempDir, "template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Create a test config file
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
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test with template and config
	args := []string{templatePath, configPath}
	err = BuildCmd.Do(BuildCmd, args...)
	if err != nil {
		t.Errorf("BuildCmd.Do failed with template and config: %v", err)
	}
}

func TestBuildCmd_Do_WithTemplateConfigAndClean(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file
	templateContent := `\documentclass{article}
\title{delim[[.title]]}
\begin{document}
\maketitle
delim[[.content]]
\end{document}`
	templatePath := filepath.Join(tempDir, "template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Create a test config file
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
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test with template, config, and clean
	args := []string{templatePath, configPath, "clean"}
	err = BuildCmd.Do(BuildCmd, args...)
	if err != nil {
		t.Errorf("BuildCmd.Do failed with template, config, and clean: %v", err)
	}
}

func TestBuildCmd_Do_InvalidTemplate(t *testing.T) {
	// Test with non-existent template
	args := []string{"/path/to/nonexistent/template.tex"}
	err := BuildCmd.Do(BuildCmd, args...)
	if err == nil {
		t.Error("Expected error for non-existent template but got none")
	}
}

func TestBuildCmd_Do_InvalidConfig(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file with variables
	templateContent := `\documentclass{scrartcl}
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}

\title{delim[[.title]]}
\begin{document}
\maketitle
delim[[.content]]
\end{document}`
	templatePath := filepath.Join(tempDir, "template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Test with non-existent config
	args := []string{templatePath, "/path/to/nonexistent/config.yaml"}
	err = BuildCmd.Do(BuildCmd, args...)
	if err == nil {
		t.Error("Expected error for non-existent config but got none")
	}
}

func TestBuildCmd_Do_InvalidYAML(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file with variables
	templateContent := `\documentclass{scrartcl}
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}

\title{delim[[.title]]}
\begin{document}
\maketitle
delim[[.content]]
\end{document}`
	templatePath := filepath.Join(tempDir, "template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Create an invalid YAML config file
	configContent := `template: "template.tex"
output: "output.pdf"
invalid: [unclosed array
`
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test with invalid YAML
	args := []string{templatePath, configPath}
	err = BuildCmd.Do(BuildCmd, args...)
	if err == nil {
		t.Error("Expected error for invalid YAML but got none")
	}
}

func TestBuildCmd_Do_WithCustomOutput(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file
	templateContent := `\documentclass{article}
\title{delim[[.title]]}
\begin{document}
\maketitle
delim[[.content]]
\end{document}`
	templatePath := filepath.Join(tempDir, "template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Create a test config file with custom output
	configContent := `template: ""
output: "custom_output.pdf"
variables:
  title: "Test Document"
  content: "This is a test document."
engine: "pdflatex"
conversion:
  enabled: false
  formats: []
`
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test with custom output
	args := []string{templatePath, configPath}
	err = BuildCmd.Do(BuildCmd, args...)
	if err != nil {
		t.Errorf("BuildCmd.Do failed with custom output: %v", err)
	}
}

func TestBuildCmd_Do_WithConversion(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file
	templateContent := `\documentclass{article}
\title{delim[[.title]]}
\begin{document}
\maketitle
delim[[.content]]
\end{document}`
	templatePath := filepath.Join(tempDir, "template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Create a test config file with conversion enabled
	configContent := `template: ""
output: "output.pdf"
variables:
  title: "Test Document"
  content: "This is a test document."
engine: "pdflatex"
conversion:
  enabled: true
  formats: ["png"]
`
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test with conversion enabled
	args := []string{templatePath, configPath}
	err = BuildCmd.Do(BuildCmd, args...)
	if err != nil {
		t.Errorf("BuildCmd.Do failed with conversion: %v", err)
	}
}

func TestBuildCmd_Do_WithCustomEngine(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file
	templateContent := `\documentclass{article}
\title{delim[[.title]]}
\begin{document}
\maketitle
delim[[.content]]
\end{document}`
	templatePath := filepath.Join(tempDir, "template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Create a test config file with custom engine
	configContent := `template: ""
output: "output.pdf"
variables:
  title: "Test Document"
  content: "This is a test document."
engine: "xelatex"
conversion:
  enabled: false
  formats: []
`
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test with custom engine
	args := []string{templatePath, configPath}
	err = BuildCmd.Do(BuildCmd, args...)
	if err != nil {
		t.Errorf("BuildCmd.Do failed with custom engine: %v", err)
	}
}

func TestBuildCmd_Do_ErrorHandling(t *testing.T) {
	// Test various error conditions

	// Test with empty template path
	args := []string{""}
	err := BuildCmd.Do(BuildCmd, args...)
	if err == nil {
		t.Error("Expected error for empty template path but got none")
	}
}

func TestBuildCmd_Do_Integration(t *testing.T) {
	// Test the full integration of the BuildCmd
	// This tests the entire flow from template processing to PDF generation

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file
	templateContent := `\documentclass{scrartcl}
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}

\title{delim[[.title]]}
\author{delim[[.author]]}
\begin{document}
\maketitle
delim[[.content]]
\end{document}`
	templatePath := filepath.Join(tempDir, "template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Create a test config file
	configContent := `template: ""
output: "integration_output.pdf"
variables:
  title: "Integration Test Document"
  author: "Test Author"
  content: "This is an integration test for the BuildCmd."
engine: "pdflatex"
conversion:
  enabled: false
  formats: []
`
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test the full command execution
	args := []string{templatePath, configPath}
	err = BuildCmd.Do(BuildCmd, args...)
	if err != nil {
		t.Errorf("BuildCmd integration test failed: %v", err)
	}
}
