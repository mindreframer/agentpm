# Epic 6: Handoff & Documentation - Specification

## Overview
**Goal:** Agent handoff support and human-readable documentation  
**Duration:** 2-3 days  
**Philosophy:** Comprehensive knowledge transfer with structured handoffs and clear documentation

## User Stories
1. Generate comprehensive handoff reports for outgoing agents
2. Quickly understand current state and blockers as incoming agent
3. Generate readable documentation from epic data for humans
4. Identify recent activity and context quickly for situational awareness
5. Create structured handoff data for automated agent onboarding

## Technical Requirements
- **Dependencies:** Epic 1-5 (all previous foundations)
- **Handoff Generation:** Comprehensive XML with all relevant context
- **Documentation:** Markdown generation from epic data
- **Recent Activity:** Configurable event summarization
- **Blocker Extraction:** Integration with Epic 5 blocker detection
- **Human Readability:** Clear formatting and structure for documentation

## Handoff System Architecture

### Handoff Data Structure
```go
type HandoffReport struct {
    EpicInfo      EpicSummary     `xml:"epic_info"`
    CurrentState  CurrentState    `xml:"current_state"`
    Progress      ProgressSummary `xml:"summary"`
    RecentEvents  []Event         `xml:"recent_events>event"`
    Blockers      []Blocker       `xml:"blockers>blocker"`
    NextActions   []NextAction    `xml:"next_actions>action"`
    Context       HandoffContext  `xml:"context"`
    GeneratedAt   time.Time       `xml:"generated_at,attr"`
}

type EpicSummary struct {
    ID          string    `xml:"id,attr"`
    Name        string    `xml:"name,attr"`
    Status      string    `xml:"status,attr"`
    Started     time.Time `xml:"started,attr,omitempty"`
    Assignee    string    `xml:"assignee,attr"`
    Duration    string    `xml:"duration,attr,omitempty"`
    Description string    `xml:"description,omitempty"`
}

type CurrentState struct {
    ActivePhase    string `xml:"active_phase,attr,omitempty"`
    ActiveTask     string `xml:"active_task,attr,omitempty"`
    NextAction     string `xml:"next_action,omitempty"`
    LastActivity   time.Time `xml:"last_activity,attr,omitempty"`
    WorkingSince   time.Time `xml:"working_since,attr,omitempty"`
}

type NextAction struct {
    Priority    string `xml:"priority,attr"`
    Type        string `xml:"type,attr"`
    Description string `xml:",chardata"`
}

type HandoffContext struct {
    TotalWorkDays    int    `xml:"total_work_days"`
    RecentFocus      string `xml:"recent_focus"`
    KeyDecisions     []string `xml:"key_decisions>decision"`
    ImportantNotes   []string `xml:"important_notes>note"`
    TechnicalContext string   `xml:"technical_context,omitempty"`
}
```

## Documentation Generation System

### Markdown Documentation Structure
```go
type DocumentationSections struct {
    EpicOverview    MarkdownSection
    PhaseProgress   MarkdownSection  
    TaskStatus      MarkdownSection
    TestResults     MarkdownSection
    RecentActivity  MarkdownSection
    Blockers        MarkdownSection
    NextSteps       MarkdownSection
    Timeline        MarkdownSection
}

type MarkdownSection struct {
    Title       string
    Content     string
    Subsections []MarkdownSection
}
```

## Implementation Phases

### Phase 6A: Handoff Data Aggregation (1 day)
- Comprehensive handoff data collection from all epic components
- Current state analysis with Epic 2/4 integration
- Recent activity summarization with configurable limits
- Blocker extraction with Epic 5 integration
- Next action recommendation generation
- Context analysis and key information extraction

### Phase 6B: Documentation Generation Engine (1 day)
- Markdown generation from epic data structures
- Human-readable formatting and structure
- Progress visualization and status reporting
- Timeline generation with milestone identification
- Cross-reference generation for phases, tasks, tests
- Template-based documentation with customization

### Phase 6C: Command Implementation & Integration (0.5 days)
- `agentpm handoff` command with comprehensive output
- `agentpm docs` command with markdown generation
- Output format options and customization
- Integration with all previous epic systems
- Command validation and error handling
- Help system and usage examples

## Acceptance Criteria
- ✅ `agentpm handoff` generates complete XML with current state, progress, recent events, and blockers
- ✅ `agentpm docs` creates markdown documentation suitable for humans
- ✅ Handoff includes active work, failing tests, and next actions
- ✅ Documentation shows epic overview, phase progress, and timeline
- ✅ Recent events are summarized with most important items first

## Handoff Generation Logic

### Comprehensive Data Collection
```go
func GenerateHandoff(epic *Epic, options HandoffOptions) (*HandoffReport, error) {
    report := &HandoffReport{
        GeneratedAt: time.Now(),
    }
    
    // Epic summary information
    report.EpicInfo = extractEpicSummary(epic)
    
    // Current state from Epic 2/4
    report.CurrentState = extractCurrentState(epic)
    
    // Progress summary from Epic 4
    report.Progress = calculateProgressSummary(epic)
    
    // Recent events from Epic 5 (configurable limit)
    report.RecentEvents = extractRecentEvents(epic, options.EventLimit)
    
    // Blockers from Epic 5
    report.Blockers = extractBlockers(epic)
    
    // Next actions from Epic 2/4
    report.NextActions = generateNextActions(epic)
    
    // Context analysis
    report.Context = analyzeContext(epic)
    
    return report, nil
}
```

### Recent Activity Summarization
```go
func ExtractRecentEvents(epic *Epic, limit int) []Event {
    events := epic.Events
    
    // Sort chronologically (most recent first)
    sort.Slice(events, func(i, j int) bool {
        return events[i].Timestamp.After(events[j].Timestamp)
    })
    
    // Prioritize important event types
    prioritizedEvents := prioritizeEvents(events)
    
    // Apply limit
    if limit > 0 && len(prioritizedEvents) > limit {
        prioritizedEvents = prioritizedEvents[:limit]
    }
    
    return prioritizedEvents
}

func PrioritizeEvents(events []Event) []Event {
    // Priority order: blockers, test failures, milestones, implementation
    priority := map[EventType]int{
        EventBlocker:     1,
        EventTestFailed: 2,
        EventMilestone:  3,
        EventDecision:   4,
        EventImplementation: 5,
        EventNote:       6,
    }
    
    sort.Slice(events, func(i, j int) bool {
        pi := priority[events[i].Type]
        pj := priority[events[j].Type]
        if pi != pj {
            return pi < pj
        }
        return events[i].Timestamp.After(events[j].Timestamp)
    })
    
    return events
}
```

## Documentation Generation

### Markdown Generation Engine
```go
func GenerateDocumentation(epic *Epic, options DocOptions) (string, error) {
    doc := &MarkdownDocument{}
    
    // Epic Overview
    doc.AddSection("Epic Overview", generateOverviewSection(epic))
    
    // Phase Progress
    doc.AddSection("Phase Progress", generatePhaseProgressSection(epic))
    
    // Task Status
    doc.AddSection("Task Status", generateTaskStatusSection(epic))
    
    // Test Results
    doc.AddSection("Test Results", generateTestResultsSection(epic))
    
    // Recent Activity
    doc.AddSection("Recent Activity", generateActivitySection(epic, options.ActivityLimit))
    
    // Blockers (if any)
    if blockers := extractBlockers(epic); len(blockers) > 0 {
        doc.AddSection("Current Blockers", generateBlockersSection(blockers))
    }
    
    // Next Steps
    doc.AddSection("Next Steps", generateNextStepsSection(epic))
    
    // Timeline
    if options.IncludeTimeline {
        doc.AddSection("Timeline", generateTimelineSection(epic))
    }
    
    return doc.Render(), nil
}
```

### Human-Readable Formatting
```go
func GenerateOverviewSection(epic *Epic) string {
    progress := calculateProgress(epic)
    duration := calculateDuration(epic)
    
    return fmt.Sprintf(`
**Epic:** %s  
**Status:** %s  
**Progress:** %d%% complete (%d/%d phases, %d/%d tasks)  
**Duration:** %s  
**Last Activity:** %s  

%s
`, 
        epic.Name,
        epic.Status,
        progress.OverallProgress,
        progress.CompletedPhases, progress.TotalPhases,
        progress.CompletedTasks, progress.TotalTasks,
        duration,
        formatLastActivity(epic),
        epic.Description,
    )
}
```

## Output Examples

### agentpm handoff
```xml
<handoff epic="8" timestamp="2025-08-16T15:30:00Z">
    <epic_info>
        <name>Schools Index Pagination</name>
        <status>in_progress</status>
        <started>2025-08-15T09:00:00Z</started>
        <assignee>agent_claude</assignee>
        <duration>1 day 6 hours</duration>
    </epic_info>
    <current_state>
        <active_phase>2A</active_phase>
        <active_task>2A_1</active_task>
        <next_action>Fix mobile responsive pagination controls</next_action>
        <last_activity>2025-08-16T15:00:00Z</last_activity>
        <working_since>2025-08-16T14:15:00Z</working_since>
    </current_state>
    <summary>
        <completed_phases>2</completed_phases>
        <total_phases>4</total_phases>
        <passing_tests>12</passing_tests>
        <failing_tests>1</failing_tests>
        <completion_percentage>50</completion_percentage>
    </summary>
    <recent_events limit="3">
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
    <blockers>
        <blocker type="failing_test" source="2A_2">
            Touch targets too small, need 44px+ minimum
        </blocker>
        <blocker type="logged_blocker" source="event">
            Need design system tokens for mobile responsive design
        </blocker>
    </blockers>
    <next_actions>
        <action priority="high" type="fix">Fix failing mobile responsive test</action>
        <action priority="medium" type="implement">Complete remaining Phase 2A tasks</action>
        <action priority="low" type="plan">Plan Phase 3A activities</action>
    </next_actions>
    <context>
        <total_work_days>2</total_work_days>
        <recent_focus>Mobile responsive pagination implementation</recent_focus>
        <key_decisions>
            <decision>Use standard pagination component pattern</decision>
            <decision>Implement accessibility from the start</decision>
        </key_decisions>
        <important_notes>
            <note>Design system dependency identified</note>
            <note>Mobile testing requires physical device</note>
        </important_notes>
        <technical_context>LiveView pagination with accessibility focus</technical_context>
    </context>
</handoff>
```

### agentpm docs (Sample Markdown Output)
```markdown
# Epic 8: Schools Index Pagination

**Status:** In Progress  
**Progress:** 50% complete (2/4 phases, 5/12 tasks)  
**Duration:** 1 day 6 hours  
**Last Activity:** 2 hours ago  

Implementation of pagination controls for the schools index page with accessibility and mobile responsive design.

## Phase Progress

### ✅ Phase 1A: Foundation Setup (Completed)
- **Duration:** 4 hours
- **Tasks:** 3/3 completed
- **Tests:** 5/5 passing

### ✅ Phase 2A: Create PaginationComponent (In Progress)
- **Duration:** 2 hours 30 minutes (ongoing)
- **Tasks:** 1/2 completed
- **Tests:** 2/3 passing (1 failing)

### ⏳ Phase 3A: LiveView Integration (Pending)
- **Status:** Not started
- **Dependencies:** Phase 2A completion

### ⏳ Phase 4A: Performance Optimization (Pending)
- **Status:** Not started
- **Dependencies:** Phase 3A completion

## Current Blockers

### 🚫 Failing Test: Mobile Responsive Controls
- **Test ID:** 2A_2
- **Issue:** Touch targets too small, need 44px+ minimum
- **Phase:** 2A
- **Since:** 30 minutes ago

### 🚫 Design System Dependency
- **Type:** External dependency
- **Issue:** Need design system tokens for mobile responsive design
- **Impact:** Blocking mobile implementation
- **Since:** 15 minutes ago

## Recent Activity

### 🎯 Implementation (15 minutes ago)
Implemented basic pagination controls in Phase 2A

**Files Changed:**
- `src/components/Pagination.js` (added)
- `src/styles/pagination.css` (added)

### ❌ Test Failed (30 minutes ago)
Mobile responsive test failing in Phase 2A

**Test:** 2A_2 - Mobile pagination controls  
**Issue:** Touch targets too small, need 44px+ minimum

### 🚫 Blocker Identified (45 minutes ago)
Found design system dependency

**Details:** Need design system tokens for mobile responsive design

## Next Steps

1. **High Priority:** Fix failing mobile responsive test (Test 2A_2)
2. **Medium Priority:** Complete remaining Phase 2A tasks
3. **Low Priority:** Plan Phase 3A LiveView integration activities

## Timeline

- **Started:** August 15, 2025 at 9:00 AM
- **Phase 1A Completed:** August 15, 2025 at 1:00 PM
- **Phase 2A Started:** August 16, 2025 at 2:00 PM
- **Current Time:** August 16, 2025 at 3:30 PM

---

*Documentation generated on August 16, 2025 at 3:30 PM*
```

## Context Analysis & Intelligence

### Key Decision Extraction
```go
func ExtractKeyDecisions(epic *Epic) []string {
    decisions := []string{}
    
    // Look for decision-type events
    for _, event := range epic.Events {
        if event.Type == EventDecision {
            decisions = append(decisions, event.Message)
        }
    }
    
    // Look for milestone events that indicate decisions
    for _, event := range epic.Events {
        if event.Type == EventMilestone && containsDecisionKeywords(event.Message) {
            decisions = append(decisions, event.Message)
        }
    }
    
    return decisions
}
```

### Technical Context Analysis
```go
func AnalyzeTechnicalContext(epic *Epic) string {
    context := []string{}
    
    // Analyze file changes for technology patterns
    technologies := extractTechnologies(epic)
    if len(technologies) > 0 {
        context = append(context, fmt.Sprintf("Technologies: %s", strings.Join(technologies, ", ")))
    }
    
    // Analyze event patterns for approach
    approach := extractApproach(epic)
    if approach != "" {
        context = append(context, fmt.Sprintf("Approach: %s", approach))
    }
    
    return strings.Join(context, "; ")
}
```

## Test Scenarios (Key Examples)
- **Handoff Generation:** Create comprehensive handoff with current state, progress, events, and blockers
- **Documentation:** Generate human-readable markdown with overview, progress, and timeline
- **Recent Activity:** Summarize recent events with intelligent prioritization
- **Blocker Extraction:** Identify and categorize all blocking issues
- **Context Analysis:** Extract key decisions and technical context
- **Next Actions:** Generate intelligent next step recommendations
- **Empty States:** Handle epics with no activity, progress, or events

## Integration with Previous Epics

### Epic 1 Integration
- **Epic Loading:** Use Epic 1 epic loading for handoff data access
- **Configuration:** Use Epic 1 config for project context

### Epic 2 Integration
- **Status Analysis:** Use Epic 2 status calculation for progress summary
- **Current State:** Leverage Epic 2 current state analysis

### Epic 3 Integration
- **Lifecycle Events:** Include Epic 3 lifecycle events in timeline
- **Duration Calculation:** Use Epic 3 timestamps for duration analysis

### Epic 4 Integration
- **Phase Progress:** Use Epic 4 phase/task progress for detailed status
- **Next Actions:** Leverage Epic 4 auto-next logic for recommendations

### Epic 5 Integration
- **Event Timeline:** Use Epic 5 rich events for activity history
- **Blocker Detection:** Leverage Epic 5 blocker identification
- **Test Status:** Include Epic 5 test results in documentation

## Quality Gates
- [ ] All acceptance criteria implemented and tested
- [ ] Comprehensive handoff data with all relevant context
- [ ] Human-readable documentation with clear structure
- [ ] Intelligent recent activity summarization
- [ ] Effective blocker identification and reporting

## Performance Considerations
- **Handoff Generation:** Fast data aggregation from all epic components
- **Documentation:** Efficient markdown generation with templates
- **Event Summarization:** Optimized recent activity analysis
- **Context Analysis:** Efficient pattern recognition and extraction

This specification provides comprehensive agent handoff and human-readable documentation while leveraging all previous epic foundations for complete context and intelligence.