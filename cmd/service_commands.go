package cmd

import (
	"context"
	"fmt"

	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/service"
	"github.com/urfave/cli/v3"
)

// ServiceBasedInitCommand creates an init command using the service layer
func ServiceBasedInitCommand() *cli.Command {
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
		Action: runServiceBasedInit,
	}
}

func runServiceBasedInit(ctx context.Context, c *cli.Command) error {
	epicFile := c.String("epic")
	configPath := c.String("config")
	format := c.String("format")

	svc := service.NewEpicService(service.ServiceConfig{
		ConfigPath: configPath,
		UseMemory:  false,
	})

	result, err := svc.InitializeProject(epicFile)
	if err != nil {
		return writeServiceError(c, format, err)
	}

	return writeServiceInitResult(c, format, result)
}

// ServiceBasedConfigCommand creates a config command using the service layer
func ServiceBasedConfigCommand() *cli.Command {
	return &cli.Command{
		Name:   "config",
		Usage:  "Display current project configuration",
		Action: runServiceBasedConfig,
	}
}

func runServiceBasedConfig(ctx context.Context, c *cli.Command) error {
	configPath := c.String("config")
	format := c.String("format")

	svc := service.NewEpicService(service.ServiceConfig{
		ConfigPath: configPath,
		UseMemory:  false,
	})

	result, err := svc.GetConfiguration()
	if err != nil {
		return writeServiceError(c, format, err)
	}

	return writeServiceConfigResult(c, format, result)
}

// ServiceBasedValidateCommand creates a validate command using the service layer
func ServiceBasedValidateCommand() *cli.Command {
	return &cli.Command{
		Name:  "validate",
		Usage: "Validate epic XML structure",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Epic file to validate (overrides config)",
			},
		},
		Action: runServiceBasedValidate,
	}
}

func runServiceBasedValidate(ctx context.Context, c *cli.Command) error {
	configPath := c.String("config")
	format := c.String("format")
	fileOverride := c.String("file")

	svc := service.NewEpicService(service.ServiceConfig{
		ConfigPath: configPath,
		UseMemory:  false,
	})

	result, err := svc.ValidateEpic(fileOverride)
	if err != nil {
		return writeServiceError(c, format, err)
	}

	return writeServiceValidationResult(c, format, result)
}

// Output formatters for service results
func writeServiceInitResult(c *cli.Command, format string, result *service.InitResult) error {
	switch format {
	case "xml":
		output := fmt.Sprintf(`<init_result>
    <project_created>%v</project_created>
    <config_file>%s</config_file>
    <current_epic>%s</current_epic>
</init_result>`, result.ProjectCreated, result.ConfigFile, result.CurrentEpic)
		fmt.Fprint(c.Root().Writer, output)
	case "json":
		output := fmt.Sprintf(`{
  "project_created": %v,
  "config_file": "%s",
  "current_epic": "%s"
}`, result.ProjectCreated, result.ConfigFile, result.CurrentEpic)
		fmt.Fprint(c.Root().Writer, output)
	default: // text
		fmt.Fprintf(c.Root().Writer, "✓ Project initialized successfully\n")
		fmt.Fprintf(c.Root().Writer, "Config file: %s\n", result.ConfigFile)
		fmt.Fprintf(c.Root().Writer, "Current epic: %s\n", result.CurrentEpic)
	}
	return nil
}

func writeServiceConfigResult(c *cli.Command, format string, result *service.ConfigResult) error {
	switch format {
	case "xml":
		output := fmt.Sprintf(`<config>
    <current_epic>%s</current_epic>`, result.Config.CurrentEpic)

		if result.Config.ProjectName != "" {
			output += fmt.Sprintf(`
    <project_name>%s</project_name>`, result.Config.ProjectName)
		}

		output += fmt.Sprintf(`
    <default_assignee>%s</default_assignee>`, result.Config.DefaultAssignee)

		if !result.EpicExists {
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
  "current_epic": "%s",`, result.Config.CurrentEpic)

		if result.Config.ProjectName != "" {
			output += fmt.Sprintf(`
  "project_name": "%s",`, result.Config.ProjectName)
		}

		output += fmt.Sprintf(`
  "default_assignee": "%s"`, result.Config.DefaultAssignee)

		if !result.EpicExists {
			output += `,
  "warnings": ["Epic file not found"]`
		}

		output += `
}`
		fmt.Fprint(c.Root().Writer, output)

	default: // text
		fmt.Fprintf(c.Root().Writer, "Current Configuration:\n")
		fmt.Fprintf(c.Root().Writer, "  Current epic: %s\n", result.Config.CurrentEpic)
		if result.Config.ProjectName != "" {
			fmt.Fprintf(c.Root().Writer, "  Project name: %s\n", result.Config.ProjectName)
		}
		fmt.Fprintf(c.Root().Writer, "  Default assignee: %s\n", result.Config.DefaultAssignee)

		if !result.EpicExists {
			fmt.Fprintf(c.Root().Writer, "\n⚠ Warning: Epic file not found: %s\n", result.Config.EpicFilePath())
		}
	}

	return nil
}

func writeServiceValidationResult(c *cli.Command, format string, result *service.ValidationResult) error {
	epicName := getEpicName(result.EpicFile)
	vr := result.ValidationResult

	switch format {
	case "xml":
		output := fmt.Sprintf(`<validation_result epic="%s">
    <valid>%v</valid>`, epicName, vr.Valid)

		if len(vr.Warnings) > 0 {
			output += `
    <warnings>`
			for _, warning := range vr.Warnings {
				output += fmt.Sprintf(`
        <warning>%s</warning>`, warning)
			}
			output += `
    </warnings>`
		}

		if len(vr.Errors) > 0 {
			output += `
    <errors>`
			for _, err := range vr.Errors {
				output += fmt.Sprintf(`
        <error>%s</error>`, err)
			}
			output += `
    </errors>`
		}

		if len(vr.Checks) > 0 {
			output += `
    <checks_performed>`
			for name, status := range vr.Checks {
				output += fmt.Sprintf(`
        <check name="%s">%s</check>`, name, status)
			}
			output += `
    </checks_performed>`
		}

		output += fmt.Sprintf(`
    <message>%s</message>
</validation_result>`, vr.Message())

		fmt.Fprint(c.Root().Writer, output)

	case "json":
		output := fmt.Sprintf(`{
  "epic": "%s",
  "valid": %v,
  "message": "%s"`, epicName, vr.Valid, vr.Message())

		if len(vr.Warnings) > 0 {
			output += `,
  "warnings": [`
			for i, warning := range vr.Warnings {
				if i > 0 {
					output += `, `
				}
				output += fmt.Sprintf(`"%s"`, warning)
			}
			output += `]`
		}

		if len(vr.Errors) > 0 {
			output += `,
  "errors": [`
			for i, err := range vr.Errors {
				if i > 0 {
					output += `, `
				}
				output += fmt.Sprintf(`"%s"`, err)
			}
			output += `]`
		}

		if len(vr.Checks) > 0 {
			output += `,
  "checks_performed": {`
			i := 0
			for name, status := range vr.Checks {
				if i > 0 {
					output += `, `
				}
				output += fmt.Sprintf(`"%s": "%s"`, name, status)
				i++
			}
			output += `}`
		}

		output += `
}`
		fmt.Fprint(c.Root().Writer, output)

	default: // text
		// Use the epic package's text formatter
		output := epic.FormatValidationResult(result.ValidationResult, "text")
		fmt.Fprint(c.Root().Writer, output)
	}

	// Return error if validation failed
	if !vr.Valid {
		return fmt.Errorf("validation failed")
	}

	return nil
}

func writeServiceError(c *cli.Command, format string, err error) error {
	if svcErr, ok := err.(*service.ServiceError); ok {
		switch format {
		case "xml":
			output := fmt.Sprintf(`<error>
    <type>%s</type>
    <message>%s</message>
</error>`, svcErr.Type, svcErr.Message)
			fmt.Fprint(c.Root().ErrWriter, output)
		case "json":
			output := fmt.Sprintf(`{
  "error": {
    "type": "%s",
    "message": "%s"
  }
}`, svcErr.Type, svcErr.Message)
			fmt.Fprint(c.Root().ErrWriter, output)
		default: // text
			fmt.Fprintf(c.Root().ErrWriter, "✗ Error: %s\n", svcErr.Message)
		}
	} else {
		// Fallback for non-service errors
		fmt.Fprintf(c.Root().ErrWriter, "✗ Error: %s\n", err.Error())
	}
	return err
}
