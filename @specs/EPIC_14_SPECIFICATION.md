# EPIC-14 SPECIFICATION: Transition Chain Testing Framework

## Overview

**Epic ID:** 14  
**Name:** Transition Chain Testing Framework  
**Duration:** 3-4 days  
**Status:** pending  
**Priority:** high  

**Goal:** Implement a fluent testing framework that enables compact validation of complex state transition chains using builder patterns and assertion APIs, leveraging existing memory storage and command services.

## Business Context

This epic introduces a sophisticated testing framework specifically designed for validating complex state transition workflows in AgentPM. The framework enables developers to build initial XML states, execute transition commands, and assert final states using a fluent API that is both readable and maintainable. This addresses the current challenge of testing lengthy transition sequences where manual XML assertions are verbose and error-prone.

## User Stories

### Primary User Stories
- **As a developer, I can build initial epic states using a fluent builder pattern** so that I can create test scenarios without manual XML construction
- **As a developer, I can execute transition chains with method chaining** so that I can simulate complex workflows in a readable manner
- **As a developer, I can assert final states using fluent assertions** so that I can validate outcomes without verbose XML parsing
- **As a developer, I can leverage existing command services for transitions** so that I test the actual production code paths

### Secondary User Stories
- **As a developer, I can create reusable state builders** so that I can share common test setups across multiple test cases
- **As a developer, I can snapshot transition chain outcomes** so that I can detect regressions in complex workflows
- **As a developer, I can debug failed transitions with detailed state information** so that I can quickly identify issues in complex scenarios
- **As a developer, I can validate intermediate states during transitions** so that I can catch issues early in the chain

## Technical Requirements

### Core Dependencies
- **Memory Storage:** `internal/storage/memory.go` for isolated test state
- **Command Services:** Existing lifecycle and command services for authentic transitions
- **Snapshot Testing:** `internal/testing/snapshots.go` for regression detection
- **XML Processing:** `github.com/beevik/etree` for state manipulation

### Builder Pattern Architecture
Based on existing epic structure:
- **Epic Builder:** Create epics with metadata, status, and timing
- **Phase Builder:** Add phases with dependencies and status
- **Task Builder:** Add tasks with phase assignments and criteria
- **Test Builder:** Add tests with task assignments and validation rules

### Command Chain Integration
- Leverage existing command services: `StartEpicService`, `DonePhaseService`, etc.
- Memory storage factory for isolated test execution
- Error handling consistent with production command behavior

## Functional Requirements

### FR-1: Fluent State Builder
**Usage:** 
```go
epic := EpicBuilder("test-epic").
    WithStatus("pending").
    WithAssignee("agent_claude").
    WithPhase("1A", "Setup", "pending").
    WithPhase("1B", "Development", "pending").
    WithTask("1A_1", "1A", "Initialize Project", "pending").
    WithTask("1A_2", "1A", "Configure Tools", "pending").
    WithTest("T1A_1", "1A_1", "Test Project Init", "pending").
    Build()
```

**Behavior:**
- Constructs valid epic XML structure using builder pattern
- Validates relationships between phases, tasks, and tests
- Generates appropriate IDs and timestamps
- Creates epic compatible with existing storage interface

### FR-2: Transition Chain Execution
**Usage:**
```go
result := TransitionChain(epic).
    StartEpic().
    StartPhase("1A").
    StartTask("1A_1").
    PassTest("T1A_1").
    DoneTask("1A_1").
    DonePhase("1A").
    StartPhase("1B").
    Execute()
```

**Behavior:**
- Uses memory storage for isolated test execution
- Invokes actual command services for authentic behavior
- Maintains command execution order and dependencies
- Captures intermediate states for debugging
- Returns final epic state and execution metadata

### FR-3: Fluent Assertion API
**Usage:**
```go
result.Assert().
    EpicStatus("wip").
    PhaseStatus("1A", "done").
    PhaseStatus("1B", "wip").
    TaskStatus("1A_1", "done").
    TestStatus("T1A_1", "passed").
    HasEvent("epic_started").
    HasEvent("phase_completed").
    EventCount(6).
    NoErrors()
```

**Behavior:**
- Provides readable assertion methods for common validations
- Generates helpful error messages with state context
- Supports complex assertions on nested XML elements
- Integrates with standard Go testing framework
- Allows custom assertion predicates

### FR-4: Snapshot Integration
**Usage:**
```go
result.Assert().
    MatchSnapshot("complex_transition_flow").
    MatchXMLSnapshot("final_state", result.FinalState)
```

**Behavior:**
- Integrates with existing snapshot testing framework
- Supports both full state and selective element snapshots
- Provides clear diff output for snapshot mismatches
- Handles XML normalization for consistent comparisons

### FR-5: Error State Validation
**Usage:**
```go
result := TransitionChain(epic).
    StartEpic().
    DoneEpic().  // Should fail - no phases completed
    Execute()

result.Assert().
    HasError("cannot_complete_epic").
    ErrorContains("incomplete phases").
    EpicStatus("wip")  // Should remain in WIP state
```

**Behavior:**
- Captures and validates expected error conditions
- Maintains state consistency after failed transitions
- Provides detailed error context for debugging
- Supports testing of validation rules and business logic

### FR-6: Intermediate State Validation
**Usage:**
```go
chain := TransitionChain(epic).
    StartEpic().
    Assert().EpicStatus("wip").  // Validate intermediate state
    StartPhase("1A").
    Assert().PhaseStatus("1A", "wip").
    Execute()
```

**Behavior:**
- Allows assertions at any point in the transition chain
- Maintains chain fluency with embedded assertions
- Provides early failure detection in complex sequences
- Supports conditional branching based on intermediate states

## Non-Functional Requirements

### NFR-1: Performance
- Transition chain execution completes in < 500ms for typical scenarios
- Memory usage remains < 50MB for complex test suites
- Builder pattern construction in < 10ms for standard epics

### NFR-2: Usability
- Intuitive fluent API following Go conventions
- Clear error messages with state context and suggestions
- Comprehensive documentation with practical examples
- IDE-friendly method signatures with descriptive names

### NFR-3: Reliability
- Isolated test execution prevents cross-test interference
- Deterministic behavior with consistent state initialization
- Robust error handling for malformed builder input
- Memory cleanup after test execution

### NFR-4: Maintainability
- Modular design supporting easy extension
- Clear separation between builders, executors, and assertions
- Consistent patterns across different entity types
- Comprehensive test coverage for framework itself

## Data Model

### TransitionChainResult
```go
type TransitionChainResult struct {
    InitialState    *epic.Epic
    FinalState      *epic.Epic
    IntermediateStates []StateSnapshot
    ExecutedCommands []CommandExecution
    Errors          []TransitionError
    ExecutionTime   time.Duration
    MemoryUsage     int64
}

type StateSnapshot struct {
    Command     string
    Timestamp   time.Time
    EpicState   *epic.Epic
    Success     bool
    Error       error
}
```

### Builder Configuration
```go
type EpicBuilderConfig struct {
    ID              string
    Name            string
    Status          string
    CreatedAt       *time.Time
    Assignee        string
    Description     string
    Phases          []PhaseConfig
    DefaultValues   bool  // Auto-generate IDs, timestamps
}
```

## Error Handling

### Error Categories
1. **Builder Validation Errors:** Invalid relationships, missing required fields
2. **Transition Execution Errors:** Command failures, state validation errors
3. **Assertion Failures:** State validation mismatches, missing elements
4. **System Errors:** Memory allocation, storage failures

### Error Context Enhancement
```go
type TransitionError struct {
    Command         string
    ExpectedState   string
    ActualState     string
    Epic           *epic.Epic
    ContextualInfo  map[string]interface{}
    Suggestions     []string
}
```

## Acceptance Criteria

### AC-1: Basic Builder Pattern
- **GIVEN** I want to create a test epic
- **WHEN** I use `EpicBuilder("test").WithPhase("1A", "Setup", "planning").Build()`
- **THEN** I should get a valid epic with phase 1A in planning status

### AC-2: Simple Transition Chain
- **GIVEN** I have a built epic in planning status
- **WHEN** I run `TransitionChain(epic).StartEpic().Execute()`
- **THEN** The epic status should change to "wip"

### AC-3: Complex Workflow Validation
- **GIVEN** I have an epic with multiple phases and tasks
- **WHEN** I execute a complete workflow transition chain
- **THEN** All final states should match expected values

### AC-4: Error State Handling
- **GIVEN** I have an epic with incomplete phases
- **WHEN** I try to complete the epic
- **THEN** I should get a validation error and the epic should remain in wip status

### AC-5: Fluent Assertion API
- **GIVEN** I have executed a transition chain
- **WHEN** I use fluent assertions like `.EpicStatus("wip").PhaseStatus("1A", "done")`
- **THEN** Assertions should validate correctly with helpful error messages

### AC-6: Snapshot Integration
- **GIVEN** I have a complex transition chain result
- **WHEN** I use `.MatchSnapshot("test_scenario")`
- **THEN** The snapshot should capture the full final state

### AC-7: Intermediate State Validation
- **GIVEN** I have a multi-step transition chain
- **WHEN** I add intermediate assertions like `.StartEpic().Assert().EpicStatus("wip")`
- **THEN** Intermediate validations should execute without breaking the chain

### AC-8: Memory Isolation
- **GIVEN** I run multiple transition chain tests
- **WHEN** Tests execute in parallel or sequence
- **THEN** Each test should have isolated state without interference

## Testing Strategy

### Test Categories
- **Unit Tests (60%):** Builder pattern, assertion methods, error handling
- **Integration Tests (30%):** Command service integration, memory storage interaction
- **End-to-End Tests (10%):** Complete workflow scenarios with complex state chains

### Test Data Requirements
- **Sample Epic Structures:** Various complexity levels for builder testing
- **Transition Scenarios:** Valid and invalid transition sequences
- **Error Cases:** Malformed builders, invalid transitions, assertion failures

### Performance Testing
- **Builder Performance:** Create 1000 epics in < 1 second
- **Chain Execution:** Execute 10-step transition chains in < 100ms
- **Memory Usage:** Support 50 concurrent test scenarios

## Implementation Phases

### Phase 14A: Core Builder Framework (Day 1)
- Epic, Phase, Task, and Test builder implementations
- Validation logic for entity relationships
- Integration with existing epic data structures
- Basic error handling and validation

### Phase 14B: Transition Chain Engine (Day 1-2)
- TransitionChain executor using memory storage
- Command service integration
- State snapshot capture and management
- Error handling for failed transitions

### Phase 14C: Fluent Assertion API (Day 2-3)
- Assertion builder with fluent methods
- State validation and comparison logic
- Error message generation with context
- Integration with Go testing framework

### Phase 14D: Advanced Features & Integration (Day 3-4)
- Snapshot testing integration
- Intermediate state validation
- Performance optimization
- Comprehensive test coverage and documentation

## Definition of Done

- [ ] All acceptance criteria verified with automated tests
- [ ] Builder pattern creates valid epic structures matching existing XML schema
- [ ] Transition chains execute using actual command services
- [ ] Fluent assertions provide clear, actionable error messages
- [ ] Snapshot integration works with existing testing framework
- [ ] Memory isolation prevents test interference
- [ ] Performance meets specified benchmarks
- [ ] Test coverage > 90% for framework components
- [ ] Documentation includes practical examples and patterns
- [ ] Integration with existing CLI testing patterns

## Dependencies and Risks

### Dependencies
- **Epic 1:** Foundation CLI structure and command services (done)
- **Memory Storage:** `internal/storage/memory.go` implementation
- **Snapshot Testing:** Existing `internal/testing/snapshots.go`

### Risks
- **Medium Risk:** Command service integration complexity may require refactoring
- **Medium Risk:** Memory storage isolation may not handle all edge cases
- **Low Risk:** Performance degradation with complex transition chains
- **Low Risk:** Fluent API design may become unwieldy with many assertion types

### Mitigation Strategies
- Early prototype with simple command service integration
- Comprehensive memory isolation testing with concurrent scenarios
- Performance benchmarking from early development stages
- Iterative API design with developer feedback
- Fallback to traditional testing approaches for complex edge cases

## Notes

- This framework should serve as the foundation for all complex state transition testing
- Builder patterns should be extensible for future epic schema evolution
- Consider generating builders from XML schema for consistency
- Documentation should include migration guide from existing XML-based tests
- Framework should support both simple and complex testing scenarios
- Consider command line tool for generating test builders from existing epics