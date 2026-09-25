package glance

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDetailedFeedUsesAtomContentWhenSummaryMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		w.Write([]byte(`<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom"><title>Releases</title><link href="https://example.com/"/><entry><title>v2.1.0</title><link href="https://example.com/v2.1.0"/><id>x</id><updated>2026-09-25T00:00:00Z</updated><content type="html">&lt;h2&gt;Fixed&lt;/h2&gt;&lt;ul&gt;&lt;li&gt;Faster startup&lt;/li&gt;&lt;/ul&gt;</content></entry></feed>`))
	}))
	defer server.Close()
	widget := &rssWidget{DescriptionLength: 390}
	items, err := widget.fetchItemsFromFeedTask(rssFeedRequest{URL: server.URL, IsDetailed: true})
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%v err=%v", items, err)
	}
	if !strings.Contains(items[0].Description, "Faster startup") {
		t.Fatalf("missing author notes: %q", items[0].Description)
	}
}

func TestReleaseTagOnlyPattern(t *testing.T) {
	for _, s := range []string{"Release 0.158.0-alpha.13", "release: v2.0.16", "v0.8.6"} {
		if !releaseTagOnlyPattern.MatchString(s) {
			t.Errorf("not identified: %q", s)
		}
	}
	if releaseTagOnlyPattern.MatchString("Fixed: faster startup") {
		t.Fatal("meaningful notes marked empty")
	}
}
