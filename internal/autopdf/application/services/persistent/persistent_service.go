// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package persistent

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/BuddhiLW/AutoPDF/configs"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"github.com/rwxrob/bonzai/persisters/inyaml"
)

// PersistentService handles persistent configuration using Bonzai persisters
type PersistentService struct {
	config    *PersistentConfig
	persister *inyaml.Persister
}

// NewPersistentService creates a new persistent service
func NewPersistentService() *PersistentService {
	// Get user's home directory for config storage
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.TempDir()
	}

	configDir := filepath.Join(homeDir, configs.ConfigDirName)
	configFile := filepath.Join(configDir, configs.ConfigFileName)

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, configs.ConfigDirPerms); err != nil {
		// Fallback to temp directory if home directory fails
		configDir = configs.TempDirFallback
		configFile = filepath.Join(configDir, configs.FallbackConfigDir)
	}

	// Create YAML persister for file persistence
	persister := &inyaml.Persister{File: configFile}

	// Setup the persister
	if err := persister.Setup(); err != nil {
		// If setup fails, continue with default config
		config := DefaultPersistentConfig()
		return &PersistentService{
			config:    config,
			persister: persister,
		}
	}

	// Load existing configuration or create default
	config := DefaultPersistentConfig()

	// Load from persister using key-value approach
	if verboseLevel := persister.Get("verbose_level"); verboseLevel != "" {
		if level, err := strconv.Atoi(verboseLevel); err == nil && level >= 0 && level <= 4 {
			config.VerboseLevel = logger.LogLevel(level)
		}
	}

	if verboseEnabled := persister.Get("verbose_enabled"); verboseEnabled != "" {
		if enabled, err := strconv.ParseBool(verboseEnabled); err == nil {
			config.VerboseEnabled = enabled
		}
	}

	if cleanEnabled := persister.Get("clean_enabled"); cleanEnabled != "" {
		if enabled, err := strconv.ParseBool(cleanEnabled); err == nil {
			config.CleanEnabled = enabled
		}
	}

	if debugEnabled := persister.Get("debug_enabled"); debugEnabled != "" {
		if enabled, err := strconv.ParseBool(debugEnabled); err == nil {
			config.DebugEnabled = enabled
		}
	}

	if debugOutput := persister.Get("debug_output"); debugOutput != "" {
		config.DebugOutput = debugOutput
	}

	if forceEnabled := persister.Get("force_enabled"); forceEnabled != "" {
		if enabled, err := strconv.ParseBool(forceEnabled); err == nil {
			config.ForceEnabled = enabled
		}
	}

	return &PersistentService{
		config:    config,
		persister: persister,
	}
}

// GetConfig returns the current persistent configuration
func (ps *PersistentService) GetConfig() *PersistentConfig {
	return ps.config
}

// SaveConfig saves the current configuration to persistent storage
func (ps *PersistentService) SaveConfig() error {
	// Update timestamp
	ps.config.LastUpdated = time.Now()

	// Save each field as key-value pairs
	ps.persister.Set("verbose_level", strconv.Itoa(int(ps.config.VerboseLevel)))
	ps.persister.Set("verbose_enabled", strconv.FormatBool(ps.config.VerboseEnabled))
	ps.persister.Set("clean_enabled", strconv.FormatBool(ps.config.CleanEnabled))
	ps.persister.Set("debug_enabled", strconv.FormatBool(ps.config.DebugEnabled))
	ps.persister.Set("debug_output", ps.config.DebugOutput)
	ps.persister.Set("force_enabled", strconv.FormatBool(ps.config.ForceEnabled))
	ps.persister.Set("last_updated", ps.config.LastUpdated.Format(time.RFC3339))
	ps.persister.Set("version", ps.config.Version)

	return nil
}

// SetVerboseLevel sets the verbose level and persists it
func (ps *PersistentService) SetVerboseLevel(level logger.LogLevel) error {
	ps.config.SetVerboseLevel(level)

	// Persist to file
	return ps.SaveConfig()
}

// GetVerboseLevel returns the current verbose level
func (ps *PersistentService) GetVerboseLevel() logger.LogLevel {
	return ps.config.VerboseLevel
}

// SetCleanEnabled sets the clean setting and persists it
func (ps *PersistentService) SetCleanEnabled(enabled bool) error {
	ps.config.SetCleanEnabled(enabled)

	// Persist to file
	return ps.SaveConfig()
}

// GetCleanEnabled returns the current clean setting
func (ps *PersistentService) GetCleanEnabled() bool {
	return ps.config.CleanEnabled
}

// ToggleClean toggles the clean setting and persists it
func (ps *PersistentService) ToggleClean() (bool, error) {
	enabled := ps.config.ToggleClean()

	// Persist to file
	err := ps.SaveConfig()
	return enabled, err
}

// SetDebugEnabled sets the debug setting and persists it
func (ps *PersistentService) SetDebugEnabled(enabled bool, output string) error {
	ps.config.SetDebugEnabled(enabled, output)

	// Persist to file
	return ps.SaveConfig()
}

// GetDebugEnabled returns the current debug setting
func (ps *PersistentService) GetDebugEnabled() bool {
	return ps.config.DebugEnabled
}

// ToggleDebug toggles the debug setting and persists it
func (ps *PersistentService) ToggleDebug() (bool, error) {
	enabled := ps.config.ToggleDebug()

	// Persist to file
	err := ps.SaveConfig()
	return enabled, err
}

// SetForceEnabled sets the force setting and persists it
func (ps *PersistentService) SetForceEnabled(enabled bool) error {
	ps.config.SetForceEnabled(enabled)

	// Persist to file
	return ps.SaveConfig()
}

// GetForceEnabled returns the current force setting
func (ps *PersistentService) GetForceEnabled() bool {
	return ps.config.ForceEnabled
}

// ToggleForce toggles the force setting and persists it
func (ps *PersistentService) ToggleForce() (bool, error) {
	enabled := ps.config.ToggleForce()

	// Persist to file
	err := ps.SaveConfig()
	return enabled, err
}

// GetStatus returns the current status of all settings
func (ps *PersistentService) GetStatus() map[string]interface{} {
	return ps.config.GetStatus()
}

// ResetToDefaults resets all settings to defaults and persists them
func (ps *PersistentService) ResetToDefaults() error {
	ps.config = DefaultPersistentConfig()

	// Persist to file
	return ps.SaveConfig()
}

// GetConfigPath returns the path to the configuration file
func (ps *PersistentService) GetConfigPath() string {
	return ps.persister.File
}

// LoadFromFile loads configuration from the persistent file
func (ps *PersistentService) LoadFromFile() error {
	config := DefaultPersistentConfig()

	// Load from persister using key-value approach
	if verboseLevel := ps.persister.Get("verbose_level"); verboseLevel != "" {
		if level, err := strconv.Atoi(verboseLevel); err == nil && level >= 0 && level <= 4 {
			config.VerboseLevel = logger.LogLevel(level)
		}
	}

	if verboseEnabled := ps.persister.Get("verbose_enabled"); verboseEnabled != "" {
		if enabled, err := strconv.ParseBool(verboseEnabled); err == nil {
			config.VerboseEnabled = enabled
		}
	}

	if cleanEnabled := ps.persister.Get("clean_enabled"); cleanEnabled != "" {
		if enabled, err := strconv.ParseBool(cleanEnabled); err == nil {
			config.CleanEnabled = enabled
		}
	}

	if debugEnabled := ps.persister.Get("debug_enabled"); debugEnabled != "" {
		if enabled, err := strconv.ParseBool(debugEnabled); err == nil {
			config.DebugEnabled = enabled
		}
	}

	if debugOutput := ps.persister.Get("debug_output"); debugOutput != "" {
		config.DebugOutput = debugOutput
	}

	if forceEnabled := ps.persister.Get("force_enabled"); forceEnabled != "" {
		if enabled, err := strconv.ParseBool(forceEnabled); err == nil {
			config.ForceEnabled = enabled
		}
	}

	ps.config = config
	return nil
}

// ExportConfig exports the current configuration to a file
func (ps *PersistentService) ExportConfig(path string) error {
	exportPersister := &inyaml.Persister{File: path}
	if err := exportPersister.Setup(); err != nil {
		return fmt.Errorf("failed to setup export persister: %w", err)
	}

	// Save each field as key-value pairs
	exportPersister.Set("verbose_level", strconv.Itoa(int(ps.config.VerboseLevel)))
	exportPersister.Set("verbose_enabled", strconv.FormatBool(ps.config.VerboseEnabled))
	exportPersister.Set("clean_enabled", strconv.FormatBool(ps.config.CleanEnabled))
	exportPersister.Set("debug_enabled", strconv.FormatBool(ps.config.DebugEnabled))
	exportPersister.Set("debug_output", ps.config.DebugOutput)
	exportPersister.Set("force_enabled", strconv.FormatBool(ps.config.ForceEnabled))
	exportPersister.Set("last_updated", ps.config.LastUpdated.Format(time.RFC3339))
	exportPersister.Set("version", ps.config.Version)

	return nil
}

// ImportConfig imports configuration from a file
func (ps *PersistentService) ImportConfig(path string) error {
	importPersister := &inyaml.Persister{File: path}
	if err := importPersister.Setup(); err != nil {
		return fmt.Errorf("failed to setup import persister: %w", err)
	}

	config := DefaultPersistentConfig()

	// Load from persister using key-value approach
	if verboseLevel := importPersister.Get("verbose_level"); verboseLevel != "" {
		if level, err := strconv.Atoi(verboseLevel); err == nil && level >= 0 && level <= 4 {
			config.VerboseLevel = logger.LogLevel(level)
		}
	}

	if verboseEnabled := importPersister.Get("verbose_enabled"); verboseEnabled != "" {
		if enabled, err := strconv.ParseBool(verboseEnabled); err == nil {
			config.VerboseEnabled = enabled
		}
	}

	if cleanEnabled := importPersister.Get("clean_enabled"); cleanEnabled != "" {
		if enabled, err := strconv.ParseBool(cleanEnabled); err == nil {
			config.CleanEnabled = enabled
		}
	}

	if debugEnabled := importPersister.Get("debug_enabled"); debugEnabled != "" {
		if enabled, err := strconv.ParseBool(debugEnabled); err == nil {
			config.DebugEnabled = enabled
		}
	}

	if debugOutput := importPersister.Get("debug_output"); debugOutput != "" {
		config.DebugOutput = debugOutput
	}

	if forceEnabled := importPersister.Get("force_enabled"); forceEnabled != "" {
		if enabled, err := strconv.ParseBool(forceEnabled); err == nil {
			config.ForceEnabled = enabled
		}
	}

	ps.config = config

	// Persist to main config file
	return ps.SaveConfig()
}
