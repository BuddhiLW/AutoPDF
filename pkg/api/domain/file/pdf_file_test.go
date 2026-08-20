// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package file

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubInspector struct {
	inspection Inspection
	err        error
}

func (stub stubInspector) Inspect(string) (Inspection, error) {
	return stub.inspection, stub.err
}

func TestPDFFileValidatesInjectedEvidenceWithoutIO(t *testing.T) {
	generatedAt := time.Date(2026, time.August, 20, 0, 0, 0, 0, time.UTC)
	pdf := NewPDFFile("document.pdf", stubInspector{inspection: Inspection{
		Size: 128, Header: []byte("%PDF-1.4"), PageCount: 3, ModTime: generatedAt,
	}})

	require.NoError(t, pdf.Validate())
	metadata, err := pdf.GetMetadata()
	require.NoError(t, err)
	assert.Equal(t, int64(128), metadata.FileSize)
	assert.Equal(t, 3, metadata.PageCount)
	assert.Equal(t, generatedAt, metadata.GeneratedAt)
}

func TestPDFFileRequiresInspectionCapability(t *testing.T) {
	err := NewPDFFile("document.pdf").Validate()
	assert.ErrorIs(t, err, ErrInspectorRequired)
}

func TestPDFFileMapsMissingInspectionEvidence(t *testing.T) {
	pdf := NewPDFFile("missing.pdf", stubInspector{err: ErrPDFNotFound})
	err := pdf.Validate()
	require.Error(t, err)
	assert.True(t, errors.Is(pdf.GetErrors()[0], err))
}
