// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package converter

import (
	"fmt"
	"reflect"
	"sync"
)

// ConverterRegistry manages custom converters for specific types
type ConverterRegistry struct {
	converters map[reflect.Type]Converter
	mu         sync.RWMutex
}

// NewConverterRegistry creates a new converter registry
func NewConverterRegistry() *ConverterRegistry {
	return &ConverterRegistry{
		converters: make(map[reflect.Type]Converter),
	}
}

// Register registers a converter for a specific type
func (r *ConverterRegistry) Register(typ reflect.Type, converter Converter) error {
	if typ == nil {
		return fmt.Errorf("type cannot be nil")
	}
	if converter == nil {
		return fmt.Errorf("converter cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.converters[typ] = converter
	return nil
}

// RegisterInterface registers a converter for an interface type
func (r *ConverterRegistry) RegisterInterface(typ reflect.Type, converter Converter) error {
	if typ == nil {
		return fmt.Errorf("type cannot be nil")
	}
	if typ.Kind() != reflect.Interface {
		return fmt.Errorf("type must be an interface, got %s", typ.Kind())
	}
	if converter == nil {
		return fmt.Errorf("converter cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.converters[typ] = converter
	return nil
}

// Get retrieves a converter for a specific type
func (r *ConverterRegistry) Get(typ reflect.Type) (Converter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Direct type match
	if converter, exists := r.converters[typ]; exists {
		return converter, true
	}

	// Check for interface implementations
	for interfaceType, converter := range r.converters {
		if interfaceType.Kind() == reflect.Interface && typ.Implements(interfaceType) {
			return converter, true
		}
	}

	return nil, false
}

// Unregister removes a converter for a specific type
func (r *ConverterRegistry) Unregister(typ reflect.Type) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.converters, typ)
}

// List returns all registered types
func (r *ConverterRegistry) List() []reflect.Type {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]reflect.Type, 0, len(r.converters))
	for typ := range r.converters {
		types = append(types, typ)
	}
	return types
}

// Clear removes all registered converters
func (r *ConverterRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.converters = make(map[reflect.Type]Converter)
}

// Count returns the number of registered converters
func (r *ConverterRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.converters)
}

// Has checks if a converter is registered for a specific type
func (r *ConverterRegistry) Has(typ reflect.Type) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.converters[typ]
	return exists
}

// RegisterBuiltinConverters registers all built-in converters
func (r *ConverterRegistry) RegisterBuiltinConverters() error {
	// This will be implemented in builtin.go
	return RegisterBuiltinConverters(r)
}
