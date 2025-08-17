# Makefile for agentpm
# Go-based build system with version injection

# Variables
APP_NAME := agentpm
MODULE := github.com/mindreframer/agentpm
BINARY_NAME := $(APP_NAME)
BUILD_DIR := build
MAIN_FILE := main.go

# Version information
VERSION := $(shell cat VERSION 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GO_VERSION := $(shell go version | awk '{print $$3}')

# Build flags for version injection
VERSION_PKG := $(MODULE)/cmd
LDFLAGS := -ldflags "-X '$(VERSION_PKG).Version=$(VERSION)' -X '$(VERSION_PKG).GitCommit=$(GIT_COMMIT)' -X '$(VERSION_PKG).BuildDate=$(BUILD_DATE)'"

# Go build flags
GO_BUILD_FLAGS := -v
GO_BUILD_FLAGS_RELEASE := -v -ldflags "-s -w -X '$(VERSION_PKG).Version=$(VERSION)' -X '$(VERSION_PKG).GitCommit=$(GIT_COMMIT)' -X '$(VERSION_PKG).BuildDate=$(BUILD_DATE)'"

# Default target
.PHONY: all
all: build

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          Build the application with version injection"
	@echo "  test           Run all tests"
	@echo "  clean          Clean build artifacts"
	@echo "  install        Install to system location"
	@echo "  dev            Development build (fast, no optimizations)"
	@echo "  release        Release build (optimized, stripped)"
	@echo "  lint           Run code linting"
	@echo "  fmt            Format code"
	@echo "  deps           Install dependencies"
	@echo "  version        Show current version"
	@echo "  build-linux    Build for Linux"
	@echo "  build-macos    Build for macOS"
	@echo "  build-windows  Build for Windows"
	@echo "  build-all      Build for all platforms"
	@echo "  help           Show this help message"

# Build target with version injection
.PHONY: build
build: clean
	@echo "Building $(APP_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Development build (fast)
.PHONY: dev
dev:
	@echo "Building $(APP_NAME) development version..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Development build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Release build (optimized)
.PHONY: release
release: clean
	@echo "Building $(APP_NAME) release v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(GO_BUILD_FLAGS_RELEASE) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Release build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Test target
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Update snapshots target
.PHONY: update-snapshots
update-snapshots:
	@echo "Updating all test snapshots..."
	SNAPS_UPDATE=true go test ./cmd ./internal/testing -run ".*XML.*|.*Snapshot.*"
	@echo "Snapshots updated"

# Test with snapshot validation
.PHONY: test-snapshots
test-snapshots:
	@echo "Running tests with snapshot validation..."
	go test -v ./cmd ./internal/testing -run ".*XML.*|.*Snapshot.*"
	@echo "Snapshot tests complete"

# Clean target
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean

# Install target
.PHONY: install
install: build
	@echo "Installing $(APP_NAME) to system..."
	go install $(LDFLAGS) $(GO_BUILD_FLAGS) ./...
	@echo "Installation complete"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Show version
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Git commit: $(GIT_COMMIT)"
	@echo "Build date: $(BUILD_DATE)"
	@echo "Go version: $(GO_VERSION)"

# Cross-platform builds
.PHONY: build-linux
build-linux: clean
	@echo "Building $(APP_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
	@echo "Linux build complete: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

.PHONY: build-macos
build-macos: clean
	@echo "Building $(APP_NAME) for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-macos-amd64 $(MAIN_FILE)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-macos-arm64 $(MAIN_FILE)
	@echo "macOS builds complete: $(BUILD_DIR)/$(BINARY_NAME)-macos-*"

.PHONY: build-windows
build-windows: clean
	@echo "Building $(APP_NAME) for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)
	@echo "Windows build complete: $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe"

.PHONY: build-all
build-all: build-linux build-macos build-windows
	@echo "All platform builds complete"
	@ls -la $(BUILD_DIR)/

# Development utilities
.PHONY: run
run: dev
	./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: bench
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Verification targets
.PHONY: verify
verify: fmt lint test
	@echo "All verification checks passed"

.PHONY: check-version
check-version:
	@if [ ! -f VERSION ]; then echo "Error: VERSION file not found"; exit 1; fi
	@echo "VERSION file exists: $(VERSION)"

# Build with checksums
.PHONY: build-checksums
build-checksums: build-all
	@echo "Generating checksums..."
	@cd $(BUILD_DIR) && sha256sum * > checksums.txt
	@echo "Checksums generated: $(BUILD_DIR)/checksums.txt"