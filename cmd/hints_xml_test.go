package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/storage"
	apmtesting "github.com/mindreframer/agentpm/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestXMLOutputWithHints(t *testing.T) {
	t.Run("outputXMLErrorWithHint includes hint element", func(t *testing.T) {
		var stderr bytes.Buffer
		cmd := &cli.Command{
			ErrWriter: &stderr,
		}

		err := outputXMLErrorWithHint(cmd, "test_error", "Test error message",
			map[string]interface{}{
				"test_key": "test_value",
			}, "This is a test hint")

		require.Error(t, err)
		output := stderr.String()

		// Check XML output using snapshots
		snapshotTester := apmtesting.NewSnapshotTester()
		snapshotTester.MatchXMLSnapshot(t, output, "outputXMLErrorWithHint_includes_hint_element")
	})

	t.Run("outputXMLErrorWithHint without hint omits hint element", func(t *testing.T) {
		var stderr bytes.Buffer
		cmd := &cli.Command{
			ErrWriter: &stderr,
		}

		err := outputXMLErrorWithHint(cmd, "test_error", "Test error message",
			map[string]interface{}{
				"test_key": "test_value",
			}, "")

		require.Error(t, err)
		output := stderr.String()

		// Check XML output using snapshots
		snapshotTester := apmtesting.NewSnapshotTester()
		snapshotTester.MatchXMLSnapshot(t, output, "outputXMLErrorWithHint_without_hint_omits_hint_element")
	})

	t.Run("outputXMLErrorWithHint without details", func(t *testing.T) {
		var stderr bytes.Buffer
		cmd := &cli.Command{
			ErrWriter: &stderr,
		}

		err := outputXMLErrorWithHint(cmd, "test_error", "Test error message",
			nil, "This is a test hint")

		require.Error(t, err)
		output := stderr.String()

		// Check XML output using snapshots
		snapshotTester := apmtesting.NewSnapshotTester()
		snapshotTester.MatchXMLSnapshot(t, output, "outputXMLErrorWithHint_without_details")
	})

	t.Run("outputXMLErrorWithHint handles special characters in hint", func(t *testing.T) {
		var stderr bytes.Buffer
		cmd := &cli.Command{
			ErrWriter: &stderr,
		}

		hintWithSpecialChars := "Use 'agentpm current' & check status"
		err := outputXMLErrorWithHint(cmd, "test_error", "Test error message",
			nil, hintWithSpecialChars)

		require.Error(t, err)
		output := stderr.String()

		// Check XML output using snapshots
		snapshotTester := apmtesting.NewSnapshotTester()
		snapshotTester.MatchXMLSnapshot(t, output, "outputXMLErrorWithHint_handles_special_characters_in_hint")
	})

	t.Run("outputXMLError compatibility maintained", func(t *testing.T) {
		var stderr bytes.Buffer
		cmd := &cli.Command{
			ErrWriter: &stderr,
		}

		err := outputXMLError(cmd, "test_error", "Test error message",
			map[string]interface{}{
				"test_key": "test_value",
			})

		require.Error(t, err)
		output := stderr.String()

		// Check XML output using snapshots
		snapshotTester := apmtesting.NewSnapshotTester()
		snapshotTester.MatchXMLSnapshot(t, output, "outputXMLError_compatibility_maintained")
	})
}

func TestHintIntegrationInCommands(t *testing.T) {
	t.Run("start_phase command includes hints in XML error output", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := tempDir + "/epic.xml"

		// Create test epic using the storage system
		testEpic := &epic.Epic{
			ID:     "test-epic",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{
					ID:     "phase-1",
					Name:   "First Phase",
					Status: epic.StatusActive,
				},
				{
					ID:     "phase-2",
					Name:   "Second Phase",
					Status: epic.StatusPending,
				},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Try to start second phase while first is active
		var stderr bytes.Buffer
		cmd := StartPhaseCommand()
		cmd.Root().ErrWriter = &stderr

		args := []string{"start-phase", "phase-2", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		// Should return error with hint
		require.Error(t, err)
		output := stderr.String()

		// Verify XML error output includes hint
		assert.Contains(t, output, "<error>")
		assert.Contains(t, output, "<type>phase_constraint_violation</type>")
		assert.Contains(t, output, "<message>Cannot start phase phase-2: phase phase-1 is still active</message>")
		assert.Contains(t, output, "<hint>Complete phase 'phase-1' before starting 'phase-2'</hint>")
		assert.Contains(t, output, "</error>")
	})

	t.Run("start_task command includes hints in XML error output", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := tempDir + "/epic.xml"

		// Create test epic with active phase and task
		testEpic := &epic.Epic{
			ID:     "test-epic",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{
					ID:     "phase-1",
					Name:   "First Phase",
					Status: epic.StatusActive,
				},
			},
			Tasks: []epic.Task{
				{
					ID:      "task-1",
					PhaseID: "phase-1",
					Name:    "First Task",
					Status:  epic.StatusActive,
				},
				{
					ID:      "task-2",
					PhaseID: "phase-1",
					Name:    "Second Task",
					Status:  epic.StatusPending,
				},
			},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Try to start second task while first is active
		var stderr bytes.Buffer
		cmd := StartTaskCommand()
		cmd.Root().ErrWriter = &stderr

		args := []string{"start-task", "task-2", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		// Should return error with hint
		require.Error(t, err)
		output := stderr.String()

		// Verify XML error output includes hint
		assert.Contains(t, output, "<error>")
		assert.Contains(t, output, "<type>task_constraint_violation</type>")
		assert.Contains(t, output, "<message>Cannot start task task-2: task task-1 is already active in phase phase-1</message>")
		assert.Contains(t, output, "<hint>Complete task 'task-1' in phase 'phase-1' before starting 'task-2'</hint>")
		assert.Contains(t, output, "</error>")
	})
}

func TestHintXMLFormatting(t *testing.T) {
	t.Run("hint content is included in XML output", func(t *testing.T) {
		var stderr bytes.Buffer
		cmd := &cli.Command{
			ErrWriter: &stderr,
		}

		// Test basic hint content inclusion
		testCases := []struct {
			name string
			hint string
		}{
			{
				name: "simple hint",
				hint: "Use agentpm current to see active work",
			},
			{
				name: "hint with command",
				hint: "Run 'agentpm status' for overview",
			},
			{
				name: "hint with path",
				hint: "Check the current directory",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				stderr.Reset()

				err := outputXMLErrorWithHint(cmd, "test_error", "Test message", nil, tc.hint)
				require.Error(t, err)

				output := stderr.String()
				assert.Contains(t, output, "<hint>"+tc.hint+"</hint>")
			})
		}
	})

	t.Run("XML structure is well-formed with hints", func(t *testing.T) {
		var stderr bytes.Buffer
		cmd := &cli.Command{
			ErrWriter: &stderr,
		}

		err := outputXMLErrorWithHint(cmd, "complex_error", "Complex error message",
			map[string]interface{}{
				"entity_id":  "test-entity",
				"status":     "active",
				"suggestion": "Try running 'agentpm status'",
			}, "Complete the current task before proceeding")

		require.Error(t, err)
		output := stderr.String()

		// Verify XML structure is well-formed
		expectedElements := []string{
			"<error>",
			"<type>complex_error</type>",
			"<message>Complex error message</message>",
			"<hint>Complete the current task before proceeding</hint>",
			"<details>",
			"<entity_id>test-entity</entity_id>",
			"<status>active</status>",
			"<suggestion>Try running 'agentpm status'</suggestion>",
			"</details>",
			"</error>",
		}

		for _, element := range expectedElements {
			assert.Contains(t, output, element, "Missing XML element: %s", element)
		}

		// Verify order of elements is correct
		lines := strings.Split(strings.TrimSpace(output), "\n")
		assert.Equal(t, "<error>", strings.TrimSpace(lines[0]))
		assert.Contains(t, lines[1], "<type>complex_error</type>")
		assert.Contains(t, lines[2], "<message>Complex error message</message>")
		assert.Contains(t, lines[3], "<hint>Complete the current task before proceeding</hint>")
		assert.Contains(t, lines[4], "<details>")
		assert.Equal(t, "</error>", strings.TrimSpace(lines[len(lines)-1]))
	})
}
