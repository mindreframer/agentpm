# Epic 14 Framework - Usage Examples and Patterns

## Table of Contents

1. [Basic Usage Patterns](#basic-usage-patterns)
2. [Advanced Testing Scenarios](#advanced-testing-scenarios)
3. [Common Testing Patterns](#common-testing-patterns)
4. [Error Handling Examples](#error-handling-examples)
5. [Performance Testing](#performance-testing)
6. [Snapshot Testing](#snapshot-testing)
7. [Integration Testing](#integration-testing)
8. [Custom Validation Patterns](#custom-validation-patterns)

## Basic Usage Patterns

### Simple Epic Completion Test

```go
func TestEpicCompletion(t *testing.T) {
    // Execute the transition chain
    result := executor.NewTransitionChain().
        StartEpic("test-epic").
        ExecutePhase("1A").
        CompleteEpic().
        Execute()
    
    // Validate the results
    assertions.Assert(result).
        EpicStatus("completed").
        NoErrors().
        MustPass()
}
```

### Phase-by-Phase Validation

```go
func TestPhaseProgression(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("multi-phase-epic").
        ExecutePhase("1A").
        ExecutePhase("1B").
        ExecutePhase("2A").
        CompleteEpic().
        Execute()
    
    assertions.Assert(result).
        PhaseStatus("1A", "completed").
        PhaseStatus("1B", "completed").
        PhaseStatus("2A", "completed").
        EventSequence([]string{
            "epic_started",
            "phase_1A_started",
            "phase_1A_completed",
            "phase_1B_started", 
            "phase_1B_completed",
            "phase_2A_started",
            "phase_2A_completed",
            "epic_completed",
        }).
        MustPass()
}
```

### Task-Level Testing

```go
func TestTaskExecution(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("task-heavy-epic").
        ExecutePhase("1A").
        Execute()
    
    assertions.Assert(result).
        TaskStatus("1A_1", "completed").
        TaskStatus("1A_2", "completed").
        TaskStatus("1A_3", "pending").    // Not yet started
        EventCount(15).                   // Expected number of task events
        ExecutionTime(30 * time.Second).  // Should complete quickly
        MustPass()
}
```

## Advanced Testing Scenarios

### Complex State Transition Testing

```go
func TestComplexStateTransitions(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("complex-epic").
        ExecuteCommand("initialize_system").
        ExecutePhase("1A").
        ExecuteCommand("validate_phase_1a").
        ExecutePhase("1B").
        ExecuteCommand("checkpoint_save").
        ExecutePhase("2A").
        ExecuteCommand("finalize_system").
        CompleteEpic().
        Execute()
    
    assertions.Assert(result).
        WithDebugMode(assertions.DebugVerbose).
        StateProgression([]string{
            "pending",
            "initializing", 
            "wip",
            "validating",
            "wip",
            "checkpointing",
            "wip",
            "finalizing",
            "completed",
        }).
        IntermediateState(2, func(epic *epic.Epic) error {
            // Validate intermediate state after phase 1A
            if epic.Status != "wip" {
                return fmt.Errorf("expected active status, got %s", epic.Status)
            }
            if len(epic.Events) < 5 {
                return fmt.Errorf("expected at least 5 events, got %d", len(epic.Events))
            }
            return nil
        }).
        IntermediateState(6, func(epic *epic.Epic) error {
            // Validate checkpoint was created
            for _, event := range epic.Events {
                if event.Type == "checkpoint_created" {
                    return nil
                }
            }
            return fmt.Errorf("checkpoint_created event not found")
        }).
        MustPass()
}
```

### Timing and Performance Validation

```go
func TestTimingConstraints(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("time-sensitive-epic").
        ExecutePhase("1A").
        ExecutePhase("1B").
        CompleteEpic().
        Execute()
    
    assertions.Assert(result).
        ExecutionTime(5 * time.Second).               // Total execution time
        PhaseTransitionTiming("1A", 2 * time.Second). // Phase-specific timing
        PhaseTransitionTiming("1B", 1 * time.Second).
        PerformanceBenchmark(5*time.Second, 100).     // Max 100MB memory
        MustPass()
}
```

### Error Recovery Testing

```go
func TestErrorRecovery(t *testing.T) {
    // Configure recovery strategy
    recoveryStrategy := &assertions.RecoveryStrategy{
        CanRecover: func(err error) bool {
            return strings.Contains(err.Error(), "recoverable")
        },
        RecoverFunc: func(err error, ctx *assertions.ErrorContext) error {
            // Implement recovery logic
            return nil
        },
        ContinueFunc: func(ctx *assertions.ErrorContext) bool {
            return true
        },
    }
    
    result := executor.NewTransitionChain().
        StartEpic("error-prone-epic").
        ExecutePhase("1A").
        InjectError("recoverable_error").
        ExecutePhase("1B").
        CompleteEpic().
        Execute()
    
    assertions.Assert(result).
        WithRecoveryStrategy(recoveryStrategy).
        RecoverFromErrors().
        EpicStatus("completed").
        HasErrors().                    // Should have errors but recover
        Check() // Use Check() instead of MustPass() for graceful error handling
    
    if err := assertion.Check(); err != nil {
        t.Fatalf("Failed after recovery: %v", err)
    }
}
```

## Common Testing Patterns

### Test Suite Pattern

```go
func TestEpicTestSuite(t *testing.T) {
    testCases := []struct {
        name           string
        epicID         string
        phases         []string
        expectedStatus string
        expectedEvents int
    }{
        {
            name:           "simple_epic",
            epicID:         "simple-test",
            phases:         []string{"1A"},
            expectedStatus: "completed",
            expectedEvents: 3,
        },
        {
            name:           "complex_epic", 
            epicID:         "complex-test",
            phases:         []string{"1A", "1B", "2A"},
            expectedStatus: "completed",
            expectedEvents: 9,
        },
        {
            name:           "failing_epic",
            epicID:         "failing-test",
            phases:         []string{"1A", "INVALID"},
            expectedStatus: "failed",
            expectedEvents: 5,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            chain := executor.NewTransitionChain().StartEpic(tc.epicID)
            
            for _, phase := range tc.phases {
                chain = chain.ExecutePhase(phase)
            }
            
            result := chain.CompleteEpic().Execute()
            
            builder := assertions.Assert(result).
                EpicStatus(tc.expectedStatus).
                EventCount(tc.expectedEvents)
            
            if tc.expectedStatus == "completed" {
                builder.NoErrors()
            } else {
                builder.HasErrors()
            }
            
            builder.MustPass()
        })
    }
}
```

### Batch Assertion Pattern

```go
func TestBatchAssertions(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("batch-test-epic").
        ExecutePhase("1A").
        ExecutePhase("1B").
        CompleteEpic().
        Execute()
    
    // Define reusable assertion sets
    statusAssertions := []func(*assertions.AssertionBuilder) *assertions.AssertionBuilder{
        func(ab *assertions.AssertionBuilder) *assertions.AssertionBuilder {
            return ab.EpicStatus("completed")
        },
        func(ab *assertions.AssertionBuilder) *assertions.AssertionBuilder {
            return ab.PhaseStatus("1A", "completed")
        },
        func(ab *assertions.AssertionBuilder) *assertions.AssertionBuilder {
            return ab.PhaseStatus("1B", "completed")
        },
    }
    
    eventAssertions := []func(*assertions.AssertionBuilder) *assertions.AssertionBuilder{
        func(ab *assertions.AssertionBuilder) *assertions.AssertionBuilder {
            return ab.HasEvent("epic_started")
        },
        func(ab *assertions.AssertionBuilder) *assertions.AssertionBuilder {
            return ab.HasEvent("epic_completed")
        },
        func(ab *assertions.AssertionBuilder) *assertions.AssertionBuilder {
            return ab.EventCount(8)
        },
    }
    
    // Execute batch assertions
    assertions.Assert(result).
        BatchAssertions(statusAssertions).
        BatchAssertions(eventAssertions).
        NoErrors().
        MustPass()
}
```

## Error Handling Examples

### Negative Testing

```go
func TestEpicFailureScenarios(t *testing.T) {
    tests := []struct {
        name     string
        scenario func() *executor.TransitionChainResult
        validate func(*assertions.AssertionBuilder)
    }{
        {
            name: "invalid_phase",
            scenario: func() *executor.TransitionChainResult {
                return executor.NewTransitionChain().
                    StartEpic("test-epic").
                    ExecutePhase("INVALID_PHASE").
                    Execute()
            },
            validate: func(ab *assertions.AssertionBuilder) {
                ab.EpicStatus("failed").
                    HasErrors().
                    ErrorCount(1)
            },
        },
        {
            name: "timeout_scenario",
            scenario: func() *executor.TransitionChainResult {
                return executor.NewTransitionChain().
                    StartEpic("slow-epic").
                    WithTimeout(1 * time.Second).
                    ExecutePhase("1A").
                    Execute()
            },
            validate: func(ab *assertions.AssertionBuilder) {
                ab.HasErrors().
                    ExecutionTime(2 * time.Second) // Should have timed out
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.scenario()
            builder := assertions.Assert(result).WithDebugMode(assertions.DebugBasic)
            tt.validate(builder)
            builder.MustPass()
        })
    }
}
```

### Custom Error Validation

```go
func TestCustomErrorValidation(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("validation-epic").
        ExecutePhase("1A").
        CompleteEpic().
        Execute()
    
    assertions.Assert(result).
        CustomAssertion("dependency_validation", func(result *executor.TransitionChainResult) error {
            // Check that all dependencies were properly initialized
            for _, event := range result.FinalState.Events {
                if event.Type == "dependency_error" {
                    return fmt.Errorf("dependency validation failed: %s", event.Data)
                }
            }
            return nil
        }).
        CustomAssertion("resource_cleanup", func(result *executor.TransitionChainResult) error {
            // Verify resources were properly cleaned up
            for _, event := range result.FinalState.Events {
                if event.Type == "resource_leaked" {
                    return fmt.Errorf("resource leak detected: %s", event.Data)
                }
            }
            return nil
        }).
        MustPass()
}
```

## Performance Testing

### Load Testing Pattern

```go
func TestConcurrentEpicExecution(t *testing.T) {
    const numConcurrentEpics = 10
    
    var wg sync.WaitGroup
    results := make(chan *executor.TransitionChainResult, numConcurrentEpics)
    
    // Launch concurrent epic executions
    for i := 0; i < numConcurrentEpics; i++ {
        wg.Add(1)
        go func(epicIndex int) {
            defer wg.Done()
            
            result := executor.NewTransitionChain().
                StartEpic(fmt.Sprintf("concurrent-epic-%d", epicIndex)).
                ExecutePhase("1A").
                CompleteEpic().
                Execute()
            
            results <- result
        }(i)
    }
    
    // Wait for all to complete
    go func() {
        wg.Wait()
        close(results)
    }()
    
    // Validate all results
    var allResults []*executor.TransitionChainResult
    for result := range results {
        allResults = append(allResults, result)
        
        // Each epic should complete successfully
        assertions.Assert(result).
            EpicStatus("completed").
            NoErrors().
            ExecutionTime(5 * time.Second).
            MustPass()
    }
    
    // Validate overall performance
    if len(allResults) != numConcurrentEpics {
        t.Fatalf("Expected %d results, got %d", numConcurrentEpics, len(allResults))
    }
}
```

### Memory Usage Testing

```go
func TestMemoryUsagePattern(t *testing.T) {
    const iterations = 100
    
    var baselineMemory runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&baselineMemory)
    
    for i := 0; i < iterations; i++ {
        result := executor.NewTransitionChain().
            StartEpic(fmt.Sprintf("memory-test-%d", i)).
            ExecutePhase("1A").
            CompleteEpic().
            Execute()
        
        assertions.Assert(result).
            EpicStatus("completed").
            PerformanceBenchmark(1*time.Second, 50). // Max 50MB per iteration
            MustPass()
        
        // Force garbage collection every 10 iterations
        if i%10 == 9 {
            runtime.GC()
        }
    }
    
    // Final memory check
    var finalMemory runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&finalMemory)
    
    memoryGrowth := int64(finalMemory.Alloc) - int64(baselineMemory.Alloc)
    maxExpectedGrowth := int64(100 * 1024 * 1024) // 100MB
    
    if memoryGrowth > maxExpectedGrowth {
        t.Errorf("Memory growth too large: %d bytes (max: %d)", 
            memoryGrowth, maxExpectedGrowth)
    }
}
```

## Snapshot Testing

### Basic Snapshot Testing

```go
func TestEpicStateSnapshot(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("snapshot-epic").
        ExecutePhase("1A").
        ExecutePhase("1B").
        CompleteEpic().
        Execute()
    
    assertions.Assert(result).
        MatchSnapshot("epic_completion_state").
        MatchSelectiveSnapshot("phase_status", []string{"id", "status", "completion_time"}).
        MustPass()
}
```

### Conditional Snapshot Updates

```go
func TestSnapshotWithConditions(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("conditional-epic").
        ExecutePhase("1A").
        CompleteEpic().
        Execute()
    
    // Only update snapshots in development
    updateSnapshots := os.Getenv("UPDATE_SNAPSHOTS") == "true"
    
    config := map[string]interface{}{
        "update_mode":       updateSnapshots,
        "cross_platform":    true,
        "ignore_timestamps": true,
    }
    
    assertions.Assert(result).
        MatchSnapshotWithConfig("conditional_epic_state", config).
        MustPass()
}
```

## Integration Testing

### Integration with Standard Go Testing

```go
func TestIntegrationWithStandardTesting(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("integration-epic").
        ExecutePhase("1A").
        CompleteEpic().
        Execute()
    
    // Use Check() for more control over error handling
    if err := assertions.Assert(result).
        EpicStatus("completed").
        NoErrors().
        Check(); err != nil {
        
        // Custom error handling
        t.Logf("Epic execution failed: %v", err)
        
        if assertionErr, ok := err.(assertions.AssertionError); ok {
            t.Logf("Expected: %v", assertionErr.Expected)
            t.Logf("Actual: %v", assertionErr.Actual)
            t.Logf("Suggestions: %v", assertionErr.Suggestions)
        }
        
        t.Fatal(err)
    }
}
```

### Testify Integration

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestWithTestify(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("testify-epic").
        ExecutePhase("1A").
        CompleteEpic().
        Execute()
    
    // Use require for critical assertions
    require.NotNil(t, result)
    require.NotNil(t, result.FinalState)
    
    // Use Epic 14 framework for domain-specific assertions
    err := assertions.Assert(result).
        EpicStatus("completed").
        NoErrors().
        Check()
    
    // Use assert for non-critical checks
    assert.NoError(t, err)
    assert.Equal(t, "completed", result.FinalState.Status)
    assert.True(t, len(result.FinalState.Events) > 0)
}
```

## Custom Validation Patterns

### Domain-Specific Assertions

```go
// Custom assertion helper for business logic
func AssertBusinessRules(result *executor.TransitionChainResult) *assertions.AssertionBuilder {
    return assertions.Assert(result).
        CustomAssertion("business_rule_validation", func(result *executor.TransitionChainResult) error {
            // Validate business-specific requirements
            epic := result.FinalState
            
            // Rule 1: All phases must have completion timestamps
            for _, phase := range epic.Phases {
                if phase.Status == "completed" && phase.CompletionTime.IsZero() {
                    return fmt.Errorf("phase %s completed but missing timestamp", phase.ID)
                }
            }
            
            // Rule 2: Certain events must occur in order
            requiredSequence := []string{"approval_requested", "approval_granted", "execution_started"}
            eventTypes := make([]string, len(epic.Events))
            for i, event := range epic.Events {
                eventTypes[i] = event.Type
            }
            
            if !containsSequence(eventTypes, requiredSequence) {
                return fmt.Errorf("required event sequence not found: %v", requiredSequence)
            }
            
            return nil
        })
}

func TestBusinessRuleValidation(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("business-epic").
        ExecuteCommand("request_approval").
        ExecuteCommand("grant_approval").
        ExecutePhase("1A").
        CompleteEpic().
        Execute()
    
    AssertBusinessRules(result).
        EpicStatus("completed").
        MustPass()
}

// Helper function to check sequence containment
func containsSequence(haystack, needle []string) bool {
    if len(needle) == 0 {
        return true
    }
    
    needleIndex := 0
    for _, item := range haystack {
        if item == needle[needleIndex] {
            needleIndex++
            if needleIndex == len(needle) {
                return true
            }
        }
    }
    return false
}
```

### Fluent Custom Builders

```go
// Custom builder for specific epic types
type DatabaseMigrationAssertions struct {
    *assertions.AssertionBuilder
}

func AssertDatabaseMigration(result *executor.TransitionChainResult) *DatabaseMigrationAssertions {
    return &DatabaseMigrationAssertions{
        AssertionBuilder: assertions.Assert(result),
    }
}

func (dma *DatabaseMigrationAssertions) MigrationCompleted() *DatabaseMigrationAssertions {
    dma.AssertionBuilder.CustomAssertion("migration_completed", func(result *executor.TransitionChainResult) error {
        // Check for migration completion events
        for _, event := range result.FinalState.Events {
            if event.Type == "migration_completed" {
                return nil
            }
        }
        return fmt.Errorf("migration_completed event not found")
    })
    return dma
}

func (dma *DatabaseMigrationAssertions) SchemaVersionUpdated(expectedVersion string) *DatabaseMigrationAssertions {
    dma.AssertionBuilder.CustomAssertion("schema_version", func(result *executor.TransitionChainResult) error {
        for _, event := range result.FinalState.Events {
            if event.Type == "schema_version_updated" {
                if event.Data == expectedVersion {
                    return nil
                }
                return fmt.Errorf("schema version mismatch: expected %s, got %s", 
                    expectedVersion, event.Data)
            }
        }
        return fmt.Errorf("schema_version_updated event not found")
    })
    return dma
}

func (dma *DatabaseMigrationAssertions) MustPass() {
    dma.AssertionBuilder.MustPass()
}

func TestDatabaseMigration(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("db-migration-epic").
        ExecutePhase("backup").
        ExecutePhase("migrate").
        ExecutePhase("verify").
        CompleteEpic().
        Execute()
    
    AssertDatabaseMigration(result).
        EpicStatus("completed").
        MigrationCompleted().
        SchemaVersionUpdated("2.1.0").
        NoErrors().
        MustPass()
}
```

These examples demonstrate the flexibility and power of the Epic 14 framework for testing complex state transitions in various scenarios.