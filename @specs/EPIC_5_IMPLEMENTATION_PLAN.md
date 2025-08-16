# EPIC-5: Test Management & Event Logging Implementation Plan
## Test-Driven Development Approach

### Phase 1: Test Management Foundation + Tests (High Priority)

#### Phase 1A: Create Test Service Foundation
- [ ] Create internal/tests package
- [ ] Define TestService struct with Storage injection
- [ ] Implement test state validation logic with updated state machine
- [ ] Create test transition validation rules (pending→wip, wip→passed/failed/cancelled, passed↔failed)
- [ ] Test state constraint enforcement
- [ ] Event creation utilities for test transitions
- [ ] Timestamp handling utilities (current time vs --time flag)

#### Phase 1B: Write Test Service Foundation Tests **IMMEDIATELY AFTER 1A**
Epic 5 Test Scenarios Covered:
- [ ] **Test: Start test** (Epic 5 line 330)
- [ ] **Test: Pass test** (Epic 5 line 335)
- [ ] **Test: Fail test with details** (Epic 5 line 340)
- [ ] **Test: Cancel test with reason** (Epic 5 line 345)
- [ ] **Test: Test state validation logic with updated transitions**
- [ ] **Test: Test can transition from passed to failed and vice versa**
- [ ] **Test: Event creation for test transitions**
- [ ] **Test: Timestamp handling and injection**

#### Phase 1C: Test Commands Implementation
- [ ] Create cmd/start_test.go command
- [ ] Create cmd/pass_test.go command
- [ ] Create cmd/fail_test.go command
- [ ] Create cmd/cancel_test.go command
- [ ] Integrate with TestService for test management
- [ ] Test status transitions with enhanced state machine
- [ ] --time flag support for deterministic testing
- [ ] Simple confirmation output format (non-XML)
- [ ] Failure reason and cancellation reason handling

#### Phase 1D: Write Test Commands Tests **IMMEDIATELY AFTER 1C**
Epic 5 Test Scenarios Covered:
- [ ] **Test: Start test** (detailed workflow)
- [ ] **Test: Pass test** (with automatic event logging)
- [ ] **Test: Fail test with details** (failure note recording)
- [ ] **Test: Cancel test with reason** (cancellation reason recording)
- [ ] **Test: Simple confirmation output format**
- [ ] **Test: Test transitions from passed to failed**
- [ ] **Test: Test transitions from failed to passed**
- [ ] **Test: Timestamp injection via --time flag**

### Phase 2: Event Logging System + Tests (High Priority)

#### Phase 2A: Event Logging Infrastructure
- [ ] Enhanced event structure with metadata support
- [ ] Create cmd/log.go command for manual event logging
- [ ] Event type categorization (implementation, blocker, file_change, milestone)
- [ ] File change tracking and parsing logic
- [ ] Event validation and error handling
- [ ] Integration with current phase/task context from Epic 4
- [ ] XML output formatting for event logging confirmation

#### Phase 2B: Write Event Logging Tests **IMMEDIATELY AFTER 2A**
Epic 5 Test Scenarios Covered:
- [ ] **Test: Log implementation event** (Epic 5 line 350)
- [ ] **Test: Log event with file changes** (Epic 5 line 355)
- [ ] **Test: Log blocker event** (Epic 5 line 360)
- [ ] **Test: File change format validation** (Epic 5 line 375)
- [ ] **Test: Event type categorization**
- [ ] **Test: File change parsing with multiple files**
- [ ] **Test: Event validation and error handling**
- [ ] **Test: XML output format for event logging**

#### Phase 2C: Rich Event Structure & File Tracking
- [ ] Implement rich event XML structure with metadata
- [ ] File change parsing for multiple actions (added, modified, deleted, renamed)
- [ ] Event content formatting with structured information
- [ ] File change metadata integration
- [ ] Event content validation and sanitization
- [ ] Error handling for malformed file change syntax

#### Phase 2D: Write Rich Event Tests **IMMEDIATELY AFTER 2C**
Epic 5 Test Scenarios Covered:
- [ ] **Test: Rich event XML structure**
- [ ] **Test: File change parsing for all action types**
- [ ] **Test: Multiple file changes in single event**
- [ ] **Test: Event content formatting and validation**
- [ ] **Test: File change metadata accuracy**
- [ ] **Test: Error handling for malformed file syntax**

### Phase 3: Blocker Detection & Automatic Events + Tests (Medium Priority)

#### Phase 3A: Automatic Blocker Creation
- [ ] Automatic blocker event creation for failed tests
- [ ] Blocker event formatting with test details
- [ ] Integration between test failures and blocker generation
- [ ] Manual blocker event logging via `--type=blocker`
- [ ] Blocker impact assessment and suggestion generation
- [ ] Blocker event XML structure and metadata

#### Phase 3B: Write Blocker Detection Tests **IMMEDIATELY AFTER 3A**
Epic 5 Test Scenarios Covered:
- [ ] **Test: Automatic blocker for failed test** (Epic 5 line 365)
- [ ] **Test: Manual blocker event creation**
- [ ] **Test: Blocker event formatting with test details**
- [ ] **Test: Blocker impact assessment**
- [ ] **Test: Blocker event XML structure**
- [ ] **Test: Integration between test failures and blockers**

#### Phase 3C: Event Context Integration
- [ ] Event integration with active context from Epic 4
- [ ] Phase and task context preservation in events
- [ ] Automatic context detection and assignment
- [ ] Context validation and error handling
- [ ] Event querying enhancement for Epic 2 integration
- [ ] Context consistency across all event types

#### Phase 3D: Write Event Context Tests **IMMEDIATELY AFTER 3C**
Epic 5 Test Scenarios Covered:
- [ ] **Test: Event integration with active context** (Epic 5 line 370)
- [ ] **Test: Phase and task context preservation**
- [ ] **Test: Automatic context detection**
- [ ] **Test: Context validation and error handling**
- [ ] **Test: Event querying integration**
- [ ] **Test: Context consistency across event types**

### Phase 4: Integration & Enhanced Features + Tests (Low Priority)

#### Phase 4A: Epic 2 Integration & Enhancements
- [ ] Integration with Epic 2 failing command enhancements
- [ ] Enhanced failing tests display with failure_note details
- [ ] Test status updates integration with existing progress calculation
- [ ] Blocker events integration with handoff reporting
- [ ] Cross-command consistency for test-related queries
- [ ] Performance optimization for test and event queries

#### Phase 4B: Write Integration Tests **IMMEDIATELY AFTER 4A**
Epic 5 Test Scenarios Covered:
- [ ] **Test: Integration with Epic 2 failing command**
- [ ] **Test: Enhanced failing tests display**
- [ ] **Test: Test status integration with progress calculation**
- [ ] **Test: Blocker events in handoff reporting**
- [ ] **Test: Cross-command consistency**
- [ ] **Test: Performance with large event histories**

#### Phase 4C: Final Integration & Performance
- [ ] End-to-end integration testing across all Epic 5 commands
- [ ] Performance optimization for large event histories
- [ ] Cross-command consistency (error formats, XML output)
- [ ] Global flag consistency (--time, --file)
- [ ] Help system integration for all test and event commands
- [ ] Final error message refinement and user experience

#### Phase 4D: Write Final Integration Tests **IMMEDIATELY AFTER 4C**
Epic 5 Test Scenarios Covered:
- [ ] **Test: End-to-end test and event workflow**
- [ ] **Test: Performance with large event datasets**
- [ ] **Test: Cross-command consistency**
- [ ] **Test: Global flag handling**
- [ ] **Test: Help system completeness**
- [ ] **Test: Error message consistency and clarity**
- [ ] **Test: Integration with all previous epics**

## Development Workflow Per Phase

For **EACH** phase:

1. **Implement Code** (Phase XA or XC)
2. **Write Tests IMMEDIATELY** (Phase XB or XD) 
3. **Run Tests & Verify** - All tests must pass
4. **Run Linting/Type Checking** - Code must be clean
5. **NEVER move to next phase with failing tests**

## Epic 5 Specific Considerations

### Dependencies & Integration
- **Epic 1:** CLI framework, XML processing, storage interface
- **Epic 2:** Query service for test status queries (enhancement needed for failure details)
- **Epic 3:** Epic lifecycle for overall epic state management
- **Epic 4:** Task/phase management for event context integration
- **Event System:** Enhanced event logging with rich metadata

### Technical Requirements
- **Enhanced Test States:** Support for passed↔failed transitions (updated state machine)
- **Rich Event Logging:** Detailed events with types, metadata, and file tracking
- **Automatic Blockers:** Failed tests create automatic blocker events
- **Context Integration:** Events include active phase/task context from Epic 4
- **File Tracking:** Parse and validate file changes with actions

### Updated Test State Machine
```go
type TestStatus string

const (
    TestPending   TestStatus = "pending"
    TestWIP      TestStatus = "wip"
    TestPassed   TestStatus = "passed"
    TestFailed   TestStatus = "failed"
    TestCancelled TestStatus = "cancelled"
)

func (ts TestStatus) CanTransitionTo(target TestStatus) bool {
    transitions := map[TestStatus][]TestStatus{
        TestPending:   {TestWIP},
        TestWIP:      {TestPassed, TestFailed, TestCancelled},
        TestPassed:   {TestFailed}, // New: can transition from passed to failed
        TestFailed:   {TestPassed}, // New: can transition from failed to passed
        TestCancelled: {},          // No transitions from cancelled
    }
    
    for _, allowed := range transitions[ts] {
        if allowed == target {
            return true
        }
    }
    return false
}
```

### File Structure
```
├── cmd/
│   ├── start_test.go       # Test start command
│   ├── pass_test.go        # Test pass command
│   ├── fail_test.go        # Test fail command
│   ├── cancel_test.go      # Test cancel command
│   └── log.go             # Event logging command
├── internal/
│   └── tests/             # Test management and event logging
│       ├── service.go     # TestService with DI
│       ├── states.go      # Test state management with enhanced transitions
│       ├── events.go      # Event creation and logging
│       ├── blockers.go    # Automatic blocker detection
│       └── files.go       # File change parsing and validation
└── testdata/
    ├── epic-test-states.xml    # Epic with tests in various states
    ├── epic-rich-events.xml    # Epic with rich event history
    ├── epic-failed-tests.xml   # Epic with failed tests for blocker testing
    └── epic-file-changes.xml   # Epic with file change events
```

### Event Types & File Change Support
```go
type EventType string

const (
    EventTestStarted     EventType = "test_started"
    EventTestPassed     EventType = "test_passed"
    EventTestFailed     EventType = "test_failed"
    EventTestCancelled  EventType = "test_cancelled"
    EventImplementation EventType = "implementation"
    EventBlocker        EventType = "blocker"
    EventFileChange     EventType = "file_change"
    EventMilestone      EventType = "milestone"
)

type FileChange struct {
    Path   string
    Action string // added, modified, deleted, renamed
}

func ParseFileChanges(filesFlag string) ([]FileChange, error) {
    // Parse format: "file1.js:added,file2.js:modified"
    // Support multiple files and validate actions
}
```

## Benefits of This Approach

✅ **Immediate Feedback** - Catch issues as soon as code is written  
✅ **Working Code** - Each phase delivers tested, working functionality  
✅ **Epic 5 Coverage** - All acceptance criteria distributed across phases  
✅ **Incremental Progress** - Agents can use test commands after each phase  
✅ **Risk Mitigation** - Problems caught early, not at the end  
✅ **Quality Assurance** - No untested code makes it to later phases  

## Test Distribution Summary

- **Phase 1 Tests:** 16 scenarios (Test management foundation, enhanced state machine)
- **Phase 2 Tests:** 16 scenarios (Event logging system, rich event structure)
- **Phase 3 Tests:** 12 scenarios (Blocker detection, event context integration)
- **Phase 4 Tests:** 14 scenarios (Epic 2 integration, final integration testing)

**Total: All Epic 5 acceptance criteria and test scenarios covered across all phases**

---

## Implementation Status

### EPIC 5: TEST MANAGEMENT & EVENT LOGGING - PENDING
### Current Status: READY TO START (After Epic 1, 2, 3 & 4 Complete)

### Progress Tracking
- [ ] Phase 1A: Create Test Service Foundation
- [ ] Phase 1B: Write Test Service Foundation Tests
- [ ] Phase 1C: Test Commands Implementation
- [ ] Phase 1D: Write Test Commands Tests
- [ ] Phase 2A: Event Logging Infrastructure
- [ ] Phase 2B: Write Event Logging Tests
- [ ] Phase 2C: Rich Event Structure & File Tracking
- [ ] Phase 2D: Write Rich Event Tests
- [ ] Phase 3A: Automatic Blocker Creation
- [ ] Phase 3B: Write Blocker Detection Tests
- [ ] Phase 3C: Event Context Integration
- [ ] Phase 3D: Write Event Context Tests
- [ ] Phase 4A: Epic 2 Integration & Enhancements
- [ ] Phase 4B: Write Integration Tests
- [ ] Phase 4C: Final Integration & Performance
- [ ] Phase 4D: Write Final Integration Tests

### Definition of Done
- [ ] All acceptance criteria verified with automated tests
- [ ] Test management commands execute in < 100ms for typical epic files
- [ ] Test coverage > 90% for test management and event logging
- [ ] Event logging supports all specified types and metadata
- [ ] File change tracking parses all valid formats correctly
- [ ] Automatic blocker creation works for failed tests
- [ ] Integration with Epic 2 query commands works seamlessly
- [ ] Performance requirements met for large event histories
- [ ] Enhanced test state machine supports passed↔failed transitions