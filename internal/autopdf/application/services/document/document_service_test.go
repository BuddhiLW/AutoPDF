// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package document

import (
	"context"
	"errors"
	"testing"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing

type MockTemplateProcessor struct {
	mock.Mock
}

func (m *MockTemplateProcessor) Process(ctx context.Context, templatePath string, variables map[string]string) (string, error) {
	args := m.Called(ctx, templatePath, variables)
	return args.String(0), args.Error(1)
}

type MockLaTeXCompiler struct {
	mock.Mock
}

func (m *MockLaTeXCompiler) Compile(ctx context.Context, content string, engine string, outputPath string) (string, error) {
	args := m.Called(ctx, content, engine, outputPath)
	return args.String(0), args.Error(1)
}

type MockConverter struct {
	mock.Mock
}

func (m *MockConverter) ConvertToImages(ctx context.Context, pdfPath string, formats []string) ([]string, error) {
	args := m.Called(ctx, pdfPath, formats)
	return args.Get(0).([]string), args.Error(1)
}

type MockCleaner struct {
	mock.Mock
}

func (m *MockCleaner) Clean(ctx context.Context, pdfPath string) error {
	args := m.Called(ctx, pdfPath)
	return args.Error(0)
}

// Tests

func TestDocumentService_Build_Success(t *testing.T) {
	// Arrange
	mockTpl := new(MockTemplateProcessor)
	mockTex := new(MockLaTeXCompiler)
	mockConv := new(MockConverter)
	mockClean := new(MockCleaner)

	svc := DocumentService{
		TemplateProcessor: mockTpl,
		LaTeXCompiler:     mockTex,
		Converter:         mockConv,
		Cleaner:           mockClean,
	}

	ctx := context.Background()
	variables := config.NewVariables()
	variables.SetString("title", "Test")

	req := BuildRequest{
		TemplatePath: "template.tex",
		Variables:    variables,
		Engine:       "pdflatex",
		OutputPath:   "output.pdf",
		DoConvert:    false,
		DoClean:      false,
	}

	mockTpl.On("Process", ctx, "template.tex", mock.Anything).Return("\\documentclass{article}...", nil)
	mockTex.On("Compile", ctx, "\\documentclass{article}...", "pdflatex", "output.pdf").Return("output.pdf", nil)

	// Act
	result, err := svc.Build(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "output.pdf", result.PDFPath)
	assert.Empty(t, result.ImagePaths)
	mockTpl.AssertExpectations(t)
	mockTex.AssertExpectations(t)
}

func TestDocumentService_Build_TemplateProcessingFails(t *testing.T) {
	// Arrange
	mockTpl := new(MockTemplateProcessor)
	mockTex := new(MockLaTeXCompiler)
	mockConv := new(MockConverter)
	mockClean := new(MockCleaner)

	svc := DocumentService{
		TemplateProcessor: mockTpl,
		LaTeXCompiler:     mockTex,
		Converter:         mockConv,
		Cleaner:           mockClean,
	}

	ctx := context.Background()
	variables := config.NewVariables()
	variables.SetString("title", "Test")

	req := BuildRequest{
		TemplatePath: "template.tex",
		Variables:    variables,
		Engine:       "pdflatex",
		OutputPath:   "output.pdf",
	}

	mockTpl.On("Process", ctx, "template.tex", mock.Anything).Return("", errors.New("template error"))

	// Act
	result, err := svc.Build(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error.Error(), "template processing failed")
	mockTpl.AssertExpectations(t)
}

func TestDocumentService_Build_LaTeXCompilationFails(t *testing.T) {
	// Arrange
	mockTpl := new(MockTemplateProcessor)
	mockTex := new(MockLaTeXCompiler)
	mockConv := new(MockConverter)
	mockClean := new(MockCleaner)

	svc := DocumentService{
		TemplateProcessor: mockTpl,
		LaTeXCompiler:     mockTex,
		Converter:         mockConv,
		Cleaner:           mockClean,
	}

	ctx := context.Background()
	variables := config.NewVariables()
	variables.SetString("title", "Test")

	req := BuildRequest{
		TemplatePath: "template.tex",
		Variables:    variables,
		Engine:       "pdflatex",
		OutputPath:   "output.pdf",
	}

	mockTpl.On("Process", ctx, "template.tex", mock.Anything).Return("\\documentclass{article}...", nil)
	mockTex.On("Compile", ctx, "\\documentclass{article}...", "pdflatex", "output.pdf").Return("", errors.New("compilation error"))

	// Act
	result, err := svc.Build(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error.Error(), "LaTeX compilation failed")
	mockTpl.AssertExpectations(t)
	mockTex.AssertExpectations(t)
}

func TestDocumentService_Build_WithConversion(t *testing.T) {
	// Arrange
	mockTpl := new(MockTemplateProcessor)
	mockTex := new(MockLaTeXCompiler)
	mockConv := new(MockConverter)
	mockClean := new(MockCleaner)

	svc := DocumentService{
		TemplateProcessor: mockTpl,
		LaTeXCompiler:     mockTex,
		Converter:         mockConv,
		Cleaner:           mockClean,
	}

	ctx := context.Background()
	variables := config.NewVariables()
	variables.SetString("title", "Test")

	req := BuildRequest{
		TemplatePath: "template.tex",
		Variables:    variables,
		Engine:       "pdflatex",
		OutputPath:   "output.pdf",
		DoConvert:    true,
		Conversion: ConversionSettings{
			Enabled: true,
			Formats: []string{"png", "jpg"},
		},
	}

	mockTpl.On("Process", ctx, "template.tex", mock.Anything).Return("\\documentclass{article}...", nil)
	mockTex.On("Compile", ctx, "\\documentclass{article}...", "pdflatex", "output.pdf").Return("output.pdf", nil)
	mockConv.On("ConvertToImages", ctx, "output.pdf", []string{"png", "jpg"}).Return([]string{"output.png", "output.jpg"}, nil)

	// Act
	result, err := svc.Build(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "output.pdf", result.PDFPath)
	assert.Equal(t, []string{"output.png", "output.jpg"}, result.ImagePaths)
	mockTpl.AssertExpectations(t)
	mockTex.AssertExpectations(t)
	mockConv.AssertExpectations(t)
}

func TestDocumentService_Build_WithClean(t *testing.T) {
	// Arrange
	mockTpl := new(MockTemplateProcessor)
	mockTex := new(MockLaTeXCompiler)
	mockConv := new(MockConverter)
	mockClean := new(MockCleaner)

	svc := DocumentService{
		TemplateProcessor: mockTpl,
		LaTeXCompiler:     mockTex,
		Converter:         mockConv,
		Cleaner:           mockClean,
	}

	ctx := context.Background()
	variables := config.NewVariables()
	variables.SetString("title", "Test")

	req := BuildRequest{
		TemplatePath: "template.tex",
		Variables:    variables,
		Engine:       "pdflatex",
		OutputPath:   "output.pdf",
		DoClean:      true,
	}

	mockTpl.On("Process", ctx, "template.tex", mock.Anything).Return("\\documentclass{article}...", nil)
	mockTex.On("Compile", ctx, "\\documentclass{article}...", "pdflatex", "output.pdf").Return("output.pdf", nil)
	mockClean.On("Clean", ctx, "output.pdf").Return(nil)

	// Act
	result, err := svc.Build(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "output.pdf", result.PDFPath)
	mockTpl.AssertExpectations(t)
	mockTex.AssertExpectations(t)
	mockClean.AssertExpectations(t)
}

func TestDocumentService_ConvertDocument(t *testing.T) {
	// Arrange
	mockConv := new(MockConverter)

	svc := DocumentService{
		Converter: mockConv,
	}

	ctx := context.Background()
	pdfPath := "output.pdf"
	formats := []string{"png", "jpg"}

	mockConv.On("ConvertToImages", ctx, pdfPath, formats).Return([]string{"output.png", "output.jpg"}, nil)

	// Act
	imagePaths, err := svc.ConvertDocument(ctx, pdfPath, formats)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, []string{"output.png", "output.jpg"}, imagePaths)
	mockConv.AssertExpectations(t)
}

func TestDocumentService_CleanDocument(t *testing.T) {
	// Arrange
	mockClean := new(MockCleaner)

	svc := DocumentService{
		Cleaner: mockClean,
	}

	ctx := context.Background()
	pdfPath := "output.pdf"

	mockClean.On("Clean", ctx, pdfPath).Return(nil)

	// Act
	err := svc.CleanDocument(ctx, pdfPath)

	// Assert
	assert.NoError(t, err)
	mockClean.AssertExpectations(t)
}
