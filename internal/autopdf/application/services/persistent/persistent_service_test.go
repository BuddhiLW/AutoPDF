// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package persistent

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStore struct {
	path   string
	values map[string]string
}

func (store *testStore) Get(key string) string { return store.values[key] }
func (store *testStore) Set(key, value string) { store.values[key] = value }
func (store *testStore) Path() string          { return store.path }

type testStoreFactory struct {
	defaultStore Store
	stores       map[string]Store
	err          error
}

func (factory testStoreFactory) OpenDefault() (Store, error) {
	return factory.defaultStore, factory.err
}

func (factory testStoreFactory) Open(path string) (Store, error) {
	if factory.err != nil {
		return nil, factory.err
	}
	if store, ok := factory.stores[path]; ok {
		return store, nil
	}
	return nil, errors.New("store not found")
}

func TestPersistentServiceUsesInjectedStore(t *testing.T) {
	store := &testStore{path: "settings.yaml", values: map[string]string{"clean_enabled": "true"}}
	service := NewPersistentService(testStoreFactory{defaultStore: store})

	assert.True(t, service.GetCleanEnabled())
	require.NoError(t, service.SetForceEnabled(true))
	assert.Equal(t, "true", store.values["force_enabled"])
	assert.Equal(t, "settings.yaml", service.GetConfigPath())
}

func TestPersistentServiceReportsUnavailableStore(t *testing.T) {
	service := NewPersistentService(testStoreFactory{err: errors.New("disk unavailable")})
	assert.ErrorContains(t, service.SaveConfig(), "disk unavailable")
}
