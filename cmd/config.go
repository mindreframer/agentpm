package cmd

import (
	"context"
	"fmt"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/urfave/cli/v3"
)

func ConfigCommand() *cli.Command {
	return &cli.Command{
		Name:   "config",
		Usage:  "Display current project configuration",
		Action: runConfig,
	}
}

func runConfig(ctx context.Context, c *cli.Command) error {
	configPath := c.String("config")
	format := c.String("format")

	// Try to load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return writeError(c, format, fmt.Sprintf("Failed to load configuration: %v", err))
	}

	// Check if epic file exists and warn if missing
	storage := storage.NewFileStorage()
	epicExists := storage.EpicExists(cfg.EpicFilePath())

	return writeConfigResult(c, format, cfg, !epicExists)
}

func writeConfigResult(c *cli.Command, format string, cfg *config.Config, epicMissing bool) error {
	switch format {
	case "xml":
		output := fmt.Sprintf(`<config>
    <current_epic>%s</current_epic>`, cfg.CurrentEpic)

		if cfg.ProjectName != "" {
			output += fmt.Sprintf(`
    <project_name>%s</project_name>`, cfg.ProjectName)
		}

		output += fmt.Sprintf(`
    <default_assignee>%s</default_assignee>`, cfg.DefaultAssignee)

		if epicMissing {
			output += `
    <warnings>
        <warning>Epic file not found</warning>
    </warnings>`
		}

		output += `
</config>`
		fmt.Fprint(c.Root().Writer, output)

	case "json":
		output := fmt.Sprintf(`{
  "current_epic": "%s",`, cfg.CurrentEpic)

		if cfg.ProjectName != "" {
			output += fmt.Sprintf(`
  "project_name": "%s",`, cfg.ProjectName)
		}

		output += fmt.Sprintf(`
  "default_assignee": "%s"`, cfg.DefaultAssignee)

		if epicMissing {
			output += `,
  "warnings": ["Epic file not found"]`
		}

		output += `
}`
		fmt.Fprint(c.Root().Writer, output)

	default: // text
		fmt.Fprintf(c.Root().Writer, "Current Configuration:\n")
		fmt.Fprintf(c.Root().Writer, "  Current epic: %s\n", cfg.CurrentEpic)
		if cfg.ProjectName != "" {
			fmt.Fprintf(c.Root().Writer, "  Project name: %s\n", cfg.ProjectName)
		}
		fmt.Fprintf(c.Root().Writer, "  Default assignee: %s\n", cfg.DefaultAssignee)

		if epicMissing {
			fmt.Fprintf(c.Root().Writer, "\n⚠ Warning: Epic file not found: %s\n", cfg.EpicFilePath())
		}
	}

	return nil
}
