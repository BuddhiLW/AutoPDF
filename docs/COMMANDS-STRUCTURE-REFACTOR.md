# Commands Structure Refactor

## Overview

Refactored the `internal/autopdf/commands/` folder to follow **Single Responsibility Principle** and improve maintainability by organizing code into focused, reusable abstractions.

## New Structure

```
internal/autopdf/commands/
├── shared/                    # Shared utilities and abstractions
│   ├── args_parser.go         # Argument parsing logic
│   ├── args_parser_test.go
│   ├── config_resolver.go     # Config file and template path resolution
│   ├── config_resolver_test.go
│   ├── service_builder.go     # Application service construction
│   ├── service_builder_test.go
│   ├── result_handler.go      # Result output handling
│   └── result_handler_test.go
├── build/                     # Build command specific
│   └── build_service.go       # BuildServiceCmd
└── (future commands)          # Clean, convert, etc.
    ├── clean/
    └── convert/
```

## Benefits

### 1. **Single Responsibility Principle (SRP)**
- **Shared utilities**: Each has one clear responsibility
- **Command-specific**: Each command folder contains only its logic
- **Clear separation**: Shared vs. command-specific concerns

### 2. **Open/Closed Principle (OCP)**
- **Easy to extend**: New commands can reuse shared utilities
- **No modification**: Adding new commands doesn't require changing existing code
- **Pluggable**: Commands can be added/removed independently

### 3. **Dependency Inversion Principle (DIP)**
- **Shared abstractions**: Commands depend on shared interfaces
- **Loose coupling**: Commands don't depend on each other
- **Testable**: Each component can be tested in isolation

### 4. **Interface Segregation Principle (ISP)**
- **Focused interfaces**: Each shared utility has a specific purpose
- **No fat interfaces**: Commands only use what they need
- **Clear contracts**: Well-defined responsibilities

## Shared Utilities

### `ArgsParser`
- **Responsibility**: Parse command-line arguments
- **Reusable**: All commands can use this for argument parsing
- **Testable**: Unit tests for different argument combinations

### `ConfigResolver`
- **Responsibility**: Resolve config files and template paths
- **Reusable**: All commands that need config can use this
- **Smart**: Handles relative paths, absolute paths, and defaults

### `ServiceBuilder`
- **Responsibility**: Construct application services
- **Reusable**: All commands can build their required services
- **Flexible**: Can be extended for different service configurations

### `ResultHandler`
- **Responsibility**: Handle and display command results
- **Reusable**: All commands can use this for consistent output
- **Extensible**: Easy to add new output formats

## Command-Specific Folders

### `build/`
- **Contains**: `BuildServiceCmd` - the actual Bonzai command
- **Focus**: Build-specific logic and integration
- **Dependencies**: Uses shared utilities for common functionality

### Future Commands
- **`clean/`**: Clean command with its specific logic
- **`convert/`**: Convert command with its specific logic
- **`compile/`**: Compile command with its specific logic

## SOLID Principles Applied

### Single Responsibility Principle (SRP)
- Each shared utility has one clear responsibility
- Each command folder contains only command-specific logic
- Clear separation of concerns

### Open/Closed Principle (OCP)
- Easy to add new commands without modifying existing code
- Shared utilities can be extended without breaking existing commands
- New functionality can be added through composition

### Liskov Substitution Principle (LSP)
- All shared utilities implement their interfaces correctly
- Commands can be substituted without breaking functionality
- Consistent behavior across all commands

### Interface Segregation Principle (ISP)
- Small, focused interfaces for each shared utility
- Commands only depend on what they need
- No unnecessary dependencies

### Dependency Inversion Principle (DIP)
- Commands depend on shared abstractions, not concrete implementations
- Shared utilities can be easily mocked for testing
- Loose coupling between components

## Testing Strategy

### Unit Tests
- **Shared utilities**: Comprehensive unit tests for each utility
- **Command integration**: Integration tests for each command
- **Isolation**: Each component can be tested independently

### Test Coverage
- **ArgsParser**: Tests for different argument combinations
- **ConfigResolver**: Tests for path resolution scenarios
- **ServiceBuilder**: Tests for service construction
- **ResultHandler**: Tests for result processing

## Migration Benefits

### Before (Monolithic)
```go
// All logic in one file
func (cmd *bonzai.Cmd) Do(args ...string) error {
    // 100+ lines of mixed concerns
    // Argument parsing
    // Config resolution
    // Service building
    // Result handling
}
```

### After (Focused Abstractions)
```go
// Clean, focused command
func (cmd *bonzai.Cmd) Do(args ...string) error {
    // Parse arguments
    argsParser := shared.NewArgsParser()
    buildArgs, err := argsParser.ParseBuildArgs(args)
    
    // Resolve config
    configResolver := shared.NewConfigResolver()
    configFile, err := configResolver.ResolveConfigFile(...)
    
    // Build service
    serviceBuilder := shared.NewServiceBuilder()
    svc := serviceBuilder.BuildDocumentService(cfg)
    
    // Handle result
    resultHandler := shared.NewResultHandler()
    return resultHandler.HandleBuildResult(result)
}
```

## Future Extensibility

### Adding New Commands
1. Create new command folder (e.g., `clean/`)
2. Import shared utilities
3. Implement command-specific logic
4. Register in main command tree

### Adding New Shared Utilities
1. Add to `shared/` folder
2. Create focused interface
3. Implement with tests
4. Use in commands as needed

### Example: Clean Command
```go
// internal/autopdf/commands/clean/clean_service.go
package clean

import (
    "github.com/BuddhiLW/AutoPDF/internal/autopdf/commands/shared"
    // ... other imports
)

var CleanServiceCmd = &bonzai.Cmd{
    Name: "clean",
    Do: func(cmd *bonzai.Cmd, args ...string) error {
        // Use shared utilities
        argsParser := shared.NewArgsParser()
        configResolver := shared.NewConfigResolver()
        // ... implement clean-specific logic
    },
}
```

## Status: ✅ COMPLETE

The commands structure has been successfully refactored with:
- **Focused abstractions** in `shared/` folder
- **Command-specific logic** in dedicated folders
- **SOLID principles** applied throughout
- **Comprehensive tests** for all components
- **Future-ready** structure for new commands

The refactored structure is more maintainable, testable, and extensible while preserving all existing functionality.
