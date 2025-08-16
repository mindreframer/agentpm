package service

import (
	"fmt"
	"time"

	"github.com/memomoo/agentpm/internal/config"
	"github.com/memomoo/agentpm/internal/epic"
	"github.com/memomoo/agentpm/internal/storage"
)

// EpicService provides high-level operations for epic management
type EpicService struct {
	storage    storage.Storage
	configPath string
	timeSource func() time.Time
}

// ServiceConfig holds configuration for creating services
type ServiceConfig struct {
	ConfigPath string
	UseMemory  bool
	TimeSource func() time.Time
}

// NewEpicService creates a new epic service with the given configuration
func NewEpicService(cfg ServiceConfig) *EpicService {
	factory := storage.NewFactory(cfg.UseMemory)

	timeSource := cfg.TimeSource
	if timeSource == nil {
		timeSource = time.Now
	}

	return &EpicService{
		storage:    factory.CreateStorage(),
		configPath: cfg.ConfigPath,
		timeSource: timeSource,
	}
}

// InitializeProject initializes a new project with the given epic file
func (s *EpicService) InitializeProject(epicFile string) (*InitResult, error) {
	// Validate epic file exists and is readable
	if !s.storage.EpicExists(epicFile) {
		return nil, &ServiceError{
			Type:    ErrorTypeNotFound,
			Message: fmt.Sprintf("Epic file not found: %s", epicFile),
		}
	}

	// Load and validate epic
	_, err := s.storage.LoadEpic(epicFile)
	if err != nil {
		return nil, &ServiceError{
			Type:    ErrorTypeValidation,
			Message: fmt.Sprintf("Failed to load epic file: %v", err),
			Cause:   err,
		}
	}

	// Load existing config or create new one
	var cfg *config.Config
	if config.ConfigExists(s.configPath) {
		existingCfg, err := config.LoadConfig(s.configPath)
		if err == nil {
			cfg = existingCfg
		}
	}

	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// Update config with new epic
	cfg.CurrentEpic = epicFile

	// Save configuration atomically
	err = s.withTransaction(func() error {
		return config.SaveConfig(cfg, s.configPath)
	})

	if err != nil {
		return nil, &ServiceError{
			Type:    ErrorTypeIO,
			Message: fmt.Sprintf("Failed to save configuration: %v", err),
			Cause:   err,
		}
	}

	return &InitResult{
		ProjectCreated: true,
		ConfigFile:     s.resolveConfigPath(),
		CurrentEpic:    epicFile,
		Config:         cfg,
	}, nil
}

// GetConfiguration returns the current project configuration
func (s *EpicService) GetConfiguration() (*ConfigResult, error) {
	cfg, err := config.LoadConfig(s.configPath)
	if err != nil {
		return nil, &ServiceError{
			Type:    ErrorTypeNotFound,
			Message: fmt.Sprintf("Failed to load configuration: %v", err),
			Cause:   err,
		}
	}

	// Check if epic file exists
	epicExists := s.storage.EpicExists(cfg.EpicFilePath())

	return &ConfigResult{
		Config:     cfg,
		EpicExists: epicExists,
		ConfigPath: s.resolveConfigPath(),
	}, nil
}

// ValidateEpic validates an epic file, using config or file override
func (s *EpicService) ValidateEpic(fileOverride string) (*ValidationResult, error) {
	var epicFile string

	if fileOverride != "" {
		epicFile = fileOverride
	} else {
		// Load from config
		cfg, err := config.LoadConfig(s.configPath)
		if err != nil {
			return nil, &ServiceError{
				Type:    ErrorTypeNotFound,
				Message: fmt.Sprintf("Failed to load configuration: %v", err),
				Cause:   err,
			}
		}
		epicFile = cfg.EpicFilePath()
	}

	// Check if file exists
	if !s.storage.EpicExists(epicFile) {
		return nil, &ServiceError{
			Type:    ErrorTypeNotFound,
			Message: fmt.Sprintf("Epic file not found: %s", epicFile),
		}
	}

	// Validate the epic
	result, err := epic.ValidateFromFile(s.storage, epicFile)
	if err != nil {
		return nil, &ServiceError{
			Type:    ErrorTypeValidation,
			Message: fmt.Sprintf("Failed to validate epic: %v", err),
			Cause:   err,
		}
	}

	return &ValidationResult{
		EpicFile:         epicFile,
		ValidationResult: result,
	}, nil
}

// LoadEpic loads an epic from storage
func (s *EpicService) LoadEpic(epicFile string) (*epic.Epic, error) {
	if !s.storage.EpicExists(epicFile) {
		return nil, &ServiceError{
			Type:    ErrorTypeNotFound,
			Message: fmt.Sprintf("Epic file not found: %s", epicFile),
		}
	}

	e, err := s.storage.LoadEpic(epicFile)
	if err != nil {
		return nil, &ServiceError{
			Type:    ErrorTypeIO,
			Message: fmt.Sprintf("Failed to load epic: %v", err),
			Cause:   err,
		}
	}

	return e, nil
}

// SaveEpic saves an epic to storage
func (s *EpicService) SaveEpic(e *epic.Epic, epicFile string) error {
	if e == nil {
		return &ServiceError{
			Type:    ErrorTypeValidation,
			Message: "Epic cannot be nil",
		}
	}

	// Validate epic before saving
	result := e.Validate()
	if !result.Valid {
		return &ServiceError{
			Type:    ErrorTypeValidation,
			Message: fmt.Sprintf("Epic validation failed: %s", result.Message()),
			Details: map[string]interface{}{
				"errors":   result.Errors,
				"warnings": result.Warnings,
			},
		}
	}

	// Save atomically
	err := s.withTransaction(func() error {
		return s.storage.SaveEpic(e, epicFile)
	})

	if err != nil {
		return &ServiceError{
			Type:    ErrorTypeIO,
			Message: fmt.Sprintf("Failed to save epic: %v", err),
			Cause:   err,
		}
	}

	return nil
}

// withTransaction provides a simple transaction-like mechanism for file operations
func (s *EpicService) withTransaction(operation func() error) error {
	// In a more complex system, this would handle rollback
	// For now, it's just a wrapper for future enhancement
	return operation()
}

// resolveConfigPath returns the actual config path being used
func (s *EpicService) resolveConfigPath() string {
	if s.configPath == "" {
		return ".agentpm.json"
	}
	return s.configPath
}

// Result types for service operations
type InitResult struct {
	ProjectCreated bool           `json:"project_created"`
	ConfigFile     string         `json:"config_file"`
	CurrentEpic    string         `json:"current_epic"`
	Config         *config.Config `json:"-"` // Internal use only
}

type ConfigResult struct {
	Config     *config.Config `json:"config"`
	EpicExists bool           `json:"epic_exists"`
	ConfigPath string         `json:"config_path"`
}

type ValidationResult struct {
	EpicFile         string                 `json:"epic_file"`
	ValidationResult *epic.ValidationResult `json:"validation_result"`
}

// Error types and handling
type ErrorType string

const (
	ErrorTypeNotFound   ErrorType = "not_found"
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeIO         ErrorType = "io"
	ErrorTypeConfig     ErrorType = "config"
)

type ServiceError struct {
	Type    ErrorType              `json:"type"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Cause   error                  `json:"-"`
}

func (e *ServiceError) Error() string {
	return e.Message
}

func (e *ServiceError) Unwrap() error {
	return e.Cause
}

// IsNotFound checks if error is a not found error
func IsNotFound(err error) bool {
	if se, ok := err.(*ServiceError); ok {
		return se.Type == ErrorTypeNotFound
	}
	return false
}

// IsValidation checks if error is a validation error
func IsValidation(err error) bool {
	if se, ok := err.(*ServiceError); ok {
		return se.Type == ErrorTypeValidation
	}
	return false
}

// IsIO checks if error is an IO error
func IsIO(err error) bool {
	if se, ok := err.(*ServiceError); ok {
		return se.Type == ErrorTypeIO
	}
	return false
}
