// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package converter

import (
	"fmt"
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// DemoStruct demonstrates the struct conversion feature
type DemoStruct struct {
	Title     string    `autopdf:"title"`
	Author    string    `autopdf:"author"`
	CreatedAt time.Time `autopdf:"created_at"`
	Tags      []string  `autopdf:"tags,flatten"`
	IsActive  bool      `autopdf:"is_active"`
}

// RunDemo demonstrates the AutoPDF struct conversion feature
func RunDemo() {
	fmt.Println("AutoPDF Struct Conversion Demo")
	fmt.Println("==============================")

	// Create a sample struct
	demo := DemoStruct{
		Title:     "AutoPDF Struct Conversion",
		Author:    "AutoPDF Team",
		CreatedAt: time.Now(),
		Tags:      []string{"demo", "conversion", "autopdf"},
		IsActive:  true,
	}

	// Create converter with built-in type support
	converter := BuildWithDefaults()

	// Convert struct to Variables
	variables, err := converter.ConvertStruct(demo)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
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

	fmt.Println("\nDemo completed successfully!")
}

