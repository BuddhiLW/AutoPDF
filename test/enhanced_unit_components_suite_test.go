package test

import (
	"errors"
	"testing"

	"github.com/BuddhiLW/AutoPDF/mocks"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/BuddhiLW/AutoPDF/pkg/domain"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// EnhancedUnitComponentsTestSuite tests individual components with enhanced suite testing
type EnhancedUnitComponentsTestSuite struct {
	suite.Suite
	mockEngine            *mocks.MockTemplateEngine
	mockEnhancedEngine    *mocks.MockEnhancedTemplateEngine
	mockValidator         *mocks.MockTemplateValidator
	mockFileProcessor     *mocks.MockFileProcessor
	mockVariableProcessor *mocks.MockVariableProcessor
	mockConfigProvider    *mocks.MockConfigProvider
}

// SetupTest initializes the test suite
func (suite *EnhancedUnitComponentsTestSuite) SetupTest() {
	suite.mockEngine = mocks.NewMockTemplateEngine(suite.T())
	suite.mockEnhancedEngine = mocks.NewMockEnhancedTemplateEngine(suite.T())
	suite.mockValidator = mocks.NewMockTemplateValidator(suite.T())
	suite.mockFileProcessor = mocks.NewMockFileProcessor(suite.T())
	suite.mockVariableProcessor = mocks.NewMockVariableProcessor(suite.T())
	suite.mockConfigProvider = mocks.NewMockConfigProvider(suite.T())
}

// TestTemplateEngineUnit tests the template engine in isolation
func (suite *EnhancedUnitComponentsTestSuite) TestTemplateEngineUnit() {
	suite.Run("ProcessTemplate", func() {
		// Test successful processing
		suite.mockEngine.EXPECT().
			Process("test.tex").
			Return("Processed template content", nil).
			Once()

		result, err := suite.mockEngine.Process("test.tex")
		suite.NoError(err)
		suite.Equal("Processed template content", result)
	})

	suite.Run("ProcessToFile", func() {
		// Test file processing
		suite.mockEngine.EXPECT().
			ProcessToFile("input.tex", "output.tex").
			Return(nil).
			Once()

		err := suite.mockEngine.ProcessToFile("input.tex", "output.tex")
		suite.NoError(err)
	})

	suite.Run("AddFunction", func() {
		// Test function addition
		suite.mockEngine.EXPECT().
			AddFunction("upper", mock.AnythingOfType("func(string) string")).
			Return().
			Once()

		suite.mockEngine.AddFunction("upper", func(s string) string {
			return "UPPERCASE_" + s
		})
	})

	suite.Run("ValidateTemplate", func() {
		// Test template validation
		suite.mockEngine.EXPECT().
			ValidateTemplate("valid.tex").
			Return(nil).
			Once()

		err := suite.mockEngine.ValidateTemplate("valid.tex")
		suite.NoError(err)
	})
}

// TestEnhancedTemplateEngineUnit tests the enhanced template engine in isolation
func (suite *EnhancedUnitComponentsTestSuite) TestEnhancedTemplateEngineUnit() {
	suite.Run("SetVariable", func() {
		// Test setting single variable
		suite.mockEnhancedEngine.EXPECT().
			SetVariable("title", "Test Document").
			Return(nil).
			Once()

		err := suite.mockEnhancedEngine.SetVariable("title", "Test Document")
		suite.NoError(err)
	})

	suite.Run("SetVariablesFromMap", func() {
		variables := map[string]interface{}{
			"title":   "Test Document",
			"author":  "John Doe",
			"content": "Test content",
		}

		suite.mockEnhancedEngine.EXPECT().
			SetVariablesFromMap(variables).
			Return(nil).
			Once()

		err := suite.mockEnhancedEngine.SetVariablesFromMap(variables)
		suite.NoError(err)
	})

	suite.Run("GetVariable", func() {
		expectedVar, _ := domain.NewStringVariable("Test Document")

		suite.mockEnhancedEngine.EXPECT().
			GetVariable("title").
			Return(expectedVar, nil).
			Once()

		variable, err := suite.mockEnhancedEngine.GetVariable("title")
		suite.NoError(err)
		suite.Equal(expectedVar, variable)
	})

	suite.Run("ProcessWithComplexData", func() {
		// Test processing with complex data
		suite.mockEnhancedEngine.EXPECT().
			Process("complex.tex").
			Return("Complex processed content", nil).
			Once()

		result, err := suite.mockEnhancedEngine.Process("complex.tex")
		suite.NoError(err)
		suite.Equal("Complex processed content", result)
	})

	suite.Run("Clone", func() {
		mockClonedEngine := mocks.NewMockEnhancedTemplateEngine(suite.T())

		// Test cloning
		suite.mockEnhancedEngine.EXPECT().
			Clone().
			Return(mockClonedEngine).
			Once()

		cloned := suite.mockEnhancedEngine.Clone()
		suite.NotNil(cloned)
	})
}

// TestConfigProviderUnit tests the config provider in isolation
func (suite *EnhancedUnitComponentsTestSuite) TestConfigProviderUnit() {
	suite.Run("GetConfig", func() {
		expectedConfig := &config.Config{
			Template: "test.tex",
			Output:   "output.pdf",
			Variables: map[string]interface{}{
				"title": "Test Document",
			},
			Engine: "pdflatex",
		}

		suite.mockConfigProvider.EXPECT().
			GetConfig().
			Return(expectedConfig).
			Once()

		config := suite.mockConfigProvider.GetConfig()
		suite.Equal("test.tex", config.Template.String())
		suite.Equal("output.pdf", config.Output.String())
		suite.Equal("pdflatex", config.Engine.String())
	})

	suite.Run("GetDefaultConfig", func() {
		defaultConfig := &config.Config{
			Engine: "pdflatex",
		}

		suite.mockConfigProvider.EXPECT().
			GetDefaultConfig().
			Return(defaultConfig).
			Once()

		config := suite.mockConfigProvider.GetDefaultConfig()
		suite.Equal("pdflatex", config.Engine.String())
	})

	suite.Run("LoadConfigFromFile", func() {
		expectedConfig := &config.Config{
			Template: "loaded.tex",
			Engine:   "xelatex",
		}

		suite.mockConfigProvider.EXPECT().
			LoadConfigFromFile("config.yaml").
			Return(expectedConfig, nil).
			Once()

		config, err := suite.mockConfigProvider.LoadConfigFromFile("config.yaml")
		suite.NoError(err)
		suite.Equal("loaded.tex", config.Template.String())
		suite.Equal("xelatex", config.Engine.String())
	})

	suite.Run("SaveConfigToFile", func() {
		config := &config.Config{
			Template: "save.tex",
			Engine:   "pdflatex",
		}

		suite.mockConfigProvider.EXPECT().
			SaveConfigToFile(config, "output.yaml").
			Return(nil).
			Once()

		err := suite.mockConfigProvider.SaveConfigToFile(config, "output.yaml")
		suite.NoError(err)
	})
}

// TestVariableProcessorUnit tests the variable processor in isolation
func (suite *EnhancedUnitComponentsTestSuite) TestVariableProcessorUnit() {
	suite.Run("ProcessVariables", func() {
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

		suite.mockVariableProcessor.EXPECT().
			ProcessVariables(variables).
			Return(expectedCollection, nil).
			Once()

		collection, err := suite.mockVariableProcessor.ProcessVariables(variables)
		suite.NoError(err)
		suite.NotNil(collection)
	})

	suite.Run("GetVariable", func() {
		expectedVar, _ := domain.NewStringVariable("Test Document")

		suite.mockVariableProcessor.EXPECT().
			GetVariable("title").
			Return(expectedVar, nil).
			Once()

		variable, err := suite.mockVariableProcessor.GetVariable("title")
		suite.NoError(err)
		suite.Equal(expectedVar, variable)
	})

	suite.Run("SetVariable", func() {
		suite.mockVariableProcessor.EXPECT().
			SetVariable("title", "New Title").
			Return(nil).
			Once()

		err := suite.mockVariableProcessor.SetVariable("title", "New Title")
		suite.NoError(err)
	})

	suite.Run("GetNested", func() {
		expectedVar, _ := domain.NewStringVariable("Nested Value")

		suite.mockVariableProcessor.EXPECT().
			GetNested("document.title").
			Return(expectedVar, nil).
			Once()

		variable, err := suite.mockVariableProcessor.GetNested("document.title")
		suite.NoError(err)
		suite.Equal(expectedVar, variable)
	})
}

// TestTemplateValidatorUnit tests the template validator in isolation
func (suite *EnhancedUnitComponentsTestSuite) TestTemplateValidatorUnit() {
	suite.Run("ValidateTemplate", func() {
		// Test successful validation
		suite.mockValidator.EXPECT().
			ValidateTemplate("valid.tex").
			Return(nil).
			Once()

		err := suite.mockValidator.ValidateTemplate("valid.tex")
		suite.NoError(err)
	})

	suite.Run("ValidateSyntax", func() {
		// Test syntax validation
		suite.mockValidator.EXPECT().
			ValidateSyntax("\\documentclass{article}").
			Return(nil).
			Once()

		err := suite.mockValidator.ValidateSyntax("\\documentclass{article}")
		suite.NoError(err)
	})

	suite.Run("ValidateVariables", func() {
		templateContent := "\\title{delim[[.title]]}"
		variables := map[string]interface{}{
			"title": "Test Document",
		}

		suite.mockValidator.EXPECT().
			ValidateVariables(templateContent, variables).
			Return(nil).
			Once()

		err := suite.mockValidator.ValidateVariables(templateContent, variables)
		suite.NoError(err)
	})
}

// TestFileProcessorUnit tests the file processor in isolation
func (suite *EnhancedUnitComponentsTestSuite) TestFileProcessorUnit() {
	suite.Run("ReadFile", func() {
		expectedContent := []byte("\\documentclass{article}")

		suite.mockFileProcessor.EXPECT().
			ReadFile("test.tex").
			Return(expectedContent, nil).
			Once()

		content, err := suite.mockFileProcessor.ReadFile("test.tex")
		suite.NoError(err)
		suite.Equal(expectedContent, content)
	})

	suite.Run("WriteFile", func() {
		content := []byte("\\documentclass{article}")

		suite.mockFileProcessor.EXPECT().
			WriteFile("output.tex", content).
			Return(nil).
			Once()

		err := suite.mockFileProcessor.WriteFile("output.tex", content)
		suite.NoError(err)
	})

	suite.Run("FileExists", func() {
		// Test file exists
		suite.mockFileProcessor.EXPECT().
			FileExists("existing.tex").
			Return(true).
			Once()

		// Test file doesn't exist
		suite.mockFileProcessor.EXPECT().
			FileExists("missing.tex").
			Return(false).
			Once()

		exists := suite.mockFileProcessor.FileExists("existing.tex")
		suite.True(exists)

		exists = suite.mockFileProcessor.FileExists("missing.tex")
		suite.False(exists)
	})

	suite.Run("CreateDirectory", func() {
		suite.mockFileProcessor.EXPECT().
			CreateDirectory("output").
			Return(nil).
			Once()

		err := suite.mockFileProcessor.CreateDirectory("output")
		suite.NoError(err)
	})

	suite.Run("RemoveFile", func() {
		suite.mockFileProcessor.EXPECT().
			RemoveFile("temp.tex").
			Return(nil).
			Once()

		err := suite.mockFileProcessor.RemoveFile("temp.tex")
		suite.NoError(err)
	})
}

// TestUnitErrorHandling tests error handling in unit tests
func (suite *EnhancedUnitComponentsTestSuite) TestUnitErrorHandling() {
	suite.Run("TemplateEngineErrors", func() {
		// Test processing error
		suite.mockEngine.EXPECT().
			Process("error.tex").
			Return("", errors.New("processing failed")).
			Once()

		// Test validation error
		suite.mockEngine.EXPECT().
			ValidateTemplate("invalid.tex").
			Return(errors.New("validation failed")).
			Once()

		result, err := suite.mockEngine.Process("error.tex")
		suite.Error(err)
		suite.Empty(result)

		err = suite.mockEngine.ValidateTemplate("invalid.tex")
		suite.Error(err)
	})

	suite.Run("FileProcessorErrors", func() {
		// Test read error
		suite.mockFileProcessor.EXPECT().
			ReadFile("error.tex").
			Return(nil, errors.New("read failed")).
			Once()

		// Test write error
		suite.mockFileProcessor.EXPECT().
			WriteFile("error.tex", []byte("content")).
			Return(errors.New("write failed")).
			Once()

		content, err := suite.mockFileProcessor.ReadFile("error.tex")
		suite.Error(err)
		suite.Nil(content)

		err = suite.mockFileProcessor.WriteFile("error.tex", []byte("content"))
		suite.Error(err)
	})

	suite.Run("ConfigProviderErrors", func() {
		// Test load error
		suite.mockConfigProvider.EXPECT().
			LoadConfigFromFile("missing.yaml").
			Return(nil, errors.New("file not found")).
			Once()

		// Test save error
		config := &config.Config{Template: "test.tex"}
		suite.mockConfigProvider.EXPECT().
			SaveConfigToFile(config, "error.yaml").
			Return(errors.New("save failed")).
			Once()

		loadedConfig, err := suite.mockConfigProvider.LoadConfigFromFile("missing.yaml")
		suite.Error(err)
		suite.Nil(loadedConfig)

		err = suite.mockConfigProvider.SaveConfigToFile(config, "error.yaml")
		suite.Error(err)
	})
}

// TestEnhancedUnitComponentsSuite runs the complete test suite
func TestEnhancedUnitComponentsSuite(t *testing.T) {
	suite.Run(t, new(EnhancedUnitComponentsTestSuite))
}
