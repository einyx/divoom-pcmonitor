# Divoom PC Monitor Tool for Linux (Go Version)

A lightweight Go implementation of the Divoom PC Monitor Tool for Linux systems.

## Features

- Auto-discovers Divoom devices on your local network
- Monitors CPU usage and temperature
- Monitors memory usage
- Monitors disk temperature
- GPU monitoring support (requires nvidia-smi for NVIDIA GPUs)
- Sends real-time data to Divoom devices using Clock ID 625
- Support for TimeGate multi-LCD devices

## Prerequisites

- Go 1.21 or later
- Linux system with hardware sensors support
- Divoom device on the same network

## Installation

### Quick Installation (Recommended)

Download a pre-built package from the [releases page](https://github.com/alessio/divoom-pcmonitor-Linux/releases) or see [INSTALL.md](docs/INSTALL.md) for detailed instructions.

### Build from Source

1. Clone the repository and navigate to the directory:
```bash
cd divoom-pcmonitor-Linux
```

2. Verify build environment:
```bash
./scripts/verify-build.sh
```

3. Build the application:
```bash
make build
```

Or manually:
```bash
go build -o bin/divoom-monitor ./cmd/divoom-monitor
go build -o bin/divoom-daemon ./cmd/divoom-daemon
go build -o bin/divoom-test ./cmd/divoom-test
```

### Systemd Service Setup

For automatic startup and background monitoring:
```bash
sudo ./scripts/setup-systemd.sh
```

## Usage

Run the application (may need sudo for hardware monitoring):

```bash
sudo bin/divoom-monitor
```

Or use make:
```bash
sudo make run
```

### Menu Options

1. **Scan for Divoom devices** - Discovers all Divoom devices on your network
2. **Select device** - Choose which device to send data to
3. **Start monitoring** - Begin sending hardware data to the selected device
4. **Exit** - Quit the application

### Hardware Monitoring

The tool monitors:
- CPU: Usage percentage and temperature
- GPU: Usage and temperature (limited support)
- Memory: Usage percentage
- Disk: Temperature

Data is sent to the Divoom device every 2 seconds.

## Building

### Standard build:
```bash
make build
```

### Cross-platform build:
```bash
make build-all
```

### Platform-specific builds:
```bash
make build-linux    # Linux AMD64
make build-windows  # Windows AMD64
make build-macos    # macOS (both Intel and Apple Silicon)
```

### Manual builds:
```bash
# Current platform
go build -o bin/divoom-monitor ./cmd/divoom-monitor

# Cross-compilation examples
GOOS=linux GOARCH=arm64 go build -o bin/divoom-monitor-arm64 ./cmd/divoom-monitor
GOOS=windows GOARCH=amd64 go build -o bin/divoom-monitor.exe ./cmd/divoom-monitor
```

## Differences from C# Version

- Uses gopsutil library instead of LibreHardwareMonitor
- GPU monitoring is simplified (full support would require parsing nvidia-smi or similar)
- More lightweight and faster startup
- Single binary with no runtime dependencies

## Troubleshooting

See [TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) for common issues and solutions.

### Quick Fixes

1. **Build Issues**: Run `./scripts/verify-build.sh` to check your environment
2. **Device Not Found**: Use `divoom-test` to verify connectivity
3. **Service Issues**: Check logs with `sudo journalctl -u divoom-monitor -f`

### No Temperature Data
- Install lm-sensors: `sudo apt-get install lm-sensors`
- Run sensors detection: `sudo sensors-detect`
- Verify sensors work: `sensors`

### GPU Monitoring
- For NVIDIA GPUs, ensure nvidia-smi is installed and accessible
- AMD GPU support would require additional implementation

## License

Same as the original C# version