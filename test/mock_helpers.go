package test

import (
	"testing"

	"github.com/BuddhiLW/AutoPDF/mocks"
	"github.com/stretchr/testify/assert"
)

// MockSuite provides a comprehensive set of mocks for testing
type MockSuite struct {
	TemplateEngine         *mocks.MockTemplateEngine
	EnhancedTemplateEngine *mocks.MockEnhancedTemplateEngine
	ConfigProvider         *mocks.MockConfigProvider
	VariableProcessor      *mocks.MockVariableProcessor
	TemplateValidator      *mocks.MockTemplateValidator
	FileProcessor          *mocks.MockFileProcessor
}

// NewMockSuite creates a new mock suite with all necessary mocks
func NewMockSuite(t *testing.T) *MockSuite {
	return &MockSuite{
		TemplateEngine:         mocks.NewMockTemplateEngine(t),
		EnhancedTemplateEngine: mocks.NewMockEnhancedTemplateEngine(t),
		ConfigProvider:         mocks.NewMockConfigProvider(t),
		VariableProcessor:      mocks.NewMockVariableProcessor(t),
		TemplateValidator:      mocks.NewMockTemplateValidator(t),
		FileProcessor:          mocks.NewMockFileProcessor(t),
	}
}

// SetupBasicMocks configures basic mock expectations for common test scenarios
func (ms *MockSuite) SetupBasicMocks() {
	// Basic template engine expectations
	ms.TemplateEngine.EXPECT().
		Process("test.tex").
		Return("processed content", nil).
		Maybe()

	// Basic file processor expectations
	ms.FileProcessor.EXPECT().
		ReadFile("test.tex").
		Return([]byte("template content"), nil).
		Maybe()

	ms.FileProcessor.EXPECT().
		WriteFile("output.pdf", []byte("processed content")).
		Return(nil).
		Maybe()

	// Basic validator expectations
	ms.TemplateValidator.EXPECT().
		ValidateTemplate("test.tex").
		Return(nil).
		Maybe()

	// Basic variable processor expectations
	ms.VariableProcessor.EXPECT().
		ProcessVariables(map[string]interface{}{"key": "value"}).
		Return(nil, nil).
		Maybe()
}

// SetupErrorMocks configures mock expectations for error scenarios
func (ms *MockSuite) SetupErrorMocks() {
	// Template engine error expectations
	ms.TemplateEngine.EXPECT().
		Process("error.tex").
		Return("", assert.AnError).
		Maybe()

	// File processor error expectations
	ms.FileProcessor.EXPECT().
		ReadFile("error.tex").
		Return(nil, assert.AnError).
		Maybe()

	// Validator error expectations
	ms.TemplateValidator.EXPECT().
		ValidateTemplate("error.tex").
		Return(assert.AnError).
		Maybe()
}

// SetupPerformanceMocks configures mock expectations for performance testing
func (ms *MockSuite) SetupPerformanceMocks() {
	// High-performance template processing
	ms.TemplateEngine.EXPECT().
		Process("large.tex").
		Return("large processed content", nil).
		Maybe()

	// Large file processing
	ms.FileProcessor.EXPECT().
		ReadFile("large.tex").
		Return(make([]byte, 1024*1024), nil). // 1MB file
		Maybe()

	// Batch variable processing
	ms.VariableProcessor.EXPECT().
		ProcessVariables(map[string]interface{}{"batch": "data"}).
		Return(nil, nil).
		Maybe()
}

// SetupCartasBackendMocks configures mocks for cartas-backend specific scenarios
func (ms *MockSuite) SetupCartasBackendMocks() {
	// Funeral letter template processing
	ms.TemplateEngine.EXPECT().
		Process("funeral_letter.tex").
		Return("processed funeral letter", nil).
		Maybe()

	// Funeral letter variables
	ms.VariableProcessor.EXPECT().
		ProcessVariables(map[string]interface{}{
			"deceased_name": "John Doe",
			"funeral_date":  "2024-01-15",
			"cemetery":      "Peaceful Gardens",
		}).
		Return(nil, nil).
		Maybe()
}

// SetupEditalPdfApiMocks configures mocks for edital-pdf-api specific scenarios
func (ms *MockSuite) SetupEditalPdfApiMocks() {
	// Legal document template processing
	ms.TemplateEngine.EXPECT().
		Process("edital_template.tex").
		Return("processed legal document", nil).
		Maybe()

	// Legal document variables
	ms.VariableProcessor.EXPECT().
		ProcessVariables(map[string]interface{}{
			"auction_date":   "2024-02-01",
			"property_value": 500000.0,
			"legal_notice":   "Public auction notice",
			"court_info":     map[string]interface{}{"name": "Superior Court", "address": "123 Court St"},
		}).
		Return(nil, nil).
		Maybe()
}

// SetupIntegrationMocks configures mocks for integration testing scenarios
func (ms *MockSuite) SetupIntegrationMocks() {
	// End-to-end document generation
	ms.TemplateEngine.EXPECT().
		Process("integration.tex").
		Return("integration result", nil).
		Maybe()

	// File system operations
	ms.FileProcessor.EXPECT().
		ReadFile("integration.tex").
		Return([]byte("integration template"), nil).
		Maybe()

	ms.FileProcessor.EXPECT().
		WriteFile("integration_output.pdf", []byte("integration result")).
		Return(nil).
		Maybe()

	// Configuration management
	ms.ConfigProvider.EXPECT().
		GetConfig().
		Return(nil).
		Maybe()
}

// SetupBackwardCompatibilityMocks configures mocks for backward compatibility testing
func (ms *MockSuite) SetupBackwardCompatibilityMocks() {
	// Original API compatibility
	ms.TemplateEngine.EXPECT().
		Process("legacy.tex").
		Return("legacy processed", nil).
		Maybe()

	// Legacy variable processing
	ms.VariableProcessor.EXPECT().
		ProcessVariables(map[string]interface{}{"legacy_key": "legacy_value"}).
		Return(nil, nil).
		Maybe()

	// Legacy file operations
	ms.FileProcessor.EXPECT().
		ReadFile("legacy.tex").
		Return([]byte("legacy template"), nil).
		Maybe()
}

// ResetMocks resets all mocks to their initial state
func (ms *MockSuite) ResetMocks() {
	ms.TemplateEngine.ExpectedCalls = nil
	ms.EnhancedTemplateEngine.ExpectedCalls = nil
	ms.ConfigProvider.ExpectedCalls = nil
	ms.VariableProcessor.ExpectedCalls = nil
	ms.TemplateValidator.ExpectedCalls = nil
	ms.FileProcessor.ExpectedCalls = nil
}

// AssertExpectations asserts that all mock expectations were met
func (ms *MockSuite) AssertExpectations(t *testing.T) {
	ms.TemplateEngine.AssertExpectations(t)
	ms.EnhancedTemplateEngine.AssertExpectations(t)
	ms.ConfigProvider.AssertExpectations(t)
	ms.VariableProcessor.AssertExpectations(t)
	ms.TemplateValidator.AssertExpectations(t)
	ms.FileProcessor.AssertExpectations(t)
}
