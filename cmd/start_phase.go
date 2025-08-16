package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/phases"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/urfave/cli/v3"
)

func StartPhaseCommand() *cli.Command {
	return &cli.Command{
		Name:      "start-phase",
		Usage:     "Start a specific phase in the epic",
		ArgsUsage: "<phase-id>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "file",
				Usage: "Epic file path (overrides config)",
			},
			&cli.StringFlag{
				Name:  "time",
				Usage: "Timestamp for the phase start (ISO 8601 format)",
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

			// Start the phase
			err = phaseService.StartPhase(epicData, phaseID, timestamp)
			if err != nil {
				// Handle different error types for better error output
				if phaseErr, ok := err.(*phases.PhaseConstraintError); ok {
					return outputXMLError(cmd, "phase_constraint_violation",
						fmt.Sprintf("Cannot start phase %s: phase %s is still active", phaseID, phaseErr.ActivePhaseID),
						map[string]interface{}{
							"active_phase": phaseErr.ActivePhaseID,
							"suggestion":   fmt.Sprintf("Complete phase %s first or use 'agentpm current' to see active work", phaseErr.ActivePhaseID),
						})
				}

				if stateErr, ok := err.(*phases.PhaseStateError); ok {
					return outputXMLError(cmd, "invalid_phase_state",
						fmt.Sprintf("Cannot start phase %s: %s", phaseID, stateErr.Message),
						map[string]interface{}{
							"phase_id":       phaseID,
							"current_status": string(stateErr.CurrentStatus),
							"target_status":  string(stateErr.TargetStatus),
						})
				}

				return fmt.Errorf("failed to start phase: %w", err)
			}

			// Save the updated epic
			err = storageImpl.SaveEpic(epicData, epicFile)
			if err != nil {
				return fmt.Errorf("failed to save epic: %w", err)
			}

			// Output simple confirmation message
			fmt.Fprintf(cmd.Writer, "Phase %s started.\n", phaseID)
			return nil
		},
	}
}

// outputXMLError outputs structured XML error messages
func outputXMLError(cmd *cli.Command, errorType, message string, details map[string]interface{}) error {
	// Write XML error to stderr for structured error handling
	fmt.Fprintf(cmd.ErrWriter, "<error>\n")
	fmt.Fprintf(cmd.ErrWriter, "    <type>%s</type>\n", errorType)
	fmt.Fprintf(cmd.ErrWriter, "    <message>%s</message>\n", message)
	if len(details) > 0 {
		fmt.Fprintf(cmd.ErrWriter, "    <details>\n")
		for key, value := range details {
			fmt.Fprintf(cmd.ErrWriter, "        <%s>%v</%s>\n", key, value, key)
		}
		fmt.Fprintf(cmd.ErrWriter, "    </details>\n")
	}
	fmt.Fprintf(cmd.ErrWriter, "</error>\n")

	// Return a basic error for exit code
	return fmt.Errorf("Error: %s", message)
}
