package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.Equal(t, "agent", config.DefaultAssignee)
	assert.Empty(t, config.CurrentEpic)
	assert.Empty(t, config.ProjectName)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with existing assignee",
			config: &Config{
				CurrentEpic:     "epic-8.xml",
				ProjectName:     "TestProject",
				DefaultAssignee: "agent_claude",
			},
			wantErr: false,
		},
		{
			name: "missing current_epic",
			config: &Config{
				ProjectName:     "TestProject",
				DefaultAssignee: "agent_claude",
			},
			wantErr: true,
			errMsg:  "current_epic is required",
		},
		{
			name: "empty default_assignee gets filled",
			config: &Config{
				CurrentEpic: "epic-8.xml",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				// Only check default assignee for the specific test case
				if tt.name == "empty_default_assignee_gets_filled" {
					assert.Equal(t, "agent", tt.config.DefaultAssignee)
				}
			}
		})
	}
}

func TestConfig_EpicFilePath(t *testing.T) {
	tests := []struct {
		name        string
		currentEpic string
		expected    string
	}{
		{
			name:        "relative path",
			currentEpic: "epic-8.xml",
			expected:    "./epic-8.xml",
		},
		{
			name:        "absolute path",
			currentEpic: "/abs/path/epic-8.xml",
			expected:    "/abs/path/epic-8.xml",
		},
		{
			name:        "complex relative path",
			currentEpic: "epics/epic-8.xml",
			expected:    "./epics/epic-8.xml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{CurrentEpic: tt.currentEpic}
			result := config.EpicFilePath()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadConfig(t *testing.T) {
	t.Run("load valid config", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test-config.json")

		// Create test config
		testConfig := &Config{
			CurrentEpic:     "epic-8.xml",
			ProjectName:     "TestProject",
			DefaultAssignee: "agent_claude",
		}

		err := SaveConfig(testConfig, configPath)
		require.NoError(t, err)

		// Load the config
		loaded, err := LoadConfig(configPath)
		require.NoError(t, err)

		assert.Equal(t, testConfig.CurrentEpic, loaded.CurrentEpic)
		assert.Equal(t, testConfig.ProjectName, loaded.ProjectName)
		assert.Equal(t, testConfig.DefaultAssignee, loaded.DefaultAssignee)
	})

	t.Run("load non-existent config", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "non-existent.json")

		_, err := LoadConfig(configPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config file not found")
	})

	t.Run("load invalid JSON", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "invalid.json")

		// Write invalid JSON
		err := os.WriteFile(configPath, []byte(`{"current_epic": "test", invalid json`), 0644)
		require.NoError(t, err)

		_, err = LoadConfig(configPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse config file")
	})

	t.Run("load config with invalid data", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "invalid-data.json")

		// Write JSON with missing required field
		err := os.WriteFile(configPath, []byte(`{"project_name": "test"}`), 0644)
		require.NoError(t, err)

		_, err = LoadConfig(configPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid configuration")
	})

	t.Run("default config path", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)

		err := os.Chdir(tempDir)
		require.NoError(t, err)

		_, err = LoadConfig("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config file not found")
	})
}

func TestSaveConfig(t *testing.T) {
	t.Run("save valid config", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test-config.json")

		config := &Config{
			CurrentEpic:     "epic-8.xml",
			ProjectName:     "TestProject",
			DefaultAssignee: "agent_claude",
		}

		err := SaveConfig(config, configPath)
		assert.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(configPath)
		assert.NoError(t, err)

		// Verify content by loading back
		loaded, err := LoadConfig(configPath)
		require.NoError(t, err)
		assert.Equal(t, config.CurrentEpic, loaded.CurrentEpic)
		assert.Equal(t, config.ProjectName, loaded.ProjectName)
		assert.Equal(t, config.DefaultAssignee, loaded.DefaultAssignee)
	})

	t.Run("save invalid config", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test-config.json")

		config := &Config{
			// Missing CurrentEpic
			ProjectName:     "TestProject",
			DefaultAssignee: "agent_claude",
		}

		err := SaveConfig(config, configPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid configuration")
	})

	t.Run("save to non-existent directory", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "subdir", "test-config.json")

		config := &Config{
			CurrentEpic:     "epic-8.xml",
			DefaultAssignee: "agent",
		}

		err := SaveConfig(config, configPath)
		assert.NoError(t, err)

		// Verify directory was created
		_, err = os.Stat(filepath.Dir(configPath))
		assert.NoError(t, err)
	})

	t.Run("atomic write operation", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "atomic-test.json")

		// Create initial config
		config1 := &Config{
			CurrentEpic:     "epic-1.xml",
			DefaultAssignee: "agent",
		}
		err := SaveConfig(config1, configPath)
		require.NoError(t, err)

		// Update config
		config2 := &Config{
			CurrentEpic:     "epic-2.xml",
			ProjectName:     "Updated",
			DefaultAssignee: "agent",
		}
		err = SaveConfig(config2, configPath)
		assert.NoError(t, err)

		// Verify final state
		loaded, err := LoadConfig(configPath)
		require.NoError(t, err)
		assert.Equal(t, "epic-2.xml", loaded.CurrentEpic)
		assert.Equal(t, "Updated", loaded.ProjectName)

		// Verify no temp file left behind
		tempFile := configPath + ".tmp"
		_, err = os.Stat(tempFile)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("default config path", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)

		err := os.Chdir(tempDir)
		require.NoError(t, err)

		config := &Config{
			CurrentEpic:     "epic-8.xml",
			DefaultAssignee: "agent",
		}

		err = SaveConfig(config, "")
		assert.NoError(t, err)

		// Check that .agentpm.json was created
		_, err = os.Stat(".agentpm.json")
		assert.NoError(t, err)
	})
}

func TestConfigExists(t *testing.T) {
	t.Run("existing config", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "exists.json")

		// Create config file
		config := &Config{
			CurrentEpic:     "epic-8.xml",
			DefaultAssignee: "agent",
		}
		err := SaveConfig(config, configPath)
		require.NoError(t, err)

		assert.True(t, ConfigExists(configPath))
	})

	t.Run("non-existing config", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "does-not-exist.json")

		assert.False(t, ConfigExists(configPath))
	})

	t.Run("default path", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)

		err := os.Chdir(tempDir)
		require.NoError(t, err)

		// Should not exist initially
		assert.False(t, ConfigExists(""))

		// Create default config
		config := &Config{
			CurrentEpic:     "epic-8.xml",
			DefaultAssignee: "agent",
		}
		err = SaveConfig(config, "")
		require.NoError(t, err)

		// Should exist now
		assert.True(t, ConfigExists(""))
	})
}
