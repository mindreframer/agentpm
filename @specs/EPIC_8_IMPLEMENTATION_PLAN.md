# EPIC-8: Version Command Implementation Plan
## Test-Driven Development Approach

### Phase 1: Version Command Foundation + Tests (High Priority)

#### Phase 1A: Create Version Command Foundation
- [ ] Create cmd/version.go command with urfave/cli/v3 framework
- [ ] Define build-time injectable variables (Version, GitCommit, BuildDate)
- [ ] Implement basic version display functionality
- [ ] Add --format flag support for text/json/xml output formats
- [ ] Create VersionInfo struct for structured data
- [ ] Implement format validation and error handling
- [ ] Add Go runtime version detection

#### Phase 1B: Write Version Command Tests **IMMEDIATELY AFTER 1A**
Epic 8 Test Scenarios Covered:
- [ ] **Test: Version command displays default text format**
- [ ] **Test: Version command with --format=json outputs valid JSON**
- [ ] **Test: Version command with --format=xml outputs valid XML** 
- [ ] **Test: Version command with invalid format shows error**
- [ ] **Test: Version variables default values work in dev mode**
- [ ] **Test: All required version metadata is included**
- [ ] **Test: Command structure follows existing CLI patterns**

#### Phase 1C: Output Format Implementation
- [ ] Implement text format output (human-readable)
- [ ] Implement JSON format output (machine-readable)
- [ ] Implement XML format output (consistency with other commands)
- [ ] Add proper escaping for special characters in version strings
- [ ] Format validation and error handling
- [ ] Consistent output styling with existing commands

#### Phase 1D: Write Output Format Tests **IMMEDIATELY AFTER 1C**
Epic 8 Test Scenarios Covered:
- [ ] **Test: Text format produces human-readable output**
- [ ] **Test: JSON format produces valid, parseable JSON**
- [ ] **Test: XML format produces valid, parseable XML**
- [ ] **Test: Special characters in version strings are escaped properly**
- [ ] **Test: Format validation rejects invalid formats**
- [ ] **Test: Output consistency with global --format flag behavior**

### Phase 2: Build System & Makefile Implementation + Tests (High Priority)

#### Phase 2A: Makefile Creation
- [ ] Create comprehensive Makefile with standard targets
- [ ] Implement build target with version injection via ldflags
- [ ] Add test target for running all test suites
- [ ] Add clean target for removing build artifacts
- [ ] Add install target for system installation
- [ ] Add dev target for fast development builds
- [ ] Add release target for optimized production builds
- [ ] Add help target showing available commands

#### Phase 2B: Write Makefile Tests **IMMEDIATELY AFTER 2A**
Epic 8 Test Scenarios Covered:
- [ ] **Test: make build produces working binary with correct version**
- [ ] **Test: make test runs all tests successfully**
- [ ] **Test: make clean removes all build artifacts**
- [ ] **Test: make install works in test environment**
- [ ] **Test: make dev produces functional development build**
- [ ] **Test: make release produces optimized build**
- [ ] **Test: make help displays all available targets**

#### Phase 2C: Build Script Implementation
- [ ] Create scripts/build.sh for automated builds
- [ ] Implement VERSION file reading and validation
- [ ] Add git commit hash detection (with fallback for non-git)
- [ ] Add build timestamp generation in ISO 8601 format
- [ ] Implement proper error handling and validation
- [ ] Add support for CI/CD environment detection
- [ ] Cross-platform compatibility (Linux, macOS, Windows)

#### Phase 2D: Write Build Script Tests **IMMEDIATELY AFTER 2C**
Epic 8 Test Scenarios Covered:
- [ ] **Test: Build script reads VERSION file correctly**
- [ ] **Test: Build script handles missing git repository gracefully**
- [ ] **Test: Build script generates proper timestamp format**
- [ ] **Test: Build script validation catches invalid VERSION file**
- [ ] **Test: Build script works in CI/CD environments**
- [ ] **Test: Cross-platform build script compatibility**

### Phase 3: Version Injection & Integration + Tests (Medium Priority)

#### Phase 3A: Build-time Variable Injection
- [ ] Implement ldflags mechanism for version injection
- [ ] Set up proper Go module path targeting for variable injection
- [ ] Add fallback values for development builds
- [ ] Implement version string validation and sanitization
- [ ] Add build metadata collection (git status, build environment)
- [ ] Handle special characters and spaces in version strings
- [ ] Test injection mechanism with various version formats

#### Phase 3B: Write Variable Injection Tests **IMMEDIATELY AFTER 3A**
Epic 8 Test Scenarios Covered:
- [ ] **Test: ldflags injection works with standard version formats**
- [ ] **Test: Fallback values work when injection fails**
- [ ] **Test: Special characters in version strings are handled**
- [ ] **Test: Build metadata is properly included**
- [ ] **Test: Development builds have sensible defaults**
- [ ] **Test: Injection works across different Go module structures**

#### Phase 3C: Integration with Main CLI
- [ ] Register version command in main.go commands list
- [ ] Ensure version command follows global flag patterns
- [ ] Add version command to help system
- [ ] Test integration with existing CLI framework
- [ ] Verify consistent behavior with other commands
- [ ] Add version command to command completion if applicable

#### Phase 3D: Write Integration Tests **IMMEDIATELY AFTER 3C**
Epic 8 Test Scenarios Covered:
- [ ] **Test: Version command integrates properly with CLI framework**
- [ ] **Test: Global flags work consistently with version command**
- [ ] **Test: Help system includes version command**
- [ ] **Test: Command completion includes version command**
- [ ] **Test: Version command doesn't interfere with other commands**
- [ ] **Test: CLI framework compatibility maintained**

### Phase 4: Cross-platform & Advanced Features + Tests (Low Priority)

#### Phase 4A: Cross-platform Build Support
- [ ] Add cross-compilation targets to Makefile
- [ ] Implement GOOS/GOARCH support for multiple platforms
- [ ] Add Windows-specific build considerations
- [ ] Create platform-specific binary naming conventions
- [ ] Add checksums and build verification
- [ ] Implement parallel builds for multiple platforms

#### Phase 4B: Write Cross-platform Tests **IMMEDIATELY AFTER 4A**
Epic 8 Test Scenarios Covered:
- [ ] **Test: Cross-compilation produces working binaries**
- [ ] **Test: Platform-specific binary naming works correctly**
- [ ] **Test: Build checksums are generated properly**
- [ ] **Test: Windows builds work in Windows environment**
- [ ] **Test: Parallel builds complete successfully**
- [ ] **Test: Build verification catches corrupted binaries**

#### Phase 4C: Enhanced Build Features
- [ ] Add build caching for faster incremental builds
- [ ] Implement build reproducibility features
- [ ] Add build environment detection and reporting
- [ ] Create build artifact management
- [ ] Add semantic version validation
- [ ] Implement automated version bumping utilities

#### Phase 4D: Write Enhanced Features Tests **IMMEDIATELY AFTER 4C**
Epic 8 Test Scenarios Covered:
- [ ] **Test: Build caching improves build performance**
- [ ] **Test: Reproducible builds produce identical artifacts**
- [ ] **Test: Build environment detection works correctly**
- [ ] **Test: Semantic version validation catches invalid formats**
- [ ] **Test: Version bumping utilities work properly**
- [ ] **Test: Build artifacts are managed correctly**

## Development Workflow Per Phase

For **EACH** phase:

1. **Implement Code** (Phase XA or XC)
2. **Write Tests IMMEDIATELY** (Phase XB or XD) 
3. **Run Tests & Verify** - All tests must pass
4. **Run Linting/Type Checking** - Code must be clean
5. **NEVER move to next phase with failing tests**

## Epic 8 Specific Considerations

### Dependencies & Integration
- **Epic 1:** CLI framework, configuration management, existing command patterns
- **Epic 2:** Query service patterns for consistent output formatting
- **Build System:** VERSION file, Git repository, Go toolchain
- **Testing:** Build artifact verification, cross-platform compatibility

### Technical Requirements
- **Build-time Injection:** Use Go ldflags for embedding version information
- **Cross-platform:** Support Linux, macOS, Windows build targets
- **Format Consistency:** JSON/XML output follows existing command patterns
- **CI/CD Friendly:** Build process works in automated environments
- **Version Validation:** Proper handling of semantic versions and edge cases

### Priority Implementation Order
1. **Version Command** - Core functionality for displaying version information
2. **Makefile & Build System** - Essential developer experience improvements
3. **Build-time Injection** - Professional-grade version management
4. **Cross-platform Support** - Broader deployment compatibility

### File Structure
```
├── cmd/
│   └── version.go              # Version command implementation
├── scripts/
│   └── build.sh               # Build script with version injection
├── Makefile                   # Comprehensive build targets
├── VERSION                    # Existing version file
├── main.go                    # Updated with version command
└── testdata/
    ├── test-version.txt       # Version format test data
    ├── test-build.sh          # Build script test fixtures
    └── sample-builds/         # Cross-platform build test artifacts
```

### Version Command Implementation
```go
type VersionInfo struct {
    Version   string `json:"version"`
    GitCommit string `json:"git_commit"`
    BuildDate string `json:"build_date"`
    GoVersion string `json:"go_version"`
}

// Variables injected at build time via ldflags
var (
    Version   = "dev"      // From VERSION file
    GitCommit = "unknown"  // From git rev-parse --short HEAD
    BuildDate = "unknown"  // From date -u '+%Y-%m-%dT%H:%M:%SZ'
)

func VersionCommand() *cli.Command {
    return &cli.Command{
        Name:  "version",
        Usage: "Display version information",
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name:    "format",
                Aliases: []string{"F"},
                Usage:   "Output format: text, json, xml",
                Value:   "text",
            },
        },
        Action: executeVersionCommand,
    }
}
```

### Makefile Target Structure
```makefile
# Core targets
build:          Build application with version injection
test:           Run all tests
clean:          Clean build artifacts
install:        Install to system location
dev:            Development build (fast, no optimizations)
release:        Release build (optimized, stripped)

# Development targets
lint:           Run code linting
fmt:            Format code
deps:           Install dependencies

# Cross-platform targets
build-linux:    Build for Linux
build-macos:    Build for macOS  
build-windows:  Build for Windows
build-all:      Build for all platforms

# Utility targets
version:        Show current version
help:           Show available targets
```

## Benefits of This Approach

✅ **Immediate Feedback** - Catch issues as soon as code is written  
✅ **Working Code** - Each phase delivers tested, working functionality  
✅ **Epic 8 Coverage** - All specification requirements distributed across phases  
✅ **Incremental Progress** - Version command usable after each phase  
✅ **Risk Mitigation** - Problems caught early, not at the end  
✅ **Quality Assurance** - No untested code makes it to later phases  
✅ **Developer Experience** - Professional build tooling and version management  

## Test Distribution Summary

- **Phase 1 Tests:** 13 scenarios (Version command foundation, output formats)
- **Phase 2 Tests:** 13 scenarios (Makefile, build scripts, automation)
- **Phase 3 Tests:** 12 scenarios (Variable injection, CLI integration)
- **Phase 4 Tests:** 12 scenarios (Cross-platform, enhanced features)

**Total: All Epic 8 requirements and acceptance criteria covered across all phases**

---

## Implementation Status

### EPIC 8: VERSION COMMAND IMPLEMENTATION - PENDING ⏳
### Current Status: READY FOR IMPLEMENTATION

### Success Criteria Summary

#### Version Command Implementation
- [ ] `agentpm version` displays current version information
- [ ] `agentpm version --format=json` outputs valid JSON
- [ ] `agentpm version --format=xml` outputs valid XML
- [ ] Version information includes all required metadata
- [ ] All version command tests pass

#### Build System Enhancement  
- [ ] `make build` produces working binary with correct version
- [ ] `make test` runs all tests successfully
- [ ] `make clean` removes all build artifacts
- [ ] `make install` installs to system location
- [ ] Cross-platform builds work correctly

#### Integration & Quality
- [ ] Version injected at build time (not runtime)
- [ ] No runtime dependencies on VERSION file
- [ ] Build works in CI/CD environments
- [ ] Development builds have sensible defaults
- [ ] Release builds include all metadata

### Definition of Done
- [ ] All acceptance criteria verified with automated tests
- [ ] Version command executes in < 100ms
- [ ] Build system works across Linux, macOS, Windows
- [ ] All error cases handled gracefully with clear messages
- [ ] Build-time injection works in all environments
- [ ] Cross-platform builds produce working binaries
- [ ] Integration tests verify end-to-end build workflows