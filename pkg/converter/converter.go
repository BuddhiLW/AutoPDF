// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package converter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// Converter converts Go values to AutoPDF Variables
type Converter interface {
	Convert(value interface{}) (config.Variable, error)
	CanConvert(value interface{}) bool
}

// AutoPDFFormattable allows custom types to define their own conversion
type AutoPDFFormattable interface {
	ToAutoPDFVariable() (config.Variable, error)
}

// StructConverter is the main converter for structs
type StructConverter struct {
	registry *ConverterRegistry
	options  ConversionOptions
}

// ConversionOptions configures conversion behavior
type ConversionOptions struct {
	DefaultFlatten bool
	TagName        string // "autopdf"
	OmitEmpty      bool
}

// NewStructConverter creates a new StructConverter with default options
func NewStructConverter() *StructConverter {
	return &StructConverter{
		registry: NewConverterRegistry(),
		options: ConversionOptions{
			DefaultFlatten: false,
			TagName:        "autopdf",
			OmitEmpty:      false,
		},
	}
}

// NewStructConverterWithOptions creates a new StructConverter with custom options
func NewStructConverterWithOptions(options ConversionOptions) *StructConverter {
	return &StructConverter{
		registry: NewConverterRegistry(),
		options:  options,
	}
}

// ConvertStruct converts a struct to Variables using reflection
func (sc *StructConverter) ConvertStruct(v interface{}) (*config.Variables, error) {
	if v == nil {
		return config.NewVariables(), nil
	}

	// Check if implements AutoPDFFormattable (highest priority)
	if formattable, ok := v.(AutoPDFFormattable); ok {
		variable, err := formattable.ToAutoPDFVariable()
		if err != nil {
			return nil, fmt.Errorf("custom formatter failed: %w", err)
		}

		// If it's a MapVariable, convert to Variables
		if mapVar, ok := variable.(*config.MapVariable); ok {
			variables := config.NewVariables()
			for key, val := range mapVar.Values {
				variables.Set(key, val)
			}
			return variables, nil
		}

		// For other types, wrap in a single variable
		variables := config.NewVariables()
		variables.Set("value", variable)
		return variables, nil
	}

	// Check custom converters in registry
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if converter, exists := sc.registry.Get(val.Type()); exists {
		variable, err := converter.Convert(v)
		if err != nil {
			return nil, fmt.Errorf("custom converter failed: %w", err)
		}

		// Convert to Variables format
		variables := config.NewVariables()
		if mapVar, ok := variable.(*config.MapVariable); ok {
			for key, val := range mapVar.Values {
				variables.Set(key, val)
			}
		} else {
			variables.Set("value", variable)
		}
		return variables, nil
	}

	// Fall back to reflection-based conversion
	return sc.convertStructByReflection(v)
}

// convertStructByReflection handles the core reflection logic
func (sc *StructConverter) convertStructByReflection(v interface{}) (*config.Variables, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return config.NewVariables(), nil
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", val.Kind())
	}

	typ := val.Type()
	variables := config.NewVariables()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Parse struct tag
		tag := ParseTag(fieldType.Tag.Get(sc.options.TagName))

		// Skip fields marked with "-"
		if tag.Omit {
			continue
		}

		// Handle omitempty
		if (sc.options.OmitEmpty || tag.OmitEmpty) && isEmptyValue(field) {
			continue
		}

		// Determine field name
		fieldName := tag.Name
		if fieldName == "" {
			fieldName = fieldType.Name
		}

		// Convert field value
		variable, err := sc.convertValue(field, tag)
		if err != nil {
			return nil, fmt.Errorf("failed to convert field %s: %w", fieldName, err)
		}

		// Apply flattening or inlining based on tag
		if tag.Flatten {
			err = sc.flattenVariable(variables, fieldName, variable)
		} else if tag.Inline {
			err = sc.inlineVariable(variables, variable)
		} else {
			variables.Set(fieldName, variable)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to apply field %s: %w", fieldName, err)
		}
	}

	return variables, nil
}

// convertValue handles individual value conversion
func (sc *StructConverter) convertValue(val reflect.Value, tag FieldTag) (config.Variable, error) {
	// Handle nil pointers
	if val.Kind() == reflect.Ptr && val.IsNil() {
		return &config.StringVariable{Value: ""}, nil
	}

	// Dereference pointers
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Handle different kinds
	switch val.Kind() {
	case reflect.String:
		return &config.StringVariable{Value: val.String()}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &config.NumberVariable{Value: float64(val.Int())}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &config.NumberVariable{Value: float64(val.Uint())}, nil
	case reflect.Float32, reflect.Float64:
		return &config.NumberVariable{Value: val.Float()}, nil
	case reflect.Bool:
		return &config.BoolVariable{Value: val.Bool()}, nil
	case reflect.Struct:
		return sc.convertStructField(val, tag)
	case reflect.Slice, reflect.Array:
		return sc.convertSliceField(val, tag)
	case reflect.Map:
		return sc.convertMapField(val, tag)
	case reflect.Interface:
		if val.IsNil() {
			return &config.StringVariable{Value: ""}, nil
		}
		return sc.convertValue(val.Elem(), tag)
	default:
		// Try to convert to string as fallback
		return &config.StringVariable{Value: fmt.Sprintf("%v", val.Interface())}, nil
	}
}

// convertStructField converts a struct field
func (sc *StructConverter) convertStructField(val reflect.Value, tag FieldTag) (config.Variable, error) {
	// Check if it implements AutoPDFFormattable
	if formattable, ok := val.Interface().(AutoPDFFormattable); ok {
		return formattable.ToAutoPDFVariable()
	}

	// Check custom converters
	if converter, exists := sc.registry.Get(val.Type()); exists {
		return converter.Convert(val.Interface())
	}

	// Use reflection to convert nested struct
	nestedVars, err := sc.convertStructByReflection(val.Interface())
	if err != nil {
		return nil, err
	}

	// Convert Variables to MapVariable
	mapVar := config.NewMapVariable()
	for key, variable := range nestedVars.GetVariables() {
		mapVar.Set(key, variable)
	}

	return mapVar, nil
}

// convertSliceField converts a slice or array field
func (sc *StructConverter) convertSliceField(val reflect.Value, tag FieldTag) (config.Variable, error) {
	length := val.Len()
	sliceVar := config.NewSliceVariable()
	sliceVar.Values = make([]config.Variable, 0, length)

	for i := 0; i < length; i++ {
		element := val.Index(i)
		elementVar, err := sc.convertValue(element, FieldTag{})
		if err != nil {
			return nil, fmt.Errorf("failed to convert slice element %d: %w", i, err)
		}
		sliceVar.Values = append(sliceVar.Values, elementVar)
	}

	// Apply flattening if requested
	if tag.Flatten || sc.options.DefaultFlatten {
		return sc.flattenSlice(sliceVar, ", "), nil
	}

	return sliceVar, nil
}

// convertMapField converts a map field
func (sc *StructConverter) convertMapField(val reflect.Value, tag FieldTag) (config.Variable, error) {
	mapVar := config.NewMapVariable()

	for _, key := range val.MapKeys() {
		keyStr := fmt.Sprintf("%v", key.Interface())
		value := val.MapIndex(key)

		valueVar, err := sc.convertValue(value, FieldTag{})
		if err != nil {
			return nil, fmt.Errorf("failed to convert map value for key %s: %w", keyStr, err)
		}

		mapVar.Set(keyStr, valueVar)
	}

	return mapVar, nil
}

// flattenVariable flattens a variable using dot notation
func (sc *StructConverter) flattenVariable(variables *config.Variables, prefix string, variable config.Variable) error {
	switch v := variable.(type) {
	case *config.MapVariable:
		for key, val := range v.Values {
			flattenedKey := prefix + "." + key
			err := sc.flattenVariable(variables, flattenedKey, val)
			if err != nil {
				return err
			}
		}
	case *config.SliceVariable:
		for i, val := range v.Values {
			flattenedKey := fmt.Sprintf("%s[%d]", prefix, i)
			err := sc.flattenVariable(variables, flattenedKey, val)
			if err != nil {
				return err
			}
		}
	default:
		variables.Set(prefix, variable)
	}
	return nil
}

// inlineVariable inlines nested struct fields at parent level
func (sc *StructConverter) inlineVariable(variables *config.Variables, variable config.Variable) error {
	switch v := variable.(type) {
	case *config.MapVariable:
		for key, val := range v.Values {
			variables.Set(key, val)
		}
	default:
		// For non-map variables, just set as "value"
		variables.Set("value", variable)
	}
	return nil
}

// flattenSlice converts a SliceVariable to a comma-separated string
func (sc *StructConverter) flattenSlice(sliceVar *config.SliceVariable, separator string) config.Variable {
	if len(sliceVar.Values) == 0 {
		return &config.StringVariable{Value: ""}
	}

	var parts []string
	for _, val := range sliceVar.Values {
		parts = append(parts, val.String())
	}

	return &config.StringVariable{Value: strings.Join(parts, separator)}
}

// isEmptyValue checks if a reflect.Value is empty
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
