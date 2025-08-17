package commands

import (
	"os"
	"testing"

	"github.com/mindreframer/agentpm/internal/lifecycle"
	"github.com/mindreframer/agentpm/internal/messages"
)

func TestStartEpicService_Basic(t *testing.T) {
	tests := []struct {
		name     string
		request  StartEpicRequest
		wantErr  bool
		validate func(*testing.T, *StartEpicResult, error)
	}{
		{
			name: "valid_request_with_file",
			request: StartEpicRequest{
				EpicFile: "testdata/epic-valid.xml",
				Format:   "text",
			},
			wantErr: false,
			validate: func(t *testing.T, result *StartEpicResult, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result == nil {
					t.Error("result should not be nil")
					return
				}
			},
		},
		{
			name: "missing_epic_file_and_config",
			request: StartEpicRequest{
				ConfigPath: "/nonexistent/config.json",
				Format:     "text",
			},
			wantErr: true,
			validate: func(t *testing.T, result *StartEpicResult, err error) {
				if err == nil {
					t.Error("expected error for missing config")
				}
			},
		},
		{
			name: "invalid_time_format",
			request: StartEpicRequest{
				EpicFile: "testdata/epic-valid.xml",
				Time:     "invalid-time",
				Format:   "text",
			},
			wantErr: true,
			validate: func(t *testing.T, result *StartEpicResult, err error) {
				if err == nil {
					t.Error("expected error for invalid time format")
				}
			},
		},
		{
			name: "valid_time_format",
			request: StartEpicRequest{
				EpicFile: "testdata/epic-valid.xml",
				Time:     "2025-08-16T15:30:00Z",
				Format:   "text",
			},
			wantErr: false,
			validate: func(t *testing.T, result *StartEpicResult, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StartEpicService(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("StartEpicService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.validate(t, result, err)
		})
	}
}

func TestStartEpicService_AlreadyStarted(t *testing.T) {
	request := StartEpicRequest{
		EpicFile: "testdata/epic-already-started.xml",
		Format:   "text",
	}

	result, err := StartEpicService(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("result should not be nil")
		return
	}

	if !result.IsAlreadyStarted {
		t.Error("expected IsAlreadyStarted to be true")
	}

	if result.Message == nil {
		t.Error("expected friendly message for already started epic")
	}
}

func TestStartEpicService_AlreadyCompleted(t *testing.T) {
	request := StartEpicRequest{
		EpicFile: "testdata/epic-completed.xml",
		Format:   "text",
	}

	result, err := StartEpicService(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("result should not be nil")
		return
	}

	if !result.IsAlreadyCompleted {
		t.Error("expected IsAlreadyCompleted to be true")
	}

	if result.Message == nil {
		t.Error("expected friendly message for already completed epic")
	}
}

func TestStartEpicService_LifecycleIntegration(t *testing.T) {
	// Create a temporary copy of the epic file to avoid test isolation issues
	originalContent := `<?xml version="1.0" encoding="UTF-8"?>
<epic id="8" name="Epic Name" status="planning" created_at="2025-08-16T09:00:00Z">
    <assignee>agent_claude</assignee>
    <description>Epic description</description>
    <phases>
        <phase id="1A" name="Setup" status="planning">
            <description>Initial setup phase</description>
        </phase>
        <phase id="1B" name="Development" status="planning">
            <description>Main development phase</description>
        </phase>
    </phases>
    <tasks>
        <task id="1A_1" phase_id="1A" name="Initialize Project" status="planning" assignee="agent_claude">
            <description>Set up the project structure</description>
        </task>
        <task id="1A_2" phase_id="1A" name="Configure Tools" status="planning">
            <description>Configure development tools</description>
        </task>
    </tasks>
    <tests>
        <test id="T1A_1" task_id="1A_1" name="Test Project Init" status="planning">
            <description>Test that project initializes correctly</description>
        </test>
    </tests>
    <events>
        <event id="E1" type="created" timestamp="2025-08-16T09:00:00Z">
            <data>Epic created by agent_claude</data>
        </event>
    </events>
</epic>`

	tempFile := "testdata/epic-lifecycle-temp.xml"
	err := os.WriteFile(tempFile, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile)

	// Test that the service correctly calls lifecycle service
	request := StartEpicRequest{
		EpicFile: tempFile,
		Time:     "2025-08-16T15:30:00Z",
		Format:   "json",
	}

	result, err := StartEpicService(request)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Error("result should not be nil")
		return
	}

	if result.Result == nil {
		t.Error("lifecycle result should not be nil")
		return
	}

	// Verify the result has expected lifecycle fields
	if result.Result.EpicID == "" {
		t.Error("expected EpicID to be set")
	}

	if result.Result.NewStatus != lifecycle.LifecycleStatusWIP {
		t.Errorf("expected new status to be WIP, got %v", result.Result.NewStatus)
	}

	if result.Result.StartedAt.IsZero() {
		t.Error("expected StartedAt to be set")
	}
}

func TestStartEpicService_MessageTypes(t *testing.T) {
	tests := []struct {
		name     string
		epicFile string
		validate func(*testing.T, *StartEpicResult)
	}{
		{
			name:     "already_started_message",
			epicFile: "testdata/epic-already-started.xml",
			validate: func(t *testing.T, result *StartEpicResult) {
				if result.Message == nil {
					t.Error("expected message to be set")
					return
				}
				if result.Message.Type != messages.MessageSuccess {
					t.Errorf("expected success message type, got %v", result.Message.Type)
				}
			},
		},
		{
			name:     "already_completed_message",
			epicFile: "testdata/epic-completed.xml",
			validate: func(t *testing.T, result *StartEpicResult) {
				if result.Message == nil {
					t.Error("expected message to be set")
					return
				}
				if result.Message.Type != messages.MessageSuccess {
					t.Errorf("expected success message type, got %v", result.Message.Type)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := StartEpicRequest{
				EpicFile: tt.epicFile,
				Format:   "text",
			}

			result, err := StartEpicService(request)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("result should not be nil")
				return
			}

			tt.validate(t, result)
		})
	}
}
