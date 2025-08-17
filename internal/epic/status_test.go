package epic

import (
	"testing"
)

// TestEpicStatus tests the EpicStatus enum and methods
func TestEpicStatus(t *testing.T) {
	t.Run("valid epic statuses", func(t *testing.T) {
		validStatuses := []EpicStatus{
			EpicStatusPending,
			EpicStatusWIP,
			EpicStatusDone,
		}

		for _, status := range validStatuses {
			if !status.IsValid() {
				t.Errorf("Expected %s to be valid", status)
			}
		}
	})

	t.Run("invalid epic status", func(t *testing.T) {
		invalidStatus := EpicStatus("invalid")
		if invalidStatus.IsValid() {
			t.Errorf("Expected %s to be invalid", invalidStatus)
		}
	})

	t.Run("epic status string conversion", func(t *testing.T) {
		testCases := map[EpicStatus]string{
			EpicStatusPending: "pending",
			EpicStatusWIP:     "wip",
			EpicStatusDone:    "done",
		}

		for status, expected := range testCases {
			if status.String() != expected {
				t.Errorf("Expected %s.String() to return %s, got %s", status, expected, status.String())
			}
		}
	})

	t.Run("epic status transitions", func(t *testing.T) {
		// Test valid transitions
		validTransitions := map[EpicStatus][]EpicStatus{
			EpicStatusPending: {EpicStatusWIP},
			EpicStatusWIP:     {EpicStatusDone},
			EpicStatusDone:    {}, // Done is terminal
		}

		for from, allowedTargets := range validTransitions {
			for _, to := range allowedTargets {
				if !from.CanTransitionTo(to) {
					t.Errorf("Expected %s to be able to transition to %s", from, to)
				}
			}
		}

		// Test invalid transitions
		invalidTransitions := []struct {
			from EpicStatus
			to   EpicStatus
		}{
			{EpicStatusPending, EpicStatusDone},
			{EpicStatusWIP, EpicStatusPending},
			{EpicStatusDone, EpicStatusWIP},
			{EpicStatusDone, EpicStatusPending},
		}

		for _, tt := range invalidTransitions {
			if tt.from.CanTransitionTo(tt.to) {
				t.Errorf("Expected %s to NOT be able to transition to %s", tt.from, tt.to)
			}
		}
	})
}

// TestPhaseStatus tests the PhaseStatus enum and methods
func TestPhaseStatus(t *testing.T) {
	t.Run("valid phase statuses", func(t *testing.T) {
		validStatuses := []PhaseStatus{
			PhaseStatusPending,
			PhaseStatusWIP,
			PhaseStatusDone,
		}

		for _, status := range validStatuses {
			if !status.IsValid() {
				t.Errorf("Expected %s to be valid", status)
			}
		}
	})

	t.Run("invalid phase status", func(t *testing.T) {
		invalidStatus := PhaseStatus("invalid")
		if invalidStatus.IsValid() {
			t.Errorf("Expected %s to be invalid", invalidStatus)
		}
	})

	t.Run("phase status string conversion", func(t *testing.T) {
		testCases := map[PhaseStatus]string{
			PhaseStatusPending: "pending",
			PhaseStatusWIP:     "wip",
			PhaseStatusDone:    "done",
		}

		for status, expected := range testCases {
			if status.String() != expected {
				t.Errorf("Expected %s.String() to return %s, got %s", status, expected, status.String())
			}
		}
	})

	t.Run("phase status transitions", func(t *testing.T) {
		// Test valid transitions
		validTransitions := map[PhaseStatus][]PhaseStatus{
			PhaseStatusPending: {PhaseStatusWIP},
			PhaseStatusWIP:     {PhaseStatusDone},
			PhaseStatusDone:    {}, // Done is terminal
		}

		for from, allowedTargets := range validTransitions {
			for _, to := range allowedTargets {
				if !from.CanTransitionTo(to) {
					t.Errorf("Expected %s to be able to transition to %s", from, to)
				}
			}
		}

		// Test invalid transitions
		invalidTransitions := []struct {
			from PhaseStatus
			to   PhaseStatus
		}{
			{PhaseStatusPending, PhaseStatusDone},
			{PhaseStatusWIP, PhaseStatusPending},
			{PhaseStatusDone, PhaseStatusWIP},
			{PhaseStatusDone, PhaseStatusPending},
		}

		for _, tt := range invalidTransitions {
			if tt.from.CanTransitionTo(tt.to) {
				t.Errorf("Expected %s to NOT be able to transition to %s", tt.from, tt.to)
			}
		}
	})
}

// TestTaskStatus tests the TaskStatus enum and methods
func TestTaskStatus(t *testing.T) {
	t.Run("valid task statuses", func(t *testing.T) {
		validStatuses := []TaskStatus{
			TaskStatusPending,
			TaskStatusWIP,
			TaskStatusDone,
			TaskStatusCancelled,
		}

		for _, status := range validStatuses {
			if !status.IsValid() {
				t.Errorf("Expected %s to be valid", status)
			}
		}
	})

	t.Run("invalid task status", func(t *testing.T) {
		invalidStatus := TaskStatus("invalid")
		if invalidStatus.IsValid() {
			t.Errorf("Expected %s to be invalid", invalidStatus)
		}
	})

	t.Run("task status string conversion", func(t *testing.T) {
		testCases := map[TaskStatus]string{
			TaskStatusPending:   "pending",
			TaskStatusWIP:       "wip",
			TaskStatusDone:      "done",
			TaskStatusCancelled: "cancelled",
		}

		for status, expected := range testCases {
			if status.String() != expected {
				t.Errorf("Expected %s.String() to return %s, got %s", status, expected, status.String())
			}
		}
	})

	t.Run("task status transitions", func(t *testing.T) {
		// Test valid transitions
		validTransitions := map[TaskStatus][]TaskStatus{
			TaskStatusPending:   {TaskStatusWIP, TaskStatusCancelled},
			TaskStatusWIP:       {TaskStatusDone, TaskStatusCancelled},
			TaskStatusDone:      {}, // Done is terminal
			TaskStatusCancelled: {}, // Cancelled is terminal
		}

		for from, allowedTargets := range validTransitions {
			for _, to := range allowedTargets {
				if !from.CanTransitionTo(to) {
					t.Errorf("Expected %s to be able to transition to %s", from, to)
				}
			}
		}

		// Test invalid transitions
		invalidTransitions := []struct {
			from TaskStatus
			to   TaskStatus
		}{
			{TaskStatusPending, TaskStatusDone},
			{TaskStatusWIP, TaskStatusPending},
			{TaskStatusDone, TaskStatusWIP},
			{TaskStatusDone, TaskStatusPending},
			{TaskStatusCancelled, TaskStatusWIP},
			{TaskStatusCancelled, TaskStatusDone},
			{TaskStatusCancelled, TaskStatusPending},
		}

		for _, tt := range invalidTransitions {
			if tt.from.CanTransitionTo(tt.to) {
				t.Errorf("Expected %s to NOT be able to transition to %s", tt.from, tt.to)
			}
		}
	})
}

// TestTestResult tests the TestResult enum and methods
func TestTestResult(t *testing.T) {
	t.Run("valid test results", func(t *testing.T) {
		validResults := []TestResult{
			TestResultPassing,
			TestResultFailing,
		}

		for _, result := range validResults {
			if !result.IsValid() {
				t.Errorf("Expected %s to be valid", result)
			}
		}
	})

	t.Run("invalid test result", func(t *testing.T) {
		invalidResult := TestResult("invalid")
		if invalidResult.IsValid() {
			t.Errorf("Expected %s to be invalid", invalidResult)
		}
	})

	t.Run("test result string conversion", func(t *testing.T) {
		testCases := map[TestResult]string{
			TestResultPassing: "passing",
			TestResultFailing: "failing",
		}

		for result, expected := range testCases {
			if result.String() != expected {
				t.Errorf("Expected %s.String() to return %s, got %s", result, expected, result.String())
			}
		}
	})
}

// TestTestStatusTransitions tests the updated TestStatus transitions according to Epic 13
func TestTestStatusTransitions(t *testing.T) {
	t.Run("valid test status transitions", func(t *testing.T) {
		// Test valid transitions according to Epic 13 specification
		validTransitions := map[TestStatus][]TestStatus{
			TestStatusPending:   {TestStatusWIP, TestStatusCancelled},
			TestStatusWIP:       {TestStatusDone, TestStatusCancelled},
			TestStatusDone:      {TestStatusWIP}, // Can go back to WIP for failing tests
			TestStatusCancelled: {},              // Cancelled is terminal
		}

		for from, allowedTargets := range validTransitions {
			for _, to := range allowedTargets {
				if !from.CanTransitionTo(to) {
					t.Errorf("Expected %s to be able to transition to %s", from, to)
				}
			}
		}
	})

	t.Run("invalid test status transitions", func(t *testing.T) {
		invalidTransitions := []struct {
			from TestStatus
			to   TestStatus
		}{
			{TestStatusPending, TestStatusDone},
			{TestStatusWIP, TestStatusPending},
			{TestStatusDone, TestStatusPending},
			{TestStatusCancelled, TestStatusWIP},
			{TestStatusCancelled, TestStatusDone},
			{TestStatusCancelled, TestStatusPending},
		}

		for _, tt := range invalidTransitions {
			if tt.from.CanTransitionTo(tt.to) {
				t.Errorf("Expected %s to NOT be able to transition to %s", tt.from, tt.to)
			}
		}
	})
}

// TestStatusValidation tests the validation functions
func TestStatusValidation(t *testing.T) {
	t.Run("validate epic status", func(t *testing.T) {
		// Valid statuses
		validCases := []string{"pending", "wip", "done"}
		for _, status := range validCases {
			result, err := ValidateEpicStatus(status)
			if err != nil {
				t.Errorf("Expected %s to be valid, got error: %v", status, err)
			}
			if result.String() != status {
				t.Errorf("Expected validated status to be %s, got %s", status, result.String())
			}
		}

		// Invalid status
		_, err := ValidateEpicStatus("invalid")
		if err == nil {
			t.Error("Expected error for invalid epic status")
		}
	})

	t.Run("validate phase status", func(t *testing.T) {
		// Valid statuses
		validCases := []string{"pending", "wip", "done"}
		for _, status := range validCases {
			result, err := ValidatePhaseStatus(status)
			if err != nil {
				t.Errorf("Expected %s to be valid, got error: %v", status, err)
			}
			if result.String() != status {
				t.Errorf("Expected validated status to be %s, got %s", status, result.String())
			}
		}

		// Invalid status
		_, err := ValidatePhaseStatus("invalid")
		if err == nil {
			t.Error("Expected error for invalid phase status")
		}
	})

	t.Run("validate task status", func(t *testing.T) {
		// Valid statuses
		validCases := []string{"pending", "wip", "done", "cancelled"}
		for _, status := range validCases {
			result, err := ValidateTaskStatus(status)
			if err != nil {
				t.Errorf("Expected %s to be valid, got error: %v", status, err)
			}
			if result.String() != status {
				t.Errorf("Expected validated status to be %s, got %s", status, result.String())
			}
		}

		// Invalid status
		_, err := ValidateTaskStatus("invalid")
		if err == nil {
			t.Error("Expected error for invalid task status")
		}
	})

	t.Run("validate test result", func(t *testing.T) {
		// Valid results
		validCases := []string{"passing", "failing"}
		for _, result := range validCases {
			validated, err := ValidateTestResult(result)
			if err != nil {
				t.Errorf("Expected %s to be valid, got error: %v", result, err)
			}
			if validated.String() != result {
				t.Errorf("Expected validated result to be %s, got %s", result, validated.String())
			}
		}

		// Invalid result
		_, err := ValidateTestResult("invalid")
		if err == nil {
			t.Error("Expected error for invalid test result")
		}
	})
}

// TestStatusEnumConstants tests that status enum constants match expected string values
func TestStatusEnumConstants(t *testing.T) {
	t.Run("epic status constants", func(t *testing.T) {
		expected := map[EpicStatus]string{
			EpicStatusPending: "pending",
			EpicStatusWIP:     "wip",
			EpicStatusDone:    "done",
		}

		for status, expectedStr := range expected {
			if string(status) != expectedStr {
				t.Errorf("Expected %s constant to have value %s, got %s", status, expectedStr, string(status))
			}
		}
	})

	t.Run("phase status constants", func(t *testing.T) {
		expected := map[PhaseStatus]string{
			PhaseStatusPending: "pending",
			PhaseStatusWIP:     "wip",
			PhaseStatusDone:    "done",
		}

		for status, expectedStr := range expected {
			if string(status) != expectedStr {
				t.Errorf("Expected %s constant to have value %s, got %s", status, expectedStr, string(status))
			}
		}
	})

	t.Run("task status constants", func(t *testing.T) {
		expected := map[TaskStatus]string{
			TaskStatusPending:   "pending",
			TaskStatusWIP:       "wip",
			TaskStatusDone:      "done",
			TaskStatusCancelled: "cancelled",
		}

		for status, expectedStr := range expected {
			if string(status) != expectedStr {
				t.Errorf("Expected %s constant to have value %s, got %s", status, expectedStr, string(status))
			}
		}
	})

	t.Run("test status constants", func(t *testing.T) {
		expected := map[TestStatus]string{
			TestStatusPending:   "pending",
			TestStatusWIP:       "wip",
			TestStatusDone:      "done",
			TestStatusCancelled: "cancelled",
		}

		for status, expectedStr := range expected {
			if string(status) != expectedStr {
				t.Errorf("Expected %s constant to have value %s, got %s", status, expectedStr, string(status))
			}
		}
	})

	t.Run("test result constants", func(t *testing.T) {
		expected := map[TestResult]string{
			TestResultPassing: "passing",
			TestResultFailing: "failing",
		}

		for result, expectedStr := range expected {
			if string(result) != expectedStr {
				t.Errorf("Expected %s constant to have value %s, got %s", result, expectedStr, string(result))
			}
		}
	})
}

// TestStatusComparison tests status enum comparison operations
func TestStatusComparison(t *testing.T) {
	t.Run("epic status equality", func(t *testing.T) {
		if EpicStatusPending != EpicStatusPending {
			t.Error("Expected EpicStatusPending to equal itself")
		}
		if EpicStatusPending == EpicStatusWIP {
			t.Error("Expected EpicStatusPending to not equal EpicStatusWIP")
		}
	})

	t.Run("phase status equality", func(t *testing.T) {
		if PhaseStatusWIP != PhaseStatusWIP {
			t.Error("Expected PhaseStatusWIP to equal itself")
		}
		if PhaseStatusWIP == PhaseStatusDone {
			t.Error("Expected PhaseStatusWIP to not equal PhaseStatusDone")
		}
	})

	t.Run("task status equality", func(t *testing.T) {
		if TaskStatusDone != TaskStatusDone {
			t.Error("Expected TaskStatusDone to equal itself")
		}
		if TaskStatusDone == TaskStatusCancelled {
			t.Error("Expected TaskStatusDone to not equal TaskStatusCancelled")
		}
	})

	t.Run("test result equality", func(t *testing.T) {
		if TestResultPassing != TestResultPassing {
			t.Error("Expected TestResultPassing to equal itself")
		}
		if TestResultPassing == TestResultFailing {
			t.Error("Expected TestResultPassing to not equal TestResultFailing")
		}
	})
}
