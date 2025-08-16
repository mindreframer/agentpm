# EPIC-6: Handoff & Documentation Implementation Plan
## Test-Driven Development Approach

### Phase 1: Report Service Foundation + Tests (High Priority)

#### Phase 1A: Create Report Service Foundation
- [ ] Create internal/reports package
- [ ] Define ReportService struct with Storage injection
- [ ] Implement basic data extraction utilities from epic XML
- [ ] Simple progress calculation functions (counts and percentages)
- [ ] Current state extraction (active_phase, active_task)
- [ ] Basic blocker detection (failed tests + blocker events)
- [ ] Epic info extraction (name, status, timestamps)

#### Phase 1B: Write Report Service Foundation Tests **IMMEDIATELY AFTER 1A**
Epic 6 Test Scenarios Covered:
- [ ] **Test: Comprehensive handoff report** (Epic 6 line 415)
- [ ] **Test: Handoff report for completed epic** (Epic 6 line 420)
- [ ] **Test: Progress calculation accuracy**
- [ ] **Test: Current state extraction**
- [ ] **Test: Basic blocker detection from failed tests**
- [ ] **Test: Basic blocker detection from blocker events**
- [ ] **Test: Epic info extraction from XML**

#### Phase 1C: Handoff XML Report Generation
- [ ] Create cmd/handoff.go command
- [ ] Implement XML handoff report structure
- [ ] Epic info section generation
- [ ] Current state section generation
- [ ] Progress summary section generation
- [ ] Blockers section generation
- [ ] Recent events section generation (chronological)
- [ ] --limit flag support for event count

#### Phase 1D: Write Handoff Command Tests **IMMEDIATELY AFTER 1C**
Epic 6 Test Scenarios Covered:
- [ ] **Test: Comprehensive handoff report** (detailed XML structure)
- [ ] **Test: Handoff report for completed epic** (completion state)
- [ ] **Test: Recent events limit** (Epic 6 line 430)
- [ ] **Test: Blocker identification** (Epic 6 line 425)
- [ ] **Test: XML handoff output format**
- [ ] **Test: --limit flag functionality**
- [ ] **Test: Handoff with no recent activity**

### Phase 2: Documentation Generation + Tests (High Priority)

#### Phase 2A: Markdown Documentation Infrastructure
- [ ] Create simple markdown template system
- [ ] Basic data formatting utilities (timestamps, percentages)
- [ ] Status icon mapping (✅ 🔄 ⏳ for phase states)
- [ ] Simple progress statistics calculation
- [ ] Phase list generation with status indicators
- [ ] Test status extraction with failure notes
- [ ] File output operations with custom naming

#### Phase 2B: Write Documentation Infrastructure Tests **IMMEDIATELY AFTER 2A**
Epic 6 Test Scenarios Covered:
- [ ] **Test: Human-readable documentation** (Epic 6 line 422)
- [ ] **Test: Documentation file output** (Epic 6 line 426)
- [ ] **Test: Status icon mapping accuracy**
- [ ] **Test: Progress statistics calculation**
- [ ] **Test: Phase list generation**
- [ ] **Test: Test status extraction**
- [ ] **Test: File output with custom naming**

#### Phase 2C: Documentation Command Implementation
- [ ] Create cmd/docs.go command
- [ ] Integrate with markdown template system
- [ ] Epic header generation (name, status, dates)
- [ ] Progress section generation (phase/task/test counts)
- [ ] Current state section generation
- [ ] Failed tests section generation
- [ ] Recent events section generation
- [ ] --output flag support for custom filenames

#### Phase 2D: Write Documentation Command Tests **IMMEDIATELY AFTER 2C**
Epic 6 Test Scenarios Covered:
- [ ] **Test: Human-readable documentation** (complete workflow)
- [ ] **Test: Documentation file output** (file creation and content)
- [ ] **Test: Empty epic handling** (Epic 6 line 435)
- [ ] **Test: Custom output filename**
- [ ] **Test: Markdown format consistency**
- [ ] **Test: Documentation sections completeness**
- [ ] **Test: Documentation with various epic states**

### Phase 3: Data Processing & Event Handling + Tests (Medium Priority)

#### Phase 3A: Event Processing & Chronological Ordering
- [ ] Recent events extraction with chronological ordering
- [ ] Simple event filtering by count limit
- [ ] Event type and timestamp formatting
- [ ] Phase and task context extraction for events
- [ ] Event prioritization by basic rules (no AI interpretation)
- [ ] Simple event deduplication and grouping
- [ ] Event content formatting for both XML and Markdown

#### Phase 3B: Write Event Processing Tests **IMMEDIATELY AFTER 3A**
Epic 6 Test Scenarios Covered:
- [ ] **Test: Recent events limit** (event count management)
- [ ] **Test: Event prioritization** (Epic 6 line 440)
- [ ] **Test: Chronological event ordering**
- [ ] **Test: Event type formatting**
- [ ] **Test: Event context extraction**
- [ ] **Test: Event deduplication logic**
- [ ] **Test: Event content formatting**

#### Phase 3C: Enhanced Progress Statistics
- [ ] Phase completion percentage calculation
- [ ] Task completion percentage calculation  
- [ ] Test success rate calculation
- [ ] Simple duration calculation (epic start to current time)
- [ ] Basic work progress indicators
- [ ] Phase status distribution analysis
- [ ] Simple epic health metrics (no interpretation)

#### Phase 3D: Write Progress Statistics Tests **IMMEDIATELY AFTER 3C**
Epic 6 Test Scenarios Covered:
- [ ] **Test: Progress calculation accuracy** (various completion states)
- [ ] **Test: Phase completion percentage**
- [ ] **Test: Task completion percentage**
- [ ] **Test: Test success rate calculation**
- [ ] **Test: Duration calculation**
- [ ] **Test: Epic health metrics**
- [ ] **Test: Progress statistics with edge cases**

### Phase 4: Integration & Error Handling + Tests (Low Priority)

#### Phase 4A: Error Handling & Graceful Degradation
- [ ] Handle missing epic data gracefully
- [ ] Error handling for incomplete XML structures
- [ ] File writing error handling and recovery
- [ ] Missing event history handling
- [ ] Partial data report generation
- [ ] Clear error messages for file operations
- [ ] Graceful degradation when sections are empty

#### Phase 4B: Write Error Handling Tests **IMMEDIATELY AFTER 4A**
Epic 6 Test Scenarios Covered:
- [ ] **Test: Empty epic handling** (missing data scenarios)
- [ ] **Test: File writing error handling**
- [ ] **Test: Incomplete XML data handling**
- [ ] **Test: Missing event history**
- [ ] **Test: Partial data report generation**
- [ ] **Test: Clear error messages**
- [ ] **Test: Graceful degradation scenarios**

#### Phase 4C: Integration & Performance
- [ ] Integration testing with all previous epics
- [ ] Cross-command consistency (error formats, flag handling)
- [ ] Global flag consistency (--time, --file)
- [ ] Help system integration for report commands
- [ ] Performance optimization for large epic files
- [ ] Output format consistency and validation

#### Phase 4D: Write Integration Tests **IMMEDIATELY AFTER 4C**
Epic 6 Test Scenarios Covered:
- [ ] **Test: End-to-end handoff and documentation workflow**
- [ ] **Test: Integration with all previous epics**
- [ ] **Test: Cross-command consistency**
- [ ] **Test: Global flag handling**
- [ ] **Test: Help system completeness**
- [ ] **Test: Performance with large epic files**
- [ ] **Test: Output format consistency**

## Development Workflow Per Phase

For **EACH** phase:

1. **Implement Code** (Phase XA or XC)
2. **Write Tests IMMEDIATELY** (Phase XB or XD) 
3. **Run Tests & Verify** - All tests must pass
4. **Run Linting/Type Checking** - Code must be clean
5. **NEVER move to next phase with failing tests**

## Epic 6 Specific Considerations

### Dependencies & Integration
- **Epic 1:** CLI framework, XML processing, storage interface
- **Epic 2:** Query service integration for data extraction
- **Epic 3:** Epic lifecycle events and status information
- **Epic 4:** Task and phase management data
- **Epic 5:** Test results, event logging, and blocker information

### Technical Requirements
- **Simple Data Extraction:** No AI interpretation, just direct XML data processing
- **Basic Statistics:** Simple counts, percentages, and duration calculations
- **Template-Based Output:** Markdown generation using simple templates
- **Chronological Ordering:** Events ordered by timestamp without "intelligence"
- **Graceful Degradation:** Meaningful reports even with incomplete data

### File Structure
```
├── cmd/
│   ├── handoff.go          # XML handoff report generation
│   └── docs.go            # Markdown documentation generation
├── internal/
│   └── reports/           # Report generation service
│       ├── service.go     # ReportService with DI
│       ├── handoff.go     # Handoff report generation
│       ├── docs.go        # Documentation generation
│       ├── blockers.go    # Simple blocker detection
│       ├── events.go      # Event processing and ordering
│       ├── progress.go    # Progress statistics calculation
│       └── templates.go   # Markdown template system
└── testdata/
    ├── epic-complete.xml   # Epic with all work completed
    ├── epic-partial.xml    # Epic with mixed progress
    ├── epic-blockers.xml   # Epic with failing tests and blockers
    └── epic-empty.xml      # Epic with minimal data
```

### Simple Data Processing Approach
```go
// No AI interpretation - just direct data extraction
type ProgressStats struct {
    CompletedPhases int
    TotalPhases     int
    CompletedTasks  int
    TotalTasks      int
    PassedTests     int
    TotalTests      int
    CompletionPercentage int
}

func CalculateProgress(epic *Epic) ProgressStats {
    // Simple counting - no complex analysis
    stats := ProgressStats{}
    
    for _, phase := range epic.Phases {
        stats.TotalPhases++
        if phase.Status == "done" {
            stats.CompletedPhases++
        }
    }
    
    // Calculate percentage: (completed / total) * 100
    if stats.TotalPhases > 0 {
        stats.CompletionPercentage = (stats.CompletedPhases * 100) / stats.TotalPhases
    }
    
    return stats
}
```

### Markdown Template System
```go
const DocTemplate = `# Epic {{.ID}}: {{.Name}}

**Status:** {{.Status}}  
**Started:** {{.StartedAt}}  
**Generated:** {{.GeneratedAt}}

## Progress

**Phases:** {{.Progress.CompletedPhases}} of {{.Progress.TotalPhases}} completed ({{.Progress.CompletionPercentage}}%)  
**Tasks:** {{.Progress.CompletedTasks}} of {{.Progress.TotalTasks}} completed  
**Tests:** {{.Progress.PassedTests}} of {{.Progress.TotalTests}} passed

## Phases

{{range .Phases}}
- {{.StatusIcon}} {{.Name}} ({{.Status}})
{{end}}

## Recent Events (Last {{.EventLimit}})

{{range .RecentEvents}}
{{.Index}}. **{{.Timestamp}}** [{{.Type}}] {{.Message}}
{{end}}
`
```

## Benefits of This Approach

✅ **Immediate Feedback** - Catch issues as soon as code is written  
✅ **Working Code** - Each phase delivers tested, working functionality  
✅ **Epic 6 Coverage** - All acceptance criteria distributed across phases  
✅ **Incremental Progress** - Agents can use report commands after each phase  
✅ **Risk Mitigation** - Problems caught early, not at the end  
✅ **Quality Assurance** - No untested code makes it to later phases  

## Test Distribution Summary

- **Phase 1 Tests:** 14 scenarios (Report service foundation, handoff XML generation)
- **Phase 2 Tests:** 14 scenarios (Documentation infrastructure, markdown generation)
- **Phase 3 Tests:** 14 scenarios (Event processing, progress statistics)
- **Phase 4 Tests:** 14 scenarios (Error handling, integration testing)

**Total: All Epic 6 acceptance criteria and test scenarios covered across all phases**

---

## Implementation Status

### EPIC 6: HANDOFF & DOCUMENTATION - PENDING
### Current Status: READY TO START (After Epic 1-5 Complete)

### Progress Tracking
- [ ] Phase 1A: Create Report Service Foundation
- [ ] Phase 1B: Write Report Service Foundation Tests
- [ ] Phase 1C: Handoff XML Report Generation
- [ ] Phase 1D: Write Handoff Command Tests
- [ ] Phase 2A: Markdown Documentation Infrastructure
- [ ] Phase 2B: Write Documentation Infrastructure Tests
- [ ] Phase 2C: Documentation Command Implementation
- [ ] Phase 2D: Write Documentation Command Tests
- [ ] Phase 3A: Event Processing & Chronological Ordering
- [ ] Phase 3B: Write Event Processing Tests
- [ ] Phase 3C: Enhanced Progress Statistics
- [ ] Phase 3D: Write Progress Statistics Tests
- [ ] Phase 4A: Error Handling & Graceful Degradation
- [ ] Phase 4B: Write Error Handling Tests
- [ ] Phase 4C: Integration & Performance
- [ ] Phase 4D: Write Integration Tests

### Definition of Done
- [ ] All acceptance criteria verified with automated tests
- [ ] Handoff generation executes in < 300ms for typical epic files
- [ ] Documentation generation executes in < 500ms including file writing
- [ ] Test coverage > 85% for report generation logic
- [ ] All error cases handled gracefully with meaningful messages
- [ ] Blocker detection works for failed tests and blocker events
- [ ] Event extraction works chronologically without interpretation
- [ ] Markdown documentation is clean and well-formatted
- [ ] Integration tests verify end-to-end report workflows