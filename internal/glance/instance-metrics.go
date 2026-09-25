package glance

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Instance metrics come from the container's cgroup, not Render's account API.
// Only aggregate CPU/memory counters are published. History is in-memory and
// resets whenever a free-tier instance sleeps or is redeployed.
type instanceSample struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
}

var instanceHistory struct {
	sync.Mutex
	lastTime          time.Time
	lastTimeForSample time.Time
	lastCPU           uint64
	samples           []instanceSample
}

func readCgroupUint(path string) (uint64, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(strings.TrimSpace(string(raw)), 10, 64)
}

func handleInstanceMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	memory, errMemory := readCgroupUint("/sys/fs/cgroup/memory.current")
	memoryMax, errMax := readCgroupUint("/sys/fs/cgroup/memory.max")
	rawCPU, errCPU := os.ReadFile("/sys/fs/cgroup/cpu.stat")
	rawQuota, errQuota := os.ReadFile("/sys/fs/cgroup/cpu.max")
	if errMemory != nil || errMax != nil || errCPU != nil || errQuota != nil || memoryMax == 0 {
		json.NewEncoder(w).Encode(map[string]any{"available": false})
		return
	}
	var usage uint64
	for _, line := range strings.Split(string(rawCPU), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "usage_usec" {
			usage, _ = strconv.ParseUint(fields[1], 10, 64)
		}
	}
	quotaFields := strings.Fields(string(rawQuota))
	if len(quotaFields) != 2 || quotaFields[0] == "max" || usage == 0 {
		json.NewEncoder(w).Encode(map[string]any{"available": false})
		return
	}
	quota, qerr := strconv.ParseFloat(quotaFields[0], 64)
	period, perr := strconv.ParseFloat(quotaFields[1], 64)
	if qerr != nil || perr != nil || period <= 0 || quota <= 0 {
		json.NewEncoder(w).Encode(map[string]any{"available": false})
		return
	}
	now := time.Now()
	instanceHistory.Lock()
	cpu := 0.0
	if !instanceHistory.lastTime.IsZero() && usage >= instanceHistory.lastCPU {
		elapsed := now.Sub(instanceHistory.lastTime).Seconds()
		if elapsed > 0 {
			cpu = float64(usage-instanceHistory.lastCPU) / 1e6 / elapsed / (quota / period) * 100
		}
	}
	if cpu > 100 {
		cpu = 100
	}
	if cpu < 0 {
		cpu = 0
	}
	instanceHistory.lastCPU, instanceHistory.lastTime = usage, now
	// Avoid duplicate samples from repeated concurrent requests; roughly two
	// minutes per point with a 60-minute maximum sliding window.
	if len(instanceHistory.samples) == 0 || now.Sub(instanceHistory.lastTimeForSample) >= 2*time.Minute {
		instanceHistory.samples = append(instanceHistory.samples, instanceSample{cpu, float64(memory) / float64(memoryMax) * 100})
		instanceHistory.lastTimeForSample = now
	}
	if len(instanceHistory.samples) > 30 {
		instanceHistory.samples = instanceHistory.samples[len(instanceHistory.samples)-30:]
	}
	samples := append([]instanceSample(nil), instanceHistory.samples...)
	instanceHistory.Unlock()
	json.NewEncoder(w).Encode(map[string]any{
		"available": true, "cpuPercent": cpu,
		"memoryMiB": float64(memory) / 1048576, "memoryLimitMiB": float64(memoryMax) / 1048576,
		"memoryPercent": float64(memory) / float64(memoryMax) * 100,
		"cpuLimitCores": quota / period, "samples": samples,
	})
}
