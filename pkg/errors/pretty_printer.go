// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"fmt"
	"sort"
	"strings"
)

// ANSI color codes (no external deps)
const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorYellow  = "\033[33m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorGray    = "\033[90m"
	ColorBold    = "\033[1m"
)

// PrettyPrinter formats errors with ANSI colors and readable dumps
type PrettyPrinter struct {
	useColors bool
	dumper    DataDumper
}

func NewPrettyPrinter(useColors bool, dumper DataDumper) *PrettyPrinter {
	if dumper == nil {
		dumper = JSONDumper{}
	}
	return &PrettyPrinter{useColors: useColors, dumper: dumper}
}

func (pp *PrettyPrinter) colorize(color, text string) string {
	if !pp.useColors || text == "" {
		return text
	}
	return color + text + ColorReset
}

func (pp *PrettyPrinter) FormatError(err *DomainError) string {
	var b strings.Builder

	// Header: [CODE] Message
	header := err.Message
	if err.Code != "" {
		header = "[" + err.Code + "] " + err.Message
	}
	b.WriteString(pp.colorize(ColorBold+ColorRed, header))

	// Blame
	if err.Blame != "" {
		b.WriteString("\n\n  ")
		b.WriteString(pp.colorize(ColorYellow, "🔍 Blame: "))
		b.WriteString(err.Blame)
	}

	// Suggestions
	if len(err.Suggestions) > 0 {
		b.WriteString("\n\n  ")
		b.WriteString(pp.colorize(ColorCyan, "💡 Suggestions:"))
		for i, s := range err.Suggestions {
			b.WriteString("\n    ")
			b.WriteString(pp.colorize(ColorCyan, fmt.Sprintf("%d", i+1)))
			b.WriteString(") ")
			b.WriteString(s)
		}
	}

	// Details (stable order)
	if len(err.Details) > 0 {
		b.WriteString("\n\n  ")
		b.WriteString(pp.colorize(ColorGray, "📋 Details:"))
		keys := make([]string, 0, len(err.Details))
		for k := range err.Details {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		// Dump the whole details map as JSON for readability
		b.WriteString("\n    ")
		b.WriteString(pp.colorize(ColorGray, pp.dumper.Dump(err.Details)))
	}

	// Cause
	if err.Cause != nil {
		b.WriteString("\n\n  ")
		b.WriteString(pp.colorize(ColorMagenta, "⚠️  Caused by:"))
		b.WriteString("\n    ")
		// Prefer dump; fallback to Error()
		causeDump := pp.dumper.Dump(map[string]interface{}{"error": err.Cause.Error()})
		b.WriteString(pp.colorize(ColorMagenta, causeDump))
	}

	return b.String()
}
