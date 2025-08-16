package query

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/memomoo/agentpm/internal/epic"
	"github.com/memomoo/agentpm/internal/storage"
)

// QueryService provides read-only query operations for epic data
type QueryService struct {
	storage storage.Storage
	epic    *epic.Epic // cached for single command execution
}

// NewQueryService creates a new QueryService with the given storage implementation
func NewQueryService(storage storage.Storage) *QueryService {
	return &QueryService{
		storage: storage,
	}
}

// LoadEpic loads and caches an epic for query operations
func (qs *QueryService) LoadEpic(epicFile string) error {
	epic, err := qs.storage.LoadEpic(epicFile)
	if err != nil {
		return fmt.Errorf("failed to load epic: %w", err)
	}
	qs.epic = epic
	return nil
}

// EpicStatus represents the overall status and progress of an epic
type EpicStatus struct {
	ID                   string
	Name                 string
	Status               epic.Status
	CompletedPhases      int
	TotalPhases          int
	PassingTests         int
	FailingTests         int
	CompletionPercentage int
	CurrentPhase         string
	CurrentTask          string
}

// GetEpicStatus calculates and returns comprehensive epic status information
func (qs *QueryService) GetEpicStatus() (*EpicStatus, error) {
	if qs.epic == nil {
		return nil, fmt.Errorf("no epic loaded")
	}

	status := &EpicStatus{
		ID:     qs.epic.ID,
		Name:   qs.epic.Name,
		Status: qs.epic.Status,
	}

	// Calculate phase completion
	status.TotalPhases = len(qs.epic.Phases)
	for _, phase := range qs.epic.Phases {
		if qs.getPhaseStatus(phase.ID) == epic.StatusCompleted {
			status.CompletedPhases++
		}
	}

	// Calculate test status (completed = passing, any other status = failing/pending)
	for _, test := range qs.epic.Tests {
		if test.Status == epic.StatusCompleted {
			status.PassingTests++
		} else {
			// For now, treat non-completed tests as "failing" for reporting
			status.FailingTests++
		}
	}

	// Calculate completion percentage
	totalTasks := len(qs.epic.Tasks)
	totalTests := len(qs.epic.Tests)
	if totalTasks+totalTests > 0 {
		completedTasks := 0
		completedTests := 0

		for _, task := range qs.epic.Tasks {
			if task.Status == epic.StatusCompleted {
				completedTasks++
			}
		}

		for _, test := range qs.epic.Tests {
			if test.Status == epic.StatusCompleted {
				completedTests++
			}
		}

		status.CompletionPercentage = (completedTasks + completedTests) * 100 / (totalTasks + totalTests)
	}

	// Find current phase and task
	status.CurrentPhase = qs.findCurrentPhase()
	status.CurrentTask = qs.findCurrentTask()

	return status, nil
}

// CurrentState represents the active work state
type CurrentState struct {
	EpicStatus   epic.Status
	ActivePhase  string
	ActiveTask   string
	NextAction   string
	FailingTests int
}

// GetCurrentState returns information about currently active work
func (qs *QueryService) GetCurrentState() (*CurrentState, error) {
	if qs.epic == nil {
		return nil, fmt.Errorf("no epic loaded")
	}

	state := &CurrentState{
		EpicStatus: qs.epic.Status,
	}

	// Find active phase and task
	state.ActivePhase = qs.findCurrentPhase()
	state.ActiveTask = qs.findCurrentTask()

	// Count failing tests (non-completed tests considered failing for status purposes)
	for _, test := range qs.epic.Tests {
		if test.Status != epic.StatusCompleted {
			state.FailingTests++
		}
	}

	// Determine next action
	state.NextAction = qs.getNextAction()

	return state, nil
}

// PendingWork represents work that hasn't been completed
type PendingWork struct {
	Phases []PendingPhase
	Tasks  []PendingTask
	Tests  []PendingTest
}

type PendingPhase struct {
	ID     string
	Name   string
	Status epic.Status
}

type PendingTask struct {
	ID      string
	PhaseID string
	Name    string
	Status  epic.Status
}

type PendingTest struct {
	ID      string
	TaskID  string
	PhaseID string
	Name    string
	Status  epic.Status
}

// GetPendingWork returns all pending phases, tasks, and tests
func (qs *QueryService) GetPendingWork() (*PendingWork, error) {
	if qs.epic == nil {
		return nil, fmt.Errorf("no epic loaded")
	}

	pending := &PendingWork{}

	// Collect pending phases
	for _, phase := range qs.epic.Phases {
		if phase.Status != epic.StatusCompleted {
			pending.Phases = append(pending.Phases, PendingPhase{
				ID:     phase.ID,
				Name:   phase.Name,
				Status: phase.Status,
			})
		}
	}

	// Collect pending tasks
	for _, task := range qs.epic.Tasks {
		if task.Status != epic.StatusCompleted {
			pending.Tasks = append(pending.Tasks, PendingTask{
				ID:      task.ID,
				PhaseID: task.PhaseID,
				Name:    task.Name,
				Status:  task.Status,
			})
		}
	}

	// Collect pending tests
	for _, test := range qs.epic.Tests {
		if test.Status != epic.StatusCompleted {
			// Find the task's phase for context
			var phaseID string
			for _, task := range qs.epic.Tasks {
				if task.ID == test.TaskID {
					phaseID = task.PhaseID
					break
				}
			}

			pending.Tests = append(pending.Tests, PendingTest{
				ID:      test.ID,
				TaskID:  test.TaskID,
				PhaseID: phaseID,
				Name:    test.Name,
				Status:  test.Status,
			})
		}
	}

	return pending, nil
}

// FailingTest represents a test with failing status
type FailingTest struct {
	ID          string
	PhaseID     string
	TaskID      string
	Name        string
	Description string
	FailureNote string
}

// GetFailingTests returns tests with non-completed status (considered failing for reporting)
func (qs *QueryService) GetFailingTests() ([]FailingTest, error) {
	if qs.epic == nil {
		return nil, fmt.Errorf("no epic loaded")
	}

	var failing []FailingTest

	for _, test := range qs.epic.Tests {
		// For now, treat any non-completed test as "failing" for reporting purposes
		if test.Status != epic.StatusCompleted {
			// Find the task's phase for context
			var phaseID string
			for _, task := range qs.epic.Tasks {
				if task.ID == test.TaskID {
					phaseID = task.PhaseID
					break
				}
			}

			failing = append(failing, FailingTest{
				ID:          test.ID,
				PhaseID:     phaseID,
				TaskID:      test.TaskID,
				Name:        test.Name,
				Description: test.Description,
				FailureNote: "", // Field not available in current epic model
			})
		}
	}

	return failing, nil
}

// Event represents an epic event with metadata
type Event struct {
	Timestamp time.Time
	Agent     string
	PhaseID   string
	Type      string
	Content   string
}

// GetRecentEvents returns events in reverse chronological order with optional limit
func (qs *QueryService) GetRecentEvents(limit int) ([]Event, error) {
	if qs.epic == nil {
		return nil, fmt.Errorf("no epic loaded")
	}

	if limit <= 0 {
		limit = 10 // default limit
	}
	if limit > 100 {
		limit = 100 // max limit
	}

	var events []Event
	for _, event := range qs.epic.Events {
		events = append(events, Event{
			Timestamp: event.Timestamp,
			Agent:     "", // Field not available in current epic model
			PhaseID:   "", // Field not available in current epic model
			Type:      event.Type,
			Content:   event.Data, // Using Data field as Content
		})
	}

	// Sort by timestamp in reverse chronological order (most recent first)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})

	// Apply limit
	if len(events) > limit {
		events = events[:limit]
	}

	return events, nil
}

// Helper methods for internal logic

// getPhaseStatus determines the status of a phase based on its tasks
func (qs *QueryService) getPhaseStatus(phaseID string) epic.Status {
	phaseTasks := qs.getTasksForPhase(phaseID)

	if len(phaseTasks) == 0 {
		return epic.StatusPlanning
	}

	allCompleted := true
	hasActive := false

	for _, task := range phaseTasks {
		switch task.Status {
		case epic.StatusActive:
			hasActive = true
			allCompleted = false
		case epic.StatusPlanning, epic.StatusOnHold:
			allCompleted = false
		}
	}

	if allCompleted {
		return epic.StatusCompleted
	}
	if hasActive {
		return epic.StatusActive
	}
	return epic.StatusPlanning
}

// getTasksForPhase returns all tasks for a given phase
func (qs *QueryService) getTasksForPhase(phaseID string) []epic.Task {
	var tasks []epic.Task
	for _, task := range qs.epic.Tasks {
		if task.PhaseID == phaseID {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// findCurrentPhase finds the currently active phase (active status)
func (qs *QueryService) findCurrentPhase() string {
	for _, phase := range qs.epic.Phases {
		if phase.Status == epic.StatusActive {
			return phase.ID
		}
	}

	// If no phase is explicitly active, check for active tasks
	for _, task := range qs.epic.Tasks {
		if task.Status == epic.StatusActive {
			return task.PhaseID
		}
	}

	return ""
}

// findCurrentTask finds the currently active task (active status)
func (qs *QueryService) findCurrentTask() string {
	for _, task := range qs.epic.Tasks {
		if task.Status == epic.StatusActive {
			return task.ID
		}
	}
	return ""
}

// RelatedItem represents relationships between epic elements
type RelatedItem struct {
	Type         string // "phase", "task", "test"
	ID           string
	Name         string
	Relationship string // "dependency", "blocker", "sequence", "related"
}

// GetRelatedItems finds items related to a given phase, task, or test
func (qs *QueryService) GetRelatedItems(itemType, itemID string) ([]RelatedItem, error) {
	if qs.epic == nil {
		return nil, fmt.Errorf("no epic loaded")
	}

	var related []RelatedItem

	switch itemType {
	case "phase":
		// Find tasks in this phase
		for _, task := range qs.epic.Tasks {
			if task.PhaseID == itemID {
				related = append(related, RelatedItem{
					Type:         "task",
					ID:           task.ID,
					Name:         task.Name,
					Relationship: "contains",
				})
			}
		}

		// Find tests for tasks in this phase
		for _, test := range qs.epic.Tests {
			for _, task := range qs.epic.Tasks {
				if task.ID == test.TaskID && task.PhaseID == itemID {
					related = append(related, RelatedItem{
						Type:         "test",
						ID:           test.ID,
						Name:         test.Name,
						Relationship: "validates",
					})
				}
			}
		}

	case "task":
		// Find parent phase
		for _, task := range qs.epic.Tasks {
			if task.ID == itemID {
				for _, phase := range qs.epic.Phases {
					if phase.ID == task.PhaseID {
						related = append(related, RelatedItem{
							Type:         "phase",
							ID:           phase.ID,
							Name:         phase.Name,
							Relationship: "parent",
						})
					}
				}
				break
			}
		}

		// Find tests for this task
		for _, test := range qs.epic.Tests {
			if test.TaskID == itemID {
				related = append(related, RelatedItem{
					Type:         "test",
					ID:           test.ID,
					Name:         test.Name,
					Relationship: "validates",
				})
			}
		}

	case "test":
		// Find parent task and phase
		for _, test := range qs.epic.Tests {
			if test.ID == itemID {
				for _, task := range qs.epic.Tasks {
					if task.ID == test.TaskID {
						related = append(related, RelatedItem{
							Type:         "task",
							ID:           task.ID,
							Name:         task.Name,
							Relationship: "parent",
						})

						for _, phase := range qs.epic.Phases {
							if phase.ID == task.PhaseID {
								related = append(related, RelatedItem{
									Type:         "phase",
									ID:           phase.ID,
									Name:         phase.Name,
									Relationship: "ancestor",
								})
							}
						}
						break
					}
				}
				break
			}
		}
	}

	return related, nil
}

// ImpactAnalysis represents the potential impact of changes
type ImpactAnalysis struct {
	AffectedPhases []string
	AffectedTasks  []string
	AffectedTests  []string
	RiskLevel      string // "low", "medium", "high"
	Description    string
}

// AnalyzeImpact analyzes the potential impact of completing or modifying an item
func (qs *QueryService) AnalyzeImpact(itemType, itemID string) (*ImpactAnalysis, error) {
	if qs.epic == nil {
		return nil, fmt.Errorf("no epic loaded")
	}

	analysis := &ImpactAnalysis{
		RiskLevel: "low",
	}

	switch itemType {
	case "phase":
		// Completing a phase affects all its tasks and tests
		for _, task := range qs.epic.Tasks {
			if task.PhaseID == itemID {
				analysis.AffectedTasks = append(analysis.AffectedTasks, task.ID)

				for _, test := range qs.epic.Tests {
					if test.TaskID == task.ID {
						analysis.AffectedTests = append(analysis.AffectedTests, test.ID)
					}
				}
			}
		}

		if len(analysis.AffectedTasks) > 5 {
			analysis.RiskLevel = "medium"
		}
		if len(analysis.AffectedTasks) > 10 {
			analysis.RiskLevel = "high"
		}

		analysis.Description = fmt.Sprintf("Completing phase affects %d tasks and %d tests",
			len(analysis.AffectedTasks), len(analysis.AffectedTests))

	case "task":
		// Completing a task affects its tests and potentially dependent tasks
		for _, test := range qs.epic.Tests {
			if test.TaskID == itemID {
				analysis.AffectedTests = append(analysis.AffectedTests, test.ID)
			}
		}

		// Find parent phase
		for _, task := range qs.epic.Tasks {
			if task.ID == itemID {
				analysis.AffectedPhases = append(analysis.AffectedPhases, task.PhaseID)
				break
			}
		}

		if len(analysis.AffectedTests) > 3 {
			analysis.RiskLevel = "medium"
		}

		analysis.Description = fmt.Sprintf("Completing task affects %d tests", len(analysis.AffectedTests))

	case "test":
		// Completing a test primarily affects its parent task
		for _, test := range qs.epic.Tests {
			if test.ID == itemID {
				analysis.AffectedTasks = append(analysis.AffectedTasks, test.TaskID)

				// Find parent phase
				for _, task := range qs.epic.Tasks {
					if task.ID == test.TaskID {
						analysis.AffectedPhases = append(analysis.AffectedPhases, task.PhaseID)
						break
					}
				}
				break
			}
		}

		analysis.Description = "Completing test affects its parent task"
	}

	return analysis, nil
}

// ProgressInsight provides analysis of epic progress patterns
type ProgressInsight struct {
	Velocity            float64 // tasks/tests completed per unit time
	EstimatedCompletion string
	Bottlenecks         []string
	Recommendations     []string
}

// GetProgressInsights analyzes epic progress and provides recommendations
func (qs *QueryService) GetProgressInsights() (*ProgressInsight, error) {
	if qs.epic == nil {
		return nil, fmt.Errorf("no epic loaded")
	}

	insight := &ProgressInsight{}

	// Calculate completion velocity (simplified)
	totalItems := len(qs.epic.Tasks) + len(qs.epic.Tests)
	completedItems := 0
	for _, task := range qs.epic.Tasks {
		if task.Status == epic.StatusCompleted {
			completedItems++
		}
	}
	for _, test := range qs.epic.Tests {
		if test.Status == epic.StatusCompleted {
			completedItems++
		}
	}

	if completedItems > 0 && totalItems > 0 {
		completionRate := float64(completedItems) / float64(totalItems)
		insight.Velocity = completionRate

		if completionRate > 0.8 {
			insight.EstimatedCompletion = "Near completion"
		} else if completionRate > 0.5 {
			insight.EstimatedCompletion = "Mid-progress"
		} else {
			insight.EstimatedCompletion = "Early stage"
		}
	} else {
		insight.EstimatedCompletion = "Just started"
	}

	// Identify bottlenecks
	failingTestCount := 0
	for _, test := range qs.epic.Tests {
		if test.Status != epic.StatusCompleted {
			failingTestCount++
		}
	}

	if failingTestCount > 3 {
		insight.Bottlenecks = append(insight.Bottlenecks, "Multiple failing tests")
	}

	activeTaskCount := 0
	for _, task := range qs.epic.Tasks {
		if task.Status == epic.StatusActive {
			activeTaskCount++
		}
	}

	if activeTaskCount > 2 {
		insight.Bottlenecks = append(insight.Bottlenecks, "Too many active tasks")
	}

	// Generate recommendations
	if len(insight.Bottlenecks) > 0 {
		insight.Recommendations = append(insight.Recommendations, "Focus on resolving bottlenecks first")
	}

	if failingTestCount > 0 {
		insight.Recommendations = append(insight.Recommendations, "Prioritize fixing failing tests")
	}

	if activeTaskCount == 0 {
		insight.Recommendations = append(insight.Recommendations, "Start next planned task")
	}

	return insight, nil
}

// getNextAction provides next action recommendation based on epic state
func (qs *QueryService) getNextAction() string {
	// 1. If failing tests exist → "Fix failing tests" (non-completed tests)
	failingTests := 0
	var failingTestDescriptions []string
	for _, test := range qs.epic.Tests {
		if test.Status != epic.StatusCompleted {
			failingTests++
			if len(failingTestDescriptions) < 3 { // limit to first 3
				failingTestDescriptions = append(failingTestDescriptions, test.Name)
			}
		}
	}
	if failingTests > 0 {
		if len(failingTestDescriptions) > 0 {
			return fmt.Sprintf("Fix failing tests: %s", strings.Join(failingTestDescriptions, ", "))
		}
		return "Fix failing tests"
	}

	// 2. If active task exists → "Continue work on task"
	currentTask := qs.findCurrentTask()
	if currentTask != "" {
		for _, task := range qs.epic.Tasks {
			if task.ID == currentTask {
				return fmt.Sprintf("Continue work on: %s", task.Name)
			}
		}
	}

	// 3. If pending tasks in active phase → "Start next task"
	currentPhase := qs.findCurrentPhase()
	if currentPhase != "" {
		for _, task := range qs.epic.Tasks {
			if task.PhaseID == currentPhase && task.Status == epic.StatusPlanning {
				return fmt.Sprintf("Start next task: %s", task.Name)
			}
		}
	}

	// 4. If pending phases → "Start next phase"
	for _, phase := range qs.epic.Phases {
		if phase.Status == epic.StatusPlanning {
			return fmt.Sprintf("Start next phase: %s", phase.Name)
		}
	}

	// 5. If all complete → "Epic ready for completion"
	return "Epic ready for completion"
}
