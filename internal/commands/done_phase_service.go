package commands

import (
	"fmt"
	"time"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/messages"
	"github.com/mindreframer/agentpm/internal/phases"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
)

type DonePhaseRequest struct {
	PhaseID    string
	ConfigPath string
	EpicFile   string
	Time       string
	Format     string
}

type DonePhaseResult struct {
	PhaseID            string
	Message            *messages.Message
	IsAlreadyCompleted bool
	Error              *PhaseError
}

func DonePhaseService(request DonePhaseRequest) (*DonePhaseResult, error) {
	if request.PhaseID == "" {
		return nil, fmt.Errorf("phase ID is required")
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
	phaseService := phases.NewPhaseService(storageImpl, queryService)

	// Load epic
	epicData, err := storageImpl.LoadEpic(epicFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load epic: %w", err)
	}

	// Complete the phase
	err = phaseService.CompletePhase(epicData, request.PhaseID, timestamp)
	if err != nil {
		// Handle different error types for better error output
		if incompleteErr, ok := err.(*phases.PhaseIncompleteError); ok {
			return &DonePhaseResult{
				PhaseID: request.PhaseID,
				Error: &PhaseError{
					Type:    "incomplete_phase",
					Message: fmt.Sprintf("Cannot complete phase %s: %d tasks are still pending", request.PhaseID, len(incompleteErr.PendingTasks)),
					Details: map[string]any{
						"phase_id":      request.PhaseID,
						"pending_tasks": convertTasksToDetails(incompleteErr.PendingTasks),
						"suggestion":    fmt.Sprintf("Complete or cancel all tasks in phase %s first", request.PhaseID),
					},
				},
			}, nil
		}

		if stateErr, ok := err.(*phases.PhaseStateError); ok {
			// Check if it's an "already completed" scenario
			if stateErr.CurrentStatus == epic.StatusCompleted {
				// Phase is already completed - return friendly success message
				templates := messages.NewMessageTemplates()
				message := templates.PhaseAlreadyCompleted(request.PhaseID)
				return &DonePhaseResult{
					PhaseID:            request.PhaseID,
					Message:            message,
					IsAlreadyCompleted: true,
				}, nil
			}
			return &DonePhaseResult{
				PhaseID: request.PhaseID,
				Error: &PhaseError{
					Type:    "invalid_phase_state",
					Message: fmt.Sprintf("Cannot complete phase %s: %s", request.PhaseID, stateErr.Message),
					Details: map[string]any{
						"phase_id":       request.PhaseID,
						"current_status": string(stateErr.CurrentStatus),
						"target_status":  string(stateErr.TargetStatus),
					},
				},
			}, nil
		}

		return nil, fmt.Errorf("failed to complete phase: %w", err)
	}

	// Save the updated epic
	err = storageImpl.SaveEpic(epicData, epicFile)
	if err != nil {
		return nil, fmt.Errorf("failed to save epic: %w", err)
	}

	return &DonePhaseResult{
		PhaseID: request.PhaseID,
	}, nil
}

func convertTasksToDetails(tasks []epic.Task) []map[string]string {
	result := make([]map[string]string, len(tasks))
	for i, task := range tasks {
		result[i] = map[string]string{
			"id":     task.ID,
			"name":   task.Name,
			"status": string(task.Status),
		}
	}
	return result
}
