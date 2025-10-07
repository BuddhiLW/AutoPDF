// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package examples

import (
	"context"
	"fmt"
	"log"

	"github.com/BuddhiLW/AutoPDF/pkg/api/services"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// ExampleComplexVariables demonstrates how to use complex variables with AutoPDF
func ExampleComplexVariables() {
	// Create a config with complex variables
	cfg := &config.Config{
		Template:  config.Template("template.tex"),
		Output:    config.Output("output.pdf"),
		Engine:    config.Engine("pdflatex"),
		Variables: *config.NewVariables(),
	}

	// Set up complex nested variables
	// This represents the YAML structure:
	// variables:
	//   foo:
	//     bar:
	//       - bar1
	//       - bar2
	//     zet: [1, 2, 3]
	//   foo_bar: [foo, bar]

	// Create nested structure
	fooMap := config.NewMapVariable()

	// Set bar as an array
	barArray := config.NewSliceVariable()
	barArray.Values = []config.Variable{
		&config.StringVariable{Value: "bar1"},
		&config.StringVariable{Value: "bar2"},
	}
	fooMap.Set("bar", barArray)

	// Set zet as an array of numbers
	zetArray := config.NewSliceVariable()
	zetArray.Values = []config.Variable{
		&config.NumberVariable{Value: 1},
		&config.NumberVariable{Value: 2},
		&config.NumberVariable{Value: 3},
	}
	fooMap.Set("zet", zetArray)

	// Set the foo map
	cfg.Variables.Set("foo", fooMap)

	// Set foo_bar as a simple array
	fooBarArray := config.NewSliceVariable()
	fooBarArray.Values = []config.Variable{
		&config.StringVariable{Value: "foo"},
		&config.StringVariable{Value: "bar"},
	}
	cfg.Variables.Set("foo_bar", fooBarArray)

	// Create API service
	apiService := services.NewPDFGenerationAPIService(cfg)

	// Example 1: Using builder pattern with complex variables
	ctx := context.Background()

	// Build complex variables using the builder
	complexVars := map[string]interface{}{
		"title":  "My Document",
		"author": "AutoPDF User",
		"date":   "2025-01-07",
		"foo": map[string]interface{}{
			"bar": []string{"bar1", "bar2"},
			"zet": []int{1, 2, 3},
		},
		"foo_bar": []string{"foo", "bar"},
		"nested": map[string]interface{}{
			"deep": map[string]interface{}{
				"value":   "nested value",
				"numbers": []float64{1.5, 2.5, 3.5},
			},
		},
	}

	// Generate PDF with complex variables
	pdfBytes, imagePaths, err := apiService.GeneratePDF(ctx, "template.tex", "output.pdf", complexVars)
	if err != nil {
		log.Fatalf("Failed to generate PDF: %v", err)
	}

	fmt.Printf("Generated PDF: %d bytes\n", len(pdfBytes))
	fmt.Printf("Image paths: %v\n", imagePaths)
}

// ExampleBuilderPattern demonstrates the builder pattern for PDF generation
func ExampleBuilderPattern() {
	// Create config
	cfg := &config.Config{
		Template:  config.Template("template.tex"),
		Output:    config.Output("output.pdf"),
		Engine:    config.Engine("pdflatex"),
		Variables: *config.NewVariables(),
	}

	// Create API service
	apiService := services.NewPDFGenerationAPIService(cfg)

	// Build request using builder pattern
	options := services.NewPDFGenerationOptions("template.tex", "output.pdf").
		WithEngine("pdflatex").
		WithVariable("title", "Builder Pattern Example").
		WithVariable("author", "AutoPDF").
		WithVariable("date", "2025-01-07").
		WithVariable("metadata", map[string]interface{}{
			"version": "1.0",
			"tags":    []string{"example", "builder", "pattern"},
			"settings": map[string]interface{}{
				"verbose": true,
				"debug":   false,
			},
		}).
		WithConversion(true, "png", "jpeg").
		WithCleanup(false).
		WithTimeout(30).
		WithVerbose(true)

	// Generate PDF
	ctx := context.Background()
	pdfBytes, imagePaths, err := apiService.GeneratePDFWithOptions(ctx, *options)
	if err != nil {
		log.Fatalf("Failed to generate PDF: %v", err)
	}

	fmt.Printf("Generated PDF: %d bytes\n", len(pdfBytes))
	fmt.Printf("Image paths: %v\n", imagePaths)
}

// ExampleTemplateVariables demonstrates template variable extraction
func ExampleTemplateVariables() {
	cfg := &config.Config{
		Template:  config.Template("template.tex"),
		Output:    config.Output("output.pdf"),
		Engine:    config.Engine("pdflatex"),
		Variables: *config.NewVariables(),
	}

	apiService := services.NewPDFGenerationAPIService(cfg)

	// Extract variables from template
	variables, err := apiService.GetTemplateVariables("template.tex")
	if err != nil {
		log.Fatalf("Failed to extract template variables: %v", err)
	}

	fmt.Printf("Template variables: %v\n", variables)
}

// ExampleVariableFlattening demonstrates variable flattening
func ExampleVariableFlattening() {
	// Create complex variables
	complexVars := map[string]interface{}{
		"title":  "My Document",
		"author": "AutoPDF User",
		"metadata": map[string]interface{}{
			"version": "1.0",
			"tags":    []string{"example", "flattening"},
			"settings": map[string]interface{}{
				"verbose": true,
				"debug":   false,
			},
		},
		"items": []interface{}{
			"item1",
			"item2",
			map[string]interface{}{
				"name":  "nested item",
				"value": 42,
			},
		},
	}

	// Create config and set variables
	cfg := &config.Config{
		Template:  config.Template("template.tex"),
		Output:    config.Output("output.pdf"),
		Engine:    config.Engine("pdflatex"),
		Variables: *config.NewVariables(),
	}

	// Set complex variables
	for key, value := range complexVars {
		cfg.Variables.SetString(key, fmt.Sprintf("%v", value))
	}

	// Flatten variables
	flattened := cfg.Variables.Flatten()

	fmt.Println("Flattened variables:")
	for key, value := range flattened {
		fmt.Printf("  %s: %s\n", key, value)
	}
}

// ExampleTemplateProcessing demonstrates template processing with complex variables
func ExampleTemplateProcessing() {
	// Template content with complex variable references
	// (In real usage, this would be an existing template file)

	// Write template to file
	// (In real usage, this would be an existing template file)

	// Create variables
	variables := map[string]interface{}{
		"title":  "Complex Variables Example",
		"author": "AutoPDF",
		"date":   "2025-01-07",
		"metadata": map[string]interface{}{
			"version": "1.0",
			"tags":    []string{"example", "complex"},
			"settings": map[string]interface{}{
				"verbose": true,
				"debug":   false,
			},
		},
		"items": []interface{}{
			"First item",
			"Second item",
			map[string]interface{}{
				"name":  "Nested item",
				"value": 42,
			},
		},
	}

	// Create config
	cfg := &config.Config{
		Template:  config.Template("template.tex"),
		Output:    config.Output("output.pdf"),
		Engine:    config.Engine("pdflatex"),
		Variables: *config.NewVariables(),
	}

	// Create API service
	apiService := services.NewPDFGenerationAPIService(cfg)

	// Generate PDF
	ctx := context.Background()
	pdfBytes, imagePaths, err := apiService.GeneratePDF(ctx, "template.tex", "output.pdf", variables)
	if err != nil {
		log.Fatalf("Failed to generate PDF: %v", err)
	}

	fmt.Printf("Generated PDF: %d bytes\n", len(pdfBytes))
	fmt.Printf("Image paths: %v\n", imagePaths)
}
