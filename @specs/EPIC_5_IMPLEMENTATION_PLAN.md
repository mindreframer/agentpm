# EPIC-5: Task & Phase Management Implementation Plan
## Test-Driven Development Approach

### Phase 1: Phase Management Foundation + Tests (High Priority)

#### Phase 1A: Create Phase Service Foundation
- [ ] Create internal/phases package
- [ ] Define PhaseService struct with Storage and Query injection
- [ ] Implement phase status validation logic (pending/wip/done state machine)
- [ ] Create phase state transition validation rules
- [ ] Single active phase constraint enforcement
- [ ] Phase completion validation (all tasks done/cancelled)
- [ ] Timestamp handling utilities for phase lifecycle events
- [ ] Basic phase lookup and status management operations

#### Phase 1B: Write Phase Service Tests **IMMEDIATELY AFTER 1A**
Epic 5 Test Scenarios Covered:
- [ ] **Test: Start first phase of epic** (AC-1)
- [ ] **Test: Prevent multiple active phases** (AC-2)
- [ ] **Test: Complete phase with all tasks done** (AC-3)
- [ ] **Test: Prevent completing phase with pending tasks** (AC-4)
- [ ] **Test: Phase state transition validation**
- [ ] **Test: Single active phase constraint**
- [ ] **Test: Phase completion validation logic**

#### Phase 1C: Phase Management Commands
- [ ] Create cmd/start_phase.go command
- [ ] Create cmd/done_phase.go command
- [ ] Integrate with PhaseService for all operations
- [ ] --time flag support for deterministic testing
- [ ] Simple confirmation output messages
- [ ] Error handling for constraint violations
- [ ] XML error responses with actionable suggestions

#### Phase 1D: Write Phase Commands Tests **IMMEDIATELY AFTER 1C**
Epic 5 Test Scenarios Covered:
- [ ] **Test: start-phase command execution** (FR-1)
- [ ] **Test: done-phase command execution** (FR-1)
- [ ] **Test: Phase constraint violation errors** (FR-1)
- [ ] **Test: Command error handling and XML output**
- [ ] **Test: Simple confirmation output format**
- [ ] **Test: Timestamp injection via --time flag**

### Phase 2: Task Management Foundation + Tests (High Priority)

#### Phase 2A: Create Task Service Foundation
- [ ] Create internal/tasks package
- [ ] Define TaskService struct with Storage and Query injection
- [ ] Implement task status validation logic (pending/wip/done/cancelled state machine)
- [ ] Create task state transition validation rules
- [ ] Task-to-phase relationship validation
- [ ] Single active task per phase constraint enforcement
- [ ] Task prerequisite validation (phase must be active)
- [ ] Timestamp handling utilities for task lifecycle events

#### Phase 2B: Write Task Service Tests **IMMEDIATELY AFTER 2A**
Epic 5 Test Scenarios Covered:
- [ ] **Test: Start task in active phase** (AC-5)
- [ ] **Test: Prevent starting task in non-active phase** (AC-6)
- [ ] **Test: Cancel active task** (AC-10)
- [ ] **Test: Task state transition validation**
- [ ] **Test: Task-to-phase relationship validation**
- [ ] **Test: Single active task constraint**
- [ ] **Test: Task prerequisite checking**

#### Phase 2C: Task Management Commands
- [ ] Create cmd/start_task.go command
- [ ] Create cmd/done_task.go command
- [ ] Create cmd/cancel_task.go command
- [ ] Integrate with TaskService for all operations
- [ ] --time flag support for deterministic testing
- [ ] Simple confirmation output messages
- [ ] Error handling for invalid state transitions
- [ ] Task cancellation reason tracking

#### Phase 2D: Write Task Commands Tests **IMMEDIATELY AFTER 2C**
Epic 5 Test Scenarios Covered:
- [ ] **Test: start-task command execution** (FR-2)
- [ ] **Test: done-task command execution** (FR-2)
- [ ] **Test: cancel-task command execution** (FR-2)
- [ ] **Test: Task command error handling**
- [ ] **Test: Task cancellation reason storage**
- [ ] **Test: Output format validation**
- [ ] **Test: Cross-phase task validation**

### Phase 3: Auto-Next Intelligence + Tests (Medium Priority)

#### Phase 3A: Auto-Next Selection Algorithm
- [ ] Create internal/autonext package
- [ ] Define AutoNextService struct with Phase and Task service injection
- [ ] Implement auto-next selection algorithm with priority logic
- [ ] Current phase task selection logic
- [ ] Phase completion detection and auto-transition
- [ ] Next phase activation and first task selection
- [ ] All work completion detection
- [ ] Complex XML output formatting for decision context

#### Phase 3B: Write Auto-Next Algorithm Tests **IMMEDIATELY AFTER 3A**
Epic 5 Test Scenarios Covered:
- [ ] **Test: Auto-select next task in current phase** (AC-7)
- [ ] **Test: Auto-select next task in next phase** (AC-8)
- [ ] **Test: Handle completion of all work** (AC-9)
- [ ] **Test: Auto-next priority logic**
- [ ] **Test: Phase completion detection**
- [ ] **Test: Next phase activation logic**
- [ ] **Test: Complex XML output format**

#### Phase 3C: Start-Next Command Implementation
- [ ] Create cmd/start_next.go command
- [ ] Integrate with AutoNextService for intelligent selection
- [ ] --time flag support for deterministic testing
- [ ] XML output for complex decision making
- [ ] Different output formats based on operation type
- [ ] Error handling for edge cases
- [ ] Integration with phase and task services

#### Phase 3D: Write Start-Next Command Tests **IMMEDIATELY AFTER 3C**
Epic 5 Test Scenarios Covered:
- [ ] **Test: start-next command execution** (FR-3)
- [ ] **Test: XML output for task selection** (FR-3)
- [ ] **Test: XML output for phase activation** (FR-3)
- [ ] **Test: XML output for completion** (FR-3)
- [ ] **Test: Command error handling**
- [ ] **Test: Edge case handling**
- [ ] **Test: Integration with other services**

### Phase 4: Automatic Event Integration + Tests (Medium Priority)

#### Phase 4A: Automatic Event Creation
- [ ] Integrate Epic 4 EventService with Phase and Task services
- [ ] Automatic event creation for phase state transitions
- [ ] Automatic event creation for task state transitions
- [ ] phase_started, phase_completed event types
- [ ] task_started, task_completed, task_cancelled event types
- [ ] Rich event metadata for phase/task operations
- [ ] Event-phase-task relationship tracking

#### Phase 4B: Write Automatic Event Tests **IMMEDIATELY AFTER 4A**
Epic 5 Test Scenarios Covered:
- [ ] **Test: Automatic phase_started event creation** (FR-5)
- [ ] **Test: Automatic phase_completed event creation** (FR-5)
- [ ] **Test: Automatic task_started event creation** (FR-5)
- [ ] **Test: Automatic task_completed event creation** (FR-5)
- [ ] **Test: Automatic task_cancelled event creation** (FR-5)
- [ ] **Test: Event metadata accuracy**
- [ ] **Test: Event-phase-task relationship tracking**

#### Phase 4C: Progress Tracking & State Validation
- [ ] Enhanced epic progress calculation with phases/tasks
- [ ] Current state tracking (active phase, active task)
- [ ] Next action determination for Epic 2 integration
- [ ] Progress percentage calculation across phases/tasks
- [ ] State validation utilities for cross-command consistency
- [ ] Integration with Epic 2 query enhancements

#### Phase 4D: Write Progress Tracking Tests **IMMEDIATELY AFTER 4C**
Epic 5 Test Scenarios Covered:
- [ ] **Test: Epic progress calculation accuracy** (FR-4)
- [ ] **Test: Current state tracking** (FR-4)
- [ ] **Test: Next action determination** (FR-4)
- [ ] **Test: Progress percentage calculation**
- [ ] **Test: State validation utilities**
- [ ] **Test: Epic 2 integration**

### Phase 5: Enhanced Data Model & Integration + Tests (Low Priority)

#### Phase 5A: Enhanced Phase & Task Data Model
- [ ] Rich phase XML structure with timestamps
- [ ] Enhanced task XML structure with lifecycle data
- [ ] Epic current state tracking in XML
- [ ] Phase completion metadata
- [ ] Task cancellation reason storage
- [ ] XML schema validation for new structures
- [ ] Migration utilities for existing epic files

#### Phase 5B: Write Data Model Tests **IMMEDIATELY AFTER 5A**
Epic 5 Test Scenarios Covered:
- [ ] **Test: Rich phase XML structure** (Data Model)
- [ ] **Test: Enhanced task XML structure** (Data Model)
- [ ] **Test: Epic current state tracking** (Data Model)
- [ ] **Test: Phase completion metadata**
- [ ] **Test: Task cancellation data storage**
- [ ] **Test: XML schema validation**

#### Phase 5C: Cross-Command Integration & Performance
- [ ] Integration with Epic 2 current/status command enhancements
- [ ] Integration with Epic 3 epic lifecycle validation
- [ ] Performance optimization for phase/task operations
- [ ] Memory usage optimization for large epics
- [ ] Cross-command consistency verification
- [ ] Error message standardization

#### Phase 5D: Write Integration & Performance Tests **IMMEDIATELY AFTER 5C**
Epic 5 Test Scenarios Covered:
- [ ] **Test: Epic 2 integration functionality** (NFR-4)
- [ ] **Test: Epic 3 lifecycle integration** (NFR-4)
- [ ] **Test: Performance with large epics** (NFR-1)
- [ ] **Test: Memory usage optimization** (NFR-1)
- [ ] **Test: Cross-command consistency** (NFR-4)
- [ ] **Test: Error message standardization**

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
- **Epic 2:** Query service for current state validation and status display
- **Epic 3:** Epic lifecycle for overall epic state management
- **Epic 4:** Event service for automatic event creation
- **Testing:** Deterministic timestamp support for snapshot testing

### Technical Requirements
- **Phase State Machine:** pending → wip → done transitions
- **Task State Machine:** pending → wip → done/cancelled transitions
- **Work Constraints:** Only one active phase, only one active task per phase
- **Auto-Next Logic:** Intelligent task/phase selection with decision context
- **Event Integration:** Automatic event creation for all transitions
- **Performance:** < 150ms for phase/task operations

### File Structure
```
├── cmd/
│   ├── start_phase.go      # Phase start command
│   ├── done_phase.go       # Phase completion command
│   ├── start_task.go       # Task start command
│   ├── done_task.go        # Task completion command
│   ├── cancel_task.go      # Task cancellation command
│   └── start_next.go       # Auto-next selection command
├── internal/
│   ├── phases/             # Phase management service
│   │   ├── service.go      # PhaseService with DI
│   │   ├── validation.go   # Phase state validation
│   │   └── transitions.go  # Phase state transitions
│   ├── tasks/              # Task management service
│   │   ├── service.go      # TaskService with DI
│   │   ├── validation.go   # Task state validation
│   │   └── transitions.go  # Task state transitions
│   └── autonext/           # Auto-next selection logic
│       ├── service.go      # AutoNextService with DI
│       ├── algorithm.go    # Selection algorithm
│       └── formatting.go   # XML output formatting
└── testdata/
    ├── phase-pending.xml   # Phase in pending state
    ├── phase-wip.xml       # Phase in progress
    ├── task-multi.xml      # Multiple tasks in various states
    └── epic-complex.xml    # Complex multi-phase epic
```

### State Machine Implementation
```go
type PhaseStatus string

const (
    PhaseStatusPending PhaseStatus = "pending"
    PhaseStatusWIP     PhaseStatus = "wip" 
    PhaseStatusDone    PhaseStatus = "done"
)

type TaskStatus string

const (
    TaskStatusPending   TaskStatus = "pending"
    TaskStatusWIP      TaskStatus = "wip"
    TaskStatusDone     TaskStatus = "done"
    TaskStatusCancelled TaskStatus = "cancelled"
)

func (p PhaseStatus) CanTransitionTo(target PhaseStatus) bool {
    transitions := map[PhaseStatus][]PhaseStatus{
        PhaseStatusPending: {PhaseStatusWIP},
        PhaseStatusWIP:     {PhaseStatusDone},
        PhaseStatusDone:    {}, // No transitions from done
    }
    
    for _, allowed := range transitions[p] {
        if allowed == target {
            return true
        }
    }
    return false
}

func (t TaskStatus) CanTransitionTo(target TaskStatus) bool {
    transitions := map[TaskStatus][]TaskStatus{
        TaskStatusPending:   {TaskStatusWIP},
        TaskStatusWIP:      {TaskStatusDone, TaskStatusCancelled},
        TaskStatusDone:     {}, // No transitions from done
        TaskStatusCancelled: {}, // No transitions from cancelled
    }
    
    for _, allowed := range transitions[t] {
        if allowed == target {
            return true
        }
    }
    return false
}
```

### Auto-Next Selection Algorithm
```go
type AutoNextResult struct {
    Action       string // "start_task", "start_phase", "complete_epic"
    PhaseID      string
    TaskID       string
    Message      string
    XMLOutput    string
}

func (s *AutoNextService) SelectNext() (*AutoNextResult, error) {
    // 1. Check for active phase with pending tasks
    // 2. Check for completed phase that needs completion
    // 3. Check for next pending phase to activate
    // 4. Return completion message if all done
}
```

## Benefits of This Approach

✅ **Immediate Feedback** - Catch issues as soon as code is written  
✅ **Working Code** - Each phase delivers tested, working functionality  
✅ **Epic 5 Coverage** - All acceptance criteria distributed across phases  
✅ **Incremental Progress** - Agents can use phase/task management after each phase  
✅ **Risk Mitigation** - Problems caught early, not at the end  
✅ **Quality Assurance** - No untested code makes it to later phases  

## Test Distribution Summary

- **Phase 1 Tests:** 13 scenarios (Phase service foundation, phase commands)
- **Phase 2 Tests:** 14 scenarios (Task service foundation, task commands) 
- **Phase 3 Tests:** 14 scenarios (Auto-next algorithm, start-next command)
- **Phase 4 Tests:** 14 scenarios (Automatic events, progress tracking)
- **Phase 5 Tests:** 12 scenarios (Data model, integration, performance)

**Total: All Epic 5 acceptance criteria and test scenarios covered across all phases**

---

## Implementation Status

### EPIC 5: TASK & PHASE MANAGEMENT - PENDING
### Current Status: READY FOR IMPLEMENTATION

### Progress Tracking
- [ ] Phase 1A: Create Phase Service Foundation
- [ ] Phase 1B: Write Phase Service Tests
- [ ] Phase 1C: Phase Management Commands
- [ ] Phase 1D: Write Phase Commands Tests
- [ ] Phase 2A: Create Task Service Foundation
- [ ] Phase 2B: Write Task Service Tests
- [ ] Phase 2C: Task Management Commands
- [ ] Phase 2D: Write Task Commands Tests
- [ ] Phase 3A: Auto-Next Selection Algorithm
- [ ] Phase 3B: Write Auto-Next Algorithm Tests
- [ ] Phase 3C: Start-Next Command Implementation
- [ ] Phase 3D: Write Start-Next Command Tests
- [ ] Phase 4A: Automatic Event Creation
- [ ] Phase 4B: Write Automatic Event Tests
- [ ] Phase 4C: Progress Tracking & State Validation
- [ ] Phase 4D: Write Progress Tracking Tests
- [ ] Phase 5A: Enhanced Phase & Task Data Model
- [ ] Phase 5B: Write Data Model Tests
- [ ] Phase 5C: Cross-Command Integration & Performance
- [ ] Phase 5D: Write Integration & Performance Tests

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