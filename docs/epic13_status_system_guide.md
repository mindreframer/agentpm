# Epic 13 Status System Guide

## Overview

Epic 13 introduces a unified status system that streamlines all status enums across epics, phases, tasks, and tests. This system provides consistent status values with proper validation rules and business logic enforcement.

## Unified Status Enums

### Epic Status
```go
type EpicStatus string

const (
    EpicStatusPending EpicStatus = "pending"
    EpicStatusWIP     EpicStatus = "wip" 
    EpicStatusDone    EpicStatus = "done"
)
```

### Phase Status  
```go
type PhaseStatus string

const (
    PhaseStatusPending PhaseStatus = "pending"
    PhaseStatusWIP     PhaseStatus = "wip"
    PhaseStatusDone    PhaseStatus = "done"
)
```

### Task Status
```go
type TaskStatus string

const (
    TaskStatusPending   TaskStatus = "pending"
    TaskStatusWIP       TaskStatus = "wip" 
    TaskStatusDone      TaskStatus = "done"
    TaskStatusCancelled TaskStatus = "cancelled"
)
```

### Test Status & Result
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

## Status Transitions

### Valid Transitions

#### Epic Status Transitions
- `pending` → `wip`
- `wip` → `done`
- `done` is terminal (no further transitions)

#### Phase Status Transitions
- `pending` → `wip`
- `wip` → `done`
- `done` is terminal (no further transitions)

#### Task Status Transitions
- `pending` → `wip` or `cancelled`
- `wip` → `done` or `cancelled`
- `done` is terminal (no further transitions)
- `cancelled` is terminal (no further transitions)

#### Test Status Transitions
- `pending` → `wip` or `cancelled`
- `wip` → `done` or `cancelled`
- `done` → `wip` (can go back to WIP for failing tests)
- `cancelled` is terminal (no further transitions)

### Business Rules

#### Phase Completion Rules
- A phase cannot be marked as "done" if it has pending OR wip tasks
- A phase cannot be marked as "done" if it has pending OR wip tests
- Error messages display the exact NUMBER of pending/wip tasks or tests

#### Task Completion Rules  
- A task cannot be marked as "done" if it has pending OR wip tests
- Error messages display the exact NUMBER of pending/wip tests

#### Test State Rules
- A failing test cannot be marked as "done" - it can only be cancelled with a reason
- A "passing" AND "done" test can be changed to "failing" and "wip" for re-testing

## CLI Commands

### Simple Test Commands

#### Pass Command
```bash
agentpm pass <test-id>
```
- Sets status: `done`, result: `passing`
- Example: `agentpm pass A4_1_1`

#### Fail Command
```bash
agentpm fail <test-id>
```
- Sets status: `wip`, result: `failing`
- Example: `agentpm fail A4_1_1`

#### Cancel Command
```bash
agentpm cancel test <test-id> --reason "<reason>"
```
- Sets status: `cancelled`, result: `failing`
- Requires a reason for cancellation
- Example: `agentpm cancel test A4_1_1 --reason "Test case no longer relevant"`

#### Easy Pass/Fail Transitions
```bash
# Mark test as passing
agentpm pass A4_1_1    # status: done, result: passing

# Mark test as failing  
agentpm fail A4_1_1    # status: wip, result: failing  

# Mark test as passing again
agentpm pass A4_1_1    # status: done, result: passing
```

### Batch Commands

#### Batch Pass Command
```bash
agentpm pass-batch <test-id1> <test-id2> <test-id3>...
```
- Batch mark multiple tests as passing
- All-or-nothing validation: if ANY test fails validation, NO changes are made
- Example: `agentpm pass-batch A1_1_1 A1_1_2 A1_1_3`

#### Batch Fail Command
```bash
agentpm fail-batch <test-id1> <test-id2> <test-id3>...
```
- Batch mark multiple tests as failing
- All-or-nothing validation: if ANY test fails validation, NO changes are made
- Example: `agentpm fail-batch A1_1_1 A1_1_2 A1_1_3`

#### Batch Validation Rules
- All test IDs must exist in the current epic
- All tests must belong to tasks in the current active phase
- All tests must have valid status for the requested transition
- Tests marked as cancelled cannot be changed without explicit cancellation reversal

## Error Messages

### Phase Completion Error
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

### Task Completion Error
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

### Batch Operation Error
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

### Batch Operation Success
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

## Performance Characteristics

### Validation Performance
- Status validation adds < 10ms to command execution
- Bulk validation operations scale linearly with entity count
- Memory usage remains constant regardless of epic size

### Benchmarks
Based on integration testing with Epic 13:
- Average validation time: ~100μs for moderate-sized epics (200 tasks, 400 tests)
- Large epic validation: ~300μs for large epics (500 tasks, 500 tests)
- Memory usage: Constant across different epic sizes

## Troubleshooting Guide

### Common Issues

#### Issue: "Invalid test status" errors
**Cause:** Test has an invalid legacy status value  
**Solution:** The Epic 13 system gracefully handles invalid legacy statuses by converting them to valid defaults. This should not normally occur with new epics.

#### Issue: "Phase cannot be completed" errors
**Cause:** Phase has pending or wip tasks/tests  
**Solution:** Complete all tasks and tests in the phase before marking phase as done
```bash
# Check phase status
agentpm status

# Complete individual tests
agentpm pass <test-id>

# Complete multiple tests at once
agentpm pass-batch <test-id1> <test-id2> <test-id3>
```

#### Issue: "Task cannot be completed" errors
**Cause:** Task has pending or wip tests  
**Solution:** Complete all tests for the task before marking task as done
```bash
# List pending tests for task
agentpm pending

# Complete tests
agentpm pass-batch <test-id1> <test-id2>
```

#### Issue: "Batch operation failed" errors
**Cause:** One or more tests in batch have validation issues  
**Solution:** Review error details and fix issues before retrying
- Check test IDs exist
- Ensure tests belong to current active phase
- Verify test status allows the requested transition

#### Issue: Performance issues with large epics
**Cause:** Epic size may be affecting validation performance  
**Solution:** Monitor validation times and consider epic decomposition
```bash
# Run validation with timing
time agentpm validate
```

### Debug Commands

#### Check Current Status
```bash
agentpm status
```

#### Validate Epic Structure
```bash
agentpm validate
```

#### List Pending Items
```bash
agentpm pending
```

#### List Failing Items
```bash
agentpm failing
```

## Data Model Changes

### XML Schema
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

### Backward Compatibility

The Epic 13 status system maintains backward compatibility with existing epics:

- Legacy status values are automatically converted to Epic 13 equivalents
- Invalid legacy statuses default to "pending"
- Existing commands continue to work with status validation

#### Legacy Status Mapping
- `planning` → `pending`
- `active` → `wip`
- `completed` → `done`
- Invalid values → `pending`

## Best Practices

### For Agents
1. **Use batch commands** for efficiency when updating multiple tests
2. **Check status before operations** to avoid validation errors
3. **Provide reasons for cancellations** to maintain audit trail
4. **Follow the workflow**: pending → wip → done
5. **Handle validation errors** by checking counts and completing prerequisites

### For Development
1. **Always run validation** after making status changes
2. **Use specific error messages** to guide troubleshooting
3. **Test performance** with realistic epic sizes
4. **Monitor validation timing** in production
5. **Follow status transition rules** strictly

### Example Workflows

#### Complete a Task
```bash
# 1. Check current status
agentpm status

# 2. Complete all tests for the task
agentpm pass-batch test1 test2 test3

# 3. Mark task as done
agentpm done task task-1

# 4. Verify completion
agentpm status
```

#### Complete a Phase
```bash
# 1. Complete all tasks in phase
agentpm done task task-1
agentpm done task task-2

# 2. Mark phase as done
agentpm done phase phase-1

# 3. Verify completion
agentpm status
```

#### Handle Failing Tests
```bash
# 1. Identify failing tests
agentpm failing

# 2. Either fix and pass, or cancel with reason
agentpm pass test-1  # If fixed
agentpm cancel test test-2 --reason "Test case obsolete"

# 3. Continue with workflow
agentpm done task task-1
```

## Configuration Options

### Status Validation Settings
Currently, status validation is always enabled. Future versions may support:
- Validation level configuration (strict/permissive)
- Performance tuning options
- Custom status transition rules

### Logging Controls
Status validation operations are logged at appropriate levels:
- Info: Successful status transitions
- Warn: Performance concerns (>10ms validation)
- Error: Validation failures and business rule violations

## Migration Guide

### From Pre-Epic 13 Systems
1. **Backup your epic files** before migration
2. **Run validation** to identify any issues
3. **Update commands** to use new simplified syntax
4. **Test workflows** with the new status system
5. **Monitor performance** after migration

### Command Changes
- `agentpm pass <test-id>` (unchanged)
- `agentpm fail <test-id>` (unchanged)
- `agentpm pass-batch <test-ids>` (new)
- `agentpm fail-batch <test-ids>` (new)
- `agentpm cancel test <test-id> --reason "<reason>"` (enhanced)

### Status Field Changes
- Tests now have separate `status` and `result` fields
- All entities use consistent status values
- Legacy statuses are automatically converted

This guide provides comprehensive coverage of the Epic 13 status system. For additional help, refer to the CLI help system or contact support.