package glance

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"
)

const publicIntelligenceURL = "https://raw.githubusercontent.com/Mr-shubham-kumar/glance/signal-data/snapshot.json"

var intelligenceState = struct {
	sync.Mutex
	body    []byte
	checked time.Time
	fetchOK bool
}{}

var intelligenceClient = &http.Client{Timeout: 5 * time.Second}

func handlePublicIntelligence(w http.ResponseWriter, _ *http.Request) {
	intelligenceState.Lock()
	defer intelligenceState.Unlock()

	if time.Since(intelligenceState.checked) >= 30*time.Minute || intelligenceState.checked.IsZero() {
		intelligenceState.checked = time.Now()
		intelligenceState.fetchOK = false
		if res, err := intelligenceClient.Get(publicIntelligenceURL); err == nil {
			if res.StatusCode == http.StatusOK {
				if body, err := io.ReadAll(io.LimitReader(res.Body, 512*1024+1)); err == nil && len(body) <= 512*1024 {
					var snapshot struct {
						SchemaVersion int    `json:"schema_version"`
						GeneratedAt   string `json:"generated_at"`
					}
					if json.Unmarshal(body, &snapshot) == nil && snapshot.SchemaVersion == 1 && snapshot.GeneratedAt != "" {
						intelligenceState.body = body
						intelligenceState.fetchOK = true
					}
				}
			}
			res.Body.Close()
		}
	}

	var payload map[string]any
	if len(intelligenceState.body) > 0 && json.Unmarshal(intelligenceState.body, &payload) == nil {
		generated, _ := payload["generated_at"].(string)
		at, err := time.Parse(time.RFC3339, generated)
		if err != nil || time.Since(at) > 12*time.Hour || !intelligenceState.fetchOK {
			payload["data_fresh"] = false
			payload["brief"] = []any{}
			payload["delivery_status"] = "stale"
			alert := map[string]any{"title": "Intelligence snapshot is stale", "url": publicIntelligenceURL,
				"reason": "The latest public snapshot could not be verified within 12 hours", "kind": "delivery"}
			if existing, ok := payload["alerts"].([]any); ok {
				payload["alerts"] = append([]any{alert}, existing...)
			} else {
				payload["alerts"] = []any{alert}
			}
		} else {
			payload["delivery_status"] = "ok"
		}
	} else {
		payload = map[string]any{
			"schema_version": 1, "generated_at": "unavailable", "data_fresh": false,
			"delivery_status": "unavailable", "brief": []any{},
			"items":   map[string]any{"radar": []any{}, "experiments": []any{}},
			"receipt": map[string]any{"input_count": 0, "unique_count": 0, "failed_sources": 0},
			"alerts": []any{map[string]any{"title": "Intelligence snapshot unavailable", "url": publicIntelligenceURL,
				"reason": "No published snapshot is available yet", "kind": "delivery"}},
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(payload)
}
