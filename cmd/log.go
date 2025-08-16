package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/memomoo/agentpm/internal/config"
	"github.com/memomoo/agentpm/internal/epic"
	"github.com/memomoo/agentpm/internal/query"
	"github.com/memomoo/agentpm/internal/storage"
	"github.com/urfave/cli/v3"
)

func LogCommand() *cli.Command {
	return &cli.Command{
		Name:      "log",
		Usage:     "Log an event to the current epic",
		ArgsUsage: "<message>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "file",
				Usage: "Epic file path (overrides config)",
			},
			&cli.StringFlag{
				Name:  "files",
				Usage: "Files involved in this event (format: 'path:action,path2:action2')",
			},
			&cli.StringFlag{
				Name:  "type",
				Usage: "Event type: implementation, blocker, issue, etc.",
				Value: "implementation",
			},
			&cli.StringFlag{
				Name:  "time",
				Usage: "Timestamp for the event (ISO 8601 format)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() < 1 {
				return fmt.Errorf("event message is required")
			}

			message := cmd.Args().First()

			// Get epic file path
			epicFile := cmd.String("file")
			if epicFile == "" {
				cfg, err := config.LoadConfig(cmd.String("config"))
				if err != nil {
					return fmt.Errorf("failed to load configuration: %w", err)
				}
				epicFile = cfg.CurrentEpic
			}

			if epicFile == "" {
				return fmt.Errorf("no epic file specified (use --file flag or set current epic)")
			}

			// Parse timestamp if provided
			var timestamp time.Time
			if timeStr := cmd.String("time"); timeStr != "" {
				var err error
				timestamp, err = time.Parse(time.RFC3339, timeStr)
				if err != nil {
					return fmt.Errorf("invalid time format: %s (use ISO 8601 format like 2025-08-16T15:30:00Z)", timeStr)
				}
			} else {
				timestamp = time.Now()
			}

			// Validate event type
			eventType := cmd.String("type")
			if !isValidEventType(eventType) {
				return fmt.Errorf("invalid event type: %s (valid types: implementation, blocker, issue, milestone, decision, note)", eventType)
			}

			// Parse files flag
			files, err := parseFilesFlag(cmd.String("files"))
			if err != nil {
				return fmt.Errorf("invalid files format: %w", err)
			}

			// Initialize services
			storageImpl := storage.NewFileStorage()
			queryService := query.NewQueryService(storageImpl)
			logService := NewLogService(storageImpl, queryService)

			// Log the event
			err = logService.LogEvent(epicFile, message, eventType, files, timestamp)
			if err != nil {
				return fmt.Errorf("failed to log event: %w", err)
			}

			// Output confirmation
			fmt.Fprintf(cmd.Writer, "Event logged: %s\n", message)
			return nil
		},
	}
}

func isValidEventType(eventType string) bool {
	validTypes := []string{
		"implementation",
		"blocker",
		"issue",
		"milestone",
		"decision",
		"note",
	}

	for _, validType := range validTypes {
		if eventType == validType {
			return true
		}
	}
	return false
}

type FileAction struct {
	Path   string
	Action string
}

func parseFilesFlag(filesStr string) ([]FileAction, error) {
	if filesStr == "" {
		return nil, nil
	}

	var files []FileAction
	parts := strings.Split(filesStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split by last colon to handle paths with colons
		colonIndex := strings.LastIndex(part, ":")
		if colonIndex == -1 {
			return nil, fmt.Errorf("invalid file format '%s': expected 'path:action'", part)
		}

		path := part[:colonIndex]
		action := part[colonIndex+1:]

		if path == "" || action == "" {
			return nil, fmt.Errorf("invalid file format '%s': path and action cannot be empty", part)
		}

		if !isValidFileAction(action) {
			return nil, fmt.Errorf("invalid file action '%s': valid actions are added, modified, deleted, renamed", action)
		}

		files = append(files, FileAction{
			Path:   path,
			Action: action,
		})
	}

	return files, nil
}

func isValidFileAction(action string) bool {
	validActions := []string{"added", "modified", "deleted", "renamed"}
	for _, validAction := range validActions {
		if action == validAction {
			return true
		}
	}
	return false
}

type LogService struct {
	storage storage.Storage
	query   *query.QueryService
}

func NewLogService(storage storage.Storage, query *query.QueryService) *LogService {
	return &LogService{
		storage: storage,
		query:   query,
	}
}

func (ls *LogService) LogEvent(epicFile, message, eventType string, files []FileAction, timestamp time.Time) error {
	// Load epic
	epicData, err := ls.storage.LoadEpic(epicFile)
	if err != nil {
		return fmt.Errorf("failed to load epic: %w", err)
	}

	// Create event data
	eventData := message
	if len(files) > 0 {
		var filesList []string
		for _, file := range files {
			filesList = append(filesList, fmt.Sprintf("%s:%s", file.Path, file.Action))
		}
		eventData = fmt.Sprintf("%s [files: %s]", message, strings.Join(filesList, ", "))
	}

	// Add event to epic
	err = ls.addEventToEpic(epicData, eventType, eventData, timestamp)
	if err != nil {
		return fmt.Errorf("failed to add event: %w", err)
	}

	// Save epic atomically
	err = ls.storage.SaveEpic(epicData, epicFile)
	if err != nil {
		return fmt.Errorf("failed to save epic: %w", err)
	}

	return nil
}

func (ls *LogService) addEventToEpic(epicData *epic.Epic, eventType, eventData string, timestamp time.Time) error {
	// Generate simple event ID using timestamp
	eventID := fmt.Sprintf("event_%d", timestamp.Unix())

	// Create new event
	newEvent := epic.Event{
		ID:        eventID,
		Type:      eventType,
		Timestamp: timestamp,
		Data:      eventData,
	}

	// Add event to epic
	epicData.Events = append(epicData.Events, newEvent)

	return nil
}
