// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestComplexVariablesYAML(t *testing.T) {
	// Test YAML with complex nested variables
	yamlContent := `
template: "template.tex"
output: "output.pdf"
engine: "pdflatex"
variables:
  title: "My Document"
  author: "AutoPDF User"
  date: "2025-01-07"
  metadata:
    version: "1.0"
    tags: 
      - "example"
      - "complex"
      - "variables"
    settings:
      verbose: true
      debug: false
      timeout: 30
  items:
    - "First item"
    - "Second item"
    - name: "Nested item"
      value: 42
      enabled: true
  foo:
    bar: 
      - "bar1"
      - "bar2"
    zet: [1, 2, 3]
  foo_bar: ["foo", "bar"]
conversion:
  enabled: true
  formats: ["png", "jpeg"]
`

	// Parse YAML
	var config Config
	err := yaml.Unmarshal([]byte(yamlContent), &config)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	// Test basic fields
	if config.Template.String() != "template.tex" {
		t.Errorf("Expected template 'template.tex', got '%s'", config.Template.String())
	}
	if config.Output.String() != "output.pdf" {
		t.Errorf("Expected output 'output.pdf', got '%s'", config.Output.String())
	}
	if config.Engine.String() != "pdflatex" {
		t.Errorf("Expected engine 'pdflatex', got '%s'", config.Engine.String())
	}

	// Test conversion settings
	if !config.Conversion.Enabled {
		t.Error("Expected conversion to be enabled")
	}
	if len(config.Conversion.Formats) != 2 {
		t.Errorf("Expected 2 formats, got %d", len(config.Conversion.Formats))
	}

	// Test complex variables
	if config.Variables.VariableSet == nil {
		t.Fatal("Variables should not be nil")
	}

	// Test simple variables
	title, exists := config.Variables.GetString("title")
	if !exists || title != "My Document" {
		t.Errorf("Expected title 'My Document', got '%s' (exists: %v)", title, exists)
	}

	author, exists := config.Variables.GetString("author")
	if !exists || author != "AutoPDF User" {
		t.Errorf("Expected author 'AutoPDF User', got '%s' (exists: %v)", author, exists)
	}

	// Test nested variables
	version, exists := config.Variables.GetString("metadata.version")
	if !exists || version != "1.0" {
		t.Errorf("Expected metadata.version '1.0', got '%s' (exists: %v)", version, exists)
	}

	// Test array variables
	firstTag, exists := config.Variables.GetString("metadata.tags[0]")
	if !exists || firstTag != "example" {
		t.Errorf("Expected metadata.tags[0] 'example', got '%s' (exists: %v)", firstTag, exists)
	}

	// Test nested object in array
	nestedName, exists := config.Variables.GetString("items[2].name")
	if !exists || nestedName != "Nested item" {
		t.Errorf("Expected items[2].name 'Nested item', got '%s' (exists: %v)", nestedName, exists)
	}

	// Test flattening
	flattened := config.Variables.Flatten()
	
	// Check some flattened keys
	expectedKeys := []string{
		"title",
		"author", 
		"date",
		"metadata.version",
		"metadata.tags[0]",
		"metadata.tags[1]",
		"metadata.tags[2]",
		"metadata.settings.verbose",
		"metadata.settings.debug",
		"metadata.settings.timeout",
		"items[0]",
		"items[1]",
		"items[2].name",
		"items[2].value",
		"items[2].enabled",
		"foo.bar[0]",
		"foo.bar[1]",
		"foo.zet[0]",
		"foo.zet[1]",
		"foo.zet[2]",
		"foo_bar[0]",
		"foo_bar[1]",
	}

	for _, key := range expectedKeys {
		if _, exists := flattened[key]; !exists {
			t.Errorf("Expected flattened key '%s' not found", key)
		}
	}

	// Test specific flattened values
	if flattened["title"] != "My Document" {
		t.Errorf("Expected flattened title 'My Document', got '%s'", flattened["title"])
	}
	if flattened["metadata.version"] != "1.0" {
		t.Errorf("Expected flattened metadata.version '1.0', got '%s'", flattened["metadata.version"])
	}
	if flattened["metadata.settings.verbose"] != "true" {
		t.Errorf("Expected flattened metadata.settings.verbose 'true', got '%s'", flattened["metadata.settings.verbose"])
	}
	if flattened["items[2].name"] != "Nested item" {
		t.Errorf("Expected flattened items[2].name 'Nested item', got '%s'", flattened["items[2].name"])
	}
}

func TestComplexVariablesYAMLMarshal(t *testing.T) {
	// Create a config with complex variables
	cfg := &Config{
		Template: Template("template.tex"),
		Output:   Output("output.pdf"),
		Engine:   Engine("pdflatex"),
		Variables: *NewVariables(),
		Conversion: Conversion{
			Enabled: true,
			Formats: []string{"png", "jpeg"},
		},
	}

	// Set complex variables
	cfg.Variables.SetString("title", "My Document")
	cfg.Variables.SetString("author", "AutoPDF User")
	cfg.Variables.SetString("date", "2025-01-07")
	
	// Set nested metadata
	metadata := NewMapVariable()
	metadata.Set("version", &StringVariable{Value: "1.0"})
	
	tags := NewSliceVariable()
	tags.Values = []Variable{
		&StringVariable{Value: "example"},
		&StringVariable{Value: "complex"},
		&StringVariable{Value: "variables"},
	}
	metadata.Set("tags", tags)
	
	settings := NewMapVariable()
	settings.Set("verbose", &BoolVariable{Value: true})
	settings.Set("debug", &BoolVariable{Value: false})
	settings.Set("timeout", &NumberVariable{Value: 30})
	metadata.Set("settings", settings)
	
	cfg.Variables.Set("metadata", metadata)

	// Marshal to YAML
	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal to YAML: %v", err)
	}

	// Parse it back
	var parsedConfig Config
	err = yaml.Unmarshal(yamlData, &parsedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	// Verify the round-trip worked
	title, exists := parsedConfig.Variables.GetString("title")
	if !exists || title != "My Document" {
		t.Errorf("Round-trip failed for title: got '%s' (exists: %v)", title, exists)
	}

	version, exists := parsedConfig.Variables.GetString("metadata.version")
	if !exists || version != "1.0" {
		t.Errorf("Round-trip failed for metadata.version: got '%s' (exists: %v)", version, exists)
	}

	verbose, exists := parsedConfig.Variables.GetString("metadata.settings.verbose")
	if !exists || verbose != "true" {
		t.Errorf("Round-trip failed for metadata.settings.verbose: got '%s' (exists: %v)", verbose, exists)
	}
}

func TestComplexVariablesCLI(t *testing.T) {
	// Test that the CLI can handle complex variables
	// This simulates what happens when the CLI loads a config file
	
	yamlContent := `
template: "template.tex"
output: "output.pdf"
engine: "pdflatex"
variables:
  title: "CLI Test Document"
  author: "AutoPDF CLI"
  metadata:
    version: "2.0"
    features:
      - "complex variables"
      - "yaml parsing"
      - "cli integration"
    settings:
      verbose: true
      debug: false
  items:
    - name: "Feature 1"
      enabled: true
    - name: "Feature 2" 
      enabled: false
conversion:
  enabled: false
  formats: []
`

	// Parse the config (this is what the CLI does)
	var config Config
	err := yaml.Unmarshal([]byte(yamlContent), &config)
	if err != nil {
		t.Fatalf("CLI config parsing failed: %v", err)
	}

	// Test that the CLI can access variables
	title, exists := config.Variables.GetString("title")
	if !exists || title != "CLI Test Document" {
		t.Errorf("CLI title access failed: got '%s' (exists: %v)", title, exists)
	}

	// Test nested access
	version, exists := config.Variables.GetString("metadata.version")
	if !exists || version != "2.0" {
		t.Errorf("CLI nested access failed: got '%s' (exists: %v)", version, exists)
	}

	// Test array access
	firstFeature, exists := config.Variables.GetString("metadata.features[0]")
	if !exists || firstFeature != "complex variables" {
		t.Errorf("CLI array access failed: got '%s' (exists: %v)", firstFeature, exists)
	}

	// Test nested object in array
	feature1Name, exists := config.Variables.GetString("items[0].name")
	if !exists || feature1Name != "Feature 1" {
		t.Errorf("CLI nested array access failed: got '%s' (exists: %v)", feature1Name, exists)
	}

	// Test that flattening works for template processing
	flattened := config.Variables.Flatten()
	
	// Verify key flattened variables exist
	expectedFlattened := map[string]string{
		"title": "CLI Test Document",
		"author": "AutoPDF CLI",
		"metadata.version": "2.0",
		"metadata.features[0]": "complex variables",
		"metadata.features[1]": "yaml parsing",
		"metadata.features[2]": "cli integration",
		"metadata.settings.verbose": "true",
		"metadata.settings.debug": "false",
		"items[0].name": "Feature 1",
		"items[0].enabled": "true",
		"items[1].name": "Feature 2",
		"items[1].enabled": "false",
	}

	for key, expectedValue := range expectedFlattened {
		if actualValue, exists := flattened[key]; !exists {
			t.Errorf("Expected flattened key '%s' not found", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected flattened value for '%s' to be '%s', got '%s'", key, expectedValue, actualValue)
		}
	}
}
