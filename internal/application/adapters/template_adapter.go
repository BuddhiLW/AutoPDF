// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"context"
	"os"
	"path/filepath"

	"github.com/BuddhiLW/AutoPDF/internal/template"
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
	// Create a config with the provided variables, but preserve the template from the original config
	cfg := &config.Config{
		Template:  tpa.config.Template, // Use the template from the original config
		Variables: config.Variables(variables),
		Engine:    tpa.config.Engine,
		Output:    tpa.config.Output,
	}

	// If no template is set in the config, use the provided templatePath
	if cfg.Template == "" {
		cfg.Template = config.Template(templatePath)
	}

	// Create the template engine
	engine := template.NewEngine(cfg)

	// Process the template using the template path from config
	configTemplatePath := cfg.Template.String()
	// log.Printf("Template adapter processing template: %s", configTemplatePath)
	result, err := engine.Process(configTemplatePath)
	if err != nil {
		// log.Printf("Template processing failed: %v", err)
		return "", err
	}

	return result, nil
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
