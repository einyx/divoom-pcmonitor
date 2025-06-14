# Makefile for divoom-pcmonitor

# Variables
VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS = -s -w -X main.version=$(VERSION)
BUILD_DIR = bin
DIST_DIR = dist

# Default target
.PHONY: all
all: build

# Build for current platform
.PHONY: build
build:
	@echo "Building for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/divoom-monitor ./cmd/divoom-monitor
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/divoom-daemon ./cmd/divoom-daemon
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/divoom-auto ./cmd/divoom-auto
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/divoom-test ./cmd/divoom-test
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/hardware-test ./cmd/hardware-test
	@echo "Build complete!"

# Build for all platforms
.PHONY: build-all
build-all:
	@echo "Building for all platforms..."
	./build-cross-platform.sh

# Build packages (DEB, RPM, Windows installer)
.PHONY: build-packages
build-packages:
	@echo "Building all packages..."
	./scripts/build-packages.sh $(VERSION)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) $(DIST_DIR) builds
	@echo "Clean complete!"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run the interactive monitor
.PHONY: run
run: build
	$(BUILD_DIR)/divoom-monitor

# Run the auto monitor
.PHONY: run-auto
run-auto: build
	$(BUILD_DIR)/divoom-auto

# Run the daemon
.PHONY: run-daemon
run-daemon: build
	$(BUILD_DIR)/divoom-daemon

# Run device test
.PHONY: run-test
run-test: build
	$(BUILD_DIR)/divoom-test

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Static analysis
.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Development setup
.PHONY: dev
dev: deps fmt vet test build
	@echo "Development setup complete!"

# Build for specific OS/ARCH
.PHONY: build-linux
build-linux:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/divoom-monitor-linux ./cmd/divoom-monitor
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/divoom-daemon-linux ./cmd/divoom-daemon
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/divoom-auto-linux ./cmd/divoom-auto

.PHONY: build-windows
build-windows:
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/divoom-monitor-windows.exe ./cmd/divoom-monitor
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/divoom-daemon-windows.exe ./cmd/divoom-daemon
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/divoom-auto-windows.exe ./cmd/divoom-auto

.PHONY: build-macos
build-macos:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/divoom-monitor-macos-amd64 ./cmd/divoom-monitor
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/divoom-monitor-macos-arm64 ./cmd/divoom-monitor
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/divoom-auto-macos-amd64 ./cmd/divoom-auto
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/divoom-auto-macos-arm64 ./cmd/divoom-auto

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build       - Build for current platform"
	@echo "  build-all   - Build for all platforms"
	@echo "  build-linux - Build for Linux"
	@echo "  build-windows - Build for Windows"
	@echo "  build-macos - Build for macOS"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  run         - Run interactive monitor"
	@echo "  run-auto    - Run auto monitor"
	@echo "  run-test    - Run device test"
	@echo "  deps        - Install dependencies"
	@echo "  fmt         - Format code"
	@echo "  vet         - Run static analysis"
	@echo "  dev         - Development setup"
	@echo "  help        - Show this help"