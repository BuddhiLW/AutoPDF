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

// NoopLogger is the default public logger and deliberately emits nothing.
type NoopLogger struct{}

// NewNoopLogger returns a Logger suitable for libraries that do not need logging.
func NewNoopLogger() Logger { return NoopLogger{} }

func (NoopLogger) Debug(context.Context, string, ...LogField) {}
func (NoopLogger) Info(context.Context, string, ...LogField)  {}
func (NoopLogger) Warn(context.Context, string, ...LogField)  {}
func (NoopLogger) Error(context.Context, string, ...LogField) {}

// LogField represents a key-value pair for structured logging
type LogField struct {
	Key   string
	Value interface{}
}

// NewLogField creates a new log field
func NewLogField(key string, value interface{}) LogField {
	return LogField{Key: key, Value: value}
}

type internalLoggerAdapter struct {
	logger Logger
}

func newInternalLoggerAdapter(logger Logger) autopdfports.Logger {
	if logger == nil {
		return &noopInternalLoggerAdapter{}
	}
	return &internalLoggerAdapter{logger: logger}
}

// Debug logs a debug-level message
func (a *internalLoggerAdapter) Debug(ctx context.Context, msg string, fields ...autopdfports.LogField) {
	if a.logger != nil {
		pubFields := make([]LogField, len(fields))
		for i, f := range fields {
			pubFields[i] = LogField{Key: f.Key, Value: f.Value}
		}
		a.logger.Debug(ctx, msg, pubFields...)
	}
}

// Info logs an info-level message
func (a *internalLoggerAdapter) Info(ctx context.Context, msg string, fields ...autopdfports.LogField) {
	if a.logger != nil {
		pubFields := make([]LogField, len(fields))
		for i, f := range fields {
			pubFields[i] = LogField{Key: f.Key, Value: f.Value}
		}
		a.logger.Info(ctx, msg, pubFields...)
	}
}

// Warn logs a warning-level message
func (a *internalLoggerAdapter) Warn(ctx context.Context, msg string, fields ...autopdfports.LogField) {
	if a.logger != nil {
		pubFields := make([]LogField, len(fields))
		for i, f := range fields {
			pubFields[i] = LogField{Key: f.Key, Value: f.Value}
		}
		a.logger.Warn(ctx, msg, pubFields...)
	}
}

// Error logs an error-level message
func (a *internalLoggerAdapter) Error(ctx context.Context, msg string, fields ...autopdfports.LogField) {
	if a.logger != nil {
		pubFields := make([]LogField, len(fields))
		for i, f := range fields {
			pubFields[i] = LogField{Key: f.Key, Value: f.Value}
		}
		a.logger.Error(ctx, msg, pubFields...)
	}
}

type noopInternalLoggerAdapter struct{}

// Debug does nothing
func (n *noopInternalLoggerAdapter) Debug(ctx context.Context, msg string, fields ...autopdfports.LogField) {
}

// Info does nothing
func (n *noopInternalLoggerAdapter) Info(ctx context.Context, msg string, fields ...autopdfports.LogField) {
}

// Warn does nothing
func (n *noopInternalLoggerAdapter) Warn(ctx context.Context, msg string, fields ...autopdfports.LogField) {
}

// Error does nothing
func (n *noopInternalLoggerAdapter) Error(ctx context.Context, msg string, fields ...autopdfports.LogField) {
}
