package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewVariableCollection(t *testing.T) {
	vc := NewVariableCollection()
	if vc.variables == nil {
		t.Error("Expected variables map to be initialized")
	}
	if vc.Size() != 0 {
		t.Errorf("Expected size 0, got %d", vc.Size())
	}
}

func TestSetAndGet(t *testing.T) {
	vc := NewVariableCollection()
	v, _ := NewStringVariable("test")

	vc.Set("key", v)

	retrieved, exists := vc.Get("key")
	if !exists {
		t.Error("Expected variable to exist")
	}
	if retrieved != v {
		t.Error("Expected retrieved variable to be the same as set")
	}
}

func TestGetNested(t *testing.T) {
	vc := NewVariableCollection()

	// Create a nested structure
	user, err := NewObjectVariable(map[string]interface{}{
		"name": "John",
		"age":  30,
		"address": map[string]interface{}{
			"street": "123 Main St",
			"city":   "New York",
		},
	})
	require.NoError(t, err)
	vc.Set("user", user)

	// Test nested access
	tests := []struct {
		key      string
		expected interface{}
		hasError bool
	}{
		{"user.name", "John", false},
		{"user.age", 30, false},
		{"user.address.street", "123 Main St", false},
		{"user.address.city", "New York", false},
		{"user.nonexistent", nil, true},
		{"nonexistent.key", nil, true},
		{"user.address.nonexistent", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result, err := vc.GetNested(tt.key)
			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for key '%s', got nil", tt.key)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for key '%s': %v", tt.key, err)
				return
			}

			if result.Value != tt.expected {
				t.Errorf("Expected value %v for key '%s', got %v", tt.expected, tt.key, result.Value)
			}
		})
	}
}

func TestSetNested(t *testing.T) {
	vc := NewVariableCollection()

	// Test setting nested values
	nameVar, err := NewStringVariable("John")
	require.NoError(t, err)
	ageVar, err := NewNumberVariable(30)
	require.NoError(t, err)
	streetVar, err := NewStringVariable("123 Main St")
	require.NoError(t, err)
	cityVar, err := NewStringVariable("New York")
	require.NoError(t, err)

	tests := []struct {
		key   string
		value *Variable
	}{
		{"user.name", nameVar},
		{"user.age", ageVar},
		{"user.address.street", streetVar},
		{"user.address.city", cityVar},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			err := vc.SetNested(tt.key, tt.value)
			if err != nil {
				t.Errorf("Unexpected error setting '%s': %v", tt.key, err)
				return
			}

			// Verify the value was set correctly
			result, err := vc.GetNested(tt.key)
			if err != nil {
				t.Errorf("Error retrieving '%s': %v", tt.key, err)
				return
			}

			if result.Value != tt.value.Value {
				t.Errorf("Expected value %v for key '%s', got %v", tt.value.Value, tt.key, result.Value)
			}
		})
	}
}

func TestGetAll(t *testing.T) {
	vc := NewVariableCollection()
	v1, _ := NewStringVariable("test1")
	v2, _ := NewStringVariable("test2")

	vc.Set("key1", v1)
	vc.Set("key2", v2)

	all := vc.GetAll()
	if len(all) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(all))
	}

	if all["key1"] != v1 {
		t.Error("Expected key1 to match v1")
	}
	if all["key2"] != v2 {
		t.Error("Expected key2 to match v2")
	}
}

func TestSize(t *testing.T) {
	vc := NewVariableCollection()
	if vc.Size() != 0 {
		t.Errorf("Expected size 0, got %d", vc.Size())
	}

	v, err := NewStringVariable("value")
	require.NoError(t, err)
	vc.Set("key", v)
	require.NoError(t, err)
	if vc.Size() != 1 {
		t.Errorf("Expected size 1, got %d", vc.Size())
	}
}

func TestClear(t *testing.T) {
	vc := NewVariableCollection()
	v, err := NewStringVariable("value")
	require.NoError(t, err)
	vc.Set("key", v)
	require.NoError(t, err)

	if vc.Size() != 1 {
		t.Errorf("Expected size 1, got %d", vc.Size())
	}

	vc.Clear()
	if vc.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", vc.Size())
	}
}

func TestKeys(t *testing.T) {
	vc := NewVariableCollection()
	v1, err := NewStringVariable("value1")
	require.NoError(t, err)
	vc.Set("key1", v1)
	require.NoError(t, err)
	v2, err := NewStringVariable("value2")
	require.NoError(t, err)
	vc.Set("key2", v2)
	require.NoError(t, err)

	keys := vc.Keys()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}

	// Check that both keys are present
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}

	if !keyMap["key1"] {
		t.Error("Expected key1 to be present")
	}
	if !keyMap["key2"] {
		t.Error("Expected key2 to be present")
	}
}

func TestHas(t *testing.T) {
	vc := NewVariableCollection()

	if vc.Has("key") {
		t.Error("Expected key to not exist")
	}

	v, err := NewStringVariable("value")
	require.NoError(t, err)
	vc.Set("key", v)
	require.NoError(t, err)
	if !vc.Has("key") {
		t.Error("Expected key to exist")
	}
}

func TestDelete(t *testing.T) {
	vc := NewVariableCollection()
	v, err := NewStringVariable("value")
	require.NoError(t, err)
	vc.Set("key", v)
	require.NoError(t, err)

	if !vc.Has("key") {
		t.Error("Expected key to exist before deletion")
	}

	vc.Delete("key")
	if vc.Has("key") {
		t.Error("Expected key to not exist after deletion")
	}
}

func TestToMap(t *testing.T) {
	vc := NewVariableCollection()
	v1, err := NewStringVariable("test")
	require.NoError(t, err)
	vc.Set("string", v1)
	require.NoError(t, err)
	v2, err := NewNumberVariable(42)
	require.NoError(t, err)
	vc.Set("number", v2)
	require.NoError(t, err)
	v3 := NewBooleanVariable(true)
	require.NoError(t, err)
	vc.Set("boolean", v3)
	require.NoError(t, err)

	result := vc.ToMap()

	if result["string"] != "test" {
		t.Errorf("Expected string 'test', got %v", result["string"])
	}
	if result["number"] != float64(42) {
		t.Errorf("Expected number 42, got %v", result["number"])
	}
	if result["boolean"] != true {
		t.Errorf("Expected boolean true, got %v", result["boolean"])
	}
}

func TestFromMap(t *testing.T) {
	data := map[string]interface{}{
		"string":  "test",
		"number":  42,
		"boolean": true,
		"object": map[string]interface{}{
			"nested": "value",
		},
	}

	vc := FromMap(data)

	if vc.Size() != 4 {
		t.Errorf("Expected size 4, got %d", vc.Size())
	}

	// Test string variable
	strVar, exists := vc.Get("string")
	if !exists {
		t.Error("Expected string variable to exist")
	}
	if strVar.Type != VariableTypeString {
		t.Errorf("Expected string type, got %s", strVar.Type)
	}

	// Test number variable
	numVar, exists := vc.Get("number")
	if !exists {
		t.Error("Expected number variable to exist")
	}
	if numVar.Type != VariableTypeNumber {
		t.Errorf("Expected number type, got %s", numVar.Type)
	}

	// Test boolean variable
	boolVar, exists := vc.Get("boolean")
	if !exists {
		t.Error("Expected boolean variable to exist")
	}
	if boolVar.Type != VariableTypeBoolean {
		t.Errorf("Expected boolean type, got %s", boolVar.Type)
	}

	// Test object variable
	objVar, exists := vc.Get("object")
	if !exists {
		t.Error("Expected object variable to exist")
	}
	if objVar.Type != VariableTypeObject {
		t.Errorf("Expected object type, got %s", objVar.Type)
	}
}
