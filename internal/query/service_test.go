package query

import (
	"testing"
	"time"

	"github.com/memomoo/agentpm/internal/epic"
	"github.com/memomoo/agentpm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestEpic() *epic.Epic {
	return &epic.Epic{
		ID:        "test-epic-1",
		Name:      "Test Epic",
		Status:    epic.StatusActive,
		CreatedAt: time.Date(2025, 8, 16, 9, 0, 0, 0, time.UTC),
		Assignee:  "test_agent",
		Phases: []epic.Phase{
			{ID: "P1", Name: "Phase 1", Status: epic.StatusCompleted},
			{ID: "P2", Name: "Phase 2", Status: epic.StatusActive},
			{ID: "P3", Name: "Phase 3", Status: epic.StatusPlanning},
			{ID: "P4", Name: "Phase 4", Status: epic.StatusPlanning},
		},
		Tasks: []epic.Task{
			{ID: "T1", PhaseID: "P1", Name: "Task 1", Status: epic.StatusCompleted},
			{ID: "T2", PhaseID: "P1", Name: "Task 2", Status: epic.StatusCompleted},
			{ID: "T3", PhaseID: "P2", Name: "Task 3", Status: epic.StatusActive},
			{ID: "T4", PhaseID: "P2", Name: "Task 4", Status: epic.StatusPlanning},
			{ID: "T5", PhaseID: "P3", Name: "Task 5", Status: epic.StatusPlanning},
		},
		Tests: []epic.Test{
			{ID: "TEST1", TaskID: "T1", Name: "Test 1", Status: epic.StatusCompleted},
			{ID: "TEST2", TaskID: "T2", Name: "Test 2", Status: epic.StatusCompleted},
			{ID: "TEST3", TaskID: "T3", Name: "Test 3", Status: epic.StatusPlanning}, // "failing"
			{ID: "TEST4", TaskID: "T4", Name: "Test 4", Status: epic.StatusPlanning}, // "failing"
		},
		Events: []epic.Event{
			{
				ID:        "E1",
				Type:      "created",
				Timestamp: time.Date(2025, 8, 16, 9, 0, 0, 0, time.UTC),
				Data:      "Epic created",
			},
			{
				ID:        "E2",
				Type:      "task_completed",
				Timestamp: time.Date(2025, 8, 16, 10, 0, 0, 0, time.UTC),
				Data:      "Task T1 completed",
			},
			{
				ID:        "E3",
				Type:      "phase_started",
				Timestamp: time.Date(2025, 8, 16, 11, 0, 0, 0, time.UTC),
				Data:      "Phase P2 started",
			},
		},
	}
}

func createCompletedEpic() *epic.Epic {
	return &epic.Epic{
		ID:     "completed-epic",
		Name:   "Completed Epic",
		Status: epic.StatusCompleted,
		Phases: []epic.Phase{
			{ID: "P1", Name: "Phase 1", Status: epic.StatusCompleted},
			{ID: "P2", Name: "Phase 2", Status: epic.StatusCompleted},
		},
		Tasks: []epic.Task{
			{ID: "T1", PhaseID: "P1", Name: "Task 1", Status: epic.StatusCompleted},
			{ID: "T2", PhaseID: "P2", Name: "Task 2", Status: epic.StatusCompleted},
		},
		Tests: []epic.Test{
			{ID: "TEST1", TaskID: "T1", Name: "Test 1", Status: epic.StatusCompleted},
			{ID: "TEST2", TaskID: "T2", Name: "Test 2", Status: epic.StatusCompleted},
		},
		Events: []epic.Event{},
	}
}

func TestQueryService_LoadEpic(t *testing.T) {
	t.Run("successful load", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createTestEpic()
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		qs := NewQueryService(storage)
		err = qs.LoadEpic("test.xml")
		require.NoError(t, err)
		assert.Equal(t, testEpic.ID, qs.epic.ID)
	})

	t.Run("load error", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		qs := NewQueryService(storage)

		err := qs.LoadEpic("missing.xml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load epic")
	})
}

func TestQueryService_GetEpicStatus(t *testing.T) {
	t.Run("epic with mixed completion", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createTestEpic()
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		qs := NewQueryService(storage)
		err = qs.LoadEpic("test.xml")
		require.NoError(t, err)

		status, err := qs.GetEpicStatus()
		require.NoError(t, err)

		assert.Equal(t, "test-epic-1", status.ID)
		assert.Equal(t, "Test Epic", status.Name)
		assert.Equal(t, epic.StatusActive, status.Status)
		assert.Equal(t, 1, status.CompletedPhases) // P1 is completed
		assert.Equal(t, 4, status.TotalPhases)
		assert.Equal(t, 2, status.PassingTests) // TEST1, TEST2
		assert.Equal(t, 2, status.FailingTests) // TEST3, TEST4
		assert.Equal(t, "P2", status.CurrentPhase)
		assert.Equal(t, "T3", status.CurrentTask)

		// Completion: (2 completed tasks + 2 completed tests) / (5 total tasks + 4 total tests) = 4/9 = 44%
		assert.Equal(t, 44, status.CompletionPercentage)
	})

	t.Run("completed epic", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createCompletedEpic()
		err := storage.SaveEpic(testEpic, "completed.xml")
		require.NoError(t, err)

		qs := NewQueryService(storage)
		err = qs.LoadEpic("completed.xml")
		require.NoError(t, err)

		status, err := qs.GetEpicStatus()
		require.NoError(t, err)

		assert.Equal(t, epic.StatusCompleted, status.Status)
		assert.Equal(t, 2, status.CompletedPhases)
		assert.Equal(t, 2, status.TotalPhases)
		assert.Equal(t, 2, status.PassingTests)
		assert.Equal(t, 0, status.FailingTests)
		assert.Equal(t, "", status.CurrentPhase) // no active phase
		assert.Equal(t, "", status.CurrentTask)  // no active task
		assert.Equal(t, 100, status.CompletionPercentage)
	})

	t.Run("no epic loaded", func(t *testing.T) {
		qs := NewQueryService(storage.NewMemoryStorage())

		_, err := qs.GetEpicStatus()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no epic loaded")
	})
}

func TestQueryService_GetCurrentState(t *testing.T) {
	t.Run("epic with active work", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createTestEpic()
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		qs := NewQueryService(storage)
		err = qs.LoadEpic("test.xml")
		require.NoError(t, err)

		state, err := qs.GetCurrentState()
		require.NoError(t, err)

		assert.Equal(t, epic.StatusActive, state.EpicStatus)
		assert.Equal(t, "P2", state.ActivePhase)
		assert.Equal(t, "T3", state.ActiveTask)
		assert.Equal(t, 2, state.FailingTests)
		assert.Contains(t, state.NextAction, "Fix failing tests")
	})

	t.Run("epic with no active work", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createCompletedEpic()
		err := storage.SaveEpic(testEpic, "completed.xml")
		require.NoError(t, err)

		qs := NewQueryService(storage)
		err = qs.LoadEpic("completed.xml")
		require.NoError(t, err)

		state, err := qs.GetCurrentState()
		require.NoError(t, err)

		assert.Equal(t, epic.StatusCompleted, state.EpicStatus)
		assert.Equal(t, "", state.ActivePhase)
		assert.Equal(t, "", state.ActiveTask)
		assert.Equal(t, 0, state.FailingTests)
		assert.Equal(t, "Epic ready for completion", state.NextAction)
	})
}

func TestQueryService_GetPendingWork(t *testing.T) {
	t.Run("epic with pending work", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createTestEpic()
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		qs := NewQueryService(storage)
		err = qs.LoadEpic("test.xml")
		require.NoError(t, err)

		pending, err := qs.GetPendingWork()
		require.NoError(t, err)

		// Check pending phases (P2 is active, P3 and P4 are planning/pending)
		assert.Len(t, pending.Phases, 3)            // P2, P3, P4 (all not completed)
		assert.Equal(t, "P2", pending.Phases[0].ID) // active = pending for this purpose
		assert.Equal(t, "P3", pending.Phases[1].ID)
		assert.Equal(t, "P4", pending.Phases[2].ID)

		// Check pending tasks (T3 is active, T4 and T5 are planning)
		assert.Len(t, pending.Tasks, 3)            // T3, T4, T5 (all not completed)
		assert.Equal(t, "T3", pending.Tasks[0].ID) // active = pending for this purpose
		assert.Equal(t, "T4", pending.Tasks[1].ID)
		assert.Equal(t, "T5", pending.Tasks[2].ID)

		// Check pending tests (TEST3, TEST4)
		assert.Len(t, pending.Tests, 2)
		assert.Equal(t, "TEST3", pending.Tests[0].ID)
		assert.Equal(t, "TEST4", pending.Tests[1].ID)
	})

	t.Run("completed epic with no pending work", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createCompletedEpic()
		err := storage.SaveEpic(testEpic, "completed.xml")
		require.NoError(t, err)

		qs := NewQueryService(storage)
		err = qs.LoadEpic("completed.xml")
		require.NoError(t, err)

		pending, err := qs.GetPendingWork()
		require.NoError(t, err)

		assert.Len(t, pending.Phases, 0)
		assert.Len(t, pending.Tasks, 0)
		assert.Len(t, pending.Tests, 0)
	})
}

func TestQueryService_GetFailingTests(t *testing.T) {
	t.Run("epic with failing tests", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createTestEpic()
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		qs := NewQueryService(storage)
		err = qs.LoadEpic("test.xml")
		require.NoError(t, err)

		failing, err := qs.GetFailingTests()
		require.NoError(t, err)

		assert.Len(t, failing, 2) // TEST3, TEST4
		assert.Equal(t, "TEST3", failing[0].ID)
		assert.Equal(t, "P2", failing[0].PhaseID) // T3 is in P2
		assert.Equal(t, "TEST4", failing[1].ID)
		assert.Equal(t, "P2", failing[1].PhaseID) // T4 is in P2
	})

	t.Run("epic with all passing tests", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createCompletedEpic()
		err := storage.SaveEpic(testEpic, "completed.xml")
		require.NoError(t, err)

		qs := NewQueryService(storage)
		err = qs.LoadEpic("completed.xml")
		require.NoError(t, err)

		failing, err := qs.GetFailingTests()
		require.NoError(t, err)

		assert.Len(t, failing, 0)
	})
}

func TestQueryService_GetRecentEvents(t *testing.T) {
	t.Run("get events with default limit", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createTestEpic()
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		qs := NewQueryService(storage)
		err = qs.LoadEpic("test.xml")
		require.NoError(t, err)

		events, err := qs.GetRecentEvents(0) // default limit
		require.NoError(t, err)

		assert.Len(t, events, 3)
		// Should be in reverse chronological order (most recent first)
		assert.Equal(t, "phase_started", events[0].Type) // most recent
		assert.Equal(t, "task_completed", events[1].Type)
		assert.Equal(t, "created", events[2].Type) // oldest
	})

	t.Run("get events with custom limit", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createTestEpic()
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		qs := NewQueryService(storage)
		err = qs.LoadEpic("test.xml")
		require.NoError(t, err)

		events, err := qs.GetRecentEvents(2)
		require.NoError(t, err)

		assert.Len(t, events, 2)
		assert.Equal(t, "phase_started", events[0].Type)  // most recent
		assert.Equal(t, "task_completed", events[1].Type) // second most recent
	})

	t.Run("limit exceeding max", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createTestEpic()
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		qs := NewQueryService(storage)
		err = qs.LoadEpic("test.xml")
		require.NoError(t, err)

		events, err := qs.GetRecentEvents(200) // exceeds max of 100
		require.NoError(t, err)

		assert.Len(t, events, 3) // all available events
	})
}

func TestQueryService_PhaseStatusDetermination(t *testing.T) {
	storage := storage.NewMemoryStorage()
	testEpic := createTestEpic()
	err := storage.SaveEpic(testEpic, "test.xml")
	require.NoError(t, err)

	qs := NewQueryService(storage)
	err = qs.LoadEpic("test.xml")
	require.NoError(t, err)

	t.Run("completed phase", func(t *testing.T) {
		status := qs.getPhaseStatus("P1")
		assert.Equal(t, epic.StatusCompleted, status)
	})

	t.Run("active phase", func(t *testing.T) {
		status := qs.getPhaseStatus("P2")
		assert.Equal(t, epic.StatusActive, status)
	})

	t.Run("planning phase", func(t *testing.T) {
		status := qs.getPhaseStatus("P3")
		assert.Equal(t, epic.StatusPlanning, status)
	})

	t.Run("empty phase", func(t *testing.T) {
		status := qs.getPhaseStatus("P_EMPTY")
		assert.Equal(t, epic.StatusPlanning, status)
	})
}

func TestQueryService_NextActionLogic(t *testing.T) {
	storage := storage.NewMemoryStorage()

	t.Run("fix failing tests priority", func(t *testing.T) {
		testEpic := createTestEpic()
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		qs := NewQueryService(storage)
		err = qs.LoadEpic("test.xml")
		require.NoError(t, err)

		action := qs.getNextAction()
		assert.Contains(t, action, "Fix failing tests")
	})

	t.Run("continue active task", func(t *testing.T) {
		testEpic := createTestEpic()
		// Remove failing tests
		testEpic.Tests = []epic.Test{
			{ID: "TEST1", TaskID: "T1", Name: "Test 1", Status: epic.StatusCompleted},
			{ID: "TEST2", TaskID: "T2", Name: "Test 2", Status: epic.StatusCompleted},
		}
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		qs := NewQueryService(storage)
		err = qs.LoadEpic("test.xml")
		require.NoError(t, err)

		action := qs.getNextAction()
		assert.Contains(t, action, "Continue work on: Task 3")
	})

	t.Run("start next task in active phase", func(t *testing.T) {
		testEpic := createTestEpic()
		// Remove failing tests and active task
		testEpic.Tests = []epic.Test{
			{ID: "TEST1", TaskID: "T1", Name: "Test 1", Status: epic.StatusCompleted},
			{ID: "TEST2", TaskID: "T2", Name: "Test 2", Status: epic.StatusCompleted},
		}
		testEpic.Tasks[2].Status = epic.StatusCompleted // T3 completed
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		qs := NewQueryService(storage)
		err = qs.LoadEpic("test.xml")
		require.NoError(t, err)

		action := qs.getNextAction()
		assert.Contains(t, action, "Start next task: Task 4")
	})

	t.Run("start next phase", func(t *testing.T) {
		testEpic := createTestEpic()
		// Complete all work in active phase
		testEpic.Tests = []epic.Test{
			{ID: "TEST1", TaskID: "T1", Name: "Test 1", Status: epic.StatusCompleted},
			{ID: "TEST2", TaskID: "T2", Name: "Test 2", Status: epic.StatusCompleted},
		}
		testEpic.Tasks[2].Status = epic.StatusCompleted  // T3 completed
		testEpic.Tasks[3].Status = epic.StatusCompleted  // T4 completed
		testEpic.Phases[1].Status = epic.StatusCompleted // P2 completed
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		qs := NewQueryService(storage)
		err = qs.LoadEpic("test.xml")
		require.NoError(t, err)

		action := qs.getNextAction()
		assert.Contains(t, action, "Start next phase: Phase 3")
	})

	t.Run("epic ready for completion", func(t *testing.T) {
		testEpic := createCompletedEpic()
		err := storage.SaveEpic(testEpic, "completed.xml")
		require.NoError(t, err)

		qs := NewQueryService(storage)
		err = qs.LoadEpic("completed.xml")
		require.NoError(t, err)

		action := qs.getNextAction()
		assert.Equal(t, "Epic ready for completion", action)
	})
}

func TestQueryService_EdgeCases(t *testing.T) {
	t.Run("epic loading and caching", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createTestEpic()
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		qs := NewQueryService(storage)

		// Load epic
		err = qs.LoadEpic("test.xml")
		require.NoError(t, err)

		// Verify epic is cached
		assert.NotNil(t, qs.epic)
		assert.Equal(t, testEpic.ID, qs.epic.ID)

		// Multiple operations should work with cached epic
		status, err := qs.GetEpicStatus()
		require.NoError(t, err)
		assert.Equal(t, testEpic.ID, status.ID)

		state, err := qs.GetCurrentState()
		require.NoError(t, err)
		assert.Equal(t, testEpic.Status, state.EpicStatus)
	})

	t.Run("operations without loaded epic", func(t *testing.T) {
		qs := NewQueryService(storage.NewMemoryStorage())

		_, err := qs.GetEpicStatus()
		assert.Error(t, err)

		_, err = qs.GetCurrentState()
		assert.Error(t, err)

		_, err = qs.GetPendingWork()
		assert.Error(t, err)

		_, err = qs.GetFailingTests()
		assert.Error(t, err)

		_, err = qs.GetRecentEvents(10)
		assert.Error(t, err)
	})
}
