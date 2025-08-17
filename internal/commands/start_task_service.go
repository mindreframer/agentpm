package commands

import (
	"fmt"
	"time"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/hints"
	"github.com/mindreframer/agentpm/internal/messages"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/mindreframer/agentpm/internal/tasks"
)

type StartTaskRequest struct {
	TaskID     string
	ConfigPath string
	EpicFile   string
	Time       string
	Format     string
}

type StartTaskResult struct {
	TaskID          string
	Message         *messages.Message
	IsAlreadyActive bool
	Error           *TaskError
}

type TaskError struct {
	Type    string
	Message string
	Details map[string]any
	Hint    string
}

func StartTaskService(request StartTaskRequest) (*StartTaskResult, error) {
	if request.TaskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}

	// Get epic file path
	epicFile := request.EpicFile
	if epicFile == "" {
		configPath := request.ConfigPath
		if configPath == "" {
			configPath = "./.agentpm.json"
		}

		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load configuration: %w", err)
		}
		epicFile = cfg.CurrentEpic
	}

	if epicFile == "" {
		return nil, fmt.Errorf("no epic file specified (use --file flag or set current epic)")
	}

	// Parse timestamp if provided
	var timestamp time.Time
	if request.Time != "" {
		var err error
		timestamp, err = time.Parse(time.RFC3339, request.Time)
		if err != nil {
			return nil, fmt.Errorf("invalid time format: %s (use ISO 8601 format like 2025-08-16T15:30:00Z)", request.Time)
		}
	} else {
		timestamp = time.Now()
	}

	// Initialize services
	storageImpl := storage.NewFileStorage()
	queryService := query.NewQueryService(storageImpl)
	taskService := tasks.NewTaskService(storageImpl, queryService)

	// Load epic
	epicData, err := storageImpl.LoadEpic(epicFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load epic: %w", err)
	}

	// Start the task
	err = taskService.StartTask(epicData, request.TaskID, timestamp)
	if err != nil {
		// Handle different error types for better error output
		if _, ok := err.(*tasks.TaskAlreadyActiveError); ok {
			// Task is already active - return friendly success message
			templates := messages.NewMessageTemplates()
			message := templates.TaskAlreadyActive(request.TaskID)
			return &StartTaskResult{
				TaskID:          request.TaskID,
				Message:         message,
				IsAlreadyActive: true,
			}, nil
		}

		if phaseErr, ok := err.(*tasks.TaskPhaseError); ok {
			// Generate context-aware hint for task phase violations
			hintCtx := &hints.HintContext{
				ErrorType:     "TaskPhaseError",
				OperationType: "start",
				EntityType:    "task",
				EntityID:      request.TaskID,
				AdditionalData: map[string]any{
					"phase_id":     phaseErr.PhaseID,
					"phase_status": phaseErr.PhaseStatus,
				},
			}

			hintRegistry := hints.DefaultHintRegistry()
			hint := hintRegistry.GenerateHint(hintCtx)

			var hintText string
			if hint != nil {
				hintText = hint.Content
			}

			return &StartTaskResult{
				TaskID: request.TaskID,
				Error: &TaskError{
					Type:    "task_phase_violation",
					Message: fmt.Sprintf("Cannot start task %s: phase %s is not active", request.TaskID, phaseErr.PhaseID),
					Details: map[string]any{
						"task_id":      request.TaskID,
						"phase_id":     phaseErr.PhaseID,
						"phase_status": string(phaseErr.PhaseStatus),
						"suggestion":   fmt.Sprintf("Start phase %s first or use 'agentpm current' to see active work", phaseErr.PhaseID),
					},
					Hint: hintText,
				},
			}, nil
		}

		if constraintErr, ok := err.(*tasks.TaskConstraintError); ok {
			// Generate context-aware hint for task constraint violations
			hintCtx := &hints.HintContext{
				ErrorType:     "TaskConstraintError",
				OperationType: "start",
				EntityType:    "task",
				EntityID:      request.TaskID,
				AdditionalData: map[string]any{
					"active_task_id": constraintErr.ActiveTaskID,
					"phase_id":       constraintErr.PhaseID,
				},
			}

			hintRegistry := hints.DefaultHintRegistry()
			hint := hintRegistry.GenerateHint(hintCtx)

			var hintText string
			if hint != nil {
				hintText = hint.Content
			}

			return &StartTaskResult{
				TaskID: request.TaskID,
				Error: &TaskError{
					Type:    "task_constraint_violation",
					Message: fmt.Sprintf("Cannot start task %s: task %s is already active in phase %s", request.TaskID, constraintErr.ActiveTaskID, constraintErr.PhaseID),
					Details: map[string]any{
						"task_id":        request.TaskID,
						"active_task_id": constraintErr.ActiveTaskID,
						"phase_id":       constraintErr.PhaseID,
						"suggestion":     fmt.Sprintf("Complete task %s first or use 'agentpm current' to see active work", constraintErr.ActiveTaskID),
					},
					Hint: hintText,
				},
			}, nil
		}

		if stateErr, ok := err.(*tasks.TaskStateError); ok {
			// Generate context-aware hint for task state errors
			hintCtx := &hints.HintContext{
				ErrorType:     "TaskStateError",
				OperationType: "start",
				EntityType:    "task",
				EntityID:      request.TaskID,
				CurrentStatus: string(stateErr.CurrentStatus),
				TargetStatus:  string(stateErr.TargetStatus),
				AdditionalData: map[string]any{
					"current_status": stateErr.CurrentStatus,
					"target_status":  stateErr.TargetStatus,
				},
			}

			hintRegistry := hints.DefaultHintRegistry()
			hint := hintRegistry.GenerateHint(hintCtx)

			var hintText string
			if hint != nil {
				hintText = hint.Content
			}

			return &StartTaskResult{
				TaskID: request.TaskID,
				Error: &TaskError{
					Type:    "invalid_task_state",
					Message: fmt.Sprintf("Cannot start task %s: %s", request.TaskID, stateErr.Message),
					Details: map[string]any{
						"task_id":        request.TaskID,
						"current_status": string(stateErr.CurrentStatus),
						"target_status":  string(stateErr.TargetStatus),
					},
					Hint: hintText,
				},
			}, nil
		}

		return nil, fmt.Errorf("failed to start task: %w", err)
	}

	// Update current_state after starting task (Epic 7)
	updateCurrentStateAfterTaskStart(epicData, request.TaskID)

	// Save the updated epic
	err = storageImpl.SaveEpic(epicData, epicFile)
	if err != nil {
		return nil, fmt.Errorf("failed to save epic: %w", err)
	}

	return &StartTaskResult{
		TaskID: request.TaskID,
	}, nil
}

// updateCurrentStateAfterTaskStart updates the epic's current_state when a task is started
func updateCurrentStateAfterTaskStart(epicData *epic.Epic, taskID string) {
	// Ensure current_state exists
	if epicData.CurrentState == nil {
		epicData.CurrentState = &epic.CurrentState{}
	}

	// Find the task to get its phase
	var taskPhaseID string
	var taskName string
	for _, task := range epicData.Tasks {
		if task.ID == taskID {
			taskPhaseID = task.PhaseID
			taskName = task.Name
			break
		}
	}

	// Update current state
	epicData.CurrentState.ActivePhase = taskPhaseID
	epicData.CurrentState.ActiveTask = taskID
	epicData.CurrentState.NextAction = fmt.Sprintf("Continue work on: %s", taskName)
}
