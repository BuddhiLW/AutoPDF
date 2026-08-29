// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BuddhiLW/AutoPDF/v2/configs"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/services/persistent"
	"github.com/rwxrob/bonzai/persisters/inyaml"
)

// YAMLStoreFactory keeps home-directory and inyaml effects at the infrastructure boundary.
type YAMLStoreFactory struct{}

func NewYAMLStoreFactory() *YAMLStoreFactory { return &YAMLStoreFactory{} }

func (*YAMLStoreFactory) OpenDefault() (persistent.Store, error) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		homeDirectory = os.TempDir()
	}
	configDirectory := filepath.Join(homeDirectory, configs.ConfigDirName)
	if err := os.MkdirAll(configDirectory, configs.ConfigDirPerms); err != nil {
		configDirectory = configs.TempDirFallback
		if fallbackErr := os.MkdirAll(configDirectory, configs.ConfigDirPerms); fallbackErr != nil {
			return nil, fmt.Errorf("create persistent config directory: %w", fallbackErr)
		}
	}
	return openYAMLStore(filepath.Join(configDirectory, configs.ConfigFileName))
}

func (*YAMLStoreFactory) Open(path string) (persistent.Store, error) {
	if path == "" {
		return nil, fmt.Errorf("persistent config path must not be empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), configs.ConfigDirPerms); err != nil {
		return nil, err
	}
	return openYAMLStore(path)
}

type yamlStore struct{ persister *inyaml.Persister }

func openYAMLStore(path string) (*yamlStore, error) {
	persister := &inyaml.Persister{File: path}
	if err := persister.Setup(); err != nil {
		return nil, err
	}
	return &yamlStore{persister: persister}, nil
}

func (store *yamlStore) Get(key string) string { return store.persister.Get(key) }
func (store *yamlStore) Set(key, value string) { store.persister.Set(key, value) }
func (store *yamlStore) Path() string          { return store.persister.File }
