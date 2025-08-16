package cmd

import (
	"context"
	"fmt"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/urfave/cli/v3"
)

func ValidateCommand() *cli.Command {
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
		Action: runValidate,
	}
}

func runValidate(ctx context.Context, c *cli.Command) error {
	configPath := c.String("config")
	format := c.String("format")
	epicFile := c.String("file")

	// Determine which epic file to validate
	if epicFile == "" {
		// Load from config
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return writeError(c, format, fmt.Sprintf("Failed to load configuration: %v", err))
		}
		epicFile = cfg.EpicFilePath()
	}

	// Create storage and validate
	storage := storage.NewFileStorage()

	// Check if file exists first
	if !storage.EpicExists(epicFile) {
		return writeError(c, format, fmt.Sprintf("Epic file not found: %s", epicFile))
	}

	// Validate the epic
	result, err := epic.ValidateFromFile(storage, epicFile)
	if err != nil {
		return writeError(c, format, fmt.Sprintf("Failed to validate epic: %v", err))
	}

	// Format and write validation result
	return writeValidationResult(c, format, result, epicFile)
}

func writeValidationResult(c *cli.Command, format string, result *epic.ValidationResult, epicFile string) error {
	switch format {
	case "xml":
		output := fmt.Sprintf(`<validation_result epic="%s">
    <valid>%v</valid>`, getEpicName(epicFile), result.Valid)

		if len(result.Warnings) > 0 {
			output += `
    <warnings>`
			for _, warning := range result.Warnings {
				output += fmt.Sprintf(`
        <warning>%s</warning>`, warning)
			}
			output += `
    </warnings>`
		}

		if len(result.Errors) > 0 {
			output += `
    <errors>`
			for _, err := range result.Errors {
				output += fmt.Sprintf(`
        <error>%s</error>`, err)
			}
			output += `
    </errors>`
		}

		if len(result.Checks) > 0 {
			output += `
    <checks_performed>`
			for name, status := range result.Checks {
				output += fmt.Sprintf(`
        <check name="%s">%s</check>`, name, status)
			}
			output += `
    </checks_performed>`
		}

		output += fmt.Sprintf(`
    <message>%s</message>
</validation_result>`, result.Message())

		fmt.Fprint(c.Root().Writer, output)

	case "json":
		output := fmt.Sprintf(`{
  "epic": "%s",
  "valid": %v,
  "message": "%s"`, getEpicName(epicFile), result.Valid, result.Message())

		if len(result.Warnings) > 0 {
			output += `,
  "warnings": [`
			for i, warning := range result.Warnings {
				if i > 0 {
					output += `, `
				}
				output += fmt.Sprintf(`"%s"`, warning)
			}
			output += `]`
		}

		if len(result.Errors) > 0 {
			output += `,
  "errors": [`
			for i, err := range result.Errors {
				if i > 0 {
					output += `, `
				}
				output += fmt.Sprintf(`"%s"`, err)
			}
			output += `]`
		}

		if len(result.Checks) > 0 {
			output += `,
  "checks_performed": {`
			i := 0
			for name, status := range result.Checks {
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
		output := epic.FormatValidationResult(result, "text")
		fmt.Fprint(c.Root().Writer, output)
	}

	// Return error if validation failed
	if !result.Valid {
		return fmt.Errorf("validation failed")
	}

	return nil
}

// getEpicName extracts epic name/ID from file path for display
func getEpicName(filePath string) string {
	// Simple extraction - could be enhanced to read actual epic ID from file
	// For now, just use the filename without extension
	name := filePath
	if len(name) > 4 && name[len(name)-4:] == ".xml" {
		name = name[:len(name)-4]
	}

	// Extract just the filename from path
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '/' || name[i] == '\\' {
			name = name[i+1:]
			break
		}
	}

	return name
}
