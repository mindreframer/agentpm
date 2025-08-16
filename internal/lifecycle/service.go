package lifecycle

import (
	"fmt"
	"time"

	"github.com/memomoo/agentpm/internal/epic"
	"github.com/memomoo/agentpm/internal/query"
	"github.com/memomoo/agentpm/internal/storage"
)

// LifecycleService manages epic lifecycle transitions and validation
type LifecycleService struct {
	storage      storage.Storage
	queryService *query.QueryService
	timeSource   func() time.Time
}

// NewLifecycleService creates a new LifecycleService with dependency injection
func NewLifecycleService(storage storage.Storage, queryService *query.QueryService) *LifecycleService {
	return &LifecycleService{
		storage:      storage,
		queryService: queryService,
		timeSource:   time.Now,
	}
}

// WithTimeSource allows injection of custom time source for deterministic testing
func (ls *LifecycleService) WithTimeSource(timeSource func() time.Time) *LifecycleService {
	ls.timeSource = timeSource
	return ls
}

// EpicLifecycleStatus represents the lifecycle status of an epic
type EpicLifecycleStatus string

const (
	LifecycleStatusPending EpicLifecycleStatus = "pending"
	LifecycleStatusWIP     EpicLifecycleStatus = "wip"
	LifecycleStatusDone    EpicLifecycleStatus = "done"
)

// String implements the Stringer interface
func (s EpicLifecycleStatus) String() string {
	return string(s)
}

// IsValid checks if the status is a valid epic lifecycle status
func (s EpicLifecycleStatus) IsValid() bool {
	switch s {
	case LifecycleStatusPending, LifecycleStatusWIP, LifecycleStatusDone:
		return true
	default:
		return false
	}
}

// ToEpicStatus converts lifecycle status to epic.Status for storage
func (s EpicLifecycleStatus) ToEpicStatus() epic.Status {
	switch s {
	case LifecycleStatusPending:
		return epic.StatusPlanning
	case LifecycleStatusWIP:
		return epic.StatusActive
	case LifecycleStatusDone:
		return epic.StatusCompleted
	default:
		return epic.StatusPlanning
	}
}

// FromEpicStatus converts epic.Status to lifecycle status
func FromEpicStatus(status epic.Status) EpicLifecycleStatus {
	switch status {
	case epic.StatusPlanning:
		return LifecycleStatusPending
	case epic.StatusActive:
		return LifecycleStatusWIP
	case epic.StatusCompleted:
		return LifecycleStatusDone
	default:
		return LifecycleStatusPending
	}
}

// CanTransitionTo checks if the current status can transition to the target status
func (s EpicLifecycleStatus) CanTransitionTo(target EpicLifecycleStatus) bool {
	transitions := map[EpicLifecycleStatus][]EpicLifecycleStatus{
		LifecycleStatusPending: {LifecycleStatusWIP},
		LifecycleStatusWIP:     {LifecycleStatusDone},
		LifecycleStatusDone:    {}, // No transitions from done
	}

	for _, allowed := range transitions[s] {
		if allowed == target {
			return true
		}
	}
	return false
}

// StartEpicRequest represents a request to start an epic
type StartEpicRequest struct {
	EpicFile  string
	Timestamp *time.Time // optional, for deterministic testing
}

// StartEpicResult represents the result of starting an epic
type StartEpicResult struct {
	EpicID         string
	PreviousStatus EpicLifecycleStatus
	NewStatus      EpicLifecycleStatus
	StartedAt      time.Time
	Message        string
	EventCreated   bool
}

// StartEpic transitions an epic from pending to wip status
func (ls *LifecycleService) StartEpic(request StartEpicRequest) (*StartEpicResult, error) {
	// Load the epic
	loadedEpic, err := ls.storage.LoadEpic(request.EpicFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load epic: %w", err)
	}

	// Validate current status
	currentStatus := FromEpicStatus(loadedEpic.Status)
	if !currentStatus.CanTransitionTo(LifecycleStatusWIP) {
		return nil, &TransitionError{
			EpicID:        loadedEpic.ID,
			CurrentStatus: currentStatus,
			TargetStatus:  LifecycleStatusWIP,
			Message:       fmt.Sprintf("Epic is already started (current status: %s)", currentStatus),
			Suggestion:    "Use 'agentpm current' to see active work",
		}
	}

	// Determine timestamp
	startTime := ls.timeSource()
	if request.Timestamp != nil {
		startTime = *request.Timestamp
	}

	// Update epic status
	loadedEpic.Status = LifecycleStatusWIP.ToEpicStatus()

	// Event logging will be implemented in a later epic

	// Save the updated epic
	if err := ls.storage.SaveEpic(loadedEpic, request.EpicFile); err != nil {
		return nil, fmt.Errorf("failed to save epic: %w", err)
	}

	return &StartEpicResult{
		EpicID:         loadedEpic.ID,
		PreviousStatus: currentStatus,
		NewStatus:      LifecycleStatusWIP,
		StartedAt:      startTime,
		Message:        fmt.Sprintf("Epic %s started. Status changed to %s.", loadedEpic.ID, LifecycleStatusWIP),
		EventCreated:   false, // Event logging will be implemented in a later epic
	}, nil
}

// DoneEpic transitions an epic from wip to done status (simplified version of CompleteEpic)
func (ls *LifecycleService) DoneEpic(request DoneEpicRequest) (*DoneEpicResult, error) {
	// Load the epic
	loadedEpic, err := ls.storage.LoadEpic(request.EpicFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load epic: %w", err)
	}

	// Validate current status
	currentStatus := FromEpicStatus(loadedEpic.Status)
	if !currentStatus.CanTransitionTo(LifecycleStatusDone) {
		return nil, &TransitionError{
			EpicID:        loadedEpic.ID,
			CurrentStatus: currentStatus,
			TargetStatus:  LifecycleStatusDone,
			Message:       fmt.Sprintf("Epic cannot be completed from status: %s", currentStatus),
			Suggestion:    "Epic must be started first using 'agentpm start-epic'",
		}
	}

	// Validate completion requirements using enhanced validation
	if err := ls.validateCompletionRequirementsEnhanced(loadedEpic); err != nil {
		return nil, err
	}

	// Determine timestamp
	completedTime := ls.timeSource()
	if request.Timestamp != nil {
		completedTime = *request.Timestamp
	}

	// Calculate duration - for now we'll use a simple placeholder since StartedAt isn't in the epic model yet
	duration := time.Duration(0)

	// Update epic status
	loadedEpic.Status = LifecycleStatusDone.ToEpicStatus()

	// Create completion summary
	totalPhases := len(loadedEpic.Phases)
	totalTasks := len(loadedEpic.Tasks)
	totalTests := len(loadedEpic.Tests)
	totalItems := totalPhases + totalTasks + totalTests

	summary := fmt.Sprintf("Epic completed with %d phases, %d tasks, and %d tests (%d total items)",
		totalPhases, totalTasks, totalTests, totalItems)

	// Event logging will be implemented in a later epic

	// Save the updated epic
	if err := ls.storage.SaveEpic(loadedEpic, request.EpicFile); err != nil {
		return nil, fmt.Errorf("failed to save epic: %w", err)
	}

	return &DoneEpicResult{
		EpicID:         loadedEpic.ID,
		PreviousStatus: currentStatus,
		NewStatus:      LifecycleStatusDone,
		CompletedAt:    completedTime,
		Duration:       duration,
		Summary:        summary,
		Message:        fmt.Sprintf("Epic %s completed successfully. All phases and tests complete.", loadedEpic.ID),
		EventCreated:   false, // Event logging will be implemented in a later epic
	}, nil
}

// CompletionValidationError represents validation errors preventing epic completion
type CompletionValidationError struct {
	EpicID        string
	PendingPhases []PendingPhase
	FailingTests  []FailingTest
	Message       string
}

func (e *CompletionValidationError) Error() string {
	return e.Message
}

// PendingPhase represents a phase that is not completed
type PendingPhase struct {
	ID   string
	Name string
}

// FailingTest represents a test that is failing
type FailingTest struct {
	ID          string
	Name        string
	Description string
}

// DoneEpicRequest represents a request to mark an epic as done
type DoneEpicRequest struct {
	EpicFile  string
	Timestamp *time.Time // optional, for deterministic testing
}

// DoneEpicResult represents the result of marking an epic as done
type DoneEpicResult struct {
	EpicID         string
	PreviousStatus EpicLifecycleStatus
	NewStatus      EpicLifecycleStatus
	CompletedAt    time.Time
	Duration       time.Duration
	Summary        string
	Message        string
	EventCreated   bool
}

// CompleteEpicRequest represents a request to complete an epic
type CompleteEpicRequest struct {
	EpicFile  string
	Timestamp *time.Time // optional, for deterministic testing
}

// CompleteEpicResult represents the result of completing an epic
type CompleteEpicResult struct {
	EpicID         string
	PreviousStatus EpicLifecycleStatus
	NewStatus      EpicLifecycleStatus
	CompletedAt    time.Time
	Summary        EpicSummary
	Message        string
	EventCreated   bool
}

// EpicSummary provides statistics about the completed epic
type EpicSummary struct {
	TotalPhases int
	TotalTasks  int
	TotalTests  int
	Duration    string
}

// CompleteEpic transitions an epic from wip to done status
func (ls *LifecycleService) CompleteEpic(request CompleteEpicRequest) (*CompleteEpicResult, error) {
	// Load the epic
	loadedEpic, err := ls.storage.LoadEpic(request.EpicFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load epic: %w", err)
	}

	// Validate current status
	currentStatus := FromEpicStatus(loadedEpic.Status)
	if !currentStatus.CanTransitionTo(LifecycleStatusDone) {
		return nil, &TransitionError{
			EpicID:        loadedEpic.ID,
			CurrentStatus: currentStatus,
			TargetStatus:  LifecycleStatusDone,
			Message:       fmt.Sprintf("Epic cannot be completed from status: %s", currentStatus),
			Suggestion:    "Epic must be started first using 'agentpm start-epic'",
		}
	}

	// Validate completion requirements using enhanced validation
	if err := ls.validateCompletionRequirementsEnhanced(loadedEpic); err != nil {
		return nil, err
	}

	// Determine timestamp
	completedTime := ls.timeSource()
	if request.Timestamp != nil {
		completedTime = *request.Timestamp
	}

	// Calculate duration - for now we'll use a simple placeholder since StartedAt isn't in the epic model yet
	duration := "unknown"

	// Update epic status
	loadedEpic.Status = LifecycleStatusDone.ToEpicStatus()

	// Create summary
	summary := EpicSummary{
		TotalPhases: len(loadedEpic.Phases),
		TotalTasks:  len(loadedEpic.Tasks),
		TotalTests:  len(loadedEpic.Tests),
		Duration:    duration,
	}

	// Event logging will be implemented in a later epic

	// Save the updated epic
	if err := ls.storage.SaveEpic(loadedEpic, request.EpicFile); err != nil {
		return nil, fmt.Errorf("failed to save epic: %w", err)
	}

	return &CompleteEpicResult{
		EpicID:         loadedEpic.ID,
		PreviousStatus: currentStatus,
		NewStatus:      LifecycleStatusDone,
		CompletedAt:    completedTime,
		Summary:        summary,
		Message:        fmt.Sprintf("Epic %s completed successfully. All phases and tests complete.", loadedEpic.ID),
		EventCreated:   false, // Event logging will be implemented in a later epic
	}, nil
}

// validateCompletionRequirements checks if an epic can be completed
func (ls *LifecycleService) validateCompletionRequirements(loadedEpic *epic.Epic) error {
	var pendingPhases []PendingPhase
	var failingTests []FailingTest

	// Check phase completion
	for _, phase := range loadedEpic.Phases {
		if phase.Status != epic.StatusCompleted {
			pendingPhases = append(pendingPhases, PendingPhase{
				ID:   phase.ID,
				Name: phase.Name,
			})
		}
	}

	// Check test status (non-completed tests are considered failing)
	for _, test := range loadedEpic.Tests {
		if test.Status != epic.StatusCompleted {
			failingTests = append(failingTests, FailingTest{
				ID:          test.ID,
				Name:        test.Name,
				Description: test.Description,
			})
		}
	}

	// Return validation error if issues found
	if len(pendingPhases) > 0 || len(failingTests) > 0 {
		message := "Cannot complete epic with pending work"
		if len(pendingPhases) > 0 && len(failingTests) > 0 {
			message = fmt.Sprintf("Cannot complete epic: %d pending phases, %d failing tests",
				len(pendingPhases), len(failingTests))
		} else if len(pendingPhases) > 0 {
			message = fmt.Sprintf("Cannot complete epic: %d pending phases", len(pendingPhases))
		} else {
			message = fmt.Sprintf("Cannot complete epic: %d failing tests", len(failingTests))
		}

		return &CompletionValidationError{
			EpicID:        loadedEpic.ID,
			PendingPhases: pendingPhases,
			FailingTests:  failingTests,
			Message:       message,
		}
	}

	return nil
}

// TransitionError represents an invalid state transition error
type TransitionError struct {
	EpicID        string
	CurrentStatus EpicLifecycleStatus
	TargetStatus  EpicLifecycleStatus
	Message       string
	Suggestion    string
}

func (e *TransitionError) Error() string {
	return e.Message
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		if hours > 0 {
			return fmt.Sprintf("%d days, %d hours", days, hours)
		}
		return fmt.Sprintf("%d days", days)
	}
	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%d hours, %d minutes", hours, minutes)
		}
		return fmt.Sprintf("%d hours", hours)
	}
	if minutes > 0 {
		return fmt.Sprintf("%d minutes", minutes)
	}
	return "less than a minute"
}
