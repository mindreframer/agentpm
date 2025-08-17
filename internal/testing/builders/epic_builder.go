package builders

import (
	"fmt"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
)

// EpicBuilder provides a fluent API for building test epics
type EpicBuilder struct {
	id            string
	name          string
	status        string
	createdAt     *time.Time
	assignee      string
	description   string
	workflow      string
	requirements  string
	dependencies  string
	phases        []PhaseConfig
	tasks         []TaskConfig
	tests         []TestConfig
	events        []EventConfig
	defaultValues bool
}

// PhaseConfig holds configuration for a phase to be built
type PhaseConfig struct {
	ID           string
	Name         string
	Description  string
	Deliverables string
	Status       string
}

// TaskConfig holds configuration for a task to be built
type TaskConfig struct {
	ID                 string
	PhaseID            string
	Name               string
	Description        string
	AcceptanceCriteria string
	Status             string
	Assignee           string
}

// TestConfig holds configuration for a test to be built
type TestConfig struct {
	ID          string
	TaskID      string
	PhaseID     string
	Name        string
	Description string
	Status      string
	TestStatus  string
	TestResult  string
}

// EventConfig holds configuration for an event to be built
type EventConfig struct {
	ID        string
	Type      string
	Timestamp *time.Time
	Data      string
}

// NewEpicBuilder creates a new EpicBuilder with the given ID
func NewEpicBuilder(id string) *EpicBuilder {
	return &EpicBuilder{
		id:            id,
		name:          id,
		status:        string(epic.StatusPlanning),
		defaultValues: true,
		phases:        make([]PhaseConfig, 0),
		tasks:         make([]TaskConfig, 0),
		tests:         make([]TestConfig, 0),
		events:        make([]EventConfig, 0),
	}
}

// CreateEpicBuilder creates a new EpicBuilder (alias for NewEpicBuilder for fluent syntax)
func CreateEpicBuilder(id string) *EpicBuilder {
	return NewEpicBuilder(id)
}

// WithName sets the epic name
func (b *EpicBuilder) WithName(name string) *EpicBuilder {
	b.name = name
	return b
}

// WithStatus sets the epic status
func (b *EpicBuilder) WithStatus(status string) *EpicBuilder {
	b.status = status
	return b
}

// WithCreatedAt sets the creation timestamp
func (b *EpicBuilder) WithCreatedAt(timestamp time.Time) *EpicBuilder {
	b.createdAt = &timestamp
	return b
}

// WithAssignee sets the epic assignee
func (b *EpicBuilder) WithAssignee(assignee string) *EpicBuilder {
	b.assignee = assignee
	return b
}

// WithDescription sets the epic description
func (b *EpicBuilder) WithDescription(description string) *EpicBuilder {
	b.description = description
	return b
}

// WithWorkflow sets the epic workflow
func (b *EpicBuilder) WithWorkflow(workflow string) *EpicBuilder {
	b.workflow = workflow
	return b
}

// WithRequirements sets the epic requirements
func (b *EpicBuilder) WithRequirements(requirements string) *EpicBuilder {
	b.requirements = requirements
	return b
}

// WithDependencies sets the epic dependencies
func (b *EpicBuilder) WithDependencies(dependencies string) *EpicBuilder {
	b.dependencies = dependencies
	return b
}

// WithPhase adds a phase to the epic
func (b *EpicBuilder) WithPhase(id, name, status string) *EpicBuilder {
	b.phases = append(b.phases, PhaseConfig{
		ID:     id,
		Name:   name,
		Status: status,
	})
	return b
}

// WithPhaseDescriptive adds a phase with description and deliverables
func (b *EpicBuilder) WithPhaseDescriptive(id, name, description, deliverables, status string) *EpicBuilder {
	b.phases = append(b.phases, PhaseConfig{
		ID:           id,
		Name:         name,
		Description:  description,
		Deliverables: deliverables,
		Status:       status,
	})
	return b
}

// WithTask adds a task to the epic
func (b *EpicBuilder) WithTask(id, phaseID, name, status string) *EpicBuilder {
	b.tasks = append(b.tasks, TaskConfig{
		ID:      id,
		PhaseID: phaseID,
		Name:    name,
		Status:  status,
	})
	return b
}

// WithTaskDescriptive adds a task with description and acceptance criteria
func (b *EpicBuilder) WithTaskDescriptive(id, phaseID, name, description, acceptanceCriteria, status, assignee string) *EpicBuilder {
	b.tasks = append(b.tasks, TaskConfig{
		ID:                 id,
		PhaseID:            phaseID,
		Name:               name,
		Description:        description,
		AcceptanceCriteria: acceptanceCriteria,
		Status:             status,
		Assignee:           assignee,
	})
	return b
}

// WithTest adds a test to the epic
func (b *EpicBuilder) WithTest(id, taskID, phaseID, name, status string) *EpicBuilder {
	b.tests = append(b.tests, TestConfig{
		ID:      id,
		TaskID:  taskID,
		PhaseID: phaseID,
		Name:    name,
		Status:  status,
	})
	return b
}

// WithTestDescriptive adds a test with description and Epic 13 status
func (b *EpicBuilder) WithTestDescriptive(id, taskID, phaseID, name, description, status, testStatus, testResult string) *EpicBuilder {
	b.tests = append(b.tests, TestConfig{
		ID:          id,
		TaskID:      taskID,
		PhaseID:     phaseID,
		Name:        name,
		Description: description,
		Status:      status,
		TestStatus:  testStatus,
		TestResult:  testResult,
	})
	return b
}

// WithEvent adds an event to the epic
func (b *EpicBuilder) WithEvent(id, eventType, data string, timestamp time.Time) *EpicBuilder {
	b.events = append(b.events, EventConfig{
		ID:        id,
		Type:      eventType,
		Timestamp: &timestamp,
		Data:      data,
	})
	return b
}

// DisableDefaultValues disables automatic generation of default values
func (b *EpicBuilder) DisableDefaultValues() *EpicBuilder {
	b.defaultValues = false
	return b
}

// Build constructs the epic from the builder configuration
func (b *EpicBuilder) Build() (*epic.Epic, error) {
	// Validate required fields
	if err := b.validate(); err != nil {
		return nil, err
	}

	// Determine creation time
	createdAt := time.Now()
	if b.createdAt != nil {
		createdAt = *b.createdAt
	}

	// Create the epic structure
	result := &epic.Epic{
		ID:           b.id,
		Name:         b.name,
		Status:       epic.Status(b.status),
		CreatedAt:    createdAt,
		Assignee:     b.assignee,
		Description:  b.description,
		Workflow:     b.workflow,
		Requirements: b.requirements,
		Dependencies: b.dependencies,
	}

	// Add metadata if default values are enabled
	if b.defaultValues {
		result.Metadata = &epic.EpicMetadata{
			Created:         createdAt,
			Assignee:        b.assignee,
			EstimatedEffort: "",
		}

		result.CurrentState = &epic.CurrentState{
			ActivePhase: "",
			ActiveTask:  "",
			NextAction:  "Start next phase",
		}
	}

	// Build phases
	result.Phases = make([]epic.Phase, len(b.phases))
	for i, phaseConfig := range b.phases {
		result.Phases[i] = epic.Phase{
			ID:           phaseConfig.ID,
			Name:         phaseConfig.Name,
			Description:  phaseConfig.Description,
			Deliverables: phaseConfig.Deliverables,
			Status:       epic.Status(phaseConfig.Status),
		}
	}

	// Build tasks
	result.Tasks = make([]epic.Task, len(b.tasks))
	for i, taskConfig := range b.tasks {
		result.Tasks[i] = epic.Task{
			ID:                 taskConfig.ID,
			PhaseID:            taskConfig.PhaseID,
			Name:               taskConfig.Name,
			Description:        taskConfig.Description,
			AcceptanceCriteria: taskConfig.AcceptanceCriteria,
			Status:             epic.Status(taskConfig.Status),
			Assignee:           taskConfig.Assignee,
		}
	}

	// Build tests
	result.Tests = make([]epic.Test, len(b.tests))
	for i, testConfig := range b.tests {
		result.Tests[i] = epic.Test{
			ID:          testConfig.ID,
			TaskID:      testConfig.TaskID,
			PhaseID:     testConfig.PhaseID,
			Name:        testConfig.Name,
			Description: testConfig.Description,
			Status:      epic.Status(testConfig.Status),
		}

		// Set Epic 13 unified status system fields if provided
		if testConfig.TestStatus != "" {
			result.Tests[i].TestStatus = epic.TestStatus(testConfig.TestStatus)
		}
		if testConfig.TestResult != "" {
			result.Tests[i].TestResult = epic.TestResult(testConfig.TestResult)
		}
	}

	// Build events
	result.Events = make([]epic.Event, len(b.events))
	for i, eventConfig := range b.events {
		timestamp := time.Now()
		if eventConfig.Timestamp != nil {
			timestamp = *eventConfig.Timestamp
		}

		result.Events[i] = epic.Event{
			ID:        eventConfig.ID,
			Type:      eventConfig.Type,
			Timestamp: timestamp,
			Data:      eventConfig.Data,
		}
	}

	return result, nil
}

// validate checks the builder configuration for errors
func (b *EpicBuilder) validate() error {
	// Check required fields
	if b.id == "" {
		return fmt.Errorf("epic ID is required")
	}
	if b.name == "" {
		return fmt.Errorf("epic name is required")
	}

	// Validate status
	if !epic.Status(b.status).IsValid() {
		return fmt.Errorf("invalid epic status: %s", b.status)
	}

	// Validate phase relationships
	phaseIDs := make(map[string]bool)
	for _, phase := range b.phases {
		if phase.ID == "" {
			return fmt.Errorf("phase ID is required")
		}
		if phaseIDs[phase.ID] {
			return fmt.Errorf("duplicate phase ID: %s", phase.ID)
		}
		phaseIDs[phase.ID] = true

		if !epic.Status(phase.Status).IsValid() {
			return fmt.Errorf("invalid phase status for phase %s: %s", phase.ID, phase.Status)
		}
	}

	// Validate task relationships
	taskIDs := make(map[string]bool)
	for _, task := range b.tasks {
		if task.ID == "" {
			return fmt.Errorf("task ID is required")
		}
		if taskIDs[task.ID] {
			return fmt.Errorf("duplicate task ID: %s", task.ID)
		}
		taskIDs[task.ID] = true

		// Check that task's phase exists
		if task.PhaseID != "" && !phaseIDs[task.PhaseID] {
			return fmt.Errorf("task %s references non-existent phase: %s", task.ID, task.PhaseID)
		}

		if !epic.Status(task.Status).IsValid() {
			return fmt.Errorf("invalid task status for task %s: %s", task.ID, task.Status)
		}
	}

	// Validate test relationships
	testIDs := make(map[string]bool)
	for _, test := range b.tests {
		if test.ID == "" {
			return fmt.Errorf("test ID is required")
		}
		if testIDs[test.ID] {
			return fmt.Errorf("duplicate test ID: %s", test.ID)
		}
		testIDs[test.ID] = true

		// Check that test's task exists
		if test.TaskID != "" && !taskIDs[test.TaskID] {
			return fmt.Errorf("test %s references non-existent task: %s", test.ID, test.TaskID)
		}

		// Check that test's phase exists
		if test.PhaseID != "" && !phaseIDs[test.PhaseID] {
			return fmt.Errorf("test %s references non-existent phase: %s", test.ID, test.PhaseID)
		}

		if !epic.Status(test.Status).IsValid() {
			return fmt.Errorf("invalid test status for test %s: %s", test.ID, test.Status)
		}

		// Validate Epic 13 unified status fields if provided
		if test.TestStatus != "" && !epic.TestStatus(test.TestStatus).IsValid() {
			return fmt.Errorf("invalid test status for test %s: %s", test.ID, test.TestStatus)
		}
		if test.TestResult != "" && !epic.TestResult(test.TestResult).IsValid() {
			return fmt.Errorf("invalid test result for test %s: %s", test.ID, test.TestResult)
		}
	}

	return nil
}
