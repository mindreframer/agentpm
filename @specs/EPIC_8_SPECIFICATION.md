# EPIC-8: Version Command Implementation

## Overview

**Epic ID:** 8  
**Name:** Version Command Implementation  
**Status:** pending  
**Priority:** low  

**Goal:** Implement a comprehensive version command with build-time version injection, supporting multiple output formats and proper build tooling integration.

## Implementation Tasks Required

### 🔥 **High Priority: Core Version Command**

#### 1. **Implement `agentpm version` command** 
- Create `cmd/version.go` with urfave/cli/v3 framework
- Support `--format` flag (text, json, xml)
- Display version, git commit, build date, Go version
- Follow existing CLI command patterns
- Example: `agentpm version --format=json`

#### 2. **Build-time Version Injection**
- Use Go `-ldflags` for build-time variable injection
- Read version from existing `VERSION` file
- Inject git commit hash and build timestamp
- Embed version information directly in binary
- No runtime file dependencies

### 🏗️ **Medium Priority: Build Tooling**

#### 3. **Create comprehensive Makefile**
- `make build` - Build with version injection
- `make test` - Run all tests
- `make clean` - Clean build artifacts
- `make install` - Install to system
- `make dev` - Development build
- `make release` - Release build with optimizations
- Cross-platform build targets

#### 4. **Build script integration**
- Create `scripts/build.sh` for automated builds
- Proper error handling and validation
- Support for CI/CD environments
- Version validation from VERSION file

### 📊 **Low Priority: Enhanced Features**

#### 5. **Output format support**
- **Text format** (default): Human-readable version info
- **JSON format**: Machine-readable structured output
- **XML format**: XML structured output for consistency
- Proper error handling for invalid formats

#### 6. **Build metadata enhancement**
- Git commit hash (short and full)
- Build timestamp in ISO 8601 format
- Go version information
- Build environment details (optional)

## Technical Requirements

### Command Implementation
- Follow existing CLI patterns using `github.com/urfave/cli/v3`
- Implement comprehensive test coverage for version command
- Support global `--format` flag consistency
- Proper error handling for build-time injection failures

### Build System Enhancement
- Makefile with standard targets (build, test, clean, install)
- Cross-platform support (Linux, macOS, Windows)
- Development vs release build configurations
- Version validation and build verification

### Version Injection Mechanism
- Use Go build flags (`-ldflags -X`) for variable injection
- Target package-level variables in `cmd/version.go`
- Fallback values for development builds
- Proper escaping for special characters in version strings

## Success Criteria

### Commands
- [ ] `agentpm version` displays current version information
- [ ] `agentpm version --format=json` outputs valid JSON
- [ ] `agentpm version --format=xml` outputs valid XML
- [ ] Version information includes all required metadata
- [ ] All version command tests pass

### Build System
- [ ] `make build` produces working binary with correct version
- [ ] `make test` runs all tests successfully
- [ ] `make clean` removes all build artifacts
- [ ] `make install` installs to system location
- [ ] Cross-platform builds work correctly

### Integration
- [ ] Version injected at build time (not runtime)
- [ ] No runtime dependencies on VERSION file
- [ ] Build works in CI/CD environments
- [ ] Development builds have sensible defaults
- [ ] Release builds include all metadata

## Dependencies

- Epic 1: Foundation & Configuration (completed)
- Epic 2: Query Commands (completed)
- Epic 3: Epic Lifecycle (completed)
- Epic 4: Task & Phase Management (completed)  
- Epic 5: Test Management & Event Logging (completed)
- Epic 6: Handoff & Documentation (completed)
- Epic 7: Missing Features Implementation (completed)

## Estimated Effort

**Core Version Command:** 4-6 hours
**Build Tooling & Makefile:** 3-4 hours  
**Testing & Integration:** 2-3 hours
**Total Epic:** 1-2 days

## Implementation Details

### File Structure
```
├── cmd/
│   └── version.go              # New version command
├── scripts/
│   └── build.sh               # Build script with version injection
├── Makefile                   # Comprehensive build targets
├── VERSION                    # Existing version file
└── main.go                    # Updated with version command registration
```

### Version Command Interface
```bash
# Basic usage
agentpm version

# JSON output
agentpm version --format=json

# XML output  
agentpm version --format=xml
```

### Expected Output Examples

**Text Format:**
```
agentpm version 1.0.0
Git commit: abc123de
Built: 2025-08-16T14:30:00Z
Go version: go1.21.0
```

**JSON Format:**
```json
{
  "version": "1.0.0",
  "git_commit": "abc123de",
  "build_date": "2025-08-16T14:30:00Z", 
  "go_version": "go1.21.0"
}
```

**XML Format:**
```xml
<version_info>
    <version>1.0.0</version>
    <git_commit>abc123de</git_commit>
    <build_date>2025-08-16T14:30:00Z</build_date>
    <go_version>go1.21.0</go_version>
</version_info>
```

### Makefile Targets
```makefile
build:          Build the application with version injection
test:           Run all tests
clean:          Clean build artifacts
install:        Install to system location
dev:            Development build (fast, no optimizations)
release:        Release build (optimized, stripped)
lint:           Run code linting
fmt:            Format code
deps:           Install dependencies
help:           Show available targets
```

### Build-time Variable Injection
```go
// Variables injected at build time via ldflags
var (
    Version   = "dev"      // From VERSION file
    GitCommit = "unknown"  // From git rev-parse
    BuildDate = "unknown"  // From build timestamp
)
```

## Notes

This epic focuses on developer experience and build tooling enhancement. The version command provides essential metadata for debugging and support, while the Makefile standardizes the build process across different environments. The build-time injection ensures the binary is self-contained and doesn't rely on external files at runtime.

The implementation follows Go best practices for version management and integrates seamlessly with existing CLI patterns established in previous epics.