package epic

import (
	"time"
)

type Status string

const (
	StatusPlanning  Status = "planning"
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
	StatusOnHold    Status = "on_hold"
	StatusCancelled Status = "cancelled"
)

type Epic struct {
	ID          string    `xml:"id,attr"`
	Name        string    `xml:"name,attr"`
	Status      Status    `xml:"status,attr"`
	CreatedAt   time.Time `xml:"created_at,attr"`
	Assignee    string    `xml:"assignee"`
	Description string    `xml:"description"`
	Phases      []Phase   `xml:"phases>phase"`
	Tasks       []Task    `xml:"tasks>task"`
	Tests       []Test    `xml:"tests>test"`
	Events      []Event   `xml:"events>event"`
}

type Phase struct {
	ID          string `xml:"id,attr"`
	Name        string `xml:"name,attr"`
	Description string `xml:"description"`
	Status      Status `xml:"status,attr"`
}

type Task struct {
	ID          string `xml:"id,attr"`
	PhaseID     string `xml:"phase_id,attr"`
	Name        string `xml:"name,attr"`
	Description string `xml:"description"`
	Status      Status `xml:"status,attr"`
	Assignee    string `xml:"assignee,attr,omitempty"`
}

type Test struct {
	ID          string `xml:"id,attr"`
	TaskID      string `xml:"task_id,attr"`
	Name        string `xml:"name,attr"`
	Description string `xml:"description"`
	Status      Status `xml:"status,attr"`
}

type Event struct {
	ID        string    `xml:"id,attr"`
	Type      string    `xml:"type,attr"`
	Timestamp time.Time `xml:"timestamp,attr"`
	Data      string    `xml:"data"`
}

func (s Status) IsValid() bool {
	switch s {
	case StatusPlanning, StatusActive, StatusCompleted, StatusOnHold, StatusCancelled:
		return true
	default:
		return false
	}
}

func NewEpic(id, name string) *Epic {
	return &Epic{
		ID:        id,
		Name:      name,
		Status:    StatusPlanning,
		CreatedAt: time.Now(),
	}
}
