# Epic 3: Epic Lifecycle Management - Specification

## Overview
**Goal:** Manage epic creation, status transitions, and project switching  
**Duration:** 3-4 days  
**Philosophy:** Controlled state machine with comprehensive event logging and validation

## User Stories
1. Start working on a new epic with proper state transition
2. Pause work when blocked or interrupted with reason logging
3. Resume paused work and continue progress tracking
4. Switch between different epic files for multi-epic workflows
5. Complete an epic when all work is done with validation

## Technical Requirements
- **Dependencies:** Epic 1 (CLI, config, validation) + Epic 2 (status analysis)
- **State Machine:** Epic status lifecycle with transition validation
- **Event Logging:** Timestamp tracking for all state transitions
- **Atomic Operations:** Safe file updates with rollback capability
- **Validation Rules:** Enforce valid status transitions and completion criteria

## Epic Status Lifecycle

### Status Values & Transitions
```go
type EpicStatus string

const (
    StatusPlanning   EpicStatus = "planning"
    StatusInProgress EpicStatus = "in_progress" 
    StatusPaused     EpicStatus = "paused"
    StatusCompleted  EpicStatus = "completed"
    StatusCancelled  EpicStatus = "cancelled"
)

type StatusTransition struct {
    From      EpicStatus
    To        EpicStatus
    Valid     bool
    Condition string // Optional condition for transition
}
```

### Valid Transition Matrix
```
planning → in_progress (always valid)
in_progress → paused (always valid)  
in_progress → completed (requires all work done)
in_progress → cancelled (always valid)
paused → in_progress (always valid)
paused → cancelled (always valid)
completed → [no transitions] (terminal state)
cancelled → [no transitions] (terminal state)
```

### Event Logging System
```go
type LifecycleEvent struct {
    Type        string    `xml:"type,attr"`        // "epic_started", "epic_paused", etc.
    Timestamp   time.Time `xml:"timestamp,attr"`
    Agent       string    `xml:"agent,attr"`
    PreviousStatus string `xml:"previous_status,attr,omitempty"`
    NewStatus   string    `xml:"new_status,attr,omitempty"`
    Reason      string    `xml:"reason,attr,omitempty"`
    Duration    string    `xml:"duration,attr,omitempty"` // For resume operations
    Message     string    `xml:",chardata"`
}
```

## Implementation Phases

### Phase 3A: State Machine Engine (1 day)
- Epic status transition validation engine
- Status transition matrix implementation
- Pre-condition checking for transitions
- Atomic epic file updates with rollback
- Status change validation and enforcement

### Phase 3B: Event Logging System (1 day)
- Lifecycle event creation and logging
- Timestamp management and formatting
- Event serialization and XML integration
- Event history management within epic files
- Agent attribution and context tracking

### Phase 3C: Epic Startup Operations (0.5 days)
- `agentpm start-epic` command implementation
- Epic initialization and status transition
- Start event logging and timestamp recording
- Validation for already started epics
- Error handling and user feedback

### Phase 3D: Pause & Resume Operations (1 day)
- `agentpm pause-epic [reason]` command implementation
- `agentpm resume-epic` command implementation
- Pause duration calculation and tracking
- Reason logging for pause operations
- Resume validation and state restoration

### Phase 3E: Epic Completion & Project Switching (0.5 days)
- `agentpm complete-epic` command implementation
- `agentpm switch <epic-file>` command implementation
- Completion validation (all work done)
- Configuration file updates for project switching
- Epic file validation before switching

## Acceptance Criteria
- ✅ `agentpm start-epic` changes status from planning to in_progress
- ✅ `agentpm pause-epic "reason"` pauses with optional reason logging
- ✅ `agentpm resume-epic` resumes from paused state
- ✅ `agentpm switch epic-9.xml` updates current_epic in config
- ✅ `agentpm complete-epic` marks epic as completed with validation
- ✅ Status transitions are validated (can't resume non-paused epic)
- ✅ All lifecycle changes create timestamped events

## Validation Rules

### Epic Start Validation
1. **Current Status:** Must be "planning"
2. **Epic Structure:** Must have valid phases and tasks
3. **Not Already Started:** Cannot start epic that's already in progress

### Epic Pause Validation
1. **Current Status:** Must be "in_progress"
2. **Valid Epic:** Epic must exist and be parseable
3. **Not Already Paused:** Cannot pause already paused epic

### Epic Resume Validation
1. **Current Status:** Must be "paused"
2. **Valid Epic:** Epic must exist and be parseable
3. **Not Active:** Cannot resume non-paused epic

### Epic Completion Validation
1. **Current Status:** Must be "in_progress"
2. **All Phases Complete:** All phases must have status "completed"
3. **All Tests Passing:** No tests with status "failing"
4. **No Pending Work:** All tasks must be completed

### Project Switch Validation
1. **Target File Exists:** Epic file must exist and be readable
2. **Valid Epic XML:** Target epic must pass validation
3. **Configuration Update:** Must successfully update .agentpm.json

## Output Examples

### agentpm start-epic
```xml
<epic_started epic="8">
    <previous_status>planning</previous_status>
    <new_status>in_progress</new_status>
    <started_at>2025-08-16T15:30:00Z</started_at>
    <message>Epic 8 started. Status changed to in_progress.</message>
</epic_started>
```

### agentpm pause-epic "Waiting for design approval"
```xml
<epic_paused epic="8">
    <previous_status>in_progress</previous_status>
    <new_status>paused</new_status>
    <paused_at>2025-08-16T15:30:00Z</paused_at>
    <reason>Waiting for design approval</reason>
    <message>Epic 8 paused. Status changed to paused.</message>
</epic_paused>
```

### agentpm resume-epic
```xml
<epic_resumed epic="8">
    <previous_status>paused</previous_status>
    <new_status>in_progress</new_status>
    <resumed_at>2025-08-17T09:00:00Z</resumed_at>
    <pause_duration>17 hours 30 minutes</pause_duration>
    <message>Epic 8 resumed. Status changed to in_progress.</message>
</epic_resumed>
```

### agentpm complete-epic
```xml
<epic_completed epic="8">
    <previous_status>in_progress</previous_status>
    <new_status>completed</new_status>
    <completed_at>2025-08-20T16:45:00Z</completed_at>
    <summary>
        <total_phases>4</total_phases>
        <total_tasks>12</total_tasks>
        <total_tests>25</total_tests>
        <duration>5 days</duration>
    </summary>
    <message>Epic 8 completed successfully. All phases and tests complete.</message>
</epic_completed>
```

### agentpm switch epic-9.xml
```xml
<switch_result>
    <previous_epic>epic-8.xml</previous_epic>
    <current_epic>epic-9.xml</current_epic>
    <updated>.agentpm.json</updated>
</switch_result>
```

## State Machine Operations

### Epic Start Operation
```go
func StartEpic(epic *Epic) (*LifecycleEvent, error) {
    // 1. Validate current status is "planning"
    // 2. Update epic status to "in_progress"  
    // 3. Create start event with timestamp
    // 4. Save epic file atomically
    // 5. Return success event
}
```

### Epic Pause Operation
```go
func PauseEpic(epic *Epic, reason string) (*LifecycleEvent, error) {
    // 1. Validate current status is "in_progress"
    // 2. Update epic status to "paused"
    // 3. Create pause event with reason and timestamp
    // 4. Save epic file atomically
    // 5. Return success event
}
```

### Epic Resume Operation
```go
func ResumeEpic(epic *Epic) (*LifecycleEvent, error) {
    // 1. Validate current status is "paused"
    // 2. Calculate pause duration from last pause event
    // 3. Update epic status to "in_progress"
    // 4. Create resume event with duration
    // 5. Save epic file atomically
    // 6. Return success event
}
```

### Epic Complete Operation
```go
func CompleteEpic(epic *Epic) (*LifecycleEvent, error) {
    // 1. Validate current status is "in_progress"
    // 2. Validate all phases are completed
    // 3. Validate all tests are passing
    // 4. Calculate epic duration and summary
    // 5. Update epic status to "completed"
    // 6. Create completion event
    // 7. Save epic file atomically
    // 8. Return success event
}
```

## File Operations & Atomicity

### Atomic Epic Updates
1. **Read Current Epic:** Load and parse current epic file
2. **Validate Operation:** Check transition validity and preconditions
3. **Create Backup:** Save backup of current epic file
4. **Update Epic:** Modify epic status and add event
5. **Write Atomically:** Write to temporary file, then rename
6. **Cleanup:** Remove backup on success, restore on failure

### Configuration Updates
1. **Load Config:** Read current .agentpm.json
2. **Validate Target:** Ensure target epic file exists and is valid
3. **Update Config:** Modify current_epic field
4. **Save Atomically:** Write config with atomic operation
5. **Verify:** Confirm config update success

## Error Handling

### Invalid Transition Errors
```xml
<error type="invalid_transition">
    <current_status>completed</current_status>
    <attempted_action>start-epic</attempted_action>
    <message>Cannot start epic that is already completed</message>
    <valid_actions>
        <action>switch to different epic</action>
    </valid_actions>
</error>
```

### Completion Validation Errors
```xml
<error type="completion_blocked">
    <current_status>in_progress</current_status>
    <blockers>
        <pending_tasks>
            <task id="2A_2" phase_id="2A">Add accessibility features</task>
        </pending_tasks>
        <failing_tests>
            <test id="2A_3" phase_id="2A">Mobile responsive test</test>
        </failing_tests>
    </blockers>
    <message>Epic cannot be completed. Fix failing tests and complete pending tasks.</message>
</error>
```

## Test Scenarios (Key Examples)
- **Epic Start:** Start from planning status, handle already started epics
- **Epic Pause:** Pause with/without reason, handle non-active epics
- **Epic Resume:** Resume from paused state, calculate duration, handle invalid states
- **Epic Complete:** Complete with validation, handle incomplete work
- **Project Switch:** Switch between valid epic files, handle missing files
- **State Validation:** Enforce transition rules, prevent invalid operations
- **Event Logging:** Create timestamped events for all lifecycle changes

## Integration with Previous Epics

### Epic 1 Integration
- **Epic Loading:** Use Epic 1 epic loading and validation
- **Configuration:** Extend Epic 1 config system for project switching
- **Storage:** Leverage Epic 1 storage abstraction for atomic operations
- **Error Handling:** Build on Epic 1 error handling patterns

### Epic 2 Integration
- **Status Analysis:** Use Epic 2 progress calculation for completion validation
- **Pending Work:** Use Epic 2 pending work discovery for completion checks
- **Current State:** Integrate with Epic 2 current state analysis

## Quality Gates
- [ ] All acceptance criteria implemented and tested
- [ ] State transition matrix enforced correctly
- [ ] Atomic file operations prevent data corruption
- [ ] Comprehensive event logging for all lifecycle changes
- [ ] Validation prevents invalid epic completion

## Performance Considerations
- **Atomic Operations:** File operations complete within 100ms
- **Event Logging:** Minimal overhead for event creation
- **Validation:** Completion validation within 200ms for large epics
- **Configuration Updates:** Config file updates within 50ms

This specification provides robust epic lifecycle management while maintaining data integrity and comprehensive audit trails through event logging.