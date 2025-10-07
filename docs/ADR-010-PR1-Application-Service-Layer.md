// ADR-010: PR#1 - Application Service Layer Introduction

## Status
Accepted

## Context
The current Bonzai CLI commands (`internal/tex/build.go`, `internal/autopdf/cmd.go`) contain orchestration logic mixed with CLI concerns. To enable DDD refactoring, we need to separate the CLI layer from business logic orchestration.

## Decision

**Introduce Application Service Layer as Seam**

We will create an `application.DocumentService` that orchestrates the document generation workflow, while keeping the Bonzai CLI layer thin and focused only on argument parsing and output formatting.

### Key Components

1. **Application Service** (`internal/application/document_service.go`)
   - Orchestrates render → compile → convert → clean workflow
   - Depends on port interfaces (Dependency Inversion)
   - Pure business logic orchestration

2. **Port Interfaces** (`internal/application/ports.go`)
   - Pure transport types (strings, maps)
   - No domain dependencies
   - Clean contracts for adapters

3. **Adapters** (`internal/application/adapters/`)
   - Wrap existing packages (`template`, `tex`, `converter`)
   - Implement port interfaces
   - Handle type translation

4. **Thin CLI Layer** (`internal/autopdf/commands/build_service.go`)
   - Parse positional arguments
   - Wire up service with adapters
   - Format output
   - No business logic

## Implementation

### Application Service
```go
type DocumentService struct {
    TemplateProcessor TemplateProcessor
    LaTeXCompiler     LaTeXCompiler
    Converter         Converter
    Cleaner           Cleaner
}

func (s *DocumentService) Build(ctx context.Context, req BuildRequest) (BuildResult, error) {
    // Orchestrate: template → compile → convert → clean
}
```

### Port Interfaces (Pure Transport Types)
```go
type TemplateProcessor interface {
    Process(ctx context.Context, templatePath string, variables map[string]string) (string, error)
}

type LaTeXCompiler interface {
    Compile(ctx context.Context, content string, engine string, outputPath string) (string, error)
}
```

### Adapters (Wrap Existing Code)
```go
type TemplateProcessorAdapter struct {
    config *config.Config
}

func (tpa *TemplateProcessorAdapter) Process(ctx context.Context, templatePath string, variables map[string]string) (string, error) {
    // Call existing template.Engine
}
```

### Thin CLI Layer
```go
var BuildServiceCmd = &bonzai.Cmd{
    Do: func(cmd *bonzai.Cmd, args ...string) error {
        // Parse args
        templateFile, configFile, doClean := parseArgs(args)
        
        // Wire up service
        svc := application.DocumentService{
            TemplateProcessor: adapters.NewTemplateProcessorAdapter(cfg),
            LaTeXCompiler:     adapters.NewLaTeXCompilerAdapter(cfg),
            Converter:         adapters.NewConverterAdapter(cfg),
            Cleaner:           adapters.NewCleanerAdapter(),
        }
        
        // Execute
        result, err := svc.Build(ctx, req)
        
        // Format output
        fmt.Printf("Successfully built PDF: %s\n", result.PDFPath)
    },
}
```

## Benefits

1. **Strangler Seam**: Can refactor internals without touching CLI
2. **Testability**: Application service can be unit tested with mocks
3. **DDD Shape**: Domain/application/infrastructure separation introduced
4. **SOLID**: Dependency Inversion Principle applied throughout
5. **Zero Behavior Change**: User-visible functionality remains identical
6. **Reversible**: Can rollback easily if needed

## Public Contract Tests

Added `test/contract/cli_public_contract_test.go` to lock down CLI behavior:

```go
func TestCLIPublicContract_Build(t *testing.T) {
    // Ensures positional args work
    // Verifies exit codes
    // Checks output file exists
    // Validates stdout format
}

func TestCLIPublicContract_NoFlags(t *testing.T) {
    // Ensures flags are rejected (Bonzai philosophy)
}

func TestCLIPublicContract_PositionalArgs(t *testing.T) {
    // Tests all valid argument combinations
}
```

## Migration Strategy

### Phase 1: Coexistence (This PR)
- New `BuildServiceCmd` alongside existing `BuildCmd`
- Both commands work identically
- Public contract tests ensure compatibility
- Feature flag via environment variable

### Phase 2: Gradual Adoption
- Update documentation to reference new command
- Monitor usage and feedback
- Fix any issues found

### Phase 3: Deprecation
- Mark old command as deprecated
- Remove old command after migration period

## Success Criteria

- [ ] Application service layer implemented
- [ ] Port interfaces defined with pure transport types
- [ ] Adapters wrap existing packages
- [ ] Thin CLI layer delegates to service
- [ ] Public contract tests pass
- [ ] Unit tests for application service
- [ ] Zero behavior change for users
- [ ] Documentation updated

## File Structure

```
internal/
├── application/
│   ├── document_service.go       # Service orchestration
│   ├── document_service_test.go  # Unit tests
│   ├── ports.go                  # Port interfaces
│   └── adapters/
│       ├── template_adapter.go   # Wraps internal/template
│       ├── latex_adapter.go      # Wraps internal/tex
│       ├── converter_adapter.go  # Wraps internal/converter
│       └── cleaner_adapter.go    # Wraps internal/tex/clean
├── autopdf/
│   └── commands/
│       └── build_service.go      # Thin CLI layer
test/
└── contract/
    └── cli_public_contract_test.go  # Public contract tests
```

## Risks and Mitigation

### Risks
1. **Additional Abstraction Overhead**: Extra layer may impact performance
2. **Learning Curve**: Team needs to understand new structure
3. **Testing Complexity**: Need to maintain both integration and unit tests

### Mitigation
1. **Performance**: Profile and optimize if needed (minimal overhead expected)
2. **Documentation**: Clear documentation and examples
3. **Testing Strategy**: Clear separation between unit, integration, and contract tests

## Next Steps

1. Run tests to verify no behavior change
2. Add feature flag to toggle between old and new commands
3. Update documentation with migration guide
4. Monitor for issues
5. Prepare PR#2: Domain model extraction

## Related ADRs

- ADR-001: DDD Refactoring Plan
- ADR-002: Phase 1 Implementation
- ADR-004: Type Boundaries and Ownership
- ADR-007: CLI Contract Compliance
