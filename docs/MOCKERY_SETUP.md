# AutoPDF Mockery Setup Guide

## Overview

This guide explains how to use Mockery for automatic mock generation in the AutoPDF project. Mockery eliminates the need to manually maintain mocks, ensuring they stay in sync with interface changes.

## Quick Start

### 1. Generate Mocks
```bash
# Using Makefile (recommended)
make mocks

# Or directly
./scripts/generate_mocks.sh

# Or manually
mockery
```

### 2. Use Mocks in Tests
```go
func TestMyFunction(t *testing.T) {
    // Create mocks
    mockEngine := mocks.NewMockTemplateEngine(t)
    mockValidator := mocks.NewMockTemplateValidator(t)
    
    // Set expectations
    mockEngine.EXPECT().
        Process("test.tex").
        Return("processed content", nil).
        Once()
    
    // Use mocks in your test
    result, err := mockEngine.Process("test.tex")
    require.NoError(t, err)
    assert.Equal(t, "processed content", result)
}
```

## Configuration

The Mockery configuration is in `.mockery.yml`:

```yaml
with-expecter: true
outpkg: mocks
dir: mocks
packages:
  github.com/BuddhiLW/AutoPDF/internal/template:
    interfaces:
      TemplateEngine:
      EnhancedTemplateEngine:
      ConfigProvider:
      VariableProcessor:
      TemplateValidator:
      FileProcessor:
  github.com/BuddhiLW/AutoPDF/pkg/domain:
    interfaces:
      TemplateRepository:
      DocumentRepository:
      TemplateValidationService:
      VariableResolutionService:
      DocumentGenerationService:
```

## Available Mocks

### Template Engine Mocks
- `MockTemplateEngine` - Basic template processing
- `MockEnhancedTemplateEngine` - Enhanced template processing with variables

### Infrastructure Mocks
- `MockConfigProvider` - Configuration management
- `MockFileProcessor` - File system operations
- `MockTemplateValidator` - Template validation
- `MockVariableProcessor` - Variable processing

### Domain Mocks
- `MockTemplateRepository` - Template data access
- `MockDocumentRepository` - Document data access
- `MockTemplateValidationService` - Template validation service
- `MockVariableResolutionService` - Variable resolution service
- `MockDocumentGenerationService` - Document generation service

## Usage Patterns

### 1. Basic Mock Usage
```go
func TestBasicFunctionality(t *testing.T) {
    mockEngine := mocks.NewMockTemplateEngine(t)
    
    mockEngine.EXPECT().
        Process("template.tex").
        Return("result", nil).
        Once()
    
    result, err := mockEngine.Process("template.tex")
    require.NoError(t, err)
    assert.Equal(t, "result", result)
}
```

### 2. Error Scenarios
```go
func TestErrorHandling(t *testing.T) {
    mockEngine := mocks.NewMockTemplateEngine(t)
    
    mockEngine.EXPECT().
        Process("error.tex").
        Return("", errors.New("processing failed")).
        Once()
    
    _, err := mockEngine.Process("error.tex")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "processing failed")
}
```

### 3. Multiple Expectations
```go
func TestMultipleCalls(t *testing.T) {
    mockEngine := mocks.NewMockTemplateEngine(t)
    
    mockEngine.EXPECT().
        Process("template1.tex").
        Return("result1", nil).
        Once()
    
    mockEngine.EXPECT().
        Process("template2.tex").
        Return("result2", nil).
        Once()
    
    // Test multiple calls
    result1, err1 := mockEngine.Process("template1.tex")
    require.NoError(t, err1)
    assert.Equal(t, "result1", result1)
    
    result2, err2 := mockEngine.Process("template2.tex")
    require.NoError(t, err2)
    assert.Equal(t, "result2", result2)
}
```

### 4. Using Mock Suites
```go
func TestWithMockSuite(t *testing.T) {
    suite := NewMockSuite(t)
    suite.SetupBasicMocks()
    
    // Your test logic here
    // All mocks are automatically configured
}
```

## Advanced Features

### 1. Expecter Pattern
Mockery v3 generates expecter methods for type-safe mock setup:

```go
mockEngine.EXPECT().
    Process("test.tex").
    Return("result", nil).
    Once()
```

### 2. Call Verification
```go
// Verify specific number of calls
mockEngine.EXPECT().
    Process("test.tex").
    Return("result", nil).
    Times(3)

// Verify at least N calls
mockEngine.EXPECT().
    Process("test.tex").
    Return("result", nil).
    AtLeast(1)

// Verify at most N calls
mockEngine.EXPECT().
    Process("test.tex").
    Return("result", nil).
    AtMost(5)
```

### 3. Argument Matching
```go
// Match any string
mockEngine.EXPECT().
    Process(mock.AnythingOfType("string")).
    Return("result", nil).
    Once()

// Match specific pattern
mockEngine.EXPECT().
    Process(mock.MatchedBy(func(path string) bool {
        return strings.HasSuffix(path, ".tex")
    })).
    Return("result", nil).
    Once()
```

### 4. Run Functions
```go
mockEngine.EXPECT().
    Process("test.tex").
    Run(func(path string) {
        // Custom logic before return
        fmt.Printf("Processing: %s\n", path)
    }).
    Return("result", nil).
    Once()
```

## Best Practices

### 1. Use Once() for Single Calls
```go
mockEngine.EXPECT().
    Process("test.tex").
    Return("result", nil).
    Once() // Explicitly specify single call
```

### 2. Use Maybe() for Optional Calls
```go
mockEngine.EXPECT().
    Process("optional.tex").
    Return("result", nil).
    Maybe() // Call may or may not happen
```

### 3. Reset Mocks Between Tests
```go
func TestSomething(t *testing.T) {
    mockEngine := mocks.NewMockTemplateEngine(t)
    
    // Set up expectations
    mockEngine.EXPECT().
        Process("test.tex").
        Return("result", nil).
        Once()
    
    // Test logic
    // Mocks are automatically reset after test
}
```

### 4. Use Mock Suites for Complex Tests
```go
func TestComplexScenario(t *testing.T) {
    suite := NewMockSuite(t)
    suite.SetupCartasBackendMocks()
    
    // Test funeral letter generation
    // All mocks are pre-configured
}
```

## Regenerating Mocks

### When to Regenerate
- After adding new interfaces
- After modifying existing interfaces
- After updating Mockery version
- When mocks seem out of sync

### How to Regenerate
```bash
# Using Makefile (recommended)
make mocks

# Or directly
mockery

# Or with script
./scripts/generate_mocks.sh
```

### Verification
```bash
# Run tests to verify mocks work
make test-mocks

# Or specifically
go test -v ./pkg/domain -run TestBackwardCompatibilityWithMocks
```

## Troubleshooting

### Common Issues

1. **Mocks not found**
   ```bash
   # Regenerate mocks
   make mocks
   ```

2. **Interface not found**
   ```bash
   # Check if interface exists
   grep -r "type.*interface" pkg/
   
   # Add to .mockery.yml if missing
   ```

3. **Mock expectations not met**
   ```go
   // Check call order and arguments
   mockEngine.AssertExpectations(t)
   ```

4. **Import errors**
   ```bash
   # Clean and regenerate
   make clean
   make mocks
   ```

### Debug Mode
```bash
# Run with verbose output
mockery --log-level debug
```

## Integration with CI/CD

### GitHub Actions
```yaml
- name: Generate Mocks
  run: make mocks

- name: Run Tests
  run: make test
```

### Pre-commit Hook
```bash
#!/bin/bash
# .git/hooks/pre-commit
make mocks
make test
```

## Performance Considerations

### Mock Overhead
- Mocks add minimal overhead to tests
- Use `Maybe()` for optional calls to reduce setup
- Consider using real implementations for integration tests

### Memory Usage
- Mocks are automatically cleaned up after each test
- No manual cleanup required
- Memory usage is minimal

## Migration from Manual Mocks

### Before (Manual Mocks)
```go
type MockTemplateEngine struct {
    ProcessFunc func(string) (string, error)
}

func (m *MockTemplateEngine) Process(path string) (string, error) {
    return m.ProcessFunc(path)
}
```

### After (Mockery)
```go
mockEngine := mocks.NewMockTemplateEngine(t)
mockEngine.EXPECT().
    Process("test.tex").
    Return("result", nil).
    Once()
```

## Conclusion

Mockery provides a robust, maintainable solution for mock generation in the AutoPDF project. By following this guide, you can:

- Eliminate manual mock maintenance
- Ensure mocks stay in sync with interfaces
- Write more reliable and maintainable tests
- Focus on business logic rather than mock implementation

For questions or issues, refer to the [Mockery documentation](https://vektra.github.io/mockery/) or the project's test examples.
