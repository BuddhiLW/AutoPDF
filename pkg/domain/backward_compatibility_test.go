package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBackwardCompatibility ensures that all existing functionality continues to work
// This establishes the "minimum bar" for backward compatibility
func TestBackwardCompatibility(t *testing.T) {
	t.Run("Variable Creation", func(t *testing.T) {
		// Test that all original variable creation methods still work
		tests := []struct {
			name     string
			createFn func() *Variable
			expected VariableType
		}{
			{"String Variable", func() *Variable { v, _ := NewStringVariable("test"); return v }, VariableTypeString},
			{"Number Variable", func() *Variable { v, _ := NewNumberVariable(42); return v }, VariableTypeNumber},
			{"Boolean Variable", func() *Variable { return NewBooleanVariable(true) }, VariableTypeBoolean},
			{"Array Variable", func() *Variable { v, _ := NewArrayVariable([]interface{}{"item"}); return v }, VariableTypeArray},
			{"Object Variable", func() *Variable { v, _ := NewObjectVariable(map[string]interface{}{"key": "value"}); return v }, VariableTypeObject},
			{"Null Variable", func() *Variable { return NewNullVariable() }, VariableTypeNull},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				variable := tt.createFn()
				assert.Equal(t, tt.expected, variable.Type)
				assert.NotNil(t, variable)
			})
		}
	})

	t.Run("Variable Type Conversion", func(t *testing.T) {
		// Test that all type conversion methods still work
		stringVar, _ := NewStringVariable("test")
		numberVar, _ := NewNumberVariable(42)
		booleanVar := NewBooleanVariable(true)
		arrayVar, _ := NewArrayVariable([]interface{}{"item1", "item2"})
		objectVar, _ := NewObjectVariable(map[string]interface{}{"key": "value"})

		// Test AsString
		str, err := stringVar.AsString()
		require.NoError(t, err)
		assert.Equal(t, "test", str)

		_, err = numberVar.AsString()
		assert.Error(t, err)

		// Test AsNumber
		num, err := numberVar.AsNumber()
		require.NoError(t, err)
		assert.Equal(t, 42.0, num)

		_, err = stringVar.AsNumber()
		assert.Error(t, err)

		// Test AsBoolean
		boolean, err := booleanVar.AsBoolean()
		require.NoError(t, err)
		assert.Equal(t, true, boolean)

		_, err = stringVar.AsBoolean()
		assert.Error(t, err)

		// Test AsArray
		arr, err := arrayVar.AsArray()
		require.NoError(t, err)
		assert.Len(t, arr, 2)
		assert.Equal(t, "item1", arr[0])
		assert.Equal(t, "item2", arr[1])

		_, err = stringVar.AsArray()
		assert.Error(t, err)

		// Test AsObject
		obj, err := objectVar.AsObject()
		require.NoError(t, err)
		assert.Equal(t, "value", obj["key"])

		_, err = stringVar.AsObject()
		assert.Error(t, err)
	})

	t.Run("Variable Collection Basic Operations", func(t *testing.T) {
		vc := NewVariableCollection()

		// Test Set and Get
		variable, _ := NewStringVariable("test")
		vc.Set("key", variable)

		retrieved, exists := vc.Get("key")
		assert.True(t, exists)
		assert.Equal(t, variable, retrieved)

		// Test Has
		assert.True(t, vc.Has("key"))
		assert.False(t, vc.Has("nonexistent"))

		// Test Size
		assert.Equal(t, 1, vc.Size())

		// Test Keys
		keys := vc.Keys()
		assert.Len(t, keys, 1)
		assert.Contains(t, keys, "key")

		// Test Delete
		vc.Delete("key")
		assert.False(t, vc.Has("key"))
		assert.Equal(t, 0, vc.Size())

		// Test Clear
		key1, _ := NewStringVariable("value1")
		key2, _ := NewStringVariable("value2")
		vc.Set("key1", key1)
		vc.Set("key2", key2)
		assert.Equal(t, 2, vc.Size())

		vc.Clear()
		assert.Equal(t, 0, vc.Size())
	})

	t.Run("Variable Collection Nested Access", func(t *testing.T) {
		vc := NewVariableCollection()

		// Test nested object creation and access
		user, _ := NewObjectVariable(map[string]interface{}{
			"name": "John Doe",
			"age":  30,
			"address": map[string]interface{}{
				"street": "123 Main St",
				"city":   "New York",
			},
		})
		vc.Set("user", user)

		// Test nested access
		nestedVar, err := vc.GetNested("user.name")
		require.NoError(t, err)
		name, err := nestedVar.AsString()
		require.NoError(t, err)
		assert.Equal(t, "John Doe", name)

		// Test deeply nested access
		addressVar, err := vc.GetNested("user.address.street")
		require.NoError(t, err)
		street, err := addressVar.AsString()
		require.NoError(t, err)
		assert.Equal(t, "123 Main St", street)

		// Test error cases
		_, err = vc.GetNested("user.nonexistent")
		assert.Error(t, err)

		_, err = vc.GetNested("nonexistent.key")
		assert.Error(t, err)
	})

	t.Run("Variable Collection SetNested", func(t *testing.T) {
		vc := NewVariableCollection()

		// Test setting nested values
		nameVar, _ := NewStringVariable("Jane Doe")
		err := vc.SetNested("user.name", nameVar)
		require.NoError(t, err)

		ageVar, _ := NewNumberVariable(25)
		err = vc.SetNested("user.age", ageVar)
		require.NoError(t, err)

		streetVar, _ := NewStringVariable("456 Oak Ave")
		err = vc.SetNested("user.address.street", streetVar)
		require.NoError(t, err)

		// Verify nested values were set
		retrievedNameVar, err := vc.GetNested("user.name")
		require.NoError(t, err)
		name, err := retrievedNameVar.AsString()
		require.NoError(t, err)
		assert.Equal(t, "Jane Doe", name)

		retrievedAgeVar, err := vc.GetNested("user.age")
		require.NoError(t, err)
		age, err := retrievedAgeVar.AsNumber()
		require.NoError(t, err)
		assert.Equal(t, 25.0, age)

		streetVar2, err := vc.GetNested("user.address.street")
		require.NoError(t, err)
		street, err := streetVar2.AsString()
		require.NoError(t, err)
		assert.Equal(t, "456 Oak Ave", street)
	})

	t.Run("FromMap and ToMap", func(t *testing.T) {
		// Test FromMap
		data := map[string]interface{}{
			"string":  "test",
			"number":  42,
			"boolean": true,
			"object": map[string]interface{}{
				"nested": "value",
			},
			"array": []interface{}{"item1", "item2"},
		}

		vc := FromMap(data)
		assert.Equal(t, 5, vc.Size())

		// Test ToMap
		result := vc.ToMap()
		assert.Equal(t, "test", result["string"])
		assert.Equal(t, 42, result["number"])
		assert.Equal(t, true, result["boolean"])
		assert.Equal(t, "value", result["object"].(map[string]interface{})["nested"])
		assert.Equal(t, []interface{}{"item1", "item2"}, result["array"])
	})

	t.Run("Template Context Basic Operations", func(t *testing.T) {
		context := NewTemplateContext()

		// Test setting and getting variables
		variable, err := NewStringVariable("test")
		require.NoError(t, err)
		err = context.SetVariable("key", variable)
		require.NoError(t, err)

		retrieved, err := context.GetVariable("key")
		require.NoError(t, err)
		assert.Equal(t, variable, retrieved)

		// Test adding functions
		context.AddFunction("upper", func(s string) string {
			return "UPPERCASE_" + s
		})

		assert.Contains(t, context.Functions, "upper")

		// Test validation
		requiredVars := []string{"key"}
		err = context.ValidateTemplate(requiredVars)
		assert.NoError(t, err)

		// Test validation with missing variables
		requiredVars = []string{"key", "missing"}
		err = context.ValidateTemplate(requiredVars)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required variable 'missing' not found")
	})

	t.Run("Template Context Type Checking", func(t *testing.T) {
		context := NewTemplateContext()

		// Set different types of variables
		stringVar, _ := NewStringVariable("test")
		context.SetVariable("string_var", stringVar)
		numberVar, _ := NewNumberVariable(42)
		context.SetVariable("number_var", numberVar)
		context.SetVariable("boolean_var", NewBooleanVariable(true))
		arrayVar, _ := NewArrayVariable([]interface{}{"item"})
		context.SetVariable("array_var", arrayVar)
		objectVar, _ := NewObjectVariable(map[string]interface{}{"key": "value"})
		context.SetVariable("object_var", objectVar)

		// Test type checking
		isArray, err := context.IsVariableArray("array_var")
		require.NoError(t, err)
		assert.True(t, isArray)

		isObject, err := context.IsVariableObject("object_var")
		require.NoError(t, err)
		assert.True(t, isObject)

		// Test array length
		length, err := context.GetArrayLength("array_var")
		require.NoError(t, err)
		assert.Equal(t, 1, length)

		// Test object keys
		keys, err := context.GetObjectKeys("object_var")
		require.NoError(t, err)
		assert.Len(t, keys, 1)
		assert.Contains(t, keys, "key")
	})

	t.Run("Template Context Clone and Merge", func(t *testing.T) {
		context1 := NewTemplateContext()
		key1Var, _ := NewStringVariable("value1")
		context1.SetVariable("key1", key1Var)
		context1.AddFunction("func1", func() string { return "result1" })

		context2 := NewTemplateContext()
		key2Var, _ := NewStringVariable("value2")
		context2.SetVariable("key2", key2Var)
		context2.AddFunction("func2", func() string { return "result2" })

		// Test clone
		clone := context1.Clone()
		val, err := clone.GetVariable("key1")
		require.NoError(t, err)
		str, err := val.AsString()
		require.NoError(t, err)
		assert.Equal(t, "value1", str)

		// Test merge
		context1.Merge(context2)
		val, err = context1.GetVariable("key2")
		require.NoError(t, err)
		str, err = val.AsString()
		require.NoError(t, err)
		assert.Equal(t, "value2", str)

		assert.Contains(t, context1.Functions, "func1")
		assert.Contains(t, context1.Functions, "func2")
	})
}

// TestMinimumBar establishes the minimum bar for backward compatibility
func TestMinimumBar(t *testing.T) {
	t.Run("Original API Must Work", func(t *testing.T) {
		// This test ensures that the original AutoPDF API continues to work
		// This is our "minimum bar" - if this breaks, we've broken backward compatibility

		// Test 1: Simple variable creation and access
		vc := NewVariableCollection()
		titleVar, err := NewStringVariable("Test Document")
		require.NoError(t, err)
		authorVar, err := NewStringVariable("John Doe")
		require.NoError(t, err)
		vc.Set("title", titleVar)
		vc.Set("author", authorVar)

		title, exists := vc.Get("title")
		require.True(t, exists)
		titleStr, err := title.AsString()
		require.NoError(t, err)
		assert.Equal(t, "Test Document", titleStr)

		// Test 2: Basic template context operations
		context := NewTemplateContext()
		titleVar2, err := NewStringVariable("Test Document")
		require.NoError(t, err)
		err = context.SetVariable("title", titleVar2)
		require.NoError(t, err)

		retrieved, err := context.GetVariable("title")
		require.NoError(t, err)
		assert.Equal(t, "Test Document", retrieved.Value)

		// Test 3: Template data conversion
		data := context.ToTemplateData()
		assert.Equal(t, "Test Document", data["title"])
	})

	t.Run("Enhanced Features Must Not Break Original", func(t *testing.T) {
		// This test ensures that enhanced features don't break original functionality

		// Test that we can still use simple variables with enhanced features
		context := NewTemplateContext()

		// Set simple variables (original way)
		titleVar, _ := NewStringVariable("Simple Title")
		contentVar, _ := NewStringVariable("Simple Content")
		context.SetVariable("title", titleVar)
		context.SetVariable("content", contentVar)

		// Add enhanced features
		metadataVar, _ := NewObjectVariable(map[string]interface{}{
			"created": "2024-01-15",
			"version": "1.0",
		})
		context.SetVariable("metadata", metadataVar)

		// Verify original functionality still works
		title, err := context.GetVariable("title")
		require.NoError(t, err)
		titleStr, err := title.AsString()
		require.NoError(t, err)
		assert.Equal(t, "Simple Title", titleStr)

		// Verify enhanced functionality works alongside original
		metadata, err := context.GetVariable("metadata")
		require.NoError(t, err)
		metadataObj, err := metadata.AsObject()
		require.NoError(t, err)
		assert.Equal(t, "2024-01-15", metadataObj["created"])
	})

	t.Run("Error Handling Must Remain Consistent", func(t *testing.T) {
		// This test ensures that error handling remains consistent

		context := NewTemplateContext()

		// Test getting non-existent variable
		_, err := context.GetVariable("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Test type conversion errors
		stringVar, _ := NewStringVariable("test")
		_, err = stringVar.AsNumber()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a number")

		// Test validation errors
		requiredVars := []string{"missing"}
		err = context.ValidateTemplate(requiredVars)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required variable 'missing' not found")
	})
}
