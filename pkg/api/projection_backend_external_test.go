package api_test

import (
	"context"
	"errors"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/component"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/composition"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/latex"
)

type fixedProjector struct{ main string }

func (projector fixedProjector) ProjectManifest(composition.RenderManifest, []string) (latex.Projection, error) {
	return latex.Projection{Main: projector.main}, nil
}

func TestProjectionBackendsSelectByName(t *testing.T) {
	backends := api.DefaultProjectionBackends()
	latexBackend, err := backends.Select(api.ProjectionBackendLaTeX)
	if err != nil || latexBackend.NewProjector == nil || latexBackend.Builtins == nil {
		t.Fatalf("latex backend: %+v %v", latexBackend, err)
	}
	if _, err := backends.Select("troff"); !errors.Is(err, api.ErrUnknownProjectionBackend) {
		t.Fatalf("expected unknown backend, got %v", err)
	}
	if _, err := api.NewDefaultDocumentEngine("latex", api.DocumentEngineConfig{Backend: "troff"}); !errors.Is(err, api.ErrUnknownProjectionBackend) {
		t.Fatalf("engine accepted unknown backend: %v", err)
	}
	if err := backends.Register(api.ProjectionBackend{Name: "troff"}); err == nil {
		t.Fatal("incomplete backend registered")
	}
	err = backends.Register(api.ProjectionBackend{
		Name: "troff", MainExtension: ".ms",
		NewProjector: func(api.DocumentEngineConfig) api.ManifestProjector { return fixedProjector{main: "main.ms"} },
		Builtins:     component.Builtins,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := backends.ForProjection(latex.Projection{Main: "main.ms"}).Name; got != "troff" {
		t.Fatalf("main.ms selected %q", got)
	}
	if got := backends.ForProjection(latex.Projection{Main: "main.tex"}).Name; got != api.ProjectionBackendLaTeX {
		t.Fatalf("main.tex selected %q", got)
	}
	engine, err := api.NewDefaultDocumentEngine("troff", api.DocumentEngineConfig{Backends: backends})
	if err != nil {
		t.Fatal(err)
	}
	projection, err := engine.Project(context.Background(), documentSpec(), api.ProjectionOptions{})
	if err != nil || projection.Main != "main.ms" {
		t.Fatalf("target did not select its backend: %+v %v", projection, err)
	}
}
