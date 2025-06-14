package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

type AutoDeviceList struct {
	TotalData  int          `json:"TotalData"`
	DeviceList []AutoDevice `json:"DeviceList"`
}

type AutoDevice struct {
	DeviceName      string `json:"DeviceName"`
	DeviceId        int    `json:"DeviceId"`
	DevicePrivateIP string `json:"DevicePrivateIP"`
	DeviceMac       string `json:"DeviceMac"`
	Hardware        int    `json:"Hardware"`
}

type AutoHardwareData struct {
	CpuUsage    int
	GpuUsage    int
	CpuTemp     int
	GpuTemp     int
	MemoryUsage int
	DiskTemp    int
}

type AutoPCMonitorPayload struct {
	Command    string                     `json:"Command"`
	ScreenList []AutoPCMonitorScreenItem `json:"ScreenList"`
}

type AutoPCMonitorScreenItem struct {
	LcdId    int      `json:"LcdId"`
	DispData []string `json:"DispData"`
}

var (
	autoHttpClient = &http.Client{Timeout: 10 * time.Second}
)

func main() {
	fmt.Println("Divoom Auto Monitor - Sends data automatically to first found device")
	fmt.Println("===================================================================")

	// Find devices
	fmt.Println("Scanning for Divoom devices...")
	devices, err := findDevices()
	if err != nil {
		fmt.Printf("Error finding devices: %v\n", err)
		return
	}

	if len(devices) == 0 {
		fmt.Println("No Divoom devices found. Make sure your device is on the same network.")
		return
	}

	device := devices[0]
	fmt.Printf("Found device: %s (%s)\n", device.DeviceName, device.DevicePrivateIP)
	fmt.Println("Starting monitoring... Press Ctrl+C to stop")

	// Set up signal handling for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Start monitoring
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			data := getAutoHardwareData()
			if err := sendAutoDataToDevice(device, data); err != nil {
				fmt.Printf("Error sending data: %v\n", err)
			} else {
				fmt.Printf("Sent: CPU:%d%% %d°C GPU:%d%% %d°C MEM:%d%% DSK:%d°C\n",
					data.CpuUsage, data.CpuTemp, data.GpuUsage, data.GpuTemp,
					data.MemoryUsage, data.DiskTemp)
			}

		case <-c:
			fmt.Println("\nStopping monitor...")
			return
		}
	}
}

func findDevices() ([]AutoDevice, error) {
	resp, err := autoHttpClient.Get("http://app.divoom-gz.com/Device/ReturnSameLANDevice")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var deviceList AutoDeviceList
	if err := json.Unmarshal(body, &deviceList); err != nil {
		return nil, err
	}

	return deviceList.DeviceList, nil
}

func getAutoHardwareData() AutoHardwareData {
	data := AutoHardwareData{}

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

func sendAutoDataToDevice(device AutoDevice, data AutoHardwareData) error {
	// Format data according to Windows implementation
	// DispData array: [CpuUse, GpuUse, CpuTemp, GpuTemp, MemUse, DiskTemp]
	cpuUse := fmt.Sprintf("%d%%", data.CpuUsage)
	gpuUse := fmt.Sprintf("%d%%", data.GpuUsage)
	cpuTemp := fmt.Sprintf("%d°C", data.CpuTemp)
	gpuTemp := fmt.Sprintf("%d°C", data.GpuTemp)
	memUse := fmt.Sprintf("%d%%", data.MemoryUsage)
	diskTemp := fmt.Sprintf("%d°C", data.DiskTemp)

	payload := AutoPCMonitorPayload{
		Command: "Device/UpdatePCParaInfo",
		ScreenList: []AutoPCMonitorScreenItem{
			{
				LcdId:    0, // Default to first LCD
				DispData: []string{cpuUse, gpuUse, cpuTemp, gpuTemp, memUse, diskTemp},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s:80/post", device.DevicePrivateIP)
	resp, err := autoHttpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
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