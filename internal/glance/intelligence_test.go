package glance

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPublicIntelligenceUnpublishedFallback(t *testing.T) {
	intelligenceState.Lock()
	intelligenceState.body = nil
	intelligenceState.checked = time.Now()
	intelligenceState.fetchOK = false
	intelligenceState.Unlock()
	response := httptest.NewRecorder()
	handlePublicIntelligence(response, httptest.NewRequest("GET", "/api/intelligence", nil))
	var result map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["delivery_status"] != "unavailable" || result["data_fresh"] != false {
		t.Fatalf("unexpected fallback: %v", result)
	}
}

func TestPublicIntelligenceRejectsStaleBrief(t *testing.T) {
	old := time.Now().Add(-13 * time.Hour).UTC().Format(time.RFC3339)
	intelligenceState.Lock()
	intelligenceState.body, _ = json.Marshal(map[string]any{"schema_version": 1, "generated_at": old,
		"data_fresh": true, "brief": []any{map[string]any{"title": "Old recommendation"}}})
	intelligenceState.checked = time.Now()
	intelligenceState.fetchOK = true
	intelligenceState.Unlock()
	response := httptest.NewRecorder()
	handlePublicIntelligence(response, httptest.NewRequest("GET", "/api/intelligence", nil))
	var result map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["delivery_status"] != "stale" || result["data_fresh"] != false || len(result["brief"].([]any)) != 0 {
		t.Fatalf("stale recommendation was shown: %v", result)
	}
}
