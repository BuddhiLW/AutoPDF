// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package template_processor

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

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

// processTemplate processes template content with variables
func (tpa *TemplateProcessorAdapter) processTemplate(content string, variables map[string]string) (string, error) {
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

		// Simple variable lookup
		if value, exists := variables[variableName]; exists {
			return value
		}
		return "" // Variable not found
	})

	return result, nil
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
