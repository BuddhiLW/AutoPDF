package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/backend"
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/preview"
)

// PreviewBackendRequest names the backend a preview session compiles with:
// Backend directly, or an Engine / Template path resolved through the backend
// registry. All empty means the factory's default backend.
type PreviewBackendRequest struct {
	Backend  string `json:"backend,omitempty"`
	Engine   string `json:"engine,omitempty"`
	Template string `json:"template,omitempty"`
}

func (request PreviewBackendRequest) empty() bool {
	return request.Backend == "" && request.Engine == "" && request.Template == ""
}

// PreviewCompilerSelector is a PreviewCompilerFactory that can build a
// compiler for a requested backend.
type PreviewCompilerSelector interface {
	PreviewCompilerFactory
	NewCompilerFor(PreviewBackendRequest) (preview.Compiler, error)
}

var (
	ErrPreviewBackendUnsupported = errors.New("rest: no preview compiler for backend")
	ErrPreviewBackendUnresolved  = errors.New("rest: preview backend cannot be resolved")
	ErrPreviewBackendSelection   = errors.New("rest: preview compiler factory cannot select a backend")
)

// BackendPreviewCompilerFactory selects a preview compiler by backend name.
// Compilers is keyed by backend name ("latex", "typst"); Resolver maps an
// engine name or template extension to a backend; Default is the backend used
// when the request names none (empty: the resolver's default, else "latex").
type BackendPreviewCompilerFactory struct {
	Compilers map[string]PreviewCompilerFactory
	Resolver  backend.Resolver
	Default   string
}

var _ PreviewCompilerSelector = (*BackendPreviewCompilerFactory)(nil)

// NewBackendPreviewCompilerFactory wires the built-in preview compilers
// (LaTeX with Poppler pages, Typst with native pages) behind resolver.
func NewBackendPreviewCompilerFactory(resolver backend.Resolver) *BackendPreviewCompilerFactory {
	return &BackendPreviewCompilerFactory{
		Resolver: resolver,
		Compilers: map[string]PreviewCompilerFactory{
			ports.LaTeXBackendName: PreviewCompilerFactoryFunc(func() (preview.Compiler, error) {
				return preview.NewLaTeXCompiler(preview.LaTeXCompilerOptions{})
			}),
			ports.TypstBackendName: PreviewCompilerFactoryFunc(func() (preview.Compiler, error) {
				return preview.NewTypstCompiler(preview.TypstCompilerOptions{})
			}),
		},
	}
}

// NewCompiler builds the default backend's compiler.
func (factory *BackendPreviewCompilerFactory) NewCompiler() (preview.Compiler, error) {
	return factory.NewCompilerFor(PreviewBackendRequest{})
}

// NewCompilerFor builds the compiler of the backend the request resolves to.
func (factory *BackendPreviewCompilerFactory) NewCompilerFor(request PreviewBackendRequest) (preview.Compiler, error) {
	name, err := factory.backendName(request)
	if err != nil {
		return nil, err
	}
	compilerFactory, exists := factory.Compilers[name]
	if !exists || compilerFactory == nil {
		return nil, fmt.Errorf("%w %q", ErrPreviewBackendUnsupported, name)
	}
	return compilerFactory.NewCompiler()
}

func (factory *BackendPreviewCompilerFactory) backendName(request PreviewBackendRequest) (string, error) {
	if request.Backend != "" {
		return request.Backend, nil
	}
	if request.Engine == "" && request.Template == "" {
		if factory.Default != "" {
			return factory.Default, nil
		}
		if factory.Resolver == nil {
			return ports.LaTeXBackendName, nil
		}
	}
	if factory.Resolver == nil {
		return "", fmt.Errorf("%w: engine %q, template %q: no backend resolver", ErrPreviewBackendUnresolved, request.Engine, request.Template)
	}
	resolved, err := factory.Resolver.Resolve(request.Engine, request.Template)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrPreviewBackendUnresolved, err)
	}
	return resolved.Name, nil
}

// newCompiler builds the session compiler, honouring a backend request only
// when the factory can select one.
func (api *PreviewAPI) newCompiler(request PreviewBackendRequest) (preview.Compiler, error) {
	if selector, ok := api.compilerFactory.(PreviewCompilerSelector); ok {
		return selector.NewCompilerFor(request)
	}
	if !request.empty() {
		return nil, ErrPreviewBackendSelection
	}
	return api.compilerFactory.NewCompiler()
}

// sessionCompiler decodes the optional backend request of a session creation
// and builds its compiler, answering the HTTP status of a refusal.
func (api *PreviewAPI) sessionCompiler(request *http.Request) (preview.Compiler, int, error) {
	backendRequest, err := decodeBackendRequest(request)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	compiler, err := api.newCompiler(backendRequest)
	switch {
	case err == nil:
		return compiler, http.StatusCreated, nil
	case errors.Is(err, ErrPreviewBackendUnsupported), errors.Is(err, ErrPreviewBackendUnresolved), errors.Is(err, ErrPreviewBackendSelection):
		return nil, http.StatusBadRequest, err
	default:
		return nil, http.StatusInternalServerError, err
	}
}

func decodeBackendRequest(request *http.Request) (PreviewBackendRequest, error) {
	var body PreviewBackendRequest
	if request.Body == nil {
		return body, nil
	}
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil && !errors.Is(err, io.EOF) {
		return PreviewBackendRequest{}, err
	}
	return body, nil
}
