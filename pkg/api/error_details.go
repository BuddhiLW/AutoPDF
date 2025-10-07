// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"time"
)

// ErrorDetails represents structured error information
// This replaces the generic map[string]interface{} with a proper entity
type ErrorDetails struct {
	// Core identification
	Category  string    `json:"category"`
	Severity  string    `json:"severity"`
	Timestamp time.Time `json:"timestamp"`

	// Context information
	Context map[string]string `json:"context,omitempty"`

	// File information
	FilePath   string `json:"file_path,omitempty"`
	FileSize   int64  `json:"file_size,omitempty"`
	FileExists bool   `json:"file_exists,omitempty"`

	// Validation information
	Validation *ValidationDetails `json:"validation,omitempty"`

	// System information
	System *SystemDetails `json:"system,omitempty"`

	// Recovery information
	Recovery *RecoveryDetails `json:"recovery,omitempty"`
}

// ValidationDetails contains validation-specific information
type ValidationDetails struct {
	Rule        string                 `json:"rule,omitempty"`
	Expected    interface{}            `json:"expected,omitempty"`
	Actual      interface{}            `json:"actual,omitempty"`
	Constraints map[string]interface{} `json:"constraints,omitempty"`
}

// SystemDetails contains system-specific information
type SystemDetails struct {
	OS           string `json:"os,omitempty"`
	Architecture string `json:"architecture,omitempty"`
	Memory       int64  `json:"memory,omitempty"`
	DiskSpace    int64  `json:"disk_space,omitempty"`
	ProcessID    int    `json:"process_id,omitempty"`
}

// RecoveryDetails contains recovery suggestions and information
type RecoveryDetails struct {
	Suggestions []string      `json:"suggestions,omitempty"`
	Retryable   bool          `json:"retryable"`
	MaxRetries  int           `json:"max_retries,omitempty"`
	Timeout     time.Duration `json:"timeout,omitempty"`
}

// NewErrorDetails creates a new ErrorDetails with default values
func NewErrorDetails(category, severity string) *ErrorDetails {
	return &ErrorDetails{
		Category:  category,
		Severity:  severity,
		Timestamp: time.Now(),
		Context:   make(map[string]string),
	}
}

// WithFilePath adds file path information to the error details
func (ed *ErrorDetails) WithFilePath(path string) *ErrorDetails {
	ed.FilePath = path
	ed.Context[ContextKeyPDFPath] = path
	return ed
}

// WithTemplatePath adds template path information to the error details
func (ed *ErrorDetails) WithTemplatePath(path string) *ErrorDetails {
	ed.Context[ContextKeyTemplatePath] = path
	return ed
}

// WithOutputPath adds output path information to the error details
func (ed *ErrorDetails) WithOutputPath(path string) *ErrorDetails {
	ed.Context[ContextKeyOutputPath] = path
	return ed
}

// WithEngine adds engine information to the error details
func (ed *ErrorDetails) WithEngine(engine string) *ErrorDetails {
	ed.Context[ContextKeyEngine] = engine
	return ed
}

// WithFormat adds format information to the error details
func (ed *ErrorDetails) WithFormat(format string) *ErrorDetails {
	ed.Context[ContextKeyFormat] = format
	return ed
}

// WithError adds error information to the error details
func (ed *ErrorDetails) WithError(err error) *ErrorDetails {
	if err != nil {
		ed.Context[ContextKeyError] = err.Error()
	}
	return ed
}

// WithFileSize adds file size information to the error details
func (ed *ErrorDetails) WithFileSize(size int64) *ErrorDetails {
	ed.FileSize = size
	ed.Context[ContextKeyFileSize] = string(rune(size))
	return ed
}

// WithPageCount adds page count information to the error details
func (ed *ErrorDetails) WithPageCount(count int) *ErrorDetails {
	ed.Context[ContextKeyPageCount] = string(rune(count))
	return ed
}

// WithValidation adds validation details to the error details
func (ed *ErrorDetails) WithValidation(rule string, expected, actual interface{}) *ErrorDetails {
	ed.Validation = &ValidationDetails{
		Rule:     rule,
		Expected: expected,
		Actual:   actual,
	}
	return ed
}

// WithSystem adds system details to the error details
func (ed *ErrorDetails) WithSystem(os, arch string, memory, diskSpace int64, processID int) *ErrorDetails {
	ed.System = &SystemDetails{
		OS:           os,
		Architecture: arch,
		Memory:       memory,
		DiskSpace:    diskSpace,
		ProcessID:    processID,
	}
	return ed
}

// WithRecovery adds recovery details to the error details
func (ed *ErrorDetails) WithRecovery(suggestions []string, retryable bool, maxRetries int, timeout time.Duration) *ErrorDetails {
	ed.Recovery = &RecoveryDetails{
		Suggestions: suggestions,
		Retryable:   retryable,
		MaxRetries:  maxRetries,
		Timeout:     timeout,
	}
	return ed
}

// AddContext adds a key-value pair to the context
func (ed *ErrorDetails) AddContext(key, value string) *ErrorDetails {
	ed.Context[key] = value
	return ed
}

// ToMap converts ErrorDetails to a map for backward compatibility
func (ed *ErrorDetails) ToMap() map[string]interface{} {
	result := make(map[string]interface{})

	// Add core fields
	result["category"] = ed.Category
	result["severity"] = ed.Severity
	result["timestamp"] = ed.Timestamp

	// Add context
	if len(ed.Context) > 0 {
		result["context"] = ed.Context
	}

	// Add file information
	if ed.FilePath != "" {
		result["file_path"] = ed.FilePath
	}
	if ed.FileSize > 0 {
		result["file_size"] = ed.FileSize
	}
	if ed.FileExists {
		result["file_exists"] = ed.FileExists
	}

	// Add validation information
	if ed.Validation != nil {
		result["validation"] = ed.Validation
	}

	// Add system information
	if ed.System != nil {
		result["system"] = ed.System
	}

	// Add recovery information
	if ed.Recovery != nil {
		result["recovery"] = ed.Recovery
	}

	return result
}
