package tasks

import (
	"fmt"

	"github.com/mindreframer/agentpm/internal/epic"
)

// TaskStateError represents an invalid task state transition
type TaskStateError struct {
	TaskID        string
	CurrentStatus epic.Status
	TargetStatus  epic.Status
	Message       string
}

func (e *TaskStateError) Error() string {
	return fmt.Sprintf("task %s: cannot transition from %s to %s: %s",
		e.TaskID, e.CurrentStatus, e.TargetStatus, e.Message)
}

func NewTaskStateError(taskID string, current, target epic.Status, message string) *TaskStateError {
	return &TaskStateError{
		TaskID:        taskID,
		CurrentStatus: current,
		TargetStatus:  target,
		Message:       message,
	}
}

// TaskPhaseError represents a task operation attempted on inactive phase
type TaskPhaseError struct {
	TaskID      string
	PhaseID     string
	PhaseStatus epic.Status
	Message     string
}

func (e *TaskPhaseError) Error() string {
	return fmt.Sprintf("task %s: phase %s error (status: %s): %s",
		e.TaskID, e.PhaseID, e.PhaseStatus, e.Message)
}

func NewTaskPhaseError(taskID, phaseID string, phaseStatus epic.Status, message string) *TaskPhaseError {
	return &TaskPhaseError{
		TaskID:      taskID,
		PhaseID:     phaseID,
		PhaseStatus: phaseStatus,
		Message:     message,
	}
}

// TaskConstraintError represents a constraint violation (e.g., multiple active tasks in phase)
type TaskConstraintError struct {
	TaskID       string
	ActiveTaskID string
	PhaseID      string
	Message      string
}

func (e *TaskConstraintError) Error() string {
	return fmt.Sprintf("task %s: constraint violation in phase %s: %s (active task: %s)",
		e.TaskID, e.PhaseID, e.Message, e.ActiveTaskID)
}

func NewTaskConstraintError(taskID, activeTaskID, phaseID, message string) *TaskConstraintError {
	return &TaskConstraintError{
		TaskID:       taskID,
		ActiveTaskID: activeTaskID,
		PhaseID:      phaseID,
		Message:      message,
	}
}

// TaskAlreadyActiveError represents attempting to start a task that is already active
type TaskAlreadyActiveError struct {
	TaskID string
}

func (e *TaskAlreadyActiveError) Error() string {
	return fmt.Sprintf("task %s is already active", e.TaskID)
}

func NewTaskAlreadyActiveError(taskID string) *TaskAlreadyActiveError {
	return &TaskAlreadyActiveError{
		TaskID: taskID,
	}
}
