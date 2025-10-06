package domain

import (
	"fmt"
	"strings"
)

// VariableCollection represents a collection of template variables with support for nested access
type VariableCollection struct {
	variables map[string]*Variable
}

// NewVariableCollection creates a new variable collection
func NewVariableCollection() *VariableCollection {
	return &VariableCollection{
		variables: make(map[string]*Variable),
	}
}

// Set sets a variable in the collection
func (vc *VariableCollection) Set(key string, value *Variable) {
	vc.variables[key] = value
}

// Get retrieves a variable from the collection
func (vc *VariableCollection) Get(key string) (*Variable, bool) {
	v, exists := vc.variables[key]
	return v, exists
}

// GetNested retrieves a nested variable using dot notation (e.g., "user.name")
func (vc *VariableCollection) GetNested(key string) (*Variable, error) {
	parts := strings.Split(key, ".")
	if len(parts) == 1 {
		variable, exists := vc.Get(key)
		if !exists {
			return nil, fmt.Errorf("variable '%s' not found", key)
		}
		return variable, nil
	}

	// Get the root variable
	rootKey := parts[0]
	rootVar, exists := vc.Get(rootKey)
	if !exists {
		return nil, fmt.Errorf("variable '%s' not found", rootKey)
	}

	// Navigate through nested objects
	current := rootVar
	for i := 1; i < len(parts); i++ {
		if current.Type != VariableTypeObject {
			return nil, fmt.Errorf("variable '%s' is not an object", strings.Join(parts[:i], "."))
		}

		obj, err := current.AsObject()
		if err != nil {
			return nil, err
		}

		key := parts[i]
		value, exists := obj[key]
		if !exists {
			return nil, fmt.Errorf("key '%s' not found in object", key)
		}

		// Convert the value to a Variable
		current = &Variable{
			Type:  DetermineType(value),
			Value: value,
		}
	}

	return current, nil
}

// SetNested sets a nested variable using dot notation
func (vc *VariableCollection) SetNested(key string, value *Variable) error {
	parts := strings.Split(key, ".")
	if len(parts) == 1 {
		vc.Set(key, value)
		return nil
	}

	// Get or create the root variable
	rootKey := parts[0]
	rootVar, exists := vc.Get(rootKey)
	if !exists {
		rootVar, _ = NewObjectVariable(make(map[string]interface{}))
		vc.Set(rootKey, rootVar)
	}

	if rootVar.Type != VariableTypeObject {
		return fmt.Errorf("variable '%s' is not an object", rootKey)
	}

	// Navigate to the parent object
	obj, err := rootVar.AsObject()
	if err != nil {
		return err
	}

	// Create nested objects as needed
	current := obj
	for i := 1; i < len(parts)-1; i++ {
		key := parts[i]
		if _, exists := current[key]; !exists {
			current[key] = make(map[string]interface{})
		}

		nextObj, ok := current[key].(map[string]interface{})
		if !ok {
			return fmt.Errorf("key '%s' is not an object", strings.Join(parts[:i+1], "."))
		}
		current = nextObj
	}

	// Set the final value
	finalKey := parts[len(parts)-1]
	current[finalKey] = value.Value
	return nil
}

// GetAll returns all variables
func (vc *VariableCollection) GetAll() map[string]*Variable {
	return vc.variables
}

// Size returns the number of variables
func (vc *VariableCollection) Size() int {
	return len(vc.variables)
}

// Clear removes all variables
func (vc *VariableCollection) Clear() {
	vc.variables = make(map[string]*Variable)
}

// Keys returns all variable keys
func (vc *VariableCollection) Keys() []string {
	keys := make([]string, 0, len(vc.variables))
	for key := range vc.variables {
		keys = append(keys, key)
	}
	return keys
}

// Has checks if a variable exists
func (vc *VariableCollection) Has(key string) bool {
	_, exists := vc.variables[key]
	return exists
}

// Delete removes a variable
func (vc *VariableCollection) Delete(key string) {
	delete(vc.variables, key)
}

// ToMap converts the variable collection to a map for template processing
func (vc *VariableCollection) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	for key, variable := range vc.variables {
		result[key] = variable.Value
	}
	return result
}

// FromMap creates a variable collection from a map
func FromMap(data map[string]interface{}) *VariableCollection {
	vc := NewVariableCollection()
	for key, value := range data {
		variable := &Variable{
			Type:  DetermineType(value),
			Value: value,
		}
		vc.Set(key, variable)
	}
	return vc
}
