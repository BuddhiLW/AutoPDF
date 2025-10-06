# AI Agent Meta-Rules for TDD + SOLID + DDD

## Overview

This document establishes meta-rules for AI agents working with the AutoPDF codebase, based on meta-learning from TDD, SOLID, and DDD principles applied to document generation systems.

## Core Meta-Rules

### **Rule 1: Domain-First Thinking**
```
When working with this codebase:
1. Always start with domain entities and business logic
2. Identify business rules and invariants first
3. Create value objects for primitive obsession
4. Implement domain services for complex business logic
5. Never put business logic in controllers or infrastructure
6. Ask: "What business problem are we solving?"
7. Validate: "Does this code reflect the business domain?"
```

**Anti-Patterns to Avoid**:
- Anemic domain models (entities with only getters/setters)
- Business logic in controllers
- Primitive obsession (using strings/ints instead of value objects)
- Technical concerns in domain layer

### **Rule 2: TDD Red-Green-Refactor Cycle**
```
For every new feature:
1. RED: Write failing test first
2. GREEN: Write minimal code to pass
3. REFACTOR: Improve design while keeping tests green
4. Always test business rules, not implementation details
5. Use mocks for external dependencies
6. Test edge cases and error scenarios
7. Maintain test independence
```

**Test Quality Indicators**:
- Tests fail for the right reason
- Tests are readable and maintainable
- Tests run fast and independently
- Tests cover edge cases and error scenarios
- Tests document business behavior

### **Rule 3: SOLID Principle Validation**
```
Before implementing any class/function:
1. SRP: Does it have one reason to change?
2. OCP: Is it open for extension, closed for modification?
3. LSP: Can derived classes substitute base classes?
4. ISP: Are interfaces focused and minimal?
5. DIP: Does it depend on abstractions, not concretions?
```

**SOLID Validation Checklist**:
- [ ] Single Responsibility: One class, one reason to change
- [ ] Open/Closed: Open for extension, closed for modification
- [ ] Liskov Substitution: Derived classes must be substitutable
- [ ] Interface Segregation: Many specific interfaces vs one general
- [ ] Dependency Inversion: Depend on abstractions, not concretions

### **Rule 4: DDD Bounded Context Awareness**
```
When working with multiple domains:
1. Identify bounded contexts (Funeral Letters, Legal Documents, PDF Generation)
2. Keep domain logic within context boundaries
3. Use domain events for cross-context communication
4. Implement anti-corruption layers between contexts
5. Maintain clear aggregate boundaries
6. Respect context maps and relationships
```

**Context Boundaries**:
- **Funeral Letters**: Death announcements, wake scheduling, venue management
- **Legal Documents**: Court proceedings, auction notices, legal compliance
- **PDF Generation**: Template processing, document formatting, output generation

### **Rule 5: Error Handling Strategy**
```
For error handling:
1. Use domain-specific error types
2. Never return generic errors from domain layer
3. Implement error recovery strategies
4. Log errors with context
5. Use error codes for client handling
6. Fail fast with meaningful messages
7. Implement circuit breakers for external services
```

**Error Handling Patterns**:
```go
// Domain-specific errors
type FuneralLetterError struct {
    Code    string
    Message string
    Context map[string]interface{}
}

// Error recovery strategies
func (s *FuneralLetterService) CreateLetterWithRetry(request CreateLetterRequest) (*FuneralLetter, error) {
    for attempt := 1; attempt <= 3; attempt++ {
        letter, err := s.CreateLetter(request)
        if err == nil {
            return letter, nil
        }
        if !isRetryableError(err) {
            return nil, err
        }
        time.Sleep(time.Duration(attempt) * time.Second)
    }
    return nil, ErrMaxRetriesExceeded
}
```

### **Rule 6: Testing Strategy**
```
For testing:
1. Test behavior, not implementation
2. Use test doubles (mocks, stubs, fakes) appropriately
3. Test edge cases and error scenarios
4. Maintain test independence
5. Use descriptive test names
6. Test business rules, not technical details
7. Implement property-based testing for complex logic
```

**Testing Patterns**:
```go
// Behavior-driven test names
func TestFuneralLetter_ShouldScheduleWake_WhenValidDateAndVenue(t *testing.T) {
    // Test implementation
}

// Property-based testing
func TestPersonName_Validation(t *testing.T) {
    quick.Check(func(name string) bool {
        personName, err := NewPersonName(name)
        if len(name) == 0 || len(name) > 100 {
            return err != nil
        }
        return err == nil && personName.String() == strings.TrimSpace(name)
    }, nil)
}
```

### **Rule 7: Code Organization**
```
For code structure:
1. Group by feature, not by technical layer
2. Use dependency injection
3. Implement interfaces for external dependencies
4. Keep functions small and focused
5. Use meaningful names
6. Follow consistent naming conventions
7. Implement proper abstraction layers
```

**Code Organization Patterns**:
```
project/
├── domain/
│   ├── entities/
│   ├── valueobjects/
│   ├── services/
│   └── events/
├── application/
│   ├── usecases/
│   ├── services/
│   └── dto/
├── infrastructure/
│   ├── repositories/
│   ├── external/
│   └── persistence/
└── interfaces/
    ├── http/
    ├── grpc/
    └── messaging/
```

### **Rule 8: Performance Considerations**
```
For performance:
1. Test with realistic data volumes
2. Use async processing for long operations
3. Implement caching strategies
4. Monitor resource usage
5. Profile before optimizing
6. Use connection pooling
7. Implement rate limiting
```

**Performance Patterns**:
```go
// Async processing
func (s *FuneralLetterService) GeneratePDFAsync(letterID LetterID) <-chan error {
    result := make(chan error, 1)
    go func() {
        defer close(result)
        err := s.GeneratePDF(letterID)
        result <- err
    }()
    return result
}

// Caching strategy
func (s *FuneralLetterService) GetTemplate(templateType string) (*Template, error) {
    if cached, found := s.cache.Get(templateType); found {
        return cached.(*Template), nil
    }
    
    template, err := s.templateRepo.FindByType(templateType)
    if err != nil {
        return nil, err
    }
    
    s.cache.Set(templateType, template, 1*time.Hour)
    return template, nil
}
```

## Specialized Meta-Rules

### **For Domain Specialists**
```
1. Focus on business logic and domain rules
2. Implement value objects and entities
3. Create domain services
4. Ensure business invariants
5. Use domain events for cross-aggregate communication
6. Implement aggregate boundaries
7. Validate business rules
```

### **For Application Specialists**
```
1. Focus on use cases and orchestration
2. Implement application services
3. Handle cross-cutting concerns
4. Manage transactions
5. Implement CQRS patterns
6. Handle event sourcing
7. Manage application state
```

### **For Infrastructure Specialists**
```
1. Focus on external integrations
2. Implement repositories
3. Handle technical concerns
4. Manage performance
5. Implement monitoring
6. Handle deployment
7. Manage external dependencies
```

### **For QA Specialists**
```
1. Focus on testing and quality
2. Implement integration tests
3. Ensure test coverage
4. Validate performance
5. Test error scenarios
6. Implement load testing
7. Validate business rules
```

## Anti-Patterns to Avoid

### **Domain Anti-Patterns**
- **Anemic Domain Model**: Entities with only getters/setters
- **God Object**: Classes with too many responsibilities
- **Primitive Obsession**: Using primitives instead of value objects
- **Feature Envy**: Methods that use more data from other classes
- **Shotgun Surgery**: Changes require modifications in many places

### **Testing Anti-Patterns**
- **Brittle Tests**: Tests that break when implementation changes
- **Slow Tests**: Tests that take too long to run
- **Test Duplication**: Repeated test code
- **Testing Implementation**: Testing internal details instead of behavior
- **Mock Overuse**: Mocking everything instead of using real objects

### **SOLID Anti-Patterns**
- **God Class**: Classes with too many responsibilities
- **Interface Pollution**: Too many interfaces with single methods
- **Dependency Hell**: Circular dependencies
- **Rigid Design**: Hard to change without breaking existing code
- **Fragile Design**: Changes break unexpected parts

## Quality Gates

### **Code Quality Gates**
- [ ] 90%+ test coverage
- [ ] All SOLID principles followed
- [ ] Domain logic isolated
- [ ] No business logic in controllers
- [ ] Comprehensive error handling
- [ ] Performance benchmarks met
- [ ] Security vulnerabilities addressed

### **Architecture Quality Gates**
- [ ] Clear separation of concerns
- [ ] Proper abstraction layers
- [ ] Loose coupling between modules
- [ ] High cohesion within modules
- [ ] Extensible design
- [ ] Maintainable code structure

### **Business Quality Gates**
- [ ] Business rules are enforced
- [ ] Domain invariants are maintained
- [ ] User requirements are met
- [ ] Performance is acceptable
- [ ] Error handling is comprehensive
- [ ] Audit trail is maintained

## Learning and Adaptation

### **Meta-Learning Rules**
```
1. Analyze patterns in successful implementations
2. Identify common failure modes
3. Adapt rules based on project context
4. Learn from code reviews and feedback
5. Continuously improve development practices
6. Share knowledge across the team
7. Document lessons learned
```

### **Context-Aware Adaptation**
```
1. Funeral Letter Domain: Focus on scheduling, venue management, family notifications
2. Legal Document Domain: Focus on compliance, court procedures, legal validation
3. PDF Generation Domain: Focus on template processing, formatting, output quality
4. Cross-Domain: Focus on integration, event handling, data consistency
```

### **Continuous Improvement**
```
1. Regular code reviews
2. Retrospectives on development practices
3. Refactoring based on new requirements
4. Performance monitoring and optimization
5. Security assessment and updates
6. Documentation updates
7. Knowledge sharing sessions
```

## Conclusion

These meta-rules provide a comprehensive framework for AI agents working with the AutoPDF codebase, ensuring consistent application of TDD, SOLID, and DDD principles while maintaining high code quality and business value.

The rules are designed to:
- Increase assertion and confidence in code quality
- Decrease bias and errors in implementation
- Provide clear guidance for different types of specialists
- Enable continuous learning and adaptation
- Ensure business value is maintained throughout development

By following these meta-rules, AI agents can work effectively within the AutoPDF ecosystem while maintaining high standards of code quality, architecture, and business value.
