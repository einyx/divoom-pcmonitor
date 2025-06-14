#!/bin/bash

echo "Building Divoom Monitor (Go version) for Linux..."

# Download dependencies
echo "Downloading dependencies..."
go mod download

# Build standard version
echo "Building standard version..."
go build -o divoom-monitor

# Build static version
echo "Building static version..."
CGO_ENABLED=0 go build -ldflags="-s -w" -o divoom-monitor-static

# Make executables
chmod +x divoom-monitor
chmod +x divoom-monitor-static

echo "Build complete!"
echo ""
echo "Executables created:"
echo "  - divoom-monitor (standard build)"
echo "  - divoom-monitor-static (static build, no dependencies)"
echo ""
echo "To run: sudo ./divoom-monitor"