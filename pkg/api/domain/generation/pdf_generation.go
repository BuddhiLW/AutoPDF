// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package generation

import (
	"context"
	"time"
)

// PDFGenerationRequest represents a request to generate a PDF
type PDFGenerationRequest struct {
	TemplatePath string
	Variables    *TemplateVariables
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
	Verbose    int
	Debug      DebugOptions
	Force      bool
	RequestID  string // For unique file naming
	WatchMode  bool   // Enable file watching for automatic rebuilds
	WorkingDir string // Working directory for LaTeX compilation (isolates template builds)
}

// DebugOptions contains debug-specific settings
type DebugOptions struct {
	Enabled            bool
	LogToFile          bool
	LogFilePath        string
	CreateConcreteFile bool
	RequestID          string
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

// PDFMetadata contains metadata about a PDF file
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
