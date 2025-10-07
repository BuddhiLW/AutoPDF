// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package verbose

import (
	"testing"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerboseServiceCmd_LoggingLevels(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		expected string
	}{
		{
			name:     "Silent level",
			level:    0,
			expected: "Silent",
		},
		{
			name:     "Basic level",
			level:    1,
			expected: "Basic",
		},
		{
			name:     "Detailed level",
			level:    2,
			expected: "Detailed",
		},
		{
			name:     "Debug level",
			level:    3,
			expected: "Debug",
		},
		{
			name:     "Maximum level",
			level:    4,
			expected: "Maximum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test LogLevel string representation
			logLevel := logger.LogLevel(tt.level)
			assert.Equal(t, tt.expected, logLevel.String())
		})
	}
}

func TestLoggerAdapter_Creation(t *testing.T) {
	tests := []struct {
		name   string
		level  logger.LogLevel
		output string
	}{
		{
			name:   "Detailed level with stdout",
			level:  logger.Detailed,
			output: "stdout",
		},
		{
			name:   "Debug level with stderr",
			level:  logger.Debug,
			output: "stderr",
		},
		{
			name:   "Maximum level with file",
			level:  logger.Maximum,
			output: "file",
		},
		{
			name:   "Silent level with both",
			level:  logger.Silent,
			output: "both",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logger.NewLoggerAdapter(tt.level, tt.output)
			require.NotNil(t, logger)

			// Test that logger can be used without errors
			logger.Info("Test message")
			logger.Debug("Test debug message")
			logger.Warn("Test warning message")
			logger.Error("Test error message")

			// Test structured logging
			logger.InfoWithFields("Test structured message",
				"level", tt.level.String(),
				"output", tt.output,
			)

			// Test sync (may fail on stdout/stderr, which is expected)
			err := logger.Sync()
			// Sync errors on stdout/stderr are expected and can be ignored
			if err != nil && (tt.output == "stdout" || tt.output == "stderr" || tt.output == "both") {
				// This is expected for stdout/stderr
				t.Logf("Sync error (expected for %s): %v", tt.output, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoggerAdapter_AutoPDFFlowLogging(t *testing.T) {
	logger := logger.NewLoggerAdapter(logger.Detailed, "stdout")
	require.NotNil(t, logger)
	defer logger.Sync()

	// Test configuration building logging
	configPath := "/path/to/config.yaml"
	variables := map[string]string{
		"title":   "Test Document",
		"author":  "Test Author",
		"version": "1.0.0",
	}
	logger.LogConfigBuilding(configPath, variables)

	// Test data mapping logging
	templatePath := "/path/to/template.tex"
	logger.LogDataMapping(templatePath, variables)

	// Test LaTeX compilation logging
	engine := "pdflatex"
	outputPath := "/path/to/output.pdf"
	logger.LogLaTeXCompilation(engine, templatePath, outputPath)

	// Test LaTeX success logging
	fileSize := int64(1024 * 1024) // 1MB
	logger.LogLaTeXSuccess(outputPath, fileSize)

	// Test PDF generation logging
	inputPath := "/path/to/input.tex"
	logger.LogPDFGeneration(inputPath, outputPath)

	// Test PDF conversion logging
	formats := []string{"png", "jpeg", "gif"}
	logger.LogPDFConversion(outputPath, formats)

	// Test conversion success logging
	outputFiles := []string{
		"/path/to/output.png",
		"/path/to/output.jpeg",
		"/path/to/output.gif",
	}
	logger.LogConversionSuccess(outputFiles)

	// Test file cleanup logging
	directory := "/path/to/cleanup"
	filesRemoved := 5
	logger.LogFileCleanup(directory, filesRemoved)

	// Test step logging
	step := "template_processing"
	details := map[string]interface{}{
		"template":  templatePath,
		"variables": len(variables),
		"engine":    engine,
	}
	logger.LogStep(step, details)

	// Test performance logging
	duration := 2 * time.Second
	perfDetails := map[string]interface{}{
		"operations":   10,
		"memory_usage": "50MB",
		"cpu_usage":    "25%",
	}
	logger.LogPerformance("pdf_generation", duration, perfDetails)
}
