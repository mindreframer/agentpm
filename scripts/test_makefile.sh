#!/bin/bash
# Test script for Makefile functionality
# Tests various Makefile targets and verifies expected behavior

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

# Test helper to check if directory exists
dir_exists() {
    if [ -d "$1" ]; then
        return 0
    else
        echo "Directory $1 does not exist"
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

# Test helper to check binary version information
check_version_injection() {
    local binary="$1"
    if [ ! -f "$binary" ]; then
        echo "Binary $binary not found"
        return 1
    fi
    
    # Test that version shows injected information (not defaults)
    local version_output
    version_output=$("$binary" version 2>&1)
    
    if echo "$version_output" | grep -q "agentpm version 0.1.0"; then
        if echo "$version_output" | grep -q "Git commit:"; then
            if echo "$version_output" | grep -q "Built:"; then
                if echo "$version_output" | grep -q "Go version:"; then
                    return 0
                fi
            fi
        fi
    fi
    
    echo "Version injection failed. Output: $version_output"
    return 1
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test artifacts..."
    make clean >/dev/null 2>&1 || true
    rm -f coverage.out coverage.html >/dev/null 2>&1 || true
}

# Main test suite
main() {
    log_info "Starting Makefile functionality tests..."
    echo
    
    # Ensure we're in the right directory
    if [ ! -f "Makefile" ]; then
        log_error "Makefile not found. Please run this script from the project root."
        exit 1
    fi
    
    # Test 1: Help target displays all available targets
    run_test "make help displays all available targets" \
        "output_contains 'make help' 'build.*Build the application with version injection'"
    
    # Test 2: Version target shows current version
    run_test "make version shows current version" \
        "output_contains 'make version' 'Version: 0.1.0'"
    
    # Test 3: Clean target removes build artifacts
    run_test "make clean removes build artifacts" \
        "make clean && ! dir_exists build"
    
    # Test 4: Build target produces working binary with correct version
    run_test "make build produces working binary with correct version" \
        "make build && file_exists build/agentpm && check_version_injection build/agentpm"
    
    # Test 5: Dev build target works
    run_test "make dev produces functional development build" \
        "make clean && make dev && file_exists build/agentpm && check_version_injection build/agentpm"
    
    # Test 6: Release build target works
    run_test "make release produces optimized build" \
        "make clean && make release && file_exists build/agentpm && check_version_injection build/agentpm"
    
    # Test 7: Test target runs all tests successfully
    run_test "make test runs all tests successfully" \
        "make test"
    
    # Test 8: Format target works
    run_test "make fmt formats code successfully" \
        "make fmt"
    
    # Test 9: Dependencies target works
    run_test "make deps installs dependencies successfully" \
        "make deps"
    
    # Test 10: Check version file requirement
    run_test "make check-version validates VERSION file" \
        "make check-version"
    
    # Test 11: Cross-platform Linux build
    run_test "make build-linux produces Linux binary" \
        "make build-linux && file_exists build/agentpm-linux-amd64"
    
    # Test 12: Cross-platform macOS build  
    run_test "make build-macos produces macOS binaries" \
        "make build-macos && file_exists build/agentpm-macos-amd64 && file_exists build/agentpm-macos-arm64"
    
    # Test 13: Cross-platform Windows build
    run_test "make build-windows produces Windows binary" \
        "make build-windows && file_exists build/agentpm-windows-amd64.exe"
    
    # Test 14: Build all platforms
    run_test "make build-all produces all platform binaries" \
        "make build-all && file_exists build/agentpm-linux-amd64 && file_exists build/agentpm-macos-amd64 && file_exists build/agentpm-windows-amd64.exe"
    
    # Test 15: Build with checksums
    run_test "make build-checksums generates checksums" \
        "make build-checksums && file_exists build/checksums.txt"
    
    # Test 16: Version injection variables are set correctly
    run_test "Version injection sets all required variables" \
        "make build && output_contains './build/agentpm version --format=json' '\"version\": \"0.1.0\"' && output_contains './build/agentpm version --format=json' '\"git_commit\"' && output_contains './build/agentpm version --format=json' '\"build_date\"'"
    
    # Test 17: JSON version output is valid
    run_test "Version command produces valid JSON" \
        "make build && ./build/agentpm version --format=json | python3 -m json.tool >/dev/null"
    
    # Test 18: XML version output is valid
    run_test "Version command produces valid XML" \
        "make build && ./build/agentpm version --format=xml | head -1 | grep -q 'xml version'"
    
    # Test 19: Binary works after installation simulation
    run_test "Built binary executes commands correctly" \
        "make build && ./build/agentpm help >/dev/null"
    
    # Test 20: Verify build reproducibility
    run_test "Build process is deterministic" \
        "make build && cp build/agentpm build/agentpm.backup && make clean && make build && cmp build/agentpm build/agentpm.backup"
    
    echo
    log_info "Test Summary:"
    echo "  Tests Run: $TESTS_RUN"
    echo "  Passed: $TESTS_PASSED"
    echo "  Failed: $TESTS_FAILED"
    
    if [ $TESTS_FAILED -eq 0 ]; then
        log_success "All Makefile tests passed!"
        cleanup
        exit 0
    else
        log_error "Some Makefile tests failed!"
        cleanup
        exit 1
    fi
}

# Run the main test suite
main "$@"