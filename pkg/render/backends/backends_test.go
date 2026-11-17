package backends_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/document"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/backends"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/latex"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/typst"
)

func typstSpec() document.DocumentSpec {
	return document.DocumentSpec{
		SchemaVersion: document.CurrentSchemaVersion,
		ID:            "report",
		Blocks: []document.Component{
			{ID: "lead", Kind: "text", Variant: "default", Props: document.Props{"content": "Cost: $5 #tag"}},
			{
				ID: "chapter", Kind: "section", Variant: "default", Props: document.Props{"title": "Chapter"},
				Children: []document.Component{{ID: "body", Kind: "text", Variant: "default", Props: document.Props{"content": "Hello *world*"}}},
			},
		},
	}
}

func TestDefaultSelectsLaTeXAndTypstByName(t *testing.T) {
	selected := backends.Default()
	for _, name := range []string{api.ProjectionBackendLaTeX, backends.Typst} {
		if backend, err := selected.Select(name); err != nil || backend.Name != name {
			t.Fatalf("backend %q: %+v %v", name, backend, err)
		}
	}
	if got := selected.ForProjection(latex.Projection{Main: "main.typ"}).Name; got != backends.Typst {
		t.Fatalf("main.typ selected %q", got)
	}
}

func newTypstEngine(t *testing.T) *api.DocumentEngine {
	t.Helper()
	engine, err := api.NewDefaultDocumentEngine(typst.Target, api.DocumentEngineConfig{MaxWorkers: 2, Backends: backends.Default()})
	if err != nil {
		t.Fatal(err)
	}
	return engine
}

func newGenerator(t *testing.T, generate api.GeneratorFunc) *api.LocalProjectionGenerator {
	t.Helper()
	core, err := api.NewEngine(api.WithGenerator(generate))
	if err != nil {
		t.Fatal(err)
	}
	generator, err := api.NewProjectionGenerator(api.ProjectionGeneratorOptions{Engine: core, WorkspaceRoot: t.TempDir(), Backends: backends.Default()})
	if err != nil {
		t.Fatal(err)
	}
	return generator
}

func TestTypstDocumentEngineProducesTypWorkspace(t *testing.T) {
	engine := newTypstEngine(t)
	if got := len(engine.Catalog()); got != 3 {
		t.Fatalf("typst catalog has %d components", got)
	}
	var (
		request   api.Request
		workspace map[string]string
	)
	generator := newGenerator(t, api.GeneratorFunc(func(_ context.Context, req api.Request) (api.Result, error) {
		request = req
		workspace = map[string]string{}
		walkErr := filepath.Walk(req.WorkingDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			data, err := os.ReadFile(path)
			relative, _ := filepath.Rel(req.WorkingDir, path)
			workspace[filepath.ToSlash(relative)] = string(data)
			return err
		})
		return api.Result{PDF: []byte("%PDF-fake")}, walkErr
	}))
	if _, err := engine.Generate(context.Background(), typstSpec(), api.ProjectionOptions{}, generator); err != nil {
		t.Fatal(err)
	}
	if filepath.Base(request.TemplatePath) != "main.typ" || request.LaTeXEngine != "typst" {
		t.Fatalf("typst projection not routed as typst: %+v", request)
	}
	main := workspace["main.typ"]
	if !strings.Contains(main, `#include "/fragments/lead-`) || !strings.Contains(main, "#pagebreak(weak: true)") {
		t.Fatalf("main.typ wrong:\n%s", main)
	}
	var sawEscaped, sawHeading bool
	for path, content := range workspace {
		if strings.HasSuffix(path, ".tex") {
			t.Fatalf("typst workspace holds LaTeX file %s", path)
		}
		sawEscaped = sawEscaped || strings.Contains(content, `Cost: \$5 \#tag`)
		sawHeading = sawHeading || strings.Contains(content, "= Chapter")
	}
	if !sawEscaped || !sawHeading || len(workspace) != 4 {
		t.Fatalf("workspace incomplete: %v", workspace)
	}
}

func TestTypstProjectionDiagnosticsMapToComponents(t *testing.T) {
	engine := newTypstEngine(t)
	projection, err := engine.Project(context.Background(), typstSpec(), api.ProjectionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	bodyPath := projection.SourceMap["body"].Path
	generator := newGenerator(t, api.GeneratorFunc(func(context.Context, api.Request) (api.Result, error) {
		return api.Result{}, errors.New("/" + bodyPath + ":1:7: error: unknown variable: foo")
	}))
	_, err = generator.GenerateProjection(context.Background(), projection)
	var failure *api.ProjectionGenerationError
	if !errors.As(err, &failure) || len(failure.Diagnostics) != 1 {
		t.Fatalf("expected one mapped diagnostic, got %v", err)
	}
	diagnostic := failure.Diagnostics[0]
	if diagnostic.ComponentID != "body" || diagnostic.Line != 1 || diagnostic.Message != "error: unknown variable: foo" {
		t.Fatalf("diagnostic not mapped: %+v", diagnostic)
	}
}

func TestTypstDocumentEngineWorkspaceCompilesWithRealTypst(t *testing.T) {
	binary, err := exec.LookPath("typst")
	if err != nil {
		t.Skip("typst not on PATH")
	}
	engine := newTypstEngine(t)
	generator := newGenerator(t, api.GeneratorFunc(func(ctx context.Context, req api.Request) (api.Result, error) {
		output := filepath.Join(req.WorkingDir, "out.pdf")
		command := exec.CommandContext(ctx, binary, "compile", "--root", req.WorkingDir, "--diagnostic-format", "short", req.TemplatePath, output)
		if combined, err := command.CombinedOutput(); err != nil {
			return api.Result{}, errors.New(string(combined))
		}
		pdf, err := os.ReadFile(output)
		return api.Result{PDF: pdf}, err
	}))
	result, err := engine.Generate(context.Background(), typstSpec(), api.ProjectionOptions{}, generator)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(result.PDF, []byte("%PDF-")) {
		t.Fatalf("not a PDF: %q", result.PDF)
	}
}
