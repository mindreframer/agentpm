package service

import (
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
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
			wantData:  "Phase phase1 (Phase 1) started",
		},
		{
			name:      "phase completed event",
			eventType: EventPhaseCompleted,
			phaseID:   "phase1",
			taskID:    "",
			wantType:  "phase_completed",
			wantData:  "Phase phase1 (Phase 1) completed",
		},
		{
			name:      "task started event",
			eventType: EventTaskStarted,
			phaseID:   "phase1",
			taskID:    "task1",
			wantType:  "task_started",
			wantData:  "Task task1 (Task 1) started",
		},
		{
			name:      "task completed event",
			eventType: EventTaskCompleted,
			phaseID:   "phase1",
			taskID:    "task1",
			wantType:  "task_completed",
			wantData:  "Task task1 (Task 1) completed",
		},
		{
			name:      "task cancelled event",
			eventType: EventTaskCancelled,
			phaseID:   "phase1",
			taskID:    "task1",
			wantType:  "task_cancelled",
			wantData:  "Task task1 (Task 1) cancelled",
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
			CreateEvent(epicData, tt.eventType, tt.phaseID, tt.taskID, "", "", timestamp)

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
	CreateEvent(epicData, EventPhaseStarted, "nonexistent", "", "", "", timestamp)

	// Should NOT create event when entity doesn't exist
	if len(epicData.Events) != 0 {
		t.Errorf("Expected 0 events, got %d", len(epicData.Events))
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
	CreateEvent(epicData, EventTaskStarted, "phase1", "nonexistent", "", "", timestamp)

	// Should NOT create event when entity doesn't exist
	if len(epicData.Events) != 0 {
		t.Errorf("Expected 0 events, got %d", len(epicData.Events))
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
	CreateEvent(epicData, EventPhaseStarted, "phase1", "", "", "", timestamp1)
	CreateEvent(epicData, EventTaskStarted, "phase1", "task1", "", "", timestamp2)

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

func TestCreateEvent_EntityWithoutName(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		phaseID   string
		taskID    string
		wantData  string
	}{
		{
			name:      "phase with empty name",
			eventType: EventPhaseStarted,
			phaseID:   "phase1",
			taskID:    "",
			wantData:  "Phase phase1 started",
		},
		{
			name:      "task with empty name",
			eventType: EventTaskStarted,
			phaseID:   "phase1",
			taskID:    "task1",
			wantData:  "Task task1 started",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test epic with entities that have empty names
			epicData := &epic.Epic{
				ID:   "test-epic",
				Name: "Test Epic",
				Phases: []epic.Phase{
					{ID: "phase1", Name: ""}, // Empty name
				},
				Tasks: []epic.Task{
					{ID: "task1", Name: "", PhaseID: "phase1"}, // Empty name
				},
				Events: []epic.Event{},
			}

			timestamp := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

			// Create event
			CreateEvent(epicData, tt.eventType, tt.phaseID, tt.taskID, "", "", timestamp)

			// Verify event was created
			if len(epicData.Events) != 1 {
				t.Errorf("Expected 1 event, got %d", len(epicData.Events))
				return
			}

			event := epicData.Events[0]

			// Verify event data uses ID-only format when name is empty
			if event.Data != tt.wantData {
				t.Errorf("Expected event data %s, got %s", tt.wantData, event.Data)
			}
		})
	}
}

func TestCreateTestEvents(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		testID    string
		reason    string
		wantData  string
	}{
		{
			name:      "Test started event with name",
			eventType: EventTestStarted,
			testID:    "test1",
			wantData:  "Test test1 (Unit Test for Database) started",
		},
		{
			name:      "Test passed event with name",
			eventType: EventTestPassed,
			testID:    "test1",
			wantData:  "Test test1 (Unit Test for Database) passed",
		},
		{
			name:      "Test failed event with reason",
			eventType: EventTestFailed,
			testID:    "test1",
			reason:    "Connection timeout",
			wantData:  "Test test1 (Unit Test for Database) failed: Connection timeout",
		},
		{
			name:      "Test cancelled event with reason",
			eventType: EventTestCancelled,
			testID:    "test1",
			reason:    "Requirements changed",
			wantData:  "Test test1 (Unit Test for Database) cancelled: Requirements changed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create epic with test
			epicData := &epic.Epic{
				ID: "epic1",
				Tests: []epic.Test{
					{
						ID:   "test1",
						Name: "Unit Test for Database",
					},
				},
			}

			timestamp := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

			// Create event
			CreateEvent(epicData, tt.eventType, "", "", tt.testID, tt.reason, timestamp)

			// Verify event was created
			if len(epicData.Events) != 1 {
				t.Errorf("Expected 1 event, got %d", len(epicData.Events))
				return
			}

			event := epicData.Events[0]

			// Verify event details
			if event.Type != string(tt.eventType) {
				t.Errorf("Expected event type %s, got %s", tt.eventType, event.Type)
			}

			if event.Data != tt.wantData {
				t.Errorf("Expected event data %s, got %s", tt.wantData, event.Data)
			}

			if !event.Timestamp.Equal(timestamp) {
				t.Errorf("Expected timestamp %v, got %v", timestamp, event.Timestamp)
			}
		})
	}
}

func TestCreateTestEventsNonExistentTest(t *testing.T) {
	// Create epic without tests
	epicData := &epic.Epic{
		ID: "epic1",
	}

	timestamp := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create event for non-existent test
	CreateEvent(epicData, EventTestStarted, "", "", "nonexistent", "", timestamp)

	// Should NOT create event when test doesn't exist
	if len(epicData.Events) != 0 {
		t.Errorf("Expected 0 events for non-existent test, got %d", len(epicData.Events))
	}
}

func TestCreateEpicEvents(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		epicName  string
		wantData  string
	}{
		{
			name:      "Epic started event with name",
			eventType: EventEpicStarted,
			epicName:  "CLI Framework Development",
			wantData:  "Epic CLI Framework Development started",
		},
		{
			name:      "Epic completed event with name",
			eventType: EventEpicCompleted,
			epicName:  "CLI Framework Development",
			wantData:  "Epic CLI Framework Development completed",
		},
		{
			name:      "Epic started event without name",
			eventType: EventEpicStarted,
			epicName:  "",
			wantData:  "Epic epic1 started",
		},
		{
			name:      "Epic completed event without name",
			eventType: EventEpicCompleted,
			epicName:  "",
			wantData:  "Epic epic1 completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create epic
			epicData := &epic.Epic{
				ID:   "epic1",
				Name: tt.epicName,
			}

			timestamp := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

			// Create event
			CreateEvent(epicData, tt.eventType, "", "", "", "", timestamp)

			// Verify event was created
			if len(epicData.Events) != 1 {
				t.Errorf("Expected 1 event, got %d", len(epicData.Events))
				return
			}

			event := epicData.Events[0]

			// Verify event details
			if event.Type != string(tt.eventType) {
				t.Errorf("Expected event type %s, got %s", tt.eventType, event.Type)
			}

			if event.Data != tt.wantData {
				t.Errorf("Expected event data %s, got %s", tt.wantData, event.Data)
			}

			if !event.Timestamp.Equal(timestamp) {
				t.Errorf("Expected timestamp %v, got %v", timestamp, event.Timestamp)
			}
		})
	}
}

func TestFormatTestEventData(t *testing.T) {
	test := &epic.Test{
		ID:   "test1",
		Name: "Unit Test for Database",
	}

	tests := []struct {
		name     string
		fn       func(*epic.Test, string) string
		reason   string
		expected string
	}{
		{
			name:     "Format test failed data with reason",
			fn:       func(t *epic.Test, r string) string { return formatTestFailedData(t, r) },
			reason:   "Connection timeout",
			expected: "Test test1 (Unit Test for Database) failed: Connection timeout",
		},
		{
			name:     "Format test failed data without reason",
			fn:       func(t *epic.Test, r string) string { return formatTestFailedData(t, r) },
			reason:   "",
			expected: "Test test1 (Unit Test for Database) failed",
		},
		{
			name:     "Format test cancelled data with reason",
			fn:       func(t *epic.Test, r string) string { return formatTestCancelledData(t, r) },
			reason:   "Requirements changed",
			expected: "Test test1 (Unit Test for Database) cancelled: Requirements changed",
		},
		{
			name:     "Format test cancelled data without reason",
			fn:       func(t *epic.Test, r string) string { return formatTestCancelledData(t, r) },
			reason:   "",
			expected: "Test test1 (Unit Test for Database) cancelled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(test, tt.reason)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFormatTestEventDataNoName(t *testing.T) {
	test := &epic.Test{
		ID:   "test1",
		Name: "",
	}

	tests := []struct {
		name     string
		fn       func(*epic.Test) string
		expected string
	}{
		{
			name:     "Format test started data without name",
			fn:       formatTestStartedData,
			expected: "Test test1 started",
		},
		{
			name:     "Format test passed data without name",
			fn:       formatTestPassedData,
			expected: "Test test1 passed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(test)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFindTestByID(t *testing.T) {
	epicData := &epic.Epic{
		ID: "epic1",
		Tests: []epic.Test{
			{ID: "test1", Name: "Test 1"},
			{ID: "test2", Name: "Test 2"},
		},
	}

	// Test finding existing test
	test := findTestByID(epicData, "test1")
	if test == nil {
		t.Error("Expected to find test1, got nil")
	} else if test.ID != "test1" {
		t.Errorf("Expected test ID test1, got %s", test.ID)
	}

	// Test finding non-existent test
	test = findTestByID(epicData, "nonexistent")
	if test != nil {
		t.Errorf("Expected nil for non-existent test, got %v", test)
	}
}
