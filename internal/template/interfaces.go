package template

import (
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/BuddhiLW/AutoPDF/pkg/domain"
)

// TemplateEngine defines the interface for template processing engines
type TemplateEngine interface {
	Process(templatePath string) (string, error)
	ProcessToFile(templatePath, outputPath string) error
	AddFunction(name string, fn interface{})
	ValidateTemplate(templatePath string) error
}

// EnhancedTemplateEngine defines the interface for enhanced template processing
type EnhancedTemplateEngine interface {
	Process(templatePath string) (string, error)
	ProcessToFile(templatePath, outputPath string) error
	SetVariable(key string, value interface{}) error
	SetVariablesFromMap(variables map[string]interface{}) error
	GetVariable(key string) (*domain.Variable, error)
	AddFunction(name string, fn interface{})
	ValidateTemplate(templatePath string) error
	Clone() EnhancedTemplateEngine
}

// ConfigProvider defines the interface for configuration providers
type ConfigProvider interface {
	GetConfig() *config.Config
	GetDefaultConfig() *config.Config
	LoadConfigFromFile(path string) (*config.Config, error)
	SaveConfigToFile(cfg *config.Config, path string) error
}

// VariableProcessor defines the interface for variable processing
type VariableProcessor interface {
	ProcessVariables(variables map[string]interface{}) (*domain.VariableCollection, error)
	GetVariable(key string) (*domain.Variable, error)
	SetVariable(key string, value interface{}) error
	GetNested(key string) (*domain.Variable, error)
}

// TemplateValidator defines the interface for template validation
type TemplateValidator interface {
	ValidateTemplate(templatePath string) error
	ValidateSyntax(templateContent string) error
	ValidateVariables(templateContent string, variables map[string]interface{}) error
}

// FileProcessor defines the interface for file operations
type FileProcessor interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, content []byte) error
	FileExists(path string) bool
	CreateDirectory(path string) error
	RemoveFile(path string) error
}
