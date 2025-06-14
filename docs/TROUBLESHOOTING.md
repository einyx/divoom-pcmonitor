# Troubleshooting Guide

## Common Build Issues

### Error: "no required module provides package main.go"

This error typically occurs when:

1. **Running go build incorrectly**
   ```bash
   # Wrong - don't include .go extension in module path
   go build main.go
   
   # Correct - use the module path
   go build ./cmd/divoom-monitor
   # or from project root
   make build
   ```

2. **Running from wrong directory**
   ```bash
   # Make sure you're in the project root
   cd /path/to/divoom-pcmonitor-Linux
   make build
   ```

3. **Missing dependencies**
   ```bash
   # Install dependencies first
   make deps
   # or
   go mod download
   go mod tidy
   ```

### Build Commands

The correct way to build this project:

```bash
# Using Make (recommended)
make build          # Build all binaries
make clean build    # Clean and rebuild
make deps          # Install dependencies

# Using Go directly
go build -o bin/divoom-monitor ./cmd/divoom-monitor
go build -o bin/divoom-daemon ./cmd/divoom-daemon
go build -o bin/divoom-test ./cmd/divoom-test
go build -o bin/divoom-auto ./cmd/divoom-auto
```

### Module Structure

The project uses Go modules with the following structure:
- Module name: `divoom-monitor` (defined in go.mod)
- Main packages are in `cmd/` subdirectories
- Shared code is in `internal/` subdirectories

## Systemd Service Issues

### Service Won't Start

1. **Check logs**
   ```bash
   sudo journalctl -u divoom-monitor -n 50
   ```

2. **Test daemon manually**
   ```bash
   sudo -u divoom /usr/bin/divoom-daemon --device=YOUR_DEVICE_IP
   ```

3. **Check permissions**
   ```bash
   ls -la /usr/bin/divoom-*
   ls -la /var/lib/divoom
   ```

### Device Not Found

1. **Test connectivity**
   ```bash
   divoom-test
   ```

2. **Check network**
   ```bash
   # Find Divoom devices on network
   sudo nmap -sn 192.168.1.0/24 | grep -B2 "Divoom"
   ```

3. **Firewall issues**
   ```bash
   # Check if firewall is blocking
   sudo iptables -L -n | grep 80
   ```

### Manual Systemd Setup

If the automatic setup fails, you can set up systemd manually:

```bash
# 1. Create user
sudo useradd --system --home-dir /var/lib/divoom --shell /bin/false divoom

# 2. Create directories
sudo mkdir -p /var/lib/divoom
sudo chown divoom:divoom /var/lib/divoom

# 3. Copy binaries
sudo cp bin/divoom-* /usr/bin/
sudo chmod 755 /usr/bin/divoom-*

# 4. Copy service file
sudo cp packaging/systemd/divoom-monitor.service /etc/systemd/system/

# 5. Reload and enable
sudo systemctl daemon-reload
sudo systemctl enable divoom-monitor
sudo systemctl start divoom-monitor
```

## Go Environment Issues

### Check Go Installation

```bash
# Check Go version (need 1.21+)
go version

# Check Go environment
go env GOPATH
go env GOROOT
go env GO111MODULE
```

### Module Cache Issues

```bash
# Clear module cache if corrupted
go clean -modcache

# Re-download dependencies
go mod download
```

## Device-Specific Issues

### TimeGate Devices

TimeGate devices support multiple LCD displays (0-4). Configure the LCD ID:

```bash
# Edit service
sudo systemctl edit divoom-monitor

# Add configuration
[Service]
ExecStart=
ExecStart=/usr/bin/divoom-daemon --syslog --lcd=1
```

### Pixoo Devices

Pixoo devices don't support LCD selection. Remove `--lcd` flag if present.

## Getting Help

1. **Check device compatibility**
   ```bash
   divoom-test
   ```

2. **Enable debug logging**
   ```bash
   divoom-daemon --debug --device=YOUR_IP
   ```

3. **Report issues**
   - Include output of `divoom-test`
   - Include systemd logs: `sudo journalctl -u divoom-monitor -n 100`
   - Include Go version: `go version`
   - Include OS version: `cat /etc/os-release`