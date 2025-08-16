# EPIC-7: Missing Features Implementation

## Overview

**Epic ID:** 7  
**Name:** Missing Features Implementation  
**Status:** pending  
**Priority:** medium  

**Goal:** Implement missing commands and XML structure features identified from README vs implementation analysis to achieve full feature parity with the documented specification.

## Implementation Tasks Required

### 🔥 **High Priority: Missing Commands**

#### 1. **Implement `agentpm log` command** 
- Add `--files` flag for file tracking (e.g., `--files="src/Pagination.js:added"`)
- Add `--type` flag for event types (e.g., `--type=blocker`, `--type=issue`)
- Default type should be "implementation"
- Should append events to the epic XML file
- Example: `agentpm log "Implemented pagination controls" --files="src/Pagination.js:added"`

### 🏗️ **Low Priority: Enhanced Epic XML Structure Support**

#### 2. **Add `<metadata>` section support**
- `<created>` timestamp
- `<assignee>` field
- `<estimated_effort>` field
- Should be parsed and accessible via query commands

#### 3. **Add `<workflow>` section support**
- Free-text workflow instructions
- Should be displayed in docs generation
- Should be accessible via query commands

#### 4. **Add `<requirements>` section support**
- Core stories and technical requirements
- Should be displayed in docs generation
- Should be accessible via query commands

#### 5. **Add `<dependencies>` section support**
- Epic dependencies with requirement levels
- Should be displayed in docs generation
- Should validate dependency references

#### 6. **Add `<current_state>` section support**
- `<active_phase>`, `<active_task>`, `<active_test>`, `<next_action>`
- Should be automatically updated by commands
- Should be accessible via current command

#### 7. **Add `<outline>` section support**
- Quick phase overview for scanning
- Should be automatically synchronized with phases
- Should be displayed in status commands

## Priority Order Recommendation

1. **`agentpm log`** - Essential for event tracking as shown in README
2. **`agentpm pause-epic`** & **`agentpm resume-epic`** - Complete epic lifecycle
3. **Metadata section** - Basic epic information enhancement
4. **Current state section** - Improves navigation and handoff
5. **Outline section** - Better status visualization
6. **Requirements, dependencies, workflow** - Documentation enhancement

## Technical Requirements

### Command Implementation
- Follow existing CLI patterns using `github.com/urfave/cli/v3`
- Implement comprehensive test coverage for all new commands
- Ensure XML parsing and writing compatibility
- Add proper error handling and validation

### XML Structure Enhancement
- Extend `internal/epic/epic.go` structs to support new sections
- Update XML marshaling/unmarshaling tags
- Ensure backward compatibility with existing epic files
- Add validation for new fields

### Integration Requirements
- Update documentation generation to include new sections
- Enhance query commands to expose new data
- Update handoff reports to include relevant new information
- Ensure all new features work with existing workflow

## Success Criteria

### Commands
- [ ] `agentpm log` with --files and --type flags works correctly
- [ ] All new commands have comprehensive test coverage
- [ ] All tests pass after implementation

### XML Structure
- [ ] All new sections parse correctly from XML files
- [ ] Documentation generation includes new sections appropriately
- [ ] Query commands expose new data through appropriate interfaces
- [ ] Backward compatibility maintained with existing epic files
- [ ] Validation prevents invalid data in new sections

### Integration
- [ ] Handoff reports include relevant new information
- [ ] Status commands display outline information
- [ ] Current command shows current_state data
- [ ] All existing functionality continues to work unchanged

## Dependencies

- Epic 1: Foundation & Configuration (completed)
- Epic 2: Query Commands (completed)
- Epic 3: Epic Lifecycle (completed)
- Epic 4: Task & Phase Management (completed)  
- Epic 5: Test Management & Event Logging (completed)
- Epic 6: Handoff & Documentation (completed)

## Estimated Effort

**High Priority Commands:** 2-3 days
**XML Structure Enhancements:** 3-4 days  
**Total Epic:** 5-7 days

## Notes

This epic represents the final gap between the documented specification in README.md and the actual implementation. Completing this epic will ensure 100% feature parity with the documented capabilities and provide a complete agent project management solution.