// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

// Package backend is the single source of truth for the document backends
// AutoPDF can compile with: their engine names, template extensions, staged
// source suffix, auxiliary files and compiler factory. Every list of engines or
// extensions elsewhere in the module is a projection of a Registry.
package backend

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
)

// CompilerDeps are the ports a compiler factory builds a DocumentCompiler over.
type CompilerDeps struct {
	FileSystem ports.FileSystem
	Executor   ports.CommandExecutor
	Logger     ports.Logger // optional
	WorkingDir string       // used when CompileOptions.WorkingDir is empty
}

// CompilerFactory builds the backend's DocumentCompiler. opts are the backend
// options the compile will carry (nil means the backend's defaults); a factory
// may use them to pick a variant (e.g. latexmk).
type CompilerFactory func(deps CompilerDeps, opts ports.BackendOptions) (ports.DocumentCompiler, error)

// Settings are the backend-relevant values read from configuration.
type Settings struct {
	Passes     int
	UseLatexmk bool
	FormatFile bool
}

// OptionsFactory builds a backend's BackendOptions from configuration values.
// A backend refuses settings it cannot honour with an error.
type OptionsFactory func(Settings) (ports.BackendOptions, error)

// Backend describes one document backend.
type Backend struct {
	Name               string   // e.g. "latex"
	EngineNames        []string // engines it accepts; the first is its default engine
	TemplateExtensions []string // template suffixes it claims, with the dot, e.g. ".tex"
	StagedSuffix       string   // suffix of the rendered source staged before compiling
	AuxExtensions      []string // auxiliary file suffixes `clean` may remove
	NewCompiler        CompilerFactory
	NewOptions         OptionsFactory
}

// DefaultEngine is the backend's first engine name.
func (b Backend) DefaultEngine() string {
	if len(b.EngineNames) == 0 {
		return ""
	}
	return b.EngineNames[0]
}

// EngineOrDefault returns engine, or the backend's default engine when engine is empty.
func (b Backend) EngineOrDefault(engine string) string {
	if engine == "" {
		return b.DefaultEngine()
	}
	return engine
}

// Options builds the backend's options from settings; a backend without an
// OptionsFactory gets nil (its defaults).
func (b Backend) Options(settings Settings) (ports.BackendOptions, error) {
	if b.NewOptions == nil {
		return nil, nil
	}
	return b.NewOptions(settings)
}

// Compiler builds the backend's compiler for the given options.
func (b Backend) Compiler(deps CompilerDeps, opts ports.BackendOptions) (ports.DocumentCompiler, error) {
	if b.NewCompiler == nil {
		return nil, fmt.Errorf("backend %q has no compiler factory", b.Name)
	}
	return b.NewCompiler(deps, opts)
}

// ErrUnknownEngine is matched (errors.Is) by every *UnknownEngineError.
var ErrUnknownEngine = errors.New("unknown engine")

// UnknownEngineError reports an engine no registered backend accepts.
type UnknownEngineError struct {
	Engine string
	Known  []string
}

func (e *UnknownEngineError) Error() string {
	return fmt.Sprintf("unknown engine %q; known engines: %s", e.Engine, strings.Join(e.Known, ", "))
}

// Is makes errors.Is(err, ErrUnknownEngine) hold.
func (e *UnknownEngineError) Is(target error) bool { return target == ErrUnknownEngine }

// ErrInvalidBackend is matched by every registration refusal.
var ErrInvalidBackend = errors.New("invalid backend")

// Resolver picks the backend for an engine name and template path.
type Resolver interface {
	Resolve(engine, templatePath string) (Backend, error)
}

// Registry holds the registered backends in registration order. The first
// registered backend is the default. It is safe for concurrent use.
type Registry struct {
	mu       sync.RWMutex
	backends []Backend
}

var _ Resolver = (*Registry)(nil)

// NewRegistry returns an empty registry.
func NewRegistry() *Registry { return &Registry{} }

// Register adds a backend. It refuses an empty name, a missing compiler
// factory, and a name, engine name or template extension already claimed.
func (r *Registry) Register(b Backend) error {
	if b.Name == "" {
		return fmt.Errorf("%w: empty name", ErrInvalidBackend)
	}
	if b.NewCompiler == nil {
		return fmt.Errorf("%w: %q has no compiler factory", ErrInvalidBackend, b.Name)
	}
	if len(b.EngineNames) == 0 {
		return fmt.Errorf("%w: %q has no engine names", ErrInvalidBackend, b.Name)
	}
	b = cloneBackend(b)
	for i, ext := range b.TemplateExtensions {
		b.TemplateExtensions[i] = strings.ToLower(ext)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.backends {
		if existing.Name == b.Name {
			return fmt.Errorf("%w: backend %q already registered", ErrInvalidBackend, b.Name)
		}
		for _, engine := range b.EngineNames {
			if slices.Contains(existing.EngineNames, engine) {
				return fmt.Errorf("%w: engine %q already claimed by %q", ErrInvalidBackend, engine, existing.Name)
			}
		}
		for _, ext := range b.TemplateExtensions {
			if slices.Contains(existing.TemplateExtensions, ext) {
				return fmt.Errorf("%w: extension %q already claimed by %q", ErrInvalidBackend, ext, existing.Name)
			}
		}
	}
	r.backends = append(r.backends, b)
	return nil
}

// MustRegister is Register that panics on refusal; for static wiring.
func (r *Registry) MustRegister(b Backend) *Registry {
	if err := r.Register(b); err != nil {
		panic(err)
	}
	return r
}

// Resolve picks a backend: a non-empty engine name wins (unknown engines are
// refused with *UnknownEngineError), else the template extension, else the
// default backend. An engine given as an executable path is matched by its
// base name and, when that is no known engine, falls through to the extension.
func (r *Registry) Resolve(engine, templatePath string) (Backend, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.backends) == 0 {
		return Backend{}, fmt.Errorf("%w: no backends registered", ErrInvalidBackend)
	}
	if engine != "" {
		name := engine
		isPath := strings.ContainsRune(engine, '/') || strings.ContainsRune(engine, filepath.Separator)
		if isPath {
			name = filepath.Base(engine)
		}
		for _, b := range r.backends {
			if slices.Contains(b.EngineNames, name) {
				return cloneBackend(b), nil
			}
		}
		if !isPath {
			return Backend{}, &UnknownEngineError{Engine: engine, Known: r.engineNamesLocked()}
		}
		// An explicit executable path names a binary, not an engine: the
		// template extension or the default backend decides.
	}
	if ext := strings.ToLower(filepath.Ext(templatePath)); ext != "" {
		for _, b := range r.backends {
			if slices.Contains(b.TemplateExtensions, ext) {
				return cloneBackend(b), nil
			}
		}
	}
	return cloneBackend(r.backends[0]), nil
}

// Lookup returns the backend registered under name.
func (r *Registry) Lookup(name string) (Backend, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, b := range r.backends {
		if b.Name == name {
			return cloneBackend(b), true
		}
	}
	return Backend{}, false
}

// Backends returns copies of the registered backends in registration order.
func (r *Registry) Backends() []Backend {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Backend, len(r.backends))
	for i, b := range r.backends {
		out[i] = cloneBackend(b)
	}
	return out
}

// EngineNames projects every registered engine name.
func (r *Registry) EngineNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.engineNamesLocked()
}

func (r *Registry) engineNamesLocked() []string {
	return project(r.backends, func(b Backend) []string { return b.EngineNames })
}

// TemplateExtensions projects every claimed template extension.
func (r *Registry) TemplateExtensions() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return project(r.backends, func(b Backend) []string { return b.TemplateExtensions })
}

// AuxExtensions projects every backend's auxiliary file extensions.
func (r *Registry) AuxExtensions() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return project(r.backends, func(b Backend) []string { return b.AuxExtensions })
}

// DefaultEngine is the default backend's default engine ("" when empty).
func (r *Registry) DefaultEngine() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.backends) == 0 {
		return ""
	}
	return r.backends[0].DefaultEngine()
}

func project(backends []Backend, field func(Backend) []string) []string {
	var out []string
	for _, b := range backends {
		for _, value := range field(b) {
			if !slices.Contains(out, value) {
				out = append(out, value)
			}
		}
	}
	return out
}

func cloneBackend(b Backend) Backend {
	b.EngineNames = slices.Clone(b.EngineNames)
	b.TemplateExtensions = slices.Clone(b.TemplateExtensions)
	b.AuxExtensions = slices.Clone(b.AuxExtensions)
	return b
}

var (
	defaultOnce     sync.Once
	defaultRegistry *Registry
)

// Default returns the process-wide registry holding the built-in backends.
func Default() *Registry {
	defaultOnce.Do(func() { defaultRegistry = NewBuiltinRegistry() })
	return defaultRegistry
}

// NewBuiltinRegistry returns a fresh registry with the built-in backends; tests
// extend it without touching Default().
func NewBuiltinRegistry() *Registry {
	return NewRegistry().MustRegister(LaTeX()).MustRegister(Typst())
}
