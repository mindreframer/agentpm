# EPIC-2: Query & Status Commands Implementation Plan
## Test-Driven Development Approach

### Phase 1: Query Infrastructure & Progress Calculation + Tests (High Priority)

#### Phase 1A: Create Query Service Foundation
- [ ] Create internal/query package
- [ ] Define QueryService struct with Storage interface injection
- [ ] Implement epic loading and caching for single command execution
- [ ] Create progress calculation algorithms (tasks + tests completion)
- [ ] Phase status determination logic (pending → wip → done)
- [ ] Next action recommendation engine
- [ ] Error handling framework for query operations

#### Phase 1B: Write Query Service Tests **IMMEDIATELY AFTER 1A**
Epic 2 Test Scenarios Covered:
- [ ] **Test: Progress calculation with mixed completion** (Epic 2 line 275)
- [ ] **Test: Progress calculation with completed phases** (Epic 2 line 280)
- [ ] **Test: Phase status determination logic**
- [ ] **Test: Next action recommendations for different states**
- [ ] **Test: Epic loading and caching within single execution**
- [ ] **Test: Query service error handling**

#### Phase 1C: Status Command Implementation
- [ ] Create cmd/status.go command
- [ ] Integrate with QueryService for epic status retrieval
- [ ] XML output formatting for status response
- [ ] File override support (-f flag) integration
- [ ] Current phase/task identification logic
- [ ] Completion percentage calculation and display

#### Phase 1D: Write Status Command Tests **IMMEDIATELY AFTER 1C**
Epic 2 Test Scenarios Covered:
- [ ] **Test: Show epic status with progress** (Epic 2 line 67)
- [ ] **Test: Show epic status for completed epic** (Epic 2 line 72)
- [ ] **Test: Show epic status with failing tests** (Epic 2 line 77)
- [ ] **Test: Status command with file override**
- [ ] **Test: Status command error handling**
- [ ] **Test: Status XML output format validation**

### Phase 2: Current State & Pending Work Commands + Tests (High Priority)

#### Phase 2A: Current & Pending Commands Implementation
- [ ] Create cmd/current.go command
- [ ] Create cmd/pending.go command
- [ ] Active work identification logic (current phase/task)
- [ ] Next action recommendation based on epic state
- [ ] Pending work collection across all phases
- [ ] Work grouping by type (phases, tasks, tests)
- [ ] XML output formatting for both commands

#### Phase 2B: Write Current & Pending Tests **IMMEDIATELY AFTER 2A**
Epic 2 Test Scenarios Covered:
- [ ] **Test: Show current active work** (Epic 2 line 82)
- [ ] **Test: Show current state with no active work** (Epic 2 line 87)
- [ ] **Test: Show next action recommendation** (Epic 2 line 92)
- [ ] **Test: List pending tasks across phases** (Epic 2 line 100)
- [ ] **Test: Show pending tasks when all completed** (Epic 2 line 106)
- [ ] **Test: Pending work grouping and organization**

#### Phase 2C: Enhanced Query Logic & Data Processing
- [ ] Epic data validation and consistency checking
- [ ] Graceful handling of incomplete epic data
- [ ] Missing data detection and warnings
- [ ] Phase/task relationship validation
- [ ] Test association with tasks and phases
- [ ] Progress calculation edge cases handling

#### Phase 2D: Write Enhanced Query Tests **IMMEDIATELY AFTER 2C**
Epic 2 Test Scenarios Covered:
- [ ] **Test: Handle incomplete epic data gracefully**
- [ ] **Test: Progress calculation edge cases**
- [ ] **Test: Missing data warning generation**
- [ ] **Test: Phase/task relationship validation**
- [ ] **Test: Epic data consistency checking**

### Phase 3: Failing Tests & Events Commands + Tests (Medium Priority)

#### Phase 3A: Failing Tests Command Implementation
- [ ] Create cmd/failing.go command
- [ ] Test status filtering logic (only "failing" status)
- [ ] Failure details extraction and formatting
- [ ] Phase grouping for failing tests
- [ ] Empty state handling when no tests are failing
- [ ] XML output formatting for failing tests

#### Phase 3B: Write Failing Tests Command Tests **IMMEDIATELY AFTER 3A**
Epic 2 Test Scenarios Covered:
- [ ] **Test: Show only failing tests with details** (Epic 2 line 113)
- [ ] **Test: Show failing tests when all passing** (Epic 2 line 118)
- [ ] **Test: Failing tests grouped by phase**
- [ ] **Test: Failure details extraction and display**
- [ ] **Test: Empty failing tests list handling**

#### Phase 3C: Events Command Implementation
- [ ] Create cmd/events.go command
- [ ] Event timeline processing (reverse chronological order)
- [ ] Configurable limit support (--limit flag, default: 10)
- [ ] Event metadata extraction (timestamp, agent, phase, type)
- [ ] Event content formatting and display
- [ ] Empty event history handling
- [ ] XML output formatting for events

#### Phase 3D: Write Events Command Tests **IMMEDIATELY AFTER 3C**
Epic 2 Test Scenarios Covered:
- [ ] **Test: Show recent events with limit** (Epic 2 line 126)
- [ ] **Test: Show events in chronological order** (Epic 2 line 131)
- [ ] **Test: Event limit parameter handling**
- [ ] **Test: Event metadata extraction**
- [ ] **Test: Empty event history handling**
- [ ] **Test: Events XML output format**

### Phase 4: Integration & Performance Optimization + Tests (Low Priority)

#### Phase 4A: Command Integration & File Override
- [ ] Ensure all commands support -f flag consistently
- [ ] Global flag handling standardization
- [ ] Command error handling consistency
- [ ] XML output format standardization across commands
- [ ] Help system integration for all new commands
- [ ] Configuration integration (default epic file)

#### Phase 4B: Write Integration Tests **IMMEDIATELY AFTER 4A**
Epic 2 Test Scenarios Covered:
- [ ] **Test: File override support for all commands** (Epic 2 line 266)
- [ ] **Test: Global flag consistency**
- [ ] **Test: Help system completeness**
- [ ] **Test: Configuration integration**
- [ ] **Test: Command error handling consistency**

#### Phase 4C: Performance & Polish
- [ ] Query performance measurement and optimization
- [ ] Memory usage optimization for epic loading
- [ ] XML parsing efficiency improvements
- [ ] Command execution time optimization
- [ ] Error message clarity and consistency
- [ ] Code cleanup and documentation

#### Phase 4D: Final Testing & Validation **IMMEDIATELY AFTER 4C**
Epic 2 Test Scenarios Covered:
- [ ] **Test: Performance requirements met** (Epic 2 line 295)
- [ ] **Test: All commands execute within time limits**
- [ ] **Test: Memory usage within acceptable bounds**
- [ ] **Test: End-to-end query workflows**
- [ ] **Test: Cross-command consistency**
- [ ] **Test: Error handling robustness**

## Development Workflow Per Phase

For **EACH** phase:

1. **Implement Code** (Phase XA or XC)
2. **Write Tests IMMEDIATELY** (Phase XB or XD) 
3. **Run Tests & Verify** - All tests must pass
4. **Run Linting/Type Checking** - Code must be clean
5. **NEVER move to next phase with failing tests**

## Epic 2 Specific Considerations

### Dependencies & Integration
- **Epic 1:** CLI framework, XML processing, storage interface, configuration management
- **Epic 1 Storage:** FileStorage and MemoryStorage implementations
- **Epic 1 Config:** Configuration loading and epic file resolution

### Technical Requirements
- **Simple XML Parsing:** Load entire epic XML with etree for straightforward processing
- **Progress Calculation:** Task and test completion percentage algorithms
- **Query Performance:** Commands execute in < 100ms for typical epic files
- **File Override:** All commands support -f flag for multi-epic workflows
- **Error Handling:** Graceful degradation for incomplete epic data

### File Structure
```
├── cmd/
│   ├── status.go           # Epic status overview
│   ├── current.go          # Current active work
│   ├── pending.go          # Pending work listing
│   ├── failing.go          # Failing tests report
│   └── events.go           # Recent events timeline
├── internal/
│   └── query/              # Query service and logic
│       ├── service.go      # QueryService with Storage injection
│       ├── progress.go     # Progress calculation algorithms
│       ├── status.go       # Epic status determination
│       └── events.go       # Event processing utilities
└── testdata/
    ├── epic-in-progress.xml
    ├── epic-completed.xml
    ├── epic-with-failures.xml
    └── epic-empty-events.xml
```

## Benefits of This Approach

✅ **Immediate Feedback** - Catch issues as soon as code is written  
✅ **Working Code** - Each phase delivers tested, working functionality  
✅ **Epic 2 Coverage** - All acceptance criteria distributed across phases  
✅ **Incremental Progress** - Agents can use query commands after each phase  
✅ **Risk Mitigation** - Problems caught early, not at the end  
✅ **Quality Assurance** - No untested code makes it to later phases  

## Test Distribution Summary

- **Phase 1 Tests:** 12 scenarios (Query infrastructure, status command, progress calculation)
- **Phase 2 Tests:** 11 scenarios (Current/pending commands, enhanced query logic)
- **Phase 3 Tests:** 11 scenarios (Failing tests, events timeline, data processing)
- **Phase 4 Tests:** 11 scenarios (Integration, performance, cross-command consistency)

**Total: All Epic 2 acceptance criteria and test scenarios covered across all phases**

---

## Implementation Status

### EPIC 2: QUERY & STATUS COMMANDS - PENDING
### Current Status: READY TO START (After Epic 1 Complete)

### Progress Tracking
- [x] Phase 1A: Create Query Service Foundation
- [x] Phase 1B: Write Query Service Tests  
- [x] Phase 1C: Status Command Implementation
- [ ] Phase 1D: Write Status Command Tests
- [ ] Phase 2A: Current & Pending Commands Implementation
- [ ] Phase 2B: Write Current & Pending Tests
- [ ] Phase 2C: Enhanced Query Logic & Data Processing
- [ ] Phase 2D: Write Enhanced Query Tests
- [ ] Phase 3A: Failing Tests Command Implementation
- [ ] Phase 3B: Write Failing Tests Command Tests
- [ ] Phase 3C: Events Command Implementation
- [ ] Phase 3D: Write Events Command Tests
- [ ] Phase 4A: Command Integration & File Override
- [ ] Phase 4B: Write Integration Tests
- [ ] Phase 4C: Performance & Polish
- [ ] Phase 4D: Final Testing & Validation

### Definition of Done
- [ ] All acceptance criteria verified with automated tests
- [ ] Commands execute in < 100ms for typical epic files
- [ ] Test coverage > 85% for query logic
- [ ] All error cases handled gracefully with clear messages
- [ ] XML output format consistent across all commands
- [ ] File override (-f flag) works for all query commands
- [ ] Simple XML parsing approach implemented
- [ ] Integration tests verify end-to-end query workflows