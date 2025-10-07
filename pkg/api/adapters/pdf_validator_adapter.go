// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"os"
	"strings"
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/api/domain"
)

// PDFValidatorAdapter implements domain.PDFValidator
type PDFValidatorAdapter struct{}

// NewPDFValidatorAdapter creates a new PDF validator adapter
func NewPDFValidatorAdapter() *PDFValidatorAdapter {
	return &PDFValidatorAdapter{}
}

// Validate validates a PDF file
func (pva *PDFValidatorAdapter) Validate(pdfPath string) error {
	if pdfPath == "" {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodePDFValidationFailed,
			Message: "PDF path is required",
		}
	}

	// Check if file exists
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodePDFValidationFailed,
			Message: "PDF file does not exist",
			Details: map[string]interface{}{"pdf_path": pdfPath},
		}
	}

	// Check file extension
	if !strings.HasSuffix(strings.ToLower(pdfPath), ".pdf") {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodePDFValidationFailed,
			Message: "File is not a PDF",
			Details: map[string]interface{}{"pdf_path": pdfPath},
		}
	}

	// Check if file is readable
	file, err := os.Open(pdfPath)
	if err != nil {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodePDFValidationFailed,
			Message: "Cannot read PDF file",
			Details: map[string]interface{}{"pdf_path": pdfPath, "error": err.Error()},
		}
	}
	defer file.Close()

	// Check file size
	fileInfo, err := file.Stat()
	if err != nil {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodePDFValidationFailed,
			Message: "Cannot get file info",
			Details: map[string]interface{}{"pdf_path": pdfPath, "error": err.Error()},
		}
	}

	if fileInfo.Size() == 0 {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodePDFValidationFailed,
			Message: "PDF file is empty",
			Details: map[string]interface{}{"pdf_path": pdfPath},
		}
	}

	// Basic PDF header validation
	if !pva.IsValidPDF(pdfPath) {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodePDFValidationFailed,
			Message: "Invalid PDF format",
			Details: map[string]interface{}{"pdf_path": pdfPath},
		}
	}

	return nil
}

// GetMetadata extracts metadata from a PDF file
func (pva *PDFValidatorAdapter) GetMetadata(pdfPath string) (domain.PDFMetadata, error) {
	if pdfPath == "" {
		return domain.PDFMetadata{}, domain.PDFGenerationError{
			Code:    domain.ErrCodePDFValidationFailed,
			Message: "PDF path is required",
		}
	}

	// Get file info
	fileInfo, err := os.Stat(pdfPath)
	if err != nil {
		return domain.PDFMetadata{}, domain.PDFGenerationError{
			Code:    domain.ErrCodePDFValidationFailed,
			Message: "Cannot get file info",
			Details: map[string]interface{}{"pdf_path": pdfPath, "error": err.Error()},
		}
	}

	// Basic metadata extraction
	metadata := domain.PDFMetadata{
		FileSize:    fileInfo.Size(),
		GeneratedAt: fileInfo.ModTime(),
		Engine:      "unknown", // Would need to be passed from generation context
		Template:    "unknown", // Would need to be passed from generation context
		PageCount:   pva.estimatePageCount(pdfPath),
	}

	return metadata, nil
}

// IsValidPDF checks if a file is a valid PDF
func (pva *PDFValidatorAdapter) IsValidPDF(pdfPath string) bool {
	if pdfPath == "" {
		return false
	}

	// Check file extension
	if !strings.HasSuffix(strings.ToLower(pdfPath), ".pdf") {
		return false
	}

	// Check if file exists and is readable
	file, err := os.Open(pdfPath)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read first few bytes to check PDF header
	buffer := make([]byte, 8)
	n, err := file.Read(buffer)
	if err != nil || n < 4 {
		return false
	}

	// Check for PDF header signature
	header := string(buffer[:4])
	return header == "%PDF"
}

// estimatePageCount provides a rough estimate of page count
// This is a simplified implementation - in practice, you'd use a PDF library
func (pva *PDFValidatorAdapter) estimatePageCount(pdfPath string) int {
	// This is a placeholder implementation
	// In a real implementation, you would:
	// 1. Parse the PDF structure
	// 2. Count page objects
	// 3. Or use a PDF library like unidoc/unipdf

	// For now, return a default value
	// A more sophisticated implementation would require a PDF parsing library
	return 1
}

// GetFileSize returns the file size in bytes
func (pva *PDFValidatorAdapter) GetFileSize(pdfPath string) (int64, error) {
	if pdfPath == "" {
		return 0, domain.PDFGenerationError{
			Code:    domain.ErrCodePDFValidationFailed,
			Message: "PDF path is required",
		}
	}

	fileInfo, err := os.Stat(pdfPath)
	if err != nil {
		return 0, domain.PDFGenerationError{
			Code:    domain.ErrCodePDFValidationFailed,
			Message: "Cannot get file info",
			Details: map[string]interface{}{"pdf_path": pdfPath, "error": err.Error()},
		}
	}

	return fileInfo.Size(), nil
}

// GetFileModTime returns the file modification time
func (pva *PDFValidatorAdapter) GetFileModTime(pdfPath string) (time.Time, error) {
	if pdfPath == "" {
		return time.Time{}, domain.PDFGenerationError{
			Code:    domain.ErrCodePDFValidationFailed,
			Message: "PDF path is required",
		}
	}

	fileInfo, err := os.Stat(pdfPath)
	if err != nil {
		return time.Time{}, domain.PDFGenerationError{
			Code:    domain.ErrCodePDFValidationFailed,
			Message: "Cannot get file info",
			Details: map[string]interface{}{"pdf_path": pdfPath, "error": err.Error()},
		}
	}

	return fileInfo.ModTime(), nil
}

// ValidatePDFStructure performs basic structural validation
func (pva *PDFValidatorAdapter) ValidatePDFStructure(pdfPath string) error {
	if !pva.IsValidPDF(pdfPath) {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodePDFValidationFailed,
			Message: "Invalid PDF structure",
			Details: map[string]interface{}{"pdf_path": pdfPath},
		}
	}

	// Additional validation could be added here:
	// - Check for required PDF objects
	// - Validate PDF version
	// - Check for corruption
	// - Validate page structure

	return nil
}
