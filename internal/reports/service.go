package reports

import (
	"fmt"
	"strings"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/storage"
)

type ReportService struct {
	storage storage.Storage
	epic    *epic.Epic
}

func NewReportService(storage storage.Storage) *ReportService {
	return &ReportService{
		storage: storage,
	}
}

func (rs *ReportService) LoadEpic(filename string) error {
	epic, err := rs.storage.LoadEpic(filename)
	if err != nil {
		return fmt.Errorf("failed to load epic: %w", err)
	}
	rs.epic = epic
	return nil
}

type HandoffReport struct {
	EpicInfo     EpicInfo     `xml:"epic_info"`
	CurrentState CurrentState `xml:"current_state"`
	Summary      Summary      `xml:"summary"`
	RecentEvents []Event      `xml:"recent_events>event"`
	Blockers     []string     `xml:"blockers>blocker"`
	GeneratedAt  time.Time    `xml:"generated_at,attr"`
}

type EpicInfo struct {
	ID       string    `xml:"id,attr"`
	Name     string    `xml:"name"`
	Status   string    `xml:"status"`
	Started  time.Time `xml:"started"`
	Assignee string    `xml:"assignee"`
}

type CurrentState struct {
	ActivePhase string `xml:"active_phase"`
	ActiveTask  string `xml:"active_task"`
	NextAction  string `xml:"next_action"`
}

type Summary struct {
	CompletedPhases      int `xml:"completed_phases"`
	TotalPhases          int `xml:"total_phases"`
	PassingTests         int `xml:"passing_tests"`
	FailingTests         int `xml:"failing_tests"`
	CompletionPercentage int `xml:"completion_percentage"`
}

type Event struct {
	Timestamp time.Time `xml:"timestamp,attr"`
	Type      string    `xml:"type,attr"`
	PhaseID   string    `xml:"phase_id,attr,omitempty"`
	Data      string    `xml:",chardata"`
}

func (rs *ReportService) GenerateHandoffReport(limit int) (*HandoffReport, error) {
	if rs.epic == nil {
		return nil, fmt.Errorf("no epic loaded")
	}

	report := &HandoffReport{
		GeneratedAt: time.Now(),
	}

	// Extract epic info
	report.EpicInfo = EpicInfo{
		ID:       rs.epic.ID,
		Name:     rs.epic.Name,
		Status:   string(rs.epic.Status),
		Started:  rs.epic.CreatedAt,
		Assignee: rs.epic.Assignee,
	}

	// Extract current state
	report.CurrentState = CurrentState{
		ActivePhase: rs.findActivePhase(),
		ActiveTask:  rs.findActiveTask(),
		NextAction:  rs.determineNextAction(),
	}

	// Calculate summary
	report.Summary = rs.calculateSummary()

	// Get recent events
	report.RecentEvents = rs.getRecentEvents(limit)

	// Identify blockers
	report.Blockers = rs.identifyBlockers()

	return report, nil
}

func (rs *ReportService) findActivePhase() string {
	for _, phase := range rs.epic.Phases {
		if phase.Status == epic.StatusActive {
			return phase.ID
		}
	}
	return ""
}

func (rs *ReportService) findActiveTask() string {
	for _, task := range rs.epic.Tasks {
		if task.Status == epic.StatusActive {
			return task.ID
		}
	}
	return ""
}

func (rs *ReportService) determineNextAction() string {
	// Find the first pending task in the active phase
	activePhase := rs.findActivePhase()
	if activePhase == "" {
		return "Start next phase"
	}

	for _, task := range rs.epic.Tasks {
		if task.PhaseID == activePhase && task.Status == epic.StatusPending {
			return fmt.Sprintf("Start task %s: %s", task.ID, task.Name)
		}
	}

	return "Continue current work"
}

func (rs *ReportService) calculateSummary() Summary {
	summary := Summary{
		TotalPhases: len(rs.epic.Phases),
	}

	// Count completed phases
	for _, phase := range rs.epic.Phases {
		if phase.Status == epic.StatusCompleted {
			summary.CompletedPhases++
		}
	}

	// Count test status
	for _, test := range rs.epic.Tests {
		if test.Status == epic.StatusCompleted {
			summary.PassingTests++
		} else {
			summary.FailingTests++
		}
	}

	// Calculate completion percentage using weighted approach
	summary.CompletionPercentage = rs.calculateWeightedCompletion()

	return summary
}

func (rs *ReportService) calculateWeightedCompletion() int {
	totalPhases := len(rs.epic.Phases)
	totalTasks := len(rs.epic.Tasks)
	totalTests := len(rs.epic.Tests)

	if totalPhases == 0 && totalTasks == 0 && totalTests == 0 {
		return 0
	}

	phaseWeight := 40.0
	taskWeight := 40.0
	testWeight := 20.0

	var phaseCompletion, taskCompletion, testCompletion float64

	// Calculate phase completion
	if totalPhases > 0 {
		completedPhases := 0
		for _, phase := range rs.epic.Phases {
			if phase.Status == epic.StatusCompleted {
				completedPhases++
			}
		}
		phaseCompletion = float64(completedPhases) / float64(totalPhases)
	}

	// Calculate task completion
	if totalTasks > 0 {
		completedTasks := 0
		for _, task := range rs.epic.Tasks {
			if task.Status == epic.StatusCompleted {
				completedTasks++
			}
		}
		taskCompletion = float64(completedTasks) / float64(totalTasks)
	}

	// Calculate test completion
	if totalTests > 0 {
		completedTests := 0
		for _, test := range rs.epic.Tests {
			if test.Status == epic.StatusCompleted {
				completedTests++
			}
		}
		testCompletion = float64(completedTests) / float64(totalTests)
	}

	// Weight completion percentages
	weightedCompletion := (phaseCompletion*phaseWeight +
		taskCompletion*taskWeight +
		testCompletion*testWeight) / 100.0

	return int(weightedCompletion * 100)
}

func (rs *ReportService) getRecentEvents(limit int) []Event {
	events := make([]Event, 0)

	// Get events in reverse chronological order
	epicEvents := rs.epic.Events
	start := len(epicEvents) - limit
	if start < 0 {
		start = 0
	}

	for i := len(epicEvents) - 1; i >= start; i-- {
		event := epicEvents[i]
		events = append(events, Event{
			Timestamp: event.Timestamp,
			Type:      event.Type,
			Data:      event.Data,
		})
	}

	return events
}

func (rs *ReportService) identifyBlockers() []string {
	blockers := make([]string, 0)

	// Find failed tests
	for _, test := range rs.epic.Tests {
		if test.TestStatus == epic.TestStatusWIP {
			blockers = append(blockers, fmt.Sprintf("Failed test %s: %s", test.ID, test.Name))
		}
	}

	// Find blocker events
	for _, event := range rs.epic.Events {
		if event.Type == "blocker" {
			blockers = append(blockers, event.Data)
		}
	}

	return blockers
}

// DocumentationReport represents a human-readable documentation report
type DocumentationReport struct {
	EpicOverview   EpicOverview   `json:"epic_overview"`
	PhaseProgress  PhaseProgress  `json:"phase_progress"`
	TaskStatus     TaskStatus     `json:"task_status"`
	TestResults    TestResults    `json:"test_results"`
	RecentActivity RecentActivity `json:"recent_activity"`
	GeneratedAt    time.Time      `json:"generated_at"`
}

type EpicOverview struct {
	Name        string    `json:"name"`
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	Assignee    string    `json:"assignee"`
	Started     time.Time `json:"started"`
	Completion  int       `json:"completion_percentage"`
	Description string    `json:"description,omitempty"`
}

type PhaseProgress struct {
	TotalPhases     int           `json:"total_phases"`
	CompletedPhases int           `json:"completed_phases"`
	Phases          []PhaseDetail `json:"phases"`
}

type PhaseDetail struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	TaskCount   int        `json:"task_count"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type TaskStatus struct {
	TotalTasks     int          `json:"total_tasks"`
	CompletedTasks int          `json:"completed_tasks"`
	ActiveTask     string       `json:"active_task,omitempty"`
	Tasks          []TaskDetail `json:"tasks"`
}

type TaskDetail struct {
	ID          string     `json:"id"`
	PhaseID     string     `json:"phase_id"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	Assignee    string     `json:"assignee,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type TestResults struct {
	TotalTests   int          `json:"total_tests"`
	PassingTests int          `json:"passing_tests"`
	FailingTests int          `json:"failing_tests"`
	Tests        []TestDetail `json:"tests"`
}

type TestDetail struct {
	ID          string `json:"id"`
	TaskID      string `json:"task_id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	TestStatus  string `json:"test_status"`
	FailureNote string `json:"failure_note,omitempty"`
}

type RecentActivity struct {
	Events   []Event  `json:"events"`
	Blockers []string `json:"blockers"`
}

func (rs *ReportService) GenerateDocumentationReport() (*DocumentationReport, error) {
	if rs.epic == nil {
		return nil, fmt.Errorf("no epic loaded")
	}

	report := &DocumentationReport{
		GeneratedAt: time.Now(),
	}

	// Epic Overview
	report.EpicOverview = EpicOverview{
		Name:        rs.epic.Name,
		ID:          rs.epic.ID,
		Status:      string(rs.epic.Status),
		Assignee:    rs.epic.Assignee,
		Started:     rs.epic.CreatedAt,
		Completion:  rs.calculateWeightedCompletion(),
		Description: rs.epic.Description,
	}

	// Phase Progress
	report.PhaseProgress = rs.generatePhaseProgress()

	// Task Status
	report.TaskStatus = rs.generateTaskStatus()

	// Test Results
	report.TestResults = rs.generateTestResults()

	// Recent Activity
	report.RecentActivity = RecentActivity{
		Events:   rs.getRecentEvents(10), // Get more events for documentation
		Blockers: rs.identifyBlockers(),
	}

	return report, nil
}

func (rs *ReportService) generatePhaseProgress() PhaseProgress {
	progress := PhaseProgress{
		TotalPhases: len(rs.epic.Phases),
		Phases:      make([]PhaseDetail, 0, len(rs.epic.Phases)),
	}

	for _, phase := range rs.epic.Phases {
		if phase.Status == epic.StatusCompleted {
			progress.CompletedPhases++
		}

		// Count tasks in this phase
		taskCount := 0
		for _, task := range rs.epic.Tasks {
			if task.PhaseID == phase.ID {
				taskCount++
			}
		}

		phaseDetail := PhaseDetail{
			ID:          phase.ID,
			Name:        phase.Name,
			Status:      string(phase.Status),
			TaskCount:   taskCount,
			StartedAt:   phase.StartedAt,
			CompletedAt: phase.CompletedAt,
		}

		progress.Phases = append(progress.Phases, phaseDetail)
	}

	return progress
}

func (rs *ReportService) generateTaskStatus() TaskStatus {
	status := TaskStatus{
		TotalTasks: len(rs.epic.Tasks),
		Tasks:      make([]TaskDetail, 0, len(rs.epic.Tasks)),
	}

	for _, task := range rs.epic.Tasks {
		if task.Status == epic.StatusCompleted {
			status.CompletedTasks++
		}
		if task.Status == epic.StatusActive {
			status.ActiveTask = task.ID
		}

		taskDetail := TaskDetail{
			ID:          task.ID,
			PhaseID:     task.PhaseID,
			Name:        task.Name,
			Status:      string(task.Status),
			Assignee:    task.Assignee,
			StartedAt:   task.StartedAt,
			CompletedAt: task.CompletedAt,
		}

		status.Tasks = append(status.Tasks, taskDetail)
	}

	return status
}

func (rs *ReportService) generateTestResults() TestResults {
	results := TestResults{
		TotalTests: len(rs.epic.Tests),
		Tests:      make([]TestDetail, 0, len(rs.epic.Tests)),
	}

	for _, test := range rs.epic.Tests {
		if test.TestStatus == epic.TestStatusDone {
			results.PassingTests++
		} else {
			results.FailingTests++
		}

		testDetail := TestDetail{
			ID:          test.ID,
			TaskID:      test.TaskID,
			Name:        test.Name,
			Status:      string(test.Status),
			TestStatus:  string(test.TestStatus),
			FailureNote: test.FailureNote,
		}

		results.Tests = append(results.Tests, testDetail)
	}

	return results
}

func (rs *ReportService) GenerateMarkdownDocumentation() (string, error) {
	report, err := rs.GenerateDocumentationReport()
	if err != nil {
		return "", err
	}

	return rs.formatMarkdown(report), nil
}

func (rs *ReportService) formatMarkdown(report *DocumentationReport) string {
	var md strings.Builder

	// Title and overview
	md.WriteString(fmt.Sprintf("# %s\n\n", report.EpicOverview.Name))
	md.WriteString("## Epic Overview\n\n")
	md.WriteString(fmt.Sprintf("- **ID:** %s\n", report.EpicOverview.ID))
	md.WriteString(fmt.Sprintf("- **Status:** %s\n", rs.formatStatusIcon(report.EpicOverview.Status)))
	md.WriteString(fmt.Sprintf("- **Assignee:** %s\n", report.EpicOverview.Assignee))
	md.WriteString(fmt.Sprintf("- **Started:** %s\n", report.EpicOverview.Started.Format("2006-01-02 15:04:05")))
	md.WriteString(fmt.Sprintf("- **Progress:** %d%% complete\n\n", report.EpicOverview.Completion))

	if report.EpicOverview.Description != "" {
		md.WriteString(fmt.Sprintf("**Description:** %s\n\n", report.EpicOverview.Description))
	}

	// Phase Progress
	md.WriteString("## Phase Progress\n\n")
	md.WriteString(fmt.Sprintf("**Completed:** %d/%d phases\n\n",
		report.PhaseProgress.CompletedPhases, report.PhaseProgress.TotalPhases))

	md.WriteString("| Phase | Status | Tasks | Started | Completed |\n")
	md.WriteString("|-------|--------|-------|---------|----------|\n")

	for _, phase := range report.PhaseProgress.Phases {
		started := "—"
		if phase.StartedAt != nil {
			started = phase.StartedAt.Format("2006-01-02")
		}
		completed := "—"
		if phase.CompletedAt != nil {
			completed = phase.CompletedAt.Format("2006-01-02")
		}

		md.WriteString(fmt.Sprintf("| %s | %s | %d | %s | %s |\n",
			phase.Name, rs.formatStatusIcon(phase.Status), phase.TaskCount, started, completed))
	}
	md.WriteString("\n")

	// Task Status
	md.WriteString("## Task Status\n\n")
	md.WriteString(fmt.Sprintf("**Completed:** %d/%d tasks\n",
		report.TaskStatus.CompletedTasks, report.TaskStatus.TotalTasks))
	if report.TaskStatus.ActiveTask != "" {
		md.WriteString(fmt.Sprintf("**Active Task:** %s\n", report.TaskStatus.ActiveTask))
	}
	md.WriteString("\n")

	md.WriteString("| Task | Phase | Status | Assignee | Started | Completed |\n")
	md.WriteString("|------|--------|--------|----------|---------|----------|\n")

	for _, task := range report.TaskStatus.Tasks {
		started := "—"
		if task.StartedAt != nil {
			started = task.StartedAt.Format("2006-01-02")
		}
		completed := "—"
		if task.CompletedAt != nil {
			completed = task.CompletedAt.Format("2006-01-02")
		}
		assignee := task.Assignee
		if assignee == "" {
			assignee = "—"
		}

		md.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			task.Name, task.PhaseID, rs.formatStatusIcon(task.Status), assignee, started, completed))
	}
	md.WriteString("\n")

	// Test Results
	md.WriteString("## Test Results\n\n")
	md.WriteString(fmt.Sprintf("**Summary:** %d passing, %d failing (%d total)\n\n",
		report.TestResults.PassingTests, report.TestResults.FailingTests, report.TestResults.TotalTests))

	if len(report.TestResults.Tests) > 0 {
		md.WriteString("| Test | Task | Status | Notes |\n")
		md.WriteString("|------|------|--------|---------|\n")

		for _, test := range report.TestResults.Tests {
			notes := test.FailureNote
			if notes == "" {
				notes = "—"
			}

			md.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				test.Name, test.TaskID, rs.formatTestStatusWithContext(test), notes))
		}
		md.WriteString("\n")
	}

	// Blockers
	if len(report.RecentActivity.Blockers) > 0 {
		md.WriteString("## Blockers\n\n")
		for _, blocker := range report.RecentActivity.Blockers {
			md.WriteString(fmt.Sprintf("- 🚫 %s\n", blocker))
		}
		md.WriteString("\n")
	}

	// Recent Activity
	if len(report.RecentActivity.Events) > 0 {
		md.WriteString("## Recent Activity\n\n")
		for _, event := range report.RecentActivity.Events {
			md.WriteString(fmt.Sprintf("- **%s** (%s): %s\n",
				event.Timestamp.Format("2006-01-02 15:04"), event.Type, event.Data))
		}
		md.WriteString("\n")
	}

	// Footer
	md.WriteString("---\n")
	md.WriteString(fmt.Sprintf("*Generated on %s by AgentPM*\n",
		report.GeneratedAt.Format("2006-01-02 15:04:05")))

	return md.String()
}

func (rs *ReportService) formatStatusIcon(status string) string {
	switch status {
	case "completed":
		return "✅ completed"
	case "active":
		return "🔄 active"
	case "planning":
		return "⏳ planning"
	case "on_hold":
		return "⏸️ on hold"
	case "cancelled":
		return "❌ cancelled"
	default:
		return status
	}
}

func (rs *ReportService) formatTestStatusIcon(status string) string {
	switch status {
	// Epic 13 unified status system
	case "done":
		return "✅ passed"
	case "wip":
		return "🔄 in progress"
	case "pending":
		return "⏳ pending"
	case "cancelled":
		return "⏹️ cancelled"
	// Legacy status support
	case "passed":
		return "✅ passed"
	case "failed":
		return "❌ failed"
	default:
		return status
	}
}

// formatTestStatusWithContext determines test status considering Epic 13 status system
func (rs *ReportService) formatTestStatusWithContext(test TestDetail) string {
	// Epic 13 logic: If test has failure note and is wip, it's failed
	if test.FailureNote != "" && test.TestStatus == "wip" {
		return "❌ failed"
	}

	// Fall back to status-based formatting
	return rs.formatTestStatusIcon(test.TestStatus)
}
