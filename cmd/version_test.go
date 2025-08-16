package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestVersionCommand(t *testing.T) {
	oldVersion, oldCommit, oldBuildDate := Version, GitCommit, BuildDate
	defer func() {
		Version, GitCommit, BuildDate = oldVersion, oldCommit, oldBuildDate
	}()

	Version = "1.0.0"
	GitCommit = "abc123de"
	BuildDate = "2025-08-16T14:30:00Z"

	t.Run("version command displays default text format", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cmd := VersionCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err := cmd.Run(context.Background(), []string{"version"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "agentpm version 1.0.0")
		assert.Contains(t, output, "Git commit: abc123de")
		assert.Contains(t, output, "Built: 2025-08-16T14:30:00Z")
		assert.Contains(t, output, "Go version: "+runtime.Version())
	})

	t.Run("version command with format=text", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cmd := VersionCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err := cmd.Run(context.Background(), []string{"version", "--format=text"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "agentpm version 1.0.0")
		assert.Contains(t, output, "Git commit: abc123de")
		assert.Contains(t, output, "Built: 2025-08-16T14:30:00Z")
		assert.Contains(t, output, "Go version: "+runtime.Version())
	})

	t.Run("version command with format=json outputs valid JSON", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cmd := VersionCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err := cmd.Run(context.Background(), []string{"version", "--format=json"})
		require.NoError(t, err)

		output := stdout.String()

		var versionInfo VersionInfo
		err = json.Unmarshal([]byte(output), &versionInfo)
		require.NoError(t, err, "JSON output should be valid")

		assert.Equal(t, "1.0.0", versionInfo.Version)
		assert.Equal(t, "abc123de", versionInfo.GitCommit)
		assert.Equal(t, "2025-08-16T14:30:00Z", versionInfo.BuildDate)
		assert.Equal(t, runtime.Version(), versionInfo.GoVersion)
	})

	t.Run("version command with format=xml outputs valid XML", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cmd := VersionCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err := cmd.Run(context.Background(), []string{"version", "--format=xml"})
		require.NoError(t, err)

		output := stdout.String()

		assert.Contains(t, output, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
		assert.Contains(t, output, "<version_info>")
		assert.Contains(t, output, "</version_info>")

		lines := strings.Split(output, "\n")
		xmlContent := ""
		inVersionInfo := false
		for _, line := range lines {
			if strings.Contains(line, "<version_info>") {
				inVersionInfo = true
				continue
			}
			if strings.Contains(line, "</version_info>") {
				break
			}
			if inVersionInfo {
				xmlContent += line + "\n"
			}
		}

		var versionInfo VersionInfo
		err = xml.Unmarshal([]byte(xmlContent), &versionInfo)
		require.NoError(t, err, "XML content should be valid")

		assert.Equal(t, "1.0.0", versionInfo.Version)
		assert.Equal(t, "abc123de", versionInfo.GitCommit)
		assert.Equal(t, "2025-08-16T14:30:00Z", versionInfo.BuildDate)
		assert.Equal(t, runtime.Version(), versionInfo.GoVersion)
	})

	t.Run("version command with invalid format shows error", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cmd := VersionCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err := cmd.Run(context.Background(), []string{"version", "--format=invalid"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid format 'invalid'")
		assert.Contains(t, err.Error(), "must be text, json, or xml")
	})

	t.Run("version variables default values work in dev mode", func(t *testing.T) {
		Version = "dev"
		GitCommit = "unknown"
		BuildDate = "unknown"

		var stdout, stderr bytes.Buffer
		cmd := VersionCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err := cmd.Run(context.Background(), []string{"version"})
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "agentpm version dev")
		assert.Contains(t, output, "Git commit: unknown")
		assert.Contains(t, output, "Built: unknown")
		assert.Contains(t, output, "Go version: "+runtime.Version())
	})

	t.Run("all required version metadata is included", func(t *testing.T) {
		Version = "2.1.0"
		GitCommit = "def456gh"
		BuildDate = "2025-08-16T15:45:30Z"

		var stdout, stderr bytes.Buffer
		cmd := VersionCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err := cmd.Run(context.Background(), []string{"version", "--format=json"})
		require.NoError(t, err)

		var versionInfo VersionInfo
		err = json.Unmarshal([]byte(stdout.String()), &versionInfo)
		require.NoError(t, err)

		assert.NotEmpty(t, versionInfo.Version, "Version should not be empty")
		assert.NotEmpty(t, versionInfo.GitCommit, "GitCommit should not be empty")
		assert.NotEmpty(t, versionInfo.BuildDate, "BuildDate should not be empty")
		assert.NotEmpty(t, versionInfo.GoVersion, "GoVersion should not be empty")
		assert.True(t, strings.HasPrefix(versionInfo.GoVersion, "go"), "GoVersion should start with 'go'")
	})

	t.Run("command structure follows existing CLI patterns", func(t *testing.T) {
		cmd := VersionCommand()

		assert.Equal(t, "version", cmd.Name)
		assert.Equal(t, "Display version information", cmd.Usage)
		assert.Contains(t, cmd.Aliases, "ver")
		assert.Contains(t, cmd.Aliases, "v")
		assert.NotNil(t, cmd.Action)

		var formatFlag *string
		for _, flag := range cmd.Flags {
			if flag.Names()[0] == "format" {
				if sf, ok := flag.(*cli.StringFlag); ok {
					formatFlag = &sf.Value
					assert.Contains(t, flag.Names(), "F", "Format flag should have 'F' alias")
				}
			}
		}
		require.NotNil(t, formatFlag, "Format flag should exist")
		assert.Equal(t, "text", *formatFlag, "Default format should be text")
	})

	t.Run("version command supports case insensitive format", func(t *testing.T) {
		testCases := []string{"JSON", "Json", "jSoN", "XML", "Xml", "xMl", "TEXT", "Text", "tExT"}

		for _, format := range testCases {
			t.Run("format="+format, func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				cmd := VersionCommand()
				cmd.Root().Writer = &stdout
				cmd.Root().ErrWriter = &stderr

				err := cmd.Run(context.Background(), []string{"version", "--format=" + format})
				require.NoError(t, err, "Should accept format: %s", format)

				output := stdout.String()
				assert.NotEmpty(t, output, "Should produce output for format: %s", format)
			})
		}
	})

	t.Run("version command with short flag alias", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cmd := VersionCommand()
		cmd.Root().Writer = &stdout
		cmd.Root().ErrWriter = &stderr

		err := cmd.Run(context.Background(), []string{"version", "-F", "json"})
		require.NoError(t, err)

		var versionInfo VersionInfo
		err = json.Unmarshal([]byte(stdout.String()), &versionInfo)
		require.NoError(t, err, "Should work with short flag alias")
	})
}
