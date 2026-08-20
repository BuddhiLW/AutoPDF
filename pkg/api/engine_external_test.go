// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"context"
	"errors"
	"testing"

	"github.com/BuddhiLW/AutoPDF/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngineExternalContract(t *testing.T) {
	type contextKey string
	const requestKey contextKey = "request"

	want := api.Result{PDFPath: "result.pdf", PDF: []byte("%PDF-test")}
	generator := api.GeneratorFunc(func(ctx context.Context, req api.Request) (api.Result, error) {
		assert.Equal(t, "visible-to-generator", ctx.Value(requestKey))
		assert.Equal(t, "template.tex", req.TemplatePath)
		return want, nil
	})

	engine, err := api.NewEngine(api.WithGenerator(generator))
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), requestKey, "visible-to-generator")
	got, err := engine.Generate(ctx, api.Request{
		TemplatePath: "template.tex",
		OutputPath:   "result.pdf",
	})
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestEngineRejectsCancelledContext(t *testing.T) {
	called := false
	engine, err := api.NewEngine(api.WithGenerator(api.GeneratorFunc(
		func(context.Context, api.Request) (api.Result, error) {
			called = true
			return api.Result{}, nil
		},
	)))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = engine.Generate(ctx, api.Request{TemplatePath: "template.tex", OutputPath: "output.pdf"})

	assert.ErrorIs(t, err, context.Canceled)
	assert.False(t, called)
}

func TestEngineOptionsAndCapabilitiesAreSafeForExternalPrograms(t *testing.T) {
	_, err := api.NewEngine(api.WithGenerator(nil))
	assert.ErrorIs(t, err, api.ErrNilGenerator)

	engine, err := api.NewEngine(
		api.WithLogger(nil),
		api.WithCapabilities(api.Capabilities{Engines: []string{"customtex"}}),
		api.WithGenerator(api.GeneratorFunc(func(context.Context, api.Request) (api.Result, error) {
			return api.Result{}, errors.New("extension failure")
		})),
	)
	require.NoError(t, err)

	capabilities := engine.Capabilities()
	capabilities.Engines[0] = "mutated"
	assert.Equal(t, []string{"customtex"}, engine.Capabilities().Engines)

	_, err = engine.Generate(context.Background(), api.Request{TemplatePath: "template.tex", OutputPath: "output.pdf"})
	assert.EqualError(t, err, "extension failure")
}
