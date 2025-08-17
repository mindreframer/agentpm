# Epic 14 Framework - Performance Tuning Guidelines

## Overview

This guide provides comprehensive performance optimization strategies for the Epic 14 Transition Chain Testing Framework. It covers memory management, execution optimization, and scaling considerations for large test suites.

## Table of Contents

1. [Performance Fundamentals](#performance-fundamentals)
2. [Memory Management](#memory-management)
3. [Execution Optimization](#execution-optimization)
4. [Debug Mode Optimization](#debug-mode-optimization)
5. [Concurrent Testing](#concurrent-testing)
6. [Large-Scale Testing](#large-scale-testing)
7. [Monitoring and Profiling](#monitoring-and-profiling)
8. [CI/CD Optimization](#cicd-optimization)

## Performance Fundamentals

### Understanding Framework Overhead

The Epic 14 framework introduces several layers of functionality that can impact performance:

1. **State Tracking**: Every state change is captured and stored
2. **Event Logging**: All events are recorded with timestamps and context
3. **Debug Tracing**: Optional detailed execution traces
4. **State Visualization**: Visual representation of state transitions
5. **Snapshot Comparison**: State comparison for regression testing

### Performance Baseline

Typical performance characteristics:
- **Basic assertion**: ~0.1ms per assertion
- **Complex custom assertion**: ~1-5ms depending on logic
- **Snapshot testing**: ~2-10ms depending on data size
- **State visualization**: ~5-15ms for complex state graphs
- **Memory usage**: ~1-5MB per test execution (varies by complexity)

## Memory Management

### 1. Debug Mode Impact

**High Impact: Debug modes consume significant memory**

```go
// Memory usage comparison
func BenchmarkDebugModeMemory(b *testing.B) {
    b.Run("debug_off", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            result := executeWorkflow()
            assertions.Assert(result).
                WithDebugMode(assertions.DebugOff).  // Minimal memory
                EpicStatus("completed").
                MustPass()
        }
    })
    
    b.Run("debug_verbose", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            result := executeWorkflow()
            assertions.Assert(result).
                WithDebugMode(assertions.DebugVerbose).  // High memory usage
                EpicStatus("completed").
                MustPass()
        }
    })
}
```

**Best Practice: Use appropriate debug levels**
```go
func TestPerformanceOptimized(t *testing.T) {
    result := executeWorkflow()
    
    // Use DebugOff for performance-critical tests
    debugMode := assertions.DebugOff
    
    // Enable debugging only when needed
    if testing.Verbose() || os.Getenv("DEBUG_TESTS") == "true" {
        debugMode = assertions.DebugBasic
    }
    
    assertions.Assert(result).
        WithDebugMode(debugMode).
        EpicStatus("completed").
        MustPass()
}
```

### 2. State Visualization Memory Usage

**Problem: State visualization can consume significant memory**
```go
// Memory-intensive approach
func TestMemoryIntensive(t *testing.T) {
    for i := 0; i < 1000; i++ {
        result := executeWorkflow()
        
        assertions.Assert(result).
            EnableStateVisualization().  // Creates visualization data
            EpicStatus("completed").
            GetStateVisualization()      // Keeps data in memory
        
        // Visualization data not cleaned up
    }
}
```

**Solution: Selective visualization**
```go
func TestMemoryOptimized(t *testing.T) {
    for i := 0; i < 1000; i++ {
        result := executeWorkflow()
        
        builder := assertions.Assert(result).EpicStatus("completed")
        
        // Only enable visualization for failing tests or debug mode
        if t.Failed() || testing.Verbose() {
            builder = builder.EnableStateVisualization()
        }
        
        builder.MustPass()
        
        // Explicit cleanup for large loops
        result = nil
        runtime.GC()  // Periodic garbage collection
        
        if i%100 == 99 {
            runtime.GC()
        }
    }
}
```

### 3. Snapshot Data Management

**Problem: Large snapshot datasets**
```go
// Inefficient snapshot usage
func TestSnapshotMemoryIssue(t *testing.T) {
    for i := 0; i < 1000; i++ {
        result := executeWorkflow()
        
        // Creates a new snapshot for each iteration
        assertions.Assert(result).
            MatchSnapshot(fmt.Sprintf("test_%d", i)).  // 1000 different snapshots
            MustPass()
    }
}
```

**Solution: Efficient snapshot usage**
```go
func TestSnapshotMemoryOptimized(t *testing.T) {
    // Use a single snapshot for similar test data
    baselineResult := executeBaselineWorkflow()
    
    for i := 0; i < 1000; i++ {
        result := executeWorkflow()
        
        // Compare against baseline instead of creating new snapshots
        if !isEquivalentResult(result, baselineResult) {
            t.Errorf("Result %d differs from baseline", i)
        }
        
        // Only create snapshots for specific cases
        if i%100 == 0 {
            assertions.Assert(result).
                MatchSnapshot(fmt.Sprintf("checkpoint_%d", i/100)).
                MustPass()
        }
    }
}
```

### 4. Batch Processing Memory Management

```go
func TestBatchProcessingOptimized(t *testing.T) {
    const batchSize = 100
    const totalTests = 10000
    
    var baselineMemory runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&baselineMemory)
    
    for batch := 0; batch < totalTests/batchSize; batch++ {
        // Process in batches to manage memory
        for i := 0; i < batchSize; i++ {
            result := executeWorkflow()
            
            assertions.Assert(result).
                WithDebugMode(assertions.DebugOff).  // Minimal memory
                EpicStatus("completed").
                MustPass()
            
            result = nil  // Explicit cleanup
        }
        
        // Garbage collect after each batch
        runtime.GC()
        
        // Monitor memory usage
        var currentMemory runtime.MemStats
        runtime.ReadMemStats(&currentMemory)
        
        memoryGrowth := int64(currentMemory.Alloc) - int64(baselineMemory.Alloc)
        maxGrowth := int64(50 * 1024 * 1024) // 50MB threshold
        
        if memoryGrowth > maxGrowth {
            t.Fatalf("Memory growth too large after batch %d: %d bytes", 
                batch, memoryGrowth)
        }
    }
}
```

## Execution Optimization

### 1. Batch Assertions

**Inefficient: Individual assertions**
```go
func TestIndividualAssertions(t *testing.T) {
    result := executeWorkflow()
    
    // Each assertion creates overhead
    assertions.Assert(result).EpicStatus("completed").MustPass()
    assertions.Assert(result).PhaseStatus("1A", "completed").MustPass()
    assertions.Assert(result).PhaseStatus("1B", "completed").MustPass()
    assertions.Assert(result).NoErrors().MustPass()
}
```

**Efficient: Chained assertions**
```go
func TestChainedAssertions(t *testing.T) {
    result := executeWorkflow()
    
    // Single assertion chain reduces overhead
    assertions.Assert(result).
        EpicStatus("completed").
        PhaseStatus("1A", "completed").
        PhaseStatus("1B", "completed").
        NoErrors().
        MustPass()
}
```

**Most Efficient: Batch assertions**
```go
func TestBatchAssertions(t *testing.T) {
    result := executeWorkflow()
    
    // Pre-define assertion sets for reuse
    statusAssertions := []func(*assertions.AssertionBuilder) *assertions.AssertionBuilder{
        func(ab *assertions.AssertionBuilder) *assertions.AssertionBuilder {
            return ab.EpicStatus("completed")
        },
        func(ab *assertions.AssertionBuilder) *assertions.AssertionBuilder {
            return ab.PhaseStatus("1A", "completed")
        },
        func(ab *assertions.AssertionBuilder) *assertions.AssertionBuilder {
            return ab.PhaseStatus("1B", "completed")
        },
    }
    
    assertions.Assert(result).
        BatchAssertions(statusAssertions).
        NoErrors().
        MustPass()
}
```

### 2. Custom Assertion Optimization

**Inefficient: Multiple passes over data**
```go
func TestInefficient(t *testing.T) {
    result := executeWorkflow()
    
    assertions.Assert(result).
        CustomAssertion("check_phases", func(r *executor.TransitionChainResult) error {
            // First pass: check phase count
            if len(r.FinalState.Phases) != 3 {
                return fmt.Errorf("wrong phase count")
            }
            return nil
        }).
        CustomAssertion("check_tasks", func(r *executor.TransitionChainResult) error {
            // Second pass: check task count  
            if len(r.FinalState.Tasks) != 6 {
                return fmt.Errorf("wrong task count")
            }
            return nil
        }).
        CustomAssertion("check_events", func(r *executor.TransitionChainResult) error {
            // Third pass: check event count
            if len(r.FinalState.Events) < 10 {
                return fmt.Errorf("insufficient events")
            }
            return nil
        }).
        MustPass()
}
```

**Efficient: Single pass validation**
```go
func TestEfficient(t *testing.T) {
    result := executeWorkflow()
    
    assertions.Assert(result).
        CustomAssertion("comprehensive_check", func(r *executor.TransitionChainResult) error {
            // Single pass validation
            epic := r.FinalState
            
            if len(epic.Phases) != 3 {
                return fmt.Errorf("wrong phase count: expected 3, got %d", len(epic.Phases))
            }
            
            if len(epic.Tasks) != 6 {
                return fmt.Errorf("wrong task count: expected 6, got %d", len(epic.Tasks))
            }
            
            if len(epic.Events) < 10 {
                return fmt.Errorf("insufficient events: expected 10+, got %d", len(epic.Events))
            }
            
            return nil
        }).
        MustPass()
}
```

### 3. Timing Optimization

**Problem: Inefficient timing checks**
```go
func TestTimingInefficient(t *testing.T) {
    // Timing each phase individually
    start1 := time.Now()
    executePhase1()
    phase1Duration := time.Since(start1)
    
    start2 := time.Now()
    executePhase2()
    phase2Duration := time.Since(start2)
    
    // Multiple timing validations
    if phase1Duration > time.Second {
        t.Error("Phase 1 too slow")
    }
    if phase2Duration > time.Second {
        t.Error("Phase 2 too slow")
    }
}
```

**Solution: Built-in performance benchmarking**
```go
func TestTimingOptimized(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("timing-test").
        ExecutePhase("1A").
        ExecutePhase("1B").
        CompleteEpic().
        Execute()
    
    // Use built-in performance validation
    assertions.Assert(result).
        ExecutionTime(5 * time.Second).                    // Total time
        PhaseTransitionTiming("1A", 2 * time.Second).      // Phase-specific timing
        PerformanceBenchmark(5*time.Second, 100).          // Time + memory
        MustPass()
}
```

## Debug Mode Optimization

### Debug Mode Levels and Performance Impact

| Debug Mode | Memory Usage | CPU Usage | Use Case |
|------------|--------------|-----------|----------|
| `DebugOff` | Minimal (1x) | Minimal (1x) | Production tests, CI/CD |
| `DebugBasic` | Low (2-3x) | Low (1.2x) | Basic failure analysis |
| `DebugVerbose` | Medium (5-8x) | Medium (1.5x) | Development debugging |
| `DebugTrace` | High (10-15x) | High (2-3x) | Complex issue investigation |

### Conditional Debug Modes

```go
func TestConditionalDebug(t *testing.T) {
    result := executeWorkflow()
    
    // Choose debug mode based on context
    debugMode := getOptimalDebugMode(t)
    
    assertions.Assert(result).
        WithDebugMode(debugMode).
        EpicStatus("completed").
        MustPass()
}

func getOptimalDebugMode(t *testing.T) assertions.DebugMode {
    // CI environment - minimal debugging
    if os.Getenv("CI") == "true" {
        return assertions.DebugOff
    }
    
    // Test is failing - enable debugging
    if t.Failed() {
        return assertions.DebugVerbose
    }
    
    // Verbose testing requested
    if testing.Verbose() {
        return assertions.DebugBasic
    }
    
    // Environment variable override
    switch os.Getenv("DEBUG_LEVEL") {
    case "trace":
        return assertions.DebugTrace
    case "verbose":
        return assertions.DebugVerbose
    case "basic":
        return assertions.DebugBasic
    default:
        return assertions.DebugOff
    }
}
```

### Smart Debug Output

```go
func TestSmartDebugOutput(t *testing.T) {
    result := executeWorkflow()
    
    builder := assertions.Assert(result).
        WithDebugMode(assertions.DebugBasic).
        EpicStatus("completed")
    
    // Only print debug info on failure
    err := builder.Check()
    if err != nil {
        // Enable verbose debugging for failure analysis
        builder.WithDebugMode(assertions.DebugVerbose).
            PrintDebugInfo()
        
        t.Fatal(err)
    }
}
```

## Concurrent Testing

### 1. Thread-Safe Testing Patterns

```go
func TestConcurrentExecution(t *testing.T) {
    const numGoroutines = 10
    
    var wg sync.WaitGroup
    results := make(chan *executor.TransitionChainResult, numGoroutines)
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            // Each goroutine gets its own builder instance
            result := executor.NewTransitionChain().
                StartEpic(fmt.Sprintf("concurrent-epic-%d", id)).
                ExecutePhase("1A").
                CompleteEpic().
                Execute()
            
            results <- result
        }(i)
    }
    
    go func() {
        wg.Wait()
        close(results)
    }()
    
    // Validate results from all goroutines
    for result := range results {
        assertions.Assert(result).
            WithDebugMode(assertions.DebugOff).  // Minimal overhead
            EpicStatus("completed").
            MustPass()
    }
}
```

### 2. Resource Pool Management

```go
type TestResourcePool struct {
    builders chan *assertions.AssertionBuilder
    mu       sync.Mutex
    created  int
}

func NewTestResourcePool(size int) *TestResourcePool {
    pool := &TestResourcePool{
        builders: make(chan *assertions.AssertionBuilder, size),
    }
    
    // Pre-create builders
    for i := 0; i < size; i++ {
        pool.builders <- assertions.NewAssertionBuilder(nil)
    }
    
    return pool
}

func (p *TestResourcePool) GetBuilder() *assertions.AssertionBuilder {
    select {
    case builder := <-p.builders:
        return builder
    default:
        // Create new builder if pool is empty
        p.mu.Lock()
        p.created++
        p.mu.Unlock()
        return assertions.NewAssertionBuilder(nil)
    }
}

func (p *TestResourcePool) ReturnBuilder(builder *assertions.AssertionBuilder) {
    // Reset builder state
    builder = assertions.NewAssertionBuilder(nil)
    
    select {
    case p.builders <- builder:
        // Successfully returned to pool
    default:
        // Pool is full, let GC handle it
    }
}
```

### 3. Parallel Test Execution

```go
func TestParallelOptimization(t *testing.T) {
    testCases := generateTestCases(100)
    
    // Group tests by resource requirements
    fastTests := filterFastTests(testCases)
    slowTests := filterSlowTests(testCases)
    
    t.Run("fast_tests", func(t *testing.T) {
        for _, tc := range fastTests {
            tc := tc
            t.Run(tc.name, func(t *testing.T) {
                t.Parallel()
                
                result := tc.execute()
                assertions.Assert(result).
                    WithDebugMode(assertions.DebugOff).
                    EpicStatus("completed").
                    MustPass()
            })
        }
    })
    
    t.Run("slow_tests", func(t *testing.T) {
        // Run slow tests sequentially to manage resources
        for _, tc := range slowTests {
            t.Run(tc.name, func(t *testing.T) {
                result := tc.execute()
                assertions.Assert(result).
                    WithDebugMode(assertions.DebugBasic).
                    EpicStatus("completed").
                    MustPass()
            })
        }
    })
}
```

## Large-Scale Testing

### 1. Test Data Management

```go
func TestLargeScaleOptimized(t *testing.T) {
    const totalTests = 10000
    const batchSize = 100
    
    // Pre-generate test data
    testData := generateTestDataBatch(batchSize)
    
    for batch := 0; batch < totalTests/batchSize; batch++ {
        t.Run(fmt.Sprintf("batch_%d", batch), func(t *testing.T) {
            // Reuse test data for entire batch
            for i, data := range testData {
                result := executeWorkflowWithData(data)
                
                assertions.Assert(result).
                    WithDebugMode(assertions.DebugOff).
                    EpicStatus("completed").
                    MustPass()
                
                // Cleanup every 10 tests
                if i%10 == 9 {
                    runtime.GC()
                }
            }
        })
        
        // Major cleanup after each batch
        runtime.GC()
    }
}
```

### 2. Progressive Complexity Testing

```go
func TestProgressiveComplexity(t *testing.T) {
    complexityLevels := []struct {
        name       string
        iterations int
        phases     int
        tasks      int
    }{
        {"simple", 1000, 2, 4},
        {"medium", 500, 4, 8}, 
        {"complex", 100, 8, 16},
        {"extreme", 10, 16, 32},
    }
    
    for _, level := range complexityLevels {
        t.Run(level.name, func(t *testing.T) {
            for i := 0; i < level.iterations; i++ {
                result := executeComplexWorkflow(level.phases, level.tasks)
                
                // Adjust debug mode based on complexity
                debugMode := assertions.DebugOff
                if level.phases > 8 {
                    debugMode = assertions.DebugBasic
                }
                
                assertions.Assert(result).
                    WithDebugMode(debugMode).
                    EpicStatus("completed").
                    MustPass()
            }
        })
    }
}
```

### 3. Memory Monitoring

```go
func TestWithMemoryMonitoring(t *testing.T) {
    monitor := NewMemoryMonitor()
    defer monitor.Report(t)
    
    const iterations = 1000
    
    for i := 0; i < iterations; i++ {
        monitor.Checkpoint(fmt.Sprintf("iteration_%d", i))
        
        result := executeWorkflow()
        
        assertions.Assert(result).
            WithDebugMode(assertions.DebugOff).
            EpicStatus("completed").
            MustPass()
        
        if monitor.MemoryGrowthExceeded(100 * 1024 * 1024) { // 100MB
            t.Fatalf("Memory growth exceeded threshold at iteration %d", i)
        }
    }
}

type MemoryMonitor struct {
    checkpoints []MemoryCheckpoint
    baseline    runtime.MemStats
}

type MemoryCheckpoint struct {
    Name   string
    Memory runtime.MemStats
    Time   time.Time
}

func NewMemoryMonitor() *MemoryMonitor {
    m := &MemoryMonitor{}
    runtime.GC()
    runtime.ReadMemStats(&m.baseline)
    return m
}

func (m *MemoryMonitor) Checkpoint(name string) {
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    
    m.checkpoints = append(m.checkpoints, MemoryCheckpoint{
        Name:   name,
        Memory: memStats,
        Time:   time.Now(),
    })
}

func (m *MemoryMonitor) MemoryGrowthExceeded(threshold uint64) bool {
    var current runtime.MemStats
    runtime.ReadMemStats(&current)
    
    growth := current.Alloc - m.baseline.Alloc
    return growth > threshold
}

func (m *MemoryMonitor) Report(t *testing.T) {
    if len(m.checkpoints) == 0 {
        return
    }
    
    var final runtime.MemStats
    runtime.ReadMemStats(&final)
    
    totalGrowth := final.Alloc - m.baseline.Alloc
    t.Logf("Memory usage report:")
    t.Logf("  Baseline: %d bytes", m.baseline.Alloc)
    t.Logf("  Final: %d bytes", final.Alloc)
    t.Logf("  Growth: %d bytes", totalGrowth)
    t.Logf("  Checkpoints: %d", len(m.checkpoints))
}
```

## Monitoring and Profiling

### 1. Built-in Performance Monitoring

```go
func TestWithBuiltInProfiling(t *testing.T) {
    result := executor.NewTransitionChain().
        StartEpic("profiling-test").
        ExecutePhase("1A").
        CompleteEpic().
        Execute()
    
    assertions.Assert(result).
        PerformanceBenchmark(5*time.Second, 100).  // Built-in monitoring
        EpicStatus("completed").
        MustPass()
}
```

### 2. Custom Performance Metrics

```go
func TestCustomMetrics(t *testing.T) {
    metrics := NewPerformanceMetrics()
    
    metrics.Start("total_execution")
    result := executeWorkflow()
    metrics.Stop("total_execution")
    
    metrics.Start("assertion_validation")
    assertions.Assert(result).
        EpicStatus("completed").
        MustPass()
    metrics.Stop("assertion_validation")
    
    // Report metrics
    metrics.Report(t)
    
    // Validate performance requirements
    if metrics.Duration("total_execution") > 5*time.Second {
        t.Error("Execution too slow")
    }
}
```

### 3. Integration with Go's pprof

```go
//go:build performance
// +build performance

func TestWithProfiling(t *testing.T) {
    // CPU profiling
    cpuFile, err := os.Create("cpu.prof")
    if err != nil {
        t.Fatal(err)
    }
    defer cpuFile.Close()
    
    pprof.StartCPUProfile(cpuFile)
    defer pprof.StopCPUProfile()
    
    // Memory profiling
    defer func() {
        memFile, err := os.Create("mem.prof")
        if err != nil {
            t.Error(err)
            return
        }
        defer memFile.Close()
        
        runtime.GC()
        pprof.WriteHeapProfile(memFile)
    }()
    
    // Run performance test
    const iterations = 1000
    for i := 0; i < iterations; i++ {
        result := executeWorkflow()
        
        assertions.Assert(result).
            WithDebugMode(assertions.DebugOff).
            EpicStatus("completed").
            MustPass()
    }
}
```

## CI/CD Optimization

### 1. Environment-Specific Optimization

```go
func TestCIOptimized(t *testing.T) {
    result := executeWorkflow()
    
    builder := assertions.Assert(result)
    
    // Optimize for CI environment
    if isCI() {
        builder = builder.
            WithDebugMode(assertions.DebugOff).          // Minimal debug
            WithTimeout(30 * time.Second)                // Reasonable timeout
    } else {
        builder = builder.
            WithDebugMode(assertions.DebugBasic).        // More debugging locally
            EnableStateVisualization()                   // Helpful for development
    }
    
    builder.EpicStatus("completed").MustPass()
}

func isCI() bool {
    return os.Getenv("CI") != "" || 
           os.Getenv("GITHUB_ACTIONS") != "" ||
           os.Getenv("JENKINS_URL") != ""
}
```

### 2. Parallel Test Configuration

```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        test-group: [unit, integration, performance]
        
    steps:
    - uses: actions/checkout@v2
    - uses: actions/setup-go@v2
      with:
        go-version: 1.21
        
    - name: Run tests
      run: |
        case "${{ matrix.test-group }}" in
          unit)
            go test -v -short ./...
            ;;
          integration)
            go test -v -tags=integration ./...
            ;;
          performance)
            go test -v -tags=performance -timeout=30m ./...
            ;;
        esac
      env:
        CI: true
        DEBUG_LEVEL: off
```

### 3. Test Result Caching

```go
func TestWithCaching(t *testing.T) {
    cacheKey := generateCacheKey("workflow_test", getCodeVersion())
    
    // Check cache first
    if cachedResult := loadFromCache(cacheKey); cachedResult != nil {
        assertions.Assert(cachedResult).
            EpicStatus("completed").
            MustPass()
        return
    }
    
    // Execute test
    result := executeWorkflow()
    
    // Cache result for future runs
    saveToCache(cacheKey, result)
    
    assertions.Assert(result).
        EpicStatus("completed").
        MustPass()
}
```

## Performance Best Practices Summary

### ✅ DO

1. **Use appropriate debug modes** based on context (CI vs development)
2. **Chain assertions** instead of creating multiple builders
3. **Use batch assertions** for repeated patterns
4. **Monitor memory usage** in large test suites
5. **Clean up resources** explicitly in loops
6. **Use parallel testing** for independent tests
7. **Profile performance-critical** test suites
8. **Configure timeouts** appropriately for environment

### ❌ DON'T

1. **Enable verbose debugging** in CI environments
2. **Create unnecessary snapshots** for similar data
3. **Use state visualization** in large loops
4. **Ignore memory growth** in long-running tests
5. **Mix fast and slow tests** in parallel groups
6. **Over-assert on implementation details**
7. **Skip performance validation** for critical paths
8. **Use debugging features** in production test runs

### Performance Monitoring Checklist

- [ ] **Memory usage** stays within reasonable bounds (< 100MB growth per 1000 tests)
- [ ] **Execution time** meets targets (< 1ms per basic assertion)
- [ ] **Debug modes** are appropriate for environment
- [ ] **Resource cleanup** prevents memory leaks
- [ ] **Parallel execution** is safe and effective
- [ ] **CI optimization** reduces build times
- [ ] **Profiling data** is collected for optimization

By following these guidelines, your Epic 14 test suites will scale efficiently from small unit tests to large integration test suites while maintaining fast feedback cycles.