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
	Template   Template   `yaml:"template" json:"template" default:""`
	Output     Output     `yaml:"output" json:"output" default:""`
	Variables  Variables  `yaml:"variables" json:"variables" default:"{}"`
	Engine     Engine     `yaml:"engine" json:"engine" default:"pdflatex"`
	Conversion Conversion `yaml:"conversion" json:"conversion"`
}

type Template string

func (t Template) String() string {
	return string(t)
}

type Output string

func (o Output) String() string {
	return string(o)
}

type Engine string

func (e Engine) String() string {
	return string(e)
}

type Variables map[string]string

func (v Variables) String() string {
	if len(v) == 0 {
		return "{}"
	}
	s := "{"
	for k, v := range v {
		s += fmt.Sprintf("%s: %s, ", k, v)
	}
	s += "}"
	return s
}

type Conversion struct {
	Enabled bool     `yaml:"enabled" json:"enabled" default:"false"`
	Formats []string `yaml:"formats" json:"formats" default:"[]"`
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
		Template: "",
		Output:   "",
		Engine:   "pdflatex",
		Conversion: Conversion{
			Enabled: false,
			Formats: []string{},
		},
		Variables: Variables(make(map[string]string)),
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
