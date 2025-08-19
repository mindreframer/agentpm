package phases

import (
	"strings"
	"testing"

	"github.com/mindreframer/agentpm/internal/epic"
)

func TestPhaseValidationService_ValidatePhaseCompletion(t *testing.T) {
	pvs := NewPhaseValidationService()

	t.Run("phase already completed should pass", func(t *testing.T) {
		epicData := &epic.Epic{}
		phase := &epic.Phase{
			ID:     "phase1",
			Name:   "Test Phase",
			Status: epic.StatusCompleted,
		}

		err := pvs.ValidatePhaseCompletion(epicData, phase)
		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
	})

	t.Run("phase with no blocking items should pass", func(t *testing.T) {
		epicData := &epic.Epic{
			Tasks: []epic.Task{},
			Tests: []epic.Test{},
		}
		phase := &epic.Phase{
			ID:     "phase1",
			Name:   "Test Phase",
			Status: epic.StatusWIP,
		}

		err := pvs.ValidatePhaseCompletion(epicData, phase)
		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
	})

	t.Run("phase with pending tasks should fail", func(t *testing.T) {
		epicData := &epic.Epic{
			Tasks: []epic.Task{
				{
					ID:      "task1",
					PhaseID: "phase1",
					Name:    "Pending Task",
					Status:  epic.StatusPending,
				},
			},
			Tests: []epic.Test{},
		}
		phase := &epic.Phase{
			ID:     "phase1",
			Name:   "Test Phase",
			Status: epic.StatusWIP,
		}

		err := pvs.ValidatePhaseCompletion(epicData, phase)
		if err == nil {
			t.Error("Expected error but got none")
			return
		}

		statusErr, ok := err.(*epic.StatusValidationError)
		if !ok {
			t.Errorf("Expected StatusValidationError, got %T", err)
			return
		}

		if statusErr.EntityType != "phase" {
			t.Errorf("Expected EntityType 'phase', got '%s'", statusErr.EntityType)
		}
		if statusErr.EntityID != phase.ID {
			t.Errorf("Expected EntityID '%s', got '%s'", phase.ID, statusErr.EntityID)
		}
	})

	t.Run("phase with active tasks should fail", func(t *testing.T) {
		epicData := &epic.Epic{
			Tasks: []epic.Task{
				{
					ID:      "task1",
					PhaseID: "phase1",
					Name:    "Active Task",
					Status:  epic.StatusWIP,
				},
			},
			Tests: []epic.Test{},
		}
		phase := &epic.Phase{
			ID:     "phase1",
			Name:   "Test Phase",
			Status: epic.StatusWIP,
		}

		err := pvs.ValidatePhaseCompletion(epicData, phase)
		if err == nil {
			t.Error("Expected error but got none")
		}
	})

	t.Run("phase with pending tests should fail", func(t *testing.T) {
		epicData := &epic.Epic{
			Tasks: []epic.Task{},
			Tests: []epic.Test{
				{
					ID:         "test1",
					PhaseID:    "phase1",
					Name:       "Pending Test",
					TestStatus: epic.TestStatusPending,
				},
			},
		}
		phase := &epic.Phase{
			ID:     "phase1",
			Name:   "Test Phase",
			Status: epic.StatusWIP,
		}

		err := pvs.ValidatePhaseCompletion(epicData, phase)
		if err == nil {
			t.Error("Expected error but got none")
		}
	})

	t.Run("phase with wip tests should fail", func(t *testing.T) {
		epicData := &epic.Epic{
			Tasks: []epic.Task{},
			Tests: []epic.Test{
				{
					ID:         "test1",
					PhaseID:    "phase1",
					Name:       "WIP Test",
					TestStatus: epic.TestStatusWIP,
				},
			},
		}
		phase := &epic.Phase{
			ID:     "phase1",
			Name:   "Test Phase",
			Status: epic.StatusWIP,
		}

		err := pvs.ValidatePhaseCompletion(epicData, phase)
		if err == nil {
			t.Error("Expected error but got none")
		}
	})
}

func TestPhaseValidationService_ValidatePhaseStatusTransition(t *testing.T) {
	pvs := NewPhaseValidationService()

	tests := []struct {
		name          string
		currentStatus epic.Status
		targetStatus  epic.Status
		wantError     bool
	}{
		{
			name:          "pending to active should be valid",
			currentStatus: epic.StatusPending,
			targetStatus:  epic.StatusWIP,
			wantError:     false,
		},
		{
			name:          "pending to cancelled should be valid",
			currentStatus: epic.StatusPending,
			targetStatus:  epic.StatusCancelled,
			wantError:     false,
		},
		{
			name:          "active to completed should be valid",
			currentStatus: epic.StatusWIP,
			targetStatus:  epic.StatusCompleted,
			wantError:     false,
		},
		{
			name:          "pending to completed should be invalid",
			currentStatus: epic.StatusPending,
			targetStatus:  epic.StatusCompleted,
			wantError:     true,
		},
		{
			name:          "completed to any status should be invalid",
			currentStatus: epic.StatusCompleted,
			targetStatus:  epic.StatusWIP,
			wantError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pvs.ValidatePhaseStatusTransition(tt.currentStatus, tt.targetStatus)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestPhaseValidationService_CountMethods(t *testing.T) {
	pvs := NewPhaseValidationService()

	epicData := &epic.Epic{
		Tasks: []epic.Task{
			{ID: "task1", PhaseID: "phase1", Status: epic.StatusPending},
			{ID: "task2", PhaseID: "phase1", Status: epic.StatusWIP},
			{ID: "task3", PhaseID: "phase1", Status: epic.StatusCompleted},
			{ID: "task4", PhaseID: "phase2", Status: epic.StatusPending}, // Different phase
		},
		Tests: []epic.Test{
			{ID: "test1", PhaseID: "phase1", TestStatus: epic.TestStatusPending},
			{ID: "test2", PhaseID: "phase1", TestStatus: epic.TestStatusWIP},
			{ID: "test3", PhaseID: "phase1", TestStatus: epic.TestStatusDone},
			{ID: "test4", PhaseID: "phase2", TestStatus: epic.TestStatusPending}, // Different phase
		},
	}

	t.Run("countTasksByStatus", func(t *testing.T) {
		pending, active := pvs.countTasksByStatus(epicData, "phase1")
		if pending != 1 {
			t.Errorf("Expected 1 pending task, got %d", pending)
		}
		if active != 1 {
			t.Errorf("Expected 1 active task, got %d", active)
		}
	})

	t.Run("countTestsByStatus", func(t *testing.T) {
		pending, wip := pvs.countTestsByStatus(epicData, "phase1")
		if pending != 1 {
			t.Errorf("Expected 1 pending test, got %d", pending)
		}
		if wip != 1 {
			t.Errorf("Expected 1 wip test, got %d", wip)
		}
	})
}

func TestPhaseValidationService_NilPhase(t *testing.T) {
	pvs := NewPhaseValidationService()

	err := pvs.ValidatePhaseCompletion(&epic.Epic{}, nil)
	if err == nil {
		t.Errorf("Expected error for nil phase")
	}
	if !strings.Contains(err.Error(), "cannot be nil") {
		t.Errorf("Expected error message about nil phase, got: %v", err)
	}
}

func TestPhaseValidationService_ErrorMessageContainsExactCounts(t *testing.T) {
	pvs := NewPhaseValidationService()

	epicData := &epic.Epic{
		Tasks: []epic.Task{
			{ID: "task1", PhaseID: "phase1", Status: epic.StatusPending},
			{ID: "task2", PhaseID: "phase1", Status: epic.StatusPending},
			{ID: "task3", PhaseID: "phase1", Status: epic.StatusWIP},
		},
		Tests: []epic.Test{
			{ID: "test1", PhaseID: "phase1", TestStatus: epic.TestStatusPending},
			{ID: "test2", PhaseID: "phase1", TestStatus: epic.TestStatusWIP},
			{ID: "test3", PhaseID: "phase1", TestStatus: epic.TestStatusWIP},
			{ID: "test4", PhaseID: "phase1", TestStatus: epic.TestStatusWIP},
		},
	}

	phase := &epic.Phase{
		ID:     "phase1",
		Status: epic.StatusWIP,
	}

	err := pvs.ValidatePhaseCompletion(epicData, phase)
	if err == nil {
		t.Fatalf("Expected error but got none")
	}

	statusErr, ok := err.(*epic.StatusValidationError)
	if !ok {
		t.Fatalf("Expected StatusValidationError, got %T", err)
	}

	// Check that the message contains exact counts
	if !strings.Contains(statusErr.Message, "2 pending tasks") {
		t.Errorf("Expected message to contain '2 pending tasks', got: %s", statusErr.Message)
	}
	if !strings.Contains(statusErr.Message, "1 active tasks") {
		t.Errorf("Expected message to contain '1 active tasks', got: %s", statusErr.Message)
	}
	if !strings.Contains(statusErr.Message, "1 pending tests") {
		t.Errorf("Expected message to contain '1 pending tests', got: %s", statusErr.Message)
	}
	if !strings.Contains(statusErr.Message, "3 wip tests") {
		t.Errorf("Expected message to contain '3 wip tests', got: %s", statusErr.Message)
	}

	// Check that we have 7 blocking items total (2 pending tasks + 1 active task + 1 pending test + 3 wip tests)
	if len(statusErr.BlockingItems) != 7 {
		t.Errorf("Expected 7 blocking items, got %d", len(statusErr.BlockingItems))
		for i, item := range statusErr.BlockingItems {
			t.Logf("Item %d: Type=%s, ID=%s, Status=%s", i, item.Type, item.ID, item.Status)
		}
	}
}
