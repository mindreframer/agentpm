# EPIC-4: Task & Phase Management Implementation Plan
## Test-Driven Development Approach

### Phase 1: Phase Management Foundation + Tests (High Priority)

#### Phase 1A: Create Task Service Foundation
- [ ] Create internal/tasks package
- [ ] Define TaskService struct with Storage and Query service injection
- [ ] Implement phase state validation logic (pending/wip/done state machine)
- [ ] Create phase transition validation rules
- [ ] Phase constraint enforcement (only one active phase at a time)
- [ ] Event creation utilities for phase transitions
- [ ] Timestamp handling utilities (current time vs --time flag)

#### Phase 1B: Write Task Service Foundation Tests **IMMEDIATELY AFTER 1A**
Epic 4 Test Scenarios Covered:
- [ ] **Test: Start first phase of epic** (Epic 4 line 320)
- [ ] **Test: Prevent multiple active phases** (Epic 4 line 325)
- [ ] **Test: Phase state validation logic**
- [ ] **Test: Phase constraint enforcement**
- [ ] **Test: Event creation for phase transitions**
- [ ] **Test: Timestamp handling and injection**

#### Phase 1C: Phase Commands Implementation
- [ ] Create cmd/start_phase.go command
- [ ] Create cmd/done_phase.go command
- [ ] Integrate with TaskService for phase management
- [ ] Phase status transition from "pending" to "wip"
- [ ] Phase completion from "wip" to "done"
- [ ] --time flag support for deterministic testing
- [ ] Simple confirmation output format (non-XML)
- [ ] XML error output for constraint violations

#### Phase 1D: Write Phase Commands Tests **IMMEDIATELY AFTER 1C**
Epic 4 Test Scenarios Covered:
- [ ] **Test: Start first phase of epic** (detailed workflow)
- [ ] **Test: Prevent multiple active phases** (error handling)
- [ ] **Test: Complete phase with all tasks done** (Epic 4 line 330)
- [ ] **Test: Prevent completing phase with pending tasks** (Epic 4 line 335)
- [ ] **Test: Simple confirmation output format**
- [ ] **Test: XML error messages with actionable suggestions**
- [ ] **Test: Timestamp injection via --time flag**

### Phase 2: Task Management Implementation + Tests (High Priority)

#### Phase 2A: Task State Management & Commands
- [ ] Implement task state validation and transition logic
- [ ] Create cmd/start_task.go command
- [ ] Create cmd/done_task.go command
- [ ] Create cmd/cancel_task.go command
- [ ] Task-to-phase relationship validation
- [ ] Active task constraint enforcement (one per phase)
- [ ] Task status transitions (pending → wip → done/cancelled)
- [ ] Event logging for all task operations

#### Phase 2B: Write Task Management Tests **IMMEDIATELY AFTER 2A**
Epic 4 Test Scenarios Covered:
- [ ] **Test: Start task in active phase** (Epic 4 line 340)
- [ ] **Test: Prevent starting task in non-active phase** (Epic 4 line 345)
- [ ] **Test: Cancel active task** (Epic 4 line 365)
- [ ] **Test: Task state validation logic**
- [ ] **Test: Task-to-phase relationship validation**
- [ ] **Test: Active task constraint enforcement**
- [ ] **Test: Event creation for task transitions**

#### Phase 2C: Phase Completion Validation
- [ ] Implement phase completion requirements validation
- [ ] All tasks in phase must be "done" or "cancelled" before phase completion
- [ ] Task status checking for phase completion
- [ ] Detailed error reporting for incomplete phases
- [ ] Integration with phase completion commands
- [ ] Progress tracking updates after task completion

#### Phase 2D: Write Phase Completion Tests **IMMEDIATELY AFTER 2C**
Epic 4 Test Scenarios Covered:
- [ ] **Test: Complete phase with all tasks done** (validation logic)
- [ ] **Test: Prevent completing phase with pending tasks** (detailed errors)
- [ ] **Test: Phase completion with cancelled tasks allowed**
- [ ] **Test: Detailed error reporting for incomplete phases**
- [ ] **Test: Progress tracking updates correctly**

### Phase 3: Auto-Next Intelligence Implementation + Tests (Medium Priority)

#### Phase 3A: Auto-Next Selection Algorithm
- [ ] Create cmd/start_next.go command
- [ ] Implement auto-next task selection algorithm
- [ ] Current phase task selection logic
- [ ] Phase completion detection and auto-transition
- [ ] Next phase activation when current phase complete
- [ ] Smart task selection within newly activated phases
- [ ] All work completion detection

#### Phase 3B: Write Auto-Next Algorithm Tests **IMMEDIATELY AFTER 3A**
Epic 4 Test Scenarios Covered:
- [ ] **Test: Auto-select next task in current phase** (Epic 4 line 350)
- [ ] **Test: Auto-select next task in next phase** (Epic 4 line 355)
- [ ] **Test: Handle completion of all work** (Epic 4 line 360)
- [ ] **Test: Auto-next algorithm edge cases**
- [ ] **Test: Phase auto-completion logic**
- [ ] **Test: Smart task selection priorities**

#### Phase 3C: Auto-Next XML Output & Response Formatting
- [ ] Complex XML output for auto-next operations
- [ ] Task started XML response format
- [ ] Phase started XML response format (with task list)
- [ ] All work complete XML response format
- [ ] Integration with existing XML output patterns
- [ ] Auto-selected flag in XML responses

#### Phase 3D: Write Auto-Next Output Tests **IMMEDIATELY AFTER 3C**
Epic 4 Test Scenarios Covered:
- [ ] **Test: XML output for task auto-selection**
- [ ] **Test: XML output for phase auto-start**
- [ ] **Test: XML output for all work complete**
- [ ] **Test: Auto-selected flag in responses**
- [ ] **Test: XML format consistency with other commands**
- [ ] **Test: Decision context information in XML**

### Phase 4: Current State Tracking & Integration + Tests (Low Priority)

#### Phase 4A: Current State Management
- [ ] Implement current_state tracking in epic XML
- [ ] Active phase and task tracking
- [ ] Next action recommendation logic
- [ ] Integration with Epic 2 query commands
- [ ] Current state updates during phase/task transitions
- [ ] State consistency validation across commands

#### Phase 4B: Write Current State Tests **IMMEDIATELY AFTER 4A**
Epic 4 Test Scenarios Covered:
- [ ] **Test: Current state tracking updates correctly**
- [ ] **Test: Active phase and task tracking**
- [ ] **Test: Next action recommendations**
- [ ] **Test: Integration with query commands**
- [ ] **Test: State consistency across operations**

#### Phase 4C: Integration & Cross-Command Testing
- [ ] Integration between all task/phase commands
- [ ] Cross-command consistency (error formats, XML output)
- [ ] Global flag consistency (--time, --file)
- [ ] Help system integration for all task commands
- [ ] Performance optimization for state validation
- [ ] File operations safety and atomic updates

#### Phase 4D: Write Integration Tests **IMMEDIATELY AFTER 4C**
Epic 4 Test Scenarios Covered:
- [ ] **Test: End-to-end phase/task workflow**
- [ ] **Test: Cross-command consistency**
- [ ] **Test: Global flag handling**
- [ ] **Test: Help system completeness**
- [ ] **Test: Performance requirements met**
- [ ] **Test: File operations are atomic and safe**
- [ ] **Test: Integration with Epic 2 query commands**

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
- **Epic 2:** Query service for current state validation and integration
- **Epic 3:** Epic lifecycle for overall epic state management
- **Event System:** Event creation and logging utilities

### Technical Requirements
- **Sequential Work:** Only one phase active at a time, only one task active per phase
- **State Validation:** All transitions validated against current epic and phase state
- **Dual Output:** Simple confirmations for routine ops, XML for auto-next
- **Current State:** Track active phase/task in epic XML structure
- **Auto-Next Intelligence:** Smart task selection with phase auto-completion

### File Structure
```
├── cmd/
│   ├── start_phase.go      # Phase start command
│   ├── done_phase.go       # Phase completion command
│   ├── start_task.go       # Task start command
│   ├── done_task.go        # Task completion command
│   ├── cancel_task.go      # Task cancellation command
│   └── start_next.go       # Auto-next task selection
├── internal/
│   └── tasks/              # Task management service and logic
│       ├── service.go      # TaskService with DI
│       ├── phases.go       # Phase state management
│       ├── tasks.go        # Task state management
│       ├── autonext.go     # Auto-next selection algorithm
│       └── validation.go   # State validation logic
└── testdata/
    ├── epic-no-active.xml  # Epic with no active phase/task
    ├── epic-active-phase.xml # Epic with active phase
    ├── epic-active-task.xml  # Epic with active task
    └── epic-mixed-states.xml # Epic with various task states
```

### State Machine Implementation
```go
type PhaseStatus string
type TaskStatus string

const (
    StatusPending   PhaseStatus = "pending"
    StatusWIP      PhaseStatus = "wip"
    StatusDone     PhaseStatus = "done"
)

const (
    TaskPending    TaskStatus = "pending"
    TaskWIP       TaskStatus = "wip"
    TaskDone      TaskStatus = "done"
    TaskCancelled TaskStatus = "cancelled"
)

// Constraint validation
func (ts *TaskService) CanStartPhase(phaseID string) error {
    activePhase := ts.GetActivePhase()
    if activePhase != nil {
        return fmt.Errorf("cannot start phase %s: phase %s is still active", 
            phaseID, activePhase.ID)
    }
    return nil
}
```

### Auto-Next Algorithm Priority
1. **Current Phase Tasks:** Find next pending task in active phase
2. **Complete Phase:** Auto-complete phase when all tasks done/cancelled
3. **Next Phase:** Activate next pending phase when current complete
4. **First Task:** Start first pending task in newly activated phase
5. **All Complete:** Return completion message when no work remains

## Benefits of This Approach

✅ **Immediate Feedback** - Catch issues as soon as code is written  
✅ **Working Code** - Each phase delivers tested, working functionality  
✅ **Epic 4 Coverage** - All acceptance criteria distributed across phases  
✅ **Incremental Progress** - Agents can use task commands after each phase  
✅ **Risk Mitigation** - Problems caught early, not at the end  
✅ **Quality Assurance** - No untested code makes it to later phases  

## Test Distribution Summary

- **Phase 1 Tests:** 13 scenarios (Phase management foundation, phase commands)
- **Phase 2 Tests:** 11 scenarios (Task management, phase completion validation)
- **Phase 3 Tests:** 12 scenarios (Auto-next algorithm, XML output formatting)
- **Phase 4 Tests:** 13 scenarios (Current state tracking, integration testing)

**Total: All Epic 4 acceptance criteria and test scenarios covered across all phases**

---

## Implementation Status

### EPIC 4: TASK & PHASE MANAGEMENT - PENDING
### Current Status: READY TO START (After Epic 1, 2 & 3 Complete)

### Progress Tracking
- [ ] Phase 1A: Create Task Service Foundation
- [ ] Phase 1B: Write Task Service Foundation Tests
- [ ] Phase 1C: Phase Commands Implementation
- [ ] Phase 1D: Write Phase Commands Tests
- [ ] Phase 2A: Task State Management & Commands
- [ ] Phase 2B: Write Task Management Tests
- [ ] Phase 2C: Phase Completion Validation
- [ ] Phase 2D: Write Phase Completion Tests
- [ ] Phase 3A: Auto-Next Selection Algorithm
- [ ] Phase 3B: Write Auto-Next Algorithm Tests
- [ ] Phase 3C: Auto-Next XML Output & Response Formatting
- [ ] Phase 3D: Write Auto-Next Output Tests
- [ ] Phase 4A: Current State Management
- [ ] Phase 4B: Write Current State Tests
- [ ] Phase 4C: Integration & Cross-Command Testing
- [ ] Phase 4D: Write Integration Tests

### Definition of Done
- [ ] All acceptance criteria verified with automated tests
- [ ] Phase/task commands execute in < 150ms for typical epic files
- [ ] Test coverage > 90% for task management logic
- [ ] Auto-next logic works correctly in all scenarios
- [ ] All constraint violations handled with clear error messages
- [ ] Simple confirmation output for routine operations
- [ ] XML output for complex operations (auto-next)
- [ ] Event logging works for all phase/task transitions
- [ ] Integration tests verify end-to-end task workflows