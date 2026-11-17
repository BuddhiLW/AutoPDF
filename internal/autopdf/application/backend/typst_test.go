// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package backend_test

import (
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/typst"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/backend"
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypstBackendDescriptor(t *testing.T) {
	b := backend.Typst()
	assert.Equal(t, "typst", b.Name)
	assert.Equal(t, []string{"typst"}, b.EngineNames)
	assert.Equal(t, []string{".typ"}, b.TemplateExtensions)
	assert.Equal(t, typst.StagedSuffix, b.StagedSuffix)
	assert.Empty(t, b.AuxExtensions)

	registered, ok := backend.Default().Lookup("typst")
	require.True(t, ok)
	assert.Equal(t, b.EngineNames, registered.EngineNames)
}

func TestTypstResolvesByEngineAndExtension(t *testing.T) {
	registry := backend.NewBuiltinRegistry()
	for _, tc := range []struct{ engine, template string }{
		{"typst", "doc.tex"},
		{"", "dir/Doc.TYP"},
		{"", "doc.typ"},
	} {
		b, err := registry.Resolve(tc.engine, tc.template)
		require.NoError(t, err)
		assert.Equal(t, "typst", b.Name, "%+v", tc)
	}
	b, err := registry.Resolve("pdflatex", "doc.typ")
	require.NoError(t, err)
	assert.Equal(t, "latex", b.Name, "an explicit engine wins over the extension")
}

func TestTypstOptionsRefuseLaTeXOnlySettings(t *testing.T) {
	b := backend.Typst()

	opts, err := b.Options(backend.Settings{Passes: 1})
	require.NoError(t, err)
	assert.Equal(t, ports.DefaultTypstOptions(), opts)

	opts, err = b.Options(backend.Settings{})
	require.NoError(t, err)
	assert.Equal(t, ports.DefaultTypstOptions(), opts)

	_, err = b.Options(backend.Settings{Passes: 2})
	assert.ErrorIs(t, err, backend.ErrUnsupportedSetting)
	var unsupported *backend.UnsupportedSettingError
	require.ErrorAs(t, err, &unsupported)
	assert.Equal(t, "passes", unsupported.Setting)

	_, err = b.Options(backend.Settings{Passes: 1, UseLatexmk: true})
	assert.ErrorIs(t, err, backend.ErrUnsupportedSetting)
	require.ErrorAs(t, err, &unsupported)
	assert.Equal(t, "use_latexmk", unsupported.Setting)
}

func TestTypstCompilerFactory(t *testing.T) {
	b := backend.Typst()
	compiler, err := b.Compiler(backend.CompilerDeps{WorkingDir: "/w"}, nil)
	require.NoError(t, err)
	assert.IsType(t, &typst.TypstCompilerAdapter{}, compiler)

	_, err = b.Compiler(backend.CompilerDeps{}, ports.NewLaTeXOptions(2, false))
	assert.ErrorIs(t, err, ports.ErrBackendOptionsMismatch)
}
