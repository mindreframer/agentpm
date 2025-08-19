package storage

import (
	"os"
	"strings"
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
			Status:    epic.StatusWIP,
			CreatedAt: time.Now(),
			Tests: []epic.Test{
				{
					ID:          "test-1",
					TaskID:      "task-1",
					PhaseID:     "phase-1",
					Name:        "Complex Test",
					Status:      epic.StatusWIP,
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
			Status:       epic.StatusWIP,
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
			Status:    epic.StatusWIP,
			CreatedAt: time.Now(),
			Phases: []epic.Phase{
				{
					ID:          "phase-1",
					Name:        "Setup Phase",
					Status:      epic.StatusWIP,
					Description: "Setup includes <config>configuration</config> and <init>initialization</init>",
				},
			},
			Tasks: []epic.Task{
				{
					ID:          "task-1",
					PhaseID:     "phase-1",
					Name:        "Setup Task",
					Status:      epic.StatusWIP,
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
			Status:    epic.StatusWIP,
			CreatedAt: time.Now(),
			Tests: []epic.Test{
				{
					ID:          "test-1",
					TaskID:      "task-1",
					Name:        "Simple Test",
					Status:      epic.StatusWIP,
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
			Status:    epic.StatusWIP,
			CreatedAt: time.Now(),
			Tests: []epic.Test{
				{
					ID:          "test-1",
					TaskID:      "task-1",
					Name:        "Empty Test",
					Status:      epic.StatusWIP,
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
			Status:    epic.StatusWIP,
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
			Status:    epic.StatusWIP,
			CreatedAt: time.Now(),
			Tests: []epic.Test{
				{
					ID:          "test-1",
					PhaseID:     "phase-1",
					Name:        "Complex Test",
					Status:      epic.StatusWIP,
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

func TestPhaseDeliverables(t *testing.T) {
	tests := []struct {
		name         string
		deliverables string
		description  string
	}{
		{
			name:         "phase with plain text deliverables",
			deliverables: "Database schema, API endpoints, Documentation",
			description:  "Simple text deliverables",
		},
		{
			name:         "phase with markup deliverables",
			deliverables: "<ul><li>Database migration scripts</li><li>REST API with <strong>authentication</strong></li><li>Unit tests with 90% coverage</li></ul>",
			description:  "Deliverables with XML markup",
		},
		{
			name:         "phase with multiline deliverables",
			deliverables: "1. Complete database design\n2. Implement core API\n3. Write comprehensive tests\n4. Deploy to staging environment",
			description:  "Multi-line deliverables",
		},
		{
			name:         "phase with empty deliverables",
			deliverables: "",
			description:  "Empty deliverables field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			epicFile := tempDir + "/test-epic.xml"

			// Create epic with phase containing deliverables
			testEpic := &epic.Epic{
				ID:        "test-epic",
				Name:      "Test Epic",
				Status:    epic.StatusWIP,
				CreatedAt: time.Now(),
				Phases: []epic.Phase{
					{
						ID:           "phase-1",
						Name:         "Setup Phase",
						Description:  "Initial setup phase",
						Deliverables: tt.deliverables,
						Status:       epic.StatusPending,
					},
				},
			}

			fs := NewFileStorage()

			// Save the epic
			err := fs.SaveEpic(testEpic, epicFile)
			require.NoError(t, err)

			// Read the raw XML to verify deliverables are saved correctly
			xmlContent, err := os.ReadFile(epicFile)
			require.NoError(t, err)
			xmlStr := string(xmlContent)

			if tt.deliverables != "" {
				assert.Contains(t, xmlStr, "<deliverables>", "XML should contain deliverables element")
				// For markup content, check individual elements rather than exact string
				if strings.Contains(tt.deliverables, "<ul>") {
					assert.Contains(t, xmlStr, "<ul>", "XML should contain ul element")
					assert.Contains(t, xmlStr, "<li>Database migration scripts</li>", "XML should contain list items")
					assert.Contains(t, xmlStr, "<strong>authentication</strong>", "XML should contain strong markup")
				} else {
					assert.Contains(t, xmlStr, tt.deliverables, "XML should contain deliverables content")
				}
			} else {
				assert.NotContains(t, xmlStr, "<deliverables>", "XML should not contain empty deliverables element")
			}

			// Load the epic back
			loadedEpic, err := fs.LoadEpic(epicFile)
			require.NoError(t, err)
			require.Len(t, loadedEpic.Phases, 1)

			loadedPhase := loadedEpic.Phases[0]

			// Verify deliverables are preserved (allow for XML formatting differences)
			if strings.Contains(tt.deliverables, "<ul>") {
				// For markup content, check that key elements are present
				assert.Contains(t, loadedPhase.Deliverables, "<ul>", "Deliverables should contain ul element")
				assert.Contains(t, loadedPhase.Deliverables, "<li>Database migration scripts</li>", "Deliverables should contain list items")
				assert.Contains(t, loadedPhase.Deliverables, "<strong>authentication</strong>", "Deliverables should contain markup")
			} else {
				assert.Equal(t, tt.deliverables, loadedPhase.Deliverables, "Deliverables should be preserved")
			}
			assert.Equal(t, "phase-1", loadedPhase.ID)
			assert.Equal(t, "Setup Phase", loadedPhase.Name)
			assert.Equal(t, "Initial setup phase", loadedPhase.Description)
		})
	}
}

func TestTaskAcceptanceCriteria(t *testing.T) {
	tests := []struct {
		name               string
		acceptanceCriteria string
		description        string
	}{
		{
			name:               "task with plain text acceptance criteria",
			acceptanceCriteria: "User can successfully login, Database connection is established, Response time < 200ms",
			description:        "Simple text acceptance criteria",
		},
		{
			name:               "task with markup acceptance criteria",
			acceptanceCriteria: "<ol><li>User enters <em>valid credentials</em></li><li>System returns <strong>JWT token</strong></li><li>User is redirected to dashboard</li></ol>",
			description:        "Acceptance criteria with XML markup",
		},
		{
			name:               "task with multiline acceptance criteria",
			acceptanceCriteria: "GIVEN a user with valid credentials\nWHEN they submit the login form\nTHEN they should be authenticated\nAND redirected to the dashboard\nAND see their profile information",
			description:        "Multi-line BDD-style acceptance criteria",
		},
		{
			name:               "task with empty acceptance criteria",
			acceptanceCriteria: "",
			description:        "Empty acceptance criteria field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			epicFile := tempDir + "/test-epic.xml"

			// Create epic with task containing acceptance criteria
			testEpic := &epic.Epic{
				ID:        "test-epic",
				Name:      "Test Epic",
				Status:    epic.StatusWIP,
				CreatedAt: time.Now(),
				Phases: []epic.Phase{
					{
						ID:     "phase-1",
						Name:   "Development Phase",
						Status: epic.StatusWIP,
					},
				},
				Tasks: []epic.Task{
					{
						ID:                 "task-1",
						PhaseID:            "phase-1",
						Name:               "User Authentication",
						Description:        "Implement user authentication system",
						AcceptanceCriteria: tt.acceptanceCriteria,
						Status:             epic.StatusPending,
					},
				},
			}

			fs := NewFileStorage()

			// Save the epic
			err := fs.SaveEpic(testEpic, epicFile)
			require.NoError(t, err)

			// Read the raw XML to verify acceptance criteria are saved correctly
			xmlContent, err := os.ReadFile(epicFile)
			require.NoError(t, err)
			xmlStr := string(xmlContent)

			if tt.acceptanceCriteria != "" {
				assert.Contains(t, xmlStr, "<acceptance_criteria>", "XML should contain acceptance_criteria element")
				// For markup content, check individual elements; for plain text check with entity encoding
				if strings.Contains(tt.acceptanceCriteria, "<ol>") {
					assert.Contains(t, xmlStr, "<ol>", "XML should contain ol element")
					assert.Contains(t, xmlStr, "<li>User enters", "XML should contain list items")
					assert.Contains(t, xmlStr, "<strong>JWT token</strong>", "XML should contain strong markup")
				} else if strings.Contains(tt.acceptanceCriteria, "<") {
					// Handle HTML entity encoding
					expectedEncoded := strings.ReplaceAll(tt.acceptanceCriteria, "<", "&lt;")
					assert.Contains(t, xmlStr, expectedEncoded, "XML should contain acceptance criteria content with proper encoding")
				} else {
					assert.Contains(t, xmlStr, tt.acceptanceCriteria, "XML should contain acceptance criteria content")
				}
			} else {
				assert.NotContains(t, xmlStr, "<acceptance_criteria>", "XML should not contain empty acceptance_criteria element")
			}

			// Load the epic back
			loadedEpic, err := fs.LoadEpic(epicFile)
			require.NoError(t, err)
			require.Len(t, loadedEpic.Tasks, 1)

			loadedTask := loadedEpic.Tasks[0]

			// Verify acceptance criteria are preserved (allow for XML formatting differences)
			if strings.Contains(tt.acceptanceCriteria, "<ol>") {
				// For markup content, check that key elements are present
				assert.Contains(t, loadedTask.AcceptanceCriteria, "<ol>", "Acceptance criteria should contain ol element")
				assert.Contains(t, loadedTask.AcceptanceCriteria, "<li>User enters", "Acceptance criteria should contain list items")
				assert.Contains(t, loadedTask.AcceptanceCriteria, "<strong>JWT token</strong>", "Acceptance criteria should contain markup")
			} else {
				assert.Equal(t, tt.acceptanceCriteria, loadedTask.AcceptanceCriteria, "Acceptance criteria should be preserved")
			}
			assert.Equal(t, "task-1", loadedTask.ID)
			assert.Equal(t, "User Authentication", loadedTask.Name)
			assert.Equal(t, "Implement user authentication system", loadedTask.Description)
		})
	}
}

func TestPhaseDeliverablesAndTaskAcceptanceCriteria(t *testing.T) {
	t.Run("epic with both deliverables and acceptance criteria", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := tempDir + "/test-epic.xml"

		// Create epic with both deliverables and acceptance criteria
		testEpic := &epic.Epic{
			ID:        "full-epic",
			Name:      "Full Feature Epic",
			Status:    epic.StatusWIP,
			CreatedAt: time.Now(),
			Phases: []epic.Phase{
				{
					ID:           "phase-1",
					Name:         "Backend Development",
					Description:  "Develop backend services",
					Deliverables: "<ul><li>REST API</li><li>Database schema</li><li>Authentication service</li></ul>",
					Status:       epic.StatusWIP,
				},
				{
					ID:           "phase-2",
					Name:         "Frontend Development",
					Description:  "Develop user interface",
					Deliverables: "React components, CSS styles, Integration tests",
					Status:       epic.StatusPending,
				},
			},
			Tasks: []epic.Task{
				{
					ID:                 "task-1",
					PhaseID:            "phase-1",
					Name:               "API Development",
					Description:        "Create REST API endpoints",
					AcceptanceCriteria: "GIVEN valid request\nWHEN API is called\nTHEN returns correct response\nAND response time < 100ms",
					Status:             epic.StatusWIP,
				},
				{
					ID:                 "task-2",
					PhaseID:            "phase-1",
					Name:               "Database Setup",
					Description:        "Setup database schema",
					AcceptanceCriteria: "<ol><li>All tables created</li><li>Indexes optimized</li><li>Constraints enforced</li></ol>",
					Status:             epic.StatusPending,
				},
			},
		}

		fs := NewFileStorage()

		// Save the epic
		err := fs.SaveEpic(testEpic, epicFile)
		require.NoError(t, err)

		// Read the raw XML to verify both fields are saved
		xmlContent, err := os.ReadFile(epicFile)
		require.NoError(t, err)
		xmlStr := string(xmlContent)

		// Verify deliverables
		assert.Contains(t, xmlStr, "<deliverables>", "XML should contain deliverables elements")
		assert.Contains(t, xmlStr, "<li>REST API</li>", "XML should contain deliverables markup")
		assert.Contains(t, xmlStr, "React components, CSS styles", "XML should contain plain text deliverables")

		// Verify acceptance criteria
		assert.Contains(t, xmlStr, "<acceptance_criteria>", "XML should contain acceptance_criteria elements")
		assert.Contains(t, xmlStr, "GIVEN valid request", "XML should contain BDD-style criteria")
		assert.Contains(t, xmlStr, "<li>All tables created</li>", "XML should contain markup criteria")

		// Load the epic back
		loadedEpic, err := fs.LoadEpic(epicFile)
		require.NoError(t, err)
		require.Len(t, loadedEpic.Phases, 2)
		require.Len(t, loadedEpic.Tasks, 2)

		// Verify phase deliverables
		phase1 := loadedEpic.Phases[0]
		assert.Equal(t, "phase-1", phase1.ID)
		assert.Contains(t, phase1.Deliverables, "<ul>", "Phase 1 deliverables should contain ul element")
		assert.Contains(t, phase1.Deliverables, "<li>REST API</li>", "Phase 1 deliverables should be preserved")

		phase2 := loadedEpic.Phases[1]
		assert.Equal(t, "phase-2", phase2.ID)
		assert.Equal(t, "React components, CSS styles, Integration tests", phase2.Deliverables, "Phase 2 deliverables should be preserved")

		// Verify task acceptance criteria
		task1 := loadedEpic.Tasks[0]
		assert.Equal(t, "task-1", task1.ID)
		assert.Contains(t, task1.AcceptanceCriteria, "GIVEN valid request", "Task 1 acceptance criteria should be preserved")
		assert.Contains(t, task1.AcceptanceCriteria, "response time < 100ms", "Task 1 acceptance criteria should be preserved")

		task2 := loadedEpic.Tasks[1]
		assert.Equal(t, "task-2", task2.ID)
		assert.Contains(t, task2.AcceptanceCriteria, "<ol>", "Task 2 acceptance criteria should contain ol element")
		assert.Contains(t, task2.AcceptanceCriteria, "<li>All tables created</li>", "Task 2 acceptance criteria should be preserved")
		assert.Contains(t, task2.AcceptanceCriteria, "<li>Constraints enforced</li>", "Task 2 acceptance criteria should be preserved")
	})
}
