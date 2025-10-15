// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package variable_resolver

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/pkg/api"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/BuddhiLW/AutoPDF/pkg/converter"
)

// VariableResolverAdapter implements domain.VariableResolver
type VariableResolverAdapter struct {
	config    *config.Config
	logger    *logger.LoggerAdapter
	converter *converter.StructConverter
}

// NewVariableResolverAdapter creates a new variable resolver adapter
func NewVariableResolverAdapter(cfg *config.Config, logger *logger.LoggerAdapter) *VariableResolverAdapter {
	return &VariableResolverAdapter{
		config:    cfg,
		logger:    logger,
		converter: converter.BuildWithDefaults(),
	}
}

// NewVariableResolverAdapterWithConverter creates a new variable resolver adapter with custom converter
func NewVariableResolverAdapterWithConverter(cfg *config.Config, logger *logger.LoggerAdapter, conv *converter.StructConverter) *VariableResolverAdapter {
	return &VariableResolverAdapter{
		config:    cfg,
		logger:    logger,
		converter: conv,
	}
}

// Resolve resolves complex variables to simple key-value pairs
func (vra *VariableResolverAdapter) Resolve(variables map[string]interface{}) (map[string]string, error) {
	vra.logger.DebugWithFields("Starting variable resolution",
		"variable_count", len(variables),
		"variable_keys", vra.getMapKeys(variables),
	)

	result := make(map[string]string)

	for key, value := range variables {
		resolved, err := vra.resolveValue(value)
		if err != nil {
			vra.logger.ErrorWithFields("Failed to resolve variable",
				"key", key,
				"value_type", fmt.Sprintf("%T", value),
				"error", err,
			)
			return nil, domain.VariableResolutionError{
				Code:    domain.ErrCodeVariableInvalid,
				Message: api.ErrVariableResolutionFailed,
				Details: api.NewErrorDetails(api.ErrorCategoryVariable, api.ErrorSeverityHigh).
					AddContext("key", key).
					WithError(err),
			}
		}

		vra.logger.DebugWithFields("Resolved variable",
			"key", key,
			"original_type", fmt.Sprintf("%T", value),
			"resolved_value", resolved,
		)

		result[key] = resolved
	}

	vra.logger.InfoWithFields("Variable resolution complete",
		"input_count", len(variables),
		"output_count", len(result),
	)

	return result, nil
}

// getMapKeys returns the keys of a map as a slice
func (vra *VariableResolverAdapter) getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
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

// ConvertStruct converts a Go struct to map[string]interface{} using the struct converter
func (vra *VariableResolverAdapter) ConvertStruct(v interface{}) (map[string]interface{}, error) {
	vra.logger.DebugWithFields("Starting struct conversion",
		"struct_type", fmt.Sprintf("%T", v),
	)

	// Convert struct to Variables using the converter
	variables, err := vra.converter.ConvertStruct(v)
	if err != nil {
		vra.logger.ErrorWithFields("Failed to convert struct",
			"struct_type", fmt.Sprintf("%T", v),
			"error", err,
		)
		return nil, domain.VariableResolutionError{
			Code:    domain.ErrCodeVariableInvalid,
			Message: api.ErrVariableResolutionFailed,
			Details: api.NewErrorDetails(api.ErrorCategoryVariable, api.ErrorSeverityHigh).
				AddContext("struct_type", fmt.Sprintf("%T", v)).
				WithError(err),
		}
	}

	// Convert Variables to map[string]interface{} for existing pipeline
	result := vra.variablesToMap(variables)

	vra.logger.InfoWithFields("Struct conversion complete",
		"struct_type", fmt.Sprintf("%T", v),
		"output_count", len(result),
	)

	return result, nil
}

// variablesToMap converts config.Variables to map[string]interface{}
func (vra *VariableResolverAdapter) variablesToMap(variables *config.Variables) map[string]interface{} {
	result := make(map[string]interface{})

	variables.Range(func(name string, value config.Variable) bool {
		result[name] = vra.variableToInterface(value)
		return true
	})

	return result
}

// variableToInterface converts a config.Variable to interface{}
func (vra *VariableResolverAdapter) variableToInterface(variable config.Variable) interface{} {
	switch v := variable.(type) {
	case *config.StringVariable:
		return v.Value
	case *config.NumberVariable:
		return v.Value
	case *config.BoolVariable:
		return v.Value
	case *config.MapVariable:
		result := make(map[string]interface{})
		for key, val := range v.Values {
			result[key] = vra.variableToInterface(val)
		}
		return result
	case *config.SliceVariable:
		result := make([]interface{}, len(v.Values))
		for i, val := range v.Values {
			result[i] = vra.variableToInterface(val)
		}
		return result
	default:
		return variable.String()
	}
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
