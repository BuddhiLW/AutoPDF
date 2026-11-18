// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package application_test

import (
	"context"
	"fmt"
	"os"

	"github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/application/adapters/logger"
	infraadapters "github.com/BuddhiLW/AutoPDF/v2/internal/autopdf/infrastructure/adapters"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/application"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/builders"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/domain/generation"
	"github.com/BuddhiLW/AutoPDF/v2/pkg/api/services"
)

type templates struct{}

func (templates) Process(_ context.Context, _ string, vars map[string]string) (string, error) {
	return `\documentclass{article}\begin{document}` + vars["title"] + `\end{document}`, nil
}
func (templates) ValidateTemplate(string) error { return nil }
func (templates) GetTemplateVariables(string) ([]string, error) {
	return []string{"title"}, nil
}

type resolver struct{}

func (resolver) Resolve(v *generation.TemplateVariables) (map[string]string, error) {
	return v.Flatten(), nil
}
func (resolver) Flatten(v *generation.TemplateVariables) map[string]string { return v.Flatten() }
func (resolver) Validate(*generation.TemplateVariables) error              { return nil }

type validator struct{}

func (validator) Validate(string) error { return nil }
func (validator) GetMetadata(path string) (generation.PDFMetadata, error) {
	return generation.PDFMetadata{PageCount: 1, Template: path}, nil
}
func (validator) IsValidPDF(string) bool { return true }

type compiler struct{}

func (compiler) Generate(_ context.Context, req generation.PDFGenerationRequest) (generation.PDFGenerationResult, error) {
	return generation.PDFGenerationResult{PDFPath: req.OutputPath, Success: true}, nil
}
func (compiler) ValidateRequest(generation.PDFGenerationRequest) error { return nil }
func (compiler) GetSupportedEngines() []string                         { return []string{"pdflatex"} }
func (compiler) GetSupportedFormats() []string                         { return []string{"pdf"} }

type watchManager struct{ *services.WatchModeManager }

func (w watchManager) GetActiveWatches() map[string]generation.WatchInstanceInfo {
	out := make(map[string]generation.WatchInstanceInfo)
	for id, info := range w.WatchModeManager.GetActiveWatches() {
		out[id] = generation.WatchInstanceInfo{
			ID:           info.ID,
			TemplatePath: info.TemplatePath,
			RequestID:    info.RequestID,
			StartedAt:    info.StartedAt,
			Duration:     info.Duration,
		}
	}
	return out
}

func ExampleNewPDFGenerationApplicationService_watchMode() {
	portLogger := infraadapters.NewLoggerPortAdapter(logger.NewLoggerAdapter(logger.Detailed, os.DevNull))
	watches := watchManager{services.NewWatchModeManager(portLogger)}

	app := application.NewPDFGenerationApplicationService(
		templates{}, resolver{}, validator{}, compiler{},
		nil, watches, portLogger, false,
	)

	req := builders.NewPDFGenerationRequestBuilder().
		WithTemplate("example.tex").
		WithOutput("example.pdf").
		WithVariable("title", "Watch Mode").
		WithEngine("pdflatex").
		WithWatchMode(true).
		Build()

	result, err := app.GeneratePDF(context.Background(), req)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("pdf:", result.PDFPath, "success:", result.Success)
	fmt.Println("active watches:", len(app.GetActiveWatchModes()))
	fmt.Println("stop all:", app.StopAllWatchModes())
	// Output:
	// pdf: example.pdf success: true
	// active watches: 0
	// stop all: <nil>
}
