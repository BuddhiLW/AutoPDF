// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package file

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/domain"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/domain/generation"
)

var (
	ErrInspectorRequired = errors.New("PDF inspector capability is required")
	ErrPDFNotFound       = errors.New("PDF file does not exist")
)

// Inspection is the effect-free evidence used by PDFFile validation.
type Inspection struct {
	Size      int64
	ModTime   time.Time
	Header    []byte
	PageCount int
}

// Inspector is implemented at an infrastructure boundary that may access files or tools.
type Inspector interface {
	Inspect(path string) (Inspection, error)
}

// PDFFile is a domain entity whose rules operate only on injected inspection evidence.
type PDFFile struct {
	path      string
	inspector Inspector
	metadata  *generation.PDFMetadata
	errors    []error
}

// NewPDFFile creates a PDF entity. Supply an Inspector before invoking effectful queries.
func NewPDFFile(path string, inspectors ...Inspector) *PDFFile {
	var inspector Inspector
	if len(inspectors) > 0 {
		inspector = inspectors[0]
	}
	return &PDFFile{path: path, inspector: inspector, errors: []error{}}
}

func (pdf *PDFFile) Path() string { return pdf.path }

func (pdf *PDFFile) Exists() bool {
	if pdf.inspector == nil {
		return false
	}
	_, err := pdf.inspector.Inspect(pdf.path)
	return err == nil
}

func (pdf *PDFFile) IsValid() bool { return pdf.Validate() == nil }

func (pdf *PDFFile) Validate() error {
	pdf.errors = pdf.errors[:0]
	if pdf.inspector == nil {
		return pdf.record(ErrInspectorRequired)
	}
	if !strings.EqualFold(extension(pdf.path), ".pdf") {
		return pdf.record(domain.PDFGenerationError{
			Code:    domain.ErrCodePDFValidationFailed,
			Message: api.ErrPDFFileExtensionInvalid,
			Details: api.NewErrorDetails(api.ErrorCategoryPDF, api.ErrorSeverityHigh).WithFilePath(pdf.path),
		})
	}
	inspection, err := pdf.inspector.Inspect(pdf.path)
	if err != nil {
		message := api.ErrPDFFileNotReadable
		if errors.Is(err, ErrPDFNotFound) {
			message = api.ErrPDFFileNotFound
		}
		return pdf.record(domain.PDFGenerationError{
			Code:    domain.ErrCodePDFValidationFailed,
			Message: message,
			Details: api.NewErrorDetails(api.ErrorCategoryPDF, api.ErrorSeverityHigh).WithFilePath(pdf.path).WithError(err),
		})
	}
	if inspection.Size <= 0 {
		return pdf.record(domain.PDFGenerationError{
			Code:    domain.ErrCodePDFValidationFailed,
			Message: api.ErrPDFFileEmpty,
			Details: api.NewErrorDetails(api.ErrorCategoryPDF, api.ErrorSeverityHigh).WithFilePath(pdf.path),
		})
	}
	if len(inspection.Header) < 5 || string(inspection.Header[:5]) != "%PDF-" {
		return pdf.record(domain.PDFGenerationError{
			Code:    domain.ErrCodePDFValidationFailed,
			Message: api.ErrPDFHeaderInvalid,
			Details: api.NewErrorDetails(api.ErrorCategoryPDF, api.ErrorSeverityHigh).WithFilePath(pdf.path),
		})
	}
	pdf.metadata = &generation.PDFMetadata{
		FileSize:    inspection.Size,
		PageCount:   inspection.PageCount,
		GeneratedAt: inspection.ModTime,
	}
	return nil
}

func (pdf *PDFFile) GetMetadata() (*generation.PDFMetadata, error) {
	if pdf.metadata == nil {
		if err := pdf.Validate(); err != nil {
			return nil, err
		}
	}
	metadata := *pdf.metadata
	return &metadata, nil
}

func (pdf *PDFFile) GetPageCount() (int, error) {
	metadata, err := pdf.GetMetadata()
	if err != nil {
		return 0, err
	}
	return metadata.PageCount, nil
}

func (pdf *PDFFile) GetFileSize() (int64, error) {
	metadata, err := pdf.GetMetadata()
	if err != nil {
		return 0, err
	}
	return metadata.FileSize, nil
}

func (pdf *PDFFile) GetModificationTime() (time.Time, error) {
	metadata, err := pdf.GetMetadata()
	if err != nil {
		return time.Time{}, err
	}
	return metadata.GeneratedAt, nil
}

func (pdf *PDFFile) GetErrors() []error { return append([]error(nil), pdf.errors...) }
func (pdf *PDFFile) HasErrors() bool    { return len(pdf.errors) > 0 }

func (pdf *PDFFile) String() string {
	return fmt.Sprintf("PDFFile{path: %s, valid: %t, errors: %d}", pdf.path, pdf.metadata != nil && len(pdf.errors) == 0, len(pdf.errors))
}

func (pdf *PDFFile) record(err error) error {
	pdf.errors = append(pdf.errors, err)
	pdf.metadata = nil
	return err
}

func extension(path string) string {
	index := strings.LastIndex(path, ".")
	if index < 0 {
		return ""
	}
	return path[index:]
}

type PDFFileRepository interface {
	FindByPath(path string) (*PDFFile, error)
	Save(pdf *PDFFile) error
	Delete(pdf *PDFFile) error
}

type PDFFileService interface {
	CreatePDFFile(path string) (*PDFFile, error)
	ValidatePDFFile(pdf *PDFFile) error
	GetPDFMetadata(pdf *PDFFile) (*generation.PDFMetadata, error)
	ConvertToImages(pdf *PDFFile, formats []string) ([]string, error)
}
