package api

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/component"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/render/latex"
)

// ProjectionBackendLaTeX names the default projection backend.
const ProjectionBackendLaTeX = "latex"

// ErrUnknownProjectionBackend reports a backend name no ProjectionBackend
// answers to.
var ErrUnknownProjectionBackend = errors.New("api: unknown projection backend")

// ProjectionBackend is everything the document engine and the projection
// generator need to know about one source language: which projector maps a
// manifest into it, which builtin definitions render into it, and how its
// compiler reports source locations.
type ProjectionBackend struct {
	// Name is the backend key, e.g. "latex" or "typst".
	Name string
	// Engine is the engine name requested when the caller names none; empty
	// keeps the Engine's default.
	Engine string
	// MainExtension identifies a projection of this backend by its main file.
	MainExtension string
	// NewProjector builds the backend's projector for one engine config.
	NewProjector func(DocumentEngineConfig) ManifestProjector
	// Builtins returns the builtin component definitions for target.
	Builtins func(target string) ([]component.Definition, error)
	// Diagnostics extracts compiler messages: submatch 1 is the generated
	// file, 2 its line, 3 the message. Nil selects the LaTeX file:line form.
	Diagnostics *regexp.Regexp
	// RootRelativePaths marks compilers that report workspace files as
	// "/path" relative to the workspace root.
	RootRelativePaths bool
}

// ProjectionBackends selects a ProjectionBackend by name.
type ProjectionBackends map[string]ProjectionBackend

// DefaultProjectionBackends returns the backends pkg/api knows without naming
// a render target: LaTeX alone. Other targets register through Register, so
// adding one never edits this package. The returned map is fresh.
func DefaultProjectionBackends() ProjectionBackends {
	return ProjectionBackends{
		ProjectionBackendLaTeX: {
			Name:          ProjectionBackendLaTeX,
			MainExtension: ".tex",
			NewProjector: func(config DocumentEngineConfig) ManifestProjector {
				return latexProjector{trustedPreamble: config.TrustedPreamble}
			},
			Builtins:    component.Builtins,
			Diagnostics: projectionDiagnosticPattern,
		},
	}
}

// Register adds or replaces a backend under its Name.
func (backends ProjectionBackends) Register(backend ProjectionBackend) error {
	if strings.TrimSpace(backend.Name) == "" || backend.NewProjector == nil || backend.Builtins == nil {
		return fmt.Errorf("api: projection backend needs a name, projector and builtins")
	}
	backends[backend.Name] = backend
	return nil
}

// Select returns the backend registered under name.
func (backends ProjectionBackends) Select(name string) (ProjectionBackend, error) {
	backend, ok := backends[name]
	if !ok {
		names := make([]string, 0, len(backends))
		for known := range backends {
			names = append(names, known)
		}
		sort.Strings(names)
		return ProjectionBackend{}, fmt.Errorf("%w %q (known: %s)", ErrUnknownProjectionBackend, name, strings.Join(names, ", "))
	}
	return backend, nil
}

// ForProjection returns the backend whose main extension matches the
// projection's main file, defaulting to LaTeX.
func (backends ProjectionBackends) ForProjection(projection latex.Projection) ProjectionBackend {
	extension := strings.ToLower(path.Ext(projection.Main))
	names := make([]string, 0, len(backends))
	for name := range backends {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if backends[name].MainExtension == extension {
			return backends[name]
		}
	}
	return backends[ProjectionBackendLaTeX]
}

// backendName resolves the projection backend for an engine config: an
// explicit Backend wins, then a target that names a backend, then LaTeX.
func (backends ProjectionBackends) backendName(config DocumentEngineConfig) string {
	if config.Backend != "" {
		return config.Backend
	}
	if _, ok := backends[config.Target]; ok {
		return config.Target
	}
	return ProjectionBackendLaTeX
}

type rawDiagnostic struct {
	path    string
	line    int
	message string
}

func (backend ProjectionBackend) parseDiagnostics(text string) []rawDiagnostic {
	pattern := backend.Diagnostics
	if pattern == nil {
		pattern = projectionDiagnosticPattern
	}
	matches := pattern.FindAllStringSubmatch(text, -1)
	diagnostics := make([]rawDiagnostic, 0, len(matches))
	for _, match := range matches {
		line, _ := strconv.Atoi(match[2])
		diagnostics = append(diagnostics, rawDiagnostic{path: match[1], line: line, message: strings.TrimSpace(match[3])})
	}
	return diagnostics
}
