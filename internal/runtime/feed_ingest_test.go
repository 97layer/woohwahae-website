package runtime

import (
	"strings"
	"testing"
)

func TestParseFeedDocumentRSS(t *testing.T) {
	raw := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Layer Feed</title>
    <link>https://example.com/feed</link>
    <item>
      <title>First article</title>
      <link>https://example.com/post-1</link>
      <guid>post-1</guid>
      <description>First excerpt.</description>
      <pubDate>Sat, 14 Mar 2026 00:30:00 +0900</pubDate>
      <category>strategy</category>
    </item>
  </channel>
</rss>`)

	feed, err := ParseFeedDocument(raw, "https://example.com/feed.xml", 5)
	if err != nil {
		t.Fatalf("parse rss feed: %v", err)
	}
	if feed.Kind != "rss" || feed.Title != "Layer Feed" || len(feed.Items) != 1 {
		t.Fatalf("unexpected rss feed: %+v", feed)
	}
	if feed.Items[0].ID != "post-1" || feed.Items[0].Link != "https://example.com/post-1" {
		t.Fatalf("unexpected rss item: %+v", feed.Items[0])
	}
	if !containsFeedTag(feed.Items[0].Tags, "strategy") {
		t.Fatalf("expected rss tag normalization, got %+v", feed.Items[0].Tags)
	}
}

func TestParseFeedDocumentAtom(t *testing.T) {
	raw := []byte(`<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Layer Atom</title>
  <link href="https://example.com/atom.xml" rel="self"></link>
  <entry>
    <id>tag:example.com,2026:1</id>
    <title>Atom note</title>
    <link href="https://example.com/atom-note" rel="alternate"></link>
    <summary>Atom excerpt.</summary>
    <updated>2026-03-14T00:45:00+09:00</updated>
    <author><name>Founder</name></author>
    <category term="notes"></category>
  </entry>
</feed>`)

	feed, err := ParseFeedDocument(raw, "https://example.com/atom.xml", 5)
	if err != nil {
		t.Fatalf("parse atom feed: %v", err)
	}
	if feed.Kind != "atom" || feed.Title != "Layer Atom" || len(feed.Items) != 1 {
		t.Fatalf("unexpected atom feed: %+v", feed)
	}
	if !strings.Contains(feed.Items[0].ID, "tag:example.com") || feed.Items[0].Author != "Founder" {
		t.Fatalf("unexpected atom item: %+v", feed.Items[0])
	}
	if !containsFeedTag(feed.Items[0].Tags, "notes") {
		t.Fatalf("expected atom tag normalization, got %+v", feed.Items[0].Tags)
	}
}

func containsFeedTag(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
