package messages

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageType(t *testing.T) {
	t.Run("String representation", func(t *testing.T) {
		assert.Equal(t, "error", MessageError.String())
		assert.Equal(t, "warning", MessageWarning.String())
		assert.Equal(t, "info", MessageInfo.String())
		assert.Equal(t, "success", MessageSuccess.String())
	})

	t.Run("Validation", func(t *testing.T) {
		assert.True(t, MessageError.IsValid())
		assert.True(t, MessageWarning.IsValid())
		assert.True(t, MessageInfo.IsValid())
		assert.True(t, MessageSuccess.IsValid())
		assert.False(t, MessageType("invalid").IsValid())
	})
}

func TestMessage(t *testing.T) {
	t.Run("NewMessage", func(t *testing.T) {
		msg := NewMessage(MessageInfo, "Test message")

		assert.Equal(t, MessageInfo, msg.Type)
		assert.Equal(t, "Test message", msg.Content)
		assert.Empty(t, msg.Hint)
	})

	t.Run("NewMessageWithHint", func(t *testing.T) {
		msg := NewMessageWithHint(MessageError, "Error occurred", "Try this solution")

		assert.Equal(t, MessageError, msg.Type)
		assert.Equal(t, "Error occurred", msg.Content)
		assert.Equal(t, "Try this solution", msg.Hint)
	})

	t.Run("WithHint", func(t *testing.T) {
		msg := NewMessage(MessageWarning, "Warning message")
		result := msg.WithHint("Consider this")

		assert.Same(t, msg, result) // Should return the same instance
		assert.Equal(t, "Consider this", msg.Hint)
	})

	t.Run("String representation", func(t *testing.T) {
		testCases := []struct {
			name     string
			msg      *Message
			expected string
		}{
			{
				name:     "message without hint",
				msg:      NewMessage(MessageInfo, "Information message"),
				expected: "[info] Information message",
			},
			{
				name:     "message with hint",
				msg:      NewMessageWithHint(MessageError, "Error message", "Fix this way"),
				expected: "[error] Error message\nHint: Fix this way",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				assert.Equal(t, tc.expected, tc.msg.String())
			})
		}
	})

	t.Run("ToJSON", func(t *testing.T) {
		msg := NewMessageWithHint(MessageSuccess, "Operation completed", "Next steps available")

		jsonStr, err := msg.ToJSON()
		require.NoError(t, err)

		// Verify it's valid JSON
		var parsed map[string]interface{}
		err = json.Unmarshal([]byte(jsonStr), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "success", parsed["type"])
		assert.Equal(t, "Operation completed", parsed["content"])
		assert.Equal(t, "Next steps available", parsed["hint"])
	})

	t.Run("ToXML", func(t *testing.T) {
		msg := NewMessageWithHint(MessageWarning, "Warning occurred", "Please review")

		xmlStr, err := msg.ToXML()
		require.NoError(t, err)

		// Verify it's valid XML
		var parsed Message
		err = xml.Unmarshal([]byte(xmlStr), &parsed)
		require.NoError(t, err)

		assert.Equal(t, MessageWarning, parsed.Type)
		assert.Equal(t, "Warning occurred", parsed.Content)
		assert.Equal(t, "Please review", parsed.Hint)
	})
}

func TestDefaultMessageFormatter(t *testing.T) {
	t.Run("NewMessageFormatter", func(t *testing.T) {
		formatter := NewMessageFormatter()
		assert.NotNil(t, formatter)

		// Should implement the interface
		var _ MessageFormatter = formatter
	})

	t.Run("NewMessageFormatterWithConfig", func(t *testing.T) {
		config := FormatterConfig{
			ShowType:   false,
			ShowHints:  false,
			UseColors:  true,
			IndentJSON: false,
			IndentXML:  false,
		}

		formatter := NewMessageFormatterWithConfig(config)
		assert.NotNil(t, formatter)

		defaultFormatter := formatter.(*DefaultMessageFormatter)
		assert.Equal(t, config, defaultFormatter.config)
	})

	t.Run("FormatText with default config", func(t *testing.T) {
		formatter := NewMessageFormatter()

		testCases := []struct {
			name     string
			msg      *Message
			expected string
		}{
			{
				name:     "info message without hint",
				msg:      InfoMessage("Information here"),
				expected: "[info] Information here",
			},
			{
				name:     "error message with hint",
				msg:      ErrorMessageWithHint("Something failed", "Try restarting"),
				expected: "[error] Something failed\nHint: Try restarting",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := formatter.FormatText(tc.msg)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("FormatText without type display", func(t *testing.T) {
		config := DefaultFormatterConfig()
		config.ShowType = false
		formatter := NewMessageFormatterWithConfig(config)

		msg := InfoMessage("Information here")
		result := formatter.FormatText(msg)
		assert.Equal(t, "Information here", result)
	})

	t.Run("FormatText without hints", func(t *testing.T) {
		config := DefaultFormatterConfig()
		config.ShowHints = false
		formatter := NewMessageFormatterWithConfig(config)

		msg := ErrorMessageWithHint("Error occurred", "Fix this")
		result := formatter.FormatText(msg)
		assert.Equal(t, "[error] Error occurred", result)
	})

	t.Run("FormatJSON", func(t *testing.T) {
		formatter := NewMessageFormatter()
		msg := SuccessMessageWithHint("Done", "What's next")

		jsonStr, err := formatter.FormatJSON(msg)
		require.NoError(t, err)

		// Should be indented by default
		assert.Contains(t, jsonStr, "\n")
		assert.Contains(t, jsonStr, "  ")

		// Verify content
		var parsed Message
		err = json.Unmarshal([]byte(jsonStr), &parsed)
		require.NoError(t, err)
		assert.Equal(t, MessageSuccess, parsed.Type)
		assert.Equal(t, "Done", parsed.Content)
		assert.Equal(t, "What's next", parsed.Hint)
	})

	t.Run("FormatJSON without indentation", func(t *testing.T) {
		config := DefaultFormatterConfig()
		config.IndentJSON = false
		formatter := NewMessageFormatterWithConfig(config)

		msg := InfoMessage("Test")
		jsonStr, err := formatter.FormatJSON(msg)
		require.NoError(t, err)

		// Should not be indented
		assert.NotContains(t, jsonStr, "\n  ")
	})

	t.Run("FormatXML", func(t *testing.T) {
		formatter := NewMessageFormatter()
		msg := WarningMessage("Be careful")

		xmlStr, err := formatter.FormatXML(msg)
		require.NoError(t, err)

		// Should be valid XML
		assert.Contains(t, xmlStr, `type="warning"`)

		// Verify content
		var parsed Message
		err = xml.Unmarshal([]byte(xmlStr), &parsed)
		require.NoError(t, err)
		assert.Equal(t, MessageWarning, parsed.Type)
		assert.Equal(t, "Be careful", parsed.Content)
	})

	t.Run("Format delegates to FormatText", func(t *testing.T) {
		formatter := NewMessageFormatter()
		msg := InfoMessage("Test message")

		textResult := formatter.FormatText(msg)
		formatResult := formatter.Format(msg)

		assert.Equal(t, textResult, formatResult)
	})
}

func TestFormatterConfig(t *testing.T) {
	t.Run("DefaultFormatterConfig", func(t *testing.T) {
		config := DefaultFormatterConfig()

		assert.True(t, config.ShowType)
		assert.True(t, config.ShowHints)
		assert.False(t, config.UseColors) // Disabled by default
		assert.True(t, config.IndentJSON)
		assert.True(t, config.IndentXML)
	})
}

func TestPredefinedMessageCreators(t *testing.T) {
	t.Run("ErrorMessage", func(t *testing.T) {
		msg := ErrorMessage("Error content")
		assert.Equal(t, MessageError, msg.Type)
		assert.Equal(t, "Error content", msg.Content)
		assert.Empty(t, msg.Hint)
	})

	t.Run("ErrorMessageWithHint", func(t *testing.T) {
		msg := ErrorMessageWithHint("Error content", "Error hint")
		assert.Equal(t, MessageError, msg.Type)
		assert.Equal(t, "Error content", msg.Content)
		assert.Equal(t, "Error hint", msg.Hint)
	})

	t.Run("WarningMessage", func(t *testing.T) {
		msg := WarningMessage("Warning content")
		assert.Equal(t, MessageWarning, msg.Type)
		assert.Equal(t, "Warning content", msg.Content)
		assert.Empty(t, msg.Hint)
	})

	t.Run("WarningMessageWithHint", func(t *testing.T) {
		msg := WarningMessageWithHint("Warning content", "Warning hint")
		assert.Equal(t, MessageWarning, msg.Type)
		assert.Equal(t, "Warning content", msg.Content)
		assert.Equal(t, "Warning hint", msg.Hint)
	})

	t.Run("InfoMessage", func(t *testing.T) {
		msg := InfoMessage("Info content")
		assert.Equal(t, MessageInfo, msg.Type)
		assert.Equal(t, "Info content", msg.Content)
		assert.Empty(t, msg.Hint)
	})

	t.Run("InfoMessageWithHint", func(t *testing.T) {
		msg := InfoMessageWithHint("Info content", "Info hint")
		assert.Equal(t, MessageInfo, msg.Type)
		assert.Equal(t, "Info content", msg.Content)
		assert.Equal(t, "Info hint", msg.Hint)
	})

	t.Run("SuccessMessage", func(t *testing.T) {
		msg := SuccessMessage("Success content")
		assert.Equal(t, MessageSuccess, msg.Type)
		assert.Equal(t, "Success content", msg.Content)
		assert.Empty(t, msg.Hint)
	})

	t.Run("SuccessMessageWithHint", func(t *testing.T) {
		msg := SuccessMessageWithHint("Success content", "Success hint")
		assert.Equal(t, MessageSuccess, msg.Type)
		assert.Equal(t, "Success content", msg.Content)
		assert.Equal(t, "Success hint", msg.Hint)
	})
}

// Benchmark tests for performance validation
func BenchmarkMessageFormatting(b *testing.B) {
	formatter := NewMessageFormatter()
	msg := InfoMessageWithHint("Test message", "Test hint")

	b.Run("FormatText", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			formatter.FormatText(msg)
		}
	})

	b.Run("FormatJSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			formatter.FormatJSON(msg)
		}
	})

	b.Run("FormatXML", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			formatter.FormatXML(msg)
		}
	})
}
