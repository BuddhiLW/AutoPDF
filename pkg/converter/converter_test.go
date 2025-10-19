// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package converter

import (
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structs for testing
type BasicStruct struct {
	Name  string `autopdf:"name"`
	Age   int    `autopdf:"age"`
	Email string `autopdf:"email"`
}

type NestedStruct struct {
	User BasicStruct `autopdf:"user"`
	ID   int         `autopdf:"id"`
}

type FlattenStruct struct {
	User BasicStruct `autopdf:"user,flatten"`
	ID   int         `autopdf:"id"`
}

type InlineStruct struct {
	User BasicStruct `autopdf:"user,inline"`
	ID   int         `autopdf:"id"`
}

type OmitStruct struct {
	Name  string `autopdf:"name"`
	Email string `autopdf:"-"`
	Age   int    `autopdf:"age"`
}

type SliceStruct struct {
	Names []string `autopdf:"names"`
	IDs   []int    `autopdf:"ids"`
}

type FlattenSliceStruct struct {
	Names []string `autopdf:"names,flatten"`
	IDs   []int    `autopdf:"ids,flatten"`
}

type TimeStruct struct {
	CreatedAt time.Time  `autopdf:"created_at"`
	UpdatedAt *time.Time `autopdf:"updated_at"`
}

type URLStruct struct {
	Homepage url.URL  `autopdf:"homepage"`
	Profile  *url.URL `autopdf:"profile"`
}

type CustomFormattable struct {
	Value string
}

func (cf CustomFormattable) ToAutoPDFVariable() (config.Variable, error) {
	return &config.StringVariable{Value: "custom:" + cf.Value}, nil
}

type CustomStruct struct {
	Custom CustomFormattable `autopdf:"custom"`
	Normal string            `autopdf:"normal"`
}

func TestParseTag(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected FieldTag
	}{
		{
			name: "simple field name",
			tag:  "name",
			expected: FieldTag{
				Name:    "name",
				Options: []string{},
			},
		},
		{
			name: "field name with omitempty",
			tag:  "name,omitempty",
			expected: FieldTag{
				Name:      "name",
				OmitEmpty: true,
				Options:   []string{"omitempty"},
			},
		},
		{
			name: "field name with flatten",
			tag:  "user,flatten",
			expected: FieldTag{
				Name:    "user",
				Flatten: true,
				Options: []string{"flatten"},
			},
		},
		{
			name: "field name with inline",
			tag:  "user,inline",
			expected: FieldTag{
				Name:    "user",
				Inline:  true,
				Options: []string{"inline"},
			},
		},
		{
			name: "skip field",
			tag:  "-",
			expected: FieldTag{
				Omit: true,
			},
		},
		{
			name: "multiple options",
			tag:  "user,flatten,omitempty",
			expected: FieldTag{
				Name:      "user",
				OmitEmpty: true,
				Flatten:   true,
				Options:   []string{"flatten", "omitempty"},
			},
		},
		{
			name: "empty tag",
			tag:  "",
			expected: FieldTag{
				Options: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseTag(tt.tag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStructConverter_ConvertStruct_Basic(t *testing.T) {
	converter := NewStructConverter()

	basic := BasicStruct{
		Name:  "John Doe",
		Age:   30,
		Email: "john@example.com",
	}

	variables, err := converter.ConvertStruct(basic)
	require.NoError(t, err)

	// Check that all fields are present
	nameVar, exists := variables.Get("name")
	require.True(t, exists)
	assert.Equal(t, "John Doe", nameVar.String())

	ageVar, exists := variables.Get("age")
	require.True(t, exists)
	assert.Equal(t, "30", ageVar.String())

	emailVar, exists := variables.Get("email")
	require.True(t, exists)
	assert.Equal(t, "john@example.com", emailVar.String())
}

func TestStructConverter_ConvertStruct_Nested(t *testing.T) {
	converter := NewStructConverter()

	nested := NestedStruct{
		User: BasicStruct{
			Name:  "Jane Doe",
			Age:   25,
			Email: "jane@example.com",
		},
		ID: 123,
	}

	variables, err := converter.ConvertStruct(nested)
	require.NoError(t, err)

	// Check top-level field
	idVar, exists := variables.Get("id")
	require.True(t, exists)
	assert.Equal(t, "123", idVar.String())

	// Check nested field
	userVar, exists := variables.Get("user")
	require.True(t, exists)
	assert.Equal(t, config.VariableTypeMap, userVar.Type())

	// Check nested user fields
	userMap, ok := userVar.(*config.MapVariable)
	require.True(t, ok)

	nameVar, exists := userMap.Get("name")
	require.True(t, exists)
	assert.Equal(t, "Jane Doe", nameVar.String())
}

func TestStructConverter_ConvertStruct_Flatten(t *testing.T) {
	converter := NewStructConverter()

	flatten := FlattenStruct{
		User: BasicStruct{
			Name:  "Bob Smith",
			Age:   35,
			Email: "bob@example.com",
		},
		ID: 456,
	}

	variables, err := converter.ConvertStruct(flatten)
	require.NoError(t, err)

	// Check that nested fields are flattened
	nameVar, exists := variables.Get("user.name")
	require.True(t, exists)
	assert.Equal(t, "Bob Smith", nameVar.String())

	ageVar, exists := variables.Get("user.age")
	require.True(t, exists)
	assert.Equal(t, "35", ageVar.String())

	emailVar, exists := variables.Get("user.email")
	require.True(t, exists)
	assert.Equal(t, "bob@example.com", emailVar.String())

	// Check top-level field
	idVar, exists := variables.Get("id")
	require.True(t, exists)
	assert.Equal(t, "456", idVar.String())
}

func TestStructConverter_ConvertStruct_Inline(t *testing.T) {
	converter := NewStructConverter()

	inline := InlineStruct{
		User: BasicStruct{
			Name:  "Alice Johnson",
			Age:   28,
			Email: "alice@example.com",
		},
		ID: 789,
	}

	variables, err := converter.ConvertStruct(inline)
	require.NoError(t, err)

	// Check that nested fields are inlined at the same level
	nameVar, exists := variables.Get("name")
	require.True(t, exists)
	assert.Equal(t, "Alice Johnson", nameVar.String())

	ageVar, exists := variables.Get("age")
	require.True(t, exists)
	assert.Equal(t, "28", ageVar.String())

	emailVar, exists := variables.Get("email")
	require.True(t, exists)
	assert.Equal(t, "alice@example.com", emailVar.String())

	// Check top-level field
	idVar, exists := variables.Get("id")
	require.True(t, exists)
	assert.Equal(t, "789", idVar.String())
}

func TestStructConverter_ConvertStruct_Omit(t *testing.T) {
	converter := NewStructConverter()

	omit := OmitStruct{
		Name:  "Charlie Brown",
		Email: "charlie@example.com", // This should be omitted
		Age:   40,
	}

	variables, err := converter.ConvertStruct(omit)
	require.NoError(t, err)

	// Check that included fields are present
	nameVar, exists := variables.Get("name")
	require.True(t, exists)
	assert.Equal(t, "Charlie Brown", nameVar.String())

	ageVar, exists := variables.Get("age")
	require.True(t, exists)
	assert.Equal(t, "40", ageVar.String())

	// Check that omitted field is not present
	_, exists = variables.Get("email")
	assert.False(t, exists)
}

func TestStructConverter_ConvertStruct_Slices(t *testing.T) {
	converter := NewStructConverter()

	slice := SliceStruct{
		Names: []string{"Alice", "Bob", "Charlie"},
		IDs:   []int{1, 2, 3},
	}

	variables, err := converter.ConvertStruct(slice)
	require.NoError(t, err)

	// Check names slice
	namesVar, exists := variables.Get("names")
	require.True(t, exists)
	assert.Equal(t, config.VariableTypeSlice, namesVar.Type())

	namesSlice, ok := namesVar.(*config.SliceVariable)
	require.True(t, ok)
	assert.Len(t, namesSlice.Values, 3)
	assert.Equal(t, "Alice", namesSlice.Values[0].String())
	assert.Equal(t, "Bob", namesSlice.Values[1].String())
	assert.Equal(t, "Charlie", namesSlice.Values[2].String())

	// Check IDs slice
	idsVar, exists := variables.Get("ids")
	require.True(t, exists)
	assert.Equal(t, config.VariableTypeSlice, idsVar.Type())

	idsSlice, ok := idsVar.(*config.SliceVariable)
	require.True(t, ok)
	assert.Len(t, idsSlice.Values, 3)
	assert.Equal(t, "1", idsSlice.Values[0].String())
	assert.Equal(t, "2", idsSlice.Values[1].String())
	assert.Equal(t, "3", idsSlice.Values[2].String())
}

func TestStructConverter_ConvertStruct_FlattenSlices(t *testing.T) {
	converter := NewStructConverter()

	flattenSlice := FlattenSliceStruct{
		Names: []string{"Alice", "Bob", "Charlie"},
		IDs:   []int{1, 2, 3},
	}

	variables, err := converter.ConvertStruct(flattenSlice)
	require.NoError(t, err)

	// Check that slices are flattened to strings
	namesVar, exists := variables.Get("names")
	require.True(t, exists)
	assert.Equal(t, config.VariableTypeString, namesVar.Type())
	assert.Equal(t, "Alice, Bob, Charlie", namesVar.String())

	idsVar, exists := variables.Get("ids")
	require.True(t, exists)
	assert.Equal(t, config.VariableTypeString, idsVar.Type())
	assert.Equal(t, "1, 2, 3", idsVar.String())
}

func TestStructConverter_ConvertStruct_Time(t *testing.T) {
	converter := BuildWithDefaults()

	now := time.Date(2025, 1, 7, 12, 30, 45, 0, time.UTC)
	timeStruct := TimeStruct{
		CreatedAt: now,
		UpdatedAt: &now,
	}

	variables, err := converter.ConvertStruct(timeStruct)
	require.NoError(t, err)

	// Check time fields
	createdVar, exists := variables.Get("created_at")
	require.True(t, exists)
	assert.Equal(t, "2025-01-07 12:30:45", createdVar.String())

	updatedVar, exists := variables.Get("updated_at")
	require.True(t, exists)
	assert.Equal(t, "2025-01-07 12:30:45", updatedVar.String())
}

func TestStructConverter_ConvertStruct_URL(t *testing.T) {
	converter := BuildWithDefaults()

	homepageURL, _ := url.Parse("https://example.com")
	profileURL, _ := url.Parse("https://profile.example.com")
	urlStruct := URLStruct{
		Homepage: *homepageURL,
		Profile:  profileURL,
	}

	variables, err := converter.ConvertStruct(urlStruct)
	require.NoError(t, err)

	// Check URL fields
	homepageVar, exists := variables.Get("homepage")
	require.True(t, exists)
	assert.Equal(t, "https://example.com", homepageVar.String())

	profileVar, exists := variables.Get("profile")
	require.True(t, exists)
	assert.Equal(t, "https://profile.example.com", profileVar.String())
}

func TestStructConverter_ConvertStruct_CustomFormattable(t *testing.T) {
	converter := NewStructConverter()

	custom := CustomStruct{
		Custom: CustomFormattable{Value: "test"},
		Normal: "normal",
	}

	variables, err := converter.ConvertStruct(custom)
	require.NoError(t, err)

	// Check custom formattable field
	customVar, exists := variables.Get("custom")
	require.True(t, exists)
	assert.Equal(t, "custom:test", customVar.String())

	// Check normal field
	normalVar, exists := variables.Get("normal")
	require.True(t, exists)
	assert.Equal(t, "normal", normalVar.String())
}

func TestStructConverter_ConvertStruct_NilPointer(t *testing.T) {
	converter := NewStructConverter()

	type PointerStruct struct {
		Name *string `autopdf:"name"`
		Age  *int    `autopdf:"age"`
	}

	ptrStruct := PointerStruct{
		Name: nil,
		Age:  nil,
	}

	variables, err := converter.ConvertStruct(ptrStruct)
	require.NoError(t, err)

	// Check that nil pointers are converted to empty strings
	nameVar, exists := variables.Get("name")
	require.True(t, exists)
	assert.Equal(t, "", nameVar.String())

	ageVar, exists := variables.Get("age")
	require.True(t, exists)
	assert.Equal(t, "", ageVar.String())
}

func TestStructConverter_ConvertStruct_EmptySlice(t *testing.T) {
	converter := NewStructConverter()

	emptySlice := SliceStruct{
		Names: []string{},
		IDs:   []int{},
	}

	variables, err := converter.ConvertStruct(emptySlice)
	require.NoError(t, err)

	// Check empty slices
	namesVar, exists := variables.Get("names")
	require.True(t, exists)
	assert.Equal(t, config.VariableTypeSlice, namesVar.Type())

	namesSlice, ok := namesVar.(*config.SliceVariable)
	require.True(t, ok)
	assert.Len(t, namesSlice.Values, 0)
}

func TestStructConverter_ConvertStruct_InvalidInput(t *testing.T) {
	converter := NewStructConverter()

	// Test with non-struct input
	_, err := converter.ConvertStruct("not a struct")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected struct")

	// Test with nil input
	variables, err := converter.ConvertStruct(nil)
	require.NoError(t, err)
	assert.NotNil(t, variables)
}

func TestConverterRegistry(t *testing.T) {
	registry := NewConverterRegistry()

	// Test initial state
	assert.Equal(t, 0, registry.Count())
	assert.False(t, registry.Has(reflect.TypeOf(time.Time{})))

	// Register a converter
	timeType := reflect.TypeOf(time.Time{})
	timeConverter := NewTimeConverter()
	err := registry.Register(timeType, timeConverter)
	require.NoError(t, err)

	// Test registration
	assert.Equal(t, 1, registry.Count())
	assert.True(t, registry.Has(timeType))

	// Test retrieval
	retrieved, exists := registry.Get(timeType)
	require.True(t, exists)
	assert.Equal(t, timeConverter, retrieved)

	// Test unregistration
	registry.Unregister(timeType)
	assert.Equal(t, 0, registry.Count())
	assert.False(t, registry.Has(timeType))
}

func TestBuiltinConverters(t *testing.T) {
	registry := NewConverterRegistry()
	err := RegisterBuiltinConverters(registry)
	require.NoError(t, err)

	// Test that built-in types are registered
	builtinTypes := GetBuiltinConverterTypes()
	for _, typ := range builtinTypes {
		assert.True(t, registry.Has(typ), "Type %v should be registered", typ)
	}
}

func TestConverterBuilder(t *testing.T) {
	builder := NewConverterBuilder()
	converter := builder.
		WithBuiltinConverters().
		WithTimeFormat("2006-01-02").
		WithDurationFormat("seconds").
		WithSliceSeparator(" | ").
		Build()

	assert.NotNil(t, converter)
	assert.NotNil(t, converter.registry)
	assert.Equal(t, "autopdf", converter.options.TagName)
}

func TestBuildWithDefaults(t *testing.T) {
	converter := BuildWithDefaults()
	assert.NotNil(t, converter)
	assert.NotNil(t, converter.registry)
}

func TestBuildForTemplates(t *testing.T) {
	converter := BuildForTemplates()
	assert.NotNil(t, converter)
	assert.True(t, converter.options.OmitEmpty)
	assert.False(t, converter.options.DefaultFlatten)
}

func TestBuildForFlattened(t *testing.T) {
	converter := BuildForFlattened()
	assert.NotNil(t, converter)
	assert.True(t, converter.options.OmitEmpty)
	assert.True(t, converter.options.DefaultFlatten)
}
