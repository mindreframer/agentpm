package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestEpic(valid bool) *epic.Epic {
	if valid {
		return &epic.Epic{
			ID:        "test-1",
			Name:      "Test Epic",
			Status:    epic.StatusPending,
			CreatedAt: time.Date(2025, 8, 16, 9, 0, 0, 0, time.UTC),
			Assignee:  "test_agent",
			Phases: []epic.Phase{
				{ID: "P1", Name: "Phase 1", Status: epic.StatusPending},
			},
			Tasks: []epic.Task{
				{ID: "T1", PhaseID: "P1", Name: "Task 1", Status: epic.StatusPending},
			},
			Tests: []epic.Test{
				{ID: "TEST1", TaskID: "T1", Name: "Test 1", Status: epic.StatusPending},
			},
		}
	} else {
		return &epic.Epic{
			// Invalid epic - missing required fields
			Name: "Invalid Epic",
		}
	}
}

func setupTestService(tempDir string, useMemory bool) *EpicService {
	configPath := filepath.Join(tempDir, ".agentpm.json")
	return NewEpicService(ServiceConfig{
		ConfigPath: configPath,
		UseMemory:  useMemory,
		TimeSource: func() time.Time {
			return time.Date(2025, 8, 16, 9, 0, 0, 0, time.UTC)
		},
	})
}

func TestNewEpicService(t *testing.T) {
	t.Run("create service with memory storage", func(t *testing.T) {
		svc := NewEpicService(ServiceConfig{
			ConfigPath: "test-config.json",
			UseMemory:  true,
		})

		assert.NotNil(t, svc)
		assert.Equal(t, "test-config.json", svc.configPath)
		_, ok := svc.storage.(*storage.MemoryStorage)
		assert.True(t, ok, "should use memory storage")
	})

	t.Run("create service with file storage", func(t *testing.T) {
		svc := NewEpicService(ServiceConfig{
			ConfigPath: "test-config.json",
			UseMemory:  false,
		})

		assert.NotNil(t, svc)
		_, ok := svc.storage.(*storage.FileStorage)
		assert.True(t, ok, "should use file storage")
	})

	t.Run("create service with default time source", func(t *testing.T) {
		svc := NewEpicService(ServiceConfig{
			ConfigPath: "test-config.json",
		})

		assert.NotNil(t, svc.timeSource)
		// Test that time source returns a reasonable time
		now := svc.timeSource()
		assert.True(t, time.Since(now) < time.Second)
	})
}

func TestEpicService_InitializeProject(t *testing.T) {
	t.Run("initialize project with valid epic", func(t *testing.T) {
		tempDir := t.TempDir()
		svc := setupTestService(tempDir, true)

		// Store test epic
		testEpic := createTestEpic(true)
		memStorage := svc.storage.(*storage.MemoryStorage)
		memStorage.StoreEpic("test-epic.xml", testEpic)

		result, err := svc.InitializeProject("test-epic.xml")

		require.NoError(t, err)
		assert.True(t, result.ProjectCreated)
		assert.Equal(t, "test-epic.xml", result.CurrentEpic)
		assert.Contains(t, result.ConfigFile, ".agentpm.json")
		assert.NotNil(t, result.Config)
		assert.Equal(t, "test-epic.xml", result.Config.CurrentEpic)
	})

	t.Run("initialize with non-existent epic", func(t *testing.T) {
		tempDir := t.TempDir()
		svc := setupTestService(tempDir, true)

		_, err := svc.InitializeProject("missing-epic.xml")

		require.Error(t, err)
		assert.True(t, IsNotFound(err))

		svcErr := err.(*ServiceError)
		assert.Equal(t, ErrorTypeNotFound, svcErr.Type)
		assert.Contains(t, svcErr.Message, "Epic file not found")
	})

	t.Run("initialize with invalid epic", func(t *testing.T) {
		tempDir := t.TempDir()
		svc := setupTestService(tempDir, false) // Use file storage to test actual XML loading

		// Create a file with malformed XML
		epicPath := filepath.Join(tempDir, "invalid-epic.xml")
		err := os.WriteFile(epicPath, []byte(`<?xml version="1.0"?><epic>malformed xml`), 0644)
		require.NoError(t, err)

		_, err = svc.InitializeProject(epicPath)

		require.Error(t, err)
		assert.True(t, IsValidation(err))
	})

	t.Run("preserve existing config values", func(t *testing.T) {
		tempDir := t.TempDir()
		svc := setupTestService(tempDir, true)

		// Create existing config
		existingConfig := &config.Config{
			CurrentEpic:     "old-epic.xml",
			ProjectName:     "MyProject",
			DefaultAssignee: "custom_agent",
		}
		err := config.SaveConfig(existingConfig, svc.configPath)
		require.NoError(t, err)

		// Store test epic
		testEpic := createTestEpic(true)
		memStorage := svc.storage.(*storage.MemoryStorage)
		memStorage.StoreEpic("new-epic.xml", testEpic)

		result, err := svc.InitializeProject("new-epic.xml")

		require.NoError(t, err)
		assert.Equal(t, "new-epic.xml", result.Config.CurrentEpic)
		assert.Equal(t, "MyProject", result.Config.ProjectName)        // Preserved
		assert.Equal(t, "custom_agent", result.Config.DefaultAssignee) // Preserved
	})
}

func TestEpicService_GetConfiguration(t *testing.T) {
	t.Run("get configuration with existing epic", func(t *testing.T) {
		tempDir := t.TempDir()
		svc := setupTestService(tempDir, true)

		// Create config
		cfg := &config.Config{
			CurrentEpic:     "test-epic.xml",
			ProjectName:     "TestProject",
			DefaultAssignee: "test_agent",
		}
		err := config.SaveConfig(cfg, svc.configPath)
		require.NoError(t, err)

		// Store epic (need to store with the path that EpicFilePath() will return)
		testEpic := createTestEpic(true)
		memStorage := svc.storage.(*storage.MemoryStorage)
		memStorage.StoreEpic("./test-epic.xml", testEpic)

		result, err := svc.GetConfiguration()

		require.NoError(t, err)
		assert.Equal(t, cfg.CurrentEpic, result.Config.CurrentEpic)
		assert.Equal(t, cfg.ProjectName, result.Config.ProjectName)
		assert.Equal(t, cfg.DefaultAssignee, result.Config.DefaultAssignee)
		assert.True(t, result.EpicExists)
		assert.Contains(t, result.ConfigPath, ".agentpm.json")
	})

	t.Run("validate epic with file override", func(t *testing.T) {
		tempDir := t.TempDir()
		svc := setupTestService(tempDir, true)

		// Store valid epic
		testEpic := createTestEpic(true)
		memStorage := svc.storage.(*storage.MemoryStorage)
		memStorage.StoreEpic("override-epic.xml", testEpic)

		result, err := svc.ValidateEpic("override-epic.xml")

		require.NoError(t, err)
		assert.Equal(t, "override-epic.xml", result.EpicFile)
		assert.True(t, result.ValidationResult.Valid)
	})

	t.Run("validate invalid epic", func(t *testing.T) {
		tempDir := t.TempDir()
		svc := setupTestService(tempDir, true)

		// Store invalid epic
		invalidEpic := createTestEpic(false)
		memStorage := svc.storage.(*storage.MemoryStorage)
		memStorage.StoreEpic("invalid-epic.xml", invalidEpic)

		result, err := svc.ValidateEpic("invalid-epic.xml")

		require.NoError(t, err)
		assert.Equal(t, "invalid-epic.xml", result.EpicFile)
		assert.False(t, result.ValidationResult.Valid)
		assert.NotEmpty(t, result.ValidationResult.Errors)
	})

	t.Run("validate non-existent epic", func(t *testing.T) {
		tempDir := t.TempDir()
		svc := setupTestService(tempDir, true)

		_, err := svc.ValidateEpic("missing-epic.xml")

		require.Error(t, err)
		assert.True(t, IsNotFound(err))
	})

	t.Run("validate with missing config", func(t *testing.T) {
		tempDir := t.TempDir()
		svc := setupTestService(tempDir, true)

		_, err := svc.ValidateEpic("")

		require.Error(t, err)
		assert.True(t, IsNotFound(err))
	})
}

func TestEpicService_LoadEpic(t *testing.T) {
	t.Run("load existing epic", func(t *testing.T) {
		tempDir := t.TempDir()
		svc := setupTestService(tempDir, true)

		// Store epic
		testEpic := createTestEpic(true)
		memStorage := svc.storage.(*storage.MemoryStorage)
		memStorage.StoreEpic("test-epic.xml", testEpic)

		result, err := svc.LoadEpic("test-epic.xml")

		require.NoError(t, err)
		assert.Equal(t, testEpic.ID, result.ID)
		assert.Equal(t, testEpic.Name, result.Name)
	})

	t.Run("load non-existent epic", func(t *testing.T) {
		tempDir := t.TempDir()
		svc := setupTestService(tempDir, true)

		_, err := svc.LoadEpic("missing-epic.xml")

		require.Error(t, err)
		assert.True(t, IsNotFound(err))
	})
}

func TestEpicService_SaveEpic(t *testing.T) {
	t.Run("save valid epic", func(t *testing.T) {
		tempDir := t.TempDir()
		svc := setupTestService(tempDir, true)

		testEpic := createTestEpic(true)

		err := svc.SaveEpic(testEpic, "test-epic.xml")

		require.NoError(t, err)

		// Verify epic was saved
		loaded, err := svc.LoadEpic("test-epic.xml")
		require.NoError(t, err)
		assert.Equal(t, testEpic.ID, loaded.ID)
	})

	t.Run("save invalid epic", func(t *testing.T) {
		tempDir := t.TempDir()
		svc := setupTestService(tempDir, true)

		invalidEpic := createTestEpic(false)

		err := svc.SaveEpic(invalidEpic, "invalid-epic.xml")

		require.Error(t, err)
		assert.True(t, IsValidation(err))

		svcErr := err.(*ServiceError)
		assert.Contains(t, svcErr.Message, "Epic validation failed")
		assert.NotEmpty(t, svcErr.Details["errors"])
	})

	t.Run("save nil epic", func(t *testing.T) {
		tempDir := t.TempDir()
		svc := setupTestService(tempDir, true)

		err := svc.SaveEpic(nil, "nil-epic.xml")

		require.Error(t, err)
		assert.True(t, IsValidation(err))
		assert.Contains(t, err.Error(), "Epic cannot be nil")
	})
}

func TestServiceError(t *testing.T) {
	t.Run("error type checking", func(t *testing.T) {
		notFoundErr := &ServiceError{Type: ErrorTypeNotFound, Message: "not found"}
		validationErr := &ServiceError{Type: ErrorTypeValidation, Message: "validation error"}
		ioErr := &ServiceError{Type: ErrorTypeIO, Message: "io error"}

		assert.True(t, IsNotFound(notFoundErr))
		assert.False(t, IsNotFound(validationErr))

		assert.True(t, IsValidation(validationErr))
		assert.False(t, IsValidation(notFoundErr))

		assert.True(t, IsIO(ioErr))
		assert.False(t, IsIO(notFoundErr))
	})

	t.Run("error unwrapping", func(t *testing.T) {
		cause := assert.AnError
		svcErr := &ServiceError{
			Type:    ErrorTypeIO,
			Message: "wrapped error",
			Cause:   cause,
		}

		assert.Equal(t, cause, svcErr.Unwrap())
	})

	t.Run("error details", func(t *testing.T) {
		details := map[string]interface{}{
			"field": "value",
			"count": 42,
		}

		svcErr := &ServiceError{
			Type:    ErrorTypeValidation,
			Message: "validation failed",
			Details: details,
		}

		assert.Equal(t, details, svcErr.Details)
		assert.Equal(t, "value", svcErr.Details["field"])
		assert.Equal(t, 42, svcErr.Details["count"])
	})
}

func TestServiceConfig(t *testing.T) {
	t.Run("service config with custom time source", func(t *testing.T) {
		fixedTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

		svc := NewEpicService(ServiceConfig{
			ConfigPath: "test.json",
			UseMemory:  true,
			TimeSource: func() time.Time { return fixedTime },
		})

		assert.Equal(t, fixedTime, svc.timeSource())
	})
}

func TestServiceResults(t *testing.T) {
	t.Run("init result structure", func(t *testing.T) {
		cfg := &config.Config{
			CurrentEpic:     "test.xml",
			DefaultAssignee: "agent",
		}

		result := &InitResult{
			ProjectCreated: true,
			ConfigFile:     "config.json",
			CurrentEpic:    "test.xml",
			Config:         cfg,
		}

		assert.True(t, result.ProjectCreated)
		assert.Equal(t, "config.json", result.ConfigFile)
		assert.Equal(t, "test.xml", result.CurrentEpic)
		assert.Equal(t, cfg, result.Config)
	})

	t.Run("config result structure", func(t *testing.T) {
		cfg := &config.Config{
			CurrentEpic:     "test.xml",
			DefaultAssignee: "agent",
		}

		result := &ConfigResult{
			Config:     cfg,
			EpicExists: true,
			ConfigPath: "config.json",
		}

		assert.Equal(t, cfg, result.Config)
		assert.True(t, result.EpicExists)
		assert.Equal(t, "config.json", result.ConfigPath)
	})

	t.Run("validation result structure", func(t *testing.T) {
		vr := &epic.ValidationResult{
			Valid: true,
		}

		result := &ValidationResult{
			EpicFile:         "test.xml",
			ValidationResult: vr,
		}

		assert.Equal(t, "test.xml", result.EpicFile)
		assert.Equal(t, vr, result.ValidationResult)
	})
}
