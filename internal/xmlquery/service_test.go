package xmlquery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	service := NewService()
	assert.NotNil(t, service)
	assert.NotNil(t, service.engine)
}

func TestService_QueryEpicFile(t *testing.T) {
	service := NewService()

	t.Run("successful query", func(t *testing.T) {
		epicFile := createTestEpicXML(t)

		result, err := service.QueryEpicFile(epicFile, "//task")
		require.NoError(t, err)

		assert.Equal(t, "//task", result.Query)
		assert.Equal(t, epicFile, result.EpicFile)
		assert.Equal(t, 4, result.MatchCount)
		assert.Empty(t, result.Message) // No message for non-empty results
	})

	t.Run("empty result with message", func(t *testing.T) {
		epicFile := createTestEpicXML(t)

		result, err := service.QueryEpicFile(epicFile, "//nonexistent")
		require.NoError(t, err)

		assert.Equal(t, 0, result.MatchCount)
		assert.True(t, result.IsEmpty())
		assert.Equal(t, "No elements found matching query", result.Message)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := service.QueryEpicFile("nonexistent.xml", "//task")
		assert.Error(t, err)

		var fileErr *FileAccessError
		assert.ErrorAs(t, err, &fileErr)
		assert.Equal(t, "nonexistent.xml", fileErr.FilePath)
		assert.Contains(t, fileErr.Message, "does not exist")
	})

	t.Run("invalid XML file", func(t *testing.T) {
		tempDir := t.TempDir()
		invalidFile := filepath.Join(tempDir, "invalid.xml")

		err := os.WriteFile(invalidFile, []byte("<invalid>unclosed"), 0644)
		require.NoError(t, err)

		_, err = service.QueryEpicFile(invalidFile, "//task")
		assert.Error(t, err)

		var fileErr *FileAccessError
		assert.ErrorAs(t, err, &fileErr)
		assert.Equal(t, invalidFile, fileErr.FilePath)
		assert.Contains(t, fileErr.Message, "failed to load")
	})

	t.Run("invalid query syntax", func(t *testing.T) {
		epicFile := createTestEpicXML(t)

		_, err := service.QueryEpicFile(epicFile, "//task[")
		assert.Error(t, err)

		var syntaxErr *QuerySyntaxError
		assert.ErrorAs(t, err, &syntaxErr)
		assert.Equal(t, "//task[", syntaxErr.Query)
	})
}

func TestService_ValidateQuery(t *testing.T) {
	service := NewService()

	t.Run("valid queries", func(t *testing.T) {
		validQueries := []string{
			"//task",
			"//phase[@status='done']",
			"//task[@phase_id='1A']",
			"//metadata/assignee",
			"//description/text()",
		}

		for _, query := range validQueries {
			t.Run(query, func(t *testing.T) {
				err := service.ValidateQuery(query)
				assert.NoError(t, err)
			})
		}
	})

	t.Run("invalid queries", func(t *testing.T) {
		invalidQueries := []string{
			"",
			"//task[",
			"//task[@status='",
		}

		for _, query := range invalidQueries {
			t.Run(query, func(t *testing.T) {
				err := service.ValidateQuery(query)
				assert.Error(t, err)

				var syntaxErr *QuerySyntaxError
				assert.ErrorAs(t, err, &syntaxErr)
			})
		}
	})
}

func TestService_GetSupportedPatterns(t *testing.T) {
	service := NewService()

	patterns := service.GetSupportedPatterns()
	assert.NotEmpty(t, patterns)
	assert.Greater(t, len(patterns), 5)

	// Verify some expected patterns are included
	expectedPatterns := []string{
		"//task",
		"//phase[@status='done']",
		"//task[@phase_id='1A']",
		"//metadata/assignee",
		"//description/text()",
	}

	for _, expected := range expectedPatterns {
		assert.Contains(t, patterns, expected, "Pattern should be included: %s", expected)
	}
}

func TestService_IntegrationWithRealEpicStructure(t *testing.T) {
	service := NewService()
	epicFile := createTestEpicXML(t)

	t.Run("query epic metadata", func(t *testing.T) {
		result, err := service.QueryEpicFile(epicFile, "//metadata/assignee/text()")
		require.NoError(t, err)

		// Even if text() doesn't return elements, the query should execute
		assert.Equal(t, "//metadata/assignee/text()", result.Query)
	})

	t.Run("query tasks by phase", func(t *testing.T) {
		result, err := service.QueryEpicFile(epicFile, "//task[@phase_id='10A']")
		require.NoError(t, err)

		assert.Equal(t, 2, result.MatchCount)
		assert.Len(t, result.Elements, 2)

		// Verify both tasks belong to phase 10A
		for _, elem := range result.Elements {
			phaseID := elem.SelectAttrValue("phase_id", "")
			assert.Equal(t, "10A", phaseID)
		}
	})

	t.Run("query tests by status", func(t *testing.T) {
		result, err := service.QueryEpicFile(epicFile, "//test[@status='passing']")
		require.NoError(t, err)

		assert.Equal(t, 2, result.MatchCount)
		assert.Len(t, result.Elements, 2)

		// Verify all tests have passing status
		for _, elem := range result.Elements {
			status := elem.SelectAttrValue("status", "")
			assert.Equal(t, "passing", status)
		}
	})

	t.Run("query events by type", func(t *testing.T) {
		result, err := service.QueryEpicFile(epicFile, "//event[@type='task_completed']")
		require.NoError(t, err)

		assert.Equal(t, 1, result.MatchCount)
		assert.Len(t, result.Elements, 1)

		event := result.Elements[0]
		assert.Equal(t, "event", event.Tag)
		assert.Equal(t, "task_completed", event.SelectAttrValue("type", ""))
	})

	t.Run("query all active elements", func(t *testing.T) {
		result, err := service.QueryEpicFile(epicFile, "//*[@status='wip']")
		require.NoError(t, err)

		// Should find epic, phase, and task with active status
		assert.GreaterOrEqual(t, result.MatchCount, 2) // at least epic and phase

		// Verify all have active status
		for _, elem := range result.Elements {
			status := elem.SelectAttrValue("status", "")
			assert.Equal(t, "wip", status)
		}
	})

	t.Run("query nested descriptions", func(t *testing.T) {
		result, err := service.QueryEpicFile(epicFile, "//task/description")
		require.NoError(t, err)

		assert.Equal(t, 4, result.MatchCount) // 4 tasks with descriptions
		assert.Len(t, result.Elements, 4)

		// Verify all are description elements with content
		for _, elem := range result.Elements {
			assert.Equal(t, "description", elem.Tag)
			assert.NotEmpty(t, elem.Text())
		}
	})
}
