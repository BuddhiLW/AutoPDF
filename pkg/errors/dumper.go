// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"encoding/json"
)

// DataDumper abstracts pretty-printing of arbitrary Go values
type DataDumper interface {
	Dump(v interface{}) string
}

// JSONDumper is a zero-dependency dumper using json.MarshalIndent
type JSONDumper struct{}

func (d JSONDumper) Dump(v interface{}) string {
	if v == nil {
		return "null"
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		// Fallback to string if not JSON-marshallable
		return "\"<unrenderable>\""
	}
	return string(b)
}

// Optional future adapter (example only, keep zero-deps by default)
// type SpewDumper struct{ cfg spew.ConfigState }
// func (d SpewDumper) Dump(v interface{}) string { return d.cfg.Sdump(v) }
