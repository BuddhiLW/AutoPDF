// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package watch_service

import (
	"context"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/watch"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
)

// WatchServiceAdapter implements generation.WatchService interface
type WatchServiceAdapter struct {
	watchService watch.WatchService
	watchManager generation.WatchModeManager
	logger       *logger.LoggerAdapter
}

// NewWatchServiceAdapter creates a new watch service adapter
func NewWatchServiceAdapter(
	watchService watch.WatchService,
	watchManager generation.WatchModeManager,
	logger *logger.LoggerAdapter,
) *WatchServiceAdapter {
	return &WatchServiceAdapter{
		watchService: watchService,
		watchManager: watchManager,
		logger:       logger,
	}
}

// StartWatchMode starts a watch mode for the given request
func (wsa *WatchServiceAdapter) StartWatchMode(ctx context.Context, req generation.PDFGenerationRequest) error {
	wsa.logger.DebugWithFields("Starting watch mode",
		"template_path", req.TemplatePath,
		"output_path", req.OutputPath,
	)

	return wsa.watchManager.StartWatchMode(ctx, req)
}

// StopWatchMode stops a specific watch mode
func (wsa *WatchServiceAdapter) StopWatchMode(watchID string) error {
	wsa.logger.DebugWithFields("Stopping watch mode",
		"watch_id", watchID,
	)

	return wsa.watchManager.StopWatchMode(watchID)
}

// StopAllWatchModes stops all active watch modes
func (wsa *WatchServiceAdapter) StopAllWatchModes() error {
	wsa.logger.DebugWithFields("Stopping all watch modes")

	return wsa.watchManager.StopAllWatchModes()
}

// GetActiveWatchModes returns information about active watch modes
func (wsa *WatchServiceAdapter) GetActiveWatchModes() map[string]generation.WatchInstanceInfo {
	activeWatches := wsa.watchManager.GetActiveWatches()

	wsa.logger.DebugWithFields("Retrieved active watch modes",
		"count", len(activeWatches),
	)

	return activeWatches
}

// ShouldStartWatchMode determines if watch mode should be started
func (wsa *WatchServiceAdapter) ShouldStartWatchMode(req generation.PDFGenerationRequest, result generation.PDFGenerationResult) bool {
	shouldStart := req.Options.WatchMode && result.Success

	wsa.logger.DebugWithFields("Evaluating watch mode start",
		"watch_mode_enabled", req.Options.WatchMode,
		"result_success", result.Success,
		"should_start", shouldStart,
	)

	return shouldStart
}

// IsWatchModeAvailable checks if watch mode is available
func (wsa *WatchServiceAdapter) IsWatchModeAvailable() bool {
	// For now, always return true - could be enhanced with actual availability checks
	return true
}

// ConfigureExclusions configures file exclusions for watch mode
func (wsa *WatchServiceAdapter) ConfigureExclusions(exclusions []string) error {
	wsa.logger.DebugWithFields("Configuring watch exclusions",
		"exclusions", exclusions,
	)

	return wsa.watchService.ConfigureExclusions(exclusions)
}

// ConfigureInterval configures the watch interval
func (wsa *WatchServiceAdapter) ConfigureInterval(interval time.Duration) error {
	wsa.logger.DebugWithFields("Configuring watch interval",
		"interval", interval,
	)

	return wsa.watchService.ConfigureInterval(interval)
}

// StartWatching starts the watch service
func (wsa *WatchServiceAdapter) StartWatching(config watch.WatchConfiguration) error {
	wsa.logger.DebugWithFields("Starting watch service")

	return wsa.watchService.StartWatching(config)
}

// StopWatching stops the watch service
func (wsa *WatchServiceAdapter) StopWatching() error {
	wsa.logger.DebugWithFields("Stopping watch service")

	return wsa.watchService.StopWatching()
}
