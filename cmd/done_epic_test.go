package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/memomoo/agentpm/internal/config"
	"github.com/memomoo/agentpm/internal/epic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestDoneEpicCommand_Success(t *testing.T) {
	// Create temporary directory and files
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".agentpm.json")
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	// Create test config
	cfg := &config.Config{
		CurrentEpic: epicFile, // Use absolute path
	}
	require.NoError(t, config.SaveConfig(cfg, configFile))

	// Create test epic in wip status with all phases/tasks/tests completed
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusActive, // wip state
		Phases: []epic.Phase{
			{ID: "phase-1", Name: "Phase 1", Status: epic.StatusCompleted},
		},
		Tasks: []epic.Task{
			{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
		},
		Tests: []epic.Test{
			{ID: "test-1", TaskID: "task-1", Name: "Test 1", Status: epic.StatusCompleted},
		},
	}
	writeTestEpicXML(t, epicFile, testEpic)

	// Create CLI app with done-epic command
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Value: configFile},
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			DoneEpicCommand(),
		},
	}

	// Capture output
	var stdout bytes.Buffer
	app.Writer = &stdout

	// Run done-epic command
	args := []string{"agentpm", "done-epic"}
	err := app.Run(context.Background(), args)

	// Verify success
	require.NoError(t, err)

	// Check output contains expected content
	output := stdout.String()
	assert.Contains(t, output, "Epic epic-1 completed successfully")
	assert.Contains(t, output, "Status: wip → done")

	// Verify epic was updated on disk
	updatedEpic := readTestEpicXML(t, epicFile)
	assert.Equal(t, epic.StatusCompleted, updatedEpic.Status)
}

func TestDoneEpicCommand_WithFileFlag(t *testing.T) {
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "custom-epic.xml")

	// Create test epic in wip status with all completed
	testEpic := &epic.Epic{
		ID:     "epic-2",
		Name:   "Custom Epic",
		Status: epic.StatusActive,
		Phases: []epic.Phase{
			{ID: "phase-1", Name: "Phase 1", Status: epic.StatusCompleted},
		},
		Tasks: []epic.Task{
			{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
		},
		Tests: []epic.Test{
			{ID: "test-1", TaskID: "task-1", Name: "Test 1", Status: epic.StatusCompleted},
		},
	}
	writeTestEpicXML(t, epicFile, testEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			DoneEpicCommand(),
		},
	}

	var stdout bytes.Buffer
	app.Writer = &stdout

	// Run with file flag
	args := []string{"agentpm", "done-epic", "--file", epicFile}
	err := app.Run(context.Background(), args)

	require.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "Epic epic-2 completed successfully")
}

func TestDoneEpicCommand_WithTimestamp(t *testing.T) {
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	// Create test epic in wip status with all completed
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusActive,
		Phases: []epic.Phase{
			{ID: "phase-1", Name: "Phase 1", Status: epic.StatusCompleted},
		},
		Tasks: []epic.Task{
			{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
		},
		Tests: []epic.Test{
			{ID: "test-1", TaskID: "task-1", Name: "Test 1", Status: epic.StatusCompleted},
		},
	}
	writeTestEpicXML(t, epicFile, testEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			DoneEpicCommand(),
		},
	}

	var stdout bytes.Buffer
	app.Writer = &stdout

	// Run with specific timestamp
	testTime := "2025-08-16T15:30:00Z"
	args := []string{"agentpm", "done-epic", "--file", epicFile, "--time", testTime}
	err := app.Run(context.Background(), args)

	require.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "Completed at: "+testTime)
}

func TestDoneEpicCommand_JSONOutput(t *testing.T) {
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	// Create test epic in wip status with all completed
	testEpic := &epic.Epic{
		ID:     "epic-json",
		Name:   "JSON Test Epic",
		Status: epic.StatusActive,
		Phases: []epic.Phase{
			{ID: "phase-1", Name: "Phase 1", Status: epic.StatusCompleted},
		},
		Tasks: []epic.Task{
			{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
		},
		Tests: []epic.Test{
			{ID: "test-1", TaskID: "task-1", Name: "Test 1", Status: epic.StatusCompleted},
		},
	}
	writeTestEpicXML(t, epicFile, testEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "json"},
		},
		Commands: []*cli.Command{
			DoneEpicCommand(),
		},
	}

	var stdout bytes.Buffer
	app.Writer = &stdout

	// Run command
	args := []string{"agentpm", "done-epic", "--file", epicFile}
	err := app.Run(context.Background(), args)

	require.NoError(t, err)

	// Parse JSON output
	var output map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &output)
	require.NoError(t, err)

	// Verify JSON structure
	assert.Contains(t, output, "epic_completed")
	completed := output["epic_completed"].(map[string]interface{})
	assert.Equal(t, "epic-json", completed["epic_id"])
	assert.Equal(t, "wip", completed["previous_status"])
	assert.Equal(t, "done", completed["new_status"])
	assert.Equal(t, false, completed["event_created"])
}

func TestDoneEpicCommand_XMLOutput(t *testing.T) {
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	// Create test epic in wip status with all completed
	testEpic := &epic.Epic{
		ID:     "epic-xml",
		Name:   "XML Test Epic",
		Status: epic.StatusActive,
		Phases: []epic.Phase{
			{ID: "phase-1", Name: "Phase 1", Status: epic.StatusCompleted},
		},
		Tasks: []epic.Task{
			{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
		},
		Tests: []epic.Test{
			{ID: "test-1", TaskID: "task-1", Name: "Test 1", Status: epic.StatusCompleted},
		},
	}
	writeTestEpicXML(t, epicFile, testEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "xml"},
		},
		Commands: []*cli.Command{
			DoneEpicCommand(),
		},
	}

	var stdout bytes.Buffer
	app.Writer = &stdout

	// Run command
	args := []string{"agentpm", "done-epic", "--file", epicFile}
	err := app.Run(context.Background(), args)

	require.NoError(t, err)

	// Check XML output contains expected elements
	output := stdout.String()
	assert.Contains(t, output, `<epic_completed epic="epic-xml">`)
	assert.Contains(t, output, "<previous_status>wip</previous_status>")
	assert.Contains(t, output, "<new_status>done</new_status>")
	assert.Contains(t, output, "<event_created>false</event_created>")
}

func TestDoneEpicCommand_ErrorWrongStatus(t *testing.T) {
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	// Create test epic in pending status (cannot complete from pending)
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusPlanning, // pending state
	}
	writeTestEpicXML(t, epicFile, testEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			DoneEpicCommand(),
		},
	}

	var stderr bytes.Buffer
	app.ErrWriter = &stderr

	// Run command
	args := []string{"agentpm", "done-epic", "--file", epicFile}
	err := app.Run(context.Background(), args)

	// Should return transition error
	assert.Error(t, err)
	stderrOutput := stderr.String()
	assert.Contains(t, stderrOutput, "Epic cannot be completed from status: pending")
	assert.Contains(t, stderrOutput, "Epic must be started first")
}

func TestDoneEpicCommand_ErrorPendingPhases(t *testing.T) {
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	// Create test epic in wip status but with pending phases
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusActive, // wip state
		Phases: []epic.Phase{
			{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPlanning}, // pending!
		},
		Tasks: []epic.Task{
			{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
		},
		Tests: []epic.Test{
			{ID: "test-1", TaskID: "task-1", Name: "Test 1", Status: epic.StatusCompleted},
		},
	}
	writeTestEpicXML(t, epicFile, testEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			DoneEpicCommand(),
		},
	}

	var stderr bytes.Buffer
	app.ErrWriter = &stderr

	// Run command
	args := []string{"agentpm", "done-epic", "--file", epicFile}
	err := app.Run(context.Background(), args)

	// Should return validation error
	assert.Error(t, err)
	stderrOutput := stderr.String()
	assert.Contains(t, stderrOutput, "cannot be completed: 1 pending phases")
}

func TestDoneEpicCommand_ErrorFailingTests(t *testing.T) {
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	// Create test epic in wip status but with failing tests
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusActive, // wip state
		Phases: []epic.Phase{
			{ID: "phase-1", Name: "Phase 1", Status: epic.StatusCompleted},
		},
		Tasks: []epic.Task{
			{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
		},
		Tests: []epic.Test{
			{ID: "test-1", TaskID: "task-1", Name: "Test 1", Status: epic.StatusPlanning}, // failing!
		},
	}
	writeTestEpicXML(t, epicFile, testEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			DoneEpicCommand(),
		},
	}

	var stderr bytes.Buffer
	app.ErrWriter = &stderr

	// Run command
	args := []string{"agentpm", "done-epic", "--file", epicFile}
	err := app.Run(context.Background(), args)

	// Should return validation error
	assert.Error(t, err)
	stderrOutput := stderr.String()
	assert.Contains(t, stderrOutput, "cannot be completed: 1 failing tests")
}

func TestDoneEpicCommand_ErrorFileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	nonExistentFile := filepath.Join(tempDir, "non-existent.xml")

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			DoneEpicCommand(),
		},
	}

	var stderr bytes.Buffer
	app.ErrWriter = &stderr

	// Run command with non-existent file
	args := []string{"agentpm", "done-epic", "--file", nonExistentFile}
	err := app.Run(context.Background(), args)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load epic")
}
