// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package pdf_validator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOSPDFInspectorStaysAtAdapterBoundary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "document.pdf")
	require.NoError(t, os.WriteFile(path, []byte("%PDF-1.4\n/Type /Page\n/Type /Page\n"), 0644))

	validator := NewPDFValidatorAdapter()
	require.NoError(t, validator.Validate(path))
	metadata, err := validator.GetMetadata(path)
	require.NoError(t, err)
	assert.Equal(t, 2, metadata.PageCount)
}
