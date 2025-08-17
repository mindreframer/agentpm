package assertions

import (
	"fmt"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/testing/executor"
)

// AssertionBuilder provides a fluent API for asserting on TransitionChain results
type AssertionBuilder struct {
	result *executor.TransitionChainResult
	errors []AssertionError
}

// AssertionError represents a failed assertion with context
type AssertionError struct {
	Type        string
	Message     string
	Expected    interface{}
	Actual      interface{}
	Context     map[string]interface{}
	Suggestions []string
}

func (e AssertionError) Error() string {
	return e.Message
}

// NewAssertionBuilder creates a new assertion builder for the given result
func NewAssertionBuilder(result *executor.TransitionChainResult) *AssertionBuilder {
	return &AssertionBuilder{
		result: result,
		errors: make([]AssertionError, 0),
	}
}

// Assert creates a new assertion builder for the given result (fluent entry point)
func Assert(result *executor.TransitionChainResult) *AssertionBuilder {
	return NewAssertionBuilder(result)
}

// EpicStatus asserts the final epic status
func (ab *AssertionBuilder) EpicStatus(expectedStatus string) *AssertionBuilder {
	if ab.result.FinalState == nil {
		ab.addError("epic_status", "Final state is nil", expectedStatus, nil, nil)
		return ab
	}

	actualStatus := string(ab.result.FinalState.Status)
	if actualStatus != expectedStatus {
		ab.addError("epic_status",
			fmt.Sprintf("Expected epic status %s, got %s", expectedStatus, actualStatus),
			expectedStatus, actualStatus,
			map[string]interface{}{
				"epic_id": ab.result.FinalState.ID,
			})
	}
	return ab
}

// PhaseStatus asserts a specific phase status
func (ab *AssertionBuilder) PhaseStatus(phaseID, expectedStatus string) *AssertionBuilder {
	if ab.result.FinalState == nil {
		ab.addError("phase_status", "Final state is nil", expectedStatus, nil,
			map[string]interface{}{"phase_id": phaseID})
		return ab
	}

	phase := ab.findPhase(phaseID)
	if phase == nil {
		ab.addError("phase_status",
			fmt.Sprintf("Phase %s not found", phaseID),
			expectedStatus, nil,
			map[string]interface{}{"phase_id": phaseID})
		return ab
	}

	actualStatus := string(phase.Status)
	if actualStatus != expectedStatus {
		ab.addError("phase_status",
			fmt.Sprintf("Expected phase %s status %s, got %s", phaseID, expectedStatus, actualStatus),
			expectedStatus, actualStatus,
			map[string]interface{}{
				"phase_id":   phaseID,
				"phase_name": phase.Name,
			})
	}
	return ab
}

// TaskStatus asserts a specific task status
func (ab *AssertionBuilder) TaskStatus(taskID, expectedStatus string) *AssertionBuilder {
	if ab.result.FinalState == nil {
		ab.addError("task_status", "Final state is nil", expectedStatus, nil,
			map[string]interface{}{"task_id": taskID})
		return ab
	}

	task := ab.findTask(taskID)
	if task == nil {
		ab.addError("task_status",
			fmt.Sprintf("Task %s not found", taskID),
			expectedStatus, nil,
			map[string]interface{}{"task_id": taskID})
		return ab
	}

	actualStatus := string(task.Status)
	if actualStatus != expectedStatus {
		ab.addError("task_status",
			fmt.Sprintf("Expected task %s status %s, got %s", taskID, expectedStatus, actualStatus),
			expectedStatus, actualStatus,
			map[string]interface{}{
				"task_id":   taskID,
				"task_name": task.Name,
				"phase_id":  task.PhaseID,
			})
	}
	return ab
}

// TestStatus asserts a specific test status
func (ab *AssertionBuilder) TestStatus(testID, expectedStatus string) *AssertionBuilder {
	if ab.result.FinalState == nil {
		ab.addError("test_status", "Final state is nil", expectedStatus, nil,
			map[string]interface{}{"test_id": testID})
		return ab
	}

	test := ab.findTest(testID)
	if test == nil {
		ab.addError("test_status",
			fmt.Sprintf("Test %s not found", testID),
			expectedStatus, nil,
			map[string]interface{}{"test_id": testID})
		return ab
	}

	actualStatus := string(test.Status)
	if actualStatus != expectedStatus {
		ab.addError("test_status",
			fmt.Sprintf("Expected test %s status %s, got %s", testID, expectedStatus, actualStatus),
			expectedStatus, actualStatus,
			map[string]interface{}{
				"test_id":   testID,
				"test_name": test.Name,
				"task_id":   test.TaskID,
				"phase_id":  test.PhaseID,
			})
	}
	return ab
}

// TestStatusUnified asserts a specific test status using Epic 13 unified system
func (ab *AssertionBuilder) TestStatusUnified(testID, expectedStatus string) *AssertionBuilder {
	if ab.result.FinalState == nil {
		ab.addError("test_status_unified", "Final state is nil", expectedStatus, nil,
			map[string]interface{}{"test_id": testID})
		return ab
	}

	test := ab.findTest(testID)
	if test == nil {
		ab.addError("test_status_unified",
			fmt.Sprintf("Test %s not found", testID),
			expectedStatus, nil,
			map[string]interface{}{"test_id": testID})
		return ab
	}

	actualStatus := string(test.GetTestStatusUnified())
	if actualStatus != expectedStatus {
		ab.addError("test_status_unified",
			fmt.Sprintf("Expected test %s unified status %s, got %s", testID, expectedStatus, actualStatus),
			expectedStatus, actualStatus,
			map[string]interface{}{
				"test_id":       testID,
				"test_name":     test.Name,
				"legacy_status": string(test.Status),
				"test_result":   string(test.GetTestResult()),
			})
	}
	return ab
}

// TestResult asserts a specific test result using Epic 13 unified system
func (ab *AssertionBuilder) TestResult(testID, expectedResult string) *AssertionBuilder {
	if ab.result.FinalState == nil {
		ab.addError("test_result", "Final state is nil", expectedResult, nil,
			map[string]interface{}{"test_id": testID})
		return ab
	}

	test := ab.findTest(testID)
	if test == nil {
		ab.addError("test_result",
			fmt.Sprintf("Test %s not found", testID),
			expectedResult, nil,
			map[string]interface{}{"test_id": testID})
		return ab
	}

	actualResult := string(test.GetTestResult())
	if actualResult != expectedResult {
		ab.addError("test_result",
			fmt.Sprintf("Expected test %s result %s, got %s", testID, expectedResult, actualResult),
			expectedResult, actualResult,
			map[string]interface{}{
				"test_id":     testID,
				"test_name":   test.Name,
				"test_status": string(test.GetTestStatusUnified()),
			})
	}
	return ab
}

// HasEvent asserts that an event of the given type exists
func (ab *AssertionBuilder) HasEvent(eventType string) *AssertionBuilder {
	if ab.result.FinalState == nil {
		ab.addError("has_event", "Final state is nil", eventType, nil,
			map[string]interface{}{"event_type": eventType})
		return ab
	}

	found := false
	for _, event := range ab.result.FinalState.Events {
		if event.Type == eventType {
			found = true
			break
		}
	}

	if !found {
		ab.addError("has_event",
			fmt.Sprintf("Expected event type %s not found", eventType),
			eventType, nil,
			map[string]interface{}{
				"event_type":   eventType,
				"total_events": len(ab.result.FinalState.Events),
			})
	}
	return ab
}

// EventCount asserts the total number of events
func (ab *AssertionBuilder) EventCount(expectedCount int) *AssertionBuilder {
	if ab.result.FinalState == nil {
		ab.addError("event_count", "Final state is nil", expectedCount, nil, nil)
		return ab
	}

	actualCount := len(ab.result.FinalState.Events)
	if actualCount != expectedCount {
		ab.addError("event_count",
			fmt.Sprintf("Expected %d events, got %d", expectedCount, actualCount),
			expectedCount, actualCount,
			map[string]interface{}{
				"event_types": ab.getEventTypes(),
			})
	}
	return ab
}

// NoErrors asserts that the execution had no errors
func (ab *AssertionBuilder) NoErrors() *AssertionBuilder {
	if ab.result.Success {
		return ab
	}

	errorMessages := make([]string, len(ab.result.Errors))
	for i, err := range ab.result.Errors {
		errorMessages[i] = err.Error()
	}

	ab.addError("no_errors",
		fmt.Sprintf("Expected no errors, but got %d errors", len(ab.result.Errors)),
		0, len(ab.result.Errors),
		map[string]interface{}{
			"errors": errorMessages,
		})
	return ab
}

// HasErrors asserts that the execution had errors
func (ab *AssertionBuilder) HasErrors() *AssertionBuilder {
	if !ab.result.Success {
		return ab
	}

	ab.addError("has_errors",
		"Expected errors, but execution was successful",
		">0", 0, nil)
	return ab
}

// ErrorCount asserts the number of errors
func (ab *AssertionBuilder) ErrorCount(expectedCount int) *AssertionBuilder {
	actualCount := len(ab.result.Errors)
	if actualCount != expectedCount {
		ab.addError("error_count",
			fmt.Sprintf("Expected %d errors, got %d", expectedCount, actualCount),
			expectedCount, actualCount, nil)
	}
	return ab
}

// ExecutionTime asserts that execution time is within expected range
func (ab *AssertionBuilder) ExecutionTime(maxDuration time.Duration) *AssertionBuilder {
	if ab.result.ExecutionTime > maxDuration {
		ab.addError("execution_time",
			fmt.Sprintf("Expected execution time <= %v, got %v", maxDuration, ab.result.ExecutionTime),
			maxDuration, ab.result.ExecutionTime, nil)
	}
	return ab
}

// CommandCount asserts the number of executed commands
func (ab *AssertionBuilder) CommandCount(expectedCount int) *AssertionBuilder {
	actualCount := len(ab.result.ExecutedCommands)
	if actualCount != expectedCount {
		ab.addError("command_count",
			fmt.Sprintf("Expected %d commands, got %d", expectedCount, actualCount),
			expectedCount, actualCount,
			map[string]interface{}{
				"commands": ab.getCommandTypes(),
			})
	}
	return ab
}

// AllCommandsSuccessful asserts that all commands executed successfully
func (ab *AssertionBuilder) AllCommandsSuccessful() *AssertionBuilder {
	failedCommands := make([]string, 0)
	for i, cmd := range ab.result.ExecutedCommands {
		if !cmd.Success {
			failedCommands = append(failedCommands, fmt.Sprintf("Command %d: %s", i+1, cmd.Command.Type))
		}
	}

	if len(failedCommands) > 0 {
		ab.addError("all_commands_successful",
			fmt.Sprintf("Expected all commands to succeed, but %d failed", len(failedCommands)),
			0, len(failedCommands),
			map[string]interface{}{
				"failed_commands": failedCommands,
			})
	}
	return ab
}

// Check validates all assertions and returns any errors
func (ab *AssertionBuilder) Check() error {
	if len(ab.errors) == 0 {
		return nil
	}

	if len(ab.errors) == 1 {
		return ab.errors[0]
	}

	// Multiple errors - create a composite error
	return &CompositeAssertionError{
		Errors: ab.errors,
		Count:  len(ab.errors),
	}
}

// MustPass validates all assertions and panics if any fail (for test convenience)
func (ab *AssertionBuilder) MustPass() {
	if err := ab.Check(); err != nil {
		panic(fmt.Sprintf("Assertion failed: %v", err))
	}
}

// CompositeAssertionError represents multiple assertion failures
type CompositeAssertionError struct {
	Errors []AssertionError
	Count  int
}

func (e *CompositeAssertionError) Error() string {
	return fmt.Sprintf("%d assertion failures", e.Count)
}

// GetErrors returns all individual assertion errors
func (e *CompositeAssertionError) GetErrors() []AssertionError {
	return e.Errors
}

// Helper methods

func (ab *AssertionBuilder) addError(errorType, message string, expected, actual interface{}, context map[string]interface{}) {
	ab.errors = append(ab.errors, AssertionError{
		Type:     errorType,
		Message:  message,
		Expected: expected,
		Actual:   actual,
		Context:  context,
	})
}

func (ab *AssertionBuilder) findPhase(phaseID string) *epic.Phase {
	if ab.result.FinalState == nil {
		return nil
	}
	for i := range ab.result.FinalState.Phases {
		if ab.result.FinalState.Phases[i].ID == phaseID {
			return &ab.result.FinalState.Phases[i]
		}
	}
	return nil
}

func (ab *AssertionBuilder) findTask(taskID string) *epic.Task {
	if ab.result.FinalState == nil {
		return nil
	}
	for i := range ab.result.FinalState.Tasks {
		if ab.result.FinalState.Tasks[i].ID == taskID {
			return &ab.result.FinalState.Tasks[i]
		}
	}
	return nil
}

func (ab *AssertionBuilder) findTest(testID string) *epic.Test {
	if ab.result.FinalState == nil {
		return nil
	}
	for i := range ab.result.FinalState.Tests {
		if ab.result.FinalState.Tests[i].ID == testID {
			return &ab.result.FinalState.Tests[i]
		}
	}
	return nil
}

func (ab *AssertionBuilder) getEventTypes() []string {
	if ab.result.FinalState == nil {
		return nil
	}
	types := make([]string, len(ab.result.FinalState.Events))
	for i, event := range ab.result.FinalState.Events {
		types[i] = event.Type
	}
	return types
}

func (ab *AssertionBuilder) getCommandTypes() []string {
	types := make([]string, len(ab.result.ExecutedCommands))
	for i, cmd := range ab.result.ExecutedCommands {
		types[i] = cmd.Command.Type
	}
	return types
}
