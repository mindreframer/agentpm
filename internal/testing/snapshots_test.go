package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSnapshotTester(t *testing.T) {
	tester := NewSnapshotTester()
	assert.NotNil(t, tester)

	// Check that it implements the interface
	var _ SnapshotTester = tester
}

func TestNewSnapshotTesterWithConfig(t *testing.T) {
	config := SnapshotConfig{
		NormalizeXML:     false,
		SortAttributes:   false,
		RemoveTimestamps: false,
		TimestampFields:  []string{"custom_field"},
	}

	tester := NewSnapshotTesterWithConfig(config)
	assert.NotNil(t, tester)

	defaultTester := tester.(*DefaultSnapshotTester)
	assert.Equal(t, config, defaultTester.config)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.True(t, config.NormalizeXML)
	assert.True(t, config.SortAttributes)
	assert.True(t, config.RemoveTimestamps)
	assert.Contains(t, config.TimestampFields, "created_at")
	assert.Contains(t, config.TimestampFields, "started_at")
	assert.Contains(t, config.TimestampFields, "completed_at")
}

func TestNormalizeXML_EmptyInput(t *testing.T) {
	tester := NewSnapshotTester().(*DefaultSnapshotTester)

	result, err := tester.normalizeXML("")
	assert.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestNormalizeXML_NonXMLInput(t *testing.T) {
	tester := NewSnapshotTester().(*DefaultSnapshotTester)

	nonXML := "This is just plain text"
	result, err := tester.normalizeXML(nonXML)
	assert.NoError(t, err)
	assert.Equal(t, nonXML, result)
}

func TestNormalizeXML_ValidXML(t *testing.T) {
	tester := NewSnapshotTester().(*DefaultSnapshotTester)

	input := `<epic id="test-epic" status="active"><name>Test Epic</name></epic>`
	result, err := tester.normalizeXML(input)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	// Should be properly formatted XML
	assert.Contains(t, result, "<epic")
	assert.Contains(t, result, "<name>Test Epic</name>")
}

func TestNormalizeXML_WithTimestamps(t *testing.T) {
	tester := NewSnapshotTester().(*DefaultSnapshotTester)

	input := `<epic created_at="2025-08-16T09:00:00Z"><started_at>2025-08-16T10:00:00Z</started_at></epic>`
	result, err := tester.normalizeXML(input)

	assert.NoError(t, err)
	assert.Contains(t, result, `created_at="[TIMESTAMP]"`)
	assert.Contains(t, result, `<started_at>[TIMESTAMP]</started_at>`)
	assert.NotContains(t, result, "2025-08-16T09:00:00Z")
	assert.NotContains(t, result, "2025-08-16T10:00:00Z")
}

func TestNormalizeXML_DisableTimestampNormalization(t *testing.T) {
	config := DefaultConfig()
	config.RemoveTimestamps = false
	tester := NewSnapshotTesterWithConfig(config).(*DefaultSnapshotTester)

	input := `<epic created_at="2025-08-16T09:00:00Z"><name>Test</name></epic>`
	result, err := tester.normalizeXML(input)

	assert.NoError(t, err)
	assert.Contains(t, result, "2025-08-16T09:00:00Z")
	assert.NotContains(t, result, "[TIMESTAMP]")
}

// Note: Removed individual method tests since we now use XMLNormalizer
// The functionality is tested in xml_normalize_test.go

func TestMatchXMLSnapshot_Integration(t *testing.T) {
	tester := NewSnapshotTester()

	// Test with typical XML output from the application
	xmlOutput := `<status epic="test-epic">
		<name>Test Epic</name>
		<status>active</status>
		<completion_percentage>30</completion_percentage>
		<current_phase>P2</current_phase>
	</status>`

	// This should not panic or error
	// Note: In actual usage, this would create/compare snapshots
	// For unit tests, we just verify the normalization works
	require.NotPanics(t, func() {
		tester.MatchXMLSnapshot(t, xmlOutput, "test_xml_output")
	})
}

func TestMatchSnapshot_Integration(t *testing.T) {
	tester := NewSnapshotTester()

	// Test with structured data
	data := map[string]interface{}{
		"epic":   "test-epic",
		"status": "active",
		"phases": []string{"P1", "P2", "P3"},
	}

	// This should not panic or error
	require.NotPanics(t, func() {
		tester.MatchSnapshot(t, data, "test_data_output")
	})
}

// Test error handling for malformed XML
func TestNormalizeXML_MalformedXML(t *testing.T) {
	tester := NewSnapshotTester().(*DefaultSnapshotTester)

	// Malformed XML should be returned as-is (for error cases)
	malformedXML := `<epic><name>Test</epic>` // Missing closing tag
	result, err := tester.normalizeXML(malformedXML)

	assert.NoError(t, err) // Should not error, just return as-is
	assert.Equal(t, malformedXML, result)
}

// Test configuration edge cases
func TestSnapshotConfig_CustomTimestampFields(t *testing.T) {
	config := SnapshotConfig{
		NormalizeXML:     true,
		RemoveTimestamps: true,
		TimestampFields:  []string{"custom_timestamp", "another_time"},
	}

	tester := NewSnapshotTesterWithConfig(config).(*DefaultSnapshotTester)

	input := `<epic custom_timestamp="2025-08-16T09:00:00Z"><another_time>2025-08-16T10:00:00Z</another_time></epic>`
	result, err := tester.normalizeXML(input)

	assert.NoError(t, err)
	assert.Contains(t, result, `custom_timestamp="[TIMESTAMP]"`)
	assert.Contains(t, result, `<another_time>[TIMESTAMP]</another_time>`)
}

// Benchmark test for performance validation
func BenchmarkNormalizeXML(b *testing.B) {
	tester := NewSnapshotTester().(*DefaultSnapshotTester)

	xmlData := `<epic id="test-epic" status="active" created_at="2025-08-16T09:00:00Z">
		<name>Test Epic</name>
		<phases>
			<phase id="p1" status="completed" started_at="2025-08-16T09:00:00Z" completed_at="2025-08-16T10:00:00Z">
				<name>Setup Phase</name>
			</phase>
			<phase id="p2" status="active" started_at="2025-08-16T10:00:00Z">
				<name>Implementation Phase</name>
			</phase>
		</phases>
	</epic>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tester.normalizeXML(xmlData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test that snapshot framework handles realistic XML output
func TestRealisticXMLOutput(t *testing.T) {
	tester := NewSnapshotTester()

	// Realistic XML output from status command
	xmlOutput := `<status epic="epic-1">
  <name>Sample Epic</name>
  <status>active</status>
  <completion_percentage>45</completion_percentage>
  <completed_phases>2</completed_phases>
  <total_phases>4</total_phases>
  <passing_tests>3</passing_tests>
  <failing_tests>1</failing_tests>
  <current_phase>implementation</current_phase>
  <current_task>task-5</current_task>
</status>`

	// Should handle without errors
	require.NotPanics(t, func() {
		tester.MatchXMLSnapshot(t, xmlOutput, "realistic_status_output")
	})
}
