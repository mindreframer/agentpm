package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrentStateSection(t *testing.T) {
	t.Run("epic with full current_state saves and loads correctly", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create epic with current_state
		testEpic := epic.NewEpic("epic-1", "Test Epic")
		testEpic.Description = "Epic with current_state"
		testEpic.CurrentState.ActivePhase = "phase-1"
		testEpic.CurrentState.ActiveTask = "task-1"
		testEpic.CurrentState.NextAction = "Continue work on feature A"

		storage := NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Load and verify
		loadedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		assert.Equal(t, testEpic.ID, loadedEpic.ID)
		assert.Equal(t, testEpic.Name, loadedEpic.Name)
		require.NotNil(t, loadedEpic.CurrentState)
		assert.Equal(t, "phase-1", loadedEpic.CurrentState.ActivePhase)
		assert.Equal(t, "task-1", loadedEpic.CurrentState.ActiveTask)
		assert.Equal(t, "Continue work on feature A", loadedEpic.CurrentState.NextAction)
	})

	t.Run("epic with partial current_state saves and loads correctly", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create epic with partial current_state
		testEpic := epic.NewEpic("epic-2", "Partial CurrentState Epic")
		testEpic.CurrentState.ActivePhase = "phase-2"
		// Leave ActiveTask empty
		testEpic.CurrentState.NextAction = "Start next task"

		storage := NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Load and verify
		loadedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		require.NotNil(t, loadedEpic.CurrentState)
		assert.Equal(t, "phase-2", loadedEpic.CurrentState.ActivePhase)
		assert.Equal(t, "", loadedEpic.CurrentState.ActiveTask)
		assert.Equal(t, "Start next task", loadedEpic.CurrentState.NextAction)
	})

	t.Run("epic with empty current_state saves and loads correctly", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create epic with empty current_state (but section present)
		testEpic := epic.NewEpic("epic-3", "Empty CurrentState Epic")
		// CurrentState exists but all fields are empty

		storage := NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Load and verify
		loadedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		require.NotNil(t, loadedEpic.CurrentState)
		assert.Equal(t, "", loadedEpic.CurrentState.ActivePhase)
		assert.Equal(t, "", loadedEpic.CurrentState.ActiveTask)
		assert.Equal(t, "Start next phase", loadedEpic.CurrentState.NextAction)
	})

	t.Run("backward compatibility with epics without current_state", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "legacy-epic.xml")

		// Create a legacy epic structure without current_state
		legacyXML := `<?xml version="1.0" encoding="UTF-8"?>
<epic id="legacy-epic" name="Legacy Epic" status="wip" created_at="2025-08-16T10:00:00Z">
    <assignee>legacy-assignee</assignee>
    <description>Legacy epic without current_state section</description>
    <phases>
        <phase id="phase-1" name="Phase 1" status="planning"/>
    </phases>
    <tasks>
        <task id="task-1" phase_id="phase-1" name="Task 1" status="planning"/>
    </tasks>
    <events/>
</epic>`

		// Write legacy file
		err := os.WriteFile(epicFile, []byte(legacyXML), 0644)
		require.NoError(t, err)

		// Load legacy epic
		storage := NewFileStorage()
		loadedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		// Verify current_state is nil for backward compatibility
		assert.Nil(t, loadedEpic.CurrentState)
		assert.Equal(t, "legacy-epic", loadedEpic.ID)
		assert.Equal(t, "Legacy Epic", loadedEpic.Name)
		assert.Equal(t, "legacy-assignee", loadedEpic.Assignee)

		// Verify we can still save the epic without issues
		require.NoError(t, storage.SaveEpic(loadedEpic, epicFile))
	})

	t.Run("current_state section is preserved during epic operations", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create epic with current_state
		testEpic := epic.NewEpic("epic-4", "Persistence Test Epic")
		testEpic.CurrentState.ActivePhase = "phase-1"
		testEpic.CurrentState.ActiveTask = "task-1"
		testEpic.CurrentState.NextAction = "Continue current work"

		// Add some phases and tasks
		testEpic.Phases = []epic.Phase{
			{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
		}
		testEpic.Tasks = []epic.Task{
			{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusWIP},
		}

		storage := NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Load, modify, and save again
		loadedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		// Modify some non-current_state fields
		loadedEpic.Description = "Updated description"
		loadedEpic.Tasks[0].Description = "Updated task description"

		require.NoError(t, storage.SaveEpic(loadedEpic, epicFile))

		// Load again and verify current_state is still intact
		finalEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		require.NotNil(t, finalEpic.CurrentState)
		assert.Equal(t, "phase-1", finalEpic.CurrentState.ActivePhase)
		assert.Equal(t, "task-1", finalEpic.CurrentState.ActiveTask)
		assert.Equal(t, "Continue current work", finalEpic.CurrentState.NextAction)
		assert.Equal(t, "Updated description", finalEpic.Description)
		assert.Equal(t, "Updated task description", finalEpic.Tasks[0].Description)
	})

	t.Run("active_phase field validation", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "phase-epic.xml")

		testCases := []string{
			"phase-1",
			"phase_2",
			"setup-phase",
			"", // empty should be valid
		}

		for _, phaseID := range testCases {
			t.Run("phase: "+phaseID, func(t *testing.T) {
				testEpic := epic.NewEpic("phase-test", "Phase Test")
				testEpic.CurrentState.ActivePhase = phaseID

				storage := NewFileStorage()
				require.NoError(t, storage.SaveEpic(testEpic, epicFile))

				loadedEpic, err := storage.LoadEpic(epicFile)
				require.NoError(t, err)
				require.NotNil(t, loadedEpic.CurrentState)
				assert.Equal(t, phaseID, loadedEpic.CurrentState.ActivePhase)
			})
		}
	})

	t.Run("active_task field validation", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "task-epic.xml")

		testCases := []string{
			"task-1",
			"task_2",
			"setup-task",
			"", // empty should be valid
		}

		for _, taskID := range testCases {
			t.Run("task: "+taskID, func(t *testing.T) {
				testEpic := epic.NewEpic("task-test", "Task Test")
				testEpic.CurrentState.ActiveTask = taskID

				storage := NewFileStorage()
				require.NoError(t, storage.SaveEpic(testEpic, epicFile))

				loadedEpic, err := storage.LoadEpic(epicFile)
				require.NoError(t, err)
				require.NotNil(t, loadedEpic.CurrentState)
				assert.Equal(t, taskID, loadedEpic.CurrentState.ActiveTask)
			})
		}
	})

	t.Run("next_action field validation", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "action-epic.xml")

		testCases := []string{
			"Start next task",
			"Continue work on feature A",
			"Complete current phase",
			"Review and deploy",
			"", // empty should be valid
		}

		for _, action := range testCases {
			t.Run("action: "+action, func(t *testing.T) {
				testEpic := epic.NewEpic("action-test", "Action Test")
				testEpic.CurrentState.NextAction = action

				storage := NewFileStorage()
				require.NoError(t, storage.SaveEpic(testEpic, epicFile))

				loadedEpic, err := storage.LoadEpic(epicFile)
				require.NoError(t, err)
				require.NotNil(t, loadedEpic.CurrentState)
				assert.Equal(t, action, loadedEpic.CurrentState.NextAction)
			})
		}
	})

	t.Run("current_state with special characters", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "special-epic.xml")

		// Test with special characters that need XML escaping
		testEpic := epic.NewEpic("special-test", "Special Characters Test")
		testEpic.CurrentState.ActivePhase = "phase-1"
		testEpic.CurrentState.ActiveTask = "task-1"
		testEpic.CurrentState.NextAction = "Review & test features with <component> integration"

		storage := NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		loadedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		require.NotNil(t, loadedEpic.CurrentState)
		assert.Equal(t, "phase-1", loadedEpic.CurrentState.ActivePhase)
		assert.Equal(t, "task-1", loadedEpic.CurrentState.ActiveTask)
		assert.Equal(t, "Review & test features with <component> integration", loadedEpic.CurrentState.NextAction)
	})

	t.Run("current_state initialization in NewEpic", func(t *testing.T) {
		// Test that NewEpic creates a proper current_state
		testEpic := epic.NewEpic("new-epic", "New Epic Test")

		require.NotNil(t, testEpic.CurrentState)
		assert.Equal(t, "", testEpic.CurrentState.ActivePhase)
		assert.Equal(t, "", testEpic.CurrentState.ActiveTask)
		assert.Equal(t, "Start next phase", testEpic.CurrentState.NextAction)
	})
}
