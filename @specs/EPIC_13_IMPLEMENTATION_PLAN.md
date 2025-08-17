# EPIC-13: Status Enum Streamlining Implementation Plan
## Test-Driven Development Approach

### Phase 1: Unified Status Enums Definition + Tests (High Priority)

#### Phase 13A: Define Unified Status Enums
- [ ] Create new status enum types in internal/epic/status.go
- [ ] Define EpicStatus enum (pending, wip, done)
- [ ] Define PhaseStatus enum (pending, wip, done)
- [ ] Define TaskStatus enum (pending, wip, done, cancelled)
- [ ] Define TestStatus enum (pending, wip, done, cancelled)
- [ ] Define TestResult enum (passing, failing)
- [ ] Add string conversion methods for all status types
- [ ] Add validation methods for status enum values
- [ ] Create status transition mapping functions

#### Phase 13B: Write Status Enum Tests **IMMEDIATELY AFTER 13A**
Epic 13 Test Scenarios Covered:
- [ ] **Test: All status enum types are properly defined**
- [ ] **Test: String conversion methods work correctly for all statuses**
- [ ] **Test: Status validation methods reject invalid values**
- [ ] **Test: Status transition mapping functions work correctly**
- [ ] **Test: Status enum constants match expected string values**
- [ ] **Test: Status types implement proper interfaces**
- [ ] **Test: Invalid status values are handled gracefully**
- [ ] **Test: Status enum comparison operations work correctly**

#### Phase 13C: Status Validation Framework
- [ ] Create StatusValidationError struct with detailed error information
- [ ] Create BlockingItem struct for validation details
- [ ] Implement status transition validation engine
- [ ] Add business rule validation functions
- [ ] Create validation result structures
- [ ] Add detailed error message formatting with counts
- [ ] Implement validation error categorization
- [ ] Add structured validation result reporting

#### Phase 13D: Write Validation Framework Tests **IMMEDIATELY AFTER 13C**
Epic 13 Test Scenarios Covered:
- [ ] **Test: StatusValidationError contains all required fields**
- [ ] **Test: BlockingItem struct holds proper validation details**
- [ ] **Test: Status transition validation engine works correctly**
- [ ] **Test: Business rule validation functions catch violations**
- [ ] **Test: Validation result structures contain accurate data**
- [ ] **Test: Error message formatting includes proper counts**
- [ ] **Test: Validation error categorization works correctly**
- [ ] **Test: Structured validation results are properly formatted**

### Phase 2: Business Rules Implementation + Tests (High Priority)

#### Phase 2A: Phase Completion Business Rules
- [ ] Implement phase completion validation in internal/phases/validation.go
- [ ] Add pending/wip task counting for phase completion
- [ ] Add pending/wip test counting for phase completion
- [ ] Create detailed blocking item collection for phases
- [ ] Add phase completion error message generation
- [ ] Implement phase status transition rules
- [ ] Add phase completion prerequisite checking
- [ ] Integrate phase validation with existing phase service

#### Phase 2B: Write Phase Business Rules Tests **IMMEDIATELY AFTER 2A**
Epic 13 Test Scenarios Covered:
- [ ] **Test: Phase cannot be completed with pending tasks**
- [ ] **Test: Phase cannot be completed with wip tasks**
- [ ] **Test: Phase cannot be completed with pending tests**
- [ ] **Test: Phase cannot be completed with wip tests**
- [ ] **Test: Phase completion error messages include exact counts**
- [ ] **Test: Phase status transition rules are enforced**
- [ ] **Test: Phase completion prerequisite checking works**
- [ ] **Test: Phase validation integrates with phase service**

#### Phase 2C: Task Completion Business Rules
- [ ] Implement task completion validation in internal/tasks/validation.go
- [ ] Add pending/wip test counting for task completion
- [ ] Create detailed blocking item collection for tasks
- [ ] Add task completion error message generation
- [ ] Implement task status transition rules
- [ ] Add task completion prerequisite checking
- [ ] Add task cancellation validation rules
- [ ] Integrate task validation with existing task service

#### Phase 2D: Write Task Business Rules Tests **IMMEDIATELY AFTER 2C**
Epic 13 Test Scenarios Covered:
- [ ] **Test: Task cannot be completed with pending tests**
- [ ] **Test: Task cannot be completed with wip tests**
- [ ] **Test: Task completion error messages include exact counts**
- [ ] **Test: Task status transition rules are enforced**
- [ ] **Test: Task completion prerequisite checking works**
- [ ] **Test: Task cancellation validation rules work**
- [ ] **Test: Task validation integrates with task service**
- [ ] **Test: Task status transitions follow business rules**

### Phase 3: Test State Rules + CLI Commands + Tests (High Priority)

#### Phase 3A: Test State Business Rules
- [ ] Implement test state validation in internal/tests/validation.go
- [ ] Add test status transition rules (failing tests cannot be done)
- [ ] Add test result transition validation
- [ ] Create test cancellation with reason support
- [ ] Add test pass/fail transition validation
- [ ] Implement test active phase checking
- [ ] Add test state consistency validation
- [ ] Integrate test validation with existing test service

#### Phase 3B: Write Test State Rules Tests **IMMEDIATELY AFTER 3A**
Epic 13 Test Scenarios Covered:
- [ ] **Test: Failing tests cannot be marked as done**
- [ ] **Test: Test status transition rules are enforced**
- [ ] **Test: Test result transitions work correctly**
- [ ] **Test: Test cancellation requires valid reason**
- [ ] **Test: Test pass/fail transitions work correctly**
- [ ] **Test: Test active phase checking prevents wrong phase changes**
- [ ] **Test: Test state consistency validation works**
- [ ] **Test: Test validation integrates with test service**

#### Phase 3C: Simple Test CLI Commands Implementation
- [ ] Update cmd/pass.go to use new status system
- [ ] Update cmd/fail.go to use new status system
- [ ] Update cmd/cancel_test.go to use new status system with reason
- [ ] Add test active phase validation to all test commands
- [ ] Implement simple status transitions (pass: done+passing, fail: wip+failing)
- [ ] Add proper error handling for invalid test state transitions
- [ ] Ensure test commands validate test exists and is in active phase
- [ ] Add consistent command output formatting

#### Phase 3D: Write Test CLI Commands Tests **IMMEDIATELY AFTER 3C**
Epic 13 Test Scenarios Covered:
- [ ] **Test: pass command sets status=done, result=passing**
- [ ] **Test: fail command sets status=wip, result=failing**
- [ ] **Test: cancel command sets status=cancelled with reason**
- [ ] **Test: Test commands validate test is in active phase**
- [ ] **Test: Test commands handle invalid test IDs properly**
- [ ] **Test: Test state transitions work correctly**
- [ ] **Test: Test commands provide consistent output formatting**
- [ ] **Test: Test commands integrate with validation framework**

### Phase 4: Batch Commands Implementation + Tests (High Priority)

#### Phase 4A: Batch Command Validation Framework
- [ ] Create batch validation service in internal/commands/batch_validation.go
- [ ] Implement all-or-nothing validation for batch operations
- [ ] Add batch test existence validation
- [ ] Add batch active phase validation
- [ ] Add batch status transition validation
- [ ] Create comprehensive batch error reporting
- [ ] Implement batch success reporting with change details
- [ ] Add batch operation rollback capabilities

#### Phase 4B: Write Batch Validation Tests **IMMEDIATELY AFTER 4A**
Epic 13 Test Scenarios Covered:
- [ ] **Test: Batch validation rejects operations with any invalid test**
- [ ] **Test: Batch test existence validation works correctly**
- [ ] **Test: Batch active phase validation prevents wrong phase operations**
- [ ] **Test: Batch status transition validation enforces rules**
- [ ] **Test: Batch error reporting shows all validation failures**
- [ ] **Test: Batch success reporting includes all change details**
- [ ] **Test: Batch operation rollback works for partial failures**
- [ ] **Test: All-or-nothing principle is enforced**

#### Phase 4C: Batch CLI Commands Implementation
- [ ] Implement cmd/pass_batch.go with comprehensive validation
- [ ] Implement cmd/fail_batch.go with comprehensive validation
- [ ] Add batch command argument parsing and validation
- [ ] Implement batch operation execution with rollback
- [ ] Add detailed batch operation error reporting
- [ ] Add detailed batch operation success reporting
- [ ] Ensure batch commands integrate with existing validation framework
- [ ] Add batch command help text and examples

#### Phase 4D: Write Batch CLI Commands Tests **IMMEDIATELY AFTER 4C**
Epic 13 Test Scenarios Covered:
- [ ] **Test: pass-batch succeeds when all tests are valid**
- [ ] **Test: pass-batch fails with no changes when any test is invalid**
- [ ] **Test: fail-batch succeeds when all tests are valid**
- [ ] **Test: fail-batch fails with no changes when any test is invalid**
- [ ] **Test: Batch commands provide detailed error messages**
- [ ] **Test: Batch commands provide detailed success messages**
- [ ] **Test: Batch operation rollback works correctly**
- [ ] **Test: Batch commands validate tests are in active phase**

### Phase 5: Command Integration + Data Model Updates + Tests (Medium Priority)

#### Phase 5A: Update Existing Commands with New Validation
- [ ] Update cmd/done_phase.go to use new validation framework
- [ ] Update cmd/done_task.go to use new validation framework
- [ ] Update cmd/done_epic.go to use new validation framework
- [ ] Ensure all completion commands use business rule validation
- [ ] Add proper error reporting with counts to all commands
- [ ] Update command help text to reflect new validation
- [ ] Ensure backward compatibility for command interfaces
- [ ] Add consistent error output formatting across commands

#### Phase 5B: Write Command Integration Tests **IMMEDIATELY AFTER 5A**
Epic 13 Test Scenarios Covered:
- [ ] **Test: done-phase command validates with exact counts**
- [ ] **Test: done-task command validates with exact counts**
- [ ] **Test: done-epic command validates with exact counts**
- [ ] **Test: Completion commands use business rule validation**
- [ ] **Test: Error reporting includes proper counts and details**
- [ ] **Test: Command help text is accurate and helpful**
- [ ] **Test: Backward compatibility is maintained**
- [ ] **Test: Error output formatting is consistent**

#### Phase 5C: Data Model and XML Schema Updates
- [ ] Update Epic struct in internal/epic/epic.go with new status field
- [ ] Update Phase struct with new status field
- [ ] Update Task struct with new status field
- [ ] Update Test struct with new status and result fields
- [ ] Update XML marshaling/unmarshaling for new status fields
- [ ] Add status field validation in XML parsing
- [ ] Ensure new status fields are properly persisted
- [ ] Update XML schema documentation

#### Phase 5D: Write Data Model Tests **IMMEDIATELY AFTER 5C**
Epic 13 Test Scenarios Covered:
- [ ] **Test: Epic struct properly handles new status field**
- [ ] **Test: Phase struct properly handles new status field**
- [ ] **Test: Task struct properly handles new status field**
- [ ] **Test: Test struct properly handles status and result fields**
- [ ] **Test: XML marshaling includes new status fields**
- [ ] **Test: XML unmarshaling reads new status fields correctly**
- [ ] **Test: Status field validation works in XML parsing**
- [ ] **Test: New status fields are properly persisted**

### Phase 6: Integration & Performance Testing + Final Polish (Low Priority)

#### Phase 6A: Integration Testing and Performance Validation
- [ ] Create comprehensive integration test suite
- [ ] Test complete workflows with new status system
- [ ] Validate performance impact is under 10ms per validation
- [ ] Test edge cases with empty phases and orphaned tests
- [ ] Validate memory usage remains constant
- [ ] Test concurrent operations with status validation
- [ ] Add stress testing for batch operations
- [ ] Validate status system works with large epics

#### Phase 6B: Write Integration and Performance Tests **IMMEDIATELY AFTER 6A**
Epic 13 Test Scenarios Covered:
- [ ] **Test: Complete workflows work with new status system**
- [ ] **Test: Performance impact stays under 10ms requirement**
- [ ] **Test: Edge cases with empty phases work correctly**
- [ ] **Test: Orphaned tests are handled properly**
- [ ] **Test: Memory usage remains constant with epic size**
- [ ] **Test: Concurrent operations work correctly**
- [ ] **Test: Batch operations perform well under stress**
- [ ] **Test: Large epics work correctly with status validation**

#### Phase 6C: Documentation and Final Integration
- [ ] Update documentation with new status system
- [ ] Create status transition guide
- [ ] Add practical examples for all status operations
- [ ] Update CLI help text for all affected commands
- [ ] Create troubleshooting guide for status validation
- [ ] Add status system configuration options
- [ ] Implement status validation logging controls
- [ ] Final integration testing with complete system

#### Phase 6D: Write Documentation Tests **IMMEDIATELY AFTER 6C**
Epic 13 Test Scenarios Covered:
- [ ] **Test: All documentation examples work correctly**
- [ ] **Test: Status transition guide scenarios are validated**
- [ ] **Test: CLI help text includes accurate status information**
- [ ] **Test: Troubleshooting guide scenarios work**
- [ ] **Test: Status configuration options function properly**
- [ ] **Test: Status logging controls work correctly**
- [ ] **Test: Final integration scenarios pass completely**
- [ ] **Test: Complete system workflows use proper status validation**

## Development Workflow Per Phase

For **EACH** phase:

1. **Implement Code** (Phase 13A or 13C)
2. **Write Tests IMMEDIATELY** (Phase 13B or 13D) 
3. **Run Tests & Verify** - All tests must pass
4. **Run Linting/Type Checking** - Code must be clean
5. **NEVER move to next phase with failing tests**

## Epic 13 Specific Considerations

### Dependencies & Requirements
- **Epic 1:** Foundation CLI structure and XML handling (COMPLETED)
- **Existing Commands:** Current phase/task/test completion commands
- **XML Structure:** Current epic XML schema and data model
- **Test Service:** Existing internal/tests/service.go for integration
- **Phase/Task Services:** Current phase and task management functionality

### Technical Architecture
- **Status Enum System:** Unified status types across all entities
- **Validation Framework:** Comprehensive business rule enforcement
- **Batch Operations:** All-or-nothing batch command execution
- **Error Reporting:** Detailed error messages with exact counts
- **Performance:** Minimal overhead status validation (< 10ms)
- **Data Integrity:** Consistent status values across entire system

### File Structure
```
├── internal/
│   ├── epic/
│   │   ├── status.go                   # New unified status enums
│   │   └── epic.go                     # Updated with new status field
│   ├── phases/
│   │   └── validation.go               # Phase completion business rules
│   ├── tasks/
│   │   └── validation.go               # Task completion business rules
│   ├── tests/
│   │   └── validation.go               # Test state business rules
│   └── commands/
│       ├── batch_validation.go         # Batch operation validation
│       ├── pass_batch_service.go       # Batch pass operation service
│       └── fail_batch_service.go       # Batch fail operation service
├── cmd/
│   ├── pass_batch.go                   # Batch pass command
│   ├── fail_batch.go                   # Batch fail command
│   ├── pass.go                         # Updated pass command
│   ├── fail.go                         # Updated fail command
│   ├── done_phase.go                   # Updated with new validation
│   ├── done_task.go                    # Updated with new validation
│   └── done_epic.go                    # Updated with new validation
├── testdata/
│   ├── epic-status-validation.xml      # Epic for status validation testing
│   ├── epic-batch-operations.xml       # Epic for batch operation testing
│   └── epic-status-edge-cases.xml      # Edge cases for status testing
└── docs/
    └── status_system_guide.md          # Comprehensive status system documentation
```

## Testing Strategy

### Test Categories
- **Unit Tests (70%):** Status validation, business rules, batch operations
- **Integration Tests (25%):** Command interaction, data flow, complete workflows
- **Performance Tests (5%):** Validation overhead, batch operation performance

### Test Isolation
- Each test uses `t.TempDir()` for filesystem isolation
- Mock epic files with controlled status scenarios
- Status validation tests use isolated validation instances
- Performance tests use standardized operation loads

### Test Data Management
- Sample epic files with various status configurations
- Status transition test cases covering all business rules
- Edge case scenarios for comprehensive coverage
- Performance benchmarks with measurable validation loads
- Batch operation test cases with mixed valid/invalid scenarios
- Large epic files for stress testing status validation

## Benefits of This Approach

✅ **Unified Status System** - Consistent status values across all entities  
✅ **Business Rule Enforcement** - Clear validation with detailed error messages  
✅ **Batch Operation Support** - Efficient multi-test operations with validation  
✅ **Performance Validated** - Status validation overhead verified under 10ms  
✅ **Data Integrity** - Comprehensive validation prevents invalid state transitions  
✅ **Clear Error Messages** - Exact counts and actionable guidance for agents  

## Test Distribution Summary

- **Phase 1 Tests:** 16 scenarios (Status enums, validation framework)
- **Phase 2 Tests:** 16 scenarios (Phase and task business rules)
- **Phase 3 Tests:** 16 scenarios (Test state rules and CLI commands)
- **Phase 4 Tests:** 16 scenarios (Batch command validation and implementation)
- **Phase 5 Tests:** 16 scenarios (Command integration and data model updates)
- **Phase 6 Tests:** 16 scenarios (Integration testing and documentation)

**Total: All Epic 13 acceptance criteria and status-specific scenarios covered**

---

## Implementation Status

### EPIC 13: STATUS ENUM STREAMLINING - IN PROGRESS
### Current Status: PHASE 1 COMPLETE - Status System Foundation Built

### ✅ Phase 1 Achievements (COMPLETED)
- **Unified Status Enums:** Created consistent EpicStatus, PhaseStatus, TaskStatus, TestStatus, and TestResult enums
- **Status Validation Framework:** Built comprehensive StatusValidator with business rule enforcement
- **Transition Logic:** Implemented proper status transition validation for all entity types
- **Error Handling:** Created detailed StatusValidationError with blocking item reporting
- **Comprehensive Testing:** Added 100+ test cases covering all status functionality
- **Legacy Migration:** Successfully migrated from TestStatusPassed/Failed to TestStatusDone + TestResult system
- **XML Formatting:** Added XML error formatting for CLI output compliance
- **Performance Validated:** All tests pass with minimal performance overhead

### Progress Tracking
- [x] Phase 13A: Define Unified Status Enums ✅ COMPLETED 
- [x] Phase 13B: Write Status Enum Tests ✅ COMPLETED
- [x] Phase 13C: Status Validation Framework ✅ COMPLETED
- [x] Phase 13D: Write Validation Framework Tests ✅ COMPLETED
- [ ] Phase 2A: Phase Completion Business Rules
- [ ] Phase 2B: Write Phase Business Rules Tests
- [ ] Phase 2C: Task Completion Business Rules
- [ ] Phase 2D: Write Task Business Rules Tests
- [ ] Phase 3A: Test State Business Rules
- [ ] Phase 3B: Write Test State Rules Tests
- [ ] Phase 3C: Simple Test CLI Commands Implementation
- [ ] Phase 3D: Write Test CLI Commands Tests
- [ ] Phase 4A: Batch Command Validation Framework
- [ ] Phase 4B: Write Batch Validation Tests
- [ ] Phase 4C: Batch CLI Commands Implementation
- [ ] Phase 4D: Write Batch CLI Commands Tests
- [ ] Phase 5A: Update Existing Commands with New Validation
- [ ] Phase 5B: Write Command Integration Tests
- [ ] Phase 5C: Data Model and XML Schema Updates
- [ ] Phase 5D: Write Data Model Tests
- [ ] Phase 6A: Integration Testing and Performance Validation
- [ ] Phase 6B: Write Integration and Performance Tests
- [ ] Phase 6C: Documentation and Final Integration
- [ ] Phase 6D: Write Documentation Tests

### Definition of Done
- [ ] All entity types use unified status enums
- [ ] Business rules enforced with exact counts in error messages
- [ ] Failing tests cannot be marked as "done" without cancellation
- [ ] Batch commands (pass-batch, fail-batch) work with comprehensive validation
- [ ] All tests pass with new status system
- [ ] Performance impact < 10ms per validation
- [ ] Test coverage > 95% for status validation logic
- [ ] All error messages provide actionable guidance
- [ ] Documentation updated with new status values

### Dependencies
- **REQUIRED:** Current XML structure and data model
- **INTEGRATION:** Existing command infrastructure
- **COMPATIBILITY:** Current epic/phase/task/test management

### Status-Specific Considerations
- **Performance Impact:** Status validation must add minimal overhead to operations
- **Data Consistency:** All status values must follow unified system
- **Error Quality:** All error messages must include specific counts and guidance
- **Batch Operations:** All-or-nothing principle for batch commands
- **Business Rules:** Strict enforcement of completion prerequisites
- **User Experience:** Clear, actionable error messages for agents