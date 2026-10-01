package glance

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDiscoverySnapshotValidationAndFreshness(t *testing.T) {
	now := time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC)
	valid := []byte(`{"schema_version":1,"generated_at":"2026-10-01T07:00:00Z","status":"ok","stories":[{"title":"A story","url":"https://www.perplexity.ai/page/a-123","category":"Science"}]}`)
	if snapshot, ok := validateDiscoverySnapshot(valid, now); !ok || len(snapshot.Stories) != 1 {
		t.Fatalf("valid snapshot rejected: %#v, %v", snapshot, ok)
	}
	for _, test := range []struct {
		name string
		body []byte
	}{
		{"malformed", []byte(`{"schema_version":1}`)},
		{"failed collection", []byte(`{"schema_version":1,"generated_at":"2026-10-01T07:00:00Z","status":"unavailable","stories":[]}`)},
		{"stale", []byte(`{"schema_version":1,"generated_at":"2026-09-30T19:59:59Z","status":"ok","stories":[{"title":"Old","url":"https://www.perplexity.ai/page/old-123","category":"World"}]}`)},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, ok := validateDiscoverySnapshot(test.body, now); ok {
				t.Fatal("snapshot should be unavailable")
			}
		})
	}
}

func TestDiscoveryHandlerRendersUnavailableWithNoStories(t *testing.T) {
	old := discoveryFetchSnapshot
	discoveryFetchSnapshot = func() ([]byte, error) { return []byte(`not json`), nil }
	t.Cleanup(func() { discoveryFetchSnapshot = old; resetDiscoveryCache() })
	resetDiscoveryCache()
	w := httptest.NewRecorder()
	handleDiscovery(w, httptest.NewRequest("GET", "/api/discovery", nil))
	var got discoverySnapshot
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Status != "unavailable" || len(got.Stories) != 0 || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("unexpected response: %#v, headers=%v", got, w.Header())
	}
}

func TestCachedDiscoveryExpiresAtTwelveHours(t *testing.T) {
	old := discoveryFetchSnapshot
	discoveryFetchSnapshot = func() ([]byte, error) { t.Fatal("fresh cache unexpectedly refetched"); return nil, nil }
	t.Cleanup(func() { discoveryFetchSnapshot = old; resetDiscoveryCache() })
	discoveryCache.Lock()
	discoveryCache.checked = time.Now()
	discoveryCache.data = discoverySnapshot{SchemaVersion: 1, GeneratedAt: time.Now().Add(-12*time.Hour - time.Second).UTC().Format(time.RFC3339), Status: "ok", Stories: []discoveryStory{{Title: "Old story", URL: "https://www.perplexity.ai/page/old-123", Category: "World"}}}
	discoveryCache.Unlock()
	w := httptest.NewRecorder()
	handleDiscovery(w, httptest.NewRequest("GET", "/api/discovery", nil))
	var got discoverySnapshot
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Status != "unavailable" || len(got.Stories) != 0 {
		t.Fatalf("expired cached picks were served: %#v", got)
	}
}

func TestDiscoveryViewShowsOnlyPickFields(t *testing.T) {
	old := discoveryFetchSnapshot
	discoveryFetchSnapshot = func() ([]byte, error) {
		return []byte(`{"schema_version":1,"generated_at":"` + time.Now().UTC().Format(time.RFC3339) + `","status":"ok","stories":[{"title":"A title","url":"https://www.perplexity.ai/page/a-123","category":"Science"}]}`), nil
	}
	t.Cleanup(func() { discoveryFetchSnapshot = old; resetDiscoveryCache() })
	resetDiscoveryCache()
	w := httptest.NewRecorder()
	handleDiscoveryView(w, httptest.NewRequest("GET", "/api/discovery/view", nil))
	body := w.Body.String()
	for _, expected := range []string{"A title", "Science", "https://www.perplexity.ai/page/a-123"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("missing rendered field %q in %s", expected, body)
		}
	}
	if strings.Contains(body, "summary") || strings.Contains(body, "excerpt") {
		t.Fatalf("unexpected summary content: %s", body)
	}
}
