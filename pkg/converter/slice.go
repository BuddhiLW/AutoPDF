// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package converter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// SliceConverter converts slices and arrays
type SliceConverter struct {
	elementConverter Converter
	flatten          bool
	separator        string // Default: ", "
}

// NewSliceConverter creates a new SliceConverter
func NewSliceConverter() *SliceConverter {
	return &SliceConverter{
		flatten:   false,
		separator: ", ",
	}
}

// NewSliceConverterWithOptions creates a new SliceConverter with options
func NewSliceConverterWithOptions(flatten bool, separator string) *SliceConverter {
	return &SliceConverter{
		flatten:   flatten,
		separator: separator,
	}
}

// SetElementConverter sets the converter for individual elements
func (sc *SliceConverter) SetElementConverter(converter Converter) {
	sc.elementConverter = converter
}

// Convert creates SliceVariable or flattened string based on options
func (sc *SliceConverter) Convert(value interface{}) (config.Variable, error) {
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil, fmt.Errorf("expected slice or array, got %s", val.Kind())
	}

	length := val.Len()
	sliceVar := config.NewSliceVariable()
	sliceVar.Values = make([]config.Variable, 0, length)

	for i := 0; i < length; i++ {
		element := val.Index(i)

		var elementVar config.Variable
		var err error

		// Use custom element converter if available
		if sc.elementConverter != nil && sc.elementConverter.CanConvert(element.Interface()) {
			elementVar, err = sc.elementConverter.Convert(element.Interface())
		} else {
			// Use default conversion
			elementVar, err = sc.convertElement(element)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to convert slice element %d: %w", i, err)
		}

		sliceVar.Values = append(sliceVar.Values, elementVar)
	}

	// Apply flattening if requested
	if sc.flatten {
		return sc.flattenSlice(sliceVar), nil
	}

	return sliceVar, nil
}

// CanConvert checks if the value is a slice or array
func (sc *SliceConverter) CanConvert(value interface{}) bool {
	val := reflect.ValueOf(value)
	return val.Kind() == reflect.Slice || val.Kind() == reflect.Array
}

// convertElement converts a single slice element
func (sc *SliceConverter) convertElement(element reflect.Value) (config.Variable, error) {
	// Handle nil pointers
	if element.Kind() == reflect.Ptr && element.IsNil() {
		return &config.StringVariable{Value: ""}, nil
	}

	// Dereference pointers
	if element.Kind() == reflect.Ptr {
		element = element.Elem()
	}

	// Handle different kinds
	switch element.Kind() {
	case reflect.String:
		return &config.StringVariable{Value: element.String()}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &config.NumberVariable{Value: float64(element.Int())}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &config.NumberVariable{Value: float64(element.Uint())}, nil
	case reflect.Float32, reflect.Float64:
		return &config.NumberVariable{Value: element.Float()}, nil
	case reflect.Bool:
		return &config.BoolVariable{Value: element.Bool()}, nil
	case reflect.Struct:
		// For structs, create a simple string representation
		return &config.StringVariable{Value: fmt.Sprintf("%v", element.Interface())}, nil
	case reflect.Slice, reflect.Array:
		// For nested slices, create a simple string representation
		return &config.StringVariable{Value: fmt.Sprintf("%v", element.Interface())}, nil
	case reflect.Map:
		// For maps, create a simple string representation
		return &config.StringVariable{Value: fmt.Sprintf("%v", element.Interface())}, nil
	case reflect.Interface:
		if element.IsNil() {
			return &config.StringVariable{Value: ""}, nil
		}
		return sc.convertElement(element.Elem())
	default:
		// Try to convert to string as fallback
		return &config.StringVariable{Value: fmt.Sprintf("%v", element.Interface())}, nil
	}
}

// flattenSlice converts a SliceVariable to a comma-separated string
func (sc *SliceConverter) flattenSlice(sliceVar *config.SliceVariable) config.Variable {
	if len(sliceVar.Values) == 0 {
		return &config.StringVariable{Value: ""}
	}

	var parts []string
	for _, val := range sliceVar.Values {
		parts = append(parts, val.String())
	}

	return &config.StringVariable{Value: strings.Join(parts, sc.separator)}
}

// StringSliceConverter is a specialized converter for []string
type StringSliceConverter struct {
	separator string
}

// NewStringSliceConverter creates a new StringSliceConverter
func NewStringSliceConverter() *StringSliceConverter {
	return &StringSliceConverter{
		separator: ", ",
	}
}

// NewStringSliceConverterWithSeparator creates a new StringSliceConverter with custom separator
func NewStringSliceConverterWithSeparator(separator string) *StringSliceConverter {
	return &StringSliceConverter{
		separator: separator,
	}
}

// Convert converts []string to StringVariable (flattened) or SliceVariable
func (ssc *StringSliceConverter) Convert(value interface{}) (config.Variable, error) {
	stringSlice, ok := value.([]string)
	if !ok {
		return nil, fmt.Errorf("expected []string, got %T", value)
	}

	if len(stringSlice) == 0 {
		return &config.StringVariable{Value: ""}, nil
	}

	// Always flatten string slices to comma-separated strings
	return &config.StringVariable{Value: strings.Join(stringSlice, ssc.separator)}, nil
}

// CanConvert checks if the value is a []string
func (ssc *StringSliceConverter) CanConvert(value interface{}) bool {
	_, ok := value.([]string)
	return ok
}

// IntSliceConverter is a specialized converter for []int
type IntSliceConverter struct {
	separator string
}

// NewIntSliceConverter creates a new IntSliceConverter
func NewIntSliceConverter() *IntSliceConverter {
	return &IntSliceConverter{
		separator: ", ",
	}
}

// NewIntSliceConverterWithSeparator creates a new IntSliceConverter with custom separator
func NewIntSliceConverterWithSeparator(separator string) *IntSliceConverter {
	return &IntSliceConverter{
		separator: separator,
	}
}

// Convert converts []int to StringVariable (flattened) or SliceVariable
func (isc *IntSliceConverter) Convert(value interface{}) (config.Variable, error) {
	intSlice, ok := value.([]int)
	if !ok {
		return nil, fmt.Errorf("expected []int, got %T", value)
	}

	if len(intSlice) == 0 {
		return &config.StringVariable{Value: ""}, nil
	}

	// Convert to string slice
	var parts []string
	for _, val := range intSlice {
		parts = append(parts, fmt.Sprintf("%d", val))
	}

	return &config.StringVariable{Value: strings.Join(parts, isc.separator)}, nil
}

// CanConvert checks if the value is a []int
func (isc *IntSliceConverter) CanConvert(value interface{}) bool {
	_, ok := value.([]int)
	return ok
}

// FloatSliceConverter is a specialized converter for []float64
type FloatSliceConverter struct {
	separator string
}

// NewFloatSliceConverter creates a new FloatSliceConverter
func NewFloatSliceConverter() *FloatSliceConverter {
	return &FloatSliceConverter{
		separator: ", ",
	}
}

// NewFloatSliceConverterWithSeparator creates a new FloatSliceConverter with custom separator
func NewFloatSliceConverterWithSeparator(separator string) *FloatSliceConverter {
	return &FloatSliceConverter{
		separator: separator,
	}
}

// Convert converts []float64 to StringVariable (flattened) or SliceVariable
func (fsc *FloatSliceConverter) Convert(value interface{}) (config.Variable, error) {
	floatSlice, ok := value.([]float64)
	if !ok {
		return nil, fmt.Errorf("expected []float64, got %T", value)
	}

	if len(floatSlice) == 0 {
		return &config.StringVariable{Value: ""}, nil
	}

	// Convert to string slice
	var parts []string
	for _, val := range floatSlice {
		parts = append(parts, fmt.Sprintf("%g", val))
	}

	return &config.StringVariable{Value: strings.Join(parts, fsc.separator)}, nil
}

// CanConvert checks if the value is a []float64
func (fsc *FloatSliceConverter) CanConvert(value interface{}) bool {
	_, ok := value.([]float64)
	return ok
}

