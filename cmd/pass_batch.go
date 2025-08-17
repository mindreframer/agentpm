package cmd

import (
	"context"
	"fmt"

	"github.com/mindreframer/agentpm/internal/commands"
	"github.com/urfave/cli/v3"
)

func PassBatchCommand() *cli.Command {
	return &cli.Command{
		Name:      "pass-batch",
		Usage:     "Mark multiple tests as passed in a single batch operation",
		ArgsUsage: "<test-id1> <test-id2> ... <test-idN>",
		Description: `Mark multiple tests as passed using batch validation.

This command uses all-or-nothing validation - if any test cannot be passed,
the entire batch operation fails and no tests are modified.

All tests must:
- Exist in the current epic
- Be in the active phase
- Be in WIP status
- Pass Epic 13 validation rules

Examples:
  agentpm pass-batch 3A_T1 3A_T2 3A_T3                    # Pass multiple tests
  agentpm pass-batch 1B_T1 1B_T2 --time 2025-08-16T15:30:00Z # Pass with timestamp`,
		Flags:  commands.GlobalFlags(),
		Action: passBatchAction,
	}
}

func passBatchAction(ctx context.Context, c *cli.Command) error {
	if c.Args().Len() < 1 {
		return fmt.Errorf("at least one test ID is required")
	}

	testIDs := c.Args().Slice()

	// Extract router context
	routerCtx := commands.ExtractRouterContext(c)

	// Create batch request
	request := commands.BatchTestRequest{
		TestIDs:    testIDs,
		Operation:  "pass",
		ConfigPath: routerCtx.ConfigPath,
		EpicFile:   routerCtx.EpicFile,
		Time:       routerCtx.Time,
		Format:     routerCtx.Format,
	}

	// Call the batch service
	result, err := commands.PassBatchTestService(request)
	if err != nil {
		return err
	}

	// Handle service result
	if result.Error != nil {
		return fmt.Errorf("%s", result.Error.Message)
	}

	// Output success message
	if result.Result != nil {
		summary := result.Result.Summary
		fmt.Printf("Batch pass completed successfully:\n")
		fmt.Printf("- %d tests passed\n", summary.PassedTests)

		if len(result.Result.SuccessfulOperations) > 0 {
			fmt.Printf("\nPassed tests:\n")
			for _, op := range result.Result.SuccessfulOperations {
				if op.TestName != "" {
					fmt.Printf("  %s (%s)\n", op.TestID, op.TestName)
				} else {
					fmt.Printf("  %s\n", op.TestID)
				}
			}
		}
	}

	return nil
}
