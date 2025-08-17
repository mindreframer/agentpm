package testing

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TestMigrator provides utilities for migrating XML output tests to snapshot testing
type TestMigrator struct {
	normalizer *XMLNormalizer
}

// MigrationResult contains the result of a test migration
type MigrationResult struct {
	FilePath      string
	TestsFound    int
	TestsMigrated int
	XMLTestsFound int
	Errors        []error
}

// NewTestMigrator creates a new test migrator
func NewTestMigrator() *TestMigrator {
	return &TestMigrator{
		normalizer: NewXMLNormalizer(),
	}
}

// FindXMLOutputTests finds all tests that use direct XML string assertions
func (m *TestMigrator) FindXMLOutputTests(rootDir string) ([]MigrationResult, error) {
	var results []MigrationResult

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process test files
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}

		result, err := m.analyzeTestFile(path)
		if err != nil {
			return fmt.Errorf("failed to analyze %s: %w", path, err)
		}

		if result.XMLTestsFound > 0 {
			results = append(results, result)
		}

		return nil
	})

	return results, err
}

// analyzeTestFile analyzes a single test file for XML output tests
func (m *TestMigrator) analyzeTestFile(filePath string) (MigrationResult, error) {
	result := MigrationResult{
		FilePath: filePath,
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("failed to read file: %w", err))
		return result, nil
	}

	// Parse the Go file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("failed to parse Go file: %w", err))
		return result, nil
	}

	// Find test functions and analyze their content
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if strings.HasPrefix(x.Name.Name, "Test") {
				result.TestsFound++
				if m.containsXMLAssertions(string(content), x) {
					result.XMLTestsFound++
				}
			}
		}
		return true
	})

	return result, nil
}

// containsXMLAssertions checks if a test function contains XML assertions
func (m *TestMigrator) containsXMLAssertions(content string, funcDecl *ast.FuncDecl) bool {
	var funcBody string

	// If funcDecl is nil, use the entire content (for testing)
	if funcDecl == nil {
		funcBody = content
	} else {
		// Get the function body as string (simplified approach)
		funcStart := int(funcDecl.Pos())
		funcEnd := int(funcDecl.End())

		if funcStart >= len(content) || funcEnd > len(content) {
			return false
		}

		funcBody = content[funcStart:funcEnd]
	}

	// Look for patterns that indicate XML output testing
	xmlPatterns := []string{
		`assert\.Contains.*<.*>`,  // assert.Contains(t, output, "<xml>")
		`assert\.Equal.*<.*>`,     // assert.Equal(t, expected, "<xml>")
		`strings\.Contains.*<.*>`, // strings.Contains(output, "<xml>")
		`--format.*xml`,           // XML format flags
		`\.String\(\).*<.*>`,      // stdout.String() with XML
		`output.*<.*>.*</.*>`,     // XML content in output
		`<.*epic.*>`,              // Epic XML tags
		`<.*status.*>`,            // Status XML tags
		`<.*test_operation.*>`,    // Test operation XML tags
	}

	for _, pattern := range xmlPatterns {
		matched, _ := regexp.MatchString(pattern, funcBody)
		if matched {
			return true
		}
	}

	return false
}

// GenerateMigrationPlan creates a migration plan for converting XML tests to snapshots
func (m *TestMigrator) GenerateMigrationPlan(results []MigrationResult) MigrationPlan {
	plan := MigrationPlan{
		TotalFiles:     len(results),
		TotalXMLTests:  0,
		MigrationSteps: make([]MigrationStep, 0),
	}

	for _, result := range results {
		plan.TotalXMLTests += result.XMLTestsFound

		if result.XMLTestsFound > 0 {
			step := MigrationStep{
				FilePath:     result.FilePath,
				XMLTestCount: result.XMLTestsFound,
				Priority:     m.calculatePriority(result),
				Description:  fmt.Sprintf("Migrate %d XML output tests in %s", result.XMLTestsFound, filepath.Base(result.FilePath)),
			}
			plan.MigrationSteps = append(plan.MigrationSteps, step)
		}
	}

	return plan
}

// MigrationPlan represents a plan for migrating XML tests to snapshots
type MigrationPlan struct {
	TotalFiles     int
	TotalXMLTests  int
	MigrationSteps []MigrationStep
}

// MigrationStep represents a single migration step
type MigrationStep struct {
	FilePath     string
	XMLTestCount int
	Priority     string // "high", "medium", "low"
	Description  string
}

// calculatePriority determines the migration priority based on test patterns
func (m *TestMigrator) calculatePriority(result MigrationResult) string {
	// High priority: Many XML tests or core functionality
	if result.XMLTestsFound >= 5 {
		return "high"
	}

	// High priority: Core command tests
	if strings.Contains(result.FilePath, "status_test.go") ||
		strings.Contains(result.FilePath, "start_") ||
		strings.Contains(result.FilePath, "done_") {
		return "high"
	}

	// Medium priority: Other command tests
	if result.XMLTestsFound >= 2 {
		return "medium"
	}

	return "low"
}

// CreateSnapshotHelpers creates helper functions for snapshot testing in a test file
func (m *TestMigrator) CreateSnapshotHelpers() string {
	return `// Snapshot testing helpers

import (
	"testing"
	apmtesting "github.com/mindreframer/agentpm/internal/testing"
)

var snapshotTester = apmtesting.NewSnapshotTester()

// matchXMLSnapshot is a helper for XML snapshot testing
func matchXMLSnapshot(t *testing.T, xmlOutput string, testName ...string) {
	snapshotTester.MatchXMLSnapshot(t, xmlOutput, testName...)
}

// matchSnapshot is a helper for general snapshot testing
func matchSnapshot(t *testing.T, data interface{}, testName ...string) {
	snapshotTester.MatchSnapshot(t, data, testName...)
}`
}

// GenerateMigrationGuide creates a migration guide for developers
func (m *TestMigrator) GenerateMigrationGuide(plan MigrationPlan) string {
	guide := strings.Builder{}

	guide.WriteString("# XML Output Test Migration Guide\n\n")
	guide.WriteString(fmt.Sprintf("## Overview\n"))
	guide.WriteString(fmt.Sprintf("Total files to migrate: %d\n", plan.TotalFiles))
	guide.WriteString(fmt.Sprintf("Total XML tests to migrate: %d\n\n", plan.TotalXMLTests))

	guide.WriteString("## Migration Steps\n\n")

	for i, step := range plan.MigrationSteps {
		guide.WriteString(fmt.Sprintf("### %d. %s (%s priority)\n", i+1, step.Description, step.Priority))
		guide.WriteString(fmt.Sprintf("File: `%s`\n", step.FilePath))
		guide.WriteString(fmt.Sprintf("XML tests: %d\n\n", step.XMLTestCount))

		guide.WriteString("**Migration process:**\n")
		guide.WriteString("1. Add snapshot testing helpers to the file\n")
		guide.WriteString("2. Replace `assert.Contains(t, output, \"<xml>\")` with `matchXMLSnapshot(t, output)`\n")
		guide.WriteString("3. Replace `assert.Equal(t, expected, output)` with `matchXMLSnapshot(t, output)`\n")
		guide.WriteString("4. Run tests with `SNAPS_UPDATE=true go test ./...` to create initial snapshots\n")
		guide.WriteString("5. Review generated snapshots for correctness\n")
		guide.WriteString("6. Run normal tests to verify snapshot matching\n\n")
	}

	guide.WriteString("## Example Migration\n\n")
	guide.WriteString("**Before:**\n")
	guide.WriteString("```go\n")
	guide.WriteString("output := stdout.String()\n")
	guide.WriteString("assert.Contains(t, output, `<status epic=\"test-epic\">`)\n")
	guide.WriteString("assert.Contains(t, output, `<name>Test Epic</name>`)\n")
	guide.WriteString("```\n\n")

	guide.WriteString("**After:**\n")
	guide.WriteString("```go\n")
	guide.WriteString("output := stdout.String()\n")
	guide.WriteString("matchXMLSnapshot(t, output, \"status_command_xml_output\")\n")
	guide.WriteString("```\n\n")

	guide.WriteString("## Benefits\n")
	guide.WriteString("- **Less fragile:** Snapshots handle whitespace and formatting changes automatically\n")
	guide.WriteString("- **Better coverage:** Entire XML structure is validated, not just specific strings\n")
	guide.WriteString("- **Easier maintenance:** Update snapshots with a single command after intentional changes\n")
	guide.WriteString("- **Clear intent:** Snapshot names document what each test is validating\n")

	return guide.String()
}
