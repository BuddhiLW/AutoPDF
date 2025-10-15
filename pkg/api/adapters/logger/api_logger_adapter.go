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
func NewAPILoggerFactory(debugEnabled bool, logDir string) *APILoggerFactory {
	// Create base logger with appropriate level
	level := logger.Detailed
	if debugEnabled {
		level = logger.Debug
	}

	baseLogger := logger.NewLoggerAdapter(level, "stdout")

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

	// Create a new logger for this request
	requestLogger := logger.NewLoggerAdapter(logger.Debug, "stdout")

	// Add request ID to the logger context
	requestLogger = requestLogger.WithFields(zap.String("request_id", requestID))

	// Log the creation of the request logger
	requestLogger.InfoWithFields("Created request-scoped logger",
		"request_id", requestID,
		"log_file", logFile,
	)

	return requestLogger
}

// GetBaseLogger returns the base logger for non-request-scoped operations
func (f *APILoggerFactory) GetBaseLogger() *logger.LoggerAdapter {
	return f.baseLogger
}
