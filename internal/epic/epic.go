package epic

import (
	"time"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
	StatusOnHold    Status = "on_hold"
	StatusCancelled Status = "cancelled"
)

type Epic struct {
	ID           string        `xml:"id,attr"`
	Name         string        `xml:"name,attr"`
	Status       Status        `xml:"status,attr"`
	CreatedAt    time.Time     `xml:"created_at,attr"`
	Assignee     string        `xml:"assignee"`
	Description  string        `xml:"description"`
	Workflow     string        `xml:"workflow,omitempty"`
	Requirements string        `xml:"requirements,omitempty"`
	Dependencies string        `xml:"dependencies,omitempty"`
	Metadata     *EpicMetadata `xml:"metadata,omitempty"`
	CurrentState *CurrentState `xml:"current_state,omitempty"`
	Phases       []Phase       `xml:"phases>phase"`
	Tasks        []Task        `xml:"tasks>task"`
	Tests        []Test        `xml:"tests>test"`
	Events       []Event       `xml:"events>event"`
}

// Epic 13 Status System Methods

// GetEpicStatus returns the Epic 13 unified epic status
func (e *Epic) GetEpicStatus() EpicStatus {
	return e.Status.ToEpicStatus()
}

// SetEpicStatus sets the epic status using the Epic 13 unified system
func (e *Epic) SetEpicStatus(status EpicStatus) {
	e.Status = FromEpicStatus(status)
}

type EpicMetadata struct {
	Created         time.Time `xml:"created"`
	Assignee        string    `xml:"assignee"`
	EstimatedEffort string    `xml:"estimated_effort"`
}

type CurrentState struct {
	ActivePhase string `xml:"active_phase"`
	ActiveTask  string `xml:"active_task"`
	NextAction  string `xml:"next_action"`
}

type Phase struct {
	ID           string     `xml:"id,attr"`
	Name         string     `xml:"name,attr"`
	Description  string     `xml:"description"`
	Deliverables string     `xml:"deliverables"`
	Status       Status     `xml:"status,attr"`
	StartedAt    *time.Time `xml:"started_at,omitempty"`
	CompletedAt  *time.Time `xml:"completed_at,omitempty"`
}

// Epic 13 Status System Methods

// GetPhaseStatus returns the Epic 13 unified phase status
func (p *Phase) GetPhaseStatus() PhaseStatus {
	return p.Status.ToPhaseStatus()
}

// SetPhaseStatus sets the phase status using the Epic 13 unified system
func (p *Phase) SetPhaseStatus(status PhaseStatus) {
	p.Status = FromPhaseStatus(status)
}

type Task struct {
	ID                 string     `xml:"id,attr"`
	PhaseID            string     `xml:"phase_id,attr"`
	Name               string     `xml:"name,attr"`
	Description        string     `xml:"description"`
	AcceptanceCriteria string     `xml:"acceptance_criteria"`
	Status             Status     `xml:"status,attr"`
	Assignee           string     `xml:"assignee,attr,omitempty"`
	StartedAt          *time.Time `xml:"started_at,omitempty"`
	CompletedAt        *time.Time `xml:"completed_at,omitempty"`
	CancelledAt        *time.Time `xml:"cancelled_at,omitempty"`
}

// Epic 13 Status System Methods

// GetTaskStatus returns the Epic 13 unified task status
func (t *Task) GetTaskStatus() TaskStatus {
	return t.Status.ToTaskStatus()
}

// SetTaskStatus sets the task status using the Epic 13 unified system
func (t *Task) SetTaskStatus(status TaskStatus) {
	t.Status = FromTaskStatus(status)
}

type TestStatus string

const (
	TestStatusPending   TestStatus = "pending"
	TestStatusWIP       TestStatus = "wip"
	TestStatusDone      TestStatus = "done"
	TestStatusCancelled TestStatus = "cancelled"
)

func (s TestStatus) IsValid() bool {
	switch s {
	case TestStatusPending, TestStatusWIP, TestStatusDone, TestStatusCancelled:
		return true
	default:
		return false
	}
}

func (s TestStatus) CanTransitionTo(target TestStatus) bool {
	transitions := map[TestStatus][]TestStatus{
		TestStatusPending:   {TestStatusWIP, TestStatusCancelled},
		TestStatusWIP:       {TestStatusDone, TestStatusCancelled},
		TestStatusDone:      {TestStatusWIP}, // Can go back to WIP for failing tests
		TestStatusCancelled: {},              // Cancelled is terminal
	}

	for _, allowed := range transitions[s] {
		if allowed == target {
			return true
		}
	}
	return false
}

type Test struct {
	ID          string `xml:"id,attr"`
	TaskID      string `xml:"task_id,attr"`
	PhaseID     string `xml:"phase_id,attr"`
	Name        string `xml:"name,attr"`
	Description string `xml:"description"`
	Status      Status `xml:"status,attr"`
	// Epic 13 unified status system
	TestStatus         TestStatus `xml:"test_status,attr"`
	TestResult         TestResult `xml:"result,attr"`
	StartedAt          *time.Time `xml:"started_at,omitempty"`
	PassedAt           *time.Time `xml:"passed_at,omitempty"`
	FailedAt           *time.Time `xml:"failed_at,omitempty"`
	CancelledAt        *time.Time `xml:"cancelled_at,omitempty"`
	FailureNote        string     `xml:"failure_note,omitempty"`
	CancellationReason string     `xml:"cancellation_reason,omitempty"`
}

// Epic 13 Status System Methods for Test

// GetTestStatusUnified returns the unified Epic 13 test status
func (t *Test) GetTestStatusUnified() TestStatus {
	// If TestStatus is already set (new format), use it
	if t.TestStatus.IsValid() {
		return t.TestStatus
	}
	// Otherwise convert from legacy Status field
	switch t.Status {
	case StatusPending:
		return TestStatusPending
	case StatusActive:
		return TestStatusWIP
	case StatusCompleted:
		return TestStatusDone
	case StatusCancelled:
		return TestStatusCancelled
	default:
		return TestStatusPending
	}
}

// SetTestStatusUnified sets the test status using the Epic 13 unified system
func (t *Test) SetTestStatusUnified(status TestStatus) {
	t.TestStatus = status
	// Also update legacy Status field for backwards compatibility
	switch status {
	case TestStatusPending:
		t.Status = StatusPending
	case TestStatusWIP:
		t.Status = StatusActive
	case TestStatusDone:
		t.Status = StatusCompleted
	case TestStatusCancelled:
		t.Status = StatusCancelled
	}
}

// GetTestResult returns the test result with backwards compatibility
func (t *Test) GetTestResult() TestResult {
	// If TestResult is already set (new format), use it
	if t.TestResult.IsValid() {
		return t.TestResult
	}
	// For backwards compatibility, infer from status
	// WIP tests with failure timestamps are considered failing
	if t.GetTestStatusUnified() == TestStatusWIP && t.FailedAt != nil {
		return TestResultFailing
	}
	// Done tests are considered passing
	if t.GetTestStatusUnified() == TestStatusDone {
		return TestResultPassing
	}
	// Default for pending/cancelled tests
	return TestResultPassing
}

// SetTestResult sets the test result using the Epic 13 unified system
func (t *Test) SetTestResult(result TestResult) {
	t.TestResult = result
}

type Event struct {
	ID        string    `xml:"id,attr"`
	Type      string    `xml:"type,attr"`
	Timestamp time.Time `xml:"timestamp,attr"`
	Data      string    `xml:"data"`
}

func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusActive, StatusCompleted, StatusOnHold, StatusCancelled:
		return true
	default:
		return false
	}
}

// Legacy migration helpers for Epic 13 status system - DEPRECATED
// Use TestStatusDone with TestResultPassing instead of TestStatusPassed()
func TestStatusPassed() TestStatus {
	return TestStatusDone
}

// Use TestStatusWIP with TestResultFailing instead of TestStatusFailed()
func TestStatusFailed() TestStatus {
	return TestStatusWIP
}

// Epic 13 Migration Functions - Convert between old and new status systems

// ToEpicStatus converts legacy Status to Epic 13 EpicStatus
func (s Status) ToEpicStatus() EpicStatus {
	switch s {
	case StatusPending:
		return EpicStatusPending
	case StatusActive:
		return EpicStatusWIP
	case StatusCompleted:
		return EpicStatusDone
	default:
		return EpicStatusPending
	}
}

// ToPhaseStatus converts legacy Status to Epic 13 PhaseStatus
func (s Status) ToPhaseStatus() PhaseStatus {
	switch s {
	case StatusPending:
		return PhaseStatusPending
	case StatusActive:
		return PhaseStatusWIP
	case StatusCompleted:
		return PhaseStatusDone
	default:
		return PhaseStatusPending
	}
}

// ToTaskStatus converts legacy Status to Epic 13 TaskStatus
func (s Status) ToTaskStatus() TaskStatus {
	switch s {
	case StatusPending:
		return TaskStatusPending
	case StatusActive:
		return TaskStatusWIP
	case StatusCompleted:
		return TaskStatusDone
	case StatusCancelled:
		return TaskStatusCancelled
	default:
		return TaskStatusPending
	}
}

// FromEpicStatus converts Epic 13 EpicStatus back to legacy Status
func FromEpicStatus(s EpicStatus) Status {
	switch s {
	case EpicStatusPending:
		return StatusPending
	case EpicStatusWIP:
		return StatusActive
	case EpicStatusDone:
		return StatusCompleted
	default:
		return StatusPending
	}
}

// FromPhaseStatus converts Epic 13 PhaseStatus back to legacy Status
func FromPhaseStatus(s PhaseStatus) Status {
	switch s {
	case PhaseStatusPending:
		return StatusPending
	case PhaseStatusWIP:
		return StatusActive
	case PhaseStatusDone:
		return StatusCompleted
	default:
		return StatusPending
	}
}

// FromTaskStatus converts Epic 13 TaskStatus back to legacy Status
func FromTaskStatus(s TaskStatus) Status {
	switch s {
	case TaskStatusPending:
		return StatusPending
	case TaskStatusWIP:
		return StatusActive
	case TaskStatusDone:
		return StatusCompleted
	case TaskStatusCancelled:
		return StatusCancelled
	default:
		return StatusPending
	}
}

// GetTestResultFromLegacyStatus returns the appropriate TestResult for legacy passed/failed status
func GetTestResultFromLegacyStatus(legacyPassed bool) TestResult {
	if legacyPassed {
		return TestResultPassing
	}
	return TestResultFailing
}

func NewEpic(id, name string) *Epic {
	now := time.Now()
	return &Epic{
		ID:        id,
		Name:      name,
		Status:    StatusPending,
		CreatedAt: now,
		Metadata: &EpicMetadata{
			Created:         now,
			Assignee:        "",
			EstimatedEffort: "",
		},
		CurrentState: &CurrentState{
			ActivePhase: "",
			ActiveTask:  "",
			NextAction:  "Start next phase",
		},
	}
}
