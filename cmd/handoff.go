package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/memomoo/agentpm/internal/config"
	"github.com/memomoo/agentpm/internal/reports"
	"github.com/memomoo/agentpm/internal/storage"
	"github.com/urfave/cli/v3"
)

// HandoffCommand returns the handoff command for generating agent handoff reports
func HandoffCommand() *cli.Command {
	return &cli.Command{
		Name:   "handoff",
		Usage:  "Generate comprehensive handoff report for agent transitions",
		Action: handoffAction,
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
				Value:   "xml",
			},
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Usage:   "Number of recent events to include",
				Value:   5,
			},
		},
	}
}

func handoffAction(ctx context.Context, c *cli.Command) error {
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

	// Create storage and reports service
	storage := storage.NewFileStorage()
	reportsService := reports.NewReportService(storage)

	// Load epic
	err = reportsService.LoadEpic(epicFile)
	if err != nil {
		return fmt.Errorf("failed to load epic: %w", err)
	}

	// Generate handoff report
	limit := c.Int("limit")
	report, err := reportsService.GenerateHandoffReport(limit)
	if err != nil {
		return fmt.Errorf("failed to generate handoff report: %w", err)
	}

	// Output based on format
	outputFormat := c.String("format")
	switch outputFormat {
	case "json":
		return outputHandoffJSON(c, report)
	case "text":
		return outputHandoffText(c, report)
	default:
		return outputHandoffXML(c, report)
	}
}

func outputHandoffText(c *cli.Command, report *reports.HandoffReport) error {
	fmt.Fprintf(c.Root().Writer, "=== AGENT HANDOFF REPORT ===\n\n")

	// Epic Info
	fmt.Fprintf(c.Root().Writer, "Epic: %s\n", report.EpicInfo.Name)
	fmt.Fprintf(c.Root().Writer, "ID: %s\n", report.EpicInfo.ID)
	fmt.Fprintf(c.Root().Writer, "Status: %s\n", report.EpicInfo.Status)
	fmt.Fprintf(c.Root().Writer, "Assignee: %s\n", report.EpicInfo.Assignee)
	fmt.Fprintf(c.Root().Writer, "Started: %s\n\n", report.EpicInfo.Started.Format("2006-01-02 15:04:05"))

	// Current State
	fmt.Fprintf(c.Root().Writer, "CURRENT STATE:\n")
	if report.CurrentState.ActivePhase != "" {
		fmt.Fprintf(c.Root().Writer, "  Active Phase: %s\n", report.CurrentState.ActivePhase)
	}
	if report.CurrentState.ActiveTask != "" {
		fmt.Fprintf(c.Root().Writer, "  Active Task: %s\n", report.CurrentState.ActiveTask)
	}
	fmt.Fprintf(c.Root().Writer, "  Next Action: %s\n\n", report.CurrentState.NextAction)

	// Progress Summary
	fmt.Fprintf(c.Root().Writer, "PROGRESS SUMMARY:\n")
	fmt.Fprintf(c.Root().Writer, "  Completion: %d%%\n", report.Summary.CompletionPercentage)
	fmt.Fprintf(c.Root().Writer, "  Phases: %d/%d completed\n", report.Summary.CompletedPhases, report.Summary.TotalPhases)
	fmt.Fprintf(c.Root().Writer, "  Tests: %d passing, %d failing\n\n", report.Summary.PassingTests, report.Summary.FailingTests)

	// Blockers
	if len(report.Blockers) > 0 {
		fmt.Fprintf(c.Root().Writer, "BLOCKERS:\n")
		for _, blocker := range report.Blockers {
			fmt.Fprintf(c.Root().Writer, "  - %s\n", blocker)
		}
		fmt.Fprintf(c.Root().Writer, "\n")
	}

	// Recent Events
	if len(report.RecentEvents) > 0 {
		fmt.Fprintf(c.Root().Writer, "RECENT EVENTS:\n")
		for _, event := range report.RecentEvents {
			fmt.Fprintf(c.Root().Writer, "  [%s] %s: %s\n",
				event.Timestamp.Format("2006-01-02 15:04:05"),
				event.Type,
				event.Data)
		}
	}

	fmt.Fprintf(c.Root().Writer, "\nGenerated at: %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))
	return nil
}

func outputHandoffJSON(c *cli.Command, report *reports.HandoffReport) error {
	// For simplicity, we'll create a JSON-like output manually
	// In a real implementation, you'd use json.Marshal
	jsonOutput := fmt.Sprintf(`{
  "epic": "%s",
  "generated_at": "%s",
  "epic_info": {
    "id": "%s",
    "name": "%s",
    "status": "%s",
    "started": "%s",
    "assignee": "%s"
  },
  "current_state": {
    "active_phase": "%s",
    "active_task": "%s",
    "next_action": "%s"
  },
  "summary": {
    "completed_phases": %d,
    "total_phases": %d,
    "passing_tests": %d,
    "failing_tests": %d,
    "completion_percentage": %d
  },
  "recent_events": [`,
		report.EpicInfo.ID,
		report.GeneratedAt.Format(time.RFC3339),
		report.EpicInfo.ID,
		report.EpicInfo.Name,
		report.EpicInfo.Status,
		report.EpicInfo.Started.Format(time.RFC3339),
		report.EpicInfo.Assignee,
		report.CurrentState.ActivePhase,
		report.CurrentState.ActiveTask,
		report.CurrentState.NextAction,
		report.Summary.CompletedPhases,
		report.Summary.TotalPhases,
		report.Summary.PassingTests,
		report.Summary.FailingTests,
		report.Summary.CompletionPercentage,
	)

	// Add events
	for i, event := range report.RecentEvents {
		if i > 0 {
			jsonOutput += ","
		}
		jsonOutput += fmt.Sprintf(`
    {
      "timestamp": "%s",
      "type": "%s",
      "data": "%s"
    }`, event.Timestamp.Format(time.RFC3339), event.Type, event.Data)
	}

	jsonOutput += `
  ],
  "blockers": [`

	// Add blockers
	for i, blocker := range report.Blockers {
		if i > 0 {
			jsonOutput += ","
		}
		jsonOutput += fmt.Sprintf(`"%s"`, blocker)
	}

	jsonOutput += `
  ]
}`

	fmt.Fprintf(c.Root().Writer, "%s\n", jsonOutput)
	return nil
}

func outputHandoffXML(c *cli.Command, report *reports.HandoffReport) error {
	// Write XML header and data
	fmt.Fprintf(c.Root().Writer, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	fmt.Fprintf(c.Root().Writer, "<handoff epic=\"%s\" timestamp=\"%s\">\n",
		report.EpicInfo.ID,
		report.GeneratedAt.Format(time.RFC3339))

	// Write the structured XML content manually to match expected output format
	fmt.Fprintf(c.Root().Writer, "    <epic_info>\n")
	fmt.Fprintf(c.Root().Writer, "        <name>%s</name>\n", report.EpicInfo.Name)
	fmt.Fprintf(c.Root().Writer, "        <status>%s</status>\n", report.EpicInfo.Status)
	fmt.Fprintf(c.Root().Writer, "        <started>%s</started>\n", report.EpicInfo.Started.Format(time.RFC3339))
	fmt.Fprintf(c.Root().Writer, "        <assignee>%s</assignee>\n", report.EpicInfo.Assignee)
	fmt.Fprintf(c.Root().Writer, "    </epic_info>\n")

	fmt.Fprintf(c.Root().Writer, "    <current_state>\n")
	fmt.Fprintf(c.Root().Writer, "        <active_phase>%s</active_phase>\n", report.CurrentState.ActivePhase)
	fmt.Fprintf(c.Root().Writer, "        <active_task>%s</active_task>\n", report.CurrentState.ActiveTask)
	fmt.Fprintf(c.Root().Writer, "        <next_action>%s</next_action>\n", report.CurrentState.NextAction)
	fmt.Fprintf(c.Root().Writer, "    </current_state>\n")

	fmt.Fprintf(c.Root().Writer, "    <summary>\n")
	fmt.Fprintf(c.Root().Writer, "        <completed_phases>%d</completed_phases>\n", report.Summary.CompletedPhases)
	fmt.Fprintf(c.Root().Writer, "        <total_phases>%d</total_phases>\n", report.Summary.TotalPhases)
	fmt.Fprintf(c.Root().Writer, "        <passing_tests>%d</passing_tests>\n", report.Summary.PassingTests)
	fmt.Fprintf(c.Root().Writer, "        <failing_tests>%d</failing_tests>\n", report.Summary.FailingTests)
	fmt.Fprintf(c.Root().Writer, "        <completion_percentage>%d</completion_percentage>\n", report.Summary.CompletionPercentage)
	fmt.Fprintf(c.Root().Writer, "    </summary>\n")

	if len(report.RecentEvents) > 0 {
		fmt.Fprintf(c.Root().Writer, "    <recent_events limit=\"%d\">\n", len(report.RecentEvents))
		for _, event := range report.RecentEvents {
			fmt.Fprintf(c.Root().Writer, "        <event timestamp=\"%s\" type=\"%s\">\n",
				event.Timestamp.Format(time.RFC3339), event.Type)
			fmt.Fprintf(c.Root().Writer, "            %s\n", event.Data)
			fmt.Fprintf(c.Root().Writer, "        </event>\n")
		}
		fmt.Fprintf(c.Root().Writer, "    </recent_events>\n")
	}

	if len(report.Blockers) > 0 {
		fmt.Fprintf(c.Root().Writer, "    <blockers>\n")
		for _, blocker := range report.Blockers {
			fmt.Fprintf(c.Root().Writer, "        <blocker>%s</blocker>\n", blocker)
		}
		fmt.Fprintf(c.Root().Writer, "    </blockers>\n")
	}

	fmt.Fprintf(c.Root().Writer, "</handoff>\n")
	return nil
}
