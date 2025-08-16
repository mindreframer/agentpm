# AgentPM CLI Tool - Development Roadmap

<IMPORTANT>
- prefer simplicity!
- NO DATABASES, NO MULTIUSER, NO concurrency handling!
- DATA WILL ALWAYS BE XML!
- PREFER PURE TESTS, with SNAPSHOT testing for complex OUTPUT
- FOR CLI OUTPUT WE WILL SUPPORT A GENERIC SERIALIZER FOR XML (default) / JSON / human ouput (YAML like, but not really)
    - internal tests can assert on the generic golang structures via snapshot testing
- KEEP THING lightweight, no need for complex features
- as fallback we can provide a XPath query CLI interface, that the user can use. 
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
- As an agent, I can start working on a new epic
- As an agent, I can pause work when blocked or interrupted
- As an agent, I can resume paused work and continue progress
- As an agent, I can switch between different epic files
- As an agent, I can complete an epic when all work is done

### Technical Requirements:
- Epic status lifecycle: planning → in_progress → paused/completed/cancelled
- Timestamp tracking for all state transitions
- Validation rules for valid status transitions
- Automatic event logging for lifecycle changes

### Acceptance Criteria:
- `agentpm start-epic` changes status from planning to in_progress
- `agentpm pause-epic "reason"` pauses with optional reason logging
- `agentpm resume-epic` resumes from paused state
- `agentpm switch epic-9.xml` updates current_epic in config
- `agentpm complete-epic` marks epic as completed with validation
- Status transitions are validated (can't resume non-paused epic)
- All lifecycle changes create timestamped events

### Testing Strategy:
- State machine tests for all valid/invalid transitions
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
- Phase and task status tracking (pending → in_progress → completed)
- Auto-next logic to select next pending task intelligently
- Progress calculation based on completed vs total items
- Dependency validation (can't start phase 2 if phase 1 incomplete)
- Current state tracking for active phase/task

### Acceptance Criteria:
- `agentpm start-phase 2A` begins phase work with validation
- `agentpm start-task 2A_1` starts specific task within active phase
- `agentpm start-next` intelligently selects next pending task
- `agentpm complete-task 2A_1` marks task complete and updates progress
- `agentpm complete-phase 2A` completes phase when all tasks done
- Auto-next prefers tasks in current phase, then next phase
- Progress percentage updates automatically

### Testing Strategy:
- Complex task dependency scenarios with multiple phases
- Auto-next selection algorithm validation
- Progress calculation edge cases (empty phases, all complete)
- Concurrent phase/task state validation

---

## Epic 5: Test Management & Event Logging
**Duration:** 3-4 days  
**Goal:** Test status tracking and comprehensive event logging

### User Stories:
- As an agent, I can mark tests as passing or failing with details
- As an agent, I can log important events during development
- As an agent, I can track file changes and their impact
- As an agent, I can identify blockers and issues quickly
- As an agent, I can maintain a detailed activity timeline

### Technical Requirements:
- Test status management (pending → passed/failed)
- Rich event logging with types (implementation, test_failed, blocker)
- File change tracking with optional metadata
- Event querying and filtering by type/timeframe
- Blocker identification from failed tests and logged issues

### Acceptance Criteria:
- `agentpm pass-test 2A_1` marks test as passed
- `agentpm fail-test 2A_2 "Mobile responsive issue"` logs failure details
- `agentpm log "Implemented pagination" --files="src/Pagination.js:added"`
- `agentpm log "Need design tokens" --type=blocker` identifies blockers
- Events include timestamps, types, and optional metadata
- Failed tests and blocker events are easily queryable

### Testing Strategy:
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
