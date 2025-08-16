# Epic 2: Query & Status Commands - Specification

## Overview
**Goal:** Read-only operations to query epic state and progress  
**Duration:** 3-4 days  
**Philosophy:** Efficient XML output for agent consumption with intelligent state analysis

## User Stories
1. View overall epic status and progress with completion metrics
2. See current active work and next recommended actions
3. List all pending tasks across phases for planning
4. Identify failing tests that need immediate attention
5. Review recent activity and events for context

## Technical Requirements
- **Dependencies:** Epic 1 foundation (CLI, config, epic loading, validation)
- **XML Processing:** Efficient queries using `etree` XPath capabilities
- **Output Format:** Structured XML for all status commands
- **File Override:** Support `-f` flag for multi-epic workflows
- **Performance:** Fast queries without full DOM loading for simple operations

## Core Query Operations

### Status Calculation Engine
```go
type StatusSummary struct {
    EpicID              string    `xml:"epic,attr"`
    Status              string    `xml:"status"`
    CompletedPhases     int       `xml:"completed_phases"`
    TotalPhases         int       `xml:"total_phases"`
    PassingTests        int       `xml:"passing_tests"`
    FailingTests        int       `xml:"failing_tests"`
    CompletionPercentage int      `xml:"completion_percentage"`
    CurrentPhase        string    `xml:"current_phase,omitempty"`
    CurrentTask         string    `xml:"current_task,omitempty"`
}
```

### Event Query System
```go
type EventQuery struct {
    Limit       int    `xml:"limit,attr,omitempty"`
    Type        string `xml:"type,attr,omitempty"`
    PhaseID     string `xml:"phase_id,attr,omitempty"`
    Since       string `xml:"since,attr,omitempty"`
}

type Event struct {
    Timestamp string `xml:"timestamp,attr"`
    Agent     string `xml:"agent,attr"`
    PhaseID   string `xml:"phase_id,attr,omitempty"`
    Type      string `xml:"type,attr"`
    Message   string `xml:",chardata"`
}
```

## Implementation Phases

### Phase 2A: Epic Status Analysis (1 day)
- Status calculation engine for epic progress
- Phase completion percentage calculation
- Test status aggregation (passing/failing counts)
- Current active work detection (phase/task)
- Epic status validation and state analysis

### Phase 2B: Current State Intelligence (1 day)
- Active phase and task detection logic
- Next action recommendation engine
- Failing test impact analysis for recommendations
- Work prioritization algorithm
- Context-aware guidance for agents

### Phase 2C: Pending Work Discovery (1 day)
- Pending task enumeration across all phases
- Phase dependency analysis for work ordering
- Test status filtering and categorization
- Work prioritization by phase and dependencies
- Comprehensive pending work reporting

### Phase 2D: Event Querying & Timeline (0.5 days)
- Event chronological ordering and filtering
- Recent activity summarization with limits
- Event type filtering and categorization
- Timeline generation for agent context
- Event metadata extraction and formatting

### Phase 2E: Command Implementation & Integration (0.5 days)
- `agentpm status` command with progress metrics
- `agentpm current` command with next actions
- `agentpm pending` command with work breakdown
- `agentpm failing` command with test details
- `agentpm events` command with timeline filtering

## Acceptance Criteria
- ✅ `agentpm status` shows epic progress with completion percentage
- ✅ `agentpm current` displays active phase/task and next actions
- ✅ `agentpm pending` lists all pending tasks across phases
- ✅ `agentpm failing` shows only failing tests with failure details
- ✅ `agentpm events --limit=5` shows recent activity chronologically
- ✅ All commands support `-f epic-9.xml` to override current epic

## Query Logic & Algorithms

### Progress Calculation
1. **Phase Progress:** Count completed vs total phases
2. **Task Progress:** Count completed vs total tasks across all phases
3. **Test Progress:** Count passing vs total tests
4. **Overall Percentage:** Weighted average of task completion (primary metric)

### Next Action Intelligence
1. **Failing Tests Priority:** If current phase has failing tests → fix tests
2. **Current Task:** If task in progress → continue current task
3. **Next Task in Phase:** If phase active but no current task → start next task
4. **Next Phase:** If current phase complete → start next phase
5. **Epic Complete:** If all work done → completion message

### Pending Work Discovery
1. **Phase Level:** Identify phases with status "pending"
2. **Task Level:** Find tasks with status "pending" in active/pending phases
3. **Test Level:** Find tests with status "pending" or "failing"
4. **Dependency Order:** Order by phase dependencies and task sequences

## Output Examples

### agentpm status
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

### agentpm current
```xml
<current_state epic="8">
    <epic_status>in_progress</epic_status>
    <active_phase>2A</active_phase>
    <active_task>2A_1</active_task>
    <next_action>Fix mobile responsive pagination controls</next_action>
    <failing_tests>1</failing_tests>
</current_state>
```

### agentpm pending
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

### agentpm failing
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

### agentpm events --limit=3
```xml
<events epic="8" limit="3">
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

## Test Scenarios (Key Examples)
- **Status Display:** Show epic progress with accurate completion percentages
- **Current State:** Display active work and intelligent next action recommendations
- **Pending Work:** List all pending items grouped by type with proper ordering
- **Failing Tests:** Show only failing tests with detailed failure information
- **Event Timeline:** Display recent events in chronological order with filtering
- **File Override:** All commands work with `-f` flag for alternate epic files
- **Empty States:** Handle epics with no active work, failing tests, or events

## Performance Considerations
- **Efficient Parsing:** Use XPath queries to avoid loading full DOM for simple operations
- **Caching Strategy:** Cache parsed epic data for multiple query operations
- **Query Optimization:** Minimize XML traversal for status calculations
- **Memory Management:** Proper cleanup of parsed XML structures

## Error Handling
- **Missing Epic:** Clear error when epic file doesn't exist
- **Invalid Epic:** Graceful handling of malformed XML with specific errors
- **Empty Epic:** Appropriate responses for epics with no phases/tasks/tests
- **Calculation Errors:** Safe division and percentage calculations with edge cases

## Quality Gates
- [ ] All acceptance criteria implemented and tested
- [ ] Query performance under 50ms for typical epic files
- [ ] Accurate progress calculations for all epic states
- [ ] Comprehensive error handling for edge cases
- [ ] XML output format consistency across all commands

## Integration with Epic 1
- **Epic Loading:** Reuse Epic 1 epic loading and validation
- **Configuration:** Use Epic 1 config system for current epic detection
- **Storage:** Leverage Epic 1 storage abstraction for file operations
- **Error Handling:** Extend Epic 1 error handling patterns
- **CLI Framework:** Build on Epic 1 command registration system

This specification provides comprehensive read-only query operations while maintaining the simplicity and XML-focused approach established in Epic 1.