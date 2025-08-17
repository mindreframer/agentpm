# EPIC-11 SPECIFICATION: Enhanced Show Command with Full Context

## Overview

**Epic ID:** 11  
**Name:** Enhanced Show Command with Full Context  
**Duration:** 2-3 days  
**Status:** pending  
**Priority:** medium  

**Goal:** Enhance the `show` command to display full hierarchical context with complete details for all related entities, providing agents with comprehensive project understanding.

## Business Context

The current `show` command provides compact output that lacks the full context needed for effective agent decision-making. This epic transforms the show command into a powerful context-aware tool that displays complete entity hierarchies with full details including descriptions, deliverables, and acceptance criteria for all related entities.

## User Stories

### Primary User Stories
- **As an agent, I can see full context when showing a task** so that I understand the complete scope including parent phase, sibling tasks, and child tests
- **As an agent, I can see full details for all related entities** so that I have descriptions, deliverables, and acceptance criteria for comprehensive understanding
- **As an agent, I can use a --full flag for enhanced output** so that I can choose between compact and detailed views based on my needs
- **As an agent, I can see hierarchical relationships clearly** so that I understand dependencies and project structure

### Secondary User Stories
- **As an agent, I can see phase progress context** so that I understand how much work is completed vs pending in related phases
- **As an agent, I can see test coverage for tasks** so that I understand testing completeness
- **As an agent, I can see sibling entities** so that I understand parallel work and potential dependencies

## Technical Requirements

### Core Dependencies
- **Existing Show Command:** Build upon current `cmd/show.go` implementation
- **Query Service:** Leverage `internal/query/service.go` for data retrieval
- **Output Formatting:** Extend existing XML formatting patterns
- **CLI Framework:** Integrate with existing global flags and command structure

### Output Format Extensions
- **Full Entity Details:** Include description, deliverables, acceptance_criteria for all entities
- **Hierarchical Structure:** Clear parent-child-sibling relationships
- **Status Information:** Complete status context for all related entities
- **Progress Indicators:** Counts and percentages for completion tracking

## Functional Requirements

### FR-1: Enhanced Task Display with Full Context
**Command:** `agentpm show task 1A_1 --full`

**Behavior:**
- Shows the task with complete details
- Includes parent phase with full information
- Lists all sibling tasks in the same phase with full details
- Shows all child tests with complete information
- Provides progress context for the phase

**Output Format (XML):**
```xml
<task_context id="1A_1" phase_id="1A" status="completed">
    <task_details>
        <name>Initialize Project</name>
        <description>Initialize Go module with required dependencies</description>
        <acceptance_criteria>
            - Go module initializes successfully
            - Required dependencies are added to go.mod
            - Project structure follows Go conventions
        </acceptance_criteria>
        <assignee>agent_claude</assignee>
        <started_at>2025-08-16T14:30:00Z</started_at>
        <completed_at>2025-08-16T15:45:00Z</completed_at>
    </task_details>
    
    <parent_phase id="1A" status="active">
        <name>CLI Framework & Core Structure</name>
        <description>Setup basic CLI structure and initialize project</description>
        <deliverables>
            - Functional CLI framework
            - Project structure established
            - Core dependencies configured
        </deliverables>
        <started_at>2025-08-16T14:30:00Z</started_at>
        <progress>
            <total_tasks>3</total_tasks>
            <completed_tasks>1</completed_tasks>
            <pending_tasks>2</pending_tasks>
            <completion_percentage>33</completion_percentage>
        </progress>
    </parent_phase>
    
    <sibling_tasks>
        <task id="1A_2" status="pending">
            <name>Configure Tools</name>
            <description>Set up development tools and linting configuration</description>
            <acceptance_criteria>
                - golangci-lint configured
                - pre-commit hooks set up
                - IDE configuration provided
            </acceptance_criteria>
        </task>
        <task id="1A_3" status="pending">
            <name>Setup Testing Framework</name>
            <description>Initialize testing framework with basic test structure</description>
            <acceptance_criteria>
                - Test framework configured
                - Example tests created
                - Test coverage reporting enabled
            </acceptance_criteria>
        </task>
    </sibling_tasks>
    
    <child_tests>
        <test id="T1A_1" status="passed">
            <name>Test Project Init</name>
            <description>Verify that project initializes correctly with all dependencies</description>
            <acceptance_criteria>
                - go mod init succeeds
                - go mod tidy runs without errors
                - All dependencies resolve correctly
            </acceptance_criteria>
            <test_status>passed</test_status>
            <started_at>2025-08-16T15:00:00Z</started_at>
            <completed_at>2025-08-16T15:30:00Z</completed_at>
        </test>
        <test id="T1A_2" status="pending">
            <name>Test Dependency Resolution</name>
            <description>Verify all required dependencies are properly resolved</description>
            <acceptance_criteria>
                - No dependency conflicts
                - All imports resolve
                - Version constraints satisfied
            </acceptance_criteria>
            <test_status>pending</test_status>
        </test>
    </child_tests>
</task_context>
```

### FR-2: Enhanced Phase Display with Full Context
**Command:** `agentpm show phase 1A --full`

**Behavior:**
- Shows the phase with complete details
- Lists all tasks in the phase with full information
- Shows all tests for those tasks with complete details
- Provides comprehensive progress tracking
- Includes sibling phases for broader context

**Output Format (XML):**
```xml
<phase_context id="1A" status="active">
    <phase_details>
        <name>CLI Framework & Core Structure</name>
        <description>Setup basic CLI structure and initialize project</description>
        <deliverables>
            - Functional CLI framework with global flags
            - Project structure following Go conventions
            - Core dependencies configured and tested
            - Basic command structure implemented
        </deliverables>
        <started_at>2025-08-16T14:30:00Z</started_at>
    </phase_details>
    
    <progress_summary>
        <total_tasks>3</total_tasks>
        <completed_tasks>1</completed_tasks>
        <active_tasks>0</active_tasks>
        <pending_tasks>2</pending_tasks>
        <cancelled_tasks>0</cancelled_tasks>
        <completion_percentage>33</completion_percentage>
        
        <total_tests>4</total_tests>
        <passed_tests>1</passed_tests>
        <failed_tests>0</failed_tests>
        <pending_tests>3</pending_tests>
        <test_coverage_percentage>25</test_coverage_percentage>
    </progress_summary>
    
    <all_tasks>
        <task id="1A_1" status="completed">
            <name>Initialize Project</name>
            <description>Initialize Go module with required dependencies</description>
            <acceptance_criteria>
                - Go module initializes successfully
                - Required dependencies are added to go.mod
                - Project structure follows Go conventions
            </acceptance_criteria>
            <assignee>agent_claude</assignee>
            <started_at>2025-08-16T14:30:00Z</started_at>
            <completed_at>2025-08-16T15:45:00Z</completed_at>
            
            <tests>
                <test id="T1A_1" status="passed">
                    <name>Test Project Init</name>
                    <description>Verify that project initializes correctly</description>
                    <acceptance_criteria>
                        - go mod init succeeds
                        - go mod tidy runs without errors
                    </acceptance_criteria>
                    <test_status>passed</test_status>
                </test>
                <test id="T1A_2" status="pending">
                    <name>Test Dependency Resolution</name>
                    <description>Verify all required dependencies are properly resolved</description>
                    <acceptance_criteria>
                        - No dependency conflicts
                        - All imports resolve
                    </acceptance_criteria>
                    <test_status>pending</test_status>
                </test>
            </tests>
        </task>
        
        <task id="1A_2" status="pending">
            <name>Configure Tools</name>
            <description>Set up development tools and linting configuration</description>
            <acceptance_criteria>
                - golangci-lint configured with project standards
                - pre-commit hooks set up for code quality
                - IDE configuration provided for consistency
            </acceptance_criteria>
            
            <tests>
                <test id="T1A_3" status="pending">
                    <name>Test Linting Configuration</name>
                    <description>Verify linting rules are properly configured</description>
                    <acceptance_criteria>
                        - golangci-lint runs without errors
                        - Custom rules are applied
                    </acceptance_criteria>
                    <test_status>pending</test_status>
                </test>
            </tests>
        </task>
        
        <task id="1A_3" status="pending">
            <name>Setup Testing Framework</name>
            <description>Initialize testing framework with basic test structure</description>
            <acceptance_criteria>
                - Test framework configured with proper structure
                - Example tests created and passing
                - Test coverage reporting enabled
            </acceptance_criteria>
            
            <tests>
                <test id="T1A_4" status="pending">
                    <name>Test Framework Validation</name>
                    <description>Verify testing framework is properly configured</description>
                    <acceptance_criteria>
                        - Tests can be run with go test
                        - Coverage reporting works
                    </acceptance_criteria>
                    <test_status>pending</test_status>
                </test>
            </tests>
        </task>
    </all_tasks>
    
    <sibling_phases>
        <phase id="1B" status="pending">
            <name>Command Implementation</name>
            <description>Implement core CLI commands and functionality</description>
        </phase>
    </sibling_phases>
</phase_context>
```

### FR-3: Enhanced Test Display with Full Context
**Command:** `agentpm show test T1A_1 --full`

**Behavior:**
- Shows the test with complete details
- Includes parent task with full information
- Shows parent phase context
- Lists sibling tests for the same task
- Provides task and phase progress context

**Output Format (XML):**
```xml
<test_context id="T1A_1" task_id="1A_1" status="passed">
    <test_details>
        <name>Test Project Init</name>
        <description>Verify that project initializes correctly with all dependencies</description>
        <acceptance_criteria>
            - go mod init succeeds without errors
            - go mod tidy runs without errors
            - All dependencies resolve correctly
        </acceptance_criteria>
        <test_status>passed</test_status>
        <started_at>2025-08-16T15:00:00Z</started_at>
        <completed_at>2025-08-16T15:30:00Z</completed_at>
    </test_details>
    
    <parent_task id="1A_1" status="completed">
        <name>Initialize Project</name>
        <description>Initialize Go module with required dependencies</description>
        <acceptance_criteria>
            - Go module initializes successfully
            - Required dependencies are added to go.mod
            - Project structure follows Go conventions
        </acceptance_criteria>
        <assignee>agent_claude</assignee>
        <started_at>2025-08-16T14:30:00Z</started_at>
        <completed_at>2025-08-16T15:45:00Z</completed_at>
    </parent_task>
    
    <parent_phase id="1A" status="active">
        <name>CLI Framework & Core Structure</name>
        <description>Setup basic CLI structure and initialize project</description>
        <deliverables>
            - Functional CLI framework
            - Project structure established
            - Core dependencies configured
        </deliverables>
        <progress>
            <completion_percentage>33</completion_percentage>
            <total_tasks>3</total_tasks>
            <completed_tasks>1</completed_tasks>
        </progress>
    </parent_phase>
    
    <sibling_tests>
        <test id="T1A_2" status="pending">
            <name>Test Dependency Resolution</name>
            <description>Verify all required dependencies are properly resolved</description>
            <acceptance_criteria>
                - No dependency conflicts
                - All imports resolve
                - Version constraints satisfied
            </acceptance_criteria>
            <test_status>pending</test_status>
        </test>
    </sibling_tests>
</test_context>
```

### FR-4: Backward Compatibility with Compact Display
**Command:** `agentpm show task 1A_1` (without --full flag)

**Behavior:**
- Maintains current compact output format
- No breaking changes to existing functionality
- Default behavior remains unchanged

### FR-5: Global Flag Integration
**Command:** `agentpm show task 1A_1 --full --format json --file epic-2.xml`

**Behavior:**
- Supports all existing global flags
- JSON and text output formats for full context
- File override functionality
- Configuration integration

## Non-Functional Requirements

### NFR-1: Performance
- Full context display completes in < 300ms for typical epic files
- Efficient data retrieval to avoid multiple file reads
- Optimized XML generation for large result sets

### NFR-2: Usability
- Clear hierarchical structure in output
- Intuitive --full flag behavior
- Consistent formatting across all entity types
- Helpful progress indicators and summaries

### NFR-3: Compatibility
- No breaking changes to existing show command behavior
- Backward compatible with all current usage patterns
- Consistent with existing CLI patterns and conventions

### NFR-4: Extensibility
- Architecture supports future entity types
- Output format extensible for additional context types
- Plugin architecture for custom context providers

## Data Model Extensions

### Enhanced Context Schema
```xml
<entity_context type="task|phase|test" id="entity_id">
    <entity_details>
        <!-- Full entity information with all fields -->
    </entity_details>
    
    <parent_entity type="phase|task" id="parent_id">
        <!-- Complete parent information -->
    </parent_entity>
    
    <child_entities>
        <!-- All child entities with full details -->
    </child_entities>
    
    <sibling_entities>
        <!-- All sibling entities with full details -->
    </sibling_entities>
    
    <progress_context>
        <!-- Relevant progress information -->
    </progress_context>
</entity_context>
```

### Progress Information Schema
```xml
<progress>
    <total_count>integer</total_count>
    <completed_count>integer</completed_count>
    <active_count>integer</active_count>
    <pending_count>integer</pending_count>
    <cancelled_count>integer</cancelled_count>
    <completion_percentage>integer</completion_percentage>
</progress>
```

## Implementation Approach

### Phase 11A: Command Structure Enhancement (Day 1)
- Add --full flag to existing show command
- Extend command argument parsing
- Implement flag validation and help text
- Maintain backward compatibility

### Phase 11B: Context Data Retrieval (Day 1-2)
- Extend query service for hierarchical data retrieval
- Implement parent/child/sibling relationship mapping
- Add progress calculation utilities
- Optimize data retrieval performance

### Phase 11C: Enhanced Output Formatting (Day 2)
- Implement full context XML output formatting
- Add progress summary calculations
- Create hierarchical structure builders
- Support multiple output formats (XML, JSON, text)

### Phase 11D: Integration and Testing (Day 2-3)
- Integration with existing CLI framework
- Comprehensive test coverage
- Performance optimization
- Documentation and examples

## Acceptance Criteria

### AC-1: Full Task Context Display
- **GIVEN** I have an epic with phases, tasks, and tests
- **WHEN** I run `agentpm show task 1A_1 --full`
- **THEN** I should see the task with parent phase, sibling tasks, and child tests all with complete details

### AC-2: Full Phase Context Display
- **GIVEN** I have a multi-task phase
- **WHEN** I run `agentpm show phase 1A --full`
- **THEN** I should see all tasks and tests in the phase with complete details and progress summary

### AC-3: Full Test Context Display
- **GIVEN** I have tests associated with tasks
- **WHEN** I run `agentpm show test T1A_1 --full`
- **THEN** I should see the test with parent task, parent phase, and sibling tests all with complete details

### AC-4: Backward Compatibility
- **GIVEN** I have existing show command usage
- **WHEN** I run `agentpm show task 1A_1` without --full flag
- **THEN** I should get the same compact output as before

### AC-5: Progress Context
- **GIVEN** I have a phase with mixed task statuses
- **WHEN** I run `agentpm show phase 1A --full`
- **THEN** I should see accurate progress counts and percentages

### AC-6: Complete Details Display
- **GIVEN** I have entities with descriptions, deliverables, and acceptance criteria
- **WHEN** I run any show command with --full flag
- **THEN** I should see all detailed fields for all related entities

### AC-7: Output Format Support
- **GIVEN** I want JSON output for full context
- **WHEN** I run `agentpm show task 1A_1 --full --format json`
- **THEN** I should receive complete context information in JSON format

### AC-8: File Override Integration
- **GIVEN** I have multiple epic files
- **WHEN** I run `agentpm show task 1A_1 --full --file epic-2.xml`
- **THEN** the command should display context from epic-2.xml

## Testing Strategy

### Test Categories
- **Unit Tests (60%):** Context retrieval logic, progress calculations, output formatting
- **Integration Tests (30%):** CLI integration, flag processing, data retrieval
- **End-to-End Tests (10%):** Complete workflows with real epic files

### Test Data Requirements
- **Multi-level Epic Files:** Complex epics with phases, tasks, and tests
- **Various Entity States:** Mixed completion statuses for comprehensive testing
- **Edge Cases:** Empty phases, tasks without tests, completed epics

### Performance Testing
- **Response Time:** Full context display in < 300ms
- **Memory Usage:** Efficient handling of large epic structures
- **Data Retrieval:** Optimized queries for related entities

## Definition of Done

- [ ] All acceptance criteria verified with automated tests
- [ ] --full flag implemented for all show command variants
- [ ] Complete details displayed for all entity types and relationships
- [ ] Progress context accurately calculated and displayed
- [ ] Backward compatibility maintained for existing usage
- [ ] Multiple output formats (XML, JSON, text) working correctly
- [ ] Performance requirements met (< 300ms response time)
- [ ] Test coverage > 85% for new functionality
- [ ] Documentation includes practical examples
- [ ] Integration with existing CLI framework complete

## Dependencies and Risks

### Dependencies
- **Epic 1:** Foundation CLI structure (done)
- **Current Show Command:** Existing show command implementation
- **Query Service:** Current data retrieval capabilities

### Risks
- **Medium Risk:** Performance impact with large epic files and complex hierarchies
- **Low Risk:** Memory usage with extensive context information
- **Low Risk:** Output format complexity affecting readability

### Mitigation Strategies
- Implement efficient data retrieval to minimize file reads
- Add performance monitoring and optimization
- Provide clear output structure with proper indentation
- Create comprehensive test suite with performance benchmarks

## Notes

- This enhancement significantly improves agent context awareness
- Full context display should be the preferred mode for agent interactions
- Consider adding configuration option to make --full the default behavior
- Output structure should support future analysis and reporting tools
- Documentation should emphasize the value of full context for decision making