package tests

import (
	"fmt"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/service"
	"github.com/mindreframer/agentpm/internal/storage"
)

// TestService provides test management operations with dependency injection
type TestService struct {
	storage    storage.Storage
	timeSource func() time.Time
}

// ServiceConfig holds configuration for creating test services
type ServiceConfig struct {
	UseMemory  bool
	TimeSource func() time.Time
}

// NewTestService creates a new test service with the given configuration
func NewTestService(cfg ServiceConfig) *TestService {
	factory := storage.NewFactory(cfg.UseMemory)

	timeSource := cfg.TimeSource
	if timeSource == nil {
		timeSource = time.Now
	}

	return &TestService{
		storage:    factory.CreateStorage(),
		timeSource: timeSource,
	}
}

// StartTest transitions a test from pending to wip status
func (s *TestService) StartTest(epicFile, testID string, timestamp *time.Time) (*TestOperation, error) {
	e, err := s.loadAndValidateEpic(epicFile)
	if err != nil {
		return nil, err
	}

	test, err := s.findTest(e, testID)
	if err != nil {
		return nil, err
	}

	// Get current test status (prefer TestStatus, fallback to Status conversion)
	currentTestStatus := s.getTestStatus(test)

	// Check if test is already in WIP status - this is not an error, just return a result indicating it's already started
	if currentTestStatus == epic.TestStatusWIP {
		return &TestOperation{
			TestID:    testID,
			Operation: "start",
			Status:    "already_started",
			Timestamp: s.timeSource(),
		}, nil
	}

	// Validate state transition
	if !currentTestStatus.CanTransitionTo(epic.TestStatusWIP) {
		return nil, &TestError{
			Type:    ErrorTypeInvalidTransition,
			TestID:  testID,
			Current: string(currentTestStatus),
			Target:  string(epic.TestStatusWIP),
			Message: fmt.Sprintf("Cannot start test %s: test is not in pending status", testID),
		}
	}

	// Validate prerequisite: associated task/phase must be active or completed
	if err := s.validateTestPrerequisites(e, test); err != nil {
		return nil, err
	}

	// Update test status and timestamp
	s.setTestStatus(test, epic.TestStatusWIP)
	if timestamp == nil {
		now := s.timeSource()
		timestamp = &now
	}
	test.StartedAt = timestamp

	// Create event for test start
	service.CreateEvent(e, service.EventTestStarted, test.PhaseID, test.TaskID, testID, "", *timestamp)

	// Save epic
	if err := s.storage.SaveEpic(e, epicFile); err != nil {
		return nil, &TestError{
			Type:    ErrorTypeIO,
			TestID:  testID,
			Message: fmt.Sprintf("Failed to save epic: %v", err),
			Cause:   err,
		}
	}

	return &TestOperation{
		TestID:    testID,
		Operation: "started",
		Status:    string(epic.TestStatusWIP),
		Timestamp: *timestamp,
	}, nil
}

// PassTest transitions a test from wip to passed status
func (s *TestService) PassTest(epicFile, testID string, timestamp *time.Time) (*TestOperation, error) {
	e, err := s.loadAndValidateEpic(epicFile)
	if err != nil {
		return nil, err
	}

	test, err := s.findTest(e, testID)
	if err != nil {
		return nil, err
	}

	// Get current test status
	currentTestStatus := s.getTestStatus(test)

	// Validate state transition
	if !currentTestStatus.CanTransitionTo(epic.TestStatusDone) {
		return nil, &TestError{
			Type:    ErrorTypeInvalidTransition,
			TestID:  testID,
			Current: string(currentTestStatus),
			Target:  string(epic.TestStatusDone),
			Message: fmt.Sprintf("Cannot pass test %s: test is not currently in progress", testID),
		}
	}

	// Update test status and result
	s.setTestStatus(test, epic.TestStatusDone)
	test.TestResult = epic.TestResultPassing
	if timestamp == nil {
		now := s.timeSource()
		timestamp = &now
	}
	test.PassedAt = timestamp
	// Clear any previous failure note
	test.FailureNote = ""

	// Create event for test pass
	service.CreateEvent(e, service.EventTestPassed, test.PhaseID, test.TaskID, testID, "", *timestamp)

	// Save epic
	if err := s.storage.SaveEpic(e, epicFile); err != nil {
		return nil, &TestError{
			Type:    ErrorTypeIO,
			TestID:  testID,
			Message: fmt.Sprintf("Failed to save epic: %v", err),
			Cause:   err,
		}
	}

	return &TestOperation{
		TestID:    testID,
		Operation: "passed",
		Status:    string(epic.TestStatusDone),
		Timestamp: *timestamp,
	}, nil
}

// FailTest transitions a test from wip to failed status with failure details
func (s *TestService) FailTest(epicFile, testID, failureReason string, timestamp *time.Time) (*TestOperation, error) {
	e, err := s.loadAndValidateEpic(epicFile)
	if err != nil {
		return nil, err
	}

	test, err := s.findTest(e, testID)
	if err != nil {
		return nil, err
	}

	// Get current test status
	currentTestStatus := s.getTestStatus(test)

	// Epic 13: Test must be in WIP status to be failed, or Done status (which can transition to WIP)
	if currentTestStatus != epic.TestStatusWIP && currentTestStatus != epic.TestStatusDone {
		return nil, &TestError{
			Type:    ErrorTypeInvalidTransition,
			TestID:  testID,
			Current: string(currentTestStatus),
			Target:  string(epic.TestStatusWIP),
			Message: fmt.Sprintf("Cannot fail test %s: test must be in progress (wip) or done to be failed", testID),
		}
	}

	// Update test status and result
	s.setTestStatus(test, epic.TestStatusWIP)
	test.TestResult = epic.TestResultFailing
	if timestamp == nil {
		now := s.timeSource()
		timestamp = &now
	}
	test.FailedAt = timestamp
	test.FailureNote = failureReason

	// Create event for test failure
	service.CreateEvent(e, service.EventTestFailed, test.PhaseID, test.TaskID, testID, failureReason, *timestamp)

	// Save epic
	if err := s.storage.SaveEpic(e, epicFile); err != nil {
		return nil, &TestError{
			Type:    ErrorTypeIO,
			TestID:  testID,
			Message: fmt.Sprintf("Failed to save epic: %v", err),
			Cause:   err,
		}
	}

	return &TestOperation{
		TestID:        testID,
		Operation:     "failed",
		Status:        string(epic.TestStatusWIP),
		Timestamp:     *timestamp,
		FailureReason: failureReason,
	}, nil
}

// CancelTest transitions a test from wip to cancelled status with cancellation reason
func (s *TestService) CancelTest(epicFile, testID, cancellationReason string, timestamp *time.Time) (*TestOperation, error) {
	e, err := s.loadAndValidateEpic(epicFile)
	if err != nil {
		return nil, err
	}

	test, err := s.findTest(e, testID)
	if err != nil {
		return nil, err
	}

	// Get current test status
	currentTestStatus := s.getTestStatus(test)

	// Validate state transition
	if !currentTestStatus.CanTransitionTo(epic.TestStatusCancelled) {
		return nil, &TestError{
			Type:    ErrorTypeInvalidTransition,
			TestID:  testID,
			Current: string(currentTestStatus),
			Target:  string(epic.TestStatusCancelled),
			Message: fmt.Sprintf("Cannot cancel test %s: test is not currently in progress", testID),
		}
	}

	// Update test status, timestamp, and cancellation details
	s.setTestStatus(test, epic.TestStatusCancelled)
	if timestamp == nil {
		now := s.timeSource()
		timestamp = &now
	}
	test.CancelledAt = timestamp
	test.CancellationReason = cancellationReason

	// Create event for test cancellation
	service.CreateEvent(e, service.EventTestCancelled, test.PhaseID, test.TaskID, testID, cancellationReason, *timestamp)

	// Save epic
	if err := s.storage.SaveEpic(e, epicFile); err != nil {
		return nil, &TestError{
			Type:    ErrorTypeIO,
			TestID:  testID,
			Message: fmt.Sprintf("Failed to save epic: %v", err),
			Cause:   err,
		}
	}

	return &TestOperation{
		TestID:             testID,
		Operation:          "cancelled",
		Status:             string(epic.TestStatusCancelled),
		Timestamp:          *timestamp,
		CancellationReason: cancellationReason,
	}, nil
}

// Helper methods

func (s *TestService) loadAndValidateEpic(epicFile string) (*epic.Epic, error) {
	if !s.storage.EpicExists(epicFile) {
		return nil, &TestError{
			Type:    ErrorTypeNotFound,
			Message: fmt.Sprintf("Epic file not found: %s", epicFile),
		}
	}

	e, err := s.storage.LoadEpic(epicFile)
	if err != nil {
		return nil, &TestError{
			Type:    ErrorTypeIO,
			Message: fmt.Sprintf("Failed to load epic: %v", err),
			Cause:   err,
		}
	}

	return e, nil
}

func (s *TestService) findTest(e *epic.Epic, testID string) (*epic.Test, error) {
	for i := range e.Tests {
		if e.Tests[i].ID == testID {
			return &e.Tests[i], nil
		}
	}

	return nil, &TestError{
		Type:    ErrorTypeNotFound,
		TestID:  testID,
		Message: fmt.Sprintf("Test %s not found in epic", testID),
	}
}

// getTestStatus returns the current test status, preferring TestStatus over Status conversion
func (s *TestService) getTestStatus(test *epic.Test) epic.TestStatus {
	if test.TestStatus != "" && test.TestStatus.IsValid() {
		return test.TestStatus
	}

	// Convert from legacy Status field
	switch test.Status {
	case epic.StatusPlanning:
		return epic.TestStatusPending
	case epic.StatusActive:
		return epic.TestStatusWIP
	case epic.StatusCompleted:
		return epic.TestStatusDone
	case epic.StatusCancelled:
		return epic.TestStatusCancelled
	default:
		return epic.TestStatusPending
	}
}

// setTestStatus sets both TestStatus and Status fields for compatibility
func (s *TestService) setTestStatus(test *epic.Test, status epic.TestStatus) {
	test.TestStatus = status

	// Also set the legacy Status field for compatibility
	switch status {
	case epic.TestStatusPending:
		test.Status = epic.StatusPlanning
	case epic.TestStatusWIP:
		test.Status = epic.StatusActive
	case epic.TestStatusDone:
		test.Status = epic.StatusCompleted
	case epic.TestStatusCancelled:
		test.Status = epic.StatusCancelled
	}
}

func (s *TestService) validateTestPrerequisites(e *epic.Epic, test *epic.Test) error {
	// Check if associated task is active or completed
	if test.TaskID != "" {
		for _, task := range e.Tasks {
			if task.ID == test.TaskID {
				if task.Status != epic.StatusActive && task.Status != epic.StatusCompleted {
					return &TestError{
						Type:    ErrorTypeValidation,
						TestID:  test.ID,
						Message: fmt.Sprintf("Cannot start test %s: associated task %s is not active or completed (status: %s)", test.ID, test.TaskID, task.Status),
					}
				}
				break // Found the task, continue to check phase
			}
		}
	}

	// Check if associated phase is active or completed
	if test.PhaseID != "" {
		for _, phase := range e.Phases {
			if phase.ID == test.PhaseID {
				if phase.Status != epic.StatusActive && phase.Status != epic.StatusCompleted {
					return &TestError{
						Type:    ErrorTypeValidation,
						TestID:  test.ID,
						Message: fmt.Sprintf("Cannot start test %s: associated phase %s is not active or completed (status: %s)", test.ID, test.PhaseID, phase.Status),
					}
				}
				break // Found the phase
			}
		}
	}

	return nil
}

// Result types for test operations
type TestOperation struct {
	TestID             string    `json:"test_id"`
	Operation          string    `json:"operation"`
	Status             string    `json:"status"`
	Timestamp          time.Time `json:"timestamp"`
	FailureReason      string    `json:"failure_reason,omitempty"`
	CancellationReason string    `json:"cancellation_reason,omitempty"`
}

// Error types and handling
type ErrorType string

const (
	ErrorTypeNotFound          ErrorType = "not_found"
	ErrorTypeValidation        ErrorType = "validation"
	ErrorTypeIO                ErrorType = "io"
	ErrorTypeInvalidTransition ErrorType = "invalid_transition"
)

type TestError struct {
	Type    ErrorType `json:"type"`
	TestID  string    `json:"test_id,omitempty"`
	Current string    `json:"current_status,omitempty"`
	Target  string    `json:"target_status,omitempty"`
	Message string    `json:"message"`
	Cause   error     `json:"-"`
}

func (e *TestError) Error() string {
	return e.Message
}

func (e *TestError) Unwrap() error {
	return e.Cause
}

// Helper functions for error type checking
func IsNotFound(err error) bool {
	if te, ok := err.(*TestError); ok {
		return te.Type == ErrorTypeNotFound
	}
	return false
}

func IsValidation(err error) bool {
	if te, ok := err.(*TestError); ok {
		return te.Type == ErrorTypeValidation
	}
	return false
}

func IsInvalidTransition(err error) bool {
	if te, ok := err.(*TestError); ok {
		return te.Type == ErrorTypeInvalidTransition
	}
	return false
}
