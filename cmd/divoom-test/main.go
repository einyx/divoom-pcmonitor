package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TestDeviceList struct {
	TotalData  int          `json:"TotalData"`
	DeviceList []TestDevice `json:"DeviceList"`
}

type TestDevice struct {
	DeviceName      string `json:"DeviceName"`
	DeviceId        int    `json:"DeviceId"`
	DevicePrivateIP string `json:"DevicePrivateIP"`
	DeviceMac       string `json:"DeviceMac"`
	Hardware        int    `json:"Hardware"`
}

type TestTextPayload struct {
	Command     string `json:"Command"`
	TextId      int    `json:"TextId"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	Dir         int    `json:"dir"`
	Font        int    `json:"font"`
	TextWidth   int    `json:"TextWidth"`
	Speed       int    `json:"speed"`
	TextString  string `json:"TextString"`
	Color       string `json:"color"`
	Align       int    `json:"align"`
}

func main() {
	fmt.Println("Testing Divoom Device Discovery and Communication")
	fmt.Println("================================================")

	client := &http.Client{Timeout: 10 * time.Second}

	// Test 1: Device Discovery
	fmt.Println("1. Testing device discovery...")
	resp, err := client.Get("http://app.divoom-gz.com/Device/ReturnSameLANDevice")
	if err != nil {
		fmt.Printf("   ERROR: Failed to connect to Divoom service: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   ERROR: Failed to read response: %v\n", err)
		return
	}

	fmt.Printf("   Response status: %d\n", resp.StatusCode)
	fmt.Printf("   Response body: %s\n", string(body))

	var deviceList TestDeviceList
	if err := json.Unmarshal(body, &deviceList); err != nil {
		fmt.Printf("   ERROR: Failed to parse JSON: %v\n", err)
		return
	}

	if len(deviceList.DeviceList) == 0 {
		fmt.Println("   No devices found. Make sure your Divoom device is on the same network.")
		return
	}

	fmt.Printf("   Found %d device(s):\n", len(deviceList.DeviceList))
	for i, device := range deviceList.DeviceList {
		fmt.Printf("   %d. %s (%s) - Hardware: %d\n", 
			i+1, device.DeviceName, device.DevicePrivateIP, device.Hardware)
	}

	// Test 2: Send test data to first device
	device := deviceList.DeviceList[0]
	fmt.Printf("\n2. Testing communication with %s...\n", device.DeviceName)
	
	testText := "TEST CPU:50% 60C MEM:75%"
	payload := TestTextPayload{
		Command:    "Draw/SendHttpText",
		TextId:     1,
		X:          0,
		Y:          0,
		Dir:        0,
		Font:       1,
		TextWidth:  64,
		Speed:      100,
		TextString: testText,
		Color:      "#00FF00",
		Align:      1,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("   ERROR: Failed to marshal JSON: %v\n", err)
		return
	}

	fmt.Printf("   Sending payload: %s\n", string(jsonData))

	url := fmt.Sprintf("http://%s:80/post", device.DevicePrivateIP)
	fmt.Printf("   URL: %s\n", url)
	
	resp2, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("   ERROR: POST request failed: %v\n", err)
		return
	}
	defer resp2.Body.Close()

	respBody, err := io.ReadAll(resp2.Body)
	if err != nil {
		fmt.Printf("   ERROR: Failed to read response: %v\n", err)
		return
	}

	fmt.Printf("   Response status: %d\n", resp2.StatusCode)
	fmt.Printf("   Response body: %s\n", string(respBody))

	if resp2.StatusCode != http.StatusOK {
		fmt.Printf("   Device returned error status %d\n", resp2.StatusCode)
	} else {
		fmt.Println("   SUCCESS: Test message sent to device!")
	}

	// Test 3: Try Windows command format
	fmt.Println("\n3. Testing Windows command format...")
	
	// Try Windows-style PC monitoring command
	windowsPayload := map[string]interface{}{
		"Command": "Device/UpdatePCParaInfo",
		"ScreenList": []map[string]interface{}{
			{
				"LcdId":    0,
				"DispData": []string{"50%", "25%", "65C", "70C", "80%", "45C"},
			},
		},
	}
	
	jsonData2, _ := json.Marshal(windowsPayload)
	fmt.Printf("   Trying Windows command: %s\n", string(jsonData2))
	
	resp3, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData2))
	if err != nil {
		fmt.Printf("   Windows command failed: %v\n", err)
	} else {
		defer resp3.Body.Close()
		respBody3, _ := io.ReadAll(resp3.Body)
		fmt.Printf("   Windows response (%d): %s\n", resp3.StatusCode, string(respBody3))
	}
}