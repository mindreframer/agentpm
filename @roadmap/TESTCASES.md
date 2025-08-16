# AgentPM CLI Tool - Test Scenarios by Epic

## Epic 1: Foundation & Configuration - Test Scenarios

### Configuration Management Tests

#### Test: Initialize new project with epic file
**GIVEN** I am in an empty directory  
**WHEN** I run `agentpm init --epic epic-8.xml`  
**THEN** a `.agentpm.json` file should be created with `current_epic: "epic-8.xml"`

#### Test: Initialize project with existing config file
**GIVEN** I have an existing `.agentpm.json` with `current_epic: "epic-7.xml"`  
**WHEN** I run `agentpm init --epic epic-8.xml`  
**THEN** the config should be updated to `current_epic: "epic-8.xml"`

#### Test: Show current configuration
**GIVEN** I have a `.agentpm.json` with project configuration  
**WHEN** I run `agentpm config`  
**THEN** I should see the current project name, epic file, and default assignee

#### Test: Configuration with missing epic file
**GIVEN** I have a config pointing to `epic-missing.xml` that doesn't exist  
**WHEN** I run `agentpm config`  
**THEN** I should see a warning that the epic file is missing

### Epic XML Validation Tests

#### Test: Valid epic XML structure
**GIVEN** I have a well-formed epic XML with all required elements  
**WHEN** I run `agentpm validate`  
**THEN** validation should pass with success message

#### Test: Epic XML with missing required attributes
**GIVEN** I have an epic XML missing the `id` attribute  
**WHEN** I run `agentpm validate`  
**THEN** validation should fail with specific error about missing id

#### Test: Epic XML with invalid status values
**GIVEN** I have an epic XML with status "invalid_status"  
**WHEN** I run `agentpm validate`  
**THEN** validation should fail with enum value error

#### Test: Epic XML with malformed structure
**GIVEN** I have an epic XML with unclosed tags  
**WHEN** I run `agentpm validate`  
**THEN** validation should fail with XML parsing error

### File Operation Tests

#### Test: Validate with specific epic file
**GIVEN** I have multiple epic files in my directory  
**WHEN** I run `agentpm validate -f epic-9.xml`  
**THEN** validation should run on epic-9.xml instead of current epic

#### Test: Command with non-existent epic file
**GIVEN** I specify a non-existent epic file  
**WHEN** I run `agentpm validate -f missing-epic.xml`  
**THEN** I should get a clear file not found error

---

## Epic 2: Query & Status Commands - Test Scenarios

### Status Query Tests

#### Test: Show epic status with progress
**GIVEN** I have an epic with 2 completed phases and 4 total phases  
**WHEN** I run `agentpm status`  
**THEN** I should see status as "in_progress" with 50% completion

#### Test: Show epic status for completed epic
**GIVEN** I have an epic with all phases and tests completed  
**WHEN** I run `agentpm status`  
**THEN** I should see status as "completed" with 100% completion

#### Test: Show epic status with failing tests
**GIVEN** I have an epic with 5 passing tests and 2 failing tests  
**WHEN** I run `agentpm status`  
**THEN** I should see passing_tests: 5 and failing_tests: 2

### Current State Tests

#### Test: Show current active work
**GIVEN** I have an epic with active phase "2A" and active task "2A_1"  
**WHEN** I run `agentpm current`  
**THEN** I should see active_phase: "2A" and active_task: "2A_1"

#### Test: Show current state with no active work
**GIVEN** I have an epic that is started but no phase is active  
**WHEN** I run `agentpm current`  
**THEN** I should see epic_status: "in_progress" with no active phase or task

#### Test: Show next action recommendation
**GIVEN** I have failing tests in current phase  
**WHEN** I run `agentpm current`  
**THEN** I should see next_action suggesting to fix the failing tests

### Pending Work Tests

#### Test: List pending tasks across phases
**GIVEN** I have tasks in status "pending" across multiple phases  
**WHEN** I run `agentpm pending`  
**THEN** I should see all pending tasks grouped by phase

#### Test: Show pending tasks when all completed
**GIVEN** I have an epic with all tasks completed  
**WHEN** I run `agentpm pending`  
**THEN** I should see an empty pending tasks list

### Failing Tests Tests

#### Test: Show only failing tests with details
**GIVEN** I have tests with mixed pass/fail status  
**WHEN** I run `agentpm failing`  
**THEN** I should only see tests with status "failing" and their failure notes

#### Test: Show failing tests when all passing
**GIVEN** I have an epic with all tests passing  
**WHEN** I run `agentpm failing`  
**THEN** I should see an empty failing tests list

### Events Query Tests

#### Test: Show recent events with limit
**GIVEN** I have 10 events in my epic history  
**WHEN** I run `agentpm events --limit=3`  
**THEN** I should see only the 3 most recent events

#### Test: Show events in chronological order
**GIVEN** I have events with different timestamps  
**WHEN** I run `agentpm events`  
**THEN** events should be ordered from most recent to oldest

---

## Epic 3: Epic Lifecycle Management - Test Scenarios

### Epic Startup Tests

#### Test: Start epic from planning status
**GIVEN** I have an epic with status "planning"  
**WHEN** I run `agentpm start-epic`  
**THEN** the epic status should change to "in_progress" and a start event should be logged

#### Test: Start epic that is already started
**GIVEN** I have an epic with status "in_progress"  
**WHEN** I run `agentpm start-epic`  
**THEN** I should get an error that epic is already started

#### Test: Start epic with invalid status
**GIVEN** I have an epic with status "completed"  
**WHEN** I run `agentpm start-epic`  
**THEN** I should get an error that completed epics cannot be restarted

### Epic Pause/Resume Tests

#### Test: Pause epic that is in progress
**GIVEN** I have an epic with status "in_progress"  
**WHEN** I run `agentpm pause-epic "Waiting for design approval"`  
**THEN** status should change to "paused" and reason should be logged in events

#### Test: Pause epic without reason
**GIVEN** I have an epic with status "in_progress"  
**WHEN** I run `agentpm pause-epic`  
**THEN** status should change to "paused" with no reason logged

#### Test: Resume paused epic
**GIVEN** I have an epic with status "paused"  
**WHEN** I run `agentpm resume-epic`  
**THEN** status should change to "in_progress" and resume event should be logged

#### Test: Resume epic that is not paused
**GIVEN** I have an epic with status "in_progress"  
**WHEN** I run `agentpm resume-epic`  
**THEN** I should get an error that epic is not paused

### Epic Completion Tests

#### Test: Complete epic with all work done
**GIVEN** I have an epic with all phases completed and all tests passing  
**WHEN** I run `agentpm complete-epic`  
**THEN** status should change to "completed" and completion event should be logged

#### Test: Complete epic with pending work
**GIVEN** I have an epic with pending tasks  
**WHEN** I run `agentpm complete-epic`  
**THEN** I should get an error listing the pending work that must be completed

#### Test: Complete epic with failing tests
**GIVEN** I have an epic with failing tests  
**WHEN** I run `agentpm complete-epic`  
**THEN** I should get an error listing the failing tests that must be fixed

### Project Switching Tests

#### Test: Switch to different epic file
**GIVEN** I have current_epic set to "epic-8.xml"  
**WHEN** I run `agentpm switch epic-9.xml`  
**THEN** the config should update to current_epic: "epic-9.xml"

#### Test: Switch to non-existent epic file
**GIVEN** I specify an epic file that doesn't exist  
**WHEN** I run `agentpm switch missing-epic.xml`  
**THEN** I should get an error that the epic file doesn't exist

---

## Epic 4: Task & Phase Management - Test Scenarios

### Phase Management Tests

#### Test: Start first phase of epic
**GIVEN** I have an epic in "in_progress" status with no active phase  
**WHEN** I run `agentpm start-phase 1A`  
**THEN** phase "1A" should change to "in_progress" and become the active phase

#### Test: Start phase when another is active
**GIVEN** I have phase "1A" currently active  
**WHEN** I run `agentpm start-phase 2A`  
**THEN** I should get an error that phase "1A" must be completed first

#### Test: Complete phase with all tasks done
**GIVEN** I have phase "1A" with all tasks completed  
**WHEN** I run `agentpm complete-phase 1A`  
**THEN** phase "1A" should change to "completed" status

#### Test: Complete phase with pending tasks
**GIVEN** I have phase "1A" with pending tasks  
**WHEN** I run `agentpm complete-phase 1A`  
**THEN** I should get an error listing the pending tasks in that phase

### Task Management Tests

#### Test: Start specific task in active phase
**GIVEN** I have phase "2A" active with task "2A_1" pending  
**WHEN** I run `agentpm start-task 2A_1`  
**THEN** task "2A_1" should change to "in_progress" and become the active task

#### Test: Start task in non-active phase
**GIVEN** I have phase "1A" active and phase "2A" pending  
**WHEN** I run `agentpm start-task 2A_1`  
**THEN** I should get an error that phase "2A" is not active

#### Test: Complete active task
**GIVEN** I have task "2A_1" in "in_progress" status  
**WHEN** I run `agentpm complete-task 2A_1`  
**THEN** task "2A_1" should change to "completed" and progress should update

#### Test: Complete task that is not started
**GIVEN** I have task "2A_2" in "pending" status  
**WHEN** I run `agentpm complete-task 2A_2`  
**THEN** I should get an error that task must be started before completion

### Auto-Next Task Tests

#### Test: Auto-select next task in current phase
**GIVEN** I have completed task "2A_1" and task "2A_2" is pending in same phase  
**WHEN** I run `agentpm start-next`  
**THEN** task "2A_2" should be automatically selected and started

#### Test: Auto-select next task in next phase
**GIVEN** I have completed all tasks in phase "1A" and phase "2A" has pending tasks  
**WHEN** I run `agentpm start-next`  
**THEN** the first pending task in phase "2A" should be selected and started

#### Test: Auto-select when no pending tasks
**GIVEN** I have completed all tasks in all phases  
**WHEN** I run `agentpm start-next`  
**THEN** I should get a message that all work is completed

### Progress Calculation Tests

#### Test: Progress calculation with mixed completion
**GIVEN** I have 2 completed tasks out of 8 total tasks  
**WHEN** I check epic progress  
**THEN** completion percentage should be 25%

#### Test: Progress calculation with completed phases
**GIVEN** I have 1 fully completed phase out of 3 total phases  
**WHEN** I check epic progress  
**THEN** I should see 1 completed phase and overall progress reflecting task completion

---

## Epic 5: Test Management & Event Logging - Test Scenarios

### Test Status Management Tests

#### Test: Mark test as passing
**GIVEN** I have test "2A_1" in "pending" status  
**WHEN** I run `agentpm pass-test 2A_1`  
**THEN** test "2A_1" should change to "passed" status and event should be logged

#### Test: Mark test as failing with details
**GIVEN** I have test "2A_2" in "pending" status  
**WHEN** I run `agentpm fail-test 2A_2 "Mobile responsive design not working"`  
**THEN** test should change to "failing" with failure note recorded

#### Test: Update failing test to passing
**GIVEN** I have test "2A_2" in "failing" status  
**WHEN** I run `agentpm pass-test 2A_2`  
**THEN** test should change to "passed" and failure note should be cleared

#### Test: Mark non-existent test
**GIVEN** I specify a test ID that doesn't exist  
**WHEN** I run `agentpm pass-test INVALID_TEST`  
**THEN** I should get an error that test ID is not found

### Event Logging Tests

#### Test: Log simple implementation event
**GIVEN** I am working on a task  
**WHEN** I run `agentpm log "Implemented pagination controls"`  
**THEN** an event should be created with type "implementation" and the message

#### Test: Log event with file changes
**GIVEN** I have made changes to files  
**WHEN** I run `agentpm log "Added pagination component" --files="src/Pagination.js:added,src/styles.css:modified"`  
**THEN** event should include file change metadata

#### Test: Log blocker event
**GIVEN** I encounter a blocking issue  
**WHEN** I run `agentpm log "Need design system tokens" --type=blocker`  
**THEN** event should be created with type "blocker" for easy identification

#### Test: Log event with multiple metadata
**GIVEN** I want to log comprehensive information  
**WHEN** I run `agentpm log "Fixed mobile issue" --type=fix --files="src/Mobile.js:modified"`  
**THEN** event should include type, message, and file metadata

### File Change Tracking Tests

#### Test: Track single file addition
**GIVEN** I specify a file change  
**WHEN** I use `--files="src/NewFile.js:added"`  
**THEN** the file change should be parsed and stored with action "added"

#### Test: Track multiple file changes
**GIVEN** I specify multiple file changes  
**WHEN** I use `--files="src/File1.js:modified,src/File2.js:deleted,src/File3.js:added"`  
**THEN** all file changes should be parsed and stored correctly

#### Test: Invalid file change format
**GIVEN** I specify invalid file change format  
**WHEN** I use `--files="invalid-format"`  
**THEN** I should get an error about expected format "filename:action"

### Event Querying Tests

#### Test: Filter events by type
**GIVEN** I have events of different types in my epic  
**WHEN** I query events with type filter "blocker"  
**THEN** I should only see events with type "blocker"

#### Test: Query recent events with timestamp
**GIVEN** I have events from different time periods  
**WHEN** I query events from the last hour  
**THEN** I should only see events within that timeframe

---

## Epic 6: Handoff & Documentation - Test Scenarios

### Handoff Report Tests

#### Test: Generate comprehensive handoff report
**GIVEN** I have an epic with active work, recent events, and some failing tests  
**WHEN** I run `agentpm handoff`  
**THEN** I should get XML with current state, progress summary, recent events, and blockers

#### Test: Handoff report for completed epic
**GIVEN** I have a completed epic with all work done  
**WHEN** I run `agentpm handoff`  
**THEN** handoff should show completed status with 100% progress and no blockers

#### Test: Handoff report with no recent activity
**GIVEN** I have an epic with no recent events  
**WHEN** I run `agentpm handoff`  
**THEN** handoff should include current state but show empty recent events

#### Test: Handoff report identifies blockers
**GIVEN** I have failing tests and logged blocker events  
**WHEN** I run `agentpm handoff`  
**THEN** handoff should list all blockers from failed tests and blocker events

### Documentation Generation Tests

#### Test: Generate markdown documentation
**GIVEN** I have an epic with phases, tasks, and progress  
**WHEN** I run `agentpm docs`  
**THEN** I should get human-readable markdown with epic overview and status

#### Test: Documentation shows phase progress
**GIVEN** I have phases in different completion states  
**WHEN** I generate documentation  
**THEN** markdown should show which phases are completed, in progress, or pending

#### Test: Documentation includes timeline
**GIVEN** I have events logged throughout epic development  
**WHEN** I generate documentation  
**THEN** markdown should include a timeline of major milestones and activities

#### Test: Documentation with empty epic
**GIVEN** I have a newly created epic with no progress  
**WHEN** I generate documentation  
**THEN** markdown should show epic structure but indicate no progress yet

### Recent Events Summarization Tests

#### Test: Summarize recent events with limit
**GIVEN** I have 20 events in my epic history  
**WHEN** handoff includes recent events with limit 5  
**THEN** only the 5 most recent events should be included

#### Test: Prioritize important event types
**GIVEN** I have recent events of various types  
**WHEN** recent events are summarized  
**THEN** blockers and failures should be prioritized over routine implementation events

#### Test: Recent events chronological order
**GIVEN** I have events from different times  
**WHEN** recent events are included in handoff  
**THEN** they should be ordered from most recent to oldest

### Blocker Identification Tests

#### Test: Extract blockers from failing tests
**GIVEN** I have tests with status "failing" and failure notes  
**WHEN** handoff identifies blockers  
**THEN** failing test details should be included as blockers

#### Test: Extract blockers from logged events
**GIVEN** I have events with type "blocker"  
**WHEN** handoff identifies blockers  
**THEN** blocker events should be listed in the blockers section

#### Test: No blockers in healthy epic
**GIVEN** I have an epic with all tests passing and no blocker events  
**WHEN** handoff identifies blockers  
**THEN** the blockers section should be empty