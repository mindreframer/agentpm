package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/beevik/etree"
	"github.com/memomoo/agentpm/internal/epic"
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
		epicData.Description = descElem.Text()
	}

	if phasesElem := root.SelectElement("phases"); phasesElem != nil {
		for _, phaseElem := range phasesElem.SelectElements("phase") {
			phase := epic.Phase{
				ID:     phaseElem.SelectAttrValue("id", ""),
				Name:   phaseElem.SelectAttrValue("name", ""),
				Status: epic.Status(phaseElem.SelectAttrValue("status", "")),
			}
			if descElem := phaseElem.SelectElement("description"); descElem != nil {
				phase.Description = descElem.Text()
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
				task.Description = descElem.Text()
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

			if descElem := testElem.SelectElement("description"); descElem != nil {
				test.Description = descElem.Text()
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
				test.FailureNote = failureElem.Text()
			}
			if cancellationElem := testElem.SelectElement("cancellation_reason"); cancellationElem != nil {
				test.CancellationReason = cancellationElem.Text()
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
		descElem.SetText(epicData.Description)
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
				descElem.SetText(phase.Description)
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
				descElem.SetText(task.Description)
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

			if test.Description != "" {
				descElem := testElem.CreateElement("description")
				descElem.SetText(test.Description)
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
				failureElem.SetText(test.FailureNote)
			}
			if test.CancellationReason != "" {
				cancellationElem := testElem.CreateElement("cancellation_reason")
				cancellationElem.SetText(test.CancellationReason)
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
