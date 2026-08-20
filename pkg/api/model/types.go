// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

// Package model contains dependency-free value types shared by public API layers.
package model

import "time"

// ConversionOptions controls optional PDF-to-image conversion.
type ConversionOptions struct {
	Enabled bool
	Formats []string
}

// PDFMetadata describes a generated or inspected PDF.
type PDFMetadata struct {
	FileSize    int64
	PageCount   int
	GeneratedAt time.Time
	Engine      string
	Template    string
}
