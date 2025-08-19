package tasks

import (
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskService_StartTask(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	taskService := NewTaskService(storage, queryService)
	testTime := time.Date(2025, 8, 16, 15, 30, 0, 0, time.UTC)

	t.Run("start task in active phase", func(t *testing.T) {
		// AC-5: Start task in active phase
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPending},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusPending},
			},
		}

		err := taskService.StartTask(epicData, "task-1", testTime)
		require.NoError(t, err)

		// Verify task status changed to active
		task := findTaskByID(epicData, "task-1")
		require.NotNil(t, task)
		assert.Equal(t, epic.StatusWIP, task.Status)
		assert.Equal(t, testTime, *task.StartedAt)

		// Verify it's the active task in the phase
		activeTask := taskService.GetActiveTask(epicData, "phase-1")
		require.NotNil(t, activeTask)
		assert.Equal(t, "task-1", activeTask.ID)
	})

	t.Run("prevent starting task in non-active phase", func(t *testing.T) {
		// AC-6: Prevent starting task in non-active phase
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPending},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPending},
				{ID: "task-2", PhaseID: "phase-2", Name: "Task 2", Status: epic.StatusPending},
			},
		}

		err := taskService.StartTask(epicData, "task-2", testTime)
		require.Error(t, err)

		// Verify it's a phase error
		var phaseErr *TaskPhaseError
		require.ErrorAs(t, err, &phaseErr)
		assert.Equal(t, "task-2", phaseErr.TaskID)
		assert.Equal(t, "phase-2", phaseErr.PhaseID)
		assert.Equal(t, epic.StatusPending, phaseErr.PhaseStatus)

		// Verify task is still pending
		task := findTaskByID(epicData, "task-2")
		require.NotNil(t, task)
		assert.Equal(t, epic.StatusPending, task.Status)
	})

	t.Run("prevent multiple active tasks in same phase", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusWIP},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusPending},
			},
		}

		err := taskService.StartTask(epicData, "task-2", testTime)
		require.Error(t, err)

		// Verify it's a constraint error
		var constraintErr *TaskConstraintError
		require.ErrorAs(t, err, &constraintErr)
		assert.Equal(t, "task-2", constraintErr.TaskID)
		assert.Equal(t, "task-1", constraintErr.ActiveTaskID)
		assert.Equal(t, "phase-1", constraintErr.PhaseID)

		// Verify task-2 is still pending
		task := findTaskByID(epicData, "task-2")
		require.NotNil(t, task)
		assert.Equal(t, epic.StatusPending, task.Status)
	})

	t.Run("cannot start task that is not pending", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
			},
		}

		err := taskService.StartTask(epicData, "task-1", testTime)
		require.Error(t, err)

		// Verify it's a state error
		var stateErr *TaskStateError
		require.ErrorAs(t, err, &stateErr)
		assert.Equal(t, "task-1", stateErr.TaskID)
		assert.Equal(t, epic.StatusCompleted, stateErr.CurrentStatus)
		assert.Equal(t, epic.StatusWIP, stateErr.TargetStatus)
	})

	t.Run("cannot start non-existent task", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
			},
			Tasks: []epic.Task{},
		}

		err := taskService.StartTask(epicData, "non-existent", testTime)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "task non-existent not found")
	})
}

func TestTaskService_CompleteTask(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	taskService := NewTaskService(storage, queryService)
	testTime := time.Date(2025, 8, 16, 16, 30, 0, 0, time.UTC)

	t.Run("complete active task", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusWIP},
			},
		}

		err := taskService.CompleteTask(epicData, "task-1", testTime)
		require.NoError(t, err)

		// Verify task status changed to completed
		task := findTaskByID(epicData, "task-1")
		require.NotNil(t, task)
		assert.Equal(t, epic.StatusCompleted, task.Status)
		assert.Equal(t, testTime, *task.CompletedAt)

		// Verify no active task in the phase
		activeTask := taskService.GetActiveTask(epicData, "phase-1")
		assert.Nil(t, activeTask)
	})

	t.Run("cannot complete task that is not active", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPending},
			},
		}

		err := taskService.CompleteTask(epicData, "task-1", testTime)
		require.Error(t, err)

		// Verify it's a state error
		var stateErr *TaskStateError
		require.ErrorAs(t, err, &stateErr)
		assert.Equal(t, "task-1", stateErr.TaskID)
		assert.Equal(t, epic.StatusPending, stateErr.CurrentStatus)
		assert.Equal(t, epic.StatusCompleted, stateErr.TargetStatus)
	})

	t.Run("cannot complete non-existent task", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Tasks:  []epic.Task{},
		}

		err := taskService.CompleteTask(epicData, "non-existent", testTime)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "task non-existent not found")
	})
}

func TestTaskService_CancelTask(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	taskService := NewTaskService(storage, queryService)
	testTime := time.Date(2025, 8, 16, 16, 30, 0, 0, time.UTC)

	t.Run("cancel active task", func(t *testing.T) {
		// AC-10: Cancel active task
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusWIP},
			},
		}

		err := taskService.CancelTask(epicData, "task-1", testTime)
		require.NoError(t, err)

		// Verify task status changed to cancelled
		task := findTaskByID(epicData, "task-1")
		require.NotNil(t, task)
		assert.Equal(t, epic.StatusCancelled, task.Status)
		assert.Equal(t, testTime, *task.CancelledAt)

		// Verify no active task in the phase
		activeTask := taskService.GetActiveTask(epicData, "phase-1")
		assert.Nil(t, activeTask)
	})

	t.Run("cannot cancel task that is not active", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPending},
			},
		}

		err := taskService.CancelTask(epicData, "task-1", testTime)
		require.Error(t, err)

		// Verify it's a state error
		var stateErr *TaskStateError
		require.ErrorAs(t, err, &stateErr)
		assert.Equal(t, "task-1", stateErr.TaskID)
		assert.Equal(t, epic.StatusPending, stateErr.CurrentStatus)
		assert.Equal(t, epic.StatusCancelled, stateErr.TargetStatus)
	})
}

func TestTaskService_GetActiveTask(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	taskService := NewTaskService(storage, queryService)

	t.Run("returns active task when exists in phase", func(t *testing.T) {
		epicData := &epic.Epic{
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPending},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusWIP},
				{ID: "task-3", PhaseID: "phase-2", Name: "Task 3", Status: epic.StatusWIP},
			},
		}

		activeTask := taskService.GetActiveTask(epicData, "phase-1")
		require.NotNil(t, activeTask)
		assert.Equal(t, "task-2", activeTask.ID)
	})

	t.Run("returns nil when no active task in phase", func(t *testing.T) {
		epicData := &epic.Epic{
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPending},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusCompleted},
			},
		}

		activeTask := taskService.GetActiveTask(epicData, "phase-1")
		assert.Nil(t, activeTask)
	})

	t.Run("returns nil when no tasks in phase", func(t *testing.T) {
		epicData := &epic.Epic{
			Tasks: []epic.Task{},
		}

		activeTask := taskService.GetActiveTask(epicData, "phase-1")
		assert.Nil(t, activeTask)
	})
}

func TestTaskService_GetActiveTaskInEpic(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	taskService := NewTaskService(storage, queryService)

	t.Run("returns active task when exists in epic", func(t *testing.T) {
		epicData := &epic.Epic{
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPending},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusWIP},
				{ID: "task-3", PhaseID: "phase-2", Name: "Task 3", Status: epic.StatusCompleted},
			},
		}

		activeTask := taskService.GetActiveTaskInEpic(epicData)
		require.NotNil(t, activeTask)
		assert.Equal(t, "task-2", activeTask.ID)
	})

	t.Run("returns nil when no active task in epic", func(t *testing.T) {
		epicData := &epic.Epic{
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPending},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusCompleted},
			},
		}

		activeTask := taskService.GetActiveTaskInEpic(epicData)
		assert.Nil(t, activeTask)
	})
}

func TestTaskService_SingleActiveTaskConstraint(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	taskService := NewTaskService(storage, queryService)
	testTime := time.Date(2025, 8, 16, 15, 30, 0, 0, time.UTC)

	t.Run("only one task can be active per phase", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPending},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusPending},
				{ID: "task-3", PhaseID: "phase-1", Name: "Task 3", Status: epic.StatusPending},
			},
		}

		// Start first task
		err := taskService.StartTask(epicData, "task-1", testTime)
		require.NoError(t, err)

		// Verify task-1 is active
		activeTask := taskService.GetActiveTask(epicData, "phase-1")
		require.NotNil(t, activeTask)
		assert.Equal(t, "task-1", activeTask.ID)

		// Try to start second task - should fail
		err = taskService.StartTask(epicData, "task-2", testTime)
		require.Error(t, err)
		var constraintErr *TaskConstraintError
		require.ErrorAs(t, err, &constraintErr)

		// Verify task-1 is still the only active task
		activeTask = taskService.GetActiveTask(epicData, "phase-1")
		require.NotNil(t, activeTask)
		assert.Equal(t, "task-1", activeTask.ID)

		// Verify task-2 is still pending
		task2 := findTaskByID(epicData, "task-2")
		require.NotNil(t, task2)
		assert.Equal(t, epic.StatusPending, task2.Status)
	})
}

func TestTaskService_GetTasksInPhase(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	taskService := NewTaskService(storage, queryService)

	t.Run("returns all tasks for phase", func(t *testing.T) {
		epicData := &epic.Epic{
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPending},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusWIP},
				{ID: "task-3", PhaseID: "phase-2", Name: "Task 3", Status: epic.StatusCompleted},
				{ID: "task-4", PhaseID: "phase-1", Name: "Task 4", Status: epic.StatusCancelled},
			},
		}

		tasks := taskService.GetTasksInPhase(epicData, "phase-1")
		assert.Len(t, tasks, 3)

		taskIDs := make([]string, len(tasks))
		for i, task := range tasks {
			taskIDs[i] = task.ID
		}
		assert.Contains(t, taskIDs, "task-1")
		assert.Contains(t, taskIDs, "task-2")
		assert.Contains(t, taskIDs, "task-4")
	})

	t.Run("returns empty slice for phase with no tasks", func(t *testing.T) {
		epicData := &epic.Epic{
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPending},
			},
		}

		tasks := taskService.GetTasksInPhase(epicData, "phase-2")
		assert.Len(t, tasks, 0)
	})
}

func TestTaskService_GetPendingTasksInPhase(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	taskService := NewTaskService(storage, queryService)

	t.Run("returns pending tasks only", func(t *testing.T) {
		epicData := &epic.Epic{
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPending},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusWIP},
				{ID: "task-3", PhaseID: "phase-1", Name: "Task 3", Status: epic.StatusCompleted},
				{ID: "task-4", PhaseID: "phase-1", Name: "Task 4", Status: epic.StatusCancelled},
			},
		}

		pendingTasks := taskService.GetPendingTasksInPhase(epicData, "phase-1")
		assert.Len(t, pendingTasks, 2) // task-1 (planning) and task-2 (active)

		taskIDs := make([]string, len(pendingTasks))
		for i, task := range pendingTasks {
			taskIDs[i] = task.ID
		}
		assert.Contains(t, taskIDs, "task-1")
		assert.Contains(t, taskIDs, "task-2")
	})

	t.Run("returns empty slice when all tasks completed/cancelled", func(t *testing.T) {
		epicData := &epic.Epic{
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusCancelled},
			},
		}

		pendingTasks := taskService.GetPendingTasksInPhase(epicData, "phase-1")
		assert.Len(t, pendingTasks, 0)
	})
}

func TestTaskService_AutomaticEventCreation(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	taskService := NewTaskService(storage, queryService)
	testTime := time.Date(2025, 8, 16, 15, 30, 0, 0, time.UTC)

	t.Run("automatic task_started event creation", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPending},
			},
			Events: []epic.Event{}, // Start with no events
		}

		// Start task
		err := taskService.StartTask(epicData, "task-1", testTime)
		require.NoError(t, err)

		// Verify event was automatically created
		require.Len(t, epicData.Events, 1)
		event := epicData.Events[0]

		assert.Equal(t, "task_started", event.Type)
		assert.Equal(t, "Task task-1 (Task 1) started", event.Data)
		assert.Equal(t, testTime, event.Timestamp)
		assert.NotEmpty(t, event.ID)
	})

	t.Run("automatic task_completed event creation", func(t *testing.T) {
		completedTime := time.Date(2025, 8, 16, 16, 30, 0, 0, time.UTC)
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusWIP, StartedAt: &testTime},
			},
			Events: []epic.Event{}, // Start with no events
		}

		// Complete task
		err := taskService.CompleteTask(epicData, "task-1", completedTime)
		require.NoError(t, err)

		// Verify event was automatically created
		require.Len(t, epicData.Events, 1)
		event := epicData.Events[0]

		assert.Equal(t, "task_completed", event.Type)
		assert.Equal(t, "Task task-1 (Task 1) completed", event.Data)
		assert.Equal(t, completedTime, event.Timestamp)
		assert.NotEmpty(t, event.ID)
	})

	t.Run("automatic task_cancelled event creation", func(t *testing.T) {
		cancelledTime := time.Date(2025, 8, 16, 16, 30, 0, 0, time.UTC)
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusWIP, StartedAt: &testTime},
			},
			Events: []epic.Event{}, // Start with no events
		}

		// Cancel task
		err := taskService.CancelTask(epicData, "task-1", cancelledTime)
		require.NoError(t, err)

		// Verify event was automatically created
		require.Len(t, epicData.Events, 1)
		event := epicData.Events[0]

		assert.Equal(t, "task_cancelled", event.Type)
		assert.Equal(t, "Task task-1 (Task 1) cancelled", event.Data)
		assert.Equal(t, cancelledTime, event.Timestamp)
		assert.NotEmpty(t, event.ID)
	})

	t.Run("events created for multiple task operations", func(t *testing.T) {
		completedTime := time.Date(2025, 8, 16, 16, 30, 0, 0, time.UTC)
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPending},
			},
			Events: []epic.Event{}, // Start with no events
		}

		// Start task
		err := taskService.StartTask(epicData, "task-1", testTime)
		require.NoError(t, err)

		// Complete task
		err = taskService.CompleteTask(epicData, "task-1", completedTime)
		require.NoError(t, err)

		// Verify both events were created
		require.Len(t, epicData.Events, 2)

		// Verify start event
		startEvent := epicData.Events[0]
		assert.Equal(t, "task_started", startEvent.Type)
		assert.Equal(t, "Task task-1 (Task 1) started", startEvent.Data)
		assert.Equal(t, testTime, startEvent.Timestamp)

		// Verify completion event
		completeEvent := epicData.Events[1]
		assert.Equal(t, "task_completed", completeEvent.Type)
		assert.Equal(t, "Task task-1 (Task 1) completed", completeEvent.Data)
		assert.Equal(t, completedTime, completeEvent.Timestamp)

		// Verify events have different IDs
		assert.NotEqual(t, startEvent.ID, completeEvent.ID)
	})
}

// Helper function to find task by ID
func findTaskByID(epicData *epic.Epic, taskID string) *epic.Task {
	for i := range epicData.Tasks {
		if epicData.Tasks[i].ID == taskID {
			return &epicData.Tasks[i]
		}
	}
	return nil
}
