# EPIC-6 SPECIFICATION: Handoff & Documentation

## Overview

**Epic ID:** 6  
**Name:** Handoff & Documentation  
**Duration:** 2-3 days  
**Status:** pending  
**Priority:** medium  
**Depends On:** Epic 1 (Foundation), Epic 2 (Query Commands), Epic 3 (Epic Lifecycle), Epic 4 (Task & Phase Management), Epic 5 (Test Management & Event Logging)

**Goal:** Implement comprehensive agent handoff support and human-readable documentation generation, enabling seamless transitions between agents and providing clear project status communication to stakeholders.

## Business Context

Epic 6 represents the culmination of the AgentPM system by providing essential handoff capabilities for agent-to-agent transitions and comprehensive documentation for human stakeholders. This epic leverages all the data collection and tracking from previous epics to create meaningful reports that facilitate effective handoffs and project communication. The focus is on extracting maximum value from the rich data already being captured throughout the development workflow.

## User Stories

### Primary User Stories
- **As an outgoing agent, I can generate comprehensive handoff reports** so that incoming agents have complete context about current state and progress
- **As an incoming agent, I can quickly understand current state and blockers** so that I can continue work efficiently without knowledge gaps
- **As a human stakeholder, I can generate readable documentation from epic data** so that I can understand project progress and status
- **As an agent, I can identify recent activity and context quickly** so that I can understand what has been accomplished recently

### Secondary User Stories
- **As an agent, I can extract blockers from epic data** so that I can understand what obstacles need to be addressed
- **As a human, I can view project timelines and milestones** so that I can track progress over time
- **As an agent, I can get prioritized information** so that the most important details are highlighted first
- **As a stakeholder, I can understand epic structure** so that I can see how work is organized

## Technical Requirements

### Core Dependencies
- **Foundation:** Epic 1 CLI framework, XML processing, storage interface
- **Query System:** Epic 2 query service for data extraction and analysis
- **Lifecycle Data:** Epic 3 epic lifecycle events and timestamps
- **Work Progress:** Epic 4 task and phase management data and progress tracking
- **Rich Context:** Epic 5 test results, event logging, and blocker information

### Architecture Principles
- **Data Extraction:** Extract data directly from epic XML without interpretation
- **Simple Prioritization:** Order by basic rules (time, status) not "intelligence"
- **Multiple Formats:** XML for agent consumption, Markdown for human readability
- **Direct Representation:** Present data as-is without added context or analysis
- **Static Generation:** All reports are generated on-demand, no persistent state

### Report Generation Strategy
```
HANDOFF REPORT SECTIONS:
1. Epic Info (name, status, timestamps from XML)
2. Current State (active_phase, active_task from XML)
3. Progress Summary (simple counts and percentages)
4. Blockers (failed tests + blocker events)
5. Recent Events (last N events, chronological)

DOCUMENTATION SECTIONS:
1. Epic Header (name, status, dates)
2. Progress Stats (phase/task/test counts)
3. Phase List (with status icons)
4. Current State (active work)
5. Failed Tests (with failure notes)
6. Recent Events (chronological list)
```

## Functional Requirements

### FR-1: Comprehensive Handoff Report
**Command:** `agentpm handoff [--limit=N] [--time <timestamp>] [--file <epic-file>]`

**Behavior:**
- Generates complete XML handoff report with all relevant context
- Includes current state, progress summary, recent events, and blockers
- Prioritizes information by importance (blockers first, then current work)
- Supports configurable event limit (default: 10, max: 50)
- Includes timestamp for handoff generation
- Provides actionable next steps and immediate priorities

**Output Format:**
```xml
<handoff epic="8" timestamp="2025-08-16T15:30:00Z">
    <epic_info>
        <name>Schools Index Pagination</name>
        <status>wip</status>
        <started>2025-08-15T09:00:00Z</started>
        <assignee>agent_claude</assignee>
    </epic_info>
    <current_state>
        <active_phase>2A</active_phase>
        <active_task>2A_1</active_task>
        <next_action>Fix mobile responsive pagination controls</next_action>
        <immediate_priority>Address failing mobile responsive test</immediate_priority>
    </current_state>
    <summary>
        <completed_phases>2</completed_phases>
        <total_phases>4</total_phases>
        <passing_tests>12</passing_tests>
        <failing_tests>1</failing_tests>
        <completion_percentage>50</completion_percentage>
    </summary>
    <blockers>
        <blocker type="failing_test">
            <test_id>2A_2</test_id>
            <description>Mobile responsive pagination controls</description>
            <failure_note>Touch targets too small, need 44px+ minimum</failure_note>
        </blocker>
        <blocker type="event">
            <event_timestamp>2025-08-16T15:00:00Z</event_timestamp>
            <event_message>Need design system tokens for mobile responsive design</event_message>
        </blocker>
    </blockers>
    <recent_events limit="10">
        <event timestamp="2025-08-16T15:00:00Z" type="blocker">
            Found design system dependency
        </event>
        <event timestamp="2025-08-16T14:45:00Z" type="test_failed">
            Mobile responsive test failing
        </event>
        <event timestamp="2025-08-16T14:30:00Z" type="implementation">
            Implemented basic pagination controls
        </event>
    </recent_events>
</handoff>
```

### FR-2: Human-Readable Documentation
**Command:** `agentpm docs [--format=markdown] [--output=<filename>] [--time <timestamp>] [--file <epic-file>]`

**Behavior:**
- Generates simple markdown documentation directly from epic XML data
- Includes basic epic information, progress statistics, and current state
- Lists phases, tasks, tests, and events in structured format
- Supports custom output filename (default: epic-{id}-status.md)
- No interpretive text - only direct data representation
- Simple progress indicators using basic symbols

**Output Format (Markdown):**
```markdown
# Epic 8: Schools Index Pagination

**Status:** wip  
**Started:** 2025-08-15T09:00:00Z  
**Generated:** 2025-08-16T15:30:00Z

## Progress

**Phases:** 2 of 4 completed (50%)  
**Tasks:** 8 of 16 completed (50%)  
**Tests:** 12 of 13 passed (92%)

## Phases

- ✅ Phase 1A: Foundation Setup (done)
- ✅ Phase 1B: Basic Components (done)  
- 🔄 Phase 2A: Create PaginationComponent (wip)
- ⏳ Phase 2B: LiveView Integration (pending)

## Current State

**Active Phase:** 2A  
**Active Task:** 2A_1

## Failing Tests

- **2A_2:** Mobile responsive pagination controls
  - **Failed:** 2025-08-16T14:50:00Z
  - **Reason:** Touch targets too small, need 44px+ minimum

## Recent Events (Last 5)

1. **2025-08-16T15:00:00Z** [blocker] Found design system dependency
2. **2025-08-16T14:45:00Z** [test_failed] Mobile responsive test failing  
3. **2025-08-16T14:30:00Z** [implementation] Implemented basic pagination controls
4. **2025-08-16T14:15:00Z** [task_started] Started Task 2A_1
5. **2025-08-16T14:00:00Z** [phase_started] Started Phase 2A

## Blocker Events

- **2025-08-16T15:00:00Z:** Need design system tokens for mobile responsive design

---
*Generated by AgentPM*
```

**XML Output Confirmation:**
```xml
<docs_generated epic="8">
    <output_file>epic-8-status.md</output_file>
    <generated_at>2025-08-16T15:30:00Z</generated_at>
    <sections>
        <section>Epic Overview</section>
        <section>Progress Summary</section>
        <section>Current Status</section>
        <section>Issues & Blockers</section>
        <section>Recent Activity</section>
        <section>Next Steps</section>
    </sections>
    <message>Documentation generated successfully</message>
</docs_generated>
```

### FR-3: Blocker Identification & Analysis
**Internal Functionality:** Used by handoff and docs commands

**Blocker Sources:**
1. **Failed Tests:** Tests with status "failed" and failure details
2. **Blocker Events:** Events with type "blocker" from Epic 5 logging
3. **Incomplete Dependencies:** Phases that cannot start due to dependencies
4. **Stalled Work:** Tasks/phases that have been in-progress for extended periods

**Blocker Prioritization:**
- **Critical:** Failed tests, explicit blocker events
- **High:** Dependencies blocking current work
- **Medium:** General issues that may impact future work
- **Low:** Historical issues that have been resolved

**Blocker Analysis:**
- Extract blocker description and impact assessment
- Identify relationships between blockers and current work
- Provide suggested actions when possible
- Group related blockers for clarity

### FR-4: Recent Events Summarization
**Internal Functionality:** Used by handoff command

**Event Prioritization Logic:**
1. **High Priority:** blocker, test_failed, phase_completed
2. **Medium Priority:** implementation, task_completed, milestone
3. **Low Priority:** test_started, file_change

**Event Summarization:**
- Most recent events first (reverse chronological)
- Include event type, timestamp, and brief description
- Highlight high-priority events for immediate attention
- Provide context about phase/task when relevant
- Filter out routine events when limit is reached

**Event Context Enhancement:**
- Group related events (e.g., test_started followed by test_failed)
- Provide phase/task context for better understanding
- Include file change information when relevant
- Highlight patterns (e.g., multiple test failures in same area)

### FR-5: Progress Analysis & Statistics
**Internal Functionality:** Basic progress statistics for reports

**Progress Metrics:**
- Phase completion percentage (completed phases / total phases)
- Task completion percentage (completed tasks / total tasks)
- Test success rate (passed tests / total tests)
- Basic counts and ratios from epic data
- Simple time calculations (duration since start)

**No Recommendations:**
- Reports contain only factual data from epic XML
- No interpretive analysis or suggestions
- No "smart" recommendations or AI-generated content
- Focus on data presentation, not interpretation

## Non-Functional Requirements

### NFR-1: Performance
- Handoff generation executes in < 300ms for typical epic files
- Documentation generation executes in < 500ms including file writing
- Report generation scales well with large event histories
- Memory usage remains reasonable for complex epics

### NFR-2: Reliability
- All report generation is atomic and safe
- Handles incomplete or corrupted epic data gracefully
- Provides meaningful reports even with missing information
- File output operations are safe and don't overwrite without confirmation

### NFR-3: Usability
- Reports prioritize most important information first
- Markdown documentation is readable by non-technical stakeholders
- XML handoff data is structured for easy agent parsing
- Clear section organization and visual hierarchy

### NFR-4: Flexibility
- Configurable report depth and detail levels
- Multiple output formats (XML, Markdown)
- Custom output file naming and location
- Adaptable to different epic structures and states

## Data Processing Logic

### Handoff Report Generation
```go
type HandoffReport struct {
    EpicInfo       EpicInfo
    CurrentState   CurrentState
    Summary       ProgressSummary
    Blockers      []Blocker
    RecentEvents  []Event
}

func GenerateHandoffReport(epic *Epic, limit int) *HandoffReport {
    // 1. Extract basic epic info (name, status, timestamps)
    // 2. Get current state (active_phase, active_task)
    // 3. Calculate simple progress (counts and percentages)
    // 4. Find blockers (failed tests + blocker events)
    // 5. Get last N events (simple chronological)
}
```

### Documentation Generation
```go
type DocumentationData struct {
    EpicInfo      EpicInfo
    ProgressStats ProgressStats
    Phases       []PhaseInfo
    CurrentState CurrentState
    FailedTests  []TestInfo
    RecentEvents []EventInfo
}

func GenerateDocumentation(epic *Epic) *DocumentationData {
    // 1. Extract epic header info
    // 2. Calculate basic progress statistics
    // 3. List phases with status
    // 4. Get current active work
    // 5. List failed tests with failure notes
    // 6. List recent events chronologically
}
```

### Simple Blocker Detection
```go
func FindBlockers(epic *Epic) []Blocker {
    blockers := []Blocker{}
    
    // 1. Add all failed tests
    for _, test := range epic.Tests {
        if test.Status == "failed" {
            blockers = append(blockers, Blocker{
                Type: "failing_test",
                TestID: test.ID,
                FailureNote: test.FailureNote,
            })
        }
    }
    
    // 2. Add all blocker events
    for _, event := range epic.Events {
        if event.Type == "blocker" {
            blockers = append(blockers, Blocker{
                Type: "event",
                Timestamp: event.Timestamp,
                Message: event.Message,
            })
        }
    }
    
    return blockers
}
```

## Error Handling

### Error Categories
1. **Data Access Errors:** Missing epic files, corrupted XML data
2. **Generation Errors:** Failed report creation, file writing issues
3. **Analysis Errors:** Incomplete data affecting analysis quality
4. **Output Errors:** File permission issues, disk space problems

### Error Response Examples

**Missing Epic Data:**
```xml
<error>
    <type>incomplete_epic_data</type>
    <message>Cannot generate handoff: epic has no event history</message>
    <details>
        <missing_data>events</missing_data>
        <impact>Handoff report will be limited</impact>
        <suggestion>Recent activity section will be empty</suggestion>
    </details>
</error>
```

**File Writing Error:**
```xml
<error>
    <type>file_write_error</type>
    <message>Cannot write documentation to epic-8-status.md</message>
    <details>
        <file_path>epic-8-status.md</file_path>
        <error>Permission denied</error>
        <suggestion>Check file permissions or specify different output location</suggestion>
    </details>
</error>
```

### Graceful Degradation
- Generate partial reports when some data is missing
- Clearly indicate missing information sections
- Provide alternative recommendations when analysis is limited
- Continue processing when non-critical errors occur

## Acceptance Criteria

### AC-1: Comprehensive Handoff Report
- **GIVEN** I have an epic with active work, recent events, and some failing tests
- **WHEN** I run `agentpm handoff`
- **THEN** I should get XML with current state, progress summary, recent events, and blockers

### AC-2: Handoff Report for Completed Epic
- **GIVEN** I have a completed epic with all work done
- **WHEN** I run `agentpm handoff`
- **THEN** handoff should show completed status with 100% progress and no blockers

### AC-3: Human-Readable Documentation
- **GIVEN** I have an epic with phases, tasks, and progress
- **WHEN** I run `agentpm docs`
- **THEN** I should get human-readable markdown with epic overview and status

### AC-4: Documentation File Output
- **GIVEN** I specify a custom output filename
- **WHEN** I run `agentpm docs --output=my-report.md`
- **THEN** documentation should be written to my-report.md file

### AC-5: Blocker Identification
- **GIVEN** I have failing tests and logged blocker events
- **WHEN** I run `agentpm handoff`
- **THEN** handoff should list all blockers from failed tests and blocker events

### AC-6: Recent Events Limit
- **GIVEN** I have 20 events in my epic history
- **WHEN** I run `agentpm handoff --limit=5`
- **THEN** handoff should include only the 5 most recent events

### AC-7: Empty Epic Handling
- **GIVEN** I have a newly created epic with no progress
- **WHEN** I run `agentpm docs`
- **THEN** documentation should show epic structure but indicate no progress yet

### AC-8: Event Prioritization
- **GIVEN** I have recent events of various types
- **WHEN** recent events are included in handoff
- **THEN** blockers and failures should be prioritized over routine implementation events

## Testing Strategy

### Test Categories
- **Unit Tests (70%):** Report generation logic, blocker analysis, event summarization
- **Integration Tests (25%):** Full report generation, file operations, data aggregation
- **Output Format Tests (5%):** XML structure, Markdown formatting, file writing

### Test Data Requirements
- Epic files with rich event histories and various states
- Epic files with different completion levels (empty, partial, complete)
- Epic files with failing tests and blocker events
- Epic files with complex phase/task structures

### Test Isolation
- Each test uses isolated epic files in `t.TempDir()`
- File output testing in temporary directories
- Mock time for deterministic report generation
- No shared state between tests

## Implementation Phases

### Phase 6A: Handoff Report Foundation (Day 1)
- Create internal/reports package
- Implement ReportService with simple data extraction
- Handoff report generation with basic XML structure
- Simple blocker detection (failed tests + blocker events)
- Basic XML output formatting for handoff reports

### Phase 6B: Documentation Generation (Day 1-2)
- Simple markdown documentation generation
- Basic markdown templates (no complex formatting)
- File output operations with custom naming
- Simple progress statistics and counts
- Direct data extraction from epic XML

### Phase 6C: Data Processing & Events (Day 2)
- Recent events extraction (simple chronological ordering)
- Progress calculation (basic counts and percentages)
- Current state extraction (active_phase, active_task)
- Phase status mapping to simple icons
- Test status extraction with failure notes

### Phase 6D: Integration & Polish (Day 2-3)
- Integration testing with all previous epics
- Error handling and graceful degradation
- Performance optimization for large epics
- Output format refinement and consistency
- Documentation and help system completion

## Definition of Done

- [ ] All acceptance criteria verified with automated tests
- [ ] Handoff generation executes in < 300ms for typical epic files
- [ ] Documentation generation executes in < 500ms including file writing
- [ ] Test coverage > 85% for report generation logic
- [ ] All error cases handled gracefully with meaningful messages
- [ ] Blocker detection works for failed tests and blocker events
- [ ] Event extraction works chronologically without interpretation
- [ ] Markdown documentation is clean and well-formatted
- [ ] Integration tests verify end-to-end report workflows

## Dependencies and Risks

### Dependencies
- **Epic 1:** CLI framework, XML processing, storage interface
- **Epic 2:** Query service for data extraction and analysis
- **Epic 3:** Epic lifecycle events and status information
- **Epic 4:** Task and phase management data
- **Epic 5:** Test results, event logging, and rich context data

### Risks
- **Low Risk:** Report generation performance with very large event histories
- **Low Risk:** Markdown formatting complexity
- **Low Risk:** File writing permissions and error handling

### Mitigation Strategies
- Efficient data aggregation with minimal redundant processing
- Simple markdown templates with clear structure
- Comprehensive error handling for file operations
- Performance testing with large epic datasets

## Future Considerations

### Potential Enhancements (Not in Scope)
- Interactive report generation with filtering options
- Multiple output formats (HTML, PDF, JSON)
- Report templates for different stakeholder audiences
- Automated report scheduling and distribution
- Integration with external documentation systems

### Integration Points
- **Future Analytics:** Reports provide foundation for detailed project analytics
- **External Systems:** Report data could integrate with project management tools
- **Team Collaboration:** Handoff reports enable effective agent team workflows
- **Stakeholder Communication:** Documentation provides clear project visibility