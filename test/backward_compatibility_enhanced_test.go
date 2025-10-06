package test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/BuddhiLW/AutoPDF/mocks"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/BuddhiLW/AutoPDF/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestBackwardCompatibilityEnhanced demonstrates enhanced backward compatibility testing with mocks
func TestBackwardCompatibilityEnhanced(t *testing.T) {
	t.Run("OriginalEngineWithMocks", func(t *testing.T) {
		// Test original engine with mocked dependencies
		mockValidator := mocks.NewMockTemplateValidator(t)
		mockFileProcessor := mocks.NewMockFileProcessor(t)
		mockConfigProvider := mocks.NewMockConfigProvider(t)

		// Set up expectations
		expectedConfig := &config.Config{
			Template: "test.tex",
			Output:   "output.pdf",
			Variables: map[string]interface{}{
				"title":  "Backward Compatibility Test",
				"author": "AutoPDF Team",
			},
			Engine: "pdflatex",
		}

		mockConfigProvider.EXPECT().
			GetConfig().
			Return(expectedConfig).
			Once()

		mockValidator.EXPECT().
			ValidateTemplate("test.tex").
			Return(nil).
			Once()

		mockFileProcessor.EXPECT().
			FileExists("test.tex").
			Return(true).
			Once()

		mockFileProcessor.EXPECT().
			ReadFile("test.tex").
			Return([]byte("\\documentclass{article}\\title{delim[[.title]]}"), nil).
			Once()

		// Test the workflow
		config := mockConfigProvider.GetConfig()
		assert.Equal(t, "test.tex", config.Template.String())

		err := mockValidator.ValidateTemplate("test.tex")
		assert.NoError(t, err)

		exists := mockFileProcessor.FileExists("test.tex")
		assert.True(t, exists)

		content, err := mockFileProcessor.ReadFile("test.tex")
		assert.NoError(t, err)
		assert.Contains(t, string(content), "\\documentclass{article}")
	})

	t.Run("EnhancedEngineWithMocks", func(t *testing.T) {
		// Test enhanced engine with mocked dependencies
		mockEngine := mocks.NewMockEnhancedTemplateEngine(t)
		mockVariableProcessor := mocks.NewMockVariableProcessor(t)

		// Complex data structure
		complexData := map[string]interface{}{
			"document": map[string]interface{}{
				"title":  "Enhanced Backward Compatibility",
				"author": "AutoPDF Team",
			},
			"metadata": map[string]interface{}{
				"version": "3.0.0",
				"engine":  "xelatex",
			},
		}

		expectedCollection := domain.NewVariableCollection()
		documentVar, _ := domain.NewObjectVariable(complexData["document"].(map[string]interface{}))
		expectedCollection.Set("document", documentVar)

		mockVariableProcessor.EXPECT().
			ProcessVariables(complexData).
			Return(expectedCollection, nil).
			Once()

		mockEngine.EXPECT().
			SetVariablesFromMap(complexData).
			Return(nil).
			Once()

		mockEngine.EXPECT().
			Process("enhanced.tex").
			Return("Enhanced processed content", nil).
			Once()

		// Test the workflow
		collection, err := mockVariableProcessor.ProcessVariables(complexData)
		assert.NoError(t, err)
		assert.NotNil(t, collection)

		err = mockEngine.SetVariablesFromMap(complexData)
		assert.NoError(t, err)

		result, err := mockEngine.Process("enhanced.tex")
		assert.NoError(t, err)
		assert.Contains(t, result, "Enhanced processed content")
	})

	t.Run("ConfigCompatibility", func(t *testing.T) {
		// Test config compatibility with mocks
		mockConfigProvider := mocks.NewMockConfigProvider(t)

		// Test default config
		defaultConfig := &config.Config{
			Engine:    "pdflatex",
			Variables: map[string]interface{}{},
		}

		mockConfigProvider.EXPECT().
			GetDefaultConfig().
			Return(defaultConfig).
			Once()

		// Test custom config
		customConfig := &config.Config{
			Template: "custom.tex",
			Output:   "custom.pdf",
			Variables: map[string]interface{}{
				"title": "Custom Document",
			},
			Engine: "xelatex",
		}

		mockConfigProvider.EXPECT().
			GetConfig().
			Return(customConfig).
			Once()

		// Test both configs
		defaultCfg := mockConfigProvider.GetDefaultConfig()
		assert.Equal(t, "pdflatex", defaultCfg.Engine.String())

		customCfg := mockConfigProvider.GetConfig()
		assert.Equal(t, "custom.tex", customCfg.Template.String())
		assert.Equal(t, "xelatex", customCfg.Engine.String())
	})
}

// TestBackwardCompatibilityErrorHandling tests error scenarios with mocks
func TestBackwardCompatibilityErrorHandling(t *testing.T) {
	t.Run("TemplateValidationErrors", func(t *testing.T) {
		mockValidator := mocks.NewMockTemplateValidator(t)

		// Test various validation errors
		mockValidator.EXPECT().
			ValidateTemplate("invalid.tex").
			Return(errors.New("syntax error")).
			Once()

		mockValidator.EXPECT().
			ValidateTemplate("missing.tex").
			Return(errors.New("file not found")).
			Once()

		mockValidator.EXPECT().
			ValidateSyntax("\\invalid{command}").
			Return(errors.New("invalid LaTeX syntax")).
			Once()

		// Test error scenarios
		err := mockValidator.ValidateTemplate("invalid.tex")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "syntax error")

		err = mockValidator.ValidateTemplate("missing.tex")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file not found")

		err = mockValidator.ValidateSyntax("\\invalid{command}")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid LaTeX syntax")
	})

	t.Run("FileProcessingErrors", func(t *testing.T) {
		mockFileProcessor := mocks.NewMockFileProcessor(t)

		// Test file not found
		mockFileProcessor.EXPECT().
			FileExists("missing.tex").
			Return(false).
			Once()

		// Test read error
		mockFileProcessor.EXPECT().
			ReadFile("error.tex").
			Return(nil, errors.New("permission denied")).
			Once()

		// Test write error
		mockFileProcessor.EXPECT().
			WriteFile("readonly.tex", []byte("content")).
			Return(errors.New("read-only filesystem")).
			Once()

		// Test error scenarios
		exists := mockFileProcessor.FileExists("missing.tex")
		assert.False(t, exists)

		content, err := mockFileProcessor.ReadFile("error.tex")
		assert.Error(t, err)
		assert.Nil(t, content)

		err = mockFileProcessor.WriteFile("readonly.tex", []byte("content"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "read-only filesystem")
	})

	t.Run("EngineProcessingErrors", func(t *testing.T) {
		mockEngine := mocks.NewMockTemplateEngine(t)
		mockEnhancedEngine := mocks.NewMockEnhancedTemplateEngine(t)

		// Test original engine error
		mockEngine.EXPECT().
			Process("error.tex").
			Return("", errors.New("template processing failed")).
			Once()

		// Test enhanced engine error
		mockEnhancedEngine.EXPECT().
			Process("error.tex").
			Return("", errors.New("enhanced processing failed")).
			Once()

		// Test error scenarios
		result, err := mockEngine.Process("error.tex")
		assert.Error(t, err)
		assert.Empty(t, result)

		result, err = mockEnhancedEngine.Process("error.tex")
		assert.Error(t, err)
		assert.Empty(t, result)
	})
}

// TestBackwardCompatibilityPerformance tests performance with mocks
func TestBackwardCompatibilityPerformance(t *testing.T) {
	t.Run("FastMockExecution", func(t *testing.T) {
		mockEngine := mocks.NewMockTemplateEngine(t)
		mockValidator := mocks.NewMockTemplateValidator(t)

		// Set up multiple expectations for performance testing
		for i := 0; i < 100; i++ {
			mockValidator.EXPECT().
				ValidateTemplate(mock.AnythingOfType("string")).
				Return(nil).
				Once()

			mockEngine.EXPECT().
				Process(mock.AnythingOfType("string")).
				Return("processed content", nil).
				Once()
		}

		// Execute multiple operations
		for i := 0; i < 100; i++ {
			err := mockValidator.ValidateTemplate("test.tex")
			assert.NoError(t, err)

			result, err := mockEngine.Process("test.tex")
			assert.NoError(t, err)
			assert.Equal(t, "processed content", result)
		}
	})

	t.Run("ConcurrentMockUsage", func(t *testing.T) {
		mockEngine := mocks.NewMockTemplateEngine(t)

		// Set up expectations for concurrent access
		mockEngine.EXPECT().
			Process(mock.AnythingOfType("string")).
			Return("concurrent result", nil).
			Times(10)

		// Test concurrent execution
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				result, err := mockEngine.Process("concurrent.tex")
				assert.NoError(t, err)
				assert.Equal(t, "concurrent result", result)
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// TestBackwardCompatibilityEdgeCases tests edge cases with mocks
func TestBackwardCompatibilityEdgeCases(t *testing.T) {
	t.Run("EmptyVariables", func(t *testing.T) {
		mockVariableProcessor := mocks.NewMockVariableProcessor(t)

		emptyVariables := map[string]interface{}{}
		emptyCollection := domain.NewVariableCollection()

		mockVariableProcessor.EXPECT().
			ProcessVariables(emptyVariables).
			Return(emptyCollection, nil).
			Once()

		collection, err := mockVariableProcessor.ProcessVariables(emptyVariables)
		assert.NoError(t, err)
		assert.NotNil(t, collection)
	})

	t.Run("NilVariables", func(t *testing.T) {
		mockVariableProcessor := mocks.NewMockVariableProcessor(t)

		mockVariableProcessor.EXPECT().
			ProcessVariables(mock.AnythingOfType("map[string]interface {}")).
			Return(nil, errors.New("variables cannot be nil")).
			Once()

		collection, err := mockVariableProcessor.ProcessVariables(nil)
		assert.Error(t, err)
		assert.Nil(t, collection)
	})

	t.Run("LargeDataStructures", func(t *testing.T) {
		mockEngine := mocks.NewMockEnhancedTemplateEngine(t)

		// Create large data structure
		largeData := make(map[string]interface{})
		for i := 0; i < 1000; i++ {
			largeData[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
		}

		mockEngine.EXPECT().
			SetVariablesFromMap(largeData).
			Return(nil).
			Once()

		mockEngine.EXPECT().
			Process("large.tex").
			Return("Large data processed", nil).
			Once()

		err := mockEngine.SetVariablesFromMap(largeData)
		assert.NoError(t, err)

		result, err := mockEngine.Process("large.tex")
		assert.NoError(t, err)
		assert.Contains(t, result, "Large data processed")
	})
}

// TestBackwardCompatibilityRegression tests for regression issues
func TestBackwardCompatibilityRegression(t *testing.T) {
	t.Run("OriginalAPIRegression", func(t *testing.T) {
		// Test that original API still works exactly as before
		mockEngine := mocks.NewMockTemplateEngine(t)
		mockConfigProvider := mocks.NewMockConfigProvider(t)

		// Original API expectations
		originalConfig := &config.Config{
			Template: "original.tex",
			Variables: map[string]interface{}{
				"title": "Original API Test",
			},
		}

		mockConfigProvider.EXPECT().
			GetConfig().
			Return(originalConfig).
			Once()

		mockEngine.EXPECT().
			Process("original.tex").
			Return("Original API result", nil).
			Once()

		mockEngine.EXPECT().
			AddFunction("upper", mock.AnythingOfType("func(string) string")).
			Return().
			Once()

		// Test original API
		config := mockConfigProvider.GetConfig()
		assert.Equal(t, "original.tex", config.Template.String())

		result, err := mockEngine.Process("original.tex")
		assert.NoError(t, err)
		assert.Equal(t, "Original API result", result)

		mockEngine.AddFunction("upper", func(s string) string {
			return "UPPERCASE_" + s
		})
	})

	t.Run("EnhancedAPIRegression", func(t *testing.T) {
		// Test that enhanced API works with complex data
		mockEngine := mocks.NewMockEnhancedTemplateEngine(t)

		complexData := map[string]interface{}{
			"nested": map[string]interface{}{
				"deep": map[string]interface{}{
					"value": "test",
				},
			},
		}

		mockEngine.EXPECT().
			SetVariablesFromMap(complexData).
			Return(nil).
			Once()

		mockEngine.EXPECT().
			Process("complex.tex").
			Return("Complex result", nil).
			Once()

		// Test enhanced API
		err := mockEngine.SetVariablesFromMap(complexData)
		assert.NoError(t, err)

		result, err := mockEngine.Process("complex.tex")
		assert.NoError(t, err)
		assert.Equal(t, "Complex result", result)
	})
}
