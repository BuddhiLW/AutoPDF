# Options DDD Structure

## Overview

Successfully created a comprehensive Domain-Driven Design (DDD) structure for handling build command options, following the same pattern as the `clean` command in the `tex/` package. This creates a clean separation between domain logic, application services, and infrastructure adapters.

## New DDD Structure

```
internal/autopdf/
├── domain/
│   ├── options.go           # Domain model for build options
│   └── options_test.go      # Domain model tests
├── application/
│   ├── options_service.go   # Application service for options execution
│   ├── options_service_test.go
│   └── adapters/
│       ├── cleaner_adapter.go    # Clean option adapter
│       ├── logger_adapter.go     # Verbose option adapter
│       ├── debugger_adapter.go   # Debug option adapter
│       └── forcer_adapter.go     # Force option adapter
└── commands/shared/
    ├── args_parser.go           # Updated to use domain model
    └── args_parser_options_simple_test.go
```

## Domain Model

### BuildOptions (Domain Entity)
```go
type BuildOptions struct {
    Clean   CleanOption
    Verbose VerboseOption
    Debug   DebugOption
    Force   ForceOption
}
```

### Individual Option Types
- **CleanOption**: LaTeX auxiliary file cleaning with target directory
- **VerboseOption**: Logging verbosity with configurable levels
- **DebugOption**: Debug information with output destination
- **ForceOption**: Force operations with overwrite settings

### Domain Methods
- `NewBuildOptions()`: Factory method for default options
- `EnableClean(target)`: Enable clean with specific target
- `EnableVerbose(level)`: Enable verbose logging with level
- `EnableDebug(output)`: Enable debug with output destination
- `EnableForce(overwrite)`: Enable force with overwrite setting
- `HasAnyEnabled()`: Check if any options are enabled
- `GetEnabledOptions()`: Get list of enabled option names

## Application Services

### OptionsService
- **Responsibility**: Orchestrates the execution of all enabled options
- **Dependencies**: CleanerPort, LoggerPort, DebuggerPort, ForcerPort
- **Method**: `ExecuteOptions(ctx, options)` - Executes all enabled options

### Port Interfaces (Dependency Inversion)
- **CleanerPort**: `CleanAux(ctx, target) error`
- **LoggerPort**: `SetVerbosity(level)`, `Log(level, message, args...)`
- **DebuggerPort**: `EnableDebug(output)`, `Debug(message, args...)`
- **ForcerPort**: `SetForceMode(overwrite)`, `ShouldOverwrite() bool`

## Adapters (Infrastructure)

### CleanerAdapter
- **Wraps**: `tex.Cleaner` from existing `tex/` package
- **Implements**: `CleanerPort` interface
- **Functionality**: Delegates to `tex.NewCleaner(target).Clean()`

### LoggerAdapter
- **Wraps**: Go standard `log` package
- **Implements**: `LoggerPort` interface
- **Functionality**: Configurable verbosity levels, structured logging

### DebuggerAdapter
- **Implements**: `DebuggerPort` interface
- **Functionality**: Debug output to stdout, stderr, or file
- **Features**: File-based debug output with automatic cleanup

### ForcerAdapter
- **Implements**: `ForcerPort` interface
- **Functionality**: Force mode with overwrite settings
- **Features**: File existence checking with force override

## CLI Integration

### Updated Argument Parsing
- **Enhanced validation**: Distinguishes between config files and options
- **Config file validation**: Only accepts files with valid extensions (.yaml, .yml, .json, .toml)
- **Option validation**: Strict validation of known options
- **Error handling**: Clear error messages for invalid arguments

### CLI Usage Examples
```bash
# Basic usage
autopdf build template.tex

# With config file
autopdf build template.tex config.yaml

# With single option
autopdf build template.tex clean

# With multiple options
autopdf build template.tex clean verbose debug

# With config and multiple options
autopdf build template.tex config.yaml clean verbose debug force
```

### Updated CLI Limits
- **MaxArgs**: Increased from 3 to 10 to support multiple options
- **Usage**: Updated to `TEMPLATE [CONFIG] [OPTIONS...]`
- **Documentation**: Added comprehensive help text with examples

## SOLID Principles Applied

### Single Responsibility Principle (SRP)
- **Domain**: Each option type has a single responsibility
- **Application**: OptionsService only orchestrates option execution
- **Adapters**: Each adapter handles one specific option type
- **CLI**: Argument parsing separated from option execution

### Open/Closed Principle (OCP)
- **Extensible**: New options can be added without modifying existing code
- **Configurable**: Options support various parameters (levels, targets, outputs)
- **Pluggable**: Adapters can be swapped without changing application logic

### Liskov Substitution Principle (LSP)
- **Port compliance**: All adapters properly implement their interfaces
- **Behavioral consistency**: Adapters maintain expected behavior contracts
- **Substitution**: Any adapter can be substituted without breaking functionality

### Interface Segregation Principle (ISP)
- **Focused interfaces**: Each port interface has specific responsibilities
- **No fat interfaces**: Commands only depend on what they need
- **Clear contracts**: Well-defined responsibilities for each interface

### Dependency Inversion Principle (DIP)
- **Abstractions**: Application depends on port interfaces, not concrete implementations
- **Injection**: Dependencies are injected through constructor
- **Flexibility**: Easy to swap implementations for testing or different environments

## Test Coverage

### Domain Tests
- **BuildOptions**: Default values, option enabling, state management
- **Option types**: Individual option behavior and configuration
- **State queries**: `HasAnyEnabled()`, `GetEnabledOptions()`

### Application Tests
- **OptionsService**: Mock-based testing of option execution
- **Error handling**: Clean failure scenarios and error propagation
- **Integration**: Multiple options execution

### CLI Tests
- **Argument parsing**: Various argument combinations and edge cases
- **Validation**: Invalid arguments, config file validation, option validation
- **Integration**: End-to-end CLI testing with real options

## Benefits Achieved

### 1. **Clean Architecture**
- **Domain**: Pure business logic with no external dependencies
- **Application**: Orchestration layer with clear interfaces
- **Infrastructure**: Adapters that wrap existing functionality

### 2. **Maintainability**
- **Separation of concerns**: Each layer has distinct responsibilities
- **Testability**: Easy to unit test each component in isolation
- **Extensibility**: Simple to add new options or modify existing ones

### 3. **Consistency**
- **Pattern alignment**: Follows same pattern as `tex/clean` command
- **Naming conventions**: Consistent across all layers
- **Error handling**: Uniform error handling and reporting

### 4. **Flexibility**
- **Configurable options**: Each option supports various parameters
- **Pluggable adapters**: Easy to swap implementations
- **Extensible**: Simple to add new option types

## Future Extensibility

### Adding New Options
1. **Domain**: Add new option type to `BuildOptions`
2. **Application**: Add new port interface and update `OptionsService`
3. **Infrastructure**: Create new adapter implementing the port
4. **CLI**: Update argument parser to recognize new option

### Example: Adding `backup` Option
```go
// Domain
type BackupOption struct {
    Enabled bool
    Target  string
}

// Application
type BackupPort interface {
    CreateBackup(ctx context.Context, target string) error
}

// Infrastructure
type BackupAdapter struct {
    // Implementation
}
```

## Status: ✅ COMPLETE

The options DDD structure has been successfully implemented with:
- **Complete domain model** with rich behavior and state management
- **Comprehensive application services** with proper dependency injection
- **Full adapter coverage** for all option types
- **Enhanced CLI integration** with improved argument parsing
- **Extensive test coverage** across all layers
- **Working CLI functionality** with real option execution

The structure follows DDD principles, maintains SOLID design, and provides a solid foundation for future option extensions while keeping the existing `tex/clean` command pattern intact.
