package phases

import (
	"fmt"

	"github.com/mindreframer/agentpm/internal/epic"
)

type PhaseValidationService struct{}

func NewPhaseValidationService() *PhaseValidationService {
	return &PhaseValidationService{}
}

func (pvs *PhaseValidationService) ValidatePhaseCompletion(epicData *epic.Epic, phase *epic.Phase) error {
	if phase == nil {
		return fmt.Errorf("phase cannot be nil")
	}

	if phase.Status == epic.StatusCompleted {
		return nil
	}

	return pvs.checkPhaseCompletionPrerequisites(epicData, phase)
}

func (pvs *PhaseValidationService) checkPhaseCompletionPrerequisites(epicData *epic.Epic, phase *epic.Phase) error {
	var blockingItems []epic.BlockingItem

	pendingTasks, activeTasks := pvs.countTasksByStatus(epicData, phase.ID)
	pendingTests, wipTests := pvs.countTestsByStatus(epicData, phase.ID)

	// Collect blocking tasks
	for _, task := range epicData.Tasks {
		if task.PhaseID == phase.ID {
			switch task.Status {
			case epic.StatusPending:
				blockingItems = append(blockingItems, epic.BlockingItem{
					Type:   "task",
					ID:     task.ID,
					Name:   task.Name,
					Status: string(task.Status),
				})
			case epic.StatusWIP:
				blockingItems = append(blockingItems, epic.BlockingItem{
					Type:   "task",
					ID:     task.ID,
					Name:   task.Name,
					Status: string(task.Status),
				})
			}
		}
	}

	// Collect blocking tests
	for _, test := range epicData.Tests {
		if test.PhaseID == phase.ID {
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
		message := fmt.Sprintf("Phase %s cannot be completed due to %d blocking items: %d pending tasks, %d active tasks, %d pending tests, %d wip tests",
			phase.ID, len(blockingItems), pendingTasks, activeTasks, pendingTests, wipTests)

		return &epic.StatusValidationError{
			EntityType:    "phase",
			EntityID:      phase.ID,
			EntityName:    phase.Name,
			CurrentStatus: string(phase.Status),
			TargetStatus:  string(epic.StatusCompleted),
			BlockingItems: blockingItems,
			Message:       message,
		}
	}

	return nil
}

func (pvs *PhaseValidationService) countTasksByStatus(epicData *epic.Epic, phaseID string) (pending int, active int) {
	for _, task := range epicData.Tasks {
		if task.PhaseID == phaseID {
			switch task.Status {
			case epic.StatusPending:
				pending++
			case epic.StatusWIP:
				active++
			}
		}
	}
	return pending, active
}

func (pvs *PhaseValidationService) countTestsByStatus(epicData *epic.Epic, phaseID string) (pending int, wip int) {
	for _, test := range epicData.Tests {
		if test.PhaseID == phaseID {
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

func (pvs *PhaseValidationService) ValidatePhaseStatusTransition(currentStatus, targetStatus epic.Status) error {
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

func (pvs *PhaseValidationService) CanCompletePhase(epicData *epic.Epic, phase *epic.Phase) (bool, error) {
	err := pvs.ValidatePhaseCompletion(epicData, phase)
	if err != nil {
		return false, err
	}
	return true, nil
}
