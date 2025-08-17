# EPIC-10 SPECIFICATION: XML Query System

## Overview

**Epic ID:** 10  
**Name:** XML Query System  
**Duration:** 2-3 days  
**Status:** pending  
**Priority:** medium  

**Goal:** Implement XPath-like query capabilities for epic XML files using etree's path syntax, enabling agents to extract specific data from complex epic structures.

## Business Context

This epic extends AgentPM with powerful query functionality that allows agents to extract specific information from epic XML files using XPath-like syntax. The query system leverages the etree library's built-in path capabilities to provide flexible data retrieval without requiring full XML parsing knowledge.

## User Stories

### Primary User Stories
- **As an agent, I can query for specific elements using XPath syntax** so that I can extract targeted information from epic files
- **As an agent, I can filter elements by attributes** so that I can find tasks, phases, or tests matching specific criteria
- **As an agent, I can get structured XML output from queries** so that I can process results programmatically
- **As an agent, I can query against any epic file** so that I can analyze different epics without switching context

### Secondary User Stories
- **As an agent, I can get formatted text output for human readability** so that I can understand query results quickly
- **As an agent, I can validate query syntax before execution** so that I receive helpful error messages for malformed queries
- **As an agent, I can use complex XPath expressions** so that I can perform sophisticated data extraction

## Technical Requirements

### Core Dependencies
- **XPath Processing:** `github.com/beevik/etree` path compilation and execution
- **Output Formatting:** Consistent with existing AgentPM output patterns
- **Error Handling:** Comprehensive validation for query syntax and execution

### Query Syntax Support
Based on etree's XPath-like capabilities:
- **Element Selection:** `//task`, `//phase`, `/epic/metadata`
- **Attribute Filtering:** `//task[@status='done']`, `//phase[@id='1A']`
- **Index Selection:** `//task[1]`, `//phase[last()]`
- **Text Content:** `//description/text()`, `//task[@id='2A_1']/description/text()`
- **Wildcards:** `//epic/*`, `/epic/tasks/*`

### Global CLI Integration
- Inherits all global flags from existing CLI: `--file`, `--config`, `--format`, `--time`
- Supports multiple output formats: `text`, `xml`, `json`

## Functional Requirements

### FR-1: Basic Query Execution
**Command:** `agentpm query "//task[@status='done']"`

**Behavior:**
- Loads epic file from configuration or `--file` override
- Compiles and validates XPath expression
- Executes query against XML document
- Returns matching elements in specified format
- Handles empty result sets gracefully

**Output Format (XML):**
```xml
<query_result>
    <query>//task[@status='done']</query>
    <match_count>2</match_count>
    <matches>
        <task id="1A_1" phase_id="1A" status="done">
            <description>Setup CLI framework structure</description>
            <acceptance_criteria>
                - CLI framework initializes properly
                - Global flags are parsed correctly
            </acceptance_criteria>
        </task>
        <task id="2A_1" phase_id="2A" status="done">
            <description>Create reusable pagination component</description>
            <acceptance_criteria>
                - Previous/Next buttons work correctly
                - Mobile responsive design implemented
            </acceptance_criteria>
        </task>
    </matches>
</query_result>
```

**Output Format (Text):**
```
Query: //task[@status='done']
Found 2 matches:

task[id=1A_1, phase_id=1A, status=done]:
  Setup CLI framework structure
  
task[id=2A_1, phase_id=2A, status=done]:
  Create reusable pagination component
```

### FR-2: Attribute-Based Filtering
**Command:** `agentpm query "//phase[@status='pending']/@name"`

**Behavior:**
- Supports attribute value extraction
- Filters elements by multiple attribute conditions
- Returns attribute values or element content as specified

**Output Format (XML):**
```xml
<query_result>
    <query>//phase[@status='pending']/@name</query>
    <match_count>4</match_count>
    <matches>
        <attribute name="name" value="Enhanced Schools Context"/>
        <attribute name="name" value="Create PaginationComponent"/>
        <attribute name="name" value="LiveView Integration"/>
        <attribute name="name" value="Performance Optimization"/>
    </matches>
</query_result>
```

### FR-3: Complex Path Expressions
**Command:** `agentpm query "//task[@phase_id='1A']/description"`

**Behavior:**
- Supports nested element selection
- Handles complex filtering conditions
- Maintains element hierarchy in results

**Output Format (XML):**
```xml
<query_result>
    <query>//task[@phase_id='1A']/description</query>
    <match_count>2</match_count>
    <matches>
        <description>Implement list_schools_paginated with combined filtering logic</description>
        <description>Enhance SchoolCrud with QuickCrud.paginate() integration</description>
    </matches>
</query_result>
```

### FR-4: Query Validation
**Command:** `agentpm query "//invalid[syntax"` (malformed)

**Behavior:**
- Validates XPath syntax before execution
- Provides helpful error messages for common mistakes
- Suggests corrections for simple syntax errors

**Output Format (XML):**
```xml
<error>
    <type>query_syntax_error</type>
    <message>Invalid XPath expression: unclosed bracket</message>
    <query>//invalid[syntax</query>
    <position>17</position>
    <suggestion>Check for missing closing bracket ']'</suggestion>
</error>
```

### FR-5: Empty Result Handling
**Command:** `agentpm query "//nonexistent"`

**Behavior:**
- Handles queries that return no matches
- Provides clear feedback about empty results
- Maintains consistent output format

**Output Format (XML):**
```xml
<query_result>
    <query>//nonexistent</query>
    <match_count>0</match_count>
    <matches/>
    <message>No elements found matching query</message>
</query_result>
```

## Non-Functional Requirements

### NFR-1: Performance
- Query execution completes in < 200ms for typical epic files
- XPath compilation caching for repeated queries
- Efficient memory usage for large result sets

### NFR-2: Usability
- Intuitive XPath syntax following etree conventions
- Clear error messages with helpful suggestions
- Consistent output format across all query types

### NFR-3: Reliability
- Robust error handling for malformed queries
- Graceful handling of invalid XML structures
- Safe execution without modifying source files

### NFR-4: Extensibility
- Query engine design supports future XPath extensions
- Output format extensible for additional data types
- Plugin architecture for custom query functions

## Data Model

### Query Result Schema
```xml
<query_result>
    <query>string (XPath expression)</query>
    <epic_file>string (source file path)</epic_file>
    <match_count>integer</match_count>
    <execution_time_ms>integer</execution_time_ms>
    <matches>
        <!-- Variable content based on query results -->
        <element>...</element>
        <attribute>...</attribute>
        <text>...</text>
    </matches>
    <message>string (optional explanatory text)</message>
</query_result>
```

### Supported XPath Patterns
Based on etree capabilities:
- **Axis Selection:** `//`, `/`, `./`, `../`
- **Element Names:** `task`, `phase`, `test`, `epic`, `metadata`, `description`
- **Wildcards:** `*`, `node()`
- **Predicates:** `[@attr='value']`, `[position()]`, `[condition]`
- **Functions:** `text()`, `@attribute`, `position()`, `last()`

## Error Handling

### Error Categories
1. **Query Syntax Errors:** Invalid XPath expressions, unclosed brackets, malformed predicates
2. **File Access Errors:** Missing epic files, permission issues, invalid XML
3. **Execution Errors:** Runtime query failures, memory issues
4. **Configuration Errors:** Missing config, invalid file paths

### Error Output Format
```xml
<error>
    <type>error_category</type>
    <message>Human-readable error description</message>
    <query>Original query string</query>
    <epic_file>Source file path</epic_file>
    <details>
        <position>Character position in query (if applicable)</position>
        <suggestion>Helpful correction suggestion</suggestion>
    </details>
</error>
```

## Acceptance Criteria

### AC-1: Basic Element Query
- **GIVEN** I have a valid epic XML file
- **WHEN** I run `agentpm query "//task"`
- **THEN** I should see all task elements in the epic

### AC-2: Attribute Filtering
- **GIVEN** I have an epic with tasks in different states
- **WHEN** I run `agentpm query "//task[@status='done']"`
- **THEN** I should see only done tasks

### AC-3: Complex Path Navigation
- **GIVEN** I have a multi-phase epic
- **WHEN** I run `agentpm query "//task[@phase_id='1A']"`
- **THEN** I should see tasks belonging to phase 1A

### AC-4: Text Content Extraction
- **GIVEN** I have tasks with descriptions
- **WHEN** I run `agentpm query "//task[@id='1A_1']/description/text()"`
- **THEN** I should see the text content of that task's description

### AC-4b: Epic Metadata Query
- **GIVEN** I have an epic with metadata
- **WHEN** I run `agentpm query "//metadata/assignee/text()"`
- **THEN** I should see the assignee name

### AC-5: Empty Result Handling
- **GIVEN** I have a valid epic file
- **WHEN** I run `agentpm query "//nonexistent"`
- **THEN** I should get a well-formatted empty result with match_count=0

### AC-6: Query Syntax Validation
- **GIVEN** I provide an invalid XPath expression
- **WHEN** I run `agentpm query "//invalid[syntax"`
- **THEN** I should get a clear syntax error with helpful suggestions

### AC-7: File Override Support
- **GIVEN** I have multiple epic files
- **WHEN** I run `agentpm query "//task" -f epic-9.xml`
- **THEN** the query should execute against epic-9.xml

### AC-8: Output Format Support
- **GIVEN** I have a valid query
- **WHEN** I run `agentpm query "//task" --format json`
- **THEN** I should receive results in JSON format

### AC-9: Test Element Queries
- **GIVEN** I have an epic with test elements
- **WHEN** I run `agentpm query "//test[@phase_id='1A']"`
- **THEN** I should see tests belonging to phase 1A

## Testing Strategy

### Test Categories
- **Unit Tests (70%):** XPath compilation, query execution, result formatting
- **Integration Tests (25%):** File loading, CLI integration, error handling
- **End-to-End Tests (5%):** Complete query workflows with real epic files

### Test Data Requirements
- **Sample Epic Files:** Various epic structures with different elements and attributes
- **Query Test Cases:** Valid and invalid XPath expressions for comprehensive coverage
- **Edge Cases:** Empty files, malformed XML, large result sets

### Performance Testing
- **Query Speed:** Execute 100 common queries in < 1 second total
- **Memory Usage:** Handle result sets up to 1000 elements efficiently
- **File Size:** Support epic files up to 10MB without degradation

## Implementation Phases

### Phase 10A: Core Query Engine (Day 1)
- XPath expression compilation using etree
- Basic query execution against XML documents
- Result set collection and basic formatting
- Error handling for syntax and execution errors

### Phase 10B: CLI Integration (Day 1-2)
- Command definition and argument parsing
- Global flag integration (--file, --format)
- Configuration file integration
- Help system documentation

### Phase 10C: Output Formatting (Day 2)
- XML output format implementation
- Text output format for readability
- JSON output format support
- Result count and metadata inclusion

### Phase 10D: Advanced Features & Testing (Day 2-3)
- Complex XPath expression support
- Performance optimization
- Comprehensive test coverage
- Documentation and examples

## Definition of Done

- [ ] All acceptance criteria verified with automated tests
- [ ] Query execution completes in < 200ms for typical files
- [ ] Supports all etree XPath capabilities demonstrated in examples
- [ ] Comprehensive error handling with helpful messages
- [ ] Multiple output formats (xml, text, json) working correctly
- [ ] Integration with existing CLI framework complete
- [ ] Test coverage > 85% for query engine
- [ ] Performance benchmarks meet requirements
- [ ] Documentation includes practical query examples

## Dependencies and Risks

### Dependencies
- **Epic 1:** Foundation CLI structure and XML handling (done)
- **etree Library:** XPath expression compilation and execution

### Risks
- **Medium Risk:** Complex XPath expressions may not be fully supported by etree
- **Low Risk:** Performance with large epic files or complex queries
- **Low Risk:** Memory usage with large result sets

### Mitigation Strategies
- Research etree XPath limitations early and document supported syntax
- Implement result streaming for large query results
- Add query complexity warnings for potentially slow operations
- Create comprehensive test suite with performance benchmarks

## Notes

- This epic provides the foundation for advanced epic analysis and reporting
- Query syntax should be consistent with standard XPath where possible
- Consider caching compiled queries for frequently used patterns
- Output format should support future integration with analysis tools
- Documentation should include practical examples from real epic structures