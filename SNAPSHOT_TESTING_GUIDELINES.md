# Snapshot Testing Guidelines

## Overview

This project uses snapshot testing for XML output validation, replacing fragile string assertions with comprehensive snapshot comparisons. This document provides guidelines for using and maintaining snapshot tests.

## Quick Start

### Running Tests with Snapshots

```bash
# Run all tests (includes snapshot validation)
make test

# Run only snapshot-related tests
make test-snapshots

# Update all snapshots after intentional changes
make update-snapshots
```

### Creating Snapshot Tests

```go
import apmtesting "github.com/mindreframer/agentpm/internal/testing"

func TestCommandXMLOutput(t *testing.T) {
    // ... execute command and capture output ...
    
    // Replace multiple assert.Contains with single snapshot assertion
    snapshotTester := apmtesting.NewSnapshotTester()
    snapshotTester.MatchXMLSnapshot(t, output, "command_xml_output")
}
```

## Migration from String Assertions

### Before (Fragile)
```go
assert.Contains(t, output, `<epic_started epic="epic-1">`)
assert.Contains(t, output, `<previous_status>pending</previous_status>`)
assert.Contains(t, output, `<new_status>wip</new_status>`)
assert.Contains(t, output, `<event_created>false</event_created>`)
```

### After (Robust)
```go
snapshotTester := apmtesting.NewSnapshotTester()
snapshotTester.MatchXMLSnapshot(t, output, "epic_started_xml_output")
```

## Benefits

1. **Comprehensive Coverage**: Captures entire XML structure, not just fragments
2. **Regression Detection**: Automatically detects any changes in output format
3. **Maintainability**: Single assertion instead of multiple fragile checks
4. **Normalization**: Timestamps and dynamic content automatically normalized
5. **Readability**: Snapshots provide clear view of expected output

## Best Practices

### Snapshot Naming
- Use descriptive names: `"start_epic_xml_output"` not `"test1"`
- Include command and format: `"switch_command_xml_output"`
- Be consistent within test files

### When to Update Snapshots
- After intentional XML format changes
- When adding new fields to output
- When fixing bugs that change expected output
- **Never** update snapshots to make failing tests pass without review

### Reviewing Snapshot Changes
1. Always review snapshot diffs carefully
2. Ensure changes are intentional and correct
3. Verify timestamps are properly normalized
4. Check for unexpected structural changes

## XML Normalization Features

The snapshot system automatically normalizes:
- **Timestamps**: Replaced with `[TIMESTAMP]` placeholder
- **Dynamic IDs**: Can be configured for normalization
- **Whitespace**: Consistent formatting for comparison
- **Attribute Order**: Sorted for consistent comparison

## Commands Reference

```bash
# Update snapshots after code changes
go test ./... -args -test.update-snapshots

# Run specific test with snapshot validation
go test ./cmd -run TestStartEpicCommand_XMLOutput

# View snapshot files
find . -name "*.snap" -exec cat {} \;
```

## Snapshot File Structure

Snapshots are stored in `__snapshots__/` directories next to test files:

```
cmd/
├── start_epic_test.go
├── __snapshots__/
│   └── start_epic_test.snap
```

Each snapshot file contains:
```
[TestFunctionName - 1]
snapshot_name
<xml_content>
    <normalized>true</normalized>
    <timestamp>[TIMESTAMP]</timestamp>
</xml_content>
---
```

## Troubleshooting

### Snapshot Test Failures
1. **Review the diff**: Check what changed in the output
2. **Verify intentionality**: Are changes expected?
3. **Update if correct**: Run `make update-snapshots`
4. **Investigate if unexpected**: Debug the underlying issue

### Missing Snapshots
- First run of a test creates the snapshot
- Look for green "✎ Snapshot added" message
- Commit new snapshot files with your changes

## Integration with CI/CD

Snapshot tests integrate seamlessly with CI pipelines:
- Tests fail if output doesn't match snapshots
- No special configuration needed
- Snapshots should be committed to version control
- CI should never auto-update snapshots

## Epic 9 Migration Status

This snapshot testing system was introduced in Epic 9 Phase 4C. Key migrations completed:

- `TestStartEpicCommand_XMLOutput` - Epic start command
- `TestSwitchCommand_XMLOutput` - Epic switch command
- Migration validation tests added
- Build system integration completed

## Future Considerations

- Extend normalization for file paths if needed
- Add snapshot testing for JSON output formats
- Consider snapshot testing for large text outputs
- Integrate with performance testing for output size validation