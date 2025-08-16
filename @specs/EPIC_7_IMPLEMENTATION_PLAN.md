# EPIC-7: Missing Features Implementation Plan
## Test-Driven Development Approach

### Phase 1: Log Command Foundation + Tests (High Priority)

#### Phase 1A: Create Log Command Foundation
- [ ] Create cmd/log.go command with urfave/cli/v3 framework
- [ ] Define LogService struct with Storage and Query service injection
- [ ] Implement --files flag parsing for file tracking (format: "path:action")
- [ ] Implement --type flag with validation (implementation, blocker, issue, etc.)
- [ ] Create event creation utilities for log entries
- [ ] Timestamp handling utilities (current time vs --time flag)
- [ ] XML event appending logic with atomic file updates

#### Phase 1B: Write Log Command Tests **IMMEDIATELY AFTER 1A**
Epic 7 Test Scenarios Covered:
- [ ] **Test: Log with default type "implementation"**
- [ ] **Test: Log with --type=blocker flag**
- [ ] **Test: Log with --type=issue flag**
- [ ] **Test: Log with --files flag parsing**
- [ ] **Test: Log with multiple files in --files**
- [ ] **Test: Log message appending to XML events**
- [ ] **Test: Timestamp injection via --time flag**
- [ ] **Test: Error handling for invalid --type values**

#### Phase 1C: File Tracking & Event Integration
- [ ] Implement file action parsing (added, modified, deleted, renamed)
- [ ] File path validation and normalization
- [ ] Integration with existing event logging system
- [ ] XML formatting for file metadata in events
- [ ] Event type validation and standardization
- [ ] Error handling for malformed file specifications

#### Phase 1D: Write File Tracking Tests **IMMEDIATELY AFTER 1C**
Epic 7 Test Scenarios Covered:
- [ ] **Test: File action parsing accuracy**
- [ ] **Test: Multiple file tracking in single command**
- [ ] **Test: File path validation and error handling**
- [ ] **Test: XML event format compliance**
- [ ] **Test: Integration with existing event system**
- [ ] **Test: Error messages for malformed file specs**

### Phase 2
    - REMOVED, not needed!

### Phase 3: Enhanced XML Structure Support + Tests (Medium Priority)

#### Phase 3A: Metadata Section Implementation
- [ ] Extend internal/epic/epic.go structs for metadata section
- [ ] Add XML marshaling/unmarshaling tags for metadata
- [ ] Implement metadata parsing and validation
- [ ] Add metadata fields: created, assignee, estimated_effort
- [ ] Update epic initialization to include metadata
- [ ] Ensure backward compatibility with existing epic files

#### Phase 3B: Write Metadata Section Tests **IMMEDIATELY AFTER 3A**
Epic 7 Test Scenarios Covered:
- [ ] **Test: Metadata section parsing from XML**
- [ ] **Test: Metadata marshaling to XML**
- [ ] **Test: Created timestamp handling**
- [ ] **Test: Assignee field validation**
- [ ] **Test: Estimated effort format validation**
- [ ] **Test: Backward compatibility with epics without metadata**
- [ ] **Test: Default values for missing metadata fields**

#### Phase 3C: Current State Section Implementation
- [ ] Extend epic structs for current_state section
- [ ] Add XML parsing for active_phase, active_task, next_action
- [ ] Implement automatic current_state updates via commands
- [ ] Integration with start/done commands for state tracking
- [ ] Query service updates to expose current_state data
- [ ] Current command enhancement to display current_state

#### Phase 3D: Write Current State Section Tests **IMMEDIATELY AFTER 3C**
Epic 7 Test Scenarios Covered:
- [ ] **Test: Current state section parsing from XML**
- [ ] **Test: Automatic updates when starting phases/tasks**
- [ ] **Test: Current command displays current_state data**
- [ ] **Test: Next action tracking and updates**
- [ ] **Test: Query service exposes current_state correctly**
- [ ] **Test: Integration with lifecycle commands**

### Phase 4: Additional XML Sections + Integration Tests (Low Priority)

#### Phase 4A: Workflow & Requirements Sections
- [ ] Extend epic structs for workflow and requirements sections
- [ ] Add XML parsing for free-text workflow instructions
- [ ] Add XML parsing for core stories and technical requirements
- [ ] Update docs generation to include workflow and requirements
- [ ] Query service updates to expose workflow/requirements data
- [ ] Validation for workflow and requirements content

#### Phase 4B: Write Workflow & Requirements Tests **IMMEDIATELY AFTER 4A**
Epic 7 Test Scenarios Covered:
- [ ] **Test: Workflow section parsing and display**
- [ ] **Test: Requirements section parsing and display**
- [ ] **Test: Documentation generation includes new sections**
- [ ] **Test: Query commands expose workflow/requirements data**
- [ ] **Test: Content validation for workflow/requirements**
- [ ] **Test: Backward compatibility with existing epics**

#### Phase 4C: Dependencies & Outline Sections
- [ ] Extend epic structs for dependencies and outline sections
- [ ] Add XML parsing for epic dependencies with requirement levels
- [ ] Add XML parsing for outline phase overview
- [ ] Implement outline synchronization with phases
- [ ] Dependencies validation and reference checking
- [ ] Status command enhancement to display outline

#### Phase 4D: Write Dependencies & Outline Tests **IMMEDIATELY AFTER 4C**
Epic 7 Test Scenarios Covered:
- [ ] **Test: Dependencies section parsing and validation**
- [ ] **Test: Outline section parsing and synchronization**
- [ ] **Test: Dependency reference validation**
- [ ] **Test: Outline auto-sync with phase changes**
- [ ] **Test: Status command displays outline**
- [ ] **Test: Dependencies shown in docs generation**

#### Phase 4E: Full Integration & Cross-Command Testing
- [ ] Integration between all new commands and XML sections
- [ ] End-to-end workflow testing with new features
- [ ] Cross-command consistency (error formats, XML output)
- [ ] Documentation generation with all new sections
- [ ] Query command updates for all new data
- [ ] Help system integration for all new commands
- [ ] Global flag consistency across new commands

#### Phase 4F: Write Integration Tests **IMMEDIATELY AFTER 4E**
Epic 7 Test Scenarios Covered:
- [ ] **Test: End-to-end workflow with all new features**
- [ ] **Test: Documentation generation includes all sections**
- [ ] **Test: Query commands expose all new data**
- [ ] **Test: Cross-command XML format consistency**
- [ ] **Test: Help system completeness for new commands**
- [ ] **Test: Global flag handling across new features**
- [ ] **Test: Backward compatibility maintained throughout**

## Development Workflow Per Phase

For **EACH** phase:

1. **Implement Code** (Phase XA or XC)
2. **Write Tests IMMEDIATELY** (Phase XB or XD) 
3. **Run Tests & Verify** - All tests must pass
4. **Run Linting/Type Checking** - Code must be clean
5. **NEVER move to next phase with failing tests**

## Epic 7 Specific Considerations

### Dependencies & Integration
- **Epic 1:** CLI framework, XML processing, storage interface, configuration management
- **Epic 2:** Query service for data exposure and current command enhancement
- **Epic 3:** Lifecycle service for pause/resume epic functionality
- **Epic 5:** Event logging system for log command integration
- **Epic 6:** Documentation generation for new XML sections

### Technical Requirements
- **Command Framework:** Use existing urfave/cli/v3 patterns for consistency
- **XML Compatibility:** Ensure backward compatibility with existing epic files
- **Atomic Operations:** File updates must be safe and rollback-capable
- **Event Integration:** Log command integrates with existing event system
- **State Management:** Pause/resume commands integrate with lifecycle service
- **Documentation:** New XML sections appear in docs generation

### Priority Implementation Order
1. **Log Command** - Essential for event tracking as documented in README
2. SKIPPED
3. **Metadata Section** - Basic epic information enhancement
4. **Current State Section** - Improves navigation and handoff
5. **Additional Sections** - Documentation and workflow enhancements

### File Structure
```
├── cmd/
│   ├── log.go              # Event logging command with files/type flags
│   ├── pause_epic.go       # Epic pause command
│   └── resume_epic.go      # Epic resume command
├── internal/
│   ├── epic/               # Extended epic structures
│   │   ├── epic.go         # Enhanced structs for new XML sections
│   │   └── validation.go   # Validation for new fields
│   ├── lifecycle/          # Enhanced lifecycle service
│   │   └── service.go      # Pause/resume functionality
│   └── service/            # Enhanced services
│       └── log_service.go  # Log command service logic
└── testdata/
    ├── epic-with-metadata.xml      # Epic with metadata section
    ├── epic-with-current-state.xml # Epic with current_state
    ├── epic-paused.xml             # Epic in paused status
    └── epic-full-sections.xml      # Epic with all new sections
```

### Log Command Implementation Details
```go
type LogCommand struct {
    Storage storage.Interface
    Query   *query.Service
}

type LogOptions struct {
    Message     string
    Files       []string  // Format: "path:action"
    Type        string    // implementation, blocker, issue, etc.
    Timestamp   time.Time // --time flag support
}

func (c *LogCommand) Execute(opts LogOptions) error {
    // Validate type
    // Parse files with actions
    // Create event with metadata
    // Append to epic XML atomically
    // Return success/error
}
```

### Enhanced Epic Structure
```go
type Epic struct {
    // Existing fields...
    
    // New sections
    Metadata    *EpicMetadata    `xml:"metadata,omitempty"`
    Workflow    string           `xml:"workflow,omitempty"`
    Requirements string          `xml:"requirements,omitempty"`
    Dependencies []Dependency    `xml:"dependencies>dependency,omitempty"`
    CurrentState *CurrentState   `xml:"current_state,omitempty"`
    Outline     []OutlinePhase   `xml:"outline>phase,omitempty"`
}

type EpicMetadata struct {
    Created         time.Time `xml:"created"`
    Assignee        string    `xml:"assignee"`
    EstimatedEffort string    `xml:"estimated_effort"`
}

type CurrentState struct {
    ActivePhase string `xml:"active_phase"`
    ActiveTask  string `xml:"active_task"`
    NextAction  string `xml:"next_action"`
}
```

## Benefits of This Approach

✅ **Immediate Feedback** - Catch issues as soon as code is written  
✅ **Working Code** - Each phase delivers tested, working functionality  
✅ **Epic 7 Coverage** - All specification requirements distributed across phases  
✅ **Incremental Progress** - Agents can use new commands after each phase  
✅ **Risk Mitigation** - Problems caught early, not at the end  
✅ **Quality Assurance** - No untested code makes it to later phases  
✅ **Backward Compatibility** - Existing epic files continue to work  

## Test Distribution Summary

- **Phase 1 Tests:** 14 scenarios (Log command foundation, file tracking)
- **Phase 3 Tests:** 14 scenarios (Metadata and current_state sections)
- **Phase 4 Tests:** 18 scenarios (Additional sections, full integration)

**Total: All Epic 7 requirements and acceptance criteria covered across all phases**

---

## Implementation Status

### EPIC 7: MISSING FEATURES IMPLEMENTATION - PENDING ⏳
### Current Status: READY FOR IMPLEMENTATION

### Success Criteria Summary

#### Commands Implementation
- [ ] `agentpm log` with --files and --type flags works correctly
- [ ] All new commands have comprehensive test coverage
- [ ] All tests pass after implementation

#### XML Structure Enhancement
- [ ] All new sections (metadata, workflow, requirements, dependencies, current_state, outline) parse correctly
- [ ] Documentation generation includes new sections appropriately
- [ ] Query commands expose new data through appropriate interfaces
- [ ] Backward compatibility maintained with existing epic files
- [ ] Validation prevents invalid data in new sections

#### Integration & Quality
- [ ] Handoff reports include relevant new information
- [ ] Status commands display outline information
- [ ] Current command shows current_state data
- [ ] All existing functionality continues to work unchanged
- [ ] Test coverage > 90% for all new functionality

### Definition of Done
- [ ] All acceptance criteria verified with automated tests
- [ ] New commands execute in < 200ms for typical epic files
- [ ] Test coverage > 90% for new functionality
- [ ] All error cases handled gracefully with clear messages
- [ ] Backward compatibility verified with existing epic files
- [ ] Documentation generation enhanced with new sections
- [ ] Integration tests verify end-to-end workflows with new features