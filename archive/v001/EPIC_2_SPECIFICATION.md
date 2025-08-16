# Epic 2: Query & Status Commands - Detailed Technical Specification

## Overview

**Epic ID:** 2  
**Name:** Query & Status Commands  
**Duration:** 3-4 days  
**Goal:** Read-only operations to query epic state and progress  
**Status:** Planning  
**Dependencies:** Epic 1 (Foundation & Configuration)

## Executive Summary

Epic 2 implements comprehensive read-only query operations that enable agents to understand epic state, progress, and current context. This epic focuses on efficient XML querying, structured output formatting, and intelligent data filtering to provide agents with actionable information about their work state.

---

## User Stories & Acceptance Criteria

### Story 1: Epic Status Overview
**As an agent, I can view overall epic status and progress**

#### Acceptance Criteria:
- ✅ `agentpm status` shows epic progress with completion percentage
- ✅ Displays completed vs total phases, tasks, and tests
- ✅ Shows current epic status (planning, in_progress, paused, completed)
- ✅ Calculates and displays progress metrics accurately
- ✅ Supports `-f epic-9.xml` to override current epic
- ✅ Handles missing or invalid epic files gracefully

#### Technical Implementation:
```go
// Progress calculation using XPath queries
type ProgressMetrics struct {
    CompletedPhases int `xml:"completed_phases"`
    TotalPhases     int `xml:"total_phases"`
    PassingTests    int `xml:"passing_tests"`
    FailingTests    int `xml:"failing_tests"`
    CompletionPercentage int `xml:"completion_percentage"`
}

// XPath-based progress calculation (memory efficient)
func CalculateProgressWithXPath(doc *etree.Document) ProgressMetrics {
    // Use XPath to count elements without loading full structs
    totalTasks := len(doc.FindElements("//task"))
    completedTasks := len(doc.FindElements("//task[@status='completed']"))
    
    totalTests := len(doc.FindElements("//test"))
    passingTests := len(doc.FindElements("//test[@status='passed']"))
    failingTests := len(doc.FindElements("//test[@status='failing']"))
    
    totalPhases := len(doc.FindElements("//phase"))
    completedPhases := len(doc.FindElements("//phase[@status='completed']"))
    
    percentage := 0
    if totalTasks > 0 {
        percentage = (completedTasks * 100) / totalTasks
    }
    
    return ProgressMetrics{
        CompletedPhases:      completedPhases,
        TotalPhases:         totalPhases,
        PassingTests:        passingTests,
        FailingTests:        failingTests,
        CompletionPercentage: percentage,
    }
}
```

#### Output Format:
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

---

### Story 2: Current Work Context
**As an agent, I can see what I'm currently working on**

#### Acceptance Criteria:
- ✅ `agentpm current` displays active phase/task and next actions
- ✅ Shows epic status and current work assignments
- ✅ Provides intelligent next action recommendations
- ✅ Displays failing tests count and priority indicators
- ✅ Handles epics with no active work gracefully
- ✅ Supports file override with `-f` flag

#### Next Action Logic (XPath Implementation):
1. `//test[@status='failing']` exists → "Fix failing tests"
2. `//task[@status='in_progress']` exists → "Continue task: [task_description]" 
3. Active phase `//phase[@status='in_progress']` with `//task[@phase_id='X' and @status='pending']` → "Start next task in phase"
4. `//phase[@status='pending']` exists → "Start first pending phase"
5. All work complete → "Epic ready for completion"

```go
// XPath-based next action using etree queries
func GetNextActionWithXPath(doc *etree.Document) string {
    // Priority 1: XPath query for failing tests
    if failingTests := doc.FindElements("//test[@status='failing']"); len(failingTests) > 0 {
        return fmt.Sprintf("Fix %d failing test(s)", len(failingTests))
    }
    
    // Priority 2: XPath query for active task
    if activeTask := doc.FindElement("//task[@status='in_progress']"); activeTask != nil {
        return fmt.Sprintf("Continue task: %s", activeTask.Text())
    }
    
    // Continue with other XPath-based priorities...
}
```

#### Output Format:
```xml
<current_state epic="8">
    <epic_status>in_progress</epic_status>
    <active_phase>2A</active_phase>
    <active_task>2A_1</active_task>
    <next_action>Fix mobile responsive pagination controls</next_action>
    <failing_tests>1</failing_tests>
</current_state>
```

---

### Story 3: Pending Work Queue
**As an agent, I can list all pending tasks**

#### Acceptance Criteria:
- ✅ `agentpm pending` lists all pending tasks across phases
- ✅ Groups pending work by phases for organization
- ✅ Shows pending phases, tasks, and tests separately
- ✅ Prioritizes work based on phase order and dependencies
- ✅ Handles epics with no pending work gracefully
- ✅ Provides clear empty state messages

#### Output Format:
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

---

### Story 4: Failing Tests Identification
**As an agent, I can identify failing tests that need attention**

#### Acceptance Criteria:
- ✅ `agentpm failing` shows only failing tests with failure details
- ✅ Displays test descriptions and failure notes
- ✅ Groups failing tests by phase for context
- ✅ Shows failure timestamps when available
- ✅ Handles epics with no failing tests gracefully
- ✅ Provides actionable failure information

#### Output Format:
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

---

### Story 5: Recent Activity Timeline
**As an agent, I can review recent activity and events**

#### Acceptance Criteria:
- ✅ `agentpm events --limit=5` shows recent activity chronologically
- ✅ Displays events in reverse chronological order (newest first)
- ✅ Supports configurable limit for number of events
- ✅ Shows event types, timestamps, and descriptions
- ✅ Includes agent attribution and phase context
- ✅ Handles epics with no events gracefully

#### Event Types:
- `implementation` - Code changes and feature additions
- `test_failed` - Test failures and validation issues
- `test_passed` - Test successes and validations
- `blocker` - Blocking issues and dependencies
- `milestone` - Phase/task completions and major progress

#### Output Format:
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

---

## Technical Architecture

### 1. Query Engine Design

#### Epic Query Interface
```go
type EpicQuerier interface {
    // Status queries
    GetStatus() (*StatusResult, error)
    GetCurrentState() (*CurrentStateResult, error)
    
    // Work queries
    GetPendingWork() (*PendingWorkResult, error)
    GetFailingTests() (*FailingTestsResult, error)
    
    // Event queries
    GetRecentEvents(limit int) (*EventsResult, error)
    GetEventsByType(eventType string, limit int) (*EventsResult, error)
    GetEventsSince(since time.Time) (*EventsResult, error)
}
```

#### Progress Calculation Engine
```go
type ProgressCalculator struct {
    epic *Epic
}

func (pc *ProgressCalculator) CalculateCompletion() ProgressMetrics {
    totalTasks := len(pc.epic.Tasks)
    completedTasks := pc.countTasksByStatus("completed")
    
    totalTests := len(pc.epic.Tests)
    passingTests := pc.countTestsByStatus("passed")
    failingTests := pc.countTestsByStatus("failing")
    
    totalPhases := len(pc.epic.Phases)
    completedPhases := pc.countPhasesByStatus("completed")
    
    percentage := 0
    if totalTasks > 0 {
        percentage = (completedTasks * 100) / totalTasks
    }
    
    return ProgressMetrics{
        CompletedPhases:      completedPhases,
        TotalPhases:         totalPhases,
        PassingTests:        passingTests,
        FailingTests:        failingTests,
        CompletionPercentage: percentage,
    }
}
```

#### Next Action Recommendation Engine
```go
type ActionRecommender struct {
    epic *Epic
}

func (ar *ActionRecommender) GetNextAction() string {
    // Priority 1: Fix failing tests
    if failingTests := ar.epic.GetFailingTests(); len(failingTests) > 0 {
        return fmt.Sprintf("Fix %d failing test(s)", len(failingTests))
    }
    
    // Priority 2: Continue active task
    if activeTask := ar.epic.GetActiveTask(); activeTask != nil {
        return fmt.Sprintf("Continue task: %s", activeTask.Description)
    }
    
    // Priority 3: Start next task in active phase
    if activePhase := ar.epic.GetActivePhase(); activePhase != nil {
        if pendingTasks := activePhase.GetPendingTasks(); len(pendingTasks) > 0 {
            return fmt.Sprintf("Start task: %s", pendingTasks[0].Description)
        }
    }
    
    // Priority 4: Start next phase
    if pendingPhases := ar.epic.GetPendingPhases(); len(pendingPhases) > 0 {
        return fmt.Sprintf("Start phase: %s", pendingPhases[0].Name)
    }
    
    // Priority 5: Complete epic
    return "Epic ready for completion"
}
```

### 2. Data Structures

#### Query Results
```go
type StatusResult struct {
    Epic     *Epic           `xml:"epic"`
    Progress ProgressMetrics `xml:"progress"`
    Status   string          `xml:"status"`
    Name     string          `xml:"name"`
    CurrentPhase string      `xml:"current_phase"`
    CurrentTask  string      `xml:"current_task"`
}

type CurrentStateResult struct {
    EpicStatus   string `xml:"epic_status"`
    ActivePhase  string `xml:"active_phase"`
    ActiveTask   string `xml:"active_task"`
    NextAction   string `xml:"next_action"`
    FailingTests int    `xml:"failing_tests"`
}

type PendingWorkResult struct {
    Phases []Phase `xml:"phases>phase"`
    Tasks  []Task  `xml:"tasks>task"`
    Tests  []Test  `xml:"tests>test"`
}

type FailingTestsResult struct {
    Tests []FailingTest `xml:"test"`
}

type FailingTest struct {
    ID          string `xml:"id,attr"`
    PhaseID     string `xml:"phase_id,attr"`
    Description string `xml:"description"`
    FailureNote string `xml:"failure_note"`
    FailedAt    time.Time `xml:"failed_at,attr"`
}

type EventsResult struct {
    Events []Event `xml:"event"`
    Limit  int     `xml:"limit,attr"`
}
```

#### Enhanced Epic Structures
```go
// Extensions to Epic struct from Epic 1
type Epic struct {
    // ... existing fields from Epic 1
    
    // State tracking
    ActivePhaseID string `xml:"active_phase_id,attr"`
    ActiveTaskID  string `xml:"active_task_id,attr"`
    StartedAt     *time.Time `xml:"started_at,attr"`
    CompletedAt   *time.Time `xml:"completed_at,attr"`
}

type Phase struct {
    ID          string    `xml:"id,attr"`
    Name        string    `xml:"name,attr"`
    Status      string    `xml:"status,attr"` // pending, in_progress, completed
    StartedAt   *time.Time `xml:"started_at,attr"`
    CompletedAt *time.Time `xml:"completed_at,attr"`
    Description string    `xml:",chardata"`
}

type Task struct {
    ID          string    `xml:"id,attr"`
    PhaseID     string    `xml:"phase_id,attr"`
    Status      string    `xml:"status,attr"` // pending, in_progress, completed
    StartedAt   *time.Time `xml:"started_at,attr"`
    CompletedAt *time.Time `xml:"completed_at,attr"`
    Description string    `xml:",chardata"`
}

type Test struct {
    ID          string    `xml:"id,attr"`
    TaskID      string    `xml:"task_id,attr"`
    PhaseID     string    `xml:"phase_id,attr"`
    Status      string    `xml:"status,attr"` // pending, passed, failing
    Description string    `xml:"description"`
    FailureNote string    `xml:"failure_note,omitempty"`
    UpdatedAt   time.Time `xml:"updated_at,attr"`
}

type Event struct {
    ID        string            `xml:"id,attr"`
    Timestamp time.Time         `xml:"timestamp,attr"`
    Agent     string            `xml:"agent,attr"`
    PhaseID   string            `xml:"phase_id,attr"`
    TaskID    string            `xml:"task_id,attr,omitempty"`
    Type      string            `xml:"type,attr"` // implementation, test_failed, blocker, milestone
    Message   string            `xml:",chardata"`
    Metadata  map[string]string `xml:"metadata,omitempty"`
}
```

### 3. Query Optimization Using etree XPath

#### XPath Query Engine
```go
import "github.com/beevik/etree"

type XPathQuerier struct {
    doc *etree.Document
}

func NewXPathQuerier(epicPath string) (*XPathQuerier, error) {
    doc := etree.NewDocument()
    if err := doc.ReadFromFile(epicPath); err != nil {
        return nil, fmt.Errorf("failed to load epic: %w", err)
    }
    
    return &XPathQuerier{doc: doc}, nil
}

// Efficient XPath queries without loading full structs
func (xq *XPathQuerier) GetTasksByStatus(status string) ([]*etree.Element, error) {
    // XPath: //task[@status='pending']
    xpath := fmt.Sprintf("//task[@status='%s']", status)
    return xq.doc.FindElements(xpath), nil
}

func (xq *XPathQuerier) GetFailingTests() ([]*etree.Element, error) {
    // XPath: //test[@status='failing']
    return xq.doc.FindElements("//test[@status='failing']"), nil
}

func (xq *XPathQuerier) GetActivePhase() *etree.Element {
    // XPath: //phase[@status='in_progress']
    return xq.doc.FindElement("//phase[@status='in_progress']")
}

func (xq *XPathQuerier) GetActiveTask() *etree.Element {
    // XPath: //task[@status='in_progress']
    return xq.doc.FindElement("//task[@status='in_progress']")
}

func (xq *XPathQuerier) GetRecentEvents(limit int) ([]*etree.Element, error) {
    // XPath: //event[position() <= $limit]
    // Note: etree sorts by document order, we'll need timestamp sorting
    allEvents := xq.doc.FindElements("//event")
    
    // Sort by timestamp (newest first)
    sort.Slice(allEvents, func(i, j int) bool {
        timeI := allEvents[i].SelectAttrValue("timestamp", "")
        timeJ := allEvents[j].SelectAttrValue("timestamp", "")
        return timeI > timeJ // Reverse chronological
    })
    
    if len(allEvents) > limit {
        return allEvents[:limit], nil
    }
    
    return allEvents, nil
}

func (xq *XPathQuerier) CountByStatus(elementType, status string) int {
    // XPath: count(//task[@status='completed'])
    xpath := fmt.Sprintf("//%s[@status='%s']", elementType, status)
    elements := xq.doc.FindElements(xpath)
    return len(elements)
}
```

#### Performance-Optimized Queries
```go
type QueryOptions struct {
    StatusFilter []string
    PhaseFilter  []string
    TypeFilter   []string
    Limit        int
    Since        *time.Time
}

// Optimized progress calculation using XPath
func (xq *XPathQuerier) CalculateProgress() ProgressMetrics {
    // Use XPath for efficient counting without loading full objects
    totalTasks := len(xq.doc.FindElements("//task"))
    completedTasks := len(xq.doc.FindElements("//task[@status='completed']"))
    
    totalTests := len(xq.doc.FindElements("//test"))
    passingTests := len(xq.doc.FindElements("//test[@status='passed']"))
    failingTests := len(xq.doc.FindElements("//test[@status='failing']"))
    
    totalPhases := len(xq.doc.FindElements("//phase"))
    completedPhases := len(xq.doc.FindElements("//phase[@status='completed']"))
    
    percentage := 0
    if totalTasks > 0 {
        percentage = (completedTasks * 100) / totalTasks
    }
    
    return ProgressMetrics{
        CompletedPhases:      completedPhases,
        TotalPhases:         totalPhases,
        PassingTests:        passingTests,
        FailingTests:        failingTests,
        CompletionPercentage: percentage,
    }
}

// Complex XPath queries for advanced filtering
func (xq *XPathQuerier) GetTasksInPhase(phaseID string) ([]*etree.Element, error) {
    // XPath: //task[@phase_id='2A']
    xpath := fmt.Sprintf("//task[@phase_id='%s']", phaseID)
    return xq.doc.FindElements(xpath), nil
}

func (xq *XPathQuerier) GetEventsByType(eventType string, limit int) ([]*etree.Element, error) {
    // XPath: //event[@type='blocker']
    xpath := fmt.Sprintf("//event[@type='%s']", eventType)
    events := xq.doc.FindElements(xpath)
    
    // Sort by timestamp and apply limit
    sort.Slice(events, func(i, j int) bool {
        timeI := events[i].SelectAttrValue("timestamp", "")
        timeJ := events[j].SelectAttrValue("timestamp", "")
        return timeI > timeJ
    })
    
    if len(events) > limit {
        return events[:limit], nil
    }
    
    return events, nil
}
```

#### Memory-Efficient XPath Parsing
```go
// For large epics, use XPath queries to avoid loading full Go structs
type StreamingQuerier struct {
    storage Storage
}

func (sq *StreamingQuerier) GetFailingTestsOnly(epicPath string) ([]FailingTest, error) {
    xq, err := NewXPathQuerier(epicPath)
    if err != nil {
        return nil, err
    }
    
    // Use XPath to get only failing tests
    failingElements := xq.doc.FindElements("//test[@status='failing']")
    
    var results []FailingTest
    for _, elem := range failingElements {
        test := FailingTest{
            ID:          elem.SelectAttrValue("id", ""),
            PhaseID:     elem.SelectAttrValue("phase_id", ""),
            Description: elem.FindElement("description").Text(),
            FailureNote: elem.FindElement("failure_note").Text(),
        }
        
        if failedAtStr := elem.SelectAttrValue("failed_at", ""); failedAtStr != "" {
            if failedAt, err := time.Parse(time.RFC3339, failedAtStr); err == nil {
                test.FailedAt = failedAt
            }
        }
        
        results = append(results, test)
    }
    
    return results, nil
}

// XPath-based next action recommendation
func (xq *XPathQuerier) GetNextActionRecommendation() string {
    // Priority 1: Check for failing tests using XPath
    failingTests := xq.doc.FindElements("//test[@status='failing']")
    if len(failingTests) > 0 {
        return fmt.Sprintf("Fix %d failing test(s)", len(failingTests))
    }
    
    // Priority 2: Check for active task
    if activeTask := xq.doc.FindElement("//task[@status='in_progress']"); activeTask != nil {
        description := activeTask.Text()
        return fmt.Sprintf("Continue task: %s", description)
    }
    
    // Priority 3: Check for pending tasks in active phase
    if activePhase := xq.doc.FindElement("//phase[@status='in_progress']"); activePhase != nil {
        phaseID := activePhase.SelectAttrValue("id", "")
        pendingTasks := xq.doc.FindElements(fmt.Sprintf("//task[@phase_id='%s' and @status='pending']", phaseID))
        if len(pendingTasks) > 0 {
            description := pendingTasks[0].Text()
            return fmt.Sprintf("Start task: %s", description)
        }
    }
    
    // Priority 4: Check for next pending phase
    pendingPhases := xq.doc.FindElements("//phase[@status='pending']")
    if len(pendingPhases) > 0 {
        name := pendingPhases[0].SelectAttrValue("name", "")
        return fmt.Sprintf("Start phase: %s", name)
    }
    
    return "Epic ready for completion"
}
```

---

## Implementation Phases

### Phase 2A: XPath Query Engine Foundation (1 day)
**Deliverables:**
- XPathQuerier implementation using etree's XPath capabilities
- Progress calculation engine using efficient XPath queries
- Basic query result structures
- Memory-efficient XPath filtering logic

**Tasks:**
- 2A_1: Implement XPathQuerier with etree FindElements()/FindElement()
- 2A_2: Create progress calculation using XPath counting queries
- 2A_3: Build efficient XPath filtering and search methods
- 2A_4: Implement next action recommendation engine with XPath queries

**Tests:**
- XPath query accuracy with various epic structures
- Progress calculation accuracy using XPath counting
- XPath filtering logic validation
- Memory usage validation for large epics with XPath queries

### Phase 2B: Status and Current Commands (1 day)
**Deliverables:**
- `agentpm status` command implementation
- `agentpm current` command implementation
- XML output formatting for status queries
- File override functionality with `-f` flag

**Tasks:**
- 2B_1: Build `agentpm status` command with progress display
- 2B_2: Build `agentpm current` command with context awareness
- 2B_3: Implement file override functionality
- 2B_4: Add comprehensive error handling

**Tests:**
- Status command output validation
- Current command logic verification
- File override functionality
- Error handling for missing/invalid files

### Phase 2C: Work Queue Commands (1 day)
**Deliverables:**
- `agentpm pending` command implementation
- `agentpm failing` command implementation
- Work prioritization and grouping logic
- Empty state handling

**Tasks:**
- 2C_1: Build `agentpm pending` with work organization
- 2C_2: Build `agentpm failing` with failure details
- 2C_3: Implement work prioritization logic
- 2C_4: Add empty state messaging

**Tests:**
- Pending work organization validation
- Failing test detection accuracy
- Prioritization logic verification
- Empty state handling

### Phase 2D: Events and Activity Commands (1 day)
**Deliverables:**
- `agentpm events` command implementation
- Event filtering and limiting functionality
- Chronological ordering logic
- Event metadata handling

**Tasks:**
- 2D_1: Build `agentpm events` with limit support
- 2D_2: Implement event filtering by type and timeframe
- 2D_3: Add chronological ordering and pagination
- 2D_4: Handle event metadata display

**Tests:**
- Event ordering and limiting
- Event filtering logic
- Metadata handling
- Performance with large event histories

---

## Testing Strategy

### Test Categories

#### Unit Tests (75% coverage target)
- **Focus:** Query logic, calculations, filtering
- **Examples:** Progress calculation, next action logic, event filtering
- **Execution:** In-memory operations with test data
- **Isolation:** Mock epic structures with known states

#### Component Tests (20% coverage target)
- **Focus:** Command implementations with file I/O
- **Examples:** Status command with real epic files, error handling
- **Execution:** Isolated temporary directories with test epics
- **Isolation:** Each test uses `t.TempDir()` with sample data

#### Integration Tests (5% coverage target)
- **Focus:** End-to-end command workflows
- **Examples:** Complete query workflows, file override scenarios
- **Execution:** Full CLI simulation with realistic data
- **Isolation:** Complete test environments

### Test Data Scenarios

#### Epic State Variations
```go
// Test data factory for various epic states
func NewTestEpicWithProgress() *Epic {
    return testutil.NewTestEpic("progress").
        WithStatus("in_progress").
        WithPhases(
            testutil.NewPhase("1A", "completed"),
            testutil.NewPhase("2A", "in_progress"),
            testutil.NewPhase("3A", "pending"),
        ).
        WithTasks(
            testutil.NewTask("1A_1", "1A", "completed"),
            testutil.NewTask("2A_1", "2A", "in_progress"),
            testutil.NewTask("2A_2", "2A", "pending"),
        ).
        WithTests(
            testutil.NewTest("1A_1", "1A_1", "passed"),
            testutil.NewTest("2A_1", "2A_1", "failing").
                WithFailureNote("Mobile responsive issue"),
        ).
        Build()
}
```

#### Edge Case Scenarios
- Empty epic (no phases, tasks, tests)
- Completed epic (all work done)
- Epic with only failing tests
- Epic with large event history
- Epic with circular dependencies (error case)

### Performance Test Scenarios
```go
func BenchmarkXPathStatusQuery(b *testing.B) {
    // Create large test epic XML file with 1000 tasks
    testEpicPath := testutil.CreateLargeTestEpic(b, 1000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        xq, err := NewXPathQuerier(testEpicPath)
        if err != nil {
            b.Fatal(err)
        }
        
        // Benchmark XPath-based progress calculation
        _ = xq.CalculateProgress()
    }
}

func BenchmarkXPathVsStructQuery(b *testing.B) {
    testEpicPath := testutil.CreateLargeTestEpic(b, 1000)
    
    b.Run("XPath", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            xq, _ := NewXPathQuerier(testEpicPath)
            _ = len(xq.doc.FindElements("//task[@status='pending']"))
        }
    })
    
    b.Run("FullParse", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            epic, _ := LoadFullEpic(testEpicPath)
            count := 0
            for _, task := range epic.Tasks {
                if task.Status == "pending" {
                    count++
                }
            }
        }
    })
}
```

---

## Quality Gates

### Performance Requirements
- **Query Response Time:** < 50ms for typical epics
- **Large Epic Handling:** < 200ms for epics with 1000+ tasks
- **Memory Usage:** < 100MB for complex queries
- **Event Filtering:** < 100ms for 10,000 events with filters

### Accuracy Requirements
- **Progress Calculation:** 100% accuracy across all scenarios
- **Next Action Logic:** Correct recommendations in all states
- **Event Ordering:** Chronological accuracy maintained
- **Filtering Logic:** No false positives/negatives

### Usability Requirements
- **Clear Output:** XML structure easily parseable by agents
- **Error Messages:** Actionable guidance for all error conditions
- **Empty States:** Helpful messages when no data available
- **Consistency:** Uniform output format across all commands

---

## Risk Assessment & Mitigation

### Technical Risks

#### Risk: Query Performance Degradation
- **Impact:** High - Slow queries block agent workflows
- **Probability:** Medium - Large epics with many events
- **Mitigation:** Performance benchmarks, streaming parsing, query optimization

#### Risk: Memory Usage Growth
- **Impact:** Medium - Could affect system resources
- **Probability:** Medium - Large epic files and event histories
- **Mitigation:** Memory profiling, lazy loading, efficient data structures

#### Risk: XML Output Parsing Issues
- **Impact:** Medium - Could break agent integrations
- **Probability:** Low - Well-defined schema
- **Mitigation:** Schema validation, compatibility testing, versioning

### Data Integrity Risks

#### Risk: Progress Calculation Errors
- **Impact:** High - Incorrect progress misleads agents
- **Probability:** Low - Comprehensive testing
- **Mitigation:** Extensive test coverage, calculation validation, edge case testing

#### Risk: Event Ordering Corruption
- **Impact:** Medium - Confusing timeline for agents
- **Probability:** Low - Simple timestamp ordering
- **Mitigation:** Timestamp validation, ordering tests, data integrity checks

---

## Integration Points

### Epic 1 Dependencies
- **Storage Interface:** Uses established storage abstraction
- **Configuration:** Leverages current epic resolution
- **Epic Structures:** Extends foundation epic data structures
- **Error Handling:** Uses standardized error patterns

### Future Epic Preparation
- **Epic 3:** Query commands provide state for lifecycle operations
- **Epic 4:** Work queries support task management workflows
- **Epic 5:** Event queries foundation for logging operations
- **Epic 6:** Status queries provide data for handoff reports

---

## Success Metrics

### Functional Metrics
- ✅ All acceptance criteria implemented and tested
- ✅ Query commands provide accurate, actionable information
- ✅ Performance targets met for all command scenarios
- ✅ Error handling covers all edge cases

### Quality Metrics
- ✅ Test coverage ≥ 90% for query logic
- ✅ Zero critical bugs in progress calculations
- ✅ Performance benchmarks established
- ✅ Agent usability validated

### User Experience Metrics
- ✅ Clear, structured XML output for all commands
- ✅ Consistent command behavior and patterns
- ✅ Helpful error messages and guidance
- ✅ Efficient workflows for common agent tasks

---

## Future Considerations

### Extensibility Points
- **Query Filters:** Additional filtering criteria
- **Output Formats:** Alternative output formats (JSON, plain text)
- **Caching:** Query result caching for performance
- **Subscriptions:** Real-time query updates

### Known Limitations
1. **Memory Usage:** Full epic loading for complex queries
2. **Query Language:** No advanced query syntax
3. **Sorting Options:** Limited sorting capabilities
4. **Batch Queries:** No support for multiple epic queries

---

## Appendices

### Appendix A: Command Reference

#### agentpm status
```bash
agentpm status              # Status of current epic
agentpm status -f epic.xml  # Status of specific epic

# Shows epic progress, completion percentage, active work
# Performance: < 50ms for typical epics
```

#### agentpm current
```bash
agentpm current             # Current work context
agentpm current -f epic.xml # Context for specific epic

# Shows active work and next action recommendations
# Includes failing test count and priorities
```

#### agentpm pending
```bash
agentpm pending             # All pending work
agentpm pending -f epic.xml # Pending work for specific epic

# Lists pending phases, tasks, and tests
# Organized for easy agent consumption
```

#### agentpm failing
```bash
agentpm failing             # All failing tests
agentpm failing -f epic.xml # Failing tests for specific epic

# Shows failure details and actionable information
# Empty when no tests are failing
```

#### agentpm events
```bash
agentpm events                    # Recent events (default limit)
agentpm events --limit=10         # Specific number of events
agentpm events -f epic.xml        # Events for specific epic

# Chronological order (newest first)
# Includes event types and metadata
```

### Appendix B: XML Schema Reference

#### Status Output Schema
```xml
<status epic="[epic_id]">
    <name>[epic_name]</name>
    <status>[epic_status]</status>
    <progress>
        <completed_phases>[number]</completed_phases>
        <total_phases>[number]</total_phases>
        <passing_tests>[number]</passing_tests>
        <failing_tests>[number]</failing_tests>
        <completion_percentage>[0-100]</completion_percentage>
    </progress>
    <current_phase>[phase_id]</current_phase>
    <current_task>[task_id]</current_task>
</status>
```

#### Events Output Schema
```xml
<events epic="[epic_id]" limit="[number]">
    <event timestamp="[ISO8601]" agent="[agent_name]" phase_id="[phase_id]" type="[event_type]">
        [event_message]
        
        [optional_event_details]
    </event>
</events>
```

### Appendix C: Performance Benchmarks

| Command | Epic Size | Target Time | Memory Usage |
|---------|-----------|-------------|--------------|
| status | 100 tasks | < 50ms | < 50MB |
| current | 100 tasks | < 30ms | < 30MB |
| pending | 100 tasks | < 40ms | < 40MB |
| failing | 100 tasks | < 30ms | < 30MB |
| events | 1000 events | < 100ms | < 80MB |

---

**Document Version:** 1.0  
**Last Updated:** 2025-08-16  
**Dependencies:** Epic 1 (Foundation & Configuration)  
**Next Epic:** Epic 3 (Epic Lifecycle Management)  
**Owner:** Development Team  
**Stakeholders:** Agent PM Users, Integration Partners