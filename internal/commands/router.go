package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/beevik/etree"
	"github.com/mindreframer/agentpm/internal/messages"
	"github.com/urfave/cli/v3"
)

// EntityType represents the type of entity (phase, task, test)
type EntityType string

const (
	EntityTypePhase EntityType = "phase"
	EntityTypeTask  EntityType = "task"
	EntityTypeTest  EntityType = "test"
	EntityTypeEpic  EntityType = "epic"
)

// String returns the string representation of EntityType
func (et EntityType) String() string {
	return string(et)
}

// RouterContext holds common data needed by unified commands
type RouterContext struct {
	ConfigPath string
	EpicFile   string
	Format     string
	Time       string
}

// ExtractRouterContext extracts common flags from a CLI command
func ExtractRouterContext(c *cli.Command) RouterContext {
	return RouterContext{
		ConfigPath: c.String("config"),
		EpicFile:   c.String("file"),
		Format:     c.String("format"),
		Time:       c.String("time"),
	}
}

// EntityID represents an entity identifier with its detected type
type EntityID struct {
	ID   string
	Type EntityType
}

// TypeDetectionResult holds the result of entity type detection
type TypeDetectionResult struct {
	EntityID    *EntityID
	IsAmbiguous bool
	Suggestions []EntityID
	Error       error
}

// DetectEntityType attempts to determine the entity type from an ID
func DetectEntityType(id string) TypeDetectionResult {
	if id == "" {
		return TypeDetectionResult{
			Error: fmt.Errorf("entity ID cannot be empty"),
		}
	}

	// Define patterns for different entity types (order matters - more specific first)
	patterns := []struct {
		Type    EntityType
		Pattern *regexp.Regexp
	}{
		{EntityTypeTest, regexp.MustCompile(`^[0-9]+[A-Z]_T[0-9]+$`)}, // e.g., "3A_T1", "1B_T2" (most specific)
		{EntityTypeTask, regexp.MustCompile(`^[0-9]+[A-Z]_[0-9]+$`)},  // e.g., "3A_1", "1B_2" (no T prefix)
		{EntityTypePhase, regexp.MustCompile(`^[0-9]+[A-Z]$`)},        // e.g., "3A", "1B", "4C"
	}

	var matches []EntityID
	for _, patternDef := range patterns {
		if patternDef.Pattern.MatchString(id) {
			matches = append(matches, EntityID{
				ID:   id,
				Type: patternDef.Type,
			})
		}
	}

	switch len(matches) {
	case 0:
		return TypeDetectionResult{
			Error: fmt.Errorf("unable to determine entity type for ID '%s': does not match any known patterns", id),
		}
	case 1:
		return TypeDetectionResult{
			EntityID: &matches[0],
		}
	default:
		return TypeDetectionResult{
			IsAmbiguous: true,
			Suggestions: matches,
			Error:       fmt.Errorf("ambiguous entity ID '%s': could be %s", id, formatEntityTypes(matches)),
		}
	}
}

// formatEntityTypes formats a list of entity types for error messages
func formatEntityTypes(entities []EntityID) string {
	var types []string
	for _, entity := range entities {
		types = append(types, string(entity.Type))
	}
	return strings.Join(types, " or ")
}

// ValidateSubcommandArgs validates arguments for subcommand-based operations
func ValidateSubcommandArgs(subcommand string, args []string, requiredArgCount int) error {
	if len(args) != requiredArgCount {
		if requiredArgCount == 1 {
			return fmt.Errorf("%s requires exactly one argument", subcommand)
		}
		return fmt.Errorf("%s requires exactly %d arguments", subcommand, requiredArgCount)
	}
	return nil
}

// Output formatting utilities for unified commands

// OutputResult handles outputting results in different formats
func OutputResult(c *cli.Command, format string, data any) error {
	switch format {
	case "json":
		return OutputJSON(c, data)
	case "xml":
		return OutputXML(c, data)
	default:
		return OutputText(c, data)
	}
}

// OutputJSON outputs data as JSON
func OutputJSON(c *cli.Command, data any) error {
	encoder := json.NewEncoder(c.Root().Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// OutputXML outputs data as XML
func OutputXML(c *cli.Command, data any) error {
	// Create XML output based on data type
	doc := etree.NewDocument()

	switch v := data.(type) {
	case string:
		root := doc.CreateElement("result")
		root.SetText(v)
	case map[string]any:
		root := doc.CreateElement("result")
		for key, value := range v {
			elem := root.CreateElement(key)
			elem.SetText(fmt.Sprintf("%v", value))
		}
	default:
		root := doc.CreateElement("result")
		root.SetText(fmt.Sprintf("%v", data))
	}

	doc.Indent(4)
	doc.WriteTo(c.Root().Writer)
	fmt.Fprintf(c.Root().Writer, "\n")
	return nil
}

// OutputText outputs data as plain text
func OutputText(c *cli.Command, data any) error {
	switch v := data.(type) {
	case string:
		fmt.Fprintf(c.Root().Writer, "%s\n", v)
	case map[string]any:
		for key, value := range v {
			fmt.Fprintf(c.Root().Writer, "%s: %v\n", key, value)
		}
	default:
		fmt.Fprintf(c.Root().Writer, "%v\n", data)
	}
	return nil
}

// Error handling for unified commands

// OutputError handles outputting errors in different formats
func OutputError(c *cli.Command, format string, err error) error {
	switch format {
	case "json":
		return OutputErrorJSON(c, err)
	case "xml":
		return OutputErrorXML(c, err)
	default:
		return OutputErrorText(c, err)
	}
}

// OutputErrorJSON outputs error as JSON
func OutputErrorJSON(c *cli.Command, err error) error {
	output := map[string]any{
		"error": map[string]any{
			"message": err.Error(),
		},
	}

	encoder := json.NewEncoder(c.Root().ErrWriter)
	encoder.SetIndent("", "  ")
	encoder.Encode(output)
	return err
}

// OutputErrorXML outputs error as XML
func OutputErrorXML(c *cli.Command, err error) error {
	doc := etree.NewDocument()
	root := doc.CreateElement("error")

	message := root.CreateElement("message")
	message.SetText(err.Error())

	doc.Indent(4)
	doc.WriteTo(c.Root().ErrWriter)
	fmt.Fprintf(c.Root().ErrWriter, "\n")
	return err
}

// OutputErrorText outputs error as plain text
func OutputErrorText(c *cli.Command, err error) error {
	fmt.Fprintf(c.Root().ErrWriter, "Error: %v\n", err)
	return err
}

// OutputFriendlyMessage handles outputting friendly messages
func OutputFriendlyMessage(c *cli.Command, message *messages.Message, format string) error {
	formatter := messages.NewMessageFormatter()
	switch format {
	case "json":
		output, err := formatter.FormatJSON(message)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.Root().Writer, "%s\n", output)
	case "xml":
		output, err := formatter.FormatXML(message)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.Root().Writer, "%s\n", output)
	default:
		output := formatter.FormatText(message)
		fmt.Fprintf(c.Root().Writer, "%s\n", output)
	}
	return nil // Success - this is a friendly message, not an error
}

// Subcommand action creators

// CreateEpicAction creates an action function for epic-level operations (no ID needed)
func CreateEpicAction(handler func(RouterContext) error) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		routerCtx := ExtractRouterContext(c)
		return handler(routerCtx)
	}
}

// CreateEntityAction creates an action function for entity-level operations (ID required)
func CreateEntityAction(entityType EntityType, handler func(RouterContext, string) error) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() < 1 {
			return fmt.Errorf("%s ID is required", entityType)
		}

		entityID := c.Args().First()
		routerCtx := ExtractRouterContext(c)
		return handler(routerCtx, entityID)
	}
}

// CreateAutoDetectAction creates an action that auto-detects entity type
func CreateAutoDetectAction(handlers map[EntityType]func(RouterContext, string) error) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if c.Args().Len() < 1 {
			return fmt.Errorf("entity ID is required")
		}

		entityID := c.Args().First()
		routerCtx := ExtractRouterContext(c)

		// Attempt type detection
		result := DetectEntityType(entityID)
		if result.Error != nil {
			return OutputError(c, routerCtx.Format, result.Error)
		}

		if result.IsAmbiguous {
			return OutputError(c, routerCtx.Format, result.Error)
		}

		// Find appropriate handler
		handler, exists := handlers[result.EntityID.Type]
		if !exists {
			return OutputError(c, routerCtx.Format, fmt.Errorf("no handler available for entity type: %s", result.EntityID.Type))
		}

		return handler(routerCtx, entityID)
	}
}

// Global flag definitions for unified commands
func GlobalFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Usage:   "Override epic file from config",
		},
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Override config file path",
			Value:   "./.agentpm.json",
		},
		&cli.StringFlag{
			Name:    "time",
			Aliases: []string{"t"},
			Usage:   "Timestamp for current time (testing support)",
		},
		&cli.StringFlag{
			Name:    "format",
			Aliases: []string{"F"},
			Usage:   "Output format - text (default) / json / xml",
			Value:   "text",
		},
	}
}
