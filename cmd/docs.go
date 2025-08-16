package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/reports"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/urfave/cli/v3"
)

// DocsCommand returns the docs command for generating human-readable documentation
func DocsCommand() *cli.Command {
	return &cli.Command{
		Name:   "docs",
		Usage:  "Generate human-readable documentation from epic data",
		Action: docsAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Override epic file from config",
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"F"},
				Usage:   "Output format: markdown (default), json",
				Value:   "markdown",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file path (default: stdout)",
			},
		},
	}
}

func docsAction(ctx context.Context, c *cli.Command) error {
	// Load configuration
	configPath := c.String("config")
	if configPath == "" {
		configPath = "./.agentpm.json"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine epic file (command flag overrides config)
	epicFile := c.String("file")
	if epicFile == "" {
		epicFile = cfg.CurrentEpic
	}
	if epicFile == "" {
		return fmt.Errorf("no epic file specified. Use --file flag or run 'agentpm init' first")
	}

	// Create storage and reports service
	storage := storage.NewFileStorage()
	reportsService := reports.NewReportService(storage)

	// Load epic
	err = reportsService.LoadEpic(epicFile)
	if err != nil {
		return fmt.Errorf("failed to load epic: %w", err)
	}

	// Generate documentation based on format
	outputFormat := c.String("format")
	outputFile := c.String("output")

	switch outputFormat {
	case "json":
		return generateJSONDocs(c, reportsService, outputFile)
	default:
		return generateMarkdownDocs(c, reportsService, outputFile)
	}
}

func generateMarkdownDocs(c *cli.Command, reportsService *reports.ReportService, outputFile string) error {
	markdown, err := reportsService.GenerateMarkdownDocumentation()
	if err != nil {
		return fmt.Errorf("failed to generate markdown documentation: %w", err)
	}

	if outputFile != "" {
		// Write to file
		err = writeToFile(outputFile, markdown)
		if err != nil {
			return fmt.Errorf("failed to write documentation to file: %w", err)
		}
		fmt.Fprintf(c.Root().Writer, "Documentation generated: %s\n", outputFile)
	} else {
		// Write to stdout
		fmt.Fprintf(c.Root().Writer, "%s", markdown)
	}

	return nil
}

func generateJSONDocs(c *cli.Command, reportsService *reports.ReportService, outputFile string) error {
	report, err := reportsService.GenerateDocumentationReport()
	if err != nil {
		return fmt.Errorf("failed to generate documentation report: %w", err)
	}

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal documentation to JSON: %w", err)
	}

	if outputFile != "" {
		// Write to file
		err = writeToFile(outputFile, string(jsonData))
		if err != nil {
			return fmt.Errorf("failed to write documentation to file: %w", err)
		}
		fmt.Fprintf(c.Root().Writer, "Documentation generated: %s\n", outputFile)
	} else {
		// Write to stdout
		fmt.Fprintf(c.Root().Writer, "%s\n", jsonData)
	}

	return nil
}

func writeToFile(filename, content string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write file
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	return nil
}
