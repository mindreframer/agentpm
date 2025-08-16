# EPIC-3: Epic Lifecycle Management Implementation Plan
## Test-Driven Development Approach

### Phase 1: Lifecycle Service Foundation + Tests (High Priority)

#### Phase 1A: Create Lifecycle Service Foundation
- [ ] Create internal/lifecycle package
- [ ] Define LifecycleService struct with Storage and Query service injection
- [ ] Implement epic status validation logic (pending/wip/done state machine)
- [ ] Create state transition validation rules
- [ ] Epic completion validation (all phases done, no failing tests)
- [ ] Event creation utilities for lifecycle transitions
- [ ] Timestamp handling utilities (current time vs --time flag)

#### Phase 1B: Write Lifecycle Service Tests **IMMEDIATELY AFTER 1A**
Epic 3 Test Scenarios Covered:
- [ ] **Test: Start epic from planning status** (Epic 3 line 140)
- [ ] **Test: Start epic that is already started** (Epic 3 line 146)
- [ ] **Test: Start epic with invalid status** (Epic 3 line 151)
- [ ] **Test: Epic status validation logic**
- [ ] **Test: State transition rules enforcement**
- [ ] **Test: Event creation for lifecycle changes**
- [ ] **Test: Timestamp handling and injection**

#### Phase 1C: Epic Validation & Completion Logic
- [ ] Implement completion requirements validation
- [ ] Phase completion checking (all phases have status "completed")
- [ ] Test failure checking (no tests with status "failing")
- [ ] Epic summary calculation (phases, tasks, tests counts)
- [ ] Duration calculation between started_at and completed_at
- [ ] Validation error collection and detailed reporting

#### Phase 1D: Write Validation Logic Tests **IMMEDIATELY AFTER 1C**
Epic 3 Test Scenarios Covered:
- [ ] **Test: Complete epic with all work done** (Epic 3 line 180)
- [ ] **Test: Complete epic with pending work** (Epic 3 line 185)
- [ ] **Test: Complete epic with failing tests** (Epic 3 line 190)
- [ ] **Test: Completion requirements validation**
- [ ] **Test: Epic summary calculation accuracy**
- [ ] **Test: Duration calculation correctness**

### Phase 2: Start Epic Command Implementation + Tests (High Priority)

#### Phase 2A: Start Epic Command Implementation
- [ ] Create cmd/start_epic.go command
- [ ] Integrate with LifecycleService for epic startup
- [ ] Status transition from "pending" to "wip"
- [ ] Timestamp setting (started_at field)
- [ ] --time flag support for deterministic testing
- [ ] Automatic event logging for epic startup
- [ ] XML file updates with new status and timestamp
- [ ] Error handling for invalid state transitions

#### Phase 2B: Write Start Epic Command Tests **IMMEDIATELY AFTER 2A**
Epic 3 Test Scenarios Covered:
- [ ] **Test: Start epic from pending status** (detailed workflow)
- [ ] **Test: Start epic that is already started** (error handling)
- [ ] **Test: Start epic with invalid status** (error validation)
- [ ] **Test: Timestamp injection via --time flag**
- [ ] **Test: Automatic event creation**
- [ ] **Test: XML file updates are atomic**
- [ ] **Test: Error messages are clear and actionable**

#### Phase 2C: File Operations & Atomic Updates
- [ ] Implement atomic XML file updates (backup before modify)
- [ ] Rollback mechanism for failed file operations
- [ ] File locking and concurrency safety
- [ ] XML formatting preservation during updates
- [ ] Error handling for file system issues
- [ ] Backup file cleanup after successful operations

#### Phase 2D: Write File Operations Tests **IMMEDIATELY AFTER 2C**
Epic 3 Test Scenarios Covered:
- [ ] **Test: File operations are atomic and safe**
- [ ] **Test: Rollback works on file operation failure**
- [ ] **Test: XML formatting is preserved**
- [ ] **Test: Concurrent access handling**
- [ ] **Test: File system error handling**
- [ ] **Test: Backup cleanup after success**

### Phase 3: Complete Epic Command Implementation + Tests (Medium Priority)

#### Phase 3A: Complete Epic Command Implementation
- [ ] Create cmd/done_epic.go command
- [ ] Integrate with LifecycleService for epic completion
- [ ] Status transition from "wip" to "done"
- [ ] Completion validation (phases done, tests passing)
- [ ] Timestamp setting (completed_at field)
- [ ] Epic summary generation and display
- [ ] Automatic event logging for epic completion
- [ ] Detailed error reporting for incomplete work

#### Phase 3B: Write Complete Epic Command Tests **IMMEDIATELY AFTER 3A**
Epic 3 Test Scenarios Covered:
- [ ] **Test: Complete epic with all work done** (complete workflow)
- [ ] **Test: Complete epic with pending work** (validation error)
- [ ] **Test: Complete epic with failing tests** (validation error)
- [ ] **Test: Epic summary accuracy**
- [ ] **Test: Duration calculation**
- [ ] **Test: Completion event creation**
- [ ] **Test: Detailed error messages for incomplete work**

#### Phase 3C: Enhanced Validation & Error Reporting
- [ ] Detailed pending work reporting (specific phases/tasks)
- [ ] Failing test details in error messages
- [ ] Validation error categorization and formatting
- [ ] Suggestions for resolving validation failures
- [ ] Progress indicators in error messages
- [ ] XML error response formatting consistency

#### Phase 3D: Write Enhanced Validation Tests **IMMEDIATELY AFTER 3C**
Epic 3 Test Scenarios Covered:
- [ ] **Test: Detailed pending work error reporting**
- [ ] **Test: Failing test details in error messages**
- [ ] **Test: Validation error categorization**
- [ ] **Test: Actionable error suggestions**
- [ ] **Test: Progress indicators in errors**
- [ ] **Test: XML error format consistency**

### Phase 4: Switch Epic Command & Integration + Tests (Low Priority)

#### Phase 4A: Switch Epic Command Implementation
- [ ] Create cmd/switch.go command
- [ ] Configuration file reading and updating
- [ ] Epic file existence validation
- [ ] Current epic context switching
- [ ] Previous epic tracking in config
- [ ] Switch event logging (optional)
- [ ] Error handling for missing files and invalid configurations

#### Phase 4B: Write Switch Command Tests **IMMEDIATELY AFTER 4A**
Epic 3 Test Scenarios Covered:
- [ ] **Test: Switch to different epic file** (Epic 3 line 196)
- [ ] **Test: Switch to non-existent epic file** (Epic 3 line 201)
- [ ] **Test: Configuration file updates correctly**
- [ ] **Test: Previous epic tracking**
- [ ] **Test: Epic file validation**
- [ ] **Test: Switch event logging**
- [ ] **Test: Error handling for invalid files**

#### Phase 4C: Integration & Cross-Command Testing
- [ ] Integration between all lifecycle commands
- [ ] End-to-end workflow testing (init → start → complete)
- [ ] Cross-command consistency (error formats, XML output)
- [ ] Configuration integration across commands
- [ ] Help system integration for all lifecycle commands
- [ ] Global flag consistency (--time, --file)

#### Phase 4D: Write Integration Tests **IMMEDIATELY AFTER 4C**
Epic 3 Test Scenarios Covered:
- [ ] **Test: Deterministic timestamp support** (Epic 3 line 249)
- [ ] **Test: End-to-end epic lifecycle workflow**
- [ ] **Test: Cross-command consistency**
- [ ] **Test: Configuration integration**
- [ ] **Test: Help system completeness**
- [ ] **Test: Global flag handling**
- [ ] **Test: Error format consistency**

## Development Workflow Per Phase

For **EACH** phase:

1. **Implement Code** (Phase XA or XC)
2. **Write Tests IMMEDIATELY** (Phase XB or XD) 
3. **Run Tests & Verify** - All tests must pass
4. **Run Linting/Type Checking** - Code must be clean
5. **NEVER move to next phase with failing tests**

## Epic 3 Specific Considerations

### Dependencies & Integration
- **Epic 1:** CLI framework, XML processing, storage interface, configuration management
- **Epic 2:** Query service for epic validation and status checking
- **Testing:** Deterministic timestamp support for snapshot testing

### Technical Requirements
- **State Machine:** Simple pending → wip → done transitions
- **Validation:** Epic completion requires all phases done + no failing tests
- **Atomic Operations:** File updates must be safe and rollback-capable
- **Event Logging:** All lifecycle changes automatically create events
- **Timestamp Control:** --time flag for deterministic testing support

### File Structure
```
├── cmd/
│   ├── start_epic.go       # Epic startup command
│   ├── done_epic.go        # Epic completion command
│   └── switch.go           # Epic switching command
├── internal/
│   └── lifecycle/          # Lifecycle service and logic
│       ├── service.go      # LifecycleService with DI
│       ├── validation.go   # Completion validation logic
│       ├── transitions.go  # State transition rules
│       └── events.go       # Event creation utilities
└── testdata/
    ├── epic-pending.xml    # Epic ready to start
    ├── epic-wip.xml        # Epic in progress
    ├── epic-incomplete.xml # Epic with pending work
    └── epic-complete.xml   # Epic ready for completion
```

### State Machine Implementation
```go
type EpicStatus string

const (
    StatusPending   EpicStatus = "pending"
    StatusWIP      EpicStatus = "wip"
    StatusDone     EpicStatus = "done"
)

func (s EpicStatus) CanTransitionTo(target EpicStatus) bool {
    transitions := map[EpicStatus][]EpicStatus{
        StatusPending: {StatusWIP},
        StatusWIP:     {StatusDone},
        StatusDone:    {}, // No transitions from done
    }
    
    for _, allowed := range transitions[s] {
        if allowed == target {
            return true
        }
    }
    return false
}
```

## Benefits of This Approach

✅ **Immediate Feedback** - Catch issues as soon as code is written  
✅ **Working Code** - Each phase delivers tested, working functionality  
✅ **Epic 3 Coverage** - All acceptance criteria distributed across phases  
✅ **Incremental Progress** - Agents can use lifecycle commands after each phase  
✅ **Risk Mitigation** - Problems caught early, not at the end  
✅ **Quality Assurance** - No untested code makes it to later phases  

## Test Distribution Summary

- **Phase 1 Tests:** 13 scenarios (Lifecycle service foundation, validation logic)
- **Phase 2 Tests:** 12 scenarios (Start epic command, file operations, atomic updates)
- **Phase 3 Tests:** 13 scenarios (Complete epic command, enhanced validation)
- **Phase 4 Tests:** 13 scenarios (Switch command, integration, cross-command testing)

**Total: All Epic 3 acceptance criteria and test scenarios covered across all phases**

---

## Implementation Status

### EPIC 3: EPIC LIFECYCLE MANAGEMENT - COMPLETE ✅
### Current Status: FULLY IMPLEMENTED

### Progress Tracking
- [x] Phase 1A: Create Lifecycle Service Foundation
- [x] Phase 1B: Write Lifecycle Service Tests
- [x] Phase 1C: Epic Validation & Completion Logic
- [x] Phase 1D: Write Validation Logic Tests
- [x] Phase 2A: Start Epic Command Implementation
- [x] Phase 2B: Write Start Epic Command Tests
- [x] Phase 2C: File Operations & Atomic Updates
- [x] Phase 2D: Write File Operations Tests
- [x] Phase 3A: Complete Epic Command Implementation
- [x] Phase 3B: Write Complete Epic Command Tests
- [x] Phase 3C: Enhanced Validation & Error Reporting
- [x] Phase 3D: Write Enhanced Validation Tests
- [x] Phase 4A: Switch Epic Command Implementation
- [x] Phase 4B: Write Switch Command Tests
- [x] Phase 4C: Integration & Cross-Command Testing
- [x] Phase 4D: Write Integration Tests

### Definition of Done
- [x] All acceptance criteria verified with automated tests
- [x] Lifecycle commands execute in < 200ms for typical epic files
- [x] Test coverage > 90% for lifecycle logic
- [x] All error cases handled gracefully with clear messages
- [x] Automatic event logging works for all lifecycle transitions
- [x] Timestamp injection (--time flag) works for deterministic testing
- [x] File operations are atomic and safe from corruption
- [x] Integration tests verify end-to-end lifecycle workflows