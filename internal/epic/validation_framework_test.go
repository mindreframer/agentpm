package epic

import (
	"testing"
	"time"
)

// createTestEpic creates a test epic with predefined structure
func createTestEpic() *Epic {
	now := time.Now()
	epic := &Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: StatusPending,
		CurrentState: &CurrentState{
			ActivePhase: "phase-1",
		},
		Phases: []Phase{
			{
				ID:     "phase-1",
				Name:   "Implementation",
				Status: StatusPending,
			},
			{
				ID:     "phase-2",
				Name:   "Testing",
				Status: StatusPending,
			},
		},
		Tasks: []Task{
			{
				ID:      "task-1",
				PhaseID: "phase-1",
				Name:    "Implement feature",
				Status:  StatusPending,
			},
			{
				ID:      "task-2",
				PhaseID: "phase-1",
				Name:    "Write docs",
				Status:  StatusPending,
			},
		},
		Tests: []Test{
			{
				ID:         "test-1",
				TaskID:     "task-1",
				PhaseID:    "phase-1",
				Name:       "Unit test",
				TestStatus: TestStatusPending,
				TestResult: TestResultFailing,
			},
			{
				ID:         "test-2",
				TaskID:     "task-1",
				PhaseID:    "phase-1",
				Name:       "Integration test",
				TestStatus: TestStatusPending,
				TestResult: TestResultFailing,
			},
		},
		CreatedAt: now,
	}
	return epic
}

// TestStatusValidationError tests the StatusValidationError struct
func TestStatusValidationError(t *testing.T) {
	t.Run("error interface implementation", func(t *testing.T) {
		err := &StatusValidationError{
			Message: "Test error message",
		}

		if err.Error() != "Test error message" {
			t.Errorf("Expected Error() to return 'Test error message', got %s", err.Error())
		}
	})

	t.Run("contains all required fields", func(t *testing.T) {
		err := &StatusValidationError{
			EntityType:    "task",
			EntityID:      "task-1",
			EntityName:    "Test Task",
			CurrentStatus: "pending",
			TargetStatus:  "done",
			BlockingItems: []BlockingItem{{Type: "test", ID: "test-1", Name: "Test", Status: "pending"}},
			Message:       "Validation failed",
			Suggestions:   []string{"Complete tests first"},
		}

		if err.EntityType != "task" {
			t.Errorf("Expected EntityType to be 'task', got %s", err.EntityType)
		}
		if err.EntityID != "task-1" {
			t.Errorf("Expected EntityID to be 'task-1', got %s", err.EntityID)
		}
		if len(err.BlockingItems) != 1 {
			t.Errorf("Expected 1 blocking item, got %d", len(err.BlockingItems))
		}
		if len(err.Suggestions) != 1 {
			t.Errorf("Expected 1 suggestion, got %d", len(err.Suggestions))
		}
	})
}

// TestBlockingItem tests the BlockingItem struct
func TestBlockingItem(t *testing.T) {
	t.Run("holds proper validation details", func(t *testing.T) {
		item := BlockingItem{
			Type:   "test",
			ID:     "test-1",
			Name:   "Unit Test",
			Status: "pending",
			Result: "failing",
		}

		if item.Type != "test" {
			t.Errorf("Expected Type to be 'test', got %s", item.Type)
		}
		if item.ID != "test-1" {
			t.Errorf("Expected ID to be 'test-1', got %s", item.ID)
		}
		if item.Result != "failing" {
			t.Errorf("Expected Result to be 'failing', got %s", item.Result)
		}
	})
}

// TestNewStatusValidator tests the StatusValidator creation
func TestNewStatusValidator(t *testing.T) {
	epic := createTestEpic()
	validator := NewStatusValidator(epic)

	if validator == nil {
		t.Fatal("Expected validator to be created")
	}
	if validator.epic != epic {
		t.Error("Expected validator to hold reference to epic")
	}
}

// TestValidateEpicStatusTransition tests epic status transition validation
func TestValidateEpicStatusTransition(t *testing.T) {
	epic := createTestEpic()
	validator := NewStatusValidator(epic)

	t.Run("valid epic status transition", func(t *testing.T) {
		epic.Status = StatusPending
		result := validator.ValidateEpicStatusTransition(EpicStatusWIP)

		if !result.Valid {
			t.Errorf("Expected transition from pending to wip to be valid")
		}
		if result.Error != nil {
			t.Errorf("Expected no error for valid transition, got: %s", result.Error.Message)
		}
	})

	t.Run("invalid epic status transition", func(t *testing.T) {
		epic.Status = StatusPending
		result := validator.ValidateEpicStatusTransition(EpicStatusDone)

		if result.Valid {
			t.Error("Expected transition from pending to done to be invalid")
		}
		if result.Error == nil {
			t.Error("Expected error for invalid transition")
		}
		if result.Error.EntityType != "epic" {
			t.Errorf("Expected EntityType to be 'epic', got %s", result.Error.EntityType)
		}
	})

	t.Run("epic completion validation with incomplete phases", func(t *testing.T) {
		epic.Status = Status("wip")               // Maps to EpicStatusWIP, can transition to done
		epic.Phases[0].Status = Status("pending") // Phase is not done
		epic.Phases[1].Status = Status("pending") // Phase is not done
		result := validator.ValidateEpicStatusTransition(EpicStatusDone)

		if result.Valid {
			t.Error("Expected epic completion to be blocked by incomplete phases")
		}
		if result.Error == nil {
			t.Error("Expected error for blocked epic completion")
		}
		if result.BlockingCount == 0 {
			t.Error("Expected blocking count to be greater than 0")
		}
	})
}

// TestValidatePhaseStatusTransition tests phase status transition validation
func TestValidatePhaseStatusTransition(t *testing.T) {
	epic := createTestEpic()
	validator := NewStatusValidator(epic)

	t.Run("valid phase status transition", func(t *testing.T) {
		epic.Phases[0].Status = StatusPending
		result := validator.ValidatePhaseStatusTransition("phase-1", PhaseStatusWIP)

		if !result.Valid {
			t.Errorf("Expected transition from pending to wip to be valid")
		}
		if result.Error != nil {
			t.Errorf("Expected no error for valid transition, got: %s", result.Error.Message)
		}
	})

	t.Run("invalid phase status transition", func(t *testing.T) {
		epic.Phases[0].Status = StatusPending
		result := validator.ValidatePhaseStatusTransition("phase-1", PhaseStatusDone)

		if result.Valid {
			t.Error("Expected transition from pending to done to be invalid")
		}
		if result.Error == nil {
			t.Error("Expected error for invalid transition")
		}
		if result.Error.EntityType != "phase" {
			t.Errorf("Expected EntityType to be 'phase', got %s", result.Error.EntityType)
		}
	})

	t.Run("phase not found", func(t *testing.T) {
		result := validator.ValidatePhaseStatusTransition("nonexistent", PhaseStatusWIP)

		if result.Valid {
			t.Error("Expected validation to fail for nonexistent phase")
		}
		if result.Error == nil {
			t.Error("Expected error for nonexistent phase")
		}
	})

	t.Run("phase completion validation with incomplete tasks", func(t *testing.T) {
		epic.Phases[0].Status = Status("wip")    // Maps to PhaseStatusWIP, can transition to done
		epic.Tasks[0].Status = Status("pending") // Task is not done
		epic.Tasks[1].Status = Status("wip")     // Task is not done
		result := validator.ValidatePhaseStatusTransition("phase-1", PhaseStatusDone)

		if result.Valid {
			t.Error("Expected phase completion to be blocked by incomplete tasks")
		}
		if result.Error == nil {
			t.Error("Expected error for blocked phase completion")
		}
		if result.BlockingCount == 0 {
			t.Error("Expected blocking count to be greater than 0")
		}
	})
}

// TestValidateTaskStatusTransition tests task status transition validation
func TestValidateTaskStatusTransition(t *testing.T) {
	epic := createTestEpic()
	validator := NewStatusValidator(epic)

	t.Run("valid task status transition", func(t *testing.T) {
		epic.Tasks[0].Status = StatusPending
		result := validator.ValidateTaskStatusTransition("task-1", TaskStatusWIP)

		if !result.Valid {
			t.Errorf("Expected transition from pending to wip to be valid")
		}
		if result.Error != nil {
			t.Errorf("Expected no error for valid transition, got: %s", result.Error.Message)
		}
	})

	t.Run("invalid task status transition", func(t *testing.T) {
		epic.Tasks[0].Status = StatusPending
		result := validator.ValidateTaskStatusTransition("task-1", TaskStatusDone)

		if result.Valid {
			t.Error("Expected transition from pending to done to be invalid")
		}
		if result.Error == nil {
			t.Error("Expected error for invalid transition")
		}
		if result.Error.EntityType != "task" {
			t.Errorf("Expected EntityType to be 'task', got %s", result.Error.EntityType)
		}
	})

	t.Run("task not found", func(t *testing.T) {
		result := validator.ValidateTaskStatusTransition("nonexistent", TaskStatusWIP)

		if result.Valid {
			t.Error("Expected validation to fail for nonexistent task")
		}
		if result.Error == nil {
			t.Error("Expected error for nonexistent task")
		}
	})

	t.Run("task completion validation with incomplete tests", func(t *testing.T) {
		epic.Tasks[0].Status = Status("wip")         // Maps to TaskStatusWIP, can transition to done
		epic.Tests[0].TestStatus = TestStatusPending // Test is not done
		epic.Tests[1].TestStatus = TestStatusWIP     // Test is not done
		result := validator.ValidateTaskStatusTransition("task-1", TaskStatusDone)

		if result.Valid {
			t.Error("Expected task completion to be blocked by incomplete tests")
		}
		if result.Error == nil {
			t.Error("Expected error for blocked task completion")
		}
		if result.BlockingCount == 0 {
			t.Error("Expected blocking count to be greater than 0")
		}
	})
}

// TestValidateTestStatusTransition tests test status transition validation
func TestValidateTestStatusTransition(t *testing.T) {
	epic := createTestEpic()
	validator := NewStatusValidator(epic)

	t.Run("valid test status transition", func(t *testing.T) {
		epic.Tests[0].TestStatus = TestStatusPending
		result := validator.ValidateTestStatusTransition("test-1", TestStatusWIP, TestResultFailing)

		if !result.Valid {
			t.Errorf("Expected transition from pending to wip to be valid")
		}
		if result.Error != nil {
			t.Errorf("Expected no error for valid transition, got: %s", result.Error.Message)
		}
	})

	t.Run("invalid test status transition", func(t *testing.T) {
		epic.Tests[0].TestStatus = TestStatusPending
		result := validator.ValidateTestStatusTransition("test-1", TestStatusDone, TestResultPassing)

		if result.Valid {
			t.Error("Expected transition from pending to done to be invalid")
		}
		if result.Error == nil {
			t.Error("Expected error for invalid transition")
		}
		if result.Error.EntityType != "test" {
			t.Errorf("Expected EntityType to be 'test', got %s", result.Error.EntityType)
		}
	})

	t.Run("failing tests cannot be marked done", func(t *testing.T) {
		epic.Tests[0].TestStatus = TestStatusWIP
		result := validator.ValidateTestStatusTransition("test-1", TestStatusDone, TestResultFailing)

		if result.Valid {
			t.Error("Expected failing test to not be markable as done")
		}
		if result.Error == nil {
			t.Error("Expected error for failing test marked as done")
		}
	})

	t.Run("test not found", func(t *testing.T) {
		result := validator.ValidateTestStatusTransition("nonexistent", TestStatusWIP, TestResultFailing)

		if result.Valid {
			t.Error("Expected validation to fail for nonexistent test")
		}
		if result.Error == nil {
			t.Error("Expected error for nonexistent test")
		}
	})

	t.Run("test not in active phase", func(t *testing.T) {
		epic.CurrentState.ActivePhase = "phase-2"
		epic.Tests[0].TestStatus = TestStatusPending
		result := validator.ValidateTestStatusTransition("test-1", TestStatusWIP, TestResultFailing)

		if result.Valid {
			t.Error("Expected validation to fail for test not in active phase")
		}
		if result.Error == nil {
			t.Error("Expected error for test not in active phase")
		}
	})
}

// TestStatusValidationBusinessRules tests business rule enforcement
func TestStatusValidationBusinessRules(t *testing.T) {
	epic := createTestEpic()
	validator := NewStatusValidator(epic)

	t.Run("phase completion with pending tasks and tests", func(t *testing.T) {
		// Set phase to "wip" status for Epic 13 system
		epic.Phases[0].Status = Status("wip")    // Maps to PhaseStatusWIP
		epic.Tasks[0].Status = Status("pending") // Maps to TaskStatusPending
		epic.Tasks[1].Status = Status("wip")     // Maps to TaskStatusWIP
		epic.Tests[0].TestStatus = TestStatusPending
		epic.Tests[1].TestStatus = TestStatusWIP

		result := validator.ValidatePhaseStatusTransition("phase-1", PhaseStatusDone)

		if result.Valid {
			t.Error("Expected phase completion to be blocked")
		}

		// Should have 2 tasks + 2 tests = 4 blocking items
		if result.BlockingCount != 4 {
			t.Errorf("Expected 4 blocking items, got %d", result.BlockingCount)
		}

		// Check error message includes counts
		expectedMessage := "Phase cannot be completed: 2 pending/wip tasks, 2 pending/wip tests"
		if result.Error.Message != expectedMessage {
			t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Error.Message)
		}
	})

	t.Run("task completion with pending tests", func(t *testing.T) {
		epic.Tasks[0].Status = Status("wip") // Maps to TaskStatusWIP
		epic.Tests[0].TestStatus = TestStatusPending
		epic.Tests[1].TestStatus = TestStatusWIP

		result := validator.ValidateTaskStatusTransition("task-1", TaskStatusDone)

		if result.Valid {
			t.Error("Expected task completion to be blocked")
		}

		// Should have 2 tests blocking
		if result.BlockingCount != 2 {
			t.Errorf("Expected 2 blocking items, got %d", result.BlockingCount)
		}

		// Check error message includes count
		expectedMessage := "Task cannot be completed: 2 pending/wip tests"
		if result.Error.Message != expectedMessage {
			t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Error.Message)
		}
	})
}

// TestStatusValidationHelperMethods tests the helper methods
func TestStatusValidationHelperMethods(t *testing.T) {
	epic := createTestEpic()
	validator := NewStatusValidator(epic)

	t.Run("findPhase", func(t *testing.T) {
		phase := validator.findPhase("phase-1")
		if phase == nil {
			t.Error("Expected to find phase-1")
		}
		if phase.ID != "phase-1" {
			t.Errorf("Expected phase ID to be 'phase-1', got %s", phase.ID)
		}

		phase = validator.findPhase("nonexistent")
		if phase != nil {
			t.Error("Expected to not find nonexistent phase")
		}
	})

	t.Run("findTask", func(t *testing.T) {
		task := validator.findTask("task-1")
		if task == nil {
			t.Error("Expected to find task-1")
		}
		if task.ID != "task-1" {
			t.Errorf("Expected task ID to be 'task-1', got %s", task.ID)
		}

		task = validator.findTask("nonexistent")
		if task != nil {
			t.Error("Expected to not find nonexistent task")
		}
	})

	t.Run("findTest", func(t *testing.T) {
		test := validator.findTest("test-1")
		if test == nil {
			t.Error("Expected to find test-1")
		}
		if test.ID != "test-1" {
			t.Errorf("Expected test ID to be 'test-1', got %s", test.ID)
		}

		test = validator.findTest("nonexistent")
		if test != nil {
			t.Error("Expected to not find nonexistent test")
		}
	})

	t.Run("isTestInActivePhase", func(t *testing.T) {
		epic.CurrentState.ActivePhase = "phase-1"

		if !validator.isTestInActivePhase("test-1") {
			t.Error("Expected test-1 to be in active phase")
		}

		epic.CurrentState.ActivePhase = "phase-2"
		if validator.isTestInActivePhase("test-1") {
			t.Error("Expected test-1 to not be in active phase")
		}
	})

	t.Run("getActivePhaseID", func(t *testing.T) {
		epic.CurrentState.ActivePhase = "phase-1"
		activePhase := validator.getActivePhaseID()
		if activePhase != "phase-1" {
			t.Errorf("Expected active phase to be 'phase-1', got %s", activePhase)
		}

		// Test fallback when no current state
		epic.CurrentState = nil
		epic.Phases[0].Status = StatusPending
		activePhase = validator.getActivePhaseID()
		if activePhase != "phase-1" {
			t.Errorf("Expected fallback active phase to be 'phase-1', got %s", activePhase)
		}
	})
}

// TestFormatErrorXML tests the XML error formatting
func TestFormatErrorXML(t *testing.T) {
	t.Run("format error with blocking items", func(t *testing.T) {
		err := &StatusValidationError{
			EntityType: "phase",
			Message:    "Phase cannot be completed: 1 pending task, 1 wip test",
			BlockingItems: []BlockingItem{
				{Type: "task", ID: "task-1", Name: "Test Task", Status: "pending"},
				{Type: "test", ID: "test-1", Name: "Test", Status: "wip", Result: "failing"},
			},
		}

		xml := err.FormatErrorXML()

		// Check key elements are present
		if !contains(xml, "<error>") {
			t.Error("Expected XML to contain <error> tag")
		}
		if !contains(xml, "<type>phase_completion_blocked</type>") {
			t.Error("Expected XML to contain error type")
		}
		if !contains(xml, "<message>Phase cannot be completed: 1 pending task, 1 wip test</message>") {
			t.Error("Expected XML to contain error message")
		}
		if !contains(xml, "<blocking_items>") {
			t.Error("Expected XML to contain blocking items")
		}
		if !contains(xml, "<tasks count=\"1\">") {
			t.Error("Expected XML to contain tasks section")
		}
		if !contains(xml, "<tests count=\"1\">") {
			t.Error("Expected XML to contain tests section")
		}
	})

	t.Run("format error without blocking items", func(t *testing.T) {
		err := &StatusValidationError{
			EntityType: "test",
			Message:    "Test cannot transition from pending to done",
		}

		xml := err.FormatErrorXML()

		if !contains(xml, "<type>test_completion_blocked</type>") {
			t.Error("Expected XML to contain error type")
		}
		if contains(xml, "<blocking_items>") {
			t.Error("Expected XML to not contain blocking items section")
		}
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
