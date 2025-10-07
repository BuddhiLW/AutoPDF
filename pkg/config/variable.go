// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Variable represents a complex variable that can be:
// - A simple string value
// - A map of nested variables
// - A slice of values
// - Any combination of the above
type Variable interface {
	// String returns the string representation for template substitution
	String() string
	// Get retrieves a nested value by path (e.g., "foo.bar[0]")
	Get(path string) (Variable, bool)
	// Set sets a nested value by path
	Set(path string, value Variable) error
	// Keys returns the keys for map-like variables
	Keys() []string
	// Len returns the length for slice-like variables
	Len() int
	// Type returns the variable type
	Type() VariableType
	// MarshalJSON implements json.Marshaler
	MarshalJSON() ([]byte, error)
	// UnmarshalJSON implements json.Unmarshaler
	UnmarshalJSON(data []byte) error
}

// VariableType represents the type of a variable
type VariableType int

const (
	VariableTypeString VariableType = iota
	VariableTypeMap
	VariableTypeSlice
	VariableTypeNumber
	VariableTypeBool
)

// StringVariable represents a simple string variable
type StringVariable struct {
	Value string
}

func (v StringVariable) String() string {
	return v.Value
}

func (v StringVariable) Get(path string) (Variable, bool) {
	if path == "" {
		return &v, true
	}
	return nil, false
}

func (v *StringVariable) Set(path string, value Variable) error {
	if path == "" {
		if strVar, ok := value.(*StringVariable); ok {
			v.Value = strVar.Value
			return nil
		}
		return fmt.Errorf("cannot set string variable to non-string value")
	}
	return fmt.Errorf("cannot set nested path on string variable")
}

func (v StringVariable) Keys() []string {
	return nil
}

func (v StringVariable) Len() int {
	return 0
}

func (v StringVariable) Type() VariableType {
	return VariableTypeString
}

func (v StringVariable) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Value)
}

func (v *StringVariable) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &v.Value)
}

// MapVariable represents a map of nested variables
type MapVariable struct {
	Values map[string]Variable
}

func NewMapVariable() *MapVariable {
	return &MapVariable{
		Values: make(map[string]Variable),
	}
}

func (v MapVariable) String() string {
	return fmt.Sprintf("%v", v.Values)
}

func (v MapVariable) Get(path string) (Variable, bool) {
	if path == "" {
		return &v, true
	}

	parts := parsePath(path)
	if len(parts) == 0 {
		return &v, true
	}

	key := parts[0]

	// Check if the key contains array access (e.g., "tags[0]")
	if strings.Contains(key, "[") && strings.Contains(key, "]") {
		// Extract the base key and array index
		baseKey := strings.Split(key, "[")[0]
		indexStr := strings.TrimSuffix(strings.Split(key, "[")[1], "]")

		// Get the base value
		if baseVal, exists := v.Values[baseKey]; exists {
			// Try to access the array index
			if sliceVar, ok := baseVal.(*SliceVariable); ok {
				if index, err := strconv.Atoi(indexStr); err == nil && index >= 0 && index < len(sliceVar.Values) {
					if len(parts) == 1 {
						return sliceVar.Values[index], true
					}
					return sliceVar.Values[index].Get(strings.Join(parts[1:], "."))
				}
			}
		}
		return nil, false
	}

	// Regular key access
	if val, exists := v.Values[key]; exists {
		if len(parts) == 1 {
			return val, true
		}
		return val.Get(strings.Join(parts[1:], "."))
	}
	return nil, false
}

func (v *MapVariable) Set(path string, value Variable) error {
	if path == "" {
		if mapVar, ok := value.(*MapVariable); ok {
			v.Values = mapVar.Values
			return nil
		}
		return fmt.Errorf("cannot set map variable to non-map value")
	}

	parts := parsePath(path)
	if len(parts) == 1 {
		v.Values[parts[0]] = value
		return nil
	}

	key := parts[0]
	if existing, exists := v.Values[key]; exists {
		return existing.Set(strings.Join(parts[1:], "."), value)
	}

	// Create nested map
	nested := NewMapVariable()
	if err := nested.Set(strings.Join(parts[1:], "."), value); err != nil {
		return err
	}
	v.Values[key] = nested
	return nil
}

func (v MapVariable) Keys() []string {
	keys := make([]string, 0, len(v.Values))
	for k := range v.Values {
		keys = append(keys, k)
	}
	return keys
}

func (v MapVariable) Len() int {
	return len(v.Values)
}

func (v MapVariable) Type() VariableType {
	return VariableTypeMap
}

func (v MapVariable) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Values)
}

func (v *MapVariable) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	v.Values = make(map[string]Variable)
	for k, val := range raw {
		v.Values[k] = convertToVariable(val)
	}
	return nil
}

// SliceVariable represents a slice of variables
type SliceVariable struct {
	Values []Variable
}

func NewSliceVariable() *SliceVariable {
	return &SliceVariable{
		Values: make([]Variable, 0),
	}
}

func (v SliceVariable) String() string {
	strs := make([]string, len(v.Values))
	for i, val := range v.Values {
		strs[i] = val.String()
	}
	return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
}

func (v SliceVariable) Get(path string) (Variable, bool) {
	if path == "" {
		return &v, true
	}

	parts := parsePath(path)
	if len(parts) == 0 {
		return &v, true
	}

	// Check if first part is an index
	if index, err := strconv.Atoi(parts[0]); err == nil && index >= 0 && index < len(v.Values) {
		if len(parts) == 1 {
			return v.Values[index], true
		}
		return v.Values[index].Get(strings.Join(parts[1:], "."))
	}
	return nil, false
}

func (v *SliceVariable) Set(path string, value Variable) error {
	if path == "" {
		if sliceVar, ok := value.(*SliceVariable); ok {
			v.Values = sliceVar.Values
			return nil
		}
		return fmt.Errorf("cannot set slice variable to non-slice value")
	}

	parts := parsePath(path)
	if len(parts) == 0 {
		return fmt.Errorf("empty path")
	}

	// Check if first part is an index
	if index, err := strconv.Atoi(parts[0]); err == nil {
		if index < 0 || index >= len(v.Values) {
			return fmt.Errorf("index %d out of range", index)
		}
		if len(parts) == 1 {
			v.Values[index] = value
			return nil
		}
		return v.Values[index].Set(strings.Join(parts[1:], "."), value)
	}

	return fmt.Errorf("invalid path for slice variable")
}

func (v SliceVariable) Keys() []string {
	return nil
}

func (v SliceVariable) Len() int {
	return len(v.Values)
}

func (v SliceVariable) Type() VariableType {
	return VariableTypeSlice
}

func (v SliceVariable) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Values)
}

func (v *SliceVariable) UnmarshalJSON(data []byte) error {
	var raw []interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	v.Values = make([]Variable, len(raw))
	for i, val := range raw {
		v.Values[i] = convertToVariable(val)
	}
	return nil
}

// NumberVariable represents a numeric variable
type NumberVariable struct {
	Value float64
}

func (v NumberVariable) String() string {
	return strconv.FormatFloat(v.Value, 'f', -1, 64)
}

func (v NumberVariable) Get(path string) (Variable, bool) {
	if path == "" {
		return &v, true
	}
	return nil, false
}

func (v *NumberVariable) Set(path string, value Variable) error {
	if path == "" {
		if numVar, ok := value.(*NumberVariable); ok {
			v.Value = numVar.Value
			return nil
		}
		return fmt.Errorf("cannot set number variable to non-number value")
	}
	return fmt.Errorf("cannot set nested path on number variable")
}

func (v NumberVariable) Keys() []string {
	return nil
}

func (v NumberVariable) Len() int {
	return 0
}

func (v NumberVariable) Type() VariableType {
	return VariableTypeNumber
}

func (v NumberVariable) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Value)
}

func (v *NumberVariable) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &v.Value)
}

// BoolVariable represents a boolean variable
type BoolVariable struct {
	Value bool
}

func (v BoolVariable) String() string {
	return strconv.FormatBool(v.Value)
}

func (v BoolVariable) Get(path string) (Variable, bool) {
	if path == "" {
		return &v, true
	}
	return nil, false
}

func (v *BoolVariable) Set(path string, value Variable) error {
	if path == "" {
		if boolVar, ok := value.(*BoolVariable); ok {
			v.Value = boolVar.Value
			return nil
		}
		return fmt.Errorf("cannot set bool variable to non-bool value")
	}
	return fmt.Errorf("cannot set nested path on bool variable")
}

func (v BoolVariable) Keys() []string {
	return nil
}

func (v BoolVariable) Len() int {
	return 0
}

func (v BoolVariable) Type() VariableType {
	return VariableTypeBool
}

func (v BoolVariable) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Value)
}

func (v *BoolVariable) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &v.Value)
}

// Helper functions

// parsePath parses a path like "foo.bar[0].baz" into ["foo", "bar[0]", "baz"]
func parsePath(path string) []string {
	if path == "" {
		return nil
	}

	// Handle array indices in brackets
	var parts []string
	var current strings.Builder
	inBrackets := false

	for _, char := range path {
		switch char {
		case '.':
			if !inBrackets {
				if current.Len() > 0 {
					parts = append(parts, current.String())
					current.Reset()
				}
			} else {
				current.WriteRune(char)
			}
		case '[':
			inBrackets = true
			current.WriteRune(char)
		case ']':
			inBrackets = false
			current.WriteRune(char)
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// convertToVariable converts an interface{} to a Variable
func convertToVariable(val interface{}) Variable {
	if val == nil {
		return &StringVariable{Value: ""}
	}

	switch v := val.(type) {
	case string:
		return &StringVariable{Value: v}
	case float64:
		return &NumberVariable{Value: v}
	case int:
		return &NumberVariable{Value: float64(v)}
	case bool:
		return &BoolVariable{Value: v}
	case []interface{}:
		slice := NewSliceVariable()
		for _, item := range v {
			slice.Values = append(slice.Values, convertToVariable(item))
		}
		return slice
	case map[string]interface{}:
		m := NewMapVariable()
		for k, item := range v {
			m.Values[k] = convertToVariable(item)
		}
		return m
	default:
		// Try to convert to string
		return &StringVariable{Value: fmt.Sprintf("%v", v)}
	}
}

// VariableSet represents a collection of variables with complex operations
type VariableSet struct {
	variables map[string]Variable
}

func NewVariableSet() *VariableSet {
	return &VariableSet{
		variables: make(map[string]Variable),
	}
}

func (vs *VariableSet) Set(name string, value Variable) {
	vs.variables[name] = value
}

func (vs *VariableSet) Get(name string) (Variable, bool) {
	val, exists := vs.variables[name]
	return val, exists
}

func (vs *VariableSet) GetString(name string) (string, bool) {
	if val, exists := vs.GetByPath(name); exists {
		return val.String(), true
	}
	return "", false
}

func (vs *VariableSet) GetByPath(path string) (Variable, bool) {
	parts := parsePath(path)
	if len(parts) == 0 {
		return nil, false
	}

	rootName := parts[0]
	if rootVar, exists := vs.Get(rootName); exists {
		if len(parts) == 1 {
			return rootVar, true
		}
		return rootVar.Get(strings.Join(parts[1:], "."))
	}
	return nil, false
}

func (vs *VariableSet) SetByPath(path string, value Variable) error {
	parts := parsePath(path)
	if len(parts) == 0 {
		return fmt.Errorf("empty path")
	}

	rootName := parts[0]
	if len(parts) == 1 {
		vs.Set(rootName, value)
		return nil
	}

	if rootVar, exists := vs.Get(rootName); exists {
		return rootVar.Set(strings.Join(parts[1:], "."), value)
	}

	// Create nested structure
	nested := NewMapVariable()
	if err := nested.Set(strings.Join(parts[1:], "."), value); err != nil {
		return err
	}
	vs.Set(rootName, nested)
	return nil
}

func (vs *VariableSet) Keys() []string {
	keys := make([]string, 0, len(vs.variables))
	for k := range vs.variables {
		keys = append(keys, k)
	}
	return keys
}

func (vs *VariableSet) Len() int {
	return len(vs.variables)
}

// GetVariables returns a copy of the variables map for template processing
func (vs *VariableSet) GetVariables() map[string]Variable {
	result := make(map[string]Variable)
	for k, v := range vs.variables {
		result[k] = v
	}
	return result
}

func (vs *VariableSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(vs.variables)
}

func (vs *VariableSet) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	vs.variables = make(map[string]Variable)
	for k, val := range raw {
		vs.variables[k] = convertToVariable(val)
	}
	return nil
}

// RangeVariables provides iteration over variables
func (vs *VariableSet) Range(fn func(name string, value Variable) bool) {
	for name, value := range vs.variables {
		if !fn(name, value) {
			break
		}
	}
}

// FlattenVariables flattens nested variables into dot-notation paths
func (vs *VariableSet) Flatten() map[string]string {
	result := make(map[string]string)

	var flatten func(prefix string, value Variable)
	flatten = func(prefix string, value Variable) {
		switch v := value.(type) {
		case *StringVariable, *NumberVariable, *BoolVariable:
			if prefix == "" {
				return
			}
			result[prefix] = value.String()
		case *MapVariable:
			for key, val := range v.Values {
				newPrefix := key
				if prefix != "" {
					newPrefix = prefix + "." + key
				}
				flatten(newPrefix, val)
			}
		case *SliceVariable:
			for i, val := range v.Values {
				newPrefix := fmt.Sprintf("%s[%d]", prefix, i)
				flatten(newPrefix, val)
			}
		}
	}

	for name, value := range vs.variables {
		flatten(name, value)
	}

	return result
}
