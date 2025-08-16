package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mindreframer/agentpm/cmd"
	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// createRealApp creates the real CLI app with actual command implementations
func createRealApp() *cli.Command {
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
			cmd.InitCommand(),
			cmd.ConfigCommand(),
			cmd.ValidateCommand(),
			cmd.StatusCommand(),
			cmd.CurrentCommand(),
			cmd.PendingCommand(),
			cmd.FailingCommand(),
			cmd.EventsCommand(),
			cmd.StartEpicCommand(),
			cmd.DoneEpicCommand(),
			cmd.SwitchCommand(),
		},
	}
}

// Integration tests for the complete CLI workflow
func TestFullCLIWorkflow(t *testing.T) {
	t.Run("complete workflow: init -> config -> validate", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create a valid epic file
		epicPath := createValidEpic(tempDir, "workflow-epic.xml")

		// Test 1: Initialize project
		var stdout, stderr bytes.Buffer
		app := createRealApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err := app.Run(context.Background(), []string{"agentpm", "init", "--epic", epicPath})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "✓ Project initialized successfully")
		assert.Contains(t, output, epicPath)

		// Test 2: Check configuration
		stdout.Reset()
		stderr.Reset()

		err = app.Run(context.Background(), []string{"agentpm", "config"})
		require.NoError(t, err)

		output = stdout.String()
		assert.Contains(t, output, "Current Configuration:")
		assert.Contains(t, output, epicPath)
		assert.Contains(t, output, "agent")

		// Test 3: Validate epic
		stdout.Reset()
		stderr.Reset()

		err = app.Run(context.Background(), []string{"agentpm", "validate"})
		require.NoError(t, err)

		output = stdout.String()
		assert.Contains(t, output, "✓ Epic validation passed")

		// Test 4: Validate with file override
		stdout.Reset()
		stderr.Reset()

		err = app.Run(context.Background(), []string{"agentpm", "validate", "--file", epicPath})
		require.NoError(t, err)

		output = stdout.String()
		assert.Contains(t, output, "✓ Epic validation passed")
	})

	t.Run("error handling workflow", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		var stdout, stderr bytes.Buffer
		app := createRealApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		// Test 1: Initialize with non-existent epic
		err := app.Run(context.Background(), []string{"agentpm", "init", "--epic", "missing.xml"})
		require.Error(t, err)

		output := stderr.String()
		assert.Contains(t, output, "Epic file not found")

		// Test 2: Config without initialization
		stdout.Reset()
		stderr.Reset()

		err = app.Run(context.Background(), []string{"agentpm", "config"})
		require.Error(t, err)

		output = stderr.String()
		assert.Contains(t, output, "Failed to load configuration")

		// Test 3: Validate without initialization
		stdout.Reset()
		stderr.Reset()

		err = app.Run(context.Background(), []string{"agentpm", "validate"})
		require.Error(t, err)

		output = stderr.String()
		assert.Contains(t, output, "Failed to load configuration")
	})

	t.Run("multi-format output workflow", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		epicPath := createValidEpic(tempDir, "format-epic.xml")

		var stdout, stderr bytes.Buffer
		app := createRealApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		// Initialize with JSON format
		err := app.Run(context.Background(), []string{"agentpm", "--format", "json", "init", "--epic", epicPath})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, `"project_created": true`)
		assert.Contains(t, output, epicPath)

		// Config with XML format
		stdout.Reset()
		stderr.Reset()

		err = app.Run(context.Background(), []string{"agentpm", "--format", "xml", "config"})
		require.NoError(t, err)

		output = stdout.String()
		assert.Contains(t, output, "<config>")
		assert.Contains(t, output, "<current_epic>")

		// Validate with JSON format
		stdout.Reset()
		stderr.Reset()

		err = app.Run(context.Background(), []string{"agentpm", "--format", "json", "validate"})
		require.NoError(t, err)

		output = stdout.String()
		assert.Contains(t, output, `"valid": true`)
		assert.Contains(t, output, `"epic":`)
	})
}

func TestCLIPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	t.Run("command execution performance", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		epicPath := createValidEpic(tempDir, "perf-epic.xml")

		var stdout, stderr bytes.Buffer
		app := createRealApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		// Test init performance
		start := time.Now()
		err := app.Run(context.Background(), []string{"agentpm", "init", "--epic", epicPath})
		initDuration := time.Since(start)

		require.NoError(t, err)
		assert.Less(t, initDuration, 100*time.Millisecond, "Init command should complete in < 100ms")

		// Test config performance
		start = time.Now()
		err = app.Run(context.Background(), []string{"agentpm", "config"})
		configDuration := time.Since(start)

		require.NoError(t, err)
		assert.Less(t, configDuration, 50*time.Millisecond, "Config command should complete in < 50ms")

		// Test validate performance
		start = time.Now()
		err = app.Run(context.Background(), []string{"agentpm", "validate"})
		validateDuration := time.Since(start)

		require.NoError(t, err)
		assert.Less(t, validateDuration, 100*time.Millisecond, "Validate command should complete in < 100ms")

		t.Logf("Performance results: init=%v, config=%v, validate=%v",
			initDuration, configDuration, validateDuration)
	})

	t.Run("large epic file performance", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create a large epic with many phases, tasks, and tests
		epicPath := createLargeEpic(tempDir, "large-epic.xml")

		var stdout, stderr bytes.Buffer
		app := createRealApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		// Test initialization with large epic
		start := time.Now()
		err := app.Run(context.Background(), []string{"agentpm", "init", "--epic", epicPath})
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Less(t, duration, 500*time.Millisecond, "Large epic init should complete in < 500ms")

		// Test validation with large epic
		start = time.Now()
		err = app.Run(context.Background(), []string{"agentpm", "validate"})
		duration = time.Since(start)

		require.NoError(t, err)
		assert.Less(t, duration, 500*time.Millisecond, "Large epic validation should complete in < 500ms")

		t.Logf("Large epic performance: %v", duration)
	})
}

func TestCLIEdgeCases(t *testing.T) {
	t.Run("concurrent access", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		epicPath := createValidEpic(tempDir, "concurrent-epic.xml")

		// Initialize project first
		var stdout, stderr bytes.Buffer
		app := createRealApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err := app.Run(context.Background(), []string{"agentpm", "init", "--epic", epicPath})
		require.NoError(t, err)

		// Test concurrent reads (should be safe)
		done := make(chan bool, 3)

		for i := 0; i < 3; i++ {
			go func() {
				defer func() { done <- true }()

				var stdout, stderr bytes.Buffer
				app := createRealApp()
				app.Writer = &stdout
				app.ErrWriter = &stderr

				err := app.Run(context.Background(), []string{"agentpm", "config"})
				assert.NoError(t, err)

				err = app.Run(context.Background(), []string{"agentpm", "validate"})
				assert.NoError(t, err)
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 3; i++ {
			<-done
		}
	})

	t.Run("special characters in paths", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create directory with spaces and special characters
		specialDir := filepath.Join(tempDir, "special dir with spaces & symbols")
		err := os.MkdirAll(specialDir, 0755)
		require.NoError(t, err)

		epicPath := createValidEpic(specialDir, "special-epic.xml")

		var stdout, stderr bytes.Buffer
		app := createRealApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err = app.Run(context.Background(), []string{"agentpm", "init", "--epic", epicPath})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "✓ Project initialized successfully")
	})

	t.Run("very long epic file path", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create nested directory structure
		longPath := tempDir
		for i := 0; i < 10; i++ {
			longPath = filepath.Join(longPath, "very-long-directory-name-that-goes-on-and-on")
		}
		err := os.MkdirAll(longPath, 0755)
		require.NoError(t, err)

		epicPath := createValidEpic(longPath, "deep-epic.xml")

		var stdout, stderr bytes.Buffer
		app := createRealApp()
		app.Writer = &stdout
		app.ErrWriter = &stderr

		err = app.Run(context.Background(), []string{"agentpm", "init", "--epic", epicPath})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "✓ Project initialized successfully")
	})
}

func TestCLIOutputConsistency(t *testing.T) {
	t.Run("output format consistency", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		epicPath := createValidEpic(tempDir, "output-epic.xml")

		// Test that output is consistent across multiple runs
		results := make([]string, 3)

		for i := 0; i < 3; i++ {
			var stdout, stderr bytes.Buffer
			app := createRealApp()
			app.Writer = &stdout
			app.ErrWriter = &stderr

			err := app.Run(context.Background(), []string{"agentpm", "init", "--epic", epicPath})
			require.NoError(t, err)

			results[i] = stdout.String()
		}

		// All outputs should be identical (except for potential timestamps)
		for i := 1; i < len(results); i++ {
			assert.Equal(t, results[0], results[i], "Output should be consistent across runs")
		}
	})

	t.Run("error message consistency", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Test that error messages are consistent
		errorResults := make([]string, 3)

		for i := 0; i < 3; i++ {
			var stdout, stderr bytes.Buffer
			app := createRealApp()
			app.Writer = &stdout
			app.ErrWriter = &stderr

			err := app.Run(context.Background(), []string{"agentpm", "init", "--epic", "missing.xml"})
			require.Error(t, err)

			errorResults[i] = stderr.String()
		}

		// All error outputs should be identical
		for i := 1; i < len(errorResults); i++ {
			assert.Equal(t, errorResults[0], errorResults[i], "Error output should be consistent across runs")
		}
	})
}

// Helper functions for creating test epics
func createValidEpic(dir, filename string) string {
	epicPath := filepath.Join(dir, filename)

	testEpic := &epic.Epic{
		ID:        "integration-test-1",
		Name:      "Integration Test Epic",
		Status:    epic.StatusPlanning,
		CreatedAt: time.Date(2025, 8, 16, 9, 0, 0, 0, time.UTC),
		Assignee:  "test_agent",
		Phases: []epic.Phase{
			{ID: "P1", Name: "Phase 1", Status: epic.StatusPlanning},
			{ID: "P2", Name: "Phase 2", Status: epic.StatusPlanning},
		},
		Tasks: []epic.Task{
			{ID: "T1", PhaseID: "P1", Name: "Task 1", Status: epic.StatusPlanning},
			{ID: "T2", PhaseID: "P1", Name: "Task 2", Status: epic.StatusPlanning},
			{ID: "T3", PhaseID: "P2", Name: "Task 3", Status: epic.StatusPlanning},
		},
		Tests: []epic.Test{
			{ID: "TEST1", TaskID: "T1", Name: "Test 1", Status: epic.StatusPlanning},
			{ID: "TEST2", TaskID: "T2", Name: "Test 2", Status: epic.StatusPlanning},
			{ID: "TEST3", TaskID: "T3", Name: "Test 3", Status: epic.StatusPlanning},
		},
	}

	storage := storage.NewFileStorage()
	err := storage.SaveEpic(testEpic, epicPath)
	if err != nil {
		panic(err)
	}

	return epicPath
}

func createLargeEpic(dir, filename string) string {
	epicPath := filepath.Join(dir, filename)

	testEpic := &epic.Epic{
		ID:        "large-test-epic",
		Name:      "Large Test Epic",
		Status:    epic.StatusPlanning,
		CreatedAt: time.Date(2025, 8, 16, 9, 0, 0, 0, time.UTC),
		Assignee:  "test_agent",
	}

	// Create many phases
	for i := 1; i <= 50; i++ {
		testEpic.Phases = append(testEpic.Phases, epic.Phase{
			ID:     fmt.Sprintf("P%d", i),
			Name:   fmt.Sprintf("Phase %d", i),
			Status: epic.StatusPlanning,
		})
	}

	// Create many tasks
	for i := 1; i <= 200; i++ {
		phaseID := fmt.Sprintf("P%d", (i%50)+1)
		testEpic.Tasks = append(testEpic.Tasks, epic.Task{
			ID:      fmt.Sprintf("T%d", i),
			PhaseID: phaseID,
			Name:    fmt.Sprintf("Task %d", i),
			Status:  epic.StatusPlanning,
		})
	}

	// Create many tests
	for i := 1; i <= 200; i++ {
		taskID := fmt.Sprintf("T%d", i)
		testEpic.Tests = append(testEpic.Tests, epic.Test{
			ID:     fmt.Sprintf("TEST%d", i),
			TaskID: taskID,
			Name:   fmt.Sprintf("Test %d", i),
			Status: epic.StatusPlanning,
		})
	}

	storage := storage.NewFileStorage()
	err := storage.SaveEpic(testEpic, epicPath)
	if err != nil {
		panic(err)
	}

	return epicPath
}
