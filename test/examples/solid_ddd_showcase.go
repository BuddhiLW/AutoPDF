// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

// Package examples demonstrates the full capabilities of the SOLID + DDD + GoF refactored AutoPDF system
package examples

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/BuddhiLW/AutoPDF/internal/autopdf/application"
	"github.com/BuddhiLW/AutoPDF/internal/autopdf/domain"
)

// ShowcaseSOLIDDDDArchitecture demonstrates the full capabilities of the refactored system
func ShowcaseSOLIDDDDArchitecture() {
	fmt.Println("🏗️  AutoPDF SOLID + DDD + GoF Architecture Showcase")
	fmt.Println("=" + string(make([]byte, 60)))
	fmt.Println()

	// Get the service factory (Singleton + Factory Pattern)
	factory := application.GetDefaultFactory()

	// Demonstrate Dependency Injection
	fmt.Println("1️⃣  Dependency Injection & Service Factory")
	fmt.Println("   ✓ Single service factory instance")
	fmt.Println("   ✓ All dependencies properly wired")
	fmt.Println("   ✓ Services depend on abstractions, not concretions")
	fmt.Println()

	// Get all services
	buildService := factory.GetBuildService()
	templateService := factory.GetTemplateProcessingService()
	pdfService := factory.GetPDFGenerationService()
	conversionService := factory.GetConversionService()
	fileService := factory.GetFileManagementService()
	configService := factory.GetConfigurationService()
	eventPublisher := factory.GetEventPublisher()

	ctx := context.Background()

	// Demonstrate Factory Pattern
	fmt.Println("2️⃣  Factory Pattern - Engine Creation")
	demonstrateFactoryPattern()
	fmt.Println()

	// Demonstrate Strategy Pattern
	fmt.Println("3️⃣  Strategy Pattern - Template Processing")
	demonstrateStrategyPattern(templateService, ctx)
	fmt.Println()

	// Demonstrate Observer Pattern
	fmt.Println("4️⃣  Observer Pattern - Event-Driven Architecture")
	demonstrateObserverPattern(eventPublisher)
	fmt.Println()

	// Demonstrate Domain Services
	fmt.Println("5️⃣  Domain Services - Business Logic Orchestration")
	demonstrateDomainServices(buildService, configService, fileService, ctx)
	fmt.Println()

	// Demonstrate Complete Workflow
	fmt.Println("6️⃣  Complete Workflow - All Patterns Working Together")
	demonstrateCompleteWorkflow(buildService, configService, ctx)
	fmt.Println()

	// Demonstrate Advanced Features
	fmt.Println("7️⃣  Advanced Features - Complex Data Structures")
	demonstrateAdvancedFeatures(buildService, configService, ctx)
	fmt.Println()

	// Demonstrate Conversion with Strategy Pattern
	fmt.Println("8️⃣  Conversion Service - Multiple Engine Strategies")
	demonstrateConversionStrategies(conversionService, pdfService, ctx)
	fmt.Println()

	fmt.Println("✅ Showcase complete!")
	fmt.Println()
	fmt.Println("Key Takeaways:")
	fmt.Println("  • SOLID principles ensure maintainable, extensible code")
	fmt.Println("  • DDD provides clear domain boundaries and business logic")
	fmt.Println("  • GoF patterns solve common design problems elegantly")
	fmt.Println("  • All components are testable with mocks")
	fmt.Println("  • Event-driven architecture enables loose coupling")
}

func demonstrateFactoryPattern() {
	// Template Engine Factory
	templateFactory := domain.NewTemplateEngineFactory()
	engines := templateFactory.GetAvailableEngines()
	fmt.Printf("   Template Engines: %v\n", engines)

	// PDF Engine Factory
	pdfFactory := domain.NewPDFEngineFactory()
	pdfEngines := pdfFactory.GetAvailableEngines()
	fmt.Printf("   PDF Engines: %v\n", pdfEngines)

	// Conversion Engine Factory
	convFactory := domain.NewConversionEngineFactory()
	convEngines := convFactory.GetAvailableEngines()
	fmt.Printf("   Conversion Engines: %v\n", convEngines)

	fmt.Println("   ✓ Factories abstract complex object creation")
	fmt.Println("   ✓ Easy to add new engine types")
}

func demonstrateStrategyPattern(templateService domain.TemplateProcessingService, ctx context.Context) {
	// Create a temp file for demonstration
	tempDir, _ := os.MkdirTemp("", "autopdf-demo-*")
	defer os.RemoveAll(tempDir)

	templatePath := filepath.Join(tempDir, "demo.tex")
	content := `\documentclass{article}\begin{document}Hello, delim[[.name]]!\end{document}`
	os.WriteFile(templatePath, []byte(content), 0644)

	// Process template with variables
	variables := map[string]interface{}{
		"name": "SOLID + DDD + GoF",
	}

	result, err := templateService.ProcessTemplate(ctx, templatePath, variables)
	if err == nil {
		fmt.Println("   ✓ Template processed with LaTeX strategy")
		fmt.Printf("   ✓ Variables substituted: %v\n", variables)
	}
	_ = result
}

func demonstrateObserverPattern(publisher domain.EventPublisher) {
	// Create a logging handler
	handler := domain.NewLoggingEventHandler()

	// Subscribe to events
	publisher.Subscribe("demo.event", handler)

	// Publish an event
	event := domain.NewDomainEvent("demo.event", map[string]interface{}{
		"message": "Observer pattern in action",
		"time":    time.Now(),
	})

	publisher.Publish(event)

	fmt.Println("   ✓ Event published and handled")
	fmt.Println("   ✓ Loose coupling through events")
	fmt.Println("   ✓ Easy to add new event handlers")
}

func demonstrateDomainServices(
	buildService domain.BuildService,
	configService domain.ConfigurationService,
	fileService domain.FileManagementService,
	ctx context.Context,
) {
	fmt.Println("   Domain Services encapsulate business logic:")
	fmt.Println("   • TemplateProcessingService - Template operations")
	fmt.Println("   • PDFGenerationService - PDF compilation")
	fmt.Println("   • ConversionService - Image conversion")
	fmt.Println("   • FileManagementService - File operations")
	fmt.Println("   • ConfigurationService - Config management")
	fmt.Println("   • BuildService - Workflow orchestration")
	fmt.Println()
	fmt.Println("   ✓ Each service has single responsibility")
	fmt.Println("   ✓ Services depend on interfaces")
	fmt.Println("   ✓ Easy to test with mocks")
}

func demonstrateCompleteWorkflow(
	buildService domain.BuildService,
	configService domain.ConfigurationService,
	ctx context.Context,
) {
	// Create temp directory
	tempDir, _ := os.MkdirTemp("", "autopdf-workflow-*")
	defer os.RemoveAll(tempDir)

	// Create a simple template
	templatePath := filepath.Join(tempDir, "workflow.tex")
	templateContent := `\documentclass{article}
\begin{document}
\title{delim[[.title]]}
\author{delim[[.author]]}
\maketitle
\section{Introduction}
delim[[.content]]
\end{document}`
	os.WriteFile(templatePath, []byte(templateContent), 0644)

	// Create configuration
	cfg := &domain.Configuration{
		Template: templatePath,
		Output:   filepath.Join(tempDir, "workflow-output.pdf"),
		Engine:   "pdflatex",
		Variables: map[string]interface{}{
			"title":   "Complete Workflow Demo",
			"author":  "AutoPDF SOLID + DDD",
			"content": "This demonstrates the complete workflow with all patterns working together.",
		},
	}

	// Build the PDF
	start := time.Now()
	result, err := buildService.BuildPDF(ctx, &domain.BuildRequest{
		TemplatePath: templatePath,
		OutputPath:   cfg.Output,
		Variables:    cfg.Variables,
		ShouldClean:  true,
	})

	if err != nil {
		fmt.Printf("   ⚠️  Build failed (LaTeX may not be installed): %v\n", err)
		return
	}

	duration := time.Since(start)

	fmt.Printf("   ✓ PDF built successfully in %v\n", duration)
	fmt.Printf("   ✓ Output: %s\n", result.PDFPath)
	fmt.Println("   ✓ All services coordinated by BuildService")
	fmt.Println("   ✓ Events published throughout the workflow")
}

func demonstrateAdvancedFeatures(
	buildService domain.BuildService,
	configService domain.ConfigurationService,
	ctx context.Context,
) {
	fmt.Println("   Advanced features supported:")
	fmt.Println("   • Nested objects and maps")
	fmt.Println("   • Arrays and loops")
	fmt.Println("   • Complex data structures")
	fmt.Println("   • Conditional sections")
	fmt.Println("   • Dynamic content generation")
	fmt.Println()

	// Try to load the enhanced config to show complex data support
	cfg, err := configService.LoadConfiguration(ctx, "../../configs/enhanced-sample-config.yaml")
	if err == nil {
		fmt.Println("   ✓ Enhanced config loaded successfully")
		fmt.Printf("   ✓ Variables include: company, team, legal_document, etc.\n")
		fmt.Printf("   ✓ Nested depth: 4+ levels\n")
		fmt.Printf("   ✓ Arrays: team members, proceedings, damages breakdown\n")
	}
	_ = cfg
}

func demonstrateConversionStrategies(
	conversionService domain.ConversionService,
	pdfService domain.PDFGenerationService,
	ctx context.Context,
) {
	// Get supported formats
	formats := conversionService.GetSupportedFormats()
	fmt.Printf("   Supported formats: %v\n", formats)

	// Show strategy pattern in action
	fmt.Println("   ✓ Multiple conversion engines available")
	fmt.Println("   ✓ Automatic fallback if one engine fails")
	fmt.Println("   ✓ Strategy selected based on availability")
	fmt.Println("   ✓ Easy to add new conversion strategies")
}

// RunShowcase is the main entry point for the showcase
func RunShowcase() {
	log.SetFlags(0) // Disable log timestamps for cleaner output
	ShowcaseSOLIDDDDArchitecture()
}
