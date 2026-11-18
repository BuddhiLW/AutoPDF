// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package application_test

import (
	"errors"
	"reflect"
	"testing"

	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
)

func TestTypstOptionsFrom(t *testing.T) {
	base := ports.NewCompileOptions("typst", "out.pdf", ".")
	full := ports.TypstOptions{
		Root:      "/proj",
		FontPaths: []string{"/fonts"},
		Inputs:    map[string]string{"k": "v"},
	}

	cases := []struct {
		name    string
		backend ports.BackendOptions
		want    ports.TypstOptions
	}{
		{"absent means defaults", nil, ports.TypstOptions{}},
		{"value", full, full},
		{"pointer", &full, full},
		{"nil pointer means defaults", (*ports.TypstOptions)(nil), ports.TypstOptions{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ports.TypstOptionsFrom(base.WithBackendOptions(tc.backend))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestTypstOptionsFromReturnsACopy(t *testing.T) {
	in := ports.TypstOptions{FontPaths: []string{"/a"}, Inputs: map[string]string{"k": "v"}}
	got, err := ports.TypstOptionsFrom(ports.CompileOptions{Backend: in})
	if err != nil {
		t.Fatal(err)
	}
	got.FontPaths[0] = "/mutated"
	got.Inputs["k"] = "mutated"
	if in.FontPaths[0] != "/a" || in.Inputs["k"] != "v" {
		t.Fatalf("caller's options were mutated: %+v", in)
	}
}

func TestTypstOptionsFromRefusesLaTeXOptions(t *testing.T) {
	for _, latexOpts := range []ports.BackendOptions{
		ports.NewLaTeXOptions(3, false),
		ports.NewLaTeXOptions(1, true),
		ports.DefaultLaTeXOptions(),
		&ports.LaTeXOptions{Passes: 2},
	} {
		_, err := ports.TypstOptionsFrom(ports.NewCompileOptions("typst", "", "").WithBackendOptions(latexOpts))
		var mismatch *ports.BackendOptionsMismatchError
		if !errors.As(err, &mismatch) {
			t.Fatalf("%+v: err = %v, want *BackendOptionsMismatchError", latexOpts, err)
		}
		if mismatch.Want != ports.TypstBackendName || mismatch.Got != ports.LaTeXBackendName {
			t.Fatalf("mismatch = %+v", mismatch)
		}
		if !errors.Is(err, ports.ErrBackendOptionsMismatch) {
			t.Fatal("errors.Is(err, ErrBackendOptionsMismatch) = false")
		}
	}
}

func TestLaTeXOptionsFromRefusesTypstOptions(t *testing.T) {
	_, err := ports.LaTeXOptionsFrom(ports.CompileOptions{Backend: ports.TypstOptions{}})
	if !errors.Is(err, ports.ErrBackendOptionsMismatch) {
		t.Fatalf("err = %v, want ErrBackendOptionsMismatch", err)
	}
}
