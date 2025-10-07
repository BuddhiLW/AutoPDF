package template

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// TestProcess creates a temporary LaTeX template with custom delimiters, processes
// it with the template engine, and verifies that the variables were properly
// substituted.
func TestProcess(t *testing.T) {
	// Create a temporary template file
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	log.Print(tempDir)
	fmt.Println(tempDir)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test LaTeX template with our custom delimiters
	templateContent := `
// LaTeX document with custom delimiters
\documentclass{article}
// Title with variable
\title{delim[[.title]]}
// Author with variable
\author{delim[[.author]]}
// Date with variable
\date{delim[[.date]]}

\begin{document}
// Print the title
\maketitle

// Content with variable
delim[[.content]]

\end{document}
`
	// Write the template to a temporary file
	templatePath := config.Template(filepath.Join(tempDir, "test-template.tex"))
	if err := os.WriteFile(templatePath.String(), []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Create a test config with variables
	cfg := &config.Config{
		// Use the temporary template file
		Template: templatePath,
		Variables: map[string]string{
			// Title variable
			"title": "Test Document",
			// Author variable
			"author": "Test User",
			// Date variable
			"date": "April 1, 2025",
			// Content variable
			"content": "This is a test document.",
		},
	}

	// Create and use the template engine
	engine := NewEngine(cfg)
	result, err := engine.Process(templatePath.String())
	if err != nil {
		t.Fatalf("Template processing failed: %v", err)
	}

	// Verify that variables were properly substituted
	expectedSubstrings := []string{
		// Title should be "Test Document"
		`\title{Test Document}`,
		// Author should be "Test User"
		`\author{Test User}`,
		// Date should be "April 1, 2025"
		`\date{April 1, 2025}`,
		// Content should be "This is a test document."
		`This is a test document.`,
	}

	for _, substr := range expectedSubstrings {
		if !strings.Contains(result, substr) {
			t.Errorf("Expected result to contain '%s', but it doesn't: %s", substr, result)
		}
	}
}

func TestProcessToFile(t *testing.T) {
	// Create temporary directories
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test LaTeX template with our custom delimiters
	templateContent := `\title{delim[[.title]]}`
	templatePath := filepath.Join(tempDir, "test-template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Output path for the processed file
	outputPath := filepath.Join(tempDir, "output.tex")

	// Create a test config with variables
	cfg := &config.Config{
		Template: config.Template(templatePath),
		Output:   config.Output(outputPath),
		Variables: map[string]string{
			"title": "Test Document",
		},
	}

	// Create and use the template engine
	engine := NewEngine(cfg)
	err = engine.ProcessToFile(templatePath, outputPath)
	if err != nil {
		t.Fatalf("ProcessToFile failed: %v", err)
	}

	// Verify that the output file exists and contains the expected content
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output file was not created at %s", outputPath)
	}

	outputContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	expectedContent := `\title{Test Document}`
	if string(outputContent) != expectedContent {
		t.Errorf("Output file content doesn't match. Expected '%s', got '%s'", expectedContent, string(outputContent))
	}
}

func TestAddFunction(t *testing.T) {
	// Create a temporary template file
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test LaTeX template using a custom function
	templateContent := `\title{delim[[.title | upper]]}`
	templatePath := filepath.Join(tempDir, "test-template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Create a test config
	cfg := &config.Config{
		Template: config.Template(templatePath),
		Variables: map[string]string{
			"title": "test document",
		},
	}

	// Create the template engine and add a custom function
	engine := NewEngine(cfg)
	engine.AddFunction("upper", strings.ToUpper)

	// Process the template
	result, err := engine.Process(templatePath)
	if err != nil {
		t.Fatalf("Template processing failed: %v", err)
	}

	// Verify that the custom function was applied
	expectedOutput := `\title{TEST DOCUMENT}`
	if result != expectedOutput {
		t.Errorf("Custom function not applied correctly. Expected '%s', got '%s'", expectedOutput, result)
	}
}

func TestNewEngine(t *testing.T) {
	cfg := &config.Config{
		Template: "test.tex",
		Variables: map[string]string{
			"title": "Test Document",
		},
	}

	engine := NewEngine(cfg)
	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}

	if engine.Config != cfg {
		t.Error("Engine.Config not set correctly")
	}

	if engine.FuncMap == nil {
		t.Error("Engine.FuncMap should not be nil")
	}

	// Check that default functions are present
	if _, exists := engine.FuncMap["upper"]; !exists {
		t.Error("Default 'upper' function not found in FuncMap")
	}
}

func TestEngine_Process_EmptyTemplatePath(t *testing.T) {
	cfg := &config.Config{
		Variables: map[string]string{
			"title": "Test Document",
		},
	}

	engine := NewEngine(cfg)
	_, err := engine.Process("")
	if err == nil {
		t.Error("Process with empty template path should return error")
	}
}

func TestEngine_Process_NonExistentFile(t *testing.T) {
	cfg := &config.Config{
		Variables: map[string]string{
			"title": "Test Document",
		},
	}

	engine := NewEngine(cfg)
	_, err := engine.Process("/path/to/nonexistent/file.tex")
	if err == nil {
		t.Error("Process with non-existent file should return error")
	}
}

func TestEngine_Process_InvalidTemplate(t *testing.T) {
	// Create a temporary template file with invalid template syntax
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a template with invalid syntax
	templateContent := `\title{delim[[.title | invalidFunction]]}`
	templatePath := filepath.Join(tempDir, "invalid-template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	cfg := &config.Config{
		Variables: map[string]string{
			"title": "Test Document",
		},
	}

	engine := NewEngine(cfg)
	_, err = engine.Process(templatePath)
	if err == nil {
		t.Error("Process with invalid template should return error")
	}
}

func TestEngine_ProcessToFile_EmptyOutputPath(t *testing.T) {
	// Create a temporary template file
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	templateContent := `\title{delim[[.title]]}`
	templatePath := filepath.Join(tempDir, "test-template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	cfg := &config.Config{
		Template: config.Template(templatePath),
		Variables: map[string]string{
			"title": "Test Document",
		},
	}

	engine := NewEngine(cfg)
	err = engine.ProcessToFile(templatePath, "")
	if err == nil {
		t.Error("ProcessToFile with empty output path should return error")
	}
}

func TestEngine_ProcessToFile_WithConfigOutput(t *testing.T) {
	// Create a temporary template file
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	templateContent := `\title{delim[[.title]]}`
	templatePath := filepath.Join(tempDir, "test-template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	outputPath := filepath.Join(tempDir, "output.tex")
	cfg := &config.Config{
		Template: config.Template(templatePath),
		Output:   config.Output(outputPath),
		Variables: map[string]string{
			"title": "Test Document",
		},
	}

	engine := NewEngine(cfg)
	err = engine.ProcessToFile(templatePath, "")
	if err != nil {
		t.Fatalf("ProcessToFile failed: %v", err)
	}

	// Verify that the output file exists and contains the expected content
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output file was not created at %s", outputPath)
	}

	outputContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	expectedContent := `\title{Test Document}`
	if string(outputContent) != expectedContent {
		t.Errorf("Output file content doesn't match. Expected '%s', got '%s'", expectedContent, string(outputContent))
	}
}

func TestEngine_Process_WithConfigTemplate(t *testing.T) {
	// Create a temporary template file
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	templateContent := `\title{delim[[.title]]}`
	templatePath := filepath.Join(tempDir, "test-template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	cfg := &config.Config{
		Template: config.Template(templatePath),
		Variables: map[string]string{
			"title": "Test Document",
		},
	}

	engine := NewEngine(cfg)
	result, err := engine.Process("")
	if err != nil {
		t.Fatalf("Process with config template failed: %v", err)
	}

	expectedOutput := `\title{Test Document}`
	if result != expectedOutput {
		t.Errorf("Process result doesn't match. Expected '%s', got '%s'", expectedOutput, result)
	}
}

func TestEngine_Process_WithConfigTemplate_EmptyConfigTemplate(t *testing.T) {
	cfg := &config.Config{
		Template: "",
		Variables: map[string]string{
			"title": "Test Document",
		},
	}

	engine := NewEngine(cfg)
	_, err := engine.Process("")
	if err == nil {
		t.Error("Process with empty config template should return error")
	}
}
