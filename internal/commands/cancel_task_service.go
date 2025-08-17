package commands

import (
	"fmt"
	"time"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/mindreframer/agentpm/internal/tasks"
)

type CancelTaskRequest struct {
	TaskID     string
	ConfigPath string
	EpicFile   string
	Time       string
	Format     string
}

type CancelTaskResult struct {
	TaskID string
	Error  *TaskError
}

func CancelTaskService(request CancelTaskRequest) (*CancelTaskResult, error) {
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

	// Cancel the task
	err = taskService.CancelTask(epicData, request.TaskID, timestamp)
	if err != nil {
		// Handle different error types for better error output
		if stateErr, ok := err.(*tasks.TaskStateError); ok {
			return &CancelTaskResult{
				TaskID: request.TaskID,
				Error: &TaskError{
					Type:    "invalid_task_state",
					Message: fmt.Sprintf("Cannot cancel task %s: %s", request.TaskID, stateErr.Message),
					Details: map[string]any{
						"task_id":        request.TaskID,
						"current_status": string(stateErr.CurrentStatus),
						"target_status":  string(stateErr.TargetStatus),
					},
				},
			}, nil
		}

		return nil, fmt.Errorf("failed to cancel task: %w", err)
	}

	// Save the updated epic
	err = storageImpl.SaveEpic(epicData, epicFile)
	if err != nil {
		return nil, fmt.Errorf("failed to save epic: %w", err)
	}

	return &CancelTaskResult{
		TaskID: request.TaskID,
	}, nil
}
