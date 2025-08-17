package phases

import (
	"fmt"
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
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
		assert.Equal(t, "Phase phase-1 (Phase 1) started", event.Data)
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
		assert.Equal(t, "Phase phase-1 (Phase 1) completed", event.Data)
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
		assert.Equal(t, "Phase phase-1 (Phase 1) started", startEvent.Data)
		assert.Equal(t, testTime, startEvent.Timestamp)

		// Verify completion event
		completeEvent := epicData.Events[1]
		assert.Equal(t, "phase_completed", completeEvent.Type)
		assert.Equal(t, "Phase phase-1 (Phase 1) completed", completeEvent.Data)
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

// EPIC 9 PHASE 4A: Test Dependencies in Phase Management
func TestPhaseService_TestDependencyValidation(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	phaseService := NewPhaseService(storage, queryService)
	testTime := time.Date(2025, 8, 16, 15, 30, 0, 0, time.UTC)

	t.Run("Epic 9 Test: Phase completion blocked by incomplete tests", func(t *testing.T) {
		// Epic 9 line 72: Phase completion blocked by incomplete tests
		epicData := &epic.Epic{
			ID:     "epic-test-dependencies",
			Name:   "Test Dependencies Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
			},
			Tests: []epic.Test{
				{ID: "test-1", PhaseID: "phase-1", Name: "Test 1", Status: epic.StatusPending, TestStatus: epic.TestStatusPending},
				{ID: "test-2", PhaseID: "phase-1", Name: "Test 2", Status: epic.StatusActive, TestStatus: epic.TestStatusWIP},
			},
		}

		// Attempt to complete phase with incomplete tests
		err := phaseService.CompletePhase(epicData, "phase-1", testTime)
		require.Error(t, err)

		// Verify it's a test dependency error
		testDepError, ok := err.(*PhaseTestDependencyError)
		require.True(t, ok, "Expected PhaseTestDependencyError")
		assert.Equal(t, "phase-1", testDepError.PhaseID)
		assert.Len(t, testDepError.IncompleteTests, 2)
		assert.Equal(t, "test-1", testDepError.IncompleteTests[0].ID)
		assert.Equal(t, "test-2", testDepError.IncompleteTests[1].ID)
	})

	t.Run("Epic 9 Test: Phase starting blocked by incomplete prerequisite tests", func(t *testing.T) {
		// Epic 9 line 67: Phase starting blocked by incomplete tests
		epicData := &epic.Epic{
			ID:     "epic-prerequisite-tests",
			Name:   "Prerequisite Tests Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusCompleted},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPending},
			},
			Tests: []epic.Test{
				{ID: "test-1", PhaseID: "phase-1", Name: "Test 1", Status: epic.StatusActive, TestStatus: epic.TestStatusWIP},
				{ID: "test-2", PhaseID: "phase-1", Name: "Test 2", Status: epic.StatusCompleted, TestStatus: epic.TestStatusDone},
			},
		}

		// Attempt to start phase 2 with incomplete tests in phase 1
		err := phaseService.StartPhase(epicData, "phase-2", testTime)
		require.Error(t, err)

		// Verify it's a test prerequisite error
		prereqError, ok := err.(*PhaseTestPrerequisiteError)
		require.True(t, ok, "Expected PhaseTestPrerequisiteError")
		assert.Equal(t, "phase-2", prereqError.PhaseID)
		assert.Len(t, prereqError.PrerequisiteTests, 1)
		assert.Equal(t, "test-1", prereqError.PrerequisiteTests[0].ID)
	})

	t.Run("Epic 9 Test: Test completion affects phase lifecycle", func(t *testing.T) {
		// Epic 9 line 68: Test completion affects phase lifecycle
		epicData := &epic.Epic{
			ID:     "epic-lifecycle-tests",
			Name:   "Lifecycle Tests Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
			},
			Tests: []epic.Test{
				{ID: "test-1", PhaseID: "phase-1", Name: "Test 1", Status: epic.StatusCompleted, TestStatus: epic.TestStatusDone},
				{ID: "test-2", PhaseID: "phase-1", Name: "Test 2", Status: epic.StatusCompleted, TestStatus: epic.TestStatusDone},
			},
		}

		// Phase completion should succeed with all tests passed
		err := phaseService.CompletePhase(epicData, "phase-1", testTime)
		require.NoError(t, err)

		// Verify phase was completed
		phase := findPhaseByID(epicData, "phase-1")
		require.NotNil(t, phase)
		assert.Equal(t, epic.StatusCompleted, phase.Status)
		assert.Equal(t, testTime, *phase.CompletedAt)
	})

	t.Run("Epic 9 Test: Clear messaging about blocking tests", func(t *testing.T) {
		// Epic 9 line 74: Clear messaging about blocking tests
		epicData := &epic.Epic{
			ID:     "epic-clear-messaging",
			Name:   "Clear Messaging Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tests: []epic.Test{
				{ID: "test-1", PhaseID: "phase-1", Name: "Critical Test", Status: epic.StatusActive, TestStatus: epic.TestStatusWIP},
				{ID: "test-2", PhaseID: "phase-1", Name: "Integration Test", Status: epic.StatusPending, TestStatus: epic.TestStatusPending},
			},
		}

		// Attempt to complete phase with clear test names
		err := phaseService.CompletePhase(epicData, "phase-1", testTime)
		require.Error(t, err)

		// Verify error message clarity
		assert.Contains(t, err.Error(), "phase-1")
		assert.Contains(t, err.Error(), "incomplete tests")
		assert.Contains(t, err.Error(), "2")

		// Verify specific test details are accessible
		testDepError := err.(*PhaseTestDependencyError)
		assert.Equal(t, "Critical Test", testDepError.IncompleteTests[0].Name)
		assert.Equal(t, "Integration Test", testDepError.IncompleteTests[1].Name)
	})

	t.Run("Epic 9 Test: Dependency validation rule enforcement", func(t *testing.T) {
		// Test comprehensive dependency validation rules
		epicData := &epic.Epic{
			ID:     "epic-validation-rules",
			Name:   "Validation Rules Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusCompleted},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusCompleted},
				{ID: "phase-3", Name: "Phase 3", Status: epic.StatusPending},
			},
			Tests: []epic.Test{
				// Phase 1 tests - mixed completion
				{ID: "test-1-1", PhaseID: "phase-1", Name: "Test 1.1", Status: epic.StatusCompleted, TestStatus: epic.TestStatusDone},
				{ID: "test-1-2", PhaseID: "phase-1", Name: "Test 1.2", Status: epic.StatusCancelled, TestStatus: epic.TestStatusCancelled},
				// Phase 2 tests - one failed
				{ID: "test-2-1", PhaseID: "phase-2", Name: "Test 2.1", Status: epic.StatusCompleted, TestStatus: epic.TestStatusDone},
				{ID: "test-2-2", PhaseID: "phase-2", Name: "Test 2.2", Status: epic.StatusActive, TestStatus: epic.TestStatusWIP},
			},
		}

		// Attempt to start phase 3 with failed test in phase 2
		err := phaseService.StartPhase(epicData, "phase-3", testTime)
		require.Error(t, err)

		// Verify only failed tests are reported as incomplete
		prereqError := err.(*PhaseTestPrerequisiteError)
		assert.Len(t, prereqError.PrerequisiteTests, 1)
		assert.Equal(t, "test-2-2", prereqError.PrerequisiteTests[0].ID)
		assert.Equal(t, epic.TestStatusWIP, prereqError.PrerequisiteTests[0].TestStatus)
	})

	t.Run("Epic 9 Test: Backwards compatibility with existing epics", func(t *testing.T) {
		// Test backwards compatibility with epics that don't have TestStatus
		epicData := &epic.Epic{
			ID:     "epic-backwards-compat",
			Name:   "Backwards Compatibility Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tests: []epic.Test{
				// Legacy test without TestStatus - only has Status
				{ID: "test-legacy", PhaseID: "phase-1", Name: "Legacy Test", Status: epic.StatusCompleted},
				// Modern test with both Status and TestStatus
				{ID: "test-modern", PhaseID: "phase-1", Name: "Modern Test", Status: epic.StatusCompleted, TestStatus: epic.TestStatusDone},
			},
		}

		// Phase completion should succeed with legacy tests
		err := phaseService.CompletePhase(epicData, "phase-1", testTime)
		require.NoError(t, err)

		// Verify phase was completed despite mixed test formats
		phase := findPhaseByID(epicData, "phase-1")
		require.NotNil(t, phase)
		assert.Equal(t, epic.StatusCompleted, phase.Status)
	})
}

// EPIC 9 PHASE 4A: Test Completion Status Tracking
func TestPhaseService_TestCompletionStatus(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	phaseService := NewPhaseService(storage, queryService)

	t.Run("Epic 9 Test: Test completion status tracking", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-status-tracking",
			Name:   "Status Tracking Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPending},
			},
			Tests: []epic.Test{
				// Phase 1 tests
				{ID: "test-1-1", PhaseID: "phase-1", Name: "Test 1.1", Status: epic.StatusCompleted, TestStatus: epic.TestStatusDone},
				{ID: "test-1-2", PhaseID: "phase-1", Name: "Test 1.2", Status: epic.StatusActive, TestStatus: epic.TestStatusWIP},
				{ID: "test-1-3", PhaseID: "phase-1", Name: "Test 1.3", Status: epic.StatusPending, TestStatus: epic.TestStatusPending},
				// Phase 2 tests
				{ID: "test-2-1", PhaseID: "phase-2", Name: "Test 2.1", Status: epic.StatusCompleted, TestStatus: epic.TestStatusDone},
			},
		}

		// Test individual phase status
		phase1Status := phaseService.GetTestCompletionStatus(epicData, "phase-1")
		assert.Equal(t, "phase-1", phase1Status.PhaseID)
		assert.Equal(t, 3, phase1Status.TotalTests)
		assert.Equal(t, 1, phase1Status.PassedTests)
		assert.Equal(t, 1, phase1Status.FailedTests)
		assert.Equal(t, 1, phase1Status.PendingTests)
		assert.Len(t, phase1Status.IncompleteTests, 2)
		assert.False(t, phase1Status.AllTestsCompleted)

		// Test overall status
		overallStatus := phaseService.GetOverallTestCompletionStatus(epicData)
		assert.Len(t, overallStatus, 2)
		assert.Equal(t, 3, overallStatus["phase-1"].TotalTests)
		assert.Equal(t, 1, overallStatus["phase-2"].TotalTests)
		assert.True(t, overallStatus["phase-2"].AllTestsCompleted)
	})
}

// EPIC 9 PHASE 4B: Comprehensive Test Dependency Edge Cases
func TestPhaseService_TestDependencyEdgeCases(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	phaseService := NewPhaseService(storage, queryService)
	testTime := time.Date(2025, 8, 16, 15, 30, 0, 0, time.UTC)

	t.Run("Epic 9 Test: Phase with no tests should complete successfully", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-no-tests",
			Name:   "Epic With No Tests",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-empty", Name: "Empty Phase", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-empty", Name: "Task 1", Status: epic.StatusCompleted},
			},
			Tests: []epic.Test{}, // No tests
		}

		// Phase completion should succeed with no tests
		err := phaseService.CompletePhase(epicData, "phase-empty", testTime)
		require.NoError(t, err)

		// Verify phase was completed
		phase := findPhaseByID(epicData, "phase-empty")
		require.NotNil(t, phase)
		assert.Equal(t, epic.StatusCompleted, phase.Status)
	})

	t.Run("Epic 9 Test: Phase with only cancelled tests should complete successfully", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-cancelled-tests",
			Name:   "Epic With Cancelled Tests",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-cancelled", Name: "Cancelled Tests Phase", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-cancelled", Name: "Task 1", Status: epic.StatusCompleted},
			},
			Tests: []epic.Test{
				{ID: "test-1", PhaseID: "phase-cancelled", Name: "Cancelled Test 1", Status: epic.StatusCancelled, TestStatus: epic.TestStatusCancelled},
				{ID: "test-2", PhaseID: "phase-cancelled", Name: "Cancelled Test 2", Status: epic.StatusCancelled, TestStatus: epic.TestStatusCancelled},
			},
		}

		// Phase completion should succeed with only cancelled tests
		err := phaseService.CompletePhase(epicData, "phase-cancelled", testTime)
		require.NoError(t, err)

		// Verify phase was completed
		phase := findPhaseByID(epicData, "phase-cancelled")
		require.NotNil(t, phase)
		assert.Equal(t, epic.StatusCompleted, phase.Status)
	})

	t.Run("Epic 9 Test: Mixed test statuses with complex dependency chain", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-complex-dependencies",
			Name:   "Complex Dependencies Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusCompleted},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusCompleted},
				{ID: "phase-3", Name: "Phase 3", Status: epic.StatusPending},
			},
			Tests: []epic.Test{
				// Phase 1: All completed (passed and cancelled)
				{ID: "test-1-1", PhaseID: "phase-1", Name: "Test 1.1", Status: epic.StatusCompleted, TestStatus: epic.TestStatusDone},
				{ID: "test-1-2", PhaseID: "phase-1", Name: "Test 1.2", Status: epic.StatusCancelled, TestStatus: epic.TestStatusCancelled},
				// Phase 2: Mixed with one failed
				{ID: "test-2-1", PhaseID: "phase-2", Name: "Test 2.1", Status: epic.StatusCompleted, TestStatus: epic.TestStatusDone},
				{ID: "test-2-2", PhaseID: "phase-2", Name: "Test 2.2", Status: epic.StatusActive, TestStatus: epic.TestStatusWIP},
				// Phase 3: Not started
				{ID: "test-3-1", PhaseID: "phase-3", Name: "Test 3.1", Status: epic.StatusPending, TestStatus: epic.TestStatusPending},
			},
		}

		// Phase 3 start should be blocked by failed test in phase 2
		err := phaseService.StartPhase(epicData, "phase-3", testTime)
		require.Error(t, err)

		// Verify it's a prerequisite error with correct test details
		prereqError, ok := err.(*PhaseTestPrerequisiteError)
		require.True(t, ok, "Expected PhaseTestPrerequisiteError")
		assert.Equal(t, "phase-3", prereqError.PhaseID)
		assert.Len(t, prereqError.PrerequisiteTests, 1)
		assert.Equal(t, "test-2-2", prereqError.PrerequisiteTests[0].ID)
		assert.Equal(t, epic.TestStatusWIP, prereqError.PrerequisiteTests[0].TestStatus)
	})

	t.Run("Epic 9 Test: Large scale test dependency validation performance", func(t *testing.T) {
		// Create epic with many phases and tests
		phases := make([]epic.Phase, 10)
		tests := make([]epic.Test, 50)

		for i := 0; i < 10; i++ {
			phases[i] = epic.Phase{
				ID:     fmt.Sprintf("phase-%d", i+1),
				Name:   fmt.Sprintf("Phase %d", i+1),
				Status: epic.StatusCompleted,
			}
		}

		// Set last phase as pending
		phases[9].Status = epic.StatusPending

		// Create 5 tests per phase
		for i := 0; i < 50; i++ {
			phaseIndex := i / 5
			testIndex := i % 5
			tests[i] = epic.Test{
				ID:         fmt.Sprintf("test-%d-%d", phaseIndex+1, testIndex+1),
				PhaseID:    fmt.Sprintf("phase-%d", phaseIndex+1),
				Name:       fmt.Sprintf("Test %d.%d", phaseIndex+1, testIndex+1),
				Status:     epic.StatusCompleted,
				TestStatus: epic.TestStatusDone,
			}
		}

		// Make one test in phase 9 failed to block phase 10 start
		tests[44].Status = epic.StatusActive // test-9-5
		tests[44].TestStatus = epic.TestStatusWIP

		epicData := &epic.Epic{
			ID:     "epic-large-scale",
			Name:   "Large Scale Epic",
			Status: epic.StatusActive,
			Phases: phases,
			Tests:  tests,
		}

		// Phase 10 start should be efficiently blocked by failed test in phase 9
		start := time.Now()
		err := phaseService.StartPhase(epicData, "phase-10", testTime)
		duration := time.Since(start)

		require.Error(t, err)
		// Performance check: should complete in reasonable time (< 100ms)
		assert.Less(t, duration.Milliseconds(), int64(100), "Test dependency validation should be fast")

		// Verify correct error detection
		prereqError, ok := err.(*PhaseTestPrerequisiteError)
		require.True(t, ok)
		assert.Len(t, prereqError.PrerequisiteTests, 1)
		assert.Equal(t, "test-9-5", prereqError.PrerequisiteTests[0].ID)
	})

	t.Run("Epic 9 Test: Zero tests edge case", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-zero-tests",
			Name:   "Zero Tests Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPending},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPending},
			},
			Tests: []epic.Test{}, // No tests at all
		}

		// Phase 2 start should succeed with no prerequisite tests
		err := phaseService.StartPhase(epicData, "phase-2", testTime)
		require.NoError(t, err)

		// Verify phase was started
		phase := findPhaseByID(epicData, "phase-2")
		require.NotNil(t, phase)
		assert.Equal(t, epic.StatusActive, phase.Status)
	})
}

// EPIC 9 PHASE 4B: Error Message and Hint Validation Tests
func TestPhaseService_ErrorMessageClarity(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	phaseService := NewPhaseService(storage, queryService)
	testTime := time.Date(2025, 8, 16, 15, 30, 0, 0, time.UTC)

	t.Run("Epic 9 Test: Phase completion error messages contain specific test details", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-error-clarity",
			Name:   "Error Clarity Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "delivery-phase", Name: "Delivery Phase", Status: epic.StatusActive},
			},
			Tests: []epic.Test{
				{ID: "unit-test-auth", PhaseID: "delivery-phase", Name: "Unit Test - Authentication", Status: epic.StatusActive, TestStatus: epic.TestStatusWIP},
				{ID: "integration-test-api", PhaseID: "delivery-phase", Name: "Integration Test - API", Status: epic.StatusPending, TestStatus: epic.TestStatusPending},
				{ID: "e2e-test-workflow", PhaseID: "delivery-phase", Name: "E2E Test - Workflow", Status: epic.StatusActive, TestStatus: epic.TestStatusWIP},
			},
		}

		err := phaseService.CompletePhase(epicData, "delivery-phase", testTime)
		require.Error(t, err)

		// Verify error message contains phase ID and test count
		assert.Contains(t, err.Error(), "delivery-phase")
		assert.Contains(t, err.Error(), "3")
		assert.Contains(t, err.Error(), "incomplete tests")

		// Verify specific test details are accessible
		testDepError, ok := err.(*PhaseTestDependencyError)
		require.True(t, ok)
		assert.Equal(t, "delivery-phase", testDepError.PhaseID)
		assert.Len(t, testDepError.IncompleteTests, 3)

		// Verify all incomplete tests are captured with meaningful names
		testNames := make([]string, len(testDepError.IncompleteTests))
		for i, test := range testDepError.IncompleteTests {
			testNames[i] = test.Name
		}
		assert.Contains(t, testNames, "Unit Test - Authentication")
		assert.Contains(t, testNames, "Integration Test - API")
		assert.Contains(t, testNames, "E2E Test - Workflow")
	})

	t.Run("Epic 9 Test: Phase prerequisite error messages are specific and actionable", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-prereq-clarity",
			Name:   "Prerequisite Clarity Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "foundation", Name: "Foundation Phase", Status: epic.StatusCompleted},
				{ID: "implementation", Name: "Implementation Phase", Status: epic.StatusCompleted},
				{ID: "deployment", Name: "Deployment Phase", Status: epic.StatusPending},
			},
			Tests: []epic.Test{
				// Foundation phase tests - all good
				{ID: "test-foundation-1", PhaseID: "foundation", Name: "Foundation Unit Tests", Status: epic.StatusCompleted, TestStatus: epic.TestStatusDone},
				// Implementation phase tests - one failed
				{ID: "test-impl-integration", PhaseID: "implementation", Name: "Implementation Integration Test", Status: epic.StatusActive, TestStatus: epic.TestStatusWIP},
				{ID: "test-impl-performance", PhaseID: "implementation", Name: "Implementation Performance Test", Status: epic.StatusCompleted, TestStatus: epic.TestStatusDone},
			},
		}

		err := phaseService.StartPhase(epicData, "deployment", testTime)
		require.Error(t, err)

		// Verify error message contains target phase and prerequisite test count
		assert.Contains(t, err.Error(), "deployment")
		assert.Contains(t, err.Error(), "1")
		assert.Contains(t, err.Error(), "prerequisite tests")

		// Verify specific prerequisite test details
		prereqError, ok := err.(*PhaseTestPrerequisiteError)
		require.True(t, ok)
		assert.Equal(t, "deployment", prereqError.PhaseID)
		assert.Len(t, prereqError.PrerequisiteTests, 1)
		assert.Equal(t, "test-impl-integration", prereqError.PrerequisiteTests[0].ID)
		assert.Equal(t, "Implementation Integration Test", prereqError.PrerequisiteTests[0].Name)
		assert.Equal(t, epic.TestStatusWIP, prereqError.PrerequisiteTests[0].TestStatus)
	})

	t.Run("Epic 9 Test: Error messages distinguish between different test failure types", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-failure-types",
			Name:   "Failure Types Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "testing-phase", Name: "Testing Phase", Status: epic.StatusActive},
			},
			Tests: []epic.Test{
				{ID: "test-failed", PhaseID: "testing-phase", Name: "Failed Test", Status: epic.StatusActive, TestStatus: epic.TestStatusWIP},
				{ID: "test-pending", PhaseID: "testing-phase", Name: "Pending Test", Status: epic.StatusPending, TestStatus: epic.TestStatusPending},
				{ID: "test-wip", PhaseID: "testing-phase", Name: "WIP Test", Status: epic.StatusActive, TestStatus: epic.TestStatusWIP},
			},
		}

		err := phaseService.CompletePhase(epicData, "testing-phase", testTime)
		require.Error(t, err)

		testDepError, ok := err.(*PhaseTestDependencyError)
		require.True(t, ok)

		// Verify all different failure types are captured
		assert.Len(t, testDepError.IncompleteTests, 3)

		statusTypes := make(map[epic.TestStatus]int)
		for _, test := range testDepError.IncompleteTests {
			statusTypes[test.TestStatus]++
		}

		assert.Equal(t, 2, statusTypes[epic.TestStatusWIP])
		assert.Equal(t, 1, statusTypes[epic.TestStatusPending])
	})

	t.Run("Epic 9 Test: Test completion status provides comprehensive breakdown", func(t *testing.T) {
		epicData := &epic.Epic{
			ID:     "epic-status-breakdown",
			Name:   "Status Breakdown Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "comprehensive-phase", Name: "Comprehensive Phase", Status: epic.StatusActive},
			},
			Tests: []epic.Test{
				{ID: "test-passed-1", PhaseID: "comprehensive-phase", Name: "Passed Test 1", Status: epic.StatusCompleted, TestStatus: epic.TestStatusDone},
				{ID: "test-passed-2", PhaseID: "comprehensive-phase", Name: "Passed Test 2", Status: epic.StatusCompleted, TestStatus: epic.TestStatusDone},
				{ID: "test-failed-1", PhaseID: "comprehensive-phase", Name: "Failed Test 1", Status: epic.StatusActive, TestStatus: epic.TestStatusWIP},
				{ID: "test-pending-1", PhaseID: "comprehensive-phase", Name: "Pending Test 1", Status: epic.StatusPending, TestStatus: epic.TestStatusPending},
				{ID: "test-pending-2", PhaseID: "comprehensive-phase", Name: "Pending Test 2", Status: epic.StatusPending, TestStatus: epic.TestStatusPending},
				{ID: "test-cancelled", PhaseID: "comprehensive-phase", Name: "Cancelled Test", Status: epic.StatusCancelled, TestStatus: epic.TestStatusCancelled},
			},
		}

		status := phaseService.GetTestCompletionStatus(epicData, "comprehensive-phase")

		// Verify comprehensive status breakdown
		assert.Equal(t, "comprehensive-phase", status.PhaseID)
		assert.Equal(t, 6, status.TotalTests)
		assert.Equal(t, 2, status.PassedTests)   // 2 passed
		assert.Equal(t, 1, status.FailedTests)   // 1 failed
		assert.Equal(t, 2, status.PendingTests)  // 2 pending
		assert.Len(t, status.IncompleteTests, 3) // failed + pending (cancelled tests are considered complete)
		assert.False(t, status.AllTestsCompleted)

		// Verify incomplete tests don't include cancelled tests
		incompleteIDs := make([]string, len(status.IncompleteTests))
		for i, test := range status.IncompleteTests {
			incompleteIDs[i] = test.ID
		}
		assert.NotContains(t, incompleteIDs, "test-cancelled")
		assert.Contains(t, incompleteIDs, "test-failed-1")
		assert.Contains(t, incompleteIDs, "test-pending-1")
		assert.Contains(t, incompleteIDs, "test-pending-2")
	})
}
