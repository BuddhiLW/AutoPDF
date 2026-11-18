// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/backend"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/pdf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// installFakeTypst puts a `typst` script first on PATH. It answers
// --version, and for `compile ... SRC OUT` writes a PDF holding SRC's text
// (or, with mode "truncated", a PDF without its %%EOF trailer) and records
// its arguments in a log.
func installFakeTypst(t *testing.T, mode string) (logPath string) {
	t.Helper()
	bin := t.TempDir()
	logPath = filepath.Join(bin, "typst.log")
	script := `#!/bin/sh
if [ "$1" = "--version" ]; then echo "typst 0.0.0-fake"; exit 0; fi
echo "$@" >> "` + logPath + `"
for arg in "$@"; do src=$out; out=$arg; done
mkdir -p "$(dirname "$out")"
{
  printf '%%PDF-1.7\n'
  cat "$src"
  if [ "` + mode + `" != truncated ]; then printf '%%%%EOF\n'; fi
} > "$out"
`
	require.NoError(t, os.WriteFile(filepath.Join(bin, "typst"), []byte(script), 0o755))
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return logPath
}

func writeTypTemplate(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "note.typ")
	require.NoError(t, os.WriteFile(path, []byte("= delim[[.title]]\nHello #emph[world].\n"), 0o644))
	return path
}

func TestEngineCapabilitiesListTypst(t *testing.T) {
	engine, err := api.NewEngine()
	require.NoError(t, err)
	assert.Contains(t, engine.Capabilities().Engines, "typst")
}

func TestRequestEngineName(t *testing.T) {
	assert.Equal(t, "typst", api.Request{Engine: "typst", LaTeXEngine: "xelatex"}.EngineName(), "Engine wins")
	assert.Equal(t, "xelatex", api.Request{LaTeXEngine: "xelatex"}.EngineName(), "deprecated alias still read")
	assert.Equal(t, "", api.Request{}.EngineName())
}

func TestEngineGeneratesTypstThroughDefaultGenerator(t *testing.T) {
	for name, req := range map[string]api.Request{
		"by .typ extension":              {},
		"Request.Engine":                 {Engine: "typst"},
		"Engine wins over LaTeXEngine":   {Engine: "typst", LaTeXEngine: "pdflatex"},
		"deprecated LaTeXEngine = typst": {LaTeXEngine: "typst"},
	} {
		t.Run(name, func(t *testing.T) {
			logPath := installFakeTypst(t, "complete")
			template := writeTypTemplate(t)
			req.TemplatePath = template
			req.OutputPath = filepath.Join(filepath.Dir(template), "out", "note")
			req.Variables = map[string]string{"title": "Typst Title"}

			engine, err := api.NewEngine()
			require.NoError(t, err)
			result, err := engine.Generate(context.Background(), req)
			require.NoError(t, err)

			assert.Equal(t, req.OutputPath+".pdf", result.PDFPath)
			assert.Contains(t, string(result.PDF), "= Typst Title")
			assert.Equal(t, pdf.DefectNone, pdf.Inspect(result.PDF))
			log, err := os.ReadFile(logPath)
			require.NoError(t, err)
			assert.Contains(t, string(log), "compile --root")
			assert.Contains(t, string(log), ".autopdf.typ")
		})
	}
}

func TestEngineRefusesLaTeXOnlySettingsWithTypst(t *testing.T) {
	logPath := installFakeTypst(t, "complete")
	template := writeTypTemplate(t)
	engine, err := api.NewEngine()
	require.NoError(t, err)

	for name, req := range map[string]api.Request{
		"passes":      {Engine: "typst", Passes: 2},
		"use_latexmk": {UseLatexmk: true},
	} {
		t.Run(name, func(t *testing.T) {
			req.TemplatePath = template
			req.OutputPath = filepath.Join(t.TempDir(), "x.pdf")
			_, err := engine.Generate(context.Background(), req)
			require.Error(t, err)
			assert.ErrorIs(t, err, backend.ErrUnsupportedSetting)
		})
	}
	_, err = os.Stat(logPath)
	assert.True(t, os.IsNotExist(err), "typst must not run for a refused request")
}

func TestEngineRefusesUnfinishedTypstPDF(t *testing.T) {
	installFakeTypst(t, "truncated")
	template := writeTypTemplate(t)
	output := filepath.Join(filepath.Dir(template), "out", "note.pdf")
	engine, err := api.NewEngine()
	require.NoError(t, err)

	_, err = engine.Generate(context.Background(), api.Request{
		TemplatePath: template,
		OutputPath:   output,
		Variables:    map[string]string{"title": "x"},
	})
	require.Error(t, err)
	var incomplete *pdf.IncompleteError
	assert.True(t, errors.As(err, &incomplete), "got %v", err)
	assert.NoFileExists(t, output, "an unfinished PDF is never published")
}

// TestEngineTypstRealBinary compiles a tiny .typ template with the real typst
// binary through pkg/api.Engine. It is skipped when typst is not on PATH.
func TestEngineTypstRealBinary(t *testing.T) {
	if _, err := exec.LookPath("typst"); err != nil {
		t.Skip("typst not on PATH; skipping the real-binary end-to-end test")
	}
	template := writeTypTemplate(t)
	output := filepath.Join(t.TempDir(), "real.pdf")
	engine, err := api.NewEngine()
	require.NoError(t, err)

	result, err := engine.Generate(context.Background(), api.Request{
		TemplatePath: template,
		OutputPath:   output,
		Engine:       "typst",
		Variables:    map[string]string{"title": "Real Typst"},
	})
	require.NoError(t, err)
	assert.Equal(t, output, result.PDFPath)
	assert.True(t, strings.HasPrefix(string(result.PDF), "%PDF-"))
	assert.Equal(t, pdf.DefectNone, pdf.Inspect(result.PDF))
	published, err := os.ReadFile(output)
	require.NoError(t, err)
	assert.Equal(t, result.PDF, published)
}
