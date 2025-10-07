# ADR-004: Type Boundaries and Ownership

## Status
Accepted

## Context
The initial DDD refactoring plan had inconsistent type ownership between `pkg/ports` and domain value objects, creating potential coupling issues and import cycles. We need to establish clear boundaries for type ownership to prevent friction during implementation.

## Decision

**Option A: Pure Transport Types in Ports**

We will use pure transport types (strings, bytes, primitives) in port interfaces, keeping domain value objects internal to their bounded contexts. Adapters will handle translation between transport and domain types.

### Rationale
- **Prevents Import Cycles**: Ports don't depend on domain types
- **Clear Boundaries**: Domain owns its types, ports are pure contracts
- **Adapter Responsibility**: Translation logic belongs in adapters
- **Testability**: Ports can be tested with simple types
- **Flexibility**: Different bounded contexts can have different domain representations

## Implementation

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
    // Dependencies
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

## Benefits

1. **No Import Cycles**: Ports are independent of domain types
2. **Clear Ownership**: Domain owns its types, ports are pure contracts
3. **Adapter Responsibility**: Translation logic is explicit and testable
4. **Flexibility**: Different contexts can have different domain representations
5. **Testability**: Ports can be tested with simple types

## Success Criteria

- [ ] Port interfaces use only primitive/transport types
- [ ] Domain value objects are internal to bounded contexts
- [ ] Adapters handle all translation between transport and domain types
- [ ] No import cycles between ports and domain
- [ ] All type conversions are explicit and testable

## Consequences

- **Translation Overhead**: Adapters need to convert between transport and domain types
- **Type Safety**: Less compile-time type safety at port boundaries
- **Explicit Conversion**: All type conversions must be explicit in adapters

## Mitigation

- **Adapter Tests**: Comprehensive tests for type conversion logic
- **Domain Validation**: Strong validation in domain value objects
- **Clear Documentation**: Document all type conversions and their purposes
