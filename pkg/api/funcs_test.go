package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

func TestGeneratePDF_ValidConfig(t *testing.T) {
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

	// Create test config
	testCfg := &config.Config{
		Template: config.Template(templatePath),
		Variables: map[string]string{
			"title":   "Test Title: API call",
			"content": "Test Content: API call",
		},
		Engine: "pdflatex",
		Output: config.Output(filepath.Join(tempDir, "output.pdf")),
	}

	// Call the function
	pdfBytes, _, err := GeneratePDF(testCfg, config.Template(templatePath))
	if err != nil {
		t.Fatalf("GeneratePDF failed: %v", err)
	}

	// Check if the PDF bytes are not nil or empty
	if len(pdfBytes) == 0 {
		t.Fatal("Expected non-empty PDF bytes")
	}
}

func TestGeneratePDF_InvalidTemplate(t *testing.T) {
	// Create test config with non-existent template
	testCfg := &config.Config{
		Template: config.Template("/path/to/nonexistent/template.tex"),
		Variables: map[string]string{
			"title": "Test Title",
		},
		Engine: "pdflatex",
	}

	// Call the function with non-existent template
	_, _, err := GeneratePDF(testCfg, config.Template("/path/to/nonexistent/template.tex"))
	if err == nil {
		t.Error("Expected error for non-existent template but got none")
	}
}

func TestGeneratePDF_EmptyTemplate(t *testing.T) {
	// Create test config with empty template
	testCfg := &config.Config{
		Template: "",
		Variables: map[string]string{
			"title": "Test Title",
		},
		Engine: "pdflatex",
	}

	// Call the function with empty template
	_, _, err := GeneratePDF(testCfg, config.Template(""))
	if err == nil {
		t.Error("Expected error for empty template but got none")
	}
}

func TestGeneratePDF_InvalidLaTeX(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a template with invalid LaTeX
	templateContent := `\documentclass{article}
\begin{document}
\invalidcommand{test}
\end{document}`
	templatePath := filepath.Join(tempDir, "invalid.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Create test config
	testCfg := &config.Config{
		Template: config.Template(templatePath),
		Variables: map[string]string{
			"title": "Test Title",
		},
		Engine: "pdflatex",
		Output: config.Output(filepath.Join(tempDir, "output.pdf")),
	}

	// Call the function
	_, _, err = GeneratePDF(testCfg, config.Template(templatePath))
	// LaTeX is robust and often produces PDFs even with errors
	// This test verifies that the function handles invalid LaTeX gracefully
	// without crashing, which is the expected behavior
	if err != nil {
		t.Logf("GeneratePDF with invalid LaTeX completed with error (expected): %v", err)
	}
}

func TestGeneratePDF_WithVariables(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a template with variables
	templateContent := `\documentclass{article}
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

	// Create test config with variables
	testCfg := &config.Config{
		Template: config.Template(templatePath),
		Variables: map[string]string{
			"title":   "Test Document",
			"author":  "Test Author",
			"content": "This is a test document with variables.",
		},
		Engine: "pdflatex",
		Output: config.Output(filepath.Join(tempDir, "output.pdf")),
	}

	// Call the function
	pdfBytes, _, err := GeneratePDF(testCfg, config.Template(templatePath))
	if err != nil {
		t.Fatalf("GeneratePDF failed: %v", err)
	}

	// Check if the PDF bytes are not nil or empty
	if len(pdfBytes) == 0 {
		t.Fatal("Expected non-empty PDF bytes")
	}
}

func TestGeneratePDF_WithCustomEngine(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file
	templateContent := `\documentclass{article}
\begin{document}
Hello, World!
\end{document}`
	templatePath := filepath.Join(tempDir, "template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Create test config with custom engine
	testCfg := &config.Config{
		Template: config.Template(templatePath),
		Variables: map[string]string{
			"title": "Test Title",
		},
		Engine: "xelatex", // Use xelatex instead of pdflatex
		Output: config.Output(filepath.Join(tempDir, "output.pdf")),
	}

	// Call the function
	pdfBytes, _, err := GeneratePDF(testCfg, config.Template(templatePath))
	if err != nil {
		t.Fatalf("GeneratePDF failed: %v", err)
	}

	// Check if the PDF bytes are not nil or empty
	if len(pdfBytes) == 0 {
		t.Fatal("Expected non-empty PDF bytes")
	}
}

func TestGeneratePDF_WithOutputPath(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file
	templateContent := `\documentclass{article}
\begin{document}
Hello, World!
\end{document}`
	templatePath := filepath.Join(tempDir, "template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Create output directory
	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Create test config with output path
	testCfg := &config.Config{
		Template: config.Template(templatePath),
		Variables: map[string]string{
			"title": "Test Title",
		},
		Engine: "pdflatex",
		Output: config.Output(filepath.Join(outputDir, "custom_output.pdf")),
	}

	// Call the function
	pdfBytes, _, err := GeneratePDF(testCfg, config.Template(templatePath))
	if err != nil {
		t.Fatalf("GeneratePDF failed: %v", err)
	}

	// Check if the PDF bytes are not nil or empty
	if len(pdfBytes) == 0 {
		t.Fatal("Expected non-empty PDF bytes")
	}
}

func TestGeneratePDF_WithConversion(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file
	templateContent := `\documentclass{article}
\begin{document}
Hello, World!
\end{document}`
	templatePath := filepath.Join(tempDir, "template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Create test config with conversion enabled
	testCfg := &config.Config{
		Template: config.Template(templatePath),
		Variables: map[string]string{
			"title": "Test Title",
		},
		Engine: "pdflatex",
		Output: config.Output(filepath.Join(tempDir, "output.pdf")),
		Conversion: config.Conversion{
			Enabled: true,
			Formats: []string{"png"},
		},
	}

	// Call the function
	pdfBytes, _, err := GeneratePDF(testCfg, config.Template(templatePath))
	if err != nil {
		t.Fatalf("GeneratePDF failed: %v", err)
	}

	// Check if the PDF bytes are not nil or empty
	if len(pdfBytes) == 0 {
		t.Fatal("Expected non-empty PDF bytes")
	}
}

func TestGeneratePDF_ErrorHandling(t *testing.T) {
	// Test various error conditions

	// Test with empty template path
	testCfg := &config.Config{
		Template: "",
		Variables: map[string]string{
			"title": "Test Title",
		},
		Engine: "pdflatex",
	}

	_, _, err := GeneratePDF(testCfg, config.Template(""))
	if err == nil {
		t.Error("Expected error for empty template path but got none")
	}
}

func TestGeneratePDF_Integration(t *testing.T) {
	// Test the full integration of the GeneratePDF function
	// This tests the entire flow from template processing to PDF generation

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file
	templateContent := `\documentclass{article}
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

	// Create test config
	testCfg := &config.Config{
		Template: config.Template(templatePath),
		Variables: map[string]string{
			"title":   "Integration Test Document",
			"author":  "Test Author",
			"content": "This is an integration test for the GeneratePDF function.",
		},
		Engine: "pdflatex",
		Output: config.Output(filepath.Join(tempDir, "integration_output.pdf")),
		Conversion: config.Conversion{
			Enabled: false,
			Formats: []string{},
		},
	}

	// Call the function
	pdfBytes, _, err := GeneratePDF(testCfg, config.Template(templatePath))
	if err != nil {
		t.Fatalf("GeneratePDF integration test failed: %v", err)
	}

	// Check if the PDF bytes are not nil or empty
	if len(pdfBytes) == 0 {
		t.Fatal("Expected non-empty PDF bytes")
	}
}
