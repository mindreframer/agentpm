package xmlquery

import "fmt"

// QuerySyntaxError represents an error in XPath query syntax
type QuerySyntaxError struct {
	Query      string `json:"query"`
	Message    string `json:"message"`
	Position   int    `json:"position,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

func (e *QuerySyntaxError) Error() string {
	return fmt.Sprintf("query syntax error: %s (query: %s)", e.Message, e.Query)
}

// QueryExecutionError represents an error during query execution
type QueryExecutionError struct {
	Query    string `json:"query"`
	Message  string `json:"message"`
	EpicFile string `json:"epic_file,omitempty"`
}

func (e *QueryExecutionError) Error() string {
	return fmt.Sprintf("query execution error: %s (query: %s)", e.Message, e.Query)
}

// FileAccessError represents an error accessing the epic file
type FileAccessError struct {
	FilePath string `json:"file_path"`
	Message  string `json:"message"`
}

func (e *FileAccessError) Error() string {
	return fmt.Sprintf("file access error: %s (file: %s)", e.Message, e.FilePath)
}

// ConfigurationError represents an error in query configuration
type ConfigurationError struct {
	Message string `json:"message"`
}

func (e *ConfigurationError) Error() string {
	return fmt.Sprintf("configuration error: %s", e.Message)
}
