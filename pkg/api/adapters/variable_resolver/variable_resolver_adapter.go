// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package variable_resolver

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/BuddhiLW/AutoPDF/pkg/api"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// VariableResolverAdapter implements domain.VariableResolver
type VariableResolverAdapter struct {
	config *config.Config
}

// NewVariableResolverAdapter creates a new variable resolver adapter
func NewVariableResolverAdapter(cfg *config.Config) *VariableResolverAdapter {
	return &VariableResolverAdapter{
		config: cfg,
	}
}

// Resolve resolves complex variables to simple key-value pairs
func (vra *VariableResolverAdapter) Resolve(variables map[string]interface{}) (map[string]string, error) {
	result := make(map[string]string)

	for key, value := range variables {
		resolved, err := vra.resolveValue(value)
		if err != nil {
			return nil, domain.VariableResolutionError{
				Code:    domain.ErrCodeVariableInvalid,
				Message: api.ErrVariableResolutionFailed,
				Details: api.NewErrorDetails(api.ErrorCategoryVariable, api.ErrorSeverityHigh).
					AddContext("key", key).
					WithError(err),
			}
		}
		result[key] = resolved
	}

	return result, nil
}

// Flatten flattens nested variables into dot-notation paths
func (vra *VariableResolverAdapter) Flatten(variables map[string]interface{}) map[string]string {
	result := make(map[string]string)

	var flatten func(prefix string, value interface{})
	flatten = func(prefix string, value interface{}) {
		switch v := value.(type) {
		case string, int, int64, float64, bool:
			if prefix != "" {
				result[prefix] = fmt.Sprintf("%v", v)
			}
		case map[string]interface{}:
			for key, val := range v {
				newPrefix := key
				if prefix != "" {
					newPrefix = prefix + "." + key
				}
				flatten(newPrefix, val)
			}
		case []interface{}:
			for i, val := range v {
				newPrefix := fmt.Sprintf("%s[%d]", prefix, i)
				flatten(newPrefix, val)
			}
		default:
			// Try to convert to string
			if prefix != "" {
				result[prefix] = fmt.Sprintf("%v", v)
			}
		}
	}

	for key, value := range variables {
		flatten(key, value)
	}

	return result
}

// Validate validates variables for correctness
func (vra *VariableResolverAdapter) Validate(variables map[string]interface{}) error {
	for key, value := range variables {
		if err := vra.validateValue(key, value); err != nil {
			return domain.VariableResolutionError{
				Code:    domain.ErrCodeVariableInvalid,
				Message: api.ErrVariableValidationFailed,
				Details: api.NewErrorDetails(api.ErrorCategoryVariable, api.ErrorSeverityHigh).
					AddContext("key", key).
					WithError(err),
			}
		}
	}
	return nil
}

// resolveValue resolves a single value to string
func (vra *VariableResolverAdapter) resolveValue(value interface{}) (string, error) {
	if value == nil {
		return "", nil
	}

	// Check if value has ToAutoPDFFormat() method (OCP compliance)
	if formatter, ok := value.(interface{ ToAutoPDFFormat() string }); ok {
		return formatter.ToAutoPDFFormat(), nil
	}

	switch v := value.(type) {
	case string:
		return v, nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(v), nil
	case map[string]interface{}:
		// For complex objects, create a JSON-like representation
		return vra.mapToString(v), nil
	case []interface{}:
		// For arrays, create a comma-separated string
		return vra.sliceToString(v), nil
	default:
		// Try to convert to string
		return fmt.Sprintf("%v", v), nil
	}
}

// mapToString converts a map to a string representation
func (vra *VariableResolverAdapter) mapToString(m map[string]interface{}) string {
	if len(m) == 0 {
		return "{}"
	}

	var parts []string
	for key, value := range m {
		valueStr, _ := vra.resolveValue(value)
		parts = append(parts, fmt.Sprintf("%s: %s", key, valueStr))
	}

	return "{" + strings.Join(parts, ", ") + "}"
}

// sliceToString converts a slice to a string representation
func (vra *VariableResolverAdapter) sliceToString(s []interface{}) string {
	if len(s) == 0 {
		return "[]"
	}

	var parts []string
	for _, value := range s {
		valueStr, _ := vra.resolveValue(value)
		parts = append(parts, valueStr)
	}

	return "[" + strings.Join(parts, ", ") + "]"
}

// validateValue validates a single value
func (vra *VariableResolverAdapter) validateValue(key string, value interface{}) error {
	if key == "" {
		return fmt.Errorf("variable key cannot be empty")
	}

	// Check for valid key format (alphanumeric, underscore, dot)
	if !vra.isValidKey(key) {
		return fmt.Errorf("invalid variable key format: %s", key)
	}

	// Validate value based on type
	switch v := value.(type) {
	case string:
		// Strings are always valid
		return nil
	case int, int64, float64, bool:
		// Numeric and boolean values are valid
		return nil
	case map[string]interface{}:
		// Recursively validate nested maps
		for nestedKey, nestedValue := range v {
			if err := vra.validateValue(nestedKey, nestedValue); err != nil {
				return fmt.Errorf("nested variable %s.%s: %w", key, nestedKey, err)
			}
		}
		return nil
	case []interface{}:
		// Validate array elements
		for i, element := range v {
			if err := vra.validateValue(fmt.Sprintf("%s[%d]", key, i), element); err != nil {
				return fmt.Errorf("array element %s[%d]: %w", key, i, err)
			}
		}
		return nil
	case nil:
		// Nil values are valid (will be converted to empty string)
		return nil
	default:
		// Check if it's a supported type
		if vra.isSupportedType(value) {
			return nil
		}
		return fmt.Errorf("unsupported variable type: %T", value)
	}
}

// isValidKey checks if a key has a valid format
func (vra *VariableResolverAdapter) isValidKey(key string) bool {
	if key == "" {
		return false
	}

	// Allow alphanumeric, underscore, dot, and bracket characters
	for _, char := range key {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '.' || char == '[' || char == ']') {
			return false
		}
	}

	return true
}

// isSupportedType checks if a type is supported for variable resolution
func (vra *VariableResolverAdapter) isSupportedType(value interface{}) bool {
	switch value.(type) {
	case string, int, int64, float64, bool, map[string]interface{}, []interface{}:
		return true
	default:
		// Check if it's a basic type that can be converted to string
		kind := reflect.TypeOf(value).Kind()
		return kind == reflect.String || kind == reflect.Int || kind == reflect.Float64 || kind == reflect.Bool
	}
}
