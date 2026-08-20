// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package converter

import (
	"fmt"
	"testing"

	"github.com/BuddhiLW/AutoPDF/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type customIdentifier int

type customIdentifierConverter struct{}

func (customIdentifierConverter) CanConvert(value interface{}) bool {
	_, ok := value.(customIdentifier)
	return ok
}

func (customIdentifierConverter) Convert(value interface{}) (config.Variable, error) {
	identifier, ok := value.(customIdentifier)
	if !ok {
		return nil, fmt.Errorf("unexpected value %T", value)
	}
	return &config.StringVariable{Value: fmt.Sprintf("ID-%d", identifier)}, nil
}

func TestConverterBuilderRegistersCustomConverter(t *testing.T) {
	converter, err := NewConverterBuilder().
		WithCustomConverter(customIdentifier(0), customIdentifierConverter{}).
		BuildValidated()
	require.NoError(t, err)

	variables, err := converter.ConvertStruct(customIdentifier(42))
	require.NoError(t, err)
	assert.Equal(t, "ID-42", variables.Flatten()["value"])
}

func TestConverterBuilderReportsInvalidRegistration(t *testing.T) {
	builder := NewConverterBuilder().WithCustomConverter(nil, customIdentifierConverter{})
	assert.Error(t, builder.Err())
	converter, err := builder.BuildValidated()
	assert.Nil(t, converter)
	assert.Error(t, err)
}
