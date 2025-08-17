package cmd

import (
	"context"
	"fmt"

	"github.com/mindreframer/agentpm/internal/commands"
	"github.com/urfave/cli/v3"
)

func PassCommand() *cli.Command {
	return &cli.Command{
		Name:      "pass",
		Usage:     "Mark a test as passed",
		ArgsUsage: "<test-id>",
		Description: `Mark a test as passed (transitions from wip to passed).

The test must exist in the current epic and be in a valid state to pass.

Examples:
  agentpm pass 3A_T1                    # Pass test 3A_T1
  agentpm pass 1B_T2 --time 2025-08-16T15:30:00Z # Pass with specific timestamp`,
		Flags:  commands.GlobalFlags(),
		Action: passAction,
	}
}

func passAction(ctx context.Context, c *cli.Command) error {
	if c.Args().Len() < 1 {
		return fmt.Errorf("test ID is required")
	}

	testID := c.Args().First()

	// Extract router context
	routerCtx := commands.ExtractRouterContext(c)

	// Create test request
	request := commands.TestRequest{
		TestID:     testID,
		ConfigPath: routerCtx.ConfigPath,
		EpicFile:   routerCtx.EpicFile,
		Time:       routerCtx.Time,
		Format:     routerCtx.Format,
	}

	// Call the service
	result, err := commands.PassTestService(request)
	if err != nil {
		return err
	}

	// Handle service result
	if result.Error != nil {
		return fmt.Errorf("%s", result.Error.Message)
	}

	// Output success message based on result
	if result.Result != nil {
		fmt.Printf("Test %s passed.\n", testID)
	}

	return nil
}
