#!/bin/bash
# Integration test for version command within CLI framework

set -e

echo "Testing version command integration with CLI framework..."

cd "$(dirname "$0")/.."

# Build the binary
make build > /dev/null 2>&1

# Test version command is available in help
if ./build/agentpm help | grep -q "version.*Display version information"; then
    echo "✓ Version command appears in help"
else
    echo "✗ Version command not found in help"
    exit 1
fi

# Test version command works
if ./build/agentpm version > /dev/null 2>&1; then
    echo "✓ Version command executes successfully"
else
    echo "✗ Version command execution failed"
    exit 1
fi

# Test version command aliases
if ./build/agentpm ver > /dev/null 2>&1; then
    echo "✓ Version command alias 'ver' works"
else
    echo "✗ Version command alias 'ver' failed"
    exit 1
fi

# Test format flag consistency with global flag
if ./build/agentpm version --format=json > /dev/null 2>&1; then
    echo "✓ Version command supports format flag"
else
    echo "✗ Version command format flag failed"
    exit 1
fi

# Test that version command output is consistent
OUTPUT1=$(./build/agentpm version)
OUTPUT2=$(./build/agentpm version)
if [ "$OUTPUT1" = "$OUTPUT2" ]; then
    echo "✓ Version command output is consistent"
else
    echo "✗ Version command output is inconsistent"
    exit 1
fi

echo "All CLI integration tests passed!"