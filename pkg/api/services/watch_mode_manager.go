// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	ports "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/watch"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
)

// ErrWatchModeUnavailable is returned until a real filesystem watcher is configured.
var ErrWatchModeUnavailable = errors.New("watch mode is unavailable: no watch service configured")

// WatchModeManager manages watch mode instances for PDF generation
type WatchModeManager struct {
	activeWatches map[string]*WatchInstance
	mutex         sync.RWMutex
	logger        ports.Logger
}

// WatchInstance represents an active watch mode instance
type WatchInstance struct {
	ID           string
	TemplatePath string
	RequestID    string
	WatchService watch.WatchService
	Config       watch.WatchConfiguration
	StartedAt    time.Time
	Context      context.Context
	Cancel       context.CancelFunc
}

// NewWatchModeManager creates a new watch mode manager
func NewWatchModeManager(logger ports.Logger) *WatchModeManager {
	if logger == nil {
		logger = ports.NewNoOpLogger()
	}
	return &WatchModeManager{
		activeWatches: make(map[string]*WatchInstance),
		logger:        logger,
	}
}

// StartWatchMode starts watching for a PDF generation request
func (m *WatchModeManager) StartWatchMode(ctx context.Context, _ generation.PDFGenerationRequest) error {
	if m == nil {
		return ErrWatchModeUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return ErrWatchModeUnavailable
}

// StopWatchMode stops watching for a specific watch ID
func (m *WatchModeManager) StopWatchMode(watchID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	instance, exists := m.activeWatches[watchID]
	if !exists {
		return fmt.Errorf("watch instance %s not found", watchID)
	}

	// Cancel the context
	instance.Cancel()

	// Remove from active watches
	delete(m.activeWatches, watchID)

	m.logger.Info(context.Background(), "Watch mode stopped",
		ports.NewLogField("watch_id", watchID),
		ports.NewLogField("template_path", instance.TemplatePath),
		ports.NewLogField("duration", time.Since(instance.StartedAt)),
		ports.NewLogField("active_watches", len(m.activeWatches)),
	)

	return nil
}

// StopAllWatchModes stops all active watch modes
func (m *WatchModeManager) StopAllWatchModes() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	stopped := len(m.activeWatches)

	for watchID, instance := range m.activeWatches {
		instance.Cancel()
		m.logger.Info(context.Background(), "Stopped watch mode",
			ports.NewLogField("watch_id", watchID),
			ports.NewLogField("template_path", instance.TemplatePath),
		)
	}

	m.activeWatches = make(map[string]*WatchInstance)

	m.logger.Info(context.Background(), "All watch modes stopped",
		ports.NewLogField("total_stopped", stopped),
	)

	return nil
}

// GetActiveWatches returns information about active watch modes
func (m *WatchModeManager) GetActiveWatches() map[string]WatchInstanceInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	watches := make(map[string]WatchInstanceInfo)
	for watchID, instance := range m.activeWatches {
		watches[watchID] = WatchInstanceInfo{
			ID:           instance.ID,
			TemplatePath: instance.TemplatePath,
			RequestID:    instance.RequestID,
			StartedAt:    instance.StartedAt,
			Duration:     time.Since(instance.StartedAt),
		}
	}

	return watches
}

// WatchInstanceInfo provides information about a watch instance
type WatchInstanceInfo struct {
	ID           string        `json:"id"`
	TemplatePath string        `json:"template_path"`
	RequestID    string        `json:"request_id"`
	StartedAt    time.Time     `json:"started_at"`
	Duration     time.Duration `json:"duration"`
}
