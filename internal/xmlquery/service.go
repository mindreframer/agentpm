package xmlquery

import (
	"fmt"
	"os"
)

// Service provides high-level XML query operations for epic files
type Service struct {
	engine QueryEngine
}

// NewService creates a new XML query service
func NewService() *Service {
	return &Service{
		engine: NewEngine(),
	}
}

// QueryEpicFile executes an XPath query against the specified epic file
func (s *Service) QueryEpicFile(filePath, xpathExpr string) (*QueryResult, error) {
	// Validate file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, &FileAccessError{
			FilePath: filePath,
			Message:  "epic file does not exist",
		}
	}

	// Load the epic file
	if err := s.engine.LoadDocument(filePath); err != nil {
		return nil, &FileAccessError{
			FilePath: filePath,
			Message:  fmt.Sprintf("failed to load epic file: %v", err),
		}
	}

	// Execute the query
	result, err := s.engine.Execute(xpathExpr)
	if err != nil {
		return nil, err
	}

	// Add helpful message for empty results
	if result.IsEmpty() {
		result.Message = "No elements found matching query"
	}

	return result, nil
}

// QueryEpicFileFormatted executes an XPath query and returns formatted output
func (s *Service) QueryEpicFileFormatted(filePath, xpathExpr string, format OutputFormat) (string, error) {
	result, err := s.QueryEpicFile(filePath, xpathExpr)
	if err != nil {
		return "", err
	}

	formatter := NewFormatter(format)
	return formatter.Format(result)
}

// ValidateQuery validates XPath syntax without executing against a file
func (s *Service) ValidateQuery(xpathExpr string) error {
	return s.engine.ValidateQuery(xpathExpr)
}

// GetSupportedPatterns returns examples of supported XPath patterns
func (s *Service) GetSupportedPatterns() []string {
	return []string{
		"//task",                            // All task elements
		"//phase[@status='done']",           // Phases with status attribute
		"//task[@phase_id='1A']",            // Tasks in specific phase
		"//metadata/assignee",               // Nested elements
		"//test[@id='1A_1']",                // Elements by ID
		"//description/text()",              // Text content
		"//phase/@name",                     // Attribute values
		"//task[1]",                         // Position-based selection
		"//epic/*",                          // All child elements
		"//events/event[@type='completed']", // Complex attribute filtering
	}
}
