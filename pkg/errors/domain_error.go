// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"errors"
	"strings"
)

// Common error categories
var (
	ErrValidation   = errors.New("validation error")
	ErrNotFound     = errors.New("not found")
	ErrInternal     = errors.New("internal error")
	ErrInvalidInput = errors.New("invalid input")
)

// DomainError provides structured error information with Elm-style guidance
type DomainError struct {
	Category    error
	Code        string
	Message     string
	Blame       string
	Suggestions []string
	Details     map[string]interface{}
	Cause       error
}

func (e *DomainError) Error() string {
	// Compact, non-colored single-line for non-debug surfaces
	var b strings.Builder
	if e.Code != "" {
		b.WriteString("[")
		b.WriteString(e.Code)
		b.WriteString("] ")
	}
	b.WriteString(e.Message)
	if e.Blame != "" {
		b.WriteString(" | Blame: ")
		b.WriteString(e.Blame)
	}
	if e.Cause != nil {
		b.WriteString(" | Cause: ")
		b.WriteString(e.Cause.Error())
	}
	return b.String()
}

// PrettyError renders the error with ANSI colors and readable dumps
func (e *DomainError) PrettyError(useColors bool, dumper DataDumper) string {
	pp := NewPrettyPrinter(useColors, dumper)
	return pp.FormatError(e)
}

func (e *DomainError) Unwrap() error { return e.Cause }
