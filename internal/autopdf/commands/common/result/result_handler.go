// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package result

import (
	"github.com/BuddhiLW/AutoPDF/internal/application"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/wiring"
)

// ResultHandler handles the output of build results
type ResultHandler struct{}

// NewResultHandler creates a new result handler
func NewResultHandler() *ResultHandler {
	return &ResultHandler{}
}

// HandleBuildResult processes and displays the build result
func (rh *ResultHandler) HandleBuildResult(result application.BuildResult) error {
	// Create logger for user feedback
	logger := adapters.NewLoggerAdapter(adapters.Detailed, "stdout")
	logger.InfoWithFields("Successfully built PDF", "pdf_path", result.PDFPath)

	if len(result.ImagePaths) > 0 {
		logger.Info("Generated image files:")
		for _, file := range result.ImagePaths {
			logger.InfoWithFields("  - Image file", "file", file)
		}
	}

	if result.Error != nil {
		logger.WarnWithFields("Warning", "error", result.Error)
	}

	return nil
}

// NewConvertResultHandler creates a new convert result handler
func NewConvertResultHandler() *ResultHandler {
	return &ResultHandler{}
}

// HandleConvertResult processes and displays the convert result
func (rh *ResultHandler) HandleConvertResult(imageFiles []string) error {
	// Create logger for user feedback
	logger := adapters.NewLoggerAdapter(adapters.Detailed, "stdout")
	if len(imageFiles) > 0 {
		logger.Info("Generated image files:")
		for _, file := range imageFiles {
			logger.InfoWithFields("  - Image file", "file", file)
		}
	} else {
		logger.Info("No image files were generated")
	}

	return nil
}

// HandleCleanResult processes and displays the clean result
func (rh *ResultHandler) HandleCleanResult(result *wiring.CleanResult) error {
	// Create logger for user feedback
	logger := adapters.NewLoggerAdapter(adapters.Detailed, "stdout")
	if result.FilesRemoved > 0 {
		logger.InfoWithFields("Successfully cleaned auxiliary files",
			"files_removed", result.FilesRemoved,
			"directory", result.Directory)

		if len(result.CleanedFiles) > 0 {
			logger.Info("Removed files:")
			for _, file := range result.CleanedFiles {
				logger.InfoWithFields("  - Removed file", "file", file)
			}
		}
	} else {
		logger.InfoWithFields("No auxiliary files found to clean", "directory", result.Directory)
	}

	if len(result.Errors) > 0 {
		logger.Warn("Errors encountered:")
		for _, err := range result.Errors {
			logger.WarnWithFields("  - Error", "error", err)
		}
	}

	return nil
}

// HandleVerboseResult processes and displays the verbose result
func (rh *ResultHandler) HandleVerboseResult(result *wiring.VerboseResult) error {
	// Create logger for user feedback
	logger := adapters.NewLoggerAdapter(adapters.Detailed, "stdout")
	if result.Enabled {
		logger.InfoWithFields("Verbose logging enabled",
			"level", result.Level,
			"description", result.Description)
	} else {
		logger.InfoWithFields("Verbose logging disabled", "level", result.Level)
	}
	return nil
}

// HandleDebugResult processes and displays the debug result
func (rh *ResultHandler) HandleDebugResult(result *wiring.DebugResult) error {
	// Create logger for user feedback
	logger := adapters.NewLoggerAdapter(adapters.Detailed, "stdout")
	if result.Enabled {
		logger.InfoWithFields("Debug output enabled", "output", result.Output)
	} else {
		logger.Info("Debug output disabled")
	}
	return nil
}

// HandleForceResult processes and displays the force result
func (rh *ResultHandler) HandleForceResult(result *wiring.ForceResult) error {
	// Create logger for user feedback
	logger := adapters.NewLoggerAdapter(adapters.Detailed, "stdout")
	if result.Enabled {
		logger.Info("Force operations enabled - files will be overwritten if they exist")
	} else {
		logger.Info("Force operations disabled - existing files will be protected")
	}
	return nil
}
