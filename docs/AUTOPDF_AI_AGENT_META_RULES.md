# AutoPDF: AI Agent Meta-Rules for TDD + SOLID + DDD

## Overview
These meta-rules are derived from applying Test-Driven Development (TDD), SOLID principles, and Domain-Driven Design (DDD) to the AutoPDF project. They serve as guidelines for AI agents to increase assertion, decrease bias, and minimize errors when working with this codebase.

## Core Meta-Rules

### 1. Domain-First Thinking
**Rule:** Always start with the domain model and business rules before considering technical implementation.

**Rationale:** The domain is the heart of the system. Technical concerns should serve the domain, not the other way around.

**Application:**
- Begin every task by understanding the business context
- Identify domain entities, value objects, and business rules first
- Map domain concepts to technical implementation
- Ensure domain logic is not contaminated with technical concerns

**Example:**
```go
// ❌ Wrong: Technical-first approach
type TemplateEngine struct {
    fileSystem FileSystem
    processor  Processor
}

// ✅ Correct: Domain-first approach
type Template struct {
    ID          TemplateID
    Content     string
    Variables   *VariableCollection
    Metadata    TemplateMetadata
}
```

### 2. Test-Driven Assertion
**Rule:** Every piece of code must be driven by a failing test that captures the intended behavior.

**Rationale:** Tests serve as executable specifications that document and verify behavior. They prevent over-engineering and ensure code meets actual requirements.

**Application:**
- Write tests before implementing any feature
- Use tests to explore the problem space
- Let tests guide the design of interfaces and implementations
- Refactor based on test feedback

**Example:**
```go
// ❌ Wrong: Implementation-first
func ProcessTemplate(template string) string {
    // Implementation without clear requirements
    return template
}

// ✅ Correct: Test-driven
func TestProcessTemplate_WithValidTemplate_ReturnsProcessedContent(t *testing.T) {
    // Test captures the requirement
    result := ProcessTemplate("Hello {{name}}")
    assert.Equal(t, "Hello World", result)
}
```

### 3. Single Responsibility Assertion
**Rule:** Each component must have one, and only one, reason to change.

**Rationale:** Components with single responsibilities are easier to understand, test, and maintain. They reduce coupling and increase cohesion.

**Application:**
- Identify the primary responsibility of each component
- Extract secondary responsibilities into separate components
- Use composition to combine responsibilities when needed
- Ensure each component has a clear, single purpose

**Example:**
```go
// ❌ Wrong: Multiple responsibilities
type TemplateEngine struct {
    // Handles template processing, file I/O, and validation
}

// ✅ Correct: Single responsibility
type TemplateProcessor struct {
    // Only handles template processing
}

type FileHandler struct {
    // Only handles file operations
}

type TemplateValidator struct {
    // Only handles validation
}
```

### 4. Interface Segregation Assertion
**Rule:** Clients should not be forced to depend on interfaces they don't use.

**Rationale:** Large interfaces create unnecessary coupling and make testing difficult. Focused interfaces are easier to understand and implement.

**Application:**
- Design interfaces around specific use cases
- Split large interfaces into smaller, focused ones
- Use composition to combine interfaces when needed
- Ensure each interface has a clear, single purpose

**Example:**
```go
// ❌ Wrong: Large interface
type TemplateEngine interface {
    ProcessTemplate(template string) (string, error)
    ValidateTemplate(template string) error
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, content []byte) error
}

// ✅ Correct: Segregated interfaces
type TemplateProcessor interface {
    ProcessTemplate(template string) (string, error)
}

type TemplateValidator interface {
    ValidateTemplate(template string) error
}

type FileHandler interface {
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, content []byte) error
}
```

### 5. Dependency Inversion Assertion
**Rule:** High-level modules should not depend on low-level modules. Both should depend on abstractions.

**Rationale:** Depending on abstractions makes the system more flexible and testable. It allows for easy substitution of implementations.

**Application:**
- Define interfaces for all dependencies
- Inject dependencies through constructors or methods
- Use dependency injection containers when appropriate
- Ensure all dependencies are abstracted

**Example:**
```go
// ❌ Wrong: Direct dependency
type DocumentService struct {
    fileSystem FileSystem
}

// ✅ Correct: Inverted dependency
type DocumentService struct {
    fileHandler FileHandler
}

func NewDocumentService(fileHandler FileHandler) *DocumentService {
    return &DocumentService{fileHandler: fileHandler}
}
```

### 6. Domain-Driven Design Assertion
**Rule:** The domain model should reflect the business reality and be the primary source of truth.

**Rationale:** A well-designed domain model makes the system easier to understand and maintain. It serves as a common language between business and technical stakeholders.

**Application:**
- Model domain concepts as entities and value objects
- Use domain services for complex business logic
- Implement repositories for data access
- Use application services to orchestrate domain operations

**Example:**
```go
// ❌ Wrong: Anemic domain model
type Template struct {
    ID      string
    Content string
}

// ✅ Correct: Rich domain model
type Template struct {
    ID          TemplateID
    Content     string
    Variables   *VariableCollection
    Metadata    TemplateMetadata
}

func (t *Template) Validate() error {
    // Domain logic for validation
}

func (t *Template) Process(variables *VariableCollection) (string, error) {
    // Domain logic for processing
}
```

### 7. Error Handling Assertion
**Rule:** All errors must be handled explicitly and provide meaningful context.

**Rationale:** Proper error handling makes the system more robust and easier to debug. It prevents silent failures and provides clear feedback.

**Application:**
- Define custom error types for different scenarios
- Provide context in error messages
- Use error wrapping to preserve error chains
- Handle errors at the appropriate level

**Example:**
```go
// ❌ Wrong: Generic error handling
func ProcessTemplate(template string) (string, error) {
    if template == "" {
        return "", errors.New("error")
    }
    // ...
}

// ✅ Correct: Specific error handling
type TemplateError struct {
    Type    string
    Message string
    Context map[string]interface{}
}

func (te *TemplateError) Error() string {
    return fmt.Sprintf("%s: %s", te.Type, te.Message)
}

func ProcessTemplate(template string) (string, error) {
    if template == "" {
        return "", &TemplateError{
            Type:    "ValidationError",
            Message: "template cannot be empty",
            Context: map[string]interface{}{"template": template},
        }
    }
    // ...
}
```

### 8. Performance Assertion
**Rule:** Performance requirements must be explicitly defined and tested.

**Rationale:** Performance is a functional requirement that must be verified through testing. It prevents performance regressions and ensures the system meets user expectations.

**Application:**
- Define performance requirements for each component
- Create performance tests that verify requirements
- Use profiling to identify bottlenecks
- Optimize based on actual usage patterns

**Example:**
```go
// ❌ Wrong: No performance requirements
func ProcessTemplate(template string) (string, error) {
    // Implementation without performance considerations
}

// ✅ Correct: Performance requirements
func TestProcessTemplate_Performance(t *testing.T) {
    template := generateLargeTemplate(10000) // 10KB template
    
    start := time.Now()
    result, err := ProcessTemplate(template)
    duration := time.Since(start)
    
    assert.NoError(t, err)
    assert.Less(t, duration, 100*time.Millisecond) // <100ms requirement
}
```

### 9. Concurrency Assertion
**Rule:** All components must be designed for concurrent access from the start.

**Rationale:** Concurrency is a fundamental requirement in modern systems. Designing for concurrency from the start prevents race conditions and performance issues.

**Application:**
- Use immutable data structures where possible
- Implement proper locking mechanisms
- Design for thread safety
- Test concurrent scenarios

**Example:**
```go
// ❌ Wrong: Not thread-safe
type TemplateCache struct {
    cache map[string]*Template
}

// ✅ Correct: Thread-safe
type TemplateCache struct {
    cache map[string]*Template
    mutex sync.RWMutex
}

func (tc *TemplateCache) Get(key string) (*Template, bool) {
    tc.mutex.RLock()
    defer tc.mutex.RUnlock()
    template, exists := tc.cache[key]
    return template, exists
}
```

### 10. Documentation Assertion
**Rule:** All public APIs must be documented with clear examples and usage patterns.

**Rationale:** Good documentation makes the system easier to understand and use. It serves as a contract between the API and its consumers.

**Application:**
- Document all public interfaces
- Provide usage examples
- Document error conditions
- Keep documentation up-to-date

**Example:**
```go
// ❌ Wrong: No documentation
func ProcessTemplate(template string) (string, error) {
    // Implementation
}

// ✅ Correct: Well-documented
// ProcessTemplate processes a template string with variable substitution.
// It returns the processed content or an error if processing fails.
//
// Example:
//   result, err := ProcessTemplate("Hello {{name}}")
//   if err != nil {
//       return "", err
//   }
//   // result will be "Hello World" if name="World"
//
// Errors:
//   - TemplateError: if template syntax is invalid
//   - VariableError: if required variables are missing
func ProcessTemplate(template string) (string, error) {
    // Implementation
}
```

## Anti-Patterns to Avoid

### 1. Anemic Domain Model
**Problem:** Domain objects contain only data without behavior.
**Solution:** Move business logic into domain objects.

### 2. God Object
**Problem:** Single class with too many responsibilities.
**Solution:** Split into multiple focused classes.

### 3. Primitive Obsession
**Problem:** Using primitive types instead of domain-specific types.
**Solution:** Create value objects for domain concepts.

### 4. Feature Envy
**Problem:** Method uses more features of another class than its own.
**Solution:** Move the method to the appropriate class.

### 5. Shotgun Surgery
**Problem:** Making changes requires modifying many classes.
**Solution:** Improve cohesion and reduce coupling.

## Quality Gates

### 1. Test Coverage
- **Domain Layer:** >95% coverage
- **Service Layer:** >90% coverage
- **Infrastructure Layer:** >85% coverage
- **Overall:** >90% coverage

### 2. Code Quality
- **Cyclomatic Complexity:** <10 for all methods
- **Code Duplication:** <5%
- **Technical Debt:** <10% of development time
- **Static Analysis:** No critical issues

### 3. Performance
- **Template Processing:** <100ms for typical templates
- **Memory Usage:** <50MB for large documents
- **Concurrent Processing:** Support 10+ concurrent operations
- **Response Time:** <1s for all operations

### 4. Maintainability
- **Interface Segregation:** Each interface has <5 methods
- **Dependency Injection:** All dependencies injected
- **Error Handling:** Consistent error types and messages
- **Documentation:** All public APIs documented

## Implementation Checklist

### Before Starting
- [ ] Understand the business context and requirements
- [ ] Identify domain entities and business rules
- [ ] Design interfaces and abstractions
- [ ] Plan test strategy and coverage

### During Development
- [ ] Write tests before implementing features
- [ ] Apply SOLID principles consistently
- [ ] Use domain-driven design patterns
- [ ] Handle errors appropriately
- [ ] Document public APIs

### After Implementation
- [ ] Verify test coverage meets requirements
- [ ] Run performance tests
- [ ] Check code quality metrics
- [ ] Review documentation completeness
- [ ] Validate error handling

## Conclusion

These meta-rules provide a framework for AI agents to work effectively with the AutoPDF codebase. By following these rules, agents can:

1. **Increase Assertion:** Tests drive implementation and provide clear requirements
2. **Decrease Bias:** Domain-first thinking prevents technical bias
3. **Minimize Errors:** Proper error handling and validation prevent failures
4. **Improve Quality:** SOLID principles and DDD patterns ensure maintainable code

The key is to apply these rules consistently and use them as a guide for decision-making throughout the development process. They should be treated as living guidelines that evolve with the project and team experience.
