# EPIC-13 SPECIFICATION: Status Enum Streamlining

## Overview

**Epic ID:** 13  
**Name:** Status Enum Streamlining  
**Duration:** 3-4 days  
**Status:** pending  
**Priority:** high  

**Goal:** Streamline all status enums across epics, phases, tasks, and tests to a unified, consistent set of values with proper validation rules and business logic enforcement.

## Business Context

The current codebase has inconsistent status values across different entities, creating confusion and potential bugs. Multiple status enums exist (`epic.Status`, `TestStatus`) with overlapping values but different meanings. This epic will establish a unified status system that enforces business rules and provides clear transitions.

## User Stories

### Primary User Stories
- **As an agent, I can rely on consistent status values** across all entities so that I can build predictable automation
- **As an agent, I receive clear error messages** when attempting invalid status transitions so that I understand what actions are allowed
- **As an agent, I can see accurate counts** of pending/wip items when validation fails so that I know exactly what needs to be completed
- **As a developer, I can maintain the status system easily** with unified enums and validation logic so that the codebase remains consistent

### Secondary User Stories
- **As an agent, I can mark failing tests as cancelled with a reason** so that I can document why a test was abandoned
- **As an agent, I can change a passing test back to failing** when I discover issues so that I can iterate on test validation

## Technical Requirements

### Unified Status Enums

#### Epic Status
```go
type EpicStatus string

const (
    EpicStatusPending EpicStatus = "pending"
    EpicStatusWIP     EpicStatus = "wip" 
    EpicStatusDone    EpicStatus = "done"
)
```

#### Phase Status  
```go
type PhaseStatus string

const (
    PhaseStatusPending PhaseStatus = "pending"
    PhaseStatusWIP     PhaseStatus = "wip"
    PhaseStatusDone    PhaseStatus = "done"
)
```

#### Task Status
```go
type TaskStatus string

const (
    TaskStatusPending   TaskStatus = "pending"
    TaskStatusWIP       TaskStatus = "wip" 
    TaskStatusDone      TaskStatus = "done"
    TaskStatusCancelled TaskStatus = "cancelled"
)
```

#### Test Status & Result
```go
type TestStatus string

const (
    TestStatusPending   TestStatus = "pending"
    TestStatusWIP       TestStatus = "wip"
    TestStatusDone      TestStatus = "done" 
    TestStatusCancelled TestStatus = "cancelled"
)

type TestResult string

const (
    TestResultPassing TestResult = "passing"
    TestResultFailing TestResult = "failing"
)
```

### Business Rules Implementation

#### Phase Completion Rules
- **Rule:** A phase cannot be "done" if it has pending OR wip tasks
- **Rule:** A phase cannot be "done" if it has pending OR wip tests  
- **Error Message:** Must display the NUMBER of pending/wip tasks or tests

#### Task Completion Rules  
- **Rule:** A task cannot be "done" if it has pending OR wip tests
- **Error Message:** Must display the NUMBER of pending/wip tests

#### Test State Rules
- **Rule:** A failing test cannot be marked as "done" - it can only be cancelled with a reason
- **Rule:** A "passing" AND "done" test can be changed to "failing" and "wip"

## Functional Requirements

### FR-1: Status Validation Framework
**Component:** Enhanced validation service

**Behavior:**
- Validates all status transitions according to business rules
- Provides detailed error messages with counts
- Prevents invalid state transitions
- Returns structured validation results

**Implementation:**
```go
type StatusValidationError struct {
    EntityType     string            // "phase", "task", "test"
    EntityID       string
    CurrentStatus  string
    TargetStatus   string
    BlockingItems  []BlockingItem    
    Message        string
}

type BlockingItem struct {
    Type   string  // "task", "test"
    ID     string
    Name   string
    Status string
}
```

### FR-2: Phase Status Enforcement
**Commands:** `agentpm done phase <phase-id>`

**Behavior:**
- Counts pending/wip tasks in the phase
- Counts pending/wip tests for tasks in the phase
- Blocks completion if any blocking items exist
- Returns detailed error with counts

**Error Output:**
```xml
<error>
    <type>phase_completion_blocked</type>
    <message>Phase cannot be completed: 2 pending tasks, 1 wip test</message>
    <blocking_items>
        <tasks count="2">
            <task id="task-1" name="Implement feature" status="pending"/>
            <task id="task-2" name="Add validation" status="wip"/>
        </tasks>
        <tests count="1">
            <test id="test-1" name="Unit test" status="wip"/>
        </tests>
    </blocking_items>
</error>
```

### FR-3: Task Status Enforcement  
**Commands:** `agentpm done task <task-id>`

**Behavior:**
- Counts pending/wip tests for the task
- Blocks completion if any tests are not done
- Returns detailed error with test counts

**Error Output:**
```xml
<error>
    <type>task_completion_blocked</type>
    <message>Task cannot be completed: 3 pending tests</message>
    <blocking_items>
        <tests count="3">
            <test id="test-1" name="Happy path test" status="pending"/>
            <test id="test-2" name="Error case test" status="pending"/> 
            <test id="test-3" name="Integration test" status="wip"/>
        </tests>
    </blocking_items>
</error>
```

### FR-4: Simplified Test CLI Commands
**Commands:** 
- `agentpm pass <test-id>` → status: done, result: passing
- `agentpm fail <test-id>` → status: wip, result: failing  
- `agentpm cancel test <test-id> --reason "<reason>"` → status: cancelled, result: failing

**Behavior:**
- Easy transitions between pass/fail states
- Tests can only be modified when their task belongs to the current active phase
- Cancelled tests require a reason
- No complex validation - simple state changes

**Examples:**
```bash
# Mark test as passing
agentpm pass A4_1_1

# Mark test as failing  
agentpm fail A4_1_1

# Cancel test with reason
agentpm cancel test A4_1_1 --reason "Test case no longer relevant"

# Easy transitions - can go from pass to fail and back
agentpm pass A4_1_1    # status: done, result: passing
agentpm fail A4_1_1    # status: wip, result: failing  
agentpm pass A4_1_1    # status: done, result: passing
```

### FR-4.1: Batch Test Status Commands
**Commands:**
- `agentpm pass-batch <test-id1> <test-id2> <test-id3>...` → batch mark tests as passing
- `agentpm fail-batch <test-id1> <test-id2> <test-id3>...` → batch mark tests as failing

**Behavior:**
- Pre-validates all test IDs exist and their parent tasks belong to current active phase
- Pre-validates all status transitions are possible according to business rules
- If ANY validation fails, NO changes are made and detailed error is returned
- Only executes batch changes if ALL tests can be successfully updated
- Returns summary of all changes made or comprehensive error details

**Validation Rules:**
- All test IDs must exist in the current epic
- All tests must belong to tasks in the current active phase
- All tests must have valid status for the requested transition
- Tests marked as cancelled cannot be changed without explicit cancellation reversal

**Error Handling:**
```xml
<error>
    <type>batch_validation_failed</type>
    <message>Batch operation failed: 2 invalid test IDs, 1 test not in active phase</message>
    <failed_tests count="3">
        <test id="A1_1_5" error="test_not_found" message="Test ID A1_1_5 does not exist"/>
        <test id="A1_2_1" error="wrong_phase" message="Test belongs to task in phase 'done', not current active phase 'wip'"/>
        <test id="A1_3_1" error="cancelled_test" message="Test is cancelled and cannot be modified without explicit cancellation reversal"/>
    </failed_tests>
    <valid_tests count="1">
        <test id="A1_1_1" current_status="wip" target_status="done"/>
    </valid_tests>
</error>
```

**Success Output:**
```xml
<batch_result>
    <type>batch_pass_success</type>
    <message>Successfully marked 3 tests as passing</message>
    <updated_tests count="3">
        <test id="A1_1_1" old_status="wip" new_status="done" old_result="failing" new_result="passing"/>
        <test id="A1_1_2" old_status="pending" new_status="done" old_result="failing" new_result="passing"/>
        <test id="A1_1_3" old_status="wip" new_status="done" old_result="failing" new_result="passing"/>
    </updated_tests>
</batch_result>
```

**Examples:**
```bash
# Batch mark multiple tests as passing
agentpm pass-batch A1_1_1 A1_1_2 A1_1_3

# Batch mark multiple tests as failing
agentpm fail-batch A1_1_1 A1_1_2 A1_1_3

# Mixed scenarios - all or nothing
agentpm pass-batch A1_1_1 A1_1_2 INVALID_ID  # Fails entirely, no changes made
agentpm pass-batch A1_1_1 A2_1_1              # Fails if A2_1_1 not in active phase
```



## Non-Functional Requirements



### NFR-2: Performance
- Status validation adds < 10ms to command execution
- Bulk validation operations scale linearly with entity count
- Memory usage remains constant regardless of epic size

### NFR-3: Error Message Quality
- All error messages include specific counts
- Error messages provide actionable suggestions
- Validation errors are structured for programmatic consumption

## Data Model Changes

### Epic XML Schema Updates
```xml
<epic id="13" name="Status Streamlining" status="pending">
    <phases>
        <phase id="phase-1" status="pending">
            <!-- Phase content -->
        </phase>
    </phases>
    <tasks>
        <task id="task-1" phase_id="phase-1" status="pending">
            <!-- Task content -->
        </task>
    </tasks>
    <tests>
        <test id="test-1" task_id="task-1" status="pending" result="failing">
            <!-- Test content -->
        </test>
    </tests>
</epic>
```

### Go Struct Updates
```go
type Epic struct {
    // ... existing fields
    Status EpicStatus `xml:"status,attr"`
}

type Phase struct {
    // ... existing fields  
    Status PhaseStatus `xml:"status,attr"`
}

type Task struct {
    // ... existing fields
    Status TaskStatus `xml:"status,attr"`
}

type Test struct {
    // ... existing fields
    Status TestStatus `xml:"status,attr"`
    Result TestResult `xml:"result,attr"`
}
```

## Error Handling

### Validation Error Types
1. **Phase Completion Blocked:** Pending/wip tasks or tests prevent completion
2. **Task Completion Blocked:** Pending/wip tests prevent completion  
3. **Invalid Test Transition:** Failing tests cannot be marked done
4. **Status Transition Invalid:** General invalid status transitions

### Error Response Format
```xml
<validation_error>
    <entity_type>phase</entity_type>
    <entity_id>phase-1</entity_id>
    <current_status>wip</current_status>
    <target_status>done</target_status>
    <blocking_count>3</blocking_count>
    <blocking_items>
        <item type="task" id="task-1" status="pending"/>
        <item type="test" id="test-1" status="wip"/>
    </blocking_items>
    <message>Phase cannot be completed: 1 pending task, 2 wip tests</message>
    <suggestions>
        <suggestion>Complete task-1 first</suggestion>
        <suggestion>Finish tests before marking phase done</suggestion>
    </suggestions>
</validation_error>
```

## Acceptance Criteria

### AC-1: Phase Completion Validation
- **GIVEN** a phase with pending tasks
- **WHEN** I run `agentpm done-phase phase-1`
- **THEN** I should get an error with the exact count of pending tasks

### AC-2: Task Completion Validation  
- **GIVEN** a task with wip tests
- **WHEN** I run `agentpm done-task task-1`
- **THEN** I should get an error with the exact count of wip tests

### AC-3: Test Pass Command
- **GIVEN** any test in the current active phase
- **WHEN** I run `agentpm pass test-1`
- **THEN** the test should have status="done" and result="passing"

### AC-4: Test Fail Command  
- **GIVEN** any test in the current active phase
- **WHEN** I run `agentpm fail test-1`
- **THEN** the test should have status="wip" and result="failing"

### AC-5: Test Cancellation
- **GIVEN** any test
- **WHEN** I run `agentpm cancel test test-1 --reason "No longer needed"`  
- **THEN** the test should have status="cancelled" and result="failing"

### AC-6: Easy Pass/Fail Transitions
- **GIVEN** a test with status="done" and result="passing"
- **WHEN** I run `agentpm fail test-1`  
- **THEN** the test should transition to status="wip" and result="failing"
- **AND WHEN** I run `agentpm pass test-1`
- **THEN** the test should transition back to status="done" and result="passing"

### AC-7: Batch Pass Command Success
- **GIVEN** tests A1_1_1, A1_1_2, A1_1_3 exist and belong to tasks in current active phase
- **WHEN** I run `agentpm pass-batch A1_1_1 A1_1_2 A1_1_3`
- **THEN** all three tests should have status="done" and result="passing"
- **AND** I should receive a success message with details of all changes

### AC-8: Batch Pass Command Validation Failure
- **GIVEN** test A1_1_5 does not exist
- **WHEN** I run `agentpm pass-batch A1_1_1 A1_1_2 A1_1_5`
- **THEN** NO tests should be modified
- **AND** I should receive an error message explaining the invalid test ID
- **AND** the error should list which tests were valid vs invalid

### AC-9: Batch Fail Command Success  
- **GIVEN** tests A1_1_1, A1_1_2, A1_1_3 exist and belong to tasks in current active phase
- **WHEN** I run `agentpm fail-batch A1_1_1 A1_1_2 A1_1_3`
- **THEN** all three tests should have status="wip" and result="failing"
- **AND** I should receive a success message with details of all changes

### AC-10: Batch Command Phase Validation
- **GIVEN** test A2_1_1 belongs to a task in phase "done" (not current active phase)
- **WHEN** I run `agentpm pass-batch A1_1_1 A2_1_1`
- **THEN** NO tests should be modified
- **AND** I should receive an error explaining A2_1_1 is not in the active phase
- **AND** the error should show which phase A2_1_1 belongs to vs current active phase



### AC-6: Validation Error Details
- **GIVEN** any blocked completion attempt
- **WHEN** validation fails
- **THEN** I should see the exact number and list of blocking items

## Testing Strategy

### Test Categories
- **Unit Tests (70%):** Status validation logic, business rules, migration logic
- **Integration Tests (20%):** Command execution with status validation
- **End-to-End Tests (10%):** Full workflows with status transitions

### Test Coverage Areas
- All status transition validations
- Error message formatting and counts
- Batch command validation and execution
- Edge cases (empty phases, orphaned tests, etc.)

## Implementation Phases

### Phase 13A: Status Enum Definition (Day 1)
- Define unified status enums for all entity types
- Create status validation framework
- Implement transition rule engine
- Add validation error types

### Phase 13B: Business Rule Implementation (Day 2)
- Implement phase completion validation
- Implement task completion validation  
- Implement test status rules
- Add detailed error messaging with counts

### Phase 13C: Command Updates (Day 2-3)
- Update existing commands to use new validation
- Update test files and snapshots for new status system
- Validate all existing functionality still works

### Phase 13D: Integration & Testing (Day 3-4)
- Comprehensive test coverage for all rules
- Performance testing for validation operations
- Error message refinement and consistency
- Documentation updates

## Definition of Done

- [ ] All entity types use unified status enums
- [ ] Business rules enforced with exact counts in error messages
- [ ] Failing tests cannot be marked as "done" without cancellation

- [ ] All tests pass with new status system
- [ ] Performance impact < 10ms per validation
- [ ] Test coverage > 95% for status validation logic
- [ ] All error messages provide actionable guidance
- [ ] Documentation updated with new status values

## Dependencies and Risks

### Dependencies
- Requires completion of current epic XML structure
- All existing commands must be updated to use new validation

### Risks
- **Medium Risk:** Performance impact of additional validation
- **Medium Risk:** Batch operations complexity with rollback scenarios

### Mitigation Strategies
- Performance benchmarking during development
- Comprehensive testing of batch operations
- Extensive validation of status transition logic

## Notes

- This epic represents a significant refactoring that touches most of the codebase
- The unified status system will make future development much more consistent
- Batch operations provide efficiency for agents managing multiple tests
- Error messages must be extremely clear since agents rely on them for automation