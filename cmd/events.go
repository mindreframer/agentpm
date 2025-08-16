package cmd

import (
	"context"
	"fmt"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/urfave/cli/v3"
)

// EventsCommand returns the events command for displaying recent events timeline
func EventsCommand() *cli.Command {
	return &cli.Command{
		Name:    "events",
		Usage:   "Display recent events timeline",
		Aliases: []string{"evt"},
		Action:  eventsAction,
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
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Usage:   "Maximum number of events to show (default: 10, max: 100)",
				Value:   10,
			},
		},
	}
}

func eventsAction(ctx context.Context, c *cli.Command) error {
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

	// Get limit
	limit := c.Int("limit")
	if limit <= 0 {
		limit = 10
	}

	// Create storage and query service
	storage := storage.NewFileStorage()
	queryService := query.NewQueryService(storage)

	// Load epic
	err = queryService.LoadEpic(epicFile)
	if err != nil {
		return fmt.Errorf("failed to load epic: %w", err)
	}

	// Get recent events
	events, err := queryService.GetRecentEvents(limit)
	if err != nil {
		return fmt.Errorf("failed to get recent events: %w", err)
	}

	// Output based on format
	outputFormat := c.String("format")
	switch outputFormat {
	case "xml":
		return outputEventsXML(c, events, limit)
	case "json":
		return outputEventsJSON(c, events, limit)
	default:
		return outputEventsText(c, events, limit)
	}
}

func outputEventsText(c *cli.Command, events []query.Event, limit int) error {
	fmt.Fprintf(c.Root().Writer, "Recent Events Timeline (limit: %d)\n\n", limit)

	if len(events) == 0 {
		fmt.Fprintf(c.Root().Writer, "No events found.\n")
		return nil
	}

	fmt.Fprintf(c.Root().Writer, "Showing %d event(s) (most recent first):\n\n", len(events))

	for i, event := range events {
		fmt.Fprintf(c.Root().Writer, "%d. [%s] %s\n",
			i+1,
			event.Timestamp.Format("2006-01-02 15:04:05"),
			event.Type)

		if event.Agent != "" {
			fmt.Fprintf(c.Root().Writer, "   Agent: %s\n", event.Agent)
		}

		if event.PhaseID != "" {
			fmt.Fprintf(c.Root().Writer, "   Phase: %s\n", event.PhaseID)
		}

		if event.Content != "" {
			// Format content with proper indentation
			fmt.Fprintf(c.Root().Writer, "   Content: %s\n", event.Content)
		}

		fmt.Fprintf(c.Root().Writer, "\n")
	}

	return nil
}

func outputEventsJSON(c *cli.Command, events []query.Event, limit int) error {
	fmt.Fprintf(c.Root().Writer, "{\n")
	fmt.Fprintf(c.Root().Writer, "  \"events\": [\n")

	for i, event := range events {
		comma := ""
		if i < len(events)-1 {
			comma = ","
		}

		fmt.Fprintf(c.Root().Writer, "    {\n")
		fmt.Fprintf(c.Root().Writer, "      \"timestamp\": \"%s\",\n", event.Timestamp.Format("2006-01-02T15:04:05Z"))
		fmt.Fprintf(c.Root().Writer, "      \"type\": \"%s\",\n", event.Type)
		fmt.Fprintf(c.Root().Writer, "      \"agent\": \"%s\",\n", event.Agent)
		fmt.Fprintf(c.Root().Writer, "      \"phase_id\": \"%s\",\n", event.PhaseID)
		fmt.Fprintf(c.Root().Writer, "      \"content\": \"%s\"\n", event.Content)
		fmt.Fprintf(c.Root().Writer, "    }%s\n", comma)
	}

	fmt.Fprintf(c.Root().Writer, "  ],\n")
	fmt.Fprintf(c.Root().Writer, "  \"limit\": %d,\n", limit)
	fmt.Fprintf(c.Root().Writer, "  \"total\": %d\n", len(events))
	fmt.Fprintf(c.Root().Writer, "}\n")
	return nil
}

func outputEventsXML(c *cli.Command, events []query.Event, limit int) error {
	fmt.Fprintf(c.Root().Writer, "<events limit=\"%d\" total=\"%d\">\n", limit, len(events))

	for _, event := range events {
		fmt.Fprintf(c.Root().Writer, "    <event timestamp=\"%s\" type=\"%s\"",
			event.Timestamp.Format("2006-01-02T15:04:05Z"), event.Type)

		if event.Agent != "" {
			fmt.Fprintf(c.Root().Writer, " agent=\"%s\"", event.Agent)
		}

		if event.PhaseID != "" {
			fmt.Fprintf(c.Root().Writer, " phase_id=\"%s\"", event.PhaseID)
		}

		fmt.Fprintf(c.Root().Writer, ">\n")

		if event.Content != "" {
			fmt.Fprintf(c.Root().Writer, "        <content>%s</content>\n", event.Content)
		}

		fmt.Fprintf(c.Root().Writer, "    </event>\n")
	}

	fmt.Fprintf(c.Root().Writer, "</events>\n")
	return nil
}
