package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/memomoo/agentpm/internal/config"
	"github.com/memomoo/agentpm/internal/epic"
	"github.com/memomoo/agentpm/internal/reports"
	"github.com/memomoo/agentpm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestEpicForDocs() *epic.Epic {
	return &epic.Epic{
		ID:          "docs-test-epic",
		Name:        "Docs Test Epic",
		Status:      epic.StatusActive,
		CreatedAt:   time.Date(2025, 8, 16, 9, 0, 0, 0, time.UTC),
		Assignee:    "test_agent",
		Description: "Test epic for documentation generation",
		Phases: []epic.Phase{
			{ID: "P1", Name: "Setup Phase", Status: epic.StatusCompleted},
			{ID: "P2", Name: "Implementation Phase", Status: epic.StatusActive},
			{ID: "P3", Name: "Testing Phase", Status: epic.StatusPlanning},
		},
		Tasks: []epic.Task{
			{ID: "T1", PhaseID: "P1", Name: "Setup Task", Status: epic.StatusCompleted, Assignee: "test_agent"},
			{ID: "T2", PhaseID: "P2", Name: "Active Task", Status: epic.StatusActive, Assignee: "test_agent"},
			{ID: "T3", PhaseID: "P2", Name: "Pending Task", Status: epic.StatusPlanning},
			{ID: "T4", PhaseID: "P3", Name: "Future Task", Status: epic.StatusPlanning},
		},
		Tests: []epic.Test{
			{ID: "TEST1", TaskID: "T1", Name: "Setup Test", Status: epic.StatusCompleted, TestStatus: epic.TestStatusPassed},
			{ID: "TEST2", TaskID: "T2", Name: "Active Test", Status: epic.StatusPlanning, TestStatus: epic.TestStatusFailed, FailureNote: "Connection timeout"},
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
				Data:      "Found integration issue",
			},
		},
	}
}

func TestDocsCommand(t *testing.T) {
	t.Run("markdown documentation generation", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForDocs()
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

		// Create docs command
		var stdout, stderr bytes.Buffer
		cmd := DocsCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command with markdown format
		err = cmd.Run(context.Background(), []string{"docs", "--format=markdown"})
		require.NoError(t, err)

		output := stdout.String()

		// Verify markdown structure
		assert.Contains(t, output, "# Docs Test Epic")
		assert.Contains(t, output, "## Epic Overview")
		assert.Contains(t, output, "- **ID:** docs-test-epic")
		assert.Contains(t, output, "- **Status:** 🔄 active")
		assert.Contains(t, output, "- **Assignee:** test_agent")
		assert.Contains(t, output, "**Description:** Test epic for documentation generation")

		// Verify phase progress section
		assert.Contains(t, output, "## Phase Progress")
		assert.Contains(t, output, "**Completed:** 1/3 phases")
		assert.Contains(t, output, "| Phase | Status | Tasks | Started | Completed |")
		assert.Contains(t, output, "| Setup Phase | ✅ completed | 1 |")
		assert.Contains(t, output, "| Implementation Phase | 🔄 active | 2 |")

		// Verify task status section
		assert.Contains(t, output, "## Task Status")
		assert.Contains(t, output, "**Completed:** 1/4 tasks")
		assert.Contains(t, output, "**Active Task:** T2")
		assert.Contains(t, output, "| Task | Phase | Status | Assignee | Started | Completed |")
		assert.Contains(t, output, "| Setup Task | P1 | ✅ completed | test_agent |")
		assert.Contains(t, output, "| Active Task | P2 | 🔄 active | test_agent |")

		// Verify test results section
		assert.Contains(t, output, "## Test Results")
		assert.Contains(t, output, "**Summary:** 1 passing, 2 failing (3 total)")
		assert.Contains(t, output, "| Test | Task | Status | Notes |")
		assert.Contains(t, output, "| Setup Test | T1 | ✅ passed | — |")
		assert.Contains(t, output, "| Active Test | T2 | ❌ failed | Connection timeout |")

		// Verify blockers section
		assert.Contains(t, output, "## Blockers")
		assert.Contains(t, output, "- 🚫 Failed test TEST2: Active Test")
		assert.Contains(t, output, "- 🚫 Found integration issue")

		// Verify recent activity section
		assert.Contains(t, output, "## Recent Activity")
		assert.Contains(t, output, "**2025-08-16 11:00** (blocker): Found integration issue")
		assert.Contains(t, output, "**2025-08-16 10:00** (task_completed): Setup task completed")

		// Verify footer
		assert.Contains(t, output, "*Generated on")
		assert.Contains(t, output, "by AgentPM*")
	})

	t.Run("JSON documentation generation", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForDocs()
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

		// Create docs command
		var stdout, stderr bytes.Buffer
		cmd := DocsCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command with JSON format
		err = cmd.Run(context.Background(), []string{"docs", "--format=json"})
		require.NoError(t, err)

		output := stdout.String()

		// Parse JSON to verify structure
		var report reports.DocumentationReport
		err = json.Unmarshal([]byte(output), &report)
		require.NoError(t, err)

		// Verify epic overview
		assert.Equal(t, "docs-test-epic", report.EpicOverview.ID)
		assert.Equal(t, "Docs Test Epic", report.EpicOverview.Name)
		assert.Equal(t, "active", report.EpicOverview.Status)
		assert.Equal(t, "test_agent", report.EpicOverview.Assignee)
		assert.Equal(t, "Test epic for documentation generation", report.EpicOverview.Description)

		// Verify phase progress
		assert.Equal(t, 3, report.PhaseProgress.TotalPhases)
		assert.Equal(t, 1, report.PhaseProgress.CompletedPhases)
		assert.Len(t, report.PhaseProgress.Phases, 3)
		assert.Equal(t, "P1", report.PhaseProgress.Phases[0].ID)
		assert.Equal(t, "completed", report.PhaseProgress.Phases[0].Status)

		// Verify task status
		assert.Equal(t, 4, report.TaskStatus.TotalTasks)
		assert.Equal(t, 1, report.TaskStatus.CompletedTasks)
		assert.Equal(t, "T2", report.TaskStatus.ActiveTask)
		assert.Len(t, report.TaskStatus.Tasks, 4)

		// Verify test results
		assert.Equal(t, 3, report.TestResults.TotalTests)
		assert.Equal(t, 1, report.TestResults.PassingTests)
		assert.Equal(t, 2, report.TestResults.FailingTests)
		assert.Len(t, report.TestResults.Tests, 3)

		// Verify recent activity
		assert.Len(t, report.RecentActivity.Events, 3)
		assert.Len(t, report.RecentActivity.Blockers, 2)
	})

	t.Run("output to file - markdown", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForDocs()
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

		// Create docs command
		var stdout, stderr bytes.Buffer
		cmd := DocsCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		outputFile := filepath.Join(tempDir, "docs", "epic-status.md")

		// Execute command with file output
		err = cmd.Run(context.Background(), []string{"docs", "--format=markdown", "--output=" + outputFile})
		require.NoError(t, err)

		// Verify success message
		output := stdout.String()
		assert.Contains(t, output, "Documentation generated: "+outputFile)

		// Verify file was created
		assert.FileExists(t, outputFile)

		// Verify file content
		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		fileContent := string(content)

		assert.Contains(t, fileContent, "# Docs Test Epic")
		assert.Contains(t, fileContent, "## Epic Overview")
		assert.Contains(t, fileContent, "**Description:** Test epic for documentation generation")
	})

	t.Run("output to file - JSON", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic and config
		testEpic := createTestEpicForDocs()
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

		// Create docs command
		var stdout, stderr bytes.Buffer
		cmd := DocsCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		outputFile := filepath.Join(tempDir, "docs", "epic-status.json")

		// Execute command with file output
		err = cmd.Run(context.Background(), []string{"docs", "--format=json", "--output=" + outputFile})
		require.NoError(t, err)

		// Verify success message
		output := stdout.String()
		assert.Contains(t, output, "Documentation generated: "+outputFile)

		// Verify file was created
		assert.FileExists(t, outputFile)

		// Verify file content is valid JSON
		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)

		var report reports.DocumentationReport
		err = json.Unmarshal(content, &report)
		require.NoError(t, err)

		assert.Equal(t, "docs-test-epic", report.EpicOverview.ID)
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

		// Create docs command
		var stdout, stderr bytes.Buffer
		cmd := DocsCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command - should fail
		err = cmd.Run(context.Background(), []string{"docs"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load epic")
	})

	t.Run("file override flag", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForDocs()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "specific-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create config with different epic
		cfg := &config.Config{
			CurrentEpic:     "other-epic.xml",
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		// Create docs command
		var stdout, stderr bytes.Buffer
		cmd := DocsCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command with file override
		err = cmd.Run(context.Background(), []string{"docs", "--file=" + epicPath, "--format=markdown"})
		require.NoError(t, err)

		output := stdout.String()

		// Should use the overridden file, not the config file
		assert.Contains(t, output, "# Docs Test Epic")
		assert.Contains(t, output, "docs-test-epic")
	})
}

func TestMarkdownGeneration(t *testing.T) {
	t.Run("comprehensive markdown formatting", func(t *testing.T) {
		storage := storage.NewMemoryStorage()
		testEpic := createTestEpicForDocs()
		err := storage.SaveEpic(testEpic, "test.xml")
		require.NoError(t, err)

		rs := reports.NewReportService(storage)
		err = rs.LoadEpic("test.xml")
		require.NoError(t, err)

		markdown, err := rs.GenerateMarkdownDocumentation()
		require.NoError(t, err)

		// Verify markdown structure
		lines := strings.Split(markdown, "\n")

		// Should start with H1 title
		assert.True(t, strings.HasPrefix(lines[0], "# "))

		// Should have proper sections
		assert.Contains(t, markdown, "## Epic Overview")
		assert.Contains(t, markdown, "## Phase Progress")
		assert.Contains(t, markdown, "## Task Status")
		assert.Contains(t, markdown, "## Test Results")
		assert.Contains(t, markdown, "## Blockers")
		assert.Contains(t, markdown, "## Recent Activity")

		// Should have proper table formatting
		assert.Contains(t, markdown, "| Phase | Status | Tasks | Started | Completed |")
		assert.Contains(t, markdown, "|-------|--------|-------|---------|----------|")

		// Should have status icons
		assert.Contains(t, markdown, "✅ completed")
		assert.Contains(t, markdown, "🔄 active")
		assert.Contains(t, markdown, "⏳ planning")
		assert.Contains(t, markdown, "❌ failed")

		// Should have footer
		assert.Contains(t, markdown, "*Generated on")
		assert.Contains(t, markdown, "by AgentPM*")
	})
}
