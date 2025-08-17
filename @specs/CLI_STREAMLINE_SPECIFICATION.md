# CLI STREAMLINE SPECIFICATION

## Overview

**Goal:** Redesign the AgentPM CLI interface to provide a more intuitive, unified command structure that reduces cognitive load and improves developer productivity through better command organization and simplified argument patterns.

**Duration:** 5-7 days  
**Priority:** high  
**Epic Dependencies:** All existing epics (1-9) will be affected by this refactoring

## Business Context

The current CLI has grown organically with 20+ commands using inconsistent naming patterns, scattered aliases, and varying argument structures. This creates confusion for users and increases the learning curve. The streamlined interface consolidates related operations under unified command verbs, introduces consistent short aliases, and establishes clear argument patterns that make the tool more intuitive and faster to use.

## Current State Analysis

### Current CLI Surface (20 commands)
| Category | Current Commands | Issues |
|----------|------------------|---------|
| **Start Operations** | `start-epic`, `start-phase`, `start-task`, `start-next`, `start-test` | Inconsistent naming, missing unified verb |
| **Completion Operations** | `done-epic`, `done-phase`, `done-task` | No unified completion command |
| **Test Operations** | `pass-test`, `fail-test`, `cancel-test` | Scattered across CLI, inconsistent patterns |
| **Status Operations** | `status`, `current`, `pending`, `failing`, `events` | Some have aliases (`st`, `cur`), others don't |
| **Project Operations** | `init`, `switch`, `validate`, `config` | Good organization |
| **Reporting** | `handoff`, `docs`, `log` | Missing `show` inspection |
| **System** | `version` | Has aliases (`ver`, `v`) |

### Command Naming Issues
- **Verb-Object vs Object-Verb:** Mix of `start-epic` vs `epic-start`
- **Inconsistent Aliases:** Only some commands have short aliases
- **Missing Unification:** Related operations spread across multiple commands
- **Argument Inconsistency:** Some take `<id>`, others use flags, some take no args

## Desired CLI Interface Analysis

### New Command Structure (15 commands, 25% reduction)
| Category | New Commands | Consolidation |
|----------|--------------|---------------|
| **CORE WORKFLOW** | `start`, `done`, `cancel`, `next` | Unified verbs |
| **TESTING** | `pass`, `fail` | Simple test operations |
| **STATUS** | `status`, `current`, `pending`, `failing` | Consistent short aliases |
| **INSPECTION** | `show` | New unified inspection command |
| **PROJECT** | `init`, `switch`, `config`, `validate` | Maintained |
| **REPORTING** | `events`, `docs`, `handoff` | Maintained |
| **SYSTEM** | `version`, `help` | Standard system commands |

### Key Improvements
1. **Unified Verbs:** `start [epic|phase|task|test]` instead of separate commands
2. **Consistent Aliases:** Every frequent command gets a short alias
3. **Argument Standardization:** Clear patterns for IDs, options, and flags
4. **Reduced Commands:** 20 → 15 commands (25% reduction)
5. **Logical Grouping:** Related operations under common verbs

## Detailed Command Mapping

### CORE WORKFLOW COMMANDS

#### `start [epic|phase|task|test] [id] [options]`
**Current Mapping:**
- `start-epic` → `start epic`
- `start-phase <id>` → `start phase <id>`
- `start-task <id>` → `start task <id>`
- `start-next` → `next` (separate command for auto-progression)
- `start-test <id>` → `start test <id>`

**New Behavior:**
```bash
# Epic operations (no ID needed, uses current epic)
agentpm start epic [--time <timestamp>]

# Phase/Task/Test operations (ID required)
agentpm start phase <phase-id> [--time <timestamp>]
agentpm start task <task-id> [--time <timestamp>]
agentpm start test <test-id> [--time <timestamp>]

# Auto-progression remains separate
agentpm next [--time <timestamp>]
```

#### `done [epic|phase|task] [id] [options]`
**Current Mapping:**
- `done-epic` → `done epic`
- `done-phase <id>` → `done phase <id>`
- `done-task <id>` → `done task <id>`

**New Behavior:**
```bash
# Epic completion (no ID needed)
agentpm done epic [--time <timestamp>]

# Phase/Task completion (ID required)
agentpm done phase <phase-id> [--time <timestamp>]
agentpm done task <task-id> [--time <timestamp>]
```

#### `cancel [task|test] <id> [reason] [options]`
**Current Mapping:**
- `cancel-task <id>` → `cancel task <id> [reason]`
- `cancel-test <id> "reason"` → `cancel test <id> [reason]`

**New Behavior:**
```bash
# Task/Test cancellation with optional reason
agentpm cancel task <task-id> [reason] [--time <timestamp>]
agentpm cancel test <test-id> [reason] [--time <timestamp>]
```

#### `next [options]`
**Current Mapping:**
- `start-next` → `next`

**New Behavior:**
```bash
# Auto-start next available work
agentpm next [--time <timestamp>]
```

### TESTING COMMANDS

#### `pass <test-id> [options]`
**Current Mapping:**
- `pass-test <id>` → `pass <id>`

**New Behavior:**
```bash
agentpm pass <test-id> [--time <timestamp>]
```

#### `fail <test-id> [reason] [options]`
**Current Mapping:**
- `fail-test <id> "reason"` → `fail <id> [reason]`

**New Behavior:**
```bash
agentpm fail <test-id> [reason] [--time <timestamp>]
```

### STATUS COMMANDS (with new aliases)

#### `status, s [options]`
**Current Mapping:**
- `status` (alias: `st`) → `status` (alias: `s`)

#### `current, c [options]`
**Current Mapping:**
- `current` (alias: `cur`) → `current` (alias: `c`)

#### `pending, p [options]`
**Current Mapping:**
- `pending` (alias: `pend`) → `pending` (alias: `p`)

#### `failing, f [options]`
**Current Mapping:**
- `failing` (alias: `fail`) → `failing` (alias: `f`)

### NEW INSPECTION COMMAND

#### `show <type> [id] [options]`
**New Command:** Unified inspection for specific entities

**Behavior:**
```bash
# Show epic details
agentpm show epic [--file <epic-file>]

# Show specific phase details
agentpm show phase <phase-id>

# Show specific task details  
agentpm show task <task-id>

# Show specific test details
agentpm show test <test-id>

# Show event details (if events have IDs)
agentpm show event <event-id>
```

### PROJECT COMMANDS (unchanged)

#### `init`, `switch <epic>`, `config`, `validate`
**No Changes:** These commands work well as-is

### REPORTING COMMANDS (minimal changes)

#### `events`, `docs`, `handoff`
**Current Mapping:**
- `events` (alias: `evt`) → `events` (no alias change needed)
- `docs` → `docs` (unchanged)
- `handoff` → `handoff` (unchanged)

**Note:** `log <message>` command is removed - event logging should be automatic

### SYSTEM COMMANDS

#### `version, v`, `help, h`
**Current Mapping:**
- `version` (aliases: `ver`, `v`) → `version` (alias: `v`)
- Built-in help → `help` (alias: `h`)

## Technical Requirements

### Command Structure Changes

#### 1. Unified Command Implementation
**Current:** Each command is a separate function (e.g., `StartEpicCommand()`, `StartPhaseCommand()`)
**New:** Consolidated commands with sub-command routing:

```go
// cmd/start.go - New unified start command
func StartCommand() *cli.Command {
    return &cli.Command{
        Name:    "start",
        Usage:   "Start working on something",
        Subcommands: []*cli.Command{
            startEpicSubcommand(),
            startPhaseSubcommand(), 
            startTaskSubcommand(),
            startTestSubcommand(),
        },
    }
}
```

#### 2. Argument Pattern Standardization
**Consistent Patterns:**
- **Type-only operations:** `start epic`, `done epic` (no ID needed)
- **Type+ID operations:** `start phase <id>`, `done task <id>` (ID required)
- **ID-only operations:** `pass <test-id>`, `fail <test-id>` (type implied)

#### 3. Global Flag Consolidation
**Current Flags (inconsistent across commands):**
```go
// Some commands have --file/-f, others don't
// Some have --format/-F, others have different defaults
// Time flags sometimes have aliases, sometimes don't
```

**New Standardized Global Flags:**
```go
// All commands inherit these global flags
&cli.StringFlag{Name: "file", Aliases: []string{"f"}, Usage: "Override epic file from config"}
&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Usage: "Override config file path", Value: "./.agentpm.json"}
&cli.StringFlag{Name: "time", Aliases: []string{"t"}, Usage: "Timestamp for current time (testing support)"}
&cli.StringFlag{Name: "format", Aliases: []string{"F"}, Usage: "Output format - text (default) / json / xml", Value: "text"}
```

### Implementation Architecture

#### Phase 1: Command Consolidation
1. **Create unified command files:**
   - `cmd/start.go` - Consolidates start-epic, start-phase, start-task, start-test
   - `cmd/done.go` - Consolidates done-epic, done-phase, done-task  
   - `cmd/cancel.go` - Consolidates cancel-task, cancel-test
   - `cmd/show.go` - New inspection command

2. **Remove old command files:**
   - Delete all old individual command files (start-epic.go, start-phase.go, etc.)
   - Extract reusable service layer logic into internal packages
   - Update main.go command registration

#### Phase 2: Alias Standardization
1. **Update status command aliases:**
   - `status` alias: `st` → `s`
   - `current` alias: `cur` → `c` 
   - `pending` alias: `pend` → `p`
   - `failing` alias: `fail` → `f`

2. **Standardize all command aliases:**
   - Remove inconsistent aliases
   - Apply uniform alias patterns

#### Phase 3: Argument Pattern Enforcement
1. **Standardize ID argument handling:**
   - All type+ID commands validate ID format
   - Consistent error messages for missing/invalid IDs
   - Help text shows clear argument patterns

2. **Unify flag inheritance:**
   - Global flags available on all relevant commands
   - Consistent flag descriptions and defaults

## Testing Strategy

### Command Functionality Testing
1. **New Command Tests:**
   - Verify new unified commands produce expected outputs
   - Test all argument patterns and flag combinations
   - Validate error handling for invalid inputs
   - Test subcommand routing logic

2. **Integration Tests:**
   - End-to-end workflows using new command structure
   - Configuration file compatibility
   - Service layer integration verification

### User Experience Testing  
1. **Command Discovery:**
   - Help system shows logical command groupings
   - Tab completion works with new patterns
   - Error messages guide users to correct syntax

2. **Workflow Efficiency:**
   - Common workflows are faster with new syntax
   - Aliases reduce typing for frequent operations
   - Consistent patterns reduce mental overhead

## Migration Guide for Users

### Quick Reference Card
```bash
# OLD SYNTAX → NEW SYNTAX
start-epic              → start epic
start-phase 3A          → start phase 3A  
start-task 3A_1         → start task 3A_1
start-test 3A_T1        → start test 3A_T1
start-next              → next

done-epic               → done epic
done-phase 3A           → done phase 3A
done-task 3A_1          → done task 3A_1

pass-test 3A_T1         → pass 3A_T1
fail-test 3A_T1 "reason" → fail 3A_T1 reason

cancel-task 3A_1        → cancel task 3A_1
cancel-test 3A_T1 "why" → cancel test 3A_T1 why

status (alias: st)      → status (alias: s)
current (alias: cur)    → current (alias: c)
pending (alias: pend)   → pending (alias: p)
failing (alias: fail)   → failing (alias: f)

# NEW COMMANDS
agentpm show epic       → Show epic details
agentpm show phase 3A   → Show specific phase
agentpm show task 3A_1  → Show specific task
agentpm show test 3A_T1 → Show specific test
```

### Workflow Examples
```bash
# Start working on epic → work on phase → complete task
agentpm start epic
agentpm start phase 3A  
agentpm start task 3A_1
agentpm done task 3A_1

# Check status with short aliases
agentpm s               # Overall status
agentpm c               # Current work
agentpm p               # Pending items
agentpm f               # Failing tests

# Test workflow
agentpm start test 3A_T1
agentpm pass 3A_T1      # or: agentpm fail 3A_T1 "reason"

# Inspection
agentpm show task 3A_1  # Detailed task info
agentpm show phase 3A   # Phase overview
```

## Implementation Phases

### Phase 1: Core Command Consolidation (Days 1-2)
- [ ] Create `cmd/start.go` with unified start command and subcommands
- [ ] Create `cmd/done.go` with unified done command and subcommands  
- [ ] Create `cmd/cancel.go` with unified cancel command
- [ ] Extract service layer logic from old commands into reusable packages
- [ ] Delete old command files (start-epic.go, start-phase.go, etc.)
- [ ] Update `main.go` to register only new commands
- [ ] Comprehensive testing of new command paths

### Phase 2: Testing Commands & Aliases (Days 3-4)
- [ ] Create `cmd/pass.go` and `cmd/fail.go` for simplified test commands
- [ ] Implement `cmd/show.go` for unified inspection
- [ ] Update status command aliases (`st`→`s`, `cur`→`c`, `pend`→`p`, `fail`→`f`)
- [ ] Rename `start-next` to `next` command
- [ ] Standardize global flags across all commands
- [ ] Update help text and usage examples

### Phase 3: Service Layer & Testing (Days 5-6)
- [ ] Refactor extracted service logic for clean interfaces
- [ ] Comprehensive test suite for new command structure
- [ ] Integration tests for complete workflows
- [ ] Update all documentation and examples  
- [ ] Performance testing to ensure no regressions
- [ ] User experience testing with new command patterns

### Phase 4: Polish & Finalization (Day 7)
- [ ] Error message standardization across all commands
- [ ] Tab completion configuration for new structure
- [ ] Final integration testing and bug fixes
- [ ] Documentation updates and examples
- [ ] Release preparation and validation

## Success Criteria

### Functional Requirements
- [ ] All existing functionality accessible through new command structure
- [ ] New commands produce identical outputs to old functionality
- [ ] All global flags work consistently across commands
- [ ] Error handling maintains current quality and helpfulness

### User Experience Goals
- [ ] 25% reduction in total command count (20 → 15)
- [ ] Consistent short aliases for all frequent commands
- [ ] Logical command grouping reduces learning curve
- [ ] Common workflows require fewer keystrokes
- [ ] Help system clearly shows new command patterns

### Technical Goals
- [ ] Code complexity reduced through command consolidation
- [ ] Shared logic extracted into reusable components
- [ ] Flag handling standardized across all commands
- [ ] Test coverage maintained at >90% for all new commands
- [ ] No performance regressions in command execution time

## Risk Assessment

### Low Risks
- **Existing functionality:** Well-tested service layer remains unchanged
- **Flag compatibility:** Current flag patterns are preserved
- **Output formats:** JSON/XML outputs remain identical

### Medium Risks  
- **User disruption:** Breaking change will require users to update workflows immediately
  - *Mitigation:* Clear migration guide and comprehensive documentation
- **Muscle memory:** Users will need to relearn command patterns
  - *Mitigation:* Intuitive new patterns and helpful error messages

### High Risks
- **Complex command routing:** Subcommand implementation may introduce bugs
  - *Mitigation:* Extensive testing and phased rollout
- **Edge case handling:** New argument patterns may miss existing edge cases
  - *Mitigation:* Comprehensive regression testing against current test suite

## Future Considerations

### Post-Implementation Opportunities
1. **Interactive Mode:** `agentpm start` could prompt for epic/phase/task selection
2. **Tab Completion:** Enhanced completion for new command structure
3. **Command Chaining:** `agentpm start phase 3A && agentpm start task 3A_1`
4. **Workflow Templates:** `agentpm workflow start-phase-work 3A`

### CLI Evolution Path
1. **v2.0:** Complete CLI restructure with new command interface
2. **v2.1+:** Add advanced features like interactive mode and chaining

This specification provides a comprehensive roadmap for streamlining the AgentPM CLI as a breaking change that delivers immediate benefits through a cleaner, more intuitive interface.