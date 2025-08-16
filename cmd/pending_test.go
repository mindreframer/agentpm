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

func createTestEpicForPending() *epic.Epic {
	return &epic.Epic{
		ID:        "pending-test-epic",
		Name:      "Pending Test Epic",
		Status:    epic.StatusActive,
		CreatedAt: time.Date(2025, 8, 16, 9, 0, 0, 0, time.UTC),
		Assignee:  "test_agent",
		Phases: []epic.Phase{
			{ID: "P1", Name: "Setup Phase", Status: epic.StatusCompleted},
			{ID: "P2", Name: "Implementation Phase", Status: epic.StatusActive},
			{ID: "P3", Name: "Testing Phase", Status: epic.StatusPlanning},
			{ID: "P4", Name: "Deployment Phase", Status: epic.StatusPlanning},
		},
		Tasks: []epic.Task{
			{ID: "T1", PhaseID: "P1", Name: "Setup Task", Status: epic.StatusCompleted},
			{ID: "T2", PhaseID: "P2", Name: "Active Task", Status: epic.StatusActive},
			{ID: "T3", PhaseID: "P2", Name: "Pending Task 1", Status: epic.StatusPlanning},
			{ID: "T4", PhaseID: "P3", Name: "Pending Task 2", Status: epic.StatusPlanning},
			{ID: "T5", PhaseID: "P4", Name: "Pending Task 3", Status: epic.StatusPlanning},
		},
		Tests: []epic.Test{
			{ID: "TEST1", TaskID: "T1", Name: "Setup Test", Status: epic.StatusCompleted},
			{ID: "TEST2", TaskID: "T2", Name: "Active Test", Status: epic.StatusPlanning},    // pending
			{ID: "TEST3", TaskID: "T3", Name: "Pending Test 1", Status: epic.StatusPlanning}, // pending
			{ID: "TEST4", TaskID: "T4", Name: "Pending Test 2", Status: epic.StatusPlanning}, // pending
		},
		Events: []epic.Event{},
	}
}

func createCompletedEpicForPending() *epic.Epic {
	return &epic.Epic{
		ID:     "completed-pending-epic",
		Name:   "Completed Pending Epic",
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

func TestPendingCommand(t *testing.T) {
	t.Run("pending with mixed work - text format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForPending()
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

		// Create pending command
		var stdout, stderr bytes.Buffer
		cmd := PendingCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		err = cmd.Run(context.Background(), []string{"pending"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Pending Work Overview")

		// Should show pending phases (P2 active, P3 and P4 planning = 3 total)
		assert.Contains(t, output, "Phases (3):")
		assert.Contains(t, output, "P2 - Implementation Phase [active]")
		assert.Contains(t, output, "P3 - Testing Phase [planning]")
		assert.Contains(t, output, "P4 - Deployment Phase [planning]")

		// Should show pending tasks (T2 active, T3, T4, T5 planning = 4 total)
		assert.Contains(t, output, "Tasks (4):")
		assert.Contains(t, output, "T2 (P2) - Active Task [active]")
		assert.Contains(t, output, "T3 (P2) - Pending Task 1 [planning]")
		assert.Contains(t, output, "T4 (P3) - Pending Task 2 [planning]")
		assert.Contains(t, output, "T5 (P4) - Pending Task 3 [planning]")

		// Should show pending tests (TEST2, TEST3, TEST4 = 3 total)
		assert.Contains(t, output, "Tests (3):")
		assert.Contains(t, output, "TEST2 (P2/T2) - Active Test [planning]")
		assert.Contains(t, output, "TEST3 (P2/T3) - Pending Test 1 [planning]")
		assert.Contains(t, output, "TEST4 (P3/T4) - Pending Test 2 [planning]")
	})

	t.Run("pending with completed epic - text format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create completed epic and config
		testEpic := createCompletedEpicForPending()
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

		// Create pending command
		var stdout, stderr bytes.Buffer
		cmd := PendingCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		err = cmd.Run(context.Background(), []string{"pending"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Pending Work Overview")
		assert.Contains(t, output, "Phases (0):")
		assert.Contains(t, output, "(none)")
		assert.Contains(t, output, "Tasks (0):")
		assert.Contains(t, output, "Tests (0):")
	})

	t.Run("pending with file override", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create two different epics
		pendingEpic := createTestEpicForPending()
		completedEpic := createCompletedEpicForPending()

		storage := storage.NewFileStorage()
		pendingPath := filepath.Join(tempDir, "pending.xml")
		completedPath := filepath.Join(tempDir, "completed.xml")

		err := storage.SaveEpic(pendingEpic, pendingPath)
		require.NoError(t, err)
		err = storage.SaveEpic(completedEpic, completedPath)
		require.NoError(t, err)

		// Config points to pending epic
		cfg := &config.Config{
			CurrentEpic:     pendingPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		// Execute pending command with file override to completed epic
		var stdout, stderr bytes.Buffer
		cmd := PendingCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"pending", "--file", completedPath})
		require.NoError(t, err)

		output := stdout.String()
		// Should show completed epic data (no pending work), not pending epic data
		assert.Contains(t, output, "Phases (0):")
		assert.Contains(t, output, "Tasks (0):")
		assert.Contains(t, output, "Tests (0):")
	})

	t.Run("pending with JSON format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForPending()
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
		cmd := PendingCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"pending", "--format", "json"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, `"phases": [`)
		assert.Contains(t, output, `"id": "P2"`)
		assert.Contains(t, output, `"name": "Implementation Phase"`)
		assert.Contains(t, output, `"status": "active"`)
		assert.Contains(t, output, `"tasks": [`)
		assert.Contains(t, output, `"id": "T2"`)
		assert.Contains(t, output, `"phase_id": "P2"`)
		assert.Contains(t, output, `"tests": [`)
		assert.Contains(t, output, `"task_id": "T2"`)
	})

	t.Run("pending with XML format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForPending()
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
		cmd := PendingCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"pending", "--format", "xml"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, `<pending_work>`)
		assert.Contains(t, output, `<phases>`)
		assert.Contains(t, output, `<phase id="P2" name="Implementation Phase" status="active"/>`)
		assert.Contains(t, output, `<tasks>`)
		assert.Contains(t, output, `<task id="T2" phase_id="P2" status="active">Active Task</task>`)
		assert.Contains(t, output, `<tests>`)
		assert.Contains(t, output, `<test id="TEST2" task_id="T2" phase_id="P2" status="planning">Active Test</test>`)
	})

	t.Run("pending command error handling", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Test with missing config
		var stdout, stderr bytes.Buffer
		cmd := PendingCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err := cmd.Run(context.Background(), []string{"pending"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load configuration")
	})

	t.Run("pending command aliases", func(t *testing.T) {
		cmd := PendingCommand()
		assert.Equal(t, "pending", cmd.Name)
		assert.Contains(t, cmd.Aliases, "pend")
	})
}

func TestPendingCommandEdgeCases(t *testing.T) {
	t.Run("pending with empty epic", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create minimal epic with no phases/tasks/tests
		emptyEpic := &epic.Epic{
			ID:        "empty-epic",
			Name:      "Empty Epic",
			Status:    epic.StatusPlanning,
			CreatedAt: time.Now(),
			Phases:    []epic.Phase{},
			Tasks:     []epic.Task{},
			Tests:     []epic.Test{},
			Events:    []epic.Event{},
		}

		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "empty-epic.xml")
		err := storage.SaveEpic(emptyEpic, epicPath)
		require.NoError(t, err)

		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		// Execute command
		var stdout, stderr bytes.Buffer
		cmd := PendingCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"pending"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Pending Work Overview")
		assert.Contains(t, output, "Phases (0):")
		assert.Contains(t, output, "Tasks (0):")
		assert.Contains(t, output, "Tests (0):")
		assert.Contains(t, output, "(none)")
	})
}

func TestPendingCommandPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	t.Run("pending command performance", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForPending()
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
		cmd := PendingCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"pending"})
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Less(t, duration, 100*time.Millisecond, "Pending command should execute in < 100ms")

		t.Logf("Pending command executed in: %v", duration)
	})
}
