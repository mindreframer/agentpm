package xmlquery

import (
	"fmt"
	"time"

	"github.com/beevik/etree"
)

// QueryEngine provides XPath-like query capabilities for XML documents using etree
type QueryEngine interface {
	// LoadDocument loads an XML document from file path
	LoadDocument(filePath string) error

	// Execute runs an XPath query against the loaded document
	Execute(xpathExpr string) (*QueryResult, error)

	// ValidateQuery validates XPath syntax without executing
	ValidateQuery(xpathExpr string) error

	// GetDocument returns the currently loaded document
	GetDocument() *etree.Document
}

// Engine implements QueryEngine using etree library
type Engine struct {
	doc      *etree.Document
	filePath string
	cache    *QueryCache
}

// NewEngine creates a new XPath query engine with query caching
func NewEngine() *Engine {
	return &Engine{
		cache: NewQueryCache(100), // cache up to 100 compiled queries
	}
}

// LoadDocument loads an XML document from the specified file path
func (e *Engine) LoadDocument(filePath string) error {
	doc := etree.NewDocument()
	err := doc.ReadFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load XML document from %s: %w", filePath, err)
	}

	e.doc = doc
	e.filePath = filePath
	return nil
}

// Execute runs an XPath query against the loaded document
func (e *Engine) Execute(xpathExpr string) (*QueryResult, error) {
	if e.doc == nil {
		return nil, fmt.Errorf("no document loaded")
	}

	// Validate query syntax first
	if err := e.ValidateQuery(xpathExpr); err != nil {
		return nil, err
	}

	startTime := time.Now()

	// Try to get compiled path from cache first
	var path etree.Path
	var err error

	if cachedPath, found := e.cache.Get(xpathExpr); found {
		path = cachedPath
	} else {
		// Compile the XPath query using etree
		path, err = etree.CompilePath(xpathExpr)
		if err != nil {
			return nil, &QuerySyntaxError{
				Query:      xpathExpr,
				Message:    fmt.Sprintf("XPath compilation failed: %v", err),
				Suggestion: "Check XPath syntax and ensure brackets are properly closed",
			}
		}
		// Cache the compiled path for future use
		e.cache.Put(xpathExpr, path)
	}

	// Find elements matching the path
	elements := e.doc.FindElementsPath(path)

	executionTime := time.Since(startTime)

	// Create result
	result := &QueryResult{
		Query:           xpathExpr,
		EpicFile:        e.filePath,
		MatchCount:      len(elements),
		ExecutionTimeMs: int(executionTime.Nanoseconds() / 1000000),
		Elements:        elements,
	}

	return result, nil
}

// ValidateQuery validates XPath syntax without executing the query
func (e *Engine) ValidateQuery(xpathExpr string) error {
	if xpathExpr == "" {
		return &QuerySyntaxError{
			Query:      xpathExpr,
			Message:    "XPath expression cannot be empty",
			Suggestion: "Provide a valid XPath expression, e.g., '//task' or '//phase[@status=\"done\"]'",
		}
	}

	// Try to compile the path to validate syntax
	_, err := etree.CompilePath(xpathExpr)
	if err != nil {
		return &QuerySyntaxError{
			Query:      xpathExpr,
			Message:    fmt.Sprintf("Invalid XPath syntax: %v", err),
			Suggestion: e.getSuggestionForError(xpathExpr, err),
		}
	}

	return nil
}

// GetDocument returns the currently loaded XML document
func (e *Engine) GetDocument() *etree.Document {
	return e.doc
}

// getSuggestionForError provides helpful suggestions based on common XPath errors
func (e *Engine) getSuggestionForError(query string, err error) string {
	errStr := err.Error()

	// Common error patterns and suggestions
	if contains(errStr, "bracket") || contains(errStr, "[") || contains(errStr, "]") {
		return "Check for missing or unmatched brackets in predicates"
	}

	if contains(errStr, "quote") || contains(errStr, "'") || contains(errStr, "\"") {
		return "Check for missing or unmatched quotes in attribute values"
	}

	if contains(errStr, "unexpected") {
		return "Check XPath syntax - common patterns: //element, //element[@attr='value'], //element[position()]"
	}

	return "Verify XPath syntax follows etree supported patterns"
}

// contains is a helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			(len(s) > len(substr) && indexOf(s, substr) >= 0))))
}

// indexOf returns the index of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
