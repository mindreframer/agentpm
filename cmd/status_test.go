package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/memomoo/agentpm/internal/config"
	"github.com/memomoo/agentpm/internal/epic"
	"github.com/memomoo/agentpm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestEpicForStatus() *epic.Epic {
	return &epic.Epic{
		ID:        "status-test-epic",
		Name:      "Status Test Epic",
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
			{ID: "T4", PhaseID: "P3", Name: "Future Task", Status: epic.StatusPlanning},
		},
		Tests: []epic.Test{
			{ID: "TEST1", TaskID: "T1", Name: "Setup Test", Status: epic.StatusCompleted},
			{ID: "TEST2", TaskID: "T2", Name: "Active Test", Status: epic.StatusPlanning},  // "failing"
			{ID: "TEST3", TaskID: "T3", Name: "Pending Test", Status: epic.StatusPlanning}, // "failing"
		},
		Events: []epic.Event{
			{
				ID:        "E1",
				Type:      "created",
				Timestamp: time.Date(2025, 8, 16, 9, 0, 0, 0, time.UTC),
				Data:      "Epic created",
			},
		},
	}
}

func createCompletedEpicForStatus() *epic.Epic {
	return &epic.Epic{
		ID:     "completed-status-epic",
		Name:   "Completed Status Epic",
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

func TestStatusCommand(t *testing.T) {
	t.Run("status with active epic - text format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForStatus()
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

		// Create status command
		var stdout, stderr bytes.Buffer
		cmd := StatusCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		err = cmd.Run(context.Background(), []string{"status"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Epic Status: Status Test Epic")
		assert.Contains(t, output, "ID: status-test-epic")
		assert.Contains(t, output, "Status: active")
		assert.Contains(t, output, "Progress: 30% complete") // Enhanced weighted calculation: 40%*phases + 40%*tasks + 20%*tests
		assert.Contains(t, output, "Phases: 1/3 completed")
		assert.Contains(t, output, "Tests: 1 passing, 2 failing")
		assert.Contains(t, output, "Current Phase: P2")
		assert.Contains(t, output, "Current Task: T2")
	})

	t.Run("status with completed epic - text format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create completed epic and config
		testEpic := createCompletedEpicForStatus()
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

		// Create status command
		var stdout, stderr bytes.Buffer
		cmd := StatusCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		err = cmd.Run(context.Background(), []string{"status"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Epic Status: Completed Status Epic")
		assert.Contains(t, output, "Status: completed")
		assert.Contains(t, output, "Progress: 100% complete")
		assert.Contains(t, output, "Phases: 1/1 completed")
		assert.Contains(t, output, "Tests: 1 passing, 0 failing")
		// Should not show current phase/task for completed epic
		assert.NotContains(t, output, "Current Phase:")
		assert.NotContains(t, output, "Current Task:")
	})

	t.Run("status with file override", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create two different epics
		testEpic1 := createTestEpicForStatus()
		testEpic2 := createCompletedEpicForStatus()

		storage := storage.NewFileStorage()
		epic1Path := filepath.Join(tempDir, "epic1.xml")
		epic2Path := filepath.Join(tempDir, "epic2.xml")

		err := storage.SaveEpic(testEpic1, epic1Path)
		require.NoError(t, err)
		err = storage.SaveEpic(testEpic2, epic2Path)
		require.NoError(t, err)

		// Config points to epic1
		cfg := &config.Config{
			CurrentEpic:     epic1Path,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		// Execute status command with file override to epic2
		var stdout, stderr bytes.Buffer
		cmd := StatusCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"status", "--file", epic2Path})
		require.NoError(t, err)

		output := stdout.String()
		// Should show epic2 data (completed), not epic1 data (active)
		assert.Contains(t, output, "Completed Status Epic")
		assert.Contains(t, output, "Status: completed")
		assert.Contains(t, output, "Progress: 100% complete")
	})

	t.Run("status with JSON format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForStatus()
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
		cmd := StatusCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"status", "--format", "json"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, `"epic": "status-test-epic"`)
		assert.Contains(t, output, `"name": "Status Test Epic"`)
		assert.Contains(t, output, `"status": "active"`)
		assert.Contains(t, output, `"completion_percentage": 30`)
		assert.Contains(t, output, `"completed_phases": 1`)
		assert.Contains(t, output, `"total_phases": 3`)
		assert.Contains(t, output, `"passing_tests": 1`)
		assert.Contains(t, output, `"failing_tests": 2`)
		assert.Contains(t, output, `"current_phase": "P2"`)
		assert.Contains(t, output, `"current_task": "T2"`)
	})

	t.Run("status with XML format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForStatus()
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
		cmd := StatusCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"status", "--format", "xml"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, `<status epic="status-test-epic">`)
		assert.Contains(t, output, `<name>Status Test Epic</name>`)
		assert.Contains(t, output, `<status>active</status>`)
		assert.Contains(t, output, `<completion_percentage>30</completion_percentage>`)
		assert.Contains(t, output, `<completed_phases>1</completed_phases>`)
		assert.Contains(t, output, `<total_phases>3</total_phases>`)
		assert.Contains(t, output, `<passing_tests>1</passing_tests>`)
		assert.Contains(t, output, `<failing_tests>2</failing_tests>`)
		assert.Contains(t, output, `<current_phase>P2</current_phase>`)
		assert.Contains(t, output, `<current_task>T2</current_task>`)
	})

	t.Run("status command error handling", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Test with missing config
		var stdout, stderr bytes.Buffer
		cmd := StatusCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err := cmd.Run(context.Background(), []string{"status"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load configuration")
	})

	t.Run("status with missing epic file", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create config pointing to non-existent epic
		cfg := &config.Config{
			CurrentEpic:     "missing-epic.xml",
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err := config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := StatusCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"status"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load epic")
	})

	t.Run("status with no epic file specified", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create config with empty current epic (will fail validation, but we need to manually write it)
		configPath := filepath.Join(tempDir, ".agentpm.json")
		configData := `{"current_epic":"","default_assignee":"test_agent"}`
		err := os.WriteFile(configPath, []byte(configData), 0644)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := StatusCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"status"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "current_epic is required")
	})

	t.Run("status command flag handling", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and minimal config (needed for CLI to not fail completely)
		testEpic := createTestEpicForStatus()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create minimal config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		// Test file flag (-f)
		var stdout, stderr bytes.Buffer
		cmd := StatusCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"status", "-f", epicPath})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Status Test Epic")

		// Test format flag (-F)
		stdout.Reset()
		err = cmd.Run(context.Background(), []string{"status", "-f", epicPath, "-F", "json"})
		require.NoError(t, err)

		output = stdout.String()
		assert.Contains(t, output, `"epic": "status-test-epic"`)
	})

	t.Run("status with empty epic - edge case", func(t *testing.T) {
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
		cmd := StatusCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"status"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Epic Status: Empty Epic")
		assert.Contains(t, output, "Progress: 0% complete") // Division by zero handled
		assert.Contains(t, output, "Phases: 0/0 completed")
		assert.Contains(t, output, "Tests: 0 passing, 0 failing")
	})
}

func TestStatusCommandAliases(t *testing.T) {
	t.Run("status command has correct aliases", func(t *testing.T) {
		cmd := StatusCommand()
		assert.Equal(t, "status", cmd.Name)
		assert.Contains(t, cmd.Aliases, "st")
	})

	t.Run("status alias works", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForStatus()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create minimal config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		// Test that alias "st" works the same as "status"
		var stdout, stderr bytes.Buffer
		cmd := StatusCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"st", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Status Test Epic")
	})
}

func TestStatusCommandOutputConsistency(t *testing.T) {
	t.Run("output format consistency across runs", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForStatus()
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

		// Run command multiple times and verify consistent output
		results := make([]string, 3)
		for i := 0; i < 3; i++ {
			var stdout, stderr bytes.Buffer
			cmd := StatusCommand()
			cmd.Root().Writer = &stdout
			cmd.Root().ErrWriter = &stderr

			err = cmd.Run(context.Background(), []string{"status"})
			require.NoError(t, err)
			results[i] = stdout.String()
		}

		// All outputs should be identical
		for i := 1; i < len(results); i++ {
			assert.Equal(t, results[0], results[i], "Status output should be consistent across runs")
		}
	})

	t.Run("multi-format output consistency", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForStatus()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create minimal config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var textOut, jsonOut, xmlOut bytes.Buffer

		// Get text format output
		cmd := StatusCommand()
		cmd.Root().Writer = &textOut
		err = cmd.Run(context.Background(), []string{"status", "--file", epicPath, "--format", "text"})
		require.NoError(t, err)

		// Get JSON format output
		cmd = StatusCommand()
		cmd.Root().Writer = &jsonOut
		err = cmd.Run(context.Background(), []string{"status", "--file", epicPath, "--format", "json"})
		require.NoError(t, err)

		// Get XML format output
		cmd = StatusCommand()
		cmd.Root().Writer = &xmlOut
		err = cmd.Run(context.Background(), []string{"status", "--file", epicPath, "--format", "xml"})
		require.NoError(t, err)

		textOutput := textOut.String()
		jsonOutput := jsonOut.String()
		xmlOutput := xmlOut.String()

		// Verify each format contains the core information
		coreInfo := []string{"Status Test Epic", "status-test-epic", "active", "30"}

		for _, info := range coreInfo {
			assert.Contains(t, textOutput, info, "Text format missing: %s", info)
		}

		for _, info := range coreInfo {
			assert.Contains(t, jsonOutput, info, "JSON format missing: %s", info)
		}

		for _, info := range coreInfo {
			assert.Contains(t, xmlOutput, info, "XML format missing: %s", info)
		}

		// Verify format-specific structure
		assert.Contains(t, textOutput, "Epic Status:")
		assert.Contains(t, jsonOutput, `"epic":`)
		assert.Contains(t, xmlOutput, `<status epic=`)
	})
}

func TestStatusCommandPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	t.Run("status command performance", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForStatus()
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
		cmd := StatusCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"status"})
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Less(t, duration, 100*time.Millisecond, "Status command should execute in < 100ms")

		t.Logf("Status command executed in: %v", duration)
	})
}
