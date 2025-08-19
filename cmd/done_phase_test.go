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

func TestDonePhaseCommand(t *testing.T) {
	t.Run("complete phase with all tasks done", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic with active phase and completed tasks
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusCompleted},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := DonePhaseCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"done-phase", "phase-1", "--file", epicFile, "--time", "2025-08-16T16:30:00Z"}
		err := cmd.Run(context.Background(), args)

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Phase phase-1 completed.")

		// Verify phase was updated
		updatedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		var phase1 *epic.Phase
		for i := range updatedEpic.Phases {
			if updatedEpic.Phases[i].ID == "phase-1" {
				phase1 = &updatedEpic.Phases[i]
				break
			}
		}

		require.NotNil(t, phase1)
		assert.Equal(t, epic.StatusCompleted, phase1.Status)
		assert.NotNil(t, phase1.CompletedAt)
	})

	t.Run("complete phase with cancelled tasks", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic with active phase, completed and cancelled tasks
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusCancelled},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := DonePhaseCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"done-phase", "phase-1", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Phase phase-1 completed.")

		// Verify phase was completed
		updatedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		var phase1 *epic.Phase
		for i := range updatedEpic.Phases {
			if updatedEpic.Phases[i].ID == "phase-1" {
				phase1 = &updatedEpic.Phases[i]
				break
			}
		}

		require.NotNil(t, phase1)
		assert.Equal(t, epic.StatusCompleted, phase1.Status)
	})

	t.Run("prevent completing phase with pending tasks", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic with active phase and pending tasks
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusPending},
				{ID: "task-3", PhaseID: "phase-1", Name: "Task 3", Status: epic.StatusActive},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := DonePhaseCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command - should fail
		args := []string{"done-phase", "phase-1", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		output := stderr.String()
		// Epic 13 validation format
		assert.Contains(t, output, "blocking items")
		assert.Contains(t, output, "active")
		assert.Contains(t, output, "phase-1")

		// Verify phase is still active
		updatedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		var phase1 *epic.Phase
		for i := range updatedEpic.Phases {
			if updatedEpic.Phases[i].ID == "phase-1" {
				phase1 = &updatedEpic.Phases[i]
				break
			}
		}

		require.NotNil(t, phase1)
		assert.Equal(t, epic.StatusActive, phase1.Status)
		assert.Nil(t, phase1.CompletedAt)
	})

	t.Run("cannot complete phase that is not active", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic with pending phase
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPending},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := DonePhaseCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command - should fail
		args := []string{"done-phase", "phase-1", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		output := stderr.String()
		assert.Contains(t, output, "<error>")
		assert.Contains(t, output, "<type>invalid_phase_state</type>")
	})

	t.Run("require phase ID argument", func(t *testing.T) {
		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := DonePhaseCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command without phase ID
		args := []string{"done-phase"}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "phase ID is required")
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
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := DonePhaseCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command with invalid timestamp
		args := []string{"done-phase", "phase-1", "--file", epicFile, "--time", "invalid-time"}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid time format")
		assert.Contains(t, err.Error(), "use ISO 8601 format")
	})
}
