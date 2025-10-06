package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// VariableType represents the type of a variable
type VariableType string

const (
	VariableTypeString  VariableType = "string"
	VariableTypeNumber  VariableType = "number"
	VariableTypeBoolean VariableType = "boolean"
	VariableTypeArray   VariableType = "array"
	VariableTypeObject  VariableType = "object"
	VariableTypeNull    VariableType = "null"
)

// Variable errors
var (
	ErrEmptyString    = errors.New("string variable cannot be empty")
	ErrStringTooLong  = errors.New("string variable cannot exceed 1000 characters")
	ErrNegativeNumber = errors.New("number variable cannot be negative")
	ErrNumberTooLarge = errors.New("number variable is too large")
	ErrArrayTooLarge  = errors.New("array variable cannot exceed 1000 elements")
	ErrObjectTooLarge = errors.New("object variable cannot exceed 1000 keys")
	ErrInvalidType    = errors.New("invalid variable type")
	ErrInvalidValue   = errors.New("invalid variable value")
)

// Variable represents a template variable that can hold any JSON-serializable value
type Variable struct {
	Type  VariableType `json:"type"`
	Value interface{}  `json:"value"`
}

// NewStringVariable creates a new string variable with validation
func NewStringVariable(value string) (*Variable, error) {
	if strings.TrimSpace(value) == "" {
		return nil, ErrEmptyString
	}
	if len(value) > 1000 {
		return nil, ErrStringTooLong
	}
	return &Variable{
		Type:  VariableTypeString,
		Value: strings.TrimSpace(value),
	}, nil
}

// NewNumberVariable creates a new number variable with validation
func NewNumberVariable(value float64) (*Variable, error) {
	if value < 0 {
		return nil, ErrNegativeNumber
	}
	if value > 1e10 {
		return nil, ErrNumberTooLarge
	}
	return &Variable{
		Type:  VariableTypeNumber,
		Value: value,
	}, nil
}

// NewBooleanVariable creates a new boolean variable
func NewBooleanVariable(value bool) *Variable {
	return &Variable{
		Type:  VariableTypeBoolean,
		Value: value,
	}
}

// NewArrayVariable creates a new array variable with validation
func NewArrayVariable(value []interface{}) (*Variable, error) {
	if len(value) > 1000 {
		return nil, ErrArrayTooLarge
	}
	return &Variable{
		Type:  VariableTypeArray,
		Value: value,
	}, nil
}

// NewObjectVariable creates a new object variable with validation
func NewObjectVariable(value map[string]interface{}) (*Variable, error) {
	if len(value) > 1000 {
		return nil, ErrObjectTooLarge
	}
	return &Variable{
		Type:  VariableTypeObject,
		Value: value,
	}, nil
}

// NewNullVariable creates a new null variable
func NewNullVariable() *Variable {
	return &Variable{
		Type:  VariableTypeNull,
		Value: nil,
	}
}

// AsString returns the value as a string
func (v *Variable) AsString() (string, error) {
	if v.Type != VariableTypeString {
		return "", fmt.Errorf("variable is not a string, got %s", v.Type)
	}
	return v.Value.(string), nil
}

// AsNumber returns the value as a float64
func (v *Variable) AsNumber() (float64, error) {
	if v.Type != VariableTypeNumber {
		return 0, fmt.Errorf("variable is not a number, got %s", v.Type)
	}
	return v.Value.(float64), nil
}

// AsBoolean returns the value as a boolean
func (v *Variable) AsBoolean() (bool, error) {
	if v.Type != VariableTypeBoolean {
		return false, fmt.Errorf("variable is not a boolean, got %s", v.Type)
	}
	return v.Value.(bool), nil
}

// AsArray returns the value as a slice
func (v *Variable) AsArray() ([]interface{}, error) {
	if v.Type != VariableTypeArray {
		return nil, fmt.Errorf("variable is not an array, got %s", v.Type)
	}
	return v.Value.([]interface{}), nil
}

// AsObject returns the value as a map
func (v *Variable) AsObject() (map[string]interface{}, error) {
	if v.Type != VariableTypeObject {
		return nil, fmt.Errorf("variable is not an object, got %s", v.Type)
	}
	return v.Value.(map[string]interface{}), nil
}

// IsNull returns true if the variable is null
func (v *Variable) IsNull() bool {
	return v.Type == VariableTypeNull
}

// Validate validates the variable according to business rules
func (v *Variable) Validate() error {
	switch v.Type {
	case VariableTypeString:
		if str, ok := v.Value.(string); ok {
			if strings.TrimSpace(str) == "" {
				return ErrEmptyString
			}
			if len(str) > 1000 {
				return ErrStringTooLong
			}
		}
	case VariableTypeNumber:
		if num, ok := v.Value.(float64); ok {
			if num < 0 {
				return ErrNegativeNumber
			}
			if num > 1e10 {
				return ErrNumberTooLarge
			}
		}
	case VariableTypeArray:
		if arr, ok := v.Value.([]interface{}); ok {
			if len(arr) > 1000 {
				return ErrArrayTooLarge
			}
		}
	case VariableTypeObject:
		if obj, ok := v.Value.(map[string]interface{}); ok {
			if len(obj) > 1000 {
				return ErrObjectTooLarge
			}
		}
	}
	return nil
}

// CanBeUsedInTemplate returns true if the variable can be used in template processing
func (v *Variable) CanBeUsedInTemplate() bool {
	return v.Type != VariableTypeNull && v.Value != nil
}

// IsEmpty returns true if the variable is empty or null
func (v *Variable) IsEmpty() bool {
	if v.IsNull() {
		return true
	}

	switch v.Type {
	case VariableTypeString:
		if str, ok := v.Value.(string); ok {
			return strings.TrimSpace(str) == ""
		}
	case VariableTypeArray:
		if arr, ok := v.Value.([]interface{}); ok {
			return len(arr) == 0
		}
	case VariableTypeObject:
		if obj, ok := v.Value.(map[string]interface{}); ok {
			return len(obj) == 0
		}
	}

	return false
}

// GetSize returns the size of the variable
func (v *Variable) GetSize() int {
	switch v.Type {
	case VariableTypeString:
		if str, ok := v.Value.(string); ok {
			return len(str)
		}
	case VariableTypeArray:
		if arr, ok := v.Value.([]interface{}); ok {
			return len(arr)
		}
	case VariableTypeObject:
		if obj, ok := v.Value.(map[string]interface{}); ok {
			return len(obj)
		}
	}
	return 0
}

// Equals compares two variables for equality
func (v *Variable) Equals(other *Variable) bool {
	if v.Type != other.Type {
		return false
	}

	switch v.Type {
	case VariableTypeString:
		str1, _ := v.AsString()
		str2, _ := other.AsString()
		return str1 == str2
	case VariableTypeNumber:
		num1, _ := v.AsNumber()
		num2, _ := other.AsNumber()
		return num1 == num2
	case VariableTypeBoolean:
		bool1, _ := v.AsBoolean()
		bool2, _ := other.AsBoolean()
		return bool1 == bool2
	case VariableTypeNull:
		return true
	default:
		// For arrays and objects, use JSON comparison
		json1, _ := json.Marshal(v.Value)
		json2, _ := json.Marshal(other.Value)
		return string(json1) == string(json2)
	}
}

// String returns a string representation of the variable
func (v *Variable) String() string {
	switch v.Type {
	case VariableTypeString:
		return fmt.Sprintf("\"%s\"", v.Value)
	case VariableTypeNumber:
		return fmt.Sprintf("%v", v.Value)
	case VariableTypeBoolean:
		return fmt.Sprintf("%t", v.Value)
	case VariableTypeNull:
		return "null"
	default:
		jsonBytes, _ := json.Marshal(v.Value)
		return string(jsonBytes)
	}
}

// MarshalJSON implements json.Marshaler
func (v *Variable) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Value)
}

// UnmarshalJSON implements json.Unmarshaler
func (v *Variable) UnmarshalJSON(data []byte) error {
	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	v.Value = value
	v.Type = DetermineType(value)
	return nil
}

// DetermineType determines the variable type based on the Go type
func DetermineType(value interface{}) VariableType {
	if value == nil {
		return VariableTypeNull
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.String:
		return VariableTypeString
	case reflect.Float64, reflect.Int, reflect.Int64, reflect.Float32:
		return VariableTypeNumber
	case reflect.Bool:
		return VariableTypeBoolean
	case reflect.Slice, reflect.Array:
		return VariableTypeArray
	case reflect.Map:
		return VariableTypeObject
	default:
		return VariableTypeString
	}
}
