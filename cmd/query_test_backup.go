package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/mindreframer/agentpm/internal/xmlquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// createTestEpicForQuery creates a comprehensive test epic for query testing
func createTestEpicForQuery() *epic.Epic {
	return &epic.Epic{
		ID:          "query-test-epic",
		Name:        "Query Test Epic",
		Status:      epic.StatusActive,
		Description: "Epic for testing query functionality",
		Assignee:    "test_agent",
		Phases: []epic.Phase{
			{
				ID:          "10A",
				Name:        "Core Query Engine",
				Status:      epic.StatusCompleted,
				Description: "Setup XPath query engine",
			},
			{
				ID:          "10B",
				Name:        "Query Tests",
				Status:      epic.StatusActive,
				Description: "Write comprehensive tests",
			},
			{
				ID:          "10C",
				Name:        "CLI Integration",
				Status:      epic.StatusPlanning,
				Description: "Integrate query command with CLI",
			},
		},
		Tasks: []epic.Task{
			{
				ID:          "10A_1",
				PhaseID:     "10A",
				Name:        "Create QueryEngine interface",
				Status:      epic.StatusCompleted,
				Description: "Setup query engine with etree integration",
			},
			{
				ID:          "10A_2",
				PhaseID:     "10A",
				Name:        "Implement caching",
				Status:      epic.StatusCompleted,
				Description: "Add query compilation caching",
			},
			{
				ID:          "10B_1",
				PhaseID:     "10B",
				Name:        "Write unit tests",
				Status:      epic.StatusActive,
				Description: "Test XPath compilation and execution",
			},
			{
				ID:          "10C_1",
				PhaseID:     "10C",
				Name:        "Create CLI command",
				Status:      epic.StatusPlanning,
				Description: "Integrate with urfave/cli framework",
			},
		},
		Tests: []epic.Test{
			{
				ID:          "test_xpath_compilation",
				TaskID:      "10A_1",
				Name:        "XPath compilation test",
				Status:      epic.StatusCompleted,
				Description: "Verify XPath expressions compile correctly",
			},
			{
				ID:          "test_element_selection",
				TaskID:      "10A_1",
				Name:        "Element selection test",
				Status:      epic.StatusCompleted,
				Description: "Test basic element queries",
			},
			{
				ID:          "test_attribute_filtering",
				TaskID:      "10B_1",
				Name:        "Attribute filtering test",
				Status:      epic.StatusActive,
				Description: "Test attribute-based queries",
			},
		},
		Events: []epic.Event{
			{
				ID:   "event_1",
				Type: "phase_completed",
				Data: "Phase 10A completed",
			},
			{
				ID:   "event_2",
				Type: "task_started",
				Data: "Task 10B_1 started",
			},
		},
	}
}

func TestQueryCommand(t *testing.T) {
	t.Run("query command basic structure", func(t *testing.T) {
		cmd := QueryCommand()
		assert.Equal(t, "query", cmd.Name)
		assert.Equal(t, "Execute XPath queries against epic XML files", cmd.Usage)
		assert.Contains(t, cmd.Description, "XPath queries")
		assert.Contains(t, cmd.Description, "etree syntax")
		assert.Contains(t, cmd.Description, "Query patterns:")
		assert.Contains(t, cmd.Description, "//task")
		assert.Contains(t, cmd.Description, "Examples:")
	})

	t.Run("query basic element selection", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForQuery()
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

		// Test query command
		var stdout, stderr bytes.Buffer
		cmd := QueryCommand()
		app := &cli.Command{
			Name:     "agentpm",
			Writer:   &stdout,
			Commands: []*cli.Command{cmd},
		}

		// Execute query for all tasks
		err = app.Run(context.Background(), []string{"agentpm", "query", "//task", "--format", "text"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Query: //task")
		assert.Contains(t, output, "Found 4 matches")
		assert.Contains(t, output, "task[")
		assert.Contains(t, output, "10A_1")
		assert.Contains(t, output, "10A_2")
		assert.Contains(t, output, "10B_1")
		assert.Contains(t, output, "10C_1")

		// Verify all output is present
		assert.NotEqual(t, "", stdout.String())
		assert.Equal(t, "", stderr.String())
	})

	t.Run("query with attribute filtering", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForQuery()
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

		// Test attribute filtering query
		var stdout bytes.Buffer
		cmd := QueryCommand()
		app := &cli.Command{
			Name:     "agentpm",
			Writer:   &stdout,
			Commands: []*cli.Command{cmd},
		}

		// Execute query for completed tasks
		err = app.Run(context.Background(), []string{"agentpm", "query", "//task[@status='completed']", "-f", epicPath, "--format", "text"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Query: //task[@status='completed']")
		assert.Contains(t, output, "Found 2 matches")
		assert.Contains(t, output, "status=completed")
		assert.Contains(t, output, "10A_1")
		assert.Contains(t, output, "10A_2")
		// Should not contain planning or active tasks
		assert.NotContains(t, output, "10B_1")
		assert.NotContains(t, output, "10C_1")
	})

	t.Run("query with phase filtering", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForQuery()
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

		// Test phase-based filtering
		var stdout bytes.Buffer
		cmd := QueryCommand()
		app := &cli.Command{
			Name:     "agentpm",
			Writer:   &stdout,
			Commands: []*cli.Command{cmd},
		}

		// Execute query for tasks in phase 10A (using config since we're in temp dir)
		err = app.Run(context.Background(), []string{"agentpm", "query", "//task[@phase_id='10A']", "--format", "text"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Query: //task[@phase_id='10A']")
		assert.Contains(t, output, "Found 2 matches")
		assert.Contains(t, output, "phase_id=10A")
		assert.Contains(t, output, "10A_1")
		assert.Contains(t, output, "10A_2")
		// Should not contain tasks from other phases
		assert.NotContains(t, output, "10B_1")
		assert.NotContains(t, output, "10C_1")
	})

	t.Run("query XML output format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForQuery()
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

		// Test XML output
		var stdout bytes.Buffer
		cmd := QueryCommand()
		app := &cli.Command{
			Name:     "agentpm",
			Writer:   &stdout,
			Commands: []*cli.Command{cmd},
		}

		// Execute query with XML format
		err = app.Run(context.Background(), []string{"agentpm", "query", "//phase[@status='active']", "--format", "xml"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
		assert.Contains(t, output, "<query_result>")
		assert.Contains(t, output, "<query>//phase[@status=&#39;active&#39;]</query>")
		assert.Contains(t, output, "<match_count>1</match_count>")
		assert.Contains(t, output, "</query_result>")

		// Verify it's valid XML
		var result xmlquery.QueryResultXML
		xmlContent := strings.TrimPrefix(output, xml.Header)
		err = xml.Unmarshal([]byte(xmlContent), &result)
		require.NoError(t, err)
		assert.Equal(t, "//phase[@status='active']", result.Query)
		assert.Equal(t, 1, result.MatchCount)
	})

	t.Run("query JSON output format", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Create test epic
		testEpic := createTestEpicForQuery()
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

		// Test JSON output
		var stdout bytes.Buffer
		cmd := QueryCommand()
		app := &cli.Command{
			Name:     "agentpm",
			Writer:   &stdout,
			Commands: []*cli.Command{cmd},
		}

		// Execute query with JSON format
		err = app.Run(context.Background(), []string{"agentpm", "query", "//test[@status='completed']", "--format", "json"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "\"query\": \"//test[@status='completed']\"")
		assert.Contains(t, output, "\"match_count\": 2")

		// Verify it's valid JSON
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)
		assert.Equal(t, "//test[@status='completed']", result["query"])
		assert.Equal(t, float64(2), result["match_count"]) // JSON numbers are floats
	})

	t.Run("query empty results", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create test epic
		testEpic := createTestEpicForQuery()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Test query with no matches
		var stdout bytes.Buffer
		cmd := QueryCommand()
		app := &cli.Command{
			Name:     "agentpm",
			Writer:   &stdout,
			Commands: []*cli.Command{cmd},
		}

		// Execute query for non-existent elements
		err = app.Run(context.Background(), []string{"agentpm", "query", "//nonexistent", "-f", epicPath, "--format", "text"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Query: //nonexistent")
		assert.Contains(t, output, "Found 0 matches")
		assert.Contains(t, output, "No elements found matching query")
	})

	t.Run("query error cases", func(t *testing.T) {
		cmd := QueryCommand()
		app := &cli.Command{
			Name:     "agentpm",
			Commands: []*cli.Command{cmd},
		}

		t.Run("missing XPath expression", func(t *testing.T) {
			err := app.Run(context.Background(), []string{"agentpm", "query"})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "XPath expression is required")
		})

		t.Run("invalid output format", func(t *testing.T) {
			err := app.Run(context.Background(), []string{"agentpm", "query", "//task", "--format", "invalid"})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid output format")
		})

		t.Run("invalid XPath syntax", func(t *testing.T) {
			tempDir := t.TempDir()
			testEpic := createTestEpicForQuery()
			storage := storage.NewFileStorage()
			epicPath := filepath.Join(tempDir, "test-epic.xml")
			err := storage.SaveEpic(testEpic, epicPath)
			require.NoError(t, err)

			err = app.Run(context.Background(), []string{"agentpm", "query", "//task[", "-f", epicPath})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid XPath query")
		})

		t.Run("missing epic file", func(t *testing.T) {
			err := app.Run(context.Background(), []string{"agentpm", "query", "//task", "-f", "nonexistent.xml"})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "query execution failed")
		})
	})

	t.Run("query with file override", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create test epic
		testEpic := createTestEpicForQuery()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Create different config pointing to non-existent file
		cfg := &config.Config{
			CurrentEpic:     "nonexistent.xml",
			DefaultAssignee: "test_agent",
		}
		configPath := filepath.Join(tempDir, ".agentpm.json")
		err = config.SaveConfig(cfg, configPath)
		require.NoError(t, err)

		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		// Test that file override works
		var stdout bytes.Buffer
		cmd := QueryCommand()
		app := &cli.Command{
			Name:     "agentpm",
			Writer:   &stdout,
			Commands: []*cli.Command{cmd},
		}

		// Execute query with file override
		err = app.Run(context.Background(), []string{"agentpm", "query", "//epic", "-f", epicPath, "--format", "text"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Query: //epic")
		assert.Contains(t, output, "Found 1 matches")
		assert.Contains(t, output, "epic[")
	})

	t.Run("query complex XPath expressions", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create test epic
		testEpic := createTestEpicForQuery()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "test-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		var stdout bytes.Buffer
		cmd := QueryCommand()
		app := &cli.Command{
			Name:     "agentpm",
			Writer:   &stdout,
			Commands: []*cli.Command{cmd},
		}

		testCases := []struct {
			name        string
			query       string
			expectMatch bool
			contains    []string
		}{
			{
				name:        "wildcard selection",
				query:       "//epic/*",
				expectMatch: true,
				contains:    []string{"phases", "tasks", "tests"},
			},
			{
				name:        "nested element selection",
				query:       "//tasks/task",
				expectMatch: true,
				contains:    []string{"task["},
			},
			{
				name:        "multiple attribute filtering",
				query:       "//task[@phase_id='10A'][@status='completed']",
				expectMatch: true,
				contains:    []string{"10A_1", "10A_2"},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				stdout.Reset()

				err := app.Run(context.Background(), []string{"agentpm", "query", tc.query, "-f", epicPath, "--format", "text"})
				require.NoError(t, err)

				output := stdout.String()
				assert.Contains(t, output, fmt.Sprintf("Query: %s", tc.query))

				if tc.expectMatch {
					assert.NotContains(t, output, "Found 0 matches")
					for _, contain := range tc.contains {
						assert.Contains(t, output, contain)
					}
				} else {
					assert.Contains(t, output, "Found 0 matches")
				}
			})
		}
	})
}
