package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowCommand(t *testing.T) {
	t.Run("show command basic structure", func(t *testing.T) {
		cmd := ShowCommand()
		assert.Equal(t, "show", cmd.Name)
		assert.Equal(t, "Display detailed information about epic entities", cmd.Usage)
		assert.Contains(t, cmd.Description, "Entity types:")
		assert.Contains(t, cmd.Description, "epic")
		assert.Contains(t, cmd.Description, "phase")
		assert.Contains(t, cmd.Description, "task")
		assert.Contains(t, cmd.Description, "test")
	})

	t.Run("show epic displays full epic information", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForShow()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "epic", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Epic: Show Test Epic")
		assert.Contains(t, output, "ID: SHOW_TEST")
		assert.Contains(t, output, "Status: pending")
		assert.Contains(t, output, "Phases (2):")
		assert.Contains(t, output, "1A - Phase One")
		assert.Contains(t, output, "1B - Phase Two")
		assert.Contains(t, output, "Tasks (2):")
		assert.Contains(t, output, "1A_T1 (1A) - Task One")
		assert.Contains(t, output, "1B_T1 (1B) - Task Two")
		assert.Contains(t, output, "Tests (2):")
		assert.Contains(t, output, "1A_T1_TEST1 (1A_T1) - Test One")
		assert.Contains(t, output, "1B_T1_TEST1 (1B_T1) - Test Two")
	})

	t.Run("show phase displays phase details and related items", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForShow()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "phase", "1A", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Phase: Phase One")
		assert.Contains(t, output, "ID: 1A")
		assert.Contains(t, output, "Status: pending")
		assert.Contains(t, output, "Tasks (1):")
		assert.Contains(t, output, "1A_T1 - Task One")
		assert.Contains(t, output, "Tests (1):")
		assert.Contains(t, output, "1A_T1_TEST1 - Test One")
	})

	t.Run("show task displays task details and related items", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForShow()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "task", "1A_T1", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Task: Task One")
		assert.Contains(t, output, "ID: 1A_T1")
		assert.Contains(t, output, "Phase: 1A")
		assert.Contains(t, output, "Status: pending")
		assert.Contains(t, output, "Parent Phase: 1A - Phase One")
		assert.Contains(t, output, "Tests (1):")
		assert.Contains(t, output, "1A_T1_TEST1 - Test One")
	})

	t.Run("show test displays test details and related items", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForShow()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "test", "1A_T1_TEST1", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Test: Test One")
		assert.Contains(t, output, "ID: 1A_T1_TEST1")
		assert.Contains(t, output, "Task: 1A_T1")
		assert.Contains(t, output, "Status: pending")
		assert.Contains(t, output, "Parent Task: 1A_T1 - Task One")
		assert.Contains(t, output, "Parent Phase: 1A - Phase One")
	})

	t.Run("show JSON format works correctly", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForShow()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "phase", "1A", "--format", "json", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()

		// Verify it's valid JSON
		var jsonData map[string]interface{}
		err = json.Unmarshal([]byte(output), &jsonData)
		require.NoError(t, err)

		assert.Equal(t, "1A", jsonData["id"])
		assert.Equal(t, "Phase One", jsonData["name"])
		assert.Equal(t, "pending", jsonData["status"])
		assert.Contains(t, jsonData, "related")
	})

	t.Run("show XML format works correctly", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForShow()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "task", "1A_T1", "--format", "xml", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()

		// Verify XML structure
		assert.Contains(t, output, "<task id=\"1A_T1\" phase_id=\"1A\" status=\"pending\">")
		assert.Contains(t, output, "<name>Task One</name>")
		assert.Contains(t, output, "<related>")
		assert.Contains(t, output, "</task>")
	})

	t.Run("error handling for invalid entity types", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForShow()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "invalid_type", "--file", epicPath})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid entity type: invalid_type")
	})

	t.Run("error handling for missing IDs", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForShow()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "phase", "--file", epicPath})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "phase requires an ID")
	})

	t.Run("error handling for missing entity types", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForShow()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "--file", epicPath})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "entity type is required")
	})

	t.Run("error handling for non-existent entities", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForShow()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "phase", "NON_EXISTENT", "--file", epicPath})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "phase NON_EXISTENT not found")
	})

	t.Run("show epic with JSON format outputs complete epic structure", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForShow()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "epic", "--format", "json", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()

		// Verify it's valid JSON and contains expected structure
		var epicData epic.Epic
		err = json.Unmarshal([]byte(output), &epicData)
		require.NoError(t, err)

		assert.Equal(t, "SHOW_TEST", epicData.ID)
		assert.Equal(t, "Show Test Epic", epicData.Name)
		assert.Equal(t, epic.StatusPending, epicData.Status)
		assert.Len(t, epicData.Phases, 2)
		assert.Len(t, epicData.Tasks, 2)
		assert.Len(t, epicData.Tests, 2)
	})

	t.Run("show epic with XML format outputs complete epic structure", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForShow()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "epic", "--format", "xml", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()

		// Verify XML structure
		assert.Contains(t, output, "<epic id=\"SHOW_TEST\" status=\"pending\">")
		assert.Contains(t, output, "<name>Show Test Epic</name>")
		assert.Contains(t, output, "<phases>")
		assert.Contains(t, output, "<phase id=\"1A\" status=\"pending\">")
		assert.Contains(t, output, "<tasks>")
		assert.Contains(t, output, "<task id=\"1A_T1\" phase_id=\"1A\" status=\"pending\">")
		assert.Contains(t, output, "<tests>")
		assert.Contains(t, output, "<test id=\"1A_T1_TEST1\" task_id=\"1A_T1\" status=\"pending\">")
		assert.Contains(t, output, "</epic>")
	})

	t.Run("show handles empty related items gracefully", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create minimal epic with just one phase
		testEpic := &epic.Epic{
			ID:     "MIN_TEST",
			Name:   "Minimal Test Epic",
			Status: epic.StatusPending,
			Phases: []epic.Phase{
				{
					ID:     "1A",
					Name:   "Lonely Phase",
					Status: epic.StatusPending,
				},
			},
			Tasks: []epic.Task{},
			Tests: []epic.Test{},
		}

		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "phase", "1A", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Phase: Lonely Phase")
		assert.Contains(t, output, "Tasks (0):")
		assert.Contains(t, output, "(none)")
		assert.Contains(t, output, "Tests (0):")
	})

	// Test the new --full flag functionality
	t.Run("show task with --full flag displays full context including siblings", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic with sibling tasks
		testEpic := createTestEpicWithSiblings()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "task", "1A_T1", "--full", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()

		// Check for text format output (default) with siblings
		assert.Contains(t, output, "Task: Task One")
		assert.Contains(t, output, "ID: 1A_T1")
		assert.Contains(t, output, "Parent Phase:")
		assert.Contains(t, output, "Phase One")
		assert.Contains(t, output, "Sibling Tasks")
		assert.Contains(t, output, "1A_T2") // Should have sibling
		assert.Contains(t, output, "Child Tests")
	})

	t.Run("show task with --full flag displays full context", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForShow()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "task", "1A_T1", "--full", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()

		// Check for text format output (default)
		assert.Contains(t, output, "Task: Task One")
		assert.Contains(t, output, "ID: 1A_T1")
		assert.Contains(t, output, "Parent Phase:")
		assert.Contains(t, output, "Phase One")
		assert.Contains(t, output, "Child Tests")
	})

	t.Run("show phase with --full flag displays full context", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForShow()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "phase", "1A", "--full", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()

		// Check for text format output with full context
		assert.Contains(t, output, "Phase: Phase One")
		assert.Contains(t, output, "ID: 1A")
		assert.Contains(t, output, "Progress Summary:")
		assert.Contains(t, output, "All Tasks")
		assert.Contains(t, output, "Sibling Phases")
	})

	t.Run("show test with --full flag displays full context", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForShow()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "test", "1A_T1_TEST1", "--full", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()

		// Check for text format output with full context
		assert.Contains(t, output, "Test: Test One")
		assert.Contains(t, output, "ID: 1A_T1_TEST1")
		assert.Contains(t, output, "Parent Task:")
		assert.Contains(t, output, "Task One")
		assert.Contains(t, output, "Parent Phase:")
		assert.Contains(t, output, "Phase One")
	})

	t.Run("show task with --full flag and XML format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForShow()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "task", "1A_T1", "--full", "--format", "xml", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()

		// Check for XML format output with full context
		assert.Contains(t, output, "<task_context")
		assert.Contains(t, output, "id=\"1A_T1\"")
		assert.Contains(t, output, "<task_details>")
		assert.Contains(t, output, "<name>Task One</name>")
		assert.Contains(t, output, "<parent_phase")
		assert.Contains(t, output, "<child_tests>")
		assert.Contains(t, output, "</task_context>")
	})

	t.Run("show task with --full flag and JSON format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForShow()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "task", "1A_T1", "--full", "--format", "json", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()

		// Verify it's valid JSON
		var contextData map[string]interface{}
		err = json.Unmarshal([]byte(output), &contextData)
		require.NoError(t, err)

		// Check basic structure
		taskDetails, ok := contextData["task_details"].(map[string]interface{})
		require.True(t, ok, "task_details should be present")
		assert.Equal(t, "1A_T1", taskDetails["id"])
		assert.Equal(t, "Task One", taskDetails["name"])

		// Check that parent phase is included
		parentPhase, ok := contextData["parent_phase"].(map[string]interface{})
		require.True(t, ok, "parent_phase should be present")
		assert.Equal(t, "1A", parentPhase["id"])
	})

	t.Run("backward compatibility: show without --full flag works as before", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForShow()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Run without --full flag
		err = cmd.Run(context.Background(), []string{"show", "task", "1A_T1", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()

		// Should use original compact format
		assert.Contains(t, output, "Task: Task One")
		assert.Contains(t, output, "ID: 1A_T1")
		assert.Contains(t, output, "Phase: 1A")
		assert.Contains(t, output, "Status: pending")
		// Should NOT contain full context elements that are specific to full mode
		// The key difference is that full mode shows rich hierarchical context
		// while normal mode shows simple relationships
		assert.NotContains(t, output, "Progress Summary:")
		assert.NotContains(t, output, "Phase Progress:")
		assert.NotContains(t, output, "Sibling Tasks")
		assert.NotContains(t, output, "Child Tests")
	})

	t.Run("show phase with phase-level tests displays them correctly", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create epic with phase-level tests (no task_id)
		testEpic := &epic.Epic{
			ID:     "TEST_PHASE_TESTS",
			Name:   "Test Epic with Phase Tests",
			Status: epic.StatusPending,
			Phases: []epic.Phase{
				{
					ID:          "2B",
					Name:        "Test Phase",
					Description: "Phase with direct tests",
					Status:      epic.StatusPending,
				},
			},
			Tasks: []epic.Task{}, // No tasks
			Tests: []epic.Test{
				{
					ID:          "2B_1",
					TaskID:      "", // No task - phase-level test
					PhaseID:     "2B",
					Name:        "Phase Test 1",
					Description: "Direct phase test",
					Status:      epic.StatusPending,
				},
				{
					ID:          "2B_2",
					TaskID:      "", // No task - phase-level test
					PhaseID:     "2B",
					Name:        "Phase Test 2",
					Description: "Another direct phase test",
					Status:      epic.StatusPending,
				},
			},
			Events: []epic.Event{},
		}

		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "phase", "2B", "--full", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()

		// Should show phase tests
		assert.Contains(t, output, "Progress Summary:")
		assert.Contains(t, output, "Tests: 2 total")
		assert.Contains(t, output, "Phase Tests (2):")
		assert.Contains(t, output, "2B_1 - Phase Test 1")
		assert.Contains(t, output, "2B_2 - Phase Test 2")
		assert.Contains(t, output, "Direct phase test")
	})

	t.Run("show task with --full flag displays acceptance criteria", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create epic with tasks having acceptance criteria
		testEpic := &epic.Epic{
			ID:     "CRITERIA_TEST",
			Name:   "Test Epic with Acceptance Criteria",
			Status: epic.StatusPending,
			Phases: []epic.Phase{
				{
					ID:          "1A",
					Name:        "Test Phase",
					Description: "Phase for criteria testing",
					Status:      epic.StatusWIP,
				},
			},
			Tasks: []epic.Task{
				{
					ID:                 "1A_1",
					PhaseID:            "1A",
					Name:               "Main Task",
					Description:        "Task with criteria",
					AcceptanceCriteria: "- Requirement 1 must be met\n- All tests must pass\n- Code review approved",
					Status:             epic.StatusCompleted,
				},
				{
					ID:                 "1A_2",
					PhaseID:            "1A",
					Name:               "Sibling Task",
					Description:        "Another task with criteria",
					AcceptanceCriteria: "- Feature implemented\n- Documentation updated",
					Status:             epic.StatusPending,
				},
			},
			Tests:  []epic.Test{},
			Events: []epic.Event{},
		}

		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "task", "1A_1", "--full", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()

		// Should show acceptance criteria for main task
		assert.Contains(t, output, "Acceptance Criteria:")
		assert.Contains(t, output, "- Requirement 1 must be met")
		assert.Contains(t, output, "- All tests must pass")
		assert.Contains(t, output, "- Code review approved")

		// Should show acceptance criteria for sibling task
		assert.Contains(t, output, "Sibling Tasks")
		assert.Contains(t, output, "1A_2 - Sibling Task")
		assert.Contains(t, output, "- Feature implemented")
		assert.Contains(t, output, "- Documentation updated")
	})

	t.Run("show phase with --full flag displays acceptance criteria for all tasks", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create epic with tasks having acceptance criteria
		testEpic := &epic.Epic{
			ID:     "PHASE_CRITERIA_TEST",
			Name:   "Test Epic for Phase Criteria",
			Status: epic.StatusPending,
			Phases: []epic.Phase{
				{
					ID:          "1A",
					Name:        "Test Phase",
					Description: "Phase for criteria testing",
					Status:      epic.StatusWIP,
				},
			},
			Tasks: []epic.Task{
				{
					ID:                 "1A_1",
					PhaseID:            "1A",
					Name:               "Task One",
					Description:        "First task",
					AcceptanceCriteria: "- Task 1 criteria met\n- Quality assured",
					Status:             epic.StatusCompleted,
				},
				{
					ID:                 "1A_2",
					PhaseID:            "1A",
					Name:               "Task Two",
					Description:        "Second task",
					AcceptanceCriteria: "- Task 2 requirements\n- Testing complete",
					Status:             epic.StatusPending,
				},
			},
			Tests:  []epic.Test{},
			Events: []epic.Event{},
		}

		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := ShowCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err = cmd.Run(context.Background(), []string{"show", "phase", "1A", "--full", "--file", epicPath})
		require.NoError(t, err)

		output := stdout.String()

		// Should show acceptance criteria for all tasks in the phase
		assert.Contains(t, output, "All Tasks")
		assert.Contains(t, output, "1A_1 - Task One")
		assert.Contains(t, output, "- Task 1 criteria met")
		assert.Contains(t, output, "- Quality assured")
		assert.Contains(t, output, "1A_2 - Task Two")
		assert.Contains(t, output, "- Task 2 requirements")
		assert.Contains(t, output, "- Testing complete")
	})
}

func createTestEpicWithSiblings() *epic.Epic {
	return &epic.Epic{
		ID:          "SHOW_TEST",
		Name:        "Show Test Epic",
		Description: "Epic for testing show command",
		Status:      epic.StatusPending,
		Phases: []epic.Phase{
			{
				ID:          "1A",
				Name:        "Phase One",
				Description: "First phase for testing",
				Status:      epic.StatusPending,
			},
			{
				ID:          "1B",
				Name:        "Phase Two",
				Description: "Second phase for testing",
				Status:      epic.StatusPending,
			},
		},
		Tasks: []epic.Task{
			{
				ID:          "1A_T1",
				PhaseID:     "1A",
				Name:        "Task One",
				Description: "First task for testing",
				Status:      epic.StatusPending,
			},
			{
				ID:          "1A_T2",
				PhaseID:     "1A",
				Name:        "Task Two",
				Description: "Second task in same phase",
				Status:      epic.StatusPending,
			},
			{
				ID:          "1B_T1",
				PhaseID:     "1B",
				Name:        "Task Three",
				Description: "Third task for testing",
				Status:      epic.StatusPending,
			},
		},
		Tests: []epic.Test{
			{
				ID:          "1A_T1_TEST1",
				TaskID:      "1A_T1",
				Name:        "Test One",
				Description: "First test for testing",
				Status:      epic.StatusPending,
			},
			{
				ID:          "1A_T2_TEST1",
				TaskID:      "1A_T2",
				Name:        "Test Two",
				Description: "Second test for testing",
				Status:      epic.StatusPending,
			},
		},
		Events: []epic.Event{},
	}
}

func createTestEpicForShow() *epic.Epic {
	return &epic.Epic{
		ID:          "SHOW_TEST",
		Name:        "Show Test Epic",
		Description: "Epic for testing show command functionality",
		Status:      epic.StatusPending,
		Phases: []epic.Phase{
			{
				ID:          "1A",
				Name:        "Phase One",
				Description: "First phase for testing",
				Status:      epic.StatusPending,
			},
			{
				ID:          "1B",
				Name:        "Phase Two",
				Description: "Second phase for testing",
				Status:      epic.StatusPending,
			},
		},
		Tasks: []epic.Task{
			{
				ID:          "1A_T1",
				PhaseID:     "1A",
				Name:        "Task One",
				Description: "First task for testing",
				Status:      epic.StatusPending,
			},
			{
				ID:          "1B_T1",
				PhaseID:     "1B",
				Name:        "Task Two",
				Description: "Second task for testing",
				Status:      epic.StatusPending,
			},
		},
		Tests: []epic.Test{
			{
				ID:          "1A_T1_TEST1",
				TaskID:      "1A_T1",
				Name:        "Test One",
				Description: "First test for testing",
				Status:      epic.StatusPending,
			},
			{
				ID:          "1B_T1_TEST1",
				TaskID:      "1B_T1",
				Name:        "Test Two",
				Description: "Second test for testing",
				Status:      epic.StatusPending,
			},
		},
		Events: []epic.Event{},
	}
}
