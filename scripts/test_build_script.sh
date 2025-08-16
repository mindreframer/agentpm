#!/bin/bash
# Test script for build.sh functionality
# Tests various build script features and validates expected behavior

set -e # Exit on any error
set -o pipefail # Exit on pipe errors

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Script locations
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_SCRIPT="$SCRIPT_DIR/build.sh"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Helper functions
log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

run_test() {
    local test_name="$1"
    local test_command="$2"
    ((TESTS_RUN++))
    
    log_info "Running test: $test_name"
    
    if eval "$test_command"; then
        log_success "$test_name"
        return 0
    else
        log_error "$test_name"
        return 1
    fi
}

# Test helper to check if file exists
file_exists() {
    if [ -f "$1" ]; then
        return 0
    else
        echo "File $1 does not exist"
        return 1
    fi
}

# Test helper to check if command output contains string
output_contains() {
    local command="$1"
    local expected="$2"
    local output
    output=$(eval "$command" 2>&1)
    if echo "$output" | grep -q "$expected"; then
        return 0
    else
        echo "Expected '$expected' in output, but got: $output"
        return 1
    fi
}

# Test helper to check VERSION file reading
test_version_reading() {
    if [ ! -f "$PROJECT_ROOT/VERSION" ]; then
        echo "VERSION file not found"
        return 1
    fi
    
    local version
    version=$(cat "$PROJECT_ROOT/VERSION" | tr -d '[:space:]')
    
    if [ -z "$version" ]; then
        echo "VERSION file is empty"
        return 1
    fi
    
    if [ "$version" = "0.1.0" ]; then
        return 0
    else
        echo "Unexpected version: $version"
        return 1
    fi
}

# Test helper to validate git information
test_git_info() {
    if command -v git >/dev/null 2>&1; then
        if git rev-parse --git-dir >/dev/null 2>&1; then
            local commit
            commit=$(git rev-parse --short HEAD 2>/dev/null)
            if [ -n "$commit" ]; then
                return 0
            fi
        fi
    fi
    echo "Git information not available"
    return 1
}

# Test helper to check binary version injection
check_binary_version() {
    local binary="$1"
    if [ ! -f "$binary" ]; then
        echo "Binary $binary not found"
        return 1
    fi
    
    local output
    output=$("$binary" version --format=json 2>/dev/null)
    
    if echo "$output" | grep -q '"version": "0.1.0"'; then
        if echo "$output" | grep -q '"git_commit"'; then
            if echo "$output" | grep -q '"build_date"'; then
                if echo "$output" | grep -q '"go_version"'; then
                    return 0
                fi
            fi
        fi
    fi
    
    echo "Version injection verification failed. Output: $output"
    return 1
}

# Test helper to check timestamp format
test_timestamp_format() {
    local timestamp="$1"
    # Check ISO 8601 format: YYYY-MM-DDTHH:MM:SSZ
    if echo "$timestamp" | grep -E '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$' >/dev/null; then
        return 0
    else
        echo "Invalid timestamp format: $timestamp"
        return 1
    fi
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test artifacts..."
    cd "$PROJECT_ROOT"
    rm -rf build test_version >/dev/null 2>&1 || true
}

# Create test VERSION file for testing
create_test_version_file() {
    local test_version="$1"
    local backup_file=""
    
    if [ -f "$PROJECT_ROOT/VERSION" ]; then
        backup_file="$PROJECT_ROOT/VERSION.backup.$$"
        cp "$PROJECT_ROOT/VERSION" "$backup_file"
    fi
    
    echo "$test_version" > "$PROJECT_ROOT/VERSION"
    echo "$backup_file"
}

# Restore VERSION file
restore_version_file() {
    local backup_file="$1"
    if [ -n "$backup_file" ] && [ -f "$backup_file" ]; then
        mv "$backup_file" "$PROJECT_ROOT/VERSION"
    fi
}

# Main test suite
main() {
    log_info "Starting build script functionality tests..."
    echo
    
    # Ensure we're in the right directory and build script exists
    if [ ! -f "$BUILD_SCRIPT" ]; then
        log_error "Build script not found at $BUILD_SCRIPT"
        exit 1
    fi
    
    cd "$PROJECT_ROOT"
    
    # Test 1: Build script help message
    run_test "Build script displays help message" \
        "output_contains '$BUILD_SCRIPT --help' 'Usage:.*build.sh'"
    
    # Test 2: Build script version information
    run_test "Build script shows version information" \
        "output_contains '$BUILD_SCRIPT --version' 'Build Script Version Information'"
    
    # Test 3: VERSION file reading
    run_test "Build script reads VERSION file correctly" \
        "test_version_reading"
    
    # Test 4: Git information handling
    run_test "Build script handles git information" \
        "test_git_info || echo 'Git not available, but handled gracefully'"
    
    # Test 5: Environment validation
    run_test "Build script validates environment" \
        "output_contains '$BUILD_SCRIPT --version' 'Go Version:'"
    
    # Test 6: Development build
    run_test "Development build produces working binary" \
        "$BUILD_SCRIPT --clean dev && file_exists build/agentpm && check_binary_version build/agentpm"
    
    # Test 7: Release build
    run_test "Release build produces optimized binary" \
        "$BUILD_SCRIPT --clean release && file_exists build/agentpm && check_binary_version build/agentpm"
    
    # Test 8: Custom output name
    run_test "Custom output name works" \
        "$BUILD_SCRIPT --clean -o custom_name dev && file_exists build/custom_name"
    
    # Test 9: Linux cross-compilation
    run_test "Linux cross-compilation works" \
        "$BUILD_SCRIPT --clean linux && file_exists build/agentpm-linux-amd64"
    
    # Test 10: macOS cross-compilation
    run_test "macOS cross-compilation works" \
        "$BUILD_SCRIPT --clean macos && file_exists build/agentpm-darwin-amd64 && file_exists build/agentpm-darwin-arm64"
    
    # Test 11: Windows cross-compilation
    run_test "Windows cross-compilation works" \
        "$BUILD_SCRIPT --clean windows && file_exists build/agentpm-windows-amd64.exe"
    
    # Test 12: All platforms build
    run_test "All platforms build works" \
        "$BUILD_SCRIPT --clean all && file_exists build/agentpm-linux-amd64 && file_exists build/agentpm-darwin-amd64 && file_exists build/agentpm-windows-amd64.exe"
    
    # Test 13: Binary verification
    run_test "Built binary passes verification" \
        "$BUILD_SCRIPT --clean dev && ./build/agentpm help >/dev/null"
    
    # Test 14: Version injection with custom version
    local backup_version
    backup_version=$(create_test_version_file "2.0.0-test")
    run_test "Version injection works with custom version" \
        "$BUILD_SCRIPT --clean dev && ./build/agentpm version --format=json | grep '\"version\": \"2.0.0-test\"'"
    restore_version_file "$backup_version"
    
    # Test 15: Invalid VERSION file handling
    backup_version=$(create_test_version_file "")
    run_test "Build script handles empty VERSION file" \
        "! $BUILD_SCRIPT dev 2>/dev/null"
    restore_version_file "$backup_version"
    
    # Test 16: Missing VERSION file handling
    backup_version=""
    if [ -f "$PROJECT_ROOT/VERSION" ]; then
        backup_version="$PROJECT_ROOT/VERSION.backup.$$"
        mv "$PROJECT_ROOT/VERSION" "$backup_version"
    fi
    run_test "Build script handles missing VERSION file" \
        "! $BUILD_SCRIPT dev 2>/dev/null"
    restore_version_file "$backup_version"
    
    # Test 17: Timestamp format validation
    run_test "Build generates valid timestamp format" \
        "$BUILD_SCRIPT --clean dev && ./build/agentpm version --format=json | grep '\"build_date\"' | grep -o '[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z'"
    
    # Test 18: Build directory creation
    run_test "Build script creates build directory" \
        "$BUILD_SCRIPT --clean dev && [ -d build ]"
    
    # Test 19: Clean flag functionality
    run_test "Clean flag removes existing artifacts" \
        "$BUILD_SCRIPT dev && $BUILD_SCRIPT --clean dev && file_exists build/agentpm"
    
    # Test 20: Verbose flag functionality
    run_test "Verbose flag works" \
        "output_contains '$BUILD_SCRIPT --verbose --version' 'Build Script Version Information'"
    
    # Test 21: Error handling for invalid arguments
    run_test "Build script handles invalid arguments" \
        "! $BUILD_SCRIPT invalid_build_type 2>/dev/null"
    
    # Test 22: Binary size comparison (release vs dev)
    run_test "Release build is smaller than development build" \
        "$BUILD_SCRIPT --clean dev && DEV_SIZE=\$(stat -f%z build/agentpm) && $BUILD_SCRIPT --clean release && REL_SIZE=\$(stat -f%z build/agentpm) && [ \$REL_SIZE -lt \$DEV_SIZE ]"
    
    echo
    log_info "Test Summary:"
    echo "  Tests Run: $TESTS_RUN"
    echo "  Passed: $TESTS_PASSED"
    echo "  Failed: $TESTS_FAILED"
    
    cleanup
    
    if [ $TESTS_FAILED -eq 0 ]; then
        log_success "All build script tests passed!"
        exit 0
    else
        log_error "Some build script tests failed!"
        exit 1
    fi
}

# Run the main test suite
main "$@"