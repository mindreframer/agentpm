package hints

import (
	"testing"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHintRegistry_DefaultRegistry(t *testing.T) {
	registry := DefaultHintRegistry()

	assert.NotNil(t, registry)
	assert.Len(t, registry.generators, 4) // PhaseConstraint, TaskConstraint, StateTransition, Workflow
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
		assert.Equal(t, "Complete the current active phase before starting a new one", hint.Content)
		assert.Equal(t, HintCategoryActionable, hint.Category)
		assert.Equal(t, HintPriorityHigh, hint.Priority)
		assert.Equal(t, "agentpm current", hint.Command)
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
		assert.Equal(t, "Complete the current active task before starting a new one in the same phase", hint.Content)
		assert.Equal(t, HintCategoryActionable, hint.Category)
		assert.Equal(t, HintPriorityHigh, hint.Priority)
		assert.Equal(t, "agentpm current", hint.Command)
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
		assert.Equal(t, HintCategoryDiagnostic, hint.Category)
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
		assert.Equal(t, "Complete the current active task before starting a new one in the same phase", hint.Content)
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
		assert.Equal(t, "Complete the current active phase before starting a new one", hint.Content)
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
		assert.Equal(t, "Complete the current active task before starting a new one in the same phase", hint.Content)
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
		assert.Equal(t, "Check the current status and ensure the entity is in the correct state for this operation", hint.Content)
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
