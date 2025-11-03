// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"

	autopdfports "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/ports"
)

// Logger is a public interface for logging within AutoPDF
// This allows external consumers (like cartas-backend) to provide their own loggers
// and see logs from AutoPDF operations (like latexmk commands) in their own logger output
type Logger interface {
	// Debug logs a debug-level message
	Debug(ctx context.Context, msg string, fields ...LogField)
	// Info logs an info-level message
	Info(ctx context.Context, msg string, fields ...LogField)
	// Warn logs a warning-level message
	Warn(ctx context.Context, msg string, fields ...LogField)
	// Error logs an error-level message
	Error(ctx context.Context, msg string, fields ...LogField)
}

// LogField represents a key-value pair for structured logging
type LogField struct {
	Key   string
	Value interface{}
}

// NewLogField creates a new log field
func NewLogField(key string, value interface{}) LogField {
	return LogField{Key: key, Value: value}
}

// LoggerAdapter adapts AutoPDF's public Logger interface to internal ports.Logger
// This follows Adapter pattern - bridges public API to internal ports
type LoggerAdapter struct {
	logger Logger
}

// NewLoggerAdapter creates a new adapter from a public Logger
func NewLoggerAdapter(logger Logger) autopdfports.Logger {
	if logger == nil {
		return &NoOpLoggerAdapter{}
	}
	return &LoggerAdapter{logger: logger}
}

// Debug logs a debug-level message
func (a *LoggerAdapter) Debug(ctx context.Context, msg string, fields ...autopdfports.LogField) {
	if a.logger != nil {
		pubFields := make([]LogField, len(fields))
		for i, f := range fields {
			pubFields[i] = LogField{Key: f.Key, Value: f.Value}
		}
		a.logger.Debug(ctx, msg, pubFields...)
	}
}

// Info logs an info-level message
func (a *LoggerAdapter) Info(ctx context.Context, msg string, fields ...autopdfports.LogField) {
	if a.logger != nil {
		pubFields := make([]LogField, len(fields))
		for i, f := range fields {
			pubFields[i] = LogField{Key: f.Key, Value: f.Value}
		}
		a.logger.Info(ctx, msg, pubFields...)
	}
}

// Warn logs a warning-level message
func (a *LoggerAdapter) Warn(ctx context.Context, msg string, fields ...autopdfports.LogField) {
	if a.logger != nil {
		pubFields := make([]LogField, len(fields))
		for i, f := range fields {
			pubFields[i] = LogField{Key: f.Key, Value: f.Value}
		}
		a.logger.Warn(ctx, msg, pubFields...)
	}
}

// Error logs an error-level message
func (a *LoggerAdapter) Error(ctx context.Context, msg string, fields ...autopdfports.LogField) {
	if a.logger != nil {
		pubFields := make([]LogField, len(fields))
		for i, f := range fields {
			pubFields[i] = LogField{Key: f.Key, Value: f.Value}
		}
		a.logger.Error(ctx, msg, pubFields...)
	}
}

// NoOpLoggerAdapter is a no-op logger implementation for when no logger is provided
type NoOpLoggerAdapter struct{}

// Debug does nothing
func (n *NoOpLoggerAdapter) Debug(ctx context.Context, msg string, fields ...autopdfports.LogField) {}

// Info does nothing
func (n *NoOpLoggerAdapter) Info(ctx context.Context, msg string, fields ...autopdfports.LogField) {}

// Warn does nothing
func (n *NoOpLoggerAdapter) Warn(ctx context.Context, msg string, fields ...autopdfports.LogField) {}

// Error does nothing
func (n *NoOpLoggerAdapter) Error(ctx context.Context, msg string, fields ...autopdfports.LogField) {}
