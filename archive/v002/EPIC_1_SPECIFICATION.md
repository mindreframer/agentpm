# Epic 1: Foundation & Configuration - Specification

## Overview
**Goal:** Core CLI structure, configuration management, and XML handling  
**Duration:** 3-4 days  
**Philosophy:** Simple, testable foundation with comprehensive error handling

## User Stories
1. Initialize new project with epic file
2. View and manage project configuration  
3. Validate epic XML structure and correctness
4. Access comprehensive help for agent discovery

## Technical Requirements
- **CLI Framework:** `github.com/urfave/cli/v3`
- **XML Processing:** `github.com/beevik/etree` with XPath support
- **Testing:** `github.com/stretchr/testify` + `github.com/gkampitakis/go-snaps`
- **Architecture:** Dependency injection for file operations
- **Configuration:** `.agentpm.json` format
- **Output:** XML format for all structured responses

## Core Data Structures

### Configuration (.agentpm.json)
```json
{
    "current_epic": "epic-8.xml",
    "project_name": "AgentPM Project", 
    "default_assignee": "agent_claude",
    "created_at": "2025-08-16T15:30:00Z"
}
```

### Epic Structure (Foundation)
```go
type Epic struct {
    ID        string    `xml:"id,attr"`
    Name      string    `xml:"name,attr"`
    Status    string    `xml:"status,attr"`
    CreatedAt time.Time `xml:"created_at,attr"`
    
    Phases []Phase `xml:"phases>phase"`
    Tasks  []Task  `xml:"tasks>task"`
    Tests  []Test  `xml:"tests>test"`
    Events []Event `xml:"events>event"`
}
```

## Implementation Phases

### Phase 1A: CLI Setup (0.5 days)
- CLI framework with urfave/cli/v3
- Global flags (`-f`, `--help`)
- Command registration pattern
- Basic help system

### Phase 1B: Configuration Management (1 day)
- Config struct and JSON operations
- `agentpm init --epic <file>` command
- `agentpm config` command
- File validation and error handling

### Phase 1C: Epic XML Foundation (1 day)
- Epic data structures with XML tags
- XML parsing using etree
- Epic loading and file operations
- Basic file existence validation

### Phase 1D: Epic Validation (1 day)
- Validation rule engine
- `agentpm validate [-f file]` command
- Structural validation (IDs, references, enums)
- Detailed error reporting with line numbers

### Phase 1E: Storage & Testing (0.5 days)
- Storage interface abstraction
- File and memory implementations
- Dependency injection setup
- Test isolation with `t.TempDir()`

## Acceptance Criteria
- ✅ `agentpm init --epic epic-8.xml` creates `.agentpm.json` with current_epic
- ✅ `agentpm config` shows current project configuration with warnings for missing files
- ✅ `agentpm validate` checks epic XML structure and reports specific errors
- ✅ All commands have rich help with usage examples
- ✅ CLI handles missing files gracefully with clear error messages

## Validation Rules
1. **XML Structure:** Well-formed XML with proper closing tags
2. **Required Attributes:** Epic must have `id`, `name`, `status`
3. **Status Values:** Must be: `planning`, `in_progress`, `paused`, `completed`, `cancelled`
4. **Unique IDs:** All phase, task, test IDs must be unique
5. **Reference Integrity:** Task phase_id must reference existing phase

## Testing Strategy
- **Unit Tests:** Pure business logic with in-memory storage (80% coverage)
- **Component Tests:** File operations with `t.TempDir()` isolation (15% coverage)  
- **Integration Tests:** End-to-end CLI command execution (5% coverage)
- **Test Factories:** `NewTestConfig()`, `NewTestEpic()` for consistent data
- **Snapshot Testing:** Complex XML output validation

## Error Handling
- **Clear Context:** Always include what operation was attempted
- **Actionable Messages:** Suggest specific next steps
- **Line Numbers:** Include XML line numbers in validation errors
- **Agent-Friendly:** Structured error output for automated processing

## Output Examples

### agentpm init --epic epic-8.xml
```xml
<init_result>
    <project_created>true</project_created>
    <config_file>.agentpm.json</config_file>
    <current_epic>epic-8.xml</current_epic>
</init_result>
```

### agentpm config
```xml
<config>
    <current_epic>epic-8.xml</current_epic>
    <project_name>AgentPM Project</project_name>
    <default_assignee>agent_claude</default_assignee>
</config>
```

### agentpm validate
```xml
<validation_result epic="8">
    <valid>true</valid>
    <warnings>
        <warning>Task 2A_2 has no tests defined</warning>
    </warnings>
    <checks_performed>
        <check name="xml_structure">passed</check>
        <check name="phase_dependencies">passed</check>
        <check name="task_phase_mapping">passed</check>
    </checks_performed>
    <message>Epic structure is valid with 1 warning</message>
</validation_result>
```

## Test Scenarios (Key Examples)
- **Initialize:** Create config with epic file, handle existing config updates
- **Config Display:** Show current configuration, warn on missing epic file
- **Validation:** Pass valid XML, catch malformed XML, report specific structural errors
- **File Override:** Use `-f` flag to specify different epic files
- **Error Cases:** Missing files, invalid JSON, XML parsing errors

## Quality Gates
- [ ] All acceptance criteria implemented and tested
- [ ] 90%+ unit test coverage for business logic
- [ ] All CLI commands complete within 100ms
- [ ] Clear error messages for all failure cases
- [ ] Help system enables agent command discovery

This specification provides a solid foundation while keeping implementation manageable and focused on the core requirements needed for subsequent epics.