// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package persistent

import (
	"fmt"
	"strconv"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
)

// Store is the persistence port required by PersistentService.
type Store interface {
	Get(key string) string
	Set(key, value string)
	Path() string
}

// StoreFactory opens the default store and explicit import/export stores.
type StoreFactory interface {
	OpenDefault() (Store, error)
	Open(path string) (Store, error)
}

type PersistentService struct {
	config   *PersistentConfig
	store    Store
	factory  StoreFactory
	storeErr error
}

// NewPersistentService creates a service around an injected persistence boundary.
// Without a factory it remains usable in memory and never claims disk persistence.
func NewPersistentService(factories ...StoreFactory) *PersistentService {
	var factory StoreFactory
	if len(factories) > 0 {
		factory = factories[0]
	}
	store := Store(newMemoryStore())
	var storeErr error
	if factory != nil {
		opened, err := factory.OpenDefault()
		if err == nil && opened != nil {
			store = opened
		} else if err != nil {
			storeErr = fmt.Errorf("open default persistent store: %w", err)
		} else {
			storeErr = fmt.Errorf("open default persistent store: factory returned nil store")
		}
	}
	return &PersistentService{config: loadConfig(store), store: store, factory: factory, storeErr: storeErr}
}

func (service *PersistentService) GetConfig() *PersistentConfig { return service.config }

func (service *PersistentService) SaveConfig() error {
	if service.storeErr != nil {
		return service.storeErr
	}
	service.config.LastUpdated = time.Now()
	writeConfig(service.store, service.config)
	return nil
}

func (service *PersistentService) SetVerboseLevel(level logger.LogLevel) error {
	service.config.SetVerboseLevel(level)
	return service.SaveConfig()
}

func (service *PersistentService) GetVerboseLevel() logger.LogLevel {
	return service.config.VerboseLevel
}

func (service *PersistentService) SetCleanEnabled(enabled bool) error {
	service.config.SetCleanEnabled(enabled)
	return service.SaveConfig()
}

func (service *PersistentService) GetCleanEnabled() bool { return service.config.CleanEnabled }

func (service *PersistentService) ToggleClean() (bool, error) {
	enabled := service.config.ToggleClean()
	return enabled, service.SaveConfig()
}

func (service *PersistentService) SetDebugEnabled(enabled bool, output string) error {
	service.config.SetDebugEnabled(enabled, output)
	return service.SaveConfig()
}

func (service *PersistentService) GetDebugEnabled() bool { return service.config.DebugEnabled }

func (service *PersistentService) ToggleDebug() (bool, error) {
	enabled := service.config.ToggleDebug()
	return enabled, service.SaveConfig()
}

func (service *PersistentService) SetForceEnabled(enabled bool) error {
	service.config.SetForceEnabled(enabled)
	return service.SaveConfig()
}

func (service *PersistentService) GetForceEnabled() bool { return service.config.ForceEnabled }

func (service *PersistentService) ToggleForce() (bool, error) {
	enabled := service.config.ToggleForce()
	return enabled, service.SaveConfig()
}

func (service *PersistentService) GetStatus() map[string]interface{} {
	return service.config.GetStatus()
}

func (service *PersistentService) ResetToDefaults() error {
	service.config = DefaultPersistentConfig()
	return service.SaveConfig()
}

func (service *PersistentService) GetConfigPath() string { return service.store.Path() }

func (service *PersistentService) LoadFromFile() error {
	if service.storeErr != nil {
		return service.storeErr
	}
	service.config = loadConfig(service.store)
	return nil
}

func (service *PersistentService) ExportConfig(path string) error {
	if service.factory == nil {
		return fmt.Errorf("persistent store factory is not configured")
	}
	store, err := service.factory.Open(path)
	if err != nil {
		return fmt.Errorf("open export store: %w", err)
	}
	writeConfig(store, service.config)
	return nil
}

func (service *PersistentService) ImportConfig(path string) error {
	if service.factory == nil {
		return fmt.Errorf("persistent store factory is not configured")
	}
	store, err := service.factory.Open(path)
	if err != nil {
		return fmt.Errorf("open import store: %w", err)
	}
	service.config = loadConfig(store)
	return service.SaveConfig()
}

func loadConfig(store Store) *PersistentConfig {
	config := DefaultPersistentConfig()
	if level, err := strconv.Atoi(store.Get("verbose_level")); err == nil && level >= 0 && level <= 4 {
		config.VerboseLevel = logger.LogLevel(level)
	}
	if enabled, err := strconv.ParseBool(store.Get("verbose_enabled")); err == nil {
		config.VerboseEnabled = enabled
	}
	if enabled, err := strconv.ParseBool(store.Get("clean_enabled")); err == nil {
		config.CleanEnabled = enabled
	}
	if enabled, err := strconv.ParseBool(store.Get("debug_enabled")); err == nil {
		config.DebugEnabled = enabled
	}
	if output := store.Get("debug_output"); output != "" {
		config.DebugOutput = output
	}
	if enabled, err := strconv.ParseBool(store.Get("force_enabled")); err == nil {
		config.ForceEnabled = enabled
	}
	return config
}

func writeConfig(store Store, config *PersistentConfig) {
	store.Set("verbose_level", strconv.Itoa(int(config.VerboseLevel)))
	store.Set("verbose_enabled", strconv.FormatBool(config.VerboseEnabled))
	store.Set("clean_enabled", strconv.FormatBool(config.CleanEnabled))
	store.Set("debug_enabled", strconv.FormatBool(config.DebugEnabled))
	store.Set("debug_output", config.DebugOutput)
	store.Set("force_enabled", strconv.FormatBool(config.ForceEnabled))
	store.Set("last_updated", config.LastUpdated.Format(time.RFC3339))
	store.Set("version", config.Version)
}

type memoryStore struct{ values map[string]string }

func newMemoryStore() *memoryStore               { return &memoryStore{values: map[string]string{}} }
func (store *memoryStore) Get(key string) string { return store.values[key] }
func (store *memoryStore) Set(key, value string) { store.values[key] = value }
func (store *memoryStore) Path() string          { return "" }
