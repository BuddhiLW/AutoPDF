package template

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/BuddhiLW/AutoPDF/pkg/domain"
)

// EnhancedEngine represents an enhanced template processing engine with support for complex data structures
type EnhancedEngine struct {
	Context *domain.TemplateContext
	Config  *EnhancedConfig
}

// EnhancedConfig represents the configuration for the enhanced template engine
type EnhancedConfig struct {
	TemplatePath string
	OutputPath   string
	Engine       string
	Delimiters   DelimiterConfig
	Functions    map[string]interface{}
}

// DelimiterConfig represents the delimiter configuration
type DelimiterConfig struct {
	Left  string
	Right string
}

// NewEnhancedEngine creates a new enhanced template engine
func NewEnhancedEngine(config *EnhancedConfig) *EnhancedEngine {
	context := domain.NewTemplateContext()

	// Add default functions
	context.AddFunction("len", func(arr []interface{}) int {
		return len(arr)
	})
	context.AddFunction("upper", func(s string) string {
		return strings.ToUpper(s)
	})
	context.AddFunction("lower", func(s string) string {
		return strings.ToLower(s)
	})
	context.AddFunction("title", func(s string) string {
		return strings.Title(s)
	})
	context.AddFunction("join", func(sep string, arr []interface{}) string {
		var strs []string
		for _, item := range arr {
			strs = append(strs, fmt.Sprintf("%v", item))
		}
		return strings.Join(strs, sep)
	})
	context.AddFunction("range", func(arr []interface{}) []interface{} {
		return arr
	})
	context.AddFunction("index", func(arr []interface{}, i int) interface{} {
		if i >= 0 && i < len(arr) {
			return arr[i]
		}
		return nil
	})
	context.AddFunction("keys", func(obj map[string]interface{}) []string {
		var keys []string
		for k := range obj {
			keys = append(keys, k)
		}
		return keys
	})
	context.AddFunction("values", func(obj map[string]interface{}) []interface{} {
		var values []interface{}
		for _, v := range obj {
			values = append(values, v)
		}
		return values
	})

	// Add custom functions from config
	for name, fn := range config.Functions {
		context.AddFunction(name, fn)
	}

	return &EnhancedEngine{
		Context: context,
		Config:  config,
	}
}

// SetVariable sets a variable in the template context
func (ee *EnhancedEngine) SetVariable(key string, value *domain.Variable) error {
	return ee.Context.SetVariable(key, value)
}

// SetVariablesFromMap sets multiple variables from a map
func (ee *EnhancedEngine) SetVariablesFromMap(variables map[string]interface{}) error {
	for key, value := range variables {
		variable := &domain.Variable{
			Type:  domain.DetermineType(value),
			Value: value,
		}
		if err := ee.SetVariable(key, variable); err != nil {
			return fmt.Errorf("failed to set variable '%s': %v", key, err)
		}
	}
	return nil
}

// Process processes a template with support for complex data structures
func (ee *EnhancedEngine) Process(templatePath string) (string, error) {
	if templatePath == "" && ee.Config.TemplatePath != "" {
		templatePath = ee.Config.TemplatePath
	}

	if templatePath == "" {
		return "", errors.New("no template file specified")
	}

	// Read template file
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %v", err)
	}

	// Create template with custom delimiters
	delims := ee.Config.Delimiters
	if delims.Left == "" {
		delims.Left = "delim[["
	}
	if delims.Right == "" {
		delims.Right = "]]"
	}

	tmpl, err := template.New(templatePath).
		Funcs(ee.Context.Functions).
		Delims(delims.Left, delims.Right).
		Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %v", err)
	}

	// Execute template with context data
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ee.Context.ToTemplateData()); err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}

	return buf.String(), nil
}

// ProcessToFile processes the template and writes the result to a file
func (ee *EnhancedEngine) ProcessToFile(templatePath, outputPath string) error {
	if outputPath == "" && ee.Config.OutputPath != "" {
		outputPath = ee.Config.OutputPath
	}

	if outputPath == "" {
		return errors.New("no output file specified")
	}

	result, err := ee.Process(templatePath)
	if err != nil {
		return err
	}

	// Write processed content to output file
	return os.WriteFile(outputPath, []byte(result), 0644)
}

// ValidateTemplate validates that all required variables are present
func (ee *EnhancedEngine) ValidateTemplate(requiredVars []string) error {
	return ee.Context.ValidateTemplate(requiredVars)
}

// GetVariable retrieves a variable from the context
func (ee *EnhancedEngine) GetVariable(key string) (*domain.Variable, error) {
	return ee.Context.GetVariable(key)
}

// AddFunction adds a custom function to the template engine
func (ee *EnhancedEngine) AddFunction(name string, fn interface{}) {
	ee.Context.AddFunction(name, fn)
}

// Clone creates a deep copy of the enhanced engine
func (ee *EnhancedEngine) Clone() *EnhancedEngine {
	return &EnhancedEngine{
		Context: ee.Context.Clone(),
		Config:  ee.Config, // Config is not cloned as it's typically immutable
	}
}

// String returns a string representation of the enhanced engine
func (ee *EnhancedEngine) String() string {
	return fmt.Sprintf("EnhancedEngine{Context: %s, Config: %+v}", ee.Context.String(), ee.Config)
}
