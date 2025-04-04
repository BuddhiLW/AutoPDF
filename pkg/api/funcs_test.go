package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

func TestGeneratePDF(t *testing.T) {
	// Setup test data
	defaultCfg := config.GetDefaultConfig()
	testCfg := &config.Config{
		Template: "./test_data/template.tex",
		Variables: map[string]string{
			"title":   "Test Title: api call",
			"content": "Test Content: api call",
		},
	}

	// Call the function
	pdfBytes, err := GeneratePDF(testCfg, defaultCfg.Template)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check if the PDF bytes are not nil or empty
	if len(pdfBytes) == 0 {
		t.Fatal("Expected non-empty PDF bytes")
	}

	// Clean up any generated files if necessary
	outputFile := filepath.Join(os.TempDir(), "out/output.pdf")
	if err := os.Remove(outputFile); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to clean up output file: %v", err)
	}
}
