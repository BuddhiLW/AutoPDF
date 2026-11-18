// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package application_test

import (
	"errors"
	"strings"
	"testing"

	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
)

type otherBackend struct{}

func (otherBackend) BackendName() string { return "typst" }

func TestLaTeXOptionsFrom(t *testing.T) {
	base := ports.NewCompileOptions("pdflatex", "out.pdf", ".")

	cases := []struct {
		name    string
		backend ports.BackendOptions
		want    ports.LaTeXOptions
	}{
		{"absent means defaults", nil, ports.LaTeXOptions{Passes: 1}},
		{"value", ports.LaTeXOptions{Passes: 3, UseLatexmk: true}, ports.LaTeXOptions{Passes: 3, UseLatexmk: true}},
		{"pointer", &ports.LaTeXOptions{Passes: 2}, ports.LaTeXOptions{Passes: 2}},
		{"nil pointer means defaults", (*ports.LaTeXOptions)(nil), ports.LaTeXOptions{Passes: 1}},
		{"passes below range clamp to 1", ports.LaTeXOptions{Passes: 0}, ports.LaTeXOptions{Passes: 1}},
		{"passes above range clamp to 10", ports.LaTeXOptions{Passes: 42}, ports.LaTeXOptions{Passes: 10}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ports.LaTeXOptionsFrom(base.WithBackendOptions(tc.backend))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestLaTeXOptionsFromRefusesAnotherBackendsOptions(t *testing.T) {
	_, err := ports.LaTeXOptionsFrom(ports.NewCompileOptions("typst", "", "").WithBackendOptions(otherBackend{}))
	var mismatch *ports.BackendOptionsMismatchError
	if !errors.As(err, &mismatch) {
		t.Fatalf("err = %v, want *BackendOptionsMismatchError", err)
	}
	if mismatch.Want != ports.LaTeXBackendName || mismatch.Got != "typst" {
		t.Fatalf("mismatch = %+v", mismatch)
	}
	if !errors.Is(err, ports.ErrBackendOptionsMismatch) {
		t.Fatal("errors.Is(err, ErrBackendOptionsMismatch) = false")
	}
}

func TestToolchainMissingErrorNamesTheBinary(t *testing.T) {
	cause := errors.New("executable file not found in $PATH")
	err := error(&ports.ToolchainMissingError{Binary: "typst", Err: cause})
	if !strings.Contains(err.Error(), "typst") {
		t.Fatalf("message %q does not name the binary", err.Error())
	}
	if !errors.Is(err, ports.ErrToolchainMissing) || !errors.Is(err, cause) {
		t.Fatal("ToolchainMissingError must match ErrToolchainMissing and unwrap to its cause")
	}
}

func TestLaTeXCompilerIsTheDocumentCompiler(t *testing.T) {
	var c ports.DocumentCompiler
	var legacy ports.LaTeXCompiler = c
	_ = legacy
}
