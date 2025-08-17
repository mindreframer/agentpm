package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestEpicForCurrent() *epic.Epic {
	return &epic.Epic{
		ID:        "current-test-epic",
		Name:      "Current Test Epic",
		Status:    epic.StatusActive,
		CreatedAt: time.Date(2025, 8, 16, 9, 0, 0, 0, time.UTC),
		Assignee:  "test_agent",
		Phases: []epic.Phase{
			{ID: "P1", Name: "Setup Phase", Status: epic.StatusCompleted},
			{ID: "P2", Name: "Implementation Phase", Status: epic.StatusActive},
			{ID: "P3", Name: "Testing Phase", Status: epic.StatusPlanning},
		},
		Tasks: []epic.Task{
			{ID: "T1", PhaseID: "P1", Name: "Setup Task", Status: epic.StatusCompleted},
			{ID: "T2", PhaseID: "P2", Name: "Active Task", Status: epic.StatusActive},
			{ID: "T3", PhaseID: "P2", Name: "Pending Task", Status: epic.StatusPlanning},
		},
		Tests: []epic.Test{
			{ID: "TEST1", TaskID: "T1", Name: "Setup Test", Status: epic.StatusCompleted},
			{ID: "TEST2", TaskID: "T2", Name: "Active Test", Status: epic.StatusPlanning}, // "failing"
		},
		Events: []epic.Event{},
	}
}

func createNoActiveWorkEpic() *epic.Epic {
	return &epic.Epic{
		ID:     "no-active-epic",
		Name:   "No Active Work Epic",
		Status: epic.StatusCompleted,
		Phases: []epic.Phase{
			{ID: "P1", Name: "Phase 1", Status: epic.StatusCompleted},
		},
		Tasks: []epic.Task{
			{ID: "T1", PhaseID: "P1", Name: "Task 1", Status: epic.StatusCompleted},
		},
		Tests: []epic.Test{
			{ID: "TEST1", TaskID: "T1", Name: "Test 1", Status: epic.StatusCompleted},
		},
		Events: []epic.Event{},
	}
}

func TestCurrentCommand(t *testing.T) {
	t.Run("current with active work - text format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForCurrent()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		// Create current command
		var stdout, stderr bytes.Buffer
		cmd := CurrentCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		err = cmd.Run(context.Background(), []string{"current"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Current Work State")
		assert.Contains(t, output, "Epic Status: active")
		assert.Contains(t, output, "Active Phase: P2")
		assert.Contains(t, output, "Active Task: T2")
		assert.Contains(t, output, "Failing Tests: 1")
		assert.Contains(t, output, "Next Action: Fix failing tests")
	})

	t.Run("current with no active work - text format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create completed epic and config
		testEpic := createNoActiveWorkEpic()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "completed-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		// Create current command
		var stdout, stderr bytes.Buffer
		cmd := CurrentCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		err = cmd.Run(context.Background(), []string{"current"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Epic Status: completed")
		assert.Contains(t, output, "Active Phase: none")
		assert.Contains(t, output, "Active Task: none")
		assert.Contains(t, output, "Failing Tests: 0")
		assert.Contains(t, output, "Next Action: Epic ready for completion")
	})

	t.Run("current with file override", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create two different epics
		activeEpic := createTestEpicForCurrent()
		completedEpic := createNoActiveWorkEpic()

		storage := storage.NewFileStorage()
		activePath := filepath.Join(tempDir, "active.xml")
		completedPath := filepath.Join(tempDir, "completed.xml")

		err := storage.SaveEpic(activeEpic, activePath)
		require.NoError(t, err)
		err = storage.SaveEpic(completedEpic, completedPath)
		require.NoError(t, err)

		// Config points to active epic
		cfg := &config.Config{
			CurrentEpic:     activePath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		// Execute current command with file override to completed epic
		var stdout, stderr bytes.Buffer
		cmd := CurrentCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"current", "--file", completedPath})
		require.NoError(t, err)

		output := stdout.String()
		// Should show completed epic data, not active epic data
		assert.Contains(t, output, "Epic Status: completed")
		assert.Contains(t, output, "Active Phase: none")
	})

	t.Run("current with JSON format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForCurrent()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		// Execute command with JSON format
		var stdout, stderr bytes.Buffer
		cmd := CurrentCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"current", "--format", "json"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, `"epic_status": "active"`)
		assert.Contains(t, output, `"active_phase": "P2"`)
		assert.Contains(t, output, `"active_task": "T2"`)
		assert.Contains(t, output, `"failing_tests": 1`)
		assert.Contains(t, output, `"next_action": "Fix failing tests`)
	})

	t.Run("current with XML format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForCurrent()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		// Execute command with XML format
		var stdout, stderr bytes.Buffer
		cmd := CurrentCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"current", "--format", "xml"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, `<current_state>`)
		assert.Contains(t, output, `<epic_status>active</epic_status>`)
		assert.Contains(t, output, `<active_phase>P2</active_phase>`)
		assert.Contains(t, output, `<active_task>T2</active_task>`)
		assert.Contains(t, output, `<failing_tests>1</failing_tests>`)
		assert.Contains(t, output, `<next_action>Fix failing tests`)
	})

	t.Run("current command error handling", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Test with missing config
		var stdout, stderr bytes.Buffer
		cmd := CurrentCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err := cmd.Run(context.Background(), []string{"current"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load configuration")
	})

	t.Run("current command aliases", func(t *testing.T) {
		cmd := CurrentCommand()
		assert.Equal(t, "current", cmd.Name)
		assert.Contains(t, cmd.Aliases, "c")
	})
}

func TestCurrentCommandPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	t.Run("current command performance", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForCurrent()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		// Measure execution time
		start := time.Now()

		var stdout, stderr bytes.Buffer
		cmd := CurrentCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"current"})
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Less(t, duration, 100*time.Millisecond, "Current command should execute in < 100ms")

		t.Logf("Current command executed in: %v", duration)
	})
}
