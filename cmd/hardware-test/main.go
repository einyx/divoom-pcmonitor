package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

type TestHardwareData struct {
	CpuUsage    int
	GpuUsage    int
	CpuTemp     int
	GpuTemp     int
	MemoryUsage int
	DiskTemp    int
}

func main() {
	fmt.Println("Testing Hardware Data Collection")
	fmt.Println("===============================")

	data := getTestHardwareData()
	
	fmt.Printf("CPU Usage: %d%%\n", data.CpuUsage)
	fmt.Printf("CPU Temp: %d°C\n", data.CpuTemp)
	fmt.Printf("GPU Usage: %d%%\n", data.GpuUsage)
	fmt.Printf("GPU Temp: %d°C\n", data.GpuTemp)
	fmt.Printf("Memory Usage: %d%%\n", data.MemoryUsage)
	fmt.Printf("Disk Temp: %d°C\n", data.DiskTemp)

	// Format as text for Divoom
	textData := fmt.Sprintf("CPU:%d%% %d°C GPU:%d%% %d°C MEM:%d%% DSK:%d°C",
		data.CpuUsage, data.CpuTemp,
		data.GpuUsage, data.GpuTemp,
		data.MemoryUsage, data.DiskTemp)
	
	fmt.Printf("\nFormatted text: %s\n", textData)
	fmt.Printf("Text length: %d chars\n", len(textData))
}

func getTestHardwareData() TestHardwareData {
	data := TestHardwareData{}

	fmt.Println("\nCollecting hardware data...")

	// CPU Usage
	fmt.Print("Getting CPU usage... ")
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		data.CpuUsage = int(cpuPercent[0])
		fmt.Printf("OK (%d%%)\n", data.CpuUsage)
	} else {
		fmt.Printf("ERROR: %v\n", err)
	}

	// Get all temperature sensors
	fmt.Print("Getting temperature sensors... ")
	temps, err := host.SensorsTemperatures()
	if err == nil {
		fmt.Printf("OK (found %d sensors)\n", len(temps))
		
		fmt.Println("Available sensors:")
		for _, temp := range temps {
			fmt.Printf("  %s: %.1f°C\n", temp.SensorKey, temp.Temperature)
		}

		// CPU Temperature
		for _, temp := range temps {
			if strings.Contains(strings.ToLower(temp.SensorKey), "cpu") || 
			   strings.Contains(strings.ToLower(temp.SensorKey), "package") ||
			   strings.Contains(strings.ToLower(temp.SensorKey), "core") {
				data.CpuTemp = int(temp.Temperature)
				fmt.Printf("Using CPU temp from %s: %d°C\n", temp.SensorKey, data.CpuTemp)
				break
			}
		}

		// Disk Temperature
		for _, temp := range temps {
			if strings.Contains(strings.ToLower(temp.SensorKey), "nvme") || 
			   strings.Contains(strings.ToLower(temp.SensorKey), "sda") ||
			   strings.Contains(strings.ToLower(temp.SensorKey), "disk") {
				data.DiskTemp = int(temp.Temperature)
				fmt.Printf("Using disk temp from %s: %d°C\n", temp.SensorKey, data.DiskTemp)
				break
			}
		}
	} else {
		fmt.Printf("ERROR: %v\n", err)
	}

	// Memory Usage
	fmt.Print("Getting memory usage... ")
	vmStat, err := mem.VirtualMemory()
	if err == nil {
		data.MemoryUsage = int(vmStat.UsedPercent)
		fmt.Printf("OK (%d%%)\n", data.MemoryUsage)
	} else {
		fmt.Printf("ERROR: %v\n", err)
	}

	// GPU data (placeholder)
	data.GpuUsage = 0
	data.GpuTemp = 0
	fmt.Println("GPU data: Not implemented (would need nvidia-smi parsing)")

	return data
}