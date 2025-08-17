package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/hints"
	"github.com/mindreframer/agentpm/internal/messages"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/mindreframer/agentpm/internal/tasks"
	"github.com/urfave/cli/v3"
)

func StartTaskCommand() *cli.Command {
	return &cli.Command{
		Name:      "start-task",
		Usage:     "Start a specific task in the epic",
		ArgsUsage: "<task-id>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "file",
				Usage: "Epic file path (overrides config)",
			},
			&cli.StringFlag{
				Name:  "time",
				Usage: "Timestamp for the task start (ISO 8601 format)",
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

			// Start the task
			err = taskService.StartTask(epicData, taskID, timestamp)
			if err != nil {
				// Handle different error types for better error output
				if _, ok := err.(*tasks.TaskAlreadyActiveError); ok {
					// Task is already active - return friendly success message
					templates := messages.NewMessageTemplates()
					message := templates.TaskAlreadyActive(taskID)
					return outputFriendlyMessage(cmd, message, cmd.String("format"))
				}

				if phaseErr, ok := err.(*tasks.TaskPhaseError); ok {
					// Generate context-aware hint for task phase violations
					hintCtx := &hints.HintContext{
						ErrorType:     "TaskPhaseError",
						OperationType: "start",
						EntityType:    "task",
						EntityID:      taskID,
						AdditionalData: map[string]interface{}{
							"phase_id":     phaseErr.PhaseID,
							"phase_status": phaseErr.PhaseStatus,
						},
					}

					hintRegistry := hints.DefaultHintRegistry()
					hint := hintRegistry.GenerateHint(hintCtx)

					var hintText string
					if hint != nil {
						hintText = hint.Content
					}

					return outputXMLErrorWithHint(cmd, "task_phase_violation",
						fmt.Sprintf("Cannot start task %s: phase %s is not active", taskID, phaseErr.PhaseID),
						map[string]interface{}{
							"task_id":      taskID,
							"phase_id":     phaseErr.PhaseID,
							"phase_status": string(phaseErr.PhaseStatus),
							"suggestion":   fmt.Sprintf("Start phase %s first or use 'agentpm current' to see active work", phaseErr.PhaseID),
						}, hintText)
				}

				if constraintErr, ok := err.(*tasks.TaskConstraintError); ok {
					// Generate context-aware hint for task constraint violations
					hintCtx := &hints.HintContext{
						ErrorType:     "TaskConstraintError",
						OperationType: "start",
						EntityType:    "task",
						EntityID:      taskID,
						AdditionalData: map[string]interface{}{
							"active_task_id": constraintErr.ActiveTaskID,
							"phase_id":       constraintErr.PhaseID,
						},
					}

					hintRegistry := hints.DefaultHintRegistry()
					hint := hintRegistry.GenerateHint(hintCtx)

					var hintText string
					if hint != nil {
						hintText = hint.Content
					}

					return outputXMLErrorWithHint(cmd, "task_constraint_violation",
						fmt.Sprintf("Cannot start task %s: task %s is already active in phase %s", taskID, constraintErr.ActiveTaskID, constraintErr.PhaseID),
						map[string]interface{}{
							"task_id":        taskID,
							"active_task_id": constraintErr.ActiveTaskID,
							"phase_id":       constraintErr.PhaseID,
							"suggestion":     fmt.Sprintf("Complete task %s first or use 'agentpm current' to see active work", constraintErr.ActiveTaskID),
						}, hintText)
				}

				if stateErr, ok := err.(*tasks.TaskStateError); ok {
					// Generate context-aware hint for task state errors
					hintCtx := &hints.HintContext{
						ErrorType:     "TaskStateError",
						OperationType: "start",
						EntityType:    "task",
						EntityID:      taskID,
						CurrentStatus: string(stateErr.CurrentStatus),
						TargetStatus:  string(stateErr.TargetStatus),
						AdditionalData: map[string]interface{}{
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

					return outputXMLErrorWithHint(cmd, "invalid_task_state",
						fmt.Sprintf("Cannot start task %s: %s", taskID, stateErr.Message),
						map[string]interface{}{
							"task_id":        taskID,
							"current_status": string(stateErr.CurrentStatus),
							"target_status":  string(stateErr.TargetStatus),
						}, hintText)
				}

				return fmt.Errorf("failed to start task: %w", err)
			}

			// Update current_state after starting task (Epic 7)
			updateCurrentStateAfterTaskStart(epicData, taskID)

			// Save the updated epic
			err = storageImpl.SaveEpic(epicData, epicFile)
			if err != nil {
				return fmt.Errorf("failed to save epic: %w", err)
			}

			// Output simple confirmation message
			fmt.Fprintf(cmd.Writer, "Task %s started.\n", taskID)
			return nil
		},
	}
}

// updateCurrentStateAfterTaskStart updates the epic's current_state when a task is started
func updateCurrentStateAfterTaskStart(epicData *epic.Epic, taskID string) {
	// Ensure current_state exists
	if epicData.CurrentState == nil {
		epicData.CurrentState = &epic.CurrentState{}
	}

	// Find the task to get its phase
	var taskPhaseID string
	var taskName string
	for _, task := range epicData.Tasks {
		if task.ID == taskID {
			taskPhaseID = task.PhaseID
			taskName = task.Name
			break
		}
	}

	// Update current state
	epicData.CurrentState.ActivePhase = taskPhaseID
	epicData.CurrentState.ActiveTask = taskID
	epicData.CurrentState.NextAction = fmt.Sprintf("Continue work on: %s", taskName)
}
