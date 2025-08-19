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

func TestStartPhaseCommand(t *testing.T) {
	t.Run("start phase successfully", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic with pending phase
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPending},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPending},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartPhaseCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"start-phase", "phase-1", "--file", epicFile, "--time", "2025-08-16T15:30:00Z"}
		err := cmd.Run(context.Background(), args)

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Phase phase-1 started.")

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
		assert.Equal(t, epic.StatusActive, phase1.Status)
		assert.NotNil(t, phase1.StartedAt)
	})

	t.Run("prevent multiple active phases", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic with one active phase
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPending},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartPhaseCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"start-phase", "phase-2", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		output := stderr.String()
		snapshotTester := apmtesting.NewSnapshotTester()
		snapshotTester.MatchXMLSnapshot(t, output, "start_phase_constraint_violation_error")
	})

	t.Run("cannot start non-existent phase", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic
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
		cmd := StartPhaseCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"start-phase", "non-existent", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "phase non-existent not found")
	})

	t.Run("require phase ID argument", func(t *testing.T) {
		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartPhaseCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command without phase ID
		args := []string{"start-phase"}
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
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusPending},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartPhaseCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command with invalid timestamp
		args := []string{"start-phase", "phase-1", "--file", epicFile, "--time", "invalid-time"}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid time format")
		assert.Contains(t, err.Error(), "use ISO 8601 format")
	})

	t.Run("handle missing epic file", func(t *testing.T) {
		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := StartPhaseCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command with non-existent file
		args := []string{"start-phase", "phase-1", "--file", "/non/existent/file.xml"}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load epic")
	})
}
