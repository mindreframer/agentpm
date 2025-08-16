#!/bin/bash
# Test script for build-time variable injection mechanism

set -e

# Test injection by building and checking version
echo "Testing build-time variable injection..."

cd "$(dirname "$0")/.."

# Build with make
make build > /dev/null 2>&1

# Test that injected variables are not defaults
VERSION_OUTPUT=$(./build/agentpm version --format=json)

# Check version
if echo "$VERSION_OUTPUT" | grep -q '"version": "0.1.0"'; then
    echo "✓ Version injection successful"
else
    echo "✗ Version injection failed"
    exit 1
fi

# Check git commit (should not be "unknown")
if echo "$VERSION_OUTPUT" | grep -q '"git_commit": "unknown"'; then
    echo "✗ Git commit injection failed - still showing default"
    exit 1
else
    echo "✓ Git commit injection successful"
fi

# Check build date (should not be "unknown")
if echo "$VERSION_OUTPUT" | grep -q '"build_date": "unknown"'; then
    echo "✗ Build date injection failed - still showing default"
    exit 1
else
    echo "✓ Build date injection successful"
fi

# Check Go version
if echo "$VERSION_OUTPUT" | grep -q '"go_version": "go'; then
    echo "✓ Go version detection successful"
else
    echo "✗ Go version detection failed"
    exit 1
fi

echo "All variable injection tests passed!"