package commands

import (
	"testing"
)

func TestDetectEntityType_Valid(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected EntityType
		wantErr  bool
	}{
		{
			name:     "phase_id_simple",
			id:       "3A",
			expected: EntityTypePhase,
			wantErr:  false,
		},
		{
			name:     "phase_id_complex",
			id:       "1B",
			expected: EntityTypePhase,
			wantErr:  false,
		},
		{
			name:     "task_id_simple",
			id:       "3A_1",
			expected: EntityTypeTask,
			wantErr:  false,
		},
		{
			name:     "task_id_complex",
			id:       "1B_2",
			expected: EntityTypeTask,
			wantErr:  false,
		},
		{
			name:     "test_id_simple",
			id:       "3A_T1",
			expected: EntityTypeTest,
			wantErr:  false,
		},
		{
			name:     "test_id_complex",
			id:       "1B_T2",
			expected: EntityTypeTest,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectEntityType(tt.id)

			if (result.Error != nil) != tt.wantErr {
				t.Errorf("DetectEntityType() error = %v, wantErr %v", result.Error, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result.EntityID == nil {
					t.Error("expected EntityID to be set")
					return
				}

				if result.EntityID.Type != tt.expected {
					t.Errorf("expected entity type %v, got %v", tt.expected, result.EntityID.Type)
				}

				if result.EntityID.ID != tt.id {
					t.Errorf("expected ID %v, got %v", tt.id, result.EntityID.ID)
				}

				if result.IsAmbiguous {
					t.Error("expected result to not be ambiguous")
				}
			}
		})
	}
}

func TestDetectEntityType_Invalid(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{
			name: "empty_id",
			id:   "",
		},
		{
			name: "invalid_pattern",
			id:   "invalid-id",
		},
		{
			name: "numbers_only",
			id:   "123",
		},
		{
			name: "letters_only",
			id:   "ABC",
		},
		{
			name: "wrong_separator",
			id:   "3A-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectEntityType(tt.id)

			if result.Error == nil {
				t.Error("expected error for invalid ID")
				return
			}

			if result.EntityID != nil {
				t.Error("expected EntityID to be nil for invalid ID")
			}
		})
	}
}

func TestDetectEntityType_Ambiguous(t *testing.T) {
	// Note: In the current implementation, the patterns are designed to be unambiguous
	// But we test the ambiguity detection mechanism in case patterns change
	tests := []struct {
		name            string
		id              string
		expectAmbiguous bool
	}{
		{
			name:            "task_vs_test_clear",
			id:              "3A_T1", // Clearly a test
			expectAmbiguous: false,
		},
		{
			name:            "task_clear",
			id:              "3A_1", // Clearly a task
			expectAmbiguous: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectEntityType(tt.id)

			if result.IsAmbiguous != tt.expectAmbiguous {
				t.Errorf("expected ambiguity %v, got %v", tt.expectAmbiguous, result.IsAmbiguous)
			}

			if tt.expectAmbiguous && len(result.Suggestions) == 0 {
				t.Error("expected suggestions for ambiguous ID")
			}
		})
	}
}

func TestValidateSubcommandArgs(t *testing.T) {
	tests := []struct {
		name         string
		subcommand   string
		args         []string
		requiredArgs int
		wantErr      bool
	}{
		{
			name:         "correct_single_arg",
			subcommand:   "start",
			args:         []string{"3A"},
			requiredArgs: 1,
			wantErr:      false,
		},
		{
			name:         "correct_two_args",
			subcommand:   "fail",
			args:         []string{"3A_T1", "reason"},
			requiredArgs: 2,
			wantErr:      false,
		},
		{
			name:         "missing_args",
			subcommand:   "start",
			args:         []string{},
			requiredArgs: 1,
			wantErr:      true,
		},
		{
			name:         "too_many_args",
			subcommand:   "start",
			args:         []string{"3A", "extra"},
			requiredArgs: 1,
			wantErr:      true,
		},
		{
			name:         "missing_second_arg",
			subcommand:   "fail",
			args:         []string{"3A_T1"},
			requiredArgs: 2,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSubcommandArgs(tt.subcommand, tt.args, tt.requiredArgs)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSubcommandArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEntityTypeString(t *testing.T) {
	tests := []struct {
		entityType EntityType
		expected   string
	}{
		{EntityTypePhase, "phase"},
		{EntityTypeTask, "task"},
		{EntityTypeTest, "test"},
		{EntityTypeEpic, "epic"},
	}

	for _, tt := range tests {
		t.Run(string(tt.entityType), func(t *testing.T) {
			result := tt.entityType.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFormatEntityTypes(t *testing.T) {
	tests := []struct {
		name     string
		entities []EntityID
		expected string
	}{
		{
			name: "single_entity",
			entities: []EntityID{
				{ID: "3A", Type: EntityTypePhase},
			},
			expected: "phase",
		},
		{
			name: "two_entities",
			entities: []EntityID{
				{ID: "3A", Type: EntityTypePhase},
				{ID: "3A_1", Type: EntityTypeTask},
			},
			expected: "phase or task",
		},
		{
			name: "three_entities",
			entities: []EntityID{
				{ID: "3A", Type: EntityTypePhase},
				{ID: "3A_1", Type: EntityTypeTask},
				{ID: "3A_T1", Type: EntityTypeTest},
			},
			expected: "phase or task or test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatEntityTypes(tt.entities)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGlobalFlags(t *testing.T) {
	flags := GlobalFlags()

	// Check that we have the expected number of flags
	expectedFlags := 4 // file, config, time, format
	if len(flags) != expectedFlags {
		t.Errorf("expected %d global flags, got %d", expectedFlags, len(flags))
	}

	// Check for specific flags
	flagNames := make(map[string]bool)
	for _, flag := range flags {
		flagNames[flag.Names()[0]] = true
	}

	requiredFlags := []string{"file", "config", "time", "format"}
	for _, requiredFlag := range requiredFlags {
		if !flagNames[requiredFlag] {
			t.Errorf("missing required global flag: %s", requiredFlag)
		}
	}
}

// Test router context extraction
func TestRouterContext(t *testing.T) {
	// This test validates the RouterContext struct and its fields
	ctx := RouterContext{
		ConfigPath: "/test/config.json",
		EpicFile:   "/test/epic.xml",
		Format:     "json",
		Time:       "2025-08-16T15:30:00Z",
	}

	if ctx.ConfigPath != "/test/config.json" {
		t.Errorf("expected ConfigPath to be set")
	}

	if ctx.EpicFile != "/test/epic.xml" {
		t.Errorf("expected EpicFile to be set")
	}

	if ctx.Format != "json" {
		t.Errorf("expected Format to be set")
	}

	if ctx.Time != "2025-08-16T15:30:00Z" {
		t.Errorf("expected Time to be set")
	}
}
