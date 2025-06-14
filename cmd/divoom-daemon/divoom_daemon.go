package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

var version = "dev" // Set by build flags

type DaemonDeviceList struct {
	TotalData  int            `json:"TotalData"`
	DeviceList []DaemonDevice `json:"DeviceList"`
}

type DaemonDevice struct {
	DeviceName      string `json:"DeviceName"`
	DeviceId        int    `json:"DeviceId"`
	DevicePrivateIP string `json:"DevicePrivateIP"`
	DeviceMac       string `json:"DeviceMac"`
	Hardware        int    `json:"Hardware"`
}

type DaemonHardwareData struct {
	CpuUsage    int
	GpuUsage    int
	CpuTemp     int
	GpuTemp     int
	MemoryUsage int
	DiskTemp    int
}

type DaemonPCMonitorPayload struct {
	Command    string                        `json:"Command"`
	ScreenList []DaemonPCMonitorScreenItem `json:"ScreenList"`
}

type DaemonPCMonitorScreenItem struct {
	LcdId    int      `json:"LcdId"`
	DispData []string `json:"DispData"`
}

var (
	daemonHttpClient = &http.Client{Timeout: 10 * time.Second}
	logger           *log.Logger
)

func main() {
	var showVersion = flag.Bool("version", false, "Show version information")
	var showHelp = flag.Bool("help", false, "Show help information")
	var deviceIP = flag.String("device", "", "Device IP address (auto-detect if not specified)")
	var lcdId = flag.Int("lcd", 0, "LCD ID for TimeGate devices (0-4)")
	var interval = flag.Int("interval", 3, "Update interval in seconds")
	var useSyslog = flag.Bool("syslog", false, "Use syslog for logging")
	var logFile = flag.String("logfile", "", "Log file path (default: stderr)")
	flag.Parse()

	if *showVersion {
		fmt.Printf("divoom-pcmonitor Daemon version %s\n", version)
		return
	}

	if *showHelp {
		fmt.Println("divoom-pcmonitor Daemon - Background PC monitoring for Divoom devices")
		fmt.Printf("Version: %s\n\n", version)
		fmt.Println("Usage:")
		fmt.Println("  divoom-daemon [flags]")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		fmt.Println("\nSystemd service:")
		fmt.Println("  sudo systemctl enable divoom-monitor")
		fmt.Println("  sudo systemctl start divoom-monitor")
		return
	}

	// Setup logging
	setupLogging(*useSyslog, *logFile)
	
	logger.Printf("Starting divoom-pcmonitor Daemon v%s", version)
	
	// Find device
	var device *DaemonDevice
	if *deviceIP != "" {
		device = &DaemonDevice{DevicePrivateIP: *deviceIP}
		logger.Printf("Using specified device IP: %s", *deviceIP)
	} else {
		logger.Println("Auto-detecting Divoom device...")
		devices, err := findDaemonDevices()
		if err != nil {
			logger.Fatalf("Error finding devices: %v", err)
		}
		if len(devices) == 0 {
			logger.Fatal("No Divoom devices found on network")
		}
		device = &devices[0]
		logger.Printf("Found device: %s (%s)", device.DeviceName, device.DevicePrivateIP)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Start monitoring loop
	ticker := time.NewTicker(time.Duration(*interval) * time.Second)
	defer ticker.Stop()

	logger.Printf("Starting monitoring loop (interval: %ds, LCD: %d)", *interval, *lcdId)

	for {
		select {
		case <-ticker.C:
			data := getDaemonHardwareData()
			if err := sendDaemonDataToDevice(*device, data, *lcdId); err != nil {
				logger.Printf("Error sending data: %v", err)
			} else {
				logger.Printf("Sent: CPU:%d%% %d°C GPU:%d%% %d°C MEM:%d%% DSK:%d°C",
					data.CpuUsage, data.CpuTemp, data.GpuUsage, data.GpuTemp,
					data.MemoryUsage, data.DiskTemp)
			}

		case sig := <-sigChan:
			logger.Printf("Received signal: %v", sig)
			if sig == syscall.SIGHUP {
				logger.Println("Reloading configuration...")
				// Reload logic here if needed
				continue
			}
			logger.Println("Shutting down gracefully...")
			return
		}
	}
}

func setupLogging(useSyslog bool, logFile string) {
	if useSyslog {
		syslogger, err := syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, "divoom-daemon")
		if err != nil {
			log.Fatalf("Failed to connect to syslog: %v", err)
		}
		logger = log.New(syslogger, "", 0)
	} else if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		logger = log.New(file, "", log.LstdFlags)
	} else {
		logger = log.New(os.Stderr, "", log.LstdFlags)
	}
}

func findDaemonDevices() ([]DaemonDevice, error) {
	resp, err := daemonHttpClient.Get("http://app.divoom-gz.com/Device/ReturnSameLANDevice")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var deviceList DaemonDeviceList
	if err := json.Unmarshal(body, &deviceList); err != nil {
		return nil, err
	}

	return deviceList.DeviceList, nil
}

func getDaemonHardwareData() DaemonHardwareData {
	data := DaemonHardwareData{}

	// CPU Usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		data.CpuUsage = int(cpuPercent[0])
	}

	// Temperature sensors
	temps, err := host.SensorsTemperatures()
	if err == nil {
		// CPU Temperature
		for _, temp := range temps {
			if strings.Contains(strings.ToLower(temp.SensorKey), "cpu") ||
				strings.Contains(strings.ToLower(temp.SensorKey), "package") ||
				strings.Contains(strings.ToLower(temp.SensorKey), "core") {
				data.CpuTemp = int(temp.Temperature)
				break
			}
		}

		// Disk Temperature
		for _, temp := range temps {
			if strings.Contains(strings.ToLower(temp.SensorKey), "nvme") ||
				strings.Contains(strings.ToLower(temp.SensorKey), "sda") ||
				strings.Contains(strings.ToLower(temp.SensorKey), "disk") {
				data.DiskTemp = int(temp.Temperature)
				break
			}
		}
	}

	// Memory Usage
	vmStat, err := mem.VirtualMemory()
	if err == nil {
		data.MemoryUsage = int(vmStat.UsedPercent)
	}

	// GPU data (placeholder - would need nvidia-smi parsing)
	data.GpuUsage = 0
	data.GpuTemp = 0

	return data
}

func sendDaemonDataToDevice(device DaemonDevice, data DaemonHardwareData, lcdId int) error {
	// Format data according to Windows implementation
	// DispData array: [CpuUse, GpuUse, CpuTemp, GpuTemp, MemUse, DiskTemp]
	cpuUse := fmt.Sprintf("%d%%", data.CpuUsage)
	gpuUse := fmt.Sprintf("%d%%", data.GpuUsage)
	cpuTemp := fmt.Sprintf("%d°C", data.CpuTemp)
	gpuTemp := fmt.Sprintf("%d°C", data.GpuTemp)
	memUse := fmt.Sprintf("%d%%", data.MemoryUsage)
	diskTemp := fmt.Sprintf("%d°C", data.DiskTemp)

	payload := DaemonPCMonitorPayload{
		Command: "Device/UpdatePCParaInfo",
		ScreenList: []DaemonPCMonitorScreenItem{
			{
				LcdId:    lcdId,
				DispData: []string{cpuUse, gpuUse, cpuTemp, gpuTemp, memUse, diskTemp},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s:80/post", device.DevicePrivateIP)
	resp, err := daemonHttpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("device returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}