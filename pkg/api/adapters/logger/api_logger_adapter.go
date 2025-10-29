package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"go.uber.org/zap"
)

// APILoggerFactory creates request-scoped loggers for API operations
type APILoggerFactory struct {
	baseLogger *logger.LoggerAdapter
	logDir     string
}

// NewAPILoggerFactory creates a new API logger factory
// If debugEnabled is true, it creates a file-based logger in logDir
// Otherwise, it uses stdout-only logging
func NewAPILoggerFactory(debugEnabled bool, logDir string) *APILoggerFactory {
	// Create base logger with appropriate level
	level := logger.Detailed
	if debugEnabled {
		level = logger.Debug
	}

	var baseLogger *logger.LoggerAdapter

	if debugEnabled && logDir != "" {
		// Ensure log directory exists
		if err := os.MkdirAll(logDir, 0755); err == nil {
			// Create a generic log file for base logger (template-specific logs will use CreateRequestLogger)
			logFile := filepath.Join(logDir, "autopdf.log")
			// Use "both" mode: stdout for visibility + file for persistence
			// Pass custom file path to NewLoggerAdapter
			baseLogger = logger.NewLoggerAdapter(level, logFile)
			// Also add stdout for real-time visibility during development
			// Note: NewLoggerAdapter with custom path only writes to that file
			// We'll need to create a custom config for "both" mode with custom path
			// For now, use file-only and rely on application logs for stdout visibility
		} else {
			// Fallback to stdout if directory creation fails
			baseLogger = logger.NewLoggerAdapter(level, "stdout")
		}
	} else {
		baseLogger = logger.NewLoggerAdapter(level, "stdout")
	}

	return &APILoggerFactory{
		baseLogger: baseLogger,
		logDir:     logDir,
	}
}

// CreateRequestLogger creates a logger scoped to a single API request
func (f *APILoggerFactory) CreateRequestLogger(requestID string, debugEnabled bool) *logger.LoggerAdapter {
	if !debugEnabled {
		return f.baseLogger
	}

	// Ensure log directory exists
	if err := os.MkdirAll(f.logDir, 0755); err != nil {
		// Fallback to base logger if we can't create directory
		f.baseLogger.WarnWithFields("Failed to create log directory, using base logger",
			"log_dir", f.logDir,
			"error", err,
		)
		return f.baseLogger
	}

	// Create request-specific log file
	logFile := filepath.Join(f.logDir, fmt.Sprintf("autopdf-api-%s.log", requestID))

	// Create a new logger for this request with custom file path
	requestLogger := logger.NewLoggerAdapter(logger.Debug, logFile)

	// Add request ID to the logger context
	requestLogger = requestLogger.WithFields(zap.String("request_id", requestID))

	// Log the creation of the request logger
	requestLogger.InfoWithFields("Created request-scoped logger",
		"request_id", requestID,
		"log_file", logFile,
	)

	return requestLogger
}

// GetTemplateLogger creates a logger for a specific template
// This is useful for generating template-specific log files
func (f *APILoggerFactory) GetTemplateLogger(templateID string, debugEnabled bool) *logger.LoggerAdapter {
	if !debugEnabled || f.logDir == "" {
		return f.baseLogger
	}

	// Ensure log directory exists
	if err := os.MkdirAll(f.logDir, 0755); err != nil {
		// Fallback to base logger if we can't create directory
		f.baseLogger.WarnWithFields("Failed to create log directory, using base logger",
			"log_dir", f.logDir,
			"error", err,
		)
		return f.baseLogger
	}

	// Create template-specific log file
	logFile := filepath.Join(f.logDir, fmt.Sprintf("autopdf-%s.log", templateID))

	// Create a new logger for this template with Debug level and custom file path
	templateLogger := logger.NewLoggerAdapter(logger.Debug, logFile)

	// Add template ID to the logger context
	templateLogger = templateLogger.WithFields(zap.String("template_id", templateID))

	return templateLogger
}

// GetBaseLogger returns the base logger for non-request-scoped operations
func (f *APILoggerFactory) GetBaseLogger() *logger.LoggerAdapter {
	return f.baseLogger
}
