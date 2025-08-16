package service

import (
	"fmt"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
)

// EventType represents the different types of events that can be created
type EventType string

const (
	EventPhaseStarted   EventType = "phase_started"
	EventPhaseCompleted EventType = "phase_completed"
	EventTaskStarted    EventType = "task_started"
	EventTaskCompleted  EventType = "task_completed"
	EventTaskCancelled  EventType = "task_cancelled"
)

// CreateEvent creates a new event and appends it to the epic's events
func CreateEvent(epicData *epic.Epic, eventType EventType, phaseID, taskID string, timestamp time.Time) {
	// Generate a simple event ID based on timestamp and type
	eventID := fmt.Sprintf("%s_%d", string(eventType), timestamp.Unix())

	// Create the event data string based on the event type
	var data string
	switch eventType {
	case EventPhaseStarted:
		phase := findPhaseByID(epicData, phaseID)
		if phase != nil {
			data = fmt.Sprintf("Phase '%s' started", phase.Name)
		} else {
			data = fmt.Sprintf("Phase '%s' started", phaseID)
		}
	case EventPhaseCompleted:
		phase := findPhaseByID(epicData, phaseID)
		if phase != nil {
			data = fmt.Sprintf("Phase '%s' completed", phase.Name)
		} else {
			data = fmt.Sprintf("Phase '%s' completed", phaseID)
		}
	case EventTaskStarted:
		task := findTaskByID(epicData, taskID)
		if task != nil {
			data = fmt.Sprintf("Task '%s' started", task.Name)
		} else {
			data = fmt.Sprintf("Task '%s' started", taskID)
		}
	case EventTaskCompleted:
		task := findTaskByID(epicData, taskID)
		if task != nil {
			data = fmt.Sprintf("Task '%s' completed", task.Name)
		} else {
			data = fmt.Sprintf("Task '%s' completed", taskID)
		}
	case EventTaskCancelled:
		task := findTaskByID(epicData, taskID)
		if task != nil {
			data = fmt.Sprintf("Task '%s' cancelled", task.Name)
		} else {
			data = fmt.Sprintf("Task '%s' cancelled", taskID)
		}
	default:
		data = fmt.Sprintf("Event of type %s occurred", string(eventType))
	}

	// Create and append the event
	event := epic.Event{
		ID:        eventID,
		Type:      string(eventType),
		Timestamp: timestamp,
		Data:      data,
	}

	epicData.Events = append(epicData.Events, event)
}

// Helper functions to find phases and tasks by ID
func findPhaseByID(epicData *epic.Epic, phaseID string) *epic.Phase {
	for i := range epicData.Phases {
		if epicData.Phases[i].ID == phaseID {
			return &epicData.Phases[i]
		}
	}
	return nil
}

func findTaskByID(epicData *epic.Epic, taskID string) *epic.Task {
	for i := range epicData.Tasks {
		if epicData.Tasks[i].ID == taskID {
			return &epicData.Tasks[i]
		}
	}
	return nil
}
