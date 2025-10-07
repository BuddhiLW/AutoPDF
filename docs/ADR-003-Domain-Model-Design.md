# ADR-003: Domain Model Design for AutoPDF

## Status
Proposed

## Context
Based on the analysis of AutoPDF's current functionality, we need to design a rich domain model that captures the business logic and rules around document generation, template processing, and LaTeX compilation.

## Current Domain Analysis

### Core Business Concepts

1. **Document Generation**: The primary business capability
   - Takes a template and variables
   - Processes the template with variables
   - Compiles to PDF using LaTeX
   - Optionally converts to images

2. **Template Processing**: Core business logic
   - Variable substitution
   - Custom delimiters to avoid LaTeX conflicts
   - Template validation

3. **LaTeX Compilation**: Technical requirement
   - Multiple engine support (pdflatex, xelatex, lualatex)
   - Error handling and recovery
   - Output path management

4. **Document Conversion**: Optional enhancement
   - PDF to image conversion
   - Multiple format support
   - Tool integration

## Decision

We will design a domain model that captures these concepts using DDD tactical patterns:

### 1. Aggregates and Entities

#### Document Aggregate
```go
// internal/document/domain/entities/document.go
package entities

import (
    "time"
    "github.com/BuddhiLW/AutoPDF/internal/document/domain/value_objects"
)

// Document is the root aggregate for document generation
type Document struct {
    id          DocumentID
    template    value_objects.TemplatePath
    output      value_objects.OutputPath
    variables   value_objects.Variables
    engine      value_objects.EngineType
    status      DocumentStatus
    compilation *CompilationResult
    createdAt   time.Time
    updatedAt   time.Time
}

// DocumentID is the aggregate root identifier
type DocumentID string

// DocumentStatus represents the current state of document processing
type DocumentStatus string

const (
    DocumentStatusPending    DocumentStatus = "pending"
    DocumentStatusProcessing DocumentStatus = "processing"
    DocumentStatusCompleted  DocumentStatus = "completed"
    DocumentStatusFailed     DocumentStatus = "failed"
)

// Business methods
func (d *Document) StartProcessing() error {
    if d.status != DocumentStatusPending {
        return ErrDocumentNotPending
    }
    d.status = DocumentStatusProcessing
    d.updatedAt = time.Now()
    return nil
}

func (d *Document) CompleteProcessing(result CompilationResult) error {
    if d.status != DocumentStatusProcessing {
        return ErrDocumentNotProcessing
    }
    d.status = DocumentStatusCompleted
    d.compilation = &result
    d.updatedAt = time.Now()
    return nil
}

func (d *Document) FailProcessing(reason string) error {
    if d.status != DocumentStatusProcessing {
        return ErrDocumentNotProcessing
    }
    d.status = DocumentStatusFailed
    d.updatedAt = time.Now()
    return nil
}

func (d *Document) CanBeProcessed() bool {
    return d.status == DocumentStatusPending
}

func (d *Document) IsCompleted() bool {
    return d.status == DocumentStatusCompleted
}

func (d *Document) IsFailed() bool {
    return d.status == DocumentStatusFailed
}
```

#### Compilation Result Entity
```go
// internal/document/domain/entities/compilation_result.go
package entities

import (
    "time"
    "github.com/BuddhiLW/AutoPDF/internal/document/domain/value_objects"
)

// CompilationResult represents the outcome of LaTeX compilation
type CompilationResult struct {
    id          CompilationResultID
    documentID  DocumentID
    outputPath  value_objects.OutputPath
    success     bool
    errors      []CompilationError
    warnings    []CompilationWarning
    duration    time.Duration
    createdAt   time.Time
}

type CompilationResultID string

type CompilationError struct {
    Line    int
    Column  int
    Message string
    Code    string
}

type CompilationWarning struct {
    Line    int
    Column  int
    Message string
    Code    string
}

func NewCompilationResult(
    id CompilationResultID,
    documentID DocumentID,
    outputPath value_objects.OutputPath,
    success bool,
    duration time.Duration,
) *CompilationResult {
    return &CompilationResult{
        id:         id,
        documentID: documentID,
        outputPath: outputPath,
        success:    success,
        duration:   duration,
        createdAt:  time.Now(),
    }
}

func (cr *CompilationResult) AddError(err CompilationError) {
    cr.errors = append(cr.errors, err)
}

func (cr *CompilationResult) AddWarning(warn CompilationWarning) {
    cr.warnings = append(cr.warnings, warn)
}

func (cr *CompilationResult) HasErrors() bool {
    return len(cr.errors) > 0
}

func (cr *CompilationResult) HasWarnings() bool {
    return len(cr.warnings) > 0
}

func (cr *CompilationResult) IsSuccessful() bool {
    return cr.success && !cr.HasErrors()
}
```

### 2. Value Objects

#### Template Path Value Object
```go
// internal/document/domain/value_objects/template_path.go
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
    
    // Validate that the path is not a directory
    if strings.HasSuffix(path, "/") || strings.HasSuffix(path, "\\") {
        return TemplatePath{}, errors.New("template path cannot be a directory")
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

func (tp TemplatePath) Ext() string {
    return filepath.Ext(tp.value)
}

func (tp TemplatePath) Equals(other TemplatePath) bool {
    return tp.value == other.value
}

func (tp TemplatePath) IsValid() bool {
    return tp.value != ""
}
```

#### Variables Value Object
```go
// internal/document/domain/value_objects/variables.go
package value_objects

import (
    "errors"
    "strings"
)

// Variables represents a collection of template variables
type Variables struct {
    values map[string]string
}

func NewVariables(vars map[string]string) Variables {
    if vars == nil {
        vars = make(map[string]string)
    }
    return Variables{values: vars}
}

func (v Variables) Get(key string) (string, bool) {
    value, exists := v.values[key]
    return value, exists
}

func (v Variables) Set(key, value string) error {
    if strings.TrimSpace(key) == "" {
        return errors.New("variable key cannot be empty")
    }
    v.values[key] = value
    return nil
}

func (v Variables) Keys() []string {
    keys := make([]string, 0, len(v.values))
    for k := range v.values {
        keys = append(keys, k)
    }
    return keys
}

func (v Variables) Count() int {
    return len(v.values)
}

func (v Variables) IsEmpty() bool {
    return len(v.values) == 0
}

func (v Variables) ToMap() map[string]string {
    result := make(map[string]string)
    for k, v := range v.values {
        result[k] = v
    }
    return result
}
```

#### Engine Type Value Object
```go
// internal/document/domain/value_objects/engine_type.go
package value_objects

import (
    "errors"
    "strings"
)

// EngineType represents a LaTeX engine with validation
type EngineType struct {
    value string
}

var (
    EngineTypePdfLatex = EngineType{value: "pdflatex"}
    EngineTypeXeLatex  = EngineType{value: "xelatex"}
    EngineTypeLuaLatex = EngineType{value: "lualatex"}
)

var supportedEngines = map[string]EngineType{
    "pdflatex": EngineTypePdfLatex,
    "xelatex":  EngineTypeXeLatex,
    "lualatex": EngineTypeLuaLatex,
}

func NewEngineType(engine string) (EngineType, error) {
    if strings.TrimSpace(engine) == "" {
        return EngineTypePdfLatex, nil // Default to pdflatex
    }
    
    engine = strings.ToLower(strings.TrimSpace(engine))
    if et, exists := supportedEngines[engine]; exists {
        return et, nil
    }
    
    return EngineType{}, errors.New("unsupported LaTeX engine: " + engine)
}

func (et EngineType) String() string {
    return et.value
}

func (et EngineType) Equals(other EngineType) bool {
    return et.value == other.value
}

func (et EngineType) IsPdfLatex() bool {
    return et.value == "pdflatex"
}

func (et EngineType) IsXeLatex() bool {
    return et.value == "xelatex"
}

func (et EngineType) IsLuaLatex() bool {
    return et.value == "lualatex"
}

func (et EngineType) RequiresUnicode() bool {
    return et.IsXeLatex() || et.IsLuaLatex()
}
```

### 3. Domain Services

#### Template Processing Service
```go
// internal/document/domain/services/template_processor.go
package services

import (
    "context"
    "errors"
    "github.com/BuddhiLW/AutoPDF/internal/document/domain/value_objects"
)

// TemplateProcessor handles template processing business logic
type TemplateProcessor struct {
    // Dependencies injected via constructor
}

func NewTemplateProcessor() *TemplateProcessor {
    return &TemplateProcessor{}
}

// ProcessTemplate processes a template with variables
func (tp *TemplateProcessor) ProcessTemplate(
    ctx context.Context,
    templatePath value_objects.TemplatePath,
    variables value_objects.Variables,
) (value_objects.ProcessedTemplate, error) {
    
    // Validate template path
    if !templatePath.IsValid() {
        return value_objects.ProcessedTemplate{}, errors.New("invalid template path")
    }
    
    // Validate variables
    if variables.IsEmpty() {
        return value_objects.ProcessedTemplate{}, errors.New("variables cannot be empty")
    }
    
    // Business rule: Template must contain at least one variable placeholder
    // This is a domain rule that belongs in the domain service
    
    // Process template with variables
    processedContent, err := tp.processWithVariables(templatePath, variables)
    if err != nil {
        return value_objects.ProcessedTemplate{}, err
    }
    
    return value_objects.NewProcessedTemplate(templatePath, processedContent), nil
}

func (tp *TemplateProcessor) processWithVariables(
    templatePath value_objects.TemplatePath,
    variables value_objects.Variables,
) (string, error) {
    // Implementation details for template processing
    // This would use Go's text/template package with custom delimiters
    // The business logic is: use custom delimiters to avoid LaTeX conflicts
    
    // For now, return a placeholder
    return "processed template content", nil
}
```

### 4. Domain Events

#### Document Events
```go
// internal/document/domain/events/document_events.go
package events

import (
    "time"
    "github.com/BuddhiLW/AutoPDF/internal/document/domain/entities"
)

// DocumentEvent represents a domain event
type DocumentEvent interface {
    EventID() string
    AggregateID() string
    OccurredAt() time.Time
    EventType() string
}

// DocumentProcessingStarted is raised when document processing begins
type DocumentProcessingStarted struct {
    eventID     string
    documentID  entities.DocumentID
    occurredAt  time.Time
}

func NewDocumentProcessingStarted(documentID entities.DocumentID) *DocumentProcessingStarted {
    return &DocumentProcessingStarted{
        eventID:    generateEventID(),
        documentID: documentID,
        occurredAt: time.Now(),
    }
}

func (e *DocumentProcessingStarted) EventID() string {
    return e.eventID
}

func (e *DocumentProcessingStarted) AggregateID() string {
    return string(e.documentID)
}

func (e *DocumentProcessingStarted) OccurredAt() time.Time {
    return e.occurredAt
}

func (e *DocumentProcessingStarted) EventType() string {
    return "DocumentProcessingStarted"
}

// DocumentProcessingCompleted is raised when document processing completes successfully
type DocumentProcessingCompleted struct {
    eventID     string
    documentID  entities.DocumentID
    outputPath  string
    occurredAt  time.Time
}

func NewDocumentProcessingCompleted(documentID entities.DocumentID, outputPath string) *DocumentProcessingCompleted {
    return &DocumentProcessingCompleted{
        eventID:    generateEventID(),
        documentID: documentID,
        outputPath: outputPath,
        occurredAt: time.Now(),
    }
}

func (e *DocumentProcessingCompleted) EventID() string {
    return e.eventID
}

func (e *DocumentProcessingCompleted) AggregateID() string {
    return string(e.documentID)
}

func (e *DocumentProcessingCompleted) OccurredAt() time.Time {
    return e.occurredAt
}

func (e *DocumentProcessingCompleted) EventType() string {
    return "DocumentProcessingCompleted"
}

func (e *DocumentProcessingCompleted) OutputPath() string {
    return e.outputPath
}

func generateEventID() string {
    return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}
```

### 5. Domain Errors

#### Domain-Specific Errors
```go
// internal/document/domain/errors/domain_errors.go
package errors

import "errors"

// Domain-specific errors
var (
    ErrDocumentNotPending     = errors.New("document is not in pending status")
    ErrDocumentNotProcessing  = errors.New("document is not in processing status")
    ErrInvalidTemplatePath    = errors.New("invalid template path")
    ErrInvalidOutputPath      = errors.New("invalid output path")
    ErrInvalidVariables       = errors.New("invalid variables")
    ErrUnsupportedEngine      = errors.New("unsupported LaTeX engine")
    ErrTemplateProcessingFailed = errors.New("template processing failed")
    ErrLaTeXCompilationFailed   = errors.New("LaTeX compilation failed")
    ErrDocumentNotFound         = errors.New("document not found")
    ErrDocumentAlreadyExists    = errors.New("document already exists")
)

// CompilationError represents a LaTeX compilation error
type CompilationError struct {
    Line    int
    Column  int
    Message string
    Code    string
}

func (ce CompilationError) Error() string {
    return ce.Message
}

// TemplateError represents a template processing error
type TemplateError struct {
    Line    int
    Column  int
    Message string
    Code    string
}

func (te TemplateError) Error() string {
    return te.Message
}
```

## Benefits of This Design

### 1. Rich Domain Model
- **Entities** capture business concepts with behavior
- **Value Objects** ensure type safety and validation
- **Domain Services** handle complex business logic
- **Domain Events** enable loose coupling and audit trails

### 2. Business Rules Encapsulation
- Document status transitions are controlled
- Template validation is centralized
- Engine type validation prevents errors
- Variable validation ensures data quality

### 3. Extensibility
- Easy to add new LaTeX engines
- Simple to add new document types
- Straightforward to add new processing steps
- Clear extension points for new features

### 4. Testability
- Pure domain logic is easy to test
- Value objects can be tested in isolation
- Domain services can be mocked
- Business rules are explicit and testable

## Implementation Strategy

### Phase 1: Core Domain
1. Implement basic entities and value objects
2. Add domain services for core business logic
3. Create domain events for important state changes
4. Add comprehensive unit tests

### Phase 2: Business Rules
1. Implement validation rules
2. Add business rule enforcement
3. Create domain error handling
4. Add business rule tests

### Phase 3: Integration
1. Connect domain to application services
2. Implement domain event handling
3. Add integration tests
4. Performance optimization

## Success Criteria

1. **Business Logic Encapsulation**: All business rules are in the domain layer
2. **Type Safety**: Value objects prevent invalid data
3. **Testability**: Domain logic is easily testable
4. **Extensibility**: New features can be added without breaking existing code
5. **Performance**: Domain operations are efficient
6. **Documentation**: Domain model is self-documenting

## Next Steps

1. Implement core entities and value objects
2. Add domain services for business logic
3. Create domain events for state changes
4. Add comprehensive unit tests
5. Integrate with application services
