# EPIC-9: Testing and UX Cleanup

## Overview

**Epic ID:** 9
**Name:** Testing and UX Cleanup  
**Status:** pending  
**Priority:** high  

**Goal:** Improve testing reliability through snapshot testing, enhance user experience with friendly messaging and helpful error hints, and fix dependency validation to include tests alongside tasks in phase lifecycle management.

## Implementation Tasks Required

### 🧪 **High Priority: Snapshot Testing Refactor**

#### 1. **Implement snapshot testing framework**
- Use - **github.com/gkampitakis/go-snaps** for snapshot testing
- Create snapshot infrastructure for XML output testing
- Establish patterns for snapshot file organization and naming

#### 2. **Migrate XML output tests to snapshots**
- Identify all tests using direct string assertions for XML output
- Replace fragile string matching with snapshot comparisons
- Ensure snapshot tests capture meaningful output variations
- Add snapshot update command/flag for easy maintenance

#### 3. **Create snapshot maintenance workflow**
- Document how to update snapshots after intentional changes
- Establish review process for snapshot changes
- Add CI validation for snapshot consistency

### 💬 **High Priority: Friendly User Messages**

#### 4. **Implement friendly messaging system**
- Create message severity levels: `error`, `warning`, `info`, `success`
- Design consistent message formatting and presentation
- Implement friendly responses for non-critical operations

#### 5. **Handle redundant state transitions gracefully**
- Starting already-started entities (phases, tasks, tests) → friendly info message
- Completing already-completed entities → friendly info message
- Provide clear feedback about current state without error exit codes
- Example: "Phase 'Implementation' is already active. No action needed."

### 🎯 **Medium Priority: Error Hints System**

#### 6. **Design comprehensive error hint framework**
- Add `hint` field to error structures in XML and CLI output
- Create hint generation system based on error context
- Establish patterns for actionable, specific hints

#### 7. **Implement context-aware error hints**
- **Phase conflicts:** "To start 'Testing' phase, first complete active phase 'Implementation' with: `agentpm complete-phase implementation`"
- **Missing dependencies:** "Phase 'Deployment' requires completion of tasks: [list]. Complete with: `agentpm complete-task <task-id>`"
- **Invalid references:** "Task 'invalid-id' not found. List available tasks with: `agentpm query tasks`"
- **Epic state issues:** "Epic not started. Initialize with: `agentpm start-epic`"

#### 8. **Audit and enhance existing error messages**
- Catalog all current error scenarios across commands
- Brainstorm helpful hints for each error type
- Standardize error message format: `[Error]: <description>` + `[Hint]: <actionable_solution>`
- Add hints to XML error logging for consistency

### 🔗 **High Priority: Test Dependencies in Phase Management**

#### 9. **Extend phase dependency validation**
- Update phase starting logic to check for incomplete tests
- Update phase completion logic to validate test dependencies
- Ensure test completion status affects phase lifecycle decisions

#### 10. **Implement comprehensive dependency checking**
- Check both task AND test completion before allowing phase completion
- Prevent phase transitions when dependent tests are incomplete
- Provide clear messaging about blocking tests
- Example: "Cannot complete 'Implementation' phase. Incomplete tests: [test-ids]. Complete with: `agentpm complete-test <test-id>`"

## Technical Requirements

### Snapshot Testing Infrastructure
- Integrate snapshot testing library with existing test suite
- Ensure snapshots work with CI/CD pipeline
- Create helper functions for common snapshot patterns
- Document snapshot testing guidelines for team

### Message System Architecture
- Create message types and severity levels
- Implement consistent formatting across all commands
- Ensure friendly messages don't break scripting/automation
- Add configuration for message verbosity levels

### Error Enhancement Framework
- Extend error structures to include hint field
- Create hint generation utilities and templates
- Update XML schema to support error hints
- Ensure hints are contextual and actionable

### Dependency Validation Updates
- Modify phase lifecycle validation logic
- Update data models to properly track test dependencies
- Ensure backwards compatibility with existing epic files
- Add comprehensive test coverage for new dependency logic

## Success Criteria

### Snapshot Testing
- [ ] All XML output tests converted to snapshot-based testing
- [ ] Single command updates all snapshots after intentional changes
- [ ] Snapshot tests catch unintended output regressions
- [ ] Documentation for snapshot maintenance workflow

### User Experience
- [ ] Redundant operations show friendly messages instead of errors
- [ ] All error messages include actionable hints
- [ ] Message severity levels consistently applied
- [ ] User testing confirms improved experience

### Dependency Management
- [ ] Phase completion blocked by incomplete tests
- [ ] Phase starting validates test dependencies
- [ ] Clear messaging about test-related blockers
- [ ] All existing functionality preserved

### Code Quality
- [ ] Test suite more maintainable and less fragile
- [ ] Error handling consistent across all commands
- [ ] Comprehensive test coverage for new features
- [ ] Documentation updated for new patterns

## Estimated Effort

**Snapshot Testing Refactor:** 2-3 days  
**Friendly Messaging System:** 2 days  
**Error Hints Framework:** 3-4 days  
**Test Dependencies in Phases:** 2 days  
**Total Epic:** 9-11 days

## Priority Order Recommendation

1. **Test dependencies in phase management** - Critical workflow bug fix
2. **Snapshot testing refactor** - Improves development velocity
3. **Friendly messaging system** - Better user experience
4. **Error hints framework** - Enhanced usability

## Notes

This cleanup epic addresses fundamental quality-of-life improvements that will make the tool more reliable to develop and more pleasant to use. The snapshot testing refactor will significantly reduce maintenance overhead for XML output tests, while the UX improvements will make the tool more approachable for daily use.

The test dependency validation fix is particularly important as it closes a gap in the phase lifecycle logic that could lead to inconsistent project states.