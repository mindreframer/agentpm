package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

// TestCommandsBasic tests the basic functionality of CLI commands without deep inspection
func TestCommandsBasic(t *testing.T) {
	// Create test epic file for CLI testing
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	// Create a minimal epic file that the CLI can work with
	epicContent := `<?xml version="1.0" encoding="UTF-8"?>
<epic id="test-epic" name="Test Epic" status="wip" created_at="2025-08-16T14:00:00Z">
    <phases>
        <phase id="phase-1" name="Phase 1" status="wip" />
    </phases>
    <tasks>
        <task id="task-1" phase_id="phase-1" name="Task 1" status="wip" />
    </tasks>
    <tests>
        <test id="test-1" phase_id="phase-1" task_id="task-1" status="planning">
            <description>Test for start command</description>
        </test>
        <test id="test-2" phase_id="phase-1" task_id="task-1" status="wip">
            <description>Test for pass command</description>
        </test>
        <test id="test-3" phase_id="phase-1" task_id="task-1" status="wip">
            <description>Test for fail command</description>
        </test>
        <test id="test-4" phase_id="phase-1" task_id="task-1" status="wip">
            <description>Test for cancel command</description>
        </test>
    </tests>
    <events />
</epic>`

	err := os.WriteFile(epicFile, []byte(epicContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test epic file: %v", err)
	}

	testCases := []struct {
		name           string
		args           []string
		expectSuccess  bool
		expectContains string
	}{
		{
			name:           "start-test success",
			args:           []string{"agentpm", "start-test", "--file", epicFile, "test-1"},
			expectSuccess:  true,
			expectContains: "Test test-1 started",
		},
		{
			name:           "pass-test success",
			args:           []string{"agentpm", "pass-test", "--file", epicFile, "test-2"},
			expectSuccess:  true,
			expectContains: "Test test-2 passed",
		},
		{
			name:           "fail-test success",
			args:           []string{"agentpm", "fail-test", "--file", epicFile, "test-3", "Test failure reason"},
			expectSuccess:  true,
			expectContains: "Test test-3 failed: Test failure reason",
		},
		{
			name:           "cancel-test success",
			args:           []string{"agentpm", "cancel-test", "--file", epicFile, "test-4", "Test cancellation reason"},
			expectSuccess:  true,
			expectContains: "Test test-4 cancelled: Test cancellation reason",
		},
		{
			name:           "start-test missing argument",
			args:           []string{"agentpm", "start-test", "--file", epicFile},
			expectSuccess:  false,
			expectContains: "exactly one argument",
		},
		{
			name:           "fail-test missing reason",
			args:           []string{"agentpm", "fail-test", "--file", epicFile, "test-1"},
			expectSuccess:  false,
			expectContains: "exactly two arguments",
		},
		{
			name:           "start-test nonexistent test",
			args:           []string{"agentpm", "start-test", "--file", epicFile, "nonexistent"},
			expectSuccess:  false,
			expectContains: "not found",
		},
		{
			name:           "start-test JSON format",
			args:           []string{"agentpm", "start-test", "--file", epicFile, "--format", "json", "test-1"},
			expectSuccess:  true, // Will succeed with already_started status
			expectContains: `"test_id"`,
		},
		{
			name:           "start-test XML format",
			args:           []string{"agentpm", "start-test", "--file", epicFile, "--format", "xml", "test-1"},
			expectSuccess:  true, // Will succeed with already_started status
			expectContains: "<test_id>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create CLI app
			app := createBasicTestApp()
			var stdout, stderr bytes.Buffer
			app.Writer = &stdout
			app.ErrWriter = &stderr

			// Run command
			err := app.Run(context.Background(), tc.args)

			// Check success/failure expectation
			if tc.expectSuccess && err != nil {
				t.Errorf("Expected command to succeed, but got error: %v", err)
			}
			if !tc.expectSuccess && err == nil {
				t.Errorf("Expected command to fail, but it succeeded")
			}

			// Check output contains expected text
			output := stdout.String() + stderr.String()
			if err != nil {
				output += err.Error()
			}
			if !strings.Contains(output, tc.expectContains) {
				t.Errorf("Expected output to contain '%s', got: %s", tc.expectContains, output)
			}
		})
	}
}

// TestCommandsCustomTimestamp tests custom timestamp functionality
func TestCommandsCustomTimestamp(t *testing.T) {
	// Create test epic file
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	epicContent := `<?xml version="1.0" encoding="UTF-8"?>
<epic id="test-epic" name="Test Epic" status="wip" created_at="2025-08-16T14:00:00Z">
    <phases>
        <phase id="phase-1" name="Phase 1" status="wip" />
    </phases>
    <tasks>
        <task id="task-1" phase_id="phase-1" name="Task 1" status="wip" />
    </tasks>
    <tests>
        <test id="test-1" phase_id="phase-1" task_id="task-1" status="planning">
            <description>Test for custom timestamp</description>
        </test>
    </tests>
    <events />
</epic>`

	err := os.WriteFile(epicFile, []byte(epicContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test epic file: %v", err)
	}

	// Create CLI app
	app := createBasicTestApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr

	// Test with custom timestamp
	args := []string{"agentpm", "start-test", "--file", epicFile, "--time", "2025-12-25T10:30:00Z", "test-1"}
	err = app.Run(context.Background(), args)

	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Should succeed
	output := stdout.String()
	if !strings.Contains(output, "Test test-1 started") {
		t.Errorf("Expected success message, got: %s", output)
	}
}

// TestCommandsInvalidTimestamp tests error handling for invalid timestamps
func TestCommandsInvalidTimestamp(t *testing.T) {
	// Create test epic file
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	epicContent := `<?xml version="1.0" encoding="UTF-8"?>
<epic id="test-epic" name="Test Epic" status="wip" created_at="2025-08-16T14:00:00Z">
    <phases>
        <phase id="phase-1" name="Phase 1" status="wip" />
    </phases>
    <tasks>
        <task id="task-1" phase_id="phase-1" name="Task 1" status="wip" />
    </tasks>
    <tests>
        <test id="test-1" phase_id="phase-1" task_id="task-1" status="planning">
            <description>Test for invalid timestamp</description>
        </test>
    </tests>
    <events />
</epic>`

	err := os.WriteFile(epicFile, []byte(epicContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test epic file: %v", err)
	}

	// Create CLI app
	app := createBasicTestApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr

	// Test with invalid timestamp
	args := []string{"agentpm", "start-test", "--file", epicFile, "--time", "invalid-timestamp", "test-1"}
	err = app.Run(context.Background(), args)

	// Should fail
	if err == nil {
		t.Fatal("Expected command to fail with invalid timestamp, but it succeeded")
	}

	if !strings.Contains(err.Error(), "invalid time format") {
		t.Errorf("Expected invalid time format error, got: %v", err)
	}
}

// TestCommandsNonExistentFile tests error handling for non-existent epic files
func TestCommandsNonExistentFile(t *testing.T) {
	// Create CLI app
	app := createBasicTestApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr

	// Test with non-existent file
	args := []string{"agentpm", "start-test", "--file", "/nonexistent/epic.xml", "test-1"}
	err := app.Run(context.Background(), args)

	// Should fail
	if err == nil {
		t.Fatal("Expected command to fail with non-existent file, but it succeeded")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected file not found error, got: %v", err)
	}
}

// Helper function to create a basic test app with commands
func createBasicTestApp() *cli.Command {
	return &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "file",
				Usage: "Override epic file from config",
			},
			&cli.StringFlag{
				Name:  "config",
				Usage: "Override config file path",
				Value: "./.agentpm.json",
			},
			&cli.StringFlag{
				Name:  "time",
				Usage: "Timestamp for current time (testing support)",
			},
			&cli.StringFlag{
				Name:  "format",
				Usage: "Output format - text (default) / json / xml",
				Value: "text",
			},
		},
		Commands: []*cli.Command{
			StartTestCommand(),
			PassTestCommand(),
			FailTestCommand(),
			CancelTestCommand(),
		},
	}
}
