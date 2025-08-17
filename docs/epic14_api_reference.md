# Epic 14 Transition Chain Testing Framework - API Reference

## Overview

The Epic 14 Transition Chain Testing Framework provides a comprehensive, fluent API for testing complex state transitions in agentpm workflows. It enables robust validation of epic lifecycles, phase transitions, task completions, and event sequences with advanced debugging and visualization capabilities.

## Core Concepts

### AssertionBuilder
The central component that provides a fluent interface for building and executing test assertions.

```go
import "github.com/mindreframer/agentpm/internal/testing/assertions"

// Create a new assertion builder
builder := assertions.NewAssertionBuilder(result)

// Or use the fluent entry point
builder := assertions.Assert(result)
```

### TransitionChainResult
The input data structure containing the state progression and execution context.

```go
type TransitionChainResult struct {
    FinalState      *epic.Epic
    Commands        []string
    ExecutionTime   time.Duration
    Errors          []error
    IntermediateStates []interface{}
}
```

## API Methods

### Basic State Assertions

#### EpicStatus
Validates the final epic status.

```go
func (ab *AssertionBuilder) EpicStatus(expectedStatus string) *AssertionBuilder
```

**Example:**
```go
assertions.Assert(result).
    EpicStatus("completed").
    MustPass()
```

**Parameters:**
- `expectedStatus`: Expected epic status ("pending", "active", "completed", "failed")

#### PhaseStatus
Validates a specific phase status.

```go
func (ab *AssertionBuilder) PhaseStatus(phaseID, expectedStatus string) *AssertionBuilder
```

**Example:**
```go
assertions.Assert(result).
    PhaseStatus("1A", "completed").
    PhaseStatus("1B", "active").
    MustPass()
```

**Parameters:**
- `phaseID`: Unique identifier for the phase
- `expectedStatus`: Expected phase status

#### TaskStatus
Validates a specific task status.

```go
func (ab *AssertionBuilder) TaskStatus(taskID, expectedStatus string) *AssertionBuilder
```

**Example:**
```go
assertions.Assert(result).
    TaskStatus("1A_1", "completed").
    TaskStatus("1A_2", "pending").
    MustPass()
```

#### TestStatus
Validates a specific test status within the epic structure.

```go
func (ab *AssertionBuilder) TestStatus(testID, expectedStatus string) *AssertionBuilder
func (ab *AssertionBuilder) TestStatusUnified(testID, expectedStatus string) *AssertionBuilder
func (ab *AssertionBuilder) TestResult(testID, expectedResult string) *AssertionBuilder
```

**Example:**
```go
assertions.Assert(result).
    TestStatus("integration_test_1", "passed").
    TestResult("unit_test_2", "success").
    MustPass()
```

### Event and Timing Assertions

#### HasEvent
Validates the presence of a specific event type.

```go
func (ab *AssertionBuilder) HasEvent(eventType string) *AssertionBuilder
```

**Example:**
```go
assertions.Assert(result).
    HasEvent("epic_started").
    HasEvent("phase_completed").
    MustPass()
```

#### EventCount
Validates the total number of events generated.

```go
func (ab *AssertionBuilder) EventCount(expectedCount int) *AssertionBuilder
```

#### EventSequence
Validates that events occurred in a specific sequence.

```go
func (ab *AssertionBuilder) EventSequence(expectedSequence []string) *AssertionBuilder
```

**Example:**
```go
assertions.Assert(result).
    EventSequence([]string{
        "epic_started",
        "phase_started", 
        "task_completed",
        "phase_completed",
        "epic_completed",
    }).
    MustPass()
```

#### ExecutionTime
Validates that execution completed within a time limit.

```go
func (ab *AssertionBuilder) ExecutionTime(maxDuration time.Duration) *AssertionBuilder
```

**Example:**
```go
assertions.Assert(result).
    ExecutionTime(5 * time.Second).
    MustPass()
```

### Error Handling Assertions

#### NoErrors
Validates that no errors occurred during execution.

```go
func (ab *AssertionBuilder) NoErrors() *AssertionBuilder
```

#### HasErrors
Validates that errors occurred (useful for negative testing).

```go
func (ab *AssertionBuilder) HasErrors() *AssertionBuilder
```

#### ErrorCount
Validates the exact number of errors.

```go
func (ab *AssertionBuilder) ErrorCount(expectedCount int) *AssertionBuilder
```

### Command Assertions

#### CommandCount
Validates the number of commands executed.

```go
func (ab *AssertionBuilder) CommandCount(expectedCount int) *AssertionBuilder
```

#### AllCommandsSuccessful
Validates that all commands completed successfully.

```go
func (ab *AssertionBuilder) AllCommandsSuccessful() *AssertionBuilder
```

### Advanced Assertions

#### StateProgression
Validates the progression through expected states.

```go
func (ab *AssertionBuilder) StateProgression(expectedStates []string) *AssertionBuilder
```

**Example:**
```go
assertions.Assert(result).
    StateProgression([]string{"pending", "active", "completed"}).
    MustPass()
```

#### IntermediateState
Validates intermediate states using custom validation logic.

```go
func (ab *AssertionBuilder) IntermediateState(stepIndex int, validator func(*epic.Epic) error) *AssertionBuilder
```

**Example:**
```go
assertions.Assert(result).
    IntermediateState(1, func(epic *epic.Epic) error {
        if len(epic.Phases) != 2 {
            return fmt.Errorf("expected 2 phases, got %d", len(epic.Phases))
        }
        return nil
    }).
    MustPass()
```

#### PhaseTransitionTiming
Validates timing for specific phase transitions.

```go
func (ab *AssertionBuilder) PhaseTransitionTiming(phaseID string, maxDuration time.Duration) *AssertionBuilder
```

#### CustomAssertion
Allows completely custom validation logic.

```go
func (ab *AssertionBuilder) CustomAssertion(name string, validator func(*executor.TransitionChainResult) error) *AssertionBuilder
```

**Example:**
```go
assertions.Assert(result).
    CustomAssertion("task_dependency_check", func(result *executor.TransitionChainResult) error {
        // Custom validation logic
        return nil
    }).
    MustPass()
```

### Snapshot Testing

#### MatchSnapshot
Compares the final state against a stored snapshot.

```go
func (ab *AssertionBuilder) MatchSnapshot(name string) *AssertionBuilder
```

#### MatchXMLSnapshot
Compares XML representation against a stored snapshot.

```go
func (ab *AssertionBuilder) MatchXMLSnapshot(name string, element interface{}) *AssertionBuilder
```

#### MatchSelectiveSnapshot
Compares only specific fields against a snapshot.

```go
func (ab *AssertionBuilder) MatchSelectiveSnapshot(name string, fields []string) *AssertionBuilder
```

**Example:**
```go
assertions.Assert(result).
    MatchSnapshot("epic_completion_state").
    MatchSelectiveSnapshot("phase_status", []string{"id", "status", "completion_time"}).
    MustPass()
```

### Performance Testing

#### PerformanceBenchmark
Validates performance characteristics.

```go
func (ab *AssertionBuilder) PerformanceBenchmark(maxDuration time.Duration, maxMemoryMB int) *AssertionBuilder
```

**Example:**
```go
assertions.Assert(result).
    PerformanceBenchmark(1*time.Second, 50).
    MustPass()
```

#### BatchAssertions
Executes multiple assertions efficiently.

```go
func (ab *AssertionBuilder) BatchAssertions(assertions []func(*AssertionBuilder) *AssertionBuilder) *AssertionBuilder
```

### Debug and Visualization

#### WithDebugMode
Enables debug tracing and logging.

```go
func (ab *AssertionBuilder) WithDebugMode(mode DebugMode) *AssertionBuilder
```

**Debug Modes:**
- `DebugOff`: No debug output
- `DebugBasic`: Basic assertion failures only
- `DebugVerbose`: Detailed state information
- `DebugTrace`: Full execution trace

**Example:**
```go
assertions.Assert(result).
    WithDebugMode(assertions.DebugVerbose).
    EpicStatus("completed").
    PrintDebugInfo().
    MustPass()
```

#### EnableStateVisualization
Enables state transition visualization.

```go
func (ab *AssertionBuilder) EnableStateVisualization() *AssertionBuilder
```

#### PrintDebugInfo
Outputs debug information to console.

```go
func (ab *AssertionBuilder) PrintDebugInfo() *AssertionBuilder
```

### Error Recovery

#### WithRecoveryStrategy
Sets a custom error recovery strategy.

```go
func (ab *AssertionBuilder) WithRecoveryStrategy(strategy *RecoveryStrategy) *AssertionBuilder
```

#### RecoverFromErrors
Attempts to recover from assertion failures.

```go
func (ab *AssertionBuilder) RecoverFromErrors() *AssertionBuilder
```

### Execution Methods

#### Check
Executes all assertions and returns the first error (if any).

```go
func (ab *AssertionBuilder) Check() error
```

**Example:**
```go
if err := assertions.Assert(result).EpicStatus("completed").Check(); err != nil {
    t.Fatal(err)
}
```

#### MustPass
Executes all assertions and panics on any failure.

```go
func (ab *AssertionBuilder) MustPass()
```

**Example:**
```go
assertions.Assert(result).
    EpicStatus("completed").
    NoErrors().
    MustPass()
```

### Information Retrieval

#### GetErrors
Returns all assertion errors.

```go
func (ab *AssertionBuilder) GetErrors() []AssertionError
```

#### GetDebugTrace
Returns the debug trace entries.

```go
func (ab *AssertionBuilder) GetDebugTrace() []TraceEntry
```

#### GetStateVisualization
Returns the state visualization data.

```go
func (ab *AssertionBuilder) GetStateVisualization() *StateVisualization
```

## Error Handling

### AssertionError Structure

```go
type AssertionError struct {
    Type        string                 // Error classification
    Message     string                 // Human-readable error message
    Expected    interface{}            // Expected value
    Actual      interface{}            // Actual value
    Context     map[string]interface{} // Additional context
    Suggestions []string               // Helpful suggestions for fixing
}
```

### Error Recovery

The framework supports automatic error recovery through configurable strategies:

```go
type RecoveryStrategy struct {
    CanRecover   func(error) bool
    RecoverFunc  func(error, *ErrorContext) error
    ContinueFunc func(*ErrorContext) bool
}
```

## Best Practices

1. **Use fluent chaining** for readable test specifications
2. **Enable debug mode** during development for better error messages
3. **Use snapshot testing** for complex state validations
4. **Implement custom assertions** for domain-specific validations
5. **Use batch assertions** for performance in large test suites
6. **Enable state visualization** for debugging complex transition failures

## Thread Safety

The AssertionBuilder is **not thread-safe**. Create separate instances for concurrent testing or use synchronization mechanisms.

For parallel test execution, use isolated builders:

```go
func TestParallelExecution(t *testing.T) {
    t.Parallel()
    
    // Each test gets its own builder instance
    assertions.Assert(result).
        EpicStatus("completed").
        MustPass()
}
```