# TDD + SOLID + DDD Analysis & Recommendations

## Overview

This document provides a comprehensive analysis of the AutoPDF codebase for TDD, SOLID, and DDD improvements, along with meta-rules for AI agents working with this codebase.

## Current State Analysis

### ✅ **Strengths**
- **Test Coverage**: Comprehensive test suites with mocking
- **Domain Separation**: Clear separation between AutoPDF, cartas-backend, and edital-pdf-api
- **Interface Design**: Good use of interfaces in AutoPDF
- **Template Processing**: Well-structured template engine

### ⚠️ **Areas for Improvement**
- **Domain Logic**: Business rules scattered across entities
- **Service Layer**: Missing application services
- **Repository Pattern**: Incomplete implementation
- **Event-Driven Architecture**: No domain events
- **Value Objects**: Missing immutable value objects

## DDD + SOLID Refactor Plans

### 1. **Domain Layer Refactoring**

#### **Current Issues**
```go
// Current: Anemic domain model
type Letter struct {
    ID              entity.ID `json:"id,omitempty"`
    Name            string    `json:"name,omitempty"`
    // ... many fields without business logic
}
```

#### **Proposed Solution**
```go
// Domain Entity with Business Logic
type FuneralLetter struct {
    id              LetterID
    deceased        Deceased
    ceremony        Ceremony
    venue           Venue
    status          LetterStatus
    createdAt       time.Time
    updatedAt       time.Time
}

// Business Methods
func (fl *FuneralLetter) ScheduleWake(date time.Time, venue Venue) error {
    if fl.status != LetterStatusDraft {
        return ErrInvalidStatusTransition
    }
    if date.Before(time.Now()) {
        return ErrInvalidDate
    }
    fl.ceremony = NewWake(date, venue)
    fl.status = LetterStatusScheduled
    return nil
}

func (fl *FuneralLetter) CanBePublished() bool {
    return fl.status == LetterStatusScheduled && 
           fl.deceased.IsValid() && 
           fl.ceremony.IsValid()
}
```

### 2. **Value Objects Implementation**

#### **Current Issues**
```go
// Current: Primitive obsession
type Letter struct {
    Name            string    `json:"name,omitempty"`
    YearBirth       YearBirth `json:"year_birth,omitempty"`
    YearDeath       YearDeath `json:"year_death,omitempty"`
}
```

#### **Proposed Solution**
```go
// Value Objects
type PersonName struct {
    value string
}

func NewPersonName(name string) (PersonName, error) {
    if len(strings.TrimSpace(name)) == 0 {
        return PersonName{}, ErrInvalidName
    }
    if len(name) > 100 {
        return PersonName{}, ErrNameTooLong
    }
    return PersonName{value: strings.TrimSpace(name)}, nil
}

func (pn PersonName) String() string {
    return pn.value
}

type BirthYear struct {
    value int
}

func NewBirthYear(year int) (BirthYear, error) {
    currentYear := time.Now().Year()
    if year < 1900 || year > currentYear {
        return BirthYear{}, ErrInvalidBirthYear
    }
    return BirthYear{value: year}, nil
}

type Deceased struct {
    name      PersonName
    birthYear BirthYear
    deathYear DeathYear
}

func (d Deceased) AgeAtDeath() int {
    return d.deathYear.Value() - d.birthYear.Value()
}

func (d Deceased) IsValid() bool {
    return d.name.value != "" && 
           d.birthYear.Value() < d.deathYear.Value()
}
```

### 3. **Domain Services**

#### **Current Issues**
- Business logic scattered across controllers
- No domain services for complex operations

#### **Proposed Solution**
```go
// Domain Service
type FuneralLetterDomainService struct {
    letterRepo LetterRepository
    templateService TemplateService
}

func (s *FuneralLetterDomainService) CreateFuneralLetter(
    deceased Deceased,
    ceremony Ceremony,
    venue Venue,
) (*FuneralLetter, error) {
    // Business rules
    if !deceased.IsValid() {
        return nil, ErrInvalidDeceased
    }
    if !ceremony.IsValid() {
        return nil, ErrInvalidCeremony
    }
    if !venue.IsValid() {
        return nil, ErrInvalidVenue
    }
    
    letter := NewFuneralLetter(deceased, ceremony, venue)
    return letter, nil
}

func (s *FuneralLetterDomainService) GeneratePDF(letter *FuneralLetter) error {
    if !letter.CanBePublished() {
        return ErrLetterNotReadyForPublishing
    }
    
    template := s.templateService.GetTemplate(letter.CeremonyType())
    return s.templateService.ProcessTemplate(template, letter)
}
```

### 4. **Repository Pattern Implementation**

#### **Current Issues**
- Direct database access in services
- No abstraction layer

#### **Proposed Solution**
```go
// Repository Interface
type LetterRepository interface {
    Save(letter *FuneralLetter) error
    FindByID(id LetterID) (*FuneralLetter, error)
    FindByStatus(status LetterStatus) ([]*FuneralLetter, error)
    Update(letter *FuneralLetter) error
    Delete(id LetterID) error
}

// Repository Implementation
type GormLetterRepository struct {
    db *gorm.DB
}

func (r *GormLetterRepository) Save(letter *FuneralLetter) error {
    return r.db.Create(letter).Error
}

func (r *GormLetterRepository) FindByID(id LetterID) (*FuneralLetter, error) {
    var letter FuneralLetter
    err := r.db.Where("id = ?", id.Value()).First(&letter).Error
    if err != nil {
        return nil, err
    }
    return &letter, nil
}
```

### 5. **Application Services (Use Cases)**

#### **Current Issues**
- Business logic in HTTP handlers
- No clear use case boundaries

#### **Proposed Solution**
```go
// Use Case Interface
type CreateFuneralLetterUseCase interface {
    Execute(request CreateFuneralLetterRequest) (*CreateFuneralLetterResponse, error)
}

// Use Case Implementation
type CreateFuneralLetterUseCaseImpl struct {
    letterRepo LetterRepository
    domainService FuneralLetterDomainService
    eventPublisher EventPublisher
}

func (uc *CreateFuneralLetterUseCaseImpl) Execute(
    request CreateFuneralLetterRequest,
) (*CreateFuneralLetterResponse, error) {
    // 1. Validate input
    if err := request.Validate(); err != nil {
        return nil, err
    }
    
    // 2. Create domain objects
    deceased, err := NewDeceased(request.Name, request.BirthYear, request.DeathYear)
    if err != nil {
        return nil, err
    }
    
    ceremony, err := NewCeremony(request.CeremonyType, request.Date, request.Time)
    if err != nil {
        return nil, err
    }
    
    venue, err := NewVenue(request.VenueName, request.VenueAddress)
    if err != nil {
        return nil, err
    }
    
    // 3. Create letter using domain service
    letter, err := uc.domainService.CreateFuneralLetter(deceased, ceremony, venue)
    if err != nil {
        return nil, err
    }
    
    // 4. Persist
    if err := uc.letterRepo.Save(letter); err != nil {
        return nil, err
    }
    
    // 5. Publish domain event
    event := NewFuneralLetterCreatedEvent(letter.ID, letter.Deceased, letter.Ceremony)
    uc.eventPublisher.Publish(event)
    
    return &CreateFuneralLetterResponse{
        LetterID: letter.ID,
        Status:   letter.Status,
    }, nil
}
```

### 6. **Domain Events**

#### **Current Issues**
- No event-driven architecture
- No audit trail

#### **Proposed Solution**
```go
// Domain Event
type FuneralLetterCreatedEvent struct {
    LetterID    LetterID
    Deceased    Deceased
    Ceremony    Ceremony
    OccurredAt  time.Time
}

func (e FuneralLetterCreatedEvent) EventType() string {
    return "funeral_letter.created"
}

// Event Handler
type FuneralLetterCreatedEventHandler struct {
    templateService TemplateService
    pdfService     PDFService
}

func (h *FuneralLetterCreatedEventHandler) Handle(event FuneralLetterCreatedEvent) error {
    // Generate PDF automatically when letter is created
    return h.pdfService.GeneratePDF(event.LetterID)
}
```

## SOLID Principles Implementation

### 1. **Single Responsibility Principle (SRP)**

#### **Current Issues**
- Controllers handling business logic
- Services doing too many things

#### **Proposed Solution**
```go
// Before: Violates SRP
type LetterController struct {
    // Handles HTTP, validation, business logic, persistence
}

// After: Follows SRP
type LetterController struct {
    useCase CreateFuneralLetterUseCase
}

type CreateFuneralLetterUseCase struct {
    // Only handles use case logic
}

type LetterRepository struct {
    // Only handles data persistence
}

type LetterValidator struct {
    // Only handles validation
}
```

### 2. **Open/Closed Principle (OCP)**

#### **Current Issues**
- Hard-coded template processing
- No extension points

#### **Proposed Solution**
```go
// Template Processor Interface
type TemplateProcessor interface {
    Process(template Template, data interface{}) ([]byte, error)
    Supports(templateType string) bool
}

// LaTeX Template Processor
type LaTeXTemplateProcessor struct {
    engine TemplateEngine
}

func (p *LaTeXTemplateProcessor) Process(template Template, data interface{}) ([]byte, error) {
    return p.engine.Process(template.Content, data)
}

// ABNTeX Template Processor
type ABNTeXTemplateProcessor struct {
    engine TemplateEngine
}

func (p *ABNTeXTemplateProcessor) Process(template Template, data interface{}) ([]byte, error) {
    return p.engine.Process(template.Content, data)
}

// Template Processor Factory
type TemplateProcessorFactory struct {
    processors []TemplateProcessor
}

func (f *TemplateProcessorFactory) GetProcessor(templateType string) TemplateProcessor {
    for _, processor := range f.processors {
        if processor.Supports(templateType) {
            return processor
        }
    }
    return nil
}
```

### 3. **Liskov Substitution Principle (LSP)**

#### **Current Issues**
- Interface implementations not interchangeable
- Violations in template processing

#### **Proposed Solution**
```go
// Base Template Engine Interface
type TemplateEngine interface {
    Process(template string, data interface{}) (string, error)
    Validate(template string) error
}

// All implementations must be substitutable
type LaTeXEngine struct{}
type ABNTeXEngine struct{}
type MarkdownEngine struct{}

// Each must implement the same contract
func (e *LaTeXEngine) Process(template string, data interface{}) (string, error) {
    // LaTeX-specific processing
}

func (e *ABNTeXEngine) Process(template string, data interface{}) (string, error) {
    // ABNTeX-specific processing
}
```

### 4. **Interface Segregation Principle (ISP)**

#### **Current Issues**
- Fat interfaces with unused methods
- Clients depending on methods they don't use

#### **Proposed Solution**
```go
// Before: Fat interface
type DocumentProcessor interface {
    ProcessTemplate(template string, data interface{}) ([]byte, error)
    ValidateTemplate(template string) error
    GeneratePDF(content []byte) ([]byte, error)
    SendEmail(recipient string, content []byte) error
    SaveToDatabase(data interface{}) error
}

// After: Segregated interfaces
type TemplateProcessor interface {
    ProcessTemplate(template string, data interface{}) ([]byte, error)
    ValidateTemplate(template string) error
}

type PDFGenerator interface {
    GeneratePDF(content []byte) ([]byte, error)
}

type EmailService interface {
    SendEmail(recipient string, content []byte) error
}

type Repository interface {
    Save(data interface{}) error
}
```

### 5. **Dependency Inversion Principle (DIP)**

#### **Current Issues**
- High-level modules depending on low-level modules
- Hard-coded dependencies

#### **Proposed Solution**
```go
// High-level module depends on abstraction
type FuneralLetterService struct {
    letterRepo    LetterRepository
    templateService TemplateService
    pdfService    PDFService
    eventPublisher EventPublisher
}

// Constructor injection
func NewFuneralLetterService(
    letterRepo LetterRepository,
    templateService TemplateService,
    pdfService PDFService,
    eventPublisher EventPublisher,
) *FuneralLetterService {
    return &FuneralLetterService{
        letterRepo:     letterRepo,
        templateService: templateService,
        pdfService:     pdfService,
        eventPublisher: eventPublisher,
    }
}
```

## TDD Task List for Other Agents

### **Phase 1: Domain Layer TDD**

#### **Task 1: Value Objects TDD**
```go
// Test: PersonName value object
func TestPersonName_NewPersonName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    PersonName
        wantErr error
    }{
        {
            name:    "valid name",
            input:   "João da Silva",
            want:    PersonName{value: "João da Silva"},
            wantErr: nil,
        },
        {
            name:    "empty name",
            input:   "",
            want:    PersonName{},
            wantErr: ErrInvalidName,
        },
        {
            name:    "name too long",
            input:   strings.Repeat("a", 101),
            want:    PersonName{},
            wantErr: ErrNameTooLong,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := NewPersonName(tt.input)
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("NewPersonName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("NewPersonName() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

#### **Task 2: Domain Entities TDD**
```go
// Test: FuneralLetter domain entity
func TestFuneralLetter_ScheduleWake(t *testing.T) {
    letter := NewFuneralLetter(
        NewDeceased("João", 1950, 2024),
        NewCeremony(CeremonyTypeWake, time.Now().Add(24*time.Hour)),
        NewVenue("Capela São José", "Rua das Flores, 123"),
    )
    
    venue := NewVenue("Nova Capela", "Rua Nova, 456")
    date := time.Now().Add(48 * time.Hour)
    
    err := letter.ScheduleWake(date, venue)
    
    assert.NoError(t, err)
    assert.Equal(t, LetterStatusScheduled, letter.Status)
    assert.Equal(t, venue, letter.Ceremony.Venue)
}
```

#### **Task 3: Domain Services TDD**
```go
// Test: FuneralLetterDomainService
func TestFuneralLetterDomainService_CreateFuneralLetter(t *testing.T) {
    mockRepo := mocks.NewMockLetterRepository(t)
    mockTemplate := mocks.NewMockTemplateService(t)
    
    service := NewFuneralLetterDomainService(mockRepo, mockTemplate)
    
    deceased := NewDeceased("Maria", 1940, 2024)
    ceremony := NewCeremony(CeremonyTypeWake, time.Now().Add(24*time.Hour))
    venue := NewVenue("Capela", "Rua das Flores, 123")
    
    letter, err := service.CreateFuneralLetter(deceased, ceremony, venue)
    
    assert.NoError(t, err)
    assert.NotNil(t, letter)
    assert.Equal(t, deceased, letter.Deceased)
    assert.Equal(t, ceremony, letter.Ceremony)
}
```

### **Phase 2: Application Layer TDD**

#### **Task 4: Use Cases TDD**
```go
// Test: CreateFuneralLetterUseCase
func TestCreateFuneralLetterUseCase_Execute(t *testing.T) {
    mockRepo := mocks.NewMockLetterRepository(t)
    mockDomainService := mocks.NewMockFuneralLetterDomainService(t)
    mockEventPublisher := mocks.NewMockEventPublisher(t)
    
    useCase := NewCreateFuneralLetterUseCase(mockRepo, mockDomainService, mockEventPublisher)
    
    request := CreateFuneralLetterRequest{
        Name:        "João da Silva",
        BirthYear:   1950,
        DeathYear:   2024,
        CeremonyType: CeremonyTypeWake,
        Date:        time.Now().Add(24 * time.Hour),
        VenueName:   "Capela São José",
        VenueAddress: "Rua das Flores, 123",
    }
    
    expectedLetter := &FuneralLetter{
        ID: LetterID("123"),
        Status: LetterStatusDraft,
    }
    
    mockDomainService.EXPECT().
        CreateFuneralLetter(mock.Anything, mock.Anything, mock.Anything).
        Return(expectedLetter, nil).
        Once()
    
    mockRepo.EXPECT().
        Save(expectedLetter).
        Return(nil).
        Once()
    
    mockEventPublisher.EXPECT().
        Publish(mock.Anything).
        Return(nil).
        Once()
    
    response, err := useCase.Execute(request)
    
    assert.NoError(t, err)
    assert.Equal(t, expectedLetter.ID, response.LetterID)
    assert.Equal(t, expectedLetter.Status, response.Status)
}
```

### **Phase 3: Infrastructure Layer TDD**

#### **Task 5: Repository TDD**
```go
// Test: GormLetterRepository
func TestGormLetterRepository_Save(t *testing.T) {
    db := setupTestDB(t)
    repo := NewGormLetterRepository(db)
    
    letter := &FuneralLetter{
        ID: LetterID("123"),
        Status: LetterStatusDraft,
    }
    
    err := repo.Save(letter)
    
    assert.NoError(t, err)
    
    // Verify in database
    var savedLetter FuneralLetter
    err = db.Where("id = ?", "123").First(&savedLetter).Error
    assert.NoError(t, err)
    assert.Equal(t, letter.ID, savedLetter.ID)
}
```

#### **Task 6: Template Service TDD**
```go
// Test: TemplateService
func TestTemplateService_ProcessTemplate(t *testing.T) {
    mockEngine := mocks.NewMockTemplateEngine(t)
    service := NewTemplateService(mockEngine)
    
    template := Template{
        Content: "Hello {{.Name}}",
        Type:    "latex",
    }
    
    data := map[string]interface{}{
        "Name": "João",
    }
    
    expectedResult := "Hello João"
    
    mockEngine.EXPECT().
        Process(template.Content, data).
        Return(expectedResult, nil).
        Once()
    
    result, err := service.ProcessTemplate(template, data)
    
    assert.NoError(t, err)
    assert.Equal(t, expectedResult, result)
}
```

## Meta-Rules for AI Agents

### **Rule 1: Domain-First Thinking**
```
When working with this codebase:
1. Always start with domain entities and business logic
2. Identify business rules and invariants first
3. Create value objects for primitive obsession
4. Implement domain services for complex business logic
5. Never put business logic in controllers or infrastructure
```

### **Rule 2: TDD Red-Green-Refactor Cycle**
```
For every new feature:
1. RED: Write failing test first
2. GREEN: Write minimal code to pass
3. REFACTOR: Improve design while keeping tests green
4. Always test business rules, not implementation details
5. Use mocks for external dependencies
```

### **Rule 3: SOLID Principle Validation**
```
Before implementing any class/function:
1. SRP: Does it have one reason to change?
2. OCP: Is it open for extension, closed for modification?
3. LSP: Can derived classes substitute base classes?
4. ISP: Are interfaces focused and minimal?
5. DIP: Does it depend on abstractions, not concretions?
```

### **Rule 4: DDD Bounded Context Awareness**
```
When working with multiple domains:
1. Identify bounded contexts (Funeral Letters, Legal Documents, PDF Generation)
2. Keep domain logic within context boundaries
3. Use domain events for cross-context communication
4. Implement anti-corruption layers between contexts
5. Maintain clear aggregate boundaries
```

### **Rule 5: Error Handling Strategy**
```
For error handling:
1. Use domain-specific error types
2. Never return generic errors from domain layer
3. Implement error recovery strategies
4. Log errors with context
5. Use error codes for client handling
```

### **Rule 6: Testing Strategy**
```
For testing:
1. Test behavior, not implementation
2. Use test doubles (mocks, stubs, fakes) appropriately
3. Test edge cases and error scenarios
4. Maintain test independence
5. Use descriptive test names
```

### **Rule 7: Code Organization**
```
For code structure:
1. Group by feature, not by technical layer
2. Use dependency injection
3. Implement interfaces for external dependencies
4. Keep functions small and focused
5. Use meaningful names
```

### **Rule 8: Performance Considerations**
```
For performance:
1. Test with realistic data volumes
2. Use async processing for long operations
3. Implement caching strategies
4. Monitor resource usage
5. Profile before optimizing
```

## Implementation Priority

### **Phase 1: Foundation (Weeks 1-2)**
1. Implement value objects for primitive obsession
2. Create domain entities with business logic
3. Implement domain services
4. Set up repository interfaces

### **Phase 2: Application Layer (Weeks 3-4)**
1. Implement use cases
2. Create application services
3. Implement domain events
4. Set up dependency injection

### **Phase 3: Infrastructure (Weeks 5-6)**
1. Implement repository concrete classes
2. Create template processors
3. Implement event handlers
4. Set up database migrations

### **Phase 4: Integration (Weeks 7-8)**
1. Integrate all layers
2. Implement end-to-end tests
3. Performance testing
4. Documentation and deployment

## Success Metrics

### **Code Quality**
- [ ] 90%+ test coverage
- [ ] All SOLID principles followed
- [ ] Domain logic isolated
- [ ] No business logic in controllers

### **Performance**
- [ ] < 2s response time for PDF generation
- [ ] < 100ms for simple operations
- [ ] Support for 1000+ concurrent requests
- [ ] Memory usage < 512MB per request

### **Maintainability**
- [ ] Clear separation of concerns
- [ ] Easy to add new features
- [ ] Simple to test
- [ ] Well-documented APIs

This analysis provides a comprehensive roadmap for implementing TDD, SOLID, and DDD principles in the AutoPDF codebase, with specific tasks for AI agents and meta-rules for consistent development practices.
