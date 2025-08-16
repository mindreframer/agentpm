package main

import (
	"context"
	"fmt"
	"os"

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
			{
				Name:  "init",
				Usage: "Initialize a new project with an epic file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "epic",
						Usage:    "Epic file to set as current",
						Required: true,
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					fmt.Printf("Init command called with epic: %s\n", c.String("epic"))
					return nil
				},
			},
			{
				Name:  "config",
				Usage: "Display current project configuration",
				Action: func(ctx context.Context, c *cli.Command) error {
					fmt.Println("Config command called")
					return nil
				},
			},
			{
				Name:  "validate",
				Usage: "Validate epic XML structure",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Usage:   "Epic file to validate (overrides config)",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					fmt.Println("Validate command called")
					return nil
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
