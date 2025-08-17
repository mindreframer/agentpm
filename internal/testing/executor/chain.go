package executor

import (
	"fmt"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/lifecycle"
	"github.com/mindreframer/agentpm/internal/phases"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/tasks"
	"github.com/mindreframer/agentpm/internal/tests"
)

// TransitionChain provides a fluent API for executing command chains on epics
type TransitionChain struct {
	environment            *TestExecutionEnvironment
	lifecycleService       *lifecycle.LifecycleService
	phaseService           *phases.PhaseService
	taskService            *tasks.TaskService
	testService            *tests.TestService
	queryService           *query.QueryService
	commands               []ChainCommand
	timeSource             func() time.Time
	intermediateAssertions []IntermediateAssertion
}

// ChainCommand represents a command to be executed in the chain
type ChainCommand struct {
	Type        string
	Target      string
	Description string
	Timestamp   *time.Time
}

// IntermediateAssertion represents an assertion to be made during chain execution
type IntermediateAssertion struct {
	AfterCommand string
	AssertionFn  func(*epic.Epic) error
}

// TransitionChainResult contains the results of executing a transition chain
type TransitionChainResult struct {
	Environment        *TestExecutionEnvironment
	InitialState       *epic.Epic
	FinalState         *epic.Epic
	IntermediateStates []StateSnapshot
	ExecutedCommands   []CommandExecution
	Errors             []TransitionError
	ExecutionTime      time.Duration
	MemoryUsage        int64
	Success            bool
}

// CommandExecution tracks details about each executed command
type CommandExecution struct {
	Command   ChainCommand
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Success   bool
	Error     error
	EpicState *epic.Epic
}

// TransitionError represents an error that occurred during transition
type TransitionError struct {
	Command        string
	ExpectedState  string
	ActualState    string
	Epic           *epic.Epic
	ContextualInfo map[string]interface{}
	Suggestions    []string
	OriginalError  error
}

func (e TransitionError) Error() string {
	if e.OriginalError != nil {
		return fmt.Sprintf("transition error in %s: %v", e.Command, e.OriginalError)
	}
	return fmt.Sprintf("transition error in %s: expected %s, got %s", e.Command, e.ExpectedState, e.ActualState)
}

// NewTransitionChain creates a new transition chain for the given epic
func NewTransitionChain(env *TestExecutionEnvironment) *TransitionChain {
	// Create query service
	queryService := query.NewQueryService(env.GetStorage())

	// Create command services
	lifecycleService := lifecycle.NewLifecycleService(env.GetStorage(), queryService)
	phaseService := phases.NewPhaseService(env.GetStorage(), queryService)
	taskService := tasks.NewTaskService(env.GetStorage(), queryService)
	testService := tests.NewTestService(tests.ServiceConfig{
		UseMemory:  true,
		TimeSource: nil, // Will be set by WithTimeSource
	})

	return &TransitionChain{
		environment:            env,
		lifecycleService:       lifecycleService,
		phaseService:           phaseService,
		taskService:            taskService,
		testService:            testService,
		queryService:           queryService,
		commands:               make([]ChainCommand, 0),
		timeSource:             time.Now,
		intermediateAssertions: make([]IntermediateAssertion, 0),
	}
}

// CreateTransitionChain creates a new transition chain for the given epic (factory function)
func CreateTransitionChain(env *TestExecutionEnvironment) *TransitionChain {
	return NewTransitionChain(env)
}

// WithTimeSource allows injection of custom time source for deterministic testing
func (tc *TransitionChain) WithTimeSource(timeSource func() time.Time) *TransitionChain {
	tc.timeSource = timeSource
	tc.lifecycleService = tc.lifecycleService.WithTimeSource(timeSource)
	return tc
}

// StartEpic adds a start epic command to the chain
func (tc *TransitionChain) StartEpic() *TransitionChain {
	tc.commands = append(tc.commands, ChainCommand{
		Type:        "start_epic",
		Target:      "",
		Description: "Start epic transition",
	})
	return tc
}

// StartEpicAt adds a start epic command with specific timestamp
func (tc *TransitionChain) StartEpicAt(timestamp time.Time) *TransitionChain {
	tc.commands = append(tc.commands, ChainCommand{
		Type:        "start_epic",
		Target:      "",
		Description: "Start epic transition",
		Timestamp:   &timestamp,
	})
	return tc
}

// StartPhase adds a start phase command to the chain
func (tc *TransitionChain) StartPhase(phaseID string) *TransitionChain {
	tc.commands = append(tc.commands, ChainCommand{
		Type:        "start_phase",
		Target:      phaseID,
		Description: fmt.Sprintf("Start phase %s", phaseID),
	})
	return tc
}

// StartPhaseAt adds a start phase command with specific timestamp
func (tc *TransitionChain) StartPhaseAt(phaseID string, timestamp time.Time) *TransitionChain {
	tc.commands = append(tc.commands, ChainCommand{
		Type:        "start_phase",
		Target:      phaseID,
		Description: fmt.Sprintf("Start phase %s", phaseID),
		Timestamp:   &timestamp,
	})
	return tc
}

// DonePhase adds a complete phase command to the chain
func (tc *TransitionChain) DonePhase(phaseID string) *TransitionChain {
	tc.commands = append(tc.commands, ChainCommand{
		Type:        "done_phase",
		Target:      phaseID,
		Description: fmt.Sprintf("Complete phase %s", phaseID),
	})
	return tc
}

// DonePhaseAt adds a complete phase command with specific timestamp
func (tc *TransitionChain) DonePhaseAt(phaseID string, timestamp time.Time) *TransitionChain {
	tc.commands = append(tc.commands, ChainCommand{
		Type:        "done_phase",
		Target:      phaseID,
		Description: fmt.Sprintf("Complete phase %s", phaseID),
		Timestamp:   &timestamp,
	})
	return tc
}

// StartTask adds a start task command to the chain
func (tc *TransitionChain) StartTask(taskID string) *TransitionChain {
	tc.commands = append(tc.commands, ChainCommand{
		Type:        "start_task",
		Target:      taskID,
		Description: fmt.Sprintf("Start task %s", taskID),
	})
	return tc
}

// StartTaskAt adds a start task command with specific timestamp
func (tc *TransitionChain) StartTaskAt(taskID string, timestamp time.Time) *TransitionChain {
	tc.commands = append(tc.commands, ChainCommand{
		Type:        "start_task",
		Target:      taskID,
		Description: fmt.Sprintf("Start task %s", taskID),
		Timestamp:   &timestamp,
	})
	return tc
}

// DoneTask adds a complete task command to the chain
func (tc *TransitionChain) DoneTask(taskID string) *TransitionChain {
	tc.commands = append(tc.commands, ChainCommand{
		Type:        "done_task",
		Target:      taskID,
		Description: fmt.Sprintf("Complete task %s", taskID),
	})
	return tc
}

// DoneTaskAt adds a complete task command with specific timestamp
func (tc *TransitionChain) DoneTaskAt(taskID string, timestamp time.Time) *TransitionChain {
	tc.commands = append(tc.commands, ChainCommand{
		Type:        "done_task",
		Target:      taskID,
		Description: fmt.Sprintf("Complete task %s", taskID),
		Timestamp:   &timestamp,
	})
	return tc
}

// PassTest adds a pass test command to the chain
func (tc *TransitionChain) PassTest(testID string) *TransitionChain {
	tc.commands = append(tc.commands, ChainCommand{
		Type:        "pass_test",
		Target:      testID,
		Description: fmt.Sprintf("Pass test %s", testID),
	})
	return tc
}

// PassTestAt adds a pass test command with specific timestamp
func (tc *TransitionChain) PassTestAt(testID string, timestamp time.Time) *TransitionChain {
	tc.commands = append(tc.commands, ChainCommand{
		Type:        "pass_test",
		Target:      testID,
		Description: fmt.Sprintf("Pass test %s", testID),
		Timestamp:   &timestamp,
	})
	return tc
}

// FailTest adds a fail test command to the chain
func (tc *TransitionChain) FailTest(testID string) *TransitionChain {
	tc.commands = append(tc.commands, ChainCommand{
		Type:        "fail_test",
		Target:      testID,
		Description: fmt.Sprintf("Fail test %s", testID),
	})
	return tc
}

// FailTestAt adds a fail test command with specific timestamp
func (tc *TransitionChain) FailTestAt(testID string, timestamp time.Time) *TransitionChain {
	tc.commands = append(tc.commands, ChainCommand{
		Type:        "fail_test",
		Target:      testID,
		Description: fmt.Sprintf("Fail test %s", testID),
		Timestamp:   &timestamp,
	})
	return tc
}

// DoneEpic adds a complete epic command to the chain
func (tc *TransitionChain) DoneEpic() *TransitionChain {
	tc.commands = append(tc.commands, ChainCommand{
		Type:        "done_epic",
		Target:      "",
		Description: "Complete epic",
	})
	return tc
}

// DoneEpicAt adds a complete epic command with specific timestamp
func (tc *TransitionChain) DoneEpicAt(timestamp time.Time) *TransitionChain {
	tc.commands = append(tc.commands, ChainCommand{
		Type:        "done_epic",
		Target:      "",
		Description: "Complete epic",
		Timestamp:   &timestamp,
	})
	return tc
}

// Assert adds an intermediate assertion to be checked during execution
func (tc *TransitionChain) Assert() *IntermediateAssertionBuilder {
	return NewIntermediateAssertionBuilder(tc)
}

// AddIntermediateAssertion adds a custom assertion (used by assertion builder)
func (tc *TransitionChain) AddIntermediateAssertion(afterCommand string, assertionFn func(*epic.Epic) error) *TransitionChain {
	tc.intermediateAssertions = append(tc.intermediateAssertions, IntermediateAssertion{
		AfterCommand: afterCommand,
		AssertionFn:  assertionFn,
	})
	return tc
}

// Execute runs all commands in the chain and returns the results
func (tc *TransitionChain) Execute() (*TransitionChainResult, error) {
	startTime := tc.timeSource()

	initialState, err := tc.environment.GetCurrentEpic()
	if err != nil {
		return nil, fmt.Errorf("failed to get initial epic state: %w", err)
	}

	executedCommands := make([]CommandExecution, 0, len(tc.commands))
	errors := make([]TransitionError, 0)

	// Execute each command in sequence
	for i, command := range tc.commands {
		cmdStartTime := tc.timeSource()

		// Execute the command
		err := tc.executeCommand(command)

		cmdEndTime := tc.timeSource()
		cmdDuration := cmdEndTime.Sub(cmdStartTime)

		// Get epic state after command
		epicState, getErr := tc.environment.GetCurrentEpic()
		if getErr != nil {
			epicState = nil
		}

		// Record command execution
		execution := CommandExecution{
			Command:   command,
			StartTime: cmdStartTime,
			EndTime:   cmdEndTime,
			Duration:  cmdDuration,
			Success:   err == nil,
			Error:     err,
			EpicState: epicState,
		}
		executedCommands = append(executedCommands, execution)

		// If command failed, record error
		if err != nil {
			transitionErr := TransitionError{
				Command:       command.Type,
				Epic:          epicState,
				OriginalError: err,
				ContextualInfo: map[string]interface{}{
					"command_index": i,
					"target":        command.Target,
					"description":   command.Description,
				},
			}
			errors = append(errors, transitionErr)
		}

		// Check intermediate assertions
		for _, assertion := range tc.intermediateAssertions {
			if assertion.AfterCommand == command.Type || assertion.AfterCommand == fmt.Sprintf("%s:%s", command.Type, command.Target) {
				if epicState != nil {
					if assertErr := assertion.AssertionFn(epicState); assertErr != nil {
						transitionErr := TransitionError{
							Command:       fmt.Sprintf("assertion_after_%s", command.Type),
							Epic:          epicState,
							OriginalError: assertErr,
							ContextualInfo: map[string]interface{}{
								"assertion_type": "intermediate",
								"after_command":  command.Type,
							},
						}
						errors = append(errors, transitionErr)
					}
				}
			}
		}
	}

	endTime := tc.timeSource()
	executionTime := endTime.Sub(startTime)

	finalState, err := tc.environment.GetCurrentEpic()
	if err != nil {
		finalState = nil
	}

	result := &TransitionChainResult{
		Environment:        tc.environment,
		InitialState:       initialState,
		FinalState:         finalState,
		IntermediateStates: tc.environment.GetSnapshots(),
		ExecutedCommands:   executedCommands,
		Errors:             errors,
		ExecutionTime:      executionTime,
		MemoryUsage:        0, // TODO: Implement memory usage tracking
		Success:            len(errors) == 0,
	}

	return result, nil
}

// executeCommand executes a single command using the appropriate service
func (tc *TransitionChain) executeCommand(command ChainCommand) error {
	// Get current epic state
	currentEpic, err := tc.environment.GetCurrentEpic()
	if err != nil {
		return fmt.Errorf("failed to get current epic: %w", err)
	}

	// Determine timestamp
	timestamp := tc.timeSource()
	if command.Timestamp != nil {
		timestamp = *command.Timestamp
	}

	switch command.Type {
	case "start_epic":
		request := lifecycle.StartEpicRequest{
			EpicFile:  tc.environment.GetEpicFile(),
			Timestamp: &timestamp,
		}
		_, err = tc.lifecycleService.StartEpic(request)
		if err != nil {
			return err
		}

	case "done_epic":
		request := lifecycle.DoneEpicRequest{
			EpicFile:  tc.environment.GetEpicFile(),
			Timestamp: &timestamp,
		}
		_, err = tc.lifecycleService.DoneEpic(request)
		if err != nil {
			return err
		}

	case "start_phase":
		err = tc.phaseService.StartPhase(currentEpic, command.Target, timestamp)
		if err != nil {
			return err
		}
		// Save the updated epic
		err = tc.environment.SaveEpic(currentEpic, fmt.Sprintf("start_phase_%s", command.Target))
		if err != nil {
			return err
		}

	case "done_phase":
		err = tc.phaseService.CompletePhase(currentEpic, command.Target, timestamp)
		if err != nil {
			return err
		}
		// Save the updated epic
		err = tc.environment.SaveEpic(currentEpic, fmt.Sprintf("done_phase_%s", command.Target))
		if err != nil {
			return err
		}

	case "start_task":
		err = tc.taskService.StartTask(currentEpic, command.Target, timestamp)
		if err != nil {
			return err
		}
		// Save the updated epic
		err = tc.environment.SaveEpic(currentEpic, fmt.Sprintf("start_task_%s", command.Target))
		if err != nil {
			return err
		}

	case "done_task":
		err = tc.taskService.CompleteTask(currentEpic, command.Target, timestamp)
		if err != nil {
			return err
		}
		// Save the updated epic
		err = tc.environment.SaveEpic(currentEpic, fmt.Sprintf("done_task_%s", command.Target))
		if err != nil {
			return err
		}

	case "pass_test":
		_, err = tc.testService.PassTest(tc.environment.GetEpicFile(), command.Target, &timestamp)
		if err != nil {
			return err
		}

	case "fail_test":
		_, err = tc.testService.FailTest(tc.environment.GetEpicFile(), command.Target, "Test failed during transition chain", &timestamp)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown command type: %s", command.Type)
	}

	return nil
}

// IntermediateAssertionBuilder provides a fluent API for building intermediate assertions
type IntermediateAssertionBuilder struct {
	chain        *TransitionChain
	afterCommand string
}

// NewIntermediateAssertionBuilder creates a new assertion builder
func NewIntermediateAssertionBuilder(chain *TransitionChain) *IntermediateAssertionBuilder {
	// Determine the command we're asserting after (the last command added)
	var afterCommand string
	if len(chain.commands) > 0 {
		lastCmd := chain.commands[len(chain.commands)-1]
		if lastCmd.Target != "" {
			afterCommand = fmt.Sprintf("%s:%s", lastCmd.Type, lastCmd.Target)
		} else {
			afterCommand = lastCmd.Type
		}
	}

	return &IntermediateAssertionBuilder{
		chain:        chain,
		afterCommand: afterCommand,
	}
}

// EpicStatus adds an epic status assertion
func (ab *IntermediateAssertionBuilder) EpicStatus(expectedStatus string) *TransitionChain {
	assertionFn := func(e *epic.Epic) error {
		if string(e.Status) != expectedStatus {
			return fmt.Errorf("expected epic status %s, got %s", expectedStatus, e.Status)
		}
		return nil
	}
	return ab.chain.AddIntermediateAssertion(ab.afterCommand, assertionFn)
}

// PhaseStatus adds a phase status assertion
func (ab *IntermediateAssertionBuilder) PhaseStatus(phaseID, expectedStatus string) *TransitionChain {
	assertionFn := func(e *epic.Epic) error {
		for _, phase := range e.Phases {
			if phase.ID == phaseID {
				if string(phase.Status) != expectedStatus {
					return fmt.Errorf("expected phase %s status %s, got %s", phaseID, expectedStatus, phase.Status)
				}
				return nil
			}
		}
		return fmt.Errorf("phase %s not found", phaseID)
	}
	return ab.chain.AddIntermediateAssertion(ab.afterCommand, assertionFn)
}

// TaskStatus adds a task status assertion
func (ab *IntermediateAssertionBuilder) TaskStatus(taskID, expectedStatus string) *TransitionChain {
	assertionFn := func(e *epic.Epic) error {
		for _, task := range e.Tasks {
			if task.ID == taskID {
				if string(task.Status) != expectedStatus {
					return fmt.Errorf("expected task %s status %s, got %s", taskID, expectedStatus, task.Status)
				}
				return nil
			}
		}
		return fmt.Errorf("task %s not found", taskID)
	}
	return ab.chain.AddIntermediateAssertion(ab.afterCommand, assertionFn)
}

// TestStatus adds a test status assertion
func (ab *IntermediateAssertionBuilder) TestStatus(testID, expectedStatus string) *TransitionChain {
	assertionFn := func(e *epic.Epic) error {
		for _, test := range e.Tests {
			if test.ID == testID {
				if string(test.Status) != expectedStatus {
					return fmt.Errorf("expected test %s status %s, got %s", testID, expectedStatus, test.Status)
				}
				return nil
			}
		}
		return fmt.Errorf("test %s not found", testID)
	}
	return ab.chain.AddIntermediateAssertion(ab.afterCommand, assertionFn)
}
