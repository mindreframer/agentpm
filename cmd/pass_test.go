package cmd

import (
	"testing"
)

func TestPassCommand_Structure(t *testing.T) {
	cmd := PassCommand()

	// Test basic command properties
	if cmd.Name != "pass" {
		t.Errorf("expected command name 'pass', got %s", cmd.Name)
	}

	if cmd.Usage == "" {
		t.Error("expected non-empty usage")
	}

	if cmd.ArgsUsage != "<test-id>" {
		t.Errorf("expected ArgsUsage '<test-id>', got %s", cmd.ArgsUsage)
	}

	if cmd.Description == "" {
		t.Error("expected non-empty description")
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

	// Test that action function exists
	if cmd.Action == nil {
		t.Error("expected action function")
	}
}

func TestPassCommand_Help(t *testing.T) {
	cmd := PassCommand()

	// Test that help information is comprehensive
	if cmd.Description == "" {
		t.Error("expected comprehensive description")
	}

	// Check that description mentions key concepts
	description := cmd.Description
	expectedConcepts := []string{"test", "passed", "wip"}

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

func TestPassCommand_GlobalFlags_Inheritance(t *testing.T) {
	cmd := PassCommand()

	// Test that command has global flags
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

func TestPassCommand_Arguments(t *testing.T) {
	cmd := PassCommand()

	// Test that command requires test ID
	if cmd.ArgsUsage == "" {
		t.Error("pass command should specify required arguments")
	}

	// Test that it expects test ID
	if cmd.ArgsUsage != "<test-id>" {
		t.Errorf("expected args usage '<test-id>', got '%s'", cmd.ArgsUsage)
	}
}
