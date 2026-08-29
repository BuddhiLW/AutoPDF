// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package generation

import (
	"context"
	"time"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/model"
)

// PDFGenerationRequest represents a request to generate a PDF.
// Deprecated: use api.Request through api.Engine for new integrations.
type PDFGenerationRequest struct {
	TemplatePath string
	Variables    *TemplateVariables
	Engine       string
	OutputPath   string
	Options      PDFGenerationOptions
}

// PDFGenerationOptions contains optional settings for PDF generation.
// Deprecated: configure api.Request directly for new integrations.
type PDFGenerationOptions struct {
	DoClean    bool
	Conversion ConversionOptions
	Timeout    time.Duration
	Verbose    int
	Debug      DebugOptions
	Force      bool
	RequestID  string // For unique file naming
	WatchMode  bool   // Enable file watching for automatic rebuilds
	WorkingDir string // Working directory for LaTeX compilation (isolates template builds)
	Passes     int    // Number of compilation passes
	UseLatexmk bool   // Whether to use latexmk
}

// DebugOptions contains debug-specific settings
type DebugOptions struct {
	Enabled            bool
	LogToFile          bool
	LogFilePath        string
	CreateConcreteFile bool
	RequestID          string
}

// ConversionOptions contains settings for PDF to image conversion.
type ConversionOptions = model.ConversionOptions

// PDFGenerationResult represents the result of PDF generation.
// Deprecated: use api.Result through api.Engine for new integrations.
type PDFGenerationResult struct {
	PDFPath    string
	ImagePaths []string
	Success    bool
	Error      error
	Metadata   PDFMetadata
}

// PDFMetadata contains metadata about a PDF file.
type PDFMetadata = model.PDFMetadata

// PDFGenerationService defines the interface for PDF generation
type PDFGenerationService interface {
	Generate(ctx context.Context, req PDFGenerationRequest) (PDFGenerationResult, error)
	ValidateRequest(req PDFGenerationRequest) error
	GetSupportedEngines() []string
	GetSupportedFormats() []string
}

// TemplateProcessingService defines the interface for template processing
type TemplateProcessingService interface {
	Process(ctx context.Context, templatePath string, variables map[string]string) (string, error)
	ValidateTemplate(templatePath string) error
	GetTemplateVariables(templatePath string) ([]string, error)
}

// VariableResolver defines the interface for resolving complex variables
type VariableResolver interface {
	Resolve(variables *TemplateVariables) (map[string]string, error)
	Flatten(variables *TemplateVariables) map[string]string
	Validate(variables *TemplateVariables) error
}

// PDFValidator defines the interface for validating generated PDFs
type PDFValidator interface {
	Validate(pdfPath string) error
	GetMetadata(pdfPath string) (PDFMetadata, error)
	IsValidPDF(pdfPath string) bool
}

// WatchInstanceInfo provides information about a watch instance
type WatchInstanceInfo struct {
	ID           string        `json:"id"`
	TemplatePath string        `json:"template_path"`
	RequestID    string        `json:"request_id"`
	StartedAt    time.Time     `json:"started_at"`
	Duration     time.Duration `json:"duration"`
}

// WatchModeManager defines the interface for managing watch mode operations
type WatchModeManager interface {
	StartWatchMode(ctx context.Context, req PDFGenerationRequest) error
	StopWatchMode(watchID string) error
	StopAllWatchModes() error
	GetActiveWatches() map[string]WatchInstanceInfo
}

// WatchService defines the interface for watch-related operations
type WatchService interface {
	StartWatchMode(ctx context.Context, req PDFGenerationRequest) error
	StopWatchMode(watchID string) error
	StopAllWatchModes() error
	GetActiveWatchModes() map[string]WatchInstanceInfo
	ShouldStartWatchMode(req PDFGenerationRequest, result PDFGenerationResult) bool
	IsWatchModeAvailable() bool
}

// Error types and constants are now defined in domain/types.go
