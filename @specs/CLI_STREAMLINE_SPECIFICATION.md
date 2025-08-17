# CLI STREAMLINE SPECIFICATION: Unified Command Interface

## Overview

**Epic ID:** CLI-Streamline  
**Name:** CLI Interface Streamlining  
**Duration:** 5-7 days  
**Status:** pending  
**Priority:** high  
**Depends On:** Existing CLI infrastructure

**Goal:** Redesign the AgentPM CLI interface to provide a more intuitive and streamlined user experience by consolidating 18+ individual commands into 12 logical command groups with consistent syntax and behavior.

## Business Context

The current AgentPM CLI has grown organically to 18+ individual commands with inconsistent patterns, making it difficult for users to remember and navigate. The new interface groups related functionality into logical command families while maintaining all existing functionality. This streamlining improves usability without sacrificing power or flexibility.

## Current vs. Desired Interface Mapping

### CORE WORKFLOW Commands
| Desired | Current Commands | Changes Required |
|---------|------------------|------------------|
| `start [epic\|phase\|task\|test]` | `start-epic`, `start-phase`, `start-task`, various test commands | **MAJOR**: Consolidate into single `start` command with subcommands |
| `done [epic\|phase\|task]` | `done-epic`, `done-phase`, `done-task` | **MAJOR**: Consolidate into single `done` command with subcommands |
| `next` | `start-next` | **MINOR**: Rename existing command |

### STATUS Commands  
| Desired | Current Commands | Changes Required |
|---------|------------------|------------------|
| `status, s` | `status` | **NONE**: Already exists with alias |
| `current, c` | `current` | **MINOR**: Add alias `c` |
| `pending, p` | `pending` | **MINOR**: Add alias `p` |
| `failing, f` | `failing` | **MINOR**: Add alias `f` |

### INSPECTION Commands
| Desired | Current Commands | Changes Required |
|---------|------------------|------------------|
| `show <type> [id]` | Multiple status/query commands | **MAJOR**: New unified inspection interface |

### PROJECT Commands
| Desired | Current Commands | Changes Required |
|---------|------------------|------------------|
| `init` | `init` | **NONE**: Already exists |
| `switch <epic>` | `switch` | **MINOR**: Simplify arguments |
| `config` | `config` | **NONE**: Already exists |

### REPORTING Commands
| Desired | Current Commands | Changes Required |
|---------|------------------|------------------|
| `events` | `events`, `log` | **MAJOR**: Consolidate event viewing and logging |
| `docs` | `docs` | **NONE**: Already exists |
| `handoff` | `handoff` | **NONE**: Already exists |

### REMOVED Commands
| Current Command | Removal Reason | Replacement |
|----------------|----------------|-------------|
| `validate` | Incorporated into other commands | Automatic validation in all operations |
| `version` | Standard CLI practice | `agentpm --version` global flag |
| `cancel-task` | Rarely used, complex workflow | Not included in streamlined interface |

## Technical Requirements

### Core Dependencies
- **Existing CLI Framework:** urfave/cli/v3 with current flag structure
- **Command Routing:** New command multiplexer for subcommand routing
- **Backward Compatibility:** Optional support for legacy command names
- **Service Layer:** Existing service implementations remain unchanged

### Architecture Principles
- **Command Consolidation:** Group related commands under common verbs
- **Consistent Syntax:** All commands follow `verb [object] [args]` pattern
- **Preserved Functionality:** Zero loss of existing capabilities
- **Maintainable Routing:** Clean separation between command parsing and business logic
- **Extensible Design:** Easy to add new subcommands to existing command groups

## Functional Requirements

### FR-1: Unified Start Command
**Command:** `agentpm start [epic|phase|task|test] [id] [options]`

**Behavior:**
- Route to appropriate existing start command based on first argument
- Maintain all existing functionality and flags
- Support type inference when ID is provided without explicit type
- Provide helpful error messages for ambiguous cases

**Examples:**
```bash
agentpm start epic          # start-epic equivalent
agentpm start phase 3A      # start-phase 3A equivalent  
agentpm start task T1       # start-task T1 equivalent
agentpm start test TEST1    # start-test TEST1 equivalent
agentpm start T1            # Auto-detect task T1
```

**Implementation:**
- New `cmd/start.go` that routes to existing start-* command implementations
- Preserve all existing flags: `--time`, `--file`, `--config`, `--format`
- Type detection logic based on ID patterns and epic content analysis

### FR-2: Unified Done Command  
**Command:** `agentpm done [epic|phase|task] [id] [options]`

**Behavior:**
- Route to appropriate existing done command based on first argument
- Maintain all existing validation and completion logic
- Support type inference for task/phase IDs
- Generate same output formats as existing commands

**Examples:**
```bash
agentpm done epic           # done-epic equivalent
agentpm done phase 3A       # done-phase 3A equivalent
agentpm done task T1        # done-task T1 equivalent  
agentpm done T1             # Auto-detect task T1
```

**Implementation:**
- New `cmd/done.go` that routes to existing done-* command implementations
- Preserve all existing validation rules and error handling
- Maintain automatic state transitions and event logging

### FR-3: Enhanced Status Commands
**Commands:** `status|s`, `current|c`, `pending|p`, `failing|f`

**Behavior:**
- Add missing aliases to existing commands
- Maintain all existing functionality unchanged
- Ensure consistent output formatting across all status commands

**Implementation:**
- Update existing command definitions to include new aliases
- No business logic changes required

### FR-4: Unified Show Command
**Command:** `agentpm show <epic|phase|task|test> [id] [options]`

**Behavior:**
- Provide detailed inspection interface for any entity type
- Show comprehensive information including status, dependencies, history
- Support both specific ID lookup and current context inspection
- Output structured data in text/json/xml formats

**Examples:**
```bash
agentpm show epic           # Show current epic details
agentpm show phase 3A       # Show specific phase details
agentpm show task T1        # Show specific task details  
agentpm show test TEST1     # Show specific test details
```

**Implementation:**
- New comprehensive inspection service
- Leverage existing query services for data retrieval
- Rich formatting for human-readable output

### FR-5: Simplified Switch Command
**Command:** `agentpm switch <epic-file>`

**Behavior:**
- Simplify to single required argument (epic file path)
- Remove complex flag-based switching options
- Maintain same validation and configuration update logic
- Keep `--back` flag for returning to previous epic

**Examples:**
```bash
agentpm switch epic-9.xml   # Switch to epic-9.xml
agentpm switch --back       # Switch to previous epic
```

**Implementation:**
- Simplify argument parsing in existing switch command
- Remove unused flags and complex argument handling

### FR-6: Consolidated Events Command
**Command:** `agentpm events [log <message>] [options]`

**Behavior:**
- Default: Show recent events (existing events command functionality)
- With `log` subcommand: Add new event (existing log command functionality)  
- Maintain all existing event logging capabilities
- Preserve event viewing and filtering options

**Examples:**
```bash
agentpm events                      # Show recent events
agentpm events --limit 10           # Show last 10 events
agentpm events log "Fixed bug X"    # Log new event
agentpm events log --type milestone "Phase 1 complete"
```

**Implementation:**
- New events command that routes between viewing and logging
- Preserve all existing functionality from both events and log commands

## Non-Functional Requirements

### NFR-1: Performance
- Command routing overhead < 5ms
- Same execution performance as existing individual commands
- No impact on memory usage or startup time

### NFR-2: Maintainability
- Clean separation between new routing logic and existing business logic
- Preserve existing test coverage and test infrastructure
- New routing code should be unit testable in isolation

### NFR-3: User Experience
- Intuitive command discovery with help system
- Consistent error messages and help text across all commands
- Auto-completion support for shells

## Data Model Changes

### Command Structure Changes
```go
// New command structure in main.go
Commands: []*cli.Command{
    // CORE WORKFLOW  
    cmd.NewStartCommand(),     // Replaces start-epic, start-phase, start-task, test commands
    cmd.NewDoneCommand(),      // Replaces done-epic, done-phase, done-task  
    cmd.NextCommand(),         // Renamed from start-next
    
    // STATUS (enhanced with aliases)
    cmd.StatusCommand(),       // Add alias 's'
    cmd.CurrentCommand(),      // Add alias 'c'  
    cmd.PendingCommand(),      // Add alias 'p'
    cmd.FailingCommand(),      // Add alias 'f'
    
    // INSPECTION
    cmd.ShowCommand(),         // New unified inspection interface
    
    // PROJECT
    cmd.InitCommand(),         // Unchanged
    cmd.SwitchCommand(),       // Simplified arguments
    cmd.ConfigCommand(),       // Unchanged
    
    // REPORTING  
    cmd.EventsCommand(),       // Consolidates events + log
    cmd.DocsCommand(),         // Unchanged
    cmd.HandoffCommand(),      // Unchanged
}
```

### Configuration Changes
No changes to .agentpm.json structure required.

### Help System Enhancement
```
agentpm help                 # Show command overview
agentpm start --help         # Show start subcommand options  
agentpm done --help          # Show done subcommand options
agentpm show --help          # Show inspection options
```

## Error Handling

### Command Routing Errors
```xml
<error>
    <type>invalid_subcommand</type>
    <message>Unknown subcommand 'foobar' for start command</message>
    <details>
        <available_subcommands>
            <subcommand>epic</subcommand>
            <subcommand>phase</subcommand>
            <subcommand>task</subcommand>
            <subcommand>test</subcommand>
        </available_subcommands>
        <suggestion>Use 'agentpm start --help' to see available options</suggestion>
    </details>
</error>
```

### Type Detection Ambiguity
```xml
<error>
    <type>ambiguous_identifier</type>
    <message>ID 'P1' could refer to phase or task</message>
    <details>
        <matches>
            <match type="phase" id="P1" name="Planning Phase"/>
            <match type="task" id="P1" name="Project Setup"/>
        </matches>
        <suggestion>Specify type explicitly: 'agentpm start phase P1' or 'agentpm start task P1'</suggestion>
    </details>
</error>
```

## Acceptance Criteria

### AC-1: Start Command Consolidation
- **GIVEN** I want to start any type of work
- **WHEN** I use `agentpm start <type> [id]`  
- **THEN** it should route to the appropriate existing start command with same behavior

### AC-2: Done Command Consolidation  
- **GIVEN** I want to complete any type of work
- **WHEN** I use `agentpm done <type> [id]`
- **THEN** it should route to the appropriate existing done command with same behavior

### AC-3: Auto-Detection Works
- **GIVEN** I provide an unambiguous ID
- **WHEN** I use `agentpm start T1` or `agentpm done T1`
- **THEN** it should detect the type automatically and execute correctly

### AC-4: Status Aliases Work  
- **GIVEN** I want quick status access
- **WHEN** I use `agentpm s`, `agentpm c`, `agentpm p`, or `agentpm f`
- **THEN** they should work the same as the full command names

### AC-5: Show Command Provides Rich Information
- **GIVEN** I want detailed information about an entity
- **WHEN** I use `agentpm show <type> [id]`  
- **THEN** I should get comprehensive details about that entity

### AC-6: Events Command Handles Both Viewing and Logging
- **GIVEN** I want to work with events
- **WHEN** I use `agentpm events` I should see recent events
- **WHEN** I use `agentpm events log <message>` I should add a new event

### AC-7: Help System Works
- **GIVEN** I need help with commands
- **WHEN** I use `agentpm help` or `agentpm <command> --help`  
- **THEN** I should get clear guidance on available options

## Testing Strategy

### Test Categories
- **Unit Tests (60%):** Command routing logic, argument parsing, type detection
- **Integration Tests (30%):** End-to-end command execution, backward compatibility  
- **User Experience Tests (10%):** Help system, error messages, auto-completion

### Test Data Requirements
- Epic files with various entity types for type detection testing
- Test cases for ambiguous IDs and edge cases
- Complete command execution scenarios for all consolidated commands

### Test Isolation
- Mock existing command implementations for pure routing tests
- Integration tests with full command stack
- Regression tests to ensure no functionality loss

## Implementation Phases

### Phase 1: Command Consolidation Foundation (Day 1-2)
- Create new routing infrastructure for `start` and `done` commands
- Implement type detection logic for entity IDs
- Build argument parsing and validation for consolidated commands
- Unit tests for routing logic

### Phase 2: Status and Inspection Commands (Day 2-3)  
- Add aliases to existing status commands
- Implement new `show` command with comprehensive entity inspection
- Enhance help system for new command structure
- Integration testing for status workflows

### Phase 3: Events and Project Commands (Day 3-4)
- Consolidate events viewing and logging into single command
- Simplify switch command argument handling  
- Update help system and error messages
- Complete integration testing

### Phase 4: Polish and Final Testing (Day 4-5)
- Enhance error messages and user guidance
- Performance optimization and final testing
- Update all documentation and examples
- Final user acceptance testing and release preparation

## Definition of Done

- [ ] All existing functionality preserved in new command structure
- [ ] Command routing performance < 5ms overhead
- [ ] Comprehensive help system for all new commands  
- [ ] Type detection works for 95% of unambiguous cases
- [ ] All acceptance criteria verified with automated tests

- [ ] Zero regression in existing functionality
- [ ] User experience testing completed with positive feedback

## Dependencies and Risks

### Dependencies
- **Existing CLI Infrastructure:** urfave/cli/v3 framework and all current command implementations
- **Service Layer:** All existing service implementations must remain unchanged
- **Test Infrastructure:** Current test framework and test data

### Risks
- **Medium Risk:** Complex type detection logic for ambiguous IDs
- **Low Risk:** Performance impact from command routing overhead

### Mitigation Strategies
- Implement type detection as fallback mechanism with explicit type preference
- Performance testing throughout implementation to catch issues early

## Future Considerations

### Potential Enhancements (Not in Scope)
- Auto-completion scripts for popular shells
- Interactive command mode for complex workflows
- Command aliases and custom shortcuts
- Plugin system for custom commands

### Integration Points
- **Shell Integration:** Future auto-completion and shell integration features
- **IDE Integration:** Structured command interface could support IDE plugins
- **API Layer:** Command structure could inform future REST API design
- **Automation:** Simplified interface better supports scripting and automation

## Migration Strategy

### For Existing Users
1. **Immediate:** New interface available alongside existing commands
2. **Month 1-3:** Deprecation warnings on old commands
3. **Month 3-6:** Old commands still work but marked as deprecated in help
4. **Month 6+:** Consider removing old commands (major version bump)

### For Scripts and Automation
- Provide explicit flag to use legacy command names without warnings
- Document migration path for automated systems
- Maintain stable output formats during transition period

## Success Metrics

- **User Experience:** Reduced time to discover and execute commands
- **Learning Curve:** New users can become productive faster
- **Maintainability:** Cleaner command structure reduces maintenance overhead  
- **Adoption:** Positive feedback from existing users on new interface
- **Performance:** No measurable impact on command execution speed