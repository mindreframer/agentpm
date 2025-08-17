package context

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// OutputFormatter handles formatting context results for different output types
type OutputFormatter interface {
	FormatTaskContext(ctx *TaskContext, writer io.Writer) error
	FormatPhaseContext(ctx *PhaseContext, writer io.Writer) error
	FormatTestContext(ctx *TestContext, writer io.Writer) error
}

// XMLFormatter implements XML output formatting
type XMLFormatter struct{}

// JSONFormatter implements JSON output formatting
type JSONFormatter struct{}

// TextFormatter implements human-readable text output formatting
type TextFormatter struct{}

// NewFormatter creates the appropriate formatter for the given format
func NewFormatter(format string) OutputFormatter {
	switch format {
	case "json":
		return &JSONFormatter{}
	case "xml":
		return &XMLFormatter{}
	default:
		return &TextFormatter{}
	}
}

// XML Formatter Implementation

func (f *XMLFormatter) FormatTaskContext(ctx *TaskContext, writer io.Writer) error {
	// Create XML structure according to specification
	fmt.Fprintf(writer, "<task_context id=\"%s\" phase_id=\"%s\" status=\"%s\">\n",
		ctx.TaskDetails.ID, ctx.TaskDetails.PhaseID, ctx.TaskDetails.Status)

	// Task details
	fmt.Fprintf(writer, "    <task_details>\n")
	fmt.Fprintf(writer, "        <name>%s</name>\n", ctx.TaskDetails.Name)
	if ctx.TaskDetails.Description != "" {
		fmt.Fprintf(writer, "        <description>%s</description>\n", ctx.TaskDetails.Description)
	}
	if ctx.TaskDetails.AcceptanceCriteria != "" {
		fmt.Fprintf(writer, "        <acceptance_criteria>%s</acceptance_criteria>\n", ctx.TaskDetails.AcceptanceCriteria)
	}
	if ctx.TaskDetails.Assignee != "" {
		fmt.Fprintf(writer, "        <assignee>%s</assignee>\n", ctx.TaskDetails.Assignee)
	}
	if ctx.TaskDetails.StartedAt != nil {
		fmt.Fprintf(writer, "        <started_at>%s</started_at>\n", ctx.TaskDetails.StartedAt.Format(time.RFC3339))
	}
	if ctx.TaskDetails.CompletedAt != nil {
		fmt.Fprintf(writer, "        <completed_at>%s</completed_at>\n", ctx.TaskDetails.CompletedAt.Format(time.RFC3339))
	}
	fmt.Fprintf(writer, "    </task_details>\n")

	// Parent phase
	if ctx.ParentPhase != nil {
		fmt.Fprintf(writer, "    <parent_phase id=\"%s\" status=\"%s\">\n", ctx.ParentPhase.ID, ctx.ParentPhase.Status)
		fmt.Fprintf(writer, "        <name>%s</name>\n", ctx.ParentPhase.Name)
		if ctx.ParentPhase.Description != "" {
			fmt.Fprintf(writer, "        <description>%s</description>\n", ctx.ParentPhase.Description)
		}
		if ctx.ParentPhase.Deliverables != "" {
			fmt.Fprintf(writer, "        <deliverables>%s</deliverables>\n", ctx.ParentPhase.Deliverables)
		}
		if ctx.ParentPhase.StartedAt != nil {
			fmt.Fprintf(writer, "        <started_at>%s</started_at>\n", ctx.ParentPhase.StartedAt.Format(time.RFC3339))
		}
		if ctx.ParentPhase.Progress != nil {
			f.writeProgressXML(writer, ctx.ParentPhase.Progress, "        ")
		}
		fmt.Fprintf(writer, "    </parent_phase>\n")
	}

	// Sibling tasks
	if len(ctx.SiblingTasks) > 0 {
		fmt.Fprintf(writer, "    <sibling_tasks>\n")
		for _, task := range ctx.SiblingTasks {
			fmt.Fprintf(writer, "        <task id=\"%s\" status=\"%s\">\n", task.ID, task.Status)
			fmt.Fprintf(writer, "            <name>%s</name>\n", task.Name)
			if task.Description != "" {
				fmt.Fprintf(writer, "            <description>%s</description>\n", task.Description)
			}
			if task.AcceptanceCriteria != "" {
				fmt.Fprintf(writer, "            <acceptance_criteria>%s</acceptance_criteria>\n", task.AcceptanceCriteria)
			}
			fmt.Fprintf(writer, "        </task>\n")
		}
		fmt.Fprintf(writer, "    </sibling_tasks>\n")
	}

	// Child tests
	if len(ctx.ChildTests) > 0 {
		fmt.Fprintf(writer, "    <child_tests>\n")
		for _, test := range ctx.ChildTests {
			fmt.Fprintf(writer, "        <test id=\"%s\" status=\"%s\">\n", test.ID, test.Status)
			fmt.Fprintf(writer, "            <name>%s</name>\n", test.Name)
			if test.Description != "" {
				fmt.Fprintf(writer, "            <description>%s</description>\n", test.Description)
			}
			if test.TestStatus != "" {
				fmt.Fprintf(writer, "            <test_status>%s</test_status>\n", test.TestStatus)
			}
			if test.StartedAt != nil {
				fmt.Fprintf(writer, "            <started_at>%s</started_at>\n", test.StartedAt.Format(time.RFC3339))
			}
			if test.PassedAt != nil {
				fmt.Fprintf(writer, "            <passed_at>%s</passed_at>\n", test.PassedAt.Format(time.RFC3339))
			}
			if test.FailedAt != nil {
				fmt.Fprintf(writer, "            <failed_at>%s</failed_at>\n", test.FailedAt.Format(time.RFC3339))
			}
			fmt.Fprintf(writer, "        </test>\n")
		}
		fmt.Fprintf(writer, "    </child_tests>\n")
	}

	fmt.Fprintf(writer, "</task_context>\n")
	return nil
}

func (f *XMLFormatter) FormatPhaseContext(ctx *PhaseContext, writer io.Writer) error {
	fmt.Fprintf(writer, "<phase_context id=\"%s\" status=\"%s\">\n",
		ctx.PhaseDetails.ID, ctx.PhaseDetails.Status)

	// Phase details
	fmt.Fprintf(writer, "    <phase_details>\n")
	fmt.Fprintf(writer, "        <name>%s</name>\n", ctx.PhaseDetails.Name)
	if ctx.PhaseDetails.Description != "" {
		fmt.Fprintf(writer, "        <description>%s</description>\n", ctx.PhaseDetails.Description)
	}
	if ctx.PhaseDetails.Deliverables != "" {
		fmt.Fprintf(writer, "        <deliverables>%s</deliverables>\n", ctx.PhaseDetails.Deliverables)
	}
	if ctx.PhaseDetails.StartedAt != nil {
		fmt.Fprintf(writer, "        <started_at>%s</started_at>\n", ctx.PhaseDetails.StartedAt.Format(time.RFC3339))
	}
	fmt.Fprintf(writer, "    </phase_details>\n")

	// Progress summary
	if ctx.ProgressSummary != nil {
		f.writeProgressXML(writer, ctx.ProgressSummary, "    ")
	}

	// All tasks
	if len(ctx.AllTasks) > 0 {
		fmt.Fprintf(writer, "    <all_tasks>\n")
		for _, taskWithTests := range ctx.AllTasks {
			task := taskWithTests.TaskDetails
			fmt.Fprintf(writer, "        <task id=\"%s\" status=\"%s\">\n", task.ID, task.Status)
			fmt.Fprintf(writer, "            <name>%s</name>\n", task.Name)
			if task.Description != "" {
				fmt.Fprintf(writer, "            <description>%s</description>\n", task.Description)
			}
			if task.AcceptanceCriteria != "" {
				fmt.Fprintf(writer, "            <acceptance_criteria>%s</acceptance_criteria>\n", task.AcceptanceCriteria)
			}
			if task.Assignee != "" {
				fmt.Fprintf(writer, "            <assignee>%s</assignee>\n", task.Assignee)
			}
			if task.StartedAt != nil {
				fmt.Fprintf(writer, "            <started_at>%s</started_at>\n", task.StartedAt.Format(time.RFC3339))
			}
			if task.CompletedAt != nil {
				fmt.Fprintf(writer, "            <completed_at>%s</completed_at>\n", task.CompletedAt.Format(time.RFC3339))
			}

			// Tests for this task
			if len(taskWithTests.Tests) > 0 {
				fmt.Fprintf(writer, "            <tests>\n")
				for _, test := range taskWithTests.Tests {
					fmt.Fprintf(writer, "                <test id=\"%s\" status=\"%s\">\n", test.ID, test.Status)
					fmt.Fprintf(writer, "                    <name>%s</name>\n", test.Name)
					if test.Description != "" {
						fmt.Fprintf(writer, "                    <description>%s</description>\n", test.Description)
					}
					if test.TestStatus != "" {
						fmt.Fprintf(writer, "                    <test_status>%s</test_status>\n", test.TestStatus)
					}
					fmt.Fprintf(writer, "                </test>\n")
				}
				fmt.Fprintf(writer, "            </tests>\n")
			}

			fmt.Fprintf(writer, "        </task>\n")
		}
		fmt.Fprintf(writer, "    </all_tasks>\n")
	}

	// Phase tests (tests directly associated with the phase)
	if len(ctx.PhaseTests) > 0 {
		fmt.Fprintf(writer, "    <phase_tests>\n")
		for _, test := range ctx.PhaseTests {
			fmt.Fprintf(writer, "        <test id=\"%s\" status=\"%s\">\n", test.ID, test.Status)
			fmt.Fprintf(writer, "            <name>%s</name>\n", test.Name)
			if test.Description != "" {
				fmt.Fprintf(writer, "            <description>%s</description>\n", test.Description)
			}
			if test.TestStatus != "" {
				fmt.Fprintf(writer, "            <test_status>%s</test_status>\n", test.TestStatus)
			}
			fmt.Fprintf(writer, "        </test>\n")
		}
		fmt.Fprintf(writer, "    </phase_tests>\n")
	}

	// Sibling phases
	if len(ctx.SiblingPhases) > 0 {
		fmt.Fprintf(writer, "    <sibling_phases>\n")
		for _, phase := range ctx.SiblingPhases {
			fmt.Fprintf(writer, "        <phase id=\"%s\" status=\"%s\">\n", phase.ID, phase.Status)
			fmt.Fprintf(writer, "            <name>%s</name>\n", phase.Name)
			if phase.Description != "" {
				fmt.Fprintf(writer, "            <description>%s</description>\n", phase.Description)
			}
			fmt.Fprintf(writer, "        </phase>\n")
		}
		fmt.Fprintf(writer, "    </sibling_phases>\n")
	}

	fmt.Fprintf(writer, "</phase_context>\n")
	return nil
}

func (f *XMLFormatter) FormatTestContext(ctx *TestContext, writer io.Writer) error {
	fmt.Fprintf(writer, "<test_context id=\"%s\" task_id=\"%s\" status=\"%s\">\n",
		ctx.TestDetails.ID, ctx.TestDetails.TaskID, ctx.TestDetails.Status)

	// Test details
	fmt.Fprintf(writer, "    <test_details>\n")
	fmt.Fprintf(writer, "        <name>%s</name>\n", ctx.TestDetails.Name)
	if ctx.TestDetails.Description != "" {
		fmt.Fprintf(writer, "        <description>%s</description>\n", ctx.TestDetails.Description)
	}
	if ctx.TestDetails.TestStatus != "" {
		fmt.Fprintf(writer, "        <test_status>%s</test_status>\n", ctx.TestDetails.TestStatus)
	}
	if ctx.TestDetails.StartedAt != nil {
		fmt.Fprintf(writer, "        <started_at>%s</started_at>\n", ctx.TestDetails.StartedAt.Format(time.RFC3339))
	}
	if ctx.TestDetails.PassedAt != nil {
		fmt.Fprintf(writer, "        <passed_at>%s</passed_at>\n", ctx.TestDetails.PassedAt.Format(time.RFC3339))
	}
	if ctx.TestDetails.FailedAt != nil {
		fmt.Fprintf(writer, "        <failed_at>%s</failed_at>\n", ctx.TestDetails.FailedAt.Format(time.RFC3339))
	}
	fmt.Fprintf(writer, "    </test_details>\n")

	// Parent task
	if ctx.ParentTask != nil {
		fmt.Fprintf(writer, "    <parent_task id=\"%s\" status=\"%s\">\n", ctx.ParentTask.ID, ctx.ParentTask.Status)
		fmt.Fprintf(writer, "        <name>%s</name>\n", ctx.ParentTask.Name)
		if ctx.ParentTask.Description != "" {
			fmt.Fprintf(writer, "        <description>%s</description>\n", ctx.ParentTask.Description)
		}
		if ctx.ParentTask.AcceptanceCriteria != "" {
			fmt.Fprintf(writer, "        <acceptance_criteria>%s</acceptance_criteria>\n", ctx.ParentTask.AcceptanceCriteria)
		}
		fmt.Fprintf(writer, "    </parent_task>\n")
	}

	// Parent phase
	if ctx.ParentPhase != nil {
		fmt.Fprintf(writer, "    <parent_phase id=\"%s\" status=\"%s\">\n", ctx.ParentPhase.ID, ctx.ParentPhase.Status)
		fmt.Fprintf(writer, "        <name>%s</name>\n", ctx.ParentPhase.Name)
		if ctx.ParentPhase.Description != "" {
			fmt.Fprintf(writer, "        <description>%s</description>\n", ctx.ParentPhase.Description)
		}
		if ctx.ParentPhase.Deliverables != "" {
			fmt.Fprintf(writer, "        <deliverables>%s</deliverables>\n", ctx.ParentPhase.Deliverables)
		}
		if ctx.ParentPhase.Progress != nil {
			f.writeProgressXML(writer, ctx.ParentPhase.Progress, "        ")
		}
		fmt.Fprintf(writer, "    </parent_phase>\n")
	}

	// Sibling tests
	if len(ctx.SiblingTests) > 0 {
		fmt.Fprintf(writer, "    <sibling_tests>\n")
		for _, test := range ctx.SiblingTests {
			fmt.Fprintf(writer, "        <test id=\"%s\" status=\"%s\">\n", test.ID, test.Status)
			fmt.Fprintf(writer, "            <name>%s</name>\n", test.Name)
			if test.Description != "" {
				fmt.Fprintf(writer, "            <description>%s</description>\n", test.Description)
			}
			if test.TestStatus != "" {
				fmt.Fprintf(writer, "            <test_status>%s</test_status>\n", test.TestStatus)
			}
			fmt.Fprintf(writer, "        </test>\n")
		}
		fmt.Fprintf(writer, "    </sibling_tests>\n")
	}

	fmt.Fprintf(writer, "</test_context>\n")
	return nil
}

func (f *XMLFormatter) writeProgressXML(writer io.Writer, progress *ProgressSummary, indent string) {
	fmt.Fprintf(writer, "%s<progress>\n", indent)
	fmt.Fprintf(writer, "%s    <total_tasks>%d</total_tasks>\n", indent, progress.TotalTasks)
	fmt.Fprintf(writer, "%s    <completed_tasks>%d</completed_tasks>\n", indent, progress.CompletedTasks)
	fmt.Fprintf(writer, "%s    <pending_tasks>%d</pending_tasks>\n", indent, progress.PendingTasks)
	fmt.Fprintf(writer, "%s    <completion_percentage>%d</completion_percentage>\n", indent, progress.CompletionPercentage)
	fmt.Fprintf(writer, "%s    <total_tests>%d</total_tests>\n", indent, progress.TotalTests)
	fmt.Fprintf(writer, "%s    <passed_tests>%d</passed_tests>\n", indent, progress.PassedTests)
	fmt.Fprintf(writer, "%s    <failed_tests>%d</failed_tests>\n", indent, progress.FailedTests)
	fmt.Fprintf(writer, "%s    <pending_tests>%d</pending_tests>\n", indent, progress.PendingTests)
	fmt.Fprintf(writer, "%s    <test_coverage_percentage>%d</test_coverage_percentage>\n", indent, progress.TestCoveragePercentage)
	fmt.Fprintf(writer, "%s</progress>\n", indent)
}

// JSON Formatter Implementation

func (f *JSONFormatter) FormatTaskContext(ctx *TaskContext, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(ctx)
}

func (f *JSONFormatter) FormatPhaseContext(ctx *PhaseContext, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(ctx)
}

func (f *JSONFormatter) FormatTestContext(ctx *TestContext, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(ctx)
}

// Text Formatter Implementation

func (f *TextFormatter) FormatTaskContext(ctx *TaskContext, writer io.Writer) error {
	fmt.Fprintf(writer, "Task: %s (ID: %s)\n", ctx.TaskDetails.Name, ctx.TaskDetails.ID)
	fmt.Fprintf(writer, "Phase: %s\n", ctx.TaskDetails.PhaseID)
	fmt.Fprintf(writer, "Status: %s\n", ctx.TaskDetails.Status)

	if ctx.TaskDetails.Description != "" {
		fmt.Fprintf(writer, "Description: %s\n", ctx.TaskDetails.Description)
	}

	if ctx.TaskDetails.AcceptanceCriteria != "" {
		fmt.Fprintf(writer, "Acceptance Criteria:\n%s\n", indentText(ctx.TaskDetails.AcceptanceCriteria, "  "))
	}

	if ctx.TaskDetails.Assignee != "" {
		fmt.Fprintf(writer, "Assignee: %s\n", ctx.TaskDetails.Assignee)
	}

	if ctx.TaskDetails.StartedAt != nil {
		fmt.Fprintf(writer, "Started: %s\n", ctx.TaskDetails.StartedAt.Format("2006-01-02 15:04:05"))
	}

	if ctx.TaskDetails.CompletedAt != nil {
		fmt.Fprintf(writer, "Completed: %s\n", ctx.TaskDetails.CompletedAt.Format("2006-01-02 15:04:05"))
	}

	// Parent phase
	if ctx.ParentPhase != nil {
		fmt.Fprintf(writer, "\nParent Phase: %s (%s)\n", ctx.ParentPhase.Name, ctx.ParentPhase.ID)
		fmt.Fprintf(writer, "Phase Status: %s\n", ctx.ParentPhase.Status)
		if ctx.ParentPhase.Description != "" {
			fmt.Fprintf(writer, "Phase Description: %s\n", ctx.ParentPhase.Description)
		}
		if ctx.ParentPhase.Progress != nil {
			fmt.Fprintf(writer, "Phase Progress: %d%% (%d/%d tasks completed)\n",
				ctx.ParentPhase.Progress.CompletionPercentage,
				ctx.ParentPhase.Progress.CompletedTasks,
				ctx.ParentPhase.Progress.TotalTasks)
		}
	}

	// Sibling tasks
	if len(ctx.SiblingTasks) > 0 {
		fmt.Fprintf(writer, "\nSibling Tasks (%d):\n", len(ctx.SiblingTasks))
		for _, task := range ctx.SiblingTasks {
			fmt.Fprintf(writer, "  %s - %s [%s]\n", task.ID, task.Name, task.Status)
			if task.Description != "" {
				fmt.Fprintf(writer, "    Description: %s\n", task.Description)
			}
			if task.AcceptanceCriteria != "" {
				fmt.Fprintf(writer, "    Acceptance Criteria:\n%s\n", indentText(task.AcceptanceCriteria, "      "))
			}
		}
	}

	// Child tests
	if len(ctx.ChildTests) > 0 {
		fmt.Fprintf(writer, "\nChild Tests (%d):\n", len(ctx.ChildTests))
		for _, test := range ctx.ChildTests {
			fmt.Fprintf(writer, "  %s - %s [%s", test.ID, test.Name, test.Status)
			if test.TestStatus != "" {
				fmt.Fprintf(writer, "/%s", test.TestStatus)
			}
			fmt.Fprintf(writer, "]\n")
			if test.Description != "" {
				fmt.Fprintf(writer, "    %s\n", test.Description)
			}
		}
	}

	return nil
}

func (f *TextFormatter) FormatPhaseContext(ctx *PhaseContext, writer io.Writer) error {
	fmt.Fprintf(writer, "Phase: %s (ID: %s)\n", ctx.PhaseDetails.Name, ctx.PhaseDetails.ID)
	fmt.Fprintf(writer, "Status: %s\n", ctx.PhaseDetails.Status)

	if ctx.PhaseDetails.Description != "" {
		fmt.Fprintf(writer, "Description: %s\n", ctx.PhaseDetails.Description)
	}

	if ctx.PhaseDetails.Deliverables != "" {
		fmt.Fprintf(writer, "Deliverables:\n%s\n", indentText(ctx.PhaseDetails.Deliverables, "  "))
	}

	if ctx.PhaseDetails.StartedAt != nil {
		fmt.Fprintf(writer, "Started: %s\n", ctx.PhaseDetails.StartedAt.Format("2006-01-02 15:04:05"))
	}

	if ctx.PhaseDetails.CompletedAt != nil {
		fmt.Fprintf(writer, "Completed: %s\n", ctx.PhaseDetails.CompletedAt.Format("2006-01-02 15:04:05"))
	}

	// Progress summary
	if ctx.ProgressSummary != nil {
		fmt.Fprintf(writer, "\nProgress Summary:\n")
		fmt.Fprintf(writer, "  Tasks: %d total, %d completed, %d active, %d pending (%d%% complete)\n",
			ctx.ProgressSummary.TotalTasks,
			ctx.ProgressSummary.CompletedTasks,
			ctx.ProgressSummary.ActiveTasks,
			ctx.ProgressSummary.PendingTasks,
			ctx.ProgressSummary.CompletionPercentage)
		fmt.Fprintf(writer, "  Tests: %d total, %d passed, %d failed, %d pending (%d%% coverage)\n",
			ctx.ProgressSummary.TotalTests,
			ctx.ProgressSummary.PassedTests,
			ctx.ProgressSummary.FailedTests,
			ctx.ProgressSummary.PendingTests,
			ctx.ProgressSummary.TestCoveragePercentage)
	}

	// All tasks
	if len(ctx.AllTasks) > 0 {
		fmt.Fprintf(writer, "\nAll Tasks (%d):\n", len(ctx.AllTasks))
		for _, taskWithTests := range ctx.AllTasks {
			task := taskWithTests.TaskDetails
			fmt.Fprintf(writer, "  %s - %s [%s]\n", task.ID, task.Name, task.Status)
			if task.Description != "" {
				fmt.Fprintf(writer, "    Description: %s\n", task.Description)
			}
			if task.AcceptanceCriteria != "" {
				fmt.Fprintf(writer, "    Acceptance Criteria:\n%s\n", indentText(task.AcceptanceCriteria, "      "))
			}
			if len(taskWithTests.Tests) > 0 {
				fmt.Fprintf(writer, "    Tests (%d):\n", len(taskWithTests.Tests))
				for _, test := range taskWithTests.Tests {
					fmt.Fprintf(writer, "      %s - %s [%s", test.ID, test.Name, test.Status)
					if test.TestStatus != "" {
						fmt.Fprintf(writer, "/%s", test.TestStatus)
					}
					fmt.Fprintf(writer, "]\n")
				}
			}
		}
	}

	// Phase tests (tests directly associated with the phase)
	if len(ctx.PhaseTests) > 0 {
		fmt.Fprintf(writer, "\nPhase Tests (%d):\n", len(ctx.PhaseTests))
		for _, test := range ctx.PhaseTests {
			fmt.Fprintf(writer, "  %s - %s [%s", test.ID, test.Name, test.Status)
			if test.TestStatus != "" {
				fmt.Fprintf(writer, "/%s", test.TestStatus)
			}
			fmt.Fprintf(writer, "]\n")
			if test.Description != "" {
				fmt.Fprintf(writer, "    %s\n", test.Description)
			}
		}
	}

	// Sibling phases
	if len(ctx.SiblingPhases) > 0 {
		fmt.Fprintf(writer, "\nSibling Phases (%d):\n", len(ctx.SiblingPhases))
		for _, phase := range ctx.SiblingPhases {
			fmt.Fprintf(writer, "  %s - %s [%s]\n", phase.ID, phase.Name, phase.Status)
		}
	}

	return nil
}

func (f *TextFormatter) FormatTestContext(ctx *TestContext, writer io.Writer) error {
	fmt.Fprintf(writer, "Test: %s (ID: %s)\n", ctx.TestDetails.Name, ctx.TestDetails.ID)
	fmt.Fprintf(writer, "Task: %s\n", ctx.TestDetails.TaskID)
	fmt.Fprintf(writer, "Status: %s", ctx.TestDetails.Status)
	if ctx.TestDetails.TestStatus != "" {
		fmt.Fprintf(writer, " / Test Status: %s", ctx.TestDetails.TestStatus)
	}
	fmt.Fprintf(writer, "\n")

	if ctx.TestDetails.Description != "" {
		fmt.Fprintf(writer, "Description: %s\n", ctx.TestDetails.Description)
	}

	if ctx.TestDetails.StartedAt != nil {
		fmt.Fprintf(writer, "Started: %s\n", ctx.TestDetails.StartedAt.Format("2006-01-02 15:04:05"))
	}

	if ctx.TestDetails.PassedAt != nil {
		fmt.Fprintf(writer, "Passed: %s\n", ctx.TestDetails.PassedAt.Format("2006-01-02 15:04:05"))
	}

	if ctx.TestDetails.FailedAt != nil {
		fmt.Fprintf(writer, "Failed: %s\n", ctx.TestDetails.FailedAt.Format("2006-01-02 15:04:05"))
	}

	if ctx.TestDetails.FailureNote != "" {
		fmt.Fprintf(writer, "Failure Note: %s\n", ctx.TestDetails.FailureNote)
	}

	// Parent task
	if ctx.ParentTask != nil {
		fmt.Fprintf(writer, "\nParent Task: %s (%s)\n", ctx.ParentTask.Name, ctx.ParentTask.ID)
		fmt.Fprintf(writer, "Task Status: %s\n", ctx.ParentTask.Status)
		if ctx.ParentTask.Description != "" {
			fmt.Fprintf(writer, "Task Description: %s\n", ctx.ParentTask.Description)
		}
	}

	// Parent phase
	if ctx.ParentPhase != nil {
		fmt.Fprintf(writer, "\nParent Phase: %s (%s)\n", ctx.ParentPhase.Name, ctx.ParentPhase.ID)
		fmt.Fprintf(writer, "Phase Status: %s\n", ctx.ParentPhase.Status)
		if ctx.ParentPhase.Progress != nil {
			fmt.Fprintf(writer, "Phase Progress: %d%% (%d/%d tasks completed)\n",
				ctx.ParentPhase.Progress.CompletionPercentage,
				ctx.ParentPhase.Progress.CompletedTasks,
				ctx.ParentPhase.Progress.TotalTasks)
		}
	}

	// Sibling tests
	if len(ctx.SiblingTests) > 0 {
		fmt.Fprintf(writer, "\nSibling Tests (%d):\n", len(ctx.SiblingTests))
		for _, test := range ctx.SiblingTests {
			fmt.Fprintf(writer, "  %s - %s [%s", test.ID, test.Name, test.Status)
			if test.TestStatus != "" {
				fmt.Fprintf(writer, "/%s", test.TestStatus)
			}
			fmt.Fprintf(writer, "]\n")
		}
	}

	return nil
}

// Helper function to indent multi-line text
func indentText(text, indent string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}
