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
	EventTestStarted    EventType = "test_started"
	EventTestPassed     EventType = "test_passed"
	EventTestFailed     EventType = "test_failed"
	EventTestCancelled  EventType = "test_cancelled"
	EventEpicStarted    EventType = "epic_started"
	EventEpicCompleted  EventType = "epic_completed"
)

// CreateEvent creates a new event and appends it to the epic's events
// Only creates an event if the referenced entity (phase, task, test, or epic) exists
func CreateEvent(epicData *epic.Epic, eventType EventType, phaseID, taskID, testID, reason string, timestamp time.Time) {
	// Create the event data string based on the event type
	var data string
	var entityExists bool

	switch eventType {
	case EventPhaseStarted:
		phase := findPhaseByID(epicData, phaseID)
		if phase != nil {
			entityExists = true
			if phase.Name != "" {
				data = fmt.Sprintf("Phase %s (%s) started", phase.ID, phase.Name)
			} else {
				data = fmt.Sprintf("Phase %s started", phase.ID)
			}
		}
	case EventPhaseCompleted:
		phase := findPhaseByID(epicData, phaseID)
		if phase != nil {
			entityExists = true
			if phase.Name != "" {
				data = fmt.Sprintf("Phase %s (%s) completed", phase.ID, phase.Name)
			} else {
				data = fmt.Sprintf("Phase %s completed", phase.ID)
			}
		}
	case EventTaskStarted:
		task := findTaskByID(epicData, taskID)
		if task != nil {
			entityExists = true
			if task.Name != "" {
				data = fmt.Sprintf("Task %s (%s) started", task.ID, task.Name)
			} else {
				data = fmt.Sprintf("Task %s started", task.ID)
			}
		}
	case EventTaskCompleted:
		task := findTaskByID(epicData, taskID)
		if task != nil {
			entityExists = true
			if task.Name != "" {
				data = fmt.Sprintf("Task %s (%s) completed", task.ID, task.Name)
			} else {
				data = fmt.Sprintf("Task %s completed", task.ID)
			}
		}
	case EventTaskCancelled:
		task := findTaskByID(epicData, taskID)
		if task != nil {
			entityExists = true
			if task.Name != "" {
				data = fmt.Sprintf("Task %s (%s) cancelled", task.ID, task.Name)
			} else {
				data = fmt.Sprintf("Task %s cancelled", task.ID)
			}
		}
	case EventTestStarted:
		test := findTestByID(epicData, testID)
		if test != nil {
			entityExists = true
			data = formatTestStartedData(test)
		}
	case EventTestPassed:
		test := findTestByID(epicData, testID)
		if test != nil {
			entityExists = true
			data = formatTestPassedData(test)
		}
	case EventTestFailed:
		test := findTestByID(epicData, testID)
		if test != nil {
			entityExists = true
			data = formatTestFailedData(test, reason)
		}
	case EventTestCancelled:
		test := findTestByID(epicData, testID)
		if test != nil {
			entityExists = true
			data = formatTestCancelledData(test, reason)
		}
	case EventEpicStarted:
		entityExists = true
		data = formatEpicStartedData(epicData)
	case EventEpicCompleted:
		entityExists = true
		data = formatEpicCompletedData(epicData)
	default:
		// For unknown event types, we don't validate entity existence
		entityExists = true
		data = fmt.Sprintf("Event of type %s occurred", string(eventType))
	}

	// Only create event if the entity exists
	if !entityExists {
		return
	}

	// Generate a simple event ID based on timestamp and type
	eventID := fmt.Sprintf("%s_%d", string(eventType), timestamp.Unix())

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

func findTestByID(epicData *epic.Epic, testID string) *epic.Test {
	for i := range epicData.Tests {
		if epicData.Tests[i].ID == testID {
			return &epicData.Tests[i]
		}
	}
	return nil
}

// Test event data formatting functions
func formatTestStartedData(test *epic.Test) string {
	if test.Name != "" {
		return fmt.Sprintf("Test %s (%s) started", test.ID, test.Name)
	}
	return fmt.Sprintf("Test %s started", test.ID)
}

func formatTestPassedData(test *epic.Test) string {
	if test.Name != "" {
		return fmt.Sprintf("Test %s (%s) passed", test.ID, test.Name)
	}
	return fmt.Sprintf("Test %s passed", test.ID)
}

func formatTestFailedData(test *epic.Test, reason string) string {
	baseData := ""
	if test.Name != "" {
		baseData = fmt.Sprintf("Test %s (%s) failed", test.ID, test.Name)
	} else {
		baseData = fmt.Sprintf("Test %s failed", test.ID)
	}

	if reason != "" {
		baseData += fmt.Sprintf(": %s", reason)
	}
	return baseData
}

func formatTestCancelledData(test *epic.Test, reason string) string {
	baseData := ""
	if test.Name != "" {
		baseData = fmt.Sprintf("Test %s (%s) cancelled", test.ID, test.Name)
	} else {
		baseData = fmt.Sprintf("Test %s cancelled", test.ID)
	}

	if reason != "" {
		baseData += fmt.Sprintf(": %s", reason)
	}
	return baseData
}

// Epic event data formatting functions
func formatEpicStartedData(epicData *epic.Epic) string {
	if epicData.Name != "" {
		return fmt.Sprintf("Epic %s started", epicData.Name)
	}
	return fmt.Sprintf("Epic %s started", epicData.ID)
}

func formatEpicCompletedData(epicData *epic.Epic) string {
	if epicData.Name != "" {
		return fmt.Sprintf("Epic %s completed", epicData.Name)
	}
	return fmt.Sprintf("Epic %s completed", epicData.ID)
}
