// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/backend"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/adapters"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/model"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/config"
)

var (
	ErrNilGenerator     = errors.New("api: generator must not be nil")
	ErrTemplateRequired = errors.New("api: template path is required")
)

// Request is the stable input contract for programmatic PDF generation.
type Request struct {
	TemplatePath string
	OutputPath   string
	Variables    map[string]string
	// Engine names the engine ("pdflatex", "xelatex", "lualatex", "typst", or
	// any engine a registered backend accepts). Empty selects the backend by
	// the template's extension. Engine wins over LaTeXEngine when both are set.
	Engine string
	// Deprecated: use Engine. LaTeXEngine is read only when Engine is empty.
	LaTeXEngine string
	WorkingDir  string
	Passes      int
	UseLatexmk  bool
	// FormatFile precompiles the LaTeX preamble into a cached format file
	// under the working directory; refused for Typst.
	FormatFile bool
	Debug      bool
	Conversion ConversionOptions
}

// ConversionOptions controls optional PDF-to-image conversion.
type ConversionOptions = model.ConversionOptions

// PDFMetadata describes a generated or inspected PDF.
type PDFMetadata = model.PDFMetadata

// Result is the stable output contract for programmatic PDF generation.
type Result struct {
	PDFPath    string
	PDF        []byte
	ImagePaths []string
	Metadata   PDFMetadata
}

// Generator is the extension point used by Engine to perform generation.
type Generator interface {
	Generate(context.Context, Request) (Result, error)
}

// GeneratorFunc adapts a function to Generator.
type GeneratorFunc func(context.Context, Request) (Result, error)

func (f GeneratorFunc) Generate(ctx context.Context, req Request) (Result, error) {
	return f(ctx, req)
}

// Capabilities describes the engines and conversion formats exposed by an Engine.
type Capabilities struct {
	Engines           []string
	ConversionFormats []string
}

// Option configures an Engine.
type Option func(*engineOptions) error

type engineOptions struct {
	generator    Generator
	logger       Logger
	capabilities Capabilities
}

// WithGenerator replaces the default generator with an application-supplied one.
func WithGenerator(generator Generator) Option {
	return func(options *engineOptions) error {
		if generator == nil {
			return ErrNilGenerator
		}
		options.generator = generator
		return nil
	}
}

// WithLogger routes engine lifecycle messages to logger. A nil logger is a no-op.
func WithLogger(logger Logger) Option {
	return func(options *engineOptions) error {
		if logger == nil {
			logger = NewNoopLogger()
		}
		options.logger = logger
		return nil
	}
}

// WithCapabilities advertises capabilities supplied by an embedded implementation.
func WithCapabilities(capabilities Capabilities) Option {
	return func(options *engineOptions) error {
		options.capabilities = cloneCapabilities(capabilities)
		return nil
	}
}

// WithBackendRegistry advertises the engines of registry instead of the
// default backend registry's.
func WithBackendRegistry(registry *backend.Registry) Option {
	return func(options *engineOptions) error {
		if registry == nil {
			return errors.New("api: backend registry must not be nil")
		}
		options.capabilities.Engines = registry.EngineNames()
		return nil
	}
}

// Engine is the canonical embeddable AutoPDF facade.
type Engine struct {
	generator    Generator
	logger       Logger
	capabilities Capabilities
}

// NewEngine constructs an Engine. The default generator uses AutoPDF's built-in pipeline.
func NewEngine(options ...Option) (*Engine, error) {
	settings := engineOptions{
		logger: NewNoopLogger(),
		capabilities: Capabilities{
			Engines:           backend.Default().EngineNames(),
			ConversionFormats: []string{"jpeg", "jpg", "png"},
		},
	}

	for index, option := range options {
		if option == nil {
			return nil, fmt.Errorf("api: option %d is nil", index)
		}
		if err := option(&settings); err != nil {
			return nil, fmt.Errorf("api: apply option %d: %w", index, err)
		}
	}

	if settings.generator == nil {
		settings.generator = &defaultGenerator{logger: settings.logger}
	}

	return &Engine{
		generator:    settings.generator,
		logger:       settings.logger,
		capabilities: cloneCapabilities(settings.capabilities),
	}, nil
}

// Generate produces a PDF while preserving the caller's context for injected generators.
func (engine *Engine) Generate(ctx context.Context, req Request) (Result, error) {
	if engine == nil || engine.generator == nil {
		return Result{}, ErrNilGenerator
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(req.TemplatePath) == "" {
		return Result{}, ErrTemplateRequired
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	engine.logger.Debug(ctx, "generating PDF",
		NewLogField("template_path", req.TemplatePath),
		NewLogField("output_path", req.OutputPath))
	result, err := engine.generator.Generate(ctx, req)
	if err != nil {
		return Result{}, err
	}
	return result, nil
}

// Capabilities returns a defensive copy of the engine's advertised capabilities.
func (engine *Engine) Capabilities() Capabilities {
	if engine == nil {
		return Capabilities{}
	}
	return cloneCapabilities(engine.capabilities)
}

type defaultGenerator struct {
	logger Logger
}

// EngineName is the engine the request selects: Engine, else the deprecated
// LaTeXEngine, else "" (the template's backend decides).
func (req Request) EngineName() string {
	if req.Engine != "" {
		return req.Engine
	}
	return req.LaTeXEngine
}

func (generator *defaultGenerator) Generate(ctx context.Context, req Request) (Result, error) {
	cfg := config.GetDefaultConfig()
	cfg.Template = config.Template(req.TemplatePath)
	cfg.Output = config.Output(req.OutputPath)
	if req.Passes > 0 {
		cfg.Passes = req.Passes
	}
	cfg.UseLatexmk = req.UseLatexmk
	cfg.FormatFile = req.FormatFile
	selected, err := backend.Default().Resolve(req.EngineName(), req.TemplatePath)
	if err != nil {
		return Result{}, fmt.Errorf("api: %w", err)
	}
	cfg.Engine = config.Engine(selected.EngineOrDefault(req.EngineName()))
	if _, err := selected.Options(backend.Settings{Passes: cfg.Passes, UseLatexmk: cfg.UseLatexmk, FormatFile: cfg.FormatFile}); err != nil {
		return Result{}, fmt.Errorf("api: %w", err)
	}
	cfg.Conversion = config.Conversion{
		Enabled: req.Conversion.Enabled,
		Formats: append([]string(nil), req.Conversion.Formats...),
	}
	for key, value := range req.Variables {
		if err := cfg.Variables.SetString(key, value); err != nil {
			return Result{}, fmt.Errorf("api: set variable %q: %w", key, err)
		}
	}

	adapter := adapters.NewInternalApplicationAdapterWithLogger(cfg, newInternalLoggerAdapter(generator.logger))
	pdf, images, err := adapter.GeneratePDFContext(ctx, cfg, cfg.Template, req.Debug, req.WorkingDir)
	if err != nil {
		return Result{}, err
	}

	imageKeys := make([]string, 0, len(images))
	for key := range images {
		imageKeys = append(imageKeys, key)
	}
	sort.Strings(imageKeys)
	imagePaths := make([]string, 0, len(imageKeys))
	for _, key := range imageKeys {
		imagePaths = append(imagePaths, images[key])
	}

	pdfPath := req.OutputPath
	if pdfPath != "" && !strings.EqualFold(filepath.Ext(pdfPath), ".pdf") {
		pdfPath += ".pdf"
	}
	return Result{PDFPath: pdfPath, PDF: pdf, ImagePaths: imagePaths}, nil
}

func cloneCapabilities(capabilities Capabilities) Capabilities {
	return Capabilities{
		Engines:           append([]string(nil), capabilities.Engines...),
		ConversionFormats: append([]string(nil), capabilities.ConversionFormats...),
	}
}
