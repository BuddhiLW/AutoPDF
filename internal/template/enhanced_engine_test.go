package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/pkg/domain"
)

func TestNewEnhancedEngine(t *testing.T) {
	config := &EnhancedConfig{
		TemplatePath: "test.tex",
		OutputPath:   "output.tex",
		Engine:       "pdflatex",
		Delimiters: DelimiterConfig{
			Left:  "delim[[",
			Right: "]]",
		},
		Functions: make(map[string]interface{}),
	}

	engine := NewEnhancedEngine(config)
	if engine == nil {
		t.Fatal("Expected engine to be created")
	}

	if engine.Context == nil {
		t.Error("Expected context to be initialized")
	}

	if engine.Config != config {
		t.Error("Expected config to be set")
	}
}

func TestSetVariable(t *testing.T) {
	config := &EnhancedConfig{}
	engine := NewEnhancedEngine(config)

	variable := domain.NewStringVariable("test")
	err := engine.SetVariable("key", variable)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify the variable was set
	retrieved, err := engine.GetVariable("key")
	if err != nil {
		t.Errorf("Unexpected error retrieving variable: %v", err)
	}

	if retrieved != variable {
		t.Error("Expected retrieved variable to be the same as set")
	}
}

func TestSetVariablesFromMap(t *testing.T) {
	config := &EnhancedConfig{}
	engine := NewEnhancedEngine(config)

	variables := map[string]interface{}{
		"string":  "test",
		"number":  42,
		"boolean": true,
		"object": map[string]interface{}{
			"nested": "value",
		},
		"array": []interface{}{"item1", "item2"},
	}

	err := engine.SetVariablesFromMap(variables)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify all variables were set
	for key, expectedValue := range variables {
		variable, err := engine.GetVariable(key)
		if err != nil {
			t.Errorf("Error retrieving variable '%s': %v", key, err)
			continue
		}

		// For complex types like maps and slices, we need to compare differently
		switch key {
		case "object":
			// For objects, just verify the type is correct
			if variable.Type != domain.VariableTypeObject {
				t.Errorf("Expected object type for key '%s', got %s", key, variable.Type)
			}
		case "array":
			// For arrays, just verify the type is correct
			if variable.Type != domain.VariableTypeArray {
				t.Errorf("Expected array type for key '%s', got %s", key, variable.Type)
			}
		default:
			// For simple types, compare values directly
			if variable.Value != expectedValue {
				t.Errorf("Expected value %v for key '%s', got %v", expectedValue, key, variable.Value)
			}
		}
	}
}

func TestEnhancedProcess(t *testing.T) {
	// Create a temporary template file
	tempDir, err := os.MkdirTemp("", "autopdf-enhanced-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template with complex data structures
	templateContent := `
\documentclass{article}
\title{delim[[.title]]}
\author{delim[[.author]]}

\begin{document}
\maketitle

\section{User Information}
Name: delim[[.user.name]]
Age: delim[[.user.age]]
Email: delim[[.user.email]]

\section{Address}
Street: delim[[.user.address.street]]
City: delim[[.user.address.city]]
Country: delim[[.user.address.country]]

\section{Skills}
delim[[range .user.skills]]
- delim[[.name]]: delim[[.level]]
delim[[end]]

\section{Projects}
delim[[range .user.projects]]
\subsection{delim[[.name]]}
Description: delim[[.description]]
Technologies: delim[[join ", " .technologies]]
delim[[end]]

\end{document}
`

	templatePath := filepath.Join(tempDir, "test-template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Create enhanced engine with complex data
	config := &EnhancedConfig{
		TemplatePath: templatePath,
		Delimiters: DelimiterConfig{
			Left:  "delim[[",
			Right: "]]",
		},
	}
	engine := NewEnhancedEngine(config)

	// Set complex variables
	variables := map[string]interface{}{
		"title":  "Enhanced Template Test",
		"author": "AutoPDF Team",
		"user": map[string]interface{}{
			"name":  "John Doe",
			"age":   30,
			"email": "john@example.com",
			"address": map[string]interface{}{
				"street":  "123 Main St",
				"city":    "New York",
				"country": "USA",
			},
			"skills": []interface{}{
				map[string]interface{}{
					"name":  "Go",
					"level": "Expert",
				},
				map[string]interface{}{
					"name":  "LaTeX",
					"level": "Advanced",
				},
			},
			"projects": []interface{}{
				map[string]interface{}{
					"name":         "AutoPDF",
					"description":  "PDF generation library",
					"technologies": []interface{}{"Go", "LaTeX", "Templates"},
				},
				map[string]interface{}{
					"name":         "WebApp",
					"description":  "Web application",
					"technologies": []interface{}{"React", "Node.js", "PostgreSQL"},
				},
			},
		},
	}

	err = engine.SetVariablesFromMap(variables)
	if err != nil {
		t.Fatalf("Failed to set variables: %v", err)
	}

	// Process the template
	result, err := engine.Process(templatePath)
	if err != nil {
		t.Fatalf("Template processing failed: %v", err)
	}

	// Verify that variables were properly substituted
	expectedSubstrings := []string{
		`\title{Enhanced Template Test}`,
		`\author{AutoPDF Team}`,
		`Name: John Doe`,
		`Age: 30`,
		`Email: john@example.com`,
		`Street: 123 Main St`,
		`City: New York`,
		`Country: USA`,
		`- Go: Expert`,
		`- LaTeX: Advanced`,
		`\subsection{AutoPDF}`,
		`Description: PDF generation library`,
		`Technologies: Go, LaTeX, Templates`,
		`\subsection{WebApp}`,
		`Description: Web application`,
		`Technologies: React, Node.js, PostgreSQL`,
	}

	for _, substr := range expectedSubstrings {
		if !strings.Contains(result, substr) {
			t.Errorf("Expected result to contain '%s', but it doesn't", substr)
		}
	}
}

func TestEnhancedProcessToFile(t *testing.T) {
	// Create a temporary template file
	tempDir, err := os.MkdirTemp("", "autopdf-enhanced-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple test template
	templateContent := `\title{delim[[.title]]}`
	templatePath := filepath.Join(tempDir, "test-template.tex")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	outputPath := filepath.Join(tempDir, "output.tex")

	// Create enhanced engine
	config := &EnhancedConfig{
		TemplatePath: templatePath,
		OutputPath:   outputPath,
		Delimiters: DelimiterConfig{
			Left:  "delim[[",
			Right: "]]",
		},
	}
	engine := NewEnhancedEngine(config)

	// Set variables
	variables := map[string]interface{}{
		"title": "Enhanced Test Document",
	}

	err = engine.SetVariablesFromMap(variables)
	if err != nil {
		t.Fatalf("Failed to set variables: %v", err)
	}

	// Process template to file
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

	expectedContent := `\title{Enhanced Test Document}`
	if string(outputContent) != expectedContent {
		t.Errorf("Output file content doesn't match. Expected '%s', got '%s'", expectedContent, string(outputContent))
	}
}

func TestValidateTemplate(t *testing.T) {
	config := &EnhancedConfig{}
	engine := NewEnhancedEngine(config)

	// Set some variables
	variables := map[string]interface{}{
		"title":  "Test Document",
		"author": "Test Author",
	}

	err := engine.SetVariablesFromMap(variables)
	if err != nil {
		t.Fatalf("Failed to set variables: %v", err)
	}

	// Test validation with existing variables
	requiredVars := []string{"title", "author"}
	err = engine.ValidateTemplate(requiredVars)
	if err != nil {
		t.Errorf("Unexpected error validating existing variables: %v", err)
	}

	// Test validation with missing variables
	requiredVars = []string{"title", "author", "missing"}
	err = engine.ValidateTemplate(requiredVars)
	if err == nil {
		t.Error("Expected error for missing variable, got nil")
	}

	if !strings.Contains(err.Error(), "required variable 'missing' not found") {
		t.Errorf("Expected error message about missing variable, got: %v", err)
	}
}

func TestEnhancedAddFunction(t *testing.T) {
	config := &EnhancedConfig{}
	engine := NewEnhancedEngine(config)

	// Add a custom function
	engine.AddFunction("custom", func(s string) string {
		return "custom_" + s
	})

	// Verify the function was added
	if _, exists := engine.Context.Functions["custom"]; !exists {
		t.Error("Expected custom function to be added")
	}
}

func TestClone(t *testing.T) {
	config := &EnhancedConfig{}
	engine := NewEnhancedEngine(config)

	// Set some variables
	variables := map[string]interface{}{
		"title": "Original Document",
	}

	err := engine.SetVariablesFromMap(variables)
	if err != nil {
		t.Fatalf("Failed to set variables: %v", err)
	}

	// Clone the engine
	clone := engine.Clone()

	// Verify the clone has the same variables
	originalVar, err := engine.GetVariable("title")
	if err != nil {
		t.Fatalf("Failed to get original variable: %v", err)
	}

	cloneVar, err := clone.GetVariable("title")
	if err != nil {
		t.Fatalf("Failed to get clone variable: %v", err)
	}

	if originalVar.Value != cloneVar.Value {
		t.Error("Expected clone to have the same variable value")
	}

	// Modify the clone and verify it doesn't affect the original
	clone.SetVariable("title", domain.NewStringVariable("Modified Document"))

	originalVar, err = engine.GetVariable("title")
	if err != nil {
		t.Fatalf("Failed to get original variable after clone modification: %v", err)
	}

	if originalVar.Value != "Original Document" {
		t.Error("Expected original engine to be unaffected by clone modification")
	}
}

func TestString(t *testing.T) {
	config := &EnhancedConfig{
		TemplatePath: "test.tex",
		OutputPath:   "output.tex",
		Engine:       "pdflatex",
	}
	engine := NewEnhancedEngine(config)

	str := engine.String()
	if !strings.Contains(str, "EnhancedEngine") {
		t.Error("Expected string representation to contain 'EnhancedEngine'")
	}
}
