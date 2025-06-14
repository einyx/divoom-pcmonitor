#!/bin/bash
# Build environment verification script

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "divoom-pcmonitor Build Environment Check"
echo "=========================================="

# Check Go installation
echo -n "Checking Go installation... "
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "${GREEN}✓ Found $GO_VERSION${NC}"
    
    # Check Go version (need 1.21+)
    GO_VERSION_NUM=$(echo $GO_VERSION | sed 's/go//' | cut -d. -f1,2)
    REQUIRED_VERSION="1.21"
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION_NUM" | sort -V | head -n1)" = "$REQUIRED_VERSION" ]; then
        echo -e "${GREEN}✓ Go version is compatible${NC}"
    else
        echo -e "${RED}✗ Go version $GO_VERSION_NUM is too old. Need Go $REQUIRED_VERSION or later${NC}"
        exit 1
    fi
else
    echo -e "${RED}✗ Go is not installed${NC}"
    echo "Please install Go 1.21 or later from https://golang.org/dl/"
    exit 1
fi

# Check if in correct directory
echo -n "Checking project directory... "
if [ -f "go.mod" ] && [ -d "cmd" ]; then
    echo -e "${GREEN}✓ In correct directory${NC}"
else
    echo -e "${RED}✗ Not in project root directory${NC}"
    echo "Please run this script from the divoom-pcmonitor-Linux directory"
    exit 1
fi

# Check go.mod
echo -n "Checking go.mod... "
if grep -q "module divoom-monitor" go.mod; then
    echo -e "${GREEN}✓ go.mod is correct${NC}"
else
    echo -e "${RED}✗ go.mod has incorrect module name${NC}"
    exit 1
fi

# Check dependencies
echo -n "Checking dependencies... "
if go list -m all &> /dev/null; then
    echo -e "${GREEN}✓ Dependencies are valid${NC}"
else
    echo -e "${YELLOW}⚠ Dependencies need updating${NC}"
    echo "Running 'go mod tidy'..."
    go mod tidy
fi

# Try a test build
echo -n "Testing build... "
if go build -o /tmp/test-divoom-monitor ./cmd/divoom-monitor &> /dev/null; then
    echo -e "${GREEN}✓ Build successful${NC}"
    rm -f /tmp/test-divoom-monitor
else
    echo -e "${RED}✗ Build failed${NC}"
    echo "Trying verbose build to see errors:"
    go build -v -o /tmp/test-divoom-monitor ./cmd/divoom-monitor
    exit 1
fi

# Check Make
echo -n "Checking Make... "
if command -v make &> /dev/null; then
    echo -e "${GREEN}✓ Make is installed${NC}"
else
    echo -e "${YELLOW}⚠ Make is not installed (optional but recommended)${NC}"
fi

# Check systemd (for Linux)
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo -n "Checking systemd... "
    if command -v systemctl &> /dev/null; then
        echo -e "${GREEN}✓ Systemd is available${NC}"
    else
        echo -e "${YELLOW}⚠ Systemd not found (service installation won't work)${NC}"
    fi
fi

echo ""
echo -e "${GREEN}Build environment check complete!${NC}"
echo ""
echo "You can now build the project with:"
echo "  make build"
echo "or"
echo "  go build -o bin/divoom-monitor ./cmd/divoom-monitor"
echo ""
echo "For systemd service setup, run:"
echo "  sudo ./scripts/setup-systemd.sh"