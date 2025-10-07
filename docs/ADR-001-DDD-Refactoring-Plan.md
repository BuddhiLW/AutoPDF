# ADR-001: DDD Refactoring Plan for AutoPDF

## Status
Accepted

## Context
AutoPDF is currently a monolithic CLI application with tightly coupled components. The codebase follows a procedural approach with mixed concerns, making it difficult to test, extend, and maintain. We need to refactor towards Domain-Driven Design (DDD) principles while maintaining backward compatibility.

## Current Architecture Analysis

### Current Structure
```
AutoPDF/
├── cmd/autopdf/           # CLI entry point (Bonzai Tree)
├── internal/
│   ├── autopdf/           # Main command orchestration
│   ├── template/         # Template processing engine
│   ├── tex/              # LaTeX compilation & cleanup
│   └── converter/        # PDF to image conversion
├── pkg/
│   ├── config/           # Configuration management
│   ├── api/              # High-level API functions
│   └── util/             # Utilities
└── configs/              # YAML configuration files
```

### Identified Bounded Contexts

1. **Document Generation** (Core Domain)
   - Template processing
   - Variable substitution
   - LaTeX compilation
   - PDF generation

2. **Document Conversion** (Supporting Domain)
   - PDF to image conversion
   - Format management
   - Tool integration

3. **Configuration Management** (Supporting Domain)
   - YAML parsing
   - Default values
   - Validation

4. **CLI Interface** (Infrastructure)
   - Bonzai Tree commands
   - User interaction
   - File system operations

## Decision

We will refactor AutoPDF using the **Strangler Fig Pattern** to gradually migrate from the current monolithic structure to a DDD-based hexagonal architecture.

### Target Architecture

```
AutoPDF/
├── cmd/autopdf/                    # CLI entry point (unchanged)
├── internal/
│   ├── document/                  # Document Generation Bounded Context
│   │   ├── app/                   # Application services
│   │   │   ├── generate_document.go
│   │   │   └── clean_document.go
│   │   ├── domain/                # Pure domain logic
│   │   │   ├── entities/
│   │   │   │   ├── document.go
│   │   │   │   ├── template.go
│   │   │   │   └── compilation_result.go
│   │   │   ├── value_objects/
│   │   │   │   ├── template_path.go
│   │   │   │   ├── output_path.go
│   │   │   │   └── variables.go
│   │   │   ├── services/
│   │   │   │   └── template_processor.go
│   │   │   └── repositories/
│   │   │       └── document_repository.go
│   │   └── infra/                 # Infrastructure adapters
│   │       ├── latex_compiler.go
│   │       ├── template_engine.go
│   │       └── file_system.go
│   ├── conversion/                # Document Conversion Bounded Context
│   │   ├── app/
│   │   │   └── convert_document.go
│   │   ├── domain/
│   │   │   ├── entities/
│   │   │   │   └── conversion_task.go
│   │   │   ├── value_objects/
│   │   │   │   ├── image_format.go
│   │   │   │   └── conversion_options.go
│   │   │   └── services/
│   │   │       └── format_converter.go
│   │   └── infra/
│   │       ├── imagemagick_adapter.go
│   │       └── pdftoppm_adapter.go
│   ├── configuration/             # Configuration Bounded Context
│   │   ├── app/
│   │   │   └── config_service.go
│   │   ├── domain/
│   │   │   ├── entities/
│   │   │   │   └── configuration.go
│   │   │   └── value_objects/
│   │   │       ├── engine_type.go
│   │   │       └── conversion_settings.go
│   │   └── infra/
│   │       ├── yaml_parser.go
│   │       └── file_loader.go
│   └── legacy/                    # Legacy code (to be gradually removed)
│       ├── autopdf/              # Current implementation
│       ├── template/
│       ├── tex/
│       └── converter/
├── pkg/                          # Shared utilities
│   ├── errors/                   # Domain errors
│   ├── events/                   # Domain events
│   └── ports/                    # Interface definitions
└── docs/                         # Documentation
    ├── ADRs/
    └── architecture/
```

## Implementation Strategy

### Phase 1: Foundation (Days 1-30)
1. **Create hexagonal skeleton**
   - Set up bounded context directories
   - Define port interfaces
   - Create basic domain entities and value objects

2. **Establish seams**
   - Create façade interfaces for current functionality
   - Implement adapter pattern for external dependencies
   - Add feature flags for gradual migration

3. **Characterization tests**
   - Capture current behavior with golden tests
   - Ensure refactoring doesn't break existing functionality

### Phase 2: Domain Extraction (Days 31-60)
1. **Extract core domain logic**
   - Move business rules to domain entities
   - Implement domain services
   - Create value objects for type safety

2. **Implement ports and adapters**
   - Define repository interfaces
   - Create infrastructure adapters
   - Implement anti-corruption layers

3. **Migrate one use case**
   - Start with document generation
   - Implement new path behind feature flag
   - Gradually route traffic to new implementation

### Phase 3: Full Migration (Days 61-90)
1. **Complete bounded context migration**
   - Migrate all use cases to new architecture
   - Remove legacy code
   - Add architectural fitness checks

2. **Enhance domain model**
   - Add domain events
   - Implement complex business rules
   - Add validation and error handling

3. **Optimize and clean up**
   - Remove dead code
   - Improve test coverage
   - Add documentation

## SOLID Principles Application

### Single Responsibility Principle (SRP)
- Each bounded context has a single responsibility
- Domain entities focus on business logic only
- Infrastructure adapters handle external concerns

### Open/Closed Principle (OCP)
- Use Strategy pattern for different LaTeX engines
- Use Factory pattern for document processors
- Extend functionality through new implementations

### Liskov Substitution Principle (LSP)
- All adapters implement the same port interfaces
- Domain services can be substituted with different implementations
- Value objects maintain consistent behavior

### Interface Segregation Principle (ISP)
- Small, focused interfaces for each concern
- Separate interfaces for reading and writing operations
- Specific interfaces for different document types

### Dependency Inversion Principle (DIP)
- Domain depends on abstractions (ports)
- Infrastructure implements these abstractions
- Application services orchestrate without knowing implementation details

## GoF Patterns Implementation

### Strategy Pattern
- Different LaTeX engines (pdflatex, xelatex, lualatex)
- Different conversion tools (ImageMagick, pdftoppm)
- Different template processors

### Factory Pattern
- Document factory for creating different document types
- Compiler factory for different LaTeX engines
- Converter factory for different image formats

### Adapter Pattern
- File system operations
- External tool integration
- Configuration persistence

### Facade Pattern
- Legacy code wrapper during migration
- Simplified API for CLI commands
- Complex domain operations

### Decorator Pattern
- Logging decorators for operations
- Caching decorators for expensive operations
- Validation decorators for input

## Benefits

1. **Maintainability**: Clear separation of concerns
2. **Testability**: Pure domain logic with dependency injection
3. **Extensibility**: Easy to add new features without breaking existing code
4. **Flexibility**: Can swap implementations without changing domain logic
5. **Documentation**: Self-documenting code through domain language

## Risks and Mitigation

### Risks
1. **Over-engineering**: Keep domain model simple and focused
2. **Migration complexity**: Use strangler fig pattern with feature flags
3. **Performance impact**: Profile and optimize critical paths
4. **Team learning curve**: Provide training and documentation

### Mitigation
1. Start with simple domain models and evolve
2. Use comprehensive testing to ensure correctness
3. Monitor performance metrics during migration
4. Provide clear documentation and examples

## Success Criteria

1. All existing functionality preserved
2. Test coverage maintained or improved
3. Performance not degraded
4. Code is more maintainable and extensible
5. Clear domain language throughout the codebase
6. Easy to add new features
7. Team can easily understand and modify the code

## Success Criteria

- [ ] Hexagonal architecture structure created
- [ ] Bounded contexts identified and implemented
- [ ] Port interfaces defined
- [ ] Domain entities and value objects implemented
- [ ] Infrastructure adapters created
- [ ] Legacy adapters implemented
- [ ] Feature flags implemented
- [ ] Characterization tests added
- [ ] Architectural fitness checks in CI

## Next Steps

1. Create detailed implementation plan for Phase 1
2. Set up project structure and basic interfaces
3. Implement characterization tests
4. Begin with document generation bounded context
5. Establish CI/CD pipeline with architectural fitness checks
