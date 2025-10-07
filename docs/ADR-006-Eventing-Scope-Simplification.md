# ADR-006: Eventing Scope Simplification

## Status
Accepted

## Context
The initial DDD refactoring plan included a complex EventBus system in Phase 1, but this adds unnecessary complexity when no real consumers exist yet. We should defer eventing infrastructure until there's a clear need.

## Decision

**Defer EventBus Implementation to Phase 2**

For Phase 1, we will:
1. **Raise domain events** on aggregates (for future use)
2. **Store events** with the aggregate (if needed for audit)
3. **Defer publishing** until real consumers exist
4. **Focus on core domain** and port/adapter implementation

### Rationale
- **Avoid Premature Complexity**: Don't build infrastructure without consumers
- **Focus on Core**: Phase 1 should focus on domain extraction and seams
- **Incremental Value**: Events add value when there are actual subscribers
- **Simpler Testing**: Fewer moving parts in Phase 1 tests

## Implementation

### Phase 1: Simple Event Storage
```go
// internal/document/domain/entities/document.go
package entities

type Document struct {
    // ... other fields
    events []DomainEvent
}

// Add event to aggregate
func (d *Document) addEvent(event DomainEvent) {
    d.events = append(d.events, event)
}

// Get events for persistence
func (d *Document) Events() []DomainEvent {
    return d.events
}

// Clear events after processing
func (d *Document) ClearEvents() {
    d.events = make([]DomainEvent, 0)
}
```

### Phase 1: Domain Events (No Publishing)
```go
// internal/document/domain/events/document_events.go
package events

type DocumentProcessingStarted struct {
    DocumentID string
    OccurredAt time.Time
}

type DocumentProcessingCompleted struct {
    DocumentID string
    OutputPath string
    OccurredAt time.Time
}

// Simple event creation - no publishing
func NewDocumentProcessingStarted(documentID string) *DocumentProcessingStarted {
    return &DocumentProcessingStarted{
        DocumentID: documentID,
        OccurredAt: time.Now(),
    }
}
```

### Phase 2: Event Publishing (When Needed)
```go
// internal/document/infra/event_publisher.go
package infra

type EventPublisher struct {
    // Implementation when real consumers exist
}

func (ep *EventPublisher) Publish(ctx context.Context, event DomainEvent) error {
    // Implementation when needed
    return nil
}
```

## Benefits

1. **Simpler Phase 1**: Focus on core domain and seams
2. **No Premature Complexity**: Don't build infrastructure without need
3. **Easier Testing**: Fewer dependencies in Phase 1
4. **Clear Progression**: Events added when value is clear
5. **Domain Events Ready**: Events are defined and stored

## Success Criteria

- [ ] Domain events are defined and stored on aggregates
- [ ] No EventBus implementation in Phase 1
- [ ] Events can be retrieved for audit purposes
- [ ] Clear path to add publishing in Phase 2
- [ ] No event publishing infrastructure in Phase 1

## Consequences

- **Limited Event Usage**: Events stored but not published
- **Future Work**: Event publishing needs to be added later
- **Audit Trail**: Events available for audit but not real-time

## Mitigation

- **Event Storage**: Events are stored with aggregates for audit
- **Clear Interface**: Event publishing interface defined for Phase 2
- **Documentation**: Clear path to add event publishing
- **Incremental Value**: Focus on core domain value first
