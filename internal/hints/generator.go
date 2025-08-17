package hints

import (
	"github.com/mindreframer/agentpm/internal/epic"
)

// HintCategory represents different categories of hints
type HintCategory string

const (
	HintCategoryActionable    HintCategory = "actionable"    // Direct action user can take
	HintCategoryInformational HintCategory = "informational" // Context or explanation
	HintCategoryDiagnostic    HintCategory = "diagnostic"    // Debugging information
	HintCategoryWorkflow      HintCategory = "workflow"      // Process guidance
	HintCategoryConfiguration HintCategory = "configuration" // Setup/config hints
)

// HintPriority represents the priority level of a hint
type HintPriority string

const (
	HintPriorityHigh   HintPriority = "high"   // Critical action needed
	HintPriorityMedium HintPriority = "medium" // Recommended action
	HintPriorityLow    HintPriority = "low"    // Optional suggestion
)

// Hint represents a structured hint with metadata
type Hint struct {
	Content    string       `json:"content"`    // The hint text
	Category   HintCategory `json:"category"`   // Type of hint
	Priority   HintPriority `json:"priority"`   // Priority level
	Command    string       `json:"command"`    // Suggested command (optional)
	Reference  string       `json:"reference"`  // Documentation reference (optional)
	Conditions []string     `json:"conditions"` // When this hint applies
}

// HintContext contains context information for generating hints
type HintContext struct {
	Epic           *epic.Epic  // Current epic
	ActivePhase    *epic.Phase // Currently active phase (if any)
	ActiveTask     *epic.Task  // Currently active task (if any)
	ErrorType      string      // Type of error that occurred
	OperationType  string      // Operation being attempted (start, complete, etc.)
	EntityType     string      // Type of entity (epic, phase, task)
	EntityID       string      // ID of the entity
	CurrentStatus  string      // Current status of the entity
	TargetStatus   string      // Intended target status
	AdditionalData interface{} // Any additional context data
}

// HintGenerator interface for generating context-aware hints
type HintGenerator interface {
	// GenerateHint generates a hint based on the error context
	GenerateHint(ctx *HintContext) *Hint

	// CanHandle checks if this generator can handle the given context
	CanHandle(ctx *HintContext) bool

	// Priority returns the priority of this generator (higher priority = checked first)
	Priority() int
}

// HintRegistry manages multiple hint generators
type HintRegistry struct {
	generators []HintGenerator
	config     *HintRegistryConfig
}

// HintRegistryConfig controls hint generation behavior
type HintRegistryConfig struct {
	Enabled        bool
	ShowCommands   bool
	ShowReferences bool
	MinPriority    HintPriority
	MaxHints       int
	Customizations map[string]string
}

// NewHintRegistry creates a new hint registry with default configuration
func NewHintRegistry() *HintRegistry {
	return &HintRegistry{
		generators: make([]HintGenerator, 0),
		config:     DefaultHintRegistryConfig(),
	}
}

// NewHintRegistryWithConfig creates a new hint registry with custom configuration
func NewHintRegistryWithConfig(config *HintRegistryConfig) *HintRegistry {
	return &HintRegistry{
		generators: make([]HintGenerator, 0),
		config:     config,
	}
}

// DefaultHintRegistryConfig returns default registry configuration
func DefaultHintRegistryConfig() *HintRegistryConfig {
	return &HintRegistryConfig{
		Enabled:        true,
		ShowCommands:   true,
		ShowReferences: false,
		MinPriority:    HintPriorityMedium,
		MaxHints:       3,
		Customizations: make(map[string]string),
	}
}

// Register adds a hint generator to the registry
func (hr *HintRegistry) Register(generator HintGenerator) {
	hr.generators = append(hr.generators, generator)
}

// GenerateHint generates a hint using the first matching generator
func (hr *HintRegistry) GenerateHint(ctx *HintContext) *Hint {
	// Check if hints are disabled
	if !hr.config.Enabled {
		return nil
	}

	// Sort generators by priority (highest first)
	for _, generator := range hr.generators {
		if generator.CanHandle(ctx) {
			hint := generator.GenerateHint(ctx)
			if hint != nil {
				// Apply configuration filtering and customization
				hint = hr.applyConfiguration(hint, ctx)
				if hint != nil && hr.meetsMinimumPriority(hint.Priority) {
					return hint
				}
			}
		}
	}

	// Return default hint if no generator matches
	defaultHint := &Hint{
		Content:  "Check the current state and try again",
		Category: HintCategoryInformational,
		Priority: HintPriorityLow,
	}

	if hr.meetsMinimumPriority(defaultHint.Priority) {
		return hr.applyConfiguration(defaultHint, ctx)
	}

	return nil
}

// applyConfiguration applies registry configuration to a hint
func (hr *HintRegistry) applyConfiguration(hint *Hint, ctx *HintContext) *Hint {
	if hint == nil {
		return nil
	}

	// Apply customizations if available
	if customText, exists := hr.config.Customizations[ctx.ErrorType]; exists {
		hint.Content = customText
	}

	// Remove command if not enabled
	if !hr.config.ShowCommands {
		hint.Command = ""
	}

	// Remove reference if not enabled
	if !hr.config.ShowReferences {
		hint.Reference = ""
	}

	return hint
}

// meetsMinimumPriority checks if hint priority meets minimum threshold
func (hr *HintRegistry) meetsMinimumPriority(priority HintPriority) bool {
	switch hr.config.MinPriority {
	case HintPriorityHigh:
		return priority == HintPriorityHigh
	case HintPriorityMedium:
		return priority == HintPriorityHigh || priority == HintPriorityMedium
	case HintPriorityLow:
		return true // All priorities are shown
	default:
		return true
	}
}

// DefaultHintRegistry creates a registry with default generators
func DefaultHintRegistry() *HintRegistry {
	registry := NewHintRegistry()

	// Register default generators
	registry.Register(&PhaseConstraintHintGenerator{})
	registry.Register(&TaskConstraintHintGenerator{})
	registry.Register(&StateTransitionHintGenerator{})
	registry.Register(&WorkflowHintGenerator{})

	return registry
}

// Placeholder implementations for hint generators (to be expanded)

// PhaseConstraintHintGenerator generates hints for phase constraint violations
type PhaseConstraintHintGenerator struct{}

func (g *PhaseConstraintHintGenerator) CanHandle(ctx *HintContext) bool {
	return ctx.ErrorType == "PhaseConstraintError"
}

func (g *PhaseConstraintHintGenerator) GenerateHint(ctx *HintContext) *Hint {
	return &Hint{
		Content:  "Complete the current active phase before starting a new one",
		Category: HintCategoryActionable,
		Priority: HintPriorityHigh,
		Command:  "agentpm current",
	}
}

func (g *PhaseConstraintHintGenerator) Priority() int { return 100 }

// TaskConstraintHintGenerator generates hints for task constraint violations
type TaskConstraintHintGenerator struct{}

func (g *TaskConstraintHintGenerator) CanHandle(ctx *HintContext) bool {
	return ctx.ErrorType == "TaskConstraintError"
}

func (g *TaskConstraintHintGenerator) GenerateHint(ctx *HintContext) *Hint {
	return &Hint{
		Content:  "Complete the current active task before starting a new one in the same phase",
		Category: HintCategoryActionable,
		Priority: HintPriorityHigh,
		Command:  "agentpm current",
	}
}

func (g *TaskConstraintHintGenerator) Priority() int { return 100 }

// StateTransitionHintGenerator generates hints for invalid state transitions
type StateTransitionHintGenerator struct{}

func (g *StateTransitionHintGenerator) CanHandle(ctx *HintContext) bool {
	return ctx.ErrorType == "TaskStateError" || ctx.ErrorType == "PhaseStateError"
}

func (g *StateTransitionHintGenerator) GenerateHint(ctx *HintContext) *Hint {
	return &Hint{
		Content:  "Check the current status and ensure the entity is in the correct state for this operation",
		Category: HintCategoryDiagnostic,
		Priority: HintPriorityMedium,
		Command:  "agentpm status",
	}
}

func (g *StateTransitionHintGenerator) Priority() int { return 80 }

// WorkflowHintGenerator generates general workflow hints
type WorkflowHintGenerator struct{}

func (g *WorkflowHintGenerator) CanHandle(ctx *HintContext) bool {
	return true // Can handle any context as fallback
}

func (g *WorkflowHintGenerator) GenerateHint(ctx *HintContext) *Hint {
	return &Hint{
		Content:  "Use 'agentpm current' to see active work and 'agentpm status' for an overview",
		Category: HintCategoryWorkflow,
		Priority: HintPriorityLow,
		Command:  "agentpm current",
	}
}

func (g *WorkflowHintGenerator) Priority() int { return 10 }
