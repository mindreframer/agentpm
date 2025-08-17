package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mindreframer/agentpm/internal/config"
	contextpkg "github.com/mindreframer/agentpm/internal/context"
	"github.com/mindreframer/agentpm/internal/epic"
	"github.com/mindreframer/agentpm/internal/query"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/urfave/cli/v3"
)

// ShowCommand returns the show command for unified entity inspection
func ShowCommand() *cli.Command {
	return &cli.Command{
		Name:      "show",
		Usage:     "Display detailed information about epic entities",
		ArgsUsage: "<entity-type> [entity-id]",
		Description: `Display detailed information about epic entities.

Entity types:
  epic        - Show complete epic information (no ID required)
  phase <id>  - Show specific phase details
  task <id>   - Show specific task details  
  test <id>   - Show specific test details

Output formats: text (default), json, xml

The --full flag provides comprehensive context with complete details for all related entities:
- For tasks: Shows parent phase, sibling tasks, and child tests with full details
- For phases: Shows all tasks and tests in the phase with complete information
- For tests: Shows parent task, parent phase, and sibling tests with full details

Examples:
  agentpm show epic                          # Show complete epic
  agentpm show phase 1A                     # Show phase 1A details
  agentpm show task 2B_T1 --full            # Show task with full context
  agentpm show test 3A_T1 --full            # Show test with full context
  agentpm show phase 1A --format json       # Show phase in JSON format
  agentpm show task 1A_1 --full --format xml # Show task context in XML format`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Override epic file from config",
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"F"},
				Usage:   "Output format: text (default), json, xml",
				Value:   "text",
			},
			&cli.BoolFlag{
				Name:  "full",
				Usage: "Display full context with complete details for all related entities",
			},
		},
		Action: showAction,
	}
}

func showAction(ctx context.Context, c *cli.Command) error {
	if c.Args().Len() < 1 {
		return fmt.Errorf("entity type is required (epic, phase, task, test)")
	}

	entityType := c.Args().First()
	var entityID string
	if c.Args().Len() >= 2 {
		entityID = c.Args().Get(1)
	}

	// Validate arguments
	switch entityType {
	case "epic":
		// Epic doesn't need an ID
	case "phase", "task", "test":
		if entityID == "" {
			return fmt.Errorf("%s requires an ID", entityType)
		}
	default:
		return fmt.Errorf("invalid entity type: %s (must be epic, phase, task, or test)", entityType)
	}

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

	// Create storage and query service
	storage := storage.NewFileStorage()
	queryService := query.NewQueryService(storage)

	// Load epic
	err = queryService.LoadEpic(epicFile)
	if err != nil {
		return fmt.Errorf("failed to load epic: %w", err)
	}

	// Get output format and full flag
	outputFormat := c.String("format")
	useFullContext := c.Bool("full")

	// Handle entity display
	switch entityType {
	case "epic":
		return showEpic(c, queryService, outputFormat)
	case "phase":
		return showPhase(c, queryService, entityID, outputFormat, useFullContext)
	case "task":
		return showTask(c, queryService, entityID, outputFormat, useFullContext)
	case "test":
		return showTest(c, queryService, entityID, outputFormat, useFullContext)
	}

	return nil
}

func showEpic(c *cli.Command, qs *query.QueryService, format string) error {
	epic, err := qs.GetEpic()
	if err != nil {
		return fmt.Errorf("failed to get epic: %w", err)
	}

	switch format {
	case "json":
		return outputEpicJSON(c, epic)
	case "xml":
		return outputEpicXML(c, epic)
	default:
		return outputEpicText(c, epic)
	}
}

func showPhase(c *cli.Command, qs *query.QueryService, phaseID, format string, useFullContext bool) error {
	if useFullContext {
		// Use context engine for full context display
		engine := contextpkg.NewEngine(qs)
		ctx, err := engine.GetPhaseContext(phaseID, true)
		if err != nil {
			return fmt.Errorf("failed to get phase context: %w", err)
		}

		formatter := contextpkg.NewFormatter(format)
		return formatter.FormatPhaseContext(ctx, c.Root().Writer)
	}

	// Use original implementation for backward compatibility
	phase, err := qs.GetPhase(phaseID)
	if err != nil {
		return err
	}

	// Get related items for context
	related, err := qs.GetRelatedItems("phase", phaseID)
	if err != nil {
		return fmt.Errorf("failed to get related items: %w", err)
	}

	switch format {
	case "json":
		return outputPhaseJSON(c, phase, related)
	case "xml":
		return outputPhaseXML(c, phase, related)
	default:
		return outputPhaseText(c, phase, related)
	}
}

func showTask(c *cli.Command, qs *query.QueryService, taskID, format string, useFullContext bool) error {
	if useFullContext {
		// Use context engine for full context display
		engine := contextpkg.NewEngine(qs)
		ctx, err := engine.GetTaskContext(taskID, true)
		if err != nil {
			return fmt.Errorf("failed to get task context: %w", err)
		}

		formatter := contextpkg.NewFormatter(format)
		return formatter.FormatTaskContext(ctx, c.Root().Writer)
	}

	// Use original implementation for backward compatibility
	task, err := qs.GetTask(taskID)
	if err != nil {
		return err
	}

	// Get related items for context
	related, err := qs.GetRelatedItems("task", taskID)
	if err != nil {
		return fmt.Errorf("failed to get related items: %w", err)
	}

	switch format {
	case "json":
		return outputTaskJSON(c, task, related)
	case "xml":
		return outputTaskXML(c, task, related)
	default:
		return outputTaskText(c, task, related)
	}
}

func showTest(c *cli.Command, qs *query.QueryService, testID, format string, useFullContext bool) error {
	if useFullContext {
		// Use context engine for full context display
		engine := contextpkg.NewEngine(qs)
		ctx, err := engine.GetTestContext(testID, true)
		if err != nil {
			return fmt.Errorf("failed to get test context: %w", err)
		}

		formatter := contextpkg.NewFormatter(format)
		return formatter.FormatTestContext(ctx, c.Root().Writer)
	}

	// Use original implementation for backward compatibility
	test, err := qs.GetTest(testID)
	if err != nil {
		return err
	}

	// Get related items for context
	related, err := qs.GetRelatedItems("test", testID)
	if err != nil {
		return fmt.Errorf("failed to get related items: %w", err)
	}

	switch format {
	case "json":
		return outputTestJSON(c, test, related)
	case "xml":
		return outputTestXML(c, test, related)
	default:
		return outputTestText(c, test, related)
	}
}

// Epic output functions
func outputEpicText(c *cli.Command, epic *epic.Epic) error {
	fmt.Fprintf(c.Root().Writer, "Epic: %s\n", epic.Name)
	fmt.Fprintf(c.Root().Writer, "ID: %s\n", epic.ID)
	fmt.Fprintf(c.Root().Writer, "Status: %s\n", epic.Status)
	if epic.Description != "" {
		fmt.Fprintf(c.Root().Writer, "Description: %s\n", epic.Description)
	}

	fmt.Fprintf(c.Root().Writer, "\nPhases (%d):\n", len(epic.Phases))
	for _, phase := range epic.Phases {
		fmt.Fprintf(c.Root().Writer, "  %s - %s [%s]\n", phase.ID, phase.Name, phase.Status)
	}

	fmt.Fprintf(c.Root().Writer, "\nTasks (%d):\n", len(epic.Tasks))
	for _, task := range epic.Tasks {
		fmt.Fprintf(c.Root().Writer, "  %s (%s) - %s [%s]\n", task.ID, task.PhaseID, task.Name, task.Status)
	}

	fmt.Fprintf(c.Root().Writer, "\nTests (%d):\n", len(epic.Tests))
	for _, test := range epic.Tests {
		fmt.Fprintf(c.Root().Writer, "  %s (%s) - %s [%s]\n", test.ID, test.TaskID, test.Name, test.Status)
	}

	return nil
}

func outputEpicJSON(c *cli.Command, epic *epic.Epic) error {
	jsonData, err := json.MarshalIndent(epic, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal epic to JSON: %w", err)
	}
	fmt.Fprintf(c.Root().Writer, "%s\n", jsonData)
	return nil
}

func outputEpicXML(c *cli.Command, epic *epic.Epic) error {
	fmt.Fprintf(c.Root().Writer, "<epic id=\"%s\" status=\"%s\">\n", epic.ID, epic.Status)
	fmt.Fprintf(c.Root().Writer, "    <name>%s</name>\n", epic.Name)
	if epic.Description != "" {
		fmt.Fprintf(c.Root().Writer, "    <description>%s</description>\n", epic.Description)
	}

	fmt.Fprintf(c.Root().Writer, "    <phases>\n")
	for _, phase := range epic.Phases {
		fmt.Fprintf(c.Root().Writer, "        <phase id=\"%s\" status=\"%s\">\n", phase.ID, phase.Status)
		fmt.Fprintf(c.Root().Writer, "            <name>%s</name>\n", phase.Name)
		if phase.Description != "" {
			fmt.Fprintf(c.Root().Writer, "            <description>%s</description>\n", phase.Description)
		}
		fmt.Fprintf(c.Root().Writer, "        </phase>\n")
	}
	fmt.Fprintf(c.Root().Writer, "    </phases>\n")

	fmt.Fprintf(c.Root().Writer, "    <tasks>\n")
	for _, task := range epic.Tasks {
		fmt.Fprintf(c.Root().Writer, "        <task id=\"%s\" phase_id=\"%s\" status=\"%s\">\n", task.ID, task.PhaseID, task.Status)
		fmt.Fprintf(c.Root().Writer, "            <name>%s</name>\n", task.Name)
		if task.Description != "" {
			fmt.Fprintf(c.Root().Writer, "            <description>%s</description>\n", task.Description)
		}
		fmt.Fprintf(c.Root().Writer, "        </task>\n")
	}
	fmt.Fprintf(c.Root().Writer, "    </tasks>\n")

	fmt.Fprintf(c.Root().Writer, "    <tests>\n")
	for _, test := range epic.Tests {
		fmt.Fprintf(c.Root().Writer, "        <test id=\"%s\" task_id=\"%s\" status=\"%s\">\n", test.ID, test.TaskID, test.Status)
		fmt.Fprintf(c.Root().Writer, "            <name>%s</name>\n", test.Name)
		if test.Description != "" {
			fmt.Fprintf(c.Root().Writer, "            <description>%s</description>\n", test.Description)
		}
		fmt.Fprintf(c.Root().Writer, "        </test>\n")
	}
	fmt.Fprintf(c.Root().Writer, "    </tests>\n")

	fmt.Fprintf(c.Root().Writer, "</epic>\n")
	return nil
}

// Phase output functions
func outputPhaseText(c *cli.Command, phase *epic.Phase, related []query.RelatedItem) error {
	fmt.Fprintf(c.Root().Writer, "Phase: %s\n", phase.Name)
	fmt.Fprintf(c.Root().Writer, "ID: %s\n", phase.ID)
	fmt.Fprintf(c.Root().Writer, "Status: %s\n", phase.Status)
	if phase.Description != "" {
		fmt.Fprintf(c.Root().Writer, "Description: %s\n", phase.Description)
	}

	// Show related tasks
	var tasks []query.RelatedItem
	for _, item := range related {
		if item.Type == "task" {
			tasks = append(tasks, item)
		}
	}

	fmt.Fprintf(c.Root().Writer, "\nTasks (%d):\n", len(tasks))
	if len(tasks) == 0 {
		fmt.Fprintf(c.Root().Writer, "  (none)\n")
	} else {
		for _, task := range tasks {
			fmt.Fprintf(c.Root().Writer, "  %s - %s\n", task.ID, task.Name)
		}
	}

	// Show related tests
	var tests []query.RelatedItem
	for _, item := range related {
		if item.Type == "test" {
			tests = append(tests, item)
		}
	}

	fmt.Fprintf(c.Root().Writer, "\nTests (%d):\n", len(tests))
	if len(tests) == 0 {
		fmt.Fprintf(c.Root().Writer, "  (none)\n")
	} else {
		for _, test := range tests {
			fmt.Fprintf(c.Root().Writer, "  %s - %s\n", test.ID, test.Name)
		}
	}

	return nil
}

func outputPhaseJSON(c *cli.Command, phase *epic.Phase, related []query.RelatedItem) error {
	output := map[string]interface{}{
		"id":          phase.ID,
		"name":        phase.Name,
		"status":      phase.Status,
		"description": phase.Description,
		"related":     related,
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal phase to JSON: %w", err)
	}
	fmt.Fprintf(c.Root().Writer, "%s\n", jsonData)
	return nil
}

func outputPhaseXML(c *cli.Command, phase *epic.Phase, related []query.RelatedItem) error {
	fmt.Fprintf(c.Root().Writer, "<phase id=\"%s\" status=\"%s\">\n", phase.ID, phase.Status)
	fmt.Fprintf(c.Root().Writer, "    <name>%s</name>\n", phase.Name)
	if phase.Description != "" {
		fmt.Fprintf(c.Root().Writer, "    <description>%s</description>\n", phase.Description)
	}

	fmt.Fprintf(c.Root().Writer, "    <related>\n")
	for _, item := range related {
		fmt.Fprintf(c.Root().Writer, "        <%s id=\"%s\" relationship=\"%s\">%s</%s>\n",
			item.Type, item.ID, item.Relationship, item.Name, item.Type)
	}
	fmt.Fprintf(c.Root().Writer, "    </related>\n")

	fmt.Fprintf(c.Root().Writer, "</phase>\n")
	return nil
}

// Task output functions
func outputTaskText(c *cli.Command, task *epic.Task, related []query.RelatedItem) error {
	fmt.Fprintf(c.Root().Writer, "Task: %s\n", task.Name)
	fmt.Fprintf(c.Root().Writer, "ID: %s\n", task.ID)
	fmt.Fprintf(c.Root().Writer, "Phase: %s\n", task.PhaseID)
	fmt.Fprintf(c.Root().Writer, "Status: %s\n", task.Status)
	if task.Description != "" {
		fmt.Fprintf(c.Root().Writer, "Description: %s\n", task.Description)
	}

	// Show parent phase
	for _, item := range related {
		if item.Type == "phase" && item.Relationship == "parent" {
			fmt.Fprintf(c.Root().Writer, "Parent Phase: %s - %s\n", item.ID, item.Name)
			break
		}
	}

	// Show related tests
	var tests []query.RelatedItem
	for _, item := range related {
		if item.Type == "test" {
			tests = append(tests, item)
		}
	}

	fmt.Fprintf(c.Root().Writer, "\nTests (%d):\n", len(tests))
	if len(tests) == 0 {
		fmt.Fprintf(c.Root().Writer, "  (none)\n")
	} else {
		for _, test := range tests {
			fmt.Fprintf(c.Root().Writer, "  %s - %s\n", test.ID, test.Name)
		}
	}

	return nil
}

func outputTaskJSON(c *cli.Command, task *epic.Task, related []query.RelatedItem) error {
	output := map[string]interface{}{
		"id":          task.ID,
		"phase_id":    task.PhaseID,
		"name":        task.Name,
		"status":      task.Status,
		"description": task.Description,
		"related":     related,
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal task to JSON: %w", err)
	}
	fmt.Fprintf(c.Root().Writer, "%s\n", jsonData)
	return nil
}

func outputTaskXML(c *cli.Command, task *epic.Task, related []query.RelatedItem) error {
	fmt.Fprintf(c.Root().Writer, "<task id=\"%s\" phase_id=\"%s\" status=\"%s\">\n", task.ID, task.PhaseID, task.Status)
	fmt.Fprintf(c.Root().Writer, "    <name>%s</name>\n", task.Name)
	if task.Description != "" {
		fmt.Fprintf(c.Root().Writer, "    <description>%s</description>\n", task.Description)
	}

	fmt.Fprintf(c.Root().Writer, "    <related>\n")
	for _, item := range related {
		fmt.Fprintf(c.Root().Writer, "        <%s id=\"%s\" relationship=\"%s\">%s</%s>\n",
			item.Type, item.ID, item.Relationship, item.Name, item.Type)
	}
	fmt.Fprintf(c.Root().Writer, "    </related>\n")

	fmt.Fprintf(c.Root().Writer, "</task>\n")
	return nil
}

// Test output functions
func outputTestText(c *cli.Command, test *epic.Test, related []query.RelatedItem) error {
	fmt.Fprintf(c.Root().Writer, "Test: %s\n", test.Name)
	fmt.Fprintf(c.Root().Writer, "ID: %s\n", test.ID)
	fmt.Fprintf(c.Root().Writer, "Task: %s\n", test.TaskID)
	fmt.Fprintf(c.Root().Writer, "Status: %s\n", test.Status)
	if test.Description != "" {
		fmt.Fprintf(c.Root().Writer, "Description: %s\n", test.Description)
	}

	// Show parent task and phase
	for _, item := range related {
		if item.Type == "task" && item.Relationship == "parent" {
			fmt.Fprintf(c.Root().Writer, "Parent Task: %s - %s\n", item.ID, item.Name)
		}
		if item.Type == "phase" && item.Relationship == "ancestor" {
			fmt.Fprintf(c.Root().Writer, "Parent Phase: %s - %s\n", item.ID, item.Name)
		}
	}

	return nil
}

func outputTestJSON(c *cli.Command, test *epic.Test, related []query.RelatedItem) error {
	output := map[string]interface{}{
		"id":          test.ID,
		"task_id":     test.TaskID,
		"name":        test.Name,
		"status":      test.Status,
		"description": test.Description,
		"related":     related,
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal test to JSON: %w", err)
	}
	fmt.Fprintf(c.Root().Writer, "%s\n", jsonData)
	return nil
}

func outputTestXML(c *cli.Command, test *epic.Test, related []query.RelatedItem) error {
	fmt.Fprintf(c.Root().Writer, "<test id=\"%s\" task_id=\"%s\" status=\"%s\">\n", test.ID, test.TaskID, test.Status)
	fmt.Fprintf(c.Root().Writer, "    <name>%s</name>\n", test.Name)
	if test.Description != "" {
		fmt.Fprintf(c.Root().Writer, "    <description>%s</description>\n", test.Description)
	}

	fmt.Fprintf(c.Root().Writer, "    <related>\n")
	for _, item := range related {
		fmt.Fprintf(c.Root().Writer, "        <%s id=\"%s\" relationship=\"%s\">%s</%s>\n",
			item.Type, item.ID, item.Relationship, item.Name, item.Type)
	}
	fmt.Fprintf(c.Root().Writer, "    </related>\n")

	fmt.Fprintf(c.Root().Writer, "</test>\n")
	return nil
}
