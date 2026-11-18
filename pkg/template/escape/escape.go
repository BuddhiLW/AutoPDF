// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

// Package escape holds the template helpers that make variable values safe
// as document markup. Every template func map installs Funcs.
package escape

import (
	"strings"
	"text/template"
)

// TypstFuncName is the template builtin that escapes a value for Typst markup:
//
//	delim[[ escapeTypst .author ]]
const TypstFuncName = "escapeTypst"

// typstMarkupSpecials lists every character that starts markup, math, code,
// a label, a reference, a comment or a line-start construct in Typst markup.
const typstMarkupSpecials = "\\#$*_`<>@[]~=-+/"

var typstEscaper = newBackslashEscaper(typstMarkupSpecials)

func newBackslashEscaper(specials string) *strings.Replacer {
	pairs := make([]string, 0, 2*len(specials))
	for _, special := range specials {
		pairs = append(pairs, string(special), `\`+string(special))
	}
	return strings.NewReplacer(pairs...)
}

// Typst makes arbitrary text safe as Typst markup: each markup special is
// preceded by a backslash, so it renders as itself.
func Typst(value string) string { return typstEscaper.Replace(value) }

// Funcs returns the escape helpers every template func map installs.
func Funcs() template.FuncMap {
	return template.FuncMap{
		TypstFuncName: Typst,
	}
}
