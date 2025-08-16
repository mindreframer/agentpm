package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beevik/etree"
	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/lifecycle"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/urfave/cli/v3"
)

func DoneEpicCommand() *cli.Command {
	return &cli.Command{
		Name:    "done-epic",
		Aliases: []string{"done", "complete"},
		Usage:   "Mark an epic as done (transition from wip to done)",
		Description: `Complete an epic by transitioning its status from "wip" to "done".

This command:
- Changes epic status from "wip" to "done"
- Sets the completed_at timestamp
- Creates an automatic event log entry
- Validates that the epic is in a valid state to complete
- Generates a completion summary

Examples:
  agentpm done-epic                     # Complete epic from config
  agentpm done-epic --file epic-5.xml  # Complete specific epic
  agentpm done-epic --time 2025-08-16T15:30:00Z # Use specific timestamp`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Epic file to complete (overrides config)",
			},
			&cli.StringFlag{
				Name:  "time",
				Usage: "Specific timestamp for epic completion (ISO 8601 format, for deterministic testing)",
			},
		},
		Action: doneEpicAction,
	}
}

func doneEpicAction(ctx context.Context, c *cli.Command) error {
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

	// Create done epic request
	request := lifecycle.DoneEpicRequest{
		EpicFile:  epicFile,
		Timestamp: timestamp,
	}

	// Complete the epic
	result, err := lifecycleService.DoneEpic(request)
	if err != nil {
		return outputDoneEpicError(c, err, format)
	}

	// Output the result
	return outputDoneEpicResult(c, result, format)
}

func outputDoneEpicResult(c *cli.Command, result *lifecycle.DoneEpicResult, format string) error {
	switch format {
	case "json":
		return outputDoneEpicJSON(c, result)
	case "xml":
		return outputDoneEpicXML(c, result)
	default:
		return outputDoneEpicText(c, result)
	}
}

func outputDoneEpicText(c *cli.Command, result *lifecycle.DoneEpicResult) error {
	fmt.Fprintf(c.Root().Writer, "Epic %s completed successfully\n", result.EpicID)
	fmt.Fprintf(c.Root().Writer, "Status: %s → %s\n", result.PreviousStatus, result.NewStatus)
	fmt.Fprintf(c.Root().Writer, "Completed at: %s\n", result.CompletedAt.Format(time.RFC3339))

	if result.Duration > 0 {
		fmt.Fprintf(c.Root().Writer, "Duration: %s\n", result.Duration)
	}

	fmt.Fprintf(c.Root().Writer, "\n%s\n", result.Message)

	// Show completion summary if available
	if result.Summary != "" {
		fmt.Fprintf(c.Root().Writer, "\nCompletion Summary:\n%s\n", result.Summary)
	}

	return nil
}

func outputDoneEpicJSON(c *cli.Command, result *lifecycle.DoneEpicResult) error {
	output := map[string]interface{}{
		"epic_completed": map[string]interface{}{
			"epic_id":         result.EpicID,
			"previous_status": result.PreviousStatus.String(),
			"new_status":      result.NewStatus.String(),
			"completed_at":    result.CompletedAt.Format(time.RFC3339),
			"duration":        result.Duration.String(),
			"event_created":   result.EventCreated,
			"message":         result.Message,
			"summary":         result.Summary,
		},
	}

	encoder := json.NewEncoder(c.Root().Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputDoneEpicXML(c *cli.Command, result *lifecycle.DoneEpicResult) error {
	doc := etree.NewDocument()
	root := doc.CreateElement("epic_completed")
	root.SetText("\n    ")

	// Add epic attribute
	root.CreateAttr("epic", result.EpicID)

	// Add child elements
	prevStatus := root.CreateElement("previous_status")
	prevStatus.SetText(result.PreviousStatus.String())

	newStatus := root.CreateElement("new_status")
	newStatus.SetText(result.NewStatus.String())

	completedAt := root.CreateElement("completed_at")
	completedAt.SetText(result.CompletedAt.Format(time.RFC3339))

	duration := root.CreateElement("duration")
	duration.SetText(result.Duration.String())

	eventCreated := root.CreateElement("event_created")
	if result.EventCreated {
		eventCreated.SetText("true")
	} else {
		eventCreated.SetText("false")
	}

	message := root.CreateElement("message")
	message.SetText(result.Message)

	if result.Summary != "" {
		summary := root.CreateElement("summary")
		summary.SetText(result.Summary)
	}

	doc.Indent(4)
	doc.WriteTo(c.Root().Writer)
	fmt.Fprintf(c.Root().Writer, "\n") // Add newline
	return nil
}

func outputDoneEpicError(c *cli.Command, err error, format string) error {
	// Check if it's a transition error for special handling
	if transitionErr, ok := err.(*lifecycle.TransitionError); ok {
		return outputDoneTransitionError(c, transitionErr, format)
	}

	// Check if it's a completion validation error for enhanced formatting
	if validationErr, ok := err.(*lifecycle.CompletionValidationError); ok {
		return outputDoneValidationError(c, validationErr, format)
	}

	// Generic error output
	fmt.Fprintf(c.Root().ErrWriter, "Error: %v\n", err)
	return err
}

func outputDoneTransitionError(c *cli.Command, err *lifecycle.TransitionError, format string) error {
	switch format {
	case "json":
		return outputDoneTransitionErrorJSON(c, err)
	case "xml":
		return outputDoneTransitionErrorXML(c, err)
	default:
		return outputDoneTransitionErrorText(c, err)
	}
}

func outputDoneTransitionErrorText(c *cli.Command, err *lifecycle.TransitionError) error {
	fmt.Fprintf(c.Root().ErrWriter, "Error: %s\n", err.Message)
	fmt.Fprintf(c.Root().ErrWriter, "Epic: %s\n", err.EpicID)
	fmt.Fprintf(c.Root().ErrWriter, "Current status: %s\n", err.CurrentStatus)
	if err.Suggestion != "" {
		fmt.Fprintf(c.Root().ErrWriter, "Suggestion: %s\n", err.Suggestion)
	}
	return err
}

func outputDoneTransitionErrorJSON(c *cli.Command, err *lifecycle.TransitionError) error {
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

func outputDoneTransitionErrorXML(c *cli.Command, err *lifecycle.TransitionError) error {
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

func outputDoneValidationError(c *cli.Command, err *lifecycle.CompletionValidationError, format string) error {
	switch format {
	case "json":
		return outputDoneValidationErrorJSON(c, err)
	case "xml":
		return outputDoneValidationErrorXML(c, err)
	default:
		return outputDoneValidationErrorText(c, err)
	}
}

func outputDoneValidationErrorText(c *cli.Command, err *lifecycle.CompletionValidationError) error {
	fmt.Fprintf(c.Root().ErrWriter, "Error: %s\n", err.Message)
	return err
}

func outputDoneValidationErrorJSON(c *cli.Command, err *lifecycle.CompletionValidationError) error {
	output := map[string]interface{}{
		"error": map[string]interface{}{
			"type":           "completion_validation",
			"epic_id":        err.EpicID,
			"message":        err.Message,
			"pending_phases": convertPendingPhasesToJSON(err.PendingPhases),
			"failing_tests":  convertFailingTestsToJSON(err.FailingTests),
		},
	}

	encoder := json.NewEncoder(c.Root().ErrWriter)
	encoder.SetIndent("", "  ")
	encoder.Encode(output)
	return err
}

func outputDoneValidationErrorXML(c *cli.Command, err *lifecycle.CompletionValidationError) error {
	doc := etree.NewDocument()
	root := doc.CreateElement("error")

	errorType := root.CreateElement("type")
	errorType.SetText("completion_validation")

	epicID := root.CreateElement("epic_id")
	epicID.SetText(err.EpicID)

	message := root.CreateElement("message")
	message.SetText(err.Message)

	// Add pending phases
	if len(err.PendingPhases) > 0 {
		pendingPhasesElem := root.CreateElement("pending_phases")
		for _, phase := range err.PendingPhases {
			phaseElem := pendingPhasesElem.CreateElement("phase")
			phaseElem.CreateAttr("id", phase.ID)
			phaseElem.SetText(phase.Name)
		}
	}

	// Add failing tests
	if len(err.FailingTests) > 0 {
		failingTestsElem := root.CreateElement("failing_tests")
		for _, test := range err.FailingTests {
			testElem := failingTestsElem.CreateElement("test")
			testElem.CreateAttr("id", test.ID)
			testElem.SetText(test.Name)
		}
	}

	doc.Indent(4)
	doc.WriteTo(c.Root().ErrWriter)
	fmt.Fprintf(c.Root().ErrWriter, "\n") // Add newline
	return err
}

func convertPendingPhasesToJSON(phases []lifecycle.PendingPhase) []map[string]string {
	result := make([]map[string]string, len(phases))
	for i, phase := range phases {
		result[i] = map[string]string{
			"id":   phase.ID,
			"name": phase.Name,
		}
	}
	return result
}

func convertFailingTestsToJSON(tests []lifecycle.FailingTest) []map[string]string {
	result := make([]map[string]string, len(tests))
	for i, test := range tests {
		result[i] = map[string]string{
			"id":          test.ID,
			"name":        test.Name,
			"description": test.Description,
		}
	}
	return result
}
