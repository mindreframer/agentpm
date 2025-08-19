package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
)

func TestEpic4FieldsPersistence(t *testing.T) {
	// Create test epic with Epic 4 fields
	now := time.Date(2025, 8, 16, 14, 30, 0, 0, time.UTC)
	testEpic := &epic.Epic{
		ID:     "test-epic",
		Name:   "Test Epic",
		Status: epic.StatusWIP,
		Phases: []epic.Phase{
			{ID: "phase-1", Name: "Phase 1", Status: epic.StatusWIP},
		},
		Tasks: []epic.Task{
			{ID: "task-1", PhaseID: "phase-1", Name: "Task 1", Status: epic.StatusWIP},
		},
		Tests: []epic.Test{
			{
				ID:                 "test-1",
				TaskID:             "task-1",
				PhaseID:            "phase-1",
				Name:               "Test 1",
				Status:             epic.StatusWIP,
				TestStatus:         epic.TestStatusWIP,
				Description:        "Test with Epic 4 fields",
				StartedAt:          &now,
				FailureNote:        "Test failure note",
				CancellationReason: "Test cancellation reason",
			},
		},
		Events: []epic.Event{},
	}

	// Save and load the epic
	tempDir := t.TempDir()
	epicFile := filepath.Join(tempDir, "test-epic.xml")
	storage := NewFileStorage()

	fmt.Printf("SAVING epic with:\n")
	fmt.Printf("  TestStatus: '%s'\n", testEpic.Tests[0].TestStatus)
	fmt.Printf("  PhaseID: '%s'\n", testEpic.Tests[0].PhaseID)
	fmt.Printf("  StartedAt: %v\n", testEpic.Tests[0].StartedAt)
	fmt.Printf("  FailureNote: '%s'\n", testEpic.Tests[0].FailureNote)

	err := storage.SaveEpic(testEpic, epicFile)
	if err != nil {
		t.Fatalf("Failed to save epic: %v", err)
	}

	// Read the saved XML file to see what was actually saved
	content, err := os.ReadFile(epicFile)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}
	fmt.Printf("\nSaved XML:\n%s\n", string(content))

	// Load the epic back
	loadedEpic, err := storage.LoadEpic(epicFile)
	if err != nil {
		t.Fatalf("Failed to load epic: %v", err)
	}

	fmt.Printf("\nLOADED epic with:\n")
	fmt.Printf("  TestStatus: '%s'\n", loadedEpic.Tests[0].TestStatus)
	fmt.Printf("  PhaseID: '%s'\n", loadedEpic.Tests[0].PhaseID)
	fmt.Printf("  StartedAt: %v\n", loadedEpic.Tests[0].StartedAt)
	fmt.Printf("  FailureNote: '%s'\n", loadedEpic.Tests[0].FailureNote)

	// Verify all Epic 4 fields are preserved
	test := loadedEpic.Tests[0]

	if test.TestStatus != epic.TestStatusWIP {
		t.Errorf("TestStatus not preserved: expected '%s', got '%s'", epic.TestStatusWIP, test.TestStatus)
	}

	if test.PhaseID != "phase-1" {
		t.Errorf("PhaseID not preserved: expected 'phase-1', got '%s'", test.PhaseID)
	}

	if test.StartedAt == nil {
		t.Error("StartedAt not preserved")
	} else if !test.StartedAt.Equal(now) {
		t.Errorf("StartedAt not preserved correctly: expected %v, got %v", now, *test.StartedAt)
	}

	if test.FailureNote != "Test failure note" {
		t.Errorf("FailureNote not preserved: expected 'Test failure note', got '%s'", test.FailureNote)
	}

	if test.CancellationReason != "Test cancellation reason" {
		t.Errorf("CancellationReason not preserved: expected 'Test cancellation reason', got '%s'", test.CancellationReason)
	}

	fmt.Printf("\n✅ All Epic 4 fields preserved correctly!\n")
}
