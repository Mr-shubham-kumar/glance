package glance

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRedditFeedUsesSubredditAndSkipsRemovedPosts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<feed xmlns="http://www.w3.org/2005/Atom"><title>Reddit</title><link href="https://www.reddit.com/r/selfhosted+ObsidianMD/new/"/><entry><title>[ Removed by Reddit ]</title><link href="https://www.reddit.com/r/selfhosted/comments/old"/><category term="r/selfhosted"/></entry><entry><title>Useful post</title><link href="https://www.reddit.com/r/ObsidianMD/comments/new"/><category term="r/ObsidianMD"/></entry></feed>`))
	}))
	defer server.Close()
	widget := &rssWidget{}
	if err := widget.initialize(); err != nil { t.Fatal(err) }
	items, err := widget.fetchItemsFromFeedTask(rssFeedRequest{URL: server.URL + "/reddit.com/r/test"})
	if err != nil { t.Fatal(err) }
	if len(items) != 1 || items[0].ChannelName != "r/ObsidianMD" || !strings.HasSuffix(items[0].ChannelURL, "/r/ObsidianMD/") {
		t.Fatalf("unexpected Reddit items: %#v", items)
	}
}
