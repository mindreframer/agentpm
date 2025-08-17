package hints

import (
	"fmt"

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
	registry.Register(&EpicPhaseAwareHintGenerator{})

	return registry
}

// Placeholder implementations for hint generators (to be expanded)

// PhaseConstraintHintGenerator generates hints for phase constraint violations
type PhaseConstraintHintGenerator struct{}

func (g *PhaseConstraintHintGenerator) CanHandle(ctx *HintContext) bool {
	return ctx.ErrorType == "PhaseConstraintError"
}

func (g *PhaseConstraintHintGenerator) GenerateHint(ctx *HintContext) *Hint {
	hint := &Hint{
		Category: HintCategoryActionable,
		Priority: HintPriorityHigh,
	}

	// Extract active phase information from context
	var activePhaseID string
	if ctx.ActivePhase != nil {
		activePhaseID = ctx.ActivePhase.ID
	} else if ctx.AdditionalData != nil {
		if id, ok := ctx.AdditionalData.(map[string]interface{})["active_phase"]; ok {
			activePhaseID = fmt.Sprintf("%v", id)
		}
	}

	// Generate context-aware content with Epic 9 specificity
	if activePhaseID != "" && ctx.EntityID != "" {
		hint.Content = fmt.Sprintf("Complete phase '%s' before starting '%s'", activePhaseID, ctx.EntityID)
		hint.Command = fmt.Sprintf("agentpm done-phase %s", activePhaseID)
		hint.Conditions = []string{
			"Phases should be completed in dependency order",
			"Check phase prerequisites before starting",
		}
	} else if activePhaseID != "" {
		hint.Content = fmt.Sprintf("Complete the current active phase '%s' before starting a new one", activePhaseID)
		hint.Command = fmt.Sprintf("agentpm done-phase %s", activePhaseID)
		hint.Conditions = []string{
			"Only one phase can be active at a time",
			"Complete active phase before starting another",
		}
	} else {
		hint.Content = "Complete the current active phase before starting a new one"
		hint.Command = "agentpm current"
		hint.Conditions = []string{
			"Only one phase can be active at a time",
			"Complete active phase before starting another",
		}
	}

	return hint
}

func (g *PhaseConstraintHintGenerator) Priority() int { return 100 }

// TaskConstraintHintGenerator generates hints for task constraint violations
type TaskConstraintHintGenerator struct{}

func (g *TaskConstraintHintGenerator) CanHandle(ctx *HintContext) bool {
	return ctx.ErrorType == "TaskConstraintError"
}

func (g *TaskConstraintHintGenerator) GenerateHint(ctx *HintContext) *Hint {
	hint := &Hint{
		Category: HintCategoryActionable,
		Priority: HintPriorityHigh,
	}

	// Extract task and phase information from context
	var activeTaskID, phaseID string
	if ctx.ActiveTask != nil {
		activeTaskID = ctx.ActiveTask.ID
		phaseID = ctx.ActiveTask.PhaseID
	} else if ctx.AdditionalData != nil {
		if data, ok := ctx.AdditionalData.(map[string]interface{}); ok {
			if id, exists := data["active_task_id"]; exists {
				activeTaskID = fmt.Sprintf("%v", id)
			}
			if id, exists := data["phase_id"]; exists {
				phaseID = fmt.Sprintf("%v", id)
			}
		}
	}

	// Generate context-aware content with Epic 9 specificity
	if activeTaskID != "" && phaseID != "" && ctx.EntityID != "" {
		hint.Content = fmt.Sprintf("Complete task '%s' in phase '%s' before starting '%s'", activeTaskID, phaseID, ctx.EntityID)
		hint.Command = fmt.Sprintf("agentpm done-task %s", activeTaskID)
		hint.Conditions = []string{
			"Only one task per phase can be active",
			"Complete current task before starting another",
		}
	} else if activeTaskID != "" && ctx.EntityID != "" {
		hint.Content = fmt.Sprintf("Complete task '%s' before starting task '%s'", activeTaskID, ctx.EntityID)
		hint.Command = fmt.Sprintf("agentpm done-task %s", activeTaskID)
		hint.Conditions = []string{
			"Only one task per phase can be active",
			"Complete current task before starting another",
		}
	} else if activeTaskID != "" {
		hint.Content = fmt.Sprintf("Complete task '%s' before starting a new one", activeTaskID)
		hint.Command = fmt.Sprintf("agentpm done-task %s", activeTaskID)
		hint.Conditions = []string{
			"Only one task per phase can be active",
			"Complete current task before starting another",
		}
	} else {
		hint.Content = "Complete the current active task before starting a new one in the same phase"
		hint.Command = "agentpm current"
		hint.Conditions = []string{
			"Only one task can be active per phase at a time",
			"Complete or cancel active task before starting another",
		}
	}

	return hint
}

func (g *TaskConstraintHintGenerator) Priority() int { return 100 }

// StateTransitionHintGenerator generates hints for invalid state transitions
type StateTransitionHintGenerator struct{}

func (g *StateTransitionHintGenerator) CanHandle(ctx *HintContext) bool {
	return ctx.ErrorType == "TaskStateError" || ctx.ErrorType == "PhaseStateError" || ctx.ErrorType == "EpicStateError"
}

func (g *StateTransitionHintGenerator) GenerateHint(ctx *HintContext) *Hint {
	hint := &Hint{
		Category: HintCategoryActionable,
		Priority: HintPriorityMedium,
	}

	// Extract state information from context
	currentStatus := ctx.CurrentStatus
	targetStatus := ctx.TargetStatus
	entityType := ctx.EntityType
	entityID := ctx.EntityID
	operation := ctx.OperationType

	// Generate context-aware content based on state transition
	if currentStatus != "" && targetStatus != "" && entityType != "" {
		hint.Content = fmt.Sprintf("Cannot %s %s '%s': current status is '%s' but '%s' status is required",
			operation, entityType, entityID, currentStatus, getRequiredStatus(operation, entityType))

		// Provide specific next steps based on current state
		switch currentStatus {
		case "completed":
			if operation == "start" {
				hint.Content = fmt.Sprintf("%s '%s' is already completed. Use 'agentpm status' to see available work", entityType, entityID)
				hint.Command = "agentpm status"
			}
		case "planning":
			if operation == "complete" || operation == "done" {
				hint.Content = fmt.Sprintf("Start %s '%s' before marking it complete", entityType, entityID)
				hint.Command = fmt.Sprintf("agentpm start-%s %s", entityType, entityID)
			} else if operation == "start" && entityType == "epic" {
				hint.Content = fmt.Sprintf("Start %s '%s' before performing other operations", entityType, entityID)
				hint.Command = fmt.Sprintf("agentpm start-%s %s", entityType, entityID)
			}
		case "active":
			if operation == "start" {
				hint.Content = fmt.Sprintf("%s '%s' is already active. Use 'agentpm current' to see active work", entityType, entityID)
				hint.Command = "agentpm current"
			}
		default:
			hint.Content = fmt.Sprintf("Check the current status of %s '%s' and ensure it's in the correct state for %s operation", entityType, entityID, operation)
			hint.Command = "agentpm status"
		}
	} else {
		hint.Content = "Check the current status and ensure the entity is in the correct state for this operation"
		hint.Command = "agentpm status"
	}

	// Add state transition conditions
	hint.Conditions = []string{
		"Entity must be in appropriate state for the operation",
		"Check current status before attempting transitions",
	}

	return hint
}

func (g *StateTransitionHintGenerator) Priority() int { return 80 }

// getRequiredStatus returns the required status for a given operation
func getRequiredStatus(operation, entityType string) string {
	switch operation {
	case "start":
		return "pending"
	case "complete", "done":
		return "active"
	case "cancel":
		return "active"
	default:
		return "appropriate"
	}
}

// WorkflowHintGenerator generates general workflow hints
type WorkflowHintGenerator struct{}

func (g *WorkflowHintGenerator) CanHandle(ctx *HintContext) bool {
	return true // Can handle any context as fallback
}

func (g *WorkflowHintGenerator) GenerateHint(ctx *HintContext) *Hint {
	hint := &Hint{
		Category: HintCategoryWorkflow,
		Priority: HintPriorityLow,
	}

	// Generate workflow hints based on epic and current state
	if ctx.Epic != nil {
		epicStatus := string(ctx.Epic.Status)

		switch epicStatus {
		case "planning":
			hint.Content = "Epic is in planning phase. Start the first phase to begin work"
			hint.Command = "agentpm start-phase"
		case "active":
			if ctx.ActivePhase != nil {
				hint.Content = fmt.Sprintf("Continue work in active phase '%s'. Use 'agentpm current' to see current task", ctx.ActivePhase.ID)
				hint.Command = "agentpm current"
			} else {
				hint.Content = "Epic is active but no phase is started. Start a phase to begin work"
				hint.Command = "agentpm start-next"
			}
		case "completed":
			hint.Content = "Epic is completed. Use 'agentpm switch' to work on a different epic"
			hint.Command = "agentpm switch"
		case "on_hold":
			hint.Content = "Epic is on hold. Resume work by reactivating phases and tasks"
			hint.Command = "agentpm status"
		default:
			hint.Content = "Use 'agentpm status' to see epic overview and 'agentpm current' for active work"
			hint.Command = "agentpm status"
		}
	} else {
		// No epic context available
		hint.Content = "Use 'agentpm current' to see active work and 'agentpm status' for an overview"
		hint.Command = "agentpm current"
	}

	// Add workflow guidance conditions
	hint.Conditions = []string{
		"Follow the epic workflow: planning → active phases → completed",
		"Use 'agentpm current' to see what to work on next",
	}

	return hint
}

func (g *WorkflowHintGenerator) Priority() int { return 10 }

// EpicPhaseAwareHintGenerator generates hints based on epic workflow and phase relationships
type EpicPhaseAwareHintGenerator struct{}

func (g *EpicPhaseAwareHintGenerator) CanHandle(ctx *HintContext) bool {
	// Handle specific epic and phase-related scenarios
	return ctx.Epic != nil && (ctx.ErrorType == "PhaseConstraintError" || ctx.ErrorType == "TaskConstraintError")
}

func (g *EpicPhaseAwareHintGenerator) GenerateHint(ctx *HintContext) *Hint {
	hint := &Hint{
		Category: HintCategoryWorkflow,
		Priority: HintPriorityMedium,
	}

	if ctx.Epic == nil {
		return hint
	}

	epic := ctx.Epic

	// Analyze epic workflow and provide phase-aware guidance
	switch ctx.ErrorType {
	case "PhaseConstraintError":
		hint = g.generatePhaseWorkflowHint(ctx, epic)
	case "TaskConstraintError":
		hint = g.generateTaskWorkflowHint(ctx, epic)
	default:
		hint = g.generateGenericWorkflowHint(ctx, epic)
	}

	return hint
}

func (g *EpicPhaseAwareHintGenerator) generatePhaseWorkflowHint(ctx *HintContext, epicData *epic.Epic) *Hint {
	hint := &Hint{
		Category: HintCategoryWorkflow,
		Priority: HintPriorityMedium,
	}

	// Analyze phase dependencies and workflow
	targetPhaseID := ctx.EntityID
	var activePhase *epic.Phase
	var targetPhase *epic.Phase

	// Find phases
	for i := range epicData.Phases {
		if epicData.Phases[i].Status == epic.StatusActive {
			activePhase = &epicData.Phases[i]
		}
		if epicData.Phases[i].ID == targetPhaseID {
			targetPhase = &epicData.Phases[i]
		}
	}

	if activePhase != nil && targetPhase != nil {
		// Check if phases have dependencies
		dependencies := g.analyzePhaseDependencies(epicData, activePhase, targetPhase)

		if len(dependencies) > 0 {
			hint.Content = fmt.Sprintf("Complete phase '%s' before starting '%s'. Dependencies: %s",
				activePhase.ID, targetPhase.ID, dependencies[0])
			hint.Command = fmt.Sprintf("agentpm done-phase %s", activePhase.ID)
			hint.Conditions = []string{
				"Phases should be completed in dependency order",
				"Check phase prerequisites before starting",
			}
		} else {
			hint.Content = fmt.Sprintf("Complete phase '%s' before starting '%s'", activePhase.ID, targetPhase.ID)
			hint.Command = fmt.Sprintf("agentpm done-phase %s", activePhase.ID)
		}

		// Add workflow-specific reference if epic has workflow defined
		if epicData.Workflow != "" {
			hint.Reference = fmt.Sprintf("Workflow: %s", epicData.Workflow)
		}
	}

	return hint
}

func (g *EpicPhaseAwareHintGenerator) generateTaskWorkflowHint(ctx *HintContext, epicData *epic.Epic) *Hint {
	hint := &Hint{
		Category: HintCategoryWorkflow,
		Priority: HintPriorityMedium,
	}

	// Find the phase containing the tasks
	var taskPhase *epic.Phase
	var activeTask *epic.Task

	for i := range epicData.Tasks {
		if epicData.Tasks[i].Status == epic.StatusActive {
			activeTask = &epicData.Tasks[i]
			// Find the phase for this task
			for j := range epicData.Phases {
				if epicData.Phases[j].ID == epicData.Tasks[i].PhaseID {
					taskPhase = &epicData.Phases[j]
					break
				}
			}
			break
		}
	}

	if activeTask != nil && taskPhase != nil {
		hint.Content = fmt.Sprintf("Complete task '%s' in phase '%s' before starting another task in the same phase",
			activeTask.ID, taskPhase.ID)
		hint.Command = fmt.Sprintf("agentpm done-task %s", activeTask.ID)
		hint.Conditions = []string{
			"Only one task per phase can be active",
			"Complete current task before starting another",
		}

		// Add phase context
		if taskPhase.Name != "" {
			hint.Reference = fmt.Sprintf("Phase: %s", taskPhase.Name)
		}
	}

	return hint
}

func (g *EpicPhaseAwareHintGenerator) generateGenericWorkflowHint(ctx *HintContext, epicData *epic.Epic) *Hint {
	hint := &Hint{
		Category: HintCategoryWorkflow,
		Priority: HintPriorityLow,
	}

	// Provide epic-level workflow guidance
	hint.Content = fmt.Sprintf("Epic '%s' workflow guidance available", epicData.ID)
	hint.Command = "agentpm current"

	if epicData.Workflow != "" {
		hint.Reference = fmt.Sprintf("Workflow: %s", epicData.Workflow)
	}

	return hint
}

func (g *EpicPhaseAwareHintGenerator) analyzePhaseDependencies(epicData *epic.Epic, activePhase, targetPhase *epic.Phase) []string {
	var dependencies []string

	// Simple dependency analysis - in a real implementation, this could check:
	// - Phase order in the epic
	// - Explicit dependencies defined in phase metadata
	// - Task completion requirements

	// For now, check if target phase comes after active phase in the list
	activeIndex := -1
	targetIndex := -1

	for i, phase := range epicData.Phases {
		if phase.ID == activePhase.ID {
			activeIndex = i
		}
		if phase.ID == targetPhase.ID {
			targetIndex = i
		}
	}

	if activeIndex >= 0 && targetIndex >= 0 && targetIndex > activeIndex {
		dependencies = append(dependencies, "Sequential phase order")
	}

	return dependencies
}

func (g *EpicPhaseAwareHintGenerator) Priority() int { return 90 }
