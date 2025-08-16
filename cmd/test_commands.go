package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/memomoo/agentpm/internal/config"
	"github.com/memomoo/agentpm/internal/tests"
	"github.com/urfave/cli/v3"
)

// StartTestCommand creates a start-test command for beginning test work
func StartTestCommand() *cli.Command {
	return &cli.Command{
		Name:    "start-test",
		Usage:   "Start working on a test (transitions from pending to wip) - Usage: start-test <test-id>",
		Aliases: []string{"st"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Override epic file from config",
			},
			&cli.StringFlag{
				Name:    "time",
				Aliases: []string{"t"},
				Usage:   "Custom timestamp (ISO8601 format)",
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"F"},
				Usage:   "Output format: text (default), json, xml",
				Value:   "text",
			},
		},
		Action: startTestAction,
	}
}

// PassTestCommand creates a pass-test command for marking tests as passed
func PassTestCommand() *cli.Command {
	return &cli.Command{
		Name:    "pass-test",
		Usage:   "Mark a test as passed (transitions from wip to passed) - Usage: pass-test <test-id>",
		Aliases: []string{"pt"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Override epic file from config",
			},
			&cli.StringFlag{
				Name:    "time",
				Aliases: []string{"t"},
				Usage:   "Custom timestamp (ISO8601 format)",
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"F"},
				Usage:   "Output format: text (default), json, xml",
				Value:   "text",
			},
		},
		Action: passTestAction,
	}
}

// FailTestCommand creates a fail-test command for marking tests as failed
func FailTestCommand() *cli.Command {
	return &cli.Command{
		Name:    "fail-test",
		Usage:   `Mark a test as failed with reason (transitions from wip to failed) - Usage: fail-test <test-id> "reason"`,
		Aliases: []string{"ft"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Override epic file from config",
			},
			&cli.StringFlag{
				Name:    "time",
				Aliases: []string{"t"},
				Usage:   "Custom timestamp (ISO8601 format)",
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"F"},
				Usage:   "Output format: text (default), json, xml",
				Value:   "text",
			},
		},
		Action: failTestAction,
	}
}

// CancelTestCommand creates a cancel-test command for cancelling tests
func CancelTestCommand() *cli.Command {
	return &cli.Command{
		Name:    "cancel-test",
		Usage:   `Cancel a test with reason (transitions from wip to cancelled) - Usage: cancel-test <test-id> "reason"`,
		Aliases: []string{"ct"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Override epic file from config",
			},
			&cli.StringFlag{
				Name:    "time",
				Aliases: []string{"t"},
				Usage:   "Custom timestamp (ISO8601 format)",
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"F"},
				Usage:   "Output format: text (default), json, xml",
				Value:   "text",
			},
		},
		Action: cancelTestAction,
	}
}

func startTestAction(ctx context.Context, c *cli.Command) error {
	// Validate arguments
	args := c.Args()
	if args.Len() != 1 {
		return fmt.Errorf("start-test requires exactly one argument: test-id")
	}
	testID := args.Get(0)

	// Load configuration and determine epic file
	epicFile, err := getEpicFile(c)
	if err != nil {
		return err
	}

	// Parse timestamp if provided
	var timestamp *time.Time
	if timeStr := c.String("time"); timeStr != "" {
		t, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return fmt.Errorf("invalid time format: %s (expected ISO8601/RFC3339)", timeStr)
		}
		timestamp = &t
	}

	// Create test service
	service := tests.NewTestService(tests.ServiceConfig{
		UseMemory: false,
	})

	// Execute operation
	result, err := service.StartTest(epicFile, testID, timestamp)
	if err != nil {
		return writeTestError(c, c.String("format"), err)
	}

	return writeTestResult(c, c.String("format"), result)
}

func passTestAction(ctx context.Context, c *cli.Command) error {
	// Validate arguments
	args := c.Args()
	if args.Len() != 1 {
		return fmt.Errorf("pass-test requires exactly one argument: test-id")
	}
	testID := args.Get(0)

	// Load configuration and determine epic file
	epicFile, err := getEpicFile(c)
	if err != nil {
		return err
	}

	// Parse timestamp if provided
	var timestamp *time.Time
	if timeStr := c.String("time"); timeStr != "" {
		t, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return fmt.Errorf("invalid time format: %s (expected ISO8601/RFC3339)", timeStr)
		}
		timestamp = &t
	}

	// Create test service
	service := tests.NewTestService(tests.ServiceConfig{
		UseMemory: false,
	})

	// Execute operation
	result, err := service.PassTest(epicFile, testID, timestamp)
	if err != nil {
		return writeTestError(c, c.String("format"), err)
	}

	return writeTestResult(c, c.String("format"), result)
}

func failTestAction(ctx context.Context, c *cli.Command) error {
	// Validate arguments
	args := c.Args()
	if args.Len() != 2 {
		return fmt.Errorf("fail-test requires exactly two arguments: test-id \"failure-reason\"")
	}
	testID := args.Get(0)
	failureReason := args.Get(1)

	// Load configuration and determine epic file
	epicFile, err := getEpicFile(c)
	if err != nil {
		return err
	}

	// Parse timestamp if provided
	var timestamp *time.Time
	if timeStr := c.String("time"); timeStr != "" {
		t, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return fmt.Errorf("invalid time format: %s (expected ISO8601/RFC3339)", timeStr)
		}
		timestamp = &t
	}

	// Create test service
	service := tests.NewTestService(tests.ServiceConfig{
		UseMemory: false,
	})

	// Execute operation
	result, err := service.FailTest(epicFile, testID, failureReason, timestamp)
	if err != nil {
		return writeTestError(c, c.String("format"), err)
	}

	return writeTestResult(c, c.String("format"), result)
}

func cancelTestAction(ctx context.Context, c *cli.Command) error {
	// Validate arguments
	args := c.Args()
	if args.Len() != 2 {
		return fmt.Errorf("cancel-test requires exactly two arguments: test-id \"cancellation-reason\"")
	}
	testID := args.Get(0)
	cancellationReason := args.Get(1)

	// Load configuration and determine epic file
	epicFile, err := getEpicFile(c)
	if err != nil {
		return err
	}

	// Parse timestamp if provided
	var timestamp *time.Time
	if timeStr := c.String("time"); timeStr != "" {
		t, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return fmt.Errorf("invalid time format: %s (expected ISO8601/RFC3339)", timeStr)
		}
		timestamp = &t
	}

	// Create test service
	service := tests.NewTestService(tests.ServiceConfig{
		UseMemory: false,
	})

	// Execute operation
	result, err := service.CancelTest(epicFile, testID, cancellationReason, timestamp)
	if err != nil {
		return writeTestError(c, c.String("format"), err)
	}

	return writeTestResult(c, c.String("format"), result)
}

// Helper functions

func getEpicFile(c *cli.Command) (string, error) {
	// Check if file is provided directly
	epicFile := c.String("file")
	if epicFile != "" {
		return epicFile, nil
	}

	// Load configuration
	configPath := c.String("config")
	if configPath == "" {
		configPath = "./.agentpm.json"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to load configuration: %w", err)
	}

	// Use epic file from config
	if cfg.CurrentEpic == "" {
		return "", fmt.Errorf("no epic file specified. Use --file flag or run 'agentpm init' first")
	}

	return cfg.CurrentEpic, nil
}

func writeTestResult(c *cli.Command, format string, result *tests.TestOperation) error {
	switch format {
	case "xml":
		return writeTestResultXML(c, result)
	case "json":
		return writeTestResultJSON(c, result)
	default:
		return writeTestResultText(c, result)
	}
}

func writeTestResultText(c *cli.Command, result *tests.TestOperation) error {
	switch result.Operation {
	case "started":
		fmt.Fprintf(c.Root().Writer, "Test %s started.\n", result.TestID)
	case "passed":
		fmt.Fprintf(c.Root().Writer, "Test %s passed.\n", result.TestID)
	case "failed":
		fmt.Fprintf(c.Root().Writer, "Test %s failed: %s\n", result.TestID, result.FailureReason)
	case "cancelled":
		fmt.Fprintf(c.Root().Writer, "Test %s cancelled: %s\n", result.TestID, result.CancellationReason)
	}
	return nil
}

func writeTestResultJSON(c *cli.Command, result *tests.TestOperation) error {
	output := fmt.Sprintf(`{
  "test_id": "%s",
  "operation": "%s",
  "status": "%s",
  "timestamp": "%s"`,
		result.TestID,
		result.Operation,
		result.Status,
		result.Timestamp.Format(time.RFC3339),
	)

	if result.FailureReason != "" {
		output += fmt.Sprintf(`,
  "failure_reason": "%s"`, result.FailureReason)
	}

	if result.CancellationReason != "" {
		output += fmt.Sprintf(`,
  "cancellation_reason": "%s"`, result.CancellationReason)
	}

	output += `
}`

	fmt.Fprintf(c.Root().Writer, "%s\n", output)
	return nil
}

func writeTestResultXML(c *cli.Command, result *tests.TestOperation) error {
	output := fmt.Sprintf(`<test_operation>
    <test_id>%s</test_id>
    <operation>%s</operation>
    <status>%s</status>
    <timestamp>%s</timestamp>`,
		result.TestID,
		result.Operation,
		result.Status,
		result.Timestamp.Format(time.RFC3339),
	)

	if result.FailureReason != "" {
		output += fmt.Sprintf(`
    <failure_reason>%s</failure_reason>`, result.FailureReason)
	}

	if result.CancellationReason != "" {
		output += fmt.Sprintf(`
    <cancellation_reason>%s</cancellation_reason>`, result.CancellationReason)
	}

	output += `
</test_operation>`

	fmt.Fprintf(c.Root().Writer, "%s\n", output)
	return nil
}

func writeTestError(c *cli.Command, format string, err error) error {
	if testErr, ok := err.(*tests.TestError); ok {
		switch format {
		case "xml":
			output := fmt.Sprintf(`<error>
    <type>%s</type>
    <test_id>%s</test_id>
    <message>%s</message>
</error>`, testErr.Type, testErr.TestID, testErr.Message)
			fmt.Fprint(c.Root().ErrWriter, output)
		case "json":
			output := fmt.Sprintf(`{
  "error": {
    "type": "%s",
    "test_id": "%s",
    "message": "%s"
  }
}`, testErr.Type, testErr.TestID, testErr.Message)
			fmt.Fprint(c.Root().ErrWriter, output)
		default: // text
			fmt.Fprintf(c.Root().ErrWriter, "✗ Error: %s\n", testErr.Message)
		}
	} else {
		// Fallback for non-test errors
		fmt.Fprintf(c.Root().ErrWriter, "✗ Error: %s\n", err.Error())
	}
	return err
}
