package messages

import (
	"fmt"
	"strings"
)

// MessageTemplates provides predefined message templates for common scenarios
type MessageTemplates struct {
	formatter MessageFormatter
}

// NewMessageTemplates creates a new message templates handler
func NewMessageTemplates() *MessageTemplates {
	return &MessageTemplates{
		formatter: NewMessageFormatter(),
	}
}

// NewMessageTemplatesWithFormatter creates a new message templates handler with custom formatter
func NewMessageTemplatesWithFormatter(formatter MessageFormatter) *MessageTemplates {
	return &MessageTemplates{
		formatter: formatter,
	}
}

// Phase-related templates

// PhaseAlreadyActive returns a friendly message when trying to start an already active phase
func (mt *MessageTemplates) PhaseAlreadyActive(phaseID string) *Message {
	content := fmt.Sprintf("Phase '%s' is already active. No action needed.", phaseID)
	hint := "You can check the current phase with: agentpm status"
	return SuccessMessageWithHint(content, hint)
}

// PhaseAlreadyCompleted returns a friendly message when trying to complete an already completed phase
func (mt *MessageTemplates) PhaseAlreadyCompleted(phaseID string) *Message {
	content := fmt.Sprintf("Phase '%s' is already completed. No action needed.", phaseID)
	hint := "You can view completed phases with: agentpm status"
	return SuccessMessageWithHint(content, hint)
}

// PhaseStarted returns a success message for phase start
func (mt *MessageTemplates) PhaseStarted(phaseID string) *Message {
	content := fmt.Sprintf("Phase '%s' started successfully.", phaseID)
	return SuccessMessage(content)
}

// PhaseCompleted returns a success message for phase completion
func (mt *MessageTemplates) PhaseCompleted(phaseID string) *Message {
	content := fmt.Sprintf("Phase '%s' completed successfully.", phaseID)
	return SuccessMessage(content)
}

// PhaseConflict returns an error message for phase conflicts
func (mt *MessageTemplates) PhaseConflict(phaseID, activePhaseID string) *Message {
	content := fmt.Sprintf("Cannot start phase '%s'. Phase '%s' is currently active.", phaseID, activePhaseID)
	hint := fmt.Sprintf("Complete the active phase first with: agentpm complete-phase %s", activePhaseID)
	return ErrorMessageWithHint(content, hint)
}

// PhaseIncompleteDependencies returns an error message for incomplete dependencies
func (mt *MessageTemplates) PhaseIncompleteDependencies(phaseID string, incompleteTasks, incompleteTests []string) *Message {
	var blockers []string
	var hints []string

	if len(incompleteTasks) > 0 {
		blockers = append(blockers, fmt.Sprintf("tasks: %s", strings.Join(incompleteTasks, ", ")))
		hints = append(hints, "Complete tasks with: agentpm complete-task <task-id>")
	}

	if len(incompleteTests) > 0 {
		blockers = append(blockers, fmt.Sprintf("tests: %s", strings.Join(incompleteTests, ", ")))
		hints = append(hints, "Complete tests with: agentpm complete-test <test-id>")
	}

	content := fmt.Sprintf("Cannot complete phase '%s'. Incomplete dependencies: %s.",
		phaseID, strings.Join(blockers, "; "))
	hint := strings.Join(hints, " | ")

	return ErrorMessageWithHint(content, hint)
}

// Task-related templates

// TaskAlreadyActive returns a friendly message when trying to start an already active task
func (mt *MessageTemplates) TaskAlreadyActive(taskID string) *Message {
	content := fmt.Sprintf("Task '%s' is already active. No action needed.", taskID)
	hint := "You can check the current task with: agentpm status"
	return SuccessMessageWithHint(content, hint)
}

// TaskAlreadyCompleted returns a friendly message when trying to complete an already completed task
func (mt *MessageTemplates) TaskAlreadyCompleted(taskID string) *Message {
	content := fmt.Sprintf("Task '%s' is already completed. No action needed.", taskID)
	hint := "You can view completed tasks with: agentpm status"
	return SuccessMessageWithHint(content, hint)
}

// TaskStarted returns a success message for task start
func (mt *MessageTemplates) TaskStarted(taskID string) *Message {
	content := fmt.Sprintf("Task '%s' started successfully.", taskID)
	return SuccessMessage(content)
}

// TaskCompleted returns a success message for task completion
func (mt *MessageTemplates) TaskCompleted(taskID string) *Message {
	content := fmt.Sprintf("Task '%s' completed successfully.", taskID)
	return SuccessMessage(content)
}

// Test-related templates

// TestAlreadyActive returns a friendly message when trying to start an already active test
func (mt *MessageTemplates) TestAlreadyActive(testID string) *Message {
	content := fmt.Sprintf("Test '%s' is already started. No action needed.", testID)
	hint := "You can check the current test status with: agentpm status"
	return SuccessMessageWithHint(content, hint)
}

// TestAlreadyPassed returns a friendly message when a test is already passed
func (mt *MessageTemplates) TestAlreadyPassed(testID string) *Message {
	content := fmt.Sprintf("Test '%s' has already passed. No action needed.", testID)
	hint := "You can view test results with: agentpm status"
	return SuccessMessageWithHint(content, hint)
}

// TestStarted returns a success message for test start
func (mt *MessageTemplates) TestStarted(testID string) *Message {
	content := fmt.Sprintf("Test '%s' started successfully.", testID)
	return SuccessMessage(content)
}

// TestPassed returns a success message for test pass
func (mt *MessageTemplates) TestPassed(testID string) *Message {
	content := fmt.Sprintf("Test '%s' passed successfully.", testID)
	return SuccessMessage(content)
}

// TestFailed returns an error message for test failure
func (mt *MessageTemplates) TestFailed(testID, reason string) *Message {
	content := fmt.Sprintf("Test '%s' failed: %s", testID, reason)
	hint := "Review the test requirements and fix the implementation."
	return ErrorMessageWithHint(content, hint)
}

// TestCancelled returns an info message for test cancellation
func (mt *MessageTemplates) TestCancelled(testID, reason string) *Message {
	content := fmt.Sprintf("Test '%s' cancelled: %s", testID, reason)
	return InfoMessage(content)
}

// Epic-related templates

// EpicAlreadyStarted returns a friendly message when trying to start an already started epic
func (mt *MessageTemplates) EpicAlreadyStarted(epicID string) *Message {
	content := fmt.Sprintf("Epic '%s' is already started. No action needed.", epicID)
	hint := "You can check the epic status with: agentpm status"
	return SuccessMessageWithHint(content, hint)
}

// EpicAlreadyCompleted returns a friendly message when trying to complete an already completed epic
func (mt *MessageTemplates) EpicAlreadyCompleted(epicID string) *Message {
	content := fmt.Sprintf("Epic '%s' is already completed. No action needed.", epicID)
	hint := "You can view epic summary with: agentpm status"
	return SuccessMessageWithHint(content, hint)
}

// EpicStarted returns a success message for epic start
func (mt *MessageTemplates) EpicStarted(epicID string) *Message {
	content := fmt.Sprintf("Epic '%s' started successfully.", epicID)
	return SuccessMessage(content)
}

// EpicCompleted returns a success message for epic completion
func (mt *MessageTemplates) EpicCompleted(epicID string) *Message {
	content := fmt.Sprintf("Epic '%s' completed successfully.", epicID)
	return SuccessMessage(content)
}

// Generic entity templates for flexibility

// EntityNotFound returns an error message for missing entities
func (mt *MessageTemplates) EntityNotFound(entityType, entityID string) *Message {
	content := fmt.Sprintf("%s '%s' not found.", strings.Title(entityType), entityID)
	hint := fmt.Sprintf("List available %ss with: agentpm query %ss",
		strings.ToLower(entityType), strings.ToLower(entityType))
	return ErrorMessageWithHint(content, hint)
}

// InvalidEntityState returns an error message for invalid entity states
func (mt *MessageTemplates) InvalidEntityState(entityType, entityID, currentState, action string) *Message {
	content := fmt.Sprintf("Cannot %s %s '%s' in state '%s'.",
		action, strings.ToLower(entityType), entityID, currentState)
	hint := fmt.Sprintf("Check valid state transitions for %ss in the documentation.",
		strings.ToLower(entityType))
	return ErrorMessageWithHint(content, hint)
}

// OperationSuccess returns a success message for successful operations
func (mt *MessageTemplates) OperationSuccess(operation, target string) *Message {
	content := fmt.Sprintf("%s %s successfully.", strings.Title(operation), target)
	return SuccessMessage(content)
}

// ConfigurationError returns an error message for configuration issues
func (mt *MessageTemplates) ConfigurationError(issue, hint string) *Message {
	content := fmt.Sprintf("Configuration error: %s", issue)
	return ErrorMessageWithHint(content, hint)
}

// FileError returns an error message for file-related issues
func (mt *MessageTemplates) FileError(operation, filename, reason string) *Message {
	content := fmt.Sprintf("Failed to %s file '%s': %s", operation, filename, reason)
	hint := "Check file permissions and that the path exists."
	return ErrorMessageWithHint(content, hint)
}

// ValidationWarning returns a warning message for validation issues
func (mt *MessageTemplates) ValidationWarning(issue, suggestion string) *Message {
	content := fmt.Sprintf("Validation warning: %s", issue)
	hint := suggestion
	return WarningMessageWithHint(content, hint)
}

// GlobalMessageTemplates provides a global instance of message templates
var GlobalMessageTemplates = NewMessageTemplates()
