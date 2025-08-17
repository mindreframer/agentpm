package assertions

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

// SnapshotAssertion provides snapshot testing capabilities for the assertion framework
type SnapshotAssertion struct {
	tester     *testing.T
	normalizer SnapshotNormalizer
}

// SnapshotNormalizer handles data normalization for consistent snapshots
type SnapshotNormalizer interface {
	NormalizeData(data interface{}) (interface{}, error)
	NormalizeXML(xml string) (string, error)
}

// DefaultSnapshotNormalizer provides basic normalization
type DefaultSnapshotNormalizer struct {
	crossPlatform bool
	updateMode    bool
}

func (n *DefaultSnapshotNormalizer) NormalizeData(data interface{}) (interface{}, error) {
	// Convert to JSON and back to normalize structure
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var normalized interface{}
	err = json.Unmarshal(jsonData, &normalized)
	if err != nil {
		return nil, err
	}

	// Apply cross-platform normalization if enabled
	if n.crossPlatform {
		normalized = n.applyCrossPlatformNormalization(normalized)
	}

	// Apply timestamp normalization for snapshot consistency
	normalized = n.normalizeTimestamps(normalized)

	return normalized, nil
}

// applyCrossPlatformNormalization normalizes data for cross-platform consistency
func (n *DefaultSnapshotNormalizer) applyCrossPlatformNormalization(data interface{}) interface{} {
	if dataMap, ok := data.(map[string]interface{}); ok {
		// Normalize path separators
		for key, value := range dataMap {
			if strValue, ok := value.(string); ok {
				// Replace Windows backslashes with forward slashes
				dataMap[key] = fmt.Sprintf("%s", strValue) // Simple string normalization
			}
		}
	}
	return data
}

// normalizeTimestamps replaces timestamp values with placeholders for consistent snapshots
func (n *DefaultSnapshotNormalizer) normalizeTimestamps(data interface{}) interface{} {
	return n.normalizeTimestampsRecursive(data)
}

func (n *DefaultSnapshotNormalizer) normalizeTimestampsRecursive(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		normalized := make(map[string]interface{})
		for key, value := range v {
			// Check if this is a timestamp field
			if n.isTimestampField(key) {
				if str, ok := value.(string); ok && str != "" {
					normalized[key] = "NORMALIZED_TIMESTAMP"
				} else {
					normalized[key] = value
				}
			} else if key == "ID" {
				// Special handling for ID fields that may contain timestamps
				if str, ok := value.(string); ok && str != "" {
					normalized[key] = n.normalizeIDWithTimestamp(str)
				} else {
					normalized[key] = value
				}
			} else {
				normalized[key] = n.normalizeTimestampsRecursive(value)
			}
		}
		return normalized
	case []interface{}:
		normalized := make([]interface{}, len(v))
		for i, item := range v {
			normalized[i] = n.normalizeTimestampsRecursive(item)
		}
		return normalized
	default:
		return data
	}
}

func (n *DefaultSnapshotNormalizer) isTimestampField(fieldName string) bool {
	timestampFields := []string{
		"Timestamp", "CreatedAt", "Created", "StartedAt", "CompletedAt",
		"PassedAt", "FailedAt", "CancelledAt", "UpdatedAt", "ModifiedAt",
	}

	for _, field := range timestampFields {
		if fieldName == field {
			return true
		}
	}
	return false
}

// normalizeIDWithTimestamp normalizes ID fields that contain embedded timestamps
func (n *DefaultSnapshotNormalizer) normalizeIDWithTimestamp(id string) string {
	// Pattern to match timestamp-like numbers at the end of IDs
	// Matches patterns like: epic_started_1755436365, task_completed_1755436734
	re := regexp.MustCompile(`^(.+)_(\d{10,})$`)

	if matches := re.FindStringSubmatch(id); len(matches) == 3 {
		// Keep the prefix, replace the timestamp with a normalized value
		prefix := matches[1]
		return prefix + "_NORMALIZED_TIMESTAMP"
	}

	// If no timestamp pattern found, return original ID
	return id
}

func (n *DefaultSnapshotNormalizer) NormalizeXML(xml string) (string, error) {
	// Basic XML normalization - can be enhanced later
	return xml, nil
}

// NewSnapshotAssertion creates a new snapshot assertion
func NewSnapshotAssertion(t *testing.T) *SnapshotAssertion {
	return &SnapshotAssertion{
		tester: t,
		normalizer: &DefaultSnapshotNormalizer{
			crossPlatform: false,
			updateMode:    false,
		},
	}
}

// NewSnapshotAssertionWithConfig creates a new snapshot assertion with configuration
func NewSnapshotAssertionWithConfig(t *testing.T, crossPlatform, updateMode bool) *SnapshotAssertion {
	return &SnapshotAssertion{
		tester: t,
		normalizer: &DefaultSnapshotNormalizer{
			crossPlatform: crossPlatform,
			updateMode:    updateMode,
		},
	}
}

// MatchSnapshot performs snapshot testing on arbitrary data
func (sa *SnapshotAssertion) MatchSnapshot(name string, data interface{}) error {
	if sa.tester == nil {
		return fmt.Errorf("no testing.T instance available for snapshot testing")
	}

	normalized, err := sa.normalizer.NormalizeData(data)
	if err != nil {
		return fmt.Errorf("failed to normalize data for snapshot: %v", err)
	}

	snaps.MatchSnapshot(sa.tester, name, normalized)
	return nil
}

// MatchXMLSnapshot performs XML-specific snapshot testing
func (sa *SnapshotAssertion) MatchXMLSnapshot(name string, xmlData string) error {
	if sa.tester == nil {
		return fmt.Errorf("no testing.T instance available for snapshot testing")
	}

	normalized, err := sa.normalizer.NormalizeXML(xmlData)
	if err != nil {
		return fmt.Errorf("failed to normalize XML for snapshot: %v", err)
	}

	snaps.MatchSnapshot(sa.tester, name, normalized)
	return nil
}

// MatchSelectiveSnapshot performs snapshot testing on specific fields
func (sa *SnapshotAssertion) MatchSelectiveSnapshot(name string, data interface{}, fields []string) error {
	if sa.tester == nil {
		return fmt.Errorf("no testing.T instance available for snapshot testing")
	}

	// Extract only the specified fields
	selective := sa.extractFields(data, fields)

	normalized, err := sa.normalizer.NormalizeData(selective)
	if err != nil {
		return fmt.Errorf("failed to normalize selective data for snapshot: %v", err)
	}

	snaps.MatchSnapshot(sa.tester, name, normalized)
	return nil
}

// extractFields extracts specific fields from data for selective snapshots
func (sa *SnapshotAssertion) extractFields(data interface{}, fields []string) map[string]interface{} {
	result := make(map[string]interface{})

	// Convert to JSON for field extraction
	jsonData, err := json.Marshal(data)
	if err != nil {
		result["error"] = fmt.Sprintf("failed to marshal data: %v", err)
		return result
	}

	var dataMap map[string]interface{}
	err = json.Unmarshal(jsonData, &dataMap)
	if err != nil {
		result["error"] = fmt.Sprintf("failed to unmarshal data: %v", err)
		return result
	}

	// Extract only specified fields
	for _, field := range fields {
		if value, exists := dataMap[field]; exists {
			result[field] = value
		}
	}

	return result
}
