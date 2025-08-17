package commands

import (
	"testing"
)

// Basic compilation tests for all services
func TestStartPhaseService_Compiles(t *testing.T) {
	request := StartPhaseRequest{
		PhaseID:  "test-phase",
		EpicFile: "/dev/null", // Will fail, but ensures service compiles
	}

	_, err := StartPhaseService(request)
	// We expect an error since the file doesn't exist
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestStartTaskService_Compiles(t *testing.T) {
	request := StartTaskRequest{
		TaskID:   "test-task",
		EpicFile: "/dev/null", // Will fail, but ensures service compiles
	}

	_, err := StartTaskService(request)
	// We expect an error since the file doesn't exist
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestDoneEpicService_Compiles(t *testing.T) {
	request := DoneEpicRequest{
		EpicFile: "/dev/null", // Will fail, but ensures service compiles
	}

	_, err := DoneEpicService(request)
	// We expect an error since the file doesn't exist
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestDonePhaseService_Compiles(t *testing.T) {
	request := DonePhaseRequest{
		PhaseID:  "test-phase",
		EpicFile: "/dev/null", // Will fail, but ensures service compiles
	}

	_, err := DonePhaseService(request)
	// We expect an error since the file doesn't exist
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestDoneTaskService_Compiles(t *testing.T) {
	request := DoneTaskRequest{
		TaskID:   "test-task",
		EpicFile: "/dev/null", // Will fail, but ensures service compiles
	}

	_, err := DoneTaskService(request)
	// We expect an error since the file doesn't exist
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestCancelTaskService_Compiles(t *testing.T) {
	request := CancelTaskRequest{
		TaskID:   "test-task",
		EpicFile: "/dev/null", // Will fail, but ensures service compiles
	}

	_, err := CancelTaskService(request)
	// We expect an error since the file doesn't exist
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestTestServices_Compile(t *testing.T) {
	testRequest := TestRequest{
		TestID:   "test-id",
		EpicFile: "/dev/null", // Will fail, but ensures service compiles
	}

	// Test StartTestService
	_, err := StartTestService(testRequest)
	if err == nil {
		t.Error("expected error for non-existent file")
	}

	// Test PassTestService
	_, err = PassTestService(testRequest)
	if err == nil {
		t.Error("expected error for non-existent file")
	}

	// Test FailTestService
	testRequest.FailureReason = "test failure"
	_, err = FailTestService(testRequest)
	if err == nil {
		t.Error("expected error for non-existent file")
	}

	// Test CancelTestService
	testRequest.CancellationReason = "test cancellation"
	_, err = CancelTestService(testRequest)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}
