package tasks

import (
	"strings"
	"testing"

	"github.com/mindreframer/agentpm/internal/epic"
)

func TestTaskValidationService_ValidateTaskCompletion(t *testing.T) {
	tvs := NewTaskValidationService()

	t.Run("task already completed should pass", func(t *testing.T) {
		epicData := &epic.Epic{}
		task := &epic.Task{
			ID:     "task1",
			Name:   "Test Task",
			Status: epic.StatusCompleted,
		}

		err := tvs.ValidateTaskCompletion(epicData, task)
		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
	})

	t.Run("task with no blocking tests should pass", func(t *testing.T) {
		epicData := &epic.Epic{
			Tests: []epic.Test{},
		}
		task := &epic.Task{
			ID:     "task1",
			Name:   "Test Task",
			Status: epic.StatusActive,
		}

		err := tvs.ValidateTaskCompletion(epicData, task)
		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
	})

	t.Run("task with pending tests should fail", func(t *testing.T) {
		epicData := &epic.Epic{
			Tests: []epic.Test{
				{
					ID:         "test1",
					TaskID:     "task1",
					Name:       "Pending Test",
					TestStatus: epic.TestStatusPending,
				},
			},
		}
		task := &epic.Task{
			ID:     "task1",
			Name:   "Test Task",
			Status: epic.StatusActive,
		}

		err := tvs.ValidateTaskCompletion(epicData, task)
		if err == nil {
			t.Error("Expected error but got none")
			return
		}

		statusErr, ok := err.(*epic.StatusValidationError)
		if !ok {
			t.Errorf("Expected StatusValidationError, got %T", err)
			return
		}

		if statusErr.EntityType != "task" {
			t.Errorf("Expected EntityType 'task', got '%s'", statusErr.EntityType)
		}
		if statusErr.EntityID != task.ID {
			t.Errorf("Expected EntityID '%s', got '%s'", task.ID, statusErr.EntityID)
		}
	})

	t.Run("task with wip tests should fail", func(t *testing.T) {
		epicData := &epic.Epic{
			Tests: []epic.Test{
				{
					ID:         "test1",
					TaskID:     "task1",
					Name:       "WIP Test",
					TestStatus: epic.TestStatusWIP,
				},
			},
		}
		task := &epic.Task{
			ID:     "task1",
			Name:   "Test Task",
			Status: epic.StatusActive,
		}

		err := tvs.ValidateTaskCompletion(epicData, task)
		if err == nil {
			t.Error("Expected error but got none")
		}
	})

	t.Run("task with multiple blocking tests should fail with all items", func(t *testing.T) {
		epicData := &epic.Epic{
			Tests: []epic.Test{
				{
					ID:         "test1",
					TaskID:     "task1",
					Name:       "Pending Test",
					TestStatus: epic.TestStatusPending,
				},
				{
					ID:         "test2",
					TaskID:     "task1",
					Name:       "WIP Test",
					TestStatus: epic.TestStatusWIP,
				},
			},
		}
		task := &epic.Task{
			ID:     "task1",
			Name:   "Test Task",
			Status: epic.StatusActive,
		}

		err := tvs.ValidateTaskCompletion(epicData, task)
		if err == nil {
			t.Error("Expected error but got none")
			return
		}

		statusErr, ok := err.(*epic.StatusValidationError)
		if !ok {
			t.Errorf("Expected StatusValidationError, got %T", err)
			return
		}

		if len(statusErr.BlockingItems) != 2 {
			t.Errorf("Expected 2 blocking items, got %d", len(statusErr.BlockingItems))
		}
	})

	t.Run("task with done tests should pass", func(t *testing.T) {
		epicData := &epic.Epic{
			Tests: []epic.Test{
				{
					ID:         "test1",
					TaskID:     "task1",
					Name:       "Done Test",
					TestStatus: epic.TestStatusDone,
				},
			},
		}
		task := &epic.Task{
			ID:     "task1",
			Name:   "Test Task",
			Status: epic.StatusActive,
		}

		err := tvs.ValidateTaskCompletion(epicData, task)
		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
	})

	t.Run("task with cancelled tests should pass", func(t *testing.T) {
		epicData := &epic.Epic{
			Tests: []epic.Test{
				{
					ID:         "test1",
					TaskID:     "task1",
					Name:       "Cancelled Test",
					TestStatus: epic.TestStatusCancelled,
				},
			},
		}
		task := &epic.Task{
			ID:     "task1",
			Name:   "Test Task",
			Status: epic.StatusActive,
		}

		err := tvs.ValidateTaskCompletion(epicData, task)
		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
	})

	t.Run("task with tests from different task should pass", func(t *testing.T) {
		epicData := &epic.Epic{
			Tests: []epic.Test{
				{
					ID:         "test1",
					TaskID:     "other_task",
					Name:       "Test in different task",
					TestStatus: epic.TestStatusPending,
				},
			},
		}
		task := &epic.Task{
			ID:     "task1",
			Name:   "Test Task",
			Status: epic.StatusActive,
		}

		err := tvs.ValidateTaskCompletion(epicData, task)
		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
	})
}

func TestTaskValidationService_ValidateTaskStatusTransition(t *testing.T) {
	tvs := NewTaskValidationService()

	tests := []struct {
		name          string
		currentStatus epic.Status
		targetStatus  epic.Status
		wantError     bool
	}{
		{
			name:          "pending to active should be valid",
			currentStatus: epic.StatusPending,
			targetStatus:  epic.StatusActive,
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
			currentStatus: epic.StatusActive,
			targetStatus:  epic.StatusCompleted,
			wantError:     false,
		},
		{
			name:          "active to cancelled should be valid",
			currentStatus: epic.StatusActive,
			targetStatus:  epic.StatusCancelled,
			wantError:     false,
		},
		{
			name:          "active to on_hold should be valid",
			currentStatus: epic.StatusActive,
			targetStatus:  epic.StatusOnHold,
			wantError:     false,
		},
		{
			name:          "on_hold to active should be valid",
			currentStatus: epic.StatusOnHold,
			targetStatus:  epic.StatusActive,
			wantError:     false,
		},
		{
			name:          "on_hold to cancelled should be valid",
			currentStatus: epic.StatusOnHold,
			targetStatus:  epic.StatusCancelled,
			wantError:     false,
		},
		{
			name:          "pending to completed should be invalid",
			currentStatus: epic.StatusPending,
			targetStatus:  epic.StatusCompleted,
			wantError:     true,
		},
		{
			name:          "pending to on_hold should be invalid",
			currentStatus: epic.StatusPending,
			targetStatus:  epic.StatusOnHold,
			wantError:     true,
		},
		{
			name:          "completed to any status should be invalid",
			currentStatus: epic.StatusCompleted,
			targetStatus:  epic.StatusActive,
			wantError:     true,
		},
		{
			name:          "cancelled to any status should be invalid",
			currentStatus: epic.StatusCancelled,
			targetStatus:  epic.StatusActive,
			wantError:     true,
		},
		{
			name:          "on_hold to completed should be invalid",
			currentStatus: epic.StatusOnHold,
			targetStatus:  epic.StatusCompleted,
			wantError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tvs.ValidateTaskStatusTransition(tt.currentStatus, tt.targetStatus)

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

func TestTaskValidationService_ValidateTaskCancellation(t *testing.T) {
	tvs := NewTaskValidationService()

	t.Run("valid cancellation with reason should pass", func(t *testing.T) {
		task := &epic.Task{
			ID:     "task1",
			Status: epic.StatusActive,
		}

		err := tvs.ValidateTaskCancellation(task, "No longer needed")
		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
	})

	t.Run("cancellation without reason should fail", func(t *testing.T) {
		task := &epic.Task{
			ID:     "task1",
			Status: epic.StatusActive,
		}

		err := tvs.ValidateTaskCancellation(task, "")
		if err == nil {
			t.Error("Expected error but got none")
		}
		if !strings.Contains(err.Error(), "cancellation reason is required") {
			t.Errorf("Expected error about missing reason, got: %v", err)
		}
	})

	t.Run("cannot cancel already cancelled task", func(t *testing.T) {
		task := &epic.Task{
			ID:     "task1",
			Status: epic.StatusCancelled,
		}

		err := tvs.ValidateTaskCancellation(task, "Some reason")
		if err == nil {
			t.Error("Expected error but got none")
		}
		if !strings.Contains(err.Error(), "already cancelled") {
			t.Errorf("Expected error about already cancelled, got: %v", err)
		}
	})

	t.Run("cannot cancel completed task", func(t *testing.T) {
		task := &epic.Task{
			ID:     "task1",
			Status: epic.StatusCompleted,
		}

		err := tvs.ValidateTaskCancellation(task, "Some reason")
		if err == nil {
			t.Error("Expected error but got none")
		}
		if !strings.Contains(err.Error(), "cannot cancel completed task") {
			t.Errorf("Expected error about cancelling completed task, got: %v", err)
		}
	})

	t.Run("nil task should fail", func(t *testing.T) {
		err := tvs.ValidateTaskCancellation(nil, "Some reason")
		if err == nil {
			t.Error("Expected error but got none")
		}
		if !strings.Contains(err.Error(), "cannot be nil") {
			t.Errorf("Expected error about nil task, got: %v", err)
		}
	})
}

func TestTaskValidationService_CountTestsByStatus(t *testing.T) {
	tvs := NewTaskValidationService()

	epicData := &epic.Epic{
		Tests: []epic.Test{
			{ID: "test1", TaskID: "task1", TestStatus: epic.TestStatusPending},
			{ID: "test2", TaskID: "task1", TestStatus: epic.TestStatusWIP},
			{ID: "test3", TaskID: "task1", TestStatus: epic.TestStatusDone},
			{ID: "test4", TaskID: "task2", TestStatus: epic.TestStatusPending}, // Different task
		},
	}

	pending, wip := tvs.countTestsByStatus(epicData, "task1")
	if pending != 1 {
		t.Errorf("Expected 1 pending test, got %d", pending)
	}
	if wip != 1 {
		t.Errorf("Expected 1 wip test, got %d", wip)
	}
}

func TestTaskValidationService_CanCompleteTask(t *testing.T) {
	tvs := NewTaskValidationService()

	t.Run("can complete with no blocking tests", func(t *testing.T) {
		epicData := &epic.Epic{
			Tests: []epic.Test{},
		}
		task := &epic.Task{
			ID:     "task1",
			Status: epic.StatusActive,
		}

		canComplete, err := tvs.CanCompleteTask(epicData, task)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !canComplete {
			t.Errorf("Expected can complete to be true")
		}
	})

	t.Run("cannot complete with blocking tests", func(t *testing.T) {
		epicData := &epic.Epic{
			Tests: []epic.Test{
				{ID: "test1", TaskID: "task1", TestStatus: epic.TestStatusPending},
			},
		}
		task := &epic.Task{
			ID:     "task1",
			Status: epic.StatusActive,
		}

		canComplete, err := tvs.CanCompleteTask(epicData, task)
		if err == nil {
			t.Errorf("Expected error but got none")
		}
		if canComplete {
			t.Errorf("Expected can complete to be false")
		}
	})
}

func TestTaskValidationService_NilTask(t *testing.T) {
	tvs := NewTaskValidationService()

	err := tvs.ValidateTaskCompletion(&epic.Epic{}, nil)
	if err == nil {
		t.Errorf("Expected error for nil task")
	}
	if !strings.Contains(err.Error(), "cannot be nil") {
		t.Errorf("Expected error message about nil task, got: %v", err)
	}
}

func TestTaskValidationService_ErrorMessageContainsExactCounts(t *testing.T) {
	tvs := NewTaskValidationService()

	epicData := &epic.Epic{
		Tests: []epic.Test{
			{ID: "test1", TaskID: "task1", TestStatus: epic.TestStatusPending},
			{ID: "test2", TaskID: "task1", TestStatus: epic.TestStatusWIP},
			{ID: "test3", TaskID: "task1", TestStatus: epic.TestStatusWIP},
		},
	}

	task := &epic.Task{
		ID:     "task1",
		Status: epic.StatusActive,
	}

	err := tvs.ValidateTaskCompletion(epicData, task)
	if err == nil {
		t.Fatalf("Expected error but got none")
	}

	statusErr, ok := err.(*epic.StatusValidationError)
	if !ok {
		t.Fatalf("Expected StatusValidationError, got %T", err)
	}

	// Check that the message contains exact counts
	if !strings.Contains(statusErr.Message, "1 pending tests") {
		t.Errorf("Expected message to contain '1 pending tests', got: %s", statusErr.Message)
	}
	if !strings.Contains(statusErr.Message, "2 wip tests") {
		t.Errorf("Expected message to contain '2 wip tests', got: %s", statusErr.Message)
	}

	// Check that we have 3 blocking items total (1 pending test + 2 wip tests)
	if len(statusErr.BlockingItems) != 3 {
		t.Errorf("Expected 3 blocking items, got %d", len(statusErr.BlockingItems))
	}
}
