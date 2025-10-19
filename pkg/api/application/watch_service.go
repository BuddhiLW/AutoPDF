// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/watch"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
)

// WatchService encapsulates all watch-related functionality
// Implements generation.WatchService interface
type WatchService struct {
	watchService watch.WatchService
	watchManager generation.WatchModeManager
	watchGuard   *WatchModeGuard
}

// NewWatchService creates a new watch service
func NewWatchService(
	watchService watch.WatchService,
	watchManager generation.WatchModeManager,
) *WatchService {
	watchGuard := NewWatchModeGuard(watchManager)

	return &WatchService{
		watchService: watchService,
		watchManager: watchManager,
		watchGuard:   watchGuard,
	}
}

// StartWatchMode starts file watching for the given request
func (ws *WatchService) StartWatchMode(ctx context.Context, req generation.PDFGenerationRequest) error {
	if !ws.watchGuard.CanStartWatchMode() {
		return ws.watchGuard.GuardWatchModeUnavailable()
	}
	return ws.watchManager.StartWatchMode(ctx, req)
}

// StopWatchMode stops a specific watch mode
func (ws *WatchService) StopWatchMode(watchID string) error {
	if !ws.watchGuard.CanStopWatchMode() {
		return ws.watchGuard.GuardWatchModeUnavailable()
	}
	return ws.watchManager.StopWatchMode(watchID)
}

// StopAllWatchModes stops all active watch modes
func (ws *WatchService) StopAllWatchModes() error {
	if !ws.watchGuard.CanStopWatchMode() {
		return ws.watchGuard.GuardWatchModeUnavailable()
	}
	return ws.watchManager.StopAllWatchModes()
}

// GetActiveWatchModes returns information about active watch modes
func (ws *WatchService) GetActiveWatchModes() map[string]generation.WatchInstanceInfo {
	if !ws.watchGuard.CanStartWatchMode() {
		return make(map[string]generation.WatchInstanceInfo)
	}
	return ws.watchManager.GetActiveWatches()
}

// ShouldStartWatchMode determines if watch mode should be started
func (ws *WatchService) ShouldStartWatchMode(req generation.PDFGenerationRequest, result generation.PDFGenerationResult) bool {
	return ws.watchGuard.ShouldStartWatchMode(req, result)
}

// IsWatchModeAvailable checks if watch mode functionality is available
func (ws *WatchService) IsWatchModeAvailable() bool {
	return ws.watchGuard.CanStartWatchMode()
}
