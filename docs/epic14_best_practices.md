# Epic 14 Framework - Best Practices for Complex Transition Testing

## Overview

This guide provides best practices, patterns, and recommendations for effectively using the Epic 14 Transition Chain Testing Framework in complex scenarios. Following these practices will help you write maintainable, reliable, and performant tests.

## Table of Contents

1. [General Testing Principles](#general-testing-principles)
2. [Test Organization and Structure](#test-organization-and-structure)
3. [Assertion Design Patterns](#assertion-design-patterns)
4. [Error Handling and Debugging](#error-handling-and-debugging)
5. [Performance Considerations](#performance-considerations)
6. [Complex Scenario Testing](#complex-scenario-testing)
7. [Maintenance and Refactoring](#maintenance-and-refactoring)
8. [Team Collaboration](#team-collaboration)

## General Testing Principles

### 1. Test Independence and Isolation

**✅ DO: Ensure test independence**
```go
func TestEpicCompletion(t *testing.T) {
    // Each test should create its own isolated test data
    result := executor.NewTransitionChain().
        StartEpic("isolated-test-epic").
        ExecutePhase("1A").
        CompleteEpic().
        Execute()
    
    assertions.Assert(result).
        EpicStatus("completed").
        MustPass()
}
```

**❌ DON'T: Share state between tests**
```go
// Avoid shared global state
var sharedEpic *epic.Epic

func TestEpicStart(t *testing.T) {
    sharedEpic = createEpic() // This affects other tests
    // ... test logic
}

func TestEpicCompletion(t *testing.T) {
    // This test depends on TestEpicStart running first
    result := processEpic(sharedEpic)
    // ... assertions
}
```

### 2. Clear Test Intent and Naming

**✅ DO: Use descriptive test names**
```go
func TestEpicWithMultiplePhasesCompletesInCorrectOrder(t *testing.T) { /* ... */ }
func TestEpicFailsWhenPhaseHasUncompletedTasks(t *testing.T) { /* ... */ }
func TestEventSequenceValidationForComplexWorkflow(t *testing.T) { /* ... */ }
```

**❌ DON'T: Use vague or generic names**
```go
func TestEpic(t *testing.T) { /* ... */ }
func TestBasic(t *testing.T) { /* ... */ }
func TestScenario1(t *testing.T) { /* ... */ }
```

### 3. Comprehensive but Focused Assertions

**✅ DO: Assert on specific, relevant conditions**
```go
func TestPhaseTransitionRules(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("transition-rules-epic").
        ExecutePhase("1A").
        ExecutePhase("1B").
        Execute()
    
    assertions.Assert(result).
        PhaseStatus("1A", "completed").
        PhaseStatus("1B", "completed").
        EventSequence([]string{
            "epic_started",
            "phase_1A_started",
            "phase_1A_completed",
            "phase_1B_started",
            "phase_1B_completed",
        }).
        NoErrors().
        MustPass()
}
```

**❌ DON'T: Assert on irrelevant details**
```go
func TestPhaseTransition(t *testing.T) {
    result := executePhaseTransition()
    
    // Too many irrelevant assertions make tests brittle
    assertions.Assert(result).
        EpicStatus("completed").
        PhaseStatus("1A", "completed").
        TaskStatus("1A_1", "completed").
        TaskStatus("1A_2", "completed").
        TaskStatus("1A_3", "completed").
        HasEvent("epic_started").
        HasEvent("task_started").
        HasEvent("task_completed").
        EventCount(15).              // Brittle - exact count may change
        ExecutionTime(500*time.Millisecond). // Too restrictive timing
        MustPass()
}
```

## Test Organization and Structure

### 1. Logical Test Grouping

**✅ DO: Group related tests in table-driven patterns**
```go
func TestEpicStateTransitions(t *testing.T) {
    testCases := []struct {
        name           string
        operations     func(*executor.TransitionChain) *executor.TransitionChain
        expectedStatus string
        expectedPhases map[string]string
    }{
        {
            name: "single_phase_completion",
            operations: func(chain *executor.TransitionChain) *executor.TransitionChain {
                return chain.StartEpic("single-phase").
                    ExecutePhase("1A").
                    CompleteEpic()
            },
            expectedStatus: "completed",
            expectedPhases: map[string]string{"1A": "completed"},
        },
        {
            name: "multi_phase_progression",
            operations: func(chain *executor.TransitionChain) *executor.TransitionChain {
                return chain.StartEpic("multi-phase").
                    ExecutePhase("1A").
                    ExecutePhase("1B").
                    CompleteEpic()
            },
            expectedStatus: "completed",
            expectedPhases: map[string]string{
                "1A": "completed",
                "1B": "completed",
            },
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := tc.operations(executor.NewTransitionChain()).Execute()
            
            builder := assertions.Assert(result).EpicStatus(tc.expectedStatus)
            
            for phaseID, expectedStatus := range tc.expectedPhases {
                builder = builder.PhaseStatus(phaseID, expectedStatus)
            }
            
            builder.MustPass()
        })
    }
}
```

### 2. Helper Functions for Complex Setup

**✅ DO: Create reusable test setup helpers**
```go
// Test data builders
func createComplexEpicBuilder() *EpicBuilder {
    return NewEpicBuilder().
        WithPhases([]string{"1A", "1B", "2A", "2B"}).
        WithTasks(map[string][]string{
            "1A": {"1A_1", "1A_2"},
            "1B": {"1B_1", "1B_2", "1B_3"},
            "2A": {"2A_1"},
            "2B": {"2B_1", "2B_2"},
        }).
        WithDependencies(map[string][]string{
            "1B": {"1A"},
            "2A": {"1A", "1B"},
            "2B": {"2A"},
        })
}

// Assertion helpers
func assertSuccessfulCompletion(t *testing.T, result *executor.TransitionChainResult) {
    assertions.Assert(result).
        EpicStatus("completed").
        NoErrors().
        ExecutionTime(30 * time.Second).
        MustPass()
}

func assertPhaseProgression(t *testing.T, result *executor.TransitionChainResult, phases []string) {
    builder := assertions.Assert(result)
    
    for _, phaseID := range phases {
        builder = builder.PhaseStatus(phaseID, "completed")
    }
    
    builder.MustPass()
}

// Usage in tests
func TestComplexEpicWorkflow(t *testing.T) {
    epic := createComplexEpicBuilder().Build()
    
    result := executor.NewTransitionChain().
        LoadEpic(epic).
        ExecuteFullWorkflow().
        Execute()
    
    assertSuccessfulCompletion(t, result)
    assertPhaseProgression(t, result, []string{"1A", "1B", "2A", "2B"})
}
```

### 3. Test Categories and Tagging

**✅ DO: Use build tags for different test categories**
```go
//go:build integration
// +build integration

package tests

// Integration tests that require external dependencies
func TestExternalServiceIntegration(t *testing.T) {
    // ... integration test logic
}
```

```go
//go:build performance
// +build performance

package tests

// Performance tests that take longer to run
func TestLargeScaleEpicPerformance(t *testing.T) {
    // ... performance test logic
}
```

## Assertion Design Patterns

### 1. Fluent Assertion Chaining

**✅ DO: Chain related assertions logically**
```go
func TestEpicLifecycle(t *testing.T) {
    result := executeEpicLifecycle()
    
    assertions.Assert(result).
        // Epic-level assertions
        EpicStatus("completed").
        ExecutionTime(10 * time.Second).
        
        // Phase-level assertions
        PhaseStatus("1A", "completed").
        PhaseStatus("1B", "completed").
        
        // Task-level assertions
        TaskStatus("1A_1", "completed").
        TaskStatus("1A_2", "completed").
        
        // Event assertions
        HasEvent("epic_started").
        HasEvent("epic_completed").
        EventSequence([]string{"epic_started", "phase_started", "phase_completed", "epic_completed"}).
        
        // Error assertions
        NoErrors().
        
        MustPass()
}
```

### 2. Conditional Assertions

**✅ DO: Use conditional logic for complex scenarios**
```go
func TestConditionalValidation(t *testing.T) {
    result := executeConditionalWorkflow()
    
    builder := assertions.Assert(result).EpicStatus("completed")
    
    // Conditional assertions based on test environment
    if isProductionTest() {
        builder = builder.ExecutionTime(5 * time.Second)
    } else {
        builder = builder.ExecutionTime(30 * time.Second)
    }
    
    // Conditional assertions based on result state
    if hasAdvancedFeatures(result) {
        builder = builder.HasEvent("advanced_feature_enabled")
    }
    
    builder.MustPass()
}
```

### 3. Custom Assertion Patterns

**✅ DO: Create domain-specific assertions**
```go
// Domain-specific assertion helpers
func AssertBusinessRuleCompliance(result *executor.TransitionChainResult) *assertions.AssertionBuilder {
    return assertions.Assert(result).
        CustomAssertion("approval_workflow", func(result *executor.TransitionChainResult) error {
            // Check that approval workflow was followed
            approvalEvents := []string{"approval_requested", "approval_granted"}
            for _, requiredEvent := range approvalEvents {
                found := false
                for _, event := range result.FinalState.Events {
                    if event.Type == requiredEvent {
                        found = true
                        break
                    }
                }
                if !found {
                    return fmt.Errorf("missing required approval event: %s", requiredEvent)
                }
            }
            return nil
        }).
        CustomAssertion("compliance_audit_trail", func(result *executor.TransitionChainResult) error {
            // Verify audit trail requirements
            return validateAuditTrail(result.FinalState.Events)
        })
}

func TestBusinessProcess(t *testing.T) {
    result := executeBusinessProcess()
    
    AssertBusinessRuleCompliance(result).
        EpicStatus("completed").
        NoErrors().
        MustPass()
}
```

## Error Handling and Debugging

### 1. Strategic Debug Mode Usage

**✅ DO: Use appropriate debug levels**
```go
func TestComplexWorkflowDebugging(t *testing.T) {
    result := executeComplexWorkflow()
    
    // Use verbose debugging for complex tests during development
    builder := assertions.Assert(result).
        WithDebugMode(assertions.DebugVerbose).
        EnableStateVisualization()
    
    if testing.Verbose() {
        builder = builder.PrintDebugInfo()
    }
    
    builder.
        EpicStatus("completed").
        MustPass()
}

func TestProductionScenario(t *testing.T) {
    result := executeProductionScenario()
    
    // Use minimal debugging for production-like tests
    assertions.Assert(result).
        WithDebugMode(assertions.DebugBasic).
        EpicStatus("completed").
        MustPass()
}
```

### 2. Graceful Error Handling

**✅ DO: Use Check() for graceful error handling**
```go
func TestWithGracefulErrorHandling(t *testing.T) {
    result := executeRiskyOperation()
    
    err := assertions.Assert(result).
        EpicStatus("completed").
        NoErrors().
        Check()
    
    if err != nil {
        // Custom error handling and reporting
        if assertionErr, ok := err.(assertions.AssertionError); ok {
            t.Logf("Assertion failed:")
            t.Logf("  Expected: %v", assertionErr.Expected)
            t.Logf("  Actual: %v", assertionErr.Actual)
            t.Logf("  Context: %v", assertionErr.Context)
            t.Logf("  Suggestions: %v", assertionErr.Suggestions)
        }
        
        // Additional debugging information
        if result.FinalState != nil {
            t.Logf("Final epic status: %s", result.FinalState.Status)
            t.Logf("Execution time: %v", result.ExecutionTime)
        }
        
        t.Fatal(err)
    }
}
```

### 3. Recovery Strategy Implementation

**✅ DO: Implement custom recovery strategies for complex scenarios**
```go
func TestWithRecoveryStrategy(t *testing.T) {
    recoveryStrategy := &assertions.RecoveryStrategy{
        CanRecover: func(err error) bool {
            // Only recover from transient errors
            return strings.Contains(err.Error(), "transient") ||
                   strings.Contains(err.Error(), "timeout")
        },
        RecoverFunc: func(err error, ctx *assertions.ErrorContext) error {
            // Implement recovery logic
            t.Logf("Attempting recovery from error: %v", err)
            
            // Could implement retry logic, state cleanup, etc.
            return nil
        },
        ContinueFunc: func(ctx *assertions.ErrorContext) bool {
            // Decide whether to continue after recovery
            return ctx.Stage != "critical_failure"
        },
    }
    
    result := executeWithPotentialFailures()
    
    assertions.Assert(result).
        WithRecoveryStrategy(recoveryStrategy).
        RecoverFromErrors().
        EpicStatus("completed").
        MustPass()
}
```

## Performance Considerations

### 1. Efficient Test Execution

**✅ DO: Use batch assertions for repeated patterns**
```go
func TestMultipleEpicsEfficiently(t *testing.T) {
    const numEpics = 100
    
    // Prepare batch assertions
    statusAssertions := []func(*assertions.AssertionBuilder) *assertions.AssertionBuilder{
        func(ab *assertions.AssertionBuilder) *assertions.AssertionBuilder {
            return ab.EpicStatus("completed")
        },
        func(ab *assertions.AssertionBuilder) *assertions.AssertionBuilder {
            return ab.NoErrors()
        },
    }
    
    for i := 0; i < numEpics; i++ {
        result := executor.NewTransitionChain().
            StartEpic(fmt.Sprintf("batch-epic-%d", i)).
            ExecutePhase("1A").
            CompleteEpic().
            Execute()
        
        // Use batch assertions for efficiency
        assertions.Assert(result).
            BatchAssertions(statusAssertions).
            MustPass()
    }
}
```

### 2. Memory Management

**✅ DO: Be mindful of memory usage in large test suites**
```go
func TestLargeScaleMemoryManagement(t *testing.T) {
    const iterations = 1000
    
    var baselineMemory runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&baselineMemory)
    
    for i := 0; i < iterations; i++ {
        result := executeEpicWorkflow(i)
        
        // Use DebugOff for performance tests
        assertions.Assert(result).
            WithDebugMode(assertions.DebugOff).
            EpicStatus("completed").
            MustPass()
        
        // Periodic memory checks
        if i%100 == 99 {
            runtime.GC()
            var currentMemory runtime.MemStats
            runtime.ReadMemStats(&currentMemory)
            
            memoryGrowth := int64(currentMemory.Alloc) - int64(baselineMemory.Alloc)
            if memoryGrowth > 100*1024*1024 { // 100MB threshold
                t.Errorf("Memory growth too large after %d iterations: %d bytes", i+1, memoryGrowth)
            }
        }
        
        // Explicit cleanup for large objects
        result = nil
    }
}
```

### 3. Parallel Test Execution

**✅ DO: Design tests for parallel execution**
```go
func TestParallelExecution(t *testing.T) {
    testCases := []struct {
        name   string
        epicID string
        phases []string
    }{
        {"workflow_a", "epic-a", []string{"1A", "1B"}},
        {"workflow_b", "epic-b", []string{"1A", "2A"}},
        {"workflow_c", "epic-c", []string{"1A", "1B", "2A"}},
    }
    
    for _, tc := range testCases {
        tc := tc // Capture loop variable
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel() // Enable parallel execution
            
            result := executor.NewTransitionChain().
                StartEpic(tc.epicID)
            
            for _, phase := range tc.phases {
                result = result.ExecutePhase(phase)
            }
            
            finalResult := result.CompleteEpic().Execute()
            
            assertions.Assert(finalResult).
                EpicStatus("completed").
                MustPass()
        })
    }
}
```

## Complex Scenario Testing

### 1. State Machine Testing

**✅ DO: Test state transitions comprehensively**
```go
func TestStateMachineTransitions(t *testing.T) {
    testCases := []struct {
        name           string
        initialState   string
        transitions    []string
        expectedFinal  string
        expectedPath   []string
    }{
        {
            name:          "normal_progression",
            initialState:  "pending",
            transitions:   []string{"start", "activate", "complete"},
            expectedFinal: "completed",
            expectedPath:  []string{"pending", "active", "in_progress", "completed"},
        },
        {
            name:          "error_recovery",
            initialState:  "pending",
            transitions:   []string{"start", "fail", "recover", "complete"},
            expectedFinal: "completed",
            expectedPath:  []string{"pending", "active", "failed", "recovering", "completed"},
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            chain := executor.NewTransitionChain().
                WithInitialState(tc.initialState)
            
            for _, transition := range tc.transitions {
                chain = chain.ExecuteTransition(transition)
            }
            
            result := chain.Execute()
            
            assertions.Assert(result).
                EpicStatus(tc.expectedFinal).
                StateProgression(tc.expectedPath).
                MustPass()
        })
    }
}
```

### 2. Dependency Testing

**✅ DO: Test complex dependency scenarios**
```go
func TestComplexDependencies(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("dependency-test").
        WithDependencies(map[string][]string{
            "1B": {"1A"},
            "2A": {"1A", "1B"},
            "2B": {"2A"},
        }).
        ExecutePhase("1A").
        ExecutePhase("1B").
        ExecutePhase("2A").
        ExecutePhase("2B").
        CompleteEpic().
        Execute()
    
    assertions.Assert(result).
        CustomAssertion("dependency_validation", func(result *executor.TransitionChainResult) error {
            // Validate that dependencies were respected
            events := result.FinalState.Events
            
            phase1AStart := findEventTime(events, "phase_1A_started")
            phase1BStart := findEventTime(events, "phase_1B_started")
            phase2AStart := findEventTime(events, "phase_2A_started")
            
            if phase1BStart.Before(phase1AStart) {
                return fmt.Errorf("1B started before 1A")
            }
            
            if phase2AStart.Before(phase1BStart) {
                return fmt.Errorf("2A started before 1B completed")
            }
            
            return nil
        }).
        EpicStatus("completed").
        MustPass()
}
```

### 3. Event-Driven Architecture Testing

**✅ DO: Test event-driven workflows thoroughly**
```go
func TestEventDrivenWorkflow(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("event-driven-epic").
        WithEventTriggers(map[string]string{
            "approval_granted": "start_development",
            "development_complete": "start_testing",
            "testing_complete": "start_deployment",
        }).
        TriggerEvent("approval_granted").
        WaitForEvent("development_complete").
        TriggerEvent("testing_complete").
        CompleteEpic().
        Execute()
    
    assertions.Assert(result).
        EventSequence([]string{
            "epic_started",
            "approval_granted",
            "development_started",
            "development_complete",
            "testing_started",
            "testing_complete",
            "deployment_started",
            "epic_completed",
        }).
        CustomAssertion("event_causality", func(result *executor.TransitionChainResult) error {
            // Validate event causality chains
            return validateEventCausality(result.FinalState.Events)
        }).
        MustPass()
}
```

## Maintenance and Refactoring

### 1. Test Refactoring Patterns

**✅ DO: Extract common patterns into reusable components**
```go
// Base test class for epic testing
type EpicTestSuite struct {
    defaultTimeout time.Duration
    debugMode      assertions.DebugMode
}

func NewEpicTestSuite() *EpicTestSuite {
    return &EpicTestSuite{
        defaultTimeout: 30 * time.Second,
        debugMode:      assertions.DebugBasic,
    }
}

func (ets *EpicTestSuite) AssertEpicCompletion(t *testing.T, result *executor.TransitionChainResult) {
    assertions.Assert(result).
        WithDebugMode(ets.debugMode).
        EpicStatus("completed").
        NoErrors().
        ExecutionTime(ets.defaultTimeout).
        MustPass()
}

func (ets *EpicTestSuite) AssertPhaseProgression(t *testing.T, result *executor.TransitionChainResult, phases []string) {
    builder := assertions.Assert(result).WithDebugMode(ets.debugMode)
    
    for _, phaseID := range phases {
        builder = builder.PhaseStatus(phaseID, "completed")
    }
    
    builder.MustPass()
}

// Usage in tests
func TestWithSuite(t *testing.T) {
    suite := NewEpicTestSuite()
    
    result := executeEpicWorkflow()
    
    suite.AssertEpicCompletion(t, result)
    suite.AssertPhaseProgression(t, result, []string{"1A", "1B"})
}
```

### 2. Version Compatibility Testing

**✅ DO: Test backward compatibility**
```go
func TestBackwardCompatibility(t *testing.T) {
    versions := []string{"v1.0", "v1.1", "v2.0"}
    
    for _, version := range versions {
        t.Run(fmt.Sprintf("compatibility_%s", version), func(t *testing.T) {
            result := executor.NewTransitionChain().
                WithVersion(version).
                StartEpic("compatibility-test").
                ExecutePhase("1A").
                CompleteEpic().
                Execute()
            
            // Core functionality should work across versions
            assertions.Assert(result).
                EpicStatus("completed").
                NoErrors().
                MustPass()
        })
    }
}
```

## Team Collaboration

### 1. Shared Test Utilities

**✅ DO: Create shared test utilities for the team**
```go
// shared/testutils/epic_builders.go
package testutils

func NewStandardEpicBuilder() *EpicBuilder {
    return &EpicBuilder{
        phases: []string{"1A", "1B"},
        tasks:  map[string][]string{
            "1A": {"1A_1", "1A_2"},
            "1B": {"1B_1"},
        },
    }
}

func NewComplexEpicBuilder() *EpicBuilder {
    return &EpicBuilder{
        phases: []string{"1A", "1B", "2A", "2B"},
        // ... more complex setup
    }
}

// shared/testutils/assertions.go
package testutils

func AssertStandardWorkflow(t *testing.T, result *executor.TransitionChainResult) {
    assertions.Assert(result).
        EpicStatus("completed").
        PhaseStatus("1A", "completed").
        PhaseStatus("1B", "completed").
        NoErrors().
        MustPass()
}
```

### 2. Documentation and Examples

**✅ DO: Document complex test patterns**
```go
// TestComplexScenarioExample demonstrates testing a complex multi-stage
// workflow with dependencies, error handling, and performance validation.
//
// This test covers:
// - Multi-phase execution with dependencies
// - Event sequence validation
// - Error recovery scenarios
// - Performance benchmarking
// - Custom business rule validation
//
// Usage pattern:
//   result := executor.NewTransitionChain().
//       StartEpic("example-epic").
//       ExecutePhase("phase1").
//       ExecutePhase("phase2").
//       CompleteEpic().
//       Execute()
//
//   assertions.Assert(result).
//       EpicStatus("completed").
//       CustomAssertion("business_rules", validateBusinessRules).
//       MustPass()
func TestComplexScenarioExample(t *testing.T) {
    // Test implementation with detailed comments
    // ...
}
```

### 3. Code Review Guidelines

**✅ DO: Follow code review best practices**

#### Code Review Checklist for Epic 14 Tests:

- [ ] **Test Independence**: Tests don't depend on external state or other tests
- [ ] **Clear Intent**: Test name and structure clearly communicate what's being tested
- [ ] **Appropriate Assertions**: Assertions are relevant and not overly brittle
- [ ] **Error Handling**: Proper error handling and debugging setup
- [ ] **Performance**: Memory usage and execution time are reasonable
- [ ] **Maintainability**: Code is readable and follows team conventions
- [ ] **Coverage**: Test covers both happy path and error scenarios
- [ ] **Documentation**: Complex logic is documented with comments

## Anti-Patterns to Avoid

### ❌ DON'T: Over-assert on implementation details
```go
// Bad: Too many implementation-specific assertions
func TestOverAssertion(t *testing.T) {
    result := executeWorkflow()
    
    assertions.Assert(result).
        EpicStatus("completed").
        PhaseStatus("1A", "completed").
        TaskStatus("1A_1", "completed").
        TaskStatus("1A_2", "completed").
        EventCount(47).  // Brittle - exact count may change
        ExecutionTime(1234*time.Millisecond).  // Too precise
        HasEvent("internal_state_change").     // Implementation detail
        MustPass()
}
```

### ❌ DON'T: Create interdependent tests
```go
// Bad: Tests depend on each other
func TestStep1(t *testing.T) {
    globalState.epic = createEpic()
    // ... test logic
}

func TestStep2(t *testing.T) {
    // Depends on TestStep1 running first
    continueEpic(globalState.epic)
    // ... test logic
}
```

### ❌ DON'T: Ignore performance in test design
```go
// Bad: Inefficient test that will slow down CI
func TestInefficient(t *testing.T) {
    for i := 0; i < 10000; i++ {
        result := executor.NewTransitionChain().
            WithDebugMode(assertions.DebugVerbose). // Too verbose for loops
            EnableStateVisualization().              // Expensive in loops
            StartEpic(fmt.Sprintf("epic-%d", i)).
            ExecutePhase("1A").
            CompleteEpic().
            Execute()
        
        assertions.Assert(result).
            MatchSnapshot(fmt.Sprintf("epic-%d", i)). // Creates too many snapshots
            PrintDebugInfo().                         // Prints for every iteration
            MustPass()
    }
}
```

By following these best practices, your Epic 14 tests will be more reliable, maintainable, and provide better debugging information when failures occur.