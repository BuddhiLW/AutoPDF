// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package compilation

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/BuddhiLW/AutoPDF/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanHandleClaimsLaTeXTemplatesOnly(t *testing.T) {
	strategy := NewEngineStrategy(api.GeneratorFunc(
		func(context.Context, api.Request) (api.Result, error) { return api.Result{}, nil }))

	assert.True(t, strategy.CanHandle("report.tex"))
	assert.True(t, strategy.CanHandle("/abs/path/Report.TEX"), "extension match is case-insensitive")
	assert.True(t, strategy.CanHandle("thesis.latex"))
	assert.False(t, strategy.CanHandle("notes.md"))
	assert.False(t, strategy.CanHandle("noextension"))
}

func TestCanHandleIsFalseWithoutAGenerator(t *testing.T) {
	strategy := NewEngineStrategy(nil)
	assert.False(t, strategy.CanHandle("report.tex"),
		"a strategy that cannot compile must not claim work")
}

func TestCompilePassesTemplateAndDerivedOutputToGenerator(t *testing.T) {
	var seen api.Request
	strategy := NewEngineStrategy(
		api.GeneratorFunc(func(_ context.Context, request api.Request) (api.Result, error) {
			seen = request
			return api.Result{PDFPath: request.OutputPath}, nil
		}),
		WithLaTeXEngine("xelatex"),
	)

	result, err := strategy.Compile(context.Background(), "docs/report.tex", "docs/autopdf.yaml")
	require.NoError(t, err)

	assert.Equal(t, "docs/report.tex", seen.TemplatePath)
	assert.Equal(t, filepath.Join("docs", "report.pdf"), seen.OutputPath,
		"output defaults to a .pdf beside the template")
	assert.Equal(t, "xelatex", seen.LaTeXEngine)
	assert.Equal(t, "docs", seen.WorkingDir)

	assert.Equal(t, "docs/report.tex", result.TemplateFile)
	assert.Equal(t, filepath.Join("docs", "report.pdf"), result.PDFPath)
	assert.NotZero(t, result.Timestamp)
}

func TestCompileHonoursOutputDirOverride(t *testing.T) {
	var seen api.Request
	strategy := NewEngineStrategy(
		api.GeneratorFunc(func(_ context.Context, request api.Request) (api.Result, error) {
			seen = request
			return api.Result{PDFPath: request.OutputPath}, nil
		}),
		WithOutputDir("build/pdf"),
	)

	_, err := strategy.Compile(context.Background(), "docs/report.tex", "autopdf.yaml")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("build", "pdf", "report.pdf"), seen.OutputPath)
}

func TestCompilePropagatesGeneratorFailure(t *testing.T) {
	wanted := errors.New("pdflatex exited 1")
	strategy := NewEngineStrategy(api.GeneratorFunc(
		func(context.Context, api.Request) (api.Result, error) { return api.Result{}, wanted }))

	result, err := strategy.Compile(context.Background(), "report.tex", "autopdf.yaml")
	assert.Nil(t, result)
	assert.ErrorIs(t, err, wanted)
}
