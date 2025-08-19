package assertions

import (
	"strings"
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/testing/builders"
	"github.com/mindreframer/agentpm/internal/testing/executor"
)

func createTestResult(t *testing.T) *executor.TransitionChainResult {
	// Create environment and epic for testing
	env := executor.NewTestExecutionEnvironment("test-epic.xml")

	testEpic, err := builders.NewEpicBuilder("test-epic").
		WithStatus("pending").
		WithPhase("1A", "Setup", "pending").
		WithPhase("1B", "Development", "pending").
		WithTask("1A_1", "1A", "Initialize Project", "pending").
		WithTask("1A_2", "1A", "Configure Tools", "pending").
		WithTask("1B_1", "1B", "Implement Feature", "pending").
		WithTest("T1A_1", "1A_1", "1A", "Test Project Init", "pending").
		WithTest("T1A_2", "1A_2", "1A", "Test Tool Config", "pending").
		WithTest("T1B_1", "1B_1", "1B", "Test Feature", "pending").
		Build()

	if err != nil {
		t.Fatalf("Failed to build test epic: %v", err)
	}

	err = env.LoadEpic(testEpic)
	if err != nil {
		t.Fatalf("Failed to load epic: %v", err)
	}

	// Execute partial workflow
	result, err := executor.CreateTransitionChain(env).
		StartEpic().
		StartPhase("1A").
		StartTask("1A_1").
		PassTest("T1A_1").
		DoneTask("1A_1").
		StartTask("1A_2").
		PassTest("T1A_2").
		DoneTask("1A_2").
		DonePhase("1A").
		Execute()

	if err != nil {
		t.Fatalf("Failed to execute transition chain: %v", err)
	}

	return result
}

func TestAssertionBuilder_EpicStatus_Success(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		EpicStatus("wip").
		Check()

	if err != nil {
		t.Errorf("Expected assertion to pass, got error: %v", err)
	}
}

func TestAssertionBuilder_EpicStatus_Failure(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		EpicStatus("completed").
		Check()

	if err == nil {
		t.Error("Expected assertion to fail")
	}

	expectedMsg := "Expected epic status completed, got wip"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestAssertionBuilder_PhaseStatus_Success(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		PhaseStatus("1A", "completed").
		PhaseStatus("1B", "pending").
		Check()

	if err != nil {
		t.Errorf("Expected assertion to pass, got error: %v", err)
	}
}

func TestAssertionBuilder_PhaseStatus_Failure(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		PhaseStatus("1A", "pending").
		Check()

	if err == nil {
		t.Error("Expected assertion to fail")
	}

	expectedMsg := "Expected phase 1A status pending, got completed"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestAssertionBuilder_PhaseStatus_NotFound(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		PhaseStatus("INVALID", "pending").
		Check()

	if err == nil {
		t.Error("Expected assertion to fail for non-existent phase")
	}

	expectedMsg := "Phase INVALID not found"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestAssertionBuilder_TaskStatus_Success(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		TaskStatus("1A_1", "completed").
		TaskStatus("1A_2", "completed").
		TaskStatus("1B_1", "pending").
		Check()

	if err != nil {
		t.Errorf("Expected assertion to pass, got error: %v", err)
	}
}

func TestAssertionBuilder_TaskStatus_Failure(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		TaskStatus("1A_1", "pending").
		Check()

	if err == nil {
		t.Error("Expected assertion to fail")
	}

	expectedMsg := "Expected task 1A_1 status pending, got completed"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestAssertionBuilder_TestStatus_Success(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		TestStatus("T1A_1", "completed").
		TestStatus("T1A_2", "completed").
		TestStatus("T1B_1", "pending").
		Check()

	if err != nil {
		t.Errorf("Expected assertion to pass, got error: %v", err)
	}
}

func TestAssertionBuilder_TestStatusUnified_Success(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		TestStatusUnified("T1A_1", "done").
		TestStatusUnified("T1A_2", "done").
		TestStatusUnified("T1B_1", "pending").
		Check()

	if err != nil {
		t.Errorf("Expected assertion to pass, got error: %v", err)
	}
}

func TestAssertionBuilder_TestResult_Success(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		TestResult("T1A_1", "passing").
		TestResult("T1A_2", "passing").
		Check()

	if err != nil {
		t.Errorf("Expected assertion to pass, got error: %v", err)
	}
}

func TestAssertionBuilder_HasEvent_Success(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		HasEvent("epic_started").
		HasEvent("phase_started").
		HasEvent("task_started").
		HasEvent("test_passed").
		Check()

	if err != nil {
		t.Errorf("Expected assertion to pass, got error: %v", err)
	}
}

func TestAssertionBuilder_HasEvent_Failure(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		HasEvent("epic_completed").
		Check()

	if err == nil {
		t.Error("Expected assertion to fail")
	}

	expectedMsg := "Expected event type epic_completed not found"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestAssertionBuilder_EventCount_Success(t *testing.T) {
	result := createTestResult(t)

	// Count the expected events from the workflow
	// epic_started, phase_started, task_started, test_passed, task_completed, task_started, test_passed, task_completed, phase_completed
	expectedEventCount := len(result.FinalState.Events)

	err := Assert(result).
		EventCount(expectedEventCount).
		Check()

	if err != nil {
		t.Errorf("Expected assertion to pass, got error: %v", err)
	}
}

func TestAssertionBuilder_EventCount_Failure(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		EventCount(999).
		Check()

	if err == nil {
		t.Error("Expected assertion to fail")
	}

	expectedMsg := "Expected 999 events, got"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestAssertionBuilder_NoErrors_Success(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		NoErrors().
		Check()

	if err != nil {
		t.Errorf("Expected assertion to pass, got error: %v", err)
	}
}

func TestAssertionBuilder_NoErrors_Failure(t *testing.T) {
	// Create a result with errors
	env := executor.NewTestExecutionEnvironment("test-epic.xml")
	testEpic, _ := builders.NewEpicBuilder("test-epic").
		WithStatus("pending").
		WithPhase("1A", "Setup", "pending").
		Build()

	env.LoadEpic(testEpic)

	// Try to start non-existent phase to create error
	result, _ := executor.CreateTransitionChain(env).
		StartEpic().
		StartPhase("INVALID_PHASE").
		Execute()

	err := Assert(result).
		NoErrors().
		Check()

	if err == nil {
		t.Error("Expected assertion to fail when there are errors")
	}

	expectedMsg := "Expected no errors, but got"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestAssertionBuilder_HasErrors_Success(t *testing.T) {
	// Create a result with errors
	env := executor.NewTestExecutionEnvironment("test-epic.xml")
	testEpic, _ := builders.NewEpicBuilder("test-epic").
		WithStatus("pending").
		WithPhase("1A", "Setup", "pending").
		Build()

	env.LoadEpic(testEpic)

	result, _ := executor.CreateTransitionChain(env).
		StartEpic().
		StartPhase("INVALID_PHASE").
		Execute()

	err := Assert(result).
		HasErrors().
		Check()

	if err != nil {
		t.Errorf("Expected assertion to pass, got error: %v", err)
	}
}

func TestAssertionBuilder_HasErrors_Failure(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		HasErrors().
		Check()

	if err == nil {
		t.Error("Expected assertion to fail when there are no errors")
	}

	expectedMsg := "Expected errors, but execution was successful"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestAssertionBuilder_ErrorCount_Success(t *testing.T) {
	// Create a result with specific number of errors
	env := executor.NewTestExecutionEnvironment("test-epic.xml")
	testEpic, _ := builders.NewEpicBuilder("test-epic").
		WithStatus("pending").
		WithPhase("1A", "Setup", "pending").
		Build()

	env.LoadEpic(testEpic)

	result, _ := executor.CreateTransitionChain(env).
		StartEpic().
		StartPhase("INVALID_PHASE_1").
		StartPhase("INVALID_PHASE_2").
		Execute()

	expectedErrorCount := len(result.Errors)

	err := Assert(result).
		ErrorCount(expectedErrorCount).
		Check()

	if err != nil {
		t.Errorf("Expected assertion to pass, got error: %v", err)
	}
}

func TestAssertionBuilder_ExecutionTime_Success(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		ExecutionTime(time.Second). // Should be much less than 1 second
		Check()

	if err != nil {
		t.Errorf("Expected assertion to pass, got error: %v", err)
	}
}

func TestAssertionBuilder_ExecutionTime_Failure(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		ExecutionTime(time.Nanosecond). // Very small time limit
		Check()

	if err == nil {
		t.Error("Expected assertion to fail for very small time limit")
	}

	expectedMsg := "Expected execution time <="
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestAssertionBuilder_CommandCount_Success(t *testing.T) {
	result := createTestResult(t)

	expectedCommandCount := len(result.ExecutedCommands)

	err := Assert(result).
		CommandCount(expectedCommandCount).
		Check()

	if err != nil {
		t.Errorf("Expected assertion to pass, got error: %v", err)
	}
}

func TestAssertionBuilder_CommandCount_Failure(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		CommandCount(999).
		Check()

	if err == nil {
		t.Error("Expected assertion to fail")
	}

	expectedMsg := "Expected 999 commands, got"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestAssertionBuilder_AllCommandsSuccessful_Success(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		AllCommandsSuccessful().
		Check()

	if err != nil {
		t.Errorf("Expected assertion to pass, got error: %v", err)
	}
}

func TestAssertionBuilder_AllCommandsSuccessful_Failure(t *testing.T) {
	// Create a result with failed commands
	env := executor.NewTestExecutionEnvironment("test-epic.xml")
	testEpic, _ := builders.NewEpicBuilder("test-epic").
		WithStatus("pending").
		WithPhase("1A", "Setup", "pending").
		Build()

	env.LoadEpic(testEpic)

	result, _ := executor.CreateTransitionChain(env).
		StartEpic().
		StartPhase("INVALID_PHASE").
		Execute()

	err := Assert(result).
		AllCommandsSuccessful().
		Check()

	if err == nil {
		t.Error("Expected assertion to fail when there are failed commands")
	}

	expectedMsg := "Expected all commands to succeed, but"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestAssertionBuilder_ChainedAssertions_Success(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		EpicStatus("wip").
		PhaseStatus("1A", "completed").
		PhaseStatus("1B", "pending").
		TaskStatus("1A_1", "completed").
		TaskStatus("1A_2", "completed").
		TestStatusUnified("T1A_1", "done").
		TestStatusUnified("T1A_2", "done").
		TestResult("T1A_1", "passing").
		TestResult("T1A_2", "passing").
		HasEvent("epic_started").
		HasEvent("phase_completed").
		NoErrors().
		AllCommandsSuccessful().
		Check()

	if err != nil {
		t.Errorf("Expected chained assertions to pass, got error: %v", err)
	}
}

func TestAssertionBuilder_ChainedAssertions_MultipleFailures(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		EpicStatus("completed").        // Wrong - should be active
		PhaseStatus("1A", "pending").   // Wrong - should be completed
		PhaseStatus("1B", "completed"). // Wrong - should be pending
		HasEvent("epic_completed").     // Wrong - event doesn't exist
		ErrorCount(5).                  // Wrong - should be 0
		Check()

	if err == nil {
		t.Error("Expected assertions to fail")
	}

	// Should be a composite error with multiple failures
	if compositeErr, ok := err.(*CompositeAssertionError); ok {
		if compositeErr.Count != 5 {
			t.Errorf("Expected 5 assertion failures, got: %d", compositeErr.Count)
		}
	} else {
		t.Errorf("Expected CompositeAssertionError, got: %T", err)
	}
}

func TestAssertionBuilder_MustPass_Success(t *testing.T) {
	result := createTestResult(t)

	// Should not panic
	Assert(result).
		EpicStatus("wip").
		NoErrors().
		MustPass()
}

func TestAssertionBuilder_MustPass_Panic(t *testing.T) {
	result := createTestResult(t)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected MustPass to panic on assertion failure")
		}
	}()

	Assert(result).
		EpicStatus("completed"). // Wrong status
		MustPass()
}

func TestAssertionBuilder_NilFinalState(t *testing.T) {
	// Create result with nil final state
	result := &executor.TransitionChainResult{
		FinalState:       nil,
		ExecutedCommands: []executor.CommandExecution{},
		Errors:           []executor.TransitionError{},
		Success:          false,
	}

	err := Assert(result).
		EpicStatus("wip").
		Check()

	if err == nil {
		t.Error("Expected assertion to fail with nil final state")
	}

	expectedMsg := "Final state is nil"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestAssertionBuilder_AssertionError_Details(t *testing.T) {
	result := createTestResult(t)

	err := Assert(result).
		EpicStatus("completed").
		Check()

	if err == nil {
		t.Error("Expected assertion to fail")
	}

	if assertionErr, ok := err.(AssertionError); ok {
		if assertionErr.Type != "epic_status" {
			t.Errorf("Expected error type 'epic_status', got: %s", assertionErr.Type)
		}

		if assertionErr.Expected != "completed" {
			t.Errorf("Expected 'completed', got: %v", assertionErr.Expected)
		}

		if assertionErr.Actual != "wip" {
			t.Errorf("Expected 'wip', got: %v", assertionErr.Actual)
		}

		if assertionErr.Context == nil {
			t.Error("Expected context to be set")
		}

		if epicID, ok := assertionErr.Context["epic_id"]; !ok || epicID != "test-epic" {
			t.Errorf("Expected epic_id in context to be 'test-epic', got: %v", epicID)
		}
	} else {
		t.Errorf("Expected AssertionError, got: %T", err)
	}
}

func TestAssertionBuilder_ComplexScenario(t *testing.T) {
	// Create a more complex test scenario
	env := executor.NewTestExecutionEnvironment("complex-epic.xml")

	complexEpic, err := builders.NewEpicBuilder("complex-epic").
		WithStatus("pending").
		WithAssignee("agent_claude").
		WithPhase("1A", "Setup", "pending").
		WithPhase("1B", "Development", "pending").
		WithPhase("1C", "Testing", "pending").
		WithTask("1A_1", "1A", "Initialize", "pending").
		WithTask("1A_2", "1A", "Configure", "pending").
		WithTask("1B_1", "1B", "Implement Core", "pending").
		WithTask("1B_2", "1B", "Implement Features", "pending").
		WithTask("1C_1", "1C", "Unit Tests", "pending").
		WithTask("1C_2", "1C", "Integration Tests", "pending").
		WithTest("T1A_1", "1A_1", "1A", "Test Initialize", "pending").
		WithTest("T1A_2", "1A_2", "1A", "Test Configure", "pending").
		WithTest("T1B_1", "1B_1", "1B", "Test Core", "pending").
		WithTest("T1C_1", "1C_1", "1C", "Test Units", "pending").
		Build()

	if err != nil {
		t.Fatalf("Failed to build complex epic: %v", err)
	}

	err = env.LoadEpic(complexEpic)
	if err != nil {
		t.Fatalf("Failed to load complex epic: %v", err)
	}

	// Execute complete workflow
	result, err := executor.CreateTransitionChain(env).
		StartEpic().
		StartPhase("1A").
		StartTask("1A_1").
		PassTest("T1A_1").
		DoneTask("1A_1").
		StartTask("1A_2").
		PassTest("T1A_2").
		DoneTask("1A_2").
		DonePhase("1A").
		StartPhase("1B").
		StartTask("1B_1").
		PassTest("T1B_1").
		DoneTask("1B_1").
		Execute()

	if err != nil {
		t.Fatalf("Failed to execute complex workflow: %v", err)
	}

	// Comprehensive assertions
	err = Assert(result).
		EpicStatus("wip").
		PhaseStatus("1A", "completed").
		PhaseStatus("1B", "wip").
		PhaseStatus("1C", "pending").
		TaskStatus("1A_1", "completed").
		TaskStatus("1A_2", "completed").
		TaskStatus("1B_1", "completed").
		TaskStatus("1B_2", "pending").
		TaskStatus("1C_1", "pending").
		TaskStatus("1C_2", "pending").
		TestStatusUnified("T1A_1", "done").
		TestStatusUnified("T1A_2", "done").
		TestStatusUnified("T1B_1", "done").
		TestStatusUnified("T1C_1", "pending").
		TestResult("T1A_1", "passing").
		TestResult("T1A_2", "passing").
		TestResult("T1B_1", "passing").
		HasEvent("epic_started").
		HasEvent("phase_started").
		HasEvent("task_started").
		HasEvent("test_passed").
		HasEvent("task_completed").
		HasEvent("phase_completed").
		NoErrors().
		AllCommandsSuccessful().
		ExecutionTime(time.Second).
		Check()

	if err != nil {
		t.Errorf("Complex scenario assertions failed: %v", err)
	}
}

func TestAssertionBuilder_PerformanceValidation(t *testing.T) {
	result := createTestResult(t)

	start := time.Now()

	// Run assertions multiple times to check performance
	for i := 0; i < 100; i++ {
		err := Assert(result).
			EpicStatus("wip").
			PhaseStatus("1A", "completed").
			TaskStatus("1A_1", "completed").
			TestStatusUnified("T1A_1", "done").
			NoErrors().
			Check()

		if err != nil {
			t.Errorf("Assertion failed on iteration %d: %v", i, err)
		}
	}

	duration := time.Since(start)

	// Assertions should be fast - 100 iterations in less than 100ms
	if duration > 100*time.Millisecond {
		t.Errorf("Assertions took too long: %v", duration)
	}
}
