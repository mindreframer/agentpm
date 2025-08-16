# EPIC-4: Test Management & Event Logging Implementation Plan
## Test-Driven Development Approach

### Phase 1: Test Management Foundation + Tests (High Priority)

#### Phase 1A: Create Test Service Foundation
- [ ] Create internal/tests package
- [ ] Define TestService struct with Storage injection
- [ ] Implement test status validation logic (pending/wip/passed/failed/cancelled state machine)
- [ ] Create test state transition validation rules
- [ ] Test prerequisite validation (associated task/phase must be active or completed)
- [ ] Timestamp handling utilities for test lifecycle events
- [ ] Basic test lookup and status management operations

#### Phase 1B: Write Test Service Tests **IMMEDIATELY AFTER 1A**
Epic 4 Test Scenarios Covered:
- [ ] **Test: Start test from pending status** (AC-1)
- [ ] **Test: Pass test from wip status** (AC-2)
- [ ] **Test: Fail test with failure details** (AC-3)
- [ ] **Test: Cancel test with cancellation reason** (AC-4)
- [ ] **Test: Test state transition validation**
- [ ] **Test: Test prerequisite checking**
- [ ] **Test: Timestamp handling and injection**

#### Phase 1C: Test Status Management Commands
- [ ] Create cmd/start_test.go command
- [ ] Create cmd/pass_test.go command
- [ ] Create cmd/fail_test.go command
- [ ] Create cmd/cancel_test.go command
- [ ] Integrate with TestService for all operations
- [ ] --time flag support for deterministic testing
- [ ] Simple confirmation output messages
- [ ] Error handling for invalid state transitions

#### Phase 1D: Write Test Commands Tests **IMMEDIATELY AFTER 1C**
Epic 4 Test Scenarios Covered:
- [ ] **Test: start-test command execution** (FR-1)
- [ ] **Test: pass-test command execution** (FR-1)
- [ ] **Test: fail-test command with failure reason** (FR-1)
- [ ] **Test: cancel-test command with cancellation reason** (FR-1)
- [ ] **Test: Command error handling**
- [ ] **Test: Output format validation**
- [ ] **Test: Timestamp injection via --time flag**

### Phase 2: Event Logging System + Tests (High Priority)

#### Phase 2A: Event Logging Foundation
- [ ] Create internal/events package
- [ ] Define EventService struct with Storage injection
- [ ] Implement event creation with rich metadata support
- [ ] Event type categorization (implementation, blocker, file_change, milestone)
- [ ] File change parsing and validation
- [ ] Phase/task context association
- [ ] Event timestamp management

#### Phase 2B: Write Event Service Tests **IMMEDIATELY AFTER 2A**
Epic 4 Test Scenarios Covered:
- [ ] **Test: Event creation with metadata** (FR-2)
- [ ] **Test: Event type categorization** (FR-2)
- [ ] **Test: File change format parsing** (AC-6, AC-10)
- [ ] **Test: Phase/task context association** (AC-9)
- [ ] **Test: Event timestamp handling**
- [ ] **Test: Event validation logic**

#### Phase 2C: Event Logging Command Implementation
- [ ] Create cmd/log.go command
- [ ] Integrate with EventService for event creation
- [ ] --type flag support for event categorization
- [ ] --files flag support for file change tracking
- [ ] --time flag support for deterministic testing
- [ ] XML output formatting for event confirmation
- [ ] Error handling for invalid file change formats

#### Phase 2D: Write Event Command Tests **IMMEDIATELY AFTER 2C**
Epic 4 Test Scenarios Covered:
- [ ] **Test: Log implementation event** (AC-5)
- [ ] **Test: Log event with file changes** (AC-6)
- [ ] **Test: Log blocker event** (AC-7)
- [ ] **Test: File change format validation** (AC-10)
- [ ] **Test: Event command error handling**
- [ ] **Test: XML output format validation**

### Phase 3: Automatic Event Integration + Tests (Medium Priority)

#### Phase 3A: Automatic Event Creation
- [ ] Integrate EventService with TestService
- [ ] Automatic event creation for test state transitions
- [ ] test_started, test_passed, test_failed, test_cancelled event types
- [ ] Rich event metadata for test operations
- [ ] Test failure details in event messages
- [ ] Event-test relationship tracking

#### Phase 3B: Write Automatic Event Tests **IMMEDIATELY AFTER 3A**
Epic 4 Test Scenarios Covered:
- [ ] **Test: Automatic test_started event creation**
- [ ] **Test: Automatic test_passed event creation**
- [ ] **Test: Automatic test_failed event creation**
- [ ] **Test: Automatic test_cancelled event creation**
- [ ] **Test: Event metadata accuracy**
- [ ] **Test: Event-test relationship tracking**

#### Phase 3C: Blocker Detection & Management
- [ ] Automatic blocker event creation for failed tests
- [ ] Blocker impact assessment logic
- [ ] Manual blocker event support via log command
- [ ] Blocker event format standardization
- [ ] Integration with test failure reporting

#### Phase 3D: Write Blocker Detection Tests **IMMEDIATELY AFTER 3C**
Epic 4 Test Scenarios Covered:
- [ ] **Test: Automatic blocker for failed test** (AC-8)
- [ ] **Test: Manual blocker event creation** (AC-7)
- [ ] **Test: Blocker impact assessment**
- [ ] **Test: Blocker event format validation**
- [ ] **Test: Integration with test failure reporting**

### Phase 4: Enhanced Data Model & Integration + Tests (Low Priority)

#### Phase 4A: Enhanced Event & Test Data Model
- [ ] Rich event XML structure with metadata
- [ ] Enhanced test XML structure with timestamps
- [ ] File change metadata in events
- [ ] Test failure note tracking
- [ ] Event-phase-task-test relationship modeling
- [ ] XML schema validation for new structures

#### Phase 4B: Write Data Model Tests **IMMEDIATELY AFTER 4A**
Epic 4 Test Scenarios Covered:
- [ ] **Test: Rich event XML structure** (FR-3)
- [ ] **Test: Enhanced test XML structure** (FR-4)
- [ ] **Test: File change metadata parsing**
- [ ] **Test: Test failure note storage**
- [ ] **Test: Event relationship modeling**
- [ ] **Test: XML schema validation**

#### Phase 4C: Epic 2 Integration & Performance
- [ ] Integration with Epic 2 failing command enhancements
- [ ] Event querying optimization for large event histories
- [ ] Performance optimization for test management operations
- [ ] Memory usage optimization for event processing
- [ ] Cross-command consistency verification

#### Phase 4D: Write Integration & Performance Tests **IMMEDIATELY AFTER 4C**
Epic 4 Test Scenarios Covered:
- [ ] **Test: Epic 2 integration functionality** (FR-4)
- [ ] **Test: Performance with large event histories** (NFR-1)
- [ ] **Test: Memory usage optimization**
- [ ] **Test: Cross-command consistency**
- [ ] **Test: Event querying performance**
- [ ] **Test: Test management command performance**

## Development Workflow Per Phase

For **EACH** phase:

1. **Implement Code** (Phase XA or XC)
2. **Write Tests IMMEDIATELY** (Phase XB or XD) 
3. **Run Tests & Verify** - All tests must pass
4. **Run Linting/Type Checking** - Code must be clean
5. **NEVER move to next phase with failing tests**

## Epic 4 Specific Considerations

### Dependencies & Integration
- **Epic 1:** CLI framework, XML processing, storage interface
- **Epic 2:** Query service for test status queries and failing command enhancement
- **Epic 3:** Epic lifecycle for phase/task context and validation
- **Testing:** Deterministic timestamp support for snapshot testing

### Technical Requirements
- **Test State Machine:** pending → wip → passed/failed/cancelled transitions
- **Event Types:** test_started, test_passed, test_failed, test_cancelled, implementation, blocker, file_change, milestone
- **Rich Metadata:** File changes, phase/task context, timestamps, failure details
- **Automatic Integration:** Test transitions automatically create events
- **Performance:** < 100ms for test operations, efficient event processing

### File Structure
```
├── cmd/
│   ├── start_test.go       # Test start command
│   ├── pass_test.go        # Test pass command
│   ├── fail_test.go        # Test fail command
│   ├── cancel_test.go      # Test cancel command
│   └── log.go              # Event logging command
├── internal/
│   ├── tests/              # Test management service
│   │   ├── service.go      # TestService with DI
│   │   ├── validation.go   # Test state validation
│   │   └── transitions.go  # Test state transitions
│   └── events/             # Event logging service
│       ├── service.go      # EventService with DI
│       ├── types.go        # Event types and metadata
│       ├── parsing.go      # File change parsing
│       └── integration.go  # Auto event creation
└── testdata/
    ├── test-pending.xml    # Test in pending state
    ├── test-wip.xml        # Test in progress
    ├── test-failed.xml     # Test with failure details
    └── events-rich.xml     # Epic with rich event history
```

### Test State Machine Implementation
```go
type TestStatus string

const (
    TestStatusPending   TestStatus = "pending"
    TestStatusWIP      TestStatus = "wip"
    TestStatusPassed   TestStatus = "passed"
    TestStatusFailed   TestStatus = "failed"
    TestStatusCancelled TestStatus = "cancelled"
)

func (s TestStatus) CanTransitionTo(target TestStatus) bool {
    transitions := map[TestStatus][]TestStatus{
        TestStatusPending:   {TestStatusWIP},
        TestStatusWIP:      {TestStatusPassed, TestStatusFailed, TestStatusCancelled},
        TestStatusPassed:   {TestStatusFailed},  // Allow re-failing passed tests
        TestStatusFailed:   {TestStatusPassed}, // Allow fixing failed tests
        TestStatusCancelled: {}, // No transitions from cancelled
    }
    
    for _, allowed := range transitions[s] {
        if allowed == target {
            return true
        }
    }
    return false
}
```

### Event Type Implementation
```go
type EventType string

const (
    EventTypeTestStarted      EventType = "test_started"
    EventTypeTestPassed      EventType = "test_passed"
    EventTypeTestFailed      EventType = "test_failed"
    EventTypeTestCancelled   EventType = "test_cancelled"
    EventTypeImplementation  EventType = "implementation"
    EventTypeBlocker         EventType = "blocker"
    EventTypeFileChange      EventType = "file_change"
    EventTypeMilestone       EventType = "milestone"
)
```

## Benefits of This Approach

✅ **Immediate Feedback** - Catch issues as soon as code is written  
✅ **Working Code** - Each phase delivers tested, working functionality  
✅ **Epic 4 Coverage** - All acceptance criteria distributed across phases  
✅ **Incremental Progress** - Agents can use test management after each phase  
✅ **Risk Mitigation** - Problems caught early, not at the end  
✅ **Quality Assurance** - No untested code makes it to later phases  

## Test Distribution Summary

- **Phase 1 Tests:** 14 scenarios (Test service foundation, test commands)
- **Phase 2 Tests:** 12 scenarios (Event logging foundation, event commands)
- **Phase 3 Tests:** 12 scenarios (Automatic events, blocker detection)
- **Phase 4 Tests:** 12 scenarios (Data model, integration, performance)

**Total: All Epic 4 acceptance criteria and test scenarios covered across all phases**

---

## Implementation Status

### EPIC 4: TEST MANAGEMENT & EVENT LOGGING - PENDING
### Current Status: READY FOR IMPLEMENTATION

### Progress Tracking
- [ ] Phase 1A: Create Test Service Foundation
- [ ] Phase 1B: Write Test Service Tests
- [ ] Phase 1C: Test Status Management Commands
- [ ] Phase 1D: Write Test Commands Tests
- [ ] Phase 2A: Event Logging Foundation
- [ ] Phase 2B: Write Event Service Tests
- [ ] Phase 2C: Event Logging Command Implementation
- [ ] Phase 2D: Write Event Command Tests
- [ ] Phase 3A: Automatic Event Creation
- [ ] Phase 3B: Write Automatic Event Tests
- [ ] Phase 3C: Blocker Detection & Management
- [ ] Phase 3D: Write Blocker Detection Tests
- [ ] Phase 4A: Enhanced Event & Test Data Model
- [ ] Phase 4B: Write Data Model Tests
- [ ] Phase 4C: Epic 2 Integration & Performance
- [ ] Phase 4D: Write Integration & Performance Tests

### Definition of Done
- [ ] All acceptance criteria verified with automated tests
- [ ] Test management commands execute in < 100ms for typical epic files
- [ ] Test coverage > 90% for test management and event logging
- [ ] Event logging supports all specified types and metadata
- [ ] File change tracking parses all valid formats correctly
- [ ] Automatic blocker creation works for failed tests
- [ ] Integration with Epic 2 query commands works seamlessly
- [ ] Performance requirements met for large event histories