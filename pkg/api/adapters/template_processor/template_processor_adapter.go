// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package template_processor

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/BuddhiLW/AutoPDF/pkg/api"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// TemplateProcessorAdapter implements domain.TemplateProcessingService
type TemplateProcessorAdapter struct {
	config *config.Config
}

// NewTemplateProcessorAdapter creates a new template processor adapter
func NewTemplateProcessorAdapter(cfg *config.Config) *TemplateProcessorAdapter {
	return &TemplateProcessorAdapter{
		config: cfg,
	}
}

// Process processes a template with variables
func (tpa *TemplateProcessorAdapter) Process(ctx context.Context, templatePath string, variables map[string]interface{}) (string, error) {
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
		return "", domain.TemplateProcessingError{
			Code:    domain.ErrCodeTemplateNotFound,
			Message: api.ErrTemplateFileNotReadable,
			Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
				WithTemplatePath(templatePath).
				WithError(err),
		}
	}

	// Process template with variables
	processedContent, err := tpa.processTemplate(string(content), variables)
	if err != nil {
		return "", domain.TemplateProcessingError{
			Code:    domain.ErrCodeTemplateInvalid,
			Message: api.ErrTemplateProcessingFailed,
			Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
				WithTemplatePath(templatePath).
				WithError(err),
		}
	}

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

// processTemplate processes template content with variables
func (tpa *TemplateProcessorAdapter) processTemplate(content string, variables map[string]interface{}) (string, error) {
	// Use custom delimiters to avoid conflicts with LaTeX
	delimStart := "delim[["
	delimEnd := "]]"

	// Find all template variables
	pattern := regexp.MustCompile(regexp.QuoteMeta(delimStart) + `([^` + regexp.QuoteMeta(delimEnd) + `]+)` + regexp.QuoteMeta(delimEnd))

	result := pattern.ReplaceAllStringFunc(content, func(match string) string {
		// Extract variable name
		variableName := strings.TrimPrefix(match, delimStart)
		variableName = strings.TrimSuffix(variableName, delimEnd)
		variableName = strings.TrimSpace(variableName)

		// Handle dot notation and array access
		value := tpa.resolveVariableValue(variableName, variables)
		return value
	})

	return result, nil
}

// resolveVariableValue resolves a variable value from the variables map
func (tpa *TemplateProcessorAdapter) resolveVariableValue(variableName string, variables map[string]interface{}) string {
	// Handle dot notation (e.g., "foo.bar")
	parts := strings.Split(variableName, ".")

	var current interface{} = variables
	for _, part := range parts {
		// Handle array access (e.g., "foo[0]")
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			// Extract array name and index
			arrayName := strings.Split(part, "[")[0]
			indexStr := strings.TrimSuffix(strings.Split(part, "[")[1], "]")

			if currentMap, ok := current.(map[string]interface{}); ok {
				if array, exists := currentMap[arrayName]; exists {
					if arraySlice, ok := array.([]interface{}); ok {
						if index, err := strconv.Atoi(indexStr); err == nil && index >= 0 && index < len(arraySlice) {
							current = arraySlice[index]
							continue
						}
					}
				}
			}
			return "" // Variable not found
		}

		// Handle regular map access
		if currentMap, ok := current.(map[string]interface{}); ok {
			if val, exists := currentMap[part]; exists {
				current = val
			} else {
				return "" // Variable not found
			}
		} else {
			return "" // Cannot traverse further
		}
	}

	// Convert final value to string
	return fmt.Sprintf("%v", current)
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
