#!/bin/bash

# Cross-platform build script for DivoomPCMonitorTool
# Builds for Windows, Linux, and macOS

set -e

VERSION="1.0.0"
BUILD_DIR="bin"
LDFLAGS="-s -w -X main.version=${VERSION}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building DivoomPCMonitorTool v${VERSION} for multiple platforms${NC}"
echo "=================================================="

# Create build directory
mkdir -p ${BUILD_DIR}

# Build function
build_target() {
    local os=$1
    local arch=$2
    local extension=$3
    local target_name="${os}_${arch}"
    
    echo -e "${YELLOW}Building for ${os}/${arch}...${NC}"
    
    # Set environment variables
    export GOOS=${os}
    export GOARCH=${arch}
    export CGO_ENABLED=0
    
    # Build main interactive version
    go build -ldflags="${LDFLAGS}" -o "${BUILD_DIR}/divoom-monitor-${target_name}${extension}" ./cmd/divoom-monitor
    
    # Build auto version
    go build -ldflags="${LDFLAGS}" -o "${BUILD_DIR}/divoom-auto-${target_name}${extension}" ./cmd/divoom-auto
    
    # Build daemon version
    go build -ldflags="${LDFLAGS}" -o "${BUILD_DIR}/divoom-daemon-${target_name}${extension}" ./cmd/divoom-daemon
    
    # Build test version
    go build -ldflags="${LDFLAGS}" -o "${BUILD_DIR}/divoom-test-${target_name}${extension}" ./cmd/divoom-test
    
    echo -e "${GREEN}✓ Built for ${os}/${arch}${NC}"
}

# Build for different platforms
echo "Building binaries..."

# Linux builds
build_target "linux" "amd64" ""
build_target "linux" "386" ""
build_target "linux" "arm64" ""
build_target "linux" "arm" ""

# Windows builds
build_target "windows" "amd64" ".exe"
build_target "windows" "386" ".exe"
build_target "windows" "arm64" ".exe"

# macOS builds
build_target "darwin" "amd64" ""
build_target "darwin" "arm64" ""

# FreeBSD builds
build_target "freebsd" "amd64" ""
build_target "freebsd" "386" ""

echo ""
echo -e "${GREEN}Build completed successfully!${NC}"
echo "Built files:"
ls -la ${BUILD_DIR}/

# Create checksums
echo ""
echo "Creating checksums..."
cd ${BUILD_DIR}
sha256sum * > checksums.txt
cd ..

echo -e "${GREEN}Checksums created in ${BUILD_DIR}/checksums.txt${NC}"
echo ""
echo "Build artifacts ready for distribution!"