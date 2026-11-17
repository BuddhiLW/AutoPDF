// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package architecture

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/cleaner"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/compilation"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/backend"
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/commands/common/args"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/commands/common/wiring"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/domain/options"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	apiapplication "github.com/BuddhiLW/AutoPDF/v2/pkg/api/application"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/domain"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/domain/generation"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/rest"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCompiler struct{}

func (fakeCompiler) Compile(context.Context, string, ports.CompileOptions) (string, error) {
	return "fake.pdf", nil
}

type fakeOptions struct{ passes int }

func (fakeOptions) BackendName() string { return "fake" }

func fakeBackend() backend.Backend {
	return backend.Backend{
		Name:               "fake",
		EngineNames:        []string{"fakeengine"},
		TemplateExtensions: []string{".fake"},
		StagedSuffix:       ".autopdf.fake",
		AuxExtensions:      []string{".fakeaux"},
		NewCompiler: func(backend.CompilerDeps, ports.BackendOptions) (ports.DocumentCompiler, error) {
			return fakeCompiler{}, nil
		},
		NewOptions: func(s backend.Settings) (ports.BackendOptions, error) {
			return fakeOptions{passes: s.Passes}, nil
		},
	}
}

// TestRegisteredBackendAppearsInEveryProjection registers a backend no
// production code names and shows each consumer that used to hold a
// hardcoded list now sees it through the registry.
func TestRegisteredBackendAppearsInEveryProjection(t *testing.T) {
	registry := backend.NewBuiltinRegistry().MustRegister(fakeBackend())

	t.Run("pkg/api NewEngine capabilities", func(t *testing.T) {
		engine, err := api.NewEngine(api.WithBackendRegistry(registry))
		require.NoError(t, err)
		assert.Contains(t, engine.Capabilities().Engines, "fakeengine")

		defaults, err := api.NewEngine()
		require.NoError(t, err)
		assert.Equal(t, backend.Default().EngineNames(), defaults.Capabilities().Engines)
	})

	t.Run("pkg/api ValidateEngine", func(t *testing.T) {
		details, err := api.NewValidationUtilsWithRegistry(registry).ValidateEngine("fakeengine")
		require.NoError(t, err)
		assert.Nil(t, details)

		details, err = api.NewValidationUtils().ValidateEngine("fakeengine")
		require.NoError(t, err)
		assert.NotNil(t, details, "the default registry does not know the fake engine")
	})

	t.Run("pkg/api/application validateEngine", func(t *testing.T) {
		request := generation.PDFGenerationRequest{TemplatePath: "doc.fake", Engine: "fakeengine"}

		err := apiapplication.NewRequestValidationGuard(nil, nil).WithRegistry(registry).
			Validate(context.Background(), request)
		assert.NotEqual(t, domain.ErrCodeEngineNotFound, generationErrorCode(t, err),
			"the engine passes; validation stops later, at the empty output path")

		err = apiapplication.NewRequestValidationGuard(nil, nil).Validate(context.Background(), request)
		assert.Equal(t, domain.ErrCodeEngineNotFound, generationErrorCode(t, err))
	})

	t.Run("pkg/api/rest GetSupportedEngines", func(t *testing.T) {
		restAPI := rest.NewPDFGenerationAPI(config.GetDefaultConfig()).WithRegistry(registry)
		response := httptest.NewRecorder()
		restAPI.GetSupportedEngines(response, httptest.NewRequest(http.MethodGet, "/api/v1/pdf/engines", nil))
		require.Equal(t, http.StatusOK, response.Code)

		var body rest.EnginesResponse
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
		assert.Equal(t, registry.EngineNames(), body.Engines)
		assert.Equal(t, registry.DefaultEngine(), body.Default)
	})

	t.Run("compilation template extensions", func(t *testing.T) {
		generator := api.GeneratorFunc(func(context.Context, api.Request) (api.Result, error) { return api.Result{}, nil })
		assert.True(t, compilation.NewEngineStrategy(generator, compilation.WithRegistry(registry)).CanHandle("doc.fake"))
		assert.False(t, compilation.NewEngineStrategy(generator).CanHandle("doc.fake"))
		assert.Equal(t, backend.Default().TemplateExtensions(), compilation.DefaultTemplateExtensions())
	})

	t.Run("clean auxiliary extensions", func(t *testing.T) {
		assert.Contains(t, cleaner.NewCleanerAdapterWithExtensions(registry.AuxExtensions()).AuxExtensions(), ".fakeaux")
		assert.Equal(t, backend.Default().AuxExtensions(), cleaner.NewCleanerAdapter().AuxExtensions())

		dir := t.TempDir()
		aux := filepath.Join(dir, "doc.fakeaux")
		require.NoError(t, os.WriteFile(aux, []byte("x"), 0o644))
		result, err := wiring.NewServiceBuilderWithRegistry(registry).BuildCleanerService(dir).Clean()
		require.NoError(t, err)
		assert.Equal(t, []string{aux}, result.CleanedFiles)
	})

	t.Run("ServiceBuilder resolves compiler and options", func(t *testing.T) {
		builder := wiring.NewServiceBuilderWithRegistry(registry)
		cfg := &config.Config{Template: "doc.fake", Passes: 3}

		svc := builder.BuildDocumentServiceWithWorkingDir(cfg, t.TempDir())
		require.NotNil(t, svc.Compilers)
		compiler, err := svc.Compilers.CompilerFor(cfg.Engine.String(), cfg.Template.String(), nil)
		require.NoError(t, err)
		assert.Equal(t, fakeCompiler{}, compiler)

		req, err := builder.BuildRequest(&args.BuildArgs{Options: options.NewBuildOptions()}, cfg)
		require.NoError(t, err)
		assert.Equal(t, fakeOptions{passes: 3}, req.Backend)
	})
}

func generationErrorCode(t *testing.T, err error) string {
	t.Helper()
	require.Error(t, err)
	var generationErr domain.PDFGenerationError
	require.True(t, errors.As(err, &generationErr), "unexpected error type %T: %v", err, err)
	return generationErr.Code
}
