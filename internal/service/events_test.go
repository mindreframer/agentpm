package service

import (
	"testing"
	"time"

	"github.com/memomoo/agentpm/internal/epic"
)

func TestCreateEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		phaseID   string
		taskID    string
		wantType  string
		wantData  string
	}{
		{
			name:      "phase started event",
			eventType: EventPhaseStarted,
			phaseID:   "phase1",
			taskID:    "",
			wantType:  "phase_started",
			wantData:  "Phase 'Phase 1' started",
		},
		{
			name:      "phase completed event",
			eventType: EventPhaseCompleted,
			phaseID:   "phase1",
			taskID:    "",
			wantType:  "phase_completed",
			wantData:  "Phase 'Phase 1' completed",
		},
		{
			name:      "task started event",
			eventType: EventTaskStarted,
			phaseID:   "phase1",
			taskID:    "task1",
			wantType:  "task_started",
			wantData:  "Task 'Task 1' started",
		},
		{
			name:      "task completed event",
			eventType: EventTaskCompleted,
			phaseID:   "phase1",
			taskID:    "task1",
			wantType:  "task_completed",
			wantData:  "Task 'Task 1' completed",
		},
		{
			name:      "task cancelled event",
			eventType: EventTaskCancelled,
			phaseID:   "phase1",
			taskID:    "task1",
			wantType:  "task_cancelled",
			wantData:  "Task 'Task 1' cancelled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test epic with sample phase and task
			epicData := &epic.Epic{
				ID:   "test-epic",
				Name: "Test Epic",
				Phases: []epic.Phase{
					{ID: "phase1", Name: "Phase 1"},
				},
				Tasks: []epic.Task{
					{ID: "task1", Name: "Task 1", PhaseID: "phase1"},
				},
				Events: []epic.Event{},
			}

			timestamp := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

			// Create event
			CreateEvent(epicData, tt.eventType, tt.phaseID, tt.taskID, timestamp)

			// Verify event was created
			if len(epicData.Events) != 1 {
				t.Errorf("Expected 1 event, got %d", len(epicData.Events))
				return
			}

			event := epicData.Events[0]

			// Verify event type
			if event.Type != tt.wantType {
				t.Errorf("Expected event type %s, got %s", tt.wantType, event.Type)
			}

			// Verify event data
			if event.Data != tt.wantData {
				t.Errorf("Expected event data %s, got %s", tt.wantData, event.Data)
			}

			// Verify timestamp
			if !event.Timestamp.Equal(timestamp) {
				t.Errorf("Expected timestamp %v, got %v", timestamp, event.Timestamp)
			}

			// Verify event ID is generated
			if event.ID == "" {
				t.Error("Expected event ID to be generated")
			}
		})
	}
}

func TestCreateEvent_PhaseNotFound(t *testing.T) {
	epicData := &epic.Epic{
		ID:     "test-epic",
		Name:   "Test Epic",
		Phases: []epic.Phase{},
		Tasks:  []epic.Task{},
		Events: []epic.Event{},
	}

	timestamp := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create event for non-existent phase
	CreateEvent(epicData, EventPhaseStarted, "nonexistent", "", timestamp)

	// Should still create event with fallback data
	if len(epicData.Events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(epicData.Events))
		return
	}

	event := epicData.Events[0]
	expectedData := "Phase 'nonexistent' started"
	if event.Data != expectedData {
		t.Errorf("Expected event data %s, got %s", expectedData, event.Data)
	}
}

func TestCreateEvent_TaskNotFound(t *testing.T) {
	epicData := &epic.Epic{
		ID:     "test-epic",
		Name:   "Test Epic",
		Phases: []epic.Phase{},
		Tasks:  []epic.Task{},
		Events: []epic.Event{},
	}

	timestamp := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create event for non-existent task
	CreateEvent(epicData, EventTaskStarted, "phase1", "nonexistent", timestamp)

	// Should still create event with fallback data
	if len(epicData.Events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(epicData.Events))
		return
	}

	event := epicData.Events[0]
	expectedData := "Task 'nonexistent' started"
	if event.Data != expectedData {
		t.Errorf("Expected event data %s, got %s", expectedData, event.Data)
	}
}

func TestCreateEvent_MultipleEvents(t *testing.T) {
	epicData := &epic.Epic{
		ID:   "test-epic",
		Name: "Test Epic",
		Phases: []epic.Phase{
			{ID: "phase1", Name: "Phase 1"},
		},
		Tasks: []epic.Task{
			{ID: "task1", Name: "Task 1", PhaseID: "phase1"},
		},
		Events: []epic.Event{},
	}

	timestamp1 := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	timestamp2 := time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC)

	// Create multiple events
	CreateEvent(epicData, EventPhaseStarted, "phase1", "", timestamp1)
	CreateEvent(epicData, EventTaskStarted, "phase1", "task1", timestamp2)

	// Verify both events were created
	if len(epicData.Events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(epicData.Events))
		return
	}

	// Verify first event
	event1 := epicData.Events[0]
	if event1.Type != "phase_started" {
		t.Errorf("Expected first event type phase_started, got %s", event1.Type)
	}
	if !event1.Timestamp.Equal(timestamp1) {
		t.Errorf("Expected first event timestamp %v, got %v", timestamp1, event1.Timestamp)
	}

	// Verify second event
	event2 := epicData.Events[1]
	if event2.Type != "task_started" {
		t.Errorf("Expected second event type task_started, got %s", event2.Type)
	}
	if !event2.Timestamp.Equal(timestamp2) {
		t.Errorf("Expected second event timestamp %v, got %v", timestamp2, event2.Timestamp)
	}

	// Verify events have different IDs
	if event1.ID == event2.ID {
		t.Error("Expected events to have different IDs")
	}
}
