package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/epic"
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

			// Complete the task
			err = taskService.CompleteTask(epicData, taskID, timestamp)
			if err != nil {
				// Handle different error types for better error output
				if stateErr, ok := err.(*tasks.TaskStateError); ok {
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
			if task.PhaseID == phaseID && task.Status == epic.StatusPlanning {
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
