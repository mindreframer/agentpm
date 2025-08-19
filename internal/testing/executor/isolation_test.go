package executor

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/testing/builders"
)

func TestTestExecutionEnvironment_BasicOperations(t *testing.T) {
	env := NewTestExecutionEnvironment("test-epic.xml")

	// Create a test epic
	testEpic, err := builders.NewEpicBuilder("test-epic").
		WithPhase("1A", "Setup", "pending").
		WithTask("1A_1", "1A", "Init", "pending").
		Build()

	if err != nil {
		t.Fatalf("Failed to build test epic: %v", err)
	}

	// Load epic into environment
	err = env.LoadEpic(testEpic)
	if err != nil {
		t.Fatalf("Failed to load epic: %v", err)
	}

	// Get current epic and verify it matches
	currentEpic, err := env.GetCurrentEpic()
	if err != nil {
		t.Fatalf("Failed to get current epic: %v", err)
	}

	if currentEpic.ID != "test-epic" {
		t.Errorf("Expected epic ID 'test-epic', got: %s", currentEpic.ID)
	}

	if len(currentEpic.Phases) != 1 {
		t.Errorf("Expected 1 phase, got: %d", len(currentEpic.Phases))
	}

	// Verify initial snapshot was taken
	snapshots := env.GetSnapshots()
	if len(snapshots) != 1 {
		t.Fatalf("Expected 1 initial snapshot, got: %d", len(snapshots))
	}

	if snapshots[0].Command != "initial_load" {
		t.Errorf("Expected initial snapshot command 'initial_load', got: %s", snapshots[0].Command)
	}
}

func TestTestExecutionEnvironment_StateSnapshots(t *testing.T) {
	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	env := NewTestExecutionEnvironment("test-epic.xml").WithTimeSource(func() time.Time { return fixedTime })

	// Create and load test epic
	testEpic, err := builders.NewEpicBuilder("test-epic").
		WithPhase("1A", "Setup", "pending").
		Build()

	if err != nil {
		t.Fatalf("Failed to build test epic: %v", err)
	}

	err = env.LoadEpic(testEpic)
	if err != nil {
		t.Fatalf("Failed to load epic: %v", err)
	}

	// Modify epic state and save with different commands
	testEpic.Status = epic.StatusActive
	err = env.SaveEpic(testEpic, "start_epic")
	if err != nil {
		t.Fatalf("Failed to save epic: %v", err)
	}

	testEpic.Phases[0].Status = epic.StatusActive
	err = env.SaveEpic(testEpic, "start_phase_1A")
	if err != nil {
		t.Fatalf("Failed to save epic: %v", err)
	}

	// Verify snapshots
	snapshots := env.GetSnapshots()
	if len(snapshots) != 3 {
		t.Fatalf("Expected 3 snapshots, got: %d", len(snapshots))
	}

	// Check snapshot details
	expectedCommands := []string{"initial_load", "start_epic", "start_phase_1A"}
	for i, expected := range expectedCommands {
		if snapshots[i].Command != expected {
			t.Errorf("Snapshot %d: expected command '%s', got: '%s'", i, expected, snapshots[i].Command)
		}
		if !snapshots[i].Success {
			t.Errorf("Snapshot %d: expected success=true, got: %v", i, snapshots[i].Success)
		}
		if !snapshots[i].Timestamp.Equal(fixedTime) {
			t.Errorf("Snapshot %d: expected timestamp %v, got: %v", i, fixedTime, snapshots[i].Timestamp)
		}
	}

	// Verify epic status progression
	if snapshots[0].EpicState.Status != epic.StatusPending {
		t.Errorf("Initial snapshot: expected status planning, got: %s", snapshots[0].EpicState.Status)
	}
	if snapshots[1].EpicState.Status != epic.StatusActive {
		t.Errorf("Second snapshot: expected status active, got: %s", snapshots[1].EpicState.Status)
	}
	if snapshots[2].EpicState.Phases[0].Status != epic.StatusActive {
		t.Errorf("Third snapshot: expected phase status active, got: %s", snapshots[2].EpicState.Phases[0].Status)
	}
}

func TestTestExecutionEnvironment_ErrorHandling(t *testing.T) {
	env := NewTestExecutionEnvironment("test-epic.xml")

	// Test loading nil epic
	err := env.LoadEpic(nil)
	if err == nil {
		t.Error("Expected error when loading nil epic")
	}

	// Test getting epic before loading
	_, err = env.GetCurrentEpic()
	if err == nil {
		t.Error("Expected error when getting epic before loading")
	}

	// Load valid epic first
	testEpic, err := builders.NewEpicBuilder("test-epic").Build()
	if err != nil {
		t.Fatalf("Failed to build test epic: %v", err)
	}

	err = env.LoadEpic(testEpic)
	if err != nil {
		t.Fatalf("Failed to load epic: %v", err)
	}

	// Test error snapshot by trying to save nil epic (this will create an error snapshot)
	// We'll simulate this by creating a storage error scenario
	// For now, we'll test the path where SaveEpic succeeds but we check error metadata

	metadata := env.GetExecutionMetadata()
	if metadata.SuccessCount != 1 { // initial load
		t.Errorf("Expected 1 successful operation, got: %d", metadata.SuccessCount)
	}
	if metadata.ErrorCount != 0 {
		t.Errorf("Expected 0 errors, got: %d", metadata.ErrorCount)
	}
}

func TestTestExecutionEnvironment_SnapshotQueries(t *testing.T) {
	env := NewTestExecutionEnvironment("test-epic.xml")

	// Create and load test epic
	testEpic, err := builders.NewEpicBuilder("test-epic").
		WithPhase("1A", "Setup", "pending").
		Build()

	if err != nil {
		t.Fatalf("Failed to build test epic: %v", err)
	}

	err = env.LoadEpic(testEpic)
	if err != nil {
		t.Fatalf("Failed to load epic: %v", err)
	}

	// Save with specific command
	testEpic.Status = epic.StatusActive
	err = env.SaveEpic(testEpic, "start_epic")
	if err != nil {
		t.Fatalf("Failed to save epic: %v", err)
	}

	// Test GetSnapshotByCommand
	snapshot, err := env.GetSnapshotByCommand("start_epic")
	if err != nil {
		t.Fatalf("Failed to get snapshot by command: %v", err)
	}
	if snapshot.EpicState.Status != epic.StatusActive {
		t.Errorf("Expected epic status active in snapshot, got: %s", snapshot.EpicState.Status)
	}

	// Test GetSnapshotByCommand for non-existent command
	_, err = env.GetSnapshotByCommand("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent command")
	}

	// Test GetStateAtStep
	state, err := env.GetStateAtStep(1)
	if err != nil {
		t.Fatalf("Failed to get state at step: %v", err)
	}
	if state.Status != epic.StatusActive {
		t.Errorf("Expected epic status active at step 1, got: %s", state.Status)
	}

	// Test GetStateAtStep with invalid index
	_, err = env.GetStateAtStep(999)
	if err == nil {
		t.Error("Expected error for invalid step index")
	}

	// Test GetLastSnapshot
	lastSnapshot, err := env.GetLastSnapshot()
	if err != nil {
		t.Fatalf("Failed to get last snapshot: %v", err)
	}
	if lastSnapshot.Command != "start_epic" {
		t.Errorf("Expected last snapshot command 'start_epic', got: %s", lastSnapshot.Command)
	}

	// Test GetInitialState
	initialState, err := env.GetInitialState()
	if err != nil {
		t.Fatalf("Failed to get initial state: %v", err)
	}
	if initialState.Status != epic.StatusPending {
		t.Errorf("Expected initial epic status planning, got: %s", initialState.Status)
	}

	// Test GetFinalState
	finalState, err := env.GetFinalState()
	if err != nil {
		t.Fatalf("Failed to get final state: %v", err)
	}
	if finalState.Status != epic.StatusActive {
		t.Errorf("Expected final epic status active, got: %s", finalState.Status)
	}
}

func TestTestExecutionEnvironment_ExecutionSummary(t *testing.T) {
	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	env := NewTestExecutionEnvironment("test-epic.xml").WithTimeSource(func() time.Time { return fixedTime })

	// Create and load test epic
	testEpic, err := builders.NewEpicBuilder("test-epic").
		WithPhase("1A", "Setup", "pending").
		WithTask("1A_1", "1A", "Init", "pending").
		Build()

	if err != nil {
		t.Fatalf("Failed to build test epic: %v", err)
	}

	err = env.LoadEpic(testEpic)
	if err != nil {
		t.Fatalf("Failed to load epic: %v", err)
	}

	// Perform some operations
	testEpic.Status = epic.StatusActive
	err = env.SaveEpic(testEpic, "start_epic")
	if err != nil {
		t.Fatalf("Failed to save epic: %v", err)
	}

	// Finalize execution
	env.Finalize()

	// Get execution summary
	summary, err := env.GetExecutionSummary()
	if err != nil {
		t.Fatalf("Failed to get execution summary: %v", err)
	}

	// Verify summary
	if summary.InitialState.Status != epic.StatusPending {
		t.Errorf("Expected initial state status planning, got: %s", summary.InitialState.Status)
	}

	if summary.FinalState.Status != epic.StatusActive {
		t.Errorf("Expected final state status active, got: %s", summary.FinalState.Status)
	}

	if summary.CommandCount != 2 {
		t.Errorf("Expected 2 commands, got: %d", summary.CommandCount)
	}

	if summary.ErrorCount != 0 {
		t.Errorf("Expected 0 errors, got: %d", summary.ErrorCount)
	}

	if !summary.Success {
		t.Error("Expected summary.Success to be true")
	}

	if len(summary.Snapshots) != 2 {
		t.Errorf("Expected 2 snapshots in summary, got: %d", len(summary.Snapshots))
	}

	// Verify metadata
	if summary.Metadata.TotalCommands != 2 {
		t.Errorf("Expected 2 total commands in metadata, got: %d", summary.Metadata.TotalCommands)
	}

	if summary.Metadata.SuccessCount != 2 {
		t.Errorf("Expected 2 successful commands in metadata, got: %d", summary.Metadata.SuccessCount)
	}

	if summary.ExecutionTime != 0 {
		t.Errorf("Expected execution time to be 0 (same fixed time), got: %v", summary.ExecutionTime)
	}
}

func TestTestExecutionEnvironment_ConcurrentAccess(t *testing.T) {
	env := NewTestExecutionEnvironment("test-epic.xml")

	// Create test epic
	testEpic, err := builders.NewEpicBuilder("test-epic").
		WithPhase("1A", "Setup", "pending").
		Build()

	if err != nil {
		t.Fatalf("Failed to build test epic: %v", err)
	}

	err = env.LoadEpic(testEpic)
	if err != nil {
		t.Fatalf("Failed to load epic: %v", err)
	}

	// Test concurrent access with multiple goroutines
	var wg sync.WaitGroup
	numGoroutines := 10
	commandsPerGoroutine := 5

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < commandsPerGoroutine; j++ {
				// Get current epic (read operation)
				currentEpic, err := env.GetCurrentEpic()
				if err != nil {
					t.Errorf("Goroutine %d: Failed to get current epic: %v", goroutineID, err)
					return
				}

				// Modify epic slightly
				currentEpic.Description = fmt.Sprintf("Modified by goroutine %d, iteration %d", goroutineID, j)

				// Save epic (write operation)
				command := fmt.Sprintf("goroutine_%d_iter_%d", goroutineID, j)
				err = env.SaveEpic(currentEpic, command)
				if err != nil {
					t.Errorf("Goroutine %d: Failed to save epic: %v", goroutineID, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify that all operations completed successfully
	snapshots := env.GetSnapshots()
	expectedSnapshotCount := 1 + (numGoroutines * commandsPerGoroutine) // initial + concurrent operations
	if len(snapshots) != expectedSnapshotCount {
		t.Errorf("Expected %d snapshots, got: %d", expectedSnapshotCount, len(snapshots))
	}

	metadata := env.GetExecutionMetadata()
	if metadata.TotalCommands != expectedSnapshotCount {
		t.Errorf("Expected %d total commands, got: %d", expectedSnapshotCount, metadata.TotalCommands)
	}

	if metadata.ErrorCount != 0 {
		t.Errorf("Expected 0 errors, got: %d", metadata.ErrorCount)
	}
}

func TestTestExecutionEnvironment_MemoryIsolation(t *testing.T) {
	// Create two separate environments to test isolation
	env1 := NewTestExecutionEnvironment("epic1.xml")
	env2 := NewTestExecutionEnvironment("epic2.xml")

	// Create different epics for each environment
	epic1, err := builders.NewEpicBuilder("epic-1").
		WithPhase("1A", "Setup", "pending").
		Build()

	if err != nil {
		t.Fatalf("Failed to build epic1: %v", err)
	}

	epic2, err := builders.NewEpicBuilder("epic-2").
		WithPhase("2A", "Development", "pending").
		Build()

	if err != nil {
		t.Fatalf("Failed to build epic2: %v", err)
	}

	// Load epics into separate environments
	err = env1.LoadEpic(epic1)
	if err != nil {
		t.Fatalf("Failed to load epic1: %v", err)
	}

	err = env2.LoadEpic(epic2)
	if err != nil {
		t.Fatalf("Failed to load epic2: %v", err)
	}

	// Modify epic1
	epic1.Status = epic.StatusActive
	err = env1.SaveEpic(epic1, "start_epic1")
	if err != nil {
		t.Fatalf("Failed to save epic1: %v", err)
	}

	// Verify environments are isolated
	current1, err := env1.GetCurrentEpic()
	if err != nil {
		t.Fatalf("Failed to get current epic from env1: %v", err)
	}

	current2, err := env2.GetCurrentEpic()
	if err != nil {
		t.Fatalf("Failed to get current epic from env2: %v", err)
	}

	// Verify isolation - epic1 should be active, epic2 should still be planning
	if current1.Status != epic.StatusActive {
		t.Errorf("Expected epic1 status to be active, got: %s", current1.Status)
	}

	if current2.Status != epic.StatusPending {
		t.Errorf("Expected epic2 status to be planning, got: %s", current2.Status)
	}

	// Verify different IDs
	if current1.ID != "epic-1" {
		t.Errorf("Expected epic1 ID to be 'epic-1', got: %s", current1.ID)
	}

	if current2.ID != "epic-2" {
		t.Errorf("Expected epic2 ID to be 'epic-2', got: %s", current2.ID)
	}

	// Verify separate snapshots
	snapshots1 := env1.GetSnapshots()
	snapshots2 := env2.GetSnapshots()

	if len(snapshots1) != 2 { // initial_load + start_epic1
		t.Errorf("Expected 2 snapshots in env1, got: %d", len(snapshots1))
	}

	if len(snapshots2) != 1 { // initial_load only
		t.Errorf("Expected 1 snapshot in env2, got: %d", len(snapshots2))
	}
}

func TestTestExecutionEnvironment_Cleanup(t *testing.T) {
	env := NewTestExecutionEnvironment("test-epic.xml")

	// Create and load test epic
	testEpic, err := builders.NewEpicBuilder("test-epic").Build()
	if err != nil {
		t.Fatalf("Failed to build test epic: %v", err)
	}

	err = env.LoadEpic(testEpic)
	if err != nil {
		t.Fatalf("Failed to load epic: %v", err)
	}

	// Perform cleanup
	err = env.Cleanup()
	if err != nil {
		t.Errorf("Cleanup failed: %v", err)
	}

	// For memory storage, cleanup is currently a no-op
	// Verify environment is still functional after cleanup
	current, err := env.GetCurrentEpic()
	if err != nil {
		t.Fatalf("Failed to get current epic after cleanup: %v", err)
	}

	if current.ID != "test-epic" {
		t.Errorf("Expected epic ID 'test-epic' after cleanup, got: %s", current.ID)
	}
}

func TestTestExecutionEnvironment_EmptyState(t *testing.T) {
	env := NewTestExecutionEnvironment("test-epic.xml")

	// Test operations on empty environment
	_, err := env.GetLastSnapshot()
	if err == nil {
		t.Error("Expected error when getting last snapshot from empty environment")
	}

	_, err = env.GetExecutionSummary()
	if err == nil {
		t.Error("Expected error when getting execution summary from empty environment")
	}

	_, err = env.GetStateAtStep(0)
	if err == nil {
		t.Error("Expected error when getting state at step from empty environment")
	}

	// Metadata should still be available
	metadata := env.GetExecutionMetadata()
	if metadata.TotalCommands != 0 {
		t.Errorf("Expected 0 total commands in empty environment, got: %d", metadata.TotalCommands)
	}
}

// Helper function to format errors properly
func formatError(goroutineID int, message string, err error) string {
	return fmt.Sprintf("Goroutine %d: %s: %v", goroutineID, message, err)
}
