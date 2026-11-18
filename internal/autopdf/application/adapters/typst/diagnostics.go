// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package typst

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Severity is the level of a Typst diagnostic.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Diagnostic is one entry of `typst compile --diagnostic-format short` output.
// Path, Line and Column are zero when typst reported no location.
type Diagnostic struct {
	Path     string
	Line     int
	Column   int
	Severity Severity
	Message  string
	Hints    []string
}

// String renders the diagnostic in typst's short format.
func (d Diagnostic) String() string {
	if d.Path == "" {
		return fmt.Sprintf("%s: %s", d.Severity, d.Message)
	}
	return fmt.Sprintf("%s:%d:%d: %s: %s", d.Path, d.Line, d.Column, d.Severity, d.Message)
}

var (
	locatedDiagnostic   = regexp.MustCompile(`^(.*):(\d+):(\d+): (error|warning): (.*)$`)
	unlocatedDiagnostic = regexp.MustCompile(`^(error|warning): (.*)$`)
	hintLine            = regexp.MustCompile(`^\s*(?:=\s*)?hint: (.*)$`)
)

// ParseDiagnostics extracts the diagnostics from typst short-format output
// (`path:line:col: severity: message`, `severity: message` when unlocated).
// A `hint: ...` line attaches to the diagnostic before it; other lines are
// ignored.
func ParseDiagnostics(output string) []Diagnostic {
	var diags []Diagnostic
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimRight(raw, "\r")
		if m := locatedDiagnostic.FindStringSubmatch(line); m != nil {
			lineNo, _ := strconv.Atoi(m[2])
			col, _ := strconv.Atoi(m[3])
			diags = append(diags, Diagnostic{
				Path:     m[1],
				Line:     lineNo,
				Column:   col,
				Severity: Severity(m[4]),
				Message:  m[5],
			})
			continue
		}
		if m := unlocatedDiagnostic.FindStringSubmatch(line); m != nil {
			diags = append(diags, Diagnostic{Severity: Severity(m[1]), Message: m[2]})
			continue
		}
		if m := hintLine.FindStringSubmatch(line); m != nil && len(diags) > 0 {
			last := &diags[len(diags)-1]
			last.Hints = append(last.Hints, m[1])
		}
	}
	return diags
}

// ErrCompileFailed is matched (errors.Is) by every *CompileError.
var ErrCompileFailed = errors.New("typst compilation failed")

// CompileError reports a typst run that did not produce a PDF, with the
// diagnostics typst printed.
type CompileError struct {
	Command     string       // the command line that ran
	ExitCode    int          // typst's exit status
	Diagnostics []Diagnostic // parsed from stderr (and stdout)
	Output      string       // raw stderr, then stdout
	Err         error        // underlying execution error, may be nil
}

// Errors returns the diagnostics of SeverityError.
func (e *CompileError) Errors() []Diagnostic {
	var out []Diagnostic
	for _, d := range e.Diagnostics {
		if d.Severity == SeverityError {
			out = append(out, d)
		}
	}
	return out
}

func (e *CompileError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "typst compilation failed (exit status %d)", e.ExitCode)
	errs := e.Errors()
	switch {
	case len(errs) > 0:
		for _, d := range errs {
			b.WriteString("\n  ")
			b.WriteString(d.String())
		}
	case strings.TrimSpace(e.Output) != "":
		b.WriteString(":\n")
		b.WriteString(strings.TrimSpace(e.Output))
	case e.Err != nil:
		fmt.Fprintf(&b, ": %v", e.Err)
	}
	if e.Command != "" {
		fmt.Fprintf(&b, "\nCommand: %s", e.Command)
	}
	return b.String()
}

// Unwrap returns the underlying execution error.
func (e *CompileError) Unwrap() error { return e.Err }

// Is makes errors.Is(err, ErrCompileFailed) hold.
func (e *CompileError) Is(target error) bool { return target == ErrCompileFailed }
