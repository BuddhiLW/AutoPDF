// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package backend

import (
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
)

// CompilerSource resolves a backend through Resolver and builds its compiler
// over Deps. It implements document.CompilerResolver.
type CompilerSource struct {
	Resolver Resolver
	Deps     CompilerDeps
}

// CompilerFor returns the compiler of the backend chosen for engine and
// templatePath, built for opts.
func (s CompilerSource) CompilerFor(engine, templatePath string, opts ports.BackendOptions) (ports.DocumentCompiler, error) {
	resolver := s.Resolver
	if resolver == nil {
		resolver = Default()
	}
	b, err := resolver.Resolve(engine, templatePath)
	if err != nil {
		return nil, err
	}
	return b.Compiler(s.Deps, opts)
}
