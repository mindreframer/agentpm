package xmlquery

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/beevik/etree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createSampleQueryResult creates a sample QueryResult for testing formatters
func createSampleQueryResult() *QueryResult {
	// Create a simple XML document for testing
	doc := etree.NewDocument()
	root := doc.CreateElement("epic")
	root.SetText("")

	// Add a task element
	task1 := root.CreateElement("task")
	task1.CreateAttr("id", "10A_1")
	task1.CreateAttr("status", "done")
	task1.SetText("Create QueryEngine interface")

	// Add another task element
	task2 := root.CreateElement("task")
	task2.CreateAttr("id", "10B_1")
	task2.CreateAttr("status", "wip")
	task2.SetText("Write unit tests for query engine")

	// Add a phase element
	phase := root.CreateElement("phase")
	phase.CreateAttr("id", "10A")
	phase.CreateAttr("name", "Core Query Engine")
	phase.CreateAttr("status", "completed")

	return &QueryResult{
		Query:           "//task",
		EpicFile:        "/test/epic.xml",
		MatchCount:      2,
		ExecutionTimeMs: 15,
		Elements:        []*etree.Element{task1, task2},
		Message:         "",
	}
}

func createEmptyQueryResult() *QueryResult {
	return &QueryResult{
		Query:           "//nonexistent",
		EpicFile:        "/test/epic.xml",
		MatchCount:      0,
		ExecutionTimeMs: 5,
		Elements:        []*etree.Element{},
		Message:         "No elements found matching query",
	}
}

func createAttributeQueryResult() *QueryResult {
	// Create elements with attributes for testing attribute queries
	doc := etree.NewDocument()
	root := doc.CreateElement("epic")

	task := root.CreateElement("task")
	task.CreateAttr("id", "10A_1")
	task.CreateAttr("status", "done")
	task.CreateAttr("phase_id", "10A")

	return &QueryResult{
		Query:           "//task/@status",
		EpicFile:        "/test/epic.xml",
		MatchCount:      1,
		ExecutionTimeMs: 8,
		Elements:        []*etree.Element{task},
		Message:         "",
	}
}

func TestNewFormatter(t *testing.T) {
	t.Run("XML formatter", func(t *testing.T) {
		formatter := NewFormatter(FormatXML)
		assert.IsType(t, &XMLFormatter{}, formatter)
	})

	t.Run("Text formatter", func(t *testing.T) {
		formatter := NewFormatter(FormatText)
		assert.IsType(t, &TextFormatter{}, formatter)
	})

	t.Run("JSON formatter", func(t *testing.T) {
		formatter := NewFormatter(FormatJSON)
		assert.IsType(t, &JSONFormatter{}, formatter)
	})

	t.Run("Default formatter", func(t *testing.T) {
		formatter := NewFormatter(OutputFormat("invalid"))
		assert.IsType(t, &XMLFormatter{}, formatter)
	})
}

func TestXMLFormatter_Format(t *testing.T) {
	formatter := &XMLFormatter{}

	t.Run("basic element results", func(t *testing.T) {
		result := createSampleQueryResult()
		output, err := formatter.Format(result)
		require.NoError(t, err)

		// Verify it's valid XML
		assert.Contains(t, output, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
		assert.Contains(t, output, "<query_result>")
		assert.Contains(t, output, "<query>//task</query>")
		assert.Contains(t, output, "<match_count>2</match_count>")
		assert.Contains(t, output, "<execution_time_ms>15</execution_time_ms>")
		assert.Contains(t, output, "<epic_file>/test/epic.xml</epic_file>")
		assert.Contains(t, output, "</query_result>")

		// Verify elements are included
		assert.Contains(t, output, "task")
		assert.Contains(t, output, "10A_1")
		assert.Contains(t, output, "done")

		// Verify XML is well-formed
		var xmlResult QueryResultXML
		err = xml.Unmarshal([]byte(strings.TrimPrefix(output, xml.Header)), &xmlResult)
		require.NoError(t, err)
		assert.Equal(t, "//task", xmlResult.Query)
		assert.Equal(t, 2, xmlResult.MatchCount)
	})

	t.Run("empty results", func(t *testing.T) {
		result := createEmptyQueryResult()
		output, err := formatter.Format(result)
		require.NoError(t, err)

		assert.Contains(t, output, "<match_count>0</match_count>")
		assert.Contains(t, output, "<message>No elements found matching query</message>")
	})

	t.Run("attribute query results", func(t *testing.T) {
		result := createAttributeQueryResult()
		output, err := formatter.Format(result)
		require.NoError(t, err)

		// For attribute queries, should detect /@
		assert.Contains(t, output, "<query>//task/@status</query>")
		// Should format as attributes when it's an attribute query
	})
}

func TestTextFormatter_Format(t *testing.T) {
	formatter := &TextFormatter{}

	t.Run("basic element results", func(t *testing.T) {
		result := createSampleQueryResult()
		output, err := formatter.Format(result)
		require.NoError(t, err)

		// Verify basic structure
		assert.Contains(t, output, "Query: //task")
		assert.Contains(t, output, "Found 2 matches")
		assert.Contains(t, output, "(executed in 15ms)")

		// Verify elements are formatted properly
		assert.Contains(t, output, "task[")
		assert.Contains(t, output, "id=10A_1")
		assert.Contains(t, output, "status=done")
		assert.Contains(t, output, "Create QueryEngine interface")

		// Verify second task
		assert.Contains(t, output, "id=10B_1")
		assert.Contains(t, output, "status=wip")
		assert.Contains(t, output, "Write unit tests")
	})

	t.Run("empty results", func(t *testing.T) {
		result := createEmptyQueryResult()
		output, err := formatter.Format(result)
		require.NoError(t, err)

		assert.Contains(t, output, "Query: //nonexistent")
		assert.Contains(t, output, "Found 0 matches")
		assert.Contains(t, output, "No elements found matching query")
	})

	t.Run("long text truncation", func(t *testing.T) {
		result := createSampleQueryResult()
		// Modify one element to have very long text
		longText := strings.Repeat("A very long description that should be truncated ", 5)
		result.Elements[0].SetText(longText)

		output, err := formatter.Format(result)
		require.NoError(t, err)

		// Should contain truncation marker
		assert.Contains(t, output, "...")
	})
}

func TestJSONFormatter_Format(t *testing.T) {
	formatter := &JSONFormatter{}

	t.Run("basic element results", func(t *testing.T) {
		result := createSampleQueryResult()
		output, err := formatter.Format(result)
		require.NoError(t, err)

		// Verify it's valid JSON
		var jsonResult QueryResultJSON
		err = json.Unmarshal([]byte(output), &jsonResult)
		require.NoError(t, err)

		assert.Equal(t, "//task", jsonResult.Query)
		assert.Equal(t, "/test/epic.xml", jsonResult.EpicFile)
		assert.Equal(t, 2, jsonResult.MatchCount)
		assert.Equal(t, 15, jsonResult.ExecutionTimeMs)
		assert.Equal(t, "", jsonResult.Message)

		// Verify matches structure
		assert.NotNil(t, jsonResult.Matches)

		// Verify JSON structure is readable
		assert.Contains(t, output, "\"query\": \"//task\"")
		assert.Contains(t, output, "\"match_count\": 2")
		assert.Contains(t, output, "\"execution_time_ms\": 15")
	})

	t.Run("empty results", func(t *testing.T) {
		result := createEmptyQueryResult()
		output, err := formatter.Format(result)
		require.NoError(t, err)

		var jsonResult QueryResultJSON
		err = json.Unmarshal([]byte(output), &jsonResult)
		require.NoError(t, err)

		assert.Equal(t, 0, jsonResult.MatchCount)
		assert.Equal(t, "No elements found matching query", jsonResult.Message)
	})

	t.Run("proper JSON formatting", func(t *testing.T) {
		result := createSampleQueryResult()
		output, err := formatter.Format(result)
		require.NoError(t, err)

		// Should be properly indented
		lines := strings.Split(output, "\n")
		assert.Greater(t, len(lines), 5) // Multiple lines for readability

		// Should contain proper JSON structure
		assert.Contains(t, output, "{")
		assert.Contains(t, output, "}")
		assert.Contains(t, output, ":")
	})
}

func TestFormatterQueryTypeDetection(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		isAttribute bool
		isText      bool
	}{
		{
			name:        "basic element query",
			query:       "//task",
			isAttribute: false,
			isText:      false,
		},
		{
			name:        "attribute query with @",
			query:       "//task/@status",
			isAttribute: true,
			isText:      false,
		},
		{
			name:        "attribute query with filter",
			query:       "//phase[@status='done']/@name",
			isAttribute: true,
			isText:      false,
		},
		{
			name:        "text query",
			query:       "//description/text()",
			isAttribute: false,
			isText:      true,
		},
		{
			name:        "complex element query",
			query:       "//task[@phase_id='10A']",
			isAttribute: false,
			isText:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xmlFormatter := &XMLFormatter{}
			textFormatter := &TextFormatter{}
			jsonFormatter := &JSONFormatter{}

			assert.Equal(t, tt.isAttribute, xmlFormatter.isAttributeQuery(tt.query))
			assert.Equal(t, tt.isAttribute, textFormatter.isAttributeQuery(tt.query))
			assert.Equal(t, tt.isAttribute, jsonFormatter.isAttributeQuery(tt.query))

			assert.Equal(t, tt.isText, xmlFormatter.isTextQuery(tt.query))
			assert.Equal(t, tt.isText, textFormatter.isTextQuery(tt.query))
			assert.Equal(t, tt.isText, jsonFormatter.isTextQuery(tt.query))
		})
	}
}

func TestService_QueryEpicFileFormatted(t *testing.T) {
	service := NewService()
	epicFile := createTestEpicXML(t)

	t.Run("XML format", func(t *testing.T) {
		output, err := service.QueryEpicFileFormatted(epicFile, "//task", FormatXML)
		require.NoError(t, err)

		assert.Contains(t, output, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
		assert.Contains(t, output, "<query_result>")
		assert.Contains(t, output, "<query>//task</query>")
		assert.Contains(t, output, "<match_count>4</match_count>")
	})

	t.Run("text format", func(t *testing.T) {
		output, err := service.QueryEpicFileFormatted(epicFile, "//task[@status='done']", FormatText)
		require.NoError(t, err)

		assert.Contains(t, output, "Query: //task[@status='done']")
		assert.Contains(t, output, "Found 2 matches")
		assert.Contains(t, output, "task[")
		assert.Contains(t, output, "status=done")
	})

	t.Run("JSON format", func(t *testing.T) {
		output, err := service.QueryEpicFileFormatted(epicFile, "//phase", FormatJSON)
		require.NoError(t, err)

		// Should be valid JSON
		var result interface{}
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)

		assert.Contains(t, output, "\"query\": \"//phase\"")
		assert.Contains(t, output, "\"match_count\":")
	})

	t.Run("invalid file", func(t *testing.T) {
		_, err := service.QueryEpicFileFormatted("nonexistent.xml", "//task", FormatXML)
		assert.Error(t, err)

		var fileErr *FileAccessError
		assert.ErrorAs(t, err, &fileErr)
	})

	t.Run("invalid query", func(t *testing.T) {
		_, err := service.QueryEpicFileFormatted(epicFile, "//task[", FormatText)
		assert.Error(t, err)

		var syntaxErr *QuerySyntaxError
		assert.ErrorAs(t, err, &syntaxErr)
	})
}

func TestFormatterIntegrationWithRealQueries(t *testing.T) {
	service := NewService()
	epicFile := createTestEpicXML(t)

	t.Run("element query formatted as XML", func(t *testing.T) {
		output, err := service.QueryEpicFileFormatted(epicFile, "//task[@phase_id='10A']", FormatXML)
		require.NoError(t, err)

		// Should be valid XML
		assert.Contains(t, output, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")

		// Should contain query details (XML-encoded single quotes)
		assert.Contains(t, output, "<query>//task[@phase_id=&#39;10A&#39;]</query>")
		assert.Contains(t, output, "<match_count>2</match_count>")

		// Should contain element data
		assert.Contains(t, output, "<matches>")
		assert.Contains(t, output, "</matches>")
	})

	t.Run("metadata query formatted as text", func(t *testing.T) {
		output, err := service.QueryEpicFileFormatted(epicFile, "//metadata/assignee", FormatText)
		require.NoError(t, err)

		assert.Contains(t, output, "Query: //metadata/assignee")
		assert.Contains(t, output, "Found 1 matches")
		assert.Contains(t, output, "assignee[")
		assert.Contains(t, output, "test_agent")
	})

	t.Run("phase query formatted as JSON", func(t *testing.T) {
		output, err := service.QueryEpicFileFormatted(epicFile, "//phase[@status='wip']", FormatJSON)
		require.NoError(t, err)

		var result QueryResultJSON
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)

		assert.Equal(t, "//phase[@status='wip']", result.Query)
		assert.Equal(t, 1, result.MatchCount)
		assert.NotNil(t, result.Matches)
	})
}

func TestFormatterErrorHandling(t *testing.T) {
	t.Run("XML formatter with nil elements", func(t *testing.T) {
		formatter := &XMLFormatter{}
		result := &QueryResult{
			Query:           "//test",
			MatchCount:      0,
			Elements:        []*etree.Element{},
			ExecutionTimeMs: 5,
		}

		output, err := formatter.Format(result)
		require.NoError(t, err)
		assert.Contains(t, output, "<match_count>0</match_count>")
	})

	t.Run("Text formatter with empty result", func(t *testing.T) {
		formatter := &TextFormatter{}
		result := createEmptyQueryResult()

		output, err := formatter.Format(result)
		require.NoError(t, err)
		assert.Contains(t, output, "Found 0 matches")
	})

	t.Run("JSON formatter with complex structure", func(t *testing.T) {
		formatter := &JSONFormatter{}
		result := createSampleQueryResult()

		output, err := formatter.Format(result)
		require.NoError(t, err)

		// Should still be valid JSON
		var jsonResult QueryResultJSON
		err = json.Unmarshal([]byte(output), &jsonResult)
		require.NoError(t, err)
	})
}
