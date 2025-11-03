// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package watch

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/watch"
	"github.com/fsnotify/fsnotify"
)

// WatchApplicationService implements the WatchService interface
type WatchApplicationService struct {
	watcher          *fsnotify.Watcher
	config           watch.WatchConfiguration
	patternMatcher   watch.FilePatternMatcher
	debounceStrategy watch.DebounceStrategy
	changeProcessor  watch.FileChangeProcessor
	isWatching       bool
	logger           *logger.LoggerAdapter
}

// NewWatchApplicationService creates a new watch application service
func NewWatchApplicationService(
	patternMatcher watch.FilePatternMatcher,
	debounceStrategy watch.DebounceStrategy,
	changeProcessor watch.FileChangeProcessor,
	logger *logger.LoggerAdapter,
) *WatchApplicationService {
	return &WatchApplicationService{
		patternMatcher:   patternMatcher,
		debounceStrategy: debounceStrategy,
		changeProcessor:  changeProcessor,
		isWatching:       false,
		logger:           logger,
	}
}

// StartWatching begins the file watching process
func (w *WatchApplicationService) StartWatching(config watch.WatchConfiguration) error {
	if w.isWatching {
		return fmt.Errorf("already watching")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	w.watcher = watcher
	w.config = config
	w.isWatching = true

	w.logger.InfoWithFields("Starting file watcher",
		"template", config.TemplateFile,
		"config", config.ConfigFile,
		"interval", config.DebounceInterval)

	// Setup watcher directories
	if err := w.setupWatcher(); err != nil {
		w.StopWatching()
		return fmt.Errorf("failed to setup watcher: %w", err)
	}

	// Start watching loop
	go w.watchLoop()

	return nil
}

// StopWatching stops the file watching process
func (w *WatchApplicationService) StopWatching() error {
	if !w.isWatching {
		return nil
	}

	w.isWatching = false
	w.logger.InfoWithFields("Stopping file watcher")

	if w.watcher != nil {
		return w.watcher.Close()
	}
	return nil
}

// ConfigureExclusions updates exclusion patterns
func (w *WatchApplicationService) ConfigureExclusions(patterns []string) error {
	w.config.ExclusionPatterns = patterns
	w.logger.InfoWithFields("Updated exclusion patterns", "patterns", patterns)
	return nil
}

// ConfigureInterval updates the debounce interval
func (w *WatchApplicationService) ConfigureInterval(interval time.Duration) error {
	w.config.DebounceInterval = interval
	w.logger.InfoWithFields("Updated debounce interval", "interval", interval)
	return nil
}

// setupWatcher configures the file watcher
func (w *WatchApplicationService) setupWatcher() error {
	// Watch template directory
	templateDir := filepath.Dir(w.config.TemplateFile)
	if err := w.watcher.Add(templateDir); err != nil {
		return fmt.Errorf("failed to watch template directory: %w", err)
	}
	w.logger.InfoWithFields("Watching template directory", "directory", templateDir)

	// Watch config directory if different
	configDir := filepath.Dir(w.config.ConfigFile)
	if configDir != templateDir {
		if err := w.watcher.Add(configDir); err != nil {
			return fmt.Errorf("failed to watch config directory: %w", err)
		}
		w.logger.InfoWithFields("Watching config directory", "directory", configDir)
	}

	return nil
}

// watchLoop is the main watching loop
func (w *WatchApplicationService) watchLoop() {
	for w.isWatching {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleFileEvent(event)

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			// Log error but continue watching
			w.logger.ErrorWithFields("Watcher error", "error", err)
		}
	}
}

// handleFileEvent processes a file system event
func (w *WatchApplicationService) handleFileEvent(event fsnotify.Event) {
	// Normalize operation string to lowercase for comparison with domain constants
	// fsnotify returns "WRITE", "CREATE", etc. (uppercase), but domain uses lowercase
	operationStr := strings.ToLower(event.Op.String())

	changeEvent := watch.FileChangeEvent{
		FilePath:  event.Name,
		Operation: watch.FileOperation(operationStr),
		Timestamp: time.Now(),
	}

	// Following CLARITY: Info-level logging for visibility (not just Debug)
	// This helps diagnose when events are detected vs. filtered
	w.logger.InfoWithFields("File system event detected",
		"file", changeEvent.FilePath,
		"operation", changeEvent.Operation)

	// Check if we should process this event
	if !w.shouldProcessEvent(changeEvent) {
		w.logger.InfoWithFields("Event filtered out by pattern matcher",
			"file", changeEvent.FilePath,
			"excluded", w.patternMatcher.ShouldExclude(changeEvent.FilePath),
			"not_included", !w.patternMatcher.ShouldInclude(changeEvent.FilePath))
		return
	}

	// Apply debounce strategy
	if !w.debounceStrategy.ShouldTrigger(changeEvent) {
		w.logger.InfoWithFields("Event debounced",
			"file", changeEvent.FilePath)
		return
	}

	// Process the change
	if w.changeProcessor.CanProcess(changeEvent) {
		if err := w.changeProcessor.ProcessChange(changeEvent); err != nil {
			w.logger.ErrorWithFields("Failed to process file change",
				"file", changeEvent.FilePath,
				"error", err)
		} else {
			w.logger.InfoWithFields("File change processed successfully",
				"file", changeEvent.FilePath)
		}
	}
}

// shouldProcessEvent determines if an event should be processed
func (w *WatchApplicationService) shouldProcessEvent(event watch.FileChangeEvent) bool {
	// Check exclusion patterns
	if w.patternMatcher.ShouldExclude(event.FilePath) {
		return false
	}

	// Check inclusion patterns
	if !w.patternMatcher.ShouldInclude(event.FilePath) {
		return false
	}

	return true
}
