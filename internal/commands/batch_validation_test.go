package commands

import (
	"strings"
	"testing"

	"github.com/mindreframer/agentpm/internal/epic"
)

func TestBatchValidationService_ValidateBatchOperations(t *testing.T) {
	bvs := NewBatchValidationService()

	// Create test epic with active phase and tests
	testEpic := &epic.Epic{
		ID:     "test_epic",
		Status: epic.StatusActive,
		Phases: []epic.Phase{
			{
				ID:     "phase1",
				Status: epic.StatusActive,
			},
			{
				ID:     "phase2",
				Status: epic.StatusPending,
			},
		},
		Tests: []epic.Test{
			{
				ID:         "test1",
				PhaseID:    "phase1",
				Name:       "Valid WIP Test",
				TestStatus: epic.TestStatusWIP,
			},
			{
				ID:         "test2",
				PhaseID:    "phase1",
				Name:       "Valid Done Test",
				TestStatus: epic.TestStatusDone,
				TestResult: epic.TestResultPassing,
			},
			{
				ID:         "test3",
				PhaseID:    "phase2",
				Name:       "Wrong Phase Test",
				TestStatus: epic.TestStatusWIP,
			},
			{
				ID:         "test4",
				PhaseID:    "phase1",
				Name:       "Pending Test",
				TestStatus: epic.TestStatusPending,
			},
		},
	}

	t.Run("all valid operations should pass", func(t *testing.T) {
		operations := []BatchOperation{
			{TestID: "test1", OperationType: "pass"},
			{TestID: "test2", OperationType: "fail", Reason: "Found issue"},
		}

		result, err := bvs.ValidateBatchOperations(testEpic, operations)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if !result.Valid {
			t.Errorf("Expected batch to be valid, got invalid")
		}

		if result.Summary.TotalOperations != 2 {
			t.Errorf("Expected 2 total operations, got %d", result.Summary.TotalOperations)
		}

		if result.Summary.ValidOperations != 2 {
			t.Errorf("Expected 2 valid operations, got %d", result.Summary.ValidOperations)
		}

		if result.Summary.InvalidOperations != 0 {
			t.Errorf("Expected 0 invalid operations, got %d", result.Summary.InvalidOperations)
		}
	})

	t.Run("any invalid operation should fail entire batch", func(t *testing.T) {
		operations := []BatchOperation{
			{TestID: "test1", OperationType: "pass"},                  // Valid
			{TestID: "test3", OperationType: "pass"},                  // Invalid - wrong phase
			{TestID: "test2", OperationType: "fail", Reason: "Issue"}, // Valid
		}

		result, err := bvs.ValidateBatchOperations(testEpic, operations)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.Valid {
			t.Errorf("Expected batch to be invalid due to one failing operation")
		}

		if result.Summary.TotalOperations != 3 {
			t.Errorf("Expected 3 total operations, got %d", result.Summary.TotalOperations)
		}

		if result.Summary.ValidOperations != 2 {
			t.Errorf("Expected 2 valid operations, got %d", result.Summary.ValidOperations)
		}

		if result.Summary.InvalidOperations != 1 {
			t.Errorf("Expected 1 invalid operation, got %d", result.Summary.InvalidOperations)
		}
	})

	t.Run("test not found should be tracked", func(t *testing.T) {
		operations := []BatchOperation{
			{TestID: "nonexistent", OperationType: "pass"},
		}

		result, err := bvs.ValidateBatchOperations(testEpic, operations)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.Valid {
			t.Errorf("Expected batch to be invalid")
		}

		if result.Summary.TestsNotFound != 1 {
			t.Errorf("Expected 1 test not found, got %d", result.Summary.TestsNotFound)
		}
	})

	t.Run("phase violations should be tracked", func(t *testing.T) {
		operations := []BatchOperation{
			{TestID: "test3", OperationType: "pass"}, // Wrong phase
		}

		result, err := bvs.ValidateBatchOperations(testEpic, operations)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.Valid {
			t.Errorf("Expected batch to be invalid")
		}

		if result.Summary.PhaseViolations != 1 {
			t.Errorf("Expected 1 phase violation, got %d", result.Summary.PhaseViolations)
		}
	})

	t.Run("status violations should be tracked", func(t *testing.T) {
		operations := []BatchOperation{
			{TestID: "test4", OperationType: "pass"}, // Pending test can't be passed directly
		}

		result, err := bvs.ValidateBatchOperations(testEpic, operations)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.Valid {
			t.Errorf("Expected batch to be invalid")
		}

		if result.Summary.StatusViolations != 1 {
			t.Errorf("Expected 1 status violation, got %d", result.Summary.StatusViolations)
		}
	})

	t.Run("comprehensive error message should be generated", func(t *testing.T) {
		operations := []BatchOperation{
			{TestID: "nonexistent", OperationType: "pass"}, // Test not found
			{TestID: "test3", OperationType: "pass"},       // Phase violation
			{TestID: "test4", OperationType: "pass"},       // Status violation
		}

		result, err := bvs.ValidateBatchOperations(testEpic, operations)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.Valid {
			t.Errorf("Expected batch to be invalid")
		}

		if result.ErrorMessage == "" {
			t.Errorf("Expected error message to be generated")
		}

		// Check error message contains expected information
		if !strings.Contains(result.ErrorMessage, "3 of 3 operations are invalid") {
			t.Errorf("Expected error message to mention operation counts")
		}

		if !strings.Contains(result.ErrorMessage, "1 tests not found") {
			t.Errorf("Expected error message to mention tests not found")
		}

		if !strings.Contains(result.ErrorMessage, "Failed operations:") {
			t.Errorf("Expected error message to list failed operations")
		}
	})
}

func TestBatchValidationService_CancelOperationValidation(t *testing.T) {
	bvs := NewBatchValidationService()

	testEpic := &epic.Epic{
		Phases: []epic.Phase{
			{ID: "phase1", Status: epic.StatusActive},
		},
		Tests: []epic.Test{
			{
				ID:         "test1",
				PhaseID:    "phase1",
				Name:       "WIP Test",
				TestStatus: epic.TestStatusWIP,
			},
			{
				ID:         "test2",
				PhaseID:    "phase1",
				Name:       "Done Test",
				TestStatus: epic.TestStatusDone,
			},
		},
	}

	t.Run("cancel with reason should be valid", func(t *testing.T) {
		operations := []BatchOperation{
			{TestID: "test1", OperationType: "cancel", Reason: "No longer needed"},
		}

		result, err := bvs.ValidateBatchOperations(testEpic, operations)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if !result.Valid {
			t.Errorf("Expected batch to be valid")
		}
	})

	t.Run("cancel without reason should be invalid", func(t *testing.T) {
		operations := []BatchOperation{
			{TestID: "test1", OperationType: "cancel", Reason: ""},
		}

		result, err := bvs.ValidateBatchOperations(testEpic, operations)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.Valid {
			t.Errorf("Expected batch to be invalid")
		}
	})

	t.Run("cancel completed test should be invalid", func(t *testing.T) {
		operations := []BatchOperation{
			{TestID: "test2", OperationType: "cancel", Reason: "Cannot cancel"},
		}

		result, err := bvs.ValidateBatchOperations(testEpic, operations)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.Valid {
			t.Errorf("Expected batch to be invalid")
		}
	})
}

func TestBatchValidationService_EmptyOperations(t *testing.T) {
	bvs := NewBatchValidationService()
	testEpic := &epic.Epic{}

	t.Run("empty operations should return error", func(t *testing.T) {
		operations := []BatchOperation{}

		_, err := bvs.ValidateBatchOperations(testEpic, operations)
		if err == nil {
			t.Errorf("Expected error for empty operations")
		}

		if !strings.Contains(err.Error(), "no operations provided") {
			t.Errorf("Expected error message about no operations, got: %v", err)
		}
	})
}

func TestBatchValidationService_InvalidOperationType(t *testing.T) {
	bvs := NewBatchValidationService()

	testEpic := &epic.Epic{
		Phases: []epic.Phase{
			{ID: "phase1", Status: epic.StatusActive},
		},
		Tests: []epic.Test{
			{
				ID:         "test1",
				PhaseID:    "phase1",
				Name:       "Test",
				TestStatus: epic.TestStatusWIP,
			},
		},
	}

	t.Run("invalid operation type should fail", func(t *testing.T) {
		operations := []BatchOperation{
			{TestID: "test1", OperationType: "invalid_operation"},
		}

		result, err := bvs.ValidateBatchOperations(testEpic, operations)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.Valid {
			t.Errorf("Expected batch to be invalid")
		}

		if len(result.InvalidOperations) != 1 {
			t.Errorf("Expected 1 invalid operation, got %d", len(result.InvalidOperations))
		}
	})
}

func TestBatchValidationService_CreateBatchSuccessReport(t *testing.T) {
	bvs := NewBatchValidationService()

	operations := []BatchOperation{
		{TestID: "test1", OperationType: "pass"},
		{TestID: "test2", OperationType: "fail", Reason: "Failed"},
		{TestID: "test3", OperationType: "cancel", Reason: "Cancelled"},
	}

	results := []BatchOperationResult{
		{TestID: "test1", TestName: "Test 1", OperationType: "pass", Valid: true},
		{TestID: "test2", TestName: "Test 2", OperationType: "fail", Valid: true},
		{TestID: "test3", TestName: "Test 3", OperationType: "cancel", Valid: true},
	}

	report := bvs.CreateBatchSuccessReport(operations, results)

	if report.TotalOperations != 3 {
		t.Errorf("Expected 3 total operations, got %d", report.TotalOperations)
	}

	if len(report.SuccessfulOperations) != 3 {
		t.Errorf("Expected 3 successful operations, got %d", len(report.SuccessfulOperations))
	}

	if report.Summary.PassedTests != 1 {
		t.Errorf("Expected 1 passed test, got %d", report.Summary.PassedTests)
	}

	if report.Summary.FailedTests != 1 {
		t.Errorf("Expected 1 failed test, got %d", report.Summary.FailedTests)
	}

	if report.Summary.CancelledTests != 1 {
		t.Errorf("Expected 1 cancelled test, got %d", report.Summary.CancelledTests)
	}
}

func TestBatchValidationService_AllOrNothingPrinciple(t *testing.T) {
	bvs := NewBatchValidationService()

	// Create epic with many tests - some valid, some invalid
	testEpic := &epic.Epic{
		Phases: []epic.Phase{
			{ID: "phase1", Status: epic.StatusActive},
			{ID: "phase2", Status: epic.StatusPending},
		},
		Tests: []epic.Test{
			{ID: "test1", PhaseID: "phase1", Name: "Valid Test 1", TestStatus: epic.TestStatusWIP},
			{ID: "test2", PhaseID: "phase1", Name: "Valid Test 2", TestStatus: epic.TestStatusWIP},
			{ID: "test3", PhaseID: "phase1", Name: "Valid Test 3", TestStatus: epic.TestStatusWIP},
			{ID: "test4", PhaseID: "phase1", Name: "Valid Test 4", TestStatus: epic.TestStatusWIP},
			{ID: "test5", PhaseID: "phase2", Name: "Invalid Test - Wrong Phase", TestStatus: epic.TestStatusWIP},
		},
	}

	t.Run("one invalid operation should fail entire large batch", func(t *testing.T) {
		operations := []BatchOperation{
			{TestID: "test1", OperationType: "pass"},
			{TestID: "test2", OperationType: "pass"},
			{TestID: "test3", OperationType: "pass"},
			{TestID: "test4", OperationType: "pass"},
			{TestID: "test5", OperationType: "pass"}, // This one will fail - wrong phase
		}

		result, err := bvs.ValidateBatchOperations(testEpic, operations)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// All-or-nothing principle: entire batch should be invalid
		if result.Valid {
			t.Errorf("Expected entire batch to be invalid due to one failing operation")
		}

		// But individual operations should be tracked correctly
		if result.Summary.ValidOperations != 4 {
			t.Errorf("Expected 4 valid operations, got %d", result.Summary.ValidOperations)
		}

		if result.Summary.InvalidOperations != 1 {
			t.Errorf("Expected 1 invalid operation, got %d", result.Summary.InvalidOperations)
		}

		// Error message should explain the failure
		if result.ErrorMessage == "" {
			t.Errorf("Expected error message for failed batch")
		}
	})
}

func TestBatchValidationService_MixedOperationTypes(t *testing.T) {
	bvs := NewBatchValidationService()

	testEpic := &epic.Epic{
		Phases: []epic.Phase{
			{ID: "phase1", Status: epic.StatusActive},
		},
		Tests: []epic.Test{
			{ID: "test1", PhaseID: "phase1", Name: "WIP Test", TestStatus: epic.TestStatusWIP},
			{ID: "test2", PhaseID: "phase1", Name: "Done Test", TestStatus: epic.TestStatusDone, TestResult: epic.TestResultPassing},
			{ID: "test3", PhaseID: "phase1", Name: "Another WIP Test", TestStatus: epic.TestStatusWIP},
		},
	}

	t.Run("mixed valid operations should pass", func(t *testing.T) {
		operations := []BatchOperation{
			{TestID: "test1", OperationType: "pass"},
			{TestID: "test2", OperationType: "fail", Reason: "Found issue"},
			{TestID: "test3", OperationType: "cancel", Reason: "No longer needed"},
		}

		result, err := bvs.ValidateBatchOperations(testEpic, operations)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if !result.Valid {
			t.Errorf("Expected batch to be valid, got: %s", result.ErrorMessage)
		}

		if result.Summary.TotalOperations != 3 {
			t.Errorf("Expected 3 total operations, got %d", result.Summary.TotalOperations)
		}

		if result.Summary.ValidOperations != 3 {
			t.Errorf("Expected 3 valid operations, got %d", result.Summary.ValidOperations)
		}
	})
}
