package epic

import "fmt"

// EpicStatus represents the status of an epic
type EpicStatus string

const (
	EpicStatusPending EpicStatus = "pending"
	EpicStatusWIP     EpicStatus = "wip"
	EpicStatusDone    EpicStatus = "done"
)

// IsValid checks if the epic status is valid
func (s EpicStatus) IsValid() bool {
	switch s {
	case EpicStatusPending, EpicStatusWIP, EpicStatusDone:
		return true
	default:
		return false
	}
}

// String returns the string representation of the epic status
func (s EpicStatus) String() string {
	return string(s)
}

// PhaseStatus represents the status of a phase
type PhaseStatus string

const (
	PhaseStatusPending PhaseStatus = "pending"
	PhaseStatusWIP     PhaseStatus = "wip"
	PhaseStatusDone    PhaseStatus = "done"
)

// IsValid checks if the phase status is valid
func (s PhaseStatus) IsValid() bool {
	switch s {
	case PhaseStatusPending, PhaseStatusWIP, PhaseStatusDone:
		return true
	default:
		return false
	}
}

// String returns the string representation of the phase status
func (s PhaseStatus) String() string {
	return string(s)
}

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusWIP       TaskStatus = "wip"
	TaskStatusDone      TaskStatus = "done"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// IsValid checks if the task status is valid
func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskStatusPending, TaskStatusWIP, TaskStatusDone, TaskStatusCancelled:
		return true
	default:
		return false
	}
}

// String returns the string representation of the task status
func (s TaskStatus) String() string {
	return string(s)
}

// TestResult represents the result of a test
type TestResult string

const (
	TestResultPassing TestResult = "passing"
	TestResultFailing TestResult = "failing"
)

// IsValid checks if the test result is valid
func (s TestResult) IsValid() bool {
	switch s {
	case TestResultPassing, TestResultFailing:
		return true
	default:
		return false
	}
}

// String returns the string representation of the test result
func (s TestResult) String() string {
	return string(s)
}

// Status transition validation methods

// CanTransitionTo checks if an epic status can transition to another status
func (s EpicStatus) CanTransitionTo(target EpicStatus) bool {
	transitions := map[EpicStatus][]EpicStatus{
		EpicStatusPending: {EpicStatusWIP},
		EpicStatusWIP:     {EpicStatusDone},
		EpicStatusDone:    {}, // Done is terminal
	}

	for _, allowed := range transitions[s] {
		if allowed == target {
			return true
		}
	}
	return false
}

// CanTransitionTo checks if a phase status can transition to another status
func (s PhaseStatus) CanTransitionTo(target PhaseStatus) bool {
	transitions := map[PhaseStatus][]PhaseStatus{
		PhaseStatusPending: {PhaseStatusWIP},
		PhaseStatusWIP:     {PhaseStatusDone},
		PhaseStatusDone:    {}, // Done is terminal
	}

	for _, allowed := range transitions[s] {
		if allowed == target {
			return true
		}
	}
	return false
}

// CanTransitionTo checks if a task status can transition to another status
func (s TaskStatus) CanTransitionTo(target TaskStatus) bool {
	transitions := map[TaskStatus][]TaskStatus{
		TaskStatusPending:   {TaskStatusWIP, TaskStatusCancelled},
		TaskStatusWIP:       {TaskStatusDone, TaskStatusCancelled},
		TaskStatusDone:      {}, // Done is terminal
		TaskStatusCancelled: {}, // Cancelled is terminal
	}

	for _, allowed := range transitions[s] {
		if allowed == target {
			return true
		}
	}
	return false
}

// ValidateEpicStatus validates an epic status string and returns the typed status
func ValidateEpicStatus(status string) (EpicStatus, error) {
	s := EpicStatus(status)
	if !s.IsValid() {
		return "", fmt.Errorf("invalid epic status: %s", status)
	}
	return s, nil
}

// ValidatePhaseStatus validates a phase status string and returns the typed status
func ValidatePhaseStatus(status string) (PhaseStatus, error) {
	s := PhaseStatus(status)
	if !s.IsValid() {
		return "", fmt.Errorf("invalid phase status: %s", status)
	}
	return s, nil
}

// ValidateTaskStatus validates a task status string and returns the typed status
func ValidateTaskStatus(status string) (TaskStatus, error) {
	s := TaskStatus(status)
	if !s.IsValid() {
		return "", fmt.Errorf("invalid task status: %s", status)
	}
	return s, nil
}

// ValidateTestResult validates a test result string and returns the typed result
func ValidateTestResult(result string) (TestResult, error) {
	r := TestResult(result)
	if !r.IsValid() {
		return "", fmt.Errorf("invalid test result: %s", result)
	}
	return r, nil
}
