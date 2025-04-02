package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rwxrob/bonzai/persisters/inyaml"
	"gopkg.in/yaml.v3"
)

// Config defines the YAML configuration schema for AutoPDF
type Config struct {
	Template   string            `yaml:"template" json:"template"`
	Output     string            `yaml:"output" json:"output"`
	Variables  map[string]string `yaml:"variables" json:"variables"`
	Engine     string            `yaml:"engine" json:"engine"`
	Conversion struct {
		Enabled bool     `yaml:"enabled" json:"enabled"`
		Formats []string `yaml:"formats" json:"formats"`
	} `yaml:"conversion" json:"conversion"`
}

// GetConfig retrieves the configuration from the persister
func GetConfig(persister *inyaml.Persister) (*Config, error) {
	configStr := persister.Get("autopdf_config")

	if configStr == "" {
		// Return default config
		return getDefaultConfig(), nil
	}

	var config Config
	if err := yaml.Unmarshal([]byte(configStr), &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves the configuration to the persister
func SaveConfig(persister *inyaml.Persister, config *Config) error {
	if config == nil {
		return errors.New("config cannot be nil")
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	persister.Set("autopdf_config", string(data))
	return nil
}

// getDefaultConfig returns a default configuration
func getDefaultConfig() *Config {
	return &Config{
		Engine: "pdflatex",
		Conversion: struct {
			Enabled bool     `yaml:"enabled" json:"enabled"`
			Formats []string `yaml:"formats" json:"formats"`
		}{
			Enabled: false,
			Formats: []string{},
		},
		Variables: make(map[string]string),
	}
}

// ToJSON converts a Config to JSON string representation
func (c *Config) ToJSON() (string, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// NewConfigFromYAML creates a new Config from YAML data
func NewConfigFromYAML(yamlData []byte) (*Config, error) {
	var config Config
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return nil, err
	}

	// Set defaults for required fields
	if config.Engine == "" {
		config.Engine = "pdflatex"
	}

	if config.Variables == nil {
		config.Variables = make(map[string]string)
	}

	return &config, nil
}

// String returns a string representation of the Config
func (c *Config) String() string {
	return fmt.Sprintf("Engine: %s, Output: %s", c.Engine, c.Output)
}
