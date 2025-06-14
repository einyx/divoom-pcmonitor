divoom-pcmonitor for Windows
================================

This package includes three utilities for monitoring your PC with Divoom devices:

1. divoom-monitor-windows.exe - Interactive monitoring application
   Run this to manually select your device and start monitoring

2. divoom-daemon-windows.exe - Background service daemon
   This runs as a Windows service in the background

3. divoom-test-windows.exe - Device connectivity tester
   Use this to test if your Divoom device is reachable

Quick Start:
------------
1. Make sure your Divoom device is connected to the same network as your PC
2. Run "divoom-test-windows.exe" to verify connectivity
3. Run "divoom-monitor-windows.exe" for interactive monitoring
   OR
   Install as a service: "divoom-daemon-windows.exe install"
   Start the service: "divoom-daemon-windows.exe start"

Service Commands:
-----------------
Install service:   divoom-daemon-windows.exe install
Start service:     divoom-daemon-windows.exe start
Stop service:      divoom-daemon-windows.exe stop
Remove service:    divoom-daemon-windows.exe remove

For more information, visit:
https://github.com/alessio/divoom-pcmonitor-Linux

System Requirements:
--------------------
- Windows 10 or later
- Network connection to Divoom device
- Administrator privileges for service installation