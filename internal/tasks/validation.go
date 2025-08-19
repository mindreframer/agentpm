package tasks

import (
	"fmt"

	"github.com/mindreframer/agentpm/internal/epic"
)

type TaskValidationService struct{}

func NewTaskValidationService() *TaskValidationService {
	return &TaskValidationService{}
}

func (tvs *TaskValidationService) ValidateTaskCompletion(epicData *epic.Epic, task *epic.Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	if task.Status == epic.StatusCompleted {
		return nil
	}

	return tvs.checkTaskCompletionPrerequisites(epicData, task)
}

func (tvs *TaskValidationService) checkTaskCompletionPrerequisites(epicData *epic.Epic, task *epic.Task) error {
	var blockingItems []epic.BlockingItem

	pendingTests, wipTests := tvs.countTestsByStatus(epicData, task.ID)

	// Collect blocking tests
	for _, test := range epicData.Tests {
		if test.TaskID == task.ID {
			switch test.TestStatus {
			case epic.TestStatusPending:
				blockingItems = append(blockingItems, epic.BlockingItem{
					Type:   "test",
					ID:     test.ID,
					Name:   test.Name,
					Status: string(test.TestStatus),
				})
			case epic.TestStatusWIP:
				blockingItems = append(blockingItems, epic.BlockingItem{
					Type:   "test",
					ID:     test.ID,
					Name:   test.Name,
					Status: string(test.TestStatus),
				})
			}
		}
	}

	if len(blockingItems) > 0 {
		message := fmt.Sprintf("Task %s cannot be completed due to %d blocking items: %d pending tests, %d wip tests",
			task.ID, len(blockingItems), pendingTests, wipTests)

		return &epic.StatusValidationError{
			EntityType:    "task",
			EntityID:      task.ID,
			EntityName:    task.Name,
			CurrentStatus: string(task.Status),
			TargetStatus:  string(epic.StatusCompleted),
			BlockingItems: blockingItems,
			Message:       message,
		}
	}

	return nil
}

func (tvs *TaskValidationService) countTestsByStatus(epicData *epic.Epic, taskID string) (pending int, wip int) {
	for _, test := range epicData.Tests {
		if test.TaskID == taskID {
			switch test.TestStatus {
			case epic.TestStatusPending:
				pending++
			case epic.TestStatusWIP:
				wip++
			}
		}
	}
	return pending, wip
}

func (tvs *TaskValidationService) ValidateTaskStatusTransition(currentStatus, targetStatus epic.Status) error {
	validTransitions := map[epic.Status][]epic.Status{
		epic.StatusPending:   {epic.StatusWIP, epic.StatusCancelled},
		epic.StatusWIP:       {epic.StatusCompleted, epic.StatusCancelled, epic.StatusOnHold},
		epic.StatusOnHold:    {epic.StatusWIP, epic.StatusCancelled},
		epic.StatusCompleted: {},
		epic.StatusCancelled: {},
	}

	validTargets, exists := validTransitions[currentStatus]
	if !exists {
		return fmt.Errorf("invalid current status: %s", currentStatus)
	}

	for _, validTarget := range validTargets {
		if validTarget == targetStatus {
			return nil
		}
	}

	return fmt.Errorf("invalid status transition from %s to %s", currentStatus, targetStatus)
}

func (tvs *TaskValidationService) ValidateTaskCancellation(task *epic.Task, reason string) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	if task.Status == epic.StatusCancelled {
		return fmt.Errorf("task %s is already cancelled", task.ID)
	}

	if task.Status == epic.StatusCompleted {
		return fmt.Errorf("cannot cancel completed task %s", task.ID)
	}

	if reason == "" {
		return fmt.Errorf("cancellation reason is required for task %s", task.ID)
	}

	return nil
}

func (tvs *TaskValidationService) CanCompleteTask(epicData *epic.Epic, task *epic.Task) (bool, error) {
	err := tvs.ValidateTaskCompletion(epicData, task)
	if err != nil {
		return false, err
	}
	return true, nil
}
