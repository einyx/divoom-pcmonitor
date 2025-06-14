package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

var version = "dev" // Set by build flags

type DivoomDeviceList struct {
	TotalData  int            `json:"TotalData"`
	DeviceList []DivoomDevice `json:"DeviceList"`
}

type DivoomDevice struct {
	DeviceName      string `json:"DeviceName"`
	DeviceId        int    `json:"DeviceId"`
	DevicePrivateIP string `json:"DevicePrivateIP"`
	DeviceMac       string `json:"DeviceMac"`
	Hardware        int    `json:"Hardware"`
}

type HardwareData struct {
	CpuUsage    int
	GpuUsage    int
	CpuTemp     int
	GpuTemp     int
	MemoryUsage int
	DiskTemp    int
}

// PC monitoring payload for Divoom devices (Windows-style)
type PCMonitorPayload struct {
	Command    string               `json:"Command"`
	ScreenList []PCMonitorScreenItem `json:"ScreenList"`
}

type PCMonitorScreenItem struct {
	LcdId    int      `json:"LcdId"`
	DispData []string `json:"DispData"`
}

var (
	httpClient     = &http.Client{Timeout: 10 * time.Second}
	selectedDevice *DivoomDevice
	selectedLcd    = 1
	running        = true
)

func main() {
	var showVersion = flag.Bool("version", false, "Show version information")
	var showHelp = flag.Bool("help", false, "Show help information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("divoom-pcmonitor version %s\n", version)
		return
	}

	if *showHelp {
		fmt.Println("divoom-pcmonitor - PC monitoring for Divoom devices")
		fmt.Printf("Version: %s\n\n", version)
		fmt.Println("Usage:")
		fmt.Println("  divoom-monitor [flags]")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		fmt.Println("\nFeatures:")
		fmt.Println("  1. Scan for Divoom devices on your network")
		fmt.Println("  2. Select and configure your device")
		fmt.Println("  3. Monitor CPU, GPU, Memory, and Disk temperatures")
		fmt.Println("  4. Real-time display on your Divoom device")
		return
	}

	fmt.Printf("Divoom PC Monitor Tool for Linux v%s\n", version)
	fmt.Println("==========================================\n")

	reader := bufio.NewReader(os.Stdin)

	for running {
		clearScreen()
		fmt.Println("Divoom PC Monitor Tool for Linux (Go Version)")
		fmt.Println("============================================\n")

		fmt.Println("1. Scan for Divoom devices")
		fmt.Println("2. Select device")
		fmt.Println("3. Start monitoring")
		fmt.Println("4. Exit")
		fmt.Print("\nSelect option: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			scanDevices()
		case "2":
			selectDevice(reader)
		case "3":
			startMonitoring()
		case "4":
			running = false
		}
	}
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func scanDevices() {
	fmt.Println("\nScanning for devices...")

	resp, err := httpClient.Get("http://app.divoom-gz.com/Device/ReturnSameLANDevice")
	if err != nil {
		fmt.Printf("Error scanning devices: %v\n", err)
		waitForKey()
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		waitForKey()
		return
	}

	var deviceList DivoomDeviceList
	if err := json.Unmarshal(body, &deviceList); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		waitForKey()
		return
	}

	if len(deviceList.DeviceList) > 0 {
		fmt.Printf("\nFound %d device(s):\n", len(deviceList.DeviceList))
		for i, device := range deviceList.DeviceList {
			fmt.Printf("%d. %s (%s)\n", i+1, device.DeviceName, device.DevicePrivateIP)
		}
	} else {
		fmt.Println("No devices found.")
	}

	waitForKey()
}

func selectDevice(reader *bufio.Reader) {
	resp, err := httpClient.Get("http://app.divoom-gz.com/Device/ReturnSameLANDevice")
	if err != nil {
		fmt.Printf("Error getting devices: %v\n", err)
		waitForKey()
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		waitForKey()
		return
	}

	var deviceList DivoomDeviceList
	if err := json.Unmarshal(body, &deviceList); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		waitForKey()
		return
	}

	if len(deviceList.DeviceList) == 0 {
		fmt.Println("\nNo devices available. Please scan first.")
		waitForKey()
		return
	}

	fmt.Println("\nAvailable devices:")
	for i, device := range deviceList.DeviceList {
		fmt.Printf("%d. %s (%s)\n", i+1, device.DeviceName, device.DevicePrivateIP)
	}

	fmt.Print("\nSelect device number: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	selection, err := strconv.Atoi(input)
	if err != nil || selection < 1 || selection > len(deviceList.DeviceList) {
		fmt.Println("Invalid selection.")
		waitForKey()
		return
	}

	selectedDevice = &deviceList.DeviceList[selection-1]
	fmt.Printf("\nSelected: %s\n", selectedDevice.DeviceName)

	// Check if it's a TimeGate device
	if selectedDevice.Hardware == 400 {
		fmt.Print("\nTimeGate device detected. Select LCD (1-5): ")
		lcdInput, _ := reader.ReadString('\n')
		lcdInput = strings.TrimSpace(lcdInput)

		lcd, err := strconv.Atoi(lcdInput)
		if err == nil && lcd >= 1 && lcd <= 5 {
			selectedLcd = lcd
		}
	}

	waitForKey()
}

func startMonitoring() {
	if selectedDevice == nil {
		fmt.Println("\nNo device selected. Please select a device first.")
		waitForKey()
		return
	}

	clearScreen()
	fmt.Printf("Monitoring started for %s\n", selectedDevice.DeviceName)
	fmt.Println("Press 'Q' to stop monitoring\n")

	// Create a channel to signal when to stop
	stop := make(chan bool)

	// Start monitoring in a goroutine
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				data := getHardwareData()
				if err := sendDataToDevice(data); err != nil {
					fmt.Printf("\nError sending data: %v\n", err)
				}

				// Update display
				fmt.Printf("\033[3;0H") // Move cursor to line 3, column 0
				fmt.Printf("CPU: %d%% @ %d°C     \n", data.CpuUsage, data.CpuTemp)
				fmt.Printf("GPU: %d%% @ %d°C     \n", data.GpuUsage, data.GpuTemp)
				fmt.Printf("Memory: %d%%              \n", data.MemoryUsage)
				fmt.Printf("Disk Temp: %d°C            \n", data.DiskTemp)

			case <-stop:
				return
			}
		}
	}()

	// Wait for 'Q' key
	reader := bufio.NewReader(os.Stdin)
	for {
		char, _ := reader.ReadByte()
		if char == 'q' || char == 'Q' {
			stop <- true
			break
		}
	}
}

func getHardwareData() HardwareData {
	data := HardwareData{}

	// CPU Usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		data.CpuUsage = int(cpuPercent[0])
	}

	// CPU Temperature
	temps, err := host.SensorsTemperatures()
	if err == nil {
		for _, temp := range temps {
			if strings.Contains(strings.ToLower(temp.SensorKey), "cpu") || 
			   strings.Contains(strings.ToLower(temp.SensorKey), "package") {
				data.CpuTemp = int(temp.Temperature)
				break
			}
		}
	}

	// Memory Usage
	vmStat, err := mem.VirtualMemory()
	if err == nil {
		data.MemoryUsage = int(vmStat.UsedPercent)
	}

	// GPU data (simplified - gopsutil doesn't have direct GPU support)
	// You might need to parse nvidia-smi or similar for accurate GPU data
	data.GpuUsage = 0
	data.GpuTemp = 0
	
	// Try to get GPU info from nvidia-smi if available
	if gpuData := getNvidiaGPUData(); gpuData != nil {
		data.GpuUsage = gpuData.Usage
		data.GpuTemp = gpuData.Temp
	}

	// Disk Temperature (from first disk)
	for _, temp := range temps {
		if strings.Contains(strings.ToLower(temp.SensorKey), "nvme") || 
		   strings.Contains(strings.ToLower(temp.SensorKey), "sda") {
			data.DiskTemp = int(temp.Temperature)
			break
		}
	}

	return data
}

func getNvidiaGPUData() *struct{ Usage, Temp int } {
	// Check if nvidia-smi is available first
	if _, err := exec.LookPath("nvidia-smi"); err != nil {
		return nil
	}
	
	// Try to execute nvidia-smi to get GPU data with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=utilization.gpu,temperature.gpu", "--format=csv,noheader,nounits")
	cmd.Env = append(os.Environ(), "HOME=/tmp")
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Printf("GPU detection timeout\n")
		} else {
			fmt.Printf("GPU detection failed: %v\n", err)
		}
		return nil
	}

	// Parse the output
	outputStr := strings.TrimSpace(string(output))
	lines := strings.Split(outputStr, "\n")
	if len(lines) == 0 {
		fmt.Printf("GPU: No output lines from nvidia-smi\n")
		return nil
	}

	// Get first GPU data
	parts := strings.Split(lines[0], ", ")
	if len(parts) != 2 {
		fmt.Printf("GPU: Unexpected output format: %s\n", lines[0])
		return nil
	}

	usage, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	temp, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	
	if err1 != nil || err2 != nil {
		fmt.Printf("GPU: Parse error - usage: %v, temp: %v\n", err1, err2)
		return nil
	}

	fmt.Printf("GPU: Detected - Usage: %d%%, Temp: %d°C\n", usage, temp)
	return &struct{ Usage, Temp int }{Usage: usage, Temp: temp}
}

func sendDataToDevice(data HardwareData) error {
	if selectedDevice == nil {
		return fmt.Errorf("no device selected")
	}

	// Format data according to Windows implementation
	// DispData array: [CpuUse, GpuUse, CpuTemp, GpuTemp, MemUse, DiskTemp]
	cpuUse := fmt.Sprintf("%d%%", data.CpuUsage)
	gpuUse := fmt.Sprintf("%d%%", data.GpuUsage)
	cpuTemp := fmt.Sprintf("%d°C", data.CpuTemp)
	gpuTemp := fmt.Sprintf("%d°C", data.GpuTemp)
	memUse := fmt.Sprintf("%d%%", data.MemoryUsage)
	diskTemp := fmt.Sprintf("%d°C", data.DiskTemp)

	payload := PCMonitorPayload{
		Command: "Device/UpdatePCParaInfo",
		ScreenList: []PCMonitorScreenItem{
			{
				LcdId:    selectedLcd - 1, // Convert to 0-based index
				DispData: []string{cpuUse, gpuUse, cpuTemp, gpuTemp, memUse, diskTemp},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s:80/post", selectedDevice.DevicePrivateIP)
	resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("device returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func waitForKey() {
	fmt.Println("\nPress any key to continue...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadByte()
}