package tests

import (
	"fmt"

	"github.com/mindreframer/agentpm/internal/epic"
)

type TestValidationService struct{}

func NewTestValidationService() *TestValidationService {
	return &TestValidationService{}
}

// ValidateTestStatusTransition validates transitions between test statuses
func (tvs *TestValidationService) ValidateTestStatusTransition(currentStatus, targetStatus epic.TestStatus) error {
	validTransitions := map[epic.TestStatus][]epic.TestStatus{
		epic.TestStatusPending:   {epic.TestStatusWIP, epic.TestStatusCancelled},
		epic.TestStatusWIP:       {epic.TestStatusDone, epic.TestStatusCancelled},
		epic.TestStatusDone:      {epic.TestStatusWIP}, // Can go back to WIP for failing tests
		epic.TestStatusCancelled: {},                   // Cancelled is terminal
	}

	validTargets, exists := validTransitions[currentStatus]
	if !exists {
		return fmt.Errorf("invalid current test status: %s", currentStatus)
	}

	for _, validTarget := range validTargets {
		if validTarget == targetStatus {
			return nil
		}
	}

	return fmt.Errorf("invalid test status transition from %s to %s", currentStatus, targetStatus)
}

// ValidateTestResultTransition validates transitions between test results
func (tvs *TestValidationService) ValidateTestResultTransition(currentResult, targetResult epic.TestResult, testStatus epic.TestStatus) error {
	// Test must be in done status to have a meaningful result
	if testStatus == epic.TestStatusDone {
		// Any result transition is valid for done tests
		return nil
	}

	// For non-done tests, result should generally be failing if set
	if testStatus == epic.TestStatusWIP && targetResult == epic.TestResultFailing {
		return nil
	}

	if testStatus == epic.TestStatusPending {
		return fmt.Errorf("pending tests should not have results set")
	}

	if testStatus == epic.TestStatusCancelled {
		return fmt.Errorf("cancelled tests should not have results changed")
	}

	return nil
}

// ValidateTestPassTransition validates marking a test as passing
func (tvs *TestValidationService) ValidateTestPassTransition(test *epic.Test) error {
	if test == nil {
		return fmt.Errorf("test cannot be nil")
	}

	// Test must be in WIP status to be marked as passing
	if test.TestStatus != epic.TestStatusWIP {
		return fmt.Errorf("test %s must be in WIP status to be marked as passing, current status: %s", test.ID, test.TestStatus)
	}

	return nil
}

// ValidateTestFailTransition validates marking a test as failing
func (tvs *TestValidationService) ValidateTestFailTransition(test *epic.Test) error {
	if test == nil {
		return fmt.Errorf("test cannot be nil")
	}

	// Test can be marked as failing from WIP or Done status
	if test.TestStatus != epic.TestStatusWIP && test.TestStatus != epic.TestStatusDone {
		return fmt.Errorf("test %s must be in WIP or Done status to be marked as failing, current status: %s", test.ID, test.TestStatus)
	}

	return nil
}

// ValidateFailingTestCannotBeDone implements the core Epic 13 rule
func (tvs *TestValidationService) ValidateFailingTestCannotBeDone(test *epic.Test) error {
	if test == nil {
		return fmt.Errorf("test cannot be nil")
	}

	// If test is being marked as done, it cannot have a failing result
	if test.TestStatus == epic.TestStatusDone && test.TestResult == epic.TestResultFailing {
		return &epic.StatusValidationError{
			EntityType:    "test",
			EntityID:      test.ID,
			EntityName:    test.Name,
			CurrentStatus: string(test.TestStatus),
			TargetStatus:  string(epic.TestStatusDone),
			BlockingItems: []epic.BlockingItem{
				{
					Type:   "test",
					ID:     test.ID,
					Name:   test.Name,
					Status: string(test.TestStatus),
					Result: string(test.TestResult),
				},
			},
			Message: fmt.Sprintf("Test %s cannot be marked as done while result is failing", test.ID),
		}
	}

	return nil
}

// ValidateTestCancellation validates test cancellation with reason
func (tvs *TestValidationService) ValidateTestCancellation(test *epic.Test, reason string) error {
	if test == nil {
		return fmt.Errorf("test cannot be nil")
	}

	if test.TestStatus == epic.TestStatusCancelled {
		return fmt.Errorf("test %s is already cancelled", test.ID)
	}

	if test.TestStatus == epic.TestStatusDone {
		return fmt.Errorf("cannot cancel completed test %s", test.ID)
	}

	if reason == "" {
		return fmt.Errorf("cancellation reason is required for test %s", test.ID)
	}

	return nil
}

// ValidateTestActivePhase ensures test operations only happen in active phase
func (tvs *TestValidationService) ValidateTestActivePhase(epicData *epic.Epic, test *epic.Test) error {
	if test == nil {
		return fmt.Errorf("test cannot be nil")
	}

	// Find the active phase
	var activePhase *epic.Phase
	for i := range epicData.Phases {
		if epicData.Phases[i].Status == epic.StatusActive {
			activePhase = &epicData.Phases[i]
			break
		}
	}

	if activePhase == nil {
		return fmt.Errorf("no active phase found - test operations require an active phase")
	}

	// Check if test belongs to the active phase
	if test.PhaseID != activePhase.ID {
		return fmt.Errorf("test %s belongs to phase %s, but active phase is %s", test.ID, test.PhaseID, activePhase.ID)
	}

	return nil
}

// ValidateTestStateConsistency validates overall test state consistency
func (tvs *TestValidationService) ValidateTestStateConsistency(test *epic.Test) error {
	if test == nil {
		return fmt.Errorf("test cannot be nil")
	}

	// Check status and result consistency
	if test.TestStatus == epic.TestStatusPending && test.TestResult != "" {
		return fmt.Errorf("test %s: pending tests should not have results", test.ID)
	}

	if test.TestStatus == epic.TestStatusCancelled && test.TestResult != "" {
		return fmt.Errorf("test %s: cancelled tests should not have results", test.ID)
	}

	if test.TestStatus == epic.TestStatusDone && test.TestResult == epic.TestResultFailing {
		return fmt.Errorf("test %s: done tests cannot have failing results", test.ID)
	}

	// Check timestamp consistency
	if test.TestStatus != epic.TestStatusDone && test.PassedAt != nil {
		return fmt.Errorf("test %s: only done tests should have PassedAt timestamps", test.ID)
	}

	if test.TestResult != epic.TestResultFailing && test.FailedAt != nil {
		return fmt.Errorf("test %s: only failing tests should have FailedAt timestamps", test.ID)
	}

	if test.TestStatus != epic.TestStatusCancelled && test.CancelledAt != nil {
		return fmt.Errorf("test %s: only cancelled tests should have CancelledAt timestamps", test.ID)
	}

	return nil
}

// CanPassTest checks if a test can be marked as passing
func (tvs *TestValidationService) CanPassTest(epicData *epic.Epic, test *epic.Test) error {
	// Validate active phase
	if err := tvs.ValidateTestActivePhase(epicData, test); err != nil {
		return err
	}

	// Validate pass transition
	if err := tvs.ValidateTestPassTransition(test); err != nil {
		return err
	}

	return nil
}

// CanFailTest checks if a test can be marked as failing
func (tvs *TestValidationService) CanFailTest(epicData *epic.Epic, test *epic.Test) error {
	// Validate active phase
	if err := tvs.ValidateTestActivePhase(epicData, test); err != nil {
		return err
	}

	// Validate fail transition
	if err := tvs.ValidateTestFailTransition(test); err != nil {
		return err
	}

	return nil
}

// CanCancelTest checks if a test can be cancelled
func (tvs *TestValidationService) CanCancelTest(epicData *epic.Epic, test *epic.Test, reason string) error {
	// Validate active phase
	if err := tvs.ValidateTestActivePhase(epicData, test); err != nil {
		return err
	}

	// Validate cancellation
	if err := tvs.ValidateTestCancellation(test, reason); err != nil {
		return err
	}

	return nil
}
