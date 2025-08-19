package assertions

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/testing/executor"
)

// === PHASE 4D: COMPREHENSIVE ERROR HANDLING TESTS ===

func TestErrorHandling_ErrorContextIncludesRelevantStateInformation(t *testing.T) {
	// Create a complex epic state with multiple phases and tasks
	complexEpic := &epic.Epic{
		ID:     "complex-error-test",
		Name:   "Complex Error Test Epic",
		Status: epic.StatusWIP,
		Phases: []epic.Phase{
			{
				ID:     "1A",
				Name:   "Setup Phase",
				Status: epic.StatusCompleted,
			},
			{
				ID:     "1B",
				Name:   "Active Phase",
				Status: epic.StatusWIP,
			},
		},
		Tasks: []epic.Task{
			{
				ID:      "1A_1",
				Name:    "Setup Task",
				PhaseID: "1A",
				Status:  epic.StatusCompleted,
			},
			{
				ID:      "1B_1",
				Name:    "Active Task",
				PhaseID: "1B",
				Status:  epic.StatusWIP,
			},
		},
		Events: []epic.Event{
			{Type: "epic_started", Data: "Epic started"},
			{Type: "phase_started", Data: "Phase 1A started"},
			{Type: "task_started", Data: "Task 1A_1 started"},
		},
	}

	// Create result with complex state and errors
	result := &executor.TransitionChainResult{
		InitialState: &epic.Epic{ID: "complex-error-test", Status: epic.StatusPending},
		FinalState:   complexEpic,
		IntermediateStates: []executor.StateSnapshot{
			{EpicState: &epic.Epic{ID: "complex-error-test", Status: epic.StatusWIP}},
		},
		Errors: []executor.TransitionError{
			{
				Command:       "complete task 1B_1",
				ExpectedState: "completed",
				ActualState:   "wip",
				Epic:          complexEpic,
			},
		},
		ExecutionTime: time.Millisecond * 250,
	}

	// Test with comprehensive debug context
	builder := NewAssertionBuilder(result).
		WithDebugMode(DebugTrace).
		EnableStateVisualization()

	// Trigger multiple assertion failures to create rich error context
	builder.EpicStatus("completed").
		PhaseStatus("1B", "completed").
		TaskStatus("1B_1", "completed").
		HasEvent("task_completed").
		EventCount(5) // Should fail - only 3 events

	errors := builder.GetErrors()
	if len(errors) == 0 {
		t.Fatal("Expected assertion errors to test error context")
	}

	// Verify each error has enhanced suggestions based on the complex state
	for i, err := range errors {
		if len(err.Suggestions) == 0 {
			t.Errorf("Error %d should have suggestions: %s", i, err.Message)
		}

		// Verify suggestions are contextually relevant
		switch {
		case strings.Contains(err.Message, "epic status"):
			found := false
			for _, suggestion := range err.Suggestions {
				if strings.Contains(suggestion, "expected state") ||
					strings.Contains(suggestion, "transition chain") {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Epic status error should have relevant suggestions: %v", err.Suggestions)
			}

		case strings.Contains(err.Message, "Phase"):
			found := false
			for _, suggestion := range err.Suggestions {
				if strings.Contains(suggestion, "phase") {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Phase error should have phase-related suggestions: %v", err.Suggestions)
			}

		case strings.Contains(err.Message, "event"):
			found := false
			for _, suggestion := range err.Suggestions {
				if strings.Contains(suggestion, "event") || strings.Contains(suggestion, "debug") {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Event error should have event-related suggestions: %v", err.Suggestions)
			}
		}
	}

	// Verify debug trace captures the full context
	trace := builder.GetDebugTrace()
	if len(trace) == 0 {
		t.Error("Expected debug trace to capture assertion failures")
	}

	// Check that trace entries have meaningful context
	for _, entry := range trace {
		if entry.Level == "ERROR" {
			if entry.Context == nil || len(entry.Context) == 0 {
				t.Error("Error trace entries should include context information")
			}

			// Verify trace includes assertion details
			if _, hasType := entry.Context["type"]; !hasType {
				t.Error("Error trace should include assertion type")
			}
		}
	}

	// Verify state visualization captures the failure context
	viz := builder.GetStateVisualization()
	if viz == nil {
		t.Error("Expected state visualization to be available")
	} else {
		// Check timeline shows state progression
		timeline := viz.GetTimelineVisualization()
		if !strings.Contains(timeline, "complex-error-test") {
			t.Error("Timeline should show the epic being tested")
		}

		// Check graph shows meaningful transitions
		graph := viz.GetGraphVisualization()
		if !strings.Contains(graph, "State Transition Graph:") {
			t.Error("Graph should have proper header")
		}
	}
}

func TestErrorHandling_DebugModeProvidesDifferentLevelsOfDetail(t *testing.T) {
	// Create test epic
	epic := &epic.Epic{
		ID:     "debug-levels-test",
		Status: epic.StatusWIP,
	}

	result := &executor.TransitionChainResult{
		FinalState: epic,
	}

	// Test different debug levels
	debugLevels := []struct {
		mode            DebugMode
		expectedTracing bool
		expectedVerbose bool
	}{
		{DebugOff, false, false},
		{DebugBasic, true, false},
		{DebugVerbose, true, true},
		{DebugTrace, true, true},
	}

	for _, level := range debugLevels {
		t.Run(fmt.Sprintf("debug_level_%d", int(level.mode)), func(t *testing.T) {
			builder := NewAssertionBuilder(result).WithDebugMode(level.mode)

			// Trigger an assertion failure
			builder.EpicStatus("completed")

			trace := builder.GetDebugTrace()

			if level.expectedTracing && len(trace) == 0 {
				t.Error("Expected debug trace entries for this level")
			} else if !level.expectedTracing && len(trace) > 0 {
				t.Error("Expected no debug trace entries for this level")
			}

			// For verbose and trace levels, verify additional context
			if level.expectedVerbose && len(trace) > 0 {
				for _, entry := range trace {
					if entry.Level == "ERROR" {
						if entry.Location == "" {
							t.Error("Verbose mode should include location information")
						}
						if entry.Timestamp.IsZero() {
							t.Error("Verbose mode should include timestamps")
						}
					}
				}
			}
		})
	}
}

func TestErrorHandling_StateVisualizationHandlesComplexScenarios(t *testing.T) {
	// Create a complex state progression
	states := []interface{}{
		&epic.Epic{ID: "viz-test", Status: epic.StatusPending},
		&epic.Epic{ID: "viz-test", Status: epic.StatusWIP, Phases: []epic.Phase{
			{ID: "1A", Status: epic.StatusWIP},
		}},
		&epic.Epic{ID: "viz-test", Status: epic.StatusWIP, Phases: []epic.Phase{
			{ID: "1A", Status: epic.StatusCompleted},
			{ID: "1B", Status: epic.StatusWIP},
		}},
		&epic.Epic{ID: "viz-test", Status: epic.StatusCompleted, Phases: []epic.Phase{
			{ID: "1A", Status: epic.StatusCompleted},
			{ID: "1B", Status: epic.StatusCompleted},
		}},
	}

	commands := []string{
		"start epic",
		"start phase 1A",
		"complete epic", // This represents the final transition to completed state
	}

	// Create visualization
	viz := CreateStateVisualization(states, commands)

	// Test timeline visualization
	timeline := viz.GetTimelineVisualization()

	// Should show all state transitions
	if !strings.Contains(timeline, "pending") {
		t.Error("Timeline should show initial pending state")
	}
	if !strings.Contains(timeline, "wip") {
		t.Error("Timeline should show active states")
	}
	if !strings.Contains(timeline, "completed") {
		t.Error("Timeline should show completed state")
	}

	// Test graph visualization
	graph := viz.GetGraphVisualization()

	// Should show command flow
	if !strings.Contains(graph, "start epic") {
		t.Error("Graph should show start epic command")
	}
	if !strings.Contains(graph, "complete epic") {
		t.Error("Graph should show complete epic command")
	}

	// Should show success indicators
	if !strings.Contains(graph, "✓") {
		t.Error("Graph should show success indicators")
	}

	// Test that visualization handles edge cases
	emptyViz := CreateStateVisualization([]interface{}{}, []string{})
	emptyTimeline := emptyViz.GetTimelineVisualization()
	if !strings.Contains(emptyTimeline, "State Timeline:") {
		t.Error("Empty visualization should still have proper headers")
	}

	// Test mismatched states and commands
	mismatchedViz := CreateStateVisualization(states[:2], commands)
	mismatchedGraph := mismatchedViz.GetGraphVisualization()
	if !strings.Contains(mismatchedGraph, "State Transition Graph:") {
		t.Error("Mismatched visualization should still be valid")
	}
}

func TestErrorHandling_ParallelExecutionIsolation(t *testing.T) {
	const numGoroutines = 10
	const assertionsPerGoroutine = 5

	var wg sync.WaitGroup
	results := make(chan struct {
		builderID int
		errors    []AssertionError
		trace     []TraceEntry
	}, numGoroutines)

	// Launch multiple goroutines doing assertions
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(builderID int) {
			defer wg.Done()

			// Each goroutine creates its own epic with unique ID
			epic := &epic.Epic{
				ID:     fmt.Sprintf("parallel-test-%d", builderID),
				Status: epic.StatusWIP,
			}

			result := &executor.TransitionChainResult{
				FinalState: epic,
			}

			builder := NewAssertionBuilder(result).
				WithDebugMode(DebugVerbose).
				EnableStateVisualization()

			// Each builder performs multiple assertions
			for j := 0; j < assertionsPerGoroutine; j++ {
				builder.EpicStatus("completed") // Will fail
				builder.HasEvent("test_event")  // Will fail
			}

			// Collect results
			results <- struct {
				builderID int
				errors    []AssertionError
				trace     []TraceEntry
			}{
				builderID: builderID,
				errors:    builder.GetErrors(),
				trace:     builder.GetDebugTrace(),
			}
		}(i)
	}

	// Wait for all goroutines and close channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and verify results
	builderResults := make(map[int]struct {
		errors []AssertionError
		trace  []TraceEntry
	})

	for result := range results {
		builderResults[result.builderID] = struct {
			errors []AssertionError
			trace  []TraceEntry
		}{
			errors: result.errors,
			trace:  result.trace,
		}
	}

	// Verify we got results from all builders
	if len(builderResults) != numGoroutines {
		t.Errorf("Expected results from %d builders, got %d", numGoroutines, len(builderResults))
	}

	// Verify each builder has isolated state
	for builderID, result := range builderResults {
		// Each builder should have exactly the expected number of errors
		expectedErrors := assertionsPerGoroutine * 2 // Two failing assertions per iteration
		if len(result.errors) != expectedErrors {
			t.Errorf("Builder %d: expected %d errors, got %d", builderID, expectedErrors, len(result.errors))
		}

		// Each builder should have its own debug trace
		if len(result.trace) == 0 {
			t.Errorf("Builder %d: expected debug trace entries", builderID)
		}

		// Basic trace validation will be done below
	}

	// Verify no cross-contamination between builders
	// We check that each builder only has errors from its own epic ID
	for builderID, result := range builderResults {
		expectedEpicID := fmt.Sprintf("parallel-test-%d", builderID)

		// Check that debug traces contain the correct epic ID
		for _, entry := range result.trace {
			if entry.Level == "ERROR" && entry.Context != nil {
				if epicID, exists := entry.Context["epic_id"]; exists {
					if epicID != expectedEpicID {
						t.Errorf("Builder %d: found trace from wrong epic: expected %s, got %v",
							builderID, expectedEpicID, epicID)
					}
				}
			}
		}
	}
}

func TestErrorHandling_FailureAnalysisWithComplexErrorPatterns(t *testing.T) {
	// Test complex error scenarios with contextual analysis
	testCases := []struct {
		name                string
		epic                *epic.Epic
		assertions          func(*AssertionBuilder)
		expectedSuggestions map[string][]string // error pattern -> expected suggestions
	}{
		{
			name: "cascading_failures",
			epic: &epic.Epic{
				ID:     "cascade-test",
				Status: epic.StatusWIP,
				Phases: []epic.Phase{
					{ID: "1A", Status: epic.StatusWIP},
				},
			},
			assertions: func(ab *AssertionBuilder) {
				ab.EpicStatus("completed").
					PhaseStatus("1A", "completed").
					PhaseStatus("1B", "wip") // Non-existent phase
			},
			expectedSuggestions: map[string][]string{
				"epic status": {"expected state", "transition chain"},
				"Phase 1A":    {"phase"},
				"Phase 1B":    {"phase", "not found"},
			},
		},
		{
			name: "timing_and_sequence_failures",
			epic: &epic.Epic{
				ID:     "timing-test",
				Status: epic.StatusWIP,
				Events: []epic.Event{
					{Type: "epic_started", Data: "Started"},
				},
			},
			assertions: func(ab *AssertionBuilder) {
				ab.EventSequence([]string{"epic_started", "phase_started", "task_completed"}).
					EventCount(5).
					ExecutionTime(time.Nanosecond) // Impossibly short
			},
			expectedSuggestions: map[string][]string{
				"event sequence": {"debug"},
				"event count":    {"debug"},
				"execution time": {"debug"},
			},
		},
		{
			name: "state_inconsistency",
			epic: &epic.Epic{
				ID:     "inconsistent-test",
				Status: epic.StatusCompleted,
				Phases: []epic.Phase{
					{ID: "1A", Status: epic.StatusWIP}, // Inconsistent - epic completed but phase active
				},
				Tasks: []epic.Task{
					{ID: "1A_1", PhaseID: "1A", Status: epic.StatusPending}, // Also inconsistent
				},
			},
			assertions: func(ab *AssertionBuilder) {
				ab.PhaseStatus("1A", "completed").
					TaskStatus("1A_1", "completed")
			},
			expectedSuggestions: map[string][]string{
				"Phase 1A":  {"expected state"},
				"Task 1A_1": {"expected state"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := &executor.TransitionChainResult{
				FinalState: tc.epic,
			}

			builder := NewAssertionBuilder(result).
				WithDebugMode(DebugTrace).
				EnableStateVisualization()

			// Execute the test assertions
			tc.assertions(builder)

			errors := builder.GetErrors()
			if len(errors) == 0 {
				t.Error("Expected assertion failures for this test case")
			}

			// Verify suggestions are contextually appropriate
			for _, err := range errors {
				foundMatch := false
				for pattern, expectedSuggestions := range tc.expectedSuggestions {
					if strings.Contains(err.Message, pattern) {
						foundMatch = true
						for _, expectedSuggestion := range expectedSuggestions {
							found := false
							for _, actualSuggestion := range err.Suggestions {
								if strings.Contains(actualSuggestion, expectedSuggestion) {
									found = true
									break
								}
							}
							if !found {
								t.Errorf("Error '%s' missing expected suggestion containing '%s'. Got: %v",
									err.Message, expectedSuggestion, err.Suggestions)
							}
						}
						break
					}
				}

				if !foundMatch {
					t.Logf("Error '%s' didn't match any expected patterns", err.Message)
				}
			}

			// Verify debug trace captures the complexity
			trace := builder.GetDebugTrace()
			if len(trace) < len(errors) {
				t.Error("Debug trace should capture all assertion failures")
			}

			// Verify state visualization is available for complex scenarios
			viz := builder.GetStateVisualization()
			if viz == nil {
				t.Error("State visualization should be available for complex error analysis")
			}
		})
	}
}

func TestErrorHandling_MemoryUsageUnderStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory stress test in short mode")
	}

	// Get initial memory stats
	var m1 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	const numIterations = 1000
	const assertionsPerIteration = 10

	// Create many builders with debug mode and visualizations
	for i := 0; i < numIterations; i++ {
		testEpic := &epic.Epic{
			ID:     fmt.Sprintf("stress-test-%d", i),
			Status: epic.StatusWIP,
			Phases: make([]epic.Phase, 5),
			Tasks:  make([]epic.Task, 20),
			Events: make([]epic.Event, 50),
		}

		// Fill with test data
		for j := range testEpic.Phases {
			testEpic.Phases[j] = epic.Phase{
				ID:     fmt.Sprintf("%d%c", i, 'A'+j),
				Status: epic.StatusWIP,
			}
		}
		for j := range testEpic.Tasks {
			testEpic.Tasks[j] = epic.Task{
				ID:      fmt.Sprintf("%d_%d", i, j),
				PhaseID: fmt.Sprintf("%d%c", i, 'A'+(j%5)),
				Status:  epic.StatusWIP,
			}
		}
		for j := range testEpic.Events {
			testEpic.Events[j] = epic.Event{
				Type: fmt.Sprintf("event_%d", j),
				Data: fmt.Sprintf("Event data %d", j),
			}
		}

		result := &executor.TransitionChainResult{
			FinalState: testEpic,
		}

		builder := NewAssertionBuilder(result).
			WithDebugMode(DebugOff)

		// Perform many assertions to generate debug data
		for j := 0; j < assertionsPerIteration; j++ {
			builder.EpicStatus("completed").
				PhaseStatus(fmt.Sprintf("%d%c", i, 'A'), "completed").
				HasEvent(fmt.Sprintf("event_%d", j*10)).
				EventCount(100)
		}

		// Verify builder has captured data
		errors := builder.GetErrors()

		if len(errors) == 0 {
			t.Error("Expected assertion errors")
		}

		// Periodically check memory growth
		if i%100 == 99 {
			runtime.GC()
			var m2 runtime.MemStats
			runtime.ReadMemStats(&m2)

			// Check memory growth is reasonable (less than 100MB)
			memGrowth := int64(m2.Alloc) - int64(m1.Alloc)
			if memGrowth > 100*1024*1024 {
				t.Errorf("Memory growth too large after %d iterations: %d bytes", i+1, memGrowth)
			}
		}
	}

	// Final memory check
	runtime.GC()
	var m3 runtime.MemStats
	runtime.ReadMemStats(&m3)

	totalGrowth := int64(m3.Alloc) - int64(m1.Alloc)
	t.Logf("Total memory growth after %d iterations: %d bytes", numIterations, totalGrowth)

	// Memory growth should be reasonable for the amount of work done
	maxExpectedGrowth := int64(200 * 1024 * 1024) // 200MB
	if totalGrowth > maxExpectedGrowth {
		t.Errorf("Excessive memory growth: %d bytes (max expected: %d)", totalGrowth, maxExpectedGrowth)
	}
}

func TestErrorHandling_RecoveryStrategiesIntegration(t *testing.T) {
	// Test integration of error recovery with the full assertion framework
	epic := &epic.Epic{
		ID:     "recovery-integration-test",
		Status: epic.StatusWIP,
		Phases: []epic.Phase{
			{ID: "1A", Status: epic.StatusCompleted},
			{ID: "1B", Status: epic.StatusWIP},
		},
	}

	result := &executor.TransitionChainResult{
		FinalState: epic,
		Errors: []executor.TransitionError{
			{Command: "complete phase 1B", ExpectedState: "completed", ActualState: "wip"},
		},
	}

	// Create custom recovery strategy that logs recovery attempts
	recoveryLog := make([]string, 0)
	customRecovery := &RecoveryStrategy{
		CanRecover: func(err error) bool {
			return strings.Contains(err.Error(), "Expected") ||
				strings.Contains(err.Error(), "Phase")
		},
		RecoverFunc: func(err error, ctx *ErrorContext) error {
			recoveryLog = append(recoveryLog, fmt.Sprintf("Recovered: %s", err.Error()))
			return nil // Successful recovery
		},
		ContinueFunc: func(ctx *ErrorContext) bool {
			return !strings.Contains(ctx.Err.Error(), "critical")
		},
	}

	builder := NewAssertionBuilder(result).
		WithDebugMode(DebugVerbose).
		WithRecoveryStrategy(customRecovery).
		EnableStateVisualization()

	// Trigger multiple assertion failures
	builder.EpicStatus("completed"). // Recoverable
						PhaseStatus("1B", "completed"). // Recoverable
						PhaseStatus("1C", "wip").       // Recoverable (phase not found)
						HasEvent("critical_failure")    // Should be recoverable

	initialErrorCount := len(builder.GetErrors())
	if initialErrorCount == 0 {
		t.Fatal("Expected initial assertion errors")
	}

	// Attempt recovery
	builder.RecoverFromErrors()

	finalErrorCount := len(builder.GetErrors())
	if finalErrorCount >= initialErrorCount {
		t.Errorf("Recovery should have reduced errors from %d to less, got %d",
			initialErrorCount, finalErrorCount)
	}

	// Verify recovery was logged
	if len(recoveryLog) == 0 {
		t.Error("Expected recovery attempts to be logged")
	}

	// Verify debug trace captured recovery process
	trace := builder.GetDebugTrace()
	recoveryTraceFound := false
	for _, entry := range trace {
		if entry.Level == "INFO" && strings.Contains(entry.Message, "recovery") {
			recoveryTraceFound = true
			break
		}
	}
	if !recoveryTraceFound {
		t.Error("Debug trace should capture recovery process")
	}

	// Test non-recoverable error
	builder.addErrorWithContext("critical_error", "critical system failure",
		"working", "broken", nil, "system", -1)

	preRecoveryCount := len(builder.GetErrors())
	builder.RecoverFromErrors()
	postRecoveryCount := len(builder.GetErrors())

	if postRecoveryCount < preRecoveryCount {
		t.Error("Critical errors should not be recoverable")
	}
}

func TestErrorHandling_IntegrationWithGoTestingFramework(t *testing.T) {
	// Test that error handling integrates properly with Go testing patterns

	// Subtest pattern integration
	t.Run("subtest_error_isolation", func(t *testing.T) {
		epic := &epic.Epic{ID: "subtest-1", Status: epic.StatusWIP}
		result := &executor.TransitionChainResult{FinalState: epic}

		builder := NewAssertionBuilder(result).WithDebugMode(DebugBasic)
		builder.EpicStatus("completed")

		errors := builder.GetErrors()
		if len(errors) != 1 {
			t.Errorf("Expected 1 error in subtest, got %d", len(errors))
		}
	})

	t.Run("subtest_separate_context", func(t *testing.T) {
		epic := &epic.Epic{ID: "subtest-2", Status: epic.StatusPending}
		result := &executor.TransitionChainResult{FinalState: epic}

		builder := NewAssertionBuilder(result).WithDebugMode(DebugBasic)
		builder.EpicStatus("wip")

		errors := builder.GetErrors()
		if len(errors) != 1 {
			t.Errorf("Expected 1 error in separate subtest, got %d", len(errors))
		}
	})

	// Test table-driven test pattern
	testCases := []struct {
		name           string
		epicStatus     epic.Status
		expectedStatus string
		shouldFail     bool
	}{
		{"pending_to_active", epic.StatusPending, "wip", true},
		{"active_to_completed", epic.StatusWIP, "completed", true},
		{"completed_to_completed", epic.StatusCompleted, "completed", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			epic := &epic.Epic{ID: tc.name, Status: tc.epicStatus}
			result := &executor.TransitionChainResult{FinalState: epic}

			builder := NewAssertionBuilder(result).WithDebugMode(DebugBasic)
			builder.EpicStatus(tc.expectedStatus)

			errors := builder.GetErrors()
			hasErrors := len(errors) > 0

			if tc.shouldFail && !hasErrors {
				t.Error("Expected assertion to fail but it passed")
			} else if !tc.shouldFail && hasErrors {
				t.Errorf("Expected assertion to pass but got errors: %v", errors)
			}
		})
	}

	// Performance testing (note: proper benchmarks should be separate functions starting with Benchmark)
	// Basic performance validation
	epic := &epic.Epic{ID: "benchmark-test", Status: epic.StatusWIP}
	result := &executor.TransitionChainResult{FinalState: epic}

	start := time.Now()
	for i := 0; i < 1000; i++ {
		builder := NewAssertionBuilder(result).WithDebugMode(DebugOff)
		builder.EpicStatus("completed")
	}
	duration := time.Since(start)

	// Should be reasonably fast (less than 100ms for 1000 iterations)
	if duration > 100*time.Millisecond {
		t.Errorf("Performance test took too long: %v", duration)
	}
}
