package glance

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"
)

func swapIntelligenceFetch(t *testing.T, fn func() ([]byte, error)) {
	t.Helper()
	previous := intelligenceFetchSnapshot
	intelligenceFetchSnapshot = fn
	resetIntelligenceCache()
	t.Cleanup(func() {
		intelligenceFetchSnapshot = previous
		resetIntelligenceCache()
	})
}

func seedIntelligence(t *testing.T, body []byte, checked time.Time, fetchOK bool) {
	t.Helper()
	intelligenceState.Lock()
	defer intelligenceState.Unlock()
	intelligenceState.body = body
	intelligenceState.checked = checked
	intelligenceState.fetchOK = fetchOK
	intelligenceState.inflight = false
}

func validIntelligenceSnapshot(t *testing.T, generatedAt time.Time, dataFresh bool, brief []any) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"schema_version": 1,
		"generated_at":   generatedAt.UTC().Format(time.RFC3339),
		"data_fresh":     dataFresh,
		"sources": []any{map[string]any{
			"id": "openai", "name": "OpenAI", "kind": "lab", "url": "https://openai.com/news/rss.xml",
			"status": "ok", "input_count": 3, "last_success_at": generatedAt.UTC().Format(time.RFC3339),
		}},
		"items":   map[string]any{"radar": []any{}, "experiments": []any{}},
		"alerts":  []any{},
		"brief":   brief,
		"receipt": map[string]any{"input_count": 3, "unique_count": 3, "radar_count": 0, "experiment_count": 0, "excluded_count": 0, "failed_sources": 0, "stale_sources": 0},
		"cache":   map[string]any{"openai": []any{map[string]any{"title": "raw feed entry"}}},
		"docs":    map[string]any{"render": map[string]any{"status": "ok"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func intelligenceResponse(t *testing.T) map[string]any {
	t.Helper()
	response := httptest.NewRecorder()
	handlePublicIntelligence(response, httptest.NewRequest("GET", "/api/intelligence", nil))
	var result map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestPublicIntelligenceUnpublishedFallback(t *testing.T) {
	swapIntelligenceFetch(t, func() ([]byte, error) { return nil, errors.New("offline") })
	result := intelligenceResponse(t)
	if result["delivery_status"] != "unavailable" || result["data_fresh"] != false {
		t.Fatalf("unexpected fallback: %v", result)
	}
	if _, ok := result["sources"].([]any); !ok {
		t.Fatalf("fallback must include a sources array: %v", result)
	}
	receipt, ok := result["receipt"].(map[string]any)
	if !ok {
		t.Fatalf("fallback must include a receipt object: %v", result)
	}
	for _, key := range []string{"input_count", "unique_count", "radar_count", "experiment_count", "excluded_count", "failed_sources", "stale_sources"} {
		if _, ok := receipt[key]; !ok {
			t.Fatalf("fallback receipt missing %q: %v", key, receipt)
		}
	}
}

func TestPublicIntelligenceRejectsStaleBrief(t *testing.T) {
	swapIntelligenceFetch(t, func() ([]byte, error) { return nil, errors.New("offline") })
	old := time.Now().Add(-13 * time.Hour)
	body, err := json.Marshal(map[string]any{"schema_version": 1, "generated_at": old.UTC().Format(time.RFC3339),
		"data_fresh": true, "brief": []any{map[string]any{"label": "Investigate", "title": "Old recommendation"}}})
	if err != nil {
		t.Fatal(err)
	}
	seedIntelligence(t, body, time.Now(), true)
	result := intelligenceResponse(t)
	if result["delivery_status"] != "stale" || result["data_fresh"] != false || len(result["brief"].([]any)) != 0 {
		t.Fatalf("stale recommendation was shown: %v", result)
	}
	alerts := result["alerts"].([]any)
	if len(alerts) == 0 || alerts[0].(map[string]any)["kind"] != "delivery" {
		t.Fatalf("stale snapshot must lead with a delivery alert: %v", alerts)
	}
}

func TestPublicIntelligenceRejectsFutureGeneratedAt(t *testing.T) {
	swapIntelligenceFetch(t, func() ([]byte, error) {
		return validIntelligenceSnapshot(t, time.Now().Add(2*time.Hour), true, nil), nil
	})
	result := intelligenceResponse(t)
	if result["delivery_status"] != "unavailable" {
		t.Fatalf("future snapshot must be rejected at fetch: %v", result)
	}
}

func TestPublicIntelligenceFetchRejectsIncompleteSnapshot(t *testing.T) {
	swapIntelligenceFetch(t, func() ([]byte, error) {
		return []byte(`{"schema_version":1,"generated_at":"` + time.Now().UTC().Format(time.RFC3339) + `"}`), nil
	})
	result := intelligenceResponse(t)
	if result["delivery_status"] != "unavailable" {
		t.Fatalf("snapshot missing required fields must be rejected: %v", result)
	}
}

func TestPublicIntelligenceServesValidSnapshotAndStripsRawCache(t *testing.T) {
	brief := []any{map[string]any{"label": "Investigate", "title": "New model", "url": "/radar"}}
	swapIntelligenceFetch(t, func() ([]byte, error) {
		return validIntelligenceSnapshot(t, time.Now(), true, brief), nil
	})
	result := intelligenceResponse(t)
	if result["delivery_status"] != "ok" || result["data_fresh"] != true {
		t.Fatalf("valid snapshot must be served fresh: %v", result)
	}
	if len(result["brief"].([]any)) != 1 {
		t.Fatalf("fresh brief must be preserved: %v", result["brief"])
	}
	if _, ok := result["cache"]; ok {
		t.Fatalf("raw feed cache must not be served: %v", result["cache"])
	}
	if _, ok := result["docs"]; ok {
		t.Fatalf("docs fingerprints must not be served: %v", result["docs"])
	}
	if _, ok := result["sources"].([]any); !ok {
		t.Fatalf("sources array must be served: %v", result)
	}
}

func TestPublicIntelligenceTrimsBriefWhenDataNotFresh(t *testing.T) {
	brief := []any{
		map[string]any{"label": "Investigate", "title": "Stale radar item"},
		map[string]any{"label": "Check", "title": "OpenAI source failed"},
	}
	swapIntelligenceFetch(t, func() ([]byte, error) {
		return validIntelligenceSnapshot(t, time.Now(), false, brief), nil
	})
	result := intelligenceResponse(t)
	if result["delivery_status"] != "ok" || result["data_fresh"] != false {
		t.Fatalf("snapshot delivered fresh but without fresh data: %v", result)
	}
	kept := result["brief"].([]any)
	if len(kept) != 1 || kept[0].(map[string]any)["label"] != "Check" {
		t.Fatalf("recommendations must be trimmed without fresh data, health pointer kept: %v", kept)
	}
}

func TestPublicIntelligenceFailedRecheckLabelsLastGoodStale(t *testing.T) {
	swapIntelligenceFetch(t, func() ([]byte, error) { return nil, errors.New("timeout") })
	body := validIntelligenceSnapshot(t, time.Now().Add(-2*time.Hour), true,
		[]any{map[string]any{"label": "Investigate", "title": "Old radar item"}})
	seedIntelligence(t, body, time.Now().Add(-31*time.Minute), false)
	result := intelligenceResponse(t)
	if result["delivery_status"] != "stale" || result["data_fresh"] != false {
		t.Fatalf("last-good snapshot must be labeled stale after failed recheck: %v", result)
	}
	if len(result["brief"].([]any)) != 0 {
		t.Fatalf("stale snapshot must not serve recommendations: %v", result["brief"])
	}
}
