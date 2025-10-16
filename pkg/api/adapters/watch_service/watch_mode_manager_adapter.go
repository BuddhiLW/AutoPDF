// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package watch_service

import (
	"context"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
)

// WatchModeManagerAdapter implements generation.WatchModeManager interface
type WatchModeManagerAdapter struct {
	logger *logger.LoggerAdapter
}

// NewWatchModeManagerAdapter creates a new watch mode manager adapter
func NewWatchModeManagerAdapter(logger *logger.LoggerAdapter) *WatchModeManagerAdapter {
	return &WatchModeManagerAdapter{
		logger: logger,
	}
}

// StartWatchMode starts a watch mode for the given request
func (wmma *WatchModeManagerAdapter) StartWatchMode(ctx context.Context, req generation.PDFGenerationRequest) error {
	wmma.logger.DebugWithFields("Starting watch mode via manager",
		"template_path", req.TemplatePath,
		"output_path", req.OutputPath,
	)

	// For now, just log the request - could be enhanced with actual watch logic
	wmma.logger.InfoWithFields("Watch mode started",
		"template_path", req.TemplatePath,
		"output_path", req.OutputPath,
		"engine", req.Engine,
		"watch_mode", req.Options.WatchMode,
		"debug", req.Options.Debug.Enabled,
	)

	return nil
}

// StopWatchMode stops a specific watch mode
func (wmma *WatchModeManagerAdapter) StopWatchMode(watchID string) error {
	wmma.logger.DebugWithFields("Stopping watch mode via manager",
		"watch_id", watchID,
	)

	wmma.logger.InfoWithFields("Watch mode stopped", "watch_id", watchID)
	return nil
}

// StopAllWatchModes stops all active watch modes
func (wmma *WatchModeManagerAdapter) StopAllWatchModes() error {
	wmma.logger.DebugWithFields("Stopping all watch modes via manager")

	wmma.logger.InfoWithFields("All watch modes stopped")
	return nil
}

// GetActiveWatches returns information about active watch modes
func (wmma *WatchModeManagerAdapter) GetActiveWatches() map[string]generation.WatchInstanceInfo {
	wmma.logger.DebugWithFields("Retrieving active watches via manager")

	// For now, return empty map - could be enhanced with actual watch tracking
	activeWatches := make(map[string]generation.WatchInstanceInfo)

	wmma.logger.DebugWithFields("Retrieved active watches via manager",
		"count", len(activeWatches),
	)

	return activeWatches
}
