#!/bin/bash
# Build script for agentpm with version injection
# Handles version reading, git information, and build timestamp generation

set -e # Exit on any error
set -o pipefail # Exit on pipe errors

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
APP_NAME="agentpm"
MODULE="github.com/mindreframer/agentpm"
VERSION_PKG="$MODULE/cmd"
BUILD_DIR="build"
MAIN_FILE="main.go"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to read version from VERSION file
read_version() {
    local version_file="$PROJECT_ROOT/VERSION"
    
    if [ ! -f "$version_file" ]; then
        log_error "VERSION file not found at $version_file"
        return 1
    fi
    
    local version
    version=$(cat "$version_file" | tr -d '[:space:]')
    
    if [ -z "$version" ]; then
        log_error "VERSION file is empty"
        return 1
    fi
    
    # Basic semantic version validation (allows dev, beta, etc.)
    if ! echo "$version" | grep -E '^[0-9]+\.[0-9]+\.[0-9]+' >/dev/null && [ "$version" != "dev" ]; then
        log_warning "VERSION file contains non-standard version format: $version"
    fi
    
    echo "$version"
}

# Function to get git commit hash
get_git_commit() {
    local commit="unknown"
    local has_changes=false
    
    if command -v git >/dev/null 2>&1; then
        if git rev-parse --git-dir >/dev/null 2>&1; then
            commit=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
            
            # Check if there are uncommitted changes
            if ! git diff-index --quiet HEAD -- 2>/dev/null; then
                commit="${commit}-dirty"
                has_changes=true
            fi
        else
            log_warning "Not in a git repository, using 'unknown' for git commit"
        fi
    else
        log_warning "Git not available, using 'unknown' for git commit"
    fi
    
    # Store warning for later display
    if [ "$has_changes" = true ]; then
        export GIT_DIRTY_WARNING="true"
    fi
    
    echo "$commit"
}

# Function to get build timestamp
get_build_date() {
    date -u '+%Y-%m-%dT%H:%M:%SZ'
}

# Function to get Go version
get_go_version() {
    go version | awk '{print $3}' 2>/dev/null || echo "unknown"
}

# Function to validate build environment
validate_environment() {
    log_info "Validating build environment..."
    
    # Check if Go is available
    if ! command -v go >/dev/null 2>&1; then
        log_error "Go is not installed or not in PATH"
        return 1
    fi
    
    # Check if we're in the right directory
    if [ ! -f "$PROJECT_ROOT/go.mod" ]; then
        log_error "go.mod not found. Please run this script from the project root or ensure PROJECT_ROOT is correct."
        return 1
    fi
    
    # Check if main.go exists
    if [ ! -f "$PROJECT_ROOT/$MAIN_FILE" ]; then
        log_error "$MAIN_FILE not found in project root"
        return 1
    fi
    
    log_success "Build environment validation passed"
}

# Function to create build directory
setup_build_dir() {
    local build_path="$PROJECT_ROOT/$BUILD_DIR"
    
    if [ -d "$build_path" ]; then
        log_info "Cleaning existing build directory..."
        rm -rf "$build_path"
    fi
    
    log_info "Creating build directory: $build_path"
    mkdir -p "$build_path"
}

# Function to build the application
build_app() {
    local build_type="$1"
    local output_name="$2"
    local extra_flags="$3"
    
    log_info "Building $APP_NAME ($build_type)..."
    
    # Read version information
    local version
    version=$(read_version)
    if [ $? -ne 0 ]; then
        return 1
    fi
    
    local git_commit
    git_commit=$(get_git_commit)
    
    local build_date
    build_date=$(get_build_date)
    
    local go_version
    go_version=$(get_go_version)
    
    # Display build information
    log_info "Build Information:"
    echo "  Version: $version"
    echo "  Git Commit: $git_commit"
    echo "  Build Date: $build_date"
    echo "  Go Version: $go_version"
    
    # Show warning if repository has changes
    if [ "$GIT_DIRTY_WARNING" = "true" ]; then
        log_warning "Repository has uncommitted changes, marking as dirty"
    fi
    echo
    
    # Construct ldflags for version injection
    local ldflags="-ldflags"
    local version_ldflags="-X '${VERSION_PKG}.Version=${version}' -X '${VERSION_PKG}.GitCommit=${git_commit}' -X '${VERSION_PKG}.BuildDate=${build_date}'"
    
    # Add extra flags for release builds
    if [ "$build_type" = "release" ]; then
        version_ldflags="-s -w $version_ldflags"
    fi
    
    # Build the application
    local output_path="$PROJECT_ROOT/$BUILD_DIR/$output_name"
    
    cd "$PROJECT_ROOT"
    
    local build_cmd="go build $ldflags \"$version_ldflags\" $extra_flags -o \"$output_path\" $MAIN_FILE"
    
    log_info "Executing: $build_cmd"
    
    if eval "$build_cmd"; then
        log_success "Build completed: $output_path"
        
        # Verify the binary works
        if [ -x "$output_path" ]; then
            log_info "Verifying binary..."
            if "$output_path" version >/dev/null 2>&1; then
                log_success "Binary verification passed"
            else
                log_warning "Binary verification failed - version command not working"
            fi
        else
            log_error "Built binary is not executable"
            return 1
        fi
        
        return 0
    else
        log_error "Build failed"
        return 1
    fi
}

# Function to build for specific platform
build_cross_platform() {
    local goos="$1"
    local goarch="$2"
    local suffix="$3"
    
    log_info "Building for $goos/$goarch..."
    
    local output_name="$APP_NAME-$goos-$goarch$suffix"
    
    GOOS="$goos" GOARCH="$goarch" build_app "cross-platform" "$output_name" ""
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS] [BUILD_TYPE]"
    echo
    echo "Build Types:"
    echo "  dev       Development build (default)"
    echo "  release   Release build (optimized)"
    echo "  linux     Build for Linux"
    echo "  macos     Build for macOS (both amd64 and arm64)"
    echo "  windows   Build for Windows"
    echo "  all       Build for all platforms"
    echo
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -v, --verbose  Verbose output"
    echo "  -o, --output   Custom output name"
    echo "  --clean        Clean build directory before building"
    echo "  --version      Show version information only"
    echo
    echo "Examples:"
    echo "  $0                    # Development build"
    echo "  $0 release            # Release build"
    echo "  $0 --clean dev        # Clean and build development version"
    echo "  $0 -o myapp release   # Release build with custom name"
    echo "  $0 all                # Build for all platforms"
    echo
}

# Function to show version information
show_version_info() {
    local version
    version=$(read_version 2>/dev/null || echo "unknown")
    
    local git_commit
    git_commit=$(get_git_commit)
    
    local build_date
    build_date=$(get_build_date)
    
    local go_version
    go_version=$(get_go_version)
    
    echo "Build Script Version Information:"
    echo "  Version: $version"
    echo "  Git Commit: $git_commit"
    echo "  Build Date: $build_date"
    echo "  Go Version: $go_version"
}

# Main function
main() {
    local build_type="dev"
    local output_name="$APP_NAME"
    local verbose=false
    local clean=false
    local show_version=false
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -v|--verbose)
                verbose=true
                shift
                ;;
            -o|--output)
                output_name="$2"
                shift 2
                ;;
            --clean)
                clean=true
                shift
                ;;
            --version)
                show_version=true
                shift
                ;;
            dev|release|linux|macos|windows|all)
                build_type="$1"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # Show version information if requested
    if [ "$show_version" = true ]; then
        show_version_info
        exit 0
    fi
    
    # Enable verbose output if requested
    if [ "$verbose" = true ]; then
        set -x
    fi
    
    log_info "Starting $APP_NAME build process..."
    
    # Validate environment
    validate_environment
    if [ $? -ne 0 ]; then
        exit 1
    fi
    
    # Clean if requested
    if [ "$clean" = true ]; then
        log_info "Cleaning build artifacts..."
        rm -rf "$PROJECT_ROOT/$BUILD_DIR"
    fi
    
    # Setup build directory
    setup_build_dir
    
    # Build based on type
    case $build_type in
        dev)
            build_app "development" "$output_name" ""
            ;;
        release)
            build_app "release" "$output_name" ""
            ;;
        linux)
            build_cross_platform "linux" "amd64" ""
            ;;
        macos)
            build_cross_platform "darwin" "amd64" ""
            build_cross_platform "darwin" "arm64" ""
            ;;
        windows)
            build_cross_platform "windows" "amd64" ".exe"
            ;;
        all)
            build_cross_platform "linux" "amd64" ""
            build_cross_platform "darwin" "amd64" ""
            build_cross_platform "darwin" "arm64" ""
            build_cross_platform "windows" "amd64" ".exe"
            ;;
        *)
            log_error "Invalid build type: $build_type"
            show_usage
            exit 1
            ;;
    esac
    
    if [ $? -eq 0 ]; then
        log_success "Build process completed successfully!"
        
        # Show built files
        log_info "Built files:"
        ls -la "$PROJECT_ROOT/$BUILD_DIR/"
    else
        log_error "Build process failed!"
        exit 1
    fi
}

# CI/CD Environment Detection
detect_ci_environment() {
    if [ -n "$CI" ]; then
        log_info "CI/CD environment detected"
        
        if [ -n "$GITHUB_ACTIONS" ]; then
            log_info "Running in GitHub Actions"
        elif [ -n "$GITLAB_CI" ]; then
            log_info "Running in GitLab CI"
        elif [ -n "$JENKINS_URL" ]; then
            log_info "Running in Jenkins"
        else
            log_info "Running in unknown CI/CD environment"
        fi
        
        # Set appropriate defaults for CI
        set -x # Enable verbose output in CI
    fi
}

# Detect CI environment
detect_ci_environment

# Run main function
main "$@"