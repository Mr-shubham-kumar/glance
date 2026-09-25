package glance

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const metricsWindow = time.Hour
const metricsSampleInterval = 2 * time.Minute

var instanceStartedAt = time.Now()
var pageContentLoads atomic.Uint64    // page fragments, not Render's total requests
var pageContentFailures atomic.Uint64 // HTTP 500 rendering failures; upstream widget notices are separate
var revisionPattern = regexp.MustCompile(`^[a-fA-F0-9]{7,40}$`)

type instanceSample struct {
	At       time.Time `json:"at"`
	CPU      float64   `json:"cpu"`
	CPUReady bool      `json:"cpuReady"`
	Memory   float64   `json:"memory"`
	X        float64   `json:"x"` // time position on a 0-120 SVG axis, oldest on the left
}

type metricsHistory struct {
	sync.Mutex
	lastTime    time.Time
	lastCPU     uint64
	lastPercent float64
	cpuReady    bool
	samples     []instanceSample
}

var instanceHistory metricsHistory

func clampPercent(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

// record is request-driven. CPU is averaged over the interval between actual
// samples, never over an arbitrary last HTTP request. A long idle gap is a gap,
// not a fake zero on the graph. All history is memory-only.
func (h *metricsHistory) record(now time.Time, usage uint64, memory, memoryMax uint64, cpuCores float64) (float64, bool, []instanceSample) {
	h.Lock()
	defer h.Unlock()
	cutoff := now.Add(-metricsWindow)
	recent := h.samples[:0]
	for _, s := range h.samples {
		if !s.At.Before(cutoff) {
			recent = append(recent, s)
		}
	}
	h.samples = recent
	if h.lastTime.IsZero() || now.Sub(h.lastTime) >= metricsSampleInterval {
		ready := !h.lastTime.IsZero() && now.Sub(h.lastTime) <= 10*time.Minute && usage >= h.lastCPU
		if ready {
			h.lastPercent = clampPercent(float64(usage-h.lastCPU) / 1e6 / now.Sub(h.lastTime).Seconds() / cpuCores * 100)
		}
		h.cpuReady = ready
		h.lastCPU, h.lastTime = usage, now
		h.samples = append(h.samples, instanceSample{At: now, CPU: h.lastPercent, CPUReady: ready, Memory: clampPercent(float64(memory) / float64(memoryMax) * 100)})
	}
	if len(h.samples) > 30 {
		h.samples = h.samples[len(h.samples)-30:]
	}
	out := make([]instanceSample, len(h.samples))
	copy(out, h.samples)
	for i := range out {
		out[i].X = clampPercent((1-now.Sub(out[i].At).Seconds()/metricsWindow.Seconds())*100) * 1.2
	}
	return h.lastPercent, h.cpuReady, out
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
	cores := quota / period
	cpu, ready, samples := instanceHistory.record(now, usage, memory, memoryMax, cores)
	revision := os.Getenv("RENDER_GIT_COMMIT")
	if !revisionPattern.MatchString(revision) {
		revision = "unavailable"
	}
	remaining := uint64(0)
	if memory < memoryMax {
		remaining = memoryMax - memory
	}
	json.NewEncoder(w).Encode(map[string]any{
		"available": true, "observedAt": now, "startedAt": instanceStartedAt,
		"uptimeMinutes": int(now.Sub(instanceStartedAt).Minutes()),
		"revision":      revision, "pageContentLoads": pageContentLoads.Load(), "pageContentFailures": pageContentFailures.Load(),
		"cpuReady": ready, "cpuPercent": cpu, "cpuHeadroomPercent": 100 - cpu,
		"cpuLimitCores":      cores,
		"memoryMiB":          float64(memory) / 1048576,
		"memoryLimitMiB":     float64(memoryMax) / 1048576,
		"memoryRemainingMiB": float64(remaining) / 1048576,
		"memoryPercent":      float64(memory) / float64(memoryMax) * 100,
		"samples":            samples,
	})
}
