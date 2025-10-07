// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package application

import "context"

// TemplateProcessor processes templates with variables
// Pure transport types - no domain dependencies
type TemplateProcessor interface {
	Process(ctx context.Context, templatePath string, variables map[string]string) (string, error)
}

// LaTeXCompiler compiles LaTeX content to PDF
// Pure transport types - no domain dependencies
type LaTeXCompiler interface {
	Compile(ctx context.Context, content string, engine string, outputPath string) (string, error)
}

// Converter converts PDFs to images
// Pure transport types - no domain dependencies
type Converter interface {
	ConvertToImages(ctx context.Context, pdfPath string, formats []string) ([]string, error)
}

// Cleaner removes auxiliary files
// Pure transport types - no domain dependencies
type Cleaner interface {
	Clean(ctx context.Context, pdfPath string) error
}
