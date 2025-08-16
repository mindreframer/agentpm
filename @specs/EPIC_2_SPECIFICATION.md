# EPIC-2 SPECIFICATION: Query & Status Commands

## Overview

**Epic ID:** 2  
**Name:** Query & Status Commands  
**Duration:** 3-4 days  
**Status:** pending  
**Priority:** high  
**Depends On:** Epic 1 (Foundation & Configuration)

**Goal:** Implement comprehensive read-only operations to query epic state, progress tracking, and activity monitoring for agent workflow optimization.

## Business Context

Epic 2 builds upon the foundation established in Epic 1 to provide agents with essential visibility into their work progress. These query commands enable agents to understand current state, identify next actions, and track completion progress without modifying any data. This epic focuses on information retrieval and reporting that agents need for effective decision-making.

## User Stories

### Primary User Stories
- **As an agent, I can view overall epic status and progress** so that I understand how much work is completed and what remains
- **As an agent, I can see what I'm currently working on** so that I know my active tasks and next actions
- **As an agent, I can list all pending tasks** so that I can plan my upcoming work effectively
- **As an agent, I can identify failing tests that need attention** so that I can prioritize bug fixes and quality issues
- **As an agent, I can review recent activity and events** so that I can understand the development timeline and context

### Secondary User Stories
- **As an agent, I can query different epic files** so that I can work with multiple projects using the -f flag
- **As an agent, I can get structured XML output** so that I can programmatically process status information
- **As an agent, I can handle missing or corrupted data gracefully** so that I receive clear error messages when epic state is invalid

## Technical Requirements

### Core Dependencies
- **Foundation:** Epic 1 CLI framework, XML processing, and storage interface
- **XML Querying:** Enhanced etree usage for efficient data filtering and extraction
- **Performance:** Optimized parsing to avoid loading entire DOM for simple queries
- **Output Format:** Consistent XML structure for all status responses

### Architecture Principles
- **Read-Only Operations:** No data modification, only querying and reporting
- **Simple Parsing:** Load entire epic XML for straightforward processing
- **Structured Output:** All responses in XML format for agent consumption
- **File Override Support:** All commands support `-f` flag for multi-epic workflows
- **Error Resilience:** Graceful handling of incomplete or invalid epic data

### Query Optimization
- **Simple Parsing:** Load entire epic XML into memory for straightforward processing
- **Caching Strategy:** In-memory caching for repeated queries within single command execution

## Functional Requirements

### FR-1: Epic Status Overview
**Command:** `agentpm status [--file <epic-file>]`

**Behavior:**
- Displays comprehensive epic status with progress metrics
- Calculates completion percentage based on task and test completion
- Shows current active phase and task if any
- Identifies failing tests count for quick assessment
- Handles missing or incomplete epic data gracefully

**Output Format:**
```xml
<status epic="8">
    <name>Schools Index Pagination</name>
    <status>in_progress</status>
    <progress>
        <completed_phases>2</completed_phases>
        <total_phases>4</total_phases>
        <passing_tests>12</passing_tests>
        <failing_tests>1</failing_tests>
        <completion_percentage>50</completion_percentage>
    </progress>
    <current_phase>2A</current_phase>
    <current_task>2A_1</current_task>
</status>
```

### FR-2: Current Work State
**Command:** `agentpm current [--file <epic-file>]`

**Behavior:**
- Shows currently active phase and task
- Provides next action recommendations based on failing tests or pending work
- Displays epic-level status for context
- Identifies blockers or failing tests that need immediate attention
- Returns empty active elements when no work is in progress

**Output Format:**
```xml
<current_state epic="8">
    <epic_status>in_progress</epic_status>
    <active_phase>2A</active_phase>
    <active_task>2A_1</active_task>
    <next_action>Fix mobile responsive pagination controls</next_action>
    <failing_tests>1</failing_tests>
</current_state>
```

### FR-3: Pending Work Overview
**Command:** `agentpm pending [--file <epic-file>]`

**Behavior:**
- Lists all pending phases, tasks, and tests across the entire epic
- Groups items by type (phases, tasks, tests) for clear organization
- Includes task descriptions and phase associations
- Shows work distribution across phases
- Returns empty sections when no pending work exists

**Output Format:**
```xml
<pending_work epic="8">
    <phases>
        <phase id="3A" name="LiveView Integration" status="pending"/>
        <phase id="4A" name="Performance Optimization" status="pending"/>
    </phases>
    <tasks>
        <task id="2A_2" phase_id="2A" status="pending">Add accessibility features to pagination controls</task>
    </tasks>
    <tests>
        <test id="2A_3" phase_id="2A" status="pending">URL state persistence test</test>
    </tests>
</pending_work>
```

### FR-4: Failing Tests Report
**Command:** `agentpm failing [--file <epic-file>]`

**Behavior:**
- Shows only tests with "failing" status
- Includes test descriptions and failure details
- Groups by phase for context
- Shows failure notes when available
- Returns empty list when all tests are passing

**Output Format:**
```xml
<failing_tests epic="8">
    <test id="2A_2" phase_id="2A">
        <description>
            **GIVEN** I'm on mobile device  
            **WHEN** I tap pagination controls  
            **THEN** They work and are easy to tap (44px+ targets)
        </description>
        <failure_note>Touch targets too small, need 44px+ minimum</failure_note>
    </test>
</failing_tests>
```

### FR-5: Recent Events Timeline
**Command:** `agentpm events [--limit=N] [--file <epic-file>]`

**Behavior:**
- Shows recent events in reverse chronological order (most recent first)
- Supports configurable limit (default: 10, max: 100)
- Includes event metadata: timestamp, agent, phase, type
- Shows event content with proper formatting
- Handles empty event history gracefully

**Output Format:**
```xml
<events epic="8" limit="5">
    <event timestamp="2025-08-16T15:00:00Z" agent="agent_claude" phase_id="2A" type="blocker">
        Found design system dependency
        
        Blocker: Need design system tokens for mobile responsive design
    </event>
    <event timestamp="2025-08-16T14:45:00Z" agent="agent_claude" phase_id="2A" type="test_failed">
        Mobile responsive test failing
        
        Test: 2A_2 - Mobile pagination controls
        Issue: Touch targets too small, need 44px+ minimum
    </event>
    <event timestamp="2025-08-16T14:30:00Z" agent="agent_claude" phase_id="2A" type="implementation">
        Implemented basic pagination controls
        
        Files: src/components/Pagination.js (added), src/styles/pagination.css (added)
        Result: Basic controls working, all tests passing
    </event>
</events>
```

## Non-Functional Requirements

### NFR-1: Performance
- Commands execute in < 100ms for typical epic files
- Simple XML parsing with etree - load entire document
- Memory usage reasonable for expected epic file sizes (< 1MB)

### NFR-2: Scalability
- Handle typical epics with reasonable numbers of phases, tasks, and tests
- Event queries work well with normal event history
- Simple approach - no pagination or streaming needed

### NFR-3: Reliability
- Graceful degradation when XML data is incomplete
- Clear error messages for corrupted or invalid epic files
- Consistent output format even with partial data
- Safe handling of missing timestamps or malformed dates

### NFR-4: Usability (for Agents)
- Consistent XML schema across all query commands
- Predictable error handling and response format
- Clear indication when no data is available
- Support for both structured queries and quick status checks

## Data Processing Logic

### Progress Calculation
```
completion_percentage = (completed_tasks + completed_tests) / (total_tasks + total_tests) * 100
```

### Phase Status Determination
- **completed:** All tasks in phase have status "completed"
- **in_progress:** At least one task in phase has status "in_progress"  
- **pending:** All tasks in phase have status "pending"

### Next Action Logic
1. If failing tests exist → "Fix failing tests: [test descriptions]"
2. If active task exists → "Continue work on: [task description]"
3. If pending tasks in active phase → "Start next task: [task description]"
4. If pending phases → "Start next phase: [phase name]"
5. If all complete → "Epic ready for completion"

### Event Filtering
- Default: All event types in chronological order
- Future: Support for type filtering (--type=blocker,test_failed)
- Future: Date range filtering (--since=2025-08-15)

## Error Handling

### Error Categories
1. **File Access Errors:** Missing files, permission issues
2. **XML Parsing Errors:** Malformed XML, invalid structure
3. **Data Consistency Errors:** Missing references, invalid states
4. **Query Parameter Errors:** Invalid limits, malformed flags

### Error Output Format
```xml
<error>
    <type>file_not_found</type>
    <message>Epic file not found: epic-missing.xml</message>
    <details>
        <file>epic-missing.xml</file>
        <suggestion>Check file path or use 'agentpm config' to see current epic</suggestion>
    </details>
</error>
```

### Graceful Degradation
- Show available data when some sections are missing
- Calculate progress with available task/test data
- Provide warnings for incomplete data sets
- Default values for missing timestamps or metadata

## Acceptance Criteria

### AC-1: Epic Status Display
- **GIVEN** I have an epic with 2 completed phases and 4 total phases
- **WHEN** I run `agentpm status`
- **THEN** I should see status as "in_progress" with 50% completion based on task distribution

### AC-2: Current Work Identification
- **GIVEN** I have an epic with active phase "2A" and active task "2A_1"
- **WHEN** I run `agentpm current`
- **THEN** I should see active_phase: "2A" and active_task: "2A_1" with next action guidance

### AC-3: Pending Work Listing
- **GIVEN** I have tasks in "pending" status across multiple phases
- **WHEN** I run `agentpm pending`
- **THEN** I should see all pending tasks grouped by type with phase associations

### AC-4: Failing Tests Identification
- **GIVEN** I have tests with mixed pass/fail status
- **WHEN** I run `agentpm failing`
- **THEN** I should only see tests with status "failing" and their failure details

### AC-5: Recent Events Timeline
- **GIVEN** I have 10 events in my epic history
- **WHEN** I run `agentpm events --limit=3`
- **THEN** I should see only the 3 most recent events in reverse chronological order

### AC-6: File Override Support
- **GIVEN** I have multiple epic files
- **WHEN** I run `agentpm status -f epic-9.xml`
- **THEN** status should be shown for epic-9.xml instead of current configured epic

### AC-7: Empty State Handling
- **GIVEN** I have an epic with no failing tests
- **WHEN** I run `agentpm failing`
- **THEN** I should see an empty failing tests list with appropriate message

### AC-8: Performance Requirements
- **GIVEN** I have a typical epic file
- **WHEN** I run any query command
- **THEN** the command should execute in < 100ms

## Testing Strategy

### Test Categories
- **Unit Tests (70%):** Query logic, data processing, progress calculation
- **Integration Tests (25%):** XML parsing, file operations, command execution
- **Performance Tests (5%):** Large file handling, memory usage, execution time

### Test Data Requirements
- Small epic files for basic functionality
- Large epic files for performance testing
- Invalid/incomplete epic files for error handling
- Mixed status scenarios for progress calculation

### Test Isolation
- Each test uses isolated epic files in `t.TempDir()`
- Mock storage implementation for controlled test data
- Snapshot testing for XML output validation
- Performance benchmarks for execution time verification

## Implementation Phases

### Phase 2A: Basic Query Infrastructure (Day 1)
- Query service foundation with Storage interface integration
- Basic XML parsing and data extraction utilities
- Progress calculation algorithms
- Error handling framework for query operations

### Phase 2B: Status and Current Commands (Day 1-2)
- `agentpm status` command implementation
- `agentpm current` command implementation
- Phase/task status determination logic
- Next action recommendation engine

### Phase 2C: Pending and Failing Commands (Day 2-3)
- `agentpm pending` command implementation
- `agentpm failing` command implementation
- Data filtering and grouping utilities
- Test status analysis

### Phase 2D: Events and Performance Optimization (Day 3-4)
- `agentpm events` command implementation
- Event timeline processing with limits
- Query performance optimization
- Large file handling improvements

## Definition of Done

- [ ] All acceptance criteria verified with automated tests
- [ ] Commands execute in < 100ms for typical epic files
- [ ] Test coverage > 85% for query logic
- [ ] All error cases handled gracefully with clear messages
- [ ] XML output format consistent across all commands
- [ ] File override (-f flag) works for all query commands
- [ ] Simple XML parsing approach implemented
- [ ] Integration tests verify end-to-end query workflows

## Dependencies and Risks

### Dependencies
- **Epic 1:** CLI framework, XML processing, storage interface, configuration management
- **Test Data:** Representative epic files with various states and progress levels

### Risks
- **Low Risk:** XML parsing complexity for malformed epic data
- **Low Risk:** Progress calculation edge cases with missing data

### Mitigation Strategies
- Simple XML parsing with etree - no complex optimizations needed
- Comprehensive error handling and graceful degradation
- Extensive testing with edge cases and malformed data

## Future Considerations

### Potential Enhancements (Not in Scope)
- Interactive query mode with filtering options
- Export functionality for query results
- Query result caching for repeated operations
- Advanced filtering by date ranges, event types, or assignees
- Graphical progress visualization

### Integration Points
- Epic 3: Epic lifecycle commands will use these queries for validation
- Epic 4: Task management will leverage pending/current queries
- Epic 5: Event logging will integrate with events query functionality