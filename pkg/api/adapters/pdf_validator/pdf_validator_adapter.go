// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package pdf_validator

import (
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/file"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
)

// PDFValidatorAdapter implements domain.PDFValidator
type PDFValidatorAdapter struct{}

// NewPDFValidatorAdapter creates a new PDF validator adapter
func NewPDFValidatorAdapter() *PDFValidatorAdapter {
	return &PDFValidatorAdapter{}
}

// Validate validates a PDF file using the PDFFile domain entity
func (pva *PDFValidatorAdapter) Validate(pdfPath string) error {
	// Create PDF file domain entity
	pdfFile := file.NewPDFFile(pdfPath)

	// Use the domain entity's validation logic
	return pdfFile.Validate()
}

// GetMetadata extracts metadata from a PDF file using the PDFFile domain entity
func (pva *PDFValidatorAdapter) GetMetadata(pdfPath string) (generation.PDFMetadata, error) {
	// Create PDF file domain entity
	pdfFile := file.NewPDFFile(pdfPath)

	// Get metadata using the domain entity
	metadata, err := pdfFile.GetMetadata()
	if err != nil {
		return generation.PDFMetadata{}, err
	}

	return *metadata, nil
}

// IsValidPDF checks if a file is a valid PDF using the PDFFile domain entity
func (pva *PDFValidatorAdapter) IsValidPDF(pdfPath string) bool {
	// Create PDF file domain entity
	pdfFile := file.NewPDFFile(pdfPath)

	// Use the domain entity's validation logic
	err := pdfFile.Validate()
	return err == nil
}

// GetPageCount returns the page count using the PDFFile domain entity
func (pva *PDFValidatorAdapter) GetPageCount(pdfPath string) (int, error) {
	// Create PDF file domain entity
	pdfFile := file.NewPDFFile(pdfPath)

	// Get page count using the domain entity
	return pdfFile.GetPageCount()
}

// GetFileSize returns the file size using the PDFFile domain entity
func (pva *PDFValidatorAdapter) GetFileSize(pdfPath string) (int64, error) {
	// Create PDF file domain entity
	pdfFile := file.NewPDFFile(pdfPath)

	// Get file size using the domain entity
	return pdfFile.GetFileSize()
}

// GetFileModTime returns the file modification time using the PDFFile domain entity
func (pva *PDFValidatorAdapter) GetFileModTime(pdfPath string) (time.Time, error) {
	// Create PDF file domain entity
	pdfFile := file.NewPDFFile(pdfPath)

	// Get modification time using the domain entity
	return pdfFile.GetModificationTime()
}

// ValidatePDFStructure performs basic structural validation using the PDFFile domain entity
func (pva *PDFValidatorAdapter) ValidatePDFStructure(pdfPath string) error {
	// Create PDF file domain entity
	pdfFile := file.NewPDFFile(pdfPath)

	// Use the domain entity's validation logic
	return pdfFile.Validate()
}
