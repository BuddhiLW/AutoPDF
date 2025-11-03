// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package watch

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	ports "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/ports"
	documentService "github.com/BuddhiLW/AutoPDF/internal/autopdf/application/services/document"
	configPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/config"
	wiringPkg "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/common/wiring"
)

// DocumentRebuildAdapter implements RebuildService port
// Following DIP: depends on abstraction (DocumentService), not concrete implementation
// Following CLARITY: composes services to orchestrate rebuild
type DocumentRebuildAdapter struct {
	configResolver *configPkg.ConfigResolver
	serviceBuilder *wiringPkg.ServiceBuilder
	logger         *logger.LoggerAdapter
}

// NewDocumentRebuildAdapter creates a new DocumentRebuildAdapter
// Following CLARITY: explicit dependencies via constructor (Dependency Injection)
func NewDocumentRebuildAdapter(
	configResolver *configPkg.ConfigResolver,
	serviceBuilder *wiringPkg.ServiceBuilder,
	logger *logger.LoggerAdapter,
) ports.RebuildService {
	return &DocumentRebuildAdapter{
		configResolver: configResolver,
		serviceBuilder: serviceBuilder,
		logger:         logger,
	}
}

// Rebuild orchestrates a full document rebuild
// Following SRP: single responsibility - orchestrate rebuild workflow
// Following CLARITY: represents rebuild intent clearly
func (d *DocumentRebuildAdapter) Rebuild(ctx context.Context, templatePath, configPath string) (ports.RebuildResult, error) {
	d.logger.InfoWithFields("Starting rebuild",
		"template", templatePath,
		"config", configPath,
	)

	// Step 1: Load configuration
	cfg, err := d.configResolver.LoadConfigWithLogging(ctx, templatePath, configPath)
	if err != nil {
		return ports.RebuildResult{
			Success: false,
			Error:   fmt.Errorf("failed to load configuration: %w", err),
		}, err
	}

	// Step 2: Create document service with template directory as working directory
	// This ensures LaTeX can find assets (.cls files, images) in template's directory
	templateDir := filepath.Dir(cfg.Template.String())
	svc := d.serviceBuilder.BuildDocumentServiceWithWorkingDir(cfg, templateDir)

	// Step 3: Build request directly (we have all needed information)
	// Following CLARITY: direct construction with clear intent
	req := documentService.BuildRequest{
		TemplatePath: cfg.Template.String(),
		ConfigPath:   configPath,
		Variables:    &cfg.Variables,
		Engine:       cfg.Engine.String(),
		OutputPath:   cfg.Output.String(),
		DoConvert:    cfg.Conversion.Enabled,
		DoClean:      false, // Don't clean in watch mode by default
		DebugEnabled: false,
		Conversion: documentService.ConversionSettings{
			Enabled: cfg.Conversion.Enabled,
			Formats: cfg.Conversion.Formats,
		},
	}

	// Step 4: Execute rebuild
	result, err := svc.Build(ctx, req)
	if err != nil {
		d.logger.ErrorWithFields("Rebuild failed",
			"template", templatePath,
			"config", configPath,
			"error", err,
		)
		return ports.RebuildResult{
			PDFPath: result.PDFPath,
			Success: false,
			Error:   err,
		}, err
	}

	d.logger.InfoWithFields("Rebuild completed successfully",
		"template", templatePath,
		"pdf_path", result.PDFPath,
	)

	return ports.RebuildResult{
		PDFPath: result.PDFPath,
		Success: result.Success,
		Error:   result.Error,
	}, nil
}
