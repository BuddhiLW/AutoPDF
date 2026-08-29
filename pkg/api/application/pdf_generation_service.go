// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"

	ports "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/ports"
	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/domain/watch"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/domain/generation"
)

// PDFGenerationApplicationService provides a clean interface for PDF generation
type PDFGenerationApplicationService struct {
	orchestrationService *PDFOrchestrationService
}

// NewPDFGenerationApplicationService creates a new application service
func NewPDFGenerationApplicationService(
	templateService generation.TemplateProcessingService,
	variableResolver generation.VariableResolver,
	pdfValidator generation.PDFValidator,
	externalService generation.PDFGenerationService,
	watchService watch.WatchService,
	watchManager generation.WatchModeManager,
	logger ports.Logger,
	debugEnabled bool, // Accept debug flag
) *PDFGenerationApplicationService {
	// Initialize orchestration service with all dependencies
	orchestrationService := NewPDFOrchestrationService(
		templateService,
		variableResolver,
		pdfValidator,
		externalService,
		watchService,
		watchManager,
		logger,
	)

	return &PDFGenerationApplicationService{
		orchestrationService: orchestrationService,
	}
}

// GeneratePDF orchestrates the complete PDF generation workflow
func (s *PDFGenerationApplicationService) GeneratePDF(ctx context.Context, req generation.PDFGenerationRequest) (generation.PDFGenerationResult, error) {
	return s.orchestrationService.GeneratePDF(ctx, req)
}

// ValidateTemplate validates a template file
func (s *PDFGenerationApplicationService) ValidateTemplate(templatePath string) error {
	return s.orchestrationService.ValidateTemplate(templatePath)
}

// GetTemplateVariables extracts variables from a template
func (s *PDFGenerationApplicationService) GetTemplateVariables(templatePath string) ([]string, error) {
	return s.orchestrationService.GetTemplateVariables(templatePath)
}

// GetSupportedEngines returns supported LaTeX engines
func (s *PDFGenerationApplicationService) GetSupportedEngines() []string {
	return s.orchestrationService.GetSupportedEngines()
}

// GetSupportedFormats returns supported output formats
func (s *PDFGenerationApplicationService) GetSupportedFormats() []string {
	return s.orchestrationService.GetSupportedFormats()
}

// GetActiveWatchModes returns information about active watch modes
func (s *PDFGenerationApplicationService) GetActiveWatchModes() map[string]generation.WatchInstanceInfo {
	return s.orchestrationService.GetActiveWatchModes()
}

// StopWatchMode stops a specific watch mode
func (s *PDFGenerationApplicationService) StopWatchMode(watchID string) error {
	return s.orchestrationService.StopWatchMode(watchID)
}

// StopAllWatchModes stops all active watch modes
func (s *PDFGenerationApplicationService) StopAllWatchModes() error {
	return s.orchestrationService.StopAllWatchModes()
}
