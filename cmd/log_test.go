package cmd

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogCommand(t *testing.T) {
	t.Run("log event with default type implementation", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Events: []epic.Event{},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := LogCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"log", "Implemented pagination controls", "--file", epicFile, "--time", "2025-08-16T15:30:00Z"}
		err := cmd.Run(context.Background(), args)

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Event logged: Implemented pagination controls")

		// Verify event was added
		updatedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)
		require.Len(t, updatedEpic.Events, 1)

		event := updatedEpic.Events[0]
		assert.Equal(t, "implementation", event.Type)
		assert.Equal(t, "Implemented pagination controls", event.Data)
		assert.Equal(t, time.Date(2025, 8, 16, 15, 30, 0, 0, time.UTC), event.Timestamp)
	})

	t.Run("log event with type blocker", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Events: []epic.Event{},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := LogCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"log", "API rate limit blocking progress", "--type", "blocker", "--file", epicFile, "--time", "2025-08-16T15:30:00Z"}
		err := cmd.Run(context.Background(), args)

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Event logged: API rate limit blocking progress")

		// Verify event was added with correct type
		updatedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)
		require.Len(t, updatedEpic.Events, 1)

		event := updatedEpic.Events[0]
		assert.Equal(t, "blocker", event.Type)
		assert.Equal(t, "API rate limit blocking progress", event.Data)
	})

	t.Run("log event with type issue", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Events: []epic.Event{},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := LogCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"log", "Found bug in payment processing", "--type", "issue", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.NoError(t, err)

		// Verify event was added with correct type
		updatedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)
		require.Len(t, updatedEpic.Events, 1)

		event := updatedEpic.Events[0]
		assert.Equal(t, "issue", event.Type)
		assert.Equal(t, "Found bug in payment processing", event.Data)
	})

	t.Run("log event with single file", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Events: []epic.Event{},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := LogCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"log", "Added pagination component", "--files", "src/Pagination.js:added", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.NoError(t, err)

		// Verify event was added with file information
		updatedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)
		require.Len(t, updatedEpic.Events, 1)

		event := updatedEpic.Events[0]
		assert.Equal(t, "implementation", event.Type)
		assert.Equal(t, "Added pagination component [files: src/Pagination.js:added]", event.Data)
	})

	t.Run("log event with multiple files", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Events: []epic.Event{},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := LogCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command
		args := []string{"log", "Refactored user authentication", "--files", "src/auth.js:modified,src/login.js:modified,tests/auth.test.js:added", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.NoError(t, err)

		// Verify event was added with multiple files
		updatedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)
		require.Len(t, updatedEpic.Events, 1)

		event := updatedEpic.Events[0]
		assert.Equal(t, "implementation", event.Type)
		assert.Contains(t, event.Data, "Refactored user authentication [files: ")
		assert.Contains(t, event.Data, "src/auth.js:modified")
		assert.Contains(t, event.Data, "src/login.js:modified")
		assert.Contains(t, event.Data, "tests/auth.test.js:added")
	})

	t.Run("log event with timestamp from --time flag", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Events: []epic.Event{},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := LogCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command with specific timestamp
		customTime := "2025-08-16T10:15:30Z"
		args := []string{"log", "Custom timestamp event", "--time", customTime, "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.NoError(t, err)

		// Verify event has correct timestamp
		updatedEpic, err := storage.LoadEpic(epicFile)
		require.NoError(t, err)
		require.Len(t, updatedEpic.Events, 1)

		event := updatedEpic.Events[0]
		expectedTime, _ := time.Parse(time.RFC3339, customTime)
		assert.Equal(t, expectedTime, event.Timestamp)
	})

	t.Run("error handling for invalid event type", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Events: []epic.Event{},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := LogCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command with invalid type
		args := []string{"log", "Some message", "--type", "invalid-type", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid event type: invalid-type")
		assert.Contains(t, err.Error(), "valid types: implementation, blocker, issue, milestone, decision, note")
	})

	t.Run("error handling for missing message", func(t *testing.T) {
		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := LogCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command without message
		args := []string{"log"}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "event message is required")
	})

	t.Run("error handling for invalid time format", func(t *testing.T) {
		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "test-epic.xml")

		// Create test epic
		testEpic := &epic.Epic{
			ID:     "epic-1",
			Name:   "Test Epic",
			Status: epic.StatusActive,
			Events: []epic.Event{},
		}

		storage := storage.NewFileStorage()
		require.NoError(t, storage.SaveEpic(testEpic, epicFile))

		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := LogCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command with invalid timestamp
		args := []string{"log", "Some message", "--time", "invalid-time", "--file", epicFile}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid time format")
		assert.Contains(t, err.Error(), "use ISO 8601 format")
	})

	t.Run("error handling for missing epic file", func(t *testing.T) {
		// Create CLI command
		var stdout, stderr bytes.Buffer
		cmd := LogCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		// Execute command with non-existent file
		args := []string{"log", "Some message", "--file", "/non/existent/file.xml"}
		err := cmd.Run(context.Background(), args)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to log event")
		assert.Contains(t, err.Error(), "failed to load epic")
	})
}

func TestParseFilesFlag(t *testing.T) {
	t.Run("parse single file action", func(t *testing.T) {
		files, err := parseFilesFlag("src/component.js:added")
		require.NoError(t, err)
		require.Len(t, files, 1)
		assert.Equal(t, "src/component.js", files[0].Path)
		assert.Equal(t, "added", files[0].Action)
	})

	t.Run("parse multiple file actions", func(t *testing.T) {
		files, err := parseFilesFlag("src/auth.js:modified,src/login.js:modified,tests/auth.test.js:added")
		require.NoError(t, err)
		require.Len(t, files, 3)

		assert.Equal(t, "src/auth.js", files[0].Path)
		assert.Equal(t, "modified", files[0].Action)

		assert.Equal(t, "src/login.js", files[1].Path)
		assert.Equal(t, "modified", files[1].Action)

		assert.Equal(t, "tests/auth.test.js", files[2].Path)
		assert.Equal(t, "added", files[2].Action)
	})

	t.Run("handle empty files flag", func(t *testing.T) {
		files, err := parseFilesFlag("")
		require.NoError(t, err)
		assert.Nil(t, files)
	})

	t.Run("handle file path with colons", func(t *testing.T) {
		files, err := parseFilesFlag("C:\\Program Files\\app\\file.js:added")
		require.NoError(t, err)
		require.Len(t, files, 1)
		assert.Equal(t, "C:\\Program Files\\app\\file.js", files[0].Path)
		assert.Equal(t, "added", files[0].Action)
	})

	t.Run("error on missing colon", func(t *testing.T) {
		_, err := parseFilesFlag("src/component.js")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid file format")
		assert.Contains(t, err.Error(), "expected 'path:action'")
	})

	t.Run("error on empty path", func(t *testing.T) {
		_, err := parseFilesFlag(":added")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "path and action cannot be empty")
	})

	t.Run("error on empty action", func(t *testing.T) {
		_, err := parseFilesFlag("src/file.js:")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "path and action cannot be empty")
	})

	t.Run("error on invalid action", func(t *testing.T) {
		_, err := parseFilesFlag("src/file.js:invalid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid file action 'invalid'")
		assert.Contains(t, err.Error(), "valid actions are added, modified, deleted, renamed")
	})
}

func TestIsValidEventType(t *testing.T) {
	validTypes := []string{"implementation", "blocker", "issue", "milestone", "decision", "note"}
	for _, validType := range validTypes {
		t.Run("valid type: "+validType, func(t *testing.T) {
			assert.True(t, isValidEventType(validType))
		})
	}

	invalidTypes := []string{"invalid", "test", "debug", ""}
	for _, invalidType := range invalidTypes {
		t.Run("invalid type: "+invalidType, func(t *testing.T) {
			assert.False(t, isValidEventType(invalidType))
		})
	}
}

func TestIsValidFileAction(t *testing.T) {
	validActions := []string{"added", "modified", "deleted", "renamed"}
	for _, validAction := range validActions {
		t.Run("valid action: "+validAction, func(t *testing.T) {
			assert.True(t, isValidFileAction(validAction))
		})
	}

	invalidActions := []string{"invalid", "created", "updated", ""}
	for _, invalidAction := range invalidActions {
		t.Run("invalid action: "+invalidAction, func(t *testing.T) {
			assert.False(t, isValidFileAction(invalidAction))
		})
	}
}
