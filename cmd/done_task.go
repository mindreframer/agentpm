package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/memomoo/agentpm/internal/config"
	"github.com/memomoo/agentpm/internal/query"
	"github.com/memomoo/agentpm/internal/storage"
	"github.com/memomoo/agentpm/internal/tasks"
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
