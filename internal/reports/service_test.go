package reports

import (
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestEpicForReports() *epic.Epic {
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
			{ID: "TEST1", TaskID: "T1", Name: "Test 1", Status: epic.StatusCompleted, TestStatus: epic.TestStatusPassed},
			{ID: "TEST2", TaskID: "T2", Name: "Test 2", Status: epic.StatusCompleted, TestStatus: epic.TestStatusPassed},
			{ID: "TEST3", TaskID: "T3", Name: "Test 3", Status: epic.StatusPlanning, TestStatus: epic.TestStatusFailed},
			{ID: "TEST4", TaskID: "T4", Name: "Test 4", Status: epic.StatusPlanning, TestStatus: epic.TestStatusPending},
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
				Type:      "blocker",
				Timestamp: time.Date(2025, 8, 16, 11, 0, 0, 0, time.UTC),
				Data:      "Found critical dependency issue",
			},
		},
	}
}

func createCompletedEpicForReports() *epic.Epic {
	return &epic.Epic{
		ID:        "completed-epic",
		Name:      "Completed Epic",
		Status:    epic.StatusCompleted,
		CreatedAt: time.Date(2025, 8, 15, 9, 0, 0, 0, time.UTC),
		Assignee:  "test_agent",
		Phases: []epic.Phase{
			{ID: "P1", Name: "Phase 1", Status: epic.StatusCompleted},
			{ID: "P2", Name: "Phase 2", Status: epic.StatusCompleted},
		},
		Tasks: []epic.Task{
			{ID: "T1", PhaseID: "P1", Name: "Task 1", Status: epic.StatusCompleted},
			{ID: "T2", PhaseID: "P2", Name: "Task 2", Status: epic.StatusCompleted},
		},
		Tests: []epic.Test{
			{ID: "TEST1", TaskID: "T1", Name: "Test 1", Status: epic.StatusCompleted, TestStatus: epic.TestStatusPassed},
			{ID: "TEST2", TaskID: "T2", Name: "Test 2", Status: epic.StatusCompleted, TestStatus: epic.TestStatusPassed},
		},
		Events: []epic.Event{
			{
				ID:        "E1",
				Type:      "created",
				Timestamp: time.Date(2025, 8, 15, 9, 0, 0, 0, time.UTC),
				Data:      "Epic created",
			},
			{
				ID:        "E2",
				Type:      "epic_completed",
				Timestamp: time.Date(2025, 8, 16, 17, 0, 0, 0, time.UTC),
				Data:      "Epic completed successfully",
			},
		},
	}
}

func TestReportService_LoadEpic(t *testing.T) {
	t.Run("successful load", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createTestEpicForReports()
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		rs := NewReportService(storage)
		err = rs.LoadEpic("test.xml")
		require.NoError(t, err)
		assert.Equal(t, testEpic.ID, rs.epic.ID)
	})

	t.Run("load error", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		rs := NewReportService(storage)

		err := rs.LoadEpic("missing.xml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load epic")
	})
}

func TestReportService_GenerateHandoffReport(t *testing.T) {
	t.Run("comprehensive handoff report", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createTestEpicForReports()
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		rs := NewReportService(storage)
		err = rs.LoadEpic("test.xml")
		require.NoError(t, err)

		report, err := rs.GenerateHandoffReport(5)
		require.NoError(t, err)

		// Verify epic info
		assert.Equal(t, "test-epic-1", report.EpicInfo.ID)
		assert.Equal(t, "Test Epic", report.EpicInfo.Name)
		assert.Equal(t, "active", report.EpicInfo.Status)
		assert.Equal(t, "test_agent", report.EpicInfo.Assignee)

		// Verify current state
		assert.Equal(t, "P2", report.CurrentState.ActivePhase)
		assert.Equal(t, "T3", report.CurrentState.ActiveTask)
		assert.Contains(t, report.CurrentState.NextAction, "T4")

		// Verify summary
		assert.Equal(t, 1, report.Summary.CompletedPhases) // P1 completed
		assert.Equal(t, 4, report.Summary.TotalPhases)
		assert.Equal(t, 2, report.Summary.PassingTests) // TEST1, TEST2 passed
		assert.Equal(t, 2, report.Summary.FailingTests) // TEST3 failed, TEST4 pending

		// Verify events (should be in reverse chronological order)
		assert.Len(t, report.RecentEvents, 3)
		assert.Equal(t, "blocker", report.RecentEvents[0].Type)
		assert.Equal(t, "task_completed", report.RecentEvents[1].Type)

		// Verify blockers
		assert.Len(t, report.Blockers, 2) // 1 failed test + 1 blocker event
		assert.Contains(t, report.Blockers[0], "Failed test TEST3")
		assert.Contains(t, report.Blockers[1], "Found critical dependency issue")
	})

	t.Run("handoff report for completed epic", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createCompletedEpicForReports()
		err := storage.SaveEpic(testEpic, "completed.xml")
		require.NoError(t, err)

		rs := NewReportService(storage)
		err = rs.LoadEpic("completed.xml")
		require.NoError(t, err)

		report, err := rs.GenerateHandoffReport(5)
		require.NoError(t, err)

		// Verify completed epic state
		assert.Equal(t, "completed", report.EpicInfo.Status)
		assert.Equal(t, 2, report.Summary.CompletedPhases)
		assert.Equal(t, 2, report.Summary.TotalPhases)
		assert.Equal(t, 2, report.Summary.PassingTests)
		assert.Equal(t, 0, report.Summary.FailingTests)
		assert.Equal(t, "", report.CurrentState.ActivePhase) // No active phase
		assert.Equal(t, "", report.CurrentState.ActiveTask)  // No active task
		assert.Len(t, report.Blockers, 0)                    // No blockers
	})

	t.Run("recent events limit", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createTestEpicForReports()
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		rs := NewReportService(storage)
		err = rs.LoadEpic("test.xml")
		require.NoError(t, err)

		report, err := rs.GenerateHandoffReport(2) // Limit to 2 events
		require.NoError(t, err)

		assert.Len(t, report.RecentEvents, 2)
		assert.Equal(t, "blocker", report.RecentEvents[0].Type) // Most recent first
	})

	t.Run("no epic loaded", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		rs := NewReportService(storage)

		_, err := rs.GenerateHandoffReport(5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no epic loaded")
	})
}

func TestReportService_ProgressCalculation(t *testing.T) {
	t.Run("progress calculation accuracy", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createTestEpicForReports()
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		rs := NewReportService(storage)
		err = rs.LoadEpic("test.xml")
		require.NoError(t, err)

		completion := rs.calculateWeightedCompletion()

		// Expected calculation:
		// Phases: 1/4 completed = 25% * 40% weight = 10%
		// Tasks: 2/5 completed = 40% * 40% weight = 16%
		// Tests: 2/4 passed = 50% * 20% weight = 10%
		// Total: 10% + 16% + 10% = 36%
		assert.Equal(t, 36, completion)
	})
}

func TestReportService_CurrentStateExtraction(t *testing.T) {
	t.Run("current state extraction", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createTestEpicForReports()
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		rs := NewReportService(storage)
		err = rs.LoadEpic("test.xml")
		require.NoError(t, err)

		activePhase := rs.findActivePhase()
		activeTask := rs.findActiveTask()
		nextAction := rs.determineNextAction()

		assert.Equal(t, "P2", activePhase)
		assert.Equal(t, "T3", activeTask)
		assert.Contains(t, nextAction, "T4") // Next pending task
	})
}

func TestReportService_BlockerDetection(t *testing.T) {
	t.Run("basic blocker detection from failed tests", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createTestEpicForReports()
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		rs := NewReportService(storage)
		err = rs.LoadEpic("test.xml")
		require.NoError(t, err)

		blockers := rs.identifyBlockers()

		// Should find 1 failed test + 1 blocker event
		assert.Len(t, blockers, 2)
		assert.Contains(t, blockers[0], "Failed test TEST3")
		assert.Contains(t, blockers[1], "Found critical dependency issue")
	})

	t.Run("basic blocker detection from blocker events", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		// Create epic with only blocker events, no failed tests
		testEpic := &epic.Epic{
			ID:       "blocker-epic",
			Name:     "Blocker Epic",
			Status:   epic.StatusActive,
			Assignee: "test_agent",
			Tests:    []epic.Test{}, // No tests
			Events: []epic.Event{
				{
					ID:        "E1",
					Type:      "blocker",
					Timestamp: time.Date(2025, 8, 16, 9, 0, 0, 0, time.UTC),
					Data:      "Database connection issue",
				},
				{
					ID:        "E2",
					Type:      "blocker",
					Timestamp: time.Date(2025, 8, 16, 10, 0, 0, 0, time.UTC),
					Data:      "Missing API credentials",
				},
			},
		}
		err := storage.SaveEpic(testEpic, "blocker.xml")
		require.NoError(t, err)

		rs := NewReportService(storage)
		err = rs.LoadEpic("blocker.xml")
		require.NoError(t, err)

		blockers := rs.identifyBlockers()

		assert.Len(t, blockers, 2)
		assert.Contains(t, blockers[0], "Database connection issue")
		assert.Contains(t, blockers[1], "Missing API credentials")
	})
}

func TestReportService_EpicInfoExtraction(t *testing.T) {
	t.Run("epic info extraction from XML", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createTestEpicForReports()
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		rs := NewReportService(storage)
		err = rs.LoadEpic("test.xml")
		require.NoError(t, err)

		report, err := rs.GenerateHandoffReport(1)
		require.NoError(t, err)

		info := report.EpicInfo
		assert.Equal(t, "test-epic-1", info.ID)
		assert.Equal(t, "Test Epic", info.Name)
		assert.Equal(t, "active", info.Status)
		assert.Equal(t, "test_agent", info.Assignee)
		assert.Equal(t, time.Date(2025, 8, 16, 9, 0, 0, 0, time.UTC), info.Started)
	})
}
