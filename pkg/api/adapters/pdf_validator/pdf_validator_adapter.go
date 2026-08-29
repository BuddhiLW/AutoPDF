// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package pdf_validator

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/domain/file"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/domain/generation"
)

// PDFValidatorAdapter implements domain.PDFValidator
type PDFValidatorAdapter struct {
	inspector file.Inspector
}

// NewPDFValidatorAdapter creates a new PDF validator adapter
func NewPDFValidatorAdapter() *PDFValidatorAdapter {
	return &PDFValidatorAdapter{inspector: OSPDFInspector{}}
}

// NewPDFValidatorAdapterWithInspector injects a custom inspection boundary.
func NewPDFValidatorAdapterWithInspector(inspector file.Inspector) *PDFValidatorAdapter {
	return &PDFValidatorAdapter{inspector: inspector}
}

// Validate validates a PDF file using the PDFFile domain entity
func (pva *PDFValidatorAdapter) Validate(pdfPath string) error {
	// Create PDF file domain entity
	pdfFile := file.NewPDFFile(pdfPath, pva.inspector)

	// Use the domain entity's validation logic
	return pdfFile.Validate()
}

// GetMetadata extracts metadata from a PDF file using the PDFFile domain entity
func (pva *PDFValidatorAdapter) GetMetadata(pdfPath string) (generation.PDFMetadata, error) {
	// Create PDF file domain entity
	pdfFile := file.NewPDFFile(pdfPath, pva.inspector)

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
	pdfFile := file.NewPDFFile(pdfPath, pva.inspector)

	// Use the domain entity's validation logic
	err := pdfFile.Validate()
	return err == nil
}

// GetPageCount returns the page count using the PDFFile domain entity
func (pva *PDFValidatorAdapter) GetPageCount(pdfPath string) (int, error) {
	// Create PDF file domain entity
	pdfFile := file.NewPDFFile(pdfPath, pva.inspector)

	// Get page count using the domain entity
	return pdfFile.GetPageCount()
}

// GetFileSize returns the file size using the PDFFile domain entity
func (pva *PDFValidatorAdapter) GetFileSize(pdfPath string) (int64, error) {
	// Create PDF file domain entity
	pdfFile := file.NewPDFFile(pdfPath, pva.inspector)

	// Get file size using the domain entity
	return pdfFile.GetFileSize()
}

// GetFileModTime returns the file modification time using the PDFFile domain entity
func (pva *PDFValidatorAdapter) GetFileModTime(pdfPath string) (time.Time, error) {
	// Create PDF file domain entity
	pdfFile := file.NewPDFFile(pdfPath, pva.inspector)

	// Get modification time using the domain entity
	return pdfFile.GetModificationTime()
}

// ValidatePDFStructure performs basic structural validation using the PDFFile domain entity
func (pva *PDFValidatorAdapter) ValidatePDFStructure(pdfPath string) error {
	// Create PDF file domain entity
	pdfFile := file.NewPDFFile(pdfPath, pva.inspector)

	// Use the domain entity's validation logic
	return pdfFile.Validate()
}

// OSPDFInspector is the filesystem boundary for PDF inspection.
type OSPDFInspector struct{}

func (OSPDFInspector) Inspect(path string) (file.Inspection, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return file.Inspection{}, fmt.Errorf("%w: %s", file.ErrPDFNotFound, path)
		}
		return file.Inspection{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return file.Inspection{}, err
	}
	headerLength := min(len(data), 8)
	pageCount := bytes.Count(data, []byte("/Type /Page")) - bytes.Count(data, []byte("/Type /Pages"))
	if pageCount < 1 {
		pageCount = 1
	}
	return file.Inspection{
		Size:      info.Size(),
		ModTime:   info.ModTime(),
		Header:    append([]byte(nil), data[:headerLength]...),
		PageCount: pageCount,
	}, nil
}
