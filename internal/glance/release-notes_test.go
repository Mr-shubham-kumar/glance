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

func TestPreferMeaningfulReleaseSkipsAlphaAndTagOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		w.Write([]byte(`<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom"><title>Releases</title><link href="https://example.com/"/><entry><title>v3.0.0-alpha.1</title><link href="https://example.com/alpha"/><id>alpha</id><updated>2026-09-26T00:00:00Z</updated><content>Alpha features</content></entry><entry><title>v2.0.0</title><link href="https://example.com/tag"/><id>tag</id><updated>2026-09-25T00:00:00Z</updated><content>Release v2.0.0</content></entry><entry><title>v1.9.0</title><link href="https://example.com/stable"/><id>stable</id><updated>2026-09-24T00:00:00Z</updated><content>Fixed startup bug</content></entry></feed>`))
	}))
	defer server.Close()
	widget := &rssWidget{DescriptionLength: 390}
	items, err := widget.fetchItemsFromFeedTask(rssFeedRequest{URL: server.URL + "/releases.atom", IsDetailed: true, PreferMeaningful: true, Limit: 1})
	if err != nil || len(items) != 1 || items[0].Title != "v1.9.0" {
		t.Fatalf("should prefer stable meaningful notes: %v %v", items, err)
	}
}

func TestRSSDescriptionStripsEncodedComments(t *testing.T) {
	got := sanitizeFeedDescription("&lt;!-- SC_OFF --&gt;<p>Readable Reddit post</p>&lt;!-- SC_ON --&gt;")
	if got != "Readable Reddit post" {
		t.Fatalf("unexpected snippet: %q", got)
	}
}

func TestRSSFeedChannelURLCanPointToOfficialCategory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		w.Write([]byte(`<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom"><title>Trending</title><link href="https://rss-generator.example/"/><entry><title>repo</title><link href="https://github.com/example/repo"/><id>repo</id><updated>2026-09-25T00:00:00Z</updated></entry></feed>`))
	}))
	defer server.Close()
	widget := &rssWidget{}
	items, err := widget.fetchItemsFromFeedTask(rssFeedRequest{URL: server.URL, ChannelURL: "https://github.com/trending/go?since=daily"})
	if err != nil || len(items) != 1 || items[0].ChannelURL != "https://github.com/trending/go?since=daily" || items[0].Link != "https://github.com/example/repo" {
		t.Fatalf("source label should lead to official category; item should lead to repo: %v %v", items, err)
	}
}

func TestPreferMeaningfulFallsBackToStableTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		w.Write([]byte(`<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom"><title>Releases</title><link href="https://example.com/"/><entry><title>v4.0.0-beta.1</title><link href="https://example.com/beta"/><id>beta</id><updated>2026-09-26T00:00:00Z</updated><content>Beta changes</content></entry><entry><title>v3.0.0</title><link href="https://example.com/stable"/><id>stable</id><updated>2026-09-25T00:00:00Z</updated><content>Release v3.0.0</content></entry></feed>`))
	}))
	defer server.Close()
	widget := &rssWidget{DescriptionLength: 390}
	items, err := widget.fetchItemsFromFeedTask(rssFeedRequest{URL: server.URL + "/releases.atom", IsDetailed: true, PreferMeaningful: true, Limit: 1})
	if err != nil || len(items) != 1 || items[0].Title != "v3.0.0" || items[0].Description != "No substantive release notes provided by the maintainer." {
		t.Fatalf("stable tag fallback failed: %v %v", items, err)
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
