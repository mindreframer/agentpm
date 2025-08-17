package cmd

import (
	"context"
	"fmt"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/xmlquery"
	"github.com/urfave/cli/v3"
)

// QueryCommand returns the query command for XPath-based epic file queries
func QueryCommand() *cli.Command {
	return &cli.Command{
		Name:      "query",
		Usage:     "Execute XPath queries against epic XML files",
		ArgsUsage: "<xpath-expression>",
		Description: `Execute XPath queries against epic XML files using etree syntax.

This command allows you to extract specific information from epic files using
XPath-like expressions. It supports element selection, attribute filtering,
and complex path navigation.

Query patterns:
  //task                          - All task elements
  //phase[@status='active']       - Phases with specific status
  //task[@phase_id='1A']          - Tasks in specific phase
  //metadata/assignee             - Nested elements
  //test[@id='test_1']            - Elements by ID
  //description/text()            - Text content
  //phase/@name                   - Attribute values
  //task[1]                       - Position-based selection
  //epic/*                        - All child elements

Output formats: xml (default), text, json

Examples:
  agentpm query "//task"                          # All tasks
  agentpm query "//task[@status='done']"         # Completed tasks
  agentpm query "//task[@phase_id='1A']"         # Tasks in phase 1A
  agentpm query "//metadata/assignee"            # Epic assignee
  agentpm query "//phase[@status='active']"      # Active phases
  agentpm query "//test[@status='passing']"      # Passing tests
  agentpm query "//task[@status='done']" --format text  # Text output
  agentpm query "//phase" -f epic-9.xml          # Query different file`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Override epic file from config",
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"F"},
				Usage:   "Output format: xml (default), text, json",
				Value:   "xml",
			},
		},
		Action: queryAction,
	}
}

func queryAction(ctx context.Context, c *cli.Command) error {
	if c.Args().Len() < 1 {
		return fmt.Errorf("XPath expression is required")
	}

	xpathExpr := c.Args().First()

	// Determine epic file (prioritize command flag)
	epicFile := c.String("file")

	if epicFile == "" {
		// Only load config if file is not explicitly provided
		configPath := c.String("config")
		if configPath == "" {
			configPath = "./.agentpm.json"
		}

		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		epicFile = cfg.CurrentEpic
		if epicFile == "" {
			return fmt.Errorf("no epic file specified. Use --file flag or run 'agentpm init' first")
		}
	}

	// Get output format
	outputFormat := c.String("format")
	var format xmlquery.OutputFormat
	switch outputFormat {
	case "xml":
		format = xmlquery.FormatXML
	case "text":
		format = xmlquery.FormatText
	case "json":
		format = xmlquery.FormatJSON
	default:
		return fmt.Errorf("invalid output format: %s (must be xml, text, or json)", outputFormat)
	}

	// Create query service
	service := xmlquery.NewService()

	// Validate query syntax first
	if err := service.ValidateQuery(xpathExpr); err != nil {
		return fmt.Errorf("invalid XPath query: %w", err)
	}

	// Execute query with formatting
	output, err := service.QueryEpicFileFormatted(epicFile, xpathExpr, format)
	if err != nil {
		return fmt.Errorf("query execution failed: %w", err)
	}

	// Output result
	fmt.Fprint(c.Root().Writer, output)
	if outputFormat == "text" && output != "" && output[len(output)-1:] != "\n" {
		fmt.Fprint(c.Root().Writer, "\n")
	}

	return nil
}
