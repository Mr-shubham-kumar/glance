package glance

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	publicDiscoveryURL = "https://raw.githubusercontent.com/Mr-shubham-kumar/glance/discovery-data/snapshot.json"
	discoveryCacheTTL  = 30 * time.Minute
	discoveryMaxAge    = 12 * time.Hour
)

type discoveryStory struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Category    string `json:"category"`
	PublishedAt string `json:"published_at,omitempty"`
}

type discoverySnapshot struct {
	SchemaVersion int              `json:"schema_version"`
	GeneratedAt   string           `json:"generated_at"`
	Status        string           `json:"status"`
	Stories       []discoveryStory `json:"stories"`
}

var (
	discoveryClient        = &http.Client{Timeout: 5 * time.Second}
	discoveryFetchSnapshot = func() ([]byte, error) {
		response, err := discoveryClient.Get(publicDiscoveryURL)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("snapshot request returned %s", response.Status)
		}
		body, err := io.ReadAll(io.LimitReader(response.Body, 128*1024+1))
		if err != nil {
			return nil, err
		}
		if len(body) > 128*1024 {
			return nil, errors.New("snapshot exceeds 128 KiB")
		}
		return body, nil
	}
	discoveryCache = struct {
		sync.Mutex
		checked time.Time
		data    discoverySnapshot
	}{data: unavailableDiscovery()}
)

func unavailableDiscovery() discoverySnapshot {
	return discoverySnapshot{SchemaVersion: 1, GeneratedAt: "", Status: "unavailable", Stories: []discoveryStory{}}
}

func validateDiscoverySnapshot(body []byte, now time.Time) (discoverySnapshot, bool) {
	var snapshot discoverySnapshot
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&snapshot) != nil || snapshot.SchemaVersion != 1 || snapshot.Status != "ok" || len(snapshot.Stories) == 0 || len(snapshot.Stories) > 8 {
		return unavailableDiscovery(), false
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return unavailableDiscovery(), false
	}
	generated, err := time.Parse(time.RFC3339, snapshot.GeneratedAt)
	if err != nil || generated.After(now.Add(5*time.Minute)) || now.Sub(generated) > discoveryMaxAge {
		return unavailableDiscovery(), false
	}
	seen := make(map[string]bool)
	for _, story := range snapshot.Stories {
		target, err := url.Parse(story.URL)
		if strings.TrimSpace(story.Title) == "" || strings.TrimSpace(story.Category) == "" || err != nil || target.Scheme != "https" || target.Host != "www.perplexity.ai" || !strings.HasPrefix(target.Path, "/page/") || seen[target.String()] {
			return unavailableDiscovery(), false
		}
		seen[target.String()] = true
		if story.PublishedAt != "" {
			if _, err := time.Parse(time.RFC3339, story.PublishedAt); err != nil {
				return unavailableDiscovery(), false
			}
		}
	}
	return snapshot, true
}

func resetDiscoveryCache() {
	discoveryCache.Lock()
	discoveryCache.checked = time.Time{}
	discoveryCache.data = unavailableDiscovery()
	discoveryCache.Unlock()
}

func currentDiscovery() discoverySnapshot {
	discoveryCache.Lock()
	defer discoveryCache.Unlock()
	if !discoveryCache.checked.IsZero() && time.Since(discoveryCache.checked) < discoveryCacheTTL {
		if discoveryCache.data.Status == "ok" {
			generated, err := time.Parse(time.RFC3339, discoveryCache.data.GeneratedAt)
			if err != nil || time.Since(generated) > discoveryMaxAge {
				discoveryCache.data = unavailableDiscovery()
			}
		}
		return discoveryCache.data
	}
	discoveryCache.checked = time.Now()
	body, err := discoveryFetchSnapshot()
	if err != nil {
		discoveryCache.data = unavailableDiscovery()
		return discoveryCache.data
	}
	if snapshot, ok := validateDiscoverySnapshot(body, time.Now()); ok {
		discoveryCache.data = snapshot
	} else {
		discoveryCache.data = unavailableDiscovery()
	}
	return discoveryCache.data
}

func handleDiscovery(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(currentDiscovery())
}

func handleDiscoveryView(w http.ResponseWriter, _ *http.Request) {
	snapshot := currentDiscovery()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Widget-Title", "Perplexity Discover")
	w.Header().Set("Widget-Title-URL", "https://www.perplexity.ai/discover")
	w.Header().Set("Widget-Content-Type", "html")
	w.Header().Set("Widget-Content-Frameless", "true")
	if snapshot.Status != "ok" {
		_, _ = io.WriteString(w, `<section class="discovery-unavailable"><p>Perplexity picks are temporarily unavailable. The public page could not be checked recently.</p><p><a href="https://www.perplexity.ai/discover" target="_blank" rel="noreferrer">Open Perplexity Discover</a></p></section>`)
		return
	}
	_, _ = io.WriteString(w, `<ol class="discovery-stories">`)
	for _, story := range snapshot.Stories {
		_, _ = fmt.Fprintf(w, `<li class="discovery-story"><a class="discovery-story-title" href="%s" target="_blank" rel="noreferrer">%s</a><div class="discovery-story-meta"><span>%s</span>`, html.EscapeString(story.URL), html.EscapeString(story.Title), html.EscapeString(story.Category))
		if story.PublishedAt != "" {
			published, _ := time.Parse(time.RFC3339, story.PublishedAt)
			age := time.Since(published)
			if age < 0 {
				age = 0
			}
			_, _ = fmt.Fprintf(w, `<span>%s</span>`, html.EscapeString(formatDiscoveryAge(age)))
		}
		_, _ = io.WriteString(w, `</div></li>`)
	}
	_, _ = io.WriteString(w, `</ol>`)
}

func formatDiscoveryAge(age time.Duration) string {
	switch {
	case age < time.Hour:
		return fmt.Sprintf("%dm ago", int(age.Minutes()))
	case age < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(age.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(age.Hours()/24))
	}
}
