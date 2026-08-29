// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package watch_service

import (
	"context"
	"errors"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/domain/generation"
)

var ErrWatchModeUnavailable = errors.New("watch mode is unavailable: no watch manager configured")

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
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	return ErrWatchModeUnavailable
}

// StopWatchMode stops a specific watch mode
func (wmma *WatchModeManagerAdapter) StopWatchMode(watchID string) error {
	return ErrWatchModeUnavailable
}

// StopAllWatchModes stops all active watch modes
func (wmma *WatchModeManagerAdapter) StopAllWatchModes() error {
	return ErrWatchModeUnavailable
}

// GetActiveWatches returns information about active watch modes
func (wmma *WatchModeManagerAdapter) GetActiveWatches() map[string]generation.WatchInstanceInfo {
	return map[string]generation.WatchInstanceInfo{}
}
