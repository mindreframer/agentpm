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
		assert.Contains(t, output, "Status: planning")
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
		assert.Contains(t, output, "Status: planning")
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
		assert.Contains(t, output, "Status: planning")
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
		assert.Contains(t, output, "Status: planning")
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
		assert.Equal(t, "planning", jsonData["status"])
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
		assert.Contains(t, output, "<task id=\"1A_T1\" phase_id=\"1A\" status=\"planning\">")
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
		assert.Equal(t, epic.StatusPlanning, epicData.Status)
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
		assert.Contains(t, output, "<epic id=\"SHOW_TEST\" status=\"planning\">")
		assert.Contains(t, output, "<name>Show Test Epic</name>")
		assert.Contains(t, output, "<phases>")
		assert.Contains(t, output, "<phase id=\"1A\" status=\"planning\">")
		assert.Contains(t, output, "<tasks>")
		assert.Contains(t, output, "<task id=\"1A_T1\" phase_id=\"1A\" status=\"planning\">")
		assert.Contains(t, output, "<tests>")
		assert.Contains(t, output, "<test id=\"1A_T1_TEST1\" task_id=\"1A_T1\" status=\"planning\">")
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
			Status: epic.StatusPlanning,
			Phases: []epic.Phase{
				{
					ID:     "1A",
					Name:   "Lonely Phase",
					Status: epic.StatusPlanning,
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
}

// Helper function to create test epic for show command tests
func createTestEpicForShow() *epic.Epic {
	return &epic.Epic{
		ID:          "SHOW_TEST",
		Name:        "Show Test Epic",
		Description: "Epic for testing show command functionality",
		Status:      epic.StatusPlanning,
		Phases: []epic.Phase{
			{
				ID:          "1A",
				Name:        "Phase One",
				Description: "First phase for testing",
				Status:      epic.StatusPlanning,
			},
			{
				ID:          "1B",
				Name:        "Phase Two",
				Description: "Second phase for testing",
				Status:      epic.StatusPlanning,
			},
		},
		Tasks: []epic.Task{
			{
				ID:          "1A_T1",
				PhaseID:     "1A",
				Name:        "Task One",
				Description: "First task for testing",
				Status:      epic.StatusPlanning,
			},
			{
				ID:          "1B_T1",
				PhaseID:     "1B",
				Name:        "Task Two",
				Description: "Second task for testing",
				Status:      epic.StatusPlanning,
			},
		},
		Tests: []epic.Test{
			{
				ID:          "1A_T1_TEST1",
				TaskID:      "1A_T1",
				Name:        "Test One",
				Description: "First test for testing",
				Status:      epic.StatusPlanning,
			},
			{
				ID:          "1B_T1_TEST1",
				TaskID:      "1B_T1",
				Name:        "Test Two",
				Description: "Second test for testing",
				Status:      epic.StatusPlanning,
			},
		},
		Events: []epic.Event{},
	}
}
