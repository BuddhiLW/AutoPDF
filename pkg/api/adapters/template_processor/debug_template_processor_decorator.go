package template_processor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
)

// DebugTemplateProcessorDecorator adds debug capabilities to template processing
type DebugTemplateProcessorDecorator struct {
	wrapped         generation.TemplateProcessingService
	logger          *logger.LoggerAdapter
	concreteFileDir string
	requestID       string
}

// NewDebugTemplateProcessorDecorator creates a new debug decorator
func NewDebugTemplateProcessorDecorator(
	wrapped generation.TemplateProcessingService,
	logger *logger.LoggerAdapter,
	concreteFileDir string,
	requestID string,
) *DebugTemplateProcessorDecorator {
	return &DebugTemplateProcessorDecorator{
		wrapped:         wrapped,
		logger:          logger,
		concreteFileDir: concreteFileDir,
		requestID:       requestID,
	}
}

// Process processes a template with variables and creates concrete files in debug mode
func (d *DebugTemplateProcessorDecorator) Process(
	ctx context.Context,
	templatePath string,
	variables map[string]string,
) (string, error) {
	// Delegate to wrapped processor
	content, err := d.wrapped.Process(ctx, templatePath, variables)
	if err != nil {
		return "", err
	}

	// Debug behavior: create persistent concrete file
	concreteFile := d.createConcreteFile(templatePath, content)
	d.logger.InfoWithFields("Created concrete template file",
		"path", concreteFile,
		"request_id", d.requestID,
		"template_path", templatePath,
		"content_length", len(content),
	)

	return content, nil
}

// ValidateTemplate delegates to wrapped service
func (d *DebugTemplateProcessorDecorator) ValidateTemplate(templatePath string) error {
	return d.wrapped.ValidateTemplate(templatePath)
}

// GetTemplateVariables delegates to wrapped service
func (d *DebugTemplateProcessorDecorator) GetTemplateVariables(templatePath string) ([]string, error) {
	return d.wrapped.GetTemplateVariables(templatePath)
}

// createConcreteFile creates a persistent concrete file with variable substitution
func (d *DebugTemplateProcessorDecorator) createConcreteFile(
	templatePath, content string,
) string {
	// Ensure directory exists
	if err := os.MkdirAll(d.concreteFileDir, 0755); err != nil {
		d.logger.WarnWithFields("Failed to create concrete file directory",
			"directory", d.concreteFileDir,
			"error", err,
		)
		// Fallback to current directory
		d.concreteFileDir = "."
	}

	// Generate concrete file name
	baseName := strings.TrimSuffix(filepath.Base(templatePath), ".tex")
	concreteFileName := fmt.Sprintf("autopdf-concrete-%s-%s.tex", baseName, d.requestID)
	concreteFile := filepath.Join(d.concreteFileDir, concreteFileName)

	// Write content to concrete file
	if err := os.WriteFile(concreteFile, []byte(content), 0644); err != nil {
		d.logger.ErrorWithFields("Failed to write concrete file",
			"file", concreteFile,
			"error", err,
		)
		// Return the intended path even if write failed
		return concreteFile
	}

	return concreteFile
}
