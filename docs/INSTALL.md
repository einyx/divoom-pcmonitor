# DivoomPCMonitorTool Installation Guide

## Quick Installation

### Ubuntu/Debian (Recommended)
```bash
# Download the DEB package from releases
wget https://github.com/alessio/DivoomPCMonitorTool-Linux/releases/latest/download/divoom-pcmonitor-1.0.0-amd64.deb

# Install the package
sudo dpkg -i divoom-pcmonitor-1.0.0-amd64.deb

# Start the service
sudo systemctl start divoom-monitor
sudo systemctl enable divoom-monitor  # Enable on boot

# Check status
sudo systemctl status divoom-monitor
```

### RHEL/CentOS/Fedora
```bash
# Download the RPM package from releases
wget https://github.com/alessio/DivoomPCMonitorTool-Linux/releases/latest/download/divoom-pcmonitor-1.0.0-1.x86_64.rpm

# Install the package
sudo rpm -i divoom-pcmonitor-1.0.0-1.x86_64.rpm

# Start the service
sudo systemctl start divoom-monitor
sudo systemctl enable divoom-monitor  # Enable on boot
```

### Manual Installation (Any Linux)
```bash
# Download binary archive
wget https://github.com/alessio/DivoomPCMonitorTool-Linux/releases/latest/download/divoom-pcmonitor-1.0.0-linux-amd64.tar.gz

# Extract
tar -xzf divoom-pcmonitor-1.0.0-linux-amd64.tar.gz

# Copy binaries to system
sudo cp divoom-* /usr/local/bin/

# Make executable
sudo chmod +x /usr/local/bin/divoom-*
```

## Applications Included

### 1. `divoom-monitor` - Interactive Monitor
Interactive application with menu-driven interface:
```bash
divoom-monitor
```

### 2. `divoom-daemon` - Background Service
Runs as a systemd service in the background:
```bash
# Manual run
divoom-daemon --help

# Service commands
sudo systemctl start divoom-monitor
sudo systemctl stop divoom-monitor
sudo systemctl status divoom-monitor
```

### 3. `divoom-auto` - Simple Auto Monitor
Standalone automatic monitoring (legacy):
```bash
divoom-auto
```

### 4. `divoom-test` - Device Tester
Test connectivity to your Divoom device:
```bash
divoom-test
```

## Configuration

### Daemon Configuration
The daemon can be configured via command line flags or by editing the systemd service:

```bash
# Edit service configuration
sudo systemctl edit divoom-monitor
```

Add configuration in the override file:
```ini
[Service]
ExecStart=
ExecStart=/usr/bin/divoom-daemon --syslog --interval=5 --device=192.168.1.100
```

### Available Options
- `--device IP`: Specify device IP (auto-detect if not set)
- `--interval N`: Update interval in seconds (default: 3)
- `--lcd N`: LCD ID for TimeGate devices (0-4)
- `--syslog`: Use syslog for logging
- `--logfile PATH`: Log to specific file

## Troubleshooting

### Service Not Starting
```bash
# Check service logs
sudo journalctl -u divoom-monitor -f

# Check if device is reachable
divoom-test

# Manual test run
sudo -u divoom divoom-daemon --device=YOUR_DEVICE_IP
```

### Device Not Found
1. Ensure your Divoom device is on the same network
2. Check firewall settings
3. Test with `divoom-test`
4. Manually specify device IP: `--device=192.168.1.XXX`

### Permission Issues
The service runs as user `divoom` for security. If you have permission issues:
```bash
# Check user exists
getent passwd divoom

# Check service permissions
sudo systemctl status divoom-monitor
```

## Uninstallation

### Ubuntu/Debian
```bash
sudo apt remove divoom-pcmonitor
# or for complete removal:
sudo apt purge divoom-pcmonitor
```

### RHEL/CentOS/Fedora
```bash
sudo rpm -e divoom-pcmonitor
```

### Manual
```bash
sudo systemctl stop divoom-monitor
sudo systemctl disable divoom-monitor
sudo rm /usr/local/bin/divoom-*
sudo rm /etc/systemd/system/divoom-monitor.service
sudo systemctl daemon-reload
sudo userdel divoom
sudo rm -rf /var/lib/divoom
```

## Building from Source

### Prerequisites
- Go 1.19 or later
- Build tools: `make`, `dpkg-dev` (for DEB), `rpm-build` (for RPM)

### Build Commands
```bash
# Simple build
make build

# Cross-platform build
make build-all

# Build packages
make build-packages

# Clean
make clean
```

## Architecture Support

- **Linux**: amd64, 386, arm64, arm
- **Windows**: amd64, 386, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **FreeBSD**: amd64, 386

## Windows Installation

Download `DivoomPCMonitorTool-Setup.exe` from releases and run the installer.

The Windows version includes:
- Interactive GUI application
- Windows service support
- Device connectivity tester

## Hardware Requirements

- **CPU**: Any modern x86_64 or ARM processor
- **Memory**: 64MB RAM minimum
- **Network**: Connection to same network as Divoom device
- **OS**: Linux kernel 2.6+, Windows 10+, macOS 10.12+