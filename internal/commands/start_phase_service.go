package commands

import (
	"fmt"
	"time"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/hints"
	"github.com/mindreframer/agentpm/internal/messages"
	"github.com/mindreframer/agentpm/internal/phases"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
)

type StartPhaseRequest struct {
	PhaseID    string
	ConfigPath string
	EpicFile   string
	Time       string
	Format     string
}

type StartPhaseResult struct {
	PhaseID         string
	Message         *messages.Message
	IsAlreadyActive bool
	Error           *PhaseError
}

type PhaseError struct {
	Type    string
	Message string
	Details map[string]any
	Hint    string
}

func StartPhaseService(request StartPhaseRequest) (*StartPhaseResult, error) {
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

	// Start the phase
	err = phaseService.StartPhase(epicData, request.PhaseID, timestamp)
	if err != nil {
		// Handle different error types for better error output
		if _, ok := err.(*phases.PhaseAlreadyActiveError); ok {
			// Phase is already active - return friendly success message
			templates := messages.NewMessageTemplates()
			message := templates.PhaseAlreadyActive(request.PhaseID)
			return &StartPhaseResult{
				PhaseID:         request.PhaseID,
				Message:         message,
				IsAlreadyActive: true,
			}, nil
		}

		if phaseErr, ok := err.(*phases.PhaseConstraintError); ok {
			// Generate context-aware hint for phase constraint violations
			hintCtx := &hints.HintContext{
				ErrorType:     "PhaseConstraintError",
				OperationType: "start",
				EntityType:    "phase",
				EntityID:      request.PhaseID,
				AdditionalData: map[string]any{
					"active_phase": phaseErr.ActivePhaseID,
				},
			}

			hintRegistry := hints.DefaultHintRegistry()
			hint := hintRegistry.GenerateHint(hintCtx)

			var hintText string
			if hint != nil {
				hintText = hint.Content
			}

			return &StartPhaseResult{
				PhaseID: request.PhaseID,
				Error: &PhaseError{
					Type:    "phase_constraint_violation",
					Message: fmt.Sprintf("Cannot start phase %s: phase %s is still active", request.PhaseID, phaseErr.ActivePhaseID),
					Details: map[string]any{
						"active_phase": phaseErr.ActivePhaseID,
						"suggestion":   fmt.Sprintf("Complete phase %s first or use 'agentpm current' to see active work", phaseErr.ActivePhaseID),
					},
					Hint: hintText,
				},
			}, nil
		}

		if stateErr, ok := err.(*phases.PhaseStateError); ok {
			// Generate context-aware hint for phase state errors
			hintCtx := &hints.HintContext{
				ErrorType:     "PhaseStateError",
				OperationType: "start",
				EntityType:    "phase",
				EntityID:      request.PhaseID,
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

			return &StartPhaseResult{
				PhaseID: request.PhaseID,
				Error: &PhaseError{
					Type:    "invalid_phase_state",
					Message: fmt.Sprintf("Cannot start phase %s: %s", request.PhaseID, stateErr.Message),
					Details: map[string]any{
						"phase_id":       request.PhaseID,
						"current_status": string(stateErr.CurrentStatus),
						"target_status":  string(stateErr.TargetStatus),
					},
					Hint: hintText,
				},
			}, nil
		}

		return nil, fmt.Errorf("failed to start phase: %w", err)
	}

	// Save the updated epic
	err = storageImpl.SaveEpic(epicData, epicFile)
	if err != nil {
		return nil, fmt.Errorf("failed to save epic: %w", err)
	}

	return &StartPhaseResult{
		PhaseID: request.PhaseID,
	}, nil
}
