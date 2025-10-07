# ADR-002: Phase 1 Implementation - Foundation & Seams

## Status
Accepted

## Context
Following ADR-001, we need to implement Phase 1 of the DDD refactoring. This phase focuses on creating the foundation and establishing seams for safe migration.

## Decision

We will implement Phase 1 with the following components:

### 1. Hexagonal Architecture Skeleton

#### Domain Layer Structure
```
internal/
├── document/                    # Document Generation Bounded Context
│   ├── domain/
│   │   ├── entities/
│   │   │   ├── document.go
│   │   │   ├── template.go
│   │   │   └── compilation_result.go
│   │   ├── value_objects/
│   │   │   ├── template_path.go
│   │   │   ├── output_path.go
│   │   │   ├── variables.go
│   │   │   └── engine_type.go
│   │   ├── services/
│   │   │   └── template_processor.go
│   │   └── repositories/
│   │       └── document_repository.go
│   ├── app/
│   │   ├── generate_document.go
│   │   └── clean_document.go
│   └── infra/
│       ├── latex_compiler.go
│       ├── template_engine.go
│       └── file_system.go
```

### 2. Port Interfaces (Dependency Inversion)

#### Document Generation Ports
```go
// pkg/ports/document.go
package ports

import "context"

// DocumentRepository defines the contract for document persistence
type DocumentRepository interface {
    Save(ctx context.Context, doc Document) error
    FindByID(ctx context.Context, id DocumentID) (Document, error)
    Delete(ctx context.Context, id DocumentID) error
}

// TemplateProcessor defines the contract for template processing
type TemplateProcessor interface {
    Process(ctx context.Context, template Template, variables Variables) (ProcessedTemplate, error)
}

// LaTeXCompiler defines the contract for LaTeX compilation
type LaTeXCompiler interface {
    Compile(ctx context.Context, template ProcessedTemplate, options CompilationOptions) (CompilationResult, error)
}

// FileSystem defines the contract for file operations
type FileSystem interface {
    ReadFile(ctx context.Context, path string) ([]byte, error)
    WriteFile(ctx context.Context, path string, data []byte) error
    DeleteFile(ctx context.Context, path string) error
    Exists(ctx context.Context, path string) (bool, error)
}
```

### 3. Domain Entities and Value Objects

#### Document Entity
```go
// internal/document/domain/entities/document.go
package entities

import (
    "time"
    "github.com/BuddhiLW/AutoPDF/pkg/ports"
)

type DocumentID string

type Document struct {
    id          DocumentID
    template    ports.TemplatePath
    output      ports.OutputPath
    variables   ports.Variables
    engine      ports.EngineType
    status      DocumentStatus
    createdAt   time.Time
    updatedAt   time.Time
}

type DocumentStatus string

const (
    DocumentStatusPending    DocumentStatus = "pending"
    DocumentStatusProcessing DocumentStatus = "processing"
    DocumentStatusCompleted  DocumentStatus = "completed"
    DocumentStatusFailed     DocumentStatus = "failed"
)

func NewDocument(id DocumentID, template ports.TemplatePath, output ports.OutputPath, variables ports.Variables, engine ports.EngineType) *Document {
    return &Document{
        id:        id,
        template:  template,
        output:    output,
        variables: variables,
        engine:    engine,
        status:    DocumentStatusPending,
        createdAt: time.Now(),
        updatedAt: time.Now(),
    }
}

func (d *Document) ID() DocumentID {
    return d.id
}

func (d *Document) Status() DocumentStatus {
    return d.status
}

func (d *Document) MarkAsProcessing() {
    d.status = DocumentStatusProcessing
    d.updatedAt = time.Now()
}

func (d *Document) MarkAsCompleted() {
    d.status = DocumentStatusCompleted
    d.updatedAt = time.Now()
}

func (d *Document) MarkAsFailed() {
    d.status = DocumentStatusFailed
    d.updatedAt = time.Now()
}
```

#### Value Objects
```go
// internal/document/domain/value_objects/template_path.go
package value_objects

import (
    "errors"
    "path/filepath"
    "strings"
)

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

### 4. Application Services

#### Document Generation Service
```go
// internal/document/app/generate_document.go
package app

import (
    "context"
    "fmt"
    "github.com/BuddhiLW/AutoPDF/internal/document/domain/entities"
    "github.com/BuddhiLW/AutoPDF/pkg/ports"
)

type GenerateDocumentService struct {
    documentRepo    ports.DocumentRepository
    templateProc    ports.TemplateProcessor
    latexCompiler   ports.LaTeXCompiler
    fileSystem      ports.FileSystem
}

func NewGenerateDocumentService(
    documentRepo ports.DocumentRepository,
    templateProc ports.TemplateProcessor,
    latexCompiler ports.LaTeXCompiler,
    fileSystem ports.FileSystem,
) *GenerateDocumentService {
    return &GenerateDocumentService{
        documentRepo:  documentRepo,
        templateProc:  templateProc,
        latexCompiler: latexCompiler,
        fileSystem:    fileSystem,
    }
}

type GenerateDocumentCommand struct {
    TemplatePath string
    OutputPath   string
    Variables    map[string]string
    Engine       string
}

type GenerateDocumentResult struct {
    DocumentID entities.DocumentID
    OutputPath string
    Success    bool
    Error      error
}

func (s *GenerateDocumentService) Execute(ctx context.Context, cmd GenerateDocumentCommand) (GenerateDocumentResult, error) {
    // Create document entity
    docID := entities.DocumentID(fmt.Sprintf("doc_%d", time.Now().UnixNano()))
    
    templatePath, err := value_objects.NewTemplatePath(cmd.TemplatePath)
    if err != nil {
        return GenerateDocumentResult{Error: err}, err
    }
    
    outputPath, err := value_objects.NewOutputPath(cmd.OutputPath)
    if err != nil {
        return GenerateDocumentResult{Error: err}, err
    }
    
    variables := value_objects.NewVariables(cmd.Variables)
    engine := value_objects.NewEngineType(cmd.Engine)
    
    document := entities.NewDocument(docID, templatePath, outputPath, variables, engine)
    
    // Save document
    if err := s.documentRepo.Save(ctx, document); err != nil {
        return GenerateDocumentResult{Error: err}, err
    }
    
    // Process template
    document.MarkAsProcessing()
    processedTemplate, err := s.templateProc.Process(ctx, templatePath, variables)
    if err != nil {
        document.MarkAsFailed()
        s.documentRepo.Save(ctx, document)
        return GenerateDocumentResult{Error: err}, err
    }
    
    // Compile LaTeX
    compilationOptions := value_objects.NewCompilationOptions(engine, outputPath)
    result, err := s.latexCompiler.Compile(ctx, processedTemplate, compilationOptions)
    if err != nil {
        document.MarkAsFailed()
        s.documentRepo.Save(ctx, document)
        return GenerateDocumentResult{Error: err}, err
    }
    
    // Mark as completed
    document.MarkAsCompleted()
    s.documentRepo.Save(ctx, document)
    
    return GenerateDocumentResult{
        DocumentID: docID,
        OutputPath: result.OutputPath(),
        Success:    true,
    }, nil
}
```

### 5. Infrastructure Adapters

#### LaTeX Compiler Adapter
```go
// internal/document/infra/latex_compiler.go
package infra

import (
    "context"
    "fmt"
    "os/exec"
    "path/filepath"
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
    // Create output directory if it doesn't exist
    outputDir := filepath.Dir(options.OutputPath())
    if err := os.MkdirAll(outputDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create output directory: %w", err)
    }
    
    // Write processed template to temporary file
    tempFile := filepath.Join(outputDir, "temp_"+filepath.Base(template.Path()))
    if err := l.fileSystem.WriteFile(ctx, tempFile, template.Content()); err != nil {
        return nil, fmt.Errorf("failed to write template: %w", err)
    }
    
    // Execute LaTeX compilation
    cmd := exec.CommandContext(ctx, options.Engine(), 
        "-interaction=nonstopmode",
        "-jobname="+filepath.Base(options.OutputPath()),
        "-output-directory="+outputDir,
        tempFile,
    )
    
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("latex compilation failed: %w", err)
    }
    
    // Return compilation result
    return NewCompilationResult(options.OutputPath(), true), nil
}
```

### 6. Legacy Adapter (Strangler Fig)

#### Legacy Document Service Adapter
```go
// internal/legacy/document_adapter.go
package legacy

import (
    "context"
    "github.com/BuddhiLW/AutoPDF/internal/tex"
    "github.com/BuddhiLW/AutoPDF/internal/template"
    "github.com/BuddhiLW/AutoPDF/pkg/ports"
)

type LegacyDocumentAdapter struct {
    // Keep references to current implementations
    buildCmd *tex.BuildCmd
}

func NewLegacyDocumentAdapter() *LegacyDocumentAdapter {
    return &LegacyDocumentAdapter{
        buildCmd: tex.BuildCmd,
    }
}

func (l *LegacyDocumentAdapter) GenerateDocument(ctx context.Context, cmd ports.GenerateDocumentCommand) (ports.GenerateDocumentResult, error) {
    // Use current implementation as fallback
    // This ensures backward compatibility during migration
    
    // Convert new command to legacy format
    args := []string{cmd.TemplatePath, cmd.OutputPath}
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

### 7. Feature Flag Implementation

#### Feature Flag Service
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

### 8. Characterization Tests

#### Golden Tests for Current Behavior
```go
// test/characterization/document_generation_test.go
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

## Implementation Steps

### Step 1: Create Directory Structure
```bash
mkdir -p internal/document/{domain/{entities,value_objects,services,repositories},app,infra}
mkdir -p internal/conversion/{domain/{entities,value_objects,services},app,infra}
mkdir -p internal/configuration/{domain/{entities,value_objects},app,infra}
mkdir -p internal/legacy
mkdir -p pkg/ports
mkdir -p pkg/features
mkdir -p test/characterization
```

### Step 2: Implement Core Interfaces
- Define all port interfaces in `pkg/ports/`
- Create basic domain entities and value objects
- Implement infrastructure adapters

### Step 3: Create Legacy Adapters
- Wrap current functionality in adapter pattern
- Ensure backward compatibility
- Add feature flags for gradual migration

### Step 4: Add Characterization Tests
- Capture current behavior with golden tests
- Ensure refactoring doesn't break functionality
- Add performance benchmarks

### Step 5: Implement Application Services
- Create use case orchestration
- Wire up dependencies
- Add error handling and logging

## Success Criteria

1. **Backward Compatibility**: All existing functionality works unchanged
2. **Test Coverage**: Characterization tests capture current behavior
3. **Seam Establishment**: Clear boundaries between old and new code
4. **Feature Flags**: Ability to toggle between implementations
5. **Documentation**: Clear interfaces and contracts defined

## Risks and Mitigation

### Risks
1. **Interface Design**: May need to iterate on interface design
2. **Legacy Integration**: Complex integration with current code
3. **Performance**: Additional abstraction layers may impact performance

### Mitigation
1. Start with simple interfaces and evolve
2. Use adapter pattern to minimize changes to legacy code
3. Profile and optimize critical paths
4. Use feature flags to gradually migrate

## Success Criteria

- [ ] Directory structure created
- [ ] Port interfaces implemented
- [ ] Basic domain entities created
- [ ] Legacy adapters implemented
- [ ] Feature flags implemented
- [ ] Characterization tests added
- [ ] One use case migrated to new architecture

## Next Steps

1. Implement directory structure
2. Create port interfaces
3. Implement basic domain entities
4. Create legacy adapters
5. Add characterization tests
6. Implement feature flags
7. Begin gradual migration of document generation
