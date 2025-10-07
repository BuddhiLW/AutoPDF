// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"os"
	"path/filepath"
	"strings"
)

// ValidationUtils provides reusable validation functions
// These can leverage internal logic from the autopdf application layer
type ValidationUtils struct{}

// NewValidationUtils creates a new ValidationUtils instance
func NewValidationUtils() *ValidationUtils {
	return &ValidationUtils{}
}

// ValidatePDFPath validates a PDF file path
func (vu *ValidationUtils) ValidatePDFPath(pdfPath string) (*ErrorDetails, error) {
	if pdfPath == "" {
		return NewErrorDetails(ErrorCategoryPDF, ErrorSeverityHigh).
				AddContext(ContextKeyError, ErrPDFPathRequired),
			nil
	}

	// Check if path is valid
	if !filepath.IsAbs(pdfPath) && !strings.HasPrefix(pdfPath, "./") {
		// Make it absolute if it's relative
		absPath, err := filepath.Abs(pdfPath)
		if err != nil {
			return NewErrorDetails(ErrorCategoryPDF, ErrorSeverityHigh).
					WithFilePath(pdfPath).
					WithError(err).
					AddContext(ContextKeyError, ErrPDFPathInvalid),
				err
		}
		pdfPath = absPath
	}

	// Check if file exists
	fileInfo, err := os.Stat(pdfPath)
	if os.IsNotExist(err) {
		return NewErrorDetails(ErrorCategoryPDF, ErrorSeverityHigh).
				WithFilePath(pdfPath).
				AddContext(ContextKeyError, ErrPDFFileNotFound),
			nil
	}
	if err != nil {
		return NewErrorDetails(ErrorCategoryPDF, ErrorSeverityHigh).
				WithFilePath(pdfPath).
				WithError(err).
				AddContext(ContextKeyError, ErrPDFFileNotReadable),
			err
	}

	// Check file size
	if fileInfo.Size() == 0 {
		return NewErrorDetails(ErrorCategoryPDF, ErrorSeverityHigh).
				WithFilePath(pdfPath).
				WithFileSize(0).
				AddContext(ContextKeyError, ErrPDFFileEmpty),
			nil
	}

	// Check file extension
	if !strings.HasSuffix(strings.ToLower(pdfPath), ".pdf") {
		return NewErrorDetails(ErrorCategoryPDF, ErrorSeverityHigh).
				WithFilePath(pdfPath).
				WithValidation("file_extension", ".pdf", filepath.Ext(pdfPath)).
				AddContext(ContextKeyError, ErrPDFFileInvalid),
			nil
	}

	// Check file size constraints
	if fileInfo.Size() < MinPDFFileSize {
		return NewErrorDetails(ErrorCategoryPDF, ErrorSeverityMedium).
				WithFilePath(pdfPath).
				WithFileSize(fileInfo.Size()).
				WithValidation("min_file_size", MinPDFFileSize, fileInfo.Size()).
				AddContext(ContextKeyError, ErrPDFFileEmpty),
			nil
	}

	if fileInfo.Size() > MaxPDFFileSize {
		return NewErrorDetails(ErrorCategoryPDF, ErrorSeverityMedium).
				WithFilePath(pdfPath).
				WithFileSize(fileInfo.Size()).
				WithValidation("max_file_size", MaxPDFFileSize, fileInfo.Size()).
				AddContext(ContextKeyError, ErrPDFFileInvalid),
			nil
	}

	return nil, nil
}

// ValidatePDFContent validates PDF content structure
func (vu *ValidationUtils) ValidatePDFContent(pdfPath string) (*ErrorDetails, error) {
	// Open file for reading
	file, err := os.Open(pdfPath)
	if err != nil {
		return NewErrorDetails(ErrorCategoryPDF, ErrorSeverityHigh).
				WithFilePath(pdfPath).
				WithError(err).
				AddContext(ContextKeyError, ErrPDFFileNotReadable),
			err
	}
	defer file.Close()

	// Read PDF header
	buffer := make([]byte, MinPDFHeaderLength)
	n, err := file.Read(buffer)
	if err != nil || n < MinPDFHeaderLength {
		return NewErrorDetails(ErrorCategoryPDF, ErrorSeverityHigh).
				WithFilePath(pdfPath).
				WithError(err).
				WithValidation("header_length", MinPDFHeaderLength, n).
				AddContext(ContextKeyError, ErrPDFHeaderInvalid),
			err
	}

	// Check PDF header signature
	header := string(buffer[:MinPDFHeaderLength])
	if header != "%PDF" {
		return NewErrorDetails(ErrorCategoryPDF, ErrorSeverityHigh).
				WithFilePath(pdfPath).
				WithValidation("pdf_header", "%PDF", header).
				AddContext(ContextKeyError, ErrPDFHeaderInvalid),
			nil
	}

	return nil, nil
}

// ValidateTemplatePath validates a template file path
func (vu *ValidationUtils) ValidateTemplatePath(templatePath string) (*ErrorDetails, error) {
	if templatePath == "" {
		return NewErrorDetails(ErrorCategoryTemplate, ErrorSeverityHigh).
				AddContext(ContextKeyError, ErrTemplatePathRequired),
			nil
	}

	// Check if file exists
	fileInfo, err := os.Stat(templatePath)
	if os.IsNotExist(err) {
		return NewErrorDetails(ErrorCategoryTemplate, ErrorSeverityHigh).
				WithTemplatePath(templatePath).
				AddContext(ContextKeyError, ErrTemplateFileNotFound),
			nil
	}
	if err != nil {
		return NewErrorDetails(ErrorCategoryTemplate, ErrorSeverityHigh).
				WithTemplatePath(templatePath).
				WithError(err).
				AddContext(ContextKeyError, ErrTemplateFileNotReadable),
			err
	}

	// Check file size
	if fileInfo.Size() == 0 {
		return NewErrorDetails(ErrorCategoryTemplate, ErrorSeverityHigh).
				WithTemplatePath(templatePath).
				WithFileSize(0).
				AddContext(ContextKeyError, ErrTemplateFileEmpty),
			nil
	}

	// Check file extension
	if !strings.HasSuffix(strings.ToLower(templatePath), ".tex") {
		return NewErrorDetails(ErrorCategoryTemplate, ErrorSeverityHigh).
				WithTemplatePath(templatePath).
				WithValidation("file_extension", ".tex", filepath.Ext(templatePath)).
				AddContext(ContextKeyError, ErrTemplateFileInvalid),
			nil
	}

	// Check file size constraints
	if fileInfo.Size() < MinTemplateFileSize {
		return NewErrorDetails(ErrorCategoryTemplate, ErrorSeverityMedium).
				WithTemplatePath(templatePath).
				WithFileSize(fileInfo.Size()).
				WithValidation("min_file_size", MinTemplateFileSize, fileInfo.Size()).
				AddContext(ContextKeyError, ErrTemplateFileEmpty),
			nil
	}

	if fileInfo.Size() > MaxTemplateFileSize {
		return NewErrorDetails(ErrorCategoryTemplate, ErrorSeverityMedium).
				WithTemplatePath(templatePath).
				WithFileSize(fileInfo.Size()).
				WithValidation("max_file_size", MaxTemplateFileSize, fileInfo.Size()).
				AddContext(ContextKeyError, ErrTemplateFileInvalid),
			nil
	}

	return nil, nil
}

// ValidateOutputPath validates an output file path
func (vu *ValidationUtils) ValidateOutputPath(outputPath string) (*ErrorDetails, error) {
	if outputPath == "" {
		return NewErrorDetails(ErrorCategoryGeneration, ErrorSeverityHigh).
				AddContext(ContextKeyError, ErrOutputPathRequired),
			nil
	}

	// Check if directory exists or can be created
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return NewErrorDetails(ErrorCategoryGeneration, ErrorSeverityHigh).
				WithOutputPath(outputPath).
				WithError(err).
				AddContext(ContextKeyError, ErrOutputPathInvalid),
			err
	}

	// Check if file already exists
	if _, err := os.Stat(outputPath); err == nil {
		return NewErrorDetails(ErrorCategoryGeneration, ErrorSeverityMedium).
				WithOutputPath(outputPath).
				AddContext(ContextKeyError, ErrOutputPathExists),
			nil
	}

	return nil, nil
}

// ValidateEngine validates a LaTeX engine
func (vu *ValidationUtils) ValidateEngine(engine string) (*ErrorDetails, error) {
	if engine == "" {
		return NewErrorDetails(ErrorCategoryGeneration, ErrorSeverityHigh).
				AddContext(ContextKeyError, ErrEngineRequired),
			nil
	}

	// Check if engine is supported
	supportedEngines := []string{"pdflatex", "xelatex", "lualatex", "latex"}
	isSupported := false
	for _, supportedEngine := range supportedEngines {
		if engine == supportedEngine {
			isSupported = true
			break
		}
	}

	if !isSupported {
		return NewErrorDetails(ErrorCategoryGeneration, ErrorSeverityHigh).
				WithEngine(engine).
				WithValidation("supported_engines", supportedEngines, engine).
				AddContext(ContextKeyError, ErrEngineNotSupported),
			nil
	}

	return nil, nil
}

// GetFileMetadata extracts file metadata
func (vu *ValidationUtils) GetFileMetadata(filePath string) (map[string]interface{}, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	metadata := map[string]interface{}{
		"file_size": fileInfo.Size(),
		"mod_time":  fileInfo.ModTime(),
		"is_dir":    fileInfo.IsDir(),
		"mode":      fileInfo.Mode(),
		"name":      fileInfo.Name(),
		"path":      filePath,
	}

	return metadata, nil
}

// CreateRecoverySuggestions creates recovery suggestions based on error type
func (vu *ValidationUtils) CreateRecoverySuggestions(category, errorCode string) []string {
	suggestions := []string{}

	switch category {
	case ErrorCategoryPDF:
		switch errorCode {
		case ErrPDFFileNotFound:
			suggestions = append(suggestions, "Check if the file path is correct")
			suggestions = append(suggestions, "Verify the file exists in the specified location")
			suggestions = append(suggestions, "Check file permissions")
		case ErrPDFFileEmpty:
			suggestions = append(suggestions, "Regenerate the PDF file")
			suggestions = append(suggestions, "Check if the source file is corrupted")
		case ErrPDFHeaderInvalid:
			suggestions = append(suggestions, "Verify the file is a valid PDF")
			suggestions = append(suggestions, "Try regenerating the PDF")
		}
	case ErrorCategoryTemplate:
		switch errorCode {
		case ErrTemplateFileNotFound:
			suggestions = append(suggestions, "Check if the template file exists")
			suggestions = append(suggestions, "Verify the file path is correct")
		case ErrTemplateFileEmpty:
			suggestions = append(suggestions, "Check if the template file has content")
			suggestions = append(suggestions, "Verify the template is not corrupted")
		}
	case ErrorCategoryGeneration:
		switch errorCode {
		case ErrEngineNotFound:
			suggestions = append(suggestions, "Install the required LaTeX engine")
			suggestions = append(suggestions, "Check if the engine is in the PATH")
		case ErrOutputPathInvalid:
			suggestions = append(suggestions, "Check if the output directory exists")
			suggestions = append(suggestions, "Verify write permissions")
		}
	}

	return suggestions
}
