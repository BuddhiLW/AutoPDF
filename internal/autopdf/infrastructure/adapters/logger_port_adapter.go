// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"context"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	application "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/ports"
	"go.uber.org/zap"
)

// LoggerPortAdapter adapts logger.LoggerAdapter to ports.Logger interface
// This follows DIP: infrastructure implements the application port
type LoggerPortAdapter struct {
	adapter *logger.LoggerAdapter
}

// NewLoggerPortAdapter creates a new adapter from LoggerAdapter
func NewLoggerPortAdapter(adapter *logger.LoggerAdapter) application.Logger {
	if adapter == nil {
		return &NoOpLogger{}
	}
	return &LoggerPortAdapter{adapter: adapter}
}

// Debug logs a debug-level message
func (l *LoggerPortAdapter) Debug(ctx context.Context, msg string, fields ...application.LogField) {
	zapFields := l.convertFields(fields...)
	l.adapter.Debug(msg, zapFields...)
}

// Info logs an info-level message
func (l *LoggerPortAdapter) Info(ctx context.Context, msg string, fields ...application.LogField) {
	zapFields := l.convertFields(fields...)
	l.adapter.Info(msg, zapFields...)
}

// Warn logs a warning-level message
func (l *LoggerPortAdapter) Warn(ctx context.Context, msg string, fields ...application.LogField) {
	zapFields := l.convertFields(fields...)
	l.adapter.Warn(msg, zapFields...)
}

// Error logs an error-level message
func (l *LoggerPortAdapter) Error(ctx context.Context, msg string, fields ...application.LogField) {
	zapFields := l.convertFields(fields...)
	l.adapter.Error(msg, zapFields...)
}

// convertFields converts ports.LogField to zap.Field
func (l *LoggerPortAdapter) convertFields(fields ...application.LogField) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		zapFields = append(zapFields, zap.Any(field.Key, field.Value))
	}
	return zapFields
}

// NoOpLogger is a no-op logger implementation for when no logger is provided
type NoOpLogger struct{}

// Debug does nothing
func (n *NoOpLogger) Debug(ctx context.Context, msg string, fields ...application.LogField) {}

// Info does nothing
func (n *NoOpLogger) Info(ctx context.Context, msg string, fields ...application.LogField) {}

// Warn does nothing
func (n *NoOpLogger) Warn(ctx context.Context, msg string, fields ...application.LogField) {}

// Error does nothing
func (n *NoOpLogger) Error(ctx context.Context, msg string, fields ...application.LogField) {}
