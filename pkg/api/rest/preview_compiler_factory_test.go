package rest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/backend"
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	rootapi "github.com/BuddhiLW/AutoPDF/v2/pkg/api"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/preview"
	"github.com/go-chi/chi/v5"
)

type namedCompiler string

func (namedCompiler) Compile(context.Context, preview.CompileRequest) (preview.CompileOutput, error) {
	return preview.CompileOutput{}, nil
}

func namedFactory(name string) PreviewCompilerFactory {
	return PreviewCompilerFactoryFunc(func() (preview.Compiler, error) { return namedCompiler(name), nil })
}

// testRegistry registers latex plus a typst descriptor built here, standing in
// for the registry entry the wiring step adds.
func testRegistry(t *testing.T) *backend.Registry {
	t.Helper()
	registry := backend.NewBuiltinRegistry()
	err := registry.Register(backend.Backend{
		Name: ports.TypstBackendName, EngineNames: []string{"typst"}, TemplateExtensions: []string{".typ"},
		NewCompiler: func(backend.CompilerDeps, ports.BackendOptions) (ports.DocumentCompiler, error) {
			return nil, errors.New("unused")
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return registry
}

func selectorUnderTest(t *testing.T) *BackendPreviewCompilerFactory {
	return &BackendPreviewCompilerFactory{
		Resolver: testRegistry(t),
		Compilers: map[string]PreviewCompilerFactory{
			ports.LaTeXBackendName: namedFactory("latex"),
			ports.TypstBackendName: namedFactory("typst"),
		},
	}
}

func TestBackendPreviewCompilerFactorySelectsByBackend(t *testing.T) {
	factory := selectorUnderTest(t)
	cases := []struct {
		request PreviewBackendRequest
		want    string
	}{
		{PreviewBackendRequest{}, "latex"},
		{PreviewBackendRequest{Backend: "typst"}, "typst"},
		{PreviewBackendRequest{Engine: "typst"}, "typst"},
		{PreviewBackendRequest{Engine: "xelatex"}, "latex"},
		{PreviewBackendRequest{Template: "doc/main.typ"}, "typst"},
		{PreviewBackendRequest{Template: "doc/main.tex"}, "latex"},
	}
	for _, tc := range cases {
		compiler, err := factory.NewCompilerFor(tc.request)
		if err != nil || compiler != namedCompiler(tc.want) {
			t.Fatalf("%+v: got %v %v, want %s", tc.request, compiler, err, tc.want)
		}
	}
	if _, err := factory.NewCompilerFor(PreviewBackendRequest{Engine: "nope"}); !errors.Is(err, ErrPreviewBackendUnresolved) || !errors.Is(err, backend.ErrUnknownEngine) {
		t.Fatalf("unknown engine must be refused by the registry, got %v", err)
	}
	if _, err := factory.NewCompilerFor(PreviewBackendRequest{Backend: "context"}); !errors.Is(err, ErrPreviewBackendUnsupported) {
		t.Fatalf("backend without preview compiler must be refused, got %v", err)
	}
	factory.Default = ports.TypstBackendName
	if compiler, err := factory.NewCompiler(); err != nil || compiler != namedCompiler("typst") {
		t.Fatalf("default backend ignored: %v %v", compiler, err)
	}
}

func TestBuiltinPreviewCompilerFactoryBuildsBothBackends(t *testing.T) {
	factory := NewBackendPreviewCompilerFactory(testRegistry(t))
	latexCompiler, err := factory.NewCompilerFor(PreviewBackendRequest{Engine: "pdflatex"})
	if _, ok := latexCompiler.(*preview.LaTeXCompiler); err != nil || !ok {
		t.Fatalf("latex compiler not built: %T %v", latexCompiler, err)
	}
	typstCompiler, err := factory.NewCompilerFor(PreviewBackendRequest{Template: "x.typ"})
	if _, ok := typstCompiler.(*preview.TypstCompiler); err != nil || !ok {
		t.Fatalf("typst compiler not built: %T %v", typstCompiler, err)
	}
}

func TestPreviewAPICreateSessionSelectsBackend(t *testing.T) {
	engine, err := rootapi.NewDefaultDocumentEngine("latex", rootapi.DocumentEngineConfig{})
	if err != nil {
		t.Fatal(err)
	}
	selected := make(chan string, 4)
	factory := selectorUnderTest(t)
	for name := range factory.Compilers {
		factory.Compilers[name] = PreviewCompilerFactoryFunc(func() (preview.Compiler, error) {
			selected <- name
			return namedCompiler(name), nil
		})
	}
	transport, err := NewPreviewAPI(PreviewAPIOptions{Engine: engine, CompilerFactory: factory, SessionOptions: preview.Options{WorkspaceRoot: t.TempDir()}})
	if err != nil {
		t.Fatal(err)
	}
	defer transport.Close()
	router := chi.NewRouter()
	router.Mount("/preview", transport.Routes())
	server := httptest.NewServer(router)
	defer server.Close()

	post := func(body string) int {
		response, err := http.Post(server.URL+"/preview/sessions", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		return response.StatusCode
	}
	if status := post(`{"engine":"typst"}`); status != http.StatusCreated || <-selected != "typst" {
		t.Fatalf("typst session status %d", status)
	}
	if status := post(``); status != http.StatusCreated || <-selected != "latex" {
		t.Fatalf("default session status %d", status)
	}
	if status := post(`{"engine":"nope"}`); status != http.StatusBadRequest {
		t.Fatalf("unknown engine status %d", status)
	}
	if status := post(`{"bogus":1}`); status != http.StatusBadRequest {
		t.Fatalf("unknown field status %d", status)
	}
}

func TestPreviewAPIPlainFactoryRefusesBackendRequest(t *testing.T) {
	engine, err := rootapi.NewDefaultDocumentEngine("latex", rootapi.DocumentEngineConfig{})
	if err != nil {
		t.Fatal(err)
	}
	transport, err := NewPreviewAPI(PreviewAPIOptions{Engine: engine, CompilerFactory: namedFactory("only"), SessionOptions: preview.Options{WorkspaceRoot: t.TempDir()}})
	if err != nil {
		t.Fatal(err)
	}
	defer transport.Close()
	if _, err := transport.newCompiler(PreviewBackendRequest{Backend: "typst"}); !errors.Is(err, ErrPreviewBackendSelection) {
		t.Fatalf("expected selection refusal, got %v", err)
	}
	if compiler, err := transport.newCompiler(PreviewBackendRequest{}); err != nil || compiler != namedCompiler("only") {
		t.Fatalf("plain factory must still serve default sessions: %v %v", compiler, err)
	}
}
