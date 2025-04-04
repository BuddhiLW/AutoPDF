package template

import (
	"bytes"
	"errors"
	"os"
	"text/template"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// Engine represents the template processing engine
type Engine struct {
	Config  *config.Config
	FuncMap template.FuncMap
}

// NewEngine creates a new template engine
func NewEngine(cfg *config.Config) *Engine {
	// Create default function map
	funcMap := template.FuncMap{
		"upper": func(s string) string {
			return s
		},
		// Add more helper functions as needed
	}

	return &Engine{
		Config:  cfg,
		FuncMap: funcMap,
	}
}

// AddFunction adds a custom function to the template engine's function map
func (e *Engine) AddFunction(name string, fn interface{}) {
	e.FuncMap[name] = fn
}

// Process applies template substitution on LaTeX source with custom delimiters
func (e *Engine) Process(templatePath string) (string, error) {
	if templatePath == "" && e.Config.Template != "" {
		templatePath = e.Config.Template.String()
	}

	if templatePath == "" {
		return "", errors.New("no template file specified")
	}

	// Read template file
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", err
	}

	// Create new template with custom delimiters to avoid conflicts with LaTeX
	tmpl, err := template.New(templatePath).
		Funcs(e.FuncMap).
		Delims("delim[[", "]]").
		Parse(string(content))
	if err != nil {
		return "", err
	}

	// Apply template with variables
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, e.Config.Variables); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ProcessToFile processes the template and writes the result to a file
func (e *Engine) ProcessToFile(templatePath, outputPath string) error {
	if outputPath == "" && e.Config.Output != "" {
		outputPath = e.Config.Output.String()
	}

	if outputPath == "" {
		return errors.New("no output file specified")
	}

	result, err := e.Process(templatePath)
	if err != nil {
		return err
	}

	// Write processed content to output file
	return os.WriteFile(outputPath, []byte(result), 0644)
}
