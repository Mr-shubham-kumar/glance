package glance

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

const (
	publicIntelligenceURL     = "https://raw.githubusercontent.com/Mr-shubham-kumar/glance/signal-data/snapshot.json"
	intelligenceCacheTTL      = 30 * time.Minute
	intelligenceMaxAge        = 12 * time.Hour
	intelligenceMaxBodySize   = 512 * 1024
	intelligenceMaxFutureSkew = 5 * time.Minute
)

var intelligenceClient = &http.Client{Timeout: 5 * time.Second}

var intelligenceFetchSnapshot = func() ([]byte, error) {
	response, err := intelligenceClient.Get(publicIntelligenceURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("snapshot request returned %s", response.Status)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, intelligenceMaxBodySize+1))
	if err != nil {
		return nil, err
	}
	if len(body) > intelligenceMaxBodySize {
		return nil, errors.New("snapshot exceeds 512 KiB")
	}
	return body, nil
}

var intelligenceState = struct {
	sync.Mutex
	body     []byte
	checked  time.Time
	fetchOK  bool
	inflight bool
}{}

func jsonHasShape(raw json.RawMessage, shape byte) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) > 0 && trimmed[0] == shape
}

func validateIntelligenceSnapshot(body []byte, now time.Time) bool {
	var snapshot struct {
		SchemaVersion int             `json:"schema_version"`
		GeneratedAt   string          `json:"generated_at"`
		DataFresh    *bool           `json:"data_fresh"`
		Sources      json.RawMessage `json:"sources"`
		Items        json.RawMessage `json:"items"`
		Alerts       json.RawMessage `json:"alerts"`
		Receipt      json.RawMessage `json:"receipt"`
	}
	if json.Unmarshal(body, &snapshot) != nil {
		return false
	}
	if snapshot.SchemaVersion != 1 || snapshot.DataFresh == nil || snapshot.GeneratedAt == "" {
		return false
	}
	generated, err := time.Parse(time.RFC3339, snapshot.GeneratedAt)
	if err != nil || generated.After(now.Add(intelligenceMaxFutureSkew)) {
		return false
	}
	if !jsonHasShape(snapshot.Sources, '[') || !jsonHasShape(snapshot.Alerts, '[') ||
		!jsonHasShape(snapshot.Receipt, '{') || !jsonHasShape(snapshot.Items, '{') {
		return false
	}
	var items struct {
		Radar       json.RawMessage `json:"radar"`
		Experiments json.RawMessage `json:"experiments"`
	}
	if json.Unmarshal(snapshot.Items, &items) != nil ||
		!jsonHasShape(items.Radar, '[') || !jsonHasShape(items.Experiments, '[') {
		return false
	}
	return true
}

func refreshIntelligence() {
	intelligenceState.Lock()
	if !intelligenceState.checked.IsZero() && time.Since(intelligenceState.checked) < intelligenceCacheTTL {
		intelligenceState.Unlock()
		return
	}
	if intelligenceState.inflight {
		intelligenceState.Unlock()
		return
	}
	intelligenceState.inflight = true
	// Stamp before the request so a failing source does not retry on every call.
	intelligenceState.checked = time.Now()
	intelligenceState.Unlock()

	body, err := intelligenceFetchSnapshot()

	intelligenceState.Lock()
	defer intelligenceState.Unlock()
	intelligenceState.inflight = false
	if err != nil {
		slog.Warn("Failed to fetch intelligence snapshot", "url", publicIntelligenceURL, "error", err)
		intelligenceState.fetchOK = false
		return
	}
	if !validateIntelligenceSnapshot(body, time.Now()) {
		slog.Warn("Intelligence snapshot failed validation", "url", publicIntelligenceURL, "bytes", len(body))
		intelligenceState.fetchOK = false
		return
	}
	intelligenceState.body = body
	intelligenceState.fetchOK = true
}

func unavailableIntelligence() map[string]any {
	return map[string]any{
		"schema_version":  1,
		"generated_at":    "unavailable",
		"data_fresh":      false,
		"delivery_status": "unavailable",
		"brief":           []any{},
		"sources":         []any{},
		"items":           map[string]any{"radar": []any{}, "experiments": []any{}},
		"receipt": map[string]any{
			"input_count": 0, "unique_count": 0, "radar_count": 0,
			"experiment_count": 0, "excluded_count": 0,
			"failed_sources": 0, "stale_sources": 0,
		},
		"alerts": []any{map[string]any{
			"title":  "Intelligence snapshot unavailable",
			"url":    publicIntelligenceURL,
			"reason": "No published snapshot is available yet",
			"kind":   "delivery",
		}},
	}
}

// keepHealthBrief drops recommendation lines that must not appear when the
// snapshot carries no fresh source data, while preserving the health pointer.
func keepHealthBrief(brief any) []any {
	entries, ok := brief.([]any)
	if !ok {
		return []any{}
	}
	kept := make([]any, 0, len(entries))
	for _, entry := range entries {
		if item, ok := entry.(map[string]any); ok && item["label"] == "Check" {
			kept = append(kept, entry)
		}
	}
	return kept
}

func applyIntelligenceDeliveryState(payload map[string]any, fetchOK bool, now time.Time) {
	generated, _ := payload["generated_at"].(string)
	at, err := time.Parse(time.RFC3339, generated)
	stale := !fetchOK || err != nil || now.Sub(at) > intelligenceMaxAge || at.After(now.Add(intelligenceMaxFutureSkew))
	if stale {
		reason := "The latest public snapshot could not be re-verified within the last 30 minutes"
		switch {
		case err != nil:
			reason = "The latest public snapshot has an unreadable collection time"
		case at.After(now.Add(intelligenceMaxFutureSkew)):
			reason = "The latest public snapshot has a collection time from the future"
		case now.Sub(at) > intelligenceMaxAge:
			reason = "The latest public snapshot is older than the 12-hour freshness window"
		}
		payload["data_fresh"] = false
		payload["delivery_status"] = "stale"
		payload["brief"] = []any{}
		alert := map[string]any{"title": "Intelligence snapshot is stale", "url": publicIntelligenceURL,
			"reason": reason, "kind": "delivery"}
		if existing, ok := payload["alerts"].([]any); ok {
			payload["alerts"] = append([]any{alert}, existing...)
		} else {
			payload["alerts"] = []any{alert}
		}
		return
	}
	payload["delivery_status"] = "ok"
	if fresh, ok := payload["data_fresh"].(bool); !ok || !fresh {
		payload["data_fresh"] = false
		payload["brief"] = keepHealthBrief(payload["brief"])
	}
}

func handlePublicIntelligence(w http.ResponseWriter, _ *http.Request) {
	refreshIntelligence()

	intelligenceState.Lock()
	body := append([]byte(nil), intelligenceState.body...)
	fetchOK := intelligenceState.fetchOK
	intelligenceState.Unlock()

	var payload map[string]any
	if len(body) > 0 && json.Unmarshal(body, &payload) == nil {
		applyIntelligenceDeliveryState(payload, fetchOK, time.Now())
	} else {
		payload = unavailableIntelligence()
	}
	delete(payload, "cache")
	delete(payload, "docs")

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(payload)
}

func resetIntelligenceCache() {
	intelligenceState.Lock()
	defer intelligenceState.Unlock()
	intelligenceState.body = nil
	intelligenceState.checked = time.Time{}
	intelligenceState.fetchOK = false
	intelligenceState.inflight = false
}
