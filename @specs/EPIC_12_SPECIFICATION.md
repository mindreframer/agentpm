# EPIC-12 SPECIFICATION: Complete Event Logging for All XML Modifications

## Overview

**Epic ID:** 12  
**Name:** Complete Event Logging for All XML Modifications  
**Duration:** 1-2 days  
**Status:** pending  
**Priority:** high  

**Goal:** Ensure all actions that modify the XML file generate proper events for complete audit trail and observability of all state changes in the system.

## Business Context

Currently, the system has inconsistent event logging where task and phase operations properly log events, but test operations and epic-level operations completely lack event logging. This creates gaps in the audit trail and makes it difficult to track the complete history of project changes. This epic addresses these gaps to ensure comprehensive event logging for all XML modifications.

## User Stories

### Primary User Stories
- **As a project manager, I can see all test-related events in the timeline** so that I have complete visibility into testing progress
- **As an agent, I can track when tests are started, passed, failed, or cancelled** so that I understand the complete testing history
- **As a system administrator, I can audit all changes made to the project** so that I have complete traceability of all modifications
- **As a developer, I can see when epics are started and completed** so that I understand high-level project milestones

### Secondary User Stories
- **As an agent, I can see event timestamps for all operations** so that I understand the chronological sequence of all activities
- **As a project manager, I can generate reports with complete event history** so that I have comprehensive project tracking
- **As a system user, I can rely on consistent event logging behavior** so that all operations provide the same level of audit trail

## Technical Requirements

### Core Dependencies
- **Event Service:** Extend `internal/service/events.go` with new event types
- **Test Service:** Modify `internal/tests/service.go` to add event logging
- **Epic Service:** Add event logging to epic operations
- **Query Service:** Ensure events are properly retrievable through existing queries

### Event Type Extensions
- **Test Events:** Add comprehensive test-related event types
- **Epic Events:** Add epic-level milestone event types
- **Consistency:** Ensure all event types follow existing patterns

## Functional Requirements

### FR-1: Test Event Logging
**Operations:** `start-test`, `pass-test`, `fail-test`, `cancel-test`

**Behavior:**
- All test operations generate events with proper timestamps
- Events include test ID, operation type, and relevant context
- Failed tests include failure reason in event data
- Cancelled tests include cancellation reason in event data

**Event Types to Add:**
```go
const (
    EventTestStarted    EventType = "test_started"
    EventTestPassed     EventType = "test_passed"
    EventTestFailed     EventType = "test_failed"
    EventTestCancelled  EventType = "test_cancelled"
)
```

**Example Event Data:**
```xml
<event id="test_started_1629123456" type="test_started" timestamp="2025-08-16T15:30:45Z">
    <data>Test T1A_1 (Test Project Init) started</data>
</event>

<event id="test_failed_1629123567" type="test_failed" timestamp="2025-08-16T15:35:45Z">
    <data>Test T1A_2 (Test Dependency Resolution) failed: Dependency conflict detected</data>
</event>
```

### FR-2: Epic Event Logging
**Operations:** `start-epic`, `done-epic`

**Behavior:**
- Epic start and completion generate milestone events
- Events include epic metadata and timing information
- Epic completion events include summary information

**Event Types to Add:**
```go
const (
    EventEpicStarted    EventType = "epic_started"
    EventEpicCompleted  EventType = "epic_completed"
)
```

**Example Event Data:**
```xml
<event id="epic_started_1629123456" type="epic_started" timestamp="2025-08-16T14:00:00Z">
    <data>Epic CLI Framework Development started</data>
</event>

<event id="epic_completed_1629123789" type="epic_completed" timestamp="2025-08-18T16:00:00Z">
    <data>Epic CLI Framework Development completed</data>
</event>
```

### FR-3: Enhanced Event Creation Function
**Location:** `internal/service/events.go`

**Behavior:**
- Extend `CreateEvent` function to support new event types
- Add entity validation for test and epic events
- Maintain consistent event data formatting
- Support failure and cancellation reasons in event data

**Enhanced Function Signature:**
```go
func CreateEvent(epicData *epic.Epic, eventType EventType, phaseID, taskID, testID, reason string, timestamp time.Time)
```

### FR-4: Service Integration
**Test Service Integration:**
- Modify all test operations in `internal/tests/service.go`
- Add event creation calls after successful state transitions
- Include failure/cancellation reasons in event data
- Maintain existing error handling behavior

**Epic Service Integration:**
- Add event creation to epic start/completion operations
- Ensure events are created before XML is saved
- Include epic metadata in event descriptions

### FR-5: Backward Compatibility
**Existing Event Behavior:**
- All existing event types continue to work unchanged
- No breaking changes to event data structure
- Existing events timeline remains functional
- Current event query functionality preserved

## Non-Functional Requirements

### NFR-1: Performance
- Event creation adds < 5ms to operation execution time
- Minimal memory overhead for event storage
- Efficient event data serialization

### NFR-2: Reliability
- Event creation failures do not prevent operation completion
- Graceful degradation if event creation fails
- Atomic operations (save succeeds or entire operation fails)

### NFR-3: Consistency
- All event data follows consistent formatting patterns
- Event timestamps use UTC timezone consistently
- Event IDs follow existing generation patterns

### NFR-4: Observability
- All new events are discoverable through existing `events` command
- Events integrate with existing reporting and query mechanisms
- Event data is human-readable and machine-parseable

## Implementation Approach

### Phase 12A: Event Type Definition (Day 1 - 2 hours)
- Add new event type constants to `internal/service/events.go`
- Extend event creation logic for new types
- Add entity validation for test and epic events
- Unit tests for new event creation logic

### Phase 12B: Test Service Integration (Day 1 - 4 hours)
- Modify `StartTest` method to create `EventTestStarted` events
- Modify `PassTest` method to create `EventTestPassed` events
- Modify `FailTest` method to create `EventTestFailed` events with failure reason
- Modify `CancelTest` method to create `EventTestCancelled` events with cancellation reason
- Integration tests for all test event logging

### Phase 12C: Epic Service Integration (Day 1-2 - 2 hours)
- Add event logging to epic start operations
- Add event logging to epic completion operations
- Ensure proper integration with existing epic lifecycle
- Integration tests for epic event logging

### Phase 12D: Testing and Validation (Day 2 - 4 hours)
- Comprehensive test coverage for all new event types
- End-to-end testing of event timeline with all operation types
- Performance testing for event creation overhead
- Validation of event data consistency and format

## Data Model Extensions

### Enhanced Event Creation Parameters
```go
type EventCreationParams struct {
    EventType EventType
    PhaseID   string
    TaskID    string
    TestID    string
    Reason    string
    Timestamp time.Time
}
```

### Event Data Templates
```go
// Test event data templates
func formatTestStartedData(test *epic.Test) string
func formatTestPassedData(test *epic.Test) string
func formatTestFailedData(test *epic.Test, reason string) string
func formatTestCancelledData(test *epic.Test, reason string) string

// Epic event data templates
func formatEpicStartedData(epic *epic.Epic) string
func formatEpicCompletedData(epic *epic.Epic) string
```

## Acceptance Criteria

### AC-1: Test Start Event Logging
- **GIVEN** I have a test in pending status
- **WHEN** I run `agentpm start-test T1A_1`
- **THEN** an event of type "test_started" should be created with proper test information

### AC-2: Test Pass Event Logging
- **GIVEN** I have a test in WIP status
- **WHEN** I run `agentpm pass-test T1A_1`
- **THEN** an event of type "test_passed" should be created with test completion information

### AC-3: Test Fail Event Logging
- **GIVEN** I have a test in WIP status
- **WHEN** I run `agentpm fail-test T1A_1 "Dependency conflict"`
- **THEN** an event of type "test_failed" should be created with failure reason included

### AC-4: Test Cancel Event Logging
- **GIVEN** I have a test in WIP status
- **WHEN** I run `agentpm cancel-test T1A_1 "Requirements changed"`
- **THEN** an event of type "test_cancelled" should be created with cancellation reason included

### AC-5: Epic Start Event Logging
- **GIVEN** I have an epic in pending status
- **WHEN** I run `agentpm start-epic`
- **THEN** an event of type "epic_started" should be created with epic information

### AC-6: Epic Complete Event Logging
- **GIVEN** I have an epic ready for completion
- **WHEN** I run `agentpm done-epic`
- **THEN** an event of type "epic_completed" should be created with completion information

### AC-7: Event Timeline Integration
- **GIVEN** I have performed various operations including test and epic operations
- **WHEN** I run `agentpm events`
- **THEN** I should see all events including the new test and epic events in chronological order

### AC-8: Event Data Consistency
- **GIVEN** I create events of different types
- **WHEN** I examine the event data
- **THEN** all events should follow consistent formatting patterns and include relevant context

### AC-9: Backward Compatibility
- **GIVEN** I have existing events from task and phase operations
- **WHEN** I add new event types and query the timeline
- **THEN** existing events should continue to display correctly alongside new events

### AC-10: Performance Validation
- **GIVEN** I perform operations that create events
- **WHEN** I measure operation execution time
- **THEN** event creation should add less than 5ms to operation completion time

## Testing Strategy

### Test Categories
- **Unit Tests (70%):** Event creation logic, data formatting, entity validation
- **Integration Tests (25%):** Service integration, event persistence, timeline retrieval
- **Performance Tests (5%):** Event creation overhead, memory usage validation

### Test Data Requirements
- **Test Scenarios:** All test state transitions with various failure/cancellation reasons
- **Epic Scenarios:** Epic start and completion with different epic configurations
- **Mixed Operations:** Combined task, phase, test, and epic operations for timeline validation

### Validation Testing
- **Event Data Format:** Consistent formatting across all event types
- **Timeline Integration:** Proper chronological ordering and display
- **Error Handling:** Graceful handling of event creation failures

## Definition of Done

- [ ] All new event types defined and implemented in events service
- [ ] Test service operations create appropriate events with proper data
- [ ] Epic operations create milestone events with relevant context
- [ ] All events integrate properly with existing events command and timeline
- [ ] Event data follows consistent formatting patterns
- [ ] Performance impact is within acceptable limits (< 5ms overhead)
- [ ] Test coverage > 90% for new event creation functionality
- [ ] Backward compatibility maintained for existing events
- [ ] Integration tests validate complete event logging workflow
- [ ] Documentation updated with new event types and examples

## Dependencies and Risks

### Dependencies
- **Epic 1:** Foundation CLI structure (done)
- **Existing Event System:** Current events service implementation
- **Test Service:** Current test operations implementation
- **Epic Service:** Current epic operations implementation

### Risks
- **Low Risk:** Performance impact from additional event creation calls
- **Low Risk:** Event creation failures affecting operation completion
- **Medium Risk:** Event data consistency across different operation types

### Mitigation Strategies
- Implement efficient event creation with minimal overhead
- Add proper error handling to prevent operation failures
- Create comprehensive test suite to ensure event data consistency
- Add performance monitoring for event creation operations

## Notes

- This epic ensures complete audit trail for all system operations
- Event logging should be implemented as a non-blocking operation where possible
- Consider adding configuration options for event verbosity levels in future
- Event data should be structured to support future analytics and reporting features
- Implementation should follow existing patterns to maintain code consistency