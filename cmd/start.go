package cmd

import (
	"fmt"

	"github.com/mindreframer/agentpm/internal/commands"
	"github.com/urfave/cli/v3"
)

func StartCommand() *cli.Command {
	return &cli.Command{
		Name:  "start",
		Usage: "Start working on something",
		Description: `Start working on an epic, phase, task, or test.

Subcommands:
  epic           Start working on current epic (transition from pending to wip)
  phase <id>     Start working on specific phase
  task <id>      Start working on specific task
  test <id>      Start test execution

Examples:
  agentpm start epic                    # Start current epic
  agentpm start phase 3A                # Start phase 3A
  agentpm start task 3A_1               # Start task 3A_1
  agentpm start test 3A_T1              # Start test 3A_T1`,
		Flags: commands.GlobalFlags(),
		Commands: []*cli.Command{
			startEpicSubcommand(),
			startPhaseSubcommand(),
			startTaskSubcommand(),
			startTestSubcommand(),
		},
	}
}

func startEpicSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "epic",
		Usage: "Start working on current epic",
		Description: `Start an epic by transitioning its status from "pending" to "wip".
		
This command:
- Changes epic status from "pending" to "wip"
- Sets the started_at timestamp
- Creates an automatic event log entry
- Validates that the epic is in a valid state to start`,
		Action: commands.CreateEpicAction(handleStartEpic),
	}
}

func startPhaseSubcommand() *cli.Command {
	return &cli.Command{
		Name:      "phase",
		Usage:     "Start working on specific phase",
		ArgsUsage: "<phase-id>",
		Description: `Start a specific phase in the epic.

The phase must exist in the current epic and be in a valid state to start.`,
		Action: commands.CreateEntityAction(commands.EntityTypePhase, handleStartPhase),
	}
}

func startTaskSubcommand() *cli.Command {
	return &cli.Command{
		Name:      "task",
		Usage:     "Start working on specific task",
		ArgsUsage: "<task-id>",
		Description: `Start a specific task in the epic.

The task must exist in the current epic and its phase must be active.`,
		Action: commands.CreateEntityAction(commands.EntityTypeTask, handleStartTask),
	}
}

func startTestSubcommand() *cli.Command {
	return &cli.Command{
		Name:      "test",
		Usage:     "Start test execution",
		ArgsUsage: "<test-id>",
		Description: `Start working on a test (transitions from pending to wip).

The test must exist in the current epic and be in a valid state to start.`,
		Action: commands.CreateEntityAction(commands.EntityTypeTest, handleStartTest),
	}
}

// Handler functions that bridge CLI to services

func handleStartEpic(ctx commands.RouterContext) error {
	request := commands.StartEpicRequest{
		ConfigPath: ctx.ConfigPath,
		EpicFile:   ctx.EpicFile,
		Time:       ctx.Time,
		Format:     ctx.Format,
	}

	result, err := commands.StartEpicService(request)
	if err != nil {
		return err
	}

	// Handle different result types
	if result.IsAlreadyStarted || result.IsAlreadyCompleted {
		// This is a friendly message, treat as success
		return nil
	}

	// Normal success - epic was started
	return nil
}

func handleStartPhase(ctx commands.RouterContext, phaseID string) error {
	request := commands.StartPhaseRequest{
		PhaseID:    phaseID,
		ConfigPath: ctx.ConfigPath,
		EpicFile:   ctx.EpicFile,
		Time:       ctx.Time,
		Format:     ctx.Format,
	}

	result, err := commands.StartPhaseService(request)
	if err != nil {
		return err
	}

	if result.Error != nil {
		return fmt.Errorf("%s", result.Error.Message)
	}

	if result.IsAlreadyActive {
		// This is a friendly message, treat as success
		return nil
	}

	// Output success message
	fmt.Printf("Phase %s started.\n", phaseID)
	return nil
}

func handleStartTask(ctx commands.RouterContext, taskID string) error {
	request := commands.StartTaskRequest{
		TaskID:     taskID,
		ConfigPath: ctx.ConfigPath,
		EpicFile:   ctx.EpicFile,
		Time:       ctx.Time,
		Format:     ctx.Format,
	}

	result, err := commands.StartTaskService(request)
	if err != nil {
		return err
	}

	if result.Error != nil {
		return fmt.Errorf("%s", result.Error.Message)
	}

	if result.IsAlreadyActive {
		// This is a friendly message, treat as success
		return nil
	}

	// Output success message
	fmt.Printf("Task %s started.\n", taskID)
	return nil
}

func handleStartTest(ctx commands.RouterContext, testID string) error {
	request := commands.TestRequest{
		TestID:     testID,
		ConfigPath: ctx.ConfigPath,
		EpicFile:   ctx.EpicFile,
		Time:       ctx.Time,
		Format:     ctx.Format,
	}

	result, err := commands.StartTestService(request)
	if err != nil {
		return err
	}

	if result.Error != nil {
		return fmt.Errorf("%s", result.Error.Message)
	}

	// Output success message based on result
	if result.Result != nil {
		if result.Result.Status == "already_started" {
			fmt.Printf("Test %s is already started.\n", testID)
		} else {
			fmt.Printf("Test %s started.\n", testID)
		}
	}
	return nil
}
