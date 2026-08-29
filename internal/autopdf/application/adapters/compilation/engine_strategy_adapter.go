// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

// Package compilation adapts the public generation engine to the domain's
// CompilationStrategy port, so parallel compilation dispatches through one
// abstraction rather than reaching for a concrete builder.
package compilation

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/domain/parallel"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api"
)

// DefaultTemplateExtensions are the suffixes a LaTeX strategy claims.
var DefaultTemplateExtensions = []string{".tex", ".latex"}

// EngineStrategy implements parallel.CompilationStrategy on top of an
// api.Generator. It depends on the Generator interface rather than on
// *api.Engine, so a caller can substitute a fake without a real LaTeX
// toolchain.
type EngineStrategy struct {
	generator  api.Generator
	extensions []string
	engineName string
	outputDir  string
}

// EngineStrategyOption configures an EngineStrategy.
type EngineStrategyOption func(*EngineStrategy)

// WithExtensions overrides which template suffixes the strategy claims.
func WithExtensions(extensions ...string) EngineStrategyOption {
	return func(strategy *EngineStrategy) {
		if len(extensions) > 0 {
			strategy.extensions = extensions
		}
	}
}

// WithLaTeXEngine selects the LaTeX engine passed to the generator.
func WithLaTeXEngine(name string) EngineStrategyOption {
	return func(strategy *EngineStrategy) { strategy.engineName = name }
}

// WithOutputDir places generated PDFs in dir rather than beside the template.
func WithOutputDir(dir string) EngineStrategyOption {
	return func(strategy *EngineStrategy) { strategy.outputDir = dir }
}

// NewEngineStrategy builds a compilation strategy over the given generator.
func NewEngineStrategy(generator api.Generator, options ...EngineStrategyOption) *EngineStrategy {
	strategy := &EngineStrategy{
		generator:  generator,
		extensions: DefaultTemplateExtensions,
	}
	for _, option := range options {
		option(strategy)
	}
	return strategy
}

// CanHandle reports whether this strategy claims the template. Pure.
func (strategy *EngineStrategy) CanHandle(template string) bool {
	if strategy.generator == nil {
		return false
	}
	suffix := strings.ToLower(filepath.Ext(template))
	for _, extension := range strategy.extensions {
		if suffix == strings.ToLower(extension) {
			return true
		}
	}
	return false
}

// outputPathFor derives where the PDF for a template should land. Pure.
func (strategy *EngineStrategy) outputPathFor(template string) string {
	base := strings.TrimSuffix(filepath.Base(template), filepath.Ext(template)) + ".pdf"
	if strategy.outputDir == "" {
		return filepath.Join(filepath.Dir(template), base)
	}
	return filepath.Join(strategy.outputDir, base)
}

// Compile generates one PDF. This is the boundary: everything effectful is in
// the generator call.
func (strategy *EngineStrategy) Compile(
	ctx context.Context,
	template string,
	config string,
) (*parallel.BuildResult, error) {
	startedAt := time.Now()

	result, err := strategy.generator.Generate(ctx, api.Request{
		TemplatePath: template,
		OutputPath:   strategy.outputPathFor(template),
		LaTeXEngine:  strategy.engineName,
		WorkingDir:   filepath.Dir(config),
	})
	if err != nil {
		return nil, err
	}

	return &parallel.BuildResult{
		TemplateFile: template,
		PDFPath:      result.PDFPath,
		Duration:     time.Since(startedAt),
		Timestamp:    time.Now(),
	}, nil
}
