# ADR-007: CLI Contract Compliance

## Status
Accepted

## Context
AutoPDF uses Bonzai Tree CLI which follows a "no flags" philosophy - commands are positional arguments only. The DDD refactoring must maintain this contract and ensure feature toggles use environment variables or configuration.

## Decision

**Maintain Bonzai CLI Contract**

1. **Positional Arguments Only**: No flags in command signatures
2. **Environment Variables**: Feature toggles via env vars
3. **Configuration**: Feature settings in YAML config
4. **Public Contract Tests**: Test CLI behavior explicitly

### Rationale
- **Bonzai Philosophy**: Follows the framework's design principles
- **User Experience**: Consistent with existing CLI usage
- **Backward Compatibility**: No breaking changes to CLI interface
- **Clear Contract**: Explicit testing of CLI behavior

## Implementation

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

### Application Service Entry Point
```go
// internal/document/app/generate_document_service.go
package app

type GenerateDocumentService struct {
    legacyAdapter  LegacyDocumentAdapter
    newService     NewDocumentService
    featureFlags   *features.FeatureFlags
}

func (s *GenerateDocumentService) Execute(ctx context.Context, cmd GenerateDocumentCommand) (GenerateDocumentResult, error) {
    // Route based on feature flags
    if s.featureFlags.UseNewDocumentGeneration {
        return s.newService.Execute(ctx, cmd)
    }
    
    // Fall back to legacy
    return s.legacyAdapter.Execute(ctx, cmd)
}
```

### Public Contract Tests
```go
// test/contract/cli_contract_test.go
package contract

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

## Benefits

1. **Bonzai Compliance**: Follows framework philosophy
2. **User Experience**: Consistent CLI interface
3. **Backward Compatibility**: No breaking changes
4. **Clear Contract**: Explicit testing of CLI behavior
5. **Feature Toggles**: Environment-based feature control

## Success Criteria

- [ ] All commands use positional arguments only
- [ ] No flags in command signatures
- [ ] Feature toggles via environment variables
- [ ] Public contract tests pass
- [ ] CLI behavior is documented and tested

## Consequences

- **Limited CLI Options**: No command-line flags for features
- **Environment Dependencies**: Feature toggles require env vars
- **Testing Complexity**: Need to test CLI behavior explicitly

## Mitigation

- **Clear Documentation**: Document environment variables
- **Default Behavior**: Sensible defaults for all features
- **Configuration**: YAML config for complex settings
- **Public Tests**: Comprehensive CLI contract testing
