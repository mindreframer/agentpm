package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/memomoo/agentpm/internal/epic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadataSection(t *testing.T) {
	t.Run("epic with full metadata saves and loads correctly", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create epic with metadata
		testEpic := epic.NewEpic("epic-1", "Test Epic")
		testEpic.Description = "Epic with metadata"
		testEpic.Metadata.Assignee = "john.doe@company.com"
		testEpic.Metadata.EstimatedEffort = "3 days"

		storage := NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Load and verify
		loadedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		assert.Equal(t, testEpic.ID, loadedEpic.ID)
		assert.Equal(t, testEpic.Name, loadedEpic.Name)
		require.NotNil(t, loadedEpic.Metadata)
		assert.Equal(t, "john.doe@company.com", loadedEpic.Metadata.Assignee)
		assert.Equal(t, "3 days", loadedEpic.Metadata.EstimatedEffort)

		// Check that created timestamp is preserved (within reasonable tolerance)
		timeDiff := testEpic.Metadata.Created.Sub(loadedEpic.Metadata.Created)
		assert.True(t, timeDiff < time.Second, "Created timestamp should be preserved accurately")
	})

	t.Run("epic with partial metadata saves and loads correctly", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create epic with partial metadata
		testEpic := epic.NewEpic("epic-2", "Partial Metadata Epic")
		testEpic.Metadata.Assignee = "jane.smith@company.com"
		// Leave EstimatedEffort empty

		storage := NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Load and verify
		loadedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		require.NotNil(t, loadedEpic.Metadata)
		assert.Equal(t, "jane.smith@company.com", loadedEpic.Metadata.Assignee)
		assert.Equal(t, "", loadedEpic.Metadata.EstimatedEffort)
		assert.False(t, loadedEpic.Metadata.Created.IsZero())
	})

	t.Run("epic with empty metadata saves and loads correctly", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create epic with empty metadata (but metadata section present)
		testEpic := epic.NewEpic("epic-3", "Empty Metadata Epic")
		// Metadata exists but all fields are empty/default

		storage := NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Load and verify
		loadedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		require.NotNil(t, loadedEpic.Metadata)
		assert.Equal(t, "", loadedEpic.Metadata.Assignee)
		assert.Equal(t, "", loadedEpic.Metadata.EstimatedEffort)
		assert.False(t, loadedEpic.Metadata.Created.IsZero())
	})

	t.Run("backward compatibility with epics without metadata", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "legacy-epic.xml")

		// Create a legacy epic structure without metadata
		legacyXML := `<?xml version="1.0" encoding="UTF-8"?>
<epic id="legacy-epic" name="Legacy Epic" status="active" created_at="2025-08-16T10:00:00Z">
    <assignee>legacy-assignee</assignee>
    <description>Legacy epic without metadata section</description>
    <phases>
        <phase id="phase-1" name="Phase 1" status="planning"/>
    </phases>
    <tasks>
        <task id="task-1" phase_id="phase-1" name="Task 1" status="planning"/>
    </tasks>
    <events/>
</epic>`

		// Write legacy file
		err := writeFileContent(epicFile, legacyXML)
		require.NoError(t, err)

		// Load legacy epic
		storage := NewFileStorage()
		loadedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		// Verify metadata is nil for backward compatibility
		assert.Nil(t, loadedEpic.Metadata)
		assert.Equal(t, "legacy-epic", loadedEpic.ID)
		assert.Equal(t, "Legacy Epic", loadedEpic.Name)
		assert.Equal(t, "legacy-assignee", loadedEpic.Assignee)

		// Verify we can still save the epic without issues
		require.NoError(t, storage.SaveEpic(loadedEpic, epicFile))
	})

	t.Run("metadata section is preserved during epic operations", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create epic with metadata
		testEpic := epic.NewEpic("epic-4", "Persistence Test Epic")
		testEpic.Metadata.Assignee = "test@company.com"
		testEpic.Metadata.EstimatedEffort = "1 week"

		// Add some phases and tasks
		testEpic.Phases = []epic.Phase{
			{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPlanning},
		}
		testEpic.Tasks = []epic.Task{
			{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPlanning},
		}

		storage := NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Load, modify, and save again
		loadedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		// Modify some non-metadata fields
		loadedEpic.Description = "Updated description"
		loadedEpic.Tasks[0].Status = epic.StatusActive

		require.NoError(t, storage.SaveEpic(loadedEpic, epicFile))

		// Load again and verify metadata is still intact
		finalEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		require.NotNil(t, finalEpic.Metadata)
		assert.Equal(t, "test@company.com", finalEpic.Metadata.Assignee)
		assert.Equal(t, "1 week", finalEpic.Metadata.EstimatedEffort)
		assert.Equal(t, "Updated description", finalEpic.Description)
		assert.Equal(t, epic.StatusActive, finalEpic.Tasks[0].Status)
	})

	t.Run("created timestamp handling", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "timestamp-epic.xml")

		// Create epic and capture the creation time
		beforeCreate := time.Now().Add(-time.Second) // Allow 1 second buffer
		testEpic := epic.NewEpic("timestamp-epic", "Timestamp Test")
		afterCreate := time.Now().Add(time.Second) // Allow 1 second buffer

		storage := NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Load and verify timestamp is within expected range
		loadedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		require.NotNil(t, loadedEpic.Metadata)
		assert.True(t, loadedEpic.Metadata.Created.After(beforeCreate),
			"Created timestamp should be after beforeCreate: %v vs %v",
			loadedEpic.Metadata.Created, beforeCreate)
		assert.True(t, loadedEpic.Metadata.Created.Before(afterCreate),
			"Created timestamp should be before afterCreate: %v vs %v",
			loadedEpic.Metadata.Created, afterCreate)
	})

	t.Run("estimated effort format validation", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "effort-epic.xml")

		testCases := []string{
			"2 hours",
			"3 days",
			"1 week",
			"2 months",
			"0.5 days",
			"", // empty should be valid too
		}

		for _, effort := range testCases {
			t.Run("effort: "+effort, func(t *testing.T) {
				testEpic := epic.NewEpic("effort-test", "Effort Test")
				testEpic.Metadata.EstimatedEffort = effort

				storage := NewFileStorage()
				require.NoError(t, storage.SaveEpic(testEpic, epicFile))

				loadedEpic, err := storage.LoadEpic(epicFile)
				require.NoError(t, err)
				require.NotNil(t, loadedEpic.Metadata)
				assert.Equal(t, effort, loadedEpic.Metadata.EstimatedEffort)
			})
		}
	})

	t.Run("assignee field handling", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "assignee-epic.xml")

		testCases := []string{
			"john.doe@company.com",
			"jane-smith",
			"user123",
			"", // empty should be valid
		}

		for _, assignee := range testCases {
			t.Run("assignee: "+assignee, func(t *testing.T) {
				testEpic := epic.NewEpic("assignee-test", "Assignee Test")
				testEpic.Metadata.Assignee = assignee

				storage := NewFileStorage()
				require.NoError(t, storage.SaveEpic(testEpic, epicFile))

				loadedEpic, err := storage.LoadEpic(epicFile)
				require.NoError(t, err)
				require.NotNil(t, loadedEpic.Metadata)
				assert.Equal(t, assignee, loadedEpic.Metadata.Assignee)
			})
		}
	})
}

// Helper function to write content to file
func writeFileContent(filePath, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}
