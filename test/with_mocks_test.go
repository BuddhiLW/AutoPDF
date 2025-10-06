package test

import (
	"errors"
	"testing"

	"strings"

	"github.com/BuddhiLW/AutoPDF/mocks"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/BuddhiLW/AutoPDF/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestTemplateEngineWithMocks demonstrates how to use Mockery-generated mocks
func TestTemplateEngineWithMocks(t *testing.T) {
	// Create a mock template engine
	mockEngine := mocks.NewMockTemplateEngine(t)

	// Set up expectations
	mockEngine.EXPECT().
		Process("test.tex").
		Return("processed content", nil).
		Once()

	mockEngine.EXPECT().
		AddFunction("upper", mock.AnythingOfType("func(string) string")).
		Return().
		Once()

	// Test the mock
	result, err := mockEngine.Process("test.tex")
	assert.NoError(t, err)
	assert.Equal(t, "processed content", result)

	mockEngine.AddFunction("upper", func(s string) string {
		return "UPPERCASE_" + s
	})
}

// TestEnhancedTemplateEngineWithMocks demonstrates enhanced engine mocking
func TestEnhancedTemplateEngineWithMocks(t *testing.T) {
	// Create a mock enhanced template engine
	mockEngine := mocks.NewMockEnhancedTemplateEngine(t)

	// Set up expectations for complex operations
	mockEngine.EXPECT().
		SetVariablesFromMap(map[string]interface{}{
			"title":   "Test Document",
			"author":  "John Doe",
			"content": "This is a test document.",
		}).
		Return(nil).
		Once()

	mockEngine.EXPECT().
		Process("enhanced.tex").
		Return("Enhanced processed content", nil).
		Once()

	// Test the mock
	err := mockEngine.SetVariablesFromMap(map[string]interface{}{
		"title":   "Test Document",
		"author":  "John Doe",
		"content": "This is a test document.",
	})
	assert.NoError(t, err)

	result, err := mockEngine.Process("enhanced.tex")
	assert.NoError(t, err)
	assert.Equal(t, "Enhanced processed content", result)
}

// TestConfigProviderWithMocks demonstrates config provider mocking
func TestConfigProviderWithMocks(t *testing.T) {
	// Create a mock config provider
	mockProvider := mocks.NewMockConfigProvider(t)

	// Create expected config
	expectedConfig := &config.Config{
		Template: "test.tex",
		Output:   "output.pdf",
		Variables: map[string]interface{}{
			"title": "Test Document",
		},
		Engine: "pdflatex",
	}

	// Set up expectations
	mockProvider.EXPECT().
		GetConfig().
		Return(expectedConfig).
		Once()

	mockProvider.EXPECT().
		GetDefaultConfig().
		Return(&config.Config{
			Engine: "pdflatex",
		}).
		Once()

	// Test the mock
	config := mockProvider.GetConfig()
	assert.Equal(t, "test.tex", config.Template.String())
	assert.Equal(t, "output.pdf", config.Output.String())
	assert.Equal(t, "pdflatex", config.Engine.String())

	defaultConfig := mockProvider.GetDefaultConfig()
	assert.Equal(t, "pdflatex", defaultConfig.Engine.String())
}

// TestVariableProcessorWithMocks demonstrates variable processor mocking
func TestVariableProcessorWithMocks(t *testing.T) {
	// Create a mock variable processor
	mockProcessor := mocks.NewMockVariableProcessor(t)

	// Create test variables
	testVariables := map[string]interface{}{
		"title":   "Test Document",
		"author":  "John Doe",
		"content": "This is a test document.",
	}

	// Create expected variable collection
	expectedCollection := domain.NewVariableCollection()
	titleVar, _ := domain.NewStringVariable("Test Document")
	authorVar, _ := domain.NewStringVariable("John Doe")
	contentVar, _ := domain.NewStringVariable("This is a test document.")
	expectedCollection.Set("title", titleVar)
	expectedCollection.Set("author", authorVar)
	expectedCollection.Set("content", contentVar)

	// Set up expectations
	mockProcessor.EXPECT().
		ProcessVariables(testVariables).
		Return(expectedCollection, nil).
		Once()

	mockProcessor.EXPECT().
		GetVariable("title").
		Return(titleVar, nil).
		Once()

	// Test the mock
	collection, err := mockProcessor.ProcessVariables(testVariables)
	assert.NoError(t, err)
	assert.NotNil(t, collection)

	variable, err := mockProcessor.GetVariable("title")
	assert.NoError(t, err)
	assert.Equal(t, "Test Document", variable.Value)
}

// TestTemplateValidatorWithMocks demonstrates template validator mocking
func TestTemplateValidatorWithMocks(t *testing.T) {
	// Create a mock template validator
	mockValidator := mocks.NewMockTemplateValidator(t)

	// Set up expectations
	mockValidator.EXPECT().
		ValidateTemplate("valid.tex").
		Return(nil).
		Once()

	mockValidator.EXPECT().
		ValidateTemplate("invalid.tex").
		Return(errors.New("template validation failed")).
		Once()

	mockValidator.EXPECT().
		ValidateSyntax("\\documentclass{article}").
		Return(nil).
		Once()

	// Test the mock
	err := mockValidator.ValidateTemplate("valid.tex")
	assert.NoError(t, err)

	err = mockValidator.ValidateTemplate("invalid.tex")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template validation failed")

	err = mockValidator.ValidateSyntax("\\documentclass{article}")
	assert.NoError(t, err)
}

// TestFileProcessorWithMocks demonstrates file processor mocking
func TestFileProcessorWithMocks(t *testing.T) {
	// Create a mock file processor
	mockProcessor := mocks.NewMockFileProcessor(t)

	// Set up expectations
	mockProcessor.EXPECT().
		ReadFile("test.tex").
		Return([]byte("\\documentclass{article}"), nil).
		Once()

	mockProcessor.EXPECT().
		WriteFile("output.tex", []byte("processed content")).
		Return(nil).
		Once()

	mockProcessor.EXPECT().
		FileExists("test.tex").
		Return(true).
		Once()

	mockProcessor.EXPECT().
		CreateDirectory("output").
		Return(nil).
		Once()

	// Test the mock
	content, err := mockProcessor.ReadFile("test.tex")
	assert.NoError(t, err)
	assert.Equal(t, "\\documentclass{article}", string(content))

	err = mockProcessor.WriteFile("output.tex", []byte("processed content"))
	assert.NoError(t, err)

	exists := mockProcessor.FileExists("test.tex")
	assert.True(t, exists)

	err = mockProcessor.CreateDirectory("output")
	assert.NoError(t, err)
}

// TestComplexWorkflowWithMocks demonstrates a complex workflow using multiple mocks
func TestComplexWorkflowWithMocks(t *testing.T) {
	// Create multiple mocks for a complex workflow
	mockEngine := mocks.NewMockTemplateEngine(t)
	mockValidator := mocks.NewMockTemplateValidator(t)
	mockFileProcessor := mocks.NewMockFileProcessor(t)

	// Set up a complex workflow
	// 1. Validate template
	mockValidator.EXPECT().
		ValidateTemplate("template.tex").
		Return(nil).
		Once()

	// 2. Read template file
	mockFileProcessor.EXPECT().
		ReadFile("template.tex").
		Return([]byte("\\documentclass{article}\\title{delim[[.title]]}"), nil).
		Once()

	// 3. Process template
	mockEngine.EXPECT().
		Process("template.tex").
		Return("\\documentclass{article}\\title{Test Document}", nil).
		Once()

	// 4. Write output file
	mockFileProcessor.EXPECT().
		WriteFile("output.tex", []byte("\\documentclass{article}\\title{Test Document}")).
		Return(nil).
		Once()

	// Simulate the workflow
	err := mockValidator.ValidateTemplate("template.tex")
	assert.NoError(t, err)

	content, err := mockFileProcessor.ReadFile("template.tex")
	assert.NoError(t, err)
	assert.Contains(t, string(content), "\\documentclass{article}")

	result, err := mockEngine.Process("template.tex")
	assert.NoError(t, err)
	assert.Contains(t, result, "Test Document")

	err = mockFileProcessor.WriteFile("output.tex", []byte(result))
	assert.NoError(t, err)
}

// TestMockExpectations demonstrates how to test that all expectations were met
func TestMockExpectations(t *testing.T) {
	// Create a mock that we won't call all expected methods
	mockEngine := mocks.NewMockTemplateEngine(t)

	// Set up expectations but don't call all of them
	mockEngine.EXPECT().
		Process("test.tex").
		Return("result", nil).
		Once()

	mockEngine.EXPECT().
		AddFunction("upper", mock.Anything).
		Return().
		Once()

	// Call both expected methods
	result, err := mockEngine.Process("test.tex")
	assert.NoError(t, err)
	assert.Equal(t, "result", result)

	mockEngine.AddFunction("upper", func(s string) string {
		return strings.ToUpper(s)
	})

	// The test will fail because we didn't call AddFunction
	// This demonstrates how mocks help ensure all expected interactions occur
}
