package executor

import (
	"fmt"
	"sync"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/storage"
)

// TestExecutionEnvironment provides isolated storage for test execution
type TestExecutionEnvironment struct {
	storage    storage.Storage
	epicFile   string
	mu         sync.RWMutex
	snapshots  []StateSnapshot
	metadata   ExecutionMetadata
	timeSource func() time.Time
}

// StateSnapshot captures the state of an epic at a specific point in time
type StateSnapshot struct {
	Command   string
	Timestamp time.Time
	EpicState *epic.Epic
	Success   bool
	Error     error
	Metadata  map[string]interface{}
}

// ExecutionMetadata tracks execution information
type ExecutionMetadata struct {
	StartTime     time.Time
	EndTime       *time.Time
	TotalCommands int
	SuccessCount  int
	ErrorCount    int
	MemoryUsage   int64
}

// NewTestExecutionEnvironment creates a new isolated test environment
func NewTestExecutionEnvironment(epicFile string) *TestExecutionEnvironment {
	return &TestExecutionEnvironment{
		storage:    storage.NewMemoryStorage(),
		epicFile:   epicFile,
		snapshots:  make([]StateSnapshot, 0),
		metadata:   ExecutionMetadata{StartTime: time.Now()},
		timeSource: time.Now,
	}
}

// WithTimeSource allows injection of custom time source for deterministic testing
func (env *TestExecutionEnvironment) WithTimeSource(timeSource func() time.Time) *TestExecutionEnvironment {
	env.timeSource = timeSource
	// Update start time to match the time source for consistency in tests
	env.metadata.StartTime = timeSource()
	return env
}

// LoadEpic loads an epic into the isolated environment
func (env *TestExecutionEnvironment) LoadEpic(e *epic.Epic) error {
	env.mu.Lock()
	defer env.mu.Unlock()

	if e == nil {
		return fmt.Errorf("epic cannot be nil")
	}

	// Store the epic in isolated memory storage
	err := env.storage.SaveEpic(e, env.epicFile)
	if err != nil {
		return fmt.Errorf("failed to load epic into test environment: %w", err)
	}

	// Take initial snapshot
	env.addSnapshot("initial_load", e, true, nil)
	env.metadata.SuccessCount++

	return nil
}

// GetCurrentEpic returns the current state of the epic
func (env *TestExecutionEnvironment) GetCurrentEpic() (*epic.Epic, error) {
	env.mu.RLock()
	defer env.mu.RUnlock()

	return env.storage.LoadEpic(env.epicFile)
}

// SaveEpic saves the epic state and takes a snapshot
func (env *TestExecutionEnvironment) SaveEpic(e *epic.Epic, command string) error {
	env.mu.Lock()
	defer env.mu.Unlock()

	err := env.storage.SaveEpic(e, env.epicFile)
	if err != nil {
		env.addSnapshot(command, e, false, err)
		env.metadata.ErrorCount++
		return err
	}

	env.addSnapshot(command, e, true, nil)
	env.metadata.SuccessCount++
	return nil
}

// GetStorage returns the underlying storage (for command service integration)
func (env *TestExecutionEnvironment) GetStorage() storage.Storage {
	return env.storage
}

// GetEpicFile returns the epic file path being used
func (env *TestExecutionEnvironment) GetEpicFile() string {
	return env.epicFile
}

// GetSnapshots returns all state snapshots taken during execution
func (env *TestExecutionEnvironment) GetSnapshots() []StateSnapshot {
	env.mu.RLock()
	defer env.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]StateSnapshot, len(env.snapshots))
	copy(result, env.snapshots)
	return result
}

// GetExecutionMetadata returns execution metadata
func (env *TestExecutionEnvironment) GetExecutionMetadata() ExecutionMetadata {
	env.mu.RLock()
	defer env.mu.RUnlock()

	// Create a copy with current end time if not set
	metadata := env.metadata
	if metadata.EndTime == nil {
		now := env.timeSource()
		metadata.EndTime = &now
	}

	return metadata
}

// Finalize marks the execution as complete
func (env *TestExecutionEnvironment) Finalize() {
	env.mu.Lock()
	defer env.mu.Unlock()

	if env.metadata.EndTime == nil {
		now := env.timeSource()
		env.metadata.EndTime = &now
	}
}

// Cleanup performs cleanup of resources (currently no-op for memory storage)
func (env *TestExecutionEnvironment) Cleanup() error {
	env.mu.Lock()
	defer env.mu.Unlock()

	// For memory storage, cleanup is automatic via GC
	// In future implementations with persistent storage, this would clean up temp files
	return nil
}

// addSnapshot adds a state snapshot (must be called with mutex held)
func (env *TestExecutionEnvironment) addSnapshot(command string, e *epic.Epic, success bool, err error) {
	// Create a deep copy of the epic to prevent mutation
	epicCopy := *e
	if e.Phases != nil {
		epicCopy.Phases = make([]epic.Phase, len(e.Phases))
		copy(epicCopy.Phases, e.Phases)
	}
	if e.Tasks != nil {
		epicCopy.Tasks = make([]epic.Task, len(e.Tasks))
		copy(epicCopy.Tasks, e.Tasks)
	}
	if e.Tests != nil {
		epicCopy.Tests = make([]epic.Test, len(e.Tests))
		copy(epicCopy.Tests, e.Tests)
	}
	if e.Events != nil {
		epicCopy.Events = make([]epic.Event, len(e.Events))
		copy(epicCopy.Events, e.Events)
	}

	snapshot := StateSnapshot{
		Command:   command,
		Timestamp: env.timeSource(),
		EpicState: &epicCopy,
		Success:   success,
		Error:     err,
		Metadata:  make(map[string]interface{}),
	}

	// Add basic metadata
	snapshot.Metadata["epic_id"] = e.ID
	snapshot.Metadata["epic_status"] = string(e.Status)
	snapshot.Metadata["phase_count"] = len(e.Phases)
	snapshot.Metadata["task_count"] = len(e.Tasks)
	snapshot.Metadata["test_count"] = len(e.Tests)
	snapshot.Metadata["event_count"] = len(e.Events)

	env.snapshots = append(env.snapshots, snapshot)
	env.metadata.TotalCommands++
}

// GetStateAtStep returns the epic state after a specific command step
func (env *TestExecutionEnvironment) GetStateAtStep(stepIndex int) (*epic.Epic, error) {
	env.mu.RLock()
	defer env.mu.RUnlock()

	if stepIndex < 0 || stepIndex >= len(env.snapshots) {
		return nil, fmt.Errorf("step index %d out of range (0-%d)", stepIndex, len(env.snapshots)-1)
	}

	return env.snapshots[stepIndex].EpicState, nil
}

// GetSnapshotByCommand returns the first snapshot matching the given command
func (env *TestExecutionEnvironment) GetSnapshotByCommand(command string) (*StateSnapshot, error) {
	env.mu.RLock()
	defer env.mu.RUnlock()

	for i := range env.snapshots {
		if env.snapshots[i].Command == command {
			return &env.snapshots[i], nil
		}
	}

	return nil, fmt.Errorf("no snapshot found for command: %s", command)
}

// GetLastSnapshot returns the most recent snapshot
func (env *TestExecutionEnvironment) GetLastSnapshot() (*StateSnapshot, error) {
	env.mu.RLock()
	defer env.mu.RUnlock()

	if len(env.snapshots) == 0 {
		return nil, fmt.Errorf("no snapshots available")
	}

	lastIndex := len(env.snapshots) - 1
	return &env.snapshots[lastIndex], nil
}

// GetInitialState returns the initial loaded state
func (env *TestExecutionEnvironment) GetInitialState() (*epic.Epic, error) {
	return env.GetStateAtStep(0)
}

// GetFinalState returns the final state after all commands
func (env *TestExecutionEnvironment) GetFinalState() (*epic.Epic, error) {
	env.mu.RLock()
	defer env.mu.RUnlock()

	if len(env.snapshots) == 0 {
		return nil, fmt.Errorf("no snapshots available")
	}

	return env.snapshots[len(env.snapshots)-1].EpicState, nil
}

// ExecutionSummary provides a summary of the execution
type ExecutionSummary struct {
	Environment   *TestExecutionEnvironment
	InitialState  *epic.Epic
	FinalState    *epic.Epic
	Snapshots     []StateSnapshot
	Metadata      ExecutionMetadata
	Success       bool
	ErrorCount    int
	CommandCount  int
	ExecutionTime time.Duration
}

// GetExecutionSummary returns a comprehensive summary of the execution
func (env *TestExecutionEnvironment) GetExecutionSummary() (*ExecutionSummary, error) {
	env.mu.RLock()
	defer env.mu.RUnlock()

	if len(env.snapshots) == 0 {
		return nil, fmt.Errorf("no execution data available")
	}

	initialState := env.snapshots[0].EpicState
	finalState := env.snapshots[len(env.snapshots)-1].EpicState

	metadata := env.metadata
	if metadata.EndTime == nil {
		now := env.timeSource()
		metadata.EndTime = &now
	}

	var executionTime time.Duration
	if metadata.EndTime != nil {
		executionTime = metadata.EndTime.Sub(metadata.StartTime)
	}

	// Create a copy of snapshots
	snapshots := make([]StateSnapshot, len(env.snapshots))
	copy(snapshots, env.snapshots)

	return &ExecutionSummary{
		Environment:   env,
		InitialState:  initialState,
		FinalState:    finalState,
		Snapshots:     snapshots,
		Metadata:      metadata,
		Success:       metadata.ErrorCount == 0,
		ErrorCount:    metadata.ErrorCount,
		CommandCount:  metadata.TotalCommands,
		ExecutionTime: executionTime,
	}, nil
}
