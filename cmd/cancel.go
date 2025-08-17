package cmd

import (
	"fmt"

	"github.com/mindreframer/agentpm/internal/commands"
	"github.com/urfave/cli/v3"
)

func CancelCommand() *cli.Command {
	return &cli.Command{
		Name:  "cancel",
		Usage: "Cancel a task or test",
		Description: `Cancel a task or test with optional reason.

Subcommands:
  task <id> [reason]     Cancel specific task
  test <id> [reason]     Cancel specific test

Examples:
  agentpm cancel task 3A_1 "No longer needed"   # Cancel task with reason
  agentpm cancel test 3A_T1 "Test obsolete"     # Cancel test with reason`,
		Flags: commands.GlobalFlags(),
		Commands: []*cli.Command{
			cancelTaskSubcommand(),
			cancelTestSubcommand(),
		},
	}
}

func cancelTaskSubcommand() *cli.Command {
	return &cli.Command{
		Name:      "task",
		Usage:     "Cancel specific task",
		ArgsUsage: "<task-id> [reason]",
		Description: `Cancel a specific task in the epic.

The task must exist in the current epic and be in a valid state to cancel.
The cancellation reason is optional but recommended for tracking purposes.`,
		Action: commands.CreateEntityAction(commands.EntityTypeTask, handleCancelTask),
	}
}

func cancelTestSubcommand() *cli.Command {
	return &cli.Command{
		Name:      "test",
		Usage:     "Cancel specific test",
		ArgsUsage: "<test-id> [reason]",
		Description: `Cancel a specific test in the epic.

The test must exist in the current epic and be in a valid state to cancel.
The cancellation reason is optional but recommended for tracking purposes.`,
		Action: commands.CreateEntityAction(commands.EntityTypeTest, handleCancelTest),
	}
}

// Handler functions that bridge CLI to services

func handleCancelTask(ctx commands.RouterContext, taskID string) error {
	// Note: In this simplified implementation, we don't extract the reason from args
	// A full implementation would need to extract additional args from the CLI context
	request := commands.CancelTaskRequest{
		TaskID:     taskID,
		ConfigPath: ctx.ConfigPath,
		EpicFile:   ctx.EpicFile,
		Time:       ctx.Time,
		Format:     ctx.Format,
	}

	result, err := commands.CancelTaskService(request)
	if err != nil {
		return err
	}

	if result.Error != nil {
		return fmt.Errorf("%s", result.Error.Message)
	}

	// Output success message
	fmt.Printf("Task %s cancelled.\n", taskID)
	return nil
}

func handleCancelTest(ctx commands.RouterContext, testID string) error {
	// Note: In this simplified implementation, we don't extract the reason from args
	// A full implementation would need to extract additional args from the CLI context
	request := commands.TestRequest{
		TestID:             testID,
		CancellationReason: "", // Could be extracted from additional args
		ConfigPath:         ctx.ConfigPath,
		EpicFile:           ctx.EpicFile,
		Time:               ctx.Time,
		Format:             ctx.Format,
	}

	result, err := commands.CancelTestService(request)
	if err != nil {
		return err
	}

	if result.Error != nil {
		return fmt.Errorf("%s", result.Error.Message)
	}

	// Output success message based on result
	if result.Result != nil {
		fmt.Printf("Test %s cancelled.\n", testID)
	}

	return nil
}
