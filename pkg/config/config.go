package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

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
	FormatFile FormatFile `yaml:"format_file" json:"format_file" default:""` // Optional precompiled format file path
	WorkingDir WorkingDir `yaml:"working_dir" json:"working_dir" default:""` // Working directory for LaTeX execution
}

func (c *Config) String() string {
	data, err := yaml.Marshal(c)
	if err != nil {
		return ""
	}
	return string(data)
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

// FormatFile represents a precompiled LaTeX format file path
// Value Object: Encapsulates format file path with domain meaning
type FormatFile string

func (f FormatFile) String() string {
	return string(f)
}

// IsEmpty checks if format file path is empty
func (f FormatFile) IsEmpty() bool {
	return strings.TrimSpace(string(f)) == ""
}

// WorkingDir represents the working directory for LaTeX compilation
// Value Object: Encapsulates working directory path with domain meaning
type WorkingDir string

func (w WorkingDir) String() string {
	return string(w)
}

// IsEmpty checks if working directory path is empty
func (w WorkingDir) IsEmpty() bool {
	return strings.TrimSpace(string(w)) == ""
}

// Variables represents a collection of complex variables
type Variables struct {
	*VariableSet
}

// NewVariables creates a new Variables collection
func NewVariables() *Variables {
	return &Variables{
		VariableSet: NewVariableSet(),
	}
}

// String returns the string representation
func (v Variables) String() string {
	if v.VariableSet == nil || v.Len() == 0 {
		return "{}"
	}

	flattened := v.Flatten()
	s := "{"
	first := true
	for k, val := range flattened {
		if !first {
			s += ", "
		}
		s += fmt.Sprintf("%s: %s", k, val)
		first = false
	}
	s += "}"
	return s
}

// GetString gets a string value by path (for backward compatibility)
func (v Variables) GetString(path string) (string, bool) {
	if v.VariableSet == nil {
		return "", false
	}
	return v.VariableSet.GetString(path)
}

// SetString sets a string value by path (for backward compatibility)
func (v *Variables) SetString(path string, value string) error {
	if v.VariableSet == nil {
		v.VariableSet = NewVariableSet()
	}
	return v.VariableSet.SetByPath(path, &StringVariable{Value: value})
}

// MarshalJSON implements json.Marshaler
func (v Variables) MarshalJSON() ([]byte, error) {
	if v.VariableSet == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return v.VariableSet.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler
func (v *Variables) UnmarshalJSON(data []byte) error {
	if v.VariableSet == nil {
		v.VariableSet = NewVariableSet()
	}
	return v.VariableSet.UnmarshalJSON(data)
}

// MarshalYAML implements yaml.Marshaler
func (v Variables) MarshalYAML() (interface{}, error) {
	if v.VariableSet == nil {
		return map[string]interface{}{}, nil
	}
	return v.VariableSet.variables, nil
}

// UnmarshalYAML implements yaml.Unmarshaler
func (v *Variables) UnmarshalYAML(value *yaml.Node) error {
	if v.VariableSet == nil {
		v.VariableSet = NewVariableSet()
	}

	// Handle different YAML node types
	switch value.Kind {
	case yaml.MappingNode:
		// Parse as map
		return v.unmarshalYAMLMap(value)
	case yaml.SequenceNode:
		// Parse as array
		return v.unmarshalYAMLSequence(value)
	case yaml.ScalarNode:
		// Parse as scalar
		return v.unmarshalYAMLScalar(value)
	default:
		return fmt.Errorf("unsupported YAML node type: %v", value.Kind)
	}
}

// unmarshalYAMLMap unmarshals a YAML mapping node
func (v *Variables) unmarshalYAMLMap(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping node, got %v", node.Kind)
	}

	for i := 0; i < len(node.Content); i += 2 {
		if i+1 >= len(node.Content) {
			return fmt.Errorf("incomplete key-value pair in mapping")
		}

		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		if keyNode.Kind != yaml.ScalarNode {
			return fmt.Errorf("mapping key must be scalar, got %v", keyNode.Kind)
		}

		key := keyNode.Value
		value := v.convertYAMLNodeToVariable(valueNode)

		v.VariableSet.Set(key, value)
	}

	return nil
}

// unmarshalYAMLSequence unmarshals a YAML sequence node
func (v *Variables) unmarshalYAMLSequence(node *yaml.Node) error {
	if node.Kind != yaml.SequenceNode {
		return fmt.Errorf("expected sequence node, got %v", node.Kind)
	}

	// For root level, we don't support arrays directly
	// This would be used if Variables was an array, but it's a map
	return fmt.Errorf("variables must be a mapping, not a sequence")
}

// unmarshalYAMLScalar unmarshals a YAML scalar node
func (v *Variables) unmarshalYAMLScalar(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode {
		return fmt.Errorf("expected scalar node, got %v", node.Kind)
	}

	// For root level, we don't support scalars directly
	// This would be used if Variables was a scalar, but it's a map
	return fmt.Errorf("variables must be a mapping, not a scalar")
}

// convertYAMLNodeToVariable converts a YAML node to a Variable
func (v *Variables) convertYAMLNodeToVariable(node *yaml.Node) Variable {
	switch node.Kind {
	case yaml.ScalarNode:
		// Try to parse as different types
		if node.Tag == "!!bool" {
			if node.Value == "true" {
				return &BoolVariable{Value: true}
			}
			return &BoolVariable{Value: false}
		}
		if node.Tag == "!!int" {
			if intVal, err := strconv.Atoi(node.Value); err == nil {
				return &NumberVariable{Value: float64(intVal)}
			}
		}
		if node.Tag == "!!float" {
			if floatVal, err := strconv.ParseFloat(node.Value, 64); err == nil {
				return &NumberVariable{Value: floatVal}
			}
		}
		// Default to string
		return &StringVariable{Value: node.Value}

	case yaml.MappingNode:
		// Convert to MapVariable
		mapVar := NewMapVariable()
		for i := 0; i < len(node.Content); i += 2 {
			if i+1 >= len(node.Content) {
				break
			}
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			if keyNode.Kind == yaml.ScalarNode {
				key := keyNode.Value
				value := v.convertYAMLNodeToVariable(valueNode)
				mapVar.Set(key, value)
			}
		}
		return mapVar

	case yaml.SequenceNode:
		// Convert to SliceVariable
		sliceVar := NewSliceVariable()
		for _, itemNode := range node.Content {
			item := v.convertYAMLNodeToVariable(itemNode)
			sliceVar.Values = append(sliceVar.Values, item)
		}
		return sliceVar

	default:
		// Fallback to string
		return &StringVariable{Value: node.Value}
	}
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
		return GetDefaultConfig(), nil
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

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *Config {
	return &Config{
		Template: "",
		Output:   "",
		Engine:   "pdflatex",
		Conversion: Conversion{
			Enabled: false,
			Formats: []string{},
		},
		Variables: *NewVariables(),
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

	if config.Variables.VariableSet == nil {
		config.Variables = *NewVariables()
	}

	return &config, nil
}

func (c *Config) Marshal() ([]byte, error) {
	return yaml.Marshal(c)
}

func (c *Config) Unmarshal(data []byte) error {
	return yaml.Unmarshal(data, c)
}
