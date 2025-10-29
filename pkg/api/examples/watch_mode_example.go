// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application/adapters/logger"
	"github.com/BuddhiLW/AutoPDF/pkg/api/application"
	"github.com/BuddhiLW/AutoPDF/pkg/api/builders"
	"github.com/BuddhiLW/AutoPDF/pkg/api/domain/generation"
	"github.com/BuddhiLW/AutoPDF/pkg/api/services"
)

// WatchModeExample demonstrates how to use watch mode with PDF generation
func WatchModeExample() {
	fmt.Println("=== Watch Mode Example ===")

	// Create logger
	logger := logger.NewLoggerAdapter(logger.Detailed, "stdout")

	// Create watch mode manager
	servicesWatchManager := services.NewWatchModeManager(logger)

	// Create adapter to match the application interface
	watchManager := &WatchModeManagerAdapter{
		manager: servicesWatchManager,
	}

	// Create mock services for demonstration
	// In a real application, these would be actual implementations
	mockTemplateService := &MockTemplateService{}
	mockVariableResolver := &MockVariableResolver{}
	mockPDFValidator := &MockPDFValidator{}
	mockExternalService := &MockPDFGenerationService{}

	// Create application service with watch mode support
	appService := application.NewPDFGenerationApplicationService(
		mockTemplateService,
		mockVariableResolver,
		mockPDFValidator,
		mockExternalService,
		nil, // watchService (not needed for this example)
		watchManager,
		logger,
		false, // Default debug to false for examples
	)

	// Create a PDF generation request with watch mode enabled
	request := builders.NewPDFGenerationRequestBuilder().
		WithTemplate("example.tex").
		WithVariables(map[string]interface{}{
			"title":   "Watch Mode Test",
			"author":  "AutoPDF",
			"content": "This is a test document with watch mode enabled.",
		}).
		WithEngine("pdflatex").
		WithWatchMode(true). // Enable watch mode
		Build()

	// Set output path manually since WithOutputPath doesn't exist
	request.OutputPath = "output.pdf"

	fmt.Printf("Generated request with watch mode: %t\n", request.Options.WatchMode)

	// Generate PDF with watch mode
	ctx := context.Background()
	result, err := appService.GeneratePDF(ctx, request)
	if err != nil {
		log.Printf("PDF generation failed: %v", err)
		return
	}

	fmt.Printf("PDF generation result: Success=%t, PDFPath=%s\n", result.Success, result.PDFPath)

	// Check active watch modes
	activeWatches := appService.GetActiveWatchModes()
	fmt.Printf("Active watch modes: %d\n", len(activeWatches))

	for watchID, info := range activeWatches {
		fmt.Printf("  Watch ID: %s\n", watchID)
		fmt.Printf("    Template: %s\n", info.TemplatePath)
		fmt.Printf("    Request ID: %s\n", info.RequestID)
		fmt.Printf("    Started: %s\n", info.StartedAt.Format(time.RFC3339))
		fmt.Printf("    Duration: %s\n", info.Duration)
	}

	// Wait a bit to see the watch mode in action
	fmt.Println("Waiting 10 seconds to observe watch mode...")
	time.Sleep(10 * time.Second)

	// Check active watch modes again
	activeWatches = appService.GetActiveWatchModes()
	fmt.Printf("Active watch modes after 10 seconds: %d\n", len(activeWatches))

	// Stop all watch modes
	err = appService.StopAllWatchModes()
	if err != nil {
		log.Printf("Failed to stop watch modes: %v", err)
		return
	}

	fmt.Println("All watch modes stopped successfully")

	// Verify no active watch modes
	activeWatches = appService.GetActiveWatchModes()
	fmt.Printf("Active watch modes after stopping: %d\n", len(activeWatches))

	fmt.Println("=== Watch Mode Example Complete ===")
}

// Mock services for demonstration
type MockTemplateService struct{}

func (m *MockTemplateService) Process(ctx context.Context, templatePath string, variables map[string]string) (string, error) {
	return "\\documentclass{article}\n\\begin{document}\n\\title{" + variables["title"] + "}\n\\author{" + variables["author"] + "}\n\\maketitle\n" + variables["content"] + "\n\\end{document}", nil
}

func (m *MockTemplateService) ValidateTemplate(templatePath string) error {
	return nil
}

func (m *MockTemplateService) GetTemplateVariables(templatePath string) ([]string, error) {
	return []string{"title", "author", "content"}, nil
}

type MockVariableResolver struct{}

func (m *MockVariableResolver) Resolve(variables *generation.TemplateVariables) (map[string]string, error) {
	if variables == nil {
		return make(map[string]string), nil
	}
	return variables.Flatten(), nil
}

func (m *MockVariableResolver) Validate(variables *generation.TemplateVariables) error {
	return nil
}

func (m *MockVariableResolver) Flatten(variables *generation.TemplateVariables) map[string]string {
	if variables == nil {
		return make(map[string]string)
	}
	return variables.Flatten()
}

type MockPDFValidator struct{}

func (m *MockPDFValidator) Validate(pdfPath string) error {
	return nil
}

func (m *MockPDFValidator) GetMetadata(pdfPath string) (generation.PDFMetadata, error) {
	return generation.PDFMetadata{
		FileSize:    1024,
		PageCount:   1,
		GeneratedAt: time.Now(),
		Engine:      "pdflatex",
		Template:    "example.tex",
	}, nil
}

func (m *MockPDFValidator) IsValidPDF(pdfPath string) bool {
	return true
}

type MockPDFGenerationService struct{}

func (m *MockPDFGenerationService) Generate(ctx context.Context, req generation.PDFGenerationRequest) (generation.PDFGenerationResult, error) {
	return generation.PDFGenerationResult{
		PDFPath:    "output.pdf",
		ImagePaths: []string{},
		Success:    true,
		Error:      nil,
		Metadata: generation.PDFMetadata{
			FileSize:    1024,
			PageCount:   1,
			GeneratedAt: time.Now(),
			Engine:      req.Engine,
			Template:    req.TemplatePath,
		},
	}, nil
}

func (m *MockPDFGenerationService) ValidateRequest(req generation.PDFGenerationRequest) error {
	return nil
}

func (m *MockPDFGenerationService) GetSupportedEngines() []string {
	return []string{"pdflatex", "xelatex", "lualatex"}
}

func (m *MockPDFGenerationService) GetSupportedFormats() []string {
	return []string{"pdf", "png", "jpeg", "svg"}
}

// WatchModeManagerAdapter adapts services.WatchModeManager to application.WatchModeManager
type WatchModeManagerAdapter struct {
	manager *services.WatchModeManager
}

func (a *WatchModeManagerAdapter) StartWatchMode(ctx context.Context, req generation.PDFGenerationRequest) error {
	return a.manager.StartWatchMode(ctx, req)
}

func (a *WatchModeManagerAdapter) StopWatchMode(watchID string) error {
	return a.manager.StopWatchMode(watchID)
}

func (a *WatchModeManagerAdapter) StopAllWatchModes() error {
	return a.manager.StopAllWatchModes()
}

func (a *WatchModeManagerAdapter) GetActiveWatches() map[string]generation.WatchInstanceInfo {
	servicesWatches := a.manager.GetActiveWatches()
	watches := make(map[string]generation.WatchInstanceInfo)
	for id, info := range servicesWatches {
		watches[id] = generation.WatchInstanceInfo{
			ID:           info.ID,
			TemplatePath: info.TemplatePath,
			RequestID:    info.RequestID,
			StartedAt:    info.StartedAt,
			Duration:     info.Duration,
		}
	}
	return watches
}
