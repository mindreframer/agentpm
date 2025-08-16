package phases

import (
	"testing"
	"time"

	"github.com/memomoo/agentpm/internal/epic"
	"github.com/memomoo/agentpm/internal/query"
	"github.com/memomoo/agentpm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhaseService_StartPhase(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	phaseService := NewPhaseService(storage, queryService)
	testTime := time.Date(2025, 8, 16, 15, 30, 0, 0, time.UTC)

	t.Run("start first phase of epic", func(t *testing.T) {
		// AC-1: Start first phase of epic
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPlanning},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPlanning},
			},
		}

		err := phaseService.StartPhase(epicData, "phase-1", testTime)
		require.NoError(t, err)

		// Verify phase status changed to active
		phase := findPhaseByID(epicData, "phase-1")
		require.NotNil(t, phase)
		assert.Equal(t, epic.StatusActive, phase.Status)
		assert.Equal(t, testTime, *phase.StartedAt)

		// Verify it's the active phase
		activePhase := phaseService.GetActivePhase(epicData)
		require.NotNil(t, activePhase)
		assert.Equal(t, "phase-1", activePhase.ID)
	})

	t.Run("prevent multiple active phases", func(t *testing.T) {
		// AC-2: Prevent multiple active phases
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPlanning},
			},
		}

		err := phaseService.StartPhase(epicData, "phase-2", testTime)
		require.Error(t, err)

		// Verify it's a constraint error
		var constraintErr *PhaseConstraintError
		require.ErrorAs(t, err, &constraintErr)
		assert.Equal(t, "phase-2", constraintErr.PhaseID)
		assert.Equal(t, "phase-1", constraintErr.ActivePhaseID)

		// Verify phase-2 is still pending
		phase := findPhaseByID(epicData, "phase-2")
		require.NotNil(t, phase)
		assert.Equal(t, epic.StatusPlanning, phase.Status)
	})

	t.Run("cannot start phase that is not pending", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusCompleted},
			},
		}

		err := phaseService.StartPhase(epicData, "phase-1", testTime)
		require.Error(t, err)

		// Verify it's a state error
		var stateErr *PhaseStateError
		require.ErrorAs(t, err, &stateErr)
		assert.Equal(t, "phase-1", stateErr.PhaseID)
		assert.Equal(t, epic.StatusCompleted, stateErr.CurrentStatus)
		assert.Equal(t, epic.StatusActive, stateErr.TargetStatus)
	})

	t.Run("cannot start non-existent phase", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{},
		}

		err := phaseService.StartPhase(epicData, "non-existent", testTime)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "phase non-existent not found")
	})
}

func TestPhaseService_CompletePhase(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	phaseService := NewPhaseService(storage, queryService)
	testTime := time.Date(2025, 8, 16, 16, 30, 0, 0, time.UTC)

	t.Run("complete phase with all tasks done", func(t *testing.T) {
		// AC-3: Complete phase with all tasks done
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusCompleted},
			},
		}

		err := phaseService.CompletePhase(epicData, "phase-1", testTime)
		require.NoError(t, err)

		// Verify phase status changed to completed
		phase := findPhaseByID(epicData, "phase-1")
		require.NotNil(t, phase)
		assert.Equal(t, epic.StatusCompleted, phase.Status)
		assert.Equal(t, testTime, *phase.CompletedAt)

		// Verify no active phase exists
		activePhase := phaseService.GetActivePhase(epicData)
		assert.Nil(t, activePhase)
	})

	t.Run("complete phase with cancelled tasks", func(t *testing.T) {
		// Cancelled tasks should allow phase completion
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusCancelled},
			},
		}

		err := phaseService.CompletePhase(epicData, "phase-1", testTime)
		require.NoError(t, err)

		// Verify phase completed successfully
		phase := findPhaseByID(epicData, "phase-1")
		require.NotNil(t, phase)
		assert.Equal(t, epic.StatusCompleted, phase.Status)
	})

	t.Run("prevent completing phase with pending tasks", func(t *testing.T) {
		// AC-4: Prevent completing phase with pending tasks
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
				{ID: "task-3", PhaseID: "phase-1", Name: "Task 3", Status: epic.StatusActive},
			},
		}

		err := phaseService.CompletePhase(epicData, "phase-1", testTime)
		require.Error(t, err)

		// Verify it's an incomplete error
		var incompleteErr *PhaseIncompleteError
		require.ErrorAs(t, err, &incompleteErr)
		assert.Equal(t, "phase-1", incompleteErr.PhaseID)
		assert.Len(t, incompleteErr.PendingTasks, 2) // task-2 (planning) and task-3 (active)

		// Verify phase is still active
		phase := findPhaseByID(epicData, "phase-1")
		require.NotNil(t, phase)
		assert.Equal(t, epic.StatusActive, phase.Status)
		assert.Nil(t, phase.CompletedAt)
	})

	t.Run("cannot complete phase that is not active", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPlanning},
			},
		}

		err := phaseService.CompletePhase(epicData, "phase-1", testTime)
		require.Error(t, err)

		// Verify it's a state error
		var stateErr *PhaseStateError
		require.ErrorAs(t, err, &stateErr)
		assert.Equal(t, "phase-1", stateErr.PhaseID)
		assert.Equal(t, epic.StatusPlanning, stateErr.CurrentStatus)
		assert.Equal(t, epic.StatusCompleted, stateErr.TargetStatus)
	})

	t.Run("cannot complete non-existent phase", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{},
		}

		err := phaseService.CompletePhase(epicData, "non-existent", testTime)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "phase non-existent not found")
	})
}

func TestPhaseService_GetActivePhase(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	phaseService := NewPhaseService(storage, queryService)

	t.Run("returns active phase when exists", func(t *testing.T) {
		epicData := &epic.Epic{
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPlanning},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusActive},
				{ID: "phase-3", Name: "Phase 3", Status: epic.StatusPlanning},
			},
		}

		activePhase := phaseService.GetActivePhase(epicData)
		require.NotNil(t, activePhase)
		assert.Equal(t, "phase-2", activePhase.ID)
		assert.Equal(t, epic.StatusActive, activePhase.Status)
	})

	t.Run("returns nil when no active phase", func(t *testing.T) {
		epicData := &epic.Epic{
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPlanning},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusCompleted},
			},
		}

		activePhase := phaseService.GetActivePhase(epicData)
		assert.Nil(t, activePhase)
	})

	t.Run("returns nil when no phases exist", func(t *testing.T) {
		epicData := &epic.Epic{
			Phases: []epic.Phase{},
		}

		activePhase := phaseService.GetActivePhase(epicData)
		assert.Nil(t, activePhase)
	})
}

func TestPhaseService_SingleActivePhaseConstraint(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	phaseService := NewPhaseService(storage, queryService)
	testTime := time.Date(2025, 8, 16, 15, 30, 0, 0, time.UTC)

	t.Run("only one phase can be active at a time", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPlanning},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPlanning},
				{ID: "phase-3", Name: "Phase 3", Status: epic.StatusPlanning},
			},
		}

		// Start first phase
		err := phaseService.StartPhase(epicData, "phase-1", testTime)
		require.NoError(t, err)

		// Verify phase-1 is active
		activePhase := phaseService.GetActivePhase(epicData)
		require.NotNil(t, activePhase)
		assert.Equal(t, "phase-1", activePhase.ID)

		// Try to start second phase - should fail
		err = phaseService.StartPhase(epicData, "phase-2", testTime)
		require.Error(t, err)
		var constraintErr *PhaseConstraintError
		require.ErrorAs(t, err, &constraintErr)

		// Verify phase-1 is still the only active phase
		activePhase = phaseService.GetActivePhase(epicData)
		require.NotNil(t, activePhase)
		assert.Equal(t, "phase-1", activePhase.ID)

		// Verify phase-2 is still pending
		phase2 := findPhaseByID(epicData, "phase-2")
		require.NotNil(t, phase2)
		assert.Equal(t, epic.StatusPlanning, phase2.Status)
	})
}

func TestPhaseService_AutomaticEventCreation(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	phaseService := NewPhaseService(storage, queryService)
	testTime := time.Date(2025, 8, 16, 15, 30, 0, 0, time.UTC)

	t.Run("automatic phase_started event creation", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPlanning},
			},
			Events: []epic.Event{}, // Start with no events
		}

		// Start phase
		err := phaseService.StartPhase(epicData, "phase-1", testTime)
		require.NoError(t, err)

		// Verify event was automatically created
		require.Len(t, epicData.Events, 1)
		event := epicData.Events[0]

		assert.Equal(t, "phase_started", event.Type)
		assert.Equal(t, "Phase 'Phase 1' started", event.Data)
		assert.Equal(t, testTime, event.Timestamp)
		assert.NotEmpty(t, event.ID)
	})

	t.Run("automatic phase_completed event creation", func(t *testing.T) {
		completedTime := time.Date(2025, 8, 16, 16, 30, 0, 0, time.UTC)
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive, StartedAt: &testTime},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
			},
			Events: []epic.Event{}, // Start with no events
		}

		// Complete phase
		err := phaseService.CompletePhase(epicData, "phase-1", completedTime)
		require.NoError(t, err)

		// Verify event was automatically created
		require.Len(t, epicData.Events, 1)
		event := epicData.Events[0]

		assert.Equal(t, "phase_completed", event.Type)
		assert.Equal(t, "Phase 'Phase 1' completed", event.Data)
		assert.Equal(t, completedTime, event.Timestamp)
		assert.NotEmpty(t, event.ID)
	})

	t.Run("events created for multiple phase operations", func(t *testing.T) {
		completedTime := time.Date(2025, 8, 16, 16, 30, 0, 0, time.UTC)
		epicData := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPlanning},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
			},
			Events: []epic.Event{}, // Start with no events
		}

		// Start phase
		err := phaseService.StartPhase(epicData, "phase-1", testTime)
		require.NoError(t, err)

		// Complete phase
		err = phaseService.CompletePhase(epicData, "phase-1", completedTime)
		require.NoError(t, err)

		// Verify both events were created
		require.Len(t, epicData.Events, 2)

		// Verify start event
		startEvent := epicData.Events[0]
		assert.Equal(t, "phase_started", startEvent.Type)
		assert.Equal(t, "Phase 'Phase 1' started", startEvent.Data)
		assert.Equal(t, testTime, startEvent.Timestamp)

		// Verify completion event
		completeEvent := epicData.Events[1]
		assert.Equal(t, "phase_completed", completeEvent.Type)
		assert.Equal(t, "Phase 'Phase 1' completed", completeEvent.Data)
		assert.Equal(t, completedTime, completeEvent.Timestamp)

		// Verify events have different IDs
		assert.NotEqual(t, startEvent.ID, completeEvent.ID)
	})
}

// Helper function to find phase by ID
func findPhaseByID(epicData *epic.Epic, phaseID string) *epic.Phase {
	for i := range epicData.Phases {
		if epicData.Phases[i].ID == phaseID {
			return &epicData.Phases[i]
		}
	}
	return nil
}
