package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/BuddhiLW/AutoPDF/pkg/component"
	"github.com/BuddhiLW/AutoPDF/pkg/composition"
	"github.com/BuddhiLW/AutoPDF/pkg/document"
	"github.com/BuddhiLW/AutoPDF/pkg/preview"
	"github.com/BuddhiLW/AutoPDF/pkg/render/latex"
)

var ErrProjectionGeneratorRequired = errors.New("api: projection generator must not be nil")

// DocumentEngineConfig controls pure composition. The existing Engine and its
// Request/Result PDF boundary remain unchanged.
type DocumentEngineConfig struct {
	Target          string
	MaxWorkers      int
	Cache           composition.Cache
	Values          map[string]any
	TrustedPreamble []byte
}

// ProjectionOptions contains request-scoped projection choices.
type ProjectionOptions struct {
	FocusSections []string
}

// ProjectionGenerator is the final effect boundary for production PDF output.
// Implementations materialize the projection and invoke their chosen compiler.
type ProjectionGenerator interface {
	GenerateProjection(context.Context, latex.Projection) (Result, error)
}

// ProjectionGeneratorFunc adapts a function to ProjectionGenerator.
type ProjectionGeneratorFunc func(context.Context, latex.Projection) (Result, error)

func (function ProjectionGeneratorFunc) GenerateProjection(ctx context.Context, projection latex.Projection) (Result, error) {
	return function(ctx, projection)
}

// DocumentEngine is the embeddable component -> manifest -> projection facade.
type DocumentEngine struct {
	catalog         *component.Catalog
	composer        *composition.Composer
	trustedPreamble []byte
}

// NewDocumentEngine constructs a facade around an immutable component catalog.
func NewDocumentEngine(catalog *component.Catalog, config DocumentEngineConfig) (*DocumentEngine, error) {
	composer, err := composition.New(catalog, config.Target, composition.Options{
		MaxWorkers: config.MaxWorkers,
		Cache:      config.Cache,
		Values:     config.Values,
	})
	if err != nil {
		return nil, fmt.Errorf("api: create document engine: %w", err)
	}
	return &DocumentEngine{
		catalog: catalog, composer: composer,
		trustedPreamble: append([]byte(nil), config.TrustedPreamble...),
	}, nil
}

// NewDefaultDocumentEngine builds the minimal flow/section/artifact catalog.
func NewDefaultDocumentEngine(target string, config DocumentEngineConfig) (*DocumentEngine, error) {
	definitions, err := component.Builtins(target)
	if err != nil {
		return nil, fmt.Errorf("api: create built-in components: %w", err)
	}
	catalog, err := component.NewCatalog(definitions...)
	if err != nil {
		return nil, fmt.Errorf("api: create built-in catalog: %w", err)
	}
	config.Target = target
	return NewDocumentEngine(catalog, config)
}

// Catalog returns a defensive editor-discovery snapshot.
func (engine *DocumentEngine) Catalog() []component.Metadata {
	if engine == nil {
		return nil
	}
	return engine.catalog.Metadata()
}

// Validate reports structural problems without rendering.
func (engine *DocumentEngine) Validate(spec document.DocumentSpec) document.Problems {
	return spec.Problems()
}

// Compose renders a stable, content-addressed manifest.
func (engine *DocumentEngine) Compose(ctx context.Context, spec document.DocumentSpec) (composition.RenderManifest, error) {
	if engine == nil || engine.composer == nil {
		return composition.RenderManifest{}, composition.ErrCatalogRequired
	}
	return engine.composer.Compose(ctx, spec)
}

// Project composes and deterministically maps fragments to LaTeX source files.
func (engine *DocumentEngine) Project(ctx context.Context, spec document.DocumentSpec, options ProjectionOptions) (latex.Projection, error) {
	manifest, err := engine.Compose(ctx, spec)
	if err != nil {
		return latex.Projection{}, err
	}
	return latex.Project(manifest, latex.Options{
		TrustedPreamble: append([]byte(nil), engine.trustedPreamble...),
		FocusSections:   append([]string(nil), options.FocusSections...),
	})
}

// Generate reaches the production effect boundary while returning the same
// stable Result type used by Engine.Generate.
func (engine *DocumentEngine) Generate(ctx context.Context, spec document.DocumentSpec, options ProjectionOptions, generator ProjectionGenerator) (Result, error) {
	if generator == nil {
		return Result{}, ErrProjectionGeneratorRequired
	}
	projection, err := engine.Project(ctx, spec, options)
	if err != nil {
		return Result{}, err
	}
	return generator.GenerateProjection(ctx, projection)
}

// NewPreviewSession creates a revisioned session using the same projection
// values as production. The compiler is an explicit effect adapter.
func (engine *DocumentEngine) NewPreviewSession(compiler preview.Compiler, options preview.Options) (*preview.Session, error) {
	return preview.NewSession(compiler, options)
}

// Preview projects a document and submits it to a long-lived session.
func (engine *DocumentEngine) Preview(ctx context.Context, session *preview.Session, spec document.DocumentSpec, options ProjectionOptions) <-chan preview.Result {
	if session == nil {
		result := make(chan preview.Result, 1)
		result <- preview.Result{Err: preview.ErrClosed}
		close(result)
		return result
	}
	projection, err := engine.Project(ctx, spec, options)
	if err != nil {
		result := make(chan preview.Result, 1)
		result <- preview.Result{Err: err}
		close(result)
		return result
	}
	return session.Submit(ctx, projection)
}
