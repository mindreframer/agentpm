package xmlquery

import (
	"github.com/beevik/etree"
)

// QueryResult represents the result of an XPath query execution
type QueryResult struct {
	Query           string           `json:"query"`
	EpicFile        string           `json:"epic_file"`
	MatchCount      int              `json:"match_count"`
	ExecutionTimeMs int              `json:"execution_time_ms"`
	Elements        []*etree.Element `json:"-"` // Elements are not directly JSON serializable
	Message         string           `json:"message,omitempty"`
}

// IsEmpty returns true if the query result contains no matches
func (qr *QueryResult) IsEmpty() bool {
	return qr.MatchCount == 0
}

// GetElementTexts extracts text content from all matched elements
func (qr *QueryResult) GetElementTexts() []string {
	var texts []string
	for _, elem := range qr.Elements {
		if elem != nil {
			texts = append(texts, elem.Text())
		}
	}
	return texts
}

// GetAttributeValues extracts attribute values if the query targets attributes
func (qr *QueryResult) GetAttributeValues(attrName string) []string {
	var values []string
	for _, elem := range qr.Elements {
		if elem != nil {
			if attr := elem.SelectAttr(attrName); attr != nil {
				values = append(values, attr.Value)
			}
		}
	}
	return values
}

// GetElementsByTag filters matched elements by tag name
func (qr *QueryResult) GetElementsByTag(tag string) []*etree.Element {
	var filtered []*etree.Element
	for _, elem := range qr.Elements {
		if elem != nil && elem.Tag == tag {
			filtered = append(filtered, elem)
		}
	}
	return filtered
}
