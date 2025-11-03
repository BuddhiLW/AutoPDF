// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package errors

// DomainErrorFactory builds common AutoPDF errors with context
type DomainErrorFactory struct {
	formatter StringFormatter
}

func NewDomainErrorFactory(formatter StringFormatter) *DomainErrorFactory {
	if formatter == nil {
		formatter = &DefaultStringFormatter{}
	}
	return &DomainErrorFactory{formatter: formatter}
}

// Document processing errors

func (f *DomainErrorFactory) TemplateProcessingFailed(templatePath string, cause error) error {
	return NewInternalError(
		"TEMPLATE_PROCESSING_FAILED",
		"Failed to process template with variables",
	).WithBlame(f.formatter.Format("templatePath: %s", templatePath)).
		WithDetail("template_path", templatePath).
		WithCause(cause).
		WithSuggestions(
			"Check if the template file exists and is readable",
			"Verify all required variables are provided",
			"Check template syntax for errors",
			"Ensure template file permissions are correct",
		).Build()
}

func (f *DomainErrorFactory) LaTeXCompilationFailed(outputPath string, cause error) error {
	return NewInternalError(
		"LATEX_COMPILATION_FAILED",
		"Failed to compile LaTeX content to PDF",
	).WithBlame(f.formatter.Format("outputPath: %s", outputPath)).
		WithDetail("output_path", outputPath).
		WithCause(cause).
		WithSuggestions(
			"Check LaTeX syntax in the generated content",
			"Verify all required LaTeX packages are installed",
			"Check if the output directory is writable",
			"Review LaTeX compilation logs for specific errors",
		).Build()
}

func (f *DomainErrorFactory) PDFConversionFailed(pdfPath string, cause error) error {
	return NewInternalError(
		"PDF_CONVERSION_FAILED",
		"Failed to convert PDF to images",
	).WithBlame(f.formatter.Format("pdfPath: %s", pdfPath)).
		WithDetail("pdf_path", pdfPath).
		WithCause(cause).
		WithSuggestions(
			"Check if the PDF file exists and is valid",
			"Verify image conversion tools are installed",
			"Check available disk space for image output",
			"Review conversion tool logs for specific errors",
		).Build()
}

func (f *DomainErrorFactory) CleanupFailed(pdfPath string, cause error) error {
	return NewInternalError(
		"CLEANUP_FAILED",
		"Failed to clean auxiliary files",
	).WithBlame(f.formatter.Format("pdfPath: %s", pdfPath)).
		WithDetail("pdf_path", pdfPath).
		WithCause(cause).
		WithSuggestions(
			"Check file permissions for the output directory",
			"Verify auxiliary files are not locked by other processes",
			"Try manual cleanup of auxiliary files",
			"Check available disk space",
		).Build()
}

// Configuration and validation errors

func (f *DomainErrorFactory) EngineInvalid(engine string) error {
	return NewInvalidInputError("ENGINE_INVALID", "Invalid LaTeX engine").
		WithDetail("engine", engine).
		WithSuggestions(
			"Use one of: pdflatex, xelatex, lualatex",
		).Build()
}

func (f *DomainErrorFactory) OutputPathEmpty() error {
	return NewInvalidInputError("OUTPUT_PATH_EMPTY", "Output path must not be empty").
		WithSuggestions("Provide a valid output path").Build()
}

func (f *DomainErrorFactory) TemplatePathEmpty() error {
	return NewInvalidInputError("TEMPLATE_PATH_EMPTY", "Template path must not be empty").
		WithSuggestions("Provide a valid template path").Build()
}

func (f *DomainErrorFactory) VariableMissing(variable, template string) error {
	return NewValidationError("VARIABLE_MISSING", "Required template variable is missing").
		WithDetails(map[string]interface{}{"variable": variable, "template": template}).
		WithSuggestions(
			"Add the missing variable to the request",
			"Check the template's required variables",
		).Build()
}

func (f *DomainErrorFactory) PassesInvalid(passes int) error {
	return NewInvalidInputError("PASSES_INVALID", "Compilation passes must be between 1 and 10").
		WithDetail("passes", passes).
		WithSuggestions("Choose a value between 1 and 10").Build()
}
