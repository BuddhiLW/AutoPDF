// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package backend_test

import (
	"context"
	"errors"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/latex"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/backend"
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubCompiler struct{ name string }

func (s stubCompiler) Compile(context.Context, string, ports.CompileOptions) (string, error) {
	return s.name, nil
}

type fakeOptions struct{}

func (fakeOptions) BackendName() string { return "fake" }

// FakeBackend is a backend no production code knows about.
func FakeBackend() backend.Backend {
	return backend.Backend{
		Name:               "fake",
		EngineNames:        []string{"fakeengine"},
		TemplateExtensions: []string{".FAKE"},
		StagedSuffix:       ".autopdf.fake",
		AuxExtensions:      []string{".fakeaux"},
		NewCompiler: func(backend.CompilerDeps, ports.BackendOptions) (ports.DocumentCompiler, error) {
			return stubCompiler{name: "fake"}, nil
		},
		NewOptions: func(s backend.Settings) (ports.BackendOptions, error) {
			if s.UseLatexmk {
				return nil, errors.New("fake backend has no latexmk")
			}
			return fakeOptions{}, nil
		},
	}
}

func TestDefaultRegistryHoldsOnlyLaTeX(t *testing.T) {
	registry := backend.Default()
	require.Len(t, registry.Backends(), 1)
	assert.Equal(t, []string{"pdflatex", "xelatex", "lualatex"}, registry.EngineNames())
	assert.Equal(t, []string{".tex", ".latex"}, registry.TemplateExtensions())
	assert.Equal(t, "pdflatex", registry.DefaultEngine())
	assert.Contains(t, registry.AuxExtensions(), ".aux")
	b, ok := registry.Lookup(ports.LaTeXBackendName)
	require.True(t, ok)
	assert.Equal(t, latex.StagedSuffix, b.StagedSuffix)
}

func TestResolveOrder(t *testing.T) {
	registry := backend.NewBuiltinRegistry().MustRegister(FakeBackend())

	t.Run("engine name wins over extension", func(t *testing.T) {
		b, err := registry.Resolve("xelatex", "doc.fake")
		require.NoError(t, err)
		assert.Equal(t, "latex", b.Name)
	})
	t.Run("extension when engine is empty, case-insensitive", func(t *testing.T) {
		b, err := registry.Resolve("", "dir/Doc.Fake")
		require.NoError(t, err)
		assert.Equal(t, "fake", b.Name)
	})
	t.Run("default latex when nothing matches", func(t *testing.T) {
		b, err := registry.Resolve("", "notes.md")
		require.NoError(t, err)
		assert.Equal(t, "latex", b.Name)
		assert.Equal(t, "pdflatex", b.EngineOrDefault(""))
	})
	t.Run("engine given as an executable path", func(t *testing.T) {
		b, err := registry.Resolve("/usr/bin/xelatex", "doc.fake")
		require.NoError(t, err)
		assert.Equal(t, "latex", b.Name, "base name matches an engine")

		b, err = registry.Resolve("/tmp/fake-latex-0", "doc.fake")
		require.NoError(t, err)
		assert.Equal(t, "fake", b.Name, "an unknown binary path falls through to the extension")
	})
	t.Run("unknown engine is a named error listing known engines", func(t *testing.T) {
		_, err := registry.Resolve("context", "doc.tex")
		require.Error(t, err)
		assert.ErrorIs(t, err, backend.ErrUnknownEngine)
		var unknown *backend.UnknownEngineError
		require.ErrorAs(t, err, &unknown)
		assert.Equal(t, "context", unknown.Engine)
		assert.Equal(t, []string{"pdflatex", "xelatex", "lualatex", "fakeengine"}, unknown.Known)
		assert.Contains(t, err.Error(), "fakeengine")
	})
}

func TestRegisterRefusesConflicts(t *testing.T) {
	registry := backend.NewBuiltinRegistry()
	clash := FakeBackend()
	clash.EngineNames = []string{"xelatex"}
	assert.ErrorIs(t, registry.Register(clash), backend.ErrInvalidBackend)

	clash = FakeBackend()
	clash.TemplateExtensions = []string{".TEX"}
	assert.ErrorIs(t, registry.Register(clash), backend.ErrInvalidBackend)

	noFactory := FakeBackend()
	noFactory.NewCompiler = nil
	assert.ErrorIs(t, registry.Register(noFactory), backend.ErrInvalidBackend)

	require.NoError(t, registry.Register(FakeBackend()))
	assert.ErrorIs(t, registry.Register(FakeBackend()), backend.ErrInvalidBackend)
}

func TestLaTeXCompilerFactorySelectsByOptions(t *testing.T) {
	b := backend.LaTeX()
	plain, err := b.Compiler(backend.CompilerDeps{WorkingDir: "/w"}, ports.NewLaTeXOptions(2, false))
	require.NoError(t, err)
	assert.IsType(t, &latex.LaTeXCompilerAdapter{}, plain)

	mk, err := b.Compiler(backend.CompilerDeps{}, ports.NewLaTeXOptions(1, true))
	require.NoError(t, err)
	assert.IsType(t, &latex.LatexmkCompilerAdapter{}, mk)

	_, err = b.Compiler(backend.CompilerDeps{}, fakeOptions{})
	assert.ErrorIs(t, err, ports.ErrBackendOptionsMismatch)

	opts, err := b.Options(backend.Settings{Passes: 3, UseLatexmk: true})
	require.NoError(t, err)
	assert.Equal(t, ports.LaTeXOptions{Passes: 3, UseLatexmk: true}, opts)
}

func TestCompilerSourceResolvesPerRequest(t *testing.T) {
	registry := backend.NewBuiltinRegistry().MustRegister(FakeBackend())
	source := backend.CompilerSource{Resolver: registry}

	compiler, err := source.CompilerFor("", "doc.fake", nil)
	require.NoError(t, err)
	assert.Equal(t, stubCompiler{name: "fake"}, compiler)

	_, err = source.CompilerFor("nope", "doc.tex", nil)
	assert.ErrorIs(t, err, backend.ErrUnknownEngine)
}
