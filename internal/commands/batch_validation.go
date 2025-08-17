package commands

import (
	"fmt"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/tests"
)

// BatchValidationService handles validation for batch test operations
type BatchValidationService struct {
	testValidator *tests.TestValidationService
}

// NewBatchValidationService creates a new batch validation service
func NewBatchValidationService() *BatchValidationService {
	return &BatchValidationService{
		testValidator: tests.NewTestValidationService(),
	}
}

// BatchOperation represents a single operation in a batch
type BatchOperation struct {
	TestID        string
	OperationType string // "pass", "fail", "cancel"
	Reason        string // For fail/cancel operations
}

// BatchValidationResult represents the result of batch validation
type BatchValidationResult struct {
	Valid             bool                   `json:"valid"`
	ValidOperations   []BatchOperationResult `json:"valid_operations"`
	InvalidOperations []BatchOperationResult `json:"invalid_operations"`
	Summary           BatchValidationSummary `json:"summary"`
	ErrorMessage      string                 `json:"error_message,omitempty"`
}

// BatchOperationResult represents the validation result for a single operation
type BatchOperationResult struct {
	TestID        string                      `json:"test_id"`
	TestName      string                      `json:"test_name"`
	OperationType string                      `json:"operation_type"`
	Valid         bool                        `json:"valid"`
	Error         *epic.StatusValidationError `json:"error,omitempty"`
	Test          *epic.Test                  `json:"test,omitempty"`
}

// BatchValidationSummary provides counts and overall status
type BatchValidationSummary struct {
	TotalOperations   int `json:"total_operations"`
	ValidOperations   int `json:"valid_operations"`
	InvalidOperations int `json:"invalid_operations"`
	TestsNotFound     int `json:"tests_not_found"`
	PhaseViolations   int `json:"phase_violations"`
	StatusViolations  int `json:"status_violations"`
}

// ValidateBatchOperations performs all-or-nothing validation for batch operations
func (bvs *BatchValidationService) ValidateBatchOperations(epicData *epic.Epic, operations []BatchOperation) (*BatchValidationResult, error) {
	if len(operations) == 0 {
		return nil, fmt.Errorf("no operations provided for batch validation")
	}

	result := &BatchValidationResult{
		Valid:             true,
		ValidOperations:   []BatchOperationResult{},
		InvalidOperations: []BatchOperationResult{},
		Summary: BatchValidationSummary{
			TotalOperations: len(operations),
		},
	}

	// Validate each operation
	for _, op := range operations {
		opResult := bvs.validateSingleOperation(epicData, op)

		if opResult.Valid {
			result.ValidOperations = append(result.ValidOperations, opResult)
			result.Summary.ValidOperations++
		} else {
			result.InvalidOperations = append(result.InvalidOperations, opResult)
			result.Summary.InvalidOperations++
			result.Valid = false // All-or-nothing: any failure invalidates the batch

			// Count specific error types
			if opResult.Test == nil {
				result.Summary.TestsNotFound++
			} else if opResult.Error != nil {
				if containsPhaseError(opResult.Error.Message) {
					result.Summary.PhaseViolations++
				} else {
					result.Summary.StatusViolations++
				}
			}
		}
	}

	// Generate error message if validation failed
	if !result.Valid {
		result.ErrorMessage = bvs.generateBatchErrorMessage(result)
	}

	return result, nil
}

// validateSingleOperation validates a single operation in the batch
func (bvs *BatchValidationService) validateSingleOperation(epicData *epic.Epic, op BatchOperation) BatchOperationResult {
	// Find the test
	var test *epic.Test
	for i := range epicData.Tests {
		if epicData.Tests[i].ID == op.TestID {
			test = &epicData.Tests[i]
			break
		}
	}

	result := BatchOperationResult{
		TestID:        op.TestID,
		OperationType: op.OperationType,
		Test:          test,
	}

	if test != nil {
		result.TestName = test.Name
	}

	// Test existence validation
	if test == nil {
		result.Valid = false
		result.Error = &epic.StatusValidationError{
			EntityType: "test",
			EntityID:   op.TestID,
			Message:    fmt.Sprintf("Test %s not found", op.TestID),
		}
		return result
	}

	// Operation-specific validation
	var err error
	switch op.OperationType {
	case "pass":
		err = bvs.testValidator.CanPassTest(epicData, test)
	case "fail":
		err = bvs.testValidator.CanFailTest(epicData, test)
	case "cancel":
		err = bvs.testValidator.CanCancelTest(epicData, test, op.Reason)
	default:
		err = fmt.Errorf("invalid operation type: %s", op.OperationType)
	}

	if err != nil {
		result.Valid = false
		if statusErr, ok := err.(*epic.StatusValidationError); ok {
			result.Error = statusErr
		} else {
			result.Error = &epic.StatusValidationError{
				EntityType: "test",
				EntityID:   op.TestID,
				Message:    err.Error(),
			}
		}
	} else {
		result.Valid = true
	}

	return result
}

// generateBatchErrorMessage creates a comprehensive error message for batch failures
func (bvs *BatchValidationService) generateBatchErrorMessage(result *BatchValidationResult) string {
	summary := result.Summary
	msg := fmt.Sprintf("Batch operation failed: %d of %d operations are invalid",
		summary.InvalidOperations, summary.TotalOperations)

	if summary.TestsNotFound > 0 {
		msg += fmt.Sprintf("\n- %d tests not found", summary.TestsNotFound)
	}
	if summary.PhaseViolations > 0 {
		msg += fmt.Sprintf("\n- %d phase violations (tests not in active phase)", summary.PhaseViolations)
	}
	if summary.StatusViolations > 0 {
		msg += fmt.Sprintf("\n- %d status violations (invalid transitions)", summary.StatusViolations)
	}

	msg += "\n\nFailed operations:"
	for _, invalid := range result.InvalidOperations {
		msg += fmt.Sprintf("\n- %s (%s): %s", invalid.TestID, invalid.OperationType, invalid.Error.Message)
	}

	return msg
}

// containsPhaseError checks if an error message is related to phase validation
func containsPhaseError(message string) bool {
	phaseIndicators := []string{
		"active phase",
		"no active phase",
		"belongs to phase",
		"wrong phase",
	}

	for _, indicator := range phaseIndicators {
		if len(message) > 0 && contains(message, indicator) {
			return true
		}
	}
	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// CreateBatchSuccessReport generates a detailed success report for completed batch operations
func (bvs *BatchValidationService) CreateBatchSuccessReport(operations []BatchOperation, results []BatchOperationResult) BatchSuccessReport {
	report := BatchSuccessReport{
		TotalOperations:      len(operations),
		SuccessfulOperations: []BatchOperationSummary{},
		Summary:              BatchOperationsSummary{},
	}

	passCount := 0
	failCount := 0
	cancelCount := 0

	for i, op := range operations {
		if i < len(results) && results[i].Valid {
			summary := BatchOperationSummary{
				TestID:        op.TestID,
				TestName:      results[i].TestName,
				OperationType: op.OperationType,
				Reason:        op.Reason,
			}
			report.SuccessfulOperations = append(report.SuccessfulOperations, summary)

			switch op.OperationType {
			case "pass":
				passCount++
			case "fail":
				failCount++
			case "cancel":
				cancelCount++
			}
		}
	}

	report.Summary.PassedTests = passCount
	report.Summary.FailedTests = failCount
	report.Summary.CancelledTests = cancelCount

	return report
}

// BatchSuccessReport represents a successful batch operation report
type BatchSuccessReport struct {
	TotalOperations      int                     `json:"total_operations"`
	SuccessfulOperations []BatchOperationSummary `json:"successful_operations"`
	Summary              BatchOperationsSummary  `json:"summary"`
}

// BatchOperationSummary represents a summary of a single successful operation
type BatchOperationSummary struct {
	TestID        string `json:"test_id"`
	TestName      string `json:"test_name"`
	OperationType string `json:"operation_type"`
	Reason        string `json:"reason,omitempty"`
}

// BatchOperationsSummary provides counts by operation type
type BatchOperationsSummary struct {
	PassedTests    int `json:"passed_tests"`
	FailedTests    int `json:"failed_tests"`
	CancelledTests int `json:"cancelled_tests"`
}
