package domain

import (
	"testing"
)

func TestNewStringVariable(t *testing.T) {
	v, _ := NewStringVariable("test")
	if v.Type != VariableTypeString {
		t.Errorf("Expected type %s, got %s", VariableTypeString, v.Type)
	}
	if v.Value != "test" {
		t.Errorf("Expected value 'test', got %v", v.Value)
	}
}

func TestNewNumberVariable(t *testing.T) {
	v, err := NewNumberVariable(42.5)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if v.Type != VariableTypeNumber {
		t.Errorf("Expected type %s, got %s", VariableTypeNumber, v.Type)
	}
	if v.Value != 42.5 {
		t.Errorf("Expected value 42.5, got %v", v.Value)
	}
}

func TestNewBooleanVariable(t *testing.T) {
	v := NewBooleanVariable(true)
	if v.Type != VariableTypeBoolean {
		t.Errorf("Expected type %s, got %s", VariableTypeBoolean, v.Type)
	}
	if v.Value != true {
		t.Errorf("Expected value true, got %v", v.Value)
	}
}

func TestNewArrayVariable(t *testing.T) {
	arr := []interface{}{"item1", "item2", 42}
	v, _ := NewArrayVariable(arr)
	if v.Type != VariableTypeArray {
		t.Errorf("Expected type %s, got %s", VariableTypeArray, v.Type)
	}
	if len(v.Value.([]interface{})) != 3 {
		t.Errorf("Expected array length 3, got %d", len(v.Value.([]interface{})))
	}
}

func TestNewObjectVariable(t *testing.T) {
	obj := map[string]interface{}{
		"name": "John",
		"age":  30,
	}
	v, _ := NewObjectVariable(obj)
	if v.Type != VariableTypeObject {
		t.Errorf("Expected type %s, got %s", VariableTypeObject, v.Type)
	}
	if v.Value.(map[string]interface{})["name"] != "John" {
		t.Errorf("Expected name 'John', got %v", v.Value.(map[string]interface{})["name"])
	}
}

func TestNewNullVariable(t *testing.T) {
	v := NewNullVariable()
	if v.Type != VariableTypeNull {
		t.Errorf("Expected type %s, got %s", VariableTypeNull, v.Type)
	}
	if v.Value != nil {
		t.Errorf("Expected value nil, got %v", v.Value)
	}
}

func TestAsString(t *testing.T) {
	v, _ := NewStringVariable("test")
	result, err := v.AsString()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != "test" {
		t.Errorf("Expected 'test', got '%s'", result)
	}

	// Test with wrong type
	v2, err := NewNumberVariable(42)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	_, err = v2.AsString()
	if err == nil {
		t.Error("Expected error for wrong type, got nil")
	}
}

func TestAsNumber(t *testing.T) {
	v, err := NewNumberVariable(42.5)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	result, err := v.AsNumber()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != 42.5 {
		t.Errorf("Expected 42.5, got %f", result)
	}

	// Test with wrong type
	v2, err := NewStringVariable("test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	_, err = v2.AsNumber()
	if err == nil {
		t.Error("Expected error for wrong type, got nil")
	}
}

func TestAsBoolean(t *testing.T) {
	v := NewBooleanVariable(true)
	result, err := v.AsBoolean()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != true {
		t.Errorf("Expected true, got %t", result)
	}

	// Test with wrong type
	v2, err := NewStringVariable("test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	_, err = v2.AsBoolean()
	if err == nil {
		t.Error("Expected error for wrong type, got nil")
	}
}

func TestAsArray(t *testing.T) {
	arr := []interface{}{"item1", "item2"}
	v, _ := NewArrayVariable(arr)
	result, err := v.AsArray()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected array length 2, got %d", len(result))
	}

	// Test with wrong type
	v2, err := NewStringVariable("test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	_, err = v2.AsArray()
	if err == nil {
		t.Error("Expected error for wrong type, got nil")
	}
}

func TestAsObject(t *testing.T) {
	obj := map[string]interface{}{
		"name": "John",
		"age":  30,
	}
	v, err := NewObjectVariable(obj)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	result, err := v.AsObject()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result["name"] != "John" {
		t.Errorf("Expected name 'John', got %v", result["name"])
	}

	// Test with wrong type
	v2, err := NewStringVariable("test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	_, err = v2.AsObject()
	if err == nil {
		t.Error("Expected error for wrong type, got nil")
	}
}

func TestIsNull(t *testing.T) {
	v := NewNullVariable()
	if !v.IsNull() {
		t.Error("Expected null variable to be null")
	}

	v2, err := NewStringVariable("test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if v2.IsNull() {
		t.Error("Expected non-null variable to not be null")
	}
}

func TestString(t *testing.T) {
	stringVar, err := NewStringVariable("test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	numberVar, err := NewNumberVariable(42.5)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	booleanVar := NewBooleanVariable(true)
	nullVar := NewNullVariable()

	tests := []struct {
		name     string
		variable *Variable
		expected string
	}{
		{
			name:     "string",
			variable: stringVar,
			expected: "\"test\"",
		},
		{
			name:     "number",
			variable: numberVar,
			expected: "42.5",
		},
		{
			name:     "boolean",
			variable: booleanVar,
			expected: "true",
		},
		{
			name:     "null",
			variable: nullVar,
			expected: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.variable.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	v, _ := NewStringVariable("test")
	data, err := v.MarshalJSON()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if string(data) != "\"test\"" {
		t.Errorf("Expected '\"test\"', got '%s'", string(data))
	}
}

func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected VariableType
	}{
		{
			name:     "string",
			json:     "\"test\"",
			expected: VariableTypeString,
		},
		{
			name:     "number",
			json:     "42.5",
			expected: VariableTypeNumber,
		},
		{
			name:     "boolean",
			json:     "true",
			expected: VariableTypeBoolean,
		},
		{
			name:     "null",
			json:     "null",
			expected: VariableTypeNull,
		},
		{
			name:     "array",
			json:     "[\"item1\", \"item2\"]",
			expected: VariableTypeArray,
		},
		{
			name:     "object",
			json:     "{\"name\": \"John\", \"age\": 30}",
			expected: VariableTypeObject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v Variable
			err := v.UnmarshalJSON([]byte(tt.json))
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if v.Type != tt.expected {
				t.Errorf("Expected type %s, got %s", tt.expected, v.Type)
			}
		})
	}
}
