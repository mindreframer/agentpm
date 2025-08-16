package epic

import (
	"fmt"
	"strings"
)

type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Warnings []string          `json:"warnings,omitempty"`
	Errors   []string          `json:"errors,omitempty"`
	Checks   map[string]string `json:"checks_performed"`
}

func (vr *ValidationResult) AddError(msg string) {
	vr.Errors = append(vr.Errors, msg)
	vr.Valid = false
}

func (vr *ValidationResult) AddWarning(msg string) {
	vr.Warnings = append(vr.Warnings, msg)
}

func (vr *ValidationResult) SetCheck(name, status string) {
	if vr.Checks == nil {
		vr.Checks = make(map[string]string)
	}
	vr.Checks[name] = status
}

func (vr *ValidationResult) Message() string {
	if len(vr.Errors) > 0 {
		return fmt.Sprintf("Epic validation failed with %d error(s)", len(vr.Errors))
	}
	if len(vr.Warnings) > 0 {
		return fmt.Sprintf("Epic structure is valid with %d warning(s)", len(vr.Warnings))
	}
	return "Epic structure is valid"
}

func (e *Epic) Validate() *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Checks: make(map[string]string),
	}

	e.validateBasicStructure(result)
	e.validateStatusValues(result)
	e.validatePhaseDependencies(result)
	e.validateTaskPhaseMapping(result)
	e.validateTestCoverage(result)

	return result
}

func (e *Epic) validateBasicStructure(result *ValidationResult) {
	// Check required fields
	if e.ID == "" {
		result.AddError("Epic ID is required")
	}
	if e.Name == "" {
		result.AddError("Epic name is required")
	}
	if e.Status == "" {
		result.AddError("Epic status is required")
	}
	if e.CreatedAt.IsZero() {
		result.AddError("Epic created_at timestamp is required")
	}

	// Check for duplicate phase IDs
	phaseIDs := make(map[string]bool)
	for _, phase := range e.Phases {
		if phase.ID == "" {
			result.AddError("Phase ID is required")
			continue
		}
		if phaseIDs[phase.ID] {
			result.AddError(fmt.Sprintf("Duplicate phase ID: %s", phase.ID))
		}
		phaseIDs[phase.ID] = true
	}

	// Check for duplicate task IDs
	taskIDs := make(map[string]bool)
	for _, task := range e.Tasks {
		if task.ID == "" {
			result.AddError("Task ID is required")
			continue
		}
		if taskIDs[task.ID] {
			result.AddError(fmt.Sprintf("Duplicate task ID: %s", task.ID))
		}
		taskIDs[task.ID] = true
	}

	// Check for duplicate test IDs
	testIDs := make(map[string]bool)
	for _, test := range e.Tests {
		if test.ID == "" {
			result.AddError("Test ID is required")
			continue
		}
		if testIDs[test.ID] {
			result.AddError(fmt.Sprintf("Duplicate test ID: %s", test.ID))
		}
		testIDs[test.ID] = true
	}

	if len(result.Errors) == 0 {
		result.SetCheck("xml_structure", "passed")
	} else {
		result.SetCheck("xml_structure", "failed")
	}
}

func (e *Epic) validateStatusValues(result *ValidationResult) {
	// Validate epic status
	if !e.Status.IsValid() {
		result.AddError(fmt.Sprintf("Invalid epic status: %s", e.Status))
	}

	// Validate phase statuses
	for _, phase := range e.Phases {
		if !phase.Status.IsValid() {
			result.AddError(fmt.Sprintf("Invalid phase status for %s: %s", phase.ID, phase.Status))
		}
	}

	// Validate task statuses
	for _, task := range e.Tasks {
		if !task.Status.IsValid() {
			result.AddError(fmt.Sprintf("Invalid task status for %s: %s", task.ID, task.Status))
		}
	}

	// Validate test statuses
	for _, test := range e.Tests {
		if !test.Status.IsValid() {
			result.AddError(fmt.Sprintf("Invalid test status for %s: %s", test.ID, test.Status))
		}
	}

	if len(result.Errors) == 0 {
		result.SetCheck("status_values", "passed")
	} else {
		result.SetCheck("status_values", "failed")
	}
}

func (e *Epic) validatePhaseDependencies(result *ValidationResult) {
	// Build phase map for quick lookup
	phaseMap := make(map[string]*Phase)
	for i := range e.Phases {
		phaseMap[e.Phases[i].ID] = &e.Phases[i]
	}

	// For now, just ensure phase IDs are valid
	// Future: implement dependency chain validation

	if len(phaseMap) == len(e.Phases) {
		result.SetCheck("phase_dependencies", "passed")
	} else {
		result.SetCheck("phase_dependencies", "failed")
	}
}

func (e *Epic) validateTaskPhaseMapping(result *ValidationResult) {
	// Build phase map for quick lookup
	phaseMap := make(map[string]bool)
	for _, phase := range e.Phases {
		phaseMap[phase.ID] = true
	}

	// Check that all tasks reference valid phases
	for _, task := range e.Tasks {
		if task.PhaseID != "" && !phaseMap[task.PhaseID] {
			result.AddError(fmt.Sprintf("Task %s references non-existent phase: %s", task.ID, task.PhaseID))
		}
	}

	if len(result.Errors) == 0 {
		result.SetCheck("task_phase_mapping", "passed")
	} else {
		result.SetCheck("task_phase_mapping", "failed")
	}
}

func (e *Epic) validateTestCoverage(result *ValidationResult) {
	// Build task map for quick lookup
	taskMap := make(map[string]bool)
	for _, task := range e.Tasks {
		taskMap[task.ID] = true
	}

	// Check that all tests reference valid tasks
	for _, test := range e.Tests {
		if test.TaskID != "" && !taskMap[test.TaskID] {
			result.AddError(fmt.Sprintf("Test %s references non-existent task: %s", test.ID, test.TaskID))
		}
	}

	// Check for tasks without tests
	tasksWithTests := make(map[string]bool)
	for _, test := range e.Tests {
		if test.TaskID != "" {
			tasksWithTests[test.TaskID] = true
		}
	}

	for _, task := range e.Tasks {
		if !tasksWithTests[task.ID] {
			result.AddWarning(fmt.Sprintf("Task %s has no tests defined", task.ID))
		}
	}

	if len(result.Errors) == 0 {
		if len(result.Warnings) > 0 {
			result.SetCheck("test_coverage", "warning")
		} else {
			result.SetCheck("test_coverage", "passed")
		}
	} else {
		result.SetCheck("test_coverage", "failed")
	}
}

// ValidateFromFile loads an epic from file and validates it
func ValidateFromFile(storage interface{ LoadEpic(string) (*Epic, error) }, filePath string) (*ValidationResult, error) {
	epic, err := storage.LoadEpic(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load epic: %w", err)
	}

	return epic.Validate(), nil
}

// FormatValidationResult formats validation result as a string
func FormatValidationResult(result *ValidationResult, format string) string {
	switch strings.ToLower(format) {
	case "xml":
		return formatValidationResultXML(result)
	case "json":
		return formatValidationResultJSON(result)
	default: // text
		return formatValidationResultText(result)
	}
}

func formatValidationResultXML(result *ValidationResult) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("<validation_result>\n"))
	output.WriteString(fmt.Sprintf("    <valid>%v</valid>\n", result.Valid))

	if len(result.Warnings) > 0 {
		output.WriteString("    <warnings>\n")
		for _, warning := range result.Warnings {
			output.WriteString(fmt.Sprintf("        <warning>%s</warning>\n", warning))
		}
		output.WriteString("    </warnings>\n")
	}

	if len(result.Errors) > 0 {
		output.WriteString("    <errors>\n")
		for _, err := range result.Errors {
			output.WriteString(fmt.Sprintf("        <error>%s</error>\n", err))
		}
		output.WriteString("    </errors>\n")
	}

	if len(result.Checks) > 0 {
		output.WriteString("    <checks_performed>\n")
		for name, status := range result.Checks {
			output.WriteString(fmt.Sprintf("        <check name=\"%s\">%s</check>\n", name, status))
		}
		output.WriteString("    </checks_performed>\n")
	}

	output.WriteString(fmt.Sprintf("    <message>%s</message>\n", result.Message()))
	output.WriteString("</validation_result>")

	return output.String()
}

func formatValidationResultJSON(result *ValidationResult) string {
	// Simple JSON formatting - could use encoding/json but keeping it minimal
	var output strings.Builder

	output.WriteString("{\n")
	output.WriteString(fmt.Sprintf("  \"valid\": %v,\n", result.Valid))
	output.WriteString(fmt.Sprintf("  \"message\": \"%s\"", result.Message()))

	if len(result.Warnings) > 0 {
		output.WriteString(",\n  \"warnings\": [")
		for i, warning := range result.Warnings {
			if i > 0 {
				output.WriteString(", ")
			}
			output.WriteString(fmt.Sprintf("\"%s\"", warning))
		}
		output.WriteString("]")
	}

	if len(result.Errors) > 0 {
		output.WriteString(",\n  \"errors\": [")
		for i, err := range result.Errors {
			if i > 0 {
				output.WriteString(", ")
			}
			output.WriteString(fmt.Sprintf("\"%s\"", err))
		}
		output.WriteString("]")
	}

	output.WriteString("\n}")

	return output.String()
}

func formatValidationResultText(result *ValidationResult) string {
	var output strings.Builder

	if result.Valid {
		output.WriteString("✓ Epic validation passed")
	} else {
		output.WriteString("✗ Epic validation failed")
	}

	if len(result.Warnings) > 0 {
		output.WriteString(fmt.Sprintf(" (%d warnings)", len(result.Warnings)))
	}

	if len(result.Errors) > 0 {
		output.WriteString(fmt.Sprintf(" (%d errors)", len(result.Errors)))
	}

	output.WriteString("\n\n")

	if len(result.Errors) > 0 {
		output.WriteString("Errors:\n")
		for _, err := range result.Errors {
			output.WriteString(fmt.Sprintf("  • %s\n", err))
		}
		output.WriteString("\n")
	}

	if len(result.Warnings) > 0 {
		output.WriteString("Warnings:\n")
		for _, warning := range result.Warnings {
			output.WriteString(fmt.Sprintf("  • %s\n", warning))
		}
		output.WriteString("\n")
	}

	output.WriteString("Checks performed:\n")
	for name, status := range result.Checks {
		var icon string
		switch status {
		case "passed":
			icon = "✓"
		case "failed":
			icon = "✗"
		case "warning":
			icon = "⚠"
		default:
			icon = "?"
		}
		output.WriteString(fmt.Sprintf("  %s %s: %s\n", icon, name, status))
	}

	return output.String()
}
