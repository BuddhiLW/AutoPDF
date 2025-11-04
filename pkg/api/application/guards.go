// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"

	"github.com/BuddhiLW/AutoPDF/pkg/api"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
)

// Constants for content preview and validation
const (
	DefaultContentPreviewLength = 500
	MinTemplatePathLength       = 1
	MinEngineNameLength         = 1
	MinOutputPathLength         = 1
)

// ValidationGuard defines the interface for validation guards
type ValidationGuard interface {
	Validate(ctx context.Context, req generation.PDFGenerationRequest) error
}

// RequestValidationGuard validates PDF generation requests
type RequestValidationGuard struct {
	templateService  generation.TemplateProcessingService
	variableResolver generation.VariableResolver
}

// NewRequestValidationGuard creates a new request validation guard
func NewRequestValidationGuard(
	templateService generation.TemplateProcessingService,
	variableResolver generation.VariableResolver,
) *RequestValidationGuard {
	return &RequestValidationGuard{
		templateService:  templateService,
		variableResolver: variableResolver,
	}
}

// Validate implements ValidationGuard interface
func (g *RequestValidationGuard) Validate(ctx context.Context, req generation.PDFGenerationRequest) error {
	validators := []func() error{
		func() error { return g.validateTemplatePath(req.TemplatePath) },
		func() error { return g.validateEngine(req.Engine) },
		func() error { return g.validateOutputPath(req.OutputPath) },
		func() error { return g.validateTemplateFile(req.TemplatePath) },
		func() error { return g.validateVariables(req.Variables) },
	}

	for _, validator := range validators {
		if err := validator(); err != nil {
			return err
		}
	}

	return nil
}

// validateTemplatePath guards against empty template paths
func (g *RequestValidationGuard) validateTemplatePath(templatePath string) error {
	if len(templatePath) < MinTemplatePathLength {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodeTemplateNotFound,
			Message: "Template path is required",
		}
	}
	return nil
}

// validateEngine guards against empty engine names
func (g *RequestValidationGuard) validateEngine(engine string) error {
	if len(engine) < MinEngineNameLength {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodeEngineNotFound,
			Message: "LaTeX engine is required",
		}
	}
	return nil
}

// validateOutputPath guards against empty output paths
func (g *RequestValidationGuard) validateOutputPath(outputPath string) error {
	if len(outputPath) < MinOutputPathLength {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodeOutputPathInvalid,
			Message: "Output path is required",
		}
	}
	return nil
}

// validateTemplateFile guards against invalid template files
func (g *RequestValidationGuard) validateTemplateFile(templatePath string) error {
	if err := g.templateService.ValidateTemplate(templatePath); err != nil {
		// Format the error message properly to avoid literal %s
		errorMessage := fmt.Sprintf(api.ErrTemplateValidationFailed, err.Error())
		return domain.TemplateProcessingError{
			Code:    domain.ErrCodeTemplateInvalid,
			Message: errorMessage,
			Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
				WithTemplatePath(templatePath).
				WithError(err),
		}
	}
	return nil
}

// validateVariables guards against invalid variables
func (g *RequestValidationGuard) validateVariables(variables *generation.TemplateVariables) error {
	if err := g.variableResolver.Validate(variables); err != nil {
		// Format the error message properly to avoid literal %s
		errorMessage := fmt.Sprintf(api.ErrVariableValidationFailed, err.Error())
		return domain.VariableResolutionError{
			Code:    domain.ErrCodeVariableInvalid,
			Message: errorMessage,
			Details: api.NewErrorDetails(api.ErrorCategoryVariable, api.ErrorSeverityHigh).
				WithError(err),
		}
	}
	return nil
}

// WatchModeGuard guards watch mode operations
type WatchModeGuard struct {
	watchManager generation.WatchModeManager
}

// NewWatchModeGuard creates a new watch mode guard
func NewWatchModeGuard(watchManager generation.WatchModeManager) *WatchModeGuard {
	return &WatchModeGuard{
		watchManager: watchManager,
	}
}

// CanStartWatchMode checks if watch mode can be started
func (g *WatchModeGuard) CanStartWatchMode() bool {
	return g.watchManager != nil
}

// CanStopWatchMode checks if watch mode can be stopped
func (g *WatchModeGuard) CanStopWatchMode() bool {
	return g.watchManager != nil
}

// ShouldStartWatchMode determines if watch mode should be started based on request and result
func (g *WatchModeGuard) ShouldStartWatchMode(req generation.PDFGenerationRequest, result generation.PDFGenerationResult) bool {
	return req.Options.WatchMode && result.Success
}

// GuardWatchModeUnavailable returns error when watch mode is unavailable
func (g *WatchModeGuard) GuardWatchModeUnavailable() error {
	return domain.PDFGenerationError{
		Code:    domain.ErrCodeWatchServiceUnavailable,
		Message: "Watch mode manager is not available",
	}
}

// PDFValidationGuard guards PDF validation operations
type PDFValidationGuard struct{}

// NewPDFValidationGuard creates a new PDF validation guard
func NewPDFValidationGuard() *PDFValidationGuard {
	return &PDFValidationGuard{}
}

// ShouldValidatePDF determines if PDF should be validated
func (g *PDFValidationGuard) ShouldValidatePDF(result generation.PDFGenerationResult) bool {
	return result.Success && len(result.PDFPath) > 0
}

// ContentPreviewGuard guards content preview operations
type ContentPreviewGuard struct{}

// NewContentPreviewGuard creates a new content preview guard
func NewContentPreviewGuard() *ContentPreviewGuard {
	return &ContentPreviewGuard{}
}

// GetPreviewLength calculates appropriate preview length
func (g *ContentPreviewGuard) GetPreviewLength(contentLength int) int {
	if contentLength > DefaultContentPreviewLength {
		return DefaultContentPreviewLength
	}
	return contentLength
}
