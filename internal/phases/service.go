package phases

import (
	"fmt"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/service"
	"github.com/mindreframer/agentpm/internal/storage"
)

type PhaseService struct {
	storage storage.Storage
	query   *query.QueryService
}

func NewPhaseService(storage storage.Storage, query *query.QueryService) *PhaseService {
	return &PhaseService{
		storage: storage,
		query:   query,
	}
}

// PhaseStartResult represents the result of starting a phase
type PhaseStartResult struct {
	AlreadyActive bool
	Error         error
}

// StartPhase transitions a phase from pending to active
func (s *PhaseService) StartPhase(epicData *epic.Epic, phaseID string, timestamp time.Time) error {
	// Find the phase
	phase := s.findPhase(epicData, phaseID)
	if phase == nil {
		return fmt.Errorf("phase %s not found", phaseID)
	}

	// Check if phase is already active
	if phase.Status == epic.StatusWIP {
		// Return a special error type to indicate "already started"
		return NewPhaseAlreadyActiveError(phaseID)
	}

	// Validate phase can be started
	if err := s.validatePhaseStart(epicData, phase); err != nil {
		return err
	}

	// Transition phase status and set timestamp
	phase.Status = epic.StatusWIP
	phase.StartedAt = &timestamp

	// Create automatic event for phase start
	service.CreateEvent(epicData, service.EventPhaseStarted, phaseID, "", "", "", timestamp)

	return nil
}

// CompletePhase transitions a phase from wip to done
func (s *PhaseService) CompletePhase(epicData *epic.Epic, phaseID string, timestamp time.Time) error {
	// Find the phase
	phase := s.findPhase(epicData, phaseID)
	if phase == nil {
		return fmt.Errorf("phase %s not found", phaseID)
	}

	// Validate phase can be completed
	if err := s.validatePhaseCompletion(epicData, phase); err != nil {
		return err
	}

	// Transition phase status and set timestamp
	phase.Status = epic.StatusCompleted
	phase.CompletedAt = &timestamp

	// Create automatic event for phase completion
	service.CreateEvent(epicData, service.EventPhaseCompleted, phaseID, "", "", "", timestamp)

	return nil
}

// GetActivePhase returns the currently active phase, if any
func (s *PhaseService) GetActivePhase(epicData *epic.Epic) *epic.Phase {
	for i := range epicData.Phases {
		if epicData.Phases[i].Status == epic.StatusWIP {
			return &epicData.Phases[i]
		}
	}
	return nil
}

// findPhase returns a pointer to the phase with the given ID
func (s *PhaseService) findPhase(epicData *epic.Epic, phaseID string) *epic.Phase {
	for i := range epicData.Phases {
		if epicData.Phases[i].ID == phaseID {
			return &epicData.Phases[i]
		}
	}
	return nil
}

// validatePhaseStart checks if a phase can be started
func (s *PhaseService) validatePhaseStart(epicData *epic.Epic, phase *epic.Phase) error {
	// Check if phase is already active - this is not an error, just a no-op
	if phase.Status == epic.StatusWIP {
		return nil // Let the caller handle this as "already started"
	}

	// Check phase is in pending status (accept both "planning" and "pending" for backward compatibility)
	if phase.Status != epic.StatusPending {
		return NewPhaseStateError(phase.ID, phase.Status, epic.StatusWIP, "Phase is not in pending state")
	}

	// Check no other phase is active
	activePhase := s.GetActivePhase(epicData)
	if activePhase != nil && activePhase.ID != phase.ID {
		return NewPhaseConstraintError(phase.ID, activePhase.ID, "Cannot start phase: another phase is already active")
	}

	// Check prerequisite tests from earlier phases are completed
	prerequisiteTests := s.getIncompleteTestsInEarlierPhases(epicData, phase.ID)
	if len(prerequisiteTests) > 0 {
		return NewPhaseTestPrerequisiteError(phase.ID, prerequisiteTests)
	}

	return nil
}

// validatePhaseCompletion checks if a phase can be completed
func (s *PhaseService) validatePhaseCompletion(epicData *epic.Epic, phase *epic.Phase) error {
	// Check phase is in active status
	if phase.Status != epic.StatusWIP {
		return NewPhaseStateError(phase.ID, phase.Status, epic.StatusCompleted, "Phase is not in active state")
	}

	// Check all tasks in phase are completed or cancelled
	pendingTasks := s.getPendingTasksInPhase(epicData, phase.ID)
	if len(pendingTasks) > 0 {
		return NewPhaseIncompleteError(phase.ID, pendingTasks)
	}

	// Check all tests in phase are completed (passed)
	incompleteTests := s.getIncompleteTestsInPhase(epicData, phase.ID)
	if len(incompleteTests) > 0 {
		return NewPhaseTestDependencyError(phase.ID, incompleteTests)
	}

	return nil
}

// getPendingTasksInPhase returns tasks in the phase that are not done or cancelled
func (s *PhaseService) getPendingTasksInPhase(epicData *epic.Epic, phaseID string) []epic.Task {
	var pendingTasks []epic.Task
	for _, task := range epicData.Tasks {
		if task.PhaseID == phaseID && task.Status != epic.StatusCompleted && task.Status != epic.StatusCancelled {
			pendingTasks = append(pendingTasks, task)
		}
	}
	return pendingTasks
}

// getIncompleteTestsInPhase returns tests in the phase that are not passed
func (s *PhaseService) getIncompleteTestsInPhase(epicData *epic.Epic, phaseID string) []epic.Test {
	var incompleteTests []epic.Test
	for _, test := range epicData.Tests {
		if test.PhaseID == phaseID && !s.isTestCompleted(test) {
			incompleteTests = append(incompleteTests, test)
		}
	}
	return incompleteTests
}

// isTestCompleted checks if a test is considered completed (passed)
func (s *PhaseService) isTestCompleted(test epic.Test) bool {
	// Test is completed if Status is completed AND TestStatus is passed
	// Cancelled tests are also considered "completed" for dependency purposes
	// For backwards compatibility, if TestStatus is not set, only check Status
	if test.TestStatus != "" {
		return (test.Status == epic.StatusCompleted && test.TestStatus == epic.TestStatusDone) ||
			(test.Status == epic.StatusCancelled && test.TestStatus == epic.TestStatusCancelled)
	}
	return test.Status == epic.StatusCompleted || test.Status == epic.StatusCancelled
}

// getIncompleteTestsInEarlierPhases returns tests from earlier phases that are not completed
func (s *PhaseService) getIncompleteTestsInEarlierPhases(epicData *epic.Epic, currentPhaseID string) []epic.Test {
	var incompleteTests []epic.Test
	currentPhaseIndex := s.getPhaseIndex(epicData, currentPhaseID)

	// Check all phases before the current one
	for i := 0; i < currentPhaseIndex; i++ {
		phaseID := epicData.Phases[i].ID
		phaseIncompleteTests := s.getIncompleteTestsInPhase(epicData, phaseID)
		incompleteTests = append(incompleteTests, phaseIncompleteTests...)
	}

	return incompleteTests
}

// getPhaseIndex returns the index of a phase in the phases slice
func (s *PhaseService) getPhaseIndex(epicData *epic.Epic, phaseID string) int {
	for i, phase := range epicData.Phases {
		if phase.ID == phaseID {
			return i
		}
	}
	return -1
}

// GetTestCompletionStatus returns detailed status of tests in a phase
func (s *PhaseService) GetTestCompletionStatus(epicData *epic.Epic, phaseID string) TestCompletionStatus {
	var totalTests, passedTests, failedTests, pendingTests int
	var incompleteTests []epic.Test

	for _, test := range epicData.Tests {
		if test.PhaseID == phaseID {
			totalTests++

			// Check specific test status for accurate counting
			if test.TestStatus != "" {
				switch test.TestStatus {
				case epic.TestStatusDone:
					passedTests++
				case epic.TestStatusWIP:
					failedTests++
					incompleteTests = append(incompleteTests, test)
				case epic.TestStatusCancelled:
					// Cancelled tests are complete but not "passed"
				default: // pending, wip
					pendingTests++
					incompleteTests = append(incompleteTests, test)
				}
			} else {
				// Legacy test without TestStatus - use Status field
				if test.Status == epic.StatusCompleted {
					passedTests++
				} else if test.Status == epic.StatusCancelled {
					// Cancelled tests are complete but not "passed"
				} else {
					pendingTests++
					incompleteTests = append(incompleteTests, test)
				}
			}
		}
	}

	return TestCompletionStatus{
		PhaseID:           phaseID,
		TotalTests:        totalTests,
		PassedTests:       passedTests,
		FailedTests:       failedTests,
		PendingTests:      pendingTests,
		IncompleteTests:   incompleteTests,
		AllTestsCompleted: len(incompleteTests) == 0 && totalTests > 0,
	}
}

// GetOverallTestCompletionStatus returns completion status for all phases
func (s *PhaseService) GetOverallTestCompletionStatus(epicData *epic.Epic) map[string]TestCompletionStatus {
	statusMap := make(map[string]TestCompletionStatus)

	for _, phase := range epicData.Phases {
		statusMap[phase.ID] = s.GetTestCompletionStatus(epicData, phase.ID)
	}

	return statusMap
}

// TestCompletionStatus represents the test completion status for a phase
type TestCompletionStatus struct {
	PhaseID           string
	TotalTests        int
	PassedTests       int
	FailedTests       int
	PendingTests      int
	IncompleteTests   []epic.Test
	AllTestsCompleted bool
}
