package glance

import (
	"strings"
	"testing"
	"time"
)

func TestDiscoverRSSUsesPublisherExcerptAndEscapesIt(t *testing.T) {
	w := &rssWidget{
		Style: "discover",
		FeedRequests: []rssFeedRequest{{URL: "https://example.com/feed"}},
	}
	w.Type = "rss"
	if err := w.initialize(); err != nil {
		t.Fatal(err)
	}
	if !w.FeedRequests[0].IsDetailed {
		t.Fatal("discover style must extract publisher descriptions")
	}
	w.ContentAvailable = true
	w.Items = rssFeedItemList{{
		Title: "A useful story", Link: "https://example.com/story",
		ChannelName: "Example publisher", Description: "Source excerpt <script>alert(1)</script>",
		PublishedAt: time.Now(),
	}}
	html := string(w.Render())
	for _, want := range []string{"A useful story", "Example publisher", "Source excerpt", "Read at source"} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered card missing %q", want)
		}
	}
	if strings.Contains(html, "<script>") {
		t.Fatal("publisher text was rendered as executable HTML")
	}
}
