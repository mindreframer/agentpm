package xmlquery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestEpicXML creates a sample epic XML file for testing
func createTestEpicXML(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")

	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<epic id="epic-10" name="XML Query System" status="active">
    <metadata>
        <assignee>test_agent</assignee>
        <created_at>2025-08-17T10:00:00Z</created_at>
        <priority>medium</priority>
    </metadata>
    <outline>
        <phase id="10A" name="Core Query Engine" status="completed">
            <description>Setup XPath query engine with etree integration</description>
        </phase>
        <phase id="10B" name="Query Engine Tests" status="active">
            <description>Write comprehensive tests for query functionality</description>
        </phase>
        <phase id="10C" name="Result Formatting" status="pending">
            <description>Implement XML, text, and JSON output formatters</description>
        </phase>
    </outline>
    <tasks>
        <task id="10A_1" phase_id="10A" status="done">
            <description>Create QueryEngine interface and implementation</description>
            <acceptance_criteria>
                - XPath expressions compile successfully
                - Query execution returns proper results
                - Error handling for invalid syntax
            </acceptance_criteria>
        </task>
        <task id="10A_2" phase_id="10A" status="done">
            <description>Implement query caching for performance</description>
            <acceptance_criteria>
                - Compiled queries are cached
                - Cache eviction works correctly
                - Performance improvement measurable
            </acceptance_criteria>
        </task>
        <task id="10B_1" phase_id="10B" status="active">
            <description>Write unit tests for query engine</description>
            <acceptance_criteria>
                - Test XPath compilation
                - Test element selection
                - Test attribute filtering
            </acceptance_criteria>
        </task>
        <task id="10C_1" phase_id="10C" status="pending">
            <description>Create output formatters</description>
            <acceptance_criteria>
                - XML format structured correctly
                - Text format human readable
                - JSON format valid
            </acceptance_criteria>
        </task>
    </tasks>
    <tests>
        <test id="test_xpath_compilation" phase_id="10A" task_id="10A_1" status="passing">
            <description>Verify XPath expressions compile without errors</description>
            <expected_outcome>Valid XPath expressions should compile successfully</expected_outcome>
        </test>
        <test id="test_element_selection" phase_id="10A" task_id="10A_1" status="passing">
            <description>Test basic element selection queries</description>
            <expected_outcome>Simple queries like //task should return all task elements</expected_outcome>
        </test>
        <test id="test_attribute_filtering" phase_id="10B" task_id="10B_1" status="pending">
            <description>Test attribute-based filtering</description>
            <expected_outcome>Queries with attribute filters should work correctly</expected_outcome>
        </test>
    </tests>
    <events>
        <event type="phase_started" timestamp="2025-08-17T09:00:00Z">
            <data>Phase 10A started</data>
        </event>
        <event type="task_completed" timestamp="2025-08-17T10:30:00Z">
            <data>Task 10A_1 completed</data>
        </event>
        <event type="phase_completed" timestamp="2025-08-17T11:00:00Z">
            <data>Phase 10A completed</data>
        </event>
    </events>
</epic>`

	err := os.WriteFile(epicFile, []byte(xmlContent), 0644)
	require.NoError(t, err)

	return epicFile
}

func TestNewEngine(t *testing.T) {
	engine := NewEngine()
	assert.NotNil(t, engine)
	assert.NotNil(t, engine.cache)
	assert.Equal(t, 0, engine.cache.Size())
}

func TestEngine_LoadDocument(t *testing.T) {
	t.Run("successful load", func(t *testing.T) {
		engine := NewEngine()
		epicFile := createTestEpicXML(t)

		err := engine.LoadDocument(epicFile)
		require.NoError(t, err)

		assert.NotNil(t, engine.GetDocument())
		assert.Equal(t, epicFile, engine.filePath)
	})

	t.Run("file not found", func(t *testing.T) {
		engine := NewEngine()

		err := engine.LoadDocument("nonexistent.xml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load XML document")
	})

	t.Run("invalid XML", func(t *testing.T) {
		engine := NewEngine()
		tempDir := t.TempDir()
		invalidFile := filepath.Join(tempDir, "invalid.xml")

		err := os.WriteFile(invalidFile, []byte("<invalid>unclosed tag"), 0644)
		require.NoError(t, err)

		err = engine.LoadDocument(invalidFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load XML document")
	})
}

func TestEngine_ValidateQuery(t *testing.T) {
	engine := NewEngine()

	t.Run("valid queries", func(t *testing.T) {
		validQueries := []string{
			"//task",
			"//phase[@status='active']",
			"//task[@phase_id='10A']",
			"//metadata/assignee",
			"//test[@id='test_1']",
			"/epic/tasks/task",
			"//description/text()",
			"//phase[1]",
			"//epic/*",
		}

		for _, query := range validQueries {
			t.Run(query, func(t *testing.T) {
				err := engine.ValidateQuery(query)
				assert.NoError(t, err, "Query should be valid: %s", query)
			})
		}
	})

	t.Run("invalid queries", func(t *testing.T) {
		invalidQueries := []string{
			"",                 // empty query
			"//task[",          // unclosed bracket
			"//task[@status='", // unclosed quote
			// Removed "//task[status=active]" as etree accepts some unquoted values
		}

		for _, query := range invalidQueries {
			t.Run(query, func(t *testing.T) {
				err := engine.ValidateQuery(query)
				assert.Error(t, err, "Query should be invalid: %s", query)

				if err != nil {
					var syntaxErr *QuerySyntaxError
					if assert.ErrorAs(t, err, &syntaxErr) {
						assert.Equal(t, query, syntaxErr.Query)
						assert.NotEmpty(t, syntaxErr.Message)
					}
				}
			})
		}
	})
}

func TestEngine_Execute(t *testing.T) {
	engine := NewEngine()
	epicFile := createTestEpicXML(t)

	err := engine.LoadDocument(epicFile)
	require.NoError(t, err)

	t.Run("basic element selection", func(t *testing.T) {
		result, err := engine.Execute("//task")
		require.NoError(t, err)

		assert.Equal(t, "//task", result.Query)
		assert.Equal(t, epicFile, result.EpicFile)
		assert.Equal(t, 4, result.MatchCount) // 4 tasks in test XML
		assert.GreaterOrEqual(t, result.ExecutionTimeMs, 0)
		assert.Len(t, result.Elements, 4)

		// Verify all elements are tasks
		for _, elem := range result.Elements {
			assert.Equal(t, "task", elem.Tag)
		}
	})

	t.Run("attribute filtering", func(t *testing.T) {
		result, err := engine.Execute("//task[@status='done']")
		require.NoError(t, err)

		assert.Equal(t, 2, result.MatchCount) // 2 done tasks
		assert.Len(t, result.Elements, 2)

		// Verify all tasks have status="done"
		for _, elem := range result.Elements {
			assert.Equal(t, "task", elem.Tag)
			status := elem.SelectAttrValue("status", "")
			assert.Equal(t, "done", status)
		}
	})

	t.Run("phase filtering", func(t *testing.T) {
		result, err := engine.Execute("//task[@phase_id='10A']")
		require.NoError(t, err)

		assert.Equal(t, 2, result.MatchCount) // 2 tasks in phase 10A
		assert.Len(t, result.Elements, 2)

		// Verify all tasks belong to phase 10A
		for _, elem := range result.Elements {
			phaseID := elem.SelectAttrValue("phase_id", "")
			assert.Equal(t, "10A", phaseID)
		}
	})

	t.Run("nested element selection", func(t *testing.T) {
		result, err := engine.Execute("//metadata/assignee")
		require.NoError(t, err)

		assert.Equal(t, 1, result.MatchCount) // 1 assignee element
		assert.Len(t, result.Elements, 1)

		assignee := result.Elements[0]
		assert.Equal(t, "assignee", assignee.Tag)
		assert.Equal(t, "test_agent", assignee.Text())
	})

	t.Run("text content selection", func(t *testing.T) {
		result, err := engine.Execute("//metadata/assignee/text()")
		require.NoError(t, err)

		// Note: text() queries return text nodes, which might behave differently
		// This test verifies the query executes without error
		assert.Equal(t, "//metadata/assignee/text()", result.Query)
		assert.GreaterOrEqual(t, result.MatchCount, 0)
	})

	t.Run("position-based selection", func(t *testing.T) {
		result, err := engine.Execute("//task[1]")
		require.NoError(t, err)

		// Should return first task element
		assert.GreaterOrEqual(t, result.MatchCount, 1)
		if result.MatchCount > 0 {
			assert.Equal(t, "task", result.Elements[0].Tag)
		}
	})

	t.Run("wildcard selection", func(t *testing.T) {
		result, err := engine.Execute("//outline/*")
		require.NoError(t, err)

		assert.Equal(t, 3, result.MatchCount) // 3 phase elements in outline
		assert.Len(t, result.Elements, 3)

		// Verify all are phase elements
		for _, elem := range result.Elements {
			assert.Equal(t, "phase", elem.Tag)
		}
	})

	t.Run("complex attribute filtering", func(t *testing.T) {
		result, err := engine.Execute("//phase[@status='active']")
		require.NoError(t, err)

		assert.Equal(t, 1, result.MatchCount) // 1 active phase
		assert.Len(t, result.Elements, 1)

		phase := result.Elements[0]
		assert.Equal(t, "phase", phase.Tag)
		assert.Equal(t, "active", phase.SelectAttrValue("status", ""))
		assert.Equal(t, "10B", phase.SelectAttrValue("id", ""))
	})

	t.Run("empty result", func(t *testing.T) {
		result, err := engine.Execute("//nonexistent")
		require.NoError(t, err)

		assert.Equal(t, 0, result.MatchCount)
		assert.Empty(t, result.Elements)
		assert.True(t, result.IsEmpty())
	})

	t.Run("no document loaded", func(t *testing.T) {
		emptyEngine := NewEngine()

		_, err := emptyEngine.Execute("//task")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no document loaded")
	})

	t.Run("invalid query execution", func(t *testing.T) {
		_, err := engine.Execute("//task[")
		assert.Error(t, err)

		var syntaxErr *QuerySyntaxError
		assert.ErrorAs(t, err, &syntaxErr)
		assert.Equal(t, "//task[", syntaxErr.Query)
	})
}

func TestEngine_QueryCaching(t *testing.T) {
	engine := NewEngine()
	epicFile := createTestEpicXML(t)

	err := engine.LoadDocument(epicFile)
	require.NoError(t, err)

	t.Run("cache stores compiled queries", func(t *testing.T) {
		query := "//task[@status='done']"

		// First execution should compile and cache
		result1, err := engine.Execute(query)
		require.NoError(t, err)
		assert.Equal(t, 1, engine.cache.Size())

		// Second execution should use cached version
		result2, err := engine.Execute(query)
		require.NoError(t, err)

		// Results should be identical
		assert.Equal(t, result1.MatchCount, result2.MatchCount)
		assert.Equal(t, result1.Query, result2.Query)

		// Cache should still have one entry
		assert.Equal(t, 1, engine.cache.Size())
	})

	t.Run("different queries create separate cache entries", func(t *testing.T) {
		// Clear cache first
		engine.cache.Clear()
		assert.Equal(t, 0, engine.cache.Size())

		queries := []string{
			"//task",
			"//phase",
			"//test",
		}

		for i, query := range queries {
			_, err := engine.Execute(query)
			require.NoError(t, err)
			assert.Equal(t, i+1, engine.cache.Size())
		}
	})
}

func TestQueryResult_HelperMethods(t *testing.T) {
	engine := NewEngine()
	epicFile := createTestEpicXML(t)

	err := engine.LoadDocument(epicFile)
	require.NoError(t, err)

	t.Run("IsEmpty method", func(t *testing.T) {
		// Non-empty result
		result, err := engine.Execute("//task")
		require.NoError(t, err)
		assert.False(t, result.IsEmpty())

		// Empty result
		emptyResult, err := engine.Execute("//nonexistent")
		require.NoError(t, err)
		assert.True(t, emptyResult.IsEmpty())
	})

	t.Run("GetElementTexts method", func(t *testing.T) {
		result, err := engine.Execute("//metadata/assignee")
		require.NoError(t, err)

		texts := result.GetElementTexts()
		assert.Len(t, texts, 1)
		assert.Equal(t, "test_agent", texts[0])
	})

	t.Run("GetAttributeValues method", func(t *testing.T) {
		result, err := engine.Execute("//task")
		require.NoError(t, err)

		statuses := result.GetAttributeValues("status")
		assert.Len(t, statuses, 4) // 4 tasks with status attributes
		assert.Contains(t, statuses, "done")
		assert.Contains(t, statuses, "active")
		assert.Contains(t, statuses, "pending")
	})

	t.Run("GetElementsByTag method", func(t *testing.T) {
		// Get all elements under tasks
		result, err := engine.Execute("//tasks/*")
		require.NoError(t, err)

		tasks := result.GetElementsByTag("task")
		assert.Len(t, tasks, 4) // All should be task elements

		// Verify all are indeed task elements
		for _, task := range tasks {
			assert.Equal(t, "task", task.Tag)
		}
	})
}
