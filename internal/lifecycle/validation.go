package lifecycle

import (
	"fmt"
	"strings"

	"github.com/mindreframer/agentpm/internal/epic"
)

// ValidationResult contains detailed validation information
type ValidationResult struct {
	IsValid       bool
	PendingPhases []PendingPhase
	FailingTests  []FailingTest
	Summary       ValidationSummary
	Suggestions   []string
}

// ValidationSummary provides counts of validation issues
type ValidationSummary struct {
	TotalPhases       int
	CompletedPhases   int
	PendingPhases     int
	TotalTasks        int
	CompletedTasks    int
	PendingTasks      int
	TotalTests        int
	PassingTests      int
	FailingTests      int
	CompletionPercent int
}

// ValidateEpicCompletion performs comprehensive validation for epic completion
func (ls *LifecycleService) ValidateEpicCompletion(epicFile string) (*ValidationResult, error) {
	// Load the epic
	loadedEpic, err := ls.storage.LoadEpic(epicFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load epic: %w", err)
	}

	result := &ValidationResult{
		IsValid: true,
	}

	// Calculate summary statistics
	result.Summary = ls.calculateValidationSummary(loadedEpic)

	// Check phase completion
	result.PendingPhases = ls.findPendingPhases(loadedEpic)
	if len(result.PendingPhases) > 0 {
		result.IsValid = false
	}

	// Check test status
	result.FailingTests = ls.findFailingTests(loadedEpic)
	if len(result.FailingTests) > 0 {
		result.IsValid = false
	}

	// Generate suggestions based on validation results
	result.Suggestions = ls.generateValidationSuggestions(result)

	return result, nil
}

// calculateValidationSummary calculates summary statistics for validation
func (ls *LifecycleService) calculateValidationSummary(e *epic.Epic) ValidationSummary {
	summary := ValidationSummary{
		TotalPhases: len(e.Phases),
		TotalTasks:  len(e.Tasks),
		TotalTests:  len(e.Tests),
	}

	// Count completed phases
	for _, phase := range e.Phases {
		if phase.Status == epic.StatusCompleted {
			summary.CompletedPhases++
		}
	}
	summary.PendingPhases = summary.TotalPhases - summary.CompletedPhases

	// Count completed tasks
	for _, task := range e.Tasks {
		if task.Status == epic.StatusCompleted {
			summary.CompletedTasks++
		}
	}
	summary.PendingTasks = summary.TotalTasks - summary.CompletedTasks

	// Count passing tests
	for _, test := range e.Tests {
		if test.Status == epic.StatusCompleted {
			summary.PassingTests++
		}
	}
	summary.FailingTests = summary.TotalTests - summary.PassingTests

	// Calculate completion percentage
	totalItems := summary.TotalPhases + summary.TotalTasks + summary.TotalTests
	completedItems := summary.CompletedPhases + summary.CompletedTasks + summary.PassingTests

	if totalItems > 0 {
		summary.CompletionPercent = (completedItems * 100) / totalItems
	}

	return summary
}

// findPendingPhases identifies phases that are not completed
func (ls *LifecycleService) findPendingPhases(e *epic.Epic) []PendingPhase {
	var pending []PendingPhase

	for _, phase := range e.Phases {
		if phase.Status != epic.StatusCompleted {
			pending = append(pending, PendingPhase{
				ID:   phase.ID,
				Name: phase.Name,
			})
		}
	}

	return pending
}

// findFailingTests identifies tests that are not completed (considered failing)
func (ls *LifecycleService) findFailingTests(e *epic.Epic) []FailingTest {
	var failing []FailingTest

	for _, test := range e.Tests {
		if test.Status != epic.StatusCompleted {
			failing = append(failing, FailingTest{
				ID:          test.ID,
				Name:        test.Name,
				Description: test.Description,
			})
		}
	}

	return failing
}

// generateValidationSuggestions creates actionable suggestions based on validation results
func (ls *LifecycleService) generateValidationSuggestions(result *ValidationResult) []string {
	var suggestions []string

	if result.IsValid {
		suggestions = append(suggestions, "Epic is ready for completion")
		return suggestions
	}

	// Suggestions for pending phases
	if len(result.PendingPhases) > 0 {
		if len(result.PendingPhases) == 1 {
			suggestions = append(suggestions,
				fmt.Sprintf("Complete the pending phase: %s", result.PendingPhases[0].Name))
		} else {
			suggestions = append(suggestions,
				fmt.Sprintf("Complete %d pending phases", len(result.PendingPhases)))
		}
		suggestions = append(suggestions, "Use 'agentpm pending' to see all pending work")
	}

	// Suggestions for failing tests
	if len(result.FailingTests) > 0 {
		if len(result.FailingTests) == 1 {
			suggestions = append(suggestions,
				fmt.Sprintf("Fix the failing test: %s", result.FailingTests[0].Name))
		} else {
			suggestions = append(suggestions,
				fmt.Sprintf("Fix %d failing tests", len(result.FailingTests)))
		}
		suggestions = append(suggestions, "Use 'agentpm failing' to see test details")
	}

	// Progress suggestion
	if result.Summary.CompletionPercent > 0 {
		suggestions = append(suggestions,
			fmt.Sprintf("Epic is %d%% complete", result.Summary.CompletionPercent))
	}

	return suggestions
}

// FormatValidationError creates a detailed validation error message
func (ls *LifecycleService) FormatValidationError(result *ValidationResult, epicID string) string {
	var parts []string

	// Main error message
	if len(result.PendingPhases) > 0 && len(result.FailingTests) > 0 {
		parts = append(parts, fmt.Sprintf("Epic %s cannot be completed: %d pending phases, %d failing tests",
			epicID, len(result.PendingPhases), len(result.FailingTests)))
	} else if len(result.PendingPhases) > 0 {
		parts = append(parts, fmt.Sprintf("Epic %s cannot be completed: %d pending phases",
			epicID, len(result.PendingPhases)))
	} else if len(result.FailingTests) > 0 {
		parts = append(parts, fmt.Sprintf("Epic %s cannot be completed: %d failing tests",
			epicID, len(result.FailingTests)))
	}

	// Progress information
	parts = append(parts, fmt.Sprintf("Progress: %d%% complete (%d/%d phases, %d/%d tasks, %d/%d tests)",
		result.Summary.CompletionPercent,
		result.Summary.CompletedPhases, result.Summary.TotalPhases,
		result.Summary.CompletedTasks, result.Summary.TotalTasks,
		result.Summary.PassingTests, result.Summary.TotalTests))

	// Pending phases details (limit to first 3)
	if len(result.PendingPhases) > 0 {
		phaseNames := make([]string, 0, len(result.PendingPhases))
		for i, phase := range result.PendingPhases {
			if i >= 3 {
				phaseNames = append(phaseNames, fmt.Sprintf("... and %d more", len(result.PendingPhases)-3))
				break
			}
			phaseNames = append(phaseNames, fmt.Sprintf("%s (%s)", phase.Name, phase.ID))
		}
		parts = append(parts, fmt.Sprintf("Pending phases: %s", strings.Join(phaseNames, ", ")))
	}

	// Failing tests details (limit to first 3)
	if len(result.FailingTests) > 0 {
		testNames := make([]string, 0, len(result.FailingTests))
		for i, test := range result.FailingTests {
			if i >= 3 {
				testNames = append(testNames, fmt.Sprintf("... and %d more", len(result.FailingTests)-3))
				break
			}
			testNames = append(testNames, fmt.Sprintf("%s (%s)", test.Name, test.ID))
		}
		parts = append(parts, fmt.Sprintf("Failing tests: %s", strings.Join(testNames, ", ")))
	}

	// Suggestions
	if len(result.Suggestions) > 0 {
		parts = append(parts, fmt.Sprintf("Suggestions: %s", strings.Join(result.Suggestions, "; ")))
	}

	return strings.Join(parts, "\n")
}

// Enhanced validateCompletionRequirements using the new validation logic
func (ls *LifecycleService) validateCompletionRequirementsEnhanced(loadedEpic *epic.Epic) error {
	// Use the detailed validation logic
	var pendingPhases []PendingPhase
	var failingTests []FailingTest

	// Find pending phases
	pendingPhases = ls.findPendingPhases(loadedEpic)

	// Find failing tests
	failingTests = ls.findFailingTests(loadedEpic)

	// Return enhanced validation error if issues found
	if len(pendingPhases) > 0 || len(failingTests) > 0 {
		// Create a validation result for error formatting
		result := &ValidationResult{
			IsValid:       false,
			PendingPhases: pendingPhases,
			FailingTests:  failingTests,
			Summary:       ls.calculateValidationSummary(loadedEpic),
		}
		result.Suggestions = ls.generateValidationSuggestions(result)

		return &CompletionValidationError{
			EpicID:        loadedEpic.ID,
			PendingPhases: pendingPhases,
			FailingTests:  failingTests,
			Message:       ls.FormatValidationError(result, loadedEpic.ID),
		}
	}

	return nil
}

// Enhanced validation for Phase 4C - Epic 5 implementation

// ValidationLevel represents the severity of validation issues
type ValidationLevel string

const (
	ValidationLevelValid   ValidationLevel = "valid"
	ValidationLevelWarning ValidationLevel = "warning"
	ValidationLevelError   ValidationLevel = "error"
)

// StateValidationResult represents comprehensive state validation
type StateValidationResult struct {
	Level   ValidationLevel
	Issues  []StateValidationIssue
	Summary string
}

// StateValidationIssue represents a specific state validation problem
type StateValidationIssue struct {
	Type        string
	Level       ValidationLevel
	Message     string
	Context     map[string]string
	Suggestions []string
}

// ValidateEpicState performs comprehensive validation of epic state consistency for Epic 5
func (ls *LifecycleService) ValidateEpicState(epicData *epic.Epic) *StateValidationResult {
	result := &StateValidationResult{
		Level:  ValidationLevelValid,
		Issues: []StateValidationIssue{},
	}

	// Validate phase constraints
	ls.validatePhaseConstraints(epicData, result)

	// Validate task constraints
	ls.validateTaskConstraints(epicData, result)

	// Validate phase-task relationships
	ls.validatePhaseTaskRelationships(epicData, result)

	// Validate epic lifecycle
	ls.validateEpicLifecycle(epicData, result)

	// Generate summary
	ls.generateStateSummary(result)

	return result
}

// validatePhaseConstraints checks phase-specific constraints
func (ls *LifecycleService) validatePhaseConstraints(epicData *epic.Epic, result *StateValidationResult) {
	activePhases := []string{}

	for _, phase := range epicData.Phases {
		if phase.Status == epic.StatusActive {
			activePhases = append(activePhases, phase.ID)
		}
	}

	// Check single active phase constraint
	if len(activePhases) > 1 {
		issue := StateValidationIssue{
			Type:    "multiple_active_phases",
			Level:   ValidationLevelError,
			Message: fmt.Sprintf("Multiple active phases found: %s", strings.Join(activePhases, ", ")),
			Context: map[string]string{
				"active_phases": strings.Join(activePhases, ","),
				"count":         fmt.Sprintf("%d", len(activePhases)),
			},
			Suggestions: []string{
				"Complete or cancel all but one active phase",
				"Use done-phase command to complete finished phases",
			},
		}
		result.Issues = append(result.Issues, issue)
		if result.Level == ValidationLevelValid {
			result.Level = ValidationLevelError
		}
	}

	// Check for phases that should be completed
	for _, phase := range epicData.Phases {
		if phase.Status == epic.StatusActive {
			if ls.isPhaseReadyForCompletion(epicData, phase.ID) {
				issue := StateValidationIssue{
					Type:    "phase_ready_for_completion",
					Level:   ValidationLevelWarning,
					Message: fmt.Sprintf("Phase %s has all tasks completed and should be marked as done", phase.ID),
					Context: map[string]string{
						"phase_id":   phase.ID,
						"phase_name": phase.Name,
					},
					Suggestions: []string{
						fmt.Sprintf("Run: agentpm done-phase %s", phase.ID),
					},
				}
				result.Issues = append(result.Issues, issue)
				if result.Level == ValidationLevelValid {
					result.Level = ValidationLevelWarning
				}
			}
		}
	}
}

// validateTaskConstraints checks task-specific constraints
func (ls *LifecycleService) validateTaskConstraints(epicData *epic.Epic, result *StateValidationResult) {
	activeTasksByPhase := make(map[string][]string)

	for _, task := range epicData.Tasks {
		if task.Status == epic.StatusActive {
			activeTasksByPhase[task.PhaseID] = append(activeTasksByPhase[task.PhaseID], task.ID)
		}
	}

	// Check single active task per phase constraint
	for phaseID, activeTasks := range activeTasksByPhase {
		if len(activeTasks) > 1 {
			issue := StateValidationIssue{
				Type:    "multiple_active_tasks_in_phase",
				Level:   ValidationLevelError,
				Message: fmt.Sprintf("Multiple active tasks in phase %s: %s", phaseID, strings.Join(activeTasks, ", ")),
				Context: map[string]string{
					"phase_id":     phaseID,
					"active_tasks": strings.Join(activeTasks, ","),
					"count":        fmt.Sprintf("%d", len(activeTasks)),
				},
				Suggestions: []string{
					"Complete or cancel all but one active task in the phase",
					"Use done-task or cancel-task commands",
				},
			}
			result.Issues = append(result.Issues, issue)
			if result.Level == ValidationLevelValid {
				result.Level = ValidationLevelError
			}
		}
	}
}

// validatePhaseTaskRelationships checks consistency between phases and tasks
func (ls *LifecycleService) validatePhaseTaskRelationships(epicData *epic.Epic, result *StateValidationResult) {
	// Check for active tasks in inactive phases
	for _, task := range epicData.Tasks {
		if task.Status == epic.StatusActive {
			phase := ls.findPhaseByID(epicData, task.PhaseID)
			if phase == nil {
				issue := StateValidationIssue{
					Type:    "task_orphaned",
					Level:   ValidationLevelError,
					Message: fmt.Sprintf("Task %s references non-existent phase %s", task.ID, task.PhaseID),
					Context: map[string]string{
						"task_id":  task.ID,
						"phase_id": task.PhaseID,
					},
					Suggestions: []string{
						"Update task to reference a valid phase",
						"Create the missing phase",
					},
				}
				result.Issues = append(result.Issues, issue)
				if result.Level == ValidationLevelValid {
					result.Level = ValidationLevelError
				}
			} else if phase.Status != epic.StatusActive {
				issue := StateValidationIssue{
					Type:    "active_task_in_inactive_phase",
					Level:   ValidationLevelError,
					Message: fmt.Sprintf("Active task %s is in inactive phase %s (status: %s)", task.ID, task.PhaseID, phase.Status),
					Context: map[string]string{
						"task_id":      task.ID,
						"phase_id":     task.PhaseID,
						"phase_status": string(phase.Status),
					},
					Suggestions: []string{
						fmt.Sprintf("Start phase %s first", task.PhaseID),
						fmt.Sprintf("Cancel task %s", task.ID),
					},
				}
				result.Issues = append(result.Issues, issue)
				if result.Level == ValidationLevelValid {
					result.Level = ValidationLevelError
				}
			}
		}
	}
}

// validateEpicLifecycle checks overall epic state consistency
func (ls *LifecycleService) validateEpicLifecycle(epicData *epic.Epic, result *StateValidationResult) {
	// Check if epic is marked as completed but has pending work
	if epicData.Status == epic.StatusCompleted {
		hasPendingWork := false

		// Check for incomplete phases
		for _, phase := range epicData.Phases {
			if phase.Status != epic.StatusCompleted {
				hasPendingWork = true
				break
			}
		}

		// Check for incomplete tasks
		if !hasPendingWork {
			for _, task := range epicData.Tasks {
				if task.Status != epic.StatusCompleted && task.Status != epic.StatusCancelled {
					hasPendingWork = true
					break
				}
			}
		}

		// Check for incomplete tests
		if !hasPendingWork {
			for _, test := range epicData.Tests {
				if test.Status != epic.StatusCompleted {
					hasPendingWork = true
					break
				}
			}
		}

		if hasPendingWork {
			issue := StateValidationIssue{
				Type:    "completed_epic_with_pending_work",
				Level:   ValidationLevelError,
				Message: "Epic is marked as completed but has pending phases, tasks, or tests",
				Context: map[string]string{
					"epic_status": string(epicData.Status),
				},
				Suggestions: []string{
					"Complete all pending work before marking epic as done",
					"Change epic status back to active",
				},
			}
			result.Issues = append(result.Issues, issue)
			if result.Level == ValidationLevelValid {
				result.Level = ValidationLevelError
			}
		}
	}
}

// generateStateSummary creates a human-readable summary of validation results
func (ls *LifecycleService) generateStateSummary(result *StateValidationResult) {
	if len(result.Issues) == 0 {
		result.Summary = "Epic state is valid - no issues found"
		return
	}

	errorCount := 0
	warningCount := 0

	for _, issue := range result.Issues {
		switch issue.Level {
		case ValidationLevelError:
			errorCount++
		case ValidationLevelWarning:
			warningCount++
		}
	}

	summaryParts := []string{}
	if errorCount > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d error(s)", errorCount))
	}
	if warningCount > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d warning(s)", warningCount))
	}

	result.Summary = fmt.Sprintf("Epic state validation found %s", strings.Join(summaryParts, " and "))
}

// Helper methods

// isPhaseReadyForCompletion checks if a phase has all tasks completed
func (ls *LifecycleService) isPhaseReadyForCompletion(epicData *epic.Epic, phaseID string) bool {
	hasActiveTasks := false
	hasPendingTasks := false

	for _, task := range epicData.Tasks {
		if task.PhaseID == phaseID {
			switch task.Status {
			case epic.StatusActive:
				hasActiveTasks = true
			case epic.StatusPlanning:
				hasPendingTasks = true
			}
		}
	}

	return !hasActiveTasks && !hasPendingTasks
}

// findPhaseByID finds a phase by its ID
func (ls *LifecycleService) findPhaseByID(epicData *epic.Epic, phaseID string) *epic.Phase {
	for i := range epicData.Phases {
		if epicData.Phases[i].ID == phaseID {
			return &epicData.Phases[i]
		}
	}
	return nil
}
