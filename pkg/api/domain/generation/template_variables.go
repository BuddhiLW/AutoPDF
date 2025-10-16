// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package generation

import (
	"fmt"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/BuddhiLW/AutoPDF/pkg/converter"
)

// TemplateVariables is a Domain Value Object representing variables for template processing
// It wraps config.Variables and provides domain-specific validation and conversion methods
type TemplateVariables struct {
	variables *config.Variables
}

// NewTemplateVariables creates a new TemplateVariables from config.Variables
func NewTemplateVariables(vars *config.Variables) *TemplateVariables {
	if vars == nil {
		vars = config.NewVariables()
	}
	return &TemplateVariables{
		variables: vars,
	}
}

// NewTemplateVariablesFromMap creates TemplateVariables from a map[string]interface{}
// This is used for backward compatibility with existing API endpoints
func NewTemplateVariablesFromMap(m map[string]interface{}) (*TemplateVariables, error) {
	if m == nil {
		return NewTemplateVariables(nil), nil
	}

	vars := config.NewVariables()
	for key, value := range m {
		// Convert interface{} to config.Variable
		var variable config.Variable
		switch v := value.(type) {
		case string:
			variable = &config.StringVariable{Value: v}
		case int:
			variable = &config.NumberVariable{Value: float64(v)}
		case float64:
			variable = &config.NumberVariable{Value: v}
		case bool:
			variable = &config.BoolVariable{Value: v}
		case map[string]interface{}:
			// Recursively handle nested maps
			mapVar := config.NewMapVariable()
			for k, val := range v {
				nestedVar, err := convertInterfaceToVariable(val)
				if err != nil {
					return nil, fmt.Errorf("failed to convert nested value for key %s: %w", k, err)
				}
				mapVar.Values[k] = nestedVar
			}
			variable = mapVar
		case []interface{}:
			// Handle slices
			sliceVar := config.NewSliceVariable()
			for _, item := range v {
				itemVar, err := convertInterfaceToVariable(item)
				if err != nil {
					return nil, fmt.Errorf("failed to convert slice item: %w", err)
				}
				sliceVar.Values = append(sliceVar.Values, itemVar)
			}
			variable = sliceVar
		default:
			// Fallback: convert to string
			variable = &config.StringVariable{Value: fmt.Sprintf("%v", value)}
		}

		vars.Set(key, variable)
	}

	return &TemplateVariables{variables: vars}, nil
}

// NewTemplateVariablesFromStruct creates TemplateVariables from a struct using StructConverter
// This enables type-safe variable creation from domain objects
func NewTemplateVariablesFromStruct(s interface{}, conv *converter.StructConverter) (*TemplateVariables, error) {
	if s == nil {
		return nil, fmt.Errorf("struct cannot be nil")
	}
	if conv == nil {
		return nil, fmt.Errorf("converter cannot be nil")
	}

	vars, err := conv.ConvertStruct(s)
	if err != nil {
		return nil, fmt.Errorf("failed to convert struct to variables: %w", err)
	}

	return &TemplateVariables{variables: vars}, nil
}

// ToMap converts TemplateVariables to map[string]interface{}
// Used for backward compatibility and serialization
func (tv *TemplateVariables) ToMap() map[string]interface{} {
	if tv.variables == nil {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})
	tv.variables.Range(func(name string, value config.Variable) bool {
		result[name] = variableToInterface(value)
		return true
	})
	return result
}

// Flatten converts TemplateVariables to map[string]string for template processing
// This is the main method used by the template processor
func (tv *TemplateVariables) Flatten() map[string]string {
	if tv.variables == nil {
		return make(map[string]string)
	}
	return tv.variables.Flatten()
}

// Get retrieves a variable by name
func (tv *TemplateVariables) Get(key string) (config.Variable, bool) {
	if tv.variables == nil {
		return nil, false
	}
	return tv.variables.Get(key)
}

// GetString retrieves a variable as a string by name
func (tv *TemplateVariables) GetString(key string) (string, bool) {
	if tv.variables == nil {
		return "", false
	}
	return tv.variables.GetString(key)
}

// Set sets a variable by name
func (tv *TemplateVariables) Set(key string, value config.Variable) {
	if tv.variables == nil {
		tv.variables = config.NewVariables()
	}
	tv.variables.Set(key, value)
}

// SetString sets a string variable by name
func (tv *TemplateVariables) SetString(key string, value string) error {
	if tv.variables == nil {
		tv.variables = config.NewVariables()
	}
	return tv.variables.SetString(key, value)
}

// Keys returns all variable names
func (tv *TemplateVariables) Keys() []string {
	if tv.variables == nil {
		return []string{}
	}
	return tv.variables.Keys()
}

// Len returns the number of variables
func (tv *TemplateVariables) Len() int {
	if tv.variables == nil {
		return 0
	}
	return tv.variables.Len()
}

// Validate performs domain-specific validation on the variables
func (tv *TemplateVariables) Validate() error {
	if tv.variables == nil {
		return fmt.Errorf("variables cannot be nil")
	}

	// Domain-specific validation rules can be added here
	// For now, we just ensure variables exist
	if tv.Len() == 0 {
		// Empty variables are allowed, just return nil
		return nil
	}

	// Validate that all variables can be converted to strings
	flattened := tv.Flatten()
	if len(flattened) == 0 && tv.Len() > 0 {
		return fmt.Errorf("failed to flatten variables: no variables produced")
	}

	return nil
}

// IsEmpty returns true if there are no variables
func (tv *TemplateVariables) IsEmpty() bool {
	return tv.variables == nil || tv.Len() == 0
}

// Clone creates a deep copy of the TemplateVariables
func (tv *TemplateVariables) Clone() *TemplateVariables {
	if tv.variables == nil {
		return NewTemplateVariables(nil)
	}

	// Create a new Variables and copy all values
	newVars := config.NewVariables()
	tv.variables.Range(func(name string, value config.Variable) bool {
		newVars.Set(name, value)
		return true
	})

	return &TemplateVariables{variables: newVars}
}

// Merge merges another TemplateVariables into this one
// Variables from other will override existing variables with the same name
func (tv *TemplateVariables) Merge(other *TemplateVariables) {
	if other == nil || other.variables == nil {
		return
	}

	if tv.variables == nil {
		tv.variables = config.NewVariables()
	}

	other.variables.Range(func(name string, value config.Variable) bool {
		tv.variables.Set(name, value)
		return true
	})
}

// Helper functions

// convertInterfaceToVariable converts an interface{} to config.Variable
func convertInterfaceToVariable(value interface{}) (config.Variable, error) {
	if value == nil {
		return &config.StringVariable{Value: ""}, nil
	}

	switch v := value.(type) {
	case string:
		return &config.StringVariable{Value: v}, nil
	case int:
		return &config.NumberVariable{Value: float64(v)}, nil
	case float64:
		return &config.NumberVariable{Value: v}, nil
	case bool:
		return &config.BoolVariable{Value: v}, nil
	case map[string]interface{}:
		mapVar := config.NewMapVariable()
		for k, val := range v {
			nestedVar, err := convertInterfaceToVariable(val)
			if err != nil {
				return nil, fmt.Errorf("failed to convert nested value for key %s: %w", k, err)
			}
			mapVar.Values[k] = nestedVar
		}
		return mapVar, nil
	case []interface{}:
		sliceVar := config.NewSliceVariable()
		for _, item := range v {
			itemVar, err := convertInterfaceToVariable(item)
			if err != nil {
				return nil, fmt.Errorf("failed to convert slice item: %w", err)
			}
			sliceVar.Values = append(sliceVar.Values, itemVar)
		}
		return sliceVar, nil
	default:
		return &config.StringVariable{Value: fmt.Sprintf("%v", value)}, nil
	}
}

// variableToInterface converts a config.Variable to interface{} for serialization
func variableToInterface(v config.Variable) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case *config.StringVariable:
		return val.Value
	case *config.NumberVariable:
		return val.Value
	case *config.BoolVariable:
		return val.Value
	case *config.MapVariable:
		result := make(map[string]interface{})
		for k, nested := range val.Values {
			result[k] = variableToInterface(nested)
		}
		return result
	case *config.SliceVariable:
		result := make([]interface{}, len(val.Values))
		for i, item := range val.Values {
			result[i] = variableToInterface(item)
		}
		return result
	default:
		return val.String()
	}
}
