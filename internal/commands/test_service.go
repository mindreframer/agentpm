package commands

import (
	"fmt"
	"time"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/storage"
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

	// Load epic for validation
	storageService := storage.NewFileStorage()
	epicData, err := storageService.LoadEpic(epicFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load epic: %w", err)
	}

	// Find the test
	var test *epic.Test
	for i := range epicData.Tests {
		if epicData.Tests[i].ID == request.TestID {
			test = &epicData.Tests[i]
			break
		}
	}

	if test == nil {
		return &TestResult{
			Error: &TestError{
				Type:    "test_not_found",
				TestID:  request.TestID,
				Message: fmt.Sprintf("Test %s not found", request.TestID),
			},
		}, nil
	}

	// Epic 13 validation - check if test can be passed
	validationService := tests.NewTestValidationService()
	if err := validationService.CanPassTest(epicData, test); err != nil {
		if statusErr, ok := err.(*epic.StatusValidationError); ok {
			return &TestResult{
				Error: &TestError{
					Type:    "validation_failed",
					TestID:  request.TestID,
					Message: statusErr.Message,
				},
			}, nil
		}
		return &TestResult{
			Error: &TestError{
				Type:    "validation_failed",
				TestID:  request.TestID,
				Message: err.Error(),
			},
		}, nil
	}

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
	if request.TestID == "" {
		return nil, fmt.Errorf("fail-test requires test-id argument")
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

	// Load epic for validation
	storageService := storage.NewFileStorage()
	epicData, err := storageService.LoadEpic(epicFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load epic: %w", err)
	}

	// Find the test
	var test *epic.Test
	for i := range epicData.Tests {
		if epicData.Tests[i].ID == request.TestID {
			test = &epicData.Tests[i]
			break
		}
	}

	if test == nil {
		return &TestResult{
			Error: &TestError{
				Type:    "test_not_found",
				TestID:  request.TestID,
				Message: fmt.Sprintf("Test %s not found", request.TestID),
			},
		}, nil
	}

	// Epic 13 validation - check if test can be failed
	validationService := tests.NewTestValidationService()
	if err := validationService.CanFailTest(epicData, test); err != nil {
		if statusErr, ok := err.(*epic.StatusValidationError); ok {
			return &TestResult{
				Error: &TestError{
					Type:    "validation_failed",
					TestID:  request.TestID,
					Message: statusErr.Message,
				},
			}, nil
		}
		return &TestResult{
			Error: &TestError{
				Type:    "validation_failed",
				TestID:  request.TestID,
				Message: err.Error(),
			},
		}, nil
	}

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

	// Load epic for validation
	storageService := storage.NewFileStorage()
	epicData, err := storageService.LoadEpic(epicFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load epic: %w", err)
	}

	// Find the test
	var test *epic.Test
	for i := range epicData.Tests {
		if epicData.Tests[i].ID == request.TestID {
			test = &epicData.Tests[i]
			break
		}
	}

	if test == nil {
		return &TestResult{
			Error: &TestError{
				Type:    "test_not_found",
				TestID:  request.TestID,
				Message: fmt.Sprintf("Test %s not found", request.TestID),
			},
		}, nil
	}

	// Epic 13 validation - check if test can be cancelled
	validationService := tests.NewTestValidationService()
	if err := validationService.CanCancelTest(epicData, test, request.CancellationReason); err != nil {
		if statusErr, ok := err.(*epic.StatusValidationError); ok {
			return &TestResult{
				Error: &TestError{
					Type:    "validation_failed",
					TestID:  request.TestID,
					Message: statusErr.Message,
				},
			}, nil
		}
		return &TestResult{
			Error: &TestError{
				Type:    "validation_failed",
				TestID:  request.TestID,
				Message: err.Error(),
			},
		}, nil
	}

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
