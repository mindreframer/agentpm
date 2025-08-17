package epic

import (
	"fmt"
	"strings"
)

// StatusValidationError represents a status validation error with detailed information
type StatusValidationError struct {
	EntityType    string         `json:"entity_type"` // "epic", "phase", "task", "test"
	EntityID      string         `json:"entity_id"`
	EntityName    string         `json:"entity_name,omitempty"`
	CurrentStatus string         `json:"current_status"`
	TargetStatus  string         `json:"target_status"`
	BlockingItems []BlockingItem `json:"blocking_items"`
	Message       string         `json:"message"`
	Suggestions   []string       `json:"suggestions,omitempty"`
}

// BlockingItem represents an item that is blocking a status transition
type BlockingItem struct {
	Type   string `json:"type"` // "task", "test", "phase"
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Result string `json:"result,omitempty"` // For tests
}

// Error implements the error interface
func (e *StatusValidationError) Error() string {
	return e.Message
}

// StatusValidationResult represents the result of a status validation operation
type StatusValidationResult struct {
	Valid         bool                   `json:"valid"`
	Error         *StatusValidationError `json:"error,omitempty"`
	Warnings      []string               `json:"warnings,omitempty"`
	BlockingCount int                    `json:"blocking_count"`
}

// StatusValidator provides validation functionality for status transitions
type StatusValidator struct {
	epic *Epic
}

// NewStatusValidator creates a new status validator for an epic
func NewStatusValidator(epic *Epic) *StatusValidator {
	return &StatusValidator{epic: epic}
}

// ValidateEpicStatusTransition validates an epic status transition
func (v *StatusValidator) ValidateEpicStatusTransition(targetStatus EpicStatus) *StatusValidationResult {
	currentStatus, err := ValidateEpicStatus(string(v.epic.Status))
	if err != nil {
		return &StatusValidationResult{
			Valid: false,
			Error: &StatusValidationError{
				EntityType:    "epic",
				EntityID:      v.epic.ID,
				EntityName:    v.epic.Name,
				CurrentStatus: string(v.epic.Status),
				TargetStatus:  string(targetStatus),
				Message:       fmt.Sprintf("Invalid current epic status: %s", v.epic.Status),
			},
		}
	}

	// Check if transition is allowed
	if !currentStatus.CanTransitionTo(targetStatus) {
		return &StatusValidationResult{
			Valid: false,
			Error: &StatusValidationError{
				EntityType:    "epic",
				EntityID:      v.epic.ID,
				EntityName:    v.epic.Name,
				CurrentStatus: string(currentStatus),
				TargetStatus:  string(targetStatus),
				Message:       fmt.Sprintf("Epic cannot transition from %s to %s", currentStatus, targetStatus),
				Suggestions:   []string{fmt.Sprintf("Epic status transitions must follow: pending → wip → done")},
			},
		}
	}

	// For completion validation, check if all phases are complete
	if targetStatus == EpicStatusDone {
		return v.validateEpicCompletion()
	}

	return &StatusValidationResult{Valid: true}
}

// ValidatePhaseStatusTransition validates a phase status transition
func (v *StatusValidator) ValidatePhaseStatusTransition(phaseID string, targetStatus PhaseStatus) *StatusValidationResult {
	phase := v.findPhase(phaseID)
	if phase == nil {
		return &StatusValidationResult{
			Valid: false,
			Error: &StatusValidationError{
				EntityType: "phase",
				EntityID:   phaseID,
				Message:    fmt.Sprintf("Phase %s not found", phaseID),
			},
		}
	}

	currentStatus, err := ValidatePhaseStatus(string(phase.Status))
	if err != nil {
		return &StatusValidationResult{
			Valid: false,
			Error: &StatusValidationError{
				EntityType:    "phase",
				EntityID:      phaseID,
				EntityName:    phase.Name,
				CurrentStatus: string(phase.Status),
				TargetStatus:  string(targetStatus),
				Message:       fmt.Sprintf("Invalid current phase status: %s", phase.Status),
			},
		}
	}

	// Check if transition is allowed
	if !currentStatus.CanTransitionTo(targetStatus) {
		return &StatusValidationResult{
			Valid: false,
			Error: &StatusValidationError{
				EntityType:    "phase",
				EntityID:      phaseID,
				EntityName:    phase.Name,
				CurrentStatus: string(currentStatus),
				TargetStatus:  string(targetStatus),
				Message:       fmt.Sprintf("Phase cannot transition from %s to %s", currentStatus, targetStatus),
				Suggestions:   []string{fmt.Sprintf("Phase status transitions must follow: pending → wip → done")},
			},
		}
	}

	// For completion validation, check business rules
	if targetStatus == PhaseStatusDone {
		return v.validatePhaseCompletion(phaseID)
	}

	return &StatusValidationResult{Valid: true}
}

// ValidateTaskStatusTransition validates a task status transition
func (v *StatusValidator) ValidateTaskStatusTransition(taskID string, targetStatus TaskStatus) *StatusValidationResult {
	task := v.findTask(taskID)
	if task == nil {
		return &StatusValidationResult{
			Valid: false,
			Error: &StatusValidationError{
				EntityType: "task",
				EntityID:   taskID,
				Message:    fmt.Sprintf("Task %s not found", taskID),
			},
		}
	}

	currentStatus, err := ValidateTaskStatus(string(task.Status))
	if err != nil {
		return &StatusValidationResult{
			Valid: false,
			Error: &StatusValidationError{
				EntityType:    "task",
				EntityID:      taskID,
				EntityName:    task.Name,
				CurrentStatus: string(task.Status),
				TargetStatus:  string(targetStatus),
				Message:       fmt.Sprintf("Invalid current task status: %s", task.Status),
			},
		}
	}

	// Check if transition is allowed
	if !currentStatus.CanTransitionTo(targetStatus) {
		return &StatusValidationResult{
			Valid: false,
			Error: &StatusValidationError{
				EntityType:    "task",
				EntityID:      taskID,
				EntityName:    task.Name,
				CurrentStatus: string(currentStatus),
				TargetStatus:  string(targetStatus),
				Message:       fmt.Sprintf("Task cannot transition from %s to %s", currentStatus, targetStatus),
				Suggestions:   []string{fmt.Sprintf("Task status transitions: pending → wip → done, or pending/wip → cancelled")},
			},
		}
	}

	// For completion validation, check business rules
	if targetStatus == TaskStatusDone {
		return v.validateTaskCompletion(taskID)
	}

	return &StatusValidationResult{Valid: true}
}

// ValidateTestStatusTransition validates a test status transition
func (v *StatusValidator) ValidateTestStatusTransition(testID string, targetStatus TestStatus, targetResult TestResult) *StatusValidationResult {
	test := v.findTest(testID)
	if test == nil {
		return &StatusValidationResult{
			Valid: false,
			Error: &StatusValidationError{
				EntityType: "test",
				EntityID:   testID,
				Message:    fmt.Sprintf("Test %s not found", testID),
			},
		}
	}

	currentStatus := test.TestStatus

	// Check if transition is allowed
	if !currentStatus.CanTransitionTo(targetStatus) {
		return &StatusValidationResult{
			Valid: false,
			Error: &StatusValidationError{
				EntityType:    "test",
				EntityID:      testID,
				EntityName:    test.Name,
				CurrentStatus: string(currentStatus),
				TargetStatus:  string(targetStatus),
				Message:       fmt.Sprintf("Test cannot transition from %s to %s", currentStatus, targetStatus),
				Suggestions:   []string{fmt.Sprintf("Test status transitions: pending → wip → done, done → wip (for failing), or pending/wip → cancelled")},
			},
		}
	}

	// Business rule: failing tests cannot be marked as done
	if targetStatus == TestStatusDone && targetResult == TestResultFailing {
		return &StatusValidationResult{
			Valid: false,
			Error: &StatusValidationError{
				EntityType:    "test",
				EntityID:      testID,
				EntityName:    test.Name,
				CurrentStatus: string(currentStatus),
				TargetStatus:  string(targetStatus),
				Message:       "Failing tests cannot be marked as done - they can only be cancelled with a reason",
				Suggestions:   []string{"Cancel the test with a reason using: agentpm cancel test <test-id> --reason \"<reason>\""},
			},
		}
	}

	// Check if test belongs to current active phase
	if !v.isTestInActivePhase(testID) {
		activePhase := v.getActivePhaseID()
		testPhase := v.getTestPhaseID(testID)
		return &StatusValidationResult{
			Valid: false,
			Error: &StatusValidationError{
				EntityType:    "test",
				EntityID:      testID,
				EntityName:    test.Name,
				CurrentStatus: string(currentStatus),
				TargetStatus:  string(targetStatus),
				Message:       fmt.Sprintf("Test belongs to phase '%s' but current active phase is '%s'", testPhase, activePhase),
				Suggestions:   []string{fmt.Sprintf("Switch to phase '%s' first or select a test from the current active phase", testPhase)},
			},
		}
	}

	return &StatusValidationResult{Valid: true}
}

// validateEpicCompletion validates that an epic can be completed
func (v *StatusValidator) validateEpicCompletion() *StatusValidationResult {
	var blockingItems []BlockingItem

	// Check all phases are done
	for _, phase := range v.epic.Phases {
		phaseStatus, err := ValidatePhaseStatus(string(phase.Status))
		if err != nil || phaseStatus != PhaseStatusDone {
			blockingItems = append(blockingItems, BlockingItem{
				Type:   "phase",
				ID:     phase.ID,
				Name:   phase.Name,
				Status: string(phase.Status),
			})
		}
	}

	if len(blockingItems) > 0 {
		return &StatusValidationResult{
			Valid:         false,
			BlockingCount: len(blockingItems),
			Error: &StatusValidationError{
				EntityType:    "epic",
				EntityID:      v.epic.ID,
				EntityName:    v.epic.Name,
				CurrentStatus: string(v.epic.Status),
				TargetStatus:  string(EpicStatusDone),
				BlockingItems: blockingItems,
				Message:       fmt.Sprintf("Epic cannot be completed: %d incomplete phases", len(blockingItems)),
				Suggestions:   []string{"Complete all phases before marking epic as done"},
			},
		}
	}

	return &StatusValidationResult{Valid: true}
}

// validatePhaseCompletion validates that a phase can be completed
func (v *StatusValidator) validatePhaseCompletion(phaseID string) *StatusValidationResult {
	var blockingItems []BlockingItem

	// Check all tasks in this phase are done
	for _, task := range v.epic.Tasks {
		if task.PhaseID == phaseID {
			taskStatus, err := ValidateTaskStatus(string(task.Status))
			if err != nil || taskStatus != TaskStatusDone {
				blockingItems = append(blockingItems, BlockingItem{
					Type:   "task",
					ID:     task.ID,
					Name:   task.Name,
					Status: string(task.Status),
				})
			}
		}
	}

	// Check all tests for tasks in this phase are done
	for _, test := range v.epic.Tests {
		if test.PhaseID == phaseID {
			if test.TestStatus != TestStatusDone {
				blockingItems = append(blockingItems, BlockingItem{
					Type:   "test",
					ID:     test.ID,
					Name:   test.Name,
					Status: string(test.TestStatus),
					Result: string(test.TestResult),
				})
			}
		}
	}

	if len(blockingItems) > 0 {
		phase := v.findPhase(phaseID)
		phaseName := ""
		if phase != nil {
			phaseName = phase.Name
		}

		// Count types for better error message
		taskCount := 0
		testCount := 0
		for _, item := range blockingItems {
			if item.Type == "task" {
				taskCount++
			} else if item.Type == "test" {
				testCount++
			}
		}

		var messageParts []string
		if taskCount > 0 {
			messageParts = append(messageParts, fmt.Sprintf("%d pending/wip tasks", taskCount))
		}
		if testCount > 0 {
			messageParts = append(messageParts, fmt.Sprintf("%d pending/wip tests", testCount))
		}

		return &StatusValidationResult{
			Valid:         false,
			BlockingCount: len(blockingItems),
			Error: &StatusValidationError{
				EntityType:    "phase",
				EntityID:      phaseID,
				EntityName:    phaseName,
				CurrentStatus: string(PhaseStatusWIP),
				TargetStatus:  string(PhaseStatusDone),
				BlockingItems: blockingItems,
				Message:       fmt.Sprintf("Phase cannot be completed: %s", strings.Join(messageParts, ", ")),
				Suggestions:   []string{"Complete all tasks and tests in this phase before marking it as done"},
			},
		}
	}

	return &StatusValidationResult{Valid: true}
}

// validateTaskCompletion validates that a task can be completed
func (v *StatusValidator) validateTaskCompletion(taskID string) *StatusValidationResult {
	var blockingItems []BlockingItem

	// Check all tests for this task are done
	for _, test := range v.epic.Tests {
		if test.TaskID == taskID {
			if test.TestStatus != TestStatusDone {
				blockingItems = append(blockingItems, BlockingItem{
					Type:   "test",
					ID:     test.ID,
					Name:   test.Name,
					Status: string(test.TestStatus),
					Result: string(test.TestResult),
				})
			}
		}
	}

	if len(blockingItems) > 0 {
		task := v.findTask(taskID)
		taskName := ""
		if task != nil {
			taskName = task.Name
		}

		return &StatusValidationResult{
			Valid:         false,
			BlockingCount: len(blockingItems),
			Error: &StatusValidationError{
				EntityType:    "task",
				EntityID:      taskID,
				EntityName:    taskName,
				CurrentStatus: string(TaskStatusWIP),
				TargetStatus:  string(TaskStatusDone),
				BlockingItems: blockingItems,
				Message:       fmt.Sprintf("Task cannot be completed: %d pending/wip tests", len(blockingItems)),
				Suggestions:   []string{"Complete all tests for this task before marking it as done"},
			},
		}
	}

	return &StatusValidationResult{Valid: true}
}

// Helper methods for finding entities

func (v *StatusValidator) findPhase(phaseID string) *Phase {
	for i := range v.epic.Phases {
		if v.epic.Phases[i].ID == phaseID {
			return &v.epic.Phases[i]
		}
	}
	return nil
}

func (v *StatusValidator) findTask(taskID string) *Task {
	for i := range v.epic.Tasks {
		if v.epic.Tasks[i].ID == taskID {
			return &v.epic.Tasks[i]
		}
	}
	return nil
}

func (v *StatusValidator) findTest(testID string) *Test {
	for i := range v.epic.Tests {
		if v.epic.Tests[i].ID == testID {
			return &v.epic.Tests[i]
		}
	}
	return nil
}

func (v *StatusValidator) isTestInActivePhase(testID string) bool {
	test := v.findTest(testID)
	if test == nil {
		return false
	}

	activePhaseID := v.getActivePhaseID()
	return test.PhaseID == activePhaseID
}

func (v *StatusValidator) getActivePhaseID() string {
	if v.epic.CurrentState != nil {
		return v.epic.CurrentState.ActivePhase
	}

	// Fallback: find first non-completed phase
	for _, phase := range v.epic.Phases {
		phaseStatus, err := ValidatePhaseStatus(string(phase.Status))
		if err == nil && phaseStatus != PhaseStatusDone {
			return phase.ID
		}
	}

	return ""
}

func (v *StatusValidator) getTestPhaseID(testID string) string {
	test := v.findTest(testID)
	if test != nil {
		return test.PhaseID
	}
	return ""
}

// FormatErrorXML formats a validation error as XML for CLI output
func (e *StatusValidationError) FormatErrorXML() string {
	var xml strings.Builder

	xml.WriteString("<error>\n")
	xml.WriteString(fmt.Sprintf("    <type>%s_completion_blocked</type>\n", e.EntityType))
	xml.WriteString(fmt.Sprintf("    <message>%s</message>\n", e.Message))

	if len(e.BlockingItems) > 0 {
		xml.WriteString("    <blocking_items>\n")

		// Group by type for cleaner output
		taskItems := []BlockingItem{}
		testItems := []BlockingItem{}
		phaseItems := []BlockingItem{}

		for _, item := range e.BlockingItems {
			switch item.Type {
			case "task":
				taskItems = append(taskItems, item)
			case "test":
				testItems = append(testItems, item)
			case "phase":
				phaseItems = append(phaseItems, item)
			}
		}

		if len(taskItems) > 0 {
			xml.WriteString(fmt.Sprintf("        <tasks count=\"%d\">\n", len(taskItems)))
			for _, item := range taskItems {
				xml.WriteString(fmt.Sprintf("            <task id=\"%s\" name=\"%s\" status=\"%s\"/>\n",
					item.ID, item.Name, item.Status))
			}
			xml.WriteString("        </tasks>\n")
		}

		if len(testItems) > 0 {
			xml.WriteString(fmt.Sprintf("        <tests count=\"%d\">\n", len(testItems)))
			for _, item := range testItems {
				if item.Result != "" {
					xml.WriteString(fmt.Sprintf("            <test id=\"%s\" name=\"%s\" status=\"%s\" result=\"%s\"/>\n",
						item.ID, item.Name, item.Status, item.Result))
				} else {
					xml.WriteString(fmt.Sprintf("            <test id=\"%s\" name=\"%s\" status=\"%s\"/>\n",
						item.ID, item.Name, item.Status))
				}
			}
			xml.WriteString("        </tests>\n")
		}

		if len(phaseItems) > 0 {
			xml.WriteString(fmt.Sprintf("        <phases count=\"%d\">\n", len(phaseItems)))
			for _, item := range phaseItems {
				xml.WriteString(fmt.Sprintf("            <phase id=\"%s\" name=\"%s\" status=\"%s\"/>\n",
					item.ID, item.Name, item.Status))
			}
			xml.WriteString("        </phases>\n")
		}

		xml.WriteString("    </blocking_items>\n")
	}

	xml.WriteString("</error>")
	return xml.String()
}
