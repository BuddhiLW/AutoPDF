// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerAdapter provides structured logging using zap
type LoggerAdapter struct {
	logger *zap.Logger
	level  zapcore.Level
}

// LogLevel represents the logging level
type LogLevel int

const (
	Silent LogLevel = iota
	Basic
	Detailed
	Debug
	Maximum
)

// String returns the string representation of LogLevel
func (l LogLevel) String() string {
	switch l {
	case Silent:
		return "Silent"
	case Basic:
		return "Basic"
	case Detailed:
		return "Detailed"
	case Debug:
		return "Debug"
	case Maximum:
		return "Maximum"
	default:
		return "Unknown"
	}
}

// NewLoggerAdapter creates a new logger adapter
func NewLoggerAdapter(level LogLevel, output string) *LoggerAdapter {
	config := zap.NewProductionConfig()

	// Set log level based on verbose level
	switch level {
	case Silent:
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case Basic:
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case Detailed:
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case Debug:
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case Maximum:
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}

	// Configure output based on debug setting
	switch output {
	case "stderr":
		config.OutputPaths = []string{"stderr"}
	case "file":
		logFile := filepath.Join(os.TempDir(), fmt.Sprintf("autopdf-%d.log", time.Now().Unix()))
		config.OutputPaths = []string{logFile}
	case "both":
		logFile := filepath.Join(os.TempDir(), fmt.Sprintf("autopdf-%d.log", time.Now().Unix()))
		config.OutputPaths = []string{"stdout", logFile}
	default: // stdout
		config.OutputPaths = []string{"stdout"}
	}

	// Add caller information for detailed debugging
	if level >= Debug {
		config.EncoderConfig.CallerKey = "caller"
		config.EncoderConfig.StacktraceKey = "stacktrace"
	}

	// Use console encoder for better readability
	if level >= Detailed {
		config.EncoderConfig = zap.NewDevelopmentEncoderConfig()
		config.Encoding = "console"
	}

	logger, err := config.Build()
	if err != nil {
		// Fallback to basic logger if configuration fails
		logger, _ = zap.NewProduction()
	}

	return &LoggerAdapter{
		logger: logger,
		level:  config.Level.Level(),
	}
}

// SetLevel updates the logging level
func (la *LoggerAdapter) SetLevel(level LogLevel) {
	la.logger = la.logger.WithOptions(zap.IncreaseLevel(zapcore.Level(level)))
}

// Info logs an info message
func (la *LoggerAdapter) Info(msg string, fields ...zap.Field) {
	la.logger.Info(msg, fields...)
}

// Debug logs a debug message
func (la *LoggerAdapter) Debug(msg string, fields ...zap.Field) {
	la.logger.Debug(msg, fields...)
}

// Warn logs a warning message
func (la *LoggerAdapter) Warn(msg string, fields ...zap.Field) {
	la.logger.Warn(msg, fields...)
}

// Error logs an error message
func (la *LoggerAdapter) Error(msg string, fields ...zap.Field) {
	la.logger.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func (la *LoggerAdapter) Fatal(msg string, fields ...zap.Field) {
	la.logger.Fatal(msg, fields...)
}

// InfoWithFields logs an info message with key-value pairs
func (la *LoggerAdapter) InfoWithFields(msg string, keyValues ...interface{}) {
	fields := make([]zap.Field, 0, len(keyValues)/2)
	for i := 0; i < len(keyValues); i += 2 {
		if i+1 < len(keyValues) {
			key, ok := keyValues[i].(string)
			if ok {
				fields = append(fields, zap.Any(key, keyValues[i+1]))
			}
		}
	}
	la.logger.Info(msg, fields...)
}

// DebugWithFields logs a debug message with key-value pairs
func (la *LoggerAdapter) DebugWithFields(msg string, keyValues ...interface{}) {
	fields := make([]zap.Field, 0, len(keyValues)/2)
	for i := 0; i < len(keyValues); i += 2 {
		if i+1 < len(keyValues) {
			key, ok := keyValues[i].(string)
			if ok {
				fields = append(fields, zap.Any(key, keyValues[i+1]))
			}
		}
	}
	la.logger.Debug(msg, fields...)
}

// ErrorWithFields logs an error message with key-value pairs
func (la *LoggerAdapter) ErrorWithFields(msg string, keyValues ...interface{}) {
	fields := make([]zap.Field, 0, len(keyValues)/2)
	for i := 0; i < len(keyValues); i += 2 {
		if i+1 < len(keyValues) {
			key, ok := keyValues[i].(string)
			if ok {
				fields = append(fields, zap.Any(key, keyValues[i+1]))
			}
		}
	}
	la.logger.Error(msg, fields...)
}

// WarnWithFields logs a warning message with key-value pairs
func (la *LoggerAdapter) WarnWithFields(msg string, keyValues ...interface{}) {
	fields := make([]zap.Field, 0, len(keyValues)/2)
	for i := 0; i < len(keyValues); i += 2 {
		if i+1 < len(keyValues) {
			key, ok := keyValues[i].(string)
			if ok {
				fields = append(fields, zap.Any(key, keyValues[i+1]))
			}
		}
	}
	la.logger.Warn(msg, fields...)
}

// Sync flushes any buffered log entries
func (la *LoggerAdapter) Sync() error {
	return la.logger.Sync()
}

// WithFields creates a new logger with additional fields
func (la *LoggerAdapter) WithFields(fields ...zap.Field) *LoggerAdapter {
	return &LoggerAdapter{
		logger: la.logger.With(fields...),
		level:  la.level,
	}
}

// AutoPDF Flow Logging Methods

// LogConfigBuilding logs configuration building process
func (la *LoggerAdapter) LogConfigBuilding(configPath string, variables map[string]string) {
	la.Info("Building configuration",
		zap.String("config_path", configPath),
		zap.Int("variable_count", len(variables)),
		zap.Strings("variables", mapKeysToStrings(variables)),
	)
}

// LogDataMapping logs data mapping process
func (la *LoggerAdapter) LogDataMapping(templatePath string, variables map[string]string) {
	la.Info("Mapping data to template",
		zap.String("template_path", templatePath),
		zap.Int("variable_count", len(variables)),
	)

	// Log each variable mapping in debug mode
	for key, value := range variables {
		la.Debug("Variable mapping",
			zap.String("key", key),
			zap.String("value", value),
		)
	}
}

// LogLaTeXCompilation logs LaTeX compilation process
func (la *LoggerAdapter) LogLaTeXCompilation(engine string, templatePath string, outputPath string) {
	la.Info("Starting LaTeX compilation",
		zap.String("engine", engine),
		zap.String("template", templatePath),
		zap.String("output", outputPath),
	)
}

// LogLaTeXError logs LaTeX compilation errors
func (la *LoggerAdapter) LogLaTeXError(err error, engine string, templatePath string) {
	la.Error("LaTeX compilation failed",
		zap.Error(err),
		zap.String("engine", engine),
		zap.String("template", templatePath),
	)
}

// LogLaTeXSuccess logs successful LaTeX compilation
func (la *LoggerAdapter) LogLaTeXSuccess(outputPath string, fileSize int64) {
	la.Info("LaTeX compilation successful",
		zap.String("output_path", outputPath),
		zap.Int64("file_size_bytes", fileSize),
	)
}

// LogPDFGeneration logs PDF generation process
func (la *LoggerAdapter) LogPDFGeneration(inputPath string, outputPath string) {
	la.Info("Generating PDF",
		zap.String("input_path", inputPath),
		zap.String("output_path", outputPath),
	)
}

// LogPDFConversion logs PDF to image conversion
func (la *LoggerAdapter) LogPDFConversion(pdfPath string, formats []string) {
	la.Info("Converting PDF to images",
		zap.String("pdf_path", pdfPath),
		zap.Strings("formats", formats),
	)
}

// LogConversionError logs conversion errors
func (la *LoggerAdapter) LogConversionError(err error, pdfPath string, format string) {
	la.Error("PDF conversion failed",
		zap.Error(err),
		zap.String("pdf_path", pdfPath),
		zap.String("format", format),
	)
}

// LogConversionSuccess logs successful conversion
func (la *LoggerAdapter) LogConversionSuccess(outputFiles []string) {
	la.Info("PDF conversion successful",
		zap.Strings("output_files", outputFiles),
	)
}

// LogFileCleanup logs file cleanup process
func (la *LoggerAdapter) LogFileCleanup(directory string, filesRemoved int) {
	la.Info("Cleaning auxiliary files",
		zap.String("directory", directory),
		zap.Int("files_removed", filesRemoved),
	)
}

// LogStep logs a general step in the AutoPDF flow
func (la *LoggerAdapter) LogStep(step string, details map[string]interface{}) {
	fields := make([]zap.Field, 0, len(details)+1)
	fields = append(fields, zap.String("step", step))

	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}

	la.Info("AutoPDF step", fields...)
}

// LogPerformance logs performance metrics
func (la *LoggerAdapter) LogPerformance(operation string, duration time.Duration, details map[string]interface{}) {
	fields := make([]zap.Field, 0, len(details)+2)
	fields = append(fields,
		zap.String("operation", operation),
		zap.Duration("duration", duration),
	)

	for key, value := range details {
		fields = append(fields, zap.Any(key, value))
	}

	la.Info("Performance metric", fields...)
}

// Helper function to convert map keys to strings
func mapKeysToStrings(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
