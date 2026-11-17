// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"maps"
	"slices"
)

// TypstBackendName is the BackendName of TypstOptions.
const TypstBackendName = "typst"

// TypstOptions are the Typst-only compile knobs.
type TypstOptions struct {
	// Root is the project root passed as --root; files outside it cannot be
	// read by the document. Empty means the compile's working directory.
	Root string
	// FontPaths are extra font directories, each passed as --font-path.
	FontPaths []string
	// Inputs are passed as --input key=value and read in the document through
	// sys.inputs.
	Inputs map[string]string
}

// BackendName implements BackendOptions.
func (TypstOptions) BackendName() string { return TypstBackendName }

// DefaultTypstOptions returns options with no root override, no extra font
// paths and no inputs.
func DefaultTypstOptions() TypstOptions { return TypstOptions{} }

// clone returns a copy that shares no slice or map with o.
func (o TypstOptions) clone() TypstOptions {
	o.FontPaths = slices.Clone(o.FontPaths)
	o.Inputs = maps.Clone(o.Inputs)
	return o
}

// TypstOptionsFrom reads the Typst options carried by opts. Nil backend
// options yield DefaultTypstOptions; options of another backend (LaTeXOptions
// included, so passes and latexmk can never reach a Typst compile) are refused
// with a *BackendOptionsMismatchError.
func TypstOptionsFrom(opts CompileOptions) (TypstOptions, error) {
	switch b := opts.Backend.(type) {
	case nil:
		return DefaultTypstOptions(), nil
	case TypstOptions:
		return b.clone(), nil
	case *TypstOptions:
		if b == nil {
			return DefaultTypstOptions(), nil
		}
		return b.clone(), nil
	default:
		return TypstOptions{}, &BackendOptionsMismatchError{Want: TypstBackendName, Got: b.BackendName()}
	}
}
