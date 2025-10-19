// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package examples

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/BuddhiLW/AutoPDF/pkg/converter"
)

// Example structs for demonstration
type Address struct {
	Street string `autopdf:"street"`
	City   string `autopdf:"city"`
	State  string `autopdf:"state"`
	Zip    string `autopdf:"zip"`
}

type Contact struct {
	Email   string   `autopdf:"email"`
	Phone   string   `autopdf:"phone"`
	Website *url.URL `autopdf:"website"`
}

type User struct {
	ID        int       `autopdf:"id"`
	Name      string    `autopdf:"name"`
	Email     string    `autopdf:"email"`
	CreatedAt time.Time `autopdf:"created_at"`
	Address   Address   `autopdf:"address"`
	Contact   Contact   `autopdf:"contact"`
	Tags      []string  `autopdf:"tags"`
	IsActive  bool      `autopdf:"is_active"`
}

type FlattenedUser struct {
	ID        int       `autopdf:"id"`
	Name      string    `autopdf:"name"`
	Email     string    `autopdf:"email"`
	CreatedAt time.Time `autopdf:"created_at"`
	Address   Address   `autopdf:"address,flatten"`
	Contact   Contact   `autopdf:"contact,flatten"`
	Tags      []string  `autopdf:"tags,flatten"`
	IsActive  bool      `autopdf:"is_active"`
}

type InlinedUser struct {
	ID        int       `autopdf:"id"`
	Name      string    `autopdf:"name"`
	Email     string    `autopdf:"email"`
	CreatedAt time.Time `autopdf:"created_at"`
	Address   Address   `autopdf:"address,inline"`
	Contact   Contact   `autopdf:"contact,inline"`
	Tags      []string  `autopdf:"tags,flatten"`
	IsActive  bool      `autopdf:"is_active"`
}

type Document struct {
	Title       string                 `autopdf:"title"`
	Author      User                   `autopdf:"author"`
	CreatedAt   time.Time              `autopdf:"created_at"`
	UpdatedAt   *time.Time             `autopdf:"updated_at"`
	Content     string                 `autopdf:"content"`
	Tags        []string               `autopdf:"tags"`
	Categories  []string               `autopdf:"categories,flatten"`
	Metadata    map[string]interface{} `autopdf:"metadata"`
	IsPublished bool                   `autopdf:"is_published"`
}

type CustomFormattable struct {
	Value string
	Type  string
}

func (cf CustomFormattable) ToAutoPDFVariable() (config.Variable, error) {
	return &config.StringVariable{Value: fmt.Sprintf("%s:%s", cf.Type, cf.Value)}, nil
}

type CustomDocument struct {
	Title  string            `autopdf:"title"`
	Custom CustomFormattable `autopdf:"custom"`
	Normal string            `autopdf:"normal"`
}

// ExampleStructConversion demonstrates basic struct conversion
func ExampleStructConversion() {
	fmt.Println("=== Basic Struct Conversion Example ===")

	// Create a sample user
	user := User{
		ID:        1,
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: time.Now(),
		Address: Address{
			Street: "123 Main St",
			City:   "Anytown",
			State:  "CA",
			Zip:    "12345",
		},
		Contact: Contact{
			Email:   "john@example.com",
			Phone:   "+1-555-123-4567",
			Website: &url.URL{Scheme: "https", Host: "johndoe.com"},
		},
		Tags:     []string{"developer", "golang", "autopdf"},
		IsActive: true,
	}

	// Create converter with built-in type support
	converter := converter.BuildWithDefaults()

	// Convert struct to Variables
	variables, err := converter.ConvertStruct(user)
	if err != nil {
		log.Fatalf("Failed to convert struct: %v", err)
	}

	// Display the converted variables
	fmt.Printf("Converted %d variables:\n", variables.Len())
	variables.Range(func(name string, value config.Variable) bool {
		fmt.Printf("  %s: %s (type: %v)\n", name, value.String(), value.Type())
		return true
	})

	// Show flattened representation
	flattened := variables.Flatten()
	fmt.Printf("\nFlattened representation:\n")
	for key, value := range flattened {
		fmt.Printf("  %s: %s\n", key, value)
	}
}

// ExampleNestedStructs demonstrates nested struct handling
func ExampleNestedStructs() {
	fmt.Println("\n=== Nested Structs Example ===")

	// Create a document with nested user
	now := time.Now()
	document := Document{
		Title: "AutoPDF Struct Conversion Guide",
		Author: User{
			ID:        1,
			Name:      "Jane Smith",
			Email:     "jane@example.com",
			CreatedAt: now.Add(-24 * time.Hour),
			Address: Address{
				Street: "456 Oak Ave",
				City:   "Tech City",
				State:  "CA",
				Zip:    "90210",
			},
			Contact: Contact{
				Email:   "jane@example.com",
				Phone:   "+1-555-987-6543",
				Website: &url.URL{Scheme: "https", Host: "janesmith.dev"},
			},
			Tags:     []string{"author", "technical-writer"},
			IsActive: true,
		},
		CreatedAt:  now,
		UpdatedAt:  &now,
		Content:    "This is a comprehensive guide to AutoPDF struct conversion...",
		Tags:       []string{"documentation", "golang", "autopdf"},
		Categories: []string{"tutorial", "guide", "reference"},
		Metadata: map[string]interface{}{
			"version": "1.0",
			"status":  "published",
			"views":   42,
		},
		IsPublished: true,
	}

	// Test different conversion strategies
	strategies := []struct {
		name      string
		converter *converter.StructConverter
	}{
		{"Default (Nested)", converter.BuildWithDefaults()},
		{"Flattened", converter.BuildForFlattened()},
		{"Template Optimized", converter.BuildForTemplates()},
	}

	for _, strategy := range strategies {
		fmt.Printf("\n--- %s Strategy ---\n", strategy.name)

		variables, err := strategy.converter.ConvertStruct(document)
		if err != nil {
			log.Printf("Failed to convert with %s: %v", strategy.name, err)
			continue
		}

		// Show key variables
		keyVars := []string{"title", "author.name", "author.address.street", "tags", "categories"}
		for _, key := range keyVars {
			if val, exists := variables.Get(key); exists {
				fmt.Printf("  %s: %s\n", key, val.String())
			}
		}
	}
}

// ExampleCustomConverter demonstrates custom type conversion
func ExampleCustomConverter() {
	fmt.Println("\n=== Custom Converter Example ===")

	// Create a document with custom formattable type
	customDoc := CustomDocument{
		Title: "Custom Type Example",
		Custom: CustomFormattable{
			Value: "important-data",
			Type:  "secret",
		},
		Normal: "regular string",
	}

	// Convert with default converter
	converter := converter.BuildWithDefaults()
	variables, err := converter.ConvertStruct(customDoc)
	if err != nil {
		log.Fatalf("Failed to convert custom struct: %v", err)
	}

	fmt.Printf("Custom document variables:\n")
	variables.Range(func(name string, value config.Variable) bool {
		fmt.Printf("  %s: %s\n", name, value.String())
		return true
	})
}

// ExampleArrayFlattening demonstrates array and slice handling
func ExampleArrayFlattening() {
	fmt.Println("\n=== Array Flattening Example ===")

	// Create users with different flattening strategies
	users := []struct {
		name      string
		user      interface{}
		converter *converter.StructConverter
	}{
		{
			"Default (Nested)",
			User{
				Name:    "Alice",
				Tags:    []string{"admin", "developer"},
				Address: Address{Street: "789 Pine St", City: "Dev City"},
			},
			converter.BuildWithDefaults(),
		},
		{
			"Flattened",
			FlattenedUser{
				Name:    "Bob",
				Tags:    []string{"user", "tester"},
				Address: Address{Street: "321 Elm St", City: "Test City"},
			},
			converter.BuildWithDefaults(),
		},
		{
			"Inlined",
			InlinedUser{
				Name:    "Charlie",
				Tags:    []string{"guest", "viewer"},
				Address: Address{Street: "654 Maple St", City: "View City"},
			},
			converter.BuildWithDefaults(),
		},
	}

	for _, test := range users {
		fmt.Printf("\n--- %s ---\n", test.name)

		variables, err := test.converter.ConvertStruct(test.user)
		if err != nil {
			log.Printf("Failed to convert %s: %v", test.name, err)
			continue
		}

		// Show address-related variables
		addressVars := []string{"street", "city", "address.street", "address.city"}
		for _, key := range addressVars {
			if val, exists := variables.Get(key); exists {
				fmt.Printf("  %s: %s\n", key, val.String())
			}
		}

		// Show tags
		if val, exists := variables.Get("tags"); exists {
			fmt.Printf("  tags: %s (type: %v)\n", val.String(), val.Type())
		}
	}
}

// ExampleIntegrationWithAutoPDF demonstrates integration with AutoPDF API
func ExampleIntegrationWithAutoPDF() {
	fmt.Println("\n=== AutoPDF Integration Example ===")

	// Create a document struct
	document := Document{
		Title:       "AutoPDF Integration Test",
		CreatedAt:   time.Now(),
		Content:     "This document demonstrates AutoPDF integration with struct conversion.",
		Tags:        []string{"integration", "test", "autopdf"},
		IsPublished: true,
	}

	// This example demonstrates how to convert a struct to variables
	// that can be used with AutoPDF's PDF generation system

	// Convert struct to variables using converter
	converter := converter.BuildForTemplates()
	variables, err := converter.ConvertStruct(document)
	if err != nil {
		log.Fatalf("Failed to convert document struct: %v", err)
	}

	// Convert Variables to map[string]interface{} for API
	variablesMap := make(map[string]interface{})
	variables.Range(func(name string, value config.Variable) bool {
		variablesMap[name] = value.String()
		return true
	})

	// Convert to TemplateVariables for API usage
	templateVars := generation.NewTemplateVariables(variables)

	// Use with AutoPDF API
	request := generation.PDFGenerationRequest{
		TemplatePath: "template.tex",
		Variables:    templateVars,
		Engine:       "pdflatex",
		OutputPath:   "output.pdf",
		Options: generation.PDFGenerationOptions{
			DoConvert: false,
			DoClean:   true,
			Timeout:   30 * time.Second,
			Verbose:   1,
		},
	}

	// Generate PDF (this would normally create a real PDF)
	fmt.Printf("Generated variables for AutoPDF:\n")
	for key, value := range variablesMap {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Printf("\nPDF Generation Request prepared with %d variables\n", len(variablesMap))
	fmt.Printf("Template: %s\n", request.TemplatePath)
	fmt.Printf("Output: %s\n", request.OutputPath)
	fmt.Printf("Engine: %s\n", request.Engine)
}

// RunAllExamples runs all struct conversion examples
func RunAllExamples() {
	fmt.Println("AutoPDF Struct Conversion Examples")
	fmt.Println("===================================")

	ExampleStructConversion()
	ExampleNestedStructs()
	ExampleCustomConverter()
	ExampleArrayFlattening()
	ExampleIntegrationWithAutoPDF()

	fmt.Println("\n=== All Examples Completed ===")
}
