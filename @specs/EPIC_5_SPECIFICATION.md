# Epic 5: Test Management & Event Logging - Specification

## Overview
**Goal:** Test status tracking and comprehensive event logging  
**Duration:** 3-4 days  
**Philosophy:** Rich activity tracking with detailed test management and comprehensive audit trails

## User Stories
1. Mark tests as passing or failing with detailed failure information
2. Log important events during development with rich metadata
3. Track file changes and their impact on the project
4. Identify blockers and issues quickly through event analysis
5. Maintain a detailed activity timeline for project history

## Technical Requirements
- **Dependencies:** Epic 1 (CLI, storage), Epic 2 (queries), Epic 3 (events), Epic 4 (phases/tasks)
- **Test Management:** Test status transitions with detailed failure tracking
- **Event System:** Rich event logging with types, metadata, and file tracking
- **File Tracking:** File change monitoring with action types
- **Event Querying:** Advanced filtering and timeline generation
- **Blocker Detection:** Automatic identification of blocking issues

## Test Management System

### Test Status Lifecycle
```go
type TestStatus string

const (
    TestPending  TestStatus = "pending"
    TestPassed   TestStatus = "passed"
    TestFailing  TestStatus = "failing"
    TestSkipped  TestStatus = "skipped"
    TestBlocked  TestStatus = "blocked"
)

type Test struct {
    ID          string     `xml:"id,attr"`
    TaskID      string     `xml:"task_id,attr,omitempty"`
    PhaseID     string     `xml:"phase_id,attr,omitempty"`
    Status      TestStatus `xml:"status,attr"`
    PassedAt    time.Time  `xml:"passed_at,attr,omitempty"`
    FailedAt    time.Time  `xml:"failed_at,attr,omitempty"`
    FailureNote string     `xml:"failure_note,omitempty"`
    Description string     `xml:",chardata"`
}
```

### Test Operations
```go
type TestOperation struct {
    TestID       string `xml:"test_id,attr"`
    PreviousStatus TestStatus `xml:"previous_status,attr"`
    NewStatus    TestStatus `xml:"new_status,attr"`
    Timestamp    time.Time  `xml:"timestamp,attr"`
    FailureNote  string     `xml:"failure_note,omitempty"`
    Agent        string     `xml:"agent,attr"`
}
```

## Event Logging System

### Event Types & Categories
```go
type EventType string

const (
    EventImplementation EventType = "implementation" // Code changes, feature work
    EventTestFailed     EventType = "test_failed"    // Test failures
    EventTestPassed     EventType = "test_passed"    // Test successes
    EventBlocker        EventType = "blocker"        // Blocking issues
    EventIssue          EventType = "issue"          // Problems, bugs
    EventMilestone      EventType = "milestone"      // Important achievements
    EventDecision       EventType = "decision"       // Architecture, design decisions
    EventNote           EventType = "note"           // General observations
)

type Event struct {
    Timestamp   time.Time     `xml:"timestamp,attr"`
    Type        EventType     `xml:"type,attr"`
    Agent       string        `xml:"agent,attr"`
    PhaseID     string        `xml:"phase_id,attr,omitempty"`
    TaskID      string        `xml:"task_id,attr,omitempty"`
    TestID      string        `xml:"test_id,attr,omitempty"`
    Files       []FileChange  `xml:"files>file,omitempty"`
    Message     string        `xml:",chardata"`
}
```

### File Change Tracking
```go
type FileChangeAction string

const (
    FileAdded    FileChangeAction = "added"
    FileModified FileChangeAction = "modified"
    FileDeleted  FileChangeAction = "deleted"
    FileRenamed  FileChangeAction = "renamed"
    FileMoved    FileChangeAction = "moved"
)

type FileChange struct {
    Path   string           `xml:"path,attr"`
    Action FileChangeAction `xml:"action,attr"`
    Lines  int              `xml:"lines,attr,omitempty"`
    Note   string           `xml:"note,omitempty"`
}
```

## Implementation Phases

### Phase 5A: Test Status Management (1 day)
- Test status tracking and transitions
- Test failure information storage and retrieval
- Test status validation and enforcement
- Test history tracking and timeline
- Test-related event generation

### Phase 5B: Event Creation & Logging (1.5 days)
- Rich event logging with metadata
- Event type categorization and validation
- File change tracking and parsing
- Event serialization and storage
- Event validation and consistency checking

### Phase 5C: Event Querying & Filtering (1 day)
- Event timeline generation with filtering
- Event type and timeframe filtering
- Recent activity summarization
- Event search and discovery
- Event aggregation and statistics

### Phase 5D: Command Implementation (0.5 days)
- `agentpm pass-test <test-id>` command
- `agentpm fail-test <test-id> [reason]` command
- `agentpm log <message>` command with options
- Command validation and error handling
- XML output formatting for all commands

## Acceptance Criteria
- ✅ `agentpm pass-test 2A_1` marks test as passed
- ✅ `agentpm fail-test 2A_2 "Mobile responsive issue"` logs failure details
- ✅ `agentpm log "Implemented pagination" --files="src/Pagination.js:added"`
- ✅ `agentpm log "Need design tokens" --type=blocker` identifies blockers
- ✅ Events include timestamps, types, and optional metadata
- ✅ Failed tests and blocker events are easily queryable

## Test Management Operations

### Pass Test Operation
```go
func PassTest(testID string, agent string) (*TestOperation, error) {
    // 1. Find test by ID
    // 2. Validate test exists and is not already passed
    // 3. Update test status to "passed"
    // 4. Clear any failure notes
    // 5. Set passed timestamp
    // 6. Generate test_passed event
    // 7. Save epic file atomically
    // 8. Return operation result
}
```

### Fail Test Operation
```go
func FailTest(testID string, failureNote string, agent string) (*TestOperation, error) {
    // 1. Find test by ID
    // 2. Validate test exists
    // 3. Update test status to "failing"
    // 4. Store failure note
    // 5. Set failed timestamp
    // 6. Generate test_failed event with failure details
    // 7. Save epic file atomically
    // 8. Return operation result
}
```

## Event Logging Operations

### Basic Event Logging
```go
func LogEvent(message string, eventType EventType, options EventOptions) (*Event, error) {
    event := &Event{
        Timestamp: time.Now(),
        Type:      eventType,
        Agent:     options.Agent,
        PhaseID:   options.PhaseID,
        TaskID:    options.TaskID,
        TestID:    options.TestID,
        Files:     options.Files,
        Message:   message,
    }
    
    // Validate event
    // Append to epic events
    // Save epic file
    return event, nil
}
```

### File Change Parsing
```go
func ParseFileChanges(fileSpec string) ([]FileChange, error) {
    // Parse format: "file1.js:added,file2.css:modified,file3.py:deleted"
    // Validate file paths and actions
    // Return structured file changes
}
```

## Output Examples

### agentpm pass-test 2A_1
```xml
<test_updated epic="8" test="2A_1">
    <test_description>Basic pagination navigation test</test_description>
    <phase_id>2A</phase_id>
    <previous_status>pending</previous_status>
    <new_status>passed</new_status>
    <updated_at>2025-08-16T14:45:00Z</updated_at>
    <message>Test 2A_1 marked as passed</message>
</test_updated>
```

### agentpm fail-test 2A_2 "Touch targets too small"
```xml
<test_updated epic="8" test="2A_2">
    <test_description>Mobile pagination controls test</test_description>
    <phase_id>2A</phase_id>
    <previous_status>pending</previous_status>
    <new_status>failing</new_status>
    <failure_reason>Touch targets too small</failure_reason>
    <updated_at>2025-08-16T14:50:00Z</updated_at>
    <message>Test 2A_2 marked as failing: Touch targets too small</message>
</test_updated>
```

### agentpm log "Implemented pagination controls" --files="src/Pagination.js:added"
```xml
<event_logged epic="8">
    <timestamp>2025-08-16T14:30:00Z</timestamp>
    <agent>agent_claude</agent>
    <phase_id>2A</phase_id>
    <type>implementation</type>
    <files>
        <file path="src/Pagination.js" action="added"/>
    </files>
    <message>Event logged successfully</message>
</event_logged>
```

### agentpm log "Found accessibility issue" --type=blocker
```xml
<event_logged epic="8">
    <timestamp>2025-08-16T14:45:00Z</timestamp>
    <agent>agent_claude</agent>
    <phase_id>2A</phase_id>
    <type>blocker</type>
    <message>Blocker logged successfully</message>
</event_logged>
```

## Event Querying & Analysis

### Event Timeline Query
```go
func QueryEvents(epic *Epic, options QueryOptions) ([]Event, error) {
    events := []Event{}
    
    // Filter by type if specified
    if options.Type != "" {
        events = filterByType(epic.Events, options.Type)
    } else {
        events = epic.Events
    }
    
    // Filter by timeframe if specified
    if !options.Since.IsZero() {
        events = filterBySince(events, options.Since)
    }
    
    // Sort chronologically (newest first)
    sort.Slice(events, func(i, j int) bool {
        return events[i].Timestamp.After(events[j].Timestamp)
    })
    
    // Apply limit if specified
    if options.Limit > 0 && len(events) > options.Limit {
        events = events[:options.Limit]
    }
    
    return events, nil
}
```

### Blocker Detection
```go
func FindBlockers(epic *Epic) ([]Blocker, error) {
    blockers := []Blocker{}
    
    // Find failing tests
    for _, test := range epic.Tests {
        if test.Status == TestFailing {
            blockers = append(blockers, Blocker{
                Type:        "failing_test",
                Source:      test.ID,
                Description: test.FailureNote,
                CreatedAt:   test.FailedAt,
            })
        }
    }
    
    // Find blocker events
    for _, event := range epic.Events {
        if event.Type == EventBlocker {
            blockers = append(blockers, Blocker{
                Type:        "logged_blocker",
                Source:      "event",
                Description: event.Message,
                CreatedAt:   event.Timestamp,
            })
        }
    }
    
    return blockers, nil
}
```

## Command Options & Flags

### Log Command Options
```bash
agentpm log <message> [options]

Options:
  --type=<type>           Event type (implementation, blocker, issue, note, etc.)
  --files=<file-spec>     File changes (format: "file1:action,file2:action")
  --phase=<phase-id>      Associate with specific phase
  --task=<task-id>        Associate with specific task
  --test=<test-id>        Associate with specific test

Examples:
  agentpm log "Implemented pagination"
  agentpm log "Fixed mobile layout" --type=implementation --files="src/Mobile.css:modified"
  agentpm log "Need design approval" --type=blocker --phase=2A
  agentpm log "Discovered edge case" --type=issue --task=2A_1
```

### Event Query Options
```bash
agentpm events [options]

Options:
  --limit=<n>             Maximum number of events to return
  --type=<type>           Filter by event type
  --since=<timespec>      Events since specified time
  --phase=<phase-id>      Events for specific phase
  --task=<task-id>        Events for specific task

Examples:
  agentpm events --limit=10
  agentpm events --type=blocker
  agentpm events --since="2 hours ago"
  agentpm events --phase=2A --limit=5
```

## File Change Specifications

### File Change Format
```
Format: "path1:action1,path2:action2,path3:action3"

Actions:
- added      - New file created
- modified   - Existing file changed
- deleted    - File removed
- renamed    - File renamed (use path:renamed:newpath)
- moved      - File moved to new location

Examples:
- "src/App.js:modified"
- "src/NewComponent.js:added,src/OldComponent.js:deleted"
- "src/Component.js:renamed:src/NewComponent.js"
```

### File Change Validation
```go
func ValidateFileChanges(fileSpec string) error {
    // Validate format: "file:action,file:action"
    // Check valid actions
    // Validate file paths
    // Ensure no duplicate files
    return nil
}
```

## Integration with Previous Epics

### Epic 1 Integration
- **Storage:** Use Epic 1 storage abstraction for atomic test/event updates
- **Epic Loading:** Leverage Epic 1 epic loading for test and event access
- **Validation:** Extend Epic 1 validation for test structure

### Epic 2 Integration
- **Failing Tests:** Integrate with Epic 2 failing test queries
- **Event Timeline:** Extend Epic 2 event querying with rich filtering
- **Status Analysis:** Include test status in Epic 2 status calculations

### Epic 3 Integration
- **Event System:** Build on Epic 3 event logging infrastructure
- **Lifecycle Events:** Integrate test events with Epic 3 lifecycle events
- **Atomic Operations:** Use Epic 3 atomic operations for test updates

### Epic 4 Integration
- **Phase/Task Context:** Associate tests and events with phases/tasks
- **Progress Calculation:** Include test status in Epic 4 progress metrics
- **Current State:** Integrate test status with Epic 4 current state

## Test Scenarios (Key Examples)
- **Test Management:** Mark tests passing/failing with detailed failure information
- **Event Logging:** Log events with rich metadata, file changes, and context
- **File Tracking:** Track file changes with proper action categorization
- **Event Querying:** Query events by type, timeframe, and context
- **Blocker Detection:** Identify blockers from failing tests and logged events
- **Timeline Generation:** Create chronological activity timelines
- **Context Association:** Associate events with phases, tasks, and tests

## Validation Rules

### Test Status Validation
1. **Test Exists:** Test ID must exist in epic
2. **Valid Transition:** Test status transitions must be valid
3. **Failure Note:** Failing tests should include failure reason
4. **Agent Context:** Current agent must be provided

### Event Validation
1. **Valid Type:** Event type must be from approved list
2. **Message Required:** Event message cannot be empty
3. **File Format:** File changes must follow proper format
4. **Context Validation:** Phase/task/test IDs must exist if provided

### File Change Validation
1. **Valid Actions:** File actions must be from approved list
2. **Path Format:** File paths must be valid and reasonable
3. **No Duplicates:** Same file cannot appear multiple times
4. **Consistent Format:** All file changes must follow format

## Quality Gates
- [ ] All acceptance criteria implemented and tested
- [ ] Test status tracking with detailed failure information
- [ ] Rich event logging with metadata and file tracking
- [ ] Event querying with filtering and timeline generation
- [ ] Blocker detection from tests and events

## Performance Considerations
- **Event Storage:** Efficient event appending and querying
- **Timeline Generation:** Fast chronological sorting and filtering
- **File Parsing:** Efficient file change specification parsing
- **Blocker Detection:** Fast identification of blocking issues

This specification provides comprehensive test management and rich event logging while maintaining integration with all previous epic foundations.