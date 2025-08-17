package testing

import (
	"fmt"
	"regexp"
	"strings"
)

// XMLNormalizer provides utilities for normalizing XML content for snapshot testing
type XMLNormalizer struct {
	config NormalizationConfig
}

// NormalizationConfig holds configuration for XML normalization
type NormalizationConfig struct {
	// IndentSize controls the indentation (spaces per level)
	IndentSize int
	// RemoveWhitespace removes extra whitespace between elements
	RemoveWhitespace bool
	// SortAttributes sorts XML attributes alphabetically
	SortAttributes bool
	// NormalizeTimestamps replaces timestamp values with placeholders
	NormalizeTimestamps bool
	// TimestampFields contains field names to normalize
	TimestampFields []string
}

// DefaultNormalizationConfig returns a default configuration for XML normalization
func DefaultNormalizationConfig() NormalizationConfig {
	return NormalizationConfig{
		IndentSize:          2,
		RemoveWhitespace:    true,
		SortAttributes:      true,
		NormalizeTimestamps: true,
		TimestampFields:     []string{"created_at", "updated_at", "started_at", "completed_at", "passed_at", "failed_at", "cancelled_at"},
	}
}

// NewXMLNormalizer creates a new XML normalizer with default configuration
func NewXMLNormalizer() *XMLNormalizer {
	return &XMLNormalizer{
		config: DefaultNormalizationConfig(),
	}
}

// NewXMLNormalizerWithConfig creates a new XML normalizer with custom configuration
func NewXMLNormalizerWithConfig(config NormalizationConfig) *XMLNormalizer {
	return &XMLNormalizer{
		config: config,
	}
}

// NormalizeXML normalizes XML content for consistent snapshot comparisons
func (n *XMLNormalizer) NormalizeXML(xmlData string) (string, error) {
	if xmlData == "" {
		return "", nil
	}

	// Clean up input
	xmlData = strings.TrimSpace(xmlData)

	// Handle non-XML content (error messages, plain text, etc.)
	if !strings.HasPrefix(xmlData, "<") {
		return xmlData, nil
	}

	// Parse and reformat XML
	normalized, err := n.parseAndReformat(xmlData)
	if err != nil {
		// If parsing fails, return original content (might be malformed XML in error cases)
		return xmlData, nil
	}

	// Apply additional normalizations
	if n.config.NormalizeTimestamps {
		normalized = n.normalizeTimestamps(normalized)
	}

	if n.config.SortAttributes {
		normalized = n.sortXMLAttributes(normalized)
	}

	if n.config.RemoveWhitespace {
		normalized = n.removeExtraWhitespace(normalized)
	}

	return strings.TrimSpace(normalized), nil
}

// parseAndReformat parses XML and reformats it with consistent indentation
func (n *XMLNormalizer) parseAndReformat(xmlData string) (string, error) {
	// For now, return the input as-is since the generic XML parsing approach
	// doesn't work well with arbitrary XML structures.
	// The other normalization steps (timestamp, attribute sorting) will still work.
	return xmlData, nil
}

// normalizeTimestamps replaces timestamp values with normalized placeholders
func (n *XMLNormalizer) normalizeTimestamps(xmlData string) string {
	result := xmlData

	for _, field := range n.config.TimestampFields {
		// Pattern for XML elements: <created_at>2025-08-16T09:00:00Z</created_at>
		elementPattern := fmt.Sprintf(`<%s>[^<]*</%s>`, regexp.QuoteMeta(field), regexp.QuoteMeta(field))
		elementRe := regexp.MustCompile(elementPattern)
		result = elementRe.ReplaceAllString(result, fmt.Sprintf(`<%s>[TIMESTAMP]</%s>`, field, field))

		// Pattern for XML attributes: created_at="2025-08-16T09:00:00Z"
		attrPattern := fmt.Sprintf(`%s="[^"]*"`, regexp.QuoteMeta(field))
		attrRe := regexp.MustCompile(attrPattern)
		result = attrRe.ReplaceAllString(result, fmt.Sprintf(`%s="[TIMESTAMP]"`, field))

		// Pattern for XML attributes with single quotes: created_at='2025-08-16T09:00:00Z'
		attrSinglePattern := fmt.Sprintf(`%s='[^']*'`, regexp.QuoteMeta(field))
		attrSingleRe := regexp.MustCompile(attrSinglePattern)
		result = attrSingleRe.ReplaceAllString(result, fmt.Sprintf(`%s='[TIMESTAMP]'`, field))
	}

	// Also normalize common timestamp patterns (ISO 8601)
	iso8601Pattern := `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})`
	iso8601Re := regexp.MustCompile(iso8601Pattern)
	result = iso8601Re.ReplaceAllString(result, "[TIMESTAMP]")

	return result
}

// sortXMLAttributes sorts XML attributes alphabetically for consistent output
func (n *XMLNormalizer) sortXMLAttributes(xmlData string) string {
	// This is a simplified implementation that works for most cases
	// For complex XML with nested structures, a full XML parser would be better

	lines := strings.Split(xmlData, "\n")
	var result []string

	for _, line := range lines {
		result = append(result, n.sortAttributesInLine(line))
	}

	return strings.Join(result, "\n")
}

// sortAttributesInLine sorts attributes in a single XML line
func (n *XMLNormalizer) sortAttributesInLine(line string) string {
	// Find XML opening tags with attributes
	tagPattern := regexp.MustCompile(`(<[a-zA-Z][a-zA-Z0-9_-]*)\s+([^>]*?)(\/?>)`)

	return tagPattern.ReplaceAllStringFunc(line, func(match string) string {
		parts := tagPattern.FindStringSubmatch(match)
		if len(parts) != 4 {
			return match
		}

		tagStart := parts[1]   // e.g., "<epic"
		attributes := parts[2] // e.g., 'id="test" status="active"'
		tagEnd := parts[3]     // e.g., ">" or "/>"

		if strings.TrimSpace(attributes) == "" {
			return match
		}

		// Sort the attributes
		sortedAttrs := n.sortAttributes(attributes)

		return fmt.Sprintf("%s %s%s", tagStart, sortedAttrs, tagEnd)
	})
}

// sortAttributes sorts XML attributes alphabetically
func (n *XMLNormalizer) sortAttributes(attributes string) string {
	// Parse attributes using regex
	attrPattern := regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9_-]*)=("[^"]*"|'[^']*')`)
	matches := attrPattern.FindAllStringSubmatch(attributes, -1)

	if len(matches) == 0 {
		return attributes
	}

	// Sort attribute matches by name
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i][1] > matches[j][1] {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Rebuild attribute string
	var sortedAttrs []string
	for _, match := range matches {
		sortedAttrs = append(sortedAttrs, fmt.Sprintf("%s=%s", match[1], match[2]))
	}

	return strings.Join(sortedAttrs, " ")
}

// removeExtraWhitespace removes extra whitespace between XML elements
func (n *XMLNormalizer) removeExtraWhitespace(xmlData string) string {
	// Remove empty lines
	lines := strings.Split(xmlData, "\n")
	var nonEmptyLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}

	return strings.Join(nonEmptyLines, "\n")
}
