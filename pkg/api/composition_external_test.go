package api_test

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/document"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/preview"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/latex"
)

func documentSpec() document.DocumentSpec {
	return document.DocumentSpec{
		SchemaVersion: document.CurrentSchemaVersion,
		ID:            "example",
		Blocks: []document.Component{{
			ID: "chapter", Kind: "section", Variant: "default",
			Props:    document.Props{"content": `\section{Chapter}`},
			Children: []document.Component{{ID: "body", Kind: "text", Variant: "default", Props: document.Props{"content": "Hello"}}},
		}},
	}
}

func TestDocumentEnginePublicCompositionAndGeneration(t *testing.T) {
	engine, err := api.NewDefaultDocumentEngine("latex", api.DocumentEngineConfig{MaxWorkers: 2, TrustedPreamble: []byte(`\usepackage{xcolor}`)})
	if err != nil {
		t.Fatal(err)
	}
	if got := len(engine.Catalog()); got != 3 {
		t.Fatalf("catalog has %d components", got)
	}
	first, err := engine.Project(context.Background(), documentSpec(), api.ProjectionOptions{FocusSections: []string{"chapter"}})
	if err != nil {
		t.Fatal(err)
	}
	second, err := engine.Project(context.Background(), documentSpec(), api.ProjectionOptions{FocusSections: []string{"chapter"}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) || !strings.Contains(string(findFile(t, first, "main.tex")), `\includeonly{`) {
		t.Fatal("public projection is not deterministic or focusable")
	}
	var generatedHash string
	result, err := engine.Generate(context.Background(), documentSpec(), api.ProjectionOptions{}, api.ProjectionGeneratorFunc(func(_ context.Context, projection latex.Projection) (api.Result, error) {
		generatedHash = projection.Hash
		return api.Result{PDF: []byte(projection.Hash)}, nil
	}))
	if err != nil || generatedHash == "" || string(result.PDF) != generatedHash {
		t.Fatalf("projection generator boundary failed: %q %v", result.PDF, err)
	}
}

func TestDocumentEnginePreviewUsesSameProjection(t *testing.T) {
	engine, err := api.NewDefaultDocumentEngine("latex", api.DocumentEngineConfig{})
	if err != nil {
		t.Fatal(err)
	}
	session, err := engine.NewPreviewSession(compilerFunc(func(_ context.Context, request preview.CompileRequest) (preview.CompileOutput, error) {
		return preview.CompileOutput{Pages: []preview.Page{{Number: 1, Data: []byte(request.Main)}}}, nil
	}), preview.Options{WorkspaceRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	result := <-engine.Preview(context.Background(), session, documentSpec(), api.ProjectionOptions{})
	if result.Err != nil || result.Projection == "" || len(result.ChangedPages) != 1 {
		t.Fatalf("preview failed: %+v", result)
	}
}

func ExampleDocumentEngine_Project() {
	engine, _ := api.NewDefaultDocumentEngine("latex", api.DocumentEngineConfig{MaxWorkers: 2})
	projection, _ := engine.Project(context.Background(), documentSpec(), api.ProjectionOptions{FocusSections: []string{"chapter"}})
	fmt.Println(len(engine.Catalog()), len(projection.Files), strings.Contains(string(findProjectionFile(projection, "main.tex")), `\includeonly{`))
	// Output: 3 3 true
}

type compilerFunc func(context.Context, preview.CompileRequest) (preview.CompileOutput, error)

func (function compilerFunc) Compile(ctx context.Context, request preview.CompileRequest) (preview.CompileOutput, error) {
	return function(ctx, request)
}

func findFile(t *testing.T, projection latex.Projection, path string) []byte {
	t.Helper()
	for _, file := range projection.Files {
		if file.Path == path {
			return file.Content
		}
	}
	t.Fatalf("missing file %s", path)
	return nil
}

func findProjectionFile(projection latex.Projection, path string) []byte {
	for _, file := range projection.Files {
		if file.Path == path {
			return file.Content
		}
	}
	return nil
}
