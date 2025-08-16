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
	"github.com/urfave/cli/v3"
)

func setupTestApp() *cli.Command {
	return &cli.Command{
		Name:  "agentpm",
		Usage: "CLI tool for LLM agents to manage epic-based development work",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Override epic file from config",
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Override config file path",
				Value:   "./.agentpm.json",
			},
			&cli.StringFlag{
				Name:    "time",
				Aliases: []string{"t"},
				Usage:   "Timestamp for current time (testing support)",
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"F"},
				Usage:   "Output format - text (default) / json / xml",
				Value:   "text",
			},
		},
		Commands: []*cli.Command{
			InitCommand(),
			ConfigCommand(),
			ValidateCommand(),
		},
	}
}

func createTestEpic(tempDir string, filename string, valid bool) string {
	epicPath := filepath.Join(tempDir, filename)

	var e *epic.Epic
	if valid {
		e = &epic.Epic{
			ID:        "test-1",
			Name:      "Test Epic",
			Status:    epic.StatusPlanning,
			CreatedAt: time.Date(2025, 8, 16, 9, 0, 0, 0, time.UTC),
			Assignee:  "test_agent",
			Phases: []epic.Phase{
				{ID: "P1", Name: "Phase 1", Status: epic.StatusPlanning},
			},
			Tasks: []epic.Task{
				{ID: "T1", PhaseID: "P1", Name: "Task 1", Status: epic.StatusPlanning},
			},
			Tests: []epic.Test{
				{ID: "TEST1", TaskID: "T1", Name: "Test 1", Status: epic.StatusPlanning},
			},
		}
	} else {
		e = &epic.Epic{
			// Invalid epic - missing required fields
			Name: "Invalid Epic",
		}
	}

	storage := storage.NewFileStorage()
	err := storage.SaveEpic(e, epicPath)
	if err != nil {
		panic(err)
	}

	return epicPath
}

func TestInitCommand(t *testing.T) {
	t.Run("init with valid epic file", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		epicPath := createTestEpic(tempDir, "test-epic.xml", true)

		var stdout, stderr bytes.Buffer
		app := setupTestApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err := app.Run(context.Background(), []string{"agentpm", "init", "--epic", epicPath})

		assert.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "✓ Project initialized successfully")
		assert.Contains(t, output, epicPath)

		// Verify config file was created
		assert.True(t, config.ConfigExists(""))

		// Verify config content
		cfg, err := config.LoadConfig("")
		require.NoError(t, err)
		assert.Equal(t, epicPath, cfg.CurrentEpic)
		assert.Equal(t, "agent", cfg.DefaultAssignee)
	})

	t.Run("init with non-existent epic file", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		var stdout, stderr bytes.Buffer
		app := setupTestApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err := app.Run(context.Background(), []string{"agentpm", "init", "--epic", "missing.xml"})

		assert.Error(t, err)
		output := stderr.String()
		assert.Contains(t, output, "Epic file not found")
	})

	t.Run("init with invalid epic file", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create malformed XML file
		invalidPath := filepath.Join(tempDir, "invalid.xml")
		err := os.WriteFile(invalidPath, []byte(`<?xml version="1.0"?><epic>malformed`), 0644)
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		app := setupTestApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err = app.Run(context.Background(), []string{"agentpm", "init", "--epic", invalidPath})

		assert.Error(t, err)
		output := stderr.String()
		assert.Contains(t, output, "Failed to load epic file")
	})

	t.Run("init preserves existing config", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		epicPath := createTestEpic(tempDir, "test-epic.xml", true)

		// Create initial config
		initialCfg := &config.Config{
			CurrentEpic:     "old-epic.xml",
			ProjectName:     "MyProject",
			DefaultAssignee: "custom_agent",
		}
		err := config.SaveConfig(initialCfg, "")
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		app := setupTestApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err = app.Run(context.Background(), []string{"agentpm", "init", "--epic", epicPath})

		assert.NoError(t, err)

		// Verify preserved values
		cfg, err := config.LoadConfig("")
		require.NoError(t, err)
		assert.Equal(t, epicPath, cfg.CurrentEpic)
		assert.Equal(t, "MyProject", cfg.ProjectName)        // Preserved
		assert.Equal(t, "custom_agent", cfg.DefaultAssignee) // Preserved
	})

	t.Run("init with XML format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		epicPath := createTestEpic(tempDir, "test-epic.xml", true)

		var stdout, stderr bytes.Buffer
		app := setupTestApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err := app.Run(context.Background(), []string{"agentpm", "--format", "xml", "init", "--epic", epicPath})

		assert.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "<init_result>")
		assert.Contains(t, output, "<project_created>true</project_created>")
		assert.Contains(t, output, epicPath)
	})

	t.Run("init with JSON format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		epicPath := createTestEpic(tempDir, "test-epic.xml", true)

		var stdout, stderr bytes.Buffer
		app := setupTestApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err := app.Run(context.Background(), []string{"agentpm", "--format", "json", "init", "--epic", epicPath})

		assert.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, `"project_created": true`)
		assert.Contains(t, output, epicPath)
	})
}

func TestConfigCommand(t *testing.T) {
	t.Run("config display with valid configuration", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		epicPath := createTestEpic(tempDir, "test-epic.xml", true)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     epicPath,
			ProjectName:     "TestProject",
			DefaultAssignee: "test_agent",
		}
		err := config.SaveConfig(cfg, "")
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		app := setupTestApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err = app.Run(context.Background(), []string{"agentpm", "config"})

		assert.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "Current Configuration:")
		assert.Contains(t, output, epicPath)
		assert.Contains(t, output, "TestProject")
		assert.Contains(t, output, "test_agent")
	})

	t.Run("config with missing epic file warning", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create config with non-existent epic
		cfg := &config.Config{
			CurrentEpic:     "missing-epic.xml",
			DefaultAssignee: "agent",
		}
		err := config.SaveConfig(cfg, "")
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		app := setupTestApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err = app.Run(context.Background(), []string{"agentpm", "config"})

		assert.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "Warning")
		assert.Contains(t, output, "Epic file not found")
	})

	t.Run("config with missing configuration file", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		var stdout, stderr bytes.Buffer
		app := setupTestApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err := app.Run(context.Background(), []string{"agentpm", "config"})

		assert.Error(t, err)
		output := stderr.String()
		assert.Contains(t, output, "Failed to load configuration")
	})

	t.Run("config with XML format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		epicPath := createTestEpic(tempDir, "test-epic.xml", true)

		cfg := &config.Config{
			CurrentEpic:     epicPath,
			ProjectName:     "TestProject",
			DefaultAssignee: "test_agent",
		}
		err := config.SaveConfig(cfg, "")
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		app := setupTestApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err = app.Run(context.Background(), []string{"agentpm", "--format", "xml", "config"})

		assert.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "<config>")
		assert.Contains(t, output, "<current_epic>")
		assert.Contains(t, output, "<project_name>")
	})

	t.Run("config with JSON format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		epicPath := createTestEpic(tempDir, "test-epic.xml", true)

		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		err := config.SaveConfig(cfg, "")
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		app := setupTestApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err = app.Run(context.Background(), []string{"agentpm", "--format", "json", "config"})

		assert.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, `"current_epic":`)
		assert.Contains(t, output, `"default_assignee":`)
	})
}

func TestValidateCommand(t *testing.T) {
	t.Run("validate with valid epic from config", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		epicPath := createTestEpic(tempDir, "test-epic.xml", true)

		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "agent",
		}
		err := config.SaveConfig(cfg, "")
		require.NoError(t, err)

		var stdout, stderr bytes.Buffer
		app := setupTestApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err = app.Run(context.Background(), []string{"agentpm", "validate"})

		assert.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "✓ Epic validation passed")
	})

	t.Run("validate with file override", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		epicPath := createTestEpic(tempDir, "test-epic.xml", true)

		var stdout, stderr bytes.Buffer
		app := setupTestApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err := app.Run(context.Background(), []string{"agentpm", "validate", "--file", epicPath})

		assert.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "✓ Epic validation passed")
	})

	t.Run("validate with invalid epic", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		epicPath := createTestEpic(tempDir, "invalid-epic.xml", false)

		var stdout, stderr bytes.Buffer
		app := setupTestApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err := app.Run(context.Background(), []string{"agentpm", "validate", "--file", epicPath})

		assert.Error(t, err)
		output := stdout.String()
		assert.Contains(t, output, "✗ Epic validation failed")
		assert.Contains(t, output, "Errors:")
	})

	t.Run("validate with non-existent epic file", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		var stdout, stderr bytes.Buffer
		app := setupTestApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err := app.Run(context.Background(), []string{"agentpm", "validate", "--file", "missing.xml"})

		assert.Error(t, err)
		output := stderr.String()
		assert.Contains(t, output, "Epic file not found")
	})

	t.Run("validate with missing config", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		var stdout, stderr bytes.Buffer
		app := setupTestApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err := app.Run(context.Background(), []string{"agentpm", "validate"})

		assert.Error(t, err)
		output := stderr.String()
		assert.Contains(t, output, "Failed to load configuration")
	})

	t.Run("validate with XML format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		epicPath := createTestEpic(tempDir, "test-epic.xml", true)

		var stdout, stderr bytes.Buffer
		app := setupTestApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err := app.Run(context.Background(), []string{"agentpm", "--format", "xml", "validate", "--file", epicPath})

		assert.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "<validation_result")
		assert.Contains(t, output, "<valid>true</valid>")
		assert.Contains(t, output, "<checks_performed>")
	})

	t.Run("validate with JSON format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		epicPath := createTestEpic(tempDir, "test-epic.xml", true)

		var stdout, stderr bytes.Buffer
		app := setupTestApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err := app.Run(context.Background(), []string{"agentpm", "--format", "json", "validate", "--file", epicPath})

		assert.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, `"valid": true`)
		assert.Contains(t, output, `"epic":`)
	})
}

func TestGetEpicName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"epic-8.xml", "epic-8"},
		{"path/to/epic-8.xml", "epic-8"},
		{"/abs/path/epic-8.xml", "epic-8"},
		{"epic-8", "epic-8"},
		{"epic.xml", "epic"},
		{"", ""},
	}

	for _, tt := range tests {
		result := getEpicName(tt.input)
		assert.Equal(t, tt.expected, result, "getEpicName(%s)", tt.input)
	}
}
