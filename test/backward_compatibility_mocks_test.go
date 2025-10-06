package test

import (
	"testing"

	"github.com/BuddhiLW/AutoPDF/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBackwardCompatibilityWithMocks demonstrates how to use Mockery for backward compatibility testing
func TestBackwardCompatibilityWithMocks(t *testing.T) {
	t.Run("Template Engine Integration", func(t *testing.T) {
		// Create mocks using Mockery
		mockEngine := mocks.NewMockTemplateEngine(t)
		mockValidator := mocks.NewMockTemplateValidator(t)
		mockFileProcessor := mocks.NewMockFileProcessor(t)

		// Set up mock expectations
		mockFileProcessor.EXPECT().
			ReadFile("test.tex").
			Return([]byte("Hello {{name}}"), nil).
			Once()

		mockValidator.EXPECT().
			ValidateTemplate("test.tex").
			Return(nil).
			Once()

		mockEngine.EXPECT().
			Process("test.tex").
			Return("Hello World", nil).
			Once()

		// Test the integration
		content, err := mockFileProcessor.ReadFile("test.tex")
		require.NoError(t, err)
		assert.Equal(t, "Hello {{name}}", string(content))

		err = mockValidator.ValidateTemplate("test.tex")
		require.NoError(t, err)

		result, err := mockEngine.Process("test.tex")
		require.NoError(t, err)
		assert.Equal(t, "Hello World", result)
	})

	t.Run("Variable Processing Integration", func(t *testing.T) {
		// Create mocks for variable processing
		mockProcessor := mocks.NewMockVariableProcessor(t)

		// Set up mock expectations
		mockProcessor.EXPECT().
			ProcessVariables(map[string]interface{}{
				"name": "John Doe",
				"age":  30,
			}).
			Return(nil, nil).
			Once()

		// Test variable processing
		variables := map[string]interface{}{
			"name": "John Doe",
			"age":  30,
		}

		_, err := mockProcessor.ProcessVariables(variables)
		require.NoError(t, err)
	})

	t.Run("Configuration Integration", func(t *testing.T) {
		// Create mocks for configuration
		mockConfigProvider := mocks.NewMockConfigProvider(t)

		// Set up mock expectations
		mockConfigProvider.EXPECT().
			GetConfig().
			Return(nil).
			Once()

		// Test configuration access
		config := mockConfigProvider.GetConfig()
		assert.Nil(t, config)
	})

	t.Run("Error Handling Integration", func(t *testing.T) {
		// Create mocks for error scenarios
		mockEngine := mocks.NewMockTemplateEngine(t)
		mockValidator := mocks.NewMockTemplateValidator(t)

		// Set up error expectations
		mockValidator.EXPECT().
			ValidateTemplate("error.tex").
			Return(assert.AnError).
			Once()

		mockEngine.EXPECT().
			Process("error.tex").
			Return("", assert.AnError).
			Once()

		// Test error handling
		err := mockValidator.ValidateTemplate("error.tex")
		assert.Error(t, err)

		_, err = mockEngine.Process("error.tex")
		assert.Error(t, err)
	})

	t.Run("Enhanced Template Engine Integration", func(t *testing.T) {
		// Create mocks for enhanced template engine
		mockEnhancedEngine := mocks.NewMockEnhancedTemplateEngine(t)

		// Set up mock expectations
		mockEnhancedEngine.EXPECT().
			SetVariable("key", "value").
			Return(nil).
			Once()

		mockEnhancedEngine.EXPECT().
			GetVariable("key").
			Return(nil, nil).
			Once()

		mockEnhancedEngine.EXPECT().
			Process("enhanced.tex").
			Return("enhanced result", nil).
			Once()

		// Test enhanced functionality
		err := mockEnhancedEngine.SetVariable("key", "value")
		require.NoError(t, err)

		_, err = mockEnhancedEngine.GetVariable("key")
		require.NoError(t, err)

		result, err := mockEnhancedEngine.Process("enhanced.tex")
		require.NoError(t, err)
		assert.Equal(t, "enhanced result", result)
	})
}
