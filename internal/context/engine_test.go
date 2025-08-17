package context

import (
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
)

func TestEngine_GetTaskContext(t *testing.T) {
	tests := []struct {
		name               string
		taskID             string
		includeFullDetails bool
		wantErr            bool
		wantTaskName       string
		wantParentPhase    bool
		wantSiblingTasks   int
		wantChildTests     int
	}{
		{
			name:               "get task with full context",
			taskID:             "1A_1",
			includeFullDetails: true,
			wantErr:            false,
			wantTaskName:       "Initialize Project",
			wantParentPhase:    true,
			wantSiblingTasks:   2,
			wantChildTests:     2,
		},
		{
			name:               "get task without full context",
			taskID:             "1A_1",
			includeFullDetails: false,
			wantErr:            false,
			wantTaskName:       "Initialize Project",
			wantParentPhase:    false,
			wantSiblingTasks:   0,
			wantChildTests:     0,
		},
		{
			name:               "get non-existent task",
			taskID:             "NONEXISTENT",
			includeFullDetails: true,
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := createTestEngine(t)

			ctx, err := engine.GetTaskContext(tt.taskID, tt.includeFullDetails)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetTaskContext() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetTaskContext() unexpected error = %v", err)
				return
			}

			if ctx.TaskDetails.Name != tt.wantTaskName {
				t.Errorf("GetTaskContext() task name = %v, want %v", ctx.TaskDetails.Name, tt.wantTaskName)
			}

			if tt.wantParentPhase && ctx.ParentPhase == nil {
				t.Errorf("GetTaskContext() expected parent phase, got nil")
			} else if !tt.wantParentPhase && ctx.ParentPhase != nil {
				t.Errorf("GetTaskContext() expected no parent phase, got %v", ctx.ParentPhase.ID)
			}

			if len(ctx.SiblingTasks) != tt.wantSiblingTasks {
				t.Errorf("GetTaskContext() sibling tasks = %v, want %v", len(ctx.SiblingTasks), tt.wantSiblingTasks)
			}

			if len(ctx.ChildTests) != tt.wantChildTests {
				t.Errorf("GetTaskContext() child tests = %v, want %v", len(ctx.ChildTests), tt.wantChildTests)
			}
		})
	}
}

func TestEngine_GetPhaseContext(t *testing.T) {
	tests := []struct {
		name               string
		phaseID            string
		includeFullDetails bool
		wantErr            bool
		wantPhaseName      string
		wantAllTasks       int
		wantSiblingPhases  int
		wantProgress       bool
	}{
		{
			name:               "get phase with full context",
			phaseID:            "1A",
			includeFullDetails: true,
			wantErr:            false,
			wantPhaseName:      "CLI Framework & Core Structure",
			wantAllTasks:       3,
			wantSiblingPhases:  1,
			wantProgress:       true,
		},
		{
			name:               "get phase without full context",
			phaseID:            "1A",
			includeFullDetails: false,
			wantErr:            false,
			wantPhaseName:      "CLI Framework & Core Structure",
			wantAllTasks:       0,
			wantSiblingPhases:  0,
			wantProgress:       false,
		},
		{
			name:               "get non-existent phase",
			phaseID:            "NONEXISTENT",
			includeFullDetails: true,
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := createTestEngine(t)

			ctx, err := engine.GetPhaseContext(tt.phaseID, tt.includeFullDetails)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetPhaseContext() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetPhaseContext() unexpected error = %v", err)
				return
			}

			if ctx.PhaseDetails.Name != tt.wantPhaseName {
				t.Errorf("GetPhaseContext() phase name = %v, want %v", ctx.PhaseDetails.Name, tt.wantPhaseName)
			}

			if len(ctx.AllTasks) != tt.wantAllTasks {
				t.Errorf("GetPhaseContext() all tasks = %v, want %v", len(ctx.AllTasks), tt.wantAllTasks)
			}

			if len(ctx.SiblingPhases) != tt.wantSiblingPhases {
				t.Errorf("GetPhaseContext() sibling phases = %v, want %v", len(ctx.SiblingPhases), tt.wantSiblingPhases)
			}

			if tt.wantProgress && ctx.ProgressSummary == nil {
				t.Errorf("GetPhaseContext() expected progress summary, got nil")
			} else if !tt.wantProgress && ctx.ProgressSummary != nil {
				t.Errorf("GetPhaseContext() expected no progress summary, got %v", ctx.ProgressSummary)
			}
		})
	}
}

func TestEngine_GetTestContext(t *testing.T) {
	tests := []struct {
		name               string
		testID             string
		includeFullDetails bool
		wantErr            bool
		wantTestName       string
		wantParentTask     bool
		wantParentPhase    bool
		wantSiblingTests   int
	}{
		{
			name:               "get test with full context",
			testID:             "T1A_1",
			includeFullDetails: true,
			wantErr:            false,
			wantTestName:       "Test Project Init",
			wantParentTask:     true,
			wantParentPhase:    true,
			wantSiblingTests:   1,
		},
		{
			name:               "get test without full context",
			testID:             "T1A_1",
			includeFullDetails: false,
			wantErr:            false,
			wantTestName:       "Test Project Init",
			wantParentTask:     false,
			wantParentPhase:    false,
			wantSiblingTests:   0,
		},
		{
			name:               "get non-existent test",
			testID:             "NONEXISTENT",
			includeFullDetails: true,
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := createTestEngine(t)

			ctx, err := engine.GetTestContext(tt.testID, tt.includeFullDetails)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetTestContext() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetTestContext() unexpected error = %v", err)
				return
			}

			if ctx.TestDetails.Name != tt.wantTestName {
				t.Errorf("GetTestContext() test name = %v, want %v", ctx.TestDetails.Name, tt.wantTestName)
			}

			if tt.wantParentTask && ctx.ParentTask == nil {
				t.Errorf("GetTestContext() expected parent task, got nil")
			} else if !tt.wantParentTask && ctx.ParentTask != nil {
				t.Errorf("GetTestContext() expected no parent task, got %v", ctx.ParentTask.ID)
			}

			if tt.wantParentPhase && ctx.ParentPhase == nil {
				t.Errorf("GetTestContext() expected parent phase, got nil")
			} else if !tt.wantParentPhase && ctx.ParentPhase != nil {
				t.Errorf("GetTestContext() expected no parent phase, got %v", ctx.ParentPhase.ID)
			}

			if len(ctx.SiblingTests) != tt.wantSiblingTests {
				t.Errorf("GetTestContext() sibling tests = %v, want %v", len(ctx.SiblingTests), tt.wantSiblingTests)
			}
		})
	}
}

func TestEngine_calculatePhaseProgress(t *testing.T) {
	tests := []struct {
		name                     string
		phaseID                  string
		wantTotalTasks           int
		wantCompletedTasks       int
		wantActiveTasks          int
		wantPendingTasks         int
		wantCompletionPercentage int
		wantTotalTests           int
		wantPassedTests          int
	}{
		{
			name:                     "phase with mixed task states",
			phaseID:                  "1A",
			wantTotalTasks:           3,
			wantCompletedTasks:       1,
			wantActiveTasks:          0,
			wantPendingTasks:         2,
			wantCompletionPercentage: 33, // 1/3 * 100 = 33%
			wantTotalTests:           4,
			wantPassedTests:          1,
		},
		{
			name:                     "phase with no tasks",
			phaseID:                  "1C",
			wantTotalTasks:           0,
			wantCompletedTasks:       0,
			wantActiveTasks:          0,
			wantPendingTasks:         0,
			wantCompletionPercentage: 0,
			wantTotalTests:           0,
			wantPassedTests:          0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := createTestEngine(t)

			progress := engine.calculatePhaseProgress(tt.phaseID)

			if progress.TotalTasks != tt.wantTotalTasks {
				t.Errorf("calculatePhaseProgress() total tasks = %v, want %v", progress.TotalTasks, tt.wantTotalTasks)
			}

			if progress.CompletedTasks != tt.wantCompletedTasks {
				t.Errorf("calculatePhaseProgress() completed tasks = %v, want %v", progress.CompletedTasks, tt.wantCompletedTasks)
			}

			if progress.ActiveTasks != tt.wantActiveTasks {
				t.Errorf("calculatePhaseProgress() active tasks = %v, want %v", progress.ActiveTasks, tt.wantActiveTasks)
			}

			if progress.PendingTasks != tt.wantPendingTasks {
				t.Errorf("calculatePhaseProgress() pending tasks = %v, want %v", progress.PendingTasks, tt.wantPendingTasks)
			}

			if progress.CompletionPercentage != tt.wantCompletionPercentage {
				t.Errorf("calculatePhaseProgress() completion percentage = %v, want %v", progress.CompletionPercentage, tt.wantCompletionPercentage)
			}

			if progress.TotalTests != tt.wantTotalTests {
				t.Errorf("calculatePhaseProgress() total tests = %v, want %v", progress.TotalTests, tt.wantTotalTests)
			}

			if progress.PassedTests != tt.wantPassedTests {
				t.Errorf("calculatePhaseProgress() passed tests = %v, want %v", progress.PassedTests, tt.wantPassedTests)
			}
		})
	}
}

func TestEngine_getSiblingTasks(t *testing.T) {
	tests := []struct {
		name               string
		taskID             string
		phaseID            string
		includeFullDetails bool
		wantSiblingCount   int
		wantFirstSibling   string
	}{
		{
			name:               "get siblings with full details",
			taskID:             "1A_1",
			phaseID:            "1A",
			includeFullDetails: true,
			wantSiblingCount:   2,
			wantFirstSibling:   "1A_2",
		},
		{
			name:               "get siblings without full details",
			taskID:             "1A_1",
			phaseID:            "1A",
			includeFullDetails: false,
			wantSiblingCount:   2,
			wantFirstSibling:   "1A_2",
		},
		{
			name:               "task with no siblings",
			taskID:             "1B_1",
			phaseID:            "1B",
			includeFullDetails: true,
			wantSiblingCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := createTestEngine(t)

			siblings := engine.getSiblingTasks(tt.taskID, tt.phaseID, tt.includeFullDetails)

			if len(siblings) != tt.wantSiblingCount {
				t.Errorf("getSiblingTasks() count = %v, want %v", len(siblings), tt.wantSiblingCount)
			}

			if tt.wantSiblingCount > 0 && siblings[0].ID != tt.wantFirstSibling {
				t.Errorf("getSiblingTasks() first sibling = %v, want %v", siblings[0].ID, tt.wantFirstSibling)
			}

			// Check that full details are included when requested
			if tt.includeFullDetails && tt.wantSiblingCount > 0 {
				if siblings[0].Description == "" {
					t.Errorf("getSiblingTasks() with full details should include description")
				}
			}
		})
	}
}

func TestEngine_getChildTests(t *testing.T) {
	tests := []struct {
		name               string
		taskID             string
		includeFullDetails bool
		wantTestCount      int
		wantFirstTest      string
	}{
		{
			name:               "get child tests with full details",
			taskID:             "1A_1",
			includeFullDetails: true,
			wantTestCount:      2,
			wantFirstTest:      "T1A_1",
		},
		{
			name:               "get child tests without full details",
			taskID:             "1A_1",
			includeFullDetails: false,
			wantTestCount:      2,
			wantFirstTest:      "T1A_1",
		},
		{
			name:               "task with no tests",
			taskID:             "1B_1",
			includeFullDetails: true,
			wantTestCount:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := createTestEngine(t)

			tests := engine.getChildTests(tt.taskID, tt.includeFullDetails)

			if len(tests) != tt.wantTestCount {
				t.Errorf("getChildTests() count = %v, want %v", len(tests), tt.wantTestCount)
			}

			if tt.wantTestCount > 0 && tests[0].ID != tt.wantFirstTest {
				t.Errorf("getChildTests() first test = %v, want %v", tests[0].ID, tt.wantFirstTest)
			}

			// Check that full details are included when requested
			if tt.includeFullDetails && tt.wantTestCount > 0 {
				if tests[0].Description == "" {
					t.Errorf("getChildTests() with full details should include description")
				}
			}
		})
	}
}

// createTestEngine creates a test engine with sample data
func createTestEngine(t *testing.T) *Engine {
	// Create an in-memory storage with test data
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)

	// Create test epic with sample data
	testEpic := createTestEpic()

	// Save the epic to storage
	err := storage.SaveEpic(testEpic, "test-epic.xml")
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Load the epic into the query service
	err = queryService.LoadEpic("test-epic.xml")
	if err != nil {
		t.Fatalf("Failed to load test epic: %v", err)
	}

	return NewEngine(queryService)
}

// createTestEpic creates a sample epic for testing
func createTestEpic() *epic.Epic {
	now := time.Now()
	earlier := now.Add(-time.Hour)

	testEpic := &epic.Epic{
		ID:     "test-epic-1",
		Name:   "Test Epic",
		Status: epic.StatusActive,
		Phases: []epic.Phase{
			{
				ID:           "1A",
				Name:         "CLI Framework & Core Structure",
				Description:  "Setup basic CLI structure and initialize project",
				Deliverables: "Functional CLI framework, Project structure established, Core dependencies configured",
				Status:       epic.StatusActive,
				StartedAt:    &earlier,
			},
			{
				ID:          "1B",
				Name:        "Command Implementation",
				Description: "Implement core CLI commands and functionality",
				Status:      epic.StatusPlanning,
			},
		},
		Tasks: []epic.Task{
			{
				ID:                 "1A_1",
				PhaseID:            "1A",
				Name:               "Initialize Project",
				Description:        "Initialize Go module with required dependencies",
				AcceptanceCriteria: "Go module initializes successfully, Required dependencies are added to go.mod, Project structure follows Go conventions",
				Status:             epic.StatusCompleted,
				Assignee:           "agent_claude",
				StartedAt:          &earlier,
				CompletedAt:        &now,
			},
			{
				ID:                 "1A_2",
				PhaseID:            "1A",
				Name:               "Configure Tools",
				Description:        "Set up development tools and linting configuration",
				AcceptanceCriteria: "golangci-lint configured, pre-commit hooks set up, IDE configuration provided",
				Status:             epic.StatusPlanning,
			},
			{
				ID:                 "1A_3",
				PhaseID:            "1A",
				Name:               "Setup Testing Framework",
				Description:        "Initialize testing framework with basic test structure",
				AcceptanceCriteria: "Test framework configured, Example tests created, Test coverage reporting enabled",
				Status:             epic.StatusPlanning,
			},
			{
				ID:          "1B_1",
				PhaseID:     "1B",
				Name:        "Implement Show Command",
				Description: "Implement the show command for displaying entity information",
				Status:      epic.StatusPlanning,
			},
		},
		Tests: []epic.Test{
			{
				ID:          "T1A_1",
				TaskID:      "1A_1",
				PhaseID:     "1A",
				Name:        "Test Project Init",
				Description: "Verify that project initializes correctly with all dependencies",
				Status:      epic.StatusCompleted,
				TestStatus:  epic.TestStatusPassed,
				StartedAt:   &earlier,
				PassedAt:    &now,
			},
			{
				ID:          "T1A_2",
				TaskID:      "1A_1",
				PhaseID:     "1A",
				Name:        "Test Dependency Resolution",
				Description: "Verify all required dependencies are properly resolved",
				Status:      epic.StatusPlanning,
				TestStatus:  epic.TestStatusPending,
			},
			{
				ID:          "T1A_3",
				TaskID:      "1A_2",
				PhaseID:     "1A",
				Name:        "Test Linting Configuration",
				Description: "Verify linting rules are properly configured",
				Status:      epic.StatusPlanning,
				TestStatus:  epic.TestStatusPending,
			},
			{
				ID:          "T1A_4",
				TaskID:      "1A_3",
				PhaseID:     "1A",
				Name:        "Test Framework Validation",
				Description: "Verify testing framework is properly configured",
				Status:      epic.StatusPlanning,
				TestStatus:  epic.TestStatusPending,
			},
		},
	}

	return testEpic
}
