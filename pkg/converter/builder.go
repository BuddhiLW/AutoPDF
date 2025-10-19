// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package converter

import (
	"reflect"
	"time"
)

// ConverterBuilder provides a fluent API for converter construction
type ConverterBuilder struct {
	registry *ConverterRegistry
	options  ConversionOptions
}

// NewConverterBuilder creates a new ConverterBuilder with default options
func NewConverterBuilder() *ConverterBuilder {
	return &ConverterBuilder{
		registry: NewConverterRegistry(),
		options: ConversionOptions{
			DefaultFlatten: false,
			TagName:        "autopdf",
			OmitEmpty:      false,
		},
	}
}

// WithRegistry sets the converter registry
func (b *ConverterBuilder) WithRegistry(r *ConverterRegistry) *ConverterBuilder {
	b.registry = r
	return b
}

// WithTagName sets the struct tag name to use
func (b *ConverterBuilder) WithTagName(name string) *ConverterBuilder {
	b.options.TagName = name
	return b
}

// WithDefaultFlatten sets the default flattening behavior
func (b *ConverterBuilder) WithDefaultFlatten(flatten bool) *ConverterBuilder {
	b.options.DefaultFlatten = flatten
	return b
}

// WithOmitEmpty sets the omit empty behavior
func (b *ConverterBuilder) WithOmitEmpty(omit bool) *ConverterBuilder {
	b.options.OmitEmpty = omit
	return b
}

// WithBuiltinConverters registers all built-in converters
func (b *ConverterBuilder) WithBuiltinConverters() *ConverterBuilder {
	// Register built-in converters
	RegisterBuiltinConverters(b.registry)
	return b
}

// WithCustomConverter registers a custom converter for a specific type
func (b *ConverterBuilder) WithCustomConverter(typ interface{}, converter Converter) *ConverterBuilder {
	// This will be implemented when we have the reflect package imported
	// For now, we'll add a method that takes the type as a parameter
	return b
}

// WithTimeFormat sets the default time format for time.Time converters
func (b *ConverterBuilder) WithTimeFormat(format string) *ConverterBuilder {
	// Register time converters with custom format
	timeType := reflect.TypeOf(time.Time{})
	timePtrType := reflect.TypeOf((*time.Time)(nil))

	b.registry.Register(timeType, NewTimeConverterWithFormat(format))
	b.registry.Register(timePtrType, NewTimePtrConverterWithFormat(format))

	return b
}

// WithDurationFormat sets the default duration format for time.Duration converters
func (b *ConverterBuilder) WithDurationFormat(format string) *ConverterBuilder {
	// Register duration converters with custom format
	durationType := reflect.TypeOf(time.Duration(0))
	durationPtrType := reflect.TypeOf((*time.Duration)(nil))

	b.registry.Register(durationType, NewDurationConverterWithFormat(format))
	b.registry.Register(durationPtrType, NewDurationPtrConverterWithFormat(format))

	return b
}

// WithSliceSeparator sets the default separator for slice flattening
func (b *ConverterBuilder) WithSliceSeparator(separator string) *ConverterBuilder {
	// Register slice converters with custom separator
	stringSliceType := reflect.TypeOf([]string{})
	intSliceType := reflect.TypeOf([]int{})
	floatSliceType := reflect.TypeOf([]float64{})

	b.registry.Register(stringSliceType, NewStringSliceConverterWithSeparator(separator))
	b.registry.Register(intSliceType, NewIntSliceConverterWithSeparator(separator))
	b.registry.Register(floatSliceType, NewFloatSliceConverterWithSeparator(separator))

	return b
}

// Build creates the StructConverter with the configured options
func (b *ConverterBuilder) Build() *StructConverter {
	return &StructConverter{
		registry: b.registry,
		options:  b.options,
	}
}

// BuildWithDefaults creates a StructConverter with sensible defaults
func BuildWithDefaults() *StructConverter {
	return NewConverterBuilder().
		WithBuiltinConverters().
		WithTimeFormat("2006-01-02 15:04:05").
		WithDurationFormat("string").
		WithSliceSeparator(", ").
		Build()
}

// BuildForTemplates creates a StructConverter optimized for template usage
func BuildForTemplates() *StructConverter {
	return NewConverterBuilder().
		WithBuiltinConverters().
		WithTimeFormat("2006-01-02").
		WithDurationFormat("string").
		WithSliceSeparator(", ").
		WithDefaultFlatten(false).
		WithOmitEmpty(true).
		Build()
}

// BuildForFlattened creates a StructConverter that flattens nested structures
func BuildForFlattened() *StructConverter {
	return NewConverterBuilder().
		WithBuiltinConverters().
		WithTimeFormat("2006-01-02 15:04:05").
		WithDurationFormat("string").
		WithSliceSeparator(", ").
		WithDefaultFlatten(true).
		WithOmitEmpty(true).
		Build()
}
