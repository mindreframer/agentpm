# EPIC-4 SPECIFICATION: Test Management & Event Logging

## Overview

**Epic ID:** 4
**Name:** Test Management & Event Logging  
**Duration:** 3-4 days  
**Status:** pending  
**Priority:** high  
**Depends On:** Epic 1 (Foundation), Epic 2 (Query Commands), Epic 3 (Epic Lifecycle)

**Goal:** Implement comprehensive test status tracking and detailed event logging capabilities, enabling agents to manage test outcomes and maintain detailed activity timelines for development transparency and handoff purposes.

## Business Context

Epic 4 is identified in the roadmap as one of the "most important epics" alongside Epic 5. It provides agents with the ability to track test results and log detailed development activities. This epic is crucial for maintaining development quality through test management and ensuring comprehensive documentation of all development activities through rich event logging. The combination of test tracking and event logging provides complete visibility into both the quality and process aspects of development work.

## User Stories

### Primary User Stories
- **As an agent, I can start working on a test** so that I can track my testing activities and progress
- **As an agent, I can mark tests as passing or failing with details** so that I can maintain accurate quality metrics
- **As an agent, I can log important events during development** so that I can document my work for handoff and review
- **As an agent, I can track file changes and their impact** so that I can maintain detailed change logs
- **As an agent, I can identify blockers and issues quickly** so that I can communicate obstacles effectively

### Secondary User Stories
- **As an agent, I can maintain a detailed activity timeline** so that handoff agents understand the full development context
- **As an agent, I can categorize events by type** so that different kinds of activities can be filtered and analyzed
- **As an agent, I can log events with rich metadata** so that context is preserved for future reference
- **As an agent, I can cancel tests when specifications change** so that I can handle evolving requirements

## Technical Requirements

### Core Dependencies
- **Foundation:** Epic 1 CLI framework, XML processing, storage interface
- **Querying:** Epic 2 query service for test status queries (already implemented)
- **Lifecycle:** Epic 3 lifecycle management for epic-level validation
- **Event System:** Enhanced event logging with rich metadata and categorization

### Architecture Principles
- **Test Lifecycle:** Simple pending → wip → passed/failed/cancelled transitions
- **Rich Event Logging:** Detailed events with types, metadata, and file tracking
- **Simple Operations:** Minimal output for routine test updates
- **Detailed Context:** Rich information for blocker identification and handoff
- **Automatic Integration:** Test events automatically logged, manual events supported

### Test State Rules
```
TEST STATES: pending → wip
             wip → failed
             wip → passed
             wip → cancelled
             passed → failed
             failed → passed

CONSTRAINTS:
- Test can only be started if its associated task/phase is active or completed
- Multiple tests can be worked on simultaneously (unlike tasks)
- Test status transitions are independent of each other
- Failed tests create automatic blocker events
```

### Event Types & Categories
```
EVENT TYPES:
- test_started: When test transitions to wip
- test_passed: When test transitions to passed
- test_failed: When test transitions to failed (with failure details)
- test_cancelled: When test transitions to cancelled
- implementation: Manual logging of implementation work
- blocker: Manual logging of blocking issues
- file_change: Manual logging of file modifications
- milestone: Manual logging of significant achievements
```

## Functional Requirements

### FR-1: Test Status Management
**Commands:**
- `agentpm start-test <test-id> [--time <timestamp>] [--file <epic-file>]`
- `agentpm pass-test <test-id> [--time <timestamp>] [--file <epic-file>]`
- `agentpm fail-test <test-id> "<failure-reason>" [--time <timestamp>] [--file <epic-file>]`
- `agentpm cancel-test <test-id> "<cancellation-reason>" [--time <timestamp>] [--file <epic-file>]`

**Start Test Behavior:**
- Changes test status from "pending" to "wip"
- Sets test started_at timestamp
- Creates automatic event log for test start
- Returns simple confirmation message

**Pass Test Behavior:**
- Changes test status from "wip" to "passed"
- Sets test passed_at timestamp
- Clears any previous failure_note
- Creates automatic event log for test pass

**Fail Test Behavior:**
- Changes test status from "wip" to "failed"
- Sets test failed_at timestamp and failure_note
- Creates automatic event log for test failure
- Creates automatic blocker event if failure impacts progress

**Cancel Test Behavior:**
- Changes test status from "wip" to "cancelled"
- Sets test cancelled_at timestamp and cancellation_reason
- Creates automatic event log for test cancellation

**Output Format (Simple Confirmations):**
```
Test 2A_1 started.
Test 2A_1 passed.
Test 2A_2 failed: Mobile responsive design not working.
Test 2A_3 cancelled: Spec contradicts itself with point xyz.
```

### FR-2: Event Logging System
**Command:** `agentpm log "<message>" [--type <event-type>] [--files "<file-changes>"] [--time <timestamp>] [--file <epic-file>]`

**Behavior:**
- Creates manual event entry with specified message
- Supports event type categorization (implementation, blocker, file_change, milestone)
- Supports file change tracking with action metadata
- Defaults to "implementation" type if not specified
- Associates event with current active phase/task if available
- Sets event timestamp (from --time flag or current time)

**Event Types:**
- **implementation** (default): General development work
- **blocker**: Issues that prevent progress
- **file_change**: File modifications and additions
- **milestone**: Significant achievements or completion markers

**File Change Format:**
- Single file: `--files="src/Pagination.js:added"`
- Multiple files: `--files="src/File1.js:modified,src/File2.js:deleted,src/File3.js:added"`
- Supported actions: added, modified, deleted, renamed

**Output Format:**
```xml
<event_logged epic="8">
    <timestamp>2025-08-16T14:30:00Z</timestamp>
    <agent>agent_claude</agent>
    <phase_id>2A</phase_id>
    <type>implementation</type>
    <message>Event logged successfully</message>
</event_logged>
```

### FR-3: Rich Event Structure
**Event XML Format:**
```xml
<event timestamp="2025-08-16T14:30:00Z" agent="agent_claude" phase_id="2A" type="implementation">
    Implemented pagination controls
    
    Files: src/components/Pagination.js (added), src/styles/pagination.css (added)
    Result: Basic controls working, all tests passing
</event>

<event timestamp="2025-08-16T14:45:00Z" agent="agent_claude" phase_id="2A" type="test_failed">
    Mobile responsive test failing
    
    Test: 2A_2 - Mobile pagination controls
    Issue: Touch targets too small, need 44px+ minimum
</event>

<event timestamp="2025-08-16T15:00:00Z" agent="agent_claude" phase_id="2A" type="blocker">
    Found design system dependency
    
    Blocker: Need design system tokens for mobile responsive design
    Impact: Cannot complete mobile responsiveness without design system
</event>
```

### FR-4: Test Data Model & Integration
**Test XML Structure:**
```xml
<test id="2A_1" phase_id="2A" task_id="2A_1" status="passed">
    <description>
        **GIVEN** I'm on mobile device  
        **WHEN** I tap pagination controls  
        **THEN** They work and are easy to tap (44px+ targets)
    </description>
</test>
```

**Integration with Epic 2 Queries:**
- `agentpm failing` command already implemented in Epic 2
- Enhanced to show failure_note details
- Blocker events automatically created for failed tests
- Test status updates integrate with existing progress calculation

### FR-5: Blocker Detection & Reporting
**Automatic Blocker Creation:**
- Failed tests automatically create blocker events
- Manual blocker events via `agentpm log --type=blocker`
- Blocker events include impact assessment when possible

**Blocker Event Format:**
```xml
<event timestamp="2025-08-16T15:00:00Z" agent="agent_claude" phase_id="2A" type="blocker">
    Mobile responsive test failing
    
    Test: 2A_2 - Mobile pagination controls  
    Issue: Touch targets too small, need 44px+ minimum
    Impact: Blocks completion of mobile responsiveness task
    Suggestion: Research design system touch target standards
</event>
```

## Non-Functional Requirements

### NFR-1: Performance
- Test management commands execute in < 100ms for typical epic files
- Event logging performs efficiently without impacting workflow
- Event queries perform well with large event histories (1000+ events)

### NFR-2: Reliability
- All test status transitions are validated before execution
- Event logging is atomic and cannot corrupt epic files
- File change parsing handles malformed input gracefully
- Event timestamps are consistent and properly formatted

### NFR-3: Usability (for Agents)
- Simple confirmation messages for routine test updates
- Rich event context for development documentation
- Clear error messages for invalid test transitions
- Flexible event logging with optional metadata

### NFR-4: Integration
- Seamless integration with Epic 2 query commands
- Automatic event creation for test transitions
- Event logging preserves context from active phase/task
- Blocker events integrate with handoff reporting

## Data Model Changes

### Enhanced Event Structure
```xml
<events>
    <event timestamp="2025-08-16T14:30:00Z" agent="agent_claude" phase_id="2A" task_id="2A_1" type="implementation">
        <message>Implemented pagination controls</message>
        <files>
            <file action="added">src/components/Pagination.js</file>
            <file action="added">src/styles/pagination.css</file>
        </files>
        <metadata>
            <result>Basic controls working, all tests passing</result>
        </metadata>
    </event>
    
    <event timestamp="2025-08-16T14:45:00Z" agent="agent_claude" phase_id="2A" test_id="2A_2" type="test_failed">
        <message>Mobile responsive test failing</message>
        <test_details>
            <test_id>2A_2</test_id>
            <test_description>Mobile pagination controls</test_description>
            <failure_reason>Touch targets too small, need 44px+ minimum</failure_reason>
        </test_details>
    </event>
</events>
```

### Test Status Enhancements
```xml
<tests>
    <test id="2A_1" phase_id="2A" task_id="2A_1" status="passed">
        <timestamps>
            <started_at>2025-08-16T14:20:00Z</started_at>
            <passed_at>2025-08-16T14:25:00Z</passed_at>
        </timestamps>
        <description>Basic pagination navigation test</description>
    </test>
    
    <test id="2A_2" phase_id="2A" task_id="2A_1" status="failed">
        <timestamps>
            <started_at>2025-08-16T14:40:00Z</started_at>
            <failed_at>2025-08-16T14:50:00Z</failed_at>
        </timestamps>
        <failure_note>Touch targets too small, need 44px+ minimum</failure_note>
        <description>Mobile pagination controls test</description>
    </test>
</tests>
```

## Error Handling

### Error Categories
1. **Test State Errors:** Invalid status transitions, test not found
2. **Event Logging Errors:** Malformed file change syntax, invalid event types
3. **Integration Errors:** Test/task/phase relationship issues
4. **Validation Errors:** Missing required parameters, invalid timestamps

### Error Response Examples

**Invalid Test Transition:**
```xml
<error>
    <type>invalid_test_transition</type>
    <message>Cannot pass test 2A_2: test is not currently in progress</message>
    <details>
        <test_id>2A_2</test_id>
        <current_status>pending</current_status>
        <suggestion>Use 'agentpm start-test 2A_2' to begin the test first</suggestion>
    </details>
</error>
```

**Malformed File Change:**
```xml
<error>
    <type>invalid_file_format</type>
    <message>Invalid file change format: "invalid-format"</message>
    <details>
        <expected_format>filename:action</expected_format>
        <valid_actions>added, modified, deleted, renamed</valid_actions>
        <example>src/File.js:modified</example>
    </details>
</error>
```

## Acceptance Criteria

### AC-1: Start Test
- **GIVEN** I have test "2A_1" in "pending" status
- **WHEN** I run `agentpm start-test 2A_1`
- **THEN** test "2A_1" should change to "wip" status and event should be logged

### AC-2: Pass Test
- **GIVEN** I have test "2A_1" in "wip" status
- **WHEN** I run `agentpm pass-test 2A_1`
- **THEN** test should change to "passed" and event should be logged

### AC-3: Fail Test with Details
- **GIVEN** I have test "2A_2" in "wip" status
- **WHEN** I run `agentpm fail-test 2A_2 "Mobile responsive design not working"`
- **THEN** test should change to "failed" with failure note recorded

### AC-4: Cancel Test with Reason
- **GIVEN** I have test "2A_3" in "wip" status
- **WHEN** I run `agentpm cancel-test 2A_3 "Spec contradicts itself with point xyz"`
- **THEN** test should change to "cancelled" with cancellation reason recorded

### AC-5: Log Implementation Event
- **GIVEN** I am working on a task
- **WHEN** I run `agentpm log "Implemented pagination controls"`
- **THEN** an event should be created with type "implementation" and the message

### AC-6: Log Event with File Changes
- **GIVEN** I have made changes to files
- **WHEN** I run `agentpm log "Added pagination component" --files="src/Pagination.js:added,src/styles.css:modified"`
- **THEN** event should include file change metadata

### AC-7: Log Blocker Event
- **GIVEN** I encounter a blocking issue
- **WHEN** I run `agentpm log "Need design system tokens" --type=blocker`
- **THEN** event should be created with type "blocker" for easy identification

### AC-8: Automatic Blocker for Failed Test
- **GIVEN** I fail a test
- **WHEN** I run `agentpm fail-test 2A_2 "Critical functionality broken"`
- **THEN** both a test_failed event and a blocker event should be created

### AC-9: Event Integration with Active Context
- **GIVEN** I have phase "2A" and task "2A_1" active
- **WHEN** I log any event
- **THEN** the event should include phase_id and task_id context

### AC-10: File Change Format Validation
- **GIVEN** I specify invalid file change format
- **WHEN** I use `--files="invalid-format"`
- **THEN** I should get an error about expected format "filename:action"

## Testing Strategy

### Test Categories
- **Unit Tests (70%):** Test state transitions, event creation, validation logic
- **Integration Tests (25%):** Command execution, file operations, Epic 2 integration
- **Event Processing Tests (5%):** Event parsing, file change tracking, blocker detection

### Test Data Requirements
- Epic files with tests in various states (pending, wip, passed, failed, cancelled)
- Epic files with rich event histories
- Test files with file change metadata
- Invalid event data for error handling testing

### Test Isolation
- Each test uses isolated epic files in `t.TempDir()`
- MemoryStorage for fast unit tests
- Deterministic timestamps for consistent snapshots
- No shared state between tests

## Implementation Phases

### Phase 4A: Event Logging System (Day 2-3)
- Enhanced event structure with metadata support
- Manual event logging command implementation
- File change tracking and parsing
- Event type categorization and validation
- Integration with current phase/task context

### Phase 4B: Test Management Foundation (Day 1-2)
- Create internal/tests package
- Implement TestService with Storage injection
- Test state validation and transition logic
- Test start/pass/fail/cancel command implementation
- Basic event logging for test operations

### Phase 4C: Blocker Detection & Rich Events (Day 3)
- Automatic blocker event creation for failed tests
- Rich event formatting with detailed metadata
- Event querying enhancement for Epic 2 integration
- Blocker identification and reporting utilities

### Phase 4D: Integration & Performance (Day 3-4)
- Integration with Epic 2 failing command enhancements
- Cross-command consistency and error handling
- Performance optimization for event processing
- Comprehensive testing and documentation

## Definition of Done

- [ ] All acceptance criteria verified with automated tests
- [ ] Test management commands execute in < 100ms for typical epic files
- [ ] Test coverage > 90% for test management and event logging
- [ ] Event logging supports all specified types and metadata
- [ ] File change tracking parses all valid formats correctly
- [ ] Automatic blocker creation works for failed tests
- [ ] Integration with Epic 2 query commands works seamlessly
- [ ] Performance requirements met for large event histories

## Dependencies and Risks

### Dependencies
- **Epic 1:** CLI framework, XML processing, storage interface
- **Epic 2:** Query service for failing tests display (enhancement needed)
- **Epic 3:** Epic lifecycle for overall epic state management

### Risks
- **Low Risk:** Event logging performance with large event histories
- **Low Risk:** File change parsing complexity
- **Low Risk:** Integration complexity with existing Epic 2 commands

### Mitigation Strategies
- Efficient event processing with minimal XML parsing overhead
- Comprehensive file change format validation and error handling
- Careful integration testing with Epic 2 query enhancements
- Performance testing with large event datasets

## Future Considerations

### Potential Enhancements (Not in Scope)
- Event filtering and search capabilities
- Event export functionality
- Test execution automation integration
- Advanced blocker impact analysis
- Event-based notifications and alerts
