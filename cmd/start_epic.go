package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beevik/etree"
	"github.com/memomoo/agentpm/internal/config"
	"github.com/memomoo/agentpm/internal/lifecycle"
	"github.com/memomoo/agentpm/internal/query"
	"github.com/memomoo/agentpm/internal/storage"
	"github.com/urfave/cli/v3"
)

func StartEpicCommand() *cli.Command {
	return &cli.Command{
		Name:    "start-epic",
		Aliases: []string{"start"},
		Usage:   "Start working on an epic (transition from pending to wip)",
		Description: `Start an epic by transitioning its status from "pending" to "wip".
		
This command:
- Changes epic status from "pending" to "wip"
- Sets the started_at timestamp
- Creates an automatic event log entry
- Validates that the epic is in a valid state to start

Examples:
  agentpm start-epic                    # Start epic from config
  agentpm start-epic --file epic-5.xml # Start specific epic
  agentpm start-epic --time 2025-08-16T15:30:00Z # Use specific timestamp`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Epic file to start (overrides config)",
			},
			&cli.StringFlag{
				Name:  "time",
				Usage: "Specific timestamp for epic start (ISO 8601 format, for deterministic testing)",
			},
		},
		Action: startEpicAction,
	}
}

func startEpicAction(ctx context.Context, c *cli.Command) error {
	// Get global flags
	configPath := c.String("config")
	if configPath == "" {
		configPath = "./.agentpm.json"
	}
	format := c.String("format")

	// Determine epic file to use
	epicFile := c.String("file")
	if epicFile == "" {
		// Load configuration only if no file specified
		cfg, configErr := config.LoadConfig(configPath)
		if configErr != nil {
			return fmt.Errorf("failed to load configuration: %w", configErr)
		}

		if cfg.CurrentEpic == "" {
			return fmt.Errorf("no epic file specified and no current epic in config")
		}
		epicFile = cfg.EpicFilePath()
	}

	// Initialize services
	storageFactory := storage.NewFactory(false) // Use file storage
	storageImpl := storageFactory.CreateStorage()
	queryService := query.NewQueryService(storageImpl)
	lifecycleService := lifecycle.NewLifecycleService(storageImpl, queryService)

	// Handle custom timestamp if provided
	var timestamp *time.Time
	if timeStr := c.String("time"); timeStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return fmt.Errorf("invalid time format: %w (use ISO 8601 format like 2025-08-16T15:30:00Z)", err)
		}
		timestamp = &parsedTime
	}

	// Create start epic request
	request := lifecycle.StartEpicRequest{
		EpicFile:  epicFile,
		Timestamp: timestamp,
	}

	// Start the epic
	result, err := lifecycleService.StartEpic(request)
	if err != nil {
		return outputStartEpicError(c, err, format)
	}

	// Output the result
	return outputStartEpicResult(c, result, format)
}

func outputStartEpicResult(c *cli.Command, result *lifecycle.StartEpicResult, format string) error {
	switch format {
	case "json":
		return outputStartEpicJSON(c, result)
	case "xml":
		return outputStartEpicXML(c, result)
	default:
		return outputStartEpicText(c, result)
	}
}

func outputStartEpicText(c *cli.Command, result *lifecycle.StartEpicResult) error {
	fmt.Fprintf(c.Root().Writer, "Epic %s started successfully\n", result.EpicID)
	fmt.Fprintf(c.Root().Writer, "Status: %s → %s\n", result.PreviousStatus, result.NewStatus)
	fmt.Fprintf(c.Root().Writer, "Started at: %s\n", result.StartedAt.Format(time.RFC3339))
	fmt.Fprintf(c.Root().Writer, "\n%s\n", result.Message)
	return nil
}

func outputStartEpicJSON(c *cli.Command, result *lifecycle.StartEpicResult) error {
	output := map[string]interface{}{
		"epic_started": map[string]interface{}{
			"epic_id":         result.EpicID,
			"previous_status": result.PreviousStatus.String(),
			"new_status":      result.NewStatus.String(),
			"started_at":      result.StartedAt.Format(time.RFC3339),
			"event_created":   result.EventCreated,
			"message":         result.Message,
		},
	}

	encoder := json.NewEncoder(c.Root().Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputStartEpicXML(c *cli.Command, result *lifecycle.StartEpicResult) error {
	doc := etree.NewDocument()
	root := doc.CreateElement("epic_started")
	root.SetText("\n    ")

	// Add epic attribute
	root.CreateAttr("epic", result.EpicID)

	// Add child elements
	prevStatus := root.CreateElement("previous_status")
	prevStatus.SetText(result.PreviousStatus.String())

	newStatus := root.CreateElement("new_status")
	newStatus.SetText(result.NewStatus.String())

	startedAt := root.CreateElement("started_at")
	startedAt.SetText(result.StartedAt.Format(time.RFC3339))

	eventCreated := root.CreateElement("event_created")
	if result.EventCreated {
		eventCreated.SetText("true")
	} else {
		eventCreated.SetText("false")
	}

	message := root.CreateElement("message")
	message.SetText(result.Message)

	doc.Indent(4)
	doc.WriteTo(c.Root().Writer)
	fmt.Fprintf(c.Root().Writer, "\n") // Add newline
	return nil
}

func outputStartEpicError(c *cli.Command, err error, format string) error {
	// Check if it's a transition error for special handling
	if transitionErr, ok := err.(*lifecycle.TransitionError); ok {
		return outputTransitionError(c, transitionErr, format)
	}

	// Generic error output
	fmt.Fprintf(c.Root().ErrWriter, "Error: %v\n", err)
	return err
}

func outputTransitionError(c *cli.Command, err *lifecycle.TransitionError, format string) error {
	switch format {
	case "json":
		return outputTransitionErrorJSON(c, err)
	case "xml":
		return outputTransitionErrorXML(c, err)
	default:
		return outputTransitionErrorText(c, err)
	}
}

func outputTransitionErrorText(c *cli.Command, err *lifecycle.TransitionError) error {
	fmt.Fprintf(c.Root().ErrWriter, "Error: %s\n", err.Message)
	fmt.Fprintf(c.Root().ErrWriter, "Epic: %s\n", err.EpicID)
	fmt.Fprintf(c.Root().ErrWriter, "Current status: %s\n", err.CurrentStatus)
	if err.Suggestion != "" {
		fmt.Fprintf(c.Root().ErrWriter, "Suggestion: %s\n", err.Suggestion)
	}
	return err
}

func outputTransitionErrorJSON(c *cli.Command, err *lifecycle.TransitionError) error {
	output := map[string]interface{}{
		"error": map[string]interface{}{
			"type":           "invalid_transition",
			"epic_id":        err.EpicID,
			"current_status": err.CurrentStatus.String(),
			"target_status":  err.TargetStatus.String(),
			"message":        err.Message,
			"suggestion":     err.Suggestion,
		},
	}

	encoder := json.NewEncoder(c.Root().ErrWriter)
	encoder.SetIndent("", "  ")
	encoder.Encode(output)
	return err
}

func outputTransitionErrorXML(c *cli.Command, err *lifecycle.TransitionError) error {
	doc := etree.NewDocument()
	root := doc.CreateElement("error")

	errorType := root.CreateElement("type")
	errorType.SetText("invalid_transition")

	epicID := root.CreateElement("epic_id")
	epicID.SetText(err.EpicID)

	currentStatus := root.CreateElement("current_status")
	currentStatus.SetText(err.CurrentStatus.String())

	targetStatus := root.CreateElement("target_status")
	targetStatus.SetText(err.TargetStatus.String())

	message := root.CreateElement("message")
	message.SetText(err.Message)

	if err.Suggestion != "" {
		suggestion := root.CreateElement("suggestion")
		suggestion.SetText(err.Suggestion)
	}

	doc.Indent(4)
	doc.WriteTo(c.Root().ErrWriter)
	fmt.Fprintf(c.Root().ErrWriter, "\n") // Add newline
	return err
}
