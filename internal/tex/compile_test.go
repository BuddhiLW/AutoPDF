package tex

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

func TestReplaceExt(t *testing.T) {
	testCases := []struct {
		filename string
		newExt   string
		expected string
	}{
		{"document.tex", ".pdf", "document.pdf"},
		{"path/to/document.tex", ".pdf", "path/to/document.pdf"},
		{"document", ".pdf", "document.pdf"},
		{"document.with.dots.tex", ".pdf", "document.with.dots.pdf"},
		{".hidden", ".pdf", ".pdf"},
	}

	for _, tc := range testCases {
		result := replaceExt(tc.filename, tc.newExt)
		if result != tc.expected {
			t.Errorf("replaceExt(%s, %s) = %s; expected %s",
				tc.filename, tc.newExt, result, tc.expected)
		}
	}
}

func TestNewCompiler(t *testing.T) {
	cfg := &config.Config{
		Engine: "pdflatex",
	}

	compiler := NewCompiler(cfg)

	if compiler == nil {
		t.Fatalf("NewCompiler returned nil")
	}

	if compiler.Config != cfg {
		t.Errorf("NewCompiler did not correctly set the Config field")
	}
}

func TestCompile_InvalidInput(t *testing.T) {
	cfg := &config.Config{
		Engine: "pdflatex",
	}

	compiler := NewCompiler(cfg)

	// Test with empty file path
	_, err := compiler.Compile("")
	if err == nil {
		t.Errorf("Expected error for empty file path but got none")
	}

	// Test with non-existent file
	_, err = compiler.Compile("/path/to/nonexistent/file.tex")
	if err == nil {
		t.Errorf("Expected error for non-existent file but got none")
	}
}

// This test requires pdflatex to be installed on the system
// Skip if not available
func TestCompile_BasicDocument(t *testing.T) {
	// Check if pdflatex is installed
	_, err := exec.LookPath("pdflatex")
	if err != nil {
		t.Skip("pdflatex not found, skipping test")
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a minimal LaTeX document
	texContent := `
\documentclass{article}
\begin{document}
Hello, World!
\end{document}
`
	texFile := filepath.Join(tempDir, "test.tex")
	if err := os.WriteFile(texFile, []byte(texContent), 0644); err != nil {
		t.Fatalf("Failed to write test LaTeX file: %v", err)
	}

	// Create compiler with default config
	cfg := &config.Config{
		Engine: "pdflatex",
	}
	compiler := NewCompiler(cfg)

	// Compile the document
	pdfPath, err := compiler.Compile(texFile)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Check if PDF was created
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		t.Errorf("Expected PDF file at %s but it doesn't exist", pdfPath)
	}
}

// Test for custom output path
func TestCompile_CustomOutput(t *testing.T) {
	// Check if pdflatex is installed
	_, err := exec.LookPath("pdflatex")
	if err != nil {
		t.Skip("pdflatex not found, skipping test")
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a minimal LaTeX document
	texContent := `
\documentclass{article}
\begin{document}
Hello, World!
\end{document}
`
	texFile := filepath.Join(tempDir, "test.tex")
	if err := os.WriteFile(texFile, []byte(texContent), 0644); err != nil {
		t.Fatalf("Failed to write test LaTeX file: %v", err)
	}

	// Custom output path
	customOutput := filepath.Join(tempDir, "output", "custom.pdf")

	// Create compiler with custom output path
	cfg := &config.Config{
		Engine: "pdflatex",
		Output: config.Output(customOutput),
	}
	compiler := NewCompiler(cfg)

	// Create output directory
	if err := os.MkdirAll(filepath.Dir(customOutput), 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Compile the document
	pdfPath, err := compiler.Compile(texFile)

	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Check if the returned path is the custom output path
	if pdfPath != customOutput {
		t.Errorf("Expected PDF path to be %s, got %s", customOutput, pdfPath)
	}
}

// TODO: Implement BibTeX test in the future
// func TestCompile_WithBibtex(t *testing.T) {
// 	// This test is commented out for now as BibTeX integration
// 	// requires more complex setup and is not currently used
// 	// in the main application workflow
// }

func TestCompile_WithBibtex_InvalidInput(t *testing.T) {
	cfg := &config.Config{
		Engine: "pdflatex",
	}

	compiler := NewCompiler(cfg)

	// Test with empty file path
	_, err := compiler.CompileWithBibtex("")
	if err == nil {
		t.Errorf("Expected error for empty file path but got none")
	}

	// Test with non-existent file
	_, err = compiler.CompileWithBibtex("/path/to/nonexistent/file.tex")
	if err == nil {
		t.Errorf("Expected error for non-existent file but got none")
	}
}

func TestCompile_EngineNotInstalled(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a minimal LaTeX document
	texContent := `
\documentclass{article}
\begin{document}
Hello, World!
\end{document}
`
	texFile := filepath.Join(tempDir, "test.tex")
	if err := os.WriteFile(texFile, []byte(texContent), 0644); err != nil {
		t.Fatalf("Failed to write test LaTeX file: %v", err)
	}

	// Create compiler with non-existent engine
	cfg := &config.Config{
		Engine: "nonexistent-engine",
	}
	compiler := NewCompiler(cfg)

	// Compile should fail
	_, err = compiler.Compile(texFile)
	if err == nil {
		t.Error("Expected error for non-existent engine but got none")
	}
}

func TestCompile_DefaultEngine(t *testing.T) {
	// Check if pdflatex is installed
	_, err := exec.LookPath("pdflatex")
	if err != nil {
		t.Skip("pdflatex not found, skipping test")
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a minimal LaTeX document
	texContent := `
\documentclass{article}
\begin{document}
Hello, World!
\end{document}
`
	texFile := filepath.Join(tempDir, "test.tex")
	if err := os.WriteFile(texFile, []byte(texContent), 0644); err != nil {
		t.Fatalf("Failed to write test LaTeX file: %v", err)
	}

	// Create compiler with empty engine (should default to pdflatex)
	cfg := &config.Config{
		Engine: "",
	}
	compiler := NewCompiler(cfg)

	// Compile the document
	pdfPath, err := compiler.Compile(texFile)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Check if PDF was created
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		t.Errorf("Expected PDF file at %s but it doesn't exist", pdfPath)
	}
}

func TestCompile_NonPdfLatexEngine(t *testing.T) {
	// Check if xelatex is installed
	_, err := exec.LookPath("xelatex")
	if err != nil {
		t.Skip("xelatex not found, skipping test")
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a minimal LaTeX document
	texContent := `
\documentclass{article}
\begin{document}
Hello, World!
\end{document}
`
	texFile := filepath.Join(tempDir, "test.tex")
	if err := os.WriteFile(texFile, []byte(texContent), 0644); err != nil {
		t.Fatalf("Failed to write test LaTeX file: %v", err)
	}

	// Create compiler with xelatex engine
	cfg := &config.Config{
		Engine: "xelatex",
	}
	compiler := NewCompiler(cfg)

	// Compile the document
	pdfPath, err := compiler.Compile(texFile)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Check if PDF was created
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		t.Errorf("Expected PDF file at %s but it doesn't exist", pdfPath)
	}
}

func TestCompile_OutputPathNormalization(t *testing.T) {
	// Check if pdflatex is installed
	_, err := exec.LookPath("pdflatex")
	if err != nil {
		t.Skip("pdflatex not found, skipping test")
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "autopdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a minimal LaTeX document
	texContent := `
\documentclass{article}
\begin{document}
Hello, World!
\end{document}
`
	texFile := filepath.Join(tempDir, "test.tex")
	if err := os.WriteFile(texFile, []byte(texContent), 0644); err != nil {
		t.Fatalf("Failed to write test LaTeX file: %v", err)
	}

	// Test output path without .pdf extension
	outputPath := filepath.Join(tempDir, "output")
	cfg := &config.Config{
		Engine: "pdflatex",
		Output: config.Output(outputPath),
	}
	compiler := NewCompiler(cfg)

	// Compile the document
	pdfPath, err := compiler.Compile(texFile)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Check if PDF was created (the path should be normalized)
	if pdfPath != outputPath {
		t.Errorf("Expected PDF path to be %s, got %s", outputPath, pdfPath)
	}
}
