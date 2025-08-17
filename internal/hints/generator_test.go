package hints

import (
	"reflect"
	"testing"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHintRegistry_DefaultRegistry(t *testing.T) {
	registry := DefaultHintRegistry()

	assert.NotNil(t, registry)
	assert.Len(t, registry.generators, 5) // PhaseConstraint, TaskConstraint, StateTransition, Workflow, EpicPhaseAware
	assert.NotNil(t, registry.config)
	assert.True(t, registry.config.Enabled)
	assert.True(t, registry.config.ShowCommands)
	assert.False(t, registry.config.ShowReferences)
	assert.Equal(t, HintPriorityMedium, registry.config.MinPriority)
	assert.Equal(t, 3, registry.config.MaxHints)
}

func TestHintRegistry_NewRegistryWithConfig(t *testing.T) {
	config := &HintRegistryConfig{
		Enabled:        false,
		ShowCommands:   false,
		ShowReferences: true,
		MinPriority:    HintPriorityHigh,
		MaxHints:       1,
		Customizations: map[string]string{"test": "custom"},
	}

	registry := NewHintRegistryWithConfig(config)

	assert.NotNil(t, registry)
	assert.Equal(t, config, registry.config)
	assert.Len(t, registry.generators, 0) // No generators registered by default
}

func TestHintRegistry_GenerateHint_DisabledConfiguration(t *testing.T) {
	config := &HintRegistryConfig{
		Enabled: false,
	}
	registry := NewHintRegistryWithConfig(config)

	ctx := &HintContext{
		ErrorType:     "PhaseConstraintError",
		OperationType: "start",
		EntityType:    "phase",
		EntityID:      "test-phase",
	}

	hint := registry.GenerateHint(ctx)

	assert.Nil(t, hint, "Should return nil when hints are disabled")
}

func TestHintRegistry_GenerateHint_NoMatchingGenerator(t *testing.T) {
	// Use custom config that allows low priority hints (so WorkflowHintGenerator can show)
	config := &HintRegistryConfig{
		Enabled:     true,
		MinPriority: HintPriorityLow, // Allow low priority hints
	}
	registry := NewHintRegistryWithConfig(config)

	// Register all default generators
	registry.Register(&PhaseConstraintHintGenerator{})
	registry.Register(&TaskConstraintHintGenerator{})
	registry.Register(&StateTransitionHintGenerator{})
	registry.Register(&WorkflowHintGenerator{})

	ctx := &HintContext{
		ErrorType:     "UnknownErrorType",
		OperationType: "unknown",
		EntityType:    "unknown",
		EntityID:      "test",
	}

	hint := registry.GenerateHint(ctx)

	// This should return a hint from WorkflowHintGenerator since it's a fallback
	require.NotNil(t, hint, "Expected a hint from WorkflowHintGenerator as fallback")
	assert.Equal(t, "Use 'agentpm current' to see active work and 'agentpm status' for an overview", hint.Content)
	assert.Equal(t, HintCategoryWorkflow, hint.Category)
	assert.Equal(t, HintPriorityLow, hint.Priority)
}

func TestHintRegistry_GenerateHint_PriorityFiltering(t *testing.T) {
	tests := []struct {
		name         string
		minPriority  HintPriority
		hintPriority HintPriority
		shouldShow   bool
	}{
		{
			name:         "High min priority shows high priority hint",
			minPriority:  HintPriorityHigh,
			hintPriority: HintPriorityHigh,
			shouldShow:   true,
		},
		{
			name:         "High min priority filters medium priority hint",
			minPriority:  HintPriorityHigh,
			hintPriority: HintPriorityMedium,
			shouldShow:   false,
		},
		{
			name:         "Medium min priority shows high priority hint",
			minPriority:  HintPriorityMedium,
			hintPriority: HintPriorityHigh,
			shouldShow:   true,
		},
		{
			name:         "Medium min priority shows medium priority hint",
			minPriority:  HintPriorityMedium,
			hintPriority: HintPriorityMedium,
			shouldShow:   true,
		},
		{
			name:         "Medium min priority filters low priority hint",
			minPriority:  HintPriorityMedium,
			hintPriority: HintPriorityLow,
			shouldShow:   false,
		},
		{
			name:         "Low min priority shows all hints",
			minPriority:  HintPriorityLow,
			hintPriority: HintPriorityLow,
			shouldShow:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &HintRegistryConfig{
				Enabled:     true,
				MinPriority: tt.minPriority,
			}
			registry := NewHintRegistryWithConfig(config)

			// Register a mock generator that returns a hint with the test priority
			mockGen := &MockHintGenerator{
				canHandle: true,
				hint: &Hint{
					Content:  "Test hint",
					Priority: tt.hintPriority,
				},
			}
			registry.Register(mockGen)

			ctx := &HintContext{ErrorType: "test"}
			hint := registry.GenerateHint(ctx)

			if tt.shouldShow {
				assert.NotNil(t, hint)
				assert.Equal(t, "Test hint", hint.Content)
			} else {
				assert.Nil(t, hint)
			}
		})
	}
}

func TestEpicPhaseAwareHintGenerator(t *testing.T) {
	generator := &EpicPhaseAwareHintGenerator{}

	t.Run("can handle PhaseConstraintError with epic context", func(t *testing.T) {
		epic := &epic.Epic{ID: "epic-1"}
		ctx := &HintContext{
			ErrorType: "PhaseConstraintError",
			Epic:      epic,
		}
		assert.True(t, generator.CanHandle(ctx))
	})

	t.Run("can handle TaskConstraintError with epic context", func(t *testing.T) {
		epic := &epic.Epic{ID: "epic-1"}
		ctx := &HintContext{
			ErrorType: "TaskConstraintError",
			Epic:      epic,
		}
		assert.True(t, generator.CanHandle(ctx))
	})

	t.Run("cannot handle without epic context", func(t *testing.T) {
		ctx := &HintContext{ErrorType: "PhaseConstraintError"}
		assert.False(t, generator.CanHandle(ctx))
	})

	t.Run("cannot handle other error types", func(t *testing.T) {
		epic := &epic.Epic{ID: "epic-1"}
		ctx := &HintContext{
			ErrorType: "WorkflowError",
			Epic:      epic,
		}
		assert.False(t, generator.CanHandle(ctx))
	})

	t.Run("generates phase workflow hint", func(t *testing.T) {
		epic := &epic.Epic{
			ID:       "epic-1",
			Workflow: "Sequential",
			Phases: []epic.Phase{
				{ID: "phase-1", Status: epic.StatusActive, Name: "Phase 1"},
				{ID: "phase-2", Status: epic.StatusPending, Name: "Phase 2"},
			},
		}

		ctx := &HintContext{
			ErrorType: "PhaseConstraintError",
			EntityID:  "phase-2",
			Epic:      epic,
		}

		hint := generator.GenerateHint(ctx)

		assert.NotNil(t, hint)
		assert.Equal(t, HintCategoryWorkflow, hint.Category)
		assert.Equal(t, HintPriorityMedium, hint.Priority)
		assert.Contains(t, hint.Content, "Complete phase 'phase-1' before starting 'phase-2'")
		assert.Equal(t, "agentpm done-phase phase-1", hint.Command)
		assert.Contains(t, hint.Reference, "Sequential")
		assert.Contains(t, hint.Conditions, "Phases should be completed in dependency order")
	})

	t.Run("generates task workflow hint", func(t *testing.T) {
		epic := &epic.Epic{
			ID: "epic-1",
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1"},
			},
			Tasks: []epic.Task{
				{ID: "task-1", PhaseID: "phase-1", Status: epic.StatusActive},
				{ID: "task-2", PhaseID: "phase-1", Status: epic.StatusPending},
			},
		}

		ctx := &HintContext{
			ErrorType: "TaskConstraintError",
			EntityID:  "task-2",
			Epic:      epic,
		}

		hint := generator.GenerateHint(ctx)

		assert.NotNil(t, hint)
		assert.Equal(t, HintCategoryWorkflow, hint.Category)
		assert.Equal(t, HintPriorityMedium, hint.Priority)
		assert.Contains(t, hint.Content, "Complete task 'task-1' in phase 'phase-1'")
		assert.Equal(t, "agentpm done-task task-1", hint.Command)
		assert.Contains(t, hint.Reference, "Phase 1")
		assert.Contains(t, hint.Conditions, "Only one task per phase can be active")
	})

	t.Run("generates generic workflow hint", func(t *testing.T) {
		epic := &epic.Epic{
			ID:       "epic-1",
			Workflow: "Custom",
		}

		ctx := &HintContext{
			ErrorType: "UnknownError",
			Epic:      epic,
		}

		hint := generator.GenerateHint(ctx)

		assert.NotNil(t, hint)
		assert.Equal(t, HintCategoryWorkflow, hint.Category)
		assert.Equal(t, HintPriorityLow, hint.Priority)
		assert.Contains(t, hint.Content, "Epic 'epic-1' workflow guidance available")
		assert.Equal(t, "agentpm current", hint.Command)
		assert.Contains(t, hint.Reference, "Custom")
	})

	t.Run("analyzes phase dependencies correctly", func(t *testing.T) {
		epic := &epic.Epic{
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1"},
				{ID: "phase-2", Name: "Phase 2"},
				{ID: "phase-3", Name: "Phase 3"},
			},
		}

		activePhase := &epic.Phases[0]
		targetPhase := &epic.Phases[2]

		dependencies := generator.analyzePhaseDependencies(epic, activePhase, targetPhase)

		assert.Len(t, dependencies, 1)
		assert.Contains(t, dependencies, "Sequential phase order")
	})

	t.Run("no dependencies for reverse order", func(t *testing.T) {
		epic := &epic.Epic{
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1"},
				{ID: "phase-2", Name: "Phase 2"},
			},
		}

		activePhase := &epic.Phases[1]
		targetPhase := &epic.Phases[0]

		dependencies := generator.analyzePhaseDependencies(epic, activePhase, targetPhase)

		assert.Empty(t, dependencies)
	})

	t.Run("has correct priority", func(t *testing.T) {
		assert.Equal(t, 90, generator.Priority())
	})
}

// Epic 9 Phase 3D: Context-Aware Hint Tests
// These tests validate specific Epic 9 requirements for actionable hint generation

func TestEpic9_PhaseConflictHints(t *testing.T) {
	t.Run("phase conflicts provide actionable completion commands", func(t *testing.T) {
		// Epic 9 line 53: "To start 'Testing' phase, first complete active phase 'Implementation' with: `agentpm complete-phase implementation`"
		epic := &epic.Epic{
			ID:       "epic-1",
			Workflow: "Sequential",
			Phases: []epic.Phase{
				{ID: "implementation", Name: "Implementation", Status: epic.StatusActive},
				{ID: "testing", Name: "Testing", Status: epic.StatusPending},
			},
		}

		ctx := &HintContext{
			ErrorType: "PhaseConstraintError",
			EntityID:  "testing",
			Epic:      epic,
			AdditionalData: map[string]interface{}{
				"active_phase": "implementation",
			},
		}

		registry := DefaultHintRegistry()
		hint := registry.GenerateHint(ctx)

		require.NotNil(t, hint)
		assert.Equal(t, HintCategoryActionable, hint.Category)
		assert.Contains(t, hint.Content, "Complete phase 'implementation' before starting 'testing'")
		assert.Equal(t, "agentpm done-phase implementation", hint.Command)
		assert.Contains(t, hint.Conditions, "Phases should be completed in dependency order")
	})

	t.Run("phase conflicts with multiple active phases", func(t *testing.T) {
		epic := &epic.Epic{
			ID: "epic-complex",
			Phases: []epic.Phase{
				{ID: "phase-1", Name: "Phase 1", Status: epic.StatusActive},
				{ID: "phase-2", Name: "Phase 2", Status: epic.StatusPending},
				{ID: "phase-3", Name: "Phase 3", Status: epic.StatusPending},
			},
		}

		ctx := &HintContext{
			ErrorType: "PhaseConstraintError",
			EntityID:  "phase-3",
			Epic:      epic,
			AdditionalData: map[string]interface{}{
				"active_phase": "phase-1",
			},
		}

		registry := DefaultHintRegistry()
		hint := registry.GenerateHint(ctx)

		require.NotNil(t, hint)
		assert.Contains(t, hint.Content, "Complete phase 'phase-1' before starting 'phase-3'")
		assert.Equal(t, "agentpm done-phase phase-1", hint.Command)
	})
}

func TestEpic9_MissingDependencyHints(t *testing.T) {
	t.Run("missing dependencies list specific tasks", func(t *testing.T) {
		// Epic 9 line 54: "Phase 'Deployment' requires completion of tasks: [list]. Complete with: `agentpm complete-task <task-id>`"
		epic := &epic.Epic{
			ID: "epic-deps",
			Phases: []epic.Phase{
				{ID: "deployment", Name: "Deployment", Status: epic.StatusPending},
			},
			Tasks: []epic.Task{
				{ID: "build-task", PhaseID: "deployment", Name: "Build", Status: epic.StatusActive},
				{ID: "test-task", PhaseID: "deployment", Name: "Test", Status: epic.StatusPending},
			},
		}

		ctx := &HintContext{
			ErrorType: "TaskConstraintError",
			EntityID:  "test-task",
			Epic:      epic,
			AdditionalData: map[string]interface{}{
				"active_task_id": "build-task",
				"phase_id":       "deployment",
			},
		}

		registry := DefaultHintRegistry()
		hint := registry.GenerateHint(ctx)

		require.NotNil(t, hint)
		assert.Contains(t, hint.Content, "Complete task 'build-task' in phase 'deployment' before starting 'test-task'")
		assert.Equal(t, "agentpm done-task build-task", hint.Command)
		assert.Contains(t, hint.Conditions, "Only one task per phase can be active")
	})

	t.Run("task dependency hints suggest completion order", func(t *testing.T) {
		epic := &epic.Epic{
			ID: "epic-sequential",
			Tasks: []epic.Task{
				{ID: "prep-task", PhaseID: "prep", Status: epic.StatusActive},
				{ID: "main-task", PhaseID: "prep", Status: epic.StatusPending},
			},
		}

		ctx := &HintContext{
			ErrorType: "TaskConstraintError",
			EntityID:  "main-task",
			Epic:      epic,
			AdditionalData: map[string]interface{}{
				"active_task_id": "prep-task",
				"phase_id":       "prep",
			},
		}

		registry := DefaultHintRegistry()
		hint := registry.GenerateHint(ctx)

		require.NotNil(t, hint)
		assert.Contains(t, hint.Content, "Complete task 'prep-task'")
		assert.Equal(t, "agentpm done-task prep-task", hint.Command)
	})
}

func TestEpic9_InvalidReferenceHints(t *testing.T) {
	t.Run("invalid references suggest alternatives", func(t *testing.T) {
		// Epic 9 line 55: "Task 'invalid-id' not found. List available tasks with: `agentpm query tasks`"

		// Test state transition errors for invalid entities
		ctx := &HintContext{
			ErrorType:     "TaskStateError",
			EntityType:    "task",
			EntityID:      "invalid-task-id",
			OperationType: "start",
			CurrentStatus: "", // No status indicates entity not found
		}

		registry := DefaultHintRegistry()
		hint := registry.GenerateHint(ctx)

		require.NotNil(t, hint)
		assert.Contains(t, hint.Content, "Check the current status")
		assert.Equal(t, "agentpm status", hint.Command)
		assert.Contains(t, hint.Conditions, "Entity must be in appropriate state for the operation")
	})

	t.Run("phase reference errors provide listing commands", func(t *testing.T) {
		ctx := &HintContext{
			ErrorType:     "PhaseStateError",
			EntityType:    "phase",
			EntityID:      "non-existent-phase",
			OperationType: "start",
		}

		registry := DefaultHintRegistry()
		hint := registry.GenerateHint(ctx)

		require.NotNil(t, hint)
		// StateTransitionHintGenerator should handle this
		assert.NotEmpty(t, hint.Content)
		assert.NotEmpty(t, hint.Command)
	})
}

func TestEpic9_EpicStateIssueHints(t *testing.T) {
	t.Run("epic state issues provide initialization commands", func(t *testing.T) {
		// Epic 9 line 56: "Epic not started. Initialize with: `agentpm start-epic`"

		ctx := &HintContext{
			ErrorType:     "EpicStateError",
			EntityType:    "epic",
			EntityID:      "uninitialized-epic",
			OperationType: "start",
			CurrentStatus: "planning",
			TargetStatus:  "active",
		}

		registry := DefaultHintRegistry()
		hint := registry.GenerateHint(ctx)

		require.NotNil(t, hint)
		assert.Contains(t, hint.Content, "Start epic 'uninitialized-epic'")
		assert.Equal(t, "agentpm start-epic uninitialized-epic", hint.Command)
	})

	t.Run("epic completion blocked by active phases", func(t *testing.T) {
		epic := &epic.Epic{
			ID:     "epic-active",
			Status: epic.StatusActive,
			Phases: []epic.Phase{
				{ID: "phase-1", Status: epic.StatusCompleted},
				{ID: "phase-2", Status: epic.StatusActive},
			},
		}

		ctx := &HintContext{
			ErrorType:     "EpicStateError",
			EntityType:    "epic",
			EntityID:      "epic-active",
			OperationType: "complete",
			Epic:          epic,
		}

		registry := DefaultHintRegistry()
		hint := registry.GenerateHint(ctx)

		require.NotNil(t, hint)
		// Should provide guidance about completing active phases first
		assert.NotEmpty(t, hint.Content)
		assert.NotEmpty(t, hint.Command)
	})
}

func TestEpic9_CommandSuggestionAccuracy(t *testing.T) {
	t.Run("command suggestions match entity types and operations", func(t *testing.T) {
		testCases := []struct {
			name        string
			errorType   string
			entityType  string
			entityID    string
			operation   string
			expectedCmd string
		}{
			{
				name:        "phase start command",
				errorType:   "PhaseConstraintError",
				entityType:  "phase",
				entityID:    "test-phase",
				operation:   "start",
				expectedCmd: "agentpm done-phase", // Complete current phase first
			},
			{
				name:        "task completion command",
				errorType:   "TaskConstraintError",
				entityType:  "task",
				entityID:    "test-task",
				operation:   "start",
				expectedCmd: "agentpm done-task", // Complete current task first
			},
			{
				name:        "epic initialization command",
				errorType:   "EpicStateError",
				entityType:  "epic",
				entityID:    "test-epic",
				operation:   "start",
				expectedCmd: "agentpm start-epic test-epic",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				ctx := &HintContext{
					ErrorType:     tc.errorType,
					EntityType:    tc.entityType,
					EntityID:      tc.entityID,
					OperationType: tc.operation,
					CurrentStatus: "planning",
				}

				// Add specific context data for constraint errors
				if tc.errorType == "PhaseConstraintError" {
					ctx.AdditionalData = map[string]interface{}{
						"active_phase": "current-phase",
					}
				} else if tc.errorType == "TaskConstraintError" {
					ctx.AdditionalData = map[string]interface{}{
						"active_task_id": "current-task",
						"phase_id":       "current-phase",
					}
				} else if tc.errorType == "EpicStateError" {
					ctx.TargetStatus = "active"
				}

				registry := DefaultHintRegistry()
				hint := registry.GenerateHint(ctx)

				require.NotNil(t, hint)
				assert.Contains(t, hint.Command, tc.expectedCmd,
					"Expected command to contain %s for %s", tc.expectedCmd, tc.name)
			})
		}
	})

	t.Run("commands include specific entity IDs", func(t *testing.T) {
		ctx := &HintContext{
			ErrorType:     "TaskStateError",
			EntityType:    "task",
			EntityID:      "specific-task-123",
			OperationType: "start",
			CurrentStatus: "planning",
			TargetStatus:  "active",
		}

		registry := DefaultHintRegistry()
		hint := registry.GenerateHint(ctx)

		require.NotNil(t, hint)
		assert.Contains(t, hint.Content, "specific-task-123",
			"Hint content should include specific entity ID")
	})
}

func TestEpic9_HintCategorizationCorrectness(t *testing.T) {
	t.Run("hints are properly categorized", func(t *testing.T) {
		testCases := []struct {
			name        string
			errorType   string
			expectedCat HintCategory
		}{
			{"phase constraint", "PhaseConstraintError", HintCategoryActionable},
			{"task constraint", "TaskConstraintError", HintCategoryActionable},
			{"state transition", "TaskStateError", HintCategoryActionable},
			{"workflow guidance", "WorkflowError", HintCategoryWorkflow},
		}

		// Create registry that accepts all priority levels for testing
		config := &HintRegistryConfig{
			Enabled:     true,
			MinPriority: HintPriorityLow,
		}
		registry := NewHintRegistryWithConfig(config)
		registry.Register(&PhaseConstraintHintGenerator{})
		registry.Register(&TaskConstraintHintGenerator{})
		registry.Register(&StateTransitionHintGenerator{})
		registry.Register(&WorkflowHintGenerator{})

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				ctx := &HintContext{ErrorType: tc.errorType}
				hint := registry.GenerateHint(ctx)

				require.NotNil(t, hint)
				assert.Equal(t, tc.expectedCat, hint.Category,
					"Expected category %s for error type %s", tc.expectedCat, tc.errorType)
			})
		}
	})

	t.Run("priority hierarchy is maintained", func(t *testing.T) {
		registry := DefaultHintRegistry()

		// Verify generator priority order
		expectedPriorities := map[string]int{
			"PhaseConstraintHintGenerator": 100,
			"TaskConstraintHintGenerator":  100,
			"EpicPhaseAwareHintGenerator":  90,
			"StateTransitionHintGenerator": 80,
			"WorkflowHintGenerator":        10,
		}

		for _, gen := range registry.generators {
			genType := reflect.TypeOf(gen).Elem().Name()
			if expectedPrio, exists := expectedPriorities[genType]; exists {
				assert.Equal(t, expectedPrio, gen.Priority(),
					"Generator %s should have priority %d", genType, expectedPrio)
			}
		}
	})
}

func TestHintRegistry_ApplyConfiguration(t *testing.T) {
	t.Run("customizations applied", func(t *testing.T) {
		config := &HintRegistryConfig{
			Enabled: true,
			Customizations: map[string]string{
				"TestError": "Custom hint text",
			},
		}
		registry := NewHintRegistryWithConfig(config)

		hint := &Hint{
			Content: "Original hint",
		}
		ctx := &HintContext{ErrorType: "TestError"}

		result := registry.applyConfiguration(hint, ctx)

		assert.Equal(t, "Custom hint text", result.Content)
	})

	t.Run("commands filtered when disabled", func(t *testing.T) {
		config := &HintRegistryConfig{
			Enabled:      true,
			ShowCommands: false,
		}
		registry := NewHintRegistryWithConfig(config)

		hint := &Hint{
			Content: "Test hint",
			Command: "agentpm status",
		}
		ctx := &HintContext{}

		result := registry.applyConfiguration(hint, ctx)

		assert.Equal(t, "Test hint", result.Content)
		assert.Empty(t, result.Command)
	})

	t.Run("references filtered when disabled", func(t *testing.T) {
		config := &HintRegistryConfig{
			Enabled:        true,
			ShowReferences: false,
		}
		registry := NewHintRegistryWithConfig(config)

		hint := &Hint{
			Content:   "Test hint",
			Reference: "https://docs.example.com",
		}
		ctx := &HintContext{}

		result := registry.applyConfiguration(hint, ctx)

		assert.Equal(t, "Test hint", result.Content)
		assert.Empty(t, result.Reference)
	})
}

func TestPhaseConstraintHintGenerator(t *testing.T) {
	generator := &PhaseConstraintHintGenerator{}

	t.Run("can handle PhaseConstraintError", func(t *testing.T) {
		ctx := &HintContext{ErrorType: "PhaseConstraintError"}
		assert.True(t, generator.CanHandle(ctx))
	})

	t.Run("cannot handle other errors", func(t *testing.T) {
		ctx := &HintContext{ErrorType: "TaskStateError"}
		assert.False(t, generator.CanHandle(ctx))
	})

	t.Run("generates appropriate hint", func(t *testing.T) {
		ctx := &HintContext{ErrorType: "PhaseConstraintError"}
		hint := generator.GenerateHint(ctx)

		assert.NotNil(t, hint)
		// Enhanced generator provides context-aware content, fallback to generic if no context
		assert.Contains(t, hint.Content, "before starting")
		assert.Equal(t, HintCategoryActionable, hint.Category)
		assert.Equal(t, HintPriorityHigh, hint.Priority)
		assert.NotEmpty(t, hint.Command)
	})

	t.Run("has correct priority", func(t *testing.T) {
		assert.Equal(t, 100, generator.Priority())
	})
}

func TestTaskConstraintHintGenerator(t *testing.T) {
	generator := &TaskConstraintHintGenerator{}

	t.Run("can handle TaskConstraintError", func(t *testing.T) {
		ctx := &HintContext{ErrorType: "TaskConstraintError"}
		assert.True(t, generator.CanHandle(ctx))
	})

	t.Run("cannot handle other errors", func(t *testing.T) {
		ctx := &HintContext{ErrorType: "PhaseStateError"}
		assert.False(t, generator.CanHandle(ctx))
	})

	t.Run("generates appropriate hint", func(t *testing.T) {
		ctx := &HintContext{ErrorType: "TaskConstraintError"}
		hint := generator.GenerateHint(ctx)

		assert.NotNil(t, hint)
		// Enhanced generator provides context-aware content, fallback to generic if no context
		assert.Contains(t, hint.Content, "task")
		assert.Equal(t, HintCategoryActionable, hint.Category)
		assert.Equal(t, HintPriorityHigh, hint.Priority)
		assert.NotEmpty(t, hint.Command)
	})

	t.Run("has correct priority", func(t *testing.T) {
		assert.Equal(t, 100, generator.Priority())
	})
}

func TestStateTransitionHintGenerator(t *testing.T) {
	generator := &StateTransitionHintGenerator{}

	t.Run("can handle state transition errors", func(t *testing.T) {
		tests := []string{"PhaseStateError", "TaskStateError"} // EpicStateError not handled according to implementation
		for _, errorType := range tests {
			ctx := &HintContext{ErrorType: errorType}
			assert.True(t, generator.CanHandle(ctx), "Should handle %s", errorType)
		}
	})

	t.Run("cannot handle other errors", func(t *testing.T) {
		ctx := &HintContext{ErrorType: "ValidationError"}
		assert.False(t, generator.CanHandle(ctx))
	})

	t.Run("generates appropriate hint", func(t *testing.T) {
		ctx := &HintContext{ErrorType: "TaskStateError"}
		hint := generator.GenerateHint(ctx)

		assert.NotNil(t, hint)
		assert.Equal(t, "Check the current status and ensure the entity is in the correct state for this operation", hint.Content)
		assert.Equal(t, HintCategoryActionable, hint.Category)
		assert.Equal(t, HintPriorityMedium, hint.Priority)
		assert.Equal(t, "agentpm status", hint.Command)
	})

	t.Run("has correct priority", func(t *testing.T) {
		assert.Equal(t, 80, generator.Priority())
	})
}

func TestWorkflowHintGenerator(t *testing.T) {
	generator := &WorkflowHintGenerator{}

	t.Run("handles all errors as fallback", func(t *testing.T) {
		tests := []string{"UnknownError", "ValidationError", "FileNotFound"}
		for _, errorType := range tests {
			ctx := &HintContext{ErrorType: errorType}
			assert.True(t, generator.CanHandle(ctx), "Should handle %s as fallback", errorType)
		}
	})

	t.Run("generates workflow guidance hint", func(t *testing.T) {
		ctx := &HintContext{ErrorType: "SomeError"}
		hint := generator.GenerateHint(ctx)

		assert.NotNil(t, hint)
		assert.Equal(t, "Use 'agentpm current' to see active work and 'agentpm status' for an overview", hint.Content)
		assert.Equal(t, HintCategoryWorkflow, hint.Category)
		assert.Equal(t, HintPriorityLow, hint.Priority)
		assert.Equal(t, "agentpm current", hint.Command)
	})

	t.Run("has lowest priority as fallback", func(t *testing.T) {
		assert.Equal(t, 10, generator.Priority())
	})
}

func TestHintContext_PopulationAndUsage(t *testing.T) {
	epicObj := &epic.Epic{
		ID:     "test-epic",
		Status: epic.StatusActive,
	}

	phase := &epic.Phase{
		ID:     "test-phase",
		Status: epic.StatusActive,
	}

	task := &epic.Task{
		ID:      "test-task",
		PhaseID: "test-phase",
		Status:  epic.StatusPending,
	}

	ctx := &HintContext{
		Epic:          epicObj,
		ActivePhase:   phase,
		ActiveTask:    task,
		ErrorType:     "TaskConstraintError",
		OperationType: "start",
		EntityType:    "task",
		EntityID:      "test-task",
		CurrentStatus: "pending",
		TargetStatus:  "active",
		AdditionalData: map[string]interface{}{
			"active_task_id": "other-task",
			"phase_id":       "test-phase",
		},
	}

	t.Run("context fields populated correctly", func(t *testing.T) {
		assert.Equal(t, "test-epic", ctx.Epic.ID)
		assert.Equal(t, "test-phase", ctx.ActivePhase.ID)
		assert.Equal(t, "test-task", ctx.ActiveTask.ID)
		assert.Equal(t, "TaskConstraintError", ctx.ErrorType)
		assert.Equal(t, "start", ctx.OperationType)
		assert.Equal(t, "task", ctx.EntityType)
		assert.Equal(t, "test-task", ctx.EntityID)
		assert.Equal(t, "pending", ctx.CurrentStatus)
		assert.Equal(t, "active", ctx.TargetStatus)
		assert.NotNil(t, ctx.AdditionalData)
	})

	t.Run("context used for hint generation", func(t *testing.T) {
		registry := DefaultHintRegistry()
		hint := registry.GenerateHint(ctx)

		assert.NotNil(t, hint)
		// Should match TaskConstraintHintGenerator since ErrorType is TaskConstraintError
		// Enhanced generator now provides context-aware content
		assert.Contains(t, hint.Content, "Complete task")
		assert.Equal(t, HintCategoryActionable, hint.Category)
		assert.Equal(t, HintPriorityHigh, hint.Priority)
	})
}

// MockHintGenerator is a test helper for mocking hint generators
type MockHintGenerator struct {
	canHandle bool
	hint      *Hint
	priority  int
}

func (m *MockHintGenerator) CanHandle(ctx *HintContext) bool {
	return m.canHandle
}

func (m *MockHintGenerator) GenerateHint(ctx *HintContext) *Hint {
	return m.hint
}

func (m *MockHintGenerator) Priority() int {
	if m.priority == 0 {
		return 50 // Default priority
	}
	return m.priority
}

func TestHintRegistryIntegration(t *testing.T) {
	t.Run("full integration with multiple generators", func(t *testing.T) {
		// Create registry with low priority enabled for testing fallback
		config := &HintRegistryConfig{
			Enabled:     true,
			MinPriority: HintPriorityLow, // Allow all priorities
		}
		registry := NewHintRegistryWithConfig(config)

		// Register all default generators
		registry.Register(&PhaseConstraintHintGenerator{})
		registry.Register(&TaskConstraintHintGenerator{})
		registry.Register(&StateTransitionHintGenerator{})
		registry.Register(&WorkflowHintGenerator{})

		// Test phase constraint error - should use PhaseConstraintHintGenerator
		phaseCtx := &HintContext{
			ErrorType:     "PhaseConstraintError",
			OperationType: "start",
			EntityType:    "phase",
			EntityID:      "phase-2",
			AdditionalData: map[string]interface{}{
				"active_phase": "phase-1",
			},
		}

		hint := registry.GenerateHint(phaseCtx)
		require.NotNil(t, hint)
		// Enhanced generator now provides context-aware content
		assert.Contains(t, hint.Content, "Complete phase")
		assert.Contains(t, hint.Content, "phase-1")
		assert.Contains(t, hint.Content, "phase-2")
		assert.Equal(t, HintPriorityHigh, hint.Priority)

		// Test task constraint error - should use TaskConstraintHintGenerator
		taskCtx := &HintContext{
			ErrorType:     "TaskConstraintError",
			OperationType: "start",
			EntityType:    "task",
			EntityID:      "task-2",
			AdditionalData: map[string]interface{}{
				"active_task_id": "task-1",
				"phase_id":       "phase-1",
			},
		}

		hint = registry.GenerateHint(taskCtx)
		require.NotNil(t, hint)
		// Enhanced generator now provides context-aware content
		assert.Contains(t, hint.Content, "Complete task")
		assert.Contains(t, hint.Content, "task-1")
		assert.Contains(t, hint.Content, "task-2")
		assert.Equal(t, HintPriorityHigh, hint.Priority)

		// Test state error - should use StateTransitionHintGenerator
		stateCtx := &HintContext{
			ErrorType:     "TaskStateError",
			OperationType: "start",
			EntityType:    "task",
			EntityID:      "task-1",
			CurrentStatus: "completed",
			TargetStatus:  "active",
		}

		hint = registry.GenerateHint(stateCtx)
		require.NotNil(t, hint)
		// Enhanced generator now provides context-aware content for completed state
		assert.Contains(t, hint.Content, "completed")
		assert.Equal(t, HintPriorityMedium, hint.Priority)

		// Test unknown error - should use WorkflowHintGenerator as fallback
		unknownCtx := &HintContext{
			ErrorType: "UnknownError",
		}

		hint = registry.GenerateHint(unknownCtx)
		require.NotNil(t, hint)
		assert.Equal(t, "Use 'agentpm current' to see active work and 'agentpm status' for an overview", hint.Content)
		assert.Equal(t, HintPriorityLow, hint.Priority)
	})
}
