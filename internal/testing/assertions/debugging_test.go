package assertions

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/testing/executor"
)

// === PHASE 4C: ERROR HANDLING & DEBUGGING SUPPORT TESTS ===

func TestErrorContext_IncludesRelevantStateInformation(t *testing.T) {
	// Create a test epic state
	testEpic := &epic.Epic{
		ID:     "test-epic",
		Name:   "Test Epic",
		Status: epic.StatusWIP,
	}

	// Create debug context with some trace entries
	debugCtx := NewDebugContext(DebugVerbose)
	debugCtx.Trace("INFO", "Starting test", map[string]interface{}{"epic_id": "test-epic"})
	debugCtx.Trace("WARN", "Potential issue detected", map[string]interface{}{"phase": "1A"})

	// Create error context
	err := errors.New("assertion failed: expected status 'completed' but got 'wip'")
	errorCtx := CreateErrorContext(err, "validation", 2, testEpic, debugCtx)

	// Verify error context contains relevant information
	if errorCtx.Err == nil {
		t.Error("Expected error to be set in error context")
	}

	if errorCtx.Stage != "validation" {
		t.Errorf("Expected stage 'validation', got '%s'", errorCtx.Stage)
	}

	if errorCtx.ChainIndex != 2 {
		t.Errorf("Expected chain index 2, got %d", errorCtx.ChainIndex)
	}

	// Check state information
	if _, exists := errorCtx.StateInfo["current_state"]; !exists {
		t.Error("Expected current_state in StateInfo")
	}

	if errorCtx.StateInfo["stage"] != "validation" {
		t.Error("Expected stage in StateInfo")
	}

	// Check debug trace is included
	if len(errorCtx.DebugTrace) != 2 {
		t.Errorf("Expected 2 debug trace entries, got %d", len(errorCtx.DebugTrace))
	}

	// Check suggestions are generated
	if len(errorCtx.Suggestions) == 0 {
		t.Error("Expected error suggestions to be generated")
	}

	// Verify error message formatting
	errorMsg := errorCtx.Error()
	if !strings.Contains(errorMsg, "validation") {
		t.Error("Expected error message to contain stage information")
	}
	if !strings.Contains(errorMsg, "step 2") {
		t.Error("Expected error message to contain chain step information")
	}
}

func TestDebugMode_ProvidesUsefulExecutionDetails(t *testing.T) {
	// Create assertion builder with debug mode enabled
	result := &executor.TransitionChainResult{
		FinalState: &epic.Epic{
			ID:     "debug-test",
			Status: epic.StatusWIP,
		},
		ExecutionTime: time.Millisecond * 150,
	}

	builder := NewAssertionBuilder(result).
		WithDebugMode(DebugVerbose).
		EnableStateVisualization()

	// Perform some assertions that will trigger debug logging
	builder.EpicStatus("completed") // This should fail and generate debug info

	// Check debug trace was captured
	trace := builder.GetDebugTrace()
	if len(trace) == 0 {
		t.Error("Expected debug trace entries to be captured")
	}

	// Check that debug trace contains error information
	found := false
	for _, entry := range trace {
		if entry.Level == "ERROR" && strings.Contains(entry.Message, "Expected epic status") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected debug trace to contain error information")
	}

	// Verify debug trace entries have required fields
	for _, entry := range trace {
		if entry.Timestamp.IsZero() {
			t.Error("Expected debug trace entry to have timestamp")
		}
		if entry.Location == "" {
			t.Error("Expected debug trace entry to have location information")
		}
		if entry.Level == "" {
			t.Error("Expected debug trace entry to have level")
		}
	}
}

func TestStateVisualization_HelpsUnderstandFailures(t *testing.T) {
	// Create a result with multiple states
	states := []interface{}{
		&epic.Epic{ID: "test", Status: epic.StatusPending},
		&epic.Epic{ID: "test", Status: epic.StatusWIP},
		&epic.Epic{ID: "test", Status: epic.StatusWIP}, // Still active - failure point
	}

	commands := []string{
		"start epic",
		"start phase 1A",
		"complete phase 1A", // This command failed to complete the phase
	}

	// Create state visualization
	viz := CreateStateVisualization(states, commands)

	// Check timeline visualization
	timeline := viz.GetTimelineVisualization()
	if !strings.Contains(timeline, "State Timeline:") {
		t.Error("Expected timeline to have header")
	}
	if !strings.Contains(timeline, "pending") {
		t.Error("Expected timeline to show initial pending status")
	}
	if !strings.Contains(timeline, "wip") {
		t.Error("Expected timeline to show active status")
	}

	// Check graph visualization
	graph := viz.GetGraphVisualization()
	if !strings.Contains(graph, "State Transition Graph:") {
		t.Error("Expected graph to have header")
	}
	if !strings.Contains(graph, "start epic") {
		t.Error("Expected graph to show commands")
	}
	if !strings.Contains(graph, "✓") {
		t.Error("Expected graph to show successful transitions")
	}

	// Verify state snapshots are captured
	if len(viz.Timeline) != 3 {
		t.Errorf("Expected 3 state snapshots, got %d", len(viz.Timeline))
	}

	// Verify graph structure
	if len(viz.Graph.Nodes) != 3 {
		t.Errorf("Expected 3 graph nodes, got %d", len(viz.Graph.Nodes))
	}
	if len(viz.Graph.Edges) != 2 {
		t.Errorf("Expected 2 graph edges, got %d", len(viz.Graph.Edges))
	}
}

func TestParallelExecution_MaintainsIsolation(t *testing.T) {
	// Create multiple assertion builders that could run in parallel
	results := make([]*executor.TransitionChainResult, 3)
	builders := make([]*AssertionBuilder, 3)

	for i := 0; i < 3; i++ {
		results[i] = &executor.TransitionChainResult{
			FinalState: &epic.Epic{
				ID:     fmt.Sprintf("parallel-test-%d", i),
				Status: epic.StatusWIP,
			},
		}
		builders[i] = NewAssertionBuilder(results[i]).WithDebugMode(DebugBasic)
	}

	// Run assertions in parallel-like manner
	done := make(chan bool, 3)

	for i, builder := range builders {
		go func(b *AssertionBuilder, index int) {
			// Each builder performs its own assertions
			b.EpicStatus("completed") // Will fail
			b.HasErrors()

			// Check that each builder has isolated state
			trace := b.GetDebugTrace()
			if len(trace) == 0 {
				t.Errorf("Builder %d should have debug trace", index)
			}

			errors := b.GetErrors()
			if len(errors) == 0 {
				t.Errorf("Builder %d should have assertion errors", index)
			}

			done <- true
		}(builder, i)
	}

	// Wait for all parallel executions to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify each builder maintained isolated state
	for i, builder := range builders {
		trace := builder.GetDebugTrace()
		errors := builder.GetErrors()

		if len(trace) == 0 {
			t.Errorf("Builder %d lost debug trace after parallel execution", i)
		}
		if len(errors) == 0 {
			t.Errorf("Builder %d lost assertion errors after parallel execution", i)
		}

		// Verify debug trace contains only this builder's entries
		for _, entry := range trace {
			if !strings.Contains(entry.Message, "Epic status assertion failed") {
				continue // Skip non-assertion entries
			}
			// Each builder should only see its own assertion failures
			if entry.Context != nil {
				if epicID, exists := entry.Context["epic_id"]; exists {
					expectedID := fmt.Sprintf("parallel-test-%d", i)
					if epicID != expectedID {
						t.Errorf("Builder %d seeing trace from wrong epic: %v", i, epicID)
					}
				}
			}
		}
	}
}

func TestTestFailureAnalysis_SuggestsSolutions(t *testing.T) {
	// Test different types of failures and their suggestions
	testCases := []struct {
		name                string
		error               error
		stage               string
		chainIndex          int
		expectedSuggestions []string
	}{
		{
			name:       "assertion failure",
			error:      errors.New("assertion failed: expected status completed"),
			stage:      "validation",
			chainIndex: 1,
			expectedSuggestions: []string{
				"Check if the expected state matches the actual result",
				"Review the transition chain logic leading to this assertion",
			},
		},
		{
			name:       "timeout error",
			error:      errors.New("timeout waiting for command completion"),
			stage:      "execution",
			chainIndex: 2,
			expectedSuggestions: []string{
				"Consider increasing timeout values for slow operations",
				"Check if there are blocking operations in the chain",
			},
		},
		{
			name:       "phase not found",
			error:      errors.New("phase not found: 1B"),
			stage:      "setup",
			chainIndex: 0,
			expectedSuggestions: []string{
				"Verify phase ID exists in the epic structure",
				"Check for typos in phase identifiers",
			},
		},
		{
			name:       "task not found",
			error:      errors.New("task not found: 1A_2"),
			stage:      "execution",
			chainIndex: 1,
			expectedSuggestions: []string{
				"Verify task ID exists in the specified phase",
				"Ensure task is properly defined in the epic XML",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errorCtx := CreateErrorContext(tc.error, tc.stage, tc.chainIndex, nil, nil)

			// Check that suggestions were generated
			if len(errorCtx.Suggestions) == 0 {
				t.Error("Expected error suggestions to be generated")
			}

			// Check for expected suggestions
			for _, expectedSuggestion := range tc.expectedSuggestions {
				found := false
				for _, actualSuggestion := range errorCtx.Suggestions {
					if strings.Contains(actualSuggestion, expectedSuggestion) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected suggestion '%s' not found in: %v",
						expectedSuggestion, errorCtx.Suggestions)
				}
			}

			// Check stage-specific suggestions
			for _, suggestion := range errorCtx.Suggestions {
				if strings.Contains(suggestion, tc.stage) {
					break
				}
			}

			// Chain index specific suggestions
			if tc.chainIndex > 0 {
				chainFound := false
				for _, suggestion := range errorCtx.Suggestions {
					if strings.Contains(suggestion, fmt.Sprintf("step %d", tc.chainIndex+1)) {
						chainFound = true
						break
					}
				}
				if !chainFound {
					t.Error("Expected chain index specific suggestion")
				}
			}
		})
	}
}

func TestErrorRecovery_AndContinuationStrategies(t *testing.T) {
	// Create a result with some errors
	result := &executor.TransitionChainResult{
		FinalState: &epic.Epic{
			ID:     "recovery-test",
			Status: epic.StatusWIP,
		},
		Errors: []executor.TransitionError{
			{
				Command:       "complete phase 1A",
				ExpectedState: "completed",
				ActualState:   "wip",
			},
		},
	}

	// Create custom recovery strategy
	customRecovery := &RecoveryStrategy{
		CanRecover: func(err error) bool {
			// Can recover from assertion failures
			errorMsg := err.Error()
			return strings.Contains(errorMsg, "assertion failed") ||
				strings.Contains(errorMsg, "Expected epic status") ||
				strings.Contains(errorMsg, "Phase") && strings.Contains(errorMsg, "not found")
		},
		RecoverFunc: func(err error, ctx *ErrorContext) error {
			// Simulate successful recovery
			if ctx != nil {
				ctx.Suggestions = append(ctx.Suggestions, "Recovery strategy applied")
			}
			return nil
		},
		ContinueFunc: func(ctx *ErrorContext) bool {
			// Continue unless critical error
			return ctx != nil && !strings.Contains(ctx.Err.Error(), "critical")
		},
	}

	builder := NewAssertionBuilder(result).
		WithDebugMode(DebugBasic).
		WithRecoveryStrategy(customRecovery)

	// Trigger some assertion failures
	builder.EpicStatus("completed")        // Should fail but be recoverable
	builder.PhaseStatus("1A", "completed") // Should fail but be recoverable

	// Check initial errors
	initialErrorCount := len(builder.GetErrors())
	if initialErrorCount == 0 {
		t.Error("Expected initial assertion errors")
	}

	// Debug: Print initial errors to understand what we're working with
	t.Logf("Initial errors (%d): %v", initialErrorCount, builder.GetErrors())

	// Debug: Check if our recovery strategy recognizes these errors
	for i, err := range builder.GetErrors() {
		testErr := fmt.Errorf("%s", err.Message)
		canRecover := customRecovery.CanRecover(testErr)
		t.Logf("Error %d '%s' can be recovered: %v", i, err.Message, canRecover)
	}

	// Attempt recovery
	builder.RecoverFromErrors()

	// Check if errors were reduced after recovery
	finalErrorCount := len(builder.GetErrors())
	t.Logf("Final errors (%d): %v", finalErrorCount, builder.GetErrors())

	if finalErrorCount >= initialErrorCount {
		t.Errorf("Expected error recovery to reduce error count from %d to less, but got %d", initialErrorCount, finalErrorCount)
	}

	// Test non-recoverable error
	builder.addErrorWithContext("critical_error", "critical system failure",
		"working", "broken", nil, "execution", 1)

	preRecoveryCount := len(builder.GetErrors())
	builder.RecoverFromErrors()
	postRecoveryCount := len(builder.GetErrors())

	// Critical errors should not be recovered
	if postRecoveryCount < preRecoveryCount {
		t.Error("Critical errors should not be recoverable")
	}
}

func TestDebugInfo_PrintsUsefulInformation(t *testing.T) {
	// Create a result with state visualization enabled
	result := &executor.TransitionChainResult{
		InitialState: &epic.Epic{
			ID:     "debug-print-test",
			Status: epic.StatusPending,
		},
		FinalState: &epic.Epic{
			ID:     "debug-print-test",
			Status: epic.StatusWIP,
		},
		IntermediateStates: []executor.StateSnapshot{
			{EpicState: &epic.Epic{ID: "debug-print-test", Status: epic.StatusWIP}},
		},
		ExecutedCommands: []executor.CommandExecution{
			{
				Command: executor.ChainCommand{
					Type:   "start",
					Target: "epic",
				},
				Success: true,
			},
			{
				Command: executor.ChainCommand{
					Type:   "start",
					Target: "phase 1A",
				},
				Success: true,
			},
		},
	}

	builder := NewAssertionBuilder(result).
		WithDebugMode(DebugVerbose).
		EnableStateVisualization()

	// Trigger some assertions for debug output
	builder.EpicStatus("completed").
		HasErrors()

	// Capture debug output (in real usage this would print to stdout)
	trace := builder.GetDebugTrace()
	viz := builder.GetStateVisualization()
	errors := builder.GetErrors()

	// Verify debug information is comprehensive
	if len(trace) == 0 {
		t.Error("Expected debug trace entries")
	}

	if viz == nil {
		t.Error("Expected state visualization to be enabled")
	}

	if len(errors) == 0 {
		t.Error("Expected assertion errors for debug output")
	}

	// Verify visualization contains useful information
	if viz != nil {
		graphViz := viz.GetGraphVisualization()
		t.Logf("Graph visualization: %s", graphViz)

		// Since the test result has ExecutedCommands, check for their presence
		if !strings.Contains(graphViz, "start") {
			t.Error("Expected graph visualization to show commands")
		}

		timeline := viz.GetTimelineVisualization()
		if !strings.Contains(timeline, "State Timeline:") {
			t.Error("Expected timeline visualization to have proper format")
		}
	}

	// Test PrintDebugInfo doesn't panic (we can't easily capture stdout in tests)
	// but we can verify it completes without error
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintDebugInfo panicked: %v", r)
		}
	}()

	builder.PrintDebugInfo()
}

func TestDefaultRecoveryStrategy_BehavesCorrectly(t *testing.T) {
	strategy := DefaultRecoveryStrategy()

	// Test recoverable errors
	recoverableErrors := []error{
		errors.New("assertion failed: status mismatch"),
		errors.New("validation failed: phase not completed"),
	}

	for _, err := range recoverableErrors {
		if !strategy.CanRecover(err) {
			t.Errorf("Expected error to be recoverable: %s", err.Error())
		}
	}

	// Test non-recoverable errors
	nonRecoverableErrors := []error{
		errors.New("file not found"),
		errors.New("network connection failed"),
		errors.New("system error occurred"),
	}

	for _, err := range nonRecoverableErrors {
		if strategy.CanRecover(err) {
			t.Errorf("Expected error to not be recoverable: %s", err.Error())
		}
	}

	// Test continue function
	continueableErrors := []error{
		errors.New("assertion failed"),
		errors.New("validation warning"),
	}

	for _, err := range continueableErrors {
		ctx := &ErrorContext{Err: err}
		if !strategy.ContinueFunc(ctx) {
			t.Errorf("Expected to continue after error: %s", err.Error())
		}
	}

	// Test non-continueable errors
	criticalErrors := []error{
		errors.New("critical system failure"),
		errors.New("fatal error occurred"),
	}

	for _, err := range criticalErrors {
		ctx := &ErrorContext{Err: err}
		if strategy.ContinueFunc(ctx) {
			t.Errorf("Expected to not continue after critical error: %s", err.Error())
		}
	}
}

func TestDebugContext_TracesCorrectly(t *testing.T) {
	// Test different debug modes
	testCases := []struct {
		name     string
		mode     DebugMode
		expected bool
	}{
		{"debug off", DebugOff, false},
		{"debug basic", DebugBasic, true},
		{"debug verbose", DebugVerbose, true},
		{"debug trace", DebugTrace, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := NewDebugContext(tc.mode)

			// Add some trace entries
			ctx.Trace("INFO", "Test message", map[string]interface{}{"key": "value"})
			ctx.Trace("ERROR", "Error message", nil)

			trace := ctx.GetTraceLog()

			if tc.expected && len(trace) == 0 {
				t.Error("Expected trace entries to be captured")
			} else if !tc.expected && len(trace) > 0 {
				t.Error("Expected no trace entries when debug is off")
			}

			// Test trace entry structure
			if tc.expected {
				for _, entry := range trace {
					if entry.Timestamp.IsZero() {
						t.Error("Expected trace entry to have timestamp")
					}
					if entry.Level == "" {
						t.Error("Expected trace entry to have level")
					}
					if entry.Message == "" {
						t.Error("Expected trace entry to have message")
					}
					if entry.Location == "" {
						t.Error("Expected trace entry to have location")
					}
				}
			}
		})
	}
}
