// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
)

func main() {
	fmt.Println("AutoPDF Zap Logging Demonstration")
	fmt.Println("=================================")

	// Demonstrate different logging levels
	levels := []logger.LogLevel{
		logger.Silent,
		logger.Basic,
		logger.Detailed,
		logger.Debug,
		logger.Maximum,
	}

	for _, level := range levels {
		fmt.Printf("\n--- %s Level Logging ---\n", level.String())

		logger := logger.NewLoggerAdapter(level, "stdout")
		defer logger.Sync()

		// Simulate AutoPDF flow with logging
		simulateAutoPDFFlow(logger)
	}
}

func simulateAutoPDFFlow(logger *logger.LoggerAdapter) {
	// Configuration building
	configPath := "/path/to/autopdf.yaml"
	variables := map[string]string{
		"title":   "My Document",
		"author":  "John Doe",
		"version": "1.0.0",
		"date":    "2025-01-07",
	}
	logger.LogConfigBuilding(configPath, variables)

	// Data mapping
	templatePath := "/path/to/template.tex"
	logger.LogDataMapping(templatePath, variables)

	// LaTeX compilation
	engine := "pdflatex"
	outputPath := "/path/to/output.pdf"
	logger.LogLaTeXCompilation(engine, templatePath, outputPath)

	// Simulate compilation success
	fileSize := int64(2048 * 1024) // 2MB
	logger.LogLaTeXSuccess(outputPath, fileSize)

	// PDF generation
	inputPath := "/path/to/input.tex"
	logger.LogPDFGeneration(inputPath, outputPath)

	// PDF conversion
	formats := []string{"png", "jpeg", "gif"}
	logger.LogPDFConversion(outputPath, formats)

	// Conversion success
	outputFiles := []string{
		"/path/to/output.png",
		"/path/to/output.jpeg",
		"/path/to/output.gif",
	}
	logger.LogConversionSuccess(outputFiles)

	// File cleanup
	directory := "/path/to/cleanup"
	filesRemoved := 8
	logger.LogFileCleanup(directory, filesRemoved)

	// Step logging
	step := "template_processing"
	details := map[string]interface{}{
		"template":  templatePath,
		"variables": len(variables),
		"engine":    engine,
		"output":    outputPath,
		"formats":   len(formats),
	}
	logger.LogStep(step, details)

	// Performance logging
	duration := 3 * time.Second
	perfDetails := map[string]interface{}{
		"operations":    15,
		"memory_usage":  "75MB",
		"cpu_usage":     "30%",
		"files_created": len(outputFiles),
	}
	logger.LogPerformance("pdf_generation", duration, perfDetails)

	// Error simulation (for demonstration)
	if logger != nil {
		logger.LogLaTeXError(fmt.Errorf("simulated LaTeX error"), engine, templatePath)
	}

	// Conversion error simulation
	logger.LogConversionError(fmt.Errorf("simulated conversion error"), outputPath, "png")
}
