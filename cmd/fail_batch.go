package cmd

import (
	"context"
	"fmt"

	"github.com/mindreframer/agentpm/internal/commands"
	"github.com/urfave/cli/v3"
)

func FailBatchCommand() *cli.Command {
	return &cli.Command{
		Name:      "fail-batch",
		Usage:     "Mark multiple tests as failed in a single batch operation",
		ArgsUsage: "<test-id1> <test-id2> ... <test-idN> [reason]",
		Description: `Mark multiple tests as failed using batch validation.

This command uses all-or-nothing validation - if any test cannot be failed,
the entire batch operation fails and no tests are modified.

All tests must:
- Exist in the current epic
- Be in the active phase  
- Be in WIP or Done status
- Pass Epic 13 validation rules

The failure reason is optional but recommended for tracking purposes.

Examples:
  agentpm fail-batch 3A_T1 3A_T2 3A_T3 "Connection timeout"     # Fail multiple tests with reason
  agentpm fail-batch 1B_T1 1B_T2                                # Fail without reason
  agentpm fail-batch 3A_T1 3A_T2 "Failed" --time 2025-08-16T15:30:00Z # Fail with timestamp`,
		Flags:  commands.GlobalFlags(),
		Action: failBatchAction,
	}
}

func failBatchAction(ctx context.Context, c *cli.Command) error {
	if c.Args().Len() < 1 {
		return fmt.Errorf("at least one test ID is required")
	}

	args := c.Args().Slice()

	// Extract test IDs and optional reason
	var testIDs []string
	var failureReason string

	// If last argument looks like a reason (not a test ID), use it as reason
	if len(args) > 1 {
		// Simple heuristic: if last arg doesn't match test ID pattern, treat as reason
		lastArg := args[len(args)-1]
		if len(lastArg) > 10 || (len(lastArg) > 0 && lastArg[0] != '1' && lastArg[0] != '2' && lastArg[0] != '3') {
			testIDs = args[:len(args)-1]
			failureReason = lastArg
		} else {
			testIDs = args
		}
	} else {
		testIDs = args
	}

	// Extract router context
	routerCtx := commands.ExtractRouterContext(c)

	// Create batch request
	request := commands.BatchTestRequest{
		TestIDs:       testIDs,
		Operation:     "fail",
		FailureReason: failureReason,
		ConfigPath:    routerCtx.ConfigPath,
		EpicFile:      routerCtx.EpicFile,
		Time:          routerCtx.Time,
		Format:        routerCtx.Format,
	}

	// Call the batch service
	result, err := commands.FailBatchTestService(request)
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
		fmt.Printf("Batch fail completed successfully:\n")
		fmt.Printf("- %d tests failed\n", summary.FailedTests)

		if failureReason != "" {
			fmt.Printf("- Reason: %s\n", failureReason)
		}

		if len(result.Result.SuccessfulOperations) > 0 {
			fmt.Printf("\nFailed tests:\n")
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
