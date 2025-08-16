package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/beevik/etree"
	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/storage"
	"github.com/urfave/cli/v3"
)

func SwitchCommand() *cli.Command {
	return &cli.Command{
		Name:    "switch",
		Aliases: []string{"sw"},
		Usage:   "Switch to a different epic file",
		Description: `Switch the current epic context to a different epic file.

This command:
- Updates the configuration to point to a new epic file
- Validates that the target epic file exists and is valid
- Tracks the previous epic for easy switching back
- Maintains project context across epic switches

Examples:
  agentpm switch epic-5.xml           # Switch to epic-5.xml
  agentpm switch --back               # Switch back to previous epic
  agentpm switch /path/to/epic.xml    # Switch to epic with absolute path`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "back",
				Aliases: []string{"b"},
				Usage:   "Switch back to the previous epic",
			},
		},
		Action: switchAction,
	}
}

func switchAction(ctx context.Context, c *cli.Command) error {
	// Get global flags
	configPath := c.String("config")
	if configPath == "" {
		configPath = "./.agentpm.json"
	}
	format := c.String("format")

	// Check if we should switch back
	switchBack := c.Bool("back")

	// Get target epic file from args (unless switching back)
	var targetEpic string
	if !switchBack {
		if c.Args().Len() == 0 {
			return fmt.Errorf("target epic file is required (use --back to switch to previous epic)")
		}
		targetEpic = c.Args().First()
	}

	// Load current configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Handle switch back operation
	if switchBack {
		return handleSwitchBack(c, cfg, configPath, format)
	}

	// Handle switch to new epic
	return handleSwitchToEpic(c, cfg, configPath, targetEpic, format)
}

func handleSwitchBack(c *cli.Command, cfg *config.Config, configPath, format string) error {
	// Check if we have a previous epic to switch back to
	if cfg.PreviousEpic == "" {
		return fmt.Errorf("no previous epic to switch back to")
	}

	// Validate previous epic file still exists
	var previousPath string
	if filepath.IsAbs(cfg.PreviousEpic) {
		previousPath = cfg.PreviousEpic
	} else {
		previousPath = filepath.Join(".", cfg.PreviousEpic)
	}

	if _, err := os.Stat(previousPath); os.IsNotExist(err) {
		return fmt.Errorf("previous epic file no longer exists: %s", previousPath)
	}

	// Validate previous epic file is still valid
	if err := validateEpicFile(previousPath); err != nil {
		return fmt.Errorf("previous epic file is no longer valid: %w", err)
	}

	// Swap current and previous epics
	currentEpic := cfg.CurrentEpic
	cfg.CurrentEpic = cfg.PreviousEpic
	cfg.PreviousEpic = currentEpic

	// Save updated configuration
	if err := config.SaveConfig(cfg, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Create switch result
	result := &SwitchResult{
		PreviousEpic: currentEpic,
		NewEpic:      cfg.CurrentEpic,
		EpicPath:     previousPath,
		Message:      fmt.Sprintf("Switched back from %s to %s", currentEpic, cfg.CurrentEpic),
	}

	// Output the result
	return outputSwitchResult(c, result, format)
}

func handleSwitchToEpic(c *cli.Command, cfg *config.Config, configPath, targetEpic, format string) error {
	// Resolve target epic path
	var targetPath string
	if filepath.IsAbs(targetEpic) {
		targetPath = targetEpic
	} else {
		targetPath = filepath.Join(".", targetEpic)
	}

	// Validate target epic file exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("epic file does not exist: %s", targetPath)
	}

	// Validate target epic file is valid
	if err := validateEpicFile(targetPath); err != nil {
		return fmt.Errorf("invalid epic file: %w", err)
	}

	// Store current epic as previous (for future switch back)
	previousEpic := cfg.CurrentEpic

	// Update configuration
	cfg.PreviousEpic = cfg.CurrentEpic
	cfg.CurrentEpic = targetEpic

	// Save updated configuration
	if err := config.SaveConfig(cfg, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Create switch result
	result := &SwitchResult{
		PreviousEpic: previousEpic,
		NewEpic:      targetEpic,
		EpicPath:     targetPath,
		Message:      fmt.Sprintf("Switched from %s to %s", previousEpic, targetEpic),
	}

	// Output the result
	return outputSwitchResult(c, result, format)
}

func validateEpicFile(epicPath string) error {
	// Initialize storage to validate the epic file
	storageFactory := storage.NewFactory(false) // Use file storage
	storageImpl := storageFactory.CreateStorage()

	// Try to load the epic to validate it
	_, err := storageImpl.LoadEpic(epicPath)
	if err != nil {
		return fmt.Errorf("failed to load epic: %w", err)
	}

	return nil
}

type SwitchResult struct {
	PreviousEpic string
	NewEpic      string
	EpicPath     string
	Message      string
}

func outputSwitchResult(c *cli.Command, result *SwitchResult, format string) error {
	switch format {
	case "json":
		return outputSwitchResultJSON(c, result)
	case "xml":
		return outputSwitchResultXML(c, result)
	default:
		return outputSwitchResultText(c, result)
	}
}

func outputSwitchResultText(c *cli.Command, result *SwitchResult) error {
	fmt.Fprintf(c.Root().Writer, "Epic switched successfully\n")
	fmt.Fprintf(c.Root().Writer, "Previous: %s\n", result.PreviousEpic)
	fmt.Fprintf(c.Root().Writer, "Current: %s\n", result.NewEpic)
	fmt.Fprintf(c.Root().Writer, "Path: %s\n", result.EpicPath)
	fmt.Fprintf(c.Root().Writer, "\n%s\n", result.Message)
	return nil
}

func outputSwitchResultJSON(c *cli.Command, result *SwitchResult) error {
	output := map[string]interface{}{
		"epic_switched": map[string]interface{}{
			"previous_epic": result.PreviousEpic,
			"new_epic":      result.NewEpic,
			"epic_path":     result.EpicPath,
			"message":       result.Message,
		},
	}

	encoder := json.NewEncoder(c.Root().Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputSwitchResultXML(c *cli.Command, result *SwitchResult) error {
	doc := etree.NewDocument()
	root := doc.CreateElement("epic_switched")
	root.SetText("\n    ")

	previousEpic := root.CreateElement("previous_epic")
	previousEpic.SetText(result.PreviousEpic)

	newEpic := root.CreateElement("new_epic")
	newEpic.SetText(result.NewEpic)

	epicPath := root.CreateElement("epic_path")
	epicPath.SetText(result.EpicPath)

	message := root.CreateElement("message")
	message.SetText(result.Message)

	doc.Indent(4)
	doc.WriteTo(c.Root().Writer)
	fmt.Fprintf(c.Root().Writer, "\n") // Add newline
	return nil
}
