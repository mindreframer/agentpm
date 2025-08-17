package cmd

import (
	"context"
	"fmt"
	"time"

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
