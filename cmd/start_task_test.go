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

func TestStartTaskCommand(t *testing.T) {
	t.Run("start task in active phase successfully", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic with active phase and pending task
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPlanning},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusPlanning},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartTaskCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"start-task", "task-1", "--file", epicFile, "--time", "2025-08-16T15:30:00Z"}
		err := cmd.Run(context.Background(), args)

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Task task-1 started.")

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
		assert.Equal(t, epic.StatusActive, task1.Status)
		assert.NotNil(t, task1.StartedAt)
	})

	t.Run("prevent starting task in non-active phase", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic with one active phase and one pending phase
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPlanning},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPlanning},
				{ID: "task-2", PhaseID: "phase-2", Name: "Task 2", Status: epic.StatusPlanning},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartTaskCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command for task in inactive phase - should fail
		args := []string{"start-task", "task-2", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		output := stderr.String()

		// Verify error contains expected information
		assert.Contains(t, output, "Cannot start task task-2: phase phase-2 is not active")
		assert.Contains(t, output, "task_phase_violation")
		assert.Contains(t, output, "Start phase phase-2 first")
	})

	t.Run("prevent multiple active tasks in same phase", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic with active phase and one active task
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusActive},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusPlanning},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartTaskCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command for second task - should fail
		args := []string{"start-task", "task-2", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		output := stderr.String()

		// Verify error contains expected information
		assert.Contains(t, output, "Cannot start task task-2: task task-1 is already active")
		assert.Contains(t, output, "task_constraint_violation")
		assert.Contains(t, output, "Complete task 'task-1'")
	})

	t.Run("cannot start task that is not pending", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic with completed task
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartTaskCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command - should fail
		args := []string{"start-task", "task-1", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		output := stderr.String()

		// Verify error contains expected information
		assert.Contains(t, output, "Cannot start task task-1")
		assert.Contains(t, output, "not in pending state")
		assert.Contains(t, output, "already completed")
	})

	t.Run("cannot start non-existent task", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartTaskCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"start-task", "non-existent", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "task non-existent not found")
	})

	t.Run("require task ID argument", func(t *testing.T) {
		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartTaskCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command without task ID
		args := []string{"start-task"}
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
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPlanning},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartTaskCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command with invalid timestamp
		args := []string{"start-task", "task-1", "--file", epicFile, "--time", "invalid-time"}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid time format")
		assert.Contains(t, err.Error(), "use ISO 8601 format")
	})

	t.Run("handle missing epic file", func(t *testing.T) {
		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartTaskCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command with non-existent file
		args := []string{"start-task", "task-1", "--file", "/non/existent/file.xml"}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load epic")
	})
}
