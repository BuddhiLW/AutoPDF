# Refined Implementation Example

## Overview

This document provides a concrete implementation example that addresses all the feedback points from the DDD refactoring review. It demonstrates the refined approach with proper type boundaries, naming conventions, and robust testing.

## 1. Type Boundaries (Option A: Pure Transport Types)

### Port Interfaces (Pure Transport Types)
```go
// pkg/ports/document.go
package ports

import "context"

// DocumentRepository - pure transport types
type DocumentRepository interface {
    Save(ctx context.Context, doc DocumentData) error
    FindByID(ctx context.Context, id string) (DocumentData, error)
    Delete(ctx context.Context, id string) error
}

// TemplateProcessor - pure transport types
type TemplateProcessor interface {
    Process(ctx context.Context, templatePath string, variables map[string]string) (string, error)
    Validate(ctx context.Context, templatePath string) error
}

// LaTeXCompiler - pure transport types
type LaTeXCompiler interface {
    Compile(ctx context.Context, templateContent string, options CompilationOptions) (CompilationResult, error)
}

// Transport DTOs
type DocumentData struct {
    ID        string
    Template  string
    Output    string
    Variables map[string]string
    Engine    string
    Status    string
    CreatedAt string
    UpdatedAt string
}

type CompilationOptions struct {
    Engine     string
    OutputPath string
    Clean      bool
}

type CompilationResult struct {
    Success    bool
    OutputPath string
    Errors     []string
    Warnings   []string
    Duration   string
}
```

### Domain Value Objects (Internal)
```go
// internal/document/domain/valueobjects/template_path.go
package valueobjects

import (
    "errors"
    "path/filepath"
    "strings"
)

// TemplatePath is a domain value object with business validation
type TemplatePath struct {
    value string
}

func NewTemplatePath(path string) (TemplatePath, error) {
    if strings.TrimSpace(path) == "" {
        return TemplatePath{}, errors.New("template path cannot be empty")
    }
    
    if !strings.HasSuffix(strings.ToLower(path), ".tex") {
        return TemplatePath{}, errors.New("template must be a .tex file")
    }
    
    return TemplatePath{value: path}, nil
}

func (tp TemplatePath) String() string {
    return tp.value
}

func (tp TemplatePath) BaseName() string {
    return filepath.Base(tp.value)
}

func (tp TemplatePath) Dir() string {
    return filepath.Dir(tp.value)
}

func (tp TemplatePath) Equals(other TemplatePath) bool {
    return tp.value == other.value
}
```

### Adapter Translation
```go
// internal/document/infra/template_processor_adapter.go
package infra

import (
    "context"
    "github.com/BuddhiLW/AutoPDF/internal/document/domain/valueobjects"
    "github.com/BuddhiLW/AutoPDF/pkg/ports"
)

type TemplateProcessorAdapter struct {
    domainService *services.TemplateProcessor
}

func (tpa *TemplateProcessorAdapter) Process(
    ctx context.Context, 
    templatePath string, 
    variables map[string]string,
) (string, error) {
    // Convert transport types to domain types
    domainTemplatePath, err := valueobjects.NewTemplatePath(templatePath)
    if err != nil {
        return "", err
    }
    
    domainVariables := valueobjects.NewVariables(variables)
    
    // Use domain service
    processedTemplate, err := tpa.domainService.ProcessTemplate(ctx, domainTemplatePath, domainVariables)
    if err != nil {
        return "", err
    }
    
    // Convert back to transport type
    return processedTemplate.Content(), nil
}
```

## 2. Naming Conventions (Single Word Packages)

### Directory Structure
```bash
# Create consistent directory structure
mkdir -p internal/document/{domain/{entities,valueobjects,services,repositories,events,errors},app,infra}
mkdir -p internal/conversion/{domain/{entities,valueobjects,services},app,infra}
mkdir -p internal/configuration/{domain/{entities,valueobjects},app,infra}
```

### Package Organization
```go
// internal/document/domain/valueobjects/template_path.go
package valueobjects

// internal/document/domain/entities/document.go
package entities

// internal/document/domain/services/template_processor.go
package services

// internal/document/app/generate_document_service.go
package app

// internal/document/infra/latex_compiler_adapter.go
package infra
```

## 3. Eventing Scope (Simplified for Phase 1)

### Domain Events (No Publishing)
```go
// internal/document/domain/events/document_events.go
package events

import "time"

type DocumentProcessingStarted struct {
    DocumentID string
    OccurredAt time.Time
}

type DocumentProcessingCompleted struct {
    DocumentID string
    OutputPath string
    OccurredAt time.Time
}

// Simple event creation - no publishing
func NewDocumentProcessingStarted(documentID string) *DocumentProcessingStarted {
    return &DocumentProcessingStarted{
        DocumentID: documentID,
        OccurredAt: time.Now(),
    }
}
```

### Aggregate Event Storage
```go
// internal/document/domain/entities/document.go
package entities

type Document struct {
    // ... other fields
    events []DomainEvent
}

// Add event to aggregate
func (d *Document) addEvent(event DomainEvent) {
    d.events = append(d.events, event)
}

// Get events for persistence
func (d *Document) Events() []DomainEvent {
    return d.events
}

// Clear events after processing
func (d *Document) ClearEvents() {
    d.events = make([]DomainEvent, 0)
}
```

## 4. CLI Contract Compliance (Positional Args Only)

### CLI Command Structure
```go
// internal/autopdf/cmd.go
var Cmd = &bonzai.Cmd{
    Name:    "autopdf",
    Summary: "Generate PDFs from LaTeX templates",
    Usage:   "autopdf <command> [args...]",
    Commands: []*bonzai.Cmd{
        BuildCmd,    // autopdf build <template> <config> [clean]
        CleanCmd,    // autopdf clean <directory>
        ConvertCmd,  // autopdf convert <pdf> [formats...]
    },
}

// Build command - positional args only
var BuildCmd = &bonzai.Cmd{
    Name:    "build",
    Summary: "Build PDF from template and config",
    Usage:   "autopdf build <template> <config> [clean]",
    Commands: []*bonzai.Cmd{},
}
```

### Feature Toggles via Environment Variables
```go
// pkg/features/feature_flags.go
package features

import (
    "os"
    "strconv"
)

type FeatureFlags struct {
    UseNewDocumentGeneration bool
    UseNewTemplateProcessing bool
    UseNewLaTeXCompilation   bool
}

func NewFeatureFlags() *FeatureFlags {
    return &FeatureFlags{
        UseNewDocumentGeneration: getBoolEnv("AUTOPDF_USE_NEW_DOCUMENT_GEN", false),
        UseNewTemplateProcessing: getBoolEnv("AUTOPDF_USE_NEW_TEMPLATE_PROC", false),
        UseNewLaTeXCompilation:   getBoolEnv("AUTOPDF_USE_NEW_LATEX_COMP", false),
    }
}

func getBoolEnv(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if parsed, err := strconv.ParseBool(value); err == nil {
            return parsed
        }
    }
    return defaultValue
}
```

## 5. Robust Characterization Tests

### Template Rendering Tests
```go
// test/characterization/template_rendering_test.go
package characterization

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestTemplateRendering_GoldenTests(t *testing.T) {
    tests := []struct {
        name           string
        templatePath   string
        configPath     string
        expectedTokens []string
    }{
        {
            name:         "basic_template_rendering",
            templatePath: "testdata/template.tex",
            configPath:   "testdata/config.yaml",
            expectedTokens: []string{
                "Test Document",
                "This is a test",
                "\\documentclass{scrartcl}",
            },
        },
        {
            name:         "template_with_variables",
            templatePath: "testdata/template_with_vars.tex",
            configPath:   "testdata/config_with_vars.yaml",
            expectedTokens: []string{
                "Hello World",
                "Custom Title",
                "\\title{Custom Title}",
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Process template
            processedContent, err := processTemplate(tt.templatePath, tt.configPath)
            require.NoError(t, err)
            
            // Verify expected tokens are present
            for _, token := range tt.expectedTokens {
                assert.Contains(t, processedContent, token, "Expected token not found: %s", token)
            }
        })
    }
}
```

### PDF Existence and Size Tests
```go
// test/characterization/pdf_generation_test.go
package characterization

import (
    "os"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestPDFGeneration_ExistenceAndSize(t *testing.T) {
    tests := []struct {
        name           string
        templatePath   string
        configPath     string
        minSizeBytes   int64
        maxSizeBytes   int64
    }{
        {
            name:         "basic_pdf_generation",
            templatePath: "testdata/template.tex",
            configPath:   "testdata/config.yaml",
            minSizeBytes: 1000,  // At least 1KB
            maxSizeBytes: 100000, // At most 100KB
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Generate PDF
            outputPath, err := generatePDF(tt.templatePath, tt.configPath)
            require.NoError(t, err)
            
            // Check file exists
            assert.FileExists(t, outputPath)
            
            // Check file size
            stat, err := os.Stat(outputPath)
            require.NoError(t, err)
            
            assert.GreaterOrEqual(t, stat.Size(), tt.minSizeBytes, "PDF too small")
            assert.LessOrEqual(t, stat.Size(), tt.maxSizeBytes, "PDF too large")
        })
    }
}
```

### Public Contract Tests
```go
// test/contract/cli_contract_test.go
package contract

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestCLIContract_PositionalArgs(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        expected int // exit code
    }{
        {
            name:     "build with template and config",
            args:     []string{"build", "template.tex", "config.yaml"},
            expected: 0,
        },
        {
            name:     "build with clean option",
            args:     []string{"build", "template.tex", "config.yaml", "clean"},
            expected: 0,
        },
        {
            name:     "clean directory",
            args:     []string{"clean", "/tmp/output"},
            expected: 0,
        },
        {
            name:     "convert PDF to images",
            args:     []string{"convert", "output.pdf", "png", "jpg"},
            expected: 0,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test CLI behavior
            result := runCLICommand(tt.args...)
            assert.Equal(t, tt.expected, result.ExitCode)
        })
    }
}

func TestCLIContract_NoFlags(t *testing.T) {
    // Ensure no flags are accepted
    result := runCLICommand("build", "--help") // Should fail
    assert.NotEqual(t, 0, result.ExitCode)
}
```

## 6. Architectural Fitness Checks

### Import Boundary Checks
```bash
#!/bin/bash
# scripts/architectural_fitness.sh

# Check that domain doesn't import infrastructure
if grep -r "github.com/BuddhiLW/AutoPDF/internal/.*/infra" internal/*/domain/; then
    echo "ERROR: Domain layer imports infrastructure"
    exit 1
fi

# Check that domain doesn't import application
if grep -r "github.com/BuddhiLW/AutoPDF/internal/.*/app" internal/*/domain/; then
    echo "ERROR: Domain layer imports application"
    exit 1
fi

# Check that application doesn't import infrastructure directly
if grep -r "github.com/BuddhiLW/AutoPDF/internal/.*/infra" internal/*/app/; then
    echo "ERROR: Application layer imports infrastructure directly"
    exit 1
fi

echo "Architectural fitness checks passed"
```

### CI/CD Pipeline
```yaml
# .github/workflows/architectural-fitness.yml
name: Architectural Fitness Checks

on: [push, pull_request]

jobs:
  architectural-fitness:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - name: Run architectural fitness checks
      run: |
        chmod +x scripts/architectural_fitness.sh
        ./scripts/architectural_fitness.sh
    - name: Run tests
      run: go test ./...
    - name: Run coverage
      run: go test -coverprofile=coverage.out ./...
    - name: Upload coverage
      uses: codecov/codecov-action@v1
```

## 7. Concrete Next Moves (Low-Risk PRs)

### PR 1: Lock Type Boundaries
```bash
# Create port interfaces with pure transport types
mkdir -p pkg/ports
# Implement port interfaces
# Add import boundary checks
```

### PR 2: Adapters First
```bash
# Wrap current template, tex, converter into port adapters
# Point CLI to application service
# Behavior unchanged
```

### PR 3: Public Contract Tests
```bash
# Add CLI tests for positional args
# Test exit codes and output presence
# These are "don't-break" tests
```

### PR 4: Stabilize Characterization
```bash
# Switch golden tests to template content checks
# Add PDF existence and size tests
# Move assertions to domain/app layers
```

### PR 5: ADR Accept + Fitness Checks
```bash
# Merge import-police script
# Mark ADR-001/002 as accepted
# Add success criteria
```

## Benefits of This Refined Approach

1. **Type Safety**: Clear boundaries between transport and domain types
2. **Consistency**: Uniform naming and package structure
3. **Simplicity**: Deferred complexity until needed
4. **Compliance**: Maintains Bonzai CLI contract
5. **Robustness**: Reliable tests that don't break with LaTeX updates
6. **Quality**: Architectural fitness checks prevent regressions

## Success Metrics

- **Type Boundaries**: No import cycles, clear ownership
- **Naming**: Consistent single-word packages
- **Eventing**: Simple event storage, no premature complexity
- **CLI**: Positional args only, env vars for features
- **Tests**: Robust, fast, reliable characterization tests
- **Architecture**: Clear boundaries enforced by CI

This refined implementation addresses all the feedback points while maintaining the core DDD principles and ensuring a smooth, safe migration path.
