package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	CurrentEpic     string `json:"current_epic"`
	ProjectName     string `json:"project_name,omitempty"`
	DefaultAssignee string `json:"default_assignee,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		DefaultAssignee: "agent",
	}
}

func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = ".agentpm.json"
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", absPath)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

func SaveConfig(config *Config, configPath string) error {
	if configPath == "" {
		configPath = ".agentpm.json"
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("failed to resolve config path: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	tempFile := absPath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	if err := os.Rename(tempFile, absPath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to move config file: %w", err)
	}

	return nil
}

func (c *Config) Validate() error {
	if c.CurrentEpic == "" {
		return fmt.Errorf("current_epic is required")
	}

	if c.DefaultAssignee == "" {
		c.DefaultAssignee = "agent"
	}

	return nil
}

func (c *Config) EpicFilePath() string {
	if filepath.IsAbs(c.CurrentEpic) {
		return c.CurrentEpic
	}
	return "./" + c.CurrentEpic
}

func ConfigExists(configPath string) bool {
	if configPath == "" {
		configPath = ".agentpm.json"
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return false
	}

	_, err = os.Stat(absPath)
	return err == nil
}
