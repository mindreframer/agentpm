package cmd

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/urfave/cli/v3"
)

// TestStartTestCommand_Success tests successful start-test execution
func TestStartTestCommand_Success(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	// Create test epic with pending test
	testEpic := createTestEpicForCLI()
	storage := storage.NewFileStorage()
	err := storage.SaveEpic(testEpic, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Create CLI app and command
	app := createTestApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr

	// Test start-test command
	args := []string{"agentpm", "start-test", "--file", epicFile, "test-1"}
	err = app.Run(context.Background(), args)

	// Verify
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Test test-1 started.") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify epic was updated
	updatedEpic, err := storage.LoadEpic(epicFile)
	if err != nil {
		t.Fatalf("Failed to load updated epic: %v", err)
	}

	var updatedTest *epic.Test
	for i := range updatedEpic.Tests {
		if updatedEpic.Tests[i].ID == "test-1" {
			updatedTest = &updatedEpic.Tests[i]
			break
		}
	}

	if updatedTest == nil {
		t.Fatal("Test not found in updated epic")
	}

	if updatedTest.TestStatus != epic.TestStatusWIP {
		t.Errorf("Expected TestStatus %s, got %s", epic.TestStatusWIP, updatedTest.TestStatus)
	}

	if updatedTest.StartedAt == nil {
		t.Error("Expected StartedAt to be set")
	}
}

// TestStartTestCommand_JSONOutput tests JSON output format
func TestStartTestCommand_JSONOutput(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	testEpic := createTestEpicForCLI()
	storage := storage.NewFileStorage()
	err := storage.SaveEpic(testEpic, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Create CLI app
	app := createTestApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr

	// Test with JSON format
	args := []string{"agentpm", "start-test", "--file", epicFile, "--format", "json", "test-1"}
	err = app.Run(context.Background(), args)

	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, `"test_id": "test-1"`) {
		t.Errorf("Expected JSON output with test_id, got: %s", output)
	}
	if !strings.Contains(output, `"operation": "started"`) {
		t.Errorf("Expected JSON output with operation, got: %s", output)
	}
	if !strings.Contains(output, `"status": "wip"`) {
		t.Errorf("Expected JSON output with status, got: %s", output)
	}
}

// TestStartTestCommand_XMLOutput tests XML output format
func TestStartTestCommand_XMLOutput(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	testEpic := createTestEpicForCLI()
	storage := storage.NewFileStorage()
	err := storage.SaveEpic(testEpic, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Create CLI app
	app := createTestApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr

	// Test with XML format
	args := []string{"agentpm", "start-test", "--file", epicFile, "--format", "xml", "test-1"}
	err = app.Run(context.Background(), args)

	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "<test_operation>") {
		t.Errorf("Expected XML output with test_operation element, got: %s", output)
	}
	if !strings.Contains(output, "<test_id>test-1</test_id>") {
		t.Errorf("Expected XML output with test_id, got: %s", output)
	}
	if !strings.Contains(output, "<operation>started</operation>") {
		t.Errorf("Expected XML output with operation, got: %s", output)
	}
}

// TestStartTestCommand_InvalidTransition tests error when test is not in pending status
func TestStartTestCommand_AlreadyStarted(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	testEpic := createTestEpicForCLI()
	// Set test to WIP status (already started)
	testEpic.Tests[0].TestStatus = epic.TestStatusWIP
	testEpic.Tests[0].Status = epic.StatusActive

	storage := storage.NewFileStorage()
	err := storage.SaveEpic(testEpic, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Create CLI app
	app := createTestApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr

	// Test start-test command (should succeed with friendly message)
	args := []string{"agentpm", "start-test", "--file", epicFile, "test-1"}
	err = app.Run(context.Background(), args)

	// Should return success
	if err != nil {
		t.Fatalf("Expected command to succeed, but it failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "already started") || !strings.Contains(output, "test-1") {
		t.Errorf("Expected friendly already started message, got: %s", output)
	}
}

// TestStartTestCommand_MissingArgument tests error when test-id argument is missing
func TestStartTestCommand_MissingArgument(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	testEpic := createTestEpicForCLI()
	storage := storage.NewFileStorage()
	err := storage.SaveEpic(testEpic, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Create CLI app
	app := createTestApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr

	// Test start-test command without test-id
	args := []string{"agentpm", "start-test", "--file", epicFile}
	err = app.Run(context.Background(), args)

	// Should return error
	if err == nil {
		t.Fatal("Expected command to fail, but it succeeded")
	}

	if !strings.Contains(err.Error(), "requires exactly one argument") {
		t.Errorf("Expected missing argument error, got: %v", err)
	}
}

// TestPassTestCommand_Success tests successful pass-test execution
func TestPassTestCommand_Success(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	testEpic := createTestEpicForCLI()
	// Set test to WIP status (can pass)
	startTime := time.Date(2025, 8, 16, 14, 30, 0, 0, time.UTC)
	testEpic.Tests[0].TestStatus = epic.TestStatusWIP
	testEpic.Tests[0].Status = epic.StatusActive
	testEpic.Tests[0].StartedAt = &startTime

	storage := storage.NewFileStorage()
	err := storage.SaveEpic(testEpic, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Create CLI app
	app := createTestApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr

	// Test pass-test command
	args := []string{"agentpm", "pass-test", "--file", epicFile, "test-1"}
	err = app.Run(context.Background(), args)

	// Verify
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Test test-1 passed.") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify epic was updated
	updatedEpic, err := storage.LoadEpic(epicFile)
	if err != nil {
		t.Fatalf("Failed to load updated epic: %v", err)
	}

	var updatedTest *epic.Test
	for i := range updatedEpic.Tests {
		if updatedEpic.Tests[i].ID == "test-1" {
			updatedTest = &updatedEpic.Tests[i]
			break
		}
	}

	if updatedTest == nil {
		t.Fatal("Test not found in updated epic")
	}

	if updatedTest.TestStatus != epic.TestStatusDone {
		t.Errorf("Expected TestStatus %s, got %s", epic.TestStatusDone, updatedTest.TestStatus)
	}

	if updatedTest.PassedAt == nil {
		t.Error("Expected PassedAt to be set")
	}

	if updatedTest.FailureNote != "" {
		t.Error("Expected FailureNote to be cleared")
	}
}

// TestFailTestCommand_Success tests successful fail-test execution
func TestFailTestCommand_Success(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	testEpic := createTestEpicForCLI()
	// Set test to WIP status (can fail)
	startTime := time.Date(2025, 8, 16, 14, 30, 0, 0, time.UTC)
	testEpic.Tests[0].TestStatus = epic.TestStatusWIP
	testEpic.Tests[0].Status = epic.StatusActive
	testEpic.Tests[0].StartedAt = &startTime

	storage := storage.NewFileStorage()
	err := storage.SaveEpic(testEpic, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Create CLI app
	app := createTestApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr

	// Test fail-test command
	failureReason := "Mobile responsive design not working"
	args := []string{"agentpm", "fail-test", "--file", epicFile, "test-1", failureReason}
	err = app.Run(context.Background(), args)

	// Verify
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := stdout.String()
	expectedOutput := "Test test-1 failed: " + failureReason
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output '%s', got: %s", expectedOutput, output)
	}

	// Verify epic was updated
	updatedEpic, err := storage.LoadEpic(epicFile)
	if err != nil {
		t.Fatalf("Failed to load updated epic: %v", err)
	}

	var updatedTest *epic.Test
	for i := range updatedEpic.Tests {
		if updatedEpic.Tests[i].ID == "test-1" {
			updatedTest = &updatedEpic.Tests[i]
			break
		}
	}

	if updatedTest == nil {
		t.Fatal("Test not found in updated epic")
	}

	if updatedTest.TestStatus != epic.TestStatusWIP {
		t.Errorf("Expected TestStatus %s, got %s", epic.TestStatusWIP, updatedTest.TestStatus)
	}

	if updatedTest.FailedAt == nil {
		t.Error("Expected FailedAt to be set")
	}

	if updatedTest.FailureNote != failureReason {
		t.Errorf("Expected FailureNote '%s', got '%s'", failureReason, updatedTest.FailureNote)
	}
}

// TestFailTestCommand_MissingReason tests error when failure reason is missing
func TestFailTestCommand_MissingReason(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	testEpic := createTestEpicForCLI()
	storage := storage.NewFileStorage()
	err := storage.SaveEpic(testEpic, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Create CLI app
	app := createTestApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr

	// Test fail-test command without reason
	args := []string{"agentpm", "fail-test", "--file", epicFile, "test-1"}
	err = app.Run(context.Background(), args)

	// Should return error
	if err == nil {
		t.Fatal("Expected command to fail, but it succeeded")
	}

	if !strings.Contains(err.Error(), "requires exactly two arguments") {
		t.Errorf("Expected missing argument error, got: %v", err)
	}
}

// TestCancelTestCommand_Success tests successful cancel-test execution
func TestCancelTestCommand_Success(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	testEpic := createTestEpicForCLI()
	// Set test to WIP status (can cancel)
	startTime := time.Date(2025, 8, 16, 14, 30, 0, 0, time.UTC)
	testEpic.Tests[0].TestStatus = epic.TestStatusWIP
	testEpic.Tests[0].Status = epic.StatusActive
	testEpic.Tests[0].StartedAt = &startTime

	storage := storage.NewFileStorage()
	err := storage.SaveEpic(testEpic, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Create CLI app
	app := createTestApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr

	// Test cancel-test command
	cancellationReason := "Spec contradicts itself with point xyz"
	args := []string{"agentpm", "cancel-test", "--file", epicFile, "test-1", cancellationReason}
	err = app.Run(context.Background(), args)

	// Verify
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	output := stdout.String()
	expectedOutput := "Test test-1 cancelled: " + cancellationReason
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output '%s', got: %s", expectedOutput, output)
	}

	// Verify epic was updated
	updatedEpic, err := storage.LoadEpic(epicFile)
	if err != nil {
		t.Fatalf("Failed to load updated epic: %v", err)
	}

	var updatedTest *epic.Test
	for i := range updatedEpic.Tests {
		if updatedEpic.Tests[i].ID == "test-1" {
			updatedTest = &updatedEpic.Tests[i]
			break
		}
	}

	if updatedTest == nil {
		t.Fatal("Test not found in updated epic")
	}

	if updatedTest.TestStatus != epic.TestStatusCancelled {
		t.Errorf("Expected TestStatus %s, got %s", epic.TestStatusCancelled, updatedTest.TestStatus)
	}

	if updatedTest.CancelledAt == nil {
		t.Error("Expected CancelledAt to be set")
	}

	if updatedTest.CancellationReason != cancellationReason {
		t.Errorf("Expected CancellationReason '%s', got '%s'", cancellationReason, updatedTest.CancellationReason)
	}
}

// TestCustomTimestamp tests that custom timestamps are properly used
func TestCustomTimestamp(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	testEpic := createTestEpicForCLI()
	storage := storage.NewFileStorage()
	err := storage.SaveEpic(testEpic, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Create CLI app
	app := createTestApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr

	// Test with custom timestamp
	customTime := "2025-12-25T10:30:00Z"
	args := []string{"agentpm", "start-test", "--file", epicFile, "--time", customTime, "test-1"}
	err = app.Run(context.Background(), args)

	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Verify custom timestamp was used
	updatedEpic, err := storage.LoadEpic(epicFile)
	if err != nil {
		t.Fatalf("Failed to load updated epic: %v", err)
	}

	var updatedTest *epic.Test
	for i := range updatedEpic.Tests {
		if updatedEpic.Tests[i].ID == "test-1" {
			updatedTest = &updatedEpic.Tests[i]
			break
		}
	}

	if updatedTest == nil {
		t.Fatal("Test not found in updated epic")
	}

	if updatedTest.StartedAt == nil {
		t.Fatal("Expected StartedAt to be set")
	}

	expectedTime, _ := time.Parse(time.RFC3339, customTime)
	if !updatedTest.StartedAt.Equal(expectedTime) {
		t.Errorf("Expected custom timestamp %v, got %v", expectedTime, *updatedTest.StartedAt)
	}
}

// TestInvalidTimestamp tests error handling for invalid timestamp format
func TestInvalidTimestamp(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	testEpic := createTestEpicForCLI()
	storage := storage.NewFileStorage()
	err := storage.SaveEpic(testEpic, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Create CLI app
	app := createTestApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr

	// Test with invalid timestamp
	args := []string{"agentpm", "start-test", "--file", epicFile, "--time", "invalid-time", "test-1"}
	err = app.Run(context.Background(), args)

	// Should return error
	if err == nil {
		t.Fatal("Expected command to fail with invalid timestamp, but it succeeded")
	}

	if !strings.Contains(err.Error(), "invalid time format") {
		t.Errorf("Expected invalid time format error, got: %v", err)
	}
}

// TestNonExistentTest tests error when test doesn't exist
func TestNonExistentTest(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	testEpic := createTestEpicForCLI()
	storage := storage.NewFileStorage()
	err := storage.SaveEpic(testEpic, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	// Create CLI app
	app := createTestApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr

	// Test with non-existent test
	args := []string{"agentpm", "start-test", "--file", epicFile, "nonexistent-test"}
	err = app.Run(context.Background(), args)

	// Should return error
	if err == nil {
		t.Fatal("Expected command to fail with non-existent test, but it succeeded")
	}

	errorOutput := stderr.String()
	if !strings.Contains(errorOutput, "not found") {
		t.Errorf("Expected test not found error, got: %s", errorOutput)
	}
}

// TestNonExistentEpicFile tests error when epic file doesn't exist
func TestNonExistentEpicFile(t *testing.T) {
	// Create CLI app
	app := createTestApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr

	// Test with non-existent epic file
	args := []string{"agentpm", "start-test", "--file", "/nonexistent/epic.xml", "test-1"}
	err := app.Run(context.Background(), args)

	// Should return error
	if err == nil {
		t.Fatal("Expected command to fail with non-existent file, but it succeeded")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected file not found error, got: %v", err)
	}
}

// TestAllCommands_JSONErrorFormat tests JSON error output format for all commands
func TestAllCommands_JSONErrorFormat(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	testEpic := createTestEpicForCLI()
	// Keep test in pending status - we'll test with nonexistent test IDs to trigger failures

	storage := storage.NewFileStorage()
	err := storage.SaveEpic(testEpic, epicFile)
	if err != nil {
		t.Fatalf("Failed to save test epic: %v", err)
	}

	testCases := []struct {
		name string
		args []string
	}{
		{"start-test", []string{"agentpm", "start-test", "--file", epicFile, "--format", "json", "nonexistent"}},
		{"pass-test", []string{"agentpm", "pass-test", "--file", epicFile, "--format", "json", "nonexistent"}},
		{"fail-test", []string{"agentpm", "fail-test", "--file", epicFile, "--format", "json", "nonexistent", "reason"}},
		{"cancel-test", []string{"agentpm", "cancel-test", "--file", epicFile, "--format", "json", "nonexistent", "reason"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create CLI app
			app := createTestApp()
			var stdout, stderr bytes.Buffer
			app.Writer = &stdout
			app.ErrWriter = &stderr

			// Run command (should fail)
			err := app.Run(context.Background(), tc.args)
			if err == nil {
				t.Fatal("Expected command to fail, but it succeeded")
			}

			errorOutput := stderr.String()
			if !strings.Contains(errorOutput, `"error"`) {
				t.Errorf("Expected JSON error format, got: %s", errorOutput)
			}
		})
	}
}

// Helper functions

func createTestApp() *cli.Command {
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

func createTestEpicForCLI() *epic.Epic {
	return &epic.Epic{
		ID:     "test-epic",
		Name:   "Test Epic",
		Status: epic.StatusActive,
		Phases: []epic.Phase{
			{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
		},
		Tasks: []epic.Task{
			{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusActive},
		},
		Tests: []epic.Test{
			{
				ID:          "test-1",
				TaskID:      "task-1",
				PhaseID:     "phase-1",
				Name:        "Test 1",
				Status:      epic.StatusPlanning,
				TestStatus:  epic.TestStatusPending,
				Description: "Test for CLI commands",
			},
		},
		Events: []epic.Event{},
	}
}
