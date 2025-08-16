# EPIC-1 SPECIFICATION: Foundation & Configuration

## Overview

**Epic ID:** 1  
**Name:** Foundation & Configuration  
**Duration:** 3-4 days  
**Status:** pending  
**Priority:** high  

**Goal:** Establish core CLI structure, configuration management, and XML handling foundation for the AgentPM CLI tool.

## Business Context

This epic establishes the foundational architecture for AgentPM, a CLI tool designed specifically for LLM agents to track and manage epic-based development work. The tool must be simple, non-interactive, and provide structured XML output for agent consumption.

## User Stories

### Primary User Stories
- **As an agent, I can initialize a new project with an epic file** so that I can start tracking development work in a structured way
- **As an agent, I can configure the default epic file for my current project** so that subsequent commands operate on the correct epic
- **As an agent, I can validate epic XML structure for correctness** so that I can ensure data integrity before proceeding with work
- **As an agent, I can view current project configuration** so that I can understand the current project context

### Secondary User Stories
- **As an agent, I can get comprehensive help with examples** so that I can discover available commands and their usage
- **As an agent, I can handle missing files gracefully** so that I receive clear error messages when configuration is invalid

## Technical Requirements

### Core Dependencies
- **CLI Framework:** `github.com/urfave/cli/v3` for command-line interface
- **XML Processing:** `github.com/beevik/etree` for XML parsing, writing, and XPath-like queries
- **Testing:** `github.com/stretchr/testify` for test assertions
- **Snapshot Testing:** `github.com/gkampitakis/go-snaps` for deterministic output testing

### Architecture Principles
- **Simplicity First:** No databases, no multiuser support, no concurrency handling
- **XML-Centric:** All data storage and exchange uses XML format
- **Dependency Injection:** File operations abstracted for testability
- **Non-Interactive:** All commands suitable for agent automation
- **Deterministic Testing:** Support time injection for consistent snapshots

### Configuration Management
- **Config File:** `.agentpm.json` in project root
- **Config Schema:**
  ```json
  {
    "current_epic": "epic-8.xml",
    "project_name": "MooCRM", 
    "default_assignee": "agent_claude"
  }
  ```

### Global CLI Flags
- `--file, -f`: Override epic file from config
- `--config, -c`: Override config file path (default: `./.agentpm.json`)
- `--time, -t`: Timestamp for current time (testing support)
- `--format, -F`: Output format - text (default) / json / xml
- `--help`: Display command help

## Functional Requirements

### FR-1: Project Initialization
**Command:** `agentpm init --epic <epic-file>`

**Behavior:**
- Creates `.agentpm.json` configuration file if it doesn't exist
- Sets `current_epic` to specified epic file
- Validates that epic file exists and is readable
- Overwrites existing configuration if present

**Output Format:**
```xml
<init_result>
    <project_created>true</project_created>
    <config_file>.agentpm.json</config_file>
    <current_epic>epic-8.xml</current_epic>
</init_result>
```

### FR-2: Configuration Display  
**Command:** `agentpm config`

**Behavior:**
- Reads current `.agentpm.json` configuration
- Displays all configuration values
- Shows warning if epic file is missing
- Handles missing config file gracefully

**Output Format:**
```xml
<config>
    <current_epic>epic-8.xml</current_epic>
    <project_name>MooCRM</project_name>
    <default_assignee>agent_claude</default_assignee>
</config>
```

### FR-3: Epic Validation
**Command:** `agentpm validate [--file <epic-file>]`

**Behavior:**
- Validates XML structure and required elements
- Checks epic status values against enum constraints
- Validates phase/task/test relationships
- Reports specific validation errors with context

**Output Format:**
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
        <check name="test_coverage">warning</check>
    </checks_performed>
    <message>Epic structure is valid with 1 warning</message>
</validation_result>
```

### FR-4: Help System
**Command:** `agentpm help [command]`

**Behavior:**
- Provides comprehensive command documentation
- Includes practical examples for agent discovery
- Shows available global flags and their usage
- Contextual help for specific commands

## Non-Functional Requirements

### NFR-1: Performance
- Commands execute in < 100ms for typical epic files
- XML parsing handles files up to 10MB efficiently
- Help system responds instantly

### NFR-2: Reliability
- Graceful error handling with clear messages
- Atomic file operations to prevent corruption
- Validation prevents invalid state transitions

### NFR-3: Usability (for Agents)
- All commands non-interactive
- Structured XML output for programmatic consumption
- Consistent error message format
- Rich help with examples

### NFR-4: Maintainability
- Clean separation between CLI and business logic
- Comprehensive test coverage (>90%)
- Dependency injection for testability
- Clear code organization

## Data Model

### Epic XML Schema (Basic Structure)
```xml
<epic id="8" name="Epic Name" status="planning" created_at="2025-08-16T09:00:00Z">
    <assignee>agent_claude</assignee>
    <description>Epic description</description>
    <phases>
        <!-- Phase definitions -->
    </phases>
    <tasks>
        <!-- Task definitions -->
    </tasks>
    <tests>
        <!-- Test definitions -->
    </tests>
    <events>
        <!-- Event log -->
    </events>
</epic>
```

### Configuration Schema
```json
{
    "current_epic": "string (required)",
    "project_name": "string (optional)",
    "default_assignee": "string (optional, default: 'agent')"
}
```

## Error Handling

### Error Categories
1. **File System Errors:** Missing files, permission issues, disk space
2. **XML Parsing Errors:** Malformed XML, invalid structure
3. **Validation Errors:** Business rule violations, invalid data
4. **Configuration Errors:** Missing config, invalid values

### Error Output Format
```xml
<error>
    <type>validation_error</type>
    <message>Epic ID is required</message>
    <details>
        <field>epic.id</field>
        <constraint>not_empty</constraint>
    </details>
</error>
```

## Acceptance Criteria

### AC-1: Project Initialization
- **GIVEN** I am in an empty directory
- **WHEN** I run `agentpm init --epic epic-8.xml`
- **THEN** a `.agentpm.json` file should be created with `current_epic: "epic-8.xml"`

### AC-2: Configuration Display
- **GIVEN** I have a valid `.agentpm.json` file
- **WHEN** I run `agentpm config`
- **THEN** I should see current epic file and project configuration

### AC-3: Epic Validation Success
- **GIVEN** I have a well-formed epic XML file
- **WHEN** I run `agentpm validate`
- **THEN** validation should pass with success message

### AC-4: Epic Validation Failure
- **GIVEN** I have an epic XML with missing required attributes
- **WHEN** I run `agentpm validate`
- **THEN** validation should fail with specific error details

### AC-5: File Override Support
- **GIVEN** I have multiple epic files
- **WHEN** I run `agentpm validate -f epic-9.xml`
- **THEN** validation should run on epic-9.xml instead of current epic

### AC-6: Missing File Handling
- **GIVEN** I specify a non-existent epic file
- **WHEN** I run any command with `-f missing-epic.xml`
- **THEN** I should get a clear file not found error

### AC-7: Help System
- **GIVEN** I need command information
- **WHEN** I run `agentpm help` or `agentpm help init`
- **THEN** I should see comprehensive documentation with examples

## Testing Strategy

### Test Categories
- **Unit Tests (80%):** Business logic validation, XML parsing, configuration management
- **Integration Tests (15%):** File I/O operations, command execution
- **End-to-End Tests (5%):** Full CLI workflows

### Test Isolation
- Each test uses `t.TempDir()` for complete filesystem isolation
- Business logic tests use in-memory storage implementation
- No shared state between tests
- Parallel test execution support

### Test Data
- Test factories for consistent epic/config creation
- Golden files for complex validation scenarios
- Snapshot testing for XML output validation

## Implementation Phases

### Phase 1A: Core Structure (Day 1)
- CLI framework setup with urfave/cli
- Basic command structure and routing
- Global flag handling
- Help system foundation

### Phase 1B: Configuration Management (Day 1-2)
- Config file reading/writing
- Configuration validation
- Project initialization command
- Config display command

### Phase 1C: XML Foundation (Day 2-3)
- XML parsing with etree
- Basic epic structure definition
- XML validation framework
- Error handling patterns

### Phase 1D: Integration & Testing (Day 3-4)
- End-to-end command testing
- Error message refinement
- Performance optimization
- Documentation completion

## Definition of Done

- [ ] All acceptance criteria verified with automated tests
- [ ] Commands execute in < 100ms for typical files
- [ ] Test coverage > 90% for business logic
- [ ] All error cases handled gracefully
- [ ] Help system complete with examples
- [ ] No external dependencies beyond specified libraries
- [ ] Code passes lint and format checks
- [ ] Integration tests verify CLI behavior end-to-end

## Dependencies and Risks

### Dependencies
- No blocking dependencies on other epics
- Requires Go 1.21+ runtime environment

### Risks
- **Medium Risk:** XML schema complexity might require iteration
- **Low Risk:** CLI framework learning curve
- **Low Risk:** File system edge cases on different platforms

### Mitigation Strategies
- Start with minimal XML schema and iterate
- Use CLI framework examples and documentation
- Test on multiple platforms early
- Implement comprehensive error handling

## Notes

- This epic establishes patterns that all subsequent epics will follow
- Focus on simplicity over feature completeness
- XML output format must be stable for agent consumption
- Configuration design should support future multi-epic workflows