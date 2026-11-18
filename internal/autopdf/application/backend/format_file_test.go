// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package backend_test

import (
	"errors"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/backend"
	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
)

func TestTypstRefusesFormatFile(t *testing.T) {
	_, err := backend.Typst().Options(backend.Settings{Passes: 1, FormatFile: true})
	var unsupported *backend.UnsupportedSettingError
	if !errors.As(err, &unsupported) || unsupported.Setting != "format_file" {
		t.Fatalf("want an UnsupportedSettingError for format_file, got %v", err)
	}
	if !errors.Is(err, backend.ErrUnsupportedSetting) {
		t.Fatalf("errors.Is(err, ErrUnsupportedSetting) must hold, got %v", err)
	}
}

func TestLaTeXCarriesFormatFileIntoItsOptions(t *testing.T) {
	opts, err := backend.LaTeX().Options(backend.Settings{Passes: 2, FormatFile: true})
	if err != nil {
		t.Fatalf("options: %v", err)
	}
	if want := (ports.LaTeXOptions{Passes: 2, FormatFile: true}); opts != want {
		t.Fatalf("options = %+v, want %+v", opts, want)
	}
	opts, err = backend.LaTeX().Options(backend.Settings{Passes: 1})
	if err != nil || opts.(ports.LaTeXOptions).FormatFile {
		t.Fatalf("format_file must default to off, got %+v (%v)", opts, err)
	}
}

func TestLaTeXRefusesFormatFileWithLatexmk(t *testing.T) {
	_, err := backend.LaTeX().Options(backend.Settings{Passes: 1, UseLatexmk: true, FormatFile: true})
	if !errors.Is(err, backend.ErrUnsupportedSetting) {
		t.Fatalf("format_file with use_latexmk must be refused, got %v", err)
	}
}
