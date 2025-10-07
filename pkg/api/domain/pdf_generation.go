// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package domain

import (
	"context"
	"fmt"
	"time"
)

// PDFGenerationRequest represents a request to generate a PDF
type PDFGenerationRequest struct {
	TemplatePath string
	Variables    map[string]interface{}
	Engine       string
	OutputPath   string
	Options      PDFGenerationOptions
}

// PDFGenerationOptions contains optional settings for PDF generation
type PDFGenerationOptions struct {
	DoConvert  bool
	DoClean    bool
	Conversion ConversionOptions
	Timeout    time.Duration
	Verbose    bool
	Debug      bool
}

// ConversionOptions contains settings for PDF to image conversion
type ConversionOptions struct {
	Enabled bool
	Formats []string
}

// PDFGenerationResult represents the result of PDF generation
type PDFGenerationResult struct {
	PDFPath    string
	ImagePaths []string
	Success    bool
	Error      error
	Metadata   PDFMetadata
}

// PDFMetadata contains metadata about the generated PDF
type PDFMetadata struct {
	FileSize    int64
	PageCount   int
	GeneratedAt time.Time
	Engine      string
	Template    string
}

// PDFGenerationService defines the interface for PDF generation
type PDFGenerationService interface {
	Generate(ctx context.Context, req PDFGenerationRequest) (PDFGenerationResult, error)
	ValidateRequest(req PDFGenerationRequest) error
	GetSupportedEngines() []string
	GetSupportedFormats() []string
}

// TemplateProcessingService defines the interface for template processing
type TemplateProcessingService interface {
	Process(ctx context.Context, templatePath string, variables map[string]interface{}) (string, error)
	ValidateTemplate(templatePath string) error
	GetTemplateVariables(templatePath string) ([]string, error)
}

// VariableResolver defines the interface for resolving complex variables
type VariableResolver interface {
	Resolve(variables map[string]interface{}) (map[string]string, error)
	Flatten(variables map[string]interface{}) map[string]string
	Validate(variables map[string]interface{}) error
}

// PDFValidator defines the interface for validating generated PDFs
type PDFValidator interface {
	Validate(pdfPath string) error
	GetMetadata(pdfPath string) (PDFMetadata, error)
	IsValidPDF(pdfPath string) bool
}

// Error types for domain-specific errors
type PDFGenerationError struct {
	Code    string
	Message string
	Details map[string]interface{}
}

func (e PDFGenerationError) Error() string {
	return fmt.Sprintf("PDF generation error [%s]: %s", e.Code, e.Message)
}

type TemplateProcessingError struct {
	Code    string
	Message string
	Details map[string]interface{}
}

func (e TemplateProcessingError) Error() string {
	return fmt.Sprintf("Template processing error [%s]: %s", e.Code, e.Message)
}

type VariableResolutionError struct {
	Code    string
	Message string
	Details map[string]interface{}
}

func (e VariableResolutionError) Error() string {
	return fmt.Sprintf("Variable resolution error [%s]: %s", e.Code, e.Message)
}

// Constants for error codes
const (
	ErrCodeTemplateNotFound    = "TEMPLATE_NOT_FOUND"
	ErrCodeTemplateInvalid     = "TEMPLATE_INVALID"
	ErrCodeEngineNotFound      = "ENGINE_NOT_FOUND"
	ErrCodeOutputPathInvalid   = "OUTPUT_PATH_INVALID"
	ErrCodeVariableInvalid     = "VARIABLE_INVALID"
	ErrCodePDFGenerationFailed = "PDF_GENERATION_FAILED"
	ErrCodePDFValidationFailed = "PDF_VALIDATION_FAILED"
	ErrCodeTimeoutExceeded     = "TIMEOUT_EXCEEDED"
)
