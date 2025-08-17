package cmd

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/storage"
	apmtesting "github.com/mindreframer/agentpm/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartNextCommand(t *testing.T) {
	t.Run("start next task in current phase", func(t *testing.T) {
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
				{ID: "task-2", PhaseID: "phase-1", Name: "Task 2", Status: epic.StatusPlanning},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartNextCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"start-next", "--file", epicFile, "--time", "2025-08-16T15:30:00Z"}
		err := cmd.Run(context.Background(), args)

		require.NoError(t, err)

		// Check XML output using snapshots
		output := stdout.String()
		snapshotTester := apmtesting.NewSnapshotTester()
		snapshotTester.MatchXMLSnapshot(t, output, "start_next_task_in_current_phase")

		// Verify task was started
		updatedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		var task2 *epic.Task
		for i := range updatedEpic.Tasks {
			if updatedEpic.Tasks[i].ID == "task-2" {
				task2 = &updatedEpic.Tasks[i]
				break
			}
		}

		require.NotNil(t, task2)
		assert.Equal(t, epic.StatusActive, task2.Status)
	})

	t.Run("start next phase and task", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic with completed phase and pending phase
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPlanning},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
				{ID: "task-2", PhaseID: "phase-2", Name: "Task 2", Status: epic.StatusPlanning},
				{ID: "task-3", PhaseID: "phase-2", Name: "Task 3", Status: epic.StatusPlanning},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartNextCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"start-next", "--file", epicFile, "--time", "2025-08-16T15:30:00Z"}
		err := cmd.Run(context.Background(), args)

		require.NoError(t, err)

		// Check XML output using snapshots
		output := stdout.String()
		snapshotTester := apmtesting.NewSnapshotTester()
		snapshotTester.MatchXMLSnapshot(t, output, "start_next_phase_and_task")

		// Verify phase and task were started
		updatedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		var phase1, phase2 *epic.Phase
		var task2 *epic.Task

		for i := range updatedEpic.Phases {
			if updatedEpic.Phases[i].ID == "phase-1" {
				phase1 = &updatedEpic.Phases[i]
			} else if updatedEpic.Phases[i].ID == "phase-2" {
				phase2 = &updatedEpic.Phases[i]
			}
		}

		for i := range updatedEpic.Tasks {
			if updatedEpic.Tasks[i].ID == "task-2" {
				task2 = &updatedEpic.Tasks[i]
			}
		}

		require.NotNil(t, phase1)
		require.NotNil(t, phase2)
		require.NotNil(t, task2)

		assert.Equal(t, epic.StatusCompleted, phase1.Status)
		assert.Equal(t, epic.StatusActive, phase2.Status)
		assert.Equal(t, epic.StatusActive, task2.Status)
	})

	t.Run("all work completed", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic with all work completed
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusCompleted},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusCompleted},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusCompleted},
				{ID: "task-2", PhaseID: "phase-2", Name: "Task 2", Status: epic.StatusCompleted},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartNextCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"start-next", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.NoError(t, err)

		// Verify XML output for completion
		output := stdout.String()
		assert.Contains(t, output, "<all_complete")
		assert.Contains(t, output, `<message>All phases and tasks completed. Epic ready for completion.</message>`)
		assert.Contains(t, output, `<suggestion>Use 'agentpm done-epic' to complete the epic</suggestion>`)
	})

	t.Run("no work needed when task already active", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic with already active task
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
		cmd := StartNextCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"start-next", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.NoError(t, err)

		// Verify simple text output
		output := stdout.String()
		assert.Contains(t, output, "Task task-1 is already active in phase phase-1")
	})

	t.Run("start first phase when no active phase", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic with no active phase
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPlanning},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPlanning},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartNextCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"start-next", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.NoError(t, err)

		// Verify XML output for phase started
		output := stdout.String()
		assert.Contains(t, output, "<phase_started")
		assert.Contains(t, output, `phase="phase-1"`)
		assert.Contains(t, output, `<started_task>task-1</started_task>`)

		// Verify phase and task were started
		updatedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)

		var phase1 *epic.Phase
		var task1 *epic.Task

		for i := range updatedEpic.Phases {
			if updatedEpic.Phases[i].ID == "phase-1" {
				phase1 = &updatedEpic.Phases[i]
			}
		}

		for i := range updatedEpic.Tasks {
			if updatedEpic.Tasks[i].ID == "task-1" {
				task1 = &updatedEpic.Tasks[i]
			}
		}

		require.NotNil(t, phase1)
		require.NotNil(t, task1)

		assert.Equal(t, epic.StatusActive, phase1.Status)
		assert.Equal(t, epic.StatusActive, task1.Status)
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
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPlanning},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusPlanning},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartNextCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command with invalid timestamp
		args := []string{"start-next", "--file", epicFile, "--time", "invalid-time"}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid time format")
		assert.Contains(t, err.Error(), "use ISO 8601 format")
	})

	t.Run("handle missing epic file", func(t *testing.T) {
		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartNextCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command with non-existent file
		args := []string{"start-next", "--file", "/non/existent/file.xml"}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load epic")
	})

	t.Run("handle empty epic", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create empty epic
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Empty Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{},
			Tasks:  []epic.Task{},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartNextCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"start-next", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.NoError(t, err)

		// Should indicate all work is complete
		output := stdout.String()
		assert.Contains(t, output, "<all_complete")
		assert.Contains(t, output, "All phases and tasks completed")
	})
}
