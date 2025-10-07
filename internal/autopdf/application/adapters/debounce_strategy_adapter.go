// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"sync"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/watch"
)

// DebounceStrategyAdapter implements the DebounceStrategy interface
type DebounceStrategyAdapter struct {
	interval    time.Duration
	lastTrigger time.Time
	timer       *time.Timer
	mutex       sync.RWMutex
}

// NewDebounceStrategyAdapter creates a new debounce strategy adapter
func NewDebounceStrategyAdapter(interval time.Duration) *DebounceStrategyAdapter {
	return &DebounceStrategyAdapter{
		interval: interval,
	}
}

// ShouldTrigger determines if an event should trigger based on debounce logic
func (d *DebounceStrategyAdapter) ShouldTrigger(event watch.FileChangeEvent) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Check if enough time has passed since last trigger
	if time.Since(d.lastTrigger) < d.interval {
		return false
	}

	// Update last trigger time
	d.lastTrigger = time.Now()
	return true
}

// Reset resets the debounce state
func (d *DebounceStrategyAdapter) Reset() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.lastTrigger = time.Time{}
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
}

// ConfigureInterval updates the debounce interval
func (d *DebounceStrategyAdapter) ConfigureInterval(interval time.Duration) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.interval = interval
}

// GetInterval returns the current debounce interval
func (d *DebounceStrategyAdapter) GetInterval() time.Duration {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return d.interval
}

// GetLastTrigger returns the time of the last trigger
func (d *DebounceStrategyAdapter) GetLastTrigger() time.Time {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return d.lastTrigger
}
