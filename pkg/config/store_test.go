// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memoryStore map[string]string

func (store memoryStore) Get(key string) string { return store[key] }
func (store memoryStore) Set(key, value string) { store[key] = value }

func TestConfigPersistenceUsesStoreCapability(t *testing.T) {
	store := memoryStore{}
	cfg := GetDefaultConfig()
	cfg.Engine = "xelatex"

	require.NoError(t, SaveConfig(store, cfg))
	loaded, err := GetConfig(store)
	require.NoError(t, err)
	assert.Equal(t, Engine("xelatex"), loaded.Engine)
}

func TestConfigPersistenceRejectsNilStore(t *testing.T) {
	_, err := GetConfig(nil)
	assert.Error(t, err)
	assert.Error(t, SaveConfig(nil, GetDefaultConfig()))
}
