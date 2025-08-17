package messages

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputFormat(t *testing.T) {
	t.Run("String representation", func(t *testing.T) {
		assert.Equal(t, "text", FormatText.String())
		assert.Equal(t, "json", FormatJSON.String())
		assert.Equal(t, "xml", FormatXML.String())
	})

	t.Run("Validation", func(t *testing.T) {
		assert.True(t, FormatText.IsValid())
		assert.True(t, FormatJSON.IsValid())
		assert.True(t, FormatXML.IsValid())
		assert.False(t, OutputFormat("invalid").IsValid())
	})
}

func TestMessageOutput(t *testing.T) {
	t.Run("NewMessageOutput", func(t *testing.T) {
		output := NewMessageOutput(FormatJSON)
		assert.NotNil(t, output)
		assert.Equal(t, FormatJSON, output.format)
		assert.NotNil(t, output.formatter)
	})

	t.Run("NewMessageOutputWithFormatter", func(t *testing.T) {
		customFormatter := NewMessageFormatter()
		output := NewMessageOutputWithFormatter(FormatXML, customFormatter)

		assert.NotNil(t, output)
		assert.Equal(t, FormatXML, output.format)
		assert.Same(t, customFormatter, output.formatter)
	})

	t.Run("Format text", func(t *testing.T) {
		output := NewMessageOutput(FormatText)
		msg := InfoMessage("Test message")

		result, err := output.Format(msg)
		require.NoError(t, err)
		assert.Equal(t, "[info] Test message", result)
	})

	t.Run("Format JSON", func(t *testing.T) {
		output := NewMessageOutput(FormatJSON)
		msg := SuccessMessage("Operation completed")

		result, err := output.Format(msg)
		require.NoError(t, err)

		// Should be valid JSON
		assert.Contains(t, result, `"type": "success"`)
		assert.Contains(t, result, `"content": "Operation completed"`)
	})

	t.Run("Format XML", func(t *testing.T) {
		output := NewMessageOutput(FormatXML)
		msg := ErrorMessage("Something failed")

		result, err := output.Format(msg)
		require.NoError(t, err)

		// Should be valid XML
		assert.Contains(t, result, `type="error"`)
		assert.Contains(t, result, "Something failed")
	})

	t.Run("Format unsupported format", func(t *testing.T) {
		output := &MessageOutput{
			formatter: NewMessageFormatter(),
			format:    OutputFormat("unsupported"),
		}
		msg := InfoMessage("Test")

		result, err := output.Format(msg)
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "unsupported output format")
	})

	t.Run("FormatMultiple - text format", func(t *testing.T) {
		output := NewMessageOutput(FormatText)
		msgs := []*Message{
			InfoMessage("First message"),
			ErrorMessage("Second message"),
		}

		result, err := output.FormatMultiple(msgs)
		require.NoError(t, err)

		expected := "[info] First message\n[error] Second message"
		assert.Equal(t, expected, result)
	})

	t.Run("FormatMultiple - JSON format", func(t *testing.T) {
		output := NewMessageOutput(FormatJSON)
		msgs := []*Message{
			InfoMessage("First"),
			WarningMessage("Second"),
		}

		result, err := output.FormatMultiple(msgs)
		require.NoError(t, err)

		assert.Contains(t, result, "[")
		assert.Contains(t, result, "]")
		assert.Contains(t, result, `"type": "info"`)
		assert.Contains(t, result, `"type": "warning"`)
	})

	t.Run("FormatMultiple - XML format", func(t *testing.T) {
		output := NewMessageOutput(FormatXML)
		msgs := []*Message{
			SuccessMessage("Done"),
			InfoMessage("Note"),
		}

		result, err := output.FormatMultiple(msgs)
		require.NoError(t, err)

		assert.Contains(t, result, "<messages>")
		assert.Contains(t, result, "</messages>")
		assert.Contains(t, result, `type="success"`)
		assert.Contains(t, result, `type="info"`)
	})

	t.Run("FormatMultiple - empty list", func(t *testing.T) {
		output := NewMessageOutput(FormatText)

		result, err := output.FormatMultiple([]*Message{})
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("SetFormat and GetFormat", func(t *testing.T) {
		output := NewMessageOutput(FormatText)
		assert.Equal(t, FormatText, output.GetFormat())

		output.SetFormat(FormatJSON)
		assert.Equal(t, FormatJSON, output.GetFormat())
	})
}

func TestFriendlyResponseTemplates(t *testing.T) {
	templates := NewFriendlyResponseTemplates()

	t.Run("AlreadyStarted", func(t *testing.T) {
		msg := templates.AlreadyStarted("phase", "implementation")

		assert.Equal(t, MessageInfo, msg.Type)
		assert.Contains(t, msg.Content, "Phase 'implementation' is already started")
		assert.Contains(t, msg.Hint, "agentpm status")
	})

	t.Run("AlreadyCompleted", func(t *testing.T) {
		msg := templates.AlreadyCompleted("task", "setup-db")

		assert.Equal(t, MessageInfo, msg.Type)
		assert.Contains(t, msg.Content, "Task 'setup-db' is already completed")
		assert.Contains(t, msg.Hint, "agentpm status")
	})

	t.Run("AlreadyInState", func(t *testing.T) {
		msg := templates.AlreadyInState("epic", "feature-x", "active")

		assert.Equal(t, MessageInfo, msg.Type)
		assert.Contains(t, msg.Content, "Epic 'feature-x' is already active")
	})

	t.Run("SuccessfulTransition", func(t *testing.T) {
		msg := templates.SuccessfulTransition("phase", "testing", "started")

		assert.Equal(t, MessageSuccess, msg.Type)
		assert.Contains(t, msg.Content, "Phase 'testing' started successfully")
	})

	t.Run("OperationComplete with details", func(t *testing.T) {
		msg := templates.OperationComplete("Migration", "Moved 5 files")

		assert.Equal(t, MessageSuccess, msg.Type)
		assert.Contains(t, msg.Content, "Migration completed: Moved 5 files")
	})

	t.Run("OperationComplete without details", func(t *testing.T) {
		msg := templates.OperationComplete("Backup", "")

		assert.Equal(t, MessageSuccess, msg.Type)
		assert.Contains(t, msg.Content, "Backup completed successfully")
	})

	t.Run("ValidationError", func(t *testing.T) {
		msg := templates.ValidationError("Invalid input format", "Use YYYY-MM-DD format")

		assert.Equal(t, MessageError, msg.Type)
		assert.Equal(t, "Invalid input format", msg.Content)
		assert.Equal(t, "Use YYYY-MM-DD format", msg.Hint)
	})

	t.Run("NotFound with custom hint", func(t *testing.T) {
		msg := templates.NotFound("task", "missing-task", "Check the task list")

		assert.Equal(t, MessageError, msg.Type)
		assert.Contains(t, msg.Content, "Task 'missing-task' not found")
		assert.Equal(t, "Check the task list", msg.Hint)
	})

	t.Run("NotFound with default hint", func(t *testing.T) {
		msg := templates.NotFound("phase", "missing-phase", "")

		assert.Equal(t, MessageError, msg.Type)
		assert.Contains(t, msg.Content, "Phase 'missing-phase' not found")
		assert.Contains(t, msg.Hint, "agentpm query phases")
	})

	t.Run("InvalidState", func(t *testing.T) {
		msg := templates.InvalidState("task", "task-1", "completed", "active")

		assert.Equal(t, MessageError, msg.Type)
		assert.Contains(t, msg.Content, "Cannot perform action on task 'task-1'")
		assert.Contains(t, msg.Content, "Current state: completed, required: active")
		assert.Contains(t, msg.Hint, "First transition task to active state")
	})

	t.Run("DependencyBlocked with custom hint", func(t *testing.T) {
		blockers := []string{"task-1", "task-2"}
		msg := templates.DependencyBlocked("phase", "deployment", blockers, "Complete prerequisites first")

		assert.Equal(t, MessageError, msg.Type)
		assert.Contains(t, msg.Content, "Cannot proceed with phase 'deployment'")
		assert.Contains(t, msg.Content, "task-1, task-2")
		assert.Equal(t, "Complete prerequisites first", msg.Hint)
	})

	t.Run("DependencyBlocked with default hint", func(t *testing.T) {
		blockers := []string{"test-1"}
		msg := templates.DependencyBlocked("epic", "release", blockers, "")

		assert.Equal(t, MessageError, msg.Type)
		assert.Contains(t, msg.Content, "Cannot proceed with epic 'release'")
		assert.Contains(t, msg.Content, "test-1")
		assert.Contains(t, msg.Hint, "Complete the blocking dependencies")
	})
}

func TestGlobalTemplates(t *testing.T) {
	t.Run("GlobalTemplates is accessible", func(t *testing.T) {
		assert.NotNil(t, GlobalTemplates)

		// Test that it works
		msg := GlobalTemplates.AlreadyStarted("test", "global-test")
		assert.Equal(t, MessageInfo, msg.Type)
		assert.Contains(t, msg.Content, "Test 'global-test' is already started")
	})
}

// Performance tests
func BenchmarkMessageOutput(b *testing.B) {
	output := NewMessageOutput(FormatText)
	msg := InfoMessageWithHint("Test message", "Test hint")

	b.Run("Format", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			output.Format(msg)
		}
	})

	b.Run("FormatMultiple", func(b *testing.B) {
		msgs := []*Message{msg, msg, msg}
		for i := 0; i < b.N; i++ {
			output.FormatMultiple(msgs)
		}
	})
}
