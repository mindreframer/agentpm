package messages

import (
	"fmt"
	"strings"
)

// OutputFormat represents different output formats supported by the application
type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
	FormatXML  OutputFormat = "xml"
)

// String returns the string representation of OutputFormat
func (of OutputFormat) String() string {
	return string(of)
}

// IsValid checks if the OutputFormat is valid
func (of OutputFormat) IsValid() bool {
	switch of {
	case FormatText, FormatJSON, FormatXML:
		return true
	default:
		return false
	}
}

// MessageOutput handles message formatting and output for different formats
type MessageOutput struct {
	formatter MessageFormatter
	format    OutputFormat
}

// NewMessageOutput creates a new message output handler
func NewMessageOutput(format OutputFormat) *MessageOutput {
	return &MessageOutput{
		formatter: NewMessageFormatter(),
		format:    format,
	}
}

// NewMessageOutputWithFormatter creates a new message output handler with custom formatter
func NewMessageOutputWithFormatter(format OutputFormat, formatter MessageFormatter) *MessageOutput {
	return &MessageOutput{
		formatter: formatter,
		format:    format,
	}
}

// Format formats a message according to the configured output format
func (mo *MessageOutput) Format(msg *Message) (string, error) {
	switch mo.format {
	case FormatText:
		return mo.formatter.FormatText(msg), nil
	case FormatJSON:
		return mo.formatter.FormatJSON(msg)
	case FormatXML:
		return mo.formatter.FormatXML(msg)
	default:
		return "", fmt.Errorf("unsupported output format: %s", mo.format)
	}
}

// FormatMultiple formats multiple messages according to the configured output format
func (mo *MessageOutput) FormatMultiple(msgs []*Message) (string, error) {
	if len(msgs) == 0 {
		return "", nil
	}

	switch mo.format {
	case FormatText:
		var result []string
		for _, msg := range msgs {
			result = append(result, mo.formatter.FormatText(msg))
		}
		return strings.Join(result, "\n"), nil
	case FormatJSON:
		// For JSON, wrap multiple messages in an array
		var jsonMsgs []string
		for _, msg := range msgs {
			jsonMsg, err := mo.formatter.FormatJSON(msg)
			if err != nil {
				return "", err
			}
			jsonMsgs = append(jsonMsgs, jsonMsg)
		}
		return fmt.Sprintf("[\n%s\n]", strings.Join(jsonMsgs, ",\n")), nil
	case FormatXML:
		// For XML, wrap multiple messages in a root element
		var xmlMsgs []string
		for _, msg := range msgs {
			xmlMsg, err := mo.formatter.FormatXML(msg)
			if err != nil {
				return "", err
			}
			xmlMsgs = append(xmlMsgs, xmlMsg)
		}
		return fmt.Sprintf("<messages>\n%s\n</messages>", strings.Join(xmlMsgs, "\n")), nil
	default:
		return "", fmt.Errorf("unsupported output format: %s", mo.format)
	}
}

// SetFormat changes the output format
func (mo *MessageOutput) SetFormat(format OutputFormat) {
	mo.format = format
}

// GetFormat returns the current output format
func (mo *MessageOutput) GetFormat() OutputFormat {
	return mo.format
}

// FriendlyResponseTemplates contains templates for friendly responses to common scenarios
type FriendlyResponseTemplates struct{}

// NewFriendlyResponseTemplates creates a new friendly response template handler
func NewFriendlyResponseTemplates() *FriendlyResponseTemplates {
	return &FriendlyResponseTemplates{}
}

// AlreadyStarted returns a friendly message for entities that are already started
func (frt *FriendlyResponseTemplates) AlreadyStarted(entityType, entityID string) *Message {
	content := fmt.Sprintf("%s '%s' is already started. No action needed.",
		strings.Title(entityType), entityID)
	hint := fmt.Sprintf("You can check the current status with: agentpm status")
	return InfoMessageWithHint(content, hint)
}

// AlreadyCompleted returns a friendly message for entities that are already completed
func (frt *FriendlyResponseTemplates) AlreadyCompleted(entityType, entityID string) *Message {
	content := fmt.Sprintf("%s '%s' is already completed. No action needed.",
		strings.Title(entityType), entityID)
	hint := fmt.Sprintf("You can view completed %ss with: agentpm status", strings.ToLower(entityType))
	return InfoMessageWithHint(content, hint)
}

// AlreadyInState returns a friendly message for entities that are already in the target state
func (frt *FriendlyResponseTemplates) AlreadyInState(entityType, entityID, state string) *Message {
	content := fmt.Sprintf("%s '%s' is already %s. No action needed.",
		strings.Title(entityType), entityID, state)
	return InfoMessage(content)
}

// SuccessfulTransition returns a success message for successful state transitions
func (frt *FriendlyResponseTemplates) SuccessfulTransition(entityType, entityID, action string) *Message {
	content := fmt.Sprintf("%s '%s' %s successfully.",
		strings.Title(entityType), entityID, action)
	return SuccessMessage(content)
}

// OperationComplete returns a success message for completed operations
func (frt *FriendlyResponseTemplates) OperationComplete(operation, details string) *Message {
	if details != "" {
		content := fmt.Sprintf("%s completed: %s", operation, details)
		return SuccessMessage(content)
	}
	content := fmt.Sprintf("%s completed successfully.", operation)
	return SuccessMessage(content)
}

// ValidationError returns an error message with helpful hints for validation failures
func (frt *FriendlyResponseTemplates) ValidationError(issue, hint string) *Message {
	return ErrorMessageWithHint(issue, hint)
}

// NotFound returns an error message for missing entities with helpful hints
func (frt *FriendlyResponseTemplates) NotFound(entityType, entityID, hint string) *Message {
	content := fmt.Sprintf("%s '%s' not found.", strings.Title(entityType), entityID)
	if hint == "" {
		hint = fmt.Sprintf("List available %ss with: agentpm query %ss",
			strings.ToLower(entityType), strings.ToLower(entityType))
	}
	return ErrorMessageWithHint(content, hint)
}

// InvalidState returns an error message for invalid state transitions
func (frt *FriendlyResponseTemplates) InvalidState(entityType, entityID, currentState, requiredState string) *Message {
	content := fmt.Sprintf("Cannot perform action on %s '%s'. Current state: %s, required: %s.",
		strings.ToLower(entityType), entityID, currentState, requiredState)
	hint := fmt.Sprintf("First transition %s to %s state before proceeding.",
		strings.ToLower(entityType), requiredState)
	return ErrorMessageWithHint(content, hint)
}

// DependencyBlocked returns an error message for dependency-related blocks
func (frt *FriendlyResponseTemplates) DependencyBlocked(entityType, entityID string, blockers []string, actionHint string) *Message {
	content := fmt.Sprintf("Cannot proceed with %s '%s'. Blocked by: %s.",
		strings.ToLower(entityType), entityID, strings.Join(blockers, ", "))

	var hint string
	if actionHint != "" {
		hint = actionHint
	} else {
		hint = "Complete the blocking dependencies before proceeding."
	}
	return ErrorMessageWithHint(content, hint)
}

// GlobalTemplates provides a global instance of friendly response templates
var GlobalTemplates = NewFriendlyResponseTemplates()
