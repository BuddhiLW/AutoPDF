// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package generation_test

import (
	"testing"

	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/BuddhiLW/AutoPDF/pkg/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTemplateVariables(t *testing.T) {
	t.Run("with nil variables", func(t *testing.T) {
		tv := generation.NewTemplateVariables(nil)
		assert.NotNil(t, tv)
		assert.Equal(t, 0, tv.Len())
		assert.True(t, tv.IsEmpty())
	})

	t.Run("with existing variables", func(t *testing.T) {
		vars := config.NewVariables()
		vars.SetString("key1", "value1")

		tv := generation.NewTemplateVariables(vars)
		assert.NotNil(t, tv)
		assert.Equal(t, 1, tv.Len())
		assert.False(t, tv.IsEmpty())
	})
}

func TestNewTemplateVariablesFromMap(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]interface{}
		wantLen int
		wantErr bool
	}{
		{
			name:    "nil map",
			input:   nil,
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "empty map",
			input:   map[string]interface{}{},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "simple string variables",
			input: map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "mixed types",
			input: map[string]interface{}{
				"name":   "John Doe",
				"age":    30,
				"score":  95.5,
				"active": true,
			},
			wantLen: 4,
			wantErr: false,
		},
		{
			name: "nested map",
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"name":  "John",
					"email": "john@example.com",
				},
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "array values",
			input: map[string]interface{}{
				"tags": []interface{}{"tag1", "tag2", "tag3"},
			},
			wantLen: 1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tv, err := generation.NewTemplateVariablesFromMap(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, tv)
			assert.Equal(t, tt.wantLen, tv.Len())
		})
	}
}

func TestNewTemplateVariablesFromStruct(t *testing.T) {
	type TestStruct struct {
		Name  string `autopdf:"name"`
		Email string `autopdf:"email"`
		Age   int    `autopdf:"age"`
	}

	t.Run("valid struct", func(t *testing.T) {
		conv := converter.BuildWithDefaults()
		data := TestStruct{
			Name:  "John Doe",
			Email: "john@example.com",
			Age:   30,
		}

		tv, err := generation.NewTemplateVariablesFromStruct(data, conv)
		require.NoError(t, err)
		assert.NotNil(t, tv)
		assert.Greater(t, tv.Len(), 0)
	})

	t.Run("nil struct", func(t *testing.T) {
		conv := converter.BuildWithDefaults()

		tv, err := generation.NewTemplateVariablesFromStruct(nil, conv)
		assert.Error(t, err)
		assert.Nil(t, tv)
		assert.Contains(t, err.Error(), "struct cannot be nil")
	})

	t.Run("nil converter", func(t *testing.T) {
		data := TestStruct{Name: "Test"}

		tv, err := generation.NewTemplateVariablesFromStruct(data, nil)
		assert.Error(t, err)
		assert.Nil(t, tv)
		assert.Contains(t, err.Error(), "converter cannot be nil")
	})
}

func TestTemplateVariables_ToMap(t *testing.T) {
	t.Run("empty variables", func(t *testing.T) {
		tv := generation.NewTemplateVariables(nil)
		result := tv.ToMap()

		assert.NotNil(t, result)
		assert.Equal(t, 0, len(result))
	})

	t.Run("simple variables", func(t *testing.T) {
		input := map[string]interface{}{
			"name":   "John",
			"age":    30,
			"active": true,
		}

		tv, err := generation.NewTemplateVariablesFromMap(input)
		require.NoError(t, err)

		result := tv.ToMap()
		assert.Equal(t, 3, len(result))
		assert.Equal(t, "John", result["name"])
		assert.Equal(t, float64(30), result["age"])
		assert.Equal(t, true, result["active"])
	})
}

func TestTemplateVariables_Flatten(t *testing.T) {
	t.Run("empty variables", func(t *testing.T) {
		tv := generation.NewTemplateVariables(nil)
		result := tv.Flatten()

		assert.NotNil(t, result)
		assert.Equal(t, 0, len(result))
	})

	t.Run("simple variables", func(t *testing.T) {
		input := map[string]interface{}{
			"name":  "John",
			"email": "john@example.com",
		}

		tv, err := generation.NewTemplateVariablesFromMap(input)
		require.NoError(t, err)

		result := tv.Flatten()
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "John", result["name"])
		assert.Equal(t, "john@example.com", result["email"])
	})

	t.Run("nested variables", func(t *testing.T) {
		input := map[string]interface{}{
			"user": map[string]interface{}{
				"name":  "John",
				"email": "john@example.com",
			},
		}

		tv, err := generation.NewTemplateVariablesFromMap(input)
		require.NoError(t, err)

		result := tv.Flatten()
		assert.Contains(t, result, "user.name")
		assert.Contains(t, result, "user.email")
		assert.Equal(t, "John", result["user.name"])
		assert.Equal(t, "john@example.com", result["user.email"])
	})

	t.Run("array variables", func(t *testing.T) {
		input := map[string]interface{}{
			"tags": []interface{}{"tag1", "tag2", "tag3"},
		}

		tv, err := generation.NewTemplateVariablesFromMap(input)
		require.NoError(t, err)

		result := tv.Flatten()
		assert.Contains(t, result, "tags[0]")
		assert.Contains(t, result, "tags[1]")
		assert.Contains(t, result, "tags[2]")
		assert.Equal(t, "tag1", result["tags[0]"])
		assert.Equal(t, "tag2", result["tags[1]"])
		assert.Equal(t, "tag3", result["tags[2]"])
	})
}

func TestTemplateVariables_GetAndSet(t *testing.T) {
	t.Run("get existing variable", func(t *testing.T) {
		vars := config.NewVariables()
		vars.SetString("test_key", "test_value")

		tv := generation.NewTemplateVariables(vars)

		value, exists := tv.Get("test_key")
		assert.True(t, exists)
		assert.NotNil(t, value)
		assert.Equal(t, "test_value", value.String())
	})

	t.Run("get non-existing variable", func(t *testing.T) {
		tv := generation.NewTemplateVariables(nil)

		value, exists := tv.Get("non_existing")
		assert.False(t, exists)
		assert.Nil(t, value)
	})

	t.Run("set and get string variable", func(t *testing.T) {
		tv := generation.NewTemplateVariables(nil)

		err := tv.SetString("new_key", "new_value")
		require.NoError(t, err)

		value, exists := tv.GetString("new_key")
		assert.True(t, exists)
		assert.Equal(t, "new_value", value)
	})
}

func TestTemplateVariables_Validate(t *testing.T) {
	t.Run("nil variables", func(t *testing.T) {
		tv := &generation.TemplateVariables{}
		err := tv.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("empty variables", func(t *testing.T) {
		tv := generation.NewTemplateVariables(nil)
		err := tv.Validate()

		assert.NoError(t, err) // Empty is valid
	})

	t.Run("valid variables", func(t *testing.T) {
		input := map[string]interface{}{
			"name": "John",
			"age":  30,
		}

		tv, err := generation.NewTemplateVariablesFromMap(input)
		require.NoError(t, err)

		err = tv.Validate()
		assert.NoError(t, err)
	})
}

func TestTemplateVariables_Clone(t *testing.T) {
	t.Run("clone nil variables", func(t *testing.T) {
		tv := &generation.TemplateVariables{}
		clone := tv.Clone()

		assert.NotNil(t, clone)
		assert.Equal(t, 0, clone.Len())
	})

	t.Run("clone with variables", func(t *testing.T) {
		input := map[string]interface{}{
			"name":  "John",
			"email": "john@example.com",
		}

		tv, err := generation.NewTemplateVariablesFromMap(input)
		require.NoError(t, err)

		clone := tv.Clone()
		assert.NotNil(t, clone)
		assert.Equal(t, tv.Len(), clone.Len())

		// Verify values are the same
		flattened1 := tv.Flatten()
		flattened2 := clone.Flatten()
		assert.Equal(t, flattened1, flattened2)
	})
}

func TestTemplateVariables_Merge(t *testing.T) {
	t.Run("merge with nil", func(t *testing.T) {
		tv := generation.NewTemplateVariables(nil)
		tv.SetString("key1", "value1")

		tv.Merge(nil)
		assert.Equal(t, 1, tv.Len())
	})

	t.Run("merge non-overlapping variables", func(t *testing.T) {
		tv1, err := generation.NewTemplateVariablesFromMap(map[string]interface{}{
			"key1": "value1",
		})
		require.NoError(t, err)

		tv2, err := generation.NewTemplateVariablesFromMap(map[string]interface{}{
			"key2": "value2",
		})
		require.NoError(t, err)

		tv1.Merge(tv2)
		assert.Equal(t, 2, tv1.Len())

		val1, exists1 := tv1.GetString("key1")
		assert.True(t, exists1)
		assert.Equal(t, "value1", val1)

		val2, exists2 := tv1.GetString("key2")
		assert.True(t, exists2)
		assert.Equal(t, "value2", val2)
	})

	t.Run("merge overlapping variables", func(t *testing.T) {
		tv1, err := generation.NewTemplateVariablesFromMap(map[string]interface{}{
			"key1": "value1",
			"key2": "value2_old",
		})
		require.NoError(t, err)

		tv2, err := generation.NewTemplateVariablesFromMap(map[string]interface{}{
			"key2": "value2_new",
			"key3": "value3",
		})
		require.NoError(t, err)

		tv1.Merge(tv2)
		assert.Equal(t, 3, tv1.Len())

		val2, exists2 := tv1.GetString("key2")
		assert.True(t, exists2)
		assert.Equal(t, "value2_new", val2) // Should be overridden
	})
}

func TestTemplateVariables_Keys(t *testing.T) {
	t.Run("empty variables", func(t *testing.T) {
		tv := generation.NewTemplateVariables(nil)
		keys := tv.Keys()

		assert.NotNil(t, keys)
		assert.Equal(t, 0, len(keys))
	})

	t.Run("with variables", func(t *testing.T) {
		input := map[string]interface{}{
			"name":  "John",
			"email": "john@example.com",
			"age":   30,
		}

		tv, err := generation.NewTemplateVariablesFromMap(input)
		require.NoError(t, err)

		keys := tv.Keys()
		assert.Equal(t, 3, len(keys))
		assert.Contains(t, keys, "name")
		assert.Contains(t, keys, "email")
		assert.Contains(t, keys, "age")
	})
}

func TestTemplateVariables_IntegrationWithConverter(t *testing.T) {
	type TestUser struct {
		Name    string `autopdf:"name"`
		Email   string `autopdf:"email"`
		Age     int    `autopdf:"age"`
		Active  bool   `autopdf:"active"`
		Profile struct {
			Bio     string `autopdf:"bio"`
			Website string `autopdf:"website"`
		} `autopdf:"profile"`
	}

	t.Run("convert struct with nested fields", func(t *testing.T) {
		conv := converter.BuildWithDefaults()

		user := TestUser{
			Name:   "Jane Smith",
			Email:  "jane@example.com",
			Age:    28,
			Active: true,
		}
		user.Profile.Bio = "Software Engineer"
		user.Profile.Website = "https://janesmith.dev"

		tv, err := generation.NewTemplateVariablesFromStruct(user, conv)
		require.NoError(t, err)
		assert.NotNil(t, tv)
		assert.Greater(t, tv.Len(), 0)

		// Verify flattened output
		flattened := tv.Flatten()
		assert.Contains(t, flattened, "name")
		assert.Contains(t, flattened, "email")
		assert.Contains(t, flattened, "age")
		assert.Contains(t, flattened, "active")

		// Check nested values
		assert.Contains(t, flattened, "profile.bio")
		assert.Contains(t, flattened, "profile.website")
		assert.Equal(t, "Software Engineer", flattened["profile.bio"])
		assert.Equal(t, "https://janesmith.dev", flattened["profile.website"])
	})
}

func TestTemplateVariables_ComplexConversions(t *testing.T) {
	t.Run("deeply nested structures", func(t *testing.T) {
		input := map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": map[string]interface{}{
					"level3": map[string]interface{}{
						"value": "deep_value",
					},
				},
			},
		}

		tv, err := generation.NewTemplateVariablesFromMap(input)
		require.NoError(t, err)

		flattened := tv.Flatten()
		assert.Contains(t, flattened, "level1.level2.level3.value")
		assert.Equal(t, "deep_value", flattened["level1.level2.level3.value"])
	})

	t.Run("array of maps", func(t *testing.T) {
		input := map[string]interface{}{
			"users": []interface{}{
				map[string]interface{}{
					"name": "User1",
					"age":  25,
				},
				map[string]interface{}{
					"name": "User2",
					"age":  30,
				},
			},
		}

		tv, err := generation.NewTemplateVariablesFromMap(input)
		require.NoError(t, err)

		flattened := tv.Flatten()
		assert.Contains(t, flattened, "users[0].name")
		assert.Contains(t, flattened, "users[0].age")
		assert.Contains(t, flattened, "users[1].name")
		assert.Contains(t, flattened, "users[1].age")
	})
}

func TestTemplateVariables_EdgeCases(t *testing.T) {
	t.Run("nil values in map", func(t *testing.T) {
		input := map[string]interface{}{
			"name":  "John",
			"email": nil,
		}

		tv, err := generation.NewTemplateVariablesFromMap(input)
		require.NoError(t, err)

		flattened := tv.Flatten()
		// Nil converts to "<nil>" string representation
		assert.Contains(t, flattened, "email")
	})

	t.Run("empty strings", func(t *testing.T) {
		input := map[string]interface{}{
			"name":  "",
			"email": "test@example.com",
		}

		tv, err := generation.NewTemplateVariablesFromMap(input)
		require.NoError(t, err)

		flattened := tv.Flatten()
		assert.Equal(t, "", flattened["name"])
		assert.Equal(t, "test@example.com", flattened["email"])
	})

	t.Run("numeric types", func(t *testing.T) {
		input := map[string]interface{}{
			"int_val":   42,
			"float_val": 3.14159,
			"bool_val":  true,
		}

		tv, err := generation.NewTemplateVariablesFromMap(input)
		require.NoError(t, err)

		flattened := tv.Flatten()
		assert.Equal(t, "42", flattened["int_val"])
		assert.Contains(t, flattened["float_val"], "3.14")
		assert.Equal(t, "true", flattened["bool_val"])
	})
}
