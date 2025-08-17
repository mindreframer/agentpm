package cmd

import (
	"fmt"

	"github.com/mindreframer/agentpm/internal/commands"
	"github.com/urfave/cli/v3"
)

func DoneCommand() *cli.Command {
	return &cli.Command{
		Name:  "done",
		Usage: "Complete something",
		Description: `Complete an epic, phase, or task.

Subcommands:
  epic           Complete current epic (transition from wip to done)
  phase <id>     Complete specific phase
  task <id>      Complete specific task

Examples:
  agentpm done epic                     # Complete current epic
  agentpm done phase 3A                 # Complete phase 3A
  agentpm done task 3A_1                # Complete task 3A_1`,
		Flags: commands.GlobalFlags(),
		Commands: []*cli.Command{
			doneEpicSubcommand(),
			donePhaseSubcommand(),
			doneTaskSubcommand(),
		},
	}
}

func doneEpicSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "epic",
		Usage: "Complete current epic",
		Description: `Complete an epic by transitioning its status from "wip" to "done".

This command:
- Changes epic status from "wip" to "done"
- Sets the completed_at timestamp
- Creates an automatic event log entry
- Validates that the epic is in a valid state to complete
- Generates a completion summary`,
		Action: commands.CreateEpicAction(handleDoneEpic),
	}
}

func donePhaseSubcommand() *cli.Command {
	return &cli.Command{
		Name:      "phase",
		Usage:     "Complete specific phase",
		ArgsUsage: "<phase-id>",
		Description: `Complete a specific phase in the epic.

The phase must exist in the current epic and all its tasks must be completed or cancelled.`,
		Action: commands.CreateEntityAction(commands.EntityTypePhase, handleDonePhase),
	}
}

func doneTaskSubcommand() *cli.Command {
	return &cli.Command{
		Name:      "task",
		Usage:     "Complete specific task",
		ArgsUsage: "<task-id>",
		Description: `Complete a specific task in the epic.

The task must exist in the current epic and be in a valid state to complete.`,
		Action: commands.CreateEntityAction(commands.EntityTypeTask, handleDoneTask),
	}
}

// Handler functions that bridge CLI to services

func handleDoneEpic(ctx commands.RouterContext) error {
	request := commands.DoneEpicRequest{
		ConfigPath: ctx.ConfigPath,
		EpicFile:   ctx.EpicFile,
		Time:       ctx.Time,
		Format:     ctx.Format,
	}

	result, err := commands.DoneEpicService(request)
	if err != nil {
		return err
	}

	// Handle different result types
	if result.IsAlreadyCompleted {
		// This is a friendly message, treat as success
		return nil
	}

	if result.Error != nil {
		return fmt.Errorf("%s", result.Error.Message)
	}

	// Normal success - epic was completed
	if result.Result != nil {
		fmt.Printf("Epic %s completed successfully\n", result.Result.EpicID)
		if result.Result.Summary != "" {
			fmt.Printf("\nCompletion Summary:\n%s\n", result.Result.Summary)
		}
	}
	return nil
}

func handleDonePhase(ctx commands.RouterContext, phaseID string) error {
	request := commands.DonePhaseRequest{
		PhaseID:    phaseID,
		ConfigPath: ctx.ConfigPath,
		EpicFile:   ctx.EpicFile,
		Time:       ctx.Time,
		Format:     ctx.Format,
	}

	result, err := commands.DonePhaseService(request)
	if err != nil {
		return err
	}

	if result.Error != nil {
		return fmt.Errorf("%s", result.Error.Message)
	}

	if result.IsAlreadyCompleted {
		// This is a friendly message, treat as success
		return nil
	}

	// Output success message
	fmt.Printf("Phase %s completed.\n", phaseID)
	return nil
}

func handleDoneTask(ctx commands.RouterContext, taskID string) error {
	request := commands.DoneTaskRequest{
		TaskID:     taskID,
		ConfigPath: ctx.ConfigPath,
		EpicFile:   ctx.EpicFile,
		Time:       ctx.Time,
		Format:     ctx.Format,
	}

	result, err := commands.DoneTaskService(request)
	if err != nil {
		return err
	}

	if result.Error != nil {
		return fmt.Errorf("%s", result.Error.Message)
	}

	if result.IsAlreadyCompleted {
		// This is a friendly message, treat as success
		return nil
	}

	// Output success message
	fmt.Printf("Task %s completed.\n", taskID)
	return nil
}
