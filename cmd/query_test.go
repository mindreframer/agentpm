package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// Test helper function to create a simple epic for testing
func setupTestEpicForQuery(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	testEpic := &epic.Epic{
		ID:          "query-test-epic",
		Name:        "Query Test Epic",
		Status:      epic.StatusActive,
		Description: "Epic for testing query functionality",
		Assignee:    "test_agent",
		Phases: []epic.Phase{
			{ID: "10A", Name: "Core Query Engine", Status: epic.StatusCompleted},
			{ID: "10B", Name: "Query Tests", Status: epic.StatusActive},
		},
		Tasks: []epic.Task{
			{ID: "10A_1", PhaseID: "10A", Name: "Create QueryEngine", Status: epic.StatusCompleted},
			{ID: "10B_1", PhaseID: "10B", Name: "Write unit tests", Status: epic.StatusActive},
		},
		Tests: []epic.Test{
			{ID: "test_1", TaskID: "10A_1", Name: "XPath compilation test", Status: epic.StatusCompleted},
			{ID: "test_2", TaskID: "10B_1", Name: "Element selection test", Status: epic.StatusActive},
		},
	}

	storage := storage.NewFileStorage()
	epicPath := filepath.Join(tempDir, "test-epic.xml")
	err := storage.SaveEpic(testEpic, epicPath)
	require.NoError(t, err)

	return epicPath
}

func TestQueryCommandFixed(t *testing.T) {
	t.Run("query basic functionality", func(t *testing.T) {
		epicPath := setupTestEpicForQuery(t)

		var stdout bytes.Buffer
		cmd := QueryCommand()
		app := &cli.Command{
			Name:     "agentpm",
			Writer:   &stdout,
			Commands: []*cli.Command{cmd},
		}

		// Test basic query
		err := app.Run(context.Background(), []string{"agentpm", "query", "//task", "-f", epicPath, "--format", "text"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Query: //task")
		assert.Contains(t, output, "Found 2 matches")
		assert.Contains(t, output, "10A_1")
		assert.Contains(t, output, "10B_1")
	})

	t.Run("query attribute filtering", func(t *testing.T) {
		epicPath := setupTestEpicForQuery(t)

		var stdout bytes.Buffer
		cmd := QueryCommand()
		app := &cli.Command{
			Name:     "agentpm",
			Writer:   &stdout,
			Commands: []*cli.Command{cmd},
		}

		// Test attribute filtering
		err := app.Run(context.Background(), []string{"agentpm", "query", "//task[@status='completed']", "-f", epicPath, "--format", "text"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Query: //task[@status='completed']")
		assert.Contains(t, output, "Found 1 matches")
		assert.Contains(t, output, "10A_1")
		assert.NotContains(t, output, "10B_1")
	})

	t.Run("query JSON output", func(t *testing.T) {
		epicPath := setupTestEpicForQuery(t)

		var stdout bytes.Buffer
		cmd := QueryCommand()
		app := &cli.Command{
			Name:     "agentpm",
			Writer:   &stdout,
			Commands: []*cli.Command{cmd},
		}

		// Test JSON output
		err := app.Run(context.Background(), []string{"agentpm", "query", "//test", "-f", epicPath, "--format", "json"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "\"query\": \"//test\"")
		assert.Contains(t, output, "\"match_count\": 2")

		// Verify valid JSON
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)
	})

	t.Run("query empty results", func(t *testing.T) {
		epicPath := setupTestEpicForQuery(t)

		var stdout bytes.Buffer
		cmd := QueryCommand()
		app := &cli.Command{
			Name:     "agentpm",
			Writer:   &stdout,
			Commands: []*cli.Command{cmd},
		}

		// Test empty results
		err := app.Run(context.Background(), []string{"agentpm", "query", "//nonexistent", "-f", epicPath, "--format", "text"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Found 0 matches")
		assert.Contains(t, output, "No elements found matching query")
	})

	t.Run("query error handling", func(t *testing.T) {
		epicPath := setupTestEpicForQuery(t)

		cmd := QueryCommand()
		app := &cli.Command{
			Name:     "agentpm",
			Commands: []*cli.Command{cmd},
		}

		// Test missing XPath expression
		err := app.Run(context.Background(), []string{"agentpm", "query"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "XPath expression is required")

		// Test invalid output format
		err = app.Run(context.Background(), []string{"agentpm", "query", "//task", "-f", epicPath, "--format", "invalid"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid output format")

		// Test invalid XPath syntax
		err = app.Run(context.Background(), []string{"agentpm", "query", "//task[", "-f", epicPath})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid XPath query")

		// Test missing epic file
		err = app.Run(context.Background(), []string{"agentpm", "query", "//task", "-f", "nonexistent.xml"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "query execution failed")
	})
}
