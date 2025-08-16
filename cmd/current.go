package cmd

import (
	"context"
	"fmt"

	"github.com/memomoo/agentpm/internal/config"
	"github.com/memomoo/agentpm/internal/query"
	"github.com/memomoo/agentpm/internal/storage"
	"github.com/urfave/cli/v3"
)

// CurrentCommand returns the current command for displaying active work state
func CurrentCommand() *cli.Command {
	return &cli.Command{
		Name:    "current",
		Usage:   "Display current active work state",
		Aliases: []string{"cur"},
		Action:  currentAction,
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

func currentAction(ctx context.Context, c *cli.Command) error {
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

	// Get current state
	state, err := queryService.GetCurrentState()
	if err != nil {
		return fmt.Errorf("failed to get current state: %w", err)
	}

	// Output based on format
	outputFormat := c.String("format")
	switch outputFormat {
	case "xml":
		return outputCurrentXML(c, state)
	case "json":
		return outputCurrentJSON(c, state)
	default:
		return outputCurrentText(c, state)
	}
}

func outputCurrentText(c *cli.Command, state *query.CurrentState) error {
	fmt.Fprintf(c.Root().Writer, "Current Work State\n")
	fmt.Fprintf(c.Root().Writer, "Epic Status: %s\n", state.EpicStatus)

	if state.ActivePhase != "" {
		fmt.Fprintf(c.Root().Writer, "Active Phase: %s\n", state.ActivePhase)
	} else {
		fmt.Fprintf(c.Root().Writer, "Active Phase: none\n")
	}

	if state.ActiveTask != "" {
		fmt.Fprintf(c.Root().Writer, "Active Task: %s\n", state.ActiveTask)
	} else {
		fmt.Fprintf(c.Root().Writer, "Active Task: none\n")
	}

	fmt.Fprintf(c.Root().Writer, "Failing Tests: %d\n", state.FailingTests)
	fmt.Fprintf(c.Root().Writer, "\nNext Action: %s\n", state.NextAction)

	return nil
}

func outputCurrentJSON(c *cli.Command, state *query.CurrentState) error {
	jsonOutput := fmt.Sprintf(`{
  "epic_status": "%s",
  "active_phase": "%s",
  "active_task": "%s",
  "failing_tests": %d,
  "next_action": "%s"
}`,
		state.EpicStatus,
		state.ActivePhase,
		state.ActiveTask,
		state.FailingTests,
		state.NextAction,
	)

	fmt.Fprintf(c.Root().Writer, "%s\n", jsonOutput)
	return nil
}

func outputCurrentXML(c *cli.Command, state *query.CurrentState) error {
	xmlOutput := fmt.Sprintf(`<current_state>
    <epic_status>%s</epic_status>
    <active_phase>%s</active_phase>
    <active_task>%s</active_task>
    <next_action>%s</next_action>
    <failing_tests>%d</failing_tests>
</current_state>`,
		state.EpicStatus,
		state.ActivePhase,
		state.ActiveTask,
		state.NextAction,
		state.FailingTests,
	)

	fmt.Fprintf(c.Root().Writer, "%s\n", xmlOutput)
	return nil
}
