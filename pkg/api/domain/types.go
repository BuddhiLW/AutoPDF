// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package domain

import (
	"fmt"

	"github.com/BuddhiLW/AutoPDF/pkg/api"
)

// PDFMetadata is now defined in domain/generation package

// Error types for domain-specific errors
type PDFGenerationError struct {
	Code    string
	Message string
	Details *api.ErrorDetails
}

func (e PDFGenerationError) Error() string {
	return fmt.Sprintf("PDF generation error [%s]: %s", e.Code, e.Message)
}

type TemplateProcessingError struct {
	Code    string
	Message string
	Details *api.ErrorDetails
}

func (e TemplateProcessingError) Error() string {
	return fmt.Sprintf("Template processing error [%s]: %s", e.Code, e.Message)
}

type VariableResolutionError struct {
	Code    string
	Message string
	Details *api.ErrorDetails
}

func (e VariableResolutionError) Error() string {
	return fmt.Sprintf("Variable resolution error [%s]: %s", e.Code, e.Message)
}

// Constants for error codes
const (
	ErrCodeTemplateNotFound        = "TEMPLATE_NOT_FOUND"
	ErrCodeTemplateInvalid         = "TEMPLATE_INVALID"
	ErrCodeEngineNotFound          = "ENGINE_NOT_FOUND"
	ErrCodeOutputPathInvalid       = "OUTPUT_PATH_INVALID"
	ErrCodeVariableInvalid         = "VARIABLE_INVALID"
	ErrCodePDFGenerationFailed     = "PDF_GENERATION_FAILED"
	ErrCodePDFValidationFailed     = "PDF_VALIDATION_FAILED"
	ErrCodeTimeoutExceeded         = "TIMEOUT_EXCEEDED"
	ErrCodeWatchServiceUnavailable = "WATCH_SERVICE_UNAVAILABLE"
)
