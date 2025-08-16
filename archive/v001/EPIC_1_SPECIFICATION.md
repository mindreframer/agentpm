# Epic 1: Foundation & Configuration - Detailed Technical Specification

## Overview

**Epic ID:** 1  
**Name:** Foundation & Configuration  
**Duration:** 3-4 days  
**Goal:** Core CLI structure, configuration management, and XML handling  
**Status:** Planning  

## Executive Summary

Epic 1 establishes the foundational architecture for the AgentPM CLI tool. This epic focuses on creating a robust, testable foundation that supports project initialization, configuration management, and epic XML validation. The implementation prioritizes simplicity, dependency injection for testability, and comprehensive error handling.

---

## User Stories & Acceptance Criteria

### Story 1: Project Initialization
**As an agent, I can initialize a new project with an epic file**

#### Acceptance Criteria:
- ✅ `agentpm init --epic epic-8.xml` creates `.agentpm.json` with `current_epic: "epic-8.xml"`
- ✅ Command validates that the specified epic file exists before creating config
- ✅ If `.agentpm.json` already exists, it updates the `current_epic` field
- ✅ Command provides clear success/error messages
- ✅ Help text includes usage examples for agent discovery

#### Technical Implementation:
```go
// Command signature
agentpm init --epic <epic-file>

// Configuration file format (.agentpm.json)
{
    "current_epic": "epic-8.xml",
    "project_name": "AgentPM Project",
    "default_assignee": "agent_claude",
    "created_at": "2025-08-16T15:30:00Z",
    "version": "1.0"
}
```

#### Output Format:
```xml
<init_result>
    <project_created>true</project_created>
    <config_file>.agentpm.json</config_file>
    <current_epic>epic-8.xml</current_epic>
</init_result>
```

---

### Story 2: Configuration Management
**As an agent, I can view current project configuration**

#### Acceptance Criteria:
- ✅ `agentpm config` displays current project configuration in XML format
- ✅ Shows current epic file, project name, and default assignee
- ✅ Warns if referenced epic file doesn't exist
- ✅ Handles missing configuration file gracefully
- ✅ Supports `-f` flag to check specific config files

#### Output Format:
```xml
<config>
    <current_epic>epic-8.xml</current_epic>
    <project_name>MyApp</project_name>
    <default_assignee>agent_claude</default_assignee>
</config>
```

---

### Story 3: Epic XML Validation
**As an agent, I can validate epic XML structure for correctness**

#### Acceptance Criteria:
- ✅ `agentpm validate` checks current epic file structure
- ✅ `agentpm validate -f epic-9.xml` validates specific file
- ✅ Validates XML schema, required attributes, and enum values
- ✅ Reports specific errors with line numbers when possible
- ✅ Returns success message for valid files
- ✅ Checks for circular dependencies and invalid references

#### Validation Rules:
1. **Required Elements:** `<epic>` root with `id`, `name`, `status` attributes
2. **Status Values:** Must be one of: `planning`, `in_progress`, `paused`, `completed`, `cancelled`
3. **Phase Structure:** Valid `<phases>` with unique IDs
4. **Task Structure:** Valid `<tasks>` with proper phase references
5. **Test Structure:** Valid `<tests>` with task/phase associations
6. **XML Wellformedness:** Proper closing tags, valid characters

#### Output Format:
```xml
<validation_result epic="8">
    <valid>true</valid>
    <warnings>
        <warning>Task 2A_2 has no tests defined</warning>
    </warnings>
    <checks_performed>
        <check name="xml_structure">passed</check>
        <check name="phase_dependencies">passed</check>
        <check name="task_phase_mapping">passed</check>
        <check name="test_coverage">warning</check>
    </checks_performed>
    <message>Epic structure is valid with 1 warning</message>
</validation_result>
```

---

### Story 4: Help System
**As an agent, I can discover available commands with comprehensive help**

#### Acceptance Criteria:
- ✅ `agentpm help` shows all available commands
- ✅ `agentpm help <command>` shows detailed command help
- ✅ Each command includes usage examples
- ✅ Help system is designed for agent consumption with clear patterns
- ✅ Global flags (`-f`, `--help`) are documented consistently

---

## Technical Architecture

### 1. Project Structure
```
├── cmd/
│   ├── root.go           # Main CLI app setup
│   ├── init.go           # Init command implementation
│   ├── config.go         # Config command implementation
│   └── validate.go       # Validate command implementation
├── internal/
│   ├── config/
│   │   ├── config.go     # Configuration data structures
│   │   ├── loader.go     # Configuration loading/saving
│   │   └── config_test.go
│   ├── epic/
│   │   ├── epic.go       # Epic data structures
│   │   ├── validator.go  # Epic validation logic
│   │   └── epic_test.go
│   └── storage/
│       ├── interface.go  # Storage abstraction
│       ├── file.go       # File-based storage
│       ├── memory.go     # In-memory storage (testing)
│       └── storage_test.go
├── pkg/
│   └── testutil/
│       ├── factory.go    # Test data factories
│       └── helpers.go    # Test utilities
└── testdata/
    ├── epic-valid.xml
    ├── epic-invalid.xml
    └── config-sample.json
```

### 2. Core Dependencies
- **CLI Framework:** `github.com/urfave/cli/v3`
- **XML Processing:** `github.com/beevik/etree`
- **Testing:** `github.com/gkampitakis/go-snaps` for snapshot testing
- **Testing:** `github.com/stretchr/testify` for assertions

### 3. Data Structures

#### Configuration Structure
```go
type Config struct {
    CurrentEpic     string    `json:"current_epic"`
    ProjectName     string    `json:"project_name"`
    DefaultAssignee string    `json:"default_assignee"`
    CreatedAt       time.Time `json:"created_at"`
    Version         string    `json:"version"`
}
```

#### Epic Structure (Foundation)
```go
type Epic struct {
    ID        string    `xml:"id,attr"`
    Name      string    `xml:"name,attr"`
    Status    string    `xml:"status,attr"`
    CreatedAt time.Time `xml:"created_at,attr"`
    
    // Extended structures will be added in later epics
    Phases []Phase `xml:"phases>phase"`
    Tasks  []Task  `xml:"tasks>task"`
    Tests  []Test  `xml:"tests>test"`
    Events []Event `xml:"events>event"`
}
```

#### Storage Interface
```go
type Storage interface {
    LoadConfig(path string) (*Config, error)
    SaveConfig(path string, config *Config) error
    ConfigExists(path string) bool
    
    LoadEpic(path string) (*Epic, error)
    EpicExists(path string) bool
    
    // File operations
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte) error
}
```

### 4. Error Handling Strategy

#### Error Types
```go
type ValidationError struct {
    Field   string
    Rule    string
    Message string
    Line    int // XML line number if available
}

type ConfigError struct {
    Type    string // "missing", "invalid", "permission"
    Path    string
    Message string
}

type FileError struct {
    Operation string // "read", "write", "create"
    Path      string
    Cause     error
}
```

#### Error Messages
- **Clear Context:** Always include what was being attempted
- **Actionable:** Suggest specific next steps
- **Agent-Friendly:** Structured format for automated processing
- **Line Numbers:** Include XML line numbers in validation errors

---

## Implementation Phases

### Phase 1A: Core CLI Setup (0.5 days)
**Deliverables:**
- Basic CLI application structure using urfave/cli/v3
- Global flags (`-f`, `--help`, `--version`)
- Command registration framework
- Basic help system

**Tasks:**
- 1A_1: Set up CLI framework and project structure
- 1A_2: Implement global flags and help system
- 1A_3: Create command registration pattern

**Tests:**
- CLI framework initialization
- Global flag parsing
- Help system completeness

### Phase 1B: Configuration Management (1 day)
**Deliverables:**
- Configuration data structures
- `.agentpm.json` loading and saving
- `agentpm config` command
- `agentpm init` command

**Tasks:**
- 1B_1: Create Config struct and JSON serialization
- 1B_2: Implement configuration file operations
- 1B_3: Build `agentpm config` command
- 1B_4: Build `agentpm init` command

**Tests:**
- Configuration creation and loading
- JSON serialization round-trip
- Command output validation
- Error handling for missing/invalid configs

### Phase 1C: Epic XML Foundation (1 day)
**Deliverables:**
- Basic Epic data structures
- XML parsing using etree
- Epic loading (read-only for now)
- File existence checking

**Tasks:**
- 1C_1: Create Epic struct with XML tags
- 1C_2: Implement XML parsing with etree
- 1C_3: Create epic loading functionality
- 1C_4: Add file existence validation

**Tests:**
- XML parsing round-trip tests
- Epic structure validation
- File loading error cases
- XML malformation handling

### Phase 1D: Epic Validation (1 day)
**Deliverables:**
- Epic validation engine
- `agentpm validate` command
- Comprehensive validation rules
- Detailed error reporting

**Tasks:**
- 1D_1: Create validation rule engine
- 1D_2: Implement structural validation rules
- 1D_3: Build `agentpm validate` command
- 1D_4: Add detailed error reporting with line numbers

**Tests:**
- Valid epic validation
- Invalid epic detection
- Error message accuracy
- Edge case handling

### Phase 1E: Storage Abstraction & Testing (0.5 days)
**Deliverables:**
- Storage interface abstraction
- File-based storage implementation
- In-memory storage for testing
- Dependency injection setup

**Tasks:**
- 1E_1: Define Storage interface
- 1E_2: Create FileStorage implementation
- 1E_3: Create MemoryStorage for testing
- 1E_4: Update commands to use dependency injection

**Tests:**
- Storage interface compliance
- File operations with isolation
- Memory storage functionality
- Dependency injection validation

---

## Testing Strategy

### Test Categories

#### Unit Tests (80% coverage target)
- **Focus:** Pure business logic without I/O
- **Examples:** Configuration validation, epic parsing, validation rules
- **Execution:** In-memory operations, < 1ms per test
- **Isolation:** No file system dependencies

#### Component Tests (15% coverage target)
- **Focus:** File operations with real I/O
- **Examples:** Configuration loading/saving, epic file parsing
- **Execution:** Isolated temporary directories, < 10ms per test
- **Isolation:** Each test uses `t.TempDir()`

#### Integration Tests (5% coverage target)
- **Focus:** End-to-end CLI command execution
- **Examples:** Full command workflows, output validation
- **Execution:** Complete CLI simulation, < 100ms per test
- **Isolation:** Isolated environments with test data

### Test Data Management

#### Test Factories
```go
// pkg/testutil/factory.go
func NewTestConfig() *config.Config {
    return &config.Config{
        CurrentEpic:     "epic-test.xml",
        ProjectName:     "Test Project",
        DefaultAssignee: "test_agent",
        CreatedAt:       time.Now(),
        Version:         "1.0",
    }
}

func NewTestEpic(id string) *epic.Epic {
    return &epic.Epic{
        ID:        id,
        Name:      fmt.Sprintf("Test Epic %s", id),
        Status:    "planning",
        CreatedAt: time.Now(),
    }
}
```

#### Golden Files
- `testdata/epic-valid.xml` - Well-formed epic for positive tests
- `testdata/epic-invalid.xml` - Malformed epic for validation tests
- `testdata/config-sample.json` - Sample configuration files

### Test Execution
```bash
# Fast unit tests only
go test -short ./...

# All tests including integration
go test ./...

# Parallel execution (all tests are parallel-safe)
go test -parallel 10 ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Quality Gates

### Definition of Done
- [ ] All acceptance criteria implemented and tested
- [ ] Unit test coverage ≥ 90% for business logic
- [ ] Integration tests for all CLI commands
- [ ] All error cases handled with clear messages
- [ ] Help documentation complete and accurate
- [ ] Code review completed
- [ ] Performance benchmarks established

### Performance Targets
- **CLI Startup:** < 50ms cold start
- **Configuration Loading:** < 10ms for typical files
- **Epic Validation:** < 100ms for complex epics
- **Help Display:** < 20ms for any help command

### Security Considerations
- **File Permissions:** Validate file access before operations
- **Path Traversal:** Sanitize all file paths
- **Input Validation:** Validate all user inputs
- **Error Messages:** Don't expose sensitive system information

---

## Risk Assessment & Mitigation

### Technical Risks

#### Risk: XML Parsing Performance
- **Impact:** Medium - Could affect large epic files
- **Probability:** Low - Current epics are small
- **Mitigation:** Benchmark with large files, implement streaming if needed

#### Risk: Configuration File Corruption
- **Impact:** High - Could break agent workflows
- **Probability:** Low - Simple JSON structure
- **Mitigation:** Atomic writes, backup before updates, validation on load

#### Risk: CLI Framework Limitations
- **Impact:** Medium - Could limit future command flexibility
- **Probability:** Low - urfave/cli is mature
- **Mitigation:** Evaluate framework thoroughly, maintain abstraction layer

### Operational Risks

#### Risk: File System Permissions
- **Impact:** High - Commands would fail
- **Probability:** Medium - Various deployment environments
- **Mitigation:** Clear error messages, permission checking, fallback strategies

#### Risk: Concurrent Access
- **Impact:** Medium - Data corruption possible
- **Probability:** Low - Single-agent workflows typical
- **Mitigation:** File locking strategy design (future epic)

---

## Success Metrics

### Functional Metrics
- ✅ 100% of acceptance criteria passing
- ✅ All CLI commands execute without errors
- ✅ Validation catches all known error patterns
- ✅ Help system enables agent discovery

### Quality Metrics
- ✅ Test coverage ≥ 90% for business logic
- ✅ Zero critical security vulnerabilities
- ✅ All error paths tested and documented
- ✅ Code review approval

### Performance Metrics
- ✅ CLI commands complete within performance targets
- ✅ Memory usage reasonable for typical epics
- ✅ No performance regressions from baseline

---

## Dependencies & Assumptions

### External Dependencies
- **Go 1.21+** - Required for language features
- **github.com/urfave/cli/v3** - CLI framework
- **github.com/beevik/etree** - XML processing
- **github.com/stretchr/testify** - Testing assertions

### Internal Dependencies
- None (Epic 1 is foundational)

### Assumptions
1. Epic XML files are typically < 1MB in size
2. Agents primarily work on single epics at a time
3. File system access is available in all deployment environments
4. JSON configuration format is sufficient for current needs
5. XML is the preferred output format for structured data

---

## Future Considerations

### Extensibility Points
- **Storage Interface:** Ready for database backends
- **Validation Engine:** Pluggable validation rules
- **CLI Framework:** Supports command plugins
- **Configuration:** Extensible JSON schema

### Known Limitations
1. No concurrent access protection (addressed in future epic)
2. No configuration migration strategy (future requirement)
3. Limited XML schema validation (can be enhanced)
4. No configuration inheritance (may be needed later)

---

## Appendices

### Appendix A: Command Reference

#### agentpm init
```bash
agentpm init --epic <epic-file>

# Creates .agentpm.json with specified epic as current
# Validates epic file exists before creating config
# Updates existing config if present
```

#### agentpm config
```bash
agentpm config

# Displays current project configuration in XML format
# Shows warnings for missing referenced files
# Supports -f flag for alternate config files
```

#### agentpm validate
```bash
agentpm validate              # Validates current epic
agentpm validate -f epic.xml  # Validates specific file

# Comprehensive XML structure validation
# Reports specific errors with context
# Returns success for valid files
```

### Appendix B: Error Code Reference

| Code | Category | Description |
|------|----------|-------------|
| 1    | Config   | Configuration file not found |
| 2    | Config   | Invalid configuration format |
| 3    | Epic     | Epic file not found |
| 4    | Epic     | Invalid epic XML structure |
| 5    | Epic     | Epic validation failed |
| 10   | File     | File permission denied |
| 11   | File     | File system error |
| 20   | CLI      | Invalid command arguments |
| 21   | CLI      | Missing required flags |

### Appendix C: XML Schema Overview

```xml
<?xml version="1.0" encoding="UTF-8"?>
<epic id="8" name="Schools Index Pagination" status="planning" created_at="2025-08-15T09:00:00Z">
    <!-- Phase structure (detailed in Epic 4) -->
    <phases>
        <phase id="1A" name="Phase Name" status="pending" />
    </phases>
    
    <!-- Task structure (detailed in Epic 4) -->
    <tasks>
        <task id="1A_1" phase_id="1A" status="pending">Task description</task>
    </tasks>
    
    <!-- Test structure (detailed in Epic 5) -->
    <tests>
        <test id="1A_1" task_id="1A_1" status="pending">Test description</test>
    </tests>
    
    <!-- Event structure (detailed in Epic 5) -->
    <events>
        <event timestamp="2025-08-16T15:00:00Z" type="implementation">Event description</event>
    </events>
</epic>
```

---

**Document Version:** 1.0  
**Last Updated:** 2025-08-16  
**Next Review:** Upon Epic 1 completion  
**Owner:** Development Team  
**Stakeholders:** Agent PM Users, Integration Partners