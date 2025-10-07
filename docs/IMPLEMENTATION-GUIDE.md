# AutoPDF DDD Refactoring Implementation Guide

## Overview

This guide provides step-by-step instructions for implementing the DDD refactoring of AutoPDF, following the ADRs and maintaining backward compatibility.

## Phase 1: Foundation Setup (Days 1-30)

### Step 1: Create Directory Structure

```bash
# Create the new DDD structure
mkdir -p internal/document/{domain/{entities,value_objects,services,repositories,events,errors},app,infra}
mkdir -p internal/conversion/{domain/{entities,value_objects,services},app,infra}
mkdir -p internal/configuration/{domain/{entities,value_objects},app,infra}
mkdir -p internal/legacy
mkdir -p pkg/{ports,features,events,errors}
mkdir -p test/{characterization,integration,unit}
```

### Step 2: Implement Core Port Interfaces

Create `pkg/ports/document.go`:

```go
package ports

import (
    "context"
    "time"
)

// DocumentRepository defines the contract for document persistence
type DocumentRepository interface {
    Save(ctx context.Context, doc Document) error
    FindByID(ctx context.Context, id DocumentID) (Document, error)
    Delete(ctx context.Context, id DocumentID) error
    FindByStatus(ctx context.Context, status DocumentStatus) ([]Document, error)
}

// TemplateProcessor defines the contract for template processing
type TemplateProcessor interface {
    Process(ctx context.Context, template TemplatePath, variables Variables) (ProcessedTemplate, error)
    Validate(ctx context.Context, template TemplatePath) error
}

// LaTeXCompiler defines the contract for LaTeX compilation
type LaTeXCompiler interface {
    Compile(ctx context.Context, template ProcessedTemplate, options CompilationOptions) (CompilationResult, error)
    ValidateEngine(ctx context.Context, engine EngineType) error
}

// FileSystem defines the contract for file operations
type FileSystem interface {
    ReadFile(ctx context.Context, path string) ([]byte, error)
    WriteFile(ctx context.Context, path string, data []byte) error
    DeleteFile(ctx context.Context, path string) error
    Exists(ctx context.Context, path string) (bool, error)
    CreateDirectory(ctx context.Context, path string) error
}

// EventPublisher defines the contract for domain event publishing
type EventPublisher interface {
    Publish(ctx context.Context, event DomainEvent) error
    Subscribe(ctx context.Context, eventType string, handler EventHandler) error
}

// Domain types
type DocumentID string
type DocumentStatus string
type TemplatePath string
type OutputPath string
type Variables map[string]string
type EngineType string
type ProcessedTemplate struct {
    Path    string
    Content string
}
type CompilationOptions struct {
    Engine     EngineType
    OutputPath OutputPath
    Clean      bool
}
type CompilationResult struct {
    Success    bool
    OutputPath OutputPath
    Errors     []string
    Warnings   []string
    Duration   time.Duration
}
type DomainEvent interface {
    EventID() string
    AggregateID() string
    OccurredAt() time.Time
    EventType() string
}
type EventHandler func(ctx context.Context, event DomainEvent) error
```

### Step 3: Implement Core Domain Entities

Create `internal/document/domain/entities/document.go`:

```go
package entities

import (
    "time"
    "github.com/BuddhiLW/AutoPDF/pkg/ports"
)

// Document is the root aggregate for document generation
type Document struct {
    id          ports.DocumentID
    template    ports.TemplatePath
    output      ports.OutputPath
    variables   ports.Variables
    engine      ports.EngineType
    status      ports.DocumentStatus
    compilation *CompilationResult
    createdAt   time.Time
    updatedAt   time.Time
    events      []ports.DomainEvent
}

// DocumentStatus constants
const (
    DocumentStatusPending    ports.DocumentStatus = "pending"
    DocumentStatusProcessing ports.DocumentStatus = "processing"
    DocumentStatusCompleted  ports.DocumentStatus = "completed"
    DocumentStatusFailed     ports.DocumentStatus = "failed"
)

// NewDocument creates a new document with validation
func NewDocument(
    id ports.DocumentID,
    template ports.TemplatePath,
    output ports.OutputPath,
    variables ports.Variables,
    engine ports.EngineType,
) (*Document, error) {
    if id == "" {
        return nil, errors.New("document ID cannot be empty")
    }
    if template == "" {
        return nil, errors.New("template path cannot be empty")
    }
    if output == "" {
        return nil, errors.New("output path cannot be empty")
    }
    
    return &Document{
        id:        id,
        template:  template,
        output:    output,
        variables: variables,
        engine:    engine,
        status:    DocumentStatusPending,
        createdAt: time.Now(),
        updatedAt: time.Now(),
        events:    make([]ports.DomainEvent, 0),
    }, nil
}

// Business methods
func (d *Document) StartProcessing() error {
    if d.status != DocumentStatusPending {
        return errors.New("document is not in pending status")
    }
    d.status = DocumentStatusProcessing
    d.updatedAt = time.Now()
    d.addEvent(NewDocumentProcessingStarted(d.id))
    return nil
}

func (d *Document) CompleteProcessing(result CompilationResult) error {
    if d.status != DocumentStatusProcessing {
        return errors.New("document is not in processing status")
    }
    d.status = DocumentStatusCompleted
    d.compilation = &result
    d.updatedAt = time.Now()
    d.addEvent(NewDocumentProcessingCompleted(d.id, string(d.output)))
    return nil
}

func (d *Document) FailProcessing(reason string) error {
    if d.status != DocumentStatusProcessing {
        return errors.New("document is not in processing status")
    }
    d.status = DocumentStatusFailed
    d.updatedAt = time.Now()
    d.addEvent(NewDocumentProcessingFailed(d.id, reason))
    return nil
}

// Getters
func (d *Document) ID() ports.DocumentID {
    return d.id
}

func (d *Document) Status() ports.DocumentStatus {
    return d.status
}

func (d *Document) Template() ports.TemplatePath {
    return d.template
}

func (d *Document) Output() ports.OutputPath {
    return d.output
}

func (d *Document) Variables() ports.Variables {
    return d.variables
}

func (d *Document) Engine() ports.EngineType {
    return d.engine
}

func (d *Document) CreatedAt() time.Time {
    return d.createdAt
}

func (d *Document) UpdatedAt() time.Time {
    return d.updatedAt
}

func (d *Document) Events() []ports.DomainEvent {
    return d.events
}

func (d *Document) ClearEvents() {
    d.events = make([]ports.DomainEvent, 0)
}

// Business rules
func (d *Document) CanBeProcessed() bool {
    return d.status == DocumentStatusPending
}

func (d *Document) IsCompleted() bool {
    return d.status == DocumentStatusCompleted
}

func (d *Document) IsFailed() bool {
    return d.status == DocumentStatusFailed
}

func (d *Document) IsProcessing() bool {
    return d.status == DocumentStatusProcessing
}

// Private methods
func (d *Document) addEvent(event ports.DomainEvent) {
    d.events = append(d.events, event)
}
```

### Step 4: Implement Value Objects

Create `internal/document/domain/value_objects/template_path.go`:

```go
package value_objects

import (
    "errors"
    "path/filepath"
    "strings"
)

// TemplatePath represents a validated template file path
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

func (tp TemplatePath) IsValid() bool {
    return tp.value != ""
}
```

### Step 5: Create Legacy Adapter

Create `internal/legacy/document_adapter.go`:

```go
package legacy

import (
    "context"
    "github.com/BuddhiLW/AutoPDF/internal/tex"
    "github.com/BuddhiLW/AutoPDF/pkg/ports"
)

// LegacyDocumentAdapter wraps the current implementation
type LegacyDocumentAdapter struct {
    buildCmd *tex.BuildCmd
}

func NewLegacyDocumentAdapter() *LegacyDocumentAdapter {
    return &LegacyDocumentAdapter{
        buildCmd: tex.BuildCmd,
    }
}

func (l *LegacyDocumentAdapter) GenerateDocument(ctx context.Context, cmd ports.GenerateDocumentCommand) (ports.GenerateDocumentResult, error) {
    // Convert new command to legacy format
    args := []string{string(cmd.TemplatePath), string(cmd.OutputPath)}
    if cmd.Clean {
        args = append(args, "clean")
    }
    
    // Execute legacy build command
    err := l.buildCmd.Do(l.buildCmd, args...)
    if err != nil {
        return ports.GenerateDocumentResult{
            Success: false,
            Error:   err,
        }, err
    }
    
    return ports.GenerateDocumentResult{
        Success:    true,
        OutputPath: cmd.OutputPath,
    }, nil
}
```

### Step 6: Implement Feature Flags

Create `pkg/features/feature_flags.go`:

```go
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

### Step 7: Create Application Service

Create `internal/document/app/generate_document_service.go`:

```go
package app

import (
    "context"
    "fmt"
    "time"
    "github.com/BuddhiLW/AutoPDF/internal/document/domain/entities"
    "github.com/BuddhiLW/AutoPDF/pkg/ports"
)

type GenerateDocumentService struct {
    documentRepo    ports.DocumentRepository
    templateProc    ports.TemplateProcessor
    latexCompiler   ports.LaTeXCompiler
    fileSystem      ports.FileSystem
    eventPublisher  ports.EventPublisher
    featureFlags    *features.FeatureFlags
}

func NewGenerateDocumentService(
    documentRepo ports.DocumentRepository,
    templateProc ports.TemplateProcessor,
    latexCompiler ports.LaTeXCompiler,
    fileSystem ports.FileSystem,
    eventPublisher ports.EventPublisher,
    featureFlags *features.FeatureFlags,
) *GenerateDocumentService {
    return &GenerateDocumentService{
        documentRepo:   documentRepo,
        templateProc:   templateProc,
        latexCompiler:  latexCompiler,
        fileSystem:     fileSystem,
        eventPublisher: eventPublisher,
        featureFlags:   featureFlags,
    }
}

type GenerateDocumentCommand struct {
    TemplatePath ports.TemplatePath
    OutputPath   ports.OutputPath
    Variables    ports.Variables
    Engine       ports.EngineType
    Clean        bool
}

type GenerateDocumentResult struct {
    DocumentID ports.DocumentID
    OutputPath ports.OutputPath
    Success    bool
    Error      error
}

func (s *GenerateDocumentService) Execute(ctx context.Context, cmd GenerateDocumentCommand) (GenerateDocumentResult, error) {
    // Create document entity
    docID := ports.DocumentID(fmt.Sprintf("doc_%d", time.Now().UnixNano()))
    
    document, err := entities.NewDocument(
        docID,
        cmd.TemplatePath,
        cmd.OutputPath,
        cmd.Variables,
        cmd.Engine,
    )
    if err != nil {
        return GenerateDocumentResult{Error: err}, err
    }
    
    // Save document
    if err := s.documentRepo.Save(ctx, document); err != nil {
        return GenerateDocumentResult{Error: err}, err
    }
    
    // Start processing
    if err := document.StartProcessing(); err != nil {
        return GenerateDocumentResult{Error: err}, err
    }
    
    // Process template
    processedTemplate, err := s.templateProc.Process(ctx, cmd.TemplatePath, cmd.Variables)
    if err != nil {
        document.FailProcessing(err.Error())
        s.documentRepo.Save(ctx, document)
        return GenerateDocumentResult{Error: err}, err
    }
    
    // Compile LaTeX
    compilationOptions := ports.CompilationOptions{
        Engine:     cmd.Engine,
        OutputPath: cmd.OutputPath,
        Clean:      cmd.Clean,
    }
    
    result, err := s.latexCompiler.Compile(ctx, processedTemplate, compilationOptions)
    if err != nil {
        document.FailProcessing(err.Error())
        s.documentRepo.Save(ctx, document)
        return GenerateDocumentResult{Error: err}, err
    }
    
    // Complete processing
    compilationResult := entities.NewCompilationResult(
        entities.CompilationResultID(fmt.Sprintf("comp_%d", time.Now().UnixNano())),
        docID,
        cmd.OutputPath,
        result.Success,
        result.Duration,
    )
    
    if err := document.CompleteProcessing(*compilationResult); err != nil {
        return GenerateDocumentResult{Error: err}, err
    }
    
    // Save final state
    if err := s.documentRepo.Save(ctx, document); err != nil {
        return GenerateDocumentResult{Error: err}, err
    }
    
    // Publish events
    for _, event := range document.Events() {
        s.eventPublisher.Publish(ctx, event)
    }
    
    return GenerateDocumentResult{
        DocumentID: docID,
        OutputPath: cmd.OutputPath,
        Success:    true,
    }, nil
}
```

### Step 8: Create Characterization Tests

Create `test/characterization/document_generation_test.go`:

```go
package characterization

import (
    "testing"
    "github.com/BuddhiLW/AutoPDF/internal/legacy"
)

func TestDocumentGeneration_GoldenTests(t *testing.T) {
    tests := []struct {
        name           string
        templatePath   string
        configPath     string
        expectedOutput string
    }{
        {
            name:           "basic_document_generation",
            templatePath:   "testdata/template.tex",
            configPath:     "testdata/config.yaml",
            expectedOutput: "testdata/expected_output.pdf",
        },
        {
            name:           "document_with_variables",
            templatePath:   "testdata/template_with_vars.tex",
            configPath:     "testdata/config_with_vars.yaml",
            expectedOutput: "testdata/expected_output_with_vars.pdf",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Capture current behavior
            result := legacy.GenerateDocument(tt.templatePath, tt.configPath)
            
            // Verify output matches expected
            if !filesEqual(result.OutputPath, tt.expectedOutput) {
                t.Errorf("Output differs from expected. Got: %s, Expected: %s", 
                    result.OutputPath, tt.expectedOutput)
            }
        })
    }
}
```

## Phase 2: Domain Extraction (Days 31-60)

### Step 1: Implement Infrastructure Adapters

Create `internal/document/infra/latex_compiler_adapter.go`:

```go
package infra

import (
    "context"
    "fmt"
    "os/exec"
    "path/filepath"
    "time"
    "github.com/BuddhiLW/AutoPDF/pkg/ports"
)

type LaTeXCompilerAdapter struct {
    fileSystem ports.FileSystem
}

func NewLaTeXCompilerAdapter(fileSystem ports.FileSystem) *LaTeXCompilerAdapter {
    return &LaTeXCompilerAdapter{
        fileSystem: fileSystem,
    }
}

func (l *LaTeXCompilerAdapter) Compile(ctx context.Context, template ports.ProcessedTemplate, options ports.CompilationOptions) (ports.CompilationResult, error) {
    start := time.Now()
    
    // Create output directory if it doesn't exist
    outputDir := filepath.Dir(string(options.OutputPath))
    if err := l.fileSystem.CreateDirectory(ctx, outputDir); err != nil {
        return ports.CompilationResult{}, fmt.Errorf("failed to create output directory: %w", err)
    }
    
    // Write processed template to temporary file
    tempFile := filepath.Join(outputDir, "temp_"+filepath.Base(template.Path))
    if err := l.fileSystem.WriteFile(ctx, tempFile, []byte(template.Content)); err != nil {
        return ports.CompilationResult{}, fmt.Errorf("failed to write template: %w", err)
    }
    
    // Execute LaTeX compilation
    cmd := exec.CommandContext(ctx, string(options.Engine), 
        "-interaction=nonstopmode",
        "-jobname="+filepath.Base(string(options.OutputPath)),
        "-output-directory="+outputDir,
        tempFile,
    )
    
    if err := cmd.Run(); err != nil {
        return ports.CompilationResult{
            Success:    false,
            OutputPath: options.OutputPath,
            Errors:     []string{err.Error()},
            Duration:   time.Since(start),
        }, nil
    }
    
    return ports.CompilationResult{
        Success:    true,
        OutputPath: options.OutputPath,
        Duration:   time.Since(start),
    }, nil
}
```

### Step 2: Implement Repository Adapters

Create `internal/document/infra/memory_document_repository.go`:

```go
package infra

import (
    "context"
    "sync"
    "github.com/BuddhiLW/AutoPDF/internal/document/domain/entities"
    "github.com/BuddhiLW/AutoPDF/pkg/ports"
)

type MemoryDocumentRepository struct {
    documents map[ports.DocumentID]*entities.Document
    mutex     sync.RWMutex
}

func NewMemoryDocumentRepository() *MemoryDocumentRepository {
    return &MemoryDocumentRepository{
        documents: make(map[ports.DocumentID]*entities.Document),
    }
}

func (r *MemoryDocumentRepository) Save(ctx context.Context, doc *entities.Document) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    r.documents[doc.ID()] = doc
    return nil
}

func (r *MemoryDocumentRepository) FindByID(ctx context.Context, id ports.DocumentID) (*entities.Document, error) {
    r.mutex.RLock()
    defer r.mutex.RUnlock()
    
    doc, exists := r.documents[id]
    if !exists {
        return nil, errors.New("document not found")
    }
    
    return doc, nil
}

func (r *MemoryDocumentRepository) Delete(ctx context.Context, id ports.DocumentID) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    delete(r.documents, id)
    return nil
}

func (r *MemoryDocumentRepository) FindByStatus(ctx context.Context, status ports.DocumentStatus) ([]*entities.Document, error) {
    r.mutex.RLock()
    defer r.mutex.RUnlock()
    
    var result []*entities.Document
    for _, doc := range r.documents {
        if doc.Status() == status {
            result = append(result, doc)
        }
    }
    
    return result, nil
}
```

## Phase 3: Full Migration (Days 61-90)

### Step 1: Implement Event Handling

Create `pkg/events/event_bus.go`:

```go
package events

import (
    "context"
    "sync"
    "github.com/BuddhiLW/AutoPDF/pkg/ports"
)

type EventBus struct {
    handlers map[string][]ports.EventHandler
    mutex    sync.RWMutex
}

func NewEventBus() *EventBus {
    return &EventBus{
        handlers: make(map[string][]ports.EventHandler),
    }
}

func (eb *EventBus) Publish(ctx context.Context, event ports.DomainEvent) error {
    eb.mutex.RLock()
    handlers := eb.handlers[event.EventType()]
    eb.mutex.RUnlock()
    
    for _, handler := range handlers {
        if err := handler(ctx, event); err != nil {
            return err
        }
    }
    
    return nil
}

func (eb *EventBus) Subscribe(ctx context.Context, eventType string, handler ports.EventHandler) error {
    eb.mutex.Lock()
    defer eb.mutex.Unlock()
    
    eb.handlers[eventType] = append(eb.handlers[eventType], handler)
    return nil
}
```

### Step 2: Add Architectural Fitness Checks

Create `scripts/architectural_fitness.sh`:

```bash
#!/bin/bash

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

### Step 3: Implement CI/CD Pipeline

Create `.github/workflows/architectural-fitness.yml`:

```yaml
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

## Testing Strategy

### Unit Tests
- Test domain entities and value objects in isolation
- Test domain services with mocked dependencies
- Test business rules and validation

### Integration Tests
- Test application services with real infrastructure
- Test event handling and publishing
- Test repository implementations

### Characterization Tests
- Capture current behavior with golden tests
- Ensure refactoring doesn't break functionality
- Performance regression testing

## Success Metrics

1. **Test Coverage**: Maintain or improve test coverage
2. **Performance**: No degradation in compilation time
3. **Backward Compatibility**: All existing functionality works
4. **Code Quality**: Improved maintainability and extensibility
5. **Documentation**: Clear domain language and architecture

## Next Steps

1. Implement Phase 1 foundation
2. Add comprehensive tests
3. Begin gradual migration
4. Monitor performance and quality metrics
5. Iterate and improve based on feedback
