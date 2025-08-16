# Epic 2: Query & Status Commands Implementation Plan
## Test-Driven Development Approach

### Phase 1: Epic Status Analysis Engine + Tests (High Priority)

#### Phase 1A: Progress Calculation Engine
- [ ] Create StatusSummary struct with XML serialization
- [ ] Implement phase completion counting logic
- [ ] Add task completion percentage calculation
- [ ] Create test status aggregation (passing/failing counts)
- [ ] Implement weighted progress calculation algorithm
- [ ] Add epic status validation and state detection
- [ ] Create progress calculation utilities and helpers

#### Phase 1B: Write Progress Calculation Tests **IMMEDIATELY AFTER 1A**
Epic 2 Test Scenarios Covered:
- [ ] **Test: Show epic status with progress** (Epic 2 line 67)
- [ ] **Test: Show epic status for completed epic** (Epic 2 line 72)
- [ ] **Test: Show epic status with failing tests** (Epic 2 line 77)
- [ ] **Test: Progress calculation with mixed completion** (Epic 2 line 274)
- [ ] **Test: Progress calculation with completed phases** (Epic 2 line 279)
- [ ] **Test: Status calculation accuracy across different epic states**
- [ ] **Test: Edge cases (empty phases, all complete, no tests)**

#### Phase 1C: Implement `agentpm status` Command
- [ ] Create status command with epic file loading
- [ ] Integrate progress calculation engine
- [ ] Add XML output formatting for status summary
- [ ] Implement `-f` flag support for file override
- [ ] Add error handling for missing/invalid epic files
- [ ] Create comprehensive status display logic

#### Phase 1D: Write `agentpm status` Tests **IMMEDIATELY AFTER 1C**
Epic 2 Test Scenarios Covered:
- [ ] **Test: Status command execution and output format**
- [ ] **Test: File override with `-f` flag functionality**
- [ ] **Test: Error handling for missing epic files**
- [ ] **Test: XML output structure validation**
- [ ] **Test: Status command integration with Epic 1 foundation**

### Phase 2: Current State Intelligence + Tests (High Priority)

#### Phase 2A: Active Work Detection & Next Action Engine
- [ ] Create current state analysis logic
- [ ] Implement active phase and task detection
- [ ] Add next action recommendation algorithm
- [ ] Create failing test impact analysis for recommendations
- [ ] Implement work prioritization and guidance logic
- [ ] Add context-aware next action suggestions

#### Phase 2B: Write Current State Tests **IMMEDIATELY AFTER 2A**
Epic 2 Test Scenarios Covered:
- [ ] **Test: Show current active work** (Epic 2 line 85)
- [ ] **Test: Show current state with no active work** (Epic 2 line 90)
- [ ] **Test: Show next action recommendation** (Epic 2 line 95)
- [ ] **Test: Next action logic for different epic states**
- [ ] **Test: Failing test priority in recommendation algorithm**
- [ ] **Test: Context-aware guidance accuracy**

#### Phase 2C: Implement `agentpm current` Command
- [ ] Create current command with state analysis
- [ ] Integrate active work detection logic
- [ ] Add next action recommendation display
- [ ] Implement XML output for current state
- [ ] Add support for `-f` flag and file override
- [ ] Create current state formatting and display

#### Phase 2D: Write `agentpm current` Tests **IMMEDIATELY AFTER 2C**
Epic 2 Test Scenarios Covered:
- [ ] **Test: Current command execution and state display**
- [ ] **Test: Active work detection accuracy**
- [ ] **Test: Next action recommendation quality**
- [ ] **Test: XML output format and content validation**
- [ ] **Test: Current command error handling and edge cases**

### Phase 3: Pending Work Discovery + Tests (Medium Priority)

#### Phase 3A: Pending Work Enumeration Engine
- [ ] Create pending work discovery algorithms
- [ ] Implement phase dependency analysis
- [ ] Add pending task enumeration across phases
- [ ] Create test status filtering (pending/failing)
- [ ] Implement work ordering by dependencies
- [ ] Add comprehensive pending work categorization

#### Phase 3B: Write Pending Work Tests **IMMEDIATELY AFTER 3A**
Epic 2 Test Scenarios Covered:
- [ ] **Test: List pending tasks across phases** (Epic 2 line 101)
- [ ] **Test: Show pending tasks when all completed** (Epic 2 line 106)
- [ ] **Test: Pending work ordering and categorization**
- [ ] **Test: Phase dependency analysis accuracy**
- [ ] **Test: Work prioritization algorithm correctness**

#### Phase 3C: Implement `agentpm pending` Command
- [ ] Create pending command with work discovery
- [ ] Integrate pending work enumeration logic
- [ ] Add XML output for pending work breakdown
- [ ] Implement categorized display (phases/tasks/tests)
- [ ] Add `-f` flag support and error handling
- [ ] Create pending work formatting and organization

#### Phase 3D: Write `agentpm pending` Tests **IMMEDIATELY AFTER 3C**
Epic 2 Test Scenarios Covered:
- [ ] **Test: Pending command execution and output structure**
- [ ] **Test: Work categorization and organization**
- [ ] **Test: Pending work discovery completeness**
- [ ] **Test: XML output format validation**
- [ ] **Test: Edge cases (no pending work, all phases complete)**

### Phase 4: Failing Tests & Event Queries + Tests (Medium Priority)

#### Phase 4A: Failing Test Detection & Event Query System
- [ ] Create failing test filtering and extraction
- [ ] Implement detailed failure information display
- [ ] Add event chronological ordering system
- [ ] Create event filtering by type, limit, timeframe
- [ ] Implement recent activity summarization
- [ ] Add event metadata extraction and formatting

#### Phase 4B: Write Failing Tests & Events Tests **IMMEDIATELY AFTER 4A**
Epic 2 Test Scenarios Covered:
- [ ] **Test: Show only failing tests with details** (Epic 2 line 113)
- [ ] **Test: Show failing tests when all passing** (Epic 2 line 118)
- [ ] **Test: Show recent events with limit** (Epic 2 line 125)
- [ ] **Test: Show events in chronological order** (Epic 2 line 130)
- [ ] **Test: Event filtering and categorization accuracy**
- [ ] **Test: Failure detail extraction and display**

#### Phase 4C: Implement `agentpm failing` and `agentpm events` Commands
- [ ] Create failing command with test filtering
- [ ] Implement events command with timeline functionality
- [ ] Add detailed failure information display
- [ ] Create event limit and filtering support (--limit flag)
- [ ] Implement XML output for both commands
- [ ] Add comprehensive error handling and edge cases

#### Phase 4D: Write Command Integration Tests **IMMEDIATELY AFTER 4C**
Epic 2 Test Scenarios Covered:
- [ ] **Test: Failing command execution and test detection**
- [ ] **Test: Events command with timeline and filtering**
- [ ] **Test: Event limit parameter functionality**
- [ ] **Test: XML output consistency across commands**
- [ ] **Test: Command integration with Epic 1 systems**

### Phase 5: Performance Optimization & Final Integration + Tests (Low Priority)

#### Phase 5A: Query Optimization & Performance
- [ ] Implement efficient XPath queries for status operations
- [ ] Add caching strategy for repeated epic parsing
- [ ] Optimize XML traversal for large epic files
- [ ] Create query performance benchmarks and monitoring
- [ ] Implement memory management for parsed structures
- [ ] Add performance profiling and optimization

#### Phase 5B: Write Performance Tests **IMMEDIATELY AFTER 5A**
- [ ] **Test: Query performance under 50ms target**
- [ ] **Test: Memory usage optimization for large epics**
- [ ] **Test: XPath query efficiency and accuracy**
- [ ] **Test: Caching strategy effectiveness**
- [ ] **Test: Performance regression detection**

#### Phase 5C: Error Handling & Edge Cases
- [ ] Add comprehensive error handling for all commands
- [ ] Implement graceful handling of malformed XML
- [ ] Create appropriate responses for empty epics
- [ ] Add safe calculation handling (division by zero)
- [ ] Implement missing file and permission error handling
- [ ] Create consistent error messaging across commands

#### Phase 5D: Write Error Handling Tests **IMMEDIATELY AFTER 5C**
- [ ] **Test: Missing epic file error handling**
- [ ] **Test: Invalid epic XML graceful handling**
- [ ] **Test: Empty epic appropriate responses**
- [ ] **Test: Calculation edge cases and safety**
- [ ] **Test: Consistent error format across commands**
- [ ] **Test: Error recovery and user guidance**

#### Phase 5E: Integration Testing & Command Consistency
- [ ] Create end-to-end command workflow testing
- [ ] Implement XML output format consistency validation
- [ ] Add command chaining and interaction testing
- [ ] Create comprehensive acceptance criteria verification
- [ ] Implement integration with Epic 1 systems testing
- [ ] Add final quality assurance and validation

#### Phase 5F: Write Integration Tests **IMMEDIATELY AFTER 5E**
- [ ] **Test: End-to-end command workflows**
- [ ] **Test: XML output format consistency**
- [ ] **Test: Command interaction and chaining**
- [ ] **Test: Epic 1 integration completeness**
- [ ] **Test: All acceptance criteria verification**
- [ ] **Test: Production readiness validation**

## Development Workflow Per Phase

For **EACH** phase:

1. **Implement Code** (Phase XA, XC, or XE)
2. **Write Tests IMMEDIATELY** (Phase XB, XD, or XF) 
3. **Run Tests & Verify** - All tests must pass
4. **Run Linting/Type Checking** - Code must be clean
5. **NEVER move to next phase with failing tests**

## Epic 2 Specific Considerations

### Dependencies & Integration
- **Epic 1:** CLI framework, configuration management, epic loading, validation
- **Storage:** Leverage Epic 1 storage abstraction for file operations
- **XML Processing:** Extend Epic 1 etree usage with XPath optimization
- **Error Handling:** Build on Epic 1 error handling patterns

### Technical Requirements
- **Query Performance:** XPath queries optimized for speed
- **Memory Efficiency:** Minimal memory footprint for large epic files
- **Caching Strategy:** Smart caching for repeated operations
- **XML Consistency:** Structured output format across all commands
- **File Override:** `-f` flag support for multi-epic workflows

### Algorithm Implementations
- **Progress Calculation:** Weighted task completion as primary metric
- **Next Action Logic:** Priority-based recommendation engine
- **Work Discovery:** Dependency-aware pending work enumeration
- **Event Timeline:** Chronological ordering with filtering capabilities

### Performance Targets
- **Status Queries:** < 50ms for typical epic files
- **Command Execution:** < 100ms end-to-end for all commands
- **Memory Usage:** < 10MB for large epic files (1000+ tasks)
- **XPath Queries:** < 10ms for complex element selection

## Benefits of This Approach

✅ **Immediate Feedback** - Catch issues as soon as code is written  
✅ **Working Code** - Each phase delivers tested, working functionality  
✅ **Epic 2 Coverage** - All Epic 2 test scenarios distributed across phases  
✅ **Incremental Progress** - Query commands work after each phase  
✅ **Risk Mitigation** - Problems caught early, not at the end  
✅ **Quality Assurance** - No untested code makes it to later phases  

## Test Distribution Summary

- **Phase 1 Tests:** 7 scenarios (Status calculation, progress metrics)
- **Phase 2 Tests:** 6 scenarios (Current state, next actions, active work)
- **Phase 3 Tests:** 5 scenarios (Pending work discovery, categorization)
- **Phase 4 Tests:** 6 scenarios (Failing tests, events, timeline)
- **Phase 5 Tests:** 11 scenarios (Performance, error handling, integration)

**Total: All Epic 2 test scenarios covered across all phases**

---

## Implementation Status

### EPIC 2: QUERY & STATUS COMMANDS - STATUS: READY FOR IMPLEMENTATION

### Progress Tracking
- [ ] Phase 1A: Progress Calculation Engine
- [ ] Phase 1B: Write Progress Calculation Tests
- [ ] Phase 1C: Implement `agentpm status` Command
- [ ] Phase 1D: Write `agentpm status` Tests
- [ ] Phase 2A: Active Work Detection & Next Action Engine
- [ ] Phase 2B: Write Current State Tests
- [ ] Phase 2C: Implement `agentpm current` Command
- [ ] Phase 2D: Write `agentpm current` Tests
- [ ] Phase 3A: Pending Work Enumeration Engine
- [ ] Phase 3B: Write Pending Work Tests
- [ ] Phase 3C: Implement `agentpm pending` Command
- [ ] Phase 3D: Write `agentpm pending` Tests
- [ ] Phase 4A: Failing Test Detection & Event Query System
- [ ] Phase 4B: Write Failing Tests & Events Tests
- [ ] Phase 4C: Implement `agentpm failing` and `agentpm events` Commands
- [ ] Phase 4D: Write Command Integration Tests
- [ ] Phase 5A: Query Optimization & Performance
- [ ] Phase 5B: Write Performance Tests
- [ ] Phase 5C: Error Handling & Edge Cases
- [ ] Phase 5D: Write Error Handling Tests
- [ ] Phase 5E: Integration Testing & Command Consistency
- [ ] Phase 5F: Write Integration Tests

---

## EPIC 2 IMPLEMENTATION READY

**📋 STATUS: IMPLEMENTATION PLAN COMPLETE**

**Implementation Guidelines:**
- **3-4 day duration** with proper test-driven development
- **22 implementation phases** with immediate testing after each
- **Query intelligence** for smart agent recommendations
- **Performance optimization** for efficient epic analysis

**Quality Gates:**
- ✅ All query commands complete within 50ms performance targets
- ✅ Accurate progress calculations for all epic states
- ✅ Intelligent next action recommendations based on current state
- ✅ Comprehensive error handling for edge cases

**Next Steps:**
- Begin implementation with Phase 1A: Progress Calculation Engine
- Follow TDD approach: implement code, then write tests immediately
- Focus on query performance and intelligent recommendations
- Build foundation for Epic 3: Epic Lifecycle Management

**🚀 Epic 2: Query & Status Commands - READY FOR DEVELOPMENT! 🚀**