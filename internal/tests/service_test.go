package tests

import (
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
)

func TestNewTestService(t *testing.T) {
	cfg := ServiceConfig{
		UseMemory: true,
	}

	service := NewTestService(cfg)

	if service == nil {
		t.Fatal("NewTestService returned nil")
	}

	if service.storage == nil {
		t.Fatal("TestService storage is nil")
	}

	if service.timeSource == nil {
		t.Fatal("TestService timeSource is nil")
	}
}

func TestStartTest_Success(t *testing.T) {
	// Setup
	service, epicFile := setupTestService(t)
	testID := "test_1"

	// Create epic with a pending test using legacy Status
	e := createTestEpic()
	e.Tests = []epic.Test{
		{
			ID:      testID,
			TaskID:  "task_1",
			PhaseID: "phase_1",
			Status:  epic.StatusPending, // Use legacy status
		},
	}
	e.Tasks = []epic.Task{
		{ID: "task_1", PhaseID: "phase_1", Status: epic.StatusActive},
	}
	e.Phases = []epic.Phase{
		{ID: "phase_1", Status: epic.StatusActive},
	}

	err := service.storage.SaveEpic(e, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Test
	result, err := service.StartTest(epicFile, testID, nil)

	// Verify
	if err != nil {
		t.Fatalf("StartTest failed: %v", err)
	}

	if result.TestID != testID {
		t.Errorf("Expected TestID %s, got %s", testID, result.TestID)
	}

	if result.Operation != "started" {
		t.Errorf("Expected operation 'started', got %s", result.Operation)
	}

	if result.Status != string(epic.TestStatusWIP) {
		t.Errorf("Expected status %s, got %s", epic.TestStatusWIP, result.Status)
	}

	// Verify epic was updated
	updatedEpic, err := service.storage.LoadEpic(epicFile)
	if err != nil {
		t.Fatalf("Failed to load updated epic: %v", err)
	}

	var updatedTest *epic.Test
	for i := range updatedEpic.Tests {
		if updatedEpic.Tests[i].ID == testID {
			updatedTest = &updatedEpic.Tests[i]
			break
		}
	}

	if updatedTest == nil {
		t.Fatal("Test not found in updated epic")
	}

	// Check that both TestStatus and Status were updated
	if updatedTest.TestStatus != epic.TestStatusWIP {
		t.Errorf("Expected TestStatus %s, got %s", epic.TestStatusWIP, updatedTest.TestStatus)
	}

	if updatedTest.Status != epic.StatusActive {
		t.Errorf("Expected legacy Status %s, got %s", epic.StatusActive, updatedTest.Status)
	}

	if updatedTest.StartedAt == nil {
		t.Error("Expected StartedAt to be set")
	}
}

func TestPassTest_Success(t *testing.T) {
	// Setup
	service, epicFile := setupTestService(t)
	testID := "test_1"

	// Create epic with a test in WIP status using new TestStatus
	e := createTestEpic()
	startTime := time.Date(2025, 8, 16, 14, 30, 0, 0, time.UTC)
	e.Tests = []epic.Test{
		{
			ID:         testID,
			TaskID:     "task_1",
			PhaseID:    "phase_1",
			Status:     epic.StatusActive,  // Legacy status
			TestStatus: epic.TestStatusWIP, // New test status
			StartedAt:  &startTime,
		},
	}

	err := service.storage.SaveEpic(e, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Test
	result, err := service.PassTest(epicFile, testID, nil)

	// Verify
	if err != nil {
		t.Fatalf("PassTest failed: %v", err)
	}

	if result.TestID != testID {
		t.Errorf("Expected TestID %s, got %s", testID, result.TestID)
	}

	if result.Operation != "passed" {
		t.Errorf("Expected operation 'passed', got %s", result.Operation)
	}

	if result.Status != string(epic.TestStatusDone) {
		t.Errorf("Expected status %s, got %s", epic.TestStatusDone, result.Status)
	}

	// Verify epic was updated
	updatedEpic, err := service.storage.LoadEpic(epicFile)
	if err != nil {
		t.Fatalf("Failed to load updated epic: %v", err)
	}

	var updatedTest *epic.Test
	for i := range updatedEpic.Tests {
		if updatedEpic.Tests[i].ID == testID {
			updatedTest = &updatedEpic.Tests[i]
			break
		}
	}

	if updatedTest == nil {
		t.Fatal("Test not found in updated epic")
	}

	if updatedTest.TestStatus != epic.TestStatusDone {
		t.Errorf("Expected TestStatus %s, got %s", epic.TestStatusDone, updatedTest.TestStatus)
	}

	if updatedTest.Status != epic.StatusCompleted {
		t.Errorf("Expected legacy Status %s, got %s", epic.StatusCompleted, updatedTest.Status)
	}

	if updatedTest.PassedAt == nil {
		t.Error("Expected PassedAt to be set")
	}

	if updatedTest.FailureNote != "" {
		t.Error("Expected FailureNote to be cleared")
	}
}

func TestGetTestStatus_PreferNewStatus(t *testing.T) {
	service := &TestService{}

	// Test that TestStatus takes precedence over Status
	test := &epic.Test{
		Status:     epic.StatusPending,
		TestStatus: epic.TestStatusWIP,
	}

	result := service.getTestStatus(test)
	if result != epic.TestStatusWIP {
		t.Errorf("Expected %s, got %s", epic.TestStatusWIP, result)
	}
}

func TestGetTestStatus_FallbackToLegacy(t *testing.T) {
	service := &TestService{}

	// Test fallback to Status conversion when TestStatus is empty
	test := &epic.Test{
		Status:     epic.StatusActive,
		TestStatus: "", // Empty
	}

	result := service.getTestStatus(test)
	if result != epic.TestStatusWIP {
		t.Errorf("Expected %s, got %s", epic.TestStatusWIP, result)
	}
}

// Helper functions
func setupTestService(t *testing.T) (*TestService, string) {
	t.Helper()

	cfg := ServiceConfig{
		UseMemory: true,
		TimeSource: func() time.Time {
			return time.Date(2025, 8, 16, 14, 30, 0, 0, time.UTC)
		},
	}

	service := NewTestService(cfg)
	epicFile := "test-epic.xml"

	return service, epicFile
}

func createTestEpic() *epic.Epic {
	return &epic.Epic{
		ID:     "test-epic",
		Name:   "Test Epic",
		Status: epic.StatusActive,
		Tests:  []epic.Test{},
		Tasks:  []epic.Task{},
		Phases: []epic.Phase{},
		Events: []epic.Event{},
	}
}

// TestFailTest_Success covers AC-3: Fail Test with Details
func TestFailTest_Success(t *testing.T) {
	service, epicFile := setupTestService(t)
	testID := "test_1"
	failureReason := "Mobile responsive design not working"

	// Create epic with a test in WIP status
	e := createTestEpic()
	startTime := time.Date(2025, 8, 16, 14, 30, 0, 0, time.UTC)
	e.Tests = []epic.Test{
		{
			ID:         testID,
			TaskID:     "task_1",
			PhaseID:    "phase_1",
			Status:     epic.StatusActive,
			TestStatus: epic.TestStatusWIP,
			StartedAt:  &startTime,
		},
	}

	err := service.storage.SaveEpic(e, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Test
	result, err := service.FailTest(epicFile, testID, failureReason, nil)

	// Verify
	if err != nil {
		t.Fatalf("FailTest failed: %v", err)
	}

	if result.TestID != testID {
		t.Errorf("Expected TestID %s, got %s", testID, result.TestID)
	}

	if result.Operation != "failed" {
		t.Errorf("Expected operation 'failed', got %s", result.Operation)
	}

	if result.Status != string(epic.TestStatusWIP) {
		t.Errorf("Expected status %s, got %s", epic.TestStatusWIP, result.Status)
	}

	if result.FailureReason != failureReason {
		t.Errorf("Expected FailureReason '%s', got '%s'", failureReason, result.FailureReason)
	}

	// Verify epic was updated
	updatedEpic, err := service.storage.LoadEpic(epicFile)
	if err != nil {
		t.Fatalf("Failed to load updated epic: %v", err)
	}

	var updatedTest *epic.Test
	for i := range updatedEpic.Tests {
		if updatedEpic.Tests[i].ID == testID {
			updatedTest = &updatedEpic.Tests[i]
			break
		}
	}

	if updatedTest == nil {
		t.Fatal("Test not found in updated epic")
	}

	if updatedTest.TestStatus != epic.TestStatusWIP {
		t.Errorf("Expected TestStatus %s, got %s", epic.TestStatusWIP, updatedTest.TestStatus)
	}

	// Epic 13: Failed tests stay in WIP status, which maps to active in legacy Status
	if updatedTest.Status != epic.StatusActive {
		t.Errorf("Expected legacy Status %s, got %s", epic.StatusActive, updatedTest.Status)
	}

	if updatedTest.FailedAt == nil {
		t.Error("Expected FailedAt to be set")
	}

	if updatedTest.FailureNote != failureReason {
		t.Errorf("Expected FailureNote '%s', got '%s'", failureReason, updatedTest.FailureNote)
	}
}

// TestCancelTest_Success covers AC-4: Cancel Test with Reason
func TestCancelTest_Success(t *testing.T) {
	service, epicFile := setupTestService(t)
	testID := "test_1"
	cancellationReason := "Spec contradicts itself with point xyz"

	// Create epic with a test in WIP status
	e := createTestEpic()
	startTime := time.Date(2025, 8, 16, 14, 30, 0, 0, time.UTC)
	e.Tests = []epic.Test{
		{
			ID:         testID,
			TaskID:     "task_1",
			PhaseID:    "phase_1",
			Status:     epic.StatusActive,
			TestStatus: epic.TestStatusWIP,
			StartedAt:  &startTime,
		},
	}

	err := service.storage.SaveEpic(e, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Test
	result, err := service.CancelTest(epicFile, testID, cancellationReason, nil)

	// Verify
	if err != nil {
		t.Fatalf("CancelTest failed: %v", err)
	}

	if result.TestID != testID {
		t.Errorf("Expected TestID %s, got %s", testID, result.TestID)
	}

	if result.Operation != "cancelled" {
		t.Errorf("Expected operation 'cancelled', got %s", result.Operation)
	}

	if result.Status != string(epic.TestStatusCancelled) {
		t.Errorf("Expected status %s, got %s", epic.TestStatusCancelled, result.Status)
	}

	if result.CancellationReason != cancellationReason {
		t.Errorf("Expected CancellationReason '%s', got '%s'", cancellationReason, result.CancellationReason)
	}

	// Verify epic was updated
	updatedEpic, err := service.storage.LoadEpic(epicFile)
	if err != nil {
		t.Fatalf("Failed to load updated epic: %v", err)
	}

	var updatedTest *epic.Test
	for i := range updatedEpic.Tests {
		if updatedEpic.Tests[i].ID == testID {
			updatedTest = &updatedEpic.Tests[i]
			break
		}
	}

	if updatedTest == nil {
		t.Fatal("Test not found in updated epic")
	}

	if updatedTest.TestStatus != epic.TestStatusCancelled {
		t.Errorf("Expected TestStatus %s, got %s", epic.TestStatusCancelled, updatedTest.TestStatus)
	}

	if updatedTest.Status != epic.StatusCancelled {
		t.Errorf("Expected legacy Status %s, got %s", epic.StatusCancelled, updatedTest.Status)
	}

	if updatedTest.CancelledAt == nil {
		t.Error("Expected CancelledAt to be set")
	}

	if updatedTest.CancellationReason != cancellationReason {
		t.Errorf("Expected CancellationReason '%s', got '%s'", cancellationReason, updatedTest.CancellationReason)
	}
}

// TestStartTest_AlreadyStarted tests behavior when starting a test that's already in WIP status
func TestStartTest_AlreadyStarted(t *testing.T) {
	service, epicFile := setupTestService(t)
	testID := "test_1"

	// Create epic with a test already in WIP status (should return "already started")
	e := createTestEpic()
	e.Tests = []epic.Test{
		{
			ID:         testID,
			TaskID:     "task_1",
			PhaseID:    "phase_1",
			Status:     epic.StatusActive,
			TestStatus: epic.TestStatusWIP,
		},
	}

	err := service.storage.SaveEpic(e, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Test
	result, err := service.StartTest(epicFile, testID, nil)

	// Verify no error and "already started" result
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Operation != "start" {
		t.Errorf("Expected operation 'start', got '%s'", result.Operation)
	}

	if result.Status != "already_started" {
		t.Errorf("Expected status 'already_started', got '%s'", result.Status)
	}

	if result.TestID != testID {
		t.Errorf("Expected TestID '%s', got '%s'", testID, result.TestID)
	}
}

// TestPassTest_InvalidTransition tests error handling for invalid pass transition
func TestPassTest_InvalidTransition(t *testing.T) {
	service, epicFile := setupTestService(t)
	testID := "test_1"

	// Create epic with a test in pending status (can't pass without starting)
	e := createTestEpic()
	e.Tests = []epic.Test{
		{
			ID:         testID,
			TaskID:     "task_1",
			PhaseID:    "phase_1",
			Status:     epic.StatusPending,
			TestStatus: epic.TestStatusPending,
		},
	}

	err := service.storage.SaveEpic(e, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Test
	_, err = service.PassTest(epicFile, testID, nil)

	// Verify error
	if err == nil {
		t.Fatal("Expected error for invalid transition, got nil")
	}

	expectedErr := "Cannot pass test test_1: test is not currently in progress"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

// TestTestNotFound tests error handling when test doesn't exist
func TestTestNotFound(t *testing.T) {
	service, epicFile := setupTestService(t)
	nonExistentTestID := "nonexistent"

	// Create epic without the test
	e := createTestEpic()
	err := service.storage.SaveEpic(e, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Test StartTest
	_, err = service.StartTest(epicFile, nonExistentTestID, nil)
	if err == nil {
		t.Error("Expected error for nonexistent test in StartTest, got nil")
	}

	// Test PassTest
	_, err = service.PassTest(epicFile, nonExistentTestID, nil)
	if err == nil {
		t.Error("Expected error for nonexistent test in PassTest, got nil")
	}

	// Test FailTest
	_, err = service.FailTest(epicFile, nonExistentTestID, "reason", nil)
	if err == nil {
		t.Error("Expected error for nonexistent test in FailTest, got nil")
	}

	// Test CancelTest
	_, err = service.CancelTest(epicFile, nonExistentTestID, "reason", nil)
	if err == nil {
		t.Error("Expected error for nonexistent test in CancelTest, got nil")
	}
}

// TestCustomTimestamp tests that custom timestamps are properly used
func TestCustomTimestamp(t *testing.T) {
	service, epicFile := setupTestService(t)
	testID := "test_1"
	customTime := time.Date(2025, 12, 25, 10, 30, 0, 0, time.UTC)

	// Create epic with a pending test
	e := createTestEpic()
	e.Tests = []epic.Test{
		{
			ID:         testID,
			TaskID:     "task_1",
			PhaseID:    "phase_1",
			Status:     epic.StatusPending,
			TestStatus: epic.TestStatusPending,
		},
	}
	e.Tasks = []epic.Task{
		{ID: "task_1", PhaseID: "phase_1", Status: epic.StatusActive},
	}
	e.Phases = []epic.Phase{
		{ID: "phase_1", Status: epic.StatusActive},
	}

	err := service.storage.SaveEpic(e, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Test with custom timestamp
	_, err = service.StartTest(epicFile, testID, &customTime)
	if err != nil {
		t.Fatalf("StartTest failed: %v", err)
	}

	// Verify custom timestamp was used
	updatedEpic, err := service.storage.LoadEpic(epicFile)
	if err != nil {
		t.Fatalf("Failed to load updated epic: %v", err)
	}

	var updatedTest *epic.Test
	for i := range updatedEpic.Tests {
		if updatedEpic.Tests[i].ID == testID {
			updatedTest = &updatedEpic.Tests[i]
			break
		}
	}

	if updatedTest == nil {
		t.Fatal("Test not found in updated epic")
	}

	if updatedTest.StartedAt == nil {
		t.Fatal("Expected StartedAt to be set")
	}

	if !updatedTest.StartedAt.Equal(customTime) {
		t.Errorf("Expected custom timestamp %v, got %v", customTime, *updatedTest.StartedAt)
	}
}

// TestPassedToFailedTransition tests that passed tests can be failed (Epic 4 spec)
func TestPassedToFailedTransition(t *testing.T) {
	service, epicFile := setupTestService(t)
	testID := "test_1"
	failureReason := "Found regression issue"

	// Create epic with a test in passed status
	passedTime := time.Date(2025, 8, 16, 14, 30, 0, 0, time.UTC)
	e := createTestEpic()
	e.Tests = []epic.Test{
		{
			ID:         testID,
			TaskID:     "task_1",
			PhaseID:    "phase_1",
			Status:     epic.StatusCompleted,
			TestStatus: epic.TestStatusDone,
			PassedAt:   &passedTime,
		},
	}

	err := service.storage.SaveEpic(e, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Test failing a passed test
	result, err := service.FailTest(epicFile, testID, failureReason, nil)

	// Verify
	if err != nil {
		t.Fatalf("FailTest failed: %v", err)
	}

	if result.Status != string(epic.TestStatusWIP) {
		t.Errorf("Expected status %s, got %s", epic.TestStatusWIP, result.Status)
	}

	// Verify epic was updated
	updatedEpic, err := service.storage.LoadEpic(epicFile)
	if err != nil {
		t.Fatalf("Failed to load updated epic: %v", err)
	}

	var updatedTest *epic.Test
	for i := range updatedEpic.Tests {
		if updatedEpic.Tests[i].ID == testID {
			updatedTest = &updatedEpic.Tests[i]
			break
		}
	}

	if updatedTest == nil {
		t.Fatal("Test not found in updated epic")
	}

	if updatedTest.TestStatus != epic.TestStatusWIP {
		t.Errorf("Expected TestStatus %s, got %s", epic.TestStatusWIP, updatedTest.TestStatus)
	}

	if updatedTest.FailureNote != failureReason {
		t.Errorf("Expected FailureNote '%s', got '%s'", failureReason, updatedTest.FailureNote)
	}

	// PassedAt is not cleared when test fails (keeping historical data)
	if updatedTest.PassedAt == nil {
		t.Error("Expected PassedAt to be preserved as historical data")
	}
}

// TestFailedToPassedTransition tests that failed tests can be passed (Epic 4 spec)
func TestFailedToPassedTransition(t *testing.T) {
	service, epicFile := setupTestService(t)
	testID := "test_1"

	// Create epic with a test in failed status
	failedTime := time.Date(2025, 8, 16, 14, 30, 0, 0, time.UTC)
	e := createTestEpic()
	e.Tests = []epic.Test{
		{
			ID:          testID,
			TaskID:      "task_1",
			PhaseID:     "phase_1",
			Status:      epic.StatusCancelled,
			TestStatus:  epic.TestStatusWIP,
			FailedAt:    &failedTime,
			FailureNote: "Previous failure reason",
		},
	}

	err := service.storage.SaveEpic(e, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Test passing a failed test
	result, err := service.PassTest(epicFile, testID, nil)

	// Verify
	if err != nil {
		t.Fatalf("PassTest failed: %v", err)
	}

	if result.Status != string(epic.TestStatusDone) {
		t.Errorf("Expected status %s, got %s", epic.TestStatusDone, result.Status)
	}

	// Verify epic was updated
	updatedEpic, err := service.storage.LoadEpic(epicFile)
	if err != nil {
		t.Fatalf("Failed to load updated epic: %v", err)
	}

	var updatedTest *epic.Test
	for i := range updatedEpic.Tests {
		if updatedEpic.Tests[i].ID == testID {
			updatedTest = &updatedEpic.Tests[i]
			break
		}
	}

	if updatedTest == nil {
		t.Fatal("Test not found in updated epic")
	}

	if updatedTest.TestStatus != epic.TestStatusDone {
		t.Errorf("Expected TestStatus %s, got %s", epic.TestStatusDone, updatedTest.TestStatus)
	}

	// FailureNote should be cleared when test passes
	if updatedTest.FailureNote != "" {
		t.Errorf("Expected FailureNote to be cleared, got '%s'", updatedTest.FailureNote)
	}

	// FailedAt is preserved as historical data when test passes
	if updatedTest.FailedAt == nil {
		t.Error("Expected FailedAt to be preserved as historical data")
	}

	// PassedAt should be set
	if updatedTest.PassedAt == nil {
		t.Error("Expected PassedAt to be set")
	}
}

// TestPrerequisiteValidation tests test prerequisite validation
func TestPrerequisiteValidation(t *testing.T) {
	service, epicFile := setupTestService(t)
	testID := "test_1"

	// Create epic with test whose task is not active/completed
	e := createTestEpic()
	e.Tests = []epic.Test{
		{
			ID:         testID,
			TaskID:     "task_1",
			PhaseID:    "phase_1",
			Status:     epic.StatusPending,
			TestStatus: epic.TestStatusPending,
		},
	}
	e.Tasks = []epic.Task{
		{ID: "task_1", PhaseID: "phase_1", Status: epic.StatusPending}, // Not active
	}
	e.Phases = []epic.Phase{
		{ID: "phase_1", Status: epic.StatusActive},
	}

	err := service.storage.SaveEpic(e, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Test - should fail prerequisite validation
	_, err = service.StartTest(epicFile, testID, nil)

	// Verify error
	if err == nil {
		t.Fatal("Expected error for prerequisite validation, got nil")
	}

	expectedErr := "Cannot start test test_1: associated task task_1 is not active or completed (status: pending)"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

// TestMultipleTestsInProgress tests that multiple tests can be WIP simultaneously
func TestMultipleTestsInProgress(t *testing.T) {
	service, epicFile := setupTestService(t)

	// Create epic with multiple tests
	e := createTestEpic()
	e.Tests = []epic.Test{
		{
			ID:         "test_1",
			TaskID:     "task_1",
			PhaseID:    "phase_1",
			Status:     epic.StatusPending,
			TestStatus: epic.TestStatusPending,
		},
		{
			ID:         "test_2",
			TaskID:     "task_1",
			PhaseID:    "phase_1",
			Status:     epic.StatusPending,
			TestStatus: epic.TestStatusPending,
		},
	}
	e.Tasks = []epic.Task{
		{ID: "task_1", PhaseID: "phase_1", Status: epic.StatusActive},
	}
	e.Phases = []epic.Phase{
		{ID: "phase_1", Status: epic.StatusActive},
	}

	err := service.storage.SaveEpic(e, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Start both tests
	_, err = service.StartTest(epicFile, "test_1", nil)
	if err != nil {
		t.Fatalf("Failed to start test_1: %v", err)
	}

	_, err = service.StartTest(epicFile, "test_2", nil)
	if err != nil {
		t.Fatalf("Failed to start test_2: %v", err)
	}

	// Verify both tests are in WIP
	updatedEpic, err := service.storage.LoadEpic(epicFile)
	if err != nil {
		t.Fatalf("Failed to load updated epic: %v", err)
	}

	wipCount := 0
	for _, test := range updatedEpic.Tests {
		if service.getTestStatus(&test) == epic.TestStatusWIP {
			wipCount++
		}
	}

	if wipCount != 2 {
		t.Errorf("Expected 2 tests in WIP status, got %d", wipCount)
	}
}

// TestStatusTransitionValidation tests that status transition validation works via CanTransitionTo
func TestStatusTransitionValidation(t *testing.T) {
	testCases := []struct {
		name          string
		current       epic.TestStatus
		target        epic.TestStatus
		shouldSucceed bool
	}{
		// Epic 13 valid transitions
		{"pending to wip", epic.TestStatusPending, epic.TestStatusWIP, true},
		{"pending to cancelled", epic.TestStatusPending, epic.TestStatusCancelled, true},
		{"wip to done", epic.TestStatusWIP, epic.TestStatusDone, true},
		{"wip to cancelled", epic.TestStatusWIP, epic.TestStatusCancelled, true},
		{"done to wip", epic.TestStatusDone, epic.TestStatusWIP, true}, // Can go back to WIP for failing tests

		// Epic 13 invalid transitions
		{"pending to done", epic.TestStatusPending, epic.TestStatusDone, false},     // Must go through wip
		{"cancelled to wip", epic.TestStatusCancelled, epic.TestStatusWIP, false},   // Cancelled is terminal
		{"cancelled to done", epic.TestStatusCancelled, epic.TestStatusDone, false}, // Cancelled is terminal
		{"wip to pending", epic.TestStatusWIP, epic.TestStatusPending, false},       // Can't go backwards to pending
		{"done to pending", epic.TestStatusDone, epic.TestStatusPending, false},     // Can't go backwards to pending
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			canTransition := tc.current.CanTransitionTo(tc.target)

			if tc.shouldSucceed && !canTransition {
				t.Errorf("Expected transition to be allowed, but CanTransitionTo returned false")
			}

			if !tc.shouldSucceed && canTransition {
				t.Errorf("Expected transition to be forbidden, but CanTransitionTo returned true")
			}
		})
	}
}

func TestTestServiceEventCreation(t *testing.T) {
	// Fixed timestamp for deterministic testing
	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	timeSource := func() time.Time { return fixedTime }

	service := NewTestService(ServiceConfig{
		UseMemory:  true,
		TimeSource: timeSource,
	})

	// Create a test epic with phases, tasks, and tests
	epicData := &epic.Epic{
		ID: "test-epic",
		Phases: []epic.Phase{
			{ID: "phase1", Name: "Test Phase", Status: epic.StatusActive},
		},
		Tasks: []epic.Task{
			{ID: "task1", PhaseID: "phase1", Name: "Test Task", Status: epic.StatusActive},
		},
		Tests: []epic.Test{
			{ID: "test1", PhaseID: "phase1", TaskID: "task1", Name: "Test 1", TestStatus: epic.TestStatusPending},
		},
	}

	epicFile := "test-epic.xml"
	service.storage.SaveEpic(epicData, epicFile)

	t.Run("StartTest creates event", func(t *testing.T) {
		result, err := service.StartTest(epicFile, "test1", nil)
		if err != nil {
			t.Fatalf("StartTest failed: %v", err)
		}

		if result.Status != string(epic.TestStatusWIP) {
			t.Errorf("Expected test status WIP, got %s", result.Status)
		}

		// Load epic and verify event was created
		updatedEpic, _ := service.storage.LoadEpic(epicFile)
		if len(updatedEpic.Events) != 1 {
			t.Errorf("Expected 1 event, got %d", len(updatedEpic.Events))
		}

		event := updatedEpic.Events[0]
		if event.Type != "test_started" {
			t.Errorf("Expected event type test_started, got %s", event.Type)
		}

		if event.Data != "Test test1 (Test 1) started" {
			t.Errorf("Expected event data 'Test test1 (Test 1) started', got %s", event.Data)
		}
	})

	t.Run("PassTest creates event", func(t *testing.T) {
		// Restart the test to clear events
		service.storage.SaveEpic(&epic.Epic{
			ID: "test-epic",
			Phases: []epic.Phase{
				{ID: "phase1", Name: "Test Phase", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task1", PhaseID: "phase1", Name: "Test Task", Status: epic.StatusActive},
			},
			Tests: []epic.Test{
				{ID: "test1", PhaseID: "phase1", TaskID: "task1", Name: "Test 1", TestStatus: epic.TestStatusWIP},
			},
		}, epicFile)

		result, err := service.PassTest(epicFile, "test1", nil)
		if err != nil {
			t.Fatalf("PassTest failed: %v", err)
		}

		if result.Status != string(epic.TestStatusDone) {
			t.Errorf("Expected test status passed, got %s", result.Status)
		}

		// Load epic and verify event was created
		updatedEpic, _ := service.storage.LoadEpic(epicFile)
		if len(updatedEpic.Events) != 1 {
			t.Errorf("Expected 1 event, got %d", len(updatedEpic.Events))
		}

		event := updatedEpic.Events[0]
		if event.Type != "test_passed" {
			t.Errorf("Expected event type test_passed, got %s", event.Type)
		}

		if event.Data != "Test test1 (Test 1) passed" {
			t.Errorf("Expected event data 'Test test1 (Test 1) passed', got %s", event.Data)
		}
	})

	t.Run("FailTest creates event with reason", func(t *testing.T) {
		// Restart the test to clear events
		service.storage.SaveEpic(&epic.Epic{
			ID: "test-epic",
			Phases: []epic.Phase{
				{ID: "phase1", Name: "Test Phase", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task1", PhaseID: "phase1", Name: "Test Task", Status: epic.StatusActive},
			},
			Tests: []epic.Test{
				{ID: "test1", PhaseID: "phase1", TaskID: "task1", Name: "Test 1", TestStatus: epic.TestStatusWIP},
			},
		}, epicFile)

		failureReason := "Connection timeout"
		result, err := service.FailTest(epicFile, "test1", failureReason, nil)
		if err != nil {
			t.Fatalf("FailTest failed: %v", err)
		}

		if result.Status != string(epic.TestStatusWIP) {
			t.Errorf("Expected test status failed, got %s", result.Status)
		}

		// Load epic and verify event was created
		updatedEpic, _ := service.storage.LoadEpic(epicFile)
		if len(updatedEpic.Events) != 1 {
			t.Errorf("Expected 1 event, got %d", len(updatedEpic.Events))
		}

		event := updatedEpic.Events[0]
		if event.Type != "test_failed" {
			t.Errorf("Expected event type test_failed, got %s", event.Type)
		}

		expectedData := "Test test1 (Test 1) failed: Connection timeout"
		if event.Data != expectedData {
			t.Errorf("Expected event data '%s', got %s", expectedData, event.Data)
		}
	})

	t.Run("CancelTest creates event with reason", func(t *testing.T) {
		// Restart the test to clear events
		service.storage.SaveEpic(&epic.Epic{
			ID: "test-epic",
			Phases: []epic.Phase{
				{ID: "phase1", Name: "Test Phase", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task1", PhaseID: "phase1", Name: "Test Task", Status: epic.StatusActive},
			},
			Tests: []epic.Test{
				{ID: "test1", PhaseID: "phase1", TaskID: "task1", Name: "Test 1", TestStatus: epic.TestStatusWIP},
			},
		}, epicFile)

		cancellationReason := "Requirements changed"
		result, err := service.CancelTest(epicFile, "test1", cancellationReason, nil)
		if err != nil {
			t.Fatalf("CancelTest failed: %v", err)
		}

		if result.Status != string(epic.TestStatusCancelled) {
			t.Errorf("Expected test status cancelled, got %s", result.Status)
		}

		// Load epic and verify event was created
		updatedEpic, _ := service.storage.LoadEpic(epicFile)
		if len(updatedEpic.Events) != 1 {
			t.Errorf("Expected 1 event, got %d", len(updatedEpic.Events))
		}

		event := updatedEpic.Events[0]
		if event.Type != "test_cancelled" {
			t.Errorf("Expected event type test_cancelled, got %s", event.Type)
		}

		expectedData := "Test test1 (Test 1) cancelled: Requirements changed"
		if event.Data != expectedData {
			t.Errorf("Expected event data '%s', got %s", expectedData, event.Data)
		}
	})

	t.Run("Event timestamps match operation timestamps", func(t *testing.T) {
		// Restart the test to clear events
		service.storage.SaveEpic(&epic.Epic{
			ID: "test-epic",
			Phases: []epic.Phase{
				{ID: "phase1", Name: "Test Phase", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task1", PhaseID: "phase1", Name: "Test Task", Status: epic.StatusActive},
			},
			Tests: []epic.Test{
				{ID: "test1", PhaseID: "phase1", TaskID: "task1", Name: "Test 1", TestStatus: epic.TestStatusPending},
			},
		}, epicFile)

		customTime := time.Date(2023, 6, 15, 14, 30, 0, 0, time.UTC)
		result, err := service.StartTest(epicFile, "test1", &customTime)
		if err != nil {
			t.Fatalf("StartTest failed: %v", err)
		}

		if !result.Timestamp.Equal(customTime) {
			t.Errorf("Expected result timestamp %v, got %v", customTime, result.Timestamp)
		}

		// Load epic and verify event timestamp matches
		updatedEpic, _ := service.storage.LoadEpic(epicFile)
		if len(updatedEpic.Events) != 1 {
			t.Errorf("Expected 1 event, got %d", len(updatedEpic.Events))
		}

		event := updatedEpic.Events[0]
		if !event.Timestamp.Equal(customTime) {
			t.Errorf("Expected event timestamp %v, got %v", customTime, event.Timestamp)
		}
	})
}
