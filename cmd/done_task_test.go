package cmd

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoneTaskCommand(t *testing.T) {
	t.Run("complete active task", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic with active task
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusWIP},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := DoneTaskCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"done-task", "task-1", "--file", epicFile, "--time", "2025-08-16T16:30:00Z"}
		err := cmd.Run(context.Background(), args)

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Task task-1 completed.")

		// Verify task was updated
		updatedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		var task1 *epic.Task
		for i := range updatedEpic.Tasks {
			if updatedEpic.Tasks[i].ID == "task-1" {
				task1 = &updatedEpic.Tasks[i]
				break
			}
		}

		require.NotNil(t, task1)
		assert.Equal(t, epic.StatusCompleted, task1.Status)
		assert.NotNil(t, task1.CompletedAt)
	})

	t.Run("cannot complete task that is not active", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic with pending task
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPending},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := DoneTaskCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command - should fail
		args := []string{"done-task", "task-1", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		output := stderr.String()
		assert.Contains(t, output, "<error>")
		assert.Contains(t, output, "<type>invalid_task_state</type>")
	})

	t.Run("require task ID argument", func(t *testing.T) {
		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := DoneTaskCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command without task ID
		args := []string{"done-task"}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "task ID is required")
	})

	t.Run("handle invalid timestamp format", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusWIP},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := DoneTaskCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command with invalid timestamp
		args := []string{"done-task", "task-1", "--file", epicFile, "--time", "invalid-time"}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid time format")
		assert.Contains(t, err.Error(), "use ISO 8601 format")
	})
}

func TestCancelTaskCommand(t *testing.T) {
	t.Run("cancel active task", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic with active task
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusWIP},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := CancelTaskCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"cancel-task", "task-1", "--file", epicFile, "--time", "2025-08-16T16:30:00Z"}
		err := cmd.Run(context.Background(), args)

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Task task-1 cancelled.")

		// Verify task was updated
		updatedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		var task1 *epic.Task
		for i := range updatedEpic.Tasks {
			if updatedEpic.Tasks[i].ID == "task-1" {
				task1 = &updatedEpic.Tasks[i]
				break
			}
		}

		require.NotNil(t, task1)
		assert.Equal(t, epic.StatusCancelled, task1.Status)
		assert.NotNil(t, task1.CancelledAt)
	})

	t.Run("cannot cancel task that is not active", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic with pending task
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusWIP,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPending},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := CancelTaskCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command - should fail
		args := []string{"cancel-task", "task-1", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		output := stderr.String()
		assert.Contains(t, output, "<error>")
		assert.Contains(t, output, "<type>invalid_task_state</type>")
	})

	t.Run("require task ID argument", func(t *testing.T) {
		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := CancelTaskCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command without task ID
		args := []string{"cancel-task"}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "task ID is required")
	})
}
