package cmd

import (
	"context"
	"fmt"

	"github.com/memomoo/agentpm/internal/config"
	"github.com/memomoo/agentpm/internal/query"
	"github.com/memomoo/agentpm/internal/storage"
	"github.com/urfave/cli/v3"
)

// PendingCommand returns the pending command for displaying pending work
func PendingCommand() *cli.Command {
	return &cli.Command{
		Name:    "pending",
		Usage:   "Display pending work across all phases",
		Aliases: []string{"pend"},
		Action:  pendingAction,
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

func pendingAction(ctx context.Context, c *cli.Command) error {
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

	// Get pending work
	pending, err := queryService.GetPendingWork()
	if err != nil {
		return fmt.Errorf("failed to get pending work: %w", err)
	}

	// Output based on format
	outputFormat := c.String("format")
	switch outputFormat {
	case "xml":
		return outputPendingXML(c, pending)
	case "json":
		return outputPendingJSON(c, pending)
	default:
		return outputPendingText(c, pending)
	}
}

func outputPendingText(c *cli.Command, pending *query.PendingWork) error {
	fmt.Fprintf(c.Root().Writer, "Pending Work Overview\n\n")

	// Phases
	fmt.Fprintf(c.Root().Writer, "Phases (%d):\n", len(pending.Phases))
	if len(pending.Phases) == 0 {
		fmt.Fprintf(c.Root().Writer, "  (none)\n")
	} else {
		for _, phase := range pending.Phases {
			fmt.Fprintf(c.Root().Writer, "  %s - %s [%s]\n", phase.ID, phase.Name, phase.Status)
		}
	}

	// Tasks
	fmt.Fprintf(c.Root().Writer, "\nTasks (%d):\n", len(pending.Tasks))
	if len(pending.Tasks) == 0 {
		fmt.Fprintf(c.Root().Writer, "  (none)\n")
	} else {
		for _, task := range pending.Tasks {
			fmt.Fprintf(c.Root().Writer, "  %s (%s) - %s [%s]\n", task.ID, task.PhaseID, task.Name, task.Status)
		}
	}

	// Tests
	fmt.Fprintf(c.Root().Writer, "\nTests (%d):\n", len(pending.Tests))
	if len(pending.Tests) == 0 {
		fmt.Fprintf(c.Root().Writer, "  (none)\n")
	} else {
		for _, test := range pending.Tests {
			fmt.Fprintf(c.Root().Writer, "  %s (%s/%s) - %s [%s]\n", test.ID, test.PhaseID, test.TaskID, test.Name, test.Status)
		}
	}

	return nil
}

func outputPendingJSON(c *cli.Command, pending *query.PendingWork) error {
	fmt.Fprintf(c.Root().Writer, "{\n")
	fmt.Fprintf(c.Root().Writer, "  \"phases\": [\n")

	for i, phase := range pending.Phases {
		comma := ""
		if i < len(pending.Phases)-1 {
			comma = ","
		}
		fmt.Fprintf(c.Root().Writer, "    {\"id\": \"%s\", \"name\": \"%s\", \"status\": \"%s\"}%s\n",
			phase.ID, phase.Name, phase.Status, comma)
	}

	fmt.Fprintf(c.Root().Writer, "  ],\n")
	fmt.Fprintf(c.Root().Writer, "  \"tasks\": [\n")

	for i, task := range pending.Tasks {
		comma := ""
		if i < len(pending.Tasks)-1 {
			comma = ","
		}
		fmt.Fprintf(c.Root().Writer, "    {\"id\": \"%s\", \"phase_id\": \"%s\", \"name\": \"%s\", \"status\": \"%s\"}%s\n",
			task.ID, task.PhaseID, task.Name, task.Status, comma)
	}

	fmt.Fprintf(c.Root().Writer, "  ],\n")
	fmt.Fprintf(c.Root().Writer, "  \"tests\": [\n")

	for i, test := range pending.Tests {
		comma := ""
		if i < len(pending.Tests)-1 {
			comma = ","
		}
		fmt.Fprintf(c.Root().Writer, "    {\"id\": \"%s\", \"task_id\": \"%s\", \"phase_id\": \"%s\", \"name\": \"%s\", \"status\": \"%s\"}%s\n",
			test.ID, test.TaskID, test.PhaseID, test.Name, test.Status, comma)
	}

	fmt.Fprintf(c.Root().Writer, "  ]\n")
	fmt.Fprintf(c.Root().Writer, "}\n")
	return nil
}

func outputPendingXML(c *cli.Command, pending *query.PendingWork) error {
	fmt.Fprintf(c.Root().Writer, "<pending_work>\n")

	fmt.Fprintf(c.Root().Writer, "    <phases>\n")
	for _, phase := range pending.Phases {
		fmt.Fprintf(c.Root().Writer, "        <phase id=\"%s\" name=\"%s\" status=\"%s\"/>\n",
			phase.ID, phase.Name, phase.Status)
	}
	fmt.Fprintf(c.Root().Writer, "    </phases>\n")

	fmt.Fprintf(c.Root().Writer, "    <tasks>\n")
	for _, task := range pending.Tasks {
		fmt.Fprintf(c.Root().Writer, "        <task id=\"%s\" phase_id=\"%s\" status=\"%s\">%s</task>\n",
			task.ID, task.PhaseID, task.Status, task.Name)
	}
	fmt.Fprintf(c.Root().Writer, "    </tasks>\n")

	fmt.Fprintf(c.Root().Writer, "    <tests>\n")
	for _, test := range pending.Tests {
		fmt.Fprintf(c.Root().Writer, "        <test id=\"%s\" task_id=\"%s\" phase_id=\"%s\" status=\"%s\">%s</test>\n",
			test.ID, test.TaskID, test.PhaseID, test.Status, test.Name)
	}
	fmt.Fprintf(c.Root().Writer, "    </tests>\n")

	fmt.Fprintf(c.Root().Writer, "</pending_work>\n")
	return nil
}
