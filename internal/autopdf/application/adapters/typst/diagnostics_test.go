// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package typst_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/typst"
)

func TestParseDiagnostics(t *testing.T) {
	cases := []struct {
		name   string
		output string
		want   []typst.Diagnostic
	}{
		{"empty", "", nil},
		{"noise only", "compiling...\n\n", nil},
		{
			name:   "located error",
			output: "main.typ:12:3: error: expected expression\n",
			want: []typst.Diagnostic{
				{Path: "main.typ", Line: 12, Column: 3, Severity: typst.SeverityError, Message: "expected expression"},
			},
		},
		{
			name:   "absolute path, warning and error, CRLF",
			output: "/w/doc.autopdf.typ:1:1: warning: unknown font family: foo\r\n/w/doc.autopdf.typ:4:10: error: file not found (searched at /w/x.png)\r\n",
			want: []typst.Diagnostic{
				{Path: "/w/doc.autopdf.typ", Line: 1, Column: 1, Severity: typst.SeverityWarning, Message: "unknown font family: foo"},
				{Path: "/w/doc.autopdf.typ", Line: 4, Column: 10, Severity: typst.SeverityError, Message: "file not found (searched at /w/x.png)"},
			},
		},
		{
			name:   "message containing colons",
			output: "a.typ:2:7: error: cannot add string and integer: \"a\" + 1\n",
			want: []typst.Diagnostic{
				{Path: "a.typ", Line: 2, Column: 7, Severity: typst.SeverityError, Message: "cannot add string and integer: \"a\" + 1"},
			},
		},
		{
			name:   "path with colon (windows drive)",
			output: "C:\\proj\\a.typ:5:2: error: boom\n",
			want: []typst.Diagnostic{
				{Path: "C:\\proj\\a.typ", Line: 5, Column: 2, Severity: typst.SeverityError, Message: "boom"},
			},
		},
		{
			name:   "unlocated error with hints",
			output: "error: failed to load file\n  = hint: check the root\nhint: run with --root\n",
			want: []typst.Diagnostic{
				{Severity: typst.SeverityError, Message: "failed to load file", Hints: []string{"check the root", "run with --root"}},
			},
		},
		{
			name:   "hint before any diagnostic is dropped",
			output: "hint: stray\nx.typ:1:2: error: e\n",
			want: []typst.Diagnostic{
				{Path: "x.typ", Line: 1, Column: 2, Severity: typst.SeverityError, Message: "e"},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := typst.ParseDiagnostics(tc.output)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got  %+v\nwant %+v", got, tc.want)
			}
		})
	}
}

func TestDiagnosticStringRoundTrips(t *testing.T) {
	for _, line := range []string{
		"main.typ:12:3: error: expected expression",
		"warning: no fonts found",
	} {
		got := typst.ParseDiagnostics(line)
		if len(got) != 1 || got[0].String() != line {
			t.Errorf("%q round-tripped to %+v", line, got)
		}
	}
}

func TestCompileErrorMessage(t *testing.T) {
	cause := errors.New("exit status 1")
	withDiags := &typst.CompileError{
		Command:  "typst compile x",
		ExitCode: 1,
		Diagnostics: []typst.Diagnostic{
			{Path: "a.typ", Line: 1, Column: 2, Severity: typst.SeverityWarning, Message: "w"},
			{Path: "a.typ", Line: 3, Column: 4, Severity: typst.SeverityError, Message: "e"},
		},
		Err: cause,
	}
	msg := withDiags.Error()
	if !strings.Contains(msg, "a.typ:3:4: error: e") || strings.Contains(msg, ": warning: w") {
		t.Errorf("message %q must list errors only", msg)
	}
	if !errors.Is(withDiags, cause) || !errors.Is(withDiags, typst.ErrCompileFailed) {
		t.Error("CompileError must unwrap to its cause and match ErrCompileFailed")
	}

	raw := &typst.CompileError{ExitCode: 2, Output: "thread panicked\n"}
	if !strings.Contains(raw.Error(), "thread panicked") {
		t.Errorf("message %q must fall back to raw output", raw.Error())
	}
}
