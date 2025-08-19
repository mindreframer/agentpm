# Epic 14 Migration Guide - From XML-Based Tests to Transition Chain Framework

## Overview

This guide helps you migrate from existing XML-based testing patterns to the new Epic 14 Transition Chain Testing Framework. The new framework provides type-safe, fluent API testing with enhanced debugging and visualization capabilities.

## Table of Contents

1. [Migration Strategy](#migration-strategy)
2. [XML Test Pattern Analysis](#xml-test-pattern-analysis)
3. [Step-by-Step Migration](#step-by-step-migration)
4. [Common Migration Patterns](#common-migration-patterns)
5. [Advanced Migration Scenarios](#advanced-migration-scenarios)
6. [Migration Tools and Utilities](#migration-tools-and-utilities)
7. [Validation and Testing](#validation-and-testing)

## Migration Strategy

### Phase 1: Assessment (Recommended)
1. **Inventory existing XML tests** - Catalog all XML-based test files
2. **Identify test patterns** - Categorize tests by complexity and scope
3. **Plan migration order** - Start with simple tests, progress to complex ones
4. **Set up parallel testing** - Run both old and new tests during transition

### Phase 2: Incremental Migration
1. **Migrate simple tests first** - Basic status and state validations
2. **Convert complex scenarios** - Multi-phase and event-driven tests
3. **Add new framework features** - Leverage debugging and visualization
4. **Validate equivalence** - Ensure new tests cover same scenarios

### Phase 3: Cleanup
1. **Remove XML test dependencies** - Clean up old test files
2. **Update CI/CD pipelines** - Switch to new test execution
3. **Documentation updates** - Update team guidelines

## XML Test Pattern Analysis

### Common XML Test Patterns

#### Pattern 1: Basic Epic Status Validation
**Old XML-based approach:**
```go
func TestEpicStatusXML(t *testing.T) {
    epic := loadEpicFromXML("test-epic.xml")
    xmlOutput := executeEpicCommands(epic, []string{"start", "complete"})
    
    if !strings.Contains(xmlOutput, `status="completed"`) {
        t.Error("Epic should be completed")
    }
    if !strings.Contains(xmlOutput, `<phase id="1A" status="completed">`) {
        t.Error("Phase 1A should be completed")
    }
}
```

**New Transition Chain approach:**
```go
func TestEpicStatusNewFramework(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("test-epic").
        ExecutePhase("1A").
        CompleteEpic().
        Execute()
    
    assertions.Assert(result).
        EpicStatus("completed").
        PhaseStatus("1A", "completed").
        MustPass()
}
```

#### Pattern 2: Event Sequence Validation
**Old XML-based approach:**
```go
func TestEventSequenceXML(t *testing.T) {
    epic := loadEpicFromXML("test-epic.xml")
    xmlOutput := executeEpicCommands(epic, []string{"start", "start-phase", "complete-phase"})
    
    events := parseEventsFromXML(xmlOutput)
    expectedSequence := []string{"epic_started", "phase_started", "phase_completed"}
    
    if !validateEventSequence(events, expectedSequence) {
        t.Error("Event sequence validation failed")
    }
}
```

**New Transition Chain approach:**
```go
func TestEventSequenceNewFramework(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("test-epic").
        ExecutePhase("1A").
        Execute()
    
    assertions.Assert(result).
        EventSequence([]string{"epic_started", "phase_started", "phase_completed"}).
        EventCount(3).
        MustPass()
}
```

#### Pattern 3: Error Handling and Validation
**Old XML-based approach:**
```go
func TestErrorHandlingXML(t *testing.T) {
    epic := loadEpicFromXML("test-epic.xml")
    xmlOutput := executeEpicCommands(epic, []string{"start", "invalid-phase"})
    
    if !strings.Contains(xmlOutput, `<error>`) {
        t.Error("Should contain error element")
    }
    if !strings.Contains(xmlOutput, `status="failed"`) {
        t.Error("Epic should be in failed status")
    }
}
```

**New Transition Chain approach:**
```go
func TestErrorHandlingNewFramework(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("test-epic").
        ExecutePhase("invalid-phase").
        Execute()
    
    assertions.Assert(result).
        EpicStatus("failed").
        HasErrors().
        ErrorCount(1).
        MustPass()
}
```

## Step-by-Step Migration

### Step 1: Set Up New Framework

First, import the new testing framework:

```go
import (
    "github.com/mindreframer/agentpm/internal/testing/assertions"
    "github.com/mindreframer/agentpm/internal/testing/executor"
)
```

### Step 2: Identify Test Data Sources

#### Migrating from XML File-Based Tests

**Before (XML file loading):**
```go
func loadTestEpic() *epic.Epic {
    xmlData, err := os.ReadFile("testdata/epic-lifecycle-test.xml")
    if err != nil {
        panic(err)
    }
    return parseEpicFromXML(xmlData)
}
```

**After (Programmatic creation):**
```go
func createTestEpic() *epic.Epic {
    return &epic.Epic{
        ID:     "8",
        Name:   "Epic Name",
        Status: epic.StatusWIP,
        Phases: []epic.Phase{
            {ID: "1A", Name: "Setup", Status: epic.StatusPlanning},
            {ID: "1B", Name: "Development", Status: epic.StatusPlanning},
        },
        Tasks: []epic.Task{
            {ID: "1A_1", PhaseID: "1A", Name: "Initialize Project", Status: epic.StatusPlanning},
            {ID: "1A_2", PhaseID: "1A", Name: "Configure Tools", Status: epic.StatusPlanning},
        },
        Tests: []epic.Test{
            {ID: "T1A_1", TaskID: "1A_1", Name: "Test Project Init", Status: epic.StatusPlanning},
        },
    }
}
```

### Step 3: Convert Test Logic

#### Basic Conversion Template

```go
// Template for converting XML-based tests
func convertXMLTest(t *testing.T, 
    xmlTestName string,
    commands []string,
    expectedXMLContent []string) {
    
    // OLD: XML-based approach
    /*
    epic := loadEpicFromXML(xmlTestName + ".xml")
    xmlOutput := executeCommands(epic, commands)
    for _, expected := range expectedXMLContent {
        if !strings.Contains(xmlOutput, expected) {
            t.Errorf("XML output should contain: %s", expected)
        }
    }
    */
    
    // NEW: Transition Chain approach
    chain := executor.NewTransitionChain().StartEpic(xmlTestName)
    
    for _, command := range commands {
        chain = chain.ExecuteCommand(command)
    }
    
    result := chain.Execute()
    
    // Convert XML expectations to structured assertions
    builder := assertions.Assert(result)
    
    // Add specific assertions based on expectedXMLContent
    // This requires domain knowledge to convert XML patterns to structured assertions
    
    builder.MustPass()
}
```

### Step 4: Convert Specific Test Patterns

#### Converting Status Checks

```go
// XML pattern: strings.Contains(xml, `status="completed"`)
// Converts to: .EpicStatus("completed")

// XML pattern: strings.Contains(xml, `<phase id="1A" status="completed">`)
// Converts to: .PhaseStatus("1A", "completed")

// XML pattern: strings.Contains(xml, `<task id="1A_1" status="wip">`)
// Converts to: .TaskStatus("1A_1", "wip")
```

#### Converting Event Validations

```go
// XML pattern: Count of <event> elements
// Converts to: .EventCount(expectedCount)

// XML pattern: Specific event type presence
// Converts to: .HasEvent("event_type")

// XML pattern: Event sequence validation
// Converts to: .EventSequence([]string{"event1", "event2", "event3"})
```

#### Converting Error Conditions

```go
// XML pattern: strings.Contains(xml, `<error>`)
// Converts to: .HasErrors()

// XML pattern: strings.Contains(xml, `status="failed"`)
// Converts to: .EpicStatus("failed")

// XML pattern: Count of error elements
// Converts to: .ErrorCount(expectedCount)
```

## Common Migration Patterns

### Pattern 1: Test Suite Migration

**Before (XML-based suite):**
```go
func TestEpicLifecycleXMLSuite(t *testing.T) {
    testCases := []struct {
        name     string
        xmlFile  string
        commands []string
        checks   []string
    }{
        {
            name:     "basic_completion",
            xmlFile:  "basic-epic.xml",
            commands: []string{"start", "complete"},
            checks:   []string{`status="completed"`, `<phase id="1A" status="completed">`},
        },
        // More test cases...
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            epic := loadEpicFromXML(tc.xmlFile)
            output := executeCommands(epic, tc.commands)
            for _, check := range tc.checks {
                if !strings.Contains(output, check) {
                    t.Errorf("Missing: %s", check)
                }
            }
        })
    }
}
```

**After (Transition Chain suite):**
```go
func TestEpicLifecycleNewFrameworkSuite(t *testing.T) {
    testCases := []struct {
        name       string
        epicID     string
        operations func(*executor.TransitionChain) *executor.TransitionChain
        assertions func(*assertions.AssertionBuilder) *assertions.AssertionBuilder
    }{
        {
            name:   "basic_completion",
            epicID: "basic-epic",
            operations: func(chain *executor.TransitionChain) *executor.TransitionChain {
                return chain.StartEpic("basic-epic").
                    ExecutePhase("1A").
                    CompleteEpic()
            },
            assertions: func(ab *assertions.AssertionBuilder) *assertions.AssertionBuilder {
                return ab.EpicStatus("completed").
                    PhaseStatus("1A", "completed")
            },
        },
        // More test cases...
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := tc.operations(executor.NewTransitionChain()).Execute()
            tc.assertions(assertions.Assert(result)).MustPass()
        })
    }
}
```

### Pattern 2: Complex Event Validation Migration

**Before (XML event parsing):**
```go
func TestComplexEventValidationXML(t *testing.T) {
    epic := loadEpicFromXML("complex-epic.xml")
    output := executeCommands(epic, []string{"start", "phase1", "phase2", "complete"})
    
    events := parseEventsFromXML(output)
    
    // Validate event count
    if len(events) != 8 {
        t.Errorf("Expected 8 events, got %d", len(events))
    }
    
    // Validate event types
    expectedTypes := []string{"epic_started", "phase_started", "task_completed", "phase_completed"}
    for _, expectedType := range expectedTypes {
        found := false
        for _, event := range events {
            if event.Type == expectedType {
                found = true
                break
            }
        }
        if !found {
            t.Errorf("Missing event type: %s", expectedType)
        }
    }
    
    // Validate event sequence
    if events[0].Type != "epic_started" {
        t.Error("First event should be epic_started")
    }
    if events[len(events)-1].Type != "epic_completed" {
        t.Error("Last event should be epic_completed")
    }
}
```

**After (Structured event validation):**
```go
func TestComplexEventValidationNewFramework(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("complex-epic").
        ExecutePhase("1A").
        ExecutePhase("1B").
        CompleteEpic().
        Execute()
    
    assertions.Assert(result).
        EventCount(8).
        HasEvent("epic_started").
        HasEvent("phase_started").
        HasEvent("task_completed").
        HasEvent("phase_completed").
        EventSequence([]string{
            "epic_started",
            "phase_started",
            "task_completed", 
            "phase_completed",
            "phase_started",
            "task_completed",
            "phase_completed",
            "epic_completed",
        }).
        MustPass()
}
```

### Pattern 3: Snapshot Testing Migration

**Before (XML string comparison):**
```go
func TestEpicStateSnapshotXML(t *testing.T) {
    epic := loadEpicFromXML("snapshot-epic.xml")
    output := executeCommands(epic, []string{"start", "phase1", "complete"})
    
    expectedXML := loadExpectedXML("snapshot-epic-expected.xml")
    
    if normalizeXML(output) != normalizeXML(expectedXML) {
        t.Error("XML output does not match expected snapshot")
    }
}
```

**After (Structured snapshot testing):**
```go
func TestEpicStateSnapshotNewFramework(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("snapshot-epic").
        ExecutePhase("1A").
        CompleteEpic().
        Execute()
    
    assertions.Assert(result).
        MatchSnapshot("epic_completion_state").
        MatchSelectiveSnapshot("phase_timing", []string{"id", "start_time", "completion_time"}).
        MustPass()
}
```

## Advanced Migration Scenarios

### Scenario 1: Custom XML Validation Logic

**Before (Complex XML parsing):**
```go
func TestCustomXMLValidation(t *testing.T) {
    epic := loadEpicFromXML("custom-epic.xml")
    output := executeCommands(epic, []string{"start", "complex-operation"})
    
    doc := parseXMLDocument(output)
    
    // Custom validation logic
    phases := doc.SelectElements("//phase[@status='completed']")
    if len(phases) < 2 {
        t.Error("At least 2 phases should be completed")
    }
    
    for _, phase := range phases {
        tasks := phase.SelectElements("task[@status!='completed']")
        if len(tasks) > 0 {
            t.Errorf("Phase %s has incomplete tasks", phase.GetAttribute("id"))
        }
    }
    
    // Validate dependencies
    deps := doc.SelectElements("//dependency[@status='satisfied']")
    if len(deps) != len(doc.SelectElements("//dependency")) {
        t.Error("All dependencies should be satisfied")
    }
}
```

**After (Custom assertions):**
```go
func TestCustomValidationNewFramework(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("custom-epic").
        ExecuteCommand("complex-operation").
        Execute()
    
    assertions.Assert(result).
        CustomAssertion("completed_phases_check", func(result *executor.TransitionChainResult) error {
            completedPhases := 0
            for _, phase := range result.FinalState.Phases {
                if phase.Status == "completed" {
                    completedPhases++
                }
            }
            if completedPhases < 2 {
                return fmt.Errorf("expected at least 2 completed phases, got %d", completedPhases)
            }
            return nil
        }).
        CustomAssertion("task_completion_check", func(result *executor.TransitionChainResult) error {
            for _, phase := range result.FinalState.Phases {
                if phase.Status == "completed" {
                    for _, task := range result.FinalState.Tasks {
                        if task.PhaseID == phase.ID && task.Status != "completed" {
                            return fmt.Errorf("phase %s has incomplete task %s", phase.ID, task.ID)
                        }
                    }
                }
            }
            return nil
        }).
        CustomAssertion("dependency_check", func(result *executor.TransitionChainResult) error {
            // Custom dependency validation logic
            for _, event := range result.FinalState.Events {
                if event.Type == "dependency_unsatisfied" {
                    return fmt.Errorf("unsatisfied dependency: %s", event.Data)
                }
            }
            return nil
        }).
        MustPass()
}
```

### Scenario 2: Performance Testing Migration

**Before (XML-based timing):**
```go
func TestPerformanceXML(t *testing.T) {
    start := time.Now()
    
    epic := loadEpicFromXML("performance-epic.xml")
    output := executeCommands(epic, []string{"start", "phase1", "phase2", "complete"})
    
    duration := time.Since(start)
    
    if duration > 5*time.Second {
        t.Errorf("Execution took too long: %v", duration)
    }
    
    if !strings.Contains(output, `status="completed"`) {
        t.Error("Epic should be completed")
    }
    
    // Memory usage check would be manual
}
```

**After (Integrated performance testing):**
```go
func TestPerformanceNewFramework(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("performance-epic").
        ExecutePhase("1A").
        ExecutePhase("1B").
        CompleteEpic().
        Execute()
    
    assertions.Assert(result).
        EpicStatus("completed").
        ExecutionTime(5 * time.Second).
        PerformanceBenchmark(5*time.Second, 100). // Max 100MB memory
        MustPass()
}
```

## Migration Tools and Utilities

### Automated Migration Script

Create a migration script to help convert common patterns:

```go
package main

import (
    "bufio"
    "fmt"
    "os"
    "regexp"
    "strings"
)

type MigrationTool struct {
    patterns map[string]string
}

func NewMigrationTool() *MigrationTool {
    return &MigrationTool{
        patterns: map[string]string{
            // XML pattern -> New framework pattern
            `strings\.Contains\(.*,\s*[\'"]status="completed"[\'\"]\)`: `.EpicStatus("completed")`,
            `strings\.Contains\(.*,\s*[\'"]<phase id="([^"]+)" status="([^"]+)">[\'\"]\)`: `.PhaseStatus("$1", "$2")`,
            `strings\.Contains\(.*,\s*[\'"]<task id="([^"]+)" status="([^"]+)">[\'\"]\)`: `.TaskStatus("$1", "$2")`,
            `strings\.Contains\(.*,\s*[\'"]<error>[\'\"]\)`: `.HasErrors()`,
            `len\(.*events.*\)\s*==\s*(\d+)`: `.EventCount($1)`,
        },
    }
}

func (mt *MigrationTool) ConvertTestFile(inputFile, outputFile string) error {
    input, err := os.Open(inputFile)
    if err != nil {
        return err
    }
    defer input.Close()
    
    output, err := os.Create(outputFile)
    if err != nil {
        return err
    }
    defer output.Close()
    
    scanner := bufio.NewScanner(input)
    for scanner.Scan() {
        line := scanner.Text()
        convertedLine := mt.convertLine(line)
        fmt.Fprintln(output, convertedLine)
    }
    
    return scanner.Err()
}

func (mt *MigrationTool) convertLine(line string) string {
    for pattern, replacement := range mt.patterns {
        re := regexp.MustCompile(pattern)
        line = re.ReplaceAllString(line, replacement)
    }
    return line
}

// Usage example
func main() {
    tool := NewMigrationTool()
    
    // Convert a test file
    err := tool.ConvertTestFile("old_test.go", "new_test.go")
    if err != nil {
        fmt.Printf("Migration failed: %v\n", err)
    } else {
        fmt.Println("Migration completed successfully")
    }
}
```

### Migration Validation Tool

```go
// Tool to validate that new tests cover the same scenarios as old tests
type MigrationValidator struct {
    oldTestResults map[string]TestResult
    newTestResults map[string]TestResult
}

type TestResult struct {
    Status       string
    EventCount   int
    PhaseStates  map[string]string
    TaskStates   map[string]string
    ErrorCount   int
}

func (mv *MigrationValidator) ValidateEquivalence(testName string) error {
    oldResult, oldExists := mv.oldTestResults[testName]
    newResult, newExists := mv.newTestResults[testName]
    
    if !oldExists && !newExists {
        return fmt.Errorf("test %s not found in either old or new results", testName)
    }
    
    if !oldExists {
        return fmt.Errorf("test %s not found in old results", testName)
    }
    
    if !newExists {
        return fmt.Errorf("test %s not found in new results", testName)
    }
    
    // Compare key metrics
    if oldResult.Status != newResult.Status {
        return fmt.Errorf("status mismatch for %s: old=%s, new=%s", 
            testName, oldResult.Status, newResult.Status)
    }
    
    if oldResult.EventCount != newResult.EventCount {
        return fmt.Errorf("event count mismatch for %s: old=%d, new=%d", 
            testName, oldResult.EventCount, newResult.EventCount)
    }
    
    // Compare phase states
    for phaseID, oldState := range oldResult.PhaseStates {
        if newState, exists := newResult.PhaseStates[phaseID]; !exists || newState != oldState {
            return fmt.Errorf("phase state mismatch for %s phase %s: old=%s, new=%s", 
                testName, phaseID, oldState, newState)
        }
    }
    
    return nil
}
```

## Validation and Testing

### Parallel Testing Approach

During migration, run both old and new tests to ensure equivalence:

```go
func TestMigrationValidation(t *testing.T) {
    testCases := []string{
        "basic_epic_completion",
        "multi_phase_epic",
        "error_handling_scenario",
        "complex_event_sequence",
    }
    
    for _, testName := range testCases {
        t.Run(testName, func(t *testing.T) {
            // Run old XML-based test
            oldResult := runOldXMLTest(testName)
            
            // Run new framework test
            newResult := runNewFrameworkTest(testName)
            
            // Validate equivalence
            if err := validateTestEquivalence(oldResult, newResult); err != nil {
                t.Errorf("Test equivalence validation failed for %s: %v", testName, err)
            }
        })
    }
}
```

### Migration Checklist

- [ ] **Inventory complete** - All XML tests identified and categorized
- [ ] **Simple tests migrated** - Basic status and validation tests converted
- [ ] **Complex tests migrated** - Multi-step and event-driven tests converted
- [ ] **Custom validations migrated** - Domain-specific logic converted to custom assertions
- [ ] **Performance tests migrated** - Timing and resource usage tests converted
- [ ] **Snapshot tests migrated** - State comparison tests converted
- [ ] **Equivalence validated** - New tests produce same results as old tests
- [ ] **Documentation updated** - Team guidelines and examples updated
- [ ] **CI/CD updated** - Build pipelines switched to new framework
- [ ] **Old tests removed** - XML-based tests cleaned up after validation

### Post-Migration Benefits

After completing the migration, you'll gain:

1. **Type Safety** - Compile-time validation of test assertions
2. **Better IDE Support** - Auto-completion and refactoring support
3. **Enhanced Debugging** - Rich error messages with suggestions
4. **State Visualization** - Visual debugging of state transitions
5. **Performance Insights** - Built-in performance and memory monitoring
6. **Maintainability** - Easier to modify and extend test logic
7. **Consistency** - Standardized testing patterns across the codebase