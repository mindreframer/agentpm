package hints

import (
	"fmt"
	"strings"
	"unicode"
)

// toTitle converts the first character of a string to uppercase
func toTitle(s string) string {
	if len(s) == 0 {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// HintTemplates provides predefined hint templates for common error scenarios
type HintTemplates struct{}

// NewHintTemplates creates a new hint templates instance
func NewHintTemplates() *HintTemplates {
	return &HintTemplates{}
}

// Epic-related hint templates

func (ht *HintTemplates) EpicNotStarted(epicID string) *Hint {
	return &Hint{
		Content:    fmt.Sprintf("Epic '%s' must be started before it can be completed", epicID),
		Category:   HintCategoryActionable,
		Priority:   HintPriorityHigh,
		Command:    "agentpm start-epic",
		Reference:  "Epic lifecycle workflow",
		Conditions: []string{"epic in pending state", "trying to complete"},
	}
}

func (ht *HintTemplates) EpicAlreadyStarted(epicID string) *Hint {
	return &Hint{
		Content:    fmt.Sprintf("Epic '%s' is already in progress", epicID),
		Category:   HintCategoryInformational,
		Priority:   HintPriorityMedium,
		Command:    "agentpm status",
		Reference:  "Epic status checking",
		Conditions: []string{"epic already active", "trying to start"},
	}
}

func (ht *HintTemplates) EpicAlreadyCompleted(epicID string) *Hint {
	return &Hint{
		Content:    fmt.Sprintf("Epic '%s' is already completed", epicID),
		Category:   HintCategoryInformational,
		Priority:   HintPriorityLow,
		Command:    "agentpm status",
		Reference:  "Epic completion status",
		Conditions: []string{"epic already done", "trying to start or complete"},
	}
}

// Phase-related hint templates

func (ht *HintTemplates) PhaseNotActive(phaseID, requiredAction string) *Hint {
	return &Hint{
		Content:    fmt.Sprintf("Phase '%s' must be active to %s", phaseID, requiredAction),
		Category:   HintCategoryActionable,
		Priority:   HintPriorityHigh,
		Command:    fmt.Sprintf("agentpm start-phase %s", phaseID),
		Reference:  "Phase lifecycle workflow",
		Conditions: []string{"phase not active", "trying to operate on tasks"},
	}
}

func (ht *HintTemplates) MultipleActivePhases(currentPhaseID, attemptedPhaseID string) *Hint {
	return &Hint{
		Content:    fmt.Sprintf("Only one phase can be active at a time. Complete phase '%s' before starting '%s'", currentPhaseID, attemptedPhaseID),
		Category:   HintCategoryActionable,
		Priority:   HintPriorityHigh,
		Command:    fmt.Sprintf("agentpm done-phase %s", currentPhaseID),
		Reference:  "Phase constraint management",
		Conditions: []string{"multiple phases attempted", "phase constraint violation"},
	}
}

func (ht *HintTemplates) PhaseHasPendingTasks(phaseID string, taskCount int) *Hint {
	taskWord := "task"
	if taskCount > 1 {
		taskWord = "tasks"
	}

	return &Hint{
		Content:    fmt.Sprintf("Phase '%s' has %d pending %s. Complete or cancel all tasks before completing the phase", phaseID, taskCount, taskWord),
		Category:   HintCategoryActionable,
		Priority:   HintPriorityHigh,
		Command:    "agentpm pending",
		Reference:  "Phase completion requirements",
		Conditions: []string{"incomplete tasks", "trying to complete phase"},
	}
}

// Task-related hint templates

func (ht *HintTemplates) TaskPhaseNotActive(taskID, phaseID string) *Hint {
	return &Hint{
		Content:    fmt.Sprintf("Task '%s' belongs to phase '%s' which is not active", taskID, phaseID),
		Category:   HintCategoryActionable,
		Priority:   HintPriorityHigh,
		Command:    fmt.Sprintf("agentpm start-phase %s", phaseID),
		Reference:  "Task-phase dependencies",
		Conditions: []string{"inactive phase", "trying to start task"},
	}
}

func (ht *HintTemplates) MultipleActiveTasks(currentTaskID, attemptedTaskID, phaseID string) *Hint {
	return &Hint{
		Content:    fmt.Sprintf("Only one task can be active per phase. Complete task '%s' before starting '%s' in phase '%s'", currentTaskID, attemptedTaskID, phaseID),
		Category:   HintCategoryActionable,
		Priority:   HintPriorityHigh,
		Command:    fmt.Sprintf("agentpm done-task %s", currentTaskID),
		Reference:  "Task constraint management",
		Conditions: []string{"multiple tasks in phase", "task constraint violation"},
	}
}

func (ht *HintTemplates) TaskNotActive(taskID string) *Hint {
	return &Hint{
		Content:    fmt.Sprintf("Task '%s' must be active before it can be completed", taskID),
		Category:   HintCategoryActionable,
		Priority:   HintPriorityHigh,
		Command:    fmt.Sprintf("agentpm start-task %s", taskID),
		Reference:  "Task lifecycle workflow",
		Conditions: []string{"task not active", "trying to complete"},
	}
}

// Workflow hint templates

func (ht *HintTemplates) CheckCurrentState() *Hint {
	return &Hint{
		Content:    "Check your current work status to understand what actions are available",
		Category:   HintCategoryWorkflow,
		Priority:   HintPriorityMedium,
		Command:    "agentpm current",
		Reference:  "Workflow status checking",
		Conditions: []string{"general guidance", "workflow help"},
	}
}

func (ht *HintTemplates) ViewPendingWork() *Hint {
	return &Hint{
		Content:    "View pending work to see what needs to be completed",
		Category:   HintCategoryWorkflow,
		Priority:   HintPriorityMedium,
		Command:    "agentpm pending",
		Reference:  "Work planning",
		Conditions: []string{"planning next steps", "work organization"},
	}
}

func (ht *HintTemplates) GetOverallStatus() *Hint {
	return &Hint{
		Content:    "Get an overview of the entire epic status and progress",
		Category:   HintCategoryInformational,
		Priority:   HintPriorityLow,
		Command:    "agentpm status",
		Reference:  "Status overview",
		Conditions: []string{"general information", "progress tracking"},
	}
}

// Configuration and setup hints

func (ht *HintTemplates) ConfigurationIssue(issue string) *Hint {
	return &Hint{
		Content:    fmt.Sprintf("Configuration issue: %s", issue),
		Category:   HintCategoryConfiguration,
		Priority:   HintPriorityHigh,
		Command:    "agentpm config",
		Reference:  "Configuration management",
		Conditions: []string{"config errors", "setup issues"},
	}
}

func (ht *HintTemplates) FileNotFound(filePath string) *Hint {
	return &Hint{
		Content:    fmt.Sprintf("Epic file not found: %s. Check the file path or initialize a new epic", filePath),
		Category:   HintCategoryConfiguration,
		Priority:   HintPriorityHigh,
		Command:    "agentpm init",
		Reference:  "Epic file management",
		Conditions: []string{"missing files", "file system errors"},
	}
}

// Advanced hint templates with context awareness

func (ht *HintTemplates) ContextualWorkflowHint(entityType, operation string) *Hint {
	var content, command string

	switch strings.ToLower(entityType) {
	case "epic":
		switch strings.ToLower(operation) {
		case "start":
			content = "Starting an epic begins the project workflow"
			command = "agentpm start-epic"
		case "complete":
			content = "Completing an epic finalizes all work and generates summary"
			command = "agentpm done-epic"
		default:
			content = "Epic operations manage the overall project lifecycle"
			command = "agentpm status"
		}
	case "phase":
		switch strings.ToLower(operation) {
		case "start":
			content = "Starting a phase enables work on its associated tasks"
			command = "agentpm start-phase <phase-id>"
		case "complete":
			content = "Completing a phase requires all its tasks to be done"
			command = "agentpm done-phase <phase-id>"
		default:
			content = "Phase operations organize work into logical stages"
			command = "agentpm current"
		}
	case "task":
		switch strings.ToLower(operation) {
		case "start":
			content = "Starting a task begins focused work within an active phase"
			command = "agentpm start-task <task-id>"
		case "complete":
			content = "Completing a task marks specific work as done"
			command = "agentpm done-task <task-id>"
		default:
			content = "Task operations track individual work items"
			command = "agentpm pending"
		}
	default:
		content = "Use workflow commands to navigate and manage your project"
		command = "agentpm current"
	}

	return &Hint{
		Content:    content,
		Category:   HintCategoryWorkflow,
		Priority:   HintPriorityMedium,
		Command:    command,
		Reference:  fmt.Sprintf("%s %s operations", toTitle(entityType), operation),
		Conditions: []string{fmt.Sprintf("%s %s context", entityType, operation)},
	}
}
