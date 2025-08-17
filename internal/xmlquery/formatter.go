package xmlquery

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/beevik/etree"
)

// OutputFormat represents supported output formats
type OutputFormat string

const (
	FormatXML  OutputFormat = "xml"
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
)

// Formatter provides output formatting for query results
type Formatter interface {
	Format(result *QueryResult) (string, error)
}

// XMLFormatter formats results as structured XML
type XMLFormatter struct{}

// TextFormatter formats results as human-readable text
type TextFormatter struct{}

// JSONFormatter formats results as JSON
type JSONFormatter struct{}

// NewFormatter creates the appropriate formatter for the given format
func NewFormatter(format OutputFormat) Formatter {
	switch format {
	case FormatXML:
		return &XMLFormatter{}
	case FormatText:
		return &TextFormatter{}
	case FormatJSON:
		return &JSONFormatter{}
	default:
		return &XMLFormatter{} // default to XML
	}
}

// QueryResultXML represents the XML structure for query results
type QueryResultXML struct {
	XMLName         xml.Name   `xml:"query_result"`
	Query           string     `xml:"query"`
	EpicFile        string     `xml:"epic_file,omitempty"`
	MatchCount      int        `xml:"match_count"`
	ExecutionTimeMs int        `xml:"execution_time_ms,omitempty"`
	Matches         MatchesXML `xml:"matches"`
	Message         string     `xml:"message,omitempty"`
}

// MatchesXML contains the matched elements or attributes
type MatchesXML struct {
	Elements   []ElementXML   `xml:"element,omitempty"`
	Attributes []AttributeXML `xml:"attribute,omitempty"`
	TextNodes  []TextXML      `xml:"text,omitempty"`
}

// ElementXML represents an XML element in results
type ElementXML struct {
	Tag        string         `xml:"tag,attr"`
	Attributes []AttributeXML `xml:"attr,omitempty"`
	Text       string         `xml:",chardata"`
	Children   []ElementXML   `xml:",any"`
}

// AttributeXML represents an attribute in results
type AttributeXML struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// TextXML represents text content in results
type TextXML struct {
	Content string `xml:",chardata"`
}

// Format formats the query result as XML
func (f *XMLFormatter) Format(result *QueryResult) (string, error) {
	xmlResult := &QueryResultXML{
		Query:           result.Query,
		EpicFile:        result.EpicFile,
		MatchCount:      result.MatchCount,
		ExecutionTimeMs: result.ExecutionTimeMs,
		Message:         result.Message,
	}

	// Determine result type based on query pattern
	if f.isAttributeQuery(result.Query) {
		xmlResult.Matches.Attributes = f.formatAttributes(result)
	} else if f.isTextQuery(result.Query) {
		xmlResult.Matches.TextNodes = f.formatTextNodes(result)
	} else {
		xmlResult.Matches.Elements = f.formatElements(result)
	}

	output, err := xml.MarshalIndent(xmlResult, "", "    ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal XML: %w", err)
	}

	return xml.Header + string(output), nil
}

// Format formats the query result as human-readable text
func (f *TextFormatter) Format(result *QueryResult) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Query: %s\n", result.Query))
	sb.WriteString(fmt.Sprintf("Found %d matches", result.MatchCount))

	if result.ExecutionTimeMs > 0 {
		sb.WriteString(fmt.Sprintf(" (executed in %dms)", result.ExecutionTimeMs))
	}
	sb.WriteString(":\n")

	if result.MatchCount == 0 {
		if result.Message != "" {
			sb.WriteString(fmt.Sprintf("\n%s\n", result.Message))
		} else {
			sb.WriteString("\nNo matches found.\n")
		}
		return sb.String(), nil
	}

	sb.WriteString("\n")

	// Format based on query type
	if f.isAttributeQuery(result.Query) {
		f.formatAttributesText(&sb, result)
	} else if f.isTextQuery(result.Query) {
		f.formatTextNodesText(&sb, result)
	} else {
		f.formatElementsText(&sb, result)
	}

	return sb.String(), nil
}

// QueryResultJSON represents the JSON structure for query results
type QueryResultJSON struct {
	Query           string      `json:"query"`
	EpicFile        string      `json:"epic_file,omitempty"`
	MatchCount      int         `json:"match_count"`
	ExecutionTimeMs int         `json:"execution_time_ms,omitempty"`
	Matches         interface{} `json:"matches"`
	Message         string      `json:"message,omitempty"`
}

// Format formats the query result as JSON
func (f *JSONFormatter) Format(result *QueryResult) (string, error) {
	jsonResult := &QueryResultJSON{
		Query:           result.Query,
		EpicFile:        result.EpicFile,
		MatchCount:      result.MatchCount,
		ExecutionTimeMs: result.ExecutionTimeMs,
		Message:         result.Message,
	}

	// Determine result type and format accordingly
	if f.isAttributeQuery(result.Query) {
		jsonResult.Matches = f.formatAttributesJSON(result)
	} else if f.isTextQuery(result.Query) {
		jsonResult.Matches = f.formatTextNodesJSON(result)
	} else {
		jsonResult.Matches = f.formatElementsJSON(result)
	}

	output, err := json.MarshalIndent(jsonResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(output), nil
}

// Helper methods for determining query types

func (f *XMLFormatter) isAttributeQuery(query string) bool {
	return strings.Contains(query, "/@") || strings.HasSuffix(query, "/@*")
}

func (f *XMLFormatter) isTextQuery(query string) bool {
	return strings.HasSuffix(query, "/text()")
}

func (f *TextFormatter) isAttributeQuery(query string) bool {
	return strings.Contains(query, "/@") || strings.HasSuffix(query, "/@*")
}

func (f *TextFormatter) isTextQuery(query string) bool {
	return strings.HasSuffix(query, "/text()")
}

func (f *JSONFormatter) isAttributeQuery(query string) bool {
	return strings.Contains(query, "/@") || strings.HasSuffix(query, "/@*")
}

func (f *JSONFormatter) isTextQuery(query string) bool {
	return strings.HasSuffix(query, "/text()")
}

// Helper methods for XML formatting

func (f *XMLFormatter) formatElements(result *QueryResult) []ElementXML {
	var elements []ElementXML

	for _, elem := range result.Elements {
		xmlElem := f.convertElement(elem)
		elements = append(elements, xmlElem)
	}

	return elements
}

func (f *XMLFormatter) formatAttributes(result *QueryResult) []AttributeXML {
	var attributes []AttributeXML

	// For attribute queries, we need to extract attribute values
	// This is a simplified approach - in reality, etree handles this differently
	for _, elem := range result.Elements {
		for _, attr := range elem.Attr {
			attributes = append(attributes, AttributeXML{
				Name:  attr.Key,
				Value: attr.Value,
			})
		}
	}

	return attributes
}

func (f *XMLFormatter) formatTextNodes(result *QueryResult) []TextXML {
	var textNodes []TextXML

	for _, elem := range result.Elements {
		if elem != nil && elem.Text() != "" {
			textNodes = append(textNodes, TextXML{
				Content: elem.Text(),
			})
		}
	}

	return textNodes
}

func (f *XMLFormatter) convertElement(elem *etree.Element) ElementXML {
	xmlElem := ElementXML{
		Tag:  elem.Tag,
		Text: elem.Text(),
	}

	// Convert attributes
	for _, attr := range elem.Attr {
		xmlElem.Attributes = append(xmlElem.Attributes, AttributeXML{
			Name:  attr.Key,
			Value: attr.Value,
		})
	}

	// Convert child elements (simplified - only direct children)
	for _, child := range elem.ChildElements() {
		childXML := f.convertElement(child)
		xmlElem.Children = append(xmlElem.Children, childXML)
	}

	return xmlElem
}

// Helper methods for text formatting

func (f *TextFormatter) formatElementsText(sb *strings.Builder, result *QueryResult) {
	for i, elem := range result.Elements {
		sb.WriteString(fmt.Sprintf("%s[", elem.Tag))

		// Add attributes
		for j, attr := range elem.Attr {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%s=%s", attr.Key, attr.Value))
		}
		sb.WriteString("]")

		// Add text content if present
		if text := elem.Text(); text != "" {
			// Truncate long text content
			if len(text) > 50 {
				text = text[:47] + "..."
			}
			sb.WriteString(fmt.Sprintf(":\n  %s", text))
		}

		if i < len(result.Elements)-1 {
			sb.WriteString("\n\n")
		} else {
			sb.WriteString("\n")
		}
	}
}

func (f *TextFormatter) formatAttributesText(sb *strings.Builder, result *QueryResult) {
	for i, elem := range result.Elements {
		for _, attr := range elem.Attr {
			sb.WriteString(fmt.Sprintf("%s = %s", attr.Key, attr.Value))
			if i < len(result.Elements)-1 {
				sb.WriteString("\n")
			}
		}
	}
}

func (f *TextFormatter) formatTextNodesText(sb *strings.Builder, result *QueryResult) {
	for i, elem := range result.Elements {
		if text := elem.Text(); text != "" {
			sb.WriteString(text)
			if i < len(result.Elements)-1 {
				sb.WriteString("\n")
			}
		}
	}
}

// Helper methods for JSON formatting

func (f *JSONFormatter) formatElementsJSON(result *QueryResult) interface{} {
	var elements []map[string]interface{}

	for _, elem := range result.Elements {
		element := map[string]interface{}{
			"tag":  elem.Tag,
			"text": elem.Text(),
		}

		// Add attributes
		if len(elem.Attr) > 0 {
			attributes := make(map[string]string)
			for _, attr := range elem.Attr {
				attributes[attr.Key] = attr.Value
			}
			element["attributes"] = attributes
		}

		// Add child elements (simplified)
		if children := elem.ChildElements(); len(children) > 0 {
			var childList []map[string]interface{}
			for _, child := range children {
				childElement := map[string]interface{}{
					"tag":  child.Tag,
					"text": child.Text(),
				}
				childList = append(childList, childElement)
			}
			element["children"] = childList
		}

		elements = append(elements, element)
	}

	return elements
}

func (f *JSONFormatter) formatAttributesJSON(result *QueryResult) interface{} {
	var attributes []map[string]string

	for _, elem := range result.Elements {
		for _, attr := range elem.Attr {
			attributes = append(attributes, map[string]string{
				"name":  attr.Key,
				"value": attr.Value,
			})
		}
	}

	return attributes
}

func (f *JSONFormatter) formatTextNodesJSON(result *QueryResult) interface{} {
	var textNodes []string

	for _, elem := range result.Elements {
		if text := elem.Text(); text != "" {
			textNodes = append(textNodes, text)
		}
	}

	return textNodes
}
