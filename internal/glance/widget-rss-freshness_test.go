package glance

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRSSMaxItemAgeOmitsOldItems(t *testing.T) {
	now := time.Now().UTC()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		fmt.Fprintf(w, `<rss version="2.0"><channel><title>Reading</title><link>https://example.com</link><description>Reading</description><item><title>Old</title><link>https://example.com/old</link><pubDate>%s</pubDate></item><item><title>Fresh</title><link>https://example.com/fresh</link><pubDate>%s</pubDate></item></channel></rss>`, now.Add(-15*24*time.Hour).Format(time.RFC1123Z), now.Add(-time.Hour).Format(time.RFC1123Z))
	}))
	defer server.Close()

	widget := &rssWidget{FeedRequests: []rssFeedRequest{{URL: server.URL}}, MaxItemAge: durationField(7 * 24 * time.Hour)}
	if err := widget.initialize(); err != nil {
		t.Fatal(err)
	}
	widget.update(context.Background())
	if len(widget.Items) != 1 || widget.Items[0].Title != "Fresh" {
		t.Fatalf("expected only fresh item, got %#v", widget.Items)
	}
}
