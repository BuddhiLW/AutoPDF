// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type capturedLogger struct {
	mutex  sync.Mutex
	values []string
}

func (logger *capturedLogger) capture(message string, fields ...api.LogField) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	logger.values = append(logger.values, message, fmt.Sprint(fields))
}

func (logger *capturedLogger) Debug(_ context.Context, message string, fields ...api.LogField) {
	logger.capture(message, fields...)
}
func (logger *capturedLogger) Info(_ context.Context, message string, fields ...api.LogField) {
	logger.capture(message, fields...)
}
func (logger *capturedLogger) Warn(_ context.Context, message string, fields ...api.LogField) {
	logger.capture(message, fields...)
}
func (logger *capturedLogger) Error(_ context.Context, message string, fields ...api.LogField) {
	logger.capture(message, fields...)
}

func (logger *capturedLogger) String() string {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	return strings.Join(logger.values, " ")
}

func TestDefaultObservabilityDoesNotLogVariableValues(t *testing.T) {
	logger := &capturedLogger{}
	engine, err := api.NewEngine(
		api.WithLogger(logger),
		api.WithGenerator(api.GeneratorFunc(func(context.Context, api.Request) (api.Result, error) {
			return api.Result{PDFPath: "output.pdf", PDF: []byte("%PDF-")}, nil
		})),
	)
	require.NoError(t, err)

	_, err = engine.Generate(context.Background(), api.Request{
		TemplatePath: "template.tex",
		OutputPath:   "output.pdf",
		Variables:    map[string]string{"access_token": "top-secret-value"},
	})
	require.NoError(t, err)
	assert.NotContains(t, logger.String(), "access_token")
	assert.NotContains(t, logger.String(), "top-secret-value")
}

func TestErrorLoggingRedactsContextValues(t *testing.T) {
	logger := &capturedLogger{}
	api.NewErrorDetails(api.ErrorCategoryGeneration, api.ErrorSeverityHigh).
		AddContext("access_token", "top-secret-value").
		LogError(logger)

	assert.NotContains(t, logger.String(), "access_token")
	assert.NotContains(t, logger.String(), "top-secret-value")
}
