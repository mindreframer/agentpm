# EPIC-1: Foundation & Configuration Implementation Plan
## Test-Driven Development Approach

### Phase 1: CLI Framework & Core Structure + Tests (High Priority)

#### Phase 1A: Setup CLI Framework
- [ ] Initialize Go module with required dependencies
- [ ] Setup `github.com/urfave/cli/v3` framework
- [ ] Create main.go with basic CLI app structure
- [ ] Define global flags (--file, --config, --time, --format, --help)
- [ ] Setup command structure and routing
- [ ] Basic help system foundation
- [ ] Project directory structure creation

#### Phase 1B: Write CLI Framework Tests **IMMEDIATELY AFTER 1A**
Epic 1 Test Scenarios Covered:
- [ ] **Test: CLI app initializes correctly**
- [ ] **Test: Global flags are parsed correctly**
- [ ] **Test: Help command displays usage information**
- [ ] **Test: Invalid commands show appropriate errors**
- [ ] **Test: Version information is displayed**
- [ ] **Test: Command routing works properly**

#### Phase 1C: Configuration Management Foundation
- [ ] Create internal/config package
- [ ] Define Config struct with JSON tags
- [ ] Implement LoadConfig() and SaveConfig() functions
- [ ] Add configuration validation logic
- [ ] Default configuration handling
- [ ] Error handling for missing/invalid config files
- [ ] Configuration file path resolution

#### Phase 1D: Write Configuration Tests **IMMEDIATELY AFTER 1C**
Epic 1 Test Scenarios Covered:
- [ ] **Test: Initialize new project with epic file** (Epic 1 line 7)
- [ ] **Test: Initialize project with existing config file** (Epic 1 line 12)
- [ ] **Test: Show current configuration** (Epic 1 line 17)
- [ ] **Test: Configuration with missing epic file** (Epic 1 line 22)
- [ ] **Test: Configuration file path resolution**
- [ ] **Test: Default configuration values**

### Phase 2: XML Foundation & Storage Abstraction + Tests (High Priority)

#### Phase 2A: Create Storage Interface & XML Foundation
- [ ] Define Storage interface for dependency injection
- [ ] Create internal/storage package
- [ ] Implement FileStorage with etree XML operations
- [ ] Create MemoryStorage for testing
- [ ] Basic Epic struct definition
- [ ] XML marshaling/unmarshaling with etree
- [ ] File existence and path validation
- [ ] Storage factory pattern

#### Phase 2B: Write Storage & XML Tests **IMMEDIATELY AFTER 2A**
Epic 1 Test Scenarios Covered:
- [ ] **Test: Valid epic XML structure** (Epic 1 line 29)
- [ ] **Test: Epic XML with missing required attributes** (Epic 1 line 34)
- [ ] **Test: Epic XML with invalid status values** (Epic 1 line 39)
- [ ] **Test: Epic XML with malformed structure** (Epic 1 line 44)
- [ ] **Test: XML round-trip serialization**
- [ ] **Test: File storage operations**
- [ ] **Test: Memory storage for testing**

#### Phase 2C: Epic Validation System
- [ ] Create internal/epic package
- [ ] Implement Epic.Validate() method
- [ ] XML schema validation logic
- [ ] Status enum validation
- [ ] Required field validation
- [ ] Cross-reference validation (phase/task relationships)
- [ ] Validation error collection and reporting
- [ ] Validation result structure

#### Phase 2D: Write Validation Tests **IMMEDIATELY AFTER 2C**
Epic 1 Test Scenarios Covered:
- [ ] **Test: Validate with specific epic file** (Epic 1 line 51)
- [ ] **Test: Command with non-existent epic file** (Epic 1 line 56)
- [ ] **Test: Epic validation passes for valid structure**
- [ ] **Test: Epic validation fails for invalid structure**
- [ ] **Test: Validation error messages are clear**
- [ ] **Test: Cross-reference validation works**

### Phase 3: Core Commands Implementation + Tests (Medium Priority)

#### Phase 3A: Implement Init Command
- [ ] Create cmd/init.go command
- [ ] Epic file existence validation
- [ ] Configuration file creation logic
- [ ] Atomic file operations
- [ ] Success/error response formatting
- [ ] Integration with Storage interface
- [ ] Command-line argument parsing and validation

#### Phase 3B: Write Init Command Tests **IMMEDIATELY AFTER 3A**
Epic 1 Test Scenarios Covered:
- [ ] **Test: Initialize new project with epic file** (detailed workflow)
- [ ] **Test: Initialize with existing configuration**
- [ ] **Test: Initialize with non-existent epic file**
- [ ] **Test: Initialize with invalid epic file**
- [ ] **Test: Configuration file creation is atomic**
- [ ] **Test: Init command output format**

#### Phase 3C: Implement Config & Validate Commands
- [ ] Create cmd/config.go command
- [ ] Create cmd/validate.go command
- [ ] Configuration display logic with XML output
- [ ] Epic validation with detailed reporting
- [ ] File override support (-f flag)
- [ ] Warning handling for missing files
- [ ] Structured error reporting

#### Phase 3D: Write Config & Validate Tests **IMMEDIATELY AFTER 3C**
Epic 1 Test Scenarios Covered:
- [ ] **Test: Show current configuration** (complete workflow)
- [ ] **Test: Configuration with missing epic file** (warning handling)
- [ ] **Test: Validate with specific epic file** (file override)
- [ ] **Test: Validation passes for valid epic**
- [ ] **Test: Validation fails with detailed errors**
- [ ] **Test: Non-existent file handling**

### Phase 4: Integration & Testing + Polish (Low Priority)

#### Phase 4A: Service Layer & Dependency Injection
- [ ] Create internal/service package
- [ ] EpicService with injected storage
- [ ] Service factory with configuration
- [ ] Error handling standardization
- [ ] Transaction-like operations for file safety
- [ ] Service integration with commands
- [ ] Clean separation of concerns

#### Phase 4B: Write Service Integration Tests **IMMEDIATELY AFTER 4A**
Epic 1 Test Scenarios Covered:
- [ ] **Test: Service layer integrates properly**
- [ ] **Test: Dependency injection works correctly**
- [ ] **Test: Service error handling is consistent**
- [ ] **Test: File operations are atomic**
- [ ] **Test: Service factory configuration**

#### Phase 4C: End-to-End CLI Testing & Polish
- [ ] Integration test suite with real CLI execution
- [ ] Temporary directory isolation for tests
- [ ] CLI output format verification
- [ ] Error message consistency
- [ ] Performance optimization
- [ ] Help system completion with examples
- [ ] Code formatting and linting

#### Phase 4D: Final Testing & Documentation **IMMEDIATELY AFTER 4C**
Epic 1 Test Scenarios Covered:
- [ ] **Test: Full CLI workflow end-to-end**
- [ ] **Test: All commands work in isolation**
- [ ] **Test: Error handling across all commands**
- [ ] **Test: Help system completeness**
- [ ] **Test: Performance requirements met**
- [ ] **Test: Code quality standards met**

## Development Workflow Per Phase

For **EACH** phase:

1. **Implement Code** (Phase XA or XC)
2. **Write Tests IMMEDIATELY** (Phase XB or XD) 
3. **Run Tests & Verify** - All tests must pass
4. **Run Linting/Type Checking** - Code must be clean
5. **NEVER move to next phase with failing tests**

## Epic 1 Specific Considerations

### Dependencies & Requirements
- **Go 1.21+** runtime environment
- **github.com/urfave/cli/v3** for CLI framework
- **github.com/beevik/etree** for XML processing
- **github.com/stretchr/testify** for testing
- **github.com/gkampitakis/go-snaps** for snapshot testing

### Technical Architecture
- **Storage Interface:** FileStorage for production, MemoryStorage for testing
- **Configuration:** JSON-based .agentpm.json in project root
- **XML Processing:** etree for parsing, validation, and XPath queries
- **Dependency Injection:** Interface-based design for testability
- **Error Handling:** Structured XML error responses

### File Structure
```
├── cmd/                     # CLI commands
│   ├── init.go             # Project initialization
│   ├── config.go           # Configuration display
│   └── validate.go         # Epic validation
├── internal/
│   ├── epic/               # Epic business logic
│   │   ├── epic.go         # Epic struct and methods
│   │   └── validation.go   # Validation logic
│   ├── config/             # Configuration management
│   │   └── config.go       # Config struct and operations
│   ├── storage/            # Storage abstraction
│   │   ├── interface.go    # Storage interface
│   │   ├── file.go         # File-based storage
│   │   └── memory.go       # Memory storage for tests
│   └── service/            # Service layer
│       └── epic_service.go # Epic service with DI
├── testdata/               # Test XML files
│   ├── epic-valid.xml
│   ├── epic-invalid.xml
│   └── config-sample.json
└── main.go                 # CLI entry point
```

## Testing Strategy

### Test Categories
- **Unit Tests (80%):** Pure business logic, validation, configuration
- **Integration Tests (15%):** File I/O, XML parsing, service integration
- **CLI Tests (5%):** End-to-end command execution

### Test Isolation
- Each test uses `t.TempDir()` for filesystem isolation
- MemoryStorage for fast unit tests
- No shared state between tests
- Parallel test execution support

### Test Data Management
- Test factories for consistent epic/config creation
- Golden files in testdata/ for complex scenarios
- Snapshot testing for XML output validation

## Benefits of This Approach

✅ **Immediate Feedback** - Catch issues as soon as code is written  
✅ **Working Code** - Each phase delivers tested, working functionality  
✅ **Epic 1 Coverage** - All acceptance criteria covered across phases  
✅ **Incremental Progress** - Working CLI after each phase  
✅ **Risk Mitigation** - Problems caught early, not at the end  
✅ **Quality Assurance** - No untested code makes it to later phases  

## Test Distribution Summary

- **Phase 1 Tests:** 6 scenarios (CLI framework, configuration basics)
- **Phase 2 Tests:** 10 scenarios (XML processing, storage, validation)
- **Phase 3 Tests:** 12 scenarios (Core commands, file operations)
- **Phase 4 Tests:** 6 scenarios (Integration, end-to-end, polish)

**Total: All Epic 1 acceptance criteria and test scenarios covered**

---

## Implementation Status

### EPIC 1: FOUNDATION & CONFIGURATION - PENDING
### Current Status: READY TO START

### Progress Tracking
- [x] Phase 1A: Setup CLI Framework
- [x] Phase 1B: Write CLI Framework Tests
- [x] Phase 1C: Configuration Management Foundation
- [x] Phase 1D: Write Configuration Tests
- [x] Phase 2A: Create Storage Interface & XML Foundation
- [x] Phase 2B: Write Storage & XML Tests
- [x] Phase 2C: Epic Validation System
- [x] Phase 2D: Write Validation Tests
- [x] Phase 3A: Implement Init Command
- [x] Phase 3B: Write Init Command Tests
- [x] Phase 3C: Implement Config & Validate Commands
- [x] Phase 3D: Write Config & Validate Tests
- [ ] Phase 4A: Service Layer & Dependency Injection
- [ ] Phase 4B: Write Service Integration Tests
- [ ] Phase 4C: End-to-End CLI Testing & Polish
- [ ] Phase 4D: Final Testing & Documentation

### Definition of Done
- [ ] All acceptance criteria verified with automated tests
- [ ] Commands execute in < 100ms for typical files
- [ ] Test coverage > 90% for business logic
- [ ] All error cases handled gracefully
- [ ] Help system complete with examples
- [ ] No external dependencies beyond specified libraries
- [ ] Code passes lint and format checks
- [ ] Integration tests verify CLI behavior end-to-end