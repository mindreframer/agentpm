package commands

import (
	"fmt"
	"time"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/messages"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/mindreframer/agentpm/internal/tasks"
)

type DoneTaskRequest struct {
	TaskID     string
	ConfigPath string
	EpicFile   string
	Time       string
	Format     string
}

type DoneTaskResult struct {
	TaskID             string
	Message            *messages.Message
	IsAlreadyCompleted bool
	Error              *TaskError
}

func DoneTaskService(request DoneTaskRequest) (*DoneTaskResult, error) {
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

	// Complete the task
	err = taskService.CompleteTask(epicData, request.TaskID, timestamp)
	if err != nil {
		// Handle different error types for better error output
		if stateErr, ok := err.(*tasks.TaskStateError); ok {
			// Check if it's an "already completed" scenario
			if stateErr.CurrentStatus == epic.StatusCompleted {
				// Task is already completed - return friendly success message
				templates := messages.NewMessageTemplates()
				message := templates.TaskAlreadyCompleted(request.TaskID)
				return &DoneTaskResult{
					TaskID:             request.TaskID,
					Message:            message,
					IsAlreadyCompleted: true,
				}, nil
			}
			return &DoneTaskResult{
				TaskID: request.TaskID,
				Error: &TaskError{
					Type:    "invalid_task_state",
					Message: fmt.Sprintf("Cannot complete task %s: %s", request.TaskID, stateErr.Message),
					Details: map[string]any{
						"task_id":        request.TaskID,
						"current_status": string(stateErr.CurrentStatus),
						"target_status":  string(stateErr.TargetStatus),
					},
				},
			}, nil
		}

		return nil, fmt.Errorf("failed to complete task: %w", err)
	}

	// Update current_state after completing task (Epic 7)
	updateCurrentStateAfterTaskComplete(epicData, request.TaskID)

	// Save the updated epic
	err = storageImpl.SaveEpic(epicData, epicFile)
	if err != nil {
		return nil, fmt.Errorf("failed to save epic: %w", err)
	}

	return &DoneTaskResult{
		TaskID: request.TaskID,
	}, nil
}

// updateCurrentStateAfterTaskComplete updates the epic's current_state when a task is completed
func updateCurrentStateAfterTaskComplete(epicData *epic.Epic, taskID string) {
	// Ensure current_state exists
	if epicData.CurrentState == nil {
		epicData.CurrentState = &epic.CurrentState{}
	}

	// If this was the active task, clear it
	if epicData.CurrentState.ActiveTask == taskID {
		epicData.CurrentState.ActiveTask = ""
	}

	// Find next action based on remaining work in the current phase
	phaseID := epicData.CurrentState.ActivePhase
	if phaseID != "" {
		// Look for next pending task in the same phase
		for _, task := range epicData.Tasks {
			if task.PhaseID == phaseID && task.Status == epic.StatusPending {
				epicData.CurrentState.NextAction = fmt.Sprintf("Start next task: %s", task.Name)
				return
			}
		}
		// No more tasks in phase, suggest completing the phase
		epicData.CurrentState.NextAction = "Complete current phase"
	} else {
		epicData.CurrentState.NextAction = "Start next phase"
	}
}
