// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package latex

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	apiconfig "github.com/BuddhiLW/AutoPDF/pkg/api/config"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// LaTeXCompilerAdapter wraps the existing LaTeX compiler
type LaTeXCompilerAdapter struct {
	config *config.Config
}

// NewLaTeXCompilerAdapter creates a new LaTeX compiler adapter
func NewLaTeXCompilerAdapter(cfg *config.Config) *LaTeXCompilerAdapter {
	return &LaTeXCompilerAdapter{
		config: cfg,
	}
}

// Compile compiles LaTeX content to PDF
func (lca *LaTeXCompilerAdapter) Compile(ctx context.Context, content string, engine string, outputPath string, debugEnabled bool) (string, error) {
	if content == "" {
		return "", errors.New("no LaTeX content provided")
	}

	// Determine the engine to use
	if engine == "" {
		engine = "pdflatex" // Default engine
	}

	// Verify that the engine is installed
	if _, err := exec.LookPath(engine); err != nil {
		return "", fmt.Errorf("LaTeX engine not found: %s", engine)
	}

	// Load debug configuration from environment
	debugConfig := apiconfig.LoadDebugConfigFromEnv()

	// Create a temporary file for the LaTeX content
	tempDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Generate a concrete file name
	var concreteFileName string
	if debugEnabled {
		// In debug mode, create a persistent concrete file with a memorable name
		concreteFileName = "autopdf-concrete.tex"
		if outputPath != "" {
			baseName := filepath.Base(outputPath)
			concreteFileName = "autopdf-concrete-" + strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".tex"
		}
	} else {
		// Normal mode: use temporary file
		concreteFileName = "autopdf_temp.tex"
		if outputPath != "" {
			baseName := filepath.Base(outputPath)
			concreteFileName = "autopdf_" + strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".tex"
		}
	}
	concreteFile := filepath.Join(tempDir, concreteFileName)

	// Write the content to the concrete file
	if err := os.WriteFile(concreteFile, []byte(content), 0644); err != nil {
		return "", err
	}

	// Only clean up temp file if not in debug mode
	if !debugEnabled {
		defer os.Remove(concreteFile)
	}

	// Determine output PDF path
	var pdfPath string
	if outputPath != "" {
		pdfPath = outputPath
		if !strings.HasSuffix(pdfPath, ".pdf") {
			pdfPath = pdfPath + ".pdf"
		}
	} else {
		// Default output path
		baseName := strings.TrimSuffix(filepath.Base(concreteFile), ".tex")
		pdfPath = filepath.Join(tempDir, baseName+".pdf")
	}

	// Create output directory if needed
	outputDir := filepath.Dir(pdfPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get the base name for the LaTeX job
	baseName := strings.TrimSuffix(filepath.Base(pdfPath), ".pdf")

	// Extract template directory from first line comment if present
	// Template path should be passed via environment variable or inferred from output path
	// For now, we'll try to get it from the AUTOPDF_TEMPLATE_PATH environment variable
	templateDir := os.Getenv("AUTOPDF_TEMPLATE_DIR")

	// Create command to run LaTeX
	var cmd *exec.Cmd
	if outputDir == "." {
		cmdStr := fmt.Sprintf("%s -interaction=nonstopmode -jobname=%s %s", engine, baseName, concreteFile)
		cmd = exec.Command("sh", "-c", cmdStr)
	} else {
		// Make output directory absolute to work with working directory change
		absOutputDir, err := filepath.Abs(outputDir)
		if err == nil {
			outputDir = absOutputDir
		}
		cmdStr := fmt.Sprintf("%s -interaction=nonstopmode -jobname=%s -output-directory=%s %s", engine, baseName, outputDir, concreteFile)
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	// Set working directory to template directory if specified
	// This allows templates to reference assets with relative paths like ./assets/logo.png
	if templateDir != "" {
		cmd.Dir = templateDir
	}

	// Run the LaTeX command and capture output for debug logging
	var cmdOutput []byte
	var cmdErr error
	if debugConfig.Enabled {
		// Capture both stdout and stderr for debug logging
		cmdOutput, cmdErr = cmd.CombinedOutput()
	} else {
		// Normal execution without capturing output
		cmdErr = cmd.Run()
	}

	if cmdErr != nil {
		// Check if PDF was created despite the error
		if _, statErr := os.Stat(pdfPath); statErr != nil {
			return "", fmt.Errorf("LaTeX compilation failed: %w", cmdErr)
		}
		// PDF was created, so continue
	}

	// Create debug log file if enabled
	if debugConfig.Enabled {
		logDir := debugConfig.GetLogDirectory()
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Printf("Warning: Failed to create log directory %s: %v\n", logDir, err)
		} else {
			// Create timestamped log file
			timestamp := time.Now().Format("20060102-150405")
			logFile := filepath.Join(logDir, fmt.Sprintf("request-%s.log", timestamp))

			// Prepare log content
			logContent := fmt.Sprintf("=== AutoPDF LaTeX Compilation Log ===\n")
			logContent += fmt.Sprintf("Timestamp: %s\n", time.Now().Format("2006-01-02 15:04:05"))
			logContent += fmt.Sprintf("Engine: %s\n", engine)
			logContent += fmt.Sprintf("Input File: %s\n", concreteFile)
			logContent += fmt.Sprintf("Output PDF: %s\n", pdfPath)
			logContent += fmt.Sprintf("Command: %s\n", cmd.String())
			logContent += fmt.Sprintf("Exit Code: %v\n", cmdErr)
			logContent += fmt.Sprintf("\n=== LaTeX Output ===\n")
			if len(cmdOutput) > 0 {
				logContent += string(cmdOutput)
			} else {
				logContent += "No output captured\n"
			}

			if err := os.WriteFile(logFile, []byte(logContent), 0644); err != nil {
				fmt.Printf("Warning: Failed to create debug log file %s: %v\n", logFile, err)
			} else {
				fmt.Printf("Debug: Created log file at %s\n", logFile)
			}
		}
	}

	// Check if output PDF exists and has content
	fileInfo, statErr := os.Stat(pdfPath)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return "", errors.New("PDF output file was not created")
		}
		return "", fmt.Errorf("error checking output file: %w", statErr)
	}

	// Check if file is empty
	if fileInfo.Size() == 0 {
		return "", errors.New("PDF output file was created but is empty")
	}

	return pdfPath, nil
}
