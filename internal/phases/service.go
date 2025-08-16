package phases

import (
	"fmt"
	"time"

	"github.com/memomoo/agentpm/internal/epic"
	"github.com/memomoo/agentpm/internal/query"
	"github.com/memomoo/agentpm/internal/storage"
)

type PhaseService struct {
	storage storage.Storage
	query   *query.QueryService
}

func NewPhaseService(storage storage.Storage, query *query.QueryService) *PhaseService {
	return &PhaseService{
		storage: storage,
		query:   query,
	}
}

// StartPhase transitions a phase from pending to wip
func (s *PhaseService) StartPhase(epicData *epic.Epic, phaseID string, timestamp time.Time) error {
	// Find the phase
	phase := s.findPhase(epicData, phaseID)
	if phase == nil {
		return fmt.Errorf("phase %s not found", phaseID)
	}

	// Validate phase can be started
	if err := s.validatePhaseStart(epicData, phase); err != nil {
		return err
	}

	// Transition phase status and set timestamp
	phase.Status = epic.StatusActive
	phase.StartedAt = &timestamp

	return nil
}

// CompletePhase transitions a phase from wip to done
func (s *PhaseService) CompletePhase(epicData *epic.Epic, phaseID string, timestamp time.Time) error {
	// Find the phase
	phase := s.findPhase(epicData, phaseID)
	if phase == nil {
		return fmt.Errorf("phase %s not found", phaseID)
	}

	// Validate phase can be completed
	if err := s.validatePhaseCompletion(epicData, phase); err != nil {
		return err
	}

	// Transition phase status and set timestamp
	phase.Status = epic.StatusCompleted
	phase.CompletedAt = &timestamp

	return nil
}

// GetActivePhase returns the currently active phase, if any
func (s *PhaseService) GetActivePhase(epicData *epic.Epic) *epic.Phase {
	for i := range epicData.Phases {
		if epicData.Phases[i].Status == epic.StatusActive {
			return &epicData.Phases[i]
		}
	}
	return nil
}

// findPhase returns a pointer to the phase with the given ID
func (s *PhaseService) findPhase(epicData *epic.Epic, phaseID string) *epic.Phase {
	for i := range epicData.Phases {
		if epicData.Phases[i].ID == phaseID {
			return &epicData.Phases[i]
		}
	}
	return nil
}

// validatePhaseStart checks if a phase can be started
func (s *PhaseService) validatePhaseStart(epicData *epic.Epic, phase *epic.Phase) error {
	// Check phase is in pending status
	if phase.Status != epic.StatusPlanning {
		return NewPhaseStateError(phase.ID, phase.Status, epic.StatusActive, "Phase is not in pending state")
	}

	// Check no other phase is active
	activePhase := s.GetActivePhase(epicData)
	if activePhase != nil && activePhase.ID != phase.ID {
		return NewPhaseConstraintError(phase.ID, activePhase.ID, "Cannot start phase: another phase is already active")
	}

	return nil
}

// validatePhaseCompletion checks if a phase can be completed
func (s *PhaseService) validatePhaseCompletion(epicData *epic.Epic, phase *epic.Phase) error {
	// Check phase is in active status
	if phase.Status != epic.StatusActive {
		return NewPhaseStateError(phase.ID, phase.Status, epic.StatusCompleted, "Phase is not in active state")
	}

	// Check all tasks in phase are completed or cancelled
	pendingTasks := s.getPendingTasksInPhase(epicData, phase.ID)
	if len(pendingTasks) > 0 {
		return NewPhaseIncompleteError(phase.ID, pendingTasks)
	}

	return nil
}

// getPendingTasksInPhase returns tasks in the phase that are not done or cancelled
func (s *PhaseService) getPendingTasksInPhase(epicData *epic.Epic, phaseID string) []epic.Task {
	var pendingTasks []epic.Task
	for _, task := range epicData.Tasks {
		if task.PhaseID == phaseID && task.Status != epic.StatusCompleted && task.Status != epic.StatusCancelled {
			pendingTasks = append(pendingTasks, task)
		}
	}
	return pendingTasks
}
