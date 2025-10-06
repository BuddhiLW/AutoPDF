package test

import (
	"testing"

	"github.com/BuddhiLW/AutoPDF/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockeryDemo demonstrates that Mockery is working correctly
func TestMockeryDemo(t *testing.T) {
	t.Run("Template Engine Mock", func(t *testing.T) {
		// Create mock using Mockery
		mockEngine := mocks.NewMockTemplateEngine(t)

		// Set up expectations
		mockEngine.EXPECT().
			Process("test.tex").
			Return("Hello World", nil).
			Once()

		// Test the mock
		result, err := mockEngine.Process("test.tex")
		require.NoError(t, err)
		assert.Equal(t, "Hello World", result)
	})

	t.Run("File Processor Mock", func(t *testing.T) {
		// Create mock using Mockery
		mockFileProcessor := mocks.NewMockFileProcessor(t)

		// Set up expectations
		mockFileProcessor.EXPECT().
			ReadFile("test.tex").
			Return([]byte("template content"), nil).
			Once()

		// Test the mock
		content, err := mockFileProcessor.ReadFile("test.tex")
		require.NoError(t, err)
		assert.Equal(t, "template content", string(content))
	})

	t.Run("Template Validator Mock", func(t *testing.T) {
		// Create mock using Mockery
		mockValidator := mocks.NewMockTemplateValidator(t)

		// Set up expectations
		mockValidator.EXPECT().
			ValidateTemplate("test.tex").
			Return(nil).
			Once()

		// Test the mock
		err := mockValidator.ValidateTemplate("test.tex")
		require.NoError(t, err)
	})

	t.Run("Error Handling Mock", func(t *testing.T) {
		// Create mock using Mockery
		mockEngine := mocks.NewMockTemplateEngine(t)

		// Set up error expectation
		mockEngine.EXPECT().
			Process("error.tex").
			Return("", assert.AnError).
			Once()

		// Test error handling
		_, err := mockEngine.Process("error.tex")
		assert.Error(t, err)
	})

	t.Run("Multiple Calls Mock", func(t *testing.T) {
		// Create mock using Mockery
		mockEngine := mocks.NewMockTemplateEngine(t)

		// Set up multiple expectations
		mockEngine.EXPECT().
			Process("template1.tex").
			Return("result1", nil).
			Once()

		mockEngine.EXPECT().
			Process("template2.tex").
			Return("result2", nil).
			Once()

		// Test multiple calls
		result1, err1 := mockEngine.Process("template1.tex")
		require.NoError(t, err1)
		assert.Equal(t, "result1", result1)

		result2, err2 := mockEngine.Process("template2.tex")
		require.NoError(t, err2)
		assert.Equal(t, "result2", result2)
	})

	t.Run("Enhanced Template Engine Mock", func(t *testing.T) {
		// Create mock using Mockery
		mockEnhancedEngine := mocks.NewMockEnhancedTemplateEngine(t)

		// Set up expectations
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
