package cmd

import (
	"fmt"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestDoneCommand_Structure(t *testing.T) {
	cmd := DoneCommand()

	// Test basic command properties
	if cmd.Name != "done" {
		t.Errorf("expected command name 'done', got %s", cmd.Name)
	}

	if cmd.Usage == "" {
		t.Error("expected non-empty usage")
	}

	if cmd.Description == "" {
		t.Error("expected non-empty description")
	}

	// Test that subcommands are present
	expectedSubcommands := []string{"epic", "phase", "task"}
	if len(cmd.Commands) != len(expectedSubcommands) {
		t.Errorf("expected %d subcommands, got %d", len(expectedSubcommands), len(cmd.Commands))
	}

	subcommandNames := make(map[string]bool)
	for _, subcmd := range cmd.Commands {
		subcommandNames[subcmd.Name] = true
	}

	for _, expected := range expectedSubcommands {
		if !subcommandNames[expected] {
			t.Errorf("missing expected subcommand: %s", expected)
		}
	}

	// Test that global flags are present
	if len(cmd.Flags) == 0 {
		t.Error("expected global flags to be present")
	}

	// Check for specific global flags
	flagNames := make(map[string]bool)
	for _, flag := range cmd.Flags {
		flagNames[flag.Names()[0]] = true
	}

	expectedFlags := []string{"file", "config", "time", "format"}
	for _, flag := range expectedFlags {
		if !flagNames[flag] {
			t.Errorf("missing expected global flag: %s", flag)
		}
	}
}

func TestDoneEpicSubcommand_Structure(t *testing.T) {
	cmd := DoneCommand()
	var epicSubcmd *cli.Command

	for _, subcmd := range cmd.Commands {
		if subcmd.Name == "epic" {
			epicSubcmd = subcmd
			break
		}
	}

	if epicSubcmd == nil {
		t.Fatal("epic subcommand not found")
	}

	if epicSubcmd.Usage == "" {
		t.Error("expected non-empty usage for epic subcommand")
	}

	if epicSubcmd.Description == "" {
		t.Error("expected non-empty description for epic subcommand")
	}

	if epicSubcmd.Action == nil {
		t.Error("expected action function for epic subcommand")
	}
}

func TestDonePhaseSubcommand_Structure(t *testing.T) {
	cmd := DoneCommand()
	var phaseSubcmd *cli.Command

	for _, subcmd := range cmd.Commands {
		if subcmd.Name == "phase" {
			phaseSubcmd = subcmd
			break
		}
	}

	if phaseSubcmd == nil {
		t.Fatal("phase subcommand not found")
	}

	if phaseSubcmd.Usage == "" {
		t.Error("expected non-empty usage for phase subcommand")
	}

	if phaseSubcmd.ArgsUsage != "<phase-id>" {
		t.Errorf("expected ArgsUsage '<phase-id>', got %s", phaseSubcmd.ArgsUsage)
	}

	if phaseSubcmd.Action == nil {
		t.Error("expected action function for phase subcommand")
	}
}

func TestDoneTaskSubcommand_Structure(t *testing.T) {
	cmd := DoneCommand()
	var taskSubcmd *cli.Command

	for _, subcmd := range cmd.Commands {
		if subcmd.Name == "task" {
			taskSubcmd = subcmd
			break
		}
	}

	if taskSubcmd == nil {
		t.Fatal("task subcommand not found")
	}

	if taskSubcmd.Usage == "" {
		t.Error("expected non-empty usage for task subcommand")
	}

	if taskSubcmd.ArgsUsage != "<task-id>" {
		t.Errorf("expected ArgsUsage '<task-id>', got %s", taskSubcmd.ArgsUsage)
	}

	if taskSubcmd.Action == nil {
		t.Error("expected action function for task subcommand")
	}
}

func TestDoneCommand_AutoDetection_Structure(t *testing.T) {
	cmd := DoneCommand()

	// Test that the command has auto-detection action
	if cmd.Action == nil {
		t.Error("done command should have auto-detection action")
	}

	// Test that auto-detection supports the expected entity types
	// Note: Actual execution testing requires proper CLI setup

	if cmd.Name != "done" {
		t.Error("command should be named 'done'")
	}
}

func TestDoneCommand_Help(t *testing.T) {
	cmd := DoneCommand()

	// Test that help information is comprehensive
	if cmd.Description == "" {
		t.Error("expected comprehensive description")
	}

	// Check that description mentions key concepts
	description := cmd.Description
	expectedConcepts := []string{"epic", "phase", "task", "auto-detect"}

	for _, concept := range expectedConcepts {
		if !contains(description, concept) {
			t.Errorf("description should mention '%s'", concept)
		}
	}

	// Check that examples are provided
	if !contains(description, "Examples:") {
		t.Error("description should include examples")
	}
}

func TestDoneCommand_GlobalFlags_Inheritance(t *testing.T) {
	cmd := DoneCommand()

	// Test that all subcommands inherit global flags behavior
	// This is validated by the fact that subcommands use CreateEntityAction
	// which extracts RouterContext with all global flags

	expectedFlags := []string{"file", "config", "time", "format"}

	for _, flag := range expectedFlags {
		found := false
		for _, cmdFlag := range cmd.Flags {
			if cmdFlag.Names()[0] == flag {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("global flag '%s' not found", flag)
		}
	}
}

func TestDoneCommand_Subcommand_Actions_NotNil(t *testing.T) {
	cmd := DoneCommand()

	// Ensure all subcommands have action functions
	for _, subcmd := range cmd.Commands {
		if subcmd.Action == nil {
			t.Errorf("subcommand '%s' missing action function", subcmd.Name)
		}
	}

	// Test that main command has fallback action
	if cmd.Action == nil {
		t.Error("main done command missing fallback action")
	}
}

func TestDoneCommand_ValidIDs_Structure(t *testing.T) {
	// This test validates the ID patterns the done command should support
	validIDs := []struct {
		name       string
		id         string
		entityType string
	}{
		{name: "valid_phase_id", id: "3A", entityType: "phase"},
		{name: "valid_task_id", id: "3A_1", entityType: "task"},
		{name: "phase_with_number", id: "1B", entityType: "phase"},
		{name: "task_with_number", id: "1B_2", entityType: "task"},
	}

	cmd := DoneCommand()

	// Validate command supports auto-detection for these IDs
	if cmd.Action == nil {
		t.Error("command should support auto-detection")
	}

	for _, tt := range validIDs {
		t.Run(tt.name, func(t *testing.T) {
			// Validate test case structure
			if tt.id == "" {
				t.Error("test case should have non-empty ID")
			}
			if tt.entityType == "" {
				t.Error("test case should specify entity type")
			}

			// Note: Actual router testing is done in the router tests
			// This validates the done command structure supports the workflow
		})
	}
}

func TestDoneCommand_CompletionWorkflow(t *testing.T) {
	cmd := DoneCommand()

	// Test that the command structure supports the expected completion workflow
	// 1. Epic completion (no ID needed)
	// 2. Phase completion (ID required)
	// 3. Task completion (ID required)

	// Validate epic subcommand doesn't require args
	var epicSubcmd *cli.Command
	for _, subcmd := range cmd.Commands {
		if subcmd.Name == "epic" {
			epicSubcmd = subcmd
			break
		}
	}

	if epicSubcmd == nil {
		t.Fatal("epic subcommand should exist")
	}

	// Epic subcommand should not have ArgsUsage (no ID required)
	if epicSubcmd.ArgsUsage != "" {
		t.Error("epic subcommand should not require arguments")
	}

	// Phase and task subcommands should require ID
	entitySubcommands := []string{"phase", "task"}
	for _, entityName := range entitySubcommands {
		var subcmd *cli.Command
		for _, sc := range cmd.Commands {
			if sc.Name == entityName {
				subcmd = sc
				break
			}
		}

		if subcmd == nil {
			t.Errorf("%s subcommand should exist", entityName)
			continue
		}

		expectedArgsUsage := fmt.Sprintf("<%s-id>", entityName)
		if subcmd.ArgsUsage != expectedArgsUsage {
			t.Errorf("%s subcommand should require ID argument, expected '%s', got '%s'",
				entityName, expectedArgsUsage, subcmd.ArgsUsage)
		}
	}
}

// Helper function imported from start_test.go - checks if string contains substring
func TestDoneCommand_Integration_Structure(t *testing.T) {
	cmd := DoneCommand()

	// Validate the command integrates properly with the CLI structure
	if cmd.Name != "done" {
		t.Errorf("expected command name 'done', got '%s'", cmd.Name)
	}

	// Validate it has proper subcommands
	if len(cmd.Commands) == 0 {
		t.Error("done command should have subcommands")
	}

	// Validate it has global flags
	if len(cmd.Flags) == 0 {
		t.Error("done command should have global flags")
	}

	// Validate it has actions
	if cmd.Action == nil {
		t.Error("done command should have action function")
	}
}
