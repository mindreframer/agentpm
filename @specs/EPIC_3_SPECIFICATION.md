# EPIC-3 SPECIFICATION: Epic Lifecycle Management

## Overview

**Epic ID:** 3  
**Name:** Epic Lifecycle Management  
**Duration:** 3-4 days  
**Status:** pending  
**Priority:** high  
**Depends On:** Epic 1 (Foundation & Configuration), Epic 2 (Query & Status Commands)

**Goal:** Implement epic-level status transitions and project switching capabilities to enable agents to manage epic lifecycles from start to completion.

## Business Context

Epic 3 provides agents with the ability to control epic lifecycles through simple state transitions. Unlike complex project management systems, AgentPM focuses on the essential states that matter for LLM agents: starting work (pending → wip), and completing work (wip → done). The epic is designed to be lightweight and follow the principle that agents work sequentially on epics without complex pause/resume workflows.

## User Stories

### Primary User Stories
- **As an agent, I can switch between different epic files** so that I can work on multiple projects without configuration conflicts
- **As an agent, I can start working on a new epic** so that I can begin tracking my development activities
- **As an agent, I can complete an epic when all work is done** so that I can formally close completed projects
- **As an agent, I can view epic lifecycle events** so that I understand the timeline and state transitions

### Secondary User Stories
- **As an agent, I can validate epic completion requirements** so that I don't mark incomplete work as done
- **As an agent, I can see automatic event logging** so that all lifecycle changes are tracked for handoff purposes
- **As an agent, I can work with deterministic timestamps** so that testing and snapshots are consistent

## Technical Requirements

### Core Dependencies
- **Foundation:** Epic 1 CLI framework, XML processing, storage interface, configuration management
- **Querying:** Epic 2 query service for epic validation and status checking
- **Event Logging:** Automatic event creation for all lifecycle transitions
- **Timestamp Control:** Support for --time flag for deterministic testing

### Architecture Principles
- **Simple State Machine:** Only essential transitions (pending → wip → done)
- **Validation-First:** All state transitions validated before execution
- **Event Logging:** Automatic logging of all lifecycle changes
- **Atomic Operations:** File operations are atomic to prevent corruption
- **No Complex States:** No pause/resume/cancel - keep it simple for agents

### Epic Status Lifecycle
```
pending ──start-epic──→ wip ──done-epic──→ done
   ↑                              ↑
   │                              │
   └──── switch-epic ─────────────┘
   (switches to different epic file)
```

### Status Transition Rules
- **pending → wip:** Always allowed, sets started timestamp
- **wip → done:** Only allowed when all phases complete and no failing tests
- **done → *:** No transitions allowed from done state
- **switch-epic:** Changes current_epic in config, doesn't affect epic status

## Functional Requirements

### FR-1: Epic Startup
**Command:** `agentpm start-epic [--time <timestamp>] [--file <epic-file>]`

**Behavior:**
- Changes epic status from "pending" to "wip" 
- Sets started_at timestamp (from --time flag or current time)
- Validates epic is in "pending" status before transition
- Creates automatic event log entry for epic startup
- Updates epic XML file with new status and timestamp
- Fails gracefully if epic is already started or completed

**Output Format:**
```xml
<epic_started epic="8">
    <previous_status>pending</previous_status>
    <new_status>wip</new_status>
    <started_at>2025-08-16T15:30:00Z</started_at>
    <message>Epic 8 started. Status changed to wip.</message>
</epic_started>
```

### FR-2: Epic Completion
**Command:** `agentpm done-epic [--time <timestamp>] [--file <epic-file>]`

**Behavior:**
- Changes epic status from "wip" to "done"
- Sets completed_at timestamp (from --time flag or current time)
- Validates all phases are completed before allowing transition
- Validates no tests have "failing" status before allowing transition
- Creates automatic event log entry for epic completion
- Calculates and displays epic duration and summary statistics
- Fails with clear error if incomplete work exists

**Output Format:**
```xml
<epic_completed epic="8">
    <previous_status>wip</previous_status>
    <new_status>done</new_status>
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

### FR-3: Project Switching
**Command:** `agentpm switch <epic-file> [--time <timestamp>]`

**Behavior:**
- Updates current_epic in .agentpm.json configuration
- Validates new epic file exists and is readable
- Does not modify epic status - only changes project context
- Shows previous and new epic for confirmation
- Creates event log entry in both epics (if applicable)
- Handles non-existent files with clear error messages

**Output Format:**
```xml
<switch_result>
    <previous_epic>epic-8.xml</previous_epic>
    <current_epic>epic-9.xml</current_epic>
    <updated>.agentpm.json</updated>
    <message>Switched to epic-9.xml successfully</message>
</switch_result>
```

### FR-4: Lifecycle Validation
**Internal Functionality:** Used by start-epic and done-epic commands

**Start Epic Validation:**
- Epic status must be "pending"
- Epic XML must be valid and well-formed
- No additional prerequisites required

**Complete Epic Validation:**
- Epic status must be "wip" 
- All phases must have status "completed"
- No tests can have status "failing"
- Epic must have been started (has started_at timestamp)

**Validation Error Format:**
```xml
<validation_error epic="8">
    <error_type>incomplete_work</error_type>
    <message>Cannot complete epic with pending work</message>
    <details>
        <pending_phases>
            <phase id="3A" name="LiveView Integration"/>
        </pending_phases>
        <failing_tests>
            <test id="2A_2" description="Mobile responsive design"/>
        </failing_tests>
    </details>
</validation_error>
```

### FR-5: Automatic Event Logging
**Behavior:** All lifecycle commands automatically create event entries

**Event Types:**
- `epic_started`: When epic transitions from pending to wip
- `epic_completed`: When epic transitions from wip to done  
- `epic_switched`: When switching between epic files

**Event Format:**
```xml
<event timestamp="2025-08-16T15:30:00Z" agent="agent_claude" type="epic_started">
    Epic 8 started
    
    Status: pending → wip
    Started at: 2025-08-16T15:30:00Z
</event>
```

## Non-Functional Requirements

### NFR-1: Performance
- Lifecycle commands execute in < 200ms for typical epic files
- File operations are atomic to prevent corruption during transitions
- Validation performs efficiently without loading unnecessary data

### NFR-2: Reliability
- All file operations are transactional (backup before modify)
- Validation prevents invalid state transitions
- Clear error messages for all failure conditions
- Rollback capability if XML writing fails

### NFR-3: Consistency
- Timestamp handling consistent across all commands
- Event logging format consistent with other epic events
- Error message format consistent with other commands
- XML output schema consistent across lifecycle operations

### NFR-4: Testing Support
- --time flag allows deterministic timestamp injection
- All operations work with MemoryStorage for isolated testing
- State transitions fully testable with controlled epic data

## Data Model Changes

### Epic XML Status Fields
```xml
<epic id="8" name="Epic Name" status="wip" created_at="2025-08-15T09:00:00Z">
    <started_at>2025-08-16T15:30:00Z</started_at>
    <completed_at>2025-08-20T16:45:00Z</completed_at>
    <!-- optional, only when status="done" -->
    
    <!-- existing epic content -->
    <phases>...</phases>
    <tasks>...</tasks>
    <tests>...</tests>
    <events>
        <!-- lifecycle events automatically added -->
        <event timestamp="2025-08-16T15:30:00Z" type="epic_started">...</event>
    </events>
</epic>
```

### Configuration Changes
```json
{
    "current_epic": "epic-9.xml",
    "project_name": "MooCRM",
    "default_assignee": "agent_claude",
    "previous_epic": "epic-8.xml"
}
```

## Error Handling

### Error Categories
1. **Status Transition Errors:** Invalid state changes, already started/completed
2. **Validation Errors:** Incomplete work, failing tests, missing phases
3. **File Access Errors:** Missing epic files, permission issues, corrupt XML
4. **Configuration Errors:** Invalid config file, missing current_epic

### Error Response Examples

**Invalid State Transition:**
```xml
<error>
    <type>invalid_transition</type>
    <message>Epic is already started</message>
    <details>
        <current_status>wip</current_status>
        <started_at>2025-08-15T09:00:00Z</started_at>
        <suggestion>Use 'agentpm current' to see active work</suggestion>
    </details>
</error>
```

**Incomplete Work Error:**
```xml
<error>
    <type>incomplete_work</type>
    <message>Cannot complete epic with pending work</message>
    <details>
        <pending_phases_count>2</pending_phases_count>
        <failing_tests_count>1</failing_tests_count>
        <suggestion>Use 'agentpm pending' and 'agentpm failing' to see remaining work</suggestion>
    </details>
</error>
```

## Acceptance Criteria

### AC-1: Start Epic from Pending Status
- **GIVEN** I have an epic with status "pending"
- **WHEN** I run `agentpm start-epic`
- **THEN** the epic status should change to "wip" and a start event should be logged

### AC-2: Prevent Starting Already Started Epic
- **GIVEN** I have an epic with status "wip"
- **WHEN** I run `agentpm start-epic`
- **THEN** I should get an error that epic is already started

### AC-3: Complete Epic with All Work Done
- **GIVEN** I have an epic with all phases completed and all tests passing
- **WHEN** I run `agentpm done-epic`
- **THEN** status should change to "done" and completion event should be logged

### AC-4: Prevent Completing Epic with Pending Work
- **GIVEN** I have an epic with pending tasks
- **WHEN** I run `agentpm done-epic`
- **THEN** I should get an error listing the pending work that must be completed

### AC-5: Prevent Completing Epic with Failing Tests
- **GIVEN** I have an epic with failing tests
- **WHEN** I run `agentpm done-epic`
- **THEN** I should get an error listing the failing tests that must be fixed

### AC-6: Switch Between Epic Files
- **GIVEN** I have current_epic set to "epic-8.xml"
- **WHEN** I run `agentpm switch epic-9.xml`
- **THEN** the config should update to current_epic: "epic-9.xml"

### AC-7: Handle Non-Existent Epic File Switch
- **GIVEN** I specify an epic file that doesn't exist
- **WHEN** I run `agentpm switch missing-epic.xml`
- **THEN** I should get an error that the epic file doesn't exist

### AC-8: Deterministic Timestamp Support
- **GIVEN** I provide a specific timestamp
- **WHEN** I run `agentpm start-epic --time "2025-08-16T15:30:00Z"`
- **THEN** the started_at timestamp should be exactly "2025-08-16T15:30:00Z"

## Testing Strategy

### Test Categories
- **Unit Tests (70%):** State transition logic, validation rules, event creation
- **Integration Tests (25%):** File operations, configuration updates, end-to-end workflows
- **Edge Case Tests (5%):** Error conditions, malformed data, invalid states

### Test Data Requirements
- Epic files in various states (pending, wip, done)
- Epic files with complete and incomplete work
- Epic files with failing tests
- Multiple epic files for switching scenarios

### Test Isolation
- Each test uses isolated epic files in `t.TempDir()`
- MemoryStorage for fast unit tests
- Deterministic timestamps for consistent snapshots
- No shared state between tests

## Implementation Phases

### Phase 3A: Lifecycle Service Foundation (Day 1)
- Create internal/lifecycle package
- Implement LifecycleService with Storage injection
- Epic status validation logic
- State transition rules and validation
- Event creation utilities

### Phase 3B: Start Epic Command (Day 1-2)
- `agentpm start-epic` command implementation
- Status transition from pending to wip
- Timestamp handling (--time flag support)
- Automatic event logging for epic startup
- XML file updates and error handling

### Phase 3C: Complete Epic Command (Day 2-3)
- `agentpm done-epic` command implementation
- Completion validation (all phases done, no failing tests)
- Status transition from wip to done
- Epic summary calculation and display
- Comprehensive error reporting for incomplete work

### Phase 3D: Switch Epic Command & Integration (Day 3-4)
- `agentpm switch` command implementation
- Configuration file updates
- Epic file validation and error handling
- Integration testing across all lifecycle commands
- Performance optimization and error message refinement

## Definition of Done

- [ ] All acceptance criteria verified with automated tests
- [ ] Lifecycle commands execute in < 200ms for typical epic files
- [ ] Test coverage > 90% for lifecycle logic
- [ ] All error cases handled gracefully with clear messages
- [ ] Automatic event logging works for all lifecycle transitions
- [ ] Timestamp injection (--time flag) works for deterministic testing
- [ ] File operations are atomic and safe from corruption
- [ ] Integration tests verify end-to-end lifecycle workflows

## Dependencies and Risks

### Dependencies
- **Epic 1:** CLI framework, XML processing, storage interface, configuration management
- **Epic 2:** Query service for epic validation and status checking
- **Event System:** Event creation and logging utilities

### Risks
- **Low Risk:** File corruption during state transitions
- **Low Risk:** Validation logic complexity for completion requirements
- **Low Risk:** Configuration file handling edge cases

### Mitigation Strategies
- Atomic file operations with backup/restore capability
- Comprehensive validation testing with various epic states
- Clear error messages and graceful degradation
- Extensive integration testing for configuration changes

## Future Considerations

### Potential Enhancements (Not in Scope)
- Epic archiving and restoration capabilities
- Bulk epic operations across multiple files
- Epic templates for common project types
- Integration with external project management systems

### Integration Points
- **Epic 4:** Task management will build on lifecycle state validation
- **Epic 5:** Event logging will extend the automatic event creation patterns
- **Epic 6:** Handoff reports will include lifecycle timeline information