package builders

import (
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
)

func TestEpicBuilder_BasicConstruction(t *testing.T) {
	builder := NewEpicBuilder("test-epic")

	result, err := builder.Build()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.ID != "test-epic" {
		t.Errorf("Expected ID to be 'test-epic', got: %s", result.ID)
	}

	if result.Name != "test-epic" {
		t.Errorf("Expected Name to be 'test-epic', got: %s", result.Name)
	}

	if result.Status != epic.StatusPlanning {
		t.Errorf("Expected Status to be 'planning', got: %s", result.Status)
	}

	// Check that default metadata is created
	if result.Metadata == nil {
		t.Error("Expected Metadata to be created by default")
	}

	if result.CurrentState == nil {
		t.Error("Expected CurrentState to be created by default")
	}
}

func TestEpicBuilder_FluentAPI(t *testing.T) {
	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	result, err := NewEpicBuilder("test-epic").
		WithName("Test Epic").
		WithStatus("pending").
		WithCreatedAt(fixedTime).
		WithAssignee("agent_claude").
		WithDescription("A test epic").
		WithWorkflow("test workflow").
		WithRequirements("test requirements").
		WithDependencies("test dependencies").
		Build()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Name != "Test Epic" {
		t.Errorf("Expected Name to be 'Test Epic', got: %s", result.Name)
	}

	if result.Status != epic.StatusPending {
		t.Errorf("Expected Status to be 'pending', got: %s", result.Status)
	}

	if !result.CreatedAt.Equal(fixedTime) {
		t.Errorf("Expected CreatedAt to be %v, got: %v", fixedTime, result.CreatedAt)
	}

	if result.Assignee != "agent_claude" {
		t.Errorf("Expected Assignee to be 'agent_claude', got: %s", result.Assignee)
	}

	if result.Description != "A test epic" {
		t.Errorf("Expected Description to be 'A test epic', got: %s", result.Description)
	}

	if result.Workflow != "test workflow" {
		t.Errorf("Expected Workflow to be 'test workflow', got: %s", result.Workflow)
	}

	if result.Requirements != "test requirements" {
		t.Errorf("Expected Requirements to be 'test requirements', got: %s", result.Requirements)
	}

	if result.Dependencies != "test dependencies" {
		t.Errorf("Expected Dependencies to be 'test dependencies', got: %s", result.Dependencies)
	}
}

func TestEpicBuilder_WithPhases(t *testing.T) {
	result, err := NewEpicBuilder("test-epic").
		WithPhase("1A", "Setup", "pending").
		WithPhase("1B", "Development", "pending").
		WithPhaseDescriptive("1C", "Testing", "Test all features", "Test report", "pending").
		Build()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Phases) != 3 {
		t.Fatalf("Expected 3 phases, got: %d", len(result.Phases))
	}

	// Check first phase
	phase1 := result.Phases[0]
	if phase1.ID != "1A" || phase1.Name != "Setup" || phase1.Status != epic.StatusPending {
		t.Errorf("First phase incorrect: ID=%s, Name=%s, Status=%s", phase1.ID, phase1.Name, phase1.Status)
	}

	// Check descriptive phase
	phase3 := result.Phases[2]
	if phase3.ID != "1C" || phase3.Description != "Test all features" || phase3.Deliverables != "Test report" {
		t.Errorf("Third phase incorrect: ID=%s, Description=%s, Deliverables=%s",
			phase3.ID, phase3.Description, phase3.Deliverables)
	}
}

func TestEpicBuilder_WithTasks(t *testing.T) {
	result, err := NewEpicBuilder("test-epic").
		WithPhase("1A", "Setup", "pending").
		WithTask("1A_1", "1A", "Initialize Project", "pending").
		WithTask("1A_2", "1A", "Configure Tools", "pending").
		WithTaskDescriptive("1A_3", "1A", "Setup Database", "Set up test database", "Database is ready", "pending", "agent_claude").
		Build()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Tasks) != 3 {
		t.Fatalf("Expected 3 tasks, got: %d", len(result.Tasks))
	}

	// Check basic task
	task1 := result.Tasks[0]
	if task1.ID != "1A_1" || task1.PhaseID != "1A" || task1.Name != "Initialize Project" {
		t.Errorf("First task incorrect: ID=%s, PhaseID=%s, Name=%s", task1.ID, task1.PhaseID, task1.Name)
	}

	// Check descriptive task
	task3 := result.Tasks[2]
	if task3.Description != "Set up test database" || task3.AcceptanceCriteria != "Database is ready" || task3.Assignee != "agent_claude" {
		t.Errorf("Third task incorrect: Description=%s, AcceptanceCriteria=%s, Assignee=%s",
			task3.Description, task3.AcceptanceCriteria, task3.Assignee)
	}
}

func TestEpicBuilder_WithTests(t *testing.T) {
	result, err := NewEpicBuilder("test-epic").
		WithPhase("1A", "Setup", "pending").
		WithTask("1A_1", "1A", "Initialize Project", "pending").
		WithTest("T1A_1", "1A_1", "1A", "Test Project Init", "pending").
		WithTestDescriptive("T1A_2", "1A_1", "1A", "Test Config", "Test configuration works", "pending", "pending", "passing").
		Build()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Tests) != 2 {
		t.Fatalf("Expected 2 tests, got: %d", len(result.Tests))
	}

	// Check basic test
	test1 := result.Tests[0]
	if test1.ID != "T1A_1" || test1.TaskID != "1A_1" || test1.PhaseID != "1A" || test1.Name != "Test Project Init" {
		t.Errorf("First test incorrect: ID=%s, TaskID=%s, PhaseID=%s, Name=%s",
			test1.ID, test1.TaskID, test1.PhaseID, test1.Name)
	}

	// Check descriptive test with Epic 13 status
	test2 := result.Tests[1]
	if test2.Description != "Test configuration works" || test2.TestStatus != epic.TestStatusPending || test2.TestResult != epic.TestResultPassing {
		t.Errorf("Second test incorrect: Description=%s, TestStatus=%s, TestResult=%s",
			test2.Description, test2.TestStatus, test2.TestResult)
	}
}

func TestEpicBuilder_WithEvents(t *testing.T) {
	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	result, err := NewEpicBuilder("test-epic").
		WithEvent("evt1", "epic_started", "Epic started by user", fixedTime).
		WithEvent("evt2", "phase_started", "Phase 1A started", fixedTime.Add(time.Hour)).
		Build()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Events) != 2 {
		t.Fatalf("Expected 2 events, got: %d", len(result.Events))
	}

	event1 := result.Events[0]
	if event1.ID != "evt1" || event1.Type != "epic_started" || event1.Data != "Epic started by user" {
		t.Errorf("First event incorrect: ID=%s, Type=%s, Data=%s", event1.ID, event1.Type, event1.Data)
	}

	if !event1.Timestamp.Equal(fixedTime) {
		t.Errorf("First event timestamp incorrect: expected %v, got %v", fixedTime, event1.Timestamp)
	}
}

func TestEpicBuilder_RelationshipValidation(t *testing.T) {
	t.Run("TaskReferencesNonExistentPhase", func(t *testing.T) {
		_, err := NewEpicBuilder("test-epic").
			WithPhase("1A", "Setup", "pending").
			WithTask("1B_1", "1B", "Invalid Task", "pending"). // References non-existent phase 1B
			Build()

		if err == nil {
			t.Error("Expected error for task referencing non-existent phase")
		}

		expectedMsg := "task 1B_1 references non-existent phase: 1B"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
		}
	})

	t.Run("TestReferencesNonExistentTask", func(t *testing.T) {
		_, err := NewEpicBuilder("test-epic").
			WithPhase("1A", "Setup", "pending").
			WithTask("1A_1", "1A", "Valid Task", "pending").
			WithTest("T1A_2", "1A_2", "1A", "Invalid Test", "pending"). // References non-existent task 1A_2
			Build()

		if err == nil {
			t.Error("Expected error for test referencing non-existent task")
		}

		expectedMsg := "test T1A_2 references non-existent task: 1A_2"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
		}
	})

	t.Run("TestReferencesNonExistentPhase", func(t *testing.T) {
		_, err := NewEpicBuilder("test-epic").
			WithPhase("1A", "Setup", "pending").
			WithTask("1A_1", "1A", "Valid Task", "pending").
			WithTest("T1B_1", "1A_1", "1B", "Invalid Test", "pending"). // References non-existent phase 1B
			Build()

		if err == nil {
			t.Error("Expected error for test referencing non-existent phase")
		}

		expectedMsg := "test T1B_1 references non-existent phase: 1B"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
		}
	})
}

func TestEpicBuilder_DuplicateIDValidation(t *testing.T) {
	t.Run("DuplicatePhaseIDs", func(t *testing.T) {
		_, err := NewEpicBuilder("test-epic").
			WithPhase("1A", "Setup", "pending").
			WithPhase("1A", "Duplicate Setup", "pending"). // Duplicate phase ID
			Build()

		if err == nil {
			t.Error("Expected error for duplicate phase IDs")
		}

		expectedMsg := "duplicate phase ID: 1A"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
		}
	})

	t.Run("DuplicateTaskIDs", func(t *testing.T) {
		_, err := NewEpicBuilder("test-epic").
			WithPhase("1A", "Setup", "pending").
			WithTask("1A_1", "1A", "Task 1", "pending").
			WithTask("1A_1", "1A", "Duplicate Task", "pending"). // Duplicate task ID
			Build()

		if err == nil {
			t.Error("Expected error for duplicate task IDs")
		}

		expectedMsg := "duplicate task ID: 1A_1"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
		}
	})

	t.Run("DuplicateTestIDs", func(t *testing.T) {
		_, err := NewEpicBuilder("test-epic").
			WithPhase("1A", "Setup", "pending").
			WithTask("1A_1", "1A", "Task 1", "pending").
			WithTest("T1A_1", "1A_1", "1A", "Test 1", "pending").
			WithTest("T1A_1", "1A_1", "1A", "Duplicate Test", "pending"). // Duplicate test ID
			Build()

		if err == nil {
			t.Error("Expected error for duplicate test IDs")
		}

		expectedMsg := "duplicate test ID: T1A_1"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
		}
	})
}

func TestEpicBuilder_StatusValidation(t *testing.T) {
	t.Run("InvalidEpicStatus", func(t *testing.T) {
		_, err := NewEpicBuilder("test-epic").
			WithStatus("invalid_status").
			Build()

		if err == nil {
			t.Error("Expected error for invalid epic status")
		}

		expectedMsg := "invalid epic status: invalid_status"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
		}
	})

	t.Run("InvalidPhaseStatus", func(t *testing.T) {
		_, err := NewEpicBuilder("test-epic").
			WithPhase("1A", "Setup", "invalid_status").
			Build()

		if err == nil {
			t.Error("Expected error for invalid phase status")
		}

		expectedMsg := "invalid phase status for phase 1A: invalid_status"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
		}
	})

	t.Run("InvalidTaskStatus", func(t *testing.T) {
		_, err := NewEpicBuilder("test-epic").
			WithPhase("1A", "Setup", "pending").
			WithTask("1A_1", "1A", "Task 1", "invalid_status").
			Build()

		if err == nil {
			t.Error("Expected error for invalid task status")
		}

		expectedMsg := "invalid task status for task 1A_1: invalid_status"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
		}
	})

	t.Run("InvalidTestStatus", func(t *testing.T) {
		_, err := NewEpicBuilder("test-epic").
			WithPhase("1A", "Setup", "pending").
			WithTask("1A_1", "1A", "Task 1", "pending").
			WithTest("T1A_1", "1A_1", "1A", "Test 1", "invalid_status").
			Build()

		if err == nil {
			t.Error("Expected error for invalid test status")
		}

		expectedMsg := "invalid test status for test T1A_1: invalid_status"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
		}
	})

	t.Run("InvalidTestStatusUnified", func(t *testing.T) {
		_, err := NewEpicBuilder("test-epic").
			WithPhase("1A", "Setup", "pending").
			WithTask("1A_1", "1A", "Task 1", "pending").
			WithTestDescriptive("T1A_1", "1A_1", "1A", "Test 1", "desc", "pending", "invalid_test_status", "passing").
			Build()

		if err == nil {
			t.Error("Expected error for invalid test status (unified)")
		}

		expectedMsg := "invalid test status for test T1A_1: invalid_test_status"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
		}
	})

	t.Run("InvalidTestResult", func(t *testing.T) {
		_, err := NewEpicBuilder("test-epic").
			WithPhase("1A", "Setup", "pending").
			WithTask("1A_1", "1A", "Task 1", "pending").
			WithTestDescriptive("T1A_1", "1A_1", "1A", "Test 1", "desc", "pending", "pending", "invalid_result").
			Build()

		if err == nil {
			t.Error("Expected error for invalid test result")
		}

		expectedMsg := "invalid test result for test T1A_1: invalid_result"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
		}
	})
}

func TestEpicBuilder_RequiredFieldValidation(t *testing.T) {
	t.Run("EmptyEpicID", func(t *testing.T) {
		_, err := NewEpicBuilder("").Build()

		if err == nil {
			t.Error("Expected error for empty epic ID")
		}

		expectedMsg := "epic ID is required"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
		}
	})

	t.Run("EmptyPhaseID", func(t *testing.T) {
		_, err := NewEpicBuilder("test-epic").
			WithPhase("", "Setup", "pending").
			Build()

		if err == nil {
			t.Error("Expected error for empty phase ID")
		}

		expectedMsg := "phase ID is required"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
		}
	})

	t.Run("EmptyTaskID", func(t *testing.T) {
		_, err := NewEpicBuilder("test-epic").
			WithPhase("1A", "Setup", "pending").
			WithTask("", "1A", "Task 1", "pending").
			Build()

		if err == nil {
			t.Error("Expected error for empty task ID")
		}

		expectedMsg := "task ID is required"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
		}
	})

	t.Run("EmptyTestID", func(t *testing.T) {
		_, err := NewEpicBuilder("test-epic").
			WithPhase("1A", "Setup", "pending").
			WithTask("1A_1", "1A", "Task 1", "pending").
			WithTest("", "1A_1", "1A", "Test 1", "pending").
			Build()

		if err == nil {
			t.Error("Expected error for empty test ID")
		}

		expectedMsg := "test ID is required"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got: %s", expectedMsg, err.Error())
		}
	})
}

func TestEpicBuilder_DisableDefaultValues(t *testing.T) {
	result, err := NewEpicBuilder("test-epic").
		DisableDefaultValues().
		Build()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Metadata != nil {
		t.Error("Expected Metadata to be nil when default values are disabled")
	}

	if result.CurrentState != nil {
		t.Error("Expected CurrentState to be nil when default values are disabled")
	}
}

func TestEpicBuilder_ComplexScenario(t *testing.T) {
	// Test a complete realistic scenario matching the specification examples
	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	result, err := CreateEpicBuilder("test-epic").
		WithName("Test Epic").
		WithStatus("pending").
		WithAssignee("agent_claude").
		WithCreatedAt(fixedTime).
		WithPhase("1A", "Setup", "pending").
		WithPhase("1B", "Development", "pending").
		WithTask("1A_1", "1A", "Initialize Project", "pending").
		WithTask("1A_2", "1A", "Configure Tools", "pending").
		WithTask("1B_1", "1B", "Implement Feature", "pending").
		WithTest("T1A_1", "1A_1", "1A", "Test Project Init", "pending").
		WithTestDescriptive("T1A_2", "1A_2", "1A", "Test Tool Config", "Test tools work", "pending", "pending", "passing").
		WithEvent("evt1", "epic_created", "Epic created", fixedTime).
		Build()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Validate the complete structure
	if len(result.Phases) != 2 {
		t.Errorf("Expected 2 phases, got: %d", len(result.Phases))
	}

	if len(result.Tasks) != 3 {
		t.Errorf("Expected 3 tasks, got: %d", len(result.Tasks))
	}

	if len(result.Tests) != 2 {
		t.Errorf("Expected 2 tests, got: %d", len(result.Tests))
	}

	if len(result.Events) != 1 {
		t.Errorf("Expected 1 event, got: %d", len(result.Events))
	}

	// Validate relationships
	// Task 1A_1 should reference phase 1A
	task1A1 := findTaskByID(result.Tasks, "1A_1")
	if task1A1 == nil || task1A1.PhaseID != "1A" {
		t.Error("Task 1A_1 should reference phase 1A")
	}

	// Test T1A_1 should reference task 1A_1 and phase 1A
	testT1A1 := findTestByID(result.Tests, "T1A_1")
	if testT1A1 == nil || testT1A1.TaskID != "1A_1" || testT1A1.PhaseID != "1A" {
		t.Error("Test T1A_1 should reference task 1A_1 and phase 1A")
	}

	// Test unified status system
	testT1A2 := findTestByID(result.Tests, "T1A_2")
	if testT1A2 == nil || testT1A2.TestStatus != epic.TestStatusPending || testT1A2.TestResult != epic.TestResultPassing {
		t.Error("Test T1A_2 should have unified status fields set correctly")
	}
}

// Helper functions for tests
func findTaskByID(tasks []epic.Task, id string) *epic.Task {
	for i := range tasks {
		if tasks[i].ID == id {
			return &tasks[i]
		}
	}
	return nil
}

func findTestByID(tests []epic.Test, id string) *epic.Test {
	for i := range tests {
		if tests[i].ID == id {
			return &tests[i]
		}
	}
	return nil
}
