package tests

import (
	"strings"
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
)

func TestTestValidationService_ValidateTestStatusTransition(t *testing.T) {
	tvs := NewTestValidationService()

	tests := []struct {
		name          string
		currentStatus epic.TestStatus
		targetStatus  epic.TestStatus
		wantError     bool
	}{
		// Valid transitions from pending
		{
			name:          "pending to wip should be valid",
			currentStatus: epic.TestStatusPending,
			targetStatus:  epic.TestStatusWIP,
			wantError:     false,
		},
		{
			name:          "pending to cancelled should be valid",
			currentStatus: epic.TestStatusPending,
			targetStatus:  epic.TestStatusCancelled,
			wantError:     false,
		},
		// Valid transitions from wip
		{
			name:          "wip to done should be valid",
			currentStatus: epic.TestStatusWIP,
			targetStatus:  epic.TestStatusDone,
			wantError:     false,
		},
		{
			name:          "wip to cancelled should be valid",
			currentStatus: epic.TestStatusWIP,
			targetStatus:  epic.TestStatusCancelled,
			wantError:     false,
		},
		// Valid transitions from done
		{
			name:          "done to wip should be valid for failing tests",
			currentStatus: epic.TestStatusDone,
			targetStatus:  epic.TestStatusWIP,
			wantError:     false,
		},
		// Invalid transitions
		{
			name:          "pending to done should be invalid",
			currentStatus: epic.TestStatusPending,
			targetStatus:  epic.TestStatusDone,
			wantError:     true,
		},
		{
			name:          "done to cancelled should be invalid",
			currentStatus: epic.TestStatusDone,
			targetStatus:  epic.TestStatusCancelled,
			wantError:     true,
		},
		{
			name:          "cancelled to any status should be invalid",
			currentStatus: epic.TestStatusCancelled,
			targetStatus:  epic.TestStatusWIP,
			wantError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tvs.ValidateTestStatusTransition(tt.currentStatus, tt.targetStatus)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestTestValidationService_ValidateFailingTestCannotBeDone(t *testing.T) {
	tvs := NewTestValidationService()

	t.Run("failing tests cannot be marked as done", func(t *testing.T) {
		test := &epic.Test{
			ID:         "test1",
			Name:       "Test 1",
			TestStatus: epic.TestStatusDone,
			TestResult: epic.TestResultFailing,
		}

		err := tvs.ValidateFailingTestCannotBeDone(test)
		if err == nil {
			t.Error("Expected error but got none")
			return
		}

		statusErr, ok := err.(*epic.StatusValidationError)
		if !ok {
			t.Errorf("Expected StatusValidationError, got %T", err)
			return
		}

		if statusErr.EntityType != "test" {
			t.Errorf("Expected EntityType 'test', got '%s'", statusErr.EntityType)
		}
		if statusErr.EntityID != test.ID {
			t.Errorf("Expected EntityID '%s', got '%s'", test.ID, statusErr.EntityID)
		}
	})

	t.Run("passing tests can be marked as done", func(t *testing.T) {
		test := &epic.Test{
			ID:         "test1",
			TestStatus: epic.TestStatusDone,
			TestResult: epic.TestResultPassing,
		}

		err := tvs.ValidateFailingTestCannotBeDone(test)
		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
	})

	t.Run("wip tests are allowed regardless of result", func(t *testing.T) {
		test := &epic.Test{
			ID:         "test1",
			TestStatus: epic.TestStatusWIP,
			TestResult: epic.TestResultFailing,
		}

		err := tvs.ValidateFailingTestCannotBeDone(test)
		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
	})

	t.Run("nil test should fail", func(t *testing.T) {
		err := tvs.ValidateFailingTestCannotBeDone(nil)
		if err == nil {
			t.Error("Expected error but got none")
		}
	})
}

func TestTestValidationService_ValidateTestCancellation(t *testing.T) {
	tvs := NewTestValidationService()

	tests := []struct {
		name      string
		test      *epic.Test
		reason    string
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid cancellation with reason should pass",
			test: &epic.Test{
				ID:         "test1",
				TestStatus: epic.TestStatusWIP,
			},
			reason:    "No longer needed",
			wantError: false,
		},
		{
			name: "cancellation without reason should fail",
			test: &epic.Test{
				ID:         "test1",
				TestStatus: epic.TestStatusWIP,
			},
			reason:    "",
			wantError: true,
			errorMsg:  "cancellation reason is required",
		},
		{
			name: "cannot cancel already cancelled test",
			test: &epic.Test{
				ID:         "test1",
				TestStatus: epic.TestStatusCancelled,
			},
			reason:    "Some reason",
			wantError: true,
			errorMsg:  "already cancelled",
		},
		{
			name: "cannot cancel completed test",
			test: &epic.Test{
				ID:         "test1",
				TestStatus: epic.TestStatusDone,
			},
			reason:    "Some reason",
			wantError: true,
			errorMsg:  "cannot cancel completed test",
		},
		{
			name:      "nil test should fail",
			test:      nil,
			reason:    "Some reason",
			wantError: true,
			errorMsg:  "cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tvs.ValidateTestCancellation(tt.test, tt.reason)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestTestValidationService_ValidateTestActivePhase(t *testing.T) {
	tvs := NewTestValidationService()

	t.Run("test in active phase should pass", func(t *testing.T) {
		epicData := &epic.Epic{
			Phases: []epic.Phase{
				{ID: "phase1", Status: epic.StatusWIP},
				{ID: "phase2", Status: epic.StatusPending},
			},
		}
		test := &epic.Test{
			ID:      "test1",
			PhaseID: "phase1",
		}

		err := tvs.ValidateTestActivePhase(epicData, test)
		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
	})

	t.Run("test not in active phase should fail", func(t *testing.T) {
		epicData := &epic.Epic{
			Phases: []epic.Phase{
				{ID: "phase1", Status: epic.StatusWIP},
				{ID: "phase2", Status: epic.StatusPending},
			},
		}
		test := &epic.Test{
			ID:      "test1",
			PhaseID: "phase2",
		}

		err := tvs.ValidateTestActivePhase(epicData, test)
		if err == nil {
			t.Error("Expected error but got none")
		}
	})

	t.Run("no active phase should fail", func(t *testing.T) {
		epicData := &epic.Epic{
			Phases: []epic.Phase{
				{ID: "phase1", Status: epic.StatusPending},
				{ID: "phase2", Status: epic.StatusCompleted},
			},
		}
		test := &epic.Test{
			ID:      "test1",
			PhaseID: "phase1",
		}

		err := tvs.ValidateTestActivePhase(epicData, test)
		if err == nil {
			t.Error("Expected error but got none")
		}
		if !strings.Contains(err.Error(), "no active phase found") {
			t.Errorf("Expected error about no active phase, got: %v", err)
		}
	})

	t.Run("nil test should fail", func(t *testing.T) {
		epicData := &epic.Epic{}
		err := tvs.ValidateTestActivePhase(epicData, nil)
		if err == nil {
			t.Error("Expected error but got none")
		}
	})
}

func TestTestValidationService_ValidateTestStateConsistency(t *testing.T) {
	tvs := NewTestValidationService()
	now := time.Now()

	tests := []struct {
		name      string
		test      *epic.Test
		wantError bool
		errorMsg  string
	}{
		{
			name: "consistent test state should pass",
			test: &epic.Test{
				ID:         "test1",
				TestStatus: epic.TestStatusDone,
				TestResult: epic.TestResultPassing,
				PassedAt:   &now,
			},
			wantError: false,
		},
		{
			name: "pending test with result should fail",
			test: &epic.Test{
				ID:         "test1",
				TestStatus: epic.TestStatusPending,
				TestResult: epic.TestResultPassing,
			},
			wantError: true,
			errorMsg:  "pending tests should not have results",
		},
		{
			name: "done test with failing result should fail",
			test: &epic.Test{
				ID:         "test1",
				TestStatus: epic.TestStatusDone,
				TestResult: epic.TestResultFailing,
			},
			wantError: true,
			errorMsg:  "done tests cannot have failing results",
		},
		{
			name:      "nil test should fail",
			test:      nil,
			wantError: true,
			errorMsg:  "cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tvs.ValidateTestStateConsistency(tt.test)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}
