package glance

import (
	"testing"
	"time"
)

func TestMetricsSamplesAreWindowedAndAveragedBetweenSamples(t *testing.T) {
	var h metricsHistory
	start := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	cpu, ready, samples := h.record(start, 1_000_000, 100, 1000, .1)
	if ready || len(samples) != 1 || samples[0].CPUReady {
		t.Fatalf("first reading must warm up: %v %v %v", cpu, ready, samples)
	}
	// A second request too soon cannot change the CPU sampling baseline.
	_, ready, samples = h.record(start.Add(time.Minute), 2_000_000, 150, 1000, .1)
	if ready || len(samples) != 1 {
		t.Fatalf("too-early sample: %v %v", ready, samples)
	}
	cpu, ready, samples = h.record(start.Add(2*time.Minute), 7_000_000, 200, 1000, .1)
	if !ready || len(samples) != 2 || cpu < 49.9 || cpu > 50.1 {
		t.Fatalf("2m CPU average should be 50%%: %v %v %v", cpu, ready, samples)
	}
	_, ready, samples = h.record(start.Add(70*time.Minute), 8_000_000, 230, 1000, .1)
	if ready || len(samples) != 1 || samples[0].CPUReady {
		t.Fatalf("after idle must show gap and prune expired samples: %v %v", ready, samples)
	}
	if samples[0].X < 119 {
		t.Errorf("current point should be on right edge: %f", samples[0].X)
	}
}

func TestClampPercent(t *testing.T) {
	if clampPercent(-1) != 0 || clampPercent(125) != 100 || clampPercent(33) != 33 {
		t.Fatal("invalid normalization")
	}
}
