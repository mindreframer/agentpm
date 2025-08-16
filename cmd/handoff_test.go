package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/memomoo/agentpm/internal/config"
	"github.com/memomoo/agentpm/internal/epic"
	"github.com/memomoo/agentpm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestEpicForHandoff() *epic.Epic {
	return &epic.Epic{
		ID:        "handoff-test-epic",
		Name:      "Handoff Test Epic",
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
			{ID: "TEST1", TaskID: "T1", Name: "Setup Test", Status: epic.StatusCompleted, TestStatus: epic.TestStatusPassed},
			{ID: "TEST2", TaskID: "T2", Name: "Active Test", Status: epic.StatusPlanning, TestStatus: epic.TestStatusFailed},
			{ID: "TEST3", TaskID: "T3", Name: "Pending Test", Status: epic.StatusPlanning, TestStatus: epic.TestStatusPending},
		},
		Events: []epic.Event{
			{
				ID:        "E1",
				Type:      "created",
				Timestamp: time.Date(2025, 8, 16, 9, 0, 0, 0, time.UTC),
				Data:      "Epic created",
			},
			{
				ID:        "E2",
				Type:      "task_completed",
				Timestamp: time.Date(2025, 8, 16, 10, 0, 0, 0, time.UTC),
				Data:      "Setup task completed",
			},
			{
				ID:        "E3",
				Type:      "blocker",
				Timestamp: time.Date(2025, 8, 16, 11, 0, 0, 0, time.UTC),
				Data:      "Found dependency issue",
			},
		},
	}
}

func createCompletedEpicForHandoff() *epic.Epic {
	return &epic.Epic{
		ID:        "completed-handoff-epic",
		Name:      "Completed Handoff Epic",
		Status:    epic.StatusCompleted,
		CreatedAt: time.Date(2025, 8, 15, 9, 0, 0, 0, time.UTC),
		Assignee:  "test_agent",
		Phases: []epic.Phase{
			{ID: "P1", Name: "Phase 1", Status: epic.StatusCompleted},
			{ID: "P2", Name: "Phase 2", Status: epic.StatusCompleted},
		},
		Tasks: []epic.Task{
			{ID: "T1", PhaseID: "P1", Name: "Task 1", Status: epic.StatusCompleted},
			{ID: "T2", PhaseID: "P2", Name: "Task 2", Status: epic.StatusCompleted},
		},
		Tests: []epic.Test{
			{ID: "TEST1", TaskID: "T1", Name: "Test 1", Status: epic.StatusCompleted, TestStatus: epic.TestStatusPassed},
			{ID: "TEST2", TaskID: "T2", Name: "Test 2", Status: epic.StatusCompleted, TestStatus: epic.TestStatusPassed},
		},
		Events: []epic.Event{
			{
				ID:        "E1",
				Type:      "created",
				Timestamp: time.Date(2025, 8, 15, 9, 0, 0, 0, time.UTC),
				Data:      "Epic created",
			},
			{
				ID:        "E2",
				Type:      "epic_completed",
				Timestamp: time.Date(2025, 8, 16, 17, 0, 0, 0, time.UTC),
				Data:      "Epic completed successfully",
			},
		},
	}
}

func TestHandoffCommand(t *testing.T) {
	t.Run("comprehensive handoff report - text format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForHandoff()
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

		// Create handoff command
		var stdout, stderr bytes.Buffer
		cmd := HandoffCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command with text format
		err = cmd.Run(context.Background(), []string{"handoff", "--format=text"})
		require.NoError(t, err)

		output := stdout.String()

		// Verify epic info
		assert.Contains(t, output, "=== AGENT HANDOFF REPORT ===")
		assert.Contains(t, output, "Epic: Handoff Test Epic")
		assert.Contains(t, output, "ID: handoff-test-epic")
		assert.Contains(t, output, "Status: active")
		assert.Contains(t, output, "Assignee: test_agent")

		// Verify current state
		assert.Contains(t, output, "CURRENT STATE:")
		assert.Contains(t, output, "Active Phase: P2")
		assert.Contains(t, output, "Active Task: T2")
		assert.Contains(t, output, "Next Action:")

		// Verify progress summary
		assert.Contains(t, output, "PROGRESS SUMMARY:")
		assert.Contains(t, output, "Phases: 1/3 completed")
		assert.Contains(t, output, "Tests: 1 passing, 2 failing")

		// Verify blockers
		assert.Contains(t, output, "BLOCKERS:")
		assert.Contains(t, output, "Failed test TEST2")
		assert.Contains(t, output, "Found dependency issue")

		// Verify recent events
		assert.Contains(t, output, "RECENT EVENTS:")
		assert.Contains(t, output, "blocker: Found dependency issue")
		assert.Contains(t, output, "task_completed: Setup task completed")
	})

	t.Run("comprehensive handoff report - XML format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForHandoff()
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

		// Create handoff command
		var stdout, stderr bytes.Buffer
		cmd := HandoffCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command with XML format
		err = cmd.Run(context.Background(), []string{"handoff", "--format=xml"})
		require.NoError(t, err)

		output := stdout.String()

		// Verify XML structure
		assert.Contains(t, output, `<?xml version="1.0" encoding="UTF-8"?>`)
		assert.Contains(t, output, `<handoff epic="handoff-test-epic"`)
		assert.Contains(t, output, `<epic_info>`)
		assert.Contains(t, output, `<name>Handoff Test Epic</name>`)
		assert.Contains(t, output, `<status>active</status>`)
		assert.Contains(t, output, `<assignee>test_agent</assignee>`)
		assert.Contains(t, output, `<current_state>`)
		assert.Contains(t, output, `<active_phase>P2</active_phase>`)
		assert.Contains(t, output, `<active_task>T2</active_task>`)
		assert.Contains(t, output, `<summary>`)
		assert.Contains(t, output, `<completed_phases>1</completed_phases>`)
		assert.Contains(t, output, `<total_phases>3</total_phases>`)
		assert.Contains(t, output, `<passing_tests>1</passing_tests>`)
		assert.Contains(t, output, `<failing_tests>2</failing_tests>`)
		assert.Contains(t, output, `<recent_events`)
		assert.Contains(t, output, `<blockers>`)
		assert.Contains(t, output, `<blocker>Failed test TEST2`)
		assert.Contains(t, output, `<blocker>Found dependency issue`)
	})

	t.Run("handoff report for completed epic", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create completed epic and config
		testEpic := createCompletedEpicForHandoff()
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

		// Create handoff command
		var stdout, stderr bytes.Buffer
		cmd := HandoffCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		err = cmd.Run(context.Background(), []string{"handoff", "--format=text"})
		require.NoError(t, err)

		output := stdout.String()

		// Verify completed epic state
		assert.Contains(t, output, "Status: completed")
		assert.Contains(t, output, "Phases: 2/2 completed")
		assert.Contains(t, output, "Tests: 2 passing, 0 failing")

		// Should not have blockers section for completed epic
		assert.NotContains(t, output, "BLOCKERS:")
	})

	t.Run("recent events limit", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForHandoff()
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

		// Create handoff command
		var stdout, stderr bytes.Buffer
		cmd := HandoffCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command with limit=2
		err = cmd.Run(context.Background(), []string{"handoff", "--format=xml", "--limit=2"})
		require.NoError(t, err)

		output := stdout.String()

		// Verify limit is applied
		assert.Contains(t, output, `<recent_events limit="2">`)

		// Count event elements
		eventCount := strings.Count(output, "<event ")
		assert.Equal(t, 2, eventCount)
	})

	t.Run("blocker identification", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForHandoff()
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

		// Create handoff command
		var stdout, stderr bytes.Buffer
		cmd := HandoffCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		err = cmd.Run(context.Background(), []string{"handoff", "--format=text"})
		require.NoError(t, err)

		output := stdout.String()

		// Verify blockers are identified
		assert.Contains(t, output, "BLOCKERS:")
		assert.Contains(t, output, "Failed test TEST2: Active Test")
		assert.Contains(t, output, "Found dependency issue")
	})

	t.Run("handoff with no recent activity", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create epic with no events
		testEpic := &epic.Epic{
			ID:        "no-activity-epic",
			Name:      "No Activity Epic",
			Status:    epic.StatusPlanning,
			CreatedAt: time.Date(2025, 8, 16, 9, 0, 0, 0, time.UTC),
			Assignee:  "test_agent",
			Phases:    []epic.Phase{{ID: "P1", Name: "Phase 1", Status: epic.StatusPlanning}},
			Tasks:     []epic.Task{{ID: "T1", PhaseID: "P1", Name: "Task 1", Status: epic.StatusPlanning}},
			Tests:     []epic.Test{},
			Events:    []epic.Event{}, // No events
		}

		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "no-activity-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		cfg := &config.Config{
			CurrentEpic:     epicPath,
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		// Create handoff command
		var stdout, stderr bytes.Buffer
		cmd := HandoffCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		err = cmd.Run(context.Background(), []string{"handoff", "--format=text"})
		require.NoError(t, err)

		output := stdout.String()

		// Should still generate report without errors
		assert.Contains(t, output, "=== AGENT HANDOFF REPORT ===")
		assert.Contains(t, output, "No Activity Epic")

		// Should not have blockers or recent events sections if empty
		assert.NotContains(t, output, "BLOCKERS:")
		assert.NotContains(t, output, "RECENT EVENTS:")
	})

	t.Run("missing epic file", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create config with non-existent epic
		cfg := &config.Config{
			CurrentEpic:     "missing-epic.xml",
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err := config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		// Create handoff command
		var stdout, stderr bytes.Buffer
		cmd := HandoffCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command - should fail
		err = cmd.Run(context.Background(), []string{"handoff"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load epic")
	})
}
