package cmd

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/mindreframer/agentpm/internal/config"
	"github.com/urfave/cli/v3"
)

// FixXMLCommand returns the fix-xml command for fixing XML encoding issues
func FixXMLCommand() *cli.Command {
	return &cli.Command{
		Name:    "fix-xml",
		Usage:   "Fix XML encoding issues in epic files",
		Aliases: []string{"fix"},
		Action:  fixXMLAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Epic file to fix (default: from config)",
			},
			&cli.BoolFlag{
				Name:  "backup",
				Usage: "Create backup file before fixing (default: true)",
				Value: true,
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Show what would be fixed without making changes",
			},
		},
		Description: `Fix common XML encoding issues in epic files using error-driven analysis:

• Escape unescaped ampersands (&) to &amp;
• Escape less-than signs (<) to &lt; in text content (error-driven)
• Fix malformed character entities
• Uses XML parser errors to identify and fix only problematic characters
• Preserves valid XML structure and tags

The command will:
1. Create a backup (unless --backup=false)
2. Apply fixes to make the XML parseable
3. Validate the fixed XML can be loaded
4. Report what was fixed

Examples:
  agentpm fix-xml                    # Fix current epic file
  agentpm fix-xml -f epic.xml        # Fix specific file
  agentpm fix-xml --dry-run          # Show what would be fixed
  agentpm fix-xml --backup=false     # Fix without backup`,
	}
}

func fixXMLAction(ctx context.Context, c *cli.Command) error {
	// Load configuration
	configPath := c.String("config")
	if configPath == "" {
		configPath = "./.agentpm.json"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine epic file
	epicFile := c.String("file")
	if epicFile == "" {
		epicFile = cfg.CurrentEpic
	}
	if epicFile == "" {
		return fmt.Errorf("no epic file specified. Use --file flag or run 'agentpm init' first")
	}

	absPath, err := filepath.Abs(epicFile)
	if err != nil {
		return fmt.Errorf("failed to resolve epic file path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("epic file does not exist: %s", absPath)
	}

	// Read current content
	content, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read epic file: %w", err)
	}

	originalContent := string(content)

	// Apply fixes
	fixer := NewXMLFixer()
	fixedContent, fixes := fixer.FixXMLContent(originalContent)

	// Report what would be/was fixed
	if len(fixes) == 0 {
		fmt.Fprintf(c.Root().Writer, "No XML encoding issues found in %s\n", epicFile)
		return nil
	}

	fmt.Fprintf(c.Root().Writer, "Found %d XML encoding issues in %s:\n", len(fixes), epicFile)
	for i, fix := range fixes {
		fmt.Fprintf(c.Root().Writer, "  %d. %s\n", i+1, fix.Description)
	}

	// Dry run - just show what would be fixed
	if c.Bool("dry-run") {
		fmt.Fprintf(c.Root().Writer, "\nDry run mode - no changes made.\n")
		fmt.Fprintf(c.Root().Writer, "Run without --dry-run to apply fixes.\n")
		return nil
	}

	// Validate that fixed content can be parsed
	doc := etree.NewDocument()
	if err := doc.ReadFromString(fixedContent); err != nil {
		return fmt.Errorf("fixed XML is still invalid: %w", err)
	}

	// Create backup if requested
	if c.Bool("backup") {
		backupPath := absPath + ".backup." + time.Now().Format("20060102-150405")
		if err := os.WriteFile(backupPath, content, 0644); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		fmt.Fprintf(c.Root().Writer, "Created backup: %s\n", backupPath)
	}

	// Write fixed content
	if err := os.WriteFile(absPath, []byte(fixedContent), 0644); err != nil {
		return fmt.Errorf("failed to write fixed file: %w", err)
	}

	fmt.Fprintf(c.Root().Writer, "Successfully fixed XML encoding issues!\n")
	fmt.Fprintf(c.Root().Writer, "File updated: %s\n", epicFile)

	return nil
}

// XMLFix represents a single fix that was applied
type XMLFix struct {
	Description string
	LineNumber  int
	Before      string
	After       string
}

// XMLFixer handles fixing XML content
type XMLFixer struct {
	fixes []XMLFix
}

// NewXMLFixer creates a new XML fixer
func NewXMLFixer() *XMLFixer {
	return &XMLFixer{
		fixes: make([]XMLFix, 0),
	}
}

// FixXMLContent fixes XML encoding issues in the provided content
func (f *XMLFixer) FixXMLContent(content string) (string, []XMLFix) {
	f.fixes = make([]XMLFix, 0)

	// Apply fixes in order - start with the most common and safest fixes
	content = f.fixUnescapedAmpersands(content)
	content = f.fixUnescapedAngleBrackets(content)
	content = f.fixMalformedEntities(content)

	return content, f.fixes
}

// fixUnescapedAmpersands fixes & that should be &amp;
func (f *XMLFixer) fixUnescapedAmpersands(content string) string {
	// First, find all valid entities to preserve
	validEntityPattern := `&(?:amp|lt|gt|quot|apos|#\d+|#x[0-9a-fA-F]+);`
	validEntityRegex := regexp.MustCompile(validEntityPattern)

	// Create a map to track positions of valid entities
	validEntities := validEntityRegex.FindAllStringIndex(content, -1)
	validPositions := make(map[int]bool)
	for _, entity := range validEntities {
		for i := entity[0]; i < entity[1]; i++ {
			validPositions[i] = true
		}
	}

	// Now replace unescaped & characters
	result := ""
	for i := 0; i < len(content); i++ {
		if content[i] == '&' && !validPositions[i] {
			// This is an unescaped ampersand
			result += "&amp;"
			f.fixes = append(f.fixes, XMLFix{
				Description: "Escaped unescaped ampersand (&) to &amp;",
				Before:      "&",
				After:       "&amp;",
			})
		} else {
			result += string(content[i])
		}
	}

	return result
}

// fixUnescapedAngleBrackets fixes < and > characters using error-driven approach
func (f *XMLFixer) fixUnescapedAngleBrackets(content string) string {
	return f.fixXMLWithErrorAnalysis(content)
}

// fixXMLWithErrorAnalysis uses XML parser errors to identify and fix specific issues
func (f *XMLFixer) fixXMLWithErrorAnalysis(content string) string {
	maxIterations := 10 // Prevent infinite loops
	currentContent := content

	for i := 0; i < maxIterations; i++ {
		// Try to parse the current content with etree first
		doc := etree.NewDocument()
		err := doc.ReadFromString(currentContent)

		if err == nil {
			// Parsing successful, we're done
			break
		}

		// If etree fails, try raw XML parser to get detailed error
		rawErr := f.getRawXMLError(currentContent)

		// Analyze the error and attempt to fix it
		fixed, wasFixed := f.fixBasedOnError(currentContent, rawErr)
		if !wasFixed {
			// Can't fix this error, stop trying
			break
		}

		currentContent = fixed
	}

	return currentContent
}

// getRawXMLError uses the raw XML parser to get detailed error information
func (f *XMLFixer) getRawXMLError(content string) error {
	decoder := xml.NewDecoder(strings.NewReader(content))
	for {
		_, err := decoder.Token()
		if err != nil {
			return err
		}
	}
}

// fixBasedOnError analyzes a specific parsing error and attempts to fix it
func (f *XMLFixer) fixBasedOnError(content string, err error) (string, bool) {
	syntaxErr, ok := err.(*xml.SyntaxError)
	if !ok {
		// Not a syntax error we can analyze
		return content, false
	}

	lines := strings.Split(content, "\n")
	if syntaxErr.Line <= 0 || syntaxErr.Line > len(lines) {
		return content, false
	}

	lineIndex := syntaxErr.Line - 1
	problemLine := lines[lineIndex]

	// Analyze error message and apply appropriate fix
	switch {
	case strings.Contains(syntaxErr.Msg, "expected element name after <"):
		// Fix unescaped < in text content
		return f.fixUnescapedLessThanInLine(content, lineIndex, problemLine)

	case strings.Contains(syntaxErr.Msg, "closed by"):
		// Handle "element <command> closed by </tag>" - indicates unescaped < in text
		return f.fixUnescapedLessThanInLine(content, lineIndex, problemLine)

	case strings.Contains(syntaxErr.Msg, "invalid character entity"):
		// This should already be handled by ampersand fixing
		return content, false

	case strings.Contains(syntaxErr.Msg, "expected '>'"):
		// Fix unescaped > in attribute or text
		return f.fixUnescapedGreaterThanInLine(content, lineIndex, problemLine)

	default:
		// Unknown error type
		return content, false
	}
}

// fixUnescapedLessThanInLine fixes < characters that appear in text content
func (f *XMLFixer) fixUnescapedLessThanInLine(content string, lineIndex int, problemLine string) (string, bool) {
	// Strategy: Since we got a parsing error, we know there are problematic < characters
	// We'll be more conservative and only escape < that are clearly in text content

	lines := strings.Split(content, "\n")
	fixedLine := problemLine
	changed := false

	// Look for < characters that are inside text content (between > and <)
	// Pattern: >text<something>text< where the middle < should be escaped
	textContentRegex := regexp.MustCompile(`>([^<]*?<[^>]*?)<`)

	fixedLine = textContentRegex.ReplaceAllStringFunc(fixedLine, func(match string) string {
		// Extract the text content between > and <
		// Look for < characters within this text that should be escaped
		inner := match[1 : len(match)-1] // Remove > and < at ends

		// Escape any < characters in this text content
		if strings.Contains(inner, "<") {
			escapedInner := strings.ReplaceAll(inner, "<", "&lt;")
			changed = true

			f.fixes = append(f.fixes, XMLFix{
				Description: fmt.Sprintf("Escaped < character in text content on line %d", lineIndex+1),
				Before:      "<",
				After:       "&lt;",
			})

			return ">" + escapedInner + "<"
		}

		return match
	})

	if !changed {
		return content, false
	}

	// Reconstruct the content with the fixed line
	lines[lineIndex] = fixedLine
	return strings.Join(lines, "\n"), true
}

// fixUnescapedGreaterThanInLine fixes > characters in inappropriate contexts
func (f *XMLFixer) fixUnescapedGreaterThanInLine(content string, lineIndex int, problemLine string) (string, bool) {
	// This is trickier - > can appear in text content legally
	// For now, let's be conservative and only fix obvious cases

	// Look for > in text content (between tags)
	// Pattern: >text content with >more content<
	textWithGreaterRegex := regexp.MustCompile(`>([^<]*?)>([^<]*?)<`)

	if !textWithGreaterRegex.MatchString(problemLine) {
		return content, false
	}

	fixedLine := textWithGreaterRegex.ReplaceAllStringFunc(problemLine, func(match string) string {
		// Only fix if it looks like text content with a stray >
		parts := textWithGreaterRegex.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}

		beforeText := parts[1]
		afterText := parts[2]

		// Simple heuristic: if the text around > doesn't look like a tag, escape it
		if !strings.Contains(beforeText+afterText, "=") && // No attributes
			!regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`).MatchString(strings.TrimSpace(beforeText)) { // Not a tag name

			f.fixes = append(f.fixes, XMLFix{
				Description: fmt.Sprintf("Escaped invalid > character on line %d", lineIndex+1),
				Before:      ">",
				After:       "&gt;",
			})

			return ">" + beforeText + "&gt;" + afterText + "<"
		}

		return match
	})

	if fixedLine != problemLine {
		lines := strings.Split(content, "\n")
		lines[lineIndex] = fixedLine
		return strings.Join(lines, "\n"), true
	}

	return content, false
}

// fixAttributeQuotes fixes unescaped quotes in attribute values
func (f *XMLFixer) fixAttributeQuotes(content string) string {
	result := content

	// Pattern to match attribute values with unescaped quotes
	// This is simplified - a full implementation would need proper XML parsing
	attrPattern := `(\w+)="([^"]*)"`
	attrRegex := regexp.MustCompile(attrPattern)

	result = attrRegex.ReplaceAllStringFunc(result, func(match string) string {
		parts := attrRegex.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}

		attrName := parts[1]
		attrValue := parts[2]
		originalValue := attrValue

		// Only escape quotes that are not already escaped
		// This is a simple approach - could be more sophisticated
		attrValue = strings.ReplaceAll(attrValue, "&", "&amp;")
		attrValue = strings.ReplaceAll(attrValue, "<", "&lt;")
		attrValue = strings.ReplaceAll(attrValue, ">", "&gt;")

		if attrValue != originalValue {
			f.fixes = append(f.fixes, XMLFix{
				Description: fmt.Sprintf("Escaped special characters in attribute '%s'", attrName),
				Before:      originalValue,
				After:       attrValue,
			})
		}

		return fmt.Sprintf(`%s="%s"`, attrName, attrValue)
	})

	return result
}

// fixMalformedEntities fixes malformed character entities
func (f *XMLFixer) fixMalformedEntities(content string) string {
	result := content

	// Pattern to match malformed entities (& followed by alphanumeric but no semicolon)
	malformedEntityRegex := regexp.MustCompile(`&([a-zA-Z][a-zA-Z0-9]*)\s`)

	result = malformedEntityRegex.ReplaceAllStringFunc(result, func(match string) string {
		// Extract the entity name
		parts := malformedEntityRegex.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}

		entityName := parts[1]

		// Check if it's a known entity without semicolon
		knownEntities := map[string]string{
			"amp":  "&amp;",
			"lt":   "&lt;",
			"gt":   "&gt;",
			"quot": "&quot;",
			"apos": "&apos;",
		}

		if replacement, exists := knownEntities[entityName]; exists {
			f.fixes = append(f.fixes, XMLFix{
				Description: fmt.Sprintf("Fixed malformed entity &%s to %s;", entityName, entityName),
				Before:      "&" + entityName,
				After:       replacement,
			})
			return replacement + " "
		}

		// If it's not a known entity, escape the ampersand
		f.fixes = append(f.fixes, XMLFix{
			Description: fmt.Sprintf("Escaped unknown entity &%s", entityName),
			Before:      "&" + entityName,
			After:       "&amp;" + entityName,
		})
		return "&amp;" + entityName + " "
	})

	return result
}
