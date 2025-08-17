package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestFixXMLCommand(t *testing.T) {
	t.Run("fix unescaped ampersands", func(t *testing.T) {
		// Create broken XML with unescaped &
		brokenXML := `<?xml version="1.0" encoding="UTF-8"?>
<epic id="test" name="Test & Debug" status="planning">
    <description>Testing & validation work</description>
    <phases>
        <phase id="1" name="Setup & Config" status="pending">
            <description>Setup the system & configure it properly</description>
        </phase>
    </phases>
    <tasks></tasks>
    <tests></tests>
    <events></events>
</epic>`

		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "broken.xml")
		configFile := filepath.Join(tempDir, ".agentpm.json")

		// Write broken XML
		require.NoError(t, os.WriteFile(epicFile, []byte(brokenXML), 0644))

		// Write config
		config := `{"current_epic": "broken.xml"}`
		require.NoError(t, os.WriteFile(configFile, []byte(config), 0644))

		// Change to temp directory
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tempDir)

		// Run fix command
		app := &cli.Command{Commands: []*cli.Command{FixXMLCommand()}}
		var output strings.Builder
		app.Writer = &output

		err := app.Run(context.Background(), []string{"agentpm", "fix-xml"})
		require.NoError(t, err)

		// Check output
		outputStr := output.String()
		assert.Contains(t, outputStr, "Found")
		assert.Contains(t, outputStr, "XML encoding issues")
		assert.Contains(t, outputStr, "Successfully fixed")

		// Verify file was fixed
		fixedContent, err := os.ReadFile(epicFile)
		require.NoError(t, err)
		fixedStr := string(fixedContent)

		// Should have escaped ampersands
		assert.Contains(t, fixedStr, "Test &amp; Debug")
		assert.Contains(t, fixedStr, "Testing &amp; validation")
		assert.Contains(t, fixedStr, "Setup &amp; Config")
		assert.Contains(t, fixedStr, "Setup the system &amp; configure")

		// Verify backup was created
		backupFiles, _ := filepath.Glob(filepath.Join(tempDir, "broken.xml.backup.*"))
		assert.Len(t, backupFiles, 1)
	})

	t.Run("no fixes needed for valid XML", func(t *testing.T) {
		validXML := `<?xml version="1.0" encoding="UTF-8"?>
<epic id="test" name="Test Epic" status="planning">
    <description>Valid XML content without issues</description>
    <phases>
        <phase id="1" name="Phase One" status="pending">
            <description>Valid description</description>
        </phase>
    </phases>
    <tasks></tasks>
    <tests></tests>
    <events></events>
</epic>`

		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "valid.xml")
		configFile := filepath.Join(tempDir, ".agentpm.json")

		require.NoError(t, os.WriteFile(epicFile, []byte(validXML), 0644))
		config := `{"current_epic": "valid.xml"}`
		require.NoError(t, os.WriteFile(configFile, []byte(config), 0644))

		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tempDir)

		app := &cli.Command{Commands: []*cli.Command{FixXMLCommand()}}
		var output strings.Builder
		app.Writer = &output

		err := app.Run(context.Background(), []string{"agentpm", "fix-xml"})
		require.NoError(t, err)

		outputStr := output.String()
		assert.Contains(t, outputStr, "No XML encoding issues found")
	})

	t.Run("dry run mode", func(t *testing.T) {
		brokenXML := `<?xml version="1.0" encoding="UTF-8"?>
<epic id="test" name="Test & Debug" status="planning">
    <description>Testing & validation</description>
    <phases></phases>
    <tasks></tasks>
    <tests></tests>
    <events></events>
</epic>`

		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "broken.xml")
		configFile := filepath.Join(tempDir, ".agentpm.json")

		require.NoError(t, os.WriteFile(epicFile, []byte(brokenXML), 0644))
		config := `{"current_epic": "broken.xml"}`
		require.NoError(t, os.WriteFile(configFile, []byte(config), 0644))

		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tempDir)

		app := &cli.Command{Commands: []*cli.Command{FixXMLCommand()}}
		var output strings.Builder
		app.Writer = &output

		err := app.Run(context.Background(), []string{"agentpm", "fix-xml", "--dry-run"})
		require.NoError(t, err)

		outputStr := output.String()
		assert.Contains(t, outputStr, "Dry run mode")
		assert.Contains(t, outputStr, "no changes made")

		// Verify file was NOT changed
		unchangedContent, err := os.ReadFile(epicFile)
		require.NoError(t, err)
		assert.Equal(t, brokenXML, string(unchangedContent))

		// Verify no backup was created
		backupFiles, _ := filepath.Glob(filepath.Join(tempDir, "broken.xml.backup.*"))
		assert.Len(t, backupFiles, 0)
	})

	t.Run("no backup mode", func(t *testing.T) {
		brokenXML := `<?xml version="1.0" encoding="UTF-8"?>
<epic id="test" name="Test & Debug" status="planning">
    <description>Testing</description>
    <phases></phases>
    <tasks></tasks>
    <tests></tests>
    <events></events>
</epic>`

		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "broken.xml")
		configFile := filepath.Join(tempDir, ".agentpm.json")

		require.NoError(t, os.WriteFile(epicFile, []byte(brokenXML), 0644))
		config := `{"current_epic": "broken.xml"}`
		require.NoError(t, os.WriteFile(configFile, []byte(config), 0644))

		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tempDir)

		app := &cli.Command{Commands: []*cli.Command{FixXMLCommand()}}
		var output strings.Builder
		app.Writer = &output

		err := app.Run(context.Background(), []string{"agentpm", "fix-xml", "--backup=false"})
		require.NoError(t, err)

		// Verify no backup was created
		backupFiles, _ := filepath.Glob(filepath.Join(tempDir, "broken.xml.backup.*"))
		assert.Len(t, backupFiles, 0)

		// But file should still be fixed
		fixedContent, err := os.ReadFile(epicFile)
		require.NoError(t, err)
		assert.Contains(t, string(fixedContent), "&amp;")
	})

	t.Run("no issues found", func(t *testing.T) {
		validXML := `<?xml version="1.0" encoding="UTF-8"?>
<epic id="test" name="Test Epic" status="planning">
    <description>Valid XML content</description>
    <phases></phases>
    <tasks></tasks>
    <tests></tests>
    <events></events>
</epic>`

		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "valid.xml")
		configFile := filepath.Join(tempDir, ".agentpm.json")

		require.NoError(t, os.WriteFile(epicFile, []byte(validXML), 0644))
		config := `{"current_epic": "valid.xml"}`
		require.NoError(t, os.WriteFile(configFile, []byte(config), 0644))

		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tempDir)

		app := &cli.Command{Commands: []*cli.Command{FixXMLCommand()}}
		var output strings.Builder
		app.Writer = &output

		err := app.Run(context.Background(), []string{"agentpm", "fix-xml"})
		require.NoError(t, err)

		outputStr := output.String()
		assert.Contains(t, outputStr, "No XML encoding issues found")
	})

	t.Run("file override flag", func(t *testing.T) {
		brokenXML := `<?xml version="1.0" encoding="UTF-8"?>
<epic id="test" name="Test & Debug" status="planning">
    <description>Testing</description>
    <phases></phases>
    <tasks></tasks>
    <tests></tests>
    <events></events>
</epic>`

		tempDir := t.TempDir()
		epicFile := filepath.Join(tempDir, "custom.xml")

		// Create a dummy config file in case the command needs it
		configFile := filepath.Join(tempDir, ".agentpm.json")
		config := `{"current_epic": "dummy.xml"}`
		require.NoError(t, os.WriteFile(configFile, []byte(config), 0644))

		require.NoError(t, os.WriteFile(epicFile, []byte(brokenXML), 0644))

		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tempDir)

		app := &cli.Command{Commands: []*cli.Command{FixXMLCommand()}}
		var output strings.Builder
		app.Writer = &output

		err := app.Run(context.Background(), []string{"agentpm", "fix-xml", "-f", "custom.xml"})
		require.NoError(t, err)

		outputStr := output.String()
		assert.Contains(t, outputStr, "Successfully fixed")
	})
}

func TestXMLFixer(t *testing.T) {
	t.Run("fix unescaped ampersands", func(t *testing.T) {
		fixer := NewXMLFixer()
		content := "Test & debug and A&B but not &amp; or &lt;"

		fixed, fixes := fixer.FixXMLContent(content)

		assert.Contains(t, fixed, "Test &amp; debug")
		assert.Contains(t, fixed, "A&amp;B")
		assert.Contains(t, fixed, "&amp;") // should remain
		assert.Contains(t, fixed, "&lt;")  // should remain
		assert.Len(t, fixes, 2)            // Two & characters were fixed
	})

	t.Run("fix angle brackets with error-driven approach", func(t *testing.T) {
		fixer := NewXMLFixer()
		// Use a case that matches our test-angle-brackets.xml pattern
		content := `<epic><description>Use < command & check</description></epic>`

		fixed, fixes := fixer.FixXMLContent(content)

		// Error-driven approach should fix problematic characters
		assert.Contains(t, fixed, "&lt;")
		assert.Contains(t, fixed, "&amp;")
		assert.Len(t, fixes, 2) // Should fix both < and & characters
	})

	t.Run("fix malformed entities", func(t *testing.T) {
		fixer := NewXMLFixer()
		content := "Use &amp and &lt for escaping"

		fixed, fixes := fixer.FixXMLContent(content)

		// Current implementation escapes the & in malformed entities
		assert.Contains(t, fixed, "Use &amp;amp and &amp;lt for escaping")
		assert.Len(t, fixes, 2) // Two ampersands were escaped
	})

	t.Run("no fixes needed", func(t *testing.T) {
		fixer := NewXMLFixer()
		content := "<tag>Valid content with &amp; and &lt;</tag>"

		fixed, fixes := fixer.FixXMLContent(content)

		assert.Equal(t, content, fixed)
		assert.Len(t, fixes, 0)
	})
}
