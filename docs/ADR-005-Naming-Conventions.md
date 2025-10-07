# ADR-005: Naming Conventions and Package Structure

## Status
Accepted

## Context
The DDD refactoring plan had inconsistent naming conventions (value_objects vs valueobjects) and unclear package structure. We need to establish consistent naming conventions and package organization.

## Decision

### Package Naming Convention
- **Single word packages**: Use `valueobjects`, `entities`, `services`, `repositories`
- **No underscores**: Follow Go conventions for package names
- **Descriptive names**: Clear purpose and scope

### Directory Structure
```
internal/
├── document/                    # Document Generation Bounded Context
│   ├── domain/                  # Pure domain logic
│   │   ├── entities/           # Domain entities
│   │   ├── valueobjects/       # Value objects (single word)
│   │   ├── services/           # Domain services
│   │   ├── repositories/       # Repository interfaces
│   │   ├── events/            # Domain events
│   │   └── errors/            # Domain errors
│   ├── app/                   # Application services
│   └── infra/                 # Infrastructure adapters
```

### Naming Conventions

#### Domain Layer
- **Entities**: `Document`, `CompilationResult`
- **Value Objects**: `TemplatePath`, `Variables`, `EngineType`
- **Services**: `TemplateProcessor`, `DocumentValidator`
- **Repositories**: `DocumentRepository`, `TemplateRepository`
- **Events**: `DocumentProcessingStarted`, `DocumentProcessingCompleted`
- **Errors**: `ErrDocumentNotFound`, `ErrInvalidTemplatePath`

#### Application Layer
- **Services**: `GenerateDocumentService`, `CleanDocumentService`
- **Commands**: `GenerateDocumentCommand`, `CleanDocumentCommand`
- **Results**: `GenerateDocumentResult`, `CleanDocumentResult`

#### Infrastructure Layer
- **Adapters**: `LaTeXCompilerAdapter`, `TemplateProcessorAdapter`
- **Repositories**: `MemoryDocumentRepository`, `FileSystemRepository`

## Implementation

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

## Benefits

1. **Consistency**: Uniform naming across all packages
2. **Clarity**: Clear purpose and scope for each package
3. **Go Conventions**: Follows standard Go package naming
4. **Maintainability**: Easy to find and organize code
5. **Team Productivity**: Reduces cognitive load

## Success Criteria

- [ ] All packages use single-word names
- [ ] No underscores in package names
- [ ] Clear separation between domain, app, and infra layers
- [ ] Consistent naming patterns across all bounded contexts
- [ ] All imports follow the established structure

## Consequences

- **Migration Effort**: Need to rename existing packages
- **Import Updates**: All imports need to be updated
- **Documentation**: Need to update all documentation

## Mitigation

- **Gradual Migration**: Rename packages incrementally
- **Import Aliases**: Use aliases during transition
- **Automated Tools**: Use goimports and gofmt for consistency
- **Team Alignment**: Ensure all team members follow conventions
