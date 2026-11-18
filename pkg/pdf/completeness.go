// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

// Package pdf holds the rules about PDF bytes. It is the lowest stratum: it
// imports only the standard library, so the engine boundary and every
// embedding application can ask the same question and get the same answer.
package pdf

import (
	"bytes"
	"fmt"
)

// Defect names why a byte sequence is not a finished PDF. The set is closed.
type Defect string

const (
	// DefectNone means the bytes are a finished PDF.
	DefectNone Defect = ""
	// DefectEmpty means there are no bytes at all.
	DefectEmpty Defect = "empty"
	// DefectMissingHeader means the bytes do not start with %PDF-.
	DefectMissingHeader Defect = "missing %PDF- header"
	// DefectMissingTrailer means the writer stopped before %%EOF: a compiler
	// that died, or a file read while it was still being written.
	DefectMissingTrailer Defect = "missing %%EOF trailer"
)

// TrailerWindow is how far from the end %%EOF may sit. The PDF specification
// allows at most 1024 bytes after it.
const TrailerWindow = 1024

var (
	header  = []byte("%PDF-")
	trailer = []byte("%%EOF")
)

// Inspect returns the first defect of content, or DefectNone. Pure.
func Inspect(content []byte) Defect {
	if len(content) == 0 {
		return DefectEmpty
	}
	if !bytes.HasPrefix(content, header) {
		return DefectMissingHeader
	}
	tail := content
	if len(tail) > TrailerWindow {
		tail = tail[len(tail)-TrailerWindow:]
	}
	if !bytes.Contains(tail, trailer) {
		return DefectMissingTrailer
	}
	return DefectNone
}

// IncompleteError refuses a PDF that has a defect. Callers match it with
// errors.As to tell an unfinished document from other failures.
type IncompleteError struct {
	Path   string
	Size   int
	Defect Defect
}

func (e *IncompleteError) Error() string {
	return fmt.Sprintf("generated PDF %s is incomplete (%d bytes): %s", e.Path, e.Size, e.Defect)
}

// Check returns an *IncompleteError when content has a defect, else nil. Pure.
func Check(path string, content []byte) error {
	if defect := Inspect(content); defect != DefectNone {
		return &IncompleteError{Path: path, Size: len(content), Defect: defect}
	}
	return nil
}
