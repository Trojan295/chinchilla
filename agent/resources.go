package agent

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// MemoryStats struct
type MemoryStats struct {
	Total     int
	Available int
}

func parseMeminfo(file *os.File) *MemoryStats {
	memoryStats := &MemoryStats{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		cols := strings.Split(line, " ")
		metric := cols[0][:len(cols[0])-1]
		value, _ := strconv.Atoi(cols[len(cols)-2])

		if metric == "MemTotal" {
			memoryStats.Total = value
		} else if metric == "MemAvailable" {
			memoryStats.Available = value
		}
	}
	return memoryStats
}

// GetMemoryStats returns MemoryStats for the machine
func GetMemoryStats() (*MemoryStats, error) {
	fp, err := os.Open("/proc/meminfo")
	if err != nil {
		log.Fatalf("Cannot open /proc/meminfo: %s\n", err)
		return nil, err
	}
	defer fp.Close()

	return parseMeminfo(fp), nil
}

// CPUStats struct
type CPUStats struct {
	Cores int
}

// GetCPUStats return CpuStats for the machine
func GetCPUStats() (CPUStats, error) {
	cpuStats := CPUStats{}

	out, err := exec.Command("nproc").Output()
	if err != nil {
		log.Fatalf("Cannot exec nproc command: %s\n", err)
		return cpuStats, err
	}

	cores, _ := strconv.Atoi(strings.TrimSpace(string(out)))
	cpuStats.Cores = cores

	return cpuStats, nil
}
