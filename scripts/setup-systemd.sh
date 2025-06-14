#!/bin/bash
# Systemd service setup script for divoom-pcmonitor

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Error: This script must be run as root${NC}"
    echo "Please run: sudo $0"
    exit 1
fi

echo -e "${GREEN}divoom-pcmonitor Systemd Service Setup${NC}"
echo "=========================================="

# Check if binaries exist
if [ ! -f "bin/divoom-daemon" ]; then
    echo -e "${YELLOW}Building binaries first...${NC}"
    make build || {
        echo -e "${RED}Error: Build failed. Please check your Go installation and run 'make deps' first${NC}"
        exit 1
    }
fi

# Create divoom user if it doesn't exist
if ! id -u divoom >/dev/null 2>&1; then
    echo "Creating divoom system user..."
    useradd --system --home-dir /var/lib/divoom --shell /bin/false divoom
    echo -e "${GREEN}✓ User 'divoom' created${NC}"
else
    echo -e "${GREEN}✓ User 'divoom' already exists${NC}"
fi

# Create home directory
echo "Creating service home directory..."
mkdir -p /var/lib/divoom
chown divoom:divoom /var/lib/divoom
chmod 755 /var/lib/divoom
echo -e "${GREEN}✓ Directory /var/lib/divoom created${NC}"

# Copy binaries
echo "Installing binaries to /usr/bin..."
cp -f bin/divoom-daemon /usr/bin/
cp -f bin/divoom-monitor /usr/bin/
cp -f bin/divoom-test /usr/bin/
cp -f bin/divoom-auto /usr/bin/
chmod 755 /usr/bin/divoom-*
echo -e "${GREEN}✓ Binaries installed${NC}"

# Copy systemd service file
echo "Installing systemd service..."
cp -f packaging/systemd/divoom-monitor.service /etc/systemd/system/
chmod 644 /etc/systemd/system/divoom-monitor.service
echo -e "${GREEN}✓ Service file installed${NC}"

# Copy sysusers.d file if the directory exists
if [ -d "/usr/lib/sysusers.d" ]; then
    cp -f packaging/systemd/divoom-user.conf /usr/lib/sysusers.d/divoom.conf
    echo -e "${GREEN}✓ Sysusers configuration installed${NC}"
fi

# Reload systemd
echo "Reloading systemd daemon..."
systemctl daemon-reload
echo -e "${GREEN}✓ Systemd daemon reloaded${NC}"

# Enable service (but don't start it)
echo "Enabling service..."
systemctl enable divoom-monitor.service || true
echo -e "${GREEN}✓ Service enabled${NC}"

echo ""
echo -e "${GREEN}Installation completed successfully!${NC}"
echo ""
echo "Next steps:"
echo "1. Test device connectivity:"
echo "   divoom-test"
echo ""
echo "2. Start the service:"
echo "   sudo systemctl start divoom-monitor"
echo ""
echo "3. Check service status:"
echo "   sudo systemctl status divoom-monitor"
echo ""
echo "4. View service logs:"
echo "   sudo journalctl -u divoom-monitor -f"
echo ""
echo "5. Configure service (optional):"
echo "   sudo systemctl edit divoom-monitor"
echo "   Add custom options like:"
echo "   [Service]"
echo "   ExecStart="
echo "   ExecStart=/usr/bin/divoom-daemon --syslog --interval=5 --device=192.168.1.100"
echo ""
echo "6. Run interactive monitor:"
echo "   divoom-monitor"