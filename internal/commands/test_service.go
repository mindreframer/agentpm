package commands

import (
	"fmt"
	"time"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/tests"
)

type TestRequest struct {
	TestID             string
	FailureReason      string
	CancellationReason string
	ConfigPath         string
	EpicFile           string
	Time               string
	Format             string
}

type TestResult struct {
	Result *tests.TestOperation
	Error  *TestError
}

type TestError struct {
	Type    string
	TestID  string
	Message string
}

func StartTestService(request TestRequest) (*TestResult, error) {
	if request.TestID == "" {
		return nil, fmt.Errorf("start-test requires exactly one argument: test-id")
	}

	// Load configuration and determine epic file
	epicFile, err := getEpicFileFromRequest(request)
	if err != nil {
		return nil, err
	}

	// Parse timestamp if provided
	var timestamp *time.Time
	if request.Time != "" {
		t, err := time.Parse(time.RFC3339, request.Time)
		if err != nil {
			return nil, fmt.Errorf("invalid time format: %s (expected ISO8601/RFC3339)", request.Time)
		}
		timestamp = &t
	}

	// Create test service
	service := tests.NewTestService(tests.ServiceConfig{
		UseMemory: false,
	})

	// Execute operation
	result, err := service.StartTest(epicFile, request.TestID, timestamp)
	if err != nil {
		if testErr, ok := err.(*tests.TestError); ok {
			return &TestResult{
				Error: &TestError{
					Type:    string(testErr.Type),
					TestID:  testErr.TestID,
					Message: testErr.Message,
				},
			}, nil
		}
		return nil, err
	}

	return &TestResult{
		Result: result,
	}, nil
}

func PassTestService(request TestRequest) (*TestResult, error) {
	if request.TestID == "" {
		return nil, fmt.Errorf("pass-test requires exactly one argument: test-id")
	}

	// Load configuration and determine epic file
	epicFile, err := getEpicFileFromRequest(request)
	if err != nil {
		return nil, err
	}

	// Parse timestamp if provided
	var timestamp *time.Time
	if request.Time != "" {
		t, err := time.Parse(time.RFC3339, request.Time)
		if err != nil {
			return nil, fmt.Errorf("invalid time format: %s (expected ISO8601/RFC3339)", request.Time)
		}
		timestamp = &t
	}

	// Create test service
	service := tests.NewTestService(tests.ServiceConfig{
		UseMemory: false,
	})

	// Execute operation
	result, err := service.PassTest(epicFile, request.TestID, timestamp)
	if err != nil {
		if testErr, ok := err.(*tests.TestError); ok {
			return &TestResult{
				Error: &TestError{
					Type:    string(testErr.Type),
					TestID:  testErr.TestID,
					Message: testErr.Message,
				},
			}, nil
		}
		return nil, err
	}

	return &TestResult{
		Result: result,
	}, nil
}

func FailTestService(request TestRequest) (*TestResult, error) {
	if request.TestID == "" || request.FailureReason == "" {
		return nil, fmt.Errorf("fail-test requires exactly two arguments: test-id \"failure-reason\"")
	}

	// Load configuration and determine epic file
	epicFile, err := getEpicFileFromRequest(request)
	if err != nil {
		return nil, err
	}

	// Parse timestamp if provided
	var timestamp *time.Time
	if request.Time != "" {
		t, err := time.Parse(time.RFC3339, request.Time)
		if err != nil {
			return nil, fmt.Errorf("invalid time format: %s (expected ISO8601/RFC3339)", request.Time)
		}
		timestamp = &t
	}

	// Create test service
	service := tests.NewTestService(tests.ServiceConfig{
		UseMemory: false,
	})

	// Execute operation
	result, err := service.FailTest(epicFile, request.TestID, request.FailureReason, timestamp)
	if err != nil {
		if testErr, ok := err.(*tests.TestError); ok {
			return &TestResult{
				Error: &TestError{
					Type:    string(testErr.Type),
					TestID:  testErr.TestID,
					Message: testErr.Message,
				},
			}, nil
		}
		return nil, err
	}

	return &TestResult{
		Result: result,
	}, nil
}

func CancelTestService(request TestRequest) (*TestResult, error) {
	if request.TestID == "" || request.CancellationReason == "" {
		return nil, fmt.Errorf("cancel-test requires exactly two arguments: test-id \"cancellation-reason\"")
	}

	// Load configuration and determine epic file
	epicFile, err := getEpicFileFromRequest(request)
	if err != nil {
		return nil, err
	}

	// Parse timestamp if provided
	var timestamp *time.Time
	if request.Time != "" {
		t, err := time.Parse(time.RFC3339, request.Time)
		if err != nil {
			return nil, fmt.Errorf("invalid time format: %s (expected ISO8601/RFC3339)", request.Time)
		}
		timestamp = &t
	}

	// Create test service
	service := tests.NewTestService(tests.ServiceConfig{
		UseMemory: false,
	})

	// Execute operation
	result, err := service.CancelTest(epicFile, request.TestID, request.CancellationReason, timestamp)
	if err != nil {
		if testErr, ok := err.(*tests.TestError); ok {
			return &TestResult{
				Error: &TestError{
					Type:    string(testErr.Type),
					TestID:  testErr.TestID,
					Message: testErr.Message,
				},
			}, nil
		}
		return nil, err
	}

	return &TestResult{
		Result: result,
	}, nil
}

func getEpicFileFromRequest(request TestRequest) (string, error) {
	// Check if file is provided directly
	epicFile := request.EpicFile
	if epicFile != "" {
		return epicFile, nil
	}

	// Load configuration
	configPath := request.ConfigPath
	if configPath == "" {
		configPath = "./.agentpm.json"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to load configuration: %w", err)
	}

	// Use epic file from config
	if cfg.CurrentEpic == "" {
		return "", fmt.Errorf("no epic file specified. Use --file flag or run 'agentpm init' first")
	}

	return cfg.CurrentEpic, nil
}
