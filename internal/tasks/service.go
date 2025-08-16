package tasks

import (
	"fmt"
	"time"

	"github.com/memomoo/agentpm/internal/epic"
	"github.com/memomoo/agentpm/internal/query"
	"github.com/memomoo/agentpm/internal/service"
	"github.com/memomoo/agentpm/internal/storage"
)

type TaskService struct {
	storage storage.Storage
	query   *query.QueryService
}

func NewTaskService(storage storage.Storage, query *query.QueryService) *TaskService {
	return &TaskService{
		storage: storage,
		query:   query,
	}
}

// StartTask transitions a task from pending to wip
func (s *TaskService) StartTask(epicData *epic.Epic, taskID string, timestamp time.Time) error {
	// Find the task
	task := s.findTask(epicData, taskID)
	if task == nil {
		return fmt.Errorf("task %s not found", taskID)
	}

	// Validate task can be started
	if err := s.validateTaskStart(epicData, task); err != nil {
		return err
	}

	// Transition task status and set timestamp
	task.Status = epic.StatusActive
	task.StartedAt = &timestamp

	// Create automatic event for task start
	service.CreateEvent(epicData, service.EventTaskStarted, task.PhaseID, taskID, timestamp)

	return nil
}

// CompleteTask transitions a task from wip to done
func (s *TaskService) CompleteTask(epicData *epic.Epic, taskID string, timestamp time.Time) error {
	// Find the task
	task := s.findTask(epicData, taskID)
	if task == nil {
		return fmt.Errorf("task %s not found", taskID)
	}

	// Validate task can be completed
	if err := s.validateTaskCompletion(epicData, task); err != nil {
		return err
	}

	// Transition task status and set timestamp
	task.Status = epic.StatusCompleted
	task.CompletedAt = &timestamp

	// Create automatic event for task completion
	service.CreateEvent(epicData, service.EventTaskCompleted, task.PhaseID, taskID, timestamp)

	return nil
}

// CancelTask transitions a task from wip to cancelled
func (s *TaskService) CancelTask(epicData *epic.Epic, taskID string, timestamp time.Time) error {
	// Find the task
	task := s.findTask(epicData, taskID)
	if task == nil {
		return fmt.Errorf("task %s not found", taskID)
	}

	// Validate task can be cancelled
	if err := s.validateTaskCancellation(epicData, task); err != nil {
		return err
	}

	// Transition task status and set timestamp
	task.Status = epic.StatusCancelled
	task.CancelledAt = &timestamp

	// Create automatic event for task cancellation
	service.CreateEvent(epicData, service.EventTaskCancelled, task.PhaseID, taskID, timestamp)

	return nil
}

// GetActiveTask returns the currently active task in the given phase, if any
func (s *TaskService) GetActiveTask(epicData *epic.Epic, phaseID string) *epic.Task {
	for i := range epicData.Tasks {
		if epicData.Tasks[i].PhaseID == phaseID && epicData.Tasks[i].Status == epic.StatusActive {
			return &epicData.Tasks[i]
		}
	}
	return nil
}

// GetActiveTaskInEpic returns the currently active task in the entire epic, if any
func (s *TaskService) GetActiveTaskInEpic(epicData *epic.Epic) *epic.Task {
	for i := range epicData.Tasks {
		if epicData.Tasks[i].Status == epic.StatusActive {
			return &epicData.Tasks[i]
		}
	}
	return nil
}

// findTask returns a pointer to the task with the given ID
func (s *TaskService) findTask(epicData *epic.Epic, taskID string) *epic.Task {
	for i := range epicData.Tasks {
		if epicData.Tasks[i].ID == taskID {
			return &epicData.Tasks[i]
		}
	}
	return nil
}

// findPhase returns a pointer to the phase with the given ID
func (s *TaskService) findPhase(epicData *epic.Epic, phaseID string) *epic.Phase {
	for i := range epicData.Phases {
		if epicData.Phases[i].ID == phaseID {
			return &epicData.Phases[i]
		}
	}
	return nil
}

// validateTaskStart checks if a task can be started
func (s *TaskService) validateTaskStart(epicData *epic.Epic, task *epic.Task) error {
	// Check task is in pending status
	if task.Status != epic.StatusPlanning {
		return NewTaskStateError(task.ID, task.Status, epic.StatusActive, "Task is not in pending state")
	}

	// Check task's phase is active
	phase := s.findPhase(epicData, task.PhaseID)
	if phase == nil {
		return fmt.Errorf("phase %s not found for task %s", task.PhaseID, task.ID)
	}

	if phase.Status != epic.StatusActive {
		return NewTaskPhaseError(task.ID, task.PhaseID, phase.Status, "Cannot start task: phase is not active")
	}

	// Check no other task is active in the same phase
	activeTask := s.GetActiveTask(epicData, task.PhaseID)
	if activeTask != nil && activeTask.ID != task.ID {
		return NewTaskConstraintError(task.ID, activeTask.ID, task.PhaseID, "Cannot start task: another task is already active in this phase")
	}

	return nil
}

// validateTaskCompletion checks if a task can be completed
func (s *TaskService) validateTaskCompletion(epicData *epic.Epic, task *epic.Task) error {
	// Check task is in active status
	if task.Status != epic.StatusActive {
		return NewTaskStateError(task.ID, task.Status, epic.StatusCompleted, "Task is not in active state")
	}

	return nil
}

// validateTaskCancellation checks if a task can be cancelled
func (s *TaskService) validateTaskCancellation(epicData *epic.Epic, task *epic.Task) error {
	// Check task is in active status
	if task.Status != epic.StatusActive {
		return NewTaskStateError(task.ID, task.Status, epic.StatusCancelled, "Task is not in active state")
	}

	return nil
}

// GetTasksInPhase returns all tasks for a given phase
func (s *TaskService) GetTasksInPhase(epicData *epic.Epic, phaseID string) []epic.Task {
	var tasks []epic.Task
	for _, task := range epicData.Tasks {
		if task.PhaseID == phaseID {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// GetPendingTasksInPhase returns tasks in the phase that are not done or cancelled
func (s *TaskService) GetPendingTasksInPhase(epicData *epic.Epic, phaseID string) []epic.Task {
	var pendingTasks []epic.Task
	for _, task := range epicData.Tasks {
		if task.PhaseID == phaseID && task.Status != epic.StatusCompleted && task.Status != epic.StatusCancelled {
			pendingTasks = append(pendingTasks, task)
		}
	}
	return pendingTasks
}
