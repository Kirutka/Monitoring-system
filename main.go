package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// Для Linux-специфичных функций
const (
	tempFile = "/sys/class/thermal/thermal_zone0/temp"
	memInfo  = "/proc/meminfo"
	clearCmd = "clear"
)

func main() {
	for {
		clearScreen()
		printSystemInfo()
		time.Sleep(2 * time.Second)
	}
}

func clearScreen() {
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd := exec.Command(clearCmd)
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func printSystemInfo() {
	// CPU
	printCPUInfo()

	// Память
	printMemoryInfo()

	// Диск
	printDiskInfo()

	// Сеть
	printNetworkInfo()

	// Температура (Linux)
	if runtime.GOOS == "linux" {
		printTemperature()
	}

	// Загрузка системы
	printLoadInfo()
}

func printCPUInfo() {
	percent, _ := cpu.Percent(time.Second, false)
	info, _ := cpu.Info()

	fmt.Println("=== CPU ===")
	fmt.Printf("Model: %s\n", info[0].ModelName)
	fmt.Printf("Cores: %d\n", runtime.NumCPU())
	fmt.Printf("Usage: %.2f%%\n", percent[0])
}

func printMemoryInfo() {
	vmem, _ := mem.VirtualMemory()

	fmt.Println("\n=== Memory ===")
	fmt.Printf("Total: %s\n", formatBytes(vmem.Total))
	fmt.Printf("Available: %s\n", formatBytes(vmem.Available))
	fmt.Printf("Used: %s (%.2f%%)\n",
		formatBytes(vmem.Used),
		vmem.UsedPercent)
}

func printDiskInfo() {
	partitions, _ := disk.Partitions(true)

	fmt.Println("\n=== Disk ===")
	for _, p := range partitions {
		usage, _ := disk.Usage(p.Mountpoint)
		fmt.Printf("[%s] %s/%s (%.2f%%)\n",
			p.Mountpoint,
			formatBytes(usage.Used),
			formatBytes(usage.Total),
			usage.UsedPercent)
	}
}

func printNetworkInfo() {
	io, _ := net.IOCounters(true)

	fmt.Println("\n=== Network ===")
	for _, iface := range io {
		if iface.BytesSent > 0 || iface.BytesRecv > 0 {
			fmt.Printf("%s:\n", iface.Name)
			fmt.Printf("  Sent: %s\n", formatBytes(iface.BytesSent))
			fmt.Printf("  Recv: %s\n", formatBytes(iface.BytesRecv))
		}
	}
}

func printTemperature() {
	if temp, err := readCPUTemperature(); err == nil {
		fmt.Printf("\n=== Temperature ===\nCPU: %.1f°C\n", temp)
	}
}

func printLoadInfo() {
	avg, _ := load.Avg()

	fmt.Println("\n=== System Load ===")
	fmt.Printf("1m: %.2f, 5m: %.2f, 15m: %.2f\n",
		avg.Load1,
		avg.Load5,
		avg.Load15)
}

func readCPUTemperature() (float64, error) {
	if runtime.GOOS != "linux" {
		return 0, fmt.Errorf("not supported")
	}

	data, err := os.ReadFile(tempFile)
	if err != nil {
		return 0, err
	}

	var temp float64
	fmt.Sscanf(string(data), "%f", &temp)
	return temp / 1000, nil
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(bytes)/float64(div),
		"KMGTPE"[exp])
}
