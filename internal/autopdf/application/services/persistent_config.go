// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters"
)

// PersistentConfig represents the persistent configuration for AutoPDF CLI
type PersistentConfig struct {
	// Verbose settings
	VerboseLevel   adapters.LogLevel `yaml:"verbose_level" json:"verbose_level"`
	VerboseEnabled bool              `yaml:"verbose_enabled" json:"verbose_enabled"`

	// Clean settings
	CleanEnabled bool `yaml:"clean_enabled" json:"clean_enabled"`

	// Debug settings
	DebugEnabled bool   `yaml:"debug_enabled" json:"debug_enabled"`
	DebugOutput  string `yaml:"debug_output" json:"debug_output"`

	// Force settings
	ForceEnabled bool `yaml:"force_enabled" json:"force_enabled"`

	// Metadata
	LastUpdated time.Time `yaml:"last_updated" json:"last_updated"`
	Version     string    `yaml:"version" json:"version"`
}

// DefaultPersistentConfig returns the default persistent configuration
func DefaultPersistentConfig() *PersistentConfig {
	return &PersistentConfig{
		VerboseLevel:   adapters.Detailed, // Default to detailed logging
		VerboseEnabled: true,
		CleanEnabled:   false, // Default to not cleaning
		DebugEnabled:   false,
		DebugOutput:    "stdout",
		ForceEnabled:   false,
		LastUpdated:    time.Now(),
		Version:        "1.0.0",
	}
}

// SetVerboseLevel sets the verbose level and marks as enabled
func (pc *PersistentConfig) SetVerboseLevel(level adapters.LogLevel) {
	pc.VerboseLevel = level
	pc.VerboseEnabled = level > adapters.Silent
	pc.LastUpdated = time.Now()
}

// SetCleanEnabled sets the clean setting
func (pc *PersistentConfig) SetCleanEnabled(enabled bool) {
	pc.CleanEnabled = enabled
	pc.LastUpdated = time.Now()
}

// SetDebugEnabled sets the debug setting
func (pc *PersistentConfig) SetDebugEnabled(enabled bool, output string) {
	pc.DebugEnabled = enabled
	if output != "" {
		pc.DebugOutput = output
	}
	pc.LastUpdated = time.Now()
}

// SetForceEnabled sets the force setting
func (pc *PersistentConfig) SetForceEnabled(enabled bool) {
	pc.ForceEnabled = enabled
	pc.LastUpdated = time.Now()
}

// ToggleClean toggles the clean setting
func (pc *PersistentConfig) ToggleClean() bool {
	pc.CleanEnabled = !pc.CleanEnabled
	pc.LastUpdated = time.Now()
	return pc.CleanEnabled
}

// ToggleDebug toggles the debug setting
func (pc *PersistentConfig) ToggleDebug() bool {
	pc.DebugEnabled = !pc.DebugEnabled
	pc.LastUpdated = time.Now()
	return pc.DebugEnabled
}

// ToggleForce toggles the force setting
func (pc *PersistentConfig) ToggleForce() bool {
	pc.ForceEnabled = !pc.ForceEnabled
	pc.LastUpdated = time.Now()
	return pc.ForceEnabled
}

// GetVerboseDescription returns a human-readable description of the verbose level
func (pc *PersistentConfig) GetVerboseDescription() string {
	descriptions := map[adapters.LogLevel]string{
		adapters.Silent:   "Silent (only errors)",
		adapters.Basic:    "Basic information (warnings and above)",
		adapters.Detailed: "Detailed information (info and above)",
		adapters.Debug:    "Debug information (debug and above)",
		adapters.Maximum:  "Maximum verbosity (all logs with full introspection)",
	}

	if desc, exists := descriptions[pc.VerboseLevel]; exists {
		return desc
	}
	return "Unknown level"
}

// GetStatus returns a status summary of all settings
func (pc *PersistentConfig) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"verbose": map[string]interface{}{
			"enabled":     pc.VerboseEnabled,
			"level":       pc.VerboseLevel.String(),
			"description": pc.GetVerboseDescription(),
		},
		"clean": map[string]interface{}{
			"enabled": pc.CleanEnabled,
		},
		"debug": map[string]interface{}{
			"enabled": pc.DebugEnabled,
			"output":  pc.DebugOutput,
		},
		"force": map[string]interface{}{
			"enabled": pc.ForceEnabled,
		},
		"metadata": map[string]interface{}{
			"last_updated": pc.LastUpdated.Format(time.RFC3339),
			"version":      pc.Version,
		},
	}
}
