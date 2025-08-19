package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/epic"
	apmtesting "github.com/mindreframer/agentpm/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestStartEpicCommand_Success(t *testing.T) {
	// Create temporary directory and files
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".agentpm.json")
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	// Create test config
	cfg := &config.Config{
		CurrentEpic: epicFile, // Use absolute path
	}
	require.NoError(t, config.SaveConfig(cfg, configFile))

	// Create test epic in pending status
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusPending, // pending state
		Phases: []epic.Phase{
			{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPending},
		},
		Tasks: []epic.Task{
			{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPending},
		},
		Tests: []epic.Test{
			{ID: "test-1", TaskID: "task-1", Name: "Test 1", Status: epic.StatusPending},
		},
	}
	writeTestEpicXML(t, epicFile, testEpic)

	// Create CLI app with start-epic command
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Value: configFile},
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			StartEpicCommand(),
		},
	}

	// Capture output
	var stdout bytes.Buffer
	app.Writer = &stdout

	// Run start-epic command
	args := []string{"agentpm", "start-epic"}
	err := app.Run(context.Background(), args)

	// Verify success
	require.NoError(t, err)

	// Check output contains expected content
	output := stdout.String()
	assert.Contains(t, output, "Epic epic-1 started successfully")
	assert.Contains(t, output, "Status: pending → wip")

	// Verify epic was updated on disk
	updatedEpic := readTestEpicXML(t, epicFile)
	assert.Equal(t, epic.StatusActive, updatedEpic.Status)
}

func TestStartEpicCommand_WithFileFlag(t *testing.T) {
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "custom-epic.xml")

	// Create test epic
	testEpic := &epic.Epic{
		ID:     "epic-2",
		Name:   "Custom Epic",
		Status: epic.StatusPending,
	}
	writeTestEpicXML(t, epicFile, testEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			StartEpicCommand(),
		},
	}

	var stdout bytes.Buffer
	app.Writer = &stdout

	// Run with file flag
	args := []string{"agentpm", "start-epic", "--file", epicFile}
	err := app.Run(context.Background(), args)

	require.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "Epic epic-2 started successfully")
}

func TestStartEpicCommand_WithTimestamp(t *testing.T) {
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	// Create test epic
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusPending,
	}
	writeTestEpicXML(t, epicFile, testEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			StartEpicCommand(),
		},
	}

	var stdout bytes.Buffer
	app.Writer = &stdout

	// Run with specific timestamp
	testTime := "2025-08-16T15:30:00Z"
	args := []string{"agentpm", "start-epic", "--file", epicFile, "--time", testTime}
	err := app.Run(context.Background(), args)

	require.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "Started at: "+testTime)
}

func TestStartEpicCommand_AlreadyStarted(t *testing.T) {
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	// Create test epic already in active state
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusActive, // already started
	}
	writeTestEpicXML(t, epicFile, testEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			StartEpicCommand(),
		},
	}

	var stderr bytes.Buffer
	var stdout bytes.Buffer
	app.ErrWriter = &stderr
	app.Writer = &stdout

	// Run start-epic command
	args := []string{"agentpm", "start-epic", "--file", epicFile}
	err := app.Run(context.Background(), args)

	// Should succeed with friendly message
	assert.NoError(t, err)
	output := stdout.String() // Friendly message goes to stdout, not stderr
	assert.Contains(t, output, "already started")
	assert.Contains(t, output, "No action needed")
}

func TestStartEpicCommand_JSONOutput(t *testing.T) {
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	// Create test epic
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusPending,
	}
	writeTestEpicXML(t, epicFile, testEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "json"},
		},
		Commands: []*cli.Command{
			StartEpicCommand(),
		},
	}

	var stdout bytes.Buffer
	app.Writer = &stdout

	// Run start-epic command
	args := []string{"agentpm", "start-epic", "--file", epicFile}
	err := app.Run(context.Background(), args)

	require.NoError(t, err)

	// Parse JSON output
	var result map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &result)
	require.NoError(t, err)

	// Verify JSON structure
	epicStarted, ok := result["epic_started"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "epic-1", epicStarted["epic_id"])
	assert.Equal(t, "pending", epicStarted["previous_status"])
	assert.Equal(t, "wip", epicStarted["new_status"])
	assert.True(t, epicStarted["event_created"].(bool))
}

func TestStartEpicCommand_XMLOutput(t *testing.T) {
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	// Create test epic
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusPending,
	}
	writeTestEpicXML(t, epicFile, testEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "xml"},
		},
		Commands: []*cli.Command{
			StartEpicCommand(),
		},
	}

	var stdout bytes.Buffer
	app.Writer = &stdout

	// Run start-epic command
	args := []string{"agentpm", "start-epic", "--file", epicFile}
	err := app.Run(context.Background(), args)

	require.NoError(t, err)

	// Verify XML output structure using snapshots
	output := stdout.String()
	snapshotTester := apmtesting.NewSnapshotTester()
	snapshotTester.MatchXMLSnapshot(t, output, "start_epic_xml_output")
}

func TestStartEpicCommand_FriendlyOutput_JSON(t *testing.T) {
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	// Create test epic already completed
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusCompleted, // already completed - should return friendly message
	}
	writeTestEpicXML(t, epicFile, testEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "json"},
		},
		Commands: []*cli.Command{
			StartEpicCommand(),
		},
	}

	var stdout bytes.Buffer
	app.Writer = &stdout

	// Run start-epic command
	args := []string{"agentpm", "start-epic", "--file", epicFile}
	err := app.Run(context.Background(), args)

	assert.NoError(t, err) // Should succeed with friendly message

	// Parse JSON friendly output
	var result map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &result)
	require.NoError(t, err)

	// Verify JSON friendly message structure
	assert.Equal(t, "success", result["type"])
	content, ok := result["content"].(string)
	require.True(t, ok)
	assert.Contains(t, content, "epic-1")
	assert.Contains(t, content, "already completed")
	assert.Contains(t, content, "No action needed")

	// Verify hint is present
	hint, ok := result["hint"].(string)
	require.True(t, ok)
	assert.Contains(t, hint, "agentpm status")
}

func TestStartEpicCommand_NoEpicFile(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".agentpm.json")
	// Note: We don't create the config file, so LoadConfig should fail

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Value: configFile},
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			StartEpicCommand(),
		},
	}

	var stderr bytes.Buffer
	app.ErrWriter = &stderr

	// Run start-epic command (no --file flag, no config)
	args := []string{"agentpm", "start-epic"}
	err := app.Run(context.Background(), args)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load configuration")
}

func TestStartEpicCommand_InvalidTimestamp(t *testing.T) {
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	// Create test epic
	testEpic := &epic.Epic{
		ID:     "epic-1",
		Name:   "Test Epic",
		Status: epic.StatusPending,
	}
	writeTestEpicXML(t, epicFile, testEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			StartEpicCommand(),
		},
	}

	// Run with invalid timestamp
	args := []string{"agentpm", "start-epic", "--file", epicFile, "--time", "invalid-time"}
	err := app.Run(context.Background(), args)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid time format")
	assert.Contains(t, err.Error(), "use ISO 8601 format")
}

func TestStartEpicCommand_NonExistentFile(t *testing.T) {
	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			StartEpicCommand(),
		},
	}

	var stderr bytes.Buffer
	app.ErrWriter = &stderr

	// Run with non-existent file
	args := []string{"agentpm", "start-epic", "--file", "/non/existent/file.xml"}
	err := app.Run(context.Background(), args)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load epic")
}

// Helper functions for testing

func writeTestEpicXML(t *testing.T, filePath string, epic *epic.Epic) {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	require.NoError(t, os.MkdirAll(dir, 0755))

	// Write epic as XML
	file, err := os.Create(filePath)
	require.NoError(t, err)
	defer file.Close()

	// Simple XML writing for test purposes
	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<epic id="%s" name="%s" status="%s" created_at="2025-08-15T09:00:00Z">
    <description>Test epic description</description>
    <phases>
`, epic.ID, epic.Name, epic.Status)

	for _, phase := range epic.Phases {
		content += fmt.Sprintf(`        <phase id="%s" name="%s" status="%s">
            <description>%s</description>
        </phase>
`, phase.ID, phase.Name, phase.Status, phase.Description)
	}

	content += `    </phases>
    <tasks>
`

	for _, task := range epic.Tasks {
		content += fmt.Sprintf(`        <task id="%s" phase_id="%s" name="%s" status="%s">
            <description>%s</description>
        </task>
`, task.ID, task.PhaseID, task.Name, task.Status, task.Description)
	}

	content += `    </tasks>
    <tests>
`

	for _, test := range epic.Tests {
		content += fmt.Sprintf(`        <test id="%s" task_id="%s" name="%s" status="%s">
            <description>%s</description>
        </test>
`, test.ID, test.TaskID, test.Name, test.Status, test.Description)
	}

	content += `    </tests>
    <events>
    </events>
</epic>`

	_, err = file.WriteString(content)
	require.NoError(t, err)
}

func readTestEpicXML(t *testing.T, filePath string) *epic.Epic {
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)

	var epicData epic.Epic
	err = xml.Unmarshal(content, &epicData)
	require.NoError(t, err)

	return &epicData
}
