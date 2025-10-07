// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
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

	// Create template data that includes both flattened variables and complex structure
	templateData := map[string]interface{}{
		// Flattened variables for direct access
		"vars": variables,
		// Complex structure for range loops (if available)
		"complex": tpa.getComplexVariables(),
	}

	// Apply template with both flattened and complex variables
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", err
	}

	return buf.String(), nil
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
