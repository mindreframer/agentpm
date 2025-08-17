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

type StartEpicRequest struct {
	ConfigPath string
	EpicFile   string
	Time       string
	Format     string
}

type StartEpicResult struct {
	Result             *lifecycle.StartEpicResult
	Message            *messages.Message
	IsAlreadyStarted   bool
	IsAlreadyCompleted bool
}

func StartEpicService(request StartEpicRequest) (*StartEpicResult, error) {
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

	// Create start epic request
	lifecycleRequest := lifecycle.StartEpicRequest{
		EpicFile:  epicFile,
		Timestamp: timestamp,
	}

	// Start the epic
	result, err := lifecycleService.StartEpic(lifecycleRequest)
	if err != nil {
		// Check if it's an "already started/completed" scenario and handle with friendly message
		if transitionErr, ok := err.(*lifecycle.TransitionError); ok {
			templates := messages.NewMessageTemplates()
			if transitionErr.CurrentStatus == lifecycle.LifecycleStatusWIP {
				// Epic is already started - return friendly success message
				message := templates.EpicAlreadyStarted(transitionErr.EpicID)
				return &StartEpicResult{
					Message:          message,
					IsAlreadyStarted: true,
				}, nil
			} else if transitionErr.CurrentStatus == lifecycle.LifecycleStatusDone {
				// Epic is already completed - return friendly success message
				message := templates.EpicAlreadyCompleted(transitionErr.EpicID)
				return &StartEpicResult{
					Message:            message,
					IsAlreadyCompleted: true,
				}, nil
			}
		}
		return nil, err
	}

	// Generate success message for successful start
	templates := messages.NewMessageTemplates()
	message := templates.EpicStarted(result.EpicID)

	return &StartEpicResult{
		Result:  result,
		Message: message,
	}, nil
}
