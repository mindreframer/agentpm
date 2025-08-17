package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/mindreframer/agentpm/internal/autonext"
	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/phases"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/mindreframer/agentpm/internal/tasks"
	"github.com/urfave/cli/v3"
)

func StartNextCommand() *cli.Command {
	return &cli.Command{
		Name:    "next",
		Usage:   "Auto-start next available work",
		Aliases: []string{"start-next"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "file",
				Usage: "Epic file path (overrides config)",
			},
			&cli.StringFlag{
				Name:  "time",
				Usage: "Timestamp for the operation (ISO 8601 format)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
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
			taskService := tasks.NewTaskService(storageImpl, queryService)
			autoNextService := autonext.NewAutoNextService(storageImpl, queryService, phaseService, taskService)

			// Load epic
			epicData, err := storageImpl.LoadEpic(epicFile)
			if err != nil {
				return fmt.Errorf("failed to load epic: %w", err)
			}

			// Execute auto-next selection
			result, err := autoNextService.SelectNext(epicData, timestamp)
			if err != nil {
				return fmt.Errorf("failed to execute auto-next selection: %w", err)
			}

			// Save the updated epic (if changes were made)
			if result.Action != autonext.ActionNoWork && result.Action != autonext.ActionCompleteEpic {
				err = storageImpl.SaveEpic(epicData, epicFile)
				if err != nil {
					return fmt.Errorf("failed to save epic: %w", err)
				}
			}

			// Output result based on action type
			return outputAutoNextResult(cmd, result)
		},
	}
}

// outputAutoNextResult outputs the appropriate format based on the auto-next result
func outputAutoNextResult(cmd *cli.Command, result *autonext.AutoNextResult) error {
	switch result.Action {
	case autonext.ActionStartTask:
		// Output XML for task selection (complex decision making)
		fmt.Fprintf(cmd.Writer, "%s\n", formatTaskStartedXML(result))
		return nil

	case autonext.ActionStartPhase:
		// Output XML for phase activation (complex decision making)
		if result.XMLOutput != "" {
			fmt.Fprintf(cmd.Writer, "%s\n", result.XMLOutput)
		} else {
			fmt.Fprintf(cmd.Writer, "%s\n", formatPhaseOnlyStartedXML(result))
		}
		return nil

	case autonext.ActionCompleteEpic:
		// Output XML for completion
		fmt.Fprintf(cmd.Writer, "%s\n", formatAllCompleteXML(result))
		return nil

	case autonext.ActionNoWork:
		// Simple text output for no action needed
		fmt.Fprintf(cmd.Writer, "%s\n", result.Message)
		return nil

	default:
		return fmt.Errorf("unknown auto-next action: %s", result.Action)
	}
}

// formatTaskStartedXML creates XML output for task started in active phase
func formatTaskStartedXML(result *autonext.AutoNextResult) string {
	return fmt.Sprintf(`<task_started epic="epic-id" task="%s">
    <task_description>%s</task_description>
    <phase_id>%s</phase_id>
    <previous_status>pending</previous_status>
    <new_status>wip</new_status>
    <started_at>%s</started_at>
    <auto_selected>%t</auto_selected>
    <message>%s</message>
</task_started>`,
		result.TaskID, result.TaskName, result.PhaseID,
		result.StartedAt.Format(time.RFC3339), result.AutoSelected, result.Message)
}

// formatPhaseOnlyStartedXML creates XML output for phase started without tasks
func formatPhaseOnlyStartedXML(result *autonext.AutoNextResult) string {
	return fmt.Sprintf(`<phase_started epic="epic-id" phase="%s">
    <phase_name>%s</phase_name>
    <previous_status>pending</previous_status>
    <new_status>wip</new_status>
    <started_at>%s</started_at>
    <message>%s</message>
</phase_started>`,
		result.PhaseID, result.PhaseName,
		result.StartedAt.Format(time.RFC3339), result.Message)
}

// formatAllCompleteXML creates XML output for epic completion
func formatAllCompleteXML(result *autonext.AutoNextResult) string {
	return fmt.Sprintf(`<all_complete epic="epic-id">
    <message>%s</message>
    <suggestion>Use 'agentpm done-epic' to complete the epic</suggestion>
</all_complete>`, result.Message)
}
