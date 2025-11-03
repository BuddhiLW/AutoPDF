// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// TemplateProcessorAdapter wraps the existing template engine
type TemplateProcessorAdapter struct {
	// config is used to initialize the template engine
	config *config.Config
}

// NewTemplateProcessorAdapter creates a new template processor adapter
func NewTemplateProcessorAdapter(cfg *config.Config) *TemplateProcessorAdapter {
	return &TemplateProcessorAdapter{
		config: cfg,
	}
}

// Process processes a template with variables
func (tpa *TemplateProcessorAdapter) Process(ctx context.Context, templatePath string, variables map[string]string) (string, error) {
	if templatePath == "" {
		return "", errors.New("no template file specified")
	}

	// Read template file
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", err
	}

	// Create function map for template functions
	funcMap := template.FuncMap{
		"upper": func(s string) string {
			return s
		},
		// Add more helper functions as needed
	}

	// Create new template with custom delimiters to avoid conflicts with LaTeX
	tmpl, err := template.New(templatePath).
		Funcs(funcMap).
		Delims("delim[[", "]]").
		Parse(string(content))
	if err != nil {
		return "", err
	}

	// Reconstruct nested structure from flattened variables for direct template access
	// Variables are placed at root level so templates can use delim[[.logo_path]]
	// instead of delim[[.vars.logo_path]]
	nestedVariables := tpa.reconstructNestedStructure(variables)

	// Create template data with variables at root level
	templateData := nestedVariables
	if templateData == nil {
		templateData = make(map[string]interface{})
	}

	// Add complex structure for range loops (if available)
	complex := tpa.getComplexVariables()
	if len(complex) > 0 {
		for k, v := range complex {
			templateData[k] = v
		}
	}

	// Apply template with both flattened and complex variables
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// reconstructNestedStructure converts flattened variables back to nested structure
// Handles flattened dot-notation keys from Flatten() method
// Variables are placed at root level for direct template access (e.g., .logo_path)
func (tpa *TemplateProcessorAdapter) reconstructNestedStructure(variables map[string]string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range variables {
		// Handle flattened dot-notation keys (e.g., "velorio.dia", "velorio.periodo.inicio")
		parts := strings.Split(key, ".")
		tpa.setNestedValue(result, parts, value)
	}

	// Return variables at root level (NOT wrapped under "vars" key)
	// This allows templates to use direct access like delim[[.logo_path]]
	// instead of requiring delim[[.vars.logo_path]]
	return result
}

// setNestedValue sets a value in a nested map structure
func (tpa *TemplateProcessorAdapter) setNestedValue(data map[string]interface{}, parts []string, value string) {
	if len(parts) == 0 {
		return
	}

	if len(parts) == 1 {
		data[parts[0]] = value
		return
	}

	// Create nested map if it doesn't exist
	if _, exists := data[parts[0]]; !exists {
		data[parts[0]] = make(map[string]interface{})
	}

	// Ensure it's a map
	if nested, ok := data[parts[0]].(map[string]interface{}); ok {
		tpa.setNestedValue(nested, parts[1:], value)
	}
}

// getComplexVariables returns the complex variable structure for range loops
func (tpa *TemplateProcessorAdapter) getComplexVariables() map[string]interface{} {
	if tpa.config == nil || tpa.config.Variables.VariableSet == nil {
		return make(map[string]interface{})
	}

	// Convert the complex variables to a map that can be used in range loops
	result := make(map[string]interface{})
	for k, v := range tpa.config.Variables.VariableSet.GetVariables() {
		result[k] = tpa.convertToMap(v)
	}
	return result
}

// convertToMap converts a Variable to a map for template range loops
func (tpa *TemplateProcessorAdapter) convertToMap(variable config.Variable) interface{} {
	switch v := variable.(type) {
	case *config.StringVariable:
		return v.Value
	case *config.NumberVariable:
		return v.Value
	case *config.BoolVariable:
		return v.Value
	case *config.MapVariable:
		result := make(map[string]interface{})
		for k, val := range v.Values {
			result[k] = tpa.convertToMap(val)
		}
		return result
	case *config.SliceVariable:
		result := make([]interface{}, len(v.Values))
		for i, val := range v.Values {
			result[i] = tpa.convertToMap(val)
		}
		return result
	default:
		return variable.String()
	}
}

// ProcessToFile processes a template and writes it to a file
func (tpa *TemplateProcessorAdapter) ProcessToFile(ctx context.Context, templatePath string, variables map[string]string, outputPath string) error {
	// Process the template
	result, err := tpa.Process(ctx, templatePath, variables)
	if err != nil {
		return err
	}

	// Ensure the output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	// Write the processed template to the output file
	if err := os.WriteFile(outputPath, []byte(result), 0644); err != nil {
		return err
	}

	return nil
}
