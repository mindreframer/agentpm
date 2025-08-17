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
	"github.com/mindreframer/agentpm/internal/phases"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/urfave/cli/v3"
)

func DonePhaseCommand() *cli.Command {
	return &cli.Command{
		Name:      "done-phase",
		Usage:     "Complete a specific phase in the epic",
		ArgsUsage: "<phase-id>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "file",
				Usage: "Epic file path (overrides config)",
			},
			&cli.StringFlag{
				Name:  "time",
				Usage: "Timestamp for the phase completion (ISO 8601 format)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() < 1 {
				return fmt.Errorf("phase ID is required")
			}

			phaseID := cmd.Args().First()

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
			phaseService := phases.NewPhaseService(storageImpl, queryService)

			// Load epic
			epicData, err := storageImpl.LoadEpic(epicFile)
			if err != nil {
				return fmt.Errorf("failed to load epic: %w", err)
			}

			// Validate using Epic 13 framework before completing
			phaseValidationService := phases.NewPhaseValidationService()
			phase := findPhaseByID(epicData, phaseID)
			if phase == nil {
				return fmt.Errorf("phase %s not found", phaseID)
			}
			if validationErr := phaseValidationService.ValidatePhaseCompletion(epicData, phase); validationErr != nil {
				return outputPhaseEpic13ValidationError(cmd, validationErr, cmd.String("format"))
			}

			// Complete the phase
			err = phaseService.CompletePhase(epicData, phaseID, timestamp)
			if err != nil {
				// Handle different error types for better error output
				if incompleteErr, ok := err.(*phases.PhaseIncompleteError); ok {
					return outputPhaseIncompleteError(cmd, phaseID, incompleteErr.PendingTasks)
				}

				if stateErr, ok := err.(*phases.PhaseStateError); ok {
					// Check if it's an "already completed" scenario
					if stateErr.CurrentStatus == epic.StatusCompleted {
						// Phase is already completed - return friendly success message
						templates := messages.NewMessageTemplates()
						message := templates.PhaseAlreadyCompleted(phaseID)
						return outputFriendlyMessage(cmd, message, cmd.String("format"))
					}
					return outputXMLError(cmd, "invalid_phase_state",
						fmt.Sprintf("Cannot complete phase %s: %s", phaseID, stateErr.Message),
						map[string]interface{}{
							"phase_id":       phaseID,
							"current_status": string(stateErr.CurrentStatus),
							"target_status":  string(stateErr.TargetStatus),
						})
				}

				return fmt.Errorf("failed to complete phase: %w", err)
			}

			// Save the updated epic
			err = storageImpl.SaveEpic(epicData, epicFile)
			if err != nil {
				return fmt.Errorf("failed to save epic: %w", err)
			}

			// Output simple confirmation message
			fmt.Fprintf(cmd.Writer, "Phase %s completed.\n", phaseID)
			return nil
		},
	}
}

// outputPhaseIncompleteError outputs detailed error for incomplete phase
func outputPhaseIncompleteError(cmd *cli.Command, phaseID string, pendingTasks []epic.Task) error {
	fmt.Fprintf(cmd.ErrWriter, "<error>\n")
	fmt.Fprintf(cmd.ErrWriter, "    <type>incomplete_phase</type>\n")
	fmt.Fprintf(cmd.ErrWriter, "    <message>Cannot complete phase %s: %d tasks are still pending</message>\n", phaseID, len(pendingTasks))
	fmt.Fprintf(cmd.ErrWriter, "    <details>\n")
	fmt.Fprintf(cmd.ErrWriter, "        <phase_id>%s</phase_id>\n", phaseID)
	fmt.Fprintf(cmd.ErrWriter, "        <pending_tasks>\n")

	// Output pending tasks details (first few only)
	maxTasks := 3
	for i, task := range pendingTasks {
		if i >= maxTasks {
			fmt.Fprintf(cmd.ErrWriter, "            <task>... and %d more tasks</task>\n", len(pendingTasks)-maxTasks)
			break
		}
		// Output task details
		fmt.Fprintf(cmd.ErrWriter, "            <task id=\"%s\" status=\"%s\">%s</task>\n", task.ID, task.Status, task.Name)
	}

	fmt.Fprintf(cmd.ErrWriter, "        </pending_tasks>\n")
	fmt.Fprintf(cmd.ErrWriter, "        <suggestion>Complete or cancel all tasks in phase %s first</suggestion>\n", phaseID)
	fmt.Fprintf(cmd.ErrWriter, "    </details>\n")
	fmt.Fprintf(cmd.ErrWriter, "</error>\n")

	return fmt.Errorf("Error: Cannot complete phase %s: %d tasks are still pending", phaseID, len(pendingTasks))
}

// findPhaseByID finds a phase by its ID
func findPhaseByID(epicData *epic.Epic, phaseID string) *epic.Phase {
	for i := range epicData.Phases {
		if epicData.Phases[i].ID == phaseID {
			return &epicData.Phases[i]
		}
	}
	return nil
}

// outputPhaseEpic13ValidationError outputs Epic 13 validation errors in the specified format
func outputPhaseEpic13ValidationError(cmd *cli.Command, err error, format string) error {
	if validationErr, ok := err.(*epic.StatusValidationError); ok {
		switch format {
		case "json":
			return outputValidationErrorJSON(cmd, validationErr)
		case "xml":
			return outputValidationErrorXML(cmd, validationErr)
		default:
			return outputValidationErrorText(cmd, validationErr)
		}
	}

	// Fallback for other error types
	fmt.Fprintf(cmd.ErrWriter, "Error: %v\n", err)
	return err
}

// outputValidationErrorText outputs validation error in text format
func outputValidationErrorText(cmd *cli.Command, err *epic.StatusValidationError) error {
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

// outputValidationErrorJSON outputs validation error in JSON format
func outputValidationErrorJSON(cmd *cli.Command, err *epic.StatusValidationError) error {
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

// outputValidationErrorXML outputs validation error in XML format
func outputValidationErrorXML(cmd *cli.Command, err *epic.StatusValidationError) error {
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
