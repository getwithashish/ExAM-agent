package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os/exec"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

type CPUInfo struct {
	CPU        int      `json:"cpu"`
	VendorID   string   `json:"vendorId"`
	Family     string   `json:"family"`
	Model      string   `json:"model"`
	Stepping   int      `json:"stepping"`
	PhysicalID string   `json:"physicalId"`
	CoreID     string   `json:"coreId"`
	Cores      int      `json:"cores"`
	ModelName  string   `json:"modelName"`
	Mhz        float64  `json:"mhz"`
	CacheSize  int      `json:"cacheSize"`
	Flags      []string `json:"flags"`
	Microcode  string   `json:"microcode"`
}

type MemoryInfo struct {
	Total uint64 `json:"total"`
}

type StorageInfo struct {
	Total uint64 `json:"total"`
}

type HostInfo struct {
	Os      string `json:"os"`
	Version string `json:"platformVersion"`
}

type SystemInfo struct {
	CPUModel       string `json:"cpuModel"`
	TotalMemoryGB  float64 `json:"totalMemoryGB"`
	TotalStorageGB float64 `json:"totalStorageGB"`
	ProductModel   string `json:"productModel"`
	OS             string `json:"os"`
	OSVersion      string `json:"osVersion"`
	SerialNumber   string `json:"serialNumber"`
	Manufacturer   string `json:"manufacturer"`
}

func main() {

	cpuModel, _ := retrieveCPUInfo()

	totalMemoryBytes, _ := retrieveMemoryInfo()
	totalMemoryBytes_64 := uint64(totalMemoryBytes)
	totalMemoryGB := math.Round(bytesToGB(totalMemoryBytes_64))

	totalStorageBytes, _ := retrieveStorageInfo()
	totalStorageBytes_64 := uint64(totalStorageBytes)
	totalStorageGB := math.Round(bytesToGB(totalStorageBytes_64))

	productModel, _ := retrieveProductModel()

	os, osVersion, _ := retrieveHostInfo()
	if os == "windows"{
		os = "WINDOWS"
	}

	serialNumber, _ := retrieveSerialNumber()
	manufacturer, _ := retrieveManufacturer()

	info := SystemInfo{
		CPUModel:       cpuModel,
		TotalMemoryGB:  totalMemoryGB,
		TotalStorageGB: totalStorageGB,
		ProductModel:   productModel,
		OS:             os,
		OSVersion:      osVersion,
		SerialNumber:   serialNumber,
		Manufacturer:   manufacturer,
	}

	data, err := json.Marshal(info)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	sendRequest(data)

}

func retrieveHostInfo() (string, string, error) {
	hostInfo, err := host.Info()
	if err != nil {
		fmt.Println("Failed to retrieve host info:", err)
		return "", "", err
	}
	hostInfoJson, err := infoToString(hostInfo)
	if err != nil {
		fmt.Println("Failed to convert to JSON:", err)
		return "", "", err
	}

	var hostOsInfo HostInfo
	err_conv := json.Unmarshal([]byte(hostInfoJson), &hostOsInfo)
	if err_conv != nil {
		fmt.Println("Error unmarshalling storage info:", err_conv)
		return "", "", err_conv
	}

	return hostOsInfo.Os, hostOsInfo.Version, err
}

func retrieveStorageInfo() (uint64, error) {
	diskInfo, err := disk.Usage("/")
	if err != nil {
		fmt.Println("Failed to retrieve disk info:", err)
		return 0, err
	}
	storageInfoJson, err := infoToString(diskInfo)
	if err != nil {
		fmt.Println("Failed to convert to JSON:", err)
		return 0, err
	}

	var storageInfo StorageInfo
	err_conv := json.Unmarshal([]byte(storageInfoJson), &storageInfo)
	if err_conv != nil {
		fmt.Println("Error unmarshalling storage info:", err_conv)
		return 0, err_conv
	}
	totalStorage := storageInfo.Total
	fmt.Println("Total storage:", totalStorage)
	return totalStorage, err
}

func retrieveManufacturer() (string, error) {
	cmd := exec.Command("wmic", "baseboard", "get", "manufacturer")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing command:", err)
		return "", err
	}

	fmt.Println("Command output:")
	fmt.Println(out.String())
	lines := strings.Split(out.String(), "\n")
	if len(lines) >= 2 {
		return lines[1], nil
	}
	return "", err
}

func retrieveSerialNumber() (string, error) {
	cmd := exec.Command("wmic", "bios", "get", "serialnumber")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing command:", err)
		return "", err
	}

	fmt.Println("Command output:")
	fmt.Println(out.String())
	lines := strings.Split(out.String(), "\n")
	if len(lines) >= 2 {
		return lines[1], nil
	}
	return "", err
}

func retrieveProductModel() (string, error) {
	cmd := exec.Command("wmic", "csproduct", "get", "name")

	var out bytes.Buffer
	cmd.Stdout = &out

	// Execute command
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing command:", err)
		return "", err
	}

	// Print command output
	fmt.Println("Command output:")
	fmt.Println(out.String())
	lines := strings.Split(out.String(), "\n")
	if len(lines) >= 2 {
		return lines[1], nil
	}
	return "", err
}

func retrieveCPUInfo() (string, error) {
	cpuInfo, err := cpu.Info()
	if err != nil {
		fmt.Println("Failed to retrieve CPU info:", err)
		return "", err
	}
	cpuInfoJson, err := infoToString(cpuInfo)
	if err != nil {
		fmt.Println("Failed to convert to JSON:", err)
		return "", err
	}
	cpuModelName, err := extractCPUModelName(cpuInfoJson)
	if err != nil {
		fmt.Println("Error extracting model name:", err)
		return "", err
	}
	fmt.Println("CPU Model Name: ", cpuModelName)

	return cpuModelName, err

}

func retrieveMemoryInfo() (uint64, error) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println("Failed to retrieve memory info:", err)
		return 0, err
	}
	memoryInfoJson, err := infoToString(memInfo)
	if err != nil {
		fmt.Println("Failed to convert to JSON:", err)
		return 0, err
	}

	var memoryInfo MemoryInfo
	err_conv := json.Unmarshal([]byte(memoryInfoJson), &memoryInfo)
	if err_conv != nil {
		fmt.Println("Error unmarshalling memory info:", err_conv)
		return 0, err_conv
	}
	totalMemory := memoryInfo.Total
	fmt.Println("Total memory:", totalMemory)
	return totalMemory, err
}

func infoToString(info interface{}) (string, error) {
	// Marshal the CPU info slice to JSON
	infoJSON, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", err
	}

	return string(infoJSON), nil
}

func extractCPUModelName(cpuInfoJSON string) (string, error) {
	var cpuInfo []CPUInfo
	err := json.Unmarshal([]byte(cpuInfoJSON), &cpuInfo)
	if err != nil {
		return "", err
	}

	if len(cpuInfo) > 0 {
		return cpuInfo[0].ModelName, nil
	}
	return "", fmt.Errorf("no CPU information found")
}

func bytesToGB(bytes uint64) float64 {
	gigabyte := float64(1024 * 1024 * 1024)
	return float64(bytes) / gigabyte
}

func sendRequest(systemInfo []byte) (error){
	url := "http://localhost:8000/api/v1/asset/useragent"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(systemInfo))
	if err != nil {
		fmt.Println("Error sending JSON:", err)
		return err
	}
	defer resp.Body.Close()

	fmt.Println("Response status:", resp.Status)
	return nil
} 
