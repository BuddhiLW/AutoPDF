// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDebugYAMLParsing(t *testing.T) {
	// Simple test to debug YAML parsing
	yamlContent := `
template: "template.tex"
output: "output.pdf"
engine: "pdflatex"
variables:
  title: "My Document"
  author: "AutoPDF User"
  metadata:
    version: "1.0"
    tags: 
      - "example"
      - "complex"
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

	// Debug: Print what we got
	fmt.Printf("Config: %+v\n", config)
	fmt.Printf("Variables: %+v\n", config.Variables)
	if config.Variables.VariableSet != nil {
		fmt.Printf("VariableSet: %+v\n", config.Variables.VariableSet)
		fmt.Printf("Variables map: %+v\n", config.Variables.VariableSet.variables)
	}

	// Test basic access
	title, exists := config.Variables.GetString("title")
	fmt.Printf("Title: '%s' (exists: %v)\n", title, exists)

	author, exists := config.Variables.GetString("author")
	fmt.Printf("Author: '%s' (exists: %v)\n", author, exists)

	// Test nested access
	version, exists := config.Variables.GetString("metadata.version")
	fmt.Printf("Version: '%s' (exists: %v)\n", version, exists)

	// Test array access
	firstTag, exists := config.Variables.GetString("metadata.tags[0]")
	fmt.Printf("First tag: '%s' (exists: %v)\n", firstTag, exists)

	// Debug: Test parsePath function
	parts := parsePath("metadata.tags[0]")
	fmt.Printf("parsePath('metadata.tags[0]') = %v\n", parts)

	// Test step by step
	metadata, exists := config.Variables.GetByPath("metadata")
	if exists {
		tags, exists := metadata.Get("tags")
		if exists {
			firstTag, exists := tags.Get("0")
			fmt.Printf("Step by step: metadata -> tags -> 0 = '%s' (exists: %v)\n", firstTag.String(), exists)
		}
	}

	// Debug: Test direct access to metadata
	metadata2, exists := config.Variables.GetByPath("metadata")
	fmt.Printf("Metadata exists: %v\n", exists)
	if exists {
		fmt.Printf("Metadata type: %T\n", metadata2)
		// Try to get tags directly
		tags, exists := metadata2.Get("tags")
		fmt.Printf("Tags exists: %v, type: %T\n", exists, tags)
		if exists {
			// Try to get first tag
			firstTag, exists := tags.Get("0")
			fmt.Printf("First tag direct: '%s' (exists: %v)\n", firstTag.String(), exists)
		}
	}

	// Test flattening
	flattened := config.Variables.Flatten()
	fmt.Printf("Flattened: %+v\n", flattened)
}
