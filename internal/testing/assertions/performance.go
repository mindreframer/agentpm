package assertions

import (
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/mindreframer/agentpm/internal/testing/executor"
)

// === PHASE 5A: PERFORMANCE OPTIMIZATION ===

// PerformanceConfig contains configuration for performance optimization
type PerformanceConfig struct {
	// Builder pattern optimizations
	PreallocateErrors      int  // Pre-allocate error slice capacity
	PreallocateTrace       int  // Pre-allocate trace slice capacity
	ReuseBuilders          bool // Enable builder instance reuse
	LazyStateVisualization bool // Only create visualization when requested

	// Memory optimizations
	EnableMemoryPooling   bool // Use memory pools for frequent allocations
	CompactRepresentation bool // Use compact data structures
	AggressiveGC          bool // Enable aggressive garbage collection

	// Execution optimizations
	ConcurrentAssertions  bool // Enable concurrent assertion execution
	BatchAssertions       bool // Batch multiple assertions for efficiency
	SkipIntermediateTrace bool // Skip intermediate debug traces for performance

	// Resource management
	MaxMemoryUsage     int64         // Maximum memory usage limit (bytes)
	CleanupInterval    time.Duration // Interval for resource cleanup
	MaxConcurrentTests int           // Maximum concurrent test executions
}

// DefaultPerformanceConfig provides optimized default settings
func DefaultPerformanceConfig() *PerformanceConfig {
	return &PerformanceConfig{
		PreallocateErrors:      10,
		PreallocateTrace:       50,
		ReuseBuilders:          true,
		LazyStateVisualization: true,
		EnableMemoryPooling:    true,
		CompactRepresentation:  true,
		AggressiveGC:           false,
		ConcurrentAssertions:   true,
		BatchAssertions:        true,
		SkipIntermediateTrace:  false,
		MaxMemoryUsage:         50 * 1024 * 1024, // 50MB
		CleanupInterval:        time.Second * 30,
		MaxConcurrentTests:     runtime.NumCPU() * 2,
	}
}

// HighPerformanceConfig provides settings optimized for speed over debugging
func HighPerformanceConfig() *PerformanceConfig {
	config := DefaultPerformanceConfig()
	config.SkipIntermediateTrace = true
	config.LazyStateVisualization = true
	config.AggressiveGC = true
	config.ConcurrentAssertions = true
	config.BatchAssertions = true
	return config
}

// MemoryOptimizedConfig provides settings optimized for low memory usage
func MemoryOptimizedConfig() *PerformanceConfig {
	config := DefaultPerformanceConfig()
	config.PreallocateErrors = 5
	config.PreallocateTrace = 10
	config.CompactRepresentation = true
	config.EnableMemoryPooling = true
	config.MaxMemoryUsage = 10 * 1024 * 1024 // 10MB
	config.CleanupInterval = time.Second * 10
	return config
}

// PerformanceOptimizedAssertionBuilder extends AssertionBuilder with performance optimizations
type PerformanceOptimizedAssertionBuilder struct {
	*AssertionBuilder
	config       *PerformanceConfig
	memoryPool   *sync.Pool
	startTime    time.Time
	memoryUsage  int64
	mu           sync.RWMutex
	batchedOps   []func()
	isProcessing bool
}

// StringPool provides optimized string operations with pooling
type StringPool struct {
	pool sync.Pool
}

func NewStringPool() *StringPool {
	return &StringPool{
		pool: sync.Pool{
			New: func() interface{} {
				s := make([]string, 0, 10)
				return &s
			},
		},
	}
}

func (sp *StringPool) GetSlice() *[]string {
	return sp.pool.Get().(*[]string)
}

func (sp *StringPool) PutSlice(s *[]string) {
	*s = (*s)[:0] // Reset slice but keep capacity
	sp.pool.Put(s)
}

// Global pools for performance optimization
var (
	errorPool = sync.Pool{
		New: func() interface{} {
			return make([]AssertionError, 0, 10)
		},
	}

	tracePool = sync.Pool{
		New: func() interface{} {
			return make([]TraceEntry, 0, 50)
		},
	}

	stringPool = NewStringPool()
)

// NewOptimizedAssertionBuilder creates a performance-optimized assertion builder
func NewOptimizedAssertionBuilder(result *executor.TransitionChainResult, config *PerformanceConfig) *PerformanceOptimizedAssertionBuilder {
	if config == nil {
		config = DefaultPerformanceConfig()
	}

	// Create base builder with pre-allocated slices
	base := &AssertionBuilder{
		result:   result,
		errors:   make([]AssertionError, 0, config.PreallocateErrors),
		debugCtx: NewDebugContext(DebugOff), // Start with debug off for performance
		recovery: DefaultRecoveryStrategy(),
	}

	optimized := &PerformanceOptimizedAssertionBuilder{
		AssertionBuilder: base,
		config:           config,
		startTime:        time.Now(),
		batchedOps:       make([]func(), 0, 20),
	}

	// Initialize memory pool if enabled
	if config.EnableMemoryPooling {
		optimized.memoryPool = &sync.Pool{
			New: func() interface{} {
				return make([]byte, 1024) // 1KB buffer pool
			},
		}
	}

	return optimized
}

// GetMemoryUsage returns current memory usage estimation
func (pab *PerformanceOptimizedAssertionBuilder) GetMemoryUsage() int64 {
	pab.mu.RLock()
	defer pab.mu.RUnlock()

	// Estimate memory usage
	usage := int64(unsafe.Sizeof(*pab))
	usage += int64(len(pab.errors)) * int64(unsafe.Sizeof(AssertionError{}))

	if pab.debugCtx != nil {
		trace := pab.debugCtx.GetTraceLog()
		usage += int64(len(trace)) * int64(unsafe.Sizeof(TraceEntry{}))
	}

	return usage
}

// OptimizedEpicStatus provides a performance-optimized epic status assertion
func (pab *PerformanceOptimizedAssertionBuilder) OptimizedEpicStatus(expectedStatus string) *PerformanceOptimizedAssertionBuilder {
	if pab.config.BatchAssertions {
		pab.addBatchedOperation(func() {
			pab.AssertionBuilder.EpicStatus(expectedStatus)
		})
		return pab
	}

	pab.AssertionBuilder.EpicStatus(expectedStatus)
	return pab
}

// OptimizedPhaseStatus provides a performance-optimized phase status assertion
func (pab *PerformanceOptimizedAssertionBuilder) OptimizedPhaseStatus(phaseID, expectedStatus string) *PerformanceOptimizedAssertionBuilder {
	if pab.config.BatchAssertions {
		pab.addBatchedOperation(func() {
			pab.AssertionBuilder.PhaseStatus(phaseID, expectedStatus)
		})
		return pab
	}

	pab.AssertionBuilder.PhaseStatus(phaseID, expectedStatus)
	return pab
}

// OptimizedTaskStatus provides a performance-optimized task status assertion
func (pab *PerformanceOptimizedAssertionBuilder) OptimizedTaskStatus(taskID, expectedStatus string) *PerformanceOptimizedAssertionBuilder {
	if pab.config.BatchAssertions {
		pab.addBatchedOperation(func() {
			pab.AssertionBuilder.TaskStatus(taskID, expectedStatus)
		})
		return pab
	}

	pab.AssertionBuilder.TaskStatus(taskID, expectedStatus)
	return pab
}

// addBatchedOperation adds an operation to the batch queue
func (pab *PerformanceOptimizedAssertionBuilder) addBatchedOperation(op func()) {
	pab.mu.Lock()
	defer pab.mu.Unlock()

	pab.batchedOps = append(pab.batchedOps, op)
}

// ExecuteBatch processes all batched operations
func (pab *PerformanceOptimizedAssertionBuilder) ExecuteBatch() *PerformanceOptimizedAssertionBuilder {
	pab.mu.Lock()
	defer pab.mu.Unlock()

	if pab.isProcessing {
		return pab
	}

	pab.isProcessing = true
	defer func() { pab.isProcessing = false }()

	if pab.config.ConcurrentAssertions && len(pab.batchedOps) > 1 {
		pab.executeConcurrentBatch()
	} else {
		pab.executeSequentialBatch()
	}

	// Clear batch after execution
	pab.batchedOps = pab.batchedOps[:0]
	return pab
}

// executeConcurrentBatch executes operations concurrently
func (pab *PerformanceOptimizedAssertionBuilder) executeConcurrentBatch() {
	maxWorkers := pab.config.MaxConcurrentTests
	if maxWorkers == 0 {
		maxWorkers = runtime.NumCPU()
	}

	if len(pab.batchedOps) < maxWorkers {
		maxWorkers = len(pab.batchedOps)
	}

	jobs := make(chan func(), len(pab.batchedOps))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		go func() {
			for job := range jobs {
				job()
			}
		}()
	}

	// Send jobs
	for _, op := range pab.batchedOps {
		jobs <- op
	}
	close(jobs)

	wg.Wait()
}

// executeSequentialBatch executes operations sequentially
func (pab *PerformanceOptimizedAssertionBuilder) executeSequentialBatch() {
	for _, op := range pab.batchedOps {
		op()
	}
}

// EnableOptimizedStateVisualization creates state visualization only when needed
func (pab *PerformanceOptimizedAssertionBuilder) EnableOptimizedStateVisualization() *PerformanceOptimizedAssertionBuilder {
	if !pab.config.LazyStateVisualization {
		pab.AssertionBuilder.EnableStateVisualization()
	}
	return pab
}

// GetOptimizedStateVisualization returns state visualization, creating it lazily if needed
func (pab *PerformanceOptimizedAssertionBuilder) GetOptimizedStateVisualization() *StateVisualization {
	if pab.AssertionBuilder.visualizer == nil && pab.config.LazyStateVisualization {
		pab.AssertionBuilder.EnableStateVisualization()
	}
	return pab.AssertionBuilder.GetStateVisualization()
}

// OptimizedMustPass provides an optimized version of MustPass
func (pab *PerformanceOptimizedAssertionBuilder) OptimizedMustPass() {
	// Execute any batched operations first
	pab.ExecuteBatch()

	// Check if we're within memory limits
	if pab.config.MaxMemoryUsage > 0 {
		currentUsage := pab.GetMemoryUsage()
		if currentUsage > pab.config.MaxMemoryUsage {
			// Force garbage collection if we're over memory limit
			if pab.config.AggressiveGC {
				runtime.GC()
			}

			// Check again after GC
			currentUsage = pab.GetMemoryUsage()
			if currentUsage > pab.config.MaxMemoryUsage {
				panic("Memory usage exceeds limit: " + string(rune(currentUsage)) + " bytes")
			}
		}
	}

	pab.AssertionBuilder.MustPass()
}

// Cleanup performs resource cleanup and returns resources to pools
func (pab *PerformanceOptimizedAssertionBuilder) Cleanup() {
	pab.mu.Lock()
	defer pab.mu.Unlock()

	// Return errors to pool if using pooling
	if pab.config.EnableMemoryPooling && len(pab.errors) > 0 {
		errorPool.Put(pab.errors[:0])
	}

	// Return trace to pool if using pooling
	if pab.config.EnableMemoryPooling && pab.debugCtx != nil {
		trace := pab.debugCtx.GetTraceLog()
		if len(trace) > 0 {
			tracePool.Put(trace[:0])
		}
	}

	// Clear batched operations
	pab.batchedOps = pab.batchedOps[:0]

	// Force GC if configured
	if pab.config.AggressiveGC {
		runtime.GC()
	}
}

// GetPerformanceMetrics returns performance metrics for the builder
func (pab *PerformanceOptimizedAssertionBuilder) GetPerformanceMetrics() PerformanceMetrics {
	return PerformanceMetrics{
		ExecutionTime:   time.Since(pab.startTime),
		MemoryUsage:     pab.GetMemoryUsage(),
		ErrorCount:      len(pab.GetErrors()),
		BatchSize:       len(pab.batchedOps),
		ConcurrentTests: pab.config.MaxConcurrentTests,
		PoolingEnabled:  pab.config.EnableMemoryPooling,
	}
}

// PerformanceMetrics tracks performance data for optimized builders
type PerformanceMetrics struct {
	ExecutionTime   time.Duration
	MemoryUsage     int64
	ErrorCount      int
	BatchSize       int
	ConcurrentTests int
	PoolingEnabled  bool
}

// ConcurrentAssertionRunner provides thread-safe concurrent assertion execution
type ConcurrentAssertionRunner struct {
	maxWorkers  int
	jobQueue    chan AssertionJob
	resultQueue chan AssertionResult
	workerPool  sync.Pool
	activeJobs  int32
	mu          sync.RWMutex
}

// AssertionJob represents a single assertion to be executed
type AssertionJob struct {
	ID       string
	Builder  *PerformanceOptimizedAssertionBuilder
	Function func(*PerformanceOptimizedAssertionBuilder)
}

// AssertionResult contains the result of an assertion execution
type AssertionResult struct {
	JobID    string
	Errors   []AssertionError
	Metrics  PerformanceMetrics
	Success  bool
	Duration time.Duration
}

// NewConcurrentAssertionRunner creates a new concurrent assertion runner
func NewConcurrentAssertionRunner(maxWorkers int) *ConcurrentAssertionRunner {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}

	runner := &ConcurrentAssertionRunner{
		maxWorkers:  maxWorkers,
		jobQueue:    make(chan AssertionJob, maxWorkers*2),
		resultQueue: make(chan AssertionResult, maxWorkers*2),
		workerPool: sync.Pool{
			New: func() interface{} {
				return &assertionWorker{}
			},
		},
	}

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		go runner.worker()
	}

	return runner
}

type assertionWorker struct {
	id int
}

// worker processes assertion jobs
func (car *ConcurrentAssertionRunner) worker() {
	for job := range car.jobQueue {
		start := time.Now()

		// Execute the assertion
		job.Function(job.Builder)

		// Collect results
		result := AssertionResult{
			JobID:    job.ID,
			Errors:   job.Builder.GetErrors(),
			Metrics:  job.Builder.GetPerformanceMetrics(),
			Success:  len(job.Builder.GetErrors()) == 0,
			Duration: time.Since(start),
		}

		car.resultQueue <- result
	}
}

// SubmitJob submits an assertion job for concurrent execution
func (car *ConcurrentAssertionRunner) SubmitJob(job AssertionJob) {
	car.jobQueue <- job
}

// GetResult retrieves a completed assertion result
func (car *ConcurrentAssertionRunner) GetResult() AssertionResult {
	return <-car.resultQueue
}

// Close shuts down the concurrent assertion runner
func (car *ConcurrentAssertionRunner) Close() {
	close(car.jobQueue)
	close(car.resultQueue)
}

// MemoryManager provides memory optimization utilities
type MemoryManager struct {
	maxMemory     int64
	checkInterval time.Duration
	running       bool
	mu            sync.RWMutex
}

// NewMemoryManager creates a new memory manager
func NewMemoryManager(maxMemory int64, checkInterval time.Duration) *MemoryManager {
	return &MemoryManager{
		maxMemory:     maxMemory,
		checkInterval: checkInterval,
	}
}

// Start begins memory monitoring
func (mm *MemoryManager) Start() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if mm.running {
		return
	}

	mm.running = true
	go mm.monitor()
}

// Stop stops memory monitoring
func (mm *MemoryManager) Stop() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.running = false
}

// monitor continuously monitors memory usage
func (mm *MemoryManager) monitor() {
	ticker := time.NewTicker(mm.checkInterval)
	defer ticker.Stop()

	for range ticker.C {
		mm.mu.RLock()
		if !mm.running {
			mm.mu.RUnlock()
			return
		}
		mm.mu.RUnlock()

		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		if int64(m.Alloc) > mm.maxMemory {
			runtime.GC()
		}
	}
}
