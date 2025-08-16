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
		Aliases: []string{"st"},
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

	return nil
}

func outputStatusJSON(c *cli.Command, status *query.EpicStatus) error {
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
  "current_task": "%s"
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
	)

	fmt.Fprintf(c.Root().Writer, "%s\n", jsonOutput)
	return nil
}

func outputStatusXML(c *cli.Command, status *query.EpicStatus) error {
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
	)

	fmt.Fprintf(c.Root().Writer, "%s\n", xmlOutput)
	return nil
}
