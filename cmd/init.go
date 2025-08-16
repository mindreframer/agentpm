package cmd

import (
	"context"
	"fmt"

	"github.com/memomoo/agentpm/internal/config"
	"github.com/memomoo/agentpm/internal/storage"
	"github.com/urfave/cli/v3"
)

func InitCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialize a new project with an epic file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "epic",
				Usage:    "Epic file to set as current",
				Required: true,
			},
		},
		Action: runInit,
	}
}

func runInit(ctx context.Context, c *cli.Command) error {
	epicFile := c.String("epic")
	configPath := c.String("config")
	format := c.String("format")

	// Check if epic file exists
	storage := storage.NewFileStorage()
	if !storage.EpicExists(epicFile) {
		return writeError(c, format, fmt.Sprintf("Epic file not found: %s", epicFile))
	}

	// Try to load and validate the epic file
	_, err := storage.LoadEpic(epicFile)
	if err != nil {
		return writeError(c, format, fmt.Sprintf("Failed to load epic file: %v", err))
	}

	// Create or update configuration
	cfg := &config.Config{
		CurrentEpic:     epicFile,
		DefaultAssignee: "agent",
	}

	// If config already exists, preserve project name and assignee
	if config.ConfigExists(configPath) {
		existingCfg, err := config.LoadConfig(configPath)
		if err == nil {
			cfg.ProjectName = existingCfg.ProjectName
			if existingCfg.DefaultAssignee != "" {
				cfg.DefaultAssignee = existingCfg.DefaultAssignee
			}
		}
	}

	// Save configuration
	err = config.SaveConfig(cfg, configPath)
	if err != nil {
		return writeError(c, format, fmt.Sprintf("Failed to save configuration: %v", err))
	}

	// Write success response
	return writeInitResult(c, format, cfg, configPath)
}

func writeInitResult(c *cli.Command, format string, cfg *config.Config, configPath string) error {
	switch format {
	case "xml":
		output := fmt.Sprintf(`<init_result>
    <project_created>true</project_created>
    <config_file>%s</config_file>
    <current_epic>%s</current_epic>
</init_result>`, configPath, cfg.CurrentEpic)
		fmt.Fprint(c.Root().Writer, output)
	case "json":
		output := fmt.Sprintf(`{
  "project_created": true,
  "config_file": "%s",
  "current_epic": "%s"
}`, configPath, cfg.CurrentEpic)
		fmt.Fprint(c.Root().Writer, output)
	default: // text
		fmt.Fprintf(c.Root().Writer, "✓ Project initialized successfully\n")
		fmt.Fprintf(c.Root().Writer, "Config file: %s\n", configPath)
		fmt.Fprintf(c.Root().Writer, "Current epic: %s\n", cfg.CurrentEpic)
	}
	return nil
}

func writeError(c *cli.Command, format string, message string) error {
	switch format {
	case "xml":
		output := fmt.Sprintf(`<error>
    <type>init_error</type>
    <message>%s</message>
</error>`, message)
		fmt.Fprint(c.Root().ErrWriter, output)
	case "json":
		output := fmt.Sprintf(`{
  "error": {
    "type": "init_error",
    "message": "%s"
  }
}`, message)
		fmt.Fprint(c.Root().ErrWriter, output)
	default: // text
		fmt.Fprintf(c.Root().ErrWriter, "✗ Error: %s\n", message)
	}
	return fmt.Errorf("%s", message)
}
