# 🏗️ AutoPDF SOLID + DDD + GoF Testing Guide

This document explains how to test and use the refactored SOLID + DDD + GoF architecture in AutoPDF.

## 📋 Table of Contents

- [Architecture Overview](#architecture-overview)
- [Running Tests](#running-tests)
- [CLI Showcase](#cli-showcase)
- [Integration Tests](#integration-tests)
- [Code Examples](#code-examples)
- [Test Data](#test-data)

---

## 🏛️ Architecture Overview

### SOLID Principles

- **S**ingle Responsibility: Each service has one clear purpose
- **O**pen/Closed: Extensible through interfaces and strategies
- **L**iskov Substitution: All implementations are interchangeable
- **I**nterface Segregation: Small, focused interfaces
- **D**ependency Inversion: Depends on abstractions, not concretions

### Domain-Driven Design

- **Domain Services**: Business logic orchestration
- **Value Objects**: Immutable data with validation
- **Entities**: Objects with identity and lifecycle
- **Factories**: Complex object creation
- **Events**: Loose coupling through event-driven architecture

### Gang of Four Patterns

- **Factory Pattern**: Engine creation and selection
- **Strategy Pattern**: Template processing strategies
- **Observer Pattern**: Event-driven architecture
- **Facade Pattern**: Simple interface to complex operations
- **Singleton Pattern**: Service factory instance
- **Command Pattern**: CLI command encapsulation
- **Adapter Pattern**: Bonzai CLI to domain services

---

## 🧪 Running Tests

### Unit Tests (Domain Layer)

```bash
# Test domain engines and factories
go test ./internal/autopdf/domain/... -v

# Test with coverage
go test ./internal/autopdf/domain/... -cover -coverprofile=coverage.out
```

### Unit Tests (Domain Services with Mocks)

```bash
# Test domain services using generated mocks
go test ./internal/autopdf/domain_test/... -v
```

### Unit Tests (Application Layer)

```bash
# Test application services
go test ./internal/autopdf/application/... -v
```

### Integration Tests

```bash
# Run full integration tests with real templates
go test ./test/integration/... -v

# Run specific integration test
go test ./test/integration/... -v -run TestSOLIDDDDIntegration
```

### All Tests

```bash
# Run all tests in the project
go test ./... -v

# Run with race detection
go test ./... -race

# Run with coverage
go test ./... -cover
```

---

## 🎬 CLI Showcase

### Quick Start

```bash
# Run the comprehensive CLI showcase
./test/cli_showcase.sh
```

This script demonstrates:
- ✅ Simple templates with basic variables
- ✅ Complex templates with nested data structures
- ✅ Arrays and loops in templates
- ✅ Multiple LaTeX engines (pdflatex, xelatex)
- ✅ Real-world examples (letters, documents)
- ✅ Clean command (Domain Services)
- ✅ Conversion command (Strategy Pattern)

### Manual CLI Testing

#### 1. Build Simple Template

```bash
./autopdf build templates/sample-template.tex configs/sample-config.yaml
```

**What it demonstrates:**
- Basic variable substitution
- Simple template processing
- Domain Services in action

#### 2. Build Enhanced Complex Template

```bash
./autopdf build templates/enhanced-document.tex configs/enhanced-sample-config.yaml
```

**What it demonstrates:**
- Nested objects (company.address.street)
- Arrays and loops (team members, proceedings)
- Complex data structures (legal_document)
- Multiple levels of nesting (4+ levels deep)

#### 3. Build with XeLaTeX Engine

```bash
./autopdf build test/model_xelatex/main.tex test/model_xelatex/config.yaml
```

**What it demonstrates:**
- Factory Pattern (engine selection)
- Strategy Pattern (different PDF engines)
- Engine-specific features

#### 4. Clean Auxiliary Files

```bash
./autopdf clean <directory>
```

**What it demonstrates:**
- Domain Services (FileManagementService)
- Single Responsibility Principle
- Clean separation of concerns

#### 5. Convert PDF to Images

```bash
./autopdf convert output/document.pdf png jpg
```

**What it demonstrates:**
- Strategy Pattern (multiple conversion engines)
- Automatic fallback (ImageMagick → Poppler)
- Graceful degradation

---

## 🔬 Integration Tests

### Test Structure

```
test/
├── integration/
│   └── solid_ddd_integration_test.go  # Full integration tests
├── examples/
│   ├── solid_ddd_showcase.go          # Programmatic showcase
│   └── enhanced_template_example.go   # Enhanced features demo
└── cli_showcase.sh                     # CLI demonstration script
```

### Key Integration Tests

#### 1. Simple Sample Template Test

```go
func TestSOLIDDDDIntegration(t *testing.T) {
    t.Run("SimpleSampleTemplate", func(t *testing.T) {
        // Tests basic template processing with domain services
    })
}
```

**Tests:**
- Service factory initialization
- Configuration loading
- Template processing
- PDF generation
- File verification

#### 2. Enhanced Complex Template Test

```go
t.Run("EnhancedComplexTemplate", func(t *testing.T) {
    // Tests complex data structures and nested variables
})
```

**Tests:**
- Nested object access
- Array iteration
- Loop processing
- Complex variable resolution
- Multi-level data structures

#### 3. Conversion Integration Test

```go
func TestConversionIntegration(t *testing.T) {
    // Tests Strategy Pattern with multiple conversion engines
}
```

**Tests:**
- Multiple conversion strategies
- Automatic engine selection
- Fallback mechanisms
- Error handling

#### 4. Factory Pattern Test

```go
func TestFactoryPatterns(t *testing.T) {
    // Tests all factory implementations
}
```

**Tests:**
- TemplateEngineFactory
- PDFEngineFactory
- ConversionEngineFactory
- Engine availability
- Error handling for unsupported engines

#### 5. Event-Driven Architecture Test

```go
func TestEventDrivenArchitecture(t *testing.T) {
    // Tests Observer Pattern implementation
}
```

**Tests:**
- Event publishing
- Event subscription
- Event handling
- Multiple handlers
- Event unsubscription

---

## 💻 Code Examples

### Example 1: Using the Service Factory

```go
import "github.com/BuddhiLW/AutoPDF/internal/autopdf/application"

// Get the singleton service factory
factory := application.GetDefaultFactory()

// Get any service you need
buildService := factory.GetBuildService()
configService := factory.GetConfigurationService()
```

**Demonstrates:**
- Singleton Pattern
- Dependency Injection
- Factory Pattern

### Example 2: Building a PDF

```go
import (
    "context"
    "github.com/BuddhiLW/AutoPDF/internal/autopdf/domain"
)

ctx := context.Background()

result, err := buildService.BuildPDF(ctx, &domain.BuildRequest{
    TemplatePath: "template.tex",
    OutputPath:   "output.pdf",
    Variables: map[string]interface{}{
        "title": "My Document",
        "author": "John Doe",
    },
    ShouldClean: true,
})
```

**Demonstrates:**
- Application Service usage
- Domain Request/Response objects
- Context propagation
- Error handling

### Example 3: Using Factory Pattern

```go
// Create a PDF engine factory
factory := domain.NewPDFEngineFactory()

// Get available engines
engines := factory.GetAvailableEngines()
// Returns: ["pdflatex", "xelatex", "lualatex"]

// Create a specific engine
engine, err := factory.CreateEngine("xelatex")
```

**Demonstrates:**
- Factory Pattern
- Engine abstraction
- Strategy selection

### Example 4: Event-Driven Architecture

```go
// Get event publisher
publisher := factory.GetEventPublisher()

// Create a custom handler
type MyHandler struct{}

func (h *MyHandler) Handle(event domain.Event) error {
    log.Printf("Event: %s", event.GetEventType())
    return nil
}

// Subscribe to events
handler := &MyHandler{}
publisher.Subscribe("pdf.generated", handler)

// Events are automatically published during build
```

**Demonstrates:**
- Observer Pattern
- Event subscription
- Custom event handlers
- Loose coupling

---

## 📊 Test Data

### Available Test Configurations

#### 1. Simple Sample Config (`configs/sample-config.yaml`)

```yaml
template: "templates/document.tex"
output: "output/final.pdf"
variables:
  title: "Sample Document"
  author: "John Doe"
engine: "pdflatex"
```

**Use case:** Basic template processing

#### 2. Enhanced Sample Config (`configs/enhanced-sample-config.yaml`)

```yaml
variables:
  company:
    name: "AutoPDF Solutions"
    address:
      street: "123 Technology Drive"
      city: "San Francisco"
  team:
    - name: "John Doe"
      role: "Lead Developer"
      skills: ["Go", "LaTeX", "Templates"]
```

**Use case:** Complex nested data structures

#### 3. Complex Test Config (`internal/autopdf/test-data/complex_config.yaml`)

```yaml
variables:
  project:
    name: "AutoPDF"
    items:
      - name: "Feature A"
        status: "Complete"
      - name: "Feature B"
        status: "In Progress"
```

**Use case:** Arrays, loops, and nested structures

### Available Templates

#### 1. Simple Template (`templates/sample-template.tex`)

```latex
\title{delim[[.title]]}
\author{delim[[.author]]}
```

**Features:** Basic variable substitution

#### 2. Enhanced Template (`templates/enhanced-document.tex`)

```latex
delim[[range .team]]
\subsection{delim[[.name]]}
\item \textbf{Role:} delim[[.role]]
\item \textbf{Skills:} delim[[join ", " .skills]]
delim[[end]]
```

**Features:** Loops, nested access, array operations

---

## 🎯 Testing Best Practices

### 1. Use Integration Tests for Real Workflows

```go
// Good: Test real workflow with actual files
func TestRealWorkflow(t *testing.T) {
    cfg, _ := configService.LoadConfiguration(ctx, "real-config.yaml")
    result, _ := buildService.BuildPDF(ctx, request)
    assert.FileExists(t, result.PDFPath)
}
```

### 2. Use Mocks for Unit Tests

```go
// Good: Test service in isolation with mocks
func TestServiceLogic(t *testing.T) {
    mockEngine := &mocks.MockDomainTemplateEngine{}
    mockEngine.On("Process", ...).Return("result", nil)
    // Test service logic
}
```

### 3. Test Error Paths

```go
// Good: Test error handling
func TestErrorHandling(t *testing.T) {
    _, err := service.BuildPDF(ctx, invalidRequest)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "expected error message")
}
```

### 4. Verify Architecture Patterns

```go
// Good: Verify patterns are working
func TestFactoryPattern(t *testing.T) {
    factory := domain.NewPDFEngineFactory()
    engine, _ := factory.CreateEngine("pdflatex")
    assert.Implements(t, (*domain.PDFEngine)(nil), engine)
}
```

---

## 📚 Additional Resources

- **Main README**: `../README.md`
- **Mockery Setup**: `../docs/MOCKERY_SETUP.md`
- **Architecture Docs**: `../docs/AUTOPDF_TDD_SOLID_DDD_ANALYSIS.md`
- **API Documentation**: Run `godoc -http=:6060`

---

## 🤝 Contributing

When adding new tests:

1. ✅ Test with real files (integration tests)
2. ✅ Test with mocks (unit tests)
3. ✅ Test error paths
4. ✅ Verify architecture patterns
5. ✅ Document what you're testing
6. ✅ Use descriptive test names

---

## 📝 Summary

The refactored SOLID + DDD + GoF architecture provides:

- ✅ **Testability**: All components mockable
- ✅ **Maintainability**: Clear separation of concerns
- ✅ **Extensibility**: Easy to add new features
- ✅ **Reliability**: Comprehensive test coverage
- ✅ **Observability**: Event-driven architecture
- ✅ **Flexibility**: Multiple strategies and engines

**Run the tests, explore the examples, and see the architecture in action!** 🚀
