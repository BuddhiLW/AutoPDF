// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package builders

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestBuilderReportsStructConversionErrors(t *testing.T) {
	builder := NewPDFGenerationRequestBuilder().WithVariablesFromStruct(func() {})

	_, err := builder.BuildValidated()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "convert struct variables")
	assert.Error(t, builder.Err())
}

func TestRequestBuilderPreservesComplexVariableKey(t *testing.T) {
	request, err := NewPDFGenerationRequestBuilder().
		WithComplexVariable("profile", map[string]interface{}{"name": "Ada", "active": true}).
		BuildValidated()
	require.NoError(t, err)
	assert.Equal(t, "Ada", request.Variables.Flatten()["profile.name"])
	assert.Equal(t, "true", request.Variables.Flatten()["profile.active"])
}

func TestConfigBuilderBuildsComplexVariables(t *testing.T) {
	cfg, err := NewConfigBuilder().
		WithComplexVariable("profile", map[string]interface{}{"name": "Ada"}).
		BuildValidated()
	require.NoError(t, err)
	assert.Equal(t, "Ada", cfg.Variables.Flatten()["profile.name"])
}
