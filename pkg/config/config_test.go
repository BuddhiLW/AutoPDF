package config

import (
	"reflect"
	"strings"
	"testing"
)

func TestNewConfigFromYAML(t *testing.T) {
	// Create a temporary YAML file
	yamlData := []byte(`
template: "test.tex"
output: "output.pdf"
variables:
  title: "Test Document"
  author: "Test User"
engine: "pdflatex"
conversion:
  enabled: true
  formats:
    - "png"
    - "jpg"
`)

	// Test parsing the config
	cfg, err := NewConfigFromYAML(yamlData)
	if err != nil {
		t.Fatalf("Failed to parse YAML config: %v", err)
	}

	// Verify parsed values
	if cfg.Template != "test.tex" {
		t.Errorf("Expected Template to be 'test.tex', got '%s'", cfg.Template)
	}

	if cfg.Output != "output.pdf" {
		t.Errorf("Expected Output to be 'output.pdf', got '%s'", cfg.Output)
	}

	if cfg.Engine != "pdflatex" {
		t.Errorf("Expected Engine to be 'pdflatex', got '%s'", cfg.Engine)
	}

	if !cfg.Conversion.Enabled {
		t.Errorf("Expected Conversion.Enabled to be true")
	}

	if len(cfg.Conversion.Formats) != 2 || cfg.Conversion.Formats[0] != "png" || cfg.Conversion.Formats[1] != "jpg" {
		t.Errorf("Conversion formats not parsed correctly, got %v", cfg.Conversion.Formats)
	}

	expectedVars := map[string]string{
		"title":  "Test Document",
		"author": "Test User",
	}

	if !reflect.DeepEqual(cfg.Variables, expectedVars) {
		t.Errorf("Variables not parsed correctly. Expected %v, got %v", expectedVars, cfg.Variables)
	}
}

func TestNewConfigFromYAML_Defaults(t *testing.T) {
	// Test with minimal config that should use defaults
	yamlData := []byte(`
template: "test.tex"
output: "output.pdf"
`)

	cfg, err := NewConfigFromYAML(yamlData)
	if err != nil {
		t.Fatalf("Failed to parse minimal YAML config: %v", err)
	}

	// Check that defaults are applied
	if cfg.Engine != "pdflatex" {
		t.Errorf("Default Engine should be 'pdflatex', got '%s'", cfg.Engine)
	}

	if cfg.Variables.String() != "{}" {
		t.Errorf("Variables should be initialized to empty map, got %v", cfg.Variables)
	}
}

func TestToJSON(t *testing.T) {
	cfg := &Config{
		Template: "test.tex",
		Output:   "output.pdf",
		Engine:   "pdflatex",
		Variables: Variables(map[string]interface{}{
			"title": "Test Document",
		}),
		Conversion: Conversion{
			Enabled: true,
			Formats: []string{"png"},
		},
	}

	json, err := cfg.ToJSON()
	if err != nil {
		t.Fatalf("Failed to convert config to JSON: %v", err)
	}

	// Check that JSON contains expected values
	expectedSubstrings := []string{
		`"template": "test.tex"`,
		`"output": "output.pdf"`,
		`"engine": "pdflatex"`,
		`"title": "Test Document"`,
		`"enabled": true`,
		`"formats": [`,
		`"png"`,
	}

	for _, substr := range expectedSubstrings {
		if !contains(json, substr) {
			t.Errorf("Expected JSON to contain '%s', but it doesn't: %s", substr, json)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
