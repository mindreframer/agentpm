package messages

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
)

// MessageType represents different severity levels for messages
type MessageType string

const (
	MessageError   MessageType = "error"
	MessageWarning MessageType = "warning"
	MessageInfo    MessageType = "info"
	MessageSuccess MessageType = "success"
)

// String returns the string representation of MessageType
func (mt MessageType) String() string {
	return string(mt)
}

// IsValid checks if the MessageType is valid
func (mt MessageType) IsValid() bool {
	switch mt {
	case MessageError, MessageWarning, MessageInfo, MessageSuccess:
		return true
	default:
		return false
	}
}

// Message represents a user-facing message with type, content, and optional hint
type Message struct {
	Type    MessageType `json:"type" xml:"type,attr"`
	Content string      `json:"content" xml:",chardata"`
	Hint    string      `json:"hint,omitempty" xml:"hint,omitempty"`
}

// NewMessage creates a new message with the specified type and content
func NewMessage(msgType MessageType, content string) *Message {
	return &Message{
		Type:    msgType,
		Content: content,
	}
}

// NewMessageWithHint creates a new message with type, content, and hint
func NewMessageWithHint(msgType MessageType, content, hint string) *Message {
	return &Message{
		Type:    msgType,
		Content: content,
		Hint:    hint,
	}
}

// WithHint adds a hint to an existing message
func (m *Message) WithHint(hint string) *Message {
	m.Hint = hint
	return m
}

// String returns a formatted string representation of the message
func (m *Message) String() string {
	if m.Hint != "" {
		return fmt.Sprintf("[%s] %s\nHint: %s", m.Type, m.Content, m.Hint)
	}
	return fmt.Sprintf("[%s] %s", m.Type, m.Content)
}

// ToJSON returns the JSON representation of the message
func (m *Message) ToJSON() (string, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message to JSON: %w", err)
	}
	return string(data), nil
}

// ToXML returns the XML representation of the message
func (m *Message) ToXML() (string, error) {
	data, err := xml.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message to XML: %w", err)
	}
	return string(data), nil
}

// MessageFormatter interface for consistent message formatting across different output formats
type MessageFormatter interface {
	Format(msg *Message) string
	FormatText(msg *Message) string
	FormatJSON(msg *Message) (string, error)
	FormatXML(msg *Message) (string, error)
}

// DefaultMessageFormatter implements MessageFormatter with standard formatting
type DefaultMessageFormatter struct {
	config FormatterConfig
}

// FormatterConfig holds configuration for message formatting
type FormatterConfig struct {
	// ShowType controls whether message type is displayed
	ShowType bool
	// ShowHints controls whether hints are displayed
	ShowHints bool
	// UseColors controls whether colors are used in text output
	UseColors bool
	// IndentJSON controls whether JSON output is indented
	IndentJSON bool
	// IndentXML controls whether XML output is indented
	IndentXML bool
}

// DefaultFormatterConfig returns a default configuration for message formatting
func DefaultFormatterConfig() FormatterConfig {
	return FormatterConfig{
		ShowType:   true,
		ShowHints:  true,
		UseColors:  false, // Disabled by default for compatibility
		IndentJSON: true,
		IndentXML:  true,
	}
}

// NewMessageFormatter creates a new message formatter with default configuration
func NewMessageFormatter() MessageFormatter {
	return &DefaultMessageFormatter{
		config: DefaultFormatterConfig(),
	}
}

// NewMessageFormatterWithConfig creates a new message formatter with custom configuration
func NewMessageFormatterWithConfig(config FormatterConfig) MessageFormatter {
	return &DefaultMessageFormatter{
		config: config,
	}
}

// Format returns a formatted message using the default text format
func (f *DefaultMessageFormatter) Format(msg *Message) string {
	return f.FormatText(msg)
}

// FormatText returns a text-formatted message
func (f *DefaultMessageFormatter) FormatText(msg *Message) string {
	var result string

	if f.config.ShowType {
		if f.config.UseColors {
			result = f.colorizeType(msg.Type) + " " + msg.Content
		} else {
			result = fmt.Sprintf("[%s] %s", msg.Type, msg.Content)
		}
	} else {
		result = msg.Content
	}

	if f.config.ShowHints && msg.Hint != "" {
		if f.config.UseColors {
			result += "\n" + f.colorizeHint(msg.Hint)
		} else {
			result += "\nHint: " + msg.Hint
		}
	}

	return result
}

// FormatJSON returns a JSON-formatted message
func (f *DefaultMessageFormatter) FormatJSON(msg *Message) (string, error) {
	if f.config.IndentJSON {
		data, err := json.MarshalIndent(msg, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal message to indented JSON: %w", err)
		}
		return string(data), nil
	}
	return msg.ToJSON()
}

// FormatXML returns an XML-formatted message
func (f *DefaultMessageFormatter) FormatXML(msg *Message) (string, error) {
	if f.config.IndentXML {
		data, err := xml.MarshalIndent(msg, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal message to indented XML: %w", err)
		}
		return string(data), nil
	}
	return msg.ToXML()
}

// colorizeType adds color codes to message types (for terminals that support colors)
func (f *DefaultMessageFormatter) colorizeType(msgType MessageType) string {
	switch msgType {
	case MessageError:
		return "\033[31m[ERROR]\033[0m" // Red
	case MessageWarning:
		return "\033[33m[WARNING]\033[0m" // Yellow
	case MessageInfo:
		return "\033[34m[INFO]\033[0m" // Blue
	case MessageSuccess:
		return "\033[32m[SUCCESS]\033[0m" // Green
	default:
		return fmt.Sprintf("[%s]", msgType)
	}
}

// colorizeHint adds color codes to hints
func (f *DefaultMessageFormatter) colorizeHint(hint string) string {
	return "\033[36mHint: " + hint + "\033[0m" // Cyan
}

// Predefined message creators for common scenarios

// ErrorMessage creates an error message
func ErrorMessage(content string) *Message {
	return NewMessage(MessageError, content)
}

// ErrorMessageWithHint creates an error message with a hint
func ErrorMessageWithHint(content, hint string) *Message {
	return NewMessageWithHint(MessageError, content, hint)
}

// WarningMessage creates a warning message
func WarningMessage(content string) *Message {
	return NewMessage(MessageWarning, content)
}

// WarningMessageWithHint creates a warning message with a hint
func WarningMessageWithHint(content, hint string) *Message {
	return NewMessageWithHint(MessageWarning, content, hint)
}

// InfoMessage creates an info message
func InfoMessage(content string) *Message {
	return NewMessage(MessageInfo, content)
}

// InfoMessageWithHint creates an info message with a hint
func InfoMessageWithHint(content, hint string) *Message {
	return NewMessageWithHint(MessageInfo, content, hint)
}

// SuccessMessage creates a success message
func SuccessMessage(content string) *Message {
	return NewMessage(MessageSuccess, content)
}

// SuccessMessageWithHint creates a success message with a hint
func SuccessMessageWithHint(content, hint string) *Message {
	return NewMessageWithHint(MessageSuccess, content, hint)
}
