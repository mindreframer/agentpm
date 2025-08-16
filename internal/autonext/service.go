package autonext

import (
	"fmt"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/phases"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/mindreframer/agentpm/internal/tasks"
)

// AutoNextAction represents the type of action to take
type AutoNextAction string

const (
	ActionStartTask     AutoNextAction = "start_task"
	ActionStartPhase    AutoNextAction = "start_phase"
	ActionCompletePhase AutoNextAction = "complete_phase"
	ActionCompleteEpic  AutoNextAction = "complete_epic"
	ActionNoWork        AutoNextAction = "no_work"
)

// AutoNextResult represents the result of auto-next selection
type AutoNextResult struct {
	Action       AutoNextAction
	PhaseID      string
	TaskID       string
	Message      string
	XMLOutput    string
	PhaseName    string
	TaskName     string
	PhaseStatus  epic.Status
	TaskStatus   epic.Status
	StartedAt    time.Time
	AutoSelected bool
}

// AutoNextService provides intelligent next work selection
type AutoNextService struct {
	storage      storage.Storage
	query        *query.QueryService
	phaseService *phases.PhaseService
	taskService  *tasks.TaskService
}

func NewAutoNextService(storage storage.Storage, query *query.QueryService, phaseService *phases.PhaseService, taskService *tasks.TaskService) *AutoNextService {
	return &AutoNextService{
		storage:      storage,
		query:        query,
		phaseService: phaseService,
		taskService:  taskService,
	}
}

// SelectNext implements the auto-next selection algorithm according to Epic 5 spec
func (s *AutoNextService) SelectNext(epicData *epic.Epic, timestamp time.Time) (*AutoNextResult, error) {
	// Algorithm priority (from Epic 5 spec):
	// 1. If Active Phase Exists: Find next pending task in current active phase
	// 2. If No Active Phase: Find next pending phase and activate it, then start first pending task
	// 3. If Current Phase Complete: Complete current phase, activate next phase, start first task
	// 4. If All Work Complete: Return completion message

	activePhase := s.phaseService.GetActivePhase(epicData)

	if activePhase != nil {
		return s.handleActivePhase(epicData, activePhase, timestamp)
	} else {
		return s.handleNoActivePhase(epicData, timestamp)
	}
}

// handleActivePhase handles selection when there's an active phase
func (s *AutoNextService) handleActivePhase(epicData *epic.Epic, activePhase *epic.Phase, timestamp time.Time) (*AutoNextResult, error) {
	// Check if there's already an active task in the phase
	activeTask := s.taskService.GetActiveTask(epicData, activePhase.ID)
	if activeTask != nil {
		// There's already an active task, no action needed
		return &AutoNextResult{
			Action:  ActionNoWork,
			Message: fmt.Sprintf("Task %s is already active in phase %s", activeTask.ID, activePhase.ID),
		}, nil
	}

	// Look for next pending task in current active phase
	pendingTasks := s.taskService.GetPendingTasksInPhase(epicData, activePhase.ID)
	pendingTasksFiltered := filterPendingOnly(pendingTasks)

	if len(pendingTasksFiltered) > 0 {
		// Start the first pending task in the active phase
		task := pendingTasksFiltered[0]

		err := s.taskService.StartTask(epicData, task.ID, timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to start task %s: %w", task.ID, err)
		}

		return &AutoNextResult{
			Action:       ActionStartTask,
			PhaseID:      activePhase.ID,
			TaskID:       task.ID,
			PhaseName:    activePhase.Name,
			TaskName:     task.Name,
			PhaseStatus:  activePhase.Status,
			TaskStatus:   epic.StatusActive,
			StartedAt:    timestamp,
			AutoSelected: true,
			Message:      fmt.Sprintf("Started Task %s: %s (auto-selected)", task.ID, task.Name),
		}, nil
	}

	// No pending tasks in current phase - check if phase can be completed
	allTasksCompleted := s.areAllTasksCompletedOrCancelled(epicData, activePhase.ID)
	if allTasksCompleted {
		// Complete the current phase
		err := s.phaseService.CompletePhase(epicData, activePhase.ID, timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to complete phase %s: %w", activePhase.ID, err)
		}

		// Now look for next phase to start
		return s.handleNoActivePhase(epicData, timestamp)
	}

	// Phase has pending work but no tasks can be started (shouldn't happen with proper validation)
	return &AutoNextResult{
		Action:  ActionNoWork,
		Message: fmt.Sprintf("Phase %s has pending work but no tasks can be started", activePhase.ID),
	}, nil
}

// handleNoActivePhase handles selection when there's no active phase
func (s *AutoNextService) handleNoActivePhase(epicData *epic.Epic, timestamp time.Time) (*AutoNextResult, error) {
	// Find next pending phase
	nextPhase := s.findNextPendingPhase(epicData)
	if nextPhase == nil {
		// All phases are completed - epic is ready for completion
		return &AutoNextResult{
			Action:  ActionCompleteEpic,
			Message: "All phases and tasks completed. Epic ready for completion.",
		}, nil
	}

	// Start the next phase
	err := s.phaseService.StartPhase(epicData, nextPhase.ID, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to start phase %s: %w", nextPhase.ID, err)
	}

	// Find first pending task in the newly started phase
	pendingTasks := s.taskService.GetPendingTasksInPhase(epicData, nextPhase.ID)
	pendingTasksFiltered := filterPendingOnly(pendingTasks)

	if len(pendingTasksFiltered) > 0 {
		// Start the first pending task
		task := pendingTasksFiltered[0]

		err := s.taskService.StartTask(epicData, task.ID, timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to start task %s: %w", task.ID, err)
		}

		// Get all tasks for XML output
		allTasksInPhase := s.taskService.GetTasksInPhase(epicData, nextPhase.ID)

		return &AutoNextResult{
			Action:       ActionStartPhase,
			PhaseID:      nextPhase.ID,
			TaskID:       task.ID,
			PhaseName:    nextPhase.Name,
			TaskName:     task.Name,
			PhaseStatus:  epic.StatusActive,
			TaskStatus:   epic.StatusActive,
			StartedAt:    timestamp,
			AutoSelected: true,
			Message:      fmt.Sprintf("Started Phase %s and Task %s (auto-selected)", nextPhase.ID, task.ID),
			XMLOutput:    s.formatPhaseStartedXML(epicData.ID, nextPhase, allTasksInPhase, task.ID, timestamp),
		}, nil
	}

	// Phase started but no tasks available (empty phase)
	return &AutoNextResult{
		Action:      ActionStartPhase,
		PhaseID:     nextPhase.ID,
		PhaseName:   nextPhase.Name,
		PhaseStatus: epic.StatusActive,
		StartedAt:   timestamp,
		Message:     fmt.Sprintf("Started Phase %s (no tasks available)", nextPhase.ID),
	}, nil
}

// findNextPendingPhase returns the first phase in pending status
func (s *AutoNextService) findNextPendingPhase(epicData *epic.Epic) *epic.Phase {
	for i := range epicData.Phases {
		if epicData.Phases[i].Status == epic.StatusPlanning {
			return &epicData.Phases[i]
		}
	}
	return nil
}

// areAllTasksCompletedOrCancelled checks if all tasks in a phase are done or cancelled
func (s *AutoNextService) areAllTasksCompletedOrCancelled(epicData *epic.Epic, phaseID string) bool {
	tasks := s.taskService.GetTasksInPhase(epicData, phaseID)
	if len(tasks) == 0 {
		return true // Empty phase is considered complete
	}

	for _, task := range tasks {
		if task.Status != epic.StatusCompleted && task.Status != epic.StatusCancelled {
			return false
		}
	}
	return true
}

// filterPendingOnly filters tasks to only include those in pending status
func filterPendingOnly(tasks []epic.Task) []epic.Task {
	var pendingTasks []epic.Task
	for _, task := range tasks {
		if task.Status == epic.StatusPlanning {
			pendingTasks = append(pendingTasks, task)
		}
	}
	return pendingTasks
}

// formatPhaseStartedXML creates XML output for phase started with task selection
func (s *AutoNextService) formatPhaseStartedXML(epicID string, phase *epic.Phase, tasks []epic.Task, startedTaskID string, timestamp time.Time) string {
	xml := fmt.Sprintf(`<phase_started epic="%s" phase="%s">
    <phase_name>%s</phase_name>
    <previous_status>pending</previous_status>
    <new_status>wip</new_status>
    <started_at>%s</started_at>
    <tasks>`,
		epicID, phase.ID, phase.Name, timestamp.Format(time.RFC3339))

	for _, task := range tasks {
		xml += fmt.Sprintf(`
        <task id="%s" status="%s">%s</task>`,
			task.ID, task.Status, task.Name)
	}

	xml += fmt.Sprintf(`
    </tasks>
    <started_task>%s</started_task>
    <message>Started Phase %s and Task %s (auto-selected)</message>
</phase_started>`, startedTaskID, phase.ID, startedTaskID)

	return xml
}

// formatTaskStartedXML creates XML output for task started in active phase
func (s *AutoNextService) formatTaskStartedXML(epicID string, task *epic.Task, timestamp time.Time) string {
	return fmt.Sprintf(`<task_started epic="%s" task="%s">
    <task_description>%s</task_description>
    <phase_id>%s</phase_id>
    <previous_status>pending</previous_status>
    <new_status>wip</new_status>
    <started_at>%s</started_at>
    <auto_selected>true</auto_selected>
    <message>Started Task %s: %s (auto-selected)</message>
</task_started>`,
		epicID, task.ID, task.Name, task.PhaseID,
		timestamp.Format(time.RFC3339), task.ID, task.Name)
}

// formatAllCompleteXML creates XML output for epic completion
func (s *AutoNextService) formatAllCompleteXML(epicID string) string {
	return fmt.Sprintf(`<all_complete epic="%s">
    <message>All phases and tasks completed. Epic ready for completion.</message>
    <suggestion>Use 'agentpm done-epic' to complete the epic</suggestion>
</all_complete>`, epicID)
}
