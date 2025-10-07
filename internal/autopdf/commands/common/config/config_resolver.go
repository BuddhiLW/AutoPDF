// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/BuddhiLW/AutoPDF/configs"
	// Legacy tex functionality now integrated into adapters
	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// ConfigResolver handles config file resolution and template path resolution
type ConfigResolver struct{}

// NewConfigResolver creates a new config resolver
func NewConfigResolver() *ConfigResolver {
	return &ConfigResolver{}
}

// ResolveConfigFile determines the config file to use and creates default if needed
func (cr *ConfigResolver) ResolveConfigFile(templateFile string, providedConfigFile string) (string, error) {
	if providedConfigFile != "" {
		return providedConfigFile, nil
	}

	// Create default config if not provided
	log.Println("No config file provided, creating default config file...")
	err := cr.createDefaultConfig(templateFile)
	if err != nil {
		return "", configs.BuildError
	}
	log.Println("Default config file written to:", configs.DefaultConfigName)
	return configs.DefaultConfigName, nil
}

// ResolveTemplatePath resolves the template path, handling both config and command-line scenarios
func (cr *ConfigResolver) ResolveTemplatePath(cfg *config.Config, templateFile, configFile string) error {
	if cfg.Template == "" {
		// No template in config, use the provided one
		absTemplatePath, err := filepath.Abs(templateFile)
		if err != nil {
			return fmt.Errorf("failed to resolve template path: %w", err)
		}
		cfg.Template = config.Template(absTemplatePath)
	} else {
		// Template is set in config, but it might be relative
		// Resolve it relative to the config file's directory
		configDir := filepath.Dir(configFile)
		templatePath := cfg.Template.String()

		// If it's not already absolute, make it relative to config directory
		if !filepath.IsAbs(templatePath) {
			absTemplatePath := filepath.Join(configDir, templatePath)
			absTemplatePath, err := filepath.Abs(absTemplatePath)
			if err != nil {
				return fmt.Errorf("failed to resolve template path: %w", err)
			}
			cfg.Template = config.Template(absTemplatePath)
		}
	}
	return nil
}

// createDefaultConfig creates a default configuration file
func (cr *ConfigResolver) createDefaultConfig(templateFile string) error {
	// Create a basic default config
	defaultConfig := `template: "` + templateFile + `"
variables:
  title: "My Document"
  author: "AutoPDF User"
  date: "` + time.Now().Format("2006-01-02") + `"
engine: "pdflatex"
output: ""
conversion:
  enabled: false
  formats: ["png"]
`

	// Write the default config to file
	return os.WriteFile(configs.DefaultConfigName, []byte(defaultConfig), 0644)
}

// LoadConfig loads and parses the config file
func (cr *ConfigResolver) LoadConfig(configFile string) (*config.Config, error) {
	yamlData, err := os.ReadFile(configFile)
	if err != nil {
		return nil, configs.ReadError
	}

	cfg, err := config.NewConfigFromYAML(yamlData)
	if err != nil {
		return nil, configs.ParseError
	}

	return cfg, nil
}
