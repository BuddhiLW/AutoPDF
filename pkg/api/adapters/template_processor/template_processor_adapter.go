// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package template_processor

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/pkg/api"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// TemplateProcessorAdapter implements domain.TemplateProcessingService
type TemplateProcessorAdapter struct {
	config *config.Config
	logger *logger.LoggerAdapter
}

// NewTemplateProcessorAdapter creates a new template processor adapter
func NewTemplateProcessorAdapter(cfg *config.Config, logger *logger.LoggerAdapter) *TemplateProcessorAdapter {
	return &TemplateProcessorAdapter{
		config: cfg,
		logger: logger,
	}
}

// Process processes a template with variables
func (tpa *TemplateProcessorAdapter) Process(ctx context.Context, templatePath string, variables map[string]string) (string, error) {
	tpa.logger.DebugWithFields("Starting template processing",
		"template_path", templatePath,
		"variable_count", len(variables),
		"variable_keys", getStringMapKeys(variables),
	)

	if templatePath == "" {
		return "", domain.TemplateProcessingError{
			Code:    domain.ErrCodeTemplateNotFound,
			Message: api.ErrTemplatePathRequired,
			Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh),
		}
	}

	// Read template file
	content, err := os.ReadFile(templatePath)
	if err != nil {
		tpa.logger.ErrorWithFields("Failed to read template file",
			"template_path", templatePath,
			"error", err,
		)
		return "", domain.TemplateProcessingError{
			Code:    domain.ErrCodeTemplateNotFound,
			Message: api.ErrTemplateFileNotReadable,
			Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
				WithTemplatePath(templatePath).
				WithError(err),
		}
	}

	tpa.logger.DebugWithFields("Template file read successfully",
		"template_path", templatePath,
		"content_length", len(content),
	)

	// Process template with variables
	processedContent, err := tpa.processTemplate(string(content), variables)
	if err != nil {
		tpa.logger.ErrorWithFields("Failed to process template",
			"template_path", templatePath,
			"error", err,
		)
		return "", domain.TemplateProcessingError{
			Code:    domain.ErrCodeTemplateInvalid,
			Message: api.ErrTemplateProcessingFailed,
			Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
				WithTemplatePath(templatePath).
				WithError(err),
		}
	}

	tpa.logger.InfoWithFields("Template processing completed",
		"template_path", templatePath,
		"original_length", len(content),
		"processed_length", len(processedContent),
	)

	return processedContent, nil
}

// ValidateTemplate validates a template file
func (tpa *TemplateProcessorAdapter) ValidateTemplate(templatePath string) error {
	if templatePath == "" {
		return domain.TemplateProcessingError{
			Code:    domain.ErrCodeTemplateNotFound,
			Message: api.ErrTemplatePathRequired,
			Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh),
		}
	}

	// Check if file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return domain.TemplateProcessingError{
			Code:    domain.ErrCodeTemplateNotFound,
			Message: api.ErrTemplateFileNotFound,
			Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
				WithTemplatePath(templatePath),
		}
	}

	// Read and validate template content
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return domain.TemplateProcessingError{
			Code:    domain.ErrCodeTemplateInvalid,
			Message: api.ErrTemplateFileNotReadable,
			Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
				WithTemplatePath(templatePath).
				WithError(err),
		}
	}

	// Basic LaTeX validation
	if err := tpa.validateLaTeXContent(string(content)); err != nil {
		return domain.TemplateProcessingError{
			Code:    domain.ErrCodeTemplateInvalid,
			Message: api.ErrTemplateSyntaxInvalid,
			Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
				WithTemplatePath(templatePath).
				WithError(err),
		}
	}

	return nil
}

// GetTemplateVariables extracts variables from a template
func (tpa *TemplateProcessorAdapter) GetTemplateVariables(templatePath string) ([]string, error) {
	if templatePath == "" {
		return nil, domain.TemplateProcessingError{
			Code:    domain.ErrCodeTemplateNotFound,
			Message: api.ErrTemplatePathRequired,
			Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh),
		}
	}

	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, domain.TemplateProcessingError{
			Code:    domain.ErrCodeTemplateNotFound,
			Message: api.ErrTemplateFileNotReadable,
			Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
				WithTemplatePath(templatePath).
				WithError(err),
		}
	}

	variables := tpa.extractVariables(string(content))
	return variables, nil
}

// processTemplate processes template content with variables using Go's text/template
func (tpa *TemplateProcessorAdapter) processTemplate(content string, variables map[string]string) (string, error) {
	// Import text/template at the top of the file
	// We need to use Go's text/template to support conditionals, loops, and dot notation

	// Convert flattened variables back to nested structure for template execution
	templateData := tpa.reconstructNestedStructure(variables)

	// Create template with custom delimiters
	tmpl, err := tpa.createTemplate(content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var buf strings.Builder
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// createTemplate creates a Go template with custom delimiters
func (tpa *TemplateProcessorAdapter) createTemplate(content string) (*template.Template, error) {
	// Create function map for template functions
	funcMap := template.FuncMap{
		"eq": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
		},
	}

	// Create template with custom delimiters to avoid conflicts with LaTeX
	tmpl, err := template.New("latex").
		Funcs(funcMap).
		Delims("delim[[", "]]").
		Parse(content)

	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

// reconstructNestedStructure converts flattened variables back to nested structure
// Handles flattened dot-notation keys from Flatten() method
func (tpa *TemplateProcessorAdapter) reconstructNestedStructure(variables map[string]string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range variables {
		// Handle flattened dot-notation keys (e.g., "velorio.dia", "velorio.periodo.inicio")
		parts := splitKey(key)
		tpa.setNestedValue(result, parts, value)
	}

	return result
}

// splitKey splits a dot-notation key into parts
func splitKey(key string) []string {
	return strings.Split(key, ".")
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

// getStringMapKeys returns the keys of a string map for debugging
func getStringMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// validateLaTeXContent performs basic LaTeX validation
func (tpa *TemplateProcessorAdapter) validateLaTeXContent(content string) error {
	// Check for basic LaTeX structure
	if !strings.Contains(content, "\\documentclass") {
		return fmt.Errorf("template does not contain \\documentclass")
	}

	if !strings.Contains(content, "\\begin{document}") {
		return fmt.Errorf("template does not contain \\begin{document}")
	}

	if !strings.Contains(content, "\\end{document}") {
		return fmt.Errorf("template does not contain \\end{document}")
	}

	// Check for balanced braces (basic check)
	openBraces := strings.Count(content, "{")
	closeBraces := strings.Count(content, "}")
	if openBraces != closeBraces {
		return fmt.Errorf("unbalanced braces in template")
	}

	return nil
}

// extractVariables extracts variable names from template content
func (tpa *TemplateProcessorAdapter) extractVariables(content string) []string {
	delimStart := "delim[["
	delimEnd := "]]"

	pattern := regexp.MustCompile(regexp.QuoteMeta(delimStart) + `([^` + regexp.QuoteMeta(delimEnd) + `]+)` + regexp.QuoteMeta(delimEnd))
	matches := pattern.FindAllStringSubmatch(content, -1)

	variables := make([]string, 0, len(matches))
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			variableName := strings.TrimSpace(match[1])
			if !seen[variableName] {
				variables = append(variables, variableName)
				seen[variableName] = true
			}
		}
	}

	return variables
}
