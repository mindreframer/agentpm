package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/mindreframer/agentpm/internal/epic"
)

type FileStorage struct{}

func NewFileStorage() *FileStorage {
	return &FileStorage{}
}

func (fs *FileStorage) LoadEpic(filePath string) (*epic.Epic, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve epic file path: %w", err)
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromFile(absPath); err != nil {
		return nil, fmt.Errorf("failed to read epic file: %w", err)
	}

	root := doc.SelectElement("epic")
	if root == nil {
		return nil, fmt.Errorf("invalid epic file: missing <epic> root element")
	}

	epicData := &epic.Epic{}

	epicData.ID = root.SelectAttrValue("id", "")
	epicData.Name = root.SelectAttrValue("name", "")
	epicData.Status = epic.Status(root.SelectAttrValue("status", ""))

	// Parse created_at timestamp
	if createdAtStr := root.SelectAttrValue("created_at", ""); createdAtStr != "" {
		if t, err := time.Parse("2006-01-02T15:04:05Z", createdAtStr); err == nil {
			epicData.CreatedAt = t
		}
	}

	if assigneeElem := root.SelectElement("assignee"); assigneeElem != nil {
		epicData.Assignee = assigneeElem.Text()
	}

	if descElem := root.SelectElement("description"); descElem != nil {
		epicData.Description = getInnerXML(descElem)
	}

	if workflowElem := root.SelectElement("workflow"); workflowElem != nil {
		epicData.Workflow = getInnerXML(workflowElem)
	}

	if requirementsElem := root.SelectElement("requirements"); requirementsElem != nil {
		epicData.Requirements = getInnerXML(requirementsElem)
	}

	if dependenciesElem := root.SelectElement("dependencies"); dependenciesElem != nil {
		epicData.Dependencies = getInnerXML(dependenciesElem)
	}

	// Parse metadata section (Epic 7)
	if metadataElem := root.SelectElement("metadata"); metadataElem != nil {
		metadata := &epic.EpicMetadata{}

		if createdElem := metadataElem.SelectElement("created"); createdElem != nil {
			if t, err := time.Parse(time.RFC3339, createdElem.Text()); err == nil {
				metadata.Created = t
			}
		}

		if assigneeElem := metadataElem.SelectElement("assignee"); assigneeElem != nil {
			metadata.Assignee = assigneeElem.Text()
		}

		if effortElem := metadataElem.SelectElement("estimated_effort"); effortElem != nil {
			metadata.EstimatedEffort = effortElem.Text()
		}

		epicData.Metadata = metadata
	}

	// Parse current_state section (Epic 7)
	if currentStateElem := root.SelectElement("current_state"); currentStateElem != nil {
		currentState := &epic.CurrentState{}

		if activePhaseElem := currentStateElem.SelectElement("active_phase"); activePhaseElem != nil {
			currentState.ActivePhase = activePhaseElem.Text()
		}

		if activeTaskElem := currentStateElem.SelectElement("active_task"); activeTaskElem != nil {
			currentState.ActiveTask = activeTaskElem.Text()
		}

		if nextActionElem := currentStateElem.SelectElement("next_action"); nextActionElem != nil {
			currentState.NextAction = nextActionElem.Text()
		}

		epicData.CurrentState = currentState
	}

	if phasesElem := root.SelectElement("phases"); phasesElem != nil {
		for _, phaseElem := range phasesElem.SelectElements("phase") {
			phase := epic.Phase{
				ID:     phaseElem.SelectAttrValue("id", ""),
				Name:   phaseElem.SelectAttrValue("name", ""),
				Status: epic.Status(phaseElem.SelectAttrValue("status", "")),
			}
			if descElem := phaseElem.SelectElement("description"); descElem != nil {
				phase.Description = getInnerXML(descElem)
			}
			// Load timestamps
			if startedElem := phaseElem.SelectElement("started_at"); startedElem != nil {
				if t, err := time.Parse(time.RFC3339, startedElem.Text()); err == nil {
					phase.StartedAt = &t
				}
			}
			if completedElem := phaseElem.SelectElement("completed_at"); completedElem != nil {
				if t, err := time.Parse(time.RFC3339, completedElem.Text()); err == nil {
					phase.CompletedAt = &t
				}
			}
			epicData.Phases = append(epicData.Phases, phase)
		}
	}

	if tasksElem := root.SelectElement("tasks"); tasksElem != nil {
		for _, taskElem := range tasksElem.SelectElements("task") {
			task := epic.Task{
				ID:       taskElem.SelectAttrValue("id", ""),
				PhaseID:  taskElem.SelectAttrValue("phase_id", ""),
				Name:     taskElem.SelectAttrValue("name", ""),
				Status:   epic.Status(taskElem.SelectAttrValue("status", "")),
				Assignee: taskElem.SelectAttrValue("assignee", ""),
			}
			if descElem := taskElem.SelectElement("description"); descElem != nil {
				task.Description = getInnerXML(descElem)
			}
			// Load timestamps
			if startedElem := taskElem.SelectElement("started_at"); startedElem != nil {
				if t, err := time.Parse(time.RFC3339, startedElem.Text()); err == nil {
					task.StartedAt = &t
				}
			}
			if completedElem := taskElem.SelectElement("completed_at"); completedElem != nil {
				if t, err := time.Parse(time.RFC3339, completedElem.Text()); err == nil {
					task.CompletedAt = &t
				}
			}
			if cancelledElem := taskElem.SelectElement("cancelled_at"); cancelledElem != nil {
				if t, err := time.Parse(time.RFC3339, cancelledElem.Text()); err == nil {
					task.CancelledAt = &t
				}
			}
			epicData.Tasks = append(epicData.Tasks, task)
		}
	}

	if testsElem := root.SelectElement("tests"); testsElem != nil {
		for _, testElem := range testsElem.SelectElements("test") {
			test := epic.Test{
				ID:         testElem.SelectAttrValue("id", ""),
				TaskID:     testElem.SelectAttrValue("task_id", ""),
				PhaseID:    testElem.SelectAttrValue("phase_id", ""),
				Name:       testElem.SelectAttrValue("name", ""),
				Status:     epic.Status(testElem.SelectAttrValue("status", "")),
				TestStatus: epic.TestStatus(testElem.SelectAttrValue("test_status", "")),
			}

			// First try to get content from inner text (direct content within <test>)
			// This handles the format: <test>content here</test>
			// Only use inner text if there are no child elements like <description>
			if descElem := testElem.SelectElement("description"); descElem != nil {
				// Use description element format: <test><description>content</description></test>
				test.Description = getInnerXML(descElem)
			} else {
				// Fall back to inner text format: <test>content here</test>
				innerText := getInnerXML(testElem)
				if innerText != "" {
					test.Description = innerText
				}
			}

			// Epic 4 enhancements - load timestamp fields
			if startedElem := testElem.SelectElement("started_at"); startedElem != nil {
				if t, err := time.Parse(time.RFC3339, startedElem.Text()); err == nil {
					test.StartedAt = &t
				}
			}
			if passedElem := testElem.SelectElement("passed_at"); passedElem != nil {
				if t, err := time.Parse(time.RFC3339, passedElem.Text()); err == nil {
					test.PassedAt = &t
				}
			}
			if failedElem := testElem.SelectElement("failed_at"); failedElem != nil {
				if t, err := time.Parse(time.RFC3339, failedElem.Text()); err == nil {
					test.FailedAt = &t
				}
			}
			if cancelledElem := testElem.SelectElement("cancelled_at"); cancelledElem != nil {
				if t, err := time.Parse(time.RFC3339, cancelledElem.Text()); err == nil {
					test.CancelledAt = &t
				}
			}

			// Epic 4 note fields
			if failureElem := testElem.SelectElement("failure_note"); failureElem != nil {
				test.FailureNote = getInnerXML(failureElem)
			}
			if cancellationElem := testElem.SelectElement("cancellation_reason"); cancellationElem != nil {
				test.CancellationReason = getInnerXML(cancellationElem)
			}

			epicData.Tests = append(epicData.Tests, test)
		}
	}

	// Parse events
	if eventsElem := root.SelectElement("events"); eventsElem != nil {
		for _, eventElem := range eventsElem.SelectElements("event") {
			event := epic.Event{
				ID:   eventElem.SelectAttrValue("id", ""),
				Type: eventElem.SelectAttrValue("type", ""),
			}

			// Parse timestamp
			if timestampStr := eventElem.SelectAttrValue("timestamp", ""); timestampStr != "" {
				if t, err := time.Parse(time.RFC3339, timestampStr); err == nil {
					event.Timestamp = t
				}
			}

			// Parse data/content
			if dataElem := eventElem.SelectElement("data"); dataElem != nil {
				event.Data = dataElem.Text()
			} else {
				// For backward compatibility, use the element text
				event.Data = eventElem.Text()
			}

			epicData.Events = append(epicData.Events, event)
		}
	}

	return epicData, nil
}

func (fs *FileStorage) SaveEpic(epicData *epic.Epic, filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to resolve epic file path: %w", err)
	}

	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)

	root := doc.CreateElement("epic")
	root.CreateAttr("id", epicData.ID)
	root.CreateAttr("name", epicData.Name)
	root.CreateAttr("status", string(epicData.Status))
	root.CreateAttr("created_at", epicData.CreatedAt.Format("2006-01-02T15:04:05Z"))

	if epicData.Assignee != "" {
		assigneeElem := root.CreateElement("assignee")
		assigneeElem.SetText(epicData.Assignee)
	}

	if epicData.Description != "" {
		descElem := root.CreateElement("description")
		setInnerXML(descElem, epicData.Description)
	}

	if epicData.Workflow != "" {
		workflowElem := root.CreateElement("workflow")
		setInnerXML(workflowElem, epicData.Workflow)
	}

	if epicData.Requirements != "" {
		requirementsElem := root.CreateElement("requirements")
		setInnerXML(requirementsElem, epicData.Requirements)
	}

	if epicData.Dependencies != "" {
		dependenciesElem := root.CreateElement("dependencies")
		setInnerXML(dependenciesElem, epicData.Dependencies)
	}

	// Save metadata section (Epic 7)
	if epicData.Metadata != nil {
		metadataElem := root.CreateElement("metadata")

		if !epicData.Metadata.Created.IsZero() {
			createdElem := metadataElem.CreateElement("created")
			createdElem.SetText(epicData.Metadata.Created.Format(time.RFC3339))
		}

		if epicData.Metadata.Assignee != "" {
			assigneeElem := metadataElem.CreateElement("assignee")
			assigneeElem.SetText(epicData.Metadata.Assignee)
		}

		if epicData.Metadata.EstimatedEffort != "" {
			effortElem := metadataElem.CreateElement("estimated_effort")
			effortElem.SetText(epicData.Metadata.EstimatedEffort)
		}
	}

	// Save current_state section (Epic 7)
	if epicData.CurrentState != nil {
		currentStateElem := root.CreateElement("current_state")

		if epicData.CurrentState.ActivePhase != "" {
			activePhaseElem := currentStateElem.CreateElement("active_phase")
			activePhaseElem.SetText(epicData.CurrentState.ActivePhase)
		}

		if epicData.CurrentState.ActiveTask != "" {
			activeTaskElem := currentStateElem.CreateElement("active_task")
			activeTaskElem.SetText(epicData.CurrentState.ActiveTask)
		}

		if epicData.CurrentState.NextAction != "" {
			nextActionElem := currentStateElem.CreateElement("next_action")
			nextActionElem.SetText(epicData.CurrentState.NextAction)
		}
	}

	if len(epicData.Phases) > 0 {
		phasesElem := root.CreateElement("phases")
		for _, phase := range epicData.Phases {
			phaseElem := phasesElem.CreateElement("phase")
			phaseElem.CreateAttr("id", phase.ID)
			phaseElem.CreateAttr("name", phase.Name)
			phaseElem.CreateAttr("status", string(phase.Status))
			if phase.Description != "" {
				descElem := phaseElem.CreateElement("description")
				setInnerXML(descElem, phase.Description)
			}
			// Add timestamp elements
			if phase.StartedAt != nil {
				startedElem := phaseElem.CreateElement("started_at")
				startedElem.SetText(phase.StartedAt.Format(time.RFC3339))
			}
			if phase.CompletedAt != nil {
				completedElem := phaseElem.CreateElement("completed_at")
				completedElem.SetText(phase.CompletedAt.Format(time.RFC3339))
			}
		}
	}

	if len(epicData.Tasks) > 0 {
		tasksElem := root.CreateElement("tasks")
		for _, task := range epicData.Tasks {
			taskElem := tasksElem.CreateElement("task")
			taskElem.CreateAttr("id", task.ID)
			taskElem.CreateAttr("phase_id", task.PhaseID)
			taskElem.CreateAttr("name", task.Name)
			taskElem.CreateAttr("status", string(task.Status))
			if task.Assignee != "" {
				taskElem.CreateAttr("assignee", task.Assignee)
			}
			if task.Description != "" {
				descElem := taskElem.CreateElement("description")
				setInnerXML(descElem, task.Description)
			}
			// Add timestamp elements
			if task.StartedAt != nil {
				startedElem := taskElem.CreateElement("started_at")
				startedElem.SetText(task.StartedAt.Format(time.RFC3339))
			}
			if task.CompletedAt != nil {
				completedElem := taskElem.CreateElement("completed_at")
				completedElem.SetText(task.CompletedAt.Format(time.RFC3339))
			}
			if task.CancelledAt != nil {
				cancelledElem := taskElem.CreateElement("cancelled_at")
				cancelledElem.SetText(task.CancelledAt.Format(time.RFC3339))
			}
		}
	}

	if len(epicData.Tests) > 0 {
		testsElem := root.CreateElement("tests")
		for _, test := range epicData.Tests {
			testElem := testsElem.CreateElement("test")
			testElem.CreateAttr("id", test.ID)
			testElem.CreateAttr("task_id", test.TaskID)
			if test.PhaseID != "" {
				testElem.CreateAttr("phase_id", test.PhaseID)
			}
			testElem.CreateAttr("name", test.Name)
			testElem.CreateAttr("status", string(test.Status))

			// Epic 4 enhancements - save TestStatus and related fields
			if test.TestStatus != "" {
				testElem.CreateAttr("test_status", string(test.TestStatus))
			}

			// Check if test has any additional fields beyond description
			hasAdditionalFields := test.StartedAt != nil || test.PassedAt != nil || test.FailedAt != nil ||
				test.CancelledAt != nil || test.FailureNote != "" || test.CancellationReason != ""

			// If test only has description, save as inner text for simpler XML format
			// Otherwise, use child elements to avoid conflicts
			if test.Description != "" && !hasAdditionalFields {
				setInnerXML(testElem, test.Description)
			} else if test.Description != "" {
				descElem := testElem.CreateElement("description")
				setInnerXML(descElem, test.Description)
			}

			// Epic 4 timestamp fields
			if test.StartedAt != nil {
				startedElem := testElem.CreateElement("started_at")
				startedElem.SetText(test.StartedAt.Format(time.RFC3339))
			}
			if test.PassedAt != nil {
				passedElem := testElem.CreateElement("passed_at")
				passedElem.SetText(test.PassedAt.Format(time.RFC3339))
			}
			if test.FailedAt != nil {
				failedElem := testElem.CreateElement("failed_at")
				failedElem.SetText(test.FailedAt.Format(time.RFC3339))
			}
			if test.CancelledAt != nil {
				cancelledElem := testElem.CreateElement("cancelled_at")
				cancelledElem.SetText(test.CancelledAt.Format(time.RFC3339))
			}

			// Epic 4 note fields
			if test.FailureNote != "" {
				failureElem := testElem.CreateElement("failure_note")
				setInnerXML(failureElem, test.FailureNote)
			}
			if test.CancellationReason != "" {
				cancellationElem := testElem.CreateElement("cancellation_reason")
				setInnerXML(cancellationElem, test.CancellationReason)
			}
		}
	}

	// Save events
	if len(epicData.Events) > 0 {
		eventsElem := root.CreateElement("events")
		for _, event := range epicData.Events {
			eventElem := eventsElem.CreateElement("event")
			if event.ID != "" {
				eventElem.CreateAttr("id", event.ID)
			}
			if event.Type != "" {
				eventElem.CreateAttr("type", event.Type)
			}
			if !event.Timestamp.IsZero() {
				eventElem.CreateAttr("timestamp", event.Timestamp.Format(time.RFC3339))
			}

			// Store event data as text content
			if event.Data != "" {
				eventElem.SetText(event.Data)
			}
		}
	} else {
		// Create empty events element for consistency
		root.CreateElement("events")
	}

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create epic directory: %w", err)
	}

	// Format XML with proper indentation for better readability and git diffs
	doc.Indent(4)

	tempFile := absPath + ".tmp"
	if err := doc.WriteToFile(tempFile); err != nil {
		return fmt.Errorf("failed to write epic file: %w", err)
	}

	if err := os.Rename(tempFile, absPath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to move epic file: %w", err)
	}

	return nil
}

func (fs *FileStorage) EpicExists(filePath string) bool {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}

	_, err = os.Stat(absPath)
	return err == nil
}

// getInnerXML returns the inner XML content of an element, preserving any inner XML markup
func getInnerXML(elem *etree.Element) string {
	if elem == nil {
		return ""
	}

	// Check if the element contains any child elements (XML markup)
	hasChildElements := false
	for _, child := range elem.Child {
		if _, isElement := child.(*etree.Element); isElement {
			hasChildElements = true
			break
		}
	}

	// If no child elements, use the regular Text() method to preserve plain text formatting
	if !hasChildElements {
		return elem.Text()
	}

	// Create a temporary document to serialize the inner content
	tempDoc := etree.NewDocument()
	tempRoot := tempDoc.CreateElement("temp")

	// Copy all child tokens by recreating them
	for _, child := range elem.Child {
		switch token := child.(type) {
		case *etree.Element:
			// For elements, use Copy() method
			tempRoot.AddChild(token.Copy())
		case *etree.CharData:
			// For character data, create new CharData
			if token.IsCData() {
				tempRoot.AddChild(etree.NewCData(token.Data))
			} else {
				tempRoot.AddChild(etree.NewText(token.Data))
			}
		case *etree.Comment:
			// For comments, create new Comment
			tempRoot.CreateComment(token.Data)
		default:
			// For other token types, try to write and parse
			var buf strings.Builder
			token.WriteTo(&buf, &etree.WriteSettings{})
			tempRoot.AddChild(etree.NewText(buf.String()))
		}
	}

	// Get the XML string without formatting
	xmlStr, _ := tempDoc.WriteToString()

	// Extract content between <temp> and </temp>
	start := strings.Index(xmlStr, "<temp>")
	end := strings.LastIndex(xmlStr, "</temp>")
	if start != -1 && end != -1 {
		start += 6 // len("<temp>")
		content := xmlStr[start:end]

		// Only normalize whitespace when we have mixed content (text + XML)
		// Replace tabs with spaces but preserve newlines
		content = strings.ReplaceAll(content, "\t", " ")
		// Replace multiple spaces with single space (but keep newlines)
		for strings.Contains(content, "  ") {
			content = strings.ReplaceAll(content, "  ", " ")
		}
		content = strings.TrimSpace(content)

		return content
	}

	// Fallback to plain text if something goes wrong
	return elem.Text()
}

// setInnerXML sets the inner XML content of an element, preserving any inner XML markup
func setInnerXML(elem *etree.Element, content string) {
	if elem == nil {
		return
	}

	if content == "" {
		// For empty content, just set empty text
		elem.SetText("")
		return
	}

	// Try to parse as XML first
	tempDoc := etree.NewDocument()
	tempXML := "<temp>" + content + "</temp>"
	err := tempDoc.ReadFromString(tempXML)

	if err != nil {
		// If parsing fails, treat as plain text
		elem.SetText(content)
		return
	}

	// If parsing succeeds, copy the child elements
	tempRoot := tempDoc.SelectElement("temp")
	if tempRoot != nil {
		for _, child := range tempRoot.Child {
			switch token := child.(type) {
			case *etree.Element:
				elem.AddChild(token.Copy())
			case *etree.CharData:
				if token.IsCData() {
					elem.AddChild(etree.NewCData(token.Data))
				} else {
					elem.AddChild(etree.NewText(token.Data))
				}
			case *etree.Comment:
				elem.CreateComment(token.Data)
			}
		}
	}
}
