package epic

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEpic_Validate(t *testing.T) {
	t.Run("valid epic passes validation", func(t *testing.T) {
		epic := &Epic{
			ID:        "test-1",
			Name:      "Test Epic",
			Status:    StatusPlanning,
			CreatedAt: time.Now(),
			Phases: []Phase{
				{ID: "P1", Name: "Phase 1", Status: StatusPlanning},
			},
			Tasks: []Task{
				{ID: "T1", PhaseID: "P1", Name: "Task 1", Status: StatusPlanning},
			},
			Tests: []Test{
				{ID: "TEST1", TaskID: "T1", Name: "Test 1", Status: StatusPlanning},
			},
		}

		result := epic.Validate()
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
		assert.Empty(t, result.Warnings)
		assert.Equal(t, "passed", result.Checks["xml_structure"])
		assert.Equal(t, "passed", result.Checks["status_values"])
		assert.Equal(t, "passed", result.Checks["task_phase_mapping"])
		assert.Equal(t, "passed", result.Checks["test_coverage"])
	})

	t.Run("epic with missing required fields fails validation", func(t *testing.T) {
		epic := &Epic{
			// Missing ID, Name, Status, CreatedAt
		}

		result := epic.Validate()
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors, "Epic ID is required")
		assert.Contains(t, result.Errors, "Epic name is required")
		assert.Contains(t, result.Errors, "Epic status is required")
		assert.Contains(t, result.Errors, "Epic created_at timestamp is required")
		assert.Equal(t, "failed", result.Checks["xml_structure"])
	})

	t.Run("epic with invalid status values passes validation (Epic 13 graceful handling)", func(t *testing.T) {
		epic := &Epic{
			ID:        "test-1",
			Name:      "Test Epic",
			Status:    Status("invalid_status"), // Gets converted to "pending" by Epic 13
			CreatedAt: time.Now(),
			Phases: []Phase{
				{ID: "P1", Name: "Phase 1", Status: Status("invalid_phase_status")}, // Gets converted to "pending"
			},
			Tasks: []Task{
				{ID: "T1", PhaseID: "P1", Name: "Task 1", Status: Status("invalid_task_status")}, // Gets converted to "pending"
			},
			Tests: []Test{
				{ID: "TEST1", TaskID: "T1", Name: "Test 1", Status: Status("invalid_test_status")}, // Gets converted to "pending"
			},
		}

		result := epic.Validate()
		// Epic 13 gracefully handles invalid legacy statuses by converting them to valid defaults
		assert.True(t, result.Valid, "Epic 13 should gracefully handle invalid legacy statuses")
		assert.Equal(t, "passed", result.Checks["status_values"])

		// Verify that conversion worked correctly
		assert.Equal(t, EpicStatusPending, epic.GetEpicStatus())
		assert.Equal(t, PhaseStatusPending, epic.Phases[0].GetPhaseStatus())
		assert.Equal(t, TaskStatusPending, epic.Tasks[0].GetTaskStatus())
		assert.Equal(t, TestStatusPending, epic.Tests[0].GetTestStatusUnified())
	})

	t.Run("epic with duplicate IDs fails validation", func(t *testing.T) {
		epic := &Epic{
			ID:        "test-1",
			Name:      "Test Epic",
			Status:    StatusPlanning,
			CreatedAt: time.Now(),
			Phases: []Phase{
				{ID: "P1", Name: "Phase 1", Status: StatusPlanning},
				{ID: "P1", Name: "Phase 1 Duplicate", Status: StatusPlanning}, // Duplicate ID
			},
			Tasks: []Task{
				{ID: "T1", PhaseID: "P1", Name: "Task 1", Status: StatusPlanning},
				{ID: "T1", PhaseID: "P1", Name: "Task 1 Duplicate", Status: StatusPlanning}, // Duplicate ID
			},
			Tests: []Test{
				{ID: "TEST1", TaskID: "T1", Name: "Test 1", Status: StatusPlanning},
				{ID: "TEST1", TaskID: "T1", Name: "Test 1 Duplicate", Status: StatusPlanning}, // Duplicate ID
			},
		}

		result := epic.Validate()
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors, "Duplicate phase ID: P1")
		assert.Contains(t, result.Errors, "Duplicate task ID: T1")
		assert.Contains(t, result.Errors, "Duplicate test ID: TEST1")
		assert.Equal(t, "failed", result.Checks["xml_structure"])
	})

	t.Run("task with invalid phase reference fails validation", func(t *testing.T) {
		epic := &Epic{
			ID:        "test-1",
			Name:      "Test Epic",
			Status:    StatusPlanning,
			CreatedAt: time.Now(),
			Phases: []Phase{
				{ID: "P1", Name: "Phase 1", Status: StatusPlanning},
			},
			Tasks: []Task{
				{ID: "T1", PhaseID: "P1", Name: "Task 1", Status: StatusPlanning},
				{ID: "T2", PhaseID: "P999", Name: "Task 2", Status: StatusPlanning}, // Invalid phase
			},
		}

		result := epic.Validate()
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors, "Task T2 references non-existent phase: P999")
		assert.Equal(t, "failed", result.Checks["task_phase_mapping"])
	})

	t.Run("test with invalid task reference fails validation", func(t *testing.T) {
		epic := &Epic{
			ID:        "test-1",
			Name:      "Test Epic",
			Status:    StatusPlanning,
			CreatedAt: time.Now(),
			Tasks: []Task{
				{ID: "T1", Name: "Task 1", Status: StatusPlanning},
			},
			Tests: []Test{
				{ID: "TEST1", TaskID: "T1", Name: "Test 1", Status: StatusPlanning},
				{ID: "TEST2", TaskID: "T999", Name: "Test 2", Status: StatusPlanning}, // Invalid task
			},
		}

		result := epic.Validate()
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors, "Test TEST2 references non-existent task: T999")
		assert.Equal(t, "failed", result.Checks["test_coverage"])
	})

	t.Run("task without tests generates warning", func(t *testing.T) {
		epic := &Epic{
			ID:        "test-1",
			Name:      "Test Epic",
			Status:    StatusPlanning,
			CreatedAt: time.Now(),
			Tasks: []Task{
				{ID: "T1", Name: "Task 1", Status: StatusPlanning},
				{ID: "T2", Name: "Task 2", Status: StatusPlanning}, // No tests
			},
			Tests: []Test{
				{ID: "TEST1", TaskID: "T1", Name: "Test 1", Status: StatusPlanning},
				// No test for T2
			},
		}

		result := epic.Validate()
		assert.True(t, result.Valid)
		assert.Contains(t, result.Warnings, "Task T2 has no tests defined")
		assert.Equal(t, "warning", result.Checks["test_coverage"])
	})

	t.Run("empty phases with empty ID fails validation", func(t *testing.T) {
		epic := &Epic{
			ID:        "test-1",
			Name:      "Test Epic",
			Status:    StatusPlanning,
			CreatedAt: time.Now(),
			Phases: []Phase{
				{ID: "", Name: "Phase with no ID", Status: StatusPlanning}, // Empty ID
			},
		}

		result := epic.Validate()
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors, "Phase ID is required")
		assert.Equal(t, "failed", result.Checks["xml_structure"])
	})

	t.Run("empty tasks with empty ID fails validation", func(t *testing.T) {
		epic := &Epic{
			ID:        "test-1",
			Name:      "Test Epic",
			Status:    StatusPlanning,
			CreatedAt: time.Now(),
			Tasks: []Task{
				{ID: "", Name: "Task with no ID", Status: StatusPlanning}, // Empty ID
			},
		}

		result := epic.Validate()
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors, "Task ID is required")
		assert.Equal(t, "failed", result.Checks["xml_structure"])
	})

	t.Run("empty tests with empty ID fails validation", func(t *testing.T) {
		epic := &Epic{
			ID:        "test-1",
			Name:      "Test Epic",
			Status:    StatusPlanning,
			CreatedAt: time.Now(),
			Tests: []Test{
				{ID: "", Name: "Test with no ID", Status: StatusPlanning}, // Empty ID
			},
		}

		result := epic.Validate()
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors, "Test ID is required")
		assert.Equal(t, "failed", result.Checks["xml_structure"])
	})
}

func TestValidationResult_Message(t *testing.T) {
	t.Run("valid result message", func(t *testing.T) {
		result := &ValidationResult{Valid: true}
		assert.Equal(t, "Epic structure is valid", result.Message())
	})

	t.Run("valid result with warnings message", func(t *testing.T) {
		result := &ValidationResult{
			Valid:    true,
			Warnings: []string{"warning 1", "warning 2"},
		}
		assert.Equal(t, "Epic structure is valid with 2 warning(s)", result.Message())
	})

	t.Run("invalid result message", func(t *testing.T) {
		result := &ValidationResult{
			Valid:  false,
			Errors: []string{"error 1", "error 2", "error 3"},
		}
		assert.Equal(t, "Epic validation failed with 3 error(s)", result.Message())
	})
}

func TestValidationResult_AddError(t *testing.T) {
	result := &ValidationResult{Valid: true}

	result.AddError("test error")

	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors, "test error")
}

func TestValidationResult_AddWarning(t *testing.T) {
	result := &ValidationResult{Valid: true}

	result.AddWarning("test warning")

	assert.True(t, result.Valid) // Warnings don't make result invalid
	assert.Contains(t, result.Warnings, "test warning")
}

func TestValidationResult_SetCheck(t *testing.T) {
	result := &ValidationResult{}

	result.SetCheck("test_check", "passed")

	assert.Equal(t, "passed", result.Checks["test_check"])
}

func TestStatus_IsValid(t *testing.T) {
	validStatuses := []Status{
		StatusPlanning,
		StatusActive,
		StatusCompleted,
		StatusOnHold,
		StatusCancelled,
	}

	for _, status := range validStatuses {
		assert.True(t, status.IsValid(), "Status %s should be valid", status)
	}

	invalidStatuses := []Status{
		Status("invalid"),
		Status(""),
		Status("unknown"),
		Status("PLANNING"), // Wrong case
	}

	for _, status := range invalidStatuses {
		assert.False(t, status.IsValid(), "Status %s should be invalid", status)
	}
}

func TestFormatValidationResult(t *testing.T) {
	result := &ValidationResult{
		Valid:    false,
		Warnings: []string{"warning 1"},
		Errors:   []string{"error 1"},
		Checks:   map[string]string{"check1": "passed", "check2": "failed"},
	}

	t.Run("format as text", func(t *testing.T) {
		output := FormatValidationResult(result, "text")
		assert.Contains(t, output, "✗ Epic validation failed")
		assert.Contains(t, output, "warning 1")
		assert.Contains(t, output, "error 1")
		assert.Contains(t, output, "✓ check1: passed")
		assert.Contains(t, output, "✗ check2: failed")
	})

	t.Run("format as XML", func(t *testing.T) {
		output := FormatValidationResult(result, "xml")
		assert.Contains(t, output, "<validation_result>")
		assert.Contains(t, output, "<valid>false</valid>")
		assert.Contains(t, output, "<warning>warning 1</warning>")
		assert.Contains(t, output, "<error>error 1</error>")
		assert.Contains(t, output, "<check name=\"check1\">passed</check>")
		assert.Contains(t, output, "</validation_result>")
	})

	t.Run("format as JSON", func(t *testing.T) {
		output := FormatValidationResult(result, "json")
		assert.Contains(t, output, "\"valid\": false")
		assert.Contains(t, output, "\"warnings\": [\"warning 1\"]")
		assert.Contains(t, output, "\"errors\": [\"error 1\"]")
	})

	t.Run("format with default (text)", func(t *testing.T) {
		output := FormatValidationResult(result, "unknown")
		assert.Contains(t, output, "✗ Epic validation failed")
	})
}

func TestValidateFromFile(t *testing.T) {
	// Mock storage
	mockStorage := &MockStorage{
		epics: make(map[string]*Epic),
	}

	t.Run("validate valid epic from file", func(t *testing.T) {
		validEpic := &Epic{
			ID:        "test-1",
			Name:      "Test Epic",
			Status:    StatusPlanning,
			CreatedAt: time.Now(),
		}

		mockStorage.epics["valid.xml"] = validEpic

		result, err := ValidateFromFile(mockStorage, "valid.xml")
		require.NoError(t, err)
		assert.True(t, result.Valid)
	})

	t.Run("validate invalid epic from file", func(t *testing.T) {
		invalidEpic := &Epic{
			// Missing required fields
		}

		mockStorage.epics["invalid.xml"] = invalidEpic

		result, err := ValidateFromFile(mockStorage, "invalid.xml")
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := ValidateFromFile(mockStorage, "missing.xml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load epic")
	})
}

// MockStorage for testing
type MockStorage struct {
	epics map[string]*Epic
}

func (ms *MockStorage) LoadEpic(filePath string) (*Epic, error) {
	if epic, exists := ms.epics[filePath]; exists {
		return epic, nil
	}
	return nil, fmt.Errorf("epic file not found: %s", filePath)
}

func TestNewEpic(t *testing.T) {
	epic := NewEpic("test-id", "Test Epic")

	assert.Equal(t, "test-id", epic.ID)
	assert.Equal(t, "Test Epic", epic.Name)
	assert.Equal(t, StatusPlanning, epic.Status)
	assert.False(t, epic.CreatedAt.IsZero())
}
