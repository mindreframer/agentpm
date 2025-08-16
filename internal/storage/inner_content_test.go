package storage

import (
	"os"
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInnerXMLContentPreservation(t *testing.T) {
	t.Run("test description with inner XML is preserved", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := tempDir + "/test-epic.xml"

		// Create epic with test containing inner XML content
		testEpic := &epic.Epic{
			ID:        "test-epic",
			Name:      "Test Epic",
			Status:    epic.StatusActive,
			CreatedAt: time.Now(),
			Tests: []epic.Test{
				{
					ID:          "test-1",
					TaskID:      "task-1",
					PhaseID:     "phase-1",
					Name:        "Complex Test",
					Status:      epic.StatusActive,
					TestStatus:  epic.TestStatusWIP,
					Description: "This test has <code>inner XML</code> content with <em>markup</em>",
					FailureNote: "Failure with <strong>bold text</strong> and <link>references</link>",
				},
			},
		}

		fs := NewFileStorage()

		// Save the epic
		err := fs.SaveEpic(testEpic, epicFile)
		require.NoError(t, err)

		// Read the raw XML to verify it contains the inner content
		xmlContent, err := os.ReadFile(epicFile)
		require.NoError(t, err)
		xmlStr := string(xmlContent)

		assert.Contains(t, xmlStr, "<code>inner XML</code>", "Raw XML should contain inner markup")
		assert.Contains(t, xmlStr, "<em>markup</em>", "Raw XML should contain emphasis markup")
		assert.Contains(t, xmlStr, "<strong>bold text</strong>", "Raw XML should contain strong markup")
		assert.Contains(t, xmlStr, "<link>references</link>", "Raw XML should contain link markup")

		// Load the epic back
		loadedEpic, err := fs.LoadEpic(epicFile)
		require.NoError(t, err)
		require.Len(t, loadedEpic.Tests, 1)

		loadedTest := loadedEpic.Tests[0]

		// Verify the inner XML content is preserved (newlines may be present due to XML formatting)
		assert.Contains(t, loadedTest.Description, "<code>inner XML</code>", "Description should contain inner XML markup")
		assert.Contains(t, loadedTest.Description, "<em>markup</em>", "Description should contain emphasis markup")
		assert.Contains(t, loadedTest.FailureNote, "<strong>bold text</strong>", "Failure note should contain strong markup")
		assert.Contains(t, loadedTest.FailureNote, "<link>references</link>", "Failure note should contain link markup")

		// Also verify the basic text content is preserved
		assert.Contains(t, loadedTest.Description, "This test has", "Description should contain original text")
		assert.Contains(t, loadedTest.FailureNote, "Failure with", "Failure note should contain original text")
	})

	t.Run("epic description with inner XML is preserved", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := tempDir + "/test-epic.xml"

		// Create epic with description containing inner XML content
		testEpic := &epic.Epic{
			ID:           "test-epic",
			Name:         "Test Epic",
			Status:       epic.StatusActive,
			CreatedAt:    time.Now(),
			Description:  "Epic with <code>code samples</code> and <ul><li>list items</li></ul>",
			Workflow:     "Process includes <step>validation</step> and <step>testing</step>",
			Requirements: "Must support <feature>advanced parsing</feature> capabilities",
			Dependencies: "Depends on <lib>etree</lib> version 1.2+",
		}

		fs := NewFileStorage()

		// Save the epic
		err := fs.SaveEpic(testEpic, epicFile)
		require.NoError(t, err)

		// Load the epic back
		loadedEpic, err := fs.LoadEpic(epicFile)
		require.NoError(t, err)

		// Verify the inner XML content is preserved (note: XML parsing may add formatting)
		assert.Contains(t, loadedEpic.Description, "<code>code samples</code>", "Description should contain code markup")
		assert.Contains(t, loadedEpic.Description, "<ul>", "Description should contain list markup")
		assert.Contains(t, loadedEpic.Description, "<li>list items</li>", "Description should contain list item markup")
		assert.Contains(t, loadedEpic.Description, "</ul>", "Description should contain closing list markup")

		assert.Contains(t, loadedEpic.Workflow, "<step>validation</step>", "Workflow should contain step markup")
		assert.Contains(t, loadedEpic.Workflow, "<step>testing</step>", "Workflow should contain step markup")

		assert.Contains(t, loadedEpic.Requirements, "<feature>advanced parsing</feature>", "Requirements should contain feature markup")

		assert.Contains(t, loadedEpic.Dependencies, "<lib>etree</lib>", "Dependencies should contain lib markup")
	})

	t.Run("phase and task descriptions with inner XML are preserved", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := tempDir + "/test-epic.xml"

		// Create epic with phases and tasks containing inner XML content
		testEpic := &epic.Epic{
			ID:        "test-epic",
			Name:      "Test Epic",
			Status:    epic.StatusActive,
			CreatedAt: time.Now(),
			Phases: []epic.Phase{
				{
					ID:          "phase-1",
					Name:        "Setup Phase",
					Status:      epic.StatusActive,
					Description: "Setup includes <config>configuration</config> and <init>initialization</init>",
				},
			},
			Tasks: []epic.Task{
				{
					ID:          "task-1",
					PhaseID:     "phase-1",
					Name:        "Setup Task",
					Status:      epic.StatusActive,
					Description: "Task requires <tool>specific tools</tool> and <env>environment setup</env>",
				},
			},
		}

		fs := NewFileStorage()

		// Save the epic
		err := fs.SaveEpic(testEpic, epicFile)
		require.NoError(t, err)

		// Load the epic back
		loadedEpic, err := fs.LoadEpic(epicFile)
		require.NoError(t, err)

		// Verify the inner XML content is preserved
		require.Len(t, loadedEpic.Phases, 1)
		assert.Contains(t, loadedEpic.Phases[0].Description, "<config>configuration</config>", "Phase description should contain config markup")
		assert.Contains(t, loadedEpic.Phases[0].Description, "<init>initialization</init>", "Phase description should contain init markup")

		require.Len(t, loadedEpic.Tasks, 1)
		assert.Contains(t, loadedEpic.Tasks[0].Description, "<tool>specific tools</tool>", "Task description should contain tool markup")
		assert.Contains(t, loadedEpic.Tasks[0].Description, "<env>environment setup</env>", "Task description should contain env markup")
	})

	t.Run("plain text content without XML markup works normally", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := tempDir + "/test-epic.xml"

		// Create epic with plain text content (no XML markup)
		testEpic := &epic.Epic{
			ID:        "test-epic",
			Name:      "Test Epic",
			Status:    epic.StatusActive,
			CreatedAt: time.Now(),
			Tests: []epic.Test{
				{
					ID:          "test-1",
					TaskID:      "task-1",
					Name:        "Simple Test",
					Status:      epic.StatusActive,
					Description: "This is plain text without any markup",
				},
			},
		}

		fs := NewFileStorage()

		// Save the epic
		err := fs.SaveEpic(testEpic, epicFile)
		require.NoError(t, err)

		// Load the epic back
		loadedEpic, err := fs.LoadEpic(epicFile)
		require.NoError(t, err)

		// Verify plain text content is preserved correctly
		require.Len(t, loadedEpic.Tests, 1)
		assert.Equal(t, "This is plain text without any markup", loadedEpic.Tests[0].Description)
	})

	t.Run("empty content is handled correctly", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := tempDir + "/test-epic.xml"

		// Create epic with empty content
		testEpic := &epic.Epic{
			ID:        "test-epic",
			Name:      "Test Epic",
			Status:    epic.StatusActive,
			CreatedAt: time.Now(),
			Tests: []epic.Test{
				{
					ID:          "test-1",
					TaskID:      "task-1",
					Name:        "Empty Test",
					Status:      epic.StatusActive,
					Description: "",
					FailureNote: "",
				},
			},
		}

		fs := NewFileStorage()

		// Save the epic
		err := fs.SaveEpic(testEpic, epicFile)
		require.NoError(t, err)

		// Load the epic back
		loadedEpic, err := fs.LoadEpic(epicFile)
		require.NoError(t, err)

		// Verify empty content is handled correctly
		require.Len(t, loadedEpic.Tests, 1)
		assert.Equal(t, "", loadedEpic.Tests[0].Description)
		assert.Equal(t, "", loadedEpic.Tests[0].FailureNote)
	})

	t.Run("test with inner text format is preserved", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := tempDir + "/test-epic.xml"

		// Create epic with test in inner text format (like the user's example)
		testEpic := &epic.Epic{
			ID:        "test-epic",
			Name:      "Test Epic",
			Status:    epic.StatusActive,
			CreatedAt: time.Now(),
			Tests: []epic.Test{
				{
					ID:          "1A_1",
					PhaseID:     "1A",
					Name:        "Pagination Test",
					Status:      epic.StatusPending,
					Description: "**GIVEN** I have 100 schools in the database\n**WHEN** I call list_schools_paginated with page=2, page_size=25\n**THEN** I get schools 26-50 with pagination metadata",
				},
			},
		}

		fs := NewFileStorage()

		// Save the epic
		err := fs.SaveEpic(testEpic, epicFile)
		require.NoError(t, err)

		// Read the raw XML to verify it contains the inner text format
		xmlContent, err := os.ReadFile(epicFile)
		require.NoError(t, err)
		xmlStr := string(xmlContent)

		// Should contain the test content as inner text, not in a description element
		assert.Contains(t, xmlStr, "**GIVEN** I have 100 schools", "Raw XML should contain test content as inner text")
		assert.Contains(t, xmlStr, "**WHEN** I call list_schools_paginated", "Raw XML should contain test content as inner text")
		assert.Contains(t, xmlStr, "**THEN** I get schools 26-50", "Raw XML should contain test content as inner text")

		// Should NOT contain a description element for simple tests
		assert.NotContains(t, xmlStr, "<description>", "Simple tests should not use description element format")

		// Load the epic back
		loadedEpic, err := fs.LoadEpic(epicFile)
		require.NoError(t, err)
		require.Len(t, loadedEpic.Tests, 1)

		loadedTest := loadedEpic.Tests[0]

		// Verify the test content is preserved exactly
		assert.Equal(t, "1A_1", loadedTest.ID)
		assert.Equal(t, "1A", loadedTest.PhaseID)
		assert.Equal(t, epic.StatusPending, loadedTest.Status)
		assert.Contains(t, loadedTest.Description, "**GIVEN** I have 100 schools", "Test description should be preserved")
		assert.Contains(t, loadedTest.Description, "**WHEN** I call list_schools_paginated", "Test description should be preserved")
		assert.Contains(t, loadedTest.Description, "**THEN** I get schools 26-50", "Test description should be preserved")
	})

	t.Run("test with additional fields uses description element", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := tempDir + "/test-epic.xml"

		now := time.Now()
		// Create epic with test that has additional fields (timestamps, notes)
		testEpic := &epic.Epic{
			ID:        "test-epic",
			Name:      "Test Epic",
			Status:    epic.StatusActive,
			CreatedAt: time.Now(),
			Tests: []epic.Test{
				{
					ID:          "test-1",
					PhaseID:     "phase-1",
					Name:        "Complex Test",
					Status:      epic.StatusActive,
					TestStatus:  epic.TestStatusWIP,
					Description: "**GIVEN** test setup\n**WHEN** action\n**THEN** result",
					StartedAt:   &now,
					FailureNote: "Some failure note",
				},
			},
		}

		fs := NewFileStorage()

		// Save the epic
		err := fs.SaveEpic(testEpic, epicFile)
		require.NoError(t, err)

		// Read the raw XML to verify it uses description element format
		xmlContent, err := os.ReadFile(epicFile)
		require.NoError(t, err)
		xmlStr := string(xmlContent)

		// Should contain description element for tests with additional fields
		assert.Contains(t, xmlStr, "<description>", "Tests with additional fields should use description element")
		assert.Contains(t, xmlStr, "<started_at>", "Should contain timestamp fields")
		assert.Contains(t, xmlStr, "<failure_note>", "Should contain note fields")

		// Load the epic back
		loadedEpic, err := fs.LoadEpic(epicFile)
		require.NoError(t, err)
		require.Len(t, loadedEpic.Tests, 1)

		loadedTest := loadedEpic.Tests[0]

		// Verify all fields are preserved
		assert.Equal(t, "test-1", loadedTest.ID)
		assert.Contains(t, loadedTest.Description, "**GIVEN** test setup", "Description should be preserved")
		assert.Equal(t, "Some failure note", loadedTest.FailureNote)
		assert.NotNil(t, loadedTest.StartedAt)
	})
}
