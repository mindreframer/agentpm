# CLI STREAMLINE: Implementation Plan
## Test-Driven Development Approach

### Phase 1: Service Layer Extraction + Tests (High Priority)

#### Phase 1A: Extract Service Logic from Commands
- [ ] Create internal/commands package for shared command logic
- [ ] Extract start-epic logic into internal/commands/start_epic_service.go
- [ ] Extract start-phase logic into internal/commands/start_phase_service.go  
- [ ] Extract start-task logic into internal/commands/start_task_service.go
- [ ] Extract done-epic logic into internal/commands/done_epic_service.go
- [ ] Extract done-phase logic into internal/commands/done_phase_service.go
- [ ] Extract done-task logic into internal/commands/done_task_service.go
- [ ] Extract cancel-task logic into internal/commands/cancel_task_service.go
- [ ] Extract test command logic into internal/commands/test_service.go

#### Phase 1B: Write Service Layer Tests **IMMEDIATELY AFTER 1A**
Service Extraction Test Scenarios:
- [ ] **Test: Start epic service maintains exact same behavior**
- [ ] **Test: Start phase service preserves all validation rules**
- [ ] **Test: Start task service handles all edge cases**
- [ ] **Test: Done epic service completion logic unchanged**
- [ ] **Test: Done phase/task services preserve state transitions**
- [ ] **Test: Cancel task service maintains cancellation rules**
- [ ] **Test: Test services preserve pass/fail/cancel logic**
- [ ] **Test: All services handle error cases identically**

#### Phase 1C: Command Router Foundation
- [ ] Create internal/commands/router.go for subcommand routing
- [ ] Implement type detection logic for entity IDs (phase/task/test)
- [ ] Create argument parsing utilities for unified commands
- [ ] Implement subcommand validation and error handling
- [ ] Create help generation for subcommand structures
- [ ] Unified flag inheritance system for global flags

#### Phase 1D: Write Router Foundation Tests **IMMEDIATELY AFTER 1C**
Router Foundation Test Scenarios:
- [ ] **Test: Type detection correctly identifies phase IDs**
- [ ] **Test: Type detection correctly identifies task IDs**
- [ ] **Test: Type detection correctly identifies test IDs**
- [ ] **Test: Ambiguous ID detection provides clear errors**
- [ ] **Test: Subcommand routing to correct service functions**
- [ ] **Test: Argument parsing handles all valid patterns**
- [ ] **Test: Global flag inheritance works across subcommands**

### Phase 2: Unified Start Command Implementation + Tests (High Priority)

#### Phase 2A: Create Unified Start Command
- [ ] Create cmd/start.go with subcommand structure
- [ ] Implement `start epic` subcommand routing to extracted service
- [ ] Implement `start phase <id>` subcommand with ID validation
- [ ] Implement `start task <id>` subcommand with ID validation
- [ ] Implement `start test <id>` subcommand with ID validation
- [ ] Add type auto-detection for unambiguous IDs
- [ ] Preserve all existing flags (--time, --file, --config, --format)
- [ ] Error handling for invalid subcommands and missing IDs

#### Phase 2B: Write Start Command Tests **IMMEDIATELY AFTER 2A**
Start Command Test Scenarios:
- [ ] **Test: `start epic` produces identical output to `start-epic`**
- [ ] **Test: `start phase 3A` produces identical output to `start-phase 3A`**
- [ ] **Test: `start task T1` produces identical output to `start-task T1`**
- [ ] **Test: `start test T1` produces identical output to `start-test T1`**
- [ ] **Test: `start T1` auto-detects task type correctly**
- [ ] **Test: `start P1` handles ambiguous IDs with clear error**
- [ ] **Test: All existing flags work with new command structure**
- [ ] **Test: Invalid subcommands show helpful error messages**

#### Phase 2C: Unified Done Command Implementation
- [ ] Create cmd/done.go with subcommand structure
- [ ] Implement `done epic` subcommand routing to extracted service
- [ ] Implement `done phase <id>` subcommand with ID validation
- [ ] Implement `done task <id>` subcommand with ID validation
- [ ] Add type auto-detection for unambiguous IDs
- [ ] Preserve all existing completion validation logic
- [ ] Preserve all existing flags and error handling

#### Phase 2D: Write Done Command Tests **IMMEDIATELY AFTER 2C**
Done Command Test Scenarios:
- [ ] **Test: `done epic` produces identical output to `done-epic`**
- [ ] **Test: `done phase 3A` produces identical output to `done-phase 3A`**
- [ ] **Test: `done task T1` produces identical output to `done-task T1`**
- [ ] **Test: `done T1` auto-detects task type correctly**
- [ ] **Test: All validation rules preserved in new structure**
- [ ] **Test: Error messages remain clear and actionable**
- [ ] **Test: Completion events created identically**

### Phase 3: Testing Commands & Cancel Command + Tests (Medium Priority)

#### Phase 3A: Simplified Test Commands Implementation
- [ ] Create cmd/pass.go for simplified test passing
- [ ] Create cmd/fail.go for simplified test failing  
- [ ] Route to extracted test service logic
- [ ] Implement `pass <test-id>` with ID validation
- [ ] Implement `fail <test-id> [reason]` with optional reason
- [ ] Preserve all existing test state transition logic
- [ ] Maintain all existing flags and error handling

#### Phase 3B: Write Test Commands Tests **IMMEDIATELY AFTER 3A**
Test Commands Test Scenarios:
- [ ] **Test: `pass T1` produces identical output to `pass-test T1`**
- [ ] **Test: `fail T1 reason` produces identical output to `fail-test T1 reason`**
- [ ] **Test: Test ID validation works correctly**
- [ ] **Test: Optional reason handling in fail command**
- [ ] **Test: Test state transitions preserved**
- [ ] **Test: Event logging for test status changes**

#### Phase 3C: Unified Cancel Command Implementation
- [ ] Create cmd/cancel.go with subcommand structure
- [ ] Implement `cancel task <id> [reason]` subcommand
- [ ] Implement `cancel test <id> [reason]` subcommand
- [ ] Route to extracted service logic
- [ ] Preserve all existing cancellation validation
- [ ] Maintain optional reason handling

#### Phase 3D: Write Cancel Command Tests **IMMEDIATELY AFTER 3C**
Cancel Command Test Scenarios:
- [ ] **Test: `cancel task T1` produces identical output to `cancel-task T1`**
- [ ] **Test: `cancel test T1 reason` produces identical output to `cancel-test T1 reason`**
- [ ] **Test: Task cancellation rules preserved**
- [ ] **Test: Test cancellation rules preserved**
- [ ] **Test: Reason handling works correctly**
- [ ] **Test: Cancellation events created identically**

### Phase 4: Status Aliases & New Commands + Tests (Medium Priority)

#### Phase 4A: Update Status Command Aliases
- [ ] Update cmd/status.go to change alias from `st` to `s`
- [ ] Update cmd/current.go to change alias from `cur` to `c`
- [ ] Update cmd/pending.go to change alias from `pend` to `p`
- [ ] Update cmd/failing.go to change alias from `fail` to `f`
- [ ] Verify all existing functionality preserved
- [ ] Update help text and command descriptions

#### Phase 4B: Write Status Aliases Tests **IMMEDIATELY AFTER 4A**
Status Aliases Test Scenarios:
- [ ] **Test: `agentpm s` works identical to `agentpm status`**
- [ ] **Test: `agentpm c` works identical to `agentpm current`**
- [ ] **Test: `agentpm p` works identical to `agentpm pending`**
- [ ] **Test: `agentpm f` works identical to `agentpm failing`**
- [ ] **Test: All original functionality preserved**
- [ ] **Test: Help text shows correct aliases**

#### Phase 4C: Implement Show Command
- [ ] Create cmd/show.go for unified entity inspection
- [ ] Implement `show epic` for detailed epic inspection
- [ ] Implement `show phase <id>` for phase details
- [ ] Implement `show task <id>` for task details
- [ ] Implement `show test <id>` for test details
- [ ] Use existing query services for data retrieval
- [ ] Rich text formatting for human-readable output
- [ ] Support all output formats (text/json/xml)

#### Phase 4D: Write Show Command Tests **IMMEDIATELY AFTER 4C**
Show Command Test Scenarios:
- [ ] **Test: `show epic` displays comprehensive epic information**
- [ ] **Test: `show phase 3A` displays detailed phase information**
- [ ] **Test: `show task T1` displays comprehensive task details**
- [ ] **Test: `show test T1` displays test details and history**
- [ ] **Test: All output formats work correctly**
- [ ] **Test: Invalid IDs show helpful error messages**
- [ ] **Test: Non-existent entities handled gracefully**

### Phase 5: Rename & Cleanup + Integration Tests (Low Priority)

#### Phase 5A: Rename Next Command & Final Cleanup
- [ ] Rename cmd/start_next.go to cmd/next.go
- [ ] Update command registration in main.go
- [ ] Remove all old command files (start-epic.go, start-phase.go, etc.)
- [ ] Update main.go to register only new command structure
- [ ] Clean up unused imports and dependencies
- [ ] Update all command help text and usage examples

#### Phase 5B: Write Rename & Cleanup Tests **IMMEDIATELY AFTER 5A**
Cleanup Test Scenarios:
- [ ] **Test: `next` command works identical to `start-next`**
- [ ] **Test: All old command files successfully removed**
- [ ] **Test: Main.go registers only new commands**
- [ ] **Test: No unused imports or dependencies remain**
- [ ] **Test: Help system shows only new command structure**

#### Phase 5C: Comprehensive Integration Testing
- [ ] End-to-end workflow testing with new command structure
- [ ] Cross-command consistency verification (error formats, outputs)
- [ ] Global flag handling across all commands
- [ ] Configuration file compatibility testing
- [ ] Help system completeness and accuracy
- [ ] Performance testing to ensure no regressions
- [ ] Memory usage verification

#### Phase 5D: Write Integration Tests **IMMEDIATELY AFTER 5C**
Integration Test Scenarios:
- [ ] **Test: Complete epic lifecycle using new commands**
- [ ] **Test: Mixed command workflows (start → status → done)**
- [ ] **Test: All global flags work across command structure**
- [ ] **Test: Configuration integration works correctly**
- [ ] **Test: Help system provides complete guidance**
- [ ] **Test: Performance matches original commands**
- [ ] **Test: Error handling consistency across all commands**

## Development Workflow Per Phase

For **EACH** phase:

1. **Implement Code** (Phase XA or XC)
2. **Write Tests IMMEDIATELY** (Phase XB or XD) 
3. **Run Tests & Verify** - All tests must pass
4. **Run Linting/Type Checking** - Code must be clean
5. **NEVER move to next phase with failing tests**

## CLI Streamline Specific Considerations

### Technical Requirements
- **No Backward Compatibility:** Complete breaking change, no old command support
- **Service Layer Preservation:** All business logic extracted and preserved
- **Identical Outputs:** New commands must produce identical results
- **Type Detection:** Smart ID detection for unambiguous cases
- **Unified Patterns:** Consistent argument and flag patterns
- **Error Consistency:** Unified error messaging and help system

### File Structure Changes
```
├── cmd/
│   ├── start.go           # Unified start command with subcommands
│   ├── done.go            # Unified done command with subcommands
│   ├── cancel.go          # Unified cancel command with subcommands
│   ├── pass.go            # Simplified pass command
│   ├── fail.go            # Simplified fail command
│   ├── show.go            # New unified inspection command
│   ├── next.go            # Renamed from start_next.go
│   ├── status.go          # Updated alias (s)
│   ├── current.go         # Updated alias (c)
│   ├── pending.go         # Updated alias (p)
│   ├── failing.go         # Updated alias (f)
│   ├── init.go            # Unchanged
│   ├── switch.go          # Unchanged
│   ├── config.go          # Unchanged
│   ├── validate.go        # Unchanged
│   ├── events.go          # Unchanged
│   ├── docs.go            # Unchanged
│   ├── handoff.go         # Unchanged
│   └── version.go         # Unchanged
├── internal/
│   └── commands/          # Extracted service logic
│       ├── router.go      # Subcommand routing and type detection
│       ├── start_epic_service.go
│       ├── start_phase_service.go
│       ├── start_task_service.go
│       ├── done_epic_service.go
│       ├── done_phase_service.go
│       ├── done_task_service.go
│       ├── cancel_task_service.go
│       └── test_service.go
```

### Command Structure Implementation
```go
// cmd/start.go example
func StartCommand() *cli.Command {
    return &cli.Command{
        Name:    "start",
        Usage:   "Start working on something",
        Subcommands: []*cli.Command{
            {
                Name:   "epic",
                Usage:  "Start working on current epic",
                Action: commands.StartEpicAction,
            },
            {
                Name:      "phase",
                Usage:     "Start working on specific phase",
                ArgsUsage: "<phase-id>",
                Action:    commands.StartPhaseAction,
            },
            {
                Name:      "task", 
                Usage:     "Start working on specific task",
                ArgsUsage: "<task-id>",
                Action:    commands.StartTaskAction,
            },
            {
                Name:      "test",
                Usage:     "Start test execution",
                ArgsUsage: "<test-id>",
                Action:    commands.StartTestAction,
            },
        },
        // Fallback action for type auto-detection
        Action: commands.StartWithTypeDetection,
    }
}
```

## Benefits of This Approach

✅ **Zero Regression Risk** - Service logic extracted first, tested immediately  
✅ **Immediate Feedback** - Each command tested as soon as implemented  
✅ **Preserved Functionality** - All existing behavior maintained exactly  
✅ **Clean Breaking Change** - No backward compatibility complexity  
✅ **Incremental Progress** - Each phase delivers working functionality  
✅ **Quality Assurance** - No untested code makes it to later phases  

## Test Distribution Summary

- **Phase 1 Tests:** 15 scenarios (Service extraction, router foundation)
- **Phase 2 Tests:** 16 scenarios (Start/Done unified commands)
- **Phase 3 Tests:** 12 scenarios (Test commands, Cancel command)
- **Phase 4 Tests:** 14 scenarios (Status aliases, Show command)
- **Phase 5 Tests:** 12 scenarios (Cleanup, Integration testing)

**Total: 69 test scenarios covering complete CLI transformation**

---

## Implementation Status

### CLI STREAMLINE: BREAKING CHANGE IMPLEMENTATION - PENDING
### Current Status: NOT STARTED

### Progress Tracking
- [ ] Phase 1A: Extract Service Logic from Commands
- [ ] Phase 1B: Write Service Layer Tests
- [ ] Phase 1C: Command Router Foundation
- [ ] Phase 1D: Write Router Foundation Tests
- [ ] Phase 2A: Create Unified Start Command
- [ ] Phase 2B: Write Start Command Tests
- [ ] Phase 2C: Unified Done Command Implementation
- [ ] Phase 2D: Write Done Command Tests
- [ ] Phase 3A: Simplified Test Commands Implementation
- [ ] Phase 3B: Write Test Commands Tests
- [ ] Phase 3C: Unified Cancel Command Implementation
- [ ] Phase 3D: Write Cancel Command Tests
- [ ] Phase 4A: Update Status Command Aliases
- [ ] Phase 4B: Write Status Aliases Tests
- [ ] Phase 4C: Implement Show Command
- [ ] Phase 4D: Write Show Command Tests
- [ ] Phase 5A: Rename Next Command & Final Cleanup
- [ ] Phase 5B: Write Rename & Cleanup Tests
- [ ] Phase 5C: Comprehensive Integration Testing
- [ ] Phase 5D: Write Integration Tests

### Definition of Done
- [ ] All existing functionality accessible through new command structure
- [ ] New commands produce identical outputs to old functionality
- [ ] All global flags work consistently across commands
- [ ] Error handling maintains current quality and helpfulness
- [ ] 25% reduction in total command count achieved (20 → 15)
- [ ] Consistent short aliases for all frequent commands
- [ ] Logical command grouping reduces learning curve
- [ ] Test coverage maintained at >90% for all new commands
- [ ] No performance regressions in command execution time
- [ ] Help system clearly shows new command patterns