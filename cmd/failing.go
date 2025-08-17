package cmd

import (
	"context"
	"fmt"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/urfave/cli/v3"
)

// FailingCommand returns the failing command for displaying failing tests
func FailingCommand() *cli.Command {
	return &cli.Command{
		Name:    "failing",
		Usage:   "Display failing tests with details",
		Aliases: []string{"f"},
		Action:  failingAction,
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

func failingAction(ctx context.Context, c *cli.Command) error {
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

	// Get failing tests
	failing, err := queryService.GetFailingTests()
	if err != nil {
		return fmt.Errorf("failed to get failing tests: %w", err)
	}

	// Output based on format
	outputFormat := c.String("format")
	switch outputFormat {
	case "xml":
		return outputFailingXML(c, failing)
	case "json":
		return outputFailingJSON(c, failing)
	default:
		return outputFailingText(c, failing)
	}
}

func outputFailingText(c *cli.Command, failing []query.FailingTest) error {
	fmt.Fprintf(c.Root().Writer, "Failing Tests Report\n\n")

	if len(failing) == 0 {
		fmt.Fprintf(c.Root().Writer, "✓ All tests are passing!\n")
		return nil
	}

	fmt.Fprintf(c.Root().Writer, "Found %d failing test(s):\n\n", len(failing))

	// Group by phase for better organization
	phaseGroups := make(map[string][]query.FailingTest)
	for _, test := range failing {
		phaseGroups[test.PhaseID] = append(phaseGroups[test.PhaseID], test)
	}

	for phaseID, tests := range phaseGroups {
		if phaseID != "" {
			fmt.Fprintf(c.Root().Writer, "Phase %s:\n", phaseID)
		} else {
			fmt.Fprintf(c.Root().Writer, "Unknown Phase:\n")
		}

		for _, test := range tests {
			fmt.Fprintf(c.Root().Writer, "  ✗ %s (%s)\n", test.ID, test.TaskID)
			fmt.Fprintf(c.Root().Writer, "    %s\n", test.Name)

			if test.Description != "" {
				fmt.Fprintf(c.Root().Writer, "    Description: %s\n", test.Description)
			}

			if test.FailureNote != "" {
				fmt.Fprintf(c.Root().Writer, "    Failure: %s\n", test.FailureNote)
			}

			fmt.Fprintf(c.Root().Writer, "\n")
		}
	}

	return nil
}

func outputFailingJSON(c *cli.Command, failing []query.FailingTest) error {
	fmt.Fprintf(c.Root().Writer, "{\n")
	fmt.Fprintf(c.Root().Writer, "  \"failing_tests\": [\n")

	for i, test := range failing {
		comma := ""
		if i < len(failing)-1 {
			comma = ","
		}

		fmt.Fprintf(c.Root().Writer, "    {\n")
		fmt.Fprintf(c.Root().Writer, "      \"id\": \"%s\",\n", test.ID)
		fmt.Fprintf(c.Root().Writer, "      \"phase_id\": \"%s\",\n", test.PhaseID)
		fmt.Fprintf(c.Root().Writer, "      \"task_id\": \"%s\",\n", test.TaskID)
		fmt.Fprintf(c.Root().Writer, "      \"name\": \"%s\",\n", test.Name)
		fmt.Fprintf(c.Root().Writer, "      \"description\": \"%s\",\n", test.Description)
		fmt.Fprintf(c.Root().Writer, "      \"failure_note\": \"%s\"\n", test.FailureNote)
		fmt.Fprintf(c.Root().Writer, "    }%s\n", comma)
	}

	fmt.Fprintf(c.Root().Writer, "  ],\n")
	fmt.Fprintf(c.Root().Writer, "  \"total_failing\": %d\n", len(failing))
	fmt.Fprintf(c.Root().Writer, "}\n")
	return nil
}

func outputFailingXML(c *cli.Command, failing []query.FailingTest) error {
	fmt.Fprintf(c.Root().Writer, "<failing_tests>\n")

	for _, test := range failing {
		fmt.Fprintf(c.Root().Writer, "    <test id=\"%s\" phase_id=\"%s\" task_id=\"%s\">\n",
			test.ID, test.PhaseID, test.TaskID)
		fmt.Fprintf(c.Root().Writer, "        <name>%s</name>\n", test.Name)

		if test.Description != "" {
			fmt.Fprintf(c.Root().Writer, "        <description>%s</description>\n", test.Description)
		}

		if test.FailureNote != "" {
			fmt.Fprintf(c.Root().Writer, "        <failure_note>%s</failure_note>\n", test.FailureNote)
		}

		fmt.Fprintf(c.Root().Writer, "    </test>\n")
	}

	fmt.Fprintf(c.Root().Writer, "</failing_tests>\n")
	return nil
}
