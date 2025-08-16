package autonext

import (
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/phases"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/mindreframer/agentpm/internal/tasks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutoNextService_SelectNext(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	phaseService := phases.NewPhaseService(storage, queryService)
	taskService := tasks.NewTaskService(storage, queryService)
	autoNextService := NewAutoNextService(storage, queryService, phaseService, taskService)
	testTime := time.Date(2025, 8, 16, 15, 30, 0, 0, time.UTC)

	t.Run("auto-select next task in current phase", func(t *testing.T) {
		// AC-7: Auto-select next task in current phase
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusPlanning},
				{ID: "task-3", PhaseID: "phase-1", Name: "Task 3", Status: epic.StatusPlanning},
			},
		}

		result, err := autoNextService.SelectNext(epicData, testTime)
		require.NoError(t, err)

		assert.Equal(t, ActionStartTask, result.Action)
		assert.Equal(t, "phase-1", result.PhaseID)
		assert.Equal(t, "task-2", result.TaskID)
		assert.Equal(t, "Task 2", result.TaskName)
		assert.True(t, result.AutoSelected)
		assert.Contains(t, result.Message, "Started Task task-2: Task 2 (auto-selected)")

		// Verify task was started
		task := findTaskByID(epicData, "task-2")
		require.NotNil(t, task)
		assert.Equal(t, epic.StatusActive, task.Status)
	})

	t.Run("auto-select next task in next phase", func(t *testing.T) {
		// AC-8: Auto-select next task in next phase
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPlanning},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
				{ID: "task-2", PhaseID: "phase-2", Name: "Task 2", Status: epic.StatusPlanning},
				{ID: "task-3", PhaseID: "phase-2", Name: "Task 3", Status: epic.StatusPlanning},
			},
		}

		result, err := autoNextService.SelectNext(epicData, testTime)
		require.NoError(t, err)

		assert.Equal(t, ActionStartPhase, result.Action)
		assert.Equal(t, "phase-2", result.PhaseID) // Phase that was started
		assert.Equal(t, "task-2", result.TaskID)   // First task in new phase
		assert.Equal(t, "Phase 2", result.PhaseName)
		assert.Equal(t, "Task 2", result.TaskName)
		assert.True(t, result.AutoSelected)
		assert.Contains(t, result.Message, "Started Phase phase-2 and Task task-2 (auto-selected)")

		// Verify phase-1 was completed and phase-2 was started
		phase1 := findPhaseByID(epicData, "phase-1")
		phase2 := findPhaseByID(epicData, "phase-2")
		require.NotNil(t, phase1)
		require.NotNil(t, phase2)
		assert.Equal(t, epic.StatusCompleted, phase1.Status)
		assert.Equal(t, epic.StatusActive, phase2.Status)

		// Verify task was started
		task := findTaskByID(epicData, "task-2")
		require.NotNil(t, task)
		assert.Equal(t, epic.StatusActive, task.Status)
	})

	t.Run("handle completion of all work", func(t *testing.T) {
		// AC-9: Handle completion of all work
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusCompleted},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusCompleted},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
				{ID: "task-2", PhaseID: "phase-2", Name: "Task 2", Status: epic.StatusCompleted},
			},
		}

		result, err := autoNextService.SelectNext(epicData, testTime)
		require.NoError(t, err)

		assert.Equal(t, ActionCompleteEpic, result.Action)
		assert.Contains(t, result.Message, "All phases and tasks completed. Epic ready for completion.")
	})

	t.Run("start first phase when no active phase", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPlanning},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPlanning},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPlanning},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusPlanning},
			},
		}

		result, err := autoNextService.SelectNext(epicData, testTime)
		require.NoError(t, err)

		assert.Equal(t, ActionStartPhase, result.Action)
		assert.Equal(t, "phase-1", result.PhaseID)
		assert.Equal(t, "task-1", result.TaskID)
		assert.Equal(t, "Phase 1", result.PhaseName)
		assert.Equal(t, "Task 1", result.TaskName)
		assert.True(t, result.AutoSelected)
		assert.Contains(t, result.Message, "Started Phase phase-1 and Task task-1 (auto-selected)")

		// Verify phase was started and task was started
		phase := findPhaseByID(epicData, "phase-1")
		task := findTaskByID(epicData, "task-1")
		require.NotNil(t, phase)
		require.NotNil(t, task)
		assert.Equal(t, epic.StatusActive, phase.Status)
		assert.Equal(t, epic.StatusActive, task.Status)
	})

	t.Run("no action when task already active", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusActive},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusPlanning},
			},
		}

		result, err := autoNextService.SelectNext(epicData, testTime)
		require.NoError(t, err)

		assert.Equal(t, ActionNoWork, result.Action)
		assert.Contains(t, result.Message, "Task task-1 is already active in phase phase-1")

		// Verify no changes were made
		task1 := findTaskByID(epicData, "task-1")
		task2 := findTaskByID(epicData, "task-2")
		require.NotNil(t, task1)
		require.NotNil(t, task2)
		assert.Equal(t, epic.StatusActive, task1.Status)
		assert.Equal(t, epic.StatusPlanning, task2.Status)
	})

	t.Run("complete phase when all tasks done", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPlanning},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusCompleted},
				{ID: "task-3", PhaseID: "phase-2", Name: "Task 3", Status: epic.StatusPlanning},
			},
		}

		result, err := autoNextService.SelectNext(epicData, testTime)
		require.NoError(t, err)

		assert.Equal(t, ActionStartPhase, result.Action)
		assert.Equal(t, "phase-2", result.PhaseID) // The phase that was started
		assert.Equal(t, "task-3", result.TaskID)   // Task in next phase

		// Verify phase-1 was completed and phase-2 was started
		phase1 := findPhaseByID(epicData, "phase-1")
		phase2 := findPhaseByID(epicData, "phase-2")
		require.NotNil(t, phase1)
		require.NotNil(t, phase2)
		assert.Equal(t, epic.StatusCompleted, phase1.Status)
		assert.Equal(t, epic.StatusActive, phase2.Status)
	})

	t.Run("complete phase when all tasks done or cancelled", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPlanning},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusCancelled},
				{ID: "task-3", PhaseID: "phase-2", Name: "Task 3", Status: epic.StatusPlanning},
			},
		}

		result, err := autoNextService.SelectNext(epicData, testTime)
		require.NoError(t, err)

		assert.Equal(t, ActionStartPhase, result.Action)

		// Verify phase-1 was completed (cancelled tasks count as complete)
		phase1 := findPhaseByID(epicData, "phase-1")
		require.NotNil(t, phase1)
		assert.Equal(t, epic.StatusCompleted, phase1.Status)
	})

	t.Run("start phase with no tasks", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPlanning},
			},
			Tasks: []epic.Task{}, // No tasks in the phase
		}

		result, err := autoNextService.SelectNext(epicData, testTime)
		require.NoError(t, err)

		assert.Equal(t, ActionStartPhase, result.Action)
		assert.Equal(t, "phase-1", result.PhaseID)
		assert.Equal(t, "", result.TaskID) // No task to start
		assert.Equal(t, "Phase 1", result.PhaseName)
		assert.Contains(t, result.Message, "Started Phase phase-1 (no tasks available)")

		// Verify phase was started
		phase := findPhaseByID(epicData, "phase-1")
		require.NotNil(t, phase)
		assert.Equal(t, epic.StatusActive, phase.Status)
	})

	t.Run("priority order - prefer current phase", func(t *testing.T) {
		// Test that we prefer selecting tasks in current active phase over starting new phases
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPlanning},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusPlanning},
				{ID: "task-3", PhaseID: "phase-2", Name: "Task 3", Status: epic.StatusPlanning},
			},
		}

		result, err := autoNextService.SelectNext(epicData, testTime)
		require.NoError(t, err)

		// Should select task-2 in current active phase, not start phase-2
		assert.Equal(t, ActionStartTask, result.Action)
		assert.Equal(t, "phase-1", result.PhaseID)
		assert.Equal(t, "task-2", result.TaskID)

		// Verify phase-2 is still pending
		phase2 := findPhaseByID(epicData, "phase-2")
		require.NotNil(t, phase2)
		assert.Equal(t, epic.StatusPlanning, phase2.Status)
	})
}

func TestAutoNextService_XMLOutputFormats(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	phaseService := phases.NewPhaseService(storage, queryService)
	taskService := tasks.NewTaskService(storage, queryService)
	autoNextService := NewAutoNextService(storage, queryService, phaseService, taskService)
	testTime := time.Date(2025, 8, 16, 15, 30, 0, 0, time.UTC)

	t.Run("XML output for phase started", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-8",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "2A", Name: "Create PaginationComponent", Status: epic.StatusPlanning},
			},
			Tasks: []epic.Task{
				{ID: "2A_1", PhaseID: "2A", Name: "Create PaginationComponent with Previous/Next controls", Status: epic.StatusPlanning},
				{ID: "2A_2", PhaseID: "2A", Name: "Add accessibility features to pagination controls", Status: epic.StatusPlanning},
			},
		}

		result, err := autoNextService.SelectNext(epicData, testTime)
		require.NoError(t, err)

		assert.Equal(t, ActionStartPhase, result.Action)
		assert.NotEmpty(t, result.XMLOutput)

		// Verify XML structure matches specification
		assert.Contains(t, result.XMLOutput, `<phase_started epic="epic-8" phase="2A">`)
		assert.Contains(t, result.XMLOutput, `<phase_name>Create PaginationComponent</phase_name>`)
		assert.Contains(t, result.XMLOutput, `<previous_status>pending</previous_status>`)
		assert.Contains(t, result.XMLOutput, `<new_status>wip</new_status>`)
		assert.Contains(t, result.XMLOutput, `<started_at>2025-08-16T15:30:00Z</started_at>`)
		assert.Contains(t, result.XMLOutput, `<tasks>`)
		assert.Contains(t, result.XMLOutput, `<task id="2A_1" status="active">`)
		assert.Contains(t, result.XMLOutput, `<task id="2A_2" status="planning">`)
		assert.Contains(t, result.XMLOutput, `<started_task>2A_1</started_task>`)
		assert.Contains(t, result.XMLOutput, `<message>Started Phase 2A and Task 2A_1 (auto-selected)</message>`)
	})

	t.Run("XML output for all complete", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-8",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusCompleted},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
			},
		}

		result, err := autoNextService.SelectNext(epicData, testTime)
		require.NoError(t, err)

		assert.Equal(t, ActionCompleteEpic, result.Action)

		// Should match the all complete XML format from spec
		expectedXML := autoNextService.formatAllCompleteXML("epic-8")
		assert.Contains(t, expectedXML, `<all_complete epic="epic-8">`)
		assert.Contains(t, expectedXML, `<message>All phases and tasks completed. Epic ready for completion.</message>`)
		assert.Contains(t, expectedXML, `<suggestion>Use 'agentpm done-epic' to complete the epic</suggestion>`)
	})
}

func TestAutoNextService_EdgeCases(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	phaseService := phases.NewPhaseService(storage, queryService)
	taskService := tasks.NewTaskService(storage, queryService)
	autoNextService := NewAutoNextService(storage, queryService, phaseService, taskService)
	testTime := time.Date(2025, 8, 16, 15, 30, 0, 0, time.UTC)

	t.Run("epic with no phases", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{}, // No phases
			Tasks:  []epic.Task{},
		}

		result, err := autoNextService.SelectNext(epicData, testTime)
		require.NoError(t, err)

		assert.Equal(t, ActionCompleteEpic, result.Action)
		assert.Contains(t, result.Message, "All phases and tasks completed. Epic ready for completion.")
	})

	t.Run("phase with only cancelled tasks", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPlanning},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCancelled},
				{ID: "task-2", PhaseID: "phase-2", Name: "Task 2", Status: epic.StatusPlanning},
			},
		}

		result, err := autoNextService.SelectNext(epicData, testTime)
		require.NoError(t, err)

		// Should complete phase-1 and start phase-2
		assert.Equal(t, ActionStartPhase, result.Action)

		// Verify phase-1 was completed
		phase1 := findPhaseByID(epicData, "phase-1")
		require.NotNil(t, phase1)
		assert.Equal(t, epic.StatusCompleted, phase1.Status)
	})
}

// Helper functions

func findPhaseByID(epicData *epic.Epic, phaseID string) *epic.Phase {
	for i := range epicData.Phases {
		if epicData.Phases[i].ID == phaseID {
			return &epicData.Phases[i]
		}
	}
	return nil
}

func TestAutoNextService_AutomaticEventCreation(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	phaseService := phases.NewPhaseService(storage, queryService)
	taskService := tasks.NewTaskService(storage, queryService)
	autoNextService := NewAutoNextService(storage, queryService, phaseService, taskService)
	testTime := time.Date(2025, 8, 16, 15, 30, 0, 0, time.UTC)

	t.Run("auto-next creates events when starting phase and task", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPlanning},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPlanning},
			},
			Events: []epic.Event{}, // Start with no events
		}

		// Call SelectNext which should start phase and task
		result, err := autoNextService.SelectNext(epicData, testTime)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify the operation was a phase start (which also starts a task)
		assert.Equal(t, ActionStartPhase, result.Action)
		assert.Equal(t, "task-1", result.TaskID)
		assert.Equal(t, "phase-1", result.PhaseID)

		// Verify events were automatically created
		// Should have: phase_started and task_started events
		require.Len(t, epicData.Events, 2)

		// Verify phase started event
		phaseStartEvent := epicData.Events[0]
		assert.Equal(t, "phase_started", phaseStartEvent.Type)
		assert.Equal(t, "Phase 'Phase 1' started", phaseStartEvent.Data)
		assert.Equal(t, testTime, phaseStartEvent.Timestamp)

		// Verify task started event
		taskStartEvent := epicData.Events[1]
		assert.Equal(t, "task_started", taskStartEvent.Type)
		assert.Equal(t, "Task 'Task 1' started", taskStartEvent.Data)
		assert.Equal(t, testTime, taskStartEvent.Timestamp)

		// Verify events have different IDs
		assert.NotEqual(t, phaseStartEvent.ID, taskStartEvent.ID)
	})

	t.Run("auto-next creates events when completing phase", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive, StartedAt: &testTime},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPlanning},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
				{ID: "task-2", PhaseID: "phase-2", Name: "Task 2", Status: epic.StatusPlanning},
			},
			Events: []epic.Event{}, // Start with no events
		}

		// Call SelectNext which should complete phase-1 and start phase-2 + task-2
		result, err := autoNextService.SelectNext(epicData, testTime)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify the operation was a phase start (which also starts a task)
		assert.Equal(t, ActionStartPhase, result.Action)
		assert.Equal(t, "task-2", result.TaskID)
		assert.Equal(t, "phase-2", result.PhaseID)

		// Verify events were automatically created
		// Should have: phase_completed (phase-1), phase_started (phase-2), task_started (task-2)
		require.Len(t, epicData.Events, 3)

		// Verify phase completed event
		phaseCompleteEvent := epicData.Events[0]
		assert.Equal(t, "phase_completed", phaseCompleteEvent.Type)
		assert.Equal(t, "Phase 'Phase 1' completed", phaseCompleteEvent.Data)

		// Verify new phase started event
		phaseStartEvent := epicData.Events[1]
		assert.Equal(t, "phase_started", phaseStartEvent.Type)
		assert.Equal(t, "Phase 'Phase 2' started", phaseStartEvent.Data)

		// Verify task started event
		taskStartEvent := epicData.Events[2]
		assert.Equal(t, "task_started", taskStartEvent.Type)
		assert.Equal(t, "Task 'Task 2' started", taskStartEvent.Data)
	})
}

func findTaskByID(epicData *epic.Epic, taskID string) *epic.Task {
	for i := range epicData.Tasks {
		if epicData.Tasks[i].ID == taskID {
			return &epicData.Tasks[i]
		}
	}
	return nil
}
