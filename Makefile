# GTFS-Realtime to SIRI Converter - Makefile
# Adapted for Go project with cross-platform builds and releases

# Binary name
BINARY_NAME=gtfsrt-to-siri
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build directory
BUILD_DIR=./bin
CMD_DIR=./cmd/gtfsrt-to-siri

# Linker flags for version information
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(BUILD_DATE)"

.PHONY: all build clean test coverage lint help install deps release

## help: Display this help message
help:
	@echo "Available targets:"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /' | column -t -s ':'
	@echo ""

## all: Run tests, lint, and build
all: deps test lint build

## deps: Download Go module dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## build: Build the binary for current platform
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "✓ Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

## build-linux: Build for Linux amd64
build-linux:
	@echo "Building for Linux amd64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	@echo "✓ Built: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

## build-darwin: Build for macOS (both Intel and Apple Silicon)
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	@echo "✓ Built: $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64"
	@echo "✓ Built: $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64"

## build-windows: Build for Windows amd64
build-windows:
	@echo "Building for Windows amd64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)
	@echo "✓ Built: $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe"

## build-all: Build for all platforms (Linux, macOS, Windows)
build-all: build-linux build-darwin build-windows
	@echo "✓ All platform binaries built"

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./tests/...
	@echo "✓ All tests passed"

## test-unit: Run only unit tests
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v ./tests/unit/...

## test-integration: Run only integration tests
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v ./tests/integration/...

## test-race: Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	$(GOTEST) -race -v ./tests/...

## coverage: Generate test coverage report
coverage:
	@echo "Generating coverage report..."
	$(GOTEST) -coverprofile=coverage.out -covermode=atomic ./tests/...
	$(GOCMD) tool cover -func=coverage.out
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

## lint: Run golangci-lint
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run --timeout=5m
	@echo "✓ Linting passed"

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@echo "✓ Code formatted"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...
	@echo "✓ Vet passed"

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -f $(BINARY_NAME)
	@echo "✓ Clean complete"

## install: Install the binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(LDFLAGS) $(CMD_DIR)
	@echo "✓ Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

## run: Build and run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	$(BUILD_DIR)/$(BINARY_NAME)

## version: Display version information
version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Built:   $(BUILD_DATE)"

## release: Create release archives for all platforms
release: clean build-all
	@echo "Creating release archives..."
	@mkdir -p $(BUILD_DIR)/releases
	# Linux
	tar -czf $(BUILD_DIR)/releases/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-linux-amd64
	# macOS Intel
	tar -czf $(BUILD_DIR)/releases/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-darwin-amd64
	# macOS Apple Silicon
	tar -czf $(BUILD_DIR)/releases/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-darwin-arm64
	# Windows
	zip -j $(BUILD_DIR)/releases/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe
	@echo "✓ Release archives created in $(BUILD_DIR)/releases/"
	@ls -lh $(BUILD_DIR)/releases/

## checksums: Generate SHA256 checksums for release files
checksums:
	@echo "Generating checksums..."
	@cd $(BUILD_DIR)/releases && sha256sum * > SHA256SUMS
	@echo "✓ Checksums written to $(BUILD_DIR)/releases/SHA256SUMS"

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest
	@echo "✓ Docker image built: $(BINARY_NAME):$(VERSION)"

## ci: Run CI pipeline locally (test + lint + build)
ci: deps test lint build
	@echo "✓ CI pipeline completed successfully"

## benchmark: Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

## todo: Show TODO/FIXME comments in code
todo:
	@echo "TODOs and FIXMEs:"
	@grep -rn "TODO\|FIXME" --include="*.go" . || echo "None found"

# Default target
.DEFAULT_GOAL := help

