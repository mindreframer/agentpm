# agentpm

A lightweight project management tool designed for LLM agent collaboration. Stores project data in structured XML for fast querying and updates, with human-readable markdown documentation auto-generated from the source of truth.

## Core Concept

One agent, one epic, one XML file. The XML contains the complete execution plan created upfront. Agent updates progress and logs events as it works through the plan.

## Project Setup

### Configuration File: `.agentpm.json`
```json
{
  "current_epic": "epic-8.xml",
  "previous_epic": "epic-7.xml",
  "project_name": "MyApp", 
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
agentpm done-epic
agentpm pause-epic
agentpm resume-epic

# Start work
agentpm start-phase 2A
agentpm start-task 2A_1
agentpm start-next                  # Auto-pick next pending task

# Update progress
agentpm done-task 2A_1
agentpm done-phase 2A
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
agentpm done-task 2A_1
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
agentpm done-phase 4B
agentpm done-epic
# Output: Epic 8 completed successfully. Status changed to completed.

agentpm status
# Output: Epic completed. All 4 phases done. 47/47 tests passing.
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
        <event timestamp="2025-08-16T15:00:00Z" type="blocker">
            Need design system tokens for mobile
        </event>
        <event timestamp="2025-08-16T14:45:00Z" type="test_failed">
            Mobile responsive test failing
        </event>
        <event timestamp="2025-08-16T14:30:00Z" type="implementation">
            Implemented basic pagination structure
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
<epic id="8" name="Schools Index Pagination" status="wip" started="2025-08-15T09:00:00Z">
    <metadata>
        <created>2025-08-15T09:00:00Z</created>
        <assignee>agent_claude</assignee>
        <estimated_effort>2-3 days</estimated_effort>
    </metadata>

    <description>
        Implement efficient pagination for the schools index page to handle large datasets and improve performance.
        Replace in-memory school loading with database pagination while
        maintaining existing filtering and search functionality.
    </description>

    <workflow>
        **CRITICAL: Test-Driven Development Approach**

        For **EACH** phase:
        1. **Implement Code** - Complete the implementation tasks
        2. **Write Tests IMMEDIATELY** - Create comprehensive test coverage
        3. **Run Tests Verify** - All tests must pass before proceeding
        4. **Run Linting/Type Checking** - Code must be clean and follow standards
        5. **NEVER move to next phase with failing tests**
    </workflow>

    <requirements>
        **Core Stories:**
        - Replace in-memory school loading with database pagination
        - Add pagination controls with page navigation
        - Maintain URL state for bookmarkable paginated views
        - Preserve existing filtering (status) and search functionality
        - Display pagination metadata (showing X of Y schools)

        **Technical Requirements:**
        - Database-level pagination to handle hundreds of schools
        - URL State Management - Page numbers, filters, and search terms in URL
        - LiveView Integration - Real-time pagination without page reloads
        - Mobile Responsive - Simplified pagination controls on mobile devices
        - QuickCrud Integration - Leverage existing paginate() functionality
    </requirements>

    <dependencies>
        - Epic 1: Database schema (crm_schools table) and QuickCrud system (required)
        - Epic 3: School management LiveView pages and existing filtering (required)
        - Epic 4: Contact management for preloading optimization (optional)
    </dependencies>

    <current_state>
        <active_phase>2A</active_phase>
        <active_task>2A_1</active_task>
        <next_action>Fix mobile responsive pagination controls</next_action>
    </current_state>

    <!-- Quick overview for scanning -->
    <outline>
        <phase id="1A" name="Enhanced Schools Context" status="done" />
        <phase id="2A" name="Create PaginationComponent" status="wip" />
        <phase id="3A" name="LiveView Integration" status="pending" />
        <phase id="4A" name="Performance Optimization" status="pending" />
    </outline>

    <!-- Rich details for each phase -->
    <phases>
        <phase id="1A" name="Enhanced Schools Context" status="done">
            <description>
                Extend MyApp.Schools.Main with paginated functions and database-level pagination
            </description>
            <deliverables>
                - list_schools_paginated function with combined filtering
                - Enhanced SchoolCrud with QuickCrud.paginate() integration
                - Efficient database queries with proper indexing
            </deliverables>
        </phase>
        <phase id="2A" name="Create PaginationComponent" status="wip">
            <description>
                Create reusable pagination component with accessibility
            </description>
            <deliverables>
                - Previous/Next navigation with disabled states
                - Page number display and clickable links
                - Mobile-responsive design with touch-friendly controls
                - Accessibility features (ARIA labels, keyboard navigation)
            </deliverables>
        </phase>
        <phase id="3A" name="LiveView Integration" status="pending">
            <description>
                Integrate pagination component with SchoolsLive.Index
            </description>
            <deliverables>
                - Enhanced SchoolsLive.Index with pagination assigns
                - Event handlers for pagination navigation
                - State management for page changes
                - Loading states during pagination
            </deliverables>
        </phase>
        <phase id="4A" name="Performance Optimization" status="pending">
            <description>
                Optimize performance and add polish features
            </description>
            <deliverables>
                - Database query optimization
                - Pagination metadata caching
                - Error handling for edge cases
                - Mobile responsive improvements
            </deliverables>
        </phase>
    </phases>

    <tasks>
        <task id="1A_1" phase_id="1A" status="done">
            <description>Implement list_schools_paginated with combined filtering logic</description>
            <acceptance_criteria>
                - Function accepts opts, page, and page_size parameters
                - Integrates with existing status and search filtering
                - Returns paginated results with metadata
            </acceptance_criteria>
        </task>
        <task id="1A_2" phase_id="1A" status="done">
            <description>Enhance SchoolCrud with QuickCrud.paginate() integration</description>
            <acceptance_criteria>
                - Efficient LIMIT/OFFSET queries
                - Contact preloading for paginated results
                - Proper indexing for performance
            </acceptance_criteria>
        </task>
        <task id="2A_1" phase_id="2A" status="wip">
            <description>Create PaginationComponent with Previous/Next controls</description>
            <acceptance_criteria>
                - Previous/Next buttons with proper disabled states
                - Current page highlighting
                - Mobile-responsive with 44px+ touch targets
                - Pagination metadata display
            </acceptance_criteria>
        </task>
        <task id="2A_2" phase_id="2A" status="pending">
            <description>Add accessibility features to pagination controls</description>
            <acceptance_criteria>
                - ARIA labels for screen readers
                - Keyboard navigation support
                - Focus management
                - High contrast support
            </acceptance_criteria>
        </task>
    </tasks>

    <tests>
        <test id="1A_1" phase_id="1A" status="passed">
            **GIVEN** I have 100 schools in the database
            **WHEN** I call list_schools_paginated with page=2, page_size=25
            **THEN** I get schools 26-50 with pagination metadata
        </test>
        <test id="1A_2" phase_id="1A" status="passed">
            **GIVEN** I have schools with different statuses
            **WHEN** I call list_schools_by_status_paginated with status=engaged
            **THEN** Only engaged schools are returned with pagination
        </test>
        <test id="2A_1" phase_id="2A" status="passed">
            **GIVEN** I have 100 schools displayed
            **WHEN** I click the "Next" button
            **THEN** I see page 2 and schools 26-50
        </test>
        <test id="2A_2" phase_id="2A" status="cancelled">
            **GIVEN** I'm on mobile device
            **WHEN** I tap pagination controls
            **THEN** They work and are easy to tap (44px+ targets)
        </test>
        <test id="2A_3" phase_id="2A" status="pending">
            **GIVEN** I'm on page 2 of schools
            **WHEN** I refresh the browser
            **THEN** I stay on page 2 with URL showing ?page=2
        </test>
    </tests>

    <events>
        <event timestamp="2025-08-15T09:00:00Z" agent="agent_claude" type="epic_started">
            Started Epic 8: Schools Pagination
        </event>
        <event timestamp="2025-08-15T10:30:00Z" agent="agent_claude" phase_id="1A" type="phase_completed">
            Completed Phase 1A: Enhanced Schools Context

            Result: All context functions implemented and tested
        </event>
        <event timestamp="2025-08-16T14:30:00Z" agent="agent_claude" phase_id="2A" type="implementation">
            Implemented basic pagination controls

            Files: src/components/Pagination.js (added), src/styles/pagination.css (added)
            Result: Basic controls working, all tests passing
        </event>
        <event timestamp="2025-08-16T14:45:00Z" agent="agent_claude" phase_id="2A" type="test_failed">
            Mobile responsive test failing

            Test: 2A_2 - Mobile pagination controls
            Issue: Touch targets too small, need 44px+ minimum
        </event>
        <event timestamp="2025-08-16T15:00:00Z" agent="agent_claude" phase_id="2A" type="blocker">
            Found design system dependency

            Blocker: Need design system tokens for mobile responsive design
        </event>
    </events>

    
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
✅ **Minimal Overhead** - 23 core commands, no orchestration complexity  
✅ **Self-Management** - Agent tracks its own progress and blockers  
✅ **Human Transparency** - Generate readable docs anytime  
✅ **Pause/Resume** - Handle interruptions and context switches gracefully  

**Total commands: 23 core commands for complete agent workflow management**

Perfect for LLM agent self-management and clean handoffs between agents or sessions.