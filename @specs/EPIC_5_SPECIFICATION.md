# EPIC-5 SPECIFICATION: Task & Phase Management

## Overview

**Epic ID:** 5
**Name:** Task & Phase Management  
**Duration:** 4-5 days  
**Status:** pending  
**Priority:** high  
**Depends On:** Epic 1 (Foundation), Epic 2 (Query Commands), Epic 3 (Epic Lifecycle), Epic 4 (Tests + Event logging)

**Goal:** Implement granular work tracking at phase and task levels, enabling agents to manage detailed development workflows within epics through structured phase and task state transitions.

## Business Context

Epic 5 is identified in the roadmap as one of the "most important epics" alongside Epic 4. It provides agents with fine-grained control over their development workflow by managing phases and tasks within epics. The system enforces sequential work patterns - agents cannot start multiple phases simultaneously and must complete all tasks within a phase before marking it complete. This structured approach helps agents maintain focus and ensures comprehensive completion of work.

## User Stories

### Primary User Stories
- **As an agent, I can start working on a specific phase** so that I can begin organized work within an epic
- **As an agent, I can begin individual tasks within phases** so that I can track granular progress
- **As an agent, I can automatically pick the next pending task** so that I can maintain workflow momentum without manual selection
- **As an agent, I can mark tasks and phases as completed** so that I can track progress through complex work plans
- **As an agent, I can cancel tasks when needed** so that I can handle changing requirements or blocked work

### Secondary User Stories
- **As an agent, I can track progress through complex work plans** so that I understand my position in the overall epic
- **As an agent, I can receive minimal confirmation output** so that routine operations don't overwhelm me with information
- **As an agent, I can get detailed XML output for smart operations** so that I can make informed decisions about next steps

## Technical Requirements

### Core Dependencies
- **Foundation:** Epic 1 CLI framework, XML processing, storage interface
- **Querying:** Epic 2 query service for current state and progress tracking
- **Lifecycle:** Epic 3 lifecycle management for epic-level state validation
- **Event System:** Automatic event logging for all phase/task transitions

### Architecture Principles
- **Sequential Work:** Only one phase can be active at a time per epic
- **Phase Completion:** All tasks in a phase must be completed before phase completion
- **Auto-Next Intelligence:** Smart selection of next pending task based on current context
- **Minimal Output:** Simple confirmations for routine operations, XML for complex decisions
- **State Validation:** All transitions validated against current epic and phase state

### Work State Rules
```
PHASE STATES: pending → wip → done
TASK STATES:  pending → wip → done
              pending → wip → cancelled

CONSTRAINTS:
- Only one phase can be "wip" at a time
- Only one task can be "wip" at a time within a phase
- Phase can only be completed when all tasks are "done" or "cancelled"
- Cannot start new phase if another phase is active
- Cannot start task in non-active phase
```

### Auto-Next Logic Priority
1. **Prefer Current Phase:** Select next pending task in currently active phase
2. **Complete Phase First:** Only move to next phase when current phase is fully complete
3. **Sequential Phases:** Activate next pending phase in order when previous is complete
4. **Intelligent Selection:** Choose first pending task in newly activated phase

## Functional Requirements

### FR-1: Phase Management
**Commands:** 
- `agentpm start-phase <phase-id> [--time <timestamp>] [--file <epic-file>]`
- `agentpm done-phase <phase-id> [--time <timestamp>] [--file <epic-file>]`

**Start Phase Behavior:**
- Changes phase status from "pending" to "wip"
- Validates no other phase is currently active
- Sets phase started_at timestamp
- Creates automatic event log for phase start
- Returns simple confirmation message

**Complete Phase Behavior:**
- Changes phase status from "wip" to "done"
- Validates all tasks in phase are "done" or "cancelled"
- Sets phase completed_at timestamp
- Creates automatic event log for phase completion
- Returns simple confirmation message

**Output Format (Simple Confirmations):**
```
Phase 2A started.
Phase 2A completed.
```

**Error Output Format:**
```xml
<error>
    <type>phase_constraint_violation</type>
    <message>Cannot start phase 2A: phase 1A is still active</message>
    <details>
        <active_phase>1A</active_phase>
        <active_task>1A_2</active_task>
        <suggestion>Complete phase 1A first or use 'agentpm current' to see active work</suggestion>
    </details>
</error>
```

### FR-2: Task Management
**Commands:**
- `agentpm start-task <task-id> [--time <timestamp>] [--file <epic-file>]`
- `agentpm done-task <task-id> [--time <timestamp>] [--file <epic-file>]`
- `agentpm cancel-task <task-id> [--time <timestamp>] [--file <epic-file>]`

**Start Task Behavior:**
- Changes task status from "pending" to "wip"
- Validates task belongs to currently active phase
- Validates no other task is currently active in the phase
- Sets task started_at timestamp
- Creates automatic event log for task start

**Complete Task Behavior:**
- Changes task status from "wip" to "done"
- Sets task completed_at timestamp
- Creates automatic event log for task completion

**Cancel Task Behavior:**
- Changes task status from "wip" to "cancelled"
- Sets task cancelled_at timestamp with reason
- Creates automatic event log for task cancellation

**Output Format (Simple Confirmations):**
```
Task 2A_1 started.
Task 2A_1 completed.
Task 2A_1 cancelled.
```

### FR-3: Smart Auto-Next Task Selection
**Command:** `agentpm start-next [--time <timestamp>] [--file <epic-file>]`

**Behavior Logic:**
1. **If Active Phase Exists:** Find next pending task in current active phase
2. **If No Active Phase:** Find next pending phase and activate it, then start first pending task
3. **If Current Phase Complete:** Complete current phase, activate next phase, start first task
4. **If All Work Complete:** Return completion message

**Output Format (XML for Decision Making):**
```xml
<!-- When starting task in active phase -->
<task_started epic="8" task="2A_2">
    <task_description>Add accessibility features to pagination controls</task_description>
    <phase_id>2A</phase_id>
    <previous_status>pending</previous_status>
    <new_status>wip</new_status>
    <started_at>2025-08-16T15:00:00Z</started_at>
    <auto_selected>true</auto_selected>
    <message>Started Task 2A_2: Add accessibility features to pagination controls (auto-selected)</message>
</task_started>

<!-- When starting new phase -->
<phase_started epic="8" phase="2A">
    <phase_name>Create PaginationComponent</phase_name>
    <previous_status>pending</previous_status>
    <new_status>wip</new_status>
    <started_at>2025-08-16T14:00:00Z</started_at>
    <tasks>
        <task id="2A_1" status="pending">Create PaginationComponent with Previous/Next controls</task>
        <task id="2A_2" status="pending">Add accessibility features to pagination controls</task>
    </tasks>
    <started_task>2A_1</started_task>
    <message>Started Phase 2A and Task 2A_1 (auto-selected)</message>
</phase_started>

<!-- When all work is complete -->
<all_complete epic="8">
    <message>All phases and tasks completed. Epic ready for completion.</message>
    <suggestion>Use 'agentpm done-epic' to complete the epic</suggestion>
</all_complete>
```

### FR-4: Progress Tracking & State Validation
**Internal Functionality:** Used by all phase/task commands

**Phase State Validation:**
- Phase can only be started if no other phase is active
- Phase can only be completed if all tasks are done/cancelled
- Phase transitions must be sequential (cannot skip phases)

**Task State Validation:**
- Task can only be started if its phase is active
- Task can only be started if no other task in phase is active
- Task can only be completed/cancelled if it's currently wip

**Progress Updates:**
- Automatic recalculation of epic completion percentage
- Phase completion status updates
- Epic-level progress tracking
- Event logging for all state changes

### FR-5: Automatic Event Logging
**Behavior:** All phase/task commands automatically create event entries

**Event Types:**
- `phase_started`: When phase transitions from pending to wip
- `phase_completed`: When phase transitions from wip to done
- `task_started`: When task transitions from pending to wip
- `task_completed`: When task transitions from wip to done
- `task_cancelled`: When task transitions from wip to cancelled

**Event Format:**
```xml
<event timestamp="2025-08-16T14:15:00Z" agent="agent_claude" phase_id="2A" type="task_started">
    Task 2A_1 started
    
    Task: Create PaginationComponent with Previous/Next controls
    Phase: 2A - Create PaginationComponent
    Status: pending → wip
</event>
```

## Non-Functional Requirements

### NFR-1: Performance
- Phase/task commands execute in < 150ms for typical epic files
- Auto-next logic performs efficiently without loading unnecessary data
- State validation is fast and doesn't impact user experience

### NFR-2: Reliability
- All state transitions are validated before execution
- File operations are atomic to prevent corruption
- Clear error messages for all constraint violations
- Rollback capability if operations fail

### NFR-3: Usability (for Agents)
- Simple confirmation messages for routine operations reduce noise
- XML output for complex operations (auto-next) provides decision context
- Clear error messages with actionable suggestions
- Consistent behavior across all phase/task operations

### NFR-4: Consistency
- State transition rules consistently enforced
- Event logging format consistent across all operations
- Error message format follows established patterns
- Timestamp handling consistent with other commands

## Data Model Changes

### Phase XML Structure
```xml
<phase id="2A" name="Create PaginationComponent" status="wip">
    <started_at>2025-08-16T14:00:00Z</started_at>
    <completed_at>2025-08-16T16:30:00Z</completed_at> <!-- only when status="done" -->
    <description>Implementation phase for pagination component</description>
</phase>
```

### Task XML Structure
```xml
<task id="2A_1" phase_id="2A" status="wip">
    <started_at>2025-08-16T14:15:00Z</started_at>
    <completed_at>2025-08-16T14:45:00Z</completed_at> <!-- only when status="done" -->
    <cancelled_at>2025-08-16T15:00:00Z</cancelled_at> <!-- only when status="cancelled" -->
    <description>Create PaginationComponent with Previous/Next controls</description>
</task>
```

### Epic Progress Tracking
```xml
<epic id="8">
    <current_state>
        <active_phase>2A</active_phase>
        <active_task>2A_1</active_task>
        <next_action>Fix mobile responsive pagination controls</next_action>
    </current_state>
    <!-- other epic content -->
</epic>
```

## Error Handling

### Error Categories
1. **Constraint Violations:** Multiple active phases, starting task in inactive phase
2. **State Transition Errors:** Invalid status changes, completing incomplete phases
3. **Reference Errors:** Non-existent phase/task IDs, invalid phase/task associations
4. **Completion Errors:** Attempting to complete phase with pending tasks

### Error Response Examples

**Multiple Active Phases:**
```xml
<error>
    <type>constraint_violation</type>
    <message>Cannot start phase 2A: phase 1A is still active</message>
    <details>
        <active_phase>1A</active_phase>
        <active_task>1A_2</active_task>
        <pending_tasks_in_active_phase>1</pending_tasks_in_active_phase>
        <suggestion>Complete phase 1A first or use 'agentpm done-phase 1A'</suggestion>
    </details>
</error>
```

**Incomplete Phase Completion:**
```xml
<error>
    <type>incomplete_phase</type>
    <message>Cannot complete phase 2A: 2 tasks are still pending</message>
    <details>
        <phase_id>2A</phase_id>
        <pending_tasks>
            <task id="2A_2" status="pending">Add accessibility features</task>
            <task id="2A_3" status="wip">URL state persistence</task>
        </pending_tasks>
        <suggestion>Complete or cancel all tasks in phase 2A first</suggestion>
    </details>
</error>
```

## Acceptance Criteria

### AC-1: Start First Phase of Epic
- **GIVEN** I have an epic in "wip" status with no active phase
- **WHEN** I run `agentpm start-phase 1A`
- **THEN** phase "1A" should change to "wip" and become the active phase

### AC-2: Prevent Multiple Active Phases
- **GIVEN** I have phase "1A" currently active
- **WHEN** I run `agentpm start-phase 2A`
- **THEN** I should get an error that phase "1A" must be completed first

### AC-3: Complete Phase with All Tasks Done
- **GIVEN** I have phase "1A" with all tasks completed
- **WHEN** I run `agentpm done-phase 1A`
- **THEN** phase "1A" should change to "done" status

### AC-4: Prevent Completing Phase with Pending Tasks
- **GIVEN** I have phase "1A" with pending tasks
- **WHEN** I run `agentpm done-phase 1A`
- **THEN** I should get an error listing the pending tasks in that phase

### AC-5: Start Task in Active Phase
- **GIVEN** I have phase "2A" active with task "2A_1" pending
- **WHEN** I run `agentpm start-task 2A_1`
- **THEN** task "2A_1" should change to "wip" and become the active task

### AC-6: Prevent Starting Task in Non-Active Phase
- **GIVEN** I have phase "1A" active and phase "2A" pending
- **WHEN** I run `agentpm start-task 2A_1`
- **THEN** I should get an error that phase "2A" is not active

### AC-7: Auto-Select Next Task in Current Phase
- **GIVEN** I have completed task "2A_1" and task "2A_2" is pending in same phase
- **WHEN** I run `agentpm start-next`
- **THEN** task "2A_2" should be automatically selected and started

### AC-8: Auto-Select Next Task in Next Phase
- **GIVEN** I have completed all tasks in phase "1A" and phase "2A" has pending tasks
- **WHEN** I run `agentpm start-next`
- **THEN** phase "1A" should be completed, phase "2A" should be started, and first pending task should be started

### AC-9: Handle Completion of All Work
- **GIVEN** I have completed all tasks in all phases
- **WHEN** I run `agentpm start-next`
- **THEN** I should get a message that all work is completed

### AC-10: Cancel Active Task
- **GIVEN** I have task "2A_2" in "wip" status
- **WHEN** I run `agentpm cancel-task 2A_2`
- **THEN** task "2A_2" should change to "cancelled" status

## Testing Strategy

### Test Categories
- **Unit Tests (70%):** State transition logic, validation rules, auto-next algorithms
- **Integration Tests (25%):** Command execution, file operations, cross-command consistency
- **Workflow Tests (5%):** End-to-end phase/task workflows, complex scenarios

### Test Data Requirements
- Epic files with various phase/task configurations
- Epic files with different completion states
- Epic files with mixed task statuses (pending, wip, done, cancelled)
- Multi-phase epics for sequential workflow testing

### Test Isolation
- Each test uses isolated epic files in `t.TempDir()`
- MemoryStorage for fast unit tests
- Deterministic timestamps for consistent snapshots
- No shared state between tests

## Implementation Phases

### Phase 5A: Phase Management Foundation (Day 1-2)
- Create internal/tasks package
- Implement TaskService with Storage and Query injection
- Phase state validation and transition logic
- Phase start/complete command implementation
- Basic event logging for phase operations

### Phase 5B: Task Management Implementation (Day 2-3)
- Task state validation and transition logic
- Task start/complete/cancel command implementation
- Task-to-phase relationship validation
- Active task constraint enforcement
- Event logging for task operations

### Phase 5C: Auto-Next Intelligence (Day 3-4)
- Auto-next selection algorithm implementation
- Phase completion detection and auto-transition
- Smart task selection within phases
- Complex XML output for auto-next responses
- Integration with existing query services

### Phase 5D: Integration & Performance Optimization (Day 4-5)
- Cross-command consistency and integration
- Performance optimization for state validation
- Error message refinement and user experience
- Comprehensive testing and edge case handling
- Documentation and help system updates

## Definition of Done

- [ ] All acceptance criteria verified with automated tests
- [ ] Phase/task commands execute in < 150ms for typical epic files
- [ ] Test coverage > 90% for task management logic
- [ ] Auto-next logic works correctly in all scenarios
- [ ] All constraint violations handled with clear error messages
- [ ] Simple confirmation output for routine operations
- [ ] XML output for complex operations (auto-next)
- [ ] Event logging works for all phase/task transitions
- [ ] Integration tests verify end-to-end task workflows

## Dependencies and Risks

### Dependencies
- **Epic 1:** CLI framework, XML processing, storage interface
- **Epic 2:** Query service for current state validation
- **Epic 3:** Epic lifecycle for overall epic state management
- **Epic 4:** Epic event logging + tests state management
- **Event System:** Event creation and logging utilities

### Risks
- **Medium Risk:** Auto-next logic complexity with edge cases
- **Low Risk:** State constraint validation complexity
- **Low Risk:** File corruption during concurrent task operations

### Mitigation Strategies
- Comprehensive testing of auto-next algorithm with various epic states
- Clear state machine implementation with exhaustive validation
- Atomic file operations with rollback capability
- Extensive integration testing for cross-command consistency

## Future Considerations

### Potential Enhancements (Not in Scope)
- Task dependencies within phases
- Parallel task execution within phases
- Task time tracking and estimation
- Task assignment to different agents
- Bulk task operations

### Integration Points
- **Epic 6:** Handoff reports will heavily utilize event timeline and blocker information
- **Future Development:** Event logging provides foundation for detailed project analytics
- **Quality Metrics:** Test management enables comprehensive quality tracking and reporting