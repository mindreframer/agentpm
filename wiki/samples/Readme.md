# agentpm

A lightweight project management tool designed for LLM agent collaboration. Stores project data in structured XML for fast querying and updates, with human-readable markdown documentation auto-generated from the source of truth.

## Core Concept

One agent, one epic, one XML file. The XML contains the complete execution plan created upfront. Agent updates progress and logs events as it works through the plan.

## Project Setup

### Configuration File: `.agentpm.json`
```json
{
  "current_epic": "epic-8.xml",
  "project_name": "MooCRM", 
  "default_assignee": "agent_claude"
}
```

### Project Initialization

```bash
# Initialize project with epic
agentpm init --epic epic-8.xml
# Creates .agentpm.json with current_epic: "epic-8.xml"

# Switch to different epic
agentpm switch epic-9.xml
# Updates .agentpm.json current_epic

# Show current configuration
agentpm config
```

## Essential Commands

### Basic Operations

```bash
# Work with current epic (from .agentpm.json)
agentpm status
agentpm current

# Epic-level operations
agentpm start-epic
agentpm complete-epic
agentpm pause-epic
agentpm resume-epic

# Start work
agentpm start-phase 2A
agentpm start-task 2A_1
agentpm start-next                  # Auto-pick next pending task

# Update progress
agentpm complete-task 2A_1
agentpm complete-phase 2A
agentpm fail-test 2A_1 "Button not rendering properly"
agentpm pass-test 2A_1

# Log what happened
agentpm log "Implemented pagination controls" --files="src/Pagination.js:added"
agentpm log "Found accessibility issue" --type=issue

# Override with specific file when needed
agentpm status -f epic-9.xml
agentpm start-epic -f epic-10.xml
```

### Quick Queries

```bash
# What am I working on?
agentpm current

# What's broken?
agentpm failing

# What's left to do?
agentpm pending

# Recent activity
agentpm events --limit=5
```

### Agent Handoff

```bash
# Comprehensive status for next agent
agentpm handoff

# Generate human-readable docs
agentpm docs
```

### File Validation

```bash
# Check epic structure
agentpm validate

# Generate documentation
agentpm docs > epic-status.md
```

## Agent Workflow Examples

### Starting a New Epic
```bash
# Initialize project and start epic
agentpm init --epic epic-8.xml
# Creates .agentpm.json with current_epic: "epic-8.xml"

agentpm status
# Output: Epic status: planning

agentpm start-epic
# Output: Epic 8 started. Status changed to in_progress.

agentpm current
# Output: Epic active. No current phase. Use start-phase or start-next.
```

### Daily Work Session
```bash
# Check current state
agentpm current
# Output: Epic in_progress. Phase 1A completed. Phase 2A pending.

# Start working
agentpm start-phase 2A
# Output: Started Phase 2A: Create PaginationComponent

agentpm start-task 2A_1  
# Output: Started Task 2A_1: Implement pagination controls

# Or auto-pick next task
agentpm start-next
# Output: Started Task 2A_1: Implement pagination controls (auto-selected)

# Complete some work
agentpm complete-task 2A_1
agentpm log "Implemented basic pagination structure"
agentpm pass-test 2A_1

# Hit an issue
agentpm fail-test 2A_2 "Mobile responsive design not working"
agentpm log "Need design system tokens for mobile" --type=blocker

# Continue to next task
agentpm start-next
# Output: Started Task 2A_2: Add accessibility features
```

### Working with Multiple Epics
```bash
# Switch to different epic
agentpm switch epic-9.xml
# Updates .agentpm.json current_epic

# Work on the new epic
agentpm current
# Output: Epic 9 status

# Check other epic without switching
agentpm status -f epic-8.xml
# Output: Epic 8 status (without changing current context)
```

### Completing an Epic
```bash
# All phases and tests complete
agentpm complete-phase 4B
agentpm complete-epic
# Output: Epic 8 completed successfully. Status changed to completed.

agentpm status
# Output: Epic completed. All 4 phases done. 47/47 tests passing.
```

### Pausing and Resuming Work
```bash
# Need to pause work (context switch, blocker, etc.)
agentpm pause-epic "Waiting for design system approval"
# Output: Epic 8 paused. Status changed to paused.

# Resume later
agentpm resume-epic
# Output: Epic 8 resumed. Status changed to in_progress.
```

### Agent Handoff
```bash
# Outgoing agent
agentpm handoff
# Output: Comprehensive XML with current status, recent events, blockers

# Incoming agent
agentpm current        # What's active?
agentpm failing        # What's broken?  
agentpm events         # What happened recently?
```

## Output Examples

### Epic Status
```bash
$ agentpm status
```
```xml
<status epic="8">
    <name>Schools Pagination</name>
    <status>in_progress</status>
    <progress>
        <completed_phases>2</completed_phases>
        <total_phases>4</total_phases>
        <passing_tests>12</passing_tests>
        <failing_tests>1</failing_tests>
        <completion_percentage>50</completion_percentage>
    </progress>
    <current_phase>2A</current_phase>
    <current_task>2A_1</current_task>
</status>
```

### Current Status
```bash
$ agentpm current
```
```xml
<current_state epic="8">
    <epic_status>in_progress</epic_status>
    <active_phase>2A</active_phase>
    <active_task>2A_1</active_task>
    <next_action>Fix mobile responsive pagination controls</next_action>
    <failing_tests>1</failing_tests>
</current_state>
```

### Failing Tests
```bash
$ agentpm failing
```
```xml
<failing_tests epic="8">
    <test id="2A_2" phase_id="2A">
        <given>I'm on mobile</given>
        <when>I tap pagination controls</when>
        <then>they work and are easy to tap</then>
        <failure_note>Mobile responsive design not working</failure_note>
    </test>
</failing_tests>
```

### Agent Handoff Report
```bash
$ agentpm handoff
```
```xml
<handoff epic="8" timestamp="2025-08-16T15:30:00Z">
    <epic_info>
        <name>Schools Pagination</name>
        <status>in_progress</status>
        <started>2025-08-15T09:00:00Z</started>
    </epic_info>
    <current_state>
        <active_phase>2A</active_phase>
        <active_task>2A_1</active_task>
        <next_action>Fix mobile responsive pagination controls</next_action>
    </current_state>
    <summary>
        <completed_phases>2</completed_phases>
        <total_phases>4</total_phases>
        <passing_tests>12</passing_tests>
        <failing_tests>1</failing_tests>
        <completion_percentage>50</completion_percentage>
    </summary>
    <recent_events limit="3">
        <event timestamp="2025-08-16T14:30:00Z" type="task_completed">
            <action>Implemented basic pagination structure</action>
        </event>
        <event timestamp="2025-08-16T14:45:00Z" type="test_failed">
            <test_id>2A_2</test_id>
            <note>Mobile responsive design not working</note>
        </event>
        <event timestamp="2025-08-16T15:00:00Z" type="blocker">
            <action>Need design system tokens for mobile</action>
        </event>
    </recent_events>
    <blockers>
        <blocker>Need design system tokens for mobile responsive design</blocker>
    </blockers>
</handoff>
```

## XML Structure

The epic file follows this minimal structure:

```xml
<epic id="8" name="Schools Pagination" status="in_progress" started="2025-08-15T09:00:00Z">
    <outline>
        <phase id="1A" name="Enhanced Schools Context" status="completed" />
        <phase id="2A" name="Create PaginationComponent" status="in_progress" />
        <phase id="3A" name="LiveView Integration" status="pending" />
        <phase id="4A" name="Performance Optimization" status="pending" />
    </outline>
    
    <tasks>
        <task id="1A_1" phase_id="1A" status="completed">
            <description>The Schools Context is enhanced to include pagination.</description>
        </task>
        <task id="2A_1" phase_id="2A" status="in_progress">
            <description>Create PaginationComponent with Previous/Next controls</description>
        </task>
        <task id="2A_2" phase_id="2A" status="pending">
            <description>Add accessibility features to pagination controls</description>
        </task>
    </tasks>

    <tests>
        <test id="1A_1" phase_id="1A" status="passed">
            <given>I have paginated context functions</given>
            <when>I call list_schools_paginated</when>
            <then>I get correct page size</then>
        </test>
        <test id="1A_2" phase_id="1A" status="passed">
            <given>I have 100 schools</given>
            <when>I request page 2</when>
            <then>Database query uses LIMIT/OFFSET</then>
        </test>
        <test id="2A_1" phase_id="2A" status="passed">
            <given>I have 100 schools</given>
            <when>I click "Next"</when>
            <then>I see page 2 and schools 26-50</then>
        </test>
        <test id="2A_2" phase_id="2A" status="failing">
            <given>I'm on mobile</given>
            <when>I tap pagination controls</when>
            <then>they work and are easy to tap</then>
            <failure_note>Mobile responsive design not working</failure_note>
        </test>
    </tests>

    <events>
        <event timestamp="2025-08-15T09:00:00Z" agent="agent_claude" type="epic_started">
            <action>Started Epic 8: Schools Pagination</action>
        </event>
        <event timestamp="2025-08-15T10:30:00Z" agent="agent_claude" phase_id="1A" type="phase_completed">
            <action>Completed Phase 1A: Enhanced Schools Context</action>
            <result status="success">All context functions implemented and tested</result>
        </event>
        <event timestamp="2025-08-16T14:30:00Z" agent="agent_claude" phase_id="2A" type="implementation">
            <action>Implemented basic pagination controls</action>
            <files_changed>
                <file path="src/components/Pagination.js" status="added">Created pagination component</file>
                <file path="src/styles/pagination.css" status="added">Added pagination styles</file>
            </files_changed>
            <result status="success">Basic controls working</result>
        </event>
        <event timestamp="2025-08-16T14:45:00Z" agent="agent_claude" phase_id="2A" type="test_failed">
            <test_id>2A_2</test_id>
            <action>Mobile responsive test failing</action>
            <note>Mobile responsive design not working</note>
        </event>
        <event timestamp="2025-08-16T15:00:00Z" agent="agent_claude" phase_id="2A" type="blocker">
            <action>Found design system dependency</action>
            <note>Need design system tokens for mobile responsive design</note>
        </event>
    </events>

    <current_state>
        <active_phase>2A</active_phase>
        <active_task>2A_1</active_task>
        <next_action>Fix mobile responsive pagination controls</next_action>
    </current_state>
</epic>
```

## Epic Status Lifecycle

- **`planning`** - Epic created but not started
- **`in_progress`** - Agent actively working on epic
- **`paused`** - Work temporarily stopped (blockers, context switch)
- **`completed`** - All phases and tests complete
- **`cancelled`** - Epic abandoned or deprioritized

## Key Benefits

✅ **Single File Focus** - One agent, one epic, one XML file  
✅ **Epic Lifecycle Tracking** - Clear status progression from planning to completion  
✅ **Simple Progress Tracking** - Clear start/complete commands at epic, phase, and task levels  
✅ **Agent Handoff** - Comprehensive context for next agent  
✅ **Minimal Overhead** - ~15 core commands, no orchestration complexity  
✅ **Self-Management** - Agent tracks its own progress and blockers  
✅ **Human Transparency** - Generate readable docs anytime  
✅ **Pause/Resume** - Handle interruptions and context switches gracefully  

**Total commands: 15 core commands for complete agent workflow management**

Perfect for LLM agent self-management and clean handoffs between agents or sessions.