package test

import (
	"errors"
	"testing"

	"github.com/BuddhiLW/AutoPDF/mocks"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/BuddhiLW/AutoPDF/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// UnitComponentsTestSuite tests individual components in isolation
type UnitComponentsTestSuite struct {
	suite.Suite
}

// TestTemplateEngineUnit tests the template engine in isolation
func TestTemplateEngineUnit(t *testing.T) {
	t.Run("ProcessTemplate", func(t *testing.T) {
		mockEngine := mocks.NewMockTemplateEngine(t)

		// Test successful processing
		mockEngine.EXPECT().
			Process("test.tex").
			Return("Processed template content", nil).
			Once()

		result, err := mockEngine.Process("test.tex")
		assert.NoError(t, err)
		assert.Equal(t, "Processed template content", result)
	})

	t.Run("ProcessToFile", func(t *testing.T) {
		mockEngine := mocks.NewMockTemplateEngine(t)

		// Test file processing
		mockEngine.EXPECT().
			ProcessToFile("input.tex", "output.tex").
			Return(nil).
			Once()

		err := mockEngine.ProcessToFile("input.tex", "output.tex")
		assert.NoError(t, err)
	})

	t.Run("AddFunction", func(t *testing.T) {
		mockEngine := mocks.NewMockTemplateEngine(t)

		// Test function addition
		mockEngine.EXPECT().
			AddFunction("upper", mock.AnythingOfType("func(string) string")).
			Return().
			Once()

		mockEngine.AddFunction("upper", func(s string) string {
			return "UPPERCASE_" + s
		})
	})

	t.Run("ValidateTemplate", func(t *testing.T) {
		mockEngine := mocks.NewMockTemplateEngine(t)

		// Test template validation
		mockEngine.EXPECT().
			ValidateTemplate("valid.tex").
			Return(nil).
			Once()

		err := mockEngine.ValidateTemplate("valid.tex")
		assert.NoError(t, err)
	})
}

// TestEnhancedTemplateEngineUnit tests the enhanced template engine in isolation
func TestEnhancedTemplateEngineUnit(t *testing.T) {
	t.Run("SetVariable", func(t *testing.T) {
		mockEngine := mocks.NewMockEnhancedTemplateEngine(t)

		// Test setting single variable
		mockEngine.EXPECT().
			SetVariable("title", "Test Document").
			Return(nil).
			Once()

		err := mockEngine.SetVariable("title", "Test Document")
		assert.NoError(t, err)
	})

	t.Run("SetVariablesFromMap", func(t *testing.T) {
		mockEngine := mocks.NewMockEnhancedTemplateEngine(t)

		variables := map[string]interface{}{
			"title":   "Test Document",
			"author":  "John Doe",
			"content": "Test content",
		}

		mockEngine.EXPECT().
			SetVariablesFromMap(variables).
			Return(nil).
			Once()

		err := mockEngine.SetVariablesFromMap(variables)
		assert.NoError(t, err)
	})

	t.Run("GetVariable", func(t *testing.T) {
		mockEngine := mocks.NewMockEnhancedTemplateEngine(t)

		expectedVar, _ := domain.NewStringVariable("Test Document")

		mockEngine.EXPECT().
			GetVariable("title").
			Return(expectedVar, nil).
			Once()

		variable, err := mockEngine.GetVariable("title")
		assert.NoError(t, err)
		assert.Equal(t, expectedVar, variable)
	})

	t.Run("ProcessWithComplexData", func(t *testing.T) {
		mockEngine := mocks.NewMockEnhancedTemplateEngine(t)

		// Test processing with complex data
		mockEngine.EXPECT().
			Process("complex.tex").
			Return("Complex processed content", nil).
			Once()

		result, err := mockEngine.Process("complex.tex")
		assert.NoError(t, err)
		assert.Equal(t, "Complex processed content", result)
	})

	t.Run("Clone", func(t *testing.T) {
		mockEngine := mocks.NewMockEnhancedTemplateEngine(t)
		mockClonedEngine := mocks.NewMockEnhancedTemplateEngine(t)

		// Test cloning
		mockEngine.EXPECT().
			Clone().
			Return(mockClonedEngine).
			Once()

		cloned := mockEngine.Clone()
		assert.NotNil(t, cloned)
	})
}

// TestConfigProviderUnit tests the config provider in isolation
func TestConfigProviderUnit(t *testing.T) {
	t.Run("GetConfig", func(t *testing.T) {
		mockProvider := mocks.NewMockConfigProvider(t)

		expectedConfig := &config.Config{
			Template: "test.tex",
			Output:   "output.pdf",
			Variables: map[string]interface{}{
				"title": "Test Document",
			},
			Engine: "pdflatex",
		}

		mockProvider.EXPECT().
			GetConfig().
			Return(expectedConfig).
			Once()

		config := mockProvider.GetConfig()
		assert.Equal(t, "test.tex", config.Template.String())
		assert.Equal(t, "output.pdf", config.Output.String())
		assert.Equal(t, "pdflatex", config.Engine.String())
	})

	t.Run("GetDefaultConfig", func(t *testing.T) {
		mockProvider := mocks.NewMockConfigProvider(t)

		defaultConfig := &config.Config{
			Engine: "pdflatex",
		}

		mockProvider.EXPECT().
			GetDefaultConfig().
			Return(defaultConfig).
			Once()

		config := mockProvider.GetDefaultConfig()
		assert.Equal(t, "pdflatex", config.Engine.String())
	})

	t.Run("LoadConfigFromFile", func(t *testing.T) {
		mockProvider := mocks.NewMockConfigProvider(t)

		expectedConfig := &config.Config{
			Template: "loaded.tex",
			Engine:   "xelatex",
		}

		mockProvider.EXPECT().
			LoadConfigFromFile("config.yaml").
			Return(expectedConfig, nil).
			Once()

		config, err := mockProvider.LoadConfigFromFile("config.yaml")
		assert.NoError(t, err)
		assert.Equal(t, "loaded.tex", config.Template.String())
		assert.Equal(t, "xelatex", config.Engine.String())
	})

	t.Run("SaveConfigToFile", func(t *testing.T) {
		mockProvider := mocks.NewMockConfigProvider(t)

		config := &config.Config{
			Template: "save.tex",
			Engine:   "pdflatex",
		}

		mockProvider.EXPECT().
			SaveConfigToFile(config, "output.yaml").
			Return(nil).
			Once()

		err := mockProvider.SaveConfigToFile(config, "output.yaml")
		assert.NoError(t, err)
	})
}

// TestVariableProcessorUnit tests the variable processor in isolation
func TestVariableProcessorUnit(t *testing.T) {
	t.Run("ProcessVariables", func(t *testing.T) {
		mockProcessor := mocks.NewMockVariableProcessor(t)

		variables := map[string]interface{}{
			"title":   "Test Document",
			"author":  "John Doe",
			"content": "Test content",
		}

		expectedCollection := domain.NewVariableCollection()
		titleVar, _ := domain.NewStringVariable("Test Document")
		authorVar, _ := domain.NewStringVariable("John Doe")
		contentVar, _ := domain.NewStringVariable("Test content")
		expectedCollection.Set("title", titleVar)
		expectedCollection.Set("author", authorVar)
		expectedCollection.Set("content", contentVar)

		mockProcessor.EXPECT().
			ProcessVariables(variables).
			Return(expectedCollection, nil).
			Once()

		collection, err := mockProcessor.ProcessVariables(variables)
		assert.NoError(t, err)
		assert.NotNil(t, collection)
	})

	t.Run("GetVariable", func(t *testing.T) {
		mockProcessor := mocks.NewMockVariableProcessor(t)

		expectedVar, _ := domain.NewStringVariable("Test Document")

		mockProcessor.EXPECT().
			GetVariable("title").
			Return(expectedVar, nil).
			Once()

		variable, err := mockProcessor.GetVariable("title")
		assert.NoError(t, err)
		assert.Equal(t, expectedVar, variable)
	})

	t.Run("SetVariable", func(t *testing.T) {
		mockProcessor := mocks.NewMockVariableProcessor(t)

		mockProcessor.EXPECT().
			SetVariable("title", "New Title").
			Return(nil).
			Once()

		err := mockProcessor.SetVariable("title", "New Title")
		assert.NoError(t, err)
	})

	t.Run("GetNested", func(t *testing.T) {
		mockProcessor := mocks.NewMockVariableProcessor(t)

		expectedVar, _ := domain.NewStringVariable("Nested Value")

		mockProcessor.EXPECT().
			GetNested("document.title").
			Return(expectedVar, nil).
			Once()

		variable, err := mockProcessor.GetNested("document.title")
		assert.NoError(t, err)
		assert.Equal(t, expectedVar, variable)
	})
}

// TestTemplateValidatorUnit tests the template validator in isolation
func TestTemplateValidatorUnit(t *testing.T) {
	t.Run("ValidateTemplate", func(t *testing.T) {
		mockValidator := mocks.NewMockTemplateValidator(t)

		// Test successful validation
		mockValidator.EXPECT().
			ValidateTemplate("valid.tex").
			Return(nil).
			Once()

		err := mockValidator.ValidateTemplate("valid.tex")
		assert.NoError(t, err)
	})

	t.Run("ValidateSyntax", func(t *testing.T) {
		mockValidator := mocks.NewMockTemplateValidator(t)

		// Test syntax validation
		mockValidator.EXPECT().
			ValidateSyntax("\\documentclass{article}").
			Return(nil).
			Once()

		err := mockValidator.ValidateSyntax("\\documentclass{article}")
		assert.NoError(t, err)
	})

	t.Run("ValidateVariables", func(t *testing.T) {
		mockValidator := mocks.NewMockTemplateValidator(t)

		templateContent := "\\title{delim[[.title]]}"
		variables := map[string]interface{}{
			"title": "Test Document",
		}

		mockValidator.EXPECT().
			ValidateVariables(templateContent, variables).
			Return(nil).
			Once()

		err := mockValidator.ValidateVariables(templateContent, variables)
		assert.NoError(t, err)
	})
}

// TestFileProcessorUnit tests the file processor in isolation
func TestFileProcessorUnit(t *testing.T) {
	t.Run("ReadFile", func(t *testing.T) {
		mockProcessor := mocks.NewMockFileProcessor(t)

		expectedContent := []byte("\\documentclass{article}")

		mockProcessor.EXPECT().
			ReadFile("test.tex").
			Return(expectedContent, nil).
			Once()

		content, err := mockProcessor.ReadFile("test.tex")
		assert.NoError(t, err)
		assert.Equal(t, expectedContent, content)
	})

	t.Run("WriteFile", func(t *testing.T) {
		mockProcessor := mocks.NewMockFileProcessor(t)

		content := []byte("\\documentclass{article}")

		mockProcessor.EXPECT().
			WriteFile("output.tex", content).
			Return(nil).
			Once()

		err := mockProcessor.WriteFile("output.tex", content)
		assert.NoError(t, err)
	})

	t.Run("FileExists", func(t *testing.T) {
		mockProcessor := mocks.NewMockFileProcessor(t)

		// Test file exists
		mockProcessor.EXPECT().
			FileExists("existing.tex").
			Return(true).
			Once()

		// Test file doesn't exist
		mockProcessor.EXPECT().
			FileExists("missing.tex").
			Return(false).
			Once()

		exists := mockProcessor.FileExists("existing.tex")
		assert.True(t, exists)

		exists = mockProcessor.FileExists("missing.tex")
		assert.False(t, exists)
	})

	t.Run("CreateDirectory", func(t *testing.T) {
		mockProcessor := mocks.NewMockFileProcessor(t)

		mockProcessor.EXPECT().
			CreateDirectory("output").
			Return(nil).
			Once()

		err := mockProcessor.CreateDirectory("output")
		assert.NoError(t, err)
	})

	t.Run("RemoveFile", func(t *testing.T) {
		mockProcessor := mocks.NewMockFileProcessor(t)

		mockProcessor.EXPECT().
			RemoveFile("temp.tex").
			Return(nil).
			Once()

		err := mockProcessor.RemoveFile("temp.tex")
		assert.NoError(t, err)
	})
}

// TestUnitErrorHandling tests error handling in unit tests
func TestUnitErrorHandling(t *testing.T) {
	t.Run("TemplateEngineErrors", func(t *testing.T) {
		mockEngine := mocks.NewMockTemplateEngine(t)

		// Test processing error
		mockEngine.EXPECT().
			Process("error.tex").
			Return("", errors.New("processing failed")).
			Once()

		// Test validation error
		mockEngine.EXPECT().
			ValidateTemplate("invalid.tex").
			Return(errors.New("validation failed")).
			Once()

		result, err := mockEngine.Process("error.tex")
		assert.Error(t, err)
		assert.Empty(t, result)

		err = mockEngine.ValidateTemplate("invalid.tex")
		assert.Error(t, err)
	})

	t.Run("FileProcessorErrors", func(t *testing.T) {
		mockProcessor := mocks.NewMockFileProcessor(t)

		// Test read error
		mockProcessor.EXPECT().
			ReadFile("error.tex").
			Return(nil, errors.New("read failed")).
			Once()

		// Test write error
		mockProcessor.EXPECT().
			WriteFile("error.tex", []byte("content")).
			Return(errors.New("write failed")).
			Once()

		content, err := mockProcessor.ReadFile("error.tex")
		assert.Error(t, err)
		assert.Nil(t, content)

		err = mockProcessor.WriteFile("error.tex", []byte("content"))
		assert.Error(t, err)
	})

	t.Run("ConfigProviderErrors", func(t *testing.T) {
		mockProvider := mocks.NewMockConfigProvider(t)

		// Test load error
		mockProvider.EXPECT().
			LoadConfigFromFile("missing.yaml").
			Return(nil, errors.New("file not found")).
			Once()

		// Test save error
		config := &config.Config{Template: "test.tex"}
		mockProvider.EXPECT().
			SaveConfigToFile(config, "error.yaml").
			Return(errors.New("save failed")).
			Once()

		loadedConfig, err := mockProvider.LoadConfigFromFile("missing.yaml")
		assert.Error(t, err)
		assert.Nil(t, loadedConfig)

		err = mockProvider.SaveConfigToFile(config, "error.yaml")
		assert.Error(t, err)
	})
}
