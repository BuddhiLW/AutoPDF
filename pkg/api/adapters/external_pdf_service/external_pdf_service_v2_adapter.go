// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package adapters

import (
	"context"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/decorators"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain/valueobjects"
	"github.com/BuddhiLW/AutoPDF/pkg/api"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
	"github.com/BuddhiLW/AutoPDF/pkg/config"
)

// ExternalPDFServiceV2Adapter implements domain.PDFGenerationService using V2 LaTeX compiler
type ExternalPDFServiceV2Adapter struct {
	config     *config.Config
	v2Compiler decorators.LaTeXCompiler
}

// NewExternalPDFServiceWithV2Compiler creates a new V2-based external PDF service adapter
func NewExternalPDFServiceWithV2Compiler(
	cfg *config.Config,
	compiler decorators.LaTeXCompiler,
) *ExternalPDFServiceV2Adapter {
	return &ExternalPDFServiceV2Adapter{
		config:     cfg,
		v2Compiler: compiler,
	}
}

// Generate generates a PDF using the V2 LaTeX compiler
func (epsa *ExternalPDFServiceV2Adapter) Generate(ctx context.Context, req generation.PDFGenerationRequest) (generation.PDFGenerationResult, error) {
	// Validate request first
	if err := epsa.ValidateRequest(req); err != nil {
		return generation.PDFGenerationResult{
			Success: false,
			Error:   err.(domain.PDFGenerationError),
		}, err
	}

	// Use processed template content - this should always be provided by the orchestration service
	if req.TemplateContent == "" {
		return generation.PDFGenerationResult{
			Success: false,
			Error: domain.PDFGenerationError{
				Code:    domain.ErrCodeTemplateInvalid,
				Message: "Template content is empty - template processing may have failed",
				Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh).
					WithTemplatePath(req.TemplatePath),
			},
		}, nil
	}

	templateContent := req.TemplateContent

	// Create CompilationContext from request with format file support
	// Strategy Pattern: Use format-aware constructor if format file provided
	var compCtx valueobjects.CompilationContext
	var err error

	if req.FormatFile != "" {
		// For format files, use the template file path (which points to processed content)
		// The V2 compiler will read from this file instead of using inline content
		compCtx, err = valueobjects.NewCompilationContextWithFormatFile(
			req.TemplatePath, // Use file path for format-aware compilation
			req.Engine,
			req.OutputPath,
			req.WorkingDir,
			req.FormatFile,
			req.Options.Debug.Enabled,
		)
	} else {
		compCtx, err = valueobjects.NewCompilationContextWithWorkDir(
			templateContent, // Use inline content for legacy compilation
			req.Engine,
			req.OutputPath,
			req.WorkingDir,
			req.Options.Debug.Enabled,
		)
	}

	if err != nil {
		return generation.PDFGenerationResult{
			Success: false,
			Error: domain.PDFGenerationError{
				Code:    domain.ErrCodePDFGenerationFailed,
				Message: "Failed to create compilation context",
				Details: api.NewErrorDetails(api.ErrorCategoryGeneration, api.ErrorSeverityHigh).
					WithError(err),
			},
		}, err
	}

	// Use V2 compiler
	outputPath, err := epsa.v2Compiler.Compile(ctx, compCtx)
	if err != nil {
		return generation.PDFGenerationResult{
			Success: false,
			Error: domain.PDFGenerationError{
				Code:    domain.ErrCodePDFGenerationFailed,
				Message: "V2 LaTeX compilation failed",
				Details: api.NewErrorDetails(api.ErrorCategoryGeneration, api.ErrorSeverityHigh).
					WithError(err),
			},
		}, err
	}

	// Return result
	return generation.PDFGenerationResult{
		Success:    true,
		PDFPath:    outputPath,
		ImagePaths: []string{}, // V2 compiler doesn't generate images by default
		Metadata: generation.PDFMetadata{
			FileSize:    0, // Would need to read file size
			PageCount:   1, // Would need proper PDF parsing
			GeneratedAt: time.Now(),
			Engine:      req.Engine,
			Template:    req.TemplatePath,
		},
	}, nil
}

// ValidateRequest validates a PDF generation request
func (epsa *ExternalPDFServiceV2Adapter) ValidateRequest(req generation.PDFGenerationRequest) error {
	if req.TemplatePath == "" {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodeTemplateNotFound,
			Message: api.ErrTemplatePathRequired,
			Details: api.NewErrorDetails(api.ErrorCategoryTemplate, api.ErrorSeverityHigh),
		}
	}

	if req.Engine == "" {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodeEngineNotFound,
			Message: api.ErrEngineRequired,
			Details: api.NewErrorDetails(api.ErrorCategoryGeneration, api.ErrorSeverityHigh),
		}
	}

	if req.OutputPath == "" {
		return domain.PDFGenerationError{
			Code:    domain.ErrCodeOutputPathInvalid,
			Message: api.ErrOutputPathRequired,
			Details: api.NewErrorDetails(api.ErrorCategoryGeneration, api.ErrorSeverityHigh),
		}
	}

	return nil
}

// GetSupportedEngines returns supported LaTeX engines
func (epsa *ExternalPDFServiceV2Adapter) GetSupportedEngines() []string {
	return []string{
		"pdflatex",
		"xelatex",
		"lualatex",
		"latex",
	}
}

// GetSupportedFormats returns supported output formats
func (epsa *ExternalPDFServiceV2Adapter) GetSupportedFormats() []string {
	return []string{
		"pdf",
	}
}
