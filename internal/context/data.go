package context

import (
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
)

// ProgressSummary represents progress information for a phase or context
type ProgressSummary struct {
	TotalTasks             int `json:"total_tasks" xml:"total_tasks"`
	CompletedTasks         int `json:"completed_tasks" xml:"completed_tasks"`
	ActiveTasks            int `json:"active_tasks" xml:"active_tasks"`
	PendingTasks           int `json:"pending_tasks" xml:"pending_tasks"`
	CancelledTasks         int `json:"cancelled_tasks" xml:"cancelled_tasks"`
	CompletionPercentage   int `json:"completion_percentage" xml:"completion_percentage"`
	TotalTests             int `json:"total_tests" xml:"total_tests"`
	PassedTests            int `json:"passed_tests" xml:"passed_tests"`
	FailedTests            int `json:"failed_tests" xml:"failed_tests"`
	PendingTests           int `json:"pending_tests" xml:"pending_tests"`
	TestCoveragePercentage int `json:"test_coverage_percentage" xml:"test_coverage_percentage"`
}

// TaskDetails represents detailed task information with all fields
type TaskDetails struct {
	ID                 string      `json:"id" xml:"id,attr"`
	PhaseID            string      `json:"phase_id" xml:"phase_id,attr"`
	Name               string      `json:"name" xml:"name"`
	Description        string      `json:"description" xml:"description"`
	AcceptanceCriteria string      `json:"acceptance_criteria" xml:"acceptance_criteria"`
	Status             epic.Status `json:"status" xml:"status,attr"`
	Assignee           string      `json:"assignee" xml:"assignee,attr,omitempty"`
	StartedAt          *time.Time  `json:"started_at" xml:"started_at,omitempty"`
	CompletedAt        *time.Time  `json:"completed_at" xml:"completed_at,omitempty"`
}

// PhaseDetails represents detailed phase information with all fields
type PhaseDetails struct {
	ID           string           `json:"id" xml:"id,attr"`
	Name         string           `json:"name" xml:"name"`
	Description  string           `json:"description" xml:"description"`
	Deliverables string           `json:"deliverables" xml:"deliverables"`
	Status       epic.Status      `json:"status" xml:"status,attr"`
	StartedAt    *time.Time       `json:"started_at" xml:"started_at,omitempty"`
	CompletedAt  *time.Time       `json:"completed_at" xml:"completed_at,omitempty"`
	Progress     *ProgressSummary `json:"progress,omitempty" xml:"progress,omitempty"`
}

// TestDetails represents detailed test information with all fields
type TestDetails struct {
	ID          string          `json:"id" xml:"id,attr"`
	TaskID      string          `json:"task_id" xml:"task_id,attr"`
	PhaseID     string          `json:"phase_id" xml:"phase_id,attr"`
	Name        string          `json:"name" xml:"name"`
	Description string          `json:"description" xml:"description"`
	Status      epic.Status     `json:"status" xml:"status,attr"`
	TestStatus  epic.TestStatus `json:"test_status" xml:"test_status,attr"`
	StartedAt   *time.Time      `json:"started_at" xml:"started_at,omitempty"`
	PassedAt    *time.Time      `json:"passed_at" xml:"passed_at,omitempty"`
	FailedAt    *time.Time      `json:"failed_at" xml:"failed_at,omitempty"`
	FailureNote string          `json:"failure_note" xml:"failure_note,omitempty"`
}

// TaskWithTests represents a task with its associated tests
type TaskWithTests struct {
	TaskDetails TaskDetails   `json:"task" xml:"task"`
	Tests       []TestDetails `json:"tests" xml:"tests>test"`
}

// TaskContext represents the full context around a task
type TaskContext struct {
	TaskDetails  TaskDetails   `json:"task_details" xml:"task_details"`
	ParentPhase  *PhaseDetails `json:"parent_phase,omitempty" xml:"parent_phase,omitempty"`
	SiblingTasks []TaskDetails `json:"sibling_tasks,omitempty" xml:"sibling_tasks>task,omitempty"`
	ChildTests   []TestDetails `json:"child_tests,omitempty" xml:"child_tests>test,omitempty"`
}

// PhaseContext represents the full context around a phase
type PhaseContext struct {
	PhaseDetails    PhaseDetails     `json:"phase_details" xml:"phase_details"`
	ProgressSummary *ProgressSummary `json:"progress_summary,omitempty" xml:"progress_summary,omitempty"`
	AllTasks        []TaskWithTests  `json:"all_tasks,omitempty" xml:"all_tasks>task,omitempty"`
	PhaseTests      []TestDetails    `json:"phase_tests,omitempty" xml:"phase_tests>test,omitempty"`
	SiblingPhases   []PhaseDetails   `json:"sibling_phases,omitempty" xml:"sibling_phases>phase,omitempty"`
}

// TestContext represents the full context around a test
type TestContext struct {
	TestDetails  TestDetails   `json:"test_details" xml:"test_details"`
	ParentTask   *TaskDetails  `json:"parent_task,omitempty" xml:"parent_task,omitempty"`
	ParentPhase  *PhaseDetails `json:"parent_phase,omitempty" xml:"parent_phase,omitempty"`
	SiblingTests []TestDetails `json:"sibling_tests,omitempty" xml:"sibling_tests>test,omitempty"`
}

// ContextResult is a wrapper for all context types with additional metadata
type ContextResult struct {
	Type         string        `json:"type" xml:"type,attr"`
	EntityID     string        `json:"entity_id" xml:"entity_id,attr"`
	TaskContext  *TaskContext  `json:"task_context,omitempty" xml:"task_context,omitempty"`
	PhaseContext *PhaseContext `json:"phase_context,omitempty" xml:"phase_context,omitempty"`
	TestContext  *TestContext  `json:"test_context,omitempty" xml:"test_context,omitempty"`
	FullDetails  bool          `json:"full_details" xml:"full_details,attr"`
}
