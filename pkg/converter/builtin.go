// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package converter

import (
	"fmt"
	"net/url"
	"reflect"
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// TimeConverter converts time.Time to string
type TimeConverter struct {
	Format string // Default: RFC3339
}

// NewTimeConverter creates a new TimeConverter with default format
func NewTimeConverter() *TimeConverter {
	return &TimeConverter{
		Format: time.RFC3339,
	}
}

// NewTimeConverterWithFormat creates a new TimeConverter with custom format
func NewTimeConverterWithFormat(format string) *TimeConverter {
	return &TimeConverter{
		Format: format,
	}
}

// Convert converts time.Time to StringVariable
func (tc *TimeConverter) Convert(value interface{}) (config.Variable, error) {
	timeVal, ok := value.(time.Time)
	if !ok {
		return nil, fmt.Errorf("expected time.Time, got %T", value)
	}

	return &config.StringVariable{Value: timeVal.Format(tc.Format)}, nil
}

// CanConvert checks if the value is a time.Time
func (tc *TimeConverter) CanConvert(value interface{}) bool {
	_, ok := value.(time.Time)
	return ok
}

// DurationConverter converts time.Duration to string
type DurationConverter struct {
	Format string // "seconds", "milliseconds", "string"
}

// NewDurationConverter creates a new DurationConverter with default format
func NewDurationConverter() *DurationConverter {
	return &DurationConverter{
		Format: "string",
	}
}

// NewDurationConverterWithFormat creates a new DurationConverter with custom format
func NewDurationConverterWithFormat(format string) *DurationConverter {
	return &DurationConverter{
		Format: format,
	}
}

// Convert converts time.Duration to StringVariable or NumberVariable
func (dc *DurationConverter) Convert(value interface{}) (config.Variable, error) {
	duration, ok := value.(time.Duration)
	if !ok {
		return nil, fmt.Errorf("expected time.Duration, got %T", value)
	}

	switch dc.Format {
	case "seconds":
		return &config.NumberVariable{Value: duration.Seconds()}, nil
	case "milliseconds":
		return &config.NumberVariable{Value: float64(duration.Milliseconds())}, nil
	case "nanoseconds":
		return &config.NumberVariable{Value: float64(duration.Nanoseconds())}, nil
	case "string":
		return &config.StringVariable{Value: duration.String()}, nil
	default:
		return &config.StringVariable{Value: duration.String()}, nil
	}
}

// CanConvert checks if the value is a time.Duration
func (dc *DurationConverter) CanConvert(value interface{}) bool {
	_, ok := value.(time.Duration)
	return ok
}

// URLConverter converts url.URL to string
type URLConverter struct{}

// NewURLConverter creates a new URLConverter
func NewURLConverter() *URLConverter {
	return &URLConverter{}
}

// Convert converts url.URL to StringVariable
func (uc *URLConverter) Convert(value interface{}) (config.Variable, error) {
	urlVal, ok := value.(url.URL)
	if !ok {
		return nil, fmt.Errorf("expected url.URL, got %T", value)
	}

	return &config.StringVariable{Value: urlVal.String()}, nil
}

// CanConvert checks if the value is a url.URL
func (uc *URLConverter) CanConvert(value interface{}) bool {
	_, ok := value.(url.URL)
	return ok
}

// URLPtrConverter converts *url.URL to string
type URLPtrConverter struct{}

// NewURLPtrConverter creates a new URLPtrConverter
func NewURLPtrConverter() *URLPtrConverter {
	return &URLPtrConverter{}
}

// Convert converts *url.URL to StringVariable
func (upc *URLPtrConverter) Convert(value interface{}) (config.Variable, error) {
	urlPtr, ok := value.(*url.URL)
	if !ok {
		return nil, fmt.Errorf("expected *url.URL, got %T", value)
	}

	if urlPtr == nil {
		return &config.StringVariable{Value: ""}, nil
	}

	return &config.StringVariable{Value: urlPtr.String()}, nil
}

// CanConvert checks if the value is a *url.URL
func (upc *URLPtrConverter) CanConvert(value interface{}) bool {
	_, ok := value.(*url.URL)
	return ok
}

// TimePtrConverter converts *time.Time to string
type TimePtrConverter struct {
	Format string // Default: RFC3339
}

// NewTimePtrConverter creates a new TimePtrConverter with default format
func NewTimePtrConverter() *TimePtrConverter {
	return &TimePtrConverter{
		Format: time.RFC3339,
	}
}

// NewTimePtrConverterWithFormat creates a new TimePtrConverter with custom format
func NewTimePtrConverterWithFormat(format string) *TimePtrConverter {
	return &TimePtrConverter{
		Format: format,
	}
}

// Convert converts *time.Time to StringVariable
func (tpc *TimePtrConverter) Convert(value interface{}) (config.Variable, error) {
	timePtr, ok := value.(*time.Time)
	if !ok {
		return nil, fmt.Errorf("expected *time.Time, got %T", value)
	}

	if timePtr == nil {
		return &config.StringVariable{Value: ""}, nil
	}

	return &config.StringVariable{Value: timePtr.Format(tpc.Format)}, nil
}

// CanConvert checks if the value is a *time.Time
func (tpc *TimePtrConverter) CanConvert(value interface{}) bool {
	_, ok := value.(*time.Time)
	return ok
}

// DurationPtrConverter converts *time.Duration to string
type DurationPtrConverter struct {
	Format string // "seconds", "milliseconds", "string"
}

// NewDurationPtrConverter creates a new DurationPtrConverter with default format
func NewDurationPtrConverter() *DurationPtrConverter {
	return &DurationPtrConverter{
		Format: "string",
	}
}

// NewDurationPtrConverterWithFormat creates a new DurationPtrConverter with custom format
func NewDurationPtrConverterWithFormat(format string) *DurationPtrConverter {
	return &DurationPtrConverter{
		Format: format,
	}
}

// Convert converts *time.Duration to StringVariable or NumberVariable
func (dpc *DurationPtrConverter) Convert(value interface{}) (config.Variable, error) {
	durationPtr, ok := value.(*time.Duration)
	if !ok {
		return nil, fmt.Errorf("expected *time.Duration, got %T", value)
	}

	if durationPtr == nil {
		return &config.StringVariable{Value: ""}, nil
	}

	duration := *durationPtr
	switch dpc.Format {
	case "seconds":
		return &config.NumberVariable{Value: duration.Seconds()}, nil
	case "milliseconds":
		return &config.NumberVariable{Value: float64(duration.Milliseconds())}, nil
	case "nanoseconds":
		return &config.NumberVariable{Value: float64(duration.Nanoseconds())}, nil
	case "string":
		return &config.StringVariable{Value: duration.String()}, nil
	default:
		return &config.StringVariable{Value: duration.String()}, nil
	}
}

// CanConvert checks if the value is a *time.Duration
func (dpc *DurationPtrConverter) CanConvert(value interface{}) bool {
	_, ok := value.(*time.Duration)
	return ok
}

// RegisterBuiltinConverters registers all built-in converters
func RegisterBuiltinConverters(registry *ConverterRegistry) error {
	// Register time.Time converter
	timeType := reflect.TypeOf(time.Time{})
	if err := registry.Register(timeType, NewTimeConverter()); err != nil {
		return fmt.Errorf("failed to register time.Time converter: %w", err)
	}

	// Register *time.Time converter
	timePtrType := reflect.TypeOf((*time.Time)(nil))
	if err := registry.Register(timePtrType, NewTimePtrConverter()); err != nil {
		return fmt.Errorf("failed to register *time.Time converter: %w", err)
	}

	// Register time.Duration converter
	durationType := reflect.TypeOf(time.Duration(0))
	if err := registry.Register(durationType, NewDurationConverter()); err != nil {
		return fmt.Errorf("failed to register time.Duration converter: %w", err)
	}

	// Register *time.Duration converter
	durationPtrType := reflect.TypeOf((*time.Duration)(nil))
	if err := registry.Register(durationPtrType, NewDurationPtrConverter()); err != nil {
		return fmt.Errorf("failed to register *time.Duration converter: %w", err)
	}

	// Register url.URL converter
	urlType := reflect.TypeOf(url.URL{})
	if err := registry.Register(urlType, NewURLConverter()); err != nil {
		return fmt.Errorf("failed to register url.URL converter: %w", err)
	}

	// Register *url.URL converter
	urlPtrType := reflect.TypeOf((*url.URL)(nil))
	if err := registry.Register(urlPtrType, NewURLPtrConverter()); err != nil {
		return fmt.Errorf("failed to register *url.URL converter: %w", err)
	}

	return nil
}

// GetBuiltinConverterTypes returns a list of all built-in converter types
func GetBuiltinConverterTypes() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf(time.Time{}),
		reflect.TypeOf((*time.Time)(nil)),
		reflect.TypeOf(time.Duration(0)),
		reflect.TypeOf((*time.Duration)(nil)),
		reflect.TypeOf(url.URL{}),
		reflect.TypeOf((*url.URL)(nil)),
	}
}
