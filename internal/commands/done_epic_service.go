package commands

import (
	"fmt"
	"time"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/lifecycle"
	"github.com/mindreframer/agentpm/internal/messages"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
)

type DoneEpicRequest struct {
	ConfigPath string
	EpicFile   string
	Time       string
	Format     string
}

type DoneEpicResult struct {
	Result             *lifecycle.DoneEpicResult
	Message            *messages.Message
	IsAlreadyCompleted bool
	Error              *EpicError
}

type EpicError struct {
	Type         string
	Message      string
	Details      map[string]any
	IsTransition bool
	IsValidation bool
}

func DoneEpicService(request DoneEpicRequest) (*DoneEpicResult, error) {
	// Determine epic file to use
	epicFile := request.EpicFile
	if epicFile == "" {
		// Load configuration only if no file specified
		configPath := request.ConfigPath
		if configPath == "" {
			configPath = "./.agentpm.json"
		}

		cfg, configErr := config.LoadConfig(configPath)
		if configErr != nil {
			return nil, fmt.Errorf("failed to load configuration: %w", configErr)
		}

		if cfg.CurrentEpic == "" {
			return nil, fmt.Errorf("no epic file specified and no current epic in config")
		}
		epicFile = cfg.EpicFilePath()
	}

	// Initialize services
	storageFactory := storage.NewFactory(false) // Use file storage
	storageImpl := storageFactory.CreateStorage()
	queryService := query.NewQueryService(storageImpl)
	lifecycleService := lifecycle.NewLifecycleService(storageImpl, queryService)

	// Handle custom timestamp if provided
	var timestamp *time.Time
	if request.Time != "" {
		parsedTime, err := time.Parse(time.RFC3339, request.Time)
		if err != nil {
			return nil, fmt.Errorf("invalid time format: %w (use ISO 8601 format like 2025-08-16T15:30:00Z)", err)
		}
		timestamp = &parsedTime
	}

	// Create done epic request
	lifecycleRequest := lifecycle.DoneEpicRequest{
		EpicFile:  epicFile,
		Timestamp: timestamp,
	}

	// Complete the epic
	result, err := lifecycleService.DoneEpic(lifecycleRequest)
	if err != nil {
		// Check if it's an "already completed" scenario and handle with friendly message
		if transitionErr, ok := err.(*lifecycle.TransitionError); ok {
			if transitionErr.CurrentStatus == lifecycle.LifecycleStatusDone {
				// Epic is already completed - return friendly success message
				templates := messages.NewMessageTemplates()
				message := templates.EpicAlreadyCompleted(transitionErr.EpicID)
				return &DoneEpicResult{
					Message:            message,
					IsAlreadyCompleted: true,
				}, nil
			} else {
				// Other transition errors
				return &DoneEpicResult{
					Error: &EpicError{
						Type:         "invalid_transition",
						Message:      transitionErr.Message,
						IsTransition: true,
						Details: map[string]any{
							"epic_id":        transitionErr.EpicID,
							"current_status": transitionErr.CurrentStatus.String(),
							"target_status":  transitionErr.TargetStatus.String(),
							"suggestion":     transitionErr.Suggestion,
						},
					},
				}, nil
			}
		}

		// Check if it's a completion validation error for enhanced formatting
		if validationErr, ok := err.(*lifecycle.CompletionValidationError); ok {
			return &DoneEpicResult{
				Error: &EpicError{
					Type:         "completion_validation",
					Message:      validationErr.Message,
					IsValidation: true,
					Details: map[string]any{
						"epic_id":        validationErr.EpicID,
						"pending_phases": convertPendingPhasesToDetails(validationErr.PendingPhases),
						"failing_tests":  convertFailingTestsToDetails(validationErr.FailingTests),
					},
				},
			}, nil
		}

		return nil, err
	}

	return &DoneEpicResult{
		Result: result,
	}, nil
}

func convertPendingPhasesToDetails(phases []lifecycle.PendingPhase) []map[string]string {
	result := make([]map[string]string, len(phases))
	for i, phase := range phases {
		result[i] = map[string]string{
			"id":   phase.ID,
			"name": phase.Name,
		}
	}
	return result
}

func convertFailingTestsToDetails(tests []lifecycle.FailingTest) []map[string]string {
	result := make([]map[string]string, len(tests))
	for i, test := range tests {
		result[i] = map[string]string{
			"id":          test.ID,
			"name":        test.Name,
			"description": test.Description,
		}
	}
	return result
}
