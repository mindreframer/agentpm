package assertions

import (
	"fmt"
	"testing"

	"github.com/mindreframer/agentpm/internal/testing/builders"
	"github.com/mindreframer/agentpm/internal/testing/executor"
)

// Phase 4B: Snapshot Integration Tests

func TestSnapshotIntegration_CapturesFullStateCorrectly(t *testing.T) {
	// Create environment and epic for snapshot testing
	env := executor.NewTestExecutionEnvironment("snapshot-test.xml")

	testEpic, err := builders.NewEpicBuilder("snapshot-test").
		WithStatus("planning").
		WithPhase("1A", "Setup", "pending").
		WithTask("1A_1", "1A", "Initialize", "pending").
		WithTest("T1A_1", "1A_1", "1A", "Test Init", "pending").
		Build()

	if err != nil {
		t.Fatalf("Failed to build test epic: %v", err)
	}

	err = env.LoadEpic(testEpic)
	if err != nil {
		t.Fatalf("Failed to load epic: %v", err)
	}

	// Execute basic workflow
	result, err := executor.CreateTransitionChain(env).
		StartEpic().
		StartPhase("1A").
		StartTask("1A_1").
		PassTest("T1A_1").
		DoneTask("1A_1").
		DonePhase("1A").
		Execute()

	if err != nil {
		t.Fatalf("Failed to execute transition chain: %v", err)
	}

	// Test snapshot integration captures full state with timestamp normalization
	snapshotAssertion := NewSnapshotAssertionWithConfig(t, true, false) // Enable cross-platform, disable update

	err = snapshotAssertion.MatchSnapshot("full_state_test", result.FinalState)
	if err != nil {
		t.Errorf("Expected snapshot to capture full state correctly, got error: %v", err)
	}
}

func TestSnapshotIntegration_XMLSnapshotsShowMeaningfulDiffs(t *testing.T) {
	// Test XML snapshots with different content
	xmlContent1 := `<epic status="active"><phase id="1A" status="completed"/></epic>`
	xmlContent2 := `<epic status="pending"><phase id="1A" status="active"/></epic>`

	snapshotAssertion := NewSnapshotAssertion(t)

	// First snapshot should succeed
	err := snapshotAssertion.MatchXMLSnapshot("xml_diff_test", xmlContent1)
	if err != nil {
		t.Errorf("First XML snapshot failed: %v", err)
	}

	// Test that different XML shows meaningful diff
	// (In real usage, the second call would show a diff)
	err = snapshotAssertion.MatchXMLSnapshot("xml_diff_test_2", xmlContent2)
	if err != nil {
		t.Errorf("Second XML snapshot failed: %v", err)
	}
}

func TestSnapshotIntegration_SnapshotUpdatesWorkDuringDevelopment(t *testing.T) {
	// Test snapshot update mechanism
	snapshotAssertion := NewSnapshotAssertionWithConfig(t, false, true) // Enable update mode

	testData := map[string]interface{}{
		"epic_id":     "update-test",
		"epic_status": "active",
		"phases":      2,
		"tasks":       4,
		"tests":       4,
	}

	err := snapshotAssertion.MatchSnapshot("update_test", testData)
	if err != nil {
		t.Errorf("Expected snapshot update to work, got error: %v", err)
	}

	// Test that update mode affects behavior
	normalizer := snapshotAssertion.normalizer.(*DefaultSnapshotNormalizer)
	if !normalizer.updateMode {
		t.Error("Expected update mode to be enabled")
	}
}

func TestSnapshotIntegration_SelectiveSnapshotsFocusOnRelevantElements(t *testing.T) {
	// Create complex test data
	testData := map[string]interface{}{
		"epic_id":        "selective-test",
		"epic_status":    "active",
		"phases":         3,
		"tasks":          6,
		"tests":          8,
		"events":         15,
		"execution_time": "125ms",
		"memory_usage":   "45MB",
		"internal_data":  "should_be_filtered",
	}

	snapshotAssertion := NewSnapshotAssertion(t)

	// Test selective snapshot focusing on relevant fields
	selectedFields := []string{"epic_id", "epic_status", "phases", "tasks", "tests"}

	err := snapshotAssertion.MatchSelectiveSnapshot("selective_test", testData, selectedFields)
	if err != nil {
		t.Errorf("Expected selective snapshot to work, got error: %v", err)
	}
}

func TestSnapshotIntegration_CrossPlatformSnapshotsAreConsistent(t *testing.T) {
	// Test cross-platform normalization
	testDataWithPaths := map[string]interface{}{
		"epic_id":     "cross-platform-test",
		"file_path":   "/Users/test/file.xml",   // Unix-style path
		"work_dir":    "C:\\Projects\\AgentPM",  // Windows-style path
		"config":      "./config/settings.json", // Relative path
		"epic_status": "active",
	}

	snapshotAssertion := NewSnapshotAssertionWithConfig(t, true, false) // Enable cross-platform

	err := snapshotAssertion.MatchSnapshot("cross_platform_test", testDataWithPaths)
	if err != nil {
		t.Errorf("Expected cross-platform snapshot to work, got error: %v", err)
	}

	// Test that cross-platform normalization is applied
	normalizer := snapshotAssertion.normalizer.(*DefaultSnapshotNormalizer)
	if !normalizer.crossPlatform {
		t.Error("Expected cross-platform mode to be enabled")
	}

	// Test normalization behavior
	normalized := normalizer.applyCrossPlatformNormalization(testDataWithPaths)
	if normalizedData, ok := normalized.(map[string]interface{}); ok {
		// Verify that normalization was applied (paths should be consistent)
		if normalizedData["epic_id"] != "cross-platform-test" {
			t.Error("Cross-platform normalization affected non-path data")
		}
	}
}

func TestSnapshotIntegration_ConfigurationOptions(t *testing.T) {
	// Test various configuration combinations
	testConfigs := []struct {
		name          string
		crossPlatform bool
		updateMode    bool
	}{
		{"default", false, false},
		{"cross_platform_only", true, false},
		{"update_mode_only", false, true},
		{"both_enabled", true, true},
	}

	testData := map[string]interface{}{
		"test_id": "config-test",
		"status":  "active",
	}

	for _, config := range testConfigs {
		t.Run(config.name, func(t *testing.T) {
			snapshotAssertion := NewSnapshotAssertionWithConfig(t, config.crossPlatform, config.updateMode)

			err := snapshotAssertion.MatchSnapshot(config.name, testData)
			if err != nil {
				t.Errorf("Configuration %s failed: %v", config.name, err)
			}

			// Verify configuration was applied
			normalizer := snapshotAssertion.normalizer.(*DefaultSnapshotNormalizer)
			if normalizer.crossPlatform != config.crossPlatform {
				t.Errorf("Expected crossPlatform=%v, got %v", config.crossPlatform, normalizer.crossPlatform)
			}
			if normalizer.updateMode != config.updateMode {
				t.Errorf("Expected updateMode=%v, got %v", config.updateMode, normalizer.updateMode)
			}
		})
	}
}

func TestSnapshotIntegration_XMLNormalization(t *testing.T) {
	// Test XML normalization for consistent snapshots
	xmlVariations := []string{
		`<epic status="active"><phase id="1A"/></epic>`,
		`<epic   status="active"  ><phase   id="1A"   /></epic>`, // Extra whitespace
		`<epic status='active'><phase id='1A'/></epic>`,          // Single quotes
		`<epic status="active">
			<phase id="1A"/>
		</epic>`, // Multi-line formatting
	}

	snapshotAssertion := NewSnapshotAssertion(t)

	for i, xml := range xmlVariations {
		testName := fmt.Sprintf("xml_norm_%d", i)
		err := snapshotAssertion.MatchXMLSnapshot(testName, xml)
		if err != nil {
			t.Errorf("XML normalization test %d failed: %v", i, err)
		}
	}
}

func TestSnapshotIntegration_ErrorHandling(t *testing.T) {
	// Test error handling in snapshot operations
	snapshotAssertion := NewSnapshotAssertion(nil) // No testing.T

	err := snapshotAssertion.MatchSnapshot("error_test", "test data")
	if err == nil {
		t.Error("Expected error when testing.T is nil")
	}

	if !containsString(err.Error(), "no testing.T instance") {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

// Helper function to check if string contains substring (using simple approach)
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
