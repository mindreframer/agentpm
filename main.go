package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mindreframer/agentpm/cmd"
	"github.com/urfave/cli/v3"
)

// addCategory adds a category to a command for grouped help output
func addCategory(command *cli.Command, category string) *cli.Command {
	command.Category = category
	return command
}

func main() {
	app := &cli.Command{
		Name:  "agentpm",
		Usage: "CLI tool for LLM agents to manage epic-based development work",
		Flags: []cli.Flag{
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
		},
		Commands: []*cli.Command{
			// CORE WORKFLOW - Most frequently used commands
			addCategory(cmd.StartCommand(), "CORE WORKFLOW"),
			addCategory(cmd.DoneCommand(), "CORE WORKFLOW"),
			addCategory(cmd.CancelCommand(), "CORE WORKFLOW"),
			addCategory(cmd.StartNextCommand(), "CORE WORKFLOW"),

			// TESTING - Test management commands
			addCategory(cmd.PassCommand(), "TESTING"),
			addCategory(cmd.FailCommand(), "TESTING"),

			// STATUS - Information and monitoring commands
			addCategory(cmd.StatusCommand(), "STATUS"),
			addCategory(cmd.CurrentCommand(), "STATUS"),
			addCategory(cmd.PendingCommand(), "STATUS"),
			addCategory(cmd.FailingCommand(), "STATUS"),

			// INSPECTION - Detailed entity examination
			addCategory(cmd.ShowCommand(), "INSPECTION"),
			addCategory(cmd.QueryCommand(), "INSPECTION"),

			// PROJECT - Project setup and management
			addCategory(cmd.InitCommand(), "PROJECT"),
			addCategory(cmd.SwitchCommand(), "PROJECT"),
			addCategory(cmd.ConfigCommand(), "PROJECT"),
			addCategory(cmd.ValidateCommand(), "PROJECT"),

			// REPORTING - Documentation and handoff
			addCategory(cmd.LogCommand(), "REPORTING"),
			addCategory(cmd.EventsCommand(), "REPORTING"),
			addCategory(cmd.DocsCommand(), "REPORTING"),
			addCategory(cmd.HandoffCommand(), "REPORTING"),

			// SYSTEM - Version and help
			addCategory(cmd.VersionCommand(), "SYSTEM"),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
