// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package examples

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/pkg/api"
	apilogger "github.com/BuddhiLW/AutoPDF/pkg/api/adapters/logger"
	apiconfig "github.com/BuddhiLW/AutoPDF/pkg/api/config"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
	"github.com/BuddhiLW/AutoPDF/pkg/api/middleware"
	"github.com/BuddhiLW/AutoPDF/pkg/api/services"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// ExampleComplexVariables demonstrates how to use complex variables with AutoPDF
func ExampleComplexVariables() {
	// Create a config with complex variables
	cfg := &config.Config{
		Template:  config.Template("template.tex"),
		Output:    config.Output("output.pdf"),
		Engine:    config.Engine("pdflatex"),
		Variables: *config.NewVariables(),
	}

	// Set up complex nested variables
	// This represents the YAML structure:
	// variables:
	//   foo:
	//     bar:
	//       - bar1
	//       - bar2
	//     zet: [1, 2, 3]
	//   foo_bar: [foo, bar]

	// Create nested structure
	fooMap := config.NewMapVariable()

	// Set bar as an array
	barArray := config.NewSliceVariable()
	barArray.Values = []config.Variable{
		&config.StringVariable{Value: "bar1"},
		&config.StringVariable{Value: "bar2"},
	}
	fooMap.Set("bar", barArray)

	// Set zet as an array of numbers
	zetArray := config.NewSliceVariable()
	zetArray.Values = []config.Variable{
		&config.NumberVariable{Value: 1},
		&config.NumberVariable{Value: 2},
		&config.NumberVariable{Value: 3},
	}
	fooMap.Set("zet", zetArray)

	// Set the foo map
	cfg.Variables.Set("foo", fooMap)

	// Set foo_bar as a simple array
	fooBarArray := config.NewSliceVariable()
	fooBarArray.Values = []config.Variable{
		&config.StringVariable{Value: "foo"},
		&config.StringVariable{Value: "bar"},
	}
	cfg.Variables.Set("foo_bar", fooBarArray)

	// Create logger and API service
	testLogger := logger.NewLoggerAdapter(logger.Debug, "stdout")
	apiService := services.NewPDFGenerationAPIService(cfg, testLogger)

	// Example 1: Using builder pattern with complex variables
	ctx := context.Background()

	// Build complex variables using the builder
	complexVars := map[string]interface{}{
		"title":  "My Document",
		"author": "AutoPDF User",
		"date":   "2025-01-07",
		"foo": map[string]interface{}{
			"bar": []string{"bar1", "bar2"},
			"zet": []int{1, 2, 3},
		},
		"foo_bar": []string{"foo", "bar"},
		"nested": map[string]interface{}{
			"deep": map[string]interface{}{
				"value":   "nested value",
				"numbers": []float64{1.5, 2.5, 3.5},
			},
		},
	}

	// Generate PDF with complex variables
	pdfBytes, imagePaths, err := apiService.GeneratePDF(ctx, "template.tex", "output.pdf", complexVars)
	if err != nil {
		log.Fatalf("Failed to generate PDF: %v", err)
	}

	fmt.Printf("Generated PDF: %d bytes\n", len(pdfBytes))
	fmt.Printf("Image paths: %v\n", imagePaths)
}

// ExampleBuilderPattern demonstrates the builder pattern for PDF generation
func ExampleBuilderPattern() {
	// Create config
	cfg := &config.Config{
		Template:  config.Template("template.tex"),
		Output:    config.Output("output.pdf"),
		Engine:    config.Engine("pdflatex"),
		Variables: *config.NewVariables(),
	}

	// Create logger and API service
	testLogger := logger.NewLoggerAdapter(logger.Debug, "stdout")
	apiService := services.NewPDFGenerationAPIService(cfg, testLogger)

	// Build request using builder pattern
	options := services.NewPDFGenerationOptions("template.tex", "output.pdf").
		WithEngine("pdflatex").
		WithVariable("title", "Builder Pattern Example").
		WithVariable("author", "AutoPDF").
		WithVariable("date", "2025-01-07").
		WithVariable("metadata", map[string]interface{}{
			"version": "1.0",
			"tags":    []string{"example", "builder", "pattern"},
			"settings": map[string]interface{}{
				"verbose": true,
				"debug":   false,
			},
		}).
		WithConversion(true, "png", "jpeg").
		WithCleanup(false).
		WithTimeout(30).
		WithVerbose(true)

	// Generate PDF
	ctx := context.Background()
	pdfBytes, imagePaths, err := apiService.GeneratePDFWithOptions(ctx, *options)
	if err != nil {
		log.Fatalf("Failed to generate PDF: %v", err)
	}

	fmt.Printf("Generated PDF: %d bytes\n", len(pdfBytes))
	fmt.Printf("Image paths: %v\n", imagePaths)
}

// ExampleTemplateVariables demonstrates template variable extraction
func ExampleTemplateVariables() {
	cfg := &config.Config{
		Template:  config.Template("template.tex"),
		Output:    config.Output("output.pdf"),
		Engine:    config.Engine("pdflatex"),
		Variables: *config.NewVariables(),
	}

	// Create logger and API service
	testLogger := logger.NewLoggerAdapter(logger.Debug, "stdout")
	apiService := services.NewPDFGenerationAPIService(cfg, testLogger)

	// Extract variables from template
	variables, err := apiService.GetTemplateVariables("template.tex")
	if err != nil {
		log.Fatalf("Failed to extract template variables: %v", err)
	}

	fmt.Printf("Template variables: %v\n", variables)
}

// ExampleVariableFlattening demonstrates variable flattening
func ExampleVariableFlattening() {
	// Create complex variables
	complexVars := map[string]interface{}{
		"title":  "My Document",
		"author": "AutoPDF User",
		"metadata": map[string]interface{}{
			"version": "1.0",
			"tags":    []string{"example", "flattening"},
			"settings": map[string]interface{}{
				"verbose": true,
				"debug":   false,
			},
		},
		"items": []interface{}{
			"item1",
			"item2",
			map[string]interface{}{
				"name":  "nested item",
				"value": 42,
			},
		},
	}

	// Create config and set variables
	cfg := &config.Config{
		Template:  config.Template("template.tex"),
		Output:    config.Output("output.pdf"),
		Engine:    config.Engine("pdflatex"),
		Variables: *config.NewVariables(),
	}

	// Set complex variables
	for key, value := range complexVars {
		cfg.Variables.SetString(key, fmt.Sprintf("%v", value))
	}

	// Flatten variables
	flattened := cfg.Variables.Flatten()

	fmt.Println("Flattened variables:")
	for key, value := range flattened {
		fmt.Printf("  %s: %s\n", key, value)
	}
}

// ExampleTemplateProcessing demonstrates template processing with complex variables
func ExampleTemplateProcessing() {
	// Template content with complex variable references
	// (In real usage, this would be an existing template file)

	// Write template to file
	// (In real usage, this would be an existing template file)

	// Create variables
	variables := map[string]interface{}{
		"title":  "Complex Variables Example",
		"author": "AutoPDF",
		"date":   "2025-01-07",
		"metadata": map[string]interface{}{
			"version": "1.0",
			"tags":    []string{"example", "complex"},
			"settings": map[string]interface{}{
				"verbose": true,
				"debug":   false,
			},
		},
		"items": []interface{}{
			"First item",
			"Second item",
			map[string]interface{}{
				"name":  "Nested item",
				"value": 42,
			},
		},
	}

	// Create config
	cfg := &config.Config{
		Template:  config.Template("template.tex"),
		Output:    config.Output("output.pdf"),
		Engine:    config.Engine("pdflatex"),
		Variables: *config.NewVariables(),
	}

	// Create logger and API service
	testLogger := logger.NewLoggerAdapter(logger.Debug, "stdout")
	apiService := services.NewPDFGenerationAPIService(cfg, testLogger)

	// Generate PDF
	ctx := context.Background()
	pdfBytes, imagePaths, err := apiService.GeneratePDF(ctx, "template.tex", "output.pdf", variables)
	if err != nil {
		log.Fatalf("Failed to generate PDF: %v", err)
	}

	fmt.Printf("Generated PDF: %d bytes\n", len(pdfBytes))
	fmt.Printf("Image paths: %v\n", imagePaths)
}

// ExampleDebuggingCapabilities demonstrates comprehensive debugging features
func ExampleDebuggingCapabilities() {
	fmt.Println("=== AutoPDF API Debugging Capabilities Demo ===")

	// 1. Environment-based debug configuration
	fmt.Println("\n1. Environment-based Debug Configuration:")
	debugConfig := apiconfig.LoadDebugConfigFromEnv()
	fmt.Printf("Debug enabled: %v\n", debugConfig.Enabled)
	fmt.Printf("Log directory: %s\n", debugConfig.LogDirectory)
	fmt.Printf("Concrete file directory: %s\n", debugConfig.ConcreteFileDir)

	// 2. Create API logger factory with debug enabled
	fmt.Println("\n2. Creating API Logger Factory:")
	loggerFactory := apilogger.NewAPILoggerFactory(true, "/tmp/autopdf-debug-logs")
	requestLogger := loggerFactory.CreateRequestLogger("demo-request-123", true)
	fmt.Println("Request-scoped logger created with debug enabled")

	// 3. Demonstrate structured logging
	fmt.Println("\n3. Structured Logging Demo:")
	requestLogger.InfoWithFields("Starting PDF generation demo",
		"request_id", "demo-request-123",
		"template", "debug-template.tex",
		"variables_count", 5,
	)

	requestLogger.DebugWithFields("Variable resolution starting",
		"request_id", "demo-request-123",
		"variable_keys", []string{"title", "author", "date", "metadata", "items"},
	)

	// 4. Create debug options for PDF generation
	fmt.Println("\n4. Debug Options Configuration:")
	debugOptions := generation.DebugOptions{
		Enabled:            true,
		LogToFile:          true,
		LogFilePath:        "/tmp/autopdf-debug-logs/demo-request-123.log",
		CreateConcreteFile: true,
		RequestID:          "demo-request-123",
	}

	pdfOptions := generation.PDFGenerationOptions{
		Debug:     debugOptions,
		DoClean:   true,
		Verbose:   2,
		Force:     false,
		RequestID: "demo-request-123",
		DoConvert: false,
		Timeout:   30 * time.Second,
		Conversion: generation.ConversionOptions{
			Enabled: false,
			Formats: []string{},
		},
	}

	fmt.Printf("Debug options: %+v\n", debugOptions)
	fmt.Printf("PDF generation options: %+v\n", pdfOptions)

	// 5. Demonstrate error details with structured logging
	fmt.Println("\n5. Error Details with Structured Logging:")
	errorDetails := api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
		WithTemplatePath("/test/template.tex").
		AddContext("error_type", "syntax_error").
		AddContext("template_name", "debug_template").
		AddContext("line_number", "42")

	// Log the error with structured logging
	errorDetails.LogError(requestLogger)

	fmt.Println("Error logged with structured logging - check log files for details")
}

// ExampleMalformedConfigHandling demonstrates error handling for malformed configurations
func ExampleMalformedConfigHandling() {
	fmt.Println("\n=== Malformed Configuration Error Handling Demo ===")

	// Create a logger for error demonstration
	testLogger := logger.NewLoggerAdapter(logger.Debug, "stdout")

	// 1. Missing required template file
	fmt.Println("\n1. Missing Template File Error:")
	missingTemplateError := api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
		WithTemplatePath("/nonexistent/template.tex").
		AddContext("error_type", "file_not_found").
		AddContext("operation", "template_validation").
		WithError(fmt.Errorf("template file does not exist"))

	missingTemplateError.LogErrorWithMessage(testLogger, "Template validation failed")

	// 2. Invalid variable configuration
	fmt.Println("\n2. Invalid Variable Configuration Error:")
	invalidVarError := api.NewErrorDetails(api.ErrorCategoryVariable, api.ErrorSeverityMedium).
		AddContext("error_type", "invalid_variable_type").
		AddContext("variable_name", "invalid_var").
		AddContext("expected_type", "string").
		AddContext("actual_type", "complex_object").
		WithError(fmt.Errorf("variable 'invalid_var' has unsupported type"))

	invalidVarError.LogErrorWithMessage(testLogger, "Variable validation failed")

	// 3. Malformed YAML configuration
	fmt.Println("\n3. Malformed YAML Configuration Error:")
	yamlError := api.NewErrorDetails(api.ErrorCategoryConfiguration, api.ErrorSeverityHigh).
		WithFilePath("/config/malformed.yaml").
		AddContext("error_type", "yaml_parse_error").
		AddContext("line_number", "15").
		AddContext("column_number", "8").
		WithError(fmt.Errorf("yaml: line 15: found character that cannot start any token"))

	yamlError.LogErrorWithMessage(testLogger, "Configuration parsing failed")

	// 4. Template syntax errors
	fmt.Println("\n4. Template Syntax Error:")
	templateSyntaxError := api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
		WithTemplatePath("/templates/broken.tex").
		AddContext("error_type", "latex_syntax_error").
		AddContext("line_number", "23").
		AddContext("error_context", "\\begin{document} without \\documentclass").
		WithError(fmt.Errorf("LaTeX compilation failed: missing \\documentclass"))

	templateSyntaxError.LogErrorWithMessage(testLogger, "Template compilation failed")

	// 5. Variable resolution errors
	fmt.Println("\n5. Variable Resolution Error:")
	varResolutionError := api.NewErrorDetails(api.ErrorCategoryVariable, api.ErrorSeverityMedium).
		AddContext("error_type", "variable_not_found").
		AddContext("variable_name", "undefined_var").
		AddContext("template_path", "/templates/missing-vars.tex").
		AddContext("available_variables", "title,author,date").
		WithError(fmt.Errorf("variable 'undefined_var' not found in template"))

	varResolutionError.LogErrorWithMessage(testLogger, "Variable resolution failed")
}

// ExampleTemplateValidationErrors demonstrates template validation and error reporting
func ExampleTemplateValidationErrors() {
	fmt.Println("\n=== Template Validation Error Demo ===")

	testLogger := logger.NewLoggerAdapter(logger.Debug, "stdout")

	// 1. Template with undefined variables
	fmt.Println("\n1. Template with Undefined Variables:")
	undefinedVarsError := api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityMedium).
		WithTemplatePath("/templates/undefined-vars.tex").
		AddContext("error_type", "undefined_variables").
		AddContext("undefined_variables", "missing_title,missing_author,missing_date").
		AddContext("template_content_preview", "\\title{delim[[vars.missing_title]]}\n\\author{delim[[vars.missing_author]]}").
		WithError(fmt.Errorf("template contains undefined variables"))

	undefinedVarsError.LogErrorWithMessage(testLogger, "Template validation failed - undefined variables")

	// 2. Template with malformed variable syntax
	fmt.Println("\n2. Template with Malformed Variable Syntax:")
	malformedSyntaxError := api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
		WithTemplatePath("/templates/malformed-syntax.tex").
		AddContext("error_type", "malformed_variable_syntax").
		AddContext("error_line", "\\title{delim[[vars.title}").
		AddContext("expected_syntax", "delim[[vars.variable_name]]").
		AddContext("actual_syntax", "delim[[vars.title}").
		WithError(fmt.Errorf("malformed variable syntax: missing closing delimiter"))

	malformedSyntaxError.LogErrorWithMessage(testLogger, "Template validation failed - malformed syntax")

	// 3. Template with LaTeX compilation errors
	fmt.Println("\n3. Template with LaTeX Compilation Errors:")
	latexCompilationError := api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
		WithTemplatePath("/templates/compilation-error.tex").
		AddContext("error_type", "latex_compilation_error").
		AddContext("latex_error", "Undefined control sequence. \\maketitl").
		AddContext("error_line", "15").
		AddContext("suggestion", "Did you mean \\maketitle?").
		WithError(fmt.Errorf("LaTeX compilation failed"))

	latexCompilationError.LogErrorWithMessage(testLogger, "Template compilation failed")

	// 4. Template with missing packages
	fmt.Println("\n4. Template with Missing LaTeX Packages:")
	missingPackageError := api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityMedium).
		WithTemplatePath("/templates/missing-package.tex").
		AddContext("error_type", "missing_latex_package").
		AddContext("missing_package", "geometry").
		AddContext("usage_line", "\\usepackage{geometry}").
		AddContext("suggestion", "Install geometry package or remove usage").
		WithError(fmt.Errorf("LaTeX package 'geometry' not found"))

	missingPackageError.LogErrorWithMessage(testLogger, "Template validation failed - missing package")
}

// ExampleVariableValidationErrors demonstrates variable validation and error reporting
func ExampleVariableValidationErrors() {
	fmt.Println("\n=== Variable Validation Error Demo ===")

	testLogger := logger.NewLoggerAdapter(logger.Debug, "stdout")

	// 1. Variables with wrong types
	fmt.Println("\n1. Variables with Wrong Types:")
	wrongTypeError := api.NewErrorDetails(api.ErrorCategoryVariable, api.ErrorSeverityMedium).
		AddContext("error_type", "type_mismatch").
		AddContext("variable_name", "page_count").
		AddContext("expected_type", "number").
		AddContext("actual_type", "string").
		AddContext("actual_value", "not-a-number").
		AddContext("template_usage", "delim[[vars.page_count]]").
		WithError(fmt.Errorf("variable 'page_count' expected number, got string"))

	wrongTypeError.LogErrorWithMessage(testLogger, "Variable type validation failed")

	// 2. Variables with invalid values
	fmt.Println("\n2. Variables with Invalid Values:")
	invalidValueError := api.NewErrorDetails(api.ErrorCategoryVariable, api.ErrorSeverityMedium).
		AddContext("error_type", "invalid_value").
		AddContext("variable_name", "email").
		AddContext("actual_value", "not-an-email").
		AddContext("validation_rule", "must be valid email format").
		AddContext("template_usage", "delim[[vars.email]]").
		WithError(fmt.Errorf("variable 'email' has invalid format"))

	invalidValueError.LogErrorWithMessage(testLogger, "Variable value validation failed")

	// 3. Missing required variables
	fmt.Println("\n3. Missing Required Variables:")
	missingRequiredError := api.NewErrorDetails(api.ErrorCategoryVariable, api.ErrorSeverityHigh).
		AddContext("error_type", "missing_required_variable").
		AddContext("variable_name", "title").
		AddContext("template_usage", "\\title{delim[[vars.title]]}").
		AddContext("available_variables", "author,date,content").
		WithError(fmt.Errorf("required variable 'title' is missing"))

	missingRequiredError.LogErrorWithMessage(testLogger, "Required variable validation failed")

	// 4. Variables with circular references
	fmt.Println("\n4. Variables with Circular References:")
	circularRefError := api.NewErrorDetails(api.ErrorCategoryVariable, api.ErrorSeverityHigh).
		AddContext("error_type", "circular_reference").
		AddContext("variable_chain", "var_a,var_b,var_c,var_a").
		AddContext("circular_variable", "var_a").
		WithError(fmt.Errorf("circular reference detected in variable chain"))

	circularRefError.LogErrorWithMessage(testLogger, "Variable reference validation failed")
}

// ExampleDebuggingWorkflow demonstrates a complete debugging workflow
func ExampleDebuggingWorkflow() {
	fmt.Println("\n=== Complete Debugging Workflow Demo ===")

	// 1. Set up debug environment
	fmt.Println("\n1. Setting up Debug Environment:")
	os.Setenv("AUTOPDF_API_DEBUG", "true")
	os.Setenv("AUTOPDF_API_LOG_DIR", "/tmp/autopdf-debug-workflow")
	os.Setenv("AUTOPDF_API_CONCRETE_DIR", "/tmp/autopdf-concrete-workflow")

	debugConfig := apiconfig.LoadDebugConfigFromEnv()
	fmt.Printf("Debug environment configured: %+v\n", debugConfig)

	// 2. Create request-scoped logger
	fmt.Println("\n2. Creating Request-Scoped Logger:")
	loggerFactory := apilogger.NewAPILoggerFactory(true, debugConfig.LogDirectory)
	requestLogger := loggerFactory.CreateRequestLogger("workflow-demo-456", true)
	fmt.Println("Request logger created with ID: workflow-demo-456")

	// 3. Simulate a problematic template and variables
	fmt.Println("\n3. Simulating Problematic Template and Variables:")

	// Create a template with issues
	templateContent := `
\documentclass{article}
\usepackage[utf8]{inputenc}

\title{delim[[vars.title]]}
\author{delim[[vars.author]]}
\date{delim[[vars.date]]}

\begin{document}

\maketitle

\section{Introduction}
This document has some issues:
- Missing variable: delim[[vars.missing_var]]
- Malformed syntax: delim[[vars.title}
- Undefined command: \undefinedcommand

\end{document}
`

	// Write problematic template
	templatePath := "/tmp/problematic-template.tex"
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		log.Printf("Failed to write template: %v", err)
		return
	}
	fmt.Printf("Problematic template written to: %s\n", templatePath)

	// 4. Create variables with issues
	fmt.Println("\n4. Creating Variables with Issues:")
	problematicVars := map[string]interface{}{
		"title":  "Debug Workflow Demo", // Valid
		"author": "AutoPDF Debugger",    // Valid
		"date":   "2025-01-15",          // Valid
		// Missing: "missing_var" - will cause undefined variable error
		"invalid_type": map[string]interface{}{ // Complex type that might cause issues
			"nested": "value",
		},
	}

	requestLogger.InfoWithFields("Variables prepared for debugging",
		"request_id", "workflow-demo-456",
		"variable_count", len(problematicVars),
		"template_path", templatePath,
	)

	// 5. Demonstrate error detection and logging
	fmt.Println("\n5. Error Detection and Logging:")

	// Simulate template validation errors
	templateValidationError := api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
		WithTemplatePath(templatePath).
		AddContext("error_type", "multiple_template_issues").
		AddContext("issues", "undefined variable: missing_var,malformed syntax: delim[[vars.title},undefined LaTeX command: \\undefinedcommand").
		AddContext("line_numbers", "12,13,15").
		WithError(fmt.Errorf("template validation failed with multiple issues"))

	templateValidationError.LogErrorWithMessage(requestLogger, "Template validation failed")

	// 6. Create concrete file for debugging
	fmt.Println("\n6. Creating Concrete File for Debugging:")
	_ = generation.DebugOptions{
		Enabled:            true,
		LogToFile:          true,
		LogFilePath:        filepath.Join(debugConfig.LogDirectory, "workflow-demo-456.log"),
		CreateConcreteFile: true,
		RequestID:          "workflow-demo-456",
	}

	// Simulate creating concrete file
	concreteContent := `
\documentclass{article}
\usepackage[utf8]{inputenc}

\title{Debug Workflow Demo}
\author{AutoPDF Debugger}
\date{2025-01-15}

\begin{document}

\maketitle

\section{Introduction}
This document has some issues:
- Missing variable: delim[[vars.missing_var]]
- Malformed syntax: delim[[vars.title}
- Undefined command: \undefinedcommand

\end{document}
`

	concretePath := filepath.Join(debugConfig.ConcreteFileDir, "autopdf-concrete-workflow-demo-456.tex")
	err = os.MkdirAll(debugConfig.ConcreteFileDir, 0755)
	if err == nil {
		err = os.WriteFile(concretePath, []byte(concreteContent), 0644)
		if err == nil {
			requestLogger.InfoWithFields("Concrete template file created",
				"request_id", "workflow-demo-456",
				"concrete_path", concretePath,
				"original_template", templatePath,
			)
			fmt.Printf("Concrete file created: %s\n", concretePath)
		}
	}

	// 7. Demonstrate error recovery suggestions
	fmt.Println("\n7. Error Recovery Suggestions:")
	recoverySuggestions := api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityMedium).
		WithTemplatePath(templatePath).
		AddContext("error_type", "recovery_suggestions").
		AddContext("suggestions", "missing_var: Add 'missing_var' to variables or remove from template,malformed_syntax: Fix syntax: delim[[vars.title]] (add missing closing bracket),undefined_command: Remove \\undefinedcommand or define it").
		AddContext("auto_fix_available", "true")

	recoverySuggestions.LogErrorWithMessage(requestLogger, "Error recovery suggestions provided")

	// 8. Cleanup
	fmt.Println("\n8. Cleanup:")
	os.Remove(templatePath)
	os.Remove(concretePath)
	fmt.Println("Temporary files cleaned up")

	fmt.Println("\n=== Debugging Workflow Complete ===")
	fmt.Println("Check the following for debugging information:")
	fmt.Printf("- Log file: %s/workflow-demo-456.log\n", debugConfig.LogDirectory)
	fmt.Printf("- Concrete file: %s (if created)\n", concretePath)
}

// ExampleHTTPMiddlewareDebugging demonstrates HTTP middleware debugging capabilities
func ExampleHTTPMiddlewareDebugging() {
	fmt.Println("\n=== HTTP Middleware Debugging Demo ===")

	// 1. Create debug configuration
	_ = &apiconfig.APIDebugConfig{
		Enabled:         true,
		LogDirectory:    "/tmp/autopdf-middleware-logs",
		ConcreteFileDir: "/tmp/autopdf-middleware-concrete",
	}

	// 2. Demonstrate middleware context injection
	fmt.Println("\n1. Middleware Context Injection:")

	// Simulate request headers
	requestHeaders := map[string]string{
		"X-AutoPDF-Debug":   "true",
		"X-AutoPDF-Verbose": "3",
		"X-AutoPDF-Clean":   "true",
		"X-AutoPDF-Force":   "false",
		"X-AutoPDF-Convert": "true",
		"X-AutoPDF-Timeout": "45s",
	}

	fmt.Println("Request headers:")
	for key, value := range requestHeaders {
		fmt.Printf("  %s: %s\n", key, value)
	}

	// 3. Simulate middleware processing
	fmt.Println("\n2. Middleware Processing Simulation:")

	// Generate request ID (simulating middleware)
	requestID := "middleware-demo-789"

	// Parse options from headers (simulating middleware logic)
	options := generation.PDFGenerationOptions{
		Debug: generation.DebugOptions{
			Enabled:            true,
			LogToFile:          true,
			CreateConcreteFile: true,
			RequestID:          requestID,
		},
		DoClean:   true,
		Verbose:   3,
		Force:     false,
		RequestID: requestID,
		DoConvert: true,
		Timeout:   45 * time.Second,
		Conversion: generation.ConversionOptions{
			Enabled: true,
			Formats: []string{"png", "jpeg"},
		},
	}

	fmt.Printf("Parsed options: %+v\n", options)

	// 4. Demonstrate context extraction
	fmt.Println("\n3. Context Extraction:")

	// Simulate context with options
	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.OptionsContextKey, options)
	ctx = context.WithValue(ctx, middleware.RequestIDContextKey, requestID)

	// Extract options from context
	extractedOptions, ok := middleware.GetOptionsFromContext(ctx)
	if ok {
		fmt.Printf("Successfully extracted options from context: %+v\n", extractedOptions)
	}

	extractedRequestID, ok := middleware.GetRequestIDFromContext(ctx)
	if ok {
		fmt.Printf("Successfully extracted request ID from context: %s\n", extractedRequestID)
	}

	// 5. Demonstrate request logging
	fmt.Println("\n4. Request Logging:")
	requestLogger := logger.NewLoggerAdapter(logger.Debug, "stdout")

	requestLogger.InfoWithFields("HTTP request processed with debug middleware",
		"request_id", requestID,
		"debug_enabled", options.Debug.Enabled,
		"verbose_level", options.Verbose,
		"clean_enabled", options.DoClean,
		"force_enabled", options.Force,
		"convert_enabled", options.DoConvert,
		"timeout", options.Timeout,
	)

	fmt.Println("\n=== HTTP Middleware Debugging Complete ===")
}
