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
	}
}

// PhaseConstraintError represents a constraint violation (e.g., multiple active phases)
type PhaseConstraintError struct {
	PhaseID       string
	ActivePhaseID string
	Message       string
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
	}
}

// PhaseIncompleteError represents attempting to complete a phase with pending tasks
type PhaseIncompleteError struct {
	PhaseID      string
	PendingTasks []epic.Task
}

func (e *PhaseIncompleteError) Error() string {
	return fmt.Sprintf("phase %s: cannot complete with %d pending tasks",
		e.PhaseID, len(e.PendingTasks))
}

func NewPhaseIncompleteError(phaseID string, pendingTasks []epic.Task) *PhaseIncompleteError {
	return &PhaseIncompleteError{
		PhaseID:      phaseID,
		PendingTasks: pendingTasks,
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
