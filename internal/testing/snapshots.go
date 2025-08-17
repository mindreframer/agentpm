package testing

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

// SnapshotTester interface for XML output testing
type SnapshotTester interface {
	MatchSnapshot(t *testing.T, data interface{}, optionalName ...string)
	MatchXMLSnapshot(t *testing.T, xmlData string, optionalName ...string)
	UpdateSnapshots() error
}

// DefaultSnapshotTester implements SnapshotTester using go-snaps
type DefaultSnapshotTester struct {
	config     SnapshotConfig
	normalizer *XMLNormalizer
}

// SnapshotConfig holds configuration for snapshot testing
type SnapshotConfig struct {
	// NormalizeXML controls whether XML is normalized before comparison
	NormalizeXML bool
	// SortAttributes controls whether XML attributes are sorted
	SortAttributes bool
	// RemoveTimestamps controls whether timestamp fields are normalized
	RemoveTimestamps bool
	// TimestampFields contains field names to normalize (e.g., "created_at", "updated_at")
	TimestampFields []string
}

// DefaultConfig returns a default configuration for snapshot testing
func DefaultConfig() SnapshotConfig {
	return SnapshotConfig{
		NormalizeXML:     true,
		SortAttributes:   true,
		RemoveTimestamps: true,
		TimestampFields:  []string{"created_at", "updated_at", "started_at", "completed_at", "passed_at", "failed_at", "cancelled_at"},
	}
}

// NewSnapshotTester creates a new snapshot tester with default configuration
func NewSnapshotTester() SnapshotTester {
	config := DefaultConfig()
	return &DefaultSnapshotTester{
		config:     config,
		normalizer: newXMLNormalizerFromSnapshotConfig(config),
	}
}

// NewSnapshotTesterWithConfig creates a new snapshot tester with custom configuration
func NewSnapshotTesterWithConfig(config SnapshotConfig) SnapshotTester {
	return &DefaultSnapshotTester{
		config:     config,
		normalizer: newXMLNormalizerFromSnapshotConfig(config),
	}
}

// newXMLNormalizerFromSnapshotConfig creates an XMLNormalizer configured from SnapshotConfig
func newXMLNormalizerFromSnapshotConfig(snapConfig SnapshotConfig) *XMLNormalizer {
	normConfig := NormalizationConfig{
		IndentSize:          2,
		RemoveWhitespace:    true,
		SortAttributes:      snapConfig.SortAttributes,
		NormalizeTimestamps: snapConfig.RemoveTimestamps,
		TimestampFields:     snapConfig.TimestampFields,
		NormalizePaths:      true, // Enable path normalization for snapshots
		PathFields:          []string{"previous_epic", "new_epic", "epic_path", "message"},
	}
	return NewXMLNormalizerWithConfig(normConfig)
}

// MatchSnapshot matches any data against a snapshot
func (s *DefaultSnapshotTester) MatchSnapshot(t *testing.T, data interface{}, optionalName ...string) {
	if len(optionalName) > 0 {
		snaps.MatchSnapshot(t, optionalName[0], data)
	} else {
		snaps.MatchSnapshot(t, data)
	}
}

// MatchXMLSnapshot matches XML data against a snapshot with normalization
func (s *DefaultSnapshotTester) MatchXMLSnapshot(t *testing.T, xmlData string, optionalName ...string) {
	normalizedXML := xmlData

	if s.config.NormalizeXML {
		normalized, err := s.normalizeXML(xmlData)
		if err != nil {
			t.Fatalf("Failed to normalize XML for snapshot: %v", err)
		}
		normalizedXML = normalized
	}

	if len(optionalName) > 0 {
		snaps.MatchSnapshot(t, optionalName[0], normalizedXML)
	} else {
		snaps.MatchSnapshot(t, normalizedXML)
	}
}

// UpdateSnapshots updates all snapshots (delegates to go-snaps)
func (s *DefaultSnapshotTester) UpdateSnapshots() error {
	// Note: go-snaps handles snapshot updates via environment variable SNAPS_UPDATE=true
	// This method is provided for interface completeness
	return nil
}

// normalizeXML normalizes XML content for consistent snapshot comparisons
func (s *DefaultSnapshotTester) normalizeXML(xmlData string) (string, error) {
	// Use the XMLNormalizer for normalization
	return s.normalizer.NormalizeXML(xmlData)
}
