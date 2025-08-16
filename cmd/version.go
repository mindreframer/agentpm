package cmd

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"runtime"
	"strings"

	"github.com/urfave/cli/v3"
)

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

type VersionInfo struct {
	Version   string `json:"version" xml:"version"`
	GitCommit string `json:"git_commit" xml:"git_commit"`
	BuildDate string `json:"build_date" xml:"build_date"`
	GoVersion string `json:"go_version" xml:"go_version"`
}

func VersionCommand() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Usage:   "Display version information",
		Aliases: []string{"ver", "v"},
		Action:  versionAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"F"},
				Usage:   "Output format: text (default), json, xml",
				Value:   "text",
			},
		},
	}
}

func versionAction(ctx context.Context, c *cli.Command) error {
	format := c.String("format")
	if format == "" {
		format = "text"
	}

	format = strings.ToLower(format)
	if format != "text" && format != "json" && format != "xml" {
		return fmt.Errorf("invalid format '%s': must be text, json, or xml", format)
	}

	versionInfo := VersionInfo{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
	}

	switch format {
	case "json":
		return outputJSON(c, versionInfo)
	case "xml":
		return outputXML(c, versionInfo)
	default:
		return outputText(c, versionInfo)
	}
}

func outputText(c *cli.Command, info VersionInfo) error {
	fmt.Fprintf(c.Root().Writer, "agentpm version %s\n", info.Version)
	fmt.Fprintf(c.Root().Writer, "Git commit: %s\n", info.GitCommit)
	fmt.Fprintf(c.Root().Writer, "Built: %s\n", info.BuildDate)
	fmt.Fprintf(c.Root().Writer, "Go version: %s\n", info.GoVersion)
	return nil
}

func outputJSON(c *cli.Command, info VersionInfo) error {
	encoder := json.NewEncoder(c.Root().Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(info)
}

func outputXML(c *cli.Command, info VersionInfo) error {
	fmt.Fprintf(c.Root().Writer, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	fmt.Fprintf(c.Root().Writer, "<version_info>\n")
	encoder := xml.NewEncoder(c.Root().Writer)
	encoder.Indent("", "  ")
	if err := encoder.Encode(info); err != nil {
		return err
	}
	fmt.Fprintf(c.Root().Writer, "\n</version_info>\n")
	return nil
}
