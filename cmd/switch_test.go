package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mindreframer/agentpm/internal/config"
	"github.com/mindreframer/agentpm/internal/epic"
	apmtesting "github.com/mindreframer/agentpm/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestSwitchCommand_Success(t *testing.T) {
	// Create temporary directory and files
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".agentpm.json")
	currentEpicFile := filepath.Join(tempDir, "current-epic.xml")
	targetEpicFile := filepath.Join(tempDir, "target-epic.xml")

	// Create test config with current epic
	cfg := &config.Config{
		CurrentEpic: currentEpicFile,
	}
	require.NoError(t, config.SaveConfig(cfg, configFile))

	// Create current epic file
	currentEpic := &epic.Epic{
		ID:     "current-epic",
		Name:   "Current Epic",
		Status: epic.StatusActive,
	}
	writeTestEpicXML(t, currentEpicFile, currentEpic)

	// Create target epic file
	targetEpic := &epic.Epic{
		ID:     "target-epic",
		Name:   "Target Epic",
		Status: epic.StatusPlanning,
	}
	writeTestEpicXML(t, targetEpicFile, targetEpic)

	// Create CLI app with switch command
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Value: configFile},
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			SwitchCommand(),
		},
	}

	// Capture output
	var stdout bytes.Buffer
	app.Writer = &stdout

	// Run switch command
	args := []string{"agentpm", "switch", targetEpicFile}
	err := app.Run(context.Background(), args)

	// Verify success
	require.NoError(t, err)

	// Check output contains expected content
	output := stdout.String()
	assert.Contains(t, output, "Epic switched successfully")
	assert.Contains(t, output, "Previous: "+currentEpicFile)
	assert.Contains(t, output, "Current: "+targetEpicFile)

	// Verify configuration was updated
	updatedCfg, err := config.LoadConfig(configFile)
	require.NoError(t, err)
	assert.Equal(t, targetEpicFile, updatedCfg.CurrentEpic)
	assert.Equal(t, currentEpicFile, updatedCfg.PreviousEpic)
}

func TestSwitchCommand_WithAbsolutePath(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".agentpm.json")
	currentEpicFile := filepath.Join(tempDir, "current-epic.xml")
	targetEpicFile := filepath.Join(tempDir, "target-epic.xml")

	// Create test config
	cfg := &config.Config{
		CurrentEpic: "current-epic.xml",
	}
	require.NoError(t, config.SaveConfig(cfg, configFile))

	// Create epic files
	currentEpic := &epic.Epic{ID: "current-epic", Name: "Current Epic", Status: epic.StatusActive}
	writeTestEpicXML(t, currentEpicFile, currentEpic)

	targetEpic := &epic.Epic{ID: "target-epic", Name: "Target Epic", Status: epic.StatusPlanning}
	writeTestEpicXML(t, targetEpicFile, targetEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Value: configFile},
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			SwitchCommand(),
		},
	}

	var stdout bytes.Buffer
	app.Writer = &stdout

	// Run switch command with absolute path
	args := []string{"agentpm", "switch", targetEpicFile}
	err := app.Run(context.Background(), args)

	require.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "Epic switched successfully")
	assert.Contains(t, output, "Current: "+targetEpicFile)
}

func TestSwitchCommand_SwitchBack(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".agentpm.json")
	epic1File := filepath.Join(tempDir, "epic-1.xml")
	epic2File := filepath.Join(tempDir, "epic-2.xml")

	// Create test config with both current and previous epics
	cfg := &config.Config{
		CurrentEpic:  epic2File,
		PreviousEpic: epic1File,
	}
	require.NoError(t, config.SaveConfig(cfg, configFile))

	// Create epic files
	epic1 := &epic.Epic{ID: "epic-1", Name: "Epic 1", Status: epic.StatusActive}
	writeTestEpicXML(t, epic1File, epic1)

	epic2 := &epic.Epic{ID: "epic-2", Name: "Epic 2", Status: epic.StatusPlanning}
	writeTestEpicXML(t, epic2File, epic2)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Value: configFile},
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			SwitchCommand(),
		},
	}

	var stdout bytes.Buffer
	app.Writer = &stdout

	// Run switch back command
	args := []string{"agentpm", "switch", "--back"}
	err := app.Run(context.Background(), args)

	require.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "Switched back from "+epic2File+" to "+epic1File)

	// Verify configuration was updated correctly
	updatedCfg, err := config.LoadConfig(configFile)
	require.NoError(t, err)
	assert.Equal(t, epic1File, updatedCfg.CurrentEpic)
	assert.Equal(t, epic2File, updatedCfg.PreviousEpic)
}

func TestSwitchCommand_JSONOutput(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".agentpm.json")
	currentEpicFile := filepath.Join(tempDir, "current-epic.xml")
	targetEpicFile := filepath.Join(tempDir, "target-epic.xml")

	// Create test config
	cfg := &config.Config{
		CurrentEpic: currentEpicFile,
	}
	require.NoError(t, config.SaveConfig(cfg, configFile))

	// Create epic files
	currentEpic := &epic.Epic{ID: "current-epic", Name: "Current Epic", Status: epic.StatusActive}
	writeTestEpicXML(t, currentEpicFile, currentEpic)

	targetEpic := &epic.Epic{ID: "target-epic", Name: "Target Epic", Status: epic.StatusPlanning}
	writeTestEpicXML(t, targetEpicFile, targetEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Value: configFile},
			&cli.StringFlag{Name: "format", Value: "json"},
		},
		Commands: []*cli.Command{
			SwitchCommand(),
		},
	}

	var stdout bytes.Buffer
	app.Writer = &stdout

	// Run switch command
	args := []string{"agentpm", "switch", targetEpicFile}
	err := app.Run(context.Background(), args)

	require.NoError(t, err)

	// Parse JSON output
	var output map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &output)
	require.NoError(t, err)

	// Verify JSON structure
	assert.Contains(t, output, "epic_switched")
	switched := output["epic_switched"].(map[string]interface{})
	assert.Equal(t, currentEpicFile, switched["previous_epic"])
	assert.Equal(t, targetEpicFile, switched["new_epic"])
	assert.Contains(t, switched["message"], "Switched from "+currentEpicFile+" to "+targetEpicFile)
}

func TestSwitchCommand_XMLOutput(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".agentpm.json")
	currentEpicFile := filepath.Join(tempDir, "current-epic.xml")
	targetEpicFile := filepath.Join(tempDir, "target-epic.xml")

	// Create test config
	cfg := &config.Config{
		CurrentEpic: currentEpicFile,
	}
	require.NoError(t, config.SaveConfig(cfg, configFile))

	// Create epic files
	currentEpic := &epic.Epic{ID: "current-epic", Name: "Current Epic", Status: epic.StatusActive}
	writeTestEpicXML(t, currentEpicFile, currentEpic)

	targetEpic := &epic.Epic{ID: "target-epic", Name: "Target Epic", Status: epic.StatusPlanning}
	writeTestEpicXML(t, targetEpicFile, targetEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Value: configFile},
			&cli.StringFlag{Name: "format", Value: "xml"},
		},
		Commands: []*cli.Command{
			SwitchCommand(),
		},
	}

	var stdout bytes.Buffer
	app.Writer = &stdout

	// Run switch command
	args := []string{"agentpm", "switch", targetEpicFile}
	err := app.Run(context.Background(), args)

	require.NoError(t, err)

	// Check XML output using snapshots
	output := stdout.String()
	snapshotTester := apmtesting.NewSnapshotTester()
	snapshotTester.MatchXMLSnapshot(t, output, "switch_command_xml_output")
}

func TestSwitchCommand_ErrorNonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".agentpm.json")
	currentEpicFile := filepath.Join(tempDir, "current-epic.xml")

	// Create test config
	cfg := &config.Config{
		CurrentEpic: "current-epic.xml",
	}
	require.NoError(t, config.SaveConfig(cfg, configFile))

	// Create current epic file
	currentEpic := &epic.Epic{ID: "current-epic", Name: "Current Epic", Status: epic.StatusActive}
	writeTestEpicXML(t, currentEpicFile, currentEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Value: configFile},
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			SwitchCommand(),
		},
	}

	var stderr bytes.Buffer
	app.ErrWriter = &stderr

	// Run switch command with non-existent file
	args := []string{"agentpm", "switch", "non-existent.xml"}
	err := app.Run(context.Background(), args)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "epic file does not exist")
}

func TestSwitchCommand_ErrorInvalidEpicFile(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".agentpm.json")
	currentEpicFile := filepath.Join(tempDir, "current-epic.xml")
	invalidEpicFile := filepath.Join(tempDir, "invalid-epic.xml")

	// Create test config
	cfg := &config.Config{
		CurrentEpic: currentEpicFile,
	}
	require.NoError(t, config.SaveConfig(cfg, configFile))

	// Create current epic file
	currentEpic := &epic.Epic{ID: "current-epic", Name: "Current Epic", Status: epic.StatusActive}
	writeTestEpicXML(t, currentEpicFile, currentEpic)

	// Create invalid epic file (bad XML)
	require.NoError(t, writeFile(invalidEpicFile, "invalid xml content"))

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Value: configFile},
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			SwitchCommand(),
		},
	}

	var stderr bytes.Buffer
	app.ErrWriter = &stderr

	// Run switch command with invalid epic file
	args := []string{"agentpm", "switch", invalidEpicFile}
	err := app.Run(context.Background(), args)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid epic file")
}

func TestSwitchCommand_ErrorNoTargetSpecified(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".agentpm.json")
	currentEpicFile := filepath.Join(tempDir, "current-epic.xml")

	// Create test config
	cfg := &config.Config{
		CurrentEpic: "current-epic.xml",
	}
	require.NoError(t, config.SaveConfig(cfg, configFile))

	// Create current epic file
	currentEpic := &epic.Epic{ID: "current-epic", Name: "Current Epic", Status: epic.StatusActive}
	writeTestEpicXML(t, currentEpicFile, currentEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Value: configFile},
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			SwitchCommand(),
		},
	}

	var stderr bytes.Buffer
	app.ErrWriter = &stderr

	// Run switch command without target
	args := []string{"agentpm", "switch"}
	err := app.Run(context.Background(), args)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target epic file is required")
}

func TestSwitchCommand_ErrorSwitchBackNoPrevious(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".agentpm.json")
	currentEpicFile := filepath.Join(tempDir, "current-epic.xml")

	// Create test config without previous epic
	cfg := &config.Config{
		CurrentEpic: "current-epic.xml",
		// No PreviousEpic set
	}
	require.NoError(t, config.SaveConfig(cfg, configFile))

	// Create current epic file
	currentEpic := &epic.Epic{ID: "current-epic", Name: "Current Epic", Status: epic.StatusActive}
	writeTestEpicXML(t, currentEpicFile, currentEpic)

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Value: configFile},
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			SwitchCommand(),
		},
	}

	var stderr bytes.Buffer
	app.ErrWriter = &stderr

	// Run switch back command without previous epic
	args := []string{"agentpm", "switch", "--back"}
	err := app.Run(context.Background(), args)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no previous epic to switch back to")
}

func TestSwitchCommand_ErrorMissingConfig(t *testing.T) {
	tempDir := t.TempDir()
	nonExistentConfig := filepath.Join(tempDir, "non-existent.json")

	// Create CLI app
	app := &cli.Command{
		Name: "agentpm",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Value: nonExistentConfig},
			&cli.StringFlag{Name: "format", Value: "text"},
		},
		Commands: []*cli.Command{
			SwitchCommand(),
		},
	}

	var stderr bytes.Buffer
	app.ErrWriter = &stderr

	// Run switch command with missing config
	args := []string{"agentpm", "switch", "target.xml"}
	err := app.Run(context.Background(), args)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load configuration")
}

// Helper function to write content to file
func writeFile(filePath, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}
