// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/watch"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
)

// WatchModeManager manages watch mode instances for PDF generation
type WatchModeManager struct {
	activeWatches map[string]*WatchInstance
	mutex         sync.RWMutex
	logger        *logger.LoggerAdapter
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
func NewWatchModeManager(logger *logger.LoggerAdapter) *WatchModeManager {
	return &WatchModeManager{
		activeWatches: make(map[string]*WatchInstance),
		logger:        logger,
	}
}

// StartWatchMode starts watching for a PDF generation request
func (m *WatchModeManager) StartWatchMode(ctx context.Context, req generation.PDFGenerationRequest) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Create unique watch ID
	watchID := fmt.Sprintf("%s-%s-%d", req.Options.RequestID, req.TemplatePath, time.Now().UnixNano())

	// Check if already watching this template
	for _, instance := range m.activeWatches {
		if instance.TemplatePath == req.TemplatePath {
			m.logger.InfoWithFields("Template already being watched",
				"template_path", req.TemplatePath,
				"existing_watch_id", instance.ID,
			)
			return nil // Already watching, no need to start another
		}
	}

	// Create watch configuration
	watchConfig := watch.WatchConfiguration{
		TemplateFile:      req.TemplatePath,
		ConfigFile:        "autopdf.yaml", // Default config file
		DebounceInterval:  500 * time.Millisecond,
		ExclusionPatterns: []string{"*.aux", "*.log", "*.out", "*.toc", "*.fdb_latexmk", "*.fls", "*.synctex.gz"},
		InclusionPatterns: []string{"*.tex", "*.yaml", "*.yml"},
	}

	// Create context for this watch instance
	watchCtx, cancel := context.WithCancel(ctx)

	// Create watch instance
	instance := &WatchInstance{
		ID:           watchID,
		TemplatePath: req.TemplatePath,
		RequestID:    req.Options.RequestID,
		Config:       watchConfig,
		StartedAt:    time.Now(),
		Context:      watchCtx,
		Cancel:       cancel,
	}

	// Start watching in a goroutine
	go func() {
		m.logger.InfoWithFields("Starting watch mode",
			"watch_id", watchID,
			"template_path", req.TemplatePath,
			"request_id", req.Options.RequestID,
		)

		// TODO: Create actual watch service instance
		// For now, just simulate the watch service
		m.simulateWatchService(instance)

		// Clean up when done
		m.StopWatchMode(watchID)
	}()

	// Store the instance
	m.activeWatches[watchID] = instance

	m.logger.InfoWithFields("Watch mode started successfully",
		"watch_id", watchID,
		"template_path", req.TemplatePath,
		"active_watches", len(m.activeWatches),
	)

	return nil
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

	m.logger.InfoWithFields("Watch mode stopped",
		"watch_id", watchID,
		"template_path", instance.TemplatePath,
		"duration", time.Since(instance.StartedAt),
		"active_watches", len(m.activeWatches),
	)

	return nil
}

// StopAllWatchModes stops all active watch modes
func (m *WatchModeManager) StopAllWatchModes() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for watchID, instance := range m.activeWatches {
		instance.Cancel()
		m.logger.InfoWithFields("Stopped watch mode",
			"watch_id", watchID,
			"template_path", instance.TemplatePath,
		)
	}

	m.activeWatches = make(map[string]*WatchInstance)

	m.logger.InfoWithFields("All watch modes stopped",
		"total_stopped", len(m.activeWatches),
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

// simulateWatchService simulates a watch service for testing
func (m *WatchModeManager) simulateWatchService(instance *WatchInstance) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-instance.Context.Done():
			m.logger.InfoWithFields("Watch service context cancelled",
				"watch_id", instance.ID,
			)
			return
		case <-ticker.C:
			m.logger.DebugWithFields("Watch service heartbeat",
				"watch_id", instance.ID,
				"template_path", instance.TemplatePath,
			)
		}
	}
}
