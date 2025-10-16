// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package builders

import (
	"fmt"
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/BuddhiLW/AutoPDF/pkg/converter"
)

// PDFGenerationRequestBuilder builds PDF generation requests using the Builder pattern
type PDFGenerationRequestBuilder struct {
	request generation.PDFGenerationRequest
}

// NewPDFGenerationRequestBuilder creates a new builder
func NewPDFGenerationRequestBuilder() *PDFGenerationRequestBuilder {
	return &PDFGenerationRequestBuilder{
		request: generation.PDFGenerationRequest{
			Variables: generation.NewTemplateVariables(nil),
			Options: generation.PDFGenerationOptions{
				Conversion: generation.ConversionOptions{
					Enabled: false,
					Formats: []string{},
				},
				Timeout: 30 * time.Second,
				Verbose: 0,
				Debug: generation.DebugOptions{
					Enabled: false,
				},
			},
		},
	}
}

// WithTemplate sets the template path
func (b *PDFGenerationRequestBuilder) WithTemplate(templatePath string) *PDFGenerationRequestBuilder {
	b.request.TemplatePath = templatePath
	return b
}

// WithEngine sets the LaTeX engine
func (b *PDFGenerationRequestBuilder) WithEngine(engine string) *PDFGenerationRequestBuilder {
	b.request.Engine = engine
	return b
}

// WithOutput sets the output path
func (b *PDFGenerationRequestBuilder) WithOutput(outputPath string) *PDFGenerationRequestBuilder {
	b.request.OutputPath = outputPath
	return b
}

// WithVariable sets a simple variable
func (b *PDFGenerationRequestBuilder) WithVariable(key string, value interface{}) *PDFGenerationRequestBuilder {
	if b.request.Variables == nil {
		b.request.Variables = generation.NewTemplateVariables(nil)
	}
	// Convert value to config.Variable
	var variable config.Variable
	switch v := value.(type) {
	case string:
		variable = &config.StringVariable{Value: v}
	case int:
		variable = &config.NumberVariable{Value: float64(v)}
	case float64:
		variable = &config.NumberVariable{Value: v}
	case bool:
		variable = &config.BoolVariable{Value: v}
	default:
		// Fallback to string representation
		variable = &config.StringVariable{Value: fmt.Sprintf("%v", value)}
	}
	b.request.Variables.Set(key, variable)
	return b
}

// WithVariables sets multiple variables from a map
// For backward compatibility - converts map[string]interface{} to TemplateVariables
func (b *PDFGenerationRequestBuilder) WithVariables(variables map[string]interface{}) *PDFGenerationRequestBuilder {
	templateVars, err := generation.NewTemplateVariablesFromMap(variables)
	if err != nil {
		// Log error but don't fail - set empty variables
		b.request.Variables = generation.NewTemplateVariables(nil)
		return b
	}
	b.request.Variables = templateVars
	return b
}

// WithTemplateVariables sets variables using TemplateVariables Value Object
func (b *PDFGenerationRequestBuilder) WithTemplateVariables(variables *generation.TemplateVariables) *PDFGenerationRequestBuilder {
	if variables == nil {
		b.request.Variables = generation.NewTemplateVariables(nil)
	} else {
		b.request.Variables = variables
	}
	return b
}

// WithVariablesFromStruct sets variables by converting a struct using StructConverter
func (b *PDFGenerationRequestBuilder) WithVariablesFromStruct(s interface{}) *PDFGenerationRequestBuilder {
	// Create converter with defaults
	conv := converter.BuildWithDefaults()

	templateVars, err := generation.NewTemplateVariablesFromStruct(s, conv)
	if err != nil {
		// Log error but don't fail - set empty variables
		b.request.Variables = generation.NewTemplateVariables(nil)
		return b
	}

	b.request.Variables = templateVars
	return b
}

// WithComplexVariable sets a complex nested variable
// Deprecated: Use WithVariables or WithVariablesFromStruct instead
func (b *PDFGenerationRequestBuilder) WithComplexVariable(key string, value map[string]interface{}) *PDFGenerationRequestBuilder {
	if b.request.Variables == nil {
		b.request.Variables = generation.NewTemplateVariables(nil)
	}
	// Convert nested map to Variables and set
	nestedVars, err := generation.NewTemplateVariablesFromMap(value)
	if err == nil {
		// Set as nested value - for now just merge
		b.request.Variables.Merge(nestedVars)
	}
	return b
}

// WithArrayVariable sets an array variable
// Deprecated: Use WithVariables or WithVariablesFromStruct instead
func (b *PDFGenerationRequestBuilder) WithArrayVariable(key string, values []interface{}) *PDFGenerationRequestBuilder {
	if b.request.Variables == nil {
		b.request.Variables = generation.NewTemplateVariables(nil)
	}
	// Convert array to Variables
	sliceVar := config.NewSliceVariable()
	for _, val := range values {
		// Simple conversion
		strVal := fmt.Sprintf("%v", val)
		sliceVar.Values = append(sliceVar.Values, &config.StringVariable{Value: strVal})
	}
	b.request.Variables.Set(key, sliceVar)
	return b
}

// WithConversion enables PDF to image conversion
func (b *PDFGenerationRequestBuilder) WithConversion(enabled bool, formats ...string) *PDFGenerationRequestBuilder {
	b.request.Options.DoConvert = enabled
	b.request.Options.Conversion.Enabled = enabled
	b.request.Options.Conversion.Formats = formats
	return b
}

// WithCleanup enables auxiliary file cleanup
func (b *PDFGenerationRequestBuilder) WithCleanup(enabled bool) *PDFGenerationRequestBuilder {
	b.request.Options.DoClean = enabled
	return b
}

// WithTimeout sets the generation timeout
func (b *PDFGenerationRequestBuilder) WithTimeout(timeout time.Duration) *PDFGenerationRequestBuilder {
	b.request.Options.Timeout = timeout
	return b
}

// WithVerbose enables verbose logging
func (b *PDFGenerationRequestBuilder) WithVerbose(level int) *PDFGenerationRequestBuilder {
	b.request.Options.Verbose = level
	return b
}

// WithDebug enables debug logging
func (b *PDFGenerationRequestBuilder) WithDebug(debugOptions generation.DebugOptions) *PDFGenerationRequestBuilder {
	b.request.Options.Debug = debugOptions
	return b
}

// WithWatchMode enables file watching for automatic rebuilds
func (b *PDFGenerationRequestBuilder) WithWatchMode(enabled bool) *PDFGenerationRequestBuilder {
	b.request.Options.WatchMode = enabled
	return b
}

// Build constructs the final PDF generation request
func (b *PDFGenerationRequestBuilder) Build() generation.PDFGenerationRequest {
	return b.request
}

// PDFGenerationOptionsBuilder builds PDF generation options
type PDFGenerationOptionsBuilder struct {
	options generation.PDFGenerationOptions
}

// NewPDFGenerationOptionsBuilder creates a new options builder
func NewPDFGenerationOptionsBuilder() *PDFGenerationOptionsBuilder {
	return &PDFGenerationOptionsBuilder{
		options: generation.PDFGenerationOptions{
			Conversion: generation.ConversionOptions{
				Enabled: false,
				Formats: []string{},
			},
			Timeout: 30 * time.Second,
			Verbose: 0,
			Debug: generation.DebugOptions{
				Enabled: false,
			},
		},
	}
}

// EnableConversion enables PDF to image conversion
func (b *PDFGenerationOptionsBuilder) EnableConversion(formats ...string) *PDFGenerationOptionsBuilder {
	b.options.DoConvert = true
	b.options.Conversion.Enabled = true
	b.options.Conversion.Formats = formats
	return b
}

// DisableConversion disables PDF to image conversion
func (b *PDFGenerationOptionsBuilder) DisableConversion() *PDFGenerationOptionsBuilder {
	b.options.DoConvert = false
	b.options.Conversion.Enabled = false
	b.options.Conversion.Formats = []string{}
	return b
}

// EnableCleanup enables auxiliary file cleanup
func (b *PDFGenerationOptionsBuilder) EnableCleanup() *PDFGenerationOptionsBuilder {
	b.options.DoClean = true
	return b
}

// DisableCleanup disables auxiliary file cleanup
func (b *PDFGenerationOptionsBuilder) DisableCleanup() *PDFGenerationOptionsBuilder {
	b.options.DoClean = false
	return b
}

// SetTimeout sets the generation timeout
func (b *PDFGenerationOptionsBuilder) SetTimeout(timeout time.Duration) *PDFGenerationOptionsBuilder {
	b.options.Timeout = timeout
	return b
}

// SetVerbose sets verbose logging
func (b *PDFGenerationOptionsBuilder) SetVerbose(level int) *PDFGenerationOptionsBuilder {
	b.options.Verbose = level
	return b
}

// SetDebug sets debug logging
func (b *PDFGenerationOptionsBuilder) SetDebug(debugOptions generation.DebugOptions) *PDFGenerationOptionsBuilder {
	b.options.Debug = debugOptions
	return b
}

// SetWatchMode sets file watching mode
func (b *PDFGenerationOptionsBuilder) SetWatchMode(enabled bool) *PDFGenerationOptionsBuilder {
	b.options.WatchMode = enabled
	return b
}

// Build constructs the final options
func (b *PDFGenerationOptionsBuilder) Build() generation.PDFGenerationOptions {
	return b.options
}

// ConfigBuilder builds configuration objects
type ConfigBuilder struct {
	config *config.Config
}

// NewConfigBuilder creates a new config builder
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: &config.Config{
			Variables: *config.NewVariables(),
		},
	}
}

// WithTemplate sets the template path
func (b *ConfigBuilder) WithTemplate(templatePath string) *ConfigBuilder {
	b.config.Template = config.Template(templatePath)
	return b
}

// WithOutput sets the output path
func (b *ConfigBuilder) WithOutput(outputPath string) *ConfigBuilder {
	b.config.Output = config.Output(outputPath)
	return b
}

// WithEngine sets the LaTeX engine
func (b *ConfigBuilder) WithEngine(engine string) *ConfigBuilder {
	b.config.Engine = config.Engine(engine)
	return b
}

// WithVariable sets a simple variable
func (b *ConfigBuilder) WithVariable(key, value string) *ConfigBuilder {
	b.config.Variables.SetString(key, value)
	return b
}

// WithComplexVariable sets a complex variable
func (b *ConfigBuilder) WithComplexVariable(key string, value map[string]interface{}) *ConfigBuilder {
	// Convert map to Variable and set it
	// This would need proper implementation based on the Variable interface
	return b
}

// WithConversion enables conversion
func (b *ConfigBuilder) WithConversion(enabled bool, formats ...string) *ConfigBuilder {
	b.config.Conversion.Enabled = enabled
	b.config.Conversion.Formats = formats
	return b
}

// Build constructs the final config
func (b *ConfigBuilder) Build() *config.Config {
	return b.config
}
