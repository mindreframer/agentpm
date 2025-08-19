```
epic (id: number, name: string, status: enum[pending|wip|done|cancelled], started: datetime)
├── metadata
│   ├── created (datetime, ISO8601)
│   ├── assignee (string)
│   └── estimated_effort (string, free text)
├── description (text, markdown supported)
├── workflow (text, markdown supported)
├── requirements (text, markdown supported) 
├── dependencies (text, markdown supported)
├── current_state
│   ├── active_phase (string, phase id reference)
│   ├── active_task (string, task id reference)
│   └── next_action (string, brief description)
├── outline
│   └── phase* (id: string, name: string, status: enum[pending|wip|done|cancelled])
├── phases
│   └── phase* (id: string, name: string, status: enum[pending|wip|done|cancelled])
│       ├── description (text)
│       └── deliverables (text, markdown list)
├── tasks
│   └── task* (id: string, phase_id: string, status: enum[pending|wip|done|cancelled])
│       ├── description (text)
│       └── acceptance_criteria (text, markdown list)
├── tests
│   └── test* (id: string, phase_id: string, task_id: string, status: enum[pending|wip|passed|failed|cancelled])
│       └── content (text, Given/When/Then format)
└── events
    └── event* (timestamp: datetime, agent: string, type: string, phase_id?: string)
        └── content (text, brief description)
```

**Key Patterns:**
- `*` = can have zero or more instances
- `?` = optional attribute
- `datetime` = ISO8601 format (YYYY-MM-DDTHH:MM:SSZ)
- `enum[]` = restricted values listed in brackets
- References use string IDs that should match existing elements
- Markdown formatting allowed in description/text fields

**Validation Rules:**
- `epic.id` must be unique
- `task.phase_id` must reference existing `phase.id`
- `test.phase_id` must reference existing `phase.id`
- `test.task_id` is **required** and must reference existing `task.id` (orphaned tests are not allowed)
- Status transitions should follow logical progression
- Timestamps should be chronological in events