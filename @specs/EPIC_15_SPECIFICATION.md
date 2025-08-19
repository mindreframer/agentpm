# EPIC-15 SPECIFICATION: Complex Transition Scenarios Framework

## Overview

**Epic ID:** 15  
**Name:** Complex Transition Scenarios Framework  
**Duration:** 2-3 days  
**Status:** pending  
**Priority:** high  

**Goal:** Define and implement 16 comprehensive transition scenarios for AgentPM using the EPIC 14 builder pattern framework, progressing from simple cases to medium complexity to validate all possible epic lifecycle workflows.

## Business Context

Building on the EPIC 14 Transition Chain Testing Framework, this epic defines standardized transition scenarios that can be used for regression testing, performance validation, and comprehensive workflow verification. These scenarios represent real-world usage patterns for AgentPM, ensuring that all edge cases and complex state transitions are properly handled.

## User Stories

### Primary User Stories
- **As a developer, I can run standardized transition scenarios** to validate that AgentPM handles all common workflows correctly
- **As a developer, I can use these scenarios for regression testing** so that new changes don't break existing functionality
- **As a developer, I can benchmark performance** using these scenarios to ensure AgentPM remains fast and efficient
- **As a QA engineer, I can verify complex edge cases** using predefined scenarios rather than manual testing

### Secondary User Stories
- **As a developer, I can extend these scenarios** to create custom test cases for new features
- **As a developer, I can use these scenarios for training** to understand AgentPM's capabilities and limitations
- **As a system integrator, I can validate AgentPM behavior** in different environments using standardized tests

## Technical Requirements

### Builder Pattern Integration
- All scenarios must use the EPIC 14 EpicBuilder pattern for state construction
- All scenarios must use the TransitionChain fluent API for command execution
- All scenarios must use the AssertionBuilder for state validation
- Memory isolation must be maintained across all scenarios

### Scenario Coverage Requirements
- **Simple Scenarios (1-6):** Single phase, linear progression, minimal dependencies
- **Medium Complexity Scenarios (7-12):** Multiple phases, parallel tasks, test failures and recovery
- **Complex Edge Cases (13-16):** Error conditions, validation failures, boundary conditions

## Functional Requirements

### FR-0: Complete Usage Pattern Example

**IMPORTANT:** Before showing individual scenarios, here's the complete pattern that must be followed for all scenarios:

```go
package scenarios

import (
    "github.com/mindreframer/agentpm/internal/testing/builders"
    "github.com/mindreframer/agentpm/internal/testing/executor"
    "github.com/mindreframer/agentpm/internal/testing/assertions"
)

func CompletePatternExample() {
    // Step 1: Build epic structure using EpicBuilder
    epic, err := builders.NewEpicBuilder("complete-example").
        WithPhase("1A", "Development", "pending").
        WithTask("1A_1", "1A", "Implement feature", "pending").
        WithTest("T1A_1", "1A_1", "1A", "Basic test", "pending").
        Build()
    
    if err != nil {
        panic(err) // Handle build errors appropriately
    }

    // Step 2: Create isolated test execution environment
    env := executor.CreateTestEnvironment(epic)
    defer env.Cleanup() // IMPORTANT: Always cleanup after test execution
    
    // Step 3: Execute transition chain using fluent API
    result, err := executor.NewTransitionChain(env).
        StartEpic().
        StartPhase("1A").
        StartTask("1A_1").
        PassTest("T1A_1").
        DoneTask("1A_1").
        DonePhase("1A").
        DoneEpic().
        Execute()
    
    if err != nil {
        panic(err) // Handle execution errors appropriately
    }

    // Step 4: Validate results using assertion framework
    assertions.Assert(result).
        EpicStatus("done").
        PhaseStatus("1A", "done").
        TaskStatus("1A_1", "done").
        TestStatusUnified("T1A_1", "done").
        TestResult("T1A_1", "passing").
        NoErrors().
        MustPass()
}
```

**Key Components Explained:**
- **EpicBuilder**: Constructs epic structure with phases, tasks, and tests
- **TestExecutionEnvironment**: Provides memory isolation and state management
- **TransitionChain**: Executes commands using actual AgentPM services
- **AssertionBuilder**: Validates final and intermediate states

**Note:** In the scenarios below, the environment creation step (`env := executor.CreateTestEnvironment(epic)`) is implied but must always be included in actual implementations.

**Test Naming Convention:** Each scenario must be implemented as a Go test function following this naming pattern:
- `TestEpic15_Scenario01_BasicEpicStartToCompletion()`
- `TestEpic15_Scenario02_TestFailureAndRecovery()`
- `TestEpic15_Scenario10_ComplexStateTransitionsWithAssertions()`

This ensures clear traceability between specification scenarios and actual test implementations.

### FR-1: Simple Linear Scenarios (Scenarios 1-6)

#### Scenario 1: Basic Epic Start-to-Completion
**Test Function:** `TestEpic15_Scenario01_BasicEpicStartToCompletion(t *testing.T)`  
**Description:** Simplest possible epic lifecycle with one phase, one task, one test
```go
// Build epic (Step 1 from complete pattern above)
epic, err := builders.NewEpicBuilder("simple-001").
    WithPhase("1A", "Development", "pending").
    WithTask("1A_1", "1A", "Implement feature", "pending").
    WithTest("T1A_1", "1A_1", "1A", "Basic test", "pending").
    Build()

// Create environment (Step 2 from complete pattern above)
env := executor.CreateTestEnvironment(epic)
defer env.Cleanup()

// Execute transitions (Step 3 from complete pattern above)
result, err := executor.NewTransitionChain(env).
    StartEpic().
    StartPhase("1A").
    StartTask("1A_1").
    PassTest("T1A_1").
    DoneTask("1A_1").
    DonePhase("1A").
    DoneEpic().
    Execute()

// Validate results (Step 4 from complete pattern above)
assertions.Assert(result).
    EpicStatus("done").
    PhaseStatus("1A", "done").
    TaskStatus("1A_1", "done").
    TestStatusUnified("T1A_1", "done").
    TestResult("T1A_1", "passing").
    NoErrors().
    MustPass()
```

#### Scenario 2: Test Failure and Recovery
**Test Function:** `TestEpic15_Scenario02_TestFailureAndRecovery(t *testing.T)`  
**Description:** Epic with test failure, then recovery to passing state
```go
// NOTE: Environment creation implied (see complete pattern above)
epic, err := builders.NewEpicBuilder("recovery-002").
    WithPhase("1A", "Development", "pending").
    WithTask("1A_1", "1A", "Implement feature", "pending").
    WithTest("T1A_1", "1A_1", "1A", "Unit test", "pending").
    Build()

env := executor.CreateTestEnvironment(epic)
defer env.Cleanup()

result, err := executor.NewTransitionChain(env).
    StartEpic().
    StartPhase("1A").
    StartTask("1A_1").
    FailTest("T1A_1").
    Assert().TestStatusUnified("T1A_1", "wip").TestResult("T1A_1", "failing").
    PassTest("T1A_1").
    DoneTask("1A_1").
    DonePhase("1A").
    DoneEpic().
    Execute()

assertions.Assert(result).
    EpicStatus("done").
    TestStatusUnified("T1A_1", "done").
    TestResult("T1A_1", "passing").
    HasEvent("test_failed").
    HasEvent("test_passed").
    NoErrors().
    MustPass()
```

#### Scenario 3: Multiple Tests in Single Task
**Test Function:** `TestEpic15_Scenario03_MultipleTestsInSingleTask(t *testing.T)`  
**Description:** One task with multiple tests that must all pass
```go
epic := EpicBuilder("multi-test-003").
    WithPhase("1A", "Development", "pending").
    WithTask("1A_1", "1A", "Implement feature", "pending").
    WithTest("T1A_1", "1A_1", "1A", "Unit test", "pending").
    WithTest("T1A_2", "1A_1", "1A", "Integration test", "pending").
    WithTest("T1A_3", "1A_1", "1A", "Performance test", "pending").
    Build()

result := TransitionChain(env).
    StartEpic().
    StartPhase("1A").
    StartTask("1A_1").
    PassTest("T1A_1").
    PassTest("T1A_2").
    PassTest("T1A_3").
    DoneTask("1A_1").
    DonePhase("1A").
    DoneEpic().
    Execute()

Assert(result).
    EpicStatus("done").
    TestStatusUnified("T1A_1", "done").
    TestStatusUnified("T1A_2", "done").
    TestStatusUnified("T1A_3", "done").
    EventCount(9). // epic_started, phase_started, task_started, 3 test_passed, task_completed, phase_completed, epic_completed
    NoErrors().
    MustPass()
```

#### Scenario 4: Sequential Tasks in Single Phase
**Description:** Multiple tasks that must be completed in sequence
```go
epic := EpicBuilder("sequential-004").
    WithPhase("1A", "Development", "pending").
    WithTask("1A_1", "1A", "Setup", "pending").
    WithTask("1A_2", "1A", "Implementation", "pending").
    WithTask("1A_3", "1A", "Testing", "pending").
    WithTest("T1A_1", "1A_1", "1A", "Setup test", "pending").
    WithTest("T1A_2", "1A_2", "1A", "Feature test", "pending").
    WithTest("T1A_3", "1A_3", "1A", "Final test", "pending").
    Build()

result := TransitionChain(env).
    StartEpic().
    StartPhase("1A").
    StartTask("1A_1").
    PassTest("T1A_1").
    DoneTask("1A_1").
    StartTask("1A_2").
    PassTest("T1A_2").
    DoneTask("1A_2").
    StartTask("1A_3").
    PassTest("T1A_3").
    DoneTask("1A_3").
    DonePhase("1A").
    DoneEpic().
    Execute()

Assert(result).
    EpicStatus("done").
    TaskStatus("1A_1", "done").
    TaskStatus("1A_2", "done").
    TaskStatus("1A_3", "done").
    AllCommandsSuccessful().
    NoErrors().
    MustPass()
```

#### Scenario 5: Parallel Test Execution
**Description:** Multiple tests in different tasks can be executed in parallel
```go
epic := EpicBuilder("parallel-005").
    WithPhase("1A", "Development", "pending").
    WithTask("1A_1", "1A", "Frontend", "pending").
    WithTask("1A_2", "1A", "Backend", "pending").
    WithTest("T1A_1", "1A_1", "1A", "Frontend test", "pending").
    WithTest("T1A_2", "1A_2", "1A", "Backend test", "pending").
    Build()

result := TransitionChain(env).
    StartEpic().
    StartPhase("1A").
    StartTask("1A_1").
    StartTask("1A_2").
    PassTest("T1A_1").
    PassTest("T1A_2").
    DoneTask("1A_1").
    DoneTask("1A_2").
    DonePhase("1A").
    DoneEpic().
    Execute()

Assert(result).
    EpicStatus("done").
    TaskStatus("1A_1", "done").
    TaskStatus("1A_2", "done").
    HasEvent("test_passed").
    EventCount(8). // epic, phase, 2 tasks, 2 tests, task done, phase done, epic done
    NoErrors().
    MustPass()
```

#### Scenario 6: Time-Based Transitions
**Description:** Transitions with specific timestamps for timing validation
```go
baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

epic := EpicBuilder("timed-006").
    WithCreatedAt(baseTime).
    WithPhase("1A", "Development", "pending").
    WithTask("1A_1", "1A", "Feature", "pending").
    WithTest("T1A_1", "1A_1", "1A", "Test", "pending").
    Build()

result := TransitionChain(env).
    WithTimeSource(func() time.Time { return baseTime.Add(time.Minute) }).
    StartEpicAt(baseTime.Add(time.Minute)).
    StartPhaseAt("1A", baseTime.Add(2*time.Minute)).
    StartTaskAt("1A_1", baseTime.Add(3*time.Minute)).
    PassTestAt("T1A_1", baseTime.Add(4*time.Minute)).
    DoneTaskAt("1A_1", baseTime.Add(5*time.Minute)).
    DonePhaseAt("1A", baseTime.Add(6*time.Minute)).
    DoneEpicAt(baseTime.Add(7*time.Minute)).
    Execute()

Assert(result).
    EpicStatus("done").
    ExecutionTime(10*time.Second).
    PhaseTransitionTiming("1A", 5*time.Minute).
    NoErrors().
    MustPass()
```

### FR-2: Medium Complexity Scenarios (Scenarios 7-12)

#### Scenario 7: Multi-Phase Epic with Dependencies
**Description:** Epic with multiple phases that must be completed in order
```go
epic := EpicBuilder("multi-phase-007").
    WithPhase("1A", "Planning", "pending").
    WithPhase("1B", "Development", "pending").
    WithPhase("1C", "Testing", "pending").
    WithTask("1A_1", "1A", "Requirements", "pending").
    WithTask("1B_1", "1B", "Implementation", "pending").
    WithTask("1C_1", "1C", "QA Testing", "pending").
    WithTest("T1A_1", "1A_1", "1A", "Requirement test", "pending").
    WithTest("T1B_1", "1B_1", "1B", "Unit test", "pending").
    WithTest("T1C_1", "1C_1", "1C", "Integration test", "pending").
    Build()

result := TransitionChain(env).
    StartEpic().
    StartPhase("1A").
    StartTask("1A_1").
    PassTest("T1A_1").
    DoneTask("1A_1").
    DonePhase("1A").
    StartPhase("1B").
    StartTask("1B_1").
    PassTest("T1B_1").
    DoneTask("1B_1").
    DonePhase("1B").
    StartPhase("1C").
    StartTask("1C_1").
    PassTest("T1C_1").
    DoneTask("1C_1").
    DonePhase("1C").
    DoneEpic().
    Execute()

Assert(result).
    EpicStatus("done").
    PhaseStatus("1A", "done").
    PhaseStatus("1B", "done").
    PhaseStatus("1C", "done").
    StateProgression([]string{"planning", "wip", "wip", "wip", "done"}).
    NoErrors().
    MustPass()
```

#### Scenario 8: Mixed Test Results with Recovery
**Description:** Some tests fail initially, requiring multiple recovery cycles
```go
epic := EpicBuilder("recovery-008").
    WithPhase("1A", "Development", "pending").
    WithTask("1A_1", "1A", "Feature A", "pending").
    WithTask("1A_2", "1A", "Feature B", "pending").
    WithTest("T1A_1", "1A_1", "1A", "Test A1", "pending").
    WithTest("T1A_2", "1A_1", "1A", "Test A2", "pending").
    WithTest("T1A_3", "1A_2", "1A", "Test B1", "pending").
    Build()

result := TransitionChain(env).
    StartEpic().
    StartPhase("1A").
    StartTask("1A_1").
    StartTask("1A_2").
    FailTest("T1A_1").  // First failure
    PassTest("T1A_2").  // This passes
    FailTest("T1A_3").  // Second failure
    PassTest("T1A_1").  // Recovery for first test
    PassTest("T1A_3").  // Recovery for second test
    DoneTask("1A_1").
    DoneTask("1A_2").
    DonePhase("1A").
    DoneEpic().
    Execute()

Assert(result).
    EpicStatus("done").
    TestResult("T1A_1", "passing").
    TestResult("T1A_2", "passing").
    TestResult("T1A_3", "passing").
    EventSequence([]string{"test_failed", "test_passed", "test_failed", "test_passed", "test_passed"}).
    NoErrors().
    MustPass()
```

#### Scenario 9: Batch Test Operations
**Description:** Using batch pass/fail operations on multiple tests
```go
epic := EpicBuilder("batch-009").
    WithPhase("1A", "Development", "pending").
    WithTask("1A_1", "1A", "Backend", "pending").
    WithTask("1A_2", "1A", "Frontend", "pending").
    WithTest("T1A_1", "1A_1", "1A", "API test", "pending").
    WithTest("T1A_2", "1A_1", "1A", "DB test", "pending").
    WithTest("T1A_3", "1A_2", "1A", "UI test", "pending").
    WithTest("T1A_4", "1A_2", "1A", "E2E test", "pending").
    Build()

result := TransitionChain(env).
    StartEpic().
    StartPhase("1A").
    StartTask("1A_1").
    StartTask("1A_2").
    // Simulate batch pass operation
    PassTest("T1A_1").
    PassTest("T1A_2").
    PassTest("T1A_3").
    PassTest("T1A_4").
    DoneTask("1A_1").
    DoneTask("1A_2").
    DonePhase("1A").
    DoneEpic().
    Execute()

Assert(result).
    EpicStatus("done").
    AllCommandsSuccessful().
    BatchAssertions([]func(*AssertionBuilder) *AssertionBuilder{
        func(ab *AssertionBuilder) *AssertionBuilder { return ab.TestResult("T1A_1", "passing") },
        func(ab *AssertionBuilder) *AssertionBuilder { return ab.TestResult("T1A_2", "passing") },
        func(ab *AssertionBuilder) *AssertionBuilder { return ab.TestResult("T1A_3", "passing") },
        func(ab *AssertionBuilder) *AssertionBuilder { return ab.TestResult("T1A_4", "passing") },
    }).
    NoErrors().
    MustPass()
```

#### Scenario 10: Complex State Transitions with Assertions
**Test Function:** `TestEpic15_Scenario10_ComplexStateTransitionsWithAssertions(t *testing.T)`  
**Description:** Epic with intermediate state validations throughout the transition chain
```go
epic := EpicBuilder("complex-010").
    WithPhase("1A", "Analysis", "pending").
    WithPhase("1B", "Implementation", "pending").
    WithTask("1A_1", "1A", "Research", "pending").
    WithTask("1B_1", "1B", "Code", "pending").
    WithTask("1B_2", "1B", "Test", "pending").
    WithTest("T1A_1", "1A_1", "1A", "Research validation", "pending").
    WithTest("T1B_1", "1B_1", "1B", "Code test", "pending").
    WithTest("T1B_2", "1B_2", "1B", "Integration test", "pending").
    Build()

result := TransitionChain(env).
    StartEpic().
    Assert().EpicStatus("wip").
    StartPhase("1A").
    Assert().PhaseStatus("1A", "wip").
    StartTask("1A_1").
    Assert().TaskStatus("1A_1", "wip").
    PassTest("T1A_1").
    DoneTask("1A_1").
    DonePhase("1A").
    Assert().PhaseStatus("1A", "done").
    StartPhase("1B").
    StartTask("1B_1").
    StartTask("1B_2").
    PassTest("T1B_1").
    PassTest("T1B_2").
    DoneTask("1B_1").
    DoneTask("1B_2").
    DonePhase("1B").
    DoneEpic().
    Execute()

Assert(result).
    EpicStatus("done").
    PhaseStatus("1A", "done").
    PhaseStatus("1B", "done").
    IntermediateState(2, func(e *epic.Epic) error {
        if len(e.Events) < 2 {
            return fmt.Errorf("expected at least 2 events at step 2")
        }
        return nil
    }).
    NoErrors().
    MustPass()
```

#### Scenario 11: Performance and Timing Validation
**Description:** Epic designed to test performance benchmarks and timing requirements
```go
epic := EpicBuilder("performance-011").
    WithPhase("1A", "Development", "pending").
    WithTask("1A_1", "1A", "Feature", "pending")

// Create many tests to validate performance
for i := 1; i <= 20; i++ {
    epic = epic.WithTest(fmt.Sprintf("T1A_%d", i), "1A_1", "1A", fmt.Sprintf("Test %d", i), "pending")
}

builtEpic, _ := epic.Build()
env := CreateTestEnvironment(builtEpic)

chain := TransitionChain(env).
    StartEpic().
    StartPhase("1A").
    StartTask("1A_1")

// Pass all 20 tests
for i := 1; i <= 20; i++ {
    chain = chain.PassTest(fmt.Sprintf("T1A_%d", i))
}

result := chain.
    DoneTask("1A_1").
    DonePhase("1A").
    DoneEpic().
    Execute()

Assert(result).
    EpicStatus("done").
    ExecutionTime(500*time.Millisecond). // Must complete within 500ms
    CommandCount(25). // StartEpic + StartPhase + StartTask + 20 PassTest + DoneTask + DonePhase + DoneEpic
    PerformanceBenchmark(500*time.Millisecond, 50).
    NoErrors().
    MustPass()
```

#### Scenario 12: Snapshot and Regression Testing
**Description:** Epic using snapshot testing for regression detection
```go
epic := EpicBuilder("snapshot-012").
    WithName("Snapshot Test Epic").
    WithAssignee("test_agent").
    WithPhase("1A", "Development", "pending").
    WithTask("1A_1", "1A", "Implement feature", "pending").
    WithTest("T1A_1", "1A_1", "1A", "Feature test", "pending").
    Build()

result := TransitionChain(env).
    StartEpic().
    StartPhase("1A").
    StartTask("1A_1").
    PassTest("T1A_1").
    DoneTask("1A_1").
    DonePhase("1A").
    DoneEpic().
    Execute()

Assert(result).
    EpicStatus("done").
    MatchSnapshot("complete_epic_flow").
    MatchSelectiveSnapshot("final_state_only", []string{"epic_status", "phase_status", "task_status"}).
    MatchXMLSnapshot("epic_xml", result.FinalState).
    NoErrors().
    MustPass()
```

### FR-3: Complex Edge Case Scenarios (Scenarios 13-16)

#### Scenario 13: Validation Failure - Task Completion Blocked
**Description:** Attempt to complete task with pending tests (should fail per EPIC 13)
```go
epic := EpicBuilder("validation-013").
    WithPhase("1A", "Development", "pending").
    WithTask("1A_1", "1A", "Feature", "pending").
    WithTest("T1A_1", "1A_1", "1A", "Test 1", "pending").
    WithTest("T1A_2", "1A_1", "1A", "Test 2", "pending").
    Build()

result := TransitionChain(env).
    StartEpic().
    StartPhase("1A").
    StartTask("1A_1").
    PassTest("T1A_1").
    // T1A_2 remains pending - this should block task completion
    DoneTask("1A_1"). // This should fail
    Execute()

Assert(result).
    HasErrors().
    ErrorCount(1).
    TaskStatus("1A_1", "wip"). // Should remain active due to validation failure
    TestStatusUnified("T1A_2", "pending"). // This test blocks completion
    CustomAssertion("validation_error", func(r *executor.TransitionChainResult) error {
        if len(r.Errors) == 0 {
            return fmt.Errorf("expected validation error for task completion with pending tests")
        }
        expectedError := "cannot complete task with pending tests"
        if !strings.Contains(r.Errors[0].Error(), "pending") {
            return fmt.Errorf("expected error about pending tests, got: %s", r.Errors[0].Error())
        }
        return nil
    }).
    MustPass()
```

#### Scenario 14: Phase Completion Blocked by Pending Tasks
**Description:** Attempt to complete phase with incomplete tasks (should fail per EPIC 13)
```go
epic := EpicBuilder("blocked-014").
    WithPhase("1A", "Development", "pending").
    WithTask("1A_1", "1A", "Task 1", "pending").
    WithTask("1A_2", "1A", "Task 2", "pending").
    WithTest("T1A_1", "1A_1", "1A", "Test 1", "pending").
    WithTest("T1A_2", "1A_2", "1A", "Test 2", "pending").
    Build()

result := TransitionChain(env).
    StartEpic().
    StartPhase("1A").
    StartTask("1A_1").
    PassTest("T1A_1").
    DoneTask("1A_1").
    // 1A_2 is still pending - this should block phase completion
    DonePhase("1A"). // This should fail
    Execute()

Assert(result).
    HasErrors().
    PhaseStatus("1A", "wip"). // Should remain active
    TaskStatus("1A_1", "done").   // This completed successfully
    TaskStatus("1A_2", "pending"). // This blocks phase completion
    CustomAssertion("phase_blocked", func(r *executor.TransitionChainResult) error {
        for _, err := range r.Errors {
            if strings.Contains(err.Error(), "pending") && strings.Contains(err.Error(), "task") {
                return nil // Found expected error
            }
        }
        return fmt.Errorf("expected error about pending tasks blocking phase completion")
    }).
    MustPass()
```

#### Scenario 15: Test Cancellation and Recovery
**Description:** Cancel tests and validate the state changes properly
```go
epic := EpicBuilder("cancellation-015").
    WithPhase("1A", "Development", "pending").
    WithTask("1A_1", "1A", "Feature", "pending").
    WithTask("1A_2", "1A", "Alternative", "pending").
    WithTest("T1A_1", "1A_1", "1A", "Primary test", "pending").
    WithTest("T1A_2", "1A_1", "1A", "Secondary test", "pending").
    WithTest("T1A_3", "1A_2", "1A", "Alternative test", "pending").
    Build()

// Note: Test cancellation might need to be implemented differently
// This shows the intended workflow using the current framework
result := TransitionChain(env).
    StartEpic().
    StartPhase("1A").
    StartTask("1A_1").
    StartTask("1A_2").
    FailTest("T1A_1").
    PassTest("T1A_2").
    // In a real implementation, we'd cancel T1A_1 and rely on T1A_2
    // For now, we'll pass T1A_1 to allow completion
    PassTest("T1A_1").
    PassTest("T1A_3").
    DoneTask("1A_1").
    DoneTask("1A_2").
    DonePhase("1A").
    DoneEpic().
    Execute()

Assert(result).
    EpicStatus("done").
    TestResult("T1A_1", "passing"). // Recovered
    TestResult("T1A_2", "passing").
    TestResult("T1A_3", "passing").
    EventSequence([]string{"test_failed", "test_passed", "test_passed", "test_passed"}).
    NoErrors().
    MustPass()
```

#### Scenario 16: Memory Isolation and Concurrent Execution
**Description:** Validate that multiple epic executions don't interfere with each other
```go
// Create two independent epics
epic1 := EpicBuilder("concurrent-016a").
    WithPhase("1A", "Development", "pending").
    WithTask("1A_1", "1A", "Feature A", "pending").
    WithTest("T1A_1", "1A_1", "1A", "Test A", "pending").
    Build()

epic2 := EpicBuilder("concurrent-016b").
    WithPhase("1B", "Testing", "pending").
    WithTask("1B_1", "1B", "Feature B", "pending").
    WithTest("T1B_1", "1B_1", "1B", "Test B", "pending").
    Build()

// Create separate environments
env1 := CreateTestEnvironment(epic1)
env2 := CreateTestEnvironment(epic2)

// Execute both concurrently (simulated)
result1 := TransitionChain(env1).
    StartEpic().
    StartPhase("1A").
    StartTask("1A_1").
    PassTest("T1A_1").
    DoneTask("1A_1").
    DonePhase("1A").
    DoneEpic().
    Execute()

result2 := TransitionChain(env2).
    StartEpic().
    StartPhase("1B").
    StartTask("1B_1").
    PassTest("T1B_1").
    DoneTask("1B_1").
    DonePhase("1B").
    DoneEpic().
    Execute()

// Validate both completed independently
Assert(result1).
    EpicStatus("done").
    PhaseStatus("1A", "done").
    NoErrors().
    MustPass()

Assert(result2).
    EpicStatus("done").
    PhaseStatus("1B", "done").
    NoErrors().
    CustomAssertion("isolation_check", func(r *executor.TransitionChainResult) error {
        // Verify that result2 doesn't contain any events/data from result1
        if r.FinalState.ID == "concurrent-016a" {
            return fmt.Errorf("memory isolation failed - wrong epic ID in result2")
        }
        for _, event := range r.FinalState.Events {
            if strings.Contains(event.Data, "Feature A") || strings.Contains(event.Data, "1A") {
                return fmt.Errorf("memory isolation failed - result2 contains result1 data")
            }
        }
        return nil
    }).
    MustPass()
```

## Data Model

### Scenario Definition Structure
```go
type TransitionScenario struct {
    ID          string
    Name        string
    Category    ScenarioCategory // Simple, Medium, Complex
    Description string
    EpicBuilder func() *EpicBuilder
    Execution   func(*TestExecutionEnvironment) (*TransitionChainResult, error)
    Assertions  func(*TransitionChainResult) error
    Metadata    ScenarioMetadata
}

type ScenarioCategory string
const (
    Simple   ScenarioCategory = "simple"
    Medium   ScenarioCategory = "medium"
    Complex  ScenarioCategory = "complex"
)

type ScenarioMetadata struct {
    ExpectedDuration  time.Duration
    MemoryExpectation int64
    EventCount        int
    CommandCount      int
    ValidationRules   []string
}
```

### Test Execution Framework
```go
type ScenarioExecutor struct {
    scenarios []TransitionScenario
    results   map[string]*ScenarioResult
    config    ExecutorConfig
}

type ScenarioResult struct {
    Scenario      *TransitionScenario
    ExecutionTime time.Duration
    MemoryUsage   int64
    Success       bool
    Errors        []error
    Performance   PerformanceMetrics
}
```

## Acceptance Criteria

### AC-1: Simple Scenarios (1-6)
- **GIVEN** any simple scenario (1-6)
- **WHEN** I execute the scenario using the builder pattern
- **THEN** it should complete successfully within 100ms with no errors

### AC-2: Medium Complexity Scenarios (7-12)
- **GIVEN** any medium complexity scenario (7-12)
- **WHEN** I execute the scenario with multiple phases/tasks
- **THEN** it should handle all state transitions correctly and complete within 500ms

### AC-3: Complex Edge Cases (13-16)
- **GIVEN** any complex scenario (13-16)
- **WHEN** I execute scenarios with validation failures or edge cases
- **THEN** errors should be properly captured and state should remain consistent

### AC-4: Memory Isolation
- **GIVEN** concurrent execution of scenarios 16a and 16b
- **WHEN** both scenarios run simultaneously
- **THEN** neither scenario should affect the other's state or results

### AC-5: Performance Benchmarks
- **GIVEN** all 16 scenarios
- **WHEN** executed as a test suite
- **THEN** total execution time should be < 5 seconds

### AC-6: Regression Testing
- **GIVEN** any scenario with snapshot testing
- **WHEN** executed multiple times
- **THEN** results should be consistent and match stored snapshots

### AC-7: Error Validation
- **GIVEN** scenarios 13-14 (validation failures)
- **WHEN** attempting invalid state transitions
- **THEN** specific validation errors should be thrown with helpful messages

### AC-8: Batch Operations
- **GIVEN** scenarios with multiple tests (3, 9, 11)
- **WHEN** executing batch test operations
- **THEN** all tests should be processed correctly without interference

## Implementation Phases

### Phase 15A: Simple Scenarios Implementation (Day 1)
- Implement scenarios 1-6 using EpicBuilder pattern
- Validate basic epic lifecycle workflows
- Test simple error cases and recovery
- Establish performance baselines

### Phase 15B: Medium Complexity Scenarios (Day 1-2)
- Implement scenarios 7-12 with multi-phase workflows
- Test complex state transitions and dependencies
- Validate timing and performance requirements
- Add snapshot testing capabilities

### Phase 15C: Complex Edge Cases (Day 2-3)
- Implement scenarios 13-16 with validation failures
- Test memory isolation and concurrent execution
- Validate error handling and recovery strategies
- Complete comprehensive test coverage


## Test Implementation Requirements

### Complete Test Function Naming Convention

All 16 scenarios must be implemented as Go test functions following this exact naming pattern:

**Simple Scenarios (1-6):**
- `TestEpic15_Scenario01_BasicEpicStartToCompletion(t *testing.T)`
- `TestEpic15_Scenario02_TestFailureAndRecovery(t *testing.T)`
- `TestEpic15_Scenario03_MultipleTestsInSingleTask(t *testing.T)`
- `TestEpic15_Scenario04_SequentialTasksInSinglePhase(t *testing.T)`
- `TestEpic15_Scenario05_ParallelTestExecution(t *testing.T)`
- `TestEpic15_Scenario06_TimeBasedTransitions(t *testing.T)`

**Medium Complexity Scenarios (7-12):**
- `TestEpic15_Scenario07_MultiPhaseEpicWithDependencies(t *testing.T)`
- `TestEpic15_Scenario08_MixedTestResultsWithRecovery(t *testing.T)`
- `TestEpic15_Scenario09_BatchTestOperations(t *testing.T)`
- `TestEpic15_Scenario10_ComplexStateTransitionsWithAssertions(t *testing.T)`
- `TestEpic15_Scenario11_PerformanceAndTimingValidation(t *testing.T)`
- `TestEpic15_Scenario12_SnapshotAndRegressionTesting(t *testing.T)`

**Complex Edge Cases (13-16):**
- `TestEpic15_Scenario13_ValidationFailureTaskCompletionBlocked(t *testing.T)`
- `TestEpic15_Scenario14_PhaseCompletionBlockedByPendingTasks(t *testing.T)`
- `TestEpic15_Scenario15_TestCancellationAndRecovery(t *testing.T)`
- `TestEpic15_Scenario16_MemoryIsolationAndConcurrentExecution(t *testing.T)`

**Naming Rules:**
- Prefix: `TestEpic15_`
- Scenario number: `Scenario##_` (zero-padded for single digits)
- Descriptive name: `PascalCase` without spaces or special characters
- Function signature: `(t *testing.T)`

**File Organization:**
- All tests should be in `internal/testing/scenarios/epic15_scenarios_test.go`
- Each test function should include the complete 4-step pattern from FR-0
- Tests should be grouped by complexity category with clear comments

## Definition of Done

- [ ] All 16 scenarios implemented and passing with correct test function names
- [ ] Performance benchmarks met for all scenarios
- [ ] Memory isolation validated for concurrent execution
- [ ] Error scenarios properly validate EPIC 13 business rules
- [ ] Snapshot testing working for regression detection
- [ ] Test coverage > 95% for all scenario code
- [ ] Documentation includes usage examples and patterns
- [ ] Integration with existing AgentPM test framework
- [ ] CI/CD pipeline includes scenario execution
- [ ] Performance monitoring for scenario execution times

## Dependencies and Risks

### Dependencies
- **EPIC 14:** Transition Chain Testing Framework must be complete
- **EPIC 13:** Status validation rules must be implemented
- **Memory Storage:** Isolation capabilities must be functional
- **Snapshot Testing:** Framework must support comparison operations

### Risks
- **Medium Risk:** Performance requirements may not be achievable with complex scenarios
- **Medium Risk:** Memory isolation may not work properly with concurrent execution
- **Low Risk:** Some edge cases may require additional framework features

### Mitigation Strategies
- Performance profiling during development to identify bottlenecks
- Comprehensive testing of memory isolation with stress testing
- Incremental implementation starting with simple cases
- Regular validation against real AgentPM usage patterns

## Notes

- These scenarios should become the standard test suite for all AgentPM development
- Performance benchmarks should be monitored in CI/CD to detect regressions
- Scenarios can be extended for specific feature testing beyond this epic
- The framework should support easy addition of new scenarios as AgentPM evolves
- Consider automatic generation of scenarios from real usage patterns in the future