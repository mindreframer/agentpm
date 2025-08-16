package main

import (
	"context"
	"fmt"
	"os"

	"github.com/memomoo/agentpm/cmd"
	"github.com/urfave/cli/v3"
)

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
			cmd.InitCommand(),
			cmd.ConfigCommand(),
			cmd.ValidateCommand(),
			cmd.StatusCommand(),
			cmd.CurrentCommand(),
			cmd.PendingCommand(),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
