package cmd

import (
	"context"
	"fmt"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/urfave/cli/v3"
)

// StatusCommand returns the status command for displaying epic overview
func StatusCommand() *cli.Command {
	return &cli.Command{
		Name:    "status",
		Usage:   "Display epic status and progress overview",
		Aliases: []string{"s"},
		Action:  statusAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Override epic file from config",
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"F"},
				Usage:   "Output format: text (default), json, xml",
				Value:   "text",
			},
		},
	}
}

func statusAction(ctx context.Context, c *cli.Command) error {
	// Load configuration
	configPath := c.String("config")
	if configPath == "" {
		configPath = "./.agentpm.json"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine epic file (command flag overrides config)
	epicFile := c.String("file")
	if epicFile == "" {
		epicFile = cfg.CurrentEpic
	}
	if epicFile == "" {
		return fmt.Errorf("no epic file specified. Use --file flag or run 'agentpm init' first")
	}

	// Create storage and query service
	storage := storage.NewFileStorage()
	queryService := query.NewQueryService(storage)

	// Load epic
	err = queryService.LoadEpic(epicFile)
	if err != nil {
		return fmt.Errorf("failed to load epic: %w", err)
	}

	// Get epic status
	status, err := queryService.GetEpicStatus()
	if err != nil {
		return fmt.Errorf("failed to get epic status: %w", err)
	}

	// Output based on format
	outputFormat := c.String("format")
	switch outputFormat {
	case "xml":
		return outputStatusXML(c, status)
	case "json":
		return outputStatusJSON(c, status)
	default:
		return outputStatusText(c, status)
	}
}

func outputStatusText(c *cli.Command, status *query.EpicStatus) error {
	fmt.Fprintf(c.Root().Writer, "Epic Status: %s\n", status.Name)
	fmt.Fprintf(c.Root().Writer, "ID: %s\n", status.ID)
	fmt.Fprintf(c.Root().Writer, "Status: %s\n", status.Status)
	fmt.Fprintf(c.Root().Writer, "Progress: %d%% complete\n", status.CompletionPercentage)
	fmt.Fprintf(c.Root().Writer, "\nPhases: %d/%d completed\n", status.CompletedPhases, status.TotalPhases)
	fmt.Fprintf(c.Root().Writer, "Tests: %d passing, %d failing\n", status.PassingTests, status.FailingTests)

	if status.CurrentPhase != "" {
		fmt.Fprintf(c.Root().Writer, "\nCurrent Phase: %s\n", status.CurrentPhase)
	}
	if status.CurrentTask != "" {
		fmt.Fprintf(c.Root().Writer, "Current Task: %s\n", status.CurrentTask)
	}

	// Epic 13 Enhanced Status Information
	fmt.Fprintf(c.Root().Writer, "\n--- Epic 13 Status Overview ---\n")
	fmt.Fprintf(c.Root().Writer, "Epic Status (Epic 13): %s\n", status.Epic13Status.UnifiedStatuses.EpicStatus)
	fmt.Fprintf(c.Root().Writer, "Can Complete: %t\n", status.Epic13Status.CanComplete)

	if status.Epic13Status.BlockingItems > 0 {
		fmt.Fprintf(c.Root().Writer, "Blocking Items: %d\n", status.Epic13Status.BlockingItems)
	}

	// Unified status breakdown
	fmt.Fprintf(c.Root().Writer, "\nUnified Status Breakdown:\n")
	fmt.Fprintf(c.Root().Writer, "  Phases: %d WIP, %d Done\n",
		status.Epic13Status.UnifiedStatuses.PhasesWIP,
		status.Epic13Status.UnifiedStatuses.PhasesDone)
	fmt.Fprintf(c.Root().Writer, "  Tasks:  %d WIP, %d Done\n",
		status.Epic13Status.UnifiedStatuses.TasksWIP,
		status.Epic13Status.UnifiedStatuses.TasksDone)
	fmt.Fprintf(c.Root().Writer, "  Tests:  %d WIP, %d Done\n",
		status.Epic13Status.UnifiedStatuses.TestsWIP,
		status.Epic13Status.UnifiedStatuses.TestsDone)

	// Next actions
	if len(status.Epic13Status.NextActions) > 0 {
		fmt.Fprintf(c.Root().Writer, "\nNext Actions:\n")
		for i, action := range status.Epic13Status.NextActions {
			if i >= 3 { // Limit to top 3 actions for readability
				break
			}
			fmt.Fprintf(c.Root().Writer, "  - %s\n", action)
		}
	}

	// Validation errors (if any)
	if len(status.Epic13Status.ValidationErrors) > 0 {
		fmt.Fprintf(c.Root().Writer, "\nValidation Issues:\n")
		for i, err := range status.Epic13Status.ValidationErrors {
			if i >= 5 { // Limit to top 5 errors for readability
				fmt.Fprintf(c.Root().Writer, "  ... and %d more issues\n",
					len(status.Epic13Status.ValidationErrors)-5)
				break
			}
			fmt.Fprintf(c.Root().Writer, "  - %s\n", err)
		}
	}

	return nil
}

func outputStatusJSON(c *cli.Command, status *query.EpicStatus) error {
	// Build validation errors array
	validationErrors := "[]"
	if len(status.Epic13Status.ValidationErrors) > 0 {
		validationErrors = `["`
		for i, err := range status.Epic13Status.ValidationErrors {
			if i > 0 {
				validationErrors += `", "`
			}
			validationErrors += err
		}
		validationErrors += `"]`
	}

	// Build next actions array
	nextActions := "[]"
	if len(status.Epic13Status.NextActions) > 0 {
		nextActions = `["`
		for i, action := range status.Epic13Status.NextActions {
			if i > 0 {
				nextActions += `", "`
			}
			nextActions += action
		}
		nextActions += `"]`
	}

	jsonOutput := fmt.Sprintf(`{
  "epic": "%s",
  "name": "%s",
  "status": "%s",
  "progress": {
    "completion_percentage": %d,
    "completed_phases": %d,
    "total_phases": %d,
    "passing_tests": %d,
    "failing_tests": %d
  },
  "current_phase": "%s",
  "current_task": "%s",
  "epic13_status": {
    "can_complete": %t,
    "blocking_items": %d,
    "unified_statuses": {
      "epic_status": "%s",
      "phases_wip": %d,
      "phases_done": %d,
      "tasks_wip": %d,
      "tasks_done": %d,
      "tests_wip": %d,
      "tests_done": %d
    },
    "validation_errors": %s,
    "next_actions": %s
  }
}`,
		status.ID,
		status.Name,
		status.Status,
		status.CompletionPercentage,
		status.CompletedPhases,
		status.TotalPhases,
		status.PassingTests,
		status.FailingTests,
		status.CurrentPhase,
		status.CurrentTask,
		status.Epic13Status.CanComplete,
		status.Epic13Status.BlockingItems,
		status.Epic13Status.UnifiedStatuses.EpicStatus,
		status.Epic13Status.UnifiedStatuses.PhasesWIP,
		status.Epic13Status.UnifiedStatuses.PhasesDone,
		status.Epic13Status.UnifiedStatuses.TasksWIP,
		status.Epic13Status.UnifiedStatuses.TasksDone,
		status.Epic13Status.UnifiedStatuses.TestsWIP,
		status.Epic13Status.UnifiedStatuses.TestsDone,
		validationErrors,
		nextActions,
	)

	fmt.Fprintf(c.Root().Writer, "%s\n", jsonOutput)
	return nil
}

func outputStatusXML(c *cli.Command, status *query.EpicStatus) error {
	// Build validation errors XML
	validationErrorsXML := ""
	for _, err := range status.Epic13Status.ValidationErrors {
		validationErrorsXML += fmt.Sprintf("        <error>%s</error>\n", err)
	}

	// Build next actions XML
	nextActionsXML := ""
	for _, action := range status.Epic13Status.NextActions {
		nextActionsXML += fmt.Sprintf("        <action>%s</action>\n", action)
	}

	xmlOutput := fmt.Sprintf(`<status epic="%s">
    <name>%s</name>
    <status>%s</status>
    <progress>
        <completed_phases>%d</completed_phases>
        <total_phases>%d</total_phases>
        <passing_tests>%d</passing_tests>
        <failing_tests>%d</failing_tests>
        <completion_percentage>%d</completion_percentage>
    </progress>
    <current_phase>%s</current_phase>
    <current_task>%s</current_task>
    <epic13_status>
        <can_complete>%t</can_complete>
        <blocking_items>%d</blocking_items>
        <unified_statuses>
            <epic_status>%s</epic_status>
            <phases_wip>%d</phases_wip>
            <phases_done>%d</phases_done>
            <tasks_wip>%d</tasks_wip>
            <tasks_done>%d</tasks_done>
            <tests_wip>%d</tests_wip>
            <tests_done>%d</tests_done>
        </unified_statuses>
        <validation_errors>
%s        </validation_errors>
        <next_actions>
%s        </next_actions>
    </epic13_status>
</status>`,
		status.ID,
		status.Name,
		status.Status,
		status.CompletedPhases,
		status.TotalPhases,
		status.PassingTests,
		status.FailingTests,
		status.CompletionPercentage,
		status.CurrentPhase,
		status.CurrentTask,
		status.Epic13Status.CanComplete,
		status.Epic13Status.BlockingItems,
		status.Epic13Status.UnifiedStatuses.EpicStatus,
		status.Epic13Status.UnifiedStatuses.PhasesWIP,
		status.Epic13Status.UnifiedStatuses.PhasesDone,
		status.Epic13Status.UnifiedStatuses.TasksWIP,
		status.Epic13Status.UnifiedStatuses.TasksDone,
		status.Epic13Status.UnifiedStatuses.TestsWIP,
		status.Epic13Status.UnifiedStatuses.TestsDone,
		validationErrorsXML,
		nextActionsXML,
	)

	fmt.Fprintf(c.Root().Writer, "%s\n", xmlOutput)
	return nil
}
