# Epic 4: Task & Phase Management - Specification

## Overview
**Goal:** Granular work tracking at phase and task levels  
**Duration:** 4-5 days  
**Philosophy:** Intelligent work progression with dependency validation and auto-next selection

## User Stories
1. Start working on a specific phase with dependency validation
2. Begin individual tasks within phases with proper sequencing
3. Automatically pick the next pending task intelligently
4. Mark tasks and phases as completed with progress updates
5. Track progress through complex work plans with dependency awareness

## Technical Requirements
- **Dependencies:** Epic 1 (CLI, storage), Epic 2 (status), Epic 3 (lifecycle)
- **Phase Management:** Phase status tracking with dependency validation
- **Task Management:** Task status tracking within phase context
- **Auto-Next Logic:** Intelligent task selection algorithm
- **Progress Calculation:** Real-time progress updates based on completion
- **Dependency Validation:** Enforce phase and task dependencies

## Phase & Task Status Lifecycle

### Status Values & Transitions
```go
type WorkStatus string

const (
    StatusPending     WorkStatus = "pending"
    StatusInProgress  WorkStatus = "in_progress" 
    StatusCompleted   WorkStatus = "completed"
    StatusBlocked     WorkStatus = "blocked"
    StatusSkipped     WorkStatus = "skipped"
)

type PhaseTransition struct {
    From        WorkStatus
    To          WorkStatus
    Valid       bool
    Condition   string
}

type TaskTransition struct {
    From        WorkStatus  
    To          WorkStatus
    Valid       bool
    Condition   string
}
```

### Phase & Task Data Structures
```go
type Phase struct {
    ID          string     `xml:"id,attr"`
    Name        string     `xml:"name,attr"`
    Status      WorkStatus `xml:"status,attr"`
    StartedAt   time.Time  `xml:"started_at,attr,omitempty"`
    CompletedAt time.Time  `xml:"completed_at,attr,omitempty"`
    Dependencies []string  `xml:"dependencies>dependency,omitempty"`
    Tasks       []Task     `xml:"tasks>task,omitempty"`
}

type Task struct {
    ID          string     `xml:"id,attr"`
    PhaseID     string     `xml:"phase_id,attr"`
    Status      WorkStatus `xml:"status,attr"`
    StartedAt   time.Time  `xml:"started_at,attr,omitempty"`
    CompletedAt time.Time  `xml:"completed_at,attr,omitempty"`
    Dependencies []string  `xml:"dependencies>dependency,omitempty"`
    Description string     `xml:",chardata"`
}
```

### Current State Tracking
```go
type CurrentState struct {
    ActivePhase string `xml:"active_phase,attr,omitempty"`
    ActiveTask  string `xml:"active_task,attr,omitempty"`
    LastUpdated time.Time `xml:"last_updated,attr"`
}
```

## Implementation Phases

### Phase 4A: Phase Dependency Management (1 day)
- Phase dependency validation and enforcement
- Phase prerequisite checking before activation
- Phase status transition validation
- Phase completion criteria validation
- Dependency graph analysis and cycle detection

### Phase 4B: Task Management System (1.5 days)
- Task status tracking within phase context
- Task dependency validation and sequencing
- Task assignment to phases with validation
- Task completion with phase progress updates
- Current task tracking and state management

### Phase 4C: Auto-Next Selection Algorithm (1 day)
- Intelligent next task selection logic
- Priority-based task selection with phase preference
- Dependency-aware task ordering
- Phase progression logic for task selection
- Smart work continuation algorithms

### Phase 4D: Phase Operations Implementation (1 day)
- `agentpm start-phase <phase-id>` command
- `agentpm complete-phase <phase-id>` command
- Phase validation and error handling
- Phase event logging and progress tracking
- Integration with Epic 3 lifecycle events

### Phase 4E: Task Operations Implementation (0.5 days)
- `agentpm start-task <task-id>` command
- `agentpm complete-task <task-id>` command
- `agentpm start-next` command with auto-selection
- Task validation and error handling
- Task event logging and progress updates

## Acceptance Criteria
- ✅ `agentpm start-phase 2A` begins phase work with validation
- ✅ `agentpm start-task 2A_1` starts specific task within active phase
- ✅ `agentpm start-next` intelligently selects next pending task
- ✅ `agentpm complete-task 2A_1` marks task complete and updates progress
- ✅ `agentpm complete-phase 2A` completes phase when all tasks done
- ✅ Auto-next prefers tasks in current phase, then next phase
- ✅ Progress percentage updates automatically

## Dependency Validation Rules

### Phase Dependencies
1. **Phase Prerequisites:** Cannot start phase until dependencies are completed
2. **Phase Ordering:** Phases must generally follow dependency order
3. **Circular Dependencies:** Detect and prevent circular phase dependencies
4. **Completion Blocking:** Cannot complete epic until all phases completed

### Task Dependencies
1. **Phase Context:** Tasks can only be started if their phase is active
2. **Task Prerequisites:** Cannot start task until task dependencies completed
3. **Phase Completion:** Cannot complete phase until all tasks completed
4. **Active Task Limit:** Only one task can be active at a time

### Auto-Next Selection Rules
1. **Current Phase Priority:** Prefer next pending task in current phase
2. **Dependency Order:** Respect task dependencies within phase
3. **Phase Progression:** Move to next phase if current phase complete
4. **No Available Work:** Clear message when all work completed

## Output Examples

### agentpm start-phase 2A
```xml
<phase_started epic="8" phase="2A">
    <phase_name>Create PaginationComponent</phase_name>
    <previous_status>pending</previous_status>
    <new_status>in_progress</new_status>
    <started_at>2025-08-16T14:00:00Z</started_at>
    <message>Started Phase 2A: Create PaginationComponent</message>
</phase_started>
```

### agentpm complete-phase 2A
```xml
<phase_completed epic="8" phase="2A">
    <phase_name>Create PaginationComponent</phase_name>
    <previous_status>in_progress</previous_status>
    <new_status>completed</new_status>
    <completed_at>2025-08-16T16:30:00Z</completed_at>
    <duration>2 hours 30 minutes</duration>
    <tasks_completed>2</tasks_completed>
    <tests_passing>3</tests_passing>
    <message>Phase 2A completed successfully</message>
</phase_completed>
```

### agentpm start-task 2A_1
```xml
<task_started epic="8" task="2A_1">
    <task_description>Create PaginationComponent with Previous/Next controls</task_description>
    <phase_id>2A</phase_id>
    <previous_status>pending</previous_status>
    <new_status>in_progress</new_status>
    <started_at>2025-08-16T14:15:00Z</started_at>
    <message>Started Task 2A_1: Create PaginationComponent with Previous/Next controls</message>
</task_started>
```

### agentpm start-next
```xml
<task_started epic="8" task="2A_2">
    <task_description>Add accessibility features to pagination controls</task_description>
    <phase_id>2A</phase_id>
    <previous_status>pending</previous_status>
    <new_status>in_progress</new_status>
    <started_at>2025-08-16T15:00:00Z</started_at>
    <auto_selected>true</auto_selected>
    <selection_reason>Next pending task in active phase 2A</selection_reason>
    <message>Started Task 2A_2: Add accessibility features to pagination controls (auto-selected)</message>
</task_started>
```

### agentpm complete-task 2A_1
```xml
<task_completed epic="8" task="2A_1">
    <task_description>Create PaginationComponent with Previous/Next controls</task_description>
    <phase_id>2A</phase_id>
    <previous_status>in_progress</previous_status>
    <new_status>completed</new_status>
    <completed_at>2025-08-16T14:45:00Z</completed_at>
    <duration>30 minutes</duration>
    <phase_progress>
        <completed_tasks>1</completed_tasks>
        <total_tasks>2</total_tasks>
        <completion_percentage>50</completion_percentage>
    </phase_progress>
    <message>Task 2A_1 completed successfully</message>
</task_completed>
```

## Auto-Next Selection Algorithm

### Selection Priority Logic
```go
func SelectNextTask(epic *Epic) (*Task, string, error) {
    // 1. Find current active phase
    activePhase := findActivePhase(epic)
    
    // 2. Look for pending tasks in active phase
    if activePhase != nil {
        if task := findNextPendingTask(activePhase); task != nil {
            return task, "Next pending task in active phase", nil
        }
    }
    
    // 3. Look for next phase to start
    if nextPhase := findNextPendingPhase(epic); nextPhase != nil {
        if task := findFirstPendingTask(nextPhase); task != nil {
            // Auto-start the phase first
            startPhase(nextPhase)
            return task, "First task in next phase", nil
        }
    }
    
    // 4. No work available
    return nil, "All work completed", nil
}
```

### Dependency Resolution
```go
func findNextPendingTask(phase *Phase) *Task {
    for _, task := range phase.Tasks {
        if task.Status == StatusPending {
            if areDependenciesMet(task) {
                return &task
            }
        }
    }
    return nil
}

func areDependenciesMet(task *Task) bool {
    for _, depID := range task.Dependencies {
        if depTask := findTaskByID(depID); depTask.Status != StatusCompleted {
            return false
        }
    }
    return true
}
```

## Progress Calculation Engine

### Progress Metrics
```go
type ProgressMetrics struct {
    CompletedPhases     int     `xml:"completed_phases"`
    TotalPhases        int     `xml:"total_phases"`
    CompletedTasks     int     `xml:"completed_tasks"`
    TotalTasks         int     `xml:"total_tasks"`
    PhaseProgress      float64 `xml:"phase_progress"`
    TaskProgress       float64 `xml:"task_progress"`
    OverallProgress    float64 `xml:"overall_progress"`
}

func CalculateProgress(epic *Epic) *ProgressMetrics {
    metrics := &ProgressMetrics{}
    
    // Count phases
    for _, phase := range epic.Phases {
        metrics.TotalPhases++
        if phase.Status == StatusCompleted {
            metrics.CompletedPhases++
        }
    }
    
    // Count tasks across all phases
    for _, phase := range epic.Phases {
        for _, task := range phase.Tasks {
            metrics.TotalTasks++
            if task.Status == StatusCompleted {
                metrics.CompletedTasks++
            }
        }
    }
    
    // Calculate percentages
    if metrics.TotalPhases > 0 {
        metrics.PhaseProgress = float64(metrics.CompletedPhases) / float64(metrics.TotalPhases) * 100
    }
    if metrics.TotalTasks > 0 {
        metrics.TaskProgress = float64(metrics.CompletedTasks) / float64(metrics.TotalTasks) * 100
        metrics.OverallProgress = metrics.TaskProgress // Primary metric
    }
    
    return metrics
}
```

## Validation Rules & Error Handling

### Phase Start Validation
1. **Epic Status:** Epic must be "in_progress"
2. **Phase Dependencies:** All prerequisite phases must be completed
3. **No Active Phase:** Cannot start phase if another phase is active
4. **Phase Status:** Phase must be "pending"

### Task Start Validation
1. **Phase Active:** Task's phase must be active ("in_progress")
2. **Task Dependencies:** All prerequisite tasks must be completed
3. **No Active Task:** Cannot start task if another task is active
4. **Task Status:** Task must be "pending"

### Completion Validation
1. **Phase Completion:** All tasks in phase must be completed
2. **Task Completion:** Task must be "in_progress"
3. **Epic Completion:** All phases must be completed (Epic 3 integration)

### Error Examples
```xml
<error type="dependency_not_met">
    <operation>start-phase</operation>
    <target_phase>2A</target_phase>
    <missing_dependencies>
        <dependency id="1A" status="pending">Foundation Setup</dependency>
    </missing_dependencies>
    <message>Cannot start Phase 2A. Complete Phase 1A first.</message>
</error>

<error type="phase_not_active">
    <operation>start-task</operation>
    <target_task>2A_1</target_task>
    <task_phase>2A</task_phase>
    <phase_status>pending</phase_status>
    <message>Cannot start Task 2A_1. Phase 2A must be started first.</message>
    <suggestion>Run: agentpm start-phase 2A</suggestion>
</error>
```

## Test Scenarios (Key Examples)
- **Phase Management:** Start phases with dependency validation, complete phases with task validation
- **Task Management:** Start tasks within active phases, complete tasks with progress updates
- **Auto-Next Selection:** Intelligent task selection with phase preference and dependency awareness
- **Progress Calculation:** Accurate progress metrics with phase and task completion tracking
- **Dependency Validation:** Enforce phase and task dependencies with clear error messages
- **Current State:** Track active phase and task with proper state management
- **Error Handling:** Comprehensive validation with actionable error messages

## Integration with Previous Epics

### Epic 1 Integration
- **Epic Loading:** Use Epic 1 epic loading for phase/task data access
- **Storage:** Leverage Epic 1 storage abstraction for atomic updates
- **Validation:** Extend Epic 1 validation for phase/task structure

### Epic 2 Integration
- **Status Analysis:** Integrate with Epic 2 progress calculation
- **Current State:** Extend Epic 2 current state with active phase/task
- **Pending Work:** Use Epic 4 logic for Epic 2 pending work queries

### Epic 3 Integration
- **Event Logging:** Create phase/task events using Epic 3 event system
- **Lifecycle:** Integrate with Epic 3 epic lifecycle management
- **State Management:** Use Epic 3 atomic operations for phase/task updates

## Quality Gates
- [ ] All acceptance criteria implemented and tested
- [ ] Phase and task dependency validation enforced
- [ ] Auto-next algorithm selects optimal tasks intelligently
- [ ] Progress calculation updates in real-time accurately
- [ ] Comprehensive error handling with actionable messages

## Performance Considerations
- **Dependency Resolution:** Efficient dependency graph traversal
- **Progress Calculation:** Incremental updates rather than full recalculation
- **Auto-Next Selection:** Fast task selection with minimal epic traversal
- **State Updates:** Atomic operations with minimal file I/O

This specification provides comprehensive phase and task management while maintaining intelligent work progression and dependency validation.