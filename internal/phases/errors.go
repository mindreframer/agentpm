package phases

import (
	"fmt"

	"github.com/mindreframer/agentpm/internal/epic"
)

// PhaseStateError represents an invalid phase state transition
type PhaseStateError struct {
	PhaseID       string
	CurrentStatus epic.Status
	TargetStatus  epic.Status
	Message       string
	Hint          string // Actionable hint for resolving the error
}

func (e *PhaseStateError) Error() string {
	return fmt.Sprintf("phase %s: cannot transition from %s to %s: %s",
		e.PhaseID, e.CurrentStatus, e.TargetStatus, e.Message)
}

func NewPhaseStateError(phaseID string, current, target epic.Status, message string) *PhaseStateError {
	return &PhaseStateError{
		PhaseID:       phaseID,
		CurrentStatus: current,
		TargetStatus:  target,
		Message:       message,
		Hint:          "", // Will be populated by hint generator
	}
}

func NewPhaseStateErrorWithHint(phaseID string, current, target epic.Status, message, hint string) *PhaseStateError {
	return &PhaseStateError{
		PhaseID:       phaseID,
		CurrentStatus: current,
		TargetStatus:  target,
		Message:       message,
		Hint:          hint,
	}
}

// PhaseConstraintError represents a constraint violation (e.g., multiple active phases)
type PhaseConstraintError struct {
	PhaseID       string
	ActivePhaseID string
	Message       string
	Hint          string // Actionable hint for resolving the constraint violation
}

func (e *PhaseConstraintError) Error() string {
	return fmt.Sprintf("phase %s: constraint violation: %s (active phase: %s)",
		e.PhaseID, e.Message, e.ActivePhaseID)
}

func NewPhaseConstraintError(phaseID, activePhaseID, message string) *PhaseConstraintError {
	return &PhaseConstraintError{
		PhaseID:       phaseID,
		ActivePhaseID: activePhaseID,
		Message:       message,
		Hint:          "", // Will be populated by hint generator
	}
}

func NewPhaseConstraintErrorWithHint(phaseID, activePhaseID, message, hint string) *PhaseConstraintError {
	return &PhaseConstraintError{
		PhaseID:       phaseID,
		ActivePhaseID: activePhaseID,
		Message:       message,
		Hint:          hint,
	}
}

// PhaseIncompleteError represents attempting to complete a phase with pending tasks
type PhaseIncompleteError struct {
	PhaseID      string
	PendingTasks []epic.Task
	Hint         string // Actionable hint for resolving pending tasks
}

func (e *PhaseIncompleteError) Error() string {
	return fmt.Sprintf("phase %s: cannot complete with %d pending tasks",
		e.PhaseID, len(e.PendingTasks))
}

func NewPhaseIncompleteError(phaseID string, pendingTasks []epic.Task) *PhaseIncompleteError {
	return &PhaseIncompleteError{
		PhaseID:      phaseID,
		PendingTasks: pendingTasks,
		Hint:         "", // Will be populated by hint generator
	}
}

func NewPhaseIncompleteErrorWithHint(phaseID string, pendingTasks []epic.Task, hint string) *PhaseIncompleteError {
	return &PhaseIncompleteError{
		PhaseID:      phaseID,
		PendingTasks: pendingTasks,
		Hint:         hint,
	}
}

// PhaseAlreadyActiveError represents attempting to start a phase that is already active
type PhaseAlreadyActiveError struct {
	PhaseID string
}

func (e *PhaseAlreadyActiveError) Error() string {
	return fmt.Sprintf("phase %s is already active", e.PhaseID)
}

func NewPhaseAlreadyActiveError(phaseID string) *PhaseAlreadyActiveError {
	return &PhaseAlreadyActiveError{
		PhaseID: phaseID,
	}
}

// PhaseTestDependencyError represents attempting to complete a phase with incomplete tests
type PhaseTestDependencyError struct {
	PhaseID         string
	IncompleteTests []epic.Test
	Hint            string // Actionable hint for resolving incomplete tests
}

func (e *PhaseTestDependencyError) Error() string {
	return fmt.Sprintf("phase %s: cannot complete with %d incomplete tests",
		e.PhaseID, len(e.IncompleteTests))
}

func NewPhaseTestDependencyError(phaseID string, incompleteTests []epic.Test) *PhaseTestDependencyError {
	return &PhaseTestDependencyError{
		PhaseID:         phaseID,
		IncompleteTests: incompleteTests,
		Hint:            "", // Will be populated by hint generator
	}
}

func NewPhaseTestDependencyErrorWithHint(phaseID string, incompleteTests []epic.Test, hint string) *PhaseTestDependencyError {
	return &PhaseTestDependencyError{
		PhaseID:         phaseID,
		IncompleteTests: incompleteTests,
		Hint:            hint,
	}
}

// PhaseTestPrerequisiteError represents attempting to start a phase with incomplete prerequisite tests
type PhaseTestPrerequisiteError struct {
	PhaseID           string
	PrerequisiteTests []epic.Test
	Hint              string // Actionable hint for resolving prerequisite tests
}

func (e *PhaseTestPrerequisiteError) Error() string {
	return fmt.Sprintf("phase %s: cannot start with %d incomplete prerequisite tests",
		e.PhaseID, len(e.PrerequisiteTests))
}

func NewPhaseTestPrerequisiteError(phaseID string, prerequisiteTests []epic.Test) *PhaseTestPrerequisiteError {
	return &PhaseTestPrerequisiteError{
		PhaseID:           phaseID,
		PrerequisiteTests: prerequisiteTests,
		Hint:              "", // Will be populated by hint generator
	}
}

func NewPhaseTestPrerequisiteErrorWithHint(phaseID string, prerequisiteTests []epic.Test, hint string) *PhaseTestPrerequisiteError {
	return &PhaseTestPrerequisiteError{
		PhaseID:           phaseID,
		PrerequisiteTests: prerequisiteTests,
		Hint:              hint,
	}
}
