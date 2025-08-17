package epic

import (
	"time"
)

type Status string

const (
	StatusPlanning  Status = "planning"
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

type Event struct {
	ID        string    `xml:"id,attr"`
	Type      string    `xml:"type,attr"`
	Timestamp time.Time `xml:"timestamp,attr"`
	Data      string    `xml:"data"`
}

func (s Status) IsValid() bool {
	switch s {
	case StatusPlanning, StatusPending, StatusActive, StatusCompleted, StatusOnHold, StatusCancelled:
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
		Status:    StatusPlanning,
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
