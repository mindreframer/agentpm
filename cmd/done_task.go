package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beevik/etree"
	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/messages"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/mindreframer/agentpm/internal/tasks"
	"github.com/urfave/cli/v3"
)

func DoneTaskCommand() *cli.Command {
	return &cli.Command{
		Name:      "done-task",
		Usage:     "Complete a specific task in the epic",
		ArgsUsage: "<task-id>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "file",
				Usage: "Epic file path (overrides config)",
			},
			&cli.StringFlag{
				Name:  "time",
				Usage: "Timestamp for the task completion (ISO 8601 format)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() < 1 {
				return fmt.Errorf("task ID is required")
			}

			taskID := cmd.Args().First()

			// Get epic file path
			epicFile := cmd.String("file")
			if epicFile == "" {
				cfg, err := config.LoadConfig(cmd.String("config"))
				if err != nil {
					return fmt.Errorf("failed to load configuration: %w", err)
				}
				epicFile = cfg.CurrentEpic
			}

			if epicFile == "" {
				return fmt.Errorf("no epic file specified (use --file flag or set current epic)")
			}

			// Parse timestamp if provided
			var timestamp time.Time
			if timeStr := cmd.String("time"); timeStr != "" {
				var err error
				timestamp, err = time.Parse(time.RFC3339, timeStr)
				if err != nil {
					return fmt.Errorf("invalid time format: %s (use ISO 8601 format like 2025-08-16T15:30:00Z)", timeStr)
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
				return fmt.Errorf("failed to load epic: %w", err)
			}

			// Validate using Epic 13 framework before completing
			taskValidationService := tasks.NewTaskValidationService()
			task := findTaskByID(epicData, taskID)
			if task == nil {
				return fmt.Errorf("task %s not found", taskID)
			}
			if validationErr := taskValidationService.ValidateTaskCompletion(epicData, task); validationErr != nil {
				return outputEpic13ValidationError(cmd, validationErr, cmd.String("format"))
			}

			// Complete the task
			err = taskService.CompleteTask(epicData, taskID, timestamp)
			if err != nil {
				// Handle different error types for better error output
				if stateErr, ok := err.(*tasks.TaskStateError); ok {
					// Check if it's an "already completed" scenario
					if stateErr.CurrentStatus == epic.StatusCompleted {
						// Task is already completed - return friendly success message
						templates := messages.NewMessageTemplates()
						message := templates.TaskAlreadyCompleted(taskID)
						return outputFriendlyMessage(cmd, message, cmd.String("format"))
					}
					return outputXMLError(cmd, "invalid_task_state",
						fmt.Sprintf("Cannot complete task %s: %s", taskID, stateErr.Message),
						map[string]interface{}{
							"task_id":        taskID,
							"current_status": string(stateErr.CurrentStatus),
							"target_status":  string(stateErr.TargetStatus),
						})
				}

				return fmt.Errorf("failed to complete task: %w", err)
			}

			// Update current_state after completing task (Epic 7)
			updateCurrentStateAfterTaskComplete(epicData, taskID)

			// Save the updated epic
			err = storageImpl.SaveEpic(epicData, epicFile)
			if err != nil {
				return fmt.Errorf("failed to save epic: %w", err)
			}

			// Output simple confirmation message
			fmt.Fprintf(cmd.Writer, "Task %s completed.\n", taskID)
			return nil
		},
	}
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

// findTaskByID finds a task by its ID
func findTaskByID(epicData *epic.Epic, taskID string) *epic.Task {
	for i := range epicData.Tasks {
		if epicData.Tasks[i].ID == taskID {
			return &epicData.Tasks[i]
		}
	}
	return nil
}

// outputEpic13ValidationError outputs Epic 13 validation errors in the specified format
func outputEpic13ValidationError(cmd *cli.Command, err error, format string) error {
	if validationErr, ok := err.(*epic.StatusValidationError); ok {
		switch format {
		case "json":
			return outputTaskValidationErrorJSON(cmd, validationErr)
		case "xml":
			return outputTaskValidationErrorXML(cmd, validationErr)
		default:
			return outputTaskValidationErrorText(cmd, validationErr)
		}
	}

	// Fallback for other error types
	fmt.Fprintf(cmd.ErrWriter, "Error: %v\n", err)
	return err
}

// outputTaskValidationErrorText outputs validation error in text format
func outputTaskValidationErrorText(cmd *cli.Command, err *epic.StatusValidationError) error {
	fmt.Fprintf(cmd.ErrWriter, "Error: %s\n", err.Message)
	fmt.Fprintf(cmd.ErrWriter, "Entity: %s %s (%s)\n", err.EntityType, err.EntityName, err.EntityID)
	fmt.Fprintf(cmd.ErrWriter, "Status: %s → %s\n", err.CurrentStatus, err.TargetStatus)

	if len(err.BlockingItems) > 0 {
		fmt.Fprintf(cmd.ErrWriter, "Blocking items: %d\n", len(err.BlockingItems))
		for i, item := range err.BlockingItems {
			if i >= 3 {
				fmt.Fprintf(cmd.ErrWriter, "  ... and %d more items\n", len(err.BlockingItems)-3)
				break
			}
			fmt.Fprintf(cmd.ErrWriter, "  - %s %s (%s): %s\n", item.Type, item.Name, item.ID, item.Status)
		}
	}

	return err
}

// outputTaskValidationErrorJSON outputs validation error in JSON format
func outputTaskValidationErrorJSON(cmd *cli.Command, err *epic.StatusValidationError) error {
	output := map[string]interface{}{
		"error": map[string]interface{}{
			"type":           "epic13_validation",
			"entity_type":    err.EntityType,
			"entity_id":      err.EntityID,
			"entity_name":    err.EntityName,
			"current_status": err.CurrentStatus,
			"target_status":  err.TargetStatus,
			"message":        err.Message,
			"blocking_items": err.BlockingItems,
		},
	}

	encoder := json.NewEncoder(cmd.ErrWriter)
	encoder.SetIndent("", "  ")
	encoder.Encode(output)
	return err
}

// outputTaskValidationErrorXML outputs validation error in XML format
func outputTaskValidationErrorXML(cmd *cli.Command, err *epic.StatusValidationError) error {
	doc := etree.NewDocument()
	root := doc.CreateElement("error")

	errorType := root.CreateElement("type")
	errorType.SetText("epic13_validation")

	entityType := root.CreateElement("entity_type")
	entityType.SetText(err.EntityType)

	entityID := root.CreateElement("entity_id")
	entityID.SetText(err.EntityID)

	entityName := root.CreateElement("entity_name")
	entityName.SetText(err.EntityName)

	currentStatus := root.CreateElement("current_status")
	currentStatus.SetText(err.CurrentStatus)

	targetStatus := root.CreateElement("target_status")
	targetStatus.SetText(err.TargetStatus)

	message := root.CreateElement("message")
	message.SetText(err.Message)

	if len(err.BlockingItems) > 0 {
		blockingItems := root.CreateElement("blocking_items")
		for _, item := range err.BlockingItems {
			itemElem := blockingItems.CreateElement("item")
			itemElem.CreateAttr("type", item.Type)
			itemElem.CreateAttr("id", item.ID)
			itemElem.CreateAttr("status", item.Status)
			itemElem.SetText(item.Name)
		}
	}

	doc.Indent(4)
	doc.WriteTo(cmd.ErrWriter)
	fmt.Fprintf(cmd.ErrWriter, "\n")
	return err
}
