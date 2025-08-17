# EPIC-10: XML Query System Implementation Plan
## Test-Driven Development Approach

### Phase 1: Query Engine Foundation + Tests (High Priority)

#### Phase 10A: Setup Query Engine Core
- [ ] Create internal/query package
- [ ] Define QueryEngine interface with etree integration
- [ ] Implement XPath expression compilation using etree.CompilePath()
- [ ] Create Query struct with expression, compiled path, and metadata
- [ ] Basic query execution against XML documents
- [ ] Query result collection and structuring
- [ ] Error handling for compilation and execution failures
- [ ] Query validation and syntax checking

#### Phase 10B: Write Query Engine Tests **IMMEDIATELY AFTER 10A**
Epic 10 Test Scenarios Covered:
- [ ] **Test: XPath compilation succeeds for valid expressions**
- [ ] **Test: XPath compilation fails for invalid syntax with clear errors**
- [ ] **Test: Basic element selection queries work** (`//task`, `//phase`, `//test`)
- [ ] **Test: Attribute filtering queries work** (`//task[@status='done']`, `//phase[@id='1A']`)
- [ ] **Test: Complex path expressions work** (`//task[@phase_id='1A']`, `//metadata/assignee`)
- [ ] **Test: Empty result sets are handled gracefully**
- [ ] **Test: Query execution performance meets requirements**

#### Phase 10C: Result Formatting & Output System
- [ ] Create internal/query/formatter package
- [ ] Define QueryResult struct with matches, metadata, and counts
- [ ] Implement XML output formatter for structured results
- [ ] Implement text output formatter for human readability
- [ ] Implement JSON output formatter for programmatic use
- [ ] Result metadata collection (execution time, match count)
- [ ] Error result formatting for consistent CLI output
- [ ] Output format selection and routing

#### Phase 10D: Write Result Formatting Tests **IMMEDIATELY AFTER 10C**
Epic 10 Test Scenarios Covered:
- [ ] **Test: XML output format is valid and structured**
- [ ] **Test: Text output format is readable and informative**
- [ ] **Test: JSON output format is valid and complete**
- [ ] **Test: Empty results format correctly in all formats**
- [ ] **Test: Result metadata is accurate** (count, timing, source file)
- [ ] **Test: Error formatting is consistent across formats**

### Phase 2: CLI Integration & Command Implementation + Tests (High Priority)

#### Phase 2A: Implement Query Command
- [ ] Create cmd/query.go command
- [ ] Integrate with existing CLI framework (urfave/cli/v3)
- [ ] Command-line argument parsing for query expression
- [ ] Global flag integration (--file, --format, --config)
- [ ] Epic file loading from configuration or file override
- [ ] Query engine integration and execution
- [ ] Output format selection and rendering
- [ ] Command help and usage information

#### Phase 2B: Write Query Command Tests **IMMEDIATELY AFTER 2A**
Epic 10 Test Scenarios Covered:
- [ ] **Test: Basic element query execution** (`agentpm query "//task"`)
- [ ] **Test: Attribute filtering works** (`agentpm query "//task[@status='done']"`)
- [ ] **Test: Complex path navigation** (`agentpm query "//task[@phase_id='1A']"`)
- [ ] **Test: Text content extraction** (`agentpm query "//task[@id='1A_1']/description/text()"`)
- [ ] **Test: Metadata queries work** (`agentpm query "//metadata/assignee/text()"`)
- [ ] **Test: File override support** (`agentpm query "//task" -f epic-9.xml`)
- [ ] **Test: Output format selection** (`agentpm query "//task" --format json`)
- [ ] **Test: Command help displays correctly**

#### Phase 2C: Advanced Query Features
- [ ] Support for complex XPath expressions with multiple predicates
- [ ] Attribute value extraction (`//phase/@name`, `//task/@status`)
- [ ] Position-based selection (`//task[1]`, `//phase[last()]`)
- [ ] Text content queries (`//description/text()`, `//metadata/assignee/text()`)
- [ ] Wildcard support (`//epic/*`, `//tasks/*`)
- [ ] Epic structure-specific queries (`//outline/phase`, `//events/event`)
- [ ] Query performance optimization and caching
- [ ] Memory-efficient result processing for large files

#### Phase 2D: Write Advanced Features Tests **IMMEDIATELY AFTER 2C**
Epic 10 Test Scenarios Covered:
- [ ] **Test: Multiple predicate queries work correctly**
- [ ] **Test: Attribute extraction returns proper values** (`//phase/@name`)
- [ ] **Test: Position-based selection is accurate** (`//task[1]`, `//phase[last()]`)
- [ ] **Test: Text content extraction works** (`//metadata/assignee/text()`)
- [ ] **Test: Wildcard patterns match correctly** (`//epic/*`)
- [ ] **Test: Epic structure queries work** (`//outline/phase`, `//events/event`)
- [ ] **Test: Performance requirements are met for large files**
- [ ] **Test: Memory usage stays within bounds**

### Phase 3: Error Handling & Validation + Tests (Medium Priority)

#### Phase 3A: Comprehensive Error Handling
- [ ] Create internal/query/errors package
- [ ] Define query-specific error types and categories
- [ ] XPath syntax error detection and helpful messages
- [ ] XML parsing error handling for invalid epic files
- [ ] File access error handling (missing files, permissions)
- [ ] Query execution error handling (runtime failures)
- [ ] Error position tracking for syntax errors
- [ ] Error suggestion system for common mistakes

#### Phase 3B: Write Error Handling Tests **IMMEDIATELY AFTER 3A**
Epic 10 Test Scenarios Covered:
- [ ] **Test: Query syntax validation with clear error messages**
- [ ] **Test: Invalid XPath expressions show helpful suggestions**
- [ ] **Test: Missing epic files produce appropriate errors**
- [ ] **Test: Malformed XML files are handled gracefully**
- [ ] **Test: Permission errors are caught and reported**
- [ ] **Test: Runtime query failures show context**
- [ ] **Test: Error position tracking works for syntax errors**

#### Phase 3C: Query Validation & Help System
- [ ] Pre-execution query validation
- [ ] Query complexity analysis and warnings
- [ ] Help system with practical query examples
- [ ] Documentation for supported XPath syntax
- [ ] Query pattern library for common use cases
- [ ] Interactive query building suggestions
- [ ] Performance hints for optimization

#### Phase 3D: Write Validation & Help Tests **IMMEDIATELY AFTER 3C**
Epic 10 Test Scenarios Covered:
- [ ] **Test: Query validation catches issues before execution**
- [ ] **Test: Complex query warnings are shown appropriately**
- [ ] **Test: Help system includes useful examples**
- [ ] **Test: Syntax documentation is accurate**
- [ ] **Test: Performance hints are relevant**

### Phase 4: Performance & Polish + Tests (Low Priority)

#### Phase 4A: Performance Optimization
- [ ] Query compilation caching for repeated patterns
- [ ] Result streaming for large result sets
- [ ] Memory usage optimization
- [ ] XPath execution performance tuning
- [ ] Concurrent query processing for multiple files
- [ ] Query result pagination support
- [ ] Performance benchmarking and profiling

#### Phase 4B: Write Performance Tests **IMMEDIATELY AFTER 4A**
Epic 10 Test Scenarios Covered:
- [ ] **Test: Query execution meets < 200ms requirement**
- [ ] **Test: Memory usage stays within limits for large files**
- [ ] **Test: Query caching improves repeated execution**
- [ ] **Test: Result streaming works for large result sets**
- [ ] **Test: Performance benchmarks are maintained**

#### Phase 4C: Integration & Final Testing
- [ ] End-to-end CLI testing with real epic files
- [ ] Integration testing with existing AgentPM commands
- [ ] Cross-platform compatibility testing
- [ ] Documentation completion with examples
- [ ] Code quality improvements and refactoring
- [ ] Final performance validation
- [ ] Security review for XPath injection prevention

#### Phase 4D: Final Testing & Documentation **IMMEDIATELY AFTER 4C**
Epic 10 Test Scenarios Covered:
- [ ] **Test: End-to-end workflows work correctly**
- [ ] **Test: Integration with existing commands is seamless**
- [ ] **Test: Cross-platform functionality verified**
- [ ] **Test: All documentation examples work**
- [ ] **Test: Security measures prevent XPath injection**
- [ ] **Test: Code quality standards are met**

## Development Workflow Per Phase

For **EACH** phase:

1. **Implement Code** (Phase 10A or 10C)
2. **Write Tests IMMEDIATELY** (Phase 10B or 10D) 
3. **Run Tests & Verify** - All tests must pass
4. **Run Linting/Type Checking** - Code must be clean
5. **NEVER move to next phase with failing tests**

## Epic 10 Specific Considerations

### Dependencies & Requirements
- **Epic 1:** Foundation CLI structure and XML handling (COMPLETED)
- **github.com/beevik/etree** for XPath processing and XML manipulation
- **Existing CLI framework** for command integration
- **Configuration system** for epic file resolution

### Technical Architecture
- **Query Engine:** Interface-based design with etree backend
- **Result Processing:** Streaming support for large result sets
- **Output Formatting:** Multiple formats (XML, JSON, text) with consistent structure
- **Error Handling:** Comprehensive validation and helpful error messages
- **Performance:** Caching and optimization for repeated queries

### File Structure
```
├── cmd/
│   └── query.go              # Query command implementation
├── internal/
│   ├── query/                # Query engine package
│   │   ├── engine.go         # Query engine interface and implementation
│   │   ├── result.go         # Query result structures
│   │   ├── formatter.go      # Output formatting
│   │   ├── errors.go         # Query-specific errors
│   │   └── cache.go          # Query compilation caching
│   └── service/
│       └── query_service.go  # Query service with dependency injection
├── testdata/
│   ├── epic-pagination.xml   # Sample epic with phases, tasks, tests
│   ├── epic-done.xml    # Epic with done tasks for testing
│   ├── epic-large.xml        # Large epic for performance testing
│   └── epic-complex.xml      # Complex epic with all element types
└── examples/
    └── query_examples.md     # Practical query examples for epic structure
```

## Testing Strategy

### Test Categories
- **Unit Tests (70%):** Query engine, XPath compilation, result formatting
- **Integration Tests (25%):** CLI integration, file loading, error handling
- **Performance Tests (5%):** Query speed, memory usage, large file handling

### Test Isolation
- Each test uses `t.TempDir()` for filesystem isolation
- Mock epic files for consistent test data
- Query engine tests use in-memory XML documents
- Performance tests use controlled datasets

### Test Data Management
- Sample epic files with realistic structures (tasks, phases, tests, metadata)
- Query test cases covering epic-specific XPath patterns
- Error test cases for comprehensive validation coverage
- Performance benchmarks with measurable targets
- Test epics with various states (pending, wip, done)
- Large epic files for performance testing

## Benefits of This Approach

✅ **Immediate Feedback** - Query engine issues caught during development  
✅ **Working Functionality** - Each phase delivers tested query capabilities  
✅ **Epic 10 Coverage** - All acceptance criteria covered across phases  
✅ **Performance Validated** - Speed and memory requirements verified early  
✅ **Error Handling** - Comprehensive validation and user-friendly messages  
✅ **Integration Ready** - Seamless integration with existing CLI framework  

## Test Distribution Summary

- **Phase 1 Tests:** 13 scenarios (Query engine core, result formatting)
- **Phase 2 Tests:** 16 scenarios (CLI integration, advanced features, epic structure)
- **Phase 3 Tests:** 10 scenarios (Error handling, validation, help)
- **Phase 4 Tests:** 11 scenarios (Performance, integration, polish)

**Total: All Epic 10 acceptance criteria and epic-specific query scenarios covered**

---

## Implementation Status

### EPIC 10: XML QUERY SYSTEM - PENDING
### Current Status: READY TO START (depends on Epic 1 completion)

### Progress Tracking
- [ ] Phase 10A: Setup Query Engine Core
- [ ] Phase 10B: Write Query Engine Tests
- [ ] Phase 10C: Result Formatting & Output System
- [ ] Phase 10D: Write Result Formatting Tests
- [ ] Phase 2A: Implement Query Command
- [ ] Phase 2B: Write Query Command Tests
- [ ] Phase 2C: Advanced Query Features
- [ ] Phase 2D: Write Advanced Features Tests
- [ ] Phase 3A: Comprehensive Error Handling
- [ ] Phase 3B: Write Error Handling Tests
- [ ] Phase 3C: Query Validation & Help System
- [ ] Phase 3D: Write Validation & Help Tests
- [ ] Phase 4A: Performance Optimization
- [ ] Phase 4B: Write Performance Tests
- [ ] Phase 4C: Integration & Final Testing
- [ ] Phase 4D: Final Testing & Documentation

### Definition of Done
- [ ] All acceptance criteria verified with automated tests
- [ ] Query execution completes in < 200ms for typical files
- [ ] Supports all etree XPath capabilities demonstrated in examples
- [ ] Comprehensive error handling with helpful messages
- [ ] Multiple output formats (xml, text, json) working correctly
- [ ] Integration with existing CLI framework complete
- [ ] Test coverage > 85% for query engine
- [ ] Performance benchmarks meet requirements
- [ ] Documentation includes practical query examples

### Dependencies
- **REQUIRED:** etree library XPath features validated
- **INTEGRATION:** Existing CLI framework and global flags