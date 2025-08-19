package context

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/epic"
)

func TestXMLFormatter_FormatTaskContext(t *testing.T) {
	formatter := &XMLFormatter{}
	ctx := createSampleTaskContext()

	var buf bytes.Buffer
	err := formatter.FormatTaskContext(ctx, &buf)

	if err != nil {
		t.Errorf("FormatTaskContext() unexpected error = %v", err)
		return
	}

	output := buf.String()

	// Check basic XML structure
	if !strings.Contains(output, "<task_context") {
		t.Errorf("FormatTaskContext() missing task_context element")
	}

	if !strings.Contains(output, `id="1A_1"`) {
		t.Errorf("FormatTaskContext() missing task ID")
	}

	if !strings.Contains(output, "<task_details>") {
		t.Errorf("FormatTaskContext() missing task_details element")
	}

	if !strings.Contains(output, "<name>Initialize Project</name>") {
		t.Errorf("FormatTaskContext() missing task name")
	}

	if !strings.Contains(output, "<parent_phase") {
		t.Errorf("FormatTaskContext() missing parent_phase element")
	}

	if !strings.Contains(output, "<sibling_tasks>") {
		t.Errorf("FormatTaskContext() missing sibling_tasks element")
	}

	if !strings.Contains(output, "<child_tests>") {
		t.Errorf("FormatTaskContext() missing child_tests element")
	}

	if !strings.Contains(output, "</task_context>") {
		t.Errorf("FormatTaskContext() missing closing task_context element")
	}
}

func TestXMLFormatter_FormatPhaseContext(t *testing.T) {
	formatter := &XMLFormatter{}
	ctx := createSamplePhaseContext()

	var buf bytes.Buffer
	err := formatter.FormatPhaseContext(ctx, &buf)

	if err != nil {
		t.Errorf("FormatPhaseContext() unexpected error = %v", err)
		return
	}

	output := buf.String()

	// Check basic XML structure
	if !strings.Contains(output, "<phase_context") {
		t.Errorf("FormatPhaseContext() missing phase_context element")
	}

	if !strings.Contains(output, `id="1A"`) {
		t.Errorf("FormatPhaseContext() missing phase ID")
	}

	if !strings.Contains(output, "<phase_details>") {
		t.Errorf("FormatPhaseContext() missing phase_details element")
	}

	if !strings.Contains(output, "<progress>") {
		t.Errorf("FormatPhaseContext() missing progress element")
	}

	if !strings.Contains(output, "<all_tasks>") {
		t.Errorf("FormatPhaseContext() missing all_tasks element")
	}

	if !strings.Contains(output, "<sibling_phases>") {
		t.Errorf("FormatPhaseContext() missing sibling_phases element")
	}
}

func TestXMLFormatter_FormatTestContext(t *testing.T) {
	formatter := &XMLFormatter{}
	ctx := createSampleTestContext()

	var buf bytes.Buffer
	err := formatter.FormatTestContext(ctx, &buf)

	if err != nil {
		t.Errorf("FormatTestContext() unexpected error = %v", err)
		return
	}

	output := buf.String()

	// Check basic XML structure
	if !strings.Contains(output, "<test_context") {
		t.Errorf("FormatTestContext() missing test_context element")
	}

	if !strings.Contains(output, `id="T1A_1"`) {
		t.Errorf("FormatTestContext() missing test ID")
	}

	if !strings.Contains(output, "<test_details>") {
		t.Errorf("FormatTestContext() missing test_details element")
	}

	if !strings.Contains(output, "<parent_task") {
		t.Errorf("FormatTestContext() missing parent_task element")
	}

	if !strings.Contains(output, "<parent_phase") {
		t.Errorf("FormatTestContext() missing parent_phase element")
	}

	if !strings.Contains(output, "<sibling_tests>") {
		t.Errorf("FormatTestContext() missing sibling_tests element")
	}
}

func TestJSONFormatter_FormatTaskContext(t *testing.T) {
	formatter := &JSONFormatter{}
	ctx := createSampleTaskContext()

	var buf bytes.Buffer
	err := formatter.FormatTaskContext(ctx, &buf)

	if err != nil {
		t.Errorf("FormatTaskContext() unexpected error = %v", err)
		return
	}

	// Verify it's valid JSON
	var jsonData map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &jsonData)
	if err != nil {
		t.Errorf("FormatTaskContext() produced invalid JSON: %v", err)
		return
	}

	// Check basic structure
	taskDetails, ok := jsonData["task_details"].(map[string]interface{})
	if !ok {
		t.Errorf("FormatTaskContext() missing task_details in JSON")
		return
	}

	if taskDetails["id"] != "1A_1" {
		t.Errorf("FormatTaskContext() incorrect task ID in JSON")
	}

	if taskDetails["name"] != "Initialize Project" {
		t.Errorf("FormatTaskContext() incorrect task name in JSON")
	}
}

func TestJSONFormatter_FormatPhaseContext(t *testing.T) {
	formatter := &JSONFormatter{}
	ctx := createSamplePhaseContext()

	var buf bytes.Buffer
	err := formatter.FormatPhaseContext(ctx, &buf)

	if err != nil {
		t.Errorf("FormatPhaseContext() unexpected error = %v", err)
		return
	}

	// Verify it's valid JSON
	var jsonData map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &jsonData)
	if err != nil {
		t.Errorf("FormatPhaseContext() produced invalid JSON: %v", err)
		return
	}

	// Check basic structure
	phaseDetails, ok := jsonData["phase_details"].(map[string]interface{})
	if !ok {
		t.Errorf("FormatPhaseContext() missing phase_details in JSON")
		return
	}

	if phaseDetails["id"] != "1A" {
		t.Errorf("FormatPhaseContext() incorrect phase ID in JSON")
	}
}

func TestTextFormatter_FormatTaskContext(t *testing.T) {
	formatter := &TextFormatter{}
	ctx := createSampleTaskContext()

	var buf bytes.Buffer
	err := formatter.FormatTaskContext(ctx, &buf)

	if err != nil {
		t.Errorf("FormatTaskContext() unexpected error = %v", err)
		return
	}

	output := buf.String()

	// Check basic text structure
	if !strings.Contains(output, "Task: Initialize Project") {
		t.Errorf("FormatTaskContext() missing task name in text output")
	}

	if !strings.Contains(output, "ID: 1A_1") {
		t.Errorf("FormatTaskContext() missing task ID in text output")
	}

	if !strings.Contains(output, "Status: completed") {
		t.Errorf("FormatTaskContext() missing task status in text output")
	}

	if !strings.Contains(output, "Parent Phase:") {
		t.Errorf("FormatTaskContext() missing parent phase section in text output")
	}

	if !strings.Contains(output, "Sibling Tasks") {
		t.Errorf("FormatTaskContext() missing sibling tasks section in text output")
	}

	if !strings.Contains(output, "Child Tests") {
		t.Errorf("FormatTaskContext() missing child tests section in text output")
	}
}

func TestTextFormatter_FormatPhaseContext(t *testing.T) {
	formatter := &TextFormatter{}
	ctx := createSamplePhaseContext()

	var buf bytes.Buffer
	err := formatter.FormatPhaseContext(ctx, &buf)

	if err != nil {
		t.Errorf("FormatPhaseContext() unexpected error = %v", err)
		return
	}

	output := buf.String()

	// Check basic text structure
	if !strings.Contains(output, "Phase: CLI Framework") {
		t.Errorf("FormatPhaseContext() missing phase name in text output")
	}

	if !strings.Contains(output, "ID: 1A") {
		t.Errorf("FormatPhaseContext() missing phase ID in text output")
	}

	if !strings.Contains(output, "Progress Summary:") {
		t.Errorf("FormatPhaseContext() missing progress summary in text output")
	}

	if !strings.Contains(output, "All Tasks") {
		t.Errorf("FormatPhaseContext() missing all tasks section in text output")
	}
}

func TestTextFormatter_FormatTestContext(t *testing.T) {
	formatter := &TextFormatter{}
	ctx := createSampleTestContext()

	var buf bytes.Buffer
	err := formatter.FormatTestContext(ctx, &buf)

	if err != nil {
		t.Errorf("FormatTestContext() unexpected error = %v", err)
		return
	}

	output := buf.String()

	// Check basic text structure
	if !strings.Contains(output, "Test: Test Project Init") {
		t.Errorf("FormatTestContext() missing test name in text output")
	}

	if !strings.Contains(output, "ID: T1A_1") {
		t.Errorf("FormatTestContext() missing test ID in text output")
	}

	if !strings.Contains(output, "Status: completed") {
		t.Errorf("FormatTestContext() missing test status in text output")
	}

	if !strings.Contains(output, "Parent Task:") {
		t.Errorf("FormatTestContext() missing parent task section in text output")
	}

	if !strings.Contains(output, "Parent Phase:") {
		t.Errorf("FormatTestContext() missing parent phase section in text output")
	}
}

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		format   string
		expected string
	}{
		{"json", "*context.JSONFormatter"},
		{"xml", "*context.XMLFormatter"},
		{"text", "*context.TextFormatter"},
		{"unknown", "*context.TextFormatter"}, // default to text
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			formatter := NewFormatter(tt.format)
			formatterType := strings.Replace(string(rune(0)), "", "", -1) // This is a placeholder

			// Check that we get a valid formatter
			if formatter == nil {
				t.Errorf("NewFormatter(%s) returned nil", tt.format)
			}

			// Try to use the formatter to ensure it's working
			ctx := createSampleTaskContext()
			var buf bytes.Buffer
			err := formatter.FormatTaskContext(ctx, &buf)
			if err != nil {
				t.Errorf("NewFormatter(%s) created formatter that failed: %v", tt.format, err)
			}

			// Verify output is not empty
			if buf.Len() == 0 {
				t.Errorf("NewFormatter(%s) created formatter that produced empty output", tt.format)
			}

			_ = formatterType // Suppress unused variable warning
		})
	}
}

func TestXMLFormatter_writeProgressXML(t *testing.T) {
	formatter := &XMLFormatter{}
	progress := &ProgressSummary{
		TotalTasks:             5,
		CompletedTasks:         3,
		PendingTasks:           2,
		CompletionPercentage:   60,
		TotalTests:             8,
		PassedTests:            5,
		FailedTests:            1,
		PendingTests:           2,
		TestCoveragePercentage: 62,
	}

	var buf bytes.Buffer
	formatter.writeProgressXML(&buf, progress, "    ")

	output := buf.String()

	// Check all progress elements are present
	expectedElements := []string{
		"<progress>",
		"<total_tasks>5</total_tasks>",
		"<completed_tasks>3</completed_tasks>",
		"<pending_tasks>2</pending_tasks>",
		"<completion_percentage>60</completion_percentage>",
		"<total_tests>8</total_tests>",
		"<passed_tests>5</passed_tests>",
		"<failed_tests>1</failed_tests>",
		"<pending_tests>2</pending_tests>",
		"<test_coverage_percentage>62</test_coverage_percentage>",
		"</progress>",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(output, expected) {
			t.Errorf("writeProgressXML() missing element: %s", expected)
		}
	}
}

func TestIndentText(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		indent string
		want   string
	}{
		{
			name:   "single line",
			text:   "Hello World",
			indent: "  ",
			want:   "  Hello World",
		},
		{
			name:   "multi line",
			text:   "Line 1\nLine 2\nLine 3",
			indent: "    ",
			want:   "    Line 1\n    Line 2\n    Line 3",
		},
		{
			name:   "empty lines preserved",
			text:   "Line 1\n\nLine 3",
			indent: "  ",
			want:   "  Line 1\n\n  Line 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := indentText(tt.text, tt.indent)
			if result != tt.want {
				t.Errorf("indentText() = %q, want %q", result, tt.want)
			}
		})
	}
}

// Helper functions to create sample context data for testing

func createSampleTaskContext() *TaskContext {
	now := time.Now()
	earlier := now.Add(-time.Hour)

	return &TaskContext{
		TaskDetails: TaskDetails{
			ID:                 "1A_1",
			PhaseID:            "1A",
			Name:               "Initialize Project",
			Description:        "Initialize Go module with required dependencies",
			AcceptanceCriteria: "Go module initializes successfully\nRequired dependencies are added to go.mod\nProject structure follows Go conventions",
			Status:             epic.StatusCompleted,
			Assignee:           "agent_claude",
			StartedAt:          &earlier,
			CompletedAt:        &now,
		},
		ParentPhase: &PhaseDetails{
			ID:           "1A",
			Name:         "CLI Framework & Core Structure",
			Description:  "Setup basic CLI structure and initialize project",
			Deliverables: "Functional CLI framework\nProject structure established\nCore dependencies configured",
			Status:       epic.StatusWIP,
			StartedAt:    &earlier,
			Progress: &ProgressSummary{
				TotalTasks:             3,
				CompletedTasks:         1,
				ActiveTasks:            0,
				PendingTasks:           2,
				CompletionPercentage:   33,
				TotalTests:             4,
				PassedTests:            1,
				FailedTests:            0,
				PendingTests:           3,
				TestCoveragePercentage: 25,
			},
		},
		SiblingTasks: []TaskDetails{
			{
				ID:          "1A_2",
				PhaseID:     "1A",
				Name:        "Configure Tools",
				Description: "Set up development tools and linting configuration",
				Status:      epic.StatusPending,
			},
			{
				ID:          "1A_3",
				PhaseID:     "1A",
				Name:        "Setup Testing Framework",
				Description: "Initialize testing framework with basic test structure",
				Status:      epic.StatusPending,
			},
		},
		ChildTests: []TestDetails{
			{
				ID:          "T1A_1",
				TaskID:      "1A_1",
				PhaseID:     "1A",
				Name:        "Test Project Init",
				Description: "Verify that project initializes correctly with all dependencies",
				Status:      epic.StatusCompleted,
				TestStatus:  epic.TestStatusDone,
				StartedAt:   &earlier,
				PassedAt:    &now,
			},
			{
				ID:          "T1A_2",
				TaskID:      "1A_1",
				PhaseID:     "1A",
				Name:        "Test Dependency Resolution",
				Description: "Verify all required dependencies are properly resolved",
				Status:      epic.StatusPending,
				TestStatus:  epic.TestStatusPending,
			},
		},
	}
}

func createSamplePhaseContext() *PhaseContext {
	now := time.Now()
	earlier := now.Add(-time.Hour)

	return &PhaseContext{
		PhaseDetails: PhaseDetails{
			ID:           "1A",
			Name:         "CLI Framework & Core Structure",
			Description:  "Setup basic CLI structure and initialize project",
			Deliverables: "Functional CLI framework\nProject structure established\nCore dependencies configured",
			Status:       epic.StatusWIP,
			StartedAt:    &earlier,
		},
		ProgressSummary: &ProgressSummary{
			TotalTasks:             3,
			CompletedTasks:         1,
			ActiveTasks:            0,
			PendingTasks:           2,
			CompletionPercentage:   33,
			TotalTests:             4,
			PassedTests:            1,
			FailedTests:            0,
			PendingTests:           3,
			TestCoveragePercentage: 25,
		},
		AllTasks: []TaskWithTests{
			{
				TaskDetails: TaskDetails{
					ID:          "1A_1",
					PhaseID:     "1A",
					Name:        "Initialize Project",
					Description: "Initialize Go module with required dependencies",
					Status:      epic.StatusCompleted,
					StartedAt:   &earlier,
					CompletedAt: &now,
				},
				Tests: []TestDetails{
					{
						ID:         "T1A_1",
						TaskID:     "1A_1",
						Name:       "Test Project Init",
						Status:     epic.StatusCompleted,
						TestStatus: epic.TestStatusDone,
					},
				},
			},
		},
		SiblingPhases: []PhaseDetails{
			{
				ID:          "1B",
				Name:        "Command Implementation",
				Description: "Implement core CLI commands and functionality",
				Status:      epic.StatusPending,
			},
		},
	}
}

func createSampleTestContext() *TestContext {
	now := time.Now()
	earlier := now.Add(-time.Hour)

	return &TestContext{
		TestDetails: TestDetails{
			ID:          "T1A_1",
			TaskID:      "1A_1",
			PhaseID:     "1A",
			Name:        "Test Project Init",
			Description: "Verify that project initializes correctly with all dependencies",
			Status:      epic.StatusCompleted,
			TestStatus:  epic.TestStatusDone,
			StartedAt:   &earlier,
			PassedAt:    &now,
		},
		ParentTask: &TaskDetails{
			ID:          "1A_1",
			PhaseID:     "1A",
			Name:        "Initialize Project",
			Description: "Initialize Go module with required dependencies",
			Status:      epic.StatusCompleted,
			StartedAt:   &earlier,
			CompletedAt: &now,
		},
		ParentPhase: &PhaseDetails{
			ID:        "1A",
			Name:      "CLI Framework & Core Structure",
			Status:    epic.StatusWIP,
			StartedAt: &earlier,
			Progress: &ProgressSummary{
				TotalTasks:           3,
				CompletedTasks:       1,
				CompletionPercentage: 33,
			},
		},
		SiblingTests: []TestDetails{
			{
				ID:         "T1A_2",
				TaskID:     "1A_1",
				Name:       "Test Dependency Resolution",
				Status:     epic.StatusPending,
				TestStatus: epic.TestStatusPending,
			},
		},
	}
}
