package messages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageTemplates(t *testing.T) {
	templates := NewMessageTemplates()

	t.Run("NewMessageTemplates", func(t *testing.T) {
		assert.NotNil(t, templates)
		assert.NotNil(t, templates.formatter)
	})

	t.Run("NewMessageTemplatesWithFormatter", func(t *testing.T) {
		customFormatter := NewMessageFormatter()
		templates := NewMessageTemplatesWithFormatter(customFormatter)

		assert.NotNil(t, templates)
		assert.Same(t, customFormatter, templates.formatter)
	})
}

func TestPhaseTemplates(t *testing.T) {
	templates := NewMessageTemplates()

	t.Run("PhaseAlreadyActive", func(t *testing.T) {
		msg := templates.PhaseAlreadyActive("implementation")

		assert.Equal(t, MessageInfo, msg.Type)
		assert.Contains(t, msg.Content, "Phase 'implementation' is already active")
		assert.Contains(t, msg.Hint, "agentpm status")
	})

	t.Run("PhaseAlreadyCompleted", func(t *testing.T) {
		msg := templates.PhaseAlreadyCompleted("testing")

		assert.Equal(t, MessageInfo, msg.Type)
		assert.Contains(t, msg.Content, "Phase 'testing' is already completed")
		assert.Contains(t, msg.Hint, "agentpm status")
	})

	t.Run("PhaseStarted", func(t *testing.T) {
		msg := templates.PhaseStarted("deployment")

		assert.Equal(t, MessageSuccess, msg.Type)
		assert.Contains(t, msg.Content, "Phase 'deployment' started successfully")
	})

	t.Run("PhaseCompleted", func(t *testing.T) {
		msg := templates.PhaseCompleted("analysis")

		assert.Equal(t, MessageSuccess, msg.Type)
		assert.Contains(t, msg.Content, "Phase 'analysis' completed successfully")
	})

	t.Run("PhaseConflict", func(t *testing.T) {
		msg := templates.PhaseConflict("testing", "implementation")

		assert.Equal(t, MessageError, msg.Type)
		assert.Contains(t, msg.Content, "Cannot start phase 'testing'")
		assert.Contains(t, msg.Content, "Phase 'implementation' is currently active")
		assert.Contains(t, msg.Hint, "agentpm complete-phase implementation")
	})

	t.Run("PhaseIncompleteDependencies with tasks and tests", func(t *testing.T) {
		incompleteTasks := []string{"task-1", "task-2"}
		incompleteTests := []string{"test-1"}

		msg := templates.PhaseIncompleteDependencies("deployment", incompleteTasks, incompleteTests)

		assert.Equal(t, MessageError, msg.Type)
		assert.Contains(t, msg.Content, "Cannot complete phase 'deployment'")
		assert.Contains(t, msg.Content, "tasks: task-1, task-2")
		assert.Contains(t, msg.Content, "tests: test-1")
		assert.Contains(t, msg.Hint, "agentpm complete-task")
		assert.Contains(t, msg.Hint, "agentpm complete-test")
	})

	t.Run("PhaseIncompleteDependencies with only tasks", func(t *testing.T) {
		incompleteTasks := []string{"task-1"}

		msg := templates.PhaseIncompleteDependencies("review", incompleteTasks, []string{})

		assert.Equal(t, MessageError, msg.Type)
		assert.Contains(t, msg.Content, "Cannot complete phase 'review'")
		assert.Contains(t, msg.Content, "tasks: task-1")
		assert.NotContains(t, msg.Content, "tests:")
		assert.Contains(t, msg.Hint, "agentpm complete-task")
		assert.NotContains(t, msg.Hint, "agentpm complete-test")
	})
}

func TestTaskTemplates(t *testing.T) {
	templates := NewMessageTemplates()

	t.Run("TaskAlreadyActive", func(t *testing.T) {
		msg := templates.TaskAlreadyActive("setup-database")

		assert.Equal(t, MessageInfo, msg.Type)
		assert.Contains(t, msg.Content, "Task 'setup-database' is already active")
		assert.Contains(t, msg.Hint, "agentpm status")
	})

	t.Run("TaskAlreadyCompleted", func(t *testing.T) {
		msg := templates.TaskAlreadyCompleted("create-api")

		assert.Equal(t, MessageInfo, msg.Type)
		assert.Contains(t, msg.Content, "Task 'create-api' is already completed")
		assert.Contains(t, msg.Hint, "agentpm status")
	})

	t.Run("TaskStarted", func(t *testing.T) {
		msg := templates.TaskStarted("write-docs")

		assert.Equal(t, MessageSuccess, msg.Type)
		assert.Contains(t, msg.Content, "Task 'write-docs' started successfully")
	})

	t.Run("TaskCompleted", func(t *testing.T) {
		msg := templates.TaskCompleted("refactor-code")

		assert.Equal(t, MessageSuccess, msg.Type)
		assert.Contains(t, msg.Content, "Task 'refactor-code' completed successfully")
	})
}

func TestTestTemplates(t *testing.T) {
	templates := NewMessageTemplates()

	t.Run("TestAlreadyActive", func(t *testing.T) {
		msg := templates.TestAlreadyActive("integration-test-1")

		assert.Equal(t, MessageInfo, msg.Type)
		assert.Contains(t, msg.Content, "Test 'integration-test-1' is already started")
		assert.Contains(t, msg.Hint, "agentpm status")
	})

	t.Run("TestAlreadyPassed", func(t *testing.T) {
		msg := templates.TestAlreadyPassed("unit-test-auth")

		assert.Equal(t, MessageInfo, msg.Type)
		assert.Contains(t, msg.Content, "Test 'unit-test-auth' has already passed")
		assert.Contains(t, msg.Hint, "agentpm status")
	})

	t.Run("TestStarted", func(t *testing.T) {
		msg := templates.TestStarted("e2e-checkout")

		assert.Equal(t, MessageSuccess, msg.Type)
		assert.Contains(t, msg.Content, "Test 'e2e-checkout' started successfully")
	})

	t.Run("TestPassed", func(t *testing.T) {
		msg := templates.TestPassed("performance-test")

		assert.Equal(t, MessageSuccess, msg.Type)
		assert.Contains(t, msg.Content, "Test 'performance-test' passed successfully")
	})

	t.Run("TestFailed", func(t *testing.T) {
		msg := templates.TestFailed("ui-responsive", "Layout breaks on mobile")

		assert.Equal(t, MessageError, msg.Type)
		assert.Contains(t, msg.Content, "Test 'ui-responsive' failed: Layout breaks on mobile")
	})

	t.Run("TestCancelled", func(t *testing.T) {
		msg := templates.TestCancelled("load-test", "Environment unavailable")

		assert.Equal(t, MessageInfo, msg.Type)
		assert.Contains(t, msg.Content, "Test 'load-test' cancelled: Environment unavailable")
	})
}

func TestEpicTemplates(t *testing.T) {
	templates := NewMessageTemplates()

	t.Run("EpicAlreadyStarted", func(t *testing.T) {
		msg := templates.EpicAlreadyStarted("user-authentication")

		assert.Equal(t, MessageInfo, msg.Type)
		assert.Contains(t, msg.Content, "Epic 'user-authentication' is already started")
		assert.Contains(t, msg.Hint, "agentpm status")
	})

	t.Run("EpicAlreadyCompleted", func(t *testing.T) {
		msg := templates.EpicAlreadyCompleted("payment-system")

		assert.Equal(t, MessageInfo, msg.Type)
		assert.Contains(t, msg.Content, "Epic 'payment-system' is already completed")
		assert.Contains(t, msg.Hint, "agentpm status")
	})

	t.Run("EpicStarted", func(t *testing.T) {
		msg := templates.EpicStarted("notification-service")

		assert.Equal(t, MessageSuccess, msg.Type)
		assert.Contains(t, msg.Content, "Epic 'notification-service' started successfully")
	})

	t.Run("EpicCompleted", func(t *testing.T) {
		msg := templates.EpicCompleted("search-feature")

		assert.Equal(t, MessageSuccess, msg.Type)
		assert.Contains(t, msg.Content, "Epic 'search-feature' completed successfully")
	})
}

func TestGenericEntityTemplates(t *testing.T) {
	templates := NewMessageTemplates()

	t.Run("EntityNotFound", func(t *testing.T) {
		msg := templates.EntityNotFound("sprint", "sprint-42")

		assert.Equal(t, MessageError, msg.Type)
		assert.Contains(t, msg.Content, "Sprint 'sprint-42' not found")
		assert.Contains(t, msg.Hint, "agentpm query sprints")
	})

	t.Run("InvalidEntityState", func(t *testing.T) {
		msg := templates.InvalidEntityState("feature", "feature-x", "archived", "activate")

		assert.Equal(t, MessageError, msg.Type)
		assert.Contains(t, msg.Content, "Cannot activate feature 'feature-x' in state 'archived'")
		assert.Contains(t, msg.Hint, "Check valid state transitions for features")
	})

	t.Run("OperationSuccess", func(t *testing.T) {
		msg := templates.OperationSuccess("deploy", "to production")

		assert.Equal(t, MessageSuccess, msg.Type)
		assert.Contains(t, msg.Content, "Deploy to production successfully")
	})

	t.Run("ConfigurationError", func(t *testing.T) {
		msg := templates.ConfigurationError("Missing API key", "Set AGENTPM_API_KEY environment variable")

		assert.Equal(t, MessageError, msg.Type)
		assert.Contains(t, msg.Content, "Configuration error: Missing API key")
		assert.Equal(t, "Set AGENTPM_API_KEY environment variable", msg.Hint)
	})

	t.Run("FileError", func(t *testing.T) {
		msg := templates.FileError("read", "config.json", "Permission denied")

		assert.Equal(t, MessageError, msg.Type)
		assert.Contains(t, msg.Content, "Failed to read file 'config.json': Permission denied")
		assert.Contains(t, msg.Hint, "Check file permissions")
	})

	t.Run("ValidationWarning", func(t *testing.T) {
		msg := templates.ValidationWarning("Duplicate task IDs found", "Use unique identifiers")

		assert.Equal(t, MessageWarning, msg.Type)
		assert.Contains(t, msg.Content, "Validation warning: Duplicate task IDs found")
		assert.Equal(t, "Use unique identifiers", msg.Hint)
	})
}

func TestGlobalMessageTemplates(t *testing.T) {
	t.Run("GlobalMessageTemplates is accessible", func(t *testing.T) {
		assert.NotNil(t, GlobalMessageTemplates)

		// Test that it works
		msg := GlobalMessageTemplates.PhaseStarted("global-test")
		assert.Equal(t, MessageSuccess, msg.Type)
		assert.Contains(t, msg.Content, "Phase 'global-test' started successfully")
	})
}

// Integration tests for template combinations
func TestTemplateIntegration(t *testing.T) {
	templates := NewMessageTemplates()

	t.Run("Workflow simulation", func(t *testing.T) {
		// Simulate a typical workflow with messages
		messages := []*Message{
			templates.PhaseStarted("implementation"),
			templates.TaskStarted("create-user-service"),
			templates.TestStarted("user-service-test"),
			templates.TestPassed("user-service-test"),
			templates.TaskCompleted("create-user-service"),
			templates.PhaseCompleted("implementation"),
		}

		assert.Len(t, messages, 6)

		// Check message types
		assert.Equal(t, MessageSuccess, messages[0].Type) // Phase started
		assert.Equal(t, MessageSuccess, messages[1].Type) // Task started
		assert.Equal(t, MessageSuccess, messages[2].Type) // Test started
		assert.Equal(t, MessageSuccess, messages[3].Type) // Test passed
		assert.Equal(t, MessageSuccess, messages[4].Type) // Task completed
		assert.Equal(t, MessageSuccess, messages[5].Type) // Phase completed
	})

	t.Run("Error scenario simulation", func(t *testing.T) {
		// Simulate error scenarios
		messages := []*Message{
			templates.PhaseConflict("testing", "implementation"),
			templates.EntityNotFound("task", "missing-task"),
			templates.TestFailed("integration-test", "Database connection failed"),
			templates.PhaseIncompleteDependencies("deployment", []string{"task-1"}, []string{"test-1"}),
		}

		assert.Len(t, messages, 4)

		// All should be error messages
		for _, msg := range messages {
			assert.Equal(t, MessageError, msg.Type)
			assert.NotEmpty(t, msg.Content)
			assert.NotEmpty(t, msg.Hint)
		}
	})

	t.Run("Friendly response simulation", func(t *testing.T) {
		// Simulate friendly responses for redundant operations
		messages := []*Message{
			templates.PhaseAlreadyActive("implementation"),
			templates.TaskAlreadyCompleted("setup-db"),
			templates.TestAlreadyPassed("unit-test"),
			templates.EpicAlreadyStarted("user-management"),
		}

		assert.Len(t, messages, 4)

		// All should be info messages with helpful hints
		for _, msg := range messages {
			assert.Equal(t, MessageInfo, msg.Type)
			assert.Contains(t, msg.Content, "already")
			assert.Contains(t, msg.Content, "No action needed")
			assert.NotEmpty(t, msg.Hint)
		}
	})
}
