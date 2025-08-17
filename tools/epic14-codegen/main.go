// Epic 14 Code Generation Tool
// Generates builder patterns and test scaffolding for Epic 14 framework

package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Config struct {
	InputFile   string
	OutputDir   string
	PackageName string
	TestType    string
	Verbose     bool
}

type EpicDefinition struct {
	Name        string
	ID          string
	Phases      []PhaseDefinition
	Tasks       []TaskDefinition
	Tests       []TestDefinition
	Events      []EventDefinition
	PackageName string
}

type PhaseDefinition struct {
	ID           string
	Name         string
	Description  string
	Dependencies []string
}

type TaskDefinition struct {
	ID      string
	PhaseID string
	Name    string
	Status  string
}

type TestDefinition struct {
	ID     string
	TaskID string
	Name   string
	Type   string
}

type EventDefinition struct {
	Type        string
	Description string
}

func main() {
	config := parseFlags()

	if config.Verbose {
		fmt.Printf("Epic 14 Code Generator v1.0\n")
		fmt.Printf("Input: %s\n", config.InputFile)
		fmt.Printf("Output: %s\n", config.OutputDir)
	}

	epic, err := parseEpicDefinition(config.InputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing epic definition: %v\n", err)
		os.Exit(1)
	}

	epic.PackageName = config.PackageName

	if err := generateCode(epic, config); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
		os.Exit(1)
	}

	if config.Verbose {
		fmt.Println("Code generation completed successfully!")
	}
}

func parseFlags() Config {
	var config Config

	flag.StringVar(&config.InputFile, "input", "", "Input epic definition file (.epic, .yaml, .json)")
	flag.StringVar(&config.OutputDir, "output", ".", "Output directory for generated code")
	flag.StringVar(&config.PackageName, "package", "tests", "Package name for generated code")
	flag.StringVar(&config.TestType, "type", "unit", "Test type (unit, integration, performance)")
	flag.BoolVar(&config.Verbose, "verbose", false, "Verbose output")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -input epic.yaml -output ./tests -package epic_tests\n", os.Args[0])
	}

	flag.Parse()

	if config.InputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -input flag is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	return config
}

func parseEpicDefinition(filename string) (*EpicDefinition, error) {
	ext := filepath.Ext(filename)

	switch ext {
	case ".epic":
		return parseEpicFile(filename)
	case ".yaml", ".yml":
		return parseYAMLFile(filename)
	case ".json":
		return parseJSONFile(filename)
	case ".go":
		return parseGoStructFile(filename)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

func parseEpicFile(filename string) (*EpicDefinition, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	epic := &EpicDefinition{}
	scanner := bufio.NewScanner(file)

	var currentSection string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.Trim(line, "[]")
			continue
		}

		switch currentSection {
		case "epic":
			if strings.HasPrefix(line, "name=") {
				epic.Name = strings.TrimPrefix(line, "name=")
			} else if strings.HasPrefix(line, "id=") {
				epic.ID = strings.TrimPrefix(line, "id=")
			}
		case "phases":
			if parts := strings.Split(line, ":"); len(parts) >= 2 {
				phase := PhaseDefinition{
					ID:   parts[0],
					Name: parts[1],
				}
				if len(parts) > 2 {
					phase.Description = parts[2]
				}
				epic.Phases = append(epic.Phases, phase)
			}
		case "tasks":
			if parts := strings.Split(line, ":"); len(parts) >= 3 {
				task := TaskDefinition{
					ID:      parts[0],
					PhaseID: parts[1],
					Name:    parts[2],
					Status:  "pending",
				}
				epic.Tasks = append(epic.Tasks, task)
			}
		case "events":
			if parts := strings.Split(line, ":"); len(parts) >= 2 {
				event := EventDefinition{
					Type:        parts[0],
					Description: parts[1],
				}
				epic.Events = append(epic.Events, event)
			}
		}
	}

	return epic, scanner.Err()
}

func parseYAMLFile(filename string) (*EpicDefinition, error) {
	// Implementation for YAML parsing would go here
	// For now, return a basic structure
	return &EpicDefinition{
		Name: "Generated Epic",
		ID:   "generated-epic",
	}, nil
}

func parseJSONFile(filename string) (*EpicDefinition, error) {
	// Implementation for JSON parsing would go here
	return &EpicDefinition{
		Name: "Generated Epic",
		ID:   "generated-epic",
	}, nil
}

func parseGoStructFile(filename string) (*EpicDefinition, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	epic := &EpicDefinition{}

	// Extract epic definition from Go structs
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.StructType:
			// Parse struct fields to extract epic definition
			for _, field := range x.Fields.List {
				if len(field.Names) > 0 {
					fieldName := field.Names[0].Name
					if fieldName == "ID" || fieldName == "Name" {
						// Extract string literal value if available
						// This is a simplified implementation
					}
				}
			}
		}
		return true
	})

	// Set defaults if not found in Go file
	if epic.Name == "" {
		epic.Name = "Generated Epic"
	}
	if epic.ID == "" {
		epic.ID = "generated-epic"
	}

	return epic, nil
}

func generateCode(epic *EpicDefinition, config Config) error {
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return err
	}

	generators := map[string]func(*EpicDefinition, Config) error{
		"builder":    generateBuilderCode,
		"test":       generateTestCode,
		"assertions": generateAssertionCode,
		"helpers":    generateHelperCode,
	}

	for name, generator := range generators {
		if config.Verbose {
			fmt.Printf("Generating %s code...\n", name)
		}

		if err := generator(epic, config); err != nil {
			return fmt.Errorf("error generating %s code: %v", name, err)
		}
	}

	return nil
}

func generateBuilderCode(epic *EpicDefinition, config Config) error {
	tmpl := `// Code generated by Epic 14 Code Generator. DO NOT EDIT.

package {{.PackageName}}

import (
	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/testing/executor"
)

// {{.Name}}Builder provides a fluent API for building {{.Name}} test scenarios
type {{.Name}}Builder struct {
	epic   *epic.Epic
	chain  *executor.TransitionChain
	config BuilderConfig
}

type BuilderConfig struct {
	Debug         bool
	Timeout       time.Duration
	EnableTracing bool
}

// New{{.Name}}Builder creates a new builder for {{.Name}} tests
func New{{.Name}}Builder() *{{.Name}}Builder {
	return &{{.Name}}Builder{
		epic: &epic.Epic{
			ID:     "{{.ID}}",
			Name:   "{{.Name}}",
			Status: epic.StatusPending,
			Phases: []epic.Phase{
				{{range .Phases}}
				{ID: "{{.ID}}", Name: "{{.Name}}", Status: epic.StatusPlanning},
				{{end}}
			},
			Tasks: []epic.Task{
				{{range .Tasks}}
				{ID: "{{.ID}}", PhaseID: "{{.PhaseID}}", Name: "{{.Name}}", Status: epic.StatusPlanning},
				{{end}}
			},
		},
		chain: executor.NewTransitionChain(),
		config: BuilderConfig{
			Timeout: 30 * time.Second,
		},
	}
}

// WithDebug enables debug mode for the builder
func (b *{{.Name}}Builder) WithDebug(enabled bool) *{{.Name}}Builder {
	b.config.Debug = enabled
	return b
}

// WithTimeout sets the execution timeout
func (b *{{.Name}}Builder) WithTimeout(timeout time.Duration) *{{.Name}}Builder {
	b.config.Timeout = timeout
	return b
}

// Start begins the epic execution
func (b *{{.Name}}Builder) Start() *{{.Name}}Builder {
	b.chain = b.chain.StartEpic(b.epic.ID)
	return b
}

{{range .Phases}}
// Execute{{.Name}} executes the {{.Name}} phase
func (b *{{$.Name}}Builder) Execute{{.Name}}() *{{$.Name}}Builder {
	b.chain = b.chain.ExecutePhase("{{.ID}}")
	return b
}
{{end}}

// Complete completes the epic execution
func (b *{{.Name}}Builder) Complete() *{{.Name}}Builder {
	b.chain = b.chain.CompleteEpic()
	return b
}

// Execute runs the built transition chain
func (b *{{.Name}}Builder) Execute() *executor.TransitionChainResult {
	if b.config.Timeout > 0 {
		b.chain = b.chain.WithTimeout(b.config.Timeout)
	}
	
	return b.chain.Execute()
}

// Build returns the configured epic
func (b *{{.Name}}Builder) Build() *epic.Epic {
	return b.epic
}
`

	return executeTemplate(tmpl, epic, filepath.Join(config.OutputDir, "builder.go"))
}

func generateTestCode(epic *EpicDefinition, config Config) error {
	var tmpl string

	switch config.TestType {
	case "unit":
		tmpl = generateUnitTestTemplate()
	case "integration":
		tmpl = generateIntegrationTestTemplate()
	case "performance":
		tmpl = generatePerformanceTestTemplate()
	default:
		tmpl = generateUnitTestTemplate()
	}

	filename := fmt.Sprintf("%s_test.go", strings.ToLower(config.TestType))
	return executeTemplate(tmpl, epic, filepath.Join(config.OutputDir, filename))
}

func generateUnitTestTemplate() string {
	return `// Code generated by Epic 14 Code Generator. DO NOT EDIT.

package {{.PackageName}}

import (
	"testing"
	"time"
	"github.com/mindreframer/agentpm/internal/testing/assertions"
)

func Test{{.Name}}BasicExecution(t *testing.T) {
	result := New{{.Name}}Builder().
		WithDebug(testing.Verbose()).
		Start().
		{{range .Phases}}Execute{{.Name}}().
		{{end}}Complete().
		Execute()
	
	assertions.Assert(result).
		EpicStatus("completed").
		{{range .Phases}}PhaseStatus("{{.ID}}", "completed").
		{{end}}NoErrors().
		ExecutionTime(10 * time.Second).
		MustPass()
}

{{range .Phases}}
func Test{{$.Name}}_{{.Name}}Phase(t *testing.T) {
	result := New{{$.Name}}Builder().
		WithDebug(testing.Verbose()).
		Start().
		Execute{{.Name}}().
		Execute()
	
	assertions.Assert(result).
		PhaseStatus("{{.ID}}", "completed").
		NoErrors().
		MustPass()
}
{{end}}

func Test{{.Name}}EventSequence(t *testing.T) {
	result := New{{.Name}}Builder().
		Start().
		{{range .Phases}}Execute{{.Name}}().
		{{end}}Complete().
		Execute()
	
	assertions.Assert(result).
		EventSequence([]string{
			"epic_started",
			{{range .Phases}}"phase_{{.ID}}_started",
			"phase_{{.ID}}_completed",
			{{end}}"epic_completed",
		}).
		MustPass()
}

func Test{{.Name}}ErrorHandling(t *testing.T) {
	// Test invalid phase execution
	result := New{{.Name}}Builder().
		Start().
		Execute() // Don't execute all phases
	
	// Epic should not be completed if phases are incomplete
	err := assertions.Assert(result).
		EpicStatus("completed").
		Check()
	
	if err == nil {
		t.Error("Expected error when epic is incomplete, but got none")
	}
}
`
}

func generateIntegrationTestTemplate() string {
	return `// Code generated by Epic 14 Code Generator. DO NOT EDIT.
//go:build integration
// +build integration

package {{.PackageName}}

import (
	"testing"
	"time"
	"github.com/mindreframer/agentpm/internal/testing/assertions"
)

func TestIntegration{{.Name}}FullWorkflow(t *testing.T) {
	result := New{{.Name}}Builder().
		WithDebug(false). // Reduced debug for integration tests
		WithTimeout(60 * time.Second).
		Start().
		{{range .Phases}}Execute{{.Name}}().
		{{end}}Complete().
		Execute()
	
	assertions.Assert(result).
		EpicStatus("completed").
		{{range .Phases}}PhaseStatus("{{.ID}}", "completed").
		{{end}}NoErrors().
		ExecutionTime(60 * time.Second).
		{{range .Events}}HasEvent("{{.Type}}").
		{{end}}MustPass()
}

func TestIntegration{{.Name}}StateProgression(t *testing.T) {
	result := New{{.Name}}Builder().
		WithDebug(false).
		Start().
		{{range .Phases}}Execute{{.Name}}().
		{{end}}Complete().
		Execute()
	
	assertions.Assert(result).
		StateProgression([]string{
			"pending",
			{{range .Phases}}"phase_{{.ID}}_active",
			{{end}}"completed",
		}).
		MustPass()
}

func TestIntegration{{.Name}}Dependencies(t *testing.T) {
	result := New{{.Name}}Builder().
		WithTimeout(120 * time.Second).
		Start().
		{{range .Phases}}Execute{{.Name}}().
		{{end}}Complete().
		Execute()
	
	assertions.Assert(result).
		CustomAssertion("dependency_validation", func(result *executor.TransitionChainResult) error {
			// Validate that phases executed in correct order
			events := result.FinalState.Events
			{{range $i, $phase := .Phases}}
			{{if gt $i 0}}
			// Phase {{$phase.Name}} should start after previous phase completed
			prevCompleted := false
			currentStarted := false
			
			for _, event := range events {
				if event.Type == "phase_{{index $.Phases (sub $i 1) .ID}}_completed" {
					prevCompleted = true
				}
				if event.Type == "phase_{{$phase.ID}}_started" && prevCompleted {
					currentStarted = true
					break
				}
			}
			
			if !currentStarted {
				return fmt.Errorf("phase {{$phase.Name}} did not start after dependencies")
			}
			{{end}}
			{{end}}
			
			return nil
		}).
		MustPass()
}
`
}

func generatePerformanceTestTemplate() string {
	return `// Code generated by Epic 14 Code Generator. DO NOT EDIT.
//go:build performance
// +build performance

package {{.PackageName}}

import (
	"testing"
	"time"
	"runtime"
	"github.com/mindreframer/agentpm/internal/testing/assertions"
)

func Benchmark{{.Name}}Execution(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := New{{.Name}}Builder().
			WithDebug(false). // No debug for performance tests
			Start().
			{{range .Phases}}Execute{{.Name}}().
			{{end}}Complete().
			Execute()
		
		assertions.Assert(result).
			WithDebugMode(assertions.DebugOff).
			EpicStatus("completed").
			MustPass()
	}
}

func Test{{.Name}}PerformanceBaseline(t *testing.T) {
	const iterations = 100
	
	var baselineMemory runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&baselineMemory)
	
	start := time.Now()
	
	for i := 0; i < iterations; i++ {
		result := New{{.Name}}Builder().
			WithDebug(false).
			Start().
			{{range .Phases}}Execute{{.Name}}().
			{{end}}Complete().
			Execute()
		
		assertions.Assert(result).
			WithDebugMode(assertions.DebugOff).
			EpicStatus("completed").
			MustPass()
		
		// Periodic cleanup
		if i%10 == 9 {
			runtime.GC()
		}
	}
	
	duration := time.Since(start)
	avgDuration := duration / iterations
	
	runtime.GC()
	var finalMemory runtime.MemStats
	runtime.ReadMemStats(&finalMemory)
	
	memoryGrowth := int64(finalMemory.Alloc) - int64(baselineMemory.Alloc)
	
	t.Logf("Performance baseline for {{.Name}}:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Total time: %v", duration)
	t.Logf("  Average time per iteration: %v", avgDuration)
	t.Logf("  Memory growth: %d bytes", memoryGrowth)
	
	// Performance assertions
	maxAvgDuration := 100 * time.Millisecond
	if avgDuration > maxAvgDuration {
		t.Errorf("Average execution time too slow: %v > %v", avgDuration, maxAvgDuration)
	}
	
	maxMemoryGrowth := int64(10 * 1024 * 1024) // 10MB
	if memoryGrowth > maxMemoryGrowth {
		t.Errorf("Memory growth too large: %d bytes > %d bytes", memoryGrowth, maxMemoryGrowth)
	}
}

func Test{{.Name}}ConcurrentExecution(t *testing.T) {
	const numGoroutines = 10
	
	var wg sync.WaitGroup
	results := make(chan *executor.TransitionChainResult, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			result := New{{.Name}}Builder().
				WithDebug(false).
				Start().
				{{range .Phases}}Execute{{.Name}}().
				{{end}}Complete().
				Execute()
			
			results <- result
		}(i)
	}
	
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Validate all results
	count := 0
	for result := range results {
		count++
		assertions.Assert(result).
			WithDebugMode(assertions.DebugOff).
			EpicStatus("completed").
			MustPass()
	}
	
	if count != numGoroutines {
		t.Errorf("Expected %d results, got %d", numGoroutines, count)
	}
}
`
}

func generateAssertionCode(epic *EpicDefinition, config Config) error {
	tmpl := `// Code generated by Epic 14 Code Generator. DO NOT EDIT.

package {{.PackageName}}

import (
	"fmt"
	"github.com/mindreframer/agentpm/internal/testing/assertions"
	"github.com/mindreframer/agentpm/internal/testing/executor"
)

// Assert{{.Name}} creates a specialized assertion builder for {{.Name}} tests
func Assert{{.Name}}(result *executor.TransitionChainResult) *{{.Name}}Assertions {
	return &{{.Name}}Assertions{
		AssertionBuilder: assertions.Assert(result),
		result:          result,
	}
}

// {{.Name}}Assertions provides domain-specific assertions for {{.Name}}
type {{.Name}}Assertions struct {
	*assertions.AssertionBuilder
	result *executor.TransitionChainResult
}

// AllPhasesCompleted verifies that all {{.Name}} phases are completed
func (a *{{.Name}}Assertions) AllPhasesCompleted() *{{.Name}}Assertions {
	{{range .Phases}}
	a.AssertionBuilder.PhaseStatus("{{.ID}}", "completed")
	{{end}}
	return a
}

{{range .Phases}}
// {{.Name}}PhaseCompleted verifies that the {{.Name}} phase is completed
func (a *{{$.Name}}Assertions) {{.Name}}PhaseCompleted() *{{$.Name}}Assertions {
	a.AssertionBuilder.PhaseStatus("{{.ID}}", "completed")
	return a
}
{{end}}

{{range .Tasks}}
// Task{{.Name}}Completed verifies that task {{.Name}} is completed
func (a *{{$.Name}}Assertions) Task{{.Name}}Completed() *{{$.Name}}Assertions {
	a.AssertionBuilder.TaskStatus("{{.ID}}", "completed")
	return a
}
{{end}}

// ValidateBusinessRules performs {{.Name}}-specific business rule validation
func (a *{{.Name}}Assertions) ValidateBusinessRules() *{{.Name}}Assertions {
	a.AssertionBuilder.CustomAssertion("{{.ID}}_business_rules", func(result *executor.TransitionChainResult) error {
		epic := result.FinalState
		
		// Validate phase dependencies
		{{range .Phases}}
		{{if .Dependencies}}
		if phase := findPhase(epic, "{{.ID}}"); phase != nil && phase.Status == "completed" {
			{{range .Dependencies}}
			if depPhase := findPhase(epic, "{{.}}"); depPhase == nil || depPhase.Status != "completed" {
				return fmt.Errorf("phase {{$.ID}} completed without dependency {{.}} being completed")
			}
			{{end}}
		}
		{{end}}
		{{end}}
		
		// Validate required events
		{{range .Events}}
		if !hasEvent(epic.Events, "{{.Type}}") {
			return fmt.Errorf("required event {{.Type}} not found")
		}
		{{end}}
		
		return nil
	})
	return a
}

// ValidatePerformance checks {{.Name}}-specific performance requirements
func (a *{{.Name}}Assertions) ValidatePerformance() *{{.Name}}Assertions {
	a.AssertionBuilder.
		ExecutionTime(30 * time.Second).                    // Max execution time
		PerformanceBenchmark(30*time.Second, 100)          // Time + memory limits
	return a
}

// MustPass executes all assertions and panics on failure
func (a *{{.Name}}Assertions) MustPass() {
	a.AssertionBuilder.MustPass()
}

// Check executes all assertions and returns first error
func (a *{{.Name}}Assertions) Check() error {
	return a.AssertionBuilder.Check()
}

// Helper functions
func findPhase(epic *epic.Epic, phaseID string) *epic.Phase {
	for i := range epic.Phases {
		if epic.Phases[i].ID == phaseID {
			return &epic.Phases[i]
		}
	}
	return nil
}

func hasEvent(events []epic.Event, eventType string) bool {
	for _, event := range events {
		if event.Type == eventType {
			return true
		}
	}
	return false
}
`

	return executeTemplate(tmpl, epic, filepath.Join(config.OutputDir, "assertions.go"))
}

func generateHelperCode(epic *EpicDefinition, config Config) error {
	tmpl := `// Code generated by Epic 14 Code Generator. DO NOT EDIT.

package {{.PackageName}}

import (
	"fmt"
	"testing"
	"time"
	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/testing/executor"
)

// Test data constants
const (
	{{.Name}}ID = "{{.ID}}"
	{{.Name}}Name = "{{.Name}}"
	DefaultTimeout = 30 * time.Second
)

// Phase IDs
const (
	{{range .Phases}}
	Phase{{.Name}}ID = "{{.ID}}"
	{{end}}
)

// Task IDs
const (
	{{range .Tasks}}
	Task{{.Name}}ID = "{{.ID}}"
	{{end}}
)

// Event types
const (
	{{range .Events}}
	Event{{.Type}} = "{{.Type}}"
	{{end}}
)

// Test data builders

// Create{{.Name}}Epic creates a standard {{.Name}} epic for testing
func Create{{.Name}}Epic() *epic.Epic {
	return &epic.Epic{
		ID:     {{.Name}}ID,
		Name:   {{.Name}}Name,
		Status: epic.StatusPending,
		Phases: []epic.Phase{
			{{range .Phases}}
			{
				ID:          Phase{{.Name}}ID,
				Name:        "{{.Name}}",
				Description: "{{.Description}}",
				Status:      epic.StatusPlanning,
			},
			{{end}}
		},
		Tasks: []epic.Task{
			{{range .Tasks}}
			{
				ID:      Task{{.Name}}ID,
				PhaseID: "{{.PhaseID}}",
				Name:    "{{.Name}}",
				Status:  epic.StatusPlanning,
			},
			{{end}}
		},
	}
}

// Quick test helpers

// QuickTest{{.Name}}Success runs a basic successful {{.Name}} test
func QuickTest{{.Name}}Success(t *testing.T) {
	result := New{{.Name}}Builder().
		WithDebug(testing.Verbose()).
		Start().
		{{range .Phases}}Execute{{.Name}}().
		{{end}}Complete().
		Execute()
	
	Assert{{.Name}}(result).
		AllPhasesCompleted().
		ValidateBusinessRules().
		MustPass()
}

// QuickTest{{.Name}}Performance runs a basic performance test for {{.Name}}
func QuickTest{{.Name}}Performance(t *testing.T) {
	result := New{{.Name}}Builder().
		WithDebug(false).
		Start().
		{{range .Phases}}Execute{{.Name}}().
		{{end}}Complete().
		Execute()
	
	Assert{{.Name}}(result).
		ValidatePerformance().
		MustPass()
}

// Test utilities

// Setup{{.Name}}TestSuite prepares a test suite for {{.Name}} testing
func Setup{{.Name}}TestSuite(t *testing.T) *{{.Name}}TestSuite {
	return &{{.Name}}TestSuite{
		t:       t,
		builder: New{{.Name}}Builder(),
	}
}

type {{.Name}}TestSuite struct {
	t       *testing.T
	builder *{{.Name}}Builder
}

// WithTimeout sets the timeout for the test suite
func (suite *{{.Name}}TestSuite) WithTimeout(timeout time.Duration) *{{.Name}}TestSuite {
	suite.builder.WithTimeout(timeout)
	return suite
}

// WithDebug enables debug mode for the test suite
func (suite *{{.Name}}TestSuite) WithDebug(enabled bool) *{{.Name}}TestSuite {
	suite.builder.WithDebug(enabled)
	return suite
}

// RunBasicTest executes a basic {{.Name}} test
func (suite *{{.Name}}TestSuite) RunBasicTest() {
	result := suite.builder.
		Start().
		{{range .Phases}}Execute{{.Name}}().
		{{end}}Complete().
		Execute()
	
	Assert{{.Name}}(result).
		AllPhasesCompleted().
		ValidateBusinessRules().
		MustPass()
}

{{range .Phases}}
// Run{{.Name}}PhaseTest executes only the {{.Name}} phase
func (suite *{{$.Name}}TestSuite) Run{{.Name}}PhaseTest() {
	result := suite.builder.
		Start().
		Execute{{.Name}}().
		Execute()
	
	Assert{{$.Name}}(result).
		{{.Name}}PhaseCompleted().
		MustPass()
}
{{end}}

// Error scenarios

// Test{{.Name}}IncompleteScenario tests {{.Name}} with incomplete execution
func Test{{.Name}}IncompleteScenario(t *testing.T) {
	result := New{{.Name}}Builder().
		Start().
		// Don't execute all phases
		Execute()
	
	// Should have errors or incomplete status
	err := Assert{{.Name}}(result).
		AllPhasesCompleted().
		Check()
	
	if err == nil {
		t.Error("Expected error for incomplete {{.Name}} execution")
	}
}

// Benchmark helpers

// Benchmark{{.Name}}StandardExecution benchmarks standard {{.Name}} execution
func Benchmark{{.Name}}StandardExecution(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := New{{.Name}}Builder().
			WithDebug(false).
			Start().
			{{range .Phases}}Execute{{.Name}}().
			{{end}}Complete().
			Execute()
		
		if result.FinalState.Status != "completed" {
			b.Fatalf("{{.Name}} execution failed")
		}
	}
}
`

	return executeTemplate(tmpl, epic, filepath.Join(config.OutputDir, "helpers.go"))
}

func executeTemplate(tmplText string, data interface{}, outputFile string) error {
	tmpl, err := template.New("codegen").Funcs(template.FuncMap{
		"sub": func(a, b int) int { return a - b },
	}).Parse(tmplText)
	if err != nil {
		return err
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}
