# AgentPM CLI Tool - Development Roadmap

<IMPORTANT>
- prefer simplicity!
- NO DATABASES, NO MULTIUSER, NO concurrency handling!
- DATA WILL ALWAYS BE XML!
- PREFER PURE TESTS, with SNAPSHOT testing for complex OUTPUT
- DATETIME can be injected for deterministic SNAPSHOTS and assertions!
- FOR CLI OUTPUT WE WILL SUPPORT A GENERIC SERIALIZER FOR XML (default) / JSON / human ouput
    - internal tests can assert on the generic golang structures via snapshot testing
- KEEP THING lightweight, no need for complex features
- as fallback we can provide a XPath query CLI interface, that the user can use. 
- GLOBAL CLI FLAGS
    -f : override epic file from config (default is taken from ./.agentpm.json)
    -config : override config file (default is ./.agentpm.json)
    -t : timestamp for current time, useful for testing!
    -format : text (default) / json / xml (output from the CLI)
</IMPORTANT>

## Epic 1: Foundation & Configuration
**Duration:** 3-4 days  
**Goal:** Core CLI structure, configuration management, and XML handling

### User Stories:
- As an agent, I can initialize a new project with an epic file
- As an agent, I can configure the default epic file for my current project
- As an agent, I can validate epic XML structure for correctness
- As an agent, I can view current project configuration

### Technical Requirements:
- Use `github.com/urfave/cli/v3` for CLI framework
- Use `github.com/beevik/etree` for XML parsing/writing / querying (supports XPath like queries!)
- Use `https://github.com/stretchr/testify` to implement tests
- Use `https://github.com/gkampitakis/go-snaps` for snapshot based testing

- Implement dependency injection pattern for file operations
- Support `.agentpm.json` configuration file format
- Provide comprehensive help with examples for agent discovery

### Acceptance Criteria:
- `agentpm init --epic epic-8.xml` creates `.agentpm.json` with current_epic
- `agentpm config` shows current project configuration
- `agentpm validate` checks epic XML structure and reports errors
- All commands have rich help with usage examples
- CLI handles missing files gracefully with clear error messages

### Testing Strategy:
- Each test uses `t.TempDir()` for complete isolation
- Business logic separated from CLI commands for unit testing
- Integration tests verify CLI behavior end-to-end
- Test factories provide consistent epic/config setup
- Memory storage implementation for fast unit tests

---

## Epic 2: Query & Status Commands
**Duration:** 3-4 days  
**Goal:** Read-only operations to query epic state and progress

### User Stories:
- As an agent, I can view overall epic status and progress
- As an agent, I can see what I'm currently working on
- As an agent, I can list all pending tasks
- As an agent, I can identify failing tests that need attention
- As an agent, I can review recent activity and events

### Technical Requirements:
- Implement XML query and filtering logic
- Structured XML output for all status commands
- Support file override with `-f` flag for multi-epic workflows
- Efficient parsing without loading entire DOM for simple queries

### Acceptance Criteria:
- `agentpm status` shows epic progress with completion percentage
- `agentpm current` displays active phase/task and next actions
- `agentpm pending` lists all pending tasks across phases
- `agentpm failing` shows only failing tests with failure details
- `agentpm events --limit=5` shows recent activity chronologically
- All commands support `-f epic-9.xml` to override current epic

### Testing Strategy:
- Table-driven tests for different epic states
- Test data factories for various progress scenarios
- Isolated XML parsing tests with sample files
- Command output validation with golden file patterns

---

## Epic 3: Epic Lifecycle Management
**Duration:** 3-4 days  
**Goal:** Manage epic creation, status transitions, and project switching

### User Stories:
- As an agent, I can switch between different epic files
- As an agent, I can start working on a new epic
- As an agent, I can complete an epic when all work is done
- As an agent, I can not abandon / pause / cancel anything, since I'm an LLM

### Technical Requirements:
- Epic status lifecycle: pending → wip → done
- MAKE THIS AS SIMPLE AS POSSIBLE, not need for overengineering
- Timestamp tracking for all state transitions
- Validation rules for valid status transitions
- Automatic event logging for lifecycle changes

### Acceptance Criteria:
- `agentpm switch epic-9.xml` updates current_epic in config
- `agentpm start-epic` changes status from pending to wip
- `agentpm done-epic` marks epic as done with validation
- EPICs can not be paused / cancelled / resumed, etc. 
- Status transitions are validated
- All lifecycle changes create timestamped events

### Testing Strategy:
- State machine tests for all valid/invalid transitions
- CLI MUST accept an optional timestamp for the current time, so we can test time progression
    + deterministic snapshots! (-t flag)
- Timestamp validation in isolated timezone
- Error cases for invalid state changes
- Integration tests for config file updates

---

## Epic 4: Task & Phase Management 
**Duration:** 4-5 days  
**Goal:** Granular work tracking at phase and task levels

### User Stories:
- As an agent, I can start working on a specific phase
- As an agent, I can begin individual tasks within phases
- As an agent, I can automatically pick the next pending task
- As an agent, I can mark tasks and phases as completed
- As an agent, I can track progress through complex work plans

### Technical Requirements:
- This epic and epic 5 are the most important epics of the roadmap!
- Phase and task status tracking (pending → wip → done)
- Auto-next logic to select next pending task intelligently
- Dependency validation (can't start phase 2 if phase 1 incomplete)
- Current state tracking for active phase/task

### Acceptance Criteria:
- starting new phase/task IS NOT possible, if there is another active phase/task!
- completing a phase without completing all its tasks is impossible!
- all EVENTS are logged
- `agentpm start-phase 2A` begins phase work with validation (pending -> wip)
- `agentpm start-task 2A_1` starts specific task within active phase (pending -> wip)
- `agentpm cancel-task 2A_1` starts specific task within active phase (wip -> cancelled)
- `agentpm done-task 2A_1` marks task complete and updates progress (wip -> done)
- `agentpm done-phase 2A` completes phase when all tasks done (wip -> done)

- for NON-ERROR responses the output from CLI is minimal, a simple confirmation (NOT XML!), like:
    - Phase 2A started. 
    - Phase 2A completed. 
    - Task 2A_1 started. 
    - Task 2A_1 completed. 

- `agentpm start-next` intelligently selects next pending task. It responds with XML for the started entity. 
    - IF phase: 
        - a list of all tasks with description
        + STARTED TASK ID
    - IF TASK
        - started task ID + description
    - Auto-next prefers tasks in current phase. Only when a phase is complete (with all tasks), it goes to the next


### Testing Strategy:
- Complex task dependency scenarios with multiple phases
- CLI MUST accept an optional timestamp for the current time, so we can test time progression
    + deterministic snapshots!
- Auto-next selection algorithm validation
- Progress calculation edge cases (empty phases, all complete)
- Concurrent phase/task state validation

---

## Epic 5: Test Management & Event Logging
**Duration:** 3-4 days  
**Goal:** Test status tracking and comprehensive event logging

### User Stories:
- As an agent, I can start working on a test
- As an agent, I can mark tests as passing or failing with details
- As an agent, I can log important events during development
- As an agent, I can track file changes and their impact
- As an agent, I can identify blockers and issues quickly
- As an agent, I can maintain a detailed activity timeline

### Technical Requirements:
- Test status management (pending → wip -> passed/failed/cancelled)
- All test transitions are stored as events
- Rich event logging with types (implementation, test_failed, blocker)
- File change tracking with optional metadata
- Event querying and filtering by type/timeframe
- Blocker identification from failed tests and logged issues

### Acceptance Criteria:
- `agentpm start-test 2A_1` marks the test as WIP (the agent has started working on a test)
- `agentpm pass-test 2A_1` marks test as passed
- `agentpm fail-test 2A_2 "Mobile responsive issue"` logs failure details
- `agentpm cancel-test 2A_2 "Spec contradicts itself with point xyz"` marks a test as cancelled with a comment
- `agentpm log "Implemented pagination" --files="src/Pagination.js:added"`
- `agentpm log "Need design tokens" --type=blocker` identifies blockers
- Events include timestamps, types, and optional metadata
- Failed tests and blocker events are easily queryable

### Testing Strategy:
- CLI MUST accept an optional timestamp for the current time, so we can test time progression
    + deterministic snapshots!
- Event logging with various metadata combinations
- Test status transition validation
- File tracking format parsing and validation
- Event filtering and querying logic

---

## Epic 6: Handoff & Documentation
**Duration:** 2-3 days  
**Goal:** Agent handoff support and human-readable documentation

### User Stories:
- As an outgoing agent, I can generate comprehensive handoff reports
- As an incoming agent, I can quickly understand current state and blockers
- As a human, I can generate readable documentation from epic data
- As an agent, I can identify recent activity and context quickly

### Technical Requirements:
- Comprehensive handoff XML with all relevant context
- Markdown documentation generation from XML data
- Recent events summarization with configurable limits
- Blocker extraction from failed tests and logged issues
- Human-readable formatting with proper structure

### Acceptance Criteria:
- `agentpm handoff` generates complete XML with current state, progress, recent events, and blockers
- `agentpm docs` creates markdown documentation suitable for humans
- Handoff includes active work, failing tests, and next actions
- Documentation shows epic overview, phase progress, and timeline
- Recent events are summarized with most important items first

### Testing Strategy:
- Handoff report completeness validation
- Markdown generation formatting tests
- Edge cases (empty epic, no recent activity)
- Output format stability for reliable parsing

---

## Development Guidelines

### Testing Requirements (All Epics):
- **Isolation**: Each test uses `t.TempDir()` for complete filesystem isolation
- **Speed**: Business logic tests use in-memory storage, only integration tests touch filesystem
- **Coverage**: Table-driven tests for multiple scenarios, edge cases included
- **Structure**: Tests organized by epic phase, mirroring implementation structure
- **Factories**: Consistent test data creation with `NewTestEpic()` helpers
- **Integration**: End-to-end CLI tests for critical user workflows

### Code Organization:
```
├── cmd/                 # CLI commands (minimal logic)
├── internal/
│   ├── epic/           # Core business logic
│   ├── config/         # Configuration management  
│   └── storage/        # File operations abstraction
├── testdata/           # Sample XML files for testing
└── pkg/                # Public interfaces if needed
```

### Success Criteria:
- ✅ Zero external dependencies beyond urfave/cli and etree
- ✅ All commands non-interactive for agent automation
- ✅ Comprehensive help system for agent discovery
- ✅ Fast test suite (< 1 second for unit tests)
- ✅ Simple codebase prioritizing clarity over performance
- ✅ Robust error handling with clear messages
