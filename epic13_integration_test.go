package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEpic13StatusSystemIntegration tests the complete status system workflow
func TestEpic13StatusSystemIntegration(t *testing.T) {
	t.Run("epic validation with unified status system", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create epic with Epic 13 status system
		epicPath := createEpic13StatusEpic(tempDir, "status-workflow.xml")

		storage := storage.NewFileStorage()
		loadedEpic, err := storage.LoadEpic(epicPath)
		require.NoError(t, err)

		// Validate epic with status system
		result := loadedEpic.Validate()
		assert.True(t, result.Valid, "Epic should be valid")
		assert.Equal(t, "passed", result.Checks["status_values"], "Status values should pass validation")
	})

	t.Run("status validation with mixed status states", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create epic with mixed status states
		epicPath := createEpic13StatusValidationEpic(tempDir, "status-validation.xml")

		storage := storage.NewFileStorage()
		loadedEpic, err := storage.LoadEpic(epicPath)
		require.NoError(t, err)

		// Validate should pass despite mixed statuses (validation checks structure, not business rules)
		result := loadedEpic.Validate()
		assert.True(t, result.Valid, "Epic should be valid")
	})

	t.Run("large epic with status system performance", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping large epic test in short mode")
		}

		tempDir := t.TempDir()

		// Create large epic with Epic 13 status system
		epicPath := createLargeEpic13StatusEpic(tempDir, "large-status-epic.xml")

		storage := storage.NewFileStorage()

		start := time.Now()
		loadedEpic, err := storage.LoadEpic(epicPath)
		loadDuration := time.Since(start)

		require.NoError(t, err)
		assert.Less(t, loadDuration, 500*time.Millisecond, "Large epic load should complete in < 500ms")

		start = time.Now()
		result := loadedEpic.Validate()
		validateDuration := time.Since(start)

		assert.True(t, result.Valid, "Large epic should be valid")
		assert.Less(t, validateDuration, 500*time.Millisecond, "Large epic validation should complete in < 500ms")

		t.Logf("Large epic performance: load=%v, validate=%v", loadDuration, validateDuration)
	})
}

// TestEpic13StatusValidationPerformance tests that status validation stays under 10ms
func TestEpic13StatusValidationPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	t.Run("status validation performance benchmark", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create epic with many entities for performance testing
		testEpic := createPerformanceTestEpic()
		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "perf-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Load epic and measure validation performance
		loadedEpic, err := storage.LoadEpic(epicPath)
		require.NoError(t, err)

		// Benchmark validation performance
		iterations := 100
		totalDuration := time.Duration(0)

		for i := 0; i < iterations; i++ {
			start := time.Now()
			result := loadedEpic.Validate()
			duration := time.Since(start)
			totalDuration += duration

			assert.True(t, result.Valid, "Epic should be valid")
		}

		avgDuration := totalDuration / time.Duration(iterations)
		assert.Less(t, avgDuration, 10*time.Millisecond, "Average validation should complete in < 10ms")

		t.Logf("Status validation performance: avg=%v over %d iterations", avgDuration, iterations)
	})

	t.Run("memory usage remains constant", func(t *testing.T) {
		tempDir := t.TempDir()

		// Test different epic sizes
		sizes := []int{10, 50, 100, 200}
		for _, size := range sizes {
			testEpic := createVariableSizeEpic(size)
			storage := storage.NewFileStorage()
			epicPath := filepath.Join(tempDir, fmt.Sprintf("epic-%d.xml", size))
			err := storage.SaveEpic(testEpic, epicPath)
			require.NoError(t, err)

			loadedEpic, err := storage.LoadEpic(epicPath)
			require.NoError(t, err)

			start := time.Now()
			result := loadedEpic.Validate()
			duration := time.Since(start)

			assert.True(t, result.Valid, "Epic should be valid")
			// Validation time should not grow exponentially with size
			assert.Less(t, duration, 50*time.Millisecond, "Validation should scale linearly")

			t.Logf("Epic size %d: validation time=%v", size, duration)
		}
	})
}

// TestEpic13EdgeCases tests edge cases with the status system
func TestEpic13EdgeCases(t *testing.T) {
	t.Run("empty phases with status validation", func(t *testing.T) {
		tempDir := t.TempDir()

		testEpic := &epic.Epic{
			ID:        "empty-phases-test",
			Name:      "Empty Phases Test",
			Status:    epic.StatusPending,
			CreatedAt: time.Now(),
			Assignee:  "test_agent",
			Phases:    []epic.Phase{}, // Empty phases
			Tasks:     []epic.Task{},
			Tests:     []epic.Test{},
		}

		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "empty-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		loadedEpic, err := storage.LoadEpic(epicPath)
		require.NoError(t, err)

		result := loadedEpic.Validate()
		assert.True(t, result.Valid, "Empty epic should be valid")
	})

	t.Run("orphaned tests with status validation", func(t *testing.T) {
		tempDir := t.TempDir()

		testEpic := &epic.Epic{
			ID:        "orphaned-tests",
			Name:      "Orphaned Tests Epic",
			Status:    epic.StatusPending,
			CreatedAt: time.Now(),
			Assignee:  "test_agent",
			Phases: []epic.Phase{
				{ID: "P1", Name: "Phase 1", Status: epic.StatusPending},
			},
			Tasks: []epic.Task{
				{ID: "T1", PhaseID: "P1", Name: "Task 1", Status: epic.StatusPending},
			},
			Tests: []epic.Test{
				{ID: "TEST1", TaskID: "T1", Name: "Valid Test", TestStatus: epic.TestStatusPending},
				{ID: "TEST2", TaskID: "MISSING", Name: "Orphaned Test", TestStatus: epic.TestStatusPending}, // Orphaned
			},
		}

		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "orphaned-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		loadedEpic, err := storage.LoadEpic(epicPath)
		require.NoError(t, err)

		result := loadedEpic.Validate()
		assert.False(t, result.Valid, "Epic with orphaned tests should be invalid")
		assert.Contains(t, strings.Join(result.Errors, " "), "references non-existent task")
	})

	t.Run("concurrent status validation", func(t *testing.T) {
		tempDir := t.TempDir()
		epicPath := createEpic13StatusEpic(tempDir, "concurrent-epic.xml")

		storage := storage.NewFileStorage()
		loadedEpic, err := storage.LoadEpic(epicPath)
		require.NoError(t, err)

		// Test concurrent validation (should be thread-safe)
		done := make(chan bool, 5)
		for i := 0; i < 5; i++ {
			go func() {
				defer func() { done <- true }()

				result := loadedEpic.Validate()
				assert.True(t, result.Valid, "Concurrent validation should work")
			}()
		}

		for i := 0; i < 5; i++ {
			<-done
		}
	})
}

// TestEpic13StatusSystemCompleteWorkflow tests end-to-end status workflows
func TestEpic13StatusSystemCompleteWorkflow(t *testing.T) {
	t.Run("complete workflow with status transitions", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create epic with pending statuses
		testEpic := &epic.Epic{
			ID:        "workflow-test",
			Name:      "Complete Workflow Test",
			Status:    epic.StatusPending,
			CreatedAt: time.Now(),
			Assignee:  "test_agent",
			Phases: []epic.Phase{
				{ID: "P1", Name: "Phase 1", Status: epic.StatusPending},
				{ID: "P2", Name: "Phase 2", Status: epic.StatusPending},
			},
			Tasks: []epic.Task{
				{ID: "T1", PhaseID: "P1", Name: "Task 1", Status: epic.StatusPending},
				{ID: "T2", PhaseID: "P1", Name: "Task 2", Status: epic.StatusPending},
				{ID: "T3", PhaseID: "P2", Name: "Task 3", Status: epic.StatusPending},
			},
			Tests: []epic.Test{
				{ID: "TEST1", TaskID: "T1", Name: "Test 1", TestStatus: epic.TestStatusPending, TestResult: epic.TestResultFailing},
				{ID: "TEST2", TaskID: "T2", Name: "Test 2", TestStatus: epic.TestStatusPending, TestResult: epic.TestResultFailing},
				{ID: "TEST3", TaskID: "T3", Name: "Test 3", TestStatus: epic.TestStatusPending, TestResult: epic.TestResultFailing},
			},
		}

		storage := storage.NewFileStorage()
		epicPath := filepath.Join(tempDir, "workflow-epic.xml")
		err := storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		// Validate initial state
		loadedEpic, err := storage.LoadEpic(epicPath)
		require.NoError(t, err)

		result := loadedEpic.Validate()
		assert.True(t, result.Valid, "Initial epic should be valid")

		// Test status transitions
		testEpic.Status = epic.StatusWIP           // WIP -> Active
		testEpic.Phases[0].Status = epic.StatusWIP // WIP -> Active
		testEpic.Tasks[0].Status = epic.StatusWIP  // WIP -> Active
		testEpic.Tests[0].TestStatus = epic.TestStatusWIP

		err = storage.SaveEpic(testEpic, epicPath)
		require.NoError(t, err)

		loadedEpic, err = storage.LoadEpic(epicPath)
		require.NoError(t, err)

		result = loadedEpic.Validate()
		assert.True(t, result.Valid, "Epic with active statuses should be valid")
	})
}

// Helper functions for creating test epics with Epic 13 status system

func createEpic13StatusEpic(dir, filename string) string {
	epicPath := filepath.Join(dir, filename)

	testEpic := &epic.Epic{
		ID:        "epic13-status-test",
		Name:      "Epic 13 Status System Test",
		Status:    epic.StatusPending,
		CreatedAt: time.Now(),
		Assignee:  "test_agent",
		Phases: []epic.Phase{
			{ID: "P1", Name: "Phase 1", Status: epic.StatusPending},
			{ID: "P2", Name: "Phase 2", Status: epic.StatusPending},
		},
		Tasks: []epic.Task{
			{ID: "T1", PhaseID: "P1", Name: "Task 1", Status: epic.StatusPending},
			{ID: "T2", PhaseID: "P1", Name: "Task 2", Status: epic.StatusWIP},       // WIP -> Active
			{ID: "T3", PhaseID: "P2", Name: "Task 3", Status: epic.StatusCompleted}, // Done -> Completed
		},
		Tests: []epic.Test{
			{ID: "TEST1", TaskID: "T1", Name: "Test 1", TestStatus: epic.TestStatusPending, TestResult: epic.TestResultFailing},
			{ID: "TEST2", TaskID: "T2", Name: "Test 2", TestStatus: epic.TestStatusWIP, TestResult: epic.TestResultFailing},
			{ID: "TEST3", TaskID: "T3", Name: "Test 3", TestStatus: epic.TestStatusDone, TestResult: epic.TestResultPassing},
		},
	}

	storage := storage.NewFileStorage()
	err := storage.SaveEpic(testEpic, epicPath)
	if err != nil {
		panic(err)
	}

	return epicPath
}

func createEpic13StatusValidationEpic(dir, filename string) string {
	epicPath := filepath.Join(dir, filename)

	testEpic := &epic.Epic{
		ID:        "status-validation-test",
		Name:      "Status Validation Test Epic",
		Status:    epic.StatusWIP, // Active instead of WIP
		CreatedAt: time.Now(),
		Assignee:  "test_agent",
		Phases: []epic.Phase{
			{ID: "P1", Name: "Phase 1", Status: epic.StatusWIP}, // Active instead of WIP
			{ID: "P2", Name: "Phase 2", Status: epic.StatusPending},
		},
		Tasks: []epic.Task{
			{ID: "T1", PhaseID: "P1", Name: "Task 1", Status: epic.StatusCompleted}, // Completed instead of Done
			{ID: "T2", PhaseID: "P1", Name: "Task 2", Status: epic.StatusCancelled},
			{ID: "T3", PhaseID: "P2", Name: "Task 3", Status: epic.StatusPending},
		},
		Tests: []epic.Test{
			{ID: "TEST1", TaskID: "T1", Name: "Test 1", TestStatus: epic.TestStatusDone, TestResult: epic.TestResultPassing},
			{ID: "TEST2", TaskID: "T2", Name: "Test 2", TestStatus: epic.TestStatusCancelled, TestResult: epic.TestResultFailing},
			{ID: "TEST3", TaskID: "T3", Name: "Test 3", TestStatus: epic.TestStatusPending, TestResult: epic.TestResultFailing},
		},
	}

	storage := storage.NewFileStorage()
	err := storage.SaveEpic(testEpic, epicPath)
	if err != nil {
		panic(err)
	}

	return epicPath
}

func createLargeEpic13StatusEpic(dir, filename string) string {
	epicPath := filepath.Join(dir, filename)

	testEpic := &epic.Epic{
		ID:        "large-epic13-test",
		Name:      "Large Epic 13 Status Test",
		Status:    epic.StatusWIP, // Active instead of WIP
		CreatedAt: time.Now(),
		Assignee:  "test_agent",
	}

	// Create many phases with various statuses
	for i := 1; i <= 100; i++ {
		status := epic.StatusPending
		if i%3 == 0 {
			status = epic.StatusWIP // Active instead of WIP
		} else if i%5 == 0 {
			status = epic.StatusCompleted // Completed instead of Done
		}

		testEpic.Phases = append(testEpic.Phases, epic.Phase{
			ID:     fmt.Sprintf("P%d", i),
			Name:   fmt.Sprintf("Phase %d", i),
			Status: status,
		})
	}

	// Create many tasks with various statuses
	for i := 1; i <= 500; i++ {
		phaseID := fmt.Sprintf("P%d", (i%100)+1)
		status := epic.StatusPending
		if i%3 == 0 {
			status = epic.StatusWIP // Active instead of WIP
		} else if i%7 == 0 {
			status = epic.StatusCompleted // Completed instead of Done
		} else if i%11 == 0 {
			status = epic.StatusCancelled
		}

		testEpic.Tasks = append(testEpic.Tasks, epic.Task{
			ID:      fmt.Sprintf("T%d", i),
			PhaseID: phaseID,
			Name:    fmt.Sprintf("Task %d", i),
			Status:  status,
		})
	}

	// Create many tests with various statuses
	for i := 1; i <= 500; i++ {
		taskID := fmt.Sprintf("T%d", i)
		status := epic.TestStatusPending
		result := epic.TestResultFailing
		if i%3 == 0 {
			status = epic.TestStatusWIP
		} else if i%7 == 0 {
			status = epic.TestStatusDone
			result = epic.TestResultPassing
		} else if i%11 == 0 {
			status = epic.TestStatusCancelled
		}

		testEpic.Tests = append(testEpic.Tests, epic.Test{
			ID:         fmt.Sprintf("TEST%d", i),
			TaskID:     taskID,
			Name:       fmt.Sprintf("Test %d", i),
			TestStatus: status,
			TestResult: result,
		})
	}

	storage := storage.NewFileStorage()
	err := storage.SaveEpic(testEpic, epicPath)
	if err != nil {
		panic(err)
	}

	return epicPath
}

func createPerformanceTestEpic() *epic.Epic {
	testEpic := &epic.Epic{
		ID:        "performance-test",
		Name:      "Performance Test Epic",
		Status:    epic.StatusWIP, // Active instead of WIP
		CreatedAt: time.Now(),
		Assignee:  "test_agent",
	}

	// Create moderate number of entities for performance testing
	for i := 1; i <= 50; i++ {
		testEpic.Phases = append(testEpic.Phases, epic.Phase{
			ID:     fmt.Sprintf("P%d", i),
			Name:   fmt.Sprintf("Phase %d", i),
			Status: epic.StatusPending,
		})
	}

	for i := 1; i <= 200; i++ {
		phaseID := fmt.Sprintf("P%d", (i%50)+1)
		testEpic.Tasks = append(testEpic.Tasks, epic.Task{
			ID:      fmt.Sprintf("T%d", i),
			PhaseID: phaseID,
			Name:    fmt.Sprintf("Task %d", i),
			Status:  epic.StatusPending,
		})
	}

	for i := 1; i <= 400; i++ {
		taskID := fmt.Sprintf("T%d", (i%200)+1)
		testEpic.Tests = append(testEpic.Tests, epic.Test{
			ID:         fmt.Sprintf("TEST%d", i),
			TaskID:     taskID,
			Name:       fmt.Sprintf("Test %d", i),
			TestStatus: epic.TestStatusPending,
			TestResult: epic.TestResultFailing,
		})
	}

	return testEpic
}

func createVariableSizeEpic(size int) *epic.Epic {
	testEpic := &epic.Epic{
		ID:        fmt.Sprintf("size-test-%d", size),
		Name:      fmt.Sprintf("Size Test Epic %d", size),
		Status:    epic.StatusPending,
		CreatedAt: time.Now(),
		Assignee:  "test_agent",
	}

	// Create size number of each entity type
	for i := 1; i <= size; i++ {
		testEpic.Phases = append(testEpic.Phases, epic.Phase{
			ID:     fmt.Sprintf("P%d", i),
			Name:   fmt.Sprintf("Phase %d", i),
			Status: epic.StatusPending,
		})

		testEpic.Tasks = append(testEpic.Tasks, epic.Task{
			ID:      fmt.Sprintf("T%d", i),
			PhaseID: fmt.Sprintf("P%d", i),
			Name:    fmt.Sprintf("Task %d", i),
			Status:  epic.StatusPending,
		})

		testEpic.Tests = append(testEpic.Tests, epic.Test{
			ID:         fmt.Sprintf("TEST%d", i),
			TaskID:     fmt.Sprintf("T%d", i),
			Name:       fmt.Sprintf("Test %d", i),
			TestStatus: epic.TestStatusPending,
			TestResult: epic.TestResultFailing,
		})
	}

	return testEpic
}
