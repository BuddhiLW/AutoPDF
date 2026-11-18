// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/BuddhiLW/AutoPDF/v2/configs"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/backend"
	"gopkg.in/yaml.v3"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/commands/common/args"
	// Legacy tex functionality now integrated into adapters
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
)

// ConfigResolver handles config file resolution and template path resolution
type ConfigResolver struct {
	stdin io.Reader
}

// NewConfigResolver creates a new config resolver that reads a "-" config
// from os.Stdin.
func NewConfigResolver() *ConfigResolver {
	return &ConfigResolver{stdin: os.Stdin}
}

// WithStdin returns a copy of the resolver that reads a "-" config from r.
func (cr *ConfigResolver) WithStdin(r io.Reader) *ConfigResolver {
	return &ConfigResolver{stdin: r}
}

// IsStdinConfig reports whether configFile names standard input.
func IsStdinConfig(configFile string) bool {
	return configFile == args.StdinConfig
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

// ResolveTemplatePath resolves the template path, handling both config and command-line scenarios.
// A relative template in the config resolves against the config file's
// directory, or against templateFile's directory when the config came from
// standard input ("-").
func (cr *ConfigResolver) ResolveTemplatePath(cfg *config.Config, templateFile, configFile string) error {
	if IsStdinConfig(configFile) {
		configFile = templateFile
	}
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
	// The scaffolded variable set covers every key the shipped sample template
	// reads, which is the set configs/sample-config.yaml carries. Under strict
	// rendering an omitted key refuses the build rather than printing a blank.
	engine := backend.Default().DefaultEngine()
	if selected, err := backend.Default().Resolve("", templateFile); err == nil {
		engine = selected.DefaultEngine()
	}
	defaultConfig := `template: "` + templateFile + `"
variables:
  title: "My Document"
  author: "AutoPDF User"
  date: "` + time.Now().Format("2006-01-02") + `"
  content: "Replace this with your document body."
engine: "` + engine + `"
output: ""
conversion:
  enabled: false
  formats: ["png"]
`

	// Write the default config to file
	return os.WriteFile(configs.DefaultConfigName, []byte(defaultConfig), 0644)
}

// LoadConfig loads and parses the config file, or standard input when
// configFile is "-". The content may be JSON or YAML.
func (cr *ConfigResolver) LoadConfig(configFile string) (*config.Config, error) {
	data, err := cr.readConfig(configFile)
	if err != nil {
		return nil, configs.ReadError
	}
	return ParseConfigData(data)
}

func (cr *ConfigResolver) readConfig(configFile string) ([]byte, error) {
	if !IsStdinConfig(configFile) {
		return os.ReadFile(configFile)
	}
	if cr.stdin == nil {
		return nil, fmt.Errorf("no standard input to read the config from")
	}
	return io.ReadAll(cr.stdin)
}

// ParseConfigData parses a config document. Content that is a JSON object is
// decoded as JSON; anything else is decoded as YAML.
func ParseConfigData(data []byte) (*config.Config, error) {
	yamlData, err := normalizeToYAML(data)
	if err != nil {
		return nil, configs.ParseError
	}

	cfg, err := config.NewConfigFromYAML(yamlData)
	if err != nil {
		return nil, configs.ParseError
	}
	if !engineDeclared(yamlData) {
		// An omitted engine is chosen by the template's backend once the
		// template path is known (ResolveBackend), not defaulted to LaTeX here.
		cfg.Engine = ""
	}

	return cfg, nil
}

func normalizeToYAML(data []byte) ([]byte, error) {
	trimmed := bytes.TrimSpace(bytes.TrimPrefix(data, []byte("\xef\xbb\xbf")))
	if !looksLikeJSON(trimmed) {
		return data, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.UseNumber()
	var document map[string]interface{}
	if err := decoder.Decode(&document); err != nil {
		return nil, err
	}
	return yaml.Marshal(jsonNumbersToYAML(document))
}

func jsonNumbersToYAML(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, nested := range typed {
			typed[key] = jsonNumbersToYAML(nested)
		}
		return typed
	case []interface{}:
		for index, nested := range typed {
			typed[index] = jsonNumbersToYAML(nested)
		}
		return typed
	case json.Number:
		tag := "!!float"
		if _, err := typed.Int64(); err == nil {
			tag = "!!int"
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: typed.String()}
	default:
		return value
	}
}

func looksLikeJSON(data []byte) bool {
	return len(data) > 0 && data[0] == '{' && json.Valid(data)
}

// engineDeclared reports whether the YAML sets a non-empty engine.
func engineDeclared(yamlData []byte) bool {
	var probe struct {
		Engine string `yaml:"engine"`
	}
	if err := yaml.Unmarshal(yamlData, &probe); err != nil {
		return true
	}
	return probe.Engine != ""
}

// ResolveBackend picks the backend for cfg's engine and template, fills an
// omitted engine with that backend's default engine, and refuses settings the
// backend cannot honour (e.g. passes > 1 or use_latexmk with typst).
func (cr *ConfigResolver) ResolveBackend(cfg *config.Config) error {
	return ResolveBackendWith(backend.Default(), cfg)
}

// ResolveBackendWith is ResolveBackend over an explicit registry.
func ResolveBackendWith(registry *backend.Registry, cfg *config.Config) error {
	if registry == nil {
		registry = backend.Default()
	}
	selected, err := registry.Resolve(cfg.Engine.String(), cfg.Template.String())
	if err != nil {
		return err
	}
	if cfg.Engine == "" {
		cfg.Engine = config.Engine(selected.DefaultEngine())
	}
	if _, err := selected.Options(backend.Settings{Passes: cfg.Passes, UseLatexmk: cfg.UseLatexmk, FormatFile: cfg.FormatFile}); err != nil {
		return fmt.Errorf("invalid configuration for the %s backend: %w", selected.Name, err)
	}
	return nil
}

// LoadConfigWithLogging loads configuration with integrated logging
func (cr *ConfigResolver) LoadConfigWithLogging(ctx context.Context, templateFile, providedConfigFile string) (*config.Config, error) {
	logger := configs.GetLoggerFromContext(ctx)

	// Log start
	logger.InfoWithFields("Starting AutoPDF build process", "template_file", templateFile, "config_file", providedConfigFile)

	// Resolve config file with logging
	logger.Debug("Resolving configuration file")
	configFile, err := cr.ResolveConfigFile(templateFile, providedConfigFile)
	if err != nil {
		logger.ErrorWithFields("Failed to resolve config file", "error", err)
		return nil, err
	}
	logger.InfoWithFields("Configuration file resolved", "config_file", configFile)

	// Load config with logging
	logger.Debug("Loading configuration")
	cfg, err := cr.LoadConfig(configFile)
	if err != nil {
		logger.ErrorWithFields("Failed to load configuration", "error", err)
		return nil, err
	}
	logger.LogConfigBuilding(configFile, cfg.Variables.Flatten())

	// Resolve template path with logging
	logger.Debug("Resolving template path")
	err = cr.ResolveTemplatePath(cfg, templateFile, configFile)
	if err != nil {
		logger.ErrorWithFields("Failed to resolve template path", "error", err)
		return nil, err
	}
	logger.LogDataMapping(cfg.Template.String(), cfg.Variables.Flatten())

	if err := cr.ResolveBackend(cfg); err != nil {
		logger.ErrorWithFields("Configuration refused by its backend", "error", err)
		return nil, err
	}

	return cfg, nil
}
