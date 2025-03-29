# Makefile for go-discovery

# Variables
BINARY_NAME=discovery
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOBUILD=go build ${LDFLAGS}
GOCLEAN=go clean
GOTEST=go test
GOGET=go get
GOMOD=go mod
GOFMT=go fmt
GOLINT=golangci-lint

# Build directory
BUILD_DIR=./bin

# Main target
.PHONY: all
all: clean build

# Build the application
.PHONY: build
build:
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/discovery

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run code formatter
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Run linter
.PHONY: lint
lint:
	@echo "Linting code..."
	@if command -v $(GOLINT) > /dev/null; then \
		$(GOLINT) run; \
	else \
		echo "golangci-lint not found, please install it"; \
	fi

# Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOGET) -v ./...

# Install the binary
.PHONY: install
install:
	@echo "Installing binary..."
	$(GOBUILD) -o $(GOPATH)/bin/$(BINARY_NAME) ./cmd/discovery

# Cross compilation
.PHONY: build-all
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux (amd64)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/discovery
	
	# Linux (arm64)
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/discovery
	
	# macOS (amd64)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/discovery
	
	# macOS (arm64)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/discovery
	
	# Windows (amd64)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/discovery

# Help command
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make              : Build the application"
	@echo "  make build        : Build the application"
	@echo "  make clean        : Clean build artifacts"
	@echo "  make test         : Run tests"
	@echo "  make fmt          : Format code"
	@echo "  make lint         : Lint code"
	@echo "  make tidy         : Tidy dependencies"
	@echo "  make deps         : Download dependencies"
	@echo "  make install      : Install binary"
	@echo "  make build-all    : Build for multiple platforms"
