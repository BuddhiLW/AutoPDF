// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package escape_test

import (
	"strings"
	"testing"
	"text/template"

	"github.com/BuddhiLW/AutoPDF/v2/pkg/template/escape"
)

func TestTypstEscapesEveryMarkupSpecial(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"plain text is untouched", "Hello, world 42.", "Hello, world 42."},
		{"unicode is untouched", "Ação ñ 日本", "Ação ñ 日本"},
		{"hash starts code", "#let x = 1", `\#let x \= 1`},
		{"dollar starts math", "$x^2$", `\$x^2\$`},
		{"strong", "*bold*", `\*bold\*`},
		{"emphasis", "_it_", `\_it\_`},
		{"raw", "`code`", "\\`code\\`"},
		{"label", "<intro>", `\<intro\>`},
		{"reference", "@intro", `\@intro`},
		{"content block", "#f[a[b]]", `\#f\[a\[b\]\]`},
		{"backslash", `a\b`, `a\\b`},
		{"non-breaking space", "a~b", `a\~b`},
		{"heading", "= Title", `\= Title`},
		{"list markers", "- a + b", `\- a \+ b`},
		{"comment", "a // b /* c */", `a \/\/ b \/\* c \*\/`},
		{"already escaped is escaped again", `\#`, `\\\#`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := escape.Typst(tc.input); got != tc.want {
				t.Fatalf("Typst(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestFuncsInstallsTheTypstHelperUnderItsName(t *testing.T) {
	tmpl, err := template.New("t.typ").
		Funcs(escape.Funcs()).
		Delims("delim[[", "]]").
		Parse(`delim[[ escapeTypst .v ]]`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var out strings.Builder
	if err := tmpl.Execute(&out, map[string]string{"v": "#x"}); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if out.String() != `\#x` {
		t.Fatalf("output = %q", out.String())
	}
	if escape.TypstFuncName != "escapeTypst" {
		t.Fatalf("func name = %q", escape.TypstFuncName)
	}
}
