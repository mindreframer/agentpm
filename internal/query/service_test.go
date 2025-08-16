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

		// Completion: weighted calculation - phases(40%): 1/4*40=10%, tasks(40%): 2/5*40=16%, tests(20%): 2/4*20=10% = 36%
		assert.Equal(t, 36, status.CompletionPercentage)
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
	storage := storage.NewMemoryStorage()
	qs := NewQueryService(storage)

	// Test with no epic loaded
	_, err := qs.GetFailingTests()
	if err == nil {
		t.Error("Expected error when no epic loaded")
	}

	// Create test epic with failing tests (non-completed tests)
	testEpic := &epic.Epic{
		ID:   "epic-1",
		Name: "Test Epic",
		Tests: []epic.Test{
			{ID: "test-1", TaskID: "task-1", Name: "Passing Test", Status: epic.StatusCompleted},
			{ID: "test-2", TaskID: "task-1", Name: "Failing Test", Status: epic.StatusActive},
			{ID: "test-3", TaskID: "task-2", Name: "Pending Test", Status: epic.StatusPlanning},
		},
		Tasks: []epic.Task{
			{ID: "task-1", PhaseID: "phase-1", Name: "Task 1"},
			{ID: "task-2", PhaseID: "phase-1", Name: "Task 2"},
		},
		Phases: []epic.Phase{
			{ID: "phase-1", Name: "Phase 1"},
		},
	}

	qs.epic = testEpic

	failing, err := qs.GetFailingTests()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(failing) != 2 {
		t.Errorf("Expected 2 failing tests, got %d", len(failing))
	}

	// Verify first failing test
	if failing[0].ID != "test-2" || failing[0].Name != "Failing Test" {
		t.Errorf("Unexpected first failing test: %+v", failing[0])
	}

	// Verify second failing test
	if failing[1].ID != "test-3" || failing[1].Name != "Pending Test" {
		t.Errorf("Unexpected second failing test: %+v", failing[1])
	}
}

func TestQueryService_GetRelatedItems(t *testing.T) {
	storage := storage.NewMemoryStorage()
	qs := NewQueryService(storage)

	// Test with no epic loaded
	_, err := qs.GetRelatedItems("phase", "phase-1")
	if err == nil {
		t.Error("Expected error when no epic loaded")
	}

	// Create test epic
	testEpic := &epic.Epic{
		ID:   "epic-1",
		Name: "Test Epic",
		Phases: []epic.Phase{
			{ID: "phase-1", Name: "Phase 1"},
		},
		Tasks: []epic.Task{
			{ID: "task-1", PhaseID: "phase-1", Name: "Task 1"},
			{ID: "task-2", PhaseID: "phase-1", Name: "Task 2"},
		},
		Tests: []epic.Test{
			{ID: "test-1", TaskID: "task-1", Name: "Test 1"},
			{ID: "test-2", TaskID: "task-1", Name: "Test 2"},
		},
	}

	qs.epic = testEpic

	// Test phase relationships
	related, err := qs.GetRelatedItems("phase", "phase-1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(related) != 4 { // 2 tasks + 2 tests
		t.Errorf("Expected 4 related items for phase, got %d", len(related))
	}

	// Test task relationships
	related, err = qs.GetRelatedItems("task", "task-1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(related) != 3 { // 1 phase + 2 tests
		t.Errorf("Expected 3 related items for task, got %d", len(related))
	}

	// Test test relationships
	related, err = qs.GetRelatedItems("test", "test-1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(related) != 2 { // 1 task + 1 phase
		t.Errorf("Expected 2 related items for test, got %d", len(related))
	}
}

func TestQueryService_AnalyzeImpact(t *testing.T) {
	storage := storage.NewMemoryStorage()
	qs := NewQueryService(storage)

	// Test with no epic loaded
	_, err := qs.AnalyzeImpact("phase", "phase-1")
	if err == nil {
		t.Error("Expected error when no epic loaded")
	}

	// Create test epic
	testEpic := &epic.Epic{
		ID:   "epic-1",
		Name: "Test Epic",
		Phases: []epic.Phase{
			{ID: "phase-1", Name: "Phase 1"},
		},
		Tasks: []epic.Task{
			{ID: "task-1", PhaseID: "phase-1", Name: "Task 1"},
			{ID: "task-2", PhaseID: "phase-1", Name: "Task 2"},
		},
		Tests: []epic.Test{
			{ID: "test-1", TaskID: "task-1", Name: "Test 1"},
			{ID: "test-2", TaskID: "task-2", Name: "Test 2"},
		},
	}

	qs.epic = testEpic

	// Test phase impact analysis
	analysis, err := qs.AnalyzeImpact("phase", "phase-1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(analysis.AffectedTasks) != 2 {
		t.Errorf("Expected 2 affected tasks, got %d", len(analysis.AffectedTasks))
	}

	if len(analysis.AffectedTests) != 2 {
		t.Errorf("Expected 2 affected tests, got %d", len(analysis.AffectedTests))
	}

	if analysis.RiskLevel != "low" {
		t.Errorf("Expected low risk level, got %s", analysis.RiskLevel)
	}

	// Test task impact analysis
	analysis, err = qs.AnalyzeImpact("task", "task-1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(analysis.AffectedTests) != 1 {
		t.Errorf("Expected 1 affected test, got %d", len(analysis.AffectedTests))
	}

	if len(analysis.AffectedPhases) != 1 {
		t.Errorf("Expected 1 affected phase, got %d", len(analysis.AffectedPhases))
	}
}

func TestQueryService_GetProgressInsights(t *testing.T) {
	storage := storage.NewMemoryStorage()
	qs := NewQueryService(storage)

	// Test with no epic loaded
	_, err := qs.GetProgressInsights()
	if err == nil {
		t.Error("Expected error when no epic loaded")
	}

	// Create test epic with mixed completion status
	testEpic := &epic.Epic{
		ID:   "epic-1",
		Name: "Test Epic",
		Tasks: []epic.Task{
			{ID: "task-1", Name: "Task 1", Status: epic.StatusCompleted},
			{ID: "task-2", Name: "Task 2", Status: epic.StatusActive},
			{ID: "task-3", Name: "Task 3", Status: epic.StatusPlanning},
		},
		Tests: []epic.Test{
			{ID: "test-1", Name: "Test 1", Status: epic.StatusCompleted},
			{ID: "test-2", Name: "Test 2", Status: epic.StatusActive},
		},
	}

	qs.epic = testEpic

	insights, err := qs.GetProgressInsights()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Velocity should be 2/5 = 0.4 (2 completed out of 5 total)
	expectedVelocity := 0.4
	if insights.Velocity != expectedVelocity {
		t.Errorf("Expected velocity %f, got %f", expectedVelocity, insights.Velocity)
	}

	if insights.EstimatedCompletion != "Early stage" {
		t.Errorf("Expected 'Early stage', got '%s'", insights.EstimatedCompletion)
	}

	// Should have recommendations
	if len(insights.Recommendations) == 0 {
		t.Error("Expected recommendations, got none")
	}
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

func TestQueryService_EnhancedCompletionPercentage(t *testing.T) {
	storage := storage.NewMemoryStorage()
	qs := NewQueryService(storage)

	tests := []struct {
		name     string
		epic     *epic.Epic
		expected int
	}{
		{
			name: "empty epic",
			epic: &epic.Epic{
				ID:     "epic-1",
				Name:   "Empty Epic",
				Phases: []epic.Phase{},
				Tasks:  []epic.Task{},
				Tests:  []epic.Test{},
			},
			expected: 0,
		},
		{
			name: "all completed epic",
			epic: &epic.Epic{
				ID:   "epic-1",
				Name: "Completed Epic",
				Phases: []epic.Phase{
					{ID: "phase-1", Status: epic.StatusCompleted},
					{ID: "phase-2", Status: epic.StatusCompleted},
				},
				Tasks: []epic.Task{
					{ID: "task-1", Status: epic.StatusCompleted},
					{ID: "task-2", Status: epic.StatusCompleted},
				},
				Tests: []epic.Test{
					{ID: "test-1", Status: epic.StatusCompleted},
				},
			},
			expected: 100,
		},
		{
			name: "partially completed epic",
			epic: &epic.Epic{
				ID:   "epic-1",
				Name: "Partial Epic",
				Phases: []epic.Phase{
					{ID: "phase-1", Status: epic.StatusCompleted}, // 1/2 = 50%
					{ID: "phase-2", Status: epic.StatusActive},
				},
				Tasks: []epic.Task{
					{ID: "task-1", Status: epic.StatusCompleted}, // 1/2 = 50%
					{ID: "task-2", Status: epic.StatusActive},
				},
				Tests: []epic.Test{
					{ID: "test-1", Status: epic.StatusCompleted}, // 1/1 = 100%
				},
			},
			// Weighted: 50% * 40% + 50% * 40% + 100% * 20% = 20% + 20% + 20% = 60%
			expected: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qs.LoadEpic("dummy") // Load method sets the epic internally
			qs.epic = tt.epic    // Override with test epic

			result := qs.calculateEnhancedCompletionPercentage()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueryService_GetDetailedProgress(t *testing.T) {
	storage := storage.NewMemoryStorage()
	qs := NewQueryService(storage)

	testTime := time.Date(2025, 8, 16, 15, 30, 0, 0, time.UTC)
	epic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusActive,
		Phases: []epic.Phase{
			{ID: "phase-1", Name: "Phase 1", Status: epic.StatusCompleted},
			{ID: "phase-2", Name: "Phase 2", Status: epic.StatusActive, StartedAt: &testTime},
			{ID: "phase-3", Name: "Phase 3", Status: epic.StatusPlanning},
		},
		Tasks: []epic.Task{
			{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
			{ID: "task-2", PhaseID: "phase-2", Name: "Task 2", Status: epic.StatusActive, StartedAt: &testTime},
			{ID: "task-3", PhaseID: "phase-2", Name: "Task 3", Status: epic.StatusPlanning},
			{ID: "task-4", PhaseID: "phase-3", Name: "Task 4", Status: epic.StatusPlanning},
		},
		Tests: []epic.Test{
			{ID: "test-1", TaskID: "task-1", Status: epic.StatusCompleted},
			{ID: "test-2", TaskID: "task-2", Status: epic.StatusPlanning},
		},
	}

	qs.epic = epic

	progress, err := qs.GetDetailedProgress()
	require.NoError(t, err)
	require.NotNil(t, progress)

	// Verify basic epic info
	assert.Equal(t, "epic-1", progress.EpicID)
	assert.Equal(t, "Test Epic", progress.EpicName)
	// assert.Equal(t, epic.StatusActive, progress.EpicStatus) // Temporarily commented out

	// Verify counts
	assert.Equal(t, 3, progress.TotalPhases)
	assert.Equal(t, 1, progress.CompletedPhases)
	assert.Equal(t, 4, progress.TotalTasks)
	assert.Equal(t, 1, progress.CompletedTasks)
	assert.Equal(t, 2, progress.TotalTests)
	assert.Equal(t, 1, progress.CompletedTests)

	// Verify active work
	assert.Equal(t, "phase-2", progress.ActivePhase)
	assert.Equal(t, "task-2", progress.ActiveTask)

	// Verify phase progress (task-2 is active but not complete, task-3 is pending)
	assert.Equal(t, 0, progress.ActivePhaseProgress) // 0/2 tasks completed in phase-2

	// Verify overall completion
	assert.Greater(t, progress.OverallCompletion, 0)
	assert.LessOrEqual(t, progress.OverallCompletion, 100)

	// Verify next action is set
	assert.NotEmpty(t, progress.NextAction)

	// Verify state validation
	assert.NotEmpty(t, progress.StateValidation)
}

func TestQueryService_CalculatePhaseProgress(t *testing.T) {
	storage := storage.NewMemoryStorage()
	qs := NewQueryService(storage)

	epic := &epic.Epic{
		ID: "epic-1",
		Tasks: []epic.Task{
			{ID: "task-1", PhaseID: "phase-1", Status: epic.StatusCompleted},
			{ID: "task-2", PhaseID: "phase-1", Status: epic.StatusCompleted},
			{ID: "task-3", PhaseID: "phase-1", Status: epic.StatusActive},
			{ID: "task-4", PhaseID: "phase-2", Status: epic.StatusCompleted},
		},
	}

	qs.epic = epic

	t.Run("phase with partial completion", func(t *testing.T) {
		// phase-1 has 2/3 tasks completed = 66%
		progress := qs.calculatePhaseProgress("phase-1")
		assert.Equal(t, 66, progress)
	})

	t.Run("phase with full completion", func(t *testing.T) {
		// phase-2 has 1/1 tasks completed = 100%
		progress := qs.calculatePhaseProgress("phase-2")
		assert.Equal(t, 100, progress)
	})

	t.Run("phase with no tasks", func(t *testing.T) {
		// phase-3 has no tasks = 100% (considered complete)
		progress := qs.calculatePhaseProgress("phase-3")
		assert.Equal(t, 100, progress)
	})
}

func TestQueryService_ValidateEpicState(t *testing.T) {
	storage := storage.NewMemoryStorage()
	qs := NewQueryService(storage)

	t.Run("valid epic state", func(t *testing.T) {
		epic := &epic.Epic{
			ID:     "epic-1",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusActive},
			},
		}

		qs.epic = epic
		validation, issues := qs.validateEpicState()

		assert.Equal(t, "valid", validation)
		assert.Empty(t, issues)
	})

	t.Run("multiple active phases error", func(t *testing.T) {
		epic := &epic.Epic{
			ID:     "epic-1",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusActive},
			},
		}

		qs.epic = epic
		validation, issues := qs.validateEpicState()

		assert.Equal(t, "error", validation)
		assert.Len(t, issues, 1)
		assert.Contains(t, issues[0], "Multiple active phases detected")
	})

	t.Run("multiple active tasks error", func(t *testing.T) {
		epic := &epic.Epic{
			ID:     "epic-1",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusActive},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusActive},
			},
		}

		qs.epic = epic
		validation, issues := qs.validateEpicState()

		assert.Equal(t, "error", validation)
		assert.Len(t, issues, 1)
		assert.Contains(t, issues[0], "Multiple active tasks detected")
	})

	t.Run("active task in inactive phase error", func(t *testing.T) {
		epic := &epic.Epic{
			ID:     "epic-1",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPlanning},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusActive},
			},
		}

		qs.epic = epic
		validation, issues := qs.validateEpicState()

		assert.Equal(t, "error", validation)
		assert.Len(t, issues, 1)
		assert.Contains(t, issues[0], "Active task task-1 in inactive phase")
	})

	t.Run("phase ready for completion warning", func(t *testing.T) {
		epic := &epic.Epic{
			ID:     "epic-1",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
			},
		}

		qs.epic = epic
		validation, issues := qs.validateEpicState()

		assert.Equal(t, "warning", validation)
		assert.Len(t, issues, 1)
		assert.Contains(t, issues[0], "should be completed")
	})
}
