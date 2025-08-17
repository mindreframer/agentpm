package lifecycle

import (
	"strings"
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewLifecycleService(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)

	ls := NewLifecycleService(storage, queryService)

	if ls == nil {
		t.Error("Expected lifecycle service to be created")
	}

	if ls.storage != storage {
		t.Error("Expected storage to be injected correctly")
	}

	if ls.queryService != queryService {
		t.Error("Expected query service to be injected correctly")
	}
}

func TestLifecycleService_WithTimeSource(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	ls := NewLifecycleService(storage, queryService)

	testTime := time.Date(2025, 8, 16, 15, 30, 0, 0, time.UTC)
	timeSource := func() time.Time { return testTime }

	newLS := ls.WithTimeSource(timeSource)

	if newLS != ls {
		t.Error("Expected WithTimeSource to return the same service instance")
	}

	// Test that the time source is used
	result := ls.timeSource()
	if !result.Equal(testTime) {
		t.Errorf("Expected time source to return %v, got %v", testTime, result)
	}
}

func TestEpicLifecycleStatus_String(t *testing.T) {
	tests := []struct {
		status   EpicLifecycleStatus
		expected string
	}{
		{LifecycleStatusPending, "pending"},
		{LifecycleStatusWIP, "wip"},
		{LifecycleStatusDone, "done"},
	}

	for _, test := range tests {
		if test.status.String() != test.expected {
			t.Errorf("Expected %s.String() to return %s, got %s",
				test.status, test.expected, test.status.String())
		}
	}
}

func TestEpicLifecycleStatus_IsValid(t *testing.T) {
	validStatuses := []EpicLifecycleStatus{
		LifecycleStatusPending,
		LifecycleStatusWIP,
		LifecycleStatusDone,
	}

	for _, status := range validStatuses {
		if !status.IsValid() {
			t.Errorf("Expected %s to be valid", status)
		}
	}

	invalidStatus := EpicLifecycleStatus("invalid")
	if invalidStatus.IsValid() {
		t.Error("Expected 'invalid' status to be invalid")
	}
}

func TestEpicLifecycleStatus_ToEpicStatus(t *testing.T) {
	tests := []struct {
		lifecycle EpicLifecycleStatus
		expected  epic.Status
	}{
		{LifecycleStatusPending, epic.StatusPlanning},
		{LifecycleStatusWIP, epic.StatusActive},
		{LifecycleStatusDone, epic.StatusCompleted},
	}

	for _, test := range tests {
		result := test.lifecycle.ToEpicStatus()
		if result != test.expected {
			t.Errorf("Expected %s.ToEpicStatus() to return %s, got %s",
				test.lifecycle, test.expected, result)
		}
	}
}

func TestFromEpicStatus(t *testing.T) {
	tests := []struct {
		epicStatus epic.Status
		expected   EpicLifecycleStatus
	}{
		{epic.StatusPlanning, LifecycleStatusPending},
		{epic.StatusActive, LifecycleStatusWIP},
		{epic.StatusCompleted, LifecycleStatusDone},
		{epic.StatusOnHold, LifecycleStatusPending}, // default case
	}

	for _, test := range tests {
		result := FromEpicStatus(test.epicStatus)
		if result != test.expected {
			t.Errorf("Expected FromEpicStatus(%s) to return %s, got %s",
				test.epicStatus, test.expected, result)
		}
	}
}

func TestEpicLifecycleStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		from     EpicLifecycleStatus
		to       EpicLifecycleStatus
		expected bool
	}{
		// Valid transitions
		{LifecycleStatusPending, LifecycleStatusWIP, true},
		{LifecycleStatusWIP, LifecycleStatusDone, true},

		// Invalid transitions
		{LifecycleStatusPending, LifecycleStatusDone, false},
		{LifecycleStatusWIP, LifecycleStatusPending, false},
		{LifecycleStatusDone, LifecycleStatusWIP, false},
		{LifecycleStatusDone, LifecycleStatusPending, false},

		// Self transitions (not allowed)
		{LifecycleStatusPending, LifecycleStatusPending, false},
		{LifecycleStatusWIP, LifecycleStatusWIP, false},
		{LifecycleStatusDone, LifecycleStatusDone, false},
	}

	for _, test := range tests {
		result := test.from.CanTransitionTo(test.to)
		if result != test.expected {
			t.Errorf("Expected %s.CanTransitionTo(%s) to return %t, got %t",
				test.from, test.to, test.expected, result)
		}
	}
}

func TestLifecycleService_StartEpic_Success(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	ls := NewLifecycleService(storage, queryService)

	// Set deterministic time
	testTime := time.Date(2025, 8, 16, 15, 30, 0, 0, time.UTC)
	ls.WithTimeSource(func() time.Time { return testTime })

	// Create and store a pending epic
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusPlanning, // pending
	}
	storage.StoreEpic("test-epic.xml", testEpic)

	// Start the epic
	request := StartEpicRequest{
		EpicFile: "test-epic.xml",
	}

	result, err := ls.StartEpic(request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify result
	if result.EpicID != "epic-1" {
		t.Errorf("Expected epic ID 'epic-1', got '%s'", result.EpicID)
	}

	if result.PreviousStatus != LifecycleStatusPending {
		t.Errorf("Expected previous status %s, got %s", LifecycleStatusPending, result.PreviousStatus)
	}

	if result.NewStatus != LifecycleStatusWIP {
		t.Errorf("Expected new status %s, got %s", LifecycleStatusWIP, result.NewStatus)
	}

	if !result.StartedAt.Equal(testTime) {
		t.Errorf("Expected started at %v, got %v", testTime, result.StartedAt)
	}

	if !result.EventCreated {
		t.Error("Expected event to be created")
	}

	// Verify epic was updated in storage
	updatedEpic, err := storage.LoadEpic("test-epic.xml")
	if err != nil {
		t.Fatalf("Failed to load updated epic: %v", err)
	}

	if updatedEpic.Status != epic.StatusActive {
		t.Errorf("Expected epic status to be %s, got %s", epic.StatusActive, updatedEpic.Status)
	}
}

func TestLifecycleService_StartEpic_WithTimestamp(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	ls := NewLifecycleService(storage, queryService)

	// Create and store a pending epic
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusPlanning,
	}
	storage.StoreEpic("test-epic.xml", testEpic)

	// Start the epic with specific timestamp
	specificTime := time.Date(2025, 8, 16, 15, 30, 0, 0, time.UTC)
	request := StartEpicRequest{
		EpicFile:  "test-epic.xml",
		Timestamp: &specificTime,
	}

	result, err := ls.StartEpic(request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.StartedAt.Equal(specificTime) {
		t.Errorf("Expected started at %v, got %v", specificTime, result.StartedAt)
	}
}

func TestLifecycleService_StartEpic_AlreadyStarted(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	ls := NewLifecycleService(storage, queryService)

	// Create and store an already active epic
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusActive, // already WIP
	}
	storage.StoreEpic("test-epic.xml", testEpic)

	// Try to start the epic
	request := StartEpicRequest{
		EpicFile: "test-epic.xml",
	}

	result, err := ls.StartEpic(request)
	if err == nil {
		t.Error("Expected error when starting already started epic")
	}

	if result != nil {
		t.Error("Expected nil result when error occurs")
	}

	// Check that it's a TransitionError
	var transitionErr *TransitionError
	if !isTransitionError(err, &transitionErr) {
		t.Errorf("Expected TransitionError, got %T: %v", err, err)
	} else {
		if transitionErr.CurrentStatus != LifecycleStatusWIP {
			t.Errorf("Expected current status %s, got %s", LifecycleStatusWIP, transitionErr.CurrentStatus)
		}
		if transitionErr.TargetStatus != LifecycleStatusWIP {
			t.Errorf("Expected target status %s, got %s", LifecycleStatusWIP, transitionErr.TargetStatus)
		}
	}
}

func TestLifecycleService_StartEpic_InvalidStatus(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	ls := NewLifecycleService(storage, queryService)

	// Create and store a completed epic
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusCompleted, // done
	}
	storage.StoreEpic("test-epic.xml", testEpic)

	// Try to start the epic
	request := StartEpicRequest{
		EpicFile: "test-epic.xml",
	}

	result, err := ls.StartEpic(request)
	if err == nil {
		t.Error("Expected error when starting completed epic")
	}

	if result != nil {
		t.Error("Expected nil result when error occurs")
	}

	// Check that it's a TransitionError
	var transitionErr *TransitionError
	if !isTransitionError(err, &transitionErr) {
		t.Errorf("Expected TransitionError, got %T: %v", err, err)
	}
}

func TestLifecycleService_StartEpic_NonExistentFile(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	ls := NewLifecycleService(storage, queryService)

	// Try to start non-existent epic
	request := StartEpicRequest{
		EpicFile: "non-existent.xml",
	}

	result, err := ls.StartEpic(request)
	if err == nil {
		t.Error("Expected error when starting non-existent epic")
	}

	if result != nil {
		t.Error("Expected nil result when error occurs")
	}
}

func TestTransitionError_Error(t *testing.T) {
	err := &TransitionError{
		EpicID:        "epic-1",
		CurrentStatus: LifecycleStatusWIP,
		TargetStatus:  LifecycleStatusWIP,
		Message:       "Test error message",
		Suggestion:    "Test suggestion",
	}

	if err.Error() != "Test error message" {
		t.Errorf("Expected error message 'Test error message', got '%s'", err.Error())
	}
}

func TestCompletionValidationError_Error(t *testing.T) {
	err := &CompletionValidationError{
		EpicID:  "epic-1",
		Message: "Test validation error",
	}

	if err.Error() != "Test validation error" {
		t.Errorf("Expected error message 'Test validation error', got '%s'", err.Error())
	}
}

// Helper function to check if error is TransitionError
func isTransitionError(err error, target **TransitionError) bool {
	if te, ok := err.(*TransitionError); ok {
		*target = te
		return true
	}
	return false
}

// Helper function to check if error is CompletionValidationError
func isCompletionValidationError(err error, target **CompletionValidationError) bool {
	if cve, ok := err.(*CompletionValidationError); ok {
		*target = cve
		return true
	}
	return false
}

func TestLifecycleService_CompleteEpic_Success(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	ls := NewLifecycleService(storage, queryService)

	// Set deterministic time
	testTime := time.Date(2025, 8, 20, 16, 45, 0, 0, time.UTC)
	ls.WithTimeSource(func() time.Time { return testTime })

	// Create and store a completed epic (all phases and tests done)
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusActive, // WIP
		Phases: []epic.Phase{
			{ID: "phase-1", Name: "Phase 1", Status: epic.StatusCompleted},
			{ID: "phase-2", Name: "Phase 2", Status: epic.StatusCompleted},
		},
		Tasks: []epic.Task{
			{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
			{ID: "task-2", PhaseID: "phase-2", Name: "Task 2", Status: epic.StatusCompleted},
		},
		Tests: []epic.Test{
			{ID: "test-1", TaskID: "task-1", Name: "Test 1", Status: epic.StatusCompleted},
			{ID: "test-2", TaskID: "task-2", Name: "Test 2", Status: epic.StatusCompleted},
		},
	}
	storage.StoreEpic("test-epic.xml", testEpic)

	// Complete the epic
	request := CompleteEpicRequest{
		EpicFile: "test-epic.xml",
	}

	result, err := ls.CompleteEpic(request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify result
	if result.EpicID != "epic-1" {
		t.Errorf("Expected epic ID 'epic-1', got '%s'", result.EpicID)
	}

	if result.PreviousStatus != LifecycleStatusWIP {
		t.Errorf("Expected previous status %s, got %s", LifecycleStatusWIP, result.PreviousStatus)
	}

	if result.NewStatus != LifecycleStatusDone {
		t.Errorf("Expected new status %s, got %s", LifecycleStatusDone, result.NewStatus)
	}

	if !result.CompletedAt.Equal(testTime) {
		t.Errorf("Expected completed at %v, got %v", testTime, result.CompletedAt)
	}

	if !result.EventCreated {
		t.Error("Expected event to be created")
	}

	// Verify summary
	if result.Summary.TotalPhases != 2 {
		t.Errorf("Expected 2 total phases, got %d", result.Summary.TotalPhases)
	}
	if result.Summary.TotalTasks != 2 {
		t.Errorf("Expected 2 total tasks, got %d", result.Summary.TotalTasks)
	}
	if result.Summary.TotalTests != 2 {
		t.Errorf("Expected 2 total tests, got %d", result.Summary.TotalTests)
	}

	// Verify epic was updated in storage
	updatedEpic, err := storage.LoadEpic("test-epic.xml")
	if err != nil {
		t.Fatalf("Failed to load updated epic: %v", err)
	}

	if updatedEpic.Status != epic.StatusCompleted {
		t.Errorf("Expected epic status to be %s, got %s", epic.StatusCompleted, updatedEpic.Status)
	}

	// Event logging will be implemented in a later epic
}

func TestLifecycleService_CompleteEpic_WithPendingWork(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	ls := NewLifecycleService(storage, queryService)

	// Create and store an epic with pending work
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusActive, // WIP
		Phases: []epic.Phase{
			{ID: "phase-1", Name: "Phase 1", Status: epic.StatusCompleted},
			{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPlanning}, // pending
		},
		Tasks: []epic.Task{
			{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
			{ID: "task-2", PhaseID: "phase-2", Name: "Task 2", Status: epic.StatusPlanning}, // pending
		},
		Tests: []epic.Test{
			{ID: "test-1", TaskID: "task-1", Name: "Test 1", Status: epic.StatusCompleted},
			{ID: "test-2", TaskID: "task-2", Name: "Test 2", Status: epic.StatusPlanning}, // failing
		},
	}
	storage.StoreEpic("test-epic.xml", testEpic)

	// Try to complete the epic
	request := CompleteEpicRequest{
		EpicFile: "test-epic.xml",
	}

	result, err := ls.CompleteEpic(request)
	if err == nil {
		t.Error("Expected error when completing epic with pending work")
	}

	if result != nil {
		t.Error("Expected nil result when error occurs")
	}

	// Check that it's a CompletionValidationError
	var validationErr *CompletionValidationError
	if !isCompletionValidationError(err, &validationErr) {
		t.Errorf("Expected CompletionValidationError, got %T: %v", err, err)
	} else {
		if len(validationErr.PendingPhases) != 1 {
			t.Errorf("Expected 1 pending phase, got %d", len(validationErr.PendingPhases))
		}
		if len(validationErr.FailingTests) != 1 {
			t.Errorf("Expected 1 failing test, got %d", len(validationErr.FailingTests))
		}
	}
}

func TestLifecycleService_CompleteEpic_WithFailingTests(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	ls := NewLifecycleService(storage, queryService)

	// Create and store an epic with failing tests only
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusActive, // WIP
		Phases: []epic.Phase{
			{ID: "phase-1", Name: "Phase 1", Status: epic.StatusCompleted},
			{ID: "phase-2", Name: "Phase 2", Status: epic.StatusCompleted},
		},
		Tests: []epic.Test{
			{ID: "test-1", TaskID: "task-1", Name: "Test 1", Status: epic.StatusCompleted},
			{ID: "test-2", TaskID: "task-2", Name: "Test 2", Status: epic.StatusActive},   // failing
			{ID: "test-3", TaskID: "task-2", Name: "Test 3", Status: epic.StatusPlanning}, // failing
		},
	}
	storage.StoreEpic("test-epic.xml", testEpic)

	// Try to complete the epic
	request := CompleteEpicRequest{
		EpicFile: "test-epic.xml",
	}

	result, err := ls.CompleteEpic(request)
	if err == nil {
		t.Error("Expected error when completing epic with failing tests")
	}

	if result != nil {
		t.Error("Expected nil result when error occurs")
	}

	// Check that it's a CompletionValidationError with failing tests
	var validationErr *CompletionValidationError
	if !isCompletionValidationError(err, &validationErr) {
		t.Errorf("Expected CompletionValidationError, got %T: %v", err, err)
	} else {
		if len(validationErr.PendingPhases) != 0 {
			t.Errorf("Expected 0 pending phases, got %d", len(validationErr.PendingPhases))
		}
		if len(validationErr.FailingTests) != 2 {
			t.Errorf("Expected 2 failing tests, got %d", len(validationErr.FailingTests))
		}
	}
}

func TestLifecycleService_CompleteEpic_NotStarted(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	ls := NewLifecycleService(storage, queryService)

	// Create and store a pending epic (not started)
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusPlanning, // pending, not started
	}
	storage.StoreEpic("test-epic.xml", testEpic)

	// Try to complete the epic
	request := CompleteEpicRequest{
		EpicFile: "test-epic.xml",
	}

	result, err := ls.CompleteEpic(request)
	if err == nil {
		t.Error("Expected error when completing epic that wasn't started")
	}

	if result != nil {
		t.Error("Expected nil result when error occurs")
	}

	// Check that it's a TransitionError
	var transitionErr *TransitionError
	if !isTransitionError(err, &transitionErr) {
		t.Errorf("Expected TransitionError, got %T: %v", err, err)
	} else {
		if transitionErr.CurrentStatus != LifecycleStatusPending {
			t.Errorf("Expected current status %s, got %s", LifecycleStatusPending, transitionErr.CurrentStatus)
		}
		if transitionErr.TargetStatus != LifecycleStatusDone {
			t.Errorf("Expected target status %s, got %s", LifecycleStatusDone, transitionErr.TargetStatus)
		}
	}
}

func TestLifecycleService_ValidateEpicCompletion(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	ls := NewLifecycleService(storage, queryService)

	tests := []struct {
		name            string
		epic            *epic.Epic
		expectedValid   bool
		expectedPending int
		expectedFailing int
	}{
		{
			name: "valid epic ready for completion",
			epic: &epic.Epic{
				ID:     "epic-1",
				Status: epic.StatusActive,
				Phases: []epic.Phase{
					{ID: "phase-1", Status: epic.StatusCompleted},
				},
				Tasks: []epic.Task{
					{ID: "task-1", Status: epic.StatusCompleted},
				},
				Tests: []epic.Test{
					{ID: "test-1", Status: epic.StatusCompleted},
				},
			},
			expectedValid:   true,
			expectedPending: 0,
			expectedFailing: 0,
		},
		{
			name: "epic with pending phases",
			epic: &epic.Epic{
				ID:     "epic-1",
				Status: epic.StatusActive,
				Phases: []epic.Phase{
					{ID: "phase-1", Status: epic.StatusCompleted},
					{ID: "phase-2", Status: epic.StatusPlanning},
				},
				Tests: []epic.Test{
					{ID: "test-1", Status: epic.StatusCompleted},
				},
			},
			expectedValid:   false,
			expectedPending: 1,
			expectedFailing: 0,
		},
		{
			name: "epic with failing tests",
			epic: &epic.Epic{
				ID:     "epic-1",
				Status: epic.StatusActive,
				Phases: []epic.Phase{
					{ID: "phase-1", Status: epic.StatusCompleted},
				},
				Tests: []epic.Test{
					{ID: "test-1", Status: epic.StatusCompleted},
					{ID: "test-2", Status: epic.StatusActive},
				},
			},
			expectedValid:   false,
			expectedPending: 0,
			expectedFailing: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Store the test epic
			storage.StoreEpic("test-epic.xml", test.epic)

			// Validate epic completion
			result, err := ls.ValidateEpicCompletion("test-epic.xml")
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.IsValid != test.expectedValid {
				t.Errorf("Expected IsValid %t, got %t", test.expectedValid, result.IsValid)
			}

			if len(result.PendingPhases) != test.expectedPending {
				t.Errorf("Expected %d pending phases, got %d", test.expectedPending, len(result.PendingPhases))
			}

			if len(result.FailingTests) != test.expectedFailing {
				t.Errorf("Expected %d failing tests, got %d", test.expectedFailing, len(result.FailingTests))
			}

			// Verify suggestions are generated
			if len(result.Suggestions) == 0 {
				t.Error("Expected suggestions to be generated")
			}
		})
	}
}

func TestLifecycleService_EpicEventCreation(t *testing.T) {
	factory := storage.NewFactory(true)
	storageImpl := factory.CreateStorage()
	queryService := query.NewQueryService(storageImpl)
	lifecycleService := NewLifecycleService(storageImpl, queryService)

	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	lifecycleService = lifecycleService.WithTimeSource(func() time.Time { return testTime })

	// Create a test epic
	testEpic := &epic.Epic{
		ID:     "test-epic",
		Name:   "Test Epic for Events",
		Status: epic.StatusPlanning,
		Phases: []epic.Phase{
			{ID: "phase1", Name: "Phase 1", Status: epic.StatusCompleted},
		},
		Tasks: []epic.Task{
			{ID: "task1", PhaseID: "phase1", Name: "Task 1", Status: epic.StatusCompleted},
		},
		Tests: []epic.Test{
			{ID: "test1", PhaseID: "phase1", TaskID: "task1", Name: "Test 1", Status: epic.StatusCompleted, TestStatus: epic.TestStatusPassed},
		},
	}

	epicFile := "test-epic.xml"
	err := storageImpl.SaveEpic(testEpic, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	t.Run("StartEpic creates epic_started event", func(t *testing.T) {
		request := StartEpicRequest{
			EpicFile:  epicFile,
			Timestamp: &testTime,
		}

		result, err := lifecycleService.StartEpic(request)
		if err != nil {
			t.Fatalf("StartEpic failed: %v", err)
		}

		if !result.EventCreated {
			t.Error("Expected event to be created for epic start")
		}

		// Load epic and verify event was created
		updatedEpic, err := storageImpl.LoadEpic(epicFile)
		if err != nil {
			t.Fatalf("Failed to load updated epic: %v", err)
		}

		if len(updatedEpic.Events) != 1 {
			t.Errorf("Expected 1 event, got %d", len(updatedEpic.Events))
		}

		event := updatedEpic.Events[0]
		if event.Type != "epic_started" {
			t.Errorf("Expected event type epic_started, got %s", event.Type)
		}

		expectedData := "Epic Test Epic for Events started"
		if event.Data != expectedData {
			t.Errorf("Expected event data '%s', got '%s'", expectedData, event.Data)
		}

		if !event.Timestamp.Equal(testTime) {
			t.Errorf("Expected event timestamp %v, got %v", testTime, event.Timestamp)
		}
	})

	t.Run("DoneEpic creates epic_completed event", func(t *testing.T) {
		// Prepare epic in WIP status and clear previous events
		wipEpic := &epic.Epic{
			ID:     "test-epic",
			Name:   "Test Epic for Events",
			Status: epic.StatusActive, // WIP state
			Phases: []epic.Phase{
				{ID: "phase1", Name: "Phase 1", Status: epic.StatusCompleted},
			},
			Tasks: []epic.Task{
				{ID: "task1", PhaseID: "phase1", Name: "Task 1", Status: epic.StatusCompleted},
			},
			Tests: []epic.Test{
				{ID: "test1", PhaseID: "phase1", TaskID: "task1", Name: "Test 1", Status: epic.StatusCompleted, TestStatus: epic.TestStatusPassed},
			},
			Events: []epic.Event{}, // Clear events for this test
		}

		err := storageImpl.SaveEpic(wipEpic, epicFile)
		if err != nil {
			t.Fatalf("Failed to save WIP epic: %v", err)
		}

		request := DoneEpicRequest{
			EpicFile:  epicFile,
			Timestamp: &testTime,
		}

		result, err := lifecycleService.DoneEpic(request)
		if err != nil {
			t.Fatalf("DoneEpic failed: %v", err)
		}

		if !result.EventCreated {
			t.Error("Expected event to be created for epic completion")
		}

		// Load epic and verify event was created
		updatedEpic, err := storageImpl.LoadEpic(epicFile)
		if err != nil {
			t.Fatalf("Failed to load updated epic: %v", err)
		}

		if len(updatedEpic.Events) != 1 {
			t.Errorf("Expected 1 event, got %d", len(updatedEpic.Events))
		}

		event := updatedEpic.Events[0]
		if event.Type != "epic_completed" {
			t.Errorf("Expected event type epic_completed, got %s", event.Type)
		}

		expectedData := "Epic Test Epic for Events completed"
		if event.Data != expectedData {
			t.Errorf("Expected event data '%s', got '%s'", expectedData, event.Data)
		}

		if !event.Timestamp.Equal(testTime) {
			t.Errorf("Expected event timestamp %v, got %v", testTime, event.Timestamp)
		}
	})

	t.Run("Epic events work with empty name", func(t *testing.T) {
		// Create epic without name
		noNameEpic := &epic.Epic{
			ID:     "no-name-epic",
			Name:   "",
			Status: epic.StatusPlanning,
			Phases: []epic.Phase{
				{ID: "phase1", Name: "Phase 1", Status: epic.StatusCompleted},
			},
			Tasks: []epic.Task{
				{ID: "task1", PhaseID: "phase1", Name: "Task 1", Status: epic.StatusCompleted},
			},
			Tests: []epic.Test{
				{ID: "test1", PhaseID: "phase1", TaskID: "task1", Name: "Test 1", Status: epic.StatusCompleted, TestStatus: epic.TestStatusPassed},
			},
		}

		noNameFile := "no-name-epic.xml"
		err := storageImpl.SaveEpic(noNameEpic, noNameFile)
		if err != nil {
			t.Fatalf("Failed to save no-name epic: %v", err)
		}

		request := StartEpicRequest{
			EpicFile:  noNameFile,
			Timestamp: &testTime,
		}

		_, err = lifecycleService.StartEpic(request)
		if err != nil {
			t.Fatalf("StartEpic failed: %v", err)
		}

		// Load epic and verify event uses ID when name is empty
		updatedEpic, err := storageImpl.LoadEpic(noNameFile)
		if err != nil {
			t.Fatalf("Failed to load updated epic: %v", err)
		}

		if len(updatedEpic.Events) != 1 {
			t.Errorf("Expected 1 event, got %d", len(updatedEpic.Events))
		}

		event := updatedEpic.Events[0]
		expectedData := "Epic no-name-epic started"
		if event.Data != expectedData {
			t.Errorf("Expected event data '%s', got '%s'", expectedData, event.Data)
		}
	})
}

func TestLifecycleService_CalculateValidationSummary(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	ls := NewLifecycleService(storage, queryService)

	testEpic := &epic.Epic{
		Phases: []epic.Phase{
			{ID: "phase-1", Status: epic.StatusCompleted},
			{ID: "phase-2", Status: epic.StatusPlanning},
			{ID: "phase-3", Status: epic.StatusActive},
		},
		Tasks: []epic.Task{
			{ID: "task-1", Status: epic.StatusCompleted},
			{ID: "task-2", Status: epic.StatusCompleted},
			{ID: "task-3", Status: epic.StatusPlanning},
		},
		Tests: []epic.Test{
			{ID: "test-1", Status: epic.StatusCompleted},
			{ID: "test-2", Status: epic.StatusActive},
		},
	}

	summary := ls.calculateValidationSummary(testEpic)

	// Check totals
	if summary.TotalPhases != 3 {
		t.Errorf("Expected 3 total phases, got %d", summary.TotalPhases)
	}
	if summary.TotalTasks != 3 {
		t.Errorf("Expected 3 total tasks, got %d", summary.TotalTasks)
	}
	if summary.TotalTests != 2 {
		t.Errorf("Expected 2 total tests, got %d", summary.TotalTests)
	}

	// Check completed counts
	if summary.CompletedPhases != 1 {
		t.Errorf("Expected 1 completed phase, got %d", summary.CompletedPhases)
	}
	if summary.CompletedTasks != 2 {
		t.Errorf("Expected 2 completed tasks, got %d", summary.CompletedTasks)
	}
	if summary.PassingTests != 1 {
		t.Errorf("Expected 1 passing test, got %d", summary.PassingTests)
	}

	// Check pending counts
	if summary.PendingPhases != 2 {
		t.Errorf("Expected 2 pending phases, got %d", summary.PendingPhases)
	}
	if summary.PendingTasks != 1 {
		t.Errorf("Expected 1 pending task, got %d", summary.PendingTasks)
	}
	if summary.FailingTests != 1 {
		t.Errorf("Expected 1 failing test, got %d", summary.FailingTests)
	}

	// Check completion percentage: (1+2+1)/(3+3+2) = 4/8 = 50%
	expectedPercent := 50
	if summary.CompletionPercent != expectedPercent {
		t.Errorf("Expected %d%% completion, got %d%%", expectedPercent, summary.CompletionPercent)
	}
}

func TestLifecycleService_FormatValidationError(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	ls := NewLifecycleService(storage, queryService)

	result := &ValidationResult{
		IsValid: false,
		PendingPhases: []PendingPhase{
			{ID: "phase-1", Name: "Phase 1"},
		},
		FailingTests: []FailingTest{
			{ID: "test-1", Name: "Test 1"},
		},
		Summary: ValidationSummary{
			TotalPhases:       2,
			CompletedPhases:   1,
			TotalTasks:        2,
			CompletedTasks:    2,
			TotalTests:        2,
			PassingTests:      1,
			CompletionPercent: 75,
		},
		Suggestions: []string{"Fix failing tests", "Complete pending phases"},
	}

	message := ls.FormatValidationError(result, "epic-1")

	// Check that the message contains key information
	if !strings.Contains(message, "Epic epic-1 cannot be completed") {
		t.Error("Expected message to contain epic ID and completion status")
	}
	if !strings.Contains(message, "75% complete") {
		t.Error("Expected message to contain completion percentage")
	}
	if !strings.Contains(message, "Phase 1 (phase-1)") {
		t.Error("Expected message to contain pending phase details")
	}
	if !strings.Contains(message, "Test 1 (test-1)") {
		t.Error("Expected message to contain failing test details")
	}
	if !strings.Contains(message, "Suggestions: Fix failing tests; Complete pending phases") {
		t.Error("Expected message to contain suggestions")
	}
}

func TestLifecycleService_ValidateEpicState(t *testing.T) {
	storage := storage.NewMemoryStorage()
	queryService := query.NewQueryService(storage)
	lifecycleService := NewLifecycleService(storage, queryService)

	t.Run("valid epic state", func(t *testing.T) {
		epic := &epic.Epic{
			ID:     "epic-1",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusActive},
			},
		}

		result := lifecycleService.ValidateEpicState(epic)

		assert.Equal(t, ValidationLevelValid, result.Level)
		assert.Empty(t, result.Issues)
		assert.Equal(t, "Epic state is valid - no issues found", result.Summary)
	})

	t.Run("multiple active phases error", func(t *testing.T) {
		epic := &epic.Epic{
			ID:     "epic-1",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusActive},
			},
		}

		result := lifecycleService.ValidateEpicState(epic)

		assert.Equal(t, ValidationLevelError, result.Level)
		assert.GreaterOrEqual(t, len(result.Issues), 1) // May have additional warnings

		// Find the multiple active phases issue
		var foundMultiplePhases bool
		for _, issue := range result.Issues {
			if issue.Type == "multiple_active_phases" {
				foundMultiplePhases = true
				assert.Contains(t, issue.Message, "Multiple active phases found")
				break
			}
		}
		assert.True(t, foundMultiplePhases, "Should find multiple active phases issue")
	})

	t.Run("multiple active tasks in phase error", func(t *testing.T) {
		epic := &epic.Epic{
			ID:     "epic-1",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusActive},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusActive},
			},
		}

		result := lifecycleService.ValidateEpicState(epic)

		assert.Equal(t, ValidationLevelError, result.Level)
		assert.Len(t, result.Issues, 1)
		assert.Equal(t, "multiple_active_tasks_in_phase", result.Issues[0].Type)
		assert.Contains(t, result.Issues[0].Message, "Multiple active tasks in phase")
	})

	t.Run("active task in inactive phase error", func(t *testing.T) {
		epic := &epic.Epic{
			ID:     "epic-1",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPlanning},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusActive},
			},
		}

		result := lifecycleService.ValidateEpicState(epic)

		assert.Equal(t, ValidationLevelError, result.Level)
		assert.Len(t, result.Issues, 1)
		assert.Equal(t, "active_task_in_inactive_phase", result.Issues[0].Type)
		assert.Contains(t, result.Issues[0].Message, "Active task task-1 is in inactive phase")
	})

	t.Run("phase ready for completion warning", func(t *testing.T) {
		epic := &epic.Epic{
			ID:     "epic-1",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
			},
		}

		result := lifecycleService.ValidateEpicState(epic)

		assert.Equal(t, ValidationLevelWarning, result.Level)
		assert.Len(t, result.Issues, 1)
		assert.Equal(t, "phase_ready_for_completion", result.Issues[0].Type)
		assert.Contains(t, result.Issues[0].Message, "should be marked as done")
		assert.Contains(t, result.Summary, "1 warning(s)")
	})

	t.Run("completed epic with pending work error", func(t *testing.T) {
		epic := &epic.Epic{
			ID:     "epic-1",
			Status: epic.StatusCompleted,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive}, // Still active
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusActive},
			},
		}

		result := lifecycleService.ValidateEpicState(epic)

		assert.Equal(t, ValidationLevelError, result.Level)
		assert.GreaterOrEqual(t, len(result.Issues), 1) // At least one error

		// Find the completed epic issue
		var foundCompletedEpicIssue bool
		for _, issue := range result.Issues {
			if issue.Type == "completed_epic_with_pending_work" {
				foundCompletedEpicIssue = true
				assert.Contains(t, issue.Message, "Epic is marked as completed but has pending")
				break
			}
		}
		assert.True(t, foundCompletedEpicIssue, "Should find completed epic with pending work issue")
	})
}
