package commands

import (
	"testing"
)

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
