package context

import (
	"fmt"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/query"
)

// ContextRetriever interface defines the contract for retrieving hierarchical context
type ContextRetriever interface {
	GetTaskContext(taskID string, includeFullDetails bool) (*TaskContext, error)
	GetPhaseContext(phaseID string, includeFullDetails bool) (*PhaseContext, error)
	GetTestContext(testID string, includeFullDetails bool) (*TestContext, error)
}

// Engine implements the ContextRetriever interface using a query service
type Engine struct {
	queryService *query.QueryService
}

// NewEngine creates a new context engine with the provided query service
func NewEngine(queryService *query.QueryService) *Engine {
	return &Engine{
		queryService: queryService,
	}
}

// GetTaskContext retrieves comprehensive context for a task
func (e *Engine) GetTaskContext(taskID string, includeFullDetails bool) (*TaskContext, error) {
	task, err := e.queryService.GetTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	context := &TaskContext{
		TaskDetails: TaskDetails{
			ID:                 task.ID,
			PhaseID:            task.PhaseID,
			Name:               task.Name,
			Description:        task.Description,
			AcceptanceCriteria: task.AcceptanceCriteria,
			Status:             task.Status,
			Assignee:           task.Assignee,
			StartedAt:          task.StartedAt,
			CompletedAt:        task.CompletedAt,
		},
	}

	if includeFullDetails {
		// Get parent phase with full details
		if task.PhaseID != "" {
			parentPhase, err := e.queryService.GetPhase(task.PhaseID)
			if err == nil {
				context.ParentPhase = &PhaseDetails{
					ID:           parentPhase.ID,
					Name:         parentPhase.Name,
					Description:  parentPhase.Description,
					Deliverables: parentPhase.Deliverables,
					Status:       parentPhase.Status,
					StartedAt:    parentPhase.StartedAt,
					CompletedAt:  parentPhase.CompletedAt,
				}

				// Calculate progress for parent phase
				context.ParentPhase.Progress = e.calculatePhaseProgress(task.PhaseID)
			}
		}

		// Get sibling tasks in the same phase
		if task.PhaseID != "" {
			context.SiblingTasks = e.getSiblingTasks(taskID, task.PhaseID, includeFullDetails)
		}

		// Get child tests for this task
		context.ChildTests = e.getChildTests(taskID, includeFullDetails)
	}

	return context, nil
}

// GetPhaseContext retrieves comprehensive context for a phase
func (e *Engine) GetPhaseContext(phaseID string, includeFullDetails bool) (*PhaseContext, error) {
	phase, err := e.queryService.GetPhase(phaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get phase: %w", err)
	}

	context := &PhaseContext{
		PhaseDetails: PhaseDetails{
			ID:           phase.ID,
			Name:         phase.Name,
			Description:  phase.Description,
			Deliverables: phase.Deliverables,
			Status:       phase.Status,
			StartedAt:    phase.StartedAt,
			CompletedAt:  phase.CompletedAt,
		},
	}

	if includeFullDetails {
		// Calculate progress summary
		context.ProgressSummary = e.calculatePhaseProgress(phaseID)

		// Get all tasks in this phase
		context.AllTasks = e.getTasksInPhase(phaseID, includeFullDetails)

		// Get phase-level tests (tests associated with phase but not with any specific task)
		context.PhaseTests = e.getPhaseTests(phaseID, includeFullDetails)

		// Get sibling phases
		context.SiblingPhases = e.getSiblingPhases(phaseID, includeFullDetails)
	}

	return context, nil
}

// GetTestContext retrieves comprehensive context for a test
func (e *Engine) GetTestContext(testID string, includeFullDetails bool) (*TestContext, error) {
	test, err := e.queryService.GetTest(testID)
	if err != nil {
		return nil, fmt.Errorf("failed to get test: %w", err)
	}

	context := &TestContext{
		TestDetails: TestDetails{
			ID:          test.ID,
			TaskID:      test.TaskID,
			PhaseID:     test.PhaseID,
			Name:        test.Name,
			Description: test.Description,
			Status:      test.Status,
			TestStatus:  test.TestStatus,
			StartedAt:   test.StartedAt,
			PassedAt:    test.PassedAt,
			FailedAt:    test.FailedAt,
			FailureNote: test.FailureNote,
		},
	}

	if includeFullDetails {
		// Get parent task with full details
		if test.TaskID != "" {
			parentTask, err := e.queryService.GetTask(test.TaskID)
			if err == nil {
				context.ParentTask = &TaskDetails{
					ID:                 parentTask.ID,
					PhaseID:            parentTask.PhaseID,
					Name:               parentTask.Name,
					Description:        parentTask.Description,
					AcceptanceCriteria: parentTask.AcceptanceCriteria,
					Status:             parentTask.Status,
					Assignee:           parentTask.Assignee,
					StartedAt:          parentTask.StartedAt,
					CompletedAt:        parentTask.CompletedAt,
				}
			}
		}

		// Get parent phase through the task
		if test.TaskID != "" {
			parentTask, err := e.queryService.GetTask(test.TaskID)
			if err == nil && parentTask.PhaseID != "" {
				parentPhase, err := e.queryService.GetPhase(parentTask.PhaseID)
				if err == nil {
					context.ParentPhase = &PhaseDetails{
						ID:           parentPhase.ID,
						Name:         parentPhase.Name,
						Description:  parentPhase.Description,
						Deliverables: parentPhase.Deliverables,
						Status:       parentPhase.Status,
						StartedAt:    parentPhase.StartedAt,
						CompletedAt:  parentPhase.CompletedAt,
					}
					context.ParentPhase.Progress = e.calculatePhaseProgress(parentPhase.ID)
				}
			}
		}

		// Get sibling tests for the same task
		if test.TaskID != "" {
			context.SiblingTests = e.getSiblingTests(testID, test.TaskID, includeFullDetails)
		}
	}

	return context, nil
}

// Helper methods for calculating progress and retrieving related entities

func (e *Engine) calculatePhaseProgress(phaseID string) *ProgressSummary {
	epicData, err := e.queryService.GetEpic()
	if err != nil {
		return &ProgressSummary{}
	}

	var totalTasks, completedTasks, activeTasks, pendingTasks, cancelledTasks int
	var totalTests, passedTests, failedTests, pendingTests int

	// Count tasks in this phase
	for _, task := range epicData.Tasks {
		if task.PhaseID == phaseID {
			totalTasks++
			switch task.Status {
			case epic.StatusCompleted:
				completedTasks++
			case epic.StatusActive:
				activeTasks++
			case epic.StatusCancelled:
				cancelledTasks++
			default:
				pendingTasks++
			}
		}
	}

	// Count tests for tasks in this phase AND tests directly associated with the phase
	for _, test := range epicData.Tests {
		testBelongsToPhase := false

		// Check if test belongs to a task in this phase
		if test.TaskID != "" {
			for _, task := range epicData.Tasks {
				if task.ID == test.TaskID && task.PhaseID == phaseID {
					testBelongsToPhase = true
					break
				}
			}
		} else if test.PhaseID == phaseID {
			// Test is directly associated with the phase (no task_id)
			testBelongsToPhase = true
		}

		if testBelongsToPhase {
			totalTests++
			switch test.Status {
			case epic.StatusCompleted:
				passedTests++
			default:
				if test.TestStatus == epic.TestStatusFailed {
					failedTests++
				} else {
					pendingTests++
				}
			}
		}
	}

	var completionPercentage, testCoveragePercentage int
	if totalTasks > 0 {
		completionPercentage = (completedTasks * 100) / totalTasks
	}
	if totalTests > 0 {
		testCoveragePercentage = (passedTests * 100) / totalTests
	}

	return &ProgressSummary{
		TotalTasks:             totalTasks,
		CompletedTasks:         completedTasks,
		ActiveTasks:            activeTasks,
		PendingTasks:           pendingTasks,
		CancelledTasks:         cancelledTasks,
		CompletionPercentage:   completionPercentage,
		TotalTests:             totalTests,
		PassedTests:            passedTests,
		FailedTests:            failedTests,
		PendingTests:           pendingTests,
		TestCoveragePercentage: testCoveragePercentage,
	}
}

func (e *Engine) getSiblingTasks(taskID, phaseID string, includeFullDetails bool) []TaskDetails {
	epicData, err := e.queryService.GetEpic()
	if err != nil {
		return nil
	}

	var siblings []TaskDetails
	for _, task := range epicData.Tasks {
		if task.PhaseID == phaseID && task.ID != taskID {
			taskDetails := TaskDetails{
				ID:          task.ID,
				PhaseID:     task.PhaseID,
				Name:        task.Name,
				Status:      task.Status,
				Assignee:    task.Assignee,
				StartedAt:   task.StartedAt,
				CompletedAt: task.CompletedAt,
			}

			if includeFullDetails {
				taskDetails.Description = task.Description
				taskDetails.AcceptanceCriteria = task.AcceptanceCriteria
			}

			siblings = append(siblings, taskDetails)
		}
	}

	return siblings
}

func (e *Engine) getChildTests(taskID string, includeFullDetails bool) []TestDetails {
	epicData, err := e.queryService.GetEpic()
	if err != nil {
		return nil
	}

	var tests []TestDetails
	for _, test := range epicData.Tests {
		if test.TaskID == taskID {
			testDetails := TestDetails{
				ID:         test.ID,
				TaskID:     test.TaskID,
				PhaseID:    test.PhaseID,
				Name:       test.Name,
				Status:     test.Status,
				TestStatus: test.TestStatus,
				StartedAt:  test.StartedAt,
				PassedAt:   test.PassedAt,
				FailedAt:   test.FailedAt,
			}

			if includeFullDetails {
				testDetails.Description = test.Description
				testDetails.FailureNote = test.FailureNote
			}

			tests = append(tests, testDetails)
		}
	}

	return tests
}

func (e *Engine) getTasksInPhase(phaseID string, includeFullDetails bool) []TaskWithTests {
	epicData, err := e.queryService.GetEpic()
	if err != nil {
		return nil
	}

	var tasks []TaskWithTests
	for _, task := range epicData.Tasks {
		if task.PhaseID == phaseID {
			taskDetails := TaskDetails{
				ID:          task.ID,
				PhaseID:     task.PhaseID,
				Name:        task.Name,
				Status:      task.Status,
				Assignee:    task.Assignee,
				StartedAt:   task.StartedAt,
				CompletedAt: task.CompletedAt,
			}

			if includeFullDetails {
				taskDetails.Description = task.Description
				taskDetails.AcceptanceCriteria = task.AcceptanceCriteria
			}

			taskWithTests := TaskWithTests{
				TaskDetails: taskDetails,
				Tests:       e.getChildTests(task.ID, includeFullDetails),
			}

			tasks = append(tasks, taskWithTests)
		}
	}

	return tasks
}

func (e *Engine) getPhaseTests(phaseID string, includeFullDetails bool) []TestDetails {
	epicData, err := e.queryService.GetEpic()
	if err != nil {
		return nil
	}

	var tests []TestDetails
	for _, test := range epicData.Tests {
		// Include tests that are directly associated with the phase (no task_id)
		if test.TaskID == "" && test.PhaseID == phaseID {
			testDetails := TestDetails{
				ID:         test.ID,
				TaskID:     test.TaskID,
				PhaseID:    test.PhaseID,
				Name:       test.Name,
				Status:     test.Status,
				TestStatus: test.TestStatus,
				StartedAt:  test.StartedAt,
				PassedAt:   test.PassedAt,
				FailedAt:   test.FailedAt,
			}

			if includeFullDetails {
				testDetails.Description = test.Description
				testDetails.FailureNote = test.FailureNote
			}

			tests = append(tests, testDetails)
		}
	}

	return tests
}

func (e *Engine) getSiblingPhases(phaseID string, includeFullDetails bool) []PhaseDetails {
	epicData, err := e.queryService.GetEpic()
	if err != nil {
		return nil
	}

	var siblings []PhaseDetails
	for _, phase := range epicData.Phases {
		if phase.ID != phaseID {
			phaseDetails := PhaseDetails{
				ID:          phase.ID,
				Name:        phase.Name,
				Status:      phase.Status,
				StartedAt:   phase.StartedAt,
				CompletedAt: phase.CompletedAt,
			}

			if includeFullDetails {
				phaseDetails.Description = phase.Description
				phaseDetails.Deliverables = phase.Deliverables
			}

			siblings = append(siblings, phaseDetails)
		}
	}

	return siblings
}

func (e *Engine) getSiblingTests(testID, taskID string, includeFullDetails bool) []TestDetails {
	epicData, err := e.queryService.GetEpic()
	if err != nil {
		return nil
	}

	var siblings []TestDetails
	for _, test := range epicData.Tests {
		if test.TaskID == taskID && test.ID != testID {
			testDetails := TestDetails{
				ID:         test.ID,
				TaskID:     test.TaskID,
				PhaseID:    test.PhaseID,
				Name:       test.Name,
				Status:     test.Status,
				TestStatus: test.TestStatus,
				StartedAt:  test.StartedAt,
				PassedAt:   test.PassedAt,
				FailedAt:   test.FailedAt,
			}

			if includeFullDetails {
				testDetails.Description = test.Description
				testDetails.FailureNote = test.FailureNote
			}

			siblings = append(siblings, testDetails)
		}
	}

	return siblings
}
