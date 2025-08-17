package cmd

import (
	"context"
	"fmt"

	"github.com/mindreframer/agentpm/internal/commands"
	"github.com/urfave/cli/v3"
)

func FailCommand() *cli.Command {
	return &cli.Command{
		Name:      "fail",
		Usage:     "Mark a test as failed with reason",
		ArgsUsage: "<test-id> [reason]",
		Description: `Mark a test as failed with reason (transitions from wip to failed).

The test must exist in the current epic and be in a valid state to fail.
The failure reason is optional but recommended for tracking purposes.

Examples:
  agentpm fail 3A_T1 "Connection timeout"        # Fail test with reason
  agentpm fail 1B_T2                             # Fail test without reason
  agentpm fail 3A_T1 "Failed" --time 2025-08-16T15:30:00Z # Fail with timestamp`,
		Flags:  commands.GlobalFlags(),
		Action: failAction,
	}
}

func failAction(ctx context.Context, c *cli.Command) error {
	if c.Args().Len() < 1 {
		return fmt.Errorf("test ID is required")
	}

	testID := c.Args().First()

	// Get failure reason (optional)
	var failureReason string
	if c.Args().Len() >= 2 {
		failureReason = c.Args().Get(1)
	}

	// Extract router context
	routerCtx := commands.ExtractRouterContext(c)

	// Create test request
	request := commands.TestRequest{
		TestID:        testID,
		FailureReason: failureReason,
		ConfigPath:    routerCtx.ConfigPath,
		EpicFile:      routerCtx.EpicFile,
		Time:          routerCtx.Time,
		Format:        routerCtx.Format,
	}

	// Call the service
	result, err := commands.FailTestService(request)
	if err != nil {
		return err
	}

	// Handle service result
	if result.Error != nil {
		return fmt.Errorf("%s", result.Error.Message)
	}

	// Output success message based on result
	if result.Result != nil {
		if failureReason != "" {
			fmt.Printf("Test %s failed: %s\n", testID, failureReason)
		} else {
			fmt.Printf("Test %s failed.\n", testID)
		}
	}

	return nil
}
