package assertions

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// DebugMode represents the debugging level for the framework
type DebugMode int

const (
	DebugOff DebugMode = iota
	DebugBasic
	DebugVerbose
	DebugTrace
)

// DebugContext provides debugging capabilities for the testing framework
type DebugContext struct {
	mode     DebugMode
	traceLog []TraceEntry
	enabled  bool
}

// TraceEntry represents a single debugging trace entry
type TraceEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Context   map[string]interface{}
	Location  string
}

// ErrorContext provides enhanced error information with state data
type ErrorContext struct {
	Err         error
	Stage       string
	ChainIndex  int
	StateInfo   map[string]interface{}
	Suggestions []string
	DebugTrace  []TraceEntry
}

// StateVisualization provides tools for understanding complex state transitions
type StateVisualization struct {
	StateDiffs []StateDiff
	Timeline   []StateSnapshot
	Graph      TransitionGraph
}

// StateDiff represents changes between two states
type StateDiff struct {
	Field    string
	Before   interface{}
	After    interface{}
	ChangeID string
}

// StateSnapshot captures state at a specific point in time
type StateSnapshot struct {
	Timestamp time.Time
	Index     int
	State     interface{}
	Context   string
}

// TransitionGraph represents the flow of state changes
type TransitionGraph struct {
	Nodes []GraphNode
	Edges []GraphEdge
}

// GraphNode represents a state in the transition graph
type GraphNode struct {
	ID       string
	Label    string
	State    interface{}
	Metadata map[string]interface{}
}

// GraphEdge represents a transition between states
type GraphEdge struct {
	From       string
	To         string
	Command    string
	Duration   time.Duration
	Successful bool
	Metadata   map[string]interface{}
}

// RecoveryStrategy defines how to handle and recover from errors
type RecoveryStrategy struct {
	CanRecover   func(error) bool
	RecoverFunc  func(error, *ErrorContext) error
	ContinueFunc func(*ErrorContext) bool
}

// NewDebugContext creates a new debugging context
func NewDebugContext(mode DebugMode) *DebugContext {
	return &DebugContext{
		mode:     mode,
		traceLog: make([]TraceEntry, 0),
		enabled:  mode != DebugOff,
	}
}

// Trace adds a trace entry to the debug context
func (dc *DebugContext) Trace(level, message string, context map[string]interface{}) {
	if !dc.enabled {
		return
	}

	entry := TraceEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Context:   context,
		Location:  dc.getCallerLocation(),
	}

	dc.traceLog = append(dc.traceLog, entry)

	// Print to stdout if in verbose mode
	if dc.mode >= DebugVerbose {
		fmt.Printf("[%s] %s: %s (at %s)\n",
			entry.Timestamp.Format("15:04:05.000"),
			level,
			message,
			entry.Location)
		if len(context) > 0 && dc.mode >= DebugTrace {
			fmt.Printf("  Context: %+v\n", context)
		}
	}
}

// getCallerLocation returns the file:line of the caller
func (dc *DebugContext) getCallerLocation() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}

	// Get just the filename, not the full path
	parts := strings.Split(file, "/")
	filename := parts[len(parts)-1]

	return fmt.Sprintf("%s:%d", filename, line)
}

// GetTraceLog returns all trace entries
func (dc *DebugContext) GetTraceLog() []TraceEntry {
	return dc.traceLog
}

// Clear clears the trace log
func (dc *DebugContext) Clear() {
	dc.traceLog = make([]TraceEntry, 0)
}

// CreateErrorContext creates an enhanced error context with debugging information
func CreateErrorContext(err error, stage string, chainIndex int, state interface{}, debugCtx *DebugContext) *ErrorContext {
	errorCtx := &ErrorContext{
		Err:         err,
		Stage:       stage,
		ChainIndex:  chainIndex,
		StateInfo:   make(map[string]interface{}),
		Suggestions: make([]string, 0),
	}

	// Add state information
	if state != nil {
		errorCtx.StateInfo["current_state"] = state
		errorCtx.StateInfo["stage"] = stage
		errorCtx.StateInfo["chain_index"] = chainIndex
	}

	// Add debug trace if available
	if debugCtx != nil {
		errorCtx.DebugTrace = debugCtx.GetTraceLog()
	}

	// Generate helpful suggestions based on error type
	errorCtx.Suggestions = generateErrorSuggestions(err, stage, chainIndex)

	return errorCtx
}

// generateErrorSuggestions creates helpful suggestions based on the error
func generateErrorSuggestions(err error, stage string, chainIndex int) []string {
	suggestions := make([]string, 0)
	errorMsg := err.Error()

	// Common error patterns and suggestions
	if strings.Contains(errorMsg, "assertion failed") {
		suggestions = append(suggestions, "Check if the expected state matches the actual result")
		suggestions = append(suggestions, "Review the transition chain logic leading to this assertion")
	}

	if strings.Contains(errorMsg, "timeout") {
		suggestions = append(suggestions, "Consider increasing timeout values for slow operations")
		suggestions = append(suggestions, "Check if there are blocking operations in the chain")
	}

	if strings.Contains(errorMsg, "phase not found") {
		suggestions = append(suggestions, "Verify phase ID exists in the epic structure")
		suggestions = append(suggestions, "Check for typos in phase identifiers")
	}

	if strings.Contains(errorMsg, "task not found") {
		suggestions = append(suggestions, "Verify task ID exists in the specified phase")
		suggestions = append(suggestions, "Ensure task is properly defined in the epic XML")
	}

	// Stage-specific suggestions
	switch stage {
	case "setup":
		suggestions = append(suggestions, "Review epic file initialization and configuration")
	case "execution":
		suggestions = append(suggestions, "Check command parameters and epic state")
	case "validation":
		suggestions = append(suggestions, "Verify assertion logic and expected outcomes")
	case "cleanup":
		suggestions = append(suggestions, "Ensure proper resource cleanup and state reset")
	}

	// Chain index specific suggestions
	if chainIndex > 0 {
		suggestions = append(suggestions, fmt.Sprintf("This error occurred at step %d in the chain", chainIndex+1))
		suggestions = append(suggestions, "Review previous steps that may have caused this state")
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Enable debug mode for more detailed error information")
	}

	return suggestions
}

// Error returns a formatted error string with context
func (ec *ErrorContext) Error() string {
	var parts []string

	parts = append(parts, fmt.Sprintf("Error in %s", ec.Stage))
	if ec.ChainIndex >= 0 {
		parts = append(parts, fmt.Sprintf("at chain step %d", ec.ChainIndex))
	}
	parts = append(parts, fmt.Sprintf(": %s", ec.Err.Error()))

	if len(ec.Suggestions) > 0 {
		parts = append(parts, "\nSuggestions:")
		for _, suggestion := range ec.Suggestions {
			parts = append(parts, fmt.Sprintf("  - %s", suggestion))
		}
	}

	return strings.Join(parts, " ")
}

// CreateStateVisualization generates a visualization of state transitions
func CreateStateVisualization(states []interface{}, commands []string) *StateVisualization {
	viz := &StateVisualization{
		StateDiffs: make([]StateDiff, 0),
		Timeline:   make([]StateSnapshot, 0),
		Graph: TransitionGraph{
			Nodes: make([]GraphNode, 0),
			Edges: make([]GraphEdge, 0),
		},
	}

	// Create timeline snapshots
	for i, state := range states {
		snapshot := StateSnapshot{
			Timestamp: time.Now().Add(time.Duration(i) * time.Millisecond),
			Index:     i,
			State:     state,
			Context:   fmt.Sprintf("Step %d", i),
		}
		viz.Timeline = append(viz.Timeline, snapshot)
	}

	// Create graph nodes
	for i, state := range states {
		node := GraphNode{
			ID:    fmt.Sprintf("state_%d", i),
			Label: fmt.Sprintf("State %d", i),
			State: state,
			Metadata: map[string]interface{}{
				"index": i,
				"type":  "state_node",
			},
		}
		viz.Graph.Nodes = append(viz.Graph.Nodes, node)
	}

	// Create graph edges
	for i, command := range commands {
		if i < len(states)-1 {
			edge := GraphEdge{
				From:       fmt.Sprintf("state_%d", i),
				To:         fmt.Sprintf("state_%d", i+1),
				Command:    command,
				Duration:   time.Millisecond * 100, // Default duration
				Successful: true,                   // Assume successful unless specified
				Metadata: map[string]interface{}{
					"command_index": i,
				},
			}
			viz.Graph.Edges = append(viz.Graph.Edges, edge)
		}
	}

	return viz
}

// GetGraphVisualization returns a text representation of the transition graph
func (viz *StateVisualization) GetGraphVisualization() string {
	var lines []string
	lines = append(lines, "State Transition Graph:")
	lines = append(lines, "======================")

	for _, edge := range viz.Graph.Edges {
		status := "✓"
		if !edge.Successful {
			status = "✗"
		}

		line := fmt.Sprintf("%s --[%s]--> %s %s (%.2fms)",
			edge.From,
			edge.Command,
			edge.To,
			status,
			float64(edge.Duration.Nanoseconds())/1000000.0)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// GetTimelineVisualization returns a text representation of the state timeline
func (viz *StateVisualization) GetTimelineVisualization() string {
	var lines []string
	lines = append(lines, "State Timeline:")
	lines = append(lines, "===============")

	for _, snapshot := range viz.Timeline {
		line := fmt.Sprintf("[%s] %s: %v",
			snapshot.Timestamp.Format("15:04:05.000"),
			snapshot.Context,
			snapshot.State)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// DefaultRecoveryStrategy provides a basic error recovery strategy
func DefaultRecoveryStrategy() *RecoveryStrategy {
	return &RecoveryStrategy{
		CanRecover: func(err error) bool {
			// Can recover from assertion failures but not system errors
			errorMsg := err.Error()
			return strings.Contains(errorMsg, "assertion failed") ||
				strings.Contains(errorMsg, "validation failed") ||
				strings.Contains(errorMsg, "Expected epic status") ||
				strings.Contains(errorMsg, "Phase") && strings.Contains(errorMsg, "not found")
		},
		RecoverFunc: func(err error, ctx *ErrorContext) error {
			// For now, just log the recovery attempt
			if ctx != nil {
				ctx.Suggestions = append(ctx.Suggestions, "Recovery attempted but test will continue")
			}
			return nil
		},
		ContinueFunc: func(ctx *ErrorContext) bool {
			// Continue execution unless it's a critical error
			if ctx != nil {
				errorMsg := ctx.Err.Error()
				return !strings.Contains(errorMsg, "critical") &&
					!strings.Contains(errorMsg, "fatal")
			}
			return true
		},
	}
}
