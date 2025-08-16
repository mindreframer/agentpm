package lifecycle

import (
	"fmt"
	"strings"

	"github.com/memomoo/agentpm/internal/epic"
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
