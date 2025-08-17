package executor

import (
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/testing/builders"
)

func TestTransitionChain_BasicWorkflow(t *testing.T) {
	// Create test environment
	env := NewTestExecutionEnvironment("test-epic.xml")

	// Build test epic
	testEpic, err := builders.NewEpicBuilder("test-epic").
		WithStatus("planning").
		WithPhase("1A", "Setup", "pending").
		WithTask("1A_1", "1A", "Initialize Project", "pending").
		WithTest("T1A_1", "1A_1", "1A", "Test Project Init", "pending").
		Build()

	if err != nil {
		t.Fatalf("Failed to build test epic: %v", err)
	}

	// Load epic into environment
	err = env.LoadEpic(testEpic)
	if err != nil {
		t.Fatalf("Failed to load epic: %v", err)
	}

	// Execute transition chain
	result, err := CreateTransitionChain(env).
		StartEpic().
		Execute()

	if err != nil {
		t.Fatalf("Failed to execute transition chain: %v", err)
	}

	// Verify results
	if !result.Success {
		t.Errorf("Expected successful execution, got errors: %v", result.Errors)
	}

	if len(result.ExecutedCommands) != 1 {
		t.Errorf("Expected 1 command executed, got: %d", len(result.ExecutedCommands))
	}

	// Verify epic status changed
	if result.FinalState.Status != epic.StatusActive {
		t.Errorf("Expected epic status to be active, got: %s", result.FinalState.Status)
	}

	// Verify command execution details
	cmd := result.ExecutedCommands[0]
	if cmd.Command.Type != "start_epic" {
		t.Errorf("Expected command type 'start_epic', got: %s", cmd.Command.Type)
	}

	if !cmd.Success {
		t.Errorf("Expected command success, got error: %v", cmd.Error)
	}
}

func TestTransitionChain_PhaseWorkflow(t *testing.T) {
	env := NewTestExecutionEnvironment("test-epic.xml")

	// Build test epic with active epic (already started)
	testEpic, err := builders.NewEpicBuilder("test-epic").
		WithStatus("active").
		WithPhase("1A", "Setup", "pending").
		WithTask("1A_1", "1A", "Initialize Project", "pending").
		Build()

	if err != nil {
		t.Fatalf("Failed to build test epic: %v", err)
	}

	err = env.LoadEpic(testEpic)
	if err != nil {
		t.Fatalf("Failed to load epic: %v", err)
	}

	// Execute phase workflow
	result, err := CreateTransitionChain(env).
		StartPhase("1A").
		Execute()

	if err != nil {
		t.Fatalf("Failed to execute transition chain: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got errors: %v", result.Errors)
	}

	// Verify phase status changed
	phase1A := findPhaseByID(result.FinalState, "1A")
	if phase1A == nil {
		t.Fatal("Phase 1A not found in final state")
	}

	if phase1A.Status != epic.StatusActive {
		t.Errorf("Expected phase 1A status to be active, got: %s", phase1A.Status)
	}

	// Verify phase has started timestamp
	if phase1A.StartedAt == nil {
		t.Error("Expected phase 1A to have StartedAt timestamp")
	}
}

func TestTransitionChain_TaskWorkflow(t *testing.T) {
	env := NewTestExecutionEnvironment("test-epic.xml")

	// Build test epic with active phase
	testEpic, err := builders.NewEpicBuilder("test-epic").
		WithStatus("active").
		WithPhase("1A", "Setup", "active").
		WithTask("1A_1", "1A", "Initialize Project", "pending").
		Build()

	if err != nil {
		t.Fatalf("Failed to build test epic: %v", err)
	}

	err = env.LoadEpic(testEpic)
	if err != nil {
		t.Fatalf("Failed to load epic: %v", err)
	}

	// Execute task workflow
	result, err := CreateTransitionChain(env).
		StartTask("1A_1").
		DoneTask("1A_1").
		Execute()

	if err != nil {
		t.Fatalf("Failed to execute transition chain: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got errors: %v", result.Errors)
	}

	// Verify task status progression
	task1A1 := findTaskByID(result.FinalState, "1A_1")
	if task1A1 == nil {
		t.Fatal("Task 1A_1 not found in final state")
	}

	if task1A1.Status != epic.StatusCompleted {
		t.Errorf("Expected task 1A_1 status to be completed, got: %s", task1A1.Status)
	}

	// Verify both start and completion timestamps
	if task1A1.StartedAt == nil {
		t.Error("Expected task 1A_1 to have StartedAt timestamp")
	}

	if task1A1.CompletedAt == nil {
		t.Error("Expected task 1A_1 to have CompletedAt timestamp")
	}
}

func TestTransitionChain_TestWorkflow(t *testing.T) {
	env := NewTestExecutionEnvironment("test-epic.xml")

	// Build test epic with active task
	testEpic, err := builders.NewEpicBuilder("test-epic").
		WithStatus("active").
		WithPhase("1A", "Setup", "active").
		WithTask("1A_1", "1A", "Initialize Project", "active").
		WithTest("T1A_1", "1A_1", "1A", "Test Project Init", "pending").
		Build()

	if err != nil {
		t.Fatalf("Failed to build test epic: %v", err)
	}

	err = env.LoadEpic(testEpic)
	if err != nil {
		t.Fatalf("Failed to load epic: %v", err)
	}

	// Execute test workflow - pass test
	result, err := CreateTransitionChain(env).
		PassTest("T1A_1").
		Execute()

	if err != nil {
		t.Fatalf("Failed to execute transition chain: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got errors: %v", result.Errors)
	}

	// Get updated state from storage (since test service operates on storage directly)
	finalEpic, err := env.GetCurrentEpic()
	if err != nil {
		t.Fatalf("Failed to get final epic state: %v", err)
	}

	// Verify test status
	test1A1 := findTestByID(finalEpic, "T1A_1")
	if test1A1 == nil {
		t.Fatal("Test T1A_1 not found in final state")
	}

	if test1A1.Status != epic.StatusCompleted {
		t.Errorf("Expected test T1A_1 status to be completed, got: %s", test1A1.Status)
	}

	// Verify Epic 13 unified status
	if test1A1.GetTestStatusUnified() != epic.TestStatusDone {
		t.Errorf("Expected test T1A_1 unified status to be done, got: %s", test1A1.GetTestStatusUnified())
	}

	if test1A1.GetTestResult() != epic.TestResultPassing {
		t.Errorf("Expected test T1A_1 result to be passing, got: %s", test1A1.GetTestResult())
	}
}

func TestTransitionChain_FailedTestWorkflow(t *testing.T) {
	env := NewTestExecutionEnvironment("test-epic.xml")

	// Build test epic with active task and a test in WIP status
	testEpic, err := builders.NewEpicBuilder("test-epic").
		WithStatus("active").
		WithPhase("1A", "Setup", "active").
		WithTask("1A_1", "1A", "Initialize Project", "active").
		WithTestDescriptive("T1A_1", "1A_1", "1A", "Test Project Init", "Test description", "active", "wip", "passing").
		Build()

	if err != nil {
		t.Fatalf("Failed to build test epic: %v", err)
	}

	err = env.LoadEpic(testEpic)
	if err != nil {
		t.Fatalf("Failed to load epic: %v", err)
	}

	// Execute test workflow - fail test
	result, err := CreateTransitionChain(env).
		FailTest("T1A_1").
		Execute()

	if err != nil {
		t.Fatalf("Failed to execute transition chain: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got errors: %v", result.Errors)
	}

	// Get updated state from storage
	finalEpic, err := env.GetCurrentEpic()
	if err != nil {
		t.Fatalf("Failed to get final epic state: %v", err)
	}

	// Verify test failure
	test1A1 := findTestByID(finalEpic, "T1A_1")
	if test1A1 == nil {
		t.Fatal("Test T1A_1 not found in final state")
	}

	// Test should be in WIP status with failing result
	if test1A1.GetTestStatusUnified() != epic.TestStatusWIP {
		t.Errorf("Expected test T1A_1 unified status to be wip, got: %s", test1A1.GetTestStatusUnified())
	}

	if test1A1.GetTestResult() != epic.TestResultFailing {
		t.Errorf("Expected test T1A_1 result to be failing, got: %s", test1A1.GetTestResult())
	}

	if test1A1.FailedAt == nil {
		t.Error("Expected test T1A_1 to have FailedAt timestamp")
	}
}

func TestTransitionChain_CompleteWorkflow(t *testing.T) {
	env := NewTestExecutionEnvironment("test-epic.xml")

	// Build comprehensive test epic
	testEpic, err := builders.NewEpicBuilder("test-epic").
		WithStatus("planning").
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

	// Execute complete workflow
	result, err := CreateTransitionChain(env).
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
		DonePhase("1B").
		DoneEpic().
		Execute()

	if err != nil {
		t.Fatalf("Failed to execute transition chain: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got errors: %v", result.Errors)
	}

	// Verify complete execution
	expectedCommands := 15 // All the commands above: start_epic + start_phase + start_task + pass_test + done_task + start_task + pass_test + done_task + done_phase + start_phase + start_task + pass_test + done_task + done_phase + done_epic
	if len(result.ExecutedCommands) != expectedCommands {
		t.Errorf("Expected %d commands executed, got: %d", expectedCommands, len(result.ExecutedCommands))
		// Debug: print all executed commands
		for i, cmd := range result.ExecutedCommands {
			t.Logf("Command %d: %s (%s) - Success: %v", i+1, cmd.Command.Type, cmd.Command.Target, cmd.Success)
		}
	}

	// Get final state
	finalEpic, err := env.GetCurrentEpic()
	if err != nil {
		t.Fatalf("Failed to get final epic state: %v", err)
	}

	// Verify final epic status
	if finalEpic.Status != epic.StatusCompleted {
		t.Errorf("Expected epic status to be completed, got: %s", finalEpic.Status)
	}

	// Verify all phases completed
	for _, phase := range finalEpic.Phases {
		if phase.Status != epic.StatusCompleted {
			t.Errorf("Expected phase %s to be completed, got: %s", phase.ID, phase.Status)
		}
	}

	// Verify all tasks completed
	for _, task := range finalEpic.Tasks {
		if task.Status != epic.StatusCompleted {
			t.Errorf("Expected task %s to be completed, got: %s", task.ID, task.Status)
		}
	}

	// Verify all tests passed
	for _, test := range finalEpic.Tests {
		if test.GetTestStatusUnified() != epic.TestStatusDone {
			t.Errorf("Expected test %s to be done, got: %s", test.ID, test.GetTestStatusUnified())
		}
		if test.GetTestResult() != epic.TestResultPassing {
			t.Errorf("Expected test %s to be passing, got: %s", test.ID, test.GetTestResult())
		}
	}

	// Verify execution time tracking
	if result.ExecutionTime <= 0 {
		t.Error("Expected positive execution time")
	}

	// Verify snapshots captured
	if len(result.IntermediateStates) == 0 {
		t.Error("Expected intermediate states to be captured")
	}
}

func TestTransitionChain_ErrorHandling(t *testing.T) {
	env := NewTestExecutionEnvironment("test-epic.xml")

	// Build test epic
	testEpic, err := builders.NewEpicBuilder("test-epic").
		WithStatus("planning").
		WithPhase("1A", "Setup", "pending").
		Build()

	if err != nil {
		t.Fatalf("Failed to build test epic: %v", err)
	}

	err = env.LoadEpic(testEpic)
	if err != nil {
		t.Fatalf("Failed to load epic: %v", err)
	}

	// Try to start non-existent phase
	result, err := CreateTransitionChain(env).
		StartEpic().
		StartPhase("INVALID_PHASE").
		Execute()

	if err != nil {
		t.Fatalf("Failed to execute transition chain: %v", err)
	}

	// Should have errors but still return result
	if result.Success {
		t.Error("Expected failed execution due to invalid phase")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected errors in result")
	}

	// Check error details
	if len(result.ExecutedCommands) != 2 {
		t.Errorf("Expected 2 commands attempted, got: %d", len(result.ExecutedCommands))
	}

	// First command (start epic) should succeed
	if !result.ExecutedCommands[0].Success {
		t.Error("Expected first command (start epic) to succeed")
	}

	// Second command (start invalid phase) should fail
	if result.ExecutedCommands[1].Success {
		t.Error("Expected second command (start invalid phase) to fail")
	}
}

func TestTransitionChain_TimestampControl(t *testing.T) {
	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	env := NewTestExecutionEnvironment("test-epic.xml").WithTimeSource(func() time.Time { return fixedTime })

	// Build test epic
	testEpic, err := builders.NewEpicBuilder("test-epic").
		WithStatus("planning").
		WithPhase("1A", "Setup", "pending").
		Build()

	if err != nil {
		t.Fatalf("Failed to build test epic: %v", err)
	}

	err = env.LoadEpic(testEpic)
	if err != nil {
		t.Fatalf("Failed to load epic: %v", err)
	}

	// Execute with specific timestamps
	customTime := time.Date(2023, 6, 15, 10, 30, 0, 0, time.UTC)

	result, err := CreateTransitionChain(env).
		WithTimeSource(func() time.Time { return fixedTime }).
		StartEpicAt(customTime).
		StartPhaseAt("1A", customTime.Add(time.Hour)).
		Execute()

	if err != nil {
		t.Fatalf("Failed to execute transition chain: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got errors: %v", result.Errors)
	}

	// Verify timestamp control
	cmd1 := result.ExecutedCommands[0]
	if cmd1.Command.Timestamp == nil {
		t.Error("Expected custom timestamp for first command")
	} else if !cmd1.Command.Timestamp.Equal(customTime) {
		t.Errorf("Expected first command timestamp %v, got: %v", customTime, *cmd1.Command.Timestamp)
	}

	cmd2 := result.ExecutedCommands[1]
	if cmd2.Command.Timestamp == nil {
		t.Error("Expected custom timestamp for second command")
	} else if !cmd2.Command.Timestamp.Equal(customTime.Add(time.Hour)) {
		t.Errorf("Expected second command timestamp %v, got: %v", customTime.Add(time.Hour), *cmd2.Command.Timestamp)
	}
}

func TestTransitionChain_IntermediateAssertions(t *testing.T) {
	env := NewTestExecutionEnvironment("test-epic.xml")

	// Build test epic
	testEpic, err := builders.NewEpicBuilder("test-epic").
		WithStatus("planning").
		WithPhase("1A", "Setup", "pending").
		Build()

	if err != nil {
		t.Fatalf("Failed to build test epic: %v", err)
	}

	err = env.LoadEpic(testEpic)
	if err != nil {
		t.Fatalf("Failed to load epic: %v", err)
	}

	// Execute with intermediate assertions
	result, err := CreateTransitionChain(env).
		StartEpic().
		Assert().EpicStatus("active").
		StartPhase("1A").
		Assert().PhaseStatus("1A", "active").
		Execute()

	if err != nil {
		t.Fatalf("Failed to execute transition chain: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution with assertions, got errors: %v", result.Errors)
	}

	// All commands should succeed including assertions
	for i, cmd := range result.ExecutedCommands {
		if !cmd.Success {
			t.Errorf("Expected command %d to succeed, got error: %v", i, cmd.Error)
		}
	}
}

func TestTransitionChain_FailedAssertions(t *testing.T) {
	env := NewTestExecutionEnvironment("test-epic.xml")

	// Build test epic
	testEpic, err := builders.NewEpicBuilder("test-epic").
		WithStatus("planning").
		WithPhase("1A", "Setup", "pending").
		Build()

	if err != nil {
		t.Fatalf("Failed to build test epic: %v", err)
	}

	err = env.LoadEpic(testEpic)
	if err != nil {
		t.Fatalf("Failed to load epic: %v", err)
	}

	// Execute with incorrect assertion
	result, err := CreateTransitionChain(env).
		StartEpic().
		Assert().EpicStatus("completed"). // This should fail - epic is active, not completed
		Execute()

	if err != nil {
		t.Fatalf("Failed to execute transition chain: %v", err)
	}

	// Should have assertion error
	if result.Success {
		t.Error("Expected failed execution due to assertion failure")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected assertion error in results")
	}

	// Check for assertion error
	foundAssertionError := false
	for _, err := range result.Errors {
		if err.Command == "assertion_after_start_epic" {
			foundAssertionError = true
			break
		}
	}

	if !foundAssertionError {
		t.Error("Expected to find assertion error in results")
	}
}

// Helper functions for tests
func findPhaseByID(e *epic.Epic, phaseID string) *epic.Phase {
	for i := range e.Phases {
		if e.Phases[i].ID == phaseID {
			return &e.Phases[i]
		}
	}
	return nil
}

func findTaskByID(e *epic.Epic, taskID string) *epic.Task {
	for i := range e.Tasks {
		if e.Tasks[i].ID == taskID {
			return &e.Tasks[i]
		}
	}
	return nil
}

func findTestByID(e *epic.Epic, testID string) *epic.Test {
	for i := range e.Tests {
		if e.Tests[i].ID == testID {
			return &e.Tests[i]
		}
	}
	return nil
}
