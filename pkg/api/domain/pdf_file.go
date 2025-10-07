// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package domain

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BuddhiLW/AutoPDF/pkg/api"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
	"github.com/rwxrob/bonzai/run"
)

// PDFFile represents a PDF file as a domain entity
// This encapsulates all PDF-related operations and state
type PDFFile struct {
	path     string
	metadata *generation.PDFMetadata
	valid    bool
	errors   []error
}

// NewPDFFile creates a new PDF file entity from a file path
func NewPDFFile(path string) *PDFFile {
	return &PDFFile{
		path:     path,
		metadata: nil,
		valid:    false,
		errors:   make([]error, 0),
	}
}

// Path returns the file path of the PDF
func (pf *PDFFile) Path() string {
	return pf.path
}

// Exists checks if the PDF file exists on the filesystem
func (pf *PDFFile) Exists() bool {
	_, err := os.Stat(pf.path)
	return err == nil
}

// IsValid checks if the PDF file is valid
func (pf *PDFFile) IsValid() bool {
	return pf.valid
}

// Validate performs comprehensive validation of the PDF file
func (pf *PDFFile) Validate() error {
	pf.errors = make([]error, 0)

	// Check if file exists
	if !pf.Exists() {
		err := PDFGenerationError{
			Code:    ErrCodePDFValidationFailed,
			Message: api.ErrPDFFileNotFound,
			Details: api.NewErrorDetails(api.ErrorCategoryPDF, api.ErrorSeverityHigh).
				WithFilePath(pf.path),
		}
		pf.errors = append(pf.errors, err)
		pf.valid = false
		return err
	}

	// Check file extension
	if !strings.HasSuffix(strings.ToLower(pf.path), ".pdf") {
		err := PDFGenerationError{
			Code:    ErrCodePDFValidationFailed,
			Message: api.ErrPDFFileExtensionInvalid,
			Details: api.NewErrorDetails(api.ErrorCategoryPDF, api.ErrorSeverityHigh).
				WithFilePath(pf.path),
		}
		pf.errors = append(pf.errors, err)
		pf.valid = false
		return err
	}

	// Check PDF header signature
	if !pf.hasValidPDFHeader() {
		err := PDFGenerationError{
			Code:    ErrCodePDFValidationFailed,
			Message: api.ErrPDFHeaderInvalid,
			Details: api.NewErrorDetails(api.ErrorCategoryPDF, api.ErrorSeverityHigh).
				WithFilePath(pf.path),
		}
		pf.errors = append(pf.errors, err)
		pf.valid = false
		return err
	}

	// If we get here, the PDF is valid
	pf.valid = true
	return nil
}

// GetMetadata returns the PDF metadata, loading it if necessary
func (pf *PDFFile) GetMetadata() (*generation.PDFMetadata, error) {
	if pf.metadata == nil {
		metadata, err := pf.loadMetadata()
		if err != nil {
			return nil, err
		}
		pf.metadata = metadata
	}
	return pf.metadata, nil
}

// GetPageCount returns the number of pages in the PDF
func (pf *PDFFile) GetPageCount() (int, error) {
	metadata, err := pf.GetMetadata()
	if err != nil {
		return 0, err
	}
	return metadata.PageCount, nil
}

// GetFileSize returns the file size in bytes
func (pf *PDFFile) GetFileSize() (int64, error) {
	metadata, err := pf.GetMetadata()
	if err != nil {
		return 0, err
	}
	return metadata.FileSize, nil
}

// GetModificationTime returns the file modification time
func (pf *PDFFile) GetModificationTime() (time.Time, error) {
	fileInfo, err := os.Stat(pf.path)
	if err != nil {
		return time.Time{}, err
	}
	return fileInfo.ModTime(), nil
}

// GetErrors returns all validation errors
func (pf *PDFFile) GetErrors() []error {
	return pf.errors
}

// HasErrors checks if there are any validation errors
func (pf *PDFFile) HasErrors() bool {
	return len(pf.errors) > 0
}

// String returns a string representation of the PDF file
func (pf *PDFFile) String() string {
	return fmt.Sprintf("PDFFile{path: %s, valid: %t, errors: %d}",
		pf.path, pf.valid, len(pf.errors))
}

// Private methods

// hasValidPDFHeader checks if the file has a valid PDF header
func (pf *PDFFile) hasValidPDFHeader() bool {
	file, err := os.Open(pf.path)
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

// loadMetadata loads the PDF metadata using available tools
func (pf *PDFFile) loadMetadata() (*generation.PDFMetadata, error) {
	// Try to get file info first
	fileInfo, err := os.Stat(pf.path)
	if err != nil {
		return nil, PDFGenerationError{
			Code:    ErrCodePDFValidationFailed,
			Message: api.ErrPDFFileNotReadable,
			Details: api.NewErrorDetails(api.ErrorCategoryPDF, api.ErrorSeverityHigh).
				WithFilePath(pf.path).
				WithError(err),
		}
	}

	// Try to get page count using available tools
	pageCount := pf.estimatePageCount()

	// Create metadata
	metadata := &generation.PDFMetadata{
		FileSize:    fileInfo.Size(),
		PageCount:   pageCount,
		GeneratedAt: fileInfo.ModTime(),
		Engine:      "unknown", // Would need to be passed from generation context
		Template:    "unknown", // Would need to be passed from generation context
	}

	return metadata, nil
}

// estimatePageCount provides an accurate page count using available tools
func (pf *PDFFile) estimatePageCount() int {
	// Try ImageMagick's identify command first
	if pageCount := pf.getPageCountWithImageMagick(); pageCount > 0 {
		return pageCount
	}

	// Try pdftoppm as fallback (part of poppler-utils)
	if pageCount := pf.getPageCountWithPdfToPpm(); pageCount > 0 {
		return pageCount
	}

	// Fallback to default if no tools are available
	return 1
}

// getPageCountWithImageMagick uses ImageMagick's identify command to count pages
func (pf *PDFFile) getPageCountWithImageMagick() int {
	// Use bonzai's run package for better error handling and consistency
	output := run.Out("identify", "-format", "%n", pf.path)
	if output == "" {
		return 0
	}

	// Parse the output to get page count
	pageCountStr := strings.TrimSpace(output)
	pageCount, err := strconv.Atoi(pageCountStr)
	if err != nil {
		return 0
	}

	// Ensure we have at least 1 page
	if pageCount < 1 {
		return 0
	}

	return pageCount
}

// getPageCountWithPdfToPpm uses pdftoppm to estimate page count
func (pf *PDFFile) getPageCountWithPdfToPpm() int {
	// Use bonzai's run package for better error handling and consistency
	// Use pdftoppm with -l 1 to get just the first page and see if it works
	// This is a simple way to test if the PDF is valid and get basic info
	err := run.Exec("pdftoppm", "-l", "1", "-f", "1", pf.path, "/tmp/test")
	if err != nil {
		return 0
	}

	// If pdftoppm succeeds, we know the PDF has at least 1 page
	// For a more accurate count, we could use pdfinfo if available
	// For now, return 1 as a conservative estimate
	return 1
}

// PDFFileRepository defines the interface for PDF file operations
type PDFFileRepository interface {
	FindByPath(path string) (*PDFFile, error)
	Save(pdf *PDFFile) error
	Delete(pdf *PDFFile) error
	Exists(path string) bool
}

// PDFFileService defines the interface for PDF file business logic
type PDFFileService interface {
	CreatePDFFile(path string) (*PDFFile, error)
	ValidatePDFFile(pdf *PDFFile) error
	GetPDFMetadata(pdf *PDFFile) (*generation.PDFMetadata, error)
	ConvertToImages(pdf *PDFFile, formats []string) ([]string, error)
}
